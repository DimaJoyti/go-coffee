package config

import (
	"encoding/json"
	"os"
	"strconv"
	"time"
)

// Config структура для конфігурації API Gateway
type Config struct {
	Server ServerConfig `json:"server"`
	GRPC   GRPCConfig   `json:"grpc"`
}

// ServerConfig структура для конфігурації HTTP сервера
type ServerConfig struct {
	Port int `json:"port"`
}

// GRPCConfig структура для конфігурації gRPC клієнтів
type GRPCConfig struct {
	ProducerAddress    string        `json:"producer_address"`
	ConsumerAddress    string        `json:"consumer_address"`
	ConnectionTimeout  time.Duration `json:"connection_timeout"`
	MaxRetries         int           `json:"max_retries"`
	RetryDelay         time.Duration `json:"retry_delay"`
}

// LoadConfig завантажує конфігурацію з змінних середовища або файлу конфігурації
func LoadConfig() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port: getEnvAsInt("SERVER_PORT", 8080),
		},
		GRPC: GRPCConfig{
			ProducerAddress:   getEnv("PRODUCER_GRPC_ADDRESS", "localhost:50051"),
			ConsumerAddress:   getEnv("CONSUMER_GRPC_ADDRESS", "localhost:50052"),
			ConnectionTimeout: getEnvAsDuration("GRPC_CONNECTION_TIMEOUT", 5*time.Second),
			MaxRetries:        getEnvAsInt("GRPC_MAX_RETRIES", 3),
			RetryDelay:        getEnvAsDuration("GRPC_RETRY_DELAY", 1*time.Second),
		},
	}

	// Перевірка наявності файлу конфігурації
	if configFile := getEnv("CONFIG_FILE", "config.json"); configFile != "" {
		if err := loadConfigFromFile(configFile, config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// loadConfigFromFile завантажує конфігурацію з файлу
func loadConfigFromFile(filename string, config *Config) error {
	file, err := os.Open(filename)
	if err != nil {
		// Якщо файл не існує, просто повертаємо конфігурацію за замовчуванням
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(config)
}

// getEnv отримує значення змінної середовища або значення за замовчуванням
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt отримує значення змінної середовища як int або значення за замовчуванням
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsDuration отримує значення змінної середовища як time.Duration або значення за замовчуванням
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}


