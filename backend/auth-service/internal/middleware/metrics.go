package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request counter
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP request duration histogram
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// HTTP request size histogram
	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	// HTTP response size histogram
	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path", "status"},
	)

	// Active requests gauge
	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being served",
		},
	)
)

// Metrics returns a Prometheus metrics middleware
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Increment in-flight requests
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		// Start timer
		start := time.Now()

		// Get request size
		requestSize := computeApproximateRequestSize(c.Request)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get response info
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath()

		// If path is empty (404), use the request path
		if path == "" {
			path = c.Request.URL.Path
		}

		// Record metrics
		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path, status).Observe(duration)
		httpRequestSize.WithLabelValues(method, path).Observe(float64(requestSize))
		httpResponseSize.WithLabelValues(method, path, status).Observe(float64(c.Writer.Size()))
	}
}

// computeApproximateRequestSize calculates approximate request size
func computeApproximateRequestSize(c *gin.Context) int {
	s := 0

	// Method
	s += len(c.Request.Method)

	// URL
	if c.Request.URL != nil {
		s += len(c.Request.URL.String())
	}

	// Proto
	s += len(c.Request.Proto)

	// Headers (approximate)
	for name, values := range c.Request.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}

	return s
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled    bool
	Path       string
	Namespace  string
	Subsystem  string
	MetricPath string
}

// DefaultMetricsConfig returns default metrics configuration
func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		Enabled:    true,
		Path:       "/metrics",
		Namespace:  "auth_service",
		Subsystem:  "http",
		MetricPath: "/metrics",
	}
}
