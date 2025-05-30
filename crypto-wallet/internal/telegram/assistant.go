package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/internal/ai"
)

// AssistantContext represents the context for AI assistant
type AssistantContext struct {
	UserID          int64             `json:"user_id"`
	ConversationID  string            `json:"conversation_id"`
	Messages        []AssistantMessage `json:"messages"`
	CurrentTopic    string            `json:"current_topic"`
	UserPreferences map[string]string `json:"user_preferences"`
	LastActivity    time.Time         `json:"last_activity"`
}

// AssistantMessage represents a message in the conversation
type AssistantMessage struct {
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	MessageID int       `json:"message_id,omitempty"`
}

// processAIAssistantMessage processes a message through the AI assistant
func (b *Bot) processAIAssistantMessage(ctx context.Context, message string, session *UserSession) error {
	b.logger.Info(fmt.Sprintf("Processing AI assistant message for user %d: %s", session.UserID, message))

	// Get or create assistant context
	assistantCtx := b.getAssistantContext(session)

	// Add user message to conversation
	userMsg := AssistantMessage{
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
	}
	assistantCtx.Messages = append(assistantCtx.Messages, userMsg)

	// Determine conversation context and intent
	intent, err := b.analyzeUserIntent(ctx, message, assistantCtx)
	if err != nil {
		b.logger.Error(fmt.Sprintf("Failed to analyze user intent: %v", err))
		intent = "general_help"
	}

	// Generate contextual response
	response, err := b.generateAssistantResponse(ctx, message, intent, assistantCtx, session)
	if err != nil {
		return fmt.Errorf("failed to generate assistant response: %w", err)
	}

	// Add assistant response to conversation
	assistantMsg := AssistantMessage{
		Role:      "assistant",
		Content:   response.Text,
		Timestamp: time.Now(),
	}
	assistantCtx.Messages = append(assistantCtx.Messages, assistantMsg)

	// Update session context
	assistantCtx.LastActivity = time.Now()
	assistantCtx.CurrentTopic = intent
	session.Context["assistant_context"] = assistantCtx

	// Send response with appropriate keyboard
	keyboard := b.generateContextualKeyboard(intent, session)
	return b.sendMessageWithKeyboard(session.ChatID, response.Text, keyboard)
}

// analyzeUserIntent analyzes user intent from the message
func (b *Bot) analyzeUserIntent(ctx context.Context, message string, assistantCtx *AssistantContext) (string, error) {
	// Create context for intent analysis
	conversationHistory := b.buildConversationHistory(assistantCtx.Messages, 5) // Last 5 messages

	prompt := fmt.Sprintf(`Проаналізуй намір користувача на основі повідомлення та контексту розмови.

Повідомлення користувача: "%s"

Історія розмови:
%s

Поточна тема: %s

Визнач один з наступних намірів:
- coffee_order: Замовлення кави
- payment_help: Допомога з оплатою
- menu_inquiry: Питання про меню
- crypto_help: Допомога з криптовалютами
- order_status: Статус замовлення
- general_help: Загальна допомога
- complaint: Скарга або проблема
- recommendation: Запит рекомендацій

Відповідь надай одним словом (назва наміру).`,
		message,
		conversationHistory,
		assistantCtx.CurrentTopic,
	)

	generateReq := &ai.GenerateRequest{
		UserID:      fmt.Sprintf("%d", assistantCtx.UserID),
		Message:     prompt,
		Context:     "intent_analysis",
		Temperature: 0.3, // Low temperature for consistent intent classification
	}

	response, err := b.aiService.GenerateResponse(ctx, generateReq)
	if err != nil {
		return "", err
	}

	// Clean and validate intent
	intent := strings.TrimSpace(strings.ToLower(response.Text))
	validIntents := []string{
		"coffee_order", "payment_help", "menu_inquiry", "crypto_help",
		"order_status", "general_help", "complaint", "recommendation",
	}

	for _, validIntent := range validIntents {
		if strings.Contains(intent, validIntent) {
			return validIntent, nil
		}
	}

	return "general_help", nil
}

// generateAssistantResponse generates a contextual response
func (b *Bot) generateAssistantResponse(ctx context.Context, message string, intent string, assistantCtx *AssistantContext, session *UserSession) (*ai.GenerateResponse, error) {
	// Build conversation context
	conversationHistory := b.buildConversationHistory(assistantCtx.Messages, 10)
	
	// Create intent-specific prompt
	prompt := b.buildIntentSpecificPrompt(message, intent, conversationHistory, session)

	generateReq := &ai.GenerateRequest{
		UserID:      fmt.Sprintf("%d", session.UserID),
		Message:     prompt,
		Context:     fmt.Sprintf("assistant_%s", intent),
		Temperature: 0.7,
	}

	return b.aiService.GenerateResponse(ctx, generateReq)
}

// buildIntentSpecificPrompt builds a prompt specific to the user's intent
func (b *Bot) buildIntentSpecificPrompt(message, intent, conversationHistory string, session *UserSession) string {
	baseContext := fmt.Sprintf(`Ти - AI асистент кав'ярні Web3 Coffee. Відповідай українською мовою, будь дружнім та корисним.

Повідомлення користувача: "%s"

Історія розмови:
%s

Поточний час: %s
`, message, conversationHistory, time.Now().Format("15:04"))

	switch intent {
	case "coffee_order":
		return baseContext + `
Намір: Замовлення кави

Допоможи користувачу оформити замовлення. Якщо потрібно, запитай:
- Тип кави
- Розмір (small/medium/large)
- Додатки
- Кількість

Меню:
` + b.getMenuText() + `

Після уточнення деталей, запропонуй оформити замовлення.`

	case "payment_help":
		return baseContext + `
Намір: Допомога з оплатою

Допоможи з питаннями оплати криптовалютами:
- Підтримувані валюти: BTC, ETH, USDC, USDT, SOL
- Пояснення процесу оплати
- Допомога з гаманцями
- Комісії та час підтвердження

Будь терплячим та детальним у поясненнях.`

	case "menu_inquiry":
		return baseContext + `
Намір: Питання про меню

Розкажи про наше меню, ціни, інгредієнти:

` + b.getMenuText() + `

Відповідай детально про напої, їх склад та особливості приготування.`

	case "crypto_help":
		return baseContext + `
Намір: Допомога з криптовалютами

Допоможи з питаннями про криптовалюти:
- Як створити гаманець
- Як купити криптовалюту
- Безпека транзакцій
- Курси валют
- Комісії мереж

Пояснюй простими словами, особливо для новачків.`

	case "recommendation":
		return baseContext + `
Намір: Запит рекомендацій

Надай персоналізовані рекомендації на основі:
- Часу дня
- Погоди
- Популярних напоїв
- Особистих переваг (якщо відомі)

Запропонуй 2-3 варіанти з поясненням чому саме ці напої.`

	case "complaint":
		return baseContext + `
Намір: Скарга або проблема

Відреагуй емпатично та професійно:
- Вибач за незручності
- З'ясуй деталі проблеми
- Запропонуй рішення
- Переадресуй до менеджера якщо потрібно

Будь розуміючим та готовим допомогти.`

	default:
		return baseContext + `
Намір: Загальна допомога

Надай корисну інформацію про:
- Наші послуги
- Як зробити замовлення
- Способи оплати
- Час роботи
- Контакти

Будь дружнім та інформативним.`
	}
}

// buildConversationHistory builds conversation history string
func (b *Bot) buildConversationHistory(messages []AssistantMessage, limit int) string {
	if len(messages) == 0 {
		return "Початок розмови"
	}

	start := 0
	if len(messages) > limit {
		start = len(messages) - limit
	}

	var history strings.Builder
	for i := start; i < len(messages); i++ {
		msg := messages[i]
		role := "Користувач"
		if msg.Role == "assistant" {
			role = "Асистент"
		}
		history.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
	}

	return history.String()
}

// generateContextualKeyboard generates keyboard based on conversation context
func (b *Bot) generateContextualKeyboard(intent string, session *UserSession) tgbotapi.InlineKeyboardMarkup {
	switch intent {
	case "coffee_order":
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("☕️ Популярні напої", "popular_drinks"),
				tgbotapi.NewInlineKeyboardButtonData("📋 Повне меню", "full_menu"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🎯 Рекомендації", "get_recommendations"),
				tgbotapi.NewInlineKeyboardButtonData("🛒 Оформити замовлення", "start_order"),
			),
		)

	case "payment_help":
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💳 Способи оплати", "payment_methods"),
				tgbotapi.NewInlineKeyboardButtonData("📱 Створити гаманець", "create_wallet"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💰 Курси валют", "exchange_rates"),
				tgbotapi.NewInlineKeyboardButtonData("🔒 Безпека", "security_tips"),
			),
		)

	case "menu_inquiry":
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("☕️ Еспресо напої", "category_espresso"),
				tgbotapi.NewInlineKeyboardButtonData("🥛 Молочні напої", "category_milk"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("❄️ Холодні напої", "category_cold"),
				tgbotapi.NewInlineKeyboardButtonData("💰 Ціни", "show_prices"),
			),
		)

	default:
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("☕️ Замовити каву", "start_order"),
				tgbotapi.NewInlineKeyboardButtonData("📋 Меню", "show_menu"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💳 Допомога з оплатою", "payment_help"),
				tgbotapi.NewInlineKeyboardButtonData("📞 Зв'язатися з нами", "contact_support"),
			),
		)
	}
}

// getAssistantContext gets or creates assistant context for the session
func (b *Bot) getAssistantContext(session *UserSession) *AssistantContext {
	if ctx, exists := session.Context["assistant_context"].(*AssistantContext); exists {
		return ctx
	}

	// Create new assistant context
	return &AssistantContext{
		UserID:          session.UserID,
		ConversationID:  fmt.Sprintf("conv_%d_%d", session.UserID, time.Now().Unix()),
		Messages:        []AssistantMessage{},
		CurrentTopic:    "general",
		UserPreferences: make(map[string]string),
		LastActivity:    time.Now(),
	}
}

// handleSmartSuggestions provides smart suggestions based on context
func (b *Bot) handleSmartSuggestions(ctx context.Context, session *UserSession) error {
	assistantCtx := b.getAssistantContext(session)
	
	// Analyze conversation to provide smart suggestions
	suggestions, err := b.generateSmartSuggestions(ctx, assistantCtx, session)
	if err != nil {
		return fmt.Errorf("failed to generate suggestions: %w", err)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💡 Показати поради", "show_tips"),
			tgbotapi.NewInlineKeyboardButtonData("🎯 Персональні рекомендації", "personal_recommendations"),
		),
	)

	return b.sendMessageWithKeyboard(session.ChatID, suggestions, keyboard)
}

// generateSmartSuggestions generates context-aware suggestions
func (b *Bot) generateSmartSuggestions(ctx context.Context, assistantCtx *AssistantContext, session *UserSession) (string, error) {
	prompt := fmt.Sprintf(`На основі контексту розмови надай 3-4 корисні поради або пропозиції для користувача.

Поточна тема: %s
Час дня: %s
Історія розмови: %s

Поради мають бути:
- Релевантними до контексту
- Корисними для користувача
- Пов'язаними з кавою або нашими послугами
- Короткими та зрозумілими

Формат: пронумерований список українською мовою.`,
		assistantCtx.CurrentTopic,
		b.getTimeOfDay(),
		b.buildConversationHistory(assistantCtx.Messages, 5),
	)

	generateReq := &ai.GenerateRequest{
		UserID:      fmt.Sprintf("%d", session.UserID),
		Message:     prompt,
		Context:     "smart_suggestions",
		Temperature: 0.8,
	}

	response, err := b.aiService.GenerateResponse(ctx, generateReq)
	if err != nil {
		return "", err
	}

	return "💡 *Корисні поради:*\n\n" + response.Text, nil
}
