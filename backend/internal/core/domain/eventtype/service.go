package eventtype

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateEventType(ctx context.Context, userID string, req CreateEventTypeRequest) (*EventType, error) {
	if req.Title == "" {
		return nil, errors.New("title is required")
	}
	if req.Slug == "" {
		return nil, errors.New("slug is required")
	}
	if req.Duration <= 0 {
		return nil, errors.New("duration must be positive")
	}

	// Check if slug exists for this user
	existing, err := s.repo.GetBySlug(ctx, userID, req.Slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("slug already exists")
	}

	et := &EventType{
		ID:                  uuid.New().String(),
		UserID:              userID,
		Title:               req.Title,
		Slug:                req.Slug,
		Description:         req.Description,
		Duration:            req.Duration,
		BufferBefore:        req.BufferBefore,
		BufferAfter:         req.BufferAfter,
		Color:               req.Color,
		IsActive:            true,
		MinCancelNotice:     req.MinCancelNotice,
		MinRescheduleNotice: req.MinRescheduleNotice,
		AllowGuestCancel:    req.AllowGuestCancel,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := s.repo.Create(ctx, et); err != nil {
		return nil, err
	}

	return et, nil
}

func (s *service) GetEventType(ctx context.Context, id string) (*EventType, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) GetEventTypeBySlug(ctx context.Context, userID, slug string) (*EventType, error) {
	return s.repo.GetBySlug(ctx, userID, slug)
}

func (s *service) ListEventTypes(ctx context.Context, userID string) ([]*EventType, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *service) UpdateEventType(ctx context.Context, id, userID string, req UpdateEventTypeRequest) (*EventType, error) {
	et, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if et == nil {
		return nil, errors.New("event type not found")
	}
	if et.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	if req.Title != nil {
		et.Title = *req.Title
	}
	if req.Slug != nil {
		// If slug is changing, check uniqueness
		if *req.Slug != et.Slug {
			existing, err := s.repo.GetBySlug(ctx, userID, *req.Slug)
			if err != nil {
				return nil, err
			}
			if existing != nil {
				return nil, errors.New("slug already exists")
			}
			et.Slug = *req.Slug
		}
	}
	if req.Description != nil {
		et.Description = *req.Description
	}
	if req.Duration != nil {
		if *req.Duration <= 0 {
			return nil, errors.New("duration must be positive")
		}
		et.Duration = *req.Duration
	}
	if req.BufferBefore != nil {
		et.BufferBefore = *req.BufferBefore
	}
	if req.BufferAfter != nil {
		et.BufferAfter = *req.BufferAfter
	}
	if req.Color != nil {
		et.Color = *req.Color
	}
	if req.IsActive != nil {
		et.IsActive = *req.IsActive
	}
	if req.MinCancelNotice != nil {
		et.MinCancelNotice = *req.MinCancelNotice
	}
	if req.MinRescheduleNotice != nil {
		et.MinRescheduleNotice = *req.MinRescheduleNotice
	}
	if req.AllowGuestCancel != nil {
		et.AllowGuestCancel = *req.AllowGuestCancel
	}
	et.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, et); err != nil {
		return nil, err
	}

	return et, nil
}

func (s *service) DeleteEventType(ctx context.Context, id, userID string) error {
	et, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if et == nil {
		return errors.New("event type not found")
	}
	if et.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.repo.Delete(ctx, id)
}

func (s *service) GetPublicEventTypes(ctx context.Context, hostID string) ([]*EventType, error) {
	all, err := s.repo.ListByUserID(ctx, hostID)
	if err != nil {
		return nil, err
	}

	var active []*EventType
	for _, et := range all {
		if et.IsActive {
			active = append(active, et)
		}
	}
	return active, nil
}
