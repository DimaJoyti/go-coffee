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

	"github.com/DimaJoyti/go-coffee/consumer/config"
	"github.com/DimaJoyti/go-coffee/consumer/kafka"
	"github.com/DimaJoyti/go-coffee/consumer/worker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Initialize structured logger
	logger := initLogger()
	defer logger.Sync()

	logger.Info("Starting Coffee Consumer Service...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Start health check server
	go startHealthServer(logger, cfg)

	// Create worker pool
	workerPool := worker.NewWorkerPool(cfg.Kafka.WorkerPoolSize)
	workerPool.Start()
	defer workerPool.Stop()

	// Create Kafka consumer group
	groupConsumer, err := kafka.NewGroupConsumer(cfg)
	if err != nil {
		logger.Fatal("Failed to create Kafka consumer group", zap.Error(err))
	}
	defer groupConsumer.Close()

	// Create consumer handler
	handler := kafka.NewOrderConsumerHandler()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals - used to stop the process
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Create a goroutine to run the consumer group
	go func() {
		topics := []string{cfg.Kafka.Topic, cfg.Kafka.ProcessedTopic}
		logger.Info("Starting to consume from topics", zap.Strings("topics", topics))

		// Consume messages
		for {
			// Consume messages from both topics
			if err := groupConsumer.Consume(ctx, topics, handler); err != nil {
				logger.Error("Error from consumer", zap.Error(err))
			}

			// Check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				logger.Info("Context cancelled, stopping consumer")
				return
			}
			logger.Info("Consumer restarting...")
		}
	}()

	// Wait for consumer to be ready
	handler.WaitReady()
	logger.Info("Consumer started successfully")

	// Process messages
	go func() {
		for msg := range handler.Messages() {
			// Submit message to worker pool
			workerPool.Submit(msg)
		}
	}()

	// Wait for termination signal
	<-sigchan
	logger.Info("Interrupt detected, shutting down...")
	cancel() // Cancel the context to stop the consumer

	logger.Info("Consumer stopped gracefully")
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

// startHealthServer starts a health check server for the consumer
func startHealthServer(logger *zap.Logger, cfg *config.Config) {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"coffee-consumer","timestamp":"%s"}`,
			time.Now().UTC().Format(time.RFC3339))
	})

	// Readiness check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ready","service":"coffee-consumer","timestamp":"%s"}`,
			time.Now().UTC().Format(time.RFC3339))
	})

	// Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Start health server on a different port
	healthPort := 8081
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
