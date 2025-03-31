package handlers

import (
	"net/http"
	"strconv"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/labstack/echo/v4"
)

// CartHandler handles requests related to the shopping cart
type CartHandler struct {
	cartService services.CartService
}

// NewCartHandler creates a new instance of CartHandler
func NewCartHandler(cartService services.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

// RegisterRoutes registers routes for handling cart requests
func (h *CartHandler) RegisterRoutes(router *echo.Group) {
	cart := router.Group("/cart")
	cart.GET("", h.getCart)
	cart.POST("/items", h.addItem)
	cart.DELETE("/items/:id", h.removeItem)
	cart.DELETE("", h.clearCart)
}

// getCart handles the request to get cart contents
// @Summary Get cart
// @Description Returns the user's cart contents
// @Tags cart
// @Accept json
// @Produce json
// @Success 200 {object} models.CartResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /cart [get]
func (h *CartHandler) getCart(c echo.Context) error {
	// Get user ID from context
	userID := getUserIDFromContext(c)

	// Get cart
	cart, err := h.cartService.GetCart(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, cart)
}

// addItem handles the request to add an item to the cart
// @Summary Add item to cart
// @Description Adds an item to the user's cart
// @Tags cart
// @Accept json
// @Produce json
// @Param item body models.CartItemRequest true "Item data"
// @Success 201 {object} models.CartResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /cart/items [post]
func (h *CartHandler) addItem(c echo.Context) error {
	// Get user ID from context
	userID := getUserIDFromContext(c)

	var req models.CartItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request format"})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Add item to cart
	err := h.cartService.AddItem(c.Request().Context(), userID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Get updated cart
	cart, err := h.cartService.GetCart(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, cart)
}

// removeItem handles the request to remove an item from the cart
// @Summary Remove item from cart
// @Description Removes an item from the user's cart
// @Tags cart
// @Accept json
// @Produce json
// @Param id path int true "Cart item ID"
// @Success 200 {object} models.CartResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /cart/items/{id} [delete]
func (h *CartHandler) removeItem(c echo.Context) error {
	// Get user ID from context
	userID := getUserIDFromContext(c)

	// Get item ID from request parameters
	itemID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid item ID"})
	}

	// Remove item from cart
	err = h.cartService.RemoveItem(c.Request().Context(), userID, itemID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Get updated cart
	cart, err := h.cartService.GetCart(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, cart)
}

// clearCart handles the request to clear the cart
// @Summary Clear cart
// @Description Removes all items from the user's cart
// @Tags cart
// @Accept json
// @Produce json
// @Success 204 "No Content"
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /cart [delete]
func (h *CartHandler) clearCart(c echo.Context) error {
	// Get user ID from context
	userID := getUserIDFromContext(c)

	// Clear cart
	if err := h.cartService.ClearCart(c.Request().Context(), userID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

// getUserIDFromContext extracts user ID from the request context
func getUserIDFromContext(c echo.Context) int {
	userID, ok := c.Get("user_id").(int)
	if !ok {
		return 0
	}
	return userID
}
