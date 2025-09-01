package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/paradoxe35/torra/packages/config"
)

// RedisCacheAdapter implements the Cache interface using Redis
type RedisCacheAdapter struct {
	client *redis.Client
	prefix string
}

// NewRedisCacheAdapter creates a new Redis cache adapter that implements the Cache interface
func NewRedisCacheAdapter(cfg *config.RedisConfig, prefix string) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.Database,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCacheAdapter{
		client: client,
		prefix: prefix,
	}, nil
}

func (c *RedisCacheAdapter) key(key string) string {
	if c.prefix != "" {
		return fmt.Sprintf("%s:%s", c.prefix, key)
	}
	return key
}

// Basic operations
func (c *RedisCacheAdapter) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, c.key(key)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (c *RedisCacheAdapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, c.key(key), value, ttl).Err()
}

func (c *RedisCacheAdapter) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.key(key)).Err()
}

func (c *RedisCacheAdapter) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, c.key(key)).Result()
	return n > 0, err
}

func (c *RedisCacheAdapter) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(ctx, c.key(key), ttl).Err()
}

func (c *RedisCacheAdapter) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, c.key(key)).Result()
}

// Hash operations
func (c *RedisCacheAdapter) HGet(ctx context.Context, key, field string) ([]byte, error) {
	val, err := c.client.HGet(ctx, c.key(key), field).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (c *RedisCacheAdapter) HSet(ctx context.Context, key, field string, value []byte) error {
	return c.client.HSet(ctx, c.key(key), field, value).Err()
}

func (c *RedisCacheAdapter) HGetAll(ctx context.Context, key string) (map[string][]byte, error) {
	result, err := c.client.HGetAll(ctx, c.key(key)).Result()
	if err != nil {
		return nil, err
	}

	byteMap := make(map[string][]byte)
	for field, value := range result {
		byteMap[field] = []byte(value)
	}
	return byteMap, nil
}

func (c *RedisCacheAdapter) HDel(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, c.key(key), fields...).Err()
}

// List operations
func (c *RedisCacheAdapter) LPush(ctx context.Context, key string, values ...[]byte) error {
	interfaces := make([]interface{}, len(values))
	for i, v := range values {
		interfaces[i] = v
	}
	return c.client.LPush(ctx, c.key(key), interfaces...).Err()
}

func (c *RedisCacheAdapter) RPop(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.RPop(ctx, c.key(key)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (c *RedisCacheAdapter) LRange(ctx context.Context, key string, start, stop int64) ([][]byte, error) {
	vals, err := c.client.LRange(ctx, c.key(key), start, stop).Result()
	if err != nil {
		return nil, err
	}

	result := make([][]byte, len(vals))
	for i, v := range vals {
		result[i] = []byte(v)
	}
	return result, nil
}

// Set operations
func (c *RedisCacheAdapter) SAdd(ctx context.Context, key string, members ...[]byte) error {
	interfaces := make([]interface{}, len(members))
	for i, v := range members {
		interfaces[i] = v
	}
	return c.client.SAdd(ctx, c.key(key), interfaces...).Err()
}

func (c *RedisCacheAdapter) SMembers(ctx context.Context, key string) ([][]byte, error) {
	vals, err := c.client.SMembers(ctx, c.key(key)).Result()
	if err != nil {
		return nil, err
	}

	result := make([][]byte, len(vals))
	for i, v := range vals {
		result[i] = []byte(v)
	}
	return result, nil
}

func (c *RedisCacheAdapter) SRem(ctx context.Context, key string, members ...[]byte) error {
	interfaces := make([]interface{}, len(members))
	for i, v := range members {
		interfaces[i] = v
	}
	return c.client.SRem(ctx, c.key(key), interfaces...).Err()
}

func (c *RedisCacheAdapter) SIsMember(ctx context.Context, key string, member []byte) (bool, error) {
	return c.client.SIsMember(ctx, c.key(key), member).Result()
}

// Helper methods for JSON marshaling/unmarshaling
func (c *RedisCacheAdapter) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	if data == nil {
		return nil
	}
	return json.Unmarshal(data, dest)
}

func (c *RedisCacheAdapter) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return c.Set(ctx, key, data, ttl)
}

// Close closes the Redis connection
func (c *RedisCacheAdapter) Close() error {
	return c.client.Close()
}

// Flush flushes keys with the prefix or the entire database
func (c *RedisCacheAdapter) Flush(ctx context.Context) error {
	if c.prefix != "" {
		iter := c.client.Scan(ctx, 0, c.prefix+":*", 0).Iterator()
		var keys []string
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}
		if err := iter.Err(); err != nil {
			return err
		}
		if len(keys) > 0 {
			return c.client.Del(ctx, keys...).Err()
		}
		return nil
	}
	return c.client.FlushDB(ctx).Err()
}

// Additional utility methods

// IncrBy increments the integer value of a key by the given amount
func (c *RedisCacheAdapter) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, c.key(key), value).Result()
}

// DecrBy decrements the integer value of a key by the given amount
func (c *RedisCacheAdapter) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.DecrBy(ctx, c.key(key), value).Result()
}

// SetNX sets key to hold string value if key does not exist
func (c *RedisCacheAdapter) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	return c.client.SetNX(ctx, c.key(key), value, ttl).Result()
}

// MGet gets the values of all specified keys
func (c *RedisCacheAdapter) MGet(ctx context.Context, keys ...string) ([][]byte, error) {
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.key(key)
	}

	vals, err := c.client.MGet(ctx, prefixedKeys...).Result()
	if err != nil {
		return nil, err
	}

	result := make([][]byte, len(vals))
	for i, val := range vals {
		if val != nil {
			if str, ok := val.(string); ok {
				result[i] = []byte(str)
			}
		}
	}
	return result, nil
}

// Lock implements distributed locking using Redis
func (c *RedisCacheAdapter) Lock(ctx context.Context, lockKey string, ttl time.Duration) (bool, error) {
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())
	success, err := c.client.SetNX(ctx, c.key("lock:"+lockKey), lockValue, ttl).Result()
	return success, err
}

// Unlock releases a distributed lock
func (c *RedisCacheAdapter) Unlock(ctx context.Context, lockKey string) error {
	return c.client.Del(ctx, c.key("lock:"+lockKey)).Err()
}
