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
	customMiddleware "github.com/bookshop/api/internal/middleware"
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
	cartService     services.CartService
	cartHandler     *handlers.CartHandler
	bookModule      *book.Module
	ipRateLimiter   *customMiddleware.IPRateLimiter   // IP-based rate limiter
	pathRateLimiter *customMiddleware.PathRateLimiter // Path-based rate limiter
}

// NewServer creates a new instance of HTTP server
func NewServer(
	cfg *config.Config,
	logger *logger.Logger,
	checkoutService services.CheckoutService,
	cartService services.CartService,
	bookRepo repositories.BookRepository,
	categoryRepo repositories.CategoryRepository,
	txManager repositories.TransactionManager,
) (*Server, error) {
	e := echo.New()
	e.HideBanner = true

	// Create rate limiters if enabled in config
	var ipRateLimiter *customMiddleware.IPRateLimiter
	var pathRateLimiter *customMiddleware.PathRateLimiter

	if cfg.RateLimit.Enabled {
		// Create IP-based rate limiter
		ipRateLimiter = customMiddleware.NewIPRateLimiter(cfg.RateLimit.GlobalIPLimit, *logger)

		// Create path-based rate limiter
		pathRateLimiter = customMiddleware.NewPathRateLimiter(cfg.RateLimit.DefaultPathLimit, *logger)

		// Set path-specific rate limits from configuration
		for path, limit := range cfg.RateLimit.Endpoints {
			pathRateLimiter.SetPathLimit(path, limit)
		}
	}

	// Middleware setup
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())

	// Add rate limiting middleware if enabled
	if cfg.RateLimit.Enabled {
		// First limit by IP to protect against DoS attacks
		e.Use(ipRateLimiter.Middleware())

		// Then limit by path to protect sensitive endpoints
		e.Use(pathRateLimiter.Middleware())
	}

	// Server address configuration
	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)

	// Timeout configuration
	e.Server.ReadTimeout = cfg.HTTP.ReadTimeout
	e.Server.WriteTimeout = cfg.HTTP.WriteTimeout
	e.Server.IdleTimeout = cfg.HTTP.IdleTimeout

	// Handlers initialization
	checkoutHandler := handlers.NewCheckoutHandler(checkoutService)
	cartHandler := handlers.NewCartHandler(cartService)

	// Book module initialization
	bookModule := book.NewModule(bookRepo, categoryRepo, txManager)

	server := &Server{
		echo:            e,
		config:          cfg,
		logger:          logger,
		Addr:            addr,
		checkoutService: checkoutService,
		checkoutHandler: checkoutHandler,
		cartService:     cartService,
		cartHandler:     cartHandler,
		bookModule:      bookModule,
		ipRateLimiter:   ipRateLimiter,   // Save rate limiter for cleanup during shutdown
		pathRateLimiter: pathRateLimiter, // Save rate limiter for cleanup during shutdown
	}

	// Route registration
	server.registerRoutes()

	return server, nil
}

// Start launches the HTTP server
func (s *Server) Start() error {
	if s.config.RateLimit.Enabled {
		s.logger.Info("Rate limiting enabled",
			"global_ip_limit", s.config.RateLimit.GlobalIPLimit,
			"default_path_limit", s.config.RateLimit.DefaultPathLimit)
	}
	return s.echo.Start(s.Addr)
}

// Shutdown stops the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	// Stop rate limiters if they were initialized
	if s.ipRateLimiter != nil {
		s.ipRateLimiter.Stop()
	}

	if s.pathRateLimiter != nil {
		s.pathRateLimiter.Stop()
	}

	// Stop the server
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
