package services

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// CartService определяет методы для работы с корзиной
type CartService interface {
	// AddItem добавляет товар в корзину пользователя
	AddItem(ctx context.Context, userID int, input models.CartItemRequest) error

	// GetCart возвращает корзину пользователя
	GetCart(ctx context.Context, userID int) (*models.CartResponse, error)

	// RemoveItem удаляет товар из корзины пользователя
	RemoveItem(ctx context.Context, userID int, bookID int) error

	// ClearCart очищает корзину пользователя
	ClearCart(ctx context.Context, userID int) error

	// CleanupExpiredItems удаляет истекшие товары из корзин
	CleanupExpiredItems(ctx context.Context) error
}

