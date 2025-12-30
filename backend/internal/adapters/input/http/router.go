package http

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	httpHandlers "github.com/meet-clone/backend/internal/adapters/input/http/handlers"
	"github.com/meet-clone/backend/internal/adapters/input/http/middleware"
	"github.com/meet-clone/backend/internal/adapters/input/websocket"
	"github.com/meet-clone/backend/internal/config"
)

type Router struct {
	router         *mux.Router
	authHandler    *httpHandlers.AuthHandler
	roomHandler    *httpHandlers.RoomHandler
	chatHandler    *httpHandlers.ChatHandler
	callsHandler   *httpHandlers.CallsHandler
	wsHandler      *websocket.Handler
	authMiddleware *middleware.AuthMiddleware
	config         *config.Config
}

func NewRouter(
	authHandler *httpHandlers.AuthHandler,
	roomHandler *httpHandlers.RoomHandler,
	chatHandler *httpHandlers.ChatHandler,
	callsHandler *httpHandlers.CallsHandler,
	wsHandler *websocket.Handler,
	authMiddleware *middleware.AuthMiddleware,
	cfg *config.Config,
) *Router {
	return &Router{
		router:         mux.NewRouter(),
		authHandler:    authHandler,
		roomHandler:    roomHandler,
		chatHandler:    chatHandler,
		callsHandler:   callsHandler,
		wsHandler:      wsHandler,
		authMiddleware: authMiddleware,
		config:         cfg,
	}
}

func (r *Router) Setup() http.Handler {
	// Apply CORS
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{r.config.CORSOrigin}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)

	// API version prefix
	api := r.router.PathPrefix("/api/v1").Subrouter()

	// Apply logger middleware
	api.Use(middleware.Logger)

	// Public routes - Auth
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", r.authHandler.Register).Methods("POST")
	auth.HandleFunc("/login", r.authHandler.Login).Methods("POST")

	// Protected routes - Auth
	authProtected := auth.PathPrefix("").Subrouter()
	authProtected.Use(r.authMiddleware.Authenticate)
	authProtected.HandleFunc("/me", r.authHandler.Me).Methods("GET")

	// Protected routes - Rooms
	rooms := api.PathPrefix("/rooms").Subrouter()
	rooms.Use(r.authMiddleware.Authenticate)
	rooms.HandleFunc("", r.roomHandler.CreateRoom).Methods("POST")
	rooms.HandleFunc("/{id}", r.roomHandler.GetRoom).Methods("GET")
	rooms.HandleFunc("/{id}/join", r.roomHandler.JoinRoom).Methods("POST")
	rooms.HandleFunc("/{id}/leave", r.roomHandler.LeaveRoom).Methods("POST")
	rooms.HandleFunc("/{id}", r.roomHandler.EndRoom).Methods("DELETE")
	rooms.HandleFunc("/{id}/participants", r.roomHandler.GetParticipants).Methods("GET")

	// Protected routes - Chat
	chat := api.PathPrefix("/rooms/{id}/messages").Subrouter()
	chat.Use(r.authMiddleware.Authenticate)
	chat.HandleFunc("", r.chatHandler.GetMessages).Methods("GET")

	// WebSocket route
	api.HandleFunc("/ws/room/{id}", r.wsHandler.HandleWebSocket)

	// Protected routes - Calls
	calls := api.PathPrefix("/calls").Subrouter()
	// calls.Use(r.authMiddleware.Authenticate) // Uncomment if authentication is needed, but for now we might leave it open or handle in handler
	calls.HandleFunc("/sessions", r.callsHandler.CreateSession).Methods("POST")
	calls.HandleFunc("/sessions/token", r.callsHandler.GenerateToken).Methods("POST")

	// Health check
	r.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Wrap with CORS
	// r.router.Use(corsHandler) // Use wrapper instead for better preflight handling

	return corsHandler(r.router)
}
