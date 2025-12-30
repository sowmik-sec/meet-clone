package chat

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        string    `json:"id" bson:"_id"`
	RoomID    string    `json:"room_id" bson:"room_id"`
	UserID    string    `json:"user_id" bson:"user_id"`
	UserName  string    `json:"user_name" bson:"user_name"`
	Message   string    `json:"message" bson:"message"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

func NewMessage(roomID, userID, userName, message string) *Message {
	return &Message{
		ID:        uuid.New().String(),
		RoomID:    roomID,
		UserID:    userID,
		UserName:  userName,
		Message:   message,
		Timestamp: time.Now(),
	}
}
