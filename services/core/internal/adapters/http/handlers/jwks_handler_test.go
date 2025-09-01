package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paradoxe35/torra/packages/cache"
	"github.com/paradoxe35/torra/packages/errors"
	"github.com/paradoxe35/torra/packages/i18n"
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

func TestJWKSHandler_GetJWKS(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockJWTKeyRepository)
	mockCache := cache.NewMockCache()
	mockLogger := logger.New("test", "test")
	translator, _ := i18n.New("en")

	handler := NewJWKSHandler(mockRepo, mockLogger, mockCache, translator)

	t.Run("Returns cached JWKS", func(t *testing.T) {
		// Setup
		cachedJWKS := security.JWKS{
			Keys: []security.JWK{
				{
					Kid: "cached-key-1",
					Kty: "RSA",
					Alg: "RS256",
					Use: "sig",
					N:   "test-n",
					E:   "AQAB",
				},
			},
		}
		cachedData, _ := json.Marshal(cachedJWKS)

		mockCache.On("Get", ctx, "jwks:public_keys").Return(cachedData, nil).Once()

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
		rec := httptest.NewRecorder()

		// Test
		handler.GetJWKS(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var response security.JWKS
		err := json.NewDecoder(rec.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Len(t, response.Keys, 1)
		assert.Equal(t, "cached-key-1", response.Keys[0].Kid)

		mockCache.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "FindAllActive")
	})

	t.Run("Fetches from DB when cache miss", func(t *testing.T) {
		// Setup - Generate a real RSA key pair for testing
		keyID, err := security.GenerateKeyID()
		assert.NoError(t, err)

		keyPair, err := security.GenerateRSAKeyPair(keyID, 30)
		assert.NoError(t, err)

		privateKeyPEM, err := security.EncodePrivateKeyPEM(keyPair.PrivateKey)
		assert.NoError(t, err)

		publicKeyPEM, err := security.EncodePublicKeyPEM(keyPair.PublicKey)
		assert.NoError(t, err)

		activeKeys := []*entities.JWTKey{
			{
				ID:         uuid.New(),
				KeyID:      keyID,
				PrivateKey: privateKeyPEM,
				PublicKey:  publicKeyPEM,
				Algorithm:  "RS256",
				IsActive:   true,
				ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
			},
		}

		mockCache.On("Get", ctx, "jwks:public_keys").Return(nil, errors.ErrNotFound).Once()
		mockRepo.On("FindAllActive", ctx).Return(activeKeys, nil).Once()
		mockCache.On("Set", ctx, "jwks:public_keys", mock.AnythingOfType("[]uint8"), 1*time.Hour).Return(nil).Once()

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
		rec := httptest.NewRecorder()

		// Test
		handler.GetJWKS(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var response security.JWKS
		err = json.NewDecoder(rec.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Len(t, response.Keys, 1)
		assert.Equal(t, keyID, response.Keys[0].Kid)
		assert.Equal(t, "RSA", response.Keys[0].Kty)
		assert.Equal(t, "RS256", response.Keys[0].Alg)
		assert.Equal(t, "sig", response.Keys[0].Use)

		mockCache.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Handles no active keys", func(t *testing.T) {
		// Setup
		mockCache.On("Get", ctx, "jwks:public_keys").Return(nil, errors.ErrNotFound).Once()
		mockRepo.On("FindAllActive", ctx).Return([]*entities.JWTKey{}, nil).Once()
		mockCache.On("Set", ctx, "jwks:public_keys", mock.AnythingOfType("[]uint8"), 1*time.Hour).Return(nil).Once()

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
		rec := httptest.NewRecorder()

		// Test
		handler.GetJWKS(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)

		var response security.JWKS
		err := json.NewDecoder(rec.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Len(t, response.Keys, 0)

		mockCache.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Handles repository error", func(t *testing.T) {
		// Setup
		mockCache.On("Get", ctx, "jwks:public_keys").Return(nil, errors.ErrNotFound).Once()
		mockRepo.On("FindAllActive", ctx).Return(nil, errors.New(errors.ErrorTypeInternal, "database error")).Once()

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
		rec := httptest.NewRecorder()

		// Test
		handler.GetJWKS(rec, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Internal server error")

		mockCache.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Handles invalid public key", func(t *testing.T) {
		// Setup
		activeKeys := []*entities.JWTKey{
			{
				ID:         uuid.New(),
				KeyID:      "invalid-key",
				PrivateKey: "invalid-private-key",
				PublicKey:  "invalid-public-key",
				Algorithm:  "RS256",
				IsActive:   true,
				ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
			},
		}

		mockCache.On("Get", ctx, "jwks:public_keys").Return(nil, errors.ErrNotFound).Once()
		mockRepo.On("FindAllActive", ctx).Return(activeKeys, nil).Once()
		// Should still try to cache even with invalid keys filtered out
		mockCache.On("Set", ctx, "jwks:public_keys", mock.AnythingOfType("[]uint8"), 1*time.Hour).Return(nil).Once()

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
		rec := httptest.NewRecorder()

		// Test
		handler.GetJWKS(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)

		var response security.JWKS
		err := json.NewDecoder(rec.Body).Decode(&response)
		assert.NoError(t, err)
		// Invalid key should be filtered out
		assert.Len(t, response.Keys, 0)

		mockCache.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}
