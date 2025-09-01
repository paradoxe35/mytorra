package ports

import (
	"github.com/paradoxe35/torra/packages/cache"
)

// Cache is an alias to the shared cache interface
// This allows the application layer to use the standardized cache interface
type Cache = cache.Cache
