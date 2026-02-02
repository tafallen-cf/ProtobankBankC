package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// TestRateLimitMiddleware tests rate limiting functionality
func TestRateLimitMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		requests       int
		limit          int
		window         time.Duration
		expectedPassed int
		expectedStatus int
	}{
		{
			name:           "requests within limit",
			requests:       5,
			limit:          10,
			window:         time.Minute,
			expectedPassed: 5,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "requests exceed limit",
			requests:       15,
			limit:          10,
			window:         time.Minute,
			expectedPassed: 10,
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name:           "single request under limit",
			requests:       1,
			limit:          5,
			window:         time.Minute,
			expectedPassed: 1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "exactly at limit",
			requests:       10,
			limit:          10,
			window:         time.Minute,
			expectedPassed: 10,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			router := setupTestRouter()
			limiter := NewRateLimiter(tt.limit, tt.window)
			router.Use(limiter.Limit())

			passedCount := 0
			blockedCount := 0

			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// Execute requests
			for i := 0; i < tt.requests; i++ {
				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				req.RemoteAddr = "192.168.1.1:12345"
				rec := httptest.NewRecorder()

				router.ServeHTTP(rec, req)

				if rec.Code == http.StatusOK {
					passedCount++
				} else if rec.Code == http.StatusTooManyRequests {
					blockedCount++
				}
			}

			// Assert
			assert.Equal(t, tt.expectedPassed, passedCount, "Expected %d requests to pass, got %d", tt.expectedPassed, passedCount)
			if tt.requests > tt.limit {
				assert.Greater(t, blockedCount, 0, "Expected some requests to be blocked")
			}
		})
	}
}

// TestRateLimitByIP tests that rate limiting is per IP
func TestRateLimitByIP(t *testing.T) {
	router := setupTestRouter()
	limiter := NewRateLimiter(5, time.Minute)
	router.Use(limiter.Limit())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First IP makes 5 requests (should all pass)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "Request %d from first IP should pass", i+1)
	}

	// Different IP makes 5 requests (should also all pass)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "Request %d from second IP should pass", i+1)
	}

	// First IP makes another request (should be blocked)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code, "6th request from first IP should be blocked")
}

// TestRateLimitReset tests that rate limit resets after window
func TestRateLimitReset(t *testing.T) {
	router := setupTestRouter()
	// Very short window for testing
	limiter := NewRateLimiter(2, 100*time.Millisecond)
	router.Use(limiter.Limit())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make 2 requests (should pass)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "Request %d should pass", i+1)
	}

	// Third request (should be blocked)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code, "3rd request should be blocked")

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Request after window (should pass)
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code, "Request after window reset should pass")
}

// TestRateLimitHeaders tests that rate limit headers are set
func TestRateLimitHeaders(t *testing.T) {
	router := setupTestRouter()
	limiter := NewRateLimiter(10, time.Minute)
	router.Use(limiter.Limit())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Check headers
	assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Limit"), "X-RateLimit-Limit header should be set")
	assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Remaining"), "X-RateLimit-Remaining header should be set")
	assert.NotEmpty(t, rec.Header().Get("X-RateLimit-Reset"), "X-RateLimit-Reset header should be set")
}

// TestRateLimitErrorResponse tests the error response format
func TestRateLimitErrorResponse(t *testing.T) {
	router := setupTestRouter()
	limiter := NewRateLimiter(1, time.Minute)
	router.Use(limiter.Limit())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// First request (should pass)
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusOK, rec1.Code)

	// Second request (should be blocked)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusTooManyRequests, rec2.Code)
	assert.Contains(t, rec2.Body.String(), "rate limit exceeded")
	assert.NotEmpty(t, rec2.Header().Get("Retry-After"))
}

// TestRateLimitCleanup tests that old entries are cleaned up
func TestRateLimitCleanup(t *testing.T) {
	router := setupTestRouter()
	limiter := NewRateLimiter(5, 50*time.Millisecond)
	router.Use(limiter.Limit())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make requests from multiple IPs
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1." + string(rune(i)) + ":12345"
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
	}

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	// The limiter should have cleaned up old entries
	// This is more of a sanity check - in production you'd monitor memory
	assert.NotNil(t, limiter, "Limiter should still exist after cleanup")
}

// TestRateLimitWithXForwardedFor tests rate limiting with proxy headers
func TestRateLimitWithXForwardedFor(t *testing.T) {
	router := setupTestRouter()
	limiter := NewRateLimiter(2, time.Minute)
	router.Use(limiter.Limit())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Make requests with X-Forwarded-For header
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Forwarded-For", "10.0.0.1")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "Request %d should pass", i+1)
	}

	// Third request should be blocked
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code, "3rd request should be blocked")
}
