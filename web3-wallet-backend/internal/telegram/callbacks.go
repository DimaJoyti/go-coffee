package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/internal/ai"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/models"
)

// handleCallbackQuery handles callback queries from inline keyboards
func (b *Bot) handleCallbackQuery(ctx context.Context, callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	chatID := callback.Message.Chat.ID
	data := callback.Data

	b.logger.Info(fmt.Sprintf("Received callback from user %d: %s", userID, data))

	// Get user session
	session := b.getOrCreateSession(userID, chatID, callback.From.UserName)
	session.LastActivity = time.Now()

	// Answer callback query to remove loading state
	answerCallback := tgbotapi.NewCallback(callback.ID, "")
	b.api.Request(answerCallback)

	// Handle different callback types
	switch {
	case data == "create_wallet":
		b.handleCreateWalletCallback(ctx, callback, session)
	case data == "import_wallet":
		b.handleImportWalletCallback(ctx, callback, session)
	case data == "show_menu":
		b.handleShowMenuCallback(ctx, callback, session)
	case data == "coffee_recommendations":
		b.handleCoffeeRecommendationsCallback(ctx, callback, session)
	case data == "start_order":
		b.handleStartOrderCallback(ctx, callback, session)
	case data == "confirm_order":
		b.handleConfirmOrderCallback(ctx, callback, session)
	case data == "cancel_order":
		b.handleCancelOrderCallback(ctx, callback, session)
	case data == "modify_order":
		b.handleModifyOrderCallback(ctx, callback, session)
	case strings.HasPrefix(data, "pay_"):
		b.handlePaymentCallback(ctx, callback, session, data)
	case data == "confirm_payment":
		b.handlePaymentConfirmation(ctx, callback, session)
	case data == "cancel_payment":
		b.handleCancelPaymentCallback(ctx, callback, session)
	case data == "check_payment_status":
		b.handlePaymentStatusCheck(ctx, callback, session)
	case data == "copy_address":
		b.handleCopyAddressCallback(ctx, callback, session)
	case data == "show_payment_address":
		b.handleShowPaymentAddressCallback(ctx, callback, session)
	case data == "refresh_balance":
		b.handleRefreshBalanceCallback(ctx, callback, session)
	case data == "transaction_history":
		b.handleTransactionHistoryCallback(ctx, callback, session)
	case data == "change_language":
		b.handleChangeLanguageCallback(ctx, callback, session)
	case data == "notification_settings":
		b.handleNotificationSettingsCallback(ctx, callback, session)
	case data == "favorite_settings":
		b.handleFavoriteSettingsCallback(ctx, callback, session)
	case data == "security_settings":
		b.handleSecuritySettingsCallback(ctx, callback, session)
	default:
		b.handleUnknownCallback(ctx, callback, session)
	}
}

// handleCreateWalletCallback handles wallet creation
func (b *Bot) handleCreateWalletCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	// Create new wallet
	createReq := &models.CreateWalletRequest{
		UserID: fmt.Sprintf("%d", session.UserID),
		Name:   "Telegram Wallet",
		Chain:  models.ChainEthereum,
		Type:   models.WalletTypeGenerated,
	}

	wallet, err := b.walletService.CreateWallet(ctx, createReq)
	if err != nil {
		b.logger.Error(fmt.Sprintf("Failed to create wallet: %v", err))
		b.sendMessage(session.ChatID, "❌ Помилка при створенні гаманця. Спробуйте пізніше.")
		return
	}

	session.WalletID = wallet.Wallet.ID
	session.State = StateIdle

	successText := fmt.Sprintf(`✅ *Гаманець створено успішно!*

*Адреса гаманця:*
`+"`%s`"+`

*Важливо!* 🔐
Збережіть вашу seed фразу в безпечному місці:

`+"`%s`"+`

⚠️ *Увага:* Ніколи не діліться цією фразою з іншими! Вона дає повний доступ до вашого гаманця.

Тепер ви можете:
• Перевірити баланс: /balance
• Замовити каву: /coffee
• Переглянути меню: /menu`,
		wallet.Wallet.Address,
		wallet.Mnemonic,
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💰 Перевірити баланс", "refresh_balance"),
			tgbotapi.NewInlineKeyboardButtonData("☕️ Замовити каву", "start_order"),
		),
	)

	b.sendMessageWithKeyboard(session.ChatID, successText, keyboard)
}

// handleImportWalletCallback handles wallet import
func (b *Bot) handleImportWalletCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	session.State = StateWaitingWallet

	importText := `📥 *Імпорт гаманця*

Для імпорту гаманця надішліть одне з наступного:

*1. Seed фраза (12-24 слова):*
` + "`abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about`" + `

*2. Приватний ключ:*
` + "`0x1234567890abcdef...`" + `

⚠️ *Безпека:*
• Переконайтеся, що ніхто не бачить ваш екран
• Після імпорту повідомлення буде автоматично видалено
• Ніколи не діліться цими даними з іншими

Надішліть ваші дані для імпорту:`

	b.sendMessage(session.ChatID, importText)
}

// handleShowMenuCallback shows the coffee menu
func (b *Bot) handleShowMenuCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	menuText := `📋 *Меню кав'ярні*

*☕️ Еспресо напої:*
• Еспресо - $3.50
• Американо - $4.00
• Латте - $5.50
• Капучино - $5.00
• Макіато - $5.50

*🥛 Молочні напої:*
• Флет Вайт - $5.50
• Мокка - $6.00
• Карамель Макіато - $6.50

*❄️ Холодні напої:*
• Айс Латте - $5.50
• Фрапе - $6.00
• Колд Брю - $4.50

*🍰 Додатки:*
• Додаткове молоко - $0.50
• Сироп (ваніль, карамель) - $0.75
• Додатковий шот еспресо - $1.00

*Розміри:* Small, Medium, Large (+$0.50/+$1.00)`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("☕️ Замовити", "start_order"),
			tgbotapi.NewInlineKeyboardButtonData("⭐ Рекомендації", "coffee_recommendations"),
		),
	)

	// Edit the original message
	editMsg := tgbotapi.NewEditMessageText(session.ChatID, callback.Message.MessageID, menuText)
	editMsg.ParseMode = tgbotapi.ModeMarkdown
	editMsg.ReplyMarkup = &keyboard
	b.api.Send(editMsg)
}

// handleCoffeeRecommendationsCallback provides AI-powered coffee recommendations
func (b *Bot) handleCoffeeRecommendationsCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	// Use AI to generate personalized recommendations
	prompt := `Надай персоналізовані рекомендації кави для користувача. Враховуй:
- Популярні напої
- Час дня
- Сезонні пропозиції
- Різні смакові переваги

Надай 3-4 рекомендації з коротким описом кожної українською мовою.`

	generateReq := &ai.GenerateRequest{
		UserID:      fmt.Sprintf("%d", session.UserID),
		Message:     prompt,
		Context:     "coffee_recommendations",
		Temperature: 0.8,
	}

	response, err := b.aiService.GenerateResponse(ctx, generateReq)
	if err != nil {
		b.logger.Error(fmt.Sprintf("AI recommendations failed: %v", err))
		b.sendMessage(session.ChatID, "❌ Не вдалося отримати рекомендації. Спробуйте пізніше.")
		return
	}

	recommendationsText := fmt.Sprintf(`⭐ *Персональні рекомендації*

%s

Що вас цікавить?`, response.Text)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("☕️ Замовити", "start_order"),
			tgbotapi.NewInlineKeyboardButtonData("📋 Повне меню", "show_menu"),
		),
	)

	// Edit the original message
	editMsg := tgbotapi.NewEditMessageText(session.ChatID, callback.Message.MessageID, recommendationsText)
	editMsg.ParseMode = tgbotapi.ModeMarkdown
	editMsg.ReplyMarkup = &keyboard
	b.api.Send(editMsg)
}

// handleStartOrderCallback starts the coffee ordering process
func (b *Bot) handleStartOrderCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	session.State = StateOrderingCoffee

	orderText := `☕️ *Почнемо замовлення!*

Розкажіть мені, яку каву ви хочете замовити. Я можу допомогти вам вибрати з нашого меню або створити щось особливе.

*Приклади:*
• "Хочу латте з додатковим молоком"
• "Що ви рекомендуєте для ранку?"
• "Замовити капучино середнього розміру"

Просто напишіть, що ви хочете! 🤖`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Показати меню", "show_menu"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Скасувати", "cancel_order"),
		),
	)

	// Edit the original message
	editMsg := tgbotapi.NewEditMessageText(session.ChatID, callback.Message.MessageID, orderText)
	editMsg.ParseMode = tgbotapi.ModeMarkdown
	editMsg.ReplyMarkup = &keyboard
	b.api.Send(editMsg)
}

// handleConfirmOrderCallback confirms the order
func (b *Bot) handleConfirmOrderCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	b.confirmOrder(ctx, session)
}

// handleCancelOrderCallback cancels the order
func (b *Bot) handleCancelOrderCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	session.State = StateIdle
	delete(session.Context, "pending_order")

	cancelText := `❌ *Замовлення скасовано*

Не проблема! Ви можете замовити каву в будь-який час.

Що бажаєте зробити далі?`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("☕️ Нове замовлення", "start_order"),
			tgbotapi.NewInlineKeyboardButtonData("📋 Переглянути меню", "show_menu"),
		),
	)

	// Edit the original message
	editMsg := tgbotapi.NewEditMessageText(session.ChatID, callback.Message.MessageID, cancelText)
	editMsg.ParseMode = tgbotapi.ModeMarkdown
	editMsg.ReplyMarkup = &keyboard
	b.api.Send(editMsg)
}

// handleModifyOrderCallback allows order modification
func (b *Bot) handleModifyOrderCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	session.State = StateOrderingCoffee

	modifyText := `✏️ *Зміна замовлення*

Розкажіть, що ви хочете змінити у вашому замовленні:

• Тип кави
• Розмір
• Додатки
• Кількість

Просто опишіть, що ви хочете! 🤖`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Скасувати зміни", "cancel_order"),
		),
	)

	// Edit the original message
	editMsg := tgbotapi.NewEditMessageText(session.ChatID, callback.Message.MessageID, modifyText)
	editMsg.ParseMode = tgbotapi.ModeMarkdown
	editMsg.ReplyMarkup = &keyboard
	b.api.Send(editMsg)
}

// handlePaymentCallback handles payment method selection
func (b *Bot) handlePaymentCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession, data string) {
	currency := strings.TrimPrefix(data, "pay_")

	// Get pending order
	pendingOrder, exists := session.Context["pending_order"]
	if !exists {
		answerCallback := tgbotapi.NewCallbackWithAlert(callback.ID, "❌ Замовлення не знайдено")
		b.api.Request(answerCallback)
		return
	}

	// Extract amount from order (simplified)
	amount := 5.50 // Default amount, in real implementation get from order
	if order, ok := pendingOrder.(*ai.ParsedCoffeeOrder); ok {
		amount = order.EstimatedPriceUSD
	}

	// Process payment request
	err := b.processPaymentRequest(ctx, session, currency, amount)
	if err != nil {
		b.logger.Error(fmt.Sprintf("Payment processing failed: %v", err))
		answerCallback := tgbotapi.NewCallbackWithAlert(callback.ID, "❌ Помилка обробки платежу")
		b.api.Request(answerCallback)
		return
	}

	session.State = StateProcessingPayment
}

// handleRefreshBalanceCallback refreshes wallet balance
func (b *Bot) handleRefreshBalanceCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	if session.WalletID == "" {
		answerCallback := tgbotapi.NewCallbackWithAlert(callback.ID, "❌ Спочатку налаштуйте гаманець")
		b.api.Request(answerCallback)
		return
	}

	// Simulate balance refresh
	balanceText := `💰 *Баланс оновлено*

*Поточний баланс:*
• Bitcoin (BTC): 0.00125 (~$52.30)
• Ethereum (ETH): 0.0234 (~$78.45)
• USDC: 125.50
• USDT: 89.20

*Загальна вартість:* ~$345.45 USD

🔄 Останнє оновлення: щойно`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Історія", "transaction_history"),
			tgbotapi.NewInlineKeyboardButtonData("☕️ Замовити каву", "start_order"),
		),
	)

	editMsg := tgbotapi.NewEditMessageText(session.ChatID, callback.Message.MessageID, balanceText)
	editMsg.ParseMode = tgbotapi.ModeMarkdown
	editMsg.ReplyMarkup = &keyboard
	b.api.Send(editMsg)
}

// handleTransactionHistoryCallback shows transaction history
func (b *Bot) handleTransactionHistoryCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	historyText := `📊 *Історія транзакцій*

*Останні операції:*

🟢 *Отримано*
• +0.001 BTC
• 2024-01-15 14:30
• Від: 1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa

🔴 *Відправлено*
• -25.50 USDC
• 2024-01-15 12:15
• Кава Латте Medium

🟢 *Отримано*
• +0.005 ETH
• 2024-01-14 18:45
• Від: 0x742d35Cc6634C0532925a3b8D4C0d886E

*Всього транзакцій:* 15
*Загальний обсяг:* $1,234.56`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Оновити баланс", "refresh_balance"),
			tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", "show_wallet"),
		),
	)

	editMsg := tgbotapi.NewEditMessageText(session.ChatID, callback.Message.MessageID, historyText)
	editMsg.ParseMode = tgbotapi.ModeMarkdown
	editMsg.ReplyMarkup = &keyboard
	b.api.Send(editMsg)
}

// handleChangeLanguageCallback handles language change
func (b *Bot) handleChangeLanguageCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	langText := `🌐 *Вибір мови / Language Selection*

Оберіть мову інтерфейсу:
Choose interface language:`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🇺🇦 Українська", "lang_uk"),
			tgbotapi.NewInlineKeyboardButtonData("🇺🇸 English", "lang_en"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", "settings_menu"),
		),
	)

	editMsg := tgbotapi.NewEditMessageText(session.ChatID, callback.Message.MessageID, langText)
	editMsg.ParseMode = tgbotapi.ModeMarkdown
	editMsg.ReplyMarkup = &keyboard
	b.api.Send(editMsg)
}

// handleNotificationSettingsCallback handles notification settings
func (b *Bot) handleNotificationSettingsCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	notifText := `🔔 *Налаштування сповіщень*

*Поточні налаштування:*
• Замовлення: ✅ Увімкнено
• Платежі: ✅ Увімкнено
• Промоції: ❌ Вимкнено
• Новини: ✅ Увімкнено

Оберіть, що змінити:`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📦 Замовлення", "toggle_order_notif"),
			tgbotapi.NewInlineKeyboardButtonData("💳 Платежі", "toggle_payment_notif"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎁 Промоції", "toggle_promo_notif"),
			tgbotapi.NewInlineKeyboardButtonData("📰 Новини", "toggle_news_notif"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", "settings_menu"),
		),
	)

	editMsg := tgbotapi.NewEditMessageText(session.ChatID, callback.Message.MessageID, notifText)
	editMsg.ParseMode = tgbotapi.ModeMarkdown
	editMsg.ReplyMarkup = &keyboard
	b.api.Send(editMsg)
}

// handleFavoriteSettingsCallback handles favorite settings
func (b *Bot) handleFavoriteSettingsCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	favText := `⭐ *Улюблені налаштування*

*Ваші улюблені:*
• Напій: Латте Medium
• Додатки: Додаткове молоко
• Час замовлення: 08:30, 14:00
• Спосіб оплати: USDC

*Статистика:*
• Замовлень латте: 12
• Улюблений час: Ранок
• Середня сума: $5.75`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("☕️ Змінити напій", "change_fav_drink"),
			tgbotapi.NewInlineKeyboardButtonData("💳 Змінити оплату", "change_fav_payment"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", "settings_menu"),
		),
	)

	editMsg := tgbotapi.NewEditMessageText(session.ChatID, callback.Message.MessageID, favText)
	editMsg.ParseMode = tgbotapi.ModeMarkdown
	editMsg.ReplyMarkup = &keyboard
	b.api.Send(editMsg)
}

// handleSecuritySettingsCallback handles security settings
func (b *Bot) handleSecuritySettingsCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	secText := `🔒 *Налаштування безпеки*

*Поточний стан:*
• 2FA: ❌ Вимкнено (рекомендуємо увімкнути)
• Автоблокування: ✅ 30 хвилин
• Підтвердження платежів: ✅ Увімкнено
• Резервне копіювання: ✅ Хмарне

*Рекомендації:*
• Увімкніть 2FA для додаткової безпеки
• Регулярно оновлюйте пароль
• Не діліться seed фразою`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔐 Увімкнути 2FA", "enable_2fa"),
			tgbotapi.NewInlineKeyboardButtonData("⏰ Автоблокування", "auto_lock_settings"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💾 Резервна копія", "backup_settings"),
			tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", "settings_menu"),
		),
	)

	editMsg := tgbotapi.NewEditMessageText(session.ChatID, callback.Message.MessageID, secText)
	editMsg.ParseMode = tgbotapi.ModeMarkdown
	editMsg.ReplyMarkup = &keyboard
	b.api.Send(editMsg)
}

// handleUnknownCallback handles unknown callback queries
func (b *Bot) handleUnknownCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	b.logger.Warn(fmt.Sprintf("Unknown callback data: %s", callback.Data))

	// Answer with error message
	answerCallback := tgbotapi.NewCallbackWithAlert(callback.ID, "❌ Невідома команда")
	b.api.Request(answerCallback)
}

// handleCancelPaymentCallback handles payment cancellation
func (b *Bot) handleCancelPaymentCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	session.State = StateIdle
	delete(session.Context, "payment_request")
	delete(session.Context, "payment_info")

	cancelText := `❌ *Платіж скасовано*

Не проблема! Ви можете оплатити замовлення пізніше.

Що бажаєте зробити далі?`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("☕️ Нове замовлення", "start_order"),
			tgbotapi.NewInlineKeyboardButtonData("📋 Переглянути меню", "show_menu"),
		),
	)

	editMsg := tgbotapi.NewEditMessageText(session.ChatID, callback.Message.MessageID, cancelText)
	editMsg.ParseMode = tgbotapi.ModeMarkdown
	editMsg.ReplyMarkup = &keyboard
	b.api.Send(editMsg)
}

// handleCopyAddressCallback handles address copying
func (b *Bot) handleCopyAddressCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	paymentInfo, exists := session.Context["payment_info"].(*CryptoPaymentInfo)
	if !exists {
		answerCallback := tgbotapi.NewCallbackWithAlert(callback.ID, "❌ Платіжна інформація не знайдена")
		b.api.Request(answerCallback)
		return
	}

	// Show address for easy copying
	addressText := fmt.Sprintf(`📋 *Адреса для копіювання*

*%s адреса:*
`+"`%s`"+`

Натисніть на адресу щоб скопіювати її.

*Важливо:* Відправте точно %s %s на цю адресу.`,
		paymentInfo.Currency,
		paymentInfo.Address,
		paymentInfo.TotalRequired,
		paymentInfo.Currency,
	)

	answerCallback := tgbotapi.NewCallbackWithAlert(callback.ID, "📋 Адреса показана нижче для копіювання")
	b.api.Request(answerCallback)

	b.sendMessage(session.ChatID, addressText)
}

// handleShowPaymentAddressCallback shows payment address again
func (b *Bot) handleShowPaymentAddressCallback(ctx context.Context, callback *tgbotapi.CallbackQuery, session *UserSession) {
	paymentInfo, exists := session.Context["payment_info"].(*CryptoPaymentInfo)
	if !exists {
		answerCallback := tgbotapi.NewCallbackWithAlert(callback.ID, "❌ Платіжна інформація не знайдена")
		b.api.Request(answerCallback)
		return
	}

	paymentReq, exists := session.Context["payment_request"].(*PaymentRequest)
	if !exists {
		answerCallback := tgbotapi.NewCallbackWithAlert(callback.ID, "❌ Платіжна інформація не знайдена")
		b.api.Request(answerCallback)
		return
	}

	// Resend payment instructions
	b.sendPaymentInstructions(session.ChatID, paymentInfo, paymentReq)

	answerCallback := tgbotapi.NewCallback(callback.ID, "💳 Платіжні інструкції відправлені")
	b.api.Request(answerCallback)
}
