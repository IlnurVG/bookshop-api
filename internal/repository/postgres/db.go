package postgres

import (
	"context"
	"fmt"

	"github.com/bookshop/api/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresDB creates a new PostgreSQL connection pool
func NewPostgresDB(cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("error parsing database config: %w", err)
	}

	// Make sure MaxConns is not less than 1
	maxConns := int32(cfg.MaxConns)
	if maxConns < 1 {
		maxConns = 1
	}
	poolConfig.MaxConns = maxConns
	poolConfig.ConnConfig.ConnectTimeout = cfg.Timeout

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return pool, nil
}
