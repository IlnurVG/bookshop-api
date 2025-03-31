package handlers

import (
	"net/http"
	"strconv"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/labstack/echo/v4"
)

// BookHandler handles requests related to books
type BookHandler struct {
	bookService services.BookService
}

// NewBookHandler creates a new instance of BookHandler
func NewBookHandler(bookService services.BookService) *BookHandler {
	return &BookHandler{
		bookService: bookService,
	}
}

// RegisterRoutes registers routes for handling book requests
func (h *BookHandler) RegisterRoutes(router *echo.Group) {
	books := router.Group("/books")

	// Public routes
	books.GET("", h.listBooks)
	books.GET("/:id", h.getBook)

	// Admin routes
	admin := router.Group("/admin/books")
	admin.POST("", h.createBook)
	admin.PUT("/:id", h.updateBook)
	admin.DELETE("/:id", h.deleteBook)
}

// createBook handles the request to create a book
// @Summary Create book
// @Description Creates a new book
// @Tags admin,books
// @Accept json
// @Produce json
// @Param book body models.BookCreate true "Book data"
// @Success 201 {object} models.Book
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/books [post]
func (h *BookHandler) createBook(c echo.Context) error {
	var req models.BookCreate
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request format"})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Create book
	book, err := h.bookService.Create(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, book)
}

// getBook handles the request to get book information
// @Summary Get book
// @Description Returns book information by ID
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Success 200 {object} models.Book
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books/{id} [get]
func (h *BookHandler) getBook(c echo.Context) error {
	// Get book ID from request parameters
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid book ID"})
	}

	// Get book
	book, err := h.bookService.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "book not found"})
	}

	return c.JSON(http.StatusOK, book)
}

// listBooks handles the request to get a list of books
// @Summary Get book list
// @Description Returns a filtered list of books
// @Tags books
// @Accept json
// @Produce json
// @Param category_ids query []int false "Category IDs"
// @Param min_price query number false "Minimum price"
// @Param max_price query number false "Maximum price"
// @Param in_stock query bool false "In stock only"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} models.BookListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books [get]
func (h *BookHandler) listBooks(c echo.Context) error {
	// Create filter
	filter := models.BookFilter{
		Page:     1,
		PageSize: 10,
	}

	// Get parameters from request
	if page, err := strconv.Atoi(c.QueryParam("page")); err == nil && page > 0 {
		filter.Page = page
	}

	if pageSize, err := strconv.Atoi(c.QueryParam("page_size")); err == nil && pageSize > 0 {
		filter.PageSize = pageSize
	}

	// Get categories
	if categoryIDs := c.QueryParam("category_ids"); categoryIDs != "" {
		// Parse category ID list
		for _, idStr := range c.QueryParams()["category_ids"] {
			if id, err := strconv.Atoi(idStr); err == nil && id > 0 {
				filter.CategoryIDs = append(filter.CategoryIDs, id)
			}
		}
	}

	// Get price range
	if minPrice, err := strconv.ParseFloat(c.QueryParam("min_price"), 64); err == nil && minPrice >= 0 {
		filter.MinPrice = &minPrice
	}

	if maxPrice, err := strconv.ParseFloat(c.QueryParam("max_price"), 64); err == nil && maxPrice >= 0 {
		filter.MaxPrice = &maxPrice
	}

	// Get in stock parameter
	if inStock := c.QueryParam("in_stock"); inStock != "" {
		if inStock == "true" {
			inStockValue := true
			filter.InStock = &inStockValue
		} else if inStock == "false" {
			inStockValue := false
			filter.InStock = &inStockValue
		}
	}

	// Get book list
	books, err := h.bookService.List(c.Request().Context(), filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, books)
}

// updateBook handles the request to update a book
// @Summary Update book
// @Description Updates book data
// @Tags admin,books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Param book body models.BookUpdate true "Book data"
// @Success 200 {object} models.Book
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/books/{id} [put]
func (h *BookHandler) updateBook(c echo.Context) error {
	// Get book ID from request parameters
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid book ID"})
	}

	var req models.BookUpdate
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request format"})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Update book
	book, err := h.bookService.Update(c.Request().Context(), id, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, book)
}

// deleteBook handles the request to delete a book
// @Summary Delete book
// @Description Deletes book by ID
// @Tags admin,books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/books/{id} [delete]
func (h *BookHandler) deleteBook(c echo.Context) error {
	// Get book ID from request parameters
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid book ID"})
	}

	// Delete book
	if err := h.bookService.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}
