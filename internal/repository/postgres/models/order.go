package models

import (
	"time"

	domainmodels "github.com/bookshop/api/internal/domain/models"
)

// OrderItem represents an order item for repository operations
type OrderItem struct {
	ID        int       `db:"id"`
	OrderID   int       `db:"order_id"`
	BookID    int       `db:"book_id"`
	Price     float64   `db:"price"`
	CreatedAt time.Time `db:"created_at"`
}

// Order represents an order model for repository operations
type Order struct {
	ID         int     `db:"id"`
	UserID     int     `db:"user_id"`
	Status     string  `db:"status"`
	TotalPrice float64 `db:"total_price"`
	Items      []OrderItem
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// OrderItemToDomain converts repository order item to domain model
func (oi *OrderItem) ToDomain() domainmodels.OrderItem {
	return domainmodels.OrderItem{
		ID:        oi.ID,
		OrderID:   oi.OrderID,
		BookID:    oi.BookID,
		Price:     oi.Price,
		CreatedAt: oi.CreatedAt,
	}
}

// OrderItemFromDomain converts domain order item to repository model
func OrderItemFromDomain(item domainmodels.OrderItem) OrderItem {
	return OrderItem{
		ID:        item.ID,
		OrderID:   item.OrderID,
		BookID:    item.BookID,
		Price:     item.Price,
		CreatedAt: item.CreatedAt,
	}
}

// OrderToDomain converts repository order to domain model
func (o *Order) ToDomain() *domainmodels.Order {
	domainOrder := &domainmodels.Order{
		ID:         o.ID,
		UserID:     o.UserID,
		Status:     o.Status,
		TotalPrice: o.TotalPrice,
		Items:      make([]domainmodels.OrderItem, len(o.Items)),
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  o.UpdatedAt,
	}

	for i, item := range o.Items {
		domainOrder.Items[i] = item.ToDomain()
	}

	return domainOrder
}

// OrderFromDomain converts domain order to repository model
func OrderFromDomain(order *domainmodels.Order) *Order {
	repoOrder := &Order{
		ID:         order.ID,
		UserID:     order.UserID,
		Status:     order.Status,
		TotalPrice: order.TotalPrice,
		Items:      make([]OrderItem, len(order.Items)),
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
	}

	for i, item := range order.Items {
		repoOrder.Items[i] = OrderItemFromDomain(item)
	}

	return repoOrder
}

// OrderSliceToDomain converts a slice of repository orders to domain models
func OrderSliceToDomain(orders []Order) []domainmodels.Order {
	result := make([]domainmodels.Order, len(orders))
	for i, order := range orders {
		domainOrder := order.ToDomain()
		result[i] = *domainOrder
	}
	return result
}
