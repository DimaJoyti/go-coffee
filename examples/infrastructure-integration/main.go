package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/infrastructure"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/config"
	"github.com/DimaJoyti/go-coffee/pkg/infrastructure/events"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/gorilla/mux"
)

// User represents a user entity
type User struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// UserService demonstrates how to use the infrastructure layer
type UserService struct {
	container infrastructure.ContainerInterface
	logger    *logger.Logger
}

// NewUserService creates a new user service
func NewUserService(container infrastructure.ContainerInterface, logger *logger.Logger) *UserService {
	return &UserService{
		container: container,
		logger:    logger,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, name, email string) (*User, error) {
	// Generate user ID
	userID := fmt.Sprintf("user_%d", time.Now().UnixNano())

	user := &User{
		ID:        userID,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
	}

	// Store in database
	db := s.container.GetDatabase()
	if db != nil {
		query := `INSERT INTO users (id, name, email, created_at) VALUES ($1, $2, $3, $4)`
		_, err := db.Exec(ctx, query, user.ID, user.Name, user.Email, user.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to store user in database: %w", err)
		}
	}

	// Cache the user
	cache := s.container.GetCache()
	if cache != nil {
		cacheKey := fmt.Sprintf("user:%s", user.ID)
		if err := cache.Set(ctx, cacheKey, user, time.Hour); err != nil {
			s.logger.WithError(err).Error("Failed to cache user")
		}
	}

	// Publish user created event
	eventPublisher := s.container.GetEventPublisher()
	if eventPublisher != nil {
		event := &events.Event{
			ID:            fmt.Sprintf("event_%d", time.Now().UnixNano()),
			Type:          "user.created",
			Source:        "user-service",
			AggregateID:   user.ID,
			AggregateType: "user",
			Data: map[string]interface{}{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
			},
			Timestamp: time.Now(),
		}

		if err := eventPublisher.Publish(ctx, event); err != nil {
			s.logger.WithError(err).Error("Failed to publish user created event")
		}
	}

	s.logger.InfoWithFields("User created",
		logger.String("user_id", user.ID),
		logger.String("name", user.Name),
		logger.String("email", user.Email))

	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
	// Try cache first
	cache := s.container.GetCache()
	if cache != nil {
		cacheKey := fmt.Sprintf("user:%s", userID)
		var user User
		if err := cache.Get(ctx, cacheKey, &user); err == nil {
			s.logger.Debug("User found in cache", logger.String("user_id", userID))
			return &user, nil
		}
	}

	// Fallback to database
	db := s.container.GetDatabase()
	if db != nil {
		var user User
		query := `SELECT id, name, email, created_at FROM users WHERE id = $1`
		if err := db.Get(ctx, &user, query, userID); err != nil {
			return nil, fmt.Errorf("user not found: %w", err)
		}

		// Cache for future requests
		if cache != nil {
			cacheKey := fmt.Sprintf("user:%s", userID)
			if err := cache.Set(ctx, cacheKey, &user, time.Hour); err != nil {
				s.logger.WithError(err).Error("Failed to cache user")
			}
		}

		return &user, nil
	}

	return nil, fmt.Errorf("no storage backend available")
}

// AuthenticateUser demonstrates JWT usage
func (s *UserService) AuthenticateUser(ctx context.Context, email, password string) (string, error) {
	// In a real implementation, you would verify the password
	// For this example, we'll just generate a token

	jwtService := s.container.GetJWTService()
	if jwtService == nil {
		return "", fmt.Errorf("JWT service not available")
	}

	// Generate token pair
	tokenPair, err := jwtService.GenerateTokenPair(ctx, "user123", email, "user", map[string]interface{}{
		"email": email,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenPair.AccessToken, nil
}

// UserEventHandler handles user-related events
type UserEventHandler struct {
	logger *logger.Logger
}

// Handle processes user events
func (h *UserEventHandler) Handle(ctx context.Context, event *events.Event) error {
	h.logger.InfoWithFields("Processing user event",
		logger.String("event_id", event.ID),
		logger.String("event_type", event.Type),
		logger.String("aggregate_id", event.AggregateID))

	switch event.Type {
	case "user.created":
		return h.handleUserCreated(ctx, event)
	case "user.updated":
		return h.handleUserUpdated(ctx, event)
	default:
		h.logger.WarnWithFields("Unknown event type", logger.String("event_type", event.Type))
	}

	return nil
}

func (h *UserEventHandler) handleUserCreated(ctx context.Context, event *events.Event) error {
	// Process user creation (e.g., send welcome email, update analytics, etc.)
	h.logger.InfoWithFields("User created event processed",
		logger.String("user_id", event.AggregateID))
	return nil
}

func (h *UserEventHandler) handleUserUpdated(ctx context.Context, event *events.Event) error {
	// Process user update
	h.logger.InfoWithFields("User updated event processed",
		logger.String("user_id", event.AggregateID))
	return nil
}

// CanHandle returns true if this handler can process the event type
func (h *UserEventHandler) CanHandle(eventType string) bool {
	return eventType == "user.created" || eventType == "user.updated"
}

// GetHandlerName returns the handler name
func (h *UserEventHandler) GetHandlerName() string {
	return "UserEventHandler"
}

// HTTP Handlers
type HTTPHandlers struct {
	userService *UserService
	logger      *logger.Logger
}

func NewHTTPHandlers(userService *UserService, logger *logger.Logger) *HTTPHandlers {
	return &HTTPHandlers{
		userService: userService,
		logger:      logger,
	}
}

func (h *HTTPHandlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.CreateUser(r.Context(), req.Name, req.Email)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create user")
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *HTTPHandlers) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *HTTPHandlers) AuthenticateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.userService.AuthenticateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.WithError(err).Error("Failed to authenticate user")
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	response := map[string]string{
		"access_token": token,
		"token_type":   "Bearer",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Load configuration
	cfg := config.DefaultInfrastructureConfig()

	// Override with environment variables if needed
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		cfg.Redis.Host = redisHost
	}
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		cfg.Database.Host = dbHost
	}

	// Create logger
	logger := logger.New("infrastructure-example")

	// Create and initialize infrastructure container
	container := infrastructure.NewContainer(cfg, logger)

	ctx := context.Background()
	if err := container.Initialize(ctx); err != nil {
		log.Fatal("Failed to initialize infrastructure:", err)
	}
	defer container.Shutdown(ctx)

	// Create user service
	userService := NewUserService(container, logger)

	// Set up event handler
	eventSubscriber := container.GetEventSubscriber()
	if eventSubscriber != nil {
		handler := &UserEventHandler{logger: logger}
		if err := eventSubscriber.Subscribe(ctx, []string{"user.created", "user.updated"}, handler); err != nil {
			logger.WithError(err).Error("Failed to subscribe to events")
		}
	}

	// Set up HTTP handlers
	httpHandlers := NewHTTPHandlers(userService, logger)

	// Create HTTP router
	router := mux.NewRouter()
	router.HandleFunc("/users", httpHandlers.CreateUser).Methods("POST")
	router.HandleFunc("/users/{id}", httpHandlers.GetUser).Methods("GET")
	router.HandleFunc("/auth/login", httpHandlers.AuthenticateUser).Methods("POST")

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		status, err := container.HealthCheck(r.Context())
		if err != nil {
			http.Error(w, "Health check failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}).Methods("GET")

	// Start HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		logger.Info("Starting HTTP server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTP server failed:", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	logger.Info("🚀 Infrastructure example service is running")
	logger.Info("📊 Health check: http://localhost:8080/health")
	logger.Info("👤 Create user: POST http://localhost:8080/users")
	logger.Info("🔍 Get user: GET http://localhost:8080/users/{id}")
	logger.Info("🔐 Login: POST http://localhost:8080/auth/login")

	<-c
	logger.Info("🛑 Shutting down service...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("HTTP server shutdown error")
	}

	logger.Info("✅ Service stopped gracefully")
}
