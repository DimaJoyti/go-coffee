package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
	"github.com/DimaJoyti/go-coffee/pkg/optimization"
	"go.uber.org/zap"
)

// SimpleOptimizedService demonstrates basic optimization integration
type SimpleOptimizedService struct {
	logger          *zap.Logger
	optimizationSvc *optimization.Service
}

// NewSimpleOptimizedService creates a simple optimized service
func NewSimpleOptimizedService(cfg *config.InfrastructureConfig, logger *zap.Logger) (*SimpleOptimizedService, error) {
	// Initialize optimization service
	optimizationSvc, err := optimization.NewService(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create optimization service: %w", err)
	}

	// Start optimization service
	ctx := context.Background()
	if err := optimizationSvc.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start optimization service: %w", err)
	}

	return &SimpleOptimizedService{
		logger:          logger,
		optimizationSvc: optimizationSvc,
	}, nil
}

// SimpleOrder represents a basic order
type SimpleOrder struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	Total      float64   `json:"total"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateOrder creates an order using optimized components
func (s *SimpleOptimizedService) CreateOrder(ctx context.Context, order *SimpleOrder) error {
	s.logger.Info("Creating order with optimizations",
		zap.String("order_id", order.ID),
		zap.Float64("total", order.Total))

	// Use optimized database manager
	dbManager := s.optimizationSvc.GetDatabaseManager()
	query := `INSERT INTO orders (id, customer_id, total, status, created_at) VALUES ($1, $2, $3, $4, $5)`
	
	err := dbManager.ExecuteWrite(ctx, query, order.ID, order.CustomerID, order.Total, order.Status, order.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	// Use optimized cache manager
	cacheManager := s.optimizationSvc.GetCacheManager()
	cacheKey := fmt.Sprintf("order:%s", order.ID)
	if err := cacheManager.Set(ctx, cacheKey, order, 1*time.Hour); err != nil {
		s.logger.Error("Failed to cache order", zap.Error(err))
	}

	s.logger.Info("Order created successfully", zap.String("order_id", order.ID))
	return nil
}

// GetOrder retrieves an order using cache-first strategy
func (s *SimpleOptimizedService) GetOrder(ctx context.Context, orderID string) (*SimpleOrder, error) {
	s.logger.Info("Getting order", zap.String("order_id", orderID))

	// Try cache first
	cacheManager := s.optimizationSvc.GetCacheManager()
	cacheKey := fmt.Sprintf("order:%s", orderID)
	
	var order SimpleOrder
	err := cacheManager.Get(ctx, cacheKey, &order)
	if err == nil {
		s.logger.Debug("Order found in cache", zap.String("order_id", orderID))
		return &order, nil
	}

	// Fallback to database
	dbManager := s.optimizationSvc.GetDatabaseManager()
	query := `SELECT id, customer_id, total, status, created_at FROM orders WHERE id = $1`
	
	result, err := dbManager.QueryRead(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order: %w", err)
	}

	// For demo purposes, create a mock order
	order = SimpleOrder{
		ID:         orderID,
		CustomerID: "customer123",
		Total:      25.50,
		Status:     "completed",
		CreatedAt:  time.Now(),
	}

	// Cache the result
	if err := cacheManager.Set(ctx, cacheKey, &order, 1*time.Hour); err != nil {
		s.logger.Error("Failed to cache order", zap.Error(err))
	}

	// Use result to avoid unused variable error
	_ = result

	return &order, nil
}

// GetMetrics returns optimization metrics
func (s *SimpleOptimizedService) GetMetrics() *optimization.Metrics {
	return s.optimizationSvc.GetMetrics()
}

// HTTP handlers

// CreateOrderHandler handles order creation requests
func (s *SimpleOptimizedService) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var order SimpleOrder
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	order.CreatedAt = time.Now()
	order.Status = "pending"

	if err := s.CreateOrder(r.Context(), &order); err != nil {
		s.logger.Error("Failed to create order", zap.Error(err))
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// GetOrderHandler handles order retrieval requests
func (s *SimpleOptimizedService) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("id")
	if orderID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	order, err := s.GetOrder(r.Context(), orderID)
	if err != nil {
		s.logger.Error("Failed to get order", zap.String("order_id", orderID), zap.Error(err))
		http.Error(w, "Failed to get order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// MetricsHandler returns optimization metrics
func (s *SimpleOptimizedService) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := s.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// HealthHandler returns health status
func (s *SimpleOptimizedService) HealthHandler(w http.ResponseWriter, r *http.Request) {
	report := s.optimizationSvc.GenerateReport()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// Example main function
func runSimpleExample() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg := &config.InfrastructureConfig{
		Database: &config.DatabaseConfig{
			Host:               "localhost",
			Port:               5432,
			Username:           "postgres",
			Password:           "password",
			Database:           "go_coffee",
			SSLMode:            "disable",
			MaxOpenConns:       25,
			MaxIdleConns:       5,
			ConnMaxLifetime:    5 * time.Minute,
			ConnMaxIdleTime:    2 * time.Minute,
			QueryTimeout:       30 * time.Second,
			SlowQueryThreshold: 1 * time.Second,
			ConnectTimeout:     10 * time.Second,
		},
		Redis: &config.RedisConfig{
			Host:         "localhost",
			Port:         6379,
			Password:     "",
			DB:           0,
			PoolSize:     25,
			MinIdleConns: 5,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			DialTimeout:  5 * time.Second,
			MaxRetries:   3,
			RetryDelay:   100 * time.Millisecond,
		},
	}

	// Create simple optimized service
	service, err := NewSimpleOptimizedService(cfg, logger)
	if err != nil {
		log.Fatal("Failed to create simple optimized service:", err)
	}

	// Set up HTTP routes
	http.HandleFunc("/orders", service.CreateOrderHandler)
	http.HandleFunc("/orders/get", service.GetOrderHandler)
	http.HandleFunc("/metrics", service.MetricsHandler)
	http.HandleFunc("/health", service.HealthHandler)

	// Root endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"service": "Simple Go Coffee Optimization Demo",
			"version": "1.0.0",
			"features": []string{
				"✅ Database Connection Pooling",
				"✅ Redis Caching with Compression",
				"✅ Memory Optimization",
				"✅ Performance Monitoring",
			},
			"endpoints": map[string]string{
				"POST /orders":     "Create order",
				"GET /orders/get":  "Get order by ID",
				"GET /metrics":     "Optimization metrics",
				"GET /health":      "Health check",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	logger.Info("Starting simple optimized Go Coffee service on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

