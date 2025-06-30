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
)

// Application represents the main application
type Application struct {
	config *Config
	server *http.Server
	logger Logger
}

// Config represents application configuration
type Config struct {
	Port         string `env:"PORT" envDefault:"8080"`
	DatabaseURL  string `env:"DATABASE_URL" envDefault:"postgres://localhost/feedback_analyst"`
	RedisURL     string `env:"REDIS_URL" envDefault:"redis://localhost:6379"`
	KafkaBrokers string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
	OpenAIAPIKey string `env:"OPENAI_API_KEY"`
	LogLevel     string `env:"LOG_LEVEL" envDefault:"info"`
	Environment  string `env:"ENVIRONMENT" envDefault:"development"`
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// NewLogger creates a new logger instance
func NewLogger(level string) Logger {
	return &SimpleLogger{level: level}
}

// SimpleLogger implements the Logger interface
type SimpleLogger struct {
	level string
}

func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	if l.level == "debug" {
		log.Printf("[DEBUG] "+msg, args...)
	}
}

func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

func (l *SimpleLogger) Warn(msg string, args ...interface{}) {
	log.Printf("[WARN] "+msg, args...)
}

func (l *SimpleLogger) Error(msg string, err error, args ...interface{}) {
	if err != nil {
		log.Printf("[ERROR] "+msg+": %v", append(args, err)...)
	} else {
		log.Printf("[ERROR] "+msg, args...)
	}
}

// NewApplication creates a new application instance
func NewApplication(config *Config) (*Application, error) {
	// Initialize logger
	logger := NewLogger(config.LogLevel)

	// Initialize HTTP server with basic mux
	mux := http.NewServeMux()

	// Setup basic routes
	mux.HandleFunc("/api/v1/feedback/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Feedback analysis endpoint - Enhanced Feedback Analyst v2.0", "status": "success"}`))
	})

	mux.HandleFunc("/api/v1/feedback/trends", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Feedback trends endpoint - Enhanced Feedback Analyst v2.0", "status": "success"}`))
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := fmt.Sprintf(`{
			"status": "healthy",
			"timestamp": "%s",
			"version": "2.0.0",
			"service": "feedback-analyst-agent"
		}`, time.Now().UTC().Format(time.RFC3339))
		w.Write([]byte(response))
	})

	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: mux,
	}

	return &Application{
		config: config,
		server: server,
		logger: logger,
	}, nil
}

// Start starts the application
func (app *Application) Start() error {
	app.logger.Info("Starting Enhanced Feedback Analyst Agent v2.0", "port", app.config.Port)

	// Start HTTP server in a goroutine
	go func() {
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.logger.Error("Failed to start HTTP server", err)
		}
	}()

	app.logger.Info("Enhanced Feedback Analyst Agent v2.0 started successfully", "port", app.config.Port)
	return nil
}

// Stop stops the application gracefully
func (app *Application) Stop() error {
	app.logger.Info("Stopping Enhanced Feedback Analyst Agent v2.0")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.server.Shutdown(ctx); err != nil {
		app.logger.Error("Failed to shutdown HTTP server gracefully", err)
		return err
	}

	app.logger.Info("Enhanced Feedback Analyst Agent v2.0 stopped successfully")
	return nil
}

func main() {
	fmt.Println("🚀 Starting Enhanced Feedback Analyst Agent v2.0...")

	// Load configuration from environment variables
	config := &Config{
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://localhost/feedback_analyst"),
		RedisURL:     getEnv("REDIS_URL", "redis://localhost:6379"),
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		Environment:  getEnv("ENVIRONMENT", "development"),
	}

	// Create application
	app, err := NewApplication(config)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Start application
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Stop application
	if err := app.Stop(); err != nil {
		log.Fatalf("Failed to stop application: %v", err)
	}

	fmt.Println("✅ Enhanced Feedback Analyst Agent v2.0 shutdown complete")
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
