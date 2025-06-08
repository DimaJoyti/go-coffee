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

	"github.com/DimaJoyti/go-coffee/internal/payment"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

func main() {
	// Load environment variables
	config.AutoLoadEnvFiles()

	// Initialize logger
	logger := logger.New("payment-service")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize payment service
	paymentService, err := payment.NewService(cfg, logger)
	if err != nil {
		log.Fatalf("Failed to initialize payment service: %v", err)
	}

	// Setup HTTP server
	mux := http.NewServeMux()

	// Setup routes
	payment.SetupRoutes(mux, paymentService)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"payment-service","time":"%s"}`, time.Now().UTC().Format(time.RFC3339))
	})

	// Start server
	port := fmt.Sprintf(":%d", cfg.Server.PaymentServicePort)
	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	logger.Info("Payment service started", "port", cfg.Server.PaymentServicePort)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down payment service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Payment service stopped")
}
