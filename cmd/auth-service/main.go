package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/DimaJoyti/go-coffee/internal/auth/infrastructure/container"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"gopkg.in/yaml.v3"
)

const (
	serviceName     = "auth-service"
	serviceVersion  = "1.0.0"
	defaultHTTPPort = 8080
	minPort         = 1
	maxPort         = 65535

	// Timeouts
	shutdownTimeout   = 30 * time.Second // Time to wait for graceful shutdown
	grpcShutdownDelay = 5 * time.Second  // Delay before shutting down gRPC server
	dbShutdownTimeout = 10 * time.Second // Time to wait for DB connections to close
	httpShutdownDelay = 1 * time.Second  // Delay after starting shutdown
)

var (
	configPath  = flag.String("config", "configs/auth-service.yaml", "Path to configuration file")
	logLevel    = flag.String("log-level", "", "Log level (debug, info, warn, error)")
	port        = flag.Int("port", 0, "HTTP server port (overrides config)")
	showVersion = flag.Bool("version", false, "Show version information")
	showHelp    = flag.Bool("help", false, "Show usage information")
	startTime   = time.Now() // Track service start time for uptime reporting

	// Metrics
	requestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_requests_total",
			Help: "Total number of requests processed by the auth service",
		},
		[]string{"endpoint", "method", "status"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_request_duration_seconds",
			Help:    "Time taken to process auth requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint", "method"},
	)

	dependencyHealth = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "auth_dependency_health",
			Help: "Health status of auth service dependencies (1 for healthy, 0 for unhealthy)",
		},
		[]string{"dependency"},
	)
)

// HealthStatus represents the health check response
type HealthStatus struct {
	Status    string          `json:"status"`
	Service   string          `json:"service"`
	Version   string          `json:"version"`
	Timestamp string          `json:"timestamp"`
	Uptime    string          `json:"uptime"`
	Details   map[string]bool `json:"details"`
}

// checkHealth performs health checks on all dependencies
func checkHealth(container *container.Container) *HealthStatus {
	details := map[string]bool{
		"database": true,
		"redis":    true,
		"http":     true,
	}

	// Check database connection
	if err := container.DB.Ping(); err != nil {
		details["database"] = false
	}

	// Check Redis connection via cache service
	if healthChecker, ok := container.CacheService.(interface{ Health() error }); ok {
		if err := healthChecker.Health(); err != nil {
			details["redis"] = false
		}
	}

	status := "healthy"
	for _, healthy := range details {
		if !healthy {
			status = "degraded"
			break
		}
	}

	return &HealthStatus{
		Status:    status,
		Service:   serviceName,
		Version:   serviceVersion,
		Timestamp: time.Now().Format(time.RFC3339),
		Uptime:    time.Since(startTime).String(),
		Details:   details,
	}
}

// initLogging initializes the logger with the appropriate configuration
func initLogging() *logger.Logger {
	logConfig := logger.DefaultConfig()
	if *logLevel != "" {
		switch *logLevel {
		case "debug":
			logConfig.Level = logger.DebugLevel
		case "info":
			logConfig.Level = logger.InfoLevel
		case "warn":
			logConfig.Level = logger.WarnLevel
		case "error":
			logConfig.Level = logger.ErrorLevel
		default:
			fmt.Printf("Warning: Invalid log level %q, using default\n", *logLevel)
		}
	}

	log := logger.NewLogger(logConfig)
	return log.WithFields(map[string]interface{}{
		"service": serviceName,
		"version": serviceVersion,
		"config":  *configPath,
	})
}

func main() {
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	if *showHelp {
		printUsage()
		os.Exit(0)
	}

	// Initialize logger
	log := initLogging()

	log.Info("Starting auth service")

	// Load auth service configuration
	appConfig, err := loadConfig(*configPath)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Failed to load configuration")
	}

	// Override with command line flags
	if *port != 0 {
		appConfig.HTTP.Port = *port
	}

	// Create and initialize auth service container
	authContainer, err := container.NewContainer(appConfig)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("Failed to create auth container")
	}
	defer authContainer.Close()

	// Get HTTP server from container
	httpServer := authContainer.HTTPServer

	// Start HTTP server
	serverErr := make(chan error, 1)
	go func() {
		log.WithField("port", appConfig.HTTP.Port).Info("Starting HTTP server")
		if err := httpServer.Start(); err != nil {
			serverErr <- fmt.Errorf("HTTP server failed: %w", err)
		}
	}()

	// Wait for interrupt signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.WithField("error", err.Error()).Error("Server error")
	case sig := <-sigChan:
		log.WithField("signal", sig.String()).Info("Received shutdown signal")
	}

	// Graceful shutdown
	log.Info("Starting graceful shutdown...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// First stop accepting new requests
	log.Info("Stopping HTTP server...")
	time.Sleep(httpShutdownDelay) // Allow in-flight requests to complete
	if err := httpServer.Stop(shutdownCtx); err != nil {
		log.WithField("error", err.Error()).Error("Failed to shutdown HTTP server gracefully")
	}

	// Update health status before closing dependencies
	healthStatus := checkHealth(authContainer)
	for dep, healthy := range healthStatus.Details {
		if healthy {
			dependencyHealth.WithLabelValues(dep).Set(1)
		} else {
			dependencyHealth.WithLabelValues(dep).Set(0)
		}
	}

	// Close container resources
	log.Info("Closing container resources...")
	if err := authContainer.Close(); err != nil {
		log.WithField("error", err.Error()).Error("Failed to close container resources")
	}

	log.InfoWithFields("Auth service stopped successfully",
		logger.Duration("total_shutdown_time", time.Since(startTime)),
	)
}

// getPortFromFlags returns the port from command line flags or default
func getPortFromFlags() int {
	if *port != 0 {
		return *port
	}
	return defaultHTTPPort // default port
}

// loadConfig loads configuration from file with environment variable overrides
func loadConfig(configPath string) (*container.Config, error) {
	// Start with default configuration
	cfg := container.DefaultConfig()

	if err := loadConfigFromFile(cfg, configPath); err != nil {
		return nil, err
	}

	applyEnvironmentOverrides(cfg)

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// loadConfigFromFile loads configuration from YAML file if it exists
func loadConfigFromFile(cfg *container.Config, configPath string) error {
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("reading config file: %w", err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("parsing config file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("checking config file: %w", err)
	}
	return nil
}

// applyEnvironmentOverrides applies environment variable overrides to the configuration
func applyEnvironmentOverrides(cfg *container.Config) {
	applyDatabaseEnvOverrides(cfg)
	applyRedisEnvOverrides(cfg)
	applyHTTPEnvOverrides(cfg)
	applyJWTEnvOverrides(cfg)
}

// applyDatabaseEnvOverrides applies database-related environment variables
func applyDatabaseEnvOverrides(cfg *container.Config) {
	if v := os.Getenv("AUTH_DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("AUTH_DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Database.Port = port
		}
	}
	if v := os.Getenv("AUTH_DB_NAME"); v != "" {
		cfg.Database.Database = v
	}
	if v := os.Getenv("AUTH_DB_USER"); v != "" {
		cfg.Database.Username = v
	}
	if v := os.Getenv("AUTH_DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
}

// applyRedisEnvOverrides applies Redis-related environment variables
func applyRedisEnvOverrides(cfg *container.Config) {
	if v := os.Getenv("AUTH_REDIS_HOST"); v != "" {
		cfg.Redis.Host = v
	}
	if v := os.Getenv("AUTH_REDIS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Redis.Port = port
		}
	}
	if v := os.Getenv("AUTH_REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
}

// applyHTTPEnvOverrides applies HTTP-related environment variables
func applyHTTPEnvOverrides(cfg *container.Config) {
	if v := os.Getenv("AUTH_HTTP_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.HTTP.Port = port
		}
	}
}

// applyJWTEnvOverrides applies JWT-related environment variables
func applyJWTEnvOverrides(cfg *container.Config) {
	if v := os.Getenv("AUTH_JWT_SECRET"); v != "" {
		cfg.JWT.SecretKey = v
	}
}

// validateConfig validates the configuration
func validateConfig(cfg *container.Config) error {
	// Database validation
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if cfg.Database.Port == 0 {
		return fmt.Errorf("database port is required")
	}
	if cfg.Database.Port < minPort || cfg.Database.Port > maxPort {
		return fmt.Errorf("database port must be between %d and %d", minPort, maxPort)
	}
	if cfg.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if cfg.Database.Username == "" {
		return fmt.Errorf("database username is required")
	}

	// JWT validation
	if cfg.JWT.SecretKey == "" {
		return fmt.Errorf("JWT secret key is required")
	}
	if len(cfg.JWT.SecretKey) < 32 {
		return fmt.Errorf("JWT secret key must be at least 32 characters")
	}

	// HTTP validation
	if cfg.HTTP.Port < minPort || cfg.HTTP.Port > maxPort {
		return fmt.Errorf("HTTP port must be between %d and %d", minPort, maxPort)
	}

	// Redis validation
	if cfg.Redis.Host == "" {
		return fmt.Errorf("Redis host is required")
	}
	if cfg.Redis.Port < minPort || cfg.Redis.Port > maxPort {
		return fmt.Errorf("Redis port must be between %d and %d", minPort, maxPort)
	}

	return nil
}

// Version information
func printVersion() {
	fmt.Printf("%s version %s\n", serviceName, serviceVersion)
}

// Usage information
func printUsage() {
	fmt.Printf("Usage: %s [options]\n\n", os.Args[0])
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println("\nEnvironment Variables:")
	fmt.Println("  AUTH_DB_HOST          Database host")
	fmt.Println("  AUTH_DB_PORT          Database port")
	fmt.Println("  AUTH_DB_NAME          Database name")
	fmt.Println("  AUTH_DB_USER          Database username")
	fmt.Println("  AUTH_DB_PASSWORD      Database password")
	fmt.Println("  AUTH_JWT_SECRET       JWT secret key")
	fmt.Println("  AUTH_REDIS_HOST       Redis host")
	fmt.Println("  AUTH_REDIS_PORT       Redis port")
	fmt.Println("  AUTH_REDIS_PASSWORD   Redis password")
	fmt.Println("  AUTH_HTTP_PORT        HTTP server port")
	fmt.Println("  AUTH_LOG_LEVEL        Log level (debug, info, warn, error)")
}
