package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/meet-clone/backend/internal/adapters/input/http/middleware"
	"github.com/meet-clone/backend/internal/core/domain/user"
	"github.com/meet-clone/backend/internal/pkg/errors"
)

type UserHandler struct {
	service user.Service
}

func NewUserHandler(service user.Service) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	user, err := h.service.GetPublicProfile(r.Context(), userID)
	if err != nil {
		respondError(w, errors.NewNotFoundError("User not found"), http.StatusNotFound)
		return
	}

	// Return only public fields
	profile := map[string]interface{}{
		"id":     user.ID,
		"name":   user.Name,
		"avatar": user.Avatar,
		"bio":    user.Bio,
	}

	respondJSON(w, profile, http.StatusOK)
}

type UpdateProfileRequest struct {
	Name string `json:"name"`
	Bio  string `json:"bio"`
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, errors.NewBadRequestError("invalid request body", err), http.StatusBadRequest)
		return
	}

	user, err := h.service.UpdateProfile(r.Context(), claims.UserID, req.Name, req.Bio)
	if err != nil {
		respondError(w, errors.NewInternalError("failed to update profile", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, user, http.StatusOK)
}

// Helper functions (assuming they exist in other handlers or shared)
// If not shared, I'll duplicate them or import them.
// Checking other handlers to see how they respond.
