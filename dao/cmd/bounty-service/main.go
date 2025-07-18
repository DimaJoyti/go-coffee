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

	"github.com/DimaJoyti/go-coffee/developer-dao/internal/bounty"
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
	serviceName = "bounty-service"
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

	logger.Info("Starting Bounty Service",
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

	// Initialize Bounty service
	bountyService, err := bounty.NewService(bounty.ServiceConfig{
		DB:          db,
		Redis:       redisClient,
		Logger:      logger,
		Config:      cfg,
		ServiceName: serviceName,
	})
	if err != nil {
		logger.Fatal("Failed to initialize Bounty service", zap.Error(err))
	}

	// Start gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(cfg.GRPC.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(cfg.GRPC.MaxSendMsgSize),
	)

	// Register Bounty service (simplified for now)
	// bounty.RegisterBountyServiceServer(grpcServer, bountyService)

	// Enable reflection for development
	if cfg.Environment == "development" {
		reflection.Register(grpcServer)
	}

	// Start gRPC server in goroutine
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port+1))
	if err != nil {
		logger.Fatal("Failed to listen for gRPC", zap.Error(err))
	}

	go func() {
		logger.Info("Starting gRPC server", zap.Int("port", cfg.GRPC.Port+1))
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

	// Bounty REST API endpoints
	v1 := router.Group("/api/v1")
	{
		bounties := v1.Group("/bounties")
		{
			bounties.GET("", bountyService.GetBountiesHandler)
			bounties.POST("", bountyService.CreateBountyHandler)
			bounties.GET("/:id", bountyService.GetBountyHandler)
			bounties.POST("/:id/apply", bountyService.ApplyForBountyHandler)
			bounties.POST("/:id/assign", bountyService.AssignBountyHandler)
			bounties.POST("/:id/start", bountyService.StartBountyHandler)
			bounties.POST("/:id/submit", bountyService.SubmitBountyHandler)
			bounties.POST("/:id/milestones/:milestone_id/complete", bountyService.CompleteMilestoneHandler)
			bounties.GET("/:id/applications", bountyService.GetBountyApplicationsHandler)
		}

		performance := v1.Group("/performance")
		{
			performance.POST("/verify", bountyService.VerifyPerformanceHandler)
			performance.GET("/stats", bountyService.GetPerformanceStatsHandler)
		}

		developers := v1.Group("/developers")
		{
			developers.GET("/:address/bounties", bountyService.GetDeveloperBountiesHandler)
			developers.GET("/:address/reputation", bountyService.GetDeveloperReputationHandler)
			developers.GET("/leaderboard", bountyService.GetDeveloperLeaderboardHandler)
		}
	}

	// Start HTTP server in goroutine
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port+1),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.Info("Starting HTTP server", zap.Int("port", cfg.Server.Port+1))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Start background services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start bounty monitoring service
	go func() {
		if err := bountyService.StartBountyMonitoring(ctx); err != nil {
			logger.Error("Bounty monitoring service failed", zap.Error(err))
		}
	}()

	// Start performance tracking service
	go func() {
		if err := bountyService.StartPerformanceTracking(ctx); err != nil {
			logger.Error("Performance tracking service failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Bounty Service...")

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

	logger.Info("Bounty Service stopped")
}
