package models

import (
	"time"

	bookmodels "github.com/bookshop/api/internal/app/book/models"
	domainmodels "github.com/bookshop/api/internal/domain/models"
)

// OrderItem represents an order item for service operations
type OrderItem struct {
	ID        int
	OrderID   int
	BookID    int
	Book      *bookmodels.Book
	Price     float64
	CreatedAt time.Time
}

// Order represents an order model for service operations
type Order struct {
	ID         int
	UserID     int
	Status     string
	TotalPrice float64
	Items      []OrderItem
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// OrderResponse represents an order response
type OrderResponse struct {
	ID         int
	Status     string
	TotalPrice float64
	Items      []OrderItemResponse
	CreatedAt  time.Time
}

// OrderItemResponse represents an order item in API response
type OrderItemResponse struct {
	BookID int
	Title  string
	Author string
	Price  float64
}

// OrderItemToDomain converts service order item to domain model
func (oi *OrderItem) ToDomain() domainmodels.OrderItem {
	domainItem := domainmodels.OrderItem{
		ID:        oi.ID,
		OrderID:   oi.OrderID,
		BookID:    oi.BookID,
		Price:     oi.Price,
		CreatedAt: oi.CreatedAt,
	}

	if oi.Book != nil {
		domainBook := oi.Book.ToDomain()
		domainItem.Book = domainBook
	}

	return domainItem
}

// OrderItemFromDomain converts domain order item to service model
func OrderItemFromDomain(item domainmodels.OrderItem) OrderItem {
	serviceItem := OrderItem{
		ID:        item.ID,
		OrderID:   item.OrderID,
		BookID:    item.BookID,
		Price:     item.Price,
		CreatedAt: item.CreatedAt,
	}

	if item.Book != nil {
		serviceItem.Book = bookmodels.FromDomain(item.Book)
	}

	return serviceItem
}

// OrderToDomain converts service order to domain model
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

// OrderFromDomain converts domain order to service model
func OrderFromDomain(order *domainmodels.Order) *Order {
	serviceOrder := &Order{
		ID:         order.ID,
		UserID:     order.UserID,
		Status:     order.Status,
		TotalPrice: order.TotalPrice,
		Items:      make([]OrderItem, len(order.Items)),
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
	}

	for i, item := range order.Items {
		serviceOrder.Items[i] = OrderItemFromDomain(item)
	}

	return serviceOrder
}

// OrderSliceFromDomain converts a slice of domain orders to service models
func OrderSliceFromDomain(orders []domainmodels.Order) []Order {
	result := make([]Order, len(orders))
	for i, order := range orders {
		orderCopy := order // Create a copy to avoid issues with loop variable references
		serviceOrder := OrderFromDomain(&orderCopy)
		result[i] = *serviceOrder
	}
	return result
}
