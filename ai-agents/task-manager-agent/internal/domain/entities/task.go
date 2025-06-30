package entities

import (
	"time"

	"github.com/google/uuid"
)

// Task represents a comprehensive task entity with all necessary attributes
type Task struct {
	ID                uuid.UUID              `json:"id" redis:"id"`
	Title             string                 `json:"title" redis:"title"`
	Description       string                 `json:"description" redis:"description"`
	Type              TaskType               `json:"type" redis:"type"`
	Status            TaskStatus             `json:"status" redis:"status"`
	Priority          TaskPriority           `json:"priority" redis:"priority"`
	Urgency           TaskUrgency            `json:"urgency" redis:"urgency"`
	Complexity        TaskComplexity         `json:"complexity" redis:"complexity"`
	ProjectID         *uuid.UUID             `json:"project_id,omitempty" redis:"project_id"`
	Project           *Project               `json:"project,omitempty"`
	WorkflowID        *uuid.UUID             `json:"workflow_id,omitempty" redis:"workflow_id"`
	Workflow          *Workflow              `json:"workflow,omitempty"`
	ParentTaskID      *uuid.UUID             `json:"parent_task_id,omitempty" redis:"parent_task_id"`
	ParentTask        *Task                  `json:"parent_task,omitempty"`
	SubTasks          []*Task                `json:"sub_tasks,omitempty"`
	Dependencies      []*TaskDependency      `json:"dependencies,omitempty"`
	Assignments       []*TaskAssignment      `json:"assignments,omitempty"`
	Tags              []string               `json:"tags" redis:"tags"`
	Labels            []string               `json:"labels" redis:"labels"`
	EstimatedHours    float64                `json:"estimated_hours" redis:"estimated_hours"`
	ActualHours       float64                `json:"actual_hours" redis:"actual_hours"`
	RemainingHours    float64                `json:"remaining_hours" redis:"remaining_hours"`
	ProgressPercent   float64                `json:"progress_percent" redis:"progress_percent"`
	StartDate         *time.Time             `json:"start_date,omitempty" redis:"start_date"`
	DueDate           *time.Time             `json:"due_date,omitempty" redis:"due_date"`
	CompletedDate     *time.Time             `json:"completed_date,omitempty" redis:"completed_date"`
	ScheduledStart    *time.Time             `json:"scheduled_start,omitempty" redis:"scheduled_start"`
	ScheduledEnd      *time.Time             `json:"scheduled_end,omitempty" redis:"scheduled_end"`
	Location          string                 `json:"location" redis:"location"`
	Equipment         []string               `json:"equipment" redis:"equipment"`
	Skills            []string               `json:"skills" redis:"skills"`
	Checklist         []*ChecklistItem       `json:"checklist,omitempty"`
	Attachments       []*TaskAttachment      `json:"attachments,omitempty"`
	Comments          []*TaskComment         `json:"comments,omitempty"`
	TimeEntries       []*TimeEntry           `json:"time_entries,omitempty"`
	CustomFields      map[string]interface{} `json:"custom_fields" redis:"custom_fields"`
	Metadata          map[string]interface{} `json:"metadata" redis:"metadata"`
	ExternalIDs       map[string]string      `json:"external_ids" redis:"external_ids"`
	IsRecurring       bool                   `json:"is_recurring" redis:"is_recurring"`
	RecurrenceRule    *RecurrenceRule        `json:"recurrence_rule,omitempty"`
	IsTemplate        bool                   `json:"is_template" redis:"is_template"`
	TemplateID        *uuid.UUID             `json:"template_id,omitempty" redis:"template_id"`
	IsArchived        bool                   `json:"is_archived" redis:"is_archived"`
	CreatedAt         time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" redis:"updated_at"`
	CreatedBy         uuid.UUID              `json:"created_by" redis:"created_by"`
	UpdatedBy         uuid.UUID              `json:"updated_by" redis:"updated_by"`
	Version           int64                  `json:"version" redis:"version"`
}

// TaskType defines the type of task
type TaskType string

const (
	TaskTypeGeneral     TaskType = "general"
	TaskTypeMaintenance TaskType = "maintenance"
	TaskTypeCleaning    TaskType = "cleaning"
	TaskTypeInventory   TaskType = "inventory"
	TaskTypeCustomer    TaskType = "customer"
	TaskTypeTraining    TaskType = "training"
	TaskTypeMarketing   TaskType = "marketing"
	TaskTypeFinance     TaskType = "finance"
	TaskTypeHR          TaskType = "hr"
	TaskTypeIT          TaskType = "it"
	TaskTypeQuality     TaskType = "quality"
	TaskTypeSafety      TaskType = "safety"
	TaskTypeCompliance  TaskType = "compliance"
	TaskTypeProject     TaskType = "project"
	TaskTypeBug         TaskType = "bug"
	TaskTypeFeature     TaskType = "feature"
	TaskTypeResearch    TaskType = "research"
)

// TaskStatus defines the status of a task
type TaskStatus string

const (
	StatusTodo        TaskStatus = "todo"
	StatusInProgress  TaskStatus = "in_progress"
	StatusInReview    TaskStatus = "in_review"
	StatusBlocked     TaskStatus = "blocked"
	StatusOnHold      TaskStatus = "on_hold"
	StatusCompleted   TaskStatus = "completed"
	StatusCancelled   TaskStatus = "cancelled"
	StatusDeferred    TaskStatus = "deferred"
	StatusWaitingFor  TaskStatus = "waiting_for"
	StatusApproval    TaskStatus = "approval"
	StatusTesting     TaskStatus = "testing"
	StatusDeployment  TaskStatus = "deployment"
)

// TaskPriority defines the priority of a task
type TaskPriority string

const (
	PriorityLowest  TaskPriority = "lowest"
	PriorityLow     TaskPriority = "low"
	PriorityMedium  TaskPriority = "medium"
	PriorityHigh    TaskPriority = "high"
	PriorityHighest TaskPriority = "highest"
	PriorityCritical TaskPriority = "critical"
)

// TaskUrgency defines the urgency of a task
type TaskUrgency string

const (
	UrgencyLow      TaskUrgency = "low"
	UrgencyMedium   TaskUrgency = "medium"
	UrgencyHigh     TaskUrgency = "high"
	UrgencyCritical TaskUrgency = "critical"
)

// TaskComplexity defines the complexity of a task
type TaskComplexity string

const (
	ComplexityTrivial TaskComplexity = "trivial"
	ComplexitySimple  TaskComplexity = "simple"
	ComplexityMedium  TaskComplexity = "medium"
	ComplexityComplex TaskComplexity = "complex"
	ComplexityExpert  TaskComplexity = "expert"
)

// TaskDependency represents a dependency between tasks
type TaskDependency struct {
	ID             uuid.UUID      `json:"id" redis:"id"`
	TaskID         uuid.UUID      `json:"task_id" redis:"task_id"`
	DependsOnID    uuid.UUID      `json:"depends_on_id" redis:"depends_on_id"`
	DependsOnTask  *Task          `json:"depends_on_task,omitempty"`
	Type           DependencyType `json:"type" redis:"type"`
	LagDays        int            `json:"lag_days" redis:"lag_days"`
	IsActive       bool           `json:"is_active" redis:"is_active"`
	CreatedAt      time.Time      `json:"created_at" redis:"created_at"`
	CreatedBy      uuid.UUID      `json:"created_by" redis:"created_by"`
}

// DependencyType defines the type of dependency
type DependencyType string

const (
	DependencyFinishToStart DependencyType = "finish_to_start"
	DependencyStartToStart  DependencyType = "start_to_start"
	DependencyFinishToFinish DependencyType = "finish_to_finish"
	DependencyStartToFinish DependencyType = "start_to_finish"
)

// TaskAssignment represents an assignment of a task to a user
type TaskAssignment struct {
	ID           uuid.UUID        `json:"id" redis:"id"`
	TaskID       uuid.UUID        `json:"task_id" redis:"task_id"`
	UserID       uuid.UUID        `json:"user_id" redis:"user_id"`
	User         *User            `json:"user,omitempty"`
	Role         AssignmentRole   `json:"role" redis:"role"`
	Allocation   float64          `json:"allocation" redis:"allocation"` // Percentage of time allocated
	AssignedAt   time.Time        `json:"assigned_at" redis:"assigned_at"`
	AssignedBy   uuid.UUID        `json:"assigned_by" redis:"assigned_by"`
	AcceptedAt   *time.Time       `json:"accepted_at,omitempty" redis:"accepted_at"`
	Status       AssignmentStatus `json:"status" redis:"status"`
	Notes        string           `json:"notes" redis:"notes"`
	IsActive     bool             `json:"is_active" redis:"is_active"`
}

// AssignmentRole defines the role of an assignment
type AssignmentRole string

const (
	RoleAssignee   AssignmentRole = "assignee"
	RoleReviewer   AssignmentRole = "reviewer"
	RoleApprover   AssignmentRole = "approver"
	RoleWatcher    AssignmentRole = "watcher"
	RoleCollaborator AssignmentRole = "collaborator"
)

// AssignmentStatus defines the status of an assignment
type AssignmentStatus string

const (
	AssignmentPending  AssignmentStatus = "pending"
	AssignmentAccepted AssignmentStatus = "accepted"
	AssignmentDeclined AssignmentStatus = "declined"
	AssignmentActive   AssignmentStatus = "active"
	AssignmentCompleted AssignmentStatus = "completed"
)

// ChecklistItem represents an item in a task checklist
type ChecklistItem struct {
	ID          uuid.UUID `json:"id" redis:"id"`
	TaskID      uuid.UUID `json:"task_id" redis:"task_id"`
	Title       string    `json:"title" redis:"title"`
	Description string    `json:"description" redis:"description"`
	IsCompleted bool      `json:"is_completed" redis:"is_completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty" redis:"completed_at"`
	CompletedBy *uuid.UUID `json:"completed_by,omitempty" redis:"completed_by"`
	Order       int       `json:"order" redis:"order"`
	CreatedAt   time.Time `json:"created_at" redis:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" redis:"updated_at"`
}

// TaskAttachment represents a file attachment to a task
type TaskAttachment struct {
	ID          uuid.UUID `json:"id" redis:"id"`
	TaskID      uuid.UUID `json:"task_id" redis:"task_id"`
	FileName    string    `json:"file_name" redis:"file_name"`
	FileSize    int64     `json:"file_size" redis:"file_size"`
	MimeType    string    `json:"mime_type" redis:"mime_type"`
	URL         string    `json:"url" redis:"url"`
	Description string    `json:"description" redis:"description"`
	UploadedAt  time.Time `json:"uploaded_at" redis:"uploaded_at"`
	UploadedBy  uuid.UUID `json:"uploaded_by" redis:"uploaded_by"`
}

// TaskComment represents a comment on a task
type TaskComment struct {
	ID        uuid.UUID  `json:"id" redis:"id"`
	TaskID    uuid.UUID  `json:"task_id" redis:"task_id"`
	UserID    uuid.UUID  `json:"user_id" redis:"user_id"`
	User      *User      `json:"user,omitempty"`
	Content   string     `json:"content" redis:"content"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty" redis:"parent_id"`
	IsEdited  bool       `json:"is_edited" redis:"is_edited"`
	EditedAt  *time.Time `json:"edited_at,omitempty" redis:"edited_at"`
	CreatedAt time.Time  `json:"created_at" redis:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" redis:"updated_at"`
}

// TimeEntry represents a time tracking entry for a task
type TimeEntry struct {
	ID          uuid.UUID  `json:"id" redis:"id"`
	TaskID      uuid.UUID  `json:"task_id" redis:"task_id"`
	UserID      uuid.UUID  `json:"user_id" redis:"user_id"`
	User        *User      `json:"user,omitempty"`
	Description string     `json:"description" redis:"description"`
	StartTime   time.Time  `json:"start_time" redis:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty" redis:"end_time"`
	Duration    float64    `json:"duration" redis:"duration"` // Hours
	IsBillable  bool       `json:"is_billable" redis:"is_billable"`
	HourlyRate  float64    `json:"hourly_rate" redis:"hourly_rate"`
	CreatedAt   time.Time  `json:"created_at" redis:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" redis:"updated_at"`
}

// RecurrenceRule defines how a task recurs
type RecurrenceRule struct {
	Frequency   RecurrenceFrequency `json:"frequency" redis:"frequency"`
	Interval    int                 `json:"interval" redis:"interval"`
	DaysOfWeek  []time.Weekday      `json:"days_of_week" redis:"days_of_week"`
	DayOfMonth  int                 `json:"day_of_month" redis:"day_of_month"`
	MonthOfYear int                 `json:"month_of_year" redis:"month_of_year"`
	EndDate     *time.Time          `json:"end_date,omitempty" redis:"end_date"`
	Count       int                 `json:"count" redis:"count"`
}

// RecurrenceFrequency defines how often a task recurs
type RecurrenceFrequency string

const (
	FrequencyDaily   RecurrenceFrequency = "daily"
	FrequencyWeekly  RecurrenceFrequency = "weekly"
	FrequencyMonthly RecurrenceFrequency = "monthly"
	FrequencyYearly  RecurrenceFrequency = "yearly"
)

// NewTask creates a new task with default values
func NewTask(title, description string, taskType TaskType, createdBy uuid.UUID) *Task {
	now := time.Now()
	return &Task{
		ID:              uuid.New(),
		Title:           title,
		Description:     description,
		Type:            taskType,
		Status:          StatusTodo,
		Priority:        PriorityMedium,
		Urgency:         UrgencyMedium,
		Complexity:      ComplexityMedium,
		Tags:            []string{},
		Labels:          []string{},
		EstimatedHours:  0,
		ActualHours:     0,
		RemainingHours:  0,
		ProgressPercent: 0,
		Equipment:       []string{},
		Skills:          []string{},
		CustomFields:    make(map[string]interface{}),
		Metadata:        make(map[string]interface{}),
		ExternalIDs:     make(map[string]string),
		IsRecurring:     false,
		IsTemplate:      false,
		IsArchived:      false,
		CreatedAt:       now,
		UpdatedAt:       now,
		CreatedBy:       createdBy,
		UpdatedBy:       createdBy,
		Version:         1,
	}
}

// UpdateStatus updates the task status and handles related logic
func (t *Task) UpdateStatus(newStatus TaskStatus, updatedBy uuid.UUID) {
	_ = t.Status // Store old status for potential future use
	t.Status = newStatus
	t.UpdatedBy = updatedBy
	t.UpdatedAt = time.Now()
	t.Version++

	// Handle status-specific logic
	switch newStatus {
	case StatusInProgress:
		if t.StartDate == nil {
			now := time.Now()
			t.StartDate = &now
		}
	case StatusCompleted:
		now := time.Now()
		t.CompletedDate = &now
		t.ProgressPercent = 100
		t.RemainingHours = 0
	case StatusCancelled:
		t.ProgressPercent = 0
		t.RemainingHours = 0
	}
}

// UpdateProgress updates the task progress percentage
func (t *Task) UpdateProgress(progressPercent float64, updatedBy uuid.UUID) {
	if progressPercent < 0 {
		progressPercent = 0
	}
	if progressPercent > 100 {
		progressPercent = 100
	}

	t.ProgressPercent = progressPercent
	t.UpdatedBy = updatedBy
	t.UpdatedAt = time.Now()
	t.Version++

	// Auto-update status based on progress
	if progressPercent == 100 && t.Status != StatusCompleted {
		t.UpdateStatus(StatusCompleted, updatedBy)
	} else if progressPercent > 0 && t.Status == StatusTodo {
		t.UpdateStatus(StatusInProgress, updatedBy)
	}

	// Update remaining hours based on progress
	if t.EstimatedHours > 0 {
		t.RemainingHours = t.EstimatedHours * (1 - progressPercent/100)
	}
}

// AddTimeEntry adds a time entry to the task
func (t *Task) AddTimeEntry(entry *TimeEntry) {
	t.TimeEntries = append(t.TimeEntries, entry)
	t.ActualHours += entry.Duration
	t.UpdatedAt = time.Now()
	t.Version++

	// Update progress if estimated hours are set
	if t.EstimatedHours > 0 {
		progressPercent := (t.ActualHours / t.EstimatedHours) * 100
		if progressPercent > 100 {
			progressPercent = 100
		}
		t.ProgressPercent = progressPercent
		t.RemainingHours = t.EstimatedHours - t.ActualHours
		if t.RemainingHours < 0 {
			t.RemainingHours = 0
		}
	}
}

// AddAssignment adds an assignment to the task
func (t *Task) AddAssignment(assignment *TaskAssignment) {
	t.Assignments = append(t.Assignments, assignment)
	t.UpdatedAt = time.Now()
	t.Version++
}

// AddDependency adds a dependency to the task
func (t *Task) AddDependency(dependency *TaskDependency) {
	t.Dependencies = append(t.Dependencies, dependency)
	t.UpdatedAt = time.Now()
	t.Version++
}

// AddComment adds a comment to the task
func (t *Task) AddComment(comment *TaskComment) {
	t.Comments = append(t.Comments, comment)
	t.UpdatedAt = time.Now()
	t.Version++
}

// AddAttachment adds an attachment to the task
func (t *Task) AddAttachment(attachment *TaskAttachment) {
	t.Attachments = append(t.Attachments, attachment)
	t.UpdatedAt = time.Now()
	t.Version++
}

// IsOverdue checks if the task is overdue
func (t *Task) IsOverdue() bool {
	if t.DueDate == nil || t.Status == StatusCompleted || t.Status == StatusCancelled {
		return false
	}
	return time.Now().After(*t.DueDate)
}

// IsBlocked checks if the task is blocked by dependencies
func (t *Task) IsBlocked() bool {
	for _, dep := range t.Dependencies {
		if dep.IsActive && dep.DependsOnTask != nil {
			if dep.DependsOnTask.Status != StatusCompleted {
				return true
			}
		}
	}
	return false
}

// CanStart checks if the task can be started (all dependencies are met)
func (t *Task) CanStart() bool {
	return !t.IsBlocked() && t.Status == StatusTodo
}

// GetAssignees returns all users assigned to the task
func (t *Task) GetAssignees() []*User {
	var assignees []*User
	for _, assignment := range t.Assignments {
		if assignment.Role == RoleAssignee && assignment.IsActive && assignment.User != nil {
			assignees = append(assignees, assignment.User)
		}
	}
	return assignees
}

// GetReviewers returns all users assigned as reviewers
func (t *Task) GetReviewers() []*User {
	var reviewers []*User
	for _, assignment := range t.Assignments {
		if assignment.Role == RoleReviewer && assignment.IsActive && assignment.User != nil {
			reviewers = append(reviewers, assignment.User)
		}
	}
	return reviewers
}

// CalculateCompletionRate calculates the completion rate of checklist items
func (t *Task) CalculateCompletionRate() float64 {
	if len(t.Checklist) == 0 {
		return 0
	}

	completed := 0
	for _, item := range t.Checklist {
		if item.IsCompleted {
			completed++
		}
	}

	return float64(completed) / float64(len(t.Checklist)) * 100
}

// Archive archives the task
func (t *Task) Archive(archivedBy uuid.UUID) {
	t.IsArchived = true
	t.UpdatedBy = archivedBy
	t.UpdatedAt = time.Now()
	t.Version++
}

// Unarchive unarchives the task
func (t *Task) Unarchive(unarchivedBy uuid.UUID) {
	t.IsArchived = false
	t.UpdatedBy = unarchivedBy
	t.UpdatedAt = time.Now()
	t.Version++
}

// Domain errors
var (
	ErrTaskNotFound       = NewDomainError("TASK_NOT_FOUND", "Task not found")
	ErrInvalidStatus      = NewDomainError("INVALID_STATUS", "Invalid task status")
	ErrTaskBlocked        = NewDomainError("TASK_BLOCKED", "Task is blocked by dependencies")
	ErrInvalidProgress    = NewDomainError("INVALID_PROGRESS", "Progress must be between 0 and 100")
	ErrCircularDependency = NewDomainError("CIRCULAR_DEPENDENCY", "Circular dependency detected")
	ErrTaskArchived       = NewDomainError("TASK_ARCHIVED", "Cannot modify archived task")
)

// DomainError represents a domain-specific error
type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *DomainError) Error() string {
	return e.Message
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}
