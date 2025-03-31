package middleware

import "github.com/labstack/echo/v4"

// CORS adds CORS headers to the response
func CORS() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get origin from request header
			origin := c.Request().Header.Get("Origin")
			if origin == "" {
				return next(c)
			}

			// Set CORS headers
			c.Response().Header().Set("Access-Control-Allow-Origin", origin)
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Response().Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
			if c.Request().Method == "OPTIONS" {
				return c.NoContent(204)
			}

			return next(c)
		}
	}
}
