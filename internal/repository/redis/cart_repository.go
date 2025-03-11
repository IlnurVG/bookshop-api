package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/redis/go-redis/v9"
)

const (
	// cartKeyPrefix префикс для ключей корзины
	cartKeyPrefix = "cart:"
	// cartLockKeyPrefix префикс для ключей блокировки корзины
	cartLockKeyPrefix = "cart_lock:"
)

// CartRepository реализует интерфейс repositories.CartRepository
type CartRepository struct {
	client *redis.Client
}

// NewCartRepository создает новый экземпляр репозитория корзины
func NewCartRepository(client *redis.Client) *CartRepository {
	return &CartRepository{
		client: client,
	}
}

// AddItem добавляет товар в корзину пользователя
func (r *CartRepository) AddItem(ctx context.Context, userID int, bookID int, expiresAt time.Time) error {
	// Проверяем блокировку корзины
	if r.isCartLocked(ctx, userID) {
		return fmt.Errorf("корзина заблокирована")
	}

	// Создаем элемент корзины
	item := models.CartItem{
		BookID:    bookID,
		AddedAt:   time.Now(),
		ExpiresAt: expiresAt,
	}

	// Сериализуем элемент
	itemJSON, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("ошибка сериализации элемента корзины: %w", err)
	}

	// Добавляем элемент в корзину
	key := fmt.Sprintf("%s%d", cartKeyPrefix, userID)
	if err := r.client.HSet(ctx, key, fmt.Sprint(bookID), itemJSON).Err(); err != nil {
		return fmt.Errorf("ошибка добавления элемента в корзину: %w", err)
	}

	return nil
}

// GetCart возвращает корзину пользователя
func (r *CartRepository) GetCart(ctx context.Context, userID int) (*models.Cart, error) {
	// Получаем все элементы корзины
	key := fmt.Sprintf("%s%d", cartKeyPrefix, userID)
	items, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения корзины: %w", err)
	}

	// Создаем корзину
	cart := &models.Cart{
		UserID: userID,
		Items:  make([]models.CartItem, 0, len(items)),
	}

	// Десериализуем элементы
	for _, itemJSON := range items {
		var item models.CartItem
		if err := json.Unmarshal([]byte(itemJSON), &item); err != nil {
			return nil, fmt.Errorf("ошибка десериализации элемента корзины: %w", err)
		}

		// Проверяем срок действия элемента
		if time.Now().After(item.ExpiresAt) {
			continue
		}

		cart.Items = append(cart.Items, item)
	}

	return cart, nil
}

// RemoveItem удаляет товар из корзины пользователя
func (r *CartRepository) RemoveItem(ctx context.Context, userID int, bookID int) error {
	// Проверяем блокировку корзины
	if r.isCartLocked(ctx, userID) {
		return fmt.Errorf("корзина заблокирована")
	}

	// Удаляем элемент из корзины
	key := fmt.Sprintf("%s%d", cartKeyPrefix, userID)
	if err := r.client.HDel(ctx, key, fmt.Sprint(bookID)).Err(); err != nil {
		return fmt.Errorf("ошибка удаления элемента из корзины: %w", err)
	}

	return nil
}

// ClearCart очищает корзину пользователя
func (r *CartRepository) ClearCart(ctx context.Context, userID int) error {
	// Проверяем блокировку корзины
	if r.isCartLocked(ctx, userID) {
		return fmt.Errorf("корзина заблокирована")
	}

	// Удаляем корзину
	key := fmt.Sprintf("%s%d", cartKeyPrefix, userID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("ошибка очистки корзины: %w", err)
	}

	return nil
}

// RemoveExpiredItems удаляет истекшие товары из корзин
func (r *CartRepository) RemoveExpiredItems(ctx context.Context) error {
	// Получаем все корзины
	pattern := fmt.Sprintf("%s*", cartKeyPrefix)
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()

		// Получаем все элементы корзины
		items, err := r.client.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}

		// Проверяем каждый элемент
		for bookID, itemJSON := range items {
			var item models.CartItem
			if err := json.Unmarshal([]byte(itemJSON), &item); err != nil {
				continue
			}

			// Удаляем истекшие элементы
			if time.Now().After(item.ExpiresAt) {
				r.client.HDel(ctx, key, bookID)
			}
		}
	}

	return iter.Err()
}

// LockCart блокирует корзину на время оформления заказа
func (r *CartRepository) LockCart(ctx context.Context, userID int, duration time.Duration) error {
	// Проверяем, не заблокирована ли уже корзина
	if r.isCartLocked(ctx, userID) {
		return fmt.Errorf("корзина уже заблокирована")
	}

	// Блокируем корзину
	key := fmt.Sprintf("%s%d", cartLockKeyPrefix, userID)
	if err := r.client.Set(ctx, key, "locked", duration).Err(); err != nil {
		return fmt.Errorf("ошибка блокировки корзины: %w", err)
	}

	return nil
}

// UnlockCart разблокирует корзину
func (r *CartRepository) UnlockCart(ctx context.Context, userID int) error {
	// Удаляем блокировку
	key := fmt.Sprintf("%s%d", cartLockKeyPrefix, userID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("ошибка разблокировки корзины: %w", err)
	}

	return nil
}

// isCartLocked проверяет, заблокирована ли корзина
func (r *CartRepository) isCartLocked(ctx context.Context, userID int) bool {
	key := fmt.Sprintf("%s%d", cartLockKeyPrefix, userID)
	exists, _ := r.client.Exists(ctx, key).Result()
	return exists == 1
}
