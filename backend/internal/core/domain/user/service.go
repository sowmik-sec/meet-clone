package user

import (
	"context"
	"time"

	"github.com/meet-clone/backend/internal/pkg/errors"
)

type Service interface {
	Register(ctx context.Context, email, password, name string) (*User, error)
	Login(ctx context.Context, email, password string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	UpdateGoogleToken(ctx context.Context, userID, accessToken, refreshToken string, expiry time.Time) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Register(ctx context.Context, email, password, name string) (*User, error) {
	// Check if user already exists
	existingUser, _ := s.repo.FindByEmail(ctx, email)
	if existingUser != nil {
		return nil, errors.NewAlreadyExistsError("user with this email already exists")
	}

	// Create new user
	user, err := NewUser(email, password, name)
	if err != nil {
		return nil, errors.NewInternalError("failed to create user", err)
	}

	// Save to database
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, errors.NewInternalError("failed to save user", err)
	}

	return user, nil
}

func (s *service) Login(ctx context.Context, email, password string) (*User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.NewUnauthorizedError("invalid email or password")
	}

	if !user.ComparePassword(password) {
		return nil, errors.NewUnauthorizedError("invalid email or password")
	}

	return user, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.NewNotFoundError("user not found")
	}

	return user, nil
}

func (s *service) UpdateGoogleToken(ctx context.Context, userID, accessToken, refreshToken string, expiry time.Time) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.NewNotFoundError("user not found")
	}

	user.GoogleAccessToken = accessToken
	if refreshToken != "" {
		user.GoogleRefreshToken = refreshToken
	}
	user.GoogleTokenExpiry = expiry
	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return errors.NewInternalError("failed to update user token", err)
	}
	return nil
}
