package repositories

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// OrderRepository defines methods for working with orders in storage
type OrderRepository interface {
	// Create creates a new order
	Create(ctx context.Context, order *models.Order) error

	// GetByID returns an order by ID
	GetByID(ctx context.Context, id int) (*models.Order, error)

	// GetByUserID returns a list of user's orders
	GetByUserID(ctx context.Context, userID int) ([]models.Order, error)

	// Update updates order status
	UpdateStatus(ctx context.Context, id int, status string) error

	// AddOrderItem adds an item to the order
	AddOrderItem(ctx context.Context, orderID int, item models.OrderItem) error

	// GetOrderItems returns a list of items in the order
	GetOrderItems(ctx context.Context, orderID int) ([]models.OrderItem, error)
}
