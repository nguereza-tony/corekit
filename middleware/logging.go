package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nguereza-tony/corekit/common"
	"github.com/nguereza-tony/corekit/logger"
)

// LoggingMiddleware logs HTTP requests with method, path, status, duration
func LoggingMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Get request ID
		requestID := common.GetRequestID(c)

		// Get request info
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Get response status
		status := c.Writer.Status()

		// Get error if any
		var errorMsg string
		if len(c.Errors) > 0 {
			errorMsg = c.Errors.Last().Error()
		}

		// Log based on status code
		switch {
		case status >= 500:
			log.Errorf("Request - RequestID: %s, Method: %s, Path: %s, Status: %d, Duration: %v, ClientIP: %s, UserAgent: %s, Error: %s",
				requestID, method, path, status, duration, clientIP, userAgent, errorMsg)
		case status >= 400:
			log.Warnf("Request - RequestID: %s, Method: %s, Path: %s, Status: %d, Duration: %v, ClientIP: %s, UserAgent: %s",
				requestID, method, path, status, duration, clientIP, userAgent)
		default:
			log.Debugf("Request - RequestID: %s, Method: %s, Path: %s, Status: %d, Duration: %v, ClientIP: %s, UserAgent: %s",
				requestID, method, path, status, duration, clientIP, userAgent)
		}
	}
}
