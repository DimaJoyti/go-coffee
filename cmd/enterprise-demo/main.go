package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/autoscaling"
	"github.com/DimaJoyti/go-coffee/pkg/chaos"
	"github.com/DimaJoyti/go-coffee/pkg/concurrency"
	"go.uber.org/zap"
)

// EnterpriseDemo demonstrates enterprise features without external dependencies
type EnterpriseDemo struct {
	logger          *zap.Logger
	workerPools     map[string]*concurrency.DynamicWorkerPool
	rateLimiter     *concurrency.RateLimiter
	circuitBreakers *concurrency.CircuitBreakerManager
	faultInjector   *chaos.FaultInjector
	autoScaler      *autoscaling.AutoScaler
	server          *http.Server
}

// Order represents a demo order
type Order struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	Items      []Item    `json:"items"`
	Total      float64   `json:"total"`
	Status     string    `json:"status"`
	Priority   int       `json:"priority"`
	CreatedAt  time.Time `json:"created_at"`
}

// Item represents an order item
type Item struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// NewEnterpriseDemo creates a new enterprise demo
func NewEnterpriseDemo(logger *zap.Logger) (*EnterpriseDemo, error) {
	demo := &EnterpriseDemo{
		logger:          logger,
		workerPools:     make(map[string]*concurrency.DynamicWorkerPool),
		circuitBreakers: concurrency.NewCircuitBreakerManager(logger),
	}

	// Initialize worker pools
	poolConfigs := map[string]*concurrency.WorkerPoolConfig{
		"standard": {
			MinWorkers:          3,
			MaxWorkers:          20,
			QueueSize:           500,
			WorkerTimeout:       30 * time.Second,
			ScaleUpThreshold:    0.8,
			ScaleDownThreshold:  0.2,
			ScaleUpCooldown:     1 * time.Minute,
			ScaleDownCooldown:   3 * time.Minute,
			HealthCheckInterval: 15 * time.Second,
			MetricsInterval:     10 * time.Second,
		},
		"high_priority": {
			MinWorkers:          2,
			MaxWorkers:          10,
			QueueSize:           200,
			WorkerTimeout:       15 * time.Second,
			ScaleUpThreshold:    0.6,
			ScaleDownThreshold:  0.1,
			ScaleUpCooldown:     30 * time.Second,
			ScaleDownCooldown:   2 * time.Minute,
			HealthCheckInterval: 10 * time.Second,
			MetricsInterval:     5 * time.Second,
		},
	}

	for name, config := range poolConfigs {
		pool := concurrency.NewDynamicWorkerPool(name, config, logger)
		if err := pool.Start(); err != nil {
			return nil, fmt.Errorf("failed to start worker pool %s: %w", name, err)
		}
		demo.workerPools[name] = pool
	}

	// Initialize rate limiter (local mode)
	rateLimiterConfig := &concurrency.RateLimiterConfig{
		Algorithm:         "sliding_window",
		DefaultLimit:      100,
		DefaultWindow:     1 * time.Minute,
		BurstSize:         20,
		Distributed:       false, // Local mode for demo
		CleanupInterval:   2 * time.Minute,
		HeadersEnabled:    true,
		BlockOnExceed:     false, // Just log for demo
		RetryAfterEnabled: true,
	}
	demo.rateLimiter = concurrency.NewRateLimiter(rateLimiterConfig, nil, logger)

	// Initialize chaos engineering (safe mode)
	chaosConfig := &chaos.ChaosConfig{
		Enabled:             true,
		GlobalFailureRate:   0.02, // 2% failure rate for demo
		SafeMode:            true,
		MaxConcurrentFaults: 2,
		MonitoringInterval:  30 * time.Second,
		Scenarios: map[string]*chaos.ScenarioConfig{
			"demo_latency": {
				Name:            "demo_latency",
				Enabled:         true,
				FailureRate:     0.1, // 10% of requests
				Duration:        0,   // Run indefinitely
				FaultType:       "latency",
				TargetEndpoints: []string{"/orders"},
				Parameters: map[string]interface{}{
					"min_latency": "50ms",
					"max_latency": "500ms",
				},
			},
			"demo_errors": {
				Name:            "demo_errors",
				Enabled:         true,
				FailureRate:     0.05, // 5% of requests
				Duration:        0,    // Run indefinitely
				FaultType:       "error",
				TargetEndpoints: []string{"/orders"},
				Parameters: map[string]interface{}{
					"status_code":   500,
					"error_message": "Demo chaos engineering fault",
				},
			},
		},
	}
	demo.faultInjector = chaos.NewFaultInjector(chaosConfig, logger)
	if err := demo.faultInjector.Start(); err != nil {
		return nil, fmt.Errorf("failed to start fault injector: %w", err)
	}

	// Initialize auto-scaler
	autoScalerConfig := &autoscaling.AutoScalerConfig{
		Enabled:                 true,
		MinReplicas:            2,
		MaxReplicas:            10,
		TargetCPUUtilization:   0.7,
		TargetMemoryUtilization: 0.8,
		ScaleUpCooldown:        1 * time.Minute,
		ScaleDownCooldown:      3 * time.Minute,
		MetricsWindow:          2 * time.Minute,
		EvaluationInterval:     15 * time.Second,
		PredictiveScaling:      true,
		CustomMetrics: map[string]*autoscaling.CustomMetric{
			"queue_depth": {
				Name:           "queue_depth",
				TargetValue:    50.0,
				Weight:         1.0,
				ScaleDirection: "up",
			},
		},
	}
	demo.autoScaler = autoscaling.NewAutoScaler(autoScalerConfig, logger)
	ctx := context.Background()
	if err := demo.autoScaler.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start auto-scaler: %w", err)
	}

	return demo, nil
}

// Start starts the demo service
func (ed *EnterpriseDemo) Start() error {
	ed.logger.Info("🚀 Starting Go Coffee Enterprise Demo")

	// Create HTTP server
	mux := http.NewServeMux()
	ed.setupRoutes(mux)

	// Apply middleware
	handler := ed.applyMiddleware(mux)

	ed.server = &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start metrics collection
	go ed.collectMetrics()

	// Start server
	go func() {
		ed.logger.Info("🌐 HTTP server starting on :8080")
		if err := ed.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ed.logger.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	return nil
}

// Stop stops the demo service
func (ed *EnterpriseDemo) Stop() error {
	ed.logger.Info("🛑 Stopping Go Coffee Enterprise Demo")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop HTTP server
	if ed.server != nil {
		if err := ed.server.Shutdown(ctx); err != nil {
			ed.logger.Error("HTTP server shutdown failed", zap.Error(err))
		}
	}

	// Stop worker pools
	for name, pool := range ed.workerPools {
		if err := pool.Stop(); err != nil {
			ed.logger.Error("Failed to stop worker pool", zap.String("pool", name), zap.Error(err))
		}
	}

	// Stop fault injector
	if err := ed.faultInjector.Stop(); err != nil {
		ed.logger.Error("Failed to stop fault injector", zap.Error(err))
	}

	// Stop auto-scaler
	if err := ed.autoScaler.Stop(); err != nil {
		ed.logger.Error("Failed to stop auto-scaler", zap.Error(err))
	}

	ed.logger.Info("✅ Go Coffee Enterprise Demo stopped")
	return nil
}

// setupRoutes sets up HTTP routes
func (ed *EnterpriseDemo) setupRoutes(mux *http.ServeMux) {
	// Business endpoints
	mux.HandleFunc("/orders", ed.handleOrders)
	mux.HandleFunc("/menu", ed.handleMenu)

	// Monitoring endpoints
	mux.HandleFunc("/health", ed.handleHealth)
	mux.HandleFunc("/metrics", ed.handleMetrics)
	mux.HandleFunc("/chaos", ed.handleChaos)
	mux.HandleFunc("/scaling", ed.handleScaling)
	mux.HandleFunc("/worker-pools", ed.handleWorkerPools)
	mux.HandleFunc("/circuit-breakers", ed.handleCircuitBreakers)

	// Demo endpoints
	mux.HandleFunc("/demo/load-test", ed.handleLoadTest)
	mux.HandleFunc("/demo/chaos-test", ed.handleChaosTest)

	// Root endpoint
	mux.HandleFunc("/", ed.handleRoot)
}

// applyMiddleware applies middleware to the handler
func (ed *EnterpriseDemo) applyMiddleware(handler http.Handler) http.Handler {
	// Apply middleware in reverse order
	handler = ed.faultInjector.HTTPMiddleware()(handler)
	handler = ed.rateLimiter.HTTPMiddleware()(handler)
	handler = ed.loggingMiddleware(handler)
	return handler
}

// HTTP Handlers

func (ed *EnterpriseDemo) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		ed.createOrderDemo(w, r)
	case http.MethodGet:
		ed.listOrdersDemo(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (ed *EnterpriseDemo) createOrderDemo(w http.ResponseWriter, r *http.Request) {
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	order.CreatedAt = time.Now()
	order.Status = "pending"

	// Submit to worker pool
	job := concurrency.Job{
		ID:       fmt.Sprintf("order-%s", order.ID),
		Type:     "order_processing",
		Payload:  order,
		Priority: order.Priority,
		Timeout:  10 * time.Second,
		MaxRetry: 2,
	}

	poolName := "standard"
	if order.Priority > 5 {
		poolName = "high_priority"
	}

	pool := ed.workerPools[poolName]
	if err := pool.SubmitJob(job); err != nil {
		ed.logger.Error("Failed to submit job", zap.Error(err))
		http.Error(w, "Failed to process order", http.StatusInternalServerError)
		return
	}

	// Use circuit breaker for payment simulation
	paymentCB := ed.circuitBreakers.GetOrCreate("payment_demo", &concurrency.CircuitBreakerConfig{
		FailureThreshold:     3,
		SuccessThreshold:     2,
		TimeoutThreshold:     5 * time.Second,
		OpenTimeout:          15 * time.Second,
		HalfOpenTimeout:      5 * time.Second,
		HalfOpenMaxRequests:  2,
		HalfOpenSuccessRatio: 0.5,
		ResetTimeout:         2 * time.Minute,
		MonitoringInterval:   15 * time.Second,
	})

	_, err := paymentCB.Execute(r.Context(), func(ctx context.Context) (interface{}, error) {
		// Simulate payment processing
		time.Sleep(time.Millisecond * 50)
		
		// Simulate occasional failures
		if time.Now().UnixNano()%7 == 0 {
			return nil, fmt.Errorf("payment service temporarily unavailable")
		}
		
		return map[string]string{"payment_id": fmt.Sprintf("pay_%s", order.ID)}, nil
	})

	if err != nil {
		order.Status = "payment_failed"
		ed.logger.Warn("Payment failed", zap.Error(err))
	} else {
		order.Status = "confirmed"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (ed *EnterpriseDemo) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "enterprise-demo-v1.0.0",
		"features": []string{
			"✅ Dynamic Worker Pools",
			"✅ Advanced Rate Limiting",
			"✅ Circuit Breakers",
			"✅ Chaos Engineering",
			"✅ Predictive Auto-scaling",
		},
		"components": map[string]interface{}{
			"worker_pools":     ed.getWorkerPoolStatus(),
			"circuit_breakers": ed.circuitBreakers.GetMetrics(),
			"chaos":            ed.faultInjector.GetMetrics(),
			"auto_scaling":     ed.autoScaler.GetMetrics(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (ed *EnterpriseDemo) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := map[string]interface{}{
		"worker_pools":     ed.getWorkerPoolMetrics(),
		"rate_limiter":     ed.rateLimiter.GetStats(),
		"circuit_breakers": ed.circuitBreakers.GetMetrics(),
		"chaos":            ed.faultInjector.GetMetrics(),
		"auto_scaling":     ed.autoScaler.GetMetrics(),
		"system": map[string]interface{}{
			"goroutines": runtime.NumGoroutine(),
			"memory":     ed.getMemoryStats(),
			"timestamp":  time.Now(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (ed *EnterpriseDemo) handleRoot(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"service": "🚀 Go Coffee Enterprise Demo",
		"version": "1.0.0",
		"message": "Enterprise-grade microservice with advanced concurrency, chaos engineering, and auto-scaling",
		"features": []string{
			"🔄 Dynamic Worker Pools with Auto-scaling",
			"⚡ Advanced Rate Limiting (Sliding Window)",
			"🛡️ Circuit Breakers with Fallbacks",
			"💥 Chaos Engineering (Safe Mode)",
			"📈 Predictive Auto-scaling",
			"📊 Real-time Performance Monitoring",
		},
		"endpoints": map[string]string{
			"POST /orders":           "Create order with advanced processing",
			"GET /orders":            "List orders",
			"GET /health":            "Health check with component status",
			"GET /metrics":           "Comprehensive metrics",
			"GET /chaos":             "Chaos engineering status",
			"GET /scaling":           "Auto-scaling metrics",
			"GET /worker-pools":      "Worker pool status",
			"GET /circuit-breakers":  "Circuit breaker status",
			"GET /demo/load-test":    "Generate load for testing",
			"GET /demo/chaos-test":   "Trigger chaos scenarios",
		},
		"demo_commands": []string{
			"curl -X POST http://localhost:8080/orders -d '{\"id\":\"order-123\",\"customer_id\":\"customer-456\",\"items\":[{\"product_id\":\"coffee-1\",\"name\":\"Espresso\",\"quantity\":2,\"price\":4.50}],\"total\":9.00,\"priority\":3}'",
			"curl http://localhost:8080/health",
			"curl http://localhost:8080/metrics",
			"curl http://localhost:8080/demo/load-test",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods and additional handlers...

func (ed *EnterpriseDemo) getWorkerPoolStatus() map[string]interface{} {
	status := make(map[string]interface{})
	for name, pool := range ed.workerPools {
		metrics := pool.GetMetrics()
		status[name] = map[string]interface{}{
			"active_workers": metrics.ActiveWorkers,
			"queue_depth":    metrics.QueueDepth,
			"total_jobs":     metrics.TotalJobs,
			"completed_jobs": metrics.CompletedJobs,
			"throughput":     fmt.Sprintf("%.2f jobs/sec", metrics.ThroughputPerSec),
		}
	}
	return status
}

func (ed *EnterpriseDemo) getWorkerPoolMetrics() map[string]*concurrency.WorkerPoolMetrics {
	metrics := make(map[string]*concurrency.WorkerPoolMetrics)
	for name, pool := range ed.workerPools {
		metrics[name] = pool.GetMetrics()
	}
	return metrics
}

func (ed *EnterpriseDemo) getMemoryStats() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"alloc_mb":       m.Alloc / 1024 / 1024,
		"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
		"sys_mb":         m.Sys / 1024 / 1024,
		"num_gc":         m.NumGC,
	}
}

func (ed *EnterpriseDemo) collectMetrics() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Simulate metrics collection
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		
		cpuUsage := 0.4 + (float64(time.Now().Unix()%10) / 20.0) // Simulate varying CPU
		memoryUsage := float64(m.Alloc) / float64(m.Sys)
		
		customMetrics := map[string]float64{
			"queue_depth": ed.getTotalQueueDepth(),
			"error_rate":  ed.getErrorRate(),
		}
		
		ed.autoScaler.UpdateMetrics(cpuUsage, memoryUsage, customMetrics)
	}
}

func (ed *EnterpriseDemo) getTotalQueueDepth() float64 {
	var total float64
	for _, pool := range ed.workerPools {
		metrics := pool.GetMetrics()
		total += float64(metrics.QueueDepth)
	}
	return total
}

func (ed *EnterpriseDemo) getErrorRate() float64 {
	cbMetrics := ed.circuitBreakers.GetMetrics()
	var totalRequests, totalFailures float64
	
	for _, metrics := range cbMetrics {
		totalRequests += float64(metrics.TotalRequests)
		totalFailures += float64(metrics.TotalFailures)
	}
	
	if totalRequests > 0 {
		return totalFailures / totalRequests
	}
	return 0
}

func (ed *EnterpriseDemo) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		ed.logger.Debug("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", time.Since(start)))
	})
}

// Placeholder handlers
func (ed *EnterpriseDemo) listOrdersDemo(w http.ResponseWriter, r *http.Request) {
	orders := []Order{
		{ID: "demo-1", CustomerID: "customer-1", Status: "completed", Total: 12.50, CreatedAt: time.Now().Add(-1 * time.Hour)},
		{ID: "demo-2", CustomerID: "customer-2", Status: "pending", Total: 8.75, CreatedAt: time.Now().Add(-30 * time.Minute)},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (ed *EnterpriseDemo) handleMenu(w http.ResponseWriter, r *http.Request) {
	menu := map[string]interface{}{
		"beverages": []map[string]interface{}{
			{"id": "espresso", "name": "Espresso", "price": 4.50},
			{"id": "latte", "name": "Latte", "price": 5.00},
			{"id": "cappuccino", "name": "Cappuccino", "price": 4.75},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menu)
}

func (ed *EnterpriseDemo) handleChaos(w http.ResponseWriter, r *http.Request) {
	metrics := ed.faultInjector.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (ed *EnterpriseDemo) handleScaling(w http.ResponseWriter, r *http.Request) {
	metrics := ed.autoScaler.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (ed *EnterpriseDemo) handleWorkerPools(w http.ResponseWriter, r *http.Request) {
	metrics := ed.getWorkerPoolMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (ed *EnterpriseDemo) handleCircuitBreakers(w http.ResponseWriter, r *http.Request) {
	metrics := ed.circuitBreakers.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (ed *EnterpriseDemo) handleLoadTest(w http.ResponseWriter, r *http.Request) {
	// Generate some load for demonstration
	for i := 0; i < 10; i++ {
		job := concurrency.Job{
			ID:      fmt.Sprintf("load-test-%d", i),
			Type:    "load_test",
			Payload: fmt.Sprintf("Load test job %d", i),
			Timeout: 5 * time.Second,
		}
		
		pool := ed.workerPools["standard"]
		pool.SubmitJob(job)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Generated 10 load test jobs",
		"status":  "success",
	})
}

func (ed *EnterpriseDemo) handleChaosTest(w http.ResponseWriter, r *http.Request) {
	metrics := ed.faultInjector.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Chaos engineering is active",
		"metrics": metrics,
		"note":    "Make requests to /orders to see chaos faults in action",
	})
}

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("🚀 Initializing Go Coffee Enterprise Demo")

	// Create demo service
	demo, err := NewEnterpriseDemo(logger)
	if err != nil {
		log.Fatal("Failed to create enterprise demo:", err)
	}

	// Start service
	if err := demo.Start(); err != nil {
		log.Fatal("Failed to start enterprise demo:", err)
	}

	logger.Info("✅ Go Coffee Enterprise Demo is running on http://localhost:8080")
	logger.Info("🔗 Try: curl http://localhost:8080")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	if err := demo.Stop(); err != nil {
		logger.Error("Failed to stop demo gracefully", zap.Error(err))
	}
}
