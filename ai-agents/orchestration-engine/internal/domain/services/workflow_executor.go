package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/orchestration-engine/internal/domain/entities"
)

// WorkflowExecutor handles the execution of a single workflow instance
type WorkflowExecutor struct {
	workflow       *entities.Workflow
	execution      *entities.WorkflowExecution
	agentRegistry  AgentRegistry
	executionRepo  ExecutionRepository
	eventPublisher EventPublisher
	logger         Logger
	
	currentStep    string
	stepResults    map[string]map[string]interface{}
	stepMutex      sync.RWMutex
	cancelChan     chan struct{}
	timeoutTimer   *time.Timer
	startTime      time.Time
	isHealthy      bool
	healthMutex    sync.RWMutex
}

// NewWorkflowExecutor creates a new workflow executor
func NewWorkflowExecutor(
	workflow *entities.Workflow,
	execution *entities.WorkflowExecution,
	agentRegistry AgentRegistry,
	executionRepo ExecutionRepository,
	eventPublisher EventPublisher,
	logger Logger,
) *WorkflowExecutor {
	return &WorkflowExecutor{
		workflow:       workflow,
		execution:      execution,
		agentRegistry:  agentRegistry,
		executionRepo:  executionRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
		stepResults:    make(map[string]map[string]interface{}),
		cancelChan:     make(chan struct{}),
		startTime:      time.Now(),
		isHealthy:      true,
	}
}

// Execute executes the workflow
func (we *WorkflowExecutor) Execute(ctx context.Context) error {
	we.logger.Info("Starting workflow execution", "execution_id", we.execution.ID, "workflow_id", we.workflow.ID)

	// Set up timeout if configured
	if we.workflow.Configuration != nil && we.workflow.Configuration.ExecutionTimeout > 0 {
		we.timeoutTimer = time.NewTimer(we.workflow.Configuration.ExecutionTimeout)
		go func() {
			select {
			case <-we.timeoutTimer.C:
				we.logger.Warn("Workflow execution timed out", "execution_id", we.execution.ID)
				we.Timeout(ctx)
			case <-we.cancelChan:
				we.timeoutTimer.Stop()
			}
		}()
	}

	defer func() {
		if we.timeoutTimer != nil {
			we.timeoutTimer.Stop()
		}
		close(we.cancelChan)
	}()

	// Execute workflow steps
	err := we.executeWorkflow(ctx)
	
	// Update final execution status
	now := time.Now()
	we.execution.CompletedAt = &now
	we.execution.Duration = now.Sub(we.execution.StartedAt)
	we.execution.UpdatedAt = now

	if err != nil {
		we.execution.Status = entities.WorkflowStatusFailed
		we.execution.Error = &entities.WorkflowError{
			Code:      "EXECUTION_FAILED",
			Message:   err.Error(),
			Timestamp: now,
		}
		we.logger.Error("Workflow execution failed", err, "execution_id", we.execution.ID)
	} else {
		we.execution.Status = entities.WorkflowStatusCompleted
		we.logger.Info("Workflow execution completed", "execution_id", we.execution.ID, "duration", we.execution.Duration)
	}

	// Save final execution state
	if updateErr := we.executionRepo.Update(ctx, we.execution); updateErr != nil {
		we.logger.Error("Failed to update execution", updateErr, "execution_id", we.execution.ID)
	}

	// Publish completion event
	event := &ExecutionEvent{
		ID:          uuid.New(),
		Type:        "execution.completed",
		WorkflowID:  we.workflow.ID,
		ExecutionID: we.execution.ID,
		Data: map[string]interface{}{
			"status":   we.execution.Status,
			"duration": we.execution.Duration.String(),
			"error":    we.execution.Error,
		},
		Timestamp: now,
		Source:    "workflow-executor",
	}
	we.eventPublisher.PublishExecutionEvent(ctx, event)

	return err
}

// executeWorkflow executes the workflow definition
func (we *WorkflowExecutor) executeWorkflow(ctx context.Context) error {
	if we.workflow.Definition == nil {
		return fmt.Errorf("workflow definition is nil")
	}

	definition := we.workflow.Definition
	
	// Start from the start step
	currentStepID := definition.StartStep
	if currentStepID == "" {
		return fmt.Errorf("no start step defined")
	}

	we.logger.Info("Starting workflow execution from step", "step_id", currentStepID)

	// Execute steps
	for currentStepID != "" {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-we.cancelChan:
			return fmt.Errorf("execution cancelled")
		default:
		}

		// Update current step
		we.setCurrentStep(currentStepID)
		we.execution.CurrentStep = currentStepID

		// Get step definition
		stepDef, exists := definition.Steps[currentStepID]
		if !exists {
			return fmt.Errorf("step definition not found: %s", currentStepID)
		}

		we.logger.Info("Executing step", "step_id", currentStepID, "step_name", stepDef.Name)

		// Execute step
		stepResult, err := we.executeStep(ctx, stepDef)
		if err != nil {
			we.logger.Error("Step execution failed", err, "step_id", currentStepID)
			
			// Handle step error based on error handling configuration
			if stepDef.ErrorHandling != nil {
				nextStep, handleErr := we.handleStepError(ctx, stepDef, err)
				if handleErr != nil {
					return fmt.Errorf("error handling failed: %w", handleErr)
				}
				currentStepID = nextStep
				continue
			}
			
			return fmt.Errorf("step %s failed: %w", currentStepID, err)
		}

		// Store step result
		we.setStepResult(currentStepID, stepResult)

		// Add to completed steps
		we.execution.CompletedSteps = append(we.execution.CompletedSteps, currentStepID)

		// Determine next step
		nextStepID, err := we.getNextStep(ctx, currentStepID, stepResult)
		if err != nil {
			return fmt.Errorf("failed to determine next step: %w", err)
		}

		currentStepID = nextStepID

		// Check if we've reached an end step
		if we.isEndStep(currentStepID) {
			we.logger.Info("Reached end step", "step_id", currentStepID)
			break
		}

		// Update execution progress
		if updateErr := we.executionRepo.Update(ctx, we.execution); updateErr != nil {
			we.logger.Warn("Failed to update execution progress", "error", updateErr)
		}
	}

	we.logger.Info("Workflow execution completed successfully", "execution_id", we.execution.ID)
	return nil
}

// executeStep executes a single workflow step
func (we *WorkflowExecutor) executeStep(ctx context.Context, stepDef *entities.StepDefinition) (map[string]interface{}, error) {
	stepStartTime := time.Now()

	// Create step execution record
	step := &entities.WorkflowStep{
		ID:               uuid.New(),
		WorkflowID:       we.workflow.ID,
		ExecutionID:      we.execution.ID,
		StepDefinitionID: stepDef.ID,
		Name:             stepDef.Name,
		Status:           entities.StepStatusRunning,
		StartedAt:        &stepStartTime,
		Input:            we.prepareStepInput(stepDef),
		CreatedAt:        stepStartTime,
		UpdatedAt:        stepStartTime,
	}

	// Publish step started event
	stepEvent := &StepEvent{
		ID:          uuid.New(),
		Type:        "step.started",
		WorkflowID:  we.workflow.ID,
		ExecutionID: we.execution.ID,
		StepID:      stepDef.ID,
		Data: map[string]interface{}{
			"step_name": stepDef.Name,
			"step_type": stepDef.Type,
			"input":     step.Input,
		},
		Timestamp: stepStartTime,
		Source:    "workflow-executor",
	}
	we.eventPublisher.PublishStepEvent(ctx, stepEvent)

	// Set up step timeout
	stepCtx := ctx
	if stepDef.Timeout > 0 {
		var cancel context.CancelFunc
		stepCtx, cancel = context.WithTimeout(ctx, stepDef.Timeout)
		defer cancel()
	}

	// Execute step based on type
	var result map[string]interface{}
	var err error

	switch stepDef.Type {
	case entities.StepTypeAgent:
		result, err = we.executeAgentStep(stepCtx, stepDef, step.Input)
	case entities.StepTypeCondition:
		result, err = we.executeConditionStep(stepCtx, stepDef, step.Input)
	case entities.StepTypeWait:
		result, err = we.executeWaitStep(stepCtx, stepDef, step.Input)
	case entities.StepTypeTransform:
		result, err = we.executeTransformStep(stepCtx, stepDef, step.Input)
	case entities.StepTypeValidation:
		result, err = we.executeValidationStep(stepCtx, stepDef, step.Input)
	case entities.StepTypeNotification:
		result, err = we.executeNotificationStep(stepCtx, stepDef, step.Input)
	default:
		err = fmt.Errorf("unsupported step type: %s", stepDef.Type)
	}

	// Update step completion
	stepEndTime := time.Now()
	step.CompletedAt = &stepEndTime
	step.Duration = stepEndTime.Sub(stepStartTime)
	step.UpdatedAt = stepEndTime

	if err != nil {
		step.Status = entities.StepStatusFailed
		step.Error = &entities.StepError{
			Code:      "STEP_EXECUTION_FAILED",
			Message:   err.Error(),
			Timestamp: stepEndTime,
			Retryable: we.isRetryableError(err),
		}
		we.execution.FailedSteps = append(we.execution.FailedSteps, stepDef.ID)
	} else {
		step.Status = entities.StepStatusCompleted
		step.Output = result
	}

	// Publish step completed event
	stepEvent = &StepEvent{
		ID:          uuid.New(),
		Type:        "step.completed",
		WorkflowID:  we.workflow.ID,
		ExecutionID: we.execution.ID,
		StepID:      stepDef.ID,
		Data: map[string]interface{}{
			"step_name": stepDef.Name,
			"status":    step.Status,
			"duration":  step.Duration.String(),
			"output":    result,
			"error":     step.Error,
		},
		Timestamp: stepEndTime,
		Source:    "workflow-executor",
	}
	we.eventPublisher.PublishStepEvent(ctx, stepEvent)

	we.logger.Info("Step execution completed", 
		"step_id", stepDef.ID, 
		"status", step.Status, 
		"duration", step.Duration)

	return result, err
}

// executeAgentStep executes an agent step
func (we *WorkflowExecutor) executeAgentStep(ctx context.Context, stepDef *entities.StepDefinition, input map[string]interface{}) (map[string]interface{}, error) {
	if stepDef.AgentType == "" {
		return nil, fmt.Errorf("agent type not specified")
	}

	agent, err := we.agentRegistry.GetAgent(stepDef.AgentType)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent %s: %w", stepDef.AgentType, err)
	}

	// Validate input
	if err := agent.Validate(stepDef.Action, input); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Execute agent action
	result, err := agent.Execute(ctx, stepDef.Action, input)
	if err != nil {
		return nil, fmt.Errorf("agent execution failed: %w", err)
	}

	return result, nil
}

// executeConditionStep executes a condition step
func (we *WorkflowExecutor) executeConditionStep(ctx context.Context, stepDef *entities.StepDefinition, input map[string]interface{}) (map[string]interface{}, error) {
	if len(stepDef.Conditions) == 0 {
		return nil, fmt.Errorf("no conditions defined")
	}

	result := make(map[string]interface{})
	
	for i, condition := range stepDef.Conditions {
		conditionResult, err := we.evaluateCondition(condition, input)
		if err != nil {
			return nil, fmt.Errorf("condition %d evaluation failed: %w", i, err)
		}
		result[fmt.Sprintf("condition_%d", i)] = conditionResult
	}

	return result, nil
}

// executeWaitStep executes a wait step
func (we *WorkflowExecutor) executeWaitStep(ctx context.Context, stepDef *entities.StepDefinition, input map[string]interface{}) (map[string]interface{}, error) {
	duration := time.Second // Default wait time
	
	if durationParam, exists := stepDef.Parameters["duration"]; exists {
		if durationStr, ok := durationParam.(string); ok {
			if parsedDuration, err := time.ParseDuration(durationStr); err == nil {
				duration = parsedDuration
			}
		}
	}

	we.logger.Info("Waiting", "duration", duration, "step_id", stepDef.ID)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(duration):
		return map[string]interface{}{
			"waited_duration": duration.String(),
		}, nil
	}
}

// executeTransformStep executes a transform step
func (we *WorkflowExecutor) executeTransformStep(ctx context.Context, stepDef *entities.StepDefinition, input map[string]interface{}) (map[string]interface{}, error) {
	// Simple transformation logic - in production this would be more sophisticated
	result := make(map[string]interface{})
	
	// Apply output mapping if defined
	if stepDef.OutputMapping != nil {
		for outputKey, inputKey := range stepDef.OutputMapping {
			if value, exists := input[inputKey]; exists {
				result[outputKey] = value
			}
		}
	} else {
		// Default: pass through all input
		for k, v := range input {
			result[k] = v
		}
	}

	return result, nil
}

// executeValidationStep executes a validation step
func (we *WorkflowExecutor) executeValidationStep(ctx context.Context, stepDef *entities.StepDefinition, input map[string]interface{}) (map[string]interface{}, error) {
	// Simple validation logic
	validationErrors := []string{}

	// Check required fields
	if requiredFields, exists := stepDef.Parameters["required_fields"]; exists {
		if fields, ok := requiredFields.([]interface{}); ok {
			for _, field := range fields {
				if fieldName, ok := field.(string); ok {
					if _, exists := input[fieldName]; !exists {
						validationErrors = append(validationErrors, fmt.Sprintf("required field missing: %s", fieldName))
					}
				}
			}
		}
	}

	result := map[string]interface{}{
		"valid":  len(validationErrors) == 0,
		"errors": validationErrors,
	}

	if len(validationErrors) > 0 {
		return result, fmt.Errorf("validation failed: %v", validationErrors)
	}

	return result, nil
}

// executeNotificationStep executes a notification step
func (we *WorkflowExecutor) executeNotificationStep(ctx context.Context, stepDef *entities.StepDefinition, input map[string]interface{}) (map[string]interface{}, error) {
	// Simple notification logic - in production this would integrate with notification services
	message := "Workflow notification"
	if msg, exists := stepDef.Parameters["message"]; exists {
		if msgStr, ok := msg.(string); ok {
			message = msgStr
		}
	}

	we.logger.Info("Sending notification", "message", message, "step_id", stepDef.ID)

	return map[string]interface{}{
		"notification_sent": true,
		"message":          message,
		"timestamp":        time.Now(),
	}, nil
}

// Helper methods

func (we *WorkflowExecutor) prepareStepInput(stepDef *entities.StepDefinition) map[string]interface{} {
	input := make(map[string]interface{})

	// Add workflow variables
	for k, v := range we.execution.Variables {
		input[k] = v
	}

	// Add step parameters
	if stepDef.Parameters != nil {
		for k, v := range stepDef.Parameters {
			input[k] = v
		}
	}

	// Apply input mapping
	if stepDef.InputMapping != nil {
		mappedInput := make(map[string]interface{})
		for stepInputKey, sourceKey := range stepDef.InputMapping {
			if value, exists := input[sourceKey]; exists {
				mappedInput[stepInputKey] = value
			}
		}
		// Merge mapped input
		for k, v := range mappedInput {
			input[k] = v
		}
	}

	return input
}

func (we *WorkflowExecutor) getNextStep(ctx context.Context, currentStepID string, stepResult map[string]interface{}) (string, error) {
	definition := we.workflow.Definition
	
	// Find connections from current step
	for _, connection := range definition.Connections {
		if connection.FromStep == currentStepID {
			// Check condition if present
			if connection.Condition != nil {
				conditionMet, err := we.evaluateCondition(connection.Condition, stepResult)
				if err != nil {
					return "", fmt.Errorf("failed to evaluate connection condition: %w", err)
				}
				if conditionMet {
					return connection.ToStep, nil
				}
			} else if connection.IsDefault {
				return connection.ToStep, nil
			}
		}
	}

	// No next step found
	return "", nil
}

func (we *WorkflowExecutor) isEndStep(stepID string) bool {
	if stepID == "" {
		return true
	}
	
	for _, endStep := range we.workflow.Definition.EndSteps {
		if stepID == endStep {
			return true
		}
	}
	
	return false
}

func (we *WorkflowExecutor) evaluateCondition(condition *entities.Condition, context map[string]interface{}) (bool, error) {
	// Simple condition evaluation - in production this would be more sophisticated
	switch condition.Operator {
	case entities.ConditionOperatorEquals:
		if value, exists := context[condition.Expression]; exists {
			return value == condition.Value, nil
		}
		return false, nil
	case entities.ConditionOperatorNotEquals:
		if value, exists := context[condition.Expression]; exists {
			return value != condition.Value, nil
		}
		return true, nil
	default:
		return false, fmt.Errorf("unsupported condition operator: %s", condition.Operator)
	}
}

func (we *WorkflowExecutor) handleStepError(ctx context.Context, stepDef *entities.StepDefinition, err error) (string, error) {
	errorHandling := stepDef.ErrorHandling
	
	switch errorHandling.Strategy {
	case entities.ErrorHandlingStrategyContinue:
		if errorHandling.ContinueOnError {
			return we.getNextStep(ctx, stepDef.ID, map[string]interface{}{"error": err.Error()})
		}
		return "", err
	case entities.ErrorHandlingStrategyFallback:
		if errorHandling.FallbackStep != "" {
			return errorHandling.FallbackStep, nil
		}
		return "", err
	default:
		return "", err
	}
}

func (we *WorkflowExecutor) isRetryableError(err error) bool {
	// Simple retry logic - in production this would be more sophisticated
	errorMsg := err.Error()
	retryableErrors := []string{"timeout", "connection", "temporary"}
	
	for _, retryable := range retryableErrors {
		if contains(errorMsg, retryable) {
			return true
		}
	}
	
	return false
}

func (we *WorkflowExecutor) setCurrentStep(stepID string) {
	we.stepMutex.Lock()
	defer we.stepMutex.Unlock()
	we.currentStep = stepID
}

func (we *WorkflowExecutor) setStepResult(stepID string, result map[string]interface{}) {
	we.stepMutex.Lock()
	defer we.stepMutex.Unlock()
	we.stepResults[stepID] = result
}

func (we *WorkflowExecutor) setHealthy(healthy bool) {
	we.healthMutex.Lock()
	defer we.healthMutex.Unlock()
	we.isHealthy = healthy
}

// Public methods for external control

func (we *WorkflowExecutor) Cancel(ctx context.Context) error {
	we.logger.Info("Cancelling workflow execution", "execution_id", we.execution.ID)
	
	select {
	case we.cancelChan <- struct{}{}:
		return nil
	default:
		return fmt.Errorf("execution already cancelled or completed")
	}
}

func (we *WorkflowExecutor) Stop(ctx context.Context) {
	we.Cancel(ctx)
}

func (we *WorkflowExecutor) IsTimedOut() bool {
	if we.workflow.Configuration == nil || we.workflow.Configuration.ExecutionTimeout == 0 {
		return false
	}
	
	return time.Since(we.startTime) > we.workflow.Configuration.ExecutionTimeout
}

func (we *WorkflowExecutor) IsHealthy() bool {
	we.healthMutex.RLock()
	defer we.healthMutex.RUnlock()
	return we.isHealthy
}

func (we *WorkflowExecutor) Timeout(ctx context.Context) {
	we.logger.Warn("Workflow execution timed out", "execution_id", we.execution.ID)
	
	we.execution.Status = entities.WorkflowStatusFailed
	we.execution.Error = &entities.WorkflowError{
		Code:      "EXECUTION_TIMEOUT",
		Message:   "Workflow execution timed out",
		Timestamp: time.Now(),
	}
	
	we.setHealthy(false)
	we.Cancel(ctx)
}

// Utility function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		func() bool {
			for i := 1; i < len(s)-len(substr)+1; i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())))
}
