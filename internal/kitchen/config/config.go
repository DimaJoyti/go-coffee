package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/kitchen/transport"
)

// Config represents the complete kitchen service configuration
type Config struct {
	Service     *ServiceConfig     `json:"service"`
	Database    *DatabaseConfig    `json:"database"`
	Transport   *transport.Config  `json:"transport"`
	Integration *IntegrationConfig `json:"integration"`
	AI          *AIConfig          `json:"ai"`
	Monitoring  *MonitoringConfig  `json:"monitoring"`
}

// ServiceConfig represents service-level configuration
type ServiceConfig struct {
	Name        string        `json:"name" env:"SERVICE_NAME" default:"kitchen-service"`
	Version     string        `json:"version" env:"SERVICE_VERSION" default:"1.0.0"`
	Environment string        `json:"environment" env:"ENVIRONMENT" default:"development"`
	LogLevel    string        `json:"log_level" env:"LOG_LEVEL" default:"info"`
	Timeout     time.Duration `json:"timeout" env:"SERVICE_TIMEOUT" default:"30s"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Redis *RedisConfig `json:"redis"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	URL              string        `json:"url" env:"REDIS_URL" default:"redis://localhost:6379"`
	Password         string        `json:"password" env:"REDIS_PASSWORD"`
	DB               int           `json:"db" env:"REDIS_DB" default:"0"`
	MaxRetries       int           `json:"max_retries" env:"REDIS_MAX_RETRIES" default:"3"`
	DialTimeout      time.Duration `json:"dial_timeout" env:"REDIS_DIAL_TIMEOUT" default:"5s"`
	ReadTimeout      time.Duration `json:"read_timeout" env:"REDIS_READ_TIMEOUT" default:"3s"`
	WriteTimeout     time.Duration `json:"write_timeout" env:"REDIS_WRITE_TIMEOUT" default:"3s"`
	PoolSize         int           `json:"pool_size" env:"REDIS_POOL_SIZE" default:"10"`
	MinIdleConns     int           `json:"min_idle_conns" env:"REDIS_MIN_IDLE_CONNS" default:"5"`
	MaxConnAge       time.Duration `json:"max_conn_age" env:"REDIS_MAX_CONN_AGE" default:"30m"`
	PoolTimeout      time.Duration `json:"pool_timeout" env:"REDIS_POOL_TIMEOUT" default:"4s"`
	IdleTimeout      time.Duration `json:"idle_timeout" env:"REDIS_IDLE_TIMEOUT" default:"5m"`
	IdleCheckFreq    time.Duration `json:"idle_check_freq" env:"REDIS_IDLE_CHECK_FREQ" default:"1m"`
}

// IntegrationConfig represents integration configuration
type IntegrationConfig struct {
	OrderService *OrderServiceConfig `json:"order_service"`
	Events       *EventsConfig       `json:"events"`
}

// OrderServiceConfig represents order service integration configuration
type OrderServiceConfig struct {
	Enabled     bool          `json:"enabled" env:"ORDER_SERVICE_ENABLED" default:"true"`
	Address     string        `json:"address" env:"ORDER_SERVICE_ADDRESS" default:"localhost:50051"`
	Timeout     time.Duration `json:"timeout" env:"ORDER_SERVICE_TIMEOUT" default:"10s"`
	MaxRetries  int           `json:"max_retries" env:"ORDER_SERVICE_MAX_RETRIES" default:"3"`
	SyncEnabled bool          `json:"sync_enabled" env:"ORDER_SERVICE_SYNC_ENABLED" default:"true"`
	SyncInterval time.Duration `json:"sync_interval" env:"ORDER_SERVICE_SYNC_INTERVAL" default:"30s"`
}

// EventsConfig represents event system configuration
type EventsConfig struct {
	Enabled       bool          `json:"enabled" env:"EVENTS_ENABLED" default:"true"`
	BufferSize    int           `json:"buffer_size" env:"EVENTS_BUFFER_SIZE" default:"1000"`
	WorkerCount   int           `json:"worker_count" env:"EVENTS_WORKER_COUNT" default:"5"`
	RetryAttempts int           `json:"retry_attempts" env:"EVENTS_RETRY_ATTEMPTS" default:"3"`
	RetryDelay    time.Duration `json:"retry_delay" env:"EVENTS_RETRY_DELAY" default:"1s"`
}

// AIConfig represents AI/ML configuration
type AIConfig struct {
	Enabled                bool    `json:"enabled" env:"AI_ENABLED" default:"true"`
	OptimizationEnabled    bool    `json:"optimization_enabled" env:"AI_OPTIMIZATION_ENABLED" default:"true"`
	PredictionEnabled      bool    `json:"prediction_enabled" env:"AI_PREDICTION_ENABLED" default:"true"`
	LearningRate           float64 `json:"learning_rate" env:"AI_LEARNING_RATE" default:"0.01"`
	ModelUpdateInterval    time.Duration `json:"model_update_interval" env:"AI_MODEL_UPDATE_INTERVAL" default:"1h"`
	CapacityPredictionDays int     `json:"capacity_prediction_days" env:"AI_CAPACITY_PREDICTION_DAYS" default:"7"`
}

// MonitoringConfig represents monitoring and observability configuration
type MonitoringConfig struct {
	Metrics *MetricsConfig `json:"metrics"`
	Tracing *TracingConfig `json:"tracing"`
	Health  *HealthConfig  `json:"health"`
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled       bool          `json:"enabled" env:"METRICS_ENABLED" default:"true"`
	Port          string        `json:"port" env:"METRICS_PORT" default:"9091"`
	Path          string        `json:"path" env:"METRICS_PATH" default:"/metrics"`
	Interval      time.Duration `json:"interval" env:"METRICS_INTERVAL" default:"15s"`
	Namespace     string        `json:"namespace" env:"METRICS_NAMESPACE" default:"kitchen"`
	Subsystem     string        `json:"subsystem" env:"METRICS_SUBSYSTEM" default:"service"`
}

// TracingConfig represents distributed tracing configuration
type TracingConfig struct {
	Enabled     bool    `json:"enabled" env:"TRACING_ENABLED" default:"false"`
	ServiceName string  `json:"service_name" env:"TRACING_SERVICE_NAME" default:"kitchen-service"`
	Endpoint    string  `json:"endpoint" env:"TRACING_ENDPOINT" default:"http://localhost:14268/api/traces"`
	SampleRate  float64 `json:"sample_rate" env:"TRACING_SAMPLE_RATE" default:"0.1"`
}

// HealthConfig represents health check configuration
type HealthConfig struct {
	Enabled  bool          `json:"enabled" env:"HEALTH_ENABLED" default:"true"`
	Port     string        `json:"port" env:"HEALTH_PORT" default:"8081"`
	Path     string        `json:"path" env:"HEALTH_PATH" default:"/health"`
	Interval time.Duration `json:"interval" env:"HEALTH_INTERVAL" default:"30s"`
	Timeout  time.Duration `json:"timeout" env:"HEALTH_TIMEOUT" default:"5s"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		Service: &ServiceConfig{
			Name:        getEnvOrDefault("SERVICE_NAME", "kitchen-service"),
			Version:     getEnvOrDefault("SERVICE_VERSION", "1.0.0"),
			Environment: getEnvOrDefault("ENVIRONMENT", "development"),
			LogLevel:    getEnvOrDefault("LOG_LEVEL", "info"),
			Timeout:     getEnvOrDefaultDuration("SERVICE_TIMEOUT", 30*time.Second),
		},
		Database: &DatabaseConfig{
			Redis: &RedisConfig{
				URL:              getEnvOrDefault("REDIS_URL", "redis://localhost:6379"),
				Password:         getEnvOrDefault("REDIS_PASSWORD", ""),
				DB:               getEnvOrDefaultInt("REDIS_DB", 0),
				MaxRetries:       getEnvOrDefaultInt("REDIS_MAX_RETRIES", 3),
				DialTimeout:      getEnvOrDefaultDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
				ReadTimeout:      getEnvOrDefaultDuration("REDIS_READ_TIMEOUT", 3*time.Second),
				WriteTimeout:     getEnvOrDefaultDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
				PoolSize:         getEnvOrDefaultInt("REDIS_POOL_SIZE", 10),
				MinIdleConns:     getEnvOrDefaultInt("REDIS_MIN_IDLE_CONNS", 5),
				MaxConnAge:       getEnvOrDefaultDuration("REDIS_MAX_CONN_AGE", 30*time.Minute),
				PoolTimeout:      getEnvOrDefaultDuration("REDIS_POOL_TIMEOUT", 4*time.Second),
				IdleTimeout:      getEnvOrDefaultDuration("REDIS_IDLE_TIMEOUT", 5*time.Minute),
				IdleCheckFreq:    getEnvOrDefaultDuration("REDIS_IDLE_CHECK_FREQ", 1*time.Minute),
			},
		},
		Transport: &transport.Config{
			HTTPPort:     getEnvOrDefault("HTTP_PORT", "8080"),
			GRPCPort:     getEnvOrDefault("GRPC_PORT", "9090"),
			JWTSecret:    getEnvOrDefault("JWT_SECRET", "your-secret-key"),
			EnableCORS:   getEnvOrDefaultBool("ENABLE_CORS", true),
			EnableAuth:   getEnvOrDefaultBool("ENABLE_AUTH", false),
			ReadTimeout:  getEnvOrDefaultInt("READ_TIMEOUT", 30),
			WriteTimeout: getEnvOrDefaultInt("WRITE_TIMEOUT", 30),
			IdleTimeout:  getEnvOrDefaultInt("IDLE_TIMEOUT", 120),
		},
		Integration: &IntegrationConfig{
			OrderService: &OrderServiceConfig{
				Enabled:      getEnvOrDefaultBool("ORDER_SERVICE_ENABLED", true),
				Address:      getEnvOrDefault("ORDER_SERVICE_ADDRESS", "localhost:50051"),
				Timeout:      getEnvOrDefaultDuration("ORDER_SERVICE_TIMEOUT", 10*time.Second),
				MaxRetries:   getEnvOrDefaultInt("ORDER_SERVICE_MAX_RETRIES", 3),
				SyncEnabled:  getEnvOrDefaultBool("ORDER_SERVICE_SYNC_ENABLED", true),
				SyncInterval: getEnvOrDefaultDuration("ORDER_SERVICE_SYNC_INTERVAL", 30*time.Second),
			},
			Events: &EventsConfig{
				Enabled:       getEnvOrDefaultBool("EVENTS_ENABLED", true),
				BufferSize:    getEnvOrDefaultInt("EVENTS_BUFFER_SIZE", 1000),
				WorkerCount:   getEnvOrDefaultInt("EVENTS_WORKER_COUNT", 5),
				RetryAttempts: getEnvOrDefaultInt("EVENTS_RETRY_ATTEMPTS", 3),
				RetryDelay:    getEnvOrDefaultDuration("EVENTS_RETRY_DELAY", 1*time.Second),
			},
		},
		AI: &AIConfig{
			Enabled:                getEnvOrDefaultBool("AI_ENABLED", true),
			OptimizationEnabled:    getEnvOrDefaultBool("AI_OPTIMIZATION_ENABLED", true),
			PredictionEnabled:      getEnvOrDefaultBool("AI_PREDICTION_ENABLED", true),
			LearningRate:           getEnvOrDefaultFloat("AI_LEARNING_RATE", 0.01),
			ModelUpdateInterval:    getEnvOrDefaultDuration("AI_MODEL_UPDATE_INTERVAL", 1*time.Hour),
			CapacityPredictionDays: getEnvOrDefaultInt("AI_CAPACITY_PREDICTION_DAYS", 7),
		},
		Monitoring: &MonitoringConfig{
			Metrics: &MetricsConfig{
				Enabled:   getEnvOrDefaultBool("METRICS_ENABLED", true),
				Port:      getEnvOrDefault("METRICS_PORT", "9091"),
				Path:      getEnvOrDefault("METRICS_PATH", "/metrics"),
				Interval:  getEnvOrDefaultDuration("METRICS_INTERVAL", 15*time.Second),
				Namespace: getEnvOrDefault("METRICS_NAMESPACE", "kitchen"),
				Subsystem: getEnvOrDefault("METRICS_SUBSYSTEM", "service"),
			},
			Tracing: &TracingConfig{
				Enabled:     getEnvOrDefaultBool("TRACING_ENABLED", false),
				ServiceName: getEnvOrDefault("TRACING_SERVICE_NAME", "kitchen-service"),
				Endpoint:    getEnvOrDefault("TRACING_ENDPOINT", "http://localhost:14268/api/traces"),
				SampleRate:  getEnvOrDefaultFloat("TRACING_SAMPLE_RATE", 0.1),
			},
			Health: &HealthConfig{
				Enabled:  getEnvOrDefaultBool("HEALTH_ENABLED", true),
				Port:     getEnvOrDefault("HEALTH_PORT", "8081"),
				Path:     getEnvOrDefault("HEALTH_PATH", "/health"),
				Interval: getEnvOrDefaultDuration("HEALTH_INTERVAL", 30*time.Second),
				Timeout:  getEnvOrDefaultDuration("HEALTH_TIMEOUT", 5*time.Second),
			},
		},
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Service.Name == "" {
		return fmt.Errorf("service name is required")
	}

	if c.Database.Redis.URL == "" {
		return fmt.Errorf("Redis URL is required")
	}

	if c.Transport.HTTPPort == "" {
		return fmt.Errorf("HTTP port is required")
	}

	if c.Transport.GRPCPort == "" {
		return fmt.Errorf("gRPC port is required")
	}

	if c.Integration.OrderService.Enabled && c.Integration.OrderService.Address == "" {
		return fmt.Errorf("order service address is required when integration is enabled")
	}

	return nil
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Service.Environment == "development"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Service.Environment == "production"
}

// Helper functions for environment variable parsing

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvOrDefaultBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvOrDefaultFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvOrDefaultDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
