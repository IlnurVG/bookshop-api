package book

import (
	"time"

	"github.com/bookshop/api/internal/domain/models"
)

// CreateBookRequest представляет запрос на создание книги
type CreateBookRequest struct {
	Title         string  `json:"title" validate:"required"`
	Author        string  `json:"author" validate:"required"`
	YearPublished int     `json:"year_published" validate:"required,gt=0"`
	Price         float64 `json:"price" validate:"required,gt=0"`
	Stock         int     `json:"stock" validate:"required,gte=0"`
	CategoryID    int     `json:"category_id" validate:"required,gt=0"`
}

// ToModel преобразует CreateBookRequest в модель BookCreate
func (r *CreateBookRequest) ToModel() models.BookCreate {
	return models.BookCreate{
		Title:         r.Title,
		Author:        r.Author,
		YearPublished: r.YearPublished,
		Price:         r.Price,
		Stock:         r.Stock,
		CategoryID:    r.CategoryID,
	}
}

// UpdateBookRequest представляет запрос на обновление книги
type UpdateBookRequest struct {
	Title         *string  `json:"title,omitempty"`
	Author        *string  `json:"author,omitempty"`
	YearPublished *int     `json:"year_published,omitempty" validate:"omitempty,gt=0"`
	Price         *float64 `json:"price,omitempty" validate:"omitempty,gt=0"`
	CategoryID    *int     `json:"category_id,omitempty" validate:"omitempty,gt=0"`
}

// ToModel преобразует UpdateBookRequest в модель BookUpdate
func (r *UpdateBookRequest) ToModel() models.BookUpdate {
	return models.BookUpdate{
		Title:         r.Title,
		Author:        r.Author,
		YearPublished: r.YearPublished,
		Price:         r.Price,
		CategoryID:    r.CategoryID,
	}
}

// BookResponse представляет ответ с информацией о книге
type BookResponse struct {
	ID            int       `json:"id"`
	Title         string    `json:"title"`
	Author        string    `json:"author"`
	YearPublished int       `json:"year_published"`
	Price         float64   `json:"price"`
	Stock         int       `json:"stock"`
	CategoryID    int       `json:"category_id"`
	Category      *Category `json:"category,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Category представляет информацию о категории книги
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// FromModel преобразует модель Book в BookResponse
func FromModel(book *models.Book) *BookResponse {
	response := &BookResponse{
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
		response.Category = &Category{
			ID:   book.Category.ID,
			Name: book.Category.Name,
		}
	}

	return response
}

// BookListRequest представляет запрос на получение списка книг
type BookListRequest struct {
	CategoryIDs []int    `form:"category_ids"`
	MinPrice    *float64 `form:"min_price"`
	MaxPrice    *float64 `form:"max_price"`
	InStock     *bool    `form:"in_stock"`
	Page        int      `form:"page,default=1"`
	PageSize    int      `form:"page_size,default=10"`
}

// ToModel преобразует BookListRequest в модель BookFilter
func (r *BookListRequest) ToModel() models.BookFilter {
	return models.BookFilter{
		CategoryIDs: r.CategoryIDs,
		MinPrice:    r.MinPrice,
		MaxPrice:    r.MaxPrice,
		InStock:     r.InStock,
		Page:        r.Page,
		PageSize:    r.PageSize,
	}
}

// BookListResponse представляет ответ со списком книг
type BookListResponse struct {
	Books      []BookResponse `json:"books"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// FromModelList преобразует модель BookListResponse в BookListResponse
func FromModelList(modelResponse *models.BookListResponse) *BookListResponse {
	response := &BookListResponse{
		TotalCount: modelResponse.TotalCount,
		Page:       modelResponse.Page,
		PageSize:   modelResponse.PageSize,
		TotalPages: modelResponse.TotalPages,
		Books:      make([]BookResponse, 0, len(modelResponse.Books)),
	}

	for _, book := range modelResponse.Books {
		bookCopy := book // Создаем копию, чтобы избежать проблем с указателями
		response.Books = append(response.Books, *FromModel(&bookCopy))
	}

	return response
}
