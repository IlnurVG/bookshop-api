package errors

import (
	"errors"
	"fmt"
)

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

// Unwrap returns the result of calling the Unwrap method on err, if err's type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}
