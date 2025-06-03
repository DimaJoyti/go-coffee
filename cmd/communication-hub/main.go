package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/DimaJoyti/go-coffee/internal/communication"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	redismcp "github.com/DimaJoyti/go-coffee/pkg/redis-mcp"
)

const (
	defaultPort     = "50053"
	defaultRedisURL = "redis://localhost:6379"
	serviceName     = "communication-hub"
)

func main() {
	// Initialize logger
	logger := logger.New(serviceName)
	logger.Info("🚀 Starting Communication Hub Service...")

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
		logger.Fatal("Failed to parse Redis URL: %v", err)
	}

	redisClient := redis.NewClient(opt)

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		logger.Fatal("Failed to connect to Redis: %v", err)
	}

	logger.Info("✅ Connected to Redis successfully")

	// Initialize AI service for communication optimization
	aiService, err := initializeAIService(redisClient, logger)
	if err != nil {
		logger.Fatal("Failed to initialize AI service: %v", err)
	}

	logger.Info("✅ AI service initialized successfully")

	// Initialize communication hub
	commHub := communication.NewHub(redisClient, aiService, logger)

	// Initialize message router
	messageRouter := communication.NewRouter(redisClient, aiService, logger)

	// Initialize communication service
	commService := communication.NewService(commHub, messageRouter, logger)

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(logger)),
		grpc.StreamInterceptor(streamLoggingInterceptor(logger)),
	)

	// Register Communication Service
	communication.RegisterCommunicationServiceServer(grpcServer, commService)

	// Enable reflection for development
	reflection.Register(grpcServer)

	// Start gRPC server
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Fatal("Failed to listen: %v", err)
	}

	// Start server in goroutine
	go func() {
		logger.Info("🌐 Communication Hub listening on port %s", port)
		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatal("Failed to serve gRPC: %v", err)
		}
	}()

	// Start background services
	go commHub.Start(context.Background())
	go messageRouter.Start(context.Background())

	// Initialize sample communication data
	go initializeSampleCommunicationData(redisClient, logger)

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	logger.Info("🎯 Communication Hub is running. Press Ctrl+C to stop.")
	<-c

	logger.Info("🛑 Shutting down Communication Hub...")

	// Graceful shutdown
	grpcServer.GracefulStop()

	// Stop background services
	commHub.Stop()
	messageRouter.Stop()

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		logger.Error("Error closing Redis connection: %v", err)
	}

	logger.Info("✅ Communication Hub stopped gracefully")
}

// initializeAIService creates and configures the AI service for communication
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
	aiService, err := redismcp.NewAIService(config, logger, redisClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI service: %w", err)
	}

	return aiService, nil
}

// loggingInterceptor provides request logging for gRPC unary calls
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
			logger.Error("gRPC unary request failed: method=%s, duration=%v, error=%v",
				info.FullMethod, duration, err)
		} else {
			logger.Info("gRPC unary request completed: method=%s, duration=%v",
				info.FullMethod, duration)
		}

		return resp, err
	}
}

// streamLoggingInterceptor provides request logging for gRPC streaming calls
func streamLoggingInterceptor(logger *logger.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		// Call the handler
		err := handler(srv, ss)

		// Log the request
		duration := time.Since(start)
		if err != nil {
			logger.Error("gRPC stream request failed: method=%s, duration=%v, error=%v",
				info.FullMethod, duration, err)
		} else {
			logger.Info("gRPC stream request completed: method=%s, duration=%v",
				info.FullMethod, duration)
		}

		return err
	}
}

// initializeSampleCommunicationData populates Redis with sample communication data
func initializeSampleCommunicationData(client *redis.Client, logger *logger.Logger) {
	ctx := context.Background()

	logger.Info("🏪 Initializing sample communication data in Redis...")

	// Sample registered services
	services := map[string]map[string]string{
		"comm:service:ai-order": {
			"service_id":   "ai-order-service",
			"service_name": "AI Order Management",
			"service_type": "SERVICE_TYPE_AI_ORDER",
			"endpoint":     "localhost:50051",
			"version":      "1.0.0",
			"is_active":    "true",
		},
		"comm:service:kitchen": {
			"service_id":   "kitchen-service",
			"service_name": "Kitchen Management",
			"service_type": "SERVICE_TYPE_KITCHEN",
			"endpoint":     "localhost:50052",
			"version":      "1.0.0",
			"is_active":    "true",
		},
		"comm:service:user-gateway": {
			"service_id":   "user-gateway",
			"service_name": "User Gateway",
			"service_type": "SERVICE_TYPE_USER_GATEWAY",
			"endpoint":     "localhost:8080",
			"version":      "1.0.0",
			"is_active":    "true",
		},
	}

	// Set service data
	for serviceKey, data := range services {
		for field, value := range data {
			if err := client.HSet(ctx, serviceKey, field, value).Err(); err != nil {
				logger.Error("Failed to set service data: service=%s, field=%s, error=%v",
					serviceKey, field, err)
			}
		}
		logger.Info("✅ Service data set for %s", serviceKey)
	}

	// Sample communication analytics
	analytics := map[string]float64{
		"total_messages":    1250,
		"messages_today":    156,
		"avg_response_time": 0.85,
		"failed_deliveries": 12,
		"success_rate":      98.5,
	}

	for metric, value := range analytics {
		if err := client.ZAdd(ctx, "comm:analytics:daily", &redis.Z{
			Score:  value,
			Member: metric,
		}).Err(); err != nil {
			logger.Error("Failed to add communication analytics: metric=%s, error=%v",
				metric, err)
		}
	}
	logger.Info("✅ Communication analytics data set")

	// Sample AI communication insights
	insights := map[string]string{
		"peak_communication_time": "14:00-16:00",
		"most_active_service":     "ai-order-service",
		"bottleneck_analysis":     "Kitchen service shows 15% higher response times during peak hours",
		"optimization_suggestion": "Implement message batching for non-critical notifications",
		"network_efficiency":      "92.3%",
	}

	for insight, value := range insights {
		if err := client.HSet(ctx, "comm:ai:insights", insight, value).Err(); err != nil {
			logger.Error("Failed to set AI insight: insight=%s, error=%v",
				insight, err)
		}
	}
	logger.Info("✅ AI communication insights set")

	// Sample message templates
	templates := map[string]string{
		"order_confirmed":    "Your order #{order_id} has been confirmed. Estimated completion: {estimated_time}",
		"order_ready":        "Order #{order_id} is ready for pickup at {location}!",
		"order_delayed":      "We apologize, but order #{order_id} is delayed by {delay_minutes} minutes",
		"kitchen_alert":      "Kitchen alert: {equipment_name} requires attention at {location}",
		"staff_notification": "Staff notification: {message} - Priority: {priority}",
	}

	for template, content := range templates {
		if err := client.HSet(ctx, "comm:templates", template, content).Err(); err != nil {
			logger.Error("Failed to set message template: template=%s, error=%v",
				template, err)
		}
	}
	logger.Info("✅ Message templates set")

	logger.Info("🎉 Sample communication data initialization completed successfully!")
}
