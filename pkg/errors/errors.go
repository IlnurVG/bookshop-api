package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Standard application errors
var (
	ErrNotFound           = errors.New("not found")
	ErrBadRequest         = errors.New("bad request")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInternalServer     = errors.New("internal server error")
	ErrConflict           = errors.New("conflict")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrOutOfStock         = errors.New("out of stock")
	ErrCartExpired        = errors.New("cart expired")
)

// AppError represents an application error
type AppError struct {
	Code    int
	Message string
	Err     error
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the original error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewNotFoundError creates a "not found" error
func NewNotFoundError(err error) *AppError {
	return NewAppError(http.StatusNotFound, "not found", err)
}

// NewBadRequestError creates a "bad request" error
func NewBadRequestError(err error) *AppError {
	return NewAppError(http.StatusBadRequest, "bad request", err)
}

// NewUnauthorizedError creates an "unauthorized" error
func NewUnauthorizedError(err error) *AppError {
	return NewAppError(http.StatusUnauthorized, "unauthorized", err)
}

// NewForbiddenError creates a "forbidden" error
func NewForbiddenError(err error) *AppError {
	return NewAppError(http.StatusForbidden, "forbidden", err)
}

// NewInternalServerError creates an "internal server error"
func NewInternalServerError(err error) *AppError {
	return NewAppError(http.StatusInternalServerError, "internal server error", err)
}

// NewConflictError creates a "conflict" error
func NewConflictError(err error) *AppError {
	return NewAppError(http.StatusConflict, "conflict", err)
}

// Is checks if the error is of a specific type
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As sets target to the first error in err's chain that matches target's type
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Wrap wraps an error with an additional message
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}
