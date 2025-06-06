package ai

import (
	"context"
	"fmt"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/config"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/redis"
)

// Service interface for AI operations
type Service interface {
	ProcessMessage(ctx context.Context, message string, userID string) (string, error)
	GetCoffeeRecommendation(ctx context.Context, preferences map[string]interface{}) (string, error)
	AnalyzeSpending(ctx context.Context, userID string) (string, error)
	GetMarketInsights(ctx context.Context) (string, error)
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
func NewService(cfg config.AIConfig, logger *logger.Logger, redisClient redis.Client) (*SimpleService, error) {
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

// Close closes the AI service
func (s *SimpleService) Close() error {
	if s.simpleAI != nil {
		return s.simpleAI.Close()
	}
	return nil
}
