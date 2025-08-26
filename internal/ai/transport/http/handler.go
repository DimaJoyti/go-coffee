package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/ai/orchestrator"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for the AI orchestrator
type Handler struct {
	orchestrator *orchestrator.Orchestrator
	logger       *zap.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(orchestrator *orchestrator.Orchestrator, logger *zap.Logger) *Handler {
	return &Handler{
		orchestrator: orchestrator,
		logger:       logger,
	}
}

// ListAgents handles agent listing requests
func (h *Handler) ListAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Mock response - in production, this would get actual agents
	agents := []map[string]interface{}{
		{
			"id":           "beverage-inventor",
			"type":         "beverage-inventor",
			"status":       "online",
			"capabilities": []string{"analyze_order", "create_recipe", "suggest_modifications"},
			"health":       true,
		},
		{
			"id":           "inventory-manager",
			"type":         "inventory-manager",
			"status":       "online",
			"capabilities": []string{"check_availability", "forecast_demand", "manage_stock"},
			"health":       true,
		},
		{
			"id":           "task-manager",
			"type":         "task-manager",
			"status":       "online",
			"capabilities": []string{"create_task", "assign_task", "track_progress"},
			"health":       true,
		},
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"agents": agents,
		"total":  len(agents),
	})
}

// GetAgent handles individual agent requests
func (h *Handler) GetAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract agent ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/agents/")
	agentID := strings.Split(path, "/")[0]

	if agentID == "" {
		h.writeError(w, http.StatusBadRequest, "Agent ID is required")
		return
	}

	// Mock response
	agent := map[string]interface{}{
		"id":           agentID,
		"type":         agentID,
		"status":       "online",
		"capabilities": []string{"mock_capability"},
		"health":       true,
		"last_seen":    time.Now().UTC().Format(time.RFC3339),
	}

	h.writeJSON(w, http.StatusOK, agent)
}

// RegisterAgent handles agent registration requests
func (h *Handler) RegisterAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		ID           string   `json:"id"`
		Type         string   `json:"type"`
		Capabilities []string `json:"capabilities"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]string{
		"status":   "registered",
		"agent_id": req.ID,
	})
}

// UnregisterAgent handles agent unregistration requests
func (h *Handler) UnregisterAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		AgentID string `json:"agent_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"status":   "unregistered",
		"agent_id": req.AgentID,
	})
}

// ListWorkflows handles workflow listing requests
func (h *Handler) ListWorkflows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	workflows := h.orchestrator.ListWorkflows()
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"workflows": workflows,
		"total":     len(workflows),
	})
}

// GetWorkflow handles individual workflow requests
func (h *Handler) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract workflow ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/workflows/")
	workflowID := strings.Split(path, "/")[0]

	if workflowID == "" {
		h.writeError(w, http.StatusBadRequest, "Workflow ID is required")
		return
	}

	workflow, err := h.orchestrator.GetWorkflow(workflowID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Workflow not found")
		return
	}

	h.writeJSON(w, http.StatusOK, workflow)
}

// CreateWorkflow handles workflow creation requests
func (h *Handler) CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var workflow orchestrator.Workflow
	if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.orchestrator.CreateWorkflow(&workflow); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to create workflow")
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"status":      "created",
		"workflow_id": workflow.ID,
	})
}

// ExecuteWorkflow handles workflow execution requests
func (h *Handler) ExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		WorkflowID string `json:"workflow_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.orchestrator.ExecuteWorkflow(req.WorkflowID); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"status":      "executing",
		"workflow_id": req.WorkflowID,
	})
}

// StopWorkflow handles workflow stop requests
func (h *Handler) StopWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		WorkflowID string `json:"workflow_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"status":      "stopped",
		"workflow_id": req.WorkflowID,
	})
}

// ListTasks handles task listing requests
func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	tasks := h.orchestrator.ListTasks()
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"tasks": tasks,
		"total": len(tasks),
	})
}

// GetTask handles individual task requests
func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract task ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	taskID := strings.Split(path, "/")[0]

	if taskID == "" {
		h.writeError(w, http.StatusBadRequest, "Task ID is required")
		return
	}

	task, err := h.orchestrator.GetTask(taskID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Task not found")
		return
	}

	h.writeJSON(w, http.StatusOK, task)
}

// AssignTask handles task assignment requests
func (h *Handler) AssignTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var task orchestrator.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.orchestrator.AssignTask(&task); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to assign task")
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"status":  "assigned",
		"task_id": task.ID,
	})
}

// CompleteTask handles task completion requests
func (h *Handler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		TaskID  string                 `json:"task_id"`
		Outputs map[string]interface{} `json:"outputs"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"status":  "completed",
		"task_id": req.TaskID,
	})
}

// SendMessage handles message sending requests
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var message orchestrator.AgentMessage
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.orchestrator.SendMessage(&message); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"status":     "sent",
		"message_id": message.ID,
	})
}

// BroadcastMessage handles message broadcasting requests
func (h *Handler) BroadcastMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Type    string                 `json:"type"`
		Content map[string]interface{} `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"status": "broadcasted",
		"type":   req.Type,
	})
}

// External integration handlers (placeholders)
func (h *Handler) ClickUpIntegration(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"integration": "clickup", "status": "connected"})
}

func (h *Handler) SlackIntegration(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"integration": "slack", "status": "connected"})
}

func (h *Handler) GoogleSheetsIntegration(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"integration": "google_sheets", "status": "connected"})
}

func (h *Handler) AirtableIntegration(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"integration": "airtable", "status": "connected"})
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "ok",
		"service":   "ai-orchestrator",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ReadinessCheck handles readiness check requests
func (h *Handler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "ready",
		"service":   "ai-orchestrator",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks": map[string]string{
			"kafka":  "ok",
			"agents": "ok",
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// writeJSON writes a JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}
