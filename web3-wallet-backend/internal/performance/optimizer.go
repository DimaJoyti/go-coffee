package performance

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// PerformanceOptimizer optimizes system performance
type PerformanceOptimizer struct {
	logger           *zap.Logger
	metrics          *PerformanceMetrics
	config           OptimizerConfig
	connectionPools  map[string]*ConnectionPool
	cacheManager     *CacheManager
	rateLimiter      *RateLimiter
	circuitBreaker   *CircuitBreaker
	running          bool
	stopChan         chan struct{}
	wg               sync.WaitGroup
	mutex            sync.RWMutex
}

// OptimizerConfig contains optimizer configuration
type OptimizerConfig struct {
	MaxConnections       int           `json:"max_connections"`
	ConnectionTimeout    time.Duration `json:"connection_timeout"`
	IdleTimeout          time.Duration `json:"idle_timeout"`
	MaxRetries           int           `json:"max_retries"`
	RetryDelay           time.Duration `json:"retry_delay"`
	CacheSize            int           `json:"cache_size"`
	CacheTTL             time.Duration `json:"cache_ttl"`
	RateLimitRPS         int           `json:"rate_limit_rps"`
	CircuitBreakerThreshold int        `json:"circuit_breaker_threshold"`
	MetricsInterval      time.Duration `json:"metrics_interval"`
}

// PerformanceMetrics tracks performance metrics
type PerformanceMetrics struct {
	RequestCount        int64           `json:"request_count"`
	SuccessCount        int64           `json:"success_count"`
	ErrorCount          int64           `json:"error_count"`
	AverageLatency      time.Duration   `json:"average_latency"`
	P95Latency          time.Duration   `json:"p95_latency"`
	P99Latency          time.Duration   `json:"p99_latency"`
	ThroughputRPS       decimal.Decimal `json:"throughput_rps"`
	CacheHitRate        decimal.Decimal `json:"cache_hit_rate"`
	ConnectionPoolUsage decimal.Decimal `json:"connection_pool_usage"`
	MemoryUsage         uint64          `json:"memory_usage"`
	CPUUsage            decimal.Decimal `json:"cpu_usage"`
	GoroutineCount      int             `json:"goroutine_count"`
	LastUpdate          time.Time       `json:"last_update"`
	latencies           []time.Duration
	mutex               sync.RWMutex
}

// ConnectionPool manages database/API connections
type ConnectionPool struct {
	name         string
	maxSize      int
	currentSize  int
	activeConns  int
	idleConns    chan interface{}
	busyConns    map[string]interface{}
	timeout      time.Duration
	idleTimeout  time.Duration
	mutex        sync.RWMutex
	logger       *zap.Logger
}

// CacheManager manages caching
type CacheManager struct {
	cache     map[string]*CacheEntry
	maxSize   int
	ttl       time.Duration
	hits      int64
	misses    int64
	mutex     sync.RWMutex
	logger    *zap.Logger
}

// CacheEntry represents a cache entry
type CacheEntry struct {
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
	AccessCount int64     `json:"access_count"`
	CreatedAt time.Time   `json:"created_at"`
}

// RateLimiter implements rate limiting
type RateLimiter struct {
	rps       int
	tokens    chan struct{}
	ticker    *time.Ticker
	mutex     sync.Mutex
	logger    *zap.Logger
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	name           string
	threshold      int
	timeout        time.Duration
	state          CircuitState
	failureCount   int
	lastFailure    time.Time
	successCount   int
	mutex          sync.RWMutex
	logger         *zap.Logger
}

// CircuitState represents circuit breaker state
type CircuitState string

const (
	CircuitClosed   CircuitState = "closed"
	CircuitOpen     CircuitState = "open"
	CircuitHalfOpen CircuitState = "half_open"
)

// NewPerformanceOptimizer creates a new performance optimizer
func NewPerformanceOptimizer(logger *zap.Logger, config OptimizerConfig) *PerformanceOptimizer {
	return &PerformanceOptimizer{
		logger:          logger,
		metrics:         NewPerformanceMetrics(),
		config:          config,
		connectionPools: make(map[string]*ConnectionPool),
		cacheManager:    NewCacheManager(config.CacheSize, config.CacheTTL, logger),
		rateLimiter:     NewRateLimiter(config.RateLimitRPS, logger),
		circuitBreaker:  NewCircuitBreaker("main", config.CircuitBreakerThreshold, time.Minute*5, logger),
		stopChan:        make(chan struct{}),
	}
}

// NewPerformanceMetrics creates new performance metrics
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		latencies:  make([]time.Duration, 0, 1000),
		LastUpdate: time.Now(),
	}
}

// Start starts the performance optimizer
func (po *PerformanceOptimizer) Start(ctx context.Context) error {
	po.mutex.Lock()
	defer po.mutex.Unlock()

	if po.running {
		return nil
	}

	po.running = true
	po.wg.Add(1)

	go po.metricsLoop(ctx)

	po.logger.Info("Performance optimizer started")
	return nil
}

// Stop stops the performance optimizer
func (po *PerformanceOptimizer) Stop() error {
	po.mutex.Lock()
	defer po.mutex.Unlock()

	if !po.running {
		return nil
	}

	close(po.stopChan)
	po.wg.Wait()
	po.running = false

	po.logger.Info("Performance optimizer stopped")
	return nil
}

// metricsLoop collects performance metrics
func (po *PerformanceOptimizer) metricsLoop(ctx context.Context) {
	defer po.wg.Done()

	ticker := time.NewTicker(po.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-po.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			po.collectMetrics()
		}
	}
}

// collectMetrics collects system metrics
func (po *PerformanceOptimizer) collectMetrics() {
	po.metrics.mutex.Lock()
	defer po.metrics.mutex.Unlock()

	// Collect memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	po.metrics.MemoryUsage = memStats.Alloc

	// Collect goroutine count
	po.metrics.GoroutineCount = runtime.NumGoroutine()

	// Calculate cache hit rate
	if po.cacheManager.hits+po.cacheManager.misses > 0 {
		hitRate := float64(po.cacheManager.hits) / float64(po.cacheManager.hits+po.cacheManager.misses)
		po.metrics.CacheHitRate = decimal.NewFromFloat(hitRate)
	}

	// Calculate latency percentiles
	if len(po.metrics.latencies) > 0 {
		po.calculateLatencyPercentiles()
	}

	// Calculate throughput
	if po.metrics.RequestCount > 0 {
		duration := time.Since(po.metrics.LastUpdate)
		if duration > 0 {
			rps := float64(po.metrics.RequestCount) / duration.Seconds()
			po.metrics.ThroughputRPS = decimal.NewFromFloat(rps)
		}
	}

	po.metrics.LastUpdate = time.Now()

	po.logger.Debug("Performance metrics collected",
		zap.Uint64("memory_usage", po.metrics.MemoryUsage),
		zap.Int("goroutines", po.metrics.GoroutineCount),
		zap.String("cache_hit_rate", po.metrics.CacheHitRate.String()),
		zap.String("throughput_rps", po.metrics.ThroughputRPS.String()),
	)
}

// calculateLatencyPercentiles calculates latency percentiles
func (po *PerformanceOptimizer) calculateLatencyPercentiles() {
	latencies := make([]time.Duration, len(po.metrics.latencies))
	copy(latencies, po.metrics.latencies)

	// Sort latencies (simplified bubble sort for demo)
	for i := 0; i < len(latencies); i++ {
		for j := i + 1; j < len(latencies); j++ {
			if latencies[i] > latencies[j] {
				latencies[i], latencies[j] = latencies[j], latencies[i]
			}
		}
	}

	// Calculate percentiles
	if len(latencies) > 0 {
		p95Index := int(float64(len(latencies)) * 0.95)
		p99Index := int(float64(len(latencies)) * 0.99)

		if p95Index >= len(latencies) {
			p95Index = len(latencies) - 1
		}
		if p99Index >= len(latencies) {
			p99Index = len(latencies) - 1
		}

		po.metrics.P95Latency = latencies[p95Index]
		po.metrics.P99Latency = latencies[p99Index]

		// Calculate average
		var total time.Duration
		for _, latency := range latencies {
			total += latency
		}
		po.metrics.AverageLatency = total / time.Duration(len(latencies))
	}
}

// RecordLatency records a request latency
func (po *PerformanceOptimizer) RecordLatency(latency time.Duration) {
	po.metrics.mutex.Lock()
	defer po.metrics.mutex.Unlock()

	po.metrics.RequestCount++
	po.metrics.latencies = append(po.metrics.latencies, latency)

	// Keep only last 1000 latencies
	if len(po.metrics.latencies) > 1000 {
		po.metrics.latencies = po.metrics.latencies[len(po.metrics.latencies)-1000:]
	}
}

// RecordSuccess records a successful operation
func (po *PerformanceOptimizer) RecordSuccess() {
	po.metrics.mutex.Lock()
	defer po.metrics.mutex.Unlock()

	po.metrics.SuccessCount++
}

// RecordError records an error
func (po *PerformanceOptimizer) RecordError() {
	po.metrics.mutex.Lock()
	defer po.metrics.mutex.Unlock()

	po.metrics.ErrorCount++
}

// GetMetrics returns current performance metrics
func (po *PerformanceOptimizer) GetMetrics() PerformanceMetrics {
	po.metrics.mutex.RLock()
	defer po.metrics.mutex.RUnlock()

	// Return a copy
	return PerformanceMetrics{
		RequestCount:        po.metrics.RequestCount,
		SuccessCount:        po.metrics.SuccessCount,
		ErrorCount:          po.metrics.ErrorCount,
		AverageLatency:      po.metrics.AverageLatency,
		P95Latency:          po.metrics.P95Latency,
		P99Latency:          po.metrics.P99Latency,
		ThroughputRPS:       po.metrics.ThroughputRPS,
		CacheHitRate:        po.metrics.CacheHitRate,
		ConnectionPoolUsage: po.metrics.ConnectionPoolUsage,
		MemoryUsage:         po.metrics.MemoryUsage,
		CPUUsage:            po.metrics.CPUUsage,
		GoroutineCount:      po.metrics.GoroutineCount,
		LastUpdate:          po.metrics.LastUpdate,
	}
}

// NewCacheManager creates a new cache manager
func NewCacheManager(maxSize int, ttl time.Duration, logger *zap.Logger) *CacheManager {
	return &CacheManager{
		cache:   make(map[string]*CacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
		logger:  logger,
	}
}

// Get retrieves a value from cache
func (cm *CacheManager) Get(key string) (interface{}, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	entry, exists := cm.cache[key]
	if !exists {
		cm.misses++
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		delete(cm.cache, key)
		cm.misses++
		return nil, false
	}

	entry.AccessCount++
	cm.hits++
	return entry.Value, true
}

// Set stores a value in cache
func (cm *CacheManager) Set(key string, value interface{}) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Remove oldest entries if cache is full
	if len(cm.cache) >= cm.maxSize {
		cm.evictOldest()
	}

	cm.cache[key] = &CacheEntry{
		Value:       value,
		ExpiresAt:   time.Now().Add(cm.ttl),
		AccessCount: 0,
		CreatedAt:   time.Now(),
	}
}

// evictOldest removes the oldest cache entry
func (cm *CacheManager) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range cm.cache {
		if oldestKey == "" || entry.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(cm.cache, oldestKey)
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rps int, logger *zap.Logger) *RateLimiter {
	rl := &RateLimiter{
		rps:    rps,
		tokens: make(chan struct{}, rps),
		logger: logger,
	}

	// Fill initial tokens
	for i := 0; i < rps; i++ {
		rl.tokens <- struct{}{}
	}

	// Start token refill
	rl.ticker = time.NewTicker(time.Second / time.Duration(rps))
	go rl.refillTokens()

	return rl
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// refillTokens refills the token bucket
func (rl *RateLimiter) refillTokens() {
	for range rl.ticker.C {
		select {
		case rl.tokens <- struct{}{}:
		default:
			// Bucket is full
		}
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, threshold int, timeout time.Duration, logger *zap.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		name:      name,
		threshold: threshold,
		timeout:   timeout,
		state:     CircuitClosed,
		logger:    logger,
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if cb.state == CircuitOpen {
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.state = CircuitHalfOpen
			cb.logger.Info("Circuit breaker half-open", zap.String("name", cb.name))
		} else {
			return fmt.Errorf("circuit breaker open")
		}
	}

	err := fn()
	if err != nil {
		cb.failureCount++
		cb.lastFailure = time.Now()

		if cb.failureCount >= cb.threshold {
			cb.state = CircuitOpen
			cb.logger.Warn("Circuit breaker opened", 
				zap.String("name", cb.name),
				zap.Int("failures", cb.failureCount))
		}
		return err
	}

	// Success
	cb.successCount++
	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
		cb.failureCount = 0
		cb.logger.Info("Circuit breaker closed", zap.String("name", cb.name))
	}

	return nil
}
