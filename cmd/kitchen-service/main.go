package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/kitchen/application"
	kitchenConfig "github.com/DimaJoyti/go-coffee/internal/kitchen/config"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/infrastructure/ai"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/infrastructure/repository"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/transport"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
)

const (
	serviceName = "kitchen-service"
)

func main() {
	// Initialize logger
	logger := logger.New(serviceName)
	logger.Info("🚀 Starting Kitchen Management Service with Clean Architecture...")

	// Load configuration
	config := loadConfig()

	// Initialize Redis client
	redisClient, err := initializeRedis(config.Database.Redis.URL, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Redis")
	}
	defer redisClient.Close()

	// Initialize dependencies using dependency injection
	dependencies, err := initializeDependencies(redisClient, config, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize dependencies")
	}

	// Initialize and start transport server
	transportServer := transport.NewServer(
		config.Transport,
		logger,
		dependencies.KitchenService,
		dependencies.QueueService,
		dependencies.OptimizerService,
		dependencies.NotificationService,
		dependencies.EventService,
	)

	// Initialize sample data
	go initializeSampleData(redisClient, logger)

	// Run server with graceful shutdown
	if err := transportServer.RunWithGracefulShutdown(); err != nil {
		logger.WithError(err).Fatal("Server failed")
	}

	logger.Info("✅ Kitchen Service stopped gracefully")
}

// Dependencies holds all application dependencies
type Dependencies struct {
	KitchenService      application.KitchenService
	QueueService        application.QueueService
	OptimizerService    application.OptimizerService
	NotificationService application.NotificationService
	EventService        application.EventService
}

// loadConfig loads configuration from environment variables
func loadConfig() *kitchenConfig.Config {
	config, err := kitchenConfig.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	return config
}

// initializeRedis initializes Redis client and tests connection
func initializeRedis(redisURL string, logger *logger.Logger) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	logger.Info("✅ Connected to Redis successfully")
	return client, nil
}

// loggingInterceptor provides request logging for gRPC
func loggingInterceptor(logger *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Log the request
		duration := time.Since(start)
		if err != nil {
			logger.WithFields(map[string]interface{}{
				"method":   info.FullMethod,
				"duration": duration,
			}).WithError(err).Error("gRPC request failed")
		} else {
			logger.WithFields(map[string]interface{}{
				"method":   info.FullMethod,
				"duration": duration,
			}).Info("gRPC request completed")
		}

		return resp, err
	}
}

// initializeSampleKitchenData populates Redis with sample kitchen data
func initializeSampleKitchenData(client *redis.Client, logger *logger.Logger) {
	ctx := context.Background()

	logger.Info("🏪 Initializing sample kitchen data in Redis...")

	// Sample kitchen equipment
	equipment := map[string]map[string]string{
		"kitchen:equipment:espresso-01": {
			"id":               "espresso-01",
			"name":             "Professional Espresso Machine",
			"station_type":     "ESPRESSO",
			"status":           "AVAILABLE",
			"efficiency_score": "9.2",
			"current_load":     "0",
			"max_capacity":     "4",
		},
		"kitchen:equipment:grinder-01": {
			"id":               "grinder-01",
			"name":             "Commercial Coffee Grinder",
			"station_type":     "GRINDER",
			"status":           "AVAILABLE",
			"efficiency_score": "8.8",
			"current_load":     "0",
			"max_capacity":     "2",
		},
		"kitchen:equipment:steamer-01": {
			"id":               "steamer-01",
			"name":             "Milk Steamer",
			"station_type":     "STEAMER",
			"status":           "AVAILABLE",
			"efficiency_score": "9.0",
			"current_load":     "0",
			"max_capacity":     "3",
		},
	}

	// Set equipment data
	for equipmentKey, data := range equipment {
		for field, value := range data {
			if err := client.HSet(ctx, equipmentKey, field, value).Err(); err != nil {
				logger.WithFields(map[string]interface{}{
					"equipment": equipmentKey,
					"field":     field,
				}).WithError(err).Error("Failed to set equipment data")
			}
		}
		logger.WithField("equipment", equipmentKey).Info("✅ Equipment data set")
	}

	// Sample kitchen staff
	staff := map[string]map[string]string{
		"kitchen:staff:barista-01": {
			"id":              "barista-01",
			"name":            "Alice Cooper",
			"specializations": "ESPRESSO,STEAMER",
			"skill_level":     "9.5",
			"is_available":    "true",
			"current_orders":  "0",
		},
		"kitchen:staff:barista-02": {
			"id":              "barista-02",
			"name":            "Bob Wilson",
			"specializations": "GRINDER,ASSEMBLY",
			"skill_level":     "8.7",
			"is_available":    "true",
			"current_orders":  "0",
		},
	}

	// Set staff data
	for staffKey, data := range staff {
		for field, value := range data {
			if err := client.HSet(ctx, staffKey, field, value).Err(); err != nil {
				logger.WithFields(map[string]interface{}{
					"staff": staffKey,
					"field": field,
				}).WithError(err).Error("Failed to set staff data")
			}
		}
		logger.WithField("staff", staffKey).Info("✅ Staff data set")
	}

	// Sample kitchen performance metrics
	metrics := map[string]float64{
		"avg_preparation_time":  4.2,
		"orders_completed":      156,
		"orders_in_queue":       8,
		"efficiency_rate":       92.5,
		"customer_satisfaction": 8.9,
	}

	for metric, value := range metrics {
		if err := client.ZAdd(ctx, "kitchen:metrics:daily", &redis.Z{
			Score:  value,
			Member: metric,
		}).Err(); err != nil {
			logger.WithField("metric", metric).WithError(err).Error("Failed to add kitchen metrics")
		}
	}
	logger.Info("✅ Kitchen metrics data set")

	// Sample AI optimization suggestions
	optimizations := map[string]string{
		"workflow_optimization": "Parallel processing of espresso and milk steaming can reduce preparation time by 15%",
		"staff_allocation":      "Assign Alice to espresso station during peak hours for optimal efficiency",
		"equipment_usage":       "Grinder utilization can be improved by 20% with better scheduling",
		"queue_management":      "Implement priority-based queue ordering to reduce customer wait time",
	}

	for optimization, suggestion := range optimizations {
		if err := client.HSet(ctx, "kitchen:ai:optimizations", optimization, suggestion).Err(); err != nil {
			logger.WithField("optimization", optimization).WithError(err).Error("Failed to set AI optimization")
		}
	}
	logger.Info("✅ AI optimization suggestions set")

	logger.Info("🎉 Sample kitchen data initialization completed successfully!")
}

// initializeDependencies initializes all application dependencies
func initializeDependencies(redisClient *redis.Client, config *kitchenConfig.Config, logger *logger.Logger) (*Dependencies, error) {
	// Initialize repository manager
	repoManager := repository.NewRedisRepositoryManager(redisClient, logger)

	// Initialize AI optimizer service
	optimizerService := ai.NewOptimizerService(logger)

	// Initialize notification service (mock implementation for now)
	notificationService := &MockNotificationService{logger: logger}

	// Initialize event service (mock implementation for now)
	eventService := &MockEventService{logger: logger}

	// Initialize queue service
	queueService := application.NewQueueService(
		repoManager,
		optimizerService,
		eventService,
		logger,
	)

	// Initialize kitchen service
	kitchenService := application.NewKitchenService(
		repoManager,
		queueService,
		optimizerService,
		notificationService,
		eventService,
		logger,
	)

	return &Dependencies{
		KitchenService:      kitchenService,
		QueueService:        queueService,
		OptimizerService:    optimizerService,
		NotificationService: notificationService,
		EventService:        eventService,
	}, nil
}

// initializeSampleData initializes sample data for development
func initializeSampleData(redisClient *redis.Client, logger *logger.Logger) {
	// Use the existing sample data initialization
	initializeSampleKitchenData(redisClient, logger)
}

// Helper functions for environment variables
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrDefaultBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// Mock implementations for development

// MockNotificationService is a mock implementation of NotificationService
type MockNotificationService struct {
	logger *logger.Logger
}

func (m *MockNotificationService) NotifyOrderAdded(ctx context.Context, order *domain.KitchenOrder) error {
	m.logger.WithField("order_id", order.ID()).Info("Mock: Order added notification")
	return nil
}

func (m *MockNotificationService) NotifyOrderStatusChanged(ctx context.Context, order *domain.KitchenOrder, oldStatus domain.OrderStatus) error {
	m.logger.WithFields(map[string]interface{}{
		"order_id":   order.ID(),
		"old_status": oldStatus,
		"new_status": order.Status(),
	}).Info("Mock: Order status changed notification")
	return nil
}

func (m *MockNotificationService) NotifyOrderOverdue(ctx context.Context, order *domain.KitchenOrder) error {
	m.logger.WithField("order_id", order.ID()).Warn("Mock: Order overdue notification")
	return nil
}

func (m *MockNotificationService) NotifyOrderCompleted(ctx context.Context, order *domain.KitchenOrder) error {
	m.logger.WithField("order_id", order.ID()).Info("Mock: Order completed notification")
	return nil
}

func (m *MockNotificationService) NotifyStaffAssigned(ctx context.Context, staff *domain.Staff, order *domain.KitchenOrder) error {
	m.logger.WithFields(map[string]interface{}{
		"staff_id": staff.ID(),
		"order_id": order.ID(),
	}).Info("Mock: Staff assigned notification")
	return nil
}

func (m *MockNotificationService) NotifyStaffOverloaded(ctx context.Context, staff *domain.Staff) error {
	m.logger.WithField("staff_id", staff.ID()).Warn("Mock: Staff overloaded notification")
	return nil
}

func (m *MockNotificationService) NotifyEquipmentMaintenance(ctx context.Context, equipment *domain.Equipment) error {
	m.logger.WithField("equipment_id", equipment.ID()).Warn("Mock: Equipment maintenance notification")
	return nil
}

func (m *MockNotificationService) NotifyEquipmentOverloaded(ctx context.Context, equipment *domain.Equipment) error {
	m.logger.WithField("equipment_id", equipment.ID()).Warn("Mock: Equipment overloaded notification")
	return nil
}

func (m *MockNotificationService) NotifyQueueBacklog(ctx context.Context, queueStatus *domain.QueueStatus) error {
	m.logger.WithField("queue_length", queueStatus.TotalOrders).Warn("Mock: Queue backlog notification")
	return nil
}

func (m *MockNotificationService) NotifyCapacityAlert(ctx context.Context, prediction *application.CapacityPrediction) error {
	m.logger.WithField("capacity_gap", prediction.CapacityGap).Warn("Mock: Capacity alert notification")
	return nil
}

// MockEventService is a mock implementation of EventService
type MockEventService struct {
	logger *logger.Logger
}

func (m *MockEventService) PublishEvent(ctx context.Context, event *domain.DomainEvent) error {
	m.logger.WithFields(map[string]interface{}{
		"event_type":   event.Type,
		"aggregate_id": event.AggregateID,
	}).Info("Mock: Event published")
	return nil
}

func (m *MockEventService) PublishEvents(ctx context.Context, events []*domain.DomainEvent) error {
	m.logger.WithField("event_count", len(events)).Info("Mock: Events published")
	return nil
}

func (m *MockEventService) HandleOrderEvent(ctx context.Context, event *domain.DomainEvent) error {
	m.logger.WithField("event_type", event.Type).Info("Mock: Order event handled")
	return nil
}

func (m *MockEventService) HandleEquipmentEvent(ctx context.Context, event *domain.DomainEvent) error {
	m.logger.WithField("event_type", event.Type).Info("Mock: Equipment event handled")
	return nil
}

func (m *MockEventService) HandleStaffEvent(ctx context.Context, event *domain.DomainEvent) error {
	m.logger.WithField("event_type", event.Type).Info("Mock: Staff event handled")
	return nil
}

func (m *MockEventService) HandleQueueEvent(ctx context.Context, event *domain.DomainEvent) error {
	m.logger.WithField("event_type", event.Type).Info("Mock: Queue event handled")
	return nil
}

func (m *MockEventService) SubscribeToEvents(ctx context.Context, eventTypes []string, handler domain.EventHandler) error {
	m.logger.WithField("event_types", eventTypes).Info("Mock: Subscribed to events")
	return nil
}

func (m *MockEventService) UnsubscribeFromEvents(ctx context.Context, eventTypes []string) error {
	m.logger.WithField("event_types", eventTypes).Info("Mock: Unsubscribed from events")
	return nil
}
