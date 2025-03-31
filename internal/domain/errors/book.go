package errors

import "errors"

var (
	// ErrBookNotFound indicates that a requested book was not found
	ErrBookNotFound = errors.New("book not found")

	// ErrBookOutOfStock indicates that a book is out of stock
	ErrBookOutOfStock = errors.New("book is out of stock")

	// ErrInvalidBookID indicates that a book ID is invalid
	ErrInvalidBookID = errors.New("invalid book ID")

	// ErrInvalidBookData indicates that book data is invalid
	ErrInvalidBookData = errors.New("invalid book data")

	// ErrCategoryNotFound indicates that a book's category was not found
	ErrCategoryNotFound = errors.New("category not found")
)
