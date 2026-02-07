package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/meet-clone/backend/internal/adapters/input/http/middleware"
	"github.com/meet-clone/backend/internal/core/domain/availability"
	"github.com/meet-clone/backend/internal/pkg/errors"
)

type AvailabilityHandler struct {
	service availability.Service
}

func NewAvailabilityHandler(service availability.Service) *AvailabilityHandler {
	return &AvailabilityHandler{
		service: service,
	}
}

func (h *AvailabilityHandler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	avail, err := h.service.GetAvailability(r.Context(), claims.UserID)
	if err != nil {
		respondError(w, errors.NewInternalError("failed to get availability", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, avail, http.StatusOK)
}

func (h *AvailabilityHandler) SaveAvailability(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	var req availability.Availability
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, errors.NewBadRequestError("invalid request body", err), http.StatusBadRequest)
		return
	}

	// For now, accepting full struct, but usually DTO is better. validating schedule logic in service later if needed.
	avail, err := h.service.SaveAvailability(r.Context(), claims.UserID, req.Schedule, req.Timezone, req.IsAcceptingBookings)
	if err != nil {
		respondError(w, errors.NewInternalError("failed to save availability", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, avail, http.StatusOK)
}

func (h *AvailabilityHandler) GetPublicAvailability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"] // For Phase 2 MVP, using UserID. Phase 3 username lookup.

	if userID == "" {
		respondError(w, errors.NewBadRequestError("user id required", nil), http.StatusBadRequest)
		return
	}

	avail, err := h.service.GetAvailability(r.Context(), userID)
	if err != nil {
		respondError(w, errors.NewInternalError("failed to get availability", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, avail, http.StatusOK)
}
