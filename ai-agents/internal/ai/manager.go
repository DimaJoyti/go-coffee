package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go-coffee-ai-agents/internal/ai/cache"
	"go-coffee-ai-agents/internal/ai/cost"
	"go-coffee-ai-agents/internal/ai/monitoring"
	"go-coffee-ai-agents/internal/ai/providers/gemini"
	"go-coffee-ai-agents/internal/ai/providers/openai"
	"go-coffee-ai-agents/internal/ai/ratelimit"
	"go-coffee-ai-agents/internal/common"
	"go-coffee-ai-agents/internal/config"
	"go-coffee-ai-agents/internal/observability"
)

// Manager manages multiple AI providers and provides unified access
type Manager struct {
	providers map[string]common.Provider
	config    config.AIConfig
	logger    *observability.StructuredLogger
	metrics   *observability.MetricsCollector
	tracing   *observability.TracingHelper

	// Enhanced components
	costTracker      *cost.Tracker
	rateLimiter      *ratelimit.Limiter
	cache            *cache.Cache
	metricsCollector *monitoring.MetricsCollector

	// Provider selection strategy
	strategy ProviderStrategy

	// Thread safety
	mutex sync.RWMutex
}

// ProviderStrategy defines how to select providers
type ProviderStrategy interface {
	SelectProvider(providers []common.Provider, modelType common.ModelType) (common.Provider, error)
}

// NewManager creates a new AI provider manager
func NewManager(
	config config.AIConfig,
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) *Manager {
	return &Manager{
		providers: make(map[string]common.Provider),
		config:    config,
		logger:    logger,
		metrics:   metrics,
		tracing:   tracing,
		strategy:  NewRoundRobinStrategy(),
	}
}

// Initialize initializes the AI manager and all configured providers
func (m *Manager) Initialize(ctx context.Context) error {
	ctx, span := m.tracing.StartAISpan(ctx, "INITIALIZE", "ai_manager", "")
	defer span.End()

	m.logger.InfoContext(ctx, "Initializing AI manager")

	// Initialize configured providers
	for providerName, providerConfig := range m.config.Providers {
		if !providerConfig.Enabled {
			continue
		}

		var provider common.Provider

		switch providerName {
		case "openai":
			provider = openai.NewOpenAIProvider(
				common.ProviderConfig{
					Name:       providerName,
					Type:       "openai",
					APIKey:     providerConfig.APIKey,
					BaseURL:    providerConfig.BaseURL,
					Timeout:    providerConfig.Timeout,
					MaxRetries: 3,              // Default value
					RetryDelay: time.Second,    // Default value
					Enabled:    providerConfig.Enabled,
				},
				m.logger,
				m.metrics,
				m.tracing,
			)
		case "gemini":
			provider = gemini.NewGeminiProvider(
				common.ProviderConfig{
					Name:       providerName,
					Type:       "gemini",
					APIKey:     providerConfig.APIKey,
					BaseURL:    providerConfig.BaseURL,
					Timeout:    providerConfig.Timeout,
					MaxRetries: 3,              // Default value
					RetryDelay: time.Second,    // Default value
					Enabled:    providerConfig.Enabled,
				},
				m.logger,
				m.metrics,
				m.tracing,
			)
		default:
			m.logger.WarnContext(ctx, "Unknown provider type", "provider", providerName)
			continue
		}

		if err := m.RegisterProvider(provider); err != nil {
			m.tracing.RecordError(span, err, fmt.Sprintf("Failed to register %s provider", providerName))
			return fmt.Errorf("failed to register %s provider: %w", providerName, err)
		}
	}

	m.tracing.RecordSuccess(span, "AI manager initialized successfully")
	m.logger.InfoContext(ctx, "AI manager initialized successfully",
		"providers_count", len(m.providers))

	return nil
}

// RegisterProvider registers a new AI provider
func (m *Manager) RegisterProvider(provider common.Provider) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	name := provider.GetName()
	if _, exists := m.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	m.providers[name] = provider

	m.logger.Info("Registered AI provider",
		"provider", name,
		"models_count", len(provider.GetModels()))

	return nil
}

// GetProvider returns a provider by name
func (m *Manager) GetProvider(name string) (common.Provider, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	provider, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// GetProviders returns all registered providers
func (m *Manager) GetProviders() []common.Provider {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	providers := make([]common.Provider, 0, len(m.providers))
	for _, provider := range m.providers {
		providers = append(providers, provider)
	}

	return providers
}

// GetBestProvider returns the best provider for a given model type
func (m *Manager) GetBestProvider(modelType common.ModelType) (common.Provider, error) {
	providers := m.GetProviders()
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	// Filter providers that support the model type
	var supportedProviders []common.Provider
	for _, provider := range providers {
		models := provider.GetModels()
		for _, model := range models {
			if model.Type == modelType {
				supportedProviders = append(supportedProviders, provider)
				break
			}
		}
	}

	if len(supportedProviders) == 0 {
		return nil, fmt.Errorf("no providers support model type %s", modelType)
	}

	return m.strategy.SelectProvider(supportedProviders, modelType)
}

// GenerateText generates text using the best available provider
func (m *Manager) GenerateText(ctx context.Context, request *common.TextGenerationRequest) (*common.TextGenerationResponse, error) {
	provider, err := m.getProviderForModel(request.Model)
	if err != nil {
		return nil, err
	}

	return provider.GenerateText(ctx, request)
}

// GenerateChat generates chat completion using the best available provider
func (m *Manager) GenerateChat(ctx context.Context, request *common.ChatRequest) (*common.ChatResponse, error) {
	provider, err := m.getProviderForModel(request.Model)
	if err != nil {
		return nil, err
	}

	return provider.GenerateChat(ctx, request)
}

// GenerateEmbedding generates embeddings using the best available provider
func (m *Manager) GenerateEmbedding(ctx context.Context, request *common.EmbeddingRequest) (*common.EmbeddingResponse, error) {
	provider, err := m.getProviderForModel(request.Model)
	if err != nil {
		return nil, err
	}

	return provider.GenerateEmbedding(ctx, request)
}

// GenerateBeverage generates beverage recipes using AI
func (m *Manager) GenerateBeverage(ctx context.Context, request *common.BeverageGenerationRequest) (*common.BeverageGenerationResponse, error) {
	ctx, span := m.tracing.StartAISpan(ctx, "GENERATE_BEVERAGE", "ai_manager", "")
	defer span.End()

	// Use the best chat provider
	provider, err := m.GetBestProvider(common.ModelTypeChat)
	if err != nil {
		m.tracing.RecordError(span, err, "No suitable provider for beverage generation")
		return nil, fmt.Errorf("no suitable provider for beverage generation: %w", err)
	}

	// Build prompt for beverage generation
	prompt := m.buildBeveragePrompt(request)

	// Create chat request
	chatRequest := &common.ChatRequest{
		Model: m.getBestModelForProvider(provider, common.ModelTypeChat),
		Messages: []common.ChatMessage{
			{
				Role:    "system",
				Content: "You are an expert beverage creator. Generate creative and delicious beverage recipes based on the given requirements. Return the response in JSON format.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   2000,
		Temperature: 0.8,
	}

	// Generate response
	chatResponse, err := provider.GenerateChat(ctx, chatRequest)
	if err != nil {
		m.tracing.RecordError(span, err, "Failed to generate beverage")
		return nil, fmt.Errorf("failed to generate beverage: %w", err)
	}

	// Parse response (simplified - in real implementation, parse JSON)
	response := &common.BeverageGenerationResponse{
		ID:        chatResponse.ID,
		Model:     chatResponse.Model,
		Provider:  chatResponse.Provider,
		Usage:     chatResponse.Usage,
		CreatedAt: time.Now(),
		Beverages: []common.GeneratedBeverage{
			{
				Name:        "AI Generated Beverage",
				Description: chatResponse.Message.Content,
				Ingredients: []common.BeverageIngredient{
					{Name: "Base ingredient", Quantity: "1 cup", Type: "base"},
				},
				Instructions: []string{"Mix ingredients", "Serve"},
				PrepTime:     10,
				Servings:     1,
				Difficulty:   "easy",
				Tags:         []string{request.Theme},
			},
		},
	}

	m.tracing.RecordSuccess(span, "Beverage generated successfully")
	return response, nil
}

// HealthCheck performs health checks on all providers
func (m *Manager) HealthCheck(ctx context.Context) map[string]error {
	ctx, span := m.tracing.StartAISpan(ctx, "HEALTH_CHECK", "ai_manager", "")
	defer span.End()

	providers := m.GetProviders()
	results := make(map[string]error)

	for _, provider := range providers {
		err := provider.HealthCheck(ctx)
		results[provider.GetName()] = err

		if err != nil {
			m.logger.WarnContext(ctx, "Provider health check failed",
				"provider", provider.GetName(),
				"error", err.Error())
		} else {
			m.logger.DebugContext(ctx, "Provider health check passed",
				"provider", provider.GetName())
		}
	}

	m.tracing.RecordSuccess(span, "Health check completed")
	return results
}

// GetAggregatedUsage returns aggregated usage statistics
func (m *Manager) GetAggregatedUsage() *common.AggregatedUsage {
	providers := m.GetProviders()

	aggregated := &common.AggregatedUsage{
		ProviderUsage: make(map[string]*common.UsageStatistics),
		LastUpdated:   time.Now(),
	}

	for _, provider := range providers {
		usage := provider.GetUsage()
		aggregated.ProviderUsage[provider.GetName()] = usage

		aggregated.TotalRequests += usage.TotalRequests
		aggregated.SuccessfulReqs += usage.SuccessfulReqs
		aggregated.FailedRequests += usage.FailedRequests
		aggregated.TotalTokens += usage.TotalTokens
		aggregated.TotalCost += usage.TotalCost
	}

	return aggregated
}

// Close closes all providers
func (m *Manager) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var errors []error
	for name, provider := range m.providers {
		if err := provider.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close provider %s: %w", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing providers: %v", errors)
	}

	m.logger.Info("All AI providers closed successfully")
	return nil
}

// getProviderForModel finds the provider that supports a specific model
func (m *Manager) getProviderForModel(modelID string) (common.Provider, error) {
	providers := m.GetProviders()

	for _, provider := range providers {
		models := provider.GetModels()
		for _, model := range models {
			if model.ID == modelID {
				return provider, nil
			}
		}
	}

	return nil, fmt.Errorf("no provider found for model %s", modelID)
}

// getBestModelForProvider gets the best model of a specific type from a provider
func (m *Manager) getBestModelForProvider(provider common.Provider, modelType common.ModelType) string {
	models := provider.GetModels()

	for _, model := range models {
		if model.Type == modelType {
			return model.ID
		}
	}

	// Fallback to first available model
	if len(models) > 0 {
		return models[0].ID
	}

	return ""
}

// buildBeveragePrompt builds a prompt for beverage generation
func (m *Manager) buildBeveragePrompt(request *common.BeverageGenerationRequest) string {
	prompt := fmt.Sprintf("Create a %s themed beverage recipe", request.Theme)

	if len(request.Ingredients) > 0 {
		prompt += fmt.Sprintf(" using these ingredients: %v", request.Ingredients)
	}

	if len(request.Dietary) > 0 {
		prompt += fmt.Sprintf(" with dietary restrictions: %v", request.Dietary)
	}

	if request.Complexity != "" {
		prompt += fmt.Sprintf(" with %s complexity", request.Complexity)
	}

	if request.Temperature != "" {
		prompt += fmt.Sprintf(" served %s", request.Temperature)
	}

	prompt += ". Include name, description, ingredients with quantities, instructions, prep time, and difficulty level."

	return prompt
}

// Global AI manager instance
var globalManager *Manager

// InitGlobalManager initializes the global AI manager
func InitGlobalManager(
	config config.AIConfig,
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) error {
	globalManager = NewManager(config, logger, metrics, tracing)
	return globalManager.Initialize(context.Background())
}

// GetGlobalManager returns the global AI manager
func GetGlobalManager() *Manager {
	return globalManager
}

// CloseGlobalManager closes the global AI manager
func CloseGlobalManager() error {
	if globalManager == nil {
		return nil
	}
	return globalManager.Close()
}
