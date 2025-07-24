package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server            ServerConfig            `mapstructure:"server"`
	MarketData        MarketDataConfig        `mapstructure:"market_data"`
	Redis             RedisConfig             `mapstructure:"redis"`
	Database          DatabaseConfig          `mapstructure:"database"`
	WebSocket         WebSocketConfig         `mapstructure:"websocket"`
	TechnicalAnalysis TechnicalAnalysisConfig `mapstructure:"technical_analysis"`
	Portfolio         PortfolioConfig         `mapstructure:"portfolio"`
	Alerts            AlertsConfig            `mapstructure:"alerts"`
	Integrations      IntegrationsConfig      `mapstructure:"integrations"`
	Logging           LoggingConfig           `mapstructure:"logging"`
	Monitoring        MonitoringConfig        `mapstructure:"monitoring"`
	Security          SecurityConfig          `mapstructure:"security"`
	DeFi              DeFiConfig              `mapstructure:"defi"`
	AI                AIConfig                `mapstructure:"ai"`
	HFT               *HFTConfig              `mapstructure:"hft"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// MarketDataConfig holds market data provider configuration
type MarketDataConfig struct {
	Providers   ProvidersConfig   `mapstructure:"providers"`
	Exchanges   ExchangesConfig   `mapstructure:"exchanges"`
	Cache       CacheConfig       `mapstructure:"cache"`
	Aggregation AggregationConfig `mapstructure:"aggregation"`
}

// ProvidersConfig holds configuration for market data providers
type ProvidersConfig struct {
	CoinGecko CoinGeckoConfig `mapstructure:"coingecko"`
	Binance   BinanceConfig   `mapstructure:"binance"`
	Coinbase  CoinbaseConfig  `mapstructure:"coinbase"`
}

// ExchangesConfig holds configuration for exchange integrations
type ExchangesConfig struct {
	Binance  ExchangeBinanceConfig  `mapstructure:"binance"`
	Coinbase ExchangeCoinbaseConfig `mapstructure:"coinbase"`
	Kraken   ExchangeKrakenConfig   `mapstructure:"kraken"`
}

// CoinGeckoConfig holds CoinGecko API configuration
type CoinGeckoConfig struct {
	APIKey    string        `mapstructure:"api_key"`
	BaseURL   string        `mapstructure:"base_url"`
	RateLimit int           `mapstructure:"rate_limit"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

// BinanceConfig holds Binance API configuration
type BinanceConfig struct {
	WebSocketURL string        `mapstructure:"websocket_url"`
	RestURL      string        `mapstructure:"rest_url"`
	Timeout      time.Duration `mapstructure:"timeout"`
}

// CoinbaseConfig holds Coinbase API configuration
type CoinbaseConfig struct {
	WebSocketURL string        `mapstructure:"websocket_url"`
	RestURL      string        `mapstructure:"rest_url"`
	Timeout      time.Duration `mapstructure:"timeout"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	PriceTTL      time.Duration `mapstructure:"price_ttl"`
	IndicatorTTL  time.Duration `mapstructure:"indicator_ttl"`
	MarketDataTTL time.Duration `mapstructure:"market_data_ttl"`
}

// AggregationConfig holds market data aggregation configuration
type AggregationConfig struct {
	UpdateInterval       time.Duration `mapstructure:"update_interval"`
	ArbitrageThreshold   float64       `mapstructure:"arbitrage_threshold"`
	DataQualityThreshold float64       `mapstructure:"data_quality_threshold"`
	MaxPriceDeviation    float64       `mapstructure:"max_price_deviation"`
	CacheTTL             time.Duration `mapstructure:"cache_ttl"`
	EnableArbitrage      bool          `mapstructure:"enable_arbitrage"`
	EnableDataValidation bool          `mapstructure:"enable_data_validation"`
	Symbols              []string      `mapstructure:"symbols"`
}

// ExchangeBinanceConfig holds Binance exchange configuration
type ExchangeBinanceConfig struct {
	APIKey    string `mapstructure:"api_key"`
	SecretKey string `mapstructure:"secret_key"`
	BaseURL   string `mapstructure:"base_url"`
	WSURL     string `mapstructure:"ws_url"`
	Testnet   bool   `mapstructure:"testnet"`
	Enabled   bool   `mapstructure:"enabled"`
}

// ExchangeCoinbaseConfig holds Coinbase exchange configuration
type ExchangeCoinbaseConfig struct {
	APIKey     string `mapstructure:"api_key"`
	SecretKey  string `mapstructure:"secret_key"`
	Passphrase string `mapstructure:"passphrase"`
	BaseURL    string `mapstructure:"base_url"`
	WSURL      string `mapstructure:"ws_url"`
	Sandbox    bool   `mapstructure:"sandbox"`
	Enabled    bool   `mapstructure:"enabled"`
}

// ExchangeKrakenConfig holds Kraken exchange configuration
type ExchangeKrakenConfig struct {
	APIKey    string `mapstructure:"api_key"`
	SecretKey string `mapstructure:"secret_key"`
	BaseURL   string `mapstructure:"base_url"`
	WSURL     string `mapstructure:"ws_url"`
	Enabled   bool   `mapstructure:"enabled"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	DB           int    `mapstructure:"db"`
	Password     string `mapstructure:"password"`
	MaxRetries   int    `mapstructure:"max_retries"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Name            string        `mapstructure:"name"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	ReadBufferSize  int           `mapstructure:"read_buffer_size"`
	WriteBufferSize int           `mapstructure:"write_buffer_size"`
	CheckOrigin     bool          `mapstructure:"check_origin"`
	PingPeriod      time.Duration `mapstructure:"ping_period"`
	PongWait        time.Duration `mapstructure:"pong_wait"`
	WriteWait       time.Duration `mapstructure:"write_wait"`
}

// TechnicalAnalysisConfig holds technical analysis configuration
type TechnicalAnalysisConfig struct {
	Indicators []IndicatorConfig `mapstructure:"indicators"`
	Timeframes []string          `mapstructure:"timeframes"`
}

// IndicatorConfig holds individual indicator configuration
type IndicatorConfig struct {
	Name    string `mapstructure:"name"`
	Periods []int  `mapstructure:"periods"`
	Period  int    `mapstructure:"period"`
	Fast    int    `mapstructure:"fast"`
	Slow    int    `mapstructure:"slow"`
	Signal  int    `mapstructure:"signal"`
	StdDev  int    `mapstructure:"std_dev"`
}

// PortfolioConfig holds portfolio configuration
type PortfolioConfig struct {
	SyncInterval                   time.Duration `mapstructure:"sync_interval"`
	PerformanceCalculationInterval time.Duration `mapstructure:"performance_calculation_interval"`
	RiskMetricsInterval            time.Duration `mapstructure:"risk_metrics_interval"`
}

// AlertsConfig holds alerts configuration
type AlertsConfig struct {
	MaxAlertsPerUser     int           `mapstructure:"max_alerts_per_user"`
	CheckInterval        time.Duration `mapstructure:"check_interval"`
	NotificationChannels []string      `mapstructure:"notification_channels"`
}

// IntegrationsConfig holds integration configuration
type IntegrationsConfig struct {
	GoCoffee GoCoffeeConfig `mapstructure:"go_coffee"`
}

// GoCoffeeConfig holds Go Coffee integration configuration
type GoCoffeeConfig struct {
	DeFiServiceURL   string            `mapstructure:"defi_service_url"`
	WalletServiceURL string            `mapstructure:"wallet_service_url"`
	AIAgentsURL      string            `mapstructure:"ai_agents_url"`
	KafkaBrokers     []string          `mapstructure:"kafka_brokers"`
	KafkaTopics      KafkaTopicsConfig `mapstructure:"kafka_topics"`
}

// KafkaTopicsConfig holds Kafka topics configuration
type KafkaTopicsConfig struct {
	MarketData       string `mapstructure:"market_data"`
	TradingSignals   string `mapstructure:"trading_signals"`
	PortfolioUpdates string `mapstructure:"portfolio_updates"`
	Alerts           string `mapstructure:"alerts"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	MetricsEnabled      bool          `mapstructure:"metrics_enabled"`
	TracingEnabled      bool          `mapstructure:"tracing_enabled"`
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	CORS         CORSConfig         `mapstructure:"cors"`
	RateLimiting RateLimitingConfig `mapstructure:"rate_limiting"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

// RateLimitingConfig holds rate limiting configuration
type RateLimitingConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requests_per_minute"`
	Burst             int  `mapstructure:"burst"`
}

// DeFiConfig holds DeFi configuration
type DeFiConfig struct {
	Protocols ProtocolsConfig `mapstructure:"protocols"`
	Arbitrage ArbitrageConfig `mapstructure:"arbitrage"`
}

// ProtocolsConfig holds DeFi protocols configuration
type ProtocolsConfig struct {
	UniswapV3 UniswapV3Config `mapstructure:"uniswap_v3"`
	AaveV3    AaveV3Config    `mapstructure:"aave_v3"`
	Compound  CompoundConfig  `mapstructure:"compound"`
}

// UniswapV3Config holds Uniswap V3 configuration
type UniswapV3Config struct {
	FactoryAddress string `mapstructure:"factory_address"`
	QuoterAddress  string `mapstructure:"quoter_address"`
}

// AaveV3Config holds Aave V3 configuration
type AaveV3Config struct {
	PoolAddress string `mapstructure:"pool_address"`
}

// CompoundConfig holds Compound configuration
type CompoundConfig struct {
	ComptrollerAddress string `mapstructure:"comptroller_address"`
}

// ArbitrageConfig holds arbitrage configuration
type ArbitrageConfig struct {
	MinProfitThreshold float64 `mapstructure:"min_profit_threshold"`
	MaxGasPrice        int     `mapstructure:"max_gas_price"`
	SlippageTolerance  float64 `mapstructure:"slippage_tolerance"`
}

// AIConfig holds AI configuration
type AIConfig struct {
	TradingSignals    TradingSignalsConfig    `mapstructure:"trading_signals"`
	SentimentAnalysis SentimentAnalysisConfig `mapstructure:"sentiment_analysis"`
}

// TradingSignalsConfig holds trading signals configuration
type TradingSignalsConfig struct {
	Enabled             bool    `mapstructure:"enabled"`
	ConfidenceThreshold float64 `mapstructure:"confidence_threshold"`
	MaxSignalsPerHour   int     `mapstructure:"max_signals_per_hour"`
}

// SentimentAnalysisConfig holds sentiment analysis configuration
type SentimentAnalysisConfig struct {
	Enabled        bool          `mapstructure:"enabled"`
	Sources        []string      `mapstructure:"sources"`
	UpdateInterval time.Duration `mapstructure:"update_interval"`
}

// HFTConfig holds High-Frequency Trading configuration
type HFTConfig struct {
	Enabled         bool              `mapstructure:"enabled"`
	Feeds           HFTFeedsConfig    `mapstructure:"feeds"`
	OrderManagement HFTOMSConfig      `mapstructure:"order_management"`
	StrategyEngine  HFTStrategyConfig `mapstructure:"strategy_engine"`
	RiskManagement  HFTRiskConfig     `mapstructure:"risk_management"`
}

// HFTFeedsConfig holds HFT market data feeds configuration
type HFTFeedsConfig struct {
	Providers         []string      `mapstructure:"providers"`
	BufferSize        int           `mapstructure:"buffer_size"`
	LatencyThreshold  time.Duration `mapstructure:"latency_threshold"`
	ReconnectInterval time.Duration `mapstructure:"reconnect_interval"`
}

// HFTOMSConfig holds HFT Order Management System configuration
type HFTOMSConfig struct {
	MaxOrdersPerSecond int           `mapstructure:"max_orders_per_second"`
	OrderTimeout       time.Duration `mapstructure:"order_timeout"`
	FillTimeout        time.Duration `mapstructure:"fill_timeout"`
	RetryAttempts      int           `mapstructure:"retry_attempts"`
}

// HFTStrategyConfig holds HFT strategy engine configuration
type HFTStrategyConfig struct {
	MaxStrategies     int           `mapstructure:"max_strategies"`
	SignalBufferSize  int           `mapstructure:"signal_buffer_size"`
	ExecutionTimeout  time.Duration `mapstructure:"execution_timeout"`
	PerformanceWindow time.Duration `mapstructure:"performance_window"`
}

// HFTRiskConfig holds HFT risk management configuration
type HFTRiskConfig struct {
	MaxDailyLoss       float64       `mapstructure:"max_daily_loss"`
	MaxDrawdown        float64       `mapstructure:"max_drawdown"`
	MaxPositionSize    float64       `mapstructure:"max_position_size"`
	MaxExposure        float64       `mapstructure:"max_exposure"`
	CheckInterval      time.Duration `mapstructure:"check_interval"`
	ViolationThreshold int           `mapstructure:"violation_threshold"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	// Set default values
	setDefaults()

	// Configure viper
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, use defaults and environment variables
		logrus.Warn("Config file not found, using defaults and environment variables")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Apply environment-specific overrides
	applyEnvironmentOverrides(&config)

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", 8090)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 2)
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.name", "crypto_terminal")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "5m")

	// Market data defaults
	viper.SetDefault("market_data.providers.coingecko.base_url", "https://api.coingecko.com/api/v3")
	viper.SetDefault("market_data.providers.coingecko.rate_limit", 50)
	viper.SetDefault("market_data.providers.coingecko.timeout", "10s")
	viper.SetDefault("market_data.providers.binance.rest_url", "https://api.binance.com")
	viper.SetDefault("market_data.providers.binance.websocket_url", "wss://stream.binance.com:9443/ws")
	viper.SetDefault("market_data.providers.binance.timeout", "10s")

	// Cache defaults
	viper.SetDefault("market_data.cache.price_ttl", "30s")
	viper.SetDefault("market_data.cache.indicator_ttl", "5m")
	viper.SetDefault("market_data.cache.market_data_ttl", "1m")

	// WebSocket defaults
	viper.SetDefault("websocket.read_buffer_size", 1024)
	viper.SetDefault("websocket.write_buffer_size", 1024)
	viper.SetDefault("websocket.check_origin", false)
	viper.SetDefault("websocket.ping_period", "54s")
	viper.SetDefault("websocket.pong_wait", "60s")
	viper.SetDefault("websocket.write_wait", "10s")

	// Portfolio defaults
	viper.SetDefault("portfolio.sync_interval", "5m")
	viper.SetDefault("portfolio.performance_calculation_interval", "1h")
	viper.SetDefault("portfolio.risk_metrics_interval", "30m")

	// Alerts defaults
	viper.SetDefault("alerts.max_alerts_per_user", 100)
	viper.SetDefault("alerts.check_interval", "30s")
	viper.SetDefault("alerts.notification_channels", []string{"EMAIL", "PUSH"})

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	// Monitoring defaults
	viper.SetDefault("monitoring.metrics_enabled", true)
	viper.SetDefault("monitoring.tracing_enabled", true)
	viper.SetDefault("monitoring.health_check_interval", "30s")

	// Security defaults
	viper.SetDefault("security.cors.allowed_origins", []string{"http://localhost:3000", "http://localhost:8090"})
	viper.SetDefault("security.cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("security.cors.allowed_headers", []string{"Content-Type", "Authorization", "X-Requested-With"})
	viper.SetDefault("security.rate_limiting.enabled", true)
	viper.SetDefault("security.rate_limiting.requests_per_minute", 1000)
	viper.SetDefault("security.rate_limiting.burst", 100)

	// HFT defaults
	viper.SetDefault("hft.enabled", false)
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	// Validate server configuration
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// Validate database configuration
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if config.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	// Validate Redis configuration
	if config.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	if config.Redis.Port <= 0 || config.Redis.Port > 65535 {
		return fmt.Errorf("invalid redis port: %d", config.Redis.Port)
	}

	// Validate market data configuration
	if config.MarketData.Providers.CoinGecko.BaseURL == "" {
		return fmt.Errorf("coingecko base URL is required")
	}
	if config.MarketData.Providers.Binance.RestURL == "" {
		return fmt.Errorf("binance rest URL is required")
	}

	return nil
}

// applyEnvironmentOverrides applies environment-specific configuration overrides
func applyEnvironmentOverrides(config *Config) {
	// Check for environment-specific settings
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = os.Getenv("ENV")
	}

	switch strings.ToLower(env) {
	case "development", "dev":
		applyDevelopmentOverrides(config)
	case "production", "prod":
		applyProductionOverrides(config)
	case "testing", "test":
		applyTestingOverrides(config)
	}
}

// applyDevelopmentOverrides applies development environment overrides
func applyDevelopmentOverrides(config *Config) {
	// Enable debug logging
	config.Logging.Level = "debug"
	config.Logging.Format = "text"

	// Disable CORS origin checking for development
	config.WebSocket.CheckOrigin = false

	// Enable all monitoring features
	config.Monitoring.MetricsEnabled = true
	config.Monitoring.TracingEnabled = true

	// Relaxed rate limiting for development
	config.Security.RateLimiting.RequestsPerMinute = 10000
	config.Security.RateLimiting.Burst = 1000

	logrus.Info("Applied development environment overrides")
}

// applyProductionOverrides applies production environment overrides
func applyProductionOverrides(config *Config) {
	// Production logging
	config.Logging.Level = "info"
	config.Logging.Format = "json"

	// Strict CORS checking
	config.WebSocket.CheckOrigin = true

	// Enable all monitoring features
	config.Monitoring.MetricsEnabled = true
	config.Monitoring.TracingEnabled = true

	// Strict rate limiting for production
	config.Security.RateLimiting.Enabled = true

	// Disable HFT by default in production unless explicitly enabled
	if !viper.IsSet("hft.enabled") {
		config.HFT.Enabled = false
	}

	logrus.Info("Applied production environment overrides")
}

// applyTestingOverrides applies testing environment overrides
func applyTestingOverrides(config *Config) {
	// Test logging
	config.Logging.Level = "warn"
	config.Logging.Format = "text"

	// Disable external services for testing
	config.Monitoring.MetricsEnabled = false
	config.Monitoring.TracingEnabled = false

	// Disable rate limiting for tests
	config.Security.RateLimiting.Enabled = false

	// Use test database
	if !viper.IsSet("database.name") {
		config.Database.Name = "crypto_terminal_test"
	}

	// Use test Redis DB
	if !viper.IsSet("redis.db") {
		config.Redis.DB = 15 // Use last Redis DB for tests
	}

	logrus.Info("Applied testing environment overrides")
}

// GetDSN returns the database connection string
func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

// GetRedisAddr returns the Redis address
func (r *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// GetServerAddr returns the server address
func (s *ServerConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
