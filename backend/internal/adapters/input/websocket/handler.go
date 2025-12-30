package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/meet-clone/backend/internal/core/domain/chat"
	"github.com/meet-clone/backend/internal/pkg/jwt"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type Client struct {
	conn   *websocket.Conn
	roomID string
	userID string
	send   chan []byte
}

type Hub struct {
	rooms       map[string]map[*Client]bool
	broadcast   chan *Message
	register    chan *Client
	unregister  chan *Client
	mu          sync.RWMutex
	chatService chat.Service
}

type Message struct {
	Type    string      `json:"type"`
	RoomID  string      `json:"room_id"`
	UserID  string      `json:"user_id"`
	Payload interface{} `json:"payload"`
}

func NewHub(chatService chat.Service) *Hub {
	return &Hub{
		rooms:       make(map[string]map[*Client]bool),
		broadcast:   make(chan *Message, 256),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		chatService: chatService,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.rooms[client.roomID]; !ok {
				h.rooms[client.roomID] = make(map[*Client]bool)
			}
			h.rooms[client.roomID][client] = true
			h.mu.Unlock()

			// Notify others about new participant
			h.broadcastToRoom(client.roomID, &Message{
				Type:   "participant_joined",
				RoomID: client.roomID,
				UserID: client.userID,
			})

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[client.roomID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.rooms, client.roomID)
					}
				}
			}
			h.mu.Unlock()

			// Notify others about participant leaving
			h.broadcastToRoom(client.roomID, &Message{
				Type:   "participant_left",
				RoomID: client.roomID,
				UserID: client.userID,
			})

		case message := <-h.broadcast:
			h.broadcastToRoom(message.RoomID, message)
		}
	}
}

func (h *Hub) broadcastToRoom(roomID string, message *Message) {
	h.mu.RLock()
	clients := h.rooms[roomID]
	h.mu.RUnlock()

	data, _ := json.Marshal(message)
	for client := range clients {
		select {
		case client.send <- data:
		default:
			h.mu.Lock()
			delete(h.rooms[roomID], client)
			h.mu.Unlock()
			close(client.send)
		}
	}
}

type Handler struct {
	hub        *Hub
	jwtService *jwt.JWTService
}

func NewHandler(hub *Hub, jwtService *jwt.JWTService) *Handler {
	return &Handler{
		hub:        hub,
		jwtService: jwtService,
	}
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get room ID from URL
	vars := mux.Vars(r)
	roomID := vars["id"]

	// Authenticate user via query parameter token
	token := r.URL.Query().Get("token")
	claims, err := h.jwtService.ValidateToken(token)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		conn:   conn,
		roomID: roomID,
		userID: claims.UserID,
		send:   make(chan []byte, 256),
	}

	h.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump(h.hub)
}

func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle different message types
		switch msg.Type {
		case "chat_message":
			if payload, ok := msg.Payload.(map[string]interface{}); ok {
				message, _ := payload["message"].(string)
				userName, _ := payload["user_name"].(string)

				if message != "" {
					_, err := hub.chatService.SendMessage(
						context.Background(),
						c.roomID,
						c.userID,
						userName,
						message,
					)
					if err != nil {
						log.Printf("Failed to save message: %v", err)
						continue
					}

					// Broadcast to room
					hub.broadcast <- &Message{
						Type:    "chat_message",
						RoomID:  c.roomID,
						UserID:  c.userID,
						Payload: payload,
					}
				}
			}
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}
