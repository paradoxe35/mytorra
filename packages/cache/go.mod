module github.com/paradoxe35/torra/packages/cache

go 1.25

require (
    github.com/go-redis/redis/v8 v8.11.5
    github.com/paradoxe35/torra/packages/config v0.0.0
)

replace github.com/paradoxe35/torra/packages/config => ../config