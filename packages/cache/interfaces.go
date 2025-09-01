package cache

import (
	"context"
	"time"
)

// Cache defines the interface for caching operations used across all services
// This interface supports all the caching patterns needed by the application layer
type Cache interface {
	// Basic operations
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

// ExtendedCache adds additional operations that might be needed by specific services
type ExtendedCache interface {
	Cache

	// Batch operations
	MGet(ctx context.Context, keys ...string) ([][]byte, error)
	MSet(ctx context.Context, pairs map[string][]byte, ttl time.Duration) error

	// Atomic operations
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)

	// Lock operations (for distributed locking)
	Lock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error

	// Pub/Sub operations
	Publish(ctx context.Context, channel string, message []byte) error
	Subscribe(ctx context.Context, channels ...string) (<-chan Message, error)

	// Pipeline operations for better performance
	Pipeline() Pipeline

	// Health check
	Ping(ctx context.Context) error

	// Administrative operations
	FlushDB(ctx context.Context) error
	Close() error
}

// Message represents a pub/sub message
type Message struct {
	Channel string
	Data    []byte
}

// Pipeline represents a batch of commands to be executed atomically
type Pipeline interface {
	Get(key string) *StringCmd
	Set(key string, value []byte, ttl time.Duration) *StatusCmd
	Del(keys ...string) *IntCmd
	Exec(ctx context.Context) error
}

// Command result types for pipeline operations
type StringCmd struct {
	err error
	val []byte
}

func (cmd *StringCmd) Result() ([]byte, error) {
	return cmd.val, cmd.err
}

type StatusCmd struct {
	err error
	val string
}

func (cmd *StatusCmd) Result() (string, error) {
	return cmd.val, cmd.err
}

type IntCmd struct {
	err error
	val int64
}

func (cmd *IntCmd) Result() (int64, error) {
	return cmd.val, cmd.err
}
