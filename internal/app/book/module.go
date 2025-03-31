package book

import (
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/labstack/echo/v4"
)

// Module represents a book management module
type Module struct {
	Handler *Handler
	Service services.BookService
}

// NewModule creates a new instance of the book module
func NewModule(
	bookRepo repositories.BookRepository,
	categoryRepo repositories.CategoryRepository,
) *Module {
	// Create service
	service := NewService(bookRepo, categoryRepo)

	// Create handler
	handler := NewHandler(service)

	return &Module{
		Handler: handler,
		Service: service,
	}
}

// RegisterRoutes registers routes for book request handling
func (m *Module) RegisterRoutes(router *echo.Group) {
	m.Handler.RegisterRoutes(router)
}
