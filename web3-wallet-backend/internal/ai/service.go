package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/config"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/redis"
)

// Service provides AI operations using multiple providers
type Service struct {
	config config.AIConfig
	logger *logger.Logger
	cache  redis.Client

	// AI Providers
	geminiClient    *GeminiClient
	ollamaClient    *OllamaClient
	langchainClient *LangChainClient

	// State
	mutex sync.RWMutex
}

// NewService creates a new AI service
func NewService(cfg config.AIConfig, logger *logger.Logger, cache redis.Client) (*Service, error) {
	service := &Service{
		config: cfg,
		logger: logger,
		cache:  cache,
	}

	// Initialize providers
	if cfg.Gemini.Enabled {
		geminiClient, err := NewGeminiClient(cfg.Gemini, logger)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to initialize Gemini client: %v", err))
		} else {
			service.geminiClient = geminiClient
		}
	}

	if cfg.Ollama.Enabled {
		ollamaClient, err := NewOllamaClient(cfg.Ollama, logger)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to initialize Ollama client: %v", err))
		} else {
			service.ollamaClient = ollamaClient
		}
	}

	if cfg.LangChain.Enabled {
		langchainClient, err := NewLangChainClient(cfg.LangChain, logger)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to initialize LangChain client: %v", err))
		} else {
			service.langchainClient = langchainClient
		}
	}

	return service, nil
}

// GenerateResponse generates a response using the configured AI provider
func (s *Service) GenerateResponse(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	s.logger.Info(fmt.Sprintf("Generating AI response for user %s", req.UserID))

	// Check cache first
	if s.config.Service.CacheEnabled {
		cacheKey := fmt.Sprintf("ai:response:%s:%s", req.UserID, req.MessageHash())
		if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
			s.logger.Debug(fmt.Sprintf("Cache hit for AI response: %s", cacheKey))
			return &GenerateResponse{
				Text:     cached,
				Provider: "cache",
				Cached:   true,
			}, nil
		}
	}

	// Try primary provider
	response, err := s.generateWithProvider(ctx, req, s.config.Service.DefaultProvider)
	if err != nil {
		s.logger.Warn(fmt.Sprintf("Primary provider failed: %v", err))

		// Try fallback provider
		if s.config.Service.FallbackProvider != "" && s.config.Service.FallbackProvider != s.config.Service.DefaultProvider {
			response, err = s.generateWithProvider(ctx, req, s.config.Service.FallbackProvider)
			if err != nil {
				s.logger.Error(fmt.Sprintf("Fallback provider also failed: %v", err))
				return nil, fmt.Errorf("all AI providers failed: %w", err)
			}
		} else {
			return nil, fmt.Errorf("primary AI provider failed: %w", err)
		}
	}

	// Cache the response
	if s.config.Service.CacheEnabled && response.Text != "" {
		cacheKey := fmt.Sprintf("ai:response:%s:%s", req.UserID, req.MessageHash())
		cacheTTL, _ := time.ParseDuration(s.config.Service.CacheTTL)
		if err := s.cache.Set(ctx, cacheKey, response.Text, cacheTTL); err != nil {
			s.logger.Warn(fmt.Sprintf("Failed to cache AI response: %v", err))
		}
	}

	return response, nil
}

// generateWithProvider generates response using specific provider
func (s *Service) generateWithProvider(ctx context.Context, req *GenerateRequest, provider string) (*GenerateResponse, error) {
	switch provider {
	case "gemini":
		if s.geminiClient == nil {
			return nil, fmt.Errorf("gemini client not initialized")
		}
		return s.geminiClient.GenerateResponse(ctx, req)

	case "ollama":
		if s.ollamaClient == nil {
			return nil, fmt.Errorf("ollama client not initialized")
		}
		return s.ollamaClient.GenerateResponse(ctx, req)

	case "langchain":
		if s.langchainClient == nil {
			return nil, fmt.Errorf("langchain client not initialized")
		}
		return s.langchainClient.GenerateResponse(ctx, req)

	default:
		return nil, fmt.Errorf("unknown AI provider: %s", provider)
	}
}

// ProcessCoffeeOrder processes a coffee order request with AI assistance
func (s *Service) ProcessCoffeeOrder(ctx context.Context, req *CoffeeOrderRequest) (*CoffeeOrderResponse, error) {
	s.logger.Info(fmt.Sprintf("Processing coffee order with AI for user %s", req.UserID))

	// Create AI prompt for coffee order processing
	prompt := fmt.Sprintf(`
Ти - асистент для замовлення кави в Web3 кав'ярні. Користувач хоче замовити каву.

Повідомлення користувача: "%s"

Проаналізуй запит та надай структуровану відповідь у форматі JSON з наступними полями:
- coffee_type: тип кави (latte, cappuccino, espresso, americano, etc.)
- size: розмір (small, medium, large)
- extras: додатки (milk, sugar, syrup, etc.)
- quantity: кількість
- estimated_price_usd: приблизна ціна в USD
- payment_suggestion: рекомендована криптовалюта для оплати
- friendly_response: дружня відповідь українською мовою

Якщо запит незрозумілий, попроси уточнення.
`, req.Message)

	generateReq := &GenerateRequest{
		UserID:      req.UserID,
		Message:     prompt,
		Context:     "coffee_order",
		Temperature: 0.3, // Lower temperature for more consistent responses
	}

	response, err := s.GenerateResponse(ctx, generateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI response: %w", err)
	}

	return &CoffeeOrderResponse{
		AIResponse:  response.Text,
		Provider:    response.Provider,
		ProcessedAt: time.Now(),
		Confidence:  response.Confidence,
		Suggestions: response.Suggestions,
	}, nil
}

// ProcessPaymentQuery processes a payment-related query
func (s *Service) ProcessPaymentQuery(ctx context.Context, req *PaymentQueryRequest) (*PaymentQueryResponse, error) {
	s.logger.Info(fmt.Sprintf("Processing payment query with AI for user %s", req.UserID))

	prompt := fmt.Sprintf(`
Ти - експерт з криптовалют та Web3 платежів. Користувач має питання про платежі.

Повідомлення користувача: "%s"

Контекст:
- Баланс гаманця: %s
- Доступні токени: %v
- Поточні ціни: %v

Надай корисну відповідь українською мовою про:
- Рекомендації щодо оплати
- Поточні курси криптовалют
- Комісії за транзакції
- Безпеку платежів
- Альтернативні варіанти оплати
`, req.Message, req.WalletBalance, req.AvailableTokens, req.CurrentPrices)

	generateReq := &GenerateRequest{
		UserID:      req.UserID,
		Message:     prompt,
		Context:     "payment_query",
		Temperature: 0.5,
	}

	response, err := s.GenerateResponse(ctx, generateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI response: %w", err)
	}

	return &PaymentQueryResponse{
		AIResponse:  response.Text,
		Provider:    response.Provider,
		ProcessedAt: time.Now(),
		Confidence:  response.Confidence,
	}, nil
}

// GetProviderStatus returns the status of all AI providers
func (s *Service) GetProviderStatus(ctx context.Context) map[string]bool {
	status := make(map[string]bool)

	if s.geminiClient != nil {
		status["gemini"] = s.geminiClient.IsHealthy(ctx)
	}

	if s.ollamaClient != nil {
		status["ollama"] = s.ollamaClient.IsHealthy(ctx)
	}

	if s.langchainClient != nil {
		status["langchain"] = s.langchainClient.IsHealthy(ctx)
	}

	return status
}

// Close closes all AI provider connections
func (s *Service) Close() error {
	var errors []error

	if s.geminiClient != nil {
		if err := s.geminiClient.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if s.ollamaClient != nil {
		if err := s.ollamaClient.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if s.langchainClient != nil {
		if err := s.langchainClient.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing AI providers: %v", errors)
	}

	return nil
}
