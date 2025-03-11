package redis

import (
	"context"
	"fmt"

	"github.com/bookshop/api/config"
	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("error connecting to Redis: %w", err)
	}

	return client, nil
}
