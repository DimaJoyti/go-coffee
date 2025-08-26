package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/ai/orchestrator"
	"github.com/DimaJoyti/go-coffee/internal/ai/agents"
	"github.com/DimaJoyti/go-coffee/internal/ai/messaging"
	httpTransport "github.com/DimaJoyti/go-coffee/internal/ai/transport/http"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Initialize structured logger
	logger := initLogger()
	defer logger.Sync()

	logger.Info("Starting AI Agent Orchestrator...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize Kafka messaging
	kafkaManager, err := messaging.NewKafkaManager(cfg.AI.Kafka, logger)
	if err != nil {
		logger.Fatal("Failed to create Kafka manager", zap.Error(err))
	}

	// Initialize agent registry
	agentRegistry := agents.NewRegistry(logger)

	// Register all 9 AI agents
	if err := registerAgents(agentRegistry, cfg, logger); err != nil {
		logger.Fatal("Failed to register agents", zap.Error(err))
	}

	// Create orchestrator
	orchestrator := orchestrator.NewOrchestrator(
		agentRegistry,
		kafkaManager,
		logger,
		cfg.AI.Orchestrator,
	)

	// Start orchestrator
	ctx := context.Background()
	if err := orchestrator.Start(ctx); err != nil {
		logger.Fatal("Failed to start orchestrator", zap.Error(err))
	}

	// Start health check server
	go startHealthServer(logger, cfg)

	// Create HTTP handler
	handler := httpTransport.NewHandler(orchestrator, logger)

	// Setup HTTP routes
	mux := http.NewServeMux()
	
	// Agent management endpoints
	mux.HandleFunc("/agents", handler.ListAgents)
	mux.HandleFunc("/agents/", handler.GetAgent)
	mux.HandleFunc("/agents/register", handler.RegisterAgent)
	mux.HandleFunc("/agents/unregister", handler.UnregisterAgent)
	
	// Workflow endpoints
	mux.HandleFunc("/workflows", handler.ListWorkflows)
	mux.HandleFunc("/workflows/", handler.GetWorkflow)
	mux.HandleFunc("/workflows/create", handler.CreateWorkflow)
	mux.HandleFunc("/workflows/execute", handler.ExecuteWorkflow)
	mux.HandleFunc("/workflows/stop", handler.StopWorkflow)
	
	// Task coordination endpoints
	mux.HandleFunc("/tasks", handler.ListTasks)
	mux.HandleFunc("/tasks/", handler.GetTask)
	mux.HandleFunc("/tasks/assign", handler.AssignTask)
	mux.HandleFunc("/tasks/complete", handler.CompleteTask)
	
	// Communication endpoints
	mux.HandleFunc("/messages/send", handler.SendMessage)
	mux.HandleFunc("/messages/broadcast", handler.BroadcastMessage)
	
	// External integration endpoints
	mux.HandleFunc("/integrations/clickup", handler.ClickUpIntegration)
	mux.HandleFunc("/integrations/slack", handler.SlackIntegration)
	mux.HandleFunc("/integrations/sheets", handler.GoogleSheetsIntegration)
	mux.HandleFunc("/integrations/airtable", handler.AirtableIntegration)
	
	// Observability endpoints
	mux.HandleFunc("/health", handler.HealthCheck)
	mux.HandleFunc("/ready", handler.ReadinessCheck)
	mux.Handle("/metrics", promhttp.Handler())

	// Create HTTP server with enhanced configuration
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.AI.Orchestrator.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in goroutine
	go func() {
		logger.Info("Starting AI Orchestrator HTTP server", zap.Int("port", cfg.AI.Orchestrator.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down AI Orchestrator...")

	// Create context with timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	// Stop orchestrator
	orchestrator.Stop()

	logger.Info("AI Orchestrator stopped gracefully")
}

// registerAgents registers all 9 AI agents with the registry
func registerAgents(registry *agents.Registry, cfg *config.Config, logger *zap.Logger) error {
	logger.Info("Registering AI agents...")

	// 1. Beverage Inventor Agent
	beverageAgent := agents.NewBeverageInventorAgent(cfg.AI.Agents.BeverageInventor, logger)
	if err := registry.Register("beverage-inventor", beverageAgent); err != nil {
		return fmt.Errorf("failed to register beverage inventor agent: %w", err)
	}

	// 2. Inventory Manager Agent
	inventoryAgent := agents.NewInventoryManagerAgent(cfg.AI.Agents.InventoryManager, logger)
	if err := registry.Register("inventory-manager", inventoryAgent); err != nil {
		return fmt.Errorf("failed to register inventory manager agent: %w", err)
	}

	// 3. Task Manager Agent
	taskAgent := agents.NewTaskManagerAgent(cfg.AI.Agents.TaskManager, logger)
	if err := registry.Register("task-manager", taskAgent); err != nil {
		return fmt.Errorf("failed to register task manager agent: %w", err)
	}

	// 4. Social Media Content Agent
	socialAgent := agents.NewSocialMediaAgent(cfg.AI.Agents.SocialMedia, logger)
	if err := registry.Register("social-media", socialAgent); err != nil {
		return fmt.Errorf("failed to register social media agent: %w", err)
	}

	// 5. Feedback Analyst Agent
	feedbackAgent := agents.NewFeedbackAnalystAgent(cfg.AI.Agents.FeedbackAnalyst, logger)
	if err := registry.Register("feedback-analyst", feedbackAgent); err != nil {
		return fmt.Errorf("failed to register feedback analyst agent: %w", err)
	}

	// 6. Scheduler Agent
	schedulerAgent := agents.NewSchedulerAgent(cfg.AI.Agents.Scheduler, logger)
	if err := registry.Register("scheduler", schedulerAgent); err != nil {
		return fmt.Errorf("failed to register scheduler agent: %w", err)
	}

	// 7. Inter-Location Coordinator Agent
	coordinatorAgent := agents.NewInterLocationCoordinatorAgent(cfg.AI.Agents.InterLocationCoordinator, logger)
	if err := registry.Register("inter-location-coordinator", coordinatorAgent); err != nil {
		return fmt.Errorf("failed to register inter-location coordinator agent: %w", err)
	}

	// 8. Notifier Agent
	notifierAgent := agents.NewNotifierAgent(cfg.AI.Agents.Notifier, logger)
	if err := registry.Register("notifier", notifierAgent); err != nil {
		return fmt.Errorf("failed to register notifier agent: %w", err)
	}

	// 9. Tasting Coordinator Agent
	tastingAgent := agents.NewTastingCoordinatorAgent(cfg.AI.Agents.TastingCoordinator, logger)
	if err := registry.Register("tasting-coordinator", tastingAgent); err != nil {
		return fmt.Errorf("failed to register tasting coordinator agent: %w", err)
	}

	logger.Info("Successfully registered all 9 AI agents")
	return nil
}

// initLogger initializes a structured logger with appropriate configuration
func initLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stacktrace"

	logger, err := config.Build()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	return logger
}

// startHealthServer starts a health check server for the AI orchestrator
func startHealthServer(logger *zap.Logger, cfg *config.Config) {
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"status":"ok",
			"service":"ai-orchestrator",
			"timestamp":"%s",
			"version":"1.0.0",
			"agents":9
		}`, time.Now().UTC().Format(time.RFC3339))
	})
	
	// Readiness check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"status":"ready",
			"service":"ai-orchestrator",
			"timestamp":"%s",
			"checks":{
				"kafka":"ok",
				"agents":"ok",
				"workflows":"ok"
			}
		}`, time.Now().UTC().Format(time.RFC3339))
	})
	
	// Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Start health server on a different port
	healthPort := 8095
	if cfg.AI.Orchestrator.HealthPort != 0 {
		healthPort = cfg.AI.Orchestrator.HealthPort
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", healthPort),
		Handler: mux,
	}

	logger.Info("Starting AI Orchestrator health check server", zap.Int("port", healthPort))
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Health server error", zap.Error(err))
	}
}
