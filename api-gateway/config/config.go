package config

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
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
	ProducerAddress string `json:"producer_address"`
	ConsumerAddress string `json:"consumer_address"`
}

// LoadConfig завантажує конфігурацію з змінних середовища або файлу конфігурації
func LoadConfig() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port: getEnvAsInt("SERVER_PORT", 8080),
		},
		GRPC: GRPCConfig{
			ProducerAddress: getEnv("PRODUCER_GRPC_ADDRESS", "localhost:50051"),
			ConsumerAddress: getEnv("CONSUMER_GRPC_ADDRESS", "localhost:50052"),
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

// getEnvAsSlice отримує значення змінної середовища як []string або значення за замовчуванням
func getEnvAsSlice(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		// Видаляємо квадратні дужки та пробіли
		value = strings.TrimSpace(value)
		value = strings.Trim(value, "[]")

		// Розділяємо рядок за комами
		parts := strings.Split(value, ",")

		// Видаляємо пробіли та лапки з кожної частини
		var result []string
		for _, part := range parts {
			part = strings.TrimSpace(part)
			part = strings.Trim(part, "\"'")
			if part != "" {
				result = append(result, part)
			}
		}

		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}
