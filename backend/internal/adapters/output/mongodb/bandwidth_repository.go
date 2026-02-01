package mongodb

import (
	"context"
	"errors"
	"log"

	"github.com/meet-clone/backend/internal/core/domain/bandwidth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BandwidthRepository struct {
	collection *mongo.Collection
}

func NewBandwidthRepository(client *Client) bandwidth.Repository {
	return &BandwidthRepository{
		collection: client.GetCollection("bandwidth_usage"),
	}
}

func (r *BandwidthRepository) Create(ctx context.Context, record *bandwidth.BandwidthRecord) error {
	// Use upsert to handle potential race conditions with unique index
	filter := bson.M{
		"user_id":    record.UserID,
		"room_id":    record.RoomID,
		"session_id": record.SessionID,
	}
	update := bson.M{"$set": record}
	opts := options.Update().SetUpsert(true)

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *BandwidthRepository) Update(ctx context.Context, record *bandwidth.BandwidthRecord) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": record.ID},
		bson.M{"$set": record},
	)
	return err
}

func (r *BandwidthRepository) FindBySession(ctx context.Context, userID, roomID, sessionID string) (*bandwidth.BandwidthRecord, error) {
	log.Printf("[BandwidthRepo] FindBySession: userID=%s, roomID=%s, sessionID=%s", userID, roomID, sessionID)

	var record bandwidth.BandwidthRecord
	err := r.collection.FindOne(ctx, bson.M{
		"user_id":    userID,
		"room_id":    roomID,
		"session_id": sessionID,
	}).Decode(&record)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Printf("[BandwidthRepo] FindBySession: No existing record found for userID=%s", userID)
			return nil, nil
		}
		log.Printf("[BandwidthRepo] FindBySession: Error: %v", err)
		return nil, err
	}

	log.Printf("[BandwidthRepo] FindBySession: Found existing record ID=%s for userID=%s", record.ID, userID)
	return &record, nil
}

func (r *BandwidthRepository) GetUserStats(ctx context.Context, userID string) (*bandwidth.UserBandwidthStats, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"user_id": userID}}},
		{{Key: "$group", Value: bson.M{
			"_id":                  "$user_id",
			"total_bytes_sent":     bson.M{"$sum": "$bytes_sent"},
			"total_bytes_received": bson.M{"$sum": "$bytes_received"},
			"total_duration":       bson.M{"$sum": "$duration"},
			"meeting_count":        bson.M{"$sum": 1},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stats []bandwidth.UserBandwidthStats
	if err := cursor.All(ctx, &stats); err != nil {
		return nil, err
	}

	if len(stats) == 0 {
		log.Printf("[BandwidthRepo] GetUserStats: No stats found for user %s", userID)
		return &bandwidth.UserBandwidthStats{
			UserID: userID,
		}, nil
	}

	log.Printf("[BandwidthRepo] GetUserStats result: %+v", stats[0])
	return &stats[0], nil
}

func (r *BandwidthRepository) GetUserHistory(ctx context.Context, userID string, limit int) ([]*bandwidth.BandwidthRecord, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.M{"created_at": -1})

	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var history []*bandwidth.BandwidthRecord
	if err := cursor.All(ctx, &history); err != nil {
		return nil, err
	}

	return history, nil
}
