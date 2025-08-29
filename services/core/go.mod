module github.com/paradoxe35/torra/services/core

go 1.25

require (
    github.com/go-chi/chi/v5 v5.0.11
    github.com/jmoiron/sqlx v1.3.5
    github.com/lib/pq v1.10.9
    github.com/paradoxe35/torra/packages/config v0.0.0
    github.com/paradoxe35/torra/packages/logger v0.0.0
    github.com/paradoxe35/torra/packages/security v0.0.0
    github.com/paradoxe35/torra/packages/http v0.0.0
    github.com/paradoxe35/torra/packages/database v0.0.0
    github.com/paradoxe35/torra/packages/events v0.0.0
    github.com/paradoxe35/torra/packages/errors v0.0.0
    github.com/paradoxe35/torra/packages/cache v0.0.0
)

replace (
    github.com/paradoxe35/torra/packages/config => ../../packages/config
    github.com/paradoxe35/torra/packages/logger => ../../packages/logger
    github.com/paradoxe35/torra/packages/security => ../../packages/security
    github.com/paradoxe35/torra/packages/http => ../../packages/http
    github.com/paradoxe35/torra/packages/database => ../../packages/database
    github.com/paradoxe35/torra/packages/events => ../../packages/events
    github.com/paradoxe35/torra/packages/errors => ../../packages/errors
    github.com/paradoxe35/torra/packages/cache => ../../packages/cache
)