package services

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// BookService определяет методы для работы с книгами
type BookService interface {
	// Create создает новую книгу
	Create(ctx context.Context, input models.BookCreate) (*models.Book, error)

	// GetByID возвращает книгу по ID
	GetByID(ctx context.Context, id int) (*models.Book, error)

	// List возвращает список книг с фильтрацией
	List(ctx context.Context, filter models.BookFilter) (*models.BookListResponse, error)

	// Update обновляет данные книги
	Update(ctx context.Context, id int, input models.BookUpdate) (*models.Book, error)

	// Delete удаляет книгу по ID
	Delete(ctx context.Context, id int) error

	// GetBooksByIDs возвращает книги по списку ID
	GetBooksByIDs(ctx context.Context, ids []int) ([]models.Book, error)
}

