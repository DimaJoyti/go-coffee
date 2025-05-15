# Продуктивність системи замовлення кави

Цей документ описує аспекти продуктивності системи замовлення кави та рекомендації щодо її оптимізації.

## Поточний стан продуктивності

На даний момент система замовлення кави має наступні характеристики продуктивності:

1. **Producer Service**:
   - Обробляє HTTP-запити синхронно
   - Публікує повідомлення в Kafka синхронно
   - Не має кешування
   - Не має пулінгу з'єднань

2. **Consumer Service**:
   - Споживає повідомлення з Kafka синхронно
   - Обробляє повідомлення послідовно
   - Не має паралельної обробки

3. **Kafka**:
   - Використовує одну партицію для теми
   - Не має налаштувань продуктивності

## Вузькі місця

Потенційні вузькі місця в системі:

1. **Синхронна публікація в Kafka**:
   - Producer Service чекає підтвердження від Kafka перед відповіддю клієнту
   - Це може призвести до затримок відповіді, особливо при високому навантаженні

2. **Одна партиція в Kafka**:
   - Обмежує паралельну обробку повідомлень
   - Обмежує пропускну здатність системи

3. **Послідовна обробка повідомлень**:
   - Consumer Service обробляє повідомлення одне за одним
   - Це може призвести до затримок при високому навантаженні

## Рекомендації щодо покращення продуктивності

### Producer Service

1. **Асинхронна публікація в Kafka**:
   - Використовувати асинхронний продюсер Kafka
   - Відповідати клієнту одразу після отримання запиту
   - Обробляти помилки публікації окремо

   ```go
   // Приклад асинхронного продюсера
   func NewAsyncProducer(config *config.Config) (Producer, error) {
       saramaConfig := sarama.NewConfig()
       saramaConfig.Producer.Return.Successes = true
       saramaConfig.Producer.Return.Errors = true
       
       // ...
       
       producer, err := sarama.NewAsyncProducer(config.Kafka.Brokers, saramaConfig)
       if err != nil {
           return nil, err
       }
       
       // Обробка успішних публікацій та помилок
       go func() {
           for {
               select {
               case success := <-producer.Successes():
                   log.Printf("Message published successfully: %v", success)
               case err := <-producer.Errors():
                   log.Printf("Failed to publish message: %v", err)
               }
           }
       }()
       
       return &AsyncProducer{
           producer: producer,
       }, nil
   }
   ```

2. **Пулінг з'єднань**:
   - Використовувати пул з'єднань для Kafka
   - Повторно використовувати з'єднання замість створення нових

   ```go
   // Приклад пулу з'єднань
   type ProducerPool struct {
       producers []Producer
       mutex     sync.Mutex
   }
   
   func NewProducerPool(config *config.Config, size int) (*ProducerPool, error) {
       pool := &ProducerPool{
           producers: make([]Producer, size),
       }
       
       for i := 0; i < size; i++ {
           producer, err := NewProducer(config)
           if err != nil {
               return nil, err
           }
           
           pool.producers[i] = producer
       }
       
       return pool, nil
   }
   
   func (p *ProducerPool) Get() Producer {
       p.mutex.Lock()
       defer p.mutex.Unlock()
       
       // Вибрати продюсера з пулу
       // ...
       
       return producer
   }
   ```

3. **Кешування**:
   - Кешувати часто використовувані дані
   - Використовувати in-memory кеш або Redis

   ```go
   // Приклад кешування
   type Cache interface {
       Get(key string) (interface{}, bool)
       Set(key string, value interface{})
   }
   
   type InMemoryCache struct {
       cache map[string]interface{}
       mutex sync.RWMutex
   }
   
   func NewInMemoryCache() *InMemoryCache {
       return &InMemoryCache{
           cache: make(map[string]interface{}),
       }
   }
   
   func (c *InMemoryCache) Get(key string) (interface{}, bool) {
       c.mutex.RLock()
       defer c.mutex.RUnlock()
       
       value, ok := c.cache[key]
       return value, ok
   }
   
   func (c *InMemoryCache) Set(key string, value interface{}) {
       c.mutex.Lock()
       defer c.mutex.Unlock()
       
       c.cache[key] = value
   }
   ```

### Consumer Service

1. **Паралельна обробка повідомлень**:
   - Використовувати горутини для паралельної обробки повідомлень
   - Обмежити кількість горутин за допомогою worker pool

   ```go
   // Приклад паралельної обробки
   func main() {
       // ...
       
       // Створити канал для повідомлень
       messages := make(chan *sarama.ConsumerMessage, 100)
       
       // Запустити worker pool
       for i := 0; i < 10; i++ {
           go worker(messages)
       }
       
       // Споживати повідомлення з Kafka
       go func() {
           for {
               select {
               case msg := <-consumer.Messages():
                   messages <- msg
               // ...
               }
           }
       }()
       
       // ...
   }
   
   func worker(messages <-chan *sarama.ConsumerMessage) {
       for msg := range messages {
           // Обробити повідомлення
           // ...
       }
   }
   ```

2. **Батчинг**:
   - Обробляти повідомлення пакетами замість по одному
   - Зменшити накладні витрати на обробку

   ```go
   // Приклад батчингу
   func main() {
       // ...
       
       // Створити канал для повідомлень
       messages := make(chan []*sarama.ConsumerMessage, 10)
       
       // Запустити worker pool
       for i := 0; i < 10; i++ {
           go worker(messages)
       }
       
       // Споживати повідомлення з Kafka
       go func() {
           batch := make([]*sarama.ConsumerMessage, 0, 100)
           ticker := time.NewTicker(100 * time.Millisecond)
           
           for {
               select {
               case msg := <-consumer.Messages():
                   batch = append(batch, msg)
                   
                   if len(batch) >= 100 {
                       messages <- batch
                       batch = make([]*sarama.ConsumerMessage, 0, 100)
                   }
               case <-ticker.C:
                   if len(batch) > 0 {
                       messages <- batch
                       batch = make([]*sarama.ConsumerMessage, 0, 100)
                   }
               // ...
               }
           }
       }()
       
       // ...
   }
   
   func worker(messages <-chan []*sarama.ConsumerMessage) {
       for batch := range messages {
           // Обробити пакет повідомлень
           // ...
       }
   }
   ```

### Kafka

1. **Збільшити кількість партицій**:
   - Створити тему з кількома партиціями
   - Дозволити паралельну обробку повідомлень

   ```bash
   # Приклад створення теми з кількома партиціями
   kafka-topics.sh --create --topic coffee_orders --bootstrap-server localhost:9092 --partitions 10 --replication-factor 1
   ```

2. **Налаштувати продуктивність**:
   - Оптимізувати налаштування Kafka для продуктивності
   - Налаштувати розмір пакету, буфера, тощо

   ```go
   // Приклад налаштування продуктивності
   func NewProducer(config *config.Config) (Producer, error) {
       saramaConfig := sarama.NewConfig()
       // ...
       
       // Налаштування продуктивності
       saramaConfig.Producer.Flush.Frequency = 500 * time.Millisecond
       saramaConfig.Producer.Flush.MaxMessages = 100
       saramaConfig.Producer.Flush.Bytes = 1024 * 1024 // 1 MB
       
       // ...
   }
   ```

## Моніторинг продуктивності

Для моніторингу продуктивності системи рекомендується:

1. **Метрики**:
   - Збирати метрики продуктивності
   - Використовувати Prometheus для збору метрик
   - Використовувати Grafana для візуалізації метрик

   ```go
   // Приклад збору метрик
   var (
       requestDuration = prometheus.NewHistogramVec(
           prometheus.HistogramOpts{
               Name:    "http_request_duration_seconds",
               Help:    "Duration of HTTP requests in seconds",
               Buckets: prometheus.DefBuckets,
           },
           []string{"path", "method", "status"},
       )
       
       kafkaPublishDuration = prometheus.NewHistogram(
           prometheus.HistogramOpts{
               Name:    "kafka_publish_duration_seconds",
               Help:    "Duration of Kafka publish operations in seconds",
               Buckets: prometheus.DefBuckets,
           },
       )
   )
   
   func init() {
       prometheus.MustRegister(requestDuration)
       prometheus.MustRegister(kafkaPublishDuration)
   }
   ```

2. **Трасування**:
   - Використовувати OpenTelemetry для трасування запитів
   - Відстежувати час виконання кожного етапу обробки

   ```go
   // Приклад трасування
   func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
       ctx, span := tracer.Start(r.Context(), "PlaceOrder")
       defer span.End()
       
       // ...
       
       // Трасування публікації в Kafka
       _, kafkaSpan := tracer.Start(ctx, "PublishToKafka")
       err = h.kafkaProducer.PushToQueue(h.config.Kafka.Topic, orderInBytes)
       kafkaSpan.End()
       
       // ...
   }
   ```

3. **Логування продуктивності**:
   - Логувати час виконання операцій
   - Аналізувати логи для виявлення вузьких місць

   ```go
   // Приклад логування продуктивності
   func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
       start := time.Now()
       
       // ...
       
       // Логування часу виконання
       log.Printf("PlaceOrder took %s", time.Since(start))
   }
   ```

## План оптимізації

1. **Короткострокові дії**:
   - Збільшити кількість партицій в Kafka
   - Реалізувати паралельну обробку повідомлень в Consumer Service
   - Налаштувати продуктивність Kafka

2. **Середньострокові дії**:
   - Реалізувати асинхронну публікацію в Kafka
   - Реалізувати пулінг з'єднань
   - Впровадити моніторинг продуктивності

3. **Довгострокові дії**:
   - Реалізувати кешування
   - Реалізувати батчинг
   - Оптимізувати на основі даних моніторингу

## Висновок

Оптимізація продуктивності системи замовлення кави вимагає комплексного підходу, що включає оптимізацію Producer Service, Consumer Service та Kafka. Впровадження рекомендацій, описаних у цьому документі, значно покращить продуктивність системи, особливо при високому навантаженні.
