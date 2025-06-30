package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/task-manager-agent/internal/domain/entities"
	"go-coffee-ai-agents/task-manager-agent/internal/domain/events"
	"go-coffee-ai-agents/task-manager-agent/internal/domain/repositories"
)

// WorkflowOrchestrationService provides workflow execution and orchestration
type WorkflowOrchestrationService struct {
	workflowRepo     repositories.WorkflowRepository
	taskRepo         repositories.TaskRepository
	userRepo         repositories.UserRepository
	notificationRepo repositories.NotificationRepository
	eventPublisher   EventPublisher
	actionExecutor   ActionExecutor
	conditionEvaluator ConditionEvaluator
	logger           Logger
}

// ActionExecutor defines the interface for executing workflow actions
type ActionExecutor interface {
	ExecuteAction(ctx context.Context, action *entities.WorkflowAction, context map[string]interface{}) (map[string]interface{}, error)
	CanExecute(actionType entities.ActionType) bool
	GetSupportedActions() []entities.ActionType
}

// ConditionEvaluator defines the interface for evaluating workflow conditions
type ConditionEvaluator interface {
	EvaluateCondition(ctx context.Context, condition *entities.WorkflowCondition, context map[string]interface{}) (bool, error)
	EvaluateConditions(ctx context.Context, conditions []*entities.WorkflowCondition, logic entities.ConditionLogic, context map[string]interface{}) (bool, error)
}

// WorkflowExecutionContext represents the context for workflow execution
type WorkflowExecutionContext struct {
	WorkflowID    uuid.UUID              `json:"workflow_id"`
	ExecutionID   uuid.UUID              `json:"execution_id"`
	TriggerData   map[string]interface{} `json:"trigger_data"`
	Variables     map[string]interface{} `json:"variables"`
	CurrentStep   *uuid.UUID             `json:"current_step,omitempty"`
	ExecutedBy    uuid.UUID              `json:"executed_by"`
	StartedAt     time.Time              `json:"started_at"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// StepExecutionResult represents the result of step execution
type StepExecutionResult struct {
	Success      bool                   `json:"success"`
	Output       map[string]interface{} `json:"output"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	NextSteps    []uuid.UUID            `json:"next_steps"`
	ShouldRetry  bool                   `json:"should_retry"`
	RetryDelay   time.Duration          `json:"retry_delay"`
}

// NewWorkflowOrchestrationService creates a new workflow orchestration service
func NewWorkflowOrchestrationService(
	workflowRepo repositories.WorkflowRepository,
	taskRepo repositories.TaskRepository,
	userRepo repositories.UserRepository,
	notificationRepo repositories.NotificationRepository,
	eventPublisher EventPublisher,
	actionExecutor ActionExecutor,
	conditionEvaluator ConditionEvaluator,
	logger Logger,
) *WorkflowOrchestrationService {
	return &WorkflowOrchestrationService{
		workflowRepo:       workflowRepo,
		taskRepo:           taskRepo,
		userRepo:           userRepo,
		notificationRepo:   notificationRepo,
		eventPublisher:     eventPublisher,
		actionExecutor:     actionExecutor,
		conditionEvaluator: conditionEvaluator,
		logger:             logger,
	}
}

// StartWorkflow starts a workflow execution
func (wos *WorkflowOrchestrationService) StartWorkflow(ctx context.Context, workflowID, executedBy uuid.UUID, triggerData map[string]interface{}) (*entities.WorkflowExecution, error) {
	wos.logger.Info("Starting workflow", "workflow_id", workflowID, "executed_by", executedBy)

	// Get workflow
	workflow, err := wos.workflowRepo.GetByID(ctx, workflowID)
	if err != nil {
		wos.logger.Error("Failed to get workflow", err, "workflow_id", workflowID)
		return nil, err
	}

	// Check if workflow can be executed
	if !workflow.CanExecute() {
		return nil, fmt.Errorf("workflow cannot be executed: %s", workflow.Status)
	}

	// Create execution
	execution := entities.NewWorkflowExecution(workflowID, executedBy, triggerData)
	execution.Variables = make(map[string]interface{})
	
	// Copy workflow variables to execution
	for k, v := range workflow.Variables {
		execution.Variables[k] = v
	}

	// Start execution
	execution.Start()

	// Save execution
	if err := wos.workflowRepo.CreateExecution(ctx, execution); err != nil {
		wos.logger.Error("Failed to create workflow execution", err, "workflow_id", workflowID)
		return nil, err
	}

	// Publish workflow started event
	event := events.NewWorkflowExecutionStartedEvent(workflow, execution)
	if err := wos.eventPublisher.PublishEvent(ctx, event); err != nil {
		wos.logger.Error("Failed to publish workflow started event", err, "execution_id", execution.ID)
	}

	// Start executing the workflow asynchronously
	go func() {
		if err := wos.executeWorkflow(context.Background(), workflow, execution); err != nil {
			wos.logger.Error("Workflow execution failed", err, "execution_id", execution.ID)
		}
	}()

	wos.logger.Info("Workflow started", "workflow_id", workflowID, "execution_id", execution.ID)
	return execution, nil
}

// executeWorkflow executes a workflow from start to finish
func (wos *WorkflowOrchestrationService) executeWorkflow(ctx context.Context, workflow *entities.Workflow, execution *entities.WorkflowExecution) error {
	wos.logger.Info("Executing workflow", "workflow_id", workflow.ID, "execution_id", execution.ID)

	// Create execution context
	execCtx := &WorkflowExecutionContext{
		WorkflowID:  workflow.ID,
		ExecutionID: execution.ID,
		TriggerData: execution.Context,
		Variables:   execution.Variables,
		ExecutedBy:  execution.ExecutedBy,
		StartedAt:   execution.StartedAt,
		Metadata:    make(map[string]interface{}),
	}

	// Get first step
	firstStep := workflow.GetFirstStep()
	if firstStep == nil {
		return wos.completeExecution(ctx, workflow, execution, "No steps to execute")
	}

	// Execute steps
	currentSteps := []*entities.WorkflowStep{firstStep}
	
	for len(currentSteps) > 0 {
		var nextSteps []*entities.WorkflowStep
		
		for _, step := range currentSteps {
			result, err := wos.executeStep(ctx, workflow, execution, step, execCtx)
			if err != nil {
				return wos.failExecution(ctx, workflow, execution, fmt.Sprintf("Step execution failed: %v", err))
			}
			
			if !result.Success {
				if result.ShouldRetry {
					// Handle retry logic
					wos.logger.Info("Step failed, will retry", "step_id", step.ID, "execution_id", execution.ID)
					time.Sleep(result.RetryDelay)
					continue
				} else {
					return wos.failExecution(ctx, workflow, execution, result.ErrorMessage)
				}
			}
			
			// Merge output into execution variables
			for k, v := range result.Output {
				execCtx.Variables[k] = v
				execution.Variables[k] = v
			}
			
			// Get next steps
			for _, nextStepID := range result.NextSteps {
				nextStep := workflow.GetStepByID(nextStepID)
				if nextStep != nil {
					nextSteps = append(nextSteps, nextStep)
				}
			}
		}
		
		currentSteps = nextSteps
	}

	return wos.completeExecution(ctx, workflow, execution, "")
}

// executeStep executes a single workflow step
func (wos *WorkflowOrchestrationService) executeStep(ctx context.Context, workflow *entities.Workflow, execution *entities.WorkflowExecution, step *entities.WorkflowStep, execCtx *WorkflowExecutionContext) (*StepExecutionResult, error) {
	wos.logger.Info("Executing step", "step_id", step.ID, "step_name", step.Name, "execution_id", execution.ID)

	// Create step execution record
	stepExecution := &entities.StepExecution{
		ID:          uuid.New(),
		ExecutionID: execution.ID,
		StepID:      step.ID,
		Step:        step,
		Status:      entities.StepStatusRunning,
		Input:       execCtx.Variables,
		StartedAt:   time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save step execution
	if err := wos.workflowRepo.CreateStepExecution(ctx, stepExecution); err != nil {
		wos.logger.Error("Failed to create step execution", err, "step_id", step.ID)
		return nil, err
	}

	// Update current step in execution
	execution.CurrentStep = &step.ID
	if err := wos.workflowRepo.UpdateExecution(ctx, execution); err != nil {
		wos.logger.Error("Failed to update execution current step", err, "execution_id", execution.ID)
	}

	result := &StepExecutionResult{
		Success: true,
		Output:  make(map[string]interface{}),
	}

	// Evaluate step conditions
	if len(step.Conditions) > 0 {
		// Convert StepConditions to WorkflowConditions
		workflowConditions := wos.convertStepConditionsToWorkflowConditions(step.Conditions, step.WorkflowID)
		
		conditionsMet, err := wos.conditionEvaluator.EvaluateConditions(ctx, workflowConditions, entities.LogicAnd, execCtx.Variables)
		if err != nil {
			wos.logger.Error("Failed to evaluate step conditions", err, "step_id", step.ID)
			return wos.failStep(ctx, stepExecution, fmt.Sprintf("Condition evaluation failed: %v", err))
		}
		
		if !conditionsMet {
			wos.logger.Info("Step conditions not met, skipping", "step_id", step.ID)
			stepExecution.Status = entities.StepStatusSkipped
			stepExecution.CompletedAt = &[]time.Time{time.Now()}[0]
			wos.workflowRepo.UpdateStepExecution(ctx, stepExecution)
			
			// Return next steps
			result.NextSteps = step.NextSteps
			return result, nil
		}
	}

	// Execute step based on type
	switch step.Type {
	case entities.StepTypeTask:
		err := wos.executeTaskStep(ctx, step, execCtx, result)
		if err != nil {
			return wos.failStep(ctx, stepExecution, err.Error())
		}
		
	case entities.StepTypeApproval:
		err := wos.executeApprovalStep(ctx, step, execCtx, result)
		if err != nil {
			return wos.failStep(ctx, stepExecution, err.Error())
		}
		
	case entities.StepTypeNotification:
		err := wos.executeNotificationStep(ctx, step, execCtx, result)
		if err != nil {
			return wos.failStep(ctx, stepExecution, err.Error())
		}
		
	case entities.StepTypeAction:
		err := wos.executeActionStep(ctx, step, execCtx, result)
		if err != nil {
			return wos.failStep(ctx, stepExecution, err.Error())
		}
		
	case entities.StepTypeWait:
		err := wos.executeWaitStep(ctx, step, execCtx, result)
		if err != nil {
			return wos.failStep(ctx, stepExecution, err.Error())
		}
		
	default:
		return wos.failStep(ctx, stepExecution, fmt.Sprintf("Unsupported step type: %s", step.Type))
	}

	// Complete step execution
	stepExecution.Status = entities.StepStatusCompleted
	stepExecution.Output = result.Output
	now := time.Now()
	stepExecution.CompletedAt = &now
	stepExecution.UpdatedAt = now

	if err := wos.workflowRepo.UpdateStepExecution(ctx, stepExecution); err != nil {
		wos.logger.Error("Failed to update step execution", err, "step_execution_id", stepExecution.ID)
	}

	// Set next steps
	result.NextSteps = step.NextSteps

	wos.logger.Info("Step executed successfully", "step_id", step.ID, "execution_id", execution.ID)
	return result, nil
}

// executeTaskStep executes a task creation step
func (wos *WorkflowOrchestrationService) executeTaskStep(ctx context.Context, step *entities.WorkflowStep, execCtx *WorkflowExecutionContext, result *StepExecutionResult) error {
	wos.logger.Info("Executing task step", "step_id", step.ID)

	// Extract task details from step configuration
	config := step.Configuration
	
	title, _ := config["title"].(string)
	description, _ := config["description"].(string)
	taskType, _ := config["type"].(string)
	priority, _ := config["priority"].(string)
	
	if title == "" {
		title = step.Name
	}
	if description == "" {
		description = step.Description
	}

	// Create task
	task := entities.NewTask(title, description, entities.TaskType(taskType), execCtx.ExecutedBy)
	
	if priority != "" {
		task.Priority = entities.TaskPriority(priority)
	}

	// Set project ID if available in context
	if projectID, ok := execCtx.Variables["project_id"].(uuid.UUID); ok {
		task.ProjectID = &projectID
	}

	// Create the task
	if err := wos.taskRepo.Create(ctx, task); err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Add task ID to result
	result.Output["task_id"] = task.ID
	result.Output["task_title"] = task.Title

	wos.logger.Info("Task created in workflow step", "task_id", task.ID, "step_id", step.ID)
	return nil
}

// executeApprovalStep executes an approval step
func (wos *WorkflowOrchestrationService) executeApprovalStep(ctx context.Context, step *entities.WorkflowStep, execCtx *WorkflowExecutionContext, result *StepExecutionResult) error {
	wos.logger.Info("Executing approval step", "step_id", step.ID)

	// Get approvers from step assignments
	var approverIDs []uuid.UUID
	for _, assignment := range step.Assignments {
		if assignment.Role == "approver" {
			approverIDs = append(approverIDs, assignment.UserID)
		}
	}

	if len(approverIDs) == 0 {
		return fmt.Errorf("no approvers configured for approval step")
	}

	// Create approval notifications
	for _, approverID := range approverIDs {
		notification := &repositories.Notification{
			ID:          uuid.New(),
			Type:        repositories.NotificationTypeWorkflowStarted,
			Title:       "Approval Required",
			Message:     fmt.Sprintf("Your approval is required for: %s", step.Name),
			Priority:    repositories.NotificationPriorityHigh,
			Category:    repositories.NotificationCategoryWorkflow,
			UserID:      approverID,
			RelatedID:   &execCtx.ExecutionID,
			RelatedType: "workflow_execution",
			Data: map[string]interface{}{
				"workflow_id":   execCtx.WorkflowID,
				"execution_id":  execCtx.ExecutionID,
				"step_id":       step.ID,
				"approval_type": "workflow_step",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := wos.notificationRepo.Create(ctx, notification); err != nil {
			wos.logger.Error("Failed to create approval notification", err, "approver_id", approverID)
		}
	}

	// For now, auto-approve (in real implementation, this would wait for actual approval)
	result.Output["approved"] = true
	result.Output["approved_by"] = approverIDs[0]
	result.Output["approved_at"] = time.Now()

	wos.logger.Info("Approval step executed", "step_id", step.ID)
	return nil
}

// executeNotificationStep executes a notification step
func (wos *WorkflowOrchestrationService) executeNotificationStep(ctx context.Context, step *entities.WorkflowStep, execCtx *WorkflowExecutionContext, result *StepExecutionResult) error {
	wos.logger.Info("Executing notification step", "step_id", step.ID)

	config := step.Configuration
	message, _ := config["message"].(string)
	recipients, _ := config["recipients"].([]interface{})

	if message == "" {
		message = step.Description
	}

	// Send notifications to recipients
	for _, recipient := range recipients {
		if userID, ok := recipient.(string); ok {
			if recipientUUID, err := uuid.Parse(userID); err == nil {
				notification := &repositories.Notification{
					ID:          uuid.New(),
					Type:        repositories.NotificationTypeWorkflowStarted,
					Title:       step.Name,
					Message:     message,
					Priority:    repositories.NotificationPriorityNormal,
					Category:    repositories.NotificationCategoryWorkflow,
					UserID:      recipientUUID,
					RelatedID:   &execCtx.ExecutionID,
					RelatedType: "workflow_execution",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}

				if err := wos.notificationRepo.Create(ctx, notification); err != nil {
					wos.logger.Error("Failed to create notification", err, "recipient_id", recipientUUID)
				}
			}
		}
	}

	result.Output["notifications_sent"] = len(recipients)
	wos.logger.Info("Notification step executed", "step_id", step.ID, "recipients", len(recipients))
	return nil
}

// executeActionStep executes an action step
func (wos *WorkflowOrchestrationService) executeActionStep(ctx context.Context, step *entities.WorkflowStep, execCtx *WorkflowExecutionContext, result *StepExecutionResult) error {
	wos.logger.Info("Executing action step", "step_id", step.ID)

	// Execute all actions in the step
	for _, action := range step.Actions {
		if wos.actionExecutor.CanExecute(action.Type) {
			output, err := wos.actionExecutor.ExecuteAction(ctx, &entities.WorkflowAction{
				Type:       action.Type,
				Target:     action.Target,
				Parameters: action.Parameters,
			}, execCtx.Variables)
			
			if err != nil {
				return fmt.Errorf("action execution failed: %w", err)
			}
			
			// Merge action output
			for k, v := range output {
				result.Output[k] = v
			}
		} else {
			wos.logger.Warn("Unsupported action type", "action_type", action.Type, "step_id", step.ID)
		}
	}

	wos.logger.Info("Action step executed", "step_id", step.ID)
	return nil
}

// executeWaitStep executes a wait step
func (wos *WorkflowOrchestrationService) executeWaitStep(ctx context.Context, step *entities.WorkflowStep, execCtx *WorkflowExecutionContext, result *StepExecutionResult) error {
	wos.logger.Info("Executing wait step", "step_id", step.ID)

	config := step.Configuration
	waitDuration, _ := config["duration"].(string)
	
	if waitDuration != "" {
		if duration, err := time.ParseDuration(waitDuration); err == nil {
			time.Sleep(duration)
			result.Output["waited_duration"] = duration.String()
		}
	}

	wos.logger.Info("Wait step executed", "step_id", step.ID)
	return nil
}

// Helper methods

func (wos *WorkflowOrchestrationService) failStep(ctx context.Context, stepExecution *entities.StepExecution, errorMessage string) (*StepExecutionResult, error) {
	stepExecution.Status = entities.StepStatusFailed
	stepExecution.ErrorMessage = errorMessage
	now := time.Now()
	stepExecution.FailedAt = &now
	stepExecution.UpdatedAt = now

	wos.workflowRepo.UpdateStepExecution(ctx, stepExecution)

	return &StepExecutionResult{
		Success:      false,
		ErrorMessage: errorMessage,
		ShouldRetry:  false,
	}, nil
}

func (wos *WorkflowOrchestrationService) completeExecution(ctx context.Context, workflow *entities.Workflow, execution *entities.WorkflowExecution, message string) error {
	execution.Complete()
	
	if err := wos.workflowRepo.UpdateExecution(ctx, execution); err != nil {
		wos.logger.Error("Failed to update completed execution", err, "execution_id", execution.ID)
	}

	// Publish completion event
	event := events.NewWorkflowExecutionCompletedEvent(workflow, execution)
	if err := wos.eventPublisher.PublishEvent(ctx, event); err != nil {
		wos.logger.Error("Failed to publish workflow completed event", err, "execution_id", execution.ID)
	}

	wos.logger.Info("Workflow execution completed", "workflow_id", workflow.ID, "execution_id", execution.ID)
	return nil
}

func (wos *WorkflowOrchestrationService) failExecution(ctx context.Context, workflow *entities.Workflow, execution *entities.WorkflowExecution, errorMessage string) error {
	execution.Fail(errorMessage)
	
	if err := wos.workflowRepo.UpdateExecution(ctx, execution); err != nil {
		wos.logger.Error("Failed to update failed execution", err, "execution_id", execution.ID)
	}

	// Publish failure event
	event := events.NewWorkflowExecutionFailedEvent(workflow, execution, errorMessage)
	if err := wos.eventPublisher.PublishEvent(ctx, event); err != nil {
		wos.logger.Error("Failed to publish workflow failed event", err, "execution_id", execution.ID)
	}

	wos.logger.Error("Workflow execution failed", nil, "workflow_id", workflow.ID, "execution_id", execution.ID, "error", errorMessage)
	return fmt.Errorf("workflow execution failed: %s", errorMessage)
}

// convertStepConditionsToWorkflowConditions converts StepConditions to WorkflowConditions
func (wos *WorkflowOrchestrationService) convertStepConditionsToWorkflowConditions(stepConditions []*entities.StepCondition, workflowID uuid.UUID) []*entities.WorkflowCondition {
	workflowConditions := make([]*entities.WorkflowCondition, len(stepConditions))
	
	for i, stepCondition := range stepConditions {
		workflowConditions[i] = &entities.WorkflowCondition{
			ID:         uuid.New(),
			WorkflowID: workflowID,
			Name:       "", // StepCondition doesn't have a name field
			Type:       entities.ConditionTypeField, // Default to field type
			Field:      stepCondition.Field,
			Operator:   stepCondition.Operator,
			Value:      stepCondition.Value,
			Logic:      entities.LogicAnd, // Default logic
			Configuration: make(map[string]interface{}),
			IsActive:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
	}
	
	return workflowConditions
}
