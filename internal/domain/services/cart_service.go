package services

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// CartService defines methods for working with shopping cart
type CartService interface {
	// AddItem adds an item to the user's cart
	AddItem(ctx context.Context, userID int, input models.CartItemRequest) error

	// GetCart returns the user's cart
	GetCart(ctx context.Context, userID int) (*models.CartResponse, error)

	// RemoveItem removes an item from the user's cart
	RemoveItem(ctx context.Context, userID int, bookID int) error

	// ClearCart clears the user's cart
	ClearCart(ctx context.Context, userID int) error

	// CleanupExpiredItems removes expired items from carts
	CleanupExpiredItems(ctx context.Context) error
}
