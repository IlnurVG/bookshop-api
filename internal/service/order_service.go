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
)

// OrderService handles business logic related to orders
type OrderService struct {
	orderRepo    repositories.OrderRepository
	userRepo     repositories.UserRepository
	bookRepo     repositories.BookRepository
	cartRepo     repositories.CartRepository // Used for cart management
	profileCache *cache.ProfileCache         // L1 cache for user profiles
	logger       logger.Logger
	txManager    repositories.TransactionManager
}

// NewOrderService creates a new service for working with orders
func NewOrderService(
	orderRepo repositories.OrderRepository,
	userRepo repositories.UserRepository,
	bookRepo repositories.BookRepository,
	cartRepo repositories.CartRepository,
	txManager repositories.TransactionManager,
	logger logger.Logger,
) *OrderService {
	// Create in-memory cache with 2-second TTL and 1-second cleanup interval
	profileCache := cache.NewProfileCache(2*time.Second, 1*time.Second)

	return &OrderService{
		orderRepo:    orderRepo,
		userRepo:     userRepo,
		bookRepo:     bookRepo,
		cartRepo:     cartRepo,
		profileCache: profileCache,
		txManager:    txManager,
		logger:       logger,
	}
}

// GetUserProfile returns a user profile with orders
// Uses multi-level caching strategy
func (s *OrderService) GetUserProfile(ctx context.Context, userID string) (*models.UserProfileResponse, error) {
	// First try to get profile from L1 cache (in-memory)
	cachedProfile := s.profileCache.Get(userID)

	if cachedProfile != nil {
		s.logger.Debug("Profile retrieved from L1 cache", "userID", userID)
		// Convert cached profile to response model
		return s.mapCacheProfileToResponse(cachedProfile), nil
	}

	// If not in L1 cache, try to get it from Redis (L2 cache)
	// Here we use the existing Redis client to get data
	redisKey := fmt.Sprintf("user_profile:%s", userID)
	userData, err := s.cartRepo.GetRedisClient().Get(ctx, redisKey).Bytes()

	if err == nil && len(userData) > 0 {
		s.logger.Debug("Profile retrieved from Redis cache", "userID", userID)
		// Data found in Redis
		var userProfile models.UserProfileResponse
		if err := json.Unmarshal(userData, &userProfile); err == nil {
			// Save to L1 cache for faster future requests
			s.saveProfileToCache(&userProfile)
			return &userProfile, nil
		}
	}

	s.logger.Debug("Profile not found in cache, fetching from database", "userID", userID)

	// If data is not in any cache, get it from the database
	// Convert string ID to int for repository calls
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Use a transaction for consistent reads
	var response *models.UserProfileResponse

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get user info
		user, err := s.userRepo.GetByID(txCtx, userIDInt)
		if err != nil {
			return fmt.Errorf("error getting user: %w", err)
		}

		// Get user orders
		orders, err := s.orderRepo.GetByUserID(txCtx, userIDInt)
		if err != nil {
			return fmt.Errorf("error getting orders: %w", err)
		}

		// Create response
		response = &models.UserProfileResponse{
			UUID:   userID,
			Name:   user.Email,
			Email:  user.Email,
			Orders: s.mapOrdersToProfileResponse(orders),
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to get user profile", "error", err, "userID", userID)
		return nil, err
	}

	// Save in caches for future requests
	// First in Redis (L2 cache)
	profileJSON, _ := json.Marshal(response)
	s.cartRepo.GetRedisClient().Set(ctx, redisKey, profileJSON, 5*time.Minute)

	// Then in L1 cache (in-memory)
	s.saveProfileToCache(response)

	return response, nil
}

// CreateOrder creates a new order and updates caches
func (s *OrderService) CreateOrder(ctx context.Context, userID string, input models.CreateOrderRequest) (*models.Order, error) {
	// Convert string ID to int for repository calls
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var order *models.Order

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get user's cart from Redis
		cart, err := s.cartRepo.GetCart(txCtx, userIDInt)
		if err != nil {
			return fmt.Errorf("error getting cart: %w", err)
		}

		if len(cart.Items) == 0 {
			return fmt.Errorf("cart is empty")
		}

		// Lock cart during checkout
		if err := s.cartRepo.LockCart(txCtx, userIDInt, 5*time.Minute); err != nil {
			return fmt.Errorf("error locking cart: %w", err)
		}
		defer s.cartRepo.UnlockCart(txCtx, userIDInt)

		// Create new order object
		order = &models.Order{
			UserID:     userIDInt,
			Status:     "pending",
			TotalPrice: 0, // Will calculate from cart items
			CreatedAt:  time.Now(),
			Items:      []models.OrderItem{},
		}

		// Add items from cart
		for _, item := range cart.Items {
			// Get book to get price
			book, err := s.bookRepo.GetByID(txCtx, item.BookID)
			if err != nil {
				return fmt.Errorf("error getting book: %w", err)
			}

			// Add item to order
			orderItem := models.OrderItem{
				BookID: item.BookID,
				Price:  book.Price,
				Book:   book,
			}
			order.Items = append(order.Items, orderItem)
			order.TotalPrice += book.Price
		}

		// Create order in database
		if err := s.orderRepo.Create(txCtx, order); err != nil {
			return fmt.Errorf("error creating order: %w", err)
		}

		// Clear cart after successful order creation
		if err := s.cartRepo.ClearCart(txCtx, userIDInt); err != nil {
			return fmt.Errorf("error clearing cart: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to create order", "error", err, "userID", userID)
		return nil, err
	}

	// Invalidate caches
	// L1 cache - just delete it, it will be recreated on next request
	s.profileCache.Delete(userID)

	// L2 cache (Redis) - also delete
	s.cartRepo.GetRedisClient().Del(ctx, fmt.Sprintf("user_profile:%s", userID))

	s.logger.Info("Order created successfully", "orderID", order.ID, "userID", userID)

	return order, nil
}

// UpdateOrder updates an order and caches
func (s *OrderService) UpdateOrder(ctx context.Context, orderID string, userID string, input models.UpdateOrderRequest) (*models.Order, error) {
	// Convert string ID to int for repository calls
	orderIDInt, err := strconv.Atoi(orderID)
	if err != nil {
		return nil, fmt.Errorf("invalid order ID: %w", err)
	}

	var order *models.Order

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		var err error
		// First get the order
		order, err = s.orderRepo.GetByID(txCtx, orderIDInt)
		if err != nil {
			return fmt.Errorf("error getting order: %w", err)
		}

		// Update order status if provided
		if input.Status != "" {
			if err := s.orderRepo.UpdateStatus(txCtx, orderIDInt, input.Status); err != nil {
				return fmt.Errorf("error updating order status: %w", err)
			}
			order.Status = input.Status
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to update order", "error", err, "orderID", orderID, "userID", userID)
		return nil, err
	}

	// Update caches
	// Check if profile is in L1 cache
	cachedProfile := s.profileCache.Get(userID)
	if cachedProfile != nil {
		// Find and update order in cache
		cachedOrder := &cache.Order{
			UUID:      orderID,
			Value:     order,
			UpdatedAt: time.Now(),
		}
		s.profileCache.UpdateOrder(userID, orderID, cachedOrder)
		s.logger.Debug("Updated order in L1 cache", "orderID", orderID, "userID", userID)
	} else {
		// If not in L1 cache, just invalidate L2 cache
		s.cartRepo.GetRedisClient().Del(ctx, fmt.Sprintf("user_profile:%s", userID))
		s.logger.Debug("Invalidated L2 cache for user", "userID", userID)
	}

	s.logger.Info("Order updated successfully", "orderID", orderID, "userID", userID)

	return order, nil
}

// DeleteOrder deletes an order and updates caches
func (s *OrderService) DeleteOrder(ctx context.Context, orderID string, userID string) error {
	// Convert string ID to int for repository calls
	orderIDInt, err := strconv.Atoi(orderID)
	if err != nil {
		return fmt.Errorf("invalid order ID: %w", err)
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Check if order belongs to the user
		order, err := s.orderRepo.GetByID(txCtx, orderIDInt)
		if err != nil {
			return fmt.Errorf("error getting order: %w", err)
		}

		if order.UserID != userIDInt {
			return fmt.Errorf("order does not belong to user")
		}

		// Delete order
		if err := s.orderRepo.Delete(txCtx, orderIDInt); err != nil {
			return fmt.Errorf("error deleting order: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to delete order", "error", err, "orderID", orderID, "userID", userID)
		return err
	}

	// Update caches
	// Check if profile is in L1 cache
	cachedProfile := s.profileCache.Get(userID)
	if cachedProfile != nil {
		// Remove order from cache
		s.profileCache.RemoveOrder(userID, orderID)
		s.logger.Debug("Removed order from L1 cache", "orderID", orderID, "userID", userID)
	}

	// Invalidate L2 cache
	s.cartRepo.GetRedisClient().Del(ctx, fmt.Sprintf("user_profile:%s", userID))
	s.logger.Debug("Invalidated L2 cache for user", "userID", userID)

	s.logger.Info("Order deleted successfully", "orderID", orderID, "userID", userID)

	return nil
}

// mapOrdersToProfileResponse converts domain order models to ProfileOrderResponse format
func (s *OrderService) mapOrdersToProfileResponse(orders []models.Order) []models.ProfileOrderResponse {
	response := make([]models.ProfileOrderResponse, len(orders))

	for i, order := range orders {
		response[i] = models.ProfileOrderResponse{
			UUID:      strconv.Itoa(order.ID),
			Status:    order.Status,
			Total:     order.TotalPrice,
			CreatedAt: order.CreatedAt,
		}
	}

	return response
}

// mapCacheProfileToResponse converts a cached profile to API response format
func (s *OrderService) mapCacheProfileToResponse(profile *cache.Profile) *models.UserProfileResponse {
	orders := make([]models.ProfileOrderResponse, 0, len(profile.Orders))

	for _, cachedOrder := range profile.Orders {
		// Check if Value can be converted to order
		if order, ok := cachedOrder.Value.(models.Order); ok {
			orders = append(orders, models.ProfileOrderResponse{
				UUID:      strconv.Itoa(order.ID),
				Status:    order.Status,
				Total:     order.TotalPrice,
				CreatedAt: order.CreatedAt,
			})
		} else if orderPtr, ok := cachedOrder.Value.(*models.Order); ok && orderPtr != nil {
			orders = append(orders, models.ProfileOrderResponse{
				UUID:      strconv.Itoa(orderPtr.ID),
				Status:    orderPtr.Status,
				Total:     orderPtr.TotalPrice,
				CreatedAt: orderPtr.CreatedAt,
			})
		}
	}

	return &models.UserProfileResponse{
		UUID:   profile.UUID,
		Name:   profile.Name,
		Orders: orders,
	}
}

// saveProfileToCache saves a user profile to L1 cache
func (s *OrderService) saveProfileToCache(profile *models.UserProfileResponse) {
	// Convert response model to cache format
	cachedOrders := make([]*cache.Order, len(profile.Orders))

	for i, order := range profile.Orders {
		cachedOrders[i] = &cache.Order{
			UUID:      order.UUID,
			Value:     order,
			CreatedAt: order.CreatedAt,
			UpdatedAt: time.Now(),
		}
	}

	cachedProfile := &cache.Profile{
		UUID:   profile.UUID,
		Name:   profile.Name,
		Orders: cachedOrders,
	}

	// Save to L1 cache
	s.profileCache.Set(cachedProfile)
	s.logger.Debug("Profile saved to L1 cache", "userID", profile.UUID)
}

// Shutdown properly stops the service
func (s *OrderService) Shutdown() {
	// Stop cache cleanup goroutine
	s.profileCache.Stop()
	s.logger.Info("Order service shutdown completed")
}
