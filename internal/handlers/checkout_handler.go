package handlers

import (
	"net/http"
	"strconv"

	"github.com/bookshop/api/internal/domain/services"
	"github.com/labstack/echo/v4"
)

// CheckoutHandler обрабатывает запросы, связанные с оформлением заказов
type CheckoutHandler struct {
	checkoutService services.CheckoutService
}

// NewCheckoutHandler создает новый экземпляр CheckoutHandler
func NewCheckoutHandler(checkoutService services.CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{
		checkoutService: checkoutService,
	}
}

// RegisterRoutes регистрирует маршруты для обработки заказов
func (h *CheckoutHandler) RegisterRoutes(group *echo.Group) {
	checkout := group.Group("/checkout")

	// Маршруты для заказов (требуют аутентификации)
	checkout.POST("/orders", h.createOrder)
	checkout.GET("/orders", h.getUserOrders)
	checkout.GET("/orders/:id", h.getOrderByID)
}

// createOrder обрабатывает запрос на создание заказа из корзины пользователя
// @Summary Создать заказ
// @Description Создает новый заказ из корзины пользователя
// @Tags checkout
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 201 {object} models.OrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /checkout/orders [post]
func (h *CheckoutHandler) createOrder(c echo.Context) error {
	// Получаем ID пользователя из контекста
	userID, ok := c.Get("userID").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "пользователь не авторизован"})
	}

	// Оформляем заказ
	order, err := h.checkoutService.Checkout(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Возвращаем ответ
	return c.JSON(http.StatusCreated, order.ToResponse())
}

// getUserOrders обрабатывает запрос на получение списка заказов пользователя
// @Summary Получить заказы пользователя
// @Description Возвращает список заказов текущего пользователя
// @Tags checkout
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} models.OrderResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /checkout/orders [get]
func (h *CheckoutHandler) getUserOrders(c echo.Context) error {
	// Получаем ID пользователя из контекста
	userID, ok := c.Get("userID").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "пользователь не авторизован"})
	}

	// Получаем заказы пользователя
	orders, err := h.checkoutService.GetOrdersByUserID(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Преобразуем заказы в формат ответа
	response := make([]interface{}, len(orders))
	for i, order := range orders {
		response[i] = order.ToResponse()
	}

	// Возвращаем ответ
	return c.JSON(http.StatusOK, response)
}

// getOrderByID обрабатывает запрос на получение заказа по ID
// @Summary Получить заказ по ID
// @Description Возвращает заказ по указанному ID
// @Tags checkout
// @Accept json
// @Produce json
// @Param id path int true "ID заказа"
// @Security ApiKeyAuth
// @Success 200 {object} models.OrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /checkout/orders/{id} [get]
func (h *CheckoutHandler) getOrderByID(c echo.Context) error {
	// Получаем ID пользователя из контекста
	userID, ok := c.Get("userID").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "пользователь не авторизован"})
	}

	// Получаем ID заказа из параметров запроса
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "некорректный ID заказа"})
	}

	// Получаем заказ
	order, err := h.checkoutService.GetOrderByID(c.Request().Context(), orderID, userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	// Возвращаем ответ
	return c.JSON(http.StatusOK, order.ToResponse())
}
