package models

import "time"

// Order представляет модель заказа
type Order struct {
	ID         int         `json:"id" db:"id"`
	UserID     int         `json:"user_id" db:"user_id"`
	Status     string      `json:"status" db:"status"`
	TotalPrice float64     `json:"total_price" db:"total_price"`
	Items      []OrderItem `json:"items,omitempty" db:"-"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at" db:"updated_at"`
}

// OrderItem представляет элемент заказа
type OrderItem struct {
	ID        int       `json:"id" db:"id"`
	OrderID   int       `json:"order_id" db:"order_id"`
	BookID    int       `json:"book_id" db:"book_id"`
	Book      *Book     `json:"book,omitempty" db:"-"`
	Price     float64   `json:"price" db:"price"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// OrderResponse представляет ответ с заказом
type OrderResponse struct {
	ID         int                 `json:"id"`
	Status     string              `json:"status"`
	TotalPrice float64             `json:"total_price"`
	Items      []OrderItemResponse `json:"items"`
	CreatedAt  time.Time           `json:"created_at"`
}

// OrderItemResponse представляет элемент заказа в ответе API
type OrderItemResponse struct {
	BookID int     `json:"book_id"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float64 `json:"price"`
}

// ToResponse преобразует заказ в ответ API
func (o *Order) ToResponse() OrderResponse {
	var response OrderResponse
	response.ID = o.ID
	response.Status = o.Status
	response.TotalPrice = o.TotalPrice
	response.CreatedAt = o.CreatedAt

	for _, item := range o.Items {
		if item.Book != nil {
			orderItem := OrderItemResponse{
				BookID: item.BookID,
				Title:  item.Book.Title,
				Author: item.Book.Author,
				Price:  item.Price,
			}
			response.Items = append(response.Items, orderItem)
		}
	}

	return response
}
