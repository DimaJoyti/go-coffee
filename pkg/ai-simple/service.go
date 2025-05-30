package aisimple

import (
	"context"
	"fmt"
	"strings"
)

// Service interface for simple AI operations
type Service interface {
	GenerateText(ctx context.Context, prompt string) (string, error)
	ProcessMessage(ctx context.Context, message string, userID string) (string, error)
	GetCoffeeRecommendation(ctx context.Context, preferences map[string]interface{}) (string, error)
	AnalyzeSpending(ctx context.Context, userID string) (string, error)
	GetMarketInsights(ctx context.Context) (string, error)
	Close() error
}

// SimpleService provides basic AI operations using rule-based logic
type SimpleService struct {
	name string
}

// NewService creates a new simple AI service
func NewService(name string) *SimpleService {
	return &SimpleService{
		name: name,
	}
}

// GenerateText generates text based on a prompt using simple rules
func (s *SimpleService) GenerateText(ctx context.Context, prompt string) (string, error) {
	// Simple rule-based text generation for Redis queries
	prompt = strings.ToLower(prompt)
	
	if strings.Contains(prompt, "coffee") && strings.Contains(prompt, "menu") {
		return `{
			"type": "read",
			"operation": "HGETALL",
			"key": "coffee:menu:downtown",
			"confidence": 0.8
		}`, nil
	}
	
	if strings.Contains(prompt, "inventory") {
		return `{
			"type": "read",
			"operation": "HGETALL",
			"key": "coffee:inventory:downtown",
			"confidence": 0.8
		}`, nil
	}
	
	if strings.Contains(prompt, "orders") {
		return `{
			"type": "read",
			"operation": "ZREVRANGE",
			"key": "coffee:orders:today",
			"confidence": 0.8
		}`, nil
	}
	
	if strings.Contains(prompt, "customer") {
		return `{
			"type": "read",
			"operation": "HGETALL",
			"key": "customer:123",
			"confidence": 0.8
		}`, nil
	}
	
	// Default response
	return `{
		"type": "read",
		"operation": "SCAN",
		"key": "*",
		"confidence": 0.5
	}`, nil
}

// ProcessMessage processes a message using simple rules
func (s *SimpleService) ProcessMessage(ctx context.Context, message string, userID string) (string, error) {
	message = strings.ToLower(message)
	
	if strings.Contains(message, "coffee") {
		return "I can help you with coffee-related queries. Try asking about menu, inventory, or orders.", nil
	}
	
	if strings.Contains(message, "menu") {
		return "Here are our coffee menu options. Would you like to see prices for a specific location?", nil
	}
	
	if strings.Contains(message, "order") {
		return "I can help you check order statistics and popular drinks.", nil
	}
	
	return "I'm a simple AI assistant for coffee shop operations. Ask me about menu, inventory, orders, or customers.", nil
}

// GetCoffeeRecommendation provides coffee recommendations
func (s *SimpleService) GetCoffeeRecommendation(ctx context.Context, preferences map[string]interface{}) (string, error) {
	// Simple recommendation logic
	if strength, ok := preferences["strength"]; ok && strength == "strong" {
		return "I recommend our Espresso or Americano for a strong coffee experience.", nil
	}
	
	if milk, ok := preferences["milk"]; ok && milk == "oat" {
		return "Try our Oat Milk Latte - it's creamy and delicious!", nil
	}
	
	return "Based on your preferences, I recommend our signature Latte - it's our most popular drink!", nil
}

// AnalyzeSpending analyzes user spending patterns
func (s *SimpleService) AnalyzeSpending(ctx context.Context, userID string) (string, error) {
	return fmt.Sprintf("User %s has moderate spending patterns. Average order value is $4.50.", userID), nil
}

// GetMarketInsights provides market insights
func (s *SimpleService) GetMarketInsights(ctx context.Context) (string, error) {
	return "Coffee market trends show increased demand for oat milk alternatives and specialty drinks.", nil
}

// Close closes the AI service
func (s *SimpleService) Close() error {
	return nil
}
