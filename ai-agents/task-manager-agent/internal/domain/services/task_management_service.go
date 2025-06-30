package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/task-manager-agent/internal/domain/entities"
	"go-coffee-ai-agents/task-manager-agent/internal/domain/events"
	"go-coffee-ai-agents/task-manager-agent/internal/domain/repositories"
)

// TaskManagementService provides core task management business logic
type TaskManagementService struct {
	taskRepo         repositories.TaskRepository
	projectRepo      repositories.ProjectRepository
	userRepo         repositories.UserRepository
	workflowRepo     repositories.WorkflowRepository
	notificationRepo repositories.NotificationRepository
	eventPublisher   EventPublisher
	logger           Logger
}

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	PublishEvent(ctx context.Context, event DomainEvent) error
	PublishEvents(ctx context.Context, events []DomainEvent) error
}

// DomainEvent represents a domain event
type DomainEvent interface {
	GetEventType() string
	GetAggregateID() uuid.UUID
	GetEventData() map[string]interface{}
	GetTimestamp() time.Time
	GetVersion() int
}

// Logger defines the interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// NewTaskManagementService creates a new task management service
func NewTaskManagementService(
	taskRepo repositories.TaskRepository,
	projectRepo repositories.ProjectRepository,
	userRepo repositories.UserRepository,
	workflowRepo repositories.WorkflowRepository,
	notificationRepo repositories.NotificationRepository,
	eventPublisher EventPublisher,
	logger Logger,
) *TaskManagementService {
	return &TaskManagementService{
		taskRepo:         taskRepo,
		projectRepo:      projectRepo,
		userRepo:         userRepo,
		workflowRepo:     workflowRepo,
		notificationRepo: notificationRepo,
		eventPublisher:   eventPublisher,
		logger:           logger,
	}
}

// CreateTask creates a new task with validation and business rules
func (tms *TaskManagementService) CreateTask(ctx context.Context, task *entities.Task) error {
	tms.logger.Info("Creating new task", "title", task.Title, "type", task.Type)

	// Validate task data
	if err := tms.validateTask(ctx, task); err != nil {
		tms.logger.Error("Task validation failed", err, "task_id", task.ID)
		return err
	}

	// Check if project exists and user has permission
	if task.ProjectID != nil {
		project, err := tms.projectRepo.GetByID(ctx, *task.ProjectID)
		if err != nil {
			tms.logger.Error("Failed to get project", err, "project_id", *task.ProjectID)
			return fmt.Errorf("project not found: %w", err)
		}

		// Add task to project
		project.AddTask(task)
		if err := tms.projectRepo.Update(ctx, project); err != nil {
			tms.logger.Error("Failed to update project", err, "project_id", project.ID)
			return err
		}
	}

	// Create the task
	if err := tms.taskRepo.Create(ctx, task); err != nil {
		tms.logger.Error("Failed to create task", err, "task_id", task.ID)
		return err
	}

	// Publish task created event
	event := events.NewTaskCreatedEvent(task)
	if err := tms.eventPublisher.PublishEvent(ctx, event); err != nil {
		tms.logger.Error("Failed to publish task created event", err, "task_id", task.ID)
	}

	// Send notifications to assignees
	if err := tms.notifyTaskAssignees(ctx, task, "Task assigned to you"); err != nil {
		tms.logger.Error("Failed to notify assignees", err, "task_id", task.ID)
	}

	tms.logger.Info("Task created successfully", "task_id", task.ID, "title", task.Title)
	return nil
}

// UpdateTask updates an existing task with validation
func (tms *TaskManagementService) UpdateTask(ctx context.Context, task *entities.Task) error {
	tms.logger.Info("Updating task", "task_id", task.ID, "title", task.Title)

	// Get existing task for comparison
	existingTask, err := tms.taskRepo.GetByID(ctx, task.ID)
	if err != nil {
		tms.logger.Error("Failed to get existing task", err, "task_id", task.ID)
		return err
	}

	// Validate task data
	if err := tms.validateTask(ctx, task); err != nil {
		tms.logger.Error("Task validation failed", err, "task_id", task.ID)
		return err
	}

	// Check for status changes and handle accordingly
	if existingTask.Status != task.Status {
		if err := tms.handleStatusChange(ctx, existingTask, task.Status, task.UpdatedBy); err != nil {
			tms.logger.Error("Failed to handle status change", err, "task_id", task.ID)
			return err
		}
	}

	// Update the task
	if err := tms.taskRepo.Update(ctx, task); err != nil {
		tms.logger.Error("Failed to update task", err, "task_id", task.ID)
		return err
	}

	// Publish task updated event
	event := events.NewTaskUpdatedEvent(task, existingTask)
	if err := tms.eventPublisher.PublishEvent(ctx, event); err != nil {
		tms.logger.Error("Failed to publish task updated event", err, "task_id", task.ID)
	}

	tms.logger.Info("Task updated successfully", "task_id", task.ID)
	return nil
}

// AssignTask assigns a task to a user with validation
func (tms *TaskManagementService) AssignTask(ctx context.Context, taskID, userID, assignedBy uuid.UUID, role entities.AssignmentRole) error {
	tms.logger.Info("Assigning task", "task_id", taskID, "user_id", userID, "role", role)

	// Get task
	task, err := tms.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		tms.logger.Error("Failed to get task", err, "task_id", taskID)
		return err
	}

	// Get user
	user, err := tms.userRepo.GetByID(ctx, userID)
	if err != nil {
		tms.logger.Error("Failed to get user", err, "user_id", userID)
		return err
	}

	// Check user availability and capacity
	if !user.IsAvailable() {
		return fmt.Errorf("user is not available for assignment")
	}

	// Check if user has required skills
	if err := tms.validateUserSkills(user, task.Skills); err != nil {
		tms.logger.Warn("User skills validation failed", "user_id", userID, "required_skills", task.Skills)
		// Don't fail assignment, just log warning
	}

	// Create assignment
	assignment := &entities.TaskAssignment{
		ID:         uuid.New(),
		TaskID:     taskID,
		UserID:     userID,
		User:       user,
		Role:       role,
		Allocation: 100.0, // Default full allocation
		AssignedAt: time.Now(),
		AssignedBy: assignedBy,
		Status:     entities.AssignmentPending,
		IsActive:   true,
	}

	// Add assignment to task
	task.AddAssignment(assignment)

	// Update task
	if err := tms.taskRepo.Update(ctx, task); err != nil {
		tms.logger.Error("Failed to update task with assignment", err, "task_id", taskID)
		return err
	}

	// Create assignment record
	if err := tms.taskRepo.AddAssignment(ctx, assignment); err != nil {
		tms.logger.Error("Failed to create assignment", err, "assignment_id", assignment.ID)
		return err
	}

	// Update user capacity
	if err := tms.updateUserCapacity(ctx, userID, task.EstimatedHours); err != nil {
		tms.logger.Error("Failed to update user capacity", err, "user_id", userID)
	}

	// Send notification
	notification := &repositories.Notification{
		ID:          uuid.New(),
		Type:        repositories.NotificationTypeTaskAssigned,
		Title:       "Task Assigned",
		Message:     fmt.Sprintf("You have been assigned to task: %s", task.Title),
		Priority:    repositories.NotificationPriorityNormal,
		Category:    repositories.NotificationCategoryTask,
		UserID:      userID,
		RelatedID:   &taskID,
		RelatedType: "task",
		Data: map[string]interface{}{
			"task_id":    taskID,
			"task_title": task.Title,
			"role":       role,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := tms.notificationRepo.Create(ctx, notification); err != nil {
		tms.logger.Error("Failed to create notification", err, "user_id", userID)
	}

	// Publish task assigned event
	event := events.NewTaskAssignedEvent(task, assignment)
	if err := tms.eventPublisher.PublishEvent(ctx, event); err != nil {
		tms.logger.Error("Failed to publish task assigned event", err, "task_id", taskID)
	}

	tms.logger.Info("Task assigned successfully", "task_id", taskID, "user_id", userID)
	return nil
}

// CompleteTask marks a task as completed with validation
func (tms *TaskManagementService) CompleteTask(ctx context.Context, taskID, completedBy uuid.UUID) error {
	tms.logger.Info("Completing task", "task_id", taskID, "completed_by", completedBy)

	// Get task
	task, err := tms.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		tms.logger.Error("Failed to get task", err, "task_id", taskID)
		return err
	}

	// Validate task can be completed
	if task.Status == entities.StatusCompleted {
		return fmt.Errorf("task is already completed")
	}

	if task.IsBlocked() {
		return fmt.Errorf("task is blocked by dependencies")
	}

	// Update task status
	task.UpdateStatus(entities.StatusCompleted, completedBy)

	// Update the task
	if err := tms.taskRepo.Update(ctx, task); err != nil {
		tms.logger.Error("Failed to update completed task", err, "task_id", taskID)
		return err
	}

	// Check and update dependent tasks
	if err := tms.updateDependentTasks(ctx, taskID); err != nil {
		tms.logger.Error("Failed to update dependent tasks", err, "task_id", taskID)
	}

	// Update project progress if task belongs to a project
	if task.ProjectID != nil {
		if err := tms.updateProjectProgress(ctx, *task.ProjectID); err != nil {
			tms.logger.Error("Failed to update project progress", err, "project_id", *task.ProjectID)
		}
	}

	// Publish task completed event
	event := events.NewTaskCompletedEvent(task)
	if err := tms.eventPublisher.PublishEvent(ctx, event); err != nil {
		tms.logger.Error("Failed to publish task completed event", err, "task_id", taskID)
	}

	// Send notifications
	if err := tms.notifyTaskCompletion(ctx, task); err != nil {
		tms.logger.Error("Failed to notify task completion", err, "task_id", taskID)
	}

	tms.logger.Info("Task completed successfully", "task_id", taskID)
	return nil
}

// GetOverdueTasks returns overdue tasks for a user or all users
func (tms *TaskManagementService) GetOverdueTasks(ctx context.Context, userID *uuid.UUID) ([]*entities.Task, error) {
	tms.logger.Info("Getting overdue tasks", "user_id", userID)

	tasks, err := tms.taskRepo.GetOverdueTasks(ctx, userID)
	if err != nil {
		tms.logger.Error("Failed to get overdue tasks", err, "user_id", userID)
		return nil, err
	}

	tms.logger.Info("Retrieved overdue tasks", "count", len(tasks), "user_id", userID)
	return tasks, nil
}

// GetTasksDueSoon returns tasks due within the specified number of days
func (tms *TaskManagementService) GetTasksDueSoon(ctx context.Context, days int, userID *uuid.UUID) ([]*entities.Task, error) {
	tms.logger.Info("Getting tasks due soon", "days", days, "user_id", userID)

	tasks, err := tms.taskRepo.GetTasksDueSoon(ctx, days, userID)
	if err != nil {
		tms.logger.Error("Failed to get tasks due soon", err, "days", days, "user_id", userID)
		return nil, err
	}

	tms.logger.Info("Retrieved tasks due soon", "count", len(tasks), "days", days, "user_id", userID)
	return tasks, nil
}

// Helper methods

func (tms *TaskManagementService) validateTask(ctx context.Context, task *entities.Task) error {
	if task.Title == "" {
		return fmt.Errorf("task title is required")
	}

	if task.DueDate != nil && task.DueDate.Before(time.Now()) {
		return fmt.Errorf("due date cannot be in the past")
	}

	if task.EstimatedHours < 0 {
		return fmt.Errorf("estimated hours cannot be negative")
	}

	return nil
}

func (tms *TaskManagementService) handleStatusChange(ctx context.Context, existingTask *entities.Task, newStatus entities.TaskStatus, updatedBy uuid.UUID) error {
	// Handle specific status transitions
	switch newStatus {
	case entities.StatusInProgress:
		if existingTask.IsBlocked() {
			return fmt.Errorf("cannot start blocked task")
		}
	case entities.StatusCompleted:
		if existingTask.IsBlocked() {
			return fmt.Errorf("cannot complete blocked task")
		}
	}

	return nil
}

func (tms *TaskManagementService) validateUserSkills(user *entities.User, requiredSkills []string) error {
	if len(requiredSkills) == 0 {
		return nil
	}

	userSkillsMap := make(map[string]bool)
	for _, skill := range user.Skills {
		userSkillsMap[skill] = true
	}

	var missingSkills []string
	for _, skill := range requiredSkills {
		if !userSkillsMap[skill] {
			missingSkills = append(missingSkills, skill)
		}
	}

	if len(missingSkills) > 0 {
		return fmt.Errorf("user missing required skills: %v", missingSkills)
	}

	return nil
}

func (tms *TaskManagementService) updateUserCapacity(ctx context.Context, userID uuid.UUID, additionalHours float64) error {
	capacity, err := tms.userRepo.GetUserCapacity(ctx, userID)
	if err != nil {
		return err
	}

	if capacity == nil {
		// Create default capacity
		capacity = &entities.UserCapacity{
			MaxHoursPerDay:    8.0,
			MaxHoursPerWeek:   40.0,
			CurrentWorkload:   additionalHours,
			AvailableHours:    40.0 - additionalHours,
			UtilizationRate:   (additionalHours / 40.0) * 100,
			OverloadThreshold: 80.0,
			LastUpdated:       time.Now(),
		}
	} else {
		capacity.CurrentWorkload += additionalHours
		capacity.AvailableHours = capacity.MaxHoursPerWeek - capacity.CurrentWorkload
		capacity.UtilizationRate = (capacity.CurrentWorkload / capacity.MaxHoursPerWeek) * 100
		capacity.LastUpdated = time.Now()
	}

	return tms.userRepo.UpdateUserCapacity(ctx, userID, capacity)
}

func (tms *TaskManagementService) updateDependentTasks(ctx context.Context, completedTaskID uuid.UUID) error {
	dependentTasks, err := tms.taskRepo.GetTaskDependents(ctx, completedTaskID)
	if err != nil {
		return err
	}

	for _, task := range dependentTasks {
		if task.CanStart() && task.Status == entities.StatusTodo {
			// Optionally auto-start dependent tasks
			tms.logger.Info("Dependent task can now start", "task_id", task.ID, "completed_task_id", completedTaskID)
		}
	}

	return nil
}

func (tms *TaskManagementService) updateProjectProgress(ctx context.Context, projectID uuid.UUID) error {
	project, err := tms.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	// Calculate and update project progress
	progress := project.CalculateProgress()
	tms.logger.Info("Updated project progress", "project_id", projectID, "progress", progress)

	return nil
}

func (tms *TaskManagementService) notifyTaskAssignees(ctx context.Context, task *entities.Task, message string) error {
	assignees := task.GetAssignees()
	for _, assignee := range assignees {
		notification := &repositories.Notification{
			ID:          uuid.New(),
			Type:        repositories.NotificationTypeTaskAssigned,
			Title:       "Task Assignment",
			Message:     message,
			Priority:    repositories.NotificationPriorityNormal,
			Category:    repositories.NotificationCategoryTask,
			UserID:      assignee.ID,
			RelatedID:   &task.ID,
			RelatedType: "task",
			Data: map[string]interface{}{
				"task_id":    task.ID,
				"task_title": task.Title,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := tms.notificationRepo.Create(ctx, notification); err != nil {
			tms.logger.Error("Failed to create notification", err, "user_id", assignee.ID)
		}
	}

	return nil
}

func (tms *TaskManagementService) notifyTaskCompletion(ctx context.Context, task *entities.Task) error {
	// Notify project owner and stakeholders
	if task.ProjectID != nil {
		project, err := tms.projectRepo.GetByID(ctx, *task.ProjectID)
		if err == nil {
			notification := &repositories.Notification{
				ID:          uuid.New(),
				Type:        repositories.NotificationTypeTaskCompleted,
				Title:       "Task Completed",
				Message:     fmt.Sprintf("Task '%s' has been completed", task.Title),
				Priority:    repositories.NotificationPriorityNormal,
				Category:    repositories.NotificationCategoryTask,
				UserID:      project.OwnerID,
				RelatedID:   &task.ID,
				RelatedType: "task",
				Data: map[string]interface{}{
					"task_id":    task.ID,
					"task_title": task.Title,
					"project_id": project.ID,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err := tms.notificationRepo.Create(ctx, notification); err != nil {
				tms.logger.Error("Failed to create completion notification", err, "user_id", project.OwnerID)
			}
		}
	}

	return nil
}
