package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/meet-clone/backend/internal/core/domain/room"
	"github.com/meet-clone/backend/internal/pkg/cloudflare"
	"github.com/meet-clone/backend/internal/pkg/logger"
)

type CallsHandler struct {
	service     *cloudflare.CallsService
	roomService room.Service
}

func NewCallsHandler(service *cloudflare.CallsService, roomService room.Service) *CallsHandler {
	return &CallsHandler{
		service:     service,
		roomService: roomService,
	}
}

type CreateSessionRequest struct {
	RoomID string `json:"roomId"`
}

func (h *CallsHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error.Printf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RoomID == "" {
		logger.Error.Println("Room ID is required but not provided")
		http.Error(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	logger.Info.Printf("Creating Cloudflare session for room: %s", req.RoomID)

	// Check if room exists and has a session
	roomDetails, err := h.roomService.GetRoomDetails(r.Context(), req.RoomID)
	if err != nil {
		logger.Error.Printf("Failed to get room %s: %v", req.RoomID, err)
		http.Error(w, "Failed to get room", http.StatusNotFound)
		return
	}

	// If room already has a session ID, return it
	if roomDetails.CloudflareSessionID != "" {
		logger.Info.Printf("Room %s already has session ID: %s", req.RoomID, roomDetails.CloudflareSessionID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"sessionId": roomDetails.CloudflareSessionID,
		})
		return
	}

	logger.Info.Printf("Creating new Cloudflare session for room: %s", req.RoomID)
	session, err := h.service.CreateSession(req.RoomID)
	if err != nil {
		logger.Error.Printf("Failed to create Cloudflare session for room %s: %v", req.RoomID, err)
		http.Error(w, fmt.Sprintf("Failed to create session: %v", err), http.StatusInternalServerError)
		return
	}

	logger.Info.Printf("Successfully created Cloudflare session %s for room %s", session.SessionID, req.RoomID)

	// Save session ID to room
	if err := h.roomService.SetSessionID(r.Context(), req.RoomID, session.SessionID); err != nil {
		logger.Error.Printf("Failed to update room %s with session ID %s: %v", req.RoomID, session.SessionID, err)
		http.Error(w, "Failed to persist session", http.StatusInternalServerError)
		return
	}

	logger.Info.Printf("Successfully saved session ID %s to room %s", session.SessionID, req.RoomID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

type GenerateTokenRequest struct {
	SessionID string `json:"sessionId"`
}

func (h *CallsHandler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	var req GenerateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	token, err := h.service.GenerateToken(req.SessionID)
	if err != nil {
		logger.Error.Printf("Failed to generate token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}
