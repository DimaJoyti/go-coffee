package ai

import (
	"context"
	"time"

)

// Provider represents an AI provider interface
type Provider interface {
	// GetName returns the provider name
	GetName() string
	
	// GetModels returns available models for this provider
	GetModels() []Model
	
	// GenerateText generates text using the specified model
	GenerateText(ctx context.Context, request *TextGenerationRequest) (*TextGenerationResponse, error)
	
	// GenerateChat generates chat completion using the specified model
	GenerateChat(ctx context.Context, request *ChatRequest) (*ChatResponse, error)
	
	// GenerateEmbedding generates embeddings for the given text
	GenerateEmbedding(ctx context.Context, request *EmbeddingRequest) (*EmbeddingResponse, error)
	
	// HealthCheck performs a health check on the provider
	HealthCheck(ctx context.Context) error
	
	// GetUsage returns usage statistics
	GetUsage() *UsageStatistics
	
	// Close closes the provider and cleans up resources
	Close() error
}

// Model represents an AI model
type Model struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Provider     string            `json:"provider"`
	Type         ModelType         `json:"type"`
	MaxTokens    int               `json:"max_tokens"`
	InputCost    float64           `json:"input_cost"`   // Cost per 1K tokens
	OutputCost   float64           `json:"output_cost"`  // Cost per 1K tokens
	Capabilities []string          `json:"capabilities"`
	Metadata     map[string]string `json:"metadata"`
}

// ModelType represents the type of AI model
type ModelType string

const (
	ModelTypeText      ModelType = "text"
	ModelTypeChat      ModelType = "chat"
	ModelTypeEmbedding ModelType = "embedding"
	ModelTypeImage     ModelType = "image"
	ModelTypeCode      ModelType = "code"
)

// TextGenerationRequest represents a text generation request
type TextGenerationRequest struct {
	Model       string            `json:"model"`
	Prompt      string            `json:"prompt"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Temperature float64           `json:"temperature,omitempty"`
	TopP        float64           `json:"top_p,omitempty"`
	TopK        int               `json:"top_k,omitempty"`
	Stop        []string          `json:"stop,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// TextGenerationResponse represents a text generation response
type TextGenerationResponse struct {
	ID           string            `json:"id"`
	Text         string            `json:"text"`
	Model        string            `json:"model"`
	Provider     string            `json:"provider"`
	Usage        *TokenUsage       `json:"usage"`
	FinishReason string            `json:"finish_reason"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model       string            `json:"model"`
	Messages    []ChatMessage     `json:"messages"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Temperature float64           `json:"temperature,omitempty"`
	TopP        float64           `json:"top_p,omitempty"`
	TopK        int               `json:"top_k,omitempty"`
	Stop        []string          `json:"stop,omitempty"`
	Stream      bool              `json:"stream,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	ID           string            `json:"id"`
	Message      ChatMessage       `json:"message"`
	Model        string            `json:"model"`
	Provider     string            `json:"provider"`
	Usage        *TokenUsage       `json:"usage"`
	FinishReason string            `json:"finish_reason"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"`    // system, user, assistant
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// EmbeddingRequest represents an embedding generation request
type EmbeddingRequest struct {
	Model    string            `json:"model"`
	Input    []string          `json:"input"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// EmbeddingResponse represents an embedding generation response
type EmbeddingResponse struct {
	ID         string            `json:"id"`
	Embeddings [][]float64       `json:"embeddings"`
	Model      string            `json:"model"`
	Provider   string            `json:"provider"`
	Usage      *TokenUsage       `json:"usage"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

// TokenUsage represents token usage information
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// UsageStatistics represents provider usage statistics
type UsageStatistics struct {
	TotalRequests    int64         `json:"total_requests"`
	SuccessfulReqs   int64         `json:"successful_requests"`
	FailedRequests   int64         `json:"failed_requests"`
	TotalTokens      int64         `json:"total_tokens"`
	TotalCost        float64       `json:"total_cost"`
	AverageLatency   time.Duration `json:"average_latency"`
	LastRequestTime  time.Time     `json:"last_request_time"`
	RequestsByModel  map[string]int64 `json:"requests_by_model"`
	TokensByModel    map[string]int64 `json:"tokens_by_model"`
	CostByModel      map[string]float64 `json:"cost_by_model"`
}

// ProviderConfig represents configuration for an AI provider
type ProviderConfig struct {
	Name        string            `yaml:"name"`
	Type        string            `yaml:"type"`
	APIKey      string            `yaml:"api_key"`
	BaseURL     string            `yaml:"base_url,omitempty"`
	Timeout     time.Duration     `yaml:"timeout"`
	MaxRetries  int               `yaml:"max_retries"`
	RetryDelay  time.Duration     `yaml:"retry_delay"`
	RateLimit   int               `yaml:"rate_limit,omitempty"`
	Enabled     bool              `yaml:"enabled"`
	Models      []string          `yaml:"models,omitempty"`
	Metadata    map[string]string `yaml:"metadata,omitempty"`
}

// RequestContext represents context for AI requests
type RequestContext struct {
	RequestID    string                 `json:"request_id"`
	UserID       string                 `json:"user_id,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
	Source       string                 `json:"source"`
	Priority     RequestPriority        `json:"priority"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	StartTime    time.Time              `json:"start_time"`
	Timeout      time.Duration          `json:"timeout,omitempty"`
	RetryCount   int                    `json:"retry_count"`
	MaxRetries   int                    `json:"max_retries"`
}

// RequestPriority represents the priority of an AI request
type RequestPriority string

const (
	PriorityLow    RequestPriority = "low"
	PriorityNormal RequestPriority = "normal"
	PriorityHigh   RequestPriority = "high"
	PriorityCritical RequestPriority = "critical"
)

// Error types for AI operations
type AIError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Provider string `json:"provider"`
	Model    string `json:"model,omitempty"`
	Retryable bool  `json:"retryable"`
}

func (e *AIError) Error() string {
	return e.Message
}

// Common AI error codes
const (
	ErrorCodeInvalidRequest   = "invalid_request"
	ErrorCodeAuthentication   = "authentication_error"
	ErrorCodeRateLimit        = "rate_limit_exceeded"
	ErrorCodeQuotaExceeded    = "quota_exceeded"
	ErrorCodeModelNotFound    = "model_not_found"
	ErrorCodeTimeout          = "timeout"
	ErrorCodeInternalError    = "internal_error"
	ErrorCodeServiceUnavailable = "service_unavailable"
)

// BeverageGenerationRequest represents a request to generate beverage recipes
type BeverageGenerationRequest struct {
	Theme        string            `json:"theme"`
	Ingredients  []string          `json:"ingredients,omitempty"`
	Dietary      []string          `json:"dietary_restrictions,omitempty"`
	Complexity   string            `json:"complexity,omitempty"` // simple, medium, complex
	Style        string            `json:"style,omitempty"`      // traditional, modern, fusion
	Temperature  string            `json:"temperature,omitempty"` // hot, cold, room
	Count        int               `json:"count,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// BeverageGenerationResponse represents a response containing generated beverage recipes
type BeverageGenerationResponse struct {
	ID        string                   `json:"id"`
	Beverages []GeneratedBeverage      `json:"beverages"`
	Model     string                   `json:"model"`
	Provider  string                   `json:"provider"`
	Usage     *TokenUsage              `json:"usage"`
	Metadata  map[string]string        `json:"metadata,omitempty"`
	CreatedAt time.Time                `json:"created_at"`
}

// GeneratedBeverage represents a generated beverage recipe
type GeneratedBeverage struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Ingredients  []BeverageIngredient   `json:"ingredients"`
	Instructions []string               `json:"instructions"`
	PrepTime     int                    `json:"prep_time_minutes"`
	Servings     int                    `json:"servings"`
	Difficulty   string                 `json:"difficulty"`
	Tags         []string               `json:"tags"`
	Nutrition    *NutritionInfo         `json:"nutrition,omitempty"`
	Cost         *CostEstimate          `json:"cost,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// BeverageIngredient represents an ingredient in a beverage recipe
type BeverageIngredient struct {
	Name        string  `json:"name"`
	Quantity    string  `json:"quantity"`
	Unit        string  `json:"unit,omitempty"`
	Type        string  `json:"type"`        // base, flavor, garnish, etc.
	Optional    bool    `json:"optional"`
	Substitute  string  `json:"substitute,omitempty"`
	Cost        float64 `json:"cost,omitempty"`
}

// NutritionInfo represents nutritional information
type NutritionInfo struct {
	Calories     int     `json:"calories"`
	Protein      float64 `json:"protein_g"`
	Carbs        float64 `json:"carbs_g"`
	Fat          float64 `json:"fat_g"`
	Sugar        float64 `json:"sugar_g"`
	Sodium       float64 `json:"sodium_mg"`
	Caffeine     float64 `json:"caffeine_mg,omitempty"`
}

// CostEstimate represents cost estimation
type CostEstimate struct {
	Total       float64 `json:"total"`
	Currency    string  `json:"currency"`
	PerServing  float64 `json:"per_serving"`
	Breakdown   map[string]float64 `json:"breakdown,omitempty"`
}

// TaskGenerationRequest represents a request to generate tasks
type TaskGenerationRequest struct {
	Context     string            `json:"context"`
	Goal        string            `json:"goal"`
	Priority    string            `json:"priority,omitempty"`
	Deadline    *time.Time        `json:"deadline,omitempty"`
	Skills      []string          `json:"skills,omitempty"`
	Resources   []string          `json:"resources,omitempty"`
	Count       int               `json:"count,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// TaskGenerationResponse represents a response containing generated tasks
type TaskGenerationResponse struct {
	ID        string            `json:"id"`
	Tasks     []GeneratedTask   `json:"tasks"`
	Model     string            `json:"model"`
	Provider  string            `json:"provider"`
	Usage     *TokenUsage       `json:"usage"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// GeneratedTask represents a generated task
type GeneratedTask struct {
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Priority     string                 `json:"priority"`
	EstimatedTime int                   `json:"estimated_time_hours"`
	Skills       []string               `json:"skills_required"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Deliverables []string               `json:"deliverables"`
	Acceptance   []string               `json:"acceptance_criteria"`
	Tags         []string               `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ProviderManager manages multiple AI providers
type ProviderManager interface {
	// RegisterProvider registers a new AI provider
	RegisterProvider(provider Provider) error
	
	// GetProvider returns a provider by name
	GetProvider(name string) (Provider, error)
	
	// GetProviders returns all registered providers
	GetProviders() []Provider
	
	// GetBestProvider returns the best provider for a given model type
	GetBestProvider(modelType ModelType) (Provider, error)
	
	// HealthCheck performs health checks on all providers
	HealthCheck(ctx context.Context) map[string]error
	
	// GetAggregatedUsage returns aggregated usage statistics
	GetAggregatedUsage() *AggregatedUsage
	
	// Close closes all providers
	Close() error
}

// AggregatedUsage represents aggregated usage across all providers
type AggregatedUsage struct {
	TotalRequests   int64                    `json:"total_requests"`
	SuccessfulReqs  int64                    `json:"successful_requests"`
	FailedRequests  int64                    `json:"failed_requests"`
	TotalTokens     int64                    `json:"total_tokens"`
	TotalCost       float64                  `json:"total_cost"`
	ProviderUsage   map[string]*UsageStatistics `json:"provider_usage"`
	LastUpdated     time.Time                `json:"last_updated"`
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	CircuitBreakerClosed   CircuitBreakerState = "closed"
	CircuitBreakerOpen     CircuitBreakerState = "open"
	CircuitBreakerHalfOpen CircuitBreakerState = "half_open"
)

// CircuitBreaker represents a circuit breaker for AI providers
type CircuitBreaker interface {
	// Execute executes a function with circuit breaker protection
	Execute(ctx context.Context, fn func() error) error
	
	// GetState returns the current state of the circuit breaker
	GetState() CircuitBreakerState
	
	// GetStats returns circuit breaker statistics
	GetStats() *CircuitBreakerStats
	
	// Reset resets the circuit breaker to closed state
	Reset()
}

// CircuitBreakerStats represents circuit breaker statistics
type CircuitBreakerStats struct {
	State           CircuitBreakerState `json:"state"`
	FailureCount    int64               `json:"failure_count"`
	SuccessCount    int64               `json:"success_count"`
	LastFailureTime time.Time           `json:"last_failure_time"`
	LastSuccessTime time.Time           `json:"last_success_time"`
	OpenedAt        time.Time           `json:"opened_at,omitempty"`
}

// RateLimiter represents a rate limiter for AI providers
type RateLimiter interface {
	// Allow returns true if the request is allowed
	Allow() bool
	
	// Wait waits until the request is allowed
	Wait(ctx context.Context) error
	
	// GetStats returns rate limiter statistics
	GetStats() *RateLimiterStats
}

// RateLimiterStats represents rate limiter statistics
type RateLimiterStats struct {
	RequestsPerSecond float64   `json:"requests_per_second"`
	TokensAvailable   int       `json:"tokens_available"`
	LastRefill        time.Time `json:"last_refill"`
	TotalRequests     int64     `json:"total_requests"`
	RejectedRequests  int64     `json:"rejected_requests"`
}
