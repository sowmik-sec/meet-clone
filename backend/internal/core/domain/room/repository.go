package room

import "context"

type Repository interface {
	Create(ctx context.Context, room *Room) error
	FindByID(ctx context.Context, id string) (*Room, error)
	FindByCreator(ctx context.Context, createdBy string) ([]*Room, error)
	Update(ctx context.Context, room *Room) error
	Delete(ctx context.Context, id string) error
	FindActiveRooms(ctx context.Context, limit, offset int) ([]*Room, error)
}
