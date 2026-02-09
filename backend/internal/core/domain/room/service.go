package room

import (
	"context"
	"fmt"

	"github.com/meet-clone/backend/internal/pkg/errors"
)

type Service interface {
	CreateRoom(ctx context.Context, userID string, roomType RoomType) (*Room, error)
	JoinRoom(ctx context.Context, roomID, userID, userName, avatar string) (*Room, error)
	LeaveRoom(ctx context.Context, roomID, userID string) (*Room, error)
	GetRoomDetails(ctx context.Context, roomID string) (*Room, error)
	GetUserRooms(ctx context.Context, userID string) ([]*Room, error)
	EndRoom(ctx context.Context, roomID, userID string) error
	GetActiveParticipants(ctx context.Context, roomID string) ([]Participant, error)
	SetSessionID(ctx context.Context, roomID, sessionID string) error
	GetRoomBySessionID(ctx context.Context, sessionID string) (*Room, error)
	RequestJoin(ctx context.Context, roomID, userID, userName, avatar string) (*Room, error)
	ApproveParticipant(ctx context.Context, roomID, hostID, userID string) (*Room, error)
	DenyParticipant(ctx context.Context, roomID, hostID, userID string) error
}

type CallsManager interface {
	DeleteSession(sessionID string) error
}

type service struct {
	repo         Repository
	callsManager CallsManager
}

func NewService(repo Repository, callsManager CallsManager) Service {
	return &service{
		repo:         repo,
		callsManager: callsManager,
	}
}

func (s *service) CreateRoom(ctx context.Context, userID string, roomType RoomType) (*Room, error) {
	// Set max capacity based on room type
	maxCapacity := 10
	if roomType == RoomTypeWebinar {
		maxCapacity = 100 // Webinars support more viewers
	}
	room := NewRoom(userID, maxCapacity, roomType)

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

	// Creator joins directly
	if room.CreatedBy == userID {
		if err := room.AddParticipant(userID, userName, avatar); err != nil {
			return nil, errors.NewValidationError(err.Error())
		}
		if err := s.repo.Update(ctx, room); err != nil {
			return nil, errors.NewInternalError("failed to update room", err)
		}
		return room, nil
	}

	// Non-creator goes to waiting list
	if err := room.AddToWaitingList(userID, userName, avatar); err != nil {
		// If already in waiting list or room, return the room without error
		return room, nil
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

func (s *service) GetUserRooms(ctx context.Context, userID string) ([]*Room, error) {
	rooms, err := s.repo.FindByCreator(ctx, userID)
	if err != nil {
		return nil, errors.NewInternalError("failed to get user rooms", err)
	}

	return rooms, nil
}

func (s *service) SetSessionID(ctx context.Context, roomID, sessionID string) error {
	room, err := s.repo.FindByID(ctx, roomID)
	if err != nil {
		return errors.NewNotFoundError("room not found")
	}

	room.CloudflareSessionID = sessionID
	if err := s.repo.Update(ctx, room); err != nil {
		return errors.NewInternalError("failed to update room session id", err)
	}

	return nil
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

	// Also delete Cloudflare session to kick everyone out
	if room.CloudflareSessionID != "" && s.callsManager != nil {
		if err := s.callsManager.DeleteSession(room.CloudflareSessionID); err != nil {
			// Log error but don't fail, DB state is already updated
			fmt.Printf("Failed to delete Cloudflare session %s: %v\n", room.CloudflareSessionID, err)
		}
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

func (s *service) GetRoomBySessionID(ctx context.Context, sessionID string) (*Room, error) {
	room, err := s.repo.FindBySessionID(ctx, sessionID)
	if err != nil {
		return nil, errors.NewNotFoundError("room not found")
	}

	return room, nil
}

func (s *service) RequestJoin(ctx context.Context, roomID, userID, userName, avatar string) (*Room, error) {
	return s.JoinRoom(ctx, roomID, userID, userName, avatar)
}

func (s *service) ApproveParticipant(ctx context.Context, roomID, hostID, userID string) (*Room, error) {
	room, err := s.repo.FindByID(ctx, roomID)
	if err != nil {
		return nil, errors.NewNotFoundError("room not found")
	}

	// Only room creator can approve
	if room.CreatedBy != hostID {
		return nil, errors.NewForbiddenError("only room creator can approve participants")
	}

	if err := room.ApproveParticipant(userID); err != nil {
		return nil, errors.NewValidationError(err.Error())
	}

	if err := s.repo.Update(ctx, room); err != nil {
		return nil, errors.NewInternalError("failed to update room", err)
	}

	return room, nil
}

func (s *service) DenyParticipant(ctx context.Context, roomID, hostID, userID string) error {
	room, err := s.repo.FindByID(ctx, roomID)
	if err != nil {
		return errors.NewNotFoundError("room not found")
	}

	// Only room creator can deny
	if room.CreatedBy != hostID {
		return errors.NewForbiddenError("only room creator can deny participants")
	}

	if err := room.DenyParticipant(userID); err != nil {
		return errors.NewValidationError(err.Error())
	}

	if err := s.repo.Update(ctx, room); err != nil {
		return errors.NewInternalError("failed to update room", err)
	}

	return nil
}
