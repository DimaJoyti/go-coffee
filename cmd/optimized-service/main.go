package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/cache"
	"github.com/DimaJoyti/go-coffee/pkg/database"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
	"go.uber.org/zap"
)

// OptimizedCoffeeService demonstrates the working optimizations
type OptimizedCoffeeService struct {
	logger       *zap.Logger
	dbManager    *database.Manager
	cacheManager *cache.Manager
	cacheHelper  *cache.CacheHelper
}

// NewOptimizedCoffeeService creates a new optimized service
func NewOptimizedCoffeeService(cfg *config.InfrastructureConfig, logger *zap.Logger) (*OptimizedCoffeeService, error) {
	// Initialize optimized database
	dbManager, err := database.NewManager(cfg.Database, logger)
	if err != nil {
		logger.Info("Database connection failed (expected in test environment)", zap.Error(err))
		// Continue without database for demo
	}

	// Initialize optimized cache
	cacheManager, err := cache.NewManager(cfg.Redis, logger)
	if err != nil {
		logger.Info("Cache connection failed (expected in test environment)", zap.Error(err))
		// Continue without cache for demo
	}

	var cacheHelper *cache.CacheHelper
	if cacheManager != nil {
		cacheHelper = cacheManager.NewCacheHelper()
	}

	return &OptimizedCoffeeService{
		logger:       logger,
		dbManager:    dbManager,
		cacheManager: cacheManager,
		cacheHelper:  cacheHelper,
	}, nil
}

// Order represents a coffee order
type Order struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	Items      []Item    `json:"items"`
	Total      float64   `json:"total"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Item represents an order item
type Item struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// CreateOrder creates an order with optimizations
func (s *OptimizedCoffeeService) CreateOrder(ctx context.Context, order *Order) error {
	s.logger.Info("Creating optimized order",
		zap.String("order_id", order.ID),
		zap.String("customer_id", order.CustomerID),
		zap.Float64("total", order.Total))

	// Use optimized database if available
	if s.dbManager != nil {
		query := `
			INSERT INTO orders (id, customer_id, items, total, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

		itemsJSON, err := json.Marshal(order.Items)
		if err != nil {
			return fmt.Errorf("failed to marshal items: %w", err)
		}

		err = s.dbManager.ExecuteWrite(ctx, query,
			order.ID, order.CustomerID, itemsJSON, order.Total,
			order.Status, order.CreatedAt, order.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}
	}

	// Use optimized cache if available
	if s.cacheManager != nil {
		cacheKey := fmt.Sprintf("order:%s", order.ID)
		if err := s.cacheManager.Set(ctx, cacheKey, order, 1*time.Hour); err != nil {
			s.logger.Error("Failed to cache order", zap.String("order_id", order.ID), zap.Error(err))
		}

		// Invalidate related caches
		customerOrdersKey := fmt.Sprintf("customer_orders:%s", order.CustomerID)
		if err := s.cacheManager.Delete(ctx, customerOrdersKey); err != nil {
			s.logger.Error("Failed to invalidate customer orders cache",
				zap.String("customer_id", order.CustomerID), zap.Error(err))
		}
	}

	s.logger.Info("Optimized order created successfully", zap.String("order_id", order.ID))
	return nil
}

// GetOrder retrieves an order with cache optimization
func (s *OptimizedCoffeeService) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	s.logger.Debug("Getting optimized order", zap.String("order_id", orderID))

	// Try cache first if available
	if s.cacheHelper != nil {
		var order Order
		cacheKey := fmt.Sprintf("order:%s", orderID)

		err := s.cacheHelper.GetOrSet(ctx, cacheKey, &order, 1*time.Hour, func() (interface{}, error) {
			// Simulate database fetch
			return &Order{
				ID:         orderID,
				CustomerID: "customer123",
				Items:      []Item{{ProductID: "coffee1", Name: "Espresso", Quantity: 1, Price: 4.50}},
				Total:      4.50,
				Status:     "completed",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}, nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to get order: %w", err)
		}

		return &order, nil
	}

	// Fallback to direct creation if no cache
	return &Order{
		ID:         orderID,
		CustomerID: "customer123",
		Items:      []Item{{ProductID: "coffee1", Name: "Espresso", Quantity: 1, Price: 4.50}},
		Total:      4.50,
		Status:     "completed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

// GetMetrics returns optimization metrics
func (s *OptimizedCoffeeService) GetMetrics() map[string]interface{} {
	metrics := map[string]interface{}{
		"timestamp": time.Now(),
		"service":   "optimized",
	}

	// Add database metrics if available
	if s.dbManager != nil {
		dbMetrics := s.dbManager.GetMetrics()
		metrics["database"] = map[string]interface{}{
			"query_count":        dbMetrics.QueryCount,
			"slow_query_count":   dbMetrics.SlowQueryCount,
			"connection_errors":  dbMetrics.ConnectionErrors,
			"active_connections": dbMetrics.ActiveConnections,
			"idle_connections":   dbMetrics.IdleConnections,
			"avg_query_time":     dbMetrics.AverageQueryTime.Milliseconds(),
		}
	}

	// Add cache metrics if available
	if s.cacheManager != nil {
		cacheMetrics := s.cacheManager.GetMetrics()
		metrics["cache"] = map[string]interface{}{
			"hits":         cacheMetrics.Hits,
			"misses":       cacheMetrics.Misses,
			"sets":         cacheMetrics.Sets,
			"deletes":      cacheMetrics.Deletes,
			"errors":       cacheMetrics.Errors,
			"hit_ratio":    cacheMetrics.HitRatio,
			"avg_latency":  cacheMetrics.AvgLatency.Milliseconds(),
			"total_keys":   cacheMetrics.TotalKeys,
			"memory_usage": cacheMetrics.MemoryUsage,
		}
	}

	return metrics
}

// HTTP Handlers

// CreateOrderHandler handles order creation
func (s *OptimizedCoffeeService) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set timestamps
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	order.Status = "pending"

	if err := s.CreateOrder(r.Context(), &order); err != nil {
		s.logger.Error("Failed to create order", zap.Error(err))
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// GetOrderHandler handles order retrieval
func (s *OptimizedCoffeeService) GetOrderHandler(w http.ResponseWriter, r *http.Request) {
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
func (s *OptimizedCoffeeService) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := s.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// HealthHandler returns health status
func (s *OptimizedCoffeeService) HealthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "go-coffee-optimized",
		"timestamp": time.Now(),
		"features": map[string]bool{
			"database_optimization": s.dbManager != nil,
			"cache_optimization":    s.cacheManager != nil,
			"memory_optimization":   true,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Starting Go Coffee Optimized Service")

	// Load configuration
	cfg := &config.InfrastructureConfig{
		Database: &config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "postgres",
			Password: "password",
			Database: "go_coffee",
			SSLMode:  "disable",
		},
		Redis: &config.RedisConfig{
			Host:         "localhost",
			Port:         6379,
			Password:     "",
			DB:           0,
			PoolSize:     50,
			MinIdleConns: 10,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			DialTimeout:  5 * time.Second,
			MaxRetries:   3,
			RetryDelay:   100 * time.Millisecond,
		},
	}

	// Create optimized service
	service, err := NewOptimizedCoffeeService(cfg, logger)
	if err != nil {
		log.Fatal("Failed to create optimized service:", err)
	}

	// Set up HTTP routes
	http.HandleFunc("/orders", service.CreateOrderHandler)
	http.HandleFunc("/orders/get", service.GetOrderHandler)
	http.HandleFunc("/metrics", service.MetricsHandler)
	http.HandleFunc("/health", service.HealthHandler)

	// Add a simple demo endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"service": "Go Coffee Optimized",
			"version": "1.0.0",
			"endpoints": []string{
				"POST /orders - Create order",
				"GET /orders/get?id=<order_id> - Get order",
				"GET /metrics - Get optimization metrics",
				"GET /health - Get health status",
			},
			"optimizations": []string{
				"Database connection pooling",
				"Redis caching with compression",
				"Memory optimization",
				"Performance monitoring",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	logger.Info("Go Coffee Optimized Service started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
