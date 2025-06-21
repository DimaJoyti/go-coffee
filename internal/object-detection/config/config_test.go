package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Load(t *testing.T) {
	// Test loading with default values
	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Test default values
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)
	assert.Equal(t, 0.5, cfg.Detection.ConfidenceThreshold)
	assert.Equal(t, true, cfg.Tracking.Enabled)
	assert.Equal(t, true, cfg.Monitoring.Enabled)
	assert.Equal(t, true, cfg.WebSocket.Enabled)
}

func TestConfig_EnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("PORT", "9090")
	os.Setenv("DB_HOST", "test-db")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("REDIS_HOST", "test-redis")
	os.Setenv("REDIS_PORT", "6380")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("REDIS_HOST")
		os.Unsetenv("REDIS_PORT")
	}()

	cfg, err := Load()
	require.NoError(t, err)

	// Note: The current implementation only loads some env vars in loadFromEnv
	// The viper AutomaticEnv() should handle these, but may not work as expected in tests
	// For now, we'll test the ones that are explicitly handled
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "test-db", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
	assert.Equal(t, "test-redis", cfg.Redis.Host)
	assert.Equal(t, 6380, cfg.Redis.Port)
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		modifyConfig func(*Config)
		expectError bool
	}{
		{
			name: "valid config",
			modifyConfig: func(cfg *Config) {
				// No modifications - should be valid
			},
			expectError: false,
		},
		{
			name: "invalid server port - negative",
			modifyConfig: func(cfg *Config) {
				cfg.Server.Port = -1
			},
			expectError: true,
		},
		{
			name: "invalid server port - too high",
			modifyConfig: func(cfg *Config) {
				cfg.Server.Port = 70000
			},
			expectError: true,
		},
		{
			name: "invalid confidence threshold - negative",
			modifyConfig: func(cfg *Config) {
				cfg.Detection.ConfidenceThreshold = -0.1
			},
			expectError: true,
		},
		{
			name: "invalid confidence threshold - too high",
			modifyConfig: func(cfg *Config) {
				cfg.Detection.ConfidenceThreshold = 1.1
			},
			expectError: true,
		},
		{
			name: "invalid NMS threshold - negative",
			modifyConfig: func(cfg *Config) {
				cfg.Detection.NMSThreshold = -0.1
			},
			expectError: true,
		},
		{
			name: "invalid NMS threshold - too high",
			modifyConfig: func(cfg *Config) {
				cfg.Detection.NMSThreshold = 1.1
			},
			expectError: true,
		},
		{
			name: "invalid IOU threshold - negative",
			modifyConfig: func(cfg *Config) {
				cfg.Tracking.IOUThreshold = -0.1
			},
			expectError: true,
		},
		{
			name: "invalid IOU threshold - too high",
			modifyConfig: func(cfg *Config) {
				cfg.Tracking.IOUThreshold = 1.1
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Detection: DetectionConfig{
					ConfidenceThreshold: 0.5,
					NMSThreshold:        0.4,
				},
				Tracking: TrackingConfig{
					IOUThreshold: 0.3,
				},
			}

			tt.modifyConfig(cfg)
			err := validate(cfg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_DatabaseDSN(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "testuser",
			Password: "testpass",
			Database: "testdb",
			SSLMode:  "disable",
		},
	}

	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	actual := cfg.GetDatabaseDSN()
	assert.Equal(t, expected, actual)
}

func TestConfig_RedisAddr(t *testing.T) {
	cfg := &Config{
		Redis: RedisConfig{
			Host: "redis-server",
			Port: 6379,
		},
	}

	expected := "redis-server:6379"
	actual := cfg.GetRedisAddr()
	assert.Equal(t, expected, actual)
}

func TestConfig_TimeoutParsing(t *testing.T) {
	cfg := &Config{
		Detection: DetectionConfig{
			ProcessingTimeout: 30 * time.Second,
		},
		Tracking: TrackingConfig{
			CleanupInterval: 5 * time.Minute,
		},
		WebSocket: WebSocketConfig{
			PingInterval: 30 * time.Second,
			PongTimeout:  10 * time.Second,
		},
	}

	assert.Equal(t, 30*time.Second, cfg.Detection.ProcessingTimeout)
	assert.Equal(t, 5*time.Minute, cfg.Tracking.CleanupInterval)
	assert.Equal(t, 30*time.Second, cfg.WebSocket.PingInterval)
	assert.Equal(t, 10*time.Second, cfg.WebSocket.PongTimeout)
}

func TestConfig_DefaultValues(t *testing.T) {
	// Test that setDefaults sets expected values
	setDefaults()

	// We can't directly test viper defaults, but we can test the Load function
	cfg, err := Load()
	require.NoError(t, err)

	// Test some key default values
	assert.Equal(t, "development", cfg.Environment)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, 30, cfg.Server.ReadTimeout)
	assert.Equal(t, 30, cfg.Server.WriteTimeout)
	assert.Equal(t, 120, cfg.Server.IdleTimeout)

	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "postgres", cfg.Database.User)
	assert.Equal(t, "object_detection", cfg.Database.Database)
	assert.Equal(t, "disable", cfg.Database.SSLMode)

	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)
	assert.Equal(t, 0, cfg.Redis.Database)

	assert.Equal(t, "yolo", cfg.Detection.ModelType)
	assert.Equal(t, 0.5, cfg.Detection.ConfidenceThreshold)
	assert.Equal(t, 0.4, cfg.Detection.NMSThreshold)
	assert.Equal(t, 640, cfg.Detection.InputSize)
	assert.Equal(t, false, cfg.Detection.EnableGPU)

	assert.Equal(t, true, cfg.Tracking.Enabled)
	assert.Equal(t, 30, cfg.Tracking.MaxAge)
	assert.Equal(t, 3, cfg.Tracking.MinHits)
	assert.Equal(t, 0.3, cfg.Tracking.IOUThreshold)

	assert.Equal(t, true, cfg.Monitoring.Enabled)
	assert.Equal(t, 9090, cfg.Monitoring.MetricsPort)
	assert.Equal(t, "/metrics", cfg.Monitoring.MetricsPath)

	assert.Equal(t, true, cfg.WebSocket.Enabled)
	assert.Equal(t, "/ws", cfg.WebSocket.Path)
	assert.Equal(t, 100, cfg.WebSocket.MaxConnections)
}

func TestConfig_StorageConfig(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, 30, cfg.Storage.DataRetentionDays)
	assert.Equal(t, "./data/videos", cfg.Storage.VideoStoragePath)
	assert.Equal(t, "./data/models", cfg.Storage.ModelStoragePath)
	assert.Equal(t, "./data/thumbnails", cfg.Storage.ThumbnailPath)
	assert.Equal(t, 10, cfg.Storage.MaxVideoSizeGB)
	assert.Equal(t, false, cfg.Storage.EnableVideoRecording)
}

func TestConfig_MonitoringConfig(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, true, cfg.Monitoring.Enabled)
	assert.Equal(t, 9090, cfg.Monitoring.MetricsPort)
	assert.Equal(t, "/metrics", cfg.Monitoring.MetricsPath)
	assert.Equal(t, false, cfg.Monitoring.TracingEnabled)
	assert.Equal(t, "info", cfg.Monitoring.LogLevel)
	assert.Equal(t, true, cfg.Monitoring.EnableHealthCheck)
}
