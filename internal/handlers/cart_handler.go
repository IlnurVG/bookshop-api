package handlers

import (
	"net/http"
	"strconv"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/labstack/echo/v4"
)

// CartHandler обрабатывает запросы, связанные с корзиной покупок
type CartHandler struct {
	cartService services.CartService
}

// NewCartHandler создает новый экземпляр CartHandler
func NewCartHandler(cartService services.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

// RegisterRoutes регистрирует маршруты для обработки запросов к корзине
func (h *CartHandler) RegisterRoutes(router *echo.Group) {
	cart := router.Group("/cart")
	cart.GET("", h.getCart)
	cart.POST("/items", h.addItem)
	cart.DELETE("/items/:id", h.removeItem)
	cart.DELETE("", h.clearCart)
}

// getCart обрабатывает запрос на получение содержимого корзины
// @Summary Получить корзину
// @Description Возвращает содержимое корзины пользователя
// @Tags cart
// @Accept json
// @Produce json
// @Success 200 {object} models.CartResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /cart [get]
func (h *CartHandler) getCart(c echo.Context) error {
	// Получаем ID пользователя из контекста
	userID := getUserIDFromContext(c)

	// Получаем корзину
	cart, err := h.cartService.GetCart(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, cart)
}

// addItem обрабатывает запрос на добавление товара в корзину
// @Summary Добавить товар в корзину
// @Description Добавляет товар в корзину пользователя
// @Tags cart
// @Accept json
// @Produce json
// @Param item body models.CartItemRequest true "Данные товара"
// @Success 201 {object} models.CartResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /cart/items [post]
func (h *CartHandler) addItem(c echo.Context) error {
	// Получаем ID пользователя из контекста
	userID := getUserIDFromContext(c)

	var req models.CartItemRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "неверный формат запроса"})
	}

	// Валидация запроса
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Добавляем товар в корзину
	err := h.cartService.AddItem(c.Request().Context(), userID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Получаем обновленную корзину
	cart, err := h.cartService.GetCart(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, cart)
}

// removeItem обрабатывает запрос на удаление товара из корзины
// @Summary Удалить товар из корзины
// @Description Удаляет товар из корзины пользователя
// @Tags cart
// @Accept json
// @Produce json
// @Param id path int true "ID товара в корзине"
// @Success 200 {object} models.CartResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /cart/items/{id} [delete]
func (h *CartHandler) removeItem(c echo.Context) error {
	// Получаем ID пользователя из контекста
	userID := getUserIDFromContext(c)

	// Получаем ID товара из параметров запроса
	itemID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "некорректный ID товара"})
	}

	// Удаляем товар из корзины
	err = h.cartService.RemoveItem(c.Request().Context(), userID, itemID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Получаем обновленную корзину
	cart, err := h.cartService.GetCart(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, cart)
}

// clearCart обрабатывает запрос на очистку корзины
// @Summary Очистить корзину
// @Description Удаляет все товары из корзины пользователя
// @Tags cart
// @Accept json
// @Produce json
// @Success 204 "No Content"
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /cart [delete]
func (h *CartHandler) clearCart(c echo.Context) error {
	// Получаем ID пользователя из контекста
	userID := getUserIDFromContext(c)

	// Очищаем корзину
	if err := h.cartService.ClearCart(c.Request().Context(), userID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

// getUserIDFromContext извлекает ID пользователя из контекста запроса
func getUserIDFromContext(c echo.Context) int {
	userID, ok := c.Get("user_id").(int)
	if !ok {
		return 0
	}
	return userID
}
