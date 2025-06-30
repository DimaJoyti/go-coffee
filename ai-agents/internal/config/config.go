package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go-coffee-ai-agents/internal/observability"
	"go-coffee-ai-agents/internal/resilience"
)

// Environment represents the deployment environment
type Environment string

const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
	Testing     Environment = "testing"
)

// Config represents the complete application configuration
type Config struct {
	// Service configuration
	Service ServiceConfig `yaml:"service"`
	
	// Server configuration
	Server ServerConfig `yaml:"server"`
	
	// Database configuration
	Database DatabaseConfig `yaml:"database"`
	
	// Kafka configuration
	Kafka KafkaConfig `yaml:"kafka"`
	
	// AI providers configuration
	AI AIConfig `yaml:"ai"`
	
	// External services configuration
	External ExternalConfig `yaml:"external"`
	
	// Security configuration
	Security SecurityConfig `yaml:"security"`
	
	// Observability configuration
	Observability observability.TelemetryConfig `yaml:"observability"`
	
	// Resilience configuration
	Resilience resilience.ResilienceConfig `yaml:"resilience"`
	
	// Feature flags
	Features FeatureConfig `yaml:"features"`
	
	// Environment-specific settings
	Environment Environment `yaml:"environment"`
}

// ServiceConfig contains service-level configuration
type ServiceConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Port        int    `yaml:"port"`
	Host        string `yaml:"host"`
	BasePath    string `yaml:"base_path"`
	Debug       bool   `yaml:"debug"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	MaxHeaderBytes  int           `yaml:"max_header_bytes"`
	EnableCORS      bool          `yaml:"enable_cors"`
	CORSOrigins     []string      `yaml:"cors_origins"`
	EnableMetrics   bool          `yaml:"enable_metrics"`
	MetricsPath     string        `yaml:"metrics_path"`
	HealthPath      string        `yaml:"health_path"`
}

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	Driver          string        `yaml:"driver"`
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Database        string        `yaml:"database"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	SSLMode         string        `yaml:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
	MigrationsPath  string        `yaml:"migrations_path"`
	EnableLogging   bool          `yaml:"enable_logging"`
}

// KafkaConfig contains Kafka configuration
type KafkaConfig struct {
	Brokers         []string      `yaml:"brokers"`
	ClientID        string        `yaml:"client_id"`
	GroupID         string        `yaml:"group_id"`
	EnableSASL      bool          `yaml:"enable_sasl"`
	SASLMechanism   string        `yaml:"sasl_mechanism"`
	SASLUsername    string        `yaml:"sasl_username"`
	SASLPassword    string        `yaml:"sasl_password"`
	EnableTLS       bool          `yaml:"enable_tls"`
	TLSSkipVerify   bool          `yaml:"tls_skip_verify"`
	ConnectTimeout  time.Duration `yaml:"connect_timeout"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	BatchSize       int           `yaml:"batch_size"`
	BatchTimeout    time.Duration `yaml:"batch_timeout"`
	RetryMax        int           `yaml:"retry_max"`
	RetryBackoff    time.Duration `yaml:"retry_backoff"`
	Topics          TopicsConfig  `yaml:"topics"`
}

// TopicsConfig contains Kafka topic configuration
type TopicsConfig struct {
	BeverageCreated     string `yaml:"beverage_created"`
	BeverageUpdated     string `yaml:"beverage_updated"`
	TaskCreated         string `yaml:"task_created"`
	TaskUpdated         string `yaml:"task_updated"`
	NotificationSent    string `yaml:"notification_sent"`
	AIRequestCompleted  string `yaml:"ai_request_completed"`
	SystemEvent         string `yaml:"system_event"`
}

// AIConfig contains AI provider configurations
type AIConfig struct {
	DefaultProvider string                    `yaml:"default_provider"`
	Providers       map[string]AIProviderConfig `yaml:"providers"`
	RateLimits      AIRateLimitsConfig        `yaml:"rate_limits"`
	Timeouts        AITimeoutsConfig          `yaml:"timeouts"`
}

// AIProviderConfig contains configuration for a specific AI provider
type AIProviderConfig struct {
	Enabled     bool              `yaml:"enabled"`
	APIKey      string            `yaml:"api_key"`
	BaseURL     string            `yaml:"base_url"`
	Model       string            `yaml:"model"`
	MaxTokens   int               `yaml:"max_tokens"`
	Temperature float64           `yaml:"temperature"`
	Headers     map[string]string `yaml:"headers"`
	Timeout     time.Duration     `yaml:"timeout"`
}

// AIRateLimitsConfig contains rate limiting configuration for AI providers
type AIRateLimitsConfig struct {
	RequestsPerMinute int           `yaml:"requests_per_minute"`
	TokensPerMinute   int           `yaml:"tokens_per_minute"`
	BurstSize         int           `yaml:"burst_size"`
	CooldownPeriod    time.Duration `yaml:"cooldown_period"`
}

// AITimeoutsConfig contains timeout configuration for AI operations
type AITimeoutsConfig struct {
	AnalyzeIngredients    time.Duration `yaml:"analyze_ingredients"`
	GenerateDescription   time.Duration `yaml:"generate_description"`
	SuggestImprovements   time.Duration `yaml:"suggest_improvements"`
	GenerateRecipe        time.Duration `yaml:"generate_recipe"`
}

// ExternalConfig contains external service configurations
type ExternalConfig struct {
	ClickUp      ClickUpConfig      `yaml:"clickup"`
	Slack        SlackConfig        `yaml:"slack"`
	GoogleSheets GoogleSheetsConfig `yaml:"google_sheets"`
	Email        EmailConfig        `yaml:"email"`
}

// ClickUpConfig contains ClickUp API configuration
type ClickUpConfig struct {
	Enabled     bool          `yaml:"enabled"`
	APIKey      string        `yaml:"api_key"`
	BaseURL     string        `yaml:"base_url"`
	TeamID      string        `yaml:"team_id"`
	SpaceID     string        `yaml:"space_id"`
	FolderID    string        `yaml:"folder_id"`
	ListID      string        `yaml:"list_id"`
	Timeout     time.Duration `yaml:"timeout"`
	RetryCount  int           `yaml:"retry_count"`
	RateLimit   int           `yaml:"rate_limit"`
}

// SlackConfig contains Slack API configuration
type SlackConfig struct {
	Enabled      bool          `yaml:"enabled"`
	BotToken     string        `yaml:"bot_token"`
	AppToken     string        `yaml:"app_token"`
	SigningSecret string       `yaml:"signing_secret"`
	DefaultChannel string      `yaml:"default_channel"`
	Timeout      time.Duration `yaml:"timeout"`
	RetryCount   int           `yaml:"retry_count"`
}

// GoogleSheetsConfig contains Google Sheets API configuration
type GoogleSheetsConfig struct {
	Enabled           bool          `yaml:"enabled"`
	CredentialsPath   string        `yaml:"credentials_path"`
	SpreadsheetID     string        `yaml:"spreadsheet_id"`
	DefaultSheetName  string        `yaml:"default_sheet_name"`
	Timeout           time.Duration `yaml:"timeout"`
	RetryCount        int           `yaml:"retry_count"`
}

// EmailConfig contains email service configuration
type EmailConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Provider   string `yaml:"provider"` // smtp, sendgrid, ses
	SMTPHost   string `yaml:"smtp_host"`
	SMTPPort   int    `yaml:"smtp_port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	FromEmail  string `yaml:"from_email"`
	FromName   string `yaml:"from_name"`
	EnableTLS  bool   `yaml:"enable_tls"`
}

// SecurityConfig contains security-related configuration
type SecurityConfig struct {
	JWT          JWTConfig          `yaml:"jwt"`
	API          APISecurityConfig  `yaml:"api"`
	CORS         CORSConfig         `yaml:"cors"`
	RateLimit    RateLimitConfig    `yaml:"rate_limit"`
	Encryption   EncryptionConfig   `yaml:"encryption"`
}

// JWTConfig contains JWT configuration
type JWTConfig struct {
	SecretKey      string        `yaml:"secret_key"`
	Issuer         string        `yaml:"issuer"`
	Audience       string        `yaml:"audience"`
	ExpirationTime time.Duration `yaml:"expiration_time"`
	RefreshTime    time.Duration `yaml:"refresh_time"`
	Algorithm      string        `yaml:"algorithm"`
}

// APISecurityConfig contains API security configuration
type APISecurityConfig struct {
	EnableAPIKeys    bool     `yaml:"enable_api_keys"`
	RequireHTTPS     bool     `yaml:"require_https"`
	AllowedIPs       []string `yaml:"allowed_ips"`
	BlockedIPs       []string `yaml:"blocked_ips"`
	EnableRateLimit  bool     `yaml:"enable_rate_limit"`
	MaxRequestSize   int64    `yaml:"max_request_size"`
}

// CORSConfig contains CORS configuration
type CORSConfig struct {
	Enabled          bool     `yaml:"enabled"`
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	Enabled        bool          `yaml:"enabled"`
	RequestsPerMin int           `yaml:"requests_per_min"`
	BurstSize      int           `yaml:"burst_size"`
	CleanupPeriod  time.Duration `yaml:"cleanup_period"`
	Storage        string        `yaml:"storage"` // memory, redis
	RedisURL       string        `yaml:"redis_url"`
}

// EncryptionConfig contains encryption configuration
type EncryptionConfig struct {
	Algorithm    string `yaml:"algorithm"`
	KeySize      int    `yaml:"key_size"`
	SecretKey    string `yaml:"secret_key"`
	EnableAtRest bool   `yaml:"enable_at_rest"`
	EnableInTransit bool `yaml:"enable_in_transit"`
}

// FeatureConfig contains feature flag configuration
type FeatureConfig struct {
	EnableAI              bool `yaml:"enable_ai"`
	EnableTaskCreation    bool `yaml:"enable_task_creation"`
	EnableNotifications   bool `yaml:"enable_notifications"`
	EnableMetrics         bool `yaml:"enable_metrics"`
	EnableTracing         bool `yaml:"enable_tracing"`
	EnableAuditLogging    bool `yaml:"enable_audit_logging"`
	EnableCaching         bool `yaml:"enable_caching"`
	EnableRateLimiting    bool `yaml:"enable_rate_limiting"`
	EnableCircuitBreaker  bool `yaml:"enable_circuit_breaker"`
	EnableRetry           bool `yaml:"enable_retry"`
	EnableHealthChecks    bool `yaml:"enable_health_checks"`
	EnableGracefulShutdown bool `yaml:"enable_graceful_shutdown"`
}

// GetDSN returns the database connection string
func (dc *DatabaseConfig) GetDSN() string {
	switch dc.Driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			dc.Host, dc.Port, dc.Username, dc.Password, dc.Database, dc.SSLMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
			dc.Username, dc.Password, dc.Host, dc.Port, dc.Database)
	default:
		return ""
	}
}

// GetBrokerList returns Kafka brokers as a comma-separated string
func (kc *KafkaConfig) GetBrokerList() string {
	return strings.Join(kc.Brokers, ",")
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Environment == Production
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Environment == Development
}

// IsStaging returns true if the environment is staging
func (c *Config) IsStaging() bool {
	return c.Environment == Staging
}

// IsTesting returns true if the environment is testing
func (c *Config) IsTesting() bool {
	return c.Environment == Testing
}

// GetServerAddress returns the server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Service.Host, c.Service.Port)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate service configuration
	if c.Service.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if c.Service.Port <= 0 || c.Service.Port > 65535 {
		return fmt.Errorf("invalid service port: %d", c.Service.Port)
	}

	// Validate database configuration
	if c.Database.Driver == "" {
		return fmt.Errorf("database driver is required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	// Validate Kafka configuration
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("at least one Kafka broker is required")
	}

	// Validate AI configuration
	if c.AI.DefaultProvider == "" {
		return fmt.Errorf("default AI provider is required")
	}
	if _, exists := c.AI.Providers[c.AI.DefaultProvider]; !exists {
		return fmt.Errorf("default AI provider '%s' not found in providers", c.AI.DefaultProvider)
	}

	// Validate security configuration
	if c.Security.JWT.SecretKey == "" && c.IsProduction() {
		return fmt.Errorf("JWT secret key is required in production")
	}

	return nil
}

// GetEnvOrDefault returns environment variable value or default
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvAsInt returns environment variable as integer or default
func GetEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvAsBool returns environment variable as boolean or default
func GetEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetEnvAsDuration returns environment variable as duration or default
func GetEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// GetConfigPath returns the configuration file path
func GetConfigPath() string {
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		return configPath
	}
	
	// Default paths to check
	paths := []string{
		"./config.yaml",
		"./configs/config.yaml",
		"/etc/go-coffee/config.yaml",
		filepath.Join(os.Getenv("HOME"), ".go-coffee", "config.yaml"),
	}
	
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	
	return "./config.yaml" // Default fallback
}
