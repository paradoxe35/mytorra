package middleware

import (
	"context"
	"crypto/rsa"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	httpPkg "github.com/paradoxe35/torra/packages/http"
	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/packages/security"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
)

type AuthMiddleware struct {
	jwtKeyRepo ports.JWTKeyRepository
	logger     logger.Interface
	publicKeys map[string]*rsa.PublicKey // Cache of public keys
}

func NewAuthMiddleware(jwtKeyRepo ports.JWTKeyRepository, logger logger.Interface) *AuthMiddleware {
	return &AuthMiddleware{
		jwtKeyRepo: jwtKeyRepo,
		logger:     logger,
		publicKeys: make(map[string]*rsa.PublicKey),
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httpPkg.RespondError(w, http.StatusUnauthorized, "MISSING_TOKEN", "Authorization header is required")
			return
		}

		// Check for Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			httpPkg.RespondError(w, http.StatusUnauthorized, "INVALID_TOKEN_FORMAT", "Invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Parse token without verification first to get the key ID
		unverifiedToken, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			httpPkg.RespondError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid token format")
			return
		}

		// Get key ID from token header
		keyID, ok := unverifiedToken.Header["kid"].(string)
		if !ok || keyID == "" {
			httpPkg.RespondError(w, http.StatusUnauthorized, "MISSING_KEY_ID", "Token missing key ID")
			return
		}

		// Get public key for verification
		publicKey, err := m.getPublicKey(r.Context(), keyID)
		if err != nil {
			m.logger.WithError(err).Error("Failed to get public key for token verification")
			httpPkg.RespondError(w, http.StatusUnauthorized, "INVALID_KEY", "Invalid signing key")
			return
		}

		// Parse and validate token with RS256
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return publicKey, nil
		})

		if err != nil || !token.Valid {
			m.logger.WithError(err).Debug("Token validation failed")
			httpPkg.RespondError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token")
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			httpPkg.RespondError(w, http.StatusUnauthorized, "INVALID_CLAIMS", "Invalid token claims")
			return
		}

		// Add user context
		ctx := context.WithValue(r.Context(), "user_id", claims["sub"])
		ctx = context.WithValue(ctx, "user_email", claims["email"])
		ctx = context.WithValue(ctx, "user_name", claims["name"])
		ctx = context.WithValue(ctx, "area_id", claims["area_id"])

		// Continue with authenticated request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getPublicKey retrieves the public key for a given key ID
func (m *AuthMiddleware) getPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	// Check cache first
	if publicKey, exists := m.publicKeys[keyID]; exists {
		return publicKey, nil
	}

	// Fetch from database
	jwtKey, err := m.jwtKeyRepo.FindByKeyID(ctx, keyID)
	if err != nil {
		return nil, err
	}

	// Parse public key from PEM
	publicKey, err := security.DecodePublicKeyPEM(jwtKey.PublicKey)
	if err != nil {
		return nil, err
	}

	// Cache the public key
	m.publicKeys[keyID] = publicKey

	return publicKey, nil
}

// RefreshPublicKeyCache refreshes the cache of public keys
func (m *AuthMiddleware) RefreshPublicKeyCache(ctx context.Context) error {
	// Get all active keys
	activeKeys, err := m.jwtKeyRepo.FindAllActive(ctx)
	if err != nil {
		return err
	}

	// Clear and rebuild cache
	m.publicKeys = make(map[string]*rsa.PublicKey)

	for _, key := range activeKeys {
		publicKey, err := security.DecodePublicKeyPEM(key.PublicKey)
		if err != nil {
			m.logger.WithError(err).WithField("key_id", key.KeyID).Error("Failed to decode public key")
			continue
		}
		m.publicKeys[key.KeyID] = publicKey
	}

	m.logger.Info("Public key cache refreshed", "key_count", len(m.publicKeys))
	return nil
}
