package room

import (
	"context"

	"github.com/meet-clone/backend/internal/pkg/errors"
)

type Service interface {
	CreateRoom(ctx context.Context, userID string) (*Room, error)
	JoinRoom(ctx context.Context, roomID, userID, userName, avatar string) (*Room, error)
	LeaveRoom(ctx context.Context, roomID, userID string) (*Room, error)
	GetRoomDetails(ctx context.Context, roomID string) (*Room, error)
	EndRoom(ctx context.Context, roomID, userID string) error
	GetActiveParticipants(ctx context.Context, roomID string) ([]Participant, error)
	SetSessionID(ctx context.Context, roomID, sessionID string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateRoom(ctx context.Context, userID string) (*Room, error) {
	room := NewRoom(userID, 10) // Max 10 participants for MVP

	if err := s.repo.Create(ctx, room); err != nil {
		return nil, errors.NewInternalError("failed to create room", err)
	}

	return room, nil
}

func (s *service) JoinRoom(ctx context.Context, roomID, userID, userName, avatar string) (*Room, error) {
	room, err := s.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, errors.NewNotFoundError("room not found")
	}

	if !room.IsActive() {
		return nil, errors.NewValidationError("room has ended")
	}

	if err := room.AddParticipant(userID, userName, avatar); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	if err := s.repo.Update(ctx, room); err != nil {
		return nil, errors.NewInternalError("failed to update room", err)
	}

	return room, nil
}

func (s *service) LeaveRoom(ctx context.Context, roomID, userID string) (*Room, error) {
	room, err := s.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, errors.NewNotFoundError("room not found")
	}

	if err := room.RemoveParticipant(userID); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	// End room if no active participants
	if len(room.GetActiveParticipants()) == 0 {
		room.End()
	}

	if err := s.repo.Update(ctx, room); err != nil {
		return nil, errors.NewInternalError("failed to update room", err)
	}

	return room, nil
}

func (s *service) GetRoomDetails(ctx context.Context, roomID string) (*Room, error) {
	room, err := s.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, errors.NewNotFoundError("room not found")
	}

	return room, nil
}

func (s *service) EndRoom(ctx context.Context, roomID, userID string) error {
	room, err := s.repo.FindByID(ctx, roomID)
	if err != nil {
		return errors.NewNotFoundError("room not found")
	}

	if room.CreatedBy != userID {
		return errors.NewForbiddenError("only the room creator can end the room")
	}

	room.End()

	if err := s.repo.Update(ctx, room); err != nil {
		return errors.NewInternalError("failed to end room", err)
	}

	return nil
}

func (s *service) GetActiveParticipants(ctx context.Context, roomID string) ([]Participant, error) {
	room, err := s.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, errors.NewNotFoundError("room not found")
	}

	return room.GetActiveParticipants(), nil
}
