package errors

import (
	"fmt"
	"net/http"
)

// HTTPError represents an HTTP-specific error
type HTTPError struct {
	Code    int
	Message string
	Err     error
}

// Error returns the error message
func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the original error
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// NewHTTPError creates a new HTTP error
func NewHTTPError(code int, message string, err error) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewNotFoundError creates a "not found" error
func NewNotFoundError(err error) *HTTPError {
	return NewHTTPError(http.StatusNotFound, "not found", err)
}

// NewBadRequestError creates a "bad request" error
func NewBadRequestError(err error) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, "bad request", err)
}

// NewUnauthorizedError creates an "unauthorized" error
func NewUnauthorizedError(err error) *HTTPError {
	return NewHTTPError(http.StatusUnauthorized, "unauthorized", err)
}

// NewForbiddenError creates a "forbidden" error
func NewForbiddenError(err error) *HTTPError {
	return NewHTTPError(http.StatusForbidden, "forbidden", err)
}

// NewInternalServerError creates an "internal server error"
func NewInternalServerError(err error) *HTTPError {
	return NewHTTPError(http.StatusInternalServerError, "internal server error", err)
}

// NewConflictError creates a "conflict" error
func NewConflictError(err error) *HTTPError {
	return NewHTTPError(http.StatusConflict, "conflict", err)
}
