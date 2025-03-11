package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Стандартные ошибки приложения
var (
	ErrNotFound           = errors.New("ресурс не найден")
	ErrBadRequest         = errors.New("некорректный запрос")
	ErrUnauthorized       = errors.New("не авторизован")
	ErrForbidden          = errors.New("доступ запрещен")
	ErrInternalServer     = errors.New("внутренняя ошибка сервера")
	ErrConflict           = errors.New("конфликт данных")
	ErrInvalidCredentials = errors.New("неверные учетные данные")
	ErrOutOfStock         = errors.New("товар отсутствует на складе")
	ErrCartExpired        = errors.New("корзина истекла")
)

// AppError представляет ошибку приложения
type AppError struct {
	Err     error
	Message string
	Code    int
}

// Error возвращает сообщение об ошибке
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap возвращает оригинальную ошибку
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError создает новую ошибку приложения
func NewAppError(err error, message string, code int) *AppError {
	return &AppError{
		Err:     err,
		Message: message,
		Code:    code,
	}
}

// NewNotFoundError создает ошибку "не найдено"
func NewNotFoundError(message string) *AppError {
	return NewAppError(ErrNotFound, message, http.StatusNotFound)
}

// NewBadRequestError создает ошибку "некорректный запрос"
func NewBadRequestError(message string) *AppError {
	return NewAppError(ErrBadRequest, message, http.StatusBadRequest)
}

// NewUnauthorizedError создает ошибку "не авторизован"
func NewUnauthorizedError(message string) *AppError {
	return NewAppError(ErrUnauthorized, message, http.StatusUnauthorized)
}

// NewForbiddenError создает ошибку "доступ запрещен"
func NewForbiddenError(message string) *AppError {
	return NewAppError(ErrForbidden, message, http.StatusForbidden)
}

// NewInternalServerError создает ошибку "внутренняя ошибка сервера"
func NewInternalServerError(err error) *AppError {
	return NewAppError(err, "внутренняя ошибка сервера", http.StatusInternalServerError)
}

// NewConflictError создает ошибку "конфликт данных"
func NewConflictError(message string) *AppError {
	return NewAppError(ErrConflict, message, http.StatusConflict)
}

// Is проверяет, является ли ошибка определенного типа
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As устанавливает target в первую ошибку в цепочке err, которая соответствует типу target
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Wrap оборачивает ошибку с дополнительным сообщением
func Wrap(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}
