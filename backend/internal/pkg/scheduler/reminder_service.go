package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/meet-clone/backend/internal/core/domain/appointment"
	"github.com/meet-clone/backend/internal/core/domain/user" // Assuming user repo needed to get guest/host names/email?

	// Wait, Appointment struct has GuestEmail and GuestName.
	// But we also need Host name/email? Appointment has HostID.
	// Yes, we likely need to fetch host to get their email/name if not denormalized.
	"github.com/meet-clone/backend/internal/pkg/email"
)

type ReminderService struct {
	appointmentRepo appointment.Repository
	userRepo        user.Repository // To fetch host details if needed
	emailService    email.Service
	interval        time.Duration
	frontendURL     string
}

func NewReminderService(appointmentRepo appointment.Repository, userRepo user.Repository, emailService email.Service, interval time.Duration, frontendURL string) *ReminderService {
	return &ReminderService{
		appointmentRepo: appointmentRepo,
		userRepo:        userRepo,
		emailService:    emailService,
		interval:        interval,
		frontendURL:     frontendURL,
	}
}

func (s *ReminderService) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	fmt.Println("Starting reminder scheduler...")
	for {
		select {
		case <-ticker.C:
			s.processReminders(ctx)
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (s *ReminderService) processReminders(ctx context.Context) {
	// Logic: Find appointments starting in 25-35 mins (assuming we run every 5 mins, this covers the 30 min mark with buffer)
	// Or simplify: 30 minutes from now +/- interval.

	now := time.Now()
	// Look for appointments starting between now+25m and now+35m
	// This assumes we run every 5 mins or so.
	startWindow := now.Add(25 * time.Minute)
	endWindow := now.Add(35 * time.Minute)

	appointments, err := s.appointmentRepo.FindUpcoming(ctx, startWindow, endWindow)
	if err != nil {
		log.Printf("Error fetching upcoming appointments: %v\n", err)
		return
	}

	if len(appointments) > 0 {
		log.Printf("Found %d appointments to remind\n", len(appointments))
	}

	for _, appt := range appointments {
		if err := s.sendReminder(ctx, &appt); err != nil {
			log.Printf("Failed to send reminder for appointment %s: %v\n", appt.ID, err)
			continue
		}

		// Update reminder_sent flag
		appt.ReminderSent = true
		if err := s.appointmentRepo.Update(ctx, &appt); err != nil {
			log.Printf("Failed to update reminder flag for %s: %v\n", appt.ID, err)
		}
	}
}

func (s *ReminderService) sendReminder(ctx context.Context, appt *appointment.Appointment) error {
	guestName := appt.Title    // For public bookings, Title is Guest Name
	guestEmail := appt.GuestID // GuestID is email for public bookings

	// If GuestID is a User ID (not an email), fetch user details
	// Simple check for '@'
	isEmail := false
	for _, char := range appt.GuestID {
		if char == '@' {
			isEmail = true
			break
		}
	}

	if !isEmail {
		guestUser, err := s.userRepo.FindByID(ctx, appt.GuestID)
		if err == nil && guestUser != nil {
			guestName = guestUser.Name
			guestEmail = guestUser.Email
		}
	}

	// Fetch Host info
	host, err := s.userRepo.FindByID(ctx, appt.HostID)
	if err != nil {
		return fmt.Errorf("could not find host: %w", err)
	}

	rescheduleLink := fmt.Sprintf("%s/reschedule/%s", s.frontendURL, appt.RescheduleToken)
	meetingLink := ""
	if appt.RoomID != "" {
		meetingLink = fmt.Sprintf("%s/room/%s", s.frontendURL, appt.RoomID)
	} else {
		meetingLink = fmt.Sprintf("%s/room/pending", s.frontendURL)
	}

	// Send to Guest
	if err := s.emailService.SendAppointmentReminder(
		guestEmail,
		guestName,
		appt.Title,
		appt.StartTime.Format(time.RFC1123),
		meetingLink,
		rescheduleLink,
	); err != nil {
		return err // Log or continue? The loop continues on error.
	}

	// Send to Host
	if err := s.emailService.SendAppointmentReminder(
		host.Email,
		host.Name,
		appt.Title,
		appt.StartTime.Format(time.RFC1123),
		meetingLink,
		rescheduleLink,
	); err != nil {
		return err
	}

	return nil
}
