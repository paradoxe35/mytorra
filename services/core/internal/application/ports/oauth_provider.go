package ports

import (
	"context"
)

// OAuthProvider defines the interface for OAuth provider implementations
type OAuthProvider interface {
	// GetAuthorizationURL generates the OAuth authorization URL
	GetAuthorizationURL(state, codeVerifier, redirectURI string) string

	// ExchangeCode exchanges the authorization code for tokens
	ExchangeCode(ctx context.Context, code, codeVerifier, redirectURI string) (*OAuthTokens, error)

	// GetUserProfile fetches the user profile using the access token
	GetUserProfile(ctx context.Context, accessToken string) (*OAuthUserProfile, error)

	// RefreshToken refreshes the access token using the refresh token
	RefreshToken(ctx context.Context, refreshToken string) (*OAuthTokens, error)

	// RevokeToken revokes an OAuth token
	RevokeToken(ctx context.Context, token string) error
}

// OAuthTokens represents OAuth tokens returned by the provider
type OAuthTokens struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	ExpiresIn    int64    `json:"expires_in,omitempty"`
	TokenType    string   `json:"token_type"`
	Scopes       []string `json:"scope,omitempty"`
}

// OAuthUserProfile represents user profile data from OAuth provider
type OAuthUserProfile struct {
	ID       string                 `json:"id"`
	Email    string                 `json:"email"`
	Name     string                 `json:"name"`
	Picture  string                 `json:"picture,omitempty"`
	Verified bool                   `json:"verified"`
	Raw      map[string]interface{} `json:"-"`
}
