# MyTorra Platform

A comprehensive multi-service platform for Bukavu and Goma, creating a localized digital economy that connects buyers, sellers, and service providers.

## 🚀 Quick Start

### Prerequisites

- Go 1.25 or later
- Docker & Docker Compose
- Make
- PostgreSQL, Redis, RabbitMQ (or use Docker Compose)

### Setup

```bash
# Clone the repository
git clone https://github.com/paradoxe35/torra.git
cd torra

# Initial setup
make setup

# Install dependencies
make deps

# Run tests
make test

# Start development environment
docker-compose up -d
```

## 📁 Project Structure

```
torra/
├── services/          # Microservices
│   ├── core/         # User auth, payments, subscriptions
│   ├── marketplace/  # MyTorra & TorraShop
│   ├── delivery/     # TorraFood & TorraDelivery
│   ├── booking/      # TorraBooking & TorraFlight
│   └── torracab/     # Ride-hailing service
├── packages/         # Shared packages
│   ├── config/       # Configuration management
│   ├── logger/       # Structured logging
│   ├── security/     # Security utilities
│   ├── database/     # Database utilities
│   └── ...
├── apps/            # Frontend applications
│   └── mobile/      # Flutter mobile app
├── infra/           # Infrastructure configs
├── docs/            # Documentation
├── specs/           # Specifications
└── tests/           # Integration & E2E tests
```

## 🛠️ Development

### Available Commands

```bash
# Code quality
make fmt           # Format code
make lint          # Run linters
make vet           # Run go vet

# Testing
make test          # Run all tests
make test-coverage # Run tests with coverage
make test-race     # Run tests with race detector

# Building
make build         # Build all services
make build-service SERVICE=core  # Build specific service

# Development
make run SERVICE=core  # Run specific service
make watch            # Watch for changes

# Validation
make validate      # Run all checks (fmt, lint, test)
```

### Workflow

1. Update `project.todo` with your tasks
2. Write tests first (TDD)
3. Implement features
4. Run `make validate` before committing
5. Update documentation

## 📚 Documentation

- [Service Guidelines](specs/SERVICE_GUIDELINES.md) - Development standards
- [Architecture Plan](specs/PLAN.md) - Technical architecture
- [Project Overview](specs/PROJECT.md) - Business requirements
- [API Documentation](docs/api/) - OpenAPI specifications

## 🏗️ Architecture

- **Monorepo** with Go workspace
- **Clean Architecture** for all services
- **PostgreSQL** + **PostGIS** for data
- **Redis** for caching and sessions
- **RabbitMQ** for async messaging
- **Centrifugo** for real-time features
- **Meilisearch** for search
- **Asynq** for background jobs

## 🔒 Security

- SQL injection prevention
- XSS protection
- CSRF tokens
- Rate limiting
- Input validation
- Secure password hashing (bcrypt)
- JWT authentication
- Audit logging

## 🧪 Testing

- Unit tests (minimum 80% coverage)
- Integration tests with Testcontainers
- Contract tests with Pact
- Load tests with K6
- Security tests (OWASP)

## 🚀 Deployment

Services are containerized and can be deployed to:
- Kubernetes clusters
- Docker Swarm
- Cloud Run / ECS
- Traditional VMs

## 📊 Monitoring

- Prometheus metrics
- Grafana dashboards
- Distributed tracing with Jaeger
- Centralized logging with ELK

## 🤝 Contributing

1. Check `project.todo` for available tasks
2. Follow [Service Guidelines](specs/SERVICE_GUIDELINES.md)
3. Write tests for your changes
4. Run `make validate` before pushing
5. Create detailed pull requests

## 📄 License

Proprietary - MyTorra Platform © 2024

## 📞 Support

For questions and support, contact the development team.
