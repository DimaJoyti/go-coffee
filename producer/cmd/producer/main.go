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

	"github.com/DimaJoyti/go-coffee/producer/config"
	"github.com/DimaJoyti/go-coffee/producer/handler"
	"github.com/DimaJoyti/go-coffee/producer/kafka"
	"github.com/DimaJoyti/go-coffee/producer/store"
)

func main() {
	log.Println("Starting Coffee Producer Service...")

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

	// Create HTTP handlers
	h := handler.NewHandler(kafkaProducer, cfg, orderStore)

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/order", h.PlaceOrder)
	mux.HandleFunc("/order/", h.GetOrder)
	mux.HandleFunc("/orders", h.ListOrders)
	mux.HandleFunc("/health", h.HealthCheck)

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: mux,
	}

	// Start HTTP server in goroutine
	go func() {
		log.Printf("Starting HTTP server on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
