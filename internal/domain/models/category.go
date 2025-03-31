package models

import "time"

// Category represents a book category model
type Category struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CategoryCreate represents data for creating a category
type CategoryCreate struct {
	Name string `json:"name" validate:"required"`
}

// CategoryUpdate represents data for updating a category
type CategoryUpdate struct {
	Name string `json:"name" validate:"required"`
}

// CategoryResponse represents a category response
type CategoryResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ToResponse converts a category model to API response
func (c *Category) ToResponse() CategoryResponse {
	return CategoryResponse{
		ID:   c.ID,
		Name: c.Name,
	}
}
