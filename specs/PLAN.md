Of course. Here is the final, definitive `PLAN2.md`, which incorporates all previous suggestions and seamlessly integrates the robust background job processing strategy using `Asynq`.

---

# MyTorra Platform: Definitive Production Architecture Plan (v5.0)

## 1. Guiding Principles

This architecture is the definitive blueprint for the MyTorra platform. It is designed to be robust, secure, scalable, and maintainable, enabling rapid feature development while ensuring long-term stability.

- **Pragmatic & Cost-Effective:** Prioritize managed services and serverless components to minimize operational overhead, scaling costs with revenue.
- **Developer Velocity & Safety:** Empower the team to ship features rapidly and safely through comprehensive automation, testing, and a clear monorepo structure.
- **Security by Design:** Integrate security patterns at every layer of the architecture, from authentication and secrets management to secure coding practices.
- **Observability-Driven:** Instrument every component from day one to provide deep, actionable insights into system health, performance, and user behavior.
- **Loosely Coupled & Highly Cohesive:** Both inter-service (macro) and intra-service (micro) architectures are designed to be loosely coupled, ensuring changes in one part of the system have minimal impact on others.

## 2. Definitive Monorepo Project Structure

The platform will be developed in a single monorepo to streamline dependency management and ensure code consistency. A Go Workspace (`go.work`) will manage the relationships between all Go modules.

```plaintext
mytorra-platform/
├── .gitignore
├── README.md
├── go.work
├── turbo.json          # Monorepo orchestrator configuration
│
├── services/
│   └── core/
│       ├── cmd/
│       │   ├── api/
│       │   │   └── main.go     # The main entry point for the HTTP API server
│       │   └── worker/
│       │       └── main.go     # A SEPARATE entry point for the background worker process
│       │
│       ├── internal/
│       │   ├── domain/
│       │   ├── application/
│       │   └── adapters/       # e.g., http, postgres, worker handlers
│       ├── migrations/
│       ├── Dockerfile.api      # Dockerfile for the API process
│       ├── Dockerfile.worker   # Dockerfile for the Worker process
│       └── go.mod
│
├── packages/
│   ├── auth/
│   ├── cache/
│   ├── config/
│   ├── dtos/
│   ├── events/
│   ├── features/
│   ├── i18n/
│   └── logger/
│
├── apps/
│   └── mobile/
│       ├── lib/
│       └── pubspec.yaml
│
├── infra/
│   ├── terraform/
│   ├── kubernetes/
│   └── message-bus/
│
├── docs/
│   └── api/
│       ├── core.v1.yaml
│       └── marketplace.v1.yaml
│
└── tests/
    ├── load/
    └── contract/
```

## 3. Bundled Services Definition & Responsibilities

| Service Bundle    | Description & Core Responsibilities                                                                                                                                                                                                                                              |
| :---------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`core`**        | **Foundational.** The central nervous system. Manages User Accounts, Authentication (JWT, MFA, OAuth2), Payments (via an abstracted gateway), Notifications (as a generic sender), and the generic Subscription engine. It is the source of truth for user identity and billing. |
| **`marketplace`** | **E-Commerce.** The heart of buying and selling. Manages MyTorra (used items), TorraShop (new items), the Universal Want/Request System, and interfaces with the search engine. It defines and enforces the business logic for seller subscriptions.                             |
| **`delivery`**    | **Logistics.** Manages TorraFood (restaurants, menus) and TorraDelivery (packages). It is responsible for a shared pool of delivery agents, route optimization, and order fulfillment. It defines and enforces restaurant subscription rules.                                    |
| **`booking`**     | **Reservations.** Manages TorraBooking (hotels, etc.) and TorraFlight. Responsible for availability, scheduling, and reservation lifecycle.                                                                                                                                      |
| **`torracab`**    | **Ride-Hailing.** A standalone service optimized for high-frequency, real-time operations. Manages drivers, vehicles, ride requests, fare calculation, and pushes live location data to the `realtime` service.                                                                  |
| **`media`**       | **Media Pipeline.** A dedicated utility service that handles the full lifecycle of user-generated content: secure uploads, processing (resizing, watermarking), moderation (NSFW checks), and delivery via a CDN.                                                                |
| **`realtime`**    | **Real-Time Layer.** A specialized service running Centrifugo. Its sole job is to manage persistent WebSocket connections and broadcast messages in real-time. It receives commands to publish messages from other backend services.                                             |

## 4. Internal Service Architecture: A Modular Clean Architecture

This architecture is **mandatory for every service** (especially the multi-domain `core` service) to ensure loose coupling, extensibility, and testability.

- **Principle:** Business logic is organized into modular, independent vertical slices. Communication between these internal modules happens exclusively through well-defined interfaces (ports).
- **Layers within each Service:**
  1.  **Domain (`internal/domain/`):** Pure Go structs and business logic, with no external dependencies.
  2.  **Application (`internal/application/`):** Orchestrates the business logic. Defines interfaces for external dependencies in a `ports` subdirectory.
  3.  **Adapters (`internal/adapters/`):** Implements the interfaces defined by the Application layer. This is where the outside world (HTTP, PostgreSQL, RabbitMQ) connects to the business logic.
  4.  **Wiring (`cmd/`):** Initializes and injects concrete adapter implementations into the application services via their interfaces within the `api` and `worker` entry points.

This structure is the key to building a professional, maintainable system that can evolve safely.

## 5. Technology Stack, Tools, & Go Best Practices (Expanded & Detailed)

This section prescribes the specific tools, libraries, and enforceable best practices that will ensure the MyTorra platform is built to a consistent, high-quality, and professional standard.

| Category                  | Technology / Tool                              | Rationale & Implementation Details                                                                                                                                                                                                                                                                                                                             |
| :------------------------ | :--------------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Monorepo Orchestrator** | **Turborepo**                                  | **Why:** Simplifies running tasks (build, test, lint) across the monorepo. It caches outputs and only runs tasks on services that have changed, dramatically speeding up local development and CI pipelines. **Implementation:** A `turbo.json` file at the root will define the task pipeline.                                                                |
| **Web Framework**         | **Chi**                                        | **Why:** Robust, idiomatic, and fully compatible with the Go `net/http` ecosystem. Recommended over Fiber for its standard library compliance and broader middleware support, which is critical for observability and security integrations.                                                                                                                   |
| **Database**              | **Managed PostgreSQL** + **PostGIS**           | **Why:** Reliable primary data store with powerful geospatial capabilities. Using a managed service (Cloud SQL / RDS) offloads complex operational tasks like backups, patching, and replication.                                                                                                                                                              |
| **Database ORM/Lib**      | **sqlx** (`github.com/jmoiron/sqlx`)           | **Why:** Provides powerful, non-intrusive extensions to the standard `database/sql` package without the overhead of a full ORM. It simplifies scanning query results into structs and handling named parameters. **Implementation:** `sqlx` will be used within the persistence adapters (e.g., `internal/adapters/postgres/`) to execute the handwritten SQL. |
| **DB Migrations**         | **golang-migrate**                             | **Why:** The de-facto standard for managing database schema changes through versioned SQL files. **Implementation:** Each service requiring a database will have a `migrations/` directory. Migrations will be applied automatically as part of the CI/CD deployment process.                                                                                  |
| **Advanced Search**       | **Meilisearch**                                | **Why:** Modern, fast, and easy-to-operate search engine, ideal for a startup. Its simplicity significantly reduces operational overhead compared to Elasticsearch.                                                                                                                                                                                            |
| **Real-Time Layer**       | **Centrifugo** + **Redis Pub/Sub**             | **Why:** A dedicated, high-performance, open-source real-time messaging server. Using a specialized tool is more scalable and robust than implementing complex WebSocket logic in each business service.                                                                                                                                                       |
| **Message Bus (Events)**  | **RabbitMQ** (Managed)                         | **Why:** A flexible and mature message broker for asynchronous, fire-and-forget, inter-service communication. Supports dead-letter queues and other reliability patterns out of the box.                                                                                                                                                                       |
| **Background Jobs**       | **Asynq** (`github.com/hibiken/asynq`)         | **Why:** A robust and simple Redis-backed library for deferred and guaranteed task execution. It provides automatic retries, scheduled/cron jobs, prioritization, and a critical monitoring UI (`Asynqmon`), avoiding the need for a separate job queue database.                                                                                              |
| **Caching/Sessions**      | **Managed Redis**                              | **Why:** High-performance in-memory store for caching, session management, and as the broker for `Asynq`.                                                                                                                                                                                                                                                      |
| **Secrets Mgmt**          | **HashiCorp Vault** or **AWS Secrets Manager** | **Why:** Secure, auditable, and centralized solution for all application secrets. **Implementation:** Services will fetch secrets at startup using the cloud provider's SDK, authorized by their runtime IAM role. No secrets will ever be stored in git.                                                                                                      |

### Go Best Practices, Linters, and Enforced Standards

This is a mandatory standard for all Go code committed to the repository.

| Practice                  | Tool / Library                                                         | Implementation & Enforcement                                                                                                                                                                                                                                                                          |
| :------------------------ | :--------------------------------------------------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Code Formatting**       | **gofmt**                                                              | **Enforcement:** A git pre-commit hook will run `gofmt -w` on all staged `.go` files. The CI pipeline will fail if any unformatted code is committed.                                                                                                                                                 |
| **Linting**               | **golangci-lint**                                                      | **Why:** A fast, powerful linter aggregator. **Enforcement:** A `.golangci.yml` configuration file at the root of the monorepo will enable a strict set of linters. The CI pipeline will run `golangci-lint run ./...` on every commit and fail the build on any error.                               |
| **Structured Logging**    | **slog** (Go 1.25+ Standard Library)                                   | **Why:** The new standard for structured logging in Go. **Implementation:** The shared `packages/logger/` will provide a configured `slog.Logger` instance that outputs JSON and is injected into application contexts to ensure all log lines include contextual information like `traceID`.         |
| **Configuration**         | **Viper** (`github.com/spf13/viper`)                                   | **Why:** A complete configuration solution. **Implementation:** It will be configured to read from a config file, then override with environment variables.                                                                                                                                           |
| **Dependency Injection**  | **Manual Injection** / **Wire** (`github.com/google/wire`)             | **Why:** Ensures loose coupling and testability. **Implementation:** For most services, manual dependency injection in `cmd/` is sufficient. For services with highly complex dependency graphs, `wire` can be used to automate wiring code generation.                                               |
| **Error Handling**        | **Go 1.25+ `errors` package**                                          | **Why:** Native support for error wrapping. **Best Practice:** Errors from external dependencies will be wrapped with context (e.g., `return fmt.Errorf("userRepo.FindByID: %w", err)`) to provide a clear, traceable error path.                                                                     |
| **API & Data Validation** | **Go Playground Validator** (`github.com/go-playground/validator/v10`) | **Why:** The most popular library for struct validation based on tags. **Implementation:** It will be used in the `adapters/http/` layer to validate all incoming DTOs.                                                                                                                               |
| **Internationalization**  | **`golang.org/x/text`**                                                | **Why:** The official, high-performance library for localization. **Implementation:** The shared `packages/i18n/` will provide helpers to load translation files. An HTTP middleware will inspect `Accept-Language` headers and the JWT to set the correct language printer in the request `context`. |
| **Context Propagation**   | **`context.Context`**                                                  | **Why:** Critical for handling timeouts, cancellations, and passing request-scoped data. **Mandatory Rule:** `context.Context` must be the first parameter for any function that performs I/O or makes a network call.                                                                                |
| **API Schema Generation** | **oapi-codegen**                                                       | **Why:** Enforces a design-first, contract-first approach. **Workflow:** Developers will define API changes in `docs/api/*.yaml`. The CI pipeline will run `oapi-codegen` to automatically generate Go structs and server interfaces.                                                                 |
| **Mock Generation**       | **testify/mock** or **mockery** (`github.com/vektra/mockery`)          | **Why:** Automates the creation of mock implementations for interfaces. **Implementation:** `mockery` will be run via `go generate` commands to create mock versions of interfaces, speeding up unit testing.                                                                                         |

## 6. Service Communication Patterns

### 6.1. Synchronous Communication: RESTful APIs with Enforced Schemas

For direct, request/response interactions.

- **Protocol:** **REST over HTTP/S with JSON payloads.**
- **Schema Enforcement (CRITICAL):**
  1.  **Design-First with OpenAPI:** All API contracts will be defined in **OpenAPI 3.0** specification files (`docs/api/*.yaml`). This is the single source of truth.
  2.  **Code Generation:** Use `oapi-codegen` to generate Go server stubs and request/response models (`structs`) directly from the OpenAPI spec.
  3.  **Benefit:** This guarantees that the server implementation and documentation are always in sync. For high-performance, internal-only communication in the future, services may evolve to use gRPC, but REST will remain the standard for all other synchronous APIs.

### 6.2. Asynchronous Communication: Event-Driven with a Message Bus

For decoupling services and handling real-time, fire-and-forget notifications between services.

- **Technology:** **RabbitMQ**.
- **Schema Enforcement:** All event payloads will be defined as Go structs in the shared `packages/events/` directory.
- **Idempotency:** All event consumers must be idempotent (safe to retry). This is achieved by tracking processed event IDs.

### 6.3. Example: Combined Communication Workflow

- **Scenario:** A seller lists a product, which requires a subscription check.
- **Synchronous Check (REST API):**
  1.  The `marketplace` service receives a request to create a listing.
  2.  It makes a blocking, synchronous HTTP call to the `core` service: `GET core-service/api/v1/subscriptions?entity_id={sellerID}`.
  3.  The `core` service returns the seller's subscription status, and `marketplace` enforces the business rule.
- **Asynchronous Notification (Message Bus):**
  1.  After the listing is created, the `marketplace` service publishes a `ProductListed` event to RabbitMQ.
  2.  The `core` service consumes this event and triggers its notification logic, sending an email to the seller. This decouples listing creation from notification.

## 7. Background Job Processing: A Robust Job Queue System

For tasks that must be executed reliably but not immediately (e.g., sending emails, processing images, generating reports), we will use a dedicated job queue system. This is distinct from the event bus.

- **Event Bus (RabbitMQ):** For _inter-service communication_. Its purpose is to decouple services.
- **Job Queue (Asynq/Redis):** For _intra-service background tasks_. Its purpose is to offload work reliably.

### 7.1. Implementation with `Asynq`

- **Technology:** We will use **Asynq**, a modern, Redis-backed job queue library for Go. It provides critical features like automatic retries, scheduled jobs, cron tasks, and a web UI for monitoring.
- **Architecture:** The API server and the background worker will be **two separate, independently scalable processes** that share the same core business logic.
  - `cmd/api/main.go`: The entry point for the HTTP API server. It _enqueues_ jobs.
  - `cmd/worker/main.go`: The entry point for the background worker. It _processes_ jobs.

### 7.2. Workflow: Sending a Welcome Email

1.  **Request:** A user signs up. The request hits an HTTP handler in the `core` service's API process.
2.  **Logic:** The application service creates the user in the database.
3.  **Enqueuing:** Instead of blocking the request to send an email, the service enqueues a `send-welcome-email` job with the user's ID as the payload. This is a non-blocking, sub-millisecond operation. The API immediately returns a `201 Created` response to the user.
4.  **Processing:** A `core` service worker process, running separately, picks up the job from the Redis queue.
5.  **Execution:** The worker executes the job, calling the notification service to send the email. If the email provider's API is down, `Asynq` will automatically retry the job with exponential backoff until it succeeds.

## 8. Comprehensive Testing Strategy & Tools

- **Unit Tests:**

  - **Tools:** Go's standard `testing` package, `testify/assert`, and `testify/mock`.
  - **Scope:** Test individual functions and business logic in complete isolation by mocking interface dependencies.

- **Integration Tests:**

  - **Tool:** **Testcontainers**.
  - **Scope:** Test a service with its real dependencies. Testcontainers will programmatically spin up fresh Docker containers (PostgreSQL, Redis) for each test run, providing high confidence in the data and integration layers.

- **Contract Tests (in lieu of traditional E2E):**

  - **Tool:** **Pact**, implemented using the Go library `github.com/pact-foundation/pact-go`.
  - **Rationale:** Traditional E2E tests are slow and brittle. Contract tests provide a superior alternative by verifying that inter-service communication contracts (API requests/responses) are honored. This ensures a consumer (`marketplace`) and a provider (`core`) can integrate correctly without needing to deploy both.
  - **Implementation:** Contract tests will reside in the top-level `/tests/contract` directory.
    1.  The **consumer** team (e.g., `marketplace`) writes a test using `pact-go` that mocks the provider (`core`) and defines the exact requests it will send and the responses it expects.
    2.  Running this test generates a `pact file` (a JSON contract).
    3.  In the **provider's** CI/CD pipeline, `pact-go` is used to verify the provider's API against this contract file, ensuring that any changes to the provider's API do not break the consumer's expectations.

- **Load Tests:**
  - **Tool:** **K6**.
  - **Scope:** Script realistic user scenarios to test system performance and identify bottlenecks under load before deploying to production.

---

All services and tools should use GO version 1.25 or later to leverage the latest language features and performance improvements.
