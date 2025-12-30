package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/meet-clone/backend/internal/core/domain/room"
	"github.com/meet-clone/backend/internal/pkg/errors"
	"github.com/meet-clone/backend/internal/pkg/jwt"
)

type RoomHandler struct {
	roomService room.Service
}

func NewRoomHandler(roomService room.Service) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
	}
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*jwt.Claims)
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	rm, err := h.roomService.CreateRoom(r.Context(), claims.UserID)
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
	claims, ok := r.Context().Value("user").(*jwt.Claims)
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
	claims, ok := r.Context().Value("user").(*jwt.Claims)
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
	claims, ok := r.Context().Value("user").(*jwt.Claims)
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
