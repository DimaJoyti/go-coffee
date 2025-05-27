package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/config"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
)

// LangChainClient represents a LangChain Go client
type LangChainClient struct {
	llm    llms.Model
	config config.LangChainConfig
	logger *logger.Logger
}

// NewLangChainClient creates a new LangChain client
func NewLangChainClient(cfg config.LangChainConfig, logger *logger.Logger) (*LangChainClient, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("langchain client is disabled")
	}

	// Initialize OpenAI LLM (you can extend this to support other providers)
	llm, err := openai.New(
		openai.WithModel(cfg.Model),
		openai.WithTemperature(cfg.Temperature),
		openai.WithMaxTokens(cfg.MaxTokens),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create LangChain LLM: %w", err)
	}

	client := &LangChainClient{
		llm:    llm,
		config: cfg,
		logger: logger,
	}

	logger.Info("LangChain client initialized successfully")
	return client, nil
}

// GenerateResponse generates a response using LangChain
func (l *LangChainClient) GenerateResponse(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	l.logger.Info(fmt.Sprintf("Generating LangChain response for user %s", req.UserID))

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(l.config.Timeout)*time.Second)
	defer cancel()

	// Prepare options
	var options []llms.CallOption
	
	if req.Temperature > 0 {
		options = append(options, llms.WithTemperature(req.Temperature))
	}
	
	if req.MaxTokens > 0 {
		options = append(options, llms.WithMaxTokens(req.MaxTokens))
	}

	// Add context-specific system message
	systemMessage := l.getSystemMessage(req.Context)
	
	// Create messages
	messages := []schema.ChatMessage{
		schema.SystemChatMessage{Content: systemMessage},
		schema.HumanChatMessage{Content: req.Message},
	}

	// Generate response
	resp, err := l.llm.GenerateContent(timeoutCtx, messages, options...)
	if err != nil {
		l.logger.Error(fmt.Sprintf("LangChain generation failed: %v", err))
		return nil, fmt.Errorf("langchain generation failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices from LangChain")
	}

	choice := resp.Choices[0]
	if choice.Content == "" {
		return nil, fmt.Errorf("empty response from LangChain")
	}

	// Calculate confidence based on response metadata
	confidence := l.calculateConfidence(resp)

	response := &GenerateResponse{
		Text:        choice.Content,
		Provider:    "langchain",
		Confidence:  confidence,
		GeneratedAt: time.Now(),
		Metadata: map[string]string{
			"model":        l.config.Model,
			"finish_reason": choice.FinishReason,
		},
	}

	// Add token usage information if available
	if resp.Usage.TotalTokens > 0 {
		response.Metadata["total_tokens"] = fmt.Sprintf("%d", resp.Usage.TotalTokens)
		response.Metadata["prompt_tokens"] = fmt.Sprintf("%d", resp.Usage.PromptTokens)
		response.Metadata["completion_tokens"] = fmt.Sprintf("%d", resp.Usage.CompletionTokens)
	}

	l.logger.Info(fmt.Sprintf("LangChain response generated successfully for user %s", req.UserID))
	return response, nil
}

// GenerateStreamResponse generates a streaming response using LangChain
func (l *LangChainClient) GenerateStreamResponse(ctx context.Context, req *GenerateRequest, callback func(string) error) error {
	l.logger.Info(fmt.Sprintf("Generating LangChain streaming response for user %s", req.UserID))

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(l.config.Timeout)*time.Second)
	defer cancel()

	// Prepare options
	var options []llms.CallOption
	
	if req.Temperature > 0 {
		options = append(options, llms.WithTemperature(req.Temperature))
	}
	
	if req.MaxTokens > 0 {
		options = append(options, llms.WithMaxTokens(req.MaxTokens))
	}

	// Add streaming option
	options = append(options, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		return callback(string(chunk))
	}))

	// Add context-specific system message
	systemMessage := l.getSystemMessage(req.Context)
	
	// Create messages
	messages := []schema.ChatMessage{
		schema.SystemChatMessage{Content: systemMessage},
		schema.HumanChatMessage{Content: req.Message},
	}

	// Generate streaming response
	_, err := l.llm.GenerateContent(timeoutCtx, messages, options...)
	if err != nil {
		l.logger.Error(fmt.Sprintf("LangChain streaming failed: %v", err))
		return fmt.Errorf("langchain streaming failed: %w", err)
	}

	l.logger.Info(fmt.Sprintf("LangChain streaming response completed for user %s", req.UserID))
	return nil
}

// IsHealthy checks if the LangChain client is healthy
func (l *LangChainClient) IsHealthy(ctx context.Context) bool {
	// Simple health check by generating a minimal response
	testReq := &GenerateRequest{
		UserID:  "health_check",
		Message: "Hello",
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := l.GenerateResponse(timeoutCtx, testReq)
	return err == nil
}

// Close closes the LangChain client
func (l *LangChainClient) Close() error {
	// LangChain Go doesn't require explicit closing
	return nil
}

// getSystemMessage returns context-specific system message
func (l *LangChainClient) getSystemMessage(context string) string {
	switch context {
	case "coffee_order":
		return `Ти - асистент для замовлення кави в Web3 кав'ярні. Твоя мета - допомогти користувачам замовити каву та оплатити її криптовалютою.

Основні функції:
- Допомагати з вибором кави та додатків
- Пояснювати процес оплати криптовалютою
- Надавати інформацію про меню та ціни
- Відповідати українською мовою дружньо та професійно

Завжди будь ввічливим, корисним та точним у своїх відповідях.`

	case "payment_query":
		return `Ти - експерт з криптовалют та Web3 платежів. Твоя мета - допомогти користувачам з питаннями про криптоплатежі.

Основні функції:
- Пояснювати процеси криптоплатежів
- Надавати інформацію про курси валют
- Рекомендувати оптимальні способи оплати
- Пояснювати комісії та час транзакцій
- Давати поради з безпеки

Завжди надавай точну та актуальну інформацію українською мовою.`

	case "wallet_query":
		return `Ти - помічник з управління Web3 гаманцями. Твоя мета - допомогти користувачам з питаннями про гаманці.

Основні функції:
- Пояснювати функції гаманця
- Допомагати з налаштуваннями безпеки
- Надавати інформацію про баланси та транзакції
- Пояснювати процеси відправки та отримання коштів

Завжди підкреслюй важливість безпеки та відповідай українською мовою.`

	case "menu_query":
		return `Ти - консультант з меню кав'ярні. Твоя мета - допомогти користувачам вибрати каву та напої.

Основні функції:
- Описувати різні види кави та напоїв
- Рекомендувати напої на основі смаків користувача
- Надавати інформацію про інгредієнти та алергени
- Пояснювати різницю між різними видами кави

Завжди будь ентузіастичним щодо кави та відповідай українською мовою.`

	default:
		return `Ти - корисний асистент Web3 кав'ярні. Твоя мета - допомогти користувачам з будь-якими питаннями про замовлення кави, криптоплатежі та використання платформи.

Основні принципи:
- Завжди відповідай українською мовою
- Будь дружнім та професійним
- Надавай точну та корисну інформацію
- Якщо не знаєш відповіді, чесно про це скажи

Ти можеш допомогти з:
- Замовленням кави
- Криптоплатежами
- Управлінням гаманцем
- Інформацією про меню
- Загальними питаннями про платформу`
	}
}

// calculateConfidence calculates confidence score based on response quality
func (l *LangChainClient) calculateConfidence(resp *llms.ContentResponse) float64 {
	confidence := 1.0

	if len(resp.Choices) == 0 {
		return 0.0
	}

	choice := resp.Choices[0]

	// Reduce confidence based on finish reason
	switch choice.FinishReason {
	case "stop":
		// Normal completion, no reduction
	case "length":
		confidence *= 0.8 // Slightly reduce for truncated responses
	case "content_filter":
		confidence *= 0.3 // Significantly reduce for filtered responses
	default:
		confidence *= 0.6 // Reduce for other issues
	}

	// Adjust confidence based on response length
	responseLength := len(choice.Content)
	if responseLength < 10 {
		confidence *= 0.5 // Very short responses might be incomplete
	} else if responseLength > 2000 {
		confidence *= 0.9 // Very long responses might be verbose
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

// CreateConversationChain creates a conversation chain for context-aware responses
func (l *LangChainClient) CreateConversationChain(ctx context.Context, conversationHistory []schema.ChatMessage) error {
	// This would implement conversation memory/chain functionality
	// For now, it's a placeholder for future implementation
	l.logger.Info("Creating conversation chain (placeholder)")
	return nil
}

// GetModelInfo returns information about the current model
func (l *LangChainClient) GetModelInfo() map[string]string {
	return map[string]string{
		"model":       l.config.Model,
		"temperature": fmt.Sprintf("%.2f", l.config.Temperature),
		"max_tokens":  fmt.Sprintf("%d", l.config.MaxTokens),
		"timeout":     fmt.Sprintf("%ds", l.config.Timeout),
		"provider":    "langchain",
	}
}
