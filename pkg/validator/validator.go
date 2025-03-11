package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator представляет валидатор для структур
type Validator struct {
	validate *validator.Validate
}

// NewValidator создает новый экземпляр валидатора
func NewValidator() *Validator {
	v := validator.New()

	// Регистрация функции для получения имен полей из json тегов
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{
		validate: v,
	}
}

// Validate проверяет структуру на соответствие правилам валидации
func (v *Validator) Validate(i interface{}) error {
	if err := v.validate.Struct(i); err != nil {
		return formatValidationErrors(err)
	}
	return nil
}

// ValidationError представляет ошибку валидации
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors представляет список ошибок валидации
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error возвращает строковое представление ошибок валидации
func (ve ValidationErrors) Error() string {
	var sb strings.Builder
	sb.WriteString("ошибки валидации: ")
	for i, err := range ve.Errors {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return sb.String()
}

// formatValidationErrors форматирует ошибки валидации в удобный формат
func formatValidationErrors(err error) error {
	if err == nil {
		return nil
	}

	var validationErrors ValidationErrors

	for _, err := range err.(validator.ValidationErrors) {
		field := err.Field()
		message := getErrorMessage(err)

		validationErrors.Errors = append(validationErrors.Errors, ValidationError{
			Field:   field,
			Message: message,
		})
	}

	return validationErrors
}

// getErrorMessage возвращает сообщение об ошибке в зависимости от тега валидации
func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "обязательное поле"
	case "email":
		return "некорректный email"
	case "min":
		return fmt.Sprintf("минимальное значение: %s", err.Param())
	case "max":
		return fmt.Sprintf("максимальное значение: %s", err.Param())
	case "len":
		return fmt.Sprintf("требуемая длина: %s", err.Param())
	case "gt":
		return fmt.Sprintf("должно быть больше %s", err.Param())
	case "gte":
		return fmt.Sprintf("должно быть больше или равно %s", err.Param())
	case "lt":
		return fmt.Sprintf("должно быть меньше %s", err.Param())
	case "lte":
		return fmt.Sprintf("должно быть меньше или равно %s", err.Param())
	default:
		return fmt.Sprintf("не соответствует правилу: %s", err.Tag())
	}
}
