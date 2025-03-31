package handlers

import (
	"net/http"

	"github.com/bookshop/api/internal/domain/models"
	"github.com/bookshop/api/internal/domain/services"
	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	authService services.AuthService
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRoutes registers routes for authentication handling
func (h *AuthHandler) RegisterRoutes(e *echo.Echo) {
	e.POST("/api/register", h.register)
	e.POST("/api/login", h.login)
}

// register handles user registration
// @Summary Register a new user
// @Description Creates a new user account
// @Tags authentication
// @Accept json
// @Produce json
// @Param user body models.UserRegistration true "User registration data"
// @Success 201 {object} models.User
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/register [post]
func (h *AuthHandler) register(c echo.Context) error {
	var input models.UserRegistration
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Create user
	user, err := h.authService.Register(c.Request().Context(), input)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register user")
	}

	return c.JSON(http.StatusCreated, user)
}

// login handles user login
// @Summary Log in a user
// @Description Authenticates a user and returns a token
// @Tags authentication
// @Accept json
// @Produce json
// @Param credentials body models.UserCredentials true "Login credentials"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/login [post]
func (h *AuthHandler) login(c echo.Context) error {
	var input models.UserCredentials
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Authenticate user
	accessToken, refreshToken, err := h.authService.Login(c.Request().Context(), input)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	return c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// RefreshTokenRequest represents a request to refresh a token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TokenResponse represents a response with tokens
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ErrorResponse represents a response with an error
type ErrorResponse struct {
	Error string `json:"error"`
}
