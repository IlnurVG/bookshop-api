package models

import "time"

// Book represents a book model
type Book struct {
	ID            int       `json:"id" db:"id"`
	Title         string    `json:"title" db:"title"`
	Author        string    `json:"author" db:"author"`
	YearPublished int       `json:"year_published" db:"year_published"`
	Price         float64   `json:"price" db:"price"`
	Stock         int       `json:"stock" db:"stock"`
	CategoryID    int       `json:"category_id" db:"category_id"`
	Category      *Category `json:"category,omitempty" db:"-"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// BookCreate represents data for creating a book
type BookCreate struct {
	Title         string  `json:"title" validate:"required"`
	Author        string  `json:"author" validate:"required"`
	YearPublished int     `json:"year_published" validate:"required,gt=0"`
	Price         float64 `json:"price" validate:"required,gt=0"`
	Stock         int     `json:"stock" validate:"required,gte=0"`
	CategoryID    int     `json:"category_id" validate:"required,gt=0"`
}

// BookUpdate represents data for updating a book
type BookUpdate struct {
	Title         *string  `json:"title,omitempty"`
	Author        *string  `json:"author,omitempty"`
	YearPublished *int     `json:"year_published,omitempty" validate:"omitempty,gt=0"`
	Price         *float64 `json:"price,omitempty" validate:"omitempty,gt=0"`
	CategoryID    *int     `json:"category_id,omitempty" validate:"omitempty,gt=0"`
}

// BookFilter represents book filtering parameters
type BookFilter struct {
	CategoryIDs []int    `json:"category_ids" form:"category_ids"`
	MinPrice    *float64 `json:"min_price,omitempty" form:"min_price"`
	MaxPrice    *float64 `json:"max_price,omitempty" form:"max_price"`
	InStock     *bool    `json:"in_stock,omitempty" form:"in_stock"`
	Page        int      `json:"page" form:"page"`
	PageSize    int      `json:"page_size" form:"page_size"`
}

// BookListResponse represents a response with a list of books
type BookListResponse struct {
	Books      []Book `json:"books"`
	TotalCount int    `json:"total_count"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalPages int    `json:"total_pages"`
}
