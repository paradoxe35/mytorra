module github.com/paradoxe35/torra/services/core

go 1.25

replace (
	github.com/paradoxe35/torra/packages/cache => ../../packages/cache
	github.com/paradoxe35/torra/packages/config => ../../packages/config
	github.com/paradoxe35/torra/packages/database => ../../packages/database
	github.com/paradoxe35/torra/packages/errors => ../../packages/errors
	github.com/paradoxe35/torra/packages/events => ../../packages/events
	github.com/paradoxe35/torra/packages/http => ../../packages/http
	github.com/paradoxe35/torra/packages/logger => ../../packages/logger
	github.com/paradoxe35/torra/packages/security => ../../packages/security
)
