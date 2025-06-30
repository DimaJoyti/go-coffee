package providers

import (
	"context"
	"io"
	"time"
)

// Provider represents an AI/LLM provider interface
type Provider interface {
	// Basic information
	Name() string
	Type() ProviderType
	IsAvailable(ctx context.Context) bool
	
	// Text generation
	GenerateText(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
	GenerateTextStream(ctx context.Context, req *GenerateRequest) (<-chan *StreamResponse, error)
	
	// Chat completion
	ChatCompletion(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	ChatCompletionStream(ctx context.Context, req *ChatRequest) (<-chan *StreamResponse, error)
	
	// Model management
	ListModels(ctx context.Context) ([]Model, error)
	GetModel(ctx context.Context, modelID string) (*Model, error)
	
	// Cost and usage
	CalculateCost(req *GenerateRequest, resp *GenerateResponse) (*CostInfo, error)
	GetUsage(ctx context.Context, timeRange TimeRange) (*UsageInfo, error)
	
	// Health and status
	HealthCheck(ctx context.Context) (*HealthStatus, error)
	GetLimits(ctx context.Context) (*RateLimits, error)
	
	// Configuration
	Configure(config ProviderConfig) error
	GetConfig() ProviderConfig
}

// ProviderType represents the type of AI provider
type ProviderType string

const (
	ProviderTypeOpenAI      ProviderType = "openai"
	ProviderTypeGemini      ProviderType = "gemini"
	ProviderTypeClaude      ProviderType = "claude"
	ProviderTypeOllama      ProviderType = "ollama"
	ProviderTypeHuggingFace ProviderType = "huggingface"
)

// GenerateRequest represents a text generation request
type GenerateRequest struct {
	// Model configuration
	Model       string            `json:"model"`
	Provider    string            `json:"provider,omitempty"`
	
	// Input
	Prompt      string            `json:"prompt"`
	Messages    []Message         `json:"messages,omitempty"`
	Context     string            `json:"context,omitempty"`
	
	// Generation parameters
	MaxTokens        int               `json:"max_tokens,omitempty"`
	Temperature      float64           `json:"temperature,omitempty"`
	TopP             float64           `json:"top_p,omitempty"`
	TopK             int               `json:"top_k,omitempty"`
	FrequencyPenalty float64           `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64           `json:"presence_penalty,omitempty"`
	StopSequences    []string          `json:"stop_sequences,omitempty"`
	
	// Function calling
	Functions        []Function        `json:"functions,omitempty"`
	FunctionCall     *FunctionCall     `json:"function_call,omitempty"`
	
	// Streaming
	Stream           bool              `json:"stream,omitempty"`
	
	// Metadata
	UserID           string            `json:"user_id,omitempty"`
	SessionID        string            `json:"session_id,omitempty"`
	RequestID        string            `json:"request_id,omitempty"`
	Tags             map[string]string `json:"tags,omitempty"`
	
	// Caching
	CacheKey         string            `json:"cache_key,omitempty"`
	CacheTTL         time.Duration     `json:"cache_ttl,omitempty"`
	
	// Safety and filtering
	SafetySettings   []SafetySetting   `json:"safety_settings,omitempty"`
	ContentFilter    bool              `json:"content_filter,omitempty"`
}

// GenerateResponse represents a text generation response
type GenerateResponse struct {
	// Generated content
	Text         string            `json:"text"`
	Choices      []Choice          `json:"choices,omitempty"`
	
	// Function calling
	FunctionCall *FunctionCall     `json:"function_call,omitempty"`
	
	// Usage information
	Usage        *UsageInfo        `json:"usage"`
	Cost         *CostInfo         `json:"cost,omitempty"`
	
	// Model information
	Model        string            `json:"model"`
	Provider     string            `json:"provider"`
	
	// Response metadata
	ID           string            `json:"id"`
	Created      time.Time         `json:"created"`
	FinishReason string            `json:"finish_reason"`
	
	// Safety and filtering
	SafetyRatings []SafetyRating   `json:"safety_ratings,omitempty"`
	Filtered      bool             `json:"filtered,omitempty"`
	
	// Caching
	FromCache     bool             `json:"from_cache,omitempty"`
	CacheKey      string           `json:"cache_key,omitempty"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	// Model configuration
	Model       string            `json:"model"`
	Provider    string            `json:"provider,omitempty"`
	
	// Conversation
	Messages    []Message         `json:"messages"`
	SystemPrompt string           `json:"system_prompt,omitempty"`
	
	// Generation parameters
	MaxTokens        int               `json:"max_tokens,omitempty"`
	Temperature      float64           `json:"temperature,omitempty"`
	TopP             float64           `json:"top_p,omitempty"`
	FrequencyPenalty float64           `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64           `json:"presence_penalty,omitempty"`
	StopSequences    []string          `json:"stop_sequences,omitempty"`
	
	// Function calling
	Functions        []Function        `json:"functions,omitempty"`
	FunctionCall     *FunctionCall     `json:"function_call,omitempty"`
	
	// Streaming
	Stream           bool              `json:"stream,omitempty"`
	
	// Metadata
	UserID           string            `json:"user_id,omitempty"`
	SessionID        string            `json:"session_id,omitempty"`
	RequestID        string            `json:"request_id,omitempty"`
	Tags             map[string]string `json:"tags,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	// Generated content
	Message      Message           `json:"message"`
	Choices      []Choice          `json:"choices,omitempty"`
	
	// Function calling
	FunctionCall *FunctionCall     `json:"function_call,omitempty"`
	
	// Usage information
	Usage        *UsageInfo        `json:"usage"`
	Cost         *CostInfo         `json:"cost,omitempty"`
	
	// Model information
	Model        string            `json:"model"`
	Provider     string            `json:"provider"`
	
	// Response metadata
	ID           string            `json:"id"`
	Created      time.Time         `json:"created"`
	FinishReason string            `json:"finish_reason"`
	
	// Safety and filtering
	SafetyRatings []SafetyRating   `json:"safety_ratings,omitempty"`
	Filtered      bool             `json:"filtered,omitempty"`
}

// StreamResponse represents a streaming response chunk
type StreamResponse struct {
	// Content delta
	Delta        *Delta            `json:"delta,omitempty"`
	Text         string            `json:"text,omitempty"`
	
	// Function calling
	FunctionCall *FunctionCall     `json:"function_call,omitempty"`
	
	// Response metadata
	ID           string            `json:"id"`
	Model        string            `json:"model"`
	Provider     string            `json:"provider"`
	Index        int               `json:"index"`
	FinishReason string            `json:"finish_reason,omitempty"`
	
	// Usage (final chunk only)
	Usage        *UsageInfo        `json:"usage,omitempty"`
	Cost         *CostInfo         `json:"cost,omitempty"`
	
	// Error handling
	Error        error             `json:"error,omitempty"`
	Done         bool              `json:"done"`
}

// Message represents a conversation message
type Message struct {
	Role         MessageRole       `json:"role"`
	Content      string            `json:"content"`
	Name         string            `json:"name,omitempty"`
	FunctionCall *FunctionCall     `json:"function_call,omitempty"`
	Images       []Image           `json:"images,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// MessageRole represents the role of a message sender
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleFunction  MessageRole = "function"
)

// Choice represents a generation choice
type Choice struct {
	Index        int               `json:"index"`
	Message      *Message          `json:"message,omitempty"`
	Text         string            `json:"text,omitempty"`
	FinishReason string            `json:"finish_reason"`
	LogProbs     *LogProbs         `json:"logprobs,omitempty"`
}

// Delta represents a streaming content delta
type Delta struct {
	Role         MessageRole       `json:"role,omitempty"`
	Content      string            `json:"content,omitempty"`
	FunctionCall *FunctionCall     `json:"function_call,omitempty"`
}

// Function represents a function definition for function calling
type Function struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name      string            `json:"name"`
	Arguments string            `json:"arguments"`
}

// Image represents an image input
type Image struct {
	URL       string            `json:"url,omitempty"`
	Data      []byte            `json:"data,omitempty"`
	MimeType  string            `json:"mime_type,omitempty"`
	Detail    string            `json:"detail,omitempty"` // low, high, auto
}

// Model represents an AI model
type Model struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Provider     string            `json:"provider"`
	Description  string            `json:"description"`
	
	// Capabilities
	MaxTokens         int             `json:"max_tokens"`
	ContextWindow     int             `json:"context_window"`
	SupportsStreaming bool            `json:"supports_streaming"`
	SupportsImages    bool            `json:"supports_images"`
	SupportsFunctions bool            `json:"supports_functions"`
	SupportsVision    bool            `json:"supports_vision"`
	
	// Cost information
	InputCostPer1K    float64         `json:"input_cost_per_1k"`
	OutputCostPer1K   float64         `json:"output_cost_per_1k"`
	Currency          string          `json:"currency"`
	
	// Availability
	Available         bool            `json:"available"`
	Deprecated        bool            `json:"deprecated"`
	
	// Metadata
	Version           string          `json:"version"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	Tags              []string        `json:"tags"`
}

// UsageInfo represents token usage information
type UsageInfo struct {
	PromptTokens     int             `json:"prompt_tokens"`
	CompletionTokens int             `json:"completion_tokens"`
	TotalTokens      int             `json:"total_tokens"`
	
	// Provider-specific usage
	Characters       int             `json:"characters,omitempty"`
	Words            int             `json:"words,omitempty"`
	Images           int             `json:"images,omitempty"`
	
	// Billing information
	BillableTokens   int             `json:"billable_tokens,omitempty"`
	BillableUnits    int             `json:"billable_units,omitempty"`
}

// CostInfo represents cost information
type CostInfo struct {
	InputCost        float64         `json:"input_cost"`
	OutputCost       float64         `json:"output_cost"`
	TotalCost        float64         `json:"total_cost"`
	Currency         string          `json:"currency"`
	
	// Cost breakdown
	ModelCost        float64         `json:"model_cost,omitempty"`
	ProcessingCost   float64         `json:"processing_cost,omitempty"`
	StorageCost      float64         `json:"storage_cost,omitempty"`
	
	// Billing period
	BillingPeriod    string          `json:"billing_period,omitempty"`
	EstimatedCost    float64         `json:"estimated_cost,omitempty"`
}

// HealthStatus represents provider health status
type HealthStatus struct {
	Status           string          `json:"status"` // healthy, degraded, unhealthy
	Latency          time.Duration   `json:"latency"`
	ErrorRate        float64         `json:"error_rate"`
	LastChecked      time.Time       `json:"last_checked"`
	
	// Detailed status
	APIStatus        string          `json:"api_status"`
	ModelStatus      map[string]string `json:"model_status"`
	
	// Issues
	Issues           []string        `json:"issues,omitempty"`
	Warnings         []string        `json:"warnings,omitempty"`
}

// RateLimits represents rate limiting information
type RateLimits struct {
	RequestsPerMinute int             `json:"requests_per_minute"`
	TokensPerMinute   int             `json:"tokens_per_minute"`
	RequestsPerDay    int             `json:"requests_per_day"`
	TokensPerDay      int             `json:"tokens_per_day"`
	
	// Current usage
	CurrentRequests   int             `json:"current_requests"`
	CurrentTokens     int             `json:"current_tokens"`
	
	// Reset times
	RequestsReset     time.Time       `json:"requests_reset"`
	TokensReset       time.Time       `json:"tokens_reset"`
	
	// Remaining
	RemainingRequests int             `json:"remaining_requests"`
	RemainingTokens   int             `json:"remaining_tokens"`
}

// SafetySetting represents safety configuration
type SafetySetting struct {
	Category  string            `json:"category"`
	Threshold string            `json:"threshold"`
}

// SafetyRating represents safety rating for content
type SafetyRating struct {
	Category    string            `json:"category"`
	Probability string            `json:"probability"`
	Blocked     bool              `json:"blocked"`
}

// LogProbs represents log probabilities
type LogProbs struct {
	Tokens        []string          `json:"tokens"`
	TokenLogProbs []float64         `json:"token_logprobs"`
	TopLogProbs   []map[string]float64 `json:"top_logprobs"`
}

// TimeRange represents a time range for usage queries
type TimeRange struct {
	Start time.Time             `json:"start"`
	End   time.Time             `json:"end"`
}

// ProviderConfig represents provider configuration
type ProviderConfig interface {
	GetAPIKey() string
	GetBaseURL() string
	GetTimeout() time.Duration
	GetRetryAttempts() int
	GetModel() string
	Validate() error
}

// StreamReader provides a convenient interface for reading streaming responses
type StreamReader interface {
	io.Reader
	ReadResponse() (*StreamResponse, error)
	Close() error
}

// ProviderFactory creates provider instances
type ProviderFactory interface {
	CreateProvider(providerType ProviderType, config ProviderConfig) (Provider, error)
	SupportedProviders() []ProviderType
}
