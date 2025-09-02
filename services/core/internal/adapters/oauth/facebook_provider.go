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
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"

	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
)

const (
	facebookGraphURL  = "https://graph.facebook.com/v18.0"
	facebookRevokeURL = "https://graph.facebook.com/v18.0/me/permissions"
	facebookTokenURL  = "https://graph.facebook.com/v18.0/oauth/access_token"
)

// FacebookOAuthProvider implements OAuth for Facebook using golang.org/x/oauth2
type FacebookOAuthProvider struct {
	config     *oauth2.Config
	appID      string
	appSecret  string
	httpClient *http.Client
	logger     logger.Interface
}

// NewFacebookOAuthProvider creates a new Facebook OAuth provider
func NewFacebookOAuthProvider(appID, appSecret, redirectURI string, logger logger.Interface) *FacebookOAuthProvider {
	return &FacebookOAuthProvider{
		config: &oauth2.Config{
			ClientID:     appID,
			ClientSecret: appSecret,
			RedirectURL:  redirectURI,
			Scopes: []string{
				"email",
				"public_profile",
			},
			Endpoint: facebook.Endpoint,
		},
		appID:      appID,
		appSecret:  appSecret,
		httpClient: http.DefaultClient,
		logger:     logger,
	}
}

// GetAuthorizationURL generates the Facebook OAuth authorization URL with PKCE
func (p *FacebookOAuthProvider) GetAuthorizationURL(state, codeVerifier, redirectURI string) string {
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

	// Build authorization URL with PKCE if supported
	var opts []oauth2.AuthCodeOption

	if codeVerifier != "" {
		// Add PKCE parameters (Facebook supports PKCE)
		codeChallenge := p.generateCodeChallenge(codeVerifier)
		opts = append(opts, oauth2.SetAuthURLParam("code_challenge", codeChallenge))
		opts = append(opts, oauth2.SetAuthURLParam("code_challenge_method", "S256"))
	}

	return config.AuthCodeURL(state, opts...)
}

// ExchangeCode exchanges the authorization code for tokens
func (p *FacebookOAuthProvider) ExchangeCode(ctx context.Context, code, codeVerifier, redirectURI string) (*ports.OAuthTokens, error) {
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

	// Facebook doesn't provide refresh tokens, but we can exchange for a long-lived token
	longLivedToken, err := p.exchangeForLongLivedToken(ctx, token.AccessToken)
	if err != nil {
		p.logger.WithError(err).Warn("Failed to exchange for long-lived token, using short-lived token")
		// Use the original token if long-lived exchange fails
		var expiresIn int64
		if !token.Expiry.IsZero() {
			expiresIn = int64(time.Until(token.Expiry).Seconds())
			if expiresIn < 0 {
				expiresIn = 0
			}
		}

		return &ports.OAuthTokens{
			AccessToken:  token.AccessToken,
			RefreshToken: "", // Facebook doesn't provide refresh tokens
			ExpiresIn:    expiresIn,
			TokenType:    token.TokenType,
			Scopes:       p.config.Scopes,
		}, nil
	}

	return &ports.OAuthTokens{
		AccessToken:  longLivedToken.AccessToken,
		RefreshToken: "", // Facebook doesn't provide refresh tokens
		ExpiresIn:    longLivedToken.ExpiresIn,
		TokenType:    longLivedToken.TokenType,
		Scopes:       p.config.Scopes,
	}, nil
}

// GetUserProfile fetches the user profile using the access token
func (p *FacebookOAuthProvider) GetUserProfile(ctx context.Context, accessToken string) (*ports.OAuthUserProfile, error) {
	params := url.Values{
		"fields":       {"id,email,name,picture.type(large)"},
		"access_token": {accessToken},
	}

	reqURL := fmt.Sprintf("%s/me?%s", facebookGraphURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user profile, status: %d, body: %s", resp.StatusCode, string(body))
	}

	var fbUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, &fbUser); err != nil {
		return nil, fmt.Errorf("failed to decode user profile: %w", err)
	}

	// Also unmarshal into raw map for preservation
	var rawProfile map[string]interface{}
	if err := json.Unmarshal(body, &rawProfile); err != nil {
		p.logger.WithError(err).Warn("Failed to unmarshal raw profile")
	}

	return &ports.OAuthUserProfile{
		ID:       fbUser.ID,
		Email:    fbUser.Email,
		Name:     fbUser.Name,
		Picture:  fbUser.Picture.Data.URL,
		Verified: true, // Facebook requires email verification for OAuth apps
		Raw:      rawProfile,
	}, nil
}

// RefreshToken - Facebook doesn't support refresh tokens
func (p *FacebookOAuthProvider) RefreshToken(ctx context.Context, refreshToken string) (*ports.OAuthTokens, error) {
	// Facebook doesn't provide refresh tokens
	// Apps should use long-lived tokens or implement token refresh through re-authentication
	return nil, fmt.Errorf("facebook does not support refresh tokens")
}

// RevokeToken revokes an OAuth token
func (p *FacebookOAuthProvider) RevokeToken(ctx context.Context, token string) error {
	params := url.Values{
		"access_token": {token},
	}

	reqURL := fmt.Sprintf("%s?%s", facebookRevokeURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "DELETE", reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

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

// exchangeForLongLivedToken exchanges a short-lived token for a long-lived token
func (p *FacebookOAuthProvider) exchangeForLongLivedToken(ctx context.Context, shortLivedToken string) (*struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}, error,
) {
	params := url.Values{
		"grant_type":        {"fb_exchange_token"},
		"client_id":         {p.appID},
		"client_secret":     {p.appSecret},
		"fb_exchange_token": {shortLivedToken},
	}

	reqURL := fmt.Sprintf("%s?%s", facebookTokenURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange for long-lived token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to exchange for long-lived token, status: %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// generateCodeChallenge generates PKCE code challenge from verifier
func (p *FacebookOAuthProvider) generateCodeChallenge(verifier string) string {
	h := sha256.New()
	h.Write([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
