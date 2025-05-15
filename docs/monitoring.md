# Моніторинг системи замовлення кави

Цей документ описує систему моніторингу для системи замовлення кави.

## Огляд

Система моніторингу складається з наступних компонентів:

1. **Prometheus**: Система збору та зберігання метрик.
2. **Grafana**: Система візуалізації метрик.
3. **AlertManager**: Система оповіщення про проблеми.
4. **Node Exporter**: Експортер метрик для хостів.
5. **cAdvisor**: Експортер метрик для контейнерів.
6. **Kafka Exporter**: Експортер метрик для Kafka.

## Архітектура

```
                   ┌─────────────┐
                   │   Grafana   │
                   └──────┬──────┘
                          │
                          ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│AlertManager │◄───┤  Prometheus │────►Node Exporter│
└─────────────┘    └──────┬──────┘    └─────────────┘
                          │
                          ├────────────────┬─────────────────┐
                          ▼                ▼                 ▼
                   ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
                   │  Producer   │  │   Streams   │  │  Consumer   │
                   └─────────────┘  └─────────────┘  └─────────────┘
```

## Метрики

### Producer

| Метрика | Тип | Опис |
|---------|-----|------|
| `coffee_orders_total` | Counter | Загальна кількість отриманих замовлень |
| `coffee_orders_success_total` | Counter | Загальна кількість успішно оброблених замовлень |
| `coffee_orders_failed_total` | Counter | Загальна кількість замовлень, які не вдалося обробити |
| `coffee_order_processing_seconds` | Histogram | Час обробки замовлення |
| `kafka_messages_sent_total` | Counter | Загальна кількість повідомлень, надісланих до Kafka |
| `kafka_messages_failed_total` | Counter | Загальна кількість повідомлень, які не вдалося надіслати до Kafka |
| `http_requests_total` | Counter | Загальна кількість HTTP-запитів |
| `http_request_duration_seconds` | Histogram | Тривалість HTTP-запитів |

### Consumer

| Метрика | Тип | Опис |
|---------|-----|------|
| `coffee_orders_processed_total` | Counter | Загальна кількість оброблених замовлень |
| `coffee_orders_processed_success_total` | Counter | Загальна кількість успішно оброблених замовлень |
| `coffee_orders_processed_failed_total` | Counter | Загальна кількість замовлень, які не вдалося обробити |
| `coffee_order_preparation_seconds` | Histogram | Час приготування кави |
| `kafka_messages_received_total` | Counter | Загальна кількість повідомлень, отриманих з Kafka |
| `kafka_messages_processed_total` | Counter | Загальна кількість оброблених повідомлень з Kafka |
| `kafka_messages_failed_total` | Counter | Загальна кількість повідомлень, які не вдалося обробити з Kafka |
| `worker_pool_queue_size` | Gauge | Поточний розмір черги пулу воркерів |
| `worker_pool_active_workers` | Gauge | Поточна кількість активних воркерів у пулі |
| `worker_processing_seconds` | Histogram | Час обробки повідомлення воркером |

### Streams

| Метрика | Тип | Опис |
|---------|-----|------|
| `streams_messages_processed_total` | Counter | Загальна кількість оброблених повідомлень |
| `streams_messages_success_total` | Counter | Загальна кількість успішно оброблених повідомлень |
| `streams_messages_failed_total` | Counter | Загальна кількість повідомлень, які не вдалося обробити |
| `streams_processing_seconds` | Histogram | Час обробки повідомлення |
| `streams_input_messages_total` | Counter | Загальна кількість вхідних повідомлень |
| `streams_output_messages_total` | Counter | Загальна кількість вихідних повідомлень |
| `streams_errors_total` | Counter | Загальна кількість помилок |
| `streams_running` | Gauge | Чи запущений процесор потоків (1 - запущений, 0 - не запущений) |

## Оповіщення

### Критичні оповіщення

| Оповіщення | Опис | Умова |
|------------|------|-------|
| `ProducerDown` | Producer не працює | `up{job="producer"} == 0` |
| `ConsumerDown` | Consumer не працює | `up{job="consumer"} == 0` |
| `StreamsDown` | Streams не працює | `up{job="streams"} == 0` |
| `StreamsNotRunning` | Streams не запущений | `streams_running == 0` |
| `KafkaDown` | Kafka не працює | `up{job="kafka"} == 0` |

### Попереджувальні оповіщення

| Оповіщення | Опис | Умова |
|------------|------|-------|
| `ProducerHighErrorRate` | Висока частота помилок у Producer | `rate(coffee_orders_failed_total[5m]) / rate(coffee_orders_total[5m]) > 0.05` |
| `ProducerHighLatency` | Висока затримка у Producer | `histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job="producer"}[5m])) by (le, endpoint)) > 1` |
| `ConsumerHighErrorRate` | Висока частота помилок у Consumer | `rate(coffee_orders_processed_failed_total[5m]) / rate(coffee_orders_processed_total[5m]) > 0.05` |
| `ConsumerHighLatency` | Висока затримка у Consumer | `histogram_quantile(0.95, sum(rate(coffee_order_preparation_seconds_bucket{job="consumer"}[5m])) by (le)) > 5` |
| `StreamsHighErrorRate` | Висока частота помилок у Streams | `sum(rate(streams_errors_total[5m])) / sum(rate(streams_input_messages_total[5m])) > 0.05` |
| `KafkaHighLag` | Високий лаг у Kafka | `kafka_consumergroup_lag > 1000` |
| `HighCPUUsage` | Високе використання CPU | `100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80` |
| `HighMemoryUsage` | Високе використання пам'яті | `(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes * 100 > 80` |
| `HighDiskUsage` | Високе використання диска | `100 - ((node_filesystem_avail_bytes / node_filesystem_size_bytes) * 100) > 80` |

## Дашборди

### Coffee System Overview

Дашборд "Coffee System Overview" надає загальний огляд системи замовлення кави. Він містить наступні панелі:

- **Service Status**: Статус сервісів (Producer, Consumer, Streams, Kafka).
- **System Resources**: Використання ресурсів системи (CPU, пам'ять, диск).
- **Order Rate**: Швидкість обробки замовлень.
- **HTTP Request Duration**: Тривалість HTTP-запитів.
- **Kafka Messages**: Кількість повідомлень у Kafka.
- **Worker Pool**: Статус пулу воркерів.
- **Streams Processing**: Статус обробки потоків.

## Налаштування

### Docker Compose

Для запуску системи моніторингу за допомогою Docker Compose використовуйте наступну команду:

```bash
cd monitoring
docker-compose up -d
```

### Kubernetes

Для розгортання системи моніторингу в Kubernetes використовуйте наступну команду:

```bash
kubectl apply -f kubernetes/manifests/monitoring/
```

Або за допомогою Helm:

```bash
helm install coffee-system ./kubernetes/helm/coffee-system \
  --namespace coffee-system \
  --create-namespace \
  --set monitoring.enabled=true
```

## Доступ до інтерфейсів

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (логін: admin, пароль: admin)
- **AlertManager**: http://localhost:9093

## Інтеграція з системами оповіщення

### Slack

Для інтеграції з Slack потрібно налаштувати Webhook URL в конфігурації AlertManager:

```yaml
global:
  slack_api_url: 'https://hooks.slack.com/services/YOUR_SLACK_WEBHOOK_URL'
```

### Email

Для інтеграції з Email потрібно налаштувати SMTP-сервер в конфігурації AlertManager:

```yaml
global:
  smtp_smarthost: 'smtp.example.com:587'
  smtp_from: 'alertmanager@example.com'
  smtp_auth_username: 'alertmanager'
  smtp_auth_password: 'password'
```

## Висновок

Система моніторингу дозволяє відстежувати стан системи замовлення кави та оперативно реагувати на проблеми. Вона надає інформацію про продуктивність системи, використання ресурсів та помилки, що дозволяє покращувати якість обслуговування та запобігати проблемам.
