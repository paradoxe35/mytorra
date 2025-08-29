package ports

import (
	"context"
	"time"
)

// Cache defines the interface for caching operations
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Hash operations
	HGet(ctx context.Context, key, field string) ([]byte, error)
	HSet(ctx context.Context, key, field string, value []byte) error
	HGetAll(ctx context.Context, key string) (map[string][]byte, error)
	HDel(ctx context.Context, key string, fields ...string) error

	// List operations
	LPush(ctx context.Context, key string, values ...[]byte) error
	RPop(ctx context.Context, key string) ([]byte, error)
	LRange(ctx context.Context, key string, start, stop int64) ([][]byte, error)

	// Set operations
	SAdd(ctx context.Context, key string, members ...[]byte) error
	SMembers(ctx context.Context, key string) ([][]byte, error)
	SRem(ctx context.Context, key string, members ...[]byte) error
	SIsMember(ctx context.Context, key string, member []byte) (bool, error)
}
