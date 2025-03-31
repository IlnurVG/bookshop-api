package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

// JWTConfig contains settings for JWT authentication
type JWTConfig struct {
	SecretKey string
}

// NewJWTConfig creates a new instance of JWTConfig
func NewJWTConfig(secretKey string) *JWTConfig {
	return &JWTConfig{
		SecretKey: secretKey,
	}
}

// AuthMiddleware creates middleware for JWT token validation
func AuthMiddleware(config *JWTConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization token"})
			}

			// Check token format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token format"})
			}

			// Parse token
			token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
				// Check signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid signing method")
				}
				return []byte(config.SecretKey), nil
			})

			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token: " + err.Error()})
			}

			// Verify token validity
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				// Get user ID from token
				userID, ok := claims["user_id"].(float64)
				if !ok {
					return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user ID format"})
				}

				// Save user ID in context
				c.Set("userID", int(userID))

				// Check user role if present
				if role, ok := claims["role"].(string); ok {
					c.Set("userRole", role)
				}

				return next(c)
			}

			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}
	}
}

// AdminMiddleware creates middleware for checking admin role
func AdminMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user role from context
			role, ok := c.Get("userRole").(string)
			if !ok || role != "admin" {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
			}

			return next(c)
		}
	}
}
