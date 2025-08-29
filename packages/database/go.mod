module github.com/paradoxe35/torra/packages/database

go 1.25

require (
    github.com/jmoiron/sqlx v1.3.5
    github.com/lib/pq v1.10.9
    github.com/paradoxe35/torra/packages/config v0.0.0
)

replace github.com/paradoxe35/torra/packages/config => ../config