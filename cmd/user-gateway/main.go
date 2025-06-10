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

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/DimaJoyti/go-coffee/internal/user"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/middleware"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/monitoring"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

const (
	defaultPort        = "8080"
	defaultAIOrderAddr = "localhost:50051"
	defaultKitchenAddr = "localhost:50052"
	defaultCommHubAddr = "localhost:50053"
	serviceName        = "user-gateway"
)

func main() {
	// Initialize logger
	logger := logger.New(serviceName)
	logger.Info("🚀 Starting User Gateway Service with Clean Architecture...")

	// Load infrastructure configuration
	cfg := config.DefaultInfrastructureConfig()

	// Override with environment variables
	if port := os.Getenv("HTTP_PORT"); port != "" {
		// Port will be used in server setup
	}
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		cfg.Redis.Host = redisHost
	}
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		cfg.Database.Host = dbHost
	}

	// Initialize infrastructure container
	container := infrastructure.NewContainer(cfg, logger)
	ctx := context.Background()

	if err := container.Initialize(ctx); err != nil {
		log.Fatal("Failed to initialize infrastructure:", err)
	}
	defer container.Shutdown(ctx)

	logger.Info("✅ Infrastructure initialized successfully")

	// Get configuration from environment for gRPC clients
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
		logger.WithError(err).Fatal("Failed to initialize AI Order client")
	}
	defer aiOrderClient.Close()

	kitchenClient, err := initializeKitchenClient(kitchenAddr, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Kitchen client")
	}
	defer kitchenClient.Close()

	commClient, err := initializeCommunicationClient(commHubAddr, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Communication client")
	}
	defer commClient.Close()

	logger.Info("✅ All gRPC clients initialized successfully")

	// Initialize handlers with infrastructure container
	handlers := user.NewHandlers(aiOrderClient, kitchenClient, commClient, container, logger)

	// Setup clean HTTP router (will replace Gin)
	router := setupCleanRouter(handlers, container, logger, ctx)

	// Create HTTP server with infrastructure-aware configuration
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
		logger.WithField("port", port).Info("🌐 User Gateway listening with Clean Architecture")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	logger.Info("🎯 User Gateway is running with Clean Architecture. Press Ctrl+C to stop.")
	<-c

	logger.Info("🛑 Shutting down User Gateway...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("✅ User Gateway stopped gracefully")
}

// initializeAIOrderClient creates a gRPC client for AI Order Service
func initializeAIOrderClient(addr string, logger *logger.Logger) (*grpc.ClientConn, error) {
	logger.WithField("address", addr).Info("Connecting to AI Order Service")

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AI Order Service: %w", err)
	}

	// Simple connection test
	state := conn.GetState()
	logger.WithField("state", state.String()).Info("AI Order Service connection state")

	return conn, nil
}

// initializeKitchenClient creates a gRPC client for Kitchen Service
func initializeKitchenClient(addr string, logger *logger.Logger) (*grpc.ClientConn, error) {
	logger.WithField("address", addr).Info("Connecting to Kitchen Service")

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Kitchen Service: %w", err)
	}

	// Simple connection test
	state := conn.GetState()
	logger.WithField("state", state.String()).Info("Kitchen Service connection state")

	return conn, nil
}

// initializeCommunicationClient creates a gRPC client for Communication Hub
func initializeCommunicationClient(addr string, logger *logger.Logger) (*grpc.ClientConn, error) {
	logger.WithField("address", addr).Info("Connecting to Communication Hub")

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Communication Hub: %w", err)
	}

	// Simple connection test
	state := conn.GetState()
	logger.WithField("state", state.String()).Info("Communication Hub connection state")

	return conn, nil
}

// setupCleanRouter configures a clean HTTP router using gorilla/mux (replacing Gin)
func setupCleanRouter(handlers *user.Handlers, container infrastructure.ContainerInterface, logger *logger.Logger, ctx context.Context) http.Handler {
	router := mux.NewRouter()

	// Create infrastructure-aware middleware
	mw := middleware.NewMiddleware(container, logger)

	// Get session manager from container
	sessionManager := container.GetSessionManager()

	// Initialize enhanced monitoring
	healthChecker := monitoring.NewHealthChecker(container, nil, logger)
	prometheusMetrics := monitoring.NewPrometheusMetrics(container, nil, logger)

	// Start Prometheus metrics server
	if err := prometheusMetrics.StartMetricsServer(); err != nil {
		logger.WithError(err).Error("Failed to start Prometheus metrics server")
	}

	// Start periodic metrics collection
	prometheusMetrics.StartPeriodicCollection(ctx)

	// Configure middleware with custom settings
	corsConfig := &middleware.CORSConfig{
		AllowAllOrigins:  false,
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080", "http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Request-ID", "X-API-Key"},
		ExposedHeaders:   []string{"X-Request-ID", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           86400,
	}

	rateLimitConfig := &middleware.RateLimitConfig{
		RequestsPerSecond: 20.0, // 20 requests per second for user gateway
		BurstSize:         50,   // Allow bursts up to 50 requests
		Enabled:           true,
	}

	authConfig := &middleware.AuthConfig{
		Enabled: true,
		ExcludedPaths: []string{
			"/health",
			"/metrics",
			"/api/docs",
			"/static/",
		},
		TokenHeader: "Authorization",
		TokenPrefix: "Bearer ",
	}

	// Configure session middleware
	sessionConfig := &middleware.SessionConfig{
		Enabled:        true,
		RequireSession: false, // Don't require sessions for all routes
		ExcludedPaths: []string{
			"/health",
			"/metrics",
			"/api/docs",
			"/static/",
		},
		SessionTimeout:  30 * time.Minute,
		RefreshInterval: 5 * time.Minute,
	}

	// Add comprehensive middleware chain
	router.Use(func(next http.Handler) http.Handler {
		return mw.Chain(
			next.ServeHTTP,
			mw.RequestIDMiddleware,
			mw.LoggingMiddleware,
			mw.RecoveryMiddleware,
			mw.SecurityHeadersMiddleware,
			mw.CORSMiddleware(corsConfig),
			mw.RateLimitMiddleware(rateLimitConfig),
			mw.SessionMiddleware(sessionManager, sessionConfig), // Session management
			mw.PerformanceMiddleware(prometheusMetrics, nil),    // Performance monitoring
			mw.TracingMiddleware(nil),                           // Request tracing
			mw.ValidationMiddleware(nil),                        // Use default config
			mw.MetricsMiddleware(nil),                           // Use default config
			mw.CacheMiddleware,
		)
	})

	// Health check endpoint with infrastructure health
	router.HandleFunc("/health", handlers.HealthCheck).Methods("GET")

	// Enhanced health check endpoint with detailed monitoring
	router.HandleFunc("/health/detailed", monitoring.HTTPHealthHandler(healthChecker)).Methods("GET")

	// Prometheus metrics endpoint
	router.Handle("/metrics", prometheusMetrics.GetHandler()).Methods("GET")

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Public API routes (no authentication required)
	api.HandleFunc("/docs", handlers.GetAPIDocumentation).Methods("GET")
	api.HandleFunc("/auth/login", handlers.LoginExample).Methods("POST")

	// Protected API routes (require authentication)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(func(next http.Handler) http.Handler {
		return mw.AuthenticationMiddleware(authConfig)(next.ServeHTTP)
	})

	// Order management routes (protected)
	orders := protected.PathPrefix("/orders").Subrouter()
	orders.HandleFunc("", handlers.CreateOrder).Methods("POST")
	orders.HandleFunc("/{id}", handlers.GetOrder).Methods("GET")
	orders.HandleFunc("", handlers.ListOrders).Methods("GET")
	orders.HandleFunc("/{id}/status", handlers.UpdateOrderStatus).Methods("PUT")
	orders.HandleFunc("/{id}", handlers.CancelOrder).Methods("DELETE")
	orders.HandleFunc("/{id}/predict-completion", handlers.PredictCompletionTime).Methods("GET")

	// AI recommendations routes (protected)
	recommendations := protected.PathPrefix("/recommendations").Subrouter()
	recommendations.HandleFunc("/orders", handlers.GetOrderRecommendations).Methods("GET")
	recommendations.HandleFunc("/patterns", handlers.AnalyzeOrderPatterns).Methods("GET")

	// Kitchen management routes (protected)
	kitchen := protected.PathPrefix("/kitchen").Subrouter()
	kitchen.HandleFunc("/queue", handlers.GetKitchenQueue).Methods("GET")
	kitchen.HandleFunc("/queue", handlers.AddToKitchenQueue).Methods("POST")
	kitchen.HandleFunc("/queue/{id}/status", handlers.UpdatePreparationStatus).Methods("PUT")
	kitchen.HandleFunc("/queue/{id}/complete", handlers.CompleteOrder).Methods("POST")
	kitchen.HandleFunc("/metrics", handlers.GetKitchenMetrics).Methods("GET")
	kitchen.HandleFunc("/optimize", handlers.OptimizeKitchenWorkflow).Methods("POST")
	kitchen.HandleFunc("/capacity/predict", handlers.PredictKitchenCapacity).Methods("GET")
	kitchen.HandleFunc("/ingredients", handlers.GetIngredientRequirements).Methods("GET")

	// Communication routes (protected)
	communication := protected.PathPrefix("/communication").Subrouter()
	communication.HandleFunc("/messages", handlers.SendMessage).Methods("POST")
	communication.HandleFunc("/broadcast", handlers.BroadcastMessage).Methods("POST")
	communication.HandleFunc("/messages/history", handlers.GetMessageHistory).Methods("GET")
	communication.HandleFunc("/services", handlers.GetActiveServices).Methods("GET")
	communication.HandleFunc("/notifications", handlers.SendNotification).Methods("POST")
	communication.HandleFunc("/analytics", handlers.GetCommunicationAnalytics).Methods("GET")

	// Customer routes (protected)
	customers := protected.PathPrefix("/customers").Subrouter()
	customers.HandleFunc("/{id}/profile", handlers.GetCustomerProfile).Methods("GET")
	customers.HandleFunc("/{id}/profile", handlers.UpdateCustomerProfile).Methods("PUT")
	customers.HandleFunc("/{id}/orders", handlers.GetCustomerOrders).Methods("GET")
	customers.HandleFunc("/{id}/recommendations", handlers.GetCustomerRecommendations).Methods("GET")

	// Session management routes (protected)
	auth := protected.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/profile", handlers.GetUserProfile).Methods("GET")
	auth.HandleFunc("/logout", handlers.LogoutExample).Methods("POST")

	// Analytics routes (protected)
	analytics := protected.PathPrefix("/analytics").Subrouter()
	analytics.HandleFunc("/orders", handlers.GetOrderAnalytics).Methods("GET")
	analytics.HandleFunc("/kitchen", handlers.GetKitchenAnalytics).Methods("GET")
	analytics.HandleFunc("/performance", handlers.GetPerformanceAnalytics).Methods("GET")
	analytics.HandleFunc("/ai-insights", handlers.GetAIInsights).Methods("GET")

	// WebSocket endpoint for real-time updates
	router.HandleFunc("/ws", handlers.HandleWebSocket).Methods("GET")

	// Static files (if needed)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))

	return router
}

// Note: Clean architecture migration completed - all handlers now use standard HTTP
