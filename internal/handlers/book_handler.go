package handlers

import (
	"net/http"
	"strconv"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/labstack/echo/v4"
)

// BookHandler обрабатывает запросы, связанные с книгами
type BookHandler struct {
	bookService services.BookService
}

// NewBookHandler создает новый экземпляр BookHandler
func NewBookHandler(bookService services.BookService) *BookHandler {
	return &BookHandler{
		bookService: bookService,
	}
}

// RegisterRoutes регистрирует маршруты для обработки запросов к книгам
func (h *BookHandler) RegisterRoutes(router *echo.Group) {
	books := router.Group("/books")

	// Публичные маршруты
	books.GET("", h.listBooks)
	books.GET("/:id", h.getBook)

	// Маршруты для администраторов
	admin := router.Group("/admin/books")
	admin.POST("", h.createBook)
	admin.PUT("/:id", h.updateBook)
	admin.DELETE("/:id", h.deleteBook)
}

// createBook обрабатывает запрос на создание книги
// @Summary Создать книгу
// @Description Создает новую книгу
// @Tags admin,books
// @Accept json
// @Produce json
// @Param book body models.BookCreate true "Данные книги"
// @Success 201 {object} models.Book
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/books [post]
func (h *BookHandler) createBook(c echo.Context) error {
	var req models.BookCreate
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "неверный формат запроса"})
	}

	// Валидация запроса
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Создаем книгу
	book, err := h.bookService.Create(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, book)
}

// getBook обрабатывает запрос на получение информации о книге
// @Summary Получить книгу
// @Description Возвращает информацию о книге по ID
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "ID книги"
// @Success 200 {object} models.Book
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books/{id} [get]
func (h *BookHandler) getBook(c echo.Context) error {
	// Получаем ID книги из параметров запроса
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "некорректный ID книги"})
	}

	// Получаем книгу
	book, err := h.bookService.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "книга не найдена"})
	}

	return c.JSON(http.StatusOK, book)
}

// listBooks обрабатывает запрос на получение списка книг
// @Summary Получить список книг
// @Description Возвращает список книг с фильтрацией
// @Tags books
// @Accept json
// @Produce json
// @Param category_ids query []int false "ID категорий"
// @Param min_price query number false "Минимальная цена"
// @Param max_price query number false "Максимальная цена"
// @Param in_stock query bool false "Только в наличии"
// @Param page query int false "Номер страницы"
// @Param page_size query int false "Размер страницы"
// @Success 200 {object} models.BookListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books [get]
func (h *BookHandler) listBooks(c echo.Context) error {
	// Создаем фильтр
	filter := models.BookFilter{
		Page:     1,
		PageSize: 10,
	}

	// Получаем параметры из запроса
	if page, err := strconv.Atoi(c.QueryParam("page")); err == nil && page > 0 {
		filter.Page = page
	}

	if pageSize, err := strconv.Atoi(c.QueryParam("page_size")); err == nil && pageSize > 0 {
		filter.PageSize = pageSize
	}

	// Получаем категории
	if categoryIDs := c.QueryParam("category_ids"); categoryIDs != "" {
		// Парсим список ID категорий
		for _, idStr := range c.QueryParams()["category_ids"] {
			if id, err := strconv.Atoi(idStr); err == nil && id > 0 {
				filter.CategoryIDs = append(filter.CategoryIDs, id)
			}
		}
	}

	// Получаем диапазон цен
	if minPrice, err := strconv.ParseFloat(c.QueryParam("min_price"), 64); err == nil && minPrice >= 0 {
		filter.MinPrice = &minPrice
	}

	if maxPrice, err := strconv.ParseFloat(c.QueryParam("max_price"), 64); err == nil && maxPrice >= 0 {
		filter.MaxPrice = &maxPrice
	}

	// Получаем параметр наличия на складе
	if inStock := c.QueryParam("in_stock"); inStock != "" {
		if inStock == "true" {
			inStockValue := true
			filter.InStock = &inStockValue
		} else if inStock == "false" {
			inStockValue := false
			filter.InStock = &inStockValue
		}
	}

	// Получаем список книг
	books, err := h.bookService.List(c.Request().Context(), filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, books)
}

// updateBook обрабатывает запрос на обновление книги
// @Summary Обновить книгу
// @Description Обновляет данные книги
// @Tags admin,books
// @Accept json
// @Produce json
// @Param id path int true "ID книги"
// @Param book body models.BookUpdate true "Данные книги"
// @Success 200 {object} models.Book
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/books/{id} [put]
func (h *BookHandler) updateBook(c echo.Context) error {
	// Получаем ID книги из параметров запроса
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "некорректный ID книги"})
	}

	var req models.BookUpdate
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "неверный формат запроса"})
	}

	// Валидация запроса
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Обновляем книгу
	book, err := h.bookService.Update(c.Request().Context(), id, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, book)
}

// deleteBook обрабатывает запрос на удаление книги
// @Summary Удалить книгу
// @Description Удаляет книгу по ID
// @Tags admin,books
// @Accept json
// @Produce json
// @Param id path int true "ID книги"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/books/{id} [delete]
func (h *BookHandler) deleteBook(c echo.Context) error {
	// Получаем ID книги из параметров запроса
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "некорректный ID книги"})
	}

	// Удаляем книгу
	if err := h.bookService.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}
