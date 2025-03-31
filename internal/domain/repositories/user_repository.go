package repositories

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// UserRepository defines methods for working with users in storage
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *models.User) error

	// GetByID returns a user by ID
	GetByID(ctx context.Context, id int) (*models.User, error)

	// GetByEmail returns a user by email
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// Update updates user data
	Update(ctx context.Context, user *models.User) error

	// Delete deletes a user by ID
	Delete(ctx context.Context, id int) error
}
