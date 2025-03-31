package cart

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainerrors "github.com/bookshop/api/internal/domain/errors"
	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/bookshop/api/pkg/logger"
)

const (
	// ItemExpirationTime is the lifetime of an item in the cart
	ItemExpirationTime = 24 * time.Hour
)

// Service implements services.CartService interface
type Service struct {
	cartRepo repositories.CartRepository
	bookRepo repositories.BookRepository
	logger   logger.Logger
}

// NewService creates a new instance of the cart service
func NewService(
	cartRepo repositories.CartRepository,
	bookRepo repositories.BookRepository,
	logger logger.Logger,
) services.CartService {
	return &Service{
		cartRepo: cartRepo,
		bookRepo: bookRepo,
		logger:   logger,
	}
}

// AddItem adds an item to the user's cart
func (s *Service) AddItem(ctx context.Context, userID int, input models.CartItemRequest) error {
	// Check if the book exists
	book, err := s.bookRepo.GetByID(ctx, input.BookID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return domainerrors.ErrBookNotFound
		}
		return fmt.Errorf("error getting book: %w", err)
	}

	// Check if the book is in stock
	if book.Stock <= 0 {
		return domainerrors.ErrOutOfStock
	}

	// Add item to cart
	expiresAt := time.Now().Add(ItemExpirationTime)
	if err := s.cartRepo.AddItem(ctx, userID, input.BookID, expiresAt); err != nil {
		return fmt.Errorf("error adding item to cart: %w", err)
	}

	return nil
}

// GetCart returns the user's cart
func (s *Service) GetCart(ctx context.Context, userID int) (*models.CartResponse, error) {
	// Get user's cart
	cart, err := s.cartRepo.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting cart: %w", err)
	}

	// Convert model to response
	response := cart.ToResponse()
	return &response, nil
}

// RemoveItem removes an item from the user's cart
func (s *Service) RemoveItem(ctx context.Context, userID int, bookID int) error {
	// Remove item from cart
	if err := s.cartRepo.RemoveItem(ctx, userID, bookID); err != nil {
		return fmt.Errorf("error removing item from cart: %w", err)
	}

	return nil
}

// ClearCart clears the user's cart
func (s *Service) ClearCart(ctx context.Context, userID int) error {
	// Clear the cart
	if err := s.cartRepo.ClearCart(ctx, userID); err != nil {
		return fmt.Errorf("error clearing cart: %w", err)
	}

	return nil
}

// CleanupExpiredItems removes expired items from all carts
func (s *Service) CleanupExpiredItems(ctx context.Context) error {
	// Remove expired items
	if err := s.cartRepo.RemoveExpiredItems(ctx); err != nil {
		return fmt.Errorf("error removing expired items: %w", err)
	}

	return nil
}
