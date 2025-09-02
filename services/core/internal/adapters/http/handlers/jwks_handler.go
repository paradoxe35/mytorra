package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	httpPkg "github.com/paradoxe35/mytorra/packages/http"
	"github.com/paradoxe35/mytorra/packages/i18n"
	"github.com/paradoxe35/mytorra/packages/logger"
	"github.com/paradoxe35/mytorra/packages/security"
	"github.com/paradoxe35/mytorra/services/core/internal/application/ports"
)

// JWKSHandler handles JWKS endpoint requests
type JWKSHandler struct {
	jwtKeyRepo ports.JWTKeyRepository
	logger     logger.Interface
	cache      ports.Cache
	translator *i18n.Translator
}

// NewJWKSHandler creates a new JWKS handler
func NewJWKSHandler(
	jwtKeyRepo ports.JWTKeyRepository,
	logger logger.Interface,
	cache ports.Cache,
	translator *i18n.Translator,
) *JWKSHandler {
	return &JWKSHandler{
		jwtKeyRepo: jwtKeyRepo,
		logger:     logger,
		cache:      cache,
		translator: translator,
	}
}

// GetJWKS handles GET /.well-known/jwks.json
// This endpoint is public and exposes the public keys for JWT verification
func (h *JWKSHandler) GetJWKS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Try to get from cache first
	cacheKey := "jwks:public_keys"
	if cached, err := h.cache.Get(ctx, cacheKey); err == nil && cached != nil {
		// Return cached JWKS
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
		w.Write(cached)
		return
	}

	// Get all active JWT keys from database
	activeKeys, err := h.jwtKeyRepo.FindAllActive(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch active JWT keys")
		httpPkg.RespondError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", h.translator.TWithContext(ctx, "jwks.internal_server_error"))
		return
	}

	// Convert to JWK format
	keyPairs := make([]*security.JWTKeyPair, 0, len(activeKeys))
	for _, key := range activeKeys {
		// Parse the public key from PEM format
		publicKey, err := security.DecodePublicKeyPEM(key.PublicKey)
		if err != nil {
			h.logger.WithError(err).WithField("key_id", key.KeyID).Error("Failed to decode public key")
			continue // Skip invalid keys
		}

		keyPair := &security.JWTKeyPair{
			KeyID:     key.KeyID,
			PublicKey: publicKey,
			CreatedAt: key.CreatedAt,
			ExpiresAt: key.ExpiresAt,
		}
		keyPairs = append(keyPairs, keyPair)
	}

	// Create JWKS
	jwks := security.CreateJWKS(keyPairs)

	// Marshal to JSON
	jwksJSON, err := security.MarshalJWKS(jwks)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal JWKS")
		httpPkg.RespondError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", h.translator.TWithContext(ctx, "jwks.internal_server_error"))
		return
	}

	// Cache the JWKS for 1 hour
	if err := h.cache.Set(ctx, cacheKey, jwksJSON, time.Hour); err != nil {
		h.logger.WithError(err).Error("Failed to cache JWKS")
		// Continue even if caching fails
	}

	// Return JWKS
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write(jwksJSON)
}

// GetPublicKey handles GET /.well-known/public-key
// This endpoint returns the current active public key in PEM format
func (h *JWKSHandler) GetPublicKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get the current active key
	activeKey, err := h.jwtKeyRepo.FindActiveKey(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch active JWT key")
		httpPkg.RespondError(w, http.StatusNotFound, "NO_ACTIVE_KEY_FOUND", h.translator.TWithContext(ctx, "jwks.no_active_key_found"))
		return
	}

	// Return the public key in PEM format
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	w.Header().Set("X-Key-ID", activeKey.KeyID)
	w.Write([]byte(activeKey.PublicKey))
}

// KeyInfo represents key information in the response
type KeyInfo struct {
	KeyID     string    `json:"kid"`
	Algorithm string    `json:"alg"`
	Use       string    `json:"use"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsActive  bool      `json:"is_active"`
}

// GetKeyInfo handles GET /.well-known/keys/info
// This endpoint returns information about all keys (without the actual keys)
func (h *JWKSHandler) GetKeyInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get all keys
	allKeys, err := h.jwtKeyRepo.FindAll(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to fetch JWT keys")
		httpPkg.RespondError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", h.translator.TWithContext(ctx, "jwks.internal_server_error"))
		return
	}

	// Convert to KeyInfo format
	keyInfos := make([]KeyInfo, 0, len(allKeys))
	for _, key := range allKeys {
		info := KeyInfo{
			KeyID:     key.KeyID,
			Algorithm: key.Algorithm,
			Use:       "sig",
			CreatedAt: key.CreatedAt,
			ExpiresAt: key.ExpiresAt,
			IsActive:  key.IsActive,
		}
		keyInfos = append(keyInfos, info)
	}

	// Return key information
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300") // Cache for 5 minutes
	json.NewEncoder(w).Encode(map[string]interface{}{
		"keys":       keyInfos,
		"total":      len(keyInfos),
		"active":     countActive(keyInfos),
		"expires_in": getNextExpiry(keyInfos),
	})
}

func countActive(keys []KeyInfo) int {
	count := 0
	for _, key := range keys {
		if key.IsActive {
			count++
		}
	}
	return count
}

func getNextExpiry(keys []KeyInfo) *time.Duration {
	var nextExpiry *time.Time
	now := time.Now()

	for _, key := range keys {
		if key.IsActive && key.ExpiresAt.After(now) {
			if nextExpiry == nil || key.ExpiresAt.Before(*nextExpiry) {
				nextExpiry = &key.ExpiresAt
			}
		}
	}

	if nextExpiry != nil {
		duration := nextExpiry.Sub(now)
		return &duration
	}

	return nil
}
