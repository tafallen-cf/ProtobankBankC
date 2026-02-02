package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	startTime time.Time
	version   string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
		version:   version,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string        `json:"status"`
	Service   string        `json:"service"`
	Version   string        `json:"version"`
	Uptime    string        `json:"uptime"`
	Timestamp time.Time     `json:"timestamp"`
}

// Health returns the service health status
// GET /health
func (h *HealthHandler) Health(c *gin.Context) {
	uptime := time.Since(h.startTime)

	response := HealthResponse{
		Status:    "healthy",
		Service:   "auth-service",
		Version:   h.version,
		Uptime:    uptime.String(),
		Timestamp: time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}

// Ready returns readiness status (used by Kubernetes)
// GET /ready
func (h *HealthHandler) Ready(c *gin.Context) {
	// In a production system, you would check:
	// - Database connectivity
	// - Redis connectivity
	// - Any critical dependencies

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// Live returns liveness status (used by Kubernetes)
// GET /live
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}
