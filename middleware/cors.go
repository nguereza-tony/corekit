package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nguereza-tony/corekit/config"
)

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS)
func CORSMiddleware(cfg config.CORSConfig) gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     cfg.AllowedMethods,
		AllowHeaders:     cfg.AllowedHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           time.Duration(cfg.MaxAge) * time.Second,
	}

	// If no origins are specified, default to all origins
	if len(cfg.AllowedOrigins) == 0 {
		config.AllowAllOrigins = true
	}

	return cors.New(config)
}
