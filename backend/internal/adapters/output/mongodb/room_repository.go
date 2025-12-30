package mongodb

import (
	"context"

	"github.com/meet-clone/backend/internal/core/domain/room"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RoomRepository struct {
	collection *mongo.Collection
}

func NewRoomRepository(client *Client) room.Repository {
	return &RoomRepository{
		collection: client.GetCollection("rooms"),
	}
}

func (r *RoomRepository) Create(ctx context.Context, room *room.Room) error {
	_, err := r.collection.InsertOne(ctx, room)
	return err
}

func (r *RoomRepository) FindByID(ctx context.Context, id string) (*room.Room, error) {
	var rm room.Room
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&rm)
	if err != nil {
		return nil, err
	}
	return &rm, nil
}

func (r *RoomRepository) Update(ctx context.Context, room *room.Room) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": room.ID},
		bson.M{"$set": room},
	)
	return err
}

func (r *RoomRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *RoomRepository) FindActiveRooms(ctx context.Context, limit, offset int) ([]*room.Room, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, bson.M{"status": room.RoomStatusActive}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rooms []*room.Room
	if err := cursor.All(ctx, &rooms); err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *RoomRepository) FindByCreator(ctx context.Context, createdBy string) ([]*room.Room, error) {
	opts := options.Find().
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, bson.M{
		"created_by": createdBy,
		"status":     room.RoomStatusActive,
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rooms []*room.Room
	if err := cursor.All(ctx, &rooms); err != nil {
		return nil, err
	}

	return rooms, nil
}
