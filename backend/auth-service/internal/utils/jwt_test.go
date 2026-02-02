package utils

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key-minimum-32-characters-long-for-security"

// TestGenerateAccessToken tests access token generation
func TestGenerateAccessToken(t *testing.T) {
	userID := uuid.New().String()
	email := "test@example.com"
	expiry := 15 * time.Minute

	tests := []struct {
		name    string
		userID  string
		email   string
		expiry  time.Duration
		secret  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid access token",
			userID:  userID,
			email:   email,
			expiry:  expiry,
			secret:  testSecret,
			wantErr: false,
		},
		{
			name:    "empty user ID",
			userID:  "",
			email:   email,
			expiry:  expiry,
			secret:  testSecret,
			wantErr: true,
			errMsg:  "user ID cannot be empty",
		},
		{
			name:    "empty email",
			userID:  userID,
			email:   "",
			expiry:  expiry,
			secret:  testSecret,
			wantErr: true,
			errMsg:  "email cannot be empty",
		},
		{
			name:    "empty secret",
			userID:  userID,
			email:   email,
			expiry:  expiry,
			secret:  "",
			wantErr: true,
			errMsg:  "secret cannot be empty",
		},
		{
			name:    "zero expiry",
			userID:  userID,
			email:   email,
			expiry:  0,
			secret:  testSecret,
			wantErr: true,
			errMsg:  "expiry must be positive",
		},
		{
			name:    "negative expiry",
			userID:  userID,
			email:   email,
			expiry:  -1 * time.Hour,
			secret:  testSecret,
			wantErr: true,
			errMsg:  "expiry must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateAccessToken(tt.userID, tt.email, tt.expiry, tt.secret)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Empty(t, token)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, token)

			// Token should have 3 parts (header.payload.signature)
			parts := strings.Split(token, ".")
			assert.Len(t, parts, 3)

			// Verify token can be parsed
			parsed, err := ValidateToken(token, tt.secret)
			require.NoError(t, err)
			assert.Equal(t, tt.userID, parsed.UserID)
			assert.Equal(t, tt.email, parsed.Email)
			assert.Equal(t, "access", parsed.TokenType)
		})
	}
}

// TestGenerateRefreshToken tests refresh token generation
func TestGenerateRefreshToken(t *testing.T) {
	userID := uuid.New().String()
	email := "test@example.com"
	expiry := 7 * 24 * time.Hour

	tests := []struct {
		name    string
		userID  string
		email   string
		expiry  time.Duration
		secret  string
		wantErr bool
	}{
		{
			name:    "valid refresh token",
			userID:  userID,
			email:   email,
			expiry:  expiry,
			secret:  testSecret,
			wantErr: false,
		},
		{
			name:    "empty user ID",
			userID:  "",
			email:   email,
			expiry:  expiry,
			secret:  testSecret,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateRefreshToken(tt.userID, tt.email, tt.expiry, tt.secret)

			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, token)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, token)

			// Verify token type is "refresh"
			parsed, err := ValidateToken(token, tt.secret)
			require.NoError(t, err)
			assert.Equal(t, "refresh", parsed.TokenType)
		})
	}
}

// TestValidateToken tests token validation
func TestValidateToken(t *testing.T) {
	userID := uuid.New().String()
	email := "test@example.com"

	validToken, err := GenerateAccessToken(userID, email, 15*time.Minute, testSecret)
	require.NoError(t, err)

	expiredToken, err := GenerateAccessToken(userID, email, -1*time.Hour, testSecret)
	require.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		secret  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid token",
			token:   validToken,
			secret:  testSecret,
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			secret:  testSecret,
			wantErr: true,
			errMsg:  "token cannot be empty",
		},
		{
			name:    "empty secret",
			token:   validToken,
			secret:  "",
			wantErr: true,
			errMsg:  "secret cannot be empty",
		},
		{
			name:    "invalid token format",
			token:   "invalid.token.format",
			secret:  testSecret,
			wantErr: true,
			errMsg:  "invalid token",
		},
		{
			name:    "malformed token",
			token:   "not-a-jwt-token",
			secret:  testSecret,
			wantErr: true,
			errMsg:  "invalid token",
		},
		{
			name:    "expired token",
			token:   expiredToken,
			secret:  testSecret,
			wantErr: true,
			errMsg:  "token has expired",
		},
		{
			name:    "wrong secret",
			token:   validToken,
			secret:  "wrong-secret-key-different-from-original",
			wantErr: true,
			errMsg:  "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.token, tt.secret)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, claims)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, claims)
			assert.Equal(t, userID, claims.UserID)
			assert.Equal(t, email, claims.Email)
		})
	}
}

// TestTokenExpiration tests token expiration behavior
func TestTokenExpiration(t *testing.T) {
	userID := uuid.New().String()
	email := "test@example.com"

	t.Run("token expires after duration", func(t *testing.T) {
		// Generate token that expires in 1 second
		token, err := GenerateAccessToken(userID, email, 1*time.Second, testSecret)
		require.NoError(t, err)

		// Should be valid immediately
		claims, err := ValidateToken(token, testSecret)
		require.NoError(t, err)
		assert.NotNil(t, claims)

		// Wait for expiration
		time.Sleep(2 * time.Second)

		// Should be expired now
		claims, err = ValidateToken(token, testSecret)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
		assert.Nil(t, claims)
	})

	t.Run("short-lived access token", func(t *testing.T) {
		token, err := GenerateAccessToken(userID, email, 15*time.Minute, testSecret)
		require.NoError(t, err)

		claims, err := ValidateToken(token, testSecret)
		require.NoError(t, err)
		assert.Equal(t, "access", claims.TokenType)
	})

	t.Run("long-lived refresh token", func(t *testing.T) {
		token, err := GenerateRefreshToken(userID, email, 7*24*time.Hour, testSecret)
		require.NoError(t, err)

		claims, err := ValidateToken(token, testSecret)
		require.NoError(t, err)
		assert.Equal(t, "refresh", claims.TokenType)
	})
}

// TestTokenSecurity tests security properties
func TestTokenSecurity(t *testing.T) {
	userID := uuid.New().String()
	email := "test@example.com"

	t.Run("different tokens for same user", func(t *testing.T) {
		token1, err1 := GenerateAccessToken(userID, email, 15*time.Minute, testSecret)
		token2, err2 := GenerateAccessToken(userID, email, 15*time.Minute, testSecret)

		require.NoError(t, err1)
		require.NoError(t, err2)

		// Tokens should be different due to different issued-at time
		assert.NotEqual(t, token1, token2)

		// But both should be valid
		claims1, err1 := ValidateToken(token1, testSecret)
		claims2, err2 := ValidateToken(token2, testSecret)

		require.NoError(t, err1)
		require.NoError(t, err2)

		assert.Equal(t, claims1.UserID, claims2.UserID)
		assert.Equal(t, claims1.Email, claims2.Email)
	})

	t.Run("tampering detection", func(t *testing.T) {
		token, err := GenerateAccessToken(userID, email, 15*time.Minute, testSecret)
		require.NoError(t, err)

		// Tamper with token by modifying payload
		parts := strings.Split(token, ".")
		require.Len(t, parts, 3)

		// Change middle part (payload)
		parts[1] = "tamperedpayload"
		tamperedToken := strings.Join(parts, ".")

		// Should fail validation
		claims, err := ValidateToken(tamperedToken, testSecret)
		require.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("signature verification", func(t *testing.T) {
		token, err := GenerateAccessToken(userID, email, 15*time.Minute, testSecret)
		require.NoError(t, err)

		// Change signature
		parts := strings.Split(token, ".")
		require.Len(t, parts, 3)

		parts[2] = "invalidsignature"
		invalidToken := strings.Join(parts, ".")

		// Should fail validation
		claims, err := ValidateToken(invalidToken, testSecret)
		require.Error(t, err)
		assert.Nil(t, claims)
	})
}

// TestExtractTokenFromHeader tests extracting token from Authorization header
func TestExtractTokenFromHeader(t *testing.T) {
	validToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.abc123"

	tests := []struct {
		name    string
		header  string
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid Bearer token",
			header:  "Bearer " + validToken,
			want:    validToken,
			wantErr: false,
		},
		{
			name:    "empty header",
			header:  "",
			want:    "",
			wantErr: true,
			errMsg:  "authorization header is empty",
		},
		{
			name:    "missing Bearer prefix",
			header:  validToken,
			want:    "",
			wantErr: true,
			errMsg:  "invalid authorization header format",
		},
		{
			name:    "wrong scheme",
			header:  "Basic " + validToken,
			want:    "",
			wantErr: true,
			errMsg:  "invalid authorization header format",
		},
		{
			name:    "Bearer with no token",
			header:  "Bearer ",
			want:    "",
			wantErr: true,
			errMsg:  "token is empty",
		},
		{
			name:    "extra spaces",
			header:  "Bearer  " + validToken,
			want:    validToken,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ExtractTokenFromHeader(tt.header)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Empty(t, token)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, token)
		})
	}
}

// TestTokenClaims tests token claims structure
func TestTokenClaims(t *testing.T) {
	userID := uuid.New().String()
	email := "test@example.com"

	token, err := GenerateAccessToken(userID, email, 15*time.Minute, testSecret)
	require.NoError(t, err)

	// Parse token manually to inspect claims
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	require.NoError(t, err)
	require.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// Verify required claims exist
	assert.Contains(t, claims, "user_id")
	assert.Contains(t, claims, "email")
	assert.Contains(t, claims, "token_type")
	assert.Contains(t, claims, "exp")
	assert.Contains(t, claims, "iat")
	assert.Contains(t, claims, "nbf")

	// Verify claim values
	assert.Equal(t, userID, claims["user_id"])
	assert.Equal(t, email, claims["email"])
	assert.Equal(t, "access", claims["token_type"])
}

// BenchmarkGenerateAccessToken benchmarks token generation
func BenchmarkGenerateAccessToken(b *testing.B) {
	userID := uuid.New().String()
	email := "test@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateAccessToken(userID, email, 15*time.Minute, testSecret)
	}
}

// BenchmarkValidateToken benchmarks token validation
func BenchmarkValidateToken(b *testing.B) {
	userID := uuid.New().String()
	email := "test@example.com"
	token, _ := GenerateAccessToken(userID, email, 15*time.Minute, testSecret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidateToken(token, testSecret)
	}
}

// TestTokenTypeValidation tests token type checking
func TestTokenTypeValidation(t *testing.T) {
	userID := uuid.New().String()
	email := "test@example.com"

	accessToken, err := GenerateAccessToken(userID, email, 15*time.Minute, testSecret)
	require.NoError(t, err)

	refreshToken, err := GenerateRefreshToken(userID, email, 7*24*time.Hour, testSecret)
	require.NoError(t, err)

	t.Run("access token has correct type", func(t *testing.T) {
		claims, err := ValidateToken(accessToken, testSecret)
		require.NoError(t, err)
		assert.Equal(t, "access", claims.TokenType)
	})

	t.Run("refresh token has correct type", func(t *testing.T) {
		claims, err := ValidateToken(refreshToken, testSecret)
		require.NoError(t, err)
		assert.Equal(t, "refresh", claims.TokenType)
	})

	t.Run("cannot use refresh token as access token", func(t *testing.T) {
		// This would be enforced in the middleware/service layer
		// The token itself is valid, but type should be checked
		claims, err := ValidateToken(refreshToken, testSecret)
		require.NoError(t, err)
		assert.NotEqual(t, "access", claims.TokenType)
	})
}
