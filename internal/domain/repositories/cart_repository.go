package repositories

import (
	"context"
	"time"

	"github.com/bookshop/api/internal/domain/models"
)

// CartRepository определяет методы для работы с корзиной в хранилище
type CartRepository interface {
	// AddItem добавляет товар в корзину пользователя
	AddItem(ctx context.Context, userID int, bookID int, expiresAt time.Time) error

	// GetCart возвращает корзину пользователя
	GetCart(ctx context.Context, userID int) (*models.Cart, error)

	// RemoveItem удаляет товар из корзины пользователя
	RemoveItem(ctx context.Context, userID int, bookID int) error

	// ClearCart очищает корзину пользователя
	ClearCart(ctx context.Context, userID int) error

	// GetExpiredCarts возвращает список истекших корзин
	GetExpiredCarts(ctx context.Context) ([]models.Cart, error)

	// RemoveExpiredItems удаляет истекшие товары из корзин
	RemoveExpiredItems(ctx context.Context) error

	// LockCart блокирует корзину на время оформления заказа
	// Возвращает ошибку, если корзина уже заблокирована
	LockCart(ctx context.Context, userID int, duration time.Duration) error

	// UnlockCart разблокирует корзину
	UnlockCart(ctx context.Context, userID int) error
}
