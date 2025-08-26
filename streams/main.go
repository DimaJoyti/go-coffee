package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DimaJoyti/go-coffee/streams/config"
	"github.com/DimaJoyti/go-coffee/streams/kafka"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Initialize structured logger
	logger := initLogger()
	defer logger.Sync()

	logger.Info("Starting Kafka Streams processor...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Start health check server
	go startHealthServer(logger, cfg)

	// Create stream processor
	processor, err := kafka.NewStreamProcessor(cfg)
	if err != nil {
		logger.Fatal("Failed to create stream processor", zap.Error(err))
	}

	// Start the processor
	if err := processor.Start(); err != nil {
		logger.Fatal("Failed to start stream processor", zap.Error(err))
	}
	logger.Info("Stream processor started successfully")

	// Set up signal handler for graceful shutdown
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	sig := <-sigchan
	logger.Info("Received signal, shutting down...", zap.String("signal", sig.String()))

	// Stop the processor
	processor.Stop()
	logger.Info("Stream processor stopped gracefully")
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

// startHealthServer starts a health check server for the streams processor
func startHealthServer(logger *zap.Logger, cfg *config.Config) {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"coffee-streams","timestamp":"%s"}`,
			time.Now().UTC().Format(time.RFC3339))
	})

	// Readiness check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ready","service":"coffee-streams","timestamp":"%s"}`,
			time.Now().UTC().Format(time.RFC3339))
	})

	// Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Start health server on a different port
	healthPort := 8082
	if cfg.Server.HealthPort != 0 {
		healthPort = cfg.Server.HealthPort
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", healthPort),
		Handler: mux,
	}

	logger.Info("Starting health check server", zap.Int("port", healthPort))
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Health server error", zap.Error(err))
	}
}
