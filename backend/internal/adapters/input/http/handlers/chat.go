package handlers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/meet-clone/backend/internal/core/domain/chat"
	"github.com/meet-clone/backend/internal/pkg/errors"
)

type ChatHandler struct {
	chatService chat.Service
}

func NewChatHandler(chatService chat.Service) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["id"]

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 50
	}

	messages, err := h.chatService.GetMessages(r.Context(), roomID, limit, offset)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			respondError(w, appErr, getStatusCode(appErr.Type))
			return
		}
		respondError(w, errors.NewInternalError("failed to get messages", err), http.StatusInternalServerError)
		return
	}

	respondJSON(w, messages, http.StatusOK)
}
