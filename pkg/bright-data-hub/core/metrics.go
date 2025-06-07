package core

import (
	"sync"
	"time"
)

// MetricsCollector collects and manages metrics
type MetricsCollector struct {
	enabled bool
	metrics *Metrics
	mu      sync.RWMutex
}

// Metrics represents collected metrics
type Metrics struct {
	Requests    map[string]*RequestMetrics `json:"requests"`
	Cache       *CacheMetrics              `json:"cache"`
	RateLimit   *RateLimitMetrics          `json:"rate_limit"`
	Errors      map[string]int64           `json:"errors"`
	StartTime   time.Time                  `json:"start_time"`
	LastUpdated time.Time                  `json:"last_updated"`
}

// RequestMetrics represents metrics for a specific request type
type RequestMetrics struct {
	Count       int64         `json:"count"`
	Successes   int64         `json:"successes"`
	Failures    int64         `json:"failures"`
	TotalTime   time.Duration `json:"total_time"`
	AvgLatency  time.Duration `json:"avg_latency"`
	MinLatency  time.Duration `json:"min_latency"`
	MaxLatency  time.Duration `json:"max_latency"`
	LastRequest time.Time     `json:"last_request"`
}

// CacheMetrics represents cache performance metrics
type CacheMetrics struct {
	Hits        map[string]int64 `json:"hits"`
	Misses      map[string]int64 `json:"misses"`
	TotalHits   int64            `json:"total_hits"`
	TotalMisses int64            `json:"total_misses"`
	HitRatio    float64          `json:"hit_ratio"`
}

// RateLimitMetrics represents rate limiting metrics
type RateLimitMetrics struct {
	Allowed  int64 `json:"allowed"`
	Rejected int64 `json:"rejected"`
	Total    int64 `json:"total"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(enabled bool) *MetricsCollector {
	return &MetricsCollector{
		enabled: enabled,
		metrics: &Metrics{
			Requests:    make(map[string]*RequestMetrics),
			Cache:       &CacheMetrics{
				Hits:   make(map[string]int64),
				Misses: make(map[string]int64),
			},
			RateLimit:   &RateLimitMetrics{},
			Errors:      make(map[string]int64),
			StartTime:   time.Now(),
			LastUpdated: time.Now(),
		},
	}
}

// IncrementRequests increments request count for a method
func (mc *MetricsCollector) IncrementRequests(method string) {
	if !mc.enabled {
		return
	}
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if mc.metrics.Requests[method] == nil {
		mc.metrics.Requests[method] = &RequestMetrics{
			MinLatency: time.Hour, // Initialize with high value
		}
	}
	
	mc.metrics.Requests[method].Count++
	mc.metrics.Requests[method].LastRequest = time.Now()
	mc.metrics.LastUpdated = time.Now()
}

// IncrementSuccesses increments success count for a method
func (mc *MetricsCollector) IncrementSuccesses(method string) {
	if !mc.enabled {
		return
	}
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if mc.metrics.Requests[method] == nil {
		mc.metrics.Requests[method] = &RequestMetrics{}
	}
	
	mc.metrics.Requests[method].Successes++
	mc.metrics.LastUpdated = time.Now()
}

// IncrementErrors increments error count
func (mc *MetricsCollector) IncrementErrors(method, errorType string) {
	if !mc.enabled {
		return
	}
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if mc.metrics.Requests[method] == nil {
		mc.metrics.Requests[method] = &RequestMetrics{}
	}
	
	mc.metrics.Requests[method].Failures++
	
	key := method + ":" + errorType
	mc.metrics.Errors[key]++
	mc.metrics.LastUpdated = time.Now()
}

// RecordLatency records latency for a method
func (mc *MetricsCollector) RecordLatency(method string, latency time.Duration) {
	if !mc.enabled {
		return
	}
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if mc.metrics.Requests[method] == nil {
		mc.metrics.Requests[method] = &RequestMetrics{
			MinLatency: latency,
		}
	}
	
	req := mc.metrics.Requests[method]
	req.TotalTime += latency
	
	if req.Count > 0 {
		req.AvgLatency = req.TotalTime / time.Duration(req.Count)
	}
	
	if latency < req.MinLatency {
		req.MinLatency = latency
	}
	
	if latency > req.MaxLatency {
		req.MaxLatency = latency
	}
	
	mc.metrics.LastUpdated = time.Now()
}

// IncrementCacheHits increments cache hits for a method
func (mc *MetricsCollector) IncrementCacheHits(method string) {
	if !mc.enabled {
		return
	}
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.metrics.Cache.Hits[method]++
	mc.metrics.Cache.TotalHits++
	mc.updateCacheHitRatio()
	mc.metrics.LastUpdated = time.Now()
}

// IncrementCacheMisses increments cache misses for a method
func (mc *MetricsCollector) IncrementCacheMisses(method string) {
	if !mc.enabled {
		return
	}
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.metrics.Cache.Misses[method]++
	mc.metrics.Cache.TotalMisses++
	mc.updateCacheHitRatio()
	mc.metrics.LastUpdated = time.Now()
}

// updateCacheHitRatio updates the cache hit ratio
func (mc *MetricsCollector) updateCacheHitRatio() {
	total := mc.metrics.Cache.TotalHits + mc.metrics.Cache.TotalMisses
	if total > 0 {
		mc.metrics.Cache.HitRatio = float64(mc.metrics.Cache.TotalHits) / float64(total)
	}
}

// IncrementRateLimitAllowed increments allowed rate limit counter
func (mc *MetricsCollector) IncrementRateLimitAllowed() {
	if !mc.enabled {
		return
	}
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.metrics.RateLimit.Allowed++
	mc.metrics.RateLimit.Total++
	mc.metrics.LastUpdated = time.Now()
}

// IncrementRateLimitRejected increments rejected rate limit counter
func (mc *MetricsCollector) IncrementRateLimitRejected() {
	if !mc.enabled {
		return
	}
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.metrics.RateLimit.Rejected++
	mc.metrics.RateLimit.Total++
	mc.metrics.LastUpdated = time.Now()
}

// GetMetrics returns a copy of current metrics
func (mc *MetricsCollector) GetMetrics() *Metrics {
	if !mc.enabled {
		return nil
	}
	
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	// Create a deep copy
	metricsCopy := &Metrics{
		Requests:    make(map[string]*RequestMetrics),
		Cache:       &CacheMetrics{
			Hits:        make(map[string]int64),
			Misses:      make(map[string]int64),
			TotalHits:   mc.metrics.Cache.TotalHits,
			TotalMisses: mc.metrics.Cache.TotalMisses,
			HitRatio:    mc.metrics.Cache.HitRatio,
		},
		RateLimit:   &RateLimitMetrics{
			Allowed:  mc.metrics.RateLimit.Allowed,
			Rejected: mc.metrics.RateLimit.Rejected,
			Total:    mc.metrics.RateLimit.Total,
		},
		Errors:      make(map[string]int64),
		StartTime:   mc.metrics.StartTime,
		LastUpdated: mc.metrics.LastUpdated,
	}
	
	// Copy request metrics
	for method, req := range mc.metrics.Requests {
		metricsCopy.Requests[method] = &RequestMetrics{
			Count:       req.Count,
			Successes:   req.Successes,
			Failures:    req.Failures,
			TotalTime:   req.TotalTime,
			AvgLatency:  req.AvgLatency,
			MinLatency:  req.MinLatency,
			MaxLatency:  req.MaxLatency,
			LastRequest: req.LastRequest,
		}
	}
	
	// Copy cache metrics
	for method, hits := range mc.metrics.Cache.Hits {
		metricsCopy.Cache.Hits[method] = hits
	}
	for method, misses := range mc.metrics.Cache.Misses {
		metricsCopy.Cache.Misses[method] = misses
	}
	
	// Copy error metrics
	for errorType, count := range mc.metrics.Errors {
		metricsCopy.Errors[errorType] = count
	}
	
	return metricsCopy
}

// Reset resets all metrics
func (mc *MetricsCollector) Reset() {
	if !mc.enabled {
		return
	}
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.metrics = &Metrics{
		Requests:    make(map[string]*RequestMetrics),
		Cache:       &CacheMetrics{
			Hits:   make(map[string]int64),
			Misses: make(map[string]int64),
		},
		RateLimit:   &RateLimitMetrics{},
		Errors:      make(map[string]int64),
		StartTime:   time.Now(),
		LastUpdated: time.Now(),
	}
}
