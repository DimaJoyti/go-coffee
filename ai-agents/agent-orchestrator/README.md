# 🤖 Agent Orchestrator

Розподілений оркестратор AI агентів з підтримкою Redis 8, circuit breaker, rate limiting та event sourcing.

## 🚀 Особливості

### **Основні можливості:**
- ✅ **Реєстрація та управління AI агентами**
- ✅ **Розподілене виконання завдань**
- ✅ **Пошук агентів за можливостями**
- ✅ **Моніторинг здоров'я агентів**
- ✅ **Event sourcing для аудиту**
- ✅ **Circuit breaker для стійкості**
- ✅ **Rate limiting для захисту**
- ✅ **WebSocket для real-time оновлень**

### **Redis 8 Advanced Features:**
- 🔍 **RediSearch** - векторний пошук агентів
- 📊 **RedisTimeSeries** - метрики та аналітика
- 📝 **RedisJSON** - зберігання конфігурацій
- 🌊 **RedisStreams** - event sourcing

### **Distributed Systems Patterns:**
- 🔄 **Circuit Breaker** - захист від cascade failures
- 🚦 **Rate Limiting** - контроль навантаження
- 📚 **Event Sourcing** - immutable event log
- 🔍 **Service Discovery** - автоматичне знаходження агентів

## 📋 API Endpoints

### **Agent Management**
```
POST   /api/v1/agents/register          # Реєстрація агента
DELETE /api/v1/agents/:id               # Видалення агента
GET    /api/v1/agents/:id               # Інформація про агента
GET    /api/v1/agents                   # Список всіх агентів
POST   /api/v1/agents/:id/heartbeat     # Heartbeat агента
GET    /api/v1/agents/:id/health        # Здоров'я агента
```

### **Task Management**
```
POST   /api/v1/tasks                    # Створення завдання
GET    /api/v1/tasks/:id                # Інформація про завдання
GET    /api/v1/tasks                    # Список завдань
POST   /api/v1/tasks/:id/cancel         # Скасування завдання
```

### **Agent Discovery**
```
GET    /api/v1/discovery/search         # Пошук агентів
GET    /api/v1/discovery/capabilities   # Список можливостей
POST   /api/v1/discovery/match          # Підбір агентів
```

### **Monitoring**
```
GET    /api/v1/monitoring/health        # Здоров'я системи
GET    /api/v1/monitoring/metrics       # Метрики
GET    /api/v1/monitoring/stats         # Статистика
```

### **Event Sourcing**
```
GET    /api/v1/events/stream/:type/:id  # Event stream
POST   /api/v1/events/replay            # Replay подій
```

### **WebSocket**
```
GET    /ws                              # Real-time оновлення
```

## 🛠️ Встановлення та Запуск

### **Локальний запуск:**

```bash
# 1. Клонування репозиторію
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee/cmd/agent-orchestrator

# 2. Встановлення залежностей
go mod tidy

# 3. Запуск Redis
docker run -d --name redis -p 6379:6379 redis:latest

# 4. Налаштування змінних середовища
export REDIS_URL="redis://localhost:6379"
export ORCHESTRATOR_PORT="8095"
export LOG_LEVEL="info"

# 5. Запуск оркестратора
go run .
```

### **Docker запуск:**

```bash
# Збірка образу
docker build -t agent-orchestrator .

# Запуск контейнера
docker run -d \
  --name agent-orchestrator \
  -p 8095:8095 \
  -e REDIS_URL="redis://redis:6379" \
  --link redis:redis \
  agent-orchestrator
```

### **Docker Compose:**

```yaml
version: '3.8'
services:
  redis:
    image: redis:latest
    ports:
      - "6379:6379"
    
  agent-orchestrator:
    build: .
    ports:
      - "8095:8095"
    environment:
      - REDIS_URL=redis://redis:6379
      - LOG_LEVEL=info
    depends_on:
      - redis
```

## 📊 Приклади використання

### **1. Реєстрація агента:**

```bash
curl -X POST http://localhost:8095/api/v1/agents/register \
  -H "Content-Type: application/json" \
  -d '{
    "id": "agent-001",
    "name": "Coffee Inventory Agent",
    "type": "inventory",
    "version": "1.0.0",
    "endpoint": "http://localhost:8080",
    "capabilities": ["inventory_management", "stock_tracking"],
    "metadata": {
      "location": "downtown",
      "max_concurrent_tasks": 10
    }
  }'
```

### **2. Створення завдання:**

```bash
curl -X POST http://localhost:8095/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "type": "inventory_check",
    "priority": 1,
    "data": {
      "shop_id": "downtown",
      "items": ["coffee_beans", "milk"]
    },
    "required_capabilities": ["inventory_management"],
    "timeout": "5m"
  }'
```

### **3. Пошук агентів:**

```bash
curl "http://localhost:8095/api/v1/discovery/search?q=inventory"
```

### **4. Підбір агентів за можливостями:**

```bash
curl -X POST http://localhost:8095/api/v1/discovery/match \
  -H "Content-Type: application/json" \
  -d '{
    "required_capabilities": ["inventory_management"],
    "preferred_type": "inventory",
    "max_agents": 5
  }'
```

### **5. Моніторинг здоров'я:**

```bash
curl http://localhost:8095/api/v1/monitoring/health
```

## 🔧 Конфігурація

Основні параметри конфігурації в `config.yaml`:

```yaml
server:
  port: "8095"
  
redis:
  url: "redis://localhost:6379"
  
agents:
  max_agents: 100
  health_check_interval: "30s"
  
circuit_breaker:
  failure_threshold: 5
  timeout: "30s"
  
rate_limiting:
  default_limit: 1000
  default_window: "1m"
```

## 📈 Метрики та Моніторинг

### **Ключові метрики:**
- `total_agents` - загальна кількість агентів
- `active_agents` - активні агенти
- `healthy_agents` - здорові агенти
- `task_events` - події завдань
- `agent_heartbeat` - heartbeat агентів
- `health_ratio` - коефіцієнт здоров'я

### **Health Check:**
```bash
curl http://localhost:8095/api/v1/monitoring/health
```

### **Статистика:**
```bash
curl http://localhost:8095/api/v1/monitoring/stats
```

## 🔒 Безпека

### **Rate Limiting:**
- Sliding window algorithm
- 1000 requests/minute за замовчуванням
- Адаптивне обмеження на основі навантаження

### **Circuit Breaker:**
- Захист від cascade failures
- Автоматичне відновлення
- Configurable thresholds

### **Authentication (опціонально):**
- JWT токени
- API ключі
- CORS підтримка

## 🚀 Архітектура

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   AI Agents     │    │ Agent           │    │ Redis 8         │
│                 │◄──►│ Orchestrator    │◄──►│ Advanced        │
│ - Inventory     │    │                 │    │ - Search        │
│ - Feedback      │    │ - Registration  │    │ - TimeSeries    │
│ - Scheduler     │    │ - Task Mgmt     │    │ - JSON          │
│ - Notifier      │    │ - Discovery     │    │ - Streams       │
└─────────────────┘    │ - Monitoring    │    └─────────────────┘
                       │ - Event Store   │
                       └─────────────────┘
                              │
                       ┌─────────────────┐
                       │ Distributed     │
                       │ Patterns        │
                       │ - Circuit       │
                       │   Breaker       │
                       │ - Rate Limiter  │
                       │ - Event         │
                       │   Sourcing      │
                       └─────────────────┘
```

## 🧪 Тестування

```bash
# Unit тести
go test ./...

# Integration тести
go test -tags=integration ./...

# Load тести
go test -bench=. ./...
```

## 📝 Логування

Структуроване логування з zap:

```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:00Z",
  "message": "Agent registered successfully",
  "agent_id": "agent-001",
  "agent_name": "Coffee Inventory Agent"
}
```

## 🔄 Event Sourcing

Всі важливі події зберігаються в immutable event log:

- `agent_registered`
- `agent_unregistered`
- `task_created`
- `task_completed`
- `task_failed`
- `task_cancelled`

## 🌐 WebSocket Real-time Updates

Підключення до WebSocket для отримання real-time оновлень:

```javascript
const ws = new WebSocket('ws://localhost:8095/ws');
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Status update:', data);
};
```

## 📚 Документація

- [API Reference](./docs/api.md)
- [Architecture Guide](./docs/architecture.md)
- [Deployment Guide](./docs/deployment.md)
- [Troubleshooting](./docs/troubleshooting.md)

## 🤝 Внесок

1. Fork репозиторій
2. Створіть feature branch
3. Зробіть зміни
4. Додайте тести
5. Створіть Pull Request

## 📄 Ліцензія

MIT License - див. [LICENSE](LICENSE) файл.
