package providers

import (
	"fmt"
	"strings"
	"time"
)

// Common error types for AI providers
var (
	ErrProviderNotFound     = fmt.Errorf("provider not found")
	ErrModelNotFound        = fmt.Errorf("model not found")
	ErrInvalidRequest       = fmt.Errorf("invalid request")
	ErrRateLimitExceeded    = fmt.Errorf("rate limit exceeded")
	ErrQuotaExceeded        = fmt.Errorf("quota exceeded")
	ErrProviderUnavailable  = fmt.Errorf("provider unavailable")
	ErrInvalidAPIKey        = fmt.Errorf("invalid API key")
	ErrContentFiltered      = fmt.Errorf("content filtered")
	ErrTokenLimitExceeded   = fmt.Errorf("token limit exceeded")
	ErrInvalidModel         = fmt.Errorf("invalid model")
)

// ProviderError represents a provider-specific error
type ProviderError struct {
	Provider    string            `json:"provider"`
	Code        string            `json:"code"`
	Message     string            `json:"message"`
	Type        string            `json:"type"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Retryable   bool              `json:"retryable"`
	StatusCode  int               `json:"status_code,omitempty"`
	Cause       error             `json:"-"`
}

// Error implements the error interface
func (e *ProviderError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Provider, e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *ProviderError) Unwrap() error {
	return e.Cause
}

// IsRetryable returns whether the error is retryable
func (e *ProviderError) IsRetryable() bool {
	return e.Retryable
}

// NewProviderError creates a new provider error
func NewProviderError(provider, code, message string, retryable bool) *ProviderError {
	return &ProviderError{
		Provider:  provider,
		Code:      code,
		Message:   message,
		Retryable: retryable,
	}
}

// RequestMetadata contains metadata for tracking requests
type RequestMetadata struct {
	RequestID     string            `json:"request_id"`
	UserID        string            `json:"user_id"`
	SessionID     string            `json:"session_id"`
	Timestamp     time.Time         `json:"timestamp"`
	Provider      string            `json:"provider"`
	Model         string            `json:"model"`
	UseCase       string            `json:"use_case"`
	Tags          map[string]string `json:"tags"`
	
	// Performance tracking
	StartTime     time.Time         `json:"start_time"`
	EndTime       time.Time         `json:"end_time"`
	Duration      time.Duration     `json:"duration"`
	
	// Cost tracking
	EstimatedCost float64           `json:"estimated_cost"`
	ActualCost    float64           `json:"actual_cost"`
	
	// Caching
	CacheHit      bool              `json:"cache_hit"`
	CacheKey      string            `json:"cache_key"`
}

// ModelCapabilities represents the capabilities of an AI model
type ModelCapabilities struct {
	TextGeneration    bool              `json:"text_generation"`
	ChatCompletion    bool              `json:"chat_completion"`
	FunctionCalling   bool              `json:"function_calling"`
	ImageAnalysis     bool              `json:"image_analysis"`
	ImageGeneration   bool              `json:"image_generation"`
	CodeGeneration    bool              `json:"code_generation"`
	Embedding         bool              `json:"embedding"`
	Streaming         bool              `json:"streaming"`
	
	// Context and token limits
	MaxContextTokens  int               `json:"max_context_tokens"`
	MaxOutputTokens   int               `json:"max_output_tokens"`
	
	// Supported languages
	Languages         []string          `json:"languages"`
	
	// Supported formats
	InputFormats      []string          `json:"input_formats"`
	OutputFormats     []string          `json:"output_formats"`
}

// ProviderCapabilities represents the capabilities of a provider
type ProviderCapabilities struct {
	Models            []string          `json:"models"`
	Streaming         bool              `json:"streaming"`
	FunctionCalling   bool              `json:"function_calling"`
	ImageAnalysis     bool              `json:"image_analysis"`
	ImageGeneration   bool              `json:"image_generation"`
	Embedding         bool              `json:"embedding"`
	
	// Rate limits
	MaxRequestsPerMinute int            `json:"max_requests_per_minute"`
	MaxTokensPerMinute   int            `json:"max_tokens_per_minute"`
	MaxConcurrentRequests int           `json:"max_concurrent_requests"`
	
	// Pricing
	HasFreeTier       bool              `json:"has_free_tier"`
	PayPerUse         bool              `json:"pay_per_use"`
	SubscriptionBased bool              `json:"subscription_based"`
}

// RequestOptions contains options for customizing requests
type RequestOptions struct {
	// Timeout settings
	Timeout           time.Duration     `json:"timeout,omitempty"`
	
	// Retry settings
	MaxRetries        int               `json:"max_retries,omitempty"`
	RetryDelay        time.Duration     `json:"retry_delay,omitempty"`
	
	// Caching
	EnableCache       bool              `json:"enable_cache,omitempty"`
	CacheTTL          time.Duration     `json:"cache_ttl,omitempty"`
	
	// Rate limiting
	BypassRateLimit   bool              `json:"bypass_rate_limit,omitempty"`
	Priority          int               `json:"priority,omitempty"`
	
	// Fallback
	FallbackProviders []string          `json:"fallback_providers,omitempty"`
	
	// Monitoring
	TrackCost         bool              `json:"track_cost,omitempty"`
	TrackUsage        bool              `json:"track_usage,omitempty"`
	
	// Safety
	ContentFilter     bool              `json:"content_filter,omitempty"`
	SafetyLevel       string            `json:"safety_level,omitempty"`
}

// ResponseMetadata contains metadata about the response
type ResponseMetadata struct {
	// Timing
	ProcessingTime    time.Duration     `json:"processing_time"`
	QueueTime         time.Duration     `json:"queue_time"`
	NetworkTime       time.Duration     `json:"network_time"`
	
	// Provider info
	Provider          string            `json:"provider"`
	Model             string            `json:"model"`
	ModelVersion      string            `json:"model_version"`
	
	// Quality metrics
	Confidence        float64           `json:"confidence,omitempty"`
	Relevance         float64           `json:"relevance,omitempty"`
	
	// Caching
	FromCache         bool              `json:"from_cache"`
	CacheAge          time.Duration     `json:"cache_age,omitempty"`
	
	// Rate limiting
	RateLimitRemaining int              `json:"rate_limit_remaining,omitempty"`
	RateLimitReset     time.Time        `json:"rate_limit_reset,omitempty"`
	
	// Cost
	TokensUsed        int               `json:"tokens_used"`
	Cost              float64           `json:"cost"`
	
	// Warnings
	Warnings          []string          `json:"warnings,omitempty"`
}

// PromptTemplate represents a reusable prompt template
type PromptTemplate struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Template    string            `json:"template"`
	Variables   []TemplateVariable `json:"variables"`
	UseCase     string            `json:"use_case"`
	Provider    string            `json:"provider,omitempty"`
	Model       string            `json:"model,omitempty"`
	
	// Default parameters
	DefaultParams map[string]interface{} `json:"default_params,omitempty"`
	
	// Validation
	RequiredVars  []string          `json:"required_vars"`
	OptionalVars  []string          `json:"optional_vars"`
	
	// Metadata
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Version       string            `json:"version"`
	Tags          []string          `json:"tags"`
}

// TemplateVariable represents a variable in a prompt template
type TemplateVariable struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Description  string            `json:"description"`
	Required     bool              `json:"required"`
	DefaultValue interface{}       `json:"default_value,omitempty"`
	Validation   string            `json:"validation,omitempty"`
	Examples     []string          `json:"examples,omitempty"`
}

// RenderTemplate renders a prompt template with variables
func (pt *PromptTemplate) RenderTemplate(variables map[string]interface{}) (string, error) {
	result := pt.Template
	
	// Check required variables
	for _, required := range pt.RequiredVars {
		if _, exists := variables[required]; !exists {
			return "", fmt.Errorf("required variable '%s' not provided", required)
		}
	}
	
	// Replace variables in template
	for name, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", name)
		valueStr := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, valueStr)
	}
	
	// Check for unreplaced variables
	if strings.Contains(result, "{{") && strings.Contains(result, "}}") {
		return "", fmt.Errorf("template contains unreplaced variables")
	}
	
	return result, nil
}

// Validate validates the prompt template
func (pt *PromptTemplate) Validate() error {
	if pt.ID == "" {
		return fmt.Errorf("template ID is required")
	}
	
	if pt.Name == "" {
		return fmt.Errorf("template name is required")
	}
	
	if pt.Template == "" {
		return fmt.Errorf("template content is required")
	}
	
	// Validate variables
	for _, variable := range pt.Variables {
		if variable.Name == "" {
			return fmt.Errorf("variable name is required")
		}
		
		if variable.Type == "" {
			return fmt.Errorf("variable type is required for variable '%s'", variable.Name)
		}
	}
	
	return nil
}

// ConversationContext represents the context of a conversation
type ConversationContext struct {
	ID            string            `json:"id"`
	UserID        string            `json:"user_id"`
	SessionID     string            `json:"session_id"`
	Messages      []Message         `json:"messages"`
	SystemPrompt  string            `json:"system_prompt,omitempty"`
	
	// Context management
	MaxMessages   int               `json:"max_messages"`
	MaxTokens     int               `json:"max_tokens"`
	
	// Metadata
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	LastActivity  time.Time         `json:"last_activity"`
	
	// Settings
	Provider      string            `json:"provider,omitempty"`
	Model         string            `json:"model,omitempty"`
	Temperature   float64           `json:"temperature,omitempty"`
	
	// State
	Active        bool              `json:"active"`
	Archived      bool              `json:"archived"`
	
	// Custom data
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// AddMessage adds a message to the conversation context
func (cc *ConversationContext) AddMessage(message Message) {
	cc.Messages = append(cc.Messages, message)
	cc.UpdatedAt = time.Now()
	cc.LastActivity = time.Now()
	
	// Trim messages if exceeding max
	if cc.MaxMessages > 0 && len(cc.Messages) > cc.MaxMessages {
		// Keep system message if it exists
		systemMessages := []Message{}
		userMessages := []Message{}
		
		for _, msg := range cc.Messages {
			if msg.Role == RoleSystem {
				systemMessages = append(systemMessages, msg)
			} else {
				userMessages = append(userMessages, msg)
			}
		}
		
		// Keep recent user messages
		if len(userMessages) > cc.MaxMessages-len(systemMessages) {
			keepCount := cc.MaxMessages - len(systemMessages)
			userMessages = userMessages[len(userMessages)-keepCount:]
		}
		
		cc.Messages = append(systemMessages, userMessages...)
	}
}

// GetTokenCount estimates the token count for the conversation
func (cc *ConversationContext) GetTokenCount() int {
	totalTokens := 0
	
	for _, message := range cc.Messages {
		// Rough estimation: 1 token per 4 characters
		totalTokens += len(message.Content) / 4
		
		// Add overhead for message structure
		totalTokens += 10
	}
	
	return totalTokens
}

// TrimToTokenLimit trims the conversation to fit within token limit
func (cc *ConversationContext) TrimToTokenLimit(maxTokens int) {
	if cc.MaxTokens > 0 && maxTokens > cc.MaxTokens {
		maxTokens = cc.MaxTokens
	}
	
	currentTokens := cc.GetTokenCount()
	if currentTokens <= maxTokens {
		return
	}
	
	// Keep system messages
	systemMessages := []Message{}
	otherMessages := []Message{}
	
	for _, msg := range cc.Messages {
		if msg.Role == RoleSystem {
			systemMessages = append(systemMessages, msg)
		} else {
			otherMessages = append(otherMessages, msg)
		}
	}
	
	// Calculate tokens for system messages
	systemTokens := 0
	for _, msg := range systemMessages {
		systemTokens += len(msg.Content)/4 + 10
	}
	
	// Trim other messages from the beginning
	availableTokens := maxTokens - systemTokens
	keptMessages := []Message{}
	
	for i := len(otherMessages) - 1; i >= 0; i-- {
		msgTokens := len(otherMessages[i].Content)/4 + 10
		if availableTokens >= msgTokens {
			keptMessages = append([]Message{otherMessages[i]}, keptMessages...)
			availableTokens -= msgTokens
		} else {
			break
		}
	}
	
	cc.Messages = append(systemMessages, keptMessages...)
	cc.UpdatedAt = time.Now()
}

// Clone creates a copy of the conversation context
func (cc *ConversationContext) Clone() *ConversationContext {
	clone := *cc
	
	// Deep copy messages
	clone.Messages = make([]Message, len(cc.Messages))
	copy(clone.Messages, cc.Messages)
	
	// Deep copy metadata
	if cc.Metadata != nil {
		clone.Metadata = make(map[string]interface{})
		for k, v := range cc.Metadata {
			clone.Metadata[k] = v
		}
	}
	
	return &clone
}
