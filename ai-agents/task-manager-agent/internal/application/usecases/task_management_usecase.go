package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/task-manager-agent/internal/domain/entities"
	"go-coffee-ai-agents/task-manager-agent/internal/domain/repositories"
	"go-coffee-ai-agents/task-manager-agent/internal/domain/services"
)

// TaskManagementUseCase provides application-level task management operations
type TaskManagementUseCase struct {
	taskService      *services.TaskManagementService
	workflowService  *services.WorkflowOrchestrationService
	aiService        *services.AITaskAutomationService
	taskRepo         repositories.TaskRepository
	projectRepo      repositories.ProjectRepository
	userRepo         repositories.UserRepository
	workflowRepo     repositories.WorkflowRepository
	notificationRepo repositories.NotificationRepository
	logger           services.Logger
}

// NewTaskManagementUseCase creates a new task management use case
func NewTaskManagementUseCase(
	taskService *services.TaskManagementService,
	workflowService *services.WorkflowOrchestrationService,
	aiService *services.AITaskAutomationService,
	taskRepo repositories.TaskRepository,
	projectRepo repositories.ProjectRepository,
	userRepo repositories.UserRepository,
	workflowRepo repositories.WorkflowRepository,
	notificationRepo repositories.NotificationRepository,
	logger services.Logger,
) *TaskManagementUseCase {
	return &TaskManagementUseCase{
		taskService:      taskService,
		workflowService:  workflowService,
		aiService:        aiService,
		taskRepo:         taskRepo,
		projectRepo:      projectRepo,
		userRepo:         userRepo,
		workflowRepo:     workflowRepo,
		notificationRepo: notificationRepo,
		logger:           logger,
	}
}

// CreateTaskRequest represents a request to create a task
type CreateTaskRequest struct {
	Title             string                     `json:"title" validate:"required,min=1,max=200"`
	Description       string                     `json:"description,omitempty" validate:"max=2000"`
	Type              entities.TaskType          `json:"type" validate:"required"`
	Priority          entities.TaskPriority      `json:"priority,omitempty"`
	Urgency           entities.TaskUrgency       `json:"urgency,omitempty"`
	Complexity        entities.TaskComplexity    `json:"complexity,omitempty"`
	ProjectID         *uuid.UUID                 `json:"project_id,omitempty"`
	ParentTaskID      *uuid.UUID                 `json:"parent_task_id,omitempty"`
	AssigneeIDs       []uuid.UUID                `json:"assignee_ids,omitempty"`
	Tags              []string                   `json:"tags,omitempty"`
	Labels            []string                   `json:"labels,omitempty"`
	EstimatedHours    float64                    `json:"estimated_hours,omitempty" validate:"min=0"`
	DueDate           *time.Time                 `json:"due_date,omitempty"`
	StartDate         *time.Time                 `json:"start_date,omitempty"`
	Location          string                     `json:"location,omitempty"`
	Equipment         []string                   `json:"equipment,omitempty"`
	Skills            []string                   `json:"skills,omitempty"`
	CustomFields      map[string]interface{}     `json:"custom_fields,omitempty"`
	IsRecurring       bool                       `json:"is_recurring"`
	RecurrenceRule    *entities.RecurrenceRule   `json:"recurrence_rule,omitempty"`
	AutoAssign        bool                       `json:"auto_assign"`
	TriggerWorkflow   bool                       `json:"trigger_workflow"`
	WorkflowID        *uuid.UUID                 `json:"workflow_id,omitempty"`
	CreatedBy         uuid.UUID                  `json:"created_by" validate:"required"`
}

// CreateTaskResponse represents the response from creating a task
type CreateTaskResponse struct {
	Task              *entities.Task             `json:"task"`
	Assignments       []*entities.TaskAssignment `json:"assignments,omitempty"`
	WorkflowExecution *entities.WorkflowExecution `json:"workflow_execution,omitempty"`
	Recommendations   []string                   `json:"recommendations,omitempty"`
}

// CreateTask creates a new task with advanced features
func (uc *TaskManagementUseCase) CreateTask(ctx context.Context, req *CreateTaskRequest) (*CreateTaskResponse, error) {
	uc.logger.Info("Creating task", "title", req.Title, "type", req.Type, "created_by", req.CreatedBy)

	// Create the task entity
	task := entities.NewTask(req.Title, req.Description, req.Type, req.CreatedBy)
	
	// Set optional fields
	if req.Priority != "" {
		task.Priority = req.Priority
	}
	if req.Urgency != "" {
		task.Urgency = req.Urgency
	}
	if req.Complexity != "" {
		task.Complexity = req.Complexity
	}
	if req.ProjectID != nil {
		task.ProjectID = req.ProjectID
	}
	if req.ParentTaskID != nil {
		task.ParentTaskID = req.ParentTaskID
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}
	if req.StartDate != nil {
		task.StartDate = req.StartDate
	}
	
	task.Tags = req.Tags
	task.Labels = req.Labels
	task.EstimatedHours = req.EstimatedHours
	task.Location = req.Location
	task.Equipment = req.Equipment
	task.Skills = req.Skills
	task.CustomFields = req.CustomFields
	task.IsRecurring = req.IsRecurring
	task.RecurrenceRule = req.RecurrenceRule

	// Use AI to enhance task if enabled
	if uc.aiService != nil {
		if err := uc.aiService.EnhanceTask(ctx, task); err != nil {
			uc.logger.Warn("Failed to enhance task with AI", "task_id", task.ID, "error", err)
		}
	}

	// Create the task using domain service
	if err := uc.taskService.CreateTask(ctx, task); err != nil {
		uc.logger.Error("Failed to create task", err, "task_id", task.ID)
		return nil, err
	}

	response := &CreateTaskResponse{
		Task: task,
	}

	// Handle auto-assignment
	if req.AutoAssign && len(req.AssigneeIDs) == 0 {
		assignees, err := uc.aiService.RecommendAssignees(ctx, task)
		if err != nil {
			uc.logger.Warn("Failed to get AI assignee recommendations", "task_id", task.ID, "error", err)
		} else {
			req.AssigneeIDs = assignees
		}
	}

	// Assign users to the task
	var assignments []*entities.TaskAssignment
	for _, assigneeID := range req.AssigneeIDs {
		if err := uc.taskService.AssignTask(ctx, task.ID, assigneeID, req.CreatedBy, entities.RoleAssignee); err != nil {
			uc.logger.Error("Failed to assign task", err, "task_id", task.ID, "assignee_id", assigneeID)
			continue
		}
		
		// Get the assignment for response
		taskAssignments, err := uc.taskRepo.GetTaskAssignments(ctx, task.ID)
		if err == nil {
			for _, assignment := range taskAssignments {
				if assignment.UserID == assigneeID {
					assignments = append(assignments, assignment)
					break
				}
			}
		}
	}
	response.Assignments = assignments

	// Trigger workflow if requested
	if req.TriggerWorkflow && req.WorkflowID != nil {
		execution, err := uc.workflowService.StartWorkflow(ctx, *req.WorkflowID, req.CreatedBy, map[string]interface{}{
			"task_id": task.ID,
			"trigger": "task_created",
		})
		if err != nil {
			uc.logger.Error("Failed to start workflow", err, "workflow_id", *req.WorkflowID, "task_id", task.ID)
		} else {
			response.WorkflowExecution = execution
		}
	}

	// Get AI recommendations
	if uc.aiService != nil {
		recommendations, err := uc.aiService.GetTaskRecommendations(ctx, task)
		if err != nil {
			uc.logger.Warn("Failed to get AI recommendations", "task_id", task.ID, "error", err)
		} else {
			response.Recommendations = recommendations
		}
	}

	uc.logger.Info("Task created successfully", "task_id", task.ID, "title", task.Title)
	return response, nil
}

// UpdateTaskRequest represents a request to update a task
type UpdateTaskRequest struct {
	ID                uuid.UUID                  `json:"id" validate:"required"`
	Title             *string                    `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description       *string                    `json:"description,omitempty" validate:"omitempty,max=2000"`
	Status            *entities.TaskStatus       `json:"status,omitempty"`
	Priority          *entities.TaskPriority     `json:"priority,omitempty"`
	Urgency           *entities.TaskUrgency      `json:"urgency,omitempty"`
	Complexity        *entities.TaskComplexity   `json:"complexity,omitempty"`
	ProgressPercent   *float64                   `json:"progress_percent,omitempty" validate:"omitempty,min=0,max=100"`
	EstimatedHours    *float64                   `json:"estimated_hours,omitempty" validate:"omitempty,min=0"`
	DueDate           *time.Time                 `json:"due_date,omitempty"`
	StartDate         *time.Time                 `json:"start_date,omitempty"`
	Tags              []string                   `json:"tags,omitempty"`
	Labels            []string                   `json:"labels,omitempty"`
	Location          *string                    `json:"location,omitempty"`
	Equipment         []string                   `json:"equipment,omitempty"`
	Skills            []string                   `json:"skills,omitempty"`
	CustomFields      map[string]interface{}     `json:"custom_fields,omitempty"`
	UpdatedBy         uuid.UUID                  `json:"updated_by" validate:"required"`
	TriggerWorkflow   bool                       `json:"trigger_workflow"`
}

// UpdateTask updates an existing task
func (uc *TaskManagementUseCase) UpdateTask(ctx context.Context, req *UpdateTaskRequest) (*entities.Task, error) {
	uc.logger.Info("Updating task", "task_id", req.ID, "updated_by", req.UpdatedBy)

	// Get existing task
	task, err := uc.taskRepo.GetByID(ctx, req.ID)
	if err != nil {
		uc.logger.Error("Failed to get task", err, "task_id", req.ID)
		return nil, err
	}

	// Store previous state for comparison
	previousTask := *task

	// Update fields
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		task.UpdateStatus(*req.Status, req.UpdatedBy)
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.Urgency != nil {
		task.Urgency = *req.Urgency
	}
	if req.Complexity != nil {
		task.Complexity = *req.Complexity
	}
	if req.ProgressPercent != nil {
		task.UpdateProgress(*req.ProgressPercent, req.UpdatedBy)
	}
	if req.EstimatedHours != nil {
		task.EstimatedHours = *req.EstimatedHours
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}
	if req.StartDate != nil {
		task.StartDate = req.StartDate
	}
	if req.Location != nil {
		task.Location = *req.Location
	}
	
	if req.Tags != nil {
		task.Tags = req.Tags
	}
	if req.Labels != nil {
		task.Labels = req.Labels
	}
	if req.Equipment != nil {
		task.Equipment = req.Equipment
	}
	if req.Skills != nil {
		task.Skills = req.Skills
	}
	if req.CustomFields != nil {
		task.CustomFields = req.CustomFields
	}

	task.UpdatedBy = req.UpdatedBy
	task.UpdatedAt = time.Now()
	task.Version++

	// Update the task using domain service
	if err := uc.taskService.UpdateTask(ctx, task); err != nil {
		uc.logger.Error("Failed to update task", err, "task_id", req.ID)
		return nil, err
	}

	// Trigger workflow if requested and status changed
	if req.TriggerWorkflow && previousTask.Status != task.Status {
		workflows, err := uc.workflowRepo.GetTriggersByEvent(ctx, fmt.Sprintf("task.status.%s", task.Status))
		if err == nil {
			for _, trigger := range workflows {
				if trigger.IsActive {
					_, err := uc.workflowService.StartWorkflow(ctx, trigger.WorkflowID, req.UpdatedBy, map[string]interface{}{
						"task_id":        task.ID,
						"trigger":        "status_change",
						"previous_status": previousTask.Status,
						"new_status":     task.Status,
					})
					if err != nil {
						uc.logger.Error("Failed to start workflow", err, "workflow_id", trigger.WorkflowID)
					}
				}
			}
		}
	}

	uc.logger.Info("Task updated successfully", "task_id", task.ID)
	return task, nil
}

// AssignTaskRequest represents a request to assign a task
type AssignTaskRequest struct {
	TaskID     uuid.UUID                `json:"task_id" validate:"required"`
	UserID     uuid.UUID                `json:"user_id" validate:"required"`
	Role       entities.AssignmentRole  `json:"role" validate:"required"`
	Allocation float64                  `json:"allocation" validate:"min=0,max=100"`
	AssignedBy uuid.UUID                `json:"assigned_by" validate:"required"`
	Notes      string                   `json:"notes,omitempty"`
}

// AssignTask assigns a task to a user
func (uc *TaskManagementUseCase) AssignTask(ctx context.Context, req *AssignTaskRequest) (*entities.TaskAssignment, error) {
	uc.logger.Info("Assigning task", "task_id", req.TaskID, "user_id", req.UserID, "role", req.Role)

	// Use domain service to assign task
	if err := uc.taskService.AssignTask(ctx, req.TaskID, req.UserID, req.AssignedBy, req.Role); err != nil {
		uc.logger.Error("Failed to assign task", err, "task_id", req.TaskID, "user_id", req.UserID)
		return nil, err
	}

	// Get the created assignment
	assignments, err := uc.taskRepo.GetTaskAssignments(ctx, req.TaskID)
	if err != nil {
		uc.logger.Error("Failed to get task assignments", err, "task_id", req.TaskID)
		return nil, err
	}

	// Find the assignment for this user
	for _, assignment := range assignments {
		if assignment.UserID == req.UserID && assignment.IsActive {
			// Update allocation and notes if provided
			if req.Allocation > 0 {
				assignment.Allocation = req.Allocation
			}
			if req.Notes != "" {
				assignment.Notes = req.Notes
			}
			
			if err := uc.taskRepo.UpdateAssignment(ctx, assignment); err != nil {
				uc.logger.Error("Failed to update assignment", err, "assignment_id", assignment.ID)
			}
			
			uc.logger.Info("Task assigned successfully", "task_id", req.TaskID, "user_id", req.UserID)
			return assignment, nil
		}
	}

	return nil, fmt.Errorf("assignment not found after creation")
}

// CompleteTaskRequest represents a request to complete a task
type CompleteTaskRequest struct {
	TaskID      uuid.UUID `json:"task_id" validate:"required"`
	CompletedBy uuid.UUID `json:"completed_by" validate:"required"`
	Notes       string    `json:"notes,omitempty"`
	ActualHours float64   `json:"actual_hours,omitempty" validate:"min=0"`
}

// CompleteTask marks a task as completed
func (uc *TaskManagementUseCase) CompleteTask(ctx context.Context, req *CompleteTaskRequest) (*entities.Task, error) {
	uc.logger.Info("Completing task", "task_id", req.TaskID, "completed_by", req.CompletedBy)

	// Get the task first to update actual hours if provided
	if req.ActualHours > 0 {
		task, err := uc.taskRepo.GetByID(ctx, req.TaskID)
		if err != nil {
			uc.logger.Error("Failed to get task", err, "task_id", req.TaskID)
			return nil, err
		}

		task.ActualHours = req.ActualHours
		task.UpdatedBy = req.CompletedBy
		task.UpdatedAt = time.Now()
		task.Version++

		if err := uc.taskRepo.Update(ctx, task); err != nil {
			uc.logger.Error("Failed to update task hours", err, "task_id", req.TaskID)
		}
	}

	// Use domain service to complete task
	if err := uc.taskService.CompleteTask(ctx, req.TaskID, req.CompletedBy); err != nil {
		uc.logger.Error("Failed to complete task", err, "task_id", req.TaskID)
		return nil, err
	}

	// Get updated task
	task, err := uc.taskRepo.GetByID(ctx, req.TaskID)
	if err != nil {
		uc.logger.Error("Failed to get completed task", err, "task_id", req.TaskID)
		return nil, err
	}

	// Add completion comment if notes provided
	if req.Notes != "" {
		comment := &entities.TaskComment{
			ID:        uuid.New(),
			TaskID:    req.TaskID,
			UserID:    req.CompletedBy,
			Content:   fmt.Sprintf("Task completed. Notes: %s", req.Notes),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := uc.taskRepo.AddComment(ctx, comment); err != nil {
			uc.logger.Error("Failed to add completion comment", err, "task_id", req.TaskID)
		}
	}

	uc.logger.Info("Task completed successfully", "task_id", req.TaskID)
	return task, nil
}

// GetTasksRequest represents a request to get tasks with filtering
type GetTasksRequest struct {
	Filter     *repositories.TaskFilter `json:"filter,omitempty"`
	UserID     *uuid.UUID               `json:"user_id,omitempty"`
	ProjectID  *uuid.UUID               `json:"project_id,omitempty"`
	IncludeSubTasks bool                `json:"include_sub_tasks"`
	IncludeDetails  bool                `json:"include_details"`
}

// GetTasks retrieves tasks with filtering and pagination
func (uc *TaskManagementUseCase) GetTasks(ctx context.Context, req *GetTasksRequest) ([]*entities.Task, error) {
	uc.logger.Info("Getting tasks", "user_id", req.UserID, "project_id", req.ProjectID)

	var tasks []*entities.Task
	var err error

	// Apply different query strategies based on request
	if req.UserID != nil {
		tasks, err = uc.taskRepo.ListByAssignee(ctx, *req.UserID, req.Filter)
	} else if req.ProjectID != nil {
		tasks, err = uc.taskRepo.ListByProject(ctx, *req.ProjectID, req.Filter)
	} else {
		tasks, err = uc.taskRepo.List(ctx, req.Filter)
	}

	if err != nil {
		uc.logger.Error("Failed to get tasks", err)
		return nil, err
	}

	// Load additional details if requested
	if req.IncludeDetails {
		for _, task := range tasks {
			// Load assignments
			assignments, err := uc.taskRepo.GetTaskAssignments(ctx, task.ID)
			if err == nil {
				task.Assignments = assignments
			}

			// Load comments
			comments, err := uc.taskRepo.GetTaskComments(ctx, task.ID)
			if err == nil {
				task.Comments = comments
			}

			// Load time entries
			timeEntries, err := uc.taskRepo.GetTimeEntries(ctx, task.ID)
			if err == nil {
				task.TimeEntries = timeEntries
			}
		}
	}

	uc.logger.Info("Retrieved tasks", "count", len(tasks))
	return tasks, nil
}

// GetTaskMetricsRequest represents a request to get task metrics
type GetTaskMetricsRequest struct {
	Filter    *repositories.TaskMetricsFilter `json:"filter" validate:"required"`
	UserID    *uuid.UUID                      `json:"user_id,omitempty"`
	ProjectID *uuid.UUID                      `json:"project_id,omitempty"`
	TeamID    *uuid.UUID                      `json:"team_id,omitempty"`
}

// GetTaskMetrics retrieves task metrics and analytics
func (uc *TaskManagementUseCase) GetTaskMetrics(ctx context.Context, req *GetTaskMetricsRequest) (*repositories.TaskMetrics, error) {
	uc.logger.Info("Getting task metrics", "period", req.Filter.Period)

	// Apply user/project/team filters
	if req.UserID != nil {
		req.Filter.UserIDs = []uuid.UUID{*req.UserID}
	}
	if req.ProjectID != nil {
		req.Filter.ProjectIDs = []uuid.UUID{*req.ProjectID}
	}
	if req.TeamID != nil {
		req.Filter.TeamIDs = []uuid.UUID{*req.TeamID}
	}

	metrics, err := uc.taskRepo.GetTaskMetrics(ctx, req.Filter)
	if err != nil {
		uc.logger.Error("Failed to get task metrics", err)
		return nil, err
	}

	uc.logger.Info("Retrieved task metrics", "total_tasks", metrics.TotalTasks, "completion_rate", metrics.CompletionRate)
	return metrics, nil
}
