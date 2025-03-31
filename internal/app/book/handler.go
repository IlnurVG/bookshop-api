package book

import (
	"net/http"
	"strconv"

	"github.com/bookshop/api/internal/domain/services"
	"github.com/labstack/echo/v4"
)

// Handler processes HTTP requests for book operations
type Handler struct {
	bookService services.BookService
}

// NewHandler creates a new instance of the book handler
func NewHandler(bookService services.BookService) *Handler {
	return &Handler{
		bookService: bookService,
	}
}

// RegisterRoutes registers routes for book request handling
func (h *Handler) RegisterRoutes(router *echo.Group) {
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

// createBook handles book creation request
// @Summary Create a book
// @Description Creates a new book
// @Tags admin,books
// @Accept json
// @Produce json
// @Param book body CreateBookRequest true "Book data"
// @Success 201 {object} BookResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/books [post]
func (h *Handler) createBook(c echo.Context) error {
	var req CreateBookRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request format"})
	}

	// Request validation
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Convert request to model
	input := req.ToModel()

	// Create book
	book, err := h.bookService.Create(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Convert model to response
	response := fromModel(book)

	return c.JSON(http.StatusCreated, response)
}

// getBook handles request to get book information
// @Summary Get a book
// @Description Returns book information by ID
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Success 200 {object} BookResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books/{id} [get]
func (h *Handler) getBook(c echo.Context) error {
	// Get book ID from request parameters
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid book ID"})
	}

	// Get the book
	book, err := h.bookService.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "book not found"})
	}

	// Convert model to response
	response := fromModel(book)

	return c.JSON(http.StatusOK, response)
}

// listBooks handles request to get a list of books
// @Summary Get list of books
// @Description Returns a list of books with filtering
// @Tags books
// @Accept json
// @Produce json
// @Param category_ids query []int false "Category IDs"
// @Param min_price query number false "Minimum price"
// @Param max_price query number false "Maximum price"
// @Param in_stock query bool false "Only in stock"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} BookListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /books [get]
func (h *Handler) listBooks(c echo.Context) error {
	var req BookListRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request format"})
	}

	// Convert request to model
	filter := req.ToModel()

	// Get list of books
	books, err := h.bookService.List(c.Request().Context(), filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Convert model to response
	response := fromModelList(books)

	return c.JSON(http.StatusOK, response)
}

// updateBook handles book update request
// @Summary Update a book
// @Description Updates book data
// @Tags admin,books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Param book body UpdateBookRequest true "Book data"
// @Success 200 {object} BookResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/books/{id} [put]
func (h *Handler) updateBook(c echo.Context) error {
	// Get book ID from request parameters
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid book ID"})
	}

	var req UpdateBookRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request format"})
	}

	// Request validation
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Convert request to model
	input := req.ToModel()

	// Update book
	book, err := h.bookService.Update(c.Request().Context(), id, input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Convert model to response
	response := fromModel(book)

	return c.JSON(http.StatusOK, response)
}

// deleteBook handles book deletion request
// @Summary Delete a book
// @Description Deletes a book by ID
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
func (h *Handler) deleteBook(c echo.Context) error {
	// Get book ID from request parameters
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid book ID"})
	}

	// Delete the book
	if err := h.bookService.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
