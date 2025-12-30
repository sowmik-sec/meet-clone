package application

import (
	"context"
	"errors"
	"time"

	"meet-clone/internal/modules/auth/domain"
	"meet-clone/pkg/crypto"
	"meet-clone/pkg/jwt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuthUseCase struct {
	repo       domain.Repository
	jwtService *jwt.Service
}

func NewAuthUseCase(repo domain.Repository, jwtService *jwt.Service) *AuthUseCase {
	return &AuthUseCase{
		repo:       repo,
		jwtService: jwtService,
	}
}

func (uc *AuthUseCase) Signup(ctx context.Context, req SignupRequest) (*UserResponse, error) {
	// Check if email or phone already exists
	count, err := uc.repo.CountUsersByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("this email already exists")
	}

	count, err = uc.repo.CountUsersByPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("this phone number already exists")
	}

	// Hash password
	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user entity
	now := time.Now()
	objID := bson.NewObjectID()
	userID := objID.Hex()

	// Generate tokens
	token, refreshToken, err := uc.jwtService.GenerateAllTokens(req.Email, req.FirstName, req.LastName, req.UserType, userID)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           userID, // Using Hex ID as primary ID for domain
		UserID:       userID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Password:     hashedPassword,
		Email:        req.Email,
		Phone:        req.Phone,
		UserType:     req.UserType,
		Token:        token,
		RefreshToken: refreshToken,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

func (uc *AuthUseCase) Login(ctx context.Context, req LoginRequest) (*UserResponse, error) {
	user, err := uc.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("email or password is incorrect")
	}

	passwordIsValid, msg := crypto.VerifyPassword(user.Password, req.Password)
	if !passwordIsValid {
		return nil, errors.New(msg)
	}

	token, refreshToken, err := uc.jwtService.GenerateAllTokens(user.Email, user.FirstName, user.LastName, user.UserType, user.UserID)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.UpdateTokens(ctx, user.UserID, token, refreshToken); err != nil {
		return nil, err
	}

	// Update local struct
	user.Token = token
	user.RefreshToken = refreshToken

	return toUserResponse(user), nil
}

func toUserResponse(user *domain.User) *UserResponse {
	return &UserResponse{
		ID:           user.UserID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		Phone:        user.Phone,
		UserType:     user.UserType,
		Token:        user.Token,
		RefreshToken: user.RefreshToken,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}
