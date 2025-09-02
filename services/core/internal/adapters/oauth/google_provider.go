package oauth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
)

const (
	googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	googleRevokeURL   = "https://oauth2.googleapis.com/revoke"
)

// GoogleOAuthProvider implements OAuth for Google using golang.org/x/oauth2
type GoogleOAuthProvider struct {
	config     *oauth2.Config
	httpClient *http.Client
	logger     logger.Interface
}

// NewGoogleOAuthProvider creates a new Google OAuth provider
func NewGoogleOAuthProvider(clientID, clientSecret, redirectURI string, logger logger.Interface) *GoogleOAuthProvider {
	return &GoogleOAuthProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURI,
			Scopes: []string{
				"openid",
				"email",
				"profile",
			},
			Endpoint: google.Endpoint,
		},
		httpClient: http.DefaultClient,
		logger:     logger,
	}
}

// GetAuthorizationURL generates the Google OAuth authorization URL with PKCE
func (p *GoogleOAuthProvider) GetAuthorizationURL(state, codeVerifier, redirectURI string) string {
	// Update redirect URI if different
	config := p.config
	if redirectURI != "" && redirectURI != p.config.RedirectURL {
		config = &oauth2.Config{
			ClientID:     p.config.ClientID,
			ClientSecret: p.config.ClientSecret,
			RedirectURL:  redirectURI,
			Scopes:       p.config.Scopes,
			Endpoint:     p.config.Endpoint,
		}
	}

	// Build authorization URL with PKCE
	var opts []oauth2.AuthCodeOption
	opts = append(opts, oauth2.AccessTypeOffline)                    // To get refresh token
	opts = append(opts, oauth2.SetAuthURLParam("prompt", "consent")) // Force consent to get refresh token

	if codeVerifier != "" {
		// Add PKCE parameters
		codeChallenge := p.generateCodeChallenge(codeVerifier)
		opts = append(opts, oauth2.SetAuthURLParam("code_challenge", codeChallenge))
		opts = append(opts, oauth2.SetAuthURLParam("code_challenge_method", "S256"))
	}

	return config.AuthCodeURL(state, opts...)
}

// ExchangeCode exchanges the authorization code for tokens
func (p *GoogleOAuthProvider) ExchangeCode(ctx context.Context, code, codeVerifier, redirectURI string) (*ports.OAuthTokens, error) {
	// Update redirect URI if different
	config := p.config
	if redirectURI != "" && redirectURI != p.config.RedirectURL {
		config = &oauth2.Config{
			ClientID:     p.config.ClientID,
			ClientSecret: p.config.ClientSecret,
			RedirectURL:  redirectURI,
			Scopes:       p.config.Scopes,
			Endpoint:     p.config.Endpoint,
		}
	}

	// Add PKCE verifier if provided
	var opts []oauth2.AuthCodeOption
	if codeVerifier != "" {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	}

	token, err := config.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Extract scopes from token extra fields if available
	var scopes []string
	if scopeInterface := token.Extra("scope"); scopeInterface != nil {
		if scopeStr, ok := scopeInterface.(string); ok {
			scopes = strings.Split(scopeStr, " ")
		}
	}

	// Calculate expires in seconds
	var expiresIn int64
	if !token.Expiry.IsZero() {
		expiresIn = int64(time.Until(token.Expiry).Seconds())
		if expiresIn < 0 {
			expiresIn = 0
		}
	}

	return &ports.OAuthTokens{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    token.TokenType,
		Scopes:       scopes,
	}, nil
}

// GetUserProfile fetches the user profile using the access token
func (p *GoogleOAuthProvider) GetUserProfile(ctx context.Context, accessToken string) (*ports.OAuthUserProfile, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", googleUserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user profile, status: %d, body: %s", resp.StatusCode, string(body))
	}

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Locale        string `json:"locale"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, fmt.Errorf("failed to decode user profile: %w", err)
	}

	// Also unmarshal into raw map for preservation
	var rawProfile map[string]interface{}
	if err := json.Unmarshal(body, &rawProfile); err != nil {
		p.logger.WithError(err).Warn("Failed to unmarshal raw profile")
	}

	return &ports.OAuthUserProfile{
		ID:       googleUser.ID,
		Email:    googleUser.Email,
		Name:     googleUser.Name,
		Picture:  googleUser.Picture,
		Verified: googleUser.VerifiedEmail,
		Raw:      rawProfile,
	}, nil
}

// RefreshToken refreshes the access token using the refresh token
func (p *GoogleOAuthProvider) RefreshToken(ctx context.Context, refreshToken string) (*ports.OAuthTokens, error) {
	// Create a token source with the refresh token
	tokenSource := p.config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})

	// Get a new token (this will automatically refresh it)
	token, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Extract scopes from token extra fields if available
	var scopes []string
	if scopeInterface := token.Extra("scope"); scopeInterface != nil {
		if scopeStr, ok := scopeInterface.(string); ok {
			scopes = strings.Split(scopeStr, " ")
		}
	}

	// Calculate expires in seconds
	var expiresIn int64
	if !token.Expiry.IsZero() {
		expiresIn = int64(time.Until(token.Expiry).Seconds())
		if expiresIn < 0 {
			expiresIn = 0
		}
	}

	return &ports.OAuthTokens{
		AccessToken:  token.AccessToken,
		RefreshToken: refreshToken, // Google doesn't return a new refresh token
		ExpiresIn:    expiresIn,
		TokenType:    token.TokenType,
		Scopes:       scopes,
	}, nil
}

// RevokeToken revokes an OAuth token
func (p *GoogleOAuthProvider) RevokeToken(ctx context.Context, token string) error {
	data := url.Values{
		"token": {token},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", googleRevokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to revoke token, status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// generateCodeChallenge generates PKCE code challenge from verifier
func (p *GoogleOAuthProvider) generateCodeChallenge(verifier string) string {
	h := sha256.New()
	h.Write([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
