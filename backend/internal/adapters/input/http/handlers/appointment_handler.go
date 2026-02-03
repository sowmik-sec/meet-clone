package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/meet-clone/backend/internal/adapters/input/http/middleware"
	"github.com/meet-clone/backend/internal/core/domain/appointment"
	"github.com/meet-clone/backend/internal/pkg/errors"
)

type AppointmentHandler struct {
	service appointment.Service
}

func NewAppointmentHandler(service appointment.Service) *AppointmentHandler {
	return &AppointmentHandler{
		service: service,
	}
}

func (h *AppointmentHandler) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	var req appointment.CreateAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, errors.NewBadRequestError("invalid request body", err), http.StatusBadRequest)
		return
	}

	appt, err := h.service.CreateAppointment(r.Context(), claims.UserID, req)
	if err != nil {
		respondError(w, errors.NewInternalError("failed to create appointment", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, appt, http.StatusCreated)
}

func (h *AppointmentHandler) GetUserAppointments(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	// Parse query params for filtering
	query := r.URL.Query()
	startTimeAfter := query.Get("start_time_after")
	startTimeBefore := query.Get("start_time_before")
	statusStr := query.Get("status")

	var status *appointment.AppointmentStatus
	if statusStr != "" {
		s := appointment.AppointmentStatus(statusStr)
		status = &s
	}

	filter := appointment.AppointmentFilter{
		Status: status,
	}
	if startTimeAfter != "" {
		filter.StartTimeAfter = &startTimeAfter
	}
	if startTimeBefore != "" {
		filter.StartTimeBefore = &startTimeBefore
	}

	appointments, err := h.service.GetUserAppointments(r.Context(), claims.UserID, filter)
	if err != nil {
		respondError(w, errors.NewInternalError("failed to get appointments", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, appointments, http.StatusOK)
}

func (h *AppointmentHandler) ConfirmAppointment(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.ConfirmAppointment(r.Context(), id, claims.UserID); err != nil {
		respondError(w, errors.NewInternalError("failed to confirm appointment", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AppointmentHandler) CancelAppointment(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.CancelAppointment(r.Context(), id, claims.UserID); err != nil {
		respondError(w, errors.NewInternalError("failed to cancel appointment", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AppointmentHandler) StartAppointment(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondError(w, errors.NewUnauthorizedError("unauthorized"), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	roomID, err := h.service.StartAppointment(r.Context(), id, claims.UserID)
	if err != nil {
		respondError(w, errors.NewInternalError("failed to start appointment", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{"room_id": roomID}, http.StatusOK)
}
