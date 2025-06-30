package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/internal/ai"
	"go-coffee-ai-agents/internal/httputils"
	"go-coffee-ai-agents/internal/observability"
)

// TaskHandler handles task-related HTTP requests
type TaskHandler struct {
	logger  *observability.StructuredLogger
	metrics *observability.MetricsCollector
	tracing *observability.TracingHelper
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) *TaskHandler {
	return &TaskHandler{
		logger:  logger,
		metrics: metrics,
		tracing: tracing,
	}
}

// TaskRequest represents a task creation/update request
type TaskRequest struct {
	Title        string                 `json:"title" validate:"required,min=1,max=200"`
	Description  string                 `json:"description" validate:"max=1000"`
	AssigneeID   string                 `json:"assignee_id"`
	ProjectID    string                 `json:"project_id"`
	Priority     string                 `json:"priority" validate:"required,oneof=low normal high critical"`
	Status       string                 `json:"status" validate:"required,oneof=open in_progress review done cancelled"`
	DueDate      *time.Time             `json:"due_date"`
	EstimatedTime int                   `json:"estimated_time_hours" validate:"min=0,max=1000"`
	Tags         []string               `json:"tags"`
	Dependencies []string               `json:"dependencies"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// TaskGenerateRequest represents a task generation request
type TaskGenerateRequest struct {
	Context   string    `json:"context" validate:"required,min=1,max=500"`
	Goal      string    `json:"goal" validate:"required,min=1,max=200"`
	Priority  string    `json:"priority" validate:"oneof=low normal high critical"`
	Deadline  *time.Time `json:"deadline"`
	Skills    []string  `json:"skills"`
	Resources []string  `json:"resources"`
	Count     int       `json:"count" validate:"min=1,max=20"`
}

// TaskResponse represents a task response
type TaskResponse struct {
	ID            uuid.UUID              `json:"id"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	AssigneeID    string                 `json:"assignee_id,omitempty"`
	ProjectID     string                 `json:"project_id,omitempty"`
	Priority      string                 `json:"priority"`
	Status        string                 `json:"status"`
	DueDate       *time.Time             `json:"due_date,omitempty"`
	EstimatedTime int                    `json:"estimated_time_hours"`
	ActualTime    int                    `json:"actual_time_hours"`
	Tags          []string               `json:"tags"`
	Dependencies  []string               `json:"dependencies"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	CreatedBy     string                 `json:"created_by"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	URL           string                 `json:"url"`
}

// TaskStatusUpdateRequest represents a task status update request
type TaskStatusUpdateRequest struct {
	Status    string `json:"status" validate:"required,oneof=open in_progress review done cancelled"`
	Comment   string `json:"comment" validate:"max=500"`
	ActualTime int   `json:"actual_time_hours" validate:"min=0"`
}

// List handles GET /api/v1/tasks
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "GET", "/api/v1/tasks", r.UserAgent())
	defer span.End()

	h.logger.InfoContext(ctx, "Listing tasks")

	// Parse pagination parameters
	pagination := httputils.ParsePaginationParams(r)
	sort := httputils.ParseSortParams(r, "created_at")
	filter := httputils.ParseFilterParams(r)

	h.logger.DebugContext(ctx, "Parsed request parameters",
		"page", pagination.Page,
		"per_page", pagination.PerPage,
		"sort_field", sort.Field,
		"sort_order", sort.Order,
		"search", filter.Search)

	// TODO: Implement actual task listing from database
	// For now, return mock data
	tasks := []TaskResponse{
		{
			ID:            uuid.New(),
			Title:         "Setup Mars Colony Beverage Production",
			Description:   "Establish automated beverage production system for Mars colony",
			Priority:      "high",
			Status:        "in_progress",
			EstimatedTime: 40,
			ActualTime:    15,
			Tags:          []string{"mars", "automation", "beverages"},
			Dependencies:  []string{},
			CreatedAt:     time.Now().Add(-48 * time.Hour),
			UpdatedAt:     time.Now().Add(-2 * time.Hour),
			CreatedBy:     "task-manager-agent",
			URL:           httputils.BuildResourceURL(r, "tasks", uuid.New()),
		},
	}

	total := len(tasks)

	h.tracing.RecordSuccess(span, "Tasks listed successfully")
	h.logger.InfoContext(ctx, "Tasks listed successfully",
		"count", len(tasks),
		"total", total)

	httputils.WritePaginatedResponse(w, http.StatusOK, tasks, pagination, total)
}

// Create handles POST /api/v1/tasks
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "POST", "/api/v1/tasks", r.UserAgent())
	defer span.End()

	h.logger.InfoContext(ctx, "Creating task")

	// Decode request body
	var req TaskRequest
	if err := httputils.DecodeJSONBody(r, &req); err != nil {
		h.tracing.RecordError(span, err, "Failed to decode request body")
		httputils.WriteErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// TODO: Validate request using validator
	// TODO: Create task in database

	// Create response
	taskID := uuid.New()
	response := TaskResponse{
		ID:            taskID,
		Title:         req.Title,
		Description:   req.Description,
		AssigneeID:    req.AssigneeID,
		ProjectID:     req.ProjectID,
		Priority:      req.Priority,
		Status:        req.Status,
		DueDate:       req.DueDate,
		EstimatedTime: req.EstimatedTime,
		ActualTime:    0,
		Tags:          req.Tags,
		Dependencies:  req.Dependencies,
		Metadata:      req.Metadata,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		CreatedBy:     httputils.GetUserID(r),
		URL:           httputils.BuildResourceURL(r, "tasks", taskID),
	}

	h.tracing.RecordSuccess(span, "Task created successfully")
	h.logger.InfoContext(ctx, "Task created successfully",
		"task_id", taskID,
		"title", req.Title,
		"priority", req.Priority)

	httputils.WriteJSONResponse(w, http.StatusCreated, response)
}

// Get handles GET /api/v1/tasks/{id}
func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "GET", "/api/v1/tasks/{id}", r.UserAgent())
	defer span.End()

	// Parse task ID
	taskID, err := httputils.ParseUUIDParam(r, "id")
	if err != nil {
		h.tracing.RecordError(span, err, "Invalid task ID")
		httputils.WriteErrorResponse(w, http.StatusBadRequest, "invalid_id", "Invalid task ID")
		return
	}

	h.logger.InfoContext(ctx, "Getting task", "task_id", taskID)

	// TODO: Get task from database
	// For now, return mock data
	response := TaskResponse{
		ID:            taskID,
		Title:         "Setup Mars Colony Beverage Production",
		Description:   "Establish automated beverage production system for Mars colony",
		Priority:      "high",
		Status:        "in_progress",
		EstimatedTime: 40,
		ActualTime:    15,
		Tags:          []string{"mars", "automation", "beverages"},
		Dependencies:  []string{},
		CreatedAt:     time.Now().Add(-48 * time.Hour),
		UpdatedAt:     time.Now().Add(-2 * time.Hour),
		CreatedBy:     "task-manager-agent",
		URL:           httputils.BuildResourceURL(r, "tasks", taskID),
	}

	h.tracing.RecordSuccess(span, "Task retrieved successfully")
	h.logger.InfoContext(ctx, "Task retrieved successfully", "task_id", taskID)

	httputils.WriteJSONResponse(w, http.StatusOK, response)
}

// Update handles PUT /api/v1/tasks/{id}
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "PUT", "/api/v1/tasks/{id}", r.UserAgent())
	defer span.End()

	// Parse task ID
	taskID, err := httputils.ParseUUIDParam(r, "id")
	if err != nil {
		h.tracing.RecordError(span, err, "Invalid task ID")
		httputils.WriteErrorResponse(w, http.StatusBadRequest, "invalid_id", "Invalid task ID")
		return
	}

	h.logger.InfoContext(ctx, "Updating task", "task_id", taskID)

	// Decode request body
	var req TaskRequest
	if err := httputils.DecodeJSONBody(r, &req); err != nil {
		h.tracing.RecordError(span, err, "Failed to decode request body")
		httputils.WriteErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// TODO: Update task in database

	// Create response
	response := TaskResponse{
		ID:            taskID,
		Title:         req.Title,
		Description:   req.Description,
		AssigneeID:    req.AssigneeID,
		ProjectID:     req.ProjectID,
		Priority:      req.Priority,
		Status:        req.Status,
		DueDate:       req.DueDate,
		EstimatedTime: req.EstimatedTime,
		ActualTime:    0, // TODO: Get from database
		Tags:          req.Tags,
		Dependencies:  req.Dependencies,
		Metadata:      req.Metadata,
		CreatedAt:     time.Now().Add(-48 * time.Hour), // TODO: Get from database
		UpdatedAt:     time.Now(),
		CreatedBy:     httputils.GetUserID(r),
		URL:           httputils.BuildResourceURL(r, "tasks", taskID),
	}

	h.tracing.RecordSuccess(span, "Task updated successfully")
	h.logger.InfoContext(ctx, "Task updated successfully",
		"task_id", taskID,
		"title", req.Title)

	httputils.WriteJSONResponse(w, http.StatusOK, response)
}

// Delete handles DELETE /api/v1/tasks/{id}
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "DELETE", "/api/v1/tasks/{id}", r.UserAgent())
	defer span.End()

	// Parse task ID
	taskID, err := httputils.ParseUUIDParam(r, "id")
	if err != nil {
		h.tracing.RecordError(span, err, "Invalid task ID")
		httputils.WriteErrorResponse(w, http.StatusBadRequest, "invalid_id", "Invalid task ID")
		return
	}

	h.logger.InfoContext(ctx, "Deleting task", "task_id", taskID)

	// TODO: Delete task from database

	h.tracing.RecordSuccess(span, "Task deleted successfully")
	h.logger.InfoContext(ctx, "Task deleted successfully", "task_id", taskID)

	httputils.WriteJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Task deleted successfully",
		"id":      taskID,
	})
}

// Generate handles POST /api/v1/tasks/generate
func (h *TaskHandler) Generate(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "POST", "/api/v1/tasks/generate", r.UserAgent())
	defer span.End()

	h.logger.InfoContext(ctx, "Generating tasks with AI")

	// Decode request body
	var req TaskGenerateRequest
	if err := httputils.DecodeJSONBody(r, &req); err != nil {
		h.tracing.RecordError(span, err, "Failed to decode request body")
		httputils.WriteErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Set defaults
	if req.Count == 0 {
		req.Count = 3
	}
	if req.Priority == "" {
		req.Priority = "normal"
	}

	h.logger.InfoContext(ctx, "AI task generation request",
		"context", req.Context,
		"goal", req.Goal,
		"count", req.Count,
		"priority", req.Priority)

	// Generate tasks using AI
	aiManager := ai.GetGlobalManager()
	if aiManager == nil {
		h.tracing.RecordError(span, nil, "AI manager not available")
		httputils.WriteErrorResponse(w, http.StatusServiceUnavailable, "ai_unavailable", "AI service is not available")
		return
	}

	// TODO: Implement task generation in AI manager
	// aiRequest := &ai.TaskGenerationRequest{
	// 	Context:   req.Context,
	// 	Goal:      req.Goal,
	// 	Priority:  req.Priority,
	// 	Deadline:  req.Deadline,
	// 	Skills:    req.Skills,
	// 	Resources: req.Resources,
	// 	Count:     req.Count,
	// }

	// For now, create mock tasks
	tasks := make([]TaskResponse, req.Count)
	for i := 0; i < req.Count; i++ {
		taskID := uuid.New()
		tasks[i] = TaskResponse{
			ID:            taskID,
			Title:         "AI Generated Task " + string(rune(i+1)),
			Description:   "This task was generated by AI based on the provided context and goal",
			Priority:      req.Priority,
			Status:        "open",
			EstimatedTime: 8,
			ActualTime:    0,
			Tags:          []string{"ai-generated"},
			Dependencies:  []string{},
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
			CreatedBy:     "ai-generator",
			URL:           httputils.BuildResourceURL(r, "tasks", taskID),
		}
	}

	response := map[string]interface{}{
		"tasks": tasks,
		"metadata": map[string]interface{}{
			"ai_provider":  "openai", // TODO: Get from actual AI response
			"ai_model":     "gpt-4",  // TODO: Get from actual AI response
			"generated_at": time.Now(),
		},
	}

	h.tracing.RecordSuccess(span, "Tasks generated successfully")
	h.logger.InfoContext(ctx, "Tasks generated successfully",
		"count", len(tasks),
		"context", req.Context)

	httputils.WriteJSONResponse(w, http.StatusOK, response)
}

// UpdateStatus handles PUT /api/v1/tasks/{id}/status
func (h *TaskHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "PUT", "/api/v1/tasks/{id}/status", r.UserAgent())
	defer span.End()

	// Parse task ID
	taskID, err := httputils.ParseUUIDParam(r, "id")
	if err != nil {
		h.tracing.RecordError(span, err, "Invalid task ID")
		httputils.WriteErrorResponse(w, http.StatusBadRequest, "invalid_id", "Invalid task ID")
		return
	}

	h.logger.InfoContext(ctx, "Updating task status", "task_id", taskID)

	// Decode request body
	var req TaskStatusUpdateRequest
	if err := httputils.DecodeJSONBody(r, &req); err != nil {
		h.tracing.RecordError(span, err, "Failed to decode request body")
		httputils.WriteErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// TODO: Update task status in database

	var completedAt *time.Time
	if req.Status == "done" {
		now := time.Now()
		completedAt = &now
	}

	response := map[string]interface{}{
		"id":           taskID,
		"status":       req.Status,
		"comment":      req.Comment,
		"actual_time":  req.ActualTime,
		"completed_at": completedAt,
		"updated_at":   time.Now(),
		"updated_by":   httputils.GetUserID(r),
	}

	h.tracing.RecordSuccess(span, "Task status updated successfully")
	h.logger.InfoContext(ctx, "Task status updated successfully",
		"task_id", taskID,
		"status", req.Status)

	httputils.WriteJSONResponse(w, http.StatusOK, response)
}

// Statistics handles GET /api/v1/tasks/stats
func (h *TaskHandler) Statistics(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracing.StartHTTPSpan(r.Context(), "GET", "/api/v1/tasks/stats", r.UserAgent())
	defer span.End()

	h.logger.InfoContext(ctx, "Getting task statistics")

	// TODO: Implement actual statistics
	stats := map[string]interface{}{
		"total_tasks": 156,
		"status_distribution": map[string]int{
			"open":        45,
			"in_progress": 32,
			"review":      12,
			"done":        62,
			"cancelled":   5,
		},
		"priority_distribution": map[string]int{
			"low":      23,
			"normal":   89,
			"high":     35,
			"critical": 9,
		},
		"completion_rate":     0.75,
		"average_completion_time": 18.5,
		"tasks_created_today": 8,
		"tasks_completed_today": 12,
		"last_updated":        time.Now(),
	}

	h.tracing.RecordSuccess(span, "Task statistics retrieved")
	h.logger.InfoContext(ctx, "Task statistics retrieved")

	httputils.WriteJSONResponse(w, http.StatusOK, stats)
}
