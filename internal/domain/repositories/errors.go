package repositories

import "errors"

// Definition of common repository errors
var (
	// ErrNotFound is returned when a record is not found
	ErrNotFound = errors.New("record not found")

	// ErrDuplicateKey is returned when a uniqueness constraint is violated
	ErrDuplicateKey = errors.New("record with this key already exists")
)
