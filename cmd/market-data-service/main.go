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
	marketdata "github.com/DimaJoyti/go-coffee/internal/market-data"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	redismcp "github.com/DimaJoyti/go-coffee/pkg/redis-mcp"
)

const (
	defaultPort = "50055"
	serviceName = "market-data-service"
)

func main() {
	// Initialize logger
	loggerInstance := logger.New(serviceName)
	defer loggerInstance.Sync()

	loggerInstance.Info("Starting Market Data Service")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		loggerInstance.Fatal("Failed to load configuration: %v", err)
	}

	// Get port from environment or use default
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = defaultPort
	}

	// Initialize Redis client
	redisClient, err := redismcp.NewRedisClient(cfg.Redis)
	if err != nil {
		loggerInstance.Fatal("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize market data service
	marketDataService, err := marketdata.NewService(
		redisClient,
		loggerInstance,
		cfg,
	)
	if err != nil {
		loggerInstance.Fatal("Failed to initialize market data service: %v", err)
	}

	// Create gRPC server
	server := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(loggerInstance)),
		grpc.StreamInterceptor(streamLoggingInterceptor(loggerInstance)),
	)

	// Register services
	pb.RegisterMarketDataServiceServer(server, marketDataService)

	// Register health check service
	healthServer := health.NewServer()
	healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	// Enable reflection for development
	reflection.Register(server)

	// Create listener
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		loggerInstance.Fatal("Failed to listen: %v", err)
	}

	// Start server in goroutine
	go func() {
		loggerInstance.Info("Market Data Service listening on port: %s", port)
		if err := server.Serve(lis); err != nil {
			loggerInstance.Fatal("Failed to serve: %v", err)
		}
	}()

	// Start background services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := marketDataService.Start(ctx); err != nil {
			loggerInstance.Error("Failed to start market data service: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	loggerInstance.Info("Shutting down Market Data Service")

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

	loggerInstance.Info("Market Data Service stopped")
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
			loggerInstance.ErrorWithFields("gRPC request failed",
				logger.String("method", info.FullMethod),
				logger.Duration("duration", duration),
				logger.Error(err),
			)
		} else {
			loggerInstance.InfoWithFields("gRPC request completed",
				logger.String("method", info.FullMethod),
				logger.Duration("duration", duration),
			)
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
			loggerInstance.ErrorWithFields("gRPC stream failed",
				logger.String("method", info.FullMethod),
				logger.Duration("duration", duration),
				logger.Error(err),
			)
		} else {
			loggerInstance.InfoWithFields("gRPC stream completed",
				logger.String("method", info.FullMethod),
				logger.Duration("duration", duration),
			)
		}

		return err
	}
}
