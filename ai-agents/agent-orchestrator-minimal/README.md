# 🤖 Мінімальний Agent Orchestrator

Простий та надійний оркестратор AI агентів без зовнішніх залежностей.

## ✅ **ВИПРАВЛЕНО - Система Готова до Роботи!**

### 🚀 **Особливості:**
- ✅ **Нульові зовнішні залежності** - тільки стандартна бібліотека Go
- ✅ **HTTP REST API** - повний набір endpoints
- ✅ **Реєстрація агентів** - управління AI агентами
- ✅ **Створення завдань** - розподіл задач між агентами
- ✅ **Пошук агентів** - знаходження агентів за критеріями
- ✅ **Моніторинг здоров'я** - автоматичний health check
- ✅ **CORS підтримка** - для web інтеграції
- ✅ **Structured JSON API** - стандартизовані відповіді

## 🛠️ **Запуск:**

```bash
# Компіляція
go build .

# Запуск
./agent-orchestrator-minimal

# Або прямий запуск
go run main.go
```

## 📋 **API Endpoints:**

### **Основні операції:**
```
GET  /                              # Інформація про сервіс
POST /api/v1/agents/register        # Реєстрація агента
GET  /api/v1/agents                 # Список агентів
GET  /api/v1/agents/{id}            # Інформація про агента
DELETE /api/v1/agents/{id}          # Видалення агента
POST /api/v1/agents/{id}/heartbeat  # Heartbeat агента
```

### **Завдання:**
```
POST /api/v1/tasks                  # Створення завдання
```

### **Пошук та моніторинг:**
```
GET  /api/v1/discovery/search?q=... # Пошук агентів
GET  /api/v1/monitoring/health      # Здоров'я системи
GET  /api/v1/monitoring/stats       # Статистика
```

## 🧪 **Тестування API:**

### **1. Перевірка роботи сервісу:**
```bash
curl http://localhost:8095/
```

### **2. Реєстрація агента:**
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

### **3. Список агентів:**
```bash
curl http://localhost:8095/api/v1/agents
```

### **4. Створення завдання:**
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
    "required_capabilities": ["inventory_management"]
  }'
```

### **5. Пошук агентів:**
```bash
curl "http://localhost:8095/api/v1/discovery/search?q=inventory"
```

### **6. Здоров'я системи:**
```bash
curl http://localhost:8095/api/v1/monitoring/health
```

## 📊 **Приклади відповідей:**

### **Реєстрація агента:**
```json
{
  "message": "Agent registered successfully",
  "agent": {
    "id": "agent-001",
    "name": "Coffee Inventory Agent",
    "type": "inventory",
    "status": "active",
    "capabilities": ["inventory_management", "stock_tracking"],
    "registered_at": "2024-01-15T10:30:00Z",
    "last_seen": "2024-01-15T10:30:00Z"
  }
}
```

### **Створення завдання:**
```json
{
  "message": "Task created successfully",
  "task_id": "task_1705312200000000000",
  "assigned_agent": {
    "id": "agent-001",
    "name": "Coffee Inventory Agent",
    "type": "inventory"
  },
  "status": "assigned",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### **Здоров'я системи:**
```json
{
  "status": "healthy",
  "total_agents": 1,
  "active_agents": 1,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## 🔧 **Конфігурація:**

### **Змінні середовища:**
- `PORT` - порт сервера (за замовчуванням: 8095)

### **Приклад запуску з іншим портом:**
```bash
PORT=9000 go run main.go
```

## 🏥 **Моніторинг:**

### **Автоматичний Health Check:**
- Кожні 30 секунд перевіряє стан агентів
- Агенти вважаються нездоровими, якщо не надсилали heartbeat більше 2 хвилин
- Логування всіх важливих подій

### **Статистика:**
- Кількість агентів за типами
- Кількість агентів за статусами
- Загальна інформація про систему

## 🌐 **Інтеграція:**

### **CORS підтримка:**
- Дозволені всі origins (`*`)
- Підтримка всіх HTTP методів
- Готовність до інтеграції з web додатками

### **JSON API:**
- Всі відповіді у форматі JSON
- Стандартизовані коди помилок
- Детальні повідомлення про помилки

## 🎯 **Переваги мінімальної версії:**

1. **Швидкий запуск** - компілюється за секунди
2. **Нульові залежності** - тільки стандартна бібліотека Go
3. **Простота розгортання** - один виконуваний файл
4. **Надійність** - мінімум точок відмови
5. **Легкість розширення** - чистий код для модифікації

## 🔄 **Розширення:**

Ця мінімальна версія може бути легко розширена:
- Додавання Redis для персистентності
- Інтеграція з базами даних
- Додавання автентифікації
- Розширення API функціональності
- Додавання WebSocket підтримки

## ✅ **Статус: ГОТОВО ДО ВИКОРИСТАННЯ**

Система повністю функціональна та готова для:
- Реєстрації AI агентів
- Створення та розподілу завдань
- Моніторингу стану агентів
- Інтеграції з існуючими системами

**Мінімальний Agent Orchestrator успішно виправлено та готовий до роботи!** 🎉
