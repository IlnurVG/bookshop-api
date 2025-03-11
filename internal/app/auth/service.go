package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/bookshop/api/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// Определение ошибок
var (
	ErrInvalidCredentials = errors.New("неверный email или пароль")
	ErrUserAlreadyExists  = errors.New("пользователь с таким email уже существует")
	ErrInvalidToken       = errors.New("недействительный токен")
	ErrTokenExpired       = errors.New("срок действия токена истек")
)

const (
	// AccessTokenTTL время жизни access токена
	AccessTokenTTL = 15 * time.Minute
	// RefreshTokenTTL время жизни refresh токена
	RefreshTokenTTL = 24 * 7 * time.Hour
	// DefaultCost стоимость хеширования пароля
	DefaultCost = 10
)

// Service реализует интерфейс services.AuthService
type Service struct {
	userRepo repositories.UserRepository
	tokenMgr TokenManager
	logger   logger.Logger
}

// TokenManager определяет методы для работы с токенами
type TokenManager interface {
	// CreateToken создает новый токен
	CreateToken(userID int, isAdmin bool, ttl time.Duration) (string, error)
	// ValidateToken проверяет токен и возвращает ID пользователя
	ValidateToken(token string) (int, error)
	// ParseToken разбирает токен и возвращает информацию о нем
	ParseToken(token string) (*TokenClaims, error)
}

// TokenClaims представляет данные токена
type TokenClaims struct {
	UserID  int
	IsAdmin bool
	Exp     time.Time
}

// NewService создает новый экземпляр сервиса аутентификации
func NewService(
	userRepo repositories.UserRepository,
	tokenMgr TokenManager,
	logger logger.Logger,
) services.AuthService {
	return &Service{
		userRepo: userRepo,
		tokenMgr: tokenMgr,
		logger:   logger,
	}
}

// Register регистрирует нового пользователя
func (s *Service) Register(ctx context.Context, input models.UserRegistration) (*models.User, error) {
	// Проверяем совпадение паролей
	if input.Password != input.ConfirmPassword {
		return nil, fmt.Errorf("пароли не совпадают")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("ошибка хеширования пароля: %w", err)
	}

	// Создаем пользователя
	user := &models.User{
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Сохраняем пользователя
	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repositories.ErrDuplicateKey) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	return user, nil
}

// Login аутентифицирует пользователя и возвращает токены
func (s *Service) Login(ctx context.Context, input models.UserCredentials) (string, string, error) {
	// Получаем пользователя по email
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return "", "", ErrInvalidCredentials
		}
		return "", "", fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return "", "", ErrInvalidCredentials
	}

	// Создаем access токен
	accessToken, err := s.tokenMgr.CreateToken(user.ID, user.IsAdmin, AccessTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("ошибка создания access токена: %w", err)
	}

	// Создаем refresh токен
	refreshToken, err := s.tokenMgr.CreateToken(user.ID, user.IsAdmin, RefreshTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("ошибка создания refresh токена: %w", err)
	}

	return accessToken, refreshToken, nil
}

// ValidateToken проверяет токен и возвращает ID пользователя
func (s *Service) ValidateToken(ctx context.Context, token string) (int, error) {
	return s.tokenMgr.ValidateToken(token)
}

// RefreshToken обновляет токен доступа
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	// Проверяем refresh токен
	claims, err := s.tokenMgr.ParseToken(refreshToken)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	// Проверяем срок действия токена
	if time.Now().After(claims.Exp) {
		return "", "", ErrTokenExpired
	}

	// Получаем пользователя
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return "", "", ErrInvalidToken
		}
		return "", "", fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	// Создаем новые токены
	accessToken, err := s.tokenMgr.CreateToken(user.ID, user.IsAdmin, AccessTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("ошибка создания access токена: %w", err)
	}

	newRefreshToken, err := s.tokenMgr.CreateToken(user.ID, user.IsAdmin, RefreshTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("ошибка создания refresh токена: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

// GetUserByID возвращает пользователя по ID
func (s *Service) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}
	return user, nil
}

// IsAdmin проверяет, является ли пользователь администратором
func (s *Service) IsAdmin(ctx context.Context, userID int) (bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return false, fmt.Errorf("пользователь не найден")
		}
		return false, fmt.Errorf("ошибка получения пользователя: %w", err)
	}
	return user.IsAdmin, nil
}
