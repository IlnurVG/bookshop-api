package services

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// CheckoutService определяет методы для оформления заказа
type CheckoutService interface {
	// Checkout оформляет заказ из корзины пользователя
	Checkout(ctx context.Context, userID int) (*models.Order, error)

	// GetOrdersByUserID возвращает список заказов пользователя
	GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error)

	// GetOrderByID возвращает заказ по ID
	GetOrderByID(ctx context.Context, orderID int, userID int) (*models.Order, error)
}

