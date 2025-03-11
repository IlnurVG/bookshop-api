package repositories

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// BookRepository определяет методы для работы с книгами в хранилище
type BookRepository interface {
	// Create создает новую книгу
	Create(ctx context.Context, book *models.Book) error

	// GetByID возвращает книгу по ID
	GetByID(ctx context.Context, id int) (*models.Book, error)

	// List возвращает список книг с фильтрацией
	List(ctx context.Context, filter models.BookFilter) ([]models.Book, int, error)

	// Update обновляет данные книги
	Update(ctx context.Context, book *models.Book) error

	// Delete удаляет книгу по ID
	Delete(ctx context.Context, id int) error

	// UpdateStock обновляет количество книг на складе
	UpdateStock(ctx context.Context, id int, quantity int) error

	// GetBooksByIDs возвращает книги по списку ID
	GetBooksByIDs(ctx context.Context, ids []int) ([]models.Book, error)

	// ReserveBooks резервирует книги (уменьшает доступное количество)
	// Возвращает ошибку, если какой-то из товаров недоступен
	ReserveBooks(ctx context.Context, bookIDs []int) error

	// ReleaseBooks освобождает зарезервированные книги (увеличивает доступное количество)
	ReleaseBooks(ctx context.Context, bookIDs []int) error
}

