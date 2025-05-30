package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/task-cli/models"
	"github.com/go-redis/redis/v8"
)

// TaskRepository defines the interface for task data access
type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	GetByID(ctx context.Context, id string) (*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter models.TaskFilter, offset, limit int) ([]*models.Task, int, error)
	Search(ctx context.Context, query string, offset, limit int) ([]*models.Task, int, error)
	GetStats(ctx context.Context) (*models.TaskStats, error)
	GetByAssignee(ctx context.Context, assignee string) ([]*models.Task, error)
	GetByStatus(ctx context.Context, status models.TaskStatus) ([]*models.Task, error)
	GetOverdue(ctx context.Context) ([]*models.Task, error)
}

// RedisTaskRepository implements TaskRepository using Redis
type RedisTaskRepository struct {
	client *redis.Client
}

// NewRedisTaskRepository creates a new Redis task repository
func NewRedisTaskRepository(client *redis.Client) TaskRepository {
	return &RedisTaskRepository{
		client: client,
	}
}

// Redis key patterns
const (
	taskKeyPrefix        = "task:"
	taskAllKey          = "tasks:all"
	taskStatusKeyPrefix = "tasks:by_status:"
	taskAssigneeKeyPrefix = "tasks:by_assignee:"
	taskPriorityKeyPrefix = "tasks:by_priority:"
	taskCreatorKeyPrefix = "tasks:by_creator:"
	taskTagKeyPrefix    = "tasks:by_tag:"
	taskStatsKey        = "tasks:stats"
)

// Create creates a new task in Redis
func (r *RedisTaskRepository) Create(ctx context.Context, task *models.Task) error {
	pipe := r.client.TxPipeline()

	// Store task data
	taskKey := taskKeyPrefix + task.ID
	taskJSON, err := task.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	pipe.Set(ctx, taskKey, taskJSON, 0)

	// Add to indexes
	pipe.SAdd(ctx, taskAllKey, task.ID)
	pipe.SAdd(ctx, taskStatusKeyPrefix+string(task.Status), task.ID)
	pipe.SAdd(ctx, taskPriorityKeyPrefix+string(task.Priority), task.ID)
	
	if task.Assignee != "" {
		pipe.SAdd(ctx, taskAssigneeKeyPrefix+task.Assignee, task.ID)
	}
	if task.Creator != "" {
		pipe.SAdd(ctx, taskCreatorKeyPrefix+task.Creator, task.ID)
	}

	// Add tag indexes
	for _, tag := range task.Tags {
		pipe.SAdd(ctx, taskTagKeyPrefix+tag, task.ID)
	}

	// Set expiration for task (optional, for cleanup)
	pipe.Expire(ctx, taskKey, 365*24*time.Hour) // 1 year

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

// GetByID retrieves a task by ID
func (r *RedisTaskRepository) GetByID(ctx context.Context, id string) (*models.Task, error) {
	taskKey := taskKeyPrefix + id
	data, err := r.client.Get(ctx, taskKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("task not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	task, err := models.FromJSON([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return task, nil
}

// Update updates an existing task
func (r *RedisTaskRepository) Update(ctx context.Context, task *models.Task) error {
	// Get old task for index cleanup
	oldTask, err := r.GetByID(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing task: %w", err)
	}

	pipe := r.client.TxPipeline()

	// Update task data
	taskKey := taskKeyPrefix + task.ID
	taskJSON, err := task.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	pipe.Set(ctx, taskKey, taskJSON, 0)

	// Clean old indexes
	r.cleanTaskIndexes(pipe, ctx, oldTask)

	// Add new indexes
	r.addTaskIndexes(pipe, ctx, task)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

// Delete deletes a task
func (r *RedisTaskRepository) Delete(ctx context.Context, id string) error {
	// Get task for index cleanup
	task, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get task for deletion: %w", err)
	}

	pipe := r.client.TxPipeline()

	// Delete task data
	taskKey := taskKeyPrefix + id
	pipe.Del(ctx, taskKey)

	// Clean indexes
	r.cleanTaskIndexes(pipe, ctx, task)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

// List retrieves tasks with filtering and pagination
func (r *RedisTaskRepository) List(ctx context.Context, filter models.TaskFilter, offset, limit int) ([]*models.Task, int, error) {
	// Build intersection keys for filtering
	var intersectKeys []string

	// Filter by status
	if len(filter.Status) > 0 {
		for _, status := range filter.Status {
			intersectKeys = append(intersectKeys, taskStatusKeyPrefix+string(status))
		}
	}

	// Filter by priority
	if len(filter.Priority) > 0 {
		for _, priority := range filter.Priority {
			intersectKeys = append(intersectKeys, taskPriorityKeyPrefix+string(priority))
		}
	}

	// Filter by assignee
	if len(filter.Assignee) > 0 {
		for _, assignee := range filter.Assignee {
			intersectKeys = append(intersectKeys, taskAssigneeKeyPrefix+assignee)
		}
	}

	// Filter by creator
	if len(filter.Creator) > 0 {
		for _, creator := range filter.Creator {
			intersectKeys = append(intersectKeys, taskCreatorKeyPrefix+creator)
		}
	}

	// Filter by tags
	if len(filter.Tags) > 0 {
		for _, tag := range filter.Tags {
			intersectKeys = append(intersectKeys, taskTagKeyPrefix+tag)
		}
	}

	// Get task IDs
	var taskIDs []string
	var err error

	if len(intersectKeys) == 0 {
		// No filters, get all tasks
		taskIDs, err = r.client.SMembers(ctx, taskAllKey).Result()
	} else if len(intersectKeys) == 1 {
		// Single filter
		taskIDs, err = r.client.SMembers(ctx, intersectKeys[0]).Result()
	} else {
		// Multiple filters, use intersection
		tempKey := fmt.Sprintf("temp:filter:%d", time.Now().UnixNano())
		defer r.client.Del(ctx, tempKey)

		_, err = r.client.SInterStore(ctx, tempKey, intersectKeys...).Result()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to intersect filters: %w", err)
		}

		taskIDs, err = r.client.SMembers(ctx, tempKey).Result()
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get task IDs: %w", err)
	}

	// Get tasks and apply additional filters
	var tasks []*models.Task
	for _, id := range taskIDs {
		task, err := r.GetByID(ctx, id)
		if err != nil {
			continue // Skip invalid tasks
		}

		// Apply date filters
		if filter.DueBefore != nil && task.DueDate != nil && task.DueDate.After(*filter.DueBefore) {
			continue
		}
		if filter.DueAfter != nil && task.DueDate != nil && task.DueDate.Before(*filter.DueAfter) {
			continue
		}

		// Apply search filter
		if filter.Search != "" {
			searchLower := strings.ToLower(filter.Search)
			if !strings.Contains(strings.ToLower(task.Title), searchLower) &&
				!strings.Contains(strings.ToLower(task.Description), searchLower) {
				continue
			}
		}

		tasks = append(tasks, task)
	}

	total := len(tasks)

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

// Search searches tasks by query
func (r *RedisTaskRepository) Search(ctx context.Context, query string, offset, limit int) ([]*models.Task, int, error) {
	filter := models.TaskFilter{
		Search: query,
	}
	return r.List(ctx, filter, offset, limit)
}

// GetStats returns task statistics
func (r *RedisTaskRepository) GetStats(ctx context.Context) (*models.TaskStats, error) {
	stats := &models.TaskStats{
		ByStatus:   make(map[models.TaskStatus]int),
		ByPriority: make(map[models.TaskPriority]int),
		ByAssignee: make(map[string]int),
	}

	// Get total count
	total, err := r.client.SCard(ctx, taskAllKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats.Total = int(total)

	// Get counts by status
	for _, status := range models.GetAllStatuses() {
		count, err := r.client.SCard(ctx, taskStatusKeyPrefix+string(status)).Result()
		if err != nil {
			continue
		}
		stats.ByStatus[status] = int(count)
	}

	// Get counts by priority
	for _, priority := range models.GetAllPriorities() {
		count, err := r.client.SCard(ctx, taskPriorityKeyPrefix+string(priority)).Result()
		if err != nil {
			continue
		}
		stats.ByPriority[priority] = int(count)
	}

	// Get all tasks to calculate additional stats
	taskIDs, err := r.client.SMembers(ctx, taskAllKey).Result()
	if err != nil {
		return stats, nil // Return partial stats
	}

	for _, id := range taskIDs {
		task, err := r.GetByID(ctx, id)
		if err != nil {
			continue
		}

		// Count by assignee
		if task.Assignee != "" {
			stats.ByAssignee[task.Assignee]++
		}

		// Count overdue, due today, due this week
		if task.IsOverdue() {
			stats.Overdue++
		}
		if task.IsDueToday() {
			stats.DueToday++
		}
		if task.IsDueThisWeek() {
			stats.DueThisWeek++
		}
	}

	return stats, nil
}

// Helper methods

func (r *RedisTaskRepository) cleanTaskIndexes(pipe redis.Pipeliner, ctx context.Context, task *models.Task) {
	pipe.SRem(ctx, taskAllKey, task.ID)
	pipe.SRem(ctx, taskStatusKeyPrefix+string(task.Status), task.ID)
	pipe.SRem(ctx, taskPriorityKeyPrefix+string(task.Priority), task.ID)
	
	if task.Assignee != "" {
		pipe.SRem(ctx, taskAssigneeKeyPrefix+task.Assignee, task.ID)
	}
	if task.Creator != "" {
		pipe.SRem(ctx, taskCreatorKeyPrefix+task.Creator, task.ID)
	}

	for _, tag := range task.Tags {
		pipe.SRem(ctx, taskTagKeyPrefix+tag, task.ID)
	}
}

func (r *RedisTaskRepository) addTaskIndexes(pipe redis.Pipeliner, ctx context.Context, task *models.Task) {
	pipe.SAdd(ctx, taskAllKey, task.ID)
	pipe.SAdd(ctx, taskStatusKeyPrefix+string(task.Status), task.ID)
	pipe.SAdd(ctx, taskPriorityKeyPrefix+string(task.Priority), task.ID)
	
	if task.Assignee != "" {
		pipe.SAdd(ctx, taskAssigneeKeyPrefix+task.Assignee, task.ID)
	}
	if task.Creator != "" {
		pipe.SAdd(ctx, taskCreatorKeyPrefix+task.Creator, task.ID)
	}

	for _, tag := range task.Tags {
		pipe.SAdd(ctx, taskTagKeyPrefix+tag, task.ID)
	}
}

// GetByAssignee retrieves tasks by assignee
func (r *RedisTaskRepository) GetByAssignee(ctx context.Context, assignee string) ([]*models.Task, error) {
	filter := models.TaskFilter{
		Assignee: []string{assignee},
	}
	tasks, _, err := r.List(ctx, filter, 0, 1000) // Get all tasks for assignee
	return tasks, err
}

// GetByStatus retrieves tasks by status
func (r *RedisTaskRepository) GetByStatus(ctx context.Context, status models.TaskStatus) ([]*models.Task, error) {
	filter := models.TaskFilter{
		Status: []models.TaskStatus{status},
	}
	tasks, _, err := r.List(ctx, filter, 0, 1000) // Get all tasks with status
	return tasks, err
}

// GetOverdue retrieves overdue tasks
func (r *RedisTaskRepository) GetOverdue(ctx context.Context) ([]*models.Task, error) {
	now := time.Now()
	filter := models.TaskFilter{
		DueBefore: &now,
		Status:    []models.TaskStatus{models.StatusPending, models.StatusInProgress},
	}
	tasks, _, err := r.List(ctx, filter, 0, 1000) // Get all overdue tasks
	return tasks, err
}
