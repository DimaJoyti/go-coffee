package common

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

// ProviderStrategy defines the interface for provider selection strategies
type ProviderStrategy interface {
	// SelectProvider selects a provider from the available providers for the given model type
	SelectProvider(providers []Provider, modelType ModelType) (Provider, error)
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