package health

import (
	"context"
	"fmt"
	"sync"
	"time"

)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusUnknown   HealthStatus = "unknown"
)

// CheckResult represents the result of a health check
type CheckResult struct {
	Name        string                 `json:"name"`
	Status      HealthStatus           `json:"status"`
	Message     string                 `json:"message,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Timestamp   time.Time              `json:"timestamp"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

// HealthCheck defines a health check interface
type HealthCheck interface {
	// Name returns the name of the health check
	Name() string
	
	// Check performs the health check
	Check(ctx context.Context) CheckResult
	
	// Timeout returns the timeout for this health check
	Timeout() time.Duration
	
	// Critical indicates if this check is critical for overall health
	Critical() bool
}

// HealthChecker manages and executes health checks
type HealthChecker struct {
	checks   map[string]HealthCheck
	results  map[string]CheckResult
	config   HealthConfig
	mutex    sync.RWMutex
	stopCh   chan struct{}
	running  bool
}

// HealthConfig holds health checker configuration
type HealthConfig struct {
	// Interval for periodic health checks
	CheckInterval time.Duration
	
	// Timeout for individual health checks
	DefaultTimeout time.Duration
	
	// Whether to run checks in parallel
	Parallel bool
	
	// Maximum number of concurrent checks
	MaxConcurrent int
	
	// Whether to cache results
	CacheResults bool
	
	// Cache TTL for results
	CacheTTL time.Duration
	
	// Callbacks
	OnHealthChange func(name string, oldStatus, newStatus HealthStatus)
	OnCheckFailure func(name string, err error)
}

// DefaultHealthConfig returns a default health configuration
func DefaultHealthConfig() HealthConfig {
	return HealthConfig{
		CheckInterval:  30 * time.Second,
		DefaultTimeout: 10 * time.Second,
		Parallel:       true,
		MaxConcurrent:  10,
		CacheResults:   true,
		CacheTTL:       60 * time.Second,
	}
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(config HealthConfig) *HealthChecker {
	if config.CheckInterval <= 0 {
		config.CheckInterval = 30 * time.Second
	}
	if config.DefaultTimeout <= 0 {
		config.DefaultTimeout = 10 * time.Second
	}
	if config.MaxConcurrent <= 0 {
		config.MaxConcurrent = 10
	}
	if config.CacheTTL <= 0 {
		config.CacheTTL = 60 * time.Second
	}

	return &HealthChecker{
		checks:  make(map[string]HealthCheck),
		results: make(map[string]CheckResult),
		config:  config,
		stopCh:  make(chan struct{}),
	}
}

// RegisterCheck registers a health check
func (hc *HealthChecker) RegisterCheck(check HealthCheck) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	hc.checks[check.Name()] = check
}

// UnregisterCheck unregisters a health check
func (hc *HealthChecker) UnregisterCheck(name string) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	delete(hc.checks, name)
	delete(hc.results, name)
}

// CheckAll performs all registered health checks
func (hc *HealthChecker) CheckAll(ctx context.Context) map[string]CheckResult {
	hc.mutex.RLock()
	checks := make(map[string]HealthCheck)
	for name, check := range hc.checks {
		checks[name] = check
	}
	hc.mutex.RUnlock()

	results := make(map[string]CheckResult)

	if hc.config.Parallel {
		results = hc.runChecksParallel(ctx, checks)
	} else {
		results = hc.runChecksSequential(ctx, checks)
	}

	// Update cached results
	if hc.config.CacheResults {
		hc.mutex.Lock()
		for name, result := range results {
			oldResult, exists := hc.results[name]
			hc.results[name] = result
			
			// Trigger callback if status changed
			if exists && hc.config.OnHealthChange != nil && oldResult.Status != result.Status {
				hc.config.OnHealthChange(name, oldResult.Status, result.Status)
			}
		}
		hc.mutex.Unlock()
	}

	return results
}

// runChecksParallel runs health checks in parallel
func (hc *HealthChecker) runChecksParallel(ctx context.Context, checks map[string]HealthCheck) map[string]CheckResult {
	results := make(map[string]CheckResult)
	resultsCh := make(chan CheckResult, len(checks))
	semaphore := make(chan struct{}, hc.config.MaxConcurrent)

	// Start all checks
	var wg sync.WaitGroup
	for _, check := range checks {
		wg.Add(1)
		go func(check HealthCheck) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			result := hc.runSingleCheck(ctx, check)
			resultsCh <- result
		}(check)
	}

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect results
	for result := range resultsCh {
		results[result.Name] = result
	}

	return results
}

// runChecksSequential runs health checks sequentially
func (hc *HealthChecker) runChecksSequential(ctx context.Context, checks map[string]HealthCheck) map[string]CheckResult {
	results := make(map[string]CheckResult)

	for _, check := range checks {
		result := hc.runSingleCheck(ctx, check)
		results[result.Name] = result
	}

	return results
}

// runSingleCheck runs a single health check
func (hc *HealthChecker) runSingleCheck(ctx context.Context, check HealthCheck) CheckResult {
	timeout := check.Timeout()
	if timeout <= 0 {
		timeout = hc.config.DefaultTimeout
	}

	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	
	// Run the check in a goroutine to handle panics
	resultCh := make(chan CheckResult, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				result := CheckResult{
					Name:      check.Name(),
					Status:    StatusUnhealthy,
					Message:   "Health check panicked",
					Error:     fmt.Sprintf("panic: %v", r),
					Duration:  time.Since(start),
					Timestamp: time.Now(),
				}
				resultCh <- result
			}
		}()
		
		result := check.Check(checkCtx)
		result.Duration = time.Since(start)
		result.Timestamp = time.Now()
		resultCh <- result
	}()

	// Wait for result or timeout
	select {
	case result := <-resultCh:
		return result
	case <-checkCtx.Done():
		return CheckResult{
			Name:      check.Name(),
			Status:    StatusUnhealthy,
			Message:   "Health check timed out",
			Error:     checkCtx.Err().Error(),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	}
}

// GetResult returns the cached result for a specific check
func (hc *HealthChecker) GetResult(name string) (CheckResult, bool) {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	
	result, exists := hc.results[name]
	if !exists {
		return CheckResult{}, false
	}

	// Check if result is still valid
	if hc.config.CacheResults && time.Since(result.Timestamp) > hc.config.CacheTTL {
		return CheckResult{}, false
	}

	return result, true
}

// GetOverallStatus returns the overall health status
func (hc *HealthChecker) GetOverallStatus() HealthStatus {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	if len(hc.results) == 0 {
		return StatusUnknown
	}

	hasUnhealthy := false
	hasDegraded := false

	for _, result := range hc.results {
		switch result.Status {
		case StatusUnhealthy:
			hasUnhealthy = true
		case StatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return StatusUnhealthy
	}
	if hasDegraded {
		return StatusDegraded
	}
	return StatusHealthy
}

// Start starts the periodic health checker
func (hc *HealthChecker) Start(ctx context.Context) {
	hc.mutex.Lock()
	if hc.running {
		hc.mutex.Unlock()
		return
	}
	hc.running = true
	hc.mutex.Unlock()

	ticker := time.NewTicker(hc.config.CheckInterval)
	defer ticker.Stop()

	// Run initial check
	hc.CheckAll(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-hc.stopCh:
			return
		case <-ticker.C:
			hc.CheckAll(ctx)
		}
	}
}

// Stop stops the periodic health checker
func (hc *HealthChecker) Stop() {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	if hc.running {
		close(hc.stopCh)
		hc.running = false
	}
}

// IsRunning returns whether the health checker is running
func (hc *HealthChecker) IsRunning() bool {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	return hc.running
}

// SimpleHealthCheck implements a basic health check
type SimpleHealthCheck struct {
	name     string
	checkFn  func(ctx context.Context) error
	timeout  time.Duration
	critical bool
}

// NewSimpleHealthCheck creates a new simple health check
func NewSimpleHealthCheck(name string, checkFn func(ctx context.Context) error, timeout time.Duration, critical bool) *SimpleHealthCheck {
	return &SimpleHealthCheck{
		name:     name,
		checkFn:  checkFn,
		timeout:  timeout,
		critical: critical,
	}
}

// Name returns the check name
func (shc *SimpleHealthCheck) Name() string {
	return shc.name
}

// Check performs the health check
func (shc *SimpleHealthCheck) Check(ctx context.Context) CheckResult {
	err := shc.checkFn(ctx)
	
	result := CheckResult{
		Name: shc.name,
	}

	if err == nil {
		result.Status = StatusHealthy
		result.Message = "Check passed"
	} else {
		result.Status = StatusUnhealthy
		result.Message = "Check failed"
		result.Error = err.Error()
	}

	return result
}

// Timeout returns the check timeout
func (shc *SimpleHealthCheck) Timeout() time.Duration {
	return shc.timeout
}

// Critical returns whether this check is critical
func (shc *SimpleHealthCheck) Critical() bool {
	return shc.critical
}

// DatabaseHealthCheck implements a database health check
type DatabaseHealthCheck struct {
	name     string
	pingFn   func(ctx context.Context) error
	timeout  time.Duration
}

// NewDatabaseHealthCheck creates a new database health check
func NewDatabaseHealthCheck(name string, pingFn func(ctx context.Context) error, timeout time.Duration) *DatabaseHealthCheck {
	return &DatabaseHealthCheck{
		name:    name,
		pingFn:  pingFn,
		timeout: timeout,
	}
}

// Name returns the check name
func (dhc *DatabaseHealthCheck) Name() string {
	return dhc.name
}

// Check performs the database health check
func (dhc *DatabaseHealthCheck) Check(ctx context.Context) CheckResult {
	err := dhc.pingFn(ctx)
	
	result := CheckResult{
		Name: dhc.name,
	}

	if err == nil {
		result.Status = StatusHealthy
		result.Message = "Database connection healthy"
	} else {
		result.Status = StatusUnhealthy
		result.Message = "Database connection failed"
		result.Error = err.Error()
	}

	return result
}

// Timeout returns the check timeout
func (dhc *DatabaseHealthCheck) Timeout() time.Duration {
	return dhc.timeout
}

// Critical returns whether this check is critical
func (dhc *DatabaseHealthCheck) Critical() bool {
	return true // Database is typically critical
}

// HTTPHealthCheck implements an HTTP endpoint health check
type HTTPHealthCheck struct {
	name     string
	url      string
	client   HTTPClient
	timeout  time.Duration
	critical bool
}

// HTTPClient interface for HTTP requests
type HTTPClient interface {
	Get(ctx context.Context, url string) error
}

// NewHTTPHealthCheck creates a new HTTP health check
func NewHTTPHealthCheck(name, url string, client HTTPClient, timeout time.Duration, critical bool) *HTTPHealthCheck {
	return &HTTPHealthCheck{
		name:     name,
		url:      url,
		client:   client,
		timeout:  timeout,
		critical: critical,
	}
}

// Name returns the check name
func (hhc *HTTPHealthCheck) Name() string {
	return hhc.name
}

// Check performs the HTTP health check
func (hhc *HTTPHealthCheck) Check(ctx context.Context) CheckResult {
	err := hhc.client.Get(ctx, hhc.url)
	
	result := CheckResult{
		Name: hhc.name,
		Details: map[string]interface{}{
			"url": hhc.url,
		},
	}

	if err == nil {
		result.Status = StatusHealthy
		result.Message = "HTTP endpoint healthy"
	} else {
		result.Status = StatusUnhealthy
		result.Message = "HTTP endpoint failed"
		result.Error = err.Error()
	}

	return result
}

// Timeout returns the check timeout
func (hhc *HTTPHealthCheck) Timeout() time.Duration {
	return hhc.timeout
}

// Critical returns whether this check is critical
func (hhc *HTTPHealthCheck) Critical() bool {
	return hhc.critical
}

// Global health checker
var globalHealthChecker *HealthChecker

// InitGlobalHealthChecker initializes the global health checker
func InitGlobalHealthChecker(config HealthConfig) {
	globalHealthChecker = NewHealthChecker(config)
}

// GetGlobalHealthChecker returns the global health checker
func GetGlobalHealthChecker() *HealthChecker {
	if globalHealthChecker == nil {
		globalHealthChecker = NewHealthChecker(DefaultHealthConfig())
	}
	return globalHealthChecker
}

// RegisterGlobalCheck registers a check with the global health checker
func RegisterGlobalCheck(check HealthCheck) {
	GetGlobalHealthChecker().RegisterCheck(check)
}

// CheckGlobalHealth performs all global health checks
func CheckGlobalHealth(ctx context.Context) map[string]CheckResult {
	return GetGlobalHealthChecker().CheckAll(ctx)
}
