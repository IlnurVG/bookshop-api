package models

import (
	"time"

	domainmodels "github.com/bookshop/api/internal/domain/models"
)

// Book represents a book model for service operations
type Book struct {
	ID            int
	Title         string
	Author        string
	YearPublished int
	Price         float64
	Stock         int
	CategoryID    int
	Category      *Category
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Category represents a book category
type Category struct {
	ID   int
	Name string
}

// BookCreate represents data for creating a book
type BookCreate struct {
	Title         string
	Author        string
	YearPublished int
	Price         float64
	Stock         int
	CategoryID    int
}

// BookUpdate represents data for updating a book
type BookUpdate struct {
	Title         *string
	Author        *string
	YearPublished *int
	Price         *float64
	CategoryID    *int
}

// BookFilter represents book filtering parameters
type BookFilter struct {
	CategoryIDs []int
	MinPrice    *float64
	MaxPrice    *float64
	InStock     *bool
	Page        int
	PageSize    int
}

// BookListResponse represents a response with a list of books
type BookListResponse struct {
	Books      []Book
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
}

// ToDomain converts service book model to domain model
func (b *Book) ToDomain() *domainmodels.Book {
	domainBook := &domainmodels.Book{
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

	if b.Category != nil {
		domainBook.Category = &domainmodels.Category{
			ID:   b.Category.ID,
			Name: b.Category.Name,
		}
	}

	return domainBook
}

// FromDomain converts domain model to service model
func FromDomain(book *domainmodels.Book) *Book {
	serviceBook := &Book{
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

	if book.Category != nil {
		serviceBook.Category = &Category{
			ID:   book.Category.ID,
			Name: book.Category.Name,
		}
	}

	return serviceBook
}

// BookSliceFromDomain converts a slice of domain models to service models
func BookSliceFromDomain(books []domainmodels.Book) []Book {
	result := make([]Book, len(books))
	for i, book := range books {
		bookCopy := book // Create a copy to avoid issues with loop variable references
		serviceBook := FromDomain(&bookCopy)
		result[i] = *serviceBook
	}
	return result
}

// BookCreateToDomain converts service book create model to domain model
func (bc *BookCreate) ToDomain() domainmodels.BookCreate {
	return domainmodels.BookCreate{
		Title:         bc.Title,
		Author:        bc.Author,
		YearPublished: bc.YearPublished,
		Price:         bc.Price,
		Stock:         bc.Stock,
		CategoryID:    bc.CategoryID,
	}
}

// BookUpdateToDomain converts service book update model to domain model
func (bu *BookUpdate) ToDomain() domainmodels.BookUpdate {
	return domainmodels.BookUpdate{
		Title:         bu.Title,
		Author:        bu.Author,
		YearPublished: bu.YearPublished,
		Price:         bu.Price,
		CategoryID:    bu.CategoryID,
	}
}

// BookFilterToDomain converts service book filter model to domain model
func (bf *BookFilter) ToDomain() domainmodels.BookFilter {
	return domainmodels.BookFilter{
		CategoryIDs: bf.CategoryIDs,
		MinPrice:    bf.MinPrice,
		MaxPrice:    bf.MaxPrice,
		InStock:     bf.InStock,
		Page:        bf.Page,
		PageSize:    bf.PageSize,
	}
}

// BookListResponseFromDomain converts domain book list response to service model
func BookListResponseFromDomain(dlr *domainmodels.BookListResponse) *BookListResponse {
	return &BookListResponse{
		Books:      BookSliceFromDomain(dlr.Books),
		TotalCount: dlr.TotalCount,
		Page:       dlr.Page,
		PageSize:   dlr.PageSize,
		TotalPages: dlr.TotalPages,
	}
}
