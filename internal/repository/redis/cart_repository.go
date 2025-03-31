package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/redis/go-redis/v9"
)

const (
	// cartKeyPrefix prefix for cart keys
	cartKeyPrefix = "cart:"
	// cartLockKeyPrefix prefix for cart lock keys
	cartLockKeyPrefix = "cart_lock:"
)

// CartRepository implements repositories.CartRepository interface
type CartRepository struct {
	client *redis.Client
}

// NewCartRepository creates a new instance of cart repository
func NewCartRepository(client *redis.Client) *CartRepository {
	return &CartRepository{
		client: client,
	}
}

// AddItem adds an item to the user's cart
func (r *CartRepository) AddItem(ctx context.Context, userID int, bookID int, expiresAt time.Time) error {
	// Check if cart is locked
	if r.isCartLocked(ctx, userID) {
		return fmt.Errorf("cart is locked")
	}

	// Create cart item
	item := models.CartItem{
		BookID:    bookID,
		AddedAt:   time.Now(),
		ExpiresAt: expiresAt,
	}

	// Serialize item
	itemJSON, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("error serializing cart item: %w", err)
	}

	// Add item to cart
	key := fmt.Sprintf("%s%d", cartKeyPrefix, userID)
	if err := r.client.HSet(ctx, key, fmt.Sprint(bookID), itemJSON).Err(); err != nil {
		return fmt.Errorf("error adding item to cart: %w", err)
	}

	return nil
}

// GetCart returns the user's cart
func (r *CartRepository) GetCart(ctx context.Context, userID int) (*models.Cart, error) {
	// Get all cart items
	key := fmt.Sprintf("%s%d", cartKeyPrefix, userID)
	items, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("error getting cart: %w", err)
	}

	// Create cart
	cart := &models.Cart{
		UserID: userID,
		Items:  make([]models.CartItem, 0, len(items)),
	}

	// Deserialize items
	for _, itemJSON := range items {
		var item models.CartItem
		if err := json.Unmarshal([]byte(itemJSON), &item); err != nil {
			return nil, fmt.Errorf("error deserializing cart item: %w", err)
		}

		// Check item expiration
		if time.Now().After(item.ExpiresAt) {
			continue
		}

		cart.Items = append(cart.Items, item)
	}

	return cart, nil
}

// RemoveItem removes an item from the user's cart
func (r *CartRepository) RemoveItem(ctx context.Context, userID int, bookID int) error {
	// Check if cart is locked
	if r.isCartLocked(ctx, userID) {
		return fmt.Errorf("cart is locked")
	}

	// Remove item from cart
	key := fmt.Sprintf("%s%d", cartKeyPrefix, userID)
	if err := r.client.HDel(ctx, key, fmt.Sprint(bookID)).Err(); err != nil {
		return fmt.Errorf("error removing item from cart: %w", err)
	}

	return nil
}

// ClearCart clears the user's cart
func (r *CartRepository) ClearCart(ctx context.Context, userID int) error {
	// Check if cart is locked
	if r.isCartLocked(ctx, userID) {
		return fmt.Errorf("cart is locked")
	}

	// Delete cart
	key := fmt.Sprintf("%s%d", cartKeyPrefix, userID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("error clearing cart: %w", err)
	}

	return nil
}

// RemoveExpiredItems removes expired items from carts
func (r *CartRepository) RemoveExpiredItems(ctx context.Context) error {
	// Get all carts
	pattern := fmt.Sprintf("%s*", cartKeyPrefix)
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()

		// Get all cart items
		items, err := r.client.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}

		// Check each item
		for bookID, itemJSON := range items {
			var item models.CartItem
			if err := json.Unmarshal([]byte(itemJSON), &item); err != nil {
				continue
			}

			// Remove expired items
			if time.Now().After(item.ExpiresAt) {
				r.client.HDel(ctx, key, bookID)
			}
		}
	}

	return iter.Err()
}

// LockCart locks the cart during checkout
func (r *CartRepository) LockCart(ctx context.Context, userID int, duration time.Duration) error {
	// Check if cart is already locked
	if r.isCartLocked(ctx, userID) {
		return fmt.Errorf("cart is already locked")
	}

	// Lock cart
	key := fmt.Sprintf("%s%d", cartLockKeyPrefix, userID)
	if err := r.client.Set(ctx, key, "locked", duration).Err(); err != nil {
		return fmt.Errorf("error locking cart: %w", err)
	}

	return nil
}

// UnlockCart unlocks the cart
func (r *CartRepository) UnlockCart(ctx context.Context, userID int) error {
	// Remove lock
	key := fmt.Sprintf("%s%d", cartLockKeyPrefix, userID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("error unlocking cart: %w", err)
	}

	return nil
}

// GetExpiredCarts returns a list of expired carts
func (r *CartRepository) GetExpiredCarts(ctx context.Context) ([]models.Cart, error) {
	// Get all carts
	pattern := fmt.Sprintf("%s*", cartKeyPrefix)
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()

	expiredCarts := make([]models.Cart, 0)

	for iter.Next(ctx) {
		key := iter.Val()

		// Extract user ID from key
		userID := 0
		if _, err := fmt.Sscanf(key, cartKeyPrefix+"%d", &userID); err != nil {
			continue
		}

		// Get all cart items
		items, err := r.client.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}

		hasExpired := false
		cartItems := make([]models.CartItem, 0)

		// Check each item
		for _, itemJSON := range items {
			var item models.CartItem
			if err := json.Unmarshal([]byte(itemJSON), &item); err != nil {
				continue
			}

			// Check if item is expired
			if time.Now().After(item.ExpiresAt) {
				hasExpired = true
			}

			cartItems = append(cartItems, item)
		}

		// Add cart to result if it has expired items
		if hasExpired {
			cart := models.Cart{
				UserID: userID,
				Items:  cartItems,
			}
			expiredCarts = append(expiredCarts, cart)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("error scanning carts: %w", err)
	}

	return expiredCarts, nil
}

// isCartLocked checks if the cart is locked
func (r *CartRepository) isCartLocked(ctx context.Context, userID int) bool {
	key := fmt.Sprintf("%s%d", cartLockKeyPrefix, userID)
	exists, _ := r.client.Exists(ctx, key).Result()
	return exists == 1
}

// GetRedisClient returns the underlying Redis client
func (r *CartRepository) GetRedisClient() *redis.Client {
	return r.client
}
