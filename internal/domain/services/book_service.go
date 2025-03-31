package services

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// BookService defines methods for working with books
type BookService interface {
	// Create creates a new book
	Create(ctx context.Context, input models.BookCreate) (*models.Book, error)

	// GetByID returns a book by ID
	GetByID(ctx context.Context, id int) (*models.Book, error)

	// List returns a list of books with filtering
	List(ctx context.Context, filter models.BookFilter) (*models.BookListResponse, error)

	// Update updates book data
	Update(ctx context.Context, id int, input models.BookUpdate) (*models.Book, error)

	// Delete deletes a book by ID
	Delete(ctx context.Context, id int) error

	// GetBooksByIDs returns books by a list of IDs
	GetBooksByIDs(ctx context.Context, ids []int) ([]models.Book, error)
}
