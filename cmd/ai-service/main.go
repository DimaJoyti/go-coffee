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

	"github.com/DimaJoyti/go-coffee/internal/ai"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

func main() {
	// Load environment variables
	config.AutoLoadEnvFiles()

	// Initialize logger
	logger := logger.New("ai-service")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize AI service
	aiService, err := ai.NewService(cfg, logger)
	if err != nil {
		log.Fatalf("Failed to initialize AI service: %v", err)
	}

	// Setup HTTP server
	mux := http.NewServeMux()

	// Setup routes
	ai.SetupRoutes(mux, aiService)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"ai-service","time":"%s"}`, time.Now().UTC().Format(time.RFC3339))
	})

	// Start server
	port := fmt.Sprintf(":%d", cfg.Server.AISearchPort) // Reusing AI search port
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

	logger.Info("AI service started on port %d", cfg.Server.AISearchPort)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down AI service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("AI service stopped")
}
