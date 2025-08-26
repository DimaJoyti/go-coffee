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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Initialize structured logger
	logger := initLogger()
	defer logger.Sync()

	logger.Info("Starting Coffee Producer Service...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Create Kafka producer
	kafkaProducer, err := kafka.NewProducer(cfg)
	if err != nil {
		logger.Fatal("Failed to create Kafka producer", zap.Error(err))
	}
	defer kafkaProducer.Close()

	// Create order store
	orderStore := store.NewInMemoryOrderStore()

	// Create HTTP handlers
	h := handler.NewHandler(kafkaProducer, cfg, orderStore)

	// Setup HTTP routes with middleware
	mux := http.NewServeMux()

	// Business endpoints
	mux.HandleFunc("/order", h.PlaceOrder)
	mux.HandleFunc("/order/", h.GetOrder)
	mux.HandleFunc("/orders", h.ListOrders)

	// Observability endpoints
	mux.HandleFunc("/health", h.HealthCheck)
	mux.HandleFunc("/ready", h.ReadinessCheck)
	mux.Handle("/metrics", promhttp.Handler())

	// Create HTTP server with enhanced configuration
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in goroutine
	go func() {
		logger.Info("Starting HTTP server", zap.Int("port", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	logger.Info("Server stopped gracefully")
}

// initLogger initializes a structured logger with appropriate configuration
func initLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stacktrace"

	logger, err := config.Build()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	return logger
}
