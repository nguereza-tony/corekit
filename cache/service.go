package cache

import (
	"context"
	"encoding/json"
	"time"
)

// CacheService provides cache operations with any type (uses JSON serialization)
type CacheService struct {
	cache Cache
}

// NewCacheService creates a new cache service
func NewCacheService(cache Cache) *CacheService {
	return &CacheService{
		cache: cache,
	}
}

// Get retrieves and unmarshals a value from cache
func (s *CacheService) Get(ctx context.Context, key string, dest any) error {
	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return err
	}

	return nil
}

// Set marshals and stores a value in cache
func (s *CacheService) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.cache.Set(ctx, key, string(data), ttl)
}

// Delete removes a key from cache
func (s *CacheService) Delete(ctx context.Context, key string) error {
	return s.cache.Delete(ctx, key)
}

// Exists checks if a key exists in cache
func (s *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	return s.cache.Exists(ctx, key)
}

// Expire sets expiration on an existing key
func (s *CacheService) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return s.cache.Expire(ctx, key, ttl)
}

// Increment increments a counter by 1
func (s *CacheService) Increment(ctx context.Context, key string) (int64, error) {
	return s.cache.Increment(ctx, key)
}

// IncrementBy increments a counter by a specific amount
func (s *CacheService) IncrementBy(ctx context.Context, key string, delta int64) (int64, error) {
	return s.cache.IncrementBy(ctx, key, delta)
}

// Decrement decrements a counter by 1
func (s *CacheService) Decrement(ctx context.Context, key string) (int64, error) {
	return s.cache.Decrement(ctx, key)
}

// DecrementBy decrements a counter by a specific amount
func (s *CacheService) DecrementBy(ctx context.Context, key string, delta int64) (int64, error) {
	return s.cache.DecrementBy(ctx, key, delta)
}

// Clear removes all keys from cache
func (s *CacheService) Clear(ctx context.Context) error {
	return s.cache.Clear(ctx)
}

// Ping checks if the cache is reachable
func (s *CacheService) Ping(ctx context.Context) error {
	return s.cache.Ping(ctx)
}

// Close closes the cache connection
func (s *CacheService) Close() error {
	return s.cache.Close()
}

// GetRaw returns the underlying cache
func (s *CacheService) GetRaw() Cache {
	return s.cache
}
