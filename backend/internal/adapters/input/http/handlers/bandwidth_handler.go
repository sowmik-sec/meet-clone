package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/meet-clone/backend/internal/adapters/input/http/middleware"
	"github.com/meet-clone/backend/internal/core/domain/bandwidth"
)

type BandwidthHandler struct {
	service *bandwidth.Service
}

func NewBandwidthHandler(service *bandwidth.Service) *BandwidthHandler {
	return &BandwidthHandler{
		service: service,
	}
}

type ReportBandwidthRequest struct {
	RoomID        string `json:"room_id"`
	SessionID     string `json:"session_id"`
	BytesSent     int64  `json:"bytes_sent"`
	BytesReceived int64  `json:"bytes_received"`
	PacketsSent   int64  `json:"packets_sent"`
	PacketsLost   int64  `json:"packets_lost"`
	Duration      int64  `json:"duration"`
}

func (h *BandwidthHandler) ReportBandwidth(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	var req ReportBandwidthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	log.Printf("Bandwidth Report Received: %+v", req) // DEBUG LOG

	if req.RoomID == "" {
		http.Error(w, "room_id is required", http.StatusBadRequest)
		return
	}

	err := h.service.ReportBandwidth(
		r.Context(),
		userID,
		req.RoomID,
		req.SessionID,
		req.BytesSent,
		req.BytesReceived,
		req.PacketsSent,
		req.PacketsLost,
		req.Duration,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BandwidthHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	stats, err := h.service.GetUserStats(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *BandwidthHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	history, err := h.service.GetUserHistory(r.Context(), userID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
