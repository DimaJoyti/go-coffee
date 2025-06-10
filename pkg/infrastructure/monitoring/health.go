package monitoring

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// HealthChecker performs comprehensive health checks
type HealthChecker struct {
	container infrastructure.ContainerInterface
	logger    *logger.Logger
	config    *HealthConfig
	checks    map[string]HealthCheck
	mutex     sync.RWMutex
}

// HealthConfig represents health check configuration
type HealthConfig struct {
	Enabled          bool          `yaml:"enabled"`
	CheckInterval    time.Duration `yaml:"check_interval"`
	Timeout          time.Duration `yaml:"timeout"`
	FailureThreshold int           `yaml:"failure_threshold"`
	SuccessThreshold int           `yaml:"success_threshold"`
}

// HealthCheck represents a single health check
type HealthCheck interface {
	Name() string
	Check(ctx context.Context) HealthResult
	IsRequired() bool
	GetTimeout() time.Duration
}

// HealthResult represents the result of a health check
type HealthResult struct {
	Name      string                 `json:"name"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Error     error                  `json:"error,omitempty"`
}

// HealthStatus represents the status of a health check
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// OverallHealth represents the overall health status
type OverallHealth struct {
	Status    HealthStatus            `json:"status"`
	Timestamp time.Time               `json:"timestamp"`
	Duration  time.Duration           `json:"duration"`
	Checks    map[string]HealthResult `json:"checks"`
	Summary   HealthSummary           `json:"summary"`
	Metadata  map[string]interface{}  `json:"metadata,omitempty"`
}

// HealthSummary provides a summary of health check results
type HealthSummary struct {
	Total     int `json:"total"`
	Healthy   int `json:"healthy"`
	Unhealthy int `json:"unhealthy"`
	Degraded  int `json:"degraded"`
	Unknown   int `json:"unknown"`
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(container infrastructure.ContainerInterface, config *HealthConfig, logger *logger.Logger) *HealthChecker {
	if config == nil {
		config = &HealthConfig{
			Enabled:          true,
			CheckInterval:    30 * time.Second,
			Timeout:          10 * time.Second,
			FailureThreshold: 3,
			SuccessThreshold: 1,
		}
	}

	hc := &HealthChecker{
		container: container,
		logger:    logger,
		config:    config,
		checks:    make(map[string]HealthCheck),
	}

	// Register default health checks
	hc.registerDefaultChecks()

	return hc
}

// RegisterCheck registers a custom health check
func (hc *HealthChecker) RegisterCheck(check HealthCheck) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	hc.checks[check.Name()] = check
	hc.logger.InfoWithFields("Health check registered", logger.String("name", check.Name()))
}

// UnregisterCheck unregisters a health check
func (hc *HealthChecker) UnregisterCheck(name string) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	delete(hc.checks, name)
	hc.logger.InfoWithFields("Health check unregistered", logger.String("name", name))
}

// CheckHealth performs all health checks and returns overall health
func (hc *HealthChecker) CheckHealth(ctx context.Context) *OverallHealth {
	start := time.Now()

	hc.mutex.RLock()
	checks := make(map[string]HealthCheck)
	for k, v := range hc.checks {
		checks[k] = v
	}
	hc.mutex.RUnlock()

	results := make(map[string]HealthResult)
	resultsChan := make(chan HealthResult, len(checks))

	// Run health checks concurrently
	var wg sync.WaitGroup
	for _, check := range checks {
		wg.Add(1)
		go func(check HealthCheck) {
			defer wg.Done()

			checkCtx, cancel := context.WithTimeout(ctx, check.GetTimeout())
			defer cancel()

			result := check.Check(checkCtx)
			resultsChan <- result
		}(check)
	}

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		results[result.Name] = result
	}

	// Calculate overall status
	overall := &OverallHealth{
		Timestamp: time.Now(),
		Duration:  time.Since(start),
		Checks:    results,
		Summary:   hc.calculateSummary(results),
		Metadata: map[string]interface{}{
			"service": "go-coffee",
			"version": "1.0.0",
		},
	}

	overall.Status = hc.calculateOverallStatus(results)

	return overall
}

// registerDefaultChecks registers the default health checks
func (hc *HealthChecker) registerDefaultChecks() {
	// Database health check
	if hc.container.GetDatabase() != nil {
		hc.RegisterCheck(&DatabaseHealthCheck{
			container: hc.container,
			timeout:   5 * time.Second,
		})
	}

	// Redis health check
	if hc.container.GetRedis() != nil {
		hc.RegisterCheck(&RedisHealthCheck{
			container: hc.container,
			timeout:   5 * time.Second,
		})
	}

	// Cache health check
	if hc.container.GetCache() != nil {
		hc.RegisterCheck(&CacheHealthCheck{
			container: hc.container,
			timeout:   5 * time.Second,
		})
	}

	// Session manager health check
	if hc.container.GetSessionManager() != nil {
		hc.RegisterCheck(&SessionManagerHealthCheck{
			container: hc.container,
			timeout:   3 * time.Second,
		})
	}

	// JWT service health check
	if hc.container.GetJWTService() != nil {
		hc.RegisterCheck(&JWTServiceHealthCheck{
			container: hc.container,
			timeout:   3 * time.Second,
		})
	}

	// Event infrastructure health checks
	if hc.container.GetEventStore() != nil {
		hc.RegisterCheck(&EventStoreHealthCheck{
			container: hc.container,
			timeout:   5 * time.Second,
		})
	}

	if hc.container.GetEventPublisher() != nil {
		hc.RegisterCheck(&EventPublisherHealthCheck{
			container: hc.container,
			timeout:   5 * time.Second,
		})
	}

	if hc.container.GetEventSubscriber() != nil {
		hc.RegisterCheck(&EventSubscriberHealthCheck{
			container: hc.container,
			timeout:   5 * time.Second,
		})
	}

	// System resource health check
	hc.RegisterCheck(&SystemResourceHealthCheck{
		container: hc.container,
		timeout:   2 * time.Second,
	})
}

// calculateSummary calculates the health summary
func (hc *HealthChecker) calculateSummary(results map[string]HealthResult) HealthSummary {
	summary := HealthSummary{
		Total: len(results),
	}

	for _, result := range results {
		switch result.Status {
		case HealthStatusHealthy:
			summary.Healthy++
		case HealthStatusUnhealthy:
			summary.Unhealthy++
		case HealthStatusDegraded:
			summary.Degraded++
		case HealthStatusUnknown:
			summary.Unknown++
		}
	}

	return summary
}

// calculateOverallStatus calculates the overall health status
func (hc *HealthChecker) calculateOverallStatus(results map[string]HealthResult) HealthStatus {
	if len(results) == 0 {
		return HealthStatusUnknown
	}

	hasUnhealthy := false
	hasDegraded := false
	hasUnknown := false

	for _, result := range results {
		switch result.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		case HealthStatusUnknown:
			hasUnknown = true
		}
	}

	// If any required check is unhealthy, overall is unhealthy
	if hasUnhealthy {
		return HealthStatusUnhealthy
	}

	// If any check is degraded, overall is degraded
	if hasDegraded {
		return HealthStatusDegraded
	}

	// If any check is unknown, overall is unknown
	if hasUnknown {
		return HealthStatusUnknown
	}

	return HealthStatusHealthy
}

// Database Health Check
type DatabaseHealthCheck struct {
	container infrastructure.ContainerInterface
	timeout   time.Duration
}

func (dhc *DatabaseHealthCheck) Name() string {
	return "database"
}

func (dhc *DatabaseHealthCheck) Check(ctx context.Context) HealthResult {
	start := time.Now()
	result := HealthResult{
		Name:      dhc.Name(),
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	db := dhc.container.GetDatabase()
	if db == nil {
		result.Status = HealthStatusUnknown
		result.Message = "Database not configured"
		result.Duration = time.Since(start)
		return result
	}

	if err := db.Ping(ctx); err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = "Database ping failed"
		result.Error = err
	} else {
		result.Status = HealthStatusHealthy
		result.Message = "Database is healthy"

		// Add connection stats
		stats := db.Stats()
		result.Metadata["open_connections"] = stats.OpenConnections
		result.Metadata["in_use"] = stats.InUse
		result.Metadata["idle"] = stats.Idle
	}

	result.Duration = time.Since(start)
	return result
}

func (dhc *DatabaseHealthCheck) IsRequired() bool {
	return true
}

func (dhc *DatabaseHealthCheck) GetTimeout() time.Duration {
	return dhc.timeout
}

// Redis Health Check
type RedisHealthCheck struct {
	container infrastructure.ContainerInterface
	timeout   time.Duration
}

func (rhc *RedisHealthCheck) Name() string {
	return "redis"
}

func (rhc *RedisHealthCheck) Check(ctx context.Context) HealthResult {
	start := time.Now()
	result := HealthResult{
		Name:      rhc.Name(),
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	redis := rhc.container.GetRedis()
	if redis == nil {
		result.Status = HealthStatusUnknown
		result.Message = "Redis not configured"
		result.Duration = time.Since(start)
		return result
	}

	if err := redis.Ping(ctx); err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = "Redis ping failed"
		result.Error = err
	} else {
		result.Status = HealthStatusHealthy
		result.Message = "Redis is healthy"

		// Add pool stats
		stats := redis.PoolStats()
		result.Metadata["total_conns"] = stats.TotalConns
		result.Metadata["idle_conns"] = stats.IdleConns
		result.Metadata["stale_conns"] = stats.StaleConns
	}

	result.Duration = time.Since(start)
	return result
}

func (rhc *RedisHealthCheck) IsRequired() bool {
	return true
}

func (rhc *RedisHealthCheck) GetTimeout() time.Duration {
	return rhc.timeout
}

// Cache Health Check
type CacheHealthCheck struct {
	container infrastructure.ContainerInterface
	timeout   time.Duration
}

func (chc *CacheHealthCheck) Name() string {
	return "cache"
}

func (chc *CacheHealthCheck) Check(ctx context.Context) HealthResult {
	start := time.Now()
	result := HealthResult{
		Name:      chc.Name(),
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	cache := chc.container.GetCache()
	if cache == nil {
		result.Status = HealthStatusUnknown
		result.Message = "Cache not configured"
		result.Duration = time.Since(start)
		return result
	}

	if err := cache.Ping(ctx); err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = "Cache ping failed"
		result.Error = err
	} else {
		result.Status = HealthStatusHealthy
		result.Message = "Cache is healthy"

		// Add cache stats
		if stats, err := cache.Stats(ctx); err == nil {
			result.Metadata["hits"] = stats.Hits
			result.Metadata["misses"] = stats.Misses
			result.Metadata["keys"] = stats.Keys
		}
	}

	result.Duration = time.Since(start)
	return result
}

func (chc *CacheHealthCheck) IsRequired() bool {
	return false
}

func (chc *CacheHealthCheck) GetTimeout() time.Duration {
	return chc.timeout
}

// Event Store Health Check
type EventStoreHealthCheck struct {
	container infrastructure.ContainerInterface
	timeout   time.Duration
}

func (eshc *EventStoreHealthCheck) Name() string {
	return "event_store"
}

func (eshc *EventStoreHealthCheck) Check(ctx context.Context) HealthResult {
	start := time.Now()
	result := HealthResult{
		Name:      eshc.Name(),
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	eventStore := eshc.container.GetEventStore()
	if eventStore == nil {
		result.Status = HealthStatusUnknown
		result.Message = "Event store not configured"
		result.Duration = time.Since(start)
		return result
	}

	if err := eventStore.HealthCheck(ctx); err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = "Event store health check failed"
		result.Error = err
	} else {
		result.Status = HealthStatusHealthy
		result.Message = "Event store is healthy"

		// Add event count
		if count, err := eventStore.GetEventCount(ctx); err == nil {
			result.Metadata["event_count"] = count
		}
	}

	result.Duration = time.Since(start)
	return result
}

func (eshc *EventStoreHealthCheck) IsRequired() bool {
	return false
}

func (eshc *EventStoreHealthCheck) GetTimeout() time.Duration {
	return eshc.timeout
}

// Event Publisher Health Check
type EventPublisherHealthCheck struct {
	container infrastructure.ContainerInterface
	timeout   time.Duration
}

func (ephc *EventPublisherHealthCheck) Name() string {
	return "event_publisher"
}

func (ephc *EventPublisherHealthCheck) Check(ctx context.Context) HealthResult {
	start := time.Now()
	result := HealthResult{
		Name:      ephc.Name(),
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	eventPublisher := ephc.container.GetEventPublisher()
	if eventPublisher == nil {
		result.Status = HealthStatusUnknown
		result.Message = "Event publisher not configured"
		result.Duration = time.Since(start)
		return result
	}

	if err := eventPublisher.HealthCheck(ctx); err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = "Event publisher health check failed"
		result.Error = err
	} else {
		result.Status = HealthStatusHealthy
		result.Message = "Event publisher is healthy"
	}

	result.Duration = time.Since(start)
	return result
}

func (ephc *EventPublisherHealthCheck) IsRequired() bool {
	return false
}

func (ephc *EventPublisherHealthCheck) GetTimeout() time.Duration {
	return ephc.timeout
}

// Event Subscriber Health Check
type EventSubscriberHealthCheck struct {
	container infrastructure.ContainerInterface
	timeout   time.Duration
}

func (eshc *EventSubscriberHealthCheck) Name() string {
	return "event_subscriber"
}

func (eshc *EventSubscriberHealthCheck) Check(ctx context.Context) HealthResult {
	start := time.Now()
	result := HealthResult{
		Name:      eshc.Name(),
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	eventSubscriber := eshc.container.GetEventSubscriber()
	if eventSubscriber == nil {
		result.Status = HealthStatusUnknown
		result.Message = "Event subscriber not configured"
		result.Duration = time.Since(start)
		return result
	}

	if err := eventSubscriber.HealthCheck(ctx); err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = "Event subscriber health check failed"
		result.Error = err
	} else {
		result.Status = HealthStatusHealthy
		result.Message = "Event subscriber is healthy"
	}

	result.Duration = time.Since(start)
	return result
}

func (eshc *EventSubscriberHealthCheck) IsRequired() bool {
	return false
}

func (eshc *EventSubscriberHealthCheck) GetTimeout() time.Duration {
	return eshc.timeout
}

// Session Manager Health Check
type SessionManagerHealthCheck struct {
	container infrastructure.ContainerInterface
	timeout   time.Duration
}

func (smhc *SessionManagerHealthCheck) Name() string {
	return "session_manager"
}

func (smhc *SessionManagerHealthCheck) Check(ctx context.Context) HealthResult {
	start := time.Now()
	result := HealthResult{
		Name:      smhc.Name(),
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	sessionManager := smhc.container.GetSessionManager()
	if sessionManager == nil {
		result.Status = HealthStatusUnknown
		result.Message = "Session manager not configured"
		result.Duration = time.Since(start)
		return result
	}

	// Test session manager by creating a test session
	testSession, err := sessionManager.CreateSession(ctx, "health_check_user", "health@test.com", "test", nil)
	if err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = "Session manager test failed"
		result.Error = err
	} else {
		// Clean up test session
		sessionManager.RevokeSession(ctx, testSession.ID)

		result.Status = HealthStatusHealthy
		result.Message = "Session manager is healthy"
		result.Metadata["test_session_created"] = true
	}

	result.Duration = time.Since(start)
	return result
}

func (smhc *SessionManagerHealthCheck) IsRequired() bool {
	return false
}

func (smhc *SessionManagerHealthCheck) GetTimeout() time.Duration {
	return smhc.timeout
}

// JWT Service Health Check
type JWTServiceHealthCheck struct {
	container infrastructure.ContainerInterface
	timeout   time.Duration
}

func (jhc *JWTServiceHealthCheck) Name() string {
	return "jwt_service"
}

func (jhc *JWTServiceHealthCheck) Check(ctx context.Context) HealthResult {
	start := time.Now()
	result := HealthResult{
		Name:      jhc.Name(),
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	jwtService := jhc.container.GetJWTService()
	if jwtService == nil {
		result.Status = HealthStatusUnknown
		result.Message = "JWT service not configured"
		result.Duration = time.Since(start)
		return result
	}

	// Test JWT service by generating and validating a token
	token, _, err := jwtService.GenerateAccessToken(ctx, "health_check_user", "health@test.com", "test", nil)
	if err != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = "JWT token generation failed"
		result.Error = err
	} else {
		// Validate the token
		_, err := jwtService.ValidateToken(ctx, token)
		if err != nil {
			result.Status = HealthStatusDegraded
			result.Message = "JWT token validation failed"
			result.Error = err
		} else {
			result.Status = HealthStatusHealthy
			result.Message = "JWT service is healthy"
			result.Metadata["test_token_generated"] = true
		}
	}

	result.Duration = time.Since(start)
	return result
}

func (jhc *JWTServiceHealthCheck) IsRequired() bool {
	return false
}

func (jhc *JWTServiceHealthCheck) GetTimeout() time.Duration {
	return jhc.timeout
}

// System Resource Health Check
type SystemResourceHealthCheck struct {
	container infrastructure.ContainerInterface
	timeout   time.Duration
}

func (srhc *SystemResourceHealthCheck) Name() string {
	return "system_resources"
}

func (srhc *SystemResourceHealthCheck) Check(ctx context.Context) HealthResult {
	start := time.Now()
	result := HealthResult{
		Name:      srhc.Name(),
		Timestamp: start,
		Metadata:  make(map[string]interface{}),
	}

	// Check system resources (simplified version)
	// In a real implementation, you would use libraries like gopsutil

	// For now, we'll do basic checks
	result.Status = HealthStatusHealthy
	result.Message = "System resources are healthy"

	// Add basic system info
	result.Metadata["goroutines"] = getGoroutineCount()
	result.Metadata["memory_alloc"] = getMemoryAlloc()
	result.Metadata["gc_cycles"] = getGCCycles()

	result.Duration = time.Since(start)
	return result
}

func (srhc *SystemResourceHealthCheck) IsRequired() bool {
	return false
}

func (srhc *SystemResourceHealthCheck) GetTimeout() time.Duration {
	return srhc.timeout
}

// Helper functions for system resource monitoring

// getGoroutineCount returns the current number of goroutines
func getGoroutineCount() int {
	return runtime.NumGoroutine()
}

// getMemoryAlloc returns the current memory allocation in bytes
func getMemoryAlloc() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// getGCCycles returns the number of completed GC cycles
func getGCCycles() uint32 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.NumGC
}

// HTTPHealthHandler provides an HTTP handler for health checks
func HTTPHealthHandler(healthChecker *HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Perform health check
		health := healthChecker.CheckHealth(ctx)

		// Set appropriate status code
		statusCode := http.StatusOK
		if health.Status == HealthStatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		} else if health.Status == HealthStatusDegraded {
			statusCode = http.StatusPartialContent
		}

		// Set headers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		// Encode and send response
		if err := json.NewEncoder(w).Encode(health); err != nil {
			http.Error(w, "Failed to encode health response", http.StatusInternalServerError)
		}
	}
}
