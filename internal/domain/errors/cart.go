package errors

import "errors"

var (
	// ErrCartEmpty indicates that the cart is empty
	ErrCartEmpty = errors.New("cart is empty")

	// ErrCartExpired indicates that the cart has expired
	ErrCartExpired = errors.New("cart expired")

	// ErrItemNotFound indicates that an item was not found in the cart
	ErrItemNotFound = errors.New("item not found in cart")
)
