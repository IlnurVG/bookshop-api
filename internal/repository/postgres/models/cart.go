package models

import (
	"time"

	domainmodels "github.com/bookshop/api/internal/domain/models"
)

// CartItem represents a cart item for repository operations
type CartItem struct {
	BookID    int       `db:"book_id"`
	AddedAt   time.Time `db:"added_at"`
	ExpiresAt time.Time `db:"expires_at"`
}

// Cart represents a user's shopping cart for repository operations
type Cart struct {
	UserID int `db:"user_id"`
	Items  []CartItem
}

// CartItemToDomain converts repository cart item to domain model
func (ci *CartItem) ToDomain() domainmodels.CartItem {
	return domainmodels.CartItem{
		BookID:    ci.BookID,
		AddedAt:   ci.AddedAt,
		ExpiresAt: ci.ExpiresAt,
	}
}

// CartItemFromDomain converts domain cart item to repository model
func CartItemFromDomain(item domainmodels.CartItem) CartItem {
	return CartItem{
		BookID:    item.BookID,
		AddedAt:   item.AddedAt,
		ExpiresAt: item.ExpiresAt,
	}
}

// CartToDomain converts repository cart to domain model
func (c *Cart) ToDomain() *domainmodels.Cart {
	domainCart := &domainmodels.Cart{
		UserID: c.UserID,
		Items:  make([]domainmodels.CartItem, len(c.Items)),
	}

	for i, item := range c.Items {
		domainCart.Items[i] = item.ToDomain()
	}

	return domainCart
}

// CartFromDomain converts domain cart to repository model
func CartFromDomain(cart *domainmodels.Cart) *Cart {
	repoCart := &Cart{
		UserID: cart.UserID,
		Items:  make([]CartItem, len(cart.Items)),
	}

	for i, item := range cart.Items {
		repoCart.Items[i] = CartItemFromDomain(item)
	}

	return repoCart
}
