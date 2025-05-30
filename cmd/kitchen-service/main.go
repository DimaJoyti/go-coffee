package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/DimaJoyti/go-coffee/internal/kitchen"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	redismcp "github.com/DimaJoyti/go-coffee/pkg/redis-mcp"
)

const (
	defaultPort     = "50052"
	defaultRedisURL = "redis://localhost:6379"
	serviceName     = "kitchen-service"
)

func main() {
	// Initialize logger
	logger := logger.New(serviceName)
	logger.Info("🚀 Starting Kitchen Management Service...")

	// Get configuration from environment
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = defaultPort
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = defaultRedisURL
	}

	// Initialize Redis client
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.Fatal("Failed to parse Redis URL", zap.Error(err))
	}

	redisClient := redis.NewClient(opt)

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	logger.Info("✅ Connected to Redis successfully")

	// Initialize AI service for kitchen optimization
	aiService, err := initializeAIService(redisClient, logger)
	if err != nil {
		logger.Fatal("Failed to initialize AI service", zap.Error(err))
	}

	logger.Info("✅ AI service initialized successfully")

	// Initialize kitchen repository
	kitchenRepo := kitchen.NewRedisKitchenRepository(redisClient, logger)

	// Initialize kitchen optimizer
	kitchenOptimizer := kitchen.NewKitchenOptimizer(aiService, logger)

	// Initialize queue manager
	queueManager := kitchen.NewQueueManager(redisClient, kitchenOptimizer, logger)

	// Initialize kitchen service
	kitchenService := kitchen.NewService(kitchenRepo, queueManager, kitchenOptimizer, logger)

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(logger)),
	)

	// Register Kitchen Service
	kitchen.RegisterKitchenServiceServer(grpcServer, kitchenService)

	// Enable reflection for development
	reflection.Register(grpcServer)

	// Start gRPC server
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	// Start server in goroutine
	go func() {
		logger.Info("🌐 Kitchen Service listening", zap.String("port", port))
		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// Initialize sample kitchen data
	go initializeSampleKitchenData(redisClient, logger)

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	logger.Info("🎯 Kitchen Service is running. Press Ctrl+C to stop.")
	<-c

	logger.Info("🛑 Shutting down Kitchen Service...")

	// Graceful shutdown
	grpcServer.GracefulStop()

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		logger.Error("Error closing Redis connection", zap.Error(err))
	}

	logger.Info("✅ Kitchen Service stopped gracefully")
}

// initializeAIService creates and configures the AI service for kitchen operations
func initializeAIService(redisClient *redis.Client, logger *logger.Logger) (*redismcp.AIService, error) {
	// Get AI configuration from environment
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	ollamaBaseURL := os.Getenv("OLLAMA_BASE_URL")
	if ollamaBaseURL == "" {
		ollamaBaseURL = "http://localhost:11434"
	}

	// Create AI service configuration
	config := &redismcp.AIConfig{
		Gemini: redismcp.GeminiConfig{
			APIKey: geminiAPIKey,
			Model:  "gemini-pro",
		},
		Ollama: redismcp.OllamaConfig{
			BaseURL: ollamaBaseURL,
			Model:   "llama2",
		},
	}

	// Initialize AI service
	aiService, err := redismcp.NewAIService(config, redisClient, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI service: %w", err)
	}

	return aiService, nil
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
			logger.Error("gRPC request failed",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
				zap.Error(err),
			)
		} else {
			logger.Info("gRPC request completed",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration),
			)
		}

		return resp, err
	}
}

// initializeSampleKitchenData populates Redis with sample kitchen data
func initializeSampleKitchenData(client *redis.Client, logger *logger.Logger) {
	ctx := context.Background()

	logger.Info("🏪 Initializing sample kitchen data in Redis...")

	// Sample kitchen equipment
	equipment := map[string]map[string]string{
		"kitchen:equipment:espresso-01": {
			"id":               "espresso-01",
			"name":             "Professional Espresso Machine",
			"station_type":     "ESPRESSO",
			"status":           "AVAILABLE",
			"efficiency_score": "9.2",
			"current_load":     "0",
			"max_capacity":     "4",
		},
		"kitchen:equipment:grinder-01": {
			"id":               "grinder-01",
			"name":             "Commercial Coffee Grinder",
			"station_type":     "GRINDER",
			"status":           "AVAILABLE",
			"efficiency_score": "8.8",
			"current_load":     "0",
			"max_capacity":     "2",
		},
		"kitchen:equipment:steamer-01": {
			"id":               "steamer-01",
			"name":             "Milk Steamer",
			"station_type":     "STEAMER",
			"status":           "AVAILABLE",
			"efficiency_score": "9.0",
			"current_load":     "0",
			"max_capacity":     "3",
		},
	}

	// Set equipment data
	for equipmentKey, data := range equipment {
		for field, value := range data {
			if err := client.HSet(ctx, equipmentKey, field, value).Err(); err != nil {
				logger.Error("Failed to set equipment data",
					zap.Error(err),
					zap.String("equipment", equipmentKey),
					zap.String("field", field),
				)
			}
		}
		logger.Info("✅ Equipment data set", zap.String("equipment", equipmentKey))
	}

	// Sample kitchen staff
	staff := map[string]map[string]string{
		"kitchen:staff:barista-01": {
			"id":             "barista-01",
			"name":           "Alice Cooper",
			"specializations": "ESPRESSO,STEAMER",
			"skill_level":    "9.5",
			"is_available":   "true",
			"current_orders": "0",
		},
		"kitchen:staff:barista-02": {
			"id":             "barista-02",
			"name":           "Bob Wilson",
			"specializations": "GRINDER,ASSEMBLY",
			"skill_level":    "8.7",
			"is_available":   "true",
			"current_orders": "0",
		},
	}

	// Set staff data
	for staffKey, data := range staff {
		for field, value := range data {
			if err := client.HSet(ctx, staffKey, field, value).Err(); err != nil {
				logger.Error("Failed to set staff data",
					zap.Error(err),
					zap.String("staff", staffKey),
					zap.String("field", field),
				)
			}
		}
		logger.Info("✅ Staff data set", zap.String("staff", staffKey))
	}

	// Sample kitchen performance metrics
	metrics := map[string]float64{
		"avg_preparation_time": 4.2,
		"orders_completed":     156,
		"orders_in_queue":      8,
		"efficiency_rate":      92.5,
		"customer_satisfaction": 8.9,
	}

	for metric, value := range metrics {
		if err := client.ZAdd(ctx, "kitchen:metrics:daily", &redis.Z{
			Score:  value,
			Member: metric,
		}).Err(); err != nil {
			logger.Error("Failed to add kitchen metrics",
				zap.Error(err),
				zap.String("metric", metric),
			)
		}
	}
	logger.Info("✅ Kitchen metrics data set")

	// Sample AI optimization suggestions
	optimizations := map[string]string{
		"workflow_optimization": "Parallel processing of espresso and milk steaming can reduce preparation time by 15%",
		"staff_allocation":      "Assign Alice to espresso station during peak hours for optimal efficiency",
		"equipment_usage":       "Grinder utilization can be improved by 20% with better scheduling",
		"queue_management":      "Implement priority-based queue ordering to reduce customer wait time",
	}

	for optimization, suggestion := range optimizations {
		if err := client.HSet(ctx, "kitchen:ai:optimizations", optimization, suggestion).Err(); err != nil {
			logger.Error("Failed to set AI optimization",
				zap.Error(err),
				zap.String("optimization", optimization),
			)
		}
	}
	logger.Info("✅ AI optimization suggestions set")

	logger.Info("🎉 Sample kitchen data initialization completed successfully!")
}
