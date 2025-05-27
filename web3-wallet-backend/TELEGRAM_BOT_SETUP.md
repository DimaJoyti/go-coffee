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

## 🚀 Швидкий Старт

### 1. Створення Telegram Бота

1. Знайдіть **@BotFather** в Telegram
2. Відправте `/newbot`
3. Дотримуйтесь інструкцій
4. Збережіть отриманий токен

### 2. Налаштування Environment Variables

```bash
# Windows PowerShell
$env:TELEGRAM_BOT_TOKEN="your_bot_token_here"
$env:GEMINI_API_KEY="your_gemini_api_key"

# Linux/Mac
export TELEGRAM_BOT_TOKEN="your_bot_token_here"
export GEMINI_API_KEY="your_gemini_api_key"
```

### 3. Запуск через Docker

```bash
# Перейдіть в директорію проекту
cd web3-wallet-backend

# Запустіть скрипт
./scripts/start-telegram-bot.sh start

# Або вручну
cd deployments/telegram-bot
docker-compose up -d
```

### 4. Запуск для розробки

```bash
# Встановіть залежності
go mod tidy

# Запустіть Redis
docker run -d -p 6379:6379 redis:alpine

# Запустіть Ollama (опціонально)
docker run -d -p 11434:11434 ollama/ollama
docker exec -it ollama ollama pull llama3.1

# Запустіть бота
go run cmd/telegram-bot/main.go
```

## 🎯 Основні Команди Бота

| Команда | Функція |
|---------|---------|
| `/start` | Початок роботи та налаштування гаманця |
| `/wallet` | Управління Web3 гаманцем |
| `/balance` | Перевірка балансу криптовалют |
| `/coffee` | Замовлення кави з AI |
| `/menu` | Перегляд меню |
| `/orders` | Історія замовлень |
| `/pay` | Криптоплатежі |
| `/settings` | Налаштування |
| `/help` | Допомога |

## 💬 Приклади Використання

### Замовлення Кави

```
👤 Користувач: "Хочу латте з додатковим молоком"
🤖 Бот: "✅ Зрозумів! Латте Medium з додатковим молоком - $6.00"
     [✅ Підтвердити] [✏️ Змінити] [❌ Скасувати]

👤 Користувач: [Підтвердити]
🤖 Бот: "🎉 Замовлення підтверджено! Оберіть спосіб оплати:"
     [₿ Bitcoin] [Ξ Ethereum] [💵 USDC] [💵 USDT]
```

### Створення Гаманця

```
👤 Користувач: /wallet
🤖 Бот: "🔐 Управління гаманцем"
     [🆕 Створити новий] [📥 Імпортувати]

👤 Користувач: [Створити новий]
🤖 Бот: "✅ Гаманець створено успішно!
     Адреса: 0x1234...
     ⚠️ Збережіть seed фразу: abandon abandon..."
```

## 🏗️ Архітектура

```
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

## 🔧 Конфігурація

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

## 🧪 Тестування

```bash
# Unit тести
go test ./internal/telegram/...
go test ./internal/ai/...

# Інтеграційні тести
go test ./tests/integration/...

# Тестування бота
# 1. Запустіть бота
# 2. Знайдіть його в Telegram
# 3. Відправте /start
# 4. Протестуйте команди
```

## 📊 Моніторинг

### Доступні URL після запуску

- **Grafana**: <http://localhost:3000> (admin/admin)
- **Prometheus**: <http://localhost:9090>
- **Redis**: localhost:6379
- **PostgreSQL**: localhost:5432
- **Ollama**: <http://localhost:11434>

### Логи

```bash
# Перегляд логів бота
docker-compose logs -f telegram-bot

# Або через скрипт
./scripts/start-telegram-bot.sh logs
```

## 🔒 Безпека

### Важливі моменти

1. **Ніколи не діліться** TELEGRAM_BOT_TOKEN
2. **Зберігайте в безпеці** GEMINI_API_KEY
3. **Seed фрази** зберігаються тільки у користувача
4. **Приватні ключі** не зберігаються на сервері

### Рекомендації

- Використовуйте `.env` файли для секретів
- Налаштуйте HTTPS для webhook'ів
- Регулярно оновлюйте залежності
- Моніторьте логи на підозрілу активність

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

### Environment Variables для Production

```bash
TELEGRAM_BOT_TOKEN=your_production_token
GEMINI_API_KEY=your_production_key
TELEGRAM_WEBHOOK_URL=https://yourdomain.com/webhook
DB_HOST=your_production_db
REDIS_HOST=your_production_redis
```

## 🛠️ Розширення

### Додавання нових команд

1. Додайте команду в `config.yaml`
2. Створіть обробник в `handlers.go`
3. Додайте логіку в `bot.go`

### Додавання нових AI провайдерів

1. Створіть клієнт в `internal/ai/`
2. Реалізуйте `AIProviderInterface`
3. Додайте в `service.go`

### Додавання нових криптовалют

1. Оновіть `models/wallet.go`
2. Додайте підтримку в `wallet` сервіс
3. Оновіть UI в `callbacks.go`

## 🆘 Troubleshooting

### Поширені проблеми

1. **Бот не відповідає**
   - Перевірте TELEGRAM_BOT_TOKEN
   - Перевірте інтернет з'єднання
   - Подивіться логи: `docker-compose logs telegram-bot`

2. **AI не працює**
   - Перевірте GEMINI_API_KEY
   - Переконайтеся що Ollama запущений
   - Перевірте квоти API

3. **База даних недоступна**
   - Перевірте PostgreSQL: `docker-compose ps`
   - Перевірте підключення: `docker-compose logs postgres`

4. **Redis недоступний**
   - Перевірте Redis: `docker-compose ps redis`
   - Перевірте порт 6379

## 📞 Підтримка

- **GitHub Issues**: [Створити issue](https://github.com/DimaJoyti/go-coffee/issues)
- **Documentation**: [docs/TELEGRAM_BOT.md](docs/TELEGRAM_BOT.md)
- **Email**: <support@web3coffee.com>

## 🎉 Готово

Ваш Web3 Coffee Telegram бот готовий до використання!

**Наступні кроки:**

1. Запустіть бота: `./scripts/start-telegram-bot.sh start`
2. Знайдіть бота в Telegram
3. Відправте `/start`
4. Насолоджуйтесь замовленням кави з криптоплатежами! ☕️💎
