package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server           ServerConfig           `mapstructure:"server"`
	MarketData       MarketDataConfig       `mapstructure:"market_data"`
	Redis            RedisConfig            `mapstructure:"redis"`
	Database         DatabaseConfig         `mapstructure:"database"`
	WebSocket        WebSocketConfig        `mapstructure:"websocket"`
	TechnicalAnalysis TechnicalAnalysisConfig `mapstructure:"technical_analysis"`
	Portfolio        PortfolioConfig        `mapstructure:"portfolio"`
	Alerts           AlertsConfig           `mapstructure:"alerts"`
	Integrations     IntegrationsConfig     `mapstructure:"integrations"`
	Logging          LoggingConfig          `mapstructure:"logging"`
	Monitoring       MonitoringConfig       `mapstructure:"monitoring"`
	Security         SecurityConfig         `mapstructure:"security"`
	DeFi             DeFiConfig             `mapstructure:"defi"`
	AI               AIConfig               `mapstructure:"ai"`
	HFT              *HFTConfig             `mapstructure:"hft"`
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
	Providers ProvidersConfig `mapstructure:"providers"`
	Exchanges ExchangesConfig `mapstructure:"exchanges"`
	Cache     CacheConfig     `mapstructure:"cache"`
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
	SyncInterval                    time.Duration `mapstructure:"sync_interval"`
	PerformanceCalculationInterval  time.Duration `mapstructure:"performance_calculation_interval"`
	RiskMetricsInterval            time.Duration `mapstructure:"risk_metrics_interval"`
}

// AlertsConfig holds alerts configuration
type AlertsConfig struct {
	MaxAlertsPerUser       int           `mapstructure:"max_alerts_per_user"`
	CheckInterval          time.Duration `mapstructure:"check_interval"`
	NotificationChannels   []string      `mapstructure:"notification_channels"`
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
	MetricsEnabled        bool          `mapstructure:"metrics_enabled"`
	TracingEnabled        bool          `mapstructure:"tracing_enabled"`
	HealthCheckInterval   time.Duration `mapstructure:"health_check_interval"`
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
	Enabled            bool `mapstructure:"enabled"`
	RequestsPerMinute  int  `mapstructure:"requests_per_minute"`
	Burst              int  `mapstructure:"burst"`
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
	MinProfitThreshold  float64 `mapstructure:"min_profit_threshold"`
	MaxGasPrice         int     `mapstructure:"max_gas_price"`
	SlippageTolerance   float64 `mapstructure:"slippage_tolerance"`
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
	Enabled           bool                  `mapstructure:"enabled"`
	Feeds             HFTFeedsConfig        `mapstructure:"feeds"`
	OrderManagement   HFTOMSConfig          `mapstructure:"order_management"`
	StrategyEngine    HFTStrategyConfig     `mapstructure:"strategy_engine"`
	RiskManagement    HFTRiskConfig         `mapstructure:"risk_management"`
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
	viper.SetConfigFile(configPath)
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
