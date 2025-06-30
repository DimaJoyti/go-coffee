package repositories

import (
	"context"
	"time"

	"go-coffee-ai-agents/task-manager-agent/internal/domain/entities"

	"github.com/google/uuid"
)

// TaskRepository defines the interface for task data access
type TaskRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, task *entities.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Task, error)
	Update(ctx context.Context, task *entities.Task) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Listing and filtering
	List(ctx context.Context, filter *TaskFilter) ([]*entities.Task, error)
	ListByProject(ctx context.Context, projectID uuid.UUID, filter *TaskFilter) ([]*entities.Task, error)
	ListByAssignee(ctx context.Context, userID uuid.UUID, filter *TaskFilter) ([]*entities.Task, error)
	ListByStatus(ctx context.Context, status entities.TaskStatus, filter *TaskFilter) ([]*entities.Task, error)
	ListByPriority(ctx context.Context, priority entities.TaskPriority, filter *TaskFilter) ([]*entities.Task, error)
	ListByDueDate(ctx context.Context, from, to time.Time, filter *TaskFilter) ([]*entities.Task, error)

	// Advanced queries
	GetOverdueTasks(ctx context.Context, userID *uuid.UUID) ([]*entities.Task, error)
	GetTasksDueToday(ctx context.Context, userID *uuid.UUID) ([]*entities.Task, error)
	GetTasksDueSoon(ctx context.Context, days int, userID *uuid.UUID) ([]*entities.Task, error)
	GetBlockedTasks(ctx context.Context, userID *uuid.UUID) ([]*entities.Task, error)
	GetTasksInProgress(ctx context.Context, userID *uuid.UUID) ([]*entities.Task, error)
	GetCompletedTasks(ctx context.Context, userID *uuid.UUID, since time.Time) ([]*entities.Task, error)

	// Dependencies
	GetTaskDependencies(ctx context.Context, taskID uuid.UUID) ([]*entities.TaskDependency, error)
	GetTaskDependents(ctx context.Context, taskID uuid.UUID) ([]*entities.Task, error)
	AddDependency(ctx context.Context, dependency *entities.TaskDependency) error
	RemoveDependency(ctx context.Context, dependencyID uuid.UUID) error

	// Assignments
	GetTaskAssignments(ctx context.Context, taskID uuid.UUID) ([]*entities.TaskAssignment, error)
	AddAssignment(ctx context.Context, assignment *entities.TaskAssignment) error
	UpdateAssignment(ctx context.Context, assignment *entities.TaskAssignment) error
	RemoveAssignment(ctx context.Context, assignmentID uuid.UUID) error

	// Comments and attachments
	GetTaskComments(ctx context.Context, taskID uuid.UUID) ([]*entities.TaskComment, error)
	AddComment(ctx context.Context, comment *entities.TaskComment) error
	UpdateComment(ctx context.Context, comment *entities.TaskComment) error
	DeleteComment(ctx context.Context, commentID uuid.UUID) error

	GetTaskAttachments(ctx context.Context, taskID uuid.UUID) ([]*entities.TaskAttachment, error)
	AddAttachment(ctx context.Context, attachment *entities.TaskAttachment) error
	DeleteAttachment(ctx context.Context, attachmentID uuid.UUID) error

	// Time tracking
	GetTimeEntries(ctx context.Context, taskID uuid.UUID) ([]*entities.TimeEntry, error)
	AddTimeEntry(ctx context.Context, entry *entities.TimeEntry) error
	UpdateTimeEntry(ctx context.Context, entry *entities.TimeEntry) error
	DeleteTimeEntry(ctx context.Context, entryID uuid.UUID) error

	// Analytics and reporting
	GetTaskMetrics(ctx context.Context, filter *TaskMetricsFilter) (*TaskMetrics, error)
	GetUserProductivity(ctx context.Context, userID uuid.UUID, period time.Duration) (*UserProductivity, error)
	GetProjectProgress(ctx context.Context, projectID uuid.UUID) (*ProjectProgress, error)
	GetTaskTrends(ctx context.Context, period time.Duration, groupBy string) (map[string]interface{}, error)

	// Search and advanced queries
	Search(ctx context.Context, query string, filter *TaskFilter) ([]*entities.Task, error)
	GetTasksByTags(ctx context.Context, tags []string, filter *TaskFilter) ([]*entities.Task, error)
	GetTasksByCustomField(ctx context.Context, field string, value interface{}, filter *TaskFilter) ([]*entities.Task, error)

	// Bulk operations
	BulkCreate(ctx context.Context, tasks []*entities.Task) error
	BulkUpdate(ctx context.Context, tasks []*entities.Task) error
	BulkUpdateStatus(ctx context.Context, taskIDs []uuid.UUID, status entities.TaskStatus, updatedBy uuid.UUID) error
	BulkAssign(ctx context.Context, taskIDs []uuid.UUID, userID uuid.UUID, assignedBy uuid.UUID) error

	// Archiving
	Archive(ctx context.Context, taskID uuid.UUID, archivedBy uuid.UUID) error
	Unarchive(ctx context.Context, taskID uuid.UUID, unarchivedBy uuid.UUID) error
	GetArchivedTasks(ctx context.Context, filter *TaskFilter) ([]*entities.Task, error)

	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context, repo TaskRepository) error) error
}

// TaskFilter defines filtering options for task queries
type TaskFilter struct {
	ProjectIDs        []uuid.UUID               `json:"project_ids,omitempty"`
	AssigneeIDs       []uuid.UUID               `json:"assignee_ids,omitempty"`
	CreatorIDs        []uuid.UUID               `json:"creator_ids,omitempty"`
	Types             []entities.TaskType       `json:"types,omitempty"`
	Statuses          []entities.TaskStatus     `json:"statuses,omitempty"`
	Priorities        []entities.TaskPriority   `json:"priorities,omitempty"`
	Urgencies         []entities.TaskUrgency    `json:"urgencies,omitempty"`
	Complexities      []entities.TaskComplexity `json:"complexities,omitempty"`
	Tags              []string                  `json:"tags,omitempty"`
	Labels            []string                  `json:"labels,omitempty"`
	DueDateFrom       *time.Time                `json:"due_date_from,omitempty"`
	DueDateTo         *time.Time                `json:"due_date_to,omitempty"`
	CreatedAfter      *time.Time                `json:"created_after,omitempty"`
	CreatedBefore     *time.Time                `json:"created_before,omitempty"`
	UpdatedAfter      *time.Time                `json:"updated_after,omitempty"`
	UpdatedBefore     *time.Time                `json:"updated_before,omitempty"`
	IsOverdue         *bool                     `json:"is_overdue,omitempty"`
	IsBlocked         *bool                     `json:"is_blocked,omitempty"`
	IsRecurring       *bool                     `json:"is_recurring,omitempty"`
	IsTemplate        *bool                     `json:"is_template,omitempty"`
	IsArchived        *bool                     `json:"is_archived,omitempty"`
	HasAttachments    *bool                     `json:"has_attachments,omitempty"`
	HasComments       *bool                     `json:"has_comments,omitempty"`
	MinProgress       *float64                  `json:"min_progress,omitempty"`
	MaxProgress       *float64                  `json:"max_progress,omitempty"`
	MinEstimatedHours *float64                  `json:"min_estimated_hours,omitempty"`
	MaxEstimatedHours *float64                  `json:"max_estimated_hours,omitempty"`
	CustomFields      map[string]interface{}    `json:"custom_fields,omitempty"`
	SortBy            string                    `json:"sort_by,omitempty"`
	SortOrder         string                    `json:"sort_order,omitempty"`
	Limit             int                       `json:"limit,omitempty"`
	Offset            int                       `json:"offset,omitempty"`
}

// TaskMetricsFilter defines filtering options for task metrics
type TaskMetricsFilter struct {
	ProjectIDs    []uuid.UUID   `json:"project_ids,omitempty"`
	UserIDs       []uuid.UUID   `json:"user_ids,omitempty"`
	TeamIDs       []uuid.UUID   `json:"team_ids,omitempty"`
	Period        time.Duration `json:"period"`
	StartDate     time.Time     `json:"start_date"`
	EndDate       time.Time     `json:"end_date"`
	GroupBy       string        `json:"group_by,omitempty"`
	IncludeTrends bool          `json:"include_trends"`
}

// TaskMetrics contains task metrics and analytics
type TaskMetrics struct {
	Period                string                        `json:"period"`
	TotalTasks            int                           `json:"total_tasks"`
	CompletedTasks        int                           `json:"completed_tasks"`
	InProgressTasks       int                           `json:"in_progress_tasks"`
	OverdueTasks          int                           `json:"overdue_tasks"`
	BlockedTasks          int                           `json:"blocked_tasks"`
	CompletionRate        float64                       `json:"completion_rate"`
	AverageCompletionTime float64                       `json:"average_completion_time"`
	TotalEstimatedHours   float64                       `json:"total_estimated_hours"`
	TotalActualHours      float64                       `json:"total_actual_hours"`
	EfficiencyRatio       float64                       `json:"efficiency_ratio"`
	TasksByStatus         map[entities.TaskStatus]int   `json:"tasks_by_status"`
	TasksByPriority       map[entities.TaskPriority]int `json:"tasks_by_priority"`
	TasksByType           map[entities.TaskType]int     `json:"tasks_by_type"`
	TopPerformers         []UserPerformance             `json:"top_performers"`
	TrendData             map[string][]float64          `json:"trend_data,omitempty"`
	GeneratedAt           time.Time                     `json:"generated_at"`
}

// UserProductivity contains user productivity metrics
type UserProductivity struct {
	UserID              uuid.UUID `json:"user_id"`
	Period              string    `json:"period"`
	TasksCompleted      int       `json:"tasks_completed"`
	TasksInProgress     int       `json:"tasks_in_progress"`
	TasksOverdue        int       `json:"tasks_overdue"`
	TotalHoursWorked    float64   `json:"total_hours_worked"`
	AverageTaskTime     float64   `json:"average_task_time"`
	ProductivityScore   float64   `json:"productivity_score"`
	EfficiencyRatio     float64   `json:"efficiency_ratio"`
	QualityScore        float64   `json:"quality_score"`
	OnTimeDeliveryRate  float64   `json:"on_time_delivery_rate"`
	TaskCompletionTrend []float64 `json:"task_completion_trend"`
	GeneratedAt         time.Time `json:"generated_at"`
}

// ProjectProgress contains project progress metrics
type ProjectProgress struct {
	ProjectID           uuid.UUID           `json:"project_id"`
	TotalTasks          int                 `json:"total_tasks"`
	CompletedTasks      int                 `json:"completed_tasks"`
	InProgressTasks     int                 `json:"in_progress_tasks"`
	OverdueTasks        int                 `json:"overdue_tasks"`
	ProgressPercentage  float64             `json:"progress_percentage"`
	EstimatedCompletion *time.Time          `json:"estimated_completion,omitempty"`
	IsOnTrack           bool                `json:"is_on_track"`
	RiskLevel           string              `json:"risk_level"`
	MilestoneProgress   []MilestoneProgress `json:"milestone_progress"`
	GeneratedAt         time.Time           `json:"generated_at"`
}

// MilestoneProgress contains milestone progress information
type MilestoneProgress struct {
	MilestoneID    uuid.UUID `json:"milestone_id"`
	Name           string    `json:"name"`
	DueDate        time.Time `json:"due_date"`
	Progress       float64   `json:"progress"`
	IsCompleted    bool      `json:"is_completed"`
	IsOverdue      bool      `json:"is_overdue"`
	TasksTotal     int       `json:"tasks_total"`
	TasksCompleted int       `json:"tasks_completed"`
}

// UserPerformance contains user performance metrics
type UserPerformance struct {
	UserID            uuid.UUID `json:"user_id"`
	UserName          string    `json:"user_name"`
	TasksCompleted    int       `json:"tasks_completed"`
	AverageRating     float64   `json:"average_rating"`
	OnTimeDelivery    float64   `json:"on_time_delivery"`
	ProductivityScore float64   `json:"productivity_score"`
}

// ProjectRepository defines the interface for project data access
type ProjectRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, project *entities.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Project, error)
	GetByCode(ctx context.Context, code string) (*entities.Project, error)
	Update(ctx context.Context, project *entities.Project) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Listing and filtering
	List(ctx context.Context, filter *ProjectFilter) ([]*entities.Project, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, filter *ProjectFilter) ([]*entities.Project, error)
	ListByTeam(ctx context.Context, teamID uuid.UUID, filter *ProjectFilter) ([]*entities.Project, error)
	ListByStatus(ctx context.Context, status entities.ProjectStatus, filter *ProjectFilter) ([]*entities.Project, error)

	// Members
	AddMember(ctx context.Context, member *entities.ProjectMember) error
	UpdateMember(ctx context.Context, member *entities.ProjectMember) error
	RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error
	GetMembers(ctx context.Context, projectID uuid.UUID) ([]*entities.ProjectMember, error)

	// Milestones
	AddMilestone(ctx context.Context, milestone *entities.Milestone) error
	UpdateMilestone(ctx context.Context, milestone *entities.Milestone) error
	DeleteMilestone(ctx context.Context, milestoneID uuid.UUID) error
	GetMilestones(ctx context.Context, projectID uuid.UUID) ([]*entities.Milestone, error)

	// Analytics
	GetProjectMetrics(ctx context.Context, projectID uuid.UUID, period time.Duration) (*ProjectMetrics, error)
	GetPortfolioMetrics(ctx context.Context, filter *ProjectFilter) (*PortfolioMetrics, error)

	// Search
	Search(ctx context.Context, query string, filter *ProjectFilter) ([]*entities.Project, error)

	// Bulk operations
	BulkUpdate(ctx context.Context, projects []*entities.Project) error

	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context, repo ProjectRepository) error) error
}

// ProjectFilter defines filtering options for project queries
type ProjectFilter struct {
	OwnerIDs      []uuid.UUID                `json:"owner_ids,omitempty"`
	ManagerIDs    []uuid.UUID                `json:"manager_ids,omitempty"`
	TeamIDs       []uuid.UUID                `json:"team_ids,omitempty"`
	Types         []entities.ProjectType     `json:"types,omitempty"`
	Statuses      []entities.ProjectStatus   `json:"statuses,omitempty"`
	Priorities    []entities.ProjectPriority `json:"priorities,omitempty"`
	Categories    []entities.ProjectCategory `json:"categories,omitempty"`
	Tags          []string                   `json:"tags,omitempty"`
	Labels        []string                   `json:"labels,omitempty"`
	StartDateFrom *time.Time                 `json:"start_date_from,omitempty"`
	StartDateTo   *time.Time                 `json:"start_date_to,omitempty"`
	EndDateFrom   *time.Time                 `json:"end_date_from,omitempty"`
	EndDateTo     *time.Time                 `json:"end_date_to,omitempty"`
	CreatedAfter  *time.Time                 `json:"created_after,omitempty"`
	CreatedBefore *time.Time                 `json:"created_before,omitempty"`
	IsTemplate    *bool                      `json:"is_template,omitempty"`
	IsArchived    *bool                      `json:"is_archived,omitempty"`
	MinBudget     *float64                   `json:"min_budget,omitempty"`
	MaxBudget     *float64                   `json:"max_budget,omitempty"`
	CustomFields  map[string]interface{}     `json:"custom_fields,omitempty"`
	SortBy        string                     `json:"sort_by,omitempty"`
	SortOrder     string                     `json:"sort_order,omitempty"`
	Limit         int                        `json:"limit,omitempty"`
	Offset        int                        `json:"offset,omitempty"`
}

// ProjectMetrics contains project-specific metrics
type ProjectMetrics struct {
	ProjectID           uuid.UUID `json:"project_id"`
	Period              string    `json:"period"`
	TotalTasks          int       `json:"total_tasks"`
	CompletedTasks      int       `json:"completed_tasks"`
	ProgressPercentage  float64   `json:"progress_percentage"`
	BudgetUtilization   float64   `json:"budget_utilization"`
	ScheduleVariance    float64   `json:"schedule_variance"`
	CostVariance        float64   `json:"cost_variance"`
	QualityScore        float64   `json:"quality_score"`
	RiskScore           float64   `json:"risk_score"`
	TeamProductivity    float64   `json:"team_productivity"`
	MilestoneCompletion float64   `json:"milestone_completion"`
	GeneratedAt         time.Time `json:"generated_at"`
}

// PortfolioMetrics contains portfolio-level metrics
type PortfolioMetrics struct {
	TotalProjects       int                              `json:"total_projects"`
	ActiveProjects      int                              `json:"active_projects"`
	CompletedProjects   int                              `json:"completed_projects"`
	OverdueProjects     int                              `json:"overdue_projects"`
	TotalBudget         entities.Money                   `json:"total_budget"`
	SpentBudget         entities.Money                   `json:"spent_budget"`
	ProjectsByStatus    map[entities.ProjectStatus]int   `json:"projects_by_status"`
	ProjectsByPriority  map[entities.ProjectPriority]int `json:"projects_by_priority"`
	ProjectsByType      map[entities.ProjectType]int     `json:"projects_by_type"`
	AverageCompletion   float64                          `json:"average_completion"`
	ResourceUtilization float64                          `json:"resource_utilization"`
	GeneratedAt         time.Time                        `json:"generated_at"`
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	GetByUsername(ctx context.Context, username string) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Listing and filtering
	List(ctx context.Context, filter *UserFilter) ([]*entities.User, error)
	ListByRole(ctx context.Context, role entities.UserRole, filter *UserFilter) ([]*entities.User, error)
	ListByTeam(ctx context.Context, teamID uuid.UUID, filter *UserFilter) ([]*entities.User, error)
	ListBySkills(ctx context.Context, skills []string, filter *UserFilter) ([]*entities.User, error)

	// Capacity and workload
	GetUserCapacity(ctx context.Context, userID uuid.UUID) (*entities.UserCapacity, error)
	UpdateUserCapacity(ctx context.Context, userID uuid.UUID, capacity *entities.UserCapacity) error
	GetAvailableUsers(ctx context.Context, requiredSkills []string, maxWorkload float64) ([]*entities.User, error)

	// Search
	Search(ctx context.Context, query string, filter *UserFilter) ([]*entities.User, error)

	// Bulk operations
	BulkUpdate(ctx context.Context, users []*entities.User) error

	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context, repo UserRepository) error) error
}

// UserFilter defines filtering options for user queries
type UserFilter struct {
	Roles           []entities.UserRole   `json:"roles,omitempty"`
	Statuses        []entities.UserStatus `json:"statuses,omitempty"`
	Departments     []string              `json:"departments,omitempty"`
	Locations       []string              `json:"locations,omitempty"`
	Skills          []string              `json:"skills,omitempty"`
	TeamIDs         []uuid.UUID           `json:"team_ids,omitempty"`
	IsActive        *bool                 `json:"is_active,omitempty"`
	CreatedAfter    *time.Time            `json:"created_after,omitempty"`
	CreatedBefore   *time.Time            `json:"created_before,omitempty"`
	LastActiveAfter *time.Time            `json:"last_active_after,omitempty"`
	MaxWorkload     *float64              `json:"max_workload,omitempty"`
	MinWorkload     *float64              `json:"min_workload,omitempty"`
	SortBy          string                `json:"sort_by,omitempty"`
	SortOrder       string                `json:"sort_order,omitempty"`
	Limit           int                   `json:"limit,omitempty"`
	Offset          int                   `json:"offset,omitempty"`
}

// TeamRepository defines the interface for team data access
type TeamRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, team *entities.Team) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Team, error)
	GetByName(ctx context.Context, name string) (*entities.Team, error)
	Update(ctx context.Context, team *entities.Team) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Listing and filtering
	List(ctx context.Context, filter *TeamFilter) ([]*entities.Team, error)
	ListByType(ctx context.Context, teamType entities.TeamType, filter *TeamFilter) ([]*entities.Team, error)
	ListByLeader(ctx context.Context, leaderID uuid.UUID, filter *TeamFilter) ([]*entities.Team, error)

	// Members
	AddMember(ctx context.Context, membership *entities.TeamMembership) error
	UpdateMember(ctx context.Context, membership *entities.TeamMembership) error
	RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error
	GetMembers(ctx context.Context, teamID uuid.UUID) ([]*entities.TeamMembership, error)
	GetUserTeams(ctx context.Context, userID uuid.UUID) ([]*entities.TeamMembership, error)

	// Hierarchy
	GetSubTeams(ctx context.Context, parentTeamID uuid.UUID) ([]*entities.Team, error)
	GetParentTeam(ctx context.Context, teamID uuid.UUID) (*entities.Team, error)

	// Search
	Search(ctx context.Context, query string, filter *TeamFilter) ([]*entities.Team, error)

	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context, repo TeamRepository) error) error
}

// TeamFilter defines filtering options for team queries
type TeamFilter struct {
	Types         []entities.TeamType   `json:"types,omitempty"`
	Statuses      []entities.TeamStatus `json:"statuses,omitempty"`
	LeaderIDs     []uuid.UUID           `json:"leader_ids,omitempty"`
	ParentTeamIDs []uuid.UUID           `json:"parent_team_ids,omitempty"`
	Locations     []string              `json:"locations,omitempty"`
	Skills        []string              `json:"skills,omitempty"`
	IsActive      *bool                 `json:"is_active,omitempty"`
	CreatedAfter  *time.Time            `json:"created_after,omitempty"`
	CreatedBefore *time.Time            `json:"created_before,omitempty"`
	MinMembers    *int                  `json:"min_members,omitempty"`
	MaxMembers    *int                  `json:"max_members,omitempty"`
	SortBy        string                `json:"sort_by,omitempty"`
	SortOrder     string                `json:"sort_order,omitempty"`
	Limit         int                   `json:"limit,omitempty"`
	Offset        int                   `json:"offset,omitempty"`
}
