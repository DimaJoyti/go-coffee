package infrastructure

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/events"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/monitoring"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// TestInfrastructureIntegration tests the complete infrastructure integration
func TestInfrastructureIntegration(t *testing.T) {
	// Skip if running in CI without infrastructure dependencies
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test configuration
	cfg := createTestConfig()
	logger := logger.New("integration-test")

	// Create and initialize container
	container := NewContainer(cfg, logger)
	ctx := context.Background()

	err := container.Initialize(ctx)
	require.NoError(t, err, "Failed to initialize infrastructure container")
	defer container.Shutdown(ctx)

	// Test infrastructure components
	t.Run("Redis", func(t *testing.T) {
		testRedisIntegration(t, container)
	})

	t.Run("Cache", func(t *testing.T) {
		testCacheIntegration(t, container)
	})

	t.Run("Security", func(t *testing.T) {
		testSecurityIntegration(t, container)
	})

	t.Run("Events", func(t *testing.T) {
		testEventsIntegration(t, container)
	})

	t.Run("HealthChecks", func(t *testing.T) {
		testHealthChecksIntegration(t, container)
	})

	t.Run("Metrics", func(t *testing.T) {
		testMetricsIntegration(t, container)
	})

	t.Run("SessionManagement", func(t *testing.T) {
		testSessionManagementIntegration(t, container)
	})
}

// createTestConfig creates a test configuration
func createTestConfig() *config.InfrastructureConfig {
	cfg := config.DefaultInfrastructureConfig()

	// Use test database
	cfg.Redis.DB = 15
	cfg.Database.Database = "go_coffee_test"

	// Use test JWT secret
	cfg.Security.JWT.SecretKey = "test-secret-key-for-integration-testing-32-chars"
	cfg.Security.Encryption.AESKey = "test-aes-key-32-chars-for-testing"

	// Shorter timeouts for testing
	cfg.Security.JWT.AccessTokenTTL = 5 * time.Minute
	cfg.Security.JWT.RefreshTokenTTL = 1 * time.Hour

	return cfg
}

// testRedisIntegration tests Redis functionality
func testRedisIntegration(t *testing.T, container ContainerInterface) {
	redis := container.GetRedis()
	require.NotNil(t, redis, "Redis client should be available")

	ctx := context.Background()

	// Test basic operations
	err := redis.Set(ctx, "test:key", "test:value", time.Minute)
	assert.NoError(t, err, "Should be able to set key")

	value, err := redis.Get(ctx, "test:key")
	assert.NoError(t, err, "Should be able to get key")
	assert.Equal(t, "test:value", value, "Value should match")

	// Test hash operations
	err = redis.HSet(ctx, "test:hash", "field1", "value1", "field2", "value2")
	assert.NoError(t, err, "Should be able to set hash")

	hashValue, err := redis.HGet(ctx, "test:hash", "field1")
	assert.NoError(t, err, "Should be able to get hash field")
	assert.Equal(t, "value1", hashValue, "Hash value should match")

	// Test pub/sub
	err = redis.Publish(ctx, "test:channel", "test:message")
	assert.NoError(t, err, "Should be able to publish message")

	// Cleanup
	err = redis.Del(ctx, "test:key", "test:hash")
	assert.NoError(t, err, "Should be able to delete keys")
}

// testCacheIntegration tests cache functionality
func testCacheIntegration(t *testing.T, container ContainerInterface) {
	cache := container.GetCache()
	require.NotNil(t, cache, "Cache should be available")

	ctx := context.Background()

	// Test basic cache operations
	type TestData struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	testData := &TestData{ID: "123", Name: "Test"}

	err := cache.Set(ctx, "test:cache:key", testData, time.Minute)
	assert.NoError(t, err, "Should be able to set cache value")

	var retrievedData TestData
	err = cache.Get(ctx, "test:cache:key", &retrievedData)
	assert.NoError(t, err, "Should be able to get cache value")
	assert.Equal(t, testData.ID, retrievedData.ID, "Cached data should match")
	assert.Equal(t, testData.Name, retrievedData.Name, "Cached data should match")

	// Test batch operations
	items := map[string]interface{}{
		"test:batch:1": &TestData{ID: "1", Name: "Item1"},
		"test:batch:2": &TestData{ID: "2", Name: "Item2"},
	}

	err = cache.SetMulti(ctx, items, time.Minute)
	assert.NoError(t, err, "Should be able to set multiple cache values")

	results, err := cache.GetMulti(ctx, []string{"test:batch:1", "test:batch:2"})
	assert.NoError(t, err, "Should be able to get multiple cache values")
	assert.Len(t, results, 2, "Should retrieve all cached items")

	// Test cache stats
	stats, err := cache.Stats(ctx)
	assert.NoError(t, err, "Should be able to get cache stats")
	assert.NotNil(t, stats, "Stats should not be nil")

	// Cleanup
	err = cache.DeleteMulti(ctx, []string{"test:cache:key", "test:batch:1", "test:batch:2"})
	assert.NoError(t, err, "Should be able to delete cache keys")
}

// testSecurityIntegration tests security functionality
func testSecurityIntegration(t *testing.T, container ContainerInterface) {
	jwtService := container.GetJWTService()
	require.NotNil(t, jwtService, "JWT service should be available")

	encryptionService := container.GetEncryptionService()
	require.NotNil(t, encryptionService, "Encryption service should be available")

	ctx := context.Background()

	// Test JWT operations
	tokenPair, err := jwtService.GenerateTokenPair(ctx, "user123", "test@example.com", "user", map[string]interface{}{
		"test": "metadata",
	})
	assert.NoError(t, err, "Should be able to generate token pair")
	assert.NotEmpty(t, tokenPair.AccessToken, "Access token should not be empty")
	assert.NotEmpty(t, tokenPair.RefreshToken, "Refresh token should not be empty")

	// Validate access token
	claims, err := jwtService.ValidateAccessToken(ctx, tokenPair.AccessToken)
	assert.NoError(t, err, "Should be able to validate access token")
	assert.Equal(t, "user123", claims.UserID, "User ID should match")
	assert.Equal(t, "test@example.com", claims.Email, "Email should match")

	// Test token refresh
	newTokenPair, err := jwtService.RefreshAccessToken(ctx, tokenPair.RefreshToken)
	assert.NoError(t, err, "Should be able to refresh token")
	assert.NotEmpty(t, newTokenPair.AccessToken, "New access token should not be empty")

	// Test encryption
	plaintext := "sensitive data to encrypt"
	encrypted, err := encryptionService.Encrypt(plaintext)
	assert.NoError(t, err, "Should be able to encrypt data")
	assert.NotEqual(t, plaintext, encrypted, "Encrypted data should be different")

	decrypted, err := encryptionService.Decrypt(encrypted)
	assert.NoError(t, err, "Should be able to decrypt data")
	assert.Equal(t, plaintext, decrypted, "Decrypted data should match original")

	// Test password hashing
	password := "test-password-123"
	hashedPassword, err := encryptionService.HashPassword(password)
	assert.NoError(t, err, "Should be able to hash password")
	assert.NotEqual(t, password, hashedPassword, "Hashed password should be different")

	err = encryptionService.VerifyPassword(hashedPassword, password)
	assert.NoError(t, err, "Should be able to verify password")

	err = encryptionService.VerifyPassword(hashedPassword, "wrong-password")
	assert.Error(t, err, "Should fail to verify wrong password")
}

// testEventsIntegration tests event functionality
func testEventsIntegration(t *testing.T, container ContainerInterface) {
	eventStore := container.GetEventStore()
	eventPublisher := container.GetEventPublisher()
	eventSubscriber := container.GetEventSubscriber()

	require.NotNil(t, eventStore, "Event store should be available")
	require.NotNil(t, eventPublisher, "Event publisher should be available")
	require.NotNil(t, eventSubscriber, "Event subscriber should be available")

	ctx := context.Background()

	// Test event store
	event := &events.Event{
		ID:            "test-event-123",
		Type:          "test.event",
		Source:        "integration-test",
		AggregateID:   "aggregate-123",
		AggregateType: "test-aggregate",
		Version:       1,
		Data: map[string]interface{}{
			"test": "data",
		},
		Timestamp: time.Now(),
	}

	err := eventStore.SaveEvent(ctx, event)
	assert.NoError(t, err, "Should be able to save event")

	retrievedEvent, err := eventStore.GetEvent(ctx, event.ID)
	assert.NoError(t, err, "Should be able to retrieve event")
	assert.Equal(t, event.ID, retrievedEvent.ID, "Event ID should match")
	assert.Equal(t, event.Type, retrievedEvent.Type, "Event type should match")

	// Test event publishing
	publishEvent := &events.Event{
		ID:            "publish-test-123",
		Type:          "test.publish",
		Source:        "integration-test",
		AggregateID:   "aggregate-456",
		AggregateType: "test-aggregate",
		Data: map[string]interface{}{
			"message": "test publish",
		},
		Timestamp: time.Now(),
	}

	err = eventPublisher.Publish(ctx, publishEvent)
	assert.NoError(t, err, "Should be able to publish event")

	// Test event count
	count, err := eventStore.GetEventCount(ctx)
	assert.NoError(t, err, "Should be able to get event count")
	assert.GreaterOrEqual(t, count, int64(2), "Should have at least 2 events")
}

// testHealthChecksIntegration tests health check functionality
func testHealthChecksIntegration(t *testing.T, container ContainerInterface) {
	healthConfig := &monitoring.HealthConfig{
		Enabled:          true,
		CheckInterval:    10 * time.Second,
		Timeout:          5 * time.Second,
		FailureThreshold: 3,
		SuccessThreshold: 1,
	}

	logger := logger.New("health-test")
	healthChecker := monitoring.NewHealthChecker(container, healthConfig, logger)

	ctx := context.Background()
	health := healthChecker.CheckHealth(ctx)

	assert.NotNil(t, health, "Health check result should not be nil")
	assert.NotEmpty(t, health.Checks, "Should have health checks")
	assert.Greater(t, health.Summary.Total, 0, "Should have total checks")

	// Check individual components
	if redisCheck, exists := health.Checks["redis"]; exists {
		assert.Equal(t, monitoring.HealthStatusHealthy, redisCheck.Status, "Redis should be healthy")
	}

	if cacheCheck, exists := health.Checks["cache"]; exists {
		assert.Equal(t, monitoring.HealthStatusHealthy, cacheCheck.Status, "Cache should be healthy")
	}
}

// testMetricsIntegration tests metrics functionality
func testMetricsIntegration(t *testing.T, container ContainerInterface) {
	metricsConfig := &monitoring.MetricsConfig{
		Enabled:         true,
		CollectInterval: 5 * time.Second,
		RetentionPeriod: 1 * time.Hour,
		Namespace:       "test",
		ServiceName:     "integration-test",
	}

	logger := logger.New("metrics-test")
	metricsCollector := monitoring.NewMetricsCollector(container, metricsConfig, logger)

	ctx := context.Background()
	err := metricsCollector.Start(ctx)
	assert.NoError(t, err, "Should be able to start metrics collector")
	defer metricsCollector.Stop(ctx)

	// Record some test metrics
	metricsCollector.IncrementCounter("test_counter", map[string]string{"type": "integration"})
	metricsCollector.SetGauge("test_gauge", 42.0, map[string]string{"unit": "test"})

	// Get metrics snapshot
	snapshot := metricsCollector.GetMetricsSnapshot()
	assert.NotNil(t, snapshot, "Metrics snapshot should not be nil")
	assert.Equal(t, "integration-test", snapshot.ServiceName, "Service name should match")
	assert.Greater(t, snapshot.System.GoroutineCount, 0, "Should have goroutines")

	// Test metrics export
	prometheusMetrics := metricsCollector.ExportPrometheusMetrics()
	assert.NotEmpty(t, prometheusMetrics, "Prometheus metrics should not be empty")

	jsonMetrics, err := metricsCollector.ExportJSONMetrics()
	assert.NoError(t, err, "Should be able to export JSON metrics")
	assert.NotEmpty(t, jsonMetrics, "JSON metrics should not be empty")
}

// testSessionManagementIntegration tests session management functionality
func testSessionManagementIntegration(t *testing.T, container ContainerInterface) {
	// This would test the session management functionality
	// For now, we'll just verify that the required components are available

	cache := container.GetCache()
	jwtService := container.GetJWTService()

	assert.NotNil(t, cache, "Cache should be available for session management")
	assert.NotNil(t, jwtService, "JWT service should be available for session management")

	// Test session storage in cache
	ctx := context.Background()
	sessionData := map[string]interface{}{
		"user_id":    "user123",
		"session_id": "session123",
		"created_at": time.Now(),
	}

	err := cache.Set(ctx, "session:session123", sessionData, time.Hour)
	assert.NoError(t, err, "Should be able to store session data")

	var retrievedSession map[string]interface{}
	err = cache.Get(ctx, "session:session123", &retrievedSession)
	assert.NoError(t, err, "Should be able to retrieve session data")
	assert.Equal(t, "user123", retrievedSession["user_id"], "Session user ID should match")

	// Cleanup
	err = cache.Delete(ctx, "session:session123")
	assert.NoError(t, err, "Should be able to delete session")
}

// BenchmarkInfrastructureOperations benchmarks infrastructure operations
func BenchmarkInfrastructureOperations(b *testing.B) {
	cfg := createTestConfig()
	logger := logger.New("benchmark")
	container := NewContainer(cfg, logger)

	ctx := context.Background()
	err := container.Initialize(ctx)
	if err != nil {
		b.Fatal("Failed to initialize container:", err)
	}
	defer container.Shutdown(ctx)

	b.Run("Redis Set/Get", func(b *testing.B) {
		redis := container.GetRedis()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench:key:%d", i)
			redis.Set(ctx, key, "value", time.Minute)
			redis.Get(ctx, key)
		}
	})

	b.Run("Cache Set/Get", func(b *testing.B) {
		cache := container.GetCache()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench:cache:%d", i)
			cache.Set(ctx, key, "value", time.Minute)
			var value string
			cache.Get(ctx, key, &value)
		}
	})

	b.Run("JWT Generate/Validate", func(b *testing.B) {
		jwtService := container.GetJWTService()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			tokenPair, _ := jwtService.GenerateTokenPair(ctx, "user123", "test@example.com", "user", nil)
			jwtService.ValidateAccessToken(ctx, tokenPair.AccessToken)
		}
	})

	b.Run("Encryption/Decryption", func(b *testing.B) {
		encryptionService := container.GetEncryptionService()
		plaintext := "test data to encrypt"
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			encrypted, _ := encryptionService.Encrypt(plaintext)
			encryptionService.Decrypt(encrypted)
		}
	})
}
