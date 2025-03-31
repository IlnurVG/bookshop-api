package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bookshop/api/config"
	"github.com/bookshop/api/internal/app/book"
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/bookshop/api/internal/handlers"
	"github.com/bookshop/api/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Server represents an HTTP server
type Server struct {
	echo            *echo.Echo
	config          *config.Config
	logger          *logger.Logger
	Addr            string
	checkoutService services.CheckoutService
	checkoutHandler *handlers.CheckoutHandler
	bookModule      *book.Module
}

// NewServer creates a new instance of HTTP server
func NewServer(
	cfg *config.Config,
	logger *logger.Logger,
	checkoutService services.CheckoutService,
	bookRepo repositories.BookRepository,
	categoryRepo repositories.CategoryRepository,
) (*Server, error) {
	e := echo.New()
	e.HideBanner = true

	// Middleware setup
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())

	// Server address configuration
	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)

	// Timeout configuration
	e.Server.ReadTimeout = cfg.HTTP.ReadTimeout
	e.Server.WriteTimeout = cfg.HTTP.WriteTimeout
	e.Server.IdleTimeout = cfg.HTTP.IdleTimeout

	// Handlers initialization
	checkoutHandler := handlers.NewCheckoutHandler(checkoutService)

	// Book module initialization
	bookModule := book.NewModule(bookRepo, categoryRepo)

	server := &Server{
		echo:            e,
		config:          cfg,
		logger:          logger,
		Addr:            addr,
		checkoutService: checkoutService,
		checkoutHandler: checkoutHandler,
		bookModule:      bookModule,
	}

	// Route registration
	server.registerRoutes()

	return server, nil
}

// Start launches the HTTP server
func (s *Server) Start() error {
	return s.echo.Start(s.Addr)
}

// Shutdown stops the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}

// GetEcho returns the Echo instance
func (s *Server) GetEcho() *echo.Echo {
	return s.echo
}

// HealthCheck handler for server health checking
func (s *Server) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
