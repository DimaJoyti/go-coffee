package optimization

import (
	"context"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/cache"
	"github.com/DimaJoyti/go-coffee/pkg/database"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
	"github.com/DimaJoyti/go-coffee/pkg/performance"
	"go.uber.org/zap"
)

// Service provides centralized optimization management
type Service struct {
	logger *zap.Logger
	config *config.InfrastructureConfig

	// Optimization components
	databaseManager *database.Manager
	cacheManager    *cache.Manager
	memoryOptimizer *performance.MemoryOptimizer

	// Metrics and monitoring
	metrics       *Metrics
	healthChecker *HealthChecker

	// State
	running bool
}

// NewService creates a new optimization service
func NewService(cfg *config.InfrastructureConfig, logger *zap.Logger) (*Service, error) {
	service := &Service{
		logger:  logger,
		config:  cfg,
		metrics: NewMetrics(),
	}

	// Initialize database optimization
	if err := service.initializeDatabaseOptimization(); err != nil {
		return nil, fmt.Errorf("failed to initialize database optimization: %w", err)
	}

	// Initialize cache optimization
	if err := service.initializeCacheOptimization(); err != nil {
		return nil, fmt.Errorf("failed to initialize cache optimization: %w", err)
	}

	// Initialize memory optimization
	if err := service.initializeMemoryOptimization(); err != nil {
		return nil, fmt.Errorf("failed to initialize memory optimization: %w", err)
	}

	// Initialize health checker
	service.healthChecker = NewHealthChecker(service)

	return service, nil
}

// Start starts the optimization service
func (s *Service) Start(ctx context.Context) error {
	if s.running {
		return fmt.Errorf("optimization service is already running")
	}

	s.logger.Info("Starting optimization service")

	// Start health monitoring
	go s.healthChecker.Start(ctx)

	// Start metrics collection
	go s.startMetricsCollection(ctx)

	s.running = true
	s.logger.Info("Optimization service started successfully")

	return nil
}

// Stop stops the optimization service
func (s *Service) Stop(ctx context.Context) error {
	if !s.running {
		return nil
	}

	s.logger.Info("Stopping optimization service")

	// Close database connections
	if s.databaseManager != nil {
		s.databaseManager.Close()
	}

	// Close cache connections
	if s.cacheManager != nil {
		s.cacheManager.Close()
	}

	s.running = false
	s.logger.Info("Optimization service stopped")

	return nil
}

// GetDatabaseManager returns the optimized database manager
func (s *Service) GetDatabaseManager() *database.Manager {
	return s.databaseManager
}

// GetCacheManager returns the optimized cache manager
func (s *Service) GetCacheManager() *cache.Manager {
	return s.cacheManager
}

// GetMemoryOptimizer returns the memory optimizer
func (s *Service) GetMemoryOptimizer() *performance.MemoryOptimizer {
	return s.memoryOptimizer
}

// GetMetrics returns current optimization metrics
func (s *Service) GetMetrics() *Metrics {
	return s.metrics
}

// initializeDatabaseOptimization sets up database optimization
func (s *Service) initializeDatabaseOptimization() error {
	s.logger.Info("Initializing database optimization")

	dbManager, err := database.NewManager(s.config.Database, s.logger)
	if err != nil {
		return fmt.Errorf("failed to create database manager: %w", err)
	}

	s.databaseManager = dbManager
	s.logger.Info("Database optimization initialized")

	return nil
}

// initializeCacheOptimization sets up cache optimization
func (s *Service) initializeCacheOptimization() error {
	s.logger.Info("Initializing cache optimization")

	cacheManager, err := cache.NewManager(s.config.Redis, s.logger)
	if err != nil {
		return fmt.Errorf("failed to create cache manager: %w", err)
	}

	s.cacheManager = cacheManager
	s.logger.Info("Cache optimization initialized")

	return nil
}

// initializeMemoryOptimization sets up memory optimization
func (s *Service) initializeMemoryOptimization() error {
	s.logger.Info("Initializing memory optimization")

	memoryConfig := &performance.MemoryConfig{
		GCPercent:            100,
		MaxGCPause:           10 * time.Millisecond,
		GCTriggerThreshold:   0.8,
		PoolCleanupInterval:  5 * time.Minute,
		MaxPoolSize:          1000,
		PoolIdleTimeout:      10 * time.Minute,
		MonitorInterval:      1 * time.Minute,
		MemoryThreshold:      0.8,
		LeakDetectionEnabled: true,
		EnableAutoGC:         true,
		EnablePooling:        true,
		EnableProfiling:      true,
	}

	memoryOptimizer := performance.NewMemoryOptimizer(memoryConfig, s.logger)
	s.memoryOptimizer = memoryOptimizer
	s.logger.Info("Memory optimization initialized")

	return nil
}

// startMetricsCollection starts collecting optimization metrics
func (s *Service) startMetricsCollection(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.collectMetrics()
		}
	}
}

// collectMetrics collects metrics from all optimization components
func (s *Service) collectMetrics() {
	// Collect database metrics
	if s.databaseManager != nil {
		dbMetrics := s.databaseManager.GetMetrics()
		s.metrics.UpdateDatabaseMetrics(&dbMetrics)
	}

	// Collect cache metrics
	if s.cacheManager != nil {
		cacheMetrics := s.cacheManager.GetMetrics()
		s.metrics.UpdateCacheMetrics(&cacheMetrics)
	}

	// Collect memory metrics
	if s.memoryOptimizer != nil {
		memoryStats := s.memoryOptimizer.GetOptimizationStats()
		s.metrics.UpdateMemoryMetrics(memoryStats)
	}
}

// Metrics aggregates optimization metrics
type Metrics struct {
	Database DatabaseMetrics `json:"database"`
	Cache    CacheMetrics    `json:"cache"`
	Memory   MemoryMetrics   `json:"memory"`
	System   SystemMetrics   `json:"system"`
}

// DatabaseMetrics contains database optimization metrics
type DatabaseMetrics struct {
	QueryCount        int64         `json:"query_count"`
	SlowQueryCount    int64         `json:"slow_query_count"`
	ConnectionErrors  int64         `json:"connection_errors"`
	ActiveConnections int32         `json:"active_connections"`
	IdleConnections   int32         `json:"idle_connections"`
	AverageQueryTime  time.Duration `json:"average_query_time"`
}

// CacheMetrics contains cache optimization metrics
type CacheMetrics struct {
	Hits        int64         `json:"hits"`
	Misses      int64         `json:"misses"`
	Sets        int64         `json:"sets"`
	Deletes     int64         `json:"deletes"`
	Errors      int64         `json:"errors"`
	HitRatio    float64       `json:"hit_ratio"`
	AvgLatency  time.Duration `json:"avg_latency"`
	TotalKeys   int64         `json:"total_keys"`
	MemoryUsage int64         `json:"memory_usage"`
}

// MemoryMetrics contains memory optimization metrics
type MemoryMetrics struct {
	Alloc         uint64  `json:"alloc"`
	TotalAlloc    uint64  `json:"total_alloc"`
	Sys           uint64  `json:"sys"`
	NumGC         uint32  `json:"num_gc"`
	GCCPUFraction float64 `json:"gc_cpu_fraction"`
	HeapInuse     uint64  `json:"heap_inuse"`
	HeapObjects   uint64  `json:"heap_objects"`
}

// SystemMetrics contains system-level metrics
type SystemMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	NetworkIO   int64   `json:"network_io"`
	Goroutines  int     `json:"goroutines"`
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{}
}

// UpdateDatabaseMetrics updates database metrics
func (m *Metrics) UpdateDatabaseMetrics(dbMetrics *database.DatabaseMetrics) {
	m.Database = DatabaseMetrics{
		QueryCount:        dbMetrics.QueryCount,
		SlowQueryCount:    dbMetrics.SlowQueryCount,
		ConnectionErrors:  dbMetrics.ConnectionErrors,
		ActiveConnections: dbMetrics.ActiveConnections,
		IdleConnections:   dbMetrics.IdleConnections,
		AverageQueryTime:  dbMetrics.AverageQueryTime,
	}
}

// UpdateCacheMetrics updates cache metrics
func (m *Metrics) UpdateCacheMetrics(cacheMetrics *cache.CacheMetrics) {
	m.Cache = CacheMetrics{
		Hits:        cacheMetrics.Hits,
		Misses:      cacheMetrics.Misses,
		Sets:        cacheMetrics.Sets,
		Deletes:     cacheMetrics.Deletes,
		Errors:      cacheMetrics.Errors,
		HitRatio:    cacheMetrics.HitRatio,
		AvgLatency:  cacheMetrics.AvgLatency,
		TotalKeys:   cacheMetrics.TotalKeys,
		MemoryUsage: cacheMetrics.MemoryUsage,
	}
}

// UpdateMemoryMetrics updates memory metrics
func (m *Metrics) UpdateMemoryMetrics(memoryStats map[string]interface{}) {
	// Extract memory stats from the map
	if memStats, ok := memoryStats["memory"]; ok {
		if stats, ok := memStats.(map[string]interface{}); ok {
			m.Memory = MemoryMetrics{
				Alloc:         getUint64FromMap(stats, "Alloc", 0),
				TotalAlloc:    getUint64FromMap(stats, "TotalAlloc", 0),
				Sys:           getUint64FromMap(stats, "Sys", 0),
				NumGC:         uint32(getUint64FromMap(stats, "NumGC", 0)),
				GCCPUFraction: getFloat64FromMap(stats, "GCCPUFraction", 0),
				HeapInuse:     getUint64FromMap(stats, "HeapInuse", 0),
				HeapObjects:   getUint64FromMap(stats, "HeapObjects", 0),
			}
		}
	} else {
		// If memory stats are not structured as expected, set default values
		m.Memory = MemoryMetrics{
			Alloc:         0,
			TotalAlloc:    0,
			Sys:           0,
			NumGC:         0,
			GCCPUFraction: 0,
			HeapInuse:     0,
			HeapObjects:   0,
		}
	}
}

// Helper functions to safely extract values from a map
func getUint64FromMap(m map[string]interface{}, key string, defaultVal uint64) uint64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case uint64:
			return v
		case int64:
			return uint64(v)
		case int:
			return uint64(v)
		case float64:
			return uint64(v)
		}
	}
	return defaultVal
}

func getFloat64FromMap(m map[string]interface{}, key string, defaultVal float64) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case uint64:
			return float64(v)
		}
	}
	return defaultVal
}

// HealthChecker monitors the health of optimization components
type HealthChecker struct {
	service *Service
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(service *Service) *HealthChecker {
	return &HealthChecker{service: service}
}

// Start starts the health checker
func (hc *HealthChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.checkHealth()
		}
	}
}

// checkHealth performs health checks on all components
func (hc *HealthChecker) checkHealth() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check database health
	if hc.service.databaseManager != nil {
		if err := hc.service.databaseManager.HealthCheck(ctx); err != nil {
			hc.service.logger.Error("Database health check failed", zap.Error(err))
		}
	}

	// Check cache health
	if hc.service.cacheManager != nil {
		// Implement cache health check
		if _, err := hc.service.cacheManager.Exists(ctx, "health_check"); err != nil {
			hc.service.logger.Error("Cache health check failed", zap.Error(err))
		}
	}
}

// GetHealthStatus returns the current health status
func (hc *HealthChecker) GetHealthStatus() map[string]bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status := make(map[string]bool)

	// Database health
	if hc.service.databaseManager != nil {
		status["database"] = hc.service.databaseManager.HealthCheck(ctx) == nil
	}

	// Cache health
	if hc.service.cacheManager != nil {
		_, err := hc.service.cacheManager.Exists(ctx, "health_check")
		status["cache"] = err == nil
	}

	// Memory optimizer health
	status["memory_optimizer"] = hc.service.memoryOptimizer != nil

	return status
}

// OptimizationReport provides a comprehensive optimization report
type OptimizationReport struct {
	Timestamp    time.Time              `json:"timestamp"`
	Metrics      *Metrics               `json:"metrics"`
	Health       map[string]bool        `json:"health"`
	Improvements map[string]interface{} `json:"improvements"`
}

// GenerateReport generates a comprehensive optimization report
func (s *Service) GenerateReport() *OptimizationReport {
	return &OptimizationReport{
		Timestamp:    time.Now(),
		Metrics:      s.metrics,
		Health:       s.healthChecker.GetHealthStatus(),
		Improvements: s.calculateImprovements(),
	}
}

// calculateImprovements calculates performance improvements
func (s *Service) calculateImprovements() map[string]interface{} {
	improvements := make(map[string]interface{})

	// Database improvements
	if s.databaseManager != nil {
		dbMetrics := s.databaseManager.GetMetrics()
		totalConnections := float64(dbMetrics.TotalConnections)
		queryCount := float64(dbMetrics.QueryCount)
		
		// Avoid division by zero
		connectionPoolEfficiency := 0.0
		if totalConnections > 0 {
			connectionPoolEfficiency = float64(dbMetrics.ActiveConnections) / totalConnections
		}
		
		errorRate := 0.0
		if queryCount > 0 {
			errorRate = float64(dbMetrics.ConnectionErrors) / queryCount
		}
		
		improvements["database"] = map[string]interface{}{
			"connection_pool_efficiency": connectionPoolEfficiency,
			"query_performance":          dbMetrics.AverageQueryTime.Milliseconds(),
			"error_rate":                 errorRate,
		}
	}

	// Cache improvements
	if s.cacheManager != nil {
		cacheMetrics := s.cacheManager.GetMetrics()
		improvements["cache"] = map[string]interface{}{
			"hit_ratio":    cacheMetrics.HitRatio,
			"avg_latency":  cacheMetrics.AvgLatency.Milliseconds(),
			"total_keys":   cacheMetrics.TotalKeys,
			"memory_usage": cacheMetrics.MemoryUsage,
		}
	}
	
	// Memory improvements
	if s.memoryOptimizer != nil {
		memStats := s.memoryOptimizer.GetOptimizationStats()
		improvements["memory"] = map[string]interface{}{
			"heap_optimization": memStats["heap_optimization"],
			"gc_efficiency":     memStats["gc_efficiency"],
			"memory_savings":    memStats["memory_savings"],
		}
	}

	return improvements
}
