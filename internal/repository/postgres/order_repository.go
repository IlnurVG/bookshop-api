package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OrderRepository реализует интерфейс repositories.OrderRepository
type OrderRepository struct {
	db *pgxpool.Pool
}

// NewOrderRepository создает новый экземпляр OrderRepository
func NewOrderRepository(db *pgxpool.Pool) repositories.OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

// Create создает новый заказ
func (r *OrderRepository) Create(ctx context.Context, order *models.Order) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO orders (user_id, status, total_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now

	err = tx.QueryRow(ctx, query,
		order.UserID,
		order.Status,
		order.TotalPrice,
		order.CreatedAt,
		order.UpdatedAt,
	).Scan(&order.ID)

	if err != nil {
		return fmt.Errorf("ошибка создания заказа: %w", err)
	}

	// Добавляем товары в заказ
	for i := range order.Items {
		item := &order.Items[i]
		item.OrderID = order.ID
		item.CreatedAt = now

		itemQuery := `
			INSERT INTO order_items (order_id, book_id, price, created_at)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`

		err = tx.QueryRow(ctx, itemQuery,
			item.OrderID,
			item.BookID,
			item.Price,
			item.CreatedAt,
		).Scan(&item.ID)

		if err != nil {
			return fmt.Errorf("ошибка добавления товара в заказ: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("ошибка фиксации транзакции: %w", err)
	}

	return nil
}

// GetByID возвращает заказ по ID
func (r *OrderRepository) GetByID(ctx context.Context, id int) (*models.Order, error) {
	query := `
		SELECT id, user_id, status, total_price, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	order := &models.Order{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.TotalPrice,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("заказ не найден")
		}
		return nil, fmt.Errorf("ошибка получения заказа: %w", err)
	}

	// Получаем товары заказа
	items, err := r.GetOrderItems(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	order.Items = items
	return order, nil
}

// GetByUserID возвращает список заказов пользователя
func (r *OrderRepository) GetByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	query := `
		SELECT id, user_id, status, total_price, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения заказов пользователя: %w", err)
	}
	defer rows.Close()

	orders := make([]models.Order, 0)
	for rows.Next() {
		order := models.Order{}
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.TotalPrice,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных заказа: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по результатам: %w", err)
	}

	// Получаем товары для каждого заказа
	for i := range orders {
		items, err := r.GetOrderItems(ctx, orders[i].ID)
		if err != nil {
			return nil, err
		}
		orders[i].Items = items
	}

	return orders, nil
}

// UpdateStatus обновляет статус заказа
func (r *OrderRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса заказа: %w", err)
	}

	return nil
}

// AddOrderItem добавляет товар в заказ
func (r *OrderRepository) AddOrderItem(ctx context.Context, orderID int, item models.OrderItem) error {
	query := `
		INSERT INTO order_items (order_id, book_id, price, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	item.OrderID = orderID
	item.CreatedAt = time.Now()

	err := r.db.QueryRow(ctx, query,
		item.OrderID,
		item.BookID,
		item.Price,
		item.CreatedAt,
	).Scan(&item.ID)

	if err != nil {
		return fmt.Errorf("ошибка добавления товара в заказ: %w", err)
	}

	// Обновляем общую стоимость заказа
	updateQuery := `
		UPDATE orders
		SET total_price = total_price + $1, updated_at = $2
		WHERE id = $3
	`

	_, err = r.db.Exec(ctx, updateQuery, item.Price, time.Now(), orderID)
	if err != nil {
		return fmt.Errorf("ошибка обновления стоимости заказа: %w", err)
	}

	return nil
}

// GetOrderItems возвращает список товаров в заказе
func (r *OrderRepository) GetOrderItems(ctx context.Context, orderID int) ([]models.OrderItem, error) {
	query := `
		SELECT id, order_id, book_id, price, created_at
		FROM order_items
		WHERE order_id = $1
	`

	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения товаров заказа: %w", err)
	}
	defer rows.Close()

	items := make([]models.OrderItem, 0)
	for rows.Next() {
		item := models.OrderItem{}
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.BookID,
			&item.Price,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования данных товара: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по результатам: %w", err)
	}

	return items, nil
}
