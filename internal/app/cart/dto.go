package cart

import "time"

// CartItemRequest represents a request to add an item to the cart
type CartItemRequest struct {
	BookID int `json:"book_id" binding:"required"`
	Count  int `json:"count" binding:"required,min=1"`
}

// CartItemResponse represents a cart item in the API response
type CartItemResponse struct {
	ID        int       `json:"id"`
	BookID    int       `json:"book_id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Price     float64   `json:"price"`
	Count     int       `json:"count"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CartResponse represents a cart in the API response
type CartResponse struct {
	Items      []CartItemResponse `json:"items"`
	TotalItems int                `json:"total_items"`
	TotalPrice float64            `json:"total_price"`
}

// UpdateCartItemRequest represents a request to update the quantity of an item in the cart
type UpdateCartItemRequest struct {
	Count int `json:"count" binding:"required,min=1"`
}
