package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

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

// GetEnvAsSlice отримує значення змінної середовища як слайс або повертає значення за замовчуванням
func GetEnvAsSlice(key string, defaultValue []string, separator string) []string {
	if value, exists := os.LookupEnv(key); exists {
		if value == "" {
			return defaultValue
		}
		return strings.Split(value, separator)
	}
	return defaultValue
}

// GetEnvAsJSON отримує значення змінної середовища як JSON або повертає значення за замовчуванням
func GetEnvAsJSON(key string, defaultValue interface{}) interface{} {
	if value, exists := os.LookupEnv(key); exists {
		var result interface{}
		if err := json.Unmarshal([]byte(value), &result); err == nil {
			return result
		}
	}
	return defaultValue
}
