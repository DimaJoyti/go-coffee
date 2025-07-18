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

	"github.com/DimaJoyti/go-coffee/developer-dao/internal/metrics"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/config"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/database"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/logger"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/redis"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	serviceName = "metrics-service"
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

	logger.Info("Starting TVL/MAU Metrics Service",
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

	// Initialize Metrics service
	metricsService, err := metrics.NewService(metrics.ServiceConfig{
		DB:          db,
		Redis:       redisClient,
		Logger:      logger,
		Config:      cfg,
		ServiceName: serviceName,
	})
	if err != nil {
		logger.Fatal("Failed to initialize Metrics service", zap.Error(err))
	}

	// Note: gRPC server implementation will be added in future iterations
	// Currently focusing on REST API implementation

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

	// TVL/MAU Metrics REST API endpoints
	v1 := router.Group("/api/v1")
	{
		tvl := v1.Group("/tvl")
		{
			tvl.GET("", metricsService.GetTVLMetricsHandler)
			tvl.POST("/record", metricsService.RecordTVLHandler)
			tvl.GET("/history", metricsService.GetTVLHistoryHandler)
			tvl.GET("/by-protocol", metricsService.GetTVLByProtocolHandler)
			tvl.GET("/growth", metricsService.GetTVLGrowthHandler)
		}

		mau := v1.Group("/mau")
		{
			mau.GET("", metricsService.GetMAUMetricsHandler)
			mau.POST("/record", metricsService.RecordMAUHandler)
			mau.GET("/history", metricsService.GetMAUHistoryHandler)
			mau.GET("/by-feature", metricsService.GetMAUByFeatureHandler)
			mau.GET("/growth", metricsService.GetMAUGrowthHandler)
		}

		performance := v1.Group("/performance")
		{
			performance.GET("/dashboard", metricsService.GetPerformanceDashboardHandler)
			performance.GET("/attribution", metricsService.GetAttributionAnalysisHandler)
			performance.POST("/impact", metricsService.RecordImpactHandler)
			performance.GET("/leaderboard", metricsService.GetImpactLeaderboardHandler)
		}

		analytics := v1.Group("/analytics")
		{
			analytics.GET("/overview", metricsService.GetAnalyticsOverviewHandler)
			analytics.GET("/trends", metricsService.GetTrendsAnalysisHandler)
			analytics.GET("/forecasts", metricsService.GetForecastsHandler)
			analytics.GET("/alerts", metricsService.GetAlertsHandler)
			analytics.POST("/alerts", metricsService.CreateAlertHandler)
		}

		reports := v1.Group("/reports")
		{
			reports.GET("/daily", metricsService.GetDailyReportHandler)
			reports.GET("/weekly", metricsService.GetWeeklyReportHandler)
			reports.GET("/monthly", metricsService.GetMonthlyReportHandler)
			reports.POST("/generate", metricsService.GenerateCustomReportHandler)
		}

		integrations := v1.Group("/integrations")
		{
			integrations.POST("/webhook", metricsService.HandleWebhookHandler)
			integrations.GET("/sources", metricsService.GetDataSourcesHandler)
			integrations.POST("/sources", metricsService.AddDataSourceHandler)
		}
	}

	// Start HTTP server in goroutine
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port+3),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.Info("Starting HTTP server", zap.Int("port", cfg.Server.Port+3))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Start background services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start metrics collection service
	go func() {
		if err := metricsService.StartMetricsCollection(ctx); err != nil {
			logger.Error("Metrics collection service failed", zap.Error(err))
		}
	}()

	// Start analytics processing service
	go func() {
		if err := metricsService.StartAnalyticsProcessing(ctx); err != nil {
			logger.Error("Analytics processing service failed", zap.Error(err))
		}
	}()

	// Start alerting service
	go func() {
		if err := metricsService.StartAlertingService(ctx); err != nil {
			logger.Error("Alerting service failed", zap.Error(err))
		}
	}()

	// Start reporting service
	go func() {
		if err := metricsService.StartReportingService(ctx); err != nil {
			logger.Error("Reporting service failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down TVL/MAU Metrics Service...")

	// Cancel background services
	cancel()

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server forced to shutdown", zap.Error(err))
	}

	// Note: gRPC server shutdown will be added when gRPC is implemented

	logger.Info("TVL/MAU Metrics Service stopped")
}
