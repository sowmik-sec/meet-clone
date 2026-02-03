package mongodb

import (
	"context"

	"github.com/meet-clone/backend/internal/core/domain/availability"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type availabilityRepository struct {
	collection *mongo.Collection
}

func NewAvailabilityRepository(db *mongo.Database) availability.Repository {
	return &availabilityRepository{
		collection: db.Collection("availability"),
	}
}

func (r *availabilityRepository) Save(ctx context.Context, avail *availability.Availability) error {
	opts := options.Replace().SetUpsert(true)
	_, err := r.collection.ReplaceOne(ctx, bson.M{"user_id": avail.UserID}, avail, opts)
	return err
}

func (r *availabilityRepository) Get(ctx context.Context, userID string) (*availability.Availability, error) {
	var avail availability.Availability
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&avail)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &avail, nil
}
