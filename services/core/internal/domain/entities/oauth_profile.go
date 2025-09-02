package entities

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

// OAuthProfile represents an OAuth provider profile linked to a user
type OAuthProfile struct {
	ID           uuid.UUID                  `json:"id" db:"id"`
	UserID       uuid.UUID                  `json:"user_id" db:"user_id"`
	Provider     valueobjects.OAuthProvider `json:"provider" db:"provider"`
	ProviderID   string                     `json:"provider_id" db:"provider_id"`
	Email        string                     `json:"email" db:"email"`
	Name         string                     `json:"name" db:"name"`
	Picture      string                     `json:"picture" db:"picture"`
	AccessToken  string                     `json:"-" db:"access_token"`  // Hidden from JSON
	RefreshToken string                     `json:"-" db:"refresh_token"` // Hidden from JSON
	TokenExpiry  *time.Time                 `json:"token_expiry,omitempty" db:"token_expiry"`
	Scopes       []string                   `json:"scopes" db:"scopes"`
	RawProfile   json.RawMessage            `json:"-" db:"raw_profile"` // Hidden from JSON
	CreatedAt    time.Time                  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time                  `json:"updated_at" db:"updated_at"`
}

// NewOAuthProfile creates a new OAuth profile
func NewOAuthProfile(userID uuid.UUID, provider valueobjects.OAuthProvider, providerID, email, name string) *OAuthProfile {
	now := time.Now()
	return &OAuthProfile{
		ID:         uuid.New(),
		UserID:     userID,
		Provider:   provider,
		ProviderID: providerID,
		Email:      email,
		Name:       name,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// IsTokenExpired checks if the access token has expired
func (p *OAuthProfile) IsTokenExpired() bool {
	if p.TokenExpiry == nil {
		return false // No expiry means token doesn't expire
	}
	return time.Now().After(*p.TokenExpiry)
}

// UpdateTokens updates the access and refresh tokens
func (p *OAuthProfile) UpdateTokens(accessToken, refreshToken string, expiresIn int64) {
	p.AccessToken = accessToken
	p.RefreshToken = refreshToken
	if expiresIn > 0 {
		expiry := time.Now().Add(time.Duration(expiresIn) * time.Second)
		p.TokenExpiry = &expiry
	}
	p.UpdatedAt = time.Now()
}

// UpdateProfile updates the profile information
func (p *OAuthProfile) UpdateProfile(name, email, picture string, rawProfile json.RawMessage) {
	if name != "" {
		p.Name = name
	}
	if email != "" {
		p.Email = email
	}
	if picture != "" {
		p.Picture = picture
	}
	if rawProfile != nil {
		p.RawProfile = rawProfile
	}
	p.UpdatedAt = time.Now()
}
