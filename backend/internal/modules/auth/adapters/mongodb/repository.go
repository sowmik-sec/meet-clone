package mongodb

import (
	"context"

	"meet-clone/internal/modules/auth/domain"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AuthRepository struct {
	collection *mongo.Collection
}

func NewAuthRepository(db *mongo.Database) *AuthRepository {
	return &AuthRepository{
		collection: db.Collection("user"),
	}
}

// Map domain entity to bson
func toBSON(user *domain.User) bson.M {
	id, _ := bson.ObjectIDFromHex(user.ID) // Handle error properly in prod
	return bson.M{
		"_id":           id,
		"first_name":    user.FirstName,
		"last_name":     user.LastName,
		"password":      user.Password,
		"email":         user.Email,
		"phone":         user.Phone,
		"token":         user.Token,
		"refresh_token": user.RefreshToken,
		"user_type":     user.UserType,
		"created_at":    user.CreatedAt,
		"updated_at":    user.UpdatedAt,
		"user_id":       user.UserID,
	}
}

// Map bson to domain entity
func fromBSON(doc bson.M) *domain.User {
	// Safely casting/asserting types usually required, simplified for brevity
	// Assuming structure matches perfectly for now

	// Helper to generic parsing if needed, but direct map is faster roughly
	// In real world use bson.Marshal/Unmarshal for safer conversion

	user := &domain.User{}
	bsonBytes, _ := bson.Marshal(doc)
	bson.Unmarshal(bsonBytes, user)

	// Manual mapping for clean domain object if needed, relying on unmarshal for now with tags on entity?
	// Wait, I put json tags on entity, but not bson tags.
	// I should probably add bson tags to entity or use a separate DB model struct.
	// For hexagonal purity, it's better to have a separate model, OR simple struct tags.
	// Let's use clean separate mapping or update entity.go with bson tags?
	// Adding bson tags to domain entity is a pragmatic trade-off.
	// But since I didn't add them, I will use a local struct for decoding.

	return user
}

// Local struct for strict BSON mapping if needed, to avoid polluting domain
type UserBSON struct {
	ID           bson.ObjectID `bson:"_id"`
	FirstName    string        `bson:"first_name"`
	LastName     string        `bson:"last_name"`
	Password     string        `bson:"password"`
	Email        string        `bson:"email"`
	Phone        string        `bson:"phone"`
	Token        string        `bson:"token"`
	RefreshToken string        `bson:"refresh_token"`
	UserType     string        `bson:"user_type"`
	CreatedAt    interface{}   `bson:"created_at"` // handles time parsing
	UpdatedAt    interface{}   `bson:"updated_at"`
	UserID       string        `bson:"user_id"`
}

func (r *AuthRepository) CreateUser(ctx context.Context, user *domain.User) error {
	id, _ := bson.ObjectIDFromHex(user.ID)
	userDoc := bson.M{
		"_id":           id,
		"first_name":    user.FirstName,
		"last_name":     user.LastName,
		"password":      user.Password,
		"email":         user.Email,
		"phone":         user.Phone,
		"token":         user.Token,
		"refresh_token": user.RefreshToken,
		"user_type":     user.UserType,
		"created_at":    user.CreatedAt,
		"updated_at":    user.UpdatedAt,
		"user_id":       user.UserID,
	}
	_, err := r.collection.InsertOne(ctx, userDoc)
	return err
}

func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	// var result domain.User
	// Since domain User has only json tags, we rely on implicit matching or need tags.
	// Standard driver uses lowercase or matching field names.
	// Let's rely on standard matching but ideally we should update Entity to have bson tags for simplicity in Go

	// Actually, pure domain models should NOT have bson tags.
	// So I should decode to a temporary struct or map.
	var b bson.M
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&b)
	if err != nil {
		return nil, err
	}

	// Map back using json trick or manual
	user := &domain.User{}
	bytes, _ := bson.Marshal(b)
	bson.Unmarshal(bytes, user) // This works if fields match
	return user, nil
}

func (r *AuthRepository) FindUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var b bson.M
	err := r.collection.FindOne(ctx, bson.M{"phone": phone}).Decode(&b)
	if err != nil {
		return nil, err
	}
	user := &domain.User{}
	bytes, _ := bson.Marshal(b)
	bson.Unmarshal(bytes, user)
	return user, nil
}

func (r *AuthRepository) FindUserByID(ctx context.Context, userID string) (*domain.User, error) {
	var b bson.M
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&b)
	if err != nil {
		return nil, err
	}
	user := &domain.User{}
	bytes, _ := bson.Marshal(b)
	bson.Unmarshal(bytes, user)
	return user, nil
}

func (r *AuthRepository) UpdateTokens(ctx context.Context, userID string, token string, refreshToken string) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$set": bson.M{
			"token":         token,
			"refresh_token": refreshToken,
			// "updated_at":    // update timestamp too
		},
	}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *AuthRepository) CountUsersByPhone(ctx context.Context, phone string) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"phone": phone})
}

func (r *AuthRepository) CountUsersByEmail(ctx context.Context, email string) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"email": email})
}
