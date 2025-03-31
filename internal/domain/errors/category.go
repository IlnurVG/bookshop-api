package errors

import "errors"

var (
	// ErrCategoryExists indicates that a category with the given name already exists
	ErrCategoryExists = errors.New("category with this name already exists")
)
