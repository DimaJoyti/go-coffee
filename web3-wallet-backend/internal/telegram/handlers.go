package telegram

import (
	"context"
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleCommand handles bot commands
func (b *Bot) handleCommand(ctx context.Context, message *tgbotapi.Message, session *UserSession) {
	command := message.Command()
	args := message.CommandArguments()

	b.logger.Info(fmt.Sprintf("Handling command /%s for user %d", command, session.UserID))

	switch command {
	case "start":
		b.handleStartCommand(ctx, message, session)
	case "help":
		b.handleHelpCommand(ctx, message, session)
	case "wallet":
		b.handleWalletCommand(ctx, message, session, args)
	case "balance":
		b.handleBalanceCommand(ctx, message, session)
	case "pay":
		b.handlePayCommand(ctx, message, session, args)
	case "coffee":
		b.handleCoffeeCommand(ctx, message, session)
	case "menu":
		b.handleMenuCommand(ctx, message, session)
	case "orders":
		b.handleOrdersCommand(ctx, message, session)
	case "settings":
		b.handleSettingsCommand(ctx, message, session)
	default:
		b.handleUnknownCommand(ctx, message, session)
	}
}

// handleStartCommand handles /start command
func (b *Bot) handleStartCommand(ctx context.Context, message *tgbotapi.Message, session *UserSession) {
	welcomeText := `🚀 *Welcome to Web3 Coffee Bot!*

I'm your personal assistant for ordering coffee with cryptocurrency payments! ☕️💎

*What I can do:*
• 💰 Web3 wallet management
• ☕️ Coffee ordering from menu
• 💳 Payment with Bitcoin, Ethereum and other cryptocurrencies
• 📊 Balance checking and transaction history
• 🤖 Smart AI-powered recommendations

*Main commands:*
/wallet - Wallet management
/coffee - Order coffee
/menu - View menu
/balance - Check balance
/orders - My orders
/help - Help

Let's start by setting up your wallet! 👇`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔐 Create Wallet", "create_wallet"),
			tgbotapi.NewInlineKeyboardButtonData("📥 Import Wallet", "import_wallet"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("☕️ View Menu", "show_menu"),
		),
	)

	b.sendMessageWithKeyboard(session.ChatID, welcomeText, keyboard)
}

// handleHelpCommand handles /help command
func (b *Bot) handleHelpCommand(ctx context.Context, message *tgbotapi.Message, session *UserSession) {
	helpText := `🆘 *Довідка Web3 Coffee Bot*

*Основні команди:*

🔐 */wallet* - Управління гаманцем
   • Створення нового гаманця
   • Імпорт існуючого гаманця
   • Перегляд адреси гаманця

💰 */balance* - Перевірка балансу
   • Баланс всіх криптовалют
   • Поточні курси
   • Історія транзакцій

☕️ */coffee* - Замовлення кави
   • Вибір з меню
   • Персоналізовані рекомендації
   • Розрахунок вартості

💳 */pay* - Здійснення платежу
   • Оплата замовлення
   • Вибір криптовалюти
   • Підтвердження транзакції

📋 */menu* - Перегляд меню
   • Всі види кави
   • Ціни та описи
   • Наявність

📦 */orders* - Мої замовлення
   • Поточні замовлення
   • Історія замовлень
   • Статус доставки

⚙️ */settings* - Налаштування
   • Мова інтерфейсу
   • Улюблені напої
   • Сповіщення

*Підтримка:* Якщо у вас виникли питання, просто напишіть мені повідомлення, і я допоможу! 🤖`

	b.sendMessage(session.ChatID, helpText)
}

// handleWalletCommand handles /wallet command
func (b *Bot) handleWalletCommand(ctx context.Context, message *tgbotapi.Message, session *UserSession, args string) {
	if session.WalletID == "" {
		// No wallet yet, show creation options
		walletText := `🔐 *Управління гаманцем*

У вас ще немає налаштованого гаманця. Оберіть один з варіантів:

*Створити новий гаманець:*
• Автоматично згенерується новий гаманець
• Ви отримаєте seed фразу для відновлення
• Підтримка Ethereum, Bitcoin та інших мереж

*Імпортувати існуючий гаманець:*
• Використайте seed фразу або приватний ключ
• Підключіть існуючий гаманець
• Безпечне зберігання ключів`

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🆕 Створити новий", "create_wallet"),
				tgbotapi.NewInlineKeyboardButtonData("📥 Імпортувати", "import_wallet"),
			),
		)

		b.sendMessageWithKeyboard(session.ChatID, walletText, keyboard)
	} else {
		// Show wallet info
		walletText := `🔐 *Ваш гаманець*

*Статус:* ✅ Активний
*Тип:* HD Wallet
*Мережі:* Ethereum, Bitcoin, Solana

*Доступні дії:*
• Перевірити баланс
• Відправити кошти
• Отримати кошти
• Переглянути історію

Використовуйте команди для управління гаманцем.`

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💰 Баланс", "refresh_balance"),
				tgbotapi.NewInlineKeyboardButtonData("📊 Історія", "transaction_history"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⚙️ Налаштування", "wallet_settings"),
			),
		)

		b.sendMessageWithKeyboard(session.ChatID, walletText, keyboard)
	}
}

// handleBalanceCommand handles /balance command
func (b *Bot) handleBalanceCommand(ctx context.Context, message *tgbotapi.Message, session *UserSession) {
	if session.WalletID == "" {
		b.sendMessage(session.ChatID, "❌ Спочатку налаштуйте гаманець за допомогою команди /wallet")
		return
	}

	// Simulate wallet balance (in real implementation, you'd call the wallet service)
	balanceText := fmt.Sprintf(`💰 *Баланс гаманця*

*Основні валюти:*
• Bitcoin (BTC): 0.00125 (~$52.30)
• Ethereum (ETH): 0.0234 (~$78.45)
• USDC: 125.50
• USDT: 89.20

*Загальна вартість:* ~$345.45 USD

*Адреса гаманця:*
`+"`%s`"+`

Оновлено: %s`,
		session.WalletID[:8]+"...", // Show truncated wallet ID as address
		"щойно",
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Оновити", "refresh_balance"),
			tgbotapi.NewInlineKeyboardButtonData("📊 Історія", "transaction_history"),
		),
	)

	b.sendMessageWithKeyboard(session.ChatID, balanceText, keyboard)
}

// handleCoffeeCommand handles /coffee command
func (b *Bot) handleCoffeeCommand(ctx context.Context, message *tgbotapi.Message, session *UserSession) {
	session.State = StateOrderingCoffee

	coffeeText := `☕️ *Замовлення кави*

Розкажіть мені, яку каву ви хочете замовити! Я можу допомогти вам вибрати з нашого меню або створити щось особливе.

*Приклади:*
• "Хочу латте з додатковим молоком"
• "Що ви рекомендуєте для ранку?"
• "Покажіть меню еспресо"
• "Замовити капучино середнього розміру"

Просто напишіть, що ви хочете, і я допоможу! 🤖`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Показати меню", "show_menu"),
			tgbotapi.NewInlineKeyboardButtonData("⭐ Рекомендації", "coffee_recommendations"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Скасувати", "cancel_order"),
		),
	)

	b.sendMessageWithKeyboard(session.ChatID, coffeeText, keyboard)
}

// handleMenuCommand handles /menu command
func (b *Bot) handleMenuCommand(ctx context.Context, message *tgbotapi.Message, session *UserSession) {
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

*Розміри:* Small, Medium, Large (+$0.50/+$1.00)

Всі ціни вказані в USD. Оплата приймається в BTC, ETH, USDC, USDT та інших криптовалютах! 💎`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("☕️ Замовити", "start_order"),
			tgbotapi.NewInlineKeyboardButtonData("⭐ Рекомендації", "coffee_recommendations"),
		),
	)

	b.sendMessageWithKeyboard(session.ChatID, menuText, keyboard)
}

// handleOrdersCommand handles /orders command
func (b *Bot) handleOrdersCommand(ctx context.Context, message *tgbotapi.Message, session *UserSession) {
	// This would integrate with the order service
	ordersText := `📦 *Мої замовлення*

*Поточні замовлення:*
Немає активних замовлень

*Остання історія:*
• 🕐 Сьогодні 14:30 - Латте Medium - $5.50 (Виконано)
• 🕐 Вчора 09:15 - Капучино Large - $6.00 (Виконано)
• 🕐 2 дні тому - Американо Small - $4.00 (Виконано)

*Статистика:*
• Всього замовлень: 15
• Витрачено: $82.50
• Улюблений напій: Латте
• Рівень лояльності: ⭐⭐⭐ Gold`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("☕️ Нове замовлення", "start_order"),
			tgbotapi.NewInlineKeyboardButtonData("📊 Детальна історія", "detailed_history"),
		),
	)

	b.sendMessageWithKeyboard(session.ChatID, ordersText, keyboard)
}

// handleSettingsCommand handles /settings command
func (b *Bot) handleSettingsCommand(ctx context.Context, message *tgbotapi.Message, session *UserSession) {
	settingsText := `⚙️ *Налаштування*

*Поточні налаштування:*
• Мова: 🇺🇦 Українська
• Сповіщення: ✅ Увімкнено
• Улюблена валюта: USDC
• Автоматичні рекомендації: ✅ Увімкнено

*Персоналізація:*
• Улюблені напої: Латте, Капучино
• Розмір за замовчуванням: Medium
• Додатки: Додаткове молоко

*Безпека:*
• 2FA: ❌ Вимкнено (рекомендуємо увімкнути)
• Автоматичне блокування: ✅ 30 хв`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🌐 Мова", "change_language"),
			tgbotapi.NewInlineKeyboardButtonData("🔔 Сповіщення", "notification_settings"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⭐ Улюблені", "favorite_settings"),
			tgbotapi.NewInlineKeyboardButtonData("🔒 Безпека", "security_settings"),
		),
	)

	b.sendMessageWithKeyboard(session.ChatID, settingsText, keyboard)
}

// handlePayCommand handles /pay command
func (b *Bot) handlePayCommand(ctx context.Context, message *tgbotapi.Message, session *UserSession, args string) {
	if session.WalletID == "" {
		b.sendMessage(session.ChatID, "❌ Спочатку налаштуйте гаманець за допомогою команди /wallet")
		return
	}

	// Parse amount if provided
	if args != "" {
		amount, err := strconv.ParseFloat(args, 64)
		if err != nil {
			b.sendMessage(session.ChatID, "❌ Невірний формат суми. Використовуйте: /pay 5.50")
			return
		}

		session.Context["payment_amount"] = amount
		session.State = StateProcessingPayment

		paymentText := fmt.Sprintf(`💳 *Оплата замовлення*

*Сума до оплати:* $%.2f USD

*Доступні способи оплати:*
Оберіть криптовалюту для оплати:`, amount)

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("₿ Bitcoin", "pay_btc"),
				tgbotapi.NewInlineKeyboardButtonData("Ξ Ethereum", "pay_eth"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💵 USDC", "pay_usdc"),
				tgbotapi.NewInlineKeyboardButtonData("💵 USDT", "pay_usdt"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("❌ Скасувати", "cancel_payment"),
			),
		)

		b.sendMessageWithKeyboard(session.ChatID, paymentText, keyboard)
	} else {
		b.sendMessage(session.ChatID, "💳 Для оплати вкажіть суму: /pay 5.50")
	}
}

// handleUnknownCommand handles unknown commands
func (b *Bot) handleUnknownCommand(ctx context.Context, message *tgbotapi.Message, session *UserSession) {
	unknownText := `❓ *Невідома команда*

Я не розумію цю команду. Ось що я можу:

*Основні команди:*
/start - Почати роботу
/help - Показати довідку
/wallet - Управління гаманцем
/balance - Перевірити баланс
/coffee - Замовити каву
/menu - Переглянути меню
/orders - Мої замовлення
/settings - Налаштування

Або просто напишіть мені повідомлення, і я допоможу! 🤖`

	b.sendMessage(session.ChatID, unknownText)
}
