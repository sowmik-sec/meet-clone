package room

import (
	"context"

	"github.com/meet-clone/backend/internal/pkg/errors"
)

func (s *service) SetSessionID(ctx context.Context, roomID, sessionID string) error {
	room, err := s.repo.FindByID(ctx, roomID)
	if err != nil {
		return errors.NewNotFoundError("room not found")
	}

	room.CloudflareSessionID = sessionID

	if err := s.repo.Update(ctx, room); err != nil {
		return errors.NewInternalError("failed to update room", err)
	}

	return nil
}
