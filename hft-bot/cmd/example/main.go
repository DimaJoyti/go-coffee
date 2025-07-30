package main

import (
	"log"
	"net/http"

	"github.com/DimaJoyti/go-coffee/hft-bot/internal/shared"
	"github.com/DimaJoyti/go-coffee/hft-bot/pkg/config"
	"github.com/DimaJoyti/go-coffee/hft-bot/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger using the logger package
	logger, err := logger.NewSimple(logger.SimpleConfig{
		ServiceName: "hft-bot-api",
		Environment: cfg.Service.Environment,
		Level:       "info",
		Format:      "console",
	})
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Create HTTP router using standard library
	mux := http.NewServeMux()

	// Add some example routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"message": "HFT Bot API",
			"version": "1.0.0",
			"status": "running"
		}`))
	})

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "pong"}`))
	})

	// Create service with options
	service := shared.NewService(
		"hft-bot-api",
		cfg,
		logger,
		shared.WithHTTPServer(cfg.Service.Port, mux),
		shared.WithMetricsServer(cfg.Monitoring.MetricsPort),
		shared.WithHealthServer(cfg.Monitoring.HealthCheckPort),
	)

	// Run the service
	if err := service.Run(); err != nil {
		logger.Error("Service failed", "error", err)
	}
}
