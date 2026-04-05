package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spendly/backend/internal/handler/dto"
)

type AuthHandler struct {
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// GoogleLogin handles user authentication with Google
// @Summary Google OAuth Login
// @Description Sign in or Register using Google ID Token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.LoginRequest true "Login ID Token"
// @Success 200 {object} dto.TokenResponse
// @Router /auth/google [post]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	// TODO: Verify token with Google, find/create user, generate JWT.
	// For now, return placeholder.
	resp := dto.TokenResponse{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}

	c.JSON(http.StatusOK, resp)
}

// RefreshToken handles token refresh requests
// @Summary Refresh Access Token
// @Description Obtain new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.RefreshRequest true "Refresh token"
// @Success 200 {object} dto.TokenResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Not implemented"})
}
