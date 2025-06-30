package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-coffee-ai-agents/orchestration-engine/internal/domain/services"
)

// OrchestrationHandlers provides HTTP handlers for the orchestration service
type OrchestrationHandlers struct {
	workflowEngine   *services.WorkflowEngine
	agentRegistry    *services.DefaultAgentRegistry
	analyticsService *services.AnalyticsService
	logger           Logger
}

// Logger interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// NewOrchestrationHandlers creates new orchestration handlers
func NewOrchestrationHandlers(
	workflowEngine *services.WorkflowEngine,
	agentRegistry *services.DefaultAgentRegistry,
	analyticsService *services.AnalyticsService,
	logger Logger,
) *OrchestrationHandlers {
	return &OrchestrationHandlers{
		workflowEngine:   workflowEngine,
		agentRegistry:    agentRegistry,
		analyticsService: analyticsService,
		logger:           logger,
	}
}

// RegisterRoutes registers all HTTP routes
func (h *OrchestrationHandlers) RegisterRoutes(mux *http.ServeMux) {
	// Health check
	mux.HandleFunc("/health", h.HealthCheck)
	mux.HandleFunc("/ready", h.ReadinessCheck)

	// Workflow management
	mux.HandleFunc("/api/v1/workflows", h.handleWorkflows)
	mux.HandleFunc("/api/v1/workflows/", h.handleWorkflowByID)
	mux.HandleFunc("/api/v1/workflows/execute", h.ExecuteWorkflow)
	mux.HandleFunc("/api/v1/workflows/templates", h.handleWorkflowTemplates)

	// Execution management
	mux.HandleFunc("/api/v1/executions", h.handleExecutions)
	mux.HandleFunc("/api/v1/executions/", h.handleExecutionByID)

	// Agent management
	mux.HandleFunc("/api/v1/agents", h.handleAgents)
	mux.HandleFunc("/api/v1/agents/", h.handleAgentByType)
	mux.HandleFunc("/api/v1/agents/health", h.AgentHealthCheck)

	// Analytics and monitoring
	mux.HandleFunc("/api/v1/analytics/dashboard", h.GetDashboard)
	mux.HandleFunc("/api/v1/analytics/workflows/", h.GetWorkflowAnalytics)
	mux.HandleFunc("/api/v1/analytics/agents/", h.GetAgentAnalytics)
	mux.HandleFunc("/api/v1/analytics/metrics", h.GetMetrics)

	// System management
	mux.HandleFunc("/api/v1/system/status", h.SystemStatus)
	mux.HandleFunc("/api/v1/system/metrics", h.SystemMetrics)
}

// Health check endpoint
func (h *OrchestrationHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
		"service":   "orchestration-engine",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// Readiness check endpoint
func (h *OrchestrationHandlers) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if all critical components are ready
	ready := true
	checks := map[string]bool{
		"workflow_engine": h.workflowEngine != nil,
		"agent_registry":  h.agentRegistry != nil,
		"analytics":       h.analyticsService != nil,
	}

	for service, status := range checks {
		if !status {
			ready = false
			h.logger.Warn("Service not ready", "service", service)
		}
	}

	status := "ready"
	httpStatus := http.StatusOK
	if !ready {
		status = "not_ready"
		httpStatus = http.StatusServiceUnavailable
	}

	response := map[string]interface{}{
		"status":    status,
		"checks":    checks,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	h.writeJSONResponse(w, httpStatus, response)
}

// Workflow handlers

func (h *OrchestrationHandlers) handleWorkflows(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListWorkflows(w, r)
	case http.MethodPost:
		h.CreateWorkflow(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *OrchestrationHandlers) handleWorkflowByID(w http.ResponseWriter, r *http.Request) {
	workflowID := extractIDFromPath(r.URL.Path, "/api/v1/workflows/")
	if workflowID == "" {
		http.Error(w, "Invalid workflow ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.GetWorkflow(w, r, workflowID)
	case http.MethodPut:
		h.UpdateWorkflow(w, r, workflowID)
	case http.MethodDelete:
		h.DeleteWorkflow(w, r, workflowID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *OrchestrationHandlers) ListWorkflows(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Listing workflows")

	// Parse query parameters
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Mock response for now
	response := map[string]interface{}{
		"workflows": []map[string]interface{}{
			{
				"id":          "workflow-1",
				"name":        "Content Creation Workflow",
				"description": "AI-powered content creation and publishing",
				"status":      "active",
				"created_at":  time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
				"updated_at":  time.Now().Format(time.RFC3339),
			},
			{
				"id":          "workflow-2",
				"name":        "Beverage Innovation Workflow",
				"description": "AI-powered beverage invention and testing",
				"status":      "active",
				"created_at":  time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
				"updated_at":  time.Now().Format(time.RFC3339),
			},
		},
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
			"total":  2,
		},
		"status": "success",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *OrchestrationHandlers) CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Creating workflow")

	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	name, ok := request["name"].(string)
	if !ok || name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Mock workflow creation
	workflowID := fmt.Sprintf("workflow-%d", time.Now().Unix())
	
	response := map[string]interface{}{
		"id":          workflowID,
		"name":        name,
		"description": request["description"],
		"status":      "draft",
		"created_at":  time.Now().Format(time.RFC3339),
		"message":     "Workflow created successfully",
	}

	h.writeJSONResponse(w, http.StatusCreated, response)
}

func (h *OrchestrationHandlers) GetWorkflow(w http.ResponseWriter, r *http.Request, workflowID string) {
	h.logger.Info("Getting workflow", "workflow_id", workflowID)

	// Mock workflow response
	response := map[string]interface{}{
		"id":          workflowID,
		"name":        "Sample Workflow",
		"description": "A sample workflow for demonstration",
		"type":        "sequential",
		"status":      "active",
		"definition": map[string]interface{}{
			"start_step": "validate_input",
			"steps": map[string]interface{}{
				"validate_input": map[string]interface{}{
					"type":        "validation",
					"name":        "Validate Input",
					"description": "Validate workflow input parameters",
				},
			},
		},
		"created_at": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *OrchestrationHandlers) UpdateWorkflow(w http.ResponseWriter, r *http.Request, workflowID string) {
	h.logger.Info("Updating workflow", "workflow_id", workflowID)

	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"id":         workflowID,
		"updated_at": time.Now().Format(time.RFC3339),
		"message":    "Workflow updated successfully",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *OrchestrationHandlers) DeleteWorkflow(w http.ResponseWriter, r *http.Request, workflowID string) {
	h.logger.Info("Deleting workflow", "workflow_id", workflowID)

	response := map[string]interface{}{
		"message": "Workflow deleted successfully",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// Execution handlers

func (h *OrchestrationHandlers) ExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.logger.Info("Executing workflow")

	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	workflowID, ok := request["workflow_id"].(string)
	if !ok || workflowID == "" {
		http.Error(w, "workflow_id is required", http.StatusBadRequest)
		return
	}

	// Mock execution
	executionID := fmt.Sprintf("exec-%d", time.Now().Unix())

	response := map[string]interface{}{
		"execution_id": executionID,
		"workflow_id":  workflowID,
		"status":       "running",
		"started_at":   time.Now().Format(time.RFC3339),
		"message":      "Workflow execution started",
	}

	h.writeJSONResponse(w, http.StatusAccepted, response)
}

func (h *OrchestrationHandlers) handleExecutions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.logger.Info("Listing executions")

	response := map[string]interface{}{
		"executions": []map[string]interface{}{
			{
				"id":          "exec-1",
				"workflow_id": "workflow-1",
				"status":      "completed",
				"started_at":  time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				"completed_at": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				"duration":    "1h",
			},
			{
				"id":          "exec-2",
				"workflow_id": "workflow-2",
				"status":      "running",
				"started_at":  time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
				"duration":    "30m",
			},
		},
		"status": "success",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *OrchestrationHandlers) handleExecutionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	executionID := extractIDFromPath(r.URL.Path, "/api/v1/executions/")
	if executionID == "" {
		http.Error(w, "Invalid execution ID", http.StatusBadRequest)
		return
	}

	h.logger.Info("Getting execution", "execution_id", executionID)

	response := map[string]interface{}{
		"id":          executionID,
		"workflow_id": "workflow-1",
		"status":      "completed",
		"started_at":  time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		"completed_at": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		"duration":    "1h",
		"steps": []map[string]interface{}{
			{
				"id":     "step-1",
				"name":   "Validate Input",
				"status": "completed",
				"duration": "5s",
			},
			{
				"id":     "step-2",
				"name":   "Generate Content",
				"status": "completed",
				"duration": "2m",
			},
		},
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// Agent handlers

func (h *OrchestrationHandlers) handleAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.logger.Info("Listing agents")

	agents := h.agentRegistry.ListAgents()
	agentList := make([]map[string]interface{}, 0, len(agents))

	for agentType, agent := range agents {
		status := agent.GetStatus()
		metrics := agent.GetMetrics()

		agentInfo := map[string]interface{}{
			"type":         agentType,
			"status":       string(status),
			"capabilities": agent.GetCapabilities(),
			"metrics": map[string]interface{}{
				"total_requests":     metrics.TotalRequests,
				"successful_requests": metrics.SuccessfulRequests,
				"failed_requests":    metrics.FailedRequests,
				"average_response_time": metrics.AverageResponseTime.String(),
				"current_load":       metrics.CurrentLoad,
				"last_updated":       metrics.LastUpdated.Format(time.RFC3339),
			},
		}

		agentList = append(agentList, agentInfo)
	}

	response := map[string]interface{}{
		"agents": agentList,
		"total":  len(agentList),
		"status": "success",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *OrchestrationHandlers) handleAgentByType(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentType := extractIDFromPath(r.URL.Path, "/api/v1/agents/")
	if agentType == "" {
		http.Error(w, "Invalid agent type", http.StatusBadRequest)
		return
	}

	h.logger.Info("Getting agent details", "agent_type", agentType)

	agent, err := h.agentRegistry.GetAgent(agentType)
	if err != nil {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	status := agent.GetStatus()
	metrics := agent.GetMetrics()
	health, _ := h.agentRegistry.GetAgentHealth(agentType)

	response := map[string]interface{}{
		"type":         agentType,
		"status":       string(status),
		"capabilities": agent.GetCapabilities(),
		"metrics": map[string]interface{}{
			"total_requests":      metrics.TotalRequests,
			"successful_requests": metrics.SuccessfulRequests,
			"failed_requests":     metrics.FailedRequests,
			"average_response_time": metrics.AverageResponseTime.String(),
			"current_load":        metrics.CurrentLoad,
			"last_updated":        metrics.LastUpdated.Format(time.RFC3339),
		},
		"health": health,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *OrchestrationHandlers) AgentHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.logger.Info("Checking agent health")

	agents := h.agentRegistry.ListAgents()
	healthStatus := make(map[string]interface{})

	for agentType, agent := range agents {
		status := agent.GetStatus()
		health, _ := h.agentRegistry.GetAgentHealth(agentType)

		healthStatus[agentType] = map[string]interface{}{
			"status": string(status),
			"health": health,
		}
	}

	response := map[string]interface{}{
		"agents": healthStatus,
		"status": "success",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// Analytics handlers

func (h *OrchestrationHandlers) GetDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.logger.Info("Getting dashboard data")

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	dashboardData, err := h.analyticsService.GetDashboardData(ctx)
	if err != nil {
		h.logger.Error("Failed to get dashboard data", err)
		http.Error(w, "Failed to get dashboard data", http.StatusInternalServerError)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, dashboardData)
}

func (h *OrchestrationHandlers) GetWorkflowAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	workflowID := extractIDFromPath(r.URL.Path, "/api/v1/analytics/workflows/")
	if workflowID == "" {
		http.Error(w, "Invalid workflow ID", http.StatusBadRequest)
		return
	}

	h.logger.Info("Getting workflow analytics", "workflow_id", workflowID)

	// Mock analytics response
	response := map[string]interface{}{
		"workflow_id":        workflowID,
		"total_executions":   150,
		"successful_executions": 142,
		"failed_executions":  8,
		"success_rate":       94.7,
		"average_execution_time": "45m",
		"last_execution":     time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		"throughput_per_hour": 3.2,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *OrchestrationHandlers) GetAgentAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentType := extractIDFromPath(r.URL.Path, "/api/v1/analytics/agents/")
	if agentType == "" {
		http.Error(w, "Invalid agent type", http.StatusBadRequest)
		return
	}

	h.logger.Info("Getting agent analytics", "agent_type", agentType)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	analytics, err := h.analyticsService.GetAgentMetrics(ctx, agentType)
	if err != nil {
		h.logger.Error("Failed to get agent analytics", err)
		http.Error(w, "Failed to get agent analytics", http.StatusInternalServerError)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, analytics)
}

func (h *OrchestrationHandlers) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.logger.Info("Getting system metrics")

	// Mock metrics response
	response := map[string]interface{}{
		"system": map[string]interface{}{
			"uptime":             "24h30m",
			"cpu_usage":          45.2,
			"memory_usage":       67.8,
			"active_workflows":   12,
			"running_executions": 5,
		},
		"agents": map[string]interface{}{
			"total_agents":  4,
			"online_agents": 4,
			"offline_agents": 0,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// System handlers

func (h *OrchestrationHandlers) SystemStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.logger.Info("Getting system status")

	response := map[string]interface{}{
		"status":     "operational",
		"version":    "1.0.0",
		"uptime":     "24h30m",
		"components": map[string]string{
			"workflow_engine":   "healthy",
			"agent_registry":    "healthy",
			"analytics_service": "healthy",
			"database":          "healthy",
			"cache":             "healthy",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *OrchestrationHandlers) SystemMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.logger.Info("Getting system metrics")

	response := map[string]interface{}{
		"cpu": map[string]interface{}{
			"usage_percent": 45.2,
			"cores":         8,
		},
		"memory": map[string]interface{}{
			"usage_percent": 67.8,
			"total_gb":      16,
			"used_gb":       10.8,
		},
		"disk": map[string]interface{}{
			"usage_percent": 23.4,
			"total_gb":      500,
			"used_gb":       117,
		},
		"network": map[string]interface{}{
			"bytes_in":  1024 * 1024 * 50,
			"bytes_out": 1024 * 1024 * 75,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// Workflow templates handler

func (h *OrchestrationHandlers) handleWorkflowTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.logger.Info("Getting workflow templates")

	templates := []map[string]interface{}{
		{
			"id":          "template-content-creation",
			"name":        "Content Creation Workflow",
			"description": "AI-powered content creation and publishing pipeline",
			"category":    "content",
			"complexity":  "medium",
			"estimated_duration": "30-60 minutes",
		},
		{
			"id":          "template-beverage-innovation",
			"name":        "Beverage Innovation Workflow",
			"description": "Complete beverage invention and testing pipeline",
			"category":    "innovation",
			"complexity":  "high",
			"estimated_duration": "2-4 hours",
		},
		{
			"id":          "template-feedback-analysis",
			"name":        "Feedback Analysis Workflow",
			"description": "Automated customer feedback analysis and response",
			"category":    "analytics",
			"complexity":  "low",
			"estimated_duration": "10-20 minutes",
		},
	}

	response := map[string]interface{}{
		"templates": templates,
		"total":     len(templates),
		"status":    "success",
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// Helper methods

func (h *OrchestrationHandlers) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", err)
	}
}

func extractIDFromPath(path, prefix string) string {
	if len(path) <= len(prefix) {
		return ""
	}
	
	id := path[len(prefix):]
	
	// Remove any trailing path segments
	if slashIndex := len(id); slashIndex > 0 {
		for i, char := range id {
			if char == '/' {
				slashIndex = i
				break
			}
		}
		if slashIndex < len(id) {
			id = id[:slashIndex]
		}
	}
	
	return id
}
