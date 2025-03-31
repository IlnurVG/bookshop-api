package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/bookshop/api/internal/pkg/cache"
	"github.com/bookshop/api/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// ProfileCacheService provides user profile caching functionality with L1 and L2 caches
type ProfileCacheService struct {
	userRepo    repositories.UserRepository
	orderRepo   repositories.OrderRepository
	redisClient *redis.Client
	cache       *cache.ProfileCache
	cacheWorker *cache.CacheWorker // Worker pool for cache operations
	logger      logger.Logger
}

// NewProfileCacheService creates a new profile cache service
func NewProfileCacheService(
	userRepo repositories.UserRepository,
	orderRepo repositories.OrderRepository,
	redisClient *redis.Client,
	logger logger.Logger,
) *ProfileCacheService {
	// Create in-memory cache with 2-second TTL and 1-second cleanup interval
	profileCache := cache.NewProfileCache(2*time.Second, 1*time.Second)

	// Create cache worker with 3 workers
	cacheWorker := cache.NewCacheWorker(
		profileCache,
		redisClient,
		logger,
		3, // Number of workers
		func(userID string) string {
			return fmt.Sprintf("user_profile:%s", userID)
		},
	)

	return &ProfileCacheService{
		userRepo:    userRepo,
		orderRepo:   orderRepo,
		redisClient: redisClient,
		cache:       profileCache,
		cacheWorker: cacheWorker,
		logger:      logger,
	}
}

// GetUserWithOrders gets a user with their orders, using the caching system
func (s *ProfileCacheService) GetUserWithOrders(ctx context.Context, userID int) (*models.User, []models.Order, error) {
	// Generate string ID for cache lookups
	userIDStr := strconv.Itoa(userID)

	// Try L1 cache first (in-memory)
	cachedProfile := s.cache.Get(userIDStr)
	if cachedProfile != nil {
		s.logger.Debug("Profile retrieved from L1 cache", "userID", userIDStr)

		// Convert from cache format to domain models
		user, orders, err := s.convertCacheToModels(cachedProfile)
		if err == nil {
			return user, orders, nil
		}

		s.logger.Debug("Error converting cached profile, fetching fresh data", "error", err)
	}

	// Try L2 cache (Redis)
	redisKey := fmt.Sprintf("user_profile:%d", userID)
	data, err := s.redisClient.Get(ctx, redisKey).Bytes()

	if err == nil && len(data) > 0 {
		s.logger.Debug("Profile retrieved from Redis cache", "userID", userIDStr)

		var cacheData struct {
			User   models.User
			Orders []models.Order
		}

		if err := json.Unmarshal(data, &cacheData); err == nil {
			// Save to L1 cache for future requests - use worker pool
			go s.saveToL1Cache(&cacheData.User, cacheData.Orders)
			return &cacheData.User, cacheData.Orders, nil
		}

		s.logger.Debug("Error unmarshaling Redis data, fetching fresh data", "error", err)
	}

	// Fetch fresh data from the database
	s.logger.Debug("Profile not found in cache, fetching from database", "userID", userIDStr)

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("error fetching user: %w", err)
	}

	// Get orders
	orders, err := s.orderRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("error fetching orders: %w", err)
	}

	// Save to caches using worker pool
	task := cache.CacheTask{
		Operation: cache.OperationSet,
		ProfileID: userIDStr,
		Data: struct {
			User   *models.User
			Orders []models.Order
		}{
			User:   user,
			Orders: orders,
		},
	}

	// Submit task to worker pool but don't wait for completion
	s.cacheWorker.ProcessTask(ctx, task)

	return user, orders, nil
}

// InvalidateUserCache invalidates all cache entries for a user
func (s *ProfileCacheService) InvalidateUserCache(ctx context.Context, userID int) {
	userIDStr := strconv.Itoa(userID)

	// Create delete task for worker pool
	task := cache.CacheTask{
		Operation: cache.OperationDelete,
		ProfileID: userIDStr,
	}

	// Submit task to worker pool but don't wait for completion
	s.cacheWorker.ProcessTask(ctx, task)

	s.logger.Debug("User cache invalidation submitted to worker pool", "userID", userIDStr)
}

// UpdateUserCache updates user's cache when their data changes
func (s *ProfileCacheService) UpdateUserCache(ctx context.Context, user *models.User, orders []models.Order) {
	userIDStr := strconv.Itoa(user.ID)

	// Create update task for worker pool
	task := cache.CacheTask{
		Operation: cache.OperationSet,
		ProfileID: userIDStr,
		Data: struct {
			User   *models.User
			Orders []models.Order
		}{
			User:   user,
			Orders: orders,
		},
	}

	// Submit task to worker pool but don't wait for completion
	s.cacheWorker.ProcessTask(ctx, task)

	s.logger.Debug("User cache update submitted to worker pool", "userID", userIDStr)
}

// UpdateOrderInCache updates a specific order in the cache
func (s *ProfileCacheService) UpdateOrderInCache(ctx context.Context, userID int, order *models.Order) {
	userIDStr := strconv.Itoa(userID)
	orderIDStr := strconv.Itoa(order.ID)

	// Create update task for worker pool
	task := cache.CacheTask{
		Operation: cache.OperationUpdate,
		ProfileID: userIDStr,
		OrderID:   orderIDStr,
		Data:      order,
	}

	// Submit task to worker pool but don't wait for completion
	s.cacheWorker.ProcessTask(ctx, task)

	s.logger.Debug("Order cache update submitted to worker pool", "userID", userIDStr, "orderID", orderIDStr)
}

// convertCacheToModels converts a cached profile to domain models
func (s *ProfileCacheService) convertCacheToModels(profile *cache.Profile) (*models.User, []models.Order, error) {
	// Convert user
	userID, err := strconv.Atoi(profile.UUID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid user ID in cache: %w", err)
	}

	user := &models.User{
		ID:    userID,
		Email: profile.Name, // Use Name field to store email
	}

	// Convert orders
	orders := make([]models.Order, 0, len(profile.Orders))
	for _, cachedOrder := range profile.Orders {
		// Try to convert the Value field to an Order
		if orderData, ok := cachedOrder.Value.(map[string]interface{}); ok {
			var order models.Order

			// Extract basic order data
			if id, ok := orderData["id"].(float64); ok {
				order.ID = int(id)
			}

			if status, ok := orderData["status"].(string); ok {
				order.Status = status
			}

			if price, ok := orderData["total_price"].(float64); ok {
				order.TotalPrice = price
			}

			if createdAtStr, ok := orderData["created_at"].(string); ok {
				order.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
			}

			order.UserID = userID
			orders = append(orders, order)
		}
	}

	return user, orders, nil
}

// saveToL1Cache saves user and orders to the L1 (in-memory) cache
func (s *ProfileCacheService) saveToL1Cache(user *models.User, orders []models.Order) {
	cachedOrders := make([]*cache.Order, len(orders))

	for i, order := range orders {
		// Create a copy to avoid modifying the original
		orderCopy := order

		cachedOrders[i] = &cache.Order{
			UUID:      strconv.Itoa(order.ID),
			Value:     orderCopy,
			CreatedAt: order.CreatedAt,
			UpdatedAt: order.UpdatedAt,
		}
	}

	cachedProfile := &cache.Profile{
		UUID:   strconv.Itoa(user.ID),
		Name:   user.Email, // Use Name field to store email
		Orders: cachedOrders,
	}

	s.cache.Set(cachedProfile)
	s.logger.Debug("Profile saved to L1 cache", "userID", user.ID)
}

// saveToRedis saves user and orders to the L2 (Redis) cache
func (s *ProfileCacheService) saveToRedis(ctx context.Context, user *models.User, orders []models.Order) {
	cacheData := struct {
		User   *models.User
		Orders []models.Order
	}{
		User:   user,
		Orders: orders,
	}

	// Serialize the data
	data, err := json.Marshal(cacheData)
	if err != nil {
		s.logger.Error("Failed to serialize data for Redis", "error", err)
		return
	}

	// Save to Redis with a longer TTL (5 minutes)
	redisKey := fmt.Sprintf("user_profile:%d", user.ID)
	if err := s.redisClient.Set(ctx, redisKey, data, 5*time.Minute).Err(); err != nil {
		s.logger.Error("Failed to save to Redis", "error", err)
	}
}

// Shutdown stops the ProfileCacheService and its worker pool
func (s *ProfileCacheService) Shutdown() {
	// Stop the worker pool
	s.cacheWorker.Shutdown()

	// Stop the L1 cache cleanup
	s.cache.Shutdown()

	s.logger.Info("ProfileCacheService has been shut down")
}
