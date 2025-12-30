package domain

import (
	"context"
)

type Repository interface {
	CreateUser(ctx context.Context, user *User) error
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserByPhone(ctx context.Context, phone string) (*User, error)
	FindUserByID(ctx context.Context, userID string) (*User, error)
	UpdateTokens(ctx context.Context, userID string, token string, refreshToken string) error
	CountUsersByPhone(ctx context.Context, phone string) (int64, error)
	CountUsersByEmail(ctx context.Context, email string) (int64, error)
}
