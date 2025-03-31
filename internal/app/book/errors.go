package book

import (
	"errors"
	"fmt"
)

// Definition of book-related errors
var (
	// ErrBookNotFound is returned when a book is not found
	ErrBookNotFound = errors.New("book not found")

	// ErrInvalidBookID is returned when a book ID is invalid
	ErrInvalidBookID = errors.New("invalid book ID")

	// ErrInvalidBookData is returned when book data is invalid
	ErrInvalidBookData = errors.New("invalid book data")

	// ErrCategoryNotFound is returned when a category is not found
	ErrCategoryNotFound = errors.New("category not found")

	// ErrBookOutOfStock is returned when a book is out of stock
	ErrBookOutOfStock = errors.New("book is out of stock")
)

// BookError represents a book-related error
type BookError struct {
	ID  int
	Err error
}

// Error returns the string representation of the error
func (e *BookError) Error() string {
	return fmt.Sprintf("book error (ID: %d): %v", e.ID, e.Err)
}

// Unwrap returns the original error
func (e *BookError) Unwrap() error {
	return e.Err
}

// NewBookError creates a new book error
func NewBookError(id int, err error) *BookError {
	return &BookError{
		ID:  id,
		Err: err,
	}
}
