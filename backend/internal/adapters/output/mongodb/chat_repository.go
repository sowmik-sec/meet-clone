package mongodb

import (
	"context"

	"github.com/meet-clone/backend/internal/core/domain/chat"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChatRepository struct {
	collection *mongo.Collection
}

func NewChatRepository(client *Client) chat.Repository {
	return &ChatRepository{
		collection: client.GetCollection("chat_messages"),
	}
}

func (r *ChatRepository) Create(ctx context.Context, message *chat.Message) error {
	_, err := r.collection.InsertOne(ctx, message)
	return err
}

func (r *ChatRepository) FindByRoomID(ctx context.Context, roomID string, limit, offset int) ([]*chat.Message, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"timestamp": -1})

	cursor, err := r.collection.Find(ctx, bson.M{"room_id": roomID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*chat.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *ChatRepository) DeleteByRoomID(ctx context.Context, roomID string) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"room_id": roomID})
	return err
}
