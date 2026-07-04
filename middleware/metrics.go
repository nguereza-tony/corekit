package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nguereza-tony/corekit/metrics"
)

// PrometheusMiddleware records HTTP metrics
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		method := c.Request.Method
		endpoint := c.FullPath()

		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		metrics.RecordHTTPRequest(method, endpoint, status, duration)
	}
}
