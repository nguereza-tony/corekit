package cache

import (
	"context"
	"time"
)

// Cache defines the basic cache operations
type Cache interface {
	// Get retrieves a value from cache
	// Returns ErrKeyNotFound if key doesn't exist
	Get(ctx context.Context, key string) (string, error)

	// Set stores a value in cache with expiration
	// If ttl <= 0, the key never expires
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Delete removes a key from cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)

	// Expire sets expiration on an existing key
	Expire(ctx context.Context, key string, ttl time.Duration) error

	// Increment increments a counter
	Increment(ctx context.Context, key string) (int64, error)

	// IncrementBy increments a counter by a specific amount
	IncrementBy(ctx context.Context, key string, delta int64) (int64, error)

	// Decrement decrements a counter
	Decrement(ctx context.Context, key string) (int64, error)

	// DecrementBy decrements a counter by a specific amount
	DecrementBy(ctx context.Context, key string, delta int64) (int64, error)

	// Clear removes all keys from cache
	Clear(ctx context.Context) error

	// Ping checks if the cache is reachable
	Ping(ctx context.Context) error

	// Close closes the cache connection
	Close() error
}
