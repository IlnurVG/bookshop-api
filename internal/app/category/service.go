package category

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/bookshop/api/pkg/logger"
)

// Определение ошибок
var (
	ErrCategoryNotFound = errors.New("категория не найдена")
	ErrCategoryExists   = errors.New("категория с таким именем уже существует")
)

// Service реализует интерфейс services.CategoryService
type Service struct {
	categoryRepo repositories.CategoryRepository
	logger       logger.Logger
}

// NewService создает новый экземпляр сервиса категорий
func NewService(
	categoryRepo repositories.CategoryRepository,
	logger logger.Logger,
) services.CategoryService {
	return &Service{
		categoryRepo: categoryRepo,
		logger:       logger,
	}
}

// Create создает новую категорию
func (s *Service) Create(ctx context.Context, input models.CategoryCreate) (*models.Category, error) {
	// Проверяем существование категории с таким именем
	if _, err := s.categoryRepo.GetByName(ctx, input.Name); err == nil {
		return nil, ErrCategoryExists
	} else if !errors.Is(err, repositories.ErrNotFound) {
		return nil, fmt.Errorf("ошибка проверки существования категории: %w", err)
	}

	// Создаем категорию
	category := &models.Category{
		Name:      input.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Сохраняем категорию
	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("ошибка создания категории: %w", err)
	}

	return category, nil
}

// GetByID возвращает категорию по ID
func (s *Service) GetByID(ctx context.Context, id int) (*models.Category, error) {
	// Получаем категорию
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("ошибка получения категории: %w", err)
	}

	return category, nil
}

// List возвращает список всех категорий
func (s *Service) List(ctx context.Context) ([]models.Category, error) {
	// Получаем список категорий
	categories, err := s.categoryRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка категорий: %w", err)
	}

	return categories, nil
}

// Update обновляет категорию
func (s *Service) Update(ctx context.Context, id int, input models.CategoryUpdate) (*models.Category, error) {
	// Получаем категорию
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("ошибка получения категории: %w", err)
	}

	// Проверяем, не существует ли другая категория с таким именем
	if existing, err := s.categoryRepo.GetByName(ctx, input.Name); err == nil && existing.ID != id {
		return nil, ErrCategoryExists
	} else if err != nil && !errors.Is(err, repositories.ErrNotFound) {
		return nil, fmt.Errorf("ошибка проверки существования категории: %w", err)
	}

	// Обновляем данные
	category.Name = input.Name
	category.UpdatedAt = time.Now()

	// Сохраняем изменения
	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("ошибка обновления категории: %w", err)
	}

	return category, nil
}

// Delete удаляет категорию
func (s *Service) Delete(ctx context.Context, id int) error {
	// Проверяем существование категории
	if _, err := s.categoryRepo.GetByID(ctx, id); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return ErrCategoryNotFound
		}
		return fmt.Errorf("ошибка получения категории: %w", err)
	}

	// Удаляем категорию
	if err := s.categoryRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("ошибка удаления категории: %w", err)
	}

	return nil
}
