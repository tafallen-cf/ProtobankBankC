package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/protobankbankc/auth-service/internal/config"
	"github.com/protobankbankc/auth-service/internal/handlers"
	"github.com/protobankbankc/auth-service/internal/middleware"
	"github.com/protobankbankc/auth-service/internal/repository"
	"github.com/protobankbankc/auth-service/internal/services"
	"github.com/sirupsen/logrus"
)

const version = "1.0.0"

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Set Gin mode to release if not in development
	gin.SetMode(gin.ReleaseMode)

	// Initialize database connection
	dbPool, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbPool.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(dbPool)

	// Initialize services
	authService := services.NewAuthService(
		userRepo,
		cfg.JWTSecret,
		cfg.JWTExpiry,
		cfg.RefreshTokenExpiry,
	)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	healthHandler := handlers.NewHealthHandler(version)

	// Initialize logger
	logger := middleware.NewLogger("production")

	// Setup router
	router := setupRouter(cfg, authHandler, healthHandler, logger)

	// Create server
	server := &http.Server{
		Addr:           ":" + cfg.ServicePort,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting Auth Service v%s on port %s", version, cfg.ServicePort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped successfully")
}

// initDatabase initializes the database connection pool
func initDatabase(cfg *config.Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse config using the DATABASE_URL
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Set connection pool settings
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	// Create pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established successfully")
	return pool, nil
}

// setupRouter configures the HTTP router with all routes and middleware
func setupRouter(cfg *config.Config, authHandler *handlers.AuthHandler, healthHandler *handlers.HealthHandler, logger interface{}) *gin.Engine {
	router := gin.New()

	// Recovery middleware (must be first)
	router.Use(gin.Recovery())

	// Structured logging middleware
	router.Use(middleware.Logger(logger.(*logrus.Logger)))

	// Prometheus metrics middleware
	router.Use(middleware.Metrics())

	// CORS middleware
	corsConfig := middleware.DefaultCORSConfig()
	if len(cfg.CORSOrigins) > 0 {
		// Use configured CORS origins
		corsConfig = middleware.ProductionCORSConfig(cfg.CORSOrigins)
	}
	router.Use(middleware.CORS(corsConfig))

	// Rate limiting middleware (10 requests per minute per IP)
	rateLimiter := middleware.NewRateLimiter(10, time.Minute)
	router.Use(rateLimiter.Limit())

	// Health check routes (no auth required, no rate limiting)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)
	router.GET("/live", healthHandler.Live)

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/me", authHandler.GetMe) // Requires auth header
		}
	}

	return router
}
