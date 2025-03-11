package book

import (
	"github.com/go-playground/validator/v10"
)

// RegisterValidators регистрирует пользовательские валидаторы для книг
func RegisterValidators(v *validator.Validate) {
	// Здесь можно добавить пользовательские валидаторы для книг
	// Например, валидатор для проверки ISBN, формата автора и т.д.

	// Пример регистрации валидатора для проверки ISBN
	// v.RegisterValidation("isbn", validateISBN)
}

// validateISBN проверяет корректность ISBN
// func validateISBN(fl validator.FieldLevel) bool {
// 	isbn := fl.Field().String()
// 	// Реализация проверки ISBN
// 	return true
// }
