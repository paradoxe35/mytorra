package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/packages/logger"
	"github.com/paradoxe35/mytorra/packages/security"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/mytorra/services/core/internal/application/ports"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
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

	// Create a copy of the profile with encrypted tokens for the model
	encryptedProfile := *profile
	encryptedProfile.AccessToken = encryptedAccessToken
	encryptedProfile.RefreshToken = encryptedRefreshToken

	model := models.OAuthProfileFromEntity(&encryptedProfile)

	_, err = r.db.NewInsert().
		Model(model).
		Exec(ctx)
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

	err := r.db.NewSelect().
		Model(&model).
		Where("provider = ? AND provider_id = ?", provider.String(), providerID).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("oauth profile not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get oauth profile: %w", err)
	}

	return r.modelToEntity(&model)
}

// GetOAuthProfileByUserAndProvider retrieves an OAuth profile by user ID and provider
func (r *oauthRepository) GetOAuthProfileByUserAndProvider(ctx context.Context, userID uuid.UUID, provider valueobjects.OAuthProvider) (*entities.OAuthProfile, error) {
	var model models.OAuthProfileModel

	err := r.db.NewSelect().
		Model(&model).
		Where("user_id = ? AND provider = ?", userID, provider.String()).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("oauth profile not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get oauth profile: %w", err)
	}

	return r.modelToEntity(&model)
}

// GetUserOAuthProfiles retrieves all OAuth profiles for a user
func (r *oauthRepository) GetUserOAuthProfiles(ctx context.Context, userID uuid.UUID) ([]*entities.OAuthProfile, error) {
	var models []models.OAuthProfileModel

	err := r.db.NewSelect().
		Model(&models).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Scan(ctx)
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

	updates := map[string]interface{}{
		"access_token": encryptedAccessToken,
		"token_expiry": expiry,
		"updated_at":   time.Now(),
	}

	if refreshToken != "" {
		encryptedRefreshToken, err := r.encryptor.Encrypt(refreshToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
		updates["refresh_token"] = encryptedRefreshToken
	} else {
		updates["refresh_token"] = ""
	}

	query := r.db.NewUpdate().
		Model(&models.OAuthProfileModel{}).
		Where("id = ?", profileID)

	for key, value := range updates {
		query = query.Set("? = ?", bun.Ident(key), value)
	}

	_, err = query.Exec(ctx)
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

	updates := map[string]interface{}{
		"email":         profile.Email,
		"name":          profile.Name,
		"picture":       profile.Picture,
		"access_token":  encryptedAccessToken,
		"refresh_token": encryptedRefreshToken,
		"token_expiry":  profile.TokenExpiry,
		"scopes":        profile.Scopes,
		"raw_profile":   profile.RawProfile,
		"updated_at":    time.Now(),
	}

	query := r.db.NewUpdate().
		Model(&models.OAuthProfileModel{}).
		Where("id = ?", profile.ID)

	for key, value := range updates {
		query = query.Set("? = ?", bun.Ident(key), value)
	}

	_, err := query.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update oauth profile: %w", err)
	}

	return nil
}

// DeleteOAuthProfile deletes an OAuth profile
func (r *oauthRepository) DeleteOAuthProfile(ctx context.Context, userID uuid.UUID, provider valueobjects.OAuthProvider) error {
	result, err := r.db.NewDelete().
		Model(&models.OAuthProfileModel{}).
		Where("user_id = ? AND provider = ?", userID, provider.String()).
		Exec(ctx)
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
	model := &models.OAuthStateModel{
		State:       state.State,
		RedirectURI: state.RedirectURI,
		Provider:    state.Provider.String(),
		CreatedAt:   state.CreatedAt,
		ExpiresAt:   state.ExpiresAt,
	}

	if state.CodeVerifier != "" {
		model.CodeVerifier = &state.CodeVerifier
	}

	if state.IPAddress != "" {
		model.IPAddress = &state.IPAddress
	}

	if state.UserAgent != "" {
		model.UserAgent = &state.UserAgent
	}

	if state.UserID != nil {
		uid, err := uuid.Parse(*state.UserID)
		if err == nil {
			model.UserID = &uid
		}
	}

	_, err := r.db.NewInsert().
		Model(model).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save oauth state: %w", err)
	}

	return nil
}

// GetOAuthState retrieves an OAuth state
func (r *oauthRepository) GetOAuthState(ctx context.Context, state string) (*valueobjects.OAuthState, error) {
	var model models.OAuthStateModel

	err := r.db.NewSelect().
		Model(&model).
		Where("state = ?", state).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
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

	var codeVerifier string
	if model.CodeVerifier != nil {
		codeVerifier = *model.CodeVerifier
	}

	var ipAddress string
	if model.IPAddress != nil {
		ipAddress = *model.IPAddress
	}

	var userAgent string
	if model.UserAgent != nil {
		userAgent = *model.UserAgent
	}

	return &valueobjects.OAuthState{
		State:        model.State,
		CodeVerifier: codeVerifier,
		RedirectURI:  model.RedirectURI,
		Provider:     provider,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		UserID:       userID,
		CreatedAt:    model.CreatedAt,
		ExpiresAt:    model.ExpiresAt,
	}, nil
}

// DeleteOAuthState deletes an OAuth state
func (r *oauthRepository) DeleteOAuthState(ctx context.Context, state string) error {
	_, err := r.db.NewDelete().
		Model(&models.OAuthStateModel{}).
		Where("state = ?", state).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete oauth state: %w", err)
	}

	return nil
}

// CleanupExpiredStates removes expired OAuth states
func (r *oauthRepository) CleanupExpiredStates(ctx context.Context) error {
	result, err := r.db.NewDelete().
		Model(&models.OAuthStateModel{}).
		Where("expires_at < ?", time.Now()).
		Exec(ctx)
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
	if model.RefreshToken != "" {
		decryptedRefreshToken, err = r.encryptor.Decrypt(model.RefreshToken)
		if err != nil {
			r.logger.WithError(err).Error("Failed to decrypt refresh token")
			// Don't fail, just use empty token
		}
	}

	var tokenExpiry *time.Time
	if model.TokenExpiry != nil {
		tokenExpiry = model.TokenExpiry
	}

	scopes := []string{}
	if model.Scopes != nil {
		scopes = []string(model.Scopes)
	}

	rawProfile := json.RawMessage(nil)
	if model.RawProfile != nil {
		rawProfile = json.RawMessage(model.RawProfile)
	}

	return &entities.OAuthProfile{
		ID:           model.ID,
		UserID:       model.UserID,
		Provider:     provider,
		ProviderID:   model.ProviderID,
		Email:        model.Email,
		Name:         model.Name,
		Picture:      model.Picture,
		AccessToken:  decryptedAccessToken,
		RefreshToken: decryptedRefreshToken,
		TokenExpiry:  tokenExpiry,
		Scopes:       scopes,
		RawProfile:   rawProfile,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}, nil
}
