package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator represents a validator for structs
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	validate := validator.New()

	// Register function to get field names from json tags
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{validate: validate}
}

// Validate validates a struct according to validation rules
func (v *Validator) Validate(i interface{}) error {
	if err := v.validate.Struct(i); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return &ValidationErrors{Errors: validationErrors}
		}
		return err
	}
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field string
	Tag   string
	Value interface{}
}

// ValidationErrors represents a list of validation errors
type ValidationErrors struct {
	Errors validator.ValidationErrors
}

// Error returns a string representation of validation errors
func (e *ValidationErrors) Error() string {
	return formatValidationErrors(e.Errors)
}

// formatValidationErrors formats validation errors into a readable format
func formatValidationErrors(errors validator.ValidationErrors) string {
	var messages []string
	for _, err := range errors {
		field := err.Field()
		tag := err.Tag()
		value := err.Value()
		message := getErrorMessage(field, tag, value)
		messages = append(messages, message)
	}
	return strings.Join(messages, "; ")
}

// getErrorMessage returns an error message based on the validation tag
func getErrorMessage(field, tag string, value interface{}) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %v characters", field, value)
	case "max":
		return fmt.Sprintf("%s must be at most %v characters", field, value)
	case "numeric":
		return fmt.Sprintf("%s must be numeric", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
