package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/task-manager-agent/internal/domain/entities"
)

// WorkflowRepository defines the interface for workflow data access
type WorkflowRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, workflow *entities.Workflow) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Workflow, error)
	GetByName(ctx context.Context, name string) (*entities.Workflow, error)
	Update(ctx context.Context, workflow *entities.Workflow) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Listing and filtering
	List(ctx context.Context, filter *WorkflowFilter) ([]*entities.Workflow, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, filter *WorkflowFilter) ([]*entities.Workflow, error)
	ListByType(ctx context.Context, workflowType entities.WorkflowType, filter *WorkflowFilter) ([]*entities.Workflow, error)
	ListByCategory(ctx context.Context, category entities.WorkflowCategory, filter *WorkflowFilter) ([]*entities.Workflow, error)
	ListActive(ctx context.Context, filter *WorkflowFilter) ([]*entities.Workflow, error)
	ListTemplates(ctx context.Context, filter *WorkflowFilter) ([]*entities.Workflow, error)
	
	// Steps
	AddStep(ctx context.Context, step *entities.WorkflowStep) error
	UpdateStep(ctx context.Context, step *entities.WorkflowStep) error
	DeleteStep(ctx context.Context, stepID uuid.UUID) error
	GetSteps(ctx context.Context, workflowID uuid.UUID) ([]*entities.WorkflowStep, error)
	GetStepByID(ctx context.Context, stepID uuid.UUID) (*entities.WorkflowStep, error)
	
	// Triggers
	AddTrigger(ctx context.Context, trigger *entities.WorkflowTrigger) error
	UpdateTrigger(ctx context.Context, trigger *entities.WorkflowTrigger) error
	DeleteTrigger(ctx context.Context, triggerID uuid.UUID) error
	GetTriggers(ctx context.Context, workflowID uuid.UUID) ([]*entities.WorkflowTrigger, error)
	GetTriggersByEvent(ctx context.Context, event string) ([]*entities.WorkflowTrigger, error)
	
	// Conditions and Actions
	AddCondition(ctx context.Context, condition *entities.WorkflowCondition) error
	UpdateCondition(ctx context.Context, condition *entities.WorkflowCondition) error
	DeleteCondition(ctx context.Context, conditionID uuid.UUID) error
	GetConditions(ctx context.Context, workflowID uuid.UUID) ([]*entities.WorkflowCondition, error)
	
	AddAction(ctx context.Context, action *entities.WorkflowAction) error
	UpdateAction(ctx context.Context, action *entities.WorkflowAction) error
	DeleteAction(ctx context.Context, actionID uuid.UUID) error
	GetActions(ctx context.Context, workflowID uuid.UUID) ([]*entities.WorkflowAction, error)
	
	// Executions
	CreateExecution(ctx context.Context, execution *entities.WorkflowExecution) error
	GetExecution(ctx context.Context, executionID uuid.UUID) (*entities.WorkflowExecution, error)
	UpdateExecution(ctx context.Context, execution *entities.WorkflowExecution) error
	ListExecutions(ctx context.Context, workflowID uuid.UUID, filter *ExecutionFilter) ([]*entities.WorkflowExecution, error)
	GetActiveExecutions(ctx context.Context, workflowID *uuid.UUID) ([]*entities.WorkflowExecution, error)
	
	// Step Executions
	CreateStepExecution(ctx context.Context, stepExecution *entities.StepExecution) error
	GetStepExecution(ctx context.Context, stepExecutionID uuid.UUID) (*entities.StepExecution, error)
	UpdateStepExecution(ctx context.Context, stepExecution *entities.StepExecution) error
	ListStepExecutions(ctx context.Context, executionID uuid.UUID) ([]*entities.StepExecution, error)
	GetPendingStepExecutions(ctx context.Context, userID *uuid.UUID) ([]*entities.StepExecution, error)
	
	// Analytics and metrics
	GetWorkflowMetrics(ctx context.Context, workflowID uuid.UUID, period time.Duration) (*WorkflowMetrics, error)
	GetExecutionMetrics(ctx context.Context, filter *ExecutionMetricsFilter) (*ExecutionMetrics, error)
	GetPerformanceMetrics(ctx context.Context, period time.Duration) (*WorkflowPerformanceMetrics, error)
	
	// Search
	Search(ctx context.Context, query string, filter *WorkflowFilter) ([]*entities.Workflow, error)
	
	// Bulk operations
	BulkUpdate(ctx context.Context, workflows []*entities.Workflow) error
	BulkActivate(ctx context.Context, workflowIDs []uuid.UUID, activatedBy uuid.UUID) error
	BulkDeactivate(ctx context.Context, workflowIDs []uuid.UUID, deactivatedBy uuid.UUID) error
	
	// Version management
	CreateVersion(ctx context.Context, workflowID uuid.UUID, version string, createdBy uuid.UUID) (*entities.Workflow, error)
	GetVersions(ctx context.Context, workflowID uuid.UUID) ([]*entities.Workflow, error)
	GetVersion(ctx context.Context, workflowID uuid.UUID, version string) (*entities.Workflow, error)
	
	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context, repo WorkflowRepository) error) error
}

// WorkflowFilter defines filtering options for workflow queries
type WorkflowFilter struct {
	OwnerIDs     []uuid.UUID                   `json:"owner_ids,omitempty"`
	Types        []entities.WorkflowType       `json:"types,omitempty"`
	Statuses     []entities.WorkflowStatus     `json:"statuses,omitempty"`
	Categories   []entities.WorkflowCategory   `json:"categories,omitempty"`
	Tags         []string                      `json:"tags,omitempty"`
	Labels       []string                      `json:"labels,omitempty"`
	IsActive     *bool                         `json:"is_active,omitempty"`
	IsTemplate   *bool                         `json:"is_template,omitempty"`
	CreatedAfter *time.Time                    `json:"created_after,omitempty"`
	CreatedBefore *time.Time                   `json:"created_before,omitempty"`
	UpdatedAfter *time.Time                    `json:"updated_after,omitempty"`
	UpdatedBefore *time.Time                   `json:"updated_before,omitempty"`
	Versions     []string                      `json:"versions,omitempty"`
	SortBy       string                        `json:"sort_by,omitempty"`
	SortOrder    string                        `json:"sort_order,omitempty"`
	Limit        int                           `json:"limit,omitempty"`
	Offset       int                           `json:"offset,omitempty"`
}

// ExecutionFilter defines filtering options for execution queries
type ExecutionFilter struct {
	WorkflowIDs  []uuid.UUID                   `json:"workflow_ids,omitempty"`
	TriggerIDs   []uuid.UUID                   `json:"trigger_ids,omitempty"`
	Statuses     []entities.ExecutionStatus    `json:"statuses,omitempty"`
	ExecutedBy   []uuid.UUID                   `json:"executed_by,omitempty"`
	StartedAfter *time.Time                    `json:"started_after,omitempty"`
	StartedBefore *time.Time                   `json:"started_before,omitempty"`
	CompletedAfter *time.Time                  `json:"completed_after,omitempty"`
	CompletedBefore *time.Time                 `json:"completed_before,omitempty"`
	HasErrors    *bool                         `json:"has_errors,omitempty"`
	MinDuration  *time.Duration                `json:"min_duration,omitempty"`
	MaxDuration  *time.Duration                `json:"max_duration,omitempty"`
	SortBy       string                        `json:"sort_by,omitempty"`
	SortOrder    string                        `json:"sort_order,omitempty"`
	Limit        int                           `json:"limit,omitempty"`
	Offset       int                           `json:"offset,omitempty"`
}

// ExecutionMetricsFilter defines filtering options for execution metrics
type ExecutionMetricsFilter struct {
	WorkflowIDs []uuid.UUID   `json:"workflow_ids,omitempty"`
	Period      time.Duration `json:"period"`
	StartDate   time.Time     `json:"start_date"`
	EndDate     time.Time     `json:"end_date"`
	GroupBy     string        `json:"group_by,omitempty"`
}

// WorkflowMetrics contains workflow-specific metrics
type WorkflowMetrics struct {
	WorkflowID          uuid.UUID `json:"workflow_id"`
	Period              string    `json:"period"`
	TotalExecutions     int       `json:"total_executions"`
	SuccessfulExecutions int      `json:"successful_executions"`
	FailedExecutions    int       `json:"failed_executions"`
	CancelledExecutions int       `json:"cancelled_executions"`
	SuccessRate         float64   `json:"success_rate"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	MedianExecutionTime time.Duration `json:"median_execution_time"`
	MinExecutionTime    time.Duration `json:"min_execution_time"`
	MaxExecutionTime    time.Duration `json:"max_execution_time"`
	TotalSteps          int       `json:"total_steps"`
	AverageStepsPerExecution float64 `json:"average_steps_per_execution"`
	MostFailedStep      string    `json:"most_failed_step"`
	BottleneckStep      string    `json:"bottleneck_step"`
	TriggerBreakdown    map[string]int `json:"trigger_breakdown"`
	ExecutionTrend      []ExecutionTrendPoint `json:"execution_trend"`
	GeneratedAt         time.Time `json:"generated_at"`
}

// ExecutionMetrics contains execution metrics across workflows
type ExecutionMetrics struct {
	Period              string    `json:"period"`
	TotalExecutions     int       `json:"total_executions"`
	SuccessfulExecutions int      `json:"successful_executions"`
	FailedExecutions    int       `json:"failed_executions"`
	CancelledExecutions int       `json:"cancelled_executions"`
	PendingExecutions   int       `json:"pending_executions"`
	RunningExecutions   int       `json:"running_executions"`
	SuccessRate         float64   `json:"success_rate"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	TotalExecutionTime  time.Duration `json:"total_execution_time"`
	ExecutionsByWorkflow map[uuid.UUID]int `json:"executions_by_workflow"`
	ExecutionsByStatus  map[entities.ExecutionStatus]int `json:"executions_by_status"`
	ExecutionsByTrigger map[string]int `json:"executions_by_trigger"`
	TopFailingWorkflows []WorkflowFailureInfo `json:"top_failing_workflows"`
	TopSlowWorkflows    []WorkflowPerformanceInfo `json:"top_slow_workflows"`
	HourlyDistribution  map[int]int `json:"hourly_distribution"`
	DailyDistribution   map[string]int `json:"daily_distribution"`
	GeneratedAt         time.Time `json:"generated_at"`
}

// WorkflowPerformanceMetrics contains overall workflow performance metrics
type WorkflowPerformanceMetrics struct {
	Period                  string    `json:"period"`
	TotalWorkflows          int       `json:"total_workflows"`
	ActiveWorkflows         int       `json:"active_workflows"`
	InactiveWorkflows       int       `json:"inactive_workflows"`
	TotalExecutions         int       `json:"total_executions"`
	AverageExecutionsPerWorkflow float64 `json:"average_executions_per_workflow"`
	SystemThroughput        float64   `json:"system_throughput"`
	SystemUtilization       float64   `json:"system_utilization"`
	ErrorRate               float64   `json:"error_rate"`
	AverageResponseTime     time.Duration `json:"average_response_time"`
	WorkflowsByType         map[entities.WorkflowType]int `json:"workflows_by_type"`
	WorkflowsByCategory     map[entities.WorkflowCategory]int `json:"workflows_by_category"`
	MostActiveWorkflows     []WorkflowActivityInfo `json:"most_active_workflows"`
	LeastActiveWorkflows    []WorkflowActivityInfo `json:"least_active_workflows"`
	ResourceUtilization     map[string]float64 `json:"resource_utilization"`
	GeneratedAt             time.Time `json:"generated_at"`
}

// Supporting metric types
type ExecutionTrendPoint struct {
	Date        time.Time `json:"date"`
	Executions  int       `json:"executions"`
	Successes   int       `json:"successes"`
	Failures    int       `json:"failures"`
	AvgDuration time.Duration `json:"avg_duration"`
}

type WorkflowFailureInfo struct {
	WorkflowID   uuid.UUID `json:"workflow_id"`
	WorkflowName string    `json:"workflow_name"`
	FailureCount int       `json:"failure_count"`
	FailureRate  float64   `json:"failure_rate"`
	LastFailure  time.Time `json:"last_failure"`
	CommonErrors []string  `json:"common_errors"`
}

type WorkflowPerformanceInfo struct {
	WorkflowID      uuid.UUID     `json:"workflow_id"`
	WorkflowName    string        `json:"workflow_name"`
	AverageTime     time.Duration `json:"average_time"`
	MedianTime      time.Duration `json:"median_time"`
	MaxTime         time.Duration `json:"max_time"`
	ExecutionCount  int           `json:"execution_count"`
	BottleneckSteps []string      `json:"bottleneck_steps"`
}

type WorkflowActivityInfo struct {
	WorkflowID     uuid.UUID `json:"workflow_id"`
	WorkflowName   string    `json:"workflow_name"`
	ExecutionCount int       `json:"execution_count"`
	LastExecution  time.Time `json:"last_execution"`
	SuccessRate    float64   `json:"success_rate"`
	AverageTime    time.Duration `json:"average_time"`
}

// NotificationRepository defines the interface for notification data access
type NotificationRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, notification *Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	Update(ctx context.Context, notification *Notification) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Listing and filtering
	List(ctx context.Context, filter *NotificationFilter) ([]*Notification, error)
	ListByUser(ctx context.Context, userID uuid.UUID, filter *NotificationFilter) ([]*Notification, error)
	ListUnread(ctx context.Context, userID uuid.UUID, filter *NotificationFilter) ([]*Notification, error)
	
	// Status management
	MarkAsRead(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) error
	MarkAsUnread(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	
	// Bulk operations
	BulkCreate(ctx context.Context, notifications []*Notification) error
	BulkMarkAsRead(ctx context.Context, notificationIDs []uuid.UUID, userID uuid.UUID) error
	BulkDelete(ctx context.Context, notificationIDs []uuid.UUID) error
	
	// Cleanup
	DeleteOldNotifications(ctx context.Context, olderThan time.Time) error
	
	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context, repo NotificationRepository) error) error
}

// Notification represents a notification entity
type Notification struct {
	ID          uuid.UUID              `json:"id" redis:"id"`
	Type        NotificationType       `json:"type" redis:"type"`
	Title       string                 `json:"title" redis:"title"`
	Message     string                 `json:"message" redis:"message"`
	Priority    NotificationPriority   `json:"priority" redis:"priority"`
	Category    NotificationCategory   `json:"category" redis:"category"`
	UserID      uuid.UUID              `json:"user_id" redis:"user_id"`
	RelatedID   *uuid.UUID             `json:"related_id,omitempty" redis:"related_id"`
	RelatedType string                 `json:"related_type" redis:"related_type"`
	Data        map[string]interface{} `json:"data" redis:"data"`
	IsRead      bool                   `json:"is_read" redis:"is_read"`
	ReadAt      *time.Time             `json:"read_at,omitempty" redis:"read_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty" redis:"expires_at"`
	CreatedAt   time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" redis:"updated_at"`
}

// NotificationType defines the type of notification
type NotificationType string

const (
	NotificationTypeTaskAssigned    NotificationType = "task_assigned"
	NotificationTypeTaskDue         NotificationType = "task_due"
	NotificationTypeTaskOverdue     NotificationType = "task_overdue"
	NotificationTypeTaskCompleted   NotificationType = "task_completed"
	NotificationTypeProjectUpdate   NotificationType = "project_update"
	NotificationTypeWorkflowStarted NotificationType = "workflow_started"
	NotificationTypeWorkflowFailed  NotificationType = "workflow_failed"
	NotificationTypeComment         NotificationType = "comment"
	NotificationTypeMention         NotificationType = "mention"
	NotificationTypeReminder        NotificationType = "reminder"
	NotificationTypeSystem          NotificationType = "system"
)

// NotificationPriority defines the priority of notification
type NotificationPriority string

const (
	NotificationPriorityLow      NotificationPriority = "low"
	NotificationPriorityNormal   NotificationPriority = "normal"
	NotificationPriorityHigh     NotificationPriority = "high"
	NotificationPriorityCritical NotificationPriority = "critical"
)

// NotificationCategory defines the category of notification
type NotificationCategory string

const (
	NotificationCategoryTask     NotificationCategory = "task"
	NotificationCategoryProject  NotificationCategory = "project"
	NotificationCategoryWorkflow NotificationCategory = "workflow"
	NotificationCategorySystem   NotificationCategory = "system"
	NotificationCategorySocial   NotificationCategory = "social"
)

// NotificationFilter defines filtering options for notification queries
type NotificationFilter struct {
	UserIDs       []uuid.UUID                `json:"user_ids,omitempty"`
	Types         []NotificationType         `json:"types,omitempty"`
	Priorities    []NotificationPriority     `json:"priorities,omitempty"`
	Categories    []NotificationCategory     `json:"categories,omitempty"`
	IsRead        *bool                      `json:"is_read,omitempty"`
	RelatedTypes  []string                   `json:"related_types,omitempty"`
	RelatedIDs    []uuid.UUID                `json:"related_ids,omitempty"`
	CreatedAfter  *time.Time                 `json:"created_after,omitempty"`
	CreatedBefore *time.Time                 `json:"created_before,omitempty"`
	ExpiresAfter  *time.Time                 `json:"expires_after,omitempty"`
	ExpiresBefore *time.Time                 `json:"expires_before,omitempty"`
	SortBy        string                     `json:"sort_by,omitempty"`
	SortOrder     string                     `json:"sort_order,omitempty"`
	Limit         int                        `json:"limit,omitempty"`
	Offset        int                        `json:"offset,omitempty"`
}
