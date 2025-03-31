package checkout

import (
	"time"

	"github.com/bookshop/api/internal/domain/models"
)

// CreateOrderRequest represents an order creation request
type CreateOrderRequest struct {
	DeliveryAddress string `json:"delivery_address" validate:"required,min=10,max=200"`
	PaymentMethod   string `json:"payment_method" validate:"required,oneof=card cash"`
}

// OrderItemResponse represents an order item in the API response
type OrderItemResponse struct {
	BookID int     `json:"book_id"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float64 `json:"price"`
}

// OrderResponse represents an order in the API response
type OrderResponse struct {
	ID         int                 `json:"id"`
	Status     string              `json:"status"`
	TotalPrice float64             `json:"total_price"`
	Items      []OrderItemResponse `json:"items"`
	CreatedAt  time.Time           `json:"created_at"`
}

// FromModel converts a model to an API response
func FromModel(order *models.Order) *OrderResponse {
	response := &OrderResponse{
		ID:         order.ID,
		Status:     order.Status,
		TotalPrice: order.TotalPrice,
		Items:      make([]OrderItemResponse, 0, len(order.Items)),
		CreatedAt:  order.CreatedAt,
	}

	for _, item := range order.Items {
		if item.Book != nil {
			response.Items = append(response.Items, OrderItemResponse{
				BookID: item.BookID,
				Title:  item.Book.Title,
				Author: item.Book.Author,
				Price:  item.Price,
			})
		}
	}

	return response
}

// FromModelList converts a list of models to a list of API responses
func FromModelList(orders []models.Order) []OrderResponse {
	result := make([]OrderResponse, len(orders))
	for i, order := range orders {
		result[i] = *FromModel(&order)
	}
	return result
}
