package telegram

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/ai"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/defi"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/wallet"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/redis"
)

// Bot represents the Telegram bot
type Bot struct {
	api    *tgbotapi.BotAPI
	config config.TelegramConfig
	logger *logger.Logger
	cache  redis.Client

	// Services
	aiService     ai.Service
	walletService *wallet.Service
	defiService   *defi.Service

	// Bot state
	userSessions map[int64]*UserSession
	mutex        sync.RWMutex

	// Channels
	updates  tgbotapi.UpdatesChannel
	shutdown chan struct{}
}

// UserSession represents a user session
type UserSession struct {
	UserID       int64                  `json:"user_id"`
	ChatID       int64                  `json:"chat_id"`
	Username     string                 `json:"username"`
	State        SessionState           `json:"state"`
	Context      map[string]interface{} `json:"context"`
	LastActivity time.Time              `json:"last_activity"`
	WalletID     string                 `json:"wallet_id,omitempty"`
	Language     string                 `json:"language"`
}

// SessionState represents the current state of user session
type SessionState string

const (
	StateIdle              SessionState = "idle"
	StateWaitingWallet     SessionState = "waiting_wallet"
	StateOrderingCoffee    SessionState = "ordering_coffee"
	StateProcessingPayment SessionState = "processing_payment"
	StateConfirmingOrder   SessionState = "confirming_order"
)

// NewBot creates a new Telegram bot
func NewBot(
	cfg config.TelegramConfig,
	logger *logger.Logger,
	cache redis.Client,
	aiService ai.Service,
	walletService *wallet.Service,
	defiService *defi.Service,
) (*Bot, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("telegram bot is disabled")
	}

	if cfg.BotToken == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}

	// Create bot API
	botAPI, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	botAPI.Debug = cfg.Debug

	bot := &Bot{
		api:           botAPI,
		config:        cfg,
		logger:        logger,
		cache:         cache,
		aiService:     aiService,
		walletService: walletService,
		defiService:   defiService,
		userSessions:  make(map[int64]*UserSession),
		shutdown:      make(chan struct{}),
	}

	logger.Info(fmt.Sprintf("Telegram bot initialized: @%s", botAPI.Self.UserName))
	return bot, nil
}

// Start starts the Telegram bot
func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("Starting Telegram bot...")

	// Set up commands
	if err := b.setupCommands(); err != nil {
		return fmt.Errorf("failed to setup commands: %w", err)
	}

	// Configure updates
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = b.config.Timeout
	updateConfig.AllowedUpdates = b.config.AllowedUpdates

	// Get updates channel
	updates := b.api.GetUpdatesChan(updateConfig)
	b.updates = updates

	// Start processing updates
	go b.processUpdates(ctx)

	// Start session cleanup
	go b.sessionCleanup(ctx)

	b.logger.Info("Telegram bot started successfully")
	return nil
}

// Stop stops the Telegram bot
func (b *Bot) Stop() error {
	b.logger.Info("Stopping Telegram bot...")

	close(b.shutdown)

	if b.updates != nil {
		b.api.StopReceivingUpdates()
	}

	b.logger.Info("Telegram bot stopped")
	return nil
}

// processUpdates processes incoming updates
func (b *Bot) processUpdates(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-b.shutdown:
			return
		case update := <-b.updates:
			go b.handleUpdate(ctx, update)
		}
	}
}

// handleUpdate handles a single update
func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			b.logger.Error(fmt.Sprintf("Panic in update handler: %v", r))
		}
	}()

	// Handle different types of updates
	if update.Message != nil {
		b.handleMessage(ctx, update.Message)
	} else if update.CallbackQuery != nil {
		b.handleCallbackQuery(ctx, update.CallbackQuery)
	} else if update.InlineQuery != nil {
		// Handle inline queries (for now, just ignore them)
		b.logger.Debug(fmt.Sprintf("Received inline query: %s", update.InlineQuery.Query))
	}
}

// handleMessage handles text messages
func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	b.logger.Info(fmt.Sprintf("Received message from user %d: %s", userID, message.Text))

	// Get or create user session
	session := b.getOrCreateSession(userID, chatID, message.From.UserName)

	// Update last activity
	session.LastActivity = time.Now()

	// Handle commands
	if message.IsCommand() {
		b.handleCommand(ctx, message, session)
		return
	}

	// Handle regular messages based on session state
	switch session.State {
	case StateIdle:
		b.handleIdleMessage(ctx, message, session)
	case StateWaitingWallet:
		b.handleWalletMessage(ctx, message, session)
	case StateOrderingCoffee:
		b.handleCoffeeOrderMessage(ctx, message, session)
	case StateProcessingPayment:
		b.handlePaymentMessage(ctx, message, session)
	case StateConfirmingOrder:
		b.handleOrderConfirmationMessage(ctx, message, session)
	default:
		b.handleIdleMessage(ctx, message, session)
	}
}

// getOrCreateSession gets or creates a user session
func (b *Bot) getOrCreateSession(userID int64, chatID int64, username string) *UserSession {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	session, exists := b.userSessions[userID]
	if !exists {
		session = &UserSession{
			UserID:       userID,
			ChatID:       chatID,
			Username:     username,
			State:        StateIdle,
			Context:      make(map[string]interface{}),
			LastActivity: time.Now(),
			Language:     "uk", // Default to Ukrainian
		}
		b.userSessions[userID] = session
	}

	return session
}

// sendMessage sends a message to the user
func (b *Bot) sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown

	_, err := b.api.Send(msg)
	if err != nil {
		b.logger.Error(fmt.Sprintf("Failed to send message: %v", err))
		return err
	}

	return nil
}

// sendMessageWithKeyboard sends a message with inline keyboard
func (b *Bot) sendMessageWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = keyboard

	_, err := b.api.Send(msg)
	if err != nil {
		b.logger.Error(fmt.Sprintf("Failed to send message with keyboard: %v", err))
		return err
	}

	return nil
}

// setupCommands sets up bot commands
func (b *Bot) setupCommands() error {
	var commands []tgbotapi.BotCommand

	for _, cmd := range b.config.Commands {
		commands = append(commands, tgbotapi.BotCommand{
			Command:     cmd.Command,
			Description: cmd.Description,
		})
	}

	setCommands := tgbotapi.NewSetMyCommands(commands...)
	_, err := b.api.Request(setCommands)
	if err != nil {
		return fmt.Errorf("failed to set commands: %w", err)
	}

	b.logger.Info(fmt.Sprintf("Set %d bot commands", len(commands)))
	return nil
}

// sessionCleanup cleans up inactive sessions
func (b *Bot) sessionCleanup(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-b.shutdown:
			return
		case <-ticker.C:
			b.cleanupInactiveSessions()
		}
	}
}

// cleanupInactiveSessions removes inactive sessions
func (b *Bot) cleanupInactiveSessions() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	cutoff := time.Now().Add(-2 * time.Hour) // Remove sessions older than 2 hours

	for userID, session := range b.userSessions {
		if session.LastActivity.Before(cutoff) {
			delete(b.userSessions, userID)
			b.logger.Debug(fmt.Sprintf("Cleaned up inactive session for user %d", userID))
		}
	}
}

// GetBotInfo returns bot information
func (b *Bot) GetBotInfo() map[string]interface{} {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return map[string]interface{}{
		"username":        b.api.Self.UserName,
		"active_sessions": len(b.userSessions),
		"debug_mode":      b.config.Debug,
		"webhook_url":     b.config.WebhookURL,
	}
}

// BroadcastMessage sends a message to all active users
func (b *Bot) BroadcastMessage(ctx context.Context, message string) error {
	b.mutex.RLock()
	sessions := make([]*UserSession, 0, len(b.userSessions))
	for _, session := range b.userSessions {
		sessions = append(sessions, session)
	}
	b.mutex.RUnlock()

	var errors []error
	for _, session := range sessions {
		if err := b.sendMessage(session.ChatID, message); err != nil {
			errors = append(errors, fmt.Errorf("failed to send to user %d: %w", session.UserID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("broadcast errors: %v", errors)
	}

	b.logger.Info(fmt.Sprintf("Broadcast message sent to %d users", len(sessions)))
	return nil
}

// handleIdleMessage handles messages when user is in idle state
func (b *Bot) handleIdleMessage(ctx context.Context, message *tgbotapi.Message, session *UserSession) {
	// Use AI to process general queries
	aiReq := &ai.GeneralQueryRequest{
		UserID:    fmt.Sprintf("%d", session.UserID),
		Message:   message.Text,
		Context:   "general",
		ChatID:    session.ChatID,
		MessageID: message.MessageID,
	}

	// Generate AI response
	generateReq := &ai.GenerateRequest{
		UserID:      aiReq.UserID,
		Message:     aiReq.Message,
		Context:     aiReq.Context,
		Temperature: 0.7,
	}

	response, err := b.aiService.GenerateResponse(ctx, generateReq)
	if err != nil {
		b.logger.Error(fmt.Sprintf("AI response failed: %v", err))
		b.sendMessage(session.ChatID, "Вибачте, виникла помилка. Спробуйте ще раз або використайте команди /help")
		return
	}

	// Send AI response
	b.sendMessage(session.ChatID, response.Text)
}

// handleWalletMessage handles messages when user is setting up wallet
func (b *Bot) handleWalletMessage(ctx context.Context, message *tgbotapi.Message, session *UserSession) {
	// Process wallet-related messages
	text := message.Text

	if strings.Contains(strings.ToLower(text), "seed") || strings.Contains(strings.ToLower(text), "фраза") {
		// User might be providing seed phrase
		session.Context["seed_phrase"] = text
		session.State = StateIdle

		b.sendMessage(session.ChatID, "🔐 Seed фразу отримано. Імпортую гаманець...")

		// Here you would call wallet service to import wallet
		// For now, just simulate success
		session.WalletID = "imported_wallet_" + fmt.Sprintf("%d", session.UserID)

		b.sendMessage(session.ChatID, "✅ Гаманець успішно імпортовано! Використовуйте /balance для перевірки балансу.")
	} else {
		// Use AI to help with wallet setup
		aiReq := &ai.WalletQueryRequest{
			UserID:    fmt.Sprintf("%d", session.UserID),
			Message:   text,
			ChatID:    session.ChatID,
			MessageID: message.MessageID,
		}

		generateReq := &ai.GenerateRequest{
			UserID:      aiReq.UserID,
			Message:     aiReq.Message,
			Context:     "wallet_query",
			Temperature: 0.5,
		}

		response, err := b.aiService.GenerateResponse(ctx, generateReq)
		if err != nil {
			b.logger.Error(fmt.Sprintf("AI wallet response failed: %v", err))
			b.sendMessage(session.ChatID, "Для налаштування гаманця використовуйте команду /wallet")
			return
		}

		b.sendMessage(session.ChatID, response.Text)
	}
}

// handleCoffeeOrderMessage handles messages when user is ordering coffee
func (b *Bot) handleCoffeeOrderMessage(ctx context.Context, message *tgbotapi.Message, session *UserSession) {
	// Use AI to process coffee order
	aiReq := &ai.CoffeeOrderRequest{
		UserID:    fmt.Sprintf("%d", session.UserID),
		Message:   message.Text,
		ChatID:    session.ChatID,
		MessageID: message.MessageID,
	}

	response, err := b.aiService.ProcessCoffeeOrder(ctx, aiReq)
	if err != nil {
		b.logger.Error(fmt.Sprintf("AI coffee order failed: %v", err))
		b.sendMessage(session.ChatID, "Вибачте, не вдалося обробити замовлення. Спробуйте ще раз або оберіть з меню /menu")
		return
	}

	// Send AI response
	b.sendMessage(session.ChatID, response.AIResponse)

	// If order details were parsed, show confirmation
	if response.OrderDetails != nil {
		order := response.OrderDetails
		confirmText := fmt.Sprintf(`✅ *Підтвердження замовлення*

*Ваше замовлення:*
• Напій: %s
• Розмір: %s
• Кількість: %d
• Додатки: %s

*Вартість:* $%.2f USD

Підтвердити замовлення?`,
			order.CoffeeType,
			order.Size,
			order.Quantity,
			strings.Join(order.Extras, ", "),
			order.EstimatedPriceUSD,
		)

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅ Підтвердити", "confirm_order"),
				tgbotapi.NewInlineKeyboardButtonData("✏️ Змінити", "modify_order"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("❌ Скасувати", "cancel_order"),
			),
		)

		session.Context["pending_order"] = order
		session.State = StateConfirmingOrder

		b.sendMessageWithKeyboard(session.ChatID, confirmText, keyboard)
	}
}

// handlePaymentMessage handles messages during payment processing
func (b *Bot) handlePaymentMessage(ctx context.Context, message *tgbotapi.Message, session *UserSession) {
	// Use AI to help with payment queries
	aiReq := &ai.PaymentQueryRequest{
		UserID:    fmt.Sprintf("%d", session.UserID),
		Message:   message.Text,
		ChatID:    session.ChatID,
		MessageID: message.MessageID,
	}

	response, err := b.aiService.ProcessPaymentQuery(ctx, aiReq)
	if err != nil {
		b.logger.Error(fmt.Sprintf("AI payment query failed: %v", err))
		b.sendMessage(session.ChatID, "Для допомоги з платежами використовуйте /pay або /help")
		return
	}

	b.sendMessage(session.ChatID, response.AIResponse)
}

// handleOrderConfirmationMessage handles messages during order confirmation
func (b *Bot) handleOrderConfirmationMessage(ctx context.Context, message *tgbotapi.Message, session *UserSession) {
	text := strings.ToLower(message.Text)

	if strings.Contains(text, "так") || strings.Contains(text, "підтверджую") || strings.Contains(text, "yes") {
		// Confirm order
		b.confirmOrder(ctx, session)
	} else if strings.Contains(text, "ні") || strings.Contains(text, "скасувати") || strings.Contains(text, "no") {
		// Cancel order
		session.State = StateIdle
		delete(session.Context, "pending_order")
		b.sendMessage(session.ChatID, "❌ Замовлення скасовано. Використовуйте /coffee для нового замовлення.")
	} else {
		// Use AI to understand the message
		generateReq := &ai.GenerateRequest{
			UserID:      fmt.Sprintf("%d", session.UserID),
			Message:     message.Text,
			Context:     "order_confirmation",
			Temperature: 0.3,
		}

		response, err := b.aiService.GenerateResponse(ctx, generateReq)
		if err != nil {
			b.sendMessage(session.ChatID, "Будь ласка, підтвердіть замовлення (Так/Ні) або скасуйте його.")
			return
		}

		b.sendMessage(session.ChatID, response.Text)
	}
}

// confirmOrder confirms and processes the order
func (b *Bot) confirmOrder(ctx context.Context, session *UserSession) {
	_, exists := session.Context["pending_order"]
	if !exists {
		b.sendMessage(session.ChatID, "❌ Помилка: замовлення не знайдено.")
		return
	}

	// Convert to proper type (in real implementation, you'd have proper type handling)
	session.State = StateProcessingPayment

	confirmText := fmt.Sprintf(`🎉 *Замовлення підтверджено!*

Переходимо до оплати. Оберіть спосіб оплати:`)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("₿ Bitcoin", "pay_btc"),
			tgbotapi.NewInlineKeyboardButtonData("Ξ Ethereum", "pay_eth"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💵 USDC", "pay_usdc"),
			tgbotapi.NewInlineKeyboardButtonData("💵 USDT", "pay_usdt"),
		),
	)

	b.sendMessageWithKeyboard(session.ChatID, confirmText, keyboard)
}
