package repositories

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// OrderRepository определяет методы для работы с заказами в хранилище
type OrderRepository interface {
	// Create создает новый заказ
	Create(ctx context.Context, order *models.Order) error

	// GetByID возвращает заказ по ID
	GetByID(ctx context.Context, id int) (*models.Order, error)

	// GetByUserID возвращает список заказов пользователя
	GetByUserID(ctx context.Context, userID int) ([]models.Order, error)

	// Update обновляет статус заказа
	UpdateStatus(ctx context.Context, id int, status string) error

	// AddOrderItem добавляет товар в заказ
	AddOrderItem(ctx context.Context, orderID int, item models.OrderItem) error

	// GetOrderItems возвращает список товаров в заказе
	GetOrderItems(ctx context.Context, orderID int) ([]models.OrderItem, error)
}
