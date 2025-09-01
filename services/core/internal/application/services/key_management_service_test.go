package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paradoxe35/torra/packages/errors"
	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/packages/security"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
)

// Mock JWT Key Repository
type mockJWTKeyRepository struct {
	mock.Mock
}

func (m *mockJWTKeyRepository) Create(ctx context.Context, key *entities.JWTKey) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *mockJWTKeyRepository) FindByKeyID(ctx context.Context, keyID string) (*entities.JWTKey, error) {
	args := m.Called(ctx, keyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.JWTKey), args.Error(1)
}

func (m *mockJWTKeyRepository) FindActiveKey(ctx context.Context) (*entities.JWTKey, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.JWTKey), args.Error(1)
}

func (m *mockJWTKeyRepository) FindAllActive(ctx context.Context) ([]*entities.JWTKey, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.JWTKey), args.Error(1)
}

func (m *mockJWTKeyRepository) FindAll(ctx context.Context) ([]*entities.JWTKey, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.JWTKey), args.Error(1)
}

func (m *mockJWTKeyRepository) Update(ctx context.Context, key *entities.JWTKey) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *mockJWTKeyRepository) RotateKey(ctx context.Context, newKey *entities.JWTKey) error {
	args := m.Called(ctx, newKey)
	return args.Error(0)
}

func (m *mockJWTKeyRepository) DeactivateExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockJWTKeyRepository) DeleteOldKeys(ctx context.Context, olderThan time.Duration) (int64, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(int64), args.Error(1)
}

func TestKeyManagementService_InitializeKeys(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockJWTKeyRepository)
	mockLogger := logger.New("test", "test")

	service := NewKeyManagementService(mockRepo, mockLogger)

	t.Run("Active key exists", func(t *testing.T) {
		// Setup: Active key exists
		activeKey := &entities.JWTKey{
			ID:        uuid.New(),
			KeyID:     "test-key-id",
			IsActive:  true,
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		}

		mockRepo.On("FindActiveKey", ctx).Return(activeKey, nil).Once()

		// Test
		err := service.InitializeKeys(ctx)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("No active key - generates new", func(t *testing.T) {
		// Setup: No active key
		mockRepo.On("FindActiveKey", ctx).Return(nil, errors.ErrNotFound).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.JWTKey")).Return(nil).Once()

		// Test
		err := service.InitializeKeys(ctx)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestKeyManagementService_GenerateNewKeyPair(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockJWTKeyRepository)
	mockLogger := logger.New("test", "test")

	service := NewKeyManagementService(mockRepo, mockLogger)

	t.Run("Successfully generates and stores key pair", func(t *testing.T) {
		// Setup
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.JWTKey")).
			Return(nil).
			Run(func(args mock.Arguments) {
				key := args.Get(1).(*entities.JWTKey)

				// Verify key properties
				assert.NotEmpty(t, key.KeyID)
				assert.NotEmpty(t, key.PrivateKey)
				assert.NotEmpty(t, key.PublicKey)
				assert.Equal(t, "RS256", key.Algorithm)
				assert.True(t, key.IsActive)
				assert.True(t, key.ExpiresAt.After(time.Now()))

				// Verify keys are valid PEM
				_, err := security.DecodePrivateKeyPEM(key.PrivateKey)
				assert.NoError(t, err)

				_, err = security.DecodePublicKeyPEM(key.PublicKey)
				assert.NoError(t, err)
			}).Once()

		// Test
		err := service.GenerateNewKeyPair(ctx)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Handles repository error", func(t *testing.T) {
		// Setup
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.JWTKey")).
			Return(errors.New(errors.ErrorTypeInternal, "database error")).Once()

		// Test
		err := service.GenerateNewKeyPair(ctx)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to store JWT key")
		mockRepo.AssertExpectations(t)
	})
}

func TestKeyManagementService_RotateKeys(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockJWTKeyRepository)
	mockLogger := logger.New("test", "test")

	service := NewKeyManagementService(mockRepo, mockLogger)

	t.Run("Successfully rotates keys", func(t *testing.T) {
		// Setup
		mockRepo.On("RotateKey", ctx, mock.AnythingOfType("*entities.JWTKey")).
			Return(nil).
			Run(func(args mock.Arguments) {
				newKey := args.Get(1).(*entities.JWTKey)

				// Verify new key properties
				assert.NotEmpty(t, newKey.KeyID)
				assert.NotEmpty(t, newKey.PrivateKey)
				assert.NotEmpty(t, newKey.PublicKey)
				assert.True(t, newKey.IsActive)
			}).Once()

		// Test
		err := service.RotateKeys(ctx)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestKeyManagementService_CleanupExpiredKeys(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockJWTKeyRepository)
	mockLogger := logger.New("test", "test")

	service := NewKeyManagementService(mockRepo, mockLogger)

	t.Run("Successfully cleans up expired keys", func(t *testing.T) {
		// Setup
		mockRepo.On("DeactivateExpired", ctx).Return(int64(2), nil).Once()
		mockRepo.On("DeleteOldKeys", ctx, 90*24*time.Hour).Return(int64(5), nil).Once()

		// Test
		err := service.CleanupExpiredKeys(ctx)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Handles deactivation error", func(t *testing.T) {
		// Setup
		mockRepo.On("DeactivateExpired", ctx).Return(int64(0), errors.New(errors.ErrorTypeInternal, "database error")).Once()

		// Test
		err := service.CleanupExpiredKeys(ctx)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to deactivate expired keys")
		mockRepo.AssertExpectations(t)
	})
}

func TestKeyManagementService_GetActiveKey(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockJWTKeyRepository)
	mockLogger := logger.New("test", "test")

	service := NewKeyManagementService(mockRepo, mockLogger)

	t.Run("Returns active key", func(t *testing.T) {
		// Setup
		expectedKey := &entities.JWTKey{
			ID:        uuid.New(),
			KeyID:     "active-key",
			IsActive:  true,
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		}

		mockRepo.On("FindActiveKey", ctx).Return(expectedKey, nil).Once()

		// Test
		key, err := service.GetActiveKey(ctx)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, key)
		assert.Equal(t, expectedKey.KeyID, key.KeyID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Handles no active key", func(t *testing.T) {
		// Setup
		mockRepo.On("FindActiveKey", ctx).Return(nil, errors.ErrNotFound).Once()

		// Test
		key, err := service.GetActiveKey(ctx)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, key)
		mockRepo.AssertExpectations(t)
	})
}

func TestJWTKeyGeneration(t *testing.T) {
	// Test the actual key generation without mocks
	t.Run("Generated keys are valid RSA keys", func(t *testing.T) {
		// Generate key pair
		keyID, err := security.GenerateKeyID()
		require.NoError(t, err)

		keyPair, err := security.GenerateRSAKeyPair(keyID, 30)
		require.NoError(t, err)

		// Encode to PEM
		privateKeyPEM, err := security.EncodePrivateKeyPEM(keyPair.PrivateKey)
		require.NoError(t, err)

		publicKeyPEM, err := security.EncodePublicKeyPEM(keyPair.PublicKey)
		require.NoError(t, err)

		// Decode back and verify
		decodedPrivate, err := security.DecodePrivateKeyPEM(privateKeyPEM)
		assert.NoError(t, err)
		assert.NotNil(t, decodedPrivate)

		decodedPublic, err := security.DecodePublicKeyPEM(publicKeyPEM)
		assert.NoError(t, err)
		assert.NotNil(t, decodedPublic)

		// Verify key properties
		assert.Equal(t, keyID, keyPair.KeyID)
		assert.True(t, keyPair.ExpiresAt.After(time.Now()))
	})

	t.Run("Can convert to JWK format", func(t *testing.T) {
		// Generate key pair
		keyID, err := security.GenerateKeyID()
		require.NoError(t, err)

		keyPair, err := security.GenerateRSAKeyPair(keyID, 30)
		require.NoError(t, err)

		// Convert to JWK
		jwk := security.ConvertRSAPublicKeyToJWK(keyID, keyPair.PublicKey)

		// Verify JWK properties
		assert.Equal(t, keyID, jwk.Kid)
		assert.Equal(t, "RSA", jwk.Kty)
		assert.Equal(t, "RS256", jwk.Alg)
		assert.Equal(t, "sig", jwk.Use)
		assert.NotEmpty(t, jwk.N)
		assert.NotEmpty(t, jwk.E)
	})
}
