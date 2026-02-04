package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/meet-clone/backend/internal/adapters/input/http/middleware"
	"github.com/meet-clone/backend/internal/core/domain/appointment"
	"github.com/meet-clone/backend/internal/core/domain/room"
	"github.com/meet-clone/backend/internal/pkg/errors"
)

type RoomHandler struct {
	roomService        room.Service
	appointmentService appointment.Service
}

func NewRoomHandler(roomService room.Service, appointmentService appointment.Service) *RoomHandler {
	return &RoomHandler{
		roomService:        roomService,
		appointmentService: appointmentService,
	}
}

type CreateRoomRequest struct {
	RoomType string `json:"room_type"` // "meeting" or "webinar"
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	var req CreateRoomRequest
	json.NewDecoder(r.Body).Decode(&req)

	// Default to meeting type if not specified
	roomType := room.RoomTypeMeeting
	if req.RoomType == "webinar" {
		roomType = room.RoomTypeWebinar
	}

	rm, err := h.roomService.CreateRoom(r.Context(), claims.UserID, roomType)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			respondError(w, appErr, getStatusCode(appErr.Type))
			return
		}
		respondError(w, errors.NewInternalError("failed to create room", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, rm, http.StatusCreated)
}

func (h *RoomHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["id"]

	rm, err := h.roomService.GetRoomDetails(r.Context(), roomID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			respondError(w, appErr, getStatusCode(appErr.Type))
			return
		}
		respondError(w, errors.NewInternalError("failed to get room", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, rm, http.StatusOK)
}

type JoinRoomRequest struct {
	UserName string `json:"user_name"`
	Avatar   string `json:"avatar"`
}

func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	roomID := vars["id"]

	var req JoinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, errors.NewValidationError("invalid request body"), http.StatusBadRequest)
		return
	}

	// Check Access Control
	appt, _ := h.appointmentService.GetAppointmentByRoomID(r.Context(), roomID)
	// If it's an appointment room, validate host/guest and time
	if appt != nil {
		if appt.HostID != claims.UserID && appt.GuestID != claims.UserID {
			// Not host or guest. Check if public participants are allowed?
			// For now, strict: only host and guest.
			// Ideally we should check if they are already approved participants in the room too.
			respondError(w, errors.NewForbiddenError("access denied: appointment restricted"), http.StatusForbidden)
			return
		}

		// Time validation
		// Allow 15 mins before start, until end time
		/*
			now := time.Now()
			windowStart := appt.StartTime.Add(-15 * time.Minute)
			if now.Before(windowStart) {
				respondError(w, errors.NewForbiddenError("meeting has not started yet"), http.StatusForbidden)
				return
			}
			if now.After(appt.EndTime) {
			    // Maybe allow a grace period?
			}
		*/
	}

	rm, err := h.roomService.JoinRoom(r.Context(), roomID, claims.UserID, req.UserName, req.Avatar)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			respondError(w, appErr, getStatusCode(appErr.Type))
			return
		}
		respondError(w, errors.NewInternalError("failed to join room", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, rm, http.StatusOK)
}

func (h *RoomHandler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	roomID := vars["id"]

	rm, err := h.roomService.LeaveRoom(r.Context(), roomID, claims.UserID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			respondError(w, appErr, getStatusCode(appErr.Type))
			return
		}
		respondError(w, errors.NewInternalError("failed to leave room", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, rm, http.StatusOK)
}

func (h *RoomHandler) EndRoom(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	roomID := vars["id"]

	err := h.roomService.EndRoom(r.Context(), roomID, claims.UserID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			respondError(w, appErr, getStatusCode(appErr.Type))
			return
		}
		respondError(w, errors.NewInternalError("failed to end room", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RoomHandler) GetParticipants(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["id"]

	participants, err := h.roomService.GetActiveParticipants(r.Context(), roomID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			respondError(w, appErr, getStatusCode(appErr.Type))
			return
		}
		respondError(w, errors.NewInternalError("failed to get participants", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, participants, http.StatusOK)
}

func (h *RoomHandler) GetUserRooms(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	rooms, err := h.roomService.GetUserRooms(r.Context(), claims.UserID)
	if err != nil {
		log.Printf("Error getting user rooms: %v", err) // Log to stdout
		http.Error(w, "internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, rooms, http.StatusOK)
}

func (h *RoomHandler) ApproveParticipant(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	roomID := vars["id"]
	userID := vars["userId"]

	room, err := h.roomService.ApproveParticipant(r.Context(), roomID, claims.UserID, userID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			respondError(w, appErr, getStatusCode(appErr.Type))
			return
		}
		respondError(w, errors.NewInternalError("failed to approve participant", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, room, http.StatusOK)
}

func (h *RoomHandler) DenyParticipant(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	roomID := vars["id"]
	userID := vars["userId"]

	err := h.roomService.DenyParticipant(r.Context(), roomID, claims.UserID, userID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			respondError(w, appErr, getStatusCode(appErr.Type))
			return
		}
		respondError(w, errors.NewInternalError("failed to deny participant", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
