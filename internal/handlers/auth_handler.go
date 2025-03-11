package handlers

import (
	"net/http"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/labstack/echo/v4"
)

// AuthHandler обрабатывает запросы, связанные с аутентификацией
type AuthHandler struct {
	authService services.AuthService
}

// NewAuthHandler создает новый экземпляр AuthHandler
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRoutes регистрирует маршруты для обработки аутентификации
func (h *AuthHandler) RegisterRoutes(router *echo.Group) {
	auth := router.Group("/auth")

	auth.POST("/register", h.register)
	auth.POST("/login", h.login)
	auth.POST("/refresh", h.refreshToken)
}

// register обрабатывает запрос на регистрацию нового пользователя
// @Summary Регистрация пользователя
// @Description Регистрирует нового пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.UserRegistration true "Данные пользователя"
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) register(c echo.Context) error {
	var req models.UserRegistration
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "неверный формат запроса"})
	}

	// Валидация запроса
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Регистрация пользователя
	user, err := h.authService.Register(c.Request().Context(), req)
	if err != nil {
		// Проверяем тип ошибки
		if err.Error() == "пользователь с таким email уже существует" {
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, user.ToResponse())
}

// login обрабатывает запрос на аутентификацию пользователя
// @Summary Вход пользователя
// @Description Аутентифицирует пользователя и возвращает токены
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.UserCredentials true "Учетные данные"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) login(c echo.Context) error {
	var req models.UserCredentials
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "неверный формат запроса"})
	}

	// Валидация запроса
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Аутентификация пользователя
	accessToken, refreshToken, err := h.authService.Login(c.Request().Context(), req)
	if err != nil {
		// Проверяем тип ошибки
		if err.Error() == "неверные учетные данные" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Формируем ответ
	response := map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	return c.JSON(http.StatusOK, response)
}

// refreshToken обрабатывает запрос на обновление токена
// @Summary Обновление токена
// @Description Обновляет токен доступа с помощью токена обновления
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh_token body RefreshTokenRequest true "Токен обновления"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) refreshToken(c echo.Context) error {
	var req RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "неверный формат запроса"})
	}

	// Валидация запроса
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Обновление токена
	accessToken, refreshToken, err := h.authService.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		// Проверяем тип ошибки
		if err.Error() == "недействительный токен" || err.Error() == "срок действия токена истек" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Формируем ответ
	response := map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	return c.JSON(http.StatusOK, response)
}

// RefreshTokenRequest представляет запрос на обновление токена
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TokenResponse представляет ответ с токенами
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}
