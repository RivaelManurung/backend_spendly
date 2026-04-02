package auth

import "github.com/google/uuid"

type RefreshTokenManager struct {
	// In a real app, this would store tokens in Redis or DB
}

func NewRefreshTokenManager() *RefreshTokenManager {
	return &RefreshTokenManager{}
}

func (m *RefreshTokenManager) Generate(userID uuid.UUID) (string, error) {
	// Placeholder
	return "mock_refresh_token_" + userID.String(), nil
}

func (m *RefreshTokenManager) Verify(tokenStr string) (uuid.UUID, error) {
	// Placeholder
	return uuid.Nil, nil
}
