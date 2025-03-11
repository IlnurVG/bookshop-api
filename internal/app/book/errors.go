package book

import (
	"errors"
	"fmt"
)

// Определение ошибок, связанных с книгами
var (
	// ErrBookNotFound возвращается, когда книга не найдена
	ErrBookNotFound = errors.New("книга не найдена")

	// ErrInvalidBookID возвращается при некорректном ID книги
	ErrInvalidBookID = errors.New("некорректный ID книги")

	// ErrInvalidBookData возвращается при некорректных данных книги
	ErrInvalidBookData = errors.New("некорректные данные книги")

	// ErrCategoryNotFound возвращается, когда категория не найдена
	ErrCategoryNotFound = errors.New("категория не найдена")

	// ErrBookOutOfStock возвращается, когда книга отсутствует на складе
	ErrBookOutOfStock = errors.New("книга отсутствует на складе")
)

// BookError представляет ошибку, связанную с книгой
type BookError struct {
	ID  int
	Err error
}

// Error возвращает строковое представление ошибки
func (e *BookError) Error() string {
	return fmt.Sprintf("ошибка книги (ID: %d): %v", e.ID, e.Err)
}

// Unwrap возвращает оригинальную ошибку
func (e *BookError) Unwrap() error {
	return e.Err
}

// NewBookError создает новую ошибку книги
func NewBookError(id int, err error) *BookError {
	return &BookError{
		ID:  id,
		Err: err,
	}
}
