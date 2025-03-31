package handlers

import (
	"net/http"
	"strconv"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/labstack/echo/v4"
)

// CategoryHandler handles requests related to book categories
type CategoryHandler struct {
	categoryService services.CategoryService
}

// NewCategoryHandler creates a new instance of CategoryHandler
func NewCategoryHandler(categoryService services.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// RegisterRoutes registers routes for handling category requests
func (h *CategoryHandler) RegisterRoutes(router *echo.Group) {
	// Public routes
	categories := router.Group("/categories")
	categories.GET("", h.listCategories)
	categories.GET("/:id", h.getCategory)

	// Admin routes
	admin := router.Group("/admin/categories")
	admin.POST("", h.createCategory)
	admin.PUT("/:id", h.updateCategory)
	admin.DELETE("/:id", h.deleteCategory)
}

// listCategories handles the request to get a list of categories
// @Summary Get category list
// @Description Returns a list of all categories
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {array} models.Category
// @Failure 500 {object} ErrorResponse
// @Router /categories [get]
func (h *CategoryHandler) listCategories(c echo.Context) error {
	// Get list of categories
	categories, err := h.categoryService.List(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, categories)
}

// getCategory handles the request to get category information
// @Summary Get category
// @Description Returns category information by ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} models.Category
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /categories/{id} [get]
func (h *CategoryHandler) getCategory(c echo.Context) error {
	// Get category ID from request parameters
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid category ID"})
	}

	// Get category
	category, err := h.categoryService.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "category not found"})
	}

	return c.JSON(http.StatusOK, category)
}

// createCategory handles the request to create a category
// @Summary Create category
// @Description Creates a new category
// @Tags admin,categories
// @Accept json
// @Produce json
// @Param category body models.CategoryCreate true "Category data"
// @Success 201 {object} models.Category
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/categories [post]
func (h *CategoryHandler) createCategory(c echo.Context) error {
	var req models.CategoryCreate
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request format"})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Create category
	category, err := h.categoryService.Create(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, category)
}

// updateCategory handles the request to update a category
// @Summary Update category
// @Description Updates category data
// @Tags admin,categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body models.CategoryUpdate true "Category data"
// @Success 200 {object} models.Category
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/categories/{id} [put]
func (h *CategoryHandler) updateCategory(c echo.Context) error {
	// Get category ID from request parameters
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid category ID"})
	}

	var req models.CategoryUpdate
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request format"})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Update category
	category, err := h.categoryService.Update(c.Request().Context(), id, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, category)
}

// deleteCategory handles the request to delete a category
// @Summary Delete category
// @Description Deletes a category by ID
// @Tags admin,categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/categories/{id} [delete]
func (h *CategoryHandler) deleteCategory(c echo.Context) error {
	// Get category ID from request parameters
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid category ID"})
	}

	// Delete category
	if err := h.categoryService.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}
