# Bright Data MCP Integration

Цей документ описує інтеграцію Go Coffee з Bright Data MCP для отримання реальних даних замість mock даних.

## 🎯 Що було замінено

### Backend (Go)
- **Mock дані** → **Реальні дані через Bright Data MCP**
- **Статичні ціни конкурентів** → **Live скрапінг меню Starbucks, Dunkin', Costa**
- **Фейкові новини** → **Реальний пошук новин про каву**
- **Тестові ринкові дані** → **Актуальні дані про ф'ючерси кави**
- **Симуляція соціальних трендів** → **Реальні тренди з соціальних мереж**

### Frontend (React/Next.js)
- **Hardcoded дані в компонентах** → **API виклики до реальних endpoints**
- **Статичні метрики** → **Live оновлення даних**
- **Mock аналітика** → **Реальна аналітика з Bright Data**

## 🚀 Налаштування

### 1. Змінні середовища

Створіть `.env` файл в `web-ui/` директорії:

```bash
# Bright Data MCP Configuration
BRIGHT_DATA_API_TOKEN=your_bright_data_token_here
BRIGHT_DATA_ZONE=your_zone_here
MCP_SERVER_URL=http://localhost:3001
BRIGHT_DATA_RATE_LIMIT=10
BRIGHT_DATA_CACHE_TTL=300

# Backend Configuration
PORT=8090
GIN_MODE=debug

# Frontend Configuration
NEXT_PUBLIC_API_URL=http://localhost:8090
NEXT_PUBLIC_WS_URL=ws://localhost:8090
```

### 2. Запуск MCP сервера

```bash
# Переконайтеся, що Bright Data MCP сервер запущений
# Зазвичай на порту 3001
```

### 3. Запуск додатку

```bash
# Backend
cd web-ui/backend
go run cmd/web-ui-service/main.go

# Frontend
cd web-ui/frontend
npm run dev
```

## 📊 Нові API Endpoints

### Загальні endpoints
- `GET /api/v1/scraping/data` - Всі ринкові дані
- `POST /api/v1/scraping/refresh` - Оновити дані
- `GET /api/v1/scraping/sources` - Джерела даних

### Специфічні endpoints
- `GET /api/v1/scraping/competitors` - Дані конкурентів
- `GET /api/v1/scraping/news` - Новини ринку
- `GET /api/v1/scraping/futures` - Ф'ючерси кави
- `GET /api/v1/scraping/social` - Соціальні тренди
- `GET /api/v1/scraping/stats` - Статистика сесії

### Інтерактивні endpoints
- `POST /api/v1/scraping/url` - Скрапінг URL
- `POST /api/v1/scraping/search` - Пошук в інтернеті

## 🔧 Тестування

### Тест інтеграції
```bash
cd web-ui
go run test-real-bright-data-integration.go
```

### Тест окремих компонентів
```bash
# Тест MCP клієнта
go test ./backend/internal/services -v

# Тест API endpoints
curl http://localhost:8090/api/v1/scraping/data
```

## 📈 Функції

### Rate Limiting
- Автоматичне обмеження запитів (10/хвилину за замовчуванням)
- Конфігурується через `BRIGHT_DATA_RATE_LIMIT`

### Кешування
- Автоматичне кешування відповідей (5 хвилин за замовчуванням)
- Конфігурується через `BRIGHT_DATA_CACHE_TTL`

### Fallback механізм
- При недоступності API повертаються fallback дані
- Graceful degradation без збоїв

### Error Handling
- Детальне логування помилок
- Retry механізми для failed requests
- User-friendly повідомлення про помилки

## 🎨 Frontend компоненти

### Оновлені компоненти
- `BrightDataAnalytics` - Реальна аналітика
- `DashboardOverview` - Live метрики
- `CoffeeOrders` - Актуальні замовлення
- `AIAgents` - Реальний статус агентів

### Нові hooks
- `useMarketData()` - Ринкові дані
- `useCompetitorData()` - Дані конкурентів
- `useMarketNews()` - Новини ринку
- `useSocialTrends()` - Соціальні тренди

## 🔍 Моніторинг

### Логування
- Всі MCP виклики логуються
- Cache hits/misses відстежуються
- Rate limiting events записуються

### Метрики
- Кількість успішних/неуспішних запитів
- Час відповіді API
- Використання кешу
- Статистика rate limiting

## 🚨 Troubleshooting

### Поширені проблеми

1. **MCP сервер недоступний**
   - Перевірте `MCP_SERVER_URL`
   - Переконайтеся, що сервер запущений

2. **Bright Data API помилки**
   - Перевірте `BRIGHT_DATA_API_TOKEN`
   - Перевірте ліміти API

3. **Повільні відповіді**
   - Збільште `BRIGHT_DATA_CACHE_TTL`
   - Зменште `BRIGHT_DATA_RATE_LIMIT`

### Логи
```bash
# Backend логи
tail -f web-ui/backend/logs/app.log

# Frontend логи
# Дивіться в браузері Developer Tools
```

## 🎉 Результат

Тепер Go Coffee отримує реальні дані про:
- ✅ Ціни конкурентів (Starbucks, Dunkin', Costa)
- ✅ Ринкові новини про каву
- ✅ Ф'ючерси кави та commodity ціни
- ✅ Соціальні медіа тренди
- ✅ Актуальну аналітику ринку

Всі mock дані замінені на реальні дані з Bright Data MCP!
