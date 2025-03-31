package models

import (
	"time"

	bookmodels "github.com/bookshop/api/internal/app/book/models"
	domainmodels "github.com/bookshop/api/internal/domain/models"
)

// CartItem represents a cart item for service operations
type CartItem struct {
	BookID    int
	Book      *bookmodels.Book
	AddedAt   time.Time
	ExpiresAt time.Time
}

// Cart represents a user's shopping cart for service operations
type Cart struct {
	UserID int
	Items  []CartItem
}

// CartItemRequest represents a request to add an item to the cart
type CartItemRequest struct {
	BookID int
}

// CartResponse represents a cart response
type CartResponse struct {
	Items     []CartItemResponse
	TotalCost float64
}

// CartItemResponse represents a cart item in API response
type CartItemResponse struct {
	BookID    int
	Title     string
	Author    string
	Price     float64
	AddedAt   time.Time
	ExpiresAt time.Time
}

// CartItemToDomain converts service cart item to domain model
func (ci *CartItem) ToDomain() domainmodels.CartItem {
	domainItem := domainmodels.CartItem{
		BookID:    ci.BookID,
		AddedAt:   ci.AddedAt,
		ExpiresAt: ci.ExpiresAt,
	}

	if ci.Book != nil {
		domainBook := ci.Book.ToDomain()
		domainItem.Book = domainBook
	}

	return domainItem
}

// CartItemFromDomain converts domain cart item to service model
func CartItemFromDomain(item domainmodels.CartItem) CartItem {
	serviceItem := CartItem{
		BookID:    item.BookID,
		AddedAt:   item.AddedAt,
		ExpiresAt: item.ExpiresAt,
	}

	if item.Book != nil {
		serviceItem.Book = bookmodels.FromDomain(item.Book)
	}

	return serviceItem
}

// CartToDomain converts service cart to domain model
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

// CartFromDomain converts domain cart to service model
func CartFromDomain(cart *domainmodels.Cart) *Cart {
	serviceCart := &Cart{
		UserID: cart.UserID,
		Items:  make([]CartItem, len(cart.Items)),
	}

	for i, item := range cart.Items {
		serviceCart.Items[i] = CartItemFromDomain(item)
	}

	return serviceCart
}

// CartItemRequestToDomain converts service request to domain model
func (cir *CartItemRequest) ToDomain() domainmodels.CartItemRequest {
	return domainmodels.CartItemRequest{
		BookID: cir.BookID,
	}
}
