package handlers

import (
	"net/http"
	"strconv"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/bookshop/api/internal/middleware"
	"github.com/bookshop/api/pkg/errors"

	"github.com/labstack/echo/v4"
)

// CheckoutHandler handles requests related to order processing
type CheckoutHandler struct {
	checkoutService services.CheckoutService
}

// NewCheckoutHandler creates a new instance of CheckoutHandler
func NewCheckoutHandler(checkoutService services.CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{
		checkoutService: checkoutService,
	}
}

// RegisterRoutes registers routes for order processing
func (h *CheckoutHandler) RegisterRoutes(router *echo.Group) {
	// Order routes (require authentication)
	orders := router.Group("/orders")
	orders.Use(middleware.AuthMiddleware(middleware.NewJWTConfig("your-secret-key")))
	{
		orders.POST("", h.createOrder)
		orders.GET("", h.getUserOrders)
		orders.GET("/:id", h.getOrderByID)
	}
}

// createOrder handles the request to create an order from the user's cart
// @Summary Create order
// @Description Creates a new order from the user's cart
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} models.Order
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders [post]
func (h *CheckoutHandler) createOrder(c echo.Context) error {
	// Get user ID from context
	userID := c.Get("userID").(int)

	// Create order
	order, err := h.checkoutService.Checkout(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Return response
	return c.JSON(http.StatusCreated, order)
}

// getUserOrders handles the request to get the list of user's orders
// @Summary Get user orders
// @Description Returns a list of orders for the current user
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Order
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders [get]
func (h *CheckoutHandler) getUserOrders(c echo.Context) error {
	// Get user ID from context
	userID := c.Get("userID").(int)

	// Get user's orders
	orders, err := h.checkoutService.GetOrdersByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Convert orders to response format
	response := make([]models.Order, len(orders))
	for i, order := range orders {
		response[i] = order
	}

	// Return response
	return c.JSON(http.StatusOK, response)
}

// getOrderByID handles the request to get an order by ID
// @Summary Get order by ID
// @Description Returns an order by the specified ID
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} models.Order
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders/{id} [get]
func (h *CheckoutHandler) getOrderByID(c echo.Context) error {
	// Get user ID from context
	userID := c.Get("userID").(int)

	// Get order ID from request parameters
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid order ID"})
	}

	// Get order
	order, err := h.checkoutService.GetOrderByID(c.Request().Context(), orderID, userID)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "order not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Return response
	return c.JSON(http.StatusOK, order)
}
