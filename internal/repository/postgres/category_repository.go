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

// CategoryRepository реализует интерфейс repositories.CategoryRepository
type CategoryRepository struct {
	db *pgxpool.Pool
}

// NewCategoryRepository создает новый экземпляр CategoryRepository
func NewCategoryRepository(db *pgxpool.Pool) repositories.CategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

// Create создает новую категорию
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
		return fmt.Errorf("ошибка создания категории: %w", err)
	}

	return nil
}

// GetByID возвращает категорию по ID
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
			return nil, fmt.Errorf("категория не найдена")
		}
		return nil, fmt.Errorf("ошибка получения категории: %w", err)
	}

	return category, nil
}

// GetByName возвращает категорию по имени
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
			return nil, fmt.Errorf("категория не найдена")
		}
		return nil, fmt.Errorf("ошибка получения категории: %w", err)
	}

	return category, nil
}

// List возвращает список всех категорий
func (r *CategoryRepository) List(ctx context.Context) ([]models.Category, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM categories
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка категорий: %w", err)
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
			return nil, fmt.Errorf("ошибка сканирования данных категории: %w", err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по результатам: %w", err)
	}

	return categories, nil
}

// Update обновляет данные категории
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
		return fmt.Errorf("ошибка обновления категории: %w", err)
	}

	return nil
}

// Delete удаляет категорию по ID
func (r *CategoryRepository) Delete(ctx context.Context, id int) error {
	query := `
		DELETE FROM categories
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления категории: %w", err)
	}

	return nil
}

// GetCategoriesByIDs возвращает категории по списку ID
func (r *CategoryRepository) GetCategoriesByIDs(ctx context.Context, ids []int) ([]models.Category, error) {
	if len(ids) == 0 {
		return []models.Category{}, nil
	}

	// Создаем параметры для запроса
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
		return nil, fmt.Errorf("ошибка получения категорий по ID: %w", err)
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
			return nil, fmt.Errorf("ошибка сканирования данных категории: %w", err)
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по результатам: %w", err)
	}

	return categories, nil
}
