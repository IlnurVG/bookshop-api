package errors

import "errors"

var (
	// ErrInvalidCredentials indicates that the provided email or password is incorrect
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrUserAlreadyExists indicates that a user with the given email already exists
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	// ErrInvalidToken indicates that the provided token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired indicates that the provided token has expired
	ErrTokenExpired = errors.New("token has expired")
)
