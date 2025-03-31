package cache

import (
	"context"
	"time"

	"github.com/bookshop/api/internal/pkg/workerpool"
	"github.com/bookshop/api/pkg/logger"
)

// CacheOperation represents a type of cache operation
type CacheOperation string

const (
	// OperationSet sets a value in the cache
	OperationSet CacheOperation = "set"
	// OperationDelete deletes a value from the cache
	OperationDelete CacheOperation = "delete"
	// OperationUpdate updates a value in the cache
	OperationUpdate CacheOperation = "update"
)

// CacheTask represents a task to be executed by the cache worker
type CacheTask struct {
	Operation  CacheOperation
	ProfileID  string
	Data       interface{}
	OrderID    string // Optional, only for order operations
	ResultChan chan error
}

// CacheWorker handles asynchronous cache operations
type CacheWorker struct {
	profileCache ProfileCacher // Используем интерфейс вместо конкретного типа
	redisClient  interface{}   // Redis client interface
	workerPool   *workerpool.WorkerPool
	logger       logger.Logger
	redisKeyFn   func(string) string // Function to generate Redis keys
}

// NewCacheWorker creates a new cache worker
func NewCacheWorker(
	profileCache ProfileCacher, // Используем интерфейс вместо конкретного типа
	redisClient interface{},
	logger logger.Logger,
	numWorkers int,
	redisKeyFn func(string) string,
) *CacheWorker {
	return &CacheWorker{
		profileCache: profileCache,
		redisClient:  redisClient,
		workerPool:   workerpool.New(numWorkers),
		logger:       logger,
		redisKeyFn:   redisKeyFn,
	}
}

// ProcessTask submits a cache operation task to the worker pool
func (cw *CacheWorker) ProcessTask(ctx context.Context, task CacheTask) chan error {
	if task.ResultChan == nil {
		task.ResultChan = make(chan error, 1)
	}

	// Submit task to worker pool
	cw.workerPool.Submit(func(ctx context.Context) error {
		err := cw.processTask(ctx, task)
		task.ResultChan <- err
		close(task.ResultChan)
		return err
	})

	return task.ResultChan
}

// processTask processes a cache operation task
func (cw *CacheWorker) processTask(ctx context.Context, task CacheTask) error {
	cw.logger.Debug("Processing cache task", "operation", task.Operation, "profileID", task.ProfileID)

	start := time.Now()
	var err error

	switch task.Operation {
	case OperationSet:
		err = cw.setOperation(ctx, task)
	case OperationDelete:
		err = cw.deleteOperation(ctx, task)
	case OperationUpdate:
		err = cw.updateOperation(ctx, task)
	default:
		cw.logger.Error("Unknown cache operation", "operation", task.Operation)
		return nil
	}

	if err != nil {
		cw.logger.Error("Cache operation failed",
			"operation", task.Operation,
			"profileID", task.ProfileID,
			"error", err,
			"duration", time.Since(start))
		return err
	}

	cw.logger.Debug("Cache operation completed",
		"operation", task.Operation,
		"profileID", task.ProfileID,
		"duration", time.Since(start))
	return nil
}

// setOperation handles setting a new value in both caches
func (cw *CacheWorker) setOperation(ctx context.Context, task CacheTask) error {
	// Set the value in the memory cache (L1)
	if profile, ok := task.Data.(*Profile); ok {
		cw.profileCache.Set(profile)
	}

	// Set the value in Redis (L2) if we have data and Redis client
	if task.Data != nil && cw.redisClient != nil {
		if client, ok := cw.redisClient.(interface {
			Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
		}); ok {
			redisKey := cw.redisKeyFn(task.ProfileID)
			return client.Set(ctx, redisKey, task.Data, 5*time.Minute)
		}
	}

	return nil
}

// deleteOperation handles deleting a value from both caches
func (cw *CacheWorker) deleteOperation(ctx context.Context, task CacheTask) error {
	// Delete from memory cache (L1)
	cw.profileCache.Delete(task.ProfileID)

	// Delete from Redis (L2) if we have Redis client
	if cw.redisClient != nil {
		if client, ok := cw.redisClient.(interface {
			Del(ctx context.Context, keys ...string) error
		}); ok {
			redisKey := cw.redisKeyFn(task.ProfileID)
			return client.Del(ctx, redisKey)
		}
	}

	return nil
}

// updateOperation handles updating a specific value in both caches
func (cw *CacheWorker) updateOperation(ctx context.Context, task CacheTask) error {
	// Check if we have an order ID for this update
	if task.OrderID == "" {
		return cw.setOperation(ctx, task) // If not, just treat it as a set operation
	}

	// Update the order in memory cache (L1) if it exists
	if order, ok := task.Data.(*Order); ok {
		cw.profileCache.UpdateOrder(task.ProfileID, task.OrderID, order)
	}

	// For Redis (L2), we'll need to get the entire profile, update it, and set it back
	if cw.redisClient != nil {
		// Type assertion for Redis client with necessary methods
		if client, ok := cw.redisClient.(interface {
			Get(ctx context.Context, key string) ([]byte, error)
			Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
			Del(ctx context.Context, keys ...string) error
		}); ok {
			redisKey := cw.redisKeyFn(task.ProfileID)
			// Try to get existing data
			data, err := client.Get(ctx, redisKey)
			if err == nil && len(data) > 0 {
				// We have existing data, we would need to deserialize,
				// update the specific order, and serialize back
				// For simplicity here, we're just invalidating the Redis cache
				return client.Del(ctx, redisKey)
			}
		}
	}

	return nil
}

// Shutdown stops the worker pool
func (cw *CacheWorker) Shutdown() {
	cw.workerPool.Shutdown()
	cw.logger.Info("Cache worker pool shutdown completed")
}

// Stop is an alias for Shutdown for interface consistency
func (cw *CacheWorker) Stop() {
	cw.Shutdown()
}
