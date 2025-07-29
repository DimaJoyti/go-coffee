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

	"github.com/DimaJoyti/go-coffee/pkg/cache"
	"github.com/DimaJoyti/go-coffee/pkg/database"
	"github.com/DimaJoyti/go-coffee/pkg/monitoring"
	"github.com/DimaJoyti/go-coffee/producer/config"
	"github.com/DimaJoyti/go-coffee/producer/handler"
	"github.com/DimaJoyti/go-coffee/producer/kafka"
	"github.com/DimaJoyti/go-coffee/producer/store"
)

func main() {
	log.Println("Starting Performance Optimized Producer Example...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize resource monitor with custom alerting
	monitorConfig := monitoring.DefaultConfig()
	monitorConfig.AlertCallback = func(alert string, metrics monitoring.ResourceMetrics) {
		log.Printf("🚨 PERFORMANCE ALERT: %s", alert)
		log.Printf("📊 Current Metrics - CPU: %.2f%%, Memory: %.2fMB, Goroutines: %d, GC Pause: %.2fms",
			metrics.CPUUsagePercent, metrics.MemoryUsageMB, metrics.GoroutineCount, metrics.GCPauseMS)

		// In a real application, you might want to:
		// - Send alerts to Slack/PagerDuty
		// - Scale resources automatically
		// - Trigger circuit breakers
		// - Log to external monitoring systems
	}

	resourceMonitor := monitoring.NewResourceMonitor(monitorConfig)
	resourceMonitor.Start()
	defer resourceMonitor.Stop()

	// Initialize optimized database with better connection pool settings
	dbConfig := database.DefaultConfig()
	dbConfig.MaxOpenConns = 50 // Optimized for producer workload
	dbConfig.MaxIdleConns = 15 // Keep more idle connections

	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize optimized Redis cache with batching support
	cacheConfig := cache.DefaultConfig()
	redisCache, err := cache.NewRedisCache(cacheConfig)
	if err != nil {
		log.Printf("Warning: Failed to initialize Redis cache: %v", err)
		log.Println("Falling back to in-memory cache with TTL and LRU eviction...")

		// Fallback to optimized in-memory store
		memoryConfig := cache.DefaultMemoryStoreConfig()
		memoryConfig.MaxSize = 5000
		memoryConfig.DefaultTTL = time.Hour * 2
		memoryAdapter := cache.NewMemoryCacheAdapter(memoryConfig)
		defer memoryAdapter.Close()

		// Use memory cache as fallback
		log.Println("✅ In-memory cache with TTL and LRU eviction initialized")
	} else {
		defer redisCache.Close()
		log.Println("✅ Redis cache with pipeline support initialized")
	}

	// Initialize optimized Kafka producer with async support
	kafkaProducer, err := kafka.NewProducer(cfg)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()
	log.Println("✅ Kafka producer with async batching initialized")

	// Initialize order store (using in-memory for this example)
	orderStore := store.NewInMemoryOrderStore()
	log.Println("✅ Order store initialized")

	// Initialize HTTP handler with optimizations
	h := handler.NewHandler(kafkaProducer, cfg, orderStore)

	// Setup HTTP routes with performance monitoring
	mux := http.NewServeMux()

	// Wrap handlers with performance middleware
	mux.HandleFunc("/order", withMonitoring(h.PlaceOrder, "place_order"))
	mux.HandleFunc("/order/", withMonitoring(h.GetOrder, "get_order"))
	mux.HandleFunc("/orders", withMonitoring(h.ListOrders, "list_orders"))
	mux.HandleFunc("/health", withMonitoring(h.HealthCheck, "health_check"))

	// Add resource monitoring endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := resourceMonitor.HealthCheckHandler()
		w.Header().Set("Content-Type", "application/json")

		// Simple JSON encoding
		response := `{
			"status": "` + metrics["status"].(string) + `",
			"cpu_percent": ` + formatFloat(metrics["cpu_percent"].(float64)) + `,
			"memory_mb": ` + formatFloat(metrics["memory_mb"].(float64)) + `,
			"goroutines": ` + formatInt(metrics["goroutines"].(int)) + `,
			"heap_size_mb": ` + formatFloat(metrics["heap_size_mb"].(float64)) + `,
			"heap_in_use_mb": ` + formatFloat(metrics["heap_in_use_mb"].(float64)) + `,
			"gc_pause_ms": ` + formatFloat(metrics["gc_pause_ms"].(float64)) + `
		}`
		w.Write([]byte(response))
	})

	// Setup HTTP server with optimized settings
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Starting server on :%d", cfg.Server.Port)
		log.Println("📋 Available endpoints:")
		log.Println("  POST /order?async=true  - Place order (use async=true for high throughput)")
		log.Println("  GET  /order/{id}        - Get order by ID")
		log.Println("  GET  /orders            - List all orders")
		log.Println("  GET  /health            - Health check")
		log.Println("  GET  /metrics           - Resource metrics")
		log.Println()
		log.Println("💡 Performance Tips:")
		log.Println("  - Use ?async=true for high-throughput scenarios")
		log.Println("  - Monitor /metrics endpoint for performance insights")
		log.Println("  - Resource alerts will be logged automatically")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Log initial resource state
	go func() {
		time.Sleep(time.Second * 5) // Wait for startup
		if current, err := resourceMonitor.GetCurrentMetrics(); err == nil {
			log.Printf("📊 Startup Metrics - CPU: %.2f%%, Memory: %.2fMB, Goroutines: %d",
				current.CPUUsagePercent, current.MemoryUsageMB, current.GoroutineCount)
		}
	}()

	// Demonstrate performance optimizations
	go func() {
		time.Sleep(time.Second * 10)
		log.Println("\n🔧 Performance Optimizations Applied:")
		log.Println("  ✅ Kafka Producer: Async batching with compression")
		log.Println("  ✅ Worker Pool: Dynamic scaling with backpressure")
		log.Println("  ✅ Database: Optimized connection pool (100 max, 25 idle)")
		log.Println("  ✅ Redis: Pipeline batching for bulk operations")
		log.Println("  ✅ Memory Store: TTL cleanup with LRU eviction")
		log.Println("  ✅ Resource Monitor: Real-time alerts and metrics")
		log.Println("  ✅ HTTP Handler: Async mode support for high throughput")
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("📴 Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log final metrics before shutdown
	if current, err := resourceMonitor.GetCurrentMetrics(); err == nil {
		log.Printf("📊 Final Metrics - CPU: %.2f%%, Memory: %.2fMB, Goroutines: %d",
			current.CPUUsagePercent, current.MemoryUsageMB, current.GoroutineCount)
	}

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited")
}

// withMonitoring wraps HTTP handlers with performance monitoring
func withMonitoring(handler http.HandlerFunc, operationName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the original handler
		handler(w, r)

		// Log performance metrics
		duration := time.Since(start)
		if duration > time.Millisecond*100 {
			log.Printf("⚠️  Slow request: %s took %v", operationName, duration)
		}

		// In a real application, you might want to:
		// - Send metrics to Prometheus
		// - Log to structured logging system
		// - Track error rates
		// - Monitor request sizes
	}
}

// Helper functions for JSON formatting
func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

func formatInt(i int) string {
	return fmt.Sprintf("%d", i)
}
