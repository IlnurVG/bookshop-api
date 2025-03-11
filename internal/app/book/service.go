package book

import (
	"context"
	"fmt"
	"time"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/bookshop/api/internal/domain/services"
)

// Service реализует интерфейс services.BookService
type Service struct {
	bookRepo     repositories.BookRepository
	categoryRepo repositories.CategoryRepository
}

// NewService создает новый экземпляр сервиса для работы с книгами
func NewService(bookRepo repositories.BookRepository, categoryRepo repositories.CategoryRepository) services.BookService {
	return &Service{
		bookRepo:     bookRepo,
		categoryRepo: categoryRepo,
	}
}

// Create создает новую книгу
func (s *Service) Create(ctx context.Context, input models.BookCreate) (*models.Book, error) {
	// Проверяем существование категории
	_, err := s.categoryRepo.GetByID(ctx, input.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки категории: %w", err)
	}

	// Создаем новую книгу
	now := time.Now()
	book := &models.Book{
		Title:         input.Title,
		Author:        input.Author,
		YearPublished: input.YearPublished,
		Price:         input.Price,
		Stock:         input.Stock,
		CategoryID:    input.CategoryID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Сохраняем книгу в базе данных
	if err := s.bookRepo.Create(ctx, book); err != nil {
		return nil, fmt.Errorf("ошибка создания книги: %w", err)
	}

	// Загружаем информацию о категории
	category, err := s.categoryRepo.GetByID(ctx, book.CategoryID)
	if err != nil {
		// Не возвращаем ошибку, так как книга уже создана
		return book, nil
	}
	book.Category = category

	return book, nil
}

// GetByID возвращает книгу по ID
func (s *Service) GetByID(ctx context.Context, id int) (*models.Book, error) {
	// Получаем книгу из репозитория
	book, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения книги: %w", err)
	}

	// Загружаем информацию о категории
	category, err := s.categoryRepo.GetByID(ctx, book.CategoryID)
	if err != nil {
		// Не возвращаем ошибку, так как книга найдена
		return book, nil
	}
	book.Category = category

	return book, nil
}

// List возвращает список книг с фильтрацией
func (s *Service) List(ctx context.Context, filter models.BookFilter) (*models.BookListResponse, error) {
	// Устанавливаем значения по умолчанию для пагинации
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	// Получаем список книг из репозитория
	books, totalCount, err := s.bookRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка книг: %w", err)
	}

	// Загружаем информацию о категориях
	if len(books) > 0 {
		categoryIDs := make([]int, 0, len(books))
		for _, book := range books {
			categoryIDs = append(categoryIDs, book.CategoryID)
		}

		categories, err := s.categoryRepo.GetCategoriesByIDs(ctx, categoryIDs)
		if err == nil {
			// Создаем карту категорий для быстрого доступа
			categoryMap := make(map[int]*models.Category, len(categories))
			for i := range categories {
				categoryMap[categories[i].ID] = &categories[i]
			}

			// Устанавливаем категории для книг
			for i := range books {
				if category, ok := categoryMap[books[i].CategoryID]; ok {
					books[i].Category = category
				}
			}
		}
	}

	// Вычисляем общее количество страниц
	totalPages := totalCount / filter.PageSize
	if totalCount%filter.PageSize > 0 {
		totalPages++
	}

	// Формируем ответ
	response := &models.BookListResponse{
		Books:      books,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}

	return response, nil
}

// Update обновляет данные книги
func (s *Service) Update(ctx context.Context, id int, input models.BookUpdate) (*models.Book, error) {
	// Получаем текущую книгу
	book, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения книги: %w", err)
	}

	// Обновляем поля книги
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
		// Проверяем существование категории
		_, err := s.categoryRepo.GetByID(ctx, *input.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки категории: %w", err)
		}
		book.CategoryID = *input.CategoryID
	}

	// Обновляем время изменения
	book.UpdatedAt = time.Now()

	// Сохраняем изменения в базе данных
	if err := s.bookRepo.Update(ctx, book); err != nil {
		return nil, fmt.Errorf("ошибка обновления книги: %w", err)
	}

	// Загружаем информацию о категории
	category, err := s.categoryRepo.GetByID(ctx, book.CategoryID)
	if err != nil {
		// Не возвращаем ошибку, так как книга уже обновлена
		return book, nil
	}
	book.Category = category

	return book, nil
}

// Delete удаляет книгу по ID
func (s *Service) Delete(ctx context.Context, id int) error {
	// Проверяем существование книги
	_, err := s.bookRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("ошибка получения книги: %w", err)
	}

	// Удаляем книгу
	if err := s.bookRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("ошибка удаления книги: %w", err)
	}

	return nil
}

// GetBooksByIDs возвращает книги по списку ID
func (s *Service) GetBooksByIDs(ctx context.Context, ids []int) ([]models.Book, error) {
	if len(ids) == 0 {
		return []models.Book{}, nil
	}

	// Получаем книги из репозитория
	books, err := s.bookRepo.GetBooksByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения книг: %w", err)
	}

	// Загружаем информацию о категориях
	if len(books) > 0 {
		categoryIDs := make([]int, 0, len(books))
		for _, book := range books {
			categoryIDs = append(categoryIDs, book.CategoryID)
		}

		categories, err := s.categoryRepo.GetCategoriesByIDs(ctx, categoryIDs)
		if err == nil {
			// Создаем карту категорий для быстрого доступа
			categoryMap := make(map[int]*models.Category, len(categories))
			for i := range categories {
				categoryMap[categories[i].ID] = &categories[i]
			}

			// Устанавливаем категории для книг
			for i := range books {
				if category, ok := categoryMap[books[i].CategoryID]; ok {
					books[i].Category = category
				}
			}
		}
	}

	return books, nil
}
