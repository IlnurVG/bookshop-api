package errors

import "errors"

var (
	// ErrOrderNotFound indicates that a requested order was not found
	ErrOrderNotFound = errors.New("order not found")

	// ErrInvalidOrderStatus indicates that the order status is invalid
	ErrInvalidOrderStatus = errors.New("invalid order status")

	// ErrOrderAlreadyPaid indicates that the order has already been paid
	ErrOrderAlreadyPaid = errors.New("order has already been paid")

	// ErrOrderCanceled indicates that the order has been canceled
	ErrOrderCanceled = errors.New("order has been canceled")
)
