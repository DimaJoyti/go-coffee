package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"kafka_producer/config"
	"kafka_producer/grpc"
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

	// Create HTTP handler
	h := handler.NewHandler(kafkaProducer, cfg, orderStore)

	// Create HTTP router
	mux := http.NewServeMux()

	// Register HTTP routes
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
	httpHandler := middleware.RecoverMiddleware(
		middleware.LoggingMiddleware(
			middleware.RequestIDMiddleware(
				middleware.CORSMiddleware(mux),
			),
		),
	)

	// Create gRPC server
	grpcServer := grpc.NewCoffeeServiceServer(kafkaProducer, cfg, orderStore)

	// Start gRPC server in a goroutine
	go func() {
		grpcPort := ":50051" // Порт для gRPC сервера
		log.Printf("Starting gRPC server on %s", grpcPort)
		if err := grpc.StartGRPCServer(grpcServer, grpcPort); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Start HTTP server in a goroutine
	go func() {
		serverAddr := ":" + strconv.Itoa(cfg.Server.Port)
		log.Printf("Starting HTTP server on %s", serverAddr)
		if err := http.ListenAndServe(serverAddr, httpHandler); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")
	log.Println("Servers exited properly")
}
