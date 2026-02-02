package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthHandler tests the health endpoint
func TestHealthHandler(t *testing.T) {
	// Setup
	handler := NewHealthHandler("1.0.0")
	router := setupTestRouter()
	router.GET("/health", handler.Health)

	// Wait a bit to ensure non-zero uptime
	time.Sleep(10 * time.Millisecond)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)

	var response HealthResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "auth-service", response.Service)
	assert.Equal(t, "1.0.0", response.Version)
	assert.NotEmpty(t, response.Uptime)
	assert.False(t, response.Timestamp.IsZero())

	// Check that uptime is reasonable
	assert.Contains(t, response.Uptime, "ms")
}

// TestReadyHandler tests the readiness endpoint
func TestReadyHandler(t *testing.T) {
	// Setup
	handler := NewHealthHandler("1.0.0")
	router := setupTestRouter()
	router.GET("/ready", handler.Ready)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "ready", response["status"])
}

// TestLiveHandler tests the liveness endpoint
func TestLiveHandler(t *testing.T) {
	// Setup
	handler := NewHealthHandler("1.0.0")
	router := setupTestRouter()
	router.GET("/live", handler.Live)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	rec := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "alive", response["status"])
}

// TestHealthHandlerMultipleCalls tests that uptime increases
func TestHealthHandlerMultipleCalls(t *testing.T) {
	// Setup
	handler := NewHealthHandler("1.0.0")
	router := setupTestRouter()
	router.GET("/health", handler.Health)

	// First call
	req1 := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)

	var response1 HealthResponse
	err := json.Unmarshal(rec1.Body.Bytes(), &response1)
	require.NoError(t, err)

	// Wait
	time.Sleep(50 * time.Millisecond)

	// Second call
	req2 := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)

	var response2 HealthResponse
	err = json.Unmarshal(rec2.Body.Bytes(), &response2)
	require.NoError(t, err)

	// Uptime should have increased
	assert.NotEqual(t, response1.Uptime, response2.Uptime)
	assert.True(t, response2.Timestamp.After(response1.Timestamp))
}
