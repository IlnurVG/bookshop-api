package category

import (
	"time"

	"github.com/bookshop/api/internal/domain/models"
)

// CreateCategoryRequest представляет запрос на создание категории
type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=2,max=50"`
}

// UpdateCategoryRequest представляет запрос на обновление категории
type UpdateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=2,max=50"`
}

// CategoryResponse представляет категорию в ответе API
type CategoryResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FromModel преобразует модель в ответ API
func FromModel(category *models.Category) *CategoryResponse {
	return &CategoryResponse{
		ID:        category.ID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}

// FromModelList преобразует список моделей в список ответов API
func FromModelList(categories []models.Category) []CategoryResponse {
	result := make([]CategoryResponse, len(categories))
	for i, category := range categories {
		result[i] = *FromModel(&category)
	}
	return result
}
