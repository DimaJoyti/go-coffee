package config

import (
	"fmt"
	"os"
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
	Fintech      FintechConfig      `yaml:"fintech"`
	MultiRegion  MultiRegionConfig  `yaml:"multi_region"`
}

// ServerConfig represents the HTTP server configuration
type ServerConfig struct {
	Port           int           `yaml:"port"`
	Host           string        `yaml:"host"`
	Environment    string        `yaml:"environment"`
	Timeout        time.Duration `yaml:"timeout"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
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
	Addresses              []string      `yaml:"addresses"`
	Host                   string        `yaml:"host"`
	Port                   int           `yaml:"port"`
	Password               string        `yaml:"password"`
	DB                     int           `yaml:"db"`
	PoolSize               int           `yaml:"pool_size"`
	MinIdleConns           int           `yaml:"min_idle_conns"`
	DialTimeout            time.Duration `yaml:"dial_timeout"`
	ReadTimeout            time.Duration `yaml:"read_timeout"`
	WriteTimeout           time.Duration `yaml:"write_timeout"`
	PoolTimeout            time.Duration `yaml:"pool_timeout"`
	IdleTimeout            time.Duration `yaml:"idle_timeout"`
	IdleCheckFrequency     time.Duration `yaml:"idle_check_frequency"`
	MaxRetries             int           `yaml:"max_retries"`
	MinRetryBackoff        time.Duration `yaml:"min_retry_backoff"`
	MaxRetryBackoff        time.Duration `yaml:"max_retry_backoff"`
	EnableCluster          bool          `yaml:"enable_cluster"`
	RouteByLatency         bool          `yaml:"route_by_latency"`
	RouteRandomly          bool          `yaml:"route_randomly"`
	EnableReadFromReplicas bool          `yaml:"enable_read_from_replicas"`
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
	data, err := os.ReadFile(configPath)
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
	Reddit    RedditConfig    `yaml:"reddit"`
	RAG       RAGConfig       `yaml:"rag"`
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

// RedditConfig represents the Reddit API configuration
type RedditConfig struct {
	Enabled      bool                  `yaml:"enabled"`
	ClientID     string                `yaml:"client_id"`
	ClientSecret string                `yaml:"client_secret"`
	UserAgent    string                `yaml:"user_agent"`
	Username     string                `yaml:"username"`
	Password     string                `yaml:"password"`
	BaseURL      string                `yaml:"base_url"`
	RateLimit    RedditRateLimitConfig `yaml:"rate_limit"`
	Subreddits   []string              `yaml:"subreddits"`
	ContentTypes []string              `yaml:"content_types"`
}

// RedditRateLimitConfig represents Reddit API rate limiting
type RedditRateLimitConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
	BurstSize         int `yaml:"burst_size"`
	RetryDelay        int `yaml:"retry_delay"`
}

// RAGConfig represents the RAG (Retrieval-Augmented Generation) configuration
type RAGConfig struct {
	Enabled         bool                  `yaml:"enabled"`
	VectorDB        VectorDBConfig        `yaml:"vector_db"`
	Embeddings      EmbeddingsConfig      `yaml:"embeddings"`
	Retrieval       RetrievalConfig       `yaml:"retrieval"`
	ContentAnalysis ContentAnalysisConfig `yaml:"content_analysis"`
}

// VectorDBConfig represents vector database configuration
type VectorDBConfig struct {
	Provider    string `yaml:"provider"` // pinecone, weaviate, qdrant
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	APIKey      string `yaml:"api_key"`
	Environment string `yaml:"environment"`
	IndexName   string `yaml:"index_name"`
	Dimension   int    `yaml:"dimension"`
}

// EmbeddingsConfig represents embeddings configuration
type EmbeddingsConfig struct {
	Provider    string  `yaml:"provider"` // openai, sentence-transformers, local
	Model       string  `yaml:"model"`
	APIKey      string  `yaml:"api_key"`
	Dimension   int     `yaml:"dimension"`
	BatchSize   int     `yaml:"batch_size"`
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float64 `yaml:"temperature"`
}

// RetrievalConfig represents retrieval configuration
type RetrievalConfig struct {
	TopK           int     `yaml:"top_k"`
	ScoreThreshold float64 `yaml:"score_threshold"`
	MaxDocuments   int     `yaml:"max_documents"`
	ContextWindow  int     `yaml:"context_window"`
	RerankerModel  string  `yaml:"reranker_model"`
}

// ContentAnalysisConfig represents content analysis configuration
type ContentAnalysisConfig struct {
	Enabled             bool     `yaml:"enabled"`
	ClassificationModel string   `yaml:"classification_model"`
	SentimentAnalysis   bool     `yaml:"sentiment_analysis"`
	TopicModeling       bool     `yaml:"topic_modeling"`
	TrendAnalysis       bool     `yaml:"trend_analysis"`
	Languages           []string `yaml:"languages"`
	Categories          []string `yaml:"categories"`
	MinConfidence       float64  `yaml:"min_confidence"`
}

// FintechConfig represents the fintech platform configuration
type FintechConfig struct {
	Accounts AccountsConfig `yaml:"accounts"`
	Payments PaymentsConfig `yaml:"payments"`
	Yield    YieldConfig    `yaml:"yield"`
	Trading  TradingConfig  `yaml:"trading"`
	Cards    CardsConfig    `yaml:"cards"`
}

// AccountsConfig represents the accounts module configuration
type AccountsConfig struct {
	Enabled              bool                 `yaml:"enabled"`
	KYCRequired          bool                 `yaml:"kyc_required"`
	KYCProvider          string               `yaml:"kyc_provider"`
	TwoFactorAuth        bool                 `yaml:"two_factor_auth"`
	SessionTimeout       string               `yaml:"session_timeout"`
	MaxLoginAttempts     int                  `yaml:"max_login_attempts"`
	PasswordPolicy       PasswordPolicyConfig `yaml:"password_policy"`
	AccountLimits        AccountLimitsConfig  `yaml:"account_limits"`
	ComplianceChecks     ComplianceConfig     `yaml:"compliance_checks"`
	NotificationSettings NotificationSettings `yaml:"notification_settings"`
}

// PasswordPolicyConfig represents password policy configuration
type PasswordPolicyConfig struct {
	MinLength        int  `yaml:"min_length"`
	RequireUppercase bool `yaml:"require_uppercase"`
	RequireLowercase bool `yaml:"require_lowercase"`
	RequireNumbers   bool `yaml:"require_numbers"`
	RequireSymbols   bool `yaml:"require_symbols"`
	ExpirationDays   int  `yaml:"expiration_days"`
}

// AccountLimitsConfig represents account limits configuration
type AccountLimitsConfig struct {
	DailyTransactionLimit   string `yaml:"daily_transaction_limit"`
	MonthlyTransactionLimit string `yaml:"monthly_transaction_limit"`
	MaxWalletsPerUser       int    `yaml:"max_wallets_per_user"`
	MaxCardsPerUser         int    `yaml:"max_cards_per_user"`
	MinAccountBalance       string `yaml:"min_account_balance"`
}

// ComplianceConfig represents compliance configuration
type ComplianceConfig struct {
	AMLEnabled        bool     `yaml:"aml_enabled"`
	SanctionsCheck    bool     `yaml:"sanctions_check"`
	PEPCheck          bool     `yaml:"pep_check"`
	RiskScoring       bool     `yaml:"risk_scoring"`
	TransactionLimits bool     `yaml:"transaction_limits"`
	ReportingRequired bool     `yaml:"reporting_required"`
	AllowedCountries  []string `yaml:"allowed_countries"`
	BlockedCountries  []string `yaml:"blocked_countries"`
}

// NotificationSettings represents notification settings
type NotificationSettings struct {
	EmailEnabled      bool `yaml:"email_enabled"`
	SMSEnabled        bool `yaml:"sms_enabled"`
	PushEnabled       bool `yaml:"push_enabled"`
	WebhookEnabled    bool `yaml:"webhook_enabled"`
	SecurityAlerts    bool `yaml:"security_alerts"`
	TransactionAlerts bool `yaml:"transaction_alerts"`
}

// PaymentsConfig represents the payments module configuration
type PaymentsConfig struct {
	Enabled             bool                  `yaml:"enabled"`
	SupportedCurrencies []string              `yaml:"supported_currencies"`
	SupportedNetworks   []string              `yaml:"supported_networks"`
	DefaultNetwork      string                `yaml:"default_network"`
	TransactionFees     TransactionFeesConfig `yaml:"transaction_fees"`
	PaymentMethods      PaymentMethodsConfig  `yaml:"payment_methods"`
	FraudDetection      FraudDetectionConfig  `yaml:"fraud_detection"`
	Settlement          SettlementConfig      `yaml:"settlement"`
	Webhooks            WebhooksConfig        `yaml:"webhooks"`
	Reconciliation      ReconciliationConfig  `yaml:"reconciliation"`
}

// TransactionFeesConfig represents transaction fees configuration
type TransactionFeesConfig struct {
	FeeStructure       string  `yaml:"fee_structure"` // flat, percentage, tiered
	BaseFee            string  `yaml:"base_fee"`
	PercentageFee      float64 `yaml:"percentage_fee"`
	MinFee             string  `yaml:"min_fee"`
	MaxFee             string  `yaml:"max_fee"`
	NetworkFeeMarkup   float64 `yaml:"network_fee_markup"`
	PriorityFeeEnabled bool    `yaml:"priority_fee_enabled"`
}

// PaymentMethodsConfig represents payment methods configuration
type PaymentMethodsConfig struct {
	CryptoEnabled    bool     `yaml:"crypto_enabled"`
	FiatEnabled      bool     `yaml:"fiat_enabled"`
	StablecoinOnly   bool     `yaml:"stablecoin_only"`
	SupportedTokens  []string `yaml:"supported_tokens"`
	MinPaymentAmount string   `yaml:"min_payment_amount"`
	MaxPaymentAmount string   `yaml:"max_payment_amount"`
}

// FraudDetectionConfig represents fraud detection configuration
type FraudDetectionConfig struct {
	Enabled            bool    `yaml:"enabled"`
	RiskScoreThreshold float64 `yaml:"risk_score_threshold"`
	VelocityChecks     bool    `yaml:"velocity_checks"`
	GeolocationChecks  bool    `yaml:"geolocation_checks"`
	DeviceFingerprint  bool    `yaml:"device_fingerprint"`
	MLModelEnabled     bool    `yaml:"ml_model_enabled"`
}

// SettlementConfig represents settlement configuration
type SettlementConfig struct {
	AutoSettlement      bool   `yaml:"auto_settlement"`
	SettlementSchedule  string `yaml:"settlement_schedule"`
	MinSettlementAmount string `yaml:"min_settlement_amount"`
	SettlementCurrency  string `yaml:"settlement_currency"`
	HoldPeriod          string `yaml:"hold_period"`
}

// WebhooksConfig represents webhooks configuration
type WebhooksConfig struct {
	Enabled       bool     `yaml:"enabled"`
	RetryAttempts int      `yaml:"retry_attempts"`
	RetryDelay    string   `yaml:"retry_delay"`
	Timeout       string   `yaml:"timeout"`
	SigningSecret string   `yaml:"signing_secret"`
	AllowedEvents []string `yaml:"allowed_events"`
}

// ReconciliationConfig represents reconciliation configuration
type ReconciliationConfig struct {
	Enabled           bool   `yaml:"enabled"`
	Schedule          string `yaml:"schedule"`
	ToleranceAmount   string `yaml:"tolerance_amount"`
	AutoResolve       bool   `yaml:"auto_resolve"`
	NotifyDiscrepancy bool   `yaml:"notify_discrepancy"`
}

// YieldConfig represents the yield module configuration
type YieldConfig struct {
	Enabled             bool                    `yaml:"enabled"`
	SupportedProtocols  []string                `yaml:"supported_protocols"`
	DefaultStrategy     string                  `yaml:"default_strategy"`
	AutoCompounding     bool                    `yaml:"auto_compounding"`
	RiskManagement      YieldRiskConfig         `yaml:"risk_management"`
	StakingPools        StakingPoolsConfig      `yaml:"staking_pools"`
	LiquidityMining     LiquidityMiningConfig   `yaml:"liquidity_mining"`
	YieldOptimization   YieldOptimizationConfig `yaml:"yield_optimization"`
	RewardsDistribution RewardsConfig           `yaml:"rewards_distribution"`
	PerformanceTracking PerformanceConfig       `yaml:"performance_tracking"`
}

// YieldRiskConfig represents yield risk management configuration
type YieldRiskConfig struct {
	MaxAllocation        float64  `yaml:"max_allocation"`
	DiversificationRules bool     `yaml:"diversification_rules"`
	RiskScoreThreshold   float64  `yaml:"risk_score_threshold"`
	ImpermanentLossLimit float64  `yaml:"impermanent_loss_limit"`
	AllowedProtocols     []string `yaml:"allowed_protocols"`
	BlacklistedTokens    []string `yaml:"blacklisted_tokens"`
}

// StakingPoolsConfig represents staking pools configuration
type StakingPoolsConfig struct {
	Enabled            bool     `yaml:"enabled"`
	MinStakeAmount     string   `yaml:"min_stake_amount"`
	MaxStakeAmount     string   `yaml:"max_stake_amount"`
	UnstakingPeriod    string   `yaml:"unstaking_period"`
	SupportedTokens    []string `yaml:"supported_tokens"`
	AutoRestaking      bool     `yaml:"auto_restaking"`
	SlashingProtection bool     `yaml:"slashing_protection"`
}

// LiquidityMiningConfig represents liquidity mining configuration
type LiquidityMiningConfig struct {
	Enabled              bool     `yaml:"enabled"`
	SupportedPairs       []string `yaml:"supported_pairs"`
	MinLiquidityAmount   string   `yaml:"min_liquidity_amount"`
	MaxLiquidityAmount   string   `yaml:"max_liquidity_amount"`
	ImpermanentLossAlert bool     `yaml:"impermanent_loss_alert"`
	AutoRebalancing      bool     `yaml:"auto_rebalancing"`
	FeeHarvesting        bool     `yaml:"fee_harvesting"`
}

// YieldOptimizationConfig represents yield optimization configuration
type YieldOptimizationConfig struct {
	Enabled              bool    `yaml:"enabled"`
	OptimizationInterval string  `yaml:"optimization_interval"`
	GasOptimization      bool    `yaml:"gas_optimization"`
	YieldThreshold       float64 `yaml:"yield_threshold"`
	AutoMigration        bool    `yaml:"auto_migration"`
	CompoundingFrequency string  `yaml:"compounding_frequency"`
}

// RewardsConfig represents rewards distribution configuration
type RewardsConfig struct {
	AutoClaim         bool   `yaml:"auto_claim"`
	ClaimThreshold    string `yaml:"claim_threshold"`
	ReinvestRewards   bool   `yaml:"reinvest_rewards"`
	RewardsToken      string `yaml:"rewards_token"`
	DistributionDelay string `yaml:"distribution_delay"`
}

// PerformanceConfig represents performance tracking configuration
type PerformanceConfig struct {
	Enabled          bool   `yaml:"enabled"`
	TrackingInterval string `yaml:"tracking_interval"`
	BenchmarkEnabled bool   `yaml:"benchmark_enabled"`
	ReportGeneration bool   `yaml:"report_generation"`
	AlertsEnabled    bool   `yaml:"alerts_enabled"`
}

// TradingConfig represents the trading module configuration
type TradingConfig struct {
	Enabled             bool                     `yaml:"enabled"`
	SupportedExchanges  []string                 `yaml:"supported_exchanges"`
	DefaultExchange     string                   `yaml:"default_exchange"`
	TradingPairs        []string                 `yaml:"trading_pairs"`
	OrderTypes          []string                 `yaml:"order_types"`
	RiskManagement      TradingRiskConfig        `yaml:"risk_management"`
	AlgorithmicTrading  AlgorithmicTradingConfig `yaml:"algorithmic_trading"`
	MarketData          MarketDataConfig         `yaml:"market_data"`
	ExecutionEngine     ExecutionEngineConfig    `yaml:"execution_engine"`
	PortfolioManagement PortfolioConfig          `yaml:"portfolio_management"`
}

// TradingRiskConfig represents trading risk management configuration
type TradingRiskConfig struct {
	MaxPositionSize    float64 `yaml:"max_position_size"`
	MaxDailyLoss       float64 `yaml:"max_daily_loss"`
	StopLossRequired   bool    `yaml:"stop_loss_required"`
	TakeProfitRequired bool    `yaml:"take_profit_required"`
	MaxLeverage        float64 `yaml:"max_leverage"`
	RiskScoreThreshold float64 `yaml:"risk_score_threshold"`
	VolatilityLimit    float64 `yaml:"volatility_limit"`
}

// AlgorithmicTradingConfig represents algorithmic trading configuration
type AlgorithmicTradingConfig struct {
	Enabled             bool     `yaml:"enabled"`
	SupportedStrategies []string `yaml:"supported_strategies"`
	BacktestingEnabled  bool     `yaml:"backtesting_enabled"`
	PaperTradingEnabled bool     `yaml:"paper_trading_enabled"`
	MaxActiveStrategies int      `yaml:"max_active_strategies"`
	StrategyAllocation  float64  `yaml:"strategy_allocation"`
}

// MarketDataConfig represents market data configuration
type MarketDataConfig struct {
	Enabled           bool     `yaml:"enabled"`
	DataProviders     []string `yaml:"data_providers"`
	UpdateFrequency   string   `yaml:"update_frequency"`
	HistoricalData    bool     `yaml:"historical_data"`
	RealtimeData      bool     `yaml:"realtime_data"`
	TechnicalAnalysis bool     `yaml:"technical_analysis"`
}

// ExecutionEngineConfig represents execution engine configuration
type ExecutionEngineConfig struct {
	Enabled             bool    `yaml:"enabled"`
	OrderRouting        bool    `yaml:"order_routing"`
	SmartOrderRouting   bool    `yaml:"smart_order_routing"`
	SlippageProtection  bool    `yaml:"slippage_protection"`
	MaxSlippage         float64 `yaml:"max_slippage"`
	PartialFillsEnabled bool    `yaml:"partial_fills_enabled"`
	TimeInForce         string  `yaml:"time_in_force"`
}

// PortfolioConfig represents portfolio management configuration
type PortfolioConfig struct {
	Enabled              bool    `yaml:"enabled"`
	AutoRebalancing      bool    `yaml:"auto_rebalancing"`
	RebalancingThreshold float64 `yaml:"rebalancing_threshold"`
	DiversificationRules bool    `yaml:"diversification_rules"`
	PerformanceTracking  bool    `yaml:"performance_tracking"`
	RiskAnalysis         bool    `yaml:"risk_analysis"`
}

// CardsConfig represents the cards module configuration
type CardsConfig struct {
	Enabled               bool                        `yaml:"enabled"`
	SupportedCardTypes    []string                    `yaml:"supported_card_types"`
	DefaultCardType       string                      `yaml:"default_card_type"`
	VirtualCards          VirtualCardsConfig          `yaml:"virtual_cards"`
	PhysicalCards         PhysicalCardsConfig         `yaml:"physical_cards"`
	CardSecurity          CardSecurityConfig          `yaml:"card_security"`
	SpendingControls      SpendingControlsConfig      `yaml:"spending_controls"`
	CardManagement        CardManagementConfig        `yaml:"card_management"`
	TransactionProcessing TransactionProcessingConfig `yaml:"transaction_processing"`
	RewardsProgram        CardRewardsConfig           `yaml:"rewards_program"`
}

// VirtualCardsConfig represents virtual cards configuration
type VirtualCardsConfig struct {
	Enabled               bool     `yaml:"enabled"`
	InstantIssuance       bool     `yaml:"instant_issuance"`
	MaxCardsPerUser       int      `yaml:"max_cards_per_user"`
	DefaultExpiryPeriod   string   `yaml:"default_expiry_period"`
	SupportedNetworks     []string `yaml:"supported_networks"`
	SingleUseCards        bool     `yaml:"single_use_cards"`
	MerchantSpecificCards bool     `yaml:"merchant_specific_cards"`
}

// PhysicalCardsConfig represents physical cards configuration
type PhysicalCardsConfig struct {
	Enabled             bool     `yaml:"enabled"`
	IssuanceEnabled     bool     `yaml:"issuance_enabled"`
	ShippingEnabled     bool     `yaml:"shipping_enabled"`
	ShippingCost        string   `yaml:"shipping_cost"`
	ProductionTime      string   `yaml:"production_time"`
	SupportedRegions    []string `yaml:"supported_regions"`
	CardDesigns         []string `yaml:"card_designs"`
	CustomDesignEnabled bool     `yaml:"custom_design_enabled"`
}

// CardSecurityConfig represents card security configuration
type CardSecurityConfig struct {
	CVVRotation         bool   `yaml:"cvv_rotation"`
	CVVRotationInterval string `yaml:"cvv_rotation_interval"`
	TokenizationEnabled bool   `yaml:"tokenization_enabled"`
	BiometricAuth       bool   `yaml:"biometric_auth"`
	PINRequired         bool   `yaml:"pin_required"`
	FraudDetection      bool   `yaml:"fraud_detection"`
	VelocityChecks      bool   `yaml:"velocity_checks"`
	GeofencingEnabled   bool   `yaml:"geofencing_enabled"`
}

// SpendingControlsConfig represents spending controls configuration
type SpendingControlsConfig struct {
	Enabled            bool     `yaml:"enabled"`
	DailyLimits        bool     `yaml:"daily_limits"`
	MonthlyLimits      bool     `yaml:"monthly_limits"`
	TransactionLimits  bool     `yaml:"transaction_limits"`
	MerchantCategories bool     `yaml:"merchant_categories"`
	GeographicControls bool     `yaml:"geographic_controls"`
	TimeBasedControls  bool     `yaml:"time_based_controls"`
	AllowedMerchants   []string `yaml:"allowed_merchants"`
	BlockedMerchants   []string `yaml:"blocked_merchants"`
}

// CardManagementConfig represents card management configuration
type CardManagementConfig struct {
	Enabled             bool   `yaml:"enabled"`
	SelfServiceEnabled  bool   `yaml:"self_service_enabled"`
	InstantActivation   bool   `yaml:"instant_activation"`
	InstantSuspension   bool   `yaml:"instant_suspension"`
	InstantReplacement  bool   `yaml:"instant_replacement"`
	BulkOperations      bool   `yaml:"bulk_operations"`
	AutoRenewal         bool   `yaml:"auto_renewal"`
	RenewalNotification string `yaml:"renewal_notification"`
}

// TransactionProcessingConfig represents transaction processing configuration
type TransactionProcessingConfig struct {
	Enabled              bool    `yaml:"enabled"`
	RealtimeProcessing   bool    `yaml:"realtime_processing"`
	AuthorizationTimeout string  `yaml:"authorization_timeout"`
	SettlementDelay      string  `yaml:"settlement_delay"`
	DeclineReasons       bool    `yaml:"decline_reasons"`
	PartialApprovals     bool    `yaml:"partial_approvals"`
	CurrencyConversion   bool    `yaml:"currency_conversion"`
	FXMarkup             float64 `yaml:"fx_markup"`
}

// CardRewardsConfig represents card rewards program configuration
type CardRewardsConfig struct {
	Enabled             bool               `yaml:"enabled"`
	RewardsType         string             `yaml:"rewards_type"` // cashback, points, crypto
	CashbackRate        float64            `yaml:"cashback_rate"`
	PointsMultiplier    float64            `yaml:"points_multiplier"`
	CryptoRewards       bool               `yaml:"crypto_rewards"`
	RewardsToken        string             `yaml:"rewards_token"`
	CategoryMultipliers map[string]float64 `yaml:"category_multipliers"`
	RedemptionOptions   []string           `yaml:"redemption_options"`
	MinRedemptionAmount string             `yaml:"min_redemption_amount"`
}

// MultiRegionConfig represents multi-region configuration
type MultiRegionConfig struct {
	Enabled  bool             `yaml:"enabled"`
	Regions  []Region         `yaml:"regions"`
	Failover FailoverConfig   `yaml:"failover"`
}

// Region represents a deployment region
type Region struct {
	Name     string `yaml:"name"`
	Priority int    `yaml:"priority"`
	Endpoint string `yaml:"endpoint"`
	Healthy  bool   `yaml:"healthy"`
}

// FailoverConfig represents failover configuration
type FailoverConfig struct {
	Enabled            bool          `yaml:"enabled"`
	CheckInterval      time.Duration `yaml:"check_interval"`
	Timeout            time.Duration `yaml:"timeout"`
	FailureThreshold   int           `yaml:"failure_threshold"`
	RecoveryThreshold  int           `yaml:"recovery_threshold"`
}

// GetCurrentRegion returns the current region (first healthy region with highest priority)
func (c *MultiRegionConfig) GetCurrentRegion() *Region {
	if !c.Enabled || len(c.Regions) == 0 {
		return nil
	}

	// Find the highest priority healthy region
	var currentRegion *Region
	for i := range c.Regions {
		region := &c.Regions[i]
		if region.Healthy {
			if currentRegion == nil || region.Priority < currentRegion.Priority {
				currentRegion = region
			}
		}
	}

	return currentRegion
}
