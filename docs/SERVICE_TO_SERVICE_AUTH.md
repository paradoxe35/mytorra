# Service-to-Service Authentication Guide (Go)

## Overview

This guide provides detailed Go implementation for services that need to authenticate with the MyTorra Core Service and validate JWT tokens from other services using zero-trust architecture.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Complete Implementation](#complete-implementation)
3. [Middleware](#middleware)
4. [Caching Strategies](#caching-strategies)
5. [Error Handling](#error-handling)
6. [Service Registration](#service-registration)
7. [Monitoring](#monitoring)
8. [Best Practices](#best-practices)

## Quick Start

### Required Dependencies

```bash
go get github.com/golang-jwt/jwt/v5
go get github.com/go-redis/redis/v8  # Optional for distributed cache
go get github.com/gin-gonic/gin      # If using Gin framework
```

### Environment Variables

```bash
# Core service URL
CORE_SERVICE_URL=https://core.mytorra.internal

# Cache settings
JWKS_CACHE_TTL=3600  # 1 hour in seconds
JWKS_RETRY_ATTEMPTS=3
JWKS_RETRY_DELAY=1000  # milliseconds
```

## Complete Implementation

**Note**: This implementation leverages existing MyTorra packages from `/packages/` directory including:

- `github.com/paradoxe35/mytorra/packages/cache` - Cache implementation
- `github.com/paradoxe35/mytorra/packages/logger` - Logging interface
- `github.com/paradoxe35/mytorra/packages/config` - Configuration management
- `github.com/paradoxe35/mytorra/packages/errors` - Error handling
- `github.com/paradoxe35/mytorra/packages/security` - Security utilities

### 1. JWKS Types

```go
package auth

import (
    "crypto/rsa"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "math/big"
)

// JWKS represents a JSON Web Key Set
type JWKS struct {
    Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
    Kty string `json:"kty"` // Key Type
    Kid string `json:"kid"` // Key ID
    Use string `json:"use"` // Key Use
    Alg string `json:"alg"` // Algorithm
    N   string `json:"n"`   // RSA Modulus
    E   string `json:"e"`   // RSA Exponent
}

// GetPublicKey converts JWK to RSA public key
func (j *JWK) GetPublicKey() (*rsa.PublicKey, error) {
    // Decode modulus
    nBytes, err := base64.RawURLEncoding.DecodeString(j.N)
    if err != nil {
        return nil, fmt.Errorf("failed to decode modulus: %w", err)
    }

    // Decode exponent
    eBytes, err := base64.RawURLEncoding.DecodeString(j.E)
    if err != nil {
        return nil, fmt.Errorf("failed to decode exponent: %w", err)
    }

    // Convert to big.Int
    n := new(big.Int).SetBytes(nBytes)
    e := new(big.Int).SetBytes(eBytes)

    return &rsa.PublicKey{
        N: n,
        E: int(e.Int64()),
    }, nil
}

// GetKeyByID finds a key by its ID
func (jwks *JWKS) GetKeyByID(kid string) *JWK {
    for _, key := range jwks.Keys {
        if key.Kid == kid {
            return &key
        }
    }
    return nil
}
```

### 2. Service Authenticator (Using MyTorra Packages)

```go
package auth

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"

    "github.com/golang-jwt/jwt/v5"

    // Use MyTorra packages
    "github.com/paradoxe35/mytorra/packages/cache"
    "github.com/paradoxe35/mytorra/packages/logger"
    "github.com/paradoxe35/mytorra/packages/errors"
)

// ServiceAuthenticator handles JWT validation for services
type ServiceAuthenticator struct {
    jwksEndpoint string
    cache        cache.Cache      // Using MyTorra cache interface
    localCache   *JWKSCache       // Local in-memory cache
    httpClient   *http.Client
    logger       logger.Interface // Using MyTorra logger
}

// NewServiceAuthenticator creates a new authenticator
func NewServiceAuthenticator(
    coreServiceURL string,
    cacheClient cache.Cache,
    logger logger.Interface,
) *ServiceAuthenticator {
    return &ServiceAuthenticator{
        jwksEndpoint: fmt.Sprintf("%s/.well-known/jwks.json", coreServiceURL),
        cache:        cacheClient,
        localCache:   NewJWKSCache(time.Hour),
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
        logger: logger,
    }
}

// Claims represents JWT claims
type Claims struct {
    jwt.RegisteredClaims
    Email  string `json:"email"`
    Name   string `json:"name"`
    AreaID string `json:"area_id"`
}

// ValidateToken validates a JWT token
func (s *ServiceAuthenticator) ValidateToken(tokenString string) (*Claims, error) {
    // Get JWKS (cached)
    jwks, err := s.getJWKS()
    if err != nil {
        return nil, fmt.Errorf("failed to get JWKS: %w", err)
    }

    // Parse and validate token
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Verify signing method
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }

        // Get key ID
        kid, ok := token.Header["kid"].(string)
        if !ok {
            return nil, fmt.Errorf("missing key ID")
        }

        // Find the key
        jwk := jwks.GetKeyByID(kid)
        if jwk == nil {
            // Key not found, try refreshing JWKS once
            s.logger.Debug("Key not found, refreshing JWKS", "kid", kid)
            s.cache.Invalidate()

            jwks, err = s.getJWKS()
            if err != nil {
                return nil, fmt.Errorf("failed to refresh JWKS: %w", err)
            }

            jwk = jwks.GetKeyByID(kid)
            if jwk == nil {
                return nil, fmt.Errorf("key not found: %s", kid)
            }
        }

        // Convert JWK to RSA public key
        return jwk.GetPublicKey()
    })

    if err != nil {
        return nil, fmt.Errorf("token validation failed: %w", err)
    }

    if !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }

    claims, ok := token.Claims.(*Claims)
    if !ok {
        return nil, fmt.Errorf("invalid claims")
    }

    // Additional validation
    if err := s.validateClaims(claims); err != nil {
        return nil, err
    }

    return claims, nil
}

// validateClaims performs additional claim validation
func (s *ServiceAuthenticator) validateClaims(claims *Claims) error {
    // Check issuer
    if claims.Issuer != "mytorra-core" {
        return fmt.Errorf("invalid issuer: %s", claims.Issuer)
    }

    // Check expiration
    if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
        return fmt.Errorf("token expired")
    }

    // Check not before
    if claims.NotBefore != nil && claims.NotBefore.After(time.Now()) {
        return fmt.Errorf("token not yet valid")
    }

    return nil
}

// getJWKS fetches JWKS with caching
func (s *ServiceAuthenticator) getJWKS() (*JWKS, error) {
    // Try cache first
    if jwks, ok := s.cache.Get(); ok {
        s.logger.Debug("JWKS cache hit")
        return jwks, nil
    }

    s.logger.Debug("JWKS cache miss, fetching from endpoint")

    // Fetch with retry
    jwks, err := s.fetchJWKSWithRetry()
    if err != nil {
        // Try to return stale cache on error
        if staleJWKS := s.cache.GetStale(); staleJWKS != nil {
            s.logger.Error("Failed to fetch JWKS, using stale cache", "error", err)
            return staleJWKS, nil
        }
        return nil, err
    }

    // Update cache
    s.cache.Set(jwks)

    return jwks, nil
}

// fetchJWKSWithRetry fetches JWKS with retry logic
func (s *ServiceAuthenticator) fetchJWKSWithRetry() (*JWKS, error) {
    var lastErr error

    for attempt := 0; attempt < 3; attempt++ {
        if attempt > 0 {
            // Exponential backoff
            delay := time.Duration(attempt*attempt) * time.Second
            s.logger.Debug("Retrying JWKS fetch", "attempt", attempt, "delay", delay)
            time.Sleep(delay)
        }

        resp, err := s.httpClient.Get(s.jwksEndpoint)
        if err != nil {
            lastErr = fmt.Errorf("HTTP request failed: %w", err)
            continue
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            lastErr = fmt.Errorf("unexpected status: %d", resp.StatusCode)
            continue
        }

        var jwks JWKS
        if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
            lastErr = fmt.Errorf("failed to decode JWKS: %w", err)
            continue
        }

        s.logger.Info("JWKS fetched successfully", "keys", len(jwks.Keys))
        return &jwks, nil
    }

    return nil, fmt.Errorf("failed after 3 attempts: %w", lastErr)
}
```

### 3. JWKS Cache Implementation

```go
package auth

import (
    "sync"
    "time"
)

// JWKSCache provides thread-safe caching for JWKS
type JWKSCache struct {
    mu        sync.RWMutex
    jwks      *JWKS
    fetchedAt time.Time
    ttl       time.Duration
}

// NewJWKSCache creates a new cache
func NewJWKSCache(ttl time.Duration) *JWKSCache {
    return &JWKSCache{
        ttl: ttl,
    }
}

// Get returns cached JWKS if valid
func (c *JWKSCache) Get() (*JWKS, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    if c.jwks == nil {
        return nil, false
    }

    if time.Since(c.fetchedAt) > c.ttl {
        return nil, false
    }

    return c.jwks, true
}

// GetStale returns cached JWKS even if expired
func (c *JWKSCache) GetStale() *JWKS {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.jwks
}

// Set updates the cache
func (c *JWKSCache) Set(jwks *JWKS) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.jwks = jwks
    c.fetchedAt = time.Now()
}

// Invalidate clears the cache
func (c *JWKSCache) Invalidate() {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.jwks = nil
    c.fetchedAt = time.Time{}
}
```

## Middleware

### Standard HTTP Middleware

```go
package middleware

import (
    "context"
    "net/http"
    "strings"
)

// AuthMiddleware creates HTTP middleware for JWT validation
func AuthMiddleware(authenticator *ServiceAuthenticator) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token from header
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "Missing authorization header", http.StatusUnauthorized)
                return
            }

            // Check Bearer prefix
            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || parts[0] != "Bearer" {
                http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
                return
            }

            token := parts[1]

            // Validate token
            claims, err := authenticator.ValidateToken(token)
            if err != nil {
                http.Error(w, err.Error(), http.StatusUnauthorized)
                return
            }

            // Add claims to context
            ctx := context.WithValue(r.Context(), "claims", claims)
            ctx = context.WithValue(ctx, "user_id", claims.Subject)
            ctx = context.WithValue(ctx, "user_email", claims.Email)

            // Continue with authenticated request
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// GetClaims extracts claims from context
func GetClaims(ctx context.Context) (*Claims, bool) {
    claims, ok := ctx.Value("claims").(*Claims)
    return claims, ok
}
```

### Gin Framework Middleware

```go
package middleware

import (
    "strings"

    "github.com/gin-gonic/gin"
)

// GinAuthMiddleware creates Gin middleware for JWT validation
func GinAuthMiddleware(authenticator *ServiceAuthenticator) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract token from header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "Missing authorization header"})
            c.Abort()
            return
        }

        // Check Bearer prefix
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(401, gin.H{"error": "Invalid authorization format"})
            c.Abort()
            return
        }

        token := parts[1]

        // Validate token
        claims, err := authenticator.ValidateToken(token)
        if err != nil {
            c.JSON(401, gin.H{
                "error": "Invalid token",
                "details": err.Error(),
            })
            c.Abort()
            return
        }

        // Add claims to context
        c.Set("claims", claims)
        c.Set("user_id", claims.Subject)
        c.Set("user_email", claims.Email)
        c.Set("area_id", claims.AreaID)

        c.Next()
    }
}

// GetGinClaims extracts claims from Gin context
func GetGinClaims(c *gin.Context) (*Claims, bool) {
    if claims, exists := c.Get("claims"); exists {
        if typedClaims, ok := claims.(*Claims); ok {
            return typedClaims, true
        }
    }
    return nil, false
}
```

## Caching Strategies

### Using MyTorra Cache Package

```go
package auth

import (
    "context"
    "encoding/json"
    "time"

    "github.com/paradoxe35/mytorra/packages/cache"
    "github.com/paradoxe35/mytorra/packages/config"
)

// JWKSCacheManager manages JWKS caching using MyTorra cache
type JWKSCacheManager struct {
    cache cache.Cache
    ttl   time.Duration
}

// NewJWKSCacheManager creates a new cache manager
func NewJWKSCacheManager(cfg *config.RedisConfig) (*JWKSCacheManager, error) {
    // Use the existing Redis cache from MyTorra packages
    cacheClient, err := cache.NewRedisCacheAdapter(cfg, "jwks")
    if err != nil {
        return nil, err
    }

    return &JWKSCacheManager{
        cache: cacheClient,
        ttl:   time.Hour,
    }, nil
}

// GetJWKS retrieves JWKS from cache
func (m *JWKSCacheManager) GetJWKS(ctx context.Context) (*JWKS, error) {
    data, err := m.cache.Get(ctx, "public_keys")
    if err != nil {
        return nil, err // Cache miss or error
    }

    var jwks JWKS
    if err := json.Unmarshal(data, &jwks); err != nil {
        return nil, err
    }

    return &jwks, nil
}

// SetJWKS stores JWKS in cache
func (m *JWKSCacheManager) SetJWKS(ctx context.Context, jwks *JWKS) error {
    data, err := json.Marshal(jwks)
    if err != nil {
        return err
    }

    return m.cache.Set(ctx, "public_keys", data, m.ttl)
}

// InvalidateJWKS removes JWKS from cache
func (m *JWKSCacheManager) InvalidateJWKS(ctx context.Context) error {
    return m.cache.Delete(ctx, "public_keys")
}
```

## Error Handling

### Using MyTorra Errors Package

```go
package auth

import (
    "github.com/paradoxe35/mytorra/packages/errors"
)

// Define authentication-specific errors using MyTorra error types
var (
    ErrMissingToken = errors.New(
        errors.ErrorTypeAuthorization,
        "Authorization token is required",
    )

    ErrInvalidToken = errors.New(
        errors.ErrorTypeAuthorization,
        "The provided token is invalid",
    )

    ErrExpiredToken = errors.New(
        errors.ErrorTypeAuthorization,
        "The token has expired",
    )

    ErrKeyNotFound = errors.New(
        errors.ErrorTypeNotFound,
        "The signing key was not found",
    )

    ErrJWKSFetchFailed = errors.New(
        errors.ErrorTypeExternal,
        "Failed to fetch JWKS from endpoint",
    )
)

// WrapAuthError wraps an error with authentication context
func WrapAuthError(err error, message string) *errors.AppError {
    return errors.Wrap(err, errors.ErrorTypeAuthorization, message)
}

// HandleAuthError processes authentication errors
func HandleAuthError(err error) (int, *errors.AppError) {
    // Check if it's already an AppError
    if appErr := errors.GetAppError(err); appErr != nil {
        return appErr.HTTPStatus(), appErr
    }

    // Wrap generic errors
    return 401, errors.UnauthorizedError(err.Error())
}
```

## Service Registration

### Service-to-Service Authentication

```go
package auth

import (
    "bytes"
    "context"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

// ServiceClient handles service-to-service authentication
type ServiceClient struct {
    ServiceID     string
    APIKey        string
    WebhookSecret string
    httpClient    *http.Client
    baseURL       string
}

// NewServiceClient creates a new service client
func NewServiceClient(serviceID, apiKey, webhookSecret, baseURL string) *ServiceClient {
    return &ServiceClient{
        ServiceID:     serviceID,
        APIKey:        apiKey,
        WebhookSecret: webhookSecret,
        baseURL:       baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

// CallAPI makes an authenticated service-to-service API call
func (c *ServiceClient) CallAPI(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
    // Prepare request body
    var reqBody []byte
    var err error
    if body != nil {
        reqBody, err = json.Marshal(body)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal body: %w", err)
        }
    }

    // Create request
    url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
    req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    // Add authentication headers
    timestamp := fmt.Sprintf("%d", time.Now().Unix())
    signature := c.generateSignature(method, endpoint, timestamp, reqBody)

    req.Header.Set("X-Service-ID", c.ServiceID)
    req.Header.Set("X-Timestamp", timestamp)
    req.Header.Set("X-Signature", signature)
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
    req.Header.Set("Content-Type", "application/json")

    // Make request
    return c.httpClient.Do(req)
}

// generateSignature creates HMAC signature for request
func (c *ServiceClient) generateSignature(method, endpoint, timestamp string, body []byte) string {
    // Create message to sign
    message := fmt.Sprintf("%s:%s:%s:%s", method, endpoint, timestamp, string(body))

    // Generate HMAC
    h := hmac.New(sha256.New, []byte(c.APIKey))
    h.Write([]byte(message))

    return hex.EncodeToString(h.Sum(nil))
}

// VerifyWebhook verifies webhook signature
func (c *ServiceClient) VerifyWebhook(payload []byte, signature string) bool {
    h := hmac.New(sha256.New, []byte(c.WebhookSecret))
    h.Write(payload)
    expectedSig := hex.EncodeToString(h.Sum(nil))

    return hmac.Equal([]byte(signature), []byte(expectedSig))
}
```

## Monitoring

### Prometheus Metrics

```go
package auth

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    tokenValidationTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "jwt_token_validation_total",
            Help: "Total number of JWT token validations",
        },
        []string{"status"}, // success, failure
    )

    tokenValidationDuration = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name: "jwt_token_validation_duration_seconds",
            Help: "Duration of JWT token validation",
            Buckets: prometheus.DefBuckets,
        },
    )

    jwksFetchTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "jwks_fetch_total",
            Help: "Total number of JWKS fetch operations",
        },
        []string{"status"}, // success, failure, cache_hit
    )

    jwksFetchDuration = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name: "jwks_fetch_duration_seconds",
            Help: "Duration of JWKS fetch operations",
            Buckets: prometheus.DefBuckets,
        },
    )

    jwksCacheHits = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "jwks_cache_hits_total",
            Help: "Total number of JWKS cache hits",
        },
    )

    jwksCacheMisses = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "jwks_cache_misses_total",
            Help: "Total number of JWKS cache misses",
        },
    )
)

// ValidateTokenWithMetrics wraps token validation with metrics
func (s *ServiceAuthenticator) ValidateTokenWithMetrics(token string) (*Claims, error) {
    timer := prometheus.NewTimer(tokenValidationDuration)
    defer timer.ObserveDuration()

    claims, err := s.ValidateToken(token)

    if err != nil {
        tokenValidationTotal.WithLabelValues("failure").Inc()
        return nil, err
    }

    tokenValidationTotal.WithLabelValues("success").Inc()
    return claims, nil
}
```

## Best Practices

### 1. Configuration

```go
package config

import (
    "os"
    "strconv"
    "time"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
    CoreServiceURL   string
    JWKSCacheTTL     time.Duration
    RetryAttempts    int
    RetryDelay       time.Duration
    RequestTimeout   time.Duration
}

// LoadAuthConfig loads configuration from environment
func LoadAuthConfig() *AuthConfig {
    config := &AuthConfig{
        CoreServiceURL: getEnv("CORE_SERVICE_URL", "https://core.mytorra.internal"),
        JWKSCacheTTL:   getDuration("JWKS_CACHE_TTL", time.Hour),
        RetryAttempts:  getInt("JWKS_RETRY_ATTEMPTS", 3),
        RetryDelay:     getDuration("JWKS_RETRY_DELAY", time.Second),
        RequestTimeout: getDuration("REQUEST_TIMEOUT", 10*time.Second),
    }

    return config
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}
```

### 2. Circuit Breaker Pattern

```go
package auth

import (
    "fmt"
    "sync"
    "time"
)

// CircuitBreaker implements circuit breaker pattern for JWKS fetching
type CircuitBreaker struct {
    mu            sync.RWMutex
    failures      int
    lastFailTime  time.Time
    state         string // "closed", "open", "half-open"
    threshold     int
    timeout       time.Duration
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        state:     "closed",
        threshold: threshold,
        timeout:   timeout,
    }
}

// Call executes function with circuit breaker
func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mu.RLock()
    state := cb.state
    cb.mu.RUnlock()

    if state == "open" {
        cb.mu.RLock()
        if time.Since(cb.lastFailTime) > cb.timeout {
            cb.mu.RUnlock()
            cb.mu.Lock()
            cb.state = "half-open"
            cb.mu.Unlock()
        } else {
            cb.mu.RUnlock()
            return fmt.Errorf("circuit breaker is open")
        }
    }

    err := fn()

    cb.mu.Lock()
    defer cb.mu.Unlock()

    if err != nil {
        cb.failures++
        cb.lastFailTime = time.Now()

        if cb.failures >= cb.threshold {
            cb.state = "open"
        }

        return err
    }

    // Success
    if cb.state == "half-open" {
        cb.state = "closed"
    }
    cb.failures = 0

    return nil
}
```

### 3. Health Check Endpoint

```go
package health

import (
    "encoding/json"
    "net/http"
    "time"
)

// HealthChecker checks service health
type HealthChecker struct {
    authenticator *ServiceAuthenticator
}

// HealthStatus represents health check result
type HealthStatus struct {
    Status    string    `json:"status"`
    Timestamp time.Time `json:"timestamp"`
    Checks    map[string]CheckResult `json:"checks"`
}

// CheckResult represents individual check result
type CheckResult struct {
    Status  string `json:"status"`
    Message string `json:"message,omitempty"`
}

// Handler returns health check HTTP handler
func (h *HealthChecker) Handler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        status := h.Check()

        // Set status code based on health
        statusCode := http.StatusOK
        if status.Status != "healthy" {
            statusCode = http.StatusServiceUnavailable
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(statusCode)
        json.NewEncoder(w).Encode(status)
    }
}

// Check performs health check
func (h *HealthChecker) Check() *HealthStatus {
    status := &HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Checks:    make(map[string]CheckResult),
    }

    // Check JWKS availability
    jwksCheck := h.checkJWKS()
    status.Checks["jwks"] = jwksCheck

    if jwksCheck.Status != "healthy" {
        status.Status = "unhealthy"
    }

    return status
}

func (h *HealthChecker) checkJWKS() CheckResult {
    _, err := h.authenticator.getJWKS()
    if err != nil {
        return CheckResult{
            Status:  "unhealthy",
            Message: err.Error(),
        }
    }

    return CheckResult{
        Status: "healthy",
    }
}
```

## Complete Example Service

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"

    // MyTorra packages
    "github.com/paradoxe35/mytorra/packages/cache"
    "github.com/paradoxe35/mytorra/packages/config"
    "github.com/paradoxe35/mytorra/packages/logger"

    // Your service packages
    "your-service/auth"
    "your-service/middleware"
)

func main() {
    // Load MyTorra configuration
    cfg, err := config.Load("configs/config.yaml")
    if err != nil {
        log.Fatal("Failed to load configuration:", err)
    }

    // Initialize MyTorra logger
    appLogger := logger.New(cfg.Service.Name, cfg.Service.Environment)

    // Initialize MyTorra cache
    cacheClient, err := cache.NewRedisCacheAdapter(&cfg.Redis, "service")
    if err != nil {
        appLogger.Fatal("Failed to connect to cache")
    }
    defer func() {
        if closer, ok := cacheClient.(interface{ Close() error }); ok {
            closer.Close()
        }
    }()

    // Create authenticator using MyTorra packages
    authenticator := auth.NewServiceAuthenticator(
        cfg.CoreServiceURL,
        cacheClient,
        appLogger,
    )

    // Create Gin router
    router := gin.Default()

    // Public endpoints
    router.GET("/health", healthHandler)
    router.GET("/metrics", prometheusHandler())

    // Protected endpoints
    protected := router.Group("/api/v1")
    protected.Use(middleware.GinAuthMiddleware(authenticator))
    {
        protected.GET("/users/:id", getUserHandler)
        protected.POST("/orders", createOrderHandler)
    }

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Starting service on port %s", port)
    if err := router.Run(":" + port); err != nil {
        log.Fatal(err)
    }
}

func getUserHandler(c *gin.Context) {
    claims, _ := middleware.GetGinClaims(c)

    c.JSON(200, gin.H{
        "user_id": claims.Subject,
        "email":   claims.Email,
    })
}

func createOrderHandler(c *gin.Context) {
    claims, _ := middleware.GetGinClaims(c)

    // Your business logic here
    c.JSON(201, gin.H{
        "order_id": "12345",
        "user_id":  claims.Subject,
    })
}

func healthHandler(c *gin.Context) {
    c.JSON(200, gin.H{
        "status": "healthy",
    })
}

func prometheusHandler() gin.HandlerFunc {
    // Return Prometheus metrics handler
    return gin.WrapH(promhttp.Handler())
}
```

## Testing

### Unit Tests

```go
package auth_test

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "your-service/auth"
)

func TestJWKSCache(t *testing.T) {
    cache := auth.NewJWKSCache(time.Second)

    // Test empty cache
    _, ok := cache.Get()
    assert.False(t, ok)

    // Test set and get
    jwks := &auth.JWKS{
        Keys: []auth.JWK{
            {Kid: "test-key"},
        },
    }

    cache.Set(jwks)

    cached, ok := cache.Get()
    assert.True(t, ok)
    assert.Equal(t, "test-key", cached.Keys[0].Kid)

    // Test expiration
    time.Sleep(time.Second + 100*time.Millisecond)

    _, ok = cache.Get()
    assert.False(t, ok)

    // Test stale cache
    stale := cache.GetStale()
    assert.NotNil(t, stale)
    assert.Equal(t, "test-key", stale.Keys[0].Kid)
}

func TestTokenValidation(t *testing.T) {
    // Mock JWKS endpoint
    // Test token validation
    // Test error cases
}
```

## Troubleshooting

### Common Issues and Solutions

1. **"Key not found" errors**

   - Cause: JWKS cache is stale or key rotation in progress
   - Solution: Force refresh JWKS cache, implement retry logic

2. **Performance issues**

   - Cause: Too frequent JWKS fetches
   - Solution: Increase cache TTL, use distributed cache

3. **Network timeouts**

   - Cause: Core service unreachable
   - Solution: Implement circuit breaker, use stale cache as fallback

4. **Token validation failures**
   - Check token format and signature
   - Verify JWKS endpoint is accessible
   - Ensure clock synchronization (for exp/nbf claims)

### Debug Helper

```go
func DebugToken(tokenString string) {
    parts := strings.Split(tokenString, ".")
    if len(parts) != 3 {
        log.Printf("Invalid token format: expected 3 parts, got %d", len(parts))
        return
    }

    // Decode header
    headerJSON, _ := base64.RawURLEncoding.DecodeString(parts[0])
    log.Printf("Header: %s", headerJSON)

    // Decode claims
    claimsJSON, _ := base64.RawURLEncoding.DecodeString(parts[1])
    log.Printf("Claims: %s", claimsJSON)
}
```
