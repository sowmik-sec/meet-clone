package user

import (
	"time"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        string    `json:"id" bson:"_id"`
	Email     string    `json:"email" bson:"email"`
	Password  string    `json:"-" bson:"password"`
	Name      string    `json:"name" bson:"name"`
	Avatar    string    `json:"avatar" bson:"avatar"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

func NewUser(email, password, name string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		Email:     email,
		Password:  string(hashedPassword),
		Name:      name,
		Avatar:    generateDefaultAvatar(email),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (u *User) ComparePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func generateDefaultAvatar(email string) string {
	// Use a service like UI Avatars or Gravatar
	return "https://ui-avatars.com/api/?name=" + email + "&background=random"
}
