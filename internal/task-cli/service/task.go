package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/task-cli/config"
	"github.com/DimaJoyti/go-coffee/internal/task-cli/models"
	"github.com/DimaJoyti/go-coffee/internal/task-cli/repository"
)

// TaskService defines the interface for task business logic
type TaskService interface {
	CreateTask(ctx context.Context, req models.TaskCreateRequest, creator string) (*models.Task, error)
	GetTask(ctx context.Context, id string) (*models.Task, error)
	UpdateTask(ctx context.Context, id string, req models.TaskUpdateRequest) (*models.Task, error)
	DeleteTask(ctx context.Context, id string) error
	ListTasks(ctx context.Context, filter models.TaskFilter, sortBy, sortOrder string, offset, limit int) ([]*models.Task, int, error)
	SearchTasks(ctx context.Context, query string, offset, limit int) ([]*models.Task, int, error)
	GetTaskStats(ctx context.Context) (*models.TaskStats, error)
	AssignTask(ctx context.Context, id, assignee string) (*models.Task, error)
	ChangeTaskStatus(ctx context.Context, id string, status models.TaskStatus) (*models.Task, error)
	GetTasksByAssignee(ctx context.Context, assignee string) ([]*models.Task, error)
	GetTasksByStatus(ctx context.Context, status models.TaskStatus) ([]*models.Task, error)
	GetOverdueTasks(ctx context.Context) ([]*models.Task, error)
	GetTasksDueToday(ctx context.Context) ([]*models.Task, error)
	GetTasksDueThisWeek(ctx context.Context) ([]*models.Task, error)
	BulkUpdateTasks(ctx context.Context, filter models.TaskFilter, updates models.TaskUpdateRequest) (int, error)
	ValidateTaskRequest(req models.TaskCreateRequest) error
	ValidateTaskUpdate(req models.TaskUpdateRequest) error
}

// taskService implements TaskService
type taskService struct {
	repo   repository.TaskRepository
	config *config.Config
}

// NewTaskService creates a new task service
func NewTaskService(repo repository.TaskRepository, cfg *config.Config) TaskService {
	return &taskService{
		repo:   repo,
		config: cfg,
	}
}

// CreateTask creates a new task
func (s *taskService) CreateTask(ctx context.Context, req models.TaskCreateRequest, creator string) (*models.Task, error) {
	if err := s.ValidateTaskRequest(req); err != nil {
		return nil, fmt.Errorf("invalid task request: %w", err)
	}

	// Use default creator if not provided
	if creator == "" {
		creator = s.config.CLI.DefaultUser
	}

	// Use default assignee if not provided
	if req.Assignee == "" {
		req.Assignee = creator
	}

	task := models.NewTask(req, creator)

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

// GetTask retrieves a task by ID
func (s *taskService) GetTask(ctx context.Context, id string) (*models.Task, error) {
	if id == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

// UpdateTask updates an existing task
func (s *taskService) UpdateTask(ctx context.Context, id string, req models.TaskUpdateRequest) (*models.Task, error) {
	if id == "" {
		return nil, fmt.Errorf("task ID is required")
	}

	if err := s.ValidateTaskUpdate(req); err != nil {
		return nil, fmt.Errorf("invalid update request: %w", err)
	}

	task, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	task.Update(req)

	if err := s.repo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return task, nil
}

// DeleteTask deletes a task
func (s *taskService) DeleteTask(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("task ID is required")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

// ListTasks retrieves tasks with filtering, sorting, and pagination
func (s *taskService) ListTasks(ctx context.Context, filter models.TaskFilter, sortBy, sortOrder string, offset, limit int) ([]*models.Task, int, error) {
	tasks, total, err := s.repo.List(ctx, filter, 0, 10000) // Get all matching tasks for sorting
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Sort tasks
	s.sortTasks(tasks, sortBy, sortOrder)

	// Apply pagination
	if offset >= len(tasks) {
		return []*models.Task{}, total, nil
	}

	end := offset + limit
	if end > len(tasks) {
		end = len(tasks)
	}

	return tasks[offset:end], total, nil
}

// SearchTasks searches tasks by query
func (s *taskService) SearchTasks(ctx context.Context, query string, offset, limit int) ([]*models.Task, int, error) {
	if query == "" {
		return nil, 0, fmt.Errorf("search query is required")
	}

	tasks, total, err := s.repo.Search(ctx, query, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search tasks: %w", err)
	}

	return tasks, total, nil
}

// GetTaskStats returns task statistics
func (s *taskService) GetTaskStats(ctx context.Context) (*models.TaskStats, error) {
	stats, err := s.repo.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task stats: %w", err)
	}

	return stats, nil
}

// AssignTask assigns a task to a user
func (s *taskService) AssignTask(ctx context.Context, id, assignee string) (*models.Task, error) {
	if id == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	if assignee == "" {
		return nil, fmt.Errorf("assignee is required")
	}

	req := models.TaskUpdateRequest{
		Assignee: &assignee,
	}

	return s.UpdateTask(ctx, id, req)
}

// ChangeTaskStatus changes the status of a task
func (s *taskService) ChangeTaskStatus(ctx context.Context, id string, status models.TaskStatus) (*models.Task, error) {
	if id == "" {
		return nil, fmt.Errorf("task ID is required")
	}
	if !models.ValidateStatus(string(status)) {
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	req := models.TaskUpdateRequest{
		Status: &status,
	}

	return s.UpdateTask(ctx, id, req)
}

// GetTasksByAssignee retrieves tasks by assignee
func (s *taskService) GetTasksByAssignee(ctx context.Context, assignee string) ([]*models.Task, error) {
	if assignee == "" {
		return nil, fmt.Errorf("assignee is required")
	}

	tasks, err := s.repo.GetByAssignee(ctx, assignee)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by assignee: %w", err)
	}

	return tasks, nil
}

// GetTasksByStatus retrieves tasks by status
func (s *taskService) GetTasksByStatus(ctx context.Context, status models.TaskStatus) ([]*models.Task, error) {
	if !models.ValidateStatus(string(status)) {
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	tasks, err := s.repo.GetByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks by status: %w", err)
	}

	return tasks, nil
}

// GetOverdueTasks retrieves overdue tasks
func (s *taskService) GetOverdueTasks(ctx context.Context) ([]*models.Task, error) {
	tasks, err := s.repo.GetOverdue(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue tasks: %w", err)
	}

	return tasks, nil
}

// GetTasksDueToday retrieves tasks due today
func (s *taskService) GetTasksDueToday(ctx context.Context) ([]*models.Task, error) {
	tasks, _, err := s.repo.List(ctx, models.TaskFilter{}, 0, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	var dueToday []*models.Task
	for _, task := range tasks {
		if task.IsDueToday() {
			dueToday = append(dueToday, task)
		}
	}

	return dueToday, nil
}

// GetTasksDueThisWeek retrieves tasks due this week
func (s *taskService) GetTasksDueThisWeek(ctx context.Context) ([]*models.Task, error) {
	tasks, _, err := s.repo.List(ctx, models.TaskFilter{}, 0, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	var dueThisWeek []*models.Task
	for _, task := range tasks {
		if task.IsDueThisWeek() {
			dueThisWeek = append(dueThisWeek, task)
		}
	}

	return dueThisWeek, nil
}

// BulkUpdateTasks updates multiple tasks matching the filter
func (s *taskService) BulkUpdateTasks(ctx context.Context, filter models.TaskFilter, updates models.TaskUpdateRequest) (int, error) {
	if err := s.ValidateTaskUpdate(updates); err != nil {
		return 0, fmt.Errorf("invalid update request: %w", err)
	}

	tasks, _, err := s.repo.List(ctx, filter, 0, 10000)
	if err != nil {
		return 0, fmt.Errorf("failed to get tasks for bulk update: %w", err)
	}

	updated := 0
	for _, task := range tasks {
		task.Update(updates)
		if err := s.repo.Update(ctx, task); err != nil {
			continue // Skip failed updates
		}
		updated++
	}

	return updated, nil
}

// ValidateTaskRequest validates a task creation request
func (s *taskService) ValidateTaskRequest(req models.TaskCreateRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if len(req.Title) > 200 {
		return fmt.Errorf("title must be less than 200 characters")
	}
	if len(req.Description) > 1000 {
		return fmt.Errorf("description must be less than 1000 characters")
	}
	if !models.ValidatePriority(string(req.Priority)) {
		return fmt.Errorf("invalid priority: %s", req.Priority)
	}
	if req.DueDate != nil && req.DueDate.Before(time.Now()) {
		return fmt.Errorf("due date cannot be in the past")
	}

	return nil
}

// ValidateTaskUpdate validates a task update request
func (s *taskService) ValidateTaskUpdate(req models.TaskUpdateRequest) error {
	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" {
			return fmt.Errorf("title cannot be empty")
		}
		if len(*req.Title) > 200 {
			return fmt.Errorf("title must be less than 200 characters")
		}
	}
	if req.Description != nil && len(*req.Description) > 1000 {
		return fmt.Errorf("description must be less than 1000 characters")
	}
	if req.Status != nil && !models.ValidateStatus(string(*req.Status)) {
		return fmt.Errorf("invalid status: %s", *req.Status)
	}
	if req.Priority != nil && !models.ValidatePriority(string(*req.Priority)) {
		return fmt.Errorf("invalid priority: %s", *req.Priority)
	}

	return nil
}

// sortTasks sorts tasks by the specified field and order
func (s *taskService) sortTasks(tasks []*models.Task, sortBy, sortOrder string) {
	if sortBy == "" {
		sortBy = s.config.CLI.SortBy
	}
	if sortOrder == "" {
		sortOrder = s.config.CLI.SortOrder
	}

	sort.Slice(tasks, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "title":
			less = tasks[i].Title < tasks[j].Title
		case "status":
			less = tasks[i].Status < tasks[j].Status
		case "priority":
			priorityOrder := map[models.TaskPriority]int{
				models.PriorityLow:      1,
				models.PriorityMedium:   2,
				models.PriorityHigh:     3,
				models.PriorityCritical: 4,
			}
			less = priorityOrder[tasks[i].Priority] < priorityOrder[tasks[j].Priority]
		case "assignee":
			less = tasks[i].Assignee < tasks[j].Assignee
		case "creator":
			less = tasks[i].Creator < tasks[j].Creator
		case "due_date":
			if tasks[i].DueDate == nil && tasks[j].DueDate == nil {
				less = false
			} else if tasks[i].DueDate == nil {
				less = false
			} else if tasks[j].DueDate == nil {
				less = true
			} else {
				less = tasks[i].DueDate.Before(*tasks[j].DueDate)
			}
		case "updated_at":
			less = tasks[i].UpdatedAt.Before(tasks[j].UpdatedAt)
		default: // created_at
			less = tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
		}

		if sortOrder == "desc" {
			return !less
		}
		return less
	})
}
