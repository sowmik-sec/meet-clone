package handlers

import (
	"encoding/json"
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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RoomID == "" {
		http.Error(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	// Check if room exists and has a session
	roomDetails, err := h.roomService.GetRoomDetails(r.Context(), req.RoomID)
	if err != nil {
		logger.Error.Printf("Failed to get room: %v", err)
		http.Error(w, "Failed to get room", http.StatusNotFound)
		return
	}

	// If room already has a session ID, return it
	if roomDetails.CloudflareSessionID != "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"sessionId": roomDetails.CloudflareSessionID,
		})
		return
	}

	session, err := h.service.CreateSession(req.RoomID)
	if err != nil {
		logger.Error.Printf("Failed to create session: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Save session ID to room
	if err := h.roomService.SetSessionID(r.Context(), req.RoomID, session.SessionID); err != nil {
		logger.Error.Printf("Failed to update room with session ID: %v", err)
		http.Error(w, "Failed to persist session", http.StatusInternalServerError)
		return
	}

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
