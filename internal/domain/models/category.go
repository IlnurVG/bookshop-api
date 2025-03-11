package models

import "time"

// Category представляет модель категории книг
type Category struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CategoryCreate представляет данные для создания категории
type CategoryCreate struct {
	Name string `json:"name" validate:"required"`
}

// CategoryUpdate представляет данные для обновления категории
type CategoryUpdate struct {
	Name string `json:"name" validate:"required"`
}

// CategoryResponse представляет ответ с категорией
type CategoryResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ToResponse преобразует модель категории в ответ API
func (c *Category) ToResponse() CategoryResponse {
	return CategoryResponse{
		ID:   c.ID,
		Name: c.Name,
	}
}
