package appointment

import (
	"context"
)

type AppointmentFilter struct {
	StartTimeAfter  *string
	StartTimeBefore *string
	Status          *AppointmentStatus
}

type Repository interface {
	Create(ctx context.Context, appointment *Appointment) error
	FindByID(ctx context.Context, id string) (*Appointment, error)
	FindByUser(ctx context.Context, userID string, filter AppointmentFilter) ([]Appointment, error)
	Update(ctx context.Context, appointment *Appointment) error
	Delete(ctx context.Context, id string) error
	HasConflict(ctx context.Context, userID string, startTime, endTime string) (bool, error)
	GetBookedSlots(ctx context.Context, userID string, date string) ([][]string, error)
	FindByRoomID(ctx context.Context, roomID string) (*Appointment, error)
}
