package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger returns a structured logging middleware
func Logger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Get client IP
		clientIP := c.ClientIP()

		// Build log entry
		entry := logger.WithFields(logrus.Fields{
			"status":     statusCode,
			"method":     c.Request.Method,
			"path":       path,
			"query":      raw,
			"ip":         clientIP,
			"user_agent": c.Request.UserAgent(),
			"latency":    latency,
			"latency_ms": latency.Milliseconds(),
		})

		// Add error if present
		if len(c.Errors) > 0 {
			entry = entry.WithField("errors", c.Errors.String())
		}

		// Log based on status code
		if statusCode >= 500 {
			entry.Error("Server error")
		} else if statusCode >= 400 {
			entry.Warn("Client error")
		} else {
			entry.Info("Request completed")
		}
	}
}

// NewLogger creates a new configured logger instance
func NewLogger(environment string) *logrus.Logger {
	logger := logrus.New()

	// Set log format based on environment
	if environment == "production" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// Set log level
	if environment == "development" {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	return logger
}
