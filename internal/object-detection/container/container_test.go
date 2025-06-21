package container

import (
	"context"
	"testing"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewContainer_WithValidConfig(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "test",
			Password:        "test",
			Database:        "test",
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 300,
		},
		Redis: config.RedisConfig{
			Host:         "localhost",
			Port:         6379,
			Password:     "",
			Database:     0,
			PoolSize:     10,
			MinIdleConns: 2,
		},
	}

	// Note: This test will fail if PostgreSQL and Redis are not running
	// In a real test environment, you would use test containers or mocks
	container, err := NewContainer(cfg, logger)
	if err != nil {
		// Skip test if dependencies are not available
		t.Skipf("Skipping container test due to missing dependencies: %v", err)
		return
	}

	assert.NotNil(t, container)
	assert.Equal(t, cfg, container.Config)
	assert.Equal(t, logger, container.Logger)
	assert.NotNil(t, container.Metrics)

	// Clean up
	err = container.Close()
	assert.NoError(t, err)
}

func TestContainer_HealthCheck(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "test",
			Password:        "test",
			Database:        "test",
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 300,
		},
		Redis: config.RedisConfig{
			Host:         "localhost",
			Port:         6379,
			Password:     "",
			Database:     0,
			PoolSize:     10,
			MinIdleConns: 2,
		},
	}

	container, err := NewContainer(cfg, logger)
	if err != nil {
		t.Skipf("Skipping health check test due to missing dependencies: %v", err)
		return
	}
	defer container.Close()

	ctx := context.Background()
	status := container.HealthCheck(ctx)

	assert.NotNil(t, status)
	assert.Contains(t, status, "database")
	assert.Contains(t, status, "redis")

	// If dependencies are available, they should be healthy
	if status["database"] != "not initialized" {
		assert.Equal(t, "healthy", status["database"])
	}
	if status["redis"] != "not initialized" {
		assert.Equal(t, "healthy", status["redis"])
	}
}

func TestContainer_Close(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "test",
			Password:        "test",
			Database:        "test",
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 300,
		},
		Redis: config.RedisConfig{
			Host:         "localhost",
			Port:         6379,
			Password:     "",
			Database:     0,
			PoolSize:     10,
			MinIdleConns: 2,
		},
	}

	container, err := NewContainer(cfg, logger)
	if err != nil {
		t.Skipf("Skipping close test due to missing dependencies: %v", err)
		return
	}

	// Close should not return an error
	err = container.Close()
	assert.NoError(t, err)

	// Calling close again should not panic or error
	err = container.Close()
	assert.NoError(t, err)
}

func TestContainer_DatabaseDSN(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "testhost",
			Port:     5433,
			User:     "testuser",
			Password: "testpass",
			Database: "testdb",
			SSLMode:  "require",
		},
	}

	expected := "host=testhost port=5433 user=testuser password=testpass dbname=testdb sslmode=require"
	actual := cfg.GetDatabaseDSN()
	assert.Equal(t, expected, actual)
}

func TestContainer_RedisAddr(t *testing.T) {
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Host: "redis-test",
			Port: 6380,
		},
	}

	expected := "redis-test:6380"
	actual := cfg.GetRedisAddr()
	assert.Equal(t, expected, actual)
}

func TestContainer_WithoutDependencies(t *testing.T) {
	logger := zap.NewNop()
	
	// Create container with invalid database config to test error handling
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:            "nonexistent-host",
			Port:            9999,
			User:            "invalid",
			Password:        "invalid",
			Database:        "invalid",
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 300,
		},
		Redis: config.RedisConfig{
			Host:         "nonexistent-redis",
			Port:         9999,
			Password:     "",
			Database:     0,
			PoolSize:     10,
			MinIdleConns: 2,
		},
	}

	container, err := NewContainer(cfg, logger)
	
	// Should return an error due to invalid configuration
	assert.Error(t, err)
	assert.Nil(t, container)
}

func TestContainer_HealthCheckWithoutConnections(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{}

	// Create a container without initializing connections
	container := &Container{
		Config: cfg,
		Logger: logger,
	}

	ctx := context.Background()
	status := container.HealthCheck(ctx)

	assert.NotNil(t, status)
	assert.Equal(t, "not initialized", status["database"])
	assert.Equal(t, "not initialized", status["redis"])
}
