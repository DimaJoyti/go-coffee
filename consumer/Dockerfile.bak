FROM golang:1.22-alpine AS builder

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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o consumer .

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
COPY --from=builder /app/consumer .
COPY config.json .

# Встановлення прав доступу
RUN chown -R appuser:appgroup /app

# Перехід на непривілейованого користувача
USER appuser

# Встановлення змінних середовища за замовчуванням
ENV KAFKA_BROKERS=["kafka:9092"]
ENV KAFKA_TOPIC=coffee_orders
ENV KAFKA_PROCESSED_TOPIC=processed_orders
ENV KAFKA_CONSUMER_GROUP=coffee-consumer-group
ENV KAFKA_WORKER_POOL_SIZE=3

# Запуск додатку
CMD ["./consumer"]
