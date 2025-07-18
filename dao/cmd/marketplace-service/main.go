package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DimaJoyti/go-coffee/developer-dao/internal/marketplace"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/config"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/database"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/logger"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/redis"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	serviceName = "marketplace-service"
	version     = "1.0.0"
)

func main() {
	// Load configuration
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := logger.New(cfg.Logging.Level, cfg.Logging.Format)
	defer logger.Sync()

	logger.Info("Starting Solution Marketplace Service",
		zap.String("service", serviceName),
		zap.String("version", version),
		zap.String("environment", cfg.Environment))

	// Initialize database
	db, err := database.NewPostgresConnection(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := redis.NewClient(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize Marketplace service
	marketplaceService, err := marketplace.NewService(marketplace.ServiceConfig{
		DB:          db,
		Redis:       redisClient,
		Logger:      logger,
		Config:      cfg,
		ServiceName: serviceName,
	})
	if err != nil {
		logger.Fatal("Failed to initialize Marketplace service", zap.Error(err))
	}

	// Start gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(cfg.GRPC.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(cfg.GRPC.MaxSendMsgSize),
	)

	// Register Marketplace service (simplified for now)
	// marketplace.RegisterMarketplaceServiceServer(grpcServer, marketplaceService)

	// Enable reflection for development
	if cfg.Environment == "development" {
		reflection.Register(grpcServer)
	}

	// Start gRPC server in goroutine
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port+2))
	if err != nil {
		logger.Fatal("Failed to listen for gRPC", zap.Error(err))
	}

	go func() {
		logger.Info("Starting gRPC server", zap.Int("port", cfg.GRPC.Port+2))
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// Start HTTP server for REST API and metrics
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": serviceName,
			"version": version,
			"status":  "healthy",
			"time":    time.Now().UTC(),
		})
	})

	// Metrics endpoint
	if cfg.Monitoring.Prometheus.Enabled {
		router.GET(cfg.Monitoring.Prometheus.Path, gin.WrapH(promhttp.Handler()))
	}

	// Solution Marketplace REST API endpoints
	v1 := router.Group("/api/v1")
	{
		solutions := v1.Group("/solutions")
		{
			solutions.GET("", marketplaceService.GetSolutionsHandler)
			solutions.POST("", marketplaceService.CreateSolutionHandler)
			solutions.GET("/:id", marketplaceService.GetSolutionHandler)
			solutions.PUT("/:id", marketplaceService.UpdateSolutionHandler)
			solutions.POST("/:id/review", marketplaceService.ReviewSolutionHandler)
			solutions.POST("/:id/approve", marketplaceService.ApproveSolutionHandler)
			solutions.POST("/:id/install", marketplaceService.InstallSolutionHandler)
			solutions.GET("/:id/compatibility", marketplaceService.CheckCompatibilityHandler)
			solutions.GET("/:id/reviews", marketplaceService.GetSolutionReviewsHandler)
		}

		categories := v1.Group("/categories")
		{
			categories.GET("", marketplaceService.GetCategoriesHandler)
			categories.GET("/:category/solutions", marketplaceService.GetSolutionsByCategoryHandler)
		}

		quality := v1.Group("/quality")
		{
			quality.POST("/score", marketplaceService.CalculateQualityScoreHandler)
			quality.GET("/metrics", marketplaceService.GetQualityMetricsHandler)
		}

		developers := v1.Group("/developers")
		{
			developers.GET("/:address/solutions", marketplaceService.GetDeveloperSolutionsHandler)
			developers.GET("/:address/reviews", marketplaceService.GetDeveloperReviewsHandler)
		}

		analytics := v1.Group("/analytics")
		{
			analytics.GET("/popular", marketplaceService.GetPopularSolutionsHandler)
			analytics.GET("/trending", marketplaceService.GetTrendingSolutionsHandler)
			analytics.GET("/stats", marketplaceService.GetMarketplaceStatsHandler)
		}
	}

	// Start HTTP server in goroutine
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port+2),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.Info("Starting HTTP server", zap.Int("port", cfg.Server.Port+2))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Start background services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start quality monitoring service
	go func() {
		if err := marketplaceService.StartQualityMonitoring(ctx); err != nil {
			logger.Error("Quality monitoring service failed", zap.Error(err))
		}
	}()

	// Start analytics service
	go func() {
		if err := marketplaceService.StartAnalyticsService(ctx); err != nil {
			logger.Error("Analytics service failed", zap.Error(err))
		}
	}()

	// Start compatibility checking service
	go func() {
		if err := marketplaceService.StartCompatibilityService(ctx); err != nil {
			logger.Error("Compatibility service failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Solution Marketplace Service...")

	// Cancel background services
	cancel()

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server forced to shutdown", zap.Error(err))
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()

	logger.Info("Solution Marketplace Service stopped")
}
