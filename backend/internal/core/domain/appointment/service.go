package appointment

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/meet-clone/backend/internal/core/domain/room"
	"github.com/meet-clone/backend/internal/pkg/email"
)

// Request/Response structs

type CreateAppointmentRequest struct {
	GuestID     string    `json:"guest_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Timezone    string    `json:"timezone"`
	MeetingType string    `json:"meeting_type"` // "meeting" or "webinar"
}

type CreatePublicBookingRequest struct {
	GuestName  string    `json:"guest_name"`
	GuestEmail string    `json:"guest_email"`
	StartTime  time.Time `json:"start_time"`
	Timezone   string    `json:"timezone"`
}

type UpdateAppointmentRequest struct {
	Title       *string
	Description *string
	StartTime   *time.Time
	EndTime     *time.Time
	Status      *AppointmentStatus
}

// Service Interface

type Service interface {
	CreateAppointment(ctx context.Context, hostID string, req CreateAppointmentRequest) (*Appointment, error)
	GetAppointment(ctx context.Context, id string) (*Appointment, error)
	GetUserAppointments(ctx context.Context, userID string, filter AppointmentFilter) ([]Appointment, error)
	UpdateAppointment(ctx context.Context, id string, req UpdateAppointmentRequest) (*Appointment, error)
	CancelAppointment(ctx context.Context, id, userID string) error
	ConfirmAppointment(ctx context.Context, id, hostID string) error
	StartAppointment(ctx context.Context, id, userID string) (string, error) // Returns room ID
	CreatePublicBooking(ctx context.Context, hostID string, req CreatePublicBookingRequest) (*Appointment, error)
}

// Service Implementation

type service struct {
	repo         Repository
	roomService  room.Service
	emailService email.Service
}

func NewService(repo Repository, roomService room.Service, emailService email.Service) Service {
	return &service{
		repo:         repo,
		roomService:  roomService,
		emailService: emailService,
	}
}

func (s *service) CreateAppointment(ctx context.Context, hostID string, req CreateAppointmentRequest) (*Appointment, error) {
	if req.StartTime.Before(time.Now()) {
		return nil, errors.New("start time must be in the future")
	}
	if req.EndTime.Before(req.StartTime) {
		return nil, errors.New("end time must be after start time")
	}

	appt := NewAppointment(
		hostID,
		req.GuestID,
		req.Title,
		req.Description,
		req.Timezone,
		req.MeetingType,
		req.StartTime,
		req.EndTime,
	)
	appt.ID = uuid.New().String()
	appt.Status = StatusConfirmed // Auto-confirm for now since host is creating it

	if err := s.repo.Create(ctx, appt); err != nil {
		return nil, err
	}

	// Send Invitation Email if guest ID is provided (assuming guestId is email for now or we look it up)
	// For MVP Phase 3, we assume guestID might be user ID. If we want to send email, we need email address.
	// Since Appointment struct has GuestID, we'd need to fetch User to get email.
	// For simplicity, let's assume we can notify "guest" if we find them.
	// This requires User Service to lookup email. Avoiding circular dependency?
	// AppointmentService depends on UserService? Or we passed email in request?
	// Let's assume for now we don't send invitation on create until confirmed or we just skip if we can't find email.

	return appt, nil
}

func (s *service) GetAppointment(ctx context.Context, id string) (*Appointment, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *service) GetUserAppointments(ctx context.Context, userID string, filter AppointmentFilter) ([]Appointment, error) {
	return s.repo.FindByUser(ctx, userID, filter)
}

func (s *service) UpdateAppointment(ctx context.Context, id string, req UpdateAppointmentRequest) (*Appointment, error) {
	appt, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if appt == nil {
		return nil, errors.New("appointment not found")
	}

	if req.Title != nil {
		appt.Title = *req.Title
	}
	if req.Description != nil {
		appt.Description = *req.Description
	}
	if req.StartTime != nil {
		appt.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		appt.EndTime = *req.EndTime
	}
	if req.Status != nil {
		appt.Status = *req.Status
	}
	appt.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, appt); err != nil {
		return nil, err
	}

	return appt, nil
}

func (s *service) CancelAppointment(ctx context.Context, id, userID string) error {
	appt, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if appt == nil {
		return errors.New("appointment not found")
	}

	// Only host or guest can cancel
	if appt.HostID != userID && appt.GuestID != userID {
		return errors.New("unauthorized to cancel appointment")
	}

	appt.Status = StatusCancelled
	appt.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, appt); err != nil {
		return err
	}

	// Send Cancellation Email
	if strings.Contains(appt.GuestID, "@") {
		s.emailService.SendAppointmentCancellation(appt.GuestID, appt.Title, appt.StartTime.String())
	}
	return nil
}

func (s *service) ConfirmAppointment(ctx context.Context, id, hostID string) error {
	appt, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if appt == nil {
		return errors.New("appointment not found")
	}

	if appt.HostID != hostID {
		return errors.New("only host can confirm appointment")
	}

	appt.Status = StatusConfirmed
	appt.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, appt); err != nil {
		return err
	}

	// Send Confirmation Email
	if strings.Contains(appt.GuestID, "@") {
		// For public bookings, GuestName is stored in Title usually or we just use "Guest"
		// Using GuestID as name backup if we don't have separate name field in DB yet for generic appointments
		// In CreatePublicBooking we stored GuestName in appt.Title? No, Title is "Public Booking..."
		// Wait, CreatePublicBooking:
		// req.GuestName, // Title is guest name
		// So appt.Title IS the Guest Name for public bookings.
		s.emailService.SendAppointmentConfirmation(appt.GuestID, appt.Title, appt.Title, appt.StartTime.String(), "http://localhost:3000/appointments/"+appt.ID)
	}
	return nil
}

func (s *service) StartAppointment(ctx context.Context, id, userID string) (string, error) {
	appt, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	if appt == nil {
		return "", errors.New("appointment not found")
	}

	// Only host can start
	if appt.HostID != userID {
		return "", errors.New("only host can start the meeting")
	}

	// Create room if not exists
	if appt.RoomID == "" {
		roomType := room.RoomTypeMeeting
		if appt.MeetingType == "webinar" {
			roomType = room.RoomTypeWebinar
		}

		newRoom, err := s.roomService.CreateRoom(ctx, userID, roomType)
		if err != nil {
			return "", err
		}
		appt.RoomID = newRoom.ID
		// Should we update status? Maybe.

		if err := s.repo.Update(ctx, appt); err != nil {
			return "", err
		}
	}

	return appt.RoomID, nil
}

func (s *service) CreatePublicBooking(ctx context.Context, hostID string, req CreatePublicBookingRequest) (*Appointment, error) {
	if req.StartTime.Before(time.Now()) {
		return nil, errors.New("start time must be in the future")
	}

	// Default duration 30 mins for now if not specified (or derived from slots in future)
	// For MVP simplicity: 60 mins default
	endTime := req.StartTime.Add(60 * time.Minute)

	// Check for conflicts
	hasConflict, err := s.repo.HasConflict(ctx, hostID, req.StartTime.Format(time.RFC3339), endTime.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	if hasConflict {
		return nil, errors.New("selected time slot is no longer available")
	}

	appt := NewAppointment(
		hostID,
		req.GuestEmail,                // Using email as ID for guest since they might not be reg user
		req.GuestName,                 // Title is guest name
		"Public Booking via Calendar", // Description
		req.Timezone,
		"meeting",
		req.StartTime,
		endTime,
	)
	appt.ID = uuid.New().String()
	appt.Status = StatusConfirmed // Auto-confirm public bookings for MVP

	if err := s.repo.Create(ctx, appt); err != nil {
		return nil, err
	}

	// Send Email Notifications (Host + Guest)
	s.emailService.SendAppointmentConfirmation(req.GuestEmail, req.GuestName, appt.Title, appt.StartTime.String(), "http://localhost:3000/appointments/"+appt.ID)

	return appt, nil
}
