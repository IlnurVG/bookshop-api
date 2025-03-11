package book

import (
	"github.com/bookshop/api/internal/domain/repositories"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/labstack/echo/v4"
)

// Module представляет модуль для работы с книгами
type Module struct {
	Handler *Handler
	Service services.BookService
}

// NewModule создает новый экземпляр модуля для работы с книгами
func NewModule(
	bookRepo repositories.BookRepository,
	categoryRepo repositories.CategoryRepository,
) *Module {
	// Создаем сервис
	service := NewService(bookRepo, categoryRepo)

	// Создаем обработчик
	handler := NewHandler(service)

	return &Module{
		Handler: handler,
		Service: service,
	}
}

// RegisterRoutes регистрирует маршруты для обработки запросов к книгам
func (m *Module) RegisterRoutes(router *echo.Group) {
	m.Handler.RegisterRoutes(router)
}
