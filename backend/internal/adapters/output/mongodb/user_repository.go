package mongodb

import (
	"context"

	"github.com/google/uuid"
	"github.com/meet-clone/backend/internal/core/domain/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(client *Client) user.Repository {
	return &UserRepository{
		collection: client.GetCollection("users"),
	}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	u.ID = uuid.New().String()
	_, err := r.collection.InsertOne(ctx, u)
	return err
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	var u user.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": u.ID},
		bson.M{"$set": u},
	)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
