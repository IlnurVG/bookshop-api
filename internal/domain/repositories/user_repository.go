package repositories

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// UserRepository определяет методы для работы с пользователями в хранилище
type UserRepository interface {
	// Create создает нового пользователя
	Create(ctx context.Context, user *models.User) error

	// GetByID возвращает пользователя по ID
	GetByID(ctx context.Context, id int) (*models.User, error)

	// GetByEmail возвращает пользователя по email
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// Update обновляет данные пользователя
	Update(ctx context.Context, user *models.User) error

	// Delete удаляет пользователя по ID
	Delete(ctx context.Context, id int) error
}

