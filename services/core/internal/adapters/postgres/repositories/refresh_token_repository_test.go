package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/packages/errors"
	"github.com/paradoxe35/mytorra/packages/security"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

func setupRefreshTokenRepo(t *testing.T) (*refreshTokenRepository, *database.DB, func()) {
	// Setup test database with Bun
	testDB := database.NewTestDB(t)
	ctx := context.Background()

	// Create tables for required models
	err := testDB.CreateTables(ctx, (*models.RefreshTokenModel)(nil), (*models.UserModel)(nil), (*models.AreaModel)(nil))
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	repo := NewRefreshTokenRepository(testDB.DB)

	cleanup := func() {
		testDB.Close()
	}

	return repo.(*refreshTokenRepository), testDB.DB, cleanup
}

func createTestUserForRefreshToken(t *testing.T, db *database.DB) uuid.UUID {
	ctx := context.Background()

	// Create a test area first
	areaID := uuid.New()
	area := &entities.Area{
		ID:                   areaID,
		Type:                 valueobjects.AreaTypeTown,
		Name:                 "Test Area",
		Code:                 "TEST001",
		CenterLat:            0,
		CenterLng:            0,
		OperationalRadius:    10,
		Status:               valueobjects.AreaStatusActive,
		ServiceAvailability:  make(map[string]bool),
		Settings:             make(map[string]interface{}),
		SwitchingMode:        entities.SwitchingModeManual,
		AllowCrossTrade:      false,
		NearbyAreaIDs:        []uuid.UUID{},
		TransitionZones:      []entities.TransitionZone{},
		Timezone:             "UTC",
		Languages:            []string{"en"},
		Currency:             "USD",
		TaxRate:              0.0,
		MinOrderValue:        0.0,
		DeliveryFeeStructure: make(map[string]interface{}),
		CrossAreaFees:        make(map[string]float64),
		OperationalHours:     make(map[string]interface{}),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	areaModel := models.AreaFromEntity(area)
	_, err := db.NewInsert().Model(areaModel).Exec(ctx)
	require.NoError(t, err)

	// Create a test user
	userID := uuid.New()
	email, _ := valueobjects.NewEmail("test@example.com")
	phone, _ := valueobjects.NewPhone("+1234567890")

	user := &entities.User{
		ID:           userID,
		Email:        email,
		Phone:        phone,
		Name:         "Test User",
		PasswordHash: "hashed_password",
		Status:       entities.UserStatusActive,
		AreaID:       areaID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	userModel := models.UserFromEntity(user)
	_, err = db.NewInsert().Model(userModel).Exec(ctx)
	require.NoError(t, err)

	return userID
}

func TestRefreshTokenRepository_Create(t *testing.T) {
	repo, db, cleanup := setupRefreshTokenRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test user first
	userID := createTestUserForRefreshToken(t, db)

	// Generate a secure token
	rawToken, err := security.GenerateRefreshToken()
	require.NoError(t, err)

	// Hash the token for storage
	tokenHash := security.HashRefreshToken(rawToken)

	token := &entities.RefreshToken{
		ID:         uuid.New(),
		UserID:     userID, // Use the created user ID
		TokenHash:  tokenHash,
		DeviceID:   "test-device-123",
		DeviceName: "Test Device",
		IPAddress:  "192.168.1.1",
		UserAgent:  "TestAgent/1.0",
		ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Test: Create refresh token
	err = repo.Create(ctx, token)
	assert.NoError(t, err)
}

func TestRefreshTokenRepository_FindByTokenHash(t *testing.T) {
	repo, db, cleanup := setupRefreshTokenRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test user first
	userID := createTestUserForRefreshToken(t, db)

	// Create a test token first
	rawToken, err := security.GenerateRefreshToken()
	require.NoError(t, err)
	tokenHash := security.HashRefreshToken(rawToken)

	expectedToken := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID, // Use the created user ID
		TokenHash: tokenHash,
		DeviceID:  "test-device-123",
		IPAddress: "192.168.1.1", // Add valid IP address
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create the token
	err = repo.Create(ctx, expectedToken)
	require.NoError(t, err)

	// Test: Find by token hash
	foundToken, err := repo.FindByTokenHash(ctx, tokenHash)
	assert.NoError(t, err)
	assert.NotNil(t, foundToken)
	assert.Equal(t, expectedToken.ID, foundToken.ID)
	assert.Equal(t, expectedToken.UserID, foundToken.UserID)

	// Test: Find non-existent token
	nonExistentHash := security.HashRefreshToken("non-existent-token")
	_, err = repo.FindByTokenHash(ctx, nonExistentHash)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNotFound, err)
}

func TestRefreshTokenRepository_RevokeByID(t *testing.T) {
	repo, db, cleanup := setupRefreshTokenRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test user first
	userID := createTestUserForRefreshToken(t, db)

	// Create a test token
	token := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID, // Use the created user ID
		TokenHash: security.HashRefreshToken("test-token"),
		IPAddress: "192.168.1.1", // Add valid IP address
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, token)
	require.NoError(t, err)

	// Test: Revoke token
	err = repo.RevokeByID(ctx, token.ID)
	assert.NoError(t, err)

	// Test: Verify token is revoked
	revokedToken, err := repo.FindByID(ctx, token.ID)
	assert.NoError(t, err)
	assert.NotNil(t, revokedToken.RevokedAt)

	// Test: Revoke non-existent token
	err = repo.RevokeByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNotFound, err)
}

func TestRefreshTokenRepository_FindActiveByUserID(t *testing.T) {
	repo, db, cleanup := setupRefreshTokenRepo(t)
	defer cleanup()

	ctx := context.Background()
	userID := createTestUserForRefreshToken(t, db)

	// Create multiple tokens for the user
	activeToken1 := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: security.HashRefreshToken("active-token-1"),
		IPAddress: "192.168.1.1", // Add valid IP address
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	activeToken2 := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: security.HashRefreshToken("active-token-2"),
		IPAddress: "192.168.1.2", // Add valid IP address
		ExpiresAt: time.Now().Add(15 * 24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create an expired token
	expiredToken := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: security.HashRefreshToken("expired-token"),
		IPAddress: "192.168.1.3",                  // Add valid IP address
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Already expired
		CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt: time.Now(),
	}

	// Create a revoked token
	revokedTime := time.Now()
	revokedToken := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: security.HashRefreshToken("revoked-token"),
		IPAddress: "192.168.1.4", // Add valid IP address
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		RevokedAt: &revokedTime,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create all tokens
	require.NoError(t, repo.Create(ctx, activeToken1))
	require.NoError(t, repo.Create(ctx, activeToken2))
	require.NoError(t, repo.Create(ctx, expiredToken))
	require.NoError(t, repo.Create(ctx, revokedToken))

	// Test: Find active tokens only
	activeTokens, err := repo.FindActiveByUserID(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, activeTokens, 2) // Only 2 active tokens

	// Verify they are the active tokens
	for _, token := range activeTokens {
		assert.Nil(t, token.RevokedAt)
		assert.True(t, token.ExpiresAt.After(time.Now()))
	}
}

func TestRefreshTokenRepository_RevokeAllForUser(t *testing.T) {
	repo, db, cleanup := setupRefreshTokenRepo(t)
	defer cleanup()

	ctx := context.Background()
	userID := createTestUserForRefreshToken(t, db)

	// Create multiple tokens for the user
	token1 := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: security.HashRefreshToken("token-1"),
		IPAddress: "192.168.1.1", // Add valid IP address
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	token2 := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: security.HashRefreshToken("token-2"),
		IPAddress: "192.168.1.2", // Add valid IP address
		ExpiresAt: time.Now().Add(15 * 24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create tokens
	require.NoError(t, repo.Create(ctx, token1))
	require.NoError(t, repo.Create(ctx, token2))

	// Test: Revoke all tokens for user
	err := repo.RevokeAllForUser(ctx, userID)
	assert.NoError(t, err)

	// Test: Verify all tokens are revoked
	activeTokens, err := repo.FindActiveByUserID(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, activeTokens, 0)
}

func TestRefreshTokenRepository_DeleteExpired(t *testing.T) {
	repo, db, cleanup := setupRefreshTokenRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create test users for the tokens
	userID1 := createTestUserForRefreshToken(t, db)
	userID2 := createTestUserForRefreshToken(t, db)
	userID3 := createTestUserForRefreshToken(t, db)

	// Create expired tokens
	expiredToken1 := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID1, // Use created user ID
		TokenHash: security.HashRefreshToken("expired-1"),
		IPAddress: "192.168.1.1", // Add valid IP address
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		CreatedAt: time.Now().Add(-31 * 24 * time.Hour),
		UpdatedAt: time.Now(),
	}

	expiredToken2 := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID2, // Use created user ID
		TokenHash: security.HashRefreshToken("expired-2"),
		IPAddress: "192.168.1.2", // Add valid IP address
		ExpiresAt: time.Now().Add(-24 * time.Hour),
		CreatedAt: time.Now().Add(-60 * 24 * time.Hour),
		UpdatedAt: time.Now(),
	}

	// Create active token
	activeToken := &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID3, // Use created user ID
		TokenHash: security.HashRefreshToken("active"),
		IPAddress: "192.168.1.3", // Add valid IP address
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create all tokens
	require.NoError(t, repo.Create(ctx, expiredToken1))
	require.NoError(t, repo.Create(ctx, expiredToken2))
	require.NoError(t, repo.Create(ctx, activeToken))

	// Test: Delete expired tokens
	deletedCount, err := repo.DeleteExpired(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), deletedCount)

	// Test: Verify active token still exists
	foundToken, err := repo.FindByID(ctx, activeToken.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundToken)
}

func TestRefreshTokenEntity_IsValid(t *testing.T) {
	// Test valid token
	validToken := &entities.RefreshToken{
		ExpiresAt: time.Now().Add(1 * time.Hour),
		RevokedAt: nil,
	}
	assert.True(t, validToken.IsValid())

	// Test expired token
	expiredToken := &entities.RefreshToken{
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		RevokedAt: nil,
	}
	assert.False(t, expiredToken.IsValid())

	// Test revoked token
	revokedTime := time.Now()
	revokedToken := &entities.RefreshToken{
		ExpiresAt: time.Now().Add(1 * time.Hour),
		RevokedAt: &revokedTime,
	}
	assert.False(t, revokedToken.IsValid())
}

func TestRefreshTokenEntity_GetStatus(t *testing.T) {
	// Test active status
	activeToken := &entities.RefreshToken{
		ExpiresAt: time.Now().Add(1 * time.Hour),
		RevokedAt: nil,
	}
	assert.Equal(t, entities.RefreshTokenStatusActive, activeToken.GetStatus())

	// Test expired status
	expiredToken := &entities.RefreshToken{
		ExpiresAt: time.Now().Add(-1 * time.Hour),
		RevokedAt: nil,
	}
	assert.Equal(t, entities.RefreshTokenStatusExpired, expiredToken.GetStatus())

	// Test revoked status
	revokedTime := time.Now()
	revokedToken := &entities.RefreshToken{
		ExpiresAt: time.Now().Add(1 * time.Hour),
		RevokedAt: &revokedTime,
	}
	assert.Equal(t, entities.RefreshTokenStatusRevoked, revokedToken.GetStatus())
}
