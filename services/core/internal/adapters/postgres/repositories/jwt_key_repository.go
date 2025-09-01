package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/packages/errors"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
)

type jwtKeyRepository struct {
	db *database.DB
}

// NewJWTKeyRepository creates a new JWT key repository
func NewJWTKeyRepository(db *database.DB) ports.JWTKeyRepository {
	return &jwtKeyRepository{db: db}
}

const createJWTKeyQuery = `
	INSERT INTO jwt_keys (
		id, key_id, private_key, public_key, algorithm,
		is_active, created_at, expires_at
	) VALUES (
		:id, :key_id, :private_key, :public_key, :algorithm,
		:is_active, :created_at, :expires_at
	)
`

func (r *jwtKeyRepository) Create(ctx context.Context, key *entities.JWTKey) error {
	_, err := r.db.NamedExecContext(ctx, createJWTKeyQuery, key)
	if err != nil {
		return fmt.Errorf("failed to create JWT key: %w", err)
	}
	return nil
}

const findByKeyIDQuery = `
	SELECT
		id, key_id, private_key, public_key, algorithm,
		is_active, created_at, expires_at, rotated_at
	FROM jwt_keys
	WHERE key_id = $1
`

func (r *jwtKeyRepository) FindByKeyID(ctx context.Context, keyID string) (*entities.JWTKey, error) {
	var key entities.JWTKey
	err := r.db.GetContext(ctx, &key, findByKeyIDQuery, keyID)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find JWT key by ID: %w", err)
	}
	return &key, nil
}

const findActiveKeyQuery = `
	SELECT
		id, key_id, private_key, public_key, algorithm,
		is_active, created_at, expires_at, rotated_at
	FROM jwt_keys
	WHERE is_active = true
		AND expires_at > NOW()
		AND rotated_at IS NULL
	ORDER BY created_at DESC
	LIMIT 1
`

func (r *jwtKeyRepository) FindActiveKey(ctx context.Context) (*entities.JWTKey, error) {
	var key entities.JWTKey
	err := r.db.GetContext(ctx, &key, findActiveKeyQuery)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find active JWT key: %w", err)
	}
	return &key, nil
}

const findAllActiveQuery = `
	SELECT
		id, key_id, private_key, public_key, algorithm,
		is_active, created_at, expires_at, rotated_at
	FROM jwt_keys
	WHERE is_active = true
		AND expires_at > NOW()
	ORDER BY created_at DESC
`

func (r *jwtKeyRepository) FindAllActive(ctx context.Context) ([]*entities.JWTKey, error) {
	var keys []*entities.JWTKey
	err := r.db.SelectContext(ctx, &keys, findAllActiveQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to find all active JWT keys: %w", err)
	}
	return keys, nil
}

const findAllQuery = `
	SELECT
		id, key_id, private_key, public_key, algorithm,
		is_active, created_at, expires_at, rotated_at
	FROM jwt_keys
	ORDER BY created_at DESC
`

func (r *jwtKeyRepository) FindAll(ctx context.Context) ([]*entities.JWTKey, error) {
	var keys []*entities.JWTKey
	err := r.db.SelectContext(ctx, &keys, findAllQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to find all JWT keys: %w", err)
	}
	return keys, nil
}

const updateJWTKeyQuery = `
	UPDATE jwt_keys SET
		is_active = :is_active,
		rotated_at = :rotated_at
	WHERE id = :id
`

func (r *jwtKeyRepository) Update(ctx context.Context, key *entities.JWTKey) error {
	result, err := r.db.NamedExecContext(ctx, updateJWTKeyQuery, key)
	if err != nil {
		return fmt.Errorf("failed to update JWT key: %w", err)
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

func (r *jwtKeyRepository) RotateKey(ctx context.Context, newKey *entities.JWTKey) error {
	// Start transaction
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Deactivate current active key
	now := time.Now()
	deactivateQuery := `
		UPDATE jwt_keys SET
			is_active = false,
			rotated_at = $1
		WHERE is_active = true AND rotated_at IS NULL
	`
	_, err = tx.ExecContext(ctx, deactivateQuery, now)
	if err != nil {
		return fmt.Errorf("failed to deactivate current key: %w", err)
	}

	// Insert new active key
	_, err = tx.NamedExecContext(ctx, createJWTKeyQuery, newKey)
	if err != nil {
		return fmt.Errorf("failed to create new JWT key: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

const deactivateExpiredQuery = `
	UPDATE jwt_keys SET
		is_active = false
	WHERE is_active = true
		AND expires_at <= NOW()
`

func (r *jwtKeyRepository) DeactivateExpired(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, deactivateExpiredQuery)
	if err != nil {
		return 0, fmt.Errorf("failed to deactivate expired JWT keys: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

const deleteOldKeysQuery = `
	DELETE FROM jwt_keys
	WHERE created_at < $1
`

func (r *jwtKeyRepository) DeleteOldKeys(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)
	result, err := r.db.ExecContext(ctx, deleteOldKeysQuery, cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old JWT keys: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
