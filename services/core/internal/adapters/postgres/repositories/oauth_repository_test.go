package repositories

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/packages/security"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

func setupOAuthRepositoryTest(t *testing.T) (*oauthRepository, func()) {
	if !database.IsTestingEnabled() {
		t.Skip("Skipping test: database testing not enabled")
	}

	testDB, err := database.SetupTestDB(nil) // Use default config
	require.NoError(t, err)

	encryptor := security.NewAESEncryptor("test-secret-key-32-chars-long!!!!!")
	appLogger := logger.New("test", "test")

	repo := NewOAuthRepository(testDB.DB, encryptor, appLogger).(*oauthRepository)

	cleanup := func() {
		// Clean up test data
		testDB.DB.ExecContext(context.Background(), "DELETE FROM oauth_profiles")
		testDB.DB.ExecContext(context.Background(), "DELETE FROM oauth_states")
		testDB.DB.ExecContext(context.Background(), "DELETE FROM users")
		testDB.Close()
	}

	return repo, cleanup
}

func createTestOAuthUser(t *testing.T, db *database.DB) uuid.UUID {
	userID := uuid.New()
	query := `
		INSERT INTO users (id, email, name, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := db.ExecContext(context.Background(), query,
		userID, "test@example.com", "Test User", "hash", time.Now(), time.Now())
	require.NoError(t, err)
	return userID
}

func TestOAuthRepository_CreateOAuthProfile_Success(t *testing.T) {
	repo, cleanup := setupOAuthRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()
	userID := createTestOAuthUser(t, repo.db)

	profile := &entities.OAuthProfile{
		ID:           uuid.New(),
		UserID:       userID,
		Provider:     valueobjects.OAuthProviderGoogle,
		ProviderID:   "google-123",
		Email:        "test@gmail.com",
		Name:         "Test User",
		Picture:      "https://example.com/pic.jpg",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenExpiry:  &[]time.Time{time.Now().Add(1 * time.Hour)}[0],
		Scopes:       []string{"email", "profile"},
		RawProfile:   json.RawMessage(`{"id":"google-123"}`),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.CreateOAuthProfile(ctx, profile)
	assert.NoError(t, err)

	// Verify profile was created
	var count int
	err = repo.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM oauth_profiles WHERE id = $1", profile.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestOAuthRepository_CreateOAuthProfile_DuplicateProvider(t *testing.T) {
	repo, cleanup := setupOAuthRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()
	userID := createTestOAuthUser(t, repo.db)

	profile1 := &entities.OAuthProfile{
		ID:         uuid.New(),
		UserID:     userID,
		Provider:   valueobjects.OAuthProviderGoogle,
		ProviderID: "google-123",
		Email:      "test@gmail.com",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Create first profile
	err := repo.CreateOAuthProfile(ctx, profile1)
	assert.NoError(t, err)

	// Try to create duplicate
	profile2 := &entities.OAuthProfile{
		ID:         uuid.New(),
		UserID:     uuid.New(), // Different user
		Provider:   valueobjects.OAuthProviderGoogle,
		ProviderID: "google-123", // Same provider ID
		Email:      "other@gmail.com",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = repo.CreateOAuthProfile(ctx, profile2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestOAuthRepository_GetOAuthProfile_Found(t *testing.T) {
	repo, cleanup := setupOAuthRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()
	userID := createTestOAuthUser(t, repo.db)

	// Create profile
	originalProfile := &entities.OAuthProfile{
		ID:           uuid.New(),
		UserID:       userID,
		Provider:     valueobjects.OAuthProviderGoogle,
		ProviderID:   "google-123",
		Email:        "test@gmail.com",
		Name:         "Test User",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.CreateOAuthProfile(ctx, originalProfile)
	require.NoError(t, err)

	// Get profile
	profile, err := repo.GetOAuthProfile(ctx, valueobjects.OAuthProviderGoogle, "google-123")
	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, originalProfile.ID, profile.ID)
	assert.Equal(t, originalProfile.Email, profile.Email)
	assert.Equal(t, originalProfile.AccessToken, profile.AccessToken)
}

func TestOAuthRepository_GetOAuthProfile_NotFound(t *testing.T) {
	repo, cleanup := setupOAuthRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()

	profile, err := repo.GetOAuthProfile(ctx, valueobjects.OAuthProviderGoogle, "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, profile)
	assert.Contains(t, err.Error(), "not found")
}

func TestOAuthRepository_GetUserOAuthProfiles(t *testing.T) {
	repo, cleanup := setupOAuthRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()
	userID := createTestOAuthUser(t, repo.db)

	// Create multiple profiles
	profiles := []*entities.OAuthProfile{
		{
			ID:         uuid.New(),
			UserID:     userID,
			Provider:   valueobjects.OAuthProviderGoogle,
			ProviderID: "google-123",
			Email:      "test@gmail.com",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			UserID:     userID,
			Provider:   valueobjects.OAuthProviderFacebook,
			ProviderID: "fb-456",
			Email:      "test@facebook.com",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	for _, p := range profiles {
		err := repo.CreateOAuthProfile(ctx, p)
		require.NoError(t, err)
	}

	// Get user profiles
	userProfiles, err := repo.GetUserOAuthProfiles(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, userProfiles, 2)
}

func TestOAuthRepository_UpdateOAuthTokens_Success(t *testing.T) {
	repo, cleanup := setupOAuthRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()
	userID := createTestOAuthUser(t, repo.db)

	// Create profile
	profile := &entities.OAuthProfile{
		ID:           uuid.New(),
		UserID:       userID,
		Provider:     valueobjects.OAuthProviderGoogle,
		ProviderID:   "google-123",
		AccessToken:  "old-access-token",
		RefreshToken: "old-refresh-token",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.CreateOAuthProfile(ctx, profile)
	require.NoError(t, err)

	// Update tokens
	newAccessToken := "new-access-token"
	newRefreshToken := "new-refresh-token"
	newExpiry := time.Now().Add(2 * time.Hour)

	err = repo.UpdateOAuthTokens(ctx, profile.ID, newAccessToken, newRefreshToken, &newExpiry)
	assert.NoError(t, err)

	// Verify update
	updatedProfile, err := repo.GetOAuthProfile(ctx, valueobjects.OAuthProviderGoogle, "google-123")
	assert.NoError(t, err)
	assert.Equal(t, newAccessToken, updatedProfile.AccessToken)
	assert.Equal(t, newRefreshToken, updatedProfile.RefreshToken)
	assert.NotNil(t, updatedProfile.TokenExpiry)
}

func TestOAuthRepository_DeleteOAuthProfile_Success(t *testing.T) {
	repo, cleanup := setupOAuthRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()
	userID := createTestOAuthUser(t, repo.db)

	// Create profile
	profile := &entities.OAuthProfile{
		ID:         uuid.New(),
		UserID:     userID,
		Provider:   valueobjects.OAuthProviderGoogle,
		ProviderID: "google-123",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := repo.CreateOAuthProfile(ctx, profile)
	require.NoError(t, err)

	// Delete profile
	err = repo.DeleteOAuthProfile(ctx, userID, valueobjects.OAuthProviderGoogle)
	assert.NoError(t, err)

	// Verify deletion
	var count int
	err = repo.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM oauth_profiles WHERE user_id = $1 AND provider = $2",
		userID, valueobjects.OAuthProviderGoogle.String())
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestOAuthRepository_SaveOAuthState_Success(t *testing.T) {
	repo, cleanup := setupOAuthRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()

	state := &valueobjects.OAuthState{
		State:        "test-state-123",
		CodeVerifier: "test-verifier",
		RedirectURI:  "http://localhost/callback",
		Provider:     valueobjects.OAuthProviderGoogle,
		IPAddress:    "192.168.1.1",
		UserAgent:    "Test Browser",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}

	err := repo.SaveOAuthState(ctx, state)
	assert.NoError(t, err)

	// Verify state was saved
	var count int
	err = repo.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM oauth_states WHERE state = $1", state.State)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestOAuthRepository_GetOAuthState_Valid(t *testing.T) {
	repo, cleanup := setupOAuthRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()

	originalState := &valueobjects.OAuthState{
		State:        "test-state-456",
		CodeVerifier: "test-verifier",
		RedirectURI:  "http://localhost/callback",
		Provider:     valueobjects.OAuthProviderFacebook,
		IPAddress:    "10.0.0.1",
		UserAgent:    "Test UA",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}

	err := repo.SaveOAuthState(ctx, originalState)
	require.NoError(t, err)

	// Get state
	state, err := repo.GetOAuthState(ctx, "test-state-456")
	assert.NoError(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, originalState.State, state.State)
	assert.Equal(t, originalState.CodeVerifier, state.CodeVerifier)
	assert.Equal(t, originalState.Provider, state.Provider)
}

func TestOAuthRepository_GetOAuthState_Expired(t *testing.T) {
	repo, cleanup := setupOAuthRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create expired state
	expiredState := &valueobjects.OAuthState{
		State:        "expired-state",
		CodeVerifier: "test-verifier",
		RedirectURI:  "http://localhost/callback",
		Provider:     valueobjects.OAuthProviderGoogle,
		CreatedAt:    time.Now().Add(-20 * time.Minute),
		ExpiresAt:    time.Now().Add(-10 * time.Minute), // Expired
	}

	err := repo.SaveOAuthState(ctx, expiredState)
	require.NoError(t, err)

	// Get expired state (should still return it, validation happens in service)
	state, err := repo.GetOAuthState(ctx, "expired-state")
	assert.NoError(t, err)
	assert.NotNil(t, state)
	assert.True(t, state.IsExpired())
}

func TestOAuthRepository_DeleteOAuthState_Success(t *testing.T) {
	repo, cleanup := setupOAuthRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()

	state := &valueobjects.OAuthState{
		State:        "delete-me",
		CodeVerifier: "test",
		RedirectURI:  "http://localhost",
		Provider:     valueobjects.OAuthProviderGoogle,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}

	// Save state
	err := repo.SaveOAuthState(ctx, state)
	require.NoError(t, err)

	// Delete state
	err = repo.DeleteOAuthState(ctx, "delete-me")
	assert.NoError(t, err)

	// Verify deletion
	var count int
	err = repo.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM oauth_states WHERE state = $1", "delete-me")
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestOAuthRepository_CleanupExpiredStates(t *testing.T) {
	repo, cleanup := setupOAuthRepositoryTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create mix of expired and valid states
	states := []*valueobjects.OAuthState{
		{
			State:       "expired-1",
			Provider:    valueobjects.OAuthProviderGoogle,
			CreatedAt:   time.Now().Add(-20 * time.Minute),
			ExpiresAt:   time.Now().Add(-10 * time.Minute), // Expired
			RedirectURI: "http://localhost",
		},
		{
			State:       "expired-2",
			Provider:    valueobjects.OAuthProviderFacebook,
			CreatedAt:   time.Now().Add(-15 * time.Minute),
			ExpiresAt:   time.Now().Add(-5 * time.Minute), // Expired
			RedirectURI: "http://localhost",
		},
		{
			State:       "valid-1",
			Provider:    valueobjects.OAuthProviderGoogle,
			CreatedAt:   time.Now(),
			ExpiresAt:   time.Now().Add(10 * time.Minute), // Still valid
			RedirectURI: "http://localhost",
		},
	}

	for _, s := range states {
		err := repo.SaveOAuthState(ctx, s)
		require.NoError(t, err)
	}

	// Cleanup expired states
	err := repo.CleanupExpiredStates(ctx)
	assert.NoError(t, err)

	// Verify only valid state remains
	var count int
	err = repo.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM oauth_states")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify the valid state is still there
	state, err := repo.GetOAuthState(ctx, "valid-1")
	assert.NoError(t, err)
	assert.NotNil(t, state)
}
