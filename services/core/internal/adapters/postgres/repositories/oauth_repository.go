package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/packages/security"
	"github.com/paradoxe35/torra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

type oauthRepository struct {
	db        *database.DB
	encryptor security.Encryptor
	logger    logger.Interface
}

// NewOAuthRepository creates a new OAuth repository instance
func NewOAuthRepository(db *database.DB, encryptor security.Encryptor, logger logger.Interface) ports.OAuthRepository {
	return &oauthRepository{
		db:        db,
		encryptor: encryptor,
		logger:    logger,
	}
}

// CreateOAuthProfile creates a new OAuth profile
func (r *oauthRepository) CreateOAuthProfile(ctx context.Context, profile *entities.OAuthProfile) error {
	// Encrypt sensitive tokens
	encryptedAccessToken, err := r.encryptor.Encrypt(profile.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt access token: %w", err)
	}

	encryptedRefreshToken := ""
	if profile.RefreshToken != "" {
		encryptedRefreshToken, err = r.encryptor.Encrypt(profile.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
	}

	query := `
		INSERT INTO oauth_profiles (
			id, user_id, provider, provider_id, email, name, picture,
			access_token, refresh_token, token_expiry, scopes, raw_profile,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)`

	_, err = r.db.ExecContext(ctx, query,
		profile.ID,
		profile.UserID,
		profile.Provider.String(),
		profile.ProviderID,
		sql.NullString{String: profile.Email, Valid: profile.Email != ""},
		sql.NullString{String: profile.Name, Valid: profile.Name != ""},
		sql.NullString{String: profile.Picture, Valid: profile.Picture != ""},
		encryptedAccessToken,
		sql.NullString{String: encryptedRefreshToken, Valid: encryptedRefreshToken != ""},
		profile.TokenExpiry,
		profile.Scopes,
		profile.RawProfile,
		profile.CreatedAt,
		profile.UpdatedAt,
	)
	if err != nil {
		if database.IsUniqueViolation(err) {
			return fmt.Errorf("oauth profile already exists for this provider: %w", err)
		}
		return fmt.Errorf("failed to create oauth profile: %w", err)
	}

	return nil
}

// GetOAuthProfile retrieves an OAuth profile by provider and provider ID
func (r *oauthRepository) GetOAuthProfile(ctx context.Context, provider valueobjects.OAuthProvider, providerID string) (*entities.OAuthProfile, error) {
	var model models.OAuthProfileModel
	query := `
		SELECT id, user_id, provider, provider_id, email, name, picture,
			   access_token, refresh_token, token_expiry, scopes, raw_profile,
			   created_at, updated_at
		FROM oauth_profiles
		WHERE provider = $1 AND provider_id = $2`

	err := r.db.GetContext(ctx, &model, query, provider.String(), providerID)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, fmt.Errorf("oauth profile not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get oauth profile: %w", err)
	}

	return r.modelToEntity(&model)
}

// GetOAuthProfileByUserAndProvider retrieves an OAuth profile by user ID and provider
func (r *oauthRepository) GetOAuthProfileByUserAndProvider(ctx context.Context, userID uuid.UUID, provider valueobjects.OAuthProvider) (*entities.OAuthProfile, error) {
	var model models.OAuthProfileModel
	query := `
		SELECT id, user_id, provider, provider_id, email, name, picture,
			   access_token, refresh_token, token_expiry, scopes, raw_profile,
			   created_at, updated_at
		FROM oauth_profiles
		WHERE user_id = $1 AND provider = $2`

	err := r.db.GetContext(ctx, &model, query, userID, provider.String())
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, fmt.Errorf("oauth profile not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get oauth profile: %w", err)
	}

	return r.modelToEntity(&model)
}

// GetUserOAuthProfiles retrieves all OAuth profiles for a user
func (r *oauthRepository) GetUserOAuthProfiles(ctx context.Context, userID uuid.UUID) ([]*entities.OAuthProfile, error) {
	var models []models.OAuthProfileModel
	query := `
		SELECT id, user_id, provider, provider_id, email, name, picture,
			   access_token, refresh_token, token_expiry, scopes, raw_profile,
			   created_at, updated_at
		FROM oauth_profiles
		WHERE user_id = $1
		ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &models, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user oauth profiles: %w", err)
	}

	profiles := make([]*entities.OAuthProfile, len(models))
	for i, model := range models {
		profile, err := r.modelToEntity(&model)
		if err != nil {
			r.logger.WithError(err).Error("Failed to convert oauth profile model to entity")
			continue
		}
		profiles[i] = profile
	}

	return profiles, nil
}

// UpdateOAuthTokens updates the OAuth tokens for a profile
func (r *oauthRepository) UpdateOAuthTokens(ctx context.Context, profileID uuid.UUID, accessToken, refreshToken string, expiry *time.Time) error {
	// Encrypt tokens
	encryptedAccessToken, err := r.encryptor.Encrypt(accessToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt access token: %w", err)
	}

	encryptedRefreshToken := ""
	if refreshToken != "" {
		encryptedRefreshToken, err = r.encryptor.Encrypt(refreshToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
	}

	query := `
		UPDATE oauth_profiles
		SET access_token = $2, refresh_token = $3, token_expiry = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	_, err = r.db.ExecContext(ctx, query,
		profileID,
		encryptedAccessToken,
		sql.NullString{String: encryptedRefreshToken, Valid: encryptedRefreshToken != ""},
		expiry,
	)
	if err != nil {
		return fmt.Errorf("failed to update oauth tokens: %w", err)
	}

	return nil
}

// UpdateOAuthProfile updates an OAuth profile
func (r *oauthRepository) UpdateOAuthProfile(ctx context.Context, profile *entities.OAuthProfile) error {
	// Encrypt tokens if present
	encryptedAccessToken := profile.AccessToken
	if profile.AccessToken != "" {
		var err error
		encryptedAccessToken, err = r.encryptor.Encrypt(profile.AccessToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt access token: %w", err)
		}
	}

	encryptedRefreshToken := profile.RefreshToken
	if profile.RefreshToken != "" {
		var err error
		encryptedRefreshToken, err = r.encryptor.Encrypt(profile.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
	}

	query := `
		UPDATE oauth_profiles
		SET email = $2, name = $3, picture = $4, access_token = $5,
			refresh_token = $6, token_expiry = $7, scopes = $8,
			raw_profile = $9, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		profile.ID,
		sql.NullString{String: profile.Email, Valid: profile.Email != ""},
		sql.NullString{String: profile.Name, Valid: profile.Name != ""},
		sql.NullString{String: profile.Picture, Valid: profile.Picture != ""},
		encryptedAccessToken,
		sql.NullString{String: encryptedRefreshToken, Valid: encryptedRefreshToken != ""},
		profile.TokenExpiry,
		profile.Scopes,
		profile.RawProfile,
	)
	if err != nil {
		return fmt.Errorf("failed to update oauth profile: %w", err)
	}

	return nil
}

// DeleteOAuthProfile deletes an OAuth profile
func (r *oauthRepository) DeleteOAuthProfile(ctx context.Context, userID uuid.UUID, provider valueobjects.OAuthProvider) error {
	query := `DELETE FROM oauth_profiles WHERE user_id = $1 AND provider = $2`

	result, err := r.db.ExecContext(ctx, query, userID, provider.String())
	if err != nil {
		return fmt.Errorf("failed to delete oauth profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("oauth profile not found")
	}

	return nil
}

// SaveOAuthState saves an OAuth state for CSRF protection
func (r *oauthRepository) SaveOAuthState(ctx context.Context, state *valueobjects.OAuthState) error {
	query := `
		INSERT INTO oauth_states (
			state, code_verifier, redirect_uri, provider, ip_address,
			user_agent, user_id, created_at, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`

	var userID *uuid.UUID
	if state.UserID != nil {
		uid, err := uuid.Parse(*state.UserID)
		if err == nil {
			userID = &uid
		}
	}

	_, err := r.db.ExecContext(ctx, query,
		state.State,
		sql.NullString{String: state.CodeVerifier, Valid: state.CodeVerifier != ""},
		state.RedirectURI,
		state.Provider.String(),
		sql.NullString{String: state.IPAddress, Valid: state.IPAddress != ""},
		sql.NullString{String: state.UserAgent, Valid: state.UserAgent != ""},
		userID,
		state.CreatedAt,
		state.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save oauth state: %w", err)
	}

	return nil
}

// GetOAuthState retrieves an OAuth state
func (r *oauthRepository) GetOAuthState(ctx context.Context, state string) (*valueobjects.OAuthState, error) {
	var model models.OAuthStateModel
	query := `
		SELECT state, code_verifier, redirect_uri, provider, ip_address,
			   user_agent, user_id, created_at, expires_at
		FROM oauth_states
		WHERE state = $1`

	err := r.db.GetContext(ctx, &model, query, state)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, fmt.Errorf("oauth state not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get oauth state: %w", err)
	}

	provider, err := valueobjects.ParseOAuthProvider(model.Provider)
	if err != nil {
		return nil, fmt.Errorf("invalid oauth provider: %w", err)
	}

	var userID *string
	if model.UserID != nil {
		uid := model.UserID.String()
		userID = &uid
	}

	return &valueobjects.OAuthState{
		State:        model.State,
		CodeVerifier: model.CodeVerifier.String,
		RedirectURI:  model.RedirectURI,
		Provider:     provider,
		IPAddress:    model.IPAddress.String,
		UserAgent:    model.UserAgent.String,
		UserID:       userID,
		CreatedAt:    model.CreatedAt,
		ExpiresAt:    model.ExpiresAt,
	}, nil
}

// DeleteOAuthState deletes an OAuth state
func (r *oauthRepository) DeleteOAuthState(ctx context.Context, state string) error {
	query := `DELETE FROM oauth_states WHERE state = $1`

	_, err := r.db.ExecContext(ctx, query, state)
	if err != nil {
		return fmt.Errorf("failed to delete oauth state: %w", err)
	}

	return nil
}

// CleanupExpiredStates removes expired OAuth states
func (r *oauthRepository) CleanupExpiredStates(ctx context.Context) error {
	query := `DELETE FROM oauth_states WHERE expires_at < CURRENT_TIMESTAMP`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired oauth states: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected > 0 {
		r.logger.WithField("count", rowsAffected).Info("Cleaned up expired OAuth states")
	}

	return nil
}

// modelToEntity converts database model to domain entity
func (r *oauthRepository) modelToEntity(model *models.OAuthProfileModel) (*entities.OAuthProfile, error) {
	provider, err := valueobjects.ParseOAuthProvider(model.Provider)
	if err != nil {
		return nil, fmt.Errorf("invalid oauth provider: %w", err)
	}

	// Decrypt tokens
	decryptedAccessToken := ""
	if model.AccessToken != "" {
		decryptedAccessToken, err = r.encryptor.Decrypt(model.AccessToken)
		if err != nil {
			r.logger.WithError(err).Error("Failed to decrypt access token")
			// Don't fail, just use empty token
		}
	}

	decryptedRefreshToken := ""
	if model.RefreshToken.Valid && model.RefreshToken.String != "" {
		decryptedRefreshToken, err = r.encryptor.Decrypt(model.RefreshToken.String)
		if err != nil {
			r.logger.WithError(err).Error("Failed to decrypt refresh token")
			// Don't fail, just use empty token
		}
	}

	var tokenExpiry *time.Time
	if model.TokenExpiry.Valid {
		tokenExpiry = &model.TokenExpiry.Time
	}

	scopes := []string{}
	if model.Scopes != nil {
		scopes = model.Scopes
	}

	rawProfile := json.RawMessage(nil)
	if model.RawProfile != nil {
		rawProfile = model.RawProfile
	}

	return &entities.OAuthProfile{
		ID:           model.ID,
		UserID:       model.UserID,
		Provider:     provider,
		ProviderID:   model.ProviderID,
		Email:        model.Email.String,
		Name:         model.Name.String,
		Picture:      model.Picture.String,
		AccessToken:  decryptedAccessToken,
		RefreshToken: decryptedRefreshToken,
		TokenExpiry:  tokenExpiry,
		Scopes:       scopes,
		RawProfile:   rawProfile,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}, nil
}
