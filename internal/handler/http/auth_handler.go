package http

import (
	"encoding/json"
	"net/http"

	"github.com/spendly/backend/internal/handler/dto"
	"github.com/spendly/backend/pkg/apperror"
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
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, apperror.BadRequest("Invalid JSON body", err))
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

	respondWithJSON(w, http.StatusOK, resp)
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
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	c := map[string]string{"message": "Not implemented"}
	respondWithJSON(w, http.StatusNotImplemented, c)
}
