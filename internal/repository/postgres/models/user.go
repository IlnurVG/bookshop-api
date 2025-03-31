package models

import (
	"time"

	domainmodels "github.com/bookshop/api/internal/domain/models"
)

// User represents a user model for repository operations
type User struct {
	ID           int       `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	IsAdmin      bool      `db:"is_admin"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// ToDomain converts repository model to domain model
func (u *User) ToDomain() *domainmodels.User {
	return &domainmodels.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		IsAdmin:      u.IsAdmin,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// UserFromDomain converts domain model to repository model
func UserFromDomain(user *domainmodels.User) *User {
	return &User{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		IsAdmin:      user.IsAdmin,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

// UserSliceToDomain converts a slice of repository models to domain models
func UserSliceToDomain(users []User) []domainmodels.User {
	result := make([]domainmodels.User, len(users))
	for i, user := range users {
		domainUser := user.ToDomain()
		result[i] = *domainUser
	}
	return result
}
