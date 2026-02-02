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

type WaitingParticipant struct {
	UserID      string    `json:"user_id" bson:"user_id"`
	Name        string    `json:"name" bson:"name"`
	Avatar      string    `json:"avatar" bson:"avatar"`
	RequestedAt time.Time `json:"requested_at" bson:"requested_at"`
}

type Room struct {
	ID                  string               `json:"id" bson:"_id"`
	CreatedBy           string               `json:"created_by" bson:"created_by"`
	CloudflareSessionID string               `json:"cloudflare_session_id,omitempty" bson:"cloudflare_session_id,omitempty"`
	Status              RoomStatus           `json:"status" bson:"status"`
	Participants        []Participant        `json:"participants" bson:"participants"`
	WaitingParticipants []WaitingParticipant `json:"waiting_participants" bson:"waiting_participants"`
	MaxCapacity         int                  `json:"max_capacity" bson:"max_capacity"`
	CreatedAt           time.Time            `json:"created_at" bson:"created_at"`
	EndedAt             time.Time            `json:"ended_at,omitempty" bson:"ended_at,omitempty"`
}

func NewRoom(createdBy string, maxCapacity int) *Room {
	return &Room{
		ID:                  uuid.New().String(),
		CreatedBy:           createdBy,
		Status:              RoomStatusActive,
		Participants:        []Participant{},
		WaitingParticipants: []WaitingParticipant{},
		MaxCapacity:         maxCapacity,
		CreatedAt:           time.Now(),
	}
}

func (r *Room) AddParticipant(userID, name, avatar string) error {
	if len(r.GetActiveParticipants()) >= r.MaxCapacity {
		return &RoomError{Message: "room is at maximum capacity"}
	}

	// Check if user is already in the room (actively)
	for i, p := range r.Participants {
		if p.UserID == userID && p.LeftAt.IsZero() {
			// User is trying to rejoin - this is okay, just return success
			// Update their info in case name/avatar changed
			r.Participants[i].Name = name
			r.Participants[i].Avatar = avatar
			return nil
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

func (r *Room) AddToWaitingList(userID, name, avatar string) error {
	// Check if already in waiting list
	for _, wp := range r.WaitingParticipants {
		if wp.UserID == userID {
			return &RoomError{Message: "user already in waiting list"}
		}
	}

	// Check if already a participant
	for _, p := range r.Participants {
		if p.UserID == userID && p.LeftAt.IsZero() {
			return &RoomError{Message: "user already in the room"}
		}
	}

	r.WaitingParticipants = append(r.WaitingParticipants, WaitingParticipant{
		UserID:      userID,
		Name:        name,
		Avatar:      avatar,
		RequestedAt: time.Now(),
	})

	return nil
}

func (r *Room) ApproveParticipant(userID string) error {
	// Find in waiting list
	for i, wp := range r.WaitingParticipants {
		if wp.UserID == userID {
			// Remove from waiting list
			r.WaitingParticipants = append(r.WaitingParticipants[:i], r.WaitingParticipants[i+1:]...)

			// Add to participants
			return r.AddParticipant(userID, wp.Name, wp.Avatar)
		}
	}

	return &RoomError{Message: "user not in waiting list"}
}

func (r *Room) DenyParticipant(userID string) error {
	// Find and remove from waiting list
	for i, wp := range r.WaitingParticipants {
		if wp.UserID == userID {
			r.WaitingParticipants = append(r.WaitingParticipants[:i], r.WaitingParticipants[i+1:]...)
			return nil
		}
	}

	return &RoomError{Message: "user not in waiting list"}
}

func (r *Room) IsUserApproved(userID string) bool {
	// Creator is always approved
	if r.CreatedBy == userID {
		return true
	}

	// Check if user is an active participant
	for _, p := range r.Participants {
		if p.UserID == userID && p.LeftAt.IsZero() {
			return true
		}
	}

	return false
}

type RoomError struct {
	Message string
}

func (e *RoomError) Error() string {
	return e.Message
}
