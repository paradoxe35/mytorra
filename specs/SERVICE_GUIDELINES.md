# MyTorra Platform - Service Implementation Guidelines

## Overview

This document defines the mandatory standards and patterns for implementing services in the MyTorra platform. These guidelines ensure consistency, maintainability, and professional quality across all services.

---

## 1. Service Architecture (MANDATORY)

### 1.1 Clean Architecture Structure

Every service MUST follow this structure:

```
services/<service-name>/
├── cmd/api/main.go          # HTTP API server entry point
├── cmd/worker/main.go       # Background worker entry point (using asynq)
├── internal/
│   ├── domain/              # Pure business logic
│   │   ├── entities/        # Core business entities
│   │   └── valueobjects/    # Value objects
│   ├── application/         # Use cases and orchestration
│   │   ├── services/        # Application services
│   │   ├── ports/           # Interface definitions
│   │   └── dtos/            # Data transfer objects
│   └── adapters/            # External implementations
│       ├── http/            # HTTP handlers and middleware
│       └── postgres/        # Database repositories
├── migrations/              # Database migrations
├── configs/config.yaml      # Configuration
└── go.mod                  # Module dependencies
```

### 1.2 Dependency Rules

- **Domain Layer**: No external dependencies
- **Application Layer**: Defines interfaces, orchestrates domain logic
- **Adapters Layer**: Implements interfaces, handles external communications
- **Direction**: Outer layers depend on inner layers only

---

## 2. Core Service Patterns (MANDATORY)

### 2.1 Context Keys Pattern

Use typed context keys for type safety:

```go
// Context keys for user information
type contextKey string

const (
    UserIDKey    contextKey = "user_id"
    UserEmailKey contextKey = "user_email"
    UserNameKey  contextKey = "user_name"
    AreaIDKey    contextKey = "area_id"
)

// Usage in handlers
userIDStr := ctx.Value(middleware.UserIDKey).(string)
```

### 2.2 Internationalization (i18n) Pattern

All services MUST support multiple languages:

```go
// Configuration
type I18nConfig struct {
    DefaultLanguage    string   `mapstructure:"default_language"`
    SupportedLanguages []string `mapstructure:"supported_languages"`
}

// Handler pattern
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    if err != nil {
        httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST",
            h.translator.TWithContext(ctx, "error.invalid_request_body"))
        return
    }
}

// Success messages
httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
    "message": h.translator.TWithContext(ctx, "user.created_successfully"),
})
```

### 2.3 Language Middleware Pattern

Automatic language detection from HTTP headers:

```go
// Add to router setup
r.Use(httpPkg.Language(translator, cfg.I18n.DefaultLanguage))

// Language middleware implementation
func Language(translator *i18n.Translator, defaultLang string) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            acceptLang := r.Header.Get("Accept-Language")
            lang := i18n.ParseAcceptLanguage(acceptLang, translator.GetSupportedLanguages())
            if lang == "" {
                lang = defaultLang
            }
            ctx := i18n.WithLanguage(r.Context(), lang)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### 2.4 Configuration Pattern

Configuration-driven initialization:

```go
// main.go initialization pattern
func main() {
    // Load configuration
    cfg, err := config.Load("configs/config.yaml")
    if err != nil {
        log.Fatal("Failed to load configuration:", err)
    }

    // Initialize translator from config
    translator, err := i18n.New(cfg.I18n.DefaultLanguage, cfg.I18n.SupportedLanguages...)
    if err != nil {
        appLogger.Fatal("Failed to initialize translator")
    }

    // Pass translator to all handlers
    authHandler := handlers.NewAuthHandler(authService, userService, appLogger, translator)
    userHandler := handlers.NewUserHandler(userService, appLogger, translator)
}
```

### 2.5 Error Handling Pattern

Use shared error package for consistency:

```go
// Import shared error package
import "github.com/paradoxe35/mytorra/packages/errors"

// Repository pattern
func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
    err := r.db.GetContext(ctx, &user, query, id)
    if err != nil {
        if database.IsNoRowsError(err) {
            return nil, fmt.Errorf("user not found: %w", err)
        }
        return nil, fmt.Errorf("failed to find user by id: %w", err)
    }
    return &user, nil
}

// Handler pattern with translations
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userIDStr := ctx.Value(middleware.UserIDKey).(string)

    user, err := h.userService.GetUser(ctx, userID)
    if err != nil {
        h.logger.WithError(err).Error("Failed to get user")
        httpPkg.RespondError(w, http.StatusNotFound, "USER_NOT_FOUND",
            h.translator.TWithContext(ctx, "user.not_found"))
        return
    }

    httpPkg.RespondJSON(w, http.StatusOK, user)
}
```

---

## 3. HTTP API Standards

### 3.1 Router Setup Pattern

```go
// Router initialization with middleware
func (rt *Router) Setup() *chi.Mux {
    r := chi.NewRouter()

    // Global middleware stack
    r.Use(middleware.Recoverer)
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Timeout(30 * time.Second))
    r.Use(httpPkg.SecurityHeaders)
    r.Use(httpPkg.CORS)
    r.Use(httpPkg.Language(rt.translator, rt.defaultLanguage)) // Language detection

    return r
}
```

### 3.2 Response Patterns

```go
// Success response
httpPkg.RespondJSON(w, http.StatusOK, data)

// Error response with translation
httpPkg.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR",
    h.translator.TWithContext(ctx, "error.validation_failed"))

// Success with message
httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
    "message": h.translator.TWithContext(ctx, "user.created_successfully"),
})
```

---

## 4. Core Service Analysis (CRITICAL - MANDATORY)

### 4.1 ALWAYS Inspect Core Service First

**Before implementing ANY feature, MANDATORY steps:**

```bash
# 1. Analyze core service structure
ls -la services/core/

# 2. Study related existing files
find services/core/ -name "*your_feature*" -type f

# 3. Check how core service implements similar functionality
grep -r "YourFunctionality" services/core/

# 4. Examine handler patterns
ls -la services/core/internal/adapters/http/handlers/

# 5. Study service patterns
ls -la services/core/internal/application/services/

# 6. Check repository implementations
ls -la services/core/internal/adapters/postgres/repositories/
```

### 4.2 Core Service Reference Implementation

**The core service (`services/core/`) is the REFERENCE implementation for:**

- Clean architecture structure
- Handler patterns with i18n
- Context key usage
- Error handling with translations
- Configuration management
- Database repository patterns
- Middleware implementation
- Testing patterns

**ALWAYS study core service before writing new code.**

### 4.3 Service-to-Service Authentication (CRITICAL)

**For services communicating with core service, MANDATORY reference:**

📖 **See complete implementation guide:** [`docs/SERVICE_TO_SERVICE_AUTH.md`](../docs/SERVICE_TO_SERVICE_AUTH.md)

**Key patterns from the guide:**

```go
// 1. Use shared packages for authentication
import (
    "github.com/paradoxe35/mytorra/packages/cache"
    "github.com/paradoxe35/mytorra/packages/logger"
    "github.com/paradoxe35/mytorra/packages/config"
)

// 2. JWT validation with JWKS
type ServiceAuthenticator struct {
    jwksEndpoint string
    cache        cache.Cache
    logger       logger.Interface
}

// 3. Middleware pattern (study core service auth middleware)
func AuthMiddleware(authenticator *ServiceAuthenticator) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract and validate JWT token
            // Add user context using typed keys
            // Continue with authenticated request
        })
    }
}
```

### 4.4 Required Shared Packages

All services MUST use these packages:

```go
import (
    "github.com/paradoxe35/mytorra/packages/config"     // Configuration management
    "github.com/paradoxe35/mytorra/packages/database"  // Database connections
    "github.com/paradoxe35/mytorra/packages/logger"    // Structured logging
    "github.com/paradoxe35/mytorra/packages/errors"    // Error handling
    "github.com/paradoxe35/mytorra/packages/http"      // HTTP responses
    "github.com/paradoxe35/mytorra/packages/i18n"      // Internationalization
    "github.com/paradoxe35/mytorra/packages/cache"     // Caching layer
    "github.com/paradoxe35/mytorra/packages/security"  // Security utilities
)
```

### 4.3 Database Pattern

```go
// Entity struct tags for database mapping (IMPORTANT)
// When using sqlx with named queries or direct scanning, entities need db tags
type User struct {
    ID        uuid.UUID `db:"id"`
    Email     string    `db:"email"`
    Name      string    `db:"name"`
    CreatedAt time.Time `db:"created_at"`
}

// Note: If entity has a separate model in adapters/postgres/models/,
// use the model for database operations instead of adding tags to domain entity

// Repository implementation using shared database package
type userRepository struct {
    db *database.DB
}

func NewUserRepository(db *database.DB) ports.UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
    var user entities.User
    query := `SELECT id, email, name FROM users WHERE id = $1 AND deleted_at IS NULL`

    err := r.db.GetContext(ctx, &user, query, id)
    if err != nil {
        if database.IsNoRowsError(err) {
            return nil, fmt.Errorf("user not found: %w", err)
        }
        return nil, fmt.Errorf("failed to find user: %w", err)
    }
    return &user, nil
}
```

---

## 5. Configuration Management

### 5.1 Configuration Structure

```go
// config/config.go - Based on core service pattern
type Config struct {
    Service  ServiceConfig  `mapstructure:"service"`
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    JWT      JWTConfig      `mapstructure:"jwt"`
    I18n     I18nConfig     `mapstructure:"i18n"`
}

type I18nConfig struct {
    DefaultLanguage    string   `mapstructure:"default_language"`
    SupportedLanguages []string `mapstructure:"supported_languages"`
}
```

### 5.2 Configuration File Pattern

```yaml
# configs/config.yaml
service:
  name: your-service
  version: 1.0.0
  environment: development

server:
  port: 8080
  host: 0.0.0.0
  read_timeout: 30s
  write_timeout: 30s

database:
  host: localhost
  port: 5432
  database: mytorra_service
  user: mytorra
  password: ${DB_PASSWORD}

i18n:
  default_language: en
  supported_languages:
    - en
    - fr
```

---

## 6. Testing Requirements (CRITICAL - MANDATORY)

### 6.1 ALWAYS Write Tests - NO EXCEPTIONS

**🚨 CRITICAL RULE: Every piece of code MUST have tests before being committed.**

```bash
# MANDATORY before every commit
make test

# Check test coverage (minimum 80%)
make test-coverage

# Run specific service tests
make test-service SERVICE=your-service
```

### 6.2 Test Types Required

**ALL services MUST implement:**

1. **Unit Tests** - Test business logic in isolation
2. **Integration Tests** - Test database repositories
3. **Handler Tests** - Test HTTP endpoints with mocked dependencies
4. **Service Tests** - Test application services

### 6.3 Testing Patterns (Study Core Service)

**Reference implementation:** `services/core/internal/*/`

```go
// Unit test example (study core service tests)
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewMockUserRepository(t)
    mockTranslator := mocks.NewMockTranslator(t)
    service := NewUserService(mockRepo, mockTranslator)

    user := &domain.User{
        Email: "test@example.com",
        Name:  "Test User",
    }

    mockRepo.On("Save", mock.Anything, user).Return(nil)
    mockTranslator.On("TWithContext", mock.Anything, mock.Anything).Return("translated message")

    // Act
    err := service.CreateUser(context.Background(), user)

    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
    mockTranslator.AssertExpectations(t)
}

// Handler test example (study core service handler tests)
func TestUserHandler_CreateUser(t *testing.T) {
    // Setup
    mockService := mocks.NewMockUserService(t)
    mockTranslator := mocks.NewMockTranslator(t)
    handler := NewUserHandler(mockService, logger.NewNoop(), mockTranslator)

    // Test request
    reqBody := `{"email":"test@example.com","name":"Test User"}`
    req := httptest.NewRequest("POST", "/users", strings.NewReader(reqBody))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    // Mock expectations
    mockService.On("CreateUser", mock.Anything, mock.Anything).Return(&domain.User{}, nil)

    // Execute
    handler.CreateUser(w, req)

    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
    mockService.AssertExpectations(t)
}

// Integration test example (study core service repository tests)
func TestUserRepository_Integration(t *testing.T) {
    // Use testcontainers for real database
    db := setupTestDatabase(t)
    defer db.Close()

    repo := NewUserRepository(db)

    // Test with real database
    user := &entities.User{
        Email: "test@example.com",
        Name:  "Test User",
    }

    err := repo.Save(context.Background(), user)
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

### 6.4 Test File Organization

**Follow core service pattern:**

```
services/your-service/
├── internal/
│   ├── domain/
│   │   ├── user.go
│   │   └── user_test.go          # Unit tests next to code
│   ├── application/
│   │   ├── services/
│   │   │   ├── user_service.go
│   │   │   └── user_service_test.go
│   └── adapters/
│       ├── http/handlers/
│       │   ├── user_handler.go
│       │   └── user_handler_test.go
│       └── postgres/repositories/
│           ├── user_repository.go
│           └── user_repository_test.go
```

### 6.5 Testing with i18n (Study Core Service)

```go
// Test translations in handlers
func TestHandler_WithTranslations(t *testing.T) {
    mockTranslator := mocks.NewMockTranslator(t)
    handler := NewHandler(mockService, logger.NewNoop(), mockTranslator)

    // Setup context with language
    ctx := i18n.WithLanguage(context.Background(), "fr")
    req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)

    // Expect French translation
    mockTranslator.On("TWithContext", ctx, "error.not_found").Return("Non trouvé")

    // Test error response
    w := httptest.NewRecorder()
    handler.GetUser(w, req)

    // Verify translated message in response
    assert.Contains(t, w.Body.String(), "Non trouvé")
}
```

---

## 7. Security Essentials (CRITICAL)

### 7.1 SQL Injection Prevention

**NEVER use string concatenation for SQL queries:**

```go
// ❌ DANGEROUS - NEVER DO THIS
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)

// ✅ SAFE - Use parameterized queries
const getUserQuery = `SELECT id, email, name FROM users WHERE email = $1`
err := db.Get(&user, getUserQuery, email)
```

### 7.2 Input Validation

```go
// MANDATORY: Validate ALL inputs
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email,max=255"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Password string `json:"password" validate:"required,min=12"`
}

// Validate in handler
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        httpPkg.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST",
            h.translator.TWithContext(r.Context(), "error.invalid_request_body"))
        return
    }

    if err := h.validator.Struct(req); err != nil {
        httpPkg.RespondError(w, http.StatusBadRequest, "VALIDATION_FAILED",
            h.translator.TWithContext(r.Context(), "error.validation_failed"))
        return
    }
}
```

### 7.3 Security Headers

```go
// Use shared security headers middleware
r.Use(httpPkg.SecurityHeaders)
```

---

## 8. Development Workflow (MANDATORY)

### 8.1 Before Writing ANY Code

**🚨 CRITICAL: Follow this sequence for EVERY feature:**

```bash
# 1. ANALYZE CORE SERVICE FIRST (MANDATORY)
ls -la services/core/internal/
find services/core/ -name "*similar_feature*" -type f
grep -r "SimilarFunctionality" services/core/

# 2. Check existing packages
ls -la packages/
grep -r "YourFunctionality" packages/

# 3. Study related implementations
grep -r "YourFunctionality" services/

# 4. Plan your tests BEFORE writing code
# - What will you test?
# - What mocks do you need?
# - What test cases are required?
```

### 8.2 Implementation Workflow

```bash
# 1. Write tests FIRST (TDD approach)
touch internal/domain/user_test.go
# Write failing tests

# 2. Implement code to make tests pass
touch internal/domain/user.go
# Write minimal code to pass tests

# 3. Refactor and improve
# Keep tests passing

# 4. Add integration tests
# Test with real dependencies

# 5. Add handler tests
# Test HTTP endpoints
```

### 8.3 Before Every Commit

- Run linters
- Run ALL tests (MANDATORY - NO EXCEPTIONS)
- Check test coverage (minimum 80%)
- Verify build

### 8.4 Core Service Study Checklist

**MANDATORY before implementing similar functionality:**

- [ ] Studied core service folder structure
- [ ] Analyzed related handler implementations
- [ ] Examined service layer patterns
- [ ] Reviewed repository implementations
- [ ] Checked middleware usage
- [ ] Studied error handling patterns
- [ ] Reviewed i18n integration
- [ ] Examined test patterns
- [ ] Verified shared package usage

## 9. Compliance Checklist

Before deploying any service to production, ensure:

**📋 Pre-Implementation Checklist:**

- [ ] Core service analyzed for similar functionality
- [ ] Related existing files studied
- [ ] Shared packages verified and used
- [ ] Service-to-service auth reviewed (if applicable)
- [ ] Test plan created before coding

**🏗️ Implementation Checklist:**

- [ ] Clean architecture implemented
- [ ] All tests written and passing (>80% coverage)
- [ ] Handler tests with i18n mocking
- [ ] Integration tests with real dependencies
- [ ] Unit tests for all business logic
- [ ] Shared packages used (no code duplication)
- [ ] i18n support implemented with translations
- [ ] Context keys properly typed (no string literals)
- [ ] Configuration externalized
- [ ] Error handling with translations

**🔒 Security Checklist:**

- [ ] Security headers configured
- [ ] Input validation implemented
- [ ] SQL injection prevention verified
- [ ] JWT authentication implemented (if needed)
- [ ] Service-to-service auth configured (if needed)

**🚀 Deployment Checklist:**

- [ ] Health checks working
- [ ] Graceful shutdown implemented
- [ ] Monitoring and metrics configured
- [ ] Docker images optimized
- [ ] All tests passing in CI/CD

---

## Conclusion

These guidelines are based on patterns proven in the core service implementation. Every developer must follow these standards consistently.

**🎯 Key Principles:**

1. **Study First**: Always analyze core service before implementing
2. **Test Everything**: Write tests before code, maintain >80% coverage
3. **Reuse First**: Check existing packages and implementations
4. **Translate Everything**: All user-facing text must support i18n
5. **Type Safety**: Use typed context keys and proper validation
6. **Security First**: Never compromise on security practices

**📚 Reference Sources:**

- **Core Service**: `services/core/` - Reference implementation for all patterns
- **Service Auth**: `docs/SERVICE_TO_SERVICE_AUTH.md` - Complete authentication guide
- **Shared Packages**: `packages/` - Reusable components

**❓ When in doubt:**

1. Check how core service implements it
2. Look for existing packages
3. Study related files in other services
4. Write comprehensive tests
