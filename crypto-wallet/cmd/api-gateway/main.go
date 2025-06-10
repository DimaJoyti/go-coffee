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
	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/common"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	l := logger.NewLogger(cfg.Logging)
	l.Info("Starting API Gateway service")

	// Initialize Gin router
	router := gin.New()
	router.Use(gin.Recovery())

	// Add middleware
	router.Use(common.LoggerMiddleware(l))
	router.Use(common.RequestIDMiddleware())
	router.Use(common.CORSMiddleware())
	router.Use(common.RateLimitMiddleware(cfg.Security.RateLimit))

	// Register routes
	registerRoutes(router)

	// Configure HTTP server
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// Start server in a goroutine
	go func() {
		l.Info(fmt.Sprintf("API Gateway listening on %s:%d", cfg.Server.Host, cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatal(fmt.Sprintf("Failed to start server: %v", err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	l.Info("Shutting down API Gateway...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		l.Fatal(fmt.Sprintf("Server forced to shutdown: %v", err))
	}

	l.Info("API Gateway stopped")
}

func registerRoutes(router *gin.Engine) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", func(c *gin.Context) {
				// TODO: Implement registration
				c.JSON(http.StatusOK, gin.H{"message": "Registration endpoint"})
			})
			auth.POST("/login", func(c *gin.Context) {
				// TODO: Implement login
				c.JSON(http.StatusOK, gin.H{"message": "Login endpoint"})
			})
			auth.POST("/refresh", func(c *gin.Context) {
				// TODO: Implement token refresh
				c.JSON(http.StatusOK, gin.H{"message": "Token refresh endpoint"})
			})
		}

		// Wallet routes
		wallet := v1.Group("/wallets")
		{
			wallet.GET("/", func(c *gin.Context) {
				// TODO: Implement wallet listing
				c.JSON(http.StatusOK, gin.H{"message": "List wallets endpoint"})
			})
			wallet.POST("/", func(c *gin.Context) {
				// TODO: Implement wallet creation
				c.JSON(http.StatusOK, gin.H{"message": "Create wallet endpoint"})
			})
			wallet.GET("/:id", func(c *gin.Context) {
				// TODO: Implement wallet retrieval
				c.JSON(http.StatusOK, gin.H{"message": "Get wallet endpoint", "id": c.Param("id")})
			})
		}

		// Transaction routes
		tx := v1.Group("/transactions")
		{
			tx.GET("/", func(c *gin.Context) {
				// TODO: Implement transaction listing
				c.JSON(http.StatusOK, gin.H{"message": "List transactions endpoint"})
			})
			tx.POST("/", func(c *gin.Context) {
				// TODO: Implement transaction creation
				c.JSON(http.StatusOK, gin.H{"message": "Create transaction endpoint"})
			})
			tx.GET("/:id", func(c *gin.Context) {
				// TODO: Implement transaction retrieval
				c.JSON(http.StatusOK, gin.H{"message": "Get transaction endpoint", "id": c.Param("id")})
			})
		}

		// Smart contract routes
		contract := v1.Group("/contracts")
		{
			contract.GET("/", func(c *gin.Context) {
				// TODO: Implement contract listing
				c.JSON(http.StatusOK, gin.H{"message": "List contracts endpoint"})
			})
			contract.POST("/", func(c *gin.Context) {
				// TODO: Implement contract deployment
				c.JSON(http.StatusOK, gin.H{"message": "Deploy contract endpoint"})
			})
			contract.POST("/:address/call", func(c *gin.Context) {
				// TODO: Implement contract call
				c.JSON(http.StatusOK, gin.H{"message": "Call contract endpoint", "address": c.Param("address")})
			})
		}
	}
}
