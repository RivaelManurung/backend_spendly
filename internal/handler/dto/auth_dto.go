package dto

// LoginRequest is used for Google OAuth login.
type LoginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

// TokenResponse is returned after a successful login or refresh.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Access token expiration (seconds)
	TokenType    string `json:"token_type" default:"Bearer"`
}

// RefreshRequest is used to refresh an access token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UserResponse is a data structure for sending user profile data back to the API.
type UserResponse struct {
	ID                 string `json:"id"`
	Email              string `json:"email"`
	Name               string `json:"name"`
	AvatarURL          string `json:"avatar_url"`
	CurrencyPreference string `json:"currency_preference"`
	Status             string `json:"status"`
}
