package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/repositories"
	repomodels "github.com/bookshop/api/internal/repository/postgres/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BookRepository implements repositories.BookRepository interface
type BookRepository struct {
	db *pgxpool.Pool
}

// NewBookRepository creates a new instance of BookRepository
func NewBookRepository(db *pgxpool.Pool) repositories.BookRepository {
	return &BookRepository{
		db: db,
	}
}

// Create creates a new book
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

	// Convert domain model to repository model
	repoBook := repomodels.FromDomain(book)

	err := r.db.QueryRow(ctx, query,
		repoBook.Title,
		repoBook.Author,
		repoBook.YearPublished,
		repoBook.Price,
		repoBook.Stock,
		repoBook.CategoryID,
		repoBook.CreatedAt,
		repoBook.UpdatedAt,
	).Scan(&repoBook.ID)

	if err != nil {
		return fmt.Errorf("failed to create book: %w", err)
	}

	// Update the domain model with generated ID
	book.ID = repoBook.ID

	return nil
}

// GetByID returns a book by ID
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

	repoBook := &repomodels.Book{}
	var categoryName string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&repoBook.ID,
		&repoBook.Title,
		&repoBook.Author,
		&repoBook.YearPublished,
		&repoBook.Price,
		&repoBook.Stock,
		&repoBook.CategoryID,
		&repoBook.CreatedAt,
		&repoBook.UpdatedAt,
		&categoryName,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get book: %w", err)
	}

	// Convert to domain model
	book := repoBook.ToDomain()

	// Add category if available
	if categoryName != "" {
		book.Category = &models.Category{
			ID:   book.CategoryID,
			Name: categoryName,
		}
	}

	return book, nil
}

// List returns a list of books with filtering and pagination
func (r *BookRepository) List(ctx context.Context, filter models.BookFilter) ([]models.Book, int, error) {
	// Base query to get books
	baseQuery := `
		FROM books b
		LEFT JOIN categories c ON b.category_id = c.id
		WHERE 1=1
	`

	// Add filtering conditions
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

	// Query to count total number of books
	countQuery := "SELECT COUNT(*) " + baseQuery + conditions
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count books: %w", err)
	}

	// Add pagination
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 10 // Default value
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

	// Query to get books with filtering and pagination
	query := `
		SELECT 
			b.id, b.title, b.author, b.year_published, b.price, 
			b.stock, b.category_id, b.created_at, b.updated_at,
			c.name as category_name
	` + baseQuery + conditions + pagination

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query books list: %w", err)
	}
	defer rows.Close()

	repoBooks := make([]repomodels.Book, 0)
	categoryNames := make(map[int]string)

	for rows.Next() {
		var book repomodels.Book
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
			return nil, 0, fmt.Errorf("failed to scan book row: %w", err)
		}

		repoBooks = append(repoBooks, book)
		if categoryName != "" {
			categoryNames[book.ID] = categoryName
		}
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating over rows: %w", err)
	}

	// Convert repository models to domain models
	domainBooks := repomodels.BookSliceToDomain(repoBooks)

	// Add category information
	for i := range domainBooks {
		bookID := domainBooks[i].ID
		if categoryName, ok := categoryNames[bookID]; ok {
			domainBooks[i].Category = &models.Category{
				ID:   domainBooks[i].CategoryID,
				Name: categoryName,
			}
		}
	}

	return domainBooks, total, nil
}

// Update updates book data
func (r *BookRepository) Update(ctx context.Context, book *models.Book) error {
	query := `
		UPDATE books 
		SET 
			title = $1, 
			author = $2, 
			year_published = $3, 
			price = $4, 
			stock = $5, 
			category_id = $6, 
			updated_at = $7
		WHERE id = $8
	`

	// Convert domain model to repository model
	repoBook := repomodels.FromDomain(book)

	_, err := r.db.Exec(ctx, query,
		repoBook.Title,
		repoBook.Author,
		repoBook.YearPublished,
		repoBook.Price,
		repoBook.Stock,
		repoBook.CategoryID,
		repoBook.UpdatedAt,
		repoBook.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update book: %w", err)
	}

	return nil
}

// Delete deletes a book by ID
func (r *BookRepository) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM books WHERE id = $1"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete book: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return repositories.ErrNotFound
	}

	return nil
}

// UpdateStock updates the quantity of books in stock
func (r *BookRepository) UpdateStock(ctx context.Context, id int, quantity int) error {
	query := `
		UPDATE books
		SET stock = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query, quantity, now, id)
	if err != nil {
		return fmt.Errorf("failed to update book stock: %w", err)
	}

	return nil
}

// GetBooksByIDs returns books by a list of IDs
func (r *BookRepository) GetBooksByIDs(ctx context.Context, ids []int) ([]models.Book, error) {
	if len(ids) == 0 {
		return []models.Book{}, nil
	}

	// Prepare the query with placeholders for IDs
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	// Build the query
	query := fmt.Sprintf(`
		SELECT 
			b.id, b.title, b.author, b.year_published, b.price, 
			b.stock, b.category_id, b.created_at, b.updated_at,
			c.name as category_name
		FROM books b
		LEFT JOIN categories c ON b.category_id = c.id
		WHERE b.id IN (%s)
	`, strings.Join(placeholders, ","))

	// Execute the query
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query books by IDs: %w", err)
	}
	defer rows.Close()

	// Parse the results
	repoBooks := make([]repomodels.Book, 0, len(ids))
	categoryNames := make(map[int]string)

	for rows.Next() {
		var book repomodels.Book
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
			return nil, fmt.Errorf("failed to scan book row: %w", err)
		}

		repoBooks = append(repoBooks, book)
		if categoryName != "" {
			categoryNames[book.ID] = categoryName
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through results: %w", err)
	}

	// Convert repository models to domain models
	domainBooks := repomodels.BookSliceToDomain(repoBooks)

	// Add category information
	for i := range domainBooks {
		bookID := domainBooks[i].ID
		if categoryName, ok := categoryNames[bookID]; ok {
			domainBooks[i].Category = &models.Category{
				ID:   domainBooks[i].CategoryID,
				Name: categoryName,
			}
		}
	}

	return domainBooks, nil
}

// ReserveBooks reserves books (decreases available quantity)
// Returns an error if any of the items is unavailable
func (r *BookRepository) ReserveBooks(ctx context.Context, bookIDs []int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
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
				return fmt.Errorf("not enough books on stock for ID: %d", bookID)
			}
			return fmt.Errorf("failed to reserve book: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ReleaseBooks returns reserved books back to stock
func (r *BookRepository) ReleaseBooks(ctx context.Context, bookIDs []int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
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
			return fmt.Errorf("failed to return book to stock: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
