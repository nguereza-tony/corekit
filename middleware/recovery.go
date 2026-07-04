package middleware

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nguereza-tony/corekit/common"
	"github.com/nguereza-tony/corekit/logger"
	"github.com/nguereza-tony/corekit/response"
)

// RecoveryMiddleware recovers from panics and returns a 500 Internal Server Error
func RecoveryMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		defer func() {
			if r := recover(); r != nil {
				// Calculate request duration
				duration := time.Since(startTime)

				// Get request info
				requestID := common.GetRequestID(c)
				method := c.Request.Method
				path := c.Request.URL.Path
				clientIP := c.ClientIP()
				userAgent := c.GetHeader("User-Agent")

				// Get stack trace
				stack := debug.Stack()

				// Log detailed error information
				log.Errorf("Panic recovered - RequestID: %s, Method: %s, Path: %s, ClientIP: %s, UserAgent: %s, Duration: %v, Error: %v, Stack: %s",
					requestID, method, path, clientIP, userAgent, duration, r, string(stack))

				// Determine if it's a known error type
				var message string
				switch err := r.(type) {
				case error:
					message = err.Error()
				case string:
					message = err
				default:
					message = fmt.Sprintf("%v", err)
				}

				// Return appropriate error response
				log.Error(message)
				response.InternalServerError(c, "", "internal server error")

				// Stop further handlers
				c.Abort()
			}
		}()

		c.Next()
	}
}
