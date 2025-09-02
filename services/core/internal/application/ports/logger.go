package ports

import (
	"github.com/paradoxe35/mytorra/packages/logger"
)

// Logger is an alias to the shared logger interface
// This allows the application layer to use the standardized logger interface
type Logger = logger.Interface
