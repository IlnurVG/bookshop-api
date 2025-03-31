package errors

import "errors"

var (
	// ErrNotFound indicates that a requested resource was not found
	ErrNotFound = errors.New("not found")

	// ErrDuplicateKey indicates that a uniqueness constraint was violated
	ErrDuplicateKey = errors.New("record with this key already exists")

	// ErrInvalidData indicates that provided data is invalid
	ErrInvalidData = errors.New("invalid data")

	// ErrOutOfStock indicates that an item is out of stock
	ErrOutOfStock = errors.New("out of stock")

	// ErrEmptyCart indicates that the cart is empty
	ErrEmptyCart = errors.New("cart is empty")
)
