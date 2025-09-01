module github.com/paradoxe35/torra/services/core

go 1.25

replace (
	github.com/paradoxe35/torra/packages/cache => ../../packages/cache
	github.com/paradoxe35/torra/packages/config => ../../packages/config
	github.com/paradoxe35/torra/packages/database => ../../packages/database
	github.com/paradoxe35/torra/packages/errors => ../../packages/errors
	github.com/paradoxe35/torra/packages/events => ../../packages/events
	github.com/paradoxe35/torra/packages/http => ../../packages/http
	github.com/paradoxe35/torra/packages/i18n => ../../packages/i18n
	github.com/paradoxe35/torra/packages/logger => ../../packages/logger
	github.com/paradoxe35/torra/packages/security => ../../packages/security
)

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/go-playground/validator/v10 v10.27.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.6.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/lib/pq v1.10.9
	github.com/paradoxe35/torra/packages/cache v0.0.0-00010101000000-000000000000
	github.com/paradoxe35/torra/packages/config v0.0.0
	github.com/paradoxe35/torra/packages/database v0.0.0-00010101000000-000000000000
	github.com/paradoxe35/torra/packages/errors v0.0.0-00010101000000-000000000000
	github.com/paradoxe35/torra/packages/events v0.0.0-00010101000000-000000000000
	github.com/paradoxe35/torra/packages/http v0.0.0-00010101000000-000000000000
	github.com/paradoxe35/torra/packages/i18n v0.0.0
	github.com/paradoxe35/torra/packages/logger v0.0.0
	github.com/paradoxe35/torra/packages/security v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/pquerna/otp v1.5.0 // indirect
	github.com/sagikazarmark/locafero v0.10.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.14.0 // indirect
	github.com/spf13/cast v1.9.2 // indirect
	github.com/spf13/pflag v1.0.7 // indirect
	github.com/spf13/viper v1.20.1 // indirect
	github.com/streadway/amqp v1.1.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
