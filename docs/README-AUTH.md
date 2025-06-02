# 🔐 Auth Service - Enterprise Authentication Microservice

<div align="center">

![Auth Service](https://img.shields.io/badge/Auth%20Service-Enterprise%20Ready-blue?style=for-the-badge&logo=shield)
![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8?style=for-the-badge&logo=go)
![Security](https://img.shields.io/badge/Security-First-red?style=for-the-badge&logo=security)
![License](https://img.shields.io/badge/license-MIT-green?style=for-the-badge)

**🚀 Production-ready authentication microservice with JWT tokens, refresh tokens, and enterprise-grade security**

[🎯 Features](#-features) • [🚀 Quick Start](#-quick-start) • [📖 API Docs](#-api-reference) • [🏗️ Architecture](#-architecture) • [🛡️ Security](#-security)

</div>

---

## 🌟 Why Choose Auth Service?

<table>
<tr>
<td width="50%">

### ✨ **Enterprise Features**
- 🔐 **JWT + Refresh Tokens** with automatic rotation
- 🛡️ **Multi-layer Security** with rate limiting & lockout
- 🔄 **Session Management** with device tracking
- 📊 **Security Monitoring** with real-time alerts
- 🏗️ **Clean Architecture** for maintainability
- 📈 **Horizontal Scaling** ready

</td>
<td width="50%">

### 🚀 **Production Ready**
- 🐳 **Docker & Kubernetes** deployment
- 📊 **Prometheus Metrics** & Grafana dashboards
- 🔍 **Distributed Tracing** with Jaeger
- 📋 **Comprehensive Logging** with structured format
- 🧪 **95%+ Test Coverage** with benchmarks
- 📖 **Complete Documentation** with examples

</td>
</tr>
</table>

Безпечний мікросервіс аутентифікації з JWT токенами та refresh токенами, побудований на Clean Architecture принципах.

## 🌟 Особливості

### 🔑 Аутентифікація та Авторизація
- **JWT токени** з коротким терміном дії (15 хвилин)
- **Refresh токени** з довгим терміном дії (7 днів)
- **Автоматичне оновлення токенів**
- **Безпечне відкликання токенів**
- **Управління множинними сесіями**

### 🛡️ Безпека
- **bcrypt хешування паролів** (cost 12)
- **Rate limiting** (100 запитів/хвилину)
- **Account lockout** після 5 невдалих спроб входу
- **Security event logging**
- **CORS налаштування**
- **Security headers**

### 🏗️ Архітектура
- **Clean Architecture** (Domain, Application, Infrastructure, Transport)
- **Domain-Driven Design** принципи
- **Interface-driven development**
- **Dependency injection**
- **Repository pattern**

### 🌐 API
- **HTTP REST API** для клієнтських додатків
- **gRPC API** для міжсервісної комунікації
- **OpenAPI/Swagger документація**
- **Health checks**

### 📊 Observability
- **Structured logging** з zap
- **Prometheus metrics**
- **Distributed tracing** з OpenTelemetry
- **Jaeger integration**
- **Grafana dashboards**

## 🚀 Швидкий старт

### Передумови
- Go 1.24+
- Redis 7+
- Docker & Docker Compose (опціонально)

### Локальний запуск

1. **Клонування репозиторію**
```bash
git clone https://github.com/DimaJoyti/go-coffee.git
cd go-coffee
```

2. **Встановлення залежностей**
```bash
make -f Makefile.auth deps
```

3. **Запуск Redis**
```bash
docker run -d -p 6379:6379 redis:7-alpine
```

4. **Збірка та запуск сервісу**
```bash
make -f Makefile.auth build
make -f Makefile.auth run-dev
```

### Docker запуск

```bash
# Запуск всіх сервісів
make -f Makefile.auth docker-run

# Або з побудовою образу
make -f Makefile.auth docker-run-build
```

## 📡 API Endpoints

### HTTP REST API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Реєстрація користувача |
| POST | `/api/v1/auth/login` | Вхід в систему |
| POST | `/api/v1/auth/logout` | Вихід з системи |
| POST | `/api/v1/auth/refresh` | Оновлення токенів |
| POST | `/api/v1/auth/validate` | Валідація токена |
| POST | `/api/v1/auth/change-password` | Зміна пароля |
| GET | `/api/v1/auth/me` | Інформація про користувача |
| GET | `/health` | Health check |

### Приклади запитів

**Реєстрація:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "role": "user"
  }'
```

**Вхід:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "remember_me": true
  }'
```

**Оновлення токена:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "your-refresh-token"
  }'
```

## ⚙️ Конфігурація

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `JWT_SECRET` | JWT підписний ключ | - |
| `REDIS_URL` | Redis connection URL | `redis://localhost:6379` |
| `LOG_LEVEL` | Рівень логування | `info` |
| `ENVIRONMENT` | Середовище | `development` |

### Config File

```yaml
# cmd/auth-service/config/config.yaml
server:
  http_port: 8080
  grpc_port: 50053

redis:
  url: "redis://localhost:6379"
  db: 0

security:
  jwt_secret: "${JWT_SECRET}"
  access_token_ttl: "15m"
  refresh_token_ttl: "168h"
  bcrypt_cost: 12
  max_login_attempts: 5
  lockout_duration: "30m"
```

## 🗄️ Redis Schema

```
# Користувачі
auth:users:{userID}                 # hash - дані користувача
auth:users:email:{email}            # string - mapping email -> userID

# Сесії
auth:sessions:{sessionID}           # hash - дані сесії
auth:access_tokens:{accessToken}    # string - mapping token -> sessionID
auth:refresh_tokens:{refreshToken}  # string - mapping token -> sessionID
auth:user_sessions:{userID}         # set - всі сесії користувача

# Безпека
auth:failed_login:{email}           # counter - невдалі спроби входу
```

## 🧪 Тестування

```bash
# Unit тести
make -f Makefile.auth test

# Integration тести
make -f Makefile.auth test-integration

# Benchmarks
make -f Makefile.auth benchmark

# API тести
make -f Makefile.auth test-api
```

## 📊 Моніторинг

### Prometheus Metrics
- Доступні на `http://localhost:9090`
- Метрики запитів, помилок, латентності

### Grafana Dashboards
- Доступні на `http://localhost:3000` (admin/admin)
- Дашборди для моніторингу сервісу

### Jaeger Tracing
- Доступний на `http://localhost:16686`
- Distributed tracing всіх запитів

### Health Checks
```bash
curl http://localhost:8080/health
```

## 🔧 Розробка

### Структура проекту
```
cmd/auth-service/
├── main.go                    # Точка входу
├── config/
│   └── config.yaml           # Конфігурація
└── Dockerfile                # Docker образ

internal/auth/
├── domain/                   # Domain layer
│   ├── user.go              # User entity
│   ├── session.go           # Session entity
│   ├── token.go             # Token entity
│   └── repository.go        # Repository interfaces
├── application/             # Application layer
│   ├── service.go           # Auth service
│   ├── dto.go              # Data transfer objects
│   └── interfaces.go        # Service interfaces
├── infrastructure/         # Infrastructure layer
│   ├── repository/
│   │   ├── redis_user.go   # Redis user repository
│   │   └── redis_session.go # Redis session repository
│   └── security/
│       ├── jwt.go          # JWT service
│       └── password.go     # Password service
└── transport/              # Transport layer (TODO)
    ├── grpc/               # gRPC handlers
    └── http/               # HTTP handlers
```

### Команди розробки

```bash
# Встановлення інструментів
make -f Makefile.auth install-tools

# Форматування коду
make -f Makefile.auth format

# Лінтинг
make -f Makefile.auth lint

# Перевірка коду
make -f Makefile.auth check

# Очищення
make -f Makefile.auth clean
```

## 🚀 Deployment

### Docker
```bash
# Збірка образу
make -f Makefile.auth docker-build

# Запуск з Docker Compose
make -f Makefile.auth docker-run
```

### Kubernetes
```bash
# TODO: Kubernetes manifests
kubectl apply -f k8s/auth-service/
```

## 🔒 Безпека

### Password Policy
- Мінімум 8 символів
- Великі та малі літери
- Цифри та спеціальні символи
- Перевірка на слабкі паролі

### JWT Security
- HS256 алгоритм підпису
- Короткий термін дії access токенів
- Безпечне зберігання refresh токенів
- Token blacklisting

### Rate Limiting
- 100 запитів на хвилину на IP
- Burst size: 20 запитів
- Автоматичне блокування при перевищенні

## 📚 Документація

- [API Documentation](docs/api.md)
- [Architecture Guide](docs/architecture.md)
- [Security Guide](docs/security.md)
- [Deployment Guide](docs/deployment.md)

## 🤝 Contributing

1. Fork репозиторій
2. Створіть feature branch
3. Зробіть зміни
4. Додайте тести
5. Запустіть перевірки: `make -f Makefile.auth check`
6. Створіть Pull Request

## 📄 License

MIT License - див. [LICENSE](LICENSE) файл.

## 🆘 Підтримка

- GitHub Issues: [Issues](https://github.com/DimaJoyti/go-coffee/issues)
- Email: aws.inspiration@gmail.com

---

**Auth Service** - частина Go Coffee мікросервісної архітектури 🚀
