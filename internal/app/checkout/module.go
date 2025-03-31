package checkout

import (
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/bookshop/api/internal/handlers"
	"github.com/bookshop/api/internal/service"
	"github.com/bookshop/api/pkg/logger"
	"github.com/labstack/echo/v4"
)

// Module represents a checkout management module
type Module struct {
	Handler *handlers.CheckoutHandler
	Service services.CheckoutService
}

// NewModule creates a new instance of the checkout module
func NewModule(
	orderRepo repositories.OrderRepository,
	cartRepo repositories.CartRepository,
	bookRepo repositories.BookRepository,
	txManager repositories.TransactionManager,
	logger logger.Logger,
	profileCacheService *service.ProfileCacheService,
) *Module {
	// Create service
	service := NewService(orderRepo, cartRepo, bookRepo, txManager, logger, profileCacheService)

	// Create handler
	handler := handlers.NewCheckoutHandler(service)

	return &Module{
		Handler: handler,
		Service: service,
	}
}

// RegisterRoutes registers routes for checkout request handling
func (m *Module) RegisterRoutes(router *echo.Group) {
	m.Handler.RegisterRoutes(router)
}
