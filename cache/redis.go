package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisCache implements Cache interface using Redis
type redisCache struct {
	client *redis.Client
	prefix string
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(cfg *CacheConfig) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		DialTimeout:  time.Duration(cfg.Redis.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.Redis.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Redis.WriteTimeout) * time.Second,
		MaxRetries:   cfg.Redis.MaxRetries,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	prefix := cfg.Prefix
	if prefix == "" {
		prefix = "futa"
	}

	return &redisCache{
		client: client,
		prefix: prefix,
	}, nil
}

// Get retrieves a string value from cache
func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", ErrInvalidKey
	}

	result, err := r.client.Get(ctx, r.buildKey(key)).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOperationFailed, err)
	}

	return result, nil
}

// Set stores a string value in cache with expiration
func (r *redisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}

	if ttl < 0 {
		return ErrInvalidTTL
	}

	if ttl == 0 {
		// Store without expiration
		err := r.client.Set(ctx, r.buildKey(key), value, 0).Err()
		if err != nil {
			return fmt.Errorf("%w: %v", ErrOperationFailed, err)
		}
		return nil
	}

	err := r.client.Set(ctx, r.buildKey(key), value, ttl).Err()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOperationFailed, err)
	}

	return nil
}

// Delete removes a key from cache
func (r *redisCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return ErrInvalidKey
	}
	err := r.client.Del(ctx, r.buildKey(key)).Err()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOperationFailed, err)
	}

	return nil
}

// Exists checks if a key exists in cache
func (r *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, ErrInvalidKey
	}

	result, err := r.client.Exists(ctx, r.buildKey(key)).Result()
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrOperationFailed, err)
	}

	return result > 0, nil
}

// Expire sets expiration on an existing key
func (r *redisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}

	if ttl < 0 {
		return ErrInvalidTTL
	}

	if ttl == 0 {
		// Remove expiration (persist key)
		err := r.client.Persist(ctx, r.buildKey(key)).Err()
		if err == redis.Nil {
			return ErrKeyNotFound
		}
		if err != nil {
			return fmt.Errorf("%w: %v", ErrOperationFailed, err)
		}
		return nil
	}

	// Set expiration
	ok, err := r.client.Expire(ctx, r.buildKey(key), ttl).Result()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOperationFailed, err)
	}
	if !ok {
		return ErrKeyNotFound
	}

	return nil
}

// Increment increments a counter by 1
func (r *redisCache) Increment(ctx context.Context, key string) (int64, error) {
	return r.IncrementBy(ctx, key, 1)
}

// IncrementBy increments a counter by a specific amount
func (r *redisCache) IncrementBy(ctx context.Context, key string, delta int64) (int64, error) {
	if key == "" {
		return 0, ErrInvalidKey
	}

	result, err := r.client.IncrBy(ctx, r.buildKey(key), delta).Result()
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrOperationFailed, err)
	}

	return result, nil
}

// Decrement decrements a counter by 1
func (r *redisCache) Decrement(ctx context.Context, key string) (int64, error) {
	return r.DecrementBy(ctx, key, 1)
}

// DecrementBy decrements a counter by a specific amount
func (r *redisCache) DecrementBy(ctx context.Context, key string, delta int64) (int64, error) {
	if key == "" {
		return 0, ErrInvalidKey
	}

	result, err := r.client.DecrBy(ctx, r.buildKey(key), delta).Result()
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrOperationFailed, err)
	}

	return result, nil
}

// Clear removes all keys from cache (with prefix)
func (r *redisCache) Clear(ctx context.Context) error {
	// Delete all keys with the prefix
	pattern := fmt.Sprintf("%s:*", r.prefix)

	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrOperationFailed, err)
	}

	if len(keys) == 0 {
		return nil
	}

	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOperationFailed, err)
	}

	return nil
}

// Ping checks if the cache is reachable
func (r *redisCache) Ping(ctx context.Context) error {
	err := r.client.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	return nil
}

// Close closes the Redis connection
func (r *redisCache) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// GetClient returns the underlying Redis client (for advanced operations)
func (r *redisCache) GetClient() *redis.Client {
	return r.client
}

// GetPrefix returns the key prefix
func (r *redisCache) GetPrefix() string {
	return r.prefix
}

// buildKey builds a prefixed key
func (r *redisCache) buildKey(key string) string {
	if key == "" {
		return ""
	}
	return fmt.Sprintf("%s:%s", r.prefix, key)
}
