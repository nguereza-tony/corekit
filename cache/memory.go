package cache

import (
	"context"
	"strconv"
	"sync"
	"time"
)

// memoryItem represents a cached item in memory
type memoryItem struct {
	value     string
	expiresAt time.Time
	createdAt time.Time
}

// isExpired checks if the item has expired
func (i *memoryItem) isExpired() bool {
	if i.expiresAt.IsZero() {
		return false
	}
	return time.Now().After(i.expiresAt)
}

// memoryCache implements Cache interface using in-memory map
type memoryCache struct {
	mu              sync.RWMutex
	items           map[string]*memoryItem
	cleanupInterval time.Duration
	maxItems        int
	stopCleanup     chan struct{}
	closed          bool
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(cfg *MemoryCacheConfig) Cache {
	cleanupInterval := time.Duration(cfg.CleanupInterval) * time.Second
	if cleanupInterval <= 0 {
		cleanupInterval = 60 * time.Second
	}

	maxItems := cfg.MaxItems
	if maxItems <= 0 {
		maxItems = 10000
	}

	c := &memoryCache{
		items:           make(map[string]*memoryItem),
		cleanupInterval: cleanupInterval,
		maxItems:        maxItems,
		stopCleanup:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go c.cleanupLoop()

	return c
}

// Get retrieves a value from cache
func (c *memoryCache) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", ErrInvalidKey
	}

	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return "", ErrKeyNotFound
	}

	// Check if expired
	if item.isExpired() {
		// Delete expired item
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return "", ErrKeyNotFound
	}

	return item.value, nil
}

// Set stores a value in cache with expiration
func (c *memoryCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}

	if ttl < 0 {
		return ErrInvalidTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check max items limit
	if len(c.items) >= c.maxItems {
		// Try to remove expired items first
		c.evictExpiredLocked()

		// If still full, evict oldest
		if len(c.items) >= c.maxItems {
			c.evictOldestLocked()
		}
	}

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.items[key] = &memoryItem{
		value:     value,
		expiresAt: expiresAt,
		createdAt: time.Now(),
	}

	return nil
}

// Delete removes a key from cache
func (c *memoryCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return ErrInvalidKey
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

// Exists checks if a key exists in cache
func (c *memoryCache) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, ErrInvalidKey
	}

	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return false, nil
	}

	// Check if expired
	if item.isExpired() {
		// Delete expired item
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return false, nil
	}

	return true, nil
}

// Expire sets expiration on an existing key
func (c *memoryCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}

	if ttl < 0 {
		return ErrInvalidTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists {
		return ErrKeyNotFound
	}

	// Check if expired
	if item.isExpired() {
		delete(c.items, key)
		return ErrKeyNotFound
	}

	// Set new expiration
	if ttl > 0 {
		item.expiresAt = time.Now().Add(ttl)
	} else {
		item.expiresAt = time.Time{} // No expiration
	}

	return nil
}

// Increment increments a counter by 1
func (c *memoryCache) Increment(ctx context.Context, key string) (int64, error) {
	return c.IncrementBy(ctx, key, 1)
}

// IncrementBy increments a counter by a specific amount
func (c *memoryCache) IncrementBy(ctx context.Context, key string, delta int64) (int64, error) {
	if key == "" {
		return 0, ErrInvalidKey
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if key exists
	item, exists := c.items[key]
	if !exists {
		// Create new counter with value = delta
		c.items[key] = &memoryItem{
			value:     strconv.FormatInt(delta, 10),
			expiresAt: time.Time{}, // No expiration by default
			createdAt: time.Now(),
		}
		return delta, nil
	}

	// Check if expired
	if item.isExpired() {
		delete(c.items, key)
		// Create new counter with value = delta
		c.items[key] = &memoryItem{
			value:     strconv.FormatInt(delta, 10),
			expiresAt: time.Time{},
			createdAt: time.Now(),
		}
		return delta, nil
	}

	// Parse current value
	current, err := strconv.ParseInt(item.value, 10, 64)
	if err != nil {
		return 0, ErrOperationFailed
	}

	// Calculate new value
	newValue := current + delta
	item.value = strconv.FormatInt(newValue, 10)

	return newValue, nil
}

// Decrement decrements a counter by 1
func (c *memoryCache) Decrement(ctx context.Context, key string) (int64, error) {
	return c.DecrementBy(ctx, key, 1)
}

// DecrementBy decrements a counter by a specific amount
func (c *memoryCache) DecrementBy(ctx context.Context, key string, delta int64) (int64, error) {
	return c.IncrementBy(ctx, key, -delta)
}

// Clear removes all keys from cache
func (c *memoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create new map to free memory
	c.items = make(map[string]*memoryItem)
	return nil
}

// Ping checks if the cache is reachable
func (c *memoryCache) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return ErrConnectionFailed
	}

	return nil
}

// Close closes the cache connection
func (c *memoryCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	close(c.stopCleanup)
	c.items = nil
	return nil
}

// cleanupLoop periodically removes expired items
func (c *memoryCache) cleanupLoop() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanup removes all expired items
func (c *memoryCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.evictExpiredLocked()
}

// evictExpiredLocked removes expired items (must be called with lock held)
func (c *memoryCache) evictExpiredLocked() {
	now := time.Now()
	for key, item := range c.items {
		if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
			delete(c.items, key)
		}
	}
}

// evictOldestLocked removes the oldest item (must be called with lock held)
func (c *memoryCache) evictOldestLocked() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, item := range c.items {
		if first || item.createdAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.createdAt
			first = false
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}
