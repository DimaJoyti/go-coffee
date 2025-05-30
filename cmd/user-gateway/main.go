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
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/DimaJoyti/go-coffee/internal/user"
	pb_ai_order "github.com/DimaJoyti/go-coffee/api/proto/ai_order"
	pb_kitchen "github.com/DimaJoyti/go-coffee/api/proto/kitchen"
	pb_communication "github.com/DimaJoyti/go-coffee/api/proto/communication"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

const (
	defaultPort           = "8080"
	defaultAIOrderAddr    = "localhost:50051"
	defaultKitchenAddr    = "localhost:50052"
	defaultCommHubAddr    = "localhost:50053"
	serviceName           = "user-gateway"
)

func main() {
	// Initialize logger
	logger := logger.New(serviceName)
	logger.Info("🚀 Starting User Gateway Service...")

	// Get configuration from environment
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = defaultPort
	}

	aiOrderAddr := os.Getenv("AI_ORDER_SERVICE_ADDR")
	if aiOrderAddr == "" {
		aiOrderAddr = defaultAIOrderAddr
	}

	kitchenAddr := os.Getenv("KITCHEN_SERVICE_ADDR")
	if kitchenAddr == "" {
		kitchenAddr = defaultKitchenAddr
	}

	commHubAddr := os.Getenv("COMMUNICATION_HUB_ADDR")
	if commHubAddr == "" {
		commHubAddr = defaultCommHubAddr
	}

	// Initialize gRPC clients
	aiOrderClient, err := initializeAIOrderClient(aiOrderAddr, logger)
	if err != nil {
		logger.Fatal("Failed to initialize AI Order client", zap.Error(err))
	}
	defer aiOrderClient.Close()

	kitchenClient, err := initializeKitchenClient(kitchenAddr, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Kitchen client", zap.Error(err))
	}
	defer kitchenClient.Close()

	commClient, err := initializeCommunicationClient(commHubAddr, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Communication client", zap.Error(err))
	}
	defer commClient.Close()

	logger.Info("✅ All gRPC clients initialized successfully")

	// Initialize handlers
	handlers := user.NewHandlers(aiOrderClient, kitchenClient, commClient, logger)

	// Setup Gin router
	router := setupRouter(handlers, logger)

	// Create HTTP server
	server := &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start HTTP server in goroutine
	go func() {
		logger.Info("🌐 User Gateway listening", zap.String("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	logger.Info("🎯 User Gateway is running. Press Ctrl+C to stop.")
	<-c

	logger.Info("🛑 Shutting down User Gateway...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("✅ User Gateway stopped gracefully")
}

// initializeAIOrderClient creates a gRPC client for AI Order Service
func initializeAIOrderClient(addr string, logger *logger.Logger) (*grpc.ClientConn, error) {
	logger.Info("Connecting to AI Order Service", zap.String("address", addr))

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AI Order Service: %w", err)
	}

	// Test connection
	client := pb_ai_order.NewAIOrderServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try a simple call to verify connection
	_, err = client.ListOrders(ctx, &pb_ai_order.ListOrdersRequest{
		PageSize: 1,
	})
	if err != nil {
		logger.Warn("AI Order Service connection test failed", zap.Error(err))
		// Don't fail startup, service might not be ready yet
	}

	return conn, nil
}

// initializeKitchenClient creates a gRPC client for Kitchen Service
func initializeKitchenClient(addr string, logger *logger.Logger) (*grpc.ClientConn, error) {
	logger.Info("Connecting to Kitchen Service", zap.String("address", addr))

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Kitchen Service: %w", err)
	}

	// Test connection
	client := pb_kitchen.NewKitchenServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try a simple call to verify connection
	_, err = client.GetQueue(ctx, &pb_kitchen.GetQueueRequest{
		LocationId: "test",
	})
	if err != nil {
		logger.Warn("Kitchen Service connection test failed", zap.Error(err))
		// Don't fail startup, service might not be ready yet
	}

	return conn, nil
}

// initializeCommunicationClient creates a gRPC client for Communication Hub
func initializeCommunicationClient(addr string, logger *logger.Logger) (*grpc.ClientConn, error) {
	logger.Info("Connecting to Communication Hub", zap.String("address", addr))

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Communication Hub: %w", err)
	}

	// Test connection
	client := pb_communication.NewCommunicationServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try a simple call to verify connection
	_, err = client.GetActiveServices(ctx, &pb_communication.GetActiveServicesRequest{})
	if err != nil {
		logger.Warn("Communication Hub connection test failed", zap.Error(err))
		// Don't fail startup, service might not be ready yet
	}

	return conn, nil
}

// setupRouter configures the Gin router with all routes and middleware
func setupRouter(handlers *user.Handlers, logger *logger.Logger) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(user.LoggerMiddleware(logger))
	router.Use(user.CORSMiddleware())
	router.Use(user.RateLimitMiddleware())

	// Health check endpoint
	router.GET("/health", handlers.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Order management routes
		orders := v1.Group("/orders")
		{
			orders.POST("", handlers.CreateOrder)
			orders.GET("/:id", handlers.GetOrder)
			orders.GET("", handlers.ListOrders)
			orders.PUT("/:id/status", handlers.UpdateOrderStatus)
			orders.DELETE("/:id", handlers.CancelOrder)
			orders.GET("/:id/predict-completion", handlers.PredictCompletionTime)
		}

		// AI recommendations routes
		recommendations := v1.Group("/recommendations")
		{
			recommendations.GET("/orders", handlers.GetOrderRecommendations)
			recommendations.GET("/patterns", handlers.AnalyzeOrderPatterns)
		}

		// Kitchen management routes
		kitchen := v1.Group("/kitchen")
		{
			kitchen.GET("/queue", handlers.GetKitchenQueue)
			kitchen.POST("/queue", handlers.AddToKitchenQueue)
			kitchen.PUT("/queue/:id/status", handlers.UpdatePreparationStatus)
			kitchen.POST("/queue/:id/complete", handlers.CompleteOrder)
			kitchen.GET("/metrics", handlers.GetKitchenMetrics)
			kitchen.POST("/optimize", handlers.OptimizeKitchenWorkflow)
			kitchen.GET("/capacity/predict", handlers.PredictKitchenCapacity)
			kitchen.GET("/ingredients", handlers.GetIngredientRequirements)
		}

		// Communication routes
		communication := v1.Group("/communication")
		{
			communication.POST("/messages", handlers.SendMessage)
			communication.POST("/broadcast", handlers.BroadcastMessage)
			communication.GET("/messages/history", handlers.GetMessageHistory)
			communication.GET("/services", handlers.GetActiveServices)
			communication.POST("/notifications", handlers.SendNotification)
			communication.GET("/analytics", handlers.GetCommunicationAnalytics)
		}

		// Customer routes
		customers := v1.Group("/customers")
		{
			customers.GET("/:id/profile", handlers.GetCustomerProfile)
			customers.PUT("/:id/profile", handlers.UpdateCustomerProfile)
			customers.GET("/:id/orders", handlers.GetCustomerOrders)
			customers.GET("/:id/recommendations", handlers.GetCustomerRecommendations)
		}

		// Analytics routes
		analytics := v1.Group("/analytics")
		{
			analytics.GET("/orders", handlers.GetOrderAnalytics)
			analytics.GET("/kitchen", handlers.GetKitchenAnalytics)
			analytics.GET("/performance", handlers.GetPerformanceAnalytics)
			analytics.GET("/ai-insights", handlers.GetAIInsights)
		}
	}

	// WebSocket endpoint for real-time updates
	router.GET("/ws", handlers.HandleWebSocket)

	// Static files (if needed)
	router.Static("/static", "./web/static")

	// API documentation endpoint
	router.GET("/api/docs", handlers.GetAPIDocumentation)

	return router
}
