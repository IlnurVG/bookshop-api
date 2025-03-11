package server

import (
	"net/http"

	"github.com/bookshop/api/internal/middleware"
	"github.com/labstack/echo/v4"
)

// registerRoutes регистрирует все маршруты API
func (s *Server) registerRoutes() {
	e := s.echo

	// Группа для API v1
	v1 := e.Group("/api/v1")

	// Публичные маршруты
	public := v1.Group("")
	public.GET("/health", s.HealthCheck)

	// Регистрируем маршруты для книг
	s.bookModule.RegisterRoutes(v1)

	// Маршруты для аутентификации
	auth := public.Group("/auth")
	auth.POST("/register", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Регистрация"})
	})
	auth.POST("/login", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Вход"})
	})

	// Создаем конфигурацию JWT
	jwtConfig := middleware.NewJWTConfig(s.config.JWT.Secret)

	// Защищенные маршруты (требуют аутентификации)
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(jwtConfig))

	// Маршруты для корзины
	cart := protected.Group("/cart")
	cart.GET("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Корзина пользователя"})
	})
	cart.POST("/items", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Добавление товара в корзину"})
	})
	cart.DELETE("/items/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Удаление товара из корзины"})
	})

	// Маршруты для оформления заказа
	s.checkoutHandler.RegisterRoutes(protected)

	// Маршруты для администраторов
	admin := protected.Group("/admin")
	admin.Use(middleware.AdminMiddleware())

	// Управление категориями
	adminCategories := admin.Group("/categories")
	adminCategories.POST("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Создание категории"})
	})
	adminCategories.PUT("/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Обновление категории"})
	})
	adminCategories.DELETE("/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Удаление категории"})
	})

	// Управление книгами
	adminBooks := admin.Group("/books")
	adminBooks.POST("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Создание книги"})
	})
	adminBooks.PUT("/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Обновление книги"})
	})
	adminBooks.DELETE("/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Удаление книги"})
	})
}
