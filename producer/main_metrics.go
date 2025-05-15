package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"kafka_producer/config"
	"kafka_producer/handler"
	"kafka_producer/kafka"
	"kafka_producer/metrics"
	"kafka_producer/middleware"
	"kafka_producer/store"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create Kafka producer
	kafkaProducer, err := kafka.NewProducer(cfg)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// Create order store
	orderStore := store.NewInMemoryOrderStore()

	// Create handler
	h := handler.NewHandler(kafkaProducer, cfg, orderStore)

	// Create router
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/order", h.PlaceOrder)
	mux.HandleFunc("/orders", h.ListOrders)
	mux.HandleFunc("/order/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/order/" {
			http.Error(w, "Order ID is required", http.StatusBadRequest)
			return
		}

		// Check if it's a cancel request
		if r.Method == http.MethodPost && len(path) > 7 && path[len(path)-7:] == "/cancel" {
			h.CancelOrder(w, r)
			return
		}

		// Otherwise, it's a get request
		h.GetOrder(w, r)
	})
	mux.HandleFunc("/health", h.HealthCheck)

	// Register metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Apply middleware
	// Chain middleware in order: Recover -> Logging -> RequestID -> CORS
	handler := middleware.RecoverMiddleware(
		middleware.LoggingMiddleware(
			middleware.RequestIDMiddleware(
				middleware.CORSMiddleware(mux),
			),
		),
	)

	// Start metrics server on a different port
	metricsPort := cfg.Server.Port + 1
	go func() {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", promhttp.Handler())
		metricsAddr := ":" + strconv.Itoa(metricsPort)
		fmt.Printf("Starting metrics server on %s\n", metricsAddr)
		if err := http.ListenAndServe(metricsAddr, metricsMux); err != nil {
			log.Printf("Metrics server failed: %v", err)
		}
	}()

	// Start server
	serverAddr := ":" + strconv.Itoa(cfg.Server.Port)
	fmt.Printf("Starting server on %s\n", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, handler))
}
