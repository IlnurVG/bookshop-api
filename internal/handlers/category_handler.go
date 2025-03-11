package handlers

import (
	"net/http"
	"strconv"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/labstack/echo/v4"
)

// CategoryHandler обрабатывает запросы, связанные с категориями книг
type CategoryHandler struct {
	categoryService services.CategoryService
}

// NewCategoryHandler создает новый экземпляр CategoryHandler
func NewCategoryHandler(categoryService services.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// RegisterRoutes регистрирует маршруты для обработки запросов к категориям
func (h *CategoryHandler) RegisterRoutes(router *echo.Group) {
	// Публичные маршруты
	categories := router.Group("/categories")
	categories.GET("", h.listCategories)
	categories.GET("/:id", h.getCategory)

	// Маршруты для администраторов
	admin := router.Group("/admin/categories")
	admin.POST("", h.createCategory)
	admin.PUT("/:id", h.updateCategory)
	admin.DELETE("/:id", h.deleteCategory)
}

// listCategories обрабатывает запрос на получение списка категорий
// @Summary Получить список категорий
// @Description Возвращает список всех категорий
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {array} models.Category
// @Failure 500 {object} ErrorResponse
// @Router /categories [get]
func (h *CategoryHandler) listCategories(c echo.Context) error {
	// Получаем список категорий
	categories, err := h.categoryService.List(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, categories)
}

// getCategory обрабатывает запрос на получение информации о категории
// @Summary Получить категорию
// @Description Возвращает информацию о категории по ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "ID категории"
// @Success 200 {object} models.Category
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/{id} [get]
func (h *CategoryHandler) getCategory(c echo.Context) error {
	// Получаем ID категории из параметров запроса
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "некорректный ID категории"})
	}

	// Получаем категорию
	category, err := h.categoryService.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "категория не найдена"})
	}

	return c.JSON(http.StatusOK, category)
}

// createCategory обрабатывает запрос на создание категории
// @Summary Создать категорию
// @Description Создает новую категорию
// @Tags admin,categories
// @Accept json
// @Produce json
// @Param category body models.CategoryCreate true "Данные категории"
// @Success 201 {object} models.Category
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/categories [post]
func (h *CategoryHandler) createCategory(c echo.Context) error {
	var req models.CategoryCreate
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "неверный формат запроса"})
	}

	// Валидация запроса
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Создаем категорию
	category, err := h.categoryService.Create(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, category)
}

// updateCategory обрабатывает запрос на обновление категории
// @Summary Обновить категорию
// @Description Обновляет данные категории
// @Tags admin,categories
// @Accept json
// @Produce json
// @Param id path int true "ID категории"
// @Param category body models.CategoryUpdate true "Данные категории"
// @Success 200 {object} models.Category
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/categories/{id} [put]
func (h *CategoryHandler) updateCategory(c echo.Context) error {
	// Получаем ID категории из параметров запроса
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "некорректный ID категории"})
	}

	var req models.CategoryUpdate
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "неверный формат запроса"})
	}

	// Валидация запроса
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Обновляем категорию
	category, err := h.categoryService.Update(c.Request().Context(), id, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, category)
}

// deleteCategory обрабатывает запрос на удаление категории
// @Summary Удалить категорию
// @Description Удаляет категорию по ID
// @Tags admin,categories
// @Accept json
// @Produce json
// @Param id path int true "ID категории"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/categories/{id} [delete]
func (h *CategoryHandler) deleteCategory(c echo.Context) error {
	// Получаем ID категории из параметров запроса
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "некорректный ID категории"})
	}

	// Удаляем категорию
	if err := h.categoryService.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}
