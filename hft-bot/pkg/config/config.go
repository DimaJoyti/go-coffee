package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
)

// Config represents the complete application configuration
type Config struct {
	Service    ServiceConfig    `yaml:"service" json:"service"`
	Logger     LoggerConfig     `yaml:"logger" json:"logger"`
	Database   DatabaseConfig   `yaml:"database" json:"database"`
	Redis      RedisConfig      `yaml:"redis" json:"redis"`
	Exchanges  ExchangesConfig  `yaml:"exchanges" json:"exchanges"`
	HFT        HFTConfig        `yaml:"hft" json:"hft"`
	Risk       RiskConfig       `yaml:"risk" json:"risk"`
	Strategies StrategiesConfig `yaml:"strategies" json:"strategies"`
	Monitoring MonitoringConfig `yaml:"monitoring" json:"monitoring"`
}

// ServiceConfig holds service-level configuration
type ServiceConfig struct {
	Name        string `yaml:"name" json:"name"`
	Version     string `yaml:"version" json:"version"`
	Environment string `yaml:"environment" json:"environment"`
	Port        int    `yaml:"port" json:"port"`
	Host        string `yaml:"host" json:"host"`
	PaperMode   bool   `yaml:"paper_mode" json:"paper_mode"`
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level  string `yaml:"level" json:"level"`
	Format string `yaml:"format" json:"format"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `yaml:"host" json:"host"`
	Port            int           `yaml:"port" json:"port"`
	Database        string        `yaml:"database" json:"database"`
	Username        string        `yaml:"username" json:"username"`
	Password        string        `yaml:"password" json:"password"`
	SSLMode         string        `yaml:"ssl_mode" json:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string        `yaml:"host" json:"host"`
	Port         int           `yaml:"port" json:"port"`
	Password     string        `yaml:"password" json:"password"`
	Database     int           `yaml:"database" json:"database"`
	PoolSize     int           `yaml:"pool_size" json:"pool_size"`
	MinIdleConns int           `yaml:"min_idle_conns" json:"min_idle_conns"`
	DialTimeout  time.Duration `yaml:"dial_timeout" json:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`
}

// ExchangesConfig holds exchange configurations
type ExchangesConfig struct {
	Binance  ExchangeConfig `yaml:"binance" json:"binance"`
	Coinbase ExchangeConfig `yaml:"coinbase" json:"coinbase"`
	Kraken   ExchangeConfig `yaml:"kraken" json:"kraken"`
}

// ExchangeConfig holds individual exchange configuration
type ExchangeConfig struct {
	Enabled   bool   `yaml:"enabled" json:"enabled"`
	APIKey    string `yaml:"api_key" json:"api_key"`
	APISecret string `yaml:"api_secret" json:"api_secret"`
	Sandbox   bool   `yaml:"sandbox" json:"sandbox"`
	BaseURL   string `yaml:"base_url" json:"base_url"`
	WSBaseURL string `yaml:"ws_base_url" json:"ws_base_url"`
}

// HFTConfig holds HFT-specific configuration
type HFTConfig struct {
	Enabled                 bool          `yaml:"enabled" json:"enabled"`
	MaxOrdersPerSecond      int           `yaml:"max_orders_per_second" json:"max_orders_per_second"`
	LatencyThreshold        time.Duration `yaml:"latency_threshold" json:"latency_threshold"`
	OrderTimeout            time.Duration `yaml:"order_timeout" json:"order_timeout"`
	FillTimeout             time.Duration `yaml:"fill_timeout" json:"fill_timeout"`
	RetryAttempts           int           `yaml:"retry_attempts" json:"retry_attempts"`
	BufferSize              int           `yaml:"buffer_size" json:"buffer_size"`
	ReconnectInterval       time.Duration `yaml:"reconnect_interval" json:"reconnect_interval"`
	PerformanceWindow       time.Duration `yaml:"performance_window" json:"performance_window"`
	MaxConcurrentStrategies int           `yaml:"max_concurrent_strategies" json:"max_concurrent_strategies"`
}

// RiskConfig holds risk management configuration
type RiskConfig struct {
	Enabled               bool            `yaml:"enabled" json:"enabled"`
	MaxDailyLoss          decimal.Decimal `yaml:"max_daily_loss" json:"max_daily_loss"`
	MaxDrawdown           decimal.Decimal `yaml:"max_drawdown" json:"max_drawdown"`
	MaxPositionSize       decimal.Decimal `yaml:"max_position_size" json:"max_position_size"`
	MaxExposure           decimal.Decimal `yaml:"max_exposure" json:"max_exposure"`
	CheckInterval         time.Duration   `yaml:"check_interval" json:"check_interval"`
	ViolationThreshold    int             `yaml:"violation_threshold" json:"violation_threshold"`
	EmergencyStopEnabled  bool            `yaml:"emergency_stop_enabled" json:"emergency_stop_enabled"`
	CircuitBreakerEnabled bool            `yaml:"circuit_breaker_enabled" json:"circuit_breaker_enabled"`
}

// StrategiesConfig holds strategy configurations
type StrategiesConfig struct {
	MarketMaking MarketMakingConfig `yaml:"market_making" json:"market_making"`
	Arbitrage    ArbitrageConfig    `yaml:"arbitrage" json:"arbitrage"`
	Momentum     MomentumConfig     `yaml:"momentum" json:"momentum"`
	MeanRevert   MeanRevertConfig   `yaml:"mean_revert" json:"mean_revert"`
}

// MarketMakingConfig holds market making strategy configuration
type MarketMakingConfig struct {
	Enabled         bool            `yaml:"enabled" json:"enabled"`
	SpreadPercent   decimal.Decimal `yaml:"spread_percent" json:"spread_percent"`
	OrderSize       decimal.Decimal `yaml:"order_size" json:"order_size"`
	MaxInventory    decimal.Decimal `yaml:"max_inventory" json:"max_inventory"`
	RefreshInterval time.Duration   `yaml:"refresh_interval" json:"refresh_interval"`
}

// ArbitrageConfig holds arbitrage strategy configuration
type ArbitrageConfig struct {
	Enabled          bool            `yaml:"enabled" json:"enabled"`
	MinProfitPercent decimal.Decimal `yaml:"min_profit_percent" json:"min_profit_percent"`
	MaxOrderSize     decimal.Decimal `yaml:"max_order_size" json:"max_order_size"`
	CheckInterval    time.Duration   `yaml:"check_interval" json:"check_interval"`
	ExecutionTimeout time.Duration   `yaml:"execution_timeout" json:"execution_timeout"`
}

// MomentumConfig holds momentum strategy configuration
type MomentumConfig struct {
	Enabled          bool            `yaml:"enabled" json:"enabled"`
	LookbackPeriod   time.Duration   `yaml:"lookback_period" json:"lookback_period"`
	ThresholdPercent decimal.Decimal `yaml:"threshold_percent" json:"threshold_percent"`
	OrderSize        decimal.Decimal `yaml:"order_size" json:"order_size"`
	HoldPeriod       time.Duration   `yaml:"hold_period" json:"hold_period"`
}

// MeanRevertConfig holds mean reversion strategy configuration
type MeanRevertConfig struct {
	Enabled          bool            `yaml:"enabled" json:"enabled"`
	LookbackPeriod   time.Duration   `yaml:"lookback_period" json:"lookback_period"`
	DeviationPercent decimal.Decimal `yaml:"deviation_percent" json:"deviation_percent"`
	OrderSize        decimal.Decimal `yaml:"order_size" json:"order_size"`
	HoldPeriod       time.Duration   `yaml:"hold_period" json:"hold_period"`
}

// MonitoringConfig holds monitoring and observability configuration
type MonitoringConfig struct {
	Enabled         bool   `yaml:"enabled" json:"enabled"`
	MetricsPort     int    `yaml:"metrics_port" json:"metrics_port"`
	JaegerEndpoint  string `yaml:"jaeger_endpoint" json:"jaeger_endpoint"`
	PrometheusURL   string `yaml:"prometheus_url" json:"prometheus_url"`
	HealthCheckPort int    `yaml:"health_check_port" json:"health_check_port"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Set config file path
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath("../configs")
		v.AddConfigPath("../../configs")
	}

	// Enable environment variable support
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal config
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Service defaults
	v.SetDefault("service.name", "hft-bot")
	v.SetDefault("service.version", "1.0.0")
	v.SetDefault("service.environment", "development")
	v.SetDefault("service.port", 8080)
	v.SetDefault("service.host", "0.0.0.0")
	v.SetDefault("service.paper_mode", true)

	// Logger defaults
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.database", "hft_bot")
	v.SetDefault("database.username", "postgres")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", "5m")

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.database", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 2)
	v.SetDefault("redis.dial_timeout", "5s")
	v.SetDefault("redis.read_timeout", "3s")
	v.SetDefault("redis.write_timeout", "3s")

	// HFT defaults
	v.SetDefault("hft.enabled", true)
	v.SetDefault("hft.max_orders_per_second", 100)
	v.SetDefault("hft.latency_threshold", "10ms")
	v.SetDefault("hft.order_timeout", "5s")
	v.SetDefault("hft.fill_timeout", "10s")
	v.SetDefault("hft.retry_attempts", 3)
	v.SetDefault("hft.buffer_size", 10000)
	v.SetDefault("hft.reconnect_interval", "5s")
	v.SetDefault("hft.performance_window", "24h")
	v.SetDefault("hft.max_concurrent_strategies", 10)

	// Risk defaults
	v.SetDefault("risk.enabled", true)
	v.SetDefault("risk.max_daily_loss", "10000.0")
	v.SetDefault("risk.max_drawdown", "5.0")
	v.SetDefault("risk.max_position_size", "10.0")
	v.SetDefault("risk.max_exposure", "50000.0")
	v.SetDefault("risk.check_interval", "1s")
	v.SetDefault("risk.violation_threshold", 5)
	v.SetDefault("risk.emergency_stop_enabled", true)
	v.SetDefault("risk.circuit_breaker_enabled", true)

	// Monitoring defaults
	v.SetDefault("monitoring.enabled", true)
	v.SetDefault("monitoring.metrics_port", 9090)
	v.SetDefault("monitoring.health_check_port", 8081)
}

// validate validates the configuration
func validate(config *Config) error {
	if config.Service.Name == "" {
		return fmt.Errorf("service name is required")
	}

	if config.Service.Port <= 0 || config.Service.Port > 65535 {
		return fmt.Errorf("invalid service port: %d", config.Service.Port)
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}

	return nil
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode)
}

// GetRedisAddr returns the Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
