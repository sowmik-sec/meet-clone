package application

import "time"

type SignupRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
	Password  string `json:"password" validate:"required,min=6"`
	Email     string `json:"email" validate:"email,required"`
	Phone     string `json:"phone" validate:"required"`
	UserType  string `json:"user_type" validate:"required,eq=ADMIN|eq=USER"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"email,required"`
	Password string `json:"password" validate:"required,min=6"`
}

type UserResponse struct {
	ID           string    `json:"user_id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	UserType     string    `json:"user_type"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
