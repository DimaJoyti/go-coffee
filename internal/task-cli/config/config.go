package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Redis    RedisConfig    `mapstructure:"redis"`
	CLI      CLIConfig      `mapstructure:"cli"`
	Defaults DefaultsConfig `mapstructure:"defaults"`
}

// RedisConfig represents Redis connection configuration
type RedisConfig struct {
	URL      string `mapstructure:"url"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// CLIConfig represents CLI-specific configuration
type CLIConfig struct {
	DefaultUser   string `mapstructure:"default_user"`
	DateFormat    string `mapstructure:"date_format"`
	OutputFormat  string `mapstructure:"output_format"`
	ColorOutput   bool   `mapstructure:"color_output"`
	PageSize      int    `mapstructure:"page_size"`
	SortBy        string `mapstructure:"sort_by"`
	SortOrder     string `mapstructure:"sort_order"`
	ConfigDir     string `mapstructure:"config_dir"`
	HistorySize   int    `mapstructure:"history_size"`
}

// DefaultsConfig represents default values for task creation
type DefaultsConfig struct {
	Priority string   `mapstructure:"priority"`
	Status   string   `mapstructure:"status"`
	Tags     []string `mapstructure:"tags"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	// Set default values
	setDefaults()

	// Set config file name and paths
	viper.SetConfigName("task-cli")
	viper.SetConfigType("yaml")

	// Add config paths
	configDir := getConfigDir()
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/task-cli")
	viper.AddConfigPath("/etc/task-cli")

	// Set environment variable prefix
	viper.SetEnvPrefix("TASK_CLI")
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, use defaults
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Override with environment variables if set
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		config.Redis.URL = redisURL
	}
	if defaultUser := os.Getenv("TASK_CLI_DEFAULT_USER"); defaultUser != "" {
		config.CLI.DefaultUser = defaultUser
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate Redis configuration
	if c.Redis.URL == "" && c.Redis.Host == "" {
		return fmt.Errorf("redis URL or host must be specified")
	}

	// Validate CLI configuration
	if c.CLI.PageSize <= 0 {
		c.CLI.PageSize = 20
	}
	if c.CLI.HistorySize <= 0 {
		c.CLI.HistorySize = 100
	}

	// Validate output format
	validFormats := []string{"table", "json", "yaml", "csv"}
	isValidFormat := false
	for _, format := range validFormats {
		if c.CLI.OutputFormat == format {
			isValidFormat = true
			break
		}
	}
	if !isValidFormat {
		c.CLI.OutputFormat = "table"
	}

	// Validate sort order
	if c.CLI.SortOrder != "asc" && c.CLI.SortOrder != "desc" {
		c.CLI.SortOrder = "desc"
	}

	return nil
}

// GetRedisAddr returns the Redis connection address
func (c *Config) GetRedisAddr() string {
	if c.Redis.URL != "" {
		return c.Redis.URL
	}
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// setDefaults sets default configuration values
func setDefaults() {
	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// CLI defaults
	viper.SetDefault("cli.default_user", getCurrentUser())
	viper.SetDefault("cli.date_format", "2006-01-02 15:04")
	viper.SetDefault("cli.output_format", "table")
	viper.SetDefault("cli.color_output", true)
	viper.SetDefault("cli.page_size", 20)
	viper.SetDefault("cli.sort_by", "created_at")
	viper.SetDefault("cli.sort_order", "desc")
	viper.SetDefault("cli.history_size", 100)

	// Defaults for task creation
	viper.SetDefault("defaults.priority", "medium")
	viper.SetDefault("defaults.status", "pending")
	viper.SetDefault("defaults.tags", []string{})
}

// getCurrentUser returns the current system user
func getCurrentUser() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}
	return "unknown"
}

// getConfigDir returns the configuration directory
func getConfigDir() string {
	if configDir := os.Getenv("TASK_CLI_CONFIG_DIR"); configDir != "" {
		return configDir
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "."
	}

	return filepath.Join(homeDir, ".config", "task-cli")
}

// CreateDefaultConfig creates a default configuration file
func CreateDefaultConfig() error {
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "task-cli.yaml")
	if _, err := os.Stat(configFile); err == nil {
		return fmt.Errorf("config file already exists: %s", configFile)
	}

	defaultConfig := `# Task CLI Configuration
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
  # url: "redis://localhost:6379"  # Alternative to host/port

cli:
  default_user: "` + getCurrentUser() + `"
  date_format: "2006-01-02 15:04"
  output_format: "table"  # table, json, yaml, csv
  color_output: true
  page_size: 20
  sort_by: "created_at"   # created_at, updated_at, due_date, priority, status
  sort_order: "desc"      # asc, desc
  history_size: 100

defaults:
  priority: "medium"      # low, medium, high, critical
  status: "pending"       # pending, in-progress, completed, cancelled, on-hold
  tags: []
`

	if err := os.WriteFile(configFile, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Default configuration created at: %s\n", configFile)
	return nil
}

// SaveConfig saves the current configuration to file
func (c *Config) SaveConfig() error {
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "task-cli.yaml")
	viper.SetConfigFile(configFile)

	return viper.WriteConfig()
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() string {
	configDir := getConfigDir()
	return filepath.Join(configDir, "task-cli.yaml")
}
