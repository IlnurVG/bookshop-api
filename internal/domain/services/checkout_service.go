package services

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// CheckoutService defines methods for order processing
type CheckoutService interface {
	// Checkout creates an order from the user's cart
	Checkout(ctx context.Context, userID int) (*models.Order, error)

	// GetOrdersByUserID returns a list of user's orders
	GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error)

	// GetOrderByID returns an order by ID
	GetOrderByID(ctx context.Context, orderID int, userID int) (*models.Order, error)
}
