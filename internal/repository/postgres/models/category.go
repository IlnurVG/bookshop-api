package models

import (
	"time"

	domainmodels "github.com/bookshop/api/internal/domain/models"
)

// Category represents a category model for repository operations
type Category struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// ToDomain converts repository model to domain model
func (c *Category) ToDomain() *domainmodels.Category {
	return &domainmodels.Category{
		ID:        c.ID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// CategoryFromDomain converts domain model to repository model
func CategoryFromDomain(category *domainmodels.Category) *Category {
	return &Category{
		ID:        category.ID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}

// CategorySliceToDomain converts a slice of repository models to domain models
func CategorySliceToDomain(categories []Category) []domainmodels.Category {
	result := make([]domainmodels.Category, len(categories))
	for i, category := range categories {
		domainCategory := category.ToDomain()
		result[i] = *domainCategory
	}
	return result
}
