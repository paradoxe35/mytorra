package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/packages/errors"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
)

type refreshTokenRepository struct {
	db *database.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *database.DB) ports.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

const createRefreshTokenQuery = `
	INSERT INTO refresh_tokens (
		id, user_id, token_hash, device_id, device_name,
		ip_address, user_agent, expires_at, created_at, updated_at
	) VALUES (
		:id, :user_id, :token_hash, :device_id, :device_name,
		:ip_address, :user_agent, :expires_at, :created_at, :updated_at
	)
`

func (r *refreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	_, err := r.db.NamedExecContext(ctx, createRefreshTokenQuery, token)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}
	return nil
}

const findByTokenHashQuery = `
	SELECT
		id, user_id, token_hash, device_id, device_name,
		ip_address, user_agent, expires_at, last_used_at,
		revoked_at, created_at, updated_at
	FROM refresh_tokens
	WHERE token_hash = $1
`

func (r *refreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	var token entities.RefreshToken
	err := r.db.GetContext(ctx, &token, findByTokenHashQuery, tokenHash)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find refresh token by hash: %w", err)
	}
	return &token, nil
}

const findByIDQuery = `
	SELECT
		id, user_id, token_hash, device_id, device_name,
		ip_address, user_agent, expires_at, last_used_at,
		revoked_at, created_at, updated_at
	FROM refresh_tokens
	WHERE id = $1
`

func (r *refreshTokenRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.RefreshToken, error) {
	var token entities.RefreshToken
	err := r.db.GetContext(ctx, &token, findByIDQuery, id)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find refresh token by ID: %w", err)
	}
	return &token, nil
}

const findByUserIDQuery = `
	SELECT
		id, user_id, token_hash, device_id, device_name,
		ip_address, user_agent, expires_at, last_used_at,
		revoked_at, created_at, updated_at
	FROM refresh_tokens
	WHERE user_id = $1
	ORDER BY created_at DESC
`

func (r *refreshTokenRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.RefreshToken, error) {
	var tokens []*entities.RefreshToken
	err := r.db.SelectContext(ctx, &tokens, findByUserIDQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find refresh tokens by user ID: %w", err)
	}
	return tokens, nil
}

const findActiveByUserIDQuery = `
	SELECT
		id, user_id, token_hash, device_id, device_name,
		ip_address, user_agent, expires_at, last_used_at,
		revoked_at, created_at, updated_at
	FROM refresh_tokens
	WHERE user_id = $1
		AND revoked_at IS NULL
		AND expires_at > NOW()
	ORDER BY created_at DESC
`

func (r *refreshTokenRepository) FindActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.RefreshToken, error) {
	var tokens []*entities.RefreshToken
	err := r.db.SelectContext(ctx, &tokens, findActiveByUserIDQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find active refresh tokens by user ID: %w", err)
	}
	return tokens, nil
}

const updateRefreshTokenQuery = `
	UPDATE refresh_tokens SET
		last_used_at = :last_used_at,
		revoked_at = :revoked_at,
		updated_at = :updated_at
	WHERE id = :id
`

func (r *refreshTokenRepository) Update(ctx context.Context, token *entities.RefreshToken) error {
	result, err := r.db.NamedExecContext(ctx, updateRefreshTokenQuery, token)
	if err != nil {
		return fmt.Errorf("failed to update refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

const revokeByIDQuery = `
	UPDATE refresh_tokens SET
		revoked_at = NOW(),
		updated_at = NOW()
	WHERE id = $1 AND revoked_at IS NULL
`

func (r *refreshTokenRepository) RevokeByID(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, revokeByIDQuery, id)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

const revokeAllForUserQuery = `
	UPDATE refresh_tokens SET
		revoked_at = NOW(),
		updated_at = NOW()
	WHERE user_id = $1 AND revoked_at IS NULL
`

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, revokeAllForUserQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all refresh tokens for user: %w", err)
	}
	return nil
}

const revokeByDeviceIDQuery = `
	UPDATE refresh_tokens SET
		revoked_at = NOW(),
		updated_at = NOW()
	WHERE user_id = $1 AND device_id = $2 AND revoked_at IS NULL
`

func (r *refreshTokenRepository) RevokeByDeviceID(ctx context.Context, userID uuid.UUID, deviceID string) error {
	_, err := r.db.ExecContext(ctx, revokeByDeviceIDQuery, userID, deviceID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh tokens by device ID: %w", err)
	}
	return nil
}

const deleteExpiredQuery = `
	DELETE FROM refresh_tokens
	WHERE expires_at < NOW()
`

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, deleteExpiredQuery)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

const countActiveByUserIDQuery = `
	SELECT COUNT(*)
	FROM refresh_tokens
	WHERE user_id = $1
		AND revoked_at IS NULL
		AND expires_at > NOW()
`

func (r *refreshTokenRepository) CountActiveByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.GetContext(ctx, &count, countActiveByUserIDQuery, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count active refresh tokens: %w", err)
	}
	return count, nil
}
