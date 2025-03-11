package book

import (
	"net/http"
	"strconv"

	"github.com/bookshop/api/internal/domain/services"
	"github.com/labstack/echo/v4"
)

// Handler обрабатывает HTTP-запросы для работы с книгами
type Handler struct {
	bookService services.BookService
}

// NewHandler создает новый экземпляр обработчика для книг
func NewHandler(bookService services.BookService) *Handler {
	return &Handler{
		bookService: bookService,
	}
}

// RegisterRoutes регистрирует маршруты для обработки запросов к книгам
func (h *Handler) RegisterRoutes(router *echo.Group) {
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
// @Param book body CreateBookRequest true "Данные книги"
// @Success 201 {object} BookResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/books [post]
func (h *Handler) createBook(c echo.Context) error {
	var req CreateBookRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "неверный формат запроса"})
	}

	// Валидация запроса
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Преобразуем запрос в модель
	input := req.ToModel()

	// Создаем книгу
	book, err := h.bookService.Create(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Преобразуем модель в ответ
	response := FromModel(book)

	return c.JSON(http.StatusCreated, response)
}

// getBook обрабатывает запрос на получение информации о книге
// @Summary Получить книгу
// @Description Возвращает информацию о книге по ID
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "ID книги"
// @Success 200 {object} BookResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books/{id} [get]
func (h *Handler) getBook(c echo.Context) error {
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

	// Преобразуем модель в ответ
	response := FromModel(book)

	return c.JSON(http.StatusOK, response)
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
// @Success 200 {object} BookListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books [get]
func (h *Handler) listBooks(c echo.Context) error {
	var req BookListRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "неверный формат запроса"})
	}

	// Преобразуем запрос в модель
	filter := req.ToModel()

	// Получаем список книг
	books, err := h.bookService.List(c.Request().Context(), filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Преобразуем модель в ответ
	response := FromModelList(books)

	return c.JSON(http.StatusOK, response)
}

// updateBook обрабатывает запрос на обновление книги
// @Summary Обновить книгу
// @Description Обновляет данные книги
// @Tags admin,books
// @Accept json
// @Produce json
// @Param id path int true "ID книги"
// @Param book body UpdateBookRequest true "Данные книги"
// @Success 200 {object} BookResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/books/{id} [put]
func (h *Handler) updateBook(c echo.Context) error {
	// Получаем ID книги из параметров запроса
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "некорректный ID книги"})
	}

	var req UpdateBookRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "неверный формат запроса"})
	}

	// Валидация запроса
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Преобразуем запрос в модель
	input := req.ToModel()

	// Обновляем книгу
	book, err := h.bookService.Update(c.Request().Context(), id, input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Преобразуем модель в ответ
	response := FromModel(book)

	return c.JSON(http.StatusOK, response)
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
func (h *Handler) deleteBook(c echo.Context) error {
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

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}
