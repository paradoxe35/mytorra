module github.com/paradoxe35/torra/packages/http

go 1.25

require (
	github.com/go-chi/chi/v5 v5.0.11
	github.com/google/uuid v1.5.0
	github.com/paradoxe35/torra/packages/logger v0.0.0
)

replace github.com/paradoxe35/torra/packages/logger => ../logger
