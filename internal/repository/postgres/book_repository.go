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

// BookRepository реализует интерфейс repositories.BookRepository
type BookRepository struct {
	db *pgxpool.Pool
}

// NewBookRepository создает новый экземпляр BookRepository
func NewBookRepository(db *pgxpool.Pool) repositories.BookRepository {
	return &BookRepository{
		db: db,
	}
}

// Create создает новую книгу
func (r *BookRepository) Create(ctx context.Context, book *models.Book) error {
	query := `
		INSERT INTO books (
			title, author, year_published, price, stock, 
			category_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	now := time.Now()
	book.CreatedAt = now
	book.UpdatedAt = now

	err := r.db.QueryRow(ctx, query,
		book.Title,
		book.Author,
		book.YearPublished,
		book.Price,
		book.Stock,
		book.CategoryID,
		book.CreatedAt,
		book.UpdatedAt,
	).Scan(&book.ID)

	if err != nil {
		return fmt.Errorf("ошибка создания книги: %w", err)
	}

	return nil
}

// GetByID возвращает книгу по ID
func (r *BookRepository) GetByID(ctx context.Context, id int) (*models.Book, error) {
	query := `
		SELECT 
			b.id, b.title, b.author, b.year_published, b.price, 
			b.stock, b.category_id, b.created_at, b.updated_at,
			c.name as category_name
		FROM books b
		LEFT JOIN categories c ON b.category_id = c.id
		WHERE b.id = $1
	`

	book := &models.Book{}
	var categoryName string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.YearPublished,
		&book.Price,
		&book.Stock,
		&book.CategoryID,
		&book.CreatedAt,
		&book.UpdatedAt,
		&categoryName,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("книга не найдена")
		}
		return nil, fmt.Errorf("ошибка получения книги: %w", err)
	}

	if categoryName != "" {
		book.Category = &models.Category{
			ID:   book.CategoryID,
			Name: categoryName,
		}
	}

	return book, nil
}

// List возвращает список книг с фильтрацией и пагинацией
func (r *BookRepository) List(ctx context.Context, filter models.BookFilter) ([]models.Book, int, error) {
	// Базовый запрос для получения книг
	baseQuery := `
		FROM books b
		LEFT JOIN categories c ON b.category_id = c.id
		WHERE 1=1
	`

	// Добавляем условия фильтрации
	var conditions string
	var args []interface{}
	argIndex := 1

	if len(filter.CategoryIDs) > 0 {
		conditions += " AND b.category_id IN ("
		for i, catID := range filter.CategoryIDs {
			if i > 0 {
				conditions += ","
			}
			conditions += fmt.Sprintf("$%d", argIndex)
			args = append(args, catID)
			argIndex++
		}
		conditions += ")"
	}

	if filter.MinPrice != nil && *filter.MinPrice > 0 {
		conditions += fmt.Sprintf(" AND b.price >= $%d", argIndex)
		args = append(args, *filter.MinPrice)
		argIndex++
	}

	if filter.MaxPrice != nil && *filter.MaxPrice > 0 {
		conditions += fmt.Sprintf(" AND b.price <= $%d", argIndex)
		args = append(args, *filter.MaxPrice)
		argIndex++
	}

	if filter.InStock != nil && *filter.InStock {
		conditions += " AND b.stock > 0"
	}

	// Запрос для подсчета общего количества книг
	countQuery := "SELECT COUNT(*) " + baseQuery + conditions
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка подсчета книг: %w", err)
	}

	// Добавляем пагинацию
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 10 // Значение по умолчанию
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	pagination := fmt.Sprintf(" ORDER BY b.id DESC LIMIT %d OFFSET %d", pageSize, offset)

	// Запрос для получения книг с учетом фильтрации и пагинации
	query := `
		SELECT 
			b.id, b.title, b.author, b.year_published, b.price, 
			b.stock, b.category_id, b.created_at, b.updated_at,
			c.name as category_name
	` + baseQuery + conditions + pagination

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения списка книг: %w", err)
	}
	defer rows.Close()

	books := make([]models.Book, 0)
	for rows.Next() {
		book := models.Book{}
		var categoryName string
		err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Author,
			&book.YearPublished,
			&book.Price,
			&book.Stock,
			&book.CategoryID,
			&book.CreatedAt,
			&book.UpdatedAt,
			&categoryName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("ошибка сканирования данных книги: %w", err)
		}

		if categoryName != "" {
			book.Category = &models.Category{
				ID:   book.CategoryID,
				Name: categoryName,
			}
		}

		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("ошибка при итерации по результатам: %w", err)
	}

	return books, total, nil
}

// Update обновляет данные книги
func (r *BookRepository) Update(ctx context.Context, book *models.Book) error {
	query := `
		UPDATE books
		SET title = $1, author = $2, year_published = $3, price = $4, 
			stock = $5, category_id = $6, updated_at = $7
		WHERE id = $8
	`

	book.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		book.Title,
		book.Author,
		book.YearPublished,
		book.Price,
		book.Stock,
		book.CategoryID,
		book.UpdatedAt,
		book.ID,
	)

	if err != nil {
		return fmt.Errorf("ошибка обновления книги: %w", err)
	}

	return nil
}

// Delete удаляет книгу по ID
func (r *BookRepository) Delete(ctx context.Context, id int) error {
	query := `
		DELETE FROM books
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления книги: %w", err)
	}

	return nil
}

// UpdateStock обновляет количество книг на складе
func (r *BookRepository) UpdateStock(ctx context.Context, id int, quantity int) error {
	query := `
		UPDATE books
		SET stock = stock + $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, quantity, time.Now(), id)
	if err != nil {
		return fmt.Errorf("ошибка обновления количества книг: %w", err)
	}

	return nil
}

// GetBooksByIDs возвращает книги по списку ID
func (r *BookRepository) GetBooksByIDs(ctx context.Context, ids []int) ([]models.Book, error) {
	if len(ids) == 0 {
		return []models.Book{}, nil
	}

	// Создаем параметры для запроса
	args := make([]interface{}, len(ids))
	placeholders := make([]string, len(ids))
	for i, id := range ids {
		args[i] = id
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
		SELECT 
			b.id, b.title, b.author, b.year_published, b.price, 
			b.stock, b.category_id, b.created_at, b.updated_at,
			c.name as category_name
		FROM books b
		LEFT JOIN categories c ON b.category_id = c.id
		WHERE b.id IN (%s)
	`, fmt.Sprintf(strings.Join(placeholders, ",")))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения книг по ID: %w", err)
	}
	defer rows.Close()

	books := make([]models.Book, 0, len(ids))
	for rows.Next() {
		book := models.Book{}
		var categoryName string
		err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Author,
			&book.YearPublished,
			&book.Price,
			&book.Stock,
			&book.CategoryID,
			&book.CreatedAt,
			&book.UpdatedAt,
			&categoryName,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных книги: %w", err)
		}

		if categoryName != "" {
			book.Category = &models.Category{
				ID:   book.CategoryID,
				Name: categoryName,
			}
		}

		books = append(books, book)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по результатам: %w", err)
	}

	return books, nil
}

// ReserveBooks резервирует указанное количество книг
func (r *BookRepository) ReserveBooks(ctx context.Context, bookIDs []int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, bookID := range bookIDs {
		query := `
			UPDATE books
			SET stock = stock - 1, updated_at = $1
			WHERE id = $2 AND stock > 0
			RETURNING stock
		`

		var remainingStock int
		err := tx.QueryRow(ctx, query, time.Now(), bookID).Scan(&remainingStock)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("недостаточно книг на складе для ID: %d", bookID)
			}
			return fmt.Errorf("ошибка резервирования книги: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("ошибка фиксации транзакции: %w", err)
	}

	return nil
}

// ReleaseBooks возвращает зарезервированные книги обратно на склад
func (r *BookRepository) ReleaseBooks(ctx context.Context, bookIDs []int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, bookID := range bookIDs {
		query := `
			UPDATE books
			SET stock = stock + 1, updated_at = $1
			WHERE id = $2
		`

		_, err := tx.Exec(ctx, query, time.Now(), bookID)
		if err != nil {
			return fmt.Errorf("ошибка возврата книги на склад: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("ошибка фиксации транзакции: %w", err)
	}

	return nil
}
