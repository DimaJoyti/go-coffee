package config

import (
	"encoding/json"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Kafka KafkaConfig `json:"kafka"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers            []string `json:"brokers"`
	InputTopic         string   `json:"input_topic"`
	OutputTopic        string   `json:"output_topic"`
	ApplicationID      string   `json:"application_id"`
	AutoOffsetReset    string   `json:"auto_offset_reset"`
	ProcessingGuarantee string  `json:"processing_guarantee"`
}

// LoadConfig loads configuration from environment variables or a config file
func LoadConfig() (*Config, error) {
	config := &Config{
		Kafka: KafkaConfig{
			Brokers:            getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
			InputTopic:         getEnv("KAFKA_INPUT_TOPIC", "coffee_orders"),
			OutputTopic:        getEnv("KAFKA_OUTPUT_TOPIC", "processed_orders"),
			ApplicationID:      getEnv("KAFKA_APPLICATION_ID", "coffee-streams-app"),
			AutoOffsetReset:    getEnv("KAFKA_AUTO_OFFSET_RESET", "earliest"),
			ProcessingGuarantee: getEnv("KAFKA_PROCESSING_GUARANTEE", "at_least_once"),
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
