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

// CheckoutService реализует интерфейс services.CheckoutService
type CheckoutService struct {
	cartRepository  repositories.CartRepository
	orderRepository repositories.OrderRepository
	bookRepository  repositories.BookRepository
}

// NewCheckoutService создает новый экземпляр CheckoutService
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

// Checkout оформляет заказ из корзины пользователя
func (s *CheckoutService) Checkout(ctx context.Context, userID int) (*models.Order, error) {
	// Получаем корзину пользователя
	cart, err := s.cartRepository.GetCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения корзины: %w", err)
	}

	// Проверяем, что корзина не пуста
	if len(cart.Items) == 0 {
		return nil, errors.New("корзина пуста")
	}

	// Блокируем корзину на время оформления заказа
	if err := s.cartRepository.LockCart(ctx, userID, 5*time.Minute); err != nil {
		return nil, fmt.Errorf("ошибка блокировки корзины: %w", err)
	}
	defer s.cartRepository.UnlockCart(ctx, userID)

	// Создаем новый заказ
	order := &models.Order{
		UserID:     userID,
		Status:     "created",
		TotalPrice: 0,
		Items:      make([]models.OrderItem, 0, len(cart.Items)),
	}

	// Добавляем товары из корзины в заказ
	for _, cartItem := range cart.Items {
		// Получаем актуальную информацию о книге
		book, err := s.bookRepository.GetByID(ctx, cartItem.BookID)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения информации о книге: %w", err)
		}

		// Создаем элемент заказа
		orderItem := models.OrderItem{
			BookID: book.ID,
			Price:  book.Price,
		}

		order.Items = append(order.Items, orderItem)
		order.TotalPrice += book.Price
	}

	// Сохраняем заказ в базе данных
	if err := s.orderRepository.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("ошибка создания заказа: %w", err)
	}

	// Очищаем корзину пользователя
	if err := s.cartRepository.ClearCart(ctx, userID); err != nil {
		return nil, fmt.Errorf("ошибка очистки корзины: %w", err)
	}

	return order, nil
}

// GetOrdersByUserID возвращает список заказов пользователя
func (s *CheckoutService) GetOrdersByUserID(ctx context.Context, userID int) ([]models.Order, error) {
	orders, err := s.orderRepository.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения заказов пользователя: %w", err)
	}

	// Загружаем информацию о книгах для каждого заказа
	for i := range orders {
		for j := range orders[i].Items {
			book, err := s.bookRepository.GetByID(ctx, orders[i].Items[j].BookID)
			if err != nil {
				return nil, fmt.Errorf("ошибка получения информации о книге: %w", err)
			}
			orders[i].Items[j].Book = book
		}
	}

	return orders, nil
}

// GetOrderByID возвращает заказ по ID
func (s *CheckoutService) GetOrderByID(ctx context.Context, orderID int, userID int) (*models.Order, error) {
	order, err := s.orderRepository.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения заказа: %w", err)
	}

	// Проверяем, что заказ принадлежит пользователю
	if order.UserID != userID {
		return nil, errors.New("заказ не принадлежит пользователю")
	}

	// Загружаем информацию о книгах для заказа
	for i := range order.Items {
		book, err := s.bookRepository.GetByID(ctx, order.Items[i].BookID)
		if err != nil {
			return nil, fmt.Errorf("ошибка получения информации о книге: %w", err)
		}
		order.Items[i].Book = book
	}

	return order, nil
}
