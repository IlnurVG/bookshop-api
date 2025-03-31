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

// Error definitions
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token has expired")
)

const (
	// AccessTokenTTL lifetime of the access token
	AccessTokenTTL = 15 * time.Minute
	// RefreshTokenTTL lifetime of the refresh token
	RefreshTokenTTL = 24 * 7 * time.Hour
	// DefaultCost cost of password hashing
	DefaultCost = 10
)

// Service implements services.AuthService interface
type Service struct {
	userRepo repositories.UserRepository
	tokenMgr TokenManager
	logger   logger.Logger
}

// TokenManager defines methods for working with tokens
type TokenManager interface {
	// CreateToken creates a new token
	CreateToken(userID int, isAdmin bool, ttl time.Duration) (string, error)
	// ValidateToken validates the token and returns user ID
	ValidateToken(token string) (int, error)
	// ParseToken parses the token and returns its information
	ParseToken(token string) (*TokenClaims, error)
}

// TokenClaims represents token data
type TokenClaims struct {
	UserID  int
	IsAdmin bool
	Exp     time.Time
}

// NewService creates a new instance of the authentication service
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

// Register registers a new user
func (s *Service) Register(ctx context.Context, input models.UserRegistration) (*models.User, error) {
	// Check if passwords match
	if input.Password != input.ConfirmPassword {
		return nil, fmt.Errorf("passwords do not match")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Create user
	user := &models.User{
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save user
	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repositories.ErrDuplicateKey) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return user, nil
}

// Login authenticates a user and returns tokens
func (s *Service) Login(ctx context.Context, input models.UserCredentials) (string, string, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return "", "", ErrInvalidCredentials
		}
		return "", "", fmt.Errorf("error getting user: %w", err)
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return "", "", ErrInvalidCredentials
	}

	// Create access token
	accessToken, err := s.tokenMgr.CreateToken(user.ID, user.IsAdmin, AccessTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("error creating access token: %w", err)
	}

	// Create refresh token
	refreshToken, err := s.tokenMgr.CreateToken(user.ID, user.IsAdmin, RefreshTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("error creating refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// ValidateToken validates the token and returns user ID
func (s *Service) ValidateToken(ctx context.Context, token string) (int, error) {
	return s.tokenMgr.ValidateToken(token)
}

// RefreshToken refreshes the access token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	// Verify refresh token
	claims, err := s.tokenMgr.ParseToken(refreshToken)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	// Check token expiration
	if time.Now().After(claims.Exp) {
		return "", "", ErrTokenExpired
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return "", "", ErrInvalidToken
		}
		return "", "", fmt.Errorf("error getting user: %w", err)
	}

	// Create new tokens
	accessToken, err := s.tokenMgr.CreateToken(user.ID, user.IsAdmin, AccessTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("error creating access token: %w", err)
	}

	newRefreshToken, err := s.tokenMgr.CreateToken(user.ID, user.IsAdmin, RefreshTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("error creating refresh token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

// GetUserByID returns a user by ID
func (s *Service) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	return user, nil
}

// IsAdmin checks if a user is an administrator
func (s *Service) IsAdmin(ctx context.Context, userID int) (bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return false, fmt.Errorf("user not found")
		}
		return false, fmt.Errorf("error getting user: %w", err)
	}
	return user.IsAdmin, nil
}
