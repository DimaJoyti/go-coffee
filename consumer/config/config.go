package config

import (
	"encoding/json"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Server ServerConfig `json:"server"`
	Kafka  KafkaConfig  `json:"kafka"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	HealthPort int `json:"health_port"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers        []string `json:"brokers"`
	Topic          string   `json:"topic"`
	ProcessedTopic string   `json:"processed_topic"`
	ConsumerGroup  string   `json:"consumer_group"`
	WorkerPoolSize int      `json:"worker_pool_size"`
}

// LoadConfig loads configuration from environment variables or a config file
func LoadConfig() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			HealthPort: getEnvAsInt("HEALTH_PORT", 8081),
		},
		Kafka: KafkaConfig{
			Brokers:        getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
			Topic:          getEnv("KAFKA_TOPIC", "coffee_orders"),
			ProcessedTopic: getEnv("KAFKA_PROCESSED_TOPIC", "processed_orders"),
			ConsumerGroup:  getEnv("KAFKA_CONSUMER_GROUP", "coffee-consumer-group"),
			WorkerPoolSize: getEnvAsInt("KAFKA_WORKER_POOL_SIZE", 3),
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

// Helper functions to get environment variables
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		var result []string
		if err := json.Unmarshal([]byte(value), &result); err == nil {
			return result
		}
	}
	return defaultValue
}

func loadConfigFromFile(file string, config *Config) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	return decoder.Decode(config)
}
