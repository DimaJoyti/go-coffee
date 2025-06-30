package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config holds all configuration for the beverage inventor agent
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	AI       AIConfig       `yaml:"ai"`
	Database DatabaseConfig `yaml:"database"`
	External ExternalConfig `yaml:"external"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         int    `yaml:"port"`
	Host         string `yaml:"host"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers     []string          `yaml:"brokers"`
	Topics      KafkaTopics       `yaml:"topics"`
	Consumer    KafkaConsumer     `yaml:"consumer"`
	Producer    KafkaProducer     `yaml:"producer"`
	Security    KafkaSecurity     `yaml:"security"`
}

// KafkaTopics defines all Kafka topics used by the agent
type KafkaTopics struct {
	BeverageCreated    string `yaml:"beverage_created"`
	BeverageUpdated    string `yaml:"beverage_updated"`
	IngredientDiscovered string `yaml:"ingredient_discovered"`
	TaskCreated        string `yaml:"task_created"`
	RecipeRequests     string `yaml:"recipe_requests"`
}

// KafkaConsumer holds consumer-specific configuration
type KafkaConsumer struct {
	GroupID          string `yaml:"group_id"`
	AutoOffsetReset  string `yaml:"auto_offset_reset"`
	SessionTimeout   int    `yaml:"session_timeout"`
	HeartbeatInterval int   `yaml:"heartbeat_interval"`
}

// KafkaProducer holds producer-specific configuration
type KafkaProducer struct {
	RequiredAcks int `yaml:"required_acks"`
	RetryMax     int `yaml:"retry_max"`
	BatchTimeout int `yaml:"batch_timeout"`
}

// KafkaSecurity holds Kafka security configuration
type KafkaSecurity struct {
	Enabled   bool   `yaml:"enabled"`
	Protocol  string `yaml:"protocol"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	CertFile  string `yaml:"cert_file"`
	KeyFile   string `yaml:"key_file"`
	CAFile    string `yaml:"ca_file"`
}

// AIConfig holds AI provider configuration
type AIConfig struct {
	Providers []AIProvider `yaml:"providers"`
	Default   string       `yaml:"default"`
	Timeout   int          `yaml:"timeout"`
	RetryMax  int          `yaml:"retry_max"`
}

// AIProvider holds configuration for a specific AI provider
type AIProvider struct {
	Name     string            `yaml:"name"`
	Type     string            `yaml:"type"` // gemini, openai, ollama
	Endpoint string            `yaml:"endpoint"`
	APIKey   string            `yaml:"api_key"`
	Model    string            `yaml:"model"`
	Settings map[string]interface{} `yaml:"settings"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type     string `yaml:"type"` // postgres, sqlite, memory
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
	MaxConns int    `yaml:"max_connections"`
	MaxIdle  int    `yaml:"max_idle"`
}

// ExternalConfig holds external service configurations
type ExternalConfig struct {
	ClickUp      ClickUpConfig      `yaml:"clickup"`
	Slack        SlackConfig        `yaml:"slack"`
	GoogleSheets GoogleSheetsConfig `yaml:"google_sheets"`
	Notifications NotificationConfig `yaml:"notifications"`
}

// ClickUpConfig holds ClickUp API configuration
type ClickUpConfig struct {
	Enabled   bool   `yaml:"enabled"`
	APIKey    string `yaml:"api_key"`
	TeamID    string `yaml:"team_id"`
	SpaceID   string `yaml:"space_id"`
	FolderID  string `yaml:"folder_id"`
	ListID    string `yaml:"list_id"`
	BaseURL   string `yaml:"base_url"`
	Timeout   int    `yaml:"timeout"`
}

// SlackConfig holds Slack API configuration
type SlackConfig struct {
	Enabled     bool   `yaml:"enabled"`
	BotToken    string `yaml:"bot_token"`
	Channel     string `yaml:"channel"`
	WebhookURL  string `yaml:"webhook_url"`
	Timeout     int    `yaml:"timeout"`
}

// GoogleSheetsConfig holds Google Sheets API configuration
type GoogleSheetsConfig struct {
	Enabled         bool   `yaml:"enabled"`
	CredentialsFile string `yaml:"credentials_file"`
	SpreadsheetID   string `yaml:"spreadsheet_id"`
	SheetName       string `yaml:"sheet_name"`
}

// NotificationConfig holds notification settings
type NotificationConfig struct {
	Enabled  bool     `yaml:"enabled"`
	Channels []string `yaml:"channels"` // slack, email, webhook
	Email    EmailConfig `yaml:"email"`
	Webhook  WebhookConfig `yaml:"webhook"`
}

// EmailConfig holds email notification configuration
type EmailConfig struct {
	SMTPHost     string   `yaml:"smtp_host"`
	SMTPPort     int      `yaml:"smtp_port"`
	Username     string   `yaml:"username"`
	Password     string   `yaml:"password"`
	FromAddress  string   `yaml:"from_address"`
	ToAddresses  []string `yaml:"to_addresses"`
	UseTLS       bool     `yaml:"use_tls"`
}

// WebhookConfig holds webhook notification configuration
type WebhookConfig struct {
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	Timeout int               `yaml:"timeout"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"` // json, text
	Output     string `yaml:"output"` // stdout, file
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	config := &Config{}

	// Load from file if it exists
	if configPath != "" {
		if err := loadFromFile(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config from file: %w", err)
		}
	}

	// Override with environment variables
	loadFromEnv(config)

	// Set defaults
	setDefaults(config)

	// Validate configuration
	if err := validate(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// loadFromFile loads configuration from YAML file
func loadFromFile(config *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, config)
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
	// Server configuration
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}
	if host := os.Getenv("SERVER_HOST"); host != "" {
		config.Server.Host = host
	}

	// Kafka configuration
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		config.Kafka.Brokers = strings.Split(brokers, ",")
	}
	if groupID := os.Getenv("KAFKA_GROUP_ID"); groupID != "" {
		config.Kafka.Consumer.GroupID = groupID
	}

	// AI configuration
	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		// Find Gemini provider and set API key
		for i, provider := range config.AI.Providers {
			if provider.Type == "gemini" {
				config.AI.Providers[i].APIKey = apiKey
			}
		}
	}
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		// Find OpenAI provider and set API key
		for i, provider := range config.AI.Providers {
			if provider.Type == "openai" {
				config.AI.Providers[i].APIKey = apiKey
			}
		}
	}

	// Database configuration
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		// Parse database URL (simplified)
		config.Database.Host = dbURL
	}

	// External services
	if apiKey := os.Getenv("CLICKUP_API_KEY"); apiKey != "" {
		config.External.ClickUp.APIKey = apiKey
	}
	if token := os.Getenv("SLACK_BOT_TOKEN"); token != "" {
		config.External.Slack.BotToken = token
	}
}

// setDefaults sets default values for configuration
func setDefaults(config *Config) {
	// Server defaults
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 30
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 30
	}

	// Kafka defaults
	if len(config.Kafka.Brokers) == 0 {
		config.Kafka.Brokers = []string{"localhost:9092"}
	}
	if config.Kafka.Consumer.GroupID == "" {
		config.Kafka.Consumer.GroupID = "beverage-inventor-agent"
	}
	if config.Kafka.Consumer.AutoOffsetReset == "" {
		config.Kafka.Consumer.AutoOffsetReset = "earliest"
	}

	// Topic defaults
	if config.Kafka.Topics.BeverageCreated == "" {
		config.Kafka.Topics.BeverageCreated = "beverage.created"
	}
	if config.Kafka.Topics.BeverageUpdated == "" {
		config.Kafka.Topics.BeverageUpdated = "beverage.updated"
	}
	if config.Kafka.Topics.IngredientDiscovered == "" {
		config.Kafka.Topics.IngredientDiscovered = "ingredient.discovered"
	}
	if config.Kafka.Topics.TaskCreated == "" {
		config.Kafka.Topics.TaskCreated = "task.created"
	}
	if config.Kafka.Topics.RecipeRequests == "" {
		config.Kafka.Topics.RecipeRequests = "recipe.requests"
	}

	// AI defaults
	if config.AI.Default == "" {
		config.AI.Default = "gemini"
	}
	if config.AI.Timeout == 0 {
		config.AI.Timeout = 30
	}
	if config.AI.RetryMax == 0 {
		config.AI.RetryMax = 3
	}

	// Logging defaults
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}
	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}
}

// validate validates the configuration
func validate(config *Config) error {
	// Validate required fields
	if len(config.Kafka.Brokers) == 0 {
		return fmt.Errorf("kafka brokers are required")
	}

	if config.Kafka.Consumer.GroupID == "" {
		return fmt.Errorf("kafka consumer group ID is required")
	}

	// Validate AI providers
	if len(config.AI.Providers) == 0 {
		return fmt.Errorf("at least one AI provider must be configured")
	}

	for _, provider := range config.AI.Providers {
		if provider.Name == "" {
			return fmt.Errorf("AI provider name is required")
		}
		if provider.Type == "" {
			return fmt.Errorf("AI provider type is required")
		}
	}

	return nil
}
