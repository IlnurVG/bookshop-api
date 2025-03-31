package models

import (
	"time"

	domainmodels "github.com/bookshop/api/internal/domain/models"
)

// User represents a user model for service operations
type User struct {
	ID           int
	Email        string
	PasswordHash string
	IsAdmin      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserCredentials represents user authentication credentials
type UserCredentials struct {
	Email    string
	Password string
}

// UserRegistration represents data for user registration
type UserRegistration struct {
	Email           string
	Password        string
	ConfirmPassword string
}

// ToDomain converts service model to domain model
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

// FromDomain converts domain model to service model
func FromDomain(user *domainmodels.User) *User {
	return &User{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		IsAdmin:      user.IsAdmin,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

// UserSliceFromDomain converts a slice of domain models to service models
func UserSliceFromDomain(users []domainmodels.User) []User {
	result := make([]User, len(users))
	for i, user := range users {
		userCopy := user // Create a copy to avoid issues with loop variable references
		serviceUser := FromDomain(&userCopy)
		result[i] = *serviceUser
	}
	return result
}

// UserCredentialsToDomain converts service credentials to domain credentials
func (uc *UserCredentials) ToDomain() domainmodels.UserCredentials {
	return domainmodels.UserCredentials{
		Email:    uc.Email,
		Password: uc.Password,
	}
}

// UserRegistrationToDomain converts service registration data to domain registration data
func (ur *UserRegistration) ToDomain() domainmodels.UserRegistration {
	return domainmodels.UserRegistration{
		Email:           ur.Email,
		Password:        ur.Password,
		ConfirmPassword: ur.ConfirmPassword,
	}
}
