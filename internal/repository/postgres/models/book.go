package models

import (
	"time"

	domainmodels "github.com/bookshop/api/internal/domain/models"
)

// Book represents a book model for repository operations
type Book struct {
	ID            int       `db:"id"`
	Title         string    `db:"title"`
	Author        string    `db:"author"`
	YearPublished int       `db:"year_published"`
	Price         float64   `db:"price"`
	Stock         int       `db:"stock"`
	CategoryID    int       `db:"category_id"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// ToDomain converts repository model to domain model
func (b *Book) ToDomain() *domainmodels.Book {
	return &domainmodels.Book{
		ID:            b.ID,
		Title:         b.Title,
		Author:        b.Author,
		YearPublished: b.YearPublished,
		Price:         b.Price,
		Stock:         b.Stock,
		CategoryID:    b.CategoryID,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
	}
}

// FromDomain converts domain model to repository model
func FromDomain(book *domainmodels.Book) *Book {
	return &Book{
		ID:            book.ID,
		Title:         book.Title,
		Author:        book.Author,
		YearPublished: book.YearPublished,
		Price:         book.Price,
		Stock:         book.Stock,
		CategoryID:    book.CategoryID,
		CreatedAt:     book.CreatedAt,
		UpdatedAt:     book.UpdatedAt,
	}
}

// BookSliceToDomain converts a slice of repository models to domain models
func BookSliceToDomain(books []Book) []domainmodels.Book {
	result := make([]domainmodels.Book, len(books))
	for i, book := range books {
		domainBook := book.ToDomain()
		result[i] = *domainBook
	}
	return result
}
