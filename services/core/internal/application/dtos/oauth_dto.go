package dtos

import (
	"time"

	"github.com/google/uuid"
)

// OAuthInitRequest represents the request to initiate OAuth flow
type OAuthInitRequest struct {
	Provider    string `json:"provider" validate:"required,oneof=google facebook"`
	RedirectURI string `json:"redirect_uri" validate:"required,url"`
}

// OAuthInitResponse represents the response for OAuth initiation
type OAuthInitResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	State            string `json:"state"`
}

// OAuthCallbackRequest is defined in auth_dto.go

// OAuthCallbackResponse represents the OAuth callback response
type OAuthCallbackResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token,omitempty"`
	ExpiresIn    int64         `json:"expires_in"`
	IsNewUser    bool          `json:"is_new_user"`
}

// OAuthLinkRequest represents the request to link an OAuth account
type OAuthLinkRequest struct {
	Provider    string `json:"provider" validate:"required,oneof=google facebook"`
	RedirectURI string `json:"redirect_uri" validate:"required,url"`
}

// OAuthUnlinkRequest represents the request to unlink an OAuth account
type OAuthUnlinkRequest struct {
	Provider string `json:"provider" validate:"required,oneof=google facebook"`
}

// OAuthProfileResponse represents an OAuth profile in API responses
type OAuthProfileResponse struct {
	ID          uuid.UUID `json:"id"`
	Provider    string    `json:"provider"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Picture     string    `json:"picture,omitempty"`
	ConnectedAt time.Time `json:"connected_at"`
}

// OAuthProfilesResponse represents a list of OAuth profiles
type OAuthProfilesResponse struct {
	Profiles []*OAuthProfileResponse `json:"profiles"`
}
