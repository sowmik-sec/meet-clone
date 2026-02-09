package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/meet-clone/backend/internal/adapters/input/http/middleware"
	"github.com/meet-clone/backend/internal/core/domain/billing"
	"github.com/meet-clone/backend/internal/pkg/jwt"
	"github.com/meet-clone/backend/internal/pkg/logger"
)

type BillingHandler struct {
	service *billing.Service
}

func NewBillingHandler(service *billing.Service) *BillingHandler {
	return &BillingHandler{
		service: service,
	}
}

type EndSessionRequest struct {
	RoomID string `json:"roomId"`
}

func (h *BillingHandler) EndSession(w http.ResponseWriter, r *http.Request) {
	var req EndSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := ""
	if claims, ok := r.Context().Value(middleware.UserContextKey).(*jwt.Claims); ok {
		userID = claims.UserID
	}

	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.service.EndSession(r.Context(), userID, req.RoomID); err != nil {
		logger.Error.Printf("Failed to end session: %v", err)
		http.Error(w, "Failed to end session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BillingHandler) GetCurrentUsage(w http.ResponseWriter, r *http.Request) {
	userID := ""
	if claims, ok := r.Context().Value(middleware.UserContextKey).(*jwt.Claims); ok {
		userID = claims.UserID
	}

	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	usage, err := h.service.GetCurrentPeriodUsage(r.Context(), userID)
	if err != nil {
		logger.Error.Printf("Failed to get current usage: %v", err)
		http.Error(w, "Failed to get usage stats", http.StatusInternalServerError)
		return
	}

	if usage == nil {
		// Return empty usage object if none exists yet
		usage = &billing.UserBillingPeriod{
			UserID:        userID,
			BillingPeriod: time.Now().Format("2006-01"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}

func (h *BillingHandler) GetUsageByPeriod(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	period := vars["period"] // e.g., "2026-02"

	userID := ""
	if claims, ok := r.Context().Value(middleware.UserContextKey).(*jwt.Claims); ok {
		userID = claims.UserID
	}

	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	usage, err := h.service.GetBillingPeriod(r.Context(), userID, period)
	if err != nil {
		logger.Error.Printf("Failed to get usage for period %s: %v", period, err)
		http.Error(w, "Failed to get usage stats", http.StatusInternalServerError)
		return
	}

	if usage == nil {
		http.Error(w, "Usage data not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}

func (h *BillingHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID := ""
	if claims, ok := r.Context().Value(middleware.UserContextKey).(*jwt.Claims); ok {
		userID = claims.UserID
	}

	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 12 // Default 12 months
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	history, err := h.service.GetUserHistory(r.Context(), userID, limit)
	if err != nil {
		logger.Error.Printf("Failed to get history: %v", err)
		http.Error(w, "Failed to get history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (h *BillingHandler) SyncUsage(w http.ResponseWriter, r *http.Request) {
	// TODO: Add admin check
	userID := ""
	if claims, ok := r.Context().Value(middleware.UserContextKey).(*jwt.Claims); ok {
		userID = claims.UserID
	}

	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.service.SyncCloudflareUsage(r.Context()); err != nil {
		logger.Error.Printf("Failed to sync usage: %v", err)
		http.Error(w, "Failed to sync usage", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success", "message":"Sync completed"}`))
}
