package cache

import (
	"fmt"

	"github.com/nguereza-tony/corekit/config"
)

// NewCache creates a new cache instance based on configuration
func NewCache(cfg *config.CacheConfig) (Cache, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cache configuration is nil")
	}

	switch cfg.Type {
	case "redis":
		return NewRedisCache(cfg)
	case "memory":
		return NewMemoryCache(&cfg.Memory), nil
	default:
		return NewDefaultCache(), nil
	}
}

// NewDefaultCache creates a new memory cache with default configuration
func NewDefaultCache() Cache {
	return NewMemoryCache(&config.MemoryCacheConfig{
		CleanupInterval: 60,
		MaxItems:        10000,
	})
}
