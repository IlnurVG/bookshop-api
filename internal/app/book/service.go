package book

import (
	"context"
	"fmt"
	"time"

	servicemodels "github.com/bookshop/api/internal/app/book/models"
	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/bookshop/api/internal/domain/services"
)

// Service implements services.BookService interface
type Service struct {
	bookRepo     repositories.BookRepository
	categoryRepo repositories.CategoryRepository
	txManager    repositories.TransactionManager
}

// NewService creates a new instance of the book service
func NewService(
	bookRepo repositories.BookRepository,
	categoryRepo repositories.CategoryRepository,
	txManager repositories.TransactionManager,
) services.BookService {
	return &Service{
		bookRepo:     bookRepo,
		categoryRepo: categoryRepo,
		txManager:    txManager,
	}
}

// Create creates a new book
func (s *Service) Create(ctx context.Context, input models.BookCreate) (*models.Book, error) {
	var domainBook *models.Book

	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Convert to service model
		serviceInput := servicemodels.BookCreate{
			Title:         input.Title,
			Author:        input.Author,
			YearPublished: input.YearPublished,
			Price:         input.Price,
			Stock:         input.Stock,
			CategoryID:    input.CategoryID,
		}

		// Check if the category exists
		_, err := s.categoryRepo.GetByID(txCtx, serviceInput.CategoryID)
		if err != nil {
			return fmt.Errorf("error checking category: %w", err)
		}

		// Create a new book
		now := time.Now()
		domainBook = &models.Book{
			Title:         serviceInput.Title,
			Author:        serviceInput.Author,
			YearPublished: serviceInput.YearPublished,
			Price:         serviceInput.Price,
			Stock:         serviceInput.Stock,
			CategoryID:    serviceInput.CategoryID,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		// Save book to database
		if err := s.bookRepo.Create(txCtx, domainBook); err != nil {
			return fmt.Errorf("error creating book: %w", err)
		}

		// Load category information
		category, err := s.categoryRepo.GetByID(txCtx, domainBook.CategoryID)
		if err != nil {
			// Not returning an error as the book has already been created
			return nil
		}
		domainBook.Category = category

		return nil
	})

	if err != nil {
		return nil, err
	}

	return domainBook, nil
}

// GetByID returns a book by its ID
func (s *Service) GetByID(ctx context.Context, id int) (*models.Book, error) {
	// Get book from repository
	domainBook, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting book: %w", err)
	}

	// Load category information if needed
	if domainBook.Category == nil {
		category, err := s.categoryRepo.GetByID(ctx, domainBook.CategoryID)
		if err == nil {
			domainBook.Category = category
		}
		// Not returning an error as the book was found
	}

	return domainBook, nil
}

// List returns a list of books with filtering
func (s *Service) List(ctx context.Context, filter models.BookFilter) (*models.BookListResponse, error) {
	// Set default values for pagination
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	// Get book list from repository
	books, totalCount, err := s.bookRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error getting book list: %w", err)
	}

	// Load category information
	if len(books) > 0 {
		categoryIDs := make([]int, 0, len(books))
		for _, book := range books {
			categoryIDs = append(categoryIDs, book.CategoryID)
		}

		categories, err := s.categoryRepo.GetCategoriesByIDs(ctx, categoryIDs)
		if err == nil {
			// Create a map of categories for quick access
			categoryMap := make(map[int]*models.Category, len(categories))
			for i := range categories {
				categoryMap[categories[i].ID] = &categories[i]
			}

			// Set categories for books
			for i := range books {
				if category, ok := categoryMap[books[i].CategoryID]; ok {
					books[i].Category = category
				}
			}
		}
	}

	// Calculate total number of pages
	totalPages := totalCount / filter.PageSize
	if totalCount%filter.PageSize > 0 {
		totalPages++
	}

	// Form the response
	response := &models.BookListResponse{
		Books:      books,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}

	return response, nil
}

// Update updates book data
func (s *Service) Update(ctx context.Context, id int, input models.BookUpdate) (*models.Book, error) {
	var updatedBook *models.Book

	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get current book
		book, err := s.bookRepo.GetByID(txCtx, id)
		if err != nil {
			return fmt.Errorf("error getting book: %w", err)
		}

		// Update book fields
		if input.Title != nil {
			book.Title = *input.Title
		}
		if input.Author != nil {
			book.Author = *input.Author
		}
		if input.YearPublished != nil {
			book.YearPublished = *input.YearPublished
		}
		if input.Price != nil {
			book.Price = *input.Price
		}
		if input.CategoryID != nil {
			// Check if the category exists
			_, err := s.categoryRepo.GetByID(txCtx, *input.CategoryID)
			if err != nil {
				return fmt.Errorf("error checking category: %w", err)
			}
			book.CategoryID = *input.CategoryID
		}

		// Update modification time
		book.UpdatedAt = time.Now()

		// Save changes to database
		if err := s.bookRepo.Update(txCtx, book); err != nil {
			return fmt.Errorf("error updating book: %w", err)
		}

		// Load category information
		category, err := s.categoryRepo.GetByID(txCtx, book.CategoryID)
		if err != nil {
			// Not returning an error as the book has already been updated
			updatedBook = book
			return nil
		}
		book.Category = category
		updatedBook = book

		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedBook, nil
}

// Delete deletes a book by ID
func (s *Service) Delete(ctx context.Context, id int) error {
	return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.bookRepo.Delete(txCtx, id); err != nil {
			return fmt.Errorf("error deleting book: %w", err)
		}
		return nil
	})
}

// GetBooksByIDs returns books by a list of IDs
func (s *Service) GetBooksByIDs(ctx context.Context, ids []int) ([]models.Book, error) {
	if len(ids) == 0 {
		return []models.Book{}, nil
	}

	books, err := s.bookRepo.GetBooksByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("error getting books by IDs: %w", err)
	}

	return books, nil
}
