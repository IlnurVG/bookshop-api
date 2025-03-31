package middleware

import (
	"github.com/bookshop/api/pkg/logger"
	"github.com/labstack/echo/v4"
)

// Logging adds request logging
func Logging(log logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Process the request
			err := next(c)

			// Log information about the request
			log.Info("Request processed")

			return err
		}
	}
}
