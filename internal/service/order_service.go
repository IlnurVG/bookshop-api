package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	domainerrors "github.com/bookshop/api/internal/domain/errors"
	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/bookshop/api/internal/pkg/cache"
	"github.com/bookshop/api/pkg/logger"
)

const (
	// OrderStatusNew status for new orders
	OrderStatusNew = "new"
	// OrderStatusPaid status for paid orders
	OrderStatusPaid = "paid"
	// OrderStatusCanceled status for canceled orders
	OrderStatusCanceled = "canceled"
)

// OrderService handles business logic related to orders
type OrderService struct {
	orderRepo           repositories.OrderRepository
	userRepo            repositories.UserRepository
	bookRepo            repositories.BookRepository
	cartRepo            repositories.CartRepository // Used for cart management
	profileCache        *cache.ProfileCache         // L1 cache for user profiles
	profileCacheService *ProfileCacheService        // Service for profile caching operations
	logger              logger.Logger
	txManager           repositories.TransactionManager
	orderProcessor      *OrderProcessor // Asynchronous order processor
}

// NewOrderService creates a new service for working with orders
func NewOrderService(
	orderRepo repositories.OrderRepository,
	userRepo repositories.UserRepository,
	bookRepo repositories.BookRepository,
	cartRepo repositories.CartRepository,
	txManager repositories.TransactionManager,
	logger logger.Logger,
	profileCacheService *ProfileCacheService, // Optional, can be nil
) *OrderService {
	// Create in-memory cache with 2-second TTL and 1-second cleanup interval
	profileCache := cache.NewProfileCache(2*time.Second, 1*time.Second)

	// Create order processor with 5 workers
	orderProcessor := NewOrderProcessor(
		orderRepo,
		bookRepo,
		cartRepo,
		logger,
		5, // Number of workers
	)

	return &OrderService{
		orderRepo:           orderRepo,
		userRepo:            userRepo,
		bookRepo:            bookRepo,
		cartRepo:            cartRepo,
		profileCache:        profileCache,
		profileCacheService: profileCacheService,
		txManager:           txManager,
		logger:              logger,
		orderProcessor:      orderProcessor,
	}
}

// GetUserProfile returns a user profile with orders
// Uses multi-level caching strategy
func (s *OrderService) GetUserProfile(ctx context.Context, userID string) (*models.UserProfileResponse, error) {
	// If ProfileCacheService is available, use it
	if s.profileCacheService != nil {
		userIDInt, err := strconv.Atoi(userID)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID: %w", err)
		}

		// Get user and orders from cache service
		user, orders, err := s.profileCacheService.GetUserWithOrders(ctx, userIDInt)
		if err != nil {
			s.logger.Error("Failed to get user profile from cache service", "error", err, "userID", userID)
			// Fall through to original implementation
		} else {
			// Convert to response format
			return &models.UserProfileResponse{
				UUID:   userID,
				Name:   user.Email,
				Email:  user.Email,
				Orders: s.mapOrdersToProfileResponse(orders),
			}, nil
		}
	}

	// Original implementation as fallback
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
	var cart *models.Cart

	// Get cart and validate it within a transaction
	err = s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		var err error
		// Get user's cart from Redis
		cart, err = s.cartRepo.GetCart(txCtx, userIDInt)
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

		// Create new order object
		order = &models.Order{
			UserID:     userIDInt,
			Status:     "pending",
			TotalPrice: 0, // Will calculate from cart items
			CreatedAt:  time.Now(),
			Items:      []models.OrderItem{},
		}

		// Calculate total price and populate order items
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

		return nil
	})

	if err != nil {
		s.logger.Error("Failed to prepare order", "error", err, "userID", userID)
		return nil, err
	}

	// Send order for asynchronous processing
	s.logger.Info("Sending order for asynchronous processing", "userID", userID)

	// Create processing request
	processRequest := OrderProcessRequest{
		UserID:    userIDInt,
		CartItems: cart.Items,
		Order:     order,
	}

	// Start asynchronous processing
	resultCh := s.orderProcessor.ProcessOrder(ctx, processRequest)

	// Wait for processing result with timeout
	select {
	case err := <-resultCh:
		if err != nil {
			s.logger.Error("Error during asynchronous order processing", "error", err, "userID", userID)
			return nil, fmt.Errorf("order processing error: %w", err)
		}
	case <-time.After(500 * time.Millisecond):
		// If processing takes longer, return the order with "processing" status
		s.logger.Info("Order is being processed", "orderID", order.ID, "userID", userID)
		order.Status = "processing"
	}

	// Unlock the cart as we locked it within the transaction
	s.cartRepo.UnlockCart(ctx, userIDInt)

	// Invalidate caches
	// L1 cache - just delete it, it will be recreated on next request
	s.profileCache.Delete(userID)

	// L2 cache (Redis) - also delete
	s.cartRepo.GetRedisClient().Del(ctx, fmt.Sprintf("user_profile:%s", userID))

	s.logger.Info("Order creation process started", "orderID", order.ID, "userID", userID)

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
	// Use ProfileCacheService if available
	if s.profileCacheService != nil {
		userIDInt, err := strconv.Atoi(profile.UUID)
		if err != nil {
			s.logger.Error("Failed to convert user ID", "error", err, "userID", profile.UUID)
			return
		}

		// Convert ProfileOrderResponse to Order models
		orders := make([]models.Order, 0, len(profile.Orders))
		for _, order := range profile.Orders {
			orderIDInt, err := strconv.Atoi(order.UUID)
			if err != nil {
				continue
			}

			orders = append(orders, models.Order{
				ID:         orderIDInt,
				Status:     order.Status,
				TotalPrice: order.Total,
				CreatedAt:  order.CreatedAt,
			})
		}

		// Create user model
		user := &models.User{
			ID:    userIDInt,
			Email: profile.Email,
		}

		// Use worker pool to update cache
		s.profileCacheService.UpdateUserCache(context.Background(), user, orders)
		return
	}

	// Original implementation as fallback
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

	// Stop the order processor
	s.orderProcessor.Shutdown()

	s.logger.Info("Order service shutdown completed")
}

// Checkout processes an order from the user's cart
func (s *OrderService) Checkout(ctx context.Context, userID int) (*models.Order, error) {
	var order *models.Order

	// Execute the checkout in a transaction
	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get user's cart
		cart, err := s.cartRepo.GetCart(txCtx, userID)
		if err != nil {
			return fmt.Errorf("error getting cart: %w", err)
		}

		// Check if cart is empty
		if len(cart.Items) == 0 {
			return domainerrors.ErrEmptyCart
		}

		// Get books from cart
		bookIDs := make([]int, len(cart.Items))
		for i, item := range cart.Items {
			bookIDs[i] = item.BookID
		}

		books, err := s.bookRepo.GetBooksByIDs(txCtx, bookIDs)
		if err != nil {
			return fmt.Errorf("error getting books: %w", err)
		}

		// Check if all books are in stock
		for _, book := range books {
			if book.Stock <= 0 {
				return domainerrors.ErrOutOfStock
			}
		}

		// Create order
		order = &models.Order{
			UserID:     userID,
			Status:     OrderStatusNew,
			TotalPrice: 0,
			Items:      make([]models.OrderItem, len(cart.Items)),
		}

		// Calculate total price and create order items
		for i, item := range cart.Items {
			var book *models.Book
			for _, b := range books {
				if b.ID == item.BookID {
					book = &b
					break
				}
			}

			if book == nil {
				return domainerrors.ErrBookNotFound
			}

			order.Items[i] = models.OrderItem{
				BookID: book.ID,
				Price:  book.Price,
			}
			order.TotalPrice += book.Price
		}

		// Save order
		if err := s.orderRepo.Create(txCtx, order); err != nil {
			return fmt.Errorf("error creating order: %w", err)
		}

		// Update book stock
		for _, item := range cart.Items {
			if err := s.bookRepo.DecrementStock(txCtx, item.BookID, 1); err != nil {
				return fmt.Errorf("error updating book stock: %w", err)
			}
		}

		// Clear cart
		if err := s.cartRepo.ClearCart(txCtx, userID); err != nil {
			return fmt.Errorf("error clearing cart: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Invalidate user profile cache after successful order creation
	// Use non-blocking async version to avoid blocking the request
	if s.profileCacheService != nil {
		s.profileCacheService.InvalidateUserCacheAsync(userID)
	}

	return order, nil
}

// UpdateOrderStatus updates order status
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
	var order *models.Order

	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Check if the order exists
		var err error
		order, err = s.orderRepo.GetByID(txCtx, orderID)
		if err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return fmt.Errorf("order not found")
			}
			return fmt.Errorf("error getting order: %w", err)
		}

		// Check status validity
		switch status {
		case OrderStatusPaid, OrderStatusCanceled:
			// Valid status
		default:
			return fmt.Errorf("invalid order status")
		}

		// Update status
		if err := s.orderRepo.UpdateStatus(txCtx, orderID, status); err != nil {
			return fmt.Errorf("error updating order status: %w", err)
		}

		// Update order status in our local variable
		order.Status = status

		return nil
	})

	if err != nil {
		return err
	}

	// Update the cache after successful status update
	// Use non-blocking async version to avoid blocking the request
	if s.profileCacheService != nil && order != nil {
		// Update the specific order in cache asynchronously
		s.profileCacheService.UpdateOrderInCacheAsync(order.UserID, order)
	}

	return nil
}
