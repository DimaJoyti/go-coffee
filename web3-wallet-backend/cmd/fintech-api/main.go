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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/internal/accounts"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/config"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/redis"
)

// FintechAPIServer represents the main fintech API server
type FintechAPIServer struct {
	config     *config.Config
	logger     *logger.Logger
	db         *sqlx.DB
	cache      redis.Client
	httpServer *http.Server
	
	// Services
	accountsService accounts.Service
	
	// Handlers
	accountsHandler *accounts.Handler
}

func main() {
	fmt.Println("Starting Fintech API Server...")

	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := logger.New("fintech-api")
	defer logger.Sync()

	// Create server instance
	server := &FintechAPIServer{
		config: cfg,
		logger: logger,
	}

	// Initialize server
	if err := server.initialize(); err != nil {
		logger.Fatal("Failed to initialize server", zap.Error(err))
	}

	// Start server
	if err := server.start(); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}

	// Wait for shutdown signal
	server.waitForShutdown()

	// Graceful shutdown
	if err := server.shutdown(); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}

	logger.Info("Fintech API Server stopped")
}

// initialize initializes all server components
func (s *FintechAPIServer) initialize() error {
	s.logger.Info("Initializing Fintech API Server...")

	// Initialize database
	if err := s.initDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize Redis cache
	if err := s.initCache(); err != nil {
		return fmt.Errorf("failed to initialize cache: %w", err)
	}

	// Initialize services
	if err := s.initServices(); err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	// Initialize HTTP server
	if err := s.initHTTPServer(); err != nil {
		return fmt.Errorf("failed to initialize HTTP server: %w", err)
	}

	s.logger.Info("Fintech API Server initialized successfully")
	return nil
}

// initDatabase initializes the database connection
func (s *FintechAPIServer) initDatabase() error {
	s.logger.Info("Initializing database connection...")

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		s.config.Database.Host,
		s.config.Database.Port,
		s.config.Database.Username,
		s.config.Database.Password,
		s.config.Database.Database,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(s.config.Database.MaxOpenConns)
	db.SetMaxIdleConns(s.config.Database.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(s.config.Database.ConnMaxLifetime) * time.Second)

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	s.db = db
	s.logger.Info("Database connection established")
	return nil
}

// initCache initializes the Redis cache
func (s *FintechAPIServer) initCache() error {
	s.logger.Info("Initializing Redis cache...")

	cache, err := redis.NewClient(&redis.Config{
		Host:     s.config.Redis.Host,
		Port:     s.config.Redis.Port,
		Password: s.config.Redis.Password,
		DB:       s.config.Redis.DB,
		PoolSize: s.config.Redis.PoolSize,
	})
	if err != nil {
		return fmt.Errorf("failed to create Redis client: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cache.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	s.cache = cache
	s.logger.Info("Redis cache connection established")
	return nil
}

// initServices initializes all business logic services
func (s *FintechAPIServer) initServices() error {
	s.logger.Info("Initializing services...")

	// Initialize Accounts service
	accountsRepo := accounts.NewPostgreSQLRepository(s.db)
	s.accountsService = accounts.NewService(accountsRepo, s.config.Fintech.Accounts, s.logger, s.cache)

	// TODO: Initialize other services (Payments, Yield, Trading, Cards)

	s.logger.Info("Services initialized successfully")
	return nil
}

// initHTTPServer initializes the HTTP server and routes
func (s *FintechAPIServer) initHTTPServer() error {
	s.logger.Info("Initializing HTTP server...")

	// Set Gin mode
	if s.config.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	router.Use(cors.New(corsConfig))

	// Health check endpoint
	router.GET("/health", s.healthCheck)

	// API version 1
	v1 := router.Group("/api/v1")

	// Initialize handlers
	s.accountsHandler = accounts.NewHandler(s.accountsService, s.logger)

	// Register routes
	s.accountsHandler.RegisterRoutes(v1)

	// TODO: Register other module routes
	// s.paymentsHandler.RegisterRoutes(v1)
	// s.yieldHandler.RegisterRoutes(v1)
	// s.tradingHandler.RegisterRoutes(v1)
	// s.cardsHandler.RegisterRoutes(v1)

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(s.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(s.config.Server.IdleTimeout) * time.Second,
	}

	s.logger.Info("HTTP server initialized", zap.Int("port", s.config.Server.Port))
	return nil
}

// start starts the HTTP server
func (s *FintechAPIServer) start() error {
	s.logger.Info("Starting HTTP server", zap.Int("port", s.config.Server.Port))

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	s.logger.Info("Fintech API Server started successfully")
	s.logger.Info("API Documentation available at: http://localhost:" + fmt.Sprintf("%d", s.config.Server.Port) + "/docs")
	s.logger.Info("Health check available at: http://localhost:" + fmt.Sprintf("%d", s.config.Server.Port) + "/health")

	return nil
}

// waitForShutdown waits for shutdown signal
func (s *FintechAPIServer) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	s.logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
}

// shutdown gracefully shuts down the server
func (s *FintechAPIServer) shutdown() error {
	s.logger.Info("Shutting down Fintech API Server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if s.httpServer != nil {
		s.logger.Info("Shutting down HTTP server...")
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("Error shutting down HTTP server", zap.Error(err))
		}
	}

	// Close database connection
	if s.db != nil {
		s.logger.Info("Closing database connection...")
		if err := s.db.Close(); err != nil {
			s.logger.Error("Error closing database connection", zap.Error(err))
		}
	}

	// Close Redis connection
	if s.cache != nil {
		s.logger.Info("Closing Redis connection...")
		if err := s.cache.Close(); err != nil {
			s.logger.Error("Error closing Redis connection", zap.Error(err))
		}
	}

	s.logger.Info("Fintech API Server shutdown completed")
	return nil
}

// healthCheck handles health check requests
func (s *FintechAPIServer) healthCheck(c *gin.Context) {
	status := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "fintech-api",
		"version":   "1.0.0",
		"uptime":    time.Since(time.Now()).String(), // This would be calculated from start time
	}

	// Check database health
	if s.db != nil {
		if err := s.db.Ping(); err != nil {
			status["database"] = "unhealthy"
			status["database_error"] = err.Error()
			c.JSON(http.StatusServiceUnavailable, status)
			return
		}
		status["database"] = "healthy"
	}

	// Check Redis health
	if s.cache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		
		if err := s.cache.Ping(ctx); err != nil {
			status["cache"] = "unhealthy"
			status["cache_error"] = err.Error()
			c.JSON(http.StatusServiceUnavailable, status)
			return
		}
		status["cache"] = "healthy"
	}

	// Check services health
	status["services"] = gin.H{
		"accounts": "healthy",
		"payments": "not_implemented",
		"yield":    "not_implemented",
		"trading":  "not_implemented",
		"cards":    "not_implemented",
	}

	c.JSON(http.StatusOK, status)
}
