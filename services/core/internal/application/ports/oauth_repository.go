package ports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

// OAuthRepository defines the interface for OAuth-related database operations
type OAuthRepository interface {
	// OAuth Profile Management
	CreateOAuthProfile(ctx context.Context, profile *entities.OAuthProfile) error
	GetOAuthProfile(ctx context.Context, provider valueobjects.OAuthProvider, providerID string) (*entities.OAuthProfile, error)
	GetOAuthProfileByUserAndProvider(ctx context.Context, userID uuid.UUID, provider valueobjects.OAuthProvider) (*entities.OAuthProfile, error)
	GetUserOAuthProfiles(ctx context.Context, userID uuid.UUID) ([]*entities.OAuthProfile, error)
	UpdateOAuthTokens(ctx context.Context, profileID uuid.UUID, accessToken, refreshToken string, expiry *time.Time) error
	UpdateOAuthProfile(ctx context.Context, profile *entities.OAuthProfile) error
	DeleteOAuthProfile(ctx context.Context, userID uuid.UUID, provider valueobjects.OAuthProvider) error

	// OAuth State Management (CSRF Protection)
	SaveOAuthState(ctx context.Context, state *valueobjects.OAuthState) error
	GetOAuthState(ctx context.Context, state string) (*valueobjects.OAuthState, error)
	DeleteOAuthState(ctx context.Context, state string) error
	CleanupExpiredStates(ctx context.Context) error
}
