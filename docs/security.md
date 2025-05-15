# Безпека системи замовлення кави

Цей документ описує аспекти безпеки системи замовлення кави та рекомендації щодо її захисту.

## Поточний стан безпеки

На даний момент система замовлення кави має наступні обмеження з точки зору безпеки:

1. **Відсутність автентифікації**: API не вимагає автентифікації, що означає, що будь-хто може розміщувати замовлення.
2. **Відсутність авторизації**: Немає механізму контролю доступу для обмеження того, хто може розміщувати замовлення.
3. **Відсутність шифрування**: Комунікація між клієнтом і сервером, а також між сервером і Kafka не шифрується.
4. **Відсутність валідації вхідних даних**: Хоча система перевіряє, чи є вхідні дані дійсним JSON, вона не виконує детальну валідацію полів.

## Рекомендації щодо покращення безпеки

### Автентифікація та авторизація

1. **Додати автентифікацію**:
   - Реалізувати автентифікацію на основі JWT (JSON Web Tokens)
   - Додати ендпоінт для входу в систему, який видає JWT
   - Вимагати JWT для всіх API-запитів

   ```go
   // Приклад middleware для перевірки JWT
   func JWTMiddleware(next http.Handler) http.Handler {
       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           tokenString := r.Header.Get("Authorization")
           if tokenString == "" {
               http.Error(w, "Unauthorized", http.StatusUnauthorized)
               return
           }
           
           // Перевірити JWT
           // ...
           
           next.ServeHTTP(w, r)
       })
   }
   ```

2. **Додати авторизацію**:
   - Визначити ролі (наприклад, "customer", "barista", "admin")
   - Реалізувати перевірку ролей для різних ендпоінтів

   ```go
   // Приклад middleware для перевірки ролей
   func RoleMiddleware(role string) func(http.Handler) http.Handler {
       return func(next http.Handler) http.Handler {
           return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
               // Отримати роль з JWT
               userRole := getUserRoleFromJWT(r)
               
               if userRole != role {
                   http.Error(w, "Forbidden", http.StatusForbidden)
                   return
               }
               
               next.ServeHTTP(w, r)
           })
       }
   }
   ```

### Шифрування

1. **Додати HTTPS**:
   - Налаштувати TLS для HTTP-сервера
   - Отримати сертифікат від надійного центру сертифікації або використовувати Let's Encrypt

   ```go
   // Приклад налаштування HTTPS
   func main() {
       // ...
       
       serverAddr := ":" + strconv.Itoa(cfg.Server.Port)
       fmt.Printf("Starting server on %s\n", serverAddr)
       log.Fatal(http.ListenAndServeTLS(serverAddr, "cert.pem", "key.pem", handler))
   }
   ```

2. **Налаштувати шифрування для Kafka**:
   - Налаштувати SSL/TLS для з'єднання з Kafka
   - Налаштувати SASL для автентифікації в Kafka

   ```go
   // Приклад налаштування SSL для Kafka
   func NewProducer(config *config.Config) (Producer, error) {
       saramaConfig := sarama.NewConfig()
       // ...
       
       // Налаштувати SSL
       saramaConfig.Net.TLS.Enable = true
       saramaConfig.Net.TLS.Config = &tls.Config{
           // ...
       }
       
       // ...
   }
   ```

### Валідація вхідних даних

1. **Додати детальну валідацію**:
   - Перевіряти довжину полів
   - Перевіряти формат полів
   - Перевіряти допустимі значення

   ```go
   // Приклад валідації
   func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
       // ...
       
       // Валідація
       if len(order.CustomerName) == 0 || len(order.CustomerName) > 100 {
           http.Error(w, "Invalid customer name", http.StatusBadRequest)
           return
       }
       
       if len(order.CoffeeType) == 0 || len(order.CoffeeType) > 50 {
           http.Error(w, "Invalid coffee type", http.StatusBadRequest)
           return
       }
       
       // ...
   }
   ```

2. **Використовувати бібліотеку для валідації**:
   - Використовувати бібліотеку, таку як `go-playground/validator`, для валідації структур

   ```go
   // Приклад використання validator
   type Order struct {
       CustomerName string `json:"customer_name" validate:"required,max=100"`
       CoffeeType   string `json:"coffee_type" validate:"required,max=50"`
   }
   
   func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
       // ...
       
       // Валідація
       validate := validator.New()
       if err := validate.Struct(order); err != nil {
           http.Error(w, err.Error(), http.StatusBadRequest)
           return
       }
       
       // ...
   }
   ```

### Інші рекомендації

1. **Обмеження швидкості**:
   - Реалізувати обмеження швидкості для запобігання DoS-атакам

   ```go
   // Приклад middleware для обмеження швидкості
   func RateLimitMiddleware(next http.Handler) http.Handler {
       limiter := rate.NewLimiter(rate.Limit(10), 30) // 10 запитів в секунду, бакет на 30 запитів
       
       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           if !limiter.Allow() {
               http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
               return
           }
           
           next.ServeHTTP(w, r)
       })
   }
   ```

2. **Логування безпеки**:
   - Логувати всі спроби автентифікації та авторизації
   - Логувати всі підозрілі дії

   ```go
   // Приклад логування безпеки
   func JWTMiddleware(next http.Handler) http.Handler {
       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           tokenString := r.Header.Get("Authorization")
           if tokenString == "" {
               log.Printf("Security: Unauthorized access attempt from %s", r.RemoteAddr)
               http.Error(w, "Unauthorized", http.StatusUnauthorized)
               return
           }
           
           // ...
       })
   }
   ```

3. **Сканування вразливостей**:
   - Регулярно сканувати код на наявність вразливостей
   - Використовувати інструменти, такі як `gosec`

   ```bash
   # Приклад використання gosec
   gosec ./...
   ```

## План впровадження

1. **Короткострокові дії**:
   - Додати валідацію вхідних даних
   - Налаштувати HTTPS
   - Додати обмеження швидкості

2. **Середньострокові дії**:
   - Реалізувати автентифікацію на основі JWT
   - Налаштувати шифрування для Kafka

3. **Довгострокові дії**:
   - Реалізувати авторизацію на основі ролей
   - Впровадити регулярне сканування вразливостей
   - Розробити політику безпеки

## Висновок

Хоча поточна система замовлення кави має обмеження з точки зору безпеки, впровадження рекомендацій, описаних у цьому документі, значно покращить її безпеку. Важливо розглядати безпеку як постійний процес, а не одноразове завдання.
