// Package main demonstrates the Go Coffee LangGraph integration
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

	"github.com/DimaJoyti/go-coffee/ai-agents/langgraph-integration/pkg/agents"
	"github.com/DimaJoyti/go-coffee/ai-agents/langgraph-integration/pkg/graph"
	"github.com/DimaJoyti/go-coffee/ai-agents/langgraph-integration/pkg/orchestrator"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Application represents the main application
type Application struct {
	orchestrator *orchestrator.LangGraphOrchestrator
	server       *http.Server
	config       *Config
}

// Config represents application configuration
type Config struct {
	Port         string `json:"port"`
	Environment  string `json:"environment"`
	LogLevel     string `json:"log_level"`
	EnableMetrics bool  `json:"enable_metrics"`
}

// WorkflowRequest represents a workflow execution request
type WorkflowRequest struct {
	GraphID   string                 `json:"graph_id"`
	InputData map[string]interface{} `json:"input_data"`
	Priority  string                 `json:"priority"`
	Config    map[string]interface{} `json:"config"`
}

// WorkflowResponse represents a workflow execution response
type WorkflowResponse struct {
	WorkflowID  uuid.UUID              `json:"workflow_id"`
	ExecutionID uuid.UUID              `json:"execution_id"`
	Status      graph.WorkflowStatus   `json:"status"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Duration    float64                `json:"duration_seconds,omitempty"`
}

func main() {
	fmt.Println("🚀 Starting Go Coffee LangGraph Integration...")

	// Load configuration
	config := &Config{
		Port:         getEnv("PORT", "8080"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		EnableMetrics: getEnv("ENABLE_METRICS", "true") == "true",
	}

	// Create application
	app, err := NewApplication(config)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Start application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.Start(ctx); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown application
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Failed to shutdown application: %v", err)
	}

	fmt.Println("✅ Go Coffee LangGraph Integration shutdown complete")
}

// NewApplication creates a new application instance
func NewApplication(config *Config) (*Application, error) {
	// Create orchestrator
	orchestratorConfig := &orchestrator.OrchestratorConfig{
		MaxConcurrentExecutions: 50,
		DefaultTimeout:          10 * time.Minute,
		MonitoringEnabled:       config.EnableMetrics,
	}

	orch := orchestrator.NewLangGraphOrchestrator(orchestratorConfig)

	// Register agents
	if err := registerAgents(orch); err != nil {
		return nil, fmt.Errorf("failed to register agents: %w", err)
	}

	// Register graphs
	if err := registerGraphs(orch); err != nil {
		return nil, fmt.Errorf("failed to register graphs: %w", err)
	}

	// Create HTTP server
	router := mux.NewRouter()
	setupRoutes(router, orch)

	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: router,
	}

	return &Application{
		orchestrator: orch,
		server:       server,
		config:       config,
	}, nil
}

// Start starts the application
func (app *Application) Start(ctx context.Context) error {
	log.Printf("Starting Go Coffee LangGraph Integration on port %s", app.config.Port)

	// Start HTTP server
	go func() {
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	log.Printf("✅ Go Coffee LangGraph Integration started successfully")
	return nil
}

// Shutdown shuts down the application
func (app *Application) Shutdown(ctx context.Context) error {
	log.Printf("Shutting down Go Coffee LangGraph Integration...")

	// Shutdown orchestrator
	if err := app.orchestrator.Shutdown(ctx); err != nil {
		log.Printf("Warning: Failed to shutdown orchestrator: %v", err)
	}

	// Shutdown HTTP server
	if err := app.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	return nil
}

// registerAgents registers all available agents
func registerAgents(orch *orchestrator.LangGraphOrchestrator) error {
	// Register Beverage Inventor Agent
	beverageAgent := agents.NewBeverageInventorAgent()
	if err := orch.RegisterAgent(graph.AgentTypeBeverageInventor, beverageAgent); err != nil {
		return fmt.Errorf("failed to register beverage inventor agent: %w", err)
	}

	// TODO: Register other agents as they are implemented
	// inventoryAgent := agents.NewInventoryManagerAgent()
	// if err := orch.RegisterAgent(graph.AgentTypeInventoryManager, inventoryAgent); err != nil {
	//     return fmt.Errorf("failed to register inventory manager agent: %w", err)
	// }

	log.Printf("✅ Registered %d agents", 1) // Update count as more agents are added
	return nil
}

// registerGraphs registers workflow graphs
func registerGraphs(orch *orchestrator.LangGraphOrchestrator) error {
	// Create coffee creation workflow graph
	coffeeGraph := createCoffeeCreationWorkflow()
	if err := orch.RegisterGraph("coffee_creation", coffeeGraph); err != nil {
		return fmt.Errorf("failed to register coffee creation graph: %w", err)
	}

	log.Printf("✅ Registered workflow graphs")
	return nil
}

// createCoffeeCreationWorkflow creates a coffee creation workflow graph
func createCoffeeCreationWorkflow() *graph.Graph {
	workflowGraph := graph.NewGraph(
		"Coffee Creation Workflow",
		"End-to-end coffee beverage creation and management workflow",
	)

	// Add workflow start node
	startNode := &graph.Node{
		ID:          "start",
		Type:        graph.NodeTypeControl,
		Name:        "Workflow Start",
		Description: "Initialize coffee creation workflow",
		Function: func(ctx context.Context, state *graph.AgentState) (*graph.AgentState, error) {
			state.StartExecution()
			state.AddSystemMessage("Starting coffee creation workflow")
			log.Printf("Started coffee creation workflow %s", state.WorkflowID)
			return state, nil
		},
		Timeout: 5 * time.Second,
	}

	// Add beverage invention node (will be replaced by agent node)
	beverageNode := &graph.Node{
		ID:          string(graph.AgentTypeBeverageInventor),
		Type:        graph.NodeTypeAgent,
		Name:        "Beverage Inventor",
		Description: "Create innovative coffee recipes",
		Timeout:     60 * time.Second,
		Retries:     2,
	}

	// Add completion node
	completeNode := &graph.Node{
		ID:          "complete",
		Type:        graph.NodeTypeControl,
		Name:        "Workflow Complete",
		Description: "Complete the coffee creation workflow",
		Function: func(ctx context.Context, state *graph.AgentState) (*graph.AgentState, error) {
			state.CompleteExecution()
			state.AddSystemMessage("Coffee creation workflow completed successfully")
			log.Printf("Completed coffee creation workflow %s", state.WorkflowID)
			return state, nil
		},
		Timeout: 5 * time.Second,
	}

	// Add nodes to graph
	workflowGraph.AddNode(startNode)
	workflowGraph.AddNode(beverageNode)
	workflowGraph.AddNode(completeNode)

	// Add edges
	workflowGraph.AddEdge(&graph.Edge{From: "start", To: string(graph.AgentTypeBeverageInventor)})
	workflowGraph.AddEdge(&graph.Edge{From: string(graph.AgentTypeBeverageInventor), To: "complete"})

	// Set start and end nodes
	workflowGraph.SetStartNode("start")
	workflowGraph.AddEndNode("complete")

	return workflowGraph
}

// setupRoutes sets up HTTP routes
func setupRoutes(router *mux.Router, orch *orchestrator.LangGraphOrchestrator) {
	// Health check
	router.HandleFunc("/health", healthHandler).Methods("GET")

	// Workflow execution
	router.HandleFunc("/api/v1/workflows/execute", executeWorkflowHandler(orch)).Methods("POST")

	// Execution status
	router.HandleFunc("/api/v1/executions/{id}", getExecutionStatusHandler(orch)).Methods("GET")

	// List active executions
	router.HandleFunc("/api/v1/executions", listExecutionsHandler(orch)).Methods("GET")

	// Cancel execution
	router.HandleFunc("/api/v1/executions/{id}/cancel", cancelExecutionHandler(orch)).Methods("POST")

	// Orchestrator stats
	router.HandleFunc("/api/v1/stats", getStatsHandler(orch)).Methods("GET")
}

// healthHandler handles health check requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "go-coffee-langgraph",
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// executeWorkflowHandler handles workflow execution requests
func executeWorkflowHandler(orch *orchestrator.LangGraphOrchestrator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request WorkflowRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Create initial state
		workflowID := uuid.New()
		executionID := uuid.New()
		
		state := graph.NewAgentState(workflowID, executionID)
		state.InputData = request.InputData
		
		if request.Priority != "" {
			state.Priority = graph.Priority(request.Priority)
		}

		// Execute workflow
		start := time.Now()
		finalState, err := orch.ExecuteWorkflow(r.Context(), request.GraphID, state)
		duration := time.Since(start).Seconds()

		// Create response
		response := WorkflowResponse{
			WorkflowID:  workflowID,
			ExecutionID: executionID,
			Status:      finalState.Status,
			Duration:    duration,
		}

		if err != nil {
			response.Error = err.Error()
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			response.Result = finalState.AgentOutputs
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// getExecutionStatusHandler handles execution status requests
func getExecutionStatusHandler(orch *orchestrator.LangGraphOrchestrator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		executionID, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "Invalid execution ID", http.StatusBadRequest)
			return
		}

		execution, err := orch.GetExecutionStatus(executionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(execution)
	}
}

// listExecutionsHandler handles list executions requests
func listExecutionsHandler(orch *orchestrator.LangGraphOrchestrator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		executions := orch.ListActiveExecutions()

		response := map[string]interface{}{
			"executions": executions,
			"count":      len(executions),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// cancelExecutionHandler handles execution cancellation requests
func cancelExecutionHandler(orch *orchestrator.LangGraphOrchestrator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		executionID, err := uuid.Parse(vars["id"])
		if err != nil {
			http.Error(w, "Invalid execution ID", http.StatusBadRequest)
			return
		}

		if err := orch.CancelExecution(executionID); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		response := map[string]string{
			"status":  "cancelled",
			"message": "Execution cancelled successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// getStatsHandler handles stats requests
func getStatsHandler(orch *orchestrator.LangGraphOrchestrator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := orch.GetStats()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
