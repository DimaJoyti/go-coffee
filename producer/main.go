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

	// Create handler
	h := handler.NewHandler(kafkaProducer, cfg)

	// Create router
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/order", h.PlaceOrder)
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
