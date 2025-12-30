package chat

import "context"

type Repository interface {
	Create(ctx context.Context, message *Message) error
	FindByRoomID(ctx context.Context, roomID string, limit, offset int) ([]*Message, error)
	DeleteByRoomID(ctx context.Context, roomID string) error
}
