package models

import "time"

// User представляет модель пользователя
type User struct {
	ID           int       `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	IsAdmin      bool      `json:"is_admin" db:"is_admin"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserCredentials представляет учетные данные пользователя для аутентификации
type UserCredentials struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// UserRegistration представляет данные для регистрации пользователя
type UserRegistration struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

// UserResponse представляет данные пользователя для ответа API
type UserResponse struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse преобразует модель пользователя в ответ API
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		IsAdmin:   u.IsAdmin,
		CreatedAt: u.CreatedAt,
	}
}
