package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MockBot represents a simplified Telegram bot for testing
type MockBot struct {
	api    *tgbotapi.BotAPI
	logger *log.Logger
}

// UserSession represents a simplified user session
type UserSession struct {
	UserID       int64                  `json:"user_id"`
	ChatID       int64                  `json:"chat_id"`
	Username     string                 `json:"username"`
	State        string                 `json:"state"`
	Context      map[string]interface{} `json:"context"`
	LastActivity time.Time              `json:"last_activity"`
}

// Session states
const (
	StateIdle           = "idle"
	StateOrderingCoffee = "ordering_coffee"
	StateProcessingPayment = "processing_payment"
)

// Coffee menu items
var coffeeMenu = map[string]float64{
	"espresso":   3.50,
	"americano":  4.00,
	"latte":      5.50,
	"cappuccino": 5.00,
	"mocha":      6.00,
}

func main() {
	// Get bot token from environment
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	// Create bot API
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	bot.Debug = true
	logger := log.New(os.Stdout, "[TG-BOT] ", log.LstdFlags)

	mockBot := &MockBot{
		api:    bot,
		logger: logger,
	}

	logger.Printf("Authorized on account %s", bot.Self.UserName)

	// Set up updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Println("Shutting down...")
		cancel()
	}()

	logger.Println("Bot started. Press Ctrl+C to stop.")

	// Process updates
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			go mockBot.handleUpdate(update)
		}
	}
}

// handleUpdate handles incoming updates
func (b *MockBot) handleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		b.handleMessage(update.Message)
	} else if update.CallbackQuery != nil {
		b.handleCallback(update.CallbackQuery)
	}
}

// handleMessage handles text messages
func (b *MockBot) handleMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	text := message.Text

	b.logger.Printf("Received message from %s: %s", message.From.UserName, text)

	if message.IsCommand() {
		switch message.Command() {
		case "start":
			b.handleStart(chatID)
		case "menu":
			b.handleMenu(chatID)
		case "coffee":
			b.handleCoffeeOrder(chatID)
		case "help":
			b.handleHelp(chatID)
		default:
			b.sendMessage(chatID, "❓ Невідома команда. Використовуйте /help для допомоги.")
		}
	} else {
		// Handle regular messages based on context
		b.handleRegularMessage(chatID, text)
	}
}

// handleStart handles /start command
func (b *MockBot) handleStart(chatID int64) {
	welcomeText := `🚀 *Ласкаво просимо до Web3 Coffee Bot!*

Я ваш персональний асистент для замовлення кави з оплатою криптовалютами! ☕️💎

*Що я можу:*
• ☕️ Замовлення кави з меню
• 💳 Оплата Bitcoin, Ethereum та іншими криптовалютами
• 🤖 Розумні рекомендації на основі ШІ
• 📊 Управління замовленнями

*Основні команди:*
/coffee - Замовити каву
/menu - Переглянути меню
/help - Допомога

Почнемо! 👇`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("☕️ Замовити каву", "start_order"),
			tgbotapi.NewInlineKeyboardButtonData("📋 Переглянути меню", "show_menu"),
		),
	)

	b.sendMessageWithKeyboard(chatID, welcomeText, keyboard)
}

// handleMenu handles /menu command
func (b *MockBot) handleMenu(chatID int64) {
	menuText := `📋 *Меню кав'ярні*

*☕️ Еспресо напої:*
• Еспресо - $3.50
• Американо - $4.00
• Латте - $5.50
• Капучино - $5.00
• Мокка - $6.00

*💳 Способи оплати:*
• Bitcoin (BTC)
• Ethereum (ETH)
• USDC
• USDT
• Solana (SOL)

Всі ціни вказані в USD. Курс криптовалют розраховується автоматично! 💎`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("☕️ Замовити", "start_order"),
		),
	)

	b.sendMessageWithKeyboard(chatID, menuText, keyboard)
}

// handleCoffeeOrder handles coffee ordering
func (b *MockBot) handleCoffeeOrder(chatID int64) {
	orderText := `☕️ *Замовлення кави*

Оберіть напій з нашого меню або опишіть, що ви хочете:

*Приклади:*
• "Хочу латте"
• "Замовити капучино"
• "Що ви рекомендуєте?"

Просто напишіть мені! 🤖`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("☕️ Еспресо", "order_espresso"),
			tgbotapi.NewInlineKeyboardButtonData("☕️ Американо", "order_americano"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🥛 Латте", "order_latte"),
			tgbotapi.NewInlineKeyboardButtonData("🥛 Капучино", "order_cappuccino"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🍫 Мокка", "order_mocha"),
		),
	)

	b.sendMessageWithKeyboard(chatID, orderText, keyboard)
}

// handleHelp handles /help command
func (b *MockBot) handleHelp(chatID int64) {
	helpText := `🆘 *Довідка Web3 Coffee Bot*

*Основні команди:*
/start - Почати роботу
/menu - Переглянути меню
/coffee - Замовити каву
/help - Показати довідку

*Як замовити каву:*
1. Використайте /coffee або натисніть кнопку
2. Оберіть напій з меню
3. Підтвердіть замовлення
4. Оплатіть криптовалютою
5. Отримайте каву! ☕️

*Підтримувані криптовалюти:*
• Bitcoin (BTC)
• Ethereum (ETH)
• USDC
• USDT
• Solana (SOL)

Якщо у вас є питання, просто напишіть мені! 🤖`

	b.sendMessage(chatID, helpText)
}

// handleRegularMessage handles regular text messages
func (b *MockBot) handleRegularMessage(chatID int64, text string) {
	// Simple AI-like response for coffee orders
	response := `🤖 Я розумію, що ви хочете замовити каву! 

Використайте команду /coffee або натисніть кнопку нижче, щоб почати замовлення.`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("☕️ Замовити каву", "start_order"),
		),
	)

	b.sendMessageWithKeyboard(chatID, response, keyboard)
}

// handleCallback handles callback queries
func (b *MockBot) handleCallback(callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	data := callback.Data

	// Answer callback to remove loading state
	answerCallback := tgbotapi.NewCallback(callback.ID, "")
	b.api.Request(answerCallback)

	switch data {
	case "start_order":
		b.handleCoffeeOrder(chatID)
	case "show_menu":
		b.handleMenu(chatID)
	case "order_espresso", "order_americano", "order_latte", "order_cappuccino", "order_mocha":
		b.handleSpecificOrder(chatID, data, callback.Message.MessageID)
	default:
		b.sendMessage(chatID, "❓ Невідома команда")
	}
}

// handleSpecificOrder handles specific coffee orders
func (b *MockBot) handleSpecificOrder(chatID int64, orderData string, messageID int) {
	coffeeName := orderData[6:] // Remove "order_" prefix
	price, exists := coffeeMenu[coffeeName]
	if !exists {
		price = 5.00 // Default price
	}

	confirmText := fmt.Sprintf(`✅ *Підтвердження замовлення*

*Ваше замовлення:*
• %s
• Ціна: $%.2f USD

Підтвердити замовлення?`, coffeeName, price)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Підтвердити", fmt.Sprintf("confirm_%s_%.2f", coffeeName, price)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Скасувати", "cancel_order"),
		),
	)

	// Edit the original message
	editMsg := tgbotapi.NewEditMessageText(chatID, messageID, confirmText)
	editMsg.ParseMode = tgbotapi.ModeMarkdown
	editMsg.ReplyMarkup = &keyboard
	b.api.Send(editMsg)
}

// sendMessage sends a simple text message
func (b *MockBot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	b.api.Send(msg)
}

// sendMessageWithKeyboard sends a message with inline keyboard
func (b *MockBot) sendMessageWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}
