package mongodb

import (
	"context"
	"time"

	"github.com/meet-clone/backend/internal/core/domain/appointment"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type appointmentRepository struct {
	collection *mongo.Collection
}

func NewAppointmentRepository(db *mongo.Database) appointment.Repository {
	return &appointmentRepository{
		collection: db.Collection("appointments"),
	}
}

func (r *appointmentRepository) Create(ctx context.Context, appt *appointment.Appointment) error {
	_, err := r.collection.InsertOne(ctx, appt)
	return err
}

func (r *appointmentRepository) FindByID(ctx context.Context, id string) (*appointment.Appointment, error) {
	var appt appointment.Appointment
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&appt)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Or return a specific ErrNotFound
		}
		return nil, err
	}
	return &appt, nil
}

func (r *appointmentRepository) FindByUser(ctx context.Context, userID string, filter appointment.AppointmentFilter) ([]appointment.Appointment, error) {
	// Find appointments where user is host OR guest
	query := bson.M{
		"$or": []bson.M{
			{"host_id": userID},
			{"guest_id": userID},
		},
	}

	if filter.Status != nil {
		query["status"] = *filter.Status
	}

	if filter.StartTimeAfter != nil {
		startTime, _ := time.Parse(time.RFC3339, *filter.StartTimeAfter)
		query["start_time"] = bson.M{"$gte": startTime}
	}

	if filter.StartTimeBefore != nil {
		endTime, _ := time.Parse(time.RFC3339, *filter.StartTimeBefore)
		// Usually we want appointments that START before this time, or overlap?
		// For simplicity, let's filter by start_time just like 'After'
		if _, ok := query["start_time"]; ok {
			query["start_time"].(bson.M)["$lte"] = endTime
		} else {
			query["start_time"] = bson.M{"$lte": endTime}
		}
	}

	opts := options.Find().SetSort(bson.D{{Key: "start_time", Value: 1}}) // Ascending order
	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var appointments []appointment.Appointment
	if err = cursor.All(ctx, &appointments); err != nil {
		return nil, err
	}
	// Return empty slice instead of nil if no docs
	if appointments == nil {
		appointments = []appointment.Appointment{}
	}
	return appointments, nil
}

func (r *appointmentRepository) Update(ctx context.Context, appt *appointment.Appointment) error {
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": appt.ID}, appt)
	return err
}

func (r *appointmentRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *appointmentRepository) HasConflict(ctx context.Context, userID string, startTime string, endTime string) (bool, error) {
	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return false, err
	}
	end, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return false, err
	}

	// Conflict condition:
	// (NewStart < ExistingEnd) AND (NewEnd > ExistingStart)
	// And status is CONFIRMED (ignoring cancelled)
	filter := bson.M{
		"$or": []bson.M{
			{"host_id": userID},
			{"guest_id": userID},
		},
		"status":              appointment.StatusConfirmed,
		"buffered_start_time": bson.M{"$lt": end},
		"buffered_end_time":   bson.M{"$gt": start},
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *appointmentRepository) GetBookedSlots(ctx context.Context, userID string, date string) ([][]string, error) {
	// Parse date (YYYY-MM-DD)
	startOfDay, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := bson.M{
		"$or": []bson.M{
			{"host_id": userID},
			{"guest_id": userID},
		},
		"status": appointment.StatusConfirmed,
		"buffered_start_time": bson.M{
			"$gte": startOfDay,
			"$lt":  endOfDay,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "start_time", Value: 1}})
	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var appointments []appointment.Appointment
	if err = cursor.All(ctx, &appointments); err != nil {
		return nil, err
	}

	var slots [][]string
	for _, appt := range appointments {
		slots = append(slots, []string{
			appt.BufferedStartTime.Format(time.RFC3339),
			appt.BufferedEndTime.Format(time.RFC3339),
		})
	}
	return slots, nil
}

func (r *appointmentRepository) FindUpcoming(ctx context.Context, start, end time.Time) ([]appointment.Appointment, error) {
	filter := bson.M{
		"status": appointment.StatusConfirmed,
		"start_time": bson.M{
			"$gte": start,
			"$lte": end,
		},
		"reminder_sent": bson.M{
			"$ne": true,
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var appointments []appointment.Appointment
	if err = cursor.All(ctx, &appointments); err != nil {
		return nil, err
	}
	// Return empty slice instead of nil if no docs
	if appointments == nil {
		appointments = []appointment.Appointment{}
	}
	return appointments, nil
}

func (r *appointmentRepository) FindByRoomID(ctx context.Context, roomID string) (*appointment.Appointment, error) {
	var appt appointment.Appointment
	err := r.collection.FindOne(ctx, bson.M{"room_id": roomID}).Decode(&appt)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &appt, nil
}

func (r *appointmentRepository) FindByRescheduleToken(ctx context.Context, token string) (*appointment.Appointment, error) {
	var appt appointment.Appointment
	err := r.collection.FindOne(ctx, bson.M{"reschedule_token": token}).Decode(&appt)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &appt, nil
}
