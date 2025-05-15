```mermaid
graph TD
    subgraph "Producer Service"
        ProducerMain[Main]
        ProducerConfig[Config]
        ProducerHandler[Handler]
        ProducerKafka[Kafka]
        ProducerMiddleware[Middleware]
        
        ProducerMain --> ProducerConfig
        ProducerMain --> ProducerHandler
        ProducerMain --> ProducerKafka
        ProducerMain --> ProducerMiddleware
        ProducerHandler --> ProducerKafka
    end
    
    subgraph "Consumer Service"
        ConsumerMain[Main]
        ConsumerConfig[Config]
        ConsumerKafka[Kafka]
        
        ConsumerMain --> ConsumerConfig
        ConsumerMain --> ConsumerKafka
    end
    
    subgraph "Middleware"
        RecoverMiddleware[Recover Middleware]
        LoggingMiddleware[Logging Middleware]
        RequestIDMiddleware[Request ID Middleware]
        CORSMiddleware[CORS Middleware]
        
        ProducerMiddleware --> RecoverMiddleware
        ProducerMiddleware --> LoggingMiddleware
        ProducerMiddleware --> RequestIDMiddleware
        ProducerMiddleware --> CORSMiddleware
    end
    
    subgraph "Kafka"
        KafkaBroker[Kafka Broker]
        KafkaTopic[Coffee Orders Topic]
        
        KafkaBroker --> KafkaTopic
        ProducerKafka --> KafkaBroker
        ConsumerKafka --> KafkaBroker
    end
```

Ця діаграма компонентів показує структуру системи замовлення кави та взаємозв'язки між компонентами:

**Producer Service**:
- Main: Точка входу в додаток, ініціалізує компоненти та запускає HTTP-сервер
- Config: Управляє конфігурацією додатку з змінних середовища та файлів конфігурації
- Handler: Містить HTTP-обробники для обробки запитів
- Kafka: Надає абстракцію для взаємодії з Kafka
- Middleware: Містить HTTP-middleware для логування, генерації ID запиту, підтримки CORS та обробки помилок

**Consumer Service**:
- Main: Точка входу в додаток, ініціалізує компоненти та запускає споживача
- Config: Управляє конфігурацією додатку з змінних середовища та файлів конфігурації
- Kafka: Надає абстракцію для взаємодії з Kafka

**Middleware**:
- Recover Middleware: Перехоплює паніки та повертає помилку 500
- Logging Middleware: Логує деталі запиту
- Request ID Middleware: Призначає унікальний ID для запиту
- CORS Middleware: Додає CORS-заголовки до відповіді

**Kafka**:
- Kafka Broker: Сервер Kafka
- Coffee Orders Topic: Тема Kafka для замовлень кави
