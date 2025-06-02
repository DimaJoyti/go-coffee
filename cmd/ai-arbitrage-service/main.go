package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pb "github.com/DimaJoyti/go-coffee/api/proto"
	aiarbitrage "github.com/DimaJoyti/go-coffee/internal/ai-arbitrage"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	redismcp "github.com/DimaJoyti/go-coffee/pkg/redis-mcp"
)

const (
	defaultPort = "50054"
	serviceName = "ai-arbitrage-service"
)

func main() {
	// Initialize logger
	loggerInstance := logger.New(serviceName)
	defer loggerInstance.Sync()

	loggerInstance.Info("Starting AI Arbitrage Service")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		loggerInstance.WithError(err).Fatal("Failed to load configuration")
	}

	// Get port from environment or use default
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = defaultPort
	}

	// Initialize Redis client
	redisClient, err := redismcp.NewRedisClient(cfg.Redis)
	if err != nil {
		loggerInstance.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// Initialize AI service
	aiService, err := redismcp.NewAIService(cfg.AI, loggerInstance, redisClient)
	if err != nil {
		loggerInstance.WithError(err).Fatal("Failed to initialize AI service")
	}

	// Initialize arbitrage service
	arbitrageService, err := aiarbitrage.NewService(
		redisClient,
		aiService,
		loggerInstance,
		cfg,
	)
	if err != nil {
		loggerInstance.WithError(err).Fatal("Failed to initialize arbitrage service")
	}

	// Create gRPC server
	server := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(loggerInstance)),
		grpc.StreamInterceptor(streamLoggingInterceptor(loggerInstance)),
	)

	// Register services
	pb.RegisterArbitrageServiceServer(server, arbitrageService)

	// Register health check service
	healthServer := health.NewServer()
	healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	// Enable reflection for development
	reflection.Register(server)

	// Create listener
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		loggerInstance.WithError(err).Fatal("Failed to listen")
	}

	// Start server in goroutine
	go func() {
		loggerInstance.WithField("port", port).Info("AI Arbitrage Service listening")
		if err := server.Serve(lis); err != nil {
			loggerInstance.WithError(err).Fatal("Failed to serve")
		}
	}()

	// Start background services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := arbitrageService.Start(ctx); err != nil {
			loggerInstance.WithError(err).Error("Failed to start arbitrage service")
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	loggerInstance.Info("Shutting down AI Arbitrage Service")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop background services
	cancel()

	// Stop gRPC server
	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		loggerInstance.Info("Server stopped gracefully")
	case <-shutdownCtx.Done():
		loggerInstance.Warn("Server shutdown timeout, forcing stop")
		server.Stop()
	}

	loggerInstance.Info("AI Arbitrage Service stopped")
}

// loggingInterceptor logs gRPC requests
func loggingInterceptor(loggerInstance *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		if err != nil {
			loggerInstance.WithFields(map[string]interface{}{
				"method":   info.FullMethod,
				"duration": duration.String(),
			}).WithError(err).Error("gRPC request failed")
		} else {
			loggerInstance.WithFields(map[string]interface{}{
				"method":   info.FullMethod,
				"duration": duration.String(),
			}).Info("gRPC request completed")
		}

		return resp, err
	}
}

// streamLoggingInterceptor logs gRPC stream requests
func streamLoggingInterceptor(loggerInstance *logger.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		err := handler(srv, stream)

		duration := time.Since(start)

		if err != nil {
			loggerInstance.WithFields(map[string]interface{}{
				"method":   info.FullMethod,
				"duration": duration.String(),
			}).WithError(err).Error("gRPC stream failed")
		} else {
			loggerInstance.WithFields(map[string]interface{}{
				"method":   info.FullMethod,
				"duration": duration.String(),
			}).Info("gRPC stream completed")
		}

		return err
	}
}
