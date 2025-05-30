# 🚀 Web3 Coffee Telegram Bot - Complete Setup Guide

## 📋 Overview

We have successfully created a powerful Telegram bot for Web3 coffee shop with the following capabilities:

### 🧠 AI Integration
- **Gemini AI** - intelligent responses and recommendations
- **Ollama** - local AI processing
- **LangChain Go** - natural language processing

### 💰 Web3 Features
- Crypto wallet management
- Support for BTC, ETH, USDC, USDT
- DeFi integration
- Secure transactions

### ☕ Coffee Ordering
- Natural language for orders
- Personalized recommendations
- Automated payment processing

## 🛠️ Created Files

### Core Bot Files
```text
web3-wallet-backend/
├── internal/
│   ├── ai/
│   │   ├── service.go          # AI service with Gemini/Ollama
│   │   ├── models.go           # AI models and types
│   │   ├── gemini.go           # Gemini client
│   │   ├── ollama.go           # Ollama client
│   │   └── langchain.go        # LangChain client
│   └── telegram/
│       ├── bot.go              # Main bot
│       ├── handlers.go         # Command handlers
│       └── callbacks.go        # Callback handlers
├── cmd/
│   └── telegram-bot/
│       └── main.go             # Entry point
├── config/
│   └── config.yaml             # Configuration (updated)
└── pkg/
    └── models/
        └── wallet.go           # Wallet models (updated)
```

### Deployment Files
```text
deployments/telegram-bot/
├── Dockerfile                  # Docker image
├── docker-compose.yml          # Local development
└── prometheus.yml              # Monitoring

scripts/
└── start-telegram-bot.sh       # Startup script

docs/
└── TELEGRAM_BOT.md             # Documentation
```

## 🚀 Quick Start

### 1. Create Telegram Bot

1. Find **@BotFather** in Telegram
2. Send `/newbot`
3. Follow the instructions
4. Save the received token

### 2. Environment Variables Setup

```bash
# Windows PowerShell
$env:TELEGRAM_BOT_TOKEN="your_bot_token_here"
$env:GEMINI_API_KEY="your_gemini_api_key"

# Linux/Mac
export TELEGRAM_BOT_TOKEN="your_bot_token_here"
export GEMINI_API_KEY="your_gemini_api_key"
```

### 3. Run with Docker

```bash
# Navigate to project directory
cd web3-wallet-backend

# Run startup script
./scripts/start-telegram-bot.sh start

# Or manually
cd deployments/telegram-bot
docker-compose up -d
```

### 4. Development Setup

```bash
# Install dependencies
go mod tidy

# Start Redis
docker run -d -p 6379:6379 redis:alpine

# Start Ollama (optional)
docker run -d -p 11434:11434 ollama/ollama
docker exec -it ollama ollama pull llama3.1

# Run the bot
go run cmd/telegram-bot/main.go
```

## 🎯 Main Bot Commands

| Command | Function |
|---------|----------|
| `/start` | Start working with bot and setup wallet |
| `/wallet` | Manage Web3 wallet |
| `/balance` | Check cryptocurrency balance |
| `/coffee` | Order coffee with AI |
| `/menu` | View menu |
| `/orders` | Order history |
| `/pay` | Crypto payments |
| `/settings` | Settings |
| `/help` | Help |

## 💬 Usage Examples

### Coffee Ordering
```text
👤 User: "I want a latte with extra milk"
🤖 Bot: "✅ Got it! Latte Medium with extra milk - $6.00"
     [✅ Confirm] [✏️ Modify] [❌ Cancel]

👤 User: [Confirm]
🤖 Bot: "🎉 Order confirmed! Choose payment method:"
     [₿ Bitcoin] [Ξ Ethereum] [💵 USDC] [💵 USDT]
```

### Wallet Creation
```text
👤 User: /wallet
🤖 Bot: "🔐 Wallet Management"
     [🆕 Create New] [📥 Import]

👤 User: [Create New]
🤖 Bot: "✅ Wallet created successfully!
     Address: 0x1234...
     ⚠️ Save your seed phrase: abandon abandon..."
```

## 🏗️ Architecture

```text
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Telegram API  │────│   Bot Service    │────│   AI Services   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                ┌───────────────┼───────────────┐
                │               │               │
                ▼               ▼               ▼
        ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
        │ Wallet       │ │ DeFi         │ │ Coffee       │
        │ Service      │ │ Service      │ │ Service      │
        └──────────────┘ └──────────────┘ └──────────────┘
```

## 🔧 Configuration

### config/config.yaml
```yaml
telegram:
  enabled: true
  bot_token: "${TELEGRAM_BOT_TOKEN}"
  debug: true

ai:
  enabled: true
  gemini:
    enabled: true
    api_key: "${GEMINI_API_KEY}"
    model: "gemini-1.5-flash"
  ollama:
    enabled: true
    host: "localhost"
    port: 11434
    model: "llama3.1"
```

## 🧪 Testing

```bash
# Unit tests
go test ./internal/telegram/...
go test ./internal/ai/...

# Integration tests
go test ./tests/integration/...

# Bot testing
# 1. Start the bot
# 2. Find it in Telegram
# 3. Send /start
# 4. Test commands
```

## 📊 Monitoring

### Available URLs after startup:
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Redis**: localhost:6379
- **PostgreSQL**: localhost:5432
- **Ollama**: http://localhost:11434

### Logs
```bash
# View bot logs
docker-compose logs -f telegram-bot

# Or via script
./scripts/start-telegram-bot.sh logs
```

## 🔒 Security

### Important Notes:
1. **Never share** TELEGRAM_BOT_TOKEN
2. **Keep secure** GEMINI_API_KEY
3. **Seed phrases** are stored only by users
4. **Private keys** are not stored on server

### Recommendations:
- Use `.env` files for secrets
- Setup HTTPS for webhooks
- Regularly update dependencies
- Monitor logs for suspicious activity

## 🚀 Production Deployment

### Kubernetes
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: telegram-bot
spec:
  replicas: 3
  template:
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

### Production Environment Variables
```bash
TELEGRAM_BOT_TOKEN=your_production_token
GEMINI_API_KEY=your_production_key
TELEGRAM_WEBHOOK_URL=https://yourdomain.com/webhook
DB_HOST=your_production_db
REDIS_HOST=your_production_redis
```

## 🛠️ Extensions

### Adding New Commands:
1. Add command to `config.yaml`
2. Create handler in `handlers.go`
3. Add logic to `bot.go`

### Adding New AI Providers:
1. Create client in `internal/ai/`
2. Implement `AIProviderInterface`
3. Add to `service.go`

### Adding New Cryptocurrencies:
1. Update `models/wallet.go`
2. Add support in `wallet` service
3. Update UI in `callbacks.go`

## 🆘 Troubleshooting

### Common Issues:

1. **Bot not responding**
   - Check TELEGRAM_BOT_TOKEN
   - Check internet connection
   - View logs: `docker-compose logs telegram-bot`

2. **AI not working**
   - Check GEMINI_API_KEY
   - Ensure Ollama is running
   - Check API quotas

3. **Database unavailable**
   - Check PostgreSQL: `docker-compose ps`
   - Check connection: `docker-compose logs postgres`

4. **Redis unavailable**
   - Check Redis: `docker-compose ps redis`
   - Check port 6379

## 📞 Support

- **GitHub Issues**: [Create issue](https://github.com/DimaJoyti/go-coffee/issues)
- **Documentation**: [docs/TELEGRAM_BOT.md](docs/TELEGRAM_BOT.md)
- **Email**: support@web3coffee.com

## 🎉 Ready!

Your Web3 Coffee Telegram bot is ready to use! 

**Next Steps:**
1. Start the bot: `./scripts/start-telegram-bot.sh start`
2. Find the bot in Telegram
3. Send `/start`
4. Enjoy ordering coffee with crypto payments! ☕️💎
