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
├── migrations/       # Database migrations (Atlas)
├── tests/           # Integration tests
├── Dockerfile.api
├── Dockerfile.worker
├── atlas.hcl        # Atlas DB schema config
├── go.mod
└── README.md
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
│   └── validators/  # Request validators
├── postgres/
│   ├── models/      # Database models
│   ├── repositories/# Repository implementations
│   └── types/       # Custom DB types
├── payment/         # Payment gateway adapters
├── oauth/          # OAuth providers
└── grpc/           # gRPC clients/servers
```

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

Reusable packages across all services:

```
packages/
├── cache/          # Redis cache adapter
├── config/         # Configuration loader
├── database/       # Database connection (Bun + pgx)
├── errors/         # Common error types
├── events/         # Event publishing (RabbitMQ)
├── http/           # HTTP utilities
├── i18n/           # Internationalization
├── logger/         # Structured logging
└── security/       # Security utilities
```

### Package Usage Example

```go
import (
    "github.com/paradoxe35/mytorra/packages/cache"
    "github.com/paradoxe35/mytorra/packages/config"
    "github.com/paradoxe35/mytorra/packages/database"
    "github.com/paradoxe35/mytorra/packages/logger"
)

// Load configuration
cfg, err := config.Load("configs/config.yaml")

// Initialize database
db, err := database.New(&cfg.Database)
defer db.Close()

// Initialize cache
cacheClient, err := cache.NewRedisCacheAdapter(&cfg.Redis, "service-name")
```

## Database Management

### Atlas Schema Management

Uses Atlas for declarative schema management with Bun models.

```hcl
# atlas.hcl
data "external_schema" "bun" {
  program = [
    "go", "run", "-mod=mod",
    "ariga.io/atlas-provider-bun",
    "load",
    "--path", "./internal/adapters/postgres/models",
    "--dialect", "postgres"
  ]
}

env "bun" {
  src = data.external_schema.bun.url
  dev = "docker://postgis/17-3.5-alpine/dev"
  migration {
    dir = "file://migrations"
  }
}
```

### Model Definition

```go
// internal/adapters/postgres/models/user_model.go
package models

type UserModel struct {
    bun.BaseModel `bun:"table:users,alias:u"`

    ID              uuid.UUID  `bun:"id,pk,type:uuid"`
    Email           string     `bun:"email,unique,notnull"`
    Phone           string     `bun:"phone,unique"`
    PasswordHash    string     `bun:"password_hash,notnull"`
    CreatedAt       time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp"`
    UpdatedAt       time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
}

func (m *UserModel) ToEntity() (*entities.User, error) {
    // Convert to domain entity
}

func UserFromEntity(u *entities.User) *UserModel {
    // Convert from domain entity
}
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

### Chi Router Setup

```go
// internal/adapters/http/router.go
package http

func NewRouter(services *ServiceContainer) *chi.Mux {
    r := chi.NewRouter()

    // Middleware
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(60 * time.Second))

    // Routes
    r.Route("/api/v1", func(r chi.Router) {
        // Auth routes
        r.Route("/auth", func(r chi.Router) {
            r.Post("/login", handlers.Login(services.AuthService))
            r.Post("/register", handlers.Register(services.UserService))
            r.Post("/refresh", handlers.Refresh(services.AuthService))
        })

        // Protected routes
        r.Group(func(r chi.Router) {
            r.Use(middleware.Authenticate(services.JWTService))

            r.Route("/users", func(r chi.Router) {
                r.Get("/profile", handlers.GetProfile(services.UserService))
                r.Put("/profile", handlers.UpdateProfile(services.UserService))
            })
        })
    })

    return r
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

### Event Publishing

```go
// internal/application/services/user_service.go
func (s *UserService) CreateUser(ctx context.Context, req dtos.CreateUserRequest) (*entities.User, error) {
    user := &entities.User{...}

    if err := s.userRepo.Save(ctx, user); err != nil {
        return nil, err
    }

    // Publish event
    event := events.UserCreated{
        UserID:    user.ID,
        Email:     user.Email.String(),
        CreatedAt: user.CreatedAt,
    }

    if err := s.eventPublisher.Publish(ctx, "user.created", event); err != nil {
        s.logger.Error("Failed to publish event", err)
    }

    return user, nil
}
```

## Testing Strategy

### Unit Tests

```go
// internal/application/services/auth_service_test.go
func TestAuthService_Login(t *testing.T) {
    // Setup
    userRepo := mocks.NewMockUserRepository(t)
    cacheRepo := mocks.NewMockCache(t)

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

### Integration Tests

```go
// tests/integration/auth_test.go
func TestAuthFlow(t *testing.T) {
    // Setup test database
    db := testutil.SetupTestDB(t)
    defer db.Close()

    // Setup services
    app := setupTestApp(db)

    // Test registration
    regReq := dtos.RegisterRequest{...}
    user, err := app.UserService.Register(context.Background(), regReq)
    require.NoError(t, err)

    // Test login
    loginReq := dtos.LoginRequest{...}
    auth, err := app.AuthService.Login(context.Background(), loginReq)
    require.NoError(t, err)
    require.NotEmpty(t, auth.AccessToken)
}
```

## Service Configuration

### Config Structure

```yaml
# configs/config.yaml
service:
  name: core-service
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

#### 1. Mock Creation Pattern

Create mocks inline using `stretchr/testify/mock` - NO external mock generation tools:

```go
// internal/adapters/postgres/repositories/user_repository_test.go
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

#### 2. Use Shared Package Mocks

Packages provide their own mocks - use them:

```go
import (
    "github.com/paradoxe35/mytorra/packages/cache"
    "github.com/paradoxe35/mytorra/packages/events"
    "github.com/paradoxe35/mytorra/packages/logger"
)

func TestServiceWithDependencies(t *testing.T) {
    // Use package-provided mocks
    mockCache := cache.NewMockCache()
    mockEventPublisher := events.NewMockEventPublisher()
    testLogger := logger.New("test", "test")

    // Setup expectations
    mockCache.On("Get", mock.Anything, "key").Return([]byte("value"), nil)

    service := NewMyService(mockRepo, mockCache, mockEventPublisher, testLogger)
    // Test service...
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

## Development Workflow

### 1. Local Development

```bash
# Run database
docker-compose up -d postgres redis rabbitmq

# Run migrations
atlas migrate apply --env bun

# Run tests first
go test ./...

# Run service
go run cmd/api/main.go

# Run worker
go run cmd/worker/main.go
```

### 2. Code Generation

```bash
# Generate Atlas migrations
atlas migrate diff --env bun
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
