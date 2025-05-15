package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Kafka    KafkaConfig    `json:"kafka"`
	Logging  LoggingConfig  `json:"logging"`
	Metrics  MetricsConfig  `json:"metrics"`
}

// ServerConfig represents the HTTP server configuration
type ServerConfig struct {
	Port int `json:"port"`
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

// KafkaConfig represents the Kafka configuration
type KafkaConfig struct {
	Brokers      []string `json:"brokers"`
	Topic        string   `json:"topic"`
	RetryMax     int      `json:"retry_max"`
	RequiredAcks string   `json:"required_acks"`
}

// LoggingConfig represents the logging configuration
type LoggingConfig struct {
	Level       string `json:"level"`
	Development bool   `json:"development"`
	Encoding    string `json:"encoding"`
}

// MetricsConfig represents the metrics configuration
type MetricsConfig struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
}

// LoadConfig loads the configuration from environment variables or a config file
func LoadConfig() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port: getEnvAsInt("SERVER_PORT", 4000),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "coffee_accounts"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Kafka: KafkaConfig{
			Brokers:      getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
			Topic:        getEnv("KAFKA_TOPIC", "account_events"),
			RetryMax:     getEnvAsInt("KAFKA_RETRY_MAX", 5),
			RequiredAcks: getEnv("KAFKA_REQUIRED_ACKS", "all"),
		},
		Logging: LoggingConfig{
			Level:       getEnv("LOG_LEVEL", "info"),
			Development: getEnvAsBool("LOG_DEVELOPMENT", false),
			Encoding:    getEnv("LOG_ENCODING", "json"),
		},
		Metrics: MetricsConfig{
			Enabled: getEnvAsBool("METRICS_ENABLED", true),
			Port:    getEnvAsInt("METRICS_PORT", 9090),
		},
	}

	// Check if config file exists and load it
	if configFile := getEnv("CONFIG_FILE", ""); configFile != "" {
		if err := loadConfigFromFile(configFile, config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// loadConfigFromFile loads the configuration from a file
func loadConfigFromFile(filename string, config *Config) error {
	file, err := os.Open(filename)
	if err != nil {
		// If file doesn't exist, just use the default config
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsSlice gets an environment variable as a slice or returns a default value
func getEnvAsSlice(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		// Try to parse as JSON array
		var result []string
		if err := json.Unmarshal([]byte(value), &result); err == nil {
			return result
		}
		// Fall back to comma-separated list
		return strings.Split(value, ",")
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as a boolean or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
