package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete auth service configuration
type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	Redis        RedisConfig        `mapstructure:"redis"`
	Security     SecurityConfig     `mapstructure:"security"`
	RateLimiting RateLimitingConfig `mapstructure:"rate_limiting"`
	CORS         CORSConfig         `mapstructure:"cors"`
	Logging      LoggingConfig      `mapstructure:"logging"`
	Monitoring   MonitoringConfig   `mapstructure:"monitoring"`
	TLS          TLSConfig          `mapstructure:"tls"`
	Features     FeaturesConfig     `mapstructure:"features"`
	Environment  string             `mapstructure:"environment"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host             string        `mapstructure:"host"`
	HTTPPort         int           `mapstructure:"http_port"`
	GRPCPort         int           `mapstructure:"grpc_port"`
	ReadTimeout      time.Duration `mapstructure:"read_timeout"`
	WriteTimeout     time.Duration `mapstructure:"write_timeout"`
	IdleTimeout      time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout  time.Duration `mapstructure:"shutdown_timeout"`
	MaxHeaderBytes   int           `mapstructure:"max_header_bytes"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	URL             string        `mapstructure:"url"`
	Password        string        `mapstructure:"password"`
	DB              int           `mapstructure:"db"`
	MaxRetries      int           `mapstructure:"max_retries"`
	PoolSize        int           `mapstructure:"pool_size"`
	MinIdleConns    int           `mapstructure:"min_idle_conns"`
	DialTimeout     time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	PoolTimeout     time.Duration `mapstructure:"pool_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	MaxConnAge      time.Duration `mapstructure:"max_conn_age"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	JWT      JWTConfig      `mapstructure:"jwt"`
	Password PasswordConfig `mapstructure:"password"`
	Account  AccountConfig  `mapstructure:"account"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret          string        `mapstructure:"secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
	Issuer          string        `mapstructure:"issuer"`
	Audience        string        `mapstructure:"audience"`
	Algorithm       string        `mapstructure:"algorithm"`
}

// PasswordConfig represents password configuration
type PasswordConfig struct {
	BcryptCost int                  `mapstructure:"bcrypt_cost"`
	Policy     PasswordPolicyConfig `mapstructure:"policy"`
}

// PasswordPolicyConfig represents password policy configuration
type PasswordPolicyConfig struct {
	MinLength           int      `mapstructure:"min_length"`
	MaxLength           int      `mapstructure:"max_length"`
	RequireUppercase    bool     `mapstructure:"require_uppercase"`
	RequireLowercase    bool     `mapstructure:"require_lowercase"`
	RequireNumbers      bool     `mapstructure:"require_numbers"`
	RequireSymbols      bool     `mapstructure:"require_symbols"`
	ForbiddenPatterns   []string `mapstructure:"forbidden_patterns"`
}

// AccountConfig represents account security configuration
type AccountConfig struct {
	MaxLoginAttempts     int           `mapstructure:"max_login_attempts"`
	LockoutDuration      time.Duration `mapstructure:"lockout_duration"`
	SessionTimeout       time.Duration `mapstructure:"session_timeout"`
	MaxSessionsPerUser   int           `mapstructure:"max_sessions_per_user"`
}

// RateLimitingConfig represents rate limiting configuration
type RateLimitingConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	RequestsPerMinute int           `mapstructure:"requests_per_minute"`
	BurstSize         int           `mapstructure:"burst_size"`
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	Port    int           `mapstructure:"port"`
	Path    string        `mapstructure:"path"`
	Tracing TracingConfig `mapstructure:"tracing"`
}

// TracingConfig represents tracing configuration
type TracingConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	ServiceName string `mapstructure:"service_name"`
	Endpoint    string `mapstructure:"endpoint"`
	SampleRate  float64 `mapstructure:"sample_rate"`
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	CertFile     string   `mapstructure:"cert_file"`
	KeyFile      string   `mapstructure:"key_file"`
	MinVersion   string   `mapstructure:"min_version"`
	CipherSuites []string `mapstructure:"cipher_suites"`
}

// FeaturesConfig represents feature flags configuration
type FeaturesConfig struct {
	RegistrationEnabled           bool `mapstructure:"registration_enabled"`
	PasswordResetEnabled          bool `mapstructure:"password_reset_enabled"`
	MultiFactorAuthEnabled        bool `mapstructure:"multi_factor_auth_enabled"`
	SessionAnalyticsEnabled       bool `mapstructure:"session_analytics_enabled"`
	SecurityNotificationsEnabled  bool `mapstructure:"security_notifications_enabled"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./cmd/auth-service/config")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Set default values
	setDefaults()

	// Enable environment variable support
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.http_port", 8080)
	viper.SetDefault("server.grpc_port", 50053)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")
	viper.SetDefault("server.shutdown_timeout", "30s")
	viper.SetDefault("server.max_header_bytes", 1048576)

	// Redis defaults
	viper.SetDefault("redis.url", "redis://localhost:6379")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)
	viper.SetDefault("redis.dial_timeout", "5s")
	viper.SetDefault("redis.read_timeout", "3s")
	viper.SetDefault("redis.write_timeout", "3s")
	viper.SetDefault("redis.pool_timeout", "4s")
	viper.SetDefault("redis.idle_timeout", "5m")
	viper.SetDefault("redis.max_conn_age", "30m")

	// Security defaults
	viper.SetDefault("security.jwt.access_token_ttl", "15m")
	viper.SetDefault("security.jwt.refresh_token_ttl", "168h")
	viper.SetDefault("security.jwt.issuer", "auth-service")
	viper.SetDefault("security.jwt.audience", "go-coffee-users")
	viper.SetDefault("security.jwt.algorithm", "HS256")
	viper.SetDefault("security.password.bcrypt_cost", 12)
	viper.SetDefault("security.password.policy.min_length", 8)
	viper.SetDefault("security.password.policy.max_length", 128)
	viper.SetDefault("security.password.policy.require_uppercase", true)
	viper.SetDefault("security.password.policy.require_lowercase", true)
	viper.SetDefault("security.password.policy.require_numbers", true)
	viper.SetDefault("security.password.policy.require_symbols", true)
	viper.SetDefault("security.account.max_login_attempts", 5)
	viper.SetDefault("security.account.lockout_duration", "30m")
	viper.SetDefault("security.account.session_timeout", "24h")
	viper.SetDefault("security.account.max_sessions_per_user", 10)

	// Rate limiting defaults
	viper.SetDefault("rate_limiting.enabled", true)
	viper.SetDefault("rate_limiting.requests_per_minute", 60)
	viper.SetDefault("rate_limiting.burst_size", 10)
	viper.SetDefault("rate_limiting.cleanup_interval", "1m")

	// CORS defaults
	viper.SetDefault("cors.allowed_origins", []string{"*"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"*"})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.max_age", 86400)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	// Monitoring defaults
	viper.SetDefault("monitoring.enabled", true)
	viper.SetDefault("monitoring.port", 9090)
	viper.SetDefault("monitoring.path", "/metrics")
	viper.SetDefault("monitoring.tracing.enabled", false)
	viper.SetDefault("monitoring.tracing.service_name", "auth-service")
	viper.SetDefault("monitoring.tracing.sample_rate", 0.1)

	// TLS defaults
	viper.SetDefault("tls.enabled", false)
	viper.SetDefault("tls.min_version", "1.2")

	// Feature flags defaults
	viper.SetDefault("features.registration_enabled", true)
	viper.SetDefault("features.password_reset_enabled", false)
	viper.SetDefault("features.multi_factor_auth_enabled", false)
	viper.SetDefault("features.session_analytics_enabled", true)
	viper.SetDefault("features.security_notifications_enabled", true)

	// Environment default
	viper.SetDefault("environment", "development")
}
