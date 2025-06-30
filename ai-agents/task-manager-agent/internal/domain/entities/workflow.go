package entities

import (
	"time"

	"github.com/google/uuid"
)

// Workflow represents a comprehensive workflow entity
type Workflow struct {
	ID           uuid.UUID              `json:"id" redis:"id"`
	Name         string                 `json:"name" redis:"name"`
	Description  string                 `json:"description" redis:"description"`
	Type         WorkflowType           `json:"type" redis:"type"`
	Status       WorkflowStatus         `json:"status" redis:"status"`
	Category     WorkflowCategory       `json:"category" redis:"category"`
	Version      string                 `json:"version" redis:"version"`
	OwnerID      uuid.UUID              `json:"owner_id" redis:"owner_id"`
	Owner        *User                  `json:"owner,omitempty"`
	Steps        []*WorkflowStep        `json:"steps,omitempty"`
	Triggers     []*WorkflowTrigger     `json:"triggers,omitempty"`
	Conditions   []*WorkflowCondition   `json:"conditions,omitempty"`
	Actions      []*WorkflowAction      `json:"actions,omitempty"`
	Variables    map[string]interface{} `json:"variables" redis:"variables"`
	Configuration *WorkflowConfig       `json:"configuration,omitempty"`
	Executions   []*WorkflowExecution   `json:"executions,omitempty"`
	Tags         []string               `json:"tags" redis:"tags"`
	Labels       []string               `json:"labels" redis:"labels"`
	CustomFields map[string]interface{} `json:"custom_fields" redis:"custom_fields"`
	Metadata     map[string]interface{} `json:"metadata" redis:"metadata"`
	ExternalIDs  map[string]string      `json:"external_ids" redis:"external_ids"`
	IsTemplate   bool                   `json:"is_template" redis:"is_template"`
	TemplateID   *uuid.UUID             `json:"template_id,omitempty" redis:"template_id"`
	IsActive     bool                   `json:"is_active" redis:"is_active"`
	CreatedAt    time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" redis:"updated_at"`
	CreatedBy    uuid.UUID              `json:"created_by" redis:"created_by"`
	UpdatedBy    uuid.UUID              `json:"updated_by" redis:"updated_by"`
	VersionNum   int64                  `json:"version_num" redis:"version_num"`
}

// WorkflowType defines the type of workflow
type WorkflowType string

const (
	WorkflowTypeSequential WorkflowType = "sequential"
	WorkflowTypeParallel   WorkflowType = "parallel"
	WorkflowTypeConditional WorkflowType = "conditional"
	WorkflowTypeLoop       WorkflowType = "loop"
	WorkflowTypeEvent      WorkflowType = "event"
	WorkflowTypeApproval   WorkflowType = "approval"
	WorkflowTypeAutomation WorkflowType = "automation"
)

// WorkflowStatus defines the status of a workflow
type WorkflowStatus string

const (
	WorkflowStatusDraft     WorkflowStatus = "draft"
	WorkflowStatusActive    WorkflowStatus = "active"
	WorkflowStatusInactive  WorkflowStatus = "inactive"
	WorkflowStatusArchived  WorkflowStatus = "archived"
	WorkflowStatusDeprecated WorkflowStatus = "deprecated"
)

// WorkflowCategory defines the category of workflow
type WorkflowCategory string

const (
	CategoryTaskManagement WorkflowCategory = "task_management"
	CategoryProjectManagement WorkflowCategory = "project_management"
	CategoryApprovalProcess WorkflowCategory = "approval_process"
	CategoryNotification   WorkflowCategory = "notification"
	CategoryIntegration    WorkflowCategory = "integration"
	CategoryAutomation     WorkflowCategory = "automation"
	CategoryReporting      WorkflowCategory = "reporting"
	CategoryQualityControl WorkflowCategory = "quality_control"
)

// WorkflowStep represents a step in a workflow
type WorkflowStep struct {
	ID           uuid.UUID              `json:"id" redis:"id"`
	WorkflowID   uuid.UUID              `json:"workflow_id" redis:"workflow_id"`
	Name         string                 `json:"name" redis:"name"`
	Description  string                 `json:"description" redis:"description"`
	Type         StepType               `json:"type" redis:"type"`
	Order        int                    `json:"order" redis:"order"`
	Conditions   []*StepCondition       `json:"conditions,omitempty"`
	Actions      []*StepAction          `json:"actions,omitempty"`
	Assignments  []*StepAssignment      `json:"assignments,omitempty"`
	TimeoutHours int                    `json:"timeout_hours" redis:"timeout_hours"`
	IsOptional   bool                   `json:"is_optional" redis:"is_optional"`
	IsParallel   bool                   `json:"is_parallel" redis:"is_parallel"`
	NextSteps    []uuid.UUID            `json:"next_steps" redis:"next_steps"`
	Configuration map[string]interface{} `json:"configuration" redis:"configuration"`
	IsActive     bool                   `json:"is_active" redis:"is_active"`
	CreatedAt    time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" redis:"updated_at"`
}

// StepType defines the type of workflow step
type StepType string

const (
	StepTypeTask        StepType = "task"
	StepTypeApproval    StepType = "approval"
	StepTypeReview      StepType = "review"
	StepTypeNotification StepType = "notification"
	StepTypeCondition   StepType = "condition"
	StepTypeAction      StepType = "action"
	StepTypeWait        StepType = "wait"
	StepTypeLoop        StepType = "loop"
	StepTypeSubWorkflow StepType = "sub_workflow"
)

// WorkflowTrigger represents a trigger that starts a workflow
type WorkflowTrigger struct {
	ID           uuid.UUID              `json:"id" redis:"id"`
	WorkflowID   uuid.UUID              `json:"workflow_id" redis:"workflow_id"`
	Name         string                 `json:"name" redis:"name"`
	Type         TriggerType            `json:"type" redis:"type"`
	Event        string                 `json:"event" redis:"event"`
	Conditions   []*TriggerCondition    `json:"conditions,omitempty"`
	Configuration map[string]interface{} `json:"configuration" redis:"configuration"`
	IsActive     bool                   `json:"is_active" redis:"is_active"`
	CreatedAt    time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" redis:"updated_at"`
}

// TriggerType defines the type of workflow trigger
type TriggerType string

const (
	TriggerTypeManual    TriggerType = "manual"
	TriggerTypeScheduled TriggerType = "scheduled"
	TriggerTypeEvent     TriggerType = "event"
	TriggerTypeWebhook   TriggerType = "webhook"
	TriggerTypeEmail     TriggerType = "email"
	TriggerTypeAPI       TriggerType = "api"
)

// WorkflowCondition represents a condition in a workflow
type WorkflowCondition struct {
	ID         uuid.UUID              `json:"id" redis:"id"`
	WorkflowID uuid.UUID              `json:"workflow_id" redis:"workflow_id"`
	Name       string                 `json:"name" redis:"name"`
	Type       ConditionType          `json:"type" redis:"type"`
	Field      string                 `json:"field" redis:"field"`
	Operator   ConditionOperator      `json:"operator" redis:"operator"`
	Value      interface{}            `json:"value" redis:"value"`
	Logic      ConditionLogic         `json:"logic" redis:"logic"`
	Configuration map[string]interface{} `json:"configuration" redis:"configuration"`
	IsActive   bool                   `json:"is_active" redis:"is_active"`
	CreatedAt  time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" redis:"updated_at"`
}

// ConditionType defines the type of condition
type ConditionType string

const (
	ConditionTypeField    ConditionType = "field"
	ConditionTypeTime     ConditionType = "time"
	ConditionTypeUser     ConditionType = "user"
	ConditionTypeStatus   ConditionType = "status"
	ConditionTypeCustom   ConditionType = "custom"
)

// ConditionOperator defines the operator for conditions
type ConditionOperator string

const (
	OperatorEquals       ConditionOperator = "equals"
	OperatorNotEquals    ConditionOperator = "not_equals"
	OperatorGreaterThan  ConditionOperator = "greater_than"
	OperatorLessThan     ConditionOperator = "less_than"
	OperatorContains     ConditionOperator = "contains"
	OperatorNotContains  ConditionOperator = "not_contains"
	OperatorIn           ConditionOperator = "in"
	OperatorNotIn        ConditionOperator = "not_in"
	OperatorIsEmpty      ConditionOperator = "is_empty"
	OperatorIsNotEmpty   ConditionOperator = "is_not_empty"
)

// ConditionLogic defines the logic for combining conditions
type ConditionLogic string

const (
	LogicAnd ConditionLogic = "and"
	LogicOr  ConditionLogic = "or"
	LogicNot ConditionLogic = "not"
)

// WorkflowAction represents an action in a workflow
type WorkflowAction struct {
	ID           uuid.UUID              `json:"id" redis:"id"`
	WorkflowID   uuid.UUID              `json:"workflow_id" redis:"workflow_id"`
	Name         string                 `json:"name" redis:"name"`
	Type         ActionType             `json:"type" redis:"type"`
	Target       string                 `json:"target" redis:"target"`
	Parameters   map[string]interface{} `json:"parameters" redis:"parameters"`
	Configuration map[string]interface{} `json:"configuration" redis:"configuration"`
	Order        int                    `json:"order" redis:"order"`
	IsActive     bool                   `json:"is_active" redis:"is_active"`
	CreatedAt    time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" redis:"updated_at"`
}

// ActionType defines the type of workflow action
type ActionType string

const (
	ActionTypeCreateTask     ActionType = "create_task"
	ActionTypeUpdateTask     ActionType = "update_task"
	ActionTypeAssignTask     ActionType = "assign_task"
	ActionTypeSendEmail      ActionType = "send_email"
	ActionTypeSendNotification ActionType = "send_notification"
	ActionTypeWebhook        ActionType = "webhook"
	ActionTypeAPI            ActionType = "api"
	ActionTypeScript         ActionType = "script"
	ActionTypeApproval       ActionType = "approval"
	ActionTypeDelay          ActionType = "delay"
)

// WorkflowConfig represents workflow configuration
type WorkflowConfig struct {
	MaxExecutionTime   int                    `json:"max_execution_time" redis:"max_execution_time"`
	RetryAttempts      int                    `json:"retry_attempts" redis:"retry_attempts"`
	RetryDelay         int                    `json:"retry_delay" redis:"retry_delay"`
	NotifyOnFailure    bool                   `json:"notify_on_failure" redis:"notify_on_failure"`
	NotifyOnCompletion bool                   `json:"notify_on_completion" redis:"notify_on_completion"`
	AllowParallel      bool                   `json:"allow_parallel" redis:"allow_parallel"`
	Priority           WorkflowPriority       `json:"priority" redis:"priority"`
	Settings           map[string]interface{} `json:"settings" redis:"settings"`
}

// WorkflowPriority defines the priority of workflow execution
type WorkflowPriority string

const (
	WorkflowPriorityLow    WorkflowPriority = "low"
	WorkflowPriorityNormal WorkflowPriority = "normal"
	WorkflowPriorityHigh   WorkflowPriority = "high"
	WorkflowPriorityUrgent WorkflowPriority = "urgent"
)

// WorkflowExecution represents an execution instance of a workflow
type WorkflowExecution struct {
	ID           uuid.UUID              `json:"id" redis:"id"`
	WorkflowID   uuid.UUID              `json:"workflow_id" redis:"workflow_id"`
	Workflow     *Workflow              `json:"workflow,omitempty"`
	TriggerID    *uuid.UUID             `json:"trigger_id,omitempty" redis:"trigger_id"`
	Status       ExecutionStatus        `json:"status" redis:"status"`
	Context      map[string]interface{} `json:"context" redis:"context"`
	Variables    map[string]interface{} `json:"variables" redis:"variables"`
	CurrentStep  *uuid.UUID             `json:"current_step,omitempty" redis:"current_step"`
	StepExecutions []*StepExecution     `json:"step_executions,omitempty"`
	StartedAt    time.Time              `json:"started_at" redis:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty" redis:"completed_at"`
	FailedAt     *time.Time             `json:"failed_at,omitempty" redis:"failed_at"`
	ErrorMessage string                 `json:"error_message" redis:"error_message"`
	RetryCount   int                    `json:"retry_count" redis:"retry_count"`
	ExecutedBy   uuid.UUID              `json:"executed_by" redis:"executed_by"`
	CreatedAt    time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" redis:"updated_at"`
}

// ExecutionStatus defines the status of workflow execution
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
	ExecutionStatusPaused    ExecutionStatus = "paused"
)

// StepExecution represents the execution of a workflow step
type StepExecution struct {
	ID           uuid.UUID              `json:"id" redis:"id"`
	ExecutionID  uuid.UUID              `json:"execution_id" redis:"execution_id"`
	StepID       uuid.UUID              `json:"step_id" redis:"step_id"`
	Step         *WorkflowStep          `json:"step,omitempty"`
	Status       StepExecutionStatus    `json:"status" redis:"status"`
	Input        map[string]interface{} `json:"input" redis:"input"`
	Output       map[string]interface{} `json:"output" redis:"output"`
	StartedAt    time.Time              `json:"started_at" redis:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty" redis:"completed_at"`
	FailedAt     *time.Time             `json:"failed_at,omitempty" redis:"failed_at"`
	ErrorMessage string                 `json:"error_message" redis:"error_message"`
	RetryCount   int                    `json:"retry_count" redis:"retry_count"`
	AssignedTo   *uuid.UUID             `json:"assigned_to,omitempty" redis:"assigned_to"`
	CreatedAt    time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" redis:"updated_at"`
}

// StepExecutionStatus defines the status of step execution
type StepExecutionStatus string

const (
	StepStatusPending   StepExecutionStatus = "pending"
	StepStatusRunning   StepExecutionStatus = "running"
	StepStatusCompleted StepExecutionStatus = "completed"
	StepStatusFailed    StepExecutionStatus = "failed"
	StepStatusSkipped   StepExecutionStatus = "skipped"
	StepStatusWaiting   StepExecutionStatus = "waiting"
)

// Supporting types for steps
type StepCondition struct {
	Field    string            `json:"field" redis:"field"`
	Operator ConditionOperator `json:"operator" redis:"operator"`
	Value    interface{}       `json:"value" redis:"value"`
}

type StepAction struct {
	Type       ActionType             `json:"type" redis:"type"`
	Target     string                 `json:"target" redis:"target"`
	Parameters map[string]interface{} `json:"parameters" redis:"parameters"`
}

type StepAssignment struct {
	UserID uuid.UUID `json:"user_id" redis:"user_id"`
	Role   string    `json:"role" redis:"role"`
}

type TriggerCondition struct {
	Field    string            `json:"field" redis:"field"`
	Operator ConditionOperator `json:"operator" redis:"operator"`
	Value    interface{}       `json:"value" redis:"value"`
}

// NewWorkflow creates a new workflow with default values
func NewWorkflow(name, description string, workflowType WorkflowType, ownerID, createdBy uuid.UUID) *Workflow {
	now := time.Now()
	return &Workflow{
		ID:           uuid.New(),
		Name:         name,
		Description:  description,
		Type:         workflowType,
		Status:       WorkflowStatusDraft,
		Version:      "1.0",
		OwnerID:      ownerID,
		Variables:    make(map[string]interface{}),
		Tags:         []string{},
		Labels:       []string{},
		CustomFields: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
		ExternalIDs:  make(map[string]string),
		IsTemplate:   false,
		IsActive:     false,
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    createdBy,
		UpdatedBy:    createdBy,
		VersionNum:   1,
	}
}

// Activate activates the workflow
func (w *Workflow) Activate(updatedBy uuid.UUID) {
	w.Status = WorkflowStatusActive
	w.IsActive = true
	w.UpdatedBy = updatedBy
	w.UpdatedAt = time.Now()
	w.VersionNum++
}

// Deactivate deactivates the workflow
func (w *Workflow) Deactivate(updatedBy uuid.UUID) {
	w.Status = WorkflowStatusInactive
	w.IsActive = false
	w.UpdatedBy = updatedBy
	w.UpdatedAt = time.Now()
	w.VersionNum++
}

// AddStep adds a step to the workflow
func (w *Workflow) AddStep(step *WorkflowStep) {
	step.WorkflowID = w.ID
	w.Steps = append(w.Steps, step)
	w.UpdatedAt = time.Now()
	w.VersionNum++
}

// AddTrigger adds a trigger to the workflow
func (w *Workflow) AddTrigger(trigger *WorkflowTrigger) {
	trigger.WorkflowID = w.ID
	w.Triggers = append(w.Triggers, trigger)
	w.UpdatedAt = time.Now()
	w.VersionNum++
}

// CanExecute checks if the workflow can be executed
func (w *Workflow) CanExecute() bool {
	return w.IsActive && w.Status == WorkflowStatusActive && len(w.Steps) > 0
}

// GetFirstStep returns the first step in the workflow
func (w *Workflow) GetFirstStep() *WorkflowStep {
	if len(w.Steps) == 0 {
		return nil
	}
	
	// Find step with order 1 or lowest order
	var firstStep *WorkflowStep
	minOrder := int(^uint(0) >> 1) // Max int
	
	for _, step := range w.Steps {
		if step.IsActive && step.Order < minOrder {
			minOrder = step.Order
			firstStep = step
		}
	}
	
	return firstStep
}

// GetStepByID returns a step by its ID
func (w *Workflow) GetStepByID(stepID uuid.UUID) *WorkflowStep {
	for _, step := range w.Steps {
		if step.ID == stepID {
			return step
		}
	}
	return nil
}

// NewWorkflowExecution creates a new workflow execution
func NewWorkflowExecution(workflowID, executedBy uuid.UUID, context map[string]interface{}) *WorkflowExecution {
	now := time.Now()
	return &WorkflowExecution{
		ID:         uuid.New(),
		WorkflowID: workflowID,
		Status:     ExecutionStatusPending,
		Context:    context,
		Variables:  make(map[string]interface{}),
		StartedAt:  now,
		ExecutedBy: executedBy,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Start starts the workflow execution
func (we *WorkflowExecution) Start() {
	we.Status = ExecutionStatusRunning
	we.StartedAt = time.Now()
	we.UpdatedAt = time.Now()
}

// Complete completes the workflow execution
func (we *WorkflowExecution) Complete() {
	we.Status = ExecutionStatusCompleted
	now := time.Now()
	we.CompletedAt = &now
	we.UpdatedAt = now
}

// Fail fails the workflow execution
func (we *WorkflowExecution) Fail(errorMessage string) {
	we.Status = ExecutionStatusFailed
	we.ErrorMessage = errorMessage
	now := time.Now()
	we.FailedAt = &now
	we.UpdatedAt = now
}

// Cancel cancels the workflow execution
func (we *WorkflowExecution) Cancel() {
	we.Status = ExecutionStatusCancelled
	we.UpdatedAt = time.Now()
}

// IsCompleted checks if the execution is completed
func (we *WorkflowExecution) IsCompleted() bool {
	return we.Status == ExecutionStatusCompleted
}

// IsFailed checks if the execution failed
func (we *WorkflowExecution) IsFailed() bool {
	return we.Status == ExecutionStatusFailed
}

// IsRunning checks if the execution is running
func (we *WorkflowExecution) IsRunning() bool {
	return we.Status == ExecutionStatusRunning
}
