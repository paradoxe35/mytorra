# GORM to Bun ORM Migration Guide

## Overview
The database package has been migrated from GORM to Bun ORM for better performance and more explicit SQL control.

## Key Changes

### 1. Database Connection
- **Old**: `*GormDB` wrapper around `*gorm.DB`
- **New**: `*DB` wrapper around `*bun.DB` with pgx pool support

### 2. Configuration
- Uses pgx v5 with connection pooling
- Production-optimized settings: `maxOpenConns = 4 * runtime.GOMAXPROCS(0)`
- Simple query execution mode for better performance

### 3. Models

#### BaseModel
```go
// Bun BaseModel with soft delete support
type BaseModel struct {
    ID        uuid.UUID    `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
    CreatedAt time.Time    `bun:"created_at,nullzero,notnull,default:current_timestamp"`
    UpdatedAt time.Time    `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
    DeletedAt sql.NullTime `bun:"deleted_at,soft_delete,nullzero"`
}
```

### 4. Query Examples

#### Select
```go
// GORM
var users []UserModel
db.Where("active = ?", true).Find(&users)

// Bun
var users []UserModel
err := db.NewSelect().
    Model(&users).
    Where("active = ?", true).
    Scan(ctx)
```

#### Insert
```go
// GORM
db.Create(&user)

// Bun
_, err := db.NewInsert().
    Model(&user).
    Exec(ctx)
```

#### Update
```go
// GORM
db.Model(&user).Updates(map[string]interface{}{"name": "John"})

// Bun
_, err := db.NewUpdate().
    Model(&user).
    Set("name = ?", "John").
    Where("id = ?", user.ID).
    Exec(ctx)
```

#### Delete (Soft)
```go
// GORM
db.Delete(&user)

// Bun
_, err := db.NewDelete().
    Model(&user).
    Where("id = ?", user.ID).
    Exec(ctx)
```

### 5. Transactions
```go
err := db.Transaction(ctx, func(ctx context.Context, tx bun.Tx) error {
    // Your transaction logic here
    return nil
})
```

### 6. Pagination
```go
paginate := NewPaginate(page, perPage)
query := db.NewSelect().Model(&users)
query = paginate.Apply(query)
```

### 7. PostGIS Support
The package includes PostGIS types for spatial data:
- `Point`: Geographic point with lat/lng
- `Polygon`: Geographic polygon
- `LineString`: Geographic line
- `Geography`: Generic GeoJSON geography

### 8. Testing
Test helpers have been updated to work with Bun:
```go
func TestExample(t *testing.T) {
    db := NewTestDB(t)
    ctx := context.Background()

    // Create tables
    err := db.CreateTables(ctx, (*User)(nil))

    // Seed data
    err = db.SeedData(ctx, &user)

    // Assertions
    db.AssertExists(t, ctx, &User{}, "id", userID)
}
```

## Migration Checklist

When migrating services:
1. Replace GORM imports with Bun imports
2. Update model tags from `gorm:` to `bun:`
3. Add context to all database operations
4. Replace GORM query chains with Bun query builders
5. Update transaction handling to use context
6. Update error checking (use `pgdriver.Error` for PostgreSQL errors)

## Benefits
- Better performance with pgx driver
- More explicit SQL control
- Better connection pooling
- Native PostGIS support
- Simpler API with context support throughout
