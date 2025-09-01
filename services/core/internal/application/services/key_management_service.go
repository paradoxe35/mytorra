package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/packages/security"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
)

// KeyManagementService handles JWT key management operations
type KeyManagementService struct {
	jwtKeyRepo ports.JWTKeyRepository
	logger     logger.Interface
}

// NewKeyManagementService creates a new key management service
func NewKeyManagementService(
	jwtKeyRepo ports.JWTKeyRepository,
	logger logger.Interface,
) *KeyManagementService {
	return &KeyManagementService{
		jwtKeyRepo: jwtKeyRepo,
		logger:     logger,
	}
}

// InitializeKeys ensures at least one active key exists
func (s *KeyManagementService) InitializeKeys(ctx context.Context) error {
	// Check if there's an active key
	activeKey, err := s.jwtKeyRepo.FindActiveKey(ctx)
	if err == nil && activeKey != nil && activeKey.IsValid() {
		s.logger.Info("Active JWT key found", "key_id", activeKey.KeyID)
		return nil
	}

	// No active key found, generate a new one
	s.logger.Info("No active JWT key found, generating new key pair")
	return s.GenerateNewKeyPair(ctx)
}

// GenerateNewKeyPair generates and stores a new RSA key pair
func (s *KeyManagementService) GenerateNewKeyPair(ctx context.Context) error {
	// Generate key ID
	keyID, err := security.GenerateKeyID()
	if err != nil {
		return fmt.Errorf("failed to generate key ID: %w", err)
	}

	// Generate RSA key pair
	keyPair, err := security.GenerateRSAKeyPair(keyID, 30) // 30 days validity
	if err != nil {
		return fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	// Encode keys to PEM format
	privateKeyPEM, err := security.EncodePrivateKeyPEM(keyPair.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to encode private key: %w", err)
	}

	publicKeyPEM, err := security.EncodePublicKeyPEM(keyPair.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to encode public key: %w", err)
	}

	// Create JWT key entity
	jwtKey := &entities.JWTKey{
		ID:         uuid.New(),
		KeyID:      keyID,
		PrivateKey: privateKeyPEM,
		PublicKey:  publicKeyPEM,
		Algorithm:  string(entities.JWTKeyAlgorithmRS256),
		IsActive:   true,
		CreatedAt:  time.Now(),
		ExpiresAt:  keyPair.ExpiresAt,
	}

	// Store in database
	if err := s.jwtKeyRepo.Create(ctx, jwtKey); err != nil {
		return fmt.Errorf("failed to store JWT key: %w", err)
	}

	s.logger.Info("New JWT key pair generated and stored", "key_id", keyID, "expires_at", keyPair.ExpiresAt)
	return nil
}

// RotateKeys rotates the active key
func (s *KeyManagementService) RotateKeys(ctx context.Context) error {
	s.logger.Info("Starting key rotation")

	// Generate new key pair
	keyID, err := security.GenerateKeyID()
	if err != nil {
		return fmt.Errorf("failed to generate key ID: %w", err)
	}

	keyPair, err := security.GenerateRSAKeyPair(keyID, 30)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	privateKeyPEM, err := security.EncodePrivateKeyPEM(keyPair.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to encode private key: %w", err)
	}

	publicKeyPEM, err := security.EncodePublicKeyPEM(keyPair.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to encode public key: %w", err)
	}

	newKey := &entities.JWTKey{
		ID:         uuid.New(),
		KeyID:      keyID,
		PrivateKey: privateKeyPEM,
		PublicKey:  publicKeyPEM,
		Algorithm:  string(entities.JWTKeyAlgorithmRS256),
		IsActive:   true,
		CreatedAt:  time.Now(),
		ExpiresAt:  keyPair.ExpiresAt,
	}

	// Rotate keys (deactivate old, activate new)
	if err := s.jwtKeyRepo.RotateKey(ctx, newKey); err != nil {
		return fmt.Errorf("failed to rotate keys: %w", err)
	}

	s.logger.Info("Key rotation completed", "new_key_id", keyID)
	return nil
}

// CleanupExpiredKeys removes expired keys
func (s *KeyManagementService) CleanupExpiredKeys(ctx context.Context) error {
	// Deactivate expired keys
	deactivated, err := s.jwtKeyRepo.DeactivateExpired(ctx)
	if err != nil {
		return fmt.Errorf("failed to deactivate expired keys: %w", err)
	}

	if deactivated > 0 {
		s.logger.Info("Deactivated expired keys", "count", deactivated)
	}

	// Delete keys older than 90 days
	deleted, err := s.jwtKeyRepo.DeleteOldKeys(ctx, 90*24*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to delete old keys: %w", err)
	}

	if deleted > 0 {
		s.logger.Info("Deleted old keys", "count", deleted)
	}

	return nil
}

// GetActiveKey returns the current active key for signing
func (s *KeyManagementService) GetActiveKey(ctx context.Context) (*entities.JWTKey, error) {
	key, err := s.jwtKeyRepo.FindActiveKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active key: %w", err)
	}
	return key, nil
}
