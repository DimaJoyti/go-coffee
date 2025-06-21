package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Environment string         `mapstructure:"environment" yaml:"environment"`
	Server      ServerConfig   `mapstructure:"server" yaml:"server"`
	Database    DatabaseConfig `mapstructure:"database" yaml:"database"`
	Redis       RedisConfig    `mapstructure:"redis" yaml:"redis"`
	Detection   DetectionConfig `mapstructure:"detection" yaml:"detection"`
	Tracking    TrackingConfig `mapstructure:"tracking" yaml:"tracking"`
	Storage     StorageConfig  `mapstructure:"storage" yaml:"storage"`
	Monitoring  MonitoringConfig `mapstructure:"monitoring" yaml:"monitoring"`
	WebSocket   WebSocketConfig `mapstructure:"websocket" yaml:"websocket"`
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Port         int `mapstructure:"port" yaml:"port"`
	ReadTimeout  int `mapstructure:"read_timeout" yaml:"read_timeout"`
	WriteTimeout int `mapstructure:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  int `mapstructure:"idle_timeout" yaml:"idle_timeout"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string `mapstructure:"host" yaml:"host"`
	Port            int    `mapstructure:"port" yaml:"port"`
	User            string `mapstructure:"user" yaml:"user"`
	Password        string `mapstructure:"password" yaml:"password"`
	Database        string `mapstructure:"database" yaml:"database"`
	SSLMode         string `mapstructure:"ssl_mode" yaml:"ssl_mode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host         string `mapstructure:"host" yaml:"host"`
	Port         int    `mapstructure:"port" yaml:"port"`
	Password     string `mapstructure:"password" yaml:"password"`
	Database     int    `mapstructure:"database" yaml:"database"`
	PoolSize     int    `mapstructure:"pool_size" yaml:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns" yaml:"min_idle_conns"`
}

// DetectionConfig represents object detection configuration
type DetectionConfig struct {
	ModelPath           string        `mapstructure:"model_path" yaml:"model_path"`
	ModelType           string        `mapstructure:"model_type" yaml:"model_type"`
	ConfidenceThreshold float64       `mapstructure:"confidence_threshold" yaml:"confidence_threshold"`
	NMSThreshold        float64       `mapstructure:"nms_threshold" yaml:"nms_threshold"`
	InputSize           int           `mapstructure:"input_size" yaml:"input_size"`
	MaxDetections       int           `mapstructure:"max_detections" yaml:"max_detections"`
	ProcessingTimeout   time.Duration `mapstructure:"processing_timeout" yaml:"processing_timeout"`
	EnableGPU           bool          `mapstructure:"enable_gpu" yaml:"enable_gpu"`
	BatchSize           int           `mapstructure:"batch_size" yaml:"batch_size"`
}

// TrackingConfig represents object tracking configuration
type TrackingConfig struct {
	Enabled             bool          `mapstructure:"enabled" yaml:"enabled"`
	MaxAge              int           `mapstructure:"max_age" yaml:"max_age"`
	MinHits             int           `mapstructure:"min_hits" yaml:"min_hits"`
	IOUThreshold        float64       `mapstructure:"iou_threshold" yaml:"iou_threshold"`
	MaxDistance         float64       `mapstructure:"max_distance" yaml:"max_distance"`
	CleanupInterval     time.Duration `mapstructure:"cleanup_interval" yaml:"cleanup_interval"`
	TrajectoryMaxLength int           `mapstructure:"trajectory_max_length" yaml:"trajectory_max_length"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	DataRetentionDays   int    `mapstructure:"data_retention_days" yaml:"data_retention_days"`
	VideoStoragePath    string `mapstructure:"video_storage_path" yaml:"video_storage_path"`
	ModelStoragePath    string `mapstructure:"model_storage_path" yaml:"model_storage_path"`
	ThumbnailPath       string `mapstructure:"thumbnail_path" yaml:"thumbnail_path"`
	MaxVideoSizeGB      int    `mapstructure:"max_video_size_gb" yaml:"max_video_size_gb"`
	EnableVideoRecording bool   `mapstructure:"enable_video_recording" yaml:"enable_video_recording"`
}

// MonitoringConfig represents monitoring and observability configuration
type MonitoringConfig struct {
	Enabled           bool   `mapstructure:"enabled" yaml:"enabled"`
	MetricsPort       int    `mapstructure:"metrics_port" yaml:"metrics_port"`
	MetricsPath       string `mapstructure:"metrics_path" yaml:"metrics_path"`
	TracingEnabled    bool   `mapstructure:"tracing_enabled" yaml:"tracing_enabled"`
	TracingEndpoint   string `mapstructure:"tracing_endpoint" yaml:"tracing_endpoint"`
	LogLevel          string `mapstructure:"log_level" yaml:"log_level"`
	EnableHealthCheck bool   `mapstructure:"enable_health_check" yaml:"enable_health_check"`
}

// WebSocketConfig represents WebSocket configuration
type WebSocketConfig struct {
	Enabled         bool          `mapstructure:"enabled" yaml:"enabled"`
	Path            string        `mapstructure:"path" yaml:"path"`
	MaxConnections  int           `mapstructure:"max_connections" yaml:"max_connections"`
	ReadBufferSize  int           `mapstructure:"read_buffer_size" yaml:"read_buffer_size"`
	WriteBufferSize int           `mapstructure:"write_buffer_size" yaml:"write_buffer_size"`
	PingInterval    time.Duration `mapstructure:"ping_interval" yaml:"ping_interval"`
	PongTimeout     time.Duration `mapstructure:"pong_timeout" yaml:"pong_timeout"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	config := &Config{}

	// Set default values
	setDefaults()

	// Configure viper
	viper.SetConfigName("object-detection")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Enable environment variable support
	viper.AutomaticEnv()
	viper.SetEnvPrefix("OD")

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal config
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override with environment variables (for variables not handled by viper)
	loadFromEnv(config)

	// Validate configuration
	if err := validate(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.idle_timeout", 120)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.database", "object_detection")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", 300)

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.database", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 2)

	// Detection defaults
	viper.SetDefault("detection.model_type", "yolo")
	viper.SetDefault("detection.confidence_threshold", 0.5)
	viper.SetDefault("detection.nms_threshold", 0.4)
	viper.SetDefault("detection.input_size", 640)
	viper.SetDefault("detection.max_detections", 100)
	viper.SetDefault("detection.processing_timeout", "30s")
	viper.SetDefault("detection.enable_gpu", false)
	viper.SetDefault("detection.batch_size", 1)

	// Tracking defaults
	viper.SetDefault("tracking.enabled", true)
	viper.SetDefault("tracking.max_age", 30)
	viper.SetDefault("tracking.min_hits", 3)
	viper.SetDefault("tracking.iou_threshold", 0.3)
	viper.SetDefault("tracking.max_distance", 100.0)
	viper.SetDefault("tracking.cleanup_interval", "5m")
	viper.SetDefault("tracking.trajectory_max_length", 100)

	// Storage defaults
	viper.SetDefault("storage.data_retention_days", 30)
	viper.SetDefault("storage.video_storage_path", "./data/videos")
	viper.SetDefault("storage.model_storage_path", "./data/models")
	viper.SetDefault("storage.thumbnail_path", "./data/thumbnails")
	viper.SetDefault("storage.max_video_size_gb", 10)
	viper.SetDefault("storage.enable_video_recording", false)

	// Monitoring defaults
	viper.SetDefault("monitoring.enabled", true)
	viper.SetDefault("monitoring.metrics_port", 9090)
	viper.SetDefault("monitoring.metrics_path", "/metrics")
	viper.SetDefault("monitoring.tracing_enabled", false)
	viper.SetDefault("monitoring.log_level", "info")
	viper.SetDefault("monitoring.enable_health_check", true)

	// WebSocket defaults
	viper.SetDefault("websocket.enabled", true)
	viper.SetDefault("websocket.path", "/ws")
	viper.SetDefault("websocket.max_connections", 100)
	viper.SetDefault("websocket.read_buffer_size", 1024)
	viper.SetDefault("websocket.write_buffer_size", 1024)
	viper.SetDefault("websocket.ping_interval", "30s")
	viper.SetDefault("websocket.pong_timeout", "10s")

	// Environment default
	viper.SetDefault("environment", "development")
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Environment = env
	}

	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}

	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		if p, err := strconv.Atoi(dbPort); err == nil {
			config.Database.Port = p
		}
	}

	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		config.Database.User = dbUser
	}

	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		config.Database.Password = dbPassword
	}

	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		config.Database.Database = dbName
	}

	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		config.Redis.Host = redisHost
	}

	if redisPort := os.Getenv("REDIS_PORT"); redisPort != "" {
		if p, err := strconv.Atoi(redisPort); err == nil {
			config.Redis.Port = p
		}
	}

	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		config.Redis.Password = redisPassword
	}
}

// validate validates the configuration
func validate(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Detection.ConfidenceThreshold < 0 || config.Detection.ConfidenceThreshold > 1 {
		return fmt.Errorf("invalid confidence threshold: %f", config.Detection.ConfidenceThreshold)
	}

	if config.Detection.NMSThreshold < 0 || config.Detection.NMSThreshold > 1 {
		return fmt.Errorf("invalid NMS threshold: %f", config.Detection.NMSThreshold)
	}

	if config.Tracking.IOUThreshold < 0 || config.Tracking.IOUThreshold > 1 {
		return fmt.Errorf("invalid IOU threshold: %f", config.Tracking.IOUThreshold)
	}

	return nil
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User,
		c.Database.Password, c.Database.Database, c.Database.SSLMode)
}

// GetRedisAddr returns the Redis connection address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}
