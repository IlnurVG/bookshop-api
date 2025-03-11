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

// Server представляет HTTP сервер
type Server struct {
	echo            *echo.Echo
	config          *config.Config
	logger          *logger.Logger
	Addr            string
	checkoutService services.CheckoutService
	checkoutHandler *handlers.CheckoutHandler
	bookModule      *book.Module
}

// NewServer создает новый экземпляр HTTP сервера
func NewServer(
	cfg *config.Config,
	logger *logger.Logger,
	checkoutService services.CheckoutService,
	bookRepo repositories.BookRepository,
	categoryRepo repositories.CategoryRepository,
) (*Server, error) {
	e := echo.New()
	e.HideBanner = true

	// Настройка middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())

	// Настройка адреса сервера
	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)

	// Настройка таймаутов
	e.Server.ReadTimeout = cfg.HTTP.ReadTimeout
	e.Server.WriteTimeout = cfg.HTTP.WriteTimeout
	e.Server.IdleTimeout = cfg.HTTP.IdleTimeout

	// Инициализация обработчиков
	checkoutHandler := handlers.NewCheckoutHandler(checkoutService)

	// Инициализация модуля книг
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

	// Регистрация маршрутов
	server.registerRoutes()

	return server, nil
}

// Start запускает HTTP сервер
func (s *Server) Start() error {
	return s.echo.Start(s.Addr)
}

// Shutdown останавливает HTTP сервер
func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}

// GetEcho возвращает экземпляр Echo
func (s *Server) GetEcho() *echo.Echo {
	return s.echo
}

// HealthCheck обработчик для проверки работоспособности сервера
func (s *Server) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
