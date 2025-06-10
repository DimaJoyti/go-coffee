package ai

import (
	"context"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/redis"
	"go.uber.org/zap"
)

// SimpleAIService represents a simple AI service for testing
type SimpleAIService struct {
	config      config.AIConfig
	logger      *logger.Logger
	redisClient redis.Client
}

// SimpleAIRequest represents a simple AI request
type SimpleAIRequest struct {
	Message     string                 `json:"message"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
}

// SimpleAIResponse represents a simple AI response
type SimpleAIResponse struct {
	Response  string                 `json:"response"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewSimpleAIService creates a new simple AI service
func NewSimpleAIService(cfg config.AIConfig, logger *logger.Logger, redisClient redis.Client) (*SimpleAIService, error) {
	return &SimpleAIService{
		config:      cfg,
		logger:      logger,
		redisClient: redisClient,
	}, nil
}

// ProcessMessage processes a message using simple pattern matching
func (s *SimpleAIService) ProcessMessage(ctx context.Context, req *SimpleAIRequest) (*SimpleAIResponse, error) {
	s.logger.Debug("Processing AI message", zap.String("message", req.Message))

	message := strings.ToLower(req.Message)
	var response string

	// Simple pattern matching for coffee-related queries
	switch {
	case strings.Contains(message, "hello") || strings.Contains(message, "hi"):
		response = "Hello! I'm your Web3 Coffee AI assistant. How can I help you today? ☕"

	case strings.Contains(message, "coffee") && strings.Contains(message, "order"):
		response = "I'd be happy to help you order coffee! Here are our popular options:\n" +
			"☕ Latte - $4.50\n" +
			"☕ Cappuccino - $4.00\n" +
			"☕ Americano - $3.00\n" +
			"☕ Espresso - $2.50\n\n" +
			"Which one would you like to order?"

	case strings.Contains(message, "menu"):
		response = "Here's our coffee menu:\n\n" +
			"🌟 **Signature Drinks:**\n" +
			"☕ Latte - $4.50\n" +
			"☕ Cappuccino - $4.00\n" +
			"☕ Macchiato - $4.25\n\n" +
			"🔥 **Espresso Based:**\n" +
			"☕ Americano - $3.00\n" +
			"☕ Espresso - $2.50\n\n" +
			"💰 You can pay with crypto: BTC, ETH, USDC, or our COFFEE tokens!"

	case strings.Contains(message, "wallet") || strings.Contains(message, "balance"):
		response = "I can help you with your Web3 wallet! Here's what I can do:\n\n" +
			"💰 Check your crypto balance\n" +
			"📤 Send transactions\n" +
			"📥 Receive payments\n" +
			"🔄 Swap tokens\n" +
			"☕ Pay for coffee with crypto\n\n" +
			"What would you like to do with your wallet?"

	case strings.Contains(message, "crypto") || strings.Contains(message, "bitcoin") || strings.Contains(message, "ethereum"):
		response = "Great! We support multiple cryptocurrencies:\n\n" +
			"🟡 Bitcoin (BTC)\n" +
			"🔵 Ethereum (ETH)\n" +
			"💚 USDC Stablecoin\n" +
			"☕ COFFEE Token (our native token)\n\n" +
			"You can use any of these to pay for your coffee orders!"

	case strings.Contains(message, "price") || strings.Contains(message, "cost"):
		response = "Here are our current prices:\n\n" +
			"☕ **Coffee Prices:**\n" +
			"• Latte: $4.50 (≈ 0.0001 BTC)\n" +
			"• Cappuccino: $4.00 (≈ 0.0009 ETH)\n" +
			"• Americano: $3.00 (≈ 3 USDC)\n" +
			"• Espresso: $2.50 (≈ 250 COFFEE)\n\n" +
			"💡 Tip: Pay with COFFEE tokens and get 10% discount!"

	case strings.Contains(message, "help"):
		response = "I'm here to help! Here's what I can assist you with:\n\n" +
			"☕ **Coffee Orders:** Browse menu, place orders\n" +
			"💰 **Wallet:** Check balance, send/receive crypto\n" +
			"🔄 **Payments:** Pay with BTC, ETH, USDC, COFFEE\n" +
			"📊 **DeFi:** Stake tokens, earn rewards\n" +
			"🎯 **Support:** Get help with any issues\n\n" +
			"Just ask me anything!"

	case strings.Contains(message, "stake") || strings.Contains(message, "staking"):
		response = "Earn rewards by staking your COFFEE tokens! 🚀\n\n" +
			"💰 **Current APY:** 12%\n" +
			"⏰ **Lock Period:** Flexible (withdraw anytime)\n" +
			"🎁 **Rewards:** Daily COFFEE token rewards\n" +
			"☕ **Bonus:** 20% discount on all coffee orders\n\n" +
			"Minimum stake: 100 COFFEE tokens\n" +
			"Would you like to start staking?"

	case strings.Contains(message, "defi") || strings.Contains(message, "yield"):
		response = "Explore DeFi opportunities with Web3 Coffee! 🌟\n\n" +
			"🏦 **Lending:** Lend USDC, earn 8% APY\n" +
			"🔄 **Liquidity Mining:** Provide ETH/COFFEE LP, earn 15% APY\n" +
			"📈 **Yield Farming:** Multiple strategies available\n" +
			"☕ **Coffee Rewards:** Extra COFFEE tokens for DeFi users\n\n" +
			"Start with as little as $100!"

	case strings.Contains(message, "payment") || strings.Contains(message, "pay"):
		response = "Multiple payment options available! 💳\n\n" +
			"🟡 **Bitcoin (BTC)** - Lightning Network supported\n" +
			"🔵 **Ethereum (ETH)** - Fast transactions\n" +
			"💚 **USDC** - Stable value, low fees\n" +
			"☕ **COFFEE Token** - Get 10% discount!\n" +
			"💳 **Traditional** - Credit/debit cards accepted\n\n" +
			"Which payment method would you prefer?"

	case strings.Contains(message, "order") && strings.Contains(message, "status"):
		response = "Let me check your order status! 📋\n\n" +
			"To track your order, I'll need:\n" +
			"🔍 Order ID or transaction hash\n" +
			"📱 Your wallet address\n\n" +
			"Recent orders are usually ready in 5-10 minutes.\n" +
			"You'll receive a notification when it's ready!"

	case strings.Contains(message, "location") || strings.Contains(message, "store"):
		response = "Find our Web3 Coffee locations! 📍\n\n" +
			"🏪 **Downtown:** 123 Crypto Street\n" +
			"🏪 **Uptown:** 456 Blockchain Ave\n" +
			"🏪 **Westside:** 789 DeFi Plaza\n\n" +
			"⏰ **Hours:** 7 AM - 9 PM daily\n" +
			"🚚 **Delivery:** Available through our app\n" +
			"📱 **Pickup:** Order ahead, skip the line!"

	case strings.Contains(message, "loyalty") || strings.Contains(message, "rewards"):
		response = "Join our Web3 Loyalty Program! 🎁\n\n" +
			"☕ **Earn:** 1 COFFEE token per $1 spent\n" +
			"🎯 **Levels:** Bronze → Silver → Gold → Diamond\n" +
			"💎 **Benefits:** Discounts, free drinks, exclusive access\n" +
			"🚀 **Bonus:** Extra rewards for crypto payments\n\n" +
			"Your loyalty tokens are stored on blockchain!"

	case strings.Contains(message, "nft") || strings.Contains(message, "collectible"):
		response = "Collect exclusive Web3 Coffee NFTs! 🎨\n\n" +
			"☕ **Coffee Art NFTs:** Limited edition designs\n" +
			"🎫 **Membership NFTs:** VIP access and perks\n" +
			"🏆 **Achievement NFTs:** Unlock through activities\n" +
			"💰 **Utility:** Use NFTs for discounts and rewards\n\n" +
			"New collections drop monthly!"

	default:
		response = "I understand you're asking about: \"" + req.Message + "\"\n\n" +
			"I'm your Web3 Coffee AI assistant! ☕ I can help you with:\n" +
			"• Coffee orders and menu\n" +
			"• Crypto payments and wallet\n" +
			"• DeFi opportunities\n" +
			"• Loyalty rewards\n" +
			"• Store locations\n\n" +
			"Could you please be more specific about what you'd like to know?"
	}

	// Add some personality and emojis
	if !strings.Contains(response, "☕") && !strings.Contains(response, "🚀") {
		response += " ☕"
	}

	return &SimpleAIResponse{
		Response: response,
		Metadata: map[string]interface{}{
			"provider":    "simple_ai",
			"model":       "pattern_matching",
			"confidence":  0.85,
			"tokens_used": len(strings.Fields(response)),
		},
		Timestamp: time.Now(),
	}, nil
}

// GetCoffeeRecommendation provides coffee recommendations
func (s *SimpleAIService) GetCoffeeRecommendation(ctx context.Context, preferences map[string]interface{}) (*SimpleAIResponse, error) {
	s.logger.Debug("Getting coffee recommendation", zap.Any("preferences", preferences))

	var recommendation string

	// Simple recommendation logic
	if strength, ok := preferences["strength"].(string); ok {
		switch strings.ToLower(strength) {
		case "strong":
			recommendation = "I recommend our **Espresso** or **Americano** for a strong coffee experience! ☕💪"
		case "mild":
			recommendation = "Try our **Latte** or **Cappuccino** for a smooth, mild coffee experience! ☕😌"
		default:
			recommendation = "Our **Macchiato** offers a perfect balance of strength and smoothness! ☕⚖️"
		}
	} else {
		recommendation = "Based on popular choices, I recommend our signature **Latte** - it's perfectly balanced and loved by 90% of our customers! ☕❤️"
	}

	return &SimpleAIResponse{
		Response: recommendation,
		Metadata: map[string]interface{}{
			"type":        "recommendation",
			"preferences": preferences,
		},
		Timestamp: time.Now(),
	}, nil
}

// AnalyzeSpending analyzes user spending patterns
func (s *SimpleAIService) AnalyzeSpending(ctx context.Context, userID string) (*SimpleAIResponse, error) {
	s.logger.Debug("Analyzing spending patterns", zap.String("user_id", userID))

	// Mock spending analysis
	analysis := "📊 **Your Coffee Spending Analysis:**\n\n" +
		"☕ **This Month:** $45.50 (9 orders)\n" +
		"📈 **Trend:** +15% vs last month\n" +
		"🏆 **Favorite:** Latte (67% of orders)\n" +
		"💰 **Savings:** $12.30 with crypto payments\n" +
		"🎯 **Recommendation:** Try our loyalty program to earn more rewards!"

	return &SimpleAIResponse{
		Response: analysis,
		Metadata: map[string]interface{}{
			"type":    "spending_analysis",
			"user_id": userID,
		},
		Timestamp: time.Now(),
	}, nil
}

// GetMarketInsights provides crypto market insights
func (s *SimpleAIService) GetMarketInsights(ctx context.Context) (*SimpleAIResponse, error) {
	s.logger.Debug("Getting market insights")

	insights := "📈 **Crypto Market Insights:**\n\n" +
		"🟡 **Bitcoin:** $43,250 (+2.5% today)\n" +
		"🔵 **Ethereum:** $2,680 (+1.8% today)\n" +
		"💚 **USDC:** $1.00 (stable)\n" +
		"☕ **COFFEE Token:** $0.125 (+5.2% today)\n\n" +
		"💡 **Tip:** COFFEE token is trending up! Great time to earn rewards through staking."

	return &SimpleAIResponse{
		Response: insights,
		Metadata: map[string]interface{}{
			"type": "market_insights",
		},
		Timestamp: time.Now(),
	}, nil
}

// Close closes the AI service
func (s *SimpleAIService) Close() error {
	s.logger.Debug("Closing simple AI service")
	return nil
}
