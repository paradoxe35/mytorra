package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/paradoxe35/mytorra/packages/config"
)

// BasicCache is the simple cache interface (keeping for backward compatibility)
type BasicCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Flush(ctx context.Context) error
}

type RedisCache struct {
	client *redis.Client
	prefix string
}

func NewRedisCache(cfg *config.RedisConfig, prefix string) (*RedisCache, error) {
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

	return &RedisCache{
		client: client,
		prefix: prefix,
	}, nil
}

func (c *RedisCache) key(key string) string {
	if c.prefix != "" {
		return fmt.Sprintf("%s:%s", c.prefix, key)
	}
	return key
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, c.key(key)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (c *RedisCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	if data == nil {
		return nil
	}
	return json.Unmarshal(data, dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var data []byte
	var err error

	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		data, err = json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
	}

	return c.client.Set(ctx, c.key(key), data, expiration).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.key(key)).Err()
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, c.key(key)).Result()
	return n > 0, err
}

func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, c.key(key)).Result()
}

func (c *RedisCache) Flush(ctx context.Context) error {
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

func (c *RedisCache) Close() error {
	return c.client.Close()
}
