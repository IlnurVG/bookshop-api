package repositories

import (
	"context"
	"time"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/redis/go-redis/v9"
)

// CartRepository defines methods for working with shopping cart in storage
type CartRepository interface {
	// AddItem adds an item to the user's cart
	AddItem(ctx context.Context, userID int, bookID int, expiresAt time.Time) error

	// GetCart returns the user's cart
	GetCart(ctx context.Context, userID int) (*models.Cart, error)

	// RemoveItem removes an item from the user's cart
	RemoveItem(ctx context.Context, userID int, bookID int) error

	// ClearCart clears the user's cart
	ClearCart(ctx context.Context, userID int) error

	// GetExpiredCarts returns a list of expired carts
	GetExpiredCarts(ctx context.Context) ([]models.Cart, error)

	// RemoveExpiredItems removes expired items from carts
	RemoveExpiredItems(ctx context.Context) error

	// LockCart locks the cart during checkout
	// Returns an error if the cart is already locked
	LockCart(ctx context.Context, userID int, duration time.Duration) error

	// UnlockCart unlocks the cart
	UnlockCart(ctx context.Context, userID int) error

	// GetRedisClient returns the underlying Redis client
	GetRedisClient() *redis.Client
}
