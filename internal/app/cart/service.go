package cart

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

// Определение ошибок
var (
	ErrBookNotFound = errors.New("книга не найдена")
	ErrCartEmpty    = errors.New("корзина пуста")
)

const (
	// ItemExpirationTime время жизни товара в корзине
	ItemExpirationTime = 24 * time.Hour
)

// Service реализует интерфейс services.CartService
type Service struct {
	cartRepo repositories.CartRepository
	bookRepo repositories.BookRepository
	logger   logger.Logger
}

// NewService создает новый экземпляр сервиса корзины
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

// AddItem добавляет товар в корзину пользователя
func (s *Service) AddItem(ctx context.Context, userID int, input models.CartItemRequest) error {
	// Проверяем существование книги
	book, err := s.bookRepo.GetByID(ctx, input.BookID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return ErrBookNotFound
		}
		return fmt.Errorf("ошибка получения книги: %w", err)
	}

	// Проверяем наличие книги на складе
	if book.Stock <= 0 {
		return fmt.Errorf("книга отсутствует на складе")
	}

	// Добавляем товар в корзину
	expiresAt := time.Now().Add(ItemExpirationTime)
	if err := s.cartRepo.AddItem(ctx, userID, input.BookID, expiresAt); err != nil {
		return fmt.Errorf("ошибка добавления товара в корзину: %w", err)
	}

	return nil
}

// GetCart возвращает корзину пользователя
func (s *Service) GetCart(ctx context.Context, userID int) (*models.CartResponse, error) {
	// Получаем корзину пользователя
	cart, err := s.cartRepo.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения корзины: %w", err)
	}

	// Преобразуем модель в ответ
	response := cart.ToResponse()
	return &response, nil
}

// RemoveItem удаляет товар из корзины пользователя
func (s *Service) RemoveItem(ctx context.Context, userID int, bookID int) error {
	// Удаляем товар из корзины
	if err := s.cartRepo.RemoveItem(ctx, userID, bookID); err != nil {
		return fmt.Errorf("ошибка удаления товара из корзины: %w", err)
	}

	return nil
}

// ClearCart очищает корзину пользователя
func (s *Service) ClearCart(ctx context.Context, userID int) error {
	// Очищаем корзину
	if err := s.cartRepo.ClearCart(ctx, userID); err != nil {
		return fmt.Errorf("ошибка очистки корзины: %w", err)
	}

	return nil
}

// CleanupExpiredItems удаляет просроченные товары из всех корзин
func (s *Service) CleanupExpiredItems(ctx context.Context) error {
	// Удаляем просроченные товары
	if err := s.cartRepo.RemoveExpiredItems(ctx); err != nil {
		return fmt.Errorf("ошибка удаления просроченных товаров: %w", err)
	}

	return nil
}
