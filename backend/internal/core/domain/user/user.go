package user

import (
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type User struct {
	ID                 string    `json:"id" bson:"_id"`
	Email              string    `json:"email" bson:"email"`
	Password           string    `json:"-" bson:"password"`
	Name               string    `json:"name" bson:"name"`
	Avatar             string    `json:"avatar" bson:"avatar"`
	GoogleAccessToken  string    `json:"-" bson:"google_access_token,omitempty"`
	GoogleRefreshToken string    `json:"-" bson:"google_refresh_token,omitempty"`
	GoogleTokenExpiry  time.Time `json:"-" bson:"google_token_expiry,omitempty"`
	CreatedAt          time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" bson:"updated_at"`
}

func NewUser(email, password, name string) (*User, error) {
	// Validate email format
	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}

	// Validate password strength (minimum 6 characters)
	if len(password) < 6 {
		return nil, ErrWeakPassword
	}

	// Validate name
	if len(name) < 2 {
		return nil, ErrInvalidName
	}

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

// Validation errors
var (
	ErrInvalidEmail = &ValidationError{Message: "invalid email format"}
	ErrWeakPassword = &ValidationError{Message: "password must be at least 6 characters"}
	ErrInvalidName  = &ValidationError{Message: "name must be at least 2 characters"}
)

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
