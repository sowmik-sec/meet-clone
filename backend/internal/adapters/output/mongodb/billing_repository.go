package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/meet-clone/backend/internal/core/domain/billing"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BillingRepository struct {
	db           *mongo.Database
	sessionsColl *mongo.Collection
	periodsColl  *mongo.Collection
	syncColl     *mongo.Collection
}

func NewBillingRepository(db *mongo.Database) *BillingRepository {
	return &BillingRepository{
		db:           db,
		sessionsColl: db.Collection("meeting_sessions"),
		periodsColl:  db.Collection("user_billing_periods"),
		syncColl:     db.Collection("cloudflare_usage_sync"),
	}
}

// CreateSession inserts a new meeting session
func (r *BillingRepository) CreateSession(ctx context.Context, session *billing.MeetingSession) error {
	_, err := r.sessionsColl.InsertOne(ctx, session)
	return err
}

// UpdateSession updates an existing session
func (r *BillingRepository) UpdateSession(ctx context.Context, session *billing.MeetingSession) error {
	filter := bson.M{"_id": session.ID}
	update := bson.M{"$set": session}
	_, err := r.sessionsColl.UpdateOne(ctx, filter, update)
	return err
}

// FindActiveSession finds a session where left_at is nil
func (r *BillingRepository) FindActiveSession(ctx context.Context, userID, roomID string) (*billing.MeetingSession, error) {
	filter := bson.M{
		"user_id": userID,
		"room_id": roomID,
		"left_at": bson.M{"$exists": false}, // or null check depending on how it's stored, usually omission or null
	}

	// Double check for nil explicitly if needed, but struct field is pointer
	// In BSON, nil pointer field usually is stored as null or omitted if omitempty.
	// Our model has `bson:"left_at,omitempty"`, so it's omitted.

	var session billing.MeetingSession
	// Sort by joined_at desc to get latest
	opts := options.FindOne().SetSort(bson.M{"joined_at": -1})
	err := r.sessionsColl.FindOne(ctx, filter, opts).Decode(&session)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

// EndSession sets left_at and participant_minutes for a session
func (r *BillingRepository) EndSession(ctx context.Context, userID, roomID string, leftAt time.Time, durationMinutes int64) error {
	// We need to find the active session to end
	filter := bson.M{
		"user_id": userID,
		"room_id": roomID,
		"left_at": bson.M{"$exists": false},
	}

	update := bson.M{
		"$set": bson.M{
			"left_at":             leftAt,
			"participant_minutes": durationMinutes,
		},
	}

	_, err := r.sessionsColl.UpdateOne(ctx, filter, update)
	return err
}

// GetOrCreateBillingPeriod retrieves or creates a billing period record
func (r *BillingRepository) GetOrCreateBillingPeriod(ctx context.Context, userID, period string) (*billing.UserBillingPeriod, error) {
	filter := bson.M{
		"user_id":        userID,
		"billing_period": period,
	}

	var userPeriod billing.UserBillingPeriod
	err := r.periodsColl.FindOne(ctx, filter).Decode(&userPeriod)
	if err == nil {
		return &userPeriod, nil
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}

	// Create new
	now := time.Now()
	newPeriod := &billing.UserBillingPeriod{
		ID:            userID + "-" + period, // Simple deterministic ID
		UserID:        userID,
		BillingPeriod: period,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	_, err = r.periodsColl.InsertOne(ctx, newPeriod)
	if err != nil {
		// Handle race condition where it might have been created
		if mongo.IsDuplicateKeyError(err) {
			return r.GetBillingPeriod(ctx, userID, period)
		}
		return nil, err
	}

	return newPeriod, nil
}

// UpdateBillingPeriod updates a billing period record
func (r *BillingRepository) UpdateBillingPeriod(ctx context.Context, period *billing.UserBillingPeriod) error {
	filter := bson.M{"_id": period.ID}
	update := bson.M{"$set": period}
	_, err := r.periodsColl.UpdateOne(ctx, filter, update)
	return err
}

// GetUserBillingHistory retrieves history sorted by period desc
func (r *BillingRepository) GetUserBillingHistory(ctx context.Context, userID string, limit int) ([]*billing.UserBillingPeriod, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().SetSort(bson.M{"billing_period": -1}).SetLimit(int64(limit))

	cursor, err := r.periodsColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var periods []*billing.UserBillingPeriod
	if err := cursor.All(ctx, &periods); err != nil {
		return nil, err
	}
	return periods, nil
}

// GetBillingPeriod gets a specific period
func (r *BillingRepository) GetBillingPeriod(ctx context.Context, userID, period string) (*billing.UserBillingPeriod, error) {
	filter := bson.M{
		"user_id":        userID,
		"billing_period": period,
	}

	var userPeriod billing.UserBillingPeriod
	err := r.periodsColl.FindOne(ctx, filter).Decode(&userPeriod)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &userPeriod, nil
}

// AggregateUserMinutes sums up participant_minutes for a user in a period
func (r *BillingRepository) AggregateUserMinutes(ctx context.Context, userID, period string) (int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"room_creator_id": userID,
			"billing_period":  period,
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":           nil,
			"total_minutes": bson.M{"$sum": "$participant_minutes"},
		}}},
	}

	cursor, err := r.sessionsColl.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}

	// Extract total_minutes safely
	total, ok := result[0]["total_minutes"].(int64) // Depends on driver decoding, assumes int64 usually for $sum of int64
	if !ok {
		// Fallback for different number types
		if i, ok := result[0]["total_minutes"].(int32); ok {
			return int64(i), nil
		}
		if f, ok := result[0]["total_minutes"].(float64); ok {
			return int64(f), nil
		}
		return 0, nil
	}

	return total, nil
}

// AggregateMeetingCounts counts total and hosted meetings for a user in a period
func (r *BillingRepository) AggregateMeetingCounts(ctx context.Context, userID, period string) (totalMeetings, hostedMeetings int, err error) {
	// Count distinct rooms where this user is the room_creator (they hosted)
	// Total meetings = distinct room_ids in sessions where room_creator_id = userID
	// Hosted meetings = distinct room_ids where is_host = true

	// Pipeline for total meetings (distinct rooms billed to this user)
	totalPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"room_creator_id": userID,
			"billing_period":  period,
			"left_at":         bson.M{"$exists": true}, // Only count completed sessions
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": "$room_id", // Group by room to count distinct meetings
		}}},
		{{Key: "$count", Value: "total"}},
	}

	cursor, err := r.sessionsColl.Aggregate(ctx, totalPipeline)
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(ctx)

	var totalResult []bson.M
	if err := cursor.All(ctx, &totalResult); err != nil {
		return 0, 0, err
	}

	if len(totalResult) > 0 {
		if t, ok := totalResult[0]["total"].(int32); ok {
			totalMeetings = int(t)
		}
	}

	// Pipeline for hosted meetings (where the user was the host in the session)
	hostedPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"room_creator_id": userID,
			"user_id":         userID, // They were the actual host (their own session)
			"billing_period":  period,
			"is_host":         true,
			"left_at":         bson.M{"$exists": true},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id": "$room_id",
		}}},
		{{Key: "$count", Value: "total"}},
	}

	hostedCursor, err := r.sessionsColl.Aggregate(ctx, hostedPipeline)
	if err != nil {
		return totalMeetings, 0, err
	}
	defer hostedCursor.Close(ctx)

	var hostedResult []bson.M
	if err := hostedCursor.All(ctx, &hostedResult); err != nil {
		return totalMeetings, 0, err
	}

	if len(hostedResult) > 0 {
		if h, ok := hostedResult[0]["total"].(int32); ok {
			hostedMeetings = int(h)
		}
	}

	return totalMeetings, hostedMeetings, nil
}

func (r *BillingRepository) GetAllUsersUsage(ctx context.Context, period string) ([]*billing.UserBillingPeriod, error) {
	filter := bson.M{"billing_period": period}
	cursor, err := r.periodsColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var periods []*billing.UserBillingPeriod
	if err := cursor.All(ctx, &periods); err != nil {
		return nil, err
	}
	return periods, nil
}

func (r *BillingRepository) GetLastSync(ctx context.Context) (*billing.CloudflareUsageSync, error) {
	opts := options.FindOne().SetSort(bson.M{"last_sync_at": -1})
	var sync billing.CloudflareUsageSync
	err := r.syncColl.FindOne(ctx, bson.M{}, opts).Decode(&sync)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &sync, nil
}

func (r *BillingRepository) SaveSyncStatus(ctx context.Context, sync *billing.CloudflareUsageSync) error {
	// Upsert based on ID if provided, otherwise insert
	if sync.ID == "" {
		_, err := r.syncColl.InsertOne(ctx, sync)
		return err
	}

	filter := bson.M{"_id": sync.ID}
	update := bson.M{"$set": sync}
	options := options.Update().SetUpsert(true)
	_, err := r.syncColl.UpdateOne(ctx, filter, update, options)
	return err
}
