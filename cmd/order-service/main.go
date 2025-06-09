package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/DimaJoyti/go-coffee/internal/order/application"
	"github.com/DimaJoyti/go-coffee/internal/order/infrastructure/repository"
	httpTransport "github.com/DimaJoyti/go-coffee/internal/order/transport/http"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

const (
	defaultHTTPPort = "8081"
	defaultGRPCPort = "50051"
	defaultRedisURL = "redis://localhost:6379"
	serviceName     = "order-service"
)

func main() {
	// Initialize logger
	logger := logger.New(serviceName)
	logger.Info("🚀 Starting Order Management Service...")

	// Get configuration from environment
	httpPort := getEnvOrDefault("HTTP_PORT", defaultHTTPPort)
	grpcPort := getEnvOrDefault("GRPC_PORT", defaultGRPCPort)
	redisURL := getEnvOrDefault("REDIS_URL", defaultRedisURL)

	// Initialize Redis client
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.WithError(err).Fatal("Failed to parse Redis URL")
	}

	redisClient := redis.NewClient(opt)

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to Redis")
	}

	logger.Info("✅ Connected to Redis successfully")

	// Initialize repositories
	orderRepo := repository.NewRedisOrderRepository(redisClient, logger)
	paymentRepo := repository.NewRedisPaymentRepository(redisClient, logger)

	// Initialize event publisher (placeholder)
	eventPublisher := &MockEventPublisher{logger: logger}

	// Initialize external services (placeholders)
	kitchenService := &MockKitchenService{logger: logger}
	authService := &MockAuthService{logger: logger}
	paymentProcessor := &MockPaymentProcessor{logger: logger}
	cryptoProcessor := &MockCryptoProcessor{logger: logger}
	loyaltyService := &MockLoyaltyService{logger: logger}

	// Initialize services
	orderService := application.NewOrderService(
		orderRepo,
		paymentRepo,
		eventPublisher,
		kitchenService,
		authService,
		logger,
	)

	paymentService := application.NewPaymentService(
		paymentRepo,
		orderRepo,
		eventPublisher,
		paymentProcessor,
		cryptoProcessor,
		loyaltyService,
		logger,
	)

	// Start HTTP server
	httpServer := startHTTPServer(httpPort, orderService, paymentService, logger)

	// Start gRPC server
	grpcServer := startGRPCServer(grpcPort, orderService, paymentService, logger)

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	logger.Info("🎯 Order Service is running. Press Ctrl+C to stop.")
	<-c

	logger.Info("🛑 Shutting down Order Service...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("HTTP server shutdown error")
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		logger.WithError(err).Error("Error closing Redis connection")
	}

	logger.Info("✅ Order Service stopped gracefully")
}

// startHTTPServer starts the HTTP server
func startHTTPServer(port string, orderService *application.OrderService, paymentService *application.PaymentService, logger *logger.Logger) *http.Server {
	// Create HTTP handler and middleware
	handler := httpTransport.NewHandler(orderService, paymentService, logger)
	middleware := httpTransport.NewMiddleware(logger)

	// Setup routes
	mux := http.NewServeMux()
	handler.SetupRoutes(mux)

	// Apply middleware chain
	finalHandler := middleware.Chain(
		mux.ServeHTTP,
		middleware.LoggingMiddleware,
		middleware.RecoveryMiddleware,
		middleware.CORSMiddleware,
		middleware.AuthMiddleware,
	)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      http.HandlerFunc(finalHandler),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.WithField("port", port).Info("🌐 HTTP Server listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	return server
}

// startGRPCServer starts the gRPC server
func startGRPCServer(port string, orderService *application.OrderService, paymentService *application.PaymentService, logger *logger.Logger) *grpc.Server {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.WithError(err).Fatal("Failed to listen for gRPC")
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(logger)),
	)

	// Register services (placeholder - will be implemented when protobuf is ready)
	// pb.RegisterOrderServiceServer(grpcServer, &OrderGRPCHandler{orderService: orderService})
	// pb.RegisterPaymentServiceServer(grpcServer, &PaymentGRPCHandler{paymentService: paymentService})

	// Enable reflection for development
	reflection.Register(grpcServer)

	go func() {
		logger.WithField("port", port).Info("🌐 gRPC Server listening")
		if err := grpcServer.Serve(listener); err != nil {
			logger.WithError(err).Fatal("Failed to serve gRPC")
		}
	}()

	return grpcServer
}

// Utility functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// loggingInterceptor provides request logging for gRPC
func loggingInterceptor(logger *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Log the request
		duration := time.Since(start)
		if err != nil {
			logger.WithFields(map[string]interface{}{
				"method":   info.FullMethod,
				"duration": duration,
			}).WithError(err).Error("gRPC request failed")
		} else {
			logger.WithFields(map[string]interface{}{
				"method":   info.FullMethod,
				"duration": duration,
			}).Info("gRPC request completed")
		}

		return resp, err
	}
}
