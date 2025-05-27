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

## 📋 Bot Commands

| Command | Description |
|---------|-------------|
| `/start` | Start working with the bot and setup wallet |
| `/wallet` | Manage Web3 wallet |
| `/balance` | Check cryptocurrency balance |
| `/coffee` | Order coffee with AI assistant |
| `/menu` | View complete coffee shop menu |
| `/orders` | My orders and history |
| `/pay` | Make cryptocurrency payment |
| `/settings` | Bot settings and personalization |
| `/help` | Help and support |

## 🛠️ Setup

### 1. Creating Telegram Bot

```bash
# 1. Find @BotFather in Telegram
# 2. Send /newbot
# 3. Follow instructions
# 4. Get your bot token
```

### 2. Environment Variables

```bash
# Required variables
export TELEGRAM_BOT_TOKEN="your_bot_token_here"
export GEMINI_API_KEY="your_gemini_api_key"

# Optional
export TELEGRAM_WEBHOOK_URL="https://yourdomain.com/webhook"
export REDIS_HOST="localhost"
export REDIS_PORT="6379"
export REDIS_PASSWORD=""
```

### 3. Configuration

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

### 4. Running

```bash
# Install dependencies
go mod tidy

# Start Redis
docker run -d -p 6379:6379 redis:alpine

# Start Ollama (optional)
ollama serve
ollama pull llama3.1

# Run the bot
go run cmd/telegram-bot/main.go
```

## 💬 Usage Examples

### Coffee Ordering

```text
User: "I want a latte with extra milk"
Bot: "✅ Got it! Latte Medium with extra milk - $6.00. Confirm?"

User: "Yes"
Bot: "🎉 Order confirmed! Choose payment method:"
[Bitcoin] [Ethereum] [USDC] [USDT]
```

### Wallet Management

```text
User: /wallet
Bot: "🔐 Wallet Management
You don't have a configured wallet yet. Choose an option:"
[🆕 Create New] [📥 Import]

User: [Create New]
Bot: "✅ Wallet created successfully! Address: 0x1234...
⚠️ Save your seed phrase: abandon abandon..."
```

### Balance Check

```text
User: /balance
Bot: "💰 Wallet Balance
• Bitcoin (BTC): 0.00125 (~$52.30)
• Ethereum (ETH): 0.0234 (~$78.45)
• USDC: 125.50
• USDT: 89.20
Total value: ~$345.45 USD"
```

## 🏗️ Architecture

```text
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

## 🔒 Security

### Wallet Security

- **HD Wallets**: Hierarchical deterministic wallets
- **Seed Phrases**: 12-24 words for recovery
- **Encryption**: AES-256 encryption of private keys
- **No Storage**: Private keys are not stored on server

### Bot Security

- **Rate Limiting**: Request rate limiting
- **Input Validation**: Validation of all input data
- **Session Management**: Secure session management
- **Audit Logging**: Logging of all operations

## 🧪 Testing

```bash
# Unit tests
go test ./internal/telegram/...
go test ./internal/ai/...

# Integration tests
go test ./tests/integration/...

# Benchmark tests
go test -bench=. ./internal/telegram/...
```

## 📊 Monitoring

### Metrics

- Number of active users
- AI response success rate
- Bot response time
- Number of orders
- Crypto payment volume

### Logging

```go
// Structured logging with zap
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
