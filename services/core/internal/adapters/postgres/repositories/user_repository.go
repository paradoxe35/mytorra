package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/packages/errors"
	"github.com/paradoxe35/torra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

type userRepository struct {
	db *database.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *database.DB) ports.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	var userModel models.UserModel
	query := `
		SELECT id, email, phone, name, password_hash,
		       profile_image_url, email_verified, phone_verified,
		       id_verified, verification_level, google_id, facebook_id,
		       mfa_enabled, mfa_secret, area_id, last_known_lat, last_known_lng,
		       status, last_login_at, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &userModel, query, id)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	return userModel.ToEntity()
}

func (r *userRepository) FindByEmail(ctx context.Context, email valueobjects.Email) (*entities.User, error) {
	var userModel models.UserModel
	query := `
		SELECT id, email, phone, name, password_hash,
		       profile_image_url, email_verified, phone_verified,
		       id_verified, verification_level, google_id, facebook_id,
		       mfa_enabled, mfa_secret, area_id, last_known_lat, last_known_lng,
		       status, last_login_at, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &userModel, query, email.String())
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, nil // Return nil, nil when user not found
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return userModel.ToEntity()
}

func (r *userRepository) FindByPhone(ctx context.Context, phone valueobjects.Phone) (*entities.User, error) {
	var userModel models.UserModel
	query := `
		SELECT id, email, phone, name, password_hash,
		       profile_image_url, email_verified, phone_verified,
		       id_verified, verification_level, google_id, facebook_id,
		       mfa_enabled, mfa_secret, area_id, last_known_lat, last_known_lng,
		       status, last_login_at, created_at, updated_at, deleted_at
		FROM users
		WHERE phone = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &userModel, query, phone.String())
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by phone: %w", err)
	}

	return userModel.ToEntity()
}

func (r *userRepository) Save(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO users (
			id, email, phone, name, password_hash,
			profile_image_url, email_verified, phone_verified,
			id_verified, verification_level, google_id, facebook_id,
			mfa_enabled, mfa_secret, area_id, last_known_lat, last_known_lng,
			status, created_at, updated_at
		) VALUES (
			:id, :email, :phone, :name, :password_hash,
			:profile_image_url, :email_verified, :phone_verified,
			:id_verified, :verification_level, :google_id, :facebook_id,
			:mfa_enabled, :mfa_secret, :area_id, :last_known_lat, :last_known_lng,
			:status, :created_at, :updated_at
		)`

	// Convert domain types to database-compatible types
	dbUser := r.toDBUser(user)

	_, err := r.db.NamedExecContext(ctx, query, dbUser)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

func (r *userRepository) Update(ctx context.Context, user *entities.User) error {
	query := `
		UPDATE users SET
			email = :email,
			phone = :phone,
			name = :name,
			password_hash = :password_hash,
			profile_image_url = :profile_image_url,
			email_verified = :email_verified,
			phone_verified = :phone_verified,
			id_verified = :id_verified,
			verification_level = :verification_level,
			google_id = :google_id,
			facebook_id = :facebook_id,
			mfa_enabled = :mfa_enabled,
			mfa_secret = :mfa_secret,
			area_id = :area_id,
			last_known_lat = :last_known_lat,
			last_known_lng = :last_known_lng,
			status = :status,
			last_login_at = :last_login_at,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	// Convert domain types to database-compatible types
	dbUser := r.toDBUser(user)

	_, err := r.db.NamedExecContext(ctx, query, dbUser)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) ListByArea(ctx context.Context, area string, limit, offset int) ([]*entities.User, error) {
	// Parse area string to UUID
	areaID, err := uuid.Parse(area)
	if err != nil {
		return nil, fmt.Errorf("invalid area ID format: %w", err)
	}

	query := `
		SELECT id, email, phone, name, password_hash,
		       profile_image_url, email_verified, phone_verified,
		       id_verified, verification_level, google_id, facebook_id,
		       mfa_enabled, mfa_secret, area_id, last_known_lat, last_known_lng,
		       status, last_login_at, created_at, updated_at, deleted_at
		FROM users
		WHERE area_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	var userModels []models.UserModel
	err = r.db.SelectContext(ctx, &userModels, query, areaID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users by area: %w", err)
	}

	users := make([]*entities.User, len(userModels))
	for i, model := range userModels {
		user, err := model.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert user model: %w", err)
		}
		users[i] = user
	}

	return users, nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`

	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email valueobjects.Email) (bool, error) {
	query := `SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL`
	return r.db.ExistsContext(ctx, query, email.String())
}

func (r *userRepository) ExistsByPhone(ctx context.Context, phone valueobjects.Phone) (bool, error) {
	query := `SELECT 1 FROM users WHERE phone = $1 AND deleted_at IS NULL`
	return r.db.ExistsContext(ctx, query, phone.String())
}

// UpdateLastLogin updates the last login timestamp for a user
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, loginTime time.Time) error {
	query := `
		UPDATE users
		SET last_login_at = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, userID, loginTime)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
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

// UpdateLocation updates the user's location coordinates
func (r *userRepository) UpdateLocation(ctx context.Context, userID uuid.UUID, lat, lng float64) error {
	query := `
		UPDATE users
		SET last_known_lat = $2, last_known_lng = $3, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, userID, lat, lng)
	if err != nil {
		return fmt.Errorf("failed to update location: %w", err)
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

// Helper function to convert domain user to database-compatible struct
func (r *userRepository) toDBUser(user *entities.User) map[string]interface{} {
	return map[string]interface{}{
		"id":                 user.ID,
		"email":              user.Email.String(),
		"phone":              user.Phone.String(),
		"name":               user.Name,
		"password_hash":      user.PasswordHash,
		"profile_image_url":  user.ProfileImageURL,
		"email_verified":     user.EmailVerified,
		"phone_verified":     user.PhoneVerified,
		"id_verified":        user.IDVerified,
		"verification_level": user.VerificationLevel,
		"google_id":          user.GoogleID,
		"facebook_id":        user.FacebookID,
		"mfa_enabled":        user.MFAEnabled,
		"mfa_secret":         user.MFASecret,
		"area_id":            user.AreaID,
		"last_known_lat":     user.LastKnownLat,
		"last_known_lng":     user.LastKnownLng,
		"status":             string(user.Status),
		"last_login_at":      user.LastLoginAt,
		"created_at":         user.CreatedAt,
		"updated_at":         user.UpdatedAt,
		"deleted_at":         user.DeletedAt,
	}
}

// InTransaction runs a function within a database transaction
func (r *userRepository) InTransaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	return r.db.InTransaction(ctx, fn)
}
