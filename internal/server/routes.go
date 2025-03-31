package server

import (
	"net/http"

	"github.com/bookshop/api/internal/middleware"
	"github.com/labstack/echo/v4"
)

// registerRoutes registers all API routes
func (s *Server) registerRoutes() {
	e := s.echo

	// API v1 group
	v1 := e.Group("/api/v1")

	// Public routes
	public := v1.Group("")
	public.GET("/health", s.HealthCheck)

	// Register book routes
	s.bookModule.RegisterRoutes(v1)

	// Authentication routes
	auth := public.Group("/auth")
	auth.POST("/register", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Registration"})
	})
	auth.POST("/login", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Login"})
	})

	// Create JWT configuration
	jwtConfig := middleware.NewJWTConfig(s.config.JWT.Secret)

	// Protected routes (require authentication)
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(jwtConfig))

	// Cart routes
	s.cartHandler.RegisterRoutes(protected)

	// Checkout routes
	s.checkoutHandler.RegisterRoutes(protected)

	// Admin routes
	admin := protected.Group("/admin")
	admin.Use(middleware.AdminMiddleware())

	// Category management
	adminCategories := admin.Group("/categories")
	adminCategories.POST("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Create category"})
	})
	adminCategories.PUT("/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Update category"})
	})
	adminCategories.DELETE("/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Delete category"})
	})

	// Book management
	adminBooks := admin.Group("/books")
	adminBooks.POST("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Create book"})
	})
	adminBooks.PUT("/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Update book"})
	})
	adminBooks.DELETE("/:id", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Delete book"})
	})
}
