package repositories

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// BookRepository defines methods for working with books in storage
type BookRepository interface {
	// Create creates a new book
	Create(ctx context.Context, book *models.Book) error

	// GetByID returns a book by ID
	GetByID(ctx context.Context, id int) (*models.Book, error)

	// List returns a list of books with filtering
	List(ctx context.Context, filter models.BookFilter) ([]models.Book, int, error)

	// Update updates book data
	Update(ctx context.Context, book *models.Book) error

	// Delete deletes a book by ID
	Delete(ctx context.Context, id int) error

	// UpdateStock updates the quantity of books in stock
	UpdateStock(ctx context.Context, id int, quantity int) error

	// DecrementStock decreases the quantity of books in stock
	// Returns an error if there are not enough books in stock
	DecrementStock(ctx context.Context, id int, quantity int) error

	// GetBooksByIDs returns books by a list of IDs
	GetBooksByIDs(ctx context.Context, ids []int) ([]models.Book, error)

	// ReserveBooks reserves books (decreases available quantity)
	// Returns an error if any of the items is unavailable
	ReserveBooks(ctx context.Context, bookIDs []int) error

	// ReleaseBooks releases reserved books (increases available quantity)
	ReleaseBooks(ctx context.Context, bookIDs []int) error
}
