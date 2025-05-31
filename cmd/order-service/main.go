package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/DimaJoyti/go-coffee/internal/order/application"
	"github.com/DimaJoyti/go-coffee/internal/order/infrastructure/repository"
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   serviceName,
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Order endpoints
		orders := v1.Group("/orders")
		{
			orders.POST("", createOrderHandler(orderService, logger))
			orders.GET("/:id", getOrderHandler(orderService, logger))
			orders.POST("/:id/confirm", confirmOrderHandler(orderService, logger))
			orders.PUT("/:id/status", updateOrderStatusHandler(orderService, logger))
			orders.DELETE("/:id", cancelOrderHandler(orderService, logger))
		}

		// Payment endpoints
		payments := v1.Group("/payments")
		{
			payments.POST("", createPaymentHandler(paymentService, logger))
			payments.POST("/:id/process", processPaymentHandler(paymentService, logger))
			payments.POST("/:id/refund", refundPaymentHandler(paymentService, logger))
		}
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
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

// HTTP Handlers

func createOrderHandler(orderService *application.OrderService, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req application.CreateOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := orderService.CreateOrder(c.Request.Context(), &req)
		if err != nil {
			logger.WithError(err).Error("Failed to create order")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, resp)
	}
}

func getOrderHandler(orderService *application.OrderService, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")
		customerID := c.Query("customer_id")

		req := &application.GetOrderRequest{
			OrderID:    orderID,
			CustomerID: customerID,
		}

		resp, err := orderService.GetOrder(c.Request.Context(), req)
		if err != nil {
			logger.WithError(err).Error("Failed to get order")
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func confirmOrderHandler(orderService *application.OrderService, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")
		
		var req application.ConfirmOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		req.OrderID = orderID

		resp, err := orderService.ConfirmOrder(c.Request.Context(), &req)
		if err != nil {
			logger.WithError(err).Error("Failed to confirm order")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func updateOrderStatusHandler(orderService *application.OrderService, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")
		
		var req application.UpdateOrderStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		req.OrderID = orderID

		resp, err := orderService.UpdateOrderStatus(c.Request.Context(), &req)
		if err != nil {
			logger.WithError(err).Error("Failed to update order status")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func cancelOrderHandler(orderService *application.OrderService, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("id")
		customerID := c.Query("customer_id")
		reason := c.Query("reason")

		req := &application.CancelOrderRequest{
			OrderID:    orderID,
			CustomerID: customerID,
			Reason:     reason,
		}

		resp, err := orderService.CancelOrder(c.Request.Context(), req)
		if err != nil {
			logger.WithError(err).Error("Failed to cancel order")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func createPaymentHandler(paymentService *application.PaymentService, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req application.CreatePaymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := paymentService.CreatePayment(c.Request.Context(), &req)
		if err != nil {
			logger.WithError(err).Error("Failed to create payment")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, resp)
	}
}

func processPaymentHandler(paymentService *application.PaymentService, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		paymentID := c.Param("id")
		
		var req application.ProcessPaymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		req.PaymentID = paymentID

		resp, err := paymentService.ProcessPayment(c.Request.Context(), &req)
		if err != nil {
			logger.WithError(err).Error("Failed to process payment")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func refundPaymentHandler(paymentService *application.PaymentService, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		paymentID := c.Param("id")
		
		var req application.RefundPaymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		req.PaymentID = paymentID

		resp, err := paymentService.RefundPayment(c.Request.Context(), &req)
		if err != nil {
			logger.WithError(err).Error("Failed to refund payment")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
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
