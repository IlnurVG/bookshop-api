package services

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// CategoryService defines methods for working with categories
type CategoryService interface {
	// Create creates a new category
	Create(ctx context.Context, input models.CategoryCreate) (*models.Category, error)

	// GetByID returns a category by ID
	GetByID(ctx context.Context, id int) (*models.Category, error)

	// List returns a list of all categories
	List(ctx context.Context) ([]models.Category, error)

	// Update updates category data
	Update(ctx context.Context, id int, input models.CategoryUpdate) (*models.Category, error)

	// Delete deletes a category by ID
	Delete(ctx context.Context, id int) error
}
