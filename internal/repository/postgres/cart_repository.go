package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// CartRepository implements repositories.CartRepository interface
type CartRepository struct {
	db *pgxpool.Pool
}

// NewCartRepository creates a new instance of CartRepository
func NewCartRepository(db *pgxpool.Pool) repositories.CartRepository {
	return &CartRepository{
		db: db,
	}
}

// AddItem adds an item to the user's cart
func (r *CartRepository) AddItem(ctx context.Context, userID int, bookID int, expiresAt time.Time) error {
	query := `
		INSERT INTO cart_items (user_id, book_id, added_at, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, book_id) 
		DO UPDATE SET expires_at = $4
	`

	_, err := r.db.Exec(ctx, query, userID, bookID, time.Now(), expiresAt)
	if err != nil {
		return fmt.Errorf("failed to add item to cart: %w", err)
	}

	return nil
}

// GetCart returns the user's cart
func (r *CartRepository) GetCart(ctx context.Context, userID int) (*models.Cart, error) {
	query := `
		SELECT ci.book_id, ci.added_at, ci.expires_at
		FROM cart_items ci
		WHERE ci.user_id = $1 AND ci.expires_at > $2
		ORDER BY ci.added_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}
	defer rows.Close()

	cart := &models.Cart{
		UserID: userID,
		Items:  make([]models.CartItem, 0),
	}

	for rows.Next() {
		item := models.CartItem{}
		err := rows.Scan(
			&item.BookID,
			&item.AddedAt,
			&item.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cart item data: %w", err)
		}
		cart.Items = append(cart.Items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate through results: %w", err)
	}

	return cart, nil
}

// RemoveItem removes an item from the user's cart
func (r *CartRepository) RemoveItem(ctx context.Context, userID int, bookID int) error {
	query := `
		DELETE FROM cart_items
		WHERE user_id = $1 AND book_id = $2
	`

	_, err := r.db.Exec(ctx, query, userID, bookID)
	if err != nil {
		return fmt.Errorf("failed to remove item from cart: %w", err)
	}

	return nil
}

// ClearCart clears the user's cart
func (r *CartRepository) ClearCart(ctx context.Context, userID int) error {
	query := `
		DELETE FROM cart_items
		WHERE user_id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	return nil
}

// GetExpiredCarts returns a list of expired carts
func (r *CartRepository) GetExpiredCarts(ctx context.Context) ([]models.Cart, error) {
	query := `
		SELECT DISTINCT user_id
		FROM cart_items
		WHERE expires_at <= $1
	`

	rows, err := r.db.Query(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get expired carts: %w", err)
	}
	defer rows.Close()

	carts := make([]models.Cart, 0)
	for rows.Next() {
		cart := models.Cart{}
		err := rows.Scan(&cart.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cart data: %w", err)
		}
		carts = append(carts, cart)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate through results: %w", err)
	}

	return carts, nil
}

// RemoveExpiredItems removes expired items from carts
func (r *CartRepository) RemoveExpiredItems(ctx context.Context) error {
	query := `
		DELETE FROM cart_items
		WHERE expires_at <= $1
	`

	_, err := r.db.Exec(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to remove expired items: %w", err)
	}

	return nil
}

// LockCart locks the cart during order checkout
func (r *CartRepository) LockCart(ctx context.Context, userID int, duration time.Duration) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if the cart is already locked
	lockQuery := `
		SELECT 1
		FROM cart_locks
		WHERE user_id = $1 AND locked_until > $2
	`

	var exists int
	err = tx.QueryRow(ctx, lockQuery, userID, time.Now()).Scan(&exists)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to check cart lock: %w", err)
	}

	if err == nil {
		// Cart is already locked
		return fmt.Errorf("cart is already locked")
	}

	// Lock the cart
	insertQuery := `
		INSERT INTO cart_locks (user_id, locked_until)
		VALUES ($1, $2)
		ON CONFLICT (user_id) 
		DO UPDATE SET locked_until = $2
	`

	_, err = tx.Exec(ctx, insertQuery, userID, time.Now().Add(duration))
	if err != nil {
		return fmt.Errorf("failed to lock cart: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UnlockCart unlocks the cart
func (r *CartRepository) UnlockCart(ctx context.Context, userID int) error {
	query := `
		DELETE FROM cart_locks
		WHERE user_id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to unlock cart: %w", err)
	}

	return nil
}

// GetRedisClient returns a Redis client (not implemented for Postgres)
func (r *CartRepository) GetRedisClient() *redis.Client {
	return nil // Postgres implementation doesn't use Redis
}
