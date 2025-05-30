package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/config"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
)

// OllamaClient represents an Ollama AI client
type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
	config     config.OllamaConfig
	logger     *logger.Logger
}

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model       string                 `json:"model"`
	Prompt      string                 `json:"prompt"`
	Stream      bool                   `json:"stream"`
	Temperature float64                `json:"temperature,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
	KeepAlive   string                 `json:"keep_alive,omitempty"`
}

// OllamaResponse represents a response from Ollama API
type OllamaResponse struct {
	Model     string    `json:"model"`
	Response  string    `json:"response"`
	Done      bool      `json:"done"`
	Context   []int     `json:"context,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// OllamaErrorResponse represents an error response from Ollama
type OllamaErrorResponse struct {
	Error string `json:"error"`
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(cfg config.OllamaConfig, logger *logger.Logger) (*OllamaClient, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("ollama client is disabled")
	}

	baseURL := fmt.Sprintf("http://%s:%d", cfg.Host, cfg.Port)
	
	client := &OllamaClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
		config: cfg,
		logger: logger,
	}

	// Test connection
	if err := client.testConnection(); err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama: %w", err)
	}

	logger.Info("Ollama client initialized successfully")
	return client, nil
}

// GenerateResponse generates a response using Ollama
func (o *OllamaClient) GenerateResponse(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	o.logger.Info(fmt.Sprintf("Generating Ollama response for user %s", req.UserID))

	temperature := o.config.Temperature
	if req.Temperature > 0 {
		temperature = req.Temperature
	}

	ollamaReq := OllamaRequest{
		Model:       o.config.Model,
		Prompt:      req.Message,
		Stream:      false,
		Temperature: temperature,
		KeepAlive:   o.config.KeepAlive,
	}

	// Add options if specified
	if req.MaxTokens > 0 {
		ollamaReq.Options = map[string]interface{}{
			"num_predict": req.MaxTokens,
		}
	}

	// Marshal request
	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		o.logger.Error(fmt.Sprintf("Ollama request failed: %v", err))
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		var errorResp OllamaErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("ollama error: %s", errorResp.Error)
		}
		return nil, fmt.Errorf("ollama request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if ollamaResp.Response == "" {
		return nil, fmt.Errorf("empty response from Ollama")
	}

	response := &GenerateResponse{
		Text:        ollamaResp.Response,
		Provider:    "ollama",
		Confidence:  0.8, // Default confidence for Ollama
		GeneratedAt: time.Now(),
		Metadata: map[string]string{
			"model":      ollamaResp.Model,
			"created_at": ollamaResp.CreatedAt.Format(time.RFC3339),
		},
	}

	o.logger.Info(fmt.Sprintf("Ollama response generated successfully for user %s", req.UserID))
	return response, nil
}

// GenerateStreamResponse generates a streaming response using Ollama
func (o *OllamaClient) GenerateStreamResponse(ctx context.Context, req *GenerateRequest, callback func(string) error) error {
	o.logger.Info(fmt.Sprintf("Generating Ollama streaming response for user %s", req.UserID))

	temperature := o.config.Temperature
	if req.Temperature > 0 {
		temperature = req.Temperature
	}

	ollamaReq := OllamaRequest{
		Model:       o.config.Model,
		Prompt:      req.Message,
		Stream:      true,
		Temperature: temperature,
		KeepAlive:   o.config.KeepAlive,
	}

	// Add options if specified
	if req.MaxTokens > 0 {
		ollamaReq.Options = map[string]interface{}{
			"num_predict": req.MaxTokens,
		}
	}

	// Marshal request
	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		o.logger.Error(fmt.Sprintf("Ollama streaming request failed: %v", err))
		return fmt.Errorf("ollama streaming request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var errorResp OllamaErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return fmt.Errorf("ollama error: %s", errorResp.Error)
		}
		return fmt.Errorf("ollama request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read streaming response
	decoder := json.NewDecoder(resp.Body)
	for {
		var ollamaResp OllamaResponse
		if err := decoder.Decode(&ollamaResp); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode streaming response: %w", err)
		}

		// Send chunk to callback
		if ollamaResp.Response != "" {
			if err := callback(ollamaResp.Response); err != nil {
				return fmt.Errorf("callback error: %w", err)
			}
		}

		// Check if done
		if ollamaResp.Done {
			break
		}
	}

	o.logger.Info(fmt.Sprintf("Ollama streaming response completed for user %s", req.UserID))
	return nil
}

// IsHealthy checks if the Ollama client is healthy
func (o *OllamaClient) IsHealthy(ctx context.Context) bool {
	return o.testConnection() == nil
}

// testConnection tests the connection to Ollama
func (o *OllamaClient) testConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to get model list
	req, err := http.NewRequestWithContext(ctx, "GET", o.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// Close closes the Ollama client
func (o *OllamaClient) Close() error {
	// HTTP client doesn't need explicit closing
	return nil
}

// ListModels lists available models in Ollama
func (o *OllamaClient) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", o.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get models with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var modelsResp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, fmt.Errorf("failed to parse models response: %w", err)
	}

	var models []string
	for _, model := range modelsResp.Models {
		models = append(models, model.Name)
	}

	return models, nil
}

// PullModel pulls a model in Ollama
func (o *OllamaClient) PullModel(ctx context.Context, modelName string) error {
	pullReq := map[string]string{
		"name": modelName,
	}

	reqBody, err := json.Marshal(pullReq)
	if err != nil {
		return fmt.Errorf("failed to marshal pull request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/pull", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to pull model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to pull model with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
