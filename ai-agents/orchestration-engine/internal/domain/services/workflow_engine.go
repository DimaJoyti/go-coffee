package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/orchestration-engine/internal/domain/entities"
)

// WorkflowEngine manages workflow execution and coordination
type WorkflowEngine struct {
	workflowRepo     WorkflowRepository
	executionRepo    ExecutionRepository
	agentRegistry    AgentRegistry
	eventPublisher   EventPublisher
	logger           Logger
	executors        map[uuid.UUID]*WorkflowExecutor
	executorsMutex   sync.RWMutex
	stopChan         chan struct{}
	maxConcurrency   int
	semaphore        chan struct{}
}

// WorkflowRepository defines the interface for workflow data access
type WorkflowRepository interface {
	Create(ctx context.Context, workflow *entities.Workflow) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Workflow, error)
	Update(ctx context.Context, workflow *entities.Workflow) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *WorkflowFilter) ([]*entities.Workflow, error)
	GetActiveWorkflows(ctx context.Context) ([]*entities.Workflow, error)
	GetWorkflowsByTrigger(ctx context.Context, triggerType entities.TriggerType) ([]*entities.Workflow, error)
}

// ExecutionRepository defines the interface for execution data access
type ExecutionRepository interface {
	Create(ctx context.Context, execution *entities.WorkflowExecution) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.WorkflowExecution, error)
	Update(ctx context.Context, execution *entities.WorkflowExecution) error
	List(ctx context.Context, filter *ExecutionFilter) ([]*entities.WorkflowExecution, error)
	GetActiveExecutions(ctx context.Context, workflowID uuid.UUID) ([]*entities.WorkflowExecution, error)
	GetExecutionHistory(ctx context.Context, workflowID uuid.UUID, limit int) ([]*entities.WorkflowExecution, error)
}

// AgentRegistry defines the interface for agent management
type AgentRegistry interface {
	RegisterAgent(agentType string, agent Agent) error
	GetAgent(agentType string) (Agent, error)
	ListAgents() map[string]Agent
	IsAgentAvailable(agentType string) bool
	GetAgentHealth(agentType string) (*AgentHealth, error)
}

// Agent defines the interface for workflow agents
type Agent interface {
	Execute(ctx context.Context, action string, input map[string]interface{}) (map[string]interface{}, error)
	GetCapabilities() []string
	GetStatus() AgentStatus
	Validate(action string, input map[string]interface{}) error
	GetMetrics() *AgentMetrics
}

// EventPublisher defines the interface for publishing workflow events
type EventPublisher interface {
	PublishWorkflowEvent(ctx context.Context, event *WorkflowEvent) error
	PublishExecutionEvent(ctx context.Context, event *ExecutionEvent) error
	PublishStepEvent(ctx context.Context, event *StepEvent) error
}

// Logger defines the interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// Supporting types

type WorkflowFilter struct {
	Status     []entities.WorkflowStatus   `json:"status,omitempty"`
	Type       []entities.WorkflowType     `json:"type,omitempty"`
	Category   []entities.WorkflowCategory `json:"category,omitempty"`
	IsActive   *bool                       `json:"is_active,omitempty"`
	IsTemplate *bool                       `json:"is_template,omitempty"`
	CreatedBy  []uuid.UUID                 `json:"created_by,omitempty"`
	Tags       []string                    `json:"tags,omitempty"`
	Limit      int                         `json:"limit,omitempty"`
	Offset     int                         `json:"offset,omitempty"`
}

type ExecutionFilter struct {
	WorkflowID []uuid.UUID                `json:"workflow_id,omitempty"`
	Status     []entities.WorkflowStatus  `json:"status,omitempty"`
	CreatedBy  []uuid.UUID                `json:"created_by,omitempty"`
	StartedAfter  *time.Time              `json:"started_after,omitempty"`
	StartedBefore *time.Time              `json:"started_before,omitempty"`
	Limit      int                        `json:"limit,omitempty"`
	Offset     int                        `json:"offset,omitempty"`
}

type AgentStatus string

const (
	AgentStatusOnline  AgentStatus = "online"
	AgentStatusOffline AgentStatus = "offline"
	AgentStatusBusy    AgentStatus = "busy"
	AgentStatusError   AgentStatus = "error"
)

type AgentHealth struct {
	Status      AgentStatus `json:"status"`
	LastSeen    time.Time   `json:"last_seen"`
	ResponseTime time.Duration `json:"response_time"`
	ErrorRate   float64     `json:"error_rate"`
	Load        float64     `json:"load"`
}

type AgentMetrics struct {
	TotalRequests    int64         `json:"total_requests"`
	SuccessfulRequests int64       `json:"successful_requests"`
	FailedRequests   int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	CurrentLoad      float64       `json:"current_load"`
	LastUpdated      time.Time     `json:"last_updated"`
}

// Event types

type WorkflowEvent struct {
	ID         uuid.UUID                  `json:"id"`
	Type       string                     `json:"type"`
	WorkflowID uuid.UUID                  `json:"workflow_id"`
	Data       map[string]interface{}     `json:"data"`
	Timestamp  time.Time                  `json:"timestamp"`
	Source     string                     `json:"source"`
}

type ExecutionEvent struct {
	ID          uuid.UUID                  `json:"id"`
	Type        string                     `json:"type"`
	WorkflowID  uuid.UUID                  `json:"workflow_id"`
	ExecutionID uuid.UUID                  `json:"execution_id"`
	Data        map[string]interface{}     `json:"data"`
	Timestamp   time.Time                  `json:"timestamp"`
	Source      string                     `json:"source"`
}

type StepEvent struct {
	ID          uuid.UUID                  `json:"id"`
	Type        string                     `json:"type"`
	WorkflowID  uuid.UUID                  `json:"workflow_id"`
	ExecutionID uuid.UUID                  `json:"execution_id"`
	StepID      string                     `json:"step_id"`
	Data        map[string]interface{}     `json:"data"`
	Timestamp   time.Time                  `json:"timestamp"`
	Source      string                     `json:"source"`
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine(
	workflowRepo WorkflowRepository,
	executionRepo ExecutionRepository,
	agentRegistry AgentRegistry,
	eventPublisher EventPublisher,
	logger Logger,
	maxConcurrency int,
) *WorkflowEngine {
	return &WorkflowEngine{
		workflowRepo:   workflowRepo,
		executionRepo:  executionRepo,
		agentRegistry:  agentRegistry,
		eventPublisher: eventPublisher,
		logger:         logger,
		executors:      make(map[uuid.UUID]*WorkflowExecutor),
		stopChan:       make(chan struct{}),
		maxConcurrency: maxConcurrency,
		semaphore:      make(chan struct{}, maxConcurrency),
	}
}

// Start starts the workflow engine
func (we *WorkflowEngine) Start(ctx context.Context) error {
	we.logger.Info("Starting workflow engine", "max_concurrency", we.maxConcurrency)

	// Start monitoring active executions
	go we.monitorExecutions(ctx)

	// Start processing scheduled workflows
	go we.processScheduledWorkflows(ctx)

	we.logger.Info("Workflow engine started successfully")
	return nil
}

// Stop stops the workflow engine
func (we *WorkflowEngine) Stop(ctx context.Context) error {
	we.logger.Info("Stopping workflow engine")

	close(we.stopChan)

	// Stop all active executors
	we.executorsMutex.Lock()
	defer we.executorsMutex.Unlock()

	for executionID, executor := range we.executors {
		we.logger.Info("Stopping executor", "execution_id", executionID)
		executor.Stop(ctx)
	}

	we.logger.Info("Workflow engine stopped")
	return nil
}

// ExecuteWorkflow starts a new workflow execution
func (we *WorkflowEngine) ExecuteWorkflow(ctx context.Context, workflowID uuid.UUID, input map[string]interface{}, triggerType string, createdBy uuid.UUID) (*entities.WorkflowExecution, error) {
	we.logger.Info("Starting workflow execution", "workflow_id", workflowID, "trigger_type", triggerType)

	// Get workflow
	workflow, err := we.workflowRepo.GetByID(ctx, workflowID)
	if err != nil {
		we.logger.Error("Failed to get workflow", err, "workflow_id", workflowID)
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	// Check if workflow can be executed
	if !workflow.CanExecute() {
		return nil, fmt.Errorf("workflow cannot be executed: status=%s, active=%t", workflow.Status, workflow.IsActive)
	}

	// Create execution
	execution := &entities.WorkflowExecution{
		ID:          uuid.New(),
		WorkflowID:  workflowID,
		Status:      entities.WorkflowStatusRunning,
		TriggerType: triggerType,
		Input:       input,
		Variables:   make(map[string]interface{}),
		StartedAt:   time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   createdBy,
	}

	// Initialize variables
	if workflow.Variables != nil {
		for k, v := range workflow.Variables {
			execution.Variables[k] = v
		}
	}
	if input != nil {
		for k, v := range input {
			execution.Variables[k] = v
		}
	}

	// Save execution
	if err := we.executionRepo.Create(ctx, execution); err != nil {
		we.logger.Error("Failed to create execution", err, "execution_id", execution.ID)
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	// Acquire semaphore for concurrency control
	select {
	case we.semaphore <- struct{}{}:
		// Continue
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Create and start executor
	executor := NewWorkflowExecutor(
		workflow,
		execution,
		we.agentRegistry,
		we.executionRepo,
		we.eventPublisher,
		we.logger,
	)

	we.executorsMutex.Lock()
	we.executors[execution.ID] = executor
	we.executorsMutex.Unlock()

	// Start execution in goroutine
	go func() {
		defer func() {
			// Release semaphore
			<-we.semaphore

			// Remove executor
			we.executorsMutex.Lock()
			delete(we.executors, execution.ID)
			we.executorsMutex.Unlock()
		}()

		if err := executor.Execute(ctx); err != nil {
			we.logger.Error("Workflow execution failed", err, "execution_id", execution.ID)
		}
	}()

	// Publish execution started event
	event := &ExecutionEvent{
		ID:          uuid.New(),
		Type:        "execution.started",
		WorkflowID:  workflowID,
		ExecutionID: execution.ID,
		Data: map[string]interface{}{
			"trigger_type": triggerType,
			"input":        input,
		},
		Timestamp: time.Now(),
		Source:    "workflow-engine",
	}
	we.eventPublisher.PublishExecutionEvent(ctx, event)

	we.logger.Info("Workflow execution started", "execution_id", execution.ID)
	return execution, nil
}

// GetExecution retrieves a workflow execution
func (we *WorkflowEngine) GetExecution(ctx context.Context, executionID uuid.UUID) (*entities.WorkflowExecution, error) {
	return we.executionRepo.GetByID(ctx, executionID)
}

// CancelExecution cancels a running workflow execution
func (we *WorkflowEngine) CancelExecution(ctx context.Context, executionID uuid.UUID) error {
	we.logger.Info("Cancelling workflow execution", "execution_id", executionID)

	we.executorsMutex.RLock()
	executor, exists := we.executors[executionID]
	we.executorsMutex.RUnlock()

	if exists {
		return executor.Cancel(ctx)
	}

	// If executor doesn't exist, update execution status directly
	execution, err := we.executionRepo.GetByID(ctx, executionID)
	if err != nil {
		return fmt.Errorf("failed to get execution: %w", err)
	}

	execution.Status = entities.WorkflowStatusCancelled
	execution.CompletedAt = &[]time.Time{time.Now()}[0]
	execution.UpdatedAt = time.Now()

	return we.executionRepo.Update(ctx, execution)
}

// ListExecutions lists workflow executions
func (we *WorkflowEngine) ListExecutions(ctx context.Context, filter *ExecutionFilter) ([]*entities.WorkflowExecution, error) {
	return we.executionRepo.List(ctx, filter)
}

// GetWorkflowMetrics retrieves metrics for a workflow
func (we *WorkflowEngine) GetWorkflowMetrics(ctx context.Context, workflowID uuid.UUID) (*entities.WorkflowMetrics, error) {
	executions, err := we.executionRepo.List(ctx, &ExecutionFilter{
		WorkflowID: []uuid.UUID{workflowID},
		Limit:      1000, // Get recent executions for metrics
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get executions: %w", err)
	}

	return we.calculateMetrics(executions), nil
}

// monitorExecutions monitors active executions for timeouts and health
func (we *WorkflowEngine) monitorExecutions(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-we.stopChan:
			return
		case <-ticker.C:
			we.checkExecutionTimeouts(ctx)
			we.checkExecutorHealth(ctx)
		}
	}
}

// processScheduledWorkflows processes workflows with schedule triggers
func (we *WorkflowEngine) processScheduledWorkflows(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-we.stopChan:
			return
		case <-ticker.C:
			we.processSchedules(ctx)
		}
	}
}

// checkExecutionTimeouts checks for timed out executions
func (we *WorkflowEngine) checkExecutionTimeouts(ctx context.Context) {
	we.executorsMutex.RLock()
	defer we.executorsMutex.RUnlock()

	for executionID, executor := range we.executors {
		if executor.IsTimedOut() {
			we.logger.Warn("Execution timed out", "execution_id", executionID)
			go executor.Timeout(ctx)
		}
	}
}

// checkExecutorHealth checks the health of active executors
func (we *WorkflowEngine) checkExecutorHealth(ctx context.Context) {
	we.executorsMutex.RLock()
	defer we.executorsMutex.RUnlock()

	for executionID, executor := range we.executors {
		if !executor.IsHealthy() {
			we.logger.Warn("Executor unhealthy", "execution_id", executionID)
			// Could implement recovery logic here
		}
	}
}

// processSchedules processes scheduled workflow triggers
func (we *WorkflowEngine) processSchedules(ctx context.Context) {
	workflows, err := we.workflowRepo.GetWorkflowsByTrigger(ctx, entities.TriggerTypeSchedule)
	if err != nil {
		we.logger.Error("Failed to get scheduled workflows", err)
		return
	}

	for _, workflow := range workflows {
		we.processWorkflowSchedules(ctx, workflow)
	}
}

// processWorkflowSchedules processes schedules for a specific workflow
func (we *WorkflowEngine) processWorkflowSchedules(ctx context.Context, workflow *entities.Workflow) {
	for _, trigger := range workflow.Triggers {
		if trigger.Type == entities.TriggerTypeSchedule && trigger.IsActive {
			if we.shouldTriggerSchedule(trigger) {
				we.logger.Info("Triggering scheduled workflow", "workflow_id", workflow.ID, "trigger_id", trigger.ID)
				
				_, err := we.ExecuteWorkflow(ctx, workflow.ID, nil, "schedule", workflow.CreatedBy)
				if err != nil {
					we.logger.Error("Failed to execute scheduled workflow", err, "workflow_id", workflow.ID)
				} else {
					trigger.LastTriggered = &[]time.Time{time.Now()}[0]
					trigger.TriggerCount++
				}
			}
		}
	}
}

// shouldTriggerSchedule determines if a schedule trigger should fire
func (we *WorkflowEngine) shouldTriggerSchedule(trigger *entities.WorkflowTrigger) bool {
	// Implement schedule logic based on trigger configuration
	// This is a simplified implementation
	if trigger.LastTriggered == nil {
		return true
	}

	// Check if enough time has passed based on schedule configuration
	// This would be more sophisticated in a real implementation
	return time.Since(*trigger.LastTriggered) > time.Hour
}

// calculateMetrics calculates workflow metrics from executions
func (we *WorkflowEngine) calculateMetrics(executions []*entities.WorkflowExecution) *entities.WorkflowMetrics {
	if len(executions) == 0 {
		return &entities.WorkflowMetrics{
			LastUpdated: time.Now(),
		}
	}

	var totalExecutions, successful, failed int64
	var totalDuration time.Duration
	var lastExecution *time.Time

	for _, exec := range executions {
		totalExecutions++
		
		if exec.Status == entities.WorkflowStatusCompleted {
			successful++
		} else if exec.Status == entities.WorkflowStatusFailed {
			failed++
		}

		totalDuration += exec.Duration
		
		if lastExecution == nil || exec.StartedAt.After(*lastExecution) {
			lastExecution = &exec.StartedAt
		}
	}

	successRate := float64(successful) / float64(totalExecutions) * 100
	errorRate := float64(failed) / float64(totalExecutions) * 100
	avgExecutionTime := totalDuration / time.Duration(totalExecutions)

	return &entities.WorkflowMetrics{
		TotalExecutions:      totalExecutions,
		SuccessfulExecutions: successful,
		FailedExecutions:     failed,
		AverageExecutionTime: avgExecutionTime,
		LastExecutionTime:    lastExecution,
		SuccessRate:          successRate,
		ErrorRate:            errorRate,
		LastUpdated:          time.Now(),
	}
}
