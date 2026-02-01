package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/meet-clone/backend/internal/adapters/input/http/middleware"
	"github.com/meet-clone/backend/internal/config"
	"github.com/meet-clone/backend/internal/core/domain/user"
	"github.com/meet-clone/backend/internal/pkg/errors"
	"github.com/meet-clone/backend/internal/pkg/jwt"
)

type AuthHandler struct {
	userService user.Service
	jwtService  *jwt.JWTService
	config      *config.Config
}

func NewAuthHandler(userService user.Service, jwtService *jwt.JWTService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtService:  jwtService,
		config:      cfg,
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
	User *user.User `json:"user"`
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

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour), // 24 hours
		HttpOnly: true,
		Path:     "/",
		Secure:   h.config.Environment == "production",
		SameSite: http.SameSiteLaxMode,
	})

	respondJSON(w, AuthResponse{User: u}, http.StatusCreated)
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

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour), // 24 hours
		HttpOnly: true,
		Path:     "/",
		Secure:   h.config.Environment == "production",
		SameSite: http.SameSiteLaxMode,
	})

	respondJSON(w, AuthResponse{User: u}, http.StatusOK)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
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

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
}
