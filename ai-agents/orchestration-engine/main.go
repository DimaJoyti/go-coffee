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

	"go-coffee-ai-agents/orchestration-engine/examples"
	"go-coffee-ai-agents/orchestration-engine/internal/domain/entities"
	"go-coffee-ai-agents/orchestration-engine/internal/domain/services"

	"github.com/google/uuid"
)

// Application represents the main orchestration engine application
type Application struct {
	config         *Config
	server         *http.Server
	workflowEngine *services.WorkflowEngine
	agentRegistry  *services.DefaultAgentRegistry
	logger         Logger
}

// Config represents application configuration
type Config struct {
	Port                   string `env:"PORT" envDefault:"8080"`
	DatabaseURL            string `env:"DATABASE_URL" envDefault:"postgres://localhost/orchestration"`
	RedisURL               string `env:"REDIS_URL" envDefault:"redis://localhost:6379"`
	KafkaBrokers           string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
	LogLevel               string `env:"LOG_LEVEL" envDefault:"info"`
	Environment            string `env:"ENVIRONMENT" envDefault:"development"`
	MaxConcurrentWorkflows int    `env:"MAX_CONCURRENT_WORKFLOWS" envDefault:"100"`

	// Agent endpoints
	SocialMediaAgentURL      string `env:"SOCIAL_MEDIA_AGENT_URL" envDefault:"http://localhost:8081"`
	FeedbackAnalystAgentURL  string `env:"FEEDBACK_ANALYST_AGENT_URL" envDefault:"http://localhost:8082"`
	BeverageInventorAgentURL string `env:"BEVERAGE_INVENTOR_AGENT_URL" envDefault:"http://localhost:8083"`
	InventoryAgentURL        string `env:"INVENTORY_AGENT_URL" envDefault:"http://localhost:8084"`
	NotifierAgentURL         string `env:"NOTIFIER_AGENT_URL" envDefault:"http://localhost:8085"`
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

// Mock repositories for demonstration
type MockWorkflowRepository struct {
	workflows map[uuid.UUID]*entities.Workflow
	logger    Logger
}

func NewMockWorkflowRepository(logger Logger) *MockWorkflowRepository {
	return &MockWorkflowRepository{
		workflows: make(map[uuid.UUID]*entities.Workflow),
		logger:    logger,
	}
}

func (r *MockWorkflowRepository) Create(ctx context.Context, workflow *entities.Workflow) error {
	r.workflows[workflow.ID] = workflow
	r.logger.Info("Workflow created", "workflow_id", workflow.ID, "name", workflow.Name)
	return nil
}

func (r *MockWorkflowRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Workflow, error) {
	workflow, exists := r.workflows[id]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", id)
	}
	return workflow, nil
}

func (r *MockWorkflowRepository) Update(ctx context.Context, workflow *entities.Workflow) error {
	r.workflows[workflow.ID] = workflow
	r.logger.Info("Workflow updated", "workflow_id", workflow.ID)
	return nil
}

func (r *MockWorkflowRepository) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.workflows, id)
	r.logger.Info("Workflow deleted", "workflow_id", id)
	return nil
}

func (r *MockWorkflowRepository) List(ctx context.Context, filter *services.WorkflowFilter) ([]*entities.Workflow, error) {
	var result []*entities.Workflow
	for _, workflow := range r.workflows {
		result = append(result, workflow)
	}
	return result, nil
}

func (r *MockWorkflowRepository) GetActiveWorkflows(ctx context.Context) ([]*entities.Workflow, error) {
	var result []*entities.Workflow
	for _, workflow := range r.workflows {
		if workflow.IsActive && workflow.Status == entities.WorkflowStatusActive {
			result = append(result, workflow)
		}
	}
	return result, nil
}

func (r *MockWorkflowRepository) GetWorkflowsByTrigger(ctx context.Context, triggerType entities.TriggerType) ([]*entities.Workflow, error) {
	var result []*entities.Workflow
	for _, workflow := range r.workflows {
		for _, trigger := range workflow.Triggers {
			if trigger.Type == triggerType && trigger.IsActive {
				result = append(result, workflow)
				break
			}
		}
	}
	return result, nil
}

type MockExecutionRepository struct {
	executions map[uuid.UUID]*entities.WorkflowExecution
	logger     Logger
}

func NewMockExecutionRepository(logger Logger) *MockExecutionRepository {
	return &MockExecutionRepository{
		executions: make(map[uuid.UUID]*entities.WorkflowExecution),
		logger:     logger,
	}
}

func (r *MockExecutionRepository) Create(ctx context.Context, execution *entities.WorkflowExecution) error {
	r.executions[execution.ID] = execution
	r.logger.Info("Execution created", "execution_id", execution.ID, "workflow_id", execution.WorkflowID)
	return nil
}

func (r *MockExecutionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.WorkflowExecution, error) {
	execution, exists := r.executions[id]
	if !exists {
		return nil, fmt.Errorf("execution not found: %s", id)
	}
	return execution, nil
}

func (r *MockExecutionRepository) Update(ctx context.Context, execution *entities.WorkflowExecution) error {
	r.executions[execution.ID] = execution
	r.logger.Info("Execution updated", "execution_id", execution.ID, "status", execution.Status)
	return nil
}

func (r *MockExecutionRepository) List(ctx context.Context, filter *services.ExecutionFilter) ([]*entities.WorkflowExecution, error) {
	var result []*entities.WorkflowExecution
	for _, execution := range r.executions {
		result = append(result, execution)
	}
	return result, nil
}

func (r *MockExecutionRepository) GetActiveExecutions(ctx context.Context, workflowID uuid.UUID) ([]*entities.WorkflowExecution, error) {
	var result []*entities.WorkflowExecution
	for _, execution := range r.executions {
		if execution.WorkflowID == workflowID && execution.Status == entities.WorkflowStatusRunning {
			result = append(result, execution)
		}
	}
	return result, nil
}

func (r *MockExecutionRepository) GetExecutionHistory(ctx context.Context, workflowID uuid.UUID, limit int) ([]*entities.WorkflowExecution, error) {
	var result []*entities.WorkflowExecution
	count := 0
	for _, execution := range r.executions {
		if execution.WorkflowID == workflowID && count < limit {
			result = append(result, execution)
			count++
		}
	}
	return result, nil
}

// Mock event publisher
type MockEventPublisher struct {
	logger Logger
}

func NewMockEventPublisher(logger Logger) *MockEventPublisher {
	return &MockEventPublisher{logger: logger}
}

func (p *MockEventPublisher) PublishWorkflowEvent(ctx context.Context, event *services.WorkflowEvent) error {
	p.logger.Info("Workflow event published", "type", event.Type, "workflow_id", event.WorkflowID)
	return nil
}

func (p *MockEventPublisher) PublishExecutionEvent(ctx context.Context, event *services.ExecutionEvent) error {
	p.logger.Info("Execution event published", "type", event.Type, "execution_id", event.ExecutionID)
	return nil
}

func (p *MockEventPublisher) PublishStepEvent(ctx context.Context, event *services.StepEvent) error {
	p.logger.Info("Step event published", "type", event.Type, "step_id", event.StepID)
	return nil
}

// NewApplication creates a new orchestration engine application
func NewApplication(config *Config) (*Application, error) {
	logger := NewLogger(config.LogLevel)
	httpClient := NewSimpleHTTPClient(logger)

	// Initialize repositories
	workflowRepo := NewMockWorkflowRepository(logger)
	executionRepo := NewMockExecutionRepository(logger)
	eventPublisher := NewMockEventPublisher(logger)

	// Initialize agent registry
	agentRegistry := services.NewDefaultAgentRegistry(logger)

	// Register agents
	socialMediaAgent := services.NewSocialMediaContentAgent(config.SocialMediaAgentURL, httpClient, logger)
	feedbackAgent := services.NewFeedbackAnalystAgent(config.FeedbackAnalystAgentURL, httpClient, logger)
	beverageAgent := services.NewBeverageInventorAgent(config.BeverageInventorAgentURL, httpClient, logger)

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
		config.MaxConcurrentWorkflows,
	)

	// Setup HTTP server
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := fmt.Sprintf(`{
			"status": "healthy",
			"timestamp": "%s",
			"version": "1.0.0",
			"service": "orchestration-engine"
		}`, time.Now().UTC().Format(time.RFC3339))
		w.Write([]byte(response))
	})

	// Workflow management endpoints
	mux.HandleFunc("/api/v1/workflows", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"message": "Workflow created", "status": "success"}`))
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"workflows": [], "status": "success"}`))
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Workflow execution endpoints
	mux.HandleFunc("/api/v1/workflows/execute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"message": "Workflow execution started", "execution_id": "` + uuid.New().String() + `"}`))
	})

	// Agent status endpoints
	mux.HandleFunc("/api/v1/agents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"agents": {"social-media-content": "online", "feedback-analyst": "online"}, "status": "success"}`))
	})

	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: mux,
	}

	return &Application{
		config:         config,
		server:         server,
		workflowEngine: workflowEngine,
		agentRegistry:  agentRegistry,
		logger:         logger,
	}, nil
}

// Start starts the orchestration engine
func (app *Application) Start(ctx context.Context) error {
	app.logger.Info("Starting Orchestration Engine", "port", app.config.Port)

	// Start workflow engine
	if err := app.workflowEngine.Start(ctx); err != nil {
		return fmt.Errorf("failed to start workflow engine: %w", err)
	}

	// Start agent health monitoring
	go app.agentRegistry.MonitorAgentHealth(ctx)

	// Create and register example workflow
	go func() {
		time.Sleep(2 * time.Second) // Wait for engine to be ready
		app.createExampleWorkflow(ctx)
	}()

	// Start HTTP server
	go func() {
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.logger.Error("Failed to start HTTP server", err)
		}
	}()

	app.logger.Info("Orchestration Engine started successfully", "port", app.config.Port)
	return nil
}

// Stop stops the orchestration engine
func (app *Application) Stop(ctx context.Context) error {
	app.logger.Info("Stopping Orchestration Engine")

	// Stop workflow engine
	if err := app.workflowEngine.Stop(ctx); err != nil {
		app.logger.Error("Failed to stop workflow engine", err)
	}

	// Stop HTTP server
	if err := app.server.Shutdown(ctx); err != nil {
		app.logger.Error("Failed to shutdown HTTP server", err)
		return err
	}

	app.logger.Info("Orchestration Engine stopped successfully")
	return nil
}

// createExampleWorkflow creates and registers the example content creation workflow
func (app *Application) createExampleWorkflow(ctx context.Context) {
	app.logger.Info("Creating example content creation workflow")

	// Create the workflow
	workflow := examples.CreateContentCreationWorkflow(uuid.New())

	// Register the workflow (in a real implementation, this would be done via API)
	// For now, just log that the workflow was created
	app.logger.Info("Example workflow created", "workflow_id", workflow.ID, "name", workflow.Name)

	// Demonstrate workflow execution
	go func() {
		time.Sleep(5 * time.Second)
		app.demonstrateWorkflowExecution(ctx, workflow.ID)
	}()
}

// demonstrateWorkflowExecution demonstrates executing the example workflow
func (app *Application) demonstrateWorkflowExecution(ctx context.Context, workflowID uuid.UUID) {
	app.logger.Info("Demonstrating workflow execution", "workflow_id", workflowID)

	// Example input for content creation workflow
	input := map[string]interface{}{
		"brand_id":            uuid.New().String(),
		"content_topic":       "New Coffee Blend Launch",
		"target_platforms":    []string{"instagram", "facebook", "twitter"},
		"content_type":        "post",
		"tone":                "exciting",
		"auto_publish":        false,
		"require_approval":    true,
		"generate_variations": true,
		"analyze_feedback":    true,
	}

	execution, err := app.workflowEngine.ExecuteWorkflow(ctx, workflowID, input, "manual", uuid.New())
	if err != nil {
		app.logger.Error("Failed to execute workflow", err, "workflow_id", workflowID)
		return
	}

	app.logger.Info("Workflow execution started", "execution_id", execution.ID, "workflow_id", workflowID)
}

func main() {
	fmt.Println("🚀 Starting Go Coffee AI Agents Orchestration Engine...")

	// Load configuration from environment variables
	config := &Config{
		Port:                    getEnv("PORT", "8080"),
		DatabaseURL:             getEnv("DATABASE_URL", "postgres://localhost/orchestration"),
		RedisURL:                getEnv("REDIS_URL", "redis://localhost:6379"),
		KafkaBrokers:            getEnv("KAFKA_BROKERS", "localhost:9092"),
		LogLevel:                getEnv("LOG_LEVEL", "info"),
		Environment:             getEnv("ENVIRONMENT", "development"),
		MaxConcurrentWorkflows:  100,
		SocialMediaAgentURL:     getEnv("SOCIAL_MEDIA_AGENT_URL", "http://localhost:8081"),
		FeedbackAnalystAgentURL: getEnv("FEEDBACK_ANALYST_AGENT_URL", "http://localhost:8082"),
		InventoryAgentURL:       getEnv("INVENTORY_AGENT_URL", "http://localhost:8083"),
		NotifierAgentURL:        getEnv("NOTIFIER_AGENT_URL", "http://localhost:8084"),
	}

	// Create application
	app, err := NewApplication(config)
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

	// Stop application
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.Stop(shutdownCtx); err != nil {
		log.Fatalf("Failed to stop application: %v", err)
	}

	fmt.Println("✅ Go Coffee AI Agents Orchestration Engine shutdown complete")
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
