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

	"github.com/DimaJoyti/go-coffee/developer-dao/internal/dao"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/config"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/database"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/logger"
	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/redis"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	serviceName = "dao-governance-service"
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

	logger.Info("Starting DAO Governance Service",
		zap.String("service", serviceName),
		zap.String("version", version),
		zap.String("environment", cfg.Environment))

	// Initialize database
	db, err := database.NewPostgresConnection(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Run database migrations
	if err := database.RunMigrations(db, cfg.Database.MigrationPath); err != nil {
		logger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	// Initialize Redis
	redisClient, err := redis.NewClient(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize DAO service
	daoService, err := dao.NewService(dao.ServiceConfig{
		DB:          db,
		Redis:       redisClient,
		Logger:      logger,
		Config:      cfg,
		ServiceName: serviceName,
	})
	if err != nil {
		logger.Fatal("Failed to initialize DAO service", zap.Error(err))
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

	// DAO REST API endpoints
	v1 := router.Group("/api/v1")
	{
		proposals := v1.Group("/proposals")
		{
			proposals.GET("", daoService.GetProposalsHandler)
			proposals.POST("", daoService.CreateProposalHandler)
			proposals.GET("/:id", daoService.GetProposalHandler)
			proposals.POST("/:id/vote", daoService.VoteOnProposalHandler)
			proposals.GET("/:id/votes", daoService.GetProposalVotesHandler)
		}

		governance := v1.Group("/governance")
		{
			governance.GET("/stats", daoService.GetGovernanceStatsHandler)
			governance.GET("/voting-power/:address", daoService.GetVotingPowerHandler)
			governance.GET("/delegate/:address", daoService.GetDelegateHandler)
			governance.POST("/delegate", daoService.DelegateVotesHandler)
		}

		developers := v1.Group("/developers")
		{
			developers.GET("", daoService.GetDevelopersHandler)
			developers.GET("/:address", daoService.GetDeveloperHandler)
			developers.POST("/:address/profile", daoService.UpdateDeveloperProfileHandler)
			developers.GET("/:address/proposals", daoService.GetDeveloperProposalsHandler)
			developers.GET("/:address/votes", daoService.GetDeveloperVotesHandler)
		}
	}

	// Start HTTP server in goroutine
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.Info("Starting HTTP server", zap.Int("port", cfg.Server.Port))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Start background services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start proposal monitoring service
	go func() {
		if err := daoService.StartProposalMonitoring(ctx); err != nil {
			logger.Error("Proposal monitoring service failed", zap.Error(err))
		}
	}()

	// Start voting power cache updater
	go func() {
		if err := daoService.StartVotingPowerUpdater(ctx); err != nil {
			logger.Error("Voting power updater failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down DAO Governance Service...")

	// Cancel background services
	cancel()

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server forced to shutdown", zap.Error(err))
	}

	// Note: gRPC server shutdown will be added when gRPC is implemented

	logger.Info("DAO Governance Service stopped")
}
