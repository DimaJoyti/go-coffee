package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Loader handles configuration loading from multiple sources
type Loader struct {
	configPath string
	envPrefix  string
}

// NewLoader creates a new configuration loader
func NewLoader(configPath, envPrefix string) *Loader {
	return &Loader{
		configPath: configPath,
		envPrefix:  envPrefix,
	}
}

// Load loads configuration from file and environment variables
func (l *Loader) Load() (*Config, error) {
	config := &Config{}

	// Load from YAML file first
	if err := l.loadFromFile(config); err != nil {
		return nil, fmt.Errorf("failed to load config from file: %w", err)
	}

	// Override with environment variables
	if err := l.loadFromEnv(config); err != nil {
		return nil, fmt.Errorf("failed to load config from environment: %w", err)
	}

	// Apply defaults
	l.applyDefaults(config)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// loadFromFile loads configuration from YAML file
func (l *Loader) loadFromFile(config *Config) error {
	if l.configPath == "" {
		return nil // No config file specified
	}

	// Check if file exists
	if _, err := os.Stat(l.configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", l.configPath)
	}

	// Read file
	data, err := ioutil.ReadFile(l.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return nil
}

// loadFromEnv loads configuration from environment variables
func (l *Loader) loadFromEnv(config *Config) error {
	return l.setFieldsFromEnv(reflect.ValueOf(config).Elem(), "")
}

// setFieldsFromEnv recursively sets struct fields from environment variables
func (l *Loader) setFieldsFromEnv(v reflect.Value, prefix string) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get YAML tag for field name
		yamlTag := fieldType.Tag.Get("yaml")
		if yamlTag == "" || yamlTag == "-" {
			continue
		}

		// Remove options from YAML tag
		fieldName := strings.Split(yamlTag, ",")[0]
		
		// Build environment variable name
		envKey := l.buildEnvKey(prefix, fieldName)

		// Handle different field types
		switch field.Kind() {
		case reflect.Struct:
			// Recursively handle nested structs
			if err := l.setFieldsFromEnv(field, envKey); err != nil {
				return err
			}
		case reflect.Map:
			// Handle maps (like AI providers)
			if err := l.setMapFromEnv(field, envKey); err != nil {
				return err
			}
		case reflect.Slice:
			// Handle slices
			if err := l.setSliceFromEnv(field, envKey); err != nil {
				return err
			}
		default:
			// Handle primitive types
			if err := l.setFieldFromEnv(field, envKey); err != nil {
				return err
			}
		}
	}

	return nil
}

// buildEnvKey builds environment variable key from prefix and field name
func (l *Loader) buildEnvKey(prefix, fieldName string) string {
	key := strings.ToUpper(fieldName)
	if prefix != "" {
		key = prefix + "_" + key
	}
	if l.envPrefix != "" {
		key = l.envPrefix + "_" + key
	}
	return key
}

// setFieldFromEnv sets a single field from environment variable
func (l *Loader) setFieldFromEnv(field reflect.Value, envKey string) error {
	envValue := os.Getenv(envKey)
	if envValue == "" {
		return nil // No environment variable set
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(envValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			// Handle time.Duration
			duration, err := time.ParseDuration(envValue)
			if err != nil {
				return fmt.Errorf("invalid duration for %s: %w", envKey, err)
			}
			field.SetInt(int64(duration))
		} else {
			// Handle regular integers
			intValue, err := strconv.ParseInt(envValue, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer for %s: %w", envKey, err)
			}
			field.SetInt(intValue)
		}
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(envValue, 64)
		if err != nil {
			return fmt.Errorf("invalid float for %s: %w", envKey, err)
		}
		field.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(envValue)
		if err != nil {
			return fmt.Errorf("invalid boolean for %s: %w", envKey, err)
		}
		field.SetBool(boolValue)
	}

	return nil
}

// setSliceFromEnv sets slice fields from environment variables
func (l *Loader) setSliceFromEnv(field reflect.Value, envKey string) error {
	envValue := os.Getenv(envKey)
	if envValue == "" {
		return nil
	}

	// Split by comma
	values := strings.Split(envValue, ",")
	slice := reflect.MakeSlice(field.Type(), len(values), len(values))

	for i, value := range values {
		value = strings.TrimSpace(value)
		elem := slice.Index(i)
		
		switch elem.Kind() {
		case reflect.String:
			elem.SetString(value)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer in slice for %s: %w", envKey, err)
			}
			elem.SetInt(intValue)
		}
	}

	field.Set(slice)
	return nil
}

// setMapFromEnv sets map fields from environment variables
func (l *Loader) setMapFromEnv(field reflect.Value, envKey string) error {
	// For maps, we look for environment variables with the pattern:
	// PREFIX_MAPKEY_FIELDNAME
	
	if field.IsNil() {
		field.Set(reflect.MakeMap(field.Type()))
	}

	// Get all environment variables that start with our prefix
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		envName := parts[0]
		envValue := parts[1]

		// Check if this env var is for our map
		if !strings.HasPrefix(envName, envKey+"_") {
			continue
		}

		// Extract map key and field name
		suffix := strings.TrimPrefix(envName, envKey+"_")
		keyParts := strings.SplitN(suffix, "_", 2)
		if len(keyParts) != 2 {
			continue
		}

		mapKey := strings.ToLower(keyParts[0])
		fieldName := keyParts[1]

		// Get or create map entry
		mapValue := field.MapIndex(reflect.ValueOf(mapKey))
		if !mapValue.IsValid() {
			// Create new map entry
			mapValue = reflect.New(field.Type().Elem()).Elem()
			field.SetMapIndex(reflect.ValueOf(mapKey), mapValue)
		}

		// Set field in map entry
		if err := l.setStructFieldFromEnv(mapValue, fieldName, envValue); err != nil {
			return err
		}
	}

	return nil
}

// setStructFieldFromEnv sets a specific field in a struct from environment value
func (l *Loader) setStructFieldFromEnv(structValue reflect.Value, fieldName, envValue string) error {
	structType := structValue.Type()

	// Find field by YAML tag
	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		yamlTag := fieldType.Tag.Get("yaml")
		if yamlTag == "" {
			continue
		}

		tagName := strings.Split(yamlTag, ",")[0]
		if strings.ToUpper(tagName) == fieldName {
			return l.setFieldValue(field, envValue)
		}
	}

	return nil
}

// setFieldValue sets a field value from string
func (l *Loader) setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			field.SetInt(int64(duration))
		} else {
			intValue, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(intValue)
		}
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	}

	return nil
}

// applyDefaults applies default values to configuration
func (l *Loader) applyDefaults(config *Config) {
	// Service defaults
	if config.Service.Host == "" {
		config.Service.Host = "0.0.0.0"
	}
	if config.Service.Port == 0 {
		config.Service.Port = 8080
	}
	if config.Service.BasePath == "" {
		config.Service.BasePath = "/api/v1"
	}

	// Server defaults
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 30 * time.Second
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 30 * time.Second
	}
	if config.Server.IdleTimeout == 0 {
		config.Server.IdleTimeout = 120 * time.Second
	}
	if config.Server.ShutdownTimeout == 0 {
		config.Server.ShutdownTimeout = 30 * time.Second
	}
	if config.Server.MaxHeaderBytes == 0 {
		config.Server.MaxHeaderBytes = 1 << 20 // 1MB
	}
	if config.Server.MetricsPath == "" {
		config.Server.MetricsPath = "/metrics"
	}
	if config.Server.HealthPath == "" {
		config.Server.HealthPath = "/health"
	}

	// Database defaults
	if config.Database.Driver == "" {
		config.Database.Driver = "postgres"
	}
	if config.Database.Port == 0 {
		config.Database.Port = 5432
	}
	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "disable"
	}
	if config.Database.MaxOpenConns == 0 {
		config.Database.MaxOpenConns = 25
	}
	if config.Database.MaxIdleConns == 0 {
		config.Database.MaxIdleConns = 5
	}
	if config.Database.ConnMaxLifetime == 0 {
		config.Database.ConnMaxLifetime = 5 * time.Minute
	}

	// Kafka defaults
	if config.Kafka.ClientID == "" {
		config.Kafka.ClientID = config.Service.Name
	}
	if config.Kafka.ConnectTimeout == 0 {
		config.Kafka.ConnectTimeout = 10 * time.Second
	}
	if config.Kafka.ReadTimeout == 0 {
		config.Kafka.ReadTimeout = 10 * time.Second
	}
	if config.Kafka.WriteTimeout == 0 {
		config.Kafka.WriteTimeout = 10 * time.Second
	}

	// AI defaults
	if config.AI.RateLimits.RequestsPerMinute == 0 {
		config.AI.RateLimits.RequestsPerMinute = 60
	}
	if config.AI.RateLimits.BurstSize == 0 {
		config.AI.RateLimits.BurstSize = 10
	}

	// Security defaults
	if config.Security.JWT.Algorithm == "" {
		config.Security.JWT.Algorithm = "HS256"
	}
	if config.Security.JWT.ExpirationTime == 0 {
		config.Security.JWT.ExpirationTime = 24 * time.Hour
	}

	// Feature flags defaults
	config.Features.EnableHealthChecks = true
	config.Features.EnableGracefulShutdown = true
}

// LoadConfig is a convenience function to load configuration
func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()
	envPrefix := GetEnvOrDefault("CONFIG_ENV_PREFIX", "GOCOFFEE")
	
	loader := NewLoader(configPath, envPrefix)
	return loader.Load()
}

// LoadConfigFromPath loads configuration from a specific path
func LoadConfigFromPath(path string) (*Config, error) {
	envPrefix := GetEnvOrDefault("CONFIG_ENV_PREFIX", "GOCOFFEE")
	
	loader := NewLoader(path, envPrefix)
	return loader.Load()
}

// LoadConfigWithPrefix loads configuration with a custom environment prefix
func LoadConfigWithPrefix(envPrefix string) (*Config, error) {
	configPath := GetConfigPath()
	
	loader := NewLoader(configPath, envPrefix)
	return loader.Load()
}
