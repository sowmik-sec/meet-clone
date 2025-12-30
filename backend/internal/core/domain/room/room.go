package room

import (
	"time"

	"github.com/google/uuid"
)

type RoomStatus string

const (
	RoomStatusActive RoomStatus = "active"
	RoomStatusEnded  RoomStatus = "ended"
)

type Participant struct {
	UserID   string    `json:"user_id" bson:"user_id"`
	Name     string    `json:"name" bson:"name"`
	Avatar   string    `json:"avatar" bson:"avatar"`
	JoinedAt time.Time `json:"joined_at" bson:"joined_at"`
	LeftAt   time.Time `json:"left_at,omitempty" bson:"left_at,omitempty"`
}

type Room struct {
	ID                  string        `json:"id" bson:"_id"`
	CreatedBy           string        `json:"created_by" bson:"created_by"`
	CloudflareSessionID string        `json:"cloudflare_session_id,omitempty" bson:"cloudflare_session_id,omitempty"`
	Status              RoomStatus    `json:"status" bson:"status"`
	Participants        []Participant `json:"participants" bson:"participants"`
	MaxCapacity         int           `json:"max_capacity" bson:"max_capacity"`
	CreatedAt           time.Time     `json:"created_at" bson:"created_at"`
	EndedAt             time.Time     `json:"ended_at,omitempty" bson:"ended_at,omitempty"`
}

func NewRoom(createdBy string, maxCapacity int) *Room {
	return &Room{
		ID:           uuid.New().String(),
		CreatedBy:    createdBy,
		Status:       RoomStatusActive,
		Participants: []Participant{},
		MaxCapacity:  maxCapacity,
		CreatedAt:    time.Now(),
	}
}

func (r *Room) AddParticipant(userID, name, avatar string) error {
	if len(r.Participants) >= r.MaxCapacity {
		return &RoomError{Message: "room is at maximum capacity"}
	}

	// Check if user is already in the room
	for _, p := range r.Participants {
		if p.UserID == userID && p.LeftAt.IsZero() {
			return &RoomError{Message: "user is already in the room"}
		}
	}

	r.Participants = append(r.Participants, Participant{
		UserID:   userID,
		Name:     name,
		Avatar:   avatar,
		JoinedAt: time.Now(),
	})

	return nil
}

func (r *Room) RemoveParticipant(userID string) error {
	for i, p := range r.Participants {
		if p.UserID == userID && p.LeftAt.IsZero() {
			r.Participants[i].LeftAt = time.Now()
			return nil
		}
	}
	return &RoomError{Message: "participant not found in room"}
}

func (r *Room) GetActiveParticipants() []Participant {
	active := []Participant{}
	for _, p := range r.Participants {
		if p.LeftAt.IsZero() {
			active = append(active, p)
		}
	}
	return active
}

func (r *Room) End() {
	r.Status = RoomStatusEnded
	r.EndedAt = time.Now()
}

func (r *Room) IsActive() bool {
	return r.Status == RoomStatusActive
}

type RoomError struct {
	Message string
}

func (e *RoomError) Error() string {
	return e.Message
}
