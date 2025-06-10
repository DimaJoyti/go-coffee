package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// EnhancedConfig provides modern configuration management with validation
type EnhancedConfig struct {
	viper  *viper.Viper
	logger *zap.Logger
}

// ConfigOptions provides configuration options
type ConfigOptions struct {
	ConfigName    string
	ConfigPaths   []string
	ConfigType    string
	EnvPrefix     string
	AutomaticEnv  bool
	AllowEmptyEnv bool
	Logger        *zap.Logger
}

// DefaultConfigOptions returns sensible defaults
func DefaultConfigOptions() *ConfigOptions {
	return &ConfigOptions{
		ConfigName:    "config",
		ConfigPaths:   []string{".", "./config", "./configs"},
		ConfigType:    "yaml",
		EnvPrefix:     "GO_COFFEE",
		AutomaticEnv:  true,
		AllowEmptyEnv: false,
	}
}

// NewEnhancedConfig creates a new enhanced configuration manager
func NewEnhancedConfig(opts *ConfigOptions) (*EnhancedConfig, error) {
	if opts == nil {
		opts = DefaultConfigOptions()
	}

	v := viper.New()

	// Set configuration file settings
	v.SetConfigName(opts.ConfigName)
	v.SetConfigType(opts.ConfigType)

	// Add configuration paths
	for _, path := range opts.ConfigPaths {
		v.AddConfigPath(path)
	}

	// Environment variable settings
	if opts.EnvPrefix != "" {
		v.SetEnvPrefix(opts.EnvPrefix)
	}

	if opts.AutomaticEnv {
		v.AutomaticEnv()
	}

	// Replace dots and dashes in env vars
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Set defaults
	setDefaults(v)

	return &EnhancedConfig{
		viper:  v,
		logger: opts.Logger,
	}, nil
}

// setDefaults sets sensible default values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.api_gateway_port", 8080)
	v.SetDefault("server.producer_port", 3000)
	v.SetDefault("server.consumer_port", 3001)
	v.SetDefault("server.streams_port", 3002)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.idle_timeout", "120s")

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.name", "go_coffee")
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", "300s")

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 5)
	v.SetDefault("redis.dial_timeout", "5s")
	v.SetDefault("redis.read_timeout", "3s")
	v.SetDefault("redis.write_timeout", "3s")

	// Kafka defaults
	v.SetDefault("kafka.brokers", []string{"localhost:9092"})
	v.SetDefault("kafka.topic", "coffee_orders")
	v.SetDefault("kafka.processed_topic", "processed_orders")
	v.SetDefault("kafka.consumer_group", "coffee-consumer-group")
	v.SetDefault("kafka.worker_pool_size", 3)
	v.SetDefault("kafka.retry_max", 5)
	v.SetDefault("kafka.required_acks", "all")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "console")
	v.SetDefault("logging.colorized", true)
	v.SetDefault("logging.json_format", false)

	// Security defaults
	v.SetDefault("security.jwt_expiry", "24h")
	v.SetDefault("security.refresh_token_expiry", "720h")
	v.SetDefault("security.jwt_issuer", "go-coffee")
	v.SetDefault("security.jwt_audience", "go-coffee-users")

	// Feature flags defaults
	v.SetDefault("features.producer_service_enabled", true)
	v.SetDefault("features.consumer_service_enabled", true)
	v.SetDefault("features.streams_service_enabled", true)
	v.SetDefault("features.api_gateway_enabled", true)
	v.SetDefault("features.web3_wallet_enabled", true)
	v.SetDefault("features.defi_service_enabled", true)
	v.SetDefault("features.ai_search_enabled", true)
	v.SetDefault("features.ai_agents_enabled", true)

	// Monitoring defaults
	v.SetDefault("monitoring.prometheus.enabled", true)
	v.SetDefault("monitoring.prometheus.port", 9090)
	v.SetDefault("monitoring.prometheus.metrics_path", "/metrics")
	v.SetDefault("monitoring.jaeger.enabled", true)
	v.SetDefault("monitoring.jaeger.sampler_type", "const")
	v.SetDefault("monitoring.jaeger.sampler_param", "1")
}

// Load loads configuration from file and environment
func (c *EnhancedConfig) Load() error {
	// Try to read config file
	if err := c.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; using defaults and env vars
			if c.logger != nil {
				c.logger.Info("Config file not found, using defaults and environment variables")
			}
		} else {
			return fmt.Errorf("error reading config file: %w", err)
		}
	} else {
		if c.logger != nil {
			c.logger.Info("Using config file", zap.String("file", c.viper.ConfigFileUsed()))
		}
	}

	return nil
}

// LoadFromFile loads configuration from a specific file
func (c *EnhancedConfig) LoadFromFile(filename string) error {
	c.viper.SetConfigFile(filename)
	return c.viper.ReadInConfig()
}

// Get returns a value by key
func (c *EnhancedConfig) Get(key string) interface{} {
	return c.viper.Get(key)
}

// GetString returns a string value
func (c *EnhancedConfig) GetString(key string) string {
	return c.viper.GetString(key)
}

// GetInt returns an int value
func (c *EnhancedConfig) GetInt(key string) int {
	return c.viper.GetInt(key)
}

// GetBool returns a bool value
func (c *EnhancedConfig) GetBool(key string) bool {
	return c.viper.GetBool(key)
}

// GetDuration returns a duration value
func (c *EnhancedConfig) GetDuration(key string) time.Duration {
	return c.viper.GetDuration(key)
}

// GetStringSlice returns a string slice value
func (c *EnhancedConfig) GetStringSlice(key string) []string {
	return c.viper.GetStringSlice(key)
}

// Set sets a value
func (c *EnhancedConfig) Set(key string, value interface{}) {
	c.viper.Set(key, value)
}

// IsSet checks if a key is set
func (c *EnhancedConfig) IsSet(key string) bool {
	return c.viper.IsSet(key)
}

// Unmarshal unmarshals config into a struct
func (c *EnhancedConfig) Unmarshal(rawVal interface{}) error {
	return c.viper.Unmarshal(rawVal)
}

// UnmarshalKey unmarshals a specific key into a struct
func (c *EnhancedConfig) UnmarshalKey(key string, rawVal interface{}) error {
	return c.viper.UnmarshalKey(key, rawVal)
}

// WriteConfig writes the current configuration to file
func (c *EnhancedConfig) WriteConfig() error {
	return c.viper.WriteConfig()
}

// WriteConfigAs writes the current configuration to a specific file
func (c *EnhancedConfig) WriteConfigAs(filename string) error {
	return c.viper.WriteConfigAs(filename)
}

// GetConfigFile returns the config file being used
func (c *EnhancedConfig) GetConfigFile() string {
	return c.viper.ConfigFileUsed()
}

// WatchConfig watches for config file changes
func (c *EnhancedConfig) WatchConfig() {
	c.viper.WatchConfig()
}

// OnConfigChange sets a callback for config changes
func (c *EnhancedConfig) OnConfigChange(run func()) {
	c.viper.OnConfigChange(func(e fsnotify.Event) {
		if c.logger != nil {
			c.logger.Info("Config file changed", zap.String("file", e.Name))
		}
		run()
	})
}

// Validate validates the configuration
func (c *EnhancedConfig) Validate() error {
	var errors []string

	// Validate required fields
	requiredFields := []string{
		"database.host",
		"database.name",
		"redis.host",
		"kafka.brokers",
	}

	for _, field := range requiredFields {
		if !c.viper.IsSet(field) || c.viper.GetString(field) == "" {
			errors = append(errors, fmt.Sprintf("%s is required", field))
		}
	}

	// Validate security settings
	if jwtSecret := c.viper.GetString("security.jwt_secret"); jwtSecret == "" || jwtSecret == "your-super-secret-jwt-key" {
		errors = append(errors, "security.jwt_secret must be set to a secure value")
	}

	// Validate ports
	ports := map[string]string{
		"server.api_gateway_port": "API Gateway port",
		"server.producer_port":    "Producer port",
		"server.consumer_port":    "Consumer port",
		"database.port":           "Database port",
		"redis.port":              "Redis port",
	}

	for key, name := range ports {
		if port := c.viper.GetInt(key); port <= 0 || port > 65535 {
			errors = append(errors, fmt.Sprintf("%s must be between 1 and 65535", name))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// PrintConfig prints the configuration (without secrets)
func (c *EnhancedConfig) PrintConfig() {
	fmt.Println("=== Go Coffee Enhanced Configuration ===")

	// Server configuration
	fmt.Println("\n🖥️  Server Configuration:")
	fmt.Printf("  Host: %s\n", c.GetString("server.host"))
	fmt.Printf("  API Gateway Port: %d\n", c.GetInt("server.api_gateway_port"))
	fmt.Printf("  Producer Port: %d\n", c.GetInt("server.producer_port"))
	fmt.Printf("  Consumer Port: %d\n", c.GetInt("server.consumer_port"))

	// Database configuration
	fmt.Println("\n🗄️  Database Configuration:")
	fmt.Printf("  Host: %s\n", c.GetString("database.host"))
	fmt.Printf("  Port: %d\n", c.GetInt("database.port"))
	fmt.Printf("  Name: %s\n", c.GetString("database.name"))
	fmt.Printf("  Max Open Connections: %d\n", c.GetInt("database.max_open_conns"))

	// Redis configuration
	fmt.Println("\n🔴 Redis Configuration:")
	fmt.Printf("  Host: %s\n", c.GetString("redis.host"))
	fmt.Printf("  Port: %d\n", c.GetInt("redis.port"))
	fmt.Printf("  Database: %d\n", c.GetInt("redis.db"))
	fmt.Printf("  Pool Size: %d\n", c.GetInt("redis.pool_size"))

	// Kafka configuration
	fmt.Println("\n📨 Kafka Configuration:")
	fmt.Printf("  Brokers: %v\n", c.GetStringSlice("kafka.brokers"))
	fmt.Printf("  Topic: %s\n", c.GetString("kafka.topic"))
	fmt.Printf("  Consumer Group: %s\n", c.GetString("kafka.consumer_group"))

	// Feature flags
	fmt.Println("\n🚩 Feature Flags:")
	fmt.Printf("  Producer Service: %t\n", c.GetBool("features.producer_service_enabled"))
	fmt.Printf("  Consumer Service: %t\n", c.GetBool("features.consumer_service_enabled"))
	fmt.Printf("  Web3 Wallet: %t\n", c.GetBool("features.web3_wallet_enabled"))
	fmt.Printf("  AI Search: %t\n", c.GetBool("features.ai_search_enabled"))

	fmt.Println("\n=== End Configuration ===")
}
