package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/meet-clone/backend/internal/adapters/input/http/middleware"
	"github.com/meet-clone/backend/internal/core/domain/billing"
	"github.com/meet-clone/backend/internal/core/domain/room"
	"github.com/meet-clone/backend/internal/pkg/cloudflare"
	"github.com/meet-clone/backend/internal/pkg/jwt"
	"github.com/meet-clone/backend/internal/pkg/logger"
)

type CallsHandler struct {
	service        *cloudflare.CallsService
	roomService    room.Service
	billingService *billing.Service
}

func NewCallsHandler(service *cloudflare.CallsService, roomService room.Service, billingService *billing.Service) *CallsHandler {
	return &CallsHandler{
		service:        service,
		roomService:    roomService,
		billingService: billingService,
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

	// Get user ID from context
	userID := ""
	if claims, ok := r.Context().Value(middleware.UserContextKey).(*jwt.Claims); ok {
		userID = claims.UserID
	}

	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Find the room associated with this session to check if user is creator
	room, err := h.roomService.GetRoomBySessionID(r.Context(), req.SessionID)
	if err != nil {
		logger.Error.Printf("Failed to get room for session %s: %v", req.SessionID, err)
		http.Error(w, "Failed to get room", http.StatusNotFound)
		return
	}

	// Check if user is approved to join (creator or approved participant)
	if !room.IsUserApproved(userID) {
		logger.Info.Printf("User %s not approved for room %s, denying token", userID, room.ID)
		http.Error(w, "You must be approved by the host to join this meeting", http.StatusForbidden)
		return
	}

	// Check if user is the room creator
	isCreator := room.CreatedBy == userID
	isWebinar := room.RoomType == "webinar"
	logger.Info.Printf("Generating token for user %s (isCreator: %v, isWebinar: %v) in room %s", userID, isCreator, isWebinar, room.ID)

	token, err := h.service.GenerateToken(req.SessionID, userID, isCreator, isWebinar)
	if err != nil {
		logger.Error.Printf("Failed to generate token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Start billing session
	// We do this asynchronously to not block the token generation
	// But for reliability, synchronous might be better. Billing is critical.
	// Let's do it synchronously for now.
	if err := h.billingService.StartSession(r.Context(), userID, room.ID, req.SessionID, room.CreatedBy, isCreator); err != nil {
		// Log error but don't fail the request, user should still be able to join
		logger.Error.Printf("Failed to start billing session for user %s in room %s: %v", userID, room.ID, err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}
