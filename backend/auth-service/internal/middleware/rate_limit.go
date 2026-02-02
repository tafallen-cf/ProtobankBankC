package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu      sync.RWMutex
	clients map[string]*client
	limit   int
	window  time.Duration
}

// client represents a rate limit client
type client struct {
	tokens    int
	lastReset time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	limiter := &RateLimiter{
		clients: make(map[string]*client),
		limit:   limit,
		window:  window,
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// Limit returns the rate limiting middleware
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		ip := getClientIP(c)

		// Check rate limit
		allowed, remaining, resetTime := rl.allow(ip)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		if !allowed {
			retryAfter := time.Until(resetTime).Seconds()
			c.Header("Retry-After", fmt.Sprintf("%.0f", retryAfter))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate limit exceeded",
				"message": fmt.Sprintf("Too many requests. Please try again in %.0f seconds.", retryAfter),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allow checks if a request is allowed for the given IP
func (rl *RateLimiter) allow(ip string) (bool, int, time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Get or create client
	cl, exists := rl.clients[ip]
	if !exists {
		cl = &client{
			tokens:    rl.limit,
			lastReset: now,
		}
		rl.clients[ip] = cl
	}

	// Check if window has expired
	if now.Sub(cl.lastReset) > rl.window {
		cl.tokens = rl.limit
		cl.lastReset = now
	}

	// Check if request is allowed
	if cl.tokens > 0 {
		cl.tokens--
		resetTime := cl.lastReset.Add(rl.window)
		return true, cl.tokens, resetTime
	}

	resetTime := cl.lastReset.Add(rl.window)
	return false, 0, resetTime
}

// cleanup removes expired clients from memory
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()

		for ip, cl := range rl.clients {
			if now.Sub(cl.lastReset) > rl.window*2 {
				delete(rl.clients, ip)
			}
		}

		rl.mu.Unlock()
	}
}

// getClientIP extracts the client IP from the request
// It checks X-Forwarded-For and X-Real-IP headers for proxy support
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	xri := c.GetHeader("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}

	return ip
}
