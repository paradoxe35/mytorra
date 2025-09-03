module github.com/paradoxe35/mytorra/services/core

go 1.25

replace (
	github.com/paradoxe35/mytorra/packages/cache => ../../packages/cache
	github.com/paradoxe35/mytorra/packages/config => ../../packages/config
	github.com/paradoxe35/mytorra/packages/database => ../../packages/database
	github.com/paradoxe35/mytorra/packages/errors => ../../packages/errors
	github.com/paradoxe35/mytorra/packages/events => ../../packages/events
	github.com/paradoxe35/mytorra/packages/http => ../../packages/http
	github.com/paradoxe35/mytorra/packages/i18n => ../../packages/i18n
	github.com/paradoxe35/mytorra/packages/logger => ../../packages/logger
	github.com/paradoxe35/mytorra/packages/security => ../../packages/security
)

require (
	ariga.io/atlas-provider-bun v0.0.2
	github.com/go-chi/chi/v5 v5.2.3
	github.com/go-playground/validator/v10 v10.27.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	github.com/paradoxe35/mytorra/packages/cache v0.0.0-00010101000000-000000000000
	github.com/paradoxe35/mytorra/packages/config v0.0.0
	github.com/paradoxe35/mytorra/packages/database v0.0.0-00010101000000-000000000000
	github.com/paradoxe35/mytorra/packages/errors v0.0.0-00010101000000-000000000000
	github.com/paradoxe35/mytorra/packages/events v0.0.0-00010101000000-000000000000
	github.com/paradoxe35/mytorra/packages/http v0.0.0-00010101000000-000000000000
	github.com/paradoxe35/mytorra/packages/i18n v0.0.0
	github.com/paradoxe35/mytorra/packages/logger v0.0.0
	github.com/paradoxe35/mytorra/packages/security v0.0.0-00010101000000-000000000000
	github.com/restayway/gogis v1.0.1
	github.com/shopspring/decimal v1.4.0
	github.com/stretchr/testify v1.11.1
	github.com/stripe/stripe-go/v76 v76.25.0
	github.com/uptrace/bun v1.2.15
	golang.org/x/oauth2 v0.30.0
)

require (
	ariga.io/atlas v0.36.2-0.20250806044935-5bb51a0a956e // indirect
	cloud.google.com/go/compute/metadata v0.8.0 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/jackc/pgerrcode v0.0.0-20240316143900-6e2875d9b438 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.5 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/pquerna/otp v1.5.0 // indirect
	github.com/puzpuzpuz/xsync/v3 v3.5.1 // indirect
	github.com/sagikazarmark/locafero v0.10.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.14.0 // indirect
	github.com/spf13/cast v1.9.2 // indirect
	github.com/spf13/pflag v1.0.7 // indirect
	github.com/spf13/viper v1.20.1 // indirect
	github.com/streadway/amqp v1.1.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/uptrace/bun/dialect/mssqldialect v1.2.15 // indirect
	github.com/uptrace/bun/dialect/mysqldialect v1.2.15 // indirect
	github.com/uptrace/bun/dialect/oracledialect v1.2.15 // indirect
	github.com/uptrace/bun/dialect/pgdialect v1.2.15 // indirect
	github.com/uptrace/bun/dialect/sqlitedialect v1.2.15 // indirect
	github.com/uptrace/bun/driver/pgdriver v1.2.15 // indirect
	github.com/uptrace/bun/extra/bundebug v1.2.15 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/mod v0.27.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	mellium.im/sasl v0.3.2 // indirect
)
