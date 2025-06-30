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

	"github.com/google/uuid"
	"go-coffee-ai-agents/orchestration-engine/examples"
	"go-coffee-ai-agents/orchestration-engine/internal/config"
	"go-coffee-ai-agents/orchestration-engine/internal/domain/services"
	httpHandlers "go-coffee-ai-agents/orchestration-engine/internal/interfaces/http"
	"go-coffee-ai-agents/orchestration-engine/internal/interfaces/websocket"
)

// Application represents the main orchestration engine application
type Application struct {
	config           *config.Config
	server           *http.Server
	workflowEngine   *services.WorkflowEngine
	agentRegistry    *services.DefaultAgentRegistry
	analyticsService *services.AnalyticsService
	webSocketHub     *websocket.Hub
	httpHandlers     *httpHandlers.OrchestrationHandlers
	logger           Logger
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// SimpleLogger implements the Logger interface
type SimpleLogger struct {
	level string
}

func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	if l.level == "debug" {
		log.Printf("[DEBUG] "+msg, args...)
	}
}

func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

func (l *SimpleLogger) Warn(msg string, args ...interface{}) {
	log.Printf("[WARN] "+msg, args...)
}

func (l *SimpleLogger) Error(msg string, err error, args ...interface{}) {
	if err != nil {
		log.Printf("[ERROR] "+msg+": %v", append(args, err)...)
	} else {
		log.Printf("[ERROR] "+msg, args...)
	}
}

// NewLogger creates a new logger instance
func NewLogger(level string) Logger {
	return &SimpleLogger{level: level}
}

// SimpleHTTPClient implements the HTTPClient interface
type SimpleHTTPClient struct {
	client *http.Client
	logger Logger
}

func NewSimpleHTTPClient(logger Logger) *SimpleHTTPClient {
	return &SimpleHTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

func (c *SimpleHTTPClient) Post(ctx context.Context, url string, data interface{}) (map[string]interface{}, error) {
	c.logger.Info("HTTP POST", "url", url)
	// Simplified implementation - in production this would make actual HTTP calls
	return map[string]interface{}{
		"status":  "success",
		"message": "Mock response from " + url,
		"data":    data,
	}, nil
}

func (c *SimpleHTTPClient) Get(ctx context.Context, url string) (map[string]interface{}, error) {
	c.logger.Info("HTTP GET", "url", url)
	return map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
	}, nil
}

func (c *SimpleHTTPClient) Put(ctx context.Context, url string, data interface{}) (map[string]interface{}, error) {
	c.logger.Info("HTTP PUT", "url", url)
	return map[string]interface{}{
		"status": "updated",
		"data":   data,
	}, nil
}

func (c *SimpleHTTPClient) Delete(ctx context.Context, url string) error {
	c.logger.Info("HTTP DELETE", "url", url)
	return nil
}

// NewApplication creates a new orchestration engine application
func NewApplication(cfg *config.Config) (*Application, error) {
	logger := NewLogger(cfg.Logging.Level)
	httpClient := NewSimpleHTTPClient(logger)

	// Initialize repositories (mock implementations for now)
	workflowRepo := NewMockWorkflowRepository(logger)
	executionRepo := NewMockExecutionRepository(logger)
	eventPublisher := NewMockEventPublisher(logger)

	// Initialize agent registry
	agentRegistry := services.NewDefaultAgentRegistry(logger)

	// Register agents
	socialMediaAgent := services.NewSocialMediaContentAgent(
		cfg.Agents.Endpoints["social-media-content"], 
		httpClient, 
		logger,
	)
	feedbackAgent := services.NewFeedbackAnalystAgent(
		cfg.Agents.Endpoints["feedback-analyst"], 
		httpClient, 
		logger,
	)
	beverageAgent := services.NewBeverageInventorAgent(
		cfg.Agents.Endpoints["beverage-inventor"], 
		httpClient, 
		logger,
	)

	agentRegistry.RegisterAgent("social-media-content", socialMediaAgent)
	agentRegistry.RegisterAgent("feedback-analyst", feedbackAgent)
	agentRegistry.RegisterAgent("beverage-inventor", beverageAgent)

	// Initialize workflow engine
	workflowEngine := services.NewWorkflowEngine(
		workflowRepo,
		executionRepo,
		agentRegistry,
		eventPublisher,
		logger,
		cfg.Analytics.MaxMetricsHistory,
	)

	// Initialize analytics service
	analyticsService := services.NewAnalyticsService(
		workflowRepo,
		executionRepo,
		agentRegistry,
		eventPublisher,
		logger,
	)

	// Initialize WebSocket hub
	var webSocketHub *websocket.Hub
	if cfg.Features.EnableWebSocket {
		webSocketHub = websocket.NewHub(analyticsService, logger)
	}

	// Initialize HTTP handlers
	httpHandlers := httpHandlers.NewOrchestrationHandlers(
		workflowEngine,
		agentRegistry,
		analyticsService,
		logger,
	)

	// Setup HTTP server
	mux := http.NewServeMux()

	// Register HTTP routes
	httpHandlers.RegisterRoutes(mux)

	// Register WebSocket endpoint if enabled
	if webSocketHub != nil {
		mux.HandleFunc("/ws", webSocketHub.WebSocketHandler)
		mux.HandleFunc("/api/v1/websocket/stats", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			stats := webSocketHub.GetStats()
			w.WriteHeader(http.StatusOK)
			// Write stats JSON - simplified for now
			fmt.Fprintf(w, `{"status":"ok","connections":%d}`, len(stats))
		})
	}

	// Add middleware
	handler := addMiddleware(mux, cfg, logger)

	server := &http.Server{
		Addr:           cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:        handler,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	return &Application{
		config:           cfg,
		server:           server,
		workflowEngine:   workflowEngine,
		agentRegistry:    agentRegistry,
		analyticsService: analyticsService,
		webSocketHub:     webSocketHub,
		httpHandlers:     httpHandlers,
		logger:           logger,
	}, nil
}

// addMiddleware adds middleware to the HTTP handler
func addMiddleware(handler http.Handler, cfg *config.Config, logger Logger) http.Handler {
	// Add CORS middleware if enabled
	if cfg.Server.EnableCORS {
		handler = corsMiddleware(handler, cfg.Server.CORSAllowedOrigins)
	}

	// Add logging middleware
	handler = loggingMiddleware(handler, logger)

	// Add rate limiting middleware if enabled
	if cfg.Server.EnableRateLimiting {
		handler = rateLimitMiddleware(handler, cfg.Server.RateLimitRPS, cfg.Server.RateLimitBurst)
	}

	// Add compression middleware if enabled
	if cfg.Server.EnableCompression {
		handler = compressionMiddleware(handler)
	}

	return handler
}

// Start starts the orchestration engine
func (app *Application) Start(ctx context.Context) error {
	app.logger.Info("Starting Go Coffee AI Agents Orchestration Engine", 
		"port", app.config.Server.Port,
		"version", "1.0.0",
		"environment", app.config.Logging.Level)

	// Start workflow engine
	if err := app.workflowEngine.Start(ctx); err != nil {
		return fmt.Errorf("failed to start workflow engine: %w", err)
	}

	// Start analytics service
	if err := app.analyticsService.Start(ctx); err != nil {
		return fmt.Errorf("failed to start analytics service: %w", err)
	}

	// Start agent health monitoring
	go app.agentRegistry.MonitorAgentHealth(ctx)

	// Start WebSocket hub if enabled
	if app.webSocketHub != nil {
		app.webSocketHub.Start(ctx)
	}

	// Create and register example workflows
	go func() {
		time.Sleep(2 * time.Second) // Wait for engine to be ready
		app.createExampleWorkflows(ctx)
	}()

	// Start HTTP server
	go func() {
		app.logger.Info("HTTP server starting", "address", app.server.Addr)
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.logger.Error("Failed to start HTTP server", err)
		}
	}()

	app.logger.Info("🚀 Orchestration Engine started successfully", 
		"port", app.config.Server.Port,
		"websocket_enabled", app.config.Features.EnableWebSocket,
		"analytics_enabled", app.config.Analytics.EnableRealTimeMetrics)

	return nil
}

// Stop stops the orchestration engine
func (app *Application) Stop(ctx context.Context) error {
	app.logger.Info("Stopping Orchestration Engine")

	// Stop WebSocket hub
	if app.webSocketHub != nil {
		app.webSocketHub.Stop()
	}

	// Stop analytics service
	if err := app.analyticsService.Stop(ctx); err != nil {
		app.logger.Error("Failed to stop analytics service", err)
	}

	// Stop workflow engine
	if err := app.workflowEngine.Stop(ctx); err != nil {
		app.logger.Error("Failed to stop workflow engine", err)
	}

	// Stop HTTP server
	if err := app.server.Shutdown(ctx); err != nil {
		app.logger.Error("Failed to shutdown HTTP server", err)
		return err
	}

	app.logger.Info("✅ Orchestration Engine stopped successfully")
	return nil
}

// createExampleWorkflows creates and registers example workflows
func (app *Application) createExampleWorkflows(ctx context.Context) {
	app.logger.Info("Creating example workflows")

	// Create content creation workflow
	contentWorkflow := examples.CreateContentCreationWorkflow(generateUUID())
	app.logger.Info("Content creation workflow created", 
		"workflow_id", contentWorkflow.ID, 
		"name", contentWorkflow.Name)

	// Create beverage innovation workflow
	beverageWorkflow := examples.CreateBeverageInnovationWorkflow(generateUUID())
	app.logger.Info("Beverage innovation workflow created", 
		"workflow_id", beverageWorkflow.ID, 
		"name", beverageWorkflow.Name)

	// Demonstrate workflow execution
	go func() {
		time.Sleep(5 * time.Second)
		app.demonstrateWorkflowExecution(ctx, contentWorkflow.ID.String())
	}()
}

// demonstrateWorkflowExecution demonstrates executing example workflows
func (app *Application) demonstrateWorkflowExecution(ctx context.Context, workflowID string) {
	app.logger.Info("Demonstrating workflow execution", "workflow_id", workflowID)

	// Example input for content creation workflow
	input := map[string]interface{}{
		"brand_id":           generateUUID().String(),
		"content_topic":      "New Coffee Blend Launch",
		"target_platforms":   []string{"instagram", "facebook", "twitter"},
		"content_type":       "post",
		"tone":              "exciting",
		"auto_publish":      false,
		"require_approval":  true,
		"generate_variations": true,
		"analyze_feedback":  true,
	}

	app.logger.Info("Mock workflow execution started", 
		"workflow_id", workflowID,
		"input_topic", input["content_topic"])

	// Broadcast workflow execution event via WebSocket
	if app.webSocketHub != nil {
		app.webSocketHub.BroadcastMessage("workflow_execution", "workflows", map[string]interface{}{
			"workflow_id": workflowID,
			"status":      "started",
			"input":       input,
			"timestamp":   time.Now(),
		})
	}
}

func main() {
	fmt.Println("🚀 Starting Go Coffee AI Agents Orchestration Engine...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Create application
	app, err := NewApplication(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start application
	if err := app.Start(ctx); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 Shutdown signal received...")

	// Stop application
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.Stop(shutdownCtx); err != nil {
		log.Fatalf("Failed to stop application: %v", err)
	}

	fmt.Println("✅ Go Coffee AI Agents Orchestration Engine shutdown complete")
}

// Helper functions

func generateUUID() uuid.UUID {
	// Generate a proper UUID
	return uuid.New()
}
