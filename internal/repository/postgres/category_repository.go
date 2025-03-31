package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CategoryRepository implements repositories.CategoryRepository interface
type CategoryRepository struct {
	db *pgxpool.Pool
}

// NewCategoryRepository creates a new instance of CategoryRepository
func NewCategoryRepository(db *pgxpool.Pool) repositories.CategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

// Create creates a new category
func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) error {
	query := `
		INSERT INTO categories (name, created_at, updated_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now

	err := r.db.QueryRow(ctx, query,
		category.Name,
		category.CreatedAt,
		category.UpdatedAt,
	).Scan(&category.ID)

	if err != nil {
		return fmt.Errorf("error creating category: %w", err)
	}

	return nil
}

// GetByID returns a category by ID
func (r *CategoryRepository) GetByID(ctx context.Context, id int) (*models.Category, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM categories
		WHERE id = $1
	`

	category := &models.Category{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("error getting category: %w", err)
	}

	return category, nil
}

// GetByName returns a category by name
func (r *CategoryRepository) GetByName(ctx context.Context, name string) (*models.Category, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM categories
		WHERE name = $1
	`

	category := &models.Category{}
	err := r.db.QueryRow(ctx, query, name).Scan(
		&category.ID,
		&category.Name,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("error getting category: %w", err)
	}

	return category, nil
}

// List returns a list of all categories
func (r *CategoryRepository) List(ctx context.Context) ([]models.Category, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM categories
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting categories list: %w", err)
	}
	defer rows.Close()

	categories := make([]models.Category, 0)
	for rows.Next() {
		category := models.Category{}
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning category data: %w", err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through results: %w", err)
	}

	return categories, nil
}

// Update updates category data
func (r *CategoryRepository) Update(ctx context.Context, category *models.Category) error {
	query := `
		UPDATE categories
		SET name = $1, updated_at = $2
		WHERE id = $3
	`

	category.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		category.Name,
		category.UpdatedAt,
		category.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating category: %w", err)
	}

	return nil
}

// Delete deletes a category by ID
func (r *CategoryRepository) Delete(ctx context.Context, id int) error {
	query := `
		DELETE FROM categories
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting category: %w", err)
	}

	return nil
}

// GetCategoriesByIDs returns categories by a list of IDs
func (r *CategoryRepository) GetCategoriesByIDs(ctx context.Context, ids []int) ([]models.Category, error) {
	if len(ids) == 0 {
		return []models.Category{}, nil
	}

	// Create parameters for the query
	args := make([]interface{}, len(ids))
	placeholders := make([]string, len(ids))
	for i, id := range ids {
		args[i] = id
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
		SELECT id, name, created_at, updated_at
		FROM categories
		WHERE id IN (%s)
		ORDER BY name
	`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error getting categories by IDs: %w", err)
	}
	defer rows.Close()

	categories := make([]models.Category, 0, len(ids))
	for rows.Next() {
		category := models.Category{}
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning category data: %w", err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through results: %w", err)
	}

	return categories, nil
}
