package category

import (
	"time"

	"github.com/bookshop/api/internal/domain/models"
)

// CreateCategoryRequest represents a request to create a category
type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=2,max=50"`
}

// UpdateCategoryRequest represents a request to update a category
type UpdateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=2,max=50"`
}

// CategoryResponse represents a category in the API response
type CategoryResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// fromModel converts a model to an API response
func fromModel(category *models.Category) *CategoryResponse {
	return &CategoryResponse{
		ID:        category.ID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}

// fromModelList converts a list of models to a list of API responses
func fromModelList(categories []models.Category) []CategoryResponse {
	result := make([]CategoryResponse, len(categories))
	for i, category := range categories {
		result[i] = *fromModel(&category)
	}
	return result
}
