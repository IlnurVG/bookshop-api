package checkout

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainerrors "github.com/bookshop/api/internal/domain/errors"
	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/bookshop/api/internal/service"
	"github.com/bookshop/api/pkg/logger"
)

const (
	// OrderStatusNew status for new orders
	OrderStatusNew = "new"
	// OrderStatusPaid status for paid orders
	OrderStatusPaid = "paid"
	// OrderStatusCanceled status for canceled orders
	OrderStatusCanceled = "canceled"
	// CartLockDuration duration of cart lock during checkout
	CartLockDuration = 5 * time.Minute
)

// Service implements services.CheckoutService interface
type Service struct {
	orderRepo           repositories.OrderRepository
	cartRepo            repositories.CartRepository
	bookRepo            repositories.BookRepository
	txManager           repositories.TransactionManager
	logger              logger.Logger
	profileCacheService *service.ProfileCacheService
}

// NewService creates a new instance of the checkout service
func NewService(
	orderRepo repositories.OrderRepository,
	cartRepo repositories.CartRepository,
	bookRepo repositories.BookRepository,
	txManager repositories.TransactionManager,
	logger logger.Logger,
	profileCacheService *service.ProfileCacheService,
) services.CheckoutService {
	return &Service{
		orderRepo:           orderRepo,
		cartRepo:            cartRepo,
		bookRepo:            bookRepo,
		txManager:           txManager,
		logger:              logger,
		profileCacheService: profileCacheService,
	}
}

// Checkout processes an order from the user's cart
func (s *Service) Checkout(ctx context.Context, userID int) (*models.Order, error) {
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
	if s.profileCacheService != nil {
		s.profileCacheService.InvalidateUserCache(context.Background(), userID)
	}

	return order, nil
}

// GetOrderByID returns an order by ID
func (s *Service) GetOrderByID(ctx context.Context, orderID int, userID int) (*models.Order, error) {
	// Get the order
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("error getting order: %w", err)
	}

	// Check if the order belongs to the user
	if order.UserID != userID {
		return nil, fmt.Errorf("order not found")
	}

	return order, nil
}

// GetOrdersByUserID returns a list of user's orders
func (s *Service) GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	// Get the list of orders
	orders, err := s.orderRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting order list: %w", err)
	}

	return orders, nil
}

// UpdateOrderStatus updates order status
func (s *Service) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
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
	if s.profileCacheService != nil && order != nil {
		// Update the specific order in cache
		s.profileCacheService.UpdateOrderInCache(context.Background(), order.UserID, order)
	}

	return nil
}
