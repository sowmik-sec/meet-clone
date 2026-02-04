package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/meet-clone/backend/internal/adapters/input/http/middleware"
	"github.com/meet-clone/backend/internal/core/domain/eventtype"
	"github.com/meet-clone/backend/internal/pkg/errors"
)

type EventTypeHandler struct {
	service eventtype.Service
}

func NewEventTypeHandler(service eventtype.Service) *EventTypeHandler {
	return &EventTypeHandler{
		service: service,
	}
}

func (h *EventTypeHandler) CreateEventType(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	var req eventtype.CreateEventTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, errors.NewBadRequestError("invalid request body", err), http.StatusBadRequest)
		return
	}

	et, err := h.service.CreateEventType(r.Context(), claims.UserID, req)
	if err != nil {
		respondError(w, errors.NewInternalError("failed to create event type", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, et, http.StatusCreated)
}

func (h *EventTypeHandler) ListEventTypes(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	list, err := h.service.ListEventTypes(r.Context(), claims.UserID)
	if err != nil {
		respondError(w, errors.NewInternalError("failed to list event types", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, list, http.StatusOK)
}

func (h *EventTypeHandler) UpdateEventType(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	var req eventtype.UpdateEventTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, errors.NewBadRequestError("invalid request body", err), http.StatusBadRequest)
		return
	}

	et, err := h.service.UpdateEventType(r.Context(), id, claims.UserID, req)
	if err != nil {
		respondError(w, errors.NewInternalError("failed to update event type", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, et, http.StatusOK)
}

func (h *EventTypeHandler) DeleteEventType(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeleteEventType(r.Context(), id, claims.UserID); err != nil {
		respondError(w, errors.NewInternalError("failed to delete event type", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *EventTypeHandler) GetPublicEventTypes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	list, err := h.service.GetPublicEventTypes(r.Context(), userID)
	if err != nil {
		respondError(w, errors.NewInternalError("failed to get public event types", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, list, http.StatusOK)
}
