# Shared Packages

This directory contains shared code that is used across multiple services in the MyTorra platform. These packages promote code reuse and maintain consistency across services.

## Package Structure

- **auth/** - JWT generation/validation, authentication helpers
- **cache/** - Redis cache abstraction and utilities
- **config/** - Configuration loading with Viper
- **database/** - Database connection utilities, common queries
- **errors/** - Common error types and handling
- **events/** - Event definitions for RabbitMQ
- **http/** - HTTP middleware, response helpers
- **logger/** - Structured logging setup with slog
- **security/** - Input validation, sanitization, password hashing
- **types/** - Shared data types and DTOs
- **utils/** - General utility functions

## Usage

Import packages in your service:

```go
import (
    "github.com/paradoxe35/torra/packages/config"
    "github.com/paradoxe35/torra/packages/logger"
    "github.com/paradoxe35/torra/packages/security"
)
```

## Guidelines

1. Packages should have zero or minimal external dependencies
2. Each package must have comprehensive tests
3. Packages should be documented with examples
4. Breaking changes require version bumping
5. Packages should be stateless where possible
