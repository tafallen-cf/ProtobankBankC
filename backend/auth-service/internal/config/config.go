package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the auth service
type Config struct {
	ServiceName string
	ServicePort string
	LogLevel    string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTSecret           string
	JWTExpiry           time.Duration
	RefreshTokenExpiry  time.Duration

	// Security
	BcryptCost int

	// Rate Limiting
	RateLimitEnabled           bool
	RateLimitRequestsPerMinute int

	// CORS
	CORSOrigins     []string
	CORSCredentials bool

	// Session
	SessionTimeout time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	viper.AutomaticEnv()

	viper.SetDefault("SERVICE_NAME", "auth-service")
	viper.SetDefault("SERVICE_PORT", "3001")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("BCRYPT_COST", 12)
	viper.SetDefault("JWT_EXPIRY", "15m")
	viper.SetDefault("REFRESH_TOKEN_EXPIRY", "168h")
	viper.SetDefault("RATE_LIMIT_ENABLED", true)
	viper.SetDefault("RATE_LIMIT_REQUESTS_PER_MINUTE", 5)
	viper.SetDefault("SESSION_TIMEOUT", "30m")

	jwtExpiry, err := time.ParseDuration(viper.GetString("JWT_EXPIRY"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY: %w", err)
	}

	refreshTokenExpiry, err := time.ParseDuration(viper.GetString("REFRESH_TOKEN_EXPIRY"))
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN_EXPIRY: %w", err)
	}

	sessionTimeout, err := time.ParseDuration(viper.GetString("SESSION_TIMEOUT"))
	if err != nil {
		return nil, fmt.Errorf("invalid SESSION_TIMEOUT: %w", err)
	}

	config := &Config{
		ServiceName: viper.GetString("SERVICE_NAME"),
		ServicePort: viper.GetString("SERVICE_PORT"),
		LogLevel:    viper.GetString("LOG_LEVEL"),

		DatabaseURL: viper.GetString("DATABASE_URL"),
		RedisURL:    viper.GetString("REDIS_URL"),

		JWTSecret:          viper.GetString("JWT_SECRET"),
		JWTExpiry:          jwtExpiry,
		RefreshTokenExpiry: refreshTokenExpiry,

		BcryptCost: viper.GetInt("BCRYPT_COST"),

		RateLimitEnabled:           viper.GetBool("RATE_LIMIT_ENABLED"),
		RateLimitRequestsPerMinute: viper.GetInt("RATE_LIMIT_REQUESTS_PER_MINUTE"),

		CORSOrigins:     viper.GetStringSlice("CORS_ORIGINS"),
		CORSCredentials: viper.GetBool("CORS_CREDENTIALS"),

		SessionTimeout: sessionTimeout,
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if c.RedisURL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}

	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	if c.BcryptCost < 10 || c.BcryptCost > 14 {
		return fmt.Errorf("BCRYPT_COST must be between 10 and 14")
	}

	return nil
}
