package cache

import (
	"fmt"
)

// NewCache creates a new cache instance based on configuration
func NewCache(cfg *CacheConfig) (Cache, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cache configuration is nil")
	}

	switch cfg.Type {
	case "redis":
		return NewRedisCache(cfg)
	case "memory":
		memoryCfg := &MemoryCacheConfig{
			CleanupInterval: cfg.Memory.CleanupInterval,
			MaxItems:        cfg.Memory.MaxItems,
		}
		return NewMemoryCache(memoryCfg), nil
	default:
		return NewDefaultCache(), nil
	}
}

// NewDefaultCache creates a new memory cache with default configuration
func NewDefaultCache() Cache {
	return NewMemoryCache(&MemoryCacheConfig{
		CleanupInterval: 60,
		MaxItems:        10000,
	})
}
