# AI Provider Integration System

This package provides comprehensive AI provider integration for the Go Coffee AI agents, featuring multi-provider support, intelligent routing, observability, and resilient AI operations with OpenAI and Google Gemini.

## Overview

The AI integration system implements:

1. **Multi-Provider Support**: OpenAI and Google Gemini integration with unified interface
2. **Intelligent Routing**: Provider selection strategies for optimal performance and cost
3. **Observability Integration**: Full tracing, metrics, and logging for all AI operations
4. **Resilience Patterns**: Circuit breakers, retries, and rate limiting
5. **Domain-Specific Operations**: Beverage generation, task creation, and content generation
6. **Usage Tracking**: Comprehensive token usage and cost tracking

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ AI Manager      │    │ Provider        │    │ Selection       │
│ • Multi-provider│───▶│ Registry        │    │ Strategies      │
│ • Coordination  │    │ • OpenAI        │    │ • Round Robin   │
└─────────────────┘    │ • Gemini        │    │ • Cost Optimized│
         │              └─────────────────┘    └─────────────────┘
         ▼                       │                       │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Base Provider   │    │ Circuit Breaker │    │ Rate Limiter    │
│ • Common Logic  │    │ • Fault Tolerance│   │ • Request Control│
│ • Observability │    │ • Auto Recovery │    │ • Token Bucket  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Components

### 1. AI Manager (`manager.go`)

Central coordinator for all AI operations:

```go
// Initialize AI manager
aiManager := ai.NewManager(aiConfig, logger, metrics, tracing)
if err := aiManager.Initialize(ctx); err != nil {
    log.Fatal("Failed to initialize AI manager:", err)
}

// Generate text using best available provider
response, err := aiManager.GenerateText(ctx, &ai.TextGenerationRequest{
    Model:       "gpt-4",
    Prompt:      "Create a coffee recipe for Mars colonists",
    MaxTokens:   500,
    Temperature: 0.8,
})

// Generate beverage recipes
beverageResponse, err := aiManager.GenerateBeverage(ctx, &ai.BeverageGenerationRequest{
    Theme:       "Mars Base",
    Ingredients: []string{"coffee", "protein powder"},
    Complexity:  "medium",
    Count:       3,
})
```

### 2. Provider Interface (`types.go`)

Unified interface for all AI providers:

```go
type Provider interface {
    GetName() string
    GetModels() []Model
    GenerateText(ctx context.Context, request *TextGenerationRequest) (*TextGenerationResponse, error)
    GenerateChat(ctx context.Context, request *ChatRequest) (*ChatResponse, error)
    GenerateEmbedding(ctx context.Context, request *EmbeddingRequest) (*EmbeddingResponse, error)
    HealthCheck(ctx context.Context) error
    GetUsage() *UsageStatistics
    Close() error
}
```

### 3. OpenAI Provider (`providers/openai/`)

Full OpenAI API integration:

```go
// OpenAI provider with comprehensive model support
type OpenAIProvider struct {
    *ai.BaseProvider
    client  *http.Client
    baseURL string
    apiKey  string
}

// Supported models
models := []ai.Model{
    {
        ID:           "gpt-4",
        Name:         "GPT-4",
        Type:         ai.ModelTypeChat,
        MaxTokens:    8192,
        InputCost:    0.03,  // $0.03 per 1K tokens
        OutputCost:   0.06,  // $0.06 per 1K tokens
        Capabilities: []string{"chat", "text", "reasoning"},
    },
    {
        ID:           "gpt-4-turbo",
        Name:         "GPT-4 Turbo",
        Type:         ai.ModelTypeChat,
        MaxTokens:    128000,
        InputCost:    0.01,
        OutputCost:   0.03,
        Capabilities: []string{"chat", "text", "reasoning", "vision"},
    },
}
```

### 4. Gemini Provider (`providers/gemini/`)

Google Gemini API integration:

```go
// Gemini provider with multimodal capabilities
type GeminiProvider struct {
    *ai.BaseProvider
    client  *http.Client
    baseURL string
    apiKey  string
}

// Supported models
models := []ai.Model{
    {
        ID:           "gemini-pro",
        Name:         "Gemini Pro",
        Type:         ai.ModelTypeChat,
        MaxTokens:    32768,
        InputCost:    0.0005,  // $0.0005 per 1K tokens
        OutputCost:   0.0015,  // $0.0015 per 1K tokens
        Capabilities: []string{"chat", "text", "reasoning"},
    },
    {
        ID:           "gemini-pro-vision",
        Name:         "Gemini Pro Vision",
        Type:         ai.ModelTypeChat,
        MaxTokens:    16384,
        InputCost:    0.00025,
        OutputCost:   0.0005,
        Capabilities: []string{"chat", "text", "vision", "multimodal"},
    },
}
```

### 5. Provider Selection Strategies (`strategies.go`)

Intelligent provider selection algorithms:

#### Round Robin Strategy
```go
strategy := ai.NewRoundRobinStrategy()
provider, err := strategy.SelectProvider(providers, ai.ModelTypeChat)
```

#### Cost Optimized Strategy
```go
strategy := ai.NewCostOptimizedStrategy()
provider, err := strategy.SelectProvider(providers, ai.ModelTypeChat)
```

#### Performance Optimized Strategy
```go
strategy := ai.NewPerformanceOptimizedStrategy()
provider, err := strategy.SelectProvider(providers, ai.ModelTypeChat)
```

#### Weighted Strategy
```go
weights := map[string]float64{
    "openai": 0.7,
    "gemini": 0.3,
}
strategy := ai.NewWeightedStrategy(weights)
provider, err := strategy.SelectProvider(providers, ai.ModelTypeChat)
```

#### Failover Strategy
```go
strategy := ai.NewFailoverStrategy("openai", []string{"gemini"})
provider, err := strategy.SelectProvider(providers, ai.ModelTypeChat)
```

## Domain-Specific Operations

### Beverage Generation

```go
// Generate beverage recipes with AI
request := &ai.BeverageGenerationRequest{
    Theme:        "Mars Base",
    Ingredients:  []string{"coffee", "protein powder", "freeze-dried fruit"},
    Dietary:      []string{"vegan", "low-sugar"},
    Complexity:   "medium",
    Style:        "modern",
    Temperature:  "hot",
    Count:        3,
}

response, err := aiManager.GenerateBeverage(ctx, request)
if err != nil {
    return fmt.Errorf("failed to generate beverages: %w", err)
}

for _, beverage := range response.Beverages {
    fmt.Printf("Generated: %s - %s\n", beverage.Name, beverage.Description)
    fmt.Printf("Prep time: %d minutes, Difficulty: %s\n", 
        beverage.PrepTime, beverage.Difficulty)
}
```

### Task Generation

```go
// Generate tasks with AI
request := &ai.TaskGenerationRequest{
    Context:  "Mars colony beverage production",
    Goal:     "Establish sustainable beverage supply chain",
    Priority: "high",
    Skills:   []string{"food science", "logistics", "automation"},
    Count:    5,
}

response, err := aiManager.GenerateTask(ctx, request)
```

### Chat Completion

```go
// Multi-turn conversation
request := &ai.ChatRequest{
    Model: "gpt-4",
    Messages: []ai.ChatMessage{
        {
            Role:    "system",
            Content: "You are an expert beverage scientist working on Mars.",
        },
        {
            Role:    "user",
            Content: "How can we create coffee using local Martian resources?",
        },
    },
    MaxTokens:   1000,
    Temperature: 0.7,
}

response, err := aiManager.GenerateChat(ctx, request)
```

## Configuration

### AI Configuration

```yaml
ai:
  openai:
    enabled: true
    api_key: ${OPENAI_API_KEY}
    base_url: https://api.openai.com/v1
    timeout: 30s
    max_retries: 3
    retry_delay: 1s
    models:
      - gpt-4
      - gpt-4-turbo
      - gpt-3.5-turbo
  
  gemini:
    enabled: true
    api_key: ${GEMINI_API_KEY}
    base_url: https://generativelanguage.googleapis.com/v1beta
    timeout: 30s
    max_retries: 3
    retry_delay: 1s
    models:
      - gemini-pro
      - gemini-pro-vision
  
  strategy: cost_optimized  # round_robin, cost_optimized, performance_optimized
  circuit_breaker:
    enabled: true
    failure_threshold: 5
    timeout: 60s
  rate_limiting:
    enabled: true
    requests_per_second: 10
```

### Environment Variables

```bash
# OpenAI
GOCOFFEE_AI_OPENAI_API_KEY=sk-your-openai-key
GOCOFFEE_AI_OPENAI_ENABLED=true

# Gemini
GOCOFFEE_AI_GEMINI_API_KEY=your-gemini-key
GOCOFFEE_AI_GEMINI_ENABLED=true

# Strategy
GOCOFFEE_AI_STRATEGY=cost_optimized
GOCOFFEE_AI_CIRCUIT_BREAKER_ENABLED=true
GOCOFFEE_AI_RATE_LIMITING_ENABLED=true
```

## Usage Examples

### Basic Text Generation

```go
// Initialize AI manager
aiManager := ai.GetGlobalManager()

// Generate text
response, err := aiManager.GenerateText(ctx, &ai.TextGenerationRequest{
    Model:       "gpt-4",
    Prompt:      "Describe the perfect coffee for a Mars colony",
    MaxTokens:   300,
    Temperature: 0.8,
})

if err != nil {
    return fmt.Errorf("text generation failed: %w", err)
}

fmt.Printf("Generated text: %s\n", response.Text)
fmt.Printf("Tokens used: %d, Cost: $%.4f\n", 
    response.Usage.TotalTokens, 
    calculateCost(response.Usage, response.Model))
```

### Multi-Provider Health Check

```go
// Check health of all providers
healthResults := aiManager.HealthCheck(ctx)

for provider, err := range healthResults {
    if err != nil {
        log.Printf("Provider %s is unhealthy: %v", provider, err)
    } else {
        log.Printf("Provider %s is healthy", provider)
    }
}
```

### Usage Statistics

```go
// Get aggregated usage statistics
usage := aiManager.GetAggregatedUsage()

fmt.Printf("Total requests: %d\n", usage.TotalRequests)
fmt.Printf("Success rate: %.2f%%\n", 
    float64(usage.SuccessfulReqs)/float64(usage.TotalRequests)*100)
fmt.Printf("Total cost: $%.2f\n", usage.TotalCost)
fmt.Printf("Total tokens: %d\n", usage.TotalTokens)

// Per-provider statistics
for provider, stats := range usage.ProviderUsage {
    fmt.Printf("Provider %s: %d requests, $%.2f cost\n", 
        provider, stats.TotalRequests, stats.TotalCost)
}
```

## Observability

### Metrics

AI operations are automatically instrumented:

- `ai_requests_total`: Total AI requests by provider and operation
- `ai_requests_success_total`: Successful AI requests
- `ai_requests_error_total`: Failed AI requests
- `ai_request_duration_seconds`: AI request duration histogram

### Tracing

All AI operations are traced:

```go
// Automatic span creation
ctx, span := tracing.StartAISpan(ctx, "GENERATE_TEXT", "openai", "gpt-4")
defer span.End()

// Span attributes
span.SetAttributes(
    attribute.String("ai.provider", "openai"),
    attribute.String("ai.model", "gpt-4"),
    attribute.String("ai.operation", "text_generation"),
    attribute.Int("ai.tokens", response.Usage.TotalTokens),
    attribute.Float64("ai.cost", cost),
)
```

### Logging

Structured logging with trace correlation:

```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "level": "info",
  "message": "AI request completed successfully",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "provider": "openai",
  "model": "gpt-4",
  "operation": "text_generation",
  "tokens": 150,
  "cost": 0.0045,
  "duration_ms": 1250
}
```

## Error Handling

### AI Error Types

```go
// Typed errors for different failure modes
type AIError struct {
    Code     string `json:"code"`
    Message  string `json:"message"`
    Provider string `json:"provider"`
    Model    string `json:"model,omitempty"`
    Retryable bool  `json:"retryable"`
}

// Common error codes
const (
    ErrorCodeInvalidRequest     = "invalid_request"
    ErrorCodeAuthentication     = "authentication_error"
    ErrorCodeRateLimit          = "rate_limit_exceeded"
    ErrorCodeQuotaExceeded      = "quota_exceeded"
    ErrorCodeModelNotFound      = "model_not_found"
    ErrorCodeTimeout            = "timeout"
    ErrorCodeInternalError      = "internal_error"
    ErrorCodeServiceUnavailable = "service_unavailable"
)
```

### Retry Logic

```go
// Automatic retry with exponential backoff
func (bp *BaseProvider) ShouldRetry(err error, retryCount int) bool {
    if retryCount >= bp.config.MaxRetries {
        return false
    }
    
    if aiErr, ok := err.(*AIError); ok {
        return aiErr.Retryable
    }
    
    return true
}

func (bp *BaseProvider) CalculateRetryDelay(retryCount int) time.Duration {
    baseDelay := bp.config.RetryDelay
    delay := time.Duration(float64(baseDelay) * (1 << retryCount))
    
    // Add jitter and cap at maximum
    jitter := time.Duration(float64(delay) * 0.25)
    delay += time.Duration(float64(jitter) * (2*time.Now().UnixNano()%1000/1000.0 - 1))
    
    maxDelay := 30 * time.Second
    if delay > maxDelay {
        delay = maxDelay
    }
    
    return delay
}
```

## Best Practices

### 1. Provider Selection
- Use cost-optimized strategy for batch operations
- Use performance-optimized strategy for real-time operations
- Implement failover for critical operations
- Monitor provider health and switch automatically

### 2. Token Management
- Track token usage and costs per operation
- Implement token budgets and alerts
- Use appropriate max_tokens limits
- Monitor and optimize prompt efficiency

### 3. Error Handling
- Implement proper retry logic with exponential backoff
- Use circuit breakers for fault tolerance
- Log errors with sufficient context
- Provide graceful degradation

### 4. Performance
- Use appropriate temperature settings for use case
- Implement request batching where possible
- Cache responses for repeated queries
- Monitor and optimize request latency

### 5. Security
- Secure API key storage and rotation
- Implement request validation and sanitization
- Monitor for unusual usage patterns
- Use appropriate content filtering

This AI provider integration provides a robust, observable, and scalable foundation for AI operations in the Go Coffee AI agent ecosystem, enabling intelligent beverage creation, task generation, and content creation with enterprise-grade reliability and performance.
