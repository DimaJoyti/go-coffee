# 🤖 Web3 Coffee Telegram Bot

Smart Telegram bot for ordering coffee with cryptocurrency payments, using AI for natural language processing.

## 🚀 Features

### 🧠 AI Integration
- **Gemini AI**: Intelligent responses and recommendations
- **Ollama**: Local AI processing for fast responses
- **LangChain Go**: Natural language processing
- **Contextual dialogs**: Bot understands conversation context

### 💰 Web3 & Crypto
- **Multi-chain support**: Ethereum, Bitcoin, Solana
- **Crypto payments**: BTC, ETH, USDC, USDT
- **DeFi integration**: Automatic currency conversion
- **Secure wallets**: HD wallets with seed phrases

### ☕ Coffee Ordering
- **Natural language**: "I want a latte with extra milk"
- **Personalization**: Recommendations based on history
- **Menu**: Complete catalog of drinks and add-ons
- **Tracking**: Real-time order status

## 📋 Команди Бота

| Команда | Опис |
|---------|------|
| `/start` | Почати роботу з ботом та налаштувати гаманець |
| `/wallet` | Управління Web3 гаманцем |
| `/balance` | Перевірити баланс криптовалют |
| `/coffee` | Замовити каву з AI асистентом |
| `/menu` | Переглянути повне меню кав'ярні |
| `/orders` | Мої замовлення та історія |
| `/pay` | Здійснити платіж криптовалютою |
| `/settings` | Налаштування бота та персоналізація |
| `/help` | Довідка та підтримка |

## 🛠️ Налаштування

### 1. Створення Telegram Бота

```bash
# 1. Знайдіть @BotFather в Telegram
# 2. Відправте /newbot
# 3. Дотримуйтесь інструкцій
# 4. Отримайте токен бота
```

### 2. Environment Variables

```bash
# Обов'язкові змінні
export TELEGRAM_BOT_TOKEN="your_bot_token_here"
export GEMINI_API_KEY="your_gemini_api_key"

# Опціональні
export TELEGRAM_WEBHOOK_URL="https://yourdomain.com/webhook"
export REDIS_HOST="localhost"
export REDIS_PORT="6379"
export REDIS_PASSWORD=""
```

### 3. Конфігурація

```yaml
# config/config.yaml
telegram:
  enabled: true
  bot_token: "${TELEGRAM_BOT_TOKEN}"
  webhook_url: "${TELEGRAM_WEBHOOK_URL}"
  debug: true
  timeout: 30

ai:
  enabled: true
  gemini:
    enabled: true
    api_key: "${GEMINI_API_KEY}"
    model: "gemini-1.5-flash"
    temperature: 0.7
  ollama:
    enabled: true
    host: "localhost"
    port: 11434
    model: "llama3.1"
```

### 4. Запуск

```bash
# Встановіть залежності
go mod tidy

# Запустіть Redis
docker run -d -p 6379:6379 redis:alpine

# Запустіть Ollama (опціонально)
ollama serve
ollama pull llama3.1

# Запустіть бота
go run cmd/telegram-bot/main.go
```

## 💬 Приклади Використання

### Замовлення Кави
```
Користувач: "Хочу латте з додатковим молоком"
Бот: "✅ Зрозумів! Латте Medium з додатковим молоком - $6.00. Підтвердити?"

Користувач: "Так"
Бот: "🎉 Замовлення підтверджено! Оберіть спосіб оплати:"
[Bitcoin] [Ethereum] [USDC] [USDT]
```

### Управління Гаманцем
```
Користувач: /wallet
Бот: "🔐 Управління гаманцем
У вас ще немає налаштованого гаманця. Оберіть варіант:"
[🆕 Створити новий] [📥 Імпортувати]

Користувач: [Створити новий]
Бот: "✅ Гаманець створено! Адреса: 0x1234...
⚠️ Збережіть seed фразу: abandon abandon..."
```

### Перевірка Балансу
```
Користувач: /balance
Бот: "💰 Баланс гаманця
• Bitcoin (BTC): 0.00125 (~$52.30)
• Ethereum (ETH): 0.0234 (~$78.45)
• USDC: 125.50
• USDT: 89.20
Загальна вартість: ~$345.45 USD"
```

## 🏗️ Архітектура

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Telegram Bot  │────│   LangChain Go   │────│  Gemini/Ollama  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌───────────────────────▼───────────────────────┐
         │            Bot Service (Go)                   │
         └───────────────────────┬───────────────────────┘
                                 │
    ┌────────────────────────────┼────────────────────────────┐
    │                            │                            │
    ▼                            ▼                            ▼
┌─────────┐              ┌──────────────┐              ┌─────────┐
│ Wallet  │              │ DeFi Service │              │ Coffee  │
│ Service │              │              │              │ Service │
└─────────┘              └──────────────┘              └─────────┘
```

## 🔒 Безпека

### Wallet Security
- **HD Wallets**: Ієрархічні детерміністичні гаманці
- **Seed Phrases**: 12-24 слова для відновлення
- **Encryption**: AES-256 шифрування приватних ключів
- **No Storage**: Приватні ключі не зберігаються на сервері

### Bot Security
- **Rate Limiting**: Обмеження кількості запитів
- **Input Validation**: Перевірка всіх вхідних даних
- **Session Management**: Безпечне управління сесіями
- **Audit Logging**: Логування всіх операцій

## 🧪 Тестування

```bash
# Unit тести
go test ./internal/telegram/...
go test ./internal/ai/...

# Integration тести
go test ./tests/integration/...

# Benchmark тести
go test -bench=. ./internal/telegram/...
```

## 📊 Моніторинг

### Metrics
- Кількість активних користувачів
- Успішність AI відповідей
- Час відповіді бота
- Кількість замовлень
- Обсяг криптоплатежів

### Logging
```go
// Structured logging з zap
logger.Info("Order processed",
    zap.String("user_id", userID),
    zap.String("order_id", orderID),
    zap.Float64("amount_usd", amount))
```

## 🚀 Deployment

### Docker
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o telegram-bot cmd/telegram-bot/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/telegram-bot .
CMD ["./telegram-bot"]
```

### Kubernetes
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: telegram-bot
spec:
  replicas: 3
  selector:
    matchLabels:
      app: telegram-bot
  template:
    metadata:
      labels:
        app: telegram-bot
    spec:
      containers:
      - name: telegram-bot
        image: web3-coffee/telegram-bot:latest
        env:
        - name: TELEGRAM_BOT_TOKEN
          valueFrom:
            secretKeyRef:
              name: telegram-secrets
              key: bot-token
```

## 🤝 Contributing

1. Fork the repository
2. Create feature branch: `git checkout -b feature/telegram-bot-enhancement`
3. Commit changes: `git commit -am 'Add new feature'`
4. Push to branch: `git push origin feature/telegram-bot-enhancement`
5. Submit Pull Request

## 📄 License

MIT License - see [LICENSE](../LICENSE) file for details.

## 🆘 Support

- **Documentation**: [docs/](../docs/)
- **Issues**: [GitHub Issues](https://github.com/DimaJoyti/go-coffee/issues)
- **Telegram**: @web3coffee_support
- **Email**: support@web3coffee.com
