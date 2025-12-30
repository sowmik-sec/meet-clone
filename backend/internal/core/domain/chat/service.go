package chat

import (
	"context"

	"github.com/meet-clone/backend/internal/pkg/errors"
)

type Service interface {
	SendMessage(ctx context.Context, roomID, userID, userName, message string) (*Message, error)
	GetMessages(ctx context.Context, roomID string, limit, offset int) ([]*Message, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) SendMessage(ctx context.Context, roomID, userID, userName, message string) (*Message, error) {
	if message == "" {
		return nil, errors.NewValidationError("message cannot be empty")
	}

	msg := NewMessage(roomID, userID, userName, message)

	if err := s.repo.Create(ctx, msg); err != nil {
		return nil, errors.NewInternalError("failed to save message", err)
	}

	return msg, nil
}

func (s *service) GetMessages(ctx context.Context, roomID string, limit, offset int) ([]*Message, error) {
	messages, err := s.repo.FindByRoomID(ctx, roomID, limit, offset)
	if err != nil {
		return nil, errors.NewInternalError("failed to retrieve messages", err)
	}

	return messages, nil
}
