package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/meet-clone/backend/internal/core/domain/user"
	"github.com/meet-clone/backend/internal/pkg/errors"
	"github.com/meet-clone/backend/internal/pkg/jwt"
)

type AuthHandler struct {
	userService user.Service
	jwtService  *jwt.JWTService
}

func NewAuthHandler(userService user.Service, jwtService *jwt.JWTService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtService:  jwtService,
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User  *user.User `json:"user"`
	Token string     `json:"token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, errors.NewValidationError("invalid request body"), http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		respondError(w, errors.NewValidationError("email, password, and name are required"), http.StatusBadRequest)
		return
	}

	u, err := h.userService.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			respondError(w, appErr, getStatusCode(appErr.Type))
			return
		}
		respondError(w, errors.NewInternalError("failed to register user", err), http.StatusInternalServerError)
		return
	}

	token, err := h.jwtService.GenerateToken(u.ID, u.Email)
	if err != nil {
		respondError(w, errors.NewInternalError("failed to generate token", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, AuthResponse{User: u, Token: token}, http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, errors.NewValidationError("invalid request body"), http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, errors.NewValidationError("email and password are required"), http.StatusBadRequest)
		return
	}

	u, err := h.userService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			respondError(w, appErr, getStatusCode(appErr.Type))
			return
		}
		respondError(w, errors.NewInternalError("failed to login", err), http.StatusInternalServerError)
		return
	}

	token, err := h.jwtService.GenerateToken(u.ID, u.Email)
	if err != nil {
		respondError(w, errors.NewInternalError("failed to generate token", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, AuthResponse{User: u, Token: token}, http.StatusOK)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*jwt.Claims)
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	u, err := h.userService.GetByID(r.Context(), claims.UserID)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			respondError(w, appErr, getStatusCode(appErr.Type))
			return
		}
		respondError(w, errors.NewInternalError("failed to get user", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, u, http.StatusOK)
}
