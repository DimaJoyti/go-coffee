package interfaces

import (
	"context"
	"time"
)

// TaskManager defines the interface for task management systems
type TaskManager interface {
	// Task operations
	CreateTask(ctx context.Context, req *CreateTaskRequest) (*Task, error)
	GetTask(ctx context.Context, taskID string) (*Task, error)
	UpdateTask(ctx context.Context, taskID string, req *UpdateTaskRequest) (*Task, error)
	DeleteTask(ctx context.Context, taskID string) error
	ListTasks(ctx context.Context, req *ListTasksRequest) (*TaskList, error)
	
	// Project operations
	CreateProject(ctx context.Context, req *CreateProjectRequest) (*Project, error)
	GetProject(ctx context.Context, projectID string) (*Project, error)
	UpdateProject(ctx context.Context, projectID string, req *UpdateProjectRequest) (*Project, error)
	DeleteProject(ctx context.Context, projectID string) error
	ListProjects(ctx context.Context, req *ListProjectsRequest) (*ProjectList, error)
	
	// Assignment operations
	AssignTask(ctx context.Context, taskID, userID string) error
	UnassignTask(ctx context.Context, taskID, userID string) error
	
	// Time tracking
	StartTimeTracking(ctx context.Context, taskID, userID string) (*TimeEntry, error)
	StopTimeTracking(ctx context.Context, entryID string) (*TimeEntry, error)
	GetTimeEntries(ctx context.Context, req *TimeEntriesRequest) (*TimeEntryList, error)
	
	// Comments and attachments
	AddComment(ctx context.Context, taskID string, req *TaskCommentRequest) (*Comment, error)
	GetComments(ctx context.Context, taskID string) ([]*Comment, error)
	AddAttachment(ctx context.Context, taskID string, req *AttachmentRequest) (*Attachment, error)
	
	// Webhooks
	RegisterWebhook(ctx context.Context, req *TaskWebhookRequest) (*TaskWebhook, error)
	UnregisterWebhook(ctx context.Context, webhookID string) error
	
	// Bulk operations
	BulkCreateTasks(ctx context.Context, tasks []*CreateTaskRequest) ([]*Task, error)
	BulkUpdateTasks(ctx context.Context, updates []*BulkTaskUpdate) ([]*Task, error)
	
	// Search and filtering
	SearchTasks(ctx context.Context, req *TaskSearchRequest) (*TaskList, error)
	
	// Provider info
	GetProviderInfo() *ProviderInfo
}

// Task represents a task in the task management system
type Task struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      TaskStatus             `json:"status"`
	Priority    TaskPriority           `json:"priority"`
	
	// Relationships
	ProjectID   string                 `json:"project_id,omitempty"`
	ParentID    string                 `json:"parent_id,omitempty"`
	AssigneeIDs []string               `json:"assignee_ids,omitempty"`
	
	// Dates
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DueDate     *time.Time             `json:"due_date,omitempty"`
	StartDate   *time.Time             `json:"start_date,omitempty"`
	
	// Metadata
	Tags        []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	
	// Progress tracking
	Progress    int                    `json:"progress"` // 0-100
	TimeEstimate *time.Duration        `json:"time_estimate,omitempty"`
	TimeSpent   *time.Duration         `json:"time_spent,omitempty"`
	
	// External references
	ExternalID  string                 `json:"external_id,omitempty"`
	URL         string                 `json:"url,omitempty"`
}

// Project represents a project in the task management system
type Project struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      ProjectStatus          `json:"status"`
	
	// Relationships
	TeamID      string                 `json:"team_id,omitempty"`
	OwnerID     string                 `json:"owner_id,omitempty"`
	MemberIDs   []string               `json:"member_ids,omitempty"`
	
	// Dates
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	StartDate   *time.Time             `json:"start_date,omitempty"`
	EndDate     *time.Time             `json:"end_date,omitempty"`
	
	// Metadata
	Tags        []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	
	// Progress tracking
	Progress    int                    `json:"progress"` // 0-100
	TaskCount   int                    `json:"task_count"`
	
	// External references
	ExternalID  string                 `json:"external_id,omitempty"`
	URL         string                 `json:"url,omitempty"`
}

// TimeEntry represents a time tracking entry
type TimeEntry struct {
	ID          string                 `json:"id"`
	TaskID      string                 `json:"task_id"`
	UserID      string                 `json:"user_id"`
	Description string                 `json:"description,omitempty"`
	
	// Time tracking
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Duration    time.Duration          `json:"duration"`
	
	// Metadata
	Billable    bool                   `json:"billable"`
	Tags        []string               `json:"tags,omitempty"`
	
	// Dates
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Comment represents a comment on a task
type Comment struct {
	ID          string                 `json:"id"`
	TaskID      string                 `json:"task_id"`
	UserID      string                 `json:"user_id"`
	Content     string                 `json:"content"`
	
	// Metadata
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	
	// Attachments
	Attachments []*Attachment          `json:"attachments,omitempty"`
}

// Attachment represents a file attachment
type Attachment struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	URL         string                 `json:"url"`
	Size        int64                  `json:"size"`
	MimeType    string                 `json:"mime_type"`
	
	// Metadata
	CreatedAt   time.Time              `json:"created_at"`
	UploadedBy  string                 `json:"uploaded_by"`
}

// Webhook represents a webhook configuration
type TaskWebhook struct {
	ID          string                 `json:"id"`
	URL         string                 `json:"url"`
	Events      []string               `json:"events"`
	Secret      string                 `json:"secret,omitempty"`
	Active      bool                   `json:"active"`
	
	// Metadata
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Request types
type CreateTaskRequest struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	ProjectID    string                 `json:"project_id,omitempty"`
	ParentID     string                 `json:"parent_id,omitempty"`
	AssigneeIDs  []string               `json:"assignee_ids,omitempty"`
	Priority     TaskPriority           `json:"priority,omitempty"`
	DueDate      *time.Time             `json:"due_date,omitempty"`
	StartDate    *time.Time             `json:"start_date,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	TimeEstimate *time.Duration         `json:"time_estimate,omitempty"`
}

type UpdateTaskRequest struct {
	Name         *string                `json:"name,omitempty"`
	Description  *string                `json:"description,omitempty"`
	Status       *TaskStatus            `json:"status,omitempty"`
	Priority     *TaskPriority          `json:"priority,omitempty"`
	DueDate      *time.Time             `json:"due_date,omitempty"`
	StartDate    *time.Time             `json:"start_date,omitempty"`
	Progress     *int                   `json:"progress,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

type ListTasksRequest struct {
	ProjectID    string                 `json:"project_id,omitempty"`
	AssigneeID   string                 `json:"assignee_id,omitempty"`
	Status       []TaskStatus           `json:"status,omitempty"`
	Priority     []TaskPriority         `json:"priority,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	DueBefore    *time.Time             `json:"due_before,omitempty"`
	DueAfter     *time.Time             `json:"due_after,omitempty"`
	Limit        int                    `json:"limit,omitempty"`
	Offset       int                    `json:"offset,omitempty"`
	OrderBy      string                 `json:"order_by,omitempty"`
	OrderDir     string                 `json:"order_dir,omitempty"`
}

type CreateProjectRequest struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	TeamID       string                 `json:"team_id,omitempty"`
	OwnerID      string                 `json:"owner_id,omitempty"`
	MemberIDs    []string               `json:"member_ids,omitempty"`
	StartDate    *time.Time             `json:"start_date,omitempty"`
	EndDate      *time.Time             `json:"end_date,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

type UpdateProjectRequest struct {
	Name         *string                `json:"name,omitempty"`
	Description  *string                `json:"description,omitempty"`
	Status       *ProjectStatus         `json:"status,omitempty"`
	StartDate    *time.Time             `json:"start_date,omitempty"`
	EndDate      *time.Time             `json:"end_date,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

type ListProjectsRequest struct {
	TeamID       string                 `json:"team_id,omitempty"`
	OwnerID      string                 `json:"owner_id,omitempty"`
	Status       []ProjectStatus        `json:"status,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Limit        int                    `json:"limit,omitempty"`
	Offset       int                    `json:"offset,omitempty"`
	OrderBy      string                 `json:"order_by,omitempty"`
	OrderDir     string                 `json:"order_dir,omitempty"`
}

type TimeEntriesRequest struct {
	TaskID       string                 `json:"task_id,omitempty"`
	UserID       string                 `json:"user_id,omitempty"`
	ProjectID    string                 `json:"project_id,omitempty"`
	StartDate    *time.Time             `json:"start_date,omitempty"`
	EndDate      *time.Time             `json:"end_date,omitempty"`
	Billable     *bool                  `json:"billable,omitempty"`
	Limit        int                    `json:"limit,omitempty"`
	Offset       int                    `json:"offset,omitempty"`
}

type TaskCommentRequest struct {
	Content      string                 `json:"content"`
	UserID       string                 `json:"user_id"`
	Attachments  []*AttachmentRequest   `json:"attachments,omitempty"`
}

type AttachmentRequest struct {
	Name         string                 `json:"name"`
	Content      []byte                 `json:"content"`
	MimeType     string                 `json:"mime_type"`
}

type TaskWebhookRequest struct {
	URL          string                 `json:"url"`
	Events       []string               `json:"events"`
	Secret       string                 `json:"secret,omitempty"`
}

type BulkTaskUpdate struct {
	TaskID       string                 `json:"task_id"`
	Updates      *UpdateTaskRequest     `json:"updates"`
}

type TaskSearchRequest struct {
	Query        string                 `json:"query"`
	ProjectID    string                 `json:"project_id,omitempty"`
	Filters      map[string]interface{} `json:"filters,omitempty"`
	Limit        int                    `json:"limit,omitempty"`
	Offset       int                    `json:"offset,omitempty"`
}

// Response types
type TaskList struct {
	Tasks        []*Task                `json:"tasks"`
	Total        int                    `json:"total"`
	Limit        int                    `json:"limit"`
	Offset       int                    `json:"offset"`
	HasMore      bool                   `json:"has_more"`
}

type ProjectList struct {
	Projects     []*Project             `json:"projects"`
	Total        int                    `json:"total"`
	Limit        int                    `json:"limit"`
	Offset       int                    `json:"offset"`
	HasMore      bool                   `json:"has_more"`
}

type TimeEntryList struct {
	Entries      []*TimeEntry           `json:"entries"`
	Total        int                    `json:"total"`
	TotalDuration time.Duration         `json:"total_duration"`
	Limit        int                    `json:"limit"`
	Offset       int                    `json:"offset"`
	HasMore      bool                   `json:"has_more"`
}

// Enums
type TaskStatus string

const (
	TaskStatusOpen       TaskStatus = "open"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusReview     TaskStatus = "review"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type TaskPriority string

const (
	TaskPriorityLow      TaskPriority = "low"
	TaskPriorityNormal   TaskPriority = "normal"
	TaskPriorityHigh     TaskPriority = "high"
	TaskPriorityUrgent   TaskPriority = "urgent"
)

type ProjectStatus string

const (
	ProjectStatusActive    ProjectStatus = "active"
	ProjectStatusArchived  ProjectStatus = "archived"
	ProjectStatusCompleted ProjectStatus = "completed"
	ProjectStatusOnHold    ProjectStatus = "on_hold"
)

// ProviderInfo contains information about the task management provider
type ProviderInfo struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Capabilities []string               `json:"capabilities"`
	RateLimits   map[string]int         `json:"rate_limits"`
}
