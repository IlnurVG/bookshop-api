package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// LockManager управляет блокировками в Redis
type LockManager struct {
	client *redis.Client
}

// NewLockManager создает новый экземпляр менеджера блокировок
func NewLockManager(client *redis.Client) *LockManager {
	return &LockManager{
		client: client,
	}
}

// Lock блокирует ресурс на указанное время
func (m *LockManager) Lock(ctx context.Context, key string, duration time.Duration) error {
	// Пытаемся установить блокировку
	ok, err := m.client.SetNX(ctx, key, "locked", duration).Result()
	if err != nil {
		return fmt.Errorf("ошибка установки блокировки: %w", err)
	}

	// Проверяем, удалось ли установить блокировку
	if !ok {
		return fmt.Errorf("ресурс уже заблокирован")
	}

	return nil
}

// Unlock разблокирует ресурс
func (m *LockManager) Unlock(ctx context.Context, key string) error {
	// Удаляем блокировку
	if err := m.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("ошибка удаления блокировки: %w", err)
	}

	return nil
}

// IsLocked проверяет, заблокирован ли ресурс
func (m *LockManager) IsLocked(ctx context.Context, key string) (bool, error) {
	// Проверяем существование блокировки
	exists, err := m.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("ошибка проверки блокировки: %w", err)
	}

	return exists == 1, nil
}

// ExtendLock продлевает время блокировки
func (m *LockManager) ExtendLock(ctx context.Context, key string, duration time.Duration) error {
	// Проверяем существование блокировки
	exists, err := m.IsLocked(ctx, key)
	if err != nil {
		return err
	}

	// Если блокировка не существует, возвращаем ошибку
	if !exists {
		return fmt.Errorf("блокировка не существует")
	}

	// Продлеваем время блокировки
	if err := m.client.Expire(ctx, key, duration).Err(); err != nil {
		return fmt.Errorf("ошибка продления блокировки: %w", err)
	}

	return nil
}
