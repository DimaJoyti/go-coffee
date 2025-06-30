package repositories

import (
	"context"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/entities"
	"github.com/google/uuid"
)

// BeverageRepository defines the interface for beverage data operations
type BeverageRepository interface {
	// Save stores a beverage in the repository
	Save(ctx context.Context, beverage *entities.Beverage) error
	
	// FindByID retrieves a beverage by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Beverage, error)
	
	// FindByTheme retrieves beverages by theme
	FindByTheme(ctx context.Context, theme string) ([]*entities.Beverage, error)
	
	// FindByStatus retrieves beverages by status
	FindByStatus(ctx context.Context, status entities.BeverageStatus) ([]*entities.Beverage, error)
	
	// FindByIngredient retrieves beverages containing a specific ingredient
	FindByIngredient(ctx context.Context, ingredient string) ([]*entities.Beverage, error)
	
	// Update updates an existing beverage
	Update(ctx context.Context, beverage *entities.Beverage) error
	
	// Delete removes a beverage from the repository
	Delete(ctx context.Context, id uuid.UUID) error
	
	// List retrieves all beverages with pagination
	List(ctx context.Context, offset, limit int) ([]*entities.Beverage, error)
	
	// Count returns the total number of beverages
	Count(ctx context.Context) (int64, error)
}

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	// PublishBeverageCreated publishes a beverage created event
	PublishBeverageCreated(ctx context.Context, beverage *entities.Beverage) error
	
	// PublishBeverageUpdated publishes a beverage updated event
	PublishBeverageUpdated(ctx context.Context, beverage *entities.Beverage) error
	
	// PublishBeverageStatusChanged publishes a beverage status change event
	PublishBeverageStatusChanged(ctx context.Context, beverage *entities.Beverage, oldStatus entities.BeverageStatus) error
}

// AIProvider defines the interface for AI/LLM providers
type AIProvider interface {
	// GenerateRecipe generates a recipe using AI
	GenerateRecipe(ctx context.Context, prompt string) (string, error)
	
	// AnalyzeIngredients analyzes ingredient compatibility
	AnalyzeIngredients(ctx context.Context, ingredients []string) (*IngredientAnalysis, error)
	
	// GenerateDescription generates a creative description
	GenerateDescription(ctx context.Context, beverage *entities.Beverage) (string, error)
	
	// SuggestImprovements suggests improvements for a recipe
	SuggestImprovements(ctx context.Context, beverage *entities.Beverage) ([]string, error)
}

// IngredientAnalysis represents the result of ingredient analysis
type IngredientAnalysis struct {
	Compatible    bool     `json:"compatible"`
	Confidence    float64  `json:"confidence"`
	Suggestions   []string `json:"suggestions"`
	Warnings      []string `json:"warnings"`
	FlavorProfile string   `json:"flavor_profile"`
}

// TaskManager defines the interface for task management operations
type TaskManager interface {
	// CreateTask creates a new task
	CreateTask(ctx context.Context, task *Task) error
	
	// UpdateTask updates an existing task
	UpdateTask(ctx context.Context, task *Task) error
	
	// GetTask retrieves a task by ID
	GetTask(ctx context.Context, taskID string) (*Task, error)
	
	// ListTasks lists tasks with filters
	ListTasks(ctx context.Context, filters TaskFilters) ([]*Task, error)
}

// Task represents a task in the task management system
type Task struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      TaskStatus        `json:"status"`
	Priority    TaskPriority      `json:"priority"`
	Assignee    string            `json:"assignee"`
	DueDate     *string           `json:"due_date,omitempty"`
	Tags        []string          `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusOpen       TaskStatus = "open"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusReview     TaskStatus = "review"
	TaskStatusClosed     TaskStatus = "closed"
)

// TaskPriority represents the priority of a task
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityNormal TaskPriority = "normal"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

// TaskFilters represents filters for listing tasks
type TaskFilters struct {
	Status   *TaskStatus   `json:"status,omitempty"`
	Priority *TaskPriority `json:"priority,omitempty"`
	Assignee *string       `json:"assignee,omitempty"`
	Tags     []string      `json:"tags,omitempty"`
	Limit    int           `json:"limit"`
	Offset   int           `json:"offset"`
}

// NotificationService defines the interface for sending notifications
type NotificationService interface {
	// SendSlackMessage sends a message to Slack
	SendSlackMessage(ctx context.Context, channel, message string) error
	
	// SendEmail sends an email notification
	SendEmail(ctx context.Context, to, subject, body string) error
	
	// SendWebhook sends a webhook notification
	SendWebhook(ctx context.Context, url string, payload interface{}) error
}
