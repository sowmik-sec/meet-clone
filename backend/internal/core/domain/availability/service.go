package availability

import (
	"context"
)

type Repository interface {
	Save(ctx context.Context, availability *Availability) error
	Get(ctx context.Context, userID string) (*Availability, error)
}

type Service interface {
	SaveAvailability(ctx context.Context, userID string, schedule []DayAvailability, timezone string, isAcceptingBookings bool) (*Availability, error)
	GetAvailability(ctx context.Context, userID string) (*Availability, error)
	// GetPublicAvailability(ctx context.Context, username string) (*Availability, error) // Will need user lookup by username
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) SaveAvailability(ctx context.Context, userID string, schedule []DayAvailability, timezone string, isAcceptingBookings bool) (*Availability, error) {
	avail := &Availability{
		UserID:              userID,
		Schedule:            schedule,
		Timezone:            timezone,
		IsAcceptingBookings: isAcceptingBookings,
	}
	// Validate schedule logic here (end > start, etc)
	if err := s.repo.Save(ctx, avail); err != nil {
		return nil, err
	}
	return avail, nil
}

func (s *service) GetAvailability(ctx context.Context, userID string) (*Availability, error) {
	avail, err := s.repo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	if avail == nil {
		return DefaultAvailability(userID), nil
	}
	return avail, nil
}
