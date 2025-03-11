package repositories

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// CategoryRepository определяет методы для работы с категориями в хранилище
type CategoryRepository interface {
	// Create создает новую категорию
	Create(ctx context.Context, category *models.Category) error

	// GetByID возвращает категорию по ID
	GetByID(ctx context.Context, id int) (*models.Category, error)

	// GetByName возвращает категорию по имени
	GetByName(ctx context.Context, name string) (*models.Category, error)

	// List возвращает список всех категорий
	List(ctx context.Context) ([]models.Category, error)

	// Update обновляет данные категории
	Update(ctx context.Context, category *models.Category) error

	// Delete удаляет категорию по ID
	Delete(ctx context.Context, id int) error

	// GetCategoriesByIDs возвращает категории по списку ID
	GetCategoriesByIDs(ctx context.Context, ids []int) ([]models.Category, error)
}

