# Core Service

The Core Service is the foundational service of the MyTorra platform, managing:
- User Accounts & Authentication
- Payment Processing
- Notifications
- Subscription Engine

## Architecture

The service follows Clean Architecture principles with the following layers:
- **Domain**: Pure business logic and entities
- **Application**: Use cases and orchestration
- **Adapters**: External world implementations (HTTP, Database, etc.)

## Quick Start

### Prerequisites
- Go 1.25+
- PostgreSQL 15+
- Redis 7+
- RabbitMQ 3.12+

### Installation

```bash
# Install dependencies
go mod download

# Run database migrations
migrate -path migrations -database "postgresql://user:pass@localhost/mytorra_core?sslmode=disable" up

# Run the service
go run cmd/api/main.go
```

### Configuration

Configuration is managed via `configs/config.yaml`. Environment variables can override config values using the `MYTORRA_` prefix.

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...
```

## API Documentation

API documentation is available at `/api/docs` when the service is running.

## Project Structure

```
services/core/
├── cmd/
│   ├── api/           # HTTP API server
│   └── worker/        # Background worker
├── internal/
│   ├── domain/        # Business logic
│   ├── application/   # Use cases
│   └── adapters/      # External implementations
├── migrations/        # Database migrations
└── configs/           # Configuration files
```
