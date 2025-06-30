package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-coffee-ai-agents/beverage-inventor-agent/internal/application/usecases"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/infrastructure/config"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/infrastructure/factory"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/infrastructure/kafka"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/infrastructure/logger"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/interfaces/handlers"
)

func main() {
	// Initialize logger
	appLogger := logger.New()

	// Load configuration
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Kafka producer
	kafkaProducer, err := kafka.NewProducer(cfg.Kafka)
	if err != nil {
		log.Fatalf("Failed to initialize Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// Initialize repositories (mock implementations for now)
	// In a real implementation, these would be actual database/external service implementations

	// Create event publisher adapter
	eventPublisher := kafka.NewEventPublisherAdapter(kafkaProducer)

	// Create enhanced services factory
	servicesFactory := factory.NewEnhancedServicesFactory(appLogger)

	// Create enhanced services (with nil AI manager for now)
	enhancedServices := servicesFactory.CreateEnhancedServices(nil)

	// Initialize use cases with enhanced services
	beverageUseCase := usecases.NewBeverageInventorUseCase(
		nil,                                    // beverageRepo - will be implemented later
		eventPublisher,                         // eventPublisher - using Kafka adapter
		nil,                                    // aiProvider - will be implemented later
		nil,                                    // taskManager - will be implemented later
		nil,                                    // notificationSvc - will be implemented later
		enhancedServices.NutritionalAnalyzer,   // nutritionalAnalyzer - enhanced service
		enhancedServices.CostCalculator,        // costCalculator - enhanced service
		enhancedServices.CompatibilityAnalyzer, // compatibilityAnalyzer - enhanced service
		enhancedServices.RecipeOptimizer,       // recipeOptimizer - enhanced service
		appLogger,
	)

	// Initialize handlers
	kafkaHandler := handlers.NewKafkaHandler(beverageUseCase, appLogger)

	// Start Kafka consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := kafkaHandler.StartConsumer(ctx, cfg.Kafka); err != nil {
			log.Fatalf("Failed to start Kafka consumer: %v", err)
		}
	}()

	appLogger.Info("Enhanced Beverage Inventor Agent started successfully",
		"version", "2.0.0",
		"kafka_brokers", cfg.Kafka.Brokers,
		"enhanced_services", "nutritional_analysis,cost_calculation,compatibility_analysis,recipe_optimization",
		"ai_capabilities", "flavor_harmony,taste_prediction,market_analysis")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	appLogger.Info("Shutting down Beverage Inventor Agent...")
	cancel()
}
