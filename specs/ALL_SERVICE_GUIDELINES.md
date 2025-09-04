# MyTorra Service Architecture Guidelines

## Overview

This document outlines the standardized architecture patterns, structure, and development guidelines for all MyTorra services. Each service follows Clean Architecture with Ports & Adapters pattern (Hexagonal Architecture).

## Service Structure

### Root Directory Layout

```
services/<service-name>/
├── cmd/
│   ├── api/          # API server entry point
│   │   └── main.go
│   └── worker/       # Background worker entry point (asynq)
│       └── main.go
├── configs/
│   └── config.yaml   # Service configuration
├── internal/
│   ├── adapters/     # External world interfaces
│   ├── application/  # Application business logic
│   └── domain/       # Core domain entities
│       ├── entities/     # Business entities
│       ├── events/       # Domain events (IMPORTANT)
│       └── valueobjects/ # Value objects
├── migrations/       # Database migrations (Atlas)
├── tests/           # Integration tests
├── Dockerfile.api
├── Dockerfile.worker
├── atlas.hcl        # Atlas DB schema config
├── go.mod
└── README.md
```

### API Server main.go Pattern

**MANDATORY**: Follow this initialization pattern for all services:

```go
// cmd/api/main.go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    // Shared packages (REQUIRED)
    "github.com/paradoxe35/mytorra/packages/cache"
    "github.com/paradoxe35/mytorra/packages/config"
    "github.com/paradoxe35/mytorra/packages/database"
    "github.com/paradoxe35/mytorra/packages/events"
    "github.com/paradoxe35/mytorra/packages/i18n"
    "github.com/paradoxe35/mytorra/packages/logger"
    "github.com/paradoxe35/mytorra/packages/security"

    // Internal packages
    adaptersHTTP "github.com/paradoxe35/mytorra/services/[service-name]/internal/adapters/http"
    "github.com/paradoxe35/mytorra/services/[service-name]/internal/adapters/http/handlers"
    "github.com/paradoxe35/mytorra/services/[service-name]/internal/adapters/http/middleware"
    "github.com/paradoxe35/mytorra/services/[service-name]/internal/adapters/postgres/repositories"
    "github.com/paradoxe35/mytorra/services/[service-name]/internal/application/services"
)

func main() {
    // 1. Load configuration
    cfg, err := config.NewServiceConfig()
    if err != nil {
        log.Fatal("Failed to load configuration:", err)
    }

    // 2. Initialize logger
    appLogger := logger.New(cfg.Service.Name, cfg.Service.Environment)

    // 3. Initialize translator
    translator, err := i18n.New(cfg.I18n.DefaultLanguage, cfg.I18n.SupportedLanguages...)
    if err != nil {
        appLogger.Fatal("Failed to initialize translator")
    }

    // 4. Initialize database
    db, err := database.New(&cfg.Database)
    if err != nil {
        appLogger.Fatal("Failed to connect to database")
    }
    defer db.Close()

    // 5. Initialize cache
    cacheClient, err := cache.NewRedisCacheAdapter(&cfg.Redis, cfg.Service.Name)
    if err != nil {
        appLogger.Fatal("Failed to connect to cache")
    }
    defer func() {
        if closer, ok := cacheClient.(interface{ Close() error }); ok {
            closer.Close()
        }
    }()

    // 6. Initialize event publisher
    eventPublisher, err := events.NewRabbitMQEventPublisher(&cfg.RabbitMQ, appLogger)
    if err != nil {
        appLogger.Warn(fmt.Sprintf("Failed to connect to RabbitMQ: %v. Using simple publisher", err))
        eventPublisher = &events.SimpleEventPublisher{Logger: appLogger}
    }
    defer eventPublisher.Close()

    // 7. Initialize repositories
    userRepo := repositories.NewUserRepository(db)
    // ... other repositories

    // 8. Initialize application services
    userService := services.NewUserService(
        userRepo,
        eventPublisher,
        cacheClient,
        appLogger,
        translator,
    )
    // ... other services

    // 9. Initialize handlers
    userHandler := handlers.NewUserHandler(userService, appLogger, translator)
    // ... other handlers

    // 10. Initialize middleware
    authMiddleware := middleware.NewAuthMiddleware(/* dependencies */)

    // 11. Setup router
    router := adaptersHTTP.NewRouter(
        userHandler,
        authMiddleware,
        translator,
        cfg.I18n.DefaultLanguage,
    )

    // 12. Create HTTP server
    srv := &http.Server{
        Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
        Handler:      router.Setup(),
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
        IdleTimeout:  cfg.Server.IdleTimeout,
    }

    // 13. Start server with graceful shutdown
    go func() {
        appLogger.Info(fmt.Sprintf("Server starting on port %d", cfg.Server.Port))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            appLogger.Fatal("Failed to start server", err)
        }
    }()

    // 14. Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // 15. Graceful shutdown
    appLogger.Info("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        appLogger.Fatal("Server forced to shutdown:", err)
    }

    appLogger.Info("Server exited")
}
```

## Architecture Layers

### 1. Domain Layer (`internal/domain/`)

Core business logic, independent of external concerns.

```
domain/
├── entities/        # Business entities
├── valueobjects/    # Value objects
└── errors/         # Domain-specific errors
```

#### Example Entity

```go
// internal/domain/entities/user.go
package entities

type User struct {
    ID              uuid.UUID
    Email           valueobjects.Email
    Phone           valueobjects.Phone
    Name            string
    PasswordHash    string
    Status          UserStatus
    CreatedAt       time.Time
}

// Domain methods
func (u *User) CanLogin() bool {
    return u.Status == UserStatusActive && u.EmailVerified
}
```

### 2. Application Layer (`internal/application/`)

Orchestrates use cases and application business logic.

```
application/
├── ports/           # Interface definitions
├── services/        # Application services
└── dtos/           # Data transfer objects
```

#### Port Interfaces

```go
// internal/application/ports/repositories.go
package ports

type UserRepository interface {
    FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
    FindByEmail(ctx context.Context, email valueobjects.Email) (*entities.User, error)
    Save(ctx context.Context, user *entities.User) error
    Update(ctx context.Context, user *entities.User) error
}
```

#### Application Service

```go
// internal/application/services/auth_service.go
package services

type AuthService struct {
    userRepo         ports.UserRepository
    refreshTokenRepo ports.RefreshTokenRepository
    cache            ports.Cache
    logger           logger.Interface
}

func NewAuthService(deps...) *AuthService {
    return &AuthService{...}
}
```

### 3. Adapters Layer (`internal/adapters/`)

Implements ports and connects to external systems.

```
adapters/
├── http/
│   ├── handlers/    # HTTP handlers
│   ├── middleware/  # HTTP middleware
│   ├── router.go    # Main router configuration (REQUIRED)
│   └── validators/  # Request validators
├── postgres/
│   ├── models/      # Bun ORM models (*_model.go files)
│   ├── repositories/# Repository implementations
│   └── types/       # Custom DB types
├── payment/         # Payment gateway adapters
├── oauth/          # OAuth providers
└── grpc/           # gRPC clients/servers
```

**IMPORTANT**: All adapters go in `internal/adapters/` directory

#### Repository Implementation

```go
// internal/adapters/postgres/repositories/user_repository.go
package repositories

require (
    "context"
    "database/sql"
    "github.com/paradoxe35/mytorra/packages/database"
    "github.com/paradoxe35/mytorra/packages/errors"
    "github.com/paradoxe35/mytorra/services/[service-name]/internal/application/ports"
    "github.com/paradoxe35/mytorra/services/[service-name]/internal/domain/entities"
    "github.com/paradoxe35/mytorra/services/[service-name]/internal/adapters/postgres/models"
    "github.com/google/uuid"
)

type userRepository struct {
    db *database.DB
}

func NewUserRepository(db *database.DB) ports.UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
    var userModel models.UserModel

    err := r.db.NewSelect().
        Model(&userModel).
        Where("id = ?", id).
        Scan(ctx)

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.ErrNotFound
        }
        return nil, err
    }

    return userModel.ToEntity()
}
```

## Shared Packages (`packages/`)

**CRITICAL**: All services MUST use these shared packages - DO NOT reinvent the wheel!

```
packages/
├── cache/          # Redis cache adapter (includes MockCache)
├── config/         # Configuration loader
├── database/       # Database connection (Bun + pgx, includes TestDB helpers)
├── errors/         # Common error types
├── events/         # Event publishing (includes MockEventPublisher)
├── http/           # HTTP utilities
├── i18n/           # Internationalization
├── logger/         # Structured logging
└── security/       # Security utilities
```

### Required Package Imports

**MANDATORY**: All services MUST use these packages:

```go
import (
    "github.com/paradoxe35/mytorra/packages/config"     // Configuration management
    "github.com/paradoxe35/mytorra/packages/database"   // Database connections
    "github.com/paradoxe35/mytorra/packages/logger"     // Structured logging
    "github.com/paradoxe35/mytorra/packages/errors"     // Error handling
    "github.com/paradoxe35/mytorra/packages/http"       // HTTP responses
    "github.com/paradoxe35/mytorra/packages/i18n"       // Internationalization
    "github.com/paradoxe35/mytorra/packages/cache"      // Caching layer
    "github.com/paradoxe35/mytorra/packages/security"   // Security utilities
    "github.com/paradoxe35/mytorra/packages/events"     // Event publishing
)
```

### Built-in Mocks

**IMPORTANT**: Most packages provide their own mocks - USE THEM:

```go
// cache package provides MockCache
mockCache := cache.NewMockCache()

// events package provides MockEventPublisher
mockEventPublisher := events.NewMockEventPublisher()

// database package provides TestDB for testing
testDB := database.NewTestDB(t)

// logger package for tests
testLogger := logger.New("test", "test")
```

### Package Usage Example

```go
// Load configuration
cfg, err := config.Load("configs/config.yaml")

// Initialize database
db, err := database.New(&cfg.Database)
defer db.Close()

// Initialize cache
cacheClient, err := cache.NewRedisCacheAdapter(&cfg.Redis, "service-name")
```

## Database Management with Bun ORM

### CRITICAL: Model-Driven Migrations with Atlas

**DO NOT create migrations manually!** Atlas automatically generates migrations by detecting Bun ORM models.

### Atlas Schema Management

Atlas reads your Bun models and generates migrations automatically:

```hcl
# atlas.hcl
data "external_schema" "bun" {
  program = [
    "go", "run", "-mod=mod",
    "ariga.io/atlas-provider-bun",
    "load",
    "--path", "./internal/adapters/postgres/models",  # Atlas scans this directory
    "--dialect", "postgres"
  ]
}

env "bun" {
  src = data.external_schema.bun.url
  dev = "docker://postgis/17-3.5-alpine/dev"
  migration {
    dir = "file://migrations"  # Auto-generated migrations go here
  }
}
```

### Model Definition Requirements

**MANDATORY**: All database models MUST:

1. Be placed in `internal/adapters/postgres/models/`
2. Use `*_model.go` suffix (e.g., `user_model.go`, `area_model.go`)
3. Extend `database.BaseModel` from shared package
4. Include conversion methods (`ToEntity()` and `FromEntity()`)

```go
// internal/adapters/postgres/models/user_model.go
package models

import (
    "github.com/uptrace/bun"
    "github.com/paradoxe35/mytorra/packages/database"
    "github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
)

// REQUIRED: Extend database.BaseModel which includes ID, CreatedAt, UpdatedAt, DeletedAt
type UserModel struct {
    database.BaseModel `bun:"table:users,alias:u"`

    // Define all fields with Bun tags
    Email           string  `bun:"email,notnull,unique"`
    Phone           string  `bun:"phone,unique"`
    Name            string  `bun:"name,notnull"`
    PasswordHash    string  `bun:"password_hash,notnull"`
    ProfileImageURL string  `bun:"profile_image_url"`

    // Foreign keys
    AreaID  uuid.UUID  `bun:"area_id,type:uuid"`
    Area    *AreaModel `bun:"rel:belongs-to,join:area_id=id"`

    // Status fields
    Status      string     `bun:"status,default:'active'"`
    LastLoginAt *time.Time `bun:"last_login_at"`
}

// REQUIRED: Convert database model to domain entity
func (m *UserModel) ToEntity() (*entities.User, error) {
    email, err := valueobjects.NewEmail(m.Email)
    if err != nil {
        return nil, err
    }

    return &entities.User{
        ID:    m.ID,
        Email: email,
        // ... map all fields
    }, nil
}

// REQUIRED: Convert domain entity to database model
func UserFromEntity(u *entities.User) *UserModel {
    return &UserModel{
        BaseModel: database.BaseModel{
            ID:        u.ID,
            CreatedAt: u.CreatedAt,
            UpdatedAt: u.UpdatedAt,
        },
        Email: u.Email.String(),
        // ... map all fields
    }
}
```

### Migration Workflow

```bash
# 1. Create/modify your Bun models in internal/adapters/postgres/models/

# 2. Atlas automatically detects changes and generates migration
atlas migrate diff --env bun

# 3. Apply migrations
atlas migrate apply --env bun

# NEVER manually write SQL migrations - let Atlas handle it!
```

## Background Workers

### Asynq Worker Implementation

```go
// cmd/worker/main.go
package main

import (
    "github.com/hibiken/asynq"
    "github.com/paradoxe35/mytorra/packages/config"
    "github.com/paradoxe35/mytorra/packages/logger"
)

func main() {
    cfg, err := config.Load("configs/config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // Create Asynq server
    srv := asynq.NewServer(
        asynq.RedisClientOpt{
            Addr:     cfg.Redis.Address(),
            Password: cfg.Redis.Password,
            DB:       cfg.Redis.DB,
        },
        asynq.Config{
            Concurrency: 10,
            Queues: map[string]int{
                "critical": 6,
                "default":  3,
                "low":      1,
            },
        },
    )

    // Register task handlers
    mux := asynq.NewServeMux()
    mux.HandleFunc(tasks.TypeEmailDelivery, handleEmailDelivery)
    mux.HandleFunc(tasks.TypeUserSync, handleUserSync)

    if err := srv.Run(mux); err != nil {
        log.Fatal(err)
    }
}
```

### Task Handler Example

```go
// internal/workers/tasks/email_task.go
package tasks

const TypeEmailDelivery = "email:deliver"

type EmailDeliveryPayload struct {
    UserID    uuid.UUID
    Template  string
    Variables map[string]interface{}
}

func NewEmailDeliveryTask(payload EmailDeliveryPayload) (*asynq.Task, error) {
    bytes, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }
    return asynq.NewTask(TypeEmailDelivery, bytes), nil
}

func HandleEmailDelivery(ctx context.Context, t *asynq.Task) error {
    var p EmailDeliveryPayload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return fmt.Errorf("unmarshal payload: %v", err)
    }

    // Process email delivery
    return emailService.Send(ctx, p)
}
```

## API Handler Pattern

### Router Configuration (REQUIRED)

**MANDATORY**: Every service MUST have a router at `internal/adapters/http/router.go`

```go
// internal/adapters/http/router.go
package http

import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    httpPkg "github.com/paradoxe35/mytorra/packages/http"
    "github.com/paradoxe35/mytorra/packages/i18n"
    "github.com/paradoxe35/mytorra/services/core/internal/adapters/http/handlers"
    httpMiddleware "github.com/paradoxe35/mytorra/services/core/internal/adapters/http/middleware"
)

type Router struct {
    authHandler    *handlers.AuthHandler
    userHandler    *handlers.UserHandler
    authMiddleware *httpMiddleware.AuthMiddleware
    translator     *i18n.Translator
    defaultLanguage string
}

func (rt *Router) Setup() *chi.Mux {
    r := chi.NewRouter()

    // Global middleware (ORDER MATTERS!)
    r.Use(middleware.Recoverer)
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Timeout(30 * time.Second))
    r.Use(httpPkg.SecurityHeaders)
    r.Use(httpPkg.CORS)
    r.Use(httpPkg.Language(rt.translator, rt.defaultLanguage))

    // Health checks (public)
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        httpPkg.RespondJSON(w, http.StatusOK, map[string]string{
            "status": "healthy",
        })
    })

    // API routes
    r.Route("/api/v1", func(r chi.Router) {
        // Public routes
        r.Route("/auth", func(r chi.Router) {
            r.Post("/register", rt.authHandler.Register)
            r.Post("/login", rt.authHandler.Login)
        })

        // Protected routes
        r.Group(func(r chi.Router) {
            r.Use(rt.authMiddleware.Authenticate)

            r.Route("/users", func(r chi.Router) {
                r.Get("/me", rt.userHandler.GetCurrentUser)
                r.Put("/me", rt.userHandler.UpdateCurrentUser)
            })
        })
    })

    return r
}
```

### Middleware Organization

**MANDATORY**: Custom middleware goes in `internal/adapters/http/middleware/`

```go
// internal/adapters/http/middleware/auth.go
package middleware

type AuthMiddleware struct {
    jwtService     *services.JWTService
    userService    *services.UserService
    logger         logger.Interface
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract token
        // Validate token
        // Add user context
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Handler Implementation

```go
// internal/adapters/http/handlers/auth_handler.go
package handlers

func Login(authService *services.AuthService) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req dtos.LoginRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            httputil.WriteError(w, errors.ErrInvalidRequest)
            return
        }

        response, err := authService.Login(r.Context(), req)
        if err != nil {
            httputil.WriteError(w, err)
            return
        }

        httputil.WriteJSON(w, http.StatusOK, response)
    }
}
```

## Event-Driven Communication

### Domain Events Pattern (MANDATORY)

**CRITICAL**: All events MUST be defined in `internal/domain/events/` directory

#### 1. Define Domain Events

```go
// internal/domain/events/billing_events.go
package events

import (
    "time"
    "github.com/google/uuid"
    "github.com/shopspring/decimal"
)

// BillingPaymentFailed event is published when a payment fails
type BillingPaymentFailed struct {
    SubscriptionID uuid.UUID       `json:"subscription_id"`
    BillingCycleID uuid.UUID       `json:"billing_cycle_id"`
    TransactionID  uuid.UUID       `json:"transaction_id"`
    Amount         decimal.Decimal `json:"amount"`
    Currency       string          `json:"currency"`
    FailureReason  string          `json:"failure_reason"`
    FailedAt       time.Time       `json:"failed_at"`
}

// REQUIRED: Implement event interface methods
func (e BillingPaymentFailed) EventType() string {
    return "billing.payment_failed"
}

func (e BillingPaymentFailed) EventData() interface{} {
    return e
}

// BillingPaymentSuccessful event
type BillingPaymentSuccessful struct {
    SubscriptionID uuid.UUID       `json:"subscription_id"`
    BillingCycleID uuid.UUID       `json:"billing_cycle_id"`
    TransactionID  uuid.UUID       `json:"transaction_id"`
    Amount         decimal.Decimal `json:"amount"`
    Currency       string          `json:"currency"`
    ProcessedAt    time.Time       `json:"processed_at"`
}

func (e BillingPaymentSuccessful) EventType() string {
    return "billing.payment_successful"
}

func (e BillingPaymentSuccessful) EventData() interface{} {
    return e
}
```

#### 2. Publish Events from Services

**ALWAYS** use the events package from `packages/events/` for publishing:

```go
// internal/application/services/billing_service.go
package services

import (
    "github.com/paradoxe35/mytorra/packages/events"  // REQUIRED: Use package events
    domainEvents "github.com/paradoxe35/mytorra/services/core/internal/domain/events"
)

type BillingService struct {
    eventBus events.EventPublisher  // From packages/events
    // ... other dependencies
}

func (b *BillingService) handlePaymentFailure(ctx context.Context, ...) {
    // Business logic...

    // Publish domain event using packages/events publisher
    b.eventBus.Publish(ctx, domainEvents.BillingPaymentFailed{
        SubscriptionID: subscription.ID,
        BillingCycleID: cycle.ID,
        TransactionID:  transaction.ID,
        Amount:         cycle.Amount,
        Currency:       cycle.Currency,
        FailureReason:  reason,
        FailedAt:       time.Now(),
    })
}

func (b *BillingService) processBillingSuccess(ctx context.Context, ...) {
    // Business logic...

    // Publish success event
    b.eventBus.Publish(ctx, domainEvents.BillingPaymentSuccessful{
        SubscriptionID: subscription.ID,
        BillingCycleID: cycle.ID,
        TransactionID:  transaction.ID,
        Amount:         cycle.Amount,
        Currency:       cycle.Currency,
        ProcessedAt:    time.Now(),
    })
}
```

#### 3. Event Naming Conventions

- **Event Types**: Use dot notation (e.g., `billing.payment_failed`, `user.created`)
- **Event Structs**: Use descriptive names (e.g., `BillingPaymentFailed`, `UserCreated`)
- **Event Files**: Group by domain context (e.g., `billing_events.go`, `user_events.go`)

### Event Publishing Best Practices

1. **Define events in domain layer** - Not in application or adapter layers
2. **Use packages/events publisher** - Never create your own publisher
3. **Include all relevant data** - Events should be self-contained
4. **Use proper timestamps** - Always include when the event occurred
5. **Don't block on publishing** - Log errors but continue processing

## Testing Strategy

### Unit Tests

```go
require (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
    "github.com/paradoxe35/mytorra/packages/cache"
    "github.com/paradoxe35/mytorra/packages/logger"
    "github.com/paradoxe35/mytorra/services/core/internal/application/services"
    "github.com/paradoxe35/mytorra/services/core/internal/application/ports/mocks"
    "github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)
// internal/application/services/auth_service_test.go
func TestAuthService_Login(t *testing.T) {
    // Setup
    userRepo := mocks.NewMockUserRepository(t)
    cacheRepo := cache.NewMockCache(t)

    service := services.NewAuthService(userRepo, cacheRepo, ...)

    // Given
    userRepo.On("FindByEmail", mock.Anything, email).Return(user, nil)

    // When
    result, err := service.Login(context.Background(), loginReq)

    // Then
    assert.NoError(t, err)
    assert.NotEmpty(t, result.AccessToken)
    userRepo.AssertExpectations(t)
}
```

## Service Configuration

### Config Structure

```yaml
# configs/config.yaml
service:
  name: service-name
  version: 1.0.0
  environment: development
  port: 8080

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  database: mytorra_core
  max_open_conns: 25
  max_idle_conns: 5

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

rabbitmq:
  url: amqp://guest:guest@localhost:5672/

jwt:
  secret: your-secret-key
  issuer: mytorra-core
  access_duration: 15m
  refresh_duration: 7d
```

## Deployment

### Docker Configuration

```dockerfile
# Dockerfile.api
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o api cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/api .
COPY --from=builder /app/configs ./configs
CMD ["./api"]
```

## Best Practices

### 1. Dependency Injection

- Use constructor injection
- Define interfaces in the application layer
- Implement in the adapters layer

### 2. Error Handling

- Use domain-specific errors
- Wrap external errors with context
- Return appropriate HTTP status codes

### 3. Context Usage

- Pass context through all layers
- Use for cancellation and timeouts
- Store request-scoped values

### 4. Logging

- Structured logging with fields
- Log at appropriate levels
- Include request IDs for tracing

### 5. Database Transactions

- Use transactions for multi-step operations
- Handle rollbacks properly
- Use appropriate isolation levels

### 6. Caching Strategy

- Cache at the repository level
- Use TTL appropriately
- Invalidate on updates

### 7. Security

- Validate all inputs
- Use parameterized queries
- Hash passwords with bcrypt
- Implement rate limiting
- Use CORS properly

### 8. API Versioning

- Use URL path versioning (/api/v1/)
- Maintain backward compatibility
- Document breaking changes

## Service Communication

### Synchronous (HTTP/gRPC)

- For real-time responses
- Service discovery via DNS/Consul
- Circuit breaker pattern
- Retry with exponential backoff

### Asynchronous (Events)

- For eventual consistency
- Use RabbitMQ/Kafka
- Idempotent handlers
- Dead letter queues

## Monitoring & Observability

### Metrics

- Prometheus metrics
- Custom business metrics
- SLI/SLO tracking

### Tracing

- OpenTelemetry integration
- Distributed tracing
- Request correlation IDs

### Logging

- Structured JSON logs
- Log aggregation (ELK/Loki)
- Error tracking (Sentry)

## Testing Philosophy & Requirements

### Why Testing is Critical

Testing is mandatory in the MyTorra platform. Every service MUST have comprehensive tests:

1. **Reliability**: We handle financial transactions and business operations
2. **Confidence**: Tests enable rapid development and fearless refactoring
3. **Documentation**: Tests serve as living documentation
4. **Quality Gates**: No code merges without passing tests
5. **Regression Prevention**: Catch bugs before production

### Testing Requirements

#### Minimum Coverage Thresholds

- **Overall Coverage**: ≥ 80%
- **Business Logic (Services)**: ≥ 90%
- **Domain Entities**: ≥ 95%
- **Critical Paths**: 100%

### Testing Patterns (Based on Core Service)

#### Mock Strategy

**IMPORTANT**: Follow this hierarchy for mocks:

1. **USE PACKAGE MOCKS FIRST** - Most shared packages provide mocks
2. **Generate with Mockery** - For your service's interfaces
3. **Create inline** - Only when needed for simple test-specific mocks

```bash
# Install mockery globally
go install github.com/vektra/mockery/v2@latest

# Generate mocks ONLY for YOUR service interfaces (not for packages!)
mockery --all --dir internal/application/ports --output internal/application/ports/mocks

# Or use go:generate directive in your interface files
//go:generate mockery --name=UserRepository --output=./mocks --filename=user_repository.go
```

#### 1. Using Package-Provided Mocks (PREFERRED)

Most shared packages already include mocks - USE THEM:

```go
import (
    "github.com/paradoxe35/mytorra/packages/cache"
    "github.com/paradoxe35/mytorra/packages/events"
    "github.com/paradoxe35/mytorra/packages/logger"
)

func TestServiceWithDependencies(t *testing.T) {
    // Use package-provided mocks - DON'T generate new ones!
    mockCache := cache.NewMockCache()
    mockEventPublisher := events.NewMockEventPublisher()
    testLogger := logger.New("test", "test")

    // Setup expectations
    mockCache.On("Get", mock.Anything, "key").Return([]byte("value"), nil)

    service := NewMyService(mockRepo, mockCache, mockEventPublisher, testLogger)
}
```

#### 2. Mock Creation Pattern for Service Interfaces

For YOUR service's interfaces, use mockery or create inline mocks:

```go
// internal/adapters/postgres/repositories/mocks.go
package repositories

import (
    "github.com/stretchr/testify/mock"
)

// Mock repository - define inline in test file
type mockUserRepository struct {
    mock.Mock
}

func (m *mockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*entities.User), args.Error(1)
}

func (m *mockUserRepository) Save(ctx context.Context, user *entities.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}
```

#### 3. Table-Driven Test Pattern

```go
func TestRegisterRequest_Validate(t *testing.T) {
    translator := setupTranslator(t)
    ctx := context.Background()

    tests := []struct {
        name    string
        request dtos.RegisterRequest
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid registration request",
            request: dtos.RegisterRequest{
                Email:           "test@example.com",
                Phone:           "+250780123456",
                Name:            "John Doe",
                Password:        "SecurePass123!@#",
                ConfirmPassword: "SecurePass123!@#",
                AcceptTerms:     true,
            },
            wantErr: false,
        },
        {
            name: "invalid email format",
            request: dtos.RegisterRequest{
                Email:           "invalid-email",
                // ... other fields
            },
            wantErr: true,
            errMsg:  "email",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.request.Validate(ctx, translator)

            if tt.wantErr {
                assert.Error(t, err)
                if tt.errMsg != "" {
                    assert.Contains(t, err.Error(), tt.errMsg)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

#### 4. Service Test Pattern

```go
func TestKeyManagementService_InitializeKeys(t *testing.T) {
    ctx := context.Background()
    mockRepo := new(mockJWTKeyRepository)
    mockLogger := logger.New("test", "test")

    service := NewKeyManagementService(mockRepo, mockLogger)

    t.Run("Active key exists", func(t *testing.T) {
        activeKey := &entities.JWTKey{
            ID:        uuid.New(),
            KeyID:     "test-key-id",
            IsActive:  true,
            ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
        }

        mockRepo.On("FindActiveKey", ctx).Return(activeKey, nil).Once()

        err := service.InitializeKeys(ctx)
        assert.NoError(t, err)
        mockRepo.AssertExpectations(t)
    })

    t.Run("No active key - creates new", func(t *testing.T) {
        mockRepo.On("FindActiveKey", ctx).Return(nil, errors.ErrNotFound).Once()
        mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.JWTKey")).Return(nil).Once()

        err := service.InitializeKeys(ctx)
        assert.NoError(t, err)
        mockRepo.AssertExpectations(t)
    })
}
```

#### 5. Repository Test Pattern

Repository tests use the `database.NewTestDB` helper for isolated testing:

```go
// internal/adapters/postgres/repositories/user_repository_test.go
package repositories

import (
    "github.com/paradoxe35/mytorra/packages/database"
    "github.com/paradoxe35/mytorra/services/core/internal/adapters/postgres/models"
)

// Setup helper creates isolated test environment
func setupUserRepo(t *testing.T) (*userRepository, *database.DB, func()) {
    // NewTestDB creates isolated schema, skips if ENABLE_DB_TESTS != "true"
    testDB := database.NewTestDB(t)
    ctx := context.Background()

    // Create required tables using Bun models
    err := testDB.CreateTables(ctx,
        (*models.UserModel)(nil),
        (*models.AreaModel)(nil))
    require.NoError(t, err)

    repo := NewUserRepository(testDB.DB)

    cleanup := func() {
        testDB.Close()
    }

    return repo.(*userRepository), testDB.DB, cleanup
}

// Test data builder
func createTestUser(t *testing.T, areaID uuid.UUID) *entities.User {
    email, err := valueobjects.NewEmail("test@example.com")
    require.NoError(t, err)

    phone, err := valueobjects.NewPhone("+1234567890")
    require.NoError(t, err)

    return &entities.User{
        ID:           uuid.New(),
        Email:        email,
        Phone:        phone,
        Name:         "Test User",
        PasswordHash: "hashed_password",
        AreaID:       areaID,
        Status:       entities.UserStatusActive,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }
}

func TestUserRepository_Save(t *testing.T) {
    repo, db, cleanup := setupUserRepo(t)
    defer cleanup()

    ctx := context.Background()

    // Create test area first (foreign key requirement)
    areaID := createTestAreaForUser(t, db)
    user := createTestUser(t, areaID)

    // Test Save
    err := repo.Save(ctx, user)
    assert.NoError(t, err)

    // Verify saved
    saved, err := repo.FindByID(ctx, user.ID)
    assert.NoError(t, err)
    assert.Equal(t, user.Email.String(), saved.Email.String())
}
```

**Note**: Repository tests automatically skip when `ENABLE_DB_TESTS != "true"`

#### 6. Handler Test Pattern (Often Skipped for Integration)

```go
// Note: Many handler tests are marked as t.Skip("Integration test - requires proper service mocks")
// This is acceptable when focusing on service layer testing

func TestAreaHandler_DetectArea(t *testing.T) {
    // For integration testing
    t.Skip("Integration test - requires proper service mocks")
}

// Or implement with inline mocks
func TestAreaHandler_GetActiveAreas(t *testing.T) {
    mockAreaService := new(MockAreaService)
    mockUserService := new(MockUserService)
    handler := NewAreaHandler(mockAreaService, mockUserService, logger.New("test", "test"))

    // Test implementation...
}
```

### Test Organization

#### File Structure

```
services/<service-name>/
├── internal/
│   ├── domain/entities/
│   │   ├── user.go
│   │   └── user_test.go              # Domain tests alongside code
│   ├── application/services/
│   │   ├── auth_service.go
│   │   └── auth_service_test.go      # Service tests with inline mocks
│   ├── adapters/http/handlers/
│   │   ├── user_handler.go
│   │   └── user_handler_test.go      # Handler tests (often integration)
│   └── adapters/postgres/repositories/
│       ├── user_repository.go
│       └── user_repository_test.go   # Repository tests
└── tests/                             # Integration/E2E tests
    └── integration/
```

### Key Testing Principles

1. **No External Mock Generation**: Create mocks inline using `testify/mock`
2. **Use Package Mocks**: Leverage mocks provided by shared packages
3. **Table-Driven Tests**: Use test tables for comprehensive coverage
4. **Test Alongside Code**: Keep test files next to implementation
5. **Focus on Business Logic**: Prioritize service and domain testing
6. **Skip When Appropriate**: Mark integration tests with `t.Skip()` when mocks are complex

### Testing Commands

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/application/services

# Run with verbose output
go test -v ./...

# Run only unit tests (skip integration)
go test -short ./...
```

### 2. Code Generation

```bash
# Generate mocks for testing
mockery --all --dir internal/application/ports --output internal/application/ports/mocks

# Generate Atlas migrations from Bun models
atlas migrate diff --env bun

# Run all go:generate directives
go generate ./...
```

## Service Registry

Each new service MUST:

1. Follow this architecture pattern
2. Use shared packages from `/packages`
3. Implement health checks
4. **Write tests alongside code (minimum 80% coverage)**
5. **Use inline mocks with testify/mock**
6. **Leverage package-provided mocks**
7. **Focus on service and domain layer testing**
8. Document API endpoints
9. Define clear domain boundaries
10. Handle errors consistently
11. Implement proper logging
12. Use dependency injection
13. Follow clean architecture principles
14. **Never merge code without passing tests**
