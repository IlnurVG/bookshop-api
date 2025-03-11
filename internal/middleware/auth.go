package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
)

// JWTConfig содержит настройки для JWT аутентификации
type JWTConfig struct {
	SecretKey string
}

// NewJWTConfig создает новый экземпляр JWTConfig
func NewJWTConfig(secretKey string) *JWTConfig {
	return &JWTConfig{
		SecretKey: secretKey,
	}
}

// AuthMiddleware создает middleware для проверки JWT токена
func AuthMiddleware(config *JWTConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Получаем токен из заголовка Authorization
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "отсутствует токен авторизации"})
			}

			// Проверяем формат токена
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "неверный формат токена"})
			}

			// Парсим токен
			token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
				// Проверяем метод подписи
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "неверный метод подписи")
				}
				return []byte(config.SecretKey), nil
			})

			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "неверный токен: " + err.Error()})
			}

			// Проверяем валидность токена
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				// Получаем ID пользователя из токена
				userID, ok := claims["user_id"].(float64)
				if !ok {
					return c.JSON(http.StatusUnauthorized, map[string]string{"error": "неверный формат ID пользователя"})
				}

				// Сохраняем ID пользователя в контексте
				c.Set("userID", int(userID))

				// Проверяем роль пользователя, если она есть
				if role, ok := claims["role"].(string); ok {
					c.Set("userRole", role)
				}

				return next(c)
			}

			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "неверный токен"})
		}
	}
}

// AdminMiddleware создает middleware для проверки роли администратора
func AdminMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Получаем роль пользователя из контекста
			role, ok := c.Get("userRole").(string)
			if !ok || role != "admin" {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "доступ запрещен"})
			}

			return next(c)
		}
	}
}
