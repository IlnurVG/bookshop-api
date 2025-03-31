package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/bookshop/api/internal/domain/services"
)

// CheckoutService implements services.CheckoutService interface
type CheckoutService struct {
	cartRepository  repositories.CartRepository
	orderRepository repositories.OrderRepository
	bookRepository  repositories.BookRepository
}

// NewCheckoutService creates a new CheckoutService instance
func NewCheckoutService(
	cartRepository repositories.CartRepository,
	orderRepository repositories.OrderRepository,
	bookRepository repositories.BookRepository,
) services.CheckoutService {
	return &CheckoutService{
		cartRepository:  cartRepository,
		orderRepository: orderRepository,
		bookRepository:  bookRepository,
	}
}

// Checkout processes an order from the user's cart
func (s *CheckoutService) Checkout(ctx context.Context, userID int) (*models.Order, error) {
	// Get user's cart
	cart, err := s.cartRepository.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting cart: %w", err)
	}

	// Check if cart is empty
	if len(cart.Items) == 0 {
		return nil, errors.New("cart is empty")
	}

	// Lock cart
	if err := s.cartRepository.LockCart(ctx, userID, 5*time.Minute); err != nil {
		return nil, fmt.Errorf("error locking cart: %w", err)
	}
	defer s.cartRepository.UnlockCart(ctx, userID)

	// Create order
	order := &models.Order{
		UserID:     userID,
		Status:     "created",
		TotalPrice: 0,
		Items:      make([]models.OrderItem, 0, len(cart.Items)),
	}

	// Add items to order
	for _, cartItem := range cart.Items {
		// Get book to get its price
		book, err := s.bookRepository.GetByID(ctx, cartItem.BookID)
		if err != nil {
			return nil, fmt.Errorf("error getting book info: %w", err)
		}

		// Create order item
		orderItem := models.OrderItem{
			BookID: book.ID,
			Price:  book.Price,
		}

		order.Items = append(order.Items, orderItem)
		order.TotalPrice += book.Price
	}

	// Save order
	if err := s.orderRepository.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("error creating order: %w", err)
	}

	// Clear cart
	if err := s.cartRepository.ClearCart(ctx, userID); err != nil {
		return nil, fmt.Errorf("error clearing cart: %w", err)
	}

	return order, nil
}

// GetOrdersByUserID returns a list of user's orders
func (s *CheckoutService) GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	orders, err := s.orderRepository.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user orders: %w", err)
	}

	// Load book information for each order
	for i := range orders {
		for j := range orders[i].Items {
			book, err := s.bookRepository.GetByID(ctx, orders[i].Items[j].BookID)
			if err != nil {
				continue
			}
			orders[i].Items[j].Book = book
		}
	}

	return orders, nil
}

// GetOrderByID returns an order by ID
func (s *CheckoutService) GetOrderByID(ctx context.Context, orderID int, userID int) (*models.Order, error) {
	order, err := s.orderRepository.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("error getting order: %w", err)
	}

	// Check if order belongs to user
	if order.UserID != userID {
		return nil, errors.New("order does not belong to user")
	}

	// Load book information for order
	for i := range order.Items {
		book, err := s.bookRepository.GetByID(ctx, order.Items[i].BookID)
		if err != nil {
			continue
		}
		order.Items[i].Book = book
	}

	return order, nil
}

// UpdateOrderStatus updates the status of an order
func (s *CheckoutService) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
	return s.orderRepository.UpdateStatus(ctx, orderID, status)
}
