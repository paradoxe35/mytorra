package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	httpPkg "github.com/paradoxe35/torra/packages/http"
	"github.com/paradoxe35/torra/packages/logger"
)

type AuthMiddleware struct {
	jwtSecret []byte
	logger    *logger.Logger
}

func NewAuthMiddleware(jwtSecret string, logger *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: []byte(jwtSecret),
		logger:    logger,
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

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return m.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			m.logger.WithError(err).Debug("Invalid token")
			httpPkg.RespondError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token")
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			httpPkg.RespondError(w, http.StatusUnauthorized, "INVALID_CLAIMS", "Invalid token claims")
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), "user_id", claims["sub"])
		ctx = context.WithValue(ctx, "user_email", claims["email"])
		ctx = context.WithValue(ctx, "area_id", claims["area_id"])

		// Continue with the request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
