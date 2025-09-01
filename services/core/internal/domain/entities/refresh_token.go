package entities

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token in the system
type RefreshToken struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  string     // Hashed version of the actual token
	DeviceID   string     // Optional: track device/client
	DeviceName string     // Optional: human-readable device name
	IPAddress  string     // IP address when token was created
	UserAgent  string     // User agent string
	ExpiresAt  time.Time  // When the token expires
	LastUsedAt *time.Time // Track last usage
	RevokedAt  *time.Time // If revoked before expiry
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// IsValid checks if the refresh token is still valid
func (rt *RefreshToken) IsValid() bool {
	now := time.Now()

	// Check if token is expired
	if now.After(rt.ExpiresAt) {
		return false
	}

	// Check if token has been revoked
	if rt.RevokedAt != nil {
		return false
	}

	return true
}

// Revoke marks the token as revoked
func (rt *RefreshToken) Revoke() {
	now := time.Now()
	rt.RevokedAt = &now
	rt.UpdatedAt = now
}

// UpdateLastUsed updates the last used timestamp
func (rt *RefreshToken) UpdateLastUsed() {
	now := time.Now()
	rt.LastUsedAt = &now
	rt.UpdatedAt = now
}

// RefreshTokenStatus represents the status of a refresh token
type RefreshTokenStatus string

const (
	RefreshTokenStatusActive  RefreshTokenStatus = "active"
	RefreshTokenStatusExpired RefreshTokenStatus = "expired"
	RefreshTokenStatusRevoked RefreshTokenStatus = "revoked"
)

// GetStatus returns the current status of the refresh token
func (rt *RefreshToken) GetStatus() RefreshTokenStatus {
	if rt.RevokedAt != nil {
		return RefreshTokenStatusRevoked
	}

	if time.Now().After(rt.ExpiresAt) {
		return RefreshTokenStatusExpired
	}

	return RefreshTokenStatusActive
}
