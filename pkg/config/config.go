package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents the application configuration
type Config struct {
	Environment string `json:"environment"`
	Debug       bool   `json:"debug"`
	LogLevel    string `json:"log_level"`
	LogFormat   string `json:"log_format"`

	// Server configuration
	Server ServerConfig `json:"server"`

	// Database configuration
	Database DatabaseConfig `json:"database"`

	// Redis configuration
	Redis RedisConfig `json:"redis"`

	// Kafka configuration
	Kafka KafkaConfig `json:"kafka"`

	// Security configuration
	Security SecurityConfig `json:"security"`

	// Web3 configuration
	Web3 Web3Config `json:"web3"`

	// AI configuration
	AI AIConfig `json:"ai"`

	// External integrations
	Integrations IntegrationsConfig `json:"integrations"`

	// Monitoring configuration
	Monitoring MonitoringConfig `json:"monitoring"`

	// Feature flags
	Features FeatureFlags `json:"features"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	APIGatewayPort    int    `json:"api_gateway_port"`
	ProducerPort      int    `json:"producer_port"`
	ConsumerPort      int    `json:"consumer_port"`
	StreamsPort       int    `json:"streams_port"`
	AISearchPort      int    `json:"ai_search_port"`
	AuthServicePort   int    `json:"auth_service_port"`
	PaymentServicePort int   `json:"payment_service_port"`
	OrderServicePort   int   `json:"order_service_port"`
	KitchenServicePort int   `json:"kitchen_service_port"`
	Host              string `json:"host"`
	ReadTimeout       string `json:"read_timeout"`
	WriteTimeout      string `json:"write_timeout"`
	IdleTimeout       string `json:"idle_timeout"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Name            string `json:"name"`
	User            string `json:"user"`
	Password        string `json:"password"`
	SSLMode         string `json:"ssl_mode"`
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime string `json:"conn_max_lifetime"`
	TestDBName      string `json:"test_db_name"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Password     string `json:"password"`
	DB           int    `json:"db"`
	PoolSize     int    `json:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns"`
	DialTimeout  string `json:"dial_timeout"`
	ReadTimeout  string `json:"read_timeout"`
	WriteTimeout string `json:"write_timeout"`
	URL          string `json:"url"`
}

// KafkaConfig represents Kafka configuration
type KafkaConfig struct {
	Brokers         []string `json:"brokers"`
	Topic           string   `json:"topic"`
	ProcessedTopic  string   `json:"processed_topic"`
	ConsumerGroup   string   `json:"consumer_group"`
	WorkerPoolSize  int      `json:"worker_pool_size"`
	RetryMax        int      `json:"retry_max"`
	RequiredAcks    string   `json:"required_acks"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	JWTSecret           string `json:"jwt_secret"`
	JWTExpiry           string `json:"jwt_expiry"`
	RefreshTokenExpiry  string `json:"refresh_token_expiry"`
	JWTIssuer           string `json:"jwt_issuer"`
	JWTAudience         string `json:"jwt_audience"`
	APIKeySecret        string `json:"api_key_secret"`
	WebhookSecret       string `json:"webhook_secret"`
	EncryptionKey       string `json:"encryption_key"`
}

// Web3Config represents Web3 configuration
type Web3Config struct {
	Ethereum EthereumConfig `json:"ethereum"`
	Bitcoin  BitcoinConfig  `json:"bitcoin"`
	Solana   SolanaConfig   `json:"solana"`
	DeFi     DeFiConfig     `json:"defi"`
}

// EthereumConfig represents Ethereum configuration
type EthereumConfig struct {
	RPCURL      string `json:"rpc_url"`
	TestnetURL  string `json:"testnet_url"`
	PrivateKey  string `json:"private_key"`
	GasLimit    int64  `json:"gas_limit"`
	GasPrice    int64  `json:"gas_price"`
}

// BitcoinConfig represents Bitcoin configuration
type BitcoinConfig struct {
	RPCURL      string `json:"rpc_url"`
	RPCUsername string `json:"rpc_username"`
	RPCPassword string `json:"rpc_password"`
}

// SolanaConfig represents Solana configuration
type SolanaConfig struct {
	RPCURL     string `json:"rpc_url"`
	TestnetURL string `json:"testnet_url"`
	PrivateKey string `json:"private_key"`
}

// DeFiConfig represents DeFi configuration
type DeFiConfig struct {
	UniswapV3Router      string `json:"uniswap_v3_router"`
	AaveLendingPool      string `json:"aave_lending_pool"`
	CompoundComptroller  string `json:"compound_comptroller"`

	// Enhanced DeFi configuration
	OneInchAPIKey        string `json:"oneinch_api_key"`
	ChainlinkEnabled     bool   `json:"chainlink_enabled"`
	ArbitrageEnabled     bool   `json:"arbitrage_enabled"`
	YieldFarmingEnabled  bool   `json:"yield_farming_enabled"`
	TradingBotsEnabled   bool   `json:"trading_bots_enabled"`
}

// AIConfig represents AI configuration
type AIConfig struct {
	GeminiAPIKey            string `json:"gemini_api_key"`
	OpenAIAPIKey            string `json:"openai_api_key"`
	OllamaURL               string `json:"ollama_url"`
	SearchEmbeddingModel    string `json:"search_embedding_model"`
	SearchVectorDimensions  int    `json:"search_vector_dimensions"`
	SearchSimilarityThreshold float64 `json:"search_similarity_threshold"`
	SearchMaxResults        int    `json:"search_max_results"`
}

// IntegrationsConfig represents external integrations configuration
type IntegrationsConfig struct {
	SMTP      SMTPConfig      `json:"smtp"`
	Twilio    TwilioConfig    `json:"twilio"`
	Slack     SlackConfig     `json:"slack"`
	ClickUp   ClickUpConfig   `json:"clickup"`
	GoogleSheets GoogleSheetsConfig `json:"google_sheets"`
}

// SMTPConfig represents SMTP configuration
type SMTPConfig struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name"`
}

// TwilioConfig represents Twilio configuration
type TwilioConfig struct {
	AccountSID string `json:"account_sid"`
	AuthToken  string `json:"auth_token"`
	FromNumber string `json:"from_number"`
}

// SlackConfig represents Slack configuration
type SlackConfig struct {
	BotToken   string `json:"bot_token"`
	WebhookURL string `json:"webhook_url"`
}

// ClickUpConfig represents ClickUp configuration
type ClickUpConfig struct {
	APIToken string `json:"api_token"`
	TeamID   string `json:"team_id"`
}

// GoogleSheetsConfig represents Google Sheets configuration
type GoogleSheetsConfig struct {
	CredentialsPath string `json:"credentials_path"`
	SpreadsheetID   string `json:"spreadsheet_id"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	Prometheus PrometheusConfig `json:"prometheus"`
	Grafana    GrafanaConfig    `json:"grafana"`
	Jaeger     JaegerConfig     `json:"jaeger"`
	Sentry     SentryConfig     `json:"sentry"`
}

// PrometheusConfig represents Prometheus configuration
type PrometheusConfig struct {
	Enabled     bool   `json:"enabled"`
	Port        int    `json:"port"`
	MetricsPath string `json:"metrics_path"`
}

// GrafanaConfig represents Grafana configuration
type GrafanaConfig struct {
	Port          int    `json:"port"`
	AdminUser     string `json:"admin_user"`
	AdminPassword string `json:"admin_password"`
}

// JaegerConfig represents Jaeger configuration
type JaegerConfig struct {
	Enabled      bool   `json:"enabled"`
	Endpoint     string `json:"endpoint"`
	SamplerType  string `json:"sampler_type"`
	SamplerParam string `json:"sampler_param"`
}

// SentryConfig represents Sentry configuration
type SentryConfig struct {
	DSN         string `json:"dsn"`
	Environment string `json:"environment"`
	Release     string `json:"release"`
}

// FeatureFlags represents feature flags configuration
type FeatureFlags struct {
	ProducerServiceEnabled      bool `json:"producer_service_enabled"`
	ConsumerServiceEnabled      bool `json:"consumer_service_enabled"`
	StreamsServiceEnabled       bool `json:"streams_service_enabled"`
	APIGatewayEnabled           bool `json:"api_gateway_enabled"`
	Web3WalletEnabled           bool `json:"web3_wallet_enabled"`
	DeFiServiceEnabled          bool `json:"defi_service_enabled"`
	SmartContractServiceEnabled bool `json:"smart_contract_service_enabled"`
	AISearchEnabled             bool `json:"ai_search_enabled"`
	AIAgentsEnabled             bool `json:"ai_agents_enabled"`
	AuthModuleEnabled           bool `json:"auth_module_enabled"`
	PaymentModuleEnabled        bool `json:"payment_module_enabled"`
	NotificationModuleEnabled   bool `json:"notification_module_enabled"`
}

// LoadConfig завантажує конфігурацію з файлу
func LoadConfigFromFile(filePath string, config interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	return nil
}

// GetEnv отримує значення змінної середовища або повертає значення за замовчуванням
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// GetEnvAsInt отримує значення змінної середовища як int або повертає значення за замовчуванням
func GetEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvAsBool отримує значення змінної середовища як bool або повертає значення за замовчуванням
func GetEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetEnvAsFloat64 отримує значення змінної середовища як float64 або повертає значення за замовчуванням
func GetEnvAsFloat64(key string, defaultValue float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// GetEnvAsSlice отримує значення змінної середовища як slice або повертає значення за замовчуванням
func GetEnvAsSlice(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// LoadEnvFile завантажує змінні середовища з .env файлу
func LoadEnvFile(filename string) error {
	if filename == "" {
		filename = ".env"
	}

	// Перевіряємо, чи існує файл
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("env file %s does not exist", filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open env file %s: %w", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Пропускаємо порожні рядки та коментарі
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Розділяємо на ключ та значення
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid line %d in %s: %s", lineNumber, filename, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Видаляємо лапки, якщо вони є
		if len(value) >= 2 {
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
		}

		// Встановлюємо змінну середовища, якщо вона ще не встановлена
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading env file %s: %w", filename, err)
	}

	return nil
}

// LoadEnvFiles завантажує кілька .env файлів у порядку пріоритету
func LoadEnvFiles(filenames ...string) error {
	var errors []string

	for _, filename := range filenames {
		if err := LoadEnvFile(filename); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) == len(filenames) {
		return fmt.Errorf("failed to load any env files: %s", strings.Join(errors, "; "))
	}

	return nil
}

// AutoLoadEnvFiles автоматично завантажує .env файли на основі середовища
func AutoLoadEnvFiles() error {
	environment := GetEnv("ENVIRONMENT", "development")

	// Список файлів у порядку пріоритету (останній перезаписує попередні)
	envFiles := []string{
		".env.example",
		".env",
		fmt.Sprintf(".env.%s", environment),
		".env.local",
		fmt.Sprintf(".env.%s.local", environment),
	}

	// Завантажуємо файли, які існують
	for _, filename := range envFiles {
		if _, err := os.Stat(filename); err == nil {
			if err := LoadEnvFile(filename); err != nil {
				fmt.Printf("Warning: failed to load %s: %v\n", filename, err)
			} else {
				fmt.Printf("Loaded environment file: %s\n", filename)
			}
		}
	}

	return nil
}

// LoadConfig loads configuration from environment variables (alias for LoadConfigFromEnv)
func LoadConfig() (*Config, error) {
	return LoadConfigFromEnv()
}

// LoadConfigFromEnv завантажує конфігурацію з змінних середовища
func LoadConfigFromEnv() (*Config, error) {
	config := &Config{
		Environment: GetEnv("ENVIRONMENT", "development"),
		Debug:       GetEnvAsBool("DEBUG", true),
		LogLevel:    GetEnv("LOG_LEVEL", "info"),
		LogFormat:   GetEnv("LOG_FORMAT", "json"),

		Server: ServerConfig{
			APIGatewayPort:     GetEnvAsInt("API_GATEWAY_PORT", 8080),
			ProducerPort:       GetEnvAsInt("PRODUCER_PORT", 3000),
			ConsumerPort:       GetEnvAsInt("CONSUMER_PORT", 3001),
			StreamsPort:        GetEnvAsInt("STREAMS_PORT", 3002),
			AISearchPort:       GetEnvAsInt("AI_SEARCH_PORT", 8092),
			AuthServicePort:    GetEnvAsInt("AUTH_SERVICE_PORT", 8091),
			PaymentServicePort: GetEnvAsInt("PAYMENT_SERVICE_PORT", 8093),
			OrderServicePort:   GetEnvAsInt("ORDER_SERVICE_PORT", 8094),
			KitchenServicePort: GetEnvAsInt("KITCHEN_SERVICE_PORT", 8095),
			Host:               GetEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:        GetEnv("SERVER_READ_TIMEOUT", "30s"),
			WriteTimeout:       GetEnv("SERVER_WRITE_TIMEOUT", "30s"),
			IdleTimeout:        GetEnv("SERVER_IDLE_TIMEOUT", "120s"),
		},

		Database: DatabaseConfig{
			Host:            GetEnv("DATABASE_HOST", "localhost"),
			Port:            GetEnvAsInt("DATABASE_PORT", 5432),
			Name:            GetEnv("DATABASE_NAME", "go_coffee"),
			User:            GetEnv("DATABASE_USER", "postgres"),
			Password:        GetEnv("DATABASE_PASSWORD", "postgres"),
			SSLMode:         GetEnv("DATABASE_SSL_MODE", "disable"),
			MaxOpenConns:    GetEnvAsInt("DATABASE_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    GetEnvAsInt("DATABASE_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: GetEnv("DATABASE_CONN_MAX_LIFETIME", "300s"),
			TestDBName:      GetEnv("TEST_DATABASE_NAME", "go_coffee_test"),
		},

		Redis: RedisConfig{
			Host:         GetEnv("REDIS_HOST", "localhost"),
			Port:         GetEnvAsInt("REDIS_PORT", 6379),
			Password:     GetEnv("REDIS_PASSWORD", ""),
			DB:           GetEnvAsInt("REDIS_DB", 0),
			PoolSize:     GetEnvAsInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: GetEnvAsInt("REDIS_MIN_IDLE_CONNS", 5),
			DialTimeout:  GetEnv("REDIS_DIAL_TIMEOUT", "5s"),
			ReadTimeout:  GetEnv("REDIS_READ_TIMEOUT", "3s"),
			WriteTimeout: GetEnv("REDIS_WRITE_TIMEOUT", "3s"),
			URL:          GetEnv("REDIS_URL", "redis://localhost:6379"),
		},

		Kafka: KafkaConfig{
			Brokers:        GetEnvAsSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
			Topic:          GetEnv("KAFKA_TOPIC", "coffee_orders"),
			ProcessedTopic: GetEnv("KAFKA_PROCESSED_TOPIC", "processed_orders"),
			ConsumerGroup:  GetEnv("KAFKA_CONSUMER_GROUP", "coffee-consumer-group"),
			WorkerPoolSize: GetEnvAsInt("KAFKA_WORKER_POOL_SIZE", 3),
			RetryMax:       GetEnvAsInt("KAFKA_RETRY_MAX", 5),
			RequiredAcks:   GetEnv("KAFKA_REQUIRED_ACKS", "all"),
		},

		Security: SecurityConfig{
			JWTSecret:          GetEnv("JWT_SECRET", "your-super-secret-jwt-key"),
			JWTExpiry:          GetEnv("JWT_EXPIRY", "24h"),
			RefreshTokenExpiry: GetEnv("REFRESH_TOKEN_EXPIRY", "720h"),
			JWTIssuer:          GetEnv("JWT_ISSUER", "go-coffee"),
			JWTAudience:        GetEnv("JWT_AUDIENCE", "go-coffee-users"),
			APIKeySecret:       GetEnv("API_KEY_SECRET", "your-api-key-secret"),
			WebhookSecret:      GetEnv("WEBHOOK_SECRET", "your-webhook-secret"),
			EncryptionKey:      GetEnv("ENCRYPTION_KEY", "your-32-character-encryption-key!!"),
		},

		Web3: Web3Config{
			Ethereum: EthereumConfig{
				RPCURL:     GetEnv("ETHEREUM_RPC_URL", "https://mainnet.infura.io/v3/your-project-id"),
				TestnetURL: GetEnv("ETHEREUM_TESTNET_RPC_URL", "https://goerli.infura.io/v3/your-project-id"),
				PrivateKey: GetEnv("ETHEREUM_PRIVATE_KEY", "your-ethereum-private-key"),
				GasLimit:   int64(GetEnvAsInt("ETHEREUM_GAS_LIMIT", 21000)),
				GasPrice:   int64(GetEnvAsInt("ETHEREUM_GAS_PRICE", 20000000000)),
			},
			Bitcoin: BitcoinConfig{
				RPCURL:      GetEnv("BITCOIN_RPC_URL", "https://your-bitcoin-node.com"),
				RPCUsername: GetEnv("BITCOIN_RPC_USERNAME", "your-bitcoin-rpc-username"),
				RPCPassword: GetEnv("BITCOIN_RPC_PASSWORD", "your-bitcoin-rpc-password"),
			},
			Solana: SolanaConfig{
				RPCURL:     GetEnv("SOLANA_RPC_URL", "https://api.mainnet-beta.solana.com"),
				TestnetURL: GetEnv("SOLANA_TESTNET_RPC_URL", "https://api.testnet.solana.com"),
				PrivateKey: GetEnv("SOLANA_PRIVATE_KEY", "your-solana-private-key"),
			},
			DeFi: DeFiConfig{
				UniswapV3Router:     GetEnv("UNISWAP_V3_ROUTER", "0xE592427A0AEce92De3Edee1F18E0157C05861564"),
				AaveLendingPool:     GetEnv("AAVE_LENDING_POOL", "0x7d2768dE32b0b80b7a3454c06BdAc94A69DDc7A9"),
				CompoundComptroller: GetEnv("COMPOUND_COMPTROLLER", "0x3d9819210A31b4961b30EF54bE2aeD79B9c9Cd3B"),
				OneInchAPIKey:       GetEnv("ONEINCH_API_KEY", ""),
				ChainlinkEnabled:    GetEnvAsBool("CHAINLINK_ENABLED", true),
				ArbitrageEnabled:    GetEnvAsBool("ARBITRAGE_ENABLED", true),
				YieldFarmingEnabled: GetEnvAsBool("YIELD_FARMING_ENABLED", true),
				TradingBotsEnabled:  GetEnvAsBool("TRADING_BOTS_ENABLED", false),
			},
		},

		AI: AIConfig{
			GeminiAPIKey:              GetEnv("GEMINI_API_KEY", "your-gemini-api-key"),
			OpenAIAPIKey:              GetEnv("OPENAI_API_KEY", "your-openai-api-key"),
			OllamaURL:                 GetEnv("OLLAMA_URL", "http://localhost:11434"),
			SearchEmbeddingModel:      GetEnv("AI_SEARCH_EMBEDDING_MODEL", "coffee_ai_v2"),
			SearchVectorDimensions:    GetEnvAsInt("AI_SEARCH_VECTOR_DIMENSIONS", 384),
			SearchSimilarityThreshold: GetEnvAsFloat64("AI_SEARCH_SIMILARITY_THRESHOLD", 0.7),
			SearchMaxResults:          GetEnvAsInt("AI_SEARCH_MAX_RESULTS", 50),
		},

		Integrations: IntegrationsConfig{
			SMTP: SMTPConfig{
				Host:      GetEnv("SMTP_HOST", "smtp.gmail.com"),
				Port:      GetEnvAsInt("SMTP_PORT", 587),
				Username:  GetEnv("SMTP_USERNAME", "your-email@gmail.com"),
				Password:  GetEnv("SMTP_PASSWORD", "your-app-password"),
				FromEmail: GetEnv("SMTP_FROM_EMAIL", "noreply@gocoffee.com"),
				FromName:  GetEnv("SMTP_FROM_NAME", "Go Coffee"),
			},
			Twilio: TwilioConfig{
				AccountSID: GetEnv("TWILIO_ACCOUNT_SID", "your-twilio-account-sid"),
				AuthToken:  GetEnv("TWILIO_AUTH_TOKEN", "your-twilio-auth-token"),
				FromNumber: GetEnv("TWILIO_FROM_NUMBER", "+1234567890"),
			},
			Slack: SlackConfig{
				BotToken:   GetEnv("SLACK_BOT_TOKEN", "xoxb-your-slack-bot-token"),
				WebhookURL: GetEnv("SLACK_WEBHOOK_URL", "https://hooks.slack.com/services/your/webhook/url"),
			},
			ClickUp: ClickUpConfig{
				APIToken: GetEnv("CLICKUP_API_TOKEN", "your-clickup-api-token"),
				TeamID:   GetEnv("CLICKUP_TEAM_ID", "your-clickup-team-id"),
			},
			GoogleSheets: GoogleSheetsConfig{
				CredentialsPath: GetEnv("GOOGLE_SHEETS_CREDENTIALS_PATH", "./credentials/google-sheets.json"),
				SpreadsheetID:   GetEnv("GOOGLE_SHEETS_SPREADSHEET_ID", "your-spreadsheet-id"),
			},
		},

		Monitoring: MonitoringConfig{
			Prometheus: PrometheusConfig{
				Enabled:     GetEnvAsBool("PROMETHEUS_ENABLED", true),
				Port:        GetEnvAsInt("PROMETHEUS_PORT", 9090),
				MetricsPath: GetEnv("PROMETHEUS_METRICS_PATH", "/metrics"),
			},
			Grafana: GrafanaConfig{
				Port:          GetEnvAsInt("GRAFANA_PORT", 3000),
				AdminUser:     GetEnv("GRAFANA_ADMIN_USER", "admin"),
				AdminPassword: GetEnv("GRAFANA_ADMIN_PASSWORD", "admin"),
			},
			Jaeger: JaegerConfig{
				Enabled:      GetEnvAsBool("JAEGER_ENABLED", true),
				Endpoint:     GetEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
				SamplerType:  GetEnv("JAEGER_SAMPLER_TYPE", "const"),
				SamplerParam: GetEnv("JAEGER_SAMPLER_PARAM", "1"),
			},
			Sentry: SentryConfig{
				DSN:         GetEnv("SENTRY_DSN", "your-sentry-dsn"),
				Environment: GetEnv("SENTRY_ENVIRONMENT", "development"),
				Release:     GetEnv("SENTRY_RELEASE", "1.0.0"),
			},
		},

		Features: FeatureFlags{
			ProducerServiceEnabled:      GetEnvAsBool("PRODUCER_SERVICE_ENABLED", true),
			ConsumerServiceEnabled:      GetEnvAsBool("CONSUMER_SERVICE_ENABLED", true),
			StreamsServiceEnabled:       GetEnvAsBool("STREAMS_SERVICE_ENABLED", true),
			APIGatewayEnabled:           GetEnvAsBool("API_GATEWAY_ENABLED", true),
			Web3WalletEnabled:           GetEnvAsBool("WEB3_WALLET_ENABLED", true),
			DeFiServiceEnabled:          GetEnvAsBool("DEFI_SERVICE_ENABLED", true),
			SmartContractServiceEnabled: GetEnvAsBool("SMART_CONTRACT_SERVICE_ENABLED", true),
			AISearchEnabled:             GetEnvAsBool("AI_SEARCH_ENABLED", true),
			AIAgentsEnabled:             GetEnvAsBool("AI_AGENTS_ENABLED", true),
			AuthModuleEnabled:           GetEnvAsBool("AUTH_MODULE_ENABLED", true),
			PaymentModuleEnabled:        GetEnvAsBool("PAYMENT_MODULE_ENABLED", true),
			NotificationModuleEnabled:   GetEnvAsBool("NOTIFICATION_MODULE_ENABLED", true),
		},
	}

	return config, nil
}

// ValidateConfig перевіряє конфігурацію на коректність
func ValidateConfig(config *Config) error {
	var errors []string

	// Перевіряємо обов'язкові поля
	if config.Environment == "" {
		errors = append(errors, "ENVIRONMENT is required")
	}

	if config.Database.Host == "" {
		errors = append(errors, "DATABASE_HOST is required")
	}

	if config.Database.Name == "" {
		errors = append(errors, "DATABASE_NAME is required")
	}

	if config.Redis.Host == "" {
		errors = append(errors, "REDIS_HOST is required")
	}

	if len(config.Kafka.Brokers) == 0 {
		errors = append(errors, "KAFKA_BROKERS is required")
	}

	if config.Security.JWTSecret == "" || config.Security.JWTSecret == "your-super-secret-jwt-key" {
		errors = append(errors, "JWT_SECRET must be set to a secure value")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// PrintConfig виводить конфігурацію (без секретів) для налагодження
func PrintConfig(config *Config) {
	fmt.Println("=== Go Coffee Configuration ===")
	fmt.Printf("Environment: %s\n", config.Environment)
	fmt.Printf("Debug: %t\n", config.Debug)
	fmt.Printf("Log Level: %s\n", config.LogLevel)
	fmt.Printf("Log Format: %s\n", config.LogFormat)
	fmt.Println()

	fmt.Println("Server Configuration:")
	fmt.Printf("  API Gateway Port: %d\n", config.Server.APIGatewayPort)
	fmt.Printf("  Producer Port: %d\n", config.Server.ProducerPort)
	fmt.Printf("  Consumer Port: %d\n", config.Server.ConsumerPort)
	fmt.Printf("  Streams Port: %d\n", config.Server.StreamsPort)
	fmt.Printf("  AI Search Port: %d\n", config.Server.AISearchPort)
	fmt.Printf("  Auth Service Port: %d\n", config.Server.AuthServicePort)
	fmt.Printf("  Host: %s\n", config.Server.Host)
	fmt.Println()

	fmt.Println("Database Configuration:")
	fmt.Printf("  Host: %s\n", config.Database.Host)
	fmt.Printf("  Port: %d\n", config.Database.Port)
	fmt.Printf("  Name: %s\n", config.Database.Name)
	fmt.Printf("  User: %s\n", config.Database.User)
	fmt.Printf("  SSL Mode: %s\n", config.Database.SSLMode)
	fmt.Printf("  Max Open Conns: %d\n", config.Database.MaxOpenConns)
	fmt.Printf("  Max Idle Conns: %d\n", config.Database.MaxIdleConns)
	fmt.Println()

	fmt.Println("Redis Configuration:")
	fmt.Printf("  Host: %s\n", config.Redis.Host)
	fmt.Printf("  Port: %d\n", config.Redis.Port)
	fmt.Printf("  DB: %d\n", config.Redis.DB)
	fmt.Printf("  Pool Size: %d\n", config.Redis.PoolSize)
	fmt.Println()

	fmt.Println("Kafka Configuration:")
	fmt.Printf("  Brokers: %v\n", config.Kafka.Brokers)
	fmt.Printf("  Topic: %s\n", config.Kafka.Topic)
	fmt.Printf("  Consumer Group: %s\n", config.Kafka.ConsumerGroup)
	fmt.Printf("  Worker Pool Size: %d\n", config.Kafka.WorkerPoolSize)
	fmt.Println()

	fmt.Println("Feature Flags:")
	fmt.Printf("  Producer Service: %t\n", config.Features.ProducerServiceEnabled)
	fmt.Printf("  Consumer Service: %t\n", config.Features.ConsumerServiceEnabled)
	fmt.Printf("  Streams Service: %t\n", config.Features.StreamsServiceEnabled)
	fmt.Printf("  API Gateway: %t\n", config.Features.APIGatewayEnabled)
	fmt.Printf("  Web3 Wallet: %t\n", config.Features.Web3WalletEnabled)
	fmt.Printf("  DeFi Service: %t\n", config.Features.DeFiServiceEnabled)
	fmt.Printf("  AI Search: %t\n", config.Features.AISearchEnabled)
	fmt.Printf("  AI Agents: %t\n", config.Features.AIAgentsEnabled)
	fmt.Println()

	fmt.Println("=== End Configuration ===")
}
