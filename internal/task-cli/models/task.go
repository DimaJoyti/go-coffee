package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	StatusPending     TaskStatus = "pending"
	StatusInProgress  TaskStatus = "in-progress"
	StatusCompleted   TaskStatus = "completed"
	StatusCancelled   TaskStatus = "cancelled"
	StatusOnHold      TaskStatus = "on-hold"
)

// TaskPriority represents the priority level of a task
type TaskPriority string

const (
	PriorityLow      TaskPriority = "low"
	PriorityMedium   TaskPriority = "medium"
	PriorityHigh     TaskPriority = "high"
	PriorityCritical TaskPriority = "critical"
)

// Task represents a task in the system
type Task struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	Assignee    string       `json:"assignee"`
	Creator     string       `json:"creator"`
	Tags        []string     `json:"tags"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
}

// TaskFilter represents filters for querying tasks
type TaskFilter struct {
	Status    []TaskStatus   `json:"status,omitempty"`
	Priority  []TaskPriority `json:"priority,omitempty"`
	Assignee  []string       `json:"assignee,omitempty"`
	Creator   []string       `json:"creator,omitempty"`
	Tags      []string       `json:"tags,omitempty"`
	DueBefore *time.Time     `json:"due_before,omitempty"`
	DueAfter  *time.Time     `json:"due_after,omitempty"`
	Search    string         `json:"search,omitempty"`
}

// TaskCreateRequest represents a request to create a new task
type TaskCreateRequest struct {
	Title       string       `json:"title" validate:"required,min=1,max=200"`
	Description string       `json:"description" validate:"max=1000"`
	Priority    TaskPriority `json:"priority" validate:"required"`
	Assignee    string       `json:"assignee" validate:"max=100"`
	Tags        []string     `json:"tags"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
}

// TaskUpdateRequest represents a request to update a task
type TaskUpdateRequest struct {
	Title       *string       `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description *string       `json:"description,omitempty" validate:"omitempty,max=1000"`
	Status      *TaskStatus   `json:"status,omitempty"`
	Priority    *TaskPriority `json:"priority,omitempty"`
	Assignee    *string       `json:"assignee,omitempty" validate:"omitempty,max=100"`
	Tags        []string      `json:"tags,omitempty"`
	DueDate     *time.Time    `json:"due_date,omitempty"`
}

// TaskStats represents task statistics
type TaskStats struct {
	Total       int                    `json:"total"`
	ByStatus    map[TaskStatus]int     `json:"by_status"`
	ByPriority  map[TaskPriority]int   `json:"by_priority"`
	ByAssignee  map[string]int         `json:"by_assignee"`
	Overdue     int                    `json:"overdue"`
	DueToday    int                    `json:"due_today"`
	DueThisWeek int                    `json:"due_this_week"`
}

// NewTask creates a new task with default values
func NewTask(req TaskCreateRequest, creator string) *Task {
	now := time.Now()
	task := &Task{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Status:      StatusPending,
		Priority:    req.Priority,
		Assignee:    req.Assignee,
		Creator:     creator,
		Tags:        req.Tags,
		CreatedAt:   now,
		UpdatedAt:   now,
		DueDate:     req.DueDate,
	}

	if task.Tags == nil {
		task.Tags = []string{}
	}

	return task
}

// Update updates the task with the provided request
func (t *Task) Update(req TaskUpdateRequest) {
	t.UpdatedAt = time.Now()

	if req.Title != nil {
		t.Title = *req.Title
	}
	if req.Description != nil {
		t.Description = *req.Description
	}
	if req.Status != nil {
		t.Status = *req.Status
		if *req.Status == StatusCompleted {
			now := time.Now()
			t.CompletedAt = &now
		} else {
			t.CompletedAt = nil
		}
	}
	if req.Priority != nil {
		t.Priority = *req.Priority
	}
	if req.Assignee != nil {
		t.Assignee = *req.Assignee
	}
	if req.Tags != nil {
		t.Tags = req.Tags
	}
	if req.DueDate != nil {
		t.DueDate = req.DueDate
	}
}

// IsOverdue checks if the task is overdue
func (t *Task) IsOverdue() bool {
	if t.DueDate == nil || t.Status == StatusCompleted || t.Status == StatusCancelled {
		return false
	}
	return time.Now().After(*t.DueDate)
}

// IsDueToday checks if the task is due today
func (t *Task) IsDueToday() bool {
	if t.DueDate == nil {
		return false
	}
	now := time.Now()
	due := *t.DueDate
	return now.Year() == due.Year() && now.YearDay() == due.YearDay()
}

// IsDueThisWeek checks if the task is due this week
func (t *Task) IsDueThisWeek() bool {
	if t.DueDate == nil {
		return false
	}
	now := time.Now()
	due := *t.DueDate
	
	// Get the start of this week (Monday)
	weekStart := now.AddDate(0, 0, -int(now.Weekday())+1)
	weekEnd := weekStart.AddDate(0, 0, 7)
	
	return due.After(weekStart) && due.Before(weekEnd)
}

// ToJSON converts the task to JSON
func (t *Task) ToJSON() ([]byte, error) {
	return json.Marshal(t)
}

// FromJSON creates a task from JSON
func FromJSON(data []byte) (*Task, error) {
	var task Task
	err := json.Unmarshal(data, &task)
	return &task, err
}

// ValidateStatus checks if the status is valid
func ValidateStatus(status string) bool {
	switch TaskStatus(status) {
	case StatusPending, StatusInProgress, StatusCompleted, StatusCancelled, StatusOnHold:
		return true
	default:
		return false
	}
}

// ValidatePriority checks if the priority is valid
func ValidatePriority(priority string) bool {
	switch TaskPriority(priority) {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical:
		return true
	default:
		return false
	}
}

// GetAllStatuses returns all valid task statuses
func GetAllStatuses() []TaskStatus {
	return []TaskStatus{StatusPending, StatusInProgress, StatusCompleted, StatusCancelled, StatusOnHold}
}

// GetAllPriorities returns all valid task priorities
func GetAllPriorities() []TaskPriority {
	return []TaskPriority{PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical}
}
