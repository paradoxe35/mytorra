package valueobjects

import (
	"time"
)

// OAuthState represents the OAuth state for CSRF protection and PKCE
type OAuthState struct {
	State        string        `json:"state"`
	CodeVerifier string        `json:"code_verifier,omitempty"` // For PKCE
	RedirectURI  string        `json:"redirect_uri"`
	Provider     OAuthProvider `json:"provider"`
	IPAddress    string        `json:"ip_address,omitempty"`
	UserAgent    string        `json:"user_agent,omitempty"`
	UserID       *string       `json:"user_id,omitempty"` // For linking existing accounts
	ExpiresAt    time.Time     `json:"expires_at"`
	CreatedAt    time.Time     `json:"created_at"`
}

// IsExpired checks if the OAuth state has expired
func (s *OAuthState) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if the OAuth state is valid
func (s *OAuthState) IsValid() bool {
	if s.State == "" || s.Provider == "" || s.RedirectURI == "" {
		return false
	}
	if !s.Provider.IsValid() {
		return false
	}
	return !s.IsExpired()
}
