package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/meet-clone/backend/internal/adapters/input/http/middleware"
	"github.com/meet-clone/backend/internal/config"
	"github.com/meet-clone/backend/internal/core/domain/user"
	"github.com/meet-clone/backend/internal/pkg/calendar"
	"github.com/meet-clone/backend/internal/pkg/errors"
	"github.com/meet-clone/backend/internal/pkg/jwt"
)

type AuthHandler struct {
	userService   user.Service
	jwtService    *jwt.JWTService
	config        *config.Config
	googleService calendar.Service
}

func NewAuthHandler(userService user.Service, jwtService *jwt.JWTService, cfg *config.Config, googleService calendar.Service) *AuthHandler {
	return &AuthHandler{
		userService:   userService,
		jwtService:    jwtService,
		config:        cfg,
		googleService: googleService,
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

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := h.googleService.GetAuthURL()
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		respondError(w, errors.NewBadRequestError("code is required", nil), http.StatusBadRequest)
		return
	}

	token, err := h.googleService.ExchangeToken(r.Context(), code)
	if err != nil {
		respondError(w, errors.NewInternalError("failed to exchange token", err), http.StatusInternalServerError)
		return
	}

	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		// If unauthenticated, we can't link account.
		// For "Sync Calendar", user MUST be logged in.
		// If implementation supports "Login with Google", logic is different.
		// Assuming "Sync" use case here as per requirement.
		respondError(w, errors.NewUnauthorizedError("must be logged in to sync calendar"), http.StatusUnauthorized)
		return
	}

	err = h.userService.UpdateGoogleToken(r.Context(), claims.UserID, token.AccessToken, token.RefreshToken, token.Expiry)
	if err != nil {
		respondError(w, errors.NewInternalError("failed to update user token", err), http.StatusInternalServerError)
		return
	}

	// Redirect back to settings or dashboard
	http.Redirect(w, r, h.config.CORSOrigin+"/dashboard?google_linked=true", http.StatusTemporaryRedirect)
}
