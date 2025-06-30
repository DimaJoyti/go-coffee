package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go-coffee-ai-agents/internal/ai/providers"
)

// MetricsCollector collects and aggregates AI provider metrics
type MetricsCollector struct {
	metrics map[string]*ProviderMetrics
	mutex   sync.RWMutex
	
	// Configuration
	config *MetricsConfig
	
	// Callbacks
	onMetricUpdate func(provider string, metrics *ProviderMetrics)
	onAlert        func(alert Alert)
}

// MetricsConfig holds metrics collection configuration
type MetricsConfig struct {
	Enabled           bool          `yaml:"enabled" json:"enabled"`
	CollectionInterval time.Duration `yaml:"collection_interval" json:"collection_interval"`
	RetentionPeriod   time.Duration `yaml:"retention_period" json:"retention_period"`
	
	// Metric types to collect
	CollectLatency    bool          `yaml:"collect_latency" json:"collect_latency"`
	CollectCost       bool          `yaml:"collect_cost" json:"collect_cost"`
	CollectTokens     bool          `yaml:"collect_tokens" json:"collect_tokens"`
	CollectErrors     bool          `yaml:"collect_errors" json:"collect_errors"`
	CollectQuality    bool          `yaml:"collect_quality" json:"collect_quality"`
	
	// Alerting thresholds
	LatencyThreshold  time.Duration `yaml:"latency_threshold" json:"latency_threshold"`
	ErrorRateThreshold float64      `yaml:"error_rate_threshold" json:"error_rate_threshold"`
	CostThreshold     float64       `yaml:"cost_threshold" json:"cost_threshold"`
	
	// Aggregation settings
	WindowSize        time.Duration `yaml:"window_size" json:"window_size"`
	BucketCount       int           `yaml:"bucket_count" json:"bucket_count"`
}

// ProviderMetrics holds metrics for a specific provider
type ProviderMetrics struct {
	Provider      string            `json:"provider"`
	
	// Request metrics
	TotalRequests int64             `json:"total_requests"`
	SuccessfulRequests int64        `json:"successful_requests"`
	FailedRequests int64            `json:"failed_requests"`
	
	// Latency metrics
	AverageLatency time.Duration    `json:"average_latency"`
	MinLatency     time.Duration    `json:"min_latency"`
	MaxLatency     time.Duration    `json:"max_latency"`
	P50Latency     time.Duration    `json:"p50_latency"`
	P95Latency     time.Duration    `json:"p95_latency"`
	P99Latency     time.Duration    `json:"p99_latency"`
	
	// Token metrics
	TotalTokens    int64             `json:"total_tokens"`
	InputTokens    int64             `json:"input_tokens"`
	OutputTokens   int64             `json:"output_tokens"`
	AverageTokensPerRequest float64  `json:"average_tokens_per_request"`
	
	// Cost metrics
	TotalCost      float64           `json:"total_cost"`
	AverageCostPerRequest float64    `json:"average_cost_per_request"`
	AverageCostPerToken float64      `json:"average_cost_per_token"`
	
	// Error metrics
	ErrorRate      float64           `json:"error_rate"`
	ErrorsByType   map[string]int64  `json:"errors_by_type"`
	
	// Quality metrics
	AverageConfidence float64        `json:"average_confidence"`
	AverageRelevance  float64        `json:"average_relevance"`
	
	// Cache metrics
	CacheHitRate   float64           `json:"cache_hit_rate"`
	CacheHits      int64             `json:"cache_hits"`
	CacheMisses    int64             `json:"cache_misses"`
	
	// Rate limiting metrics
	RateLimitHits  int64             `json:"rate_limit_hits"`
	ThrottledRequests int64          `json:"throttled_requests"`
	
	// Time-based metrics
	HourlyMetrics  map[string]*HourlyMetrics  `json:"hourly_metrics"`
	DailyMetrics   map[string]*DailyMetrics   `json:"daily_metrics"`
	
	// Model-specific metrics
	ModelMetrics   map[string]*ModelMetrics   `json:"model_metrics"`
	
	// User-specific metrics
	UserMetrics    map[string]*UserMetrics    `json:"user_metrics"`
	
	// Last updated
	LastUpdated    time.Time         `json:"last_updated"`
	
	// Latency histogram
	LatencyHistogram *Histogram      `json:"latency_histogram"`
	
	// Cost histogram
	CostHistogram    *Histogram      `json:"cost_histogram"`
}

// HourlyMetrics holds metrics for a specific hour
type HourlyMetrics struct {
	Hour          string    `json:"hour"`
	Requests      int64     `json:"requests"`
	Errors        int64     `json:"errors"`
	Tokens        int64     `json:"tokens"`
	Cost          float64   `json:"cost"`
	AverageLatency time.Duration `json:"average_latency"`
	LastUpdated   time.Time `json:"last_updated"`
}

// DailyMetrics holds metrics for a specific day
type DailyMetrics struct {
	Date          string    `json:"date"`
	Requests      int64     `json:"requests"`
	Errors        int64     `json:"errors"`
	Tokens        int64     `json:"tokens"`
	Cost          float64   `json:"cost"`
	AverageLatency time.Duration `json:"average_latency"`
	LastUpdated   time.Time `json:"last_updated"`
}

// ModelMetrics holds metrics for a specific model
type ModelMetrics struct {
	Model         string    `json:"model"`
	Requests      int64     `json:"requests"`
	Errors        int64     `json:"errors"`
	InputTokens   int64     `json:"input_tokens"`
	OutputTokens  int64     `json:"output_tokens"`
	TotalTokens   int64     `json:"total_tokens"`
	Cost          float64   `json:"cost"`
	AverageLatency time.Duration `json:"average_latency"`
	LastUsed      time.Time `json:"last_used"`
}

// UserMetrics holds metrics for a specific user
type UserMetrics struct {
	UserID        string    `json:"user_id"`
	Requests      int64     `json:"requests"`
	Errors        int64     `json:"errors"`
	Tokens        int64     `json:"tokens"`
	Cost          float64   `json:"cost"`
	AverageLatency time.Duration `json:"average_latency"`
	LastUsed      time.Time `json:"last_used"`
}

// Histogram represents a histogram of values
type Histogram struct {
	Buckets       []HistogramBucket `json:"buckets"`
	Count         int64             `json:"count"`
	Sum           float64           `json:"sum"`
	Min           float64           `json:"min"`
	Max           float64           `json:"max"`
}

// HistogramBucket represents a bucket in a histogram
type HistogramBucket struct {
	UpperBound    float64           `json:"upper_bound"`
	Count         int64             `json:"count"`
	CumulativeCount int64           `json:"cumulative_count"`
}

// Alert represents a monitoring alert
type Alert struct {
	ID            string            `json:"id"`
	Provider      string            `json:"provider"`
	Type          AlertType         `json:"type"`
	Severity      AlertSeverity     `json:"severity"`
	Message       string            `json:"message"`
	Timestamp     time.Time         `json:"timestamp"`
	Value         float64           `json:"value"`
	Threshold     float64           `json:"threshold"`
	
	// Additional context
	Model         string            `json:"model,omitempty"`
	UserID        string            `json:"user_id,omitempty"`
	RequestID     string            `json:"request_id,omitempty"`
	
	// Alert metadata
	Resolved      bool              `json:"resolved"`
	ResolvedAt    time.Time         `json:"resolved_at,omitempty"`
	Duration      time.Duration     `json:"duration,omitempty"`
}

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeLatency     AlertType = "latency"
	AlertTypeErrorRate   AlertType = "error_rate"
	AlertTypeCost        AlertType = "cost"
	AlertTypeTokenUsage  AlertType = "token_usage"
	AlertTypeRateLimit   AlertType = "rate_limit"
	AlertTypeAvailability AlertType = "availability"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(config *MetricsConfig) *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*ProviderMetrics),
		config:  config,
	}
}

// RecordRequest records metrics for a request
func (mc *MetricsCollector) RecordRequest(ctx context.Context, provider string, req *providers.GenerateRequest, resp *providers.GenerateResponse, latency time.Duration, err error) {
	if !mc.config.Enabled {
		return
	}
	
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	// Get or create provider metrics
	metrics, exists := mc.metrics[provider]
	if !exists {
		metrics = mc.createProviderMetrics(provider)
		mc.metrics[provider] = metrics
	}
	
	// Update request metrics
	metrics.TotalRequests++
	if err != nil {
		metrics.FailedRequests++
		mc.recordError(metrics, err)
	} else {
		metrics.SuccessfulRequests++
	}
	
	// Update latency metrics
	if mc.config.CollectLatency {
		mc.updateLatencyMetrics(metrics, latency)
	}
	
	// Update token metrics
	if mc.config.CollectTokens && resp != nil && resp.Usage != nil {
		mc.updateTokenMetrics(metrics, resp.Usage)
	}
	
	// Update cost metrics
	if mc.config.CollectCost && resp != nil && resp.Cost != nil {
		mc.updateCostMetrics(metrics, resp.Cost)
	}
	
	// Update quality metrics
	if mc.config.CollectQuality && resp != nil {
		mc.updateQualityMetrics(metrics, resp)
	}
	
	// Update time-based metrics
	mc.updateTimeBasedMetrics(metrics, req, resp, latency, err)
	
	// Update model-specific metrics
	mc.updateModelMetrics(metrics, req, resp, latency, err)
	
	// Update user-specific metrics
	if req.UserID != "" {
		mc.updateUserMetrics(metrics, req, resp, latency, err)
	}
	
	// Update cache metrics
	if resp != nil && resp.FromCache {
		metrics.CacheHits++
	} else {
		metrics.CacheMisses++
	}
	
	// Calculate derived metrics
	mc.calculateDerivedMetrics(metrics)
	
	metrics.LastUpdated = time.Now()
	
	// Check for alerts
	mc.checkAlerts(provider, metrics)
	
	// Trigger callback
	if mc.onMetricUpdate != nil {
		mc.onMetricUpdate(provider, metrics)
	}
}

// GetMetrics returns metrics for a provider
func (mc *MetricsCollector) GetMetrics(provider string) *ProviderMetrics {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	if metrics, exists := mc.metrics[provider]; exists {
		// Return a copy to prevent external modifications
		copy := *metrics
		return &copy
	}
	
	return nil
}

// GetAllMetrics returns metrics for all providers
func (mc *MetricsCollector) GetAllMetrics() map[string]*ProviderMetrics {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	
	result := make(map[string]*ProviderMetrics)
	for provider, metrics := range mc.metrics {
		copy := *metrics
		result[provider] = &copy
	}
	
	return result
}

// createProviderMetrics creates initial metrics for a provider
func (mc *MetricsCollector) createProviderMetrics(provider string) *ProviderMetrics {
	return &ProviderMetrics{
		Provider:         provider,
		ErrorsByType:     make(map[string]int64),
		HourlyMetrics:    make(map[string]*HourlyMetrics),
		DailyMetrics:     make(map[string]*DailyMetrics),
		ModelMetrics:     make(map[string]*ModelMetrics),
		UserMetrics:      make(map[string]*UserMetrics),
		LatencyHistogram: NewHistogram([]float64{10, 50, 100, 500, 1000, 5000, 10000}), // milliseconds
		CostHistogram:    NewHistogram([]float64{0.001, 0.01, 0.1, 1.0, 10.0}),         // dollars
	}
}

// updateLatencyMetrics updates latency-related metrics
func (mc *MetricsCollector) updateLatencyMetrics(metrics *ProviderMetrics, latency time.Duration) {
	// Update basic latency metrics
	if metrics.MinLatency == 0 || latency < metrics.MinLatency {
		metrics.MinLatency = latency
	}
	if latency > metrics.MaxLatency {
		metrics.MaxLatency = latency
	}
	
	// Update average latency
	if metrics.AverageLatency == 0 {
		metrics.AverageLatency = latency
	} else {
		metrics.AverageLatency = (metrics.AverageLatency + latency) / 2
	}
	
	// Update histogram
	metrics.LatencyHistogram.Observe(float64(latency.Milliseconds()))
}

// updateTokenMetrics updates token-related metrics
func (mc *MetricsCollector) updateTokenMetrics(metrics *ProviderMetrics, usage *providers.UsageInfo) {
	metrics.TotalTokens += int64(usage.TotalTokens)
	metrics.InputTokens += int64(usage.PromptTokens)
	metrics.OutputTokens += int64(usage.CompletionTokens)
	
	// Calculate average tokens per request
	if metrics.TotalRequests > 0 {
		metrics.AverageTokensPerRequest = float64(metrics.TotalTokens) / float64(metrics.TotalRequests)
	}
}

// updateCostMetrics updates cost-related metrics
func (mc *MetricsCollector) updateCostMetrics(metrics *ProviderMetrics, cost *providers.CostInfo) {
	metrics.TotalCost += cost.TotalCost
	
	// Calculate average cost per request
	if metrics.TotalRequests > 0 {
		metrics.AverageCostPerRequest = metrics.TotalCost / float64(metrics.TotalRequests)
	}
	
	// Calculate average cost per token
	if metrics.TotalTokens > 0 {
		metrics.AverageCostPerToken = metrics.TotalCost / float64(metrics.TotalTokens)
	}
	
	// Update histogram
	metrics.CostHistogram.Observe(cost.TotalCost)
}

// updateQualityMetrics updates quality-related metrics
func (mc *MetricsCollector) updateQualityMetrics(metrics *ProviderMetrics, resp *providers.GenerateResponse) {
	// Update confidence if available
	if resp.Cost != nil {
		// Placeholder for confidence calculation
		// This would be implemented based on response metadata
	}
	
	// Update relevance if available
	// This would be implemented based on response analysis
}

// recordError records error information
func (mc *MetricsCollector) recordError(metrics *ProviderMetrics, err error) {
	if !mc.config.CollectErrors {
		return
	}
	
	// Categorize error type
	errorType := "unknown"
	if providerErr, ok := err.(*providers.ProviderError); ok {
		errorType = providerErr.Code
	}
	
	metrics.ErrorsByType[errorType]++
}

// calculateDerivedMetrics calculates derived metrics
func (mc *MetricsCollector) calculateDerivedMetrics(metrics *ProviderMetrics) {
	// Calculate error rate
	if metrics.TotalRequests > 0 {
		metrics.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
	}
	
	// Calculate cache hit rate
	totalCacheRequests := metrics.CacheHits + metrics.CacheMisses
	if totalCacheRequests > 0 {
		metrics.CacheHitRate = float64(metrics.CacheHits) / float64(totalCacheRequests)
	}
}

// checkAlerts checks for alert conditions
func (mc *MetricsCollector) checkAlerts(provider string, metrics *ProviderMetrics) {
	// Check latency threshold
	if mc.config.LatencyThreshold > 0 && metrics.AverageLatency > mc.config.LatencyThreshold {
		alert := Alert{
			ID:        fmt.Sprintf("latency_%s_%d", provider, time.Now().Unix()),
			Provider:  provider,
			Type:      AlertTypeLatency,
			Severity:  AlertSeverityWarning,
			Message:   fmt.Sprintf("High latency detected for %s: %v", provider, metrics.AverageLatency),
			Timestamp: time.Now(),
			Value:     float64(metrics.AverageLatency.Milliseconds()),
			Threshold: float64(mc.config.LatencyThreshold.Milliseconds()),
		}
		
		if mc.onAlert != nil {
			mc.onAlert(alert)
		}
	}
	
	// Check error rate threshold
	if mc.config.ErrorRateThreshold > 0 && metrics.ErrorRate > mc.config.ErrorRateThreshold {
		alert := Alert{
			ID:        fmt.Sprintf("error_rate_%s_%d", provider, time.Now().Unix()),
			Provider:  provider,
			Type:      AlertTypeErrorRate,
			Severity:  AlertSeverityError,
			Message:   fmt.Sprintf("High error rate detected for %s: %.2f%%", provider, metrics.ErrorRate*100),
			Timestamp: time.Now(),
			Value:     metrics.ErrorRate,
			Threshold: mc.config.ErrorRateThreshold,
		}
		
		if mc.onAlert != nil {
			mc.onAlert(alert)
		}
	}
	
	// Check cost threshold
	if mc.config.CostThreshold > 0 && metrics.TotalCost > mc.config.CostThreshold {
		alert := Alert{
			ID:        fmt.Sprintf("cost_%s_%d", provider, time.Now().Unix()),
			Provider:  provider,
			Type:      AlertTypeCost,
			Severity:  AlertSeverityWarning,
			Message:   fmt.Sprintf("High cost detected for %s: $%.2f", provider, metrics.TotalCost),
			Timestamp: time.Now(),
			Value:     metrics.TotalCost,
			Threshold: mc.config.CostThreshold,
		}
		
		if mc.onAlert != nil {
			mc.onAlert(alert)
		}
	}
}

// updateTimeBasedMetrics updates hourly and daily metrics
func (mc *MetricsCollector) updateTimeBasedMetrics(metrics *ProviderMetrics, req *providers.GenerateRequest, resp *providers.GenerateResponse, latency time.Duration, err error) {
	now := time.Now()
	
	// Update hourly metrics
	hour := now.Format("2006-01-02-15")
	hourlyMetrics, exists := metrics.HourlyMetrics[hour]
	if !exists {
		hourlyMetrics = &HourlyMetrics{Hour: hour}
		metrics.HourlyMetrics[hour] = hourlyMetrics
	}
	
	hourlyMetrics.Requests++
	if err != nil {
		hourlyMetrics.Errors++
	}
	if resp != nil && resp.Usage != nil {
		hourlyMetrics.Tokens += int64(resp.Usage.TotalTokens)
	}
	if resp != nil && resp.Cost != nil {
		hourlyMetrics.Cost += resp.Cost.TotalCost
	}
	hourlyMetrics.AverageLatency = (hourlyMetrics.AverageLatency + latency) / 2
	hourlyMetrics.LastUpdated = now
	
	// Update daily metrics
	date := now.Format("2006-01-02")
	dailyMetrics, exists := metrics.DailyMetrics[date]
	if !exists {
		dailyMetrics = &DailyMetrics{Date: date}
		metrics.DailyMetrics[date] = dailyMetrics
	}
	
	dailyMetrics.Requests++
	if err != nil {
		dailyMetrics.Errors++
	}
	if resp != nil && resp.Usage != nil {
		dailyMetrics.Tokens += int64(resp.Usage.TotalTokens)
	}
	if resp != nil && resp.Cost != nil {
		dailyMetrics.Cost += resp.Cost.TotalCost
	}
	dailyMetrics.AverageLatency = (dailyMetrics.AverageLatency + latency) / 2
	dailyMetrics.LastUpdated = now
}

// updateModelMetrics updates model-specific metrics
func (mc *MetricsCollector) updateModelMetrics(metrics *ProviderMetrics, req *providers.GenerateRequest, resp *providers.GenerateResponse, latency time.Duration, err error) {
	model := req.Model
	if model == "" {
		model = "unknown"
	}
	
	modelMetrics, exists := metrics.ModelMetrics[model]
	if !exists {
		modelMetrics = &ModelMetrics{Model: model}
		metrics.ModelMetrics[model] = modelMetrics
	}
	
	modelMetrics.Requests++
	if err != nil {
		modelMetrics.Errors++
	}
	if resp != nil && resp.Usage != nil {
		modelMetrics.InputTokens += int64(resp.Usage.PromptTokens)
		modelMetrics.OutputTokens += int64(resp.Usage.CompletionTokens)
		modelMetrics.TotalTokens += int64(resp.Usage.TotalTokens)
	}
	if resp != nil && resp.Cost != nil {
		modelMetrics.Cost += resp.Cost.TotalCost
	}
	modelMetrics.AverageLatency = (modelMetrics.AverageLatency + latency) / 2
	modelMetrics.LastUsed = time.Now()
}

// updateUserMetrics updates user-specific metrics
func (mc *MetricsCollector) updateUserMetrics(metrics *ProviderMetrics, req *providers.GenerateRequest, resp *providers.GenerateResponse, latency time.Duration, err error) {
	userID := req.UserID
	
	userMetrics, exists := metrics.UserMetrics[userID]
	if !exists {
		userMetrics = &UserMetrics{UserID: userID}
		metrics.UserMetrics[userID] = userMetrics
	}
	
	userMetrics.Requests++
	if err != nil {
		userMetrics.Errors++
	}
	if resp != nil && resp.Usage != nil {
		userMetrics.Tokens += int64(resp.Usage.TotalTokens)
	}
	if resp != nil && resp.Cost != nil {
		userMetrics.Cost += resp.Cost.TotalCost
	}
	userMetrics.AverageLatency = (userMetrics.AverageLatency + latency) / 2
	userMetrics.LastUsed = time.Now()
}

// NewHistogram creates a new histogram with the given buckets
func NewHistogram(buckets []float64) *Histogram {
	histogramBuckets := make([]HistogramBucket, len(buckets))
	for i, bound := range buckets {
		histogramBuckets[i] = HistogramBucket{UpperBound: bound}
	}
	
	return &Histogram{
		Buckets: histogramBuckets,
	}
}

// Observe adds an observation to the histogram
func (h *Histogram) Observe(value float64) {
	h.Count++
	h.Sum += value
	
	if h.Count == 1 {
		h.Min = value
		h.Max = value
	} else {
		if value < h.Min {
			h.Min = value
		}
		if value > h.Max {
			h.Max = value
		}
	}
	
	// Update buckets
	for i := range h.Buckets {
		if value <= h.Buckets[i].UpperBound {
			h.Buckets[i].Count++
		}
	}
	
	// Update cumulative counts
	cumulative := int64(0)
	for i := range h.Buckets {
		cumulative += h.Buckets[i].Count
		h.Buckets[i].CumulativeCount = cumulative
	}
}

// SetMetricUpdateCallback sets the callback for metric updates
func (mc *MetricsCollector) SetMetricUpdateCallback(callback func(provider string, metrics *ProviderMetrics)) {
	mc.onMetricUpdate = callback
}

// SetAlertCallback sets the callback for alerts
func (mc *MetricsCollector) SetAlertCallback(callback func(alert Alert)) {
	mc.onAlert = callback
}
