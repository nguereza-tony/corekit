package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nguereza-tony/corekit/cache"
	"github.com/nguereza-tony/corekit/logger"
	"github.com/nguereza-tony/corekit/response"
)

// RateLimiter handles rate limiting using Redis
type RateLimiter struct {
	cacheService *cache.CacheService
	log          *logger.Logger
	rpm          int // requests per minute
	burst        int // burst allowed
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cacheService *cache.CacheService, rpm, burst int, log *logger.Logger) *RateLimiter {
	return &RateLimiter{
		cacheService: cacheService,
		log:          log,
		rpm:          rpm,
		burst:        burst,
	}
}

// Allow checks if a request is allowed for a given key
func (r *RateLimiter) Allow(ctx context.Context, key string) (bool, int, error) {
	now := time.Now().Unix()
	window := now - (now % 60) // 60 second window

	// Create sliding window key
	windowKey := fmt.Sprintf("rate_limit:%s:%d", key, window)

	// Increment counter
	count, err := r.cacheService.Increment(ctx, windowKey)
	if err != nil {
		return false, 0, err
	}

	// Set TTL for the key (2 minutes)
	if count == 1 {
		r.cacheService.Expire(ctx, windowKey, 2*time.Minute)
	}

	// Check if within limits (rpm + burst)
	limit := int64(r.rpm + r.burst)
	if count > limit {
		return false, int(limit), nil
	}

	// Calculate remaining allowed requests
	remaining := int(limit - count)
	if remaining < 0 {
		remaining = 0
	}

	return true, remaining, nil
}

// GetRateLimitKey generates a rate limit key based on request
func GetRateLimitKey(c *gin.Context) string {
	// Priority: API key > Client IP
	if apiKey := c.GetHeader("Authorization"); apiKey != "" {
		return "apikey:" + apiKey
	}

	// Fallback to client IP
	return "ip:" + c.ClientIP()
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(rateLimiter *RateLimiter, enabled bool) gin.HandlerFunc {
	if !enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Generate rate limit key
		key := GetRateLimitKey(c)

		// Check rate limit
		allowed, remaining, err := rateLimiter.Allow(c.Request.Context(), key)
		if err != nil {
			rateLimiter.log.Errorf("Rate limit check failed: %v", err)
			// On error, allow the request
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimiter.rpm+rateLimiter.burst))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Unix()+(60-time.Now().Unix()%60)))

		if !allowed {
			response.TooManyRequests(c, "too many requests, please try again later")
			return
		}

		c.Next()
	}
}
