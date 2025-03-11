package services

import (
	"context"

	"github.com/bookshop/api/internal/domain/models"
)

// AuthService определяет методы для работы с аутентификацией
type AuthService interface {
	// Register регистрирует нового пользователя
	Register(ctx context.Context, input models.UserRegistration) (*models.User, error)

	// Login аутентифицирует пользователя и возвращает токены
	Login(ctx context.Context, input models.UserCredentials) (string, string, error)

	// ValidateToken проверяет токен и возвращает ID пользователя
	ValidateToken(ctx context.Context, token string) (int, error)

	// RefreshToken обновляет токен доступа
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)

	// GetUserByID возвращает пользователя по ID
	GetUserByID(ctx context.Context, id int) (*models.User, error)

	// IsAdmin проверяет, является ли пользователь администратором
	IsAdmin(ctx context.Context, userID int) (bool, error)
}

