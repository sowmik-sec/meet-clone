package appointment

import (
	"time"
)

type AppointmentStatus string

const (
	StatusPending   AppointmentStatus = "pending"
	StatusConfirmed AppointmentStatus = "confirmed"
	StatusCancelled AppointmentStatus = "cancelled"
	StatusCompleted AppointmentStatus = "completed"
)

type Appointment struct {
	ID                string            `json:"id" bson:"_id"`
	HostID            string            `json:"host_id" bson:"host_id"`
	GuestID           string            `json:"guest_id" bson:"guest_id"`
	EventTypeID       string            `json:"event_type_id,omitempty" bson:"event_type_id,omitempty"`
	RoomID            string            `json:"room_id,omitempty" bson:"room_id,omitempty"`
	Title             string            `json:"title" bson:"title"`
	Description       string            `json:"description,omitempty" bson:"description,omitempty"`
	StartTime         time.Time         `json:"start_time" bson:"start_time"`
	EndTime           time.Time         `json:"end_time" bson:"end_time"`
	BufferedStartTime time.Time         `json:"buffered_start_time" bson:"buffered_start_time"`
	BufferedEndTime   time.Time         `json:"buffered_end_time" bson:"buffered_end_time"`
	BufferBefore      int               `json:"buffer_before,omitempty" bson:"buffer_before,omitempty"` // minutes
	BufferAfter       int               `json:"buffer_after,omitempty" bson:"buffer_after,omitempty"`   // minutes
	Timezone          string            `json:"timezone" bson:"timezone"`
	Status            AppointmentStatus `json:"status" bson:"status"`
	MeetingType       string            `json:"meeting_type" bson:"meeting_type"`           // "meeting" or "webinar"
	Answers           map[string]string `json:"answers,omitempty" bson:"answers,omitempty"` // question_id -> answer
	RescheduleToken   string            `json:"reschedule_token" bson:"reschedule_token"`
	RescheduleCount   int               `json:"reschedule_count" bson:"reschedule_count"`
	ReminderSent      bool              `json:"reminder_sent" bson:"reminder_sent"`
	GoogleEventID     string            `json:"google_event_id,omitempty" bson:"google_event_id,omitempty"`
	CreatedAt         time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" bson:"updated_at"`
}

func NewAppointment(hostID, guestID, title, description, timezone, meetingType string, startTime, endTime time.Time) *Appointment {
	return &Appointment{
		ID:          "", // Will be set by Repo or UUID
		HostID:      hostID,
		GuestID:     guestID,
		Title:       title,
		Description: description,
		StartTime:   startTime,
		EndTime:     endTime,
		Timezone:    timezone,
		Status:      StatusPending,
		MeetingType: meetingType,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
