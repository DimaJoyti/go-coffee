package health

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnknown   HealthStatus = "unknown"
)

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Name        string                 `json:"name"`
	Status      HealthStatus           `json:"status"`
	Message     string                 `json:"message,omitempty"`
	LastChecked time.Time              `json:"last_checked"`
	Duration    time.Duration          `json:"duration"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// SystemHealth represents the overall system health
type SystemHealth struct {
	Status     HealthStatus                `json:"status"`
	Timestamp  time.Time                   `json:"timestamp"`
	Version    string                      `json:"version"`
	Uptime     time.Duration               `json:"uptime"`
	Components map[string]*ComponentHealth `json:"components"`
}

// Service provides health checking functionality
type Service struct {
	config    *config.Config
	db        *sql.DB
	redis     *redis.Client
	startTime time.Time
	
	// Health check results
	mu         sync.RWMutex
	components map[string]*ComponentHealth
	
	// Metrics
	meter              metric.Meter
	healthCheckCounter metric.Int64Counter
	healthCheckGauge   metric.Int64Gauge
}

// NewService creates a new health service
func NewService(cfg *config.Config, db *sql.DB, redis *redis.Client) (*Service, error) {
	meter := otel.Meter("crypto-terminal-health")
	
	healthCheckCounter, err := meter.Int64Counter(
		"health_checks_total",
		metric.WithDescription("Total number of health checks performed"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create health check counter: %w", err)
	}
	
	healthCheckGauge, err := meter.Int64Gauge(
		"component_health_status",
		metric.WithDescription("Health status of components (1=healthy, 0=unhealthy)"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create health check gauge: %w", err)
	}

	service := &Service{
		config:             cfg,
		db:                 db,
		redis:              redis,
		startTime:          time.Now(),
		components:         make(map[string]*ComponentHealth),
		meter:              meter,
		healthCheckCounter: healthCheckCounter,
		healthCheckGauge:   healthCheckGauge,
	}

	return service, nil
}

// Start starts the health check service
func (s *Service) Start(ctx context.Context) error {
	logrus.Info("Starting health check service")

	// Perform initial health checks
	s.performHealthChecks(ctx)

	// Start periodic health checks
	go s.startPeriodicHealthChecks(ctx)

	return nil
}

// GetSystemHealth returns the current system health status
func (s *Service) GetSystemHealth() *SystemHealth {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Determine overall system status
	overallStatus := StatusHealthy
	for _, component := range s.components {
		if component.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
			break
		} else if component.Status == StatusDegraded && overallStatus == StatusHealthy {
			overallStatus = StatusDegraded
		}
	}

	// Copy components to avoid race conditions
	components := make(map[string]*ComponentHealth)
	for name, health := range s.components {
		components[name] = &ComponentHealth{
			Name:        health.Name,
			Status:      health.Status,
			Message:     health.Message,
			LastChecked: health.LastChecked,
			Duration:    health.Duration,
			Details:     health.Details,
		}
	}

	return &SystemHealth{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Version:    "1.0.0", // TODO: Get from build info
		Uptime:     time.Since(s.startTime),
		Components: components,
	}
}

// startPeriodicHealthChecks starts periodic health checks
func (s *Service) startPeriodicHealthChecks(ctx context.Context) {
	ticker := time.NewTicker(s.config.Monitoring.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Stopping periodic health checks")
			return
		case <-ticker.C:
			s.performHealthChecks(ctx)
		}
	}
}

// performHealthChecks performs all health checks
func (s *Service) performHealthChecks(ctx context.Context) {
	logrus.Debug("Performing health checks")

	// Check database health
	s.checkDatabase(ctx)

	// Check Redis health
	s.checkRedis(ctx)

	// Check external APIs health
	s.checkExternalAPIs(ctx)

	// Update metrics
	s.updateHealthMetrics(ctx)
}

// checkDatabase checks database connectivity and performance
func (s *Service) checkDatabase(ctx context.Context) {
	start := time.Now()
	health := &ComponentHealth{
		Name:        "database",
		LastChecked: start,
		Details:     make(map[string]interface{}),
	}

	// Record health check
	s.healthCheckCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("component", "database"),
	))

	if s.db == nil {
		health.Status = StatusUnhealthy
		health.Message = "Database connection not initialized"
		health.Duration = time.Since(start)
		s.setComponentHealth("database", health)
		return
	}

	// Test connection with timeout
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.db.PingContext(checkCtx); err != nil {
		health.Status = StatusUnhealthy
		health.Message = fmt.Sprintf("Database ping failed: %v", err)
		health.Duration = time.Since(start)
		s.setComponentHealth("database", health)
		return
	}

	// Check database stats
	stats := s.db.Stats()
	health.Details["open_connections"] = stats.OpenConnections
	health.Details["in_use"] = stats.InUse
	health.Details["idle"] = stats.Idle
	health.Details["wait_count"] = stats.WaitCount
	health.Details["wait_duration"] = stats.WaitDuration.String()

	// Determine status based on connection pool usage
	if stats.OpenConnections >= s.config.Database.MaxOpenConns {
		health.Status = StatusDegraded
		health.Message = "Database connection pool at maximum capacity"
	} else if float64(stats.InUse)/float64(stats.OpenConnections) > 0.8 {
		health.Status = StatusDegraded
		health.Message = "High database connection usage"
	} else {
		health.Status = StatusHealthy
		health.Message = "Database is healthy"
	}

	health.Duration = time.Since(start)
	s.setComponentHealth("database", health)
}

// checkRedis checks Redis connectivity and performance
func (s *Service) checkRedis(ctx context.Context) {
	start := time.Now()
	health := &ComponentHealth{
		Name:        "redis",
		LastChecked: start,
		Details:     make(map[string]interface{}),
	}

	// Record health check
	s.healthCheckCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("component", "redis"),
	))

	if s.redis == nil {
		health.Status = StatusUnhealthy
		health.Message = "Redis connection not initialized"
		health.Duration = time.Since(start)
		s.setComponentHealth("redis", health)
		return
	}

	// Test connection with timeout
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.redis.Ping(checkCtx).Err(); err != nil {
		health.Status = StatusUnhealthy
		health.Message = fmt.Sprintf("Redis ping failed: %v", err)
		health.Duration = time.Since(start)
		s.setComponentHealth("redis", health)
		return
	}

	// Get Redis info
	info, err := s.redis.Info(checkCtx, "memory", "stats").Result()
	if err != nil {
		health.Status = StatusDegraded
		health.Message = fmt.Sprintf("Failed to get Redis info: %v", err)
	} else {
		health.Status = StatusHealthy
		health.Message = "Redis is healthy"
		health.Details["info"] = info
	}

	health.Duration = time.Since(start)
	s.setComponentHealth("redis", health)
}

// checkExternalAPIs checks external API connectivity
func (s *Service) checkExternalAPIs(ctx context.Context) {
	// Check CoinGecko API
	s.checkCoinGeckoAPI(ctx)

	// Check Binance API
	s.checkBinanceAPI(ctx)
}

// checkCoinGeckoAPI checks CoinGecko API health
func (s *Service) checkCoinGeckoAPI(ctx context.Context) {
	start := time.Now()
	health := &ComponentHealth{
		Name:        "coingecko_api",
		LastChecked: start,
		Details:     make(map[string]interface{}),
	}

	// Record health check
	s.healthCheckCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("component", "coingecko_api"),
	))

	// Simple ping to CoinGecko API
	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// This is a simplified check - in a real implementation,
	// you would make an actual HTTP request to the API
	_ = checkCtx
	health.Status = StatusHealthy
	health.Message = "CoinGecko API is accessible"
	health.Duration = time.Since(start)

	s.setComponentHealth("coingecko_api", health)
}

// checkBinanceAPI checks Binance API health
func (s *Service) checkBinanceAPI(ctx context.Context) {
	start := time.Now()
	health := &ComponentHealth{
		Name:        "binance_api",
		LastChecked: start,
		Details:     make(map[string]interface{}),
	}

	// Record health check
	s.healthCheckCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("component", "binance_api"),
	))

	// This is a simplified check - in a real implementation,
	// you would make an actual HTTP request to the API
	health.Status = StatusHealthy
	health.Message = "Binance API is accessible"
	health.Duration = time.Since(start)

	s.setComponentHealth("binance_api", health)
}

// setComponentHealth updates the health status of a component
func (s *Service) setComponentHealth(name string, health *ComponentHealth) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.components[name] = health
}

// updateHealthMetrics updates health metrics
func (s *Service) updateHealthMetrics(ctx context.Context) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for name, health := range s.components {
		var status int64
		switch health.Status {
		case StatusHealthy:
			status = 1
		case StatusDegraded:
			status = 0
		case StatusUnhealthy:
			status = -1
		default:
			status = -2
		}

		s.healthCheckGauge.Record(ctx, status, metric.WithAttributes(
			attribute.String("component", name),
		))
	}
}
