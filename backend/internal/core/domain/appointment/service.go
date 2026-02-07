package appointment

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/meet-clone/backend/internal/core/domain/eventtype"
	"github.com/meet-clone/backend/internal/core/domain/room"
	"github.com/meet-clone/backend/internal/core/domain/user"
	"github.com/meet-clone/backend/internal/pkg/calendar"
	"github.com/meet-clone/backend/internal/pkg/email"
	"github.com/meet-clone/backend/internal/pkg/logger"
	"golang.org/x/oauth2"
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
	GuestName   string    `json:"guest_name"`
	GuestEmail  string    `json:"guest_email"`
	StartTime   time.Time `json:"start_time"`
	Timezone    string    `json:"timezone"`
	EventTypeID string    `json:"event_type_id"`
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
	GetBookedSlots(ctx context.Context, userID string, date string) ([][]string, error)
	GetAppointmentByRoomID(ctx context.Context, roomID string) (*Appointment, error)
}

// Service Implementation

type service struct {
	repo          Repository
	roomService   room.Service
	emailService  email.Service
	eventTypeRepo eventtype.Repository
	userRepo      user.Repository
	calendar      calendar.Service
}

func NewService(repo Repository, roomService room.Service, emailService email.Service, eventTypeRepo eventtype.Repository, userRepo user.Repository, calendar calendar.Service) Service {
	return &service{
		repo:          repo,
		roomService:   roomService,
		emailService:  emailService,
		eventTypeRepo: eventTypeRepo,
		userRepo:      userRepo,
		calendar:      calendar,
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
	appt.BufferedStartTime = appt.StartTime // No buffer for manual appt for now?
	appt.BufferedEndTime = appt.EndTime
	appt.ID = uuid.New().String()
	appt.Status = StatusConfirmed // Auto-confirm for now since host is creating it

	if err := s.repo.Create(ctx, appt); err != nil {
		return nil, err
	}

	// Try to sync with Google Calendar
	s.syncToGoogleCalendar(ctx, appt)

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

	// Sync with Google Calendar
	s.syncToGoogleCalendar(ctx, appt)

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

	// Check time window (allow start 15 mins before to end time)
	now := time.Now()
	windowStart := appt.StartTime.Add(-15 * time.Minute)

	if now.Before(windowStart) {
		return "", errors.New("meeting cannot be started before scheduled time")
	}
	if now.After(appt.EndTime) {
		return "", errors.New("meeting has already ended")
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

	// Default duration 30 mins
	duration := 30 * time.Minute
	var bufferBefore, bufferAfter int

	// If EventTypeID is provided, fetch duration
	if req.EventTypeID != "" {
		et, err := s.eventTypeRepo.GetByID(ctx, req.EventTypeID)
		if err == nil && et != nil {
			duration = time.Duration(et.Duration) * time.Minute
			bufferBefore = et.BufferBefore
			bufferAfter = et.BufferAfter
		}
	}

	endTime := req.StartTime.Add(duration)

	// Check for conflicts
	// Use Buffered times for conflict check
	checkStart := req.StartTime.Add(-time.Duration(bufferBefore) * time.Minute)
	checkEnd := endTime.Add(time.Duration(bufferAfter) * time.Minute)

	hasConflict, err := s.repo.HasConflict(ctx, hostID, checkStart.Format(time.RFC3339), checkEnd.Format(time.RFC3339))
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
	appt.EventTypeID = req.EventTypeID
	appt.BufferBefore = bufferBefore
	appt.BufferAfter = bufferAfter
	appt.BufferedStartTime = appt.StartTime.Add(-time.Duration(bufferBefore) * time.Minute)
	appt.BufferedEndTime = appt.EndTime.Add(time.Duration(bufferAfter) * time.Minute)
	appt.Status = StatusPending // Require host approval

	if err := s.repo.Create(ctx, appt); err != nil {
		return nil, err
	}

	// Notify Guest
	s.emailService.SendAppointmentPending(req.GuestEmail, req.GuestName, appt.Title, appt.StartTime.Format("Jan 2, 2006 at 3:04 PM"))

	// Notify Host
	hostUser, err := s.userRepo.FindByID(ctx, hostID)
	if err == nil && hostUser != nil {
		// Construct a link to the dashboard or appointment details
		// For now, just a generic link or maybe appointment specific if we have a UI for it
		link := "http://localhost:3000/dashboard"
		s.emailService.SendAppointmentInvitation(hostUser.Email, req.GuestName, appt.Title, appt.StartTime.Format("Jan 2, 2006 at 3:04 PM"), link)
	}

	return appt, nil
}

func (s *service) syncToGoogleCalendar(ctx context.Context, appt *Appointment) {
	host, err := s.userRepo.FindByID(ctx, appt.HostID)
	if err == nil && host != nil && host.GoogleAccessToken != "" {
		token := &oauth2.Token{
			AccessToken:  host.GoogleAccessToken,
			RefreshToken: host.GoogleRefreshToken,
			Expiry:       host.GoogleTokenExpiry,
			TokenType:    "Bearer",
		}

		event := calendar.Event{
			Summary:     appt.Title, // For public booking, Title is Guest Name
			Description: appt.Description,
			Start:       appt.StartTime,
			End:         appt.EndTime,
			Attendees:   []string{},
		}

		// If guest ID is email, add it
		if strings.Contains(appt.GuestID, "@") {
			event.Attendees = append(event.Attendees, appt.GuestID)
		} else {
			// Try to fetch guest user
			guest, err := s.userRepo.FindByID(ctx, appt.GuestID)
			if err == nil && guest != nil {
				event.Attendees = append(event.Attendees, guest.Email)
			}
		}

		_, err := s.calendar.CreateEvent(ctx, token, event)
		if err != nil {
			// logger.Error("Failed to create Google Calendar event: %v", err)
		}
	}
}

func (s *service) GetBookedSlots(ctx context.Context, userID string, date string) ([][]string, error) {
	// 1. Get internal booked slots
	internalSlots, err := s.repo.GetBookedSlots(ctx, userID, date)
	if err != nil {
		return nil, err
	}

	// 2. Get Google Calendar busy times
	user, err := s.userRepo.FindByID(ctx, userID)
	if err == nil && user != nil && user.GoogleAccessToken != "" {
		// Use Google Calendar API
		token := &oauth2.Token{
			AccessToken:  user.GoogleAccessToken,
			RefreshToken: user.GoogleRefreshToken,
			Expiry:       user.GoogleTokenExpiry,
			TokenType:    "Bearer",
		}

		startOfDay, _ := time.Parse("2006-01-02", date)
		endOfDay := startOfDay.Add(24 * time.Hour)

		busyTimes, err := s.calendar.GetBusyTimes(ctx, token, startOfDay, endOfDay)
		if err == nil {
			for _, busy := range busyTimes {
				internalSlots = append(internalSlots, []string{
					busy.Start.Format(time.RFC3339),
					busy.End.Format(time.RFC3339),
				})
			}
		} else {
			logger.Error.Printf("Failed to fetch Google Calendar busy times: %v", err)
		}
	}

	return internalSlots, nil
}

func (s *service) GetAppointmentByRoomID(ctx context.Context, roomID string) (*Appointment, error) {
	return s.repo.FindByRoomID(ctx, roomID)
}
