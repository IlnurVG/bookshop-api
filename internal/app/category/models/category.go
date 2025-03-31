package models

import (
	"time"

	domainmodels "github.com/bookshop/api/internal/domain/models"
)

// Category represents a category model for service operations
type Category struct {
	ID        int
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CategoryCreate represents data for creating a category
type CategoryCreate struct {
	Name string
}

// CategoryUpdate represents data for updating a category
type CategoryUpdate struct {
	Name string
}

// ToDomain converts service model to domain model
func (c *Category) ToDomain() *domainmodels.Category {
	return &domainmodels.Category{
		ID:        c.ID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// FromDomain converts domain model to service model
func FromDomain(category *domainmodels.Category) *Category {
	return &Category{
		ID:        category.ID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}

// CategorySliceFromDomain converts a slice of domain models to service models
func CategorySliceFromDomain(categories []domainmodels.Category) []Category {
	result := make([]Category, len(categories))
	for i, category := range categories {
		categoryCopy := category // Create a copy to avoid issues with loop variable references
		serviceCategory := FromDomain(&categoryCopy)
		result[i] = *serviceCategory
	}
	return result
}

// CategoryCreateToDomain converts service category create model to domain model
func (cc *CategoryCreate) ToDomain() domainmodels.CategoryCreate {
	return domainmodels.CategoryCreate{
		Name: cc.Name,
	}
}

// CategoryUpdateToDomain converts service category update model to domain model
func (cu *CategoryUpdate) ToDomain() domainmodels.CategoryUpdate {
	return domainmodels.CategoryUpdate{
		Name: cu.Name,
	}
}
