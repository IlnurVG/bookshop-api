package models

import "time"

// CartItem представляет элемент корзины
type CartItem struct {
	BookID    int       `json:"book_id" db:"book_id"`
	Book      *Book     `json:"book,omitempty" db:"-"`
	AddedAt   time.Time `json:"added_at" db:"added_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
}

// Cart представляет корзину пользователя
type Cart struct {
	UserID int        `json:"user_id" db:"user_id"`
	Items  []CartItem `json:"items" db:"-"`
}

// CartItemRequest представляет запрос на добавление товара в корзину
type CartItemRequest struct {
	BookID int `json:"book_id" validate:"required,gt=0"`
}

// CartResponse представляет ответ с корзиной
type CartResponse struct {
	Items     []CartItemResponse `json:"items"`
	TotalCost float64            `json:"total_cost"`
}

// CartItemResponse представляет элемент корзины в ответе API
type CartItemResponse struct {
	BookID    int       `json:"book_id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Price     float64   `json:"price"`
	AddedAt   time.Time `json:"added_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ToResponse преобразует корзину в ответ API
func (c *Cart) ToResponse() CartResponse {
	var response CartResponse
	var totalCost float64

	for _, item := range c.Items {
		if item.Book != nil {
			cartItem := CartItemResponse{
				BookID:    item.BookID,
				Title:     item.Book.Title,
				Author:    item.Book.Author,
				Price:     item.Book.Price,
				AddedAt:   item.AddedAt,
				ExpiresAt: item.ExpiresAt,
			}
			response.Items = append(response.Items, cartItem)
			totalCost += item.Book.Price
		}
	}

	response.TotalCost = totalCost
	return response
}
