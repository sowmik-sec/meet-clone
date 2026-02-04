package mongodb

import (
	"context"

	"github.com/meet-clone/backend/internal/core/domain/eventtype"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type eventTypeRepository struct {
	collection *mongo.Collection
}

func NewEventTypeRepository(db *mongo.Database) eventtype.Repository {
	return &eventTypeRepository{
		collection: db.Collection("event_types"),
	}
}

func (r *eventTypeRepository) Create(ctx context.Context, et *eventtype.EventType) error {
	_, err := r.collection.InsertOne(ctx, et)
	return err
}

func (r *eventTypeRepository) GetByID(ctx context.Context, id string) (*eventtype.EventType, error) {
	var et eventtype.EventType
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&et)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Or return specific NotFound error
		}
		return nil, err
	}
	return &et, nil
}

func (r *eventTypeRepository) GetBySlug(ctx context.Context, userID, slug string) (*eventtype.EventType, error) {
	var et eventtype.EventType
	filter := bson.M{
		"user_id": userID,
		"slug":    slug,
	}
	err := r.collection.FindOne(ctx, filter).Decode(&et)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &et, nil
}

func (r *eventTypeRepository) ListByUserID(ctx context.Context, userID string) ([]*eventtype.EventType, error) {
	var eventTypes []*eventtype.EventType
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &eventTypes); err != nil {
		return nil, err
	}
	if eventTypes == nil {
		return []*eventtype.EventType{}, nil
	}
	return eventTypes, nil
}

func (r *eventTypeRepository) Update(ctx context.Context, et *eventtype.EventType) error {
	filter := bson.M{"_id": et.ID}
	update := bson.M{"$set": et}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *eventTypeRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
