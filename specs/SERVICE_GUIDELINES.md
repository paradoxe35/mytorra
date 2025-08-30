# MyTorra Platform - Service Implementation Guidelines

## Overview

This document defines the mandatory standards, patterns, and best practices that MUST be followed when implementing any service within the MyTorra platform. These guidelines ensure consistency, maintainability, and professional quality across all services.

---

## 1. Service Architecture Requirements

### 1.1 Clean Architecture Structure (MANDATORY)

Every service MUST implement the following layered architecture:

```
services/<service-name>/
├── cmd/
│   ├── api/
│   │   └── main.go          # HTTP API server entry point
│   └── worker/              # OPTIONAL: Only if service needs background jobs
│       └── main.go          # Background worker entry point
├── internal/
│   ├── domain/              # Pure business logic, no dependencies
│   │   ├── entities/        # Core business entities
│   │   ├── valueobjects/    # Value objects
│   │   └── errors/          # Domain-specific errors
│   ├── application/         # Use cases and orchestration
│   │   ├── services/        # Application services
│   │   ├── ports/           # Interface definitions (contracts)
│   │   └── dtos/            # Data transfer objects
│   └── adapters/            # External world implementations
│       ├── http/            # HTTP handlers and middleware
│       ├── postgres/        # Database repositories
│       ├── redis/           # Cache implementations
│       ├── rabbitmq/        # Message bus handlers
│       └── grpc/            # gRPC clients/servers (if needed)
├── migrations/              # Database migrations
├── configs/                 # Configuration files
├── Dockerfile.api          # Docker image for API
├── Dockerfile.worker       # OPTIONAL: Docker image for Worker (if needed)
└── go.mod                  # Module dependencies
```

### 1.2 Separation of Concerns

- **Domain Layer**: Contains ONLY business logic, no framework dependencies
- **Application Layer**: Orchestrates domain logic, defines interfaces
- **Adapters Layer**: Implements interfaces, handles external communications
- **Dependency Direction**: Outer layers depend on inner layers, NEVER reverse

### 1.3 Interface-Driven Design

All external dependencies MUST be accessed through interfaces:

```go
// application/ports/user_repository.go
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*domain.User, error)
    Save(ctx context.Context, user *domain.User) error
    FindByEmail(ctx context.Context, email string) (*domain.User, error)
}

// adapters/postgres/user_repository.go
type PostgresUserRepository struct {
    db *sqlx.DB
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
    // Implementation
}
```

---

## 2. Go Development Standards

### 2.1 Required Go Version

- **Minimum Version**: Go 1.25 or later
- **Module Mode**: Always use Go modules (`go mod`)
- **Workspace**: Part of monorepo workspace (`go.work`)

### 2.2 Code Formatting & Linting

**MANDATORY Tools**:

- `gofmt`: All code MUST be formatted
- `golangci-lint`: MUST pass with zero errors
- Pre-commit hooks MUST be configured

**Required Linters Configuration**:

```yaml
# .golangci.yml
linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - ineffassign
    - gosec
    - gocritic
    - revive
    - misspell
    - unconvert
    - prealloc
    - nakedret
    - gocyclo
    - dupl
```

### 2.3 Error Handling

```go
// ALWAYS wrap errors with context
if err != nil {
    return fmt.Errorf("userRepo.FindByID failed for id %s: %w", userID, err)
}

// Define domain errors
var (
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidEmail = errors.New("invalid email format")
)

// Use errors.Is for checking
if errors.Is(err, ErrUserNotFound) {
    return nil, status.Error(codes.NotFound, "user not found")
}
```

### 2.4 Context Usage

```go
// Context MUST be first parameter for I/O operations
func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    // Use context for timeout/cancellation
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    return s.repo.FindByID(ctx, id)
}
```

---

## 3. HTTP API Standards

### 3.1 Framework & Router

- **Required**: Chi router (`github.com/go-chi/chi/v5`)
- **Middleware Stack Order**:
  1. Recovery (panic recovery)
  2. RequestID
  3. RealIP
  4. Logger
  5. Timeout
  6. Rate Limiter
  7. Authentication
  8. Authorization
  9. CORS (if needed)

### 3.2 API Design Principles

```go
// RESTful resource naming
router.Route("/api/v1/users", func(r chi.Router) {
    r.Get("/", h.ListUsers)        // GET /api/v1/users
    r.Post("/", h.CreateUser)       // POST /api/v1/users
    r.Get("/{id}", h.GetUser)       // GET /api/v1/users/{id}
    r.Put("/{id}", h.UpdateUser)    // PUT /api/v1/users/{id}
    r.Delete("/{id}", h.DeleteUser) // DELETE /api/v1/users/{id}
})

// Consistent response format
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *ErrorInfo  `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}
```

### 3.3 OpenAPI Documentation

- **MANDATORY**: All APIs MUST have OpenAPI 3.0 specs
- **Location**: `docs/api/<service>.v1.yaml`
- **Code Generation**: Use `oapi-codegen` for server stubs

---

## 4. Database Standards

### 4.1 PostgreSQL Requirements

- **Library**: `sqlx` (`github.com/jmoiron/sqlx`)
- **Migrations**: `golang-migrate`
- **Connection Pooling**: Configure max connections
- **Prepared Statements**: Use for repeated queries

### 4.2 Database Patterns

```go
// Repository pattern implementation
type PostgresUserRepo struct {
    db *sqlx.DB
}

// Use named queries for clarity
const getUserByEmailQuery = `
    SELECT id, email, name, created_at, updated_at
    FROM users
    WHERE email = :email AND deleted_at IS NULL
`

// Transaction support
func (r *PostgresUserRepo) CreateWithProfile(ctx context.Context, user *User, profile *Profile) error {
    tx, err := r.db.BeginTxx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Multiple operations
    if err := r.createUser(ctx, tx, user); err != nil {
        return err
    }
    if err := r.createProfile(ctx, tx, profile); err != nil {
        return err
    }

    return tx.Commit()
}
```

### 4.3 Migration Standards

```sql
-- migrations/001_create_users.up.sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users(created_at);
```

---

## 5. Messaging & Events

### 5.1 RabbitMQ Standards

```go
// Event definitions in packages/events/
type ProductListedEvent struct {
    EventID     string    `json:"event_id"`
    ProductID   string    `json:"product_id"`
    SellerID    string    `json:"seller_id"`
    Price       float64   `json:"price"`
    Currency    string    `json:"currency"`
    ListedAt    time.Time `json:"listed_at"`
    EventType   string    `json:"event_type"`
    Version     string    `json:"version"`
}

// Publisher interface
type EventPublisher interface {
    Publish(ctx context.Context, event interface{}) error
}

// Consumer MUST be idempotent
func (h *EventHandler) HandleProductListed(ctx context.Context, event ProductListedEvent) error {
    // Check if already processed
    if exists, err := h.repo.EventExists(ctx, event.EventID); exists || err != nil {
        return err
    }

    // Process event
    // ...

    // Mark as processed
    return h.repo.MarkEventProcessed(ctx, event.EventID)
}
```

### 5.2 Background Jobs (Asynq)

```go
// Job definitions
const (
    TypeEmailWelcome = "email:welcome"
    TypeImageResize  = "image:resize"
    TypeReportDaily  = "report:daily"
)

// Job payload
type EmailWelcomePayload struct {
    UserID string `json:"user_id"`
}

// Enqueue job (in API process)
func (s *UserService) SendWelcomeEmail(ctx context.Context, userID string) error {
    payload, _ := json.Marshal(EmailWelcomePayload{UserID: userID})
    task := asynq.NewTask(TypeEmailWelcome, payload)

    _, err := s.asynqClient.Enqueue(task,
        asynq.MaxRetry(3),
        asynq.Timeout(30*time.Second),
    )
    return err
}

// Process job (in Worker process)
func (w *EmailWorker) ProcessWelcomeEmail(ctx context.Context, t *asynq.Task) error {
    var payload EmailWelcomePayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return fmt.Errorf("unmarshal payload: %w", err)
    }

    // Send email logic
    return w.emailService.SendWelcome(ctx, payload.UserID)
}
```

---

## 6. Configuration Management

### 6.1 Viper Configuration

```go
// config/config.go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    RabbitMQ RabbitMQConfig
}

type ServerConfig struct {
    Port         int           `mapstructure:"port"`
    ReadTimeout  time.Duration `mapstructure:"read_timeout"`
    WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// Load configuration
func LoadConfig(path string) (*Config, error) {
    viper.SetConfigFile(path)
    viper.AutomaticEnv()
    viper.SetEnvPrefix("MYTORRA")
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }

    return &config, nil
}
```

### 6.2 Environment Variables

```bash
# Required environment variables
MYTORRA_DATABASE_URL=postgres://user:pass@localhost/mytorra
MYTORRA_REDIS_URL=redis://localhost:6379
MYTORRA_RABBITMQ_URL=amqp://guest:guest@localhost:5672/
MYTORRA_JWT_SECRET=${SECRET_FROM_VAULT}
MYTORRA_SERVICE_NAME=core
MYTORRA_ENVIRONMENT=development
```

---

## 7. Logging & Observability

### 7.1 Structured Logging (slog)

```go
// Initialize logger
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

// Add context fields
logger = logger.With(
    slog.String("service", "core"),
    slog.String("version", version),
    slog.String("environment", env),
)

// Use in handlers
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    logger := h.logger.With(
        slog.String("request_id", middleware.GetReqID(ctx)),
        slog.String("method", r.Method),
        slog.String("path", r.URL.Path),
    )

    userID := chi.URLParam(r, "id")
    logger.Info("fetching user", slog.String("user_id", userID))

    user, err := h.service.GetUser(ctx, userID)
    if err != nil {
        logger.Error("failed to fetch user",
            slog.String("user_id", userID),
            slog.String("error", err.Error()),
        )
        // Handle error
    }
}
```

### 7.2 Metrics & Tracing

```go
// OpenTelemetry setup
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

// Create tracer
tracer := otel.Tracer("mytorra.core")

// Use in functions
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    ctx, span := tracer.Start(ctx, "GetUser")
    defer span.End()

    span.SetAttributes(
        attribute.String("user.id", id),
    )

    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        span.RecordError(err)
        return nil, err
    }

    return user, nil
}
```

---

## 8. Testing Requirements (MANDATORY)

### 8.1 Test Coverage & Execution

- **Minimum Coverage**: 80% for business logic
- **MANDATORY**: Run `make test` before EVERY commit
- **Test Types Required**:
  - Unit tests for all domain logic
  - Integration tests for repositories
  - Contract tests for API endpoints
  - Load tests for critical paths
- **Test Execution**: ALL tests must pass before committing code
- **CI/CD**: Tests run automatically on every push

### 8.2 Test File Organization

```
service/
├── internal/
│   ├── domain/
│   │   ├── user.go
│   │   └── user_test.go          # Unit tests next to code
│   ├── application/
│   │   ├── user_service.go
│   │   └── user_service_test.go
│   └── adapters/
│       ├── postgres/
│       │   ├── user_repo.go
│       │   └── user_repo_test.go # Integration tests
│       └── http/
│           ├── user_handler.go
│           └── user_handler_test.go
└── tests/
    ├── integration/               # Full integration tests
    ├── e2e/                      # End-to-end tests
    └── load/                     # Load tests
```

### 8.3 Testing Commands

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests for specific service
make test-service SERVICE=core

# Run integration tests only
make test-integration

# Run tests with race detection
make test-race

# Generate coverage report
make coverage-html
```

### 8.4 Testing Patterns

```go
// Unit test example
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewMockUserRepository(t)
    service := NewUserService(mockRepo)

    user := &domain.User{
        Email: "test@example.com",
        Name:  "Test User",
    }

    mockRepo.On("Save", mock.Anything, user).Return(nil)

    // Act
    err := service.CreateUser(context.Background(), user)

    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}

// Integration test with testcontainers
func TestPostgresRepo_Integration(t *testing.T) {
    ctx := context.Background()

    // Start PostgreSQL container
    postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "postgres:15",
            ExposedPorts: []string{"5432/tcp"},
            Env: map[string]string{
                "POSTGRES_PASSWORD": "test",
            },
        },
        Started: true,
    })
    require.NoError(t, err)
    defer postgres.Terminate(ctx)

    // Run tests against real database
    // ...
}
```

---

## 9. Security Standards (CRITICAL)

### 9.1 SQL Injection Prevention

**NEVER construct SQL queries with string concatenation or fmt.Sprintf!**

```go
// ❌ DANGEROUS - NEVER DO THIS
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)

// ❌ DANGEROUS - SQL INJECTION VULNERABLE
query := "SELECT * FROM users WHERE id = " + userID

// ✅ SAFE - Use parameterized queries with sqlx
const getUserQuery = `
    SELECT id, email, name
    FROM users
    WHERE email = $1 AND deleted_at IS NULL
`
err := db.Get(&user, getUserQuery, email)

// ✅ SAFE - Use named parameters with sqlx
const updateUserQuery = `
    UPDATE users
    SET name = :name, updated_at = :updated_at
    WHERE id = :id AND deleted_at IS NULL
`
_, err := db.NamedExec(updateUserQuery, map[string]interface{}{
    "id":         userID,
    "name":       sanitizedName,
    "updated_at": time.Now(),
})

// ✅ SAFE - For dynamic queries, use query builders
import "github.com/Masterminds/squirrel"

query := squirrel.Select("*").From("users")
if email != "" {
    query = query.Where(squirrel.Eq{"email": email})
}
if status != "" {
    query = query.Where(squirrel.Eq{"status": status})
}
sql, args, err := query.ToSql()
```

### 9.2 Authentication & Authorization

```go
// JWT validation middleware
func JWTMiddleware(secret []byte) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractToken(r)
            if token == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            claims, err := validateToken(token, secret)
            if err != nil {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), "user_claims", claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### 9.3 Input Validation & Sanitization

```go
// MANDATORY: Validate ALL inputs
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email,max=255"`
    Name     string `json:"name" validate:"required,min=2,max=100,alphanumunicode"`
    Password string `json:"password" validate:"required,min=12,containsany=!@#$%,containsnumber"`
    Age      int    `json:"age" validate:"min=18,max=120"`
    Phone    string `json:"phone" validate:"required,e164"`
    URL      string `json:"url" validate:"omitempty,url,max=2048"`
}

// Sanitize user inputs
func sanitizeInput(input string) string {
    // Remove null bytes
    input = strings.ReplaceAll(input, "\x00", "")

    // Trim whitespace
    input = strings.TrimSpace(input)

    // Limit length
    if len(input) > 1000 {
        input = input[:1000]
    }

    return input
}

// HTML content sanitization
import "github.com/microcosm-cc/bluemonday"

func sanitizeHTML(input string) string {
    p := bluemonday.UGCPolicy()
    return p.Sanitize(input)
}

// File upload validation
func validateUpload(file multipart.File, header *multipart.FileHeader) error {
    // Check file size (10MB limit)
    if header.Size > 10*1024*1024 {
        return errors.New("file too large")
    }

    // Validate MIME type
    buffer := make([]byte, 512)
    _, err := file.Read(buffer)
    if err != nil {
        return err
    }

    contentType := http.DetectContentType(buffer)
    allowedTypes := map[string]bool{
        "image/jpeg": true,
        "image/png":  true,
        "image/gif":  true,
    }

    if !allowedTypes[contentType] {
        return errors.New("invalid file type")
    }

    // Reset file pointer
    file.Seek(0, 0)

    return nil
}

// Validate in handler
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    if err := h.validator.Struct(req); err != nil {
        respondError(w, http.StatusBadRequest, "validation failed")
        return
    }

    // Process valid request
}
```

### 9.4 XSS Prevention

```go
// Always escape output
import "html/template"

// Safe template rendering
func renderHTML(w http.ResponseWriter, data interface{}) {
    tmpl := template.Must(template.New("").Parse(`
        <h1>{{.Title | html}}</h1>
        <p>{{.Content | html}}</p>
    `))
    tmpl.Execute(w, data)
}

// JSON responses with proper content type
func respondJSON(w http.ResponseWriter, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Content-Type-Options", "nosniff")
    json.NewEncoder(w).Encode(data)
}
```

### 9.5 CSRF Protection

```go
import "github.com/gorilla/csrf"

// CSRF middleware setup
func SetupCSRF(authKey []byte) func(http.Handler) http.Handler {
    return csrf.Protect(authKey,
        csrf.Secure(true),
        csrf.HttpOnly(true),
        csrf.SameSite(csrf.SameSiteStrictMode),
        csrf.Path("/"),
    )
}

// In handlers
func (h *Handler) RenderForm(w http.ResponseWriter, r *http.Request) {
    token := csrf.Token(r)
    // Include token in form or header
    w.Header().Set("X-CSRF-Token", token)
}
```

### 9.6 Security Headers (Comprehensive)

```go
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Prevent MIME type sniffing
        w.Header().Set("X-Content-Type-Options", "nosniff")

        // Prevent clickjacking
        w.Header().Set("X-Frame-Options", "DENY")

        // XSS Protection (for older browsers)
        w.Header().Set("X-XSS-Protection", "1; mode=block")

        // Force HTTPS
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

        // Content Security Policy
        w.Header().Set("Content-Security-Policy",
            "default-src 'self'; "+
            "script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; "+
            "style-src 'self' 'unsafe-inline'; "+
            "img-src 'self' data: https:; "+
            "font-src 'self' data:; "+
            "connect-src 'self'; "+
            "frame-ancestors 'none'; "+
            "base-uri 'self'; "+
            "form-action 'self'")

        // Referrer Policy
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

        // Permissions Policy
        w.Header().Set("Permissions-Policy",
            "geolocation=(), microphone=(), camera=()")

        next.ServeHTTP(w, r)
    })
}
```

### 9.7 Password Security

```go
import "golang.org/x/crypto/bcrypt"

// Password hashing
func hashPassword(password string) (string, error) {
    // Use cost of 12 or higher
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    return string(bytes), err
}

// Password verification
func checkPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// Password strength validation
func validatePasswordStrength(password string) error {
    if len(password) < 12 {
        return errors.New("password must be at least 12 characters")
    }

    var hasUpper, hasLower, hasNumber, hasSpecial bool
    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsNumber(char):
            hasNumber = true
        case unicode.IsPunct(char) || unicode.IsSymbol(char):
            hasSpecial = true
        }
    }

    if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
        return errors.New("password must contain uppercase, lowercase, number and special character")
    }

    // Check against common passwords
    if isCommonPassword(password) {
        return errors.New("password is too common")
    }

    return nil
}
```

### 9.8 Secrets Management

```go
// NEVER hardcode secrets
// ❌ NEVER DO THIS
const apiKey = "sk_live_abcd1234"

// ✅ Load from environment or secret manager
func loadSecrets() (*Secrets, error) {
    // From environment
    secret := os.Getenv("API_SECRET")
    if secret == "" {
        // From secret manager (Vault, AWS Secrets Manager, etc.)
        client := vault.NewClient()
        secret, err := client.GetSecret("mytorra/api/secret")
        if err != nil {
            return nil, err
        }
    }

    return &Secrets{
        APISecret: secret,
    }, nil
}
```

### 9.9 Rate Limiting & DDoS Protection

```go
import "golang.org/x/time/rate"

// Per-IP rate limiting
type IPRateLimiter struct {
    visitors map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
    i.mu.Lock()
    defer i.mu.Unlock()

    limiter, exists := i.visitors[ip]
    if !exists {
        limiter = rate.NewLimiter(i.rate, i.burst)
        i.visitors[ip] = limiter
    }

    return limiter
}

func RateLimitMiddleware(rl *IPRateLimiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ip := getClientIP(r)
            limiter := rl.GetLimiter(ip)

            if !limiter.Allow() {
                http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### 9.10 Audit Logging

```go
// Log security events
type SecurityLogger struct {
    logger *slog.Logger
}

func (s *SecurityLogger) LogAuthAttempt(ctx context.Context, email string, success bool, ip string) {
    s.logger.Info("authentication attempt",
        slog.String("email", email),
        slog.Bool("success", success),
        slog.String("ip", ip),
        slog.String("request_id", middleware.GetReqID(ctx)),
        slog.Time("timestamp", time.Now()),
    )
}

func (s *SecurityLogger) LogSuspiciousActivity(ctx context.Context, activity string, details map[string]interface{}) {
    s.logger.Warn("suspicious activity detected",
        slog.String("activity", activity),
        slog.Any("details", details),
        slog.String("request_id", middleware.GetReqID(ctx)),
    )
}

// Monitor for security threats
func (s *SecurityLogger) MonitorForThreats(r *http.Request) {
    // Check for SQL injection patterns
    if containsSQLInjectionPattern(r.URL.RawQuery) {
        s.LogSuspiciousActivity(r.Context(), "potential_sql_injection", map[string]interface{}{
            "query": r.URL.RawQuery,
            "ip":    getClientIP(r),
        })
    }

    // Check for XSS patterns
    if containsXSSPattern(r.URL.RawQuery) {
        s.LogSuspiciousActivity(r.Context(), "potential_xss", map[string]interface{}{
            "query": r.URL.RawQuery,
            "ip":    getClientIP(r),
        })
    }
}
```

---

## 10. Performance Guidelines

### 10.1 Database Optimization

- Use connection pooling
- Implement query result caching
- Use prepared statements for repeated queries
- Add appropriate indexes
- Implement pagination for list endpoints
- Use batch operations where possible

### 10.2 Caching Strategy

```go
// Cache-aside pattern
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("user:%s", id)
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
        var user User
        if err := json.Unmarshal(cached, &user); err == nil {
            return &user, nil
        }
    }

    // Cache miss - get from database
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Update cache
    if data, err := json.Marshal(user); err == nil {
        s.cache.Set(ctx, cacheKey, data, 5*time.Minute)
    }

    return user, nil
}
```

### 10.3 Rate Limiting

```go
// Per-user rate limiting
limiter := rate.NewLimiter(rate.Every(time.Second), 10) // 10 requests per second

func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := getUserID(r.Context())
        limiter := getUserLimiter(userID)

        if !limiter.Allow() {
            http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

---

## 11. Deployment Standards

### 11.1 Docker Requirements

```dockerfile
# Dockerfile.api
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/api .
COPY --from=builder /app/configs ./configs
EXPOSE 8080
CMD ["./api"]
```

### 11.2 Health Checks

```go
// Health check endpoint
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
    checks := map[string]string{
        "database": "checking",
        "redis":    "checking",
        "rabbitmq": "checking",
    }

    // Check database
    if err := h.db.PingContext(r.Context()); err != nil {
        checks["database"] = "unhealthy"
    } else {
        checks["database"] = "healthy"
    }

    // Check other dependencies...

    healthy := true
    for _, status := range checks {
        if status != "healthy" {
            healthy = false
            break
        }
    }

    response := map[string]interface{}{
        "status": "ok",
        "checks": checks,
    }

    if !healthy {
        response["status"] = "degraded"
        w.WriteHeader(http.StatusServiceUnavailable)
    }

    json.NewEncoder(w).Encode(response)
}
```

### 11.3 Graceful Shutdown

```go
func main() {
    // Setup server
    srv := &http.Server{
        Addr:    ":8080",
        Handler: router,
    }

    // Start server
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("listen:", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
    <-quit

    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("server forced to shutdown:", err)
    }

    log.Println("server exited")
}
```

---

## 12. Documentation Requirements

### 12.1 Code Documentation

```go
// Package user provides user management functionality.
package user

// UserService handles all user-related business logic.
type UserService struct {
    repo UserRepository
    // other dependencies
}

// CreateUser creates a new user with the provided details.
// It validates the input, checks for duplicate emails, and stores the user.
// Returns the created user or an error if the operation fails.
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    // Implementation
}
```

### 12.2 API Documentation

- OpenAPI 3.0 specification required
- Postman collection for testing
- README with setup instructions
- Architecture decision records (ADRs)

### 12.3 Context7 Documentation Lookup (MANDATORY)

**When implementing new features or using external libraries:**

1. **Always Check Context7 First**: Before implementing any feature that involves external libraries, frameworks, or complex Go patterns, developers MUST:
   - Use Context7 to look up the latest documentation
   - Search for best practices and recent updates
   - Review code examples and implementation patterns
   - Check for security advisories or deprecation notices

2. **Examples of When to Use Context7:**
   - Implementing OAuth2 flows
   - Setting up database migrations
   - Configuring message queues (RabbitMQ)
   - Using new Go standard library features
   - Implementing WebSocket connections
   - Setting up observability (OpenTelemetry)
   - Configuring CI/CD pipelines
   - Using any third-party library

3. **Documentation Priority:**
   ```
   Context7 (Latest) → Official Docs → Community Resources → Stack Overflow
   ```

4. **Document Context7 Findings:**
   - When using Context7 to research implementation approaches, document key findings in code comments
   - Include Context7 reference links in ADRs for major architectural decisions
   - Update team knowledge base with important discoveries

**Example Usage:**
```go
// Implementation based on Context7 documentation for Chi router v5
// Ref: Context7 lookup performed on 2025-08-29 for latest middleware patterns
func SetupRouter() *chi.Mux {
    r := chi.NewRouter()

    // Middleware stack order based on Context7 best practices
    r.Use(middleware.Recovery)
    r.Use(middleware.RequestID)
    // ... rest of implementation
}
```

---

## 13. Development Best Practices (CRITICAL)

### 13.1 Package Reuse Policy (CRITICAL - MANDATORY)

**This is THE MOST IMPORTANT guideline for maintaining consistency across services**

#### Available Shared Packages

All services MUST use these shared packages from `/packages/` directory:

1. **Core Infrastructure Packages:**
   - `packages/database` - Database connections, transactions, query builders
   - `packages/config` - Configuration management and validation
   - `packages/logger` - Structured logging with context
   - `packages/errors` - Error types and handling
   - `packages/http` - HTTP response helpers and middleware

2. **Business Logic Packages:**
   - `packages/auth` - Authentication and authorization utilities
   - `packages/security` - Security utilities (hashing, encryption, validation)
   - `packages/cache` - Caching layer abstractions
   - `packages/events` - Event bus and messaging
   - `packages/i18n` - Internationalization support

3. **Domain Packages:**
   - `packages/dtos` - Shared DTOs across services
   - `packages/geo` - Geolocation utilities

#### Implementation Requirements

**MANDATORY CHECKS before implementing ANY feature:**

```bash
# 1. Check if package exists
ls -la /packages/

# 2. Check if functionality exists in packages
grep -r "YourFunctionality" /packages/

# 3. Check other services for similar implementations
grep -r "YourFunctionality" /services/
```

**Package Usage Example:**

```go
// CORRECT - Using shared packages
import (
    "github.com/paradoxe35/torra/packages/database"
    "github.com/paradoxe35/torra/packages/http"
    "github.com/paradoxe35/torra/packages/logger"
    "github.com/paradoxe35/torra/packages/errors"
    "github.com/paradoxe35/torra/packages/i18n"
)

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // Use shared http response
    http.RespondJSON(w, 201, user)

    // Use shared error handling
    if err != nil {
        http.RespondError(w, 400, "VALIDATION_ERROR", err.Error())
    }
}

// INCORRECT - Duplicating functionality
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // DON'T create custom response structures
    w.WriteHeader(201)
    json.NewEncoder(w).Encode(map[string]interface{}{...})
}
```

#### When to Create New Packages

Only create new shared packages when:
1. Functionality is needed by 2+ services
2. No existing package provides the functionality
3. The functionality is generic enough to be reused

Process for creating new packages:
1. Propose in team discussion
2. Create in `/packages/` directory
3. Add proper go.mod file
4. Include comprehensive tests
5. Update all services to use the new package

### 13.2 File and Directory Management (MANDATORY)

**ALWAYS verify existence before creating files or directories:**

1. **Before Creating Any File:**
   ```bash
   # Check if file exists
   ls -la /path/to/file.go
   # OR use Read tool to verify
   ```

2. **Before Creating Directories:**
   ```bash
   # Check directory structure
   ls -la /path/to/directory/
   ```

3. **Reuse Existing Packages:**
   - **MANDATORY**: Always check `packages/` directory for existing implementations
   - Import and use shared packages instead of duplicating code
   - Common reusable packages:
     ```go
     import (
         "github.com/paradoxe35/torra/packages/logger"
         "github.com/paradoxe35/torra/packages/errors"
         "github.com/paradoxe35/torra/packages/security"
         "github.com/paradoxe35/torra/packages/http"
         "github.com/paradoxe35/torra/packages/cache"
         "github.com/paradoxe35/torra/packages/config"
         "github.com/paradoxe35/torra/packages/i18n"
     )
     ```

4. **Package Priority Order:**
   ```
   Existing packages/ → Adapt if needed → Create new only if absolutely necessary
   ```

5. **Before Writing New Code:**
   - Search for similar implementations in the codebase
   - Check if functionality exists in packages/
   - Verify if another service already implements similar logic
   - Only create new implementations when existing code cannot be adapted

### 13.2 Inter-Service Communication

**Synchronous (REST)**:

- Use for immediate responses
- Implement circuit breakers
- Add retry logic with exponential backoff
- Set appropriate timeouts

**Asynchronous (Events)**:

- Use for fire-and-forget operations
- Ensure idempotency
- Implement dead letter queues
- Version event schemas

### 13.3 Service Discovery

```go
// Service registry interface
type ServiceRegistry interface {
    Register(service ServiceInfo) error
    Discover(serviceName string) ([]ServiceEndpoint, error)
    HealthCheck() error
}

// Use in client
func NewServiceClient(registry ServiceRegistry) *Client {
    endpoints, _ := registry.Discover("core-service")
    // Load balance between endpoints
}
```

---

## 14. Monitoring & Alerting

### 14.1 Key Metrics to Track

- Request rate, error rate, duration (RED)
- CPU, memory, disk usage
- Database connection pool metrics
- Queue depths and processing rates
- Business metrics (users, transactions, etc.)

### 14.2 Alerting Rules

```yaml
# Example Prometheus alert rules
groups:
  - name: service_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        annotations:
          summary: "High error rate detected"

      - alert: SlowRequests
        expr: histogram_quantile(0.95, http_request_duration_seconds) > 1
        for: 10m
        annotations:
          summary: "95th percentile latency is above 1s"
```

---

## 15. Development Workflow (MANDATORY)

### 15.1 Project Todo Tracking

**MANDATORY**: Keep the `project.todo` file updated throughout development:

- Mark tasks as completed when finished: `[x]`
- Update task status regularly
- Add new subtasks as discovered
- Document blockers or issues in task comments
- Review todo list at start and end of each work session

### 15.2 Before Every Commit

```bash
# 1. Update project.todo with completed tasks
# Mark completed items with [x]

# 2. Format your code
make fmt

# 3. Run linters
make lint

# 4. Run ALL tests
make test

# 5. Check test coverage
make test-coverage

# If all pass, then commit
git add .
git commit -m "feat: your feature description"
```

### 15.3 Before Opening Pull Request

```bash
# Run complete validation
make validate

# This runs:
# - Format check
# - Lint
# - All tests
# - Coverage check
# - Security scan
# - Build verification
```

## 16. Compliance Checklist

Before deploying any service to production, ensure:

- [ ] Clean architecture implemented
- [ ] All tests passing with >80% coverage
- [ ] `make validate` passes without errors
- [ ] OpenAPI documentation complete
- [ ] Security headers configured
- [ ] Rate limiting implemented
- [ ] Health checks working
- [ ] Graceful shutdown implemented
- [ ] Logging properly structured
- [ ] Metrics exposed for Prometheus
- [ ] Docker images optimized
- [ ] Configuration externalized
- [ ] Secrets managed securely
- [ ] Database migrations tested
- [ ] Error handling comprehensive
- [ ] Code reviewed and approved

---

## Conclusion

These guidelines are living standards that should be updated as the platform evolves. Every developer working on the MyTorra platform must familiarize themselves with these guidelines and follow them consistently.

For questions or clarifications, consult the architecture team or refer to the example implementations in the `services/core` directory.
