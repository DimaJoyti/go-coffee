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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o streams .

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
COPY --from=builder /app/streams .
COPY config.json .

# Встановлення прав доступу
RUN chown -R appuser:appgroup /app

# Перехід на непривілейованого користувача
USER appuser

# Встановлення змінних середовища за замовчуванням
ENV KAFKA_BROKERS=["kafka:9092"]
ENV KAFKA_INPUT_TOPIC=coffee_orders
ENV KAFKA_OUTPUT_TOPIC=processed_orders
ENV KAFKA_APPLICATION_ID=coffee-streams-app
ENV KAFKA_AUTO_OFFSET_RESET=earliest
ENV KAFKA_PROCESSING_GUARANTEE=at_least_once

# Запуск додатку
CMD ["./streams"]
