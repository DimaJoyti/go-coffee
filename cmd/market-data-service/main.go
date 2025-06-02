package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/DimaJoyti/go-coffee/internal/market-data"
	pb "github.com/DimaJoyti/go-coffee/api/proto"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/DimaJoyti/go-coffee/pkg/redis-mcp"
)

const (
	defaultPort = "50055"
	serviceName = "market-data-service"
)

func main() {
	// Initialize logger
	logger := logger.NewLogger(serviceName)
	defer logger.Sync()

	logger.Info("Starting Market Data Service")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", logger.Error(err))
	}

	// Get port from environment or use default
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = defaultPort
	}

	// Initialize Redis client
	redisClient, err := redismcp.NewRedisClient(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", logger.Error(err))
	}
	defer redisClient.Close()

	// Initialize market data service
	marketDataService, err := marketdata.NewService(
		redisClient,
		logger,
		cfg,
	)
	if err != nil {
		logger.Fatal("Failed to initialize market data service", logger.Error(err))
	}

	// Create gRPC server
	server := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(logger)),
		grpc.StreamInterceptor(streamLoggingInterceptor(logger)),
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
		logger.Fatal("Failed to listen", logger.Error(err))
	}

	// Start server in goroutine
	go func() {
		logger.Info("Market Data Service listening", logger.String("port", port))
		if err := server.Serve(lis); err != nil {
			logger.Fatal("Failed to serve", logger.Error(err))
		}
	}()

	// Start background services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := marketDataService.Start(ctx); err != nil {
			logger.Error("Failed to start market data service", logger.Error(err))
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logger.Info("Shutting down Market Data Service")

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
		logger.Info("Server stopped gracefully")
	case <-shutdownCtx.Done():
		logger.Warn("Server shutdown timeout, forcing stop")
		server.Stop()
	}

	logger.Info("Market Data Service stopped")
}

// loggingInterceptor logs gRPC requests
func loggingInterceptor(logger *logger.Logger) grpc.UnaryServerInterceptor {
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
			logger.Error("gRPC request failed",
				logger.String("method", info.FullMethod),
				logger.Duration("duration", duration),
				logger.Error(err),
			)
		} else {
			logger.Info("gRPC request completed",
				logger.String("method", info.FullMethod),
				logger.Duration("duration", duration),
			)
		}
		
		return resp, err
	}
}

// streamLoggingInterceptor logs gRPC stream requests
func streamLoggingInterceptor(logger *logger.Logger) grpc.StreamServerInterceptor {
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
			logger.Error("gRPC stream failed",
				logger.String("method", info.FullMethod),
				logger.Duration("duration", duration),
				logger.Error(err),
			)
		} else {
			logger.Info("gRPC stream completed",
				logger.String("method", info.FullMethod),
				logger.Duration("duration", duration),
			)
		}
		
		return err
	}
}
