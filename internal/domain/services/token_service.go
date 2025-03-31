package services

import "time"

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
