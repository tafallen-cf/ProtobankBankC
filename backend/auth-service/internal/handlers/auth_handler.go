package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/protobankbankc/auth-service/internal/models"
	appErrors "github.com/protobankbankc/auth-service/pkg/errors"
)

// AuthService defines the interface for auth business logic
type AuthService interface {
	Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error)
	Login(ctx context.Context, email, password string) (*models.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*models.RefreshTokenResponse, error)
	ValidateAccessToken(ctx context.Context, accessToken string) (*models.User, error)
}

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// POST /auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	// Bind and validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	// Call service
	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"user":    user,
	})
}

// Login handles user login
// POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	// Bind and validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	// Call service
	response, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		handleError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
// POST /auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest

	// Bind and validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	// Call service
	response, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		handleError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusOK, response)
}

// GetMe returns the currently authenticated user
// GET /auth/me
func (h *AuthHandler) GetMe(c *gin.Context) {
	// Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authorization header is required",
		})
		return
	}

	// Parse Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid authorization header format",
		})
		return
	}

	accessToken := parts[1]

	// Validate token and get user
	user, err := h.authService.ValidateAccessToken(c.Request.Context(), accessToken)
	if err != nil {
		handleError(c, err)
		return
	}

	// Return user
	c.JSON(http.StatusOK, user)
}

// Logout handles user logout
// POST /auth/logout
// Note: For JWT, logout is typically handled client-side by removing the token
// This endpoint is here for completeness and can be extended with token blacklisting
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a production system, you might want to:
	// 1. Add the token to a blacklist in Redis
	// 2. Track logout events for audit
	// 3. Revoke refresh tokens

	c.JSON(http.StatusOK, gin.H{
		"message": "logout successful",
	})
}

// handleError maps service errors to HTTP responses
func handleError(c *gin.Context, err error) {
	// Check if it's an AppError
	if appErr := appErrors.GetAppError(err); appErr != nil {
		c.JSON(appErr.StatusCode, gin.H{
			"error": appErr.Message,
		})
		return
	}

	// Default to internal server error
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "an unexpected error occurred",
	})
}
