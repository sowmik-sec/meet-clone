package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/meet-clone/backend/internal/pkg/errors"
)

type ErrorResponse struct {
	Error string `json:"error"`
	Type  string `json:"type"`
}

func respondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, err *errors.AppError, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: err.Message,
		Type:  string(err.Type),
	})
}

func getStatusCode(errorType errors.ErrorType) int {
	switch errorType {
	case errors.ErrorTypeValidation:
		return http.StatusBadRequest
	case errors.ErrorTypeNotFound:
		return http.StatusNotFound
	case errors.ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case errors.ErrorTypeForbidden:
		return http.StatusForbidden
	case errors.ErrorTypeAlreadyExists:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
