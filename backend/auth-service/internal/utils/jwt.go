package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenClaims represents the claims stored in JWT tokens
type TokenClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"` // "access" or "refresh"
}

// customClaims extends jwt.RegisteredClaims with our custom fields
type customClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a new JWT access token
func GenerateAccessToken(userID, email string, expiry time.Duration, secret string) (string, error) {
	return generateToken(userID, email, "access", expiry, secret)
}

// GenerateRefreshToken generates a new JWT refresh token
func GenerateRefreshToken(userID, email string, expiry time.Duration, secret string) (string, error) {
	return generateToken(userID, email, "refresh", expiry, secret)
}

// generateToken creates a JWT token with the specified parameters
func generateToken(userID, email, tokenType string, expiry time.Duration, secret string) (string, error) {
	// Validate inputs
	if userID == "" {
		return "", fmt.Errorf("user ID cannot be empty")
	}

	if email == "" {
		return "", fmt.Errorf("email cannot be empty")
	}

	if secret == "" {
		return "", fmt.Errorf("secret cannot be empty")
	}

	if expiry <= 0 {
		return "", fmt.Errorf("expiry must be positive")
	}

	// Create claims
	now := time.Now()
	claims := customClaims{
		UserID:    userID,
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString, secret string) (*TokenClaims, error) {
	// Validate inputs
	if tokenString == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	if secret == "" {
		return nil, fmt.Errorf("secret cannot be empty")
	}

	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &customClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		// Check for specific error types
		if strings.Contains(err.Error(), "token has expired") ||
		   strings.Contains(err.Error(), "token is expired") {
			return nil, fmt.Errorf("token has expired")
		}
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(*customClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Return simplified claims
	return &TokenClaims{
		UserID:    claims.UserID,
		Email:     claims.Email,
		TokenType: claims.TokenType,
	}, nil
}

// ExtractTokenFromHeader extracts the JWT token from the Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	// Validate header
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is empty")
	}

	// Split header into parts
	parts := strings.Fields(authHeader)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid authorization header format")
	}

	// Verify Bearer scheme
	if parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format: expected Bearer scheme")
	}

	// Extract token
	token := parts[1]
	if token == "" {
		return "", fmt.Errorf("token is empty")
	}

	return token, nil
}

// GetTokenExpiry returns the expiration time from a token string
func GetTokenExpiry(tokenString, secret string) (*time.Time, error) {
	token, err := jwt.ParseWithClaims(tokenString, &customClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	if claims.ExpiresAt == nil {
		return nil, fmt.Errorf("token has no expiration")
	}

	expiryTime := claims.ExpiresAt.Time
	return &expiryTime, nil
}
