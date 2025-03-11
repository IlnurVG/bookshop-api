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

// Определение ошибок
var (
	ErrEmptyCart  = errors.New("корзина пуста")
	ErrOutOfStock = errors.New("товар отсутствует на складе")
)

const (
	// OrderStatusNew статус нового заказа
	OrderStatusNew = "new"
	// OrderStatusPaid статус оплаченного заказа
	OrderStatusPaid = "paid"
	// OrderStatusCanceled статус отмененного заказа
	OrderStatusCanceled = "canceled"
	// CartLockDuration время блокировки корзины при оформлении заказа
	CartLockDuration = 5 * time.Minute
)

// Service реализует интерфейс services.CheckoutService
type Service struct {
	orderRepo repositories.OrderRepository
	cartRepo  repositories.CartRepository
	bookRepo  repositories.BookRepository
	logger    logger.Logger
}

// NewService создает новый экземпляр сервиса оформления заказа
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

// Checkout оформляет заказ из корзины пользователя
func (s *Service) Checkout(ctx context.Context, userID int) (*models.Order, error) {
	// Блокируем корзину на время оформления заказа
	if err := s.cartRepo.LockCart(ctx, userID, CartLockDuration); err != nil {
		return nil, fmt.Errorf("ошибка блокировки корзины: %w", err)
	}
	defer s.cartRepo.UnlockCart(ctx, userID)

	// Получаем корзину пользователя
	cart, err := s.cartRepo.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения корзины: %w", err)
	}

	// Проверяем, что корзина не пуста
	if len(cart.Items) == 0 {
		return nil, ErrEmptyCart
	}

	// Проверяем наличие товаров на складе и рассчитываем общую стоимость
	var totalPrice float64
	orderItems := make([]models.OrderItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		// Получаем книгу
		book, err := s.bookRepo.GetByID(ctx, item.BookID)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения книги: %w", err)
		}

		// Проверяем наличие на складе
		if book.Stock <= 0 {
			return nil, ErrOutOfStock
		}

		// Уменьшаем количество на складе
		book.Stock--
		if err := s.bookRepo.Update(ctx, book); err != nil {
			return nil, fmt.Errorf("ошибка обновления количества книг: %w", err)
		}

		// Добавляем товар в заказ
		orderItems = append(orderItems, models.OrderItem{
			BookID:    item.BookID,
			Book:      book,
			Price:     book.Price,
			CreatedAt: time.Now(),
		})

		totalPrice += book.Price
	}

	// Создаем заказ
	order := &models.Order{
		UserID:     userID,
		Status:     OrderStatusNew,
		TotalPrice: totalPrice,
		Items:      orderItems,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Сохраняем заказ
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("ошибка создания заказа: %w", err)
	}

	// Очищаем корзину
	if err := s.cartRepo.ClearCart(ctx, userID); err != nil {
		s.logger.Error("ошибка очистки корзины после создания заказа: %v", err)
	}

	return order, nil
}

// GetOrderByID возвращает заказ по ID
func (s *Service) GetOrderByID(ctx context.Context, orderID int, userID int) (*models.Order, error) {
	// Получаем заказ
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, fmt.Errorf("заказ не найден")
		}
		return nil, fmt.Errorf("ошибка получения заказа: %w", err)
	}

	// Проверяем, принадлежит ли заказ пользователю
	if order.UserID != userID {
		return nil, fmt.Errorf("заказ не найден")
	}

	return order, nil
}

// GetOrdersByUserID возвращает список заказов пользователя
func (s *Service) GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	// Получаем список заказов
	orders, err := s.orderRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка заказов: %w", err)
	}

	return orders, nil
}

// UpdateOrderStatus обновляет статус заказа
func (s *Service) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
	// Проверяем существование заказа
	if _, err := s.orderRepo.GetByID(ctx, orderID); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return fmt.Errorf("заказ не найден")
		}
		return fmt.Errorf("ошибка получения заказа: %w", err)
	}

	// Проверяем корректность статуса
	switch status {
	case OrderStatusPaid, OrderStatusCanceled:
		// Статус корректный
	default:
		return fmt.Errorf("некорректный статус заказа")
	}

	// Обновляем статус
	if err := s.orderRepo.UpdateStatus(ctx, orderID, status); err != nil {
		return fmt.Errorf("ошибка обновления статуса заказа: %w", err)
	}

	return nil
}
