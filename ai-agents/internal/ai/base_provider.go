package ai

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go-coffee-ai-agents/internal/observability"

	"github.com/google/uuid"
)

// BaseProvider provides common functionality for AI providers
type BaseProvider struct {
	name           string
	config         ProviderConfig
	models         []Model
	usage          *UsageStatistics
	logger         *observability.StructuredLogger
	metrics        *observability.MetricsCollector
	tracing        *observability.TracingHelper
	circuitBreaker CircuitBreaker
	rateLimiter    RateLimiter

	// Thread safety
	mutex sync.RWMutex
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(
	name string,
	config ProviderConfig,
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) *BaseProvider {
	return &BaseProvider{
		name:    name,
		config:  config,
		models:  make([]Model, 0),
		logger:  logger,
		metrics: metrics,
		tracing: tracing,
		usage: &UsageStatistics{
			RequestsByModel: make(map[string]int64),
			TokensByModel:   make(map[string]int64),
			CostByModel:     make(map[string]float64),
		},
	}
}

// GetName returns the provider name
func (bp *BaseProvider) GetName() string {
	return bp.name
}

// GetModels returns available models for this provider
func (bp *BaseProvider) GetModels() []Model {
	bp.mutex.RLock()
	defer bp.mutex.RUnlock()

	models := make([]Model, len(bp.models))
	copy(models, bp.models)
	return models
}

// AddModel adds a model to the provider
func (bp *BaseProvider) AddModel(model Model) {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	model.Provider = bp.name
	bp.models = append(bp.models, model)

	bp.logger.Debug("Added model to provider",
		"provider", bp.name,
		"model", model.ID,
		"type", model.Type)
}

// GetModel returns a specific model by ID
func (bp *BaseProvider) GetModel(modelID string) (*Model, error) {
	bp.mutex.RLock()
	defer bp.mutex.RUnlock()

	for _, model := range bp.models {
		if model.ID == modelID {
			return &model, nil
		}
	}

	return nil, &AIError{
		Code:      ErrorCodeModelNotFound,
		Message:   fmt.Sprintf("model %s not found in provider %s", modelID, bp.name),
		Provider:  bp.name,
		Model:     modelID,
		Retryable: false,
	}
}

// GetUsage returns usage statistics
func (bp *BaseProvider) GetUsage() *UsageStatistics {
	bp.mutex.RLock()
	defer bp.mutex.RUnlock()

	// Create a copy to avoid race conditions
	usage := &UsageStatistics{
		TotalRequests:   bp.usage.TotalRequests,
		SuccessfulReqs:  bp.usage.SuccessfulReqs,
		FailedRequests:  bp.usage.FailedRequests,
		TotalTokens:     bp.usage.TotalTokens,
		TotalCost:       bp.usage.TotalCost,
		AverageLatency:  bp.usage.AverageLatency,
		LastRequestTime: bp.usage.LastRequestTime,
		RequestsByModel: make(map[string]int64),
		TokensByModel:   make(map[string]int64),
		CostByModel:     make(map[string]float64),
	}

	for k, v := range bp.usage.RequestsByModel {
		usage.RequestsByModel[k] = v
	}
	for k, v := range bp.usage.TokensByModel {
		usage.TokensByModel[k] = v
	}
	for k, v := range bp.usage.CostByModel {
		usage.CostByModel[k] = v
	}

	return usage
}

// RecordRequest records a request in usage statistics
func (bp *BaseProvider) RecordRequest(ctx context.Context, model string, tokens int, cost float64, duration time.Duration, success bool) {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	bp.usage.TotalRequests++
	bp.usage.LastRequestTime = time.Now()

	if success {
		bp.usage.SuccessfulReqs++
	} else {
		bp.usage.FailedRequests++
	}

	bp.usage.TotalTokens += int64(tokens)
	bp.usage.TotalCost += cost

	// Update average latency
	if bp.usage.TotalRequests == 1 {
		bp.usage.AverageLatency = duration
	} else {
		// Exponential moving average
		alpha := 0.1
		bp.usage.AverageLatency = time.Duration(float64(bp.usage.AverageLatency)*(1-alpha) + float64(duration)*alpha)
	}

	// Update per-model statistics
	bp.usage.RequestsByModel[model]++
	bp.usage.TokensByModel[model] += int64(tokens)
	bp.usage.CostByModel[model] += cost

	// Record metrics
	if bp.metrics != nil {
		bp.metrics.RecordAIRequest(ctx, bp.name, "generate", duration, success)
	}

	bp.logger.DebugContext(ctx, "Recorded AI request",
		"provider", bp.name,
		"model", model,
		"tokens", tokens,
		"cost", cost,
		"duration_ms", duration.Milliseconds(),
		"success", success)
}

// ExecuteWithObservability executes a function with observability and resilience
func (bp *BaseProvider) ExecuteWithObservability(
	ctx context.Context,
	operation string,
	model string,
	fn func(context.Context) (interface{}, *TokenUsage, error),
) (interface{}, *TokenUsage, error) {
	// Start tracing
	ctx, span := bp.tracing.StartAISpan(ctx, operation, bp.name, model)
	defer span.End()

	start := time.Now()
	requestID := uuid.New().String()

	bp.logger.InfoContext(ctx, "Starting AI request",
		"provider", bp.name,
		"model", model,
		"operation", operation,
		"request_id", requestID)

	// Check rate limiter
	if bp.rateLimiter != nil {
		if err := bp.rateLimiter.Wait(ctx); err != nil {
			duration := time.Since(start)
			bp.RecordRequest(ctx, model, 0, 0, duration, false)

			aiErr := &AIError{
				Code:      ErrorCodeRateLimit,
				Message:   "rate limit exceeded",
				Provider:  bp.name,
				Model:     model,
				Retryable: true,
			}

			bp.tracing.RecordError(span, aiErr, "Rate limit exceeded")
			bp.logger.WarnContext(ctx, "Rate limit exceeded",
				"provider", bp.name,
				"model", model,
				"request_id", requestID)

			return nil, nil, aiErr
		}
	}

	// Execute with circuit breaker
	var result interface{}
	var usage *TokenUsage
	var err error

	if bp.circuitBreaker != nil {
		err = bp.circuitBreaker.Execute(ctx, func() error {
			result, usage, err = fn(ctx)
			return err
		})
	} else {
		result, usage, err = fn(ctx)
	}

	duration := time.Since(start)

	// Calculate cost
	var cost float64
	if usage != nil {
		if model, modelErr := bp.GetModel(model); modelErr == nil {
			cost = float64(usage.PromptTokens)*model.InputCost/1000 +
				float64(usage.CompletionTokens)*model.OutputCost/1000
		}
	}

	// Record usage
	tokens := 0
	if usage != nil {
		tokens = usage.TotalTokens
	}
	bp.RecordRequest(ctx, model, tokens, cost, duration, err == nil)

	if err != nil {
		bp.tracing.RecordError(span, err, "AI request failed")
		bp.logger.ErrorContext(ctx, "AI request failed", err,
			"provider", bp.name,
			"model", model,
			"operation", operation,
			"request_id", requestID,
			"duration_ms", duration.Milliseconds())

		return nil, usage, err
	}

	bp.tracing.RecordSuccess(span, "AI request completed successfully")
	bp.logger.InfoContext(ctx, "AI request completed successfully",
		"provider", bp.name,
		"model", model,
		"operation", operation,
		"request_id", requestID,
		"tokens", tokens,
		"cost", cost,
		"duration_ms", duration.Milliseconds())

	return result, usage, nil
}

// ValidateRequest validates common request parameters
func (bp *BaseProvider) ValidateRequest(model string, maxTokens int) error {
	// Check if model exists
	modelInfo, err := bp.GetModel(model)
	if err != nil {
		return err
	}

	// Check token limits
	if maxTokens > 0 && maxTokens > modelInfo.MaxTokens {
		return &AIError{
			Code:      ErrorCodeInvalidRequest,
			Message:   fmt.Sprintf("max_tokens %d exceeds model limit %d", maxTokens, modelInfo.MaxTokens),
			Provider:  bp.name,
			Model:     model,
			Retryable: false,
		}
	}

	return nil
}

// SetCircuitBreaker sets the circuit breaker for this provider
func (bp *BaseProvider) SetCircuitBreaker(cb CircuitBreaker) {
	bp.circuitBreaker = cb
}

// SetRateLimiter sets the rate limiter for this provider
func (bp *BaseProvider) SetRateLimiter(rl RateLimiter) {
	bp.rateLimiter = rl
}

// GetCircuitBreakerStats returns circuit breaker statistics
func (bp *BaseProvider) GetCircuitBreakerStats() *CircuitBreakerStats {
	if bp.circuitBreaker == nil {
		return nil
	}
	return bp.circuitBreaker.GetStats()
}

// GetRateLimiterStats returns rate limiter statistics
func (bp *BaseProvider) GetRateLimiterStats() *RateLimiterStats {
	if bp.rateLimiter == nil {
		return nil
	}
	return bp.rateLimiter.GetStats()
}

// Close closes the provider and cleans up resources
func (bp *BaseProvider) Close() error {
	bp.logger.Info("Closing AI provider", "provider", bp.name)

	// Reset circuit breaker if present
	if bp.circuitBreaker != nil {
		bp.circuitBreaker.Reset()
	}

	return nil
}

// CreateRequestContext creates a request context with default values
func (bp *BaseProvider) CreateRequestContext(source string) *RequestContext {
	return &RequestContext{
		RequestID:  uuid.New().String(),
		Source:     source,
		Priority:   PriorityNormal,
		Metadata:   make(map[string]interface{}),
		StartTime:  time.Now(),
		Timeout:    bp.config.Timeout,
		RetryCount: 0,
		MaxRetries: bp.config.MaxRetries,
	}
}

// ShouldRetry determines if a request should be retried based on the error
func (bp *BaseProvider) ShouldRetry(err error, retryCount int) bool {
	if retryCount >= bp.config.MaxRetries {
		return false
	}

	if aiErr, ok := err.(*AIError); ok {
		return aiErr.Retryable
	}

	// Default to retrying for unknown errors
	return true
}

// CalculateRetryDelay calculates the delay before retrying a request
func (bp *BaseProvider) CalculateRetryDelay(retryCount int) time.Duration {
	// Exponential backoff with jitter
	baseDelay := bp.config.RetryDelay

	// Cap retry count to prevent overflow
	if retryCount > 10 {
		retryCount = 10
	}

	// Use math.Pow to safely calculate exponential backoff
	multiplier := math.Pow(2, float64(retryCount))
	delay := time.Duration(float64(baseDelay) * multiplier)

	// Add jitter (±25%)
	jitter := time.Duration(float64(delay) * 0.25)
	delay += time.Duration(float64(jitter) * (2*float64(time.Now().UnixNano()%1000)/1000.0 - 1))

	// Cap at maximum delay
	maxDelay := 30 * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// GenerateResponseID generates a unique response ID
func (bp *BaseProvider) GenerateResponseID() string {
	return fmt.Sprintf("%s_%s", bp.name, uuid.New().String())
}

// AddMetadata adds metadata to a request context
func (bp *BaseProvider) AddMetadata(ctx *RequestContext, key string, value interface{}) {
	if ctx.Metadata == nil {
		ctx.Metadata = make(map[string]interface{})
	}
	ctx.Metadata[key] = value
}

// GetMetadata gets metadata from a request context
func (bp *BaseProvider) GetMetadata(ctx *RequestContext, key string) (interface{}, bool) {
	if ctx.Metadata == nil {
		return nil, false
	}
	value, exists := ctx.Metadata[key]
	return value, exists
}
