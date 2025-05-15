# Налаштування продуктивності Kafka

Цей документ містить рекомендації щодо налаштування продуктивності Kafka для системи замовлення кави.

## Налаштування брокера Kafka

### Налаштування пам'яті

```properties
# Розмір купи Java для Kafka
KAFKA_HEAP_OPTS="-Xmx4g -Xms4g"

# Розмір буфера сторінок
log.flush.interval.messages=10000
log.flush.interval.ms=1000
```

### Налаштування диска

```properties
# Налаштування журналу
log.retention.hours=168
log.segment.bytes=1073741824
log.retention.check.interval.ms=300000

# Налаштування компресії
compression.type=producer
```

### Налаштування мережі

```properties
# Налаштування мережі
num.network.threads=3
num.io.threads=8
socket.send.buffer.bytes=102400
socket.receive.buffer.bytes=102400
socket.request.max.bytes=104857600
```

### Налаштування реплікації

```properties
# Налаштування реплікації
default.replication.factor=3
min.insync.replicas=2
```

## Налаштування Producer

### Налаштування пакетної обробки

```go
// Налаштування пакетної обробки
saramaConfig.Producer.Flush.Bytes = 1024 * 1024 // 1 MB
saramaConfig.Producer.Flush.Messages = 100
saramaConfig.Producer.Flush.Frequency = 500 * time.Millisecond
```

### Налаштування буферизації

```go
// Налаштування буферизації
saramaConfig.Producer.Retry.Max = 5
saramaConfig.Producer.Retry.Backoff = 100 * time.Millisecond
saramaConfig.Producer.Return.Successes = true
saramaConfig.Producer.Return.Errors = true
```

### Налаштування компресії

```go
// Налаштування компресії
saramaConfig.Producer.Compression = sarama.CompressionSnappy
```

## Налаштування Consumer

### Налаштування групи споживачів

```go
// Налаштування групи споживачів
saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
saramaConfig.Consumer.Offsets.AutoCommit.Enable = true
saramaConfig.Consumer.Offsets.AutoCommit.Interval = 5 * time.Second
```

### Налаштування буферизації

```go
// Налаштування буферизації
saramaConfig.Consumer.Fetch.Min = 1
saramaConfig.Consumer.Fetch.Default = 1024 * 1024 // 1 MB
saramaConfig.Consumer.Fetch.Max = 0 // Необмежено
```

### Налаштування пулу воркерів

```go
// Налаштування пулу воркерів
workerPoolSize := runtime.NumCPU() // Кількість воркерів дорівнює кількості ядер CPU
jobQueueSize := 100 // Розмір черги завдань
```

## Налаштування тем

### Налаштування партицій

```bash
# Налаштування партицій
kafka-topics.sh --bootstrap-server localhost:9092 --alter --topic coffee_orders --partitions 3
kafka-topics.sh --bootstrap-server localhost:9092 --alter --topic processed_orders --partitions 3
```

### Налаштування реплікації

```bash
# Налаштування реплікації
kafka-topics.sh --bootstrap-server localhost:9092 --create --topic coffee_orders --partitions 3 --replication-factor 3
kafka-topics.sh --bootstrap-server localhost:9092 --create --topic processed_orders --partitions 3 --replication-factor 3
```

## Моніторинг продуктивності

### Метрики брокера

- Швидкість обробки повідомлень (повідомлень/с)
- Затримка обробки повідомлень (мс)
- Використання диска (%)
- Використання пам'яті (%)
- Використання CPU (%)

### Метрики Producer

- Швидкість відправки повідомлень (повідомлень/с)
- Затримка відправки повідомлень (мс)
- Кількість повторних спроб (повторів/с)
- Розмір пакетів (байт)

### Метрики Consumer

- Швидкість споживання повідомлень (повідомлень/с)
- Затримка споживання повідомлень (мс)
- Відставання споживача (повідомлень)
- Час обробки повідомлень (мс)

## Рекомендації щодо масштабування

### Горизонтальне масштабування

- Додавання нових брокерів Kafka
- Збільшення кількості партицій
- Збільшення кількості споживачів

### Вертикальне масштабування

- Збільшення пам'яті брокерів
- Збільшення кількості потоків вводу-виводу
- Збільшення кількості мережевих потоків

## Висновок

Налаштування продуктивності Kafka є важливим аспектом оптимізації системи замовлення кави. Правильне налаштування брокера, Producer, Consumer та тем дозволяє досягти високої продуктивності та масштабованості системи.
