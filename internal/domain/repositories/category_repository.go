package repositories

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// CategoryRepository defines methods for working with categories in storage
type CategoryRepository interface {
	// Create creates a new category
	Create(ctx context.Context, category *models.Category) error

	// GetByID returns a category by ID
	GetByID(ctx context.Context, id int) (*models.Category, error)

	// GetByName returns a category by name
	GetByName(ctx context.Context, name string) (*models.Category, error)

	// List returns a list of all categories
	List(ctx context.Context) ([]models.Category, error)

	// Update updates category data
	Update(ctx context.Context, category *models.Category) error

	// Delete deletes a category by ID
	Delete(ctx context.Context, id int) error

	// GetCategoriesByIDs returns categories by a list of IDs
	GetCategoriesByIDs(ctx context.Context, ids []int) ([]models.Category, error)
}
