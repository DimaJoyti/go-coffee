package handlers

import (
	"net/http"
	"time"

	"go-coffee-ai-agents/internal/ai"
	"go-coffee-ai-agents/internal/common"
	"go-coffee-ai-agents/internal/httputils"
	"go-coffee-ai-agents/internal/observability"
)

// AIHandler handles AI-related HTTP requests
type AIHandler struct {
	logger  *observability.StructuredLogger
	metrics *observability.MetricsCollector
	tracing *observability.TracingHelper
}

// NewAIHandler creates a new AI handler
func NewAIHandler(
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) *AIHandler {
	return &AIHandler{
		logger:  logger,
		metrics: metrics,
		tracing: tracing,
	}
}

// TextGenerationRequest represents a text generation request
type TextGenerationRequest struct {
	Model       string            `json:"model" validate:"required"`
	Prompt      string            `json:"prompt" validate:"required,min=1,max=10000"`
	MaxTokens   int               `json:"max_tokens" validate:"min=1,max=4000"`
	Temperature float64           `json:"temperature" validate:"min=0,max=2"`
	TopP        float64           `json:"top_p" validate:"min=0,max=1"`
	TopK        int               `json:"top_k" validate:"min=0"`
	Stop        []string          `json:"stop"`
	Metadata    map[string]string `json:"metadata"`
}

// ChatGenerationRequest represents a chat generation request
type ChatGenerationRequest struct {
	Model       string            `json:"model" validate:"required"`
	Messages    []ChatMessage     `json:"messages" validate:"required,min=1"`
	MaxTokens   int               `json:"max_tokens" validate:"min=1,max=4000"`
	Temperature float64           `json:"temperature" validate:"min=0,max=2"`
	TopP        float64           `json:"top_p" validate:"min=0,max=1"`
	TopK        int               `json:"top_k" validate:"min=0"`
	Stop        []string          `json:"stop"`
	Stream      bool              `json:"stream"`
	Metadata    map[string]string `json:"metadata"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role" validate:"required,oneof=system user assistant"`
	Content string `json:"content" validate:"required,min=1"`
	Name    string `json:"name"`
}

// EmbeddingRequest represents an embedding generation request
type EmbeddingRequest struct {
	Model    string            `json:"model" validate:"required"`
	Input    []string          `json:"input" validate:"required,min=1,max=100"`
	Metadata map[string]string `json:"metadata"`
}

// AIResponse represents a generic AI response
type AIResponse struct {
	ID           string                 `json:"id"`
	Model        string                 `json:"model"`
	Provider     string                 `json:"provider"`
	Usage        *TokenUsage            `json:"usage"`
	FinishReason string                 `json:"finish_reason,omitempty"`
	Metadata     map[string]string      `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	Data         interface{}            `json:"data"`
}

// TokenUsage represents token usage information
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// GenerateText handles POST /api/v1/ai/text
func (h *AIHandler) GenerateText(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "POST", "/api/v1/ai/text", r.UserAgent())
	defer span.End()

	h.logger.InfoContext(ctx, "Generating text with AI")

	// Decode request body
	var req TextGenerationRequest
	if err := httputils.DecodeJSONBody(r, &req); err != nil {
		h.tracing.RecordError(span, err, "Failed to decode request body")
		httputils.WriteErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Set defaults
	if req.MaxTokens == 0 {
		req.MaxTokens = 100
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	h.logger.InfoContext(ctx, "AI text generation request",
		"model", req.Model,
		"prompt_length", len(req.Prompt),
		"max_tokens", req.MaxTokens,
		"temperature", req.Temperature)

	// Get AI manager
	aiManager := ai.GetGlobalManager()
	if aiManager == nil {
		h.tracing.RecordError(span, nil, "AI manager not available")
		httputils.WriteErrorResponse(w, http.StatusServiceUnavailable, "ai_unavailable", "AI service is not available")
		return
	}

	// Create AI request
	aiRequest := &common.TextGenerationRequest{
		Model:       req.Model,
		Prompt:      req.Prompt,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		TopK:        req.TopK,
		Stop:        req.Stop,
		Metadata:    req.Metadata,
	}

	// Generate text
	aiResponse, err := aiManager.GenerateText(ctx, aiRequest)
	if err != nil {
		h.tracing.RecordError(span, err, "AI text generation failed")
		h.logger.ErrorContext(ctx, "AI text generation failed", err)
		httputils.WriteErrorResponse(w, http.StatusInternalServerError, "generation_failed", "Failed to generate text")
		return
	}

	// Create response
	response := AIResponse{
		ID:           aiResponse.ID,
		Model:        aiResponse.Model,
		Provider:     aiResponse.Provider,
		Usage:        convertTokenUsage(aiResponse.Usage),
		FinishReason: aiResponse.FinishReason,
		Metadata:     aiResponse.Metadata,
		CreatedAt:    aiResponse.CreatedAt,
		Data: map[string]interface{}{
			"text": aiResponse.Text,
		},
	}

	h.tracing.RecordSuccess(span, "Text generated successfully")
	h.logger.InfoContext(ctx, "Text generated successfully",
		"model", aiResponse.Model,
		"provider", aiResponse.Provider,
		"tokens_used", aiResponse.Usage.TotalTokens,
		"text_length", len(aiResponse.Text))

	httputils.WriteJSONResponse(w, http.StatusOK, response)
}

// GenerateChat handles POST /api/v1/ai/chat
func (h *AIHandler) GenerateChat(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "POST", "/api/v1/ai/chat", r.UserAgent())
	defer span.End()

	h.logger.InfoContext(ctx, "Generating chat completion with AI")

	// Decode request body
	var req ChatGenerationRequest
	if err := httputils.DecodeJSONBody(r, &req); err != nil {
		h.tracing.RecordError(span, err, "Failed to decode request body")
		httputils.WriteErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Set defaults
	if req.MaxTokens == 0 {
		req.MaxTokens = 150
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	h.logger.InfoContext(ctx, "AI chat generation request",
		"model", req.Model,
		"messages_count", len(req.Messages),
		"max_tokens", req.MaxTokens,
		"temperature", req.Temperature)

	// Get AI manager
	aiManager := ai.GetGlobalManager()
	if aiManager == nil {
		h.tracing.RecordError(span, nil, "AI manager not available")
		httputils.WriteErrorResponse(w, http.StatusServiceUnavailable, "ai_unavailable", "AI service is not available")
		return
	}

	// Convert messages
	aiMessages := make([]common.ChatMessage, len(req.Messages))
	for i, msg := range req.Messages {
		aiMessages[i] = common.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}
	}

	// Create AI request
	aiRequest := &common.ChatRequest{
		Model:       req.Model,
		Messages:    aiMessages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		TopK:        req.TopK,
		Stop:        req.Stop,
		Stream:      req.Stream,
		Metadata:    req.Metadata,
	}

	// Generate chat completion
	aiResponse, err := aiManager.GenerateChat(ctx, aiRequest)
	if err != nil {
		h.tracing.RecordError(span, err, "AI chat generation failed")
		h.logger.ErrorContext(ctx, "AI chat generation failed", err)
		httputils.WriteErrorResponse(w, http.StatusInternalServerError, "generation_failed", "Failed to generate chat completion")
		return
	}

	// Create response
	response := AIResponse{
		ID:           aiResponse.ID,
		Model:        aiResponse.Model,
		Provider:     aiResponse.Provider,
		Usage:        convertTokenUsage(aiResponse.Usage),
		FinishReason: aiResponse.FinishReason,
		Metadata:     aiResponse.Metadata,
		CreatedAt:    aiResponse.CreatedAt,
		Data: map[string]interface{}{
			"message": ChatMessage{
				Role:    aiResponse.Message.Role,
				Content: aiResponse.Message.Content,
				Name:    aiResponse.Message.Name,
			},
		},
	}

	h.tracing.RecordSuccess(span, "Chat completion generated successfully")
	h.logger.InfoContext(ctx, "Chat completion generated successfully",
		"model", aiResponse.Model,
		"provider", aiResponse.Provider,
		"tokens_used", aiResponse.Usage.TotalTokens,
		"response_length", len(aiResponse.Message.Content))

	httputils.WriteJSONResponse(w, http.StatusOK, response)
}

// GenerateEmbedding handles POST /api/v1/ai/embedding
func (h *AIHandler) GenerateEmbedding(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "POST", "/api/v1/ai/embedding", r.UserAgent())
	defer span.End()

	h.logger.InfoContext(ctx, "Generating embeddings with AI")

	// Decode request body
	var req EmbeddingRequest
	if err := httputils.DecodeJSONBody(r, &req); err != nil {
		h.tracing.RecordError(span, err, "Failed to decode request body")
		httputils.WriteErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	h.logger.InfoContext(ctx, "AI embedding generation request",
		"model", req.Model,
		"input_count", len(req.Input))

	// Get AI manager
	aiManager := ai.GetGlobalManager()
	if aiManager == nil {
		h.tracing.RecordError(span, nil, "AI manager not available")
		httputils.WriteErrorResponse(w, http.StatusServiceUnavailable, "ai_unavailable", "AI service is not available")
		return
	}

	// Create AI request
	aiRequest := &common.EmbeddingRequest{
		Model:    req.Model,
		Input:    req.Input,
		Metadata: req.Metadata,
	}

	// Generate embeddings
	aiResponse, err := aiManager.GenerateEmbedding(ctx, aiRequest)
	if err != nil {
		h.tracing.RecordError(span, err, "AI embedding generation failed")
		h.logger.ErrorContext(ctx, "AI embedding generation failed", err)
		httputils.WriteErrorResponse(w, http.StatusInternalServerError, "generation_failed", "Failed to generate embeddings")
		return
	}

	// Create response
	response := AIResponse{
		ID:        aiResponse.ID,
		Model:     aiResponse.Model,
		Provider:  aiResponse.Provider,
		Usage:     convertTokenUsage(aiResponse.Usage),
		Metadata:  aiResponse.Metadata,
		CreatedAt: aiResponse.CreatedAt,
		Data: map[string]interface{}{
			"embeddings": aiResponse.Embeddings,
		},
	}

	h.tracing.RecordSuccess(span, "Embeddings generated successfully")
	h.logger.InfoContext(ctx, "Embeddings generated successfully",
		"model", aiResponse.Model,
		"provider", aiResponse.Provider,
		"tokens_used", aiResponse.Usage.TotalTokens,
		"embeddings_count", len(aiResponse.Embeddings))

	httputils.WriteJSONResponse(w, http.StatusOK, response)
}

// ListProviders handles GET /api/v1/ai/providers
func (h *AIHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "GET", "/api/v1/ai/providers", r.UserAgent())
	defer span.End()

	h.logger.InfoContext(ctx, "Listing AI providers")

	// Get AI manager
	aiManager := ai.GetGlobalManager()
	if aiManager == nil {
		h.tracing.RecordError(span, nil, "AI manager not available")
		httputils.WriteErrorResponse(w, http.StatusServiceUnavailable, "ai_unavailable", "AI service is not available")
		return
	}

	// Get providers
	providers := aiManager.GetProviders()
	
	// Convert to response format
	providerList := make([]map[string]interface{}, len(providers))
	for i, provider := range providers {
		usage := provider.GetUsage()
		providerList[i] = map[string]interface{}{
			"name":             provider.GetName(),
			"models_count":     len(provider.GetModels()),
			"total_requests":   usage.TotalRequests,
			"successful_requests": usage.SuccessfulReqs,
			"failed_requests":  usage.FailedRequests,
			"total_tokens":     usage.TotalTokens,
			"total_cost":       usage.TotalCost,
			"average_latency":  usage.AverageLatency.Milliseconds(),
			"last_request":     usage.LastRequestTime,
		}
	}

	h.tracing.RecordSuccess(span, "AI providers listed successfully")
	h.logger.InfoContext(ctx, "AI providers listed successfully", "providers_count", len(providers))

	httputils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"providers": providerList,
		"total":     len(providerList),
	})
}

// ListModels handles GET /api/v1/ai/models
func (h *AIHandler) ListModels(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "GET", "/api/v1/ai/models", r.UserAgent())
	defer span.End()

	h.logger.InfoContext(ctx, "Listing AI models")

	// Get AI manager
	aiManager := ai.GetGlobalManager()
	if aiManager == nil {
		h.tracing.RecordError(span, nil, "AI manager not available")
		httputils.WriteErrorResponse(w, http.StatusServiceUnavailable, "ai_unavailable", "AI service is not available")
		return
	}

	// Get all models from all providers
	providers := aiManager.GetProviders()
	var allModels []common.Model
	
	for _, provider := range providers {
		models := provider.GetModels()
		allModels = append(allModels, models...)
	}

	h.tracing.RecordSuccess(span, "AI models listed successfully")
	h.logger.InfoContext(ctx, "AI models listed successfully", "models_count", len(allModels))

	httputils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"models": allModels,
		"total":  len(allModels),
	})
}

// GetUsage handles GET /api/v1/ai/usage
func (h *AIHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "GET", "/api/v1/ai/usage", r.UserAgent())
	defer span.End()

	h.logger.InfoContext(ctx, "Getting AI usage statistics")

	// Get AI manager
	aiManager := ai.GetGlobalManager()
	if aiManager == nil {
		h.tracing.RecordError(span, nil, "AI manager not available")
		httputils.WriteErrorResponse(w, http.StatusServiceUnavailable, "ai_unavailable", "AI service is not available")
		return
	}

	// Get aggregated usage
	usage := aiManager.GetAggregatedUsage()

	h.tracing.RecordSuccess(span, "AI usage statistics retrieved successfully")
	h.logger.InfoContext(ctx, "AI usage statistics retrieved successfully",
		"total_requests", usage.TotalRequests,
		"total_cost", usage.TotalCost)

	httputils.WriteJSONResponse(w, http.StatusOK, usage)
}

// HealthCheck handles GET /api/v1/ai/health
func (h *AIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "GET", "/api/v1/ai/health", r.UserAgent())
	defer span.End()

	h.logger.InfoContext(ctx, "Checking AI providers health")

	// Get AI manager
	aiManager := ai.GetGlobalManager()
	if aiManager == nil {
		h.tracing.RecordError(span, nil, "AI manager not available")
		httputils.WriteErrorResponse(w, http.StatusServiceUnavailable, "ai_unavailable", "AI service is not available")
		return
	}

	// Perform health check
	healthResults := aiManager.HealthCheck(ctx)
	
	// Calculate overall health
	healthyCount := 0
	totalCount := len(healthResults)
	
	for _, err := range healthResults {
		if err == nil {
			healthyCount++
		}
	}
	
	overallHealthy := healthyCount > 0
	statusCode := http.StatusOK
	if !overallHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	response := map[string]interface{}{
		"healthy":         overallHealthy,
		"providers_total": totalCount,
		"providers_healthy": healthyCount,
		"providers":       healthResults,
		"timestamp":       time.Now(),
	}

	if overallHealthy {
		h.tracing.RecordSuccess(span, "AI health check passed")
	} else {
		h.tracing.RecordError(span, nil, "AI health check failed")
	}

	h.logger.InfoContext(ctx, "AI health check completed",
		"healthy", overallHealthy,
		"providers_healthy", healthyCount,
		"providers_total", totalCount)

	httputils.WriteJSONResponse(w, statusCode, response)
}

// Helper functions

func convertTokenUsage(usage *common.TokenUsage) *TokenUsage {
	if usage == nil {
		return nil
	}
	
	return &TokenUsage{
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
	}
}
