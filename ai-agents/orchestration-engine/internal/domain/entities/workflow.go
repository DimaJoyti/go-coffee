package entities

import (
	"time"

	"github.com/google/uuid"
)

// Workflow represents a business workflow that coordinates multiple agents
type Workflow struct {
	ID              uuid.UUID              `json:"id" redis:"id"`
	Name            string                 `json:"name" redis:"name"`
	Description     string                 `json:"description" redis:"description"`
	Type            WorkflowType           `json:"type" redis:"type"`
	Status          WorkflowStatus         `json:"status" redis:"status"`
	Priority        WorkflowPriority       `json:"priority" redis:"priority"`
	Category        WorkflowCategory       `json:"category" redis:"category"`
	Version         string                 `json:"version" redis:"version"`
	Definition      *WorkflowDefinition    `json:"definition,omitempty"`
	Steps           []*WorkflowStep        `json:"steps,omitempty"`
	Executions      []*WorkflowExecution   `json:"executions,omitempty"`
	Triggers        []*WorkflowTrigger     `json:"triggers,omitempty"`
	Variables       map[string]interface{} `json:"variables" redis:"variables"`
	Configuration   *WorkflowConfig        `json:"configuration,omitempty"`
	Metrics         *WorkflowMetrics       `json:"metrics,omitempty"`
	Tags            []string               `json:"tags" redis:"tags"`
	IsActive        bool                   `json:"is_active" redis:"is_active"`
	IsTemplate      bool                   `json:"is_template" redis:"is_template"`
	TemplateID      *uuid.UUID             `json:"template_id,omitempty" redis:"template_id"`
	ParentID        *uuid.UUID             `json:"parent_id,omitempty" redis:"parent_id"`
	CreatedAt       time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" redis:"updated_at"`
	CreatedBy       uuid.UUID              `json:"created_by" redis:"created_by"`
	UpdatedBy       uuid.UUID              `json:"updated_by" redis:"updated_by"`
	VersionNumber   int64                  `json:"version_number" redis:"version_number"`
}

// WorkflowType defines the type of workflow
type WorkflowType string

const (
	WorkflowTypeSequential WorkflowType = "sequential"
	WorkflowTypeParallel   WorkflowType = "parallel"
	WorkflowTypeConditional WorkflowType = "conditional"
	WorkflowTypeLoop       WorkflowType = "loop"
	WorkflowTypeEvent      WorkflowType = "event_driven"
	WorkflowTypeHybrid     WorkflowType = "hybrid"
)

// WorkflowStatus defines the status of a workflow
type WorkflowStatus string

const (
	WorkflowStatusDraft     WorkflowStatus = "draft"
	WorkflowStatusActive    WorkflowStatus = "active"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusPaused    WorkflowStatus = "paused"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
	WorkflowStatusArchived  WorkflowStatus = "archived"
)

// WorkflowPriority defines the priority of a workflow
type WorkflowPriority string

const (
	WorkflowPriorityLow      WorkflowPriority = "low"
	WorkflowPriorityMedium   WorkflowPriority = "medium"
	WorkflowPriorityHigh     WorkflowPriority = "high"
	WorkflowPriorityCritical WorkflowPriority = "critical"
	WorkflowPriorityUrgent   WorkflowPriority = "urgent"
)

// WorkflowCategory defines the category of a workflow
type WorkflowCategory string

const (
	WorkflowCategoryContentCreation WorkflowCategory = "content_creation"
	WorkflowCategoryContentPublishing WorkflowCategory = "content_publishing"
	WorkflowCategoryCampaignManagement WorkflowCategory = "campaign_management"
	WorkflowCategoryAnalytics WorkflowCategory = "analytics"
	WorkflowCategoryFeedbackProcessing WorkflowCategory = "feedback_processing"
	WorkflowCategoryCustomerService WorkflowCategory = "customer_service"
	WorkflowCategoryMarketing WorkflowCategory = "marketing"
	WorkflowCategoryOperations WorkflowCategory = "operations"
)

// WorkflowDefinition represents the structure and logic of a workflow
type WorkflowDefinition struct {
	StartStep       string                 `json:"start_step"`
	EndSteps        []string               `json:"end_steps"`
	Steps           map[string]*StepDefinition `json:"steps"`
	Connections     []*StepConnection      `json:"connections"`
	ErrorHandling   *ErrorHandlingConfig   `json:"error_handling,omitempty"`
	Timeouts        map[string]time.Duration `json:"timeouts,omitempty"`
	RetryPolicies   map[string]*RetryPolicy `json:"retry_policies,omitempty"`
	Conditions      map[string]*Condition  `json:"conditions,omitempty"`
	Variables       map[string]*VariableDefinition `json:"variables,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// StepDefinition defines a single step in a workflow
type StepDefinition struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Type            StepType               `json:"type"`
	AgentType       string                 `json:"agent_type,omitempty"`
	Action          string                 `json:"action"`
	Parameters      map[string]interface{} `json:"parameters,omitempty"`
	InputMapping    map[string]string      `json:"input_mapping,omitempty"`
	OutputMapping   map[string]string      `json:"output_mapping,omitempty"`
	Conditions      []*Condition           `json:"conditions,omitempty"`
	Timeout         time.Duration          `json:"timeout,omitempty"`
	RetryPolicy     *RetryPolicy           `json:"retry_policy,omitempty"`
	ErrorHandling   *ErrorHandlingConfig   `json:"error_handling,omitempty"`
	Dependencies    []string               `json:"dependencies,omitempty"`
	IsOptional      bool                   `json:"is_optional"`
	IsParallel      bool                   `json:"is_parallel"`
	Position        *StepPosition          `json:"position,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// StepType defines the type of workflow step
type StepType string

const (
	StepTypeAgent       StepType = "agent"
	StepTypeCondition   StepType = "condition"
	StepTypeLoop        StepType = "loop"
	StepTypeParallel    StepType = "parallel"
	StepTypeWait        StepType = "wait"
	StepTypeNotification StepType = "notification"
	StepTypeTransform   StepType = "transform"
	StepTypeValidation  StepType = "validation"
	StepTypeDecision    StepType = "decision"
	StepTypeSubWorkflow StepType = "sub_workflow"
)

// StepConnection defines connections between workflow steps
type StepConnection struct {
	FromStep    string     `json:"from_step"`
	ToStep      string     `json:"to_step"`
	Condition   *Condition `json:"condition,omitempty"`
	Label       string     `json:"label,omitempty"`
	IsDefault   bool       `json:"is_default"`
	Priority    int        `json:"priority"`
}

// Condition represents a conditional expression
type Condition struct {
	Expression string                 `json:"expression"`
	Variables  map[string]interface{} `json:"variables,omitempty"`
	Operator   ConditionOperator      `json:"operator"`
	Value      interface{}            `json:"value,omitempty"`
	Conditions []*Condition           `json:"conditions,omitempty"` // For nested conditions
}

// ConditionOperator defines condition operators
type ConditionOperator string

const (
	ConditionOperatorEquals          ConditionOperator = "equals"
	ConditionOperatorNotEquals       ConditionOperator = "not_equals"
	ConditionOperatorGreaterThan     ConditionOperator = "greater_than"
	ConditionOperatorLessThan        ConditionOperator = "less_than"
	ConditionOperatorGreaterOrEqual  ConditionOperator = "greater_or_equal"
	ConditionOperatorLessOrEqual     ConditionOperator = "less_or_equal"
	ConditionOperatorContains        ConditionOperator = "contains"
	ConditionOperatorNotContains     ConditionOperator = "not_contains"
	ConditionOperatorIn              ConditionOperator = "in"
	ConditionOperatorNotIn           ConditionOperator = "not_in"
	ConditionOperatorAnd             ConditionOperator = "and"
	ConditionOperatorOr              ConditionOperator = "or"
	ConditionOperatorNot             ConditionOperator = "not"
)

// RetryPolicy defines retry behavior for failed steps
type RetryPolicy struct {
	MaxAttempts     int           `json:"max_attempts"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors,omitempty"`
	StopOnErrors    []string      `json:"stop_on_errors,omitempty"`
}

// ErrorHandlingConfig defines error handling behavior
type ErrorHandlingConfig struct {
	Strategy        ErrorHandlingStrategy `json:"strategy"`
	FallbackStep    string                `json:"fallback_step,omitempty"`
	NotifyOnError   bool                  `json:"notify_on_error"`
	ContinueOnError bool                  `json:"continue_on_error"`
	ErrorMapping    map[string]string     `json:"error_mapping,omitempty"`
}

// ErrorHandlingStrategy defines error handling strategies
type ErrorHandlingStrategy string

const (
	ErrorHandlingStrategyStop     ErrorHandlingStrategy = "stop"
	ErrorHandlingStrategyContinue ErrorHandlingStrategy = "continue"
	ErrorHandlingStrategyRetry    ErrorHandlingStrategy = "retry"
	ErrorHandlingStrategyFallback ErrorHandlingStrategy = "fallback"
	ErrorHandlingStrategyEscalate ErrorHandlingStrategy = "escalate"
)

// VariableDefinition defines a workflow variable
type VariableDefinition struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Required     bool        `json:"required"`
	Description  string      `json:"description,omitempty"`
	Validation   *Validation `json:"validation,omitempty"`
}

// Validation defines validation rules for variables
type Validation struct {
	Pattern     string      `json:"pattern,omitempty"`
	MinLength   int         `json:"min_length,omitempty"`
	MaxLength   int         `json:"max_length,omitempty"`
	MinValue    interface{} `json:"min_value,omitempty"`
	MaxValue    interface{} `json:"max_value,omitempty"`
	AllowedValues []interface{} `json:"allowed_values,omitempty"`
}

// StepPosition defines the visual position of a step in workflow designer
type StepPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// WorkflowStep represents a step instance in a workflow execution
type WorkflowStep struct {
	ID              uuid.UUID              `json:"id"`
	WorkflowID      uuid.UUID              `json:"workflow_id"`
	ExecutionID     uuid.UUID              `json:"execution_id"`
	StepDefinitionID string                `json:"step_definition_id"`
	Name            string                 `json:"name"`
	Status          StepStatus             `json:"status"`
	StartedAt       *time.Time             `json:"started_at,omitempty"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	Duration        time.Duration          `json:"duration"`
	Input           map[string]interface{} `json:"input,omitempty"`
	Output          map[string]interface{} `json:"output,omitempty"`
	Error           *StepError             `json:"error,omitempty"`
	RetryCount      int                    `json:"retry_count"`
	Logs            []*StepLog             `json:"logs,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// StepStatus defines the status of a workflow step
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
	StepStatusCancelled StepStatus = "cancelled"
	StepStatusRetrying  StepStatus = "retrying"
	StepStatusWaiting   StepStatus = "waiting"
)

// StepError represents an error that occurred during step execution
type StepError struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     string                 `json:"details,omitempty"`
	Retryable   bool                   `json:"retryable"`
	Timestamp   time.Time              `json:"timestamp"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// StepLog represents a log entry for a workflow step
type StepLog struct {
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// WorkflowExecution represents an execution instance of a workflow
type WorkflowExecution struct {
	ID              uuid.UUID              `json:"id"`
	WorkflowID      uuid.UUID              `json:"workflow_id"`
	Status          WorkflowStatus         `json:"status"`
	TriggerType     string                 `json:"trigger_type"`
	TriggerData     map[string]interface{} `json:"trigger_data,omitempty"`
	Input           map[string]interface{} `json:"input,omitempty"`
	Output          map[string]interface{} `json:"output,omitempty"`
	Variables       map[string]interface{} `json:"variables,omitempty"`
	CurrentStep     string                 `json:"current_step,omitempty"`
	CompletedSteps  []string               `json:"completed_steps,omitempty"`
	FailedSteps     []string               `json:"failed_steps,omitempty"`
	StartedAt       time.Time              `json:"started_at"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	Duration        time.Duration          `json:"duration"`
	Error           *WorkflowError         `json:"error,omitempty"`
	Metrics         *ExecutionMetrics      `json:"metrics,omitempty"`
	Logs            []*ExecutionLog        `json:"logs,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	CreatedBy       uuid.UUID              `json:"created_by"`
}

// WorkflowError represents an error that occurred during workflow execution
type WorkflowError struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     string                 `json:"details,omitempty"`
	StepID      string                 `json:"step_id,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// ExecutionMetrics represents metrics for a workflow execution
type ExecutionMetrics struct {
	TotalSteps      int           `json:"total_steps"`
	CompletedSteps  int           `json:"completed_steps"`
	FailedSteps     int           `json:"failed_steps"`
	SkippedSteps    int           `json:"skipped_steps"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageStepTime time.Duration `json:"average_step_time"`
	ResourceUsage   *ResourceUsage `json:"resource_usage,omitempty"`
}

// ResourceUsage represents resource usage metrics
type ResourceUsage struct {
	CPUTime    time.Duration `json:"cpu_time"`
	MemoryPeak int64         `json:"memory_peak"`
	NetworkIO  int64         `json:"network_io"`
	DiskIO     int64         `json:"disk_io"`
}

// ExecutionLog represents a log entry for a workflow execution
type ExecutionLog struct {
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	StepID    string                 `json:"step_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// WorkflowTrigger represents a trigger that can start a workflow
type WorkflowTrigger struct {
	ID          uuid.UUID              `json:"id"`
	WorkflowID  uuid.UUID              `json:"workflow_id"`
	Name        string                 `json:"name"`
	Type        TriggerType            `json:"type"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	Conditions  []*Condition           `json:"conditions,omitempty"`
	IsActive    bool                   `json:"is_active"`
	LastTriggered *time.Time           `json:"last_triggered,omitempty"`
	TriggerCount int64                 `json:"trigger_count"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TriggerType defines the type of workflow trigger
type TriggerType string

const (
	TriggerTypeSchedule TriggerType = "schedule"
	TriggerTypeEvent    TriggerType = "event"
	TriggerTypeWebhook  TriggerType = "webhook"
	TriggerTypeManual   TriggerType = "manual"
	TriggerTypeAPI      TriggerType = "api"
	TriggerTypeFile     TriggerType = "file"
	TriggerTypeDatabase TriggerType = "database"
	TriggerTypeQueue    TriggerType = "queue"
)

// WorkflowConfig represents configuration for a workflow
type WorkflowConfig struct {
	MaxConcurrentExecutions int                    `json:"max_concurrent_executions"`
	ExecutionTimeout        time.Duration          `json:"execution_timeout"`
	RetentionPeriod         time.Duration          `json:"retention_period"`
	NotificationSettings    *NotificationSettings  `json:"notification_settings,omitempty"`
	SecuritySettings        *SecuritySettings      `json:"security_settings,omitempty"`
	ResourceLimits          *ResourceLimits        `json:"resource_limits,omitempty"`
	Monitoring              *MonitoringConfig      `json:"monitoring,omitempty"`
}

// NotificationSettings defines notification configuration
type NotificationSettings struct {
	OnStart     bool     `json:"on_start"`
	OnComplete  bool     `json:"on_complete"`
	OnFailure   bool     `json:"on_failure"`
	OnError     bool     `json:"on_error"`
	Recipients  []string `json:"recipients,omitempty"`
	Channels    []string `json:"channels,omitempty"`
	Templates   map[string]string `json:"templates,omitempty"`
}

// SecuritySettings defines security configuration
type SecuritySettings struct {
	RequireApproval bool     `json:"require_approval"`
	AllowedUsers    []string `json:"allowed_users,omitempty"`
	AllowedRoles    []string `json:"allowed_roles,omitempty"`
	EncryptData     bool     `json:"encrypt_data"`
	AuditLevel      string   `json:"audit_level"`
}

// ResourceLimits defines resource limits for workflow execution
type ResourceLimits struct {
	MaxCPU    string `json:"max_cpu,omitempty"`
	MaxMemory string `json:"max_memory,omitempty"`
	MaxDisk   string `json:"max_disk,omitempty"`
	MaxNetwork string `json:"max_network,omitempty"`
}

// MonitoringConfig defines monitoring configuration
type MonitoringConfig struct {
	EnableMetrics bool     `json:"enable_metrics"`
	EnableTracing bool     `json:"enable_tracing"`
	EnableLogging bool     `json:"enable_logging"`
	MetricsTags   []string `json:"metrics_tags,omitempty"`
	LogLevel      string   `json:"log_level"`
}

// WorkflowMetrics represents metrics for a workflow
type WorkflowMetrics struct {
	TotalExecutions     int64         `json:"total_executions"`
	SuccessfulExecutions int64        `json:"successful_executions"`
	FailedExecutions    int64         `json:"failed_executions"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	LastExecutionTime   *time.Time    `json:"last_execution_time,omitempty"`
	SuccessRate         float64       `json:"success_rate"`
	ErrorRate           float64       `json:"error_rate"`
	ThroughputPerHour   float64       `json:"throughput_per_hour"`
	ResourceUtilization *ResourceUsage `json:"resource_utilization,omitempty"`
	LastUpdated         time.Time     `json:"last_updated"`
}

// NewWorkflow creates a new workflow instance
func NewWorkflow(name, description string, workflowType WorkflowType, createdBy uuid.UUID) *Workflow {
	now := time.Now()
	return &Workflow{
		ID:            uuid.New(),
		Name:          name,
		Description:   description,
		Type:          workflowType,
		Status:        WorkflowStatusDraft,
		Priority:      WorkflowPriorityMedium,
		Version:       "1.0.0",
		Variables:     make(map[string]interface{}),
		Tags:          []string{},
		IsActive:      false,
		IsTemplate:    false,
		CreatedAt:     now,
		UpdatedAt:     now,
		CreatedBy:     createdBy,
		UpdatedBy:     createdBy,
		VersionNumber: 1,
	}
}

// UpdateStatus updates the workflow status
func (w *Workflow) UpdateStatus(newStatus WorkflowStatus, updatedBy uuid.UUID) {
	w.Status = newStatus
	w.UpdatedBy = updatedBy
	w.UpdatedAt = time.Now()
	w.VersionNumber++
}

// AddStep adds a step to the workflow
func (w *Workflow) AddStep(step *WorkflowStep) {
	step.WorkflowID = w.ID
	w.Steps = append(w.Steps, step)
	w.UpdatedAt = time.Now()
	w.VersionNumber++
}

// AddTrigger adds a trigger to the workflow
func (w *Workflow) AddTrigger(trigger *WorkflowTrigger) {
	trigger.WorkflowID = w.ID
	w.Triggers = append(w.Triggers, trigger)
	w.UpdatedAt = time.Now()
	w.VersionNumber++
}

// IsExecutable checks if the workflow can be executed
func (w *Workflow) IsExecutable() bool {
	return w.IsActive && w.Status == WorkflowStatusActive && w.Definition != nil
}

// GetActiveExecutions returns the number of active executions
func (w *Workflow) GetActiveExecutions() int {
	count := 0
	for _, execution := range w.Executions {
		if execution.Status == WorkflowStatusRunning {
			count++
		}
	}
	return count
}

// CanExecute checks if a new execution can be started
func (w *Workflow) CanExecute() bool {
	if !w.IsExecutable() {
		return false
	}
	
	if w.Configuration != nil && w.Configuration.MaxConcurrentExecutions > 0 {
		return w.GetActiveExecutions() < w.Configuration.MaxConcurrentExecutions
	}
	
	return true
}
