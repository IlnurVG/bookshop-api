package cart

import (
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/bookshop/api/internal/handlers"
	"github.com/bookshop/api/pkg/logger"
	"github.com/labstack/echo/v4"
)

// Module represents a cart management module
type Module struct {
	Handler *handlers.CartHandler
	Service services.CartService
}

// NewModule creates a new instance of the cart module
func NewModule(
	cartRepo repositories.CartRepository,
	bookRepo repositories.BookRepository,
	txManager repositories.TransactionManager,
	logger logger.Logger,
) *Module {
	// Create service
	service := NewService(cartRepo, bookRepo, txManager, logger)

	// Create handler
	handler := handlers.NewCartHandler(service)

	return &Module{
		Handler: handler,
		Service: service,
	}
}

// RegisterRoutes registers routes for cart request handling
func (m *Module) RegisterRoutes(router *echo.Group) {
	m.Handler.RegisterRoutes(router)
}
