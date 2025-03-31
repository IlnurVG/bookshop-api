package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// LockManager manages locks in Redis
type LockManager struct {
	client *redis.Client
}

// NewLockManager creates a new instance of the lock manager
func NewLockManager(client *redis.Client) *LockManager {
	return &LockManager{
		client: client,
	}
}

// Lock locks a resource for the specified duration
func (m *LockManager) Lock(ctx context.Context, key string, duration time.Duration) error {
	// Try to set the lock
	ok, err := m.client.SetNX(ctx, key, "locked", duration).Result()
	if err != nil {
		return fmt.Errorf("error setting lock: %w", err)
	}

	// Check if the lock was successfully set
	if !ok {
		return fmt.Errorf("resource is already locked")
	}

	return nil
}

// Unlock releases a resource lock
func (m *LockManager) Unlock(ctx context.Context, key string) error {
	// Remove the lock
	if err := m.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("error removing lock: %w", err)
	}

	return nil
}

// IsLocked checks if a resource is locked
func (m *LockManager) IsLocked(ctx context.Context, key string) (bool, error) {
	// Check if the lock exists
	exists, err := m.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("error checking lock: %w", err)
	}

	return exists == 1, nil
}

// ExtendLock extends the lock duration
func (m *LockManager) ExtendLock(ctx context.Context, key string, duration time.Duration) error {
	// Check if the lock exists
	exists, err := m.IsLocked(ctx, key)
	if err != nil {
		return err
	}

	// If the lock doesn't exist, return an error
	if !exists {
		return fmt.Errorf("lock does not exist")
	}

	// Extend the lock duration
	if err := m.client.Expire(ctx, key, duration).Err(); err != nil {
		return fmt.Errorf("error extending lock: %w", err)
	}

	return nil
}
