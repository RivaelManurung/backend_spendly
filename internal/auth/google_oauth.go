package auth

import (
	"context"

	"google.golang.org/api/idtoken"
)

type GoogleAuthenticator struct {
	clientID string
}

func NewGoogleAuthenticator(clientID string) *GoogleAuthenticator {
	return &GoogleAuthenticator{clientID: clientID}
}

type GoogleProfile struct {
	ID        string `json:"sub"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"picture"`
}

func (a *GoogleAuthenticator) VerifyIDToken(ctx context.Context, idToken string) (*GoogleProfile, error) {
	// In production, use idtoken.Validate or similar.
	// requires GOOGLE_APPLICATION_CREDENTIALS or proper config.
	
	payload, err := idtoken.Validate(ctx, idToken, a.clientID)
	if err != nil {
		return nil, err
	}

	claims := payload.Claims
	profile := &GoogleProfile{
		ID:        claims["sub"].(string),
		Email:     claims["email"].(string),
		Name:      claims["name"].(string),
		AvatarURL: claims["picture"].(string),
	}

	return profile, nil
}
