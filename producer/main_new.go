package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"kafka_producer/config"
	"kafka_producer/handler"
	"kafka_producer/kafka"
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

	// Apply middleware
	// Chain middleware in order: Recover -> Logging -> RequestID -> CORS
	handler := middleware.RecoverMiddleware(
		middleware.LoggingMiddleware(
			middleware.RequestIDMiddleware(
				middleware.CORSMiddleware(mux),
			),
		),
	)

	// Start server
	serverAddr := ":" + strconv.Itoa(cfg.Server.Port)
	fmt.Printf("Starting server on %s\n", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, handler))
}
