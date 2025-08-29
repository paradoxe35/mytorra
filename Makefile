# MyTorra Platform Makefile
# Main build and development automation

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Variables
GO := go
GOFLAGS := -v
SERVICES := core marketplace delivery booking torracab media realtime
PACKAGES := config logger security http database events errors cache
COVERAGE_THRESHOLD := 80

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m # No Color

#===================#
#    Development    #
#===================#

.PHONY: setup
setup: ## Initial project setup
	@echo "$(GREEN)Setting up project...$(NC)"
	$(GO) work sync
	@echo "$(GREEN)Installing tools...$(NC)"
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	$(GO) install github.com/vektra/mockery/v2@latest
	$(GO) install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)Setup complete!$(NC)"

.PHONY: deps
deps: ## Download all dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@for pkg in $(PACKAGES); do \
		echo "Installing deps for package/$$pkg..."; \
		cd packages/$$pkg && $(GO) mod download && cd ../..; \
	done
	@for svc in $(SERVICES); do \
		if [ -d "services/$$svc" ]; then \
			echo "Installing deps for service/$$svc..."; \
			cd services/$$svc && $(GO) mod download && cd ../..; \
		fi \
	done

.PHONY: tidy
tidy: ## Run go mod tidy on all modules
	@echo "$(GREEN)Tidying modules...$(NC)"
	@for pkg in $(PACKAGES); do \
		echo "Tidying package/$$pkg..."; \
		cd packages/$$pkg && $(GO) mod tidy && cd ../..; \
	done
	@for svc in $(SERVICES); do \
		if [ -d "services/$$svc" ]; then \
			echo "Tidying service/$$svc..."; \
			cd services/$$svc && $(GO) mod tidy && cd ../..; \
		fi \
	done

#===================#
#   Code Quality    #
#===================#

.PHONY: fmt
fmt: ## Format all Go code
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GO) fmt ./...
	@for pkg in $(PACKAGES); do \
		cd packages/$$pkg && $(GO) fmt ./... && cd ../..; \
	done
	@for svc in $(SERVICES); do \
		if [ -d "services/$$svc" ]; then \
			cd services/$$svc && $(GO) fmt ./... && cd ../..; \
		fi \
	done

.PHONY: fmt-check
fmt-check: ## Check if code is formatted
	@echo "$(GREEN)Checking code formatting...$(NC)"
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "$(RED)Code is not formatted. Run 'make fmt'$(NC)"; \
		gofmt -l .; \
		exit 1; \
	fi
	@echo "$(GREEN)All code is properly formatted$(NC)"

.PHONY: lint
lint: ## Run linters on all code
	@echo "$(GREEN)Running linters...$(NC)"
	@if ! which golangci-lint > /dev/null; then \
		echo "$(RED)golangci-lint not installed. Run 'make setup'$(NC)"; \
		exit 1; \
	fi
	@for pkg in $(PACKAGES); do \
		echo "Linting package/$$pkg..."; \
		cd packages/$$pkg && golangci-lint run --timeout=5m && cd ../..; \
	done
	@for svc in $(SERVICES); do \
		if [ -d "services/$$svc" ]; then \
			echo "Linting service/$$svc..."; \
			cd services/$$svc && golangci-lint run --timeout=5m && cd ../..; \
		fi \
	done

.PHONY: vet
vet: ## Run go vet on all code
	@echo "$(GREEN)Running go vet...$(NC)"
	$(GO) vet ./...
	@for pkg in $(PACKAGES); do \
		cd packages/$$pkg && $(GO) vet ./... && cd ../..; \
	done
	@for svc in $(SERVICES); do \
		if [ -d "services/$$svc" ]; then \
			cd services/$$svc && $(GO) vet ./... && cd ../..; \
		fi \
	done

#===================#
#     Testing       #
#===================#

.PHONY: test
test: ## Run all tests
	@echo "$(GREEN)Running tests...$(NC)"
	@for pkg in $(PACKAGES); do \
		echo "Testing package/$$pkg..."; \
		cd packages/$$pkg && $(GO) test ./... $(GOFLAGS) && cd ../..; \
	done
	@for svc in $(SERVICES); do \
		if [ -d "services/$$svc" ]; then \
			echo "Testing service/$$svc..."; \
			cd services/$$svc && $(GO) test ./... $(GOFLAGS) && cd ../..; \
		fi \
	done

.PHONY: test-short
test-short: ## Run short tests only
	@echo "$(GREEN)Running short tests...$(NC)"
	$(GO) test -short ./...

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo "$(GREEN)Running tests with race detector...$(NC)"
	@for pkg in $(PACKAGES); do \
		echo "Testing package/$$pkg with race detector..."; \
		cd packages/$$pkg && $(GO) test -race ./... && cd ../..; \
	done
	@for svc in $(SERVICES); do \
		if [ -d "services/$$svc" ]; then \
			echo "Testing service/$$svc with race detector..."; \
			cd services/$$svc && $(GO) test -race ./... && cd ../..; \
		fi \
	done

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@mkdir -p coverage
	@for pkg in $(PACKAGES); do \
		echo "Testing package/$$pkg with coverage..."; \
		cd packages/$$pkg && $(GO) test -coverprofile=../../coverage/$$pkg.out ./... && cd ../..; \
	done
	@for svc in $(SERVICES); do \
		if [ -d "services/$$svc" ]; then \
			echo "Testing service/$$svc with coverage..."; \
			cd services/$$svc && $(GO) test -coverprofile=../../coverage/$$svc.out ./... && cd ../..; \
		fi \
	done
	@echo "$(GREEN)Coverage reports generated in coverage/$(NC)"

.PHONY: coverage-html
coverage-html: test-coverage ## Generate HTML coverage report
	@echo "$(GREEN)Generating HTML coverage report...$(NC)"
	@for file in coverage/*.out; do \
		if [ -f "$$file" ]; then \
			$(GO) tool cover -html=$$file -o $${file%.out}.html; \
		fi \
	done
	@echo "$(GREEN)HTML coverage reports generated in coverage/$(NC)"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(GREEN)Running integration tests...$(NC)"
	$(GO) test -tags=integration ./tests/integration/...

.PHONY: test-service
test-service: ## Test specific service (use SERVICE=core)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)Please specify SERVICE=<service_name>$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Testing service/$(SERVICE)...$(NC)"
	cd services/$(SERVICE) && $(GO) test ./... $(GOFLAGS)

.PHONY: test-package
test-package: ## Test specific package (use PACKAGE=config)
	@if [ -z "$(PACKAGE)" ]; then \
		echo "$(RED)Please specify PACKAGE=<package_name>$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Testing package/$(PACKAGE)...$(NC)"
	cd packages/$(PACKAGE) && $(GO) test ./... $(GOFLAGS)

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "$(GREEN)Running benchmarks...$(NC)"
	$(GO) test -bench=. -benchmem ./...

#===================#
#      Build        #
#===================#

.PHONY: build
build: ## Build all services
	@echo "$(GREEN)Building services...$(NC)"
	@mkdir -p bin
	@for svc in $(SERVICES); do \
		if [ -d "services/$$svc/cmd/api" ]; then \
			echo "Building $$svc API..."; \
			cd services/$$svc && $(GO) build -o ../../bin/$$svc-api ./cmd/api && cd ../..; \
		fi; \
		if [ -d "services/$$svc/cmd/worker" ]; then \
			echo "Building $$svc worker..."; \
			cd services/$$svc && $(GO) build -o ../../bin/$$svc-worker ./cmd/worker && cd ../..; \
		fi \
	done

.PHONY: build-service
build-service: ## Build specific service (use SERVICE=core)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)Please specify SERVICE=<service_name>$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Building service/$(SERVICE)...$(NC)"
	@mkdir -p bin
	@if [ -d "services/$(SERVICE)/cmd/api" ]; then \
		cd services/$(SERVICE) && $(GO) build -o ../../bin/$(SERVICE)-api ./cmd/api; \
	fi
	@if [ -d "services/$(SERVICE)/cmd/worker" ]; then \
		cd services/$(SERVICE) && $(GO) build -o ../../bin/$(SERVICE)-worker ./cmd/worker; \
	fi

.PHONY: install
install: ## Install all service binaries
	@echo "$(GREEN)Installing binaries...$(NC)"
	@for svc in $(SERVICES); do \
		if [ -d "services/$$svc/cmd/api" ]; then \
			cd services/$$svc && $(GO) install ./cmd/api && cd ../..; \
		fi; \
		if [ -d "services/$$svc/cmd/worker" ]; then \
			cd services/$$svc && $(GO) install ./cmd/worker && cd ../..; \
		fi \
	done

#===================#
#     Database      #
#===================#

.PHONY: migrate-up
migrate-up: ## Run database migrations up
	@echo "$(GREEN)Running migrations up...$(NC)"
	@for svc in $(SERVICES); do \
		if [ -d "services/$$svc/migrations" ]; then \
			echo "Migrating $$svc..."; \
			migrate -path services/$$svc/migrations -database "$${DATABASE_URL}" up; \
		fi \
	done

.PHONY: migrate-down
migrate-down: ## Run database migrations down
	@echo "$(YELLOW)Running migrations down...$(NC)"
	@for svc in $(SERVICES); do \
		if [ -d "services/$$svc/migrations" ]; then \
			echo "Rolling back $$svc..."; \
			migrate -path services/$$svc/migrations -database "$${DATABASE_URL}" down 1; \
		fi \
	done

.PHONY: migrate-create
migrate-create: ## Create new migration (use SERVICE=core NAME=create_users)
	@if [ -z "$(SERVICE)" ] || [ -z "$(NAME)" ]; then \
		echo "$(RED)Please specify SERVICE=<service_name> NAME=<migration_name>$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Creating migration for $(SERVICE): $(NAME)$(NC)"
	@mkdir -p services/$(SERVICE)/migrations
	migrate create -ext sql -dir services/$(SERVICE)/migrations -seq $(NAME)

#===================#
#      Docker       #
#===================#

.PHONY: docker-build
docker-build: ## Build Docker images for all services
	@echo "$(GREEN)Building Docker images...$(NC)"
	@for svc in $(SERVICES); do \
		if [ -f "services/$$svc/Dockerfile.api" ]; then \
			echo "Building $$svc API image..."; \
			docker build -f services/$$svc/Dockerfile.api -t mytorra/$$svc-api:latest services/$$svc; \
		fi; \
		if [ -f "services/$$svc/Dockerfile.worker" ]; then \
			echo "Building $$svc worker image..."; \
			docker build -f services/$$svc/Dockerfile.worker -t mytorra/$$svc-worker:latest services/$$svc; \
		fi \
	done

.PHONY: docker-up
docker-up: ## Start all services with docker-compose
	docker-compose up -d

.PHONY: docker-down
docker-down: ## Stop all services
	docker-compose down

.PHONY: docker-logs
docker-logs: ## Show logs from all containers
	docker-compose logs -f

#===================#
#   Documentation   #
#===================#

.PHONY: docs
docs: ## Generate API documentation
	@echo "$(GREEN)Generating API documentation...$(NC)"
	@for svc in $(SERVICES); do \
		if [ -d "services/$$svc/cmd/api" ]; then \
			cd services/$$svc && swag init -g cmd/api/main.go -o docs && cd ../..; \
		fi \
	done

#===================#
#     Security      #
#===================#

.PHONY: security
security: ## Run security checks
	@echo "$(GREEN)Running security checks...$(NC)"
	$(GO) list -json -deps ./... | nancy sleuth
	gosec ./...

.PHONY: audit
audit: ## Audit dependencies for vulnerabilities
	@echo "$(GREEN)Auditing dependencies...$(NC)"
	$(GO) list -json -m all | nancy sleuth

#===================#
#    Validation     #
#===================#

.PHONY: validate
validate: fmt-check lint vet test ## Run all validation checks
	@echo "$(GREEN)================================$(NC)"
	@echo "$(GREEN)All validation checks passed! ✓$(NC)"
	@echo "$(GREEN)================================$(NC)"

.PHONY: pre-commit
pre-commit: fmt lint test ## Run pre-commit checks
	@echo "$(GREEN)Pre-commit checks passed! ✓$(NC)"

.PHONY: ci
ci: deps validate build ## Run CI pipeline locally
	@echo "$(GREEN)CI pipeline completed successfully! ✓$(NC)"

#===================#
#     Utilities     #
#===================#

.PHONY: clean
clean: ## Clean build artifacts and temp files
	@echo "$(GREEN)Cleaning...$(NC)"
	rm -rf bin/
	rm -rf coverage/
	rm -rf vendor/
	rm -rf dist/
	find . -name "*.test" -delete
	find . -name "*.out" -delete
	@echo "$(GREEN)Clean complete!$(NC)"

.PHONY: mocks
mocks: ## Generate mocks for testing
	@echo "$(GREEN)Generating mocks...$(NC)"
	@for svc in $(SERVICES); do \
		if [ -d "services/$$svc" ]; then \
			cd services/$$svc && mockery --all && cd ../..; \
		fi \
	done

.PHONY: proto
proto: ## Generate protobuf files (if using gRPC)
	@echo "$(GREEN)Generating protobuf files...$(NC)"
	@for file in $$(find . -name "*.proto"); do \
		protoc --go_out=. --go-grpc_out=. $$file; \
	done

.PHONY: run
run: ## Run specific service locally (use SERVICE=core)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)Please specify SERVICE=<service_name>$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Running service/$(SERVICE)...$(NC)"
	cd services/$(SERVICE) && $(GO) run ./cmd/api

.PHONY: watch
watch: ## Watch for changes and rebuild (requires entr)
	@echo "$(GREEN)Watching for changes...$(NC)"
	find . -name "*.go" | entr -r make build

.PHONY: version
version: ## Show Go version and dependencies
	@echo "$(GREEN)Go version:$(NC)"
	$(GO) version
	@echo ""
	@echo "$(GREEN)Module dependencies:$(NC)"
	$(GO) list -m all

# Default target
.DEFAULT_GOAL := help