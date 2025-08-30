package redis

import (
	"context"
	"time"

	"github.com/paradoxe35/torra/packages/cache"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
)

// CacheAdapter wraps the shared cache package to implement the application port interface
type CacheAdapter struct {
	cache *cache.RedisCache
}

// NewCacheAdapter creates a new cache adapter
func NewCacheAdapter(cache *cache.RedisCache) ports.Cache {
	return &CacheAdapter{cache: cache}
}

func (c *CacheAdapter) Get(ctx context.Context, key string) ([]byte, error) {
	return c.cache.Get(ctx, key)
}

func (c *CacheAdapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.cache.Set(ctx, key, value, ttl)
}

func (c *CacheAdapter) Delete(ctx context.Context, key string) error {
	return c.cache.Delete(ctx, key)
}

func (c *CacheAdapter) Exists(ctx context.Context, key string) (bool, error) {
	return c.cache.Exists(ctx, key)
}

func (c *CacheAdapter) Expire(ctx context.Context, key string, ttl time.Duration) error {
	// The shared cache doesn't have an Expire method, so we'll implement it by getting and setting
	data, err := c.cache.Get(ctx, key)
	if err != nil {
		return err
	}
	if data == nil {
		return nil // Key doesn't exist
	}
	return c.cache.Set(ctx, key, data, ttl)
}

func (c *CacheAdapter) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.cache.TTL(ctx, key)
}

// Hash operations - these need to be implemented based on Redis commands
// For now, we'll use prefixed keys to simulate hash operations

func (c *CacheAdapter) HGet(ctx context.Context, key, field string) ([]byte, error) {
	hashKey := key + ":" + field
	return c.cache.Get(ctx, hashKey)
}

func (c *CacheAdapter) HSet(ctx context.Context, key, field string, value []byte) error {
	hashKey := key + ":" + field
	return c.cache.Set(ctx, hashKey, value, 0) // No expiration for hash fields
}

func (c *CacheAdapter) HGetAll(ctx context.Context, key string) (map[string][]byte, error) {
	// This is a simplified implementation
	// In a real implementation, you'd need to track all fields in a hash
	return make(map[string][]byte), nil
}

func (c *CacheAdapter) HDel(ctx context.Context, key string, fields ...string) error {
	for _, field := range fields {
		hashKey := key + ":" + field
		if err := c.cache.Delete(ctx, hashKey); err != nil {
			return err
		}
	}
	return nil
}

// List operations - simplified implementations

func (c *CacheAdapter) LPush(ctx context.Context, key string, values ...[]byte) error {
	// Simplified: store as a single value for now
	if len(values) > 0 {
		return c.cache.Set(ctx, key, values[0], 0)
	}
	return nil
}

func (c *CacheAdapter) RPop(ctx context.Context, key string) ([]byte, error) {
	data, err := c.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	// Delete after popping
	c.cache.Delete(ctx, key)
	return data, nil
}

func (c *CacheAdapter) LRange(ctx context.Context, key string, start, stop int64) ([][]byte, error) {
	// Simplified: return single value as array
	data, err := c.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return [][]byte{}, nil
	}
	return [][]byte{data}, nil
}

// Set operations - simplified implementations

func (c *CacheAdapter) SAdd(ctx context.Context, key string, members ...[]byte) error {
	// Simplified: store first member only
	if len(members) > 0 {
		return c.cache.Set(ctx, key, members[0], 0)
	}
	return nil
}

func (c *CacheAdapter) SMembers(ctx context.Context, key string) ([][]byte, error) {
	data, err := c.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return [][]byte{}, nil
	}
	return [][]byte{data}, nil
}

func (c *CacheAdapter) SRem(ctx context.Context, key string, members ...[]byte) error {
	// Simplified: just delete the key
	return c.cache.Delete(ctx, key)
}

func (c *CacheAdapter) SIsMember(ctx context.Context, key string, member []byte) (bool, error) {
	data, err := c.cache.Get(ctx, key)
	if err != nil {
		return false, err
	}
	// Simplified comparison
	return data != nil && string(data) == string(member), nil
}
