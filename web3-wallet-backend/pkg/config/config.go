package config

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	Redis       RedisConfig       `yaml:"redis"`
	Blockchain  BlockchainConfig  `yaml:"blockchain"`
	Security    SecurityConfig    `yaml:"security"`
	Logging     LoggingConfig     `yaml:"logging"`
	Monitoring  MonitoringConfig  `yaml:"monitoring"`
	Notification NotificationConfig `yaml:"notification"`
}

// ServerConfig represents the HTTP server configuration
type ServerConfig struct {
	Port           int           `yaml:"port"`
	Host           string        `yaml:"host"`
	Timeout        time.Duration `yaml:"timeout"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	MaxHeaderBytes int           `yaml:"max_header_bytes"`
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	Driver          string        `yaml:"driver"`
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	SSLMode         string        `yaml:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// RedisConfig represents the Redis configuration
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// BlockchainConfig represents the blockchain configuration
type BlockchainConfig struct {
	Ethereum BlockchainNetworkConfig `yaml:"ethereum"`
	BSC      BlockchainNetworkConfig `yaml:"bsc"`
	Polygon  BlockchainNetworkConfig `yaml:"polygon"`
}

// BlockchainNetworkConfig represents the configuration for a blockchain network
type BlockchainNetworkConfig struct {
	Network           string `yaml:"network"`
	RPCURL            string `yaml:"rpc_url"`
	WSURL             string `yaml:"ws_url"`
	ChainID           int    `yaml:"chain_id"`
	GasLimit          uint64 `yaml:"gas_limit"`
	GasPrice          string `yaml:"gas_price"`
	ConfirmationBlocks int    `yaml:"confirmation_blocks"`
}

// SecurityConfig represents the security configuration
type SecurityConfig struct {
	JWT        JWTConfig        `yaml:"jwt"`
	Encryption EncryptionConfig `yaml:"encryption"`
	RateLimit  RateLimitConfig  `yaml:"rate_limit"`
}

// JWTConfig represents the JWT configuration
type JWTConfig struct {
	Secret            string        `yaml:"secret"`
	Expiration        time.Duration `yaml:"expiration"`
	RefreshExpiration time.Duration `yaml:"refresh_expiration"`
}

// EncryptionConfig represents the encryption configuration
type EncryptionConfig struct {
	KeyDerivation string `yaml:"key_derivation"`
	Iterations    int    `yaml:"iterations"`
	SaltLength    int    `yaml:"salt_length"`
	KeyLength     int    `yaml:"key_length"`
}

// RateLimitConfig represents the rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerMinute int  `yaml:"requests_per_minute"`
	Burst             int  `yaml:"burst"`
}

// LoggingConfig represents the logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	FilePath   string `yaml:"file_path"`
	MaxSize    int    `yaml:"max_size"`
	MaxAge     int    `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
	Compress   bool   `yaml:"compress"`
}

// MonitoringConfig represents the monitoring configuration
type MonitoringConfig struct {
	Prometheus  PrometheusConfig  `yaml:"prometheus"`
	HealthCheck HealthCheckConfig `yaml:"health_check"`
	Metrics     MetricsConfig     `yaml:"metrics"`
}

// PrometheusConfig represents the Prometheus configuration
type PrometheusConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

// HealthCheckConfig represents the health check configuration
type HealthCheckConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
}

// MetricsConfig represents the metrics configuration
type MetricsConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
}

// NotificationConfig represents the notification configuration
type NotificationConfig struct {
	Email EmailConfig `yaml:"email"`
	SMS   SMSConfig   `yaml:"sms"`
	Push  PushConfig  `yaml:"push"`
}

// EmailConfig represents the email notification configuration
type EmailConfig struct {
	Enabled      bool   `yaml:"enabled"`
	SMTPHost     string `yaml:"smtp_host"`
	SMTPPort     int    `yaml:"smtp_port"`
	SMTPUsername string `yaml:"smtp_username"`
	SMTPPassword string `yaml:"smtp_password"`
	FromEmail    string `yaml:"from_email"`
	FromName     string `yaml:"from_name"`
}

// SMSConfig represents the SMS notification configuration
type SMSConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Provider   string `yaml:"provider"`
	AccountSID string `yaml:"account_sid"`
	AuthToken  string `yaml:"auth_token"`
	FromNumber string `yaml:"from_number"`
}

// PushConfig represents the push notification configuration
type PushConfig struct {
	Enabled         bool   `yaml:"enabled"`
	Provider        string `yaml:"provider"`
	CredentialsFile string `yaml:"credentials_file"`
}

// LoadConfig loads the configuration from a file
func LoadConfig(configPath string) (*Config, error) {
	// Read the configuration file
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the YAML configuration
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}
