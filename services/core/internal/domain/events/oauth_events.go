package events

import (
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

// UserLinkedOAuthEvent is emitted when a user links an OAuth account
type UserLinkedOAuthEvent struct {
	UserID     uuid.UUID                  `json:"user_id"`
	Provider   valueobjects.OAuthProvider `json:"provider"`
	ProviderID string                     `json:"provider_id"`
	Email      string                     `json:"email"`
	OccurredAt time.Time                  `json:"occurred_at"`
}

// NewUserLinkedOAuthEvent creates a new user linked OAuth event
func NewUserLinkedOAuthEvent(userID uuid.UUID, provider valueobjects.OAuthProvider, providerID, email string) *UserLinkedOAuthEvent {
	return &UserLinkedOAuthEvent{
		UserID:     userID,
		Provider:   provider,
		ProviderID: providerID,
		Email:      email,
		OccurredAt: time.Now(),
	}
}

// OAuthLoginEvent is emitted when a user logs in via OAuth
type OAuthLoginEvent struct {
	UserID     uuid.UUID                  `json:"user_id"`
	Provider   valueobjects.OAuthProvider `json:"provider"`
	IP         string                     `json:"ip"`
	UserAgent  string                     `json:"user_agent"`
	OccurredAt time.Time                  `json:"occurred_at"`
}

// NewOAuthLoginEvent creates a new OAuth login event
func NewOAuthLoginEvent(userID uuid.UUID, provider valueobjects.OAuthProvider, ip, userAgent string) *OAuthLoginEvent {
	return &OAuthLoginEvent{
		UserID:     userID,
		Provider:   provider,
		IP:         ip,
		UserAgent:  userAgent,
		OccurredAt: time.Now(),
	}
}

// UserUnlinkedOAuthEvent is emitted when a user unlinks an OAuth account
type UserUnlinkedOAuthEvent struct {
	UserID     uuid.UUID                  `json:"user_id"`
	Provider   valueobjects.OAuthProvider `json:"provider"`
	OccurredAt time.Time                  `json:"occurred_at"`
}

// NewUserUnlinkedOAuthEvent creates a new user unlinked OAuth event
func NewUserUnlinkedOAuthEvent(userID uuid.UUID, provider valueobjects.OAuthProvider) *UserUnlinkedOAuthEvent {
	return &UserUnlinkedOAuthEvent{
		UserID:     userID,
		Provider:   provider,
		OccurredAt: time.Now(),
	}
}
