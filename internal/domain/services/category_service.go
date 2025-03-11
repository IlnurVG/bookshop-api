package services

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// CategoryService определяет методы для работы с категориями
type CategoryService interface {
	// Create создает новую категорию
	Create(ctx context.Context, input models.CategoryCreate) (*models.Category, error)

	// GetByID возвращает категорию по ID
	GetByID(ctx context.Context, id int) (*models.Category, error)

	// List возвращает список всех категорий
	List(ctx context.Context) ([]models.Category, error)

	// Update обновляет данные категории
	Update(ctx context.Context, id int, input models.CategoryUpdate) (*models.Category, error)

	// Delete удаляет категорию по ID
	Delete(ctx context.Context, id int) error
}

