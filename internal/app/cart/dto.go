package cart

import "time"

// CartItemRequest представляет запрос на добавление товара в корзину
type CartItemRequest struct {
	BookID int `json:"book_id" binding:"required"`
	Count  int `json:"count" binding:"required,min=1"`
}

// CartItemResponse представляет элемент корзины в ответе API
type CartItemResponse struct {
	ID        int       `json:"id"`
	BookID    int       `json:"book_id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Price     float64   `json:"price"`
	Count     int       `json:"count"`
	Total     float64   `json:"total"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CartResponse представляет корзину в ответе API
type CartResponse struct {
	Items      []CartItemResponse `json:"items"`
	TotalItems int                `json:"total_items"`
	TotalPrice float64            `json:"total_price"`
}

// UpdateCartItemRequest представляет запрос на обновление количества товара в корзине
type UpdateCartItemRequest struct {
	Count int `json:"count" binding:"required,min=1"`
}
