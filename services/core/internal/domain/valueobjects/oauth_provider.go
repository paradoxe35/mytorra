package valueobjects

import "fmt"

// OAuthProvider represents supported OAuth providers
type OAuthProvider string

const (
	OAuthProviderGoogle   OAuthProvider = "google"
	OAuthProviderFacebook OAuthProvider = "facebook"
)

// String returns the string representation of the OAuth provider
func (p OAuthProvider) String() string {
	return string(p)
}

// IsValid checks if the provider is supported
func (p OAuthProvider) IsValid() bool {
	switch p {
	case OAuthProviderGoogle, OAuthProviderFacebook:
		return true
	default:
		return false
	}
}

// ParseOAuthProvider parses a string into an OAuthProvider
func ParseOAuthProvider(provider string) (OAuthProvider, error) {
	p := OAuthProvider(provider)
	if !p.IsValid() {
		return "", fmt.Errorf("unsupported OAuth provider: %s", provider)
	}
	return p, nil
}

// GetSupportedProviders returns all supported OAuth providers
func GetSupportedProviders() []OAuthProvider {
	return []OAuthProvider{
		OAuthProviderGoogle,
		OAuthProviderFacebook,
	}
}
