package config

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration
type Config struct {
	Server       ServerConfig       `yaml:"server"`
	Database     DatabaseConfig     `yaml:"database"`
	Redis        RedisConfig        `yaml:"redis"`
	Blockchain   BlockchainConfig   `yaml:"blockchain"`
	Security     SecurityConfig     `yaml:"security"`
	Logging      LoggingConfig      `yaml:"logging"`
	Monitoring   MonitoringConfig   `yaml:"monitoring"`
	Notification NotificationConfig `yaml:"notification"`
	DeFi         DeFiConfig         `yaml:"defi"`
	Services     ServicesConfig     `yaml:"services"`
	Telegram     TelegramConfig     `yaml:"telegram"`
	AI           AIConfig           `yaml:"ai"`
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
	Solana   SolanaNetworkConfig     `yaml:"solana"`
}

// BlockchainNetworkConfig represents the configuration for a blockchain network
type BlockchainNetworkConfig struct {
	Network            string `yaml:"network"`
	RPCURL             string `yaml:"rpc_url"`
	WSURL              string `yaml:"ws_url"`
	ChainID            int    `yaml:"chain_id"`
	GasLimit           uint64 `yaml:"gas_limit"`
	GasPrice           string `yaml:"gas_price"`
	ConfirmationBlocks int    `yaml:"confirmation_blocks"`
}

// SolanaNetworkConfig represents the configuration for Solana network
type SolanaNetworkConfig struct {
	Network            string `yaml:"network"`
	RPCURL             string `yaml:"rpc_url"`
	WSURL              string `yaml:"ws_url"`
	Cluster            string `yaml:"cluster"`
	Commitment         string `yaml:"commitment"`
	Timeout            string `yaml:"timeout"`
	MaxRetries         int    `yaml:"max_retries"`
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

// DeFiConfig represents the DeFi protocol configuration
type DeFiConfig struct {
	Uniswap   UniswapConfig   `yaml:"uniswap"`
	Aave      AaveConfig      `yaml:"aave"`
	Chainlink ChainlinkConfig `yaml:"chainlink"`
	OneInch   OneInchConfig   `yaml:"oneinch"`
	Coffee    CoffeeConfig    `yaml:"coffee"`
}

// UniswapConfig represents the Uniswap configuration
type UniswapConfig struct {
	Enabled         bool    `yaml:"enabled"`
	FactoryAddress  string  `yaml:"factory_address"`
	RouterAddress   string  `yaml:"router_address"`
	QuoterAddress   string  `yaml:"quoter_address"`
	PositionManager string  `yaml:"position_manager"`
	DefaultSlippage float64 `yaml:"default_slippage"`
	DefaultDeadline int     `yaml:"default_deadline"`
}

// AaveConfig represents the Aave configuration
type AaveConfig struct {
	Enabled             bool   `yaml:"enabled"`
	PoolAddress         string `yaml:"pool_address"`
	DataProviderAddress string `yaml:"data_provider_address"`
	OracleAddress       string `yaml:"oracle_address"`
	RewardsController   string `yaml:"rewards_controller"`
}

// ChainlinkConfig represents the Chainlink configuration
type ChainlinkConfig struct {
	Enabled    bool              `yaml:"enabled"`
	PriceFeeds map[string]string `yaml:"price_feeds"`
}

// OneInchConfig represents the 1inch configuration
type OneInchConfig struct {
	Enabled bool   `yaml:"enabled"`
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
}

// CoffeeConfig represents the Coffee Token configuration
type CoffeeConfig struct {
	Enabled         bool              `yaml:"enabled"`
	TokenAddresses  map[string]string `yaml:"token_addresses"`
	StakingContract string            `yaml:"staking_contract"`
	RewardsAPY      float64           `yaml:"rewards_apy"`
	MinStakeAmount  string            `yaml:"min_stake_amount"`
}

// ServicesConfig represents the services configuration
type ServicesConfig struct {
	APIGateway           ServiceConfig `yaml:"api_gateway"`
	WalletService        ServiceConfig `yaml:"wallet_service"`
	TransactionService   ServiceConfig `yaml:"transaction_service"`
	SmartContractService ServiceConfig `yaml:"smart_contract_service"`
	SecurityService      ServiceConfig `yaml:"security_service"`
	DeFiService          ServiceConfig `yaml:"defi_service"`
	TelegramBot          ServiceConfig `yaml:"telegram_bot"`
}

// ServiceConfig represents the configuration for a service
type ServiceConfig struct {
	Host     string `yaml:"host"`
	HTTPPort int    `yaml:"http_port"`
	GRPCPort int    `yaml:"grpc_port"`
	Enabled  bool   `yaml:"enabled"`
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

// Load loads the configuration from a file (alias for LoadConfig)
func Load(configPath string) (*Config, error) {
	return LoadConfig(configPath)
}

// TelegramConfig represents the Telegram bot configuration
type TelegramConfig struct {
	Enabled        bool                    `yaml:"enabled"`
	BotToken       string                  `yaml:"bot_token"`
	WebhookURL     string                  `yaml:"webhook_url"`
	WebhookPort    int                     `yaml:"webhook_port"`
	WebhookPath    string                  `yaml:"webhook_path"`
	Debug          bool                    `yaml:"debug"`
	Timeout        int                     `yaml:"timeout"`
	MaxConnections int                     `yaml:"max_connections"`
	AllowedUpdates []string                `yaml:"allowed_updates"`
	Commands       []TelegramCommandConfig `yaml:"commands"`
}

// TelegramCommandConfig represents a Telegram bot command configuration
type TelegramCommandConfig struct {
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
}

// AIConfig represents the AI services configuration
type AIConfig struct {
	Enabled   bool            `yaml:"enabled"`
	LangChain LangChainConfig `yaml:"langchain"`
	Gemini    GeminiConfig    `yaml:"gemini"`
	Ollama    OllamaConfig    `yaml:"ollama"`
	Service   AIServiceConfig `yaml:"service"`
}

// LangChainConfig represents the LangChain configuration
type LangChainConfig struct {
	Enabled     bool    `yaml:"enabled"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
	Timeout     int     `yaml:"timeout"`
}

// GeminiConfig represents the Google Gemini configuration
type GeminiConfig struct {
	Enabled        bool                 `yaml:"enabled"`
	APIKey         string               `yaml:"api_key"`
	Model          string               `yaml:"model"`
	Temperature    float64              `yaml:"temperature"`
	MaxTokens      int                  `yaml:"max_tokens"`
	Timeout        int                  `yaml:"timeout"`
	SafetySettings GeminiSafetySettings `yaml:"safety_settings"`
}

// GeminiSafetySettings represents Gemini safety settings
type GeminiSafetySettings struct {
	Harassment       string `yaml:"harassment"`
	HateSpeech       string `yaml:"hate_speech"`
	SexuallyExplicit string `yaml:"sexually_explicit"`
	DangerousContent string `yaml:"dangerous_content"`
}

// OllamaConfig represents the Ollama configuration
type OllamaConfig struct {
	Enabled     bool    `yaml:"enabled"`
	Host        string  `yaml:"host"`
	Port        int     `yaml:"port"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	Timeout     int     `yaml:"timeout"`
	KeepAlive   string  `yaml:"keep_alive"`
}

// AIServiceConfig represents the AI service configuration
type AIServiceConfig struct {
	DefaultProvider  string            `yaml:"default_provider"`
	FallbackProvider string            `yaml:"fallback_provider"`
	CacheEnabled     bool              `yaml:"cache_enabled"`
	CacheTTL         string            `yaml:"cache_ttl"`
	RateLimit        AIRateLimitConfig `yaml:"rate_limit"`
}

// AIRateLimitConfig represents AI rate limiting configuration
type AIRateLimitConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
	Burst             int `yaml:"burst"`
}
