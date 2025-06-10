package ai

import (
	"context"
	"fmt"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/redis"
)

// Service interface for AI operations
type Service interface {
	ProcessMessage(ctx context.Context, message string, userID string) (string, error)
	GetCoffeeRecommendation(ctx context.Context, preferences map[string]interface{}) (string, error)
	AnalyzeSpending(ctx context.Context, userID string) (string, error)
	GetMarketInsights(ctx context.Context) (string, error)
	GenerateResponse(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
	Close() error
}

// SimpleService provides simple AI operations
type SimpleService struct {
	config   config.AIConfig
	logger   *logger.Logger
	cache    redis.Client
	simpleAI *SimpleAIService
}

// NewService creates a new simple AI service
func NewService(cfg config.AIConfig, logger *logger.Logger, redisClient redis.Client) (Service, error) {
	// Create simple AI service
	simpleAI, err := NewSimpleAIService(cfg, logger, redisClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create simple AI service: %w", err)
	}

	return &SimpleService{
		config:   cfg,
		logger:   logger,
		cache:    redisClient,
		simpleAI: simpleAI,
	}, nil
}

// ProcessMessage processes a message using AI
func (s *SimpleService) ProcessMessage(ctx context.Context, message string, userID string) (string, error) {
	req := &SimpleAIRequest{
		Message: message,
		Context: map[string]interface{}{
			"user_id": userID,
		},
	}

	response, err := s.simpleAI.ProcessMessage(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to process message: %w", err)
	}

	return response.Response, nil
}

// GetCoffeeRecommendation gets coffee recommendation
func (s *SimpleService) GetCoffeeRecommendation(ctx context.Context, preferences map[string]interface{}) (string, error) {
	response, err := s.simpleAI.GetCoffeeRecommendation(ctx, preferences)
	if err != nil {
		return "", fmt.Errorf("failed to get recommendation: %w", err)
	}

	return response.Response, nil
}

// AnalyzeSpending analyzes user spending
func (s *SimpleService) AnalyzeSpending(ctx context.Context, userID string) (string, error) {
	response, err := s.simpleAI.AnalyzeSpending(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to analyze spending: %w", err)
	}

	return response.Response, nil
}

// GetMarketInsights gets market insights
func (s *SimpleService) GetMarketInsights(ctx context.Context) (string, error) {
	response, err := s.simpleAI.GetMarketInsights(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get market insights: %w", err)
	}

	return response.Response, nil
}

// GenerateResponse generates an AI response for content analysis
func (s *SimpleService) GenerateResponse(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	// Convert GenerateRequest to SimpleAIRequest
	simpleReq := &SimpleAIRequest{
		Message:     req.Message,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Context: map[string]interface{}{
			"user_id": req.UserID,
			"context": req.Context,
		},
	}

	// Add metadata to context
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			simpleReq.Context[k] = v
		}
	}

	// Process the message using simple AI
	response, err := s.simpleAI.ProcessMessage(ctx, simpleReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Convert SimpleAIResponse to GenerateResponse
	return &GenerateResponse{
		Text:        response.Response,
		Provider:    "simple_ai",
		Confidence:  0.85, // Default confidence for simple AI
		Metadata:    convertMetadata(response.Metadata),
		GeneratedAt: response.Timestamp,
	}, nil
}

// Close closes the AI service
func (s *SimpleService) Close() error {
	if s.simpleAI != nil {
		return s.simpleAI.Close()
	}
	return nil
}

// convertMetadata converts map[string]interface{} to map[string]string
func convertMetadata(metadata map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range metadata {
		if str, ok := v.(string); ok {
			result[k] = str
		} else {
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result
}
