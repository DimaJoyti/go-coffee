package events

import (
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/task-manager-agent/internal/domain/entities"
)

// BaseDomainEvent provides common functionality for domain events
type BaseDomainEvent struct {
	EventType   string                 `json:"event_type"`
	AggregateID uuid.UUID              `json:"aggregate_id"`
	EventData   map[string]interface{} `json:"event_data"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     int                    `json:"version"`
	UserID      *uuid.UUID             `json:"user_id,omitempty"`
	CorrelationID *uuid.UUID           `json:"correlation_id,omitempty"`
}

// GetEventType returns the event type
func (e *BaseDomainEvent) GetEventType() string {
	return e.EventType
}

// GetAggregateID returns the aggregate ID
func (e *BaseDomainEvent) GetAggregateID() uuid.UUID {
	return e.AggregateID
}

// GetEventData returns the event data
func (e *BaseDomainEvent) GetEventData() map[string]interface{} {
	return e.EventData
}

// GetTimestamp returns the event timestamp
func (e *BaseDomainEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetVersion returns the event version
func (e *BaseDomainEvent) GetVersion() int {
	return e.Version
}

// Task Events

// TaskCreatedEvent represents a task creation event
type TaskCreatedEvent struct {
	BaseDomainEvent
	Task *entities.Task `json:"task"`
}

// NewTaskCreatedEvent creates a new task created event
func NewTaskCreatedEvent(task *entities.Task) *TaskCreatedEvent {
	return &TaskCreatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventType:   "task.created",
			AggregateID: task.ID,
			EventData: map[string]interface{}{
				"task_id":     task.ID,
				"title":       task.Title,
				"type":        task.Type,
				"status":      task.Status,
				"priority":    task.Priority,
				"project_id":  task.ProjectID,
				"created_by":  task.CreatedBy,
				"created_at":  task.CreatedAt,
			},
			Timestamp: time.Now(),
			Version:   1,
			UserID:    &task.CreatedBy,
		},
		Task: task,
	}
}

// TaskUpdatedEvent represents a task update event
type TaskUpdatedEvent struct {
	BaseDomainEvent
	Task        *entities.Task `json:"task"`
	PreviousTask *entities.Task `json:"previous_task"`
	Changes     map[string]interface{} `json:"changes"`
}

// NewTaskUpdatedEvent creates a new task updated event
func NewTaskUpdatedEvent(task, previousTask *entities.Task) *TaskUpdatedEvent {
	changes := make(map[string]interface{})
	
	// Track what changed
	if task.Title != previousTask.Title {
		changes["title"] = map[string]interface{}{
			"old": previousTask.Title,
			"new": task.Title,
		}
	}
	
	if task.Status != previousTask.Status {
		changes["status"] = map[string]interface{}{
			"old": previousTask.Status,
			"new": task.Status,
		}
	}
	
	if task.Priority != previousTask.Priority {
		changes["priority"] = map[string]interface{}{
			"old": previousTask.Priority,
			"new": task.Priority,
		}
	}
	
	if task.ProgressPercent != previousTask.ProgressPercent {
		changes["progress"] = map[string]interface{}{
			"old": previousTask.ProgressPercent,
			"new": task.ProgressPercent,
		}
	}

	return &TaskUpdatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventType:   "task.updated",
			AggregateID: task.ID,
			EventData: map[string]interface{}{
				"task_id":    task.ID,
				"title":      task.Title,
				"status":     task.Status,
				"priority":   task.Priority,
				"project_id": task.ProjectID,
				"updated_by": task.UpdatedBy,
				"updated_at": task.UpdatedAt,
				"changes":    changes,
			},
			Timestamp: time.Now(),
			Version:   int(task.Version),
			UserID:    &task.UpdatedBy,
		},
		Task:        task,
		PreviousTask: previousTask,
		Changes:     changes,
	}
}

// TaskAssignedEvent represents a task assignment event
type TaskAssignedEvent struct {
	BaseDomainEvent
	Task       *entities.Task           `json:"task"`
	Assignment *entities.TaskAssignment `json:"assignment"`
}

// NewTaskAssignedEvent creates a new task assigned event
func NewTaskAssignedEvent(task *entities.Task, assignment *entities.TaskAssignment) *TaskAssignedEvent {
	return &TaskAssignedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventType:   "task.assigned",
			AggregateID: task.ID,
			EventData: map[string]interface{}{
				"task_id":       task.ID,
				"title":         task.Title,
				"assignee_id":   assignment.UserID,
				"role":          assignment.Role,
				"assigned_by":   assignment.AssignedBy,
				"assigned_at":   assignment.AssignedAt,
				"allocation":    assignment.Allocation,
			},
			Timestamp: time.Now(),
			Version:   int(task.Version),
			UserID:    &assignment.AssignedBy,
		},
		Task:       task,
		Assignment: assignment,
	}
}

// TaskCompletedEvent represents a task completion event
type TaskCompletedEvent struct {
	BaseDomainEvent
	Task           *entities.Task `json:"task"`
	CompletionTime time.Duration  `json:"completion_time"`
}

// NewTaskCompletedEvent creates a new task completed event
func NewTaskCompletedEvent(task *entities.Task) *TaskCompletedEvent {
	var completionTime time.Duration
	if task.StartDate != nil && task.CompletedDate != nil {
		completionTime = task.CompletedDate.Sub(*task.StartDate)
	}

	return &TaskCompletedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventType:   "task.completed",
			AggregateID: task.ID,
			EventData: map[string]interface{}{
				"task_id":         task.ID,
				"title":           task.Title,
				"project_id":      task.ProjectID,
				"completed_at":    task.CompletedDate,
				"completion_time": completionTime,
				"actual_hours":    task.ActualHours,
				"estimated_hours": task.EstimatedHours,
				"efficiency":      task.EstimatedHours / task.ActualHours,
			},
			Timestamp: time.Now(),
			Version:   int(task.Version),
			UserID:    &task.UpdatedBy,
		},
		Task:           task,
		CompletionTime: completionTime,
	}
}

// TaskOverdueEvent represents a task overdue event
type TaskOverdueEvent struct {
	BaseDomainEvent
	Task        *entities.Task `json:"task"`
	DaysOverdue int            `json:"days_overdue"`
}

// NewTaskOverdueEvent creates a new task overdue event
func NewTaskOverdueEvent(task *entities.Task) *TaskOverdueEvent {
	var daysOverdue int
	if task.DueDate != nil {
		daysOverdue = int(time.Since(*task.DueDate).Hours() / 24)
	}

	return &TaskOverdueEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventType:   "task.overdue",
			AggregateID: task.ID,
			EventData: map[string]interface{}{
				"task_id":      task.ID,
				"title":        task.Title,
				"due_date":     task.DueDate,
				"days_overdue": daysOverdue,
				"project_id":   task.ProjectID,
				"assignees":    task.GetAssignees(),
			},
			Timestamp: time.Now(),
			Version:   int(task.Version),
		},
		Task:        task,
		DaysOverdue: daysOverdue,
	}
}

// TaskCommentAddedEvent represents a task comment added event
type TaskCommentAddedEvent struct {
	BaseDomainEvent
	Task    *entities.Task        `json:"task"`
	Comment *entities.TaskComment `json:"comment"`
}

// NewTaskCommentAddedEvent creates a new task comment added event
func NewTaskCommentAddedEvent(task *entities.Task, comment *entities.TaskComment) *TaskCommentAddedEvent {
	return &TaskCommentAddedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventType:   "task.comment.added",
			AggregateID: task.ID,
			EventData: map[string]interface{}{
				"task_id":    task.ID,
				"comment_id": comment.ID,
				"user_id":    comment.UserID,
				"content":    comment.Content,
				"created_at": comment.CreatedAt,
			},
			Timestamp: time.Now(),
			Version:   int(task.Version),
			UserID:    &comment.UserID,
		},
		Task:    task,
		Comment: comment,
	}
}

// Project Events

// ProjectCreatedEvent represents a project creation event
type ProjectCreatedEvent struct {
	BaseDomainEvent
	Project *entities.Project `json:"project"`
}

// NewProjectCreatedEvent creates a new project created event
func NewProjectCreatedEvent(project *entities.Project) *ProjectCreatedEvent {
	return &ProjectCreatedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventType:   "project.created",
			AggregateID: project.ID,
			EventData: map[string]interface{}{
				"project_id":  project.ID,
				"name":        project.Name,
				"code":        project.Code,
				"type":        project.Type,
				"status":      project.Status,
				"priority":    project.Priority,
				"owner_id":    project.OwnerID,
				"created_by":  project.CreatedBy,
				"created_at":  project.CreatedAt,
			},
			Timestamp: time.Now(),
			Version:   1,
			UserID:    &project.CreatedBy,
		},
		Project: project,
	}
}

// ProjectStatusChangedEvent represents a project status change event
type ProjectStatusChangedEvent struct {
	BaseDomainEvent
	Project       *entities.Project        `json:"project"`
	PreviousStatus entities.ProjectStatus  `json:"previous_status"`
	NewStatus     entities.ProjectStatus   `json:"new_status"`
}

// NewProjectStatusChangedEvent creates a new project status changed event
func NewProjectStatusChangedEvent(project *entities.Project, previousStatus entities.ProjectStatus) *ProjectStatusChangedEvent {
	return &ProjectStatusChangedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventType:   "project.status.changed",
			AggregateID: project.ID,
			EventData: map[string]interface{}{
				"project_id":      project.ID,
				"name":            project.Name,
				"previous_status": previousStatus,
				"new_status":      project.Status,
				"updated_by":      project.UpdatedBy,
				"updated_at":      project.UpdatedAt,
			},
			Timestamp: time.Now(),
			Version:   int(project.Version),
			UserID:    &project.UpdatedBy,
		},
		Project:       project,
		PreviousStatus: previousStatus,
		NewStatus:     project.Status,
	}
}

// ProjectMemberAddedEvent represents a project member addition event
type ProjectMemberAddedEvent struct {
	BaseDomainEvent
	Project *entities.Project       `json:"project"`
	Member  *entities.ProjectMember `json:"member"`
}

// NewProjectMemberAddedEvent creates a new project member added event
func NewProjectMemberAddedEvent(project *entities.Project, member *entities.ProjectMember) *ProjectMemberAddedEvent {
	return &ProjectMemberAddedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventType:   "project.member.added",
			AggregateID: project.ID,
			EventData: map[string]interface{}{
				"project_id": project.ID,
				"user_id":    member.UserID,
				"role":       member.Role,
				"allocation": member.Allocation,
				"joined_at":  member.JoinedAt,
			},
			Timestamp: time.Now(),
			Version:   int(project.Version),
		},
		Project: project,
		Member:  member,
	}
}

// Workflow Events

// WorkflowExecutionStartedEvent represents a workflow execution start event
type WorkflowExecutionStartedEvent struct {
	BaseDomainEvent
	Workflow  *entities.Workflow          `json:"workflow"`
	Execution *entities.WorkflowExecution `json:"execution"`
}

// NewWorkflowExecutionStartedEvent creates a new workflow execution started event
func NewWorkflowExecutionStartedEvent(workflow *entities.Workflow, execution *entities.WorkflowExecution) *WorkflowExecutionStartedEvent {
	return &WorkflowExecutionStartedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventType:   "workflow.execution.started",
			AggregateID: workflow.ID,
			EventData: map[string]interface{}{
				"workflow_id":   workflow.ID,
				"execution_id":  execution.ID,
				"workflow_name": workflow.Name,
				"executed_by":   execution.ExecutedBy,
				"started_at":    execution.StartedAt,
				"context":       execution.Context,
			},
			Timestamp: time.Now(),
			Version:   1,
			UserID:    &execution.ExecutedBy,
		},
		Workflow:  workflow,
		Execution: execution,
	}
}

// WorkflowExecutionCompletedEvent represents a workflow execution completion event
type WorkflowExecutionCompletedEvent struct {
	BaseDomainEvent
	Workflow       *entities.Workflow          `json:"workflow"`
	Execution      *entities.WorkflowExecution `json:"execution"`
	ExecutionTime  time.Duration               `json:"execution_time"`
}

// NewWorkflowExecutionCompletedEvent creates a new workflow execution completed event
func NewWorkflowExecutionCompletedEvent(workflow *entities.Workflow, execution *entities.WorkflowExecution) *WorkflowExecutionCompletedEvent {
	var executionTime time.Duration
	if execution.CompletedAt != nil {
		executionTime = execution.CompletedAt.Sub(execution.StartedAt)
	}

	return &WorkflowExecutionCompletedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventType:   "workflow.execution.completed",
			AggregateID: workflow.ID,
			EventData: map[string]interface{}{
				"workflow_id":    workflow.ID,
				"execution_id":   execution.ID,
				"workflow_name":  workflow.Name,
				"executed_by":    execution.ExecutedBy,
				"started_at":     execution.StartedAt,
				"completed_at":   execution.CompletedAt,
				"execution_time": executionTime,
				"status":         execution.Status,
			},
			Timestamp: time.Now(),
			Version:   1,
			UserID:    &execution.ExecutedBy,
		},
		Workflow:      workflow,
		Execution:     execution,
		ExecutionTime: executionTime,
	}
}

// WorkflowExecutionFailedEvent represents a workflow execution failure event
type WorkflowExecutionFailedEvent struct {
	BaseDomainEvent
	Workflow     *entities.Workflow          `json:"workflow"`
	Execution    *entities.WorkflowExecution `json:"execution"`
	ErrorMessage string                      `json:"error_message"`
}

// NewWorkflowExecutionFailedEvent creates a new workflow execution failed event
func NewWorkflowExecutionFailedEvent(workflow *entities.Workflow, execution *entities.WorkflowExecution, errorMessage string) *WorkflowExecutionFailedEvent {
	return &WorkflowExecutionFailedEvent{
		BaseDomainEvent: BaseDomainEvent{
			EventType:   "workflow.execution.failed",
			AggregateID: workflow.ID,
			EventData: map[string]interface{}{
				"workflow_id":    workflow.ID,
				"execution_id":   execution.ID,
				"workflow_name":  workflow.Name,
				"executed_by":    execution.ExecutedBy,
				"started_at":     execution.StartedAt,
				"failed_at":      execution.FailedAt,
				"error_message":  errorMessage,
				"retry_count":    execution.RetryCount,
			},
			Timestamp: time.Now(),
			Version:   1,
			UserID:    &execution.ExecutedBy,
		},
		Workflow:     workflow,
		Execution:    execution,
		ErrorMessage: errorMessage,
	}
}
