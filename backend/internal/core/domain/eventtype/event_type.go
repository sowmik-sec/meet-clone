package eventtype

import (
	"context"
	"time"
)

type EventType struct {
	ID                  string `json:"id" bson:"_id,omitempty"`
	UserID              string `json:"user_id" bson:"user_id"`
	Title               string `json:"title" bson:"title"`
	Slug                string `json:"slug" bson:"slug"`
	Description         string `json:"description" bson:"description"`
	Duration            int    `json:"duration" bson:"duration"`           // in minutes
	BufferBefore        int    `json:"buffer_before" bson:"buffer_before"` // in minutes
	BufferAfter         int    `json:"buffer_after" bson:"buffer_after"`   // in minutes
	Color               string `json:"color" bson:"color"`                 // hex color
	IsActive            bool   `json:"is_active" bson:"is_active"`
	MinCancelNotice     int    `json:"min_cancel_notice" bson:"min_cancel_notice"`         // hours
	MinRescheduleNotice int    `json:"min_reschedule_notice" bson:"min_reschedule_notice"` // hours
	AllowGuestCancel    bool   `json:"allow_guest_cancel" bson:"allow_guest_cancel"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type CreateEventTypeRequest struct {
	Title               string `json:"title"`
	Slug                string `json:"slug"`
	Description         string `json:"description"`
	Duration            int    `json:"duration"`
	BufferBefore        int    `json:"buffer_before"`
	BufferAfter         int    `json:"buffer_after"`
	Color               string `json:"color"`
	MinCancelNotice     int    `json:"min_cancel_notice"`
	MinRescheduleNotice int    `json:"min_reschedule_notice"`
	AllowGuestCancel    bool   `json:"allow_guest_cancel"`
}

type UpdateEventTypeRequest struct {
	Title               *string `json:"title"`
	Slug                *string `json:"slug"`
	Description         *string `json:"description"`
	Duration            *int    `json:"duration"`
	BufferBefore        *int    `json:"buffer_before"`
	BufferAfter         *int    `json:"buffer_after"`
	Color               *string `json:"color"`
	IsActive            *bool   `json:"is_active"`
	MinCancelNotice     *int    `json:"min_cancel_notice"`
	MinRescheduleNotice *int    `json:"min_reschedule_notice"`
	AllowGuestCancel    *bool   `json:"allow_guest_cancel"`
}

type Repository interface {
	Create(ctx context.Context, eventType *EventType) error
	GetByID(ctx context.Context, id string) (*EventType, error)
	GetBySlug(ctx context.Context, userID, slug string) (*EventType, error)
	ListByUserID(ctx context.Context, userID string) ([]*EventType, error)
	Update(ctx context.Context, eventType *EventType) error
	Delete(ctx context.Context, id string) error
}

type Service interface {
	CreateEventType(ctx context.Context, userID string, req CreateEventTypeRequest) (*EventType, error)
	GetEventType(ctx context.Context, id string) (*EventType, error)
	GetEventTypeBySlug(ctx context.Context, userID, slug string) (*EventType, error)
	ListEventTypes(ctx context.Context, userID string) ([]*EventType, error)
	UpdateEventType(ctx context.Context, id, userID string, req UpdateEventTypeRequest) (*EventType, error)
	DeleteEventType(ctx context.Context, id, userID string) error
	GetPublicEventTypes(ctx context.Context, hostID string) ([]*EventType, error)
}
