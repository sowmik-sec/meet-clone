package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Client struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewClient(uri, dbName string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(100).
		SetMinPoolSize(10)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the database
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	return &Client{
		client: client,
		db:     client.Database(dbName),
	}, nil
}

func (c *Client) GetCollection(name string) *mongo.Collection {
	return c.db.Collection(name)
}

func (c *Client) Disconnect(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

func (c *Client) CreateIndexes(ctx context.Context) error {
	// User indexes
	userIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := c.db.Collection("users").Indexes().CreateMany(ctx, userIndexes); err != nil {
		return err
	}

	// Room indexes
	roomIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "status", Value: 1}, {Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "created_by", Value: 1}},
		},
	}
	if _, err := c.db.Collection("rooms").Indexes().CreateMany(ctx, roomIndexes); err != nil {
		return err
	}

	// Message indexes
	messageIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "room_id", Value: 1}, {Key: "timestamp", Value: -1}},
		},
	}
	if _, err := c.db.Collection("chat_messages").Indexes().CreateMany(ctx, messageIndexes); err != nil {
		return err
	}

	return nil
}
