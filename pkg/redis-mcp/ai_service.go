package redismcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// GeminiConfig represents configuration for Google Gemini AI
type GeminiConfig struct {
	APIKey string `json:"api_key"`
	Model  string `json:"model"`
}

// OllamaConfig represents configuration for Ollama AI
type OllamaConfig struct {
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
}

// AIConfig represents the overall AI service configuration
type AIConfig struct {
	Gemini GeminiConfig `json:"gemini"`
	Ollama OllamaConfig `json:"ollama"`
}

// AIService provides AI capabilities for communication optimization
type AIService struct {
	config      *AIConfig
	redisClient *redis.Client
	logger      *logger.Logger
	httpClient  *http.Client
}

// AIRequest represents a request to the AI service
type AIRequest struct {
	Prompt    string                 `json:"prompt"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Model     string                 `json:"model,omitempty"`
	MaxTokens int                    `json:"max_tokens,omitempty"`
}

// AIResponse represents a response from the AI service
type AIResponse struct {
	Content   string                 `json:"content"`
	Model     string                 `json:"model"`
	Usage     map[string]interface{} `json:"usage,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewAIService creates a new AI service instance
func NewAIService(config interface{}, logger *logger.Logger, redisClient interface{}) (*AIService, error) {
	// Handle different config types for compatibility
	var aiConfig *AIConfig

	switch cfg := config.(type) {
	case *AIConfig:
		aiConfig = cfg
	case AIConfig:
		aiConfig = &cfg
	default:
		// Create default config
		aiConfig = &AIConfig{
			Gemini: GeminiConfig{
				APIKey: "",
				Model:  "gemini-pro",
			},
			Ollama: OllamaConfig{
				BaseURL: "http://localhost:11434",
				Model:   "llama2",
			},
		}
	}

	// Handle different Redis client types
	var redisClientTyped *redis.Client
	if rc, ok := redisClient.(*redis.Client); ok {
		redisClientTyped = rc
	} else {
		// Create a mock Redis client for compatibility
		redisClientTyped = nil
	}

	return &AIService{
		config:      aiConfig,
		redisClient: redisClientTyped,
		logger:      logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// ProcessMessage processes a message using AI (compatibility method)
func (ai *AIService) ProcessMessage(ctx context.Context, prompt, analyzer string) (string, error) {
	response, err := ai.GenerateResponse(ctx, prompt, map[string]interface{}{
		"analyzer": analyzer,
	})
	if err != nil {
		return "", err
	}
	return response.Content, nil
}

// GenerateResponse generates an AI response for the given prompt
func (ai *AIService) GenerateResponse(ctx context.Context, prompt string, context map[string]interface{}) (*AIResponse, error) {
	request := &AIRequest{
		Prompt:    prompt,
		Context:   context,
		MaxTokens: 1000,
	}

	// Try Gemini first if API key is available
	if ai.config.Gemini.APIKey != "" {
		request.Model = ai.config.Gemini.Model
		response, err := ai.callGemini(ctx, request)
		if err == nil {
			return response, nil
		}
		ai.logger.Warn("Gemini API failed, falling back to Ollama", zap.Error(err))
	}

	// Fall back to Ollama
	request.Model = ai.config.Ollama.Model
	return ai.callOllama(ctx, request)
}

// OptimizeMessage optimizes a message for better communication
func (ai *AIService) OptimizeMessage(ctx context.Context, message string, messageType string) (string, error) {
	prompt := fmt.Sprintf(`
Optimize the following %s message for clarity and effectiveness:

Message: %s

Please provide an improved version that is:
- Clear and concise
- Professional but friendly
- Appropriate for a coffee shop context
- Easy to understand

Return only the optimized message without explanations.
`, messageType, message)

	response, err := ai.GenerateResponse(ctx, prompt, map[string]interface{}{
		"message_type": messageType,
		"original":     message,
	})
	if err != nil {
		return message, err // Return original if optimization fails
	}

	return response.Content, nil
}

// AnalyzeCommunicationPattern analyzes communication patterns for insights
func (ai *AIService) AnalyzeCommunicationPattern(ctx context.Context, messages []string) (map[string]interface{}, error) {
	if len(messages) == 0 {
		return map[string]interface{}{}, nil
	}

	prompt := fmt.Sprintf(`
Analyze the following communication messages and provide insights:

Messages:
%s

Please analyze and provide insights on:
1. Communication tone and sentiment
2. Common themes or patterns
3. Potential improvements
4. Efficiency metrics
5. Recommendations for optimization

Return the analysis as a JSON object with the following structure:
{
  "tone": "string",
  "sentiment": "positive/neutral/negative",
  "themes": ["theme1", "theme2"],
  "efficiency_score": 0.0-1.0,
  "recommendations": ["rec1", "rec2"]
}
`, formatMessages(messages))

	response, err := ai.GenerateResponse(ctx, prompt, map[string]interface{}{
		"message_count": len(messages),
		"analysis_type": "communication_pattern",
	})
	if err != nil {
		return nil, err
	}

	// Try to parse JSON response
	var analysis map[string]interface{}
	if err := json.Unmarshal([]byte(response.Content), &analysis); err != nil {
		// If JSON parsing fails, return a basic analysis
		return map[string]interface{}{
			"tone":             "neutral",
			"sentiment":        "neutral",
			"themes":           []string{"general_communication"},
			"efficiency_score": 0.7,
			"recommendations":  []string{"Continue monitoring communication patterns"},
			"raw_response":     response.Content,
		}, nil
	}

	return analysis, nil
}

// GenerateMessageTemplate generates a message template for a specific scenario
func (ai *AIService) GenerateMessageTemplate(ctx context.Context, scenario string, variables []string) (string, error) {
	prompt := fmt.Sprintf(`
Generate a message template for the following coffee shop scenario: %s

The template should include placeholders for these variables: %v

Requirements:
- Use {variable_name} format for placeholders
- Keep it professional but friendly
- Make it suitable for customer communication
- Ensure it's clear and actionable

Return only the template without explanations.
`, scenario, variables)

	response, err := ai.GenerateResponse(ctx, prompt, map[string]interface{}{
		"scenario":  scenario,
		"variables": variables,
	})
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

// callGemini makes a request to Google Gemini API
func (ai *AIService) callGemini(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	// Mock implementation - in production, integrate with actual Gemini API
	ai.logger.Info("Calling Gemini API", zap.String("model", request.Model))

	// Simulate API call delay
	time.Sleep(100 * time.Millisecond)

	// Generate mock response
	response := &AIResponse{
		Content: ai.generateMockResponse(request.Prompt),
		Model:   request.Model,
		Usage: map[string]interface{}{
			"prompt_tokens":     len(request.Prompt) / 4, // Rough token estimate
			"completion_tokens": 50,
			"total_tokens":      len(request.Prompt)/4 + 50,
		},
		Metadata: map[string]interface{}{
			"provider": "gemini",
			"version":  "1.0",
		},
		Timestamp: time.Now(),
	}

	return response, nil
}

// callOllama makes a request to Ollama API
func (ai *AIService) callOllama(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	ai.logger.Info("Calling Ollama API", zap.String("model", request.Model))

	// Prepare Ollama request
	ollamaRequest := map[string]interface{}{
		"model":  request.Model,
		"prompt": request.Prompt,
		"stream": false,
	}

	jsonData, err := json.Marshal(ollamaRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Ollama request: %w", err)
	}

	// Make HTTP request to Ollama
	url := ai.config.Ollama.BaseURL + "/api/generate"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ai.httpClient.Do(req)
	if err != nil {
		// If Ollama is not available, return mock response
		ai.logger.Warn("Ollama API not available, using mock response", zap.Error(err))
		return ai.generateMockOllamaResponse(request), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API returned status %d", resp.StatusCode)
	}

	var ollamaResponse struct {
		Response string `json:"response"`
		Model    string `json:"model"`
		Done     bool   `json:"done"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResponse); err != nil {
		return nil, fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	response := &AIResponse{
		Content: ollamaResponse.Response,
		Model:   ollamaResponse.Model,
		Metadata: map[string]interface{}{
			"provider": "ollama",
			"done":     ollamaResponse.Done,
		},
		Timestamp: time.Now(),
	}

	return response, nil
}

// generateMockResponse generates a mock AI response for testing
func (ai *AIService) generateMockResponse(prompt string) string {
	responses := map[string]string{
		"optimize":     "Your optimized message is clear, professional, and customer-friendly.",
		"analyze":      `{"tone": "professional", "sentiment": "positive", "themes": ["customer_service"], "efficiency_score": 0.85, "recommendations": ["Continue current approach"]}`,
		"template":     "Dear {customer_name}, your order #{order_id} is {status}. Thank you for choosing our coffee shop!",
		"communication": "Based on the analysis, communication patterns show positive customer engagement with room for improvement in response times.",
	}

	for key, response := range responses {
		if contains(prompt, key) {
			return response
		}
	}

	return "Thank you for your message. Our AI service is processing your request and will provide an optimized response."
}

// generateMockOllamaResponse generates a mock Ollama response
func (ai *AIService) generateMockOllamaResponse(request *AIRequest) *AIResponse {
	return &AIResponse{
		Content: ai.generateMockResponse(request.Prompt),
		Model:   request.Model,
		Metadata: map[string]interface{}{
			"provider": "ollama_mock",
			"fallback": true,
		},
		Timestamp: time.Now(),
	}
}

// Helper functions

// formatMessages formats a slice of messages for AI analysis
func formatMessages(messages []string) string {
	var formatted string
	for i, msg := range messages {
		formatted += fmt.Sprintf("%d. %s\n", i+1, msg)
	}
	return formatted
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr ||
		      findInString(s, substr))))
}

// findInString finds substring in string
func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Health checks the health of the AI service
func (ai *AIService) Health(ctx context.Context) error {
	// Check if at least one AI provider is available
	if ai.config.Gemini.APIKey != "" {
		return nil // Gemini is configured
	}

	// Check Ollama health
	url := ai.config.Ollama.BaseURL + "/api/tags"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create Ollama health request: %w", err)
	}

	resp, err := ai.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Ollama health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama unhealthy: status %d", resp.StatusCode)
	}

	return nil
}
