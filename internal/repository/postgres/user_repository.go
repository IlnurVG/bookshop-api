package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Коды ошибок PostgreSQL
const (
	// UniqueViolationCode - код ошибки нарушения уникальности
	UniqueViolationCode = "23505"
)

// UserRepository реализует интерфейс repositories.UserRepository
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository создает новый экземпляр UserRepository
func NewUserRepository(db *pgxpool.Pool) repositories.UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create создает нового пользователя
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (email, password_hash, is_admin, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRow(ctx, query,
		user.Email,
		user.PasswordHash,
		user.IsAdmin,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		// Проверяем, является ли ошибка нарушением уникальности
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == UniqueViolationCode {
			return repositories.ErrDuplicateKey
		}
		return fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	return nil
}

// GetByID возвращает пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, is_admin, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrNotFound
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	return user, nil
}

// GetByEmail возвращает пользователя по email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, is_admin, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrNotFound
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	return user, nil
}

// Update обновляет данные пользователя
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET email = $1, password_hash = $2, is_admin = $3, updated_at = $4
		WHERE id = $5
	`

	user.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		user.Email,
		user.PasswordHash,
		user.IsAdmin,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("ошибка обновления пользователя: %w", err)
	}

	return nil
}

// Delete удаляет пользователя по ID
func (r *UserRepository) Delete(ctx context.Context, id int) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления пользователя: %w", err)
	}

	return nil
}
