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

// Error definitions
var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrCategoryExists   = errors.New("category with this name already exists")
)

// Service implements services.CategoryService interface
type Service struct {
	categoryRepo repositories.CategoryRepository
	logger       logger.Logger
}

// NewService creates a new instance of the category service
func NewService(
	categoryRepo repositories.CategoryRepository,
	logger logger.Logger,
) services.CategoryService {
	return &Service{
		categoryRepo: categoryRepo,
		logger:       logger,
	}
}

// Create creates a new category
func (s *Service) Create(ctx context.Context, input models.CategoryCreate) (*models.Category, error) {
	// Check if a category with this name already exists
	if _, err := s.categoryRepo.GetByName(ctx, input.Name); err == nil {
		return nil, ErrCategoryExists
	} else if !errors.Is(err, repositories.ErrNotFound) {
		return nil, fmt.Errorf("error checking if category exists: %w", err)
	}

	// Create category
	category := &models.Category{
		Name:      input.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save category
	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("error creating category: %w", err)
	}

	return category, nil
}

// GetByID returns a category by ID
func (s *Service) GetByID(ctx context.Context, id int) (*models.Category, error) {
	// Get category
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("error getting category: %w", err)
	}

	return category, nil
}

// List returns a list of all categories
func (s *Service) List(ctx context.Context) ([]models.Category, error) {
	// Get list of categories
	categories, err := s.categoryRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting category list: %w", err)
	}

	return categories, nil
}

// Update updates a category
func (s *Service) Update(ctx context.Context, id int, input models.CategoryUpdate) (*models.Category, error) {
	// Get category
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("error getting category: %w", err)
	}

	// Check if another category with this name exists
	if existing, err := s.categoryRepo.GetByName(ctx, input.Name); err == nil && existing.ID != id {
		return nil, ErrCategoryExists
	} else if err != nil && !errors.Is(err, repositories.ErrNotFound) {
		return nil, fmt.Errorf("error checking if category exists: %w", err)
	}

	// Update data
	category.Name = input.Name
	category.UpdatedAt = time.Now()

	// Save changes
	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("error updating category: %w", err)
	}

	return category, nil
}

// Delete deletes a category
func (s *Service) Delete(ctx context.Context, id int) error {
	// Check if the category exists
	if _, err := s.categoryRepo.GetByID(ctx, id); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return ErrCategoryNotFound
		}
		return fmt.Errorf("error getting category: %w", err)
	}

	// Delete category
	if err := s.categoryRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("error deleting category: %w", err)
	}

	return nil
}
