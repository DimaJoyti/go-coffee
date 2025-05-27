package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/config"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
)

// GeminiClient represents a Google Gemini AI client
type GeminiClient struct {
	client  *genai.Client
	model   *genai.GenerativeModel
	config  config.GeminiConfig
	logger  *logger.Logger
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(cfg config.GeminiConfig, logger *logger.Logger) (*GeminiClient, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("gemini client is disabled")
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("gemini API key is required")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Configure the model
	model := client.GenerativeModel(cfg.Model)
	
	// Set generation config
	model.GenerationConfig = &genai.GenerationConfig{
		Temperature:     &cfg.Temperature,
		MaxOutputTokens: int32Ptr(int32(cfg.MaxTokens)),
	}

	// Set safety settings
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: parseHarmBlockThreshold(cfg.SafetySettings.Harassment),
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: parseHarmBlockThreshold(cfg.SafetySettings.HateSpeech),
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: parseHarmBlockThreshold(cfg.SafetySettings.SexuallyExplicit),
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: parseHarmBlockThreshold(cfg.SafetySettings.DangerousContent),
		},
	}

	geminiClient := &GeminiClient{
		client: client,
		model:  model,
		config: cfg,
		logger: logger,
	}

	logger.Info("Gemini client initialized successfully")
	return geminiClient, nil
}

// GenerateResponse generates a response using Gemini
func (g *GeminiClient) GenerateResponse(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	g.logger.Info(fmt.Sprintf("Generating Gemini response for user %s", req.UserID))

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(g.config.Timeout)*time.Second)
	defer cancel()

	// Override model settings if specified in request
	model := g.model
	if req.Temperature > 0 {
		tempModel := g.client.GenerativeModel(g.config.Model)
		tempModel.GenerationConfig = &genai.GenerationConfig{
			Temperature: &req.Temperature,
		}
		if req.MaxTokens > 0 {
			tempModel.GenerationConfig.MaxOutputTokens = int32Ptr(int32(req.MaxTokens))
		}
		model = tempModel
	}

	// Generate content
	resp, err := model.GenerateContent(timeoutCtx, genai.Text(req.Message))
	if err != nil {
		g.logger.Error(fmt.Sprintf("Gemini generation failed: %v", err))
		return nil, fmt.Errorf("gemini generation failed: %w", err)
	}

	if resp == nil || len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no response candidates from Gemini")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	// Extract text from the first part
	var responseText string
	for _, part := range candidate.Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			responseText += string(textPart)
		}
	}

	if responseText == "" {
		return nil, fmt.Errorf("no text content in Gemini response")
	}

	// Calculate confidence based on safety ratings and finish reason
	confidence := g.calculateConfidence(candidate)

	response := &GenerateResponse{
		Text:        responseText,
		Provider:    "gemini",
		Confidence:  confidence,
		GeneratedAt: time.Now(),
		Metadata: map[string]string{
			"model":         g.config.Model,
			"finish_reason": candidate.FinishReason.String(),
		},
	}

	// Add safety ratings to metadata
	if len(candidate.SafetyRatings) > 0 {
		for _, rating := range candidate.SafetyRatings {
			key := fmt.Sprintf("safety_%s", rating.Category.String())
			response.Metadata[key] = rating.Probability.String()
		}
	}

	g.logger.Info(fmt.Sprintf("Gemini response generated successfully for user %s", req.UserID))
	return response, nil
}

// IsHealthy checks if the Gemini client is healthy
func (g *GeminiClient) IsHealthy(ctx context.Context) bool {
	// Simple health check by generating a minimal response
	testReq := &GenerateRequest{
		UserID:  "health_check",
		Message: "Hello",
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := g.GenerateResponse(timeoutCtx, testReq)
	return err == nil
}

// Close closes the Gemini client
func (g *GeminiClient) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}

// calculateConfidence calculates confidence score based on response quality
func (g *GeminiClient) calculateConfidence(candidate *genai.Candidate) float64 {
	confidence := 1.0

	// Reduce confidence based on finish reason
	switch candidate.FinishReason {
	case genai.FinishReasonStop:
		// Normal completion, no reduction
	case genai.FinishReasonMaxTokens:
		confidence *= 0.8 // Slightly reduce for truncated responses
	case genai.FinishReasonSafety:
		confidence *= 0.3 // Significantly reduce for safety-filtered responses
	case genai.FinishReasonRecitation:
		confidence *= 0.5 // Reduce for recitation issues
	default:
		confidence *= 0.6 // Reduce for other issues
	}

	// Reduce confidence based on safety ratings
	for _, rating := range candidate.SafetyRatings {
		switch rating.Probability {
		case genai.HarmProbabilityHigh:
			confidence *= 0.2
		case genai.HarmProbabilityMedium:
			confidence *= 0.6
		case genai.HarmProbabilityLow:
			confidence *= 0.9
		case genai.HarmProbabilityNegligible:
			// No reduction
		}
	}

	// Ensure confidence is between 0 and 1
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}

// parseHarmBlockThreshold parses harm block threshold from string
func parseHarmBlockThreshold(threshold string) genai.HarmBlockThreshold {
	switch threshold {
	case "BLOCK_NONE":
		return genai.HarmBlockThresholdBlockNone
	case "BLOCK_LOW_AND_ABOVE":
		return genai.HarmBlockThresholdBlockLowAndAbove
	case "BLOCK_MEDIUM_AND_ABOVE":
		return genai.HarmBlockThresholdBlockMediumAndAbove
	case "BLOCK_ONLY_HIGH":
		return genai.HarmBlockThresholdBlockOnlyHigh
	default:
		return genai.HarmBlockThresholdBlockMediumAndAbove
	}
}

// int32Ptr returns a pointer to an int32 value
func int32Ptr(v int32) *int32 {
	return &v
}

// GenerateStreamResponse generates a streaming response using Gemini
func (g *GeminiClient) GenerateStreamResponse(ctx context.Context, req *GenerateRequest, callback func(string) error) error {
	g.logger.Info(fmt.Sprintf("Generating Gemini streaming response for user %s", req.UserID))

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(g.config.Timeout)*time.Second)
	defer cancel()

	// Generate streaming content
	iter := g.model.GenerateContentStream(timeoutCtx, genai.Text(req.Message))
	defer iter.Stop()

	for {
		resp, err := iter.Next()
		if err != nil {
			if err.Error() == "iterator stopped" {
				break
			}
			g.logger.Error(fmt.Sprintf("Gemini streaming failed: %v", err))
			return fmt.Errorf("gemini streaming failed: %w", err)
		}

		if resp == nil || len(resp.Candidates) == 0 {
			continue
		}

		candidate := resp.Candidates[0]
		if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
			continue
		}

		// Extract text from parts
		for _, part := range candidate.Content.Parts {
			if textPart, ok := part.(genai.Text); ok {
				if err := callback(string(textPart)); err != nil {
					return fmt.Errorf("callback error: %w", err)
				}
			}
		}
	}

	g.logger.Info(fmt.Sprintf("Gemini streaming response completed for user %s", req.UserID))
	return nil
}
