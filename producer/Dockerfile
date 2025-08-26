FROM golang:1.23-alpine AS builder

# Встановлення необхідних залежностей
RUN apk add --no-cache git

# Встановлення робочої директорії
WORKDIR /app

# Копіювання файлів go.mod та go.sum
COPY go.mod go.sum ./

# Завантаження залежностей
RUN go mod download

# Копіювання коду
COPY . .

# Збірка додатку
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o producer ./main.go

# Використання мінімального образу для запуску
FROM alpine:latest

# Встановлення необхідних пакетів
RUN apk --no-cache add ca-certificates tzdata

# Встановлення часового поясу
ENV TZ=Europe/Kiev

# Створення непривілейованого користувача
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Встановлення робочої директорії
WORKDIR /app

# Копіювання бінарного файлу з попереднього етапу
COPY --from=builder /app/producer .
COPY config.json .

# Встановлення прав доступу
RUN chown -R appuser:appgroup /app

# Перехід на непривілейованого користувача
USER appuser

# Відкриття портів
EXPOSE 3000 9090

# Встановлення змінних середовища за замовчуванням
ENV SERVER_PORT=3000
ENV KAFKA_BROKERS=["kafka:29092"]
ENV KAFKA_TOPIC=coffee_orders
ENV KAFKA_RETRY_MAX=5
ENV KAFKA_REQUIRED_ACKS=all

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Запуск додатку
CMD ["./producer"]
