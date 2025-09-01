package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
)

// RefreshTokenRepository defines the interface for refresh token persistence
type RefreshTokenRepository interface {
	// Create stores a new refresh token
	Create(ctx context.Context, token *entities.RefreshToken) error

	// FindByTokenHash finds a refresh token by its hash
	FindByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error)

	// FindByID finds a refresh token by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.RefreshToken, error)

	// FindByUserID returns all refresh tokens for a user
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.RefreshToken, error)

	// FindActiveByUserID returns only active (non-revoked, non-expired) tokens for a user
	FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.RefreshToken, error)

	// Update updates a refresh token
	Update(ctx context.Context, token *entities.RefreshToken) error

	// RevokeByID revokes a specific refresh token
	RevokeByID(ctx context.Context, id uuid.UUID) error

	// RevokeAllForUser revokes all refresh tokens for a user
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error

	// RevokeByDeviceID revokes all tokens for a specific device
	RevokeByDeviceID(ctx context.Context, userID uuid.UUID, deviceID string) error

	// DeleteExpired removes expired tokens (for cleanup)
	DeleteExpired(ctx context.Context) (int64, error)

	// CountActiveByUserID counts active tokens for a user
	CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}
