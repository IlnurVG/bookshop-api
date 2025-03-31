package checkout

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/bookshop/api/pkg/logger"
)

// Error definitions
var (
	ErrEmptyCart  = errors.New("cart is empty")
	ErrOutOfStock = errors.New("item is out of stock")
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
	orderRepo repositories.OrderRepository
	cartRepo  repositories.CartRepository
	bookRepo  repositories.BookRepository
	logger    logger.Logger
}

// NewService creates a new instance of the checkout service
func NewService(
	orderRepo repositories.OrderRepository,
	cartRepo repositories.CartRepository,
	bookRepo repositories.BookRepository,
	logger logger.Logger,
) services.CheckoutService {
	return &Service{
		orderRepo: orderRepo,
		cartRepo:  cartRepo,
		bookRepo:  bookRepo,
		logger:    logger,
	}
}

// Checkout processes an order from the user's cart
func (s *Service) Checkout(ctx context.Context, userID int) (*models.Order, error) {
	// Lock the cart during checkout
	if err := s.cartRepo.LockCart(ctx, userID, CartLockDuration); err != nil {
		return nil, fmt.Errorf("error locking cart: %w", err)
	}
	defer s.cartRepo.UnlockCart(ctx, userID)

	// Get the user's cart
	cart, err := s.cartRepo.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting cart: %w", err)
	}

	// Check that the cart is not empty
	if len(cart.Items) == 0 {
		return nil, ErrEmptyCart
	}

	// Check stock availability and calculate total price
	var totalPrice float64
	orderItems := make([]models.OrderItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		// Get the book
		book, err := s.bookRepo.GetByID(ctx, item.BookID)
		if err != nil {
			return nil, fmt.Errorf("error getting book: %w", err)
		}

		// Check stock availability
		if book.Stock <= 0 {
			return nil, ErrOutOfStock
		}

		// Reduce stock quantity
		book.Stock--
		if err := s.bookRepo.Update(ctx, book); err != nil {
			return nil, fmt.Errorf("error updating book quantity: %w", err)
		}

		// Add item to order
		orderItems = append(orderItems, models.OrderItem{
			BookID:    item.BookID,
			Book:      book,
			Price:     book.Price,
			CreatedAt: time.Now(),
		})

		totalPrice += book.Price
	}

	// Create order
	order := &models.Order{
		UserID:     userID,
		Status:     OrderStatusNew,
		TotalPrice: totalPrice,
		Items:      orderItems,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Save the order
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("error creating order: %w", err)
	}

	// Clear the cart
	if err := s.cartRepo.ClearCart(ctx, userID); err != nil {
		s.logger.Error("error clearing cart after order creation: %v", err)
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
	// Check if the order exists
	if _, err := s.orderRepo.GetByID(ctx, orderID); err != nil {
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
	if err := s.orderRepo.UpdateStatus(ctx, orderID, status); err != nil {
		return fmt.Errorf("error updating order status: %w", err)
	}

	return nil
}
