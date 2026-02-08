package appointment

import (
	"context"
	"time"
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
	FindAppointmentsByDate(ctx context.Context, userID string, date string) ([]Appointment, error)
	FindUpcoming(ctx context.Context, start, end time.Time) ([]Appointment, error)
	FindByRoomID(ctx context.Context, roomID string) (*Appointment, error)
	FindByRescheduleToken(ctx context.Context, token string) (*Appointment, error)
	CountBookingsForSlot(ctx context.Context, hostID string, startTime, endTime time.Time) (int, error)
	HasBookingForSlot(ctx context.Context, hostID, guestEmail string, startTime, endTime time.Time) (bool, error)
}
