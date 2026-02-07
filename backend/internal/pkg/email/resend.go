package email

import (
	"fmt"

	"github.com/resend/resend-go/v2"
)

type Service interface {
	SendAppointmentConfirmation(to string, userName string, title string, time string, link string, rescheduleLink string) error
	SendAppointmentInvitation(to string, inviterName string, title string, time string, link string) error
	SendAppointmentCancellation(to string, title string, time string) error
	SendAppointmentPending(to string, userName string, title string, time string) error
	SendAppointmentReminder(to string, userName string, title string, time string, link string, rescheduleLink string) error
}

type resendService struct {
	client *resend.Client
	from   string
}

func NewResendService(apiKey string, fromEmail string) Service {
	if apiKey == "" {
		return &noopService{} // Fallback for dev if no key
	}
	client := resend.NewClient(apiKey)
	return &resendService{
		client: client,
		from:   fromEmail,
	}
}

func (s *resendService) SendAppointmentConfirmation(to string, userName string, title string, time string, link string, rescheduleLink string) error {
	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: fmt.Sprintf("Confirmed: %s", title),
		Html: fmt.Sprintf(`
			<h1>Appointment Confirmed</h1>
			<p>Hi %s,</p>
			<p>Your appointment <strong>%s</strong> is confirmed for <strong>%s</strong>.</p>
			<p><a href="%s">Join Meeting</a></p>
			<hr/>
			<p>Need to reschedule? <a href="%s">Click here</a></p>
		`, userName, title, time, link, rescheduleLink),
	}

	_, err := s.client.Emails.Send(params)
	return err
}

// Wait, I can't just change the signature and implementation in one go without breaking callers.
// I will stick to adding the new method `SendAppointmentReminder` first, and then I will do a separate edit to update `SendAppointmentConfirmation` signature and its callers.
// Let's just add `SendAppointmentReminder` for now to be safe and atomic.

func (s *resendService) SendAppointmentInvitation(to string, inviterName string, title string, time string, link string) error {
	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: fmt.Sprintf("Invitation: %s", title),
		Html: fmt.Sprintf(`
			<h1>New Invitation</h1>
			<p>%s has invited you to <strong>%s</strong>.</p>
			<p>Time: <strong>%s</strong></p>
			<p><a href="%s">Join Meeting</a></p>
		`, inviterName, title, time, link),
	}

	_, err := s.client.Emails.Send(params)
	return err
}

func (s *resendService) SendAppointmentCancellation(to string, title string, time string) error {
	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: fmt.Sprintf("Cancelled: %s", title),
		Html: fmt.Sprintf(`
			<h1>Appointment Cancelled</h1>
			<p>The appointment <strong>%s</strong> scheduled for <strong>%s</strong> has been cancelled.</p>
		`, title, time),
	}

	_, err := s.client.Emails.Send(params)
	return err
}

func (s *resendService) SendAppointmentPending(to string, userName string, title string, time string) error {
	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: fmt.Sprintf("Pending Approval: %s", title),
		Html: fmt.Sprintf(`
			<h1>Appointment Pending</h1>
			<p>Hi %s,</p>
			<p>Your appointment request for <strong>%s</strong> at <strong>%s</strong> has been received and is pending approval.</p>
			<p>You will receive another email once the host confirms the appointment.</p>
		`, userName, title, time),
	}

	_, err := s.client.Emails.Send(params)
	return err
}

func (s *resendService) SendAppointmentReminder(to, userName, title, time, link, rescheduleLink string) error {
	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: fmt.Sprintf("Reminder: %s", title),
		Html: fmt.Sprintf(`
			<h1>Appointment Reminder</h1>
			<p>Hi %s,</p>
			<p>This is a reminder that you have an appointment <strong>%s</strong> starting in 30 minutes at <strong>%s</strong>.</p>
			<p><a href="%s">Join Meeting</a></p>
			<hr/>
			<p>Need to reschedule? <a href="%s">Click here</a></p>
		`, userName, title, time, link, rescheduleLink),
	}

	_, err := s.client.Emails.Send(params)
	return err
}

// Noop service for local dev without keys
type noopService struct{}

func (s *noopService) SendAppointmentConfirmation(to, name, title, time, link, rescheduleLink string) error {
	fmt.Printf("[Email Mock] Confirmation sent to %s for %s. Reschedule: %s\n", to, title, rescheduleLink)
	return nil
}
func (s *noopService) SendAppointmentInvitation(to, name, title, time, link string) error {
	fmt.Printf("[Email Mock] Invitation sent to %s for %s\n", to, title)
	return nil
}
func (s *noopService) SendAppointmentCancellation(to, title, time string) error {
	fmt.Printf("[Email Mock] Cancellation sent to %s for %s\n", to, title)
	return nil
}
func (s *noopService) SendAppointmentPending(to, name, title, time string) error {
	fmt.Printf("[Email Mock] Pending request sent to %s for %s\n", to, title)
	return nil
}
func (s *noopService) SendAppointmentReminder(to, name, title, time, link, rescheduleLink string) error {
	fmt.Printf("[Email Mock] Reminder sent to %s for %s. Reschedule: %s\n", to, title, rescheduleLink)
	return nil
}
