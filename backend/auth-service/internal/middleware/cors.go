package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Authorization",
			"Accept",
			"Origin",
			"Cache-Control",
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		AllowCredentials: true,
		MaxAge:           43200, // 12 hours
	}
}

// ProductionCORSConfig returns CORS configuration for production
func ProductionCORSConfig(allowedOrigins []string) *CORSConfig {
	config := DefaultCORSConfig()
	config.AllowOrigins = allowedOrigins
	return config
}

// CORS returns a CORS middleware with the given configuration
func CORS(config *CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		if config.AllowOrigins[0] == "*" || contains(config.AllowOrigins, origin) {
			// Set allowed origin
			if config.AllowOrigins[0] == "*" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// Set other CORS headers
			if config.AllowCredentials {
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Set allowed methods
			methods := ""
			for i, method := range config.AllowMethods {
				if i > 0 {
					methods += ", "
				}
				methods += method
			}
			c.Writer.Header().Set("Access-Control-Allow-Methods", methods)

			// Set allowed headers
			headers := ""
			for i, header := range config.AllowHeaders {
				if i > 0 {
					headers += ", "
				}
				headers += header
			}
			c.Writer.Header().Set("Access-Control-Allow-Headers", headers)

			// Set exposed headers
			if len(config.ExposeHeaders) > 0 {
				exposeHeaders := ""
				for i, header := range config.ExposeHeaders {
					if i > 0 {
						exposeHeaders += ", "
					}
					exposeHeaders += header
				}
				c.Writer.Header().Set("Access-Control-Expose-Headers", exposeHeaders)
			}

			// Set max age
			if config.MaxAge > 0 {
				c.Writer.Header().Set("Access-Control-Max-Age", string(rune(config.MaxAge)))
			}
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
