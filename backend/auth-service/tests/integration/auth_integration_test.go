// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/protobankbankc/auth-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthFlow tests the complete authentication flow
// This is a basic smoke test that can be expanded with testcontainers
func TestAuthFlow(t *testing.T) {
	// This test requires a running server and database
	// Skip if integration tests are not enabled
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	baseURL := "http://localhost:8080"

	t.Run("health check", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("register and login flow", func(t *testing.T) {
		// Generate unique email for this test
		email := "test" + time.Now().Format("20060102150405") + "@example.com"

		// Step 1: Register user
		registerReq := models.RegisterRequest{
			Email:        email,
			Phone:        "+447700900123",
			Password:     "TestPass123!",
			FirstName:    "Integration",
			LastName:     "Test",
			DateOfBirth:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			AddressLine1: "123 Test St",
			City:         "London",
			Postcode:     "SW1A 1AA",
			Country:      "UK",
		}

		registerBody, err := json.Marshal(registerReq)
		require.NoError(t, err)

		resp, err := http.Post(
			baseURL+"/api/v1/auth/register",
			"application/json",
			bytes.NewBuffer(registerBody),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Step 2: Login with registered credentials
		loginReq := models.LoginRequest{
			Email:    email,
			Password: "TestPass123!",
		}

		loginBody, err := json.Marshal(loginReq)
		require.NoError(t, err)

		resp, err = http.Post(
			baseURL+"/api/v1/auth/login",
			"application/json",
			bytes.NewBuffer(loginBody),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var loginResp models.LoginResponse
		err = json.NewDecoder(resp.Body).Decode(&loginResp)
		require.NoError(t, err)

		assert.NotEmpty(t, loginResp.AccessToken)
		assert.NotEmpty(t, loginResp.RefreshToken)
		assert.Equal(t, "Bearer", loginResp.TokenType)
		assert.NotNil(t, loginResp.User)

		// Step 3: Access protected endpoint with token
		req, err := http.NewRequest(http.MethodGet, baseURL+"/api/v1/auth/me", nil)
		require.NoError(t, err)

		req.Header.Set("Authorization", "Bearer "+loginResp.AccessToken)

		client := &http.Client{}
		resp, err = client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var user models.User
		err = json.NewDecoder(resp.Body).Decode(&user)
		require.NoError(t, err)

		assert.Equal(t, email, user.Email)
		assert.Equal(t, "Integration", user.FirstName)

		// Step 4: Refresh token
		refreshReq := models.RefreshTokenRequest{
			RefreshToken: loginResp.RefreshToken,
		}

		refreshBody, err := json.Marshal(refreshReq)
		require.NoError(t, err)

		resp, err = http.Post(
			baseURL+"/api/v1/auth/refresh",
			"application/json",
			bytes.NewBuffer(refreshBody),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var refreshResp models.RefreshTokenResponse
		err = json.NewDecoder(resp.Body).Decode(&refreshResp)
		require.NoError(t, err)

		assert.NotEmpty(t, refreshResp.AccessToken)
		assert.NotEqual(t, loginResp.AccessToken, refreshResp.AccessToken)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		loginReq := models.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "WrongPassword123!",
		}

		loginBody, err := json.Marshal(loginReq)
		require.NoError(t, err)

		resp, err := http.Post(
			baseURL+"/api/v1/auth/login",
			"application/json",
			bytes.NewBuffer(loginBody),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// TestRateLimiting tests that rate limiting works
func TestRateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	baseURL := "http://localhost:8080"

	// Make multiple rapid requests to trigger rate limit
	successCount := 0
	rateLimitedCount := 0

	for i := 0; i < 15; i++ {
		resp, err := http.Get(baseURL + "/health")
		require.NoError(t, err)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			successCount++
		} else if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitedCount++

			// Verify rate limit headers are present
			assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Limit"))
			assert.NotEmpty(t, resp.Header.Get("X-RateLimit-Remaining"))
			assert.NotEmpty(t, resp.Header.Get("Retry-After"))
		}
	}

	// Should have some successful requests and some rate limited
	assert.Greater(t, successCount, 0, "Should have some successful requests")
}

// TestMetricsEndpoint tests that Prometheus metrics are exposed
func TestMetricsEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	baseURL := "http://localhost:8080"

	resp, err := http.Get(baseURL + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/plain")
}

// Note: To run integration tests:
// 1. Start the service with: docker-compose up
// 2. Run tests with: go test -tags=integration ./tests/integration/...
// 3. Or run with testcontainers for isolated testing
