package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/packages/errors"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
)

func setupJWTKeyRepo(t *testing.T) (*jwtKeyRepository, func()) {
	// Skip test if database testing is not enabled
	if !database.IsTestingEnabled() {
		t.Skip("Skipping test: database testing not enabled. Set ENABLE_DB_TESTS=true or run in CI")
	}

	// Setup test database
	testDB, err := database.SetupTestDB(nil) // Use default config
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := &jwtKeyRepository{db: testDB.DB}

	cleanup := func() {
		if err := testDB.Close(); err != nil {
			t.Errorf("Failed to cleanup test database: %v", err)
		}
	}

	return repo, cleanup
}

func createTestJWTKey(t *testing.T) *entities.JWTKey {
	return &entities.JWTKey{
		ID:         uuid.New(),
		KeyID:      "test-key-id-" + uuid.New().String()[:8],
		PrivateKey: "-----BEGIN TEST PRIVATE KEY-----\nTEST_PRIVATE_KEY_DATA_FOR_TESTING_ONLY\n-----END TEST PRIVATE KEY-----",
		PublicKey:  "-----BEGIN TEST PUBLIC KEY-----\nTEST_PUBLIC_KEY_DATA_FOR_TESTING_ONLY\n-----END TEST PUBLIC KEY-----",
		Algorithm:  "RS256",
		IsActive:   true,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}
}

func TestJWTKeyRepository_Create(t *testing.T) {
	repo, cleanup := setupJWTKeyRepo(t)
	defer cleanup()

	ctx := context.Background()
	jwtKey := createTestJWTKey(t)

	// Test: Create JWT key
	err := repo.Create(ctx, jwtKey)
	assert.NoError(t, err)

	// Test: Verify JWT key was created by finding it by KeyID
	savedKey, err := repo.FindByKeyID(ctx, jwtKey.KeyID)
	assert.NoError(t, err)
	assert.NotNil(t, savedKey)
	assert.Equal(t, jwtKey.ID, savedKey.ID)
	assert.Equal(t, jwtKey.KeyID, savedKey.KeyID)
	assert.Equal(t, jwtKey.Algorithm, savedKey.Algorithm)
	assert.Equal(t, jwtKey.IsActive, savedKey.IsActive)
}

func TestJWTKeyRepository_FindActiveKey(t *testing.T) {
	repo, cleanup := setupJWTKeyRepo(t)
	defer cleanup()

	ctx := context.Background()
	jwtKey := createTestJWTKey(t)
	jwtKey.IsActive = true

	// Create JWT key first
	err := repo.Create(ctx, jwtKey)
	require.NoError(t, err)

	// Test: Find active JWT key
	foundKey, err := repo.FindActiveKey(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, foundKey)
	assert.Equal(t, jwtKey.KeyID, foundKey.KeyID)
	assert.True(t, foundKey.IsActive)

	// Test: Deactivate the key and verify no active key is found
	jwtKey.IsActive = false
	err = repo.Update(ctx, jwtKey)
	require.NoError(t, err)

	_, err = repo.FindActiveKey(ctx)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNotFound, err)
}

func TestJWTKeyRepository_FindByKeyID(t *testing.T) {
	repo, cleanup := setupJWTKeyRepo(t)
	defer cleanup()

	ctx := context.Background()
	jwtKey := createTestJWTKey(t)

	// Create JWT key first
	err := repo.Create(ctx, jwtKey)
	require.NoError(t, err)

	// Test: Find existing JWT key by KeyID
	foundKey, err := repo.FindByKeyID(ctx, jwtKey.KeyID)
	assert.NoError(t, err)
	assert.NotNil(t, foundKey)
	assert.Equal(t, jwtKey.ID, foundKey.ID)
	assert.Equal(t, jwtKey.KeyID, foundKey.KeyID)

	// Test: Find non-existent JWT key by KeyID
	_, err = repo.FindByKeyID(ctx, "non-existent-key-id")
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNotFound, err)
}

func TestJWTKeyRepository_FindAllActive(t *testing.T) {
	repo, cleanup := setupJWTKeyRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create active JWT key
	activeKey := createTestJWTKey(t)
	activeKey.IsActive = true
	activeKey.KeyID = "active-key-1"

	// Create inactive JWT key
	inactiveKey := createTestJWTKey(t)
	inactiveKey.IsActive = false
	inactiveKey.KeyID = "inactive-key-1"

	// Create expired JWT key (should be considered inactive)
	expiredKey := createTestJWTKey(t)
	expiredKey.IsActive = true
	expiredKey.ExpiresAt = time.Now().Add(-1 * time.Hour) // Already expired
	expiredKey.KeyID = "expired-key-1"

	// Save all keys
	err := repo.Create(ctx, activeKey)
	require.NoError(t, err)
	err = repo.Create(ctx, inactiveKey)
	require.NoError(t, err)
	err = repo.Create(ctx, expiredKey)
	require.NoError(t, err)

	// Test: Find only active and non-expired keys
	activeKeys, err := repo.FindAllActive(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(activeKeys), 1)

	// Verify all returned keys are active and not expired
	for _, key := range activeKeys {
		assert.True(t, key.IsActive)
		assert.True(t, key.ExpiresAt.After(time.Now()))
	}

	// Verify the active key is in the results
	found := false
	for _, key := range activeKeys {
		if key.KeyID == activeKey.KeyID {
			found = true
			break
		}
	}
	assert.True(t, found, "Active key should be found in active keys")
}

func TestJWTKeyRepository_Update(t *testing.T) {
	repo, cleanup := setupJWTKeyRepo(t)
	defer cleanup()

	ctx := context.Background()
	jwtKey := createTestJWTKey(t)
	jwtKey.IsActive = true

	// Create JWT key first
	err := repo.Create(ctx, jwtKey)
	require.NoError(t, err)

	// Test: Deactivate key using Update
	jwtKey.IsActive = false
	now := time.Now()
	jwtKey.RotatedAt = &now
	err = repo.Update(ctx, jwtKey)
	assert.NoError(t, err)

	// Test: Verify key is deactivated by checking it's not in active keys
	activeKeys, err := repo.FindAllActive(ctx)
	assert.NoError(t, err)
	for _, key := range activeKeys {
		assert.NotEqual(t, jwtKey.KeyID, key.KeyID)
	}

	// Test: Update non-existent key
	nonExistentKey := createTestJWTKey(t)
	nonExistentKey.ID = uuid.New()
	err = repo.Update(ctx, nonExistentKey)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNotFound, err)
}

func TestJWTKeyRepository_DeleteOldKeys(t *testing.T) {
	repo, cleanup := setupJWTKeyRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create old keys
	oldKey1 := createTestJWTKey(t)
	oldKey1.CreatedAt = time.Now().Add(-48 * time.Hour) // 2 days old
	oldKey1.KeyID = "old-1"

	oldKey2 := createTestJWTKey(t)
	oldKey2.CreatedAt = time.Now().Add(-25 * time.Hour) // 25 hours old
	oldKey2.KeyID = "old-2"

	// Create recent key
	recentKey := createTestJWTKey(t)
	recentKey.CreatedAt = time.Now().Add(-1 * time.Hour) // 1 hour old
	recentKey.KeyID = "recent-1"

	// Save all keys
	err := repo.Create(ctx, oldKey1)
	require.NoError(t, err)
	err = repo.Create(ctx, oldKey2)
	require.NoError(t, err)
	err = repo.Create(ctx, recentKey)
	require.NoError(t, err)

	// Test: Delete keys older than 24 hours
	deletedCount, err := repo.DeleteOldKeys(ctx, 24*time.Hour)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), deletedCount)

	// Test: Verify old keys are deleted by checking they're not in all keys
	allKeys, err := repo.FindAll(ctx)
	assert.NoError(t, err)

	// Should only have the recent key
	found := false
	for _, key := range allKeys {
		if key.KeyID == recentKey.KeyID {
			found = true
		}
		// Old keys should not be found
		assert.NotEqual(t, oldKey1.KeyID, key.KeyID)
		assert.NotEqual(t, oldKey2.KeyID, key.KeyID)
	}
	assert.True(t, found, "Recent key should still exist")
}

func TestJWTKeyRepository_RotateKey(t *testing.T) {
	repo, cleanup := setupJWTKeyRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create old active key
	oldKey := createTestJWTKey(t)
	oldKey.KeyID = "old-key"
	oldKey.IsActive = true
	oldKey.CreatedAt = time.Now().Add(-48 * time.Hour)

	// Save old key
	err := repo.Create(ctx, oldKey)
	require.NoError(t, err)

	// Create new key for rotation
	newKey := createTestJWTKey(t)
	newKey.KeyID = "new-key"
	newKey.IsActive = true

	// Test: Rotate key (this should deactivate old key and create new one)
	err = repo.RotateKey(ctx, newKey)
	assert.NoError(t, err)

	// Test: Verify old key is deactivated
	allKeys, err := repo.FindAll(ctx)
	assert.NoError(t, err)

	var oldKeyFound, newKeyFound bool
	for _, key := range allKeys {
		if key.KeyID == oldKey.KeyID {
			oldKeyFound = true
			assert.False(t, key.IsActive, "Old key should be deactivated")
			assert.NotNil(t, key.RotatedAt, "Old key should have rotated_at timestamp")
		}
		if key.KeyID == newKey.KeyID {
			newKeyFound = true
			assert.True(t, key.IsActive, "New key should be active")
		}
	}

	assert.True(t, oldKeyFound, "Old key should still exist but be deactivated")
	assert.True(t, newKeyFound, "New key should exist and be active")

	// Test: Verify only new key is in active keys
	activeKeys, err := repo.FindAllActive(ctx)
	assert.NoError(t, err)

	activeNewKeyFound := false
	for _, key := range activeKeys {
		assert.NotEqual(t, oldKey.KeyID, key.KeyID, "Old key should not be in active keys")
		if key.KeyID == newKey.KeyID {
			activeNewKeyFound = true
		}
	}
	assert.True(t, activeNewKeyFound, "New key should be in active keys")
}

func TestJWTKeyRepository_ConcurrentOperations(t *testing.T) {
	repo, cleanup := setupJWTKeyRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Test concurrent key creation
	done := make(chan bool, 2)
	createdKeys := make(chan *entities.JWTKey, 10)

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 5; i++ {
			key := createTestJWTKey(t)
			key.KeyID = fmt.Sprintf("concurrent-key-1-%d", i)
			if err := repo.Create(ctx, key); err == nil {
				createdKeys <- key
			}
		}
	}()

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 5; i++ {
			key := createTestJWTKey(t)
			key.KeyID = fmt.Sprintf("concurrent-key-2-%d", i)
			if err := repo.Create(ctx, key); err == nil {
				createdKeys <- key
			}
		}
	}()

	// Wait for both goroutines to complete
	<-done
	<-done
	close(createdKeys)

	// Verify keys were created successfully
	var keys []*entities.JWTKey
	for key := range createdKeys {
		keys = append(keys, key)
	}

	assert.GreaterOrEqual(t, len(keys), 8) // Should have created most keys successfully

	// Verify all created keys can be found
	for _, key := range keys {
		foundKey, err := repo.FindByKeyID(ctx, key.KeyID)
		assert.NoError(t, err)
		assert.Equal(t, key.KeyID, foundKey.KeyID)
	}
}
