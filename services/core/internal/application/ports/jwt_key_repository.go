package ports

import (
	"context"
	"time"

	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
)

// JWTKeyRepository defines the interface for JWT key persistence
type JWTKeyRepository interface {
	// Create stores a new JWT key
	Create(ctx context.Context, key *entities.JWTKey) error

	// FindByKeyID finds a key by its key ID
	FindByKeyID(ctx context.Context, keyID string) (*entities.JWTKey, error)

	// FindActiveKey returns the current active key for signing
	FindActiveKey(ctx context.Context) (*entities.JWTKey, error)

	// FindAllActive returns all active keys (for JWKS endpoint)
	FindAllActive(ctx context.Context) ([]*entities.JWTKey, error)

	// FindAll returns all keys (including inactive ones)
	FindAll(ctx context.Context) ([]*entities.JWTKey, error)

	// Update updates a JWT key
	Update(ctx context.Context, key *entities.JWTKey) error

	// RotateKey marks the current active key as rotated and creates a new active key
	RotateKey(ctx context.Context, newKey *entities.JWTKey) error

	// DeactivateExpired deactivates expired keys
	DeactivateExpired(ctx context.Context) (int64, error)

	// DeleteOldKeys deletes keys older than the specified duration
	DeleteOldKeys(ctx context.Context, olderThan time.Duration) (int64, error)
}
