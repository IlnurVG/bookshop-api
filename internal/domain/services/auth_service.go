package services

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// AuthService defines methods for authentication
type AuthService interface {
	// Register registers a new user
	Register(ctx context.Context, input models.UserRegistration) (*models.User, error)

	// Login authenticates a user and returns tokens
	Login(ctx context.Context, input models.UserCredentials) (string, string, error)

	// ValidateToken validates a token and returns the user ID
	ValidateToken(ctx context.Context, token string) (int, error)

	// RefreshToken refreshes the access token
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)

	// GetUserByID returns a user by ID
	GetUserByID(ctx context.Context, id int) (*models.User, error)

	// IsAdmin checks if the user is an administrator
	IsAdmin(ctx context.Context, userID int) (bool, error)
}
