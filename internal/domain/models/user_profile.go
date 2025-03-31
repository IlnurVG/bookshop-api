package models

import "time"

// UserProfileResponse represents a user profile with their orders
type UserProfileResponse struct {
	UUID   string                 `json:"uuid"`
	Name   string                 `json:"name"`
	Email  string                 `json:"email"`
	Orders []ProfileOrderResponse `json:"orders"`
}

// ProfileOrderResponse represents an order in the user profile API response
type ProfileOrderResponse struct {
	UUID      string    `json:"uuid"`
	Status    string    `json:"status"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateOrderRequest represents a request to create a new order
type CreateOrderRequest struct {
	Notes string `json:"notes,omitempty"`
}

// UpdateOrderRequest represents a request to update an existing order
type UpdateOrderRequest struct {
	Status string `json:"status,omitempty"`
	Notes  string `json:"notes,omitempty"`
}
