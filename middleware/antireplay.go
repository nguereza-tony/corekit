package middleware

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nguereza-tony/corekit/cache"
	"github.com/nguereza-tony/corekit/config"
	"github.com/nguereza-tony/corekit/logger"
	"github.com/nguereza-tony/corekit/response"
)

// AntiReplayMiddleware prevents replay attacks using nonce and timestamp
func AntiReplayMiddleware(
	cacheService *cache.CacheService,
	cfg config.AntiReplayConfig,
	log *logger.Logger,
) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Get nonce and timestamp from headers
		nonce := c.GetHeader("X-Request-Nonce")
		timestampStr := c.GetHeader("X-Request-Timestamp")

		// Skip if headers are missing
		if nonce == "" || timestampStr == "" {
			c.Next()
			return
		}

		// Validate timestamp
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			response.BadRequest(c, "ANTIREPLAY_INVALID_TIMESTAMP_FORMAT", "invalid timestamp format")
			c.Abort()
			return
		}

		// Check timestamp tolerance
		now := time.Now().Unix()
		diff := now - timestamp
		if diff < 0 {
			diff = -diff
		}

		if diff > int64(cfg.TimestampTolerance*60) {
			response.BadRequest(c, "ANTIREPLAY_TIMESTAMP_OUT_TOLERANCE", "timestamp out of tolerance")
			c.Abort()
			return
		}

		// Check if nonce has been used
		key := fmt.Sprintf("nonce:%s", nonce)
		exists, err := cacheService.Exists(c.Request.Context(), key)
		if err != nil {
			log.Errorf("Failed to check nonce: %v", err)
			c.Next()
			return
		}

		if exists {
			response.BadRequest(c, "ANTIREPLAY_NONCE_USED", "nonce already used")
			c.Abort()
			return
		}

		// Store nonce
		if err := cacheService.Set(c.Request.Context(), key, timestamp, time.Duration(cfg.CacheTTL)*time.Second); err != nil {
			log.Errorf("Failed to store nonce: %v", err)
		}

		c.Next()
	}
}
