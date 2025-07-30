package entities

import (
	"time"

	"github.com/google/uuid"
)

// Project represents a comprehensive project entity
type Project struct {
	ID           uuid.UUID            `json:"id" redis:"id"`
	Name         string               `json:"name" redis:"name"`
	Description  string               `json:"description" redis:"description"`
	Code         string               `json:"code" redis:"code"`
	Type         ProjectType          `json:"type" redis:"type"`
	Status       ProjectStatus        `json:"status" redis:"status"`
	Priority     ProjectPriority      `json:"priority" redis:"priority"`
	Category     ProjectCategory      `json:"category" redis:"category"`
	OwnerID      uuid.UUID            `json:"owner_id" redis:"owner_id"`
	Owner        *User                `json:"owner,omitempty"`
	ManagerID    *uuid.UUID           `json:"manager_id,omitempty" redis:"manager_id"`
	Manager      *User                `json:"manager,omitempty"`
	TeamID       *uuid.UUID           `json:"team_id,omitempty" redis:"team_id"`
	Team         *Team                `json:"team,omitempty"`
	Members      []*ProjectMember     `json:"members,omitempty"`
	Tasks        []*Task              `json:"tasks,omitempty"`
	Milestones   []*Milestone         `json:"milestones,omitempty"`
	Budget       *ProjectBudget       `json:"budget,omitempty"`
	Timeline     *ProjectTimeline     `json:"timeline,omitempty"`
	Resources    []*ProjectResource   `json:"resources,omitempty"`
	Deliverables []*Deliverable       `json:"deliverables,omitempty"`
	Risks        []*ProjectRisk       `json:"risks,omitempty"`
	Dependencies []*ProjectDependency `json:"dependencies,omitempty"`
	Tags         []string             `json:"tags" redis:"tags"`
	Labels       []string             `json:"labels" redis:"labels"`
	CustomFields map[string]any       `json:"custom_fields" redis:"custom_fields"`
	Metadata     map[string]any       `json:"metadata" redis:"metadata"`
	ExternalIDs  map[string]string    `json:"external_ids" redis:"external_ids"`
	IsTemplate   bool                 `json:"is_template" redis:"is_template"`
	TemplateID   *uuid.UUID           `json:"template_id,omitempty" redis:"template_id"`
	IsArchived   bool                 `json:"is_archived" redis:"is_archived"`
	CreatedAt    time.Time            `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at" redis:"updated_at"`
	CreatedBy    uuid.UUID            `json:"created_by" redis:"created_by"`
	UpdatedBy    uuid.UUID            `json:"updated_by" redis:"updated_by"`
	Version      int64                `json:"version" redis:"version"`
}

// ProjectType defines the type of project
type ProjectType string

const (
	ProjectTypeOperational ProjectType = "operational"
	ProjectTypeStrategic   ProjectType = "strategic"
	ProjectTypeMaintenance ProjectType = "maintenance"
	ProjectTypeImprovement ProjectType = "improvement"
	ProjectTypeCompliance  ProjectType = "compliance"
	ProjectTypeResearch    ProjectType = "research"
	ProjectTypeDevelopment ProjectType = "development"
	ProjectTypeMarketing   ProjectType = "marketing"
	ProjectTypeTraining    ProjectType = "training"
	ProjectTypeEvent       ProjectType = "event"
)

// ProjectStatus defines the status of a project
type ProjectStatus string

const (
	ProjectStatusPlanning  ProjectStatus = "planning"
	ProjectStatusActive    ProjectStatus = "active"
	ProjectStatusOnHold    ProjectStatus = "on_hold"
	ProjectStatusCompleted ProjectStatus = "completed"
	ProjectStatusCancelled ProjectStatus = "cancelled"
	ProjectStatusArchived  ProjectStatus = "archived"
	ProjectStatusTemplate  ProjectStatus = "template"
)

// ProjectPriority defines the priority of a project
type ProjectPriority string

const (
	ProjectPriorityLow      ProjectPriority = "low"
	ProjectPriorityMedium   ProjectPriority = "medium"
	ProjectPriorityHigh     ProjectPriority = "high"
	ProjectPriorityCritical ProjectPriority = "critical"
)

// ProjectCategory defines the category of a project
type ProjectCategory string

const (
	CategoryOperations ProjectCategory = "operations"
	CategoryIT         ProjectCategory = "it"
	CategoryHR         ProjectCategory = "hr"
	CategoryFinance    ProjectCategory = "finance"
	CategoryMarketing  ProjectCategory = "marketing"
	CategorySales      ProjectCategory = "sales"
	CategoryCustomer   ProjectCategory = "customer"
	CategoryProduct    ProjectCategory = "product"
	CategoryQuality    ProjectCategory = "quality"
	CategorySafety     ProjectCategory = "safety"
)

// ProjectMember represents a member of a project
type ProjectMember struct {
	ID          uuid.UUID  `json:"id" redis:"id"`
	ProjectID   uuid.UUID  `json:"project_id" redis:"project_id"`
	UserID      uuid.UUID  `json:"user_id" redis:"user_id"`
	User        *User      `json:"user,omitempty"`
	Role        MemberRole `json:"role" redis:"role"`
	Permissions []string   `json:"permissions" redis:"permissions"`
	Allocation  float64    `json:"allocation" redis:"allocation"` // Percentage of time allocated
	JoinedAt    time.Time  `json:"joined_at" redis:"joined_at"`
	LeftAt      *time.Time `json:"left_at,omitempty" redis:"left_at"`
	IsActive    bool       `json:"is_active" redis:"is_active"`
	Notes       string     `json:"notes" redis:"notes"`
}

// MemberRole defines the role of a project member
type MemberRole string

const (
	RoleOwner          MemberRole = "owner"
	RoleProjectManager MemberRole = "manager"
	RoleLead           MemberRole = "lead"
	RoleMember         MemberRole = "member"
	RoleContributor    MemberRole = "contributor"
	RoleObserver       MemberRole = "observer"
	RoleStakeholder    MemberRole = "stakeholder"
)

// Milestone represents a project milestone
type Milestone struct {
	ID           uuid.UUID       `json:"id" redis:"id"`
	ProjectID    uuid.UUID       `json:"project_id" redis:"project_id"`
	Name         string          `json:"name" redis:"name"`
	Description  string          `json:"description" redis:"description"`
	Type         MilestoneType   `json:"type" redis:"type"`
	Status       MilestoneStatus `json:"status" redis:"status"`
	DueDate      time.Time       `json:"due_date" redis:"due_date"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty" redis:"completed_at"`
	Tasks        []*Task         `json:"tasks,omitempty"`
	Deliverables []*Deliverable  `json:"deliverables,omitempty"`
	Order        int             `json:"order" redis:"order"`
	IsActive     bool            `json:"is_active" redis:"is_active"`
	CreatedAt    time.Time       `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" redis:"updated_at"`
	CreatedBy    uuid.UUID       `json:"created_by" redis:"created_by"`
}

// MilestoneType defines the type of milestone
type MilestoneType string

const (
	MilestoneTypePhase       MilestoneType = "phase"
	MilestoneTypeDeliverable MilestoneType = "deliverable"
	MilestoneTypeReview      MilestoneType = "review"
	MilestoneTypeApproval    MilestoneType = "approval"
	MilestoneTypeRelease     MilestoneType = "release"
	MilestoneTypeDeadline    MilestoneType = "deadline"
)

// MilestoneStatus defines the status of a milestone
type MilestoneStatus string

const (
	MilestoneStatusPending    MilestoneStatus = "pending"
	MilestoneStatusInProgress MilestoneStatus = "in_progress"
	MilestoneStatusCompleted  MilestoneStatus = "completed"
	MilestoneStatusOverdue    MilestoneStatus = "overdue"
	MilestoneStatusCancelled  MilestoneStatus = "cancelled"
)

// ProjectBudget represents project budget information
type ProjectBudget struct {
	TotalBudget     Money         `json:"total_budget" redis:"total_budget"`
	SpentAmount     Money         `json:"spent_amount" redis:"spent_amount"`
	RemainingBudget Money         `json:"remaining_budget" redis:"remaining_budget"`
	Currency        string        `json:"currency" redis:"currency"`
	BudgetItems     []*BudgetItem `json:"budget_items,omitempty"`
	LastUpdated     time.Time     `json:"last_updated" redis:"last_updated"`
}

// BudgetItem represents a budget line item
type BudgetItem struct {
	ID          uuid.UUID `json:"id" redis:"id"`
	Category    string    `json:"category" redis:"category"`
	Description string    `json:"description" redis:"description"`
	Budgeted    Money     `json:"budgeted" redis:"budgeted"`
	Actual      Money     `json:"actual" redis:"actual"`
	Variance    Money     `json:"variance" redis:"variance"`
}

// Money represents monetary values
type Money struct {
	Amount   float64 `json:"amount" redis:"amount"`
	Currency string  `json:"currency" redis:"currency"`
}

// ProjectTimeline represents project timeline information
type ProjectTimeline struct {
	StartDate    time.Time  `json:"start_date" redis:"start_date"`
	EndDate      time.Time  `json:"end_date" redis:"end_date"`
	ActualStart  *time.Time `json:"actual_start,omitempty" redis:"actual_start"`
	ActualEnd    *time.Time `json:"actual_end,omitempty" redis:"actual_end"`
	Duration     int        `json:"duration" redis:"duration"` // Days
	WorkingDays  int        `json:"working_days" redis:"working_days"`
	ProgressDays int        `json:"progress_days" redis:"progress_days"`
	IsOnTrack    bool       `json:"is_on_track" redis:"is_on_track"`
}

// ProjectResource represents a resource allocated to a project
type ProjectResource struct {
	ID          uuid.UUID    `json:"id" redis:"id"`
	ProjectID   uuid.UUID    `json:"project_id" redis:"project_id"`
	Type        ResourceType `json:"type" redis:"type"`
	Name        string       `json:"name" redis:"name"`
	Description string       `json:"description" redis:"description"`
	Quantity    float64      `json:"quantity" redis:"quantity"`
	Unit        string       `json:"unit" redis:"unit"`
	Cost        Money        `json:"cost" redis:"cost"`
	AllocatedAt time.Time    `json:"allocated_at" redis:"allocated_at"`
	ReleasedAt  *time.Time   `json:"released_at,omitempty" redis:"released_at"`
	IsActive    bool         `json:"is_active" redis:"is_active"`
}

// ResourceType defines the type of resource
type ResourceType string

const (
	ResourceTypeHuman     ResourceType = "human"
	ResourceTypeEquipment ResourceType = "equipment"
	ResourceTypeMaterial  ResourceType = "material"
	ResourceTypeSoftware  ResourceType = "software"
	ResourceTypeFacility  ResourceType = "facility"
	ResourceTypeBudget    ResourceType = "budget"
)

// Deliverable represents a project deliverable
type Deliverable struct {
	ID          uuid.UUID         `json:"id" redis:"id"`
	ProjectID   uuid.UUID         `json:"project_id" redis:"project_id"`
	MilestoneID *uuid.UUID        `json:"milestone_id,omitempty" redis:"milestone_id"`
	Name        string            `json:"name" redis:"name"`
	Description string            `json:"description" redis:"description"`
	Type        DeliverableType   `json:"type" redis:"type"`
	Status      DeliverableStatus `json:"status" redis:"status"`
	DueDate     time.Time         `json:"due_date" redis:"due_date"`
	CompletedAt *time.Time        `json:"completed_at,omitempty" redis:"completed_at"`
	OwnerID     uuid.UUID         `json:"owner_id" redis:"owner_id"`
	Owner       *User             `json:"owner,omitempty"`
	Attachments []*TaskAttachment `json:"attachments,omitempty"`
	IsActive    bool              `json:"is_active" redis:"is_active"`
	CreatedAt   time.Time         `json:"created_at" redis:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" redis:"updated_at"`
}

// DeliverableType defines the type of deliverable
type DeliverableType string

const (
	DeliverableTypeDocument DeliverableType = "document"
	DeliverableTypeReport   DeliverableType = "report"
	DeliverableTypeProduct  DeliverableType = "product"
	DeliverableTypeService  DeliverableType = "service"
	DeliverableTypeTraining DeliverableType = "training"
	DeliverableTypeProcess  DeliverableType = "process"
)

// DeliverableStatus defines the status of a deliverable
type DeliverableStatus string

const (
	DeliverableStatusPending    DeliverableStatus = "pending"
	DeliverableStatusInProgress DeliverableStatus = "in_progress"
	DeliverableStatusReview     DeliverableStatus = "review"
	DeliverableStatusCompleted  DeliverableStatus = "completed"
	DeliverableStatusRejected   DeliverableStatus = "rejected"
)

// ProjectRisk represents a project risk
type ProjectRisk struct {
	ID           uuid.UUID    `json:"id" redis:"id"`
	ProjectID    uuid.UUID    `json:"project_id" redis:"project_id"`
	Title        string       `json:"title" redis:"title"`
	Description  string       `json:"description" redis:"description"`
	Category     RiskCategory `json:"category" redis:"category"`
	Probability  RiskLevel    `json:"probability" redis:"probability"`
	Impact       RiskLevel    `json:"impact" redis:"impact"`
	RiskScore    float64      `json:"risk_score" redis:"risk_score"`
	Status       RiskStatus   `json:"status" redis:"status"`
	OwnerID      uuid.UUID    `json:"owner_id" redis:"owner_id"`
	Owner        *User        `json:"owner,omitempty"`
	Mitigation   string       `json:"mitigation" redis:"mitigation"`
	Contingency  string       `json:"contingency" redis:"contingency"`
	IdentifiedAt time.Time    `json:"identified_at" redis:"identified_at"`
	ReviewDate   *time.Time   `json:"review_date,omitempty" redis:"review_date"`
	ClosedAt     *time.Time   `json:"closed_at,omitempty" redis:"closed_at"`
	IsActive     bool         `json:"is_active" redis:"is_active"`
}

// RiskCategory defines the category of risk
type RiskCategory string

const (
	RiskCategoryTechnical   RiskCategory = "technical"
	RiskCategoryOperational RiskCategory = "operational"
	RiskCategoryFinancial   RiskCategory = "financial"
	RiskCategorySchedule    RiskCategory = "schedule"
	RiskCategoryResource    RiskCategory = "resource"
	RiskCategoryQuality     RiskCategory = "quality"
	RiskCategoryCompliance  RiskCategory = "compliance"
	RiskCategoryExternal    RiskCategory = "external"
)

// RiskLevel defines the level of risk
type RiskLevel string

const (
	RiskLevelVeryLow  RiskLevel = "very_low"
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelVeryHigh RiskLevel = "very_high"
)

// RiskStatus defines the status of a risk
type RiskStatus string

const (
	RiskStatusOpen      RiskStatus = "open"
	RiskStatusMitigated RiskStatus = "mitigated"
	RiskStatusAccepted  RiskStatus = "accepted"
	RiskStatusClosed    RiskStatus = "closed"
)

// ProjectDependency represents a dependency between projects
type ProjectDependency struct {
	ID               uuid.UUID      `json:"id" redis:"id"`
	ProjectID        uuid.UUID      `json:"project_id" redis:"project_id"`
	DependsOnID      uuid.UUID      `json:"depends_on_id" redis:"depends_on_id"`
	DependsOnProject *Project       `json:"depends_on_project,omitempty"`
	Type             DependencyType `json:"type" redis:"type"`
	Description      string         `json:"description" redis:"description"`
	IsActive         bool           `json:"is_active" redis:"is_active"`
	CreatedAt        time.Time      `json:"created_at" redis:"created_at"`
	CreatedBy        uuid.UUID      `json:"created_by" redis:"created_by"`
}

// NewProject creates a new project with default values
func NewProject(name, description string, projectType ProjectType, ownerID, createdBy uuid.UUID) *Project {
	now := time.Now()
	return &Project{
		ID:           uuid.New(),
		Name:         name,
		Description:  description,
		Code:         generateProjectCode(),
		Type:         projectType,
		Status:       ProjectStatusPlanning,
		Priority:     ProjectPriorityMedium,
		OwnerID:      ownerID,
		Tags:         []string{},
		Labels:       []string{},
		CustomFields: make(map[string]any),
		Metadata:     make(map[string]any),
		ExternalIDs:  make(map[string]string),
		IsTemplate:   false,
		IsArchived:   false,
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    createdBy,
		UpdatedBy:    createdBy,
		Version:      1,
	}
}

// generateProjectCode generates a unique project code
func generateProjectCode() string {
	return "PROJ-" + time.Now().Format("20060102-150405")
}

// UpdateStatus updates the project status
func (p *Project) UpdateStatus(newStatus ProjectStatus, updatedBy uuid.UUID) {
	p.Status = newStatus
	p.UpdatedBy = updatedBy
	p.UpdatedAt = time.Now()
	p.Version++

	// Handle status-specific logic
	switch newStatus {
	case ProjectStatusActive:
		if p.Timeline != nil && p.Timeline.ActualStart == nil {
			now := time.Now()
			p.Timeline.ActualStart = &now
		}
	case ProjectStatusCompleted:
		if p.Timeline != nil {
			now := time.Now()
			p.Timeline.ActualEnd = &now
		}
	}
}

// AddMember adds a member to the project
func (p *Project) AddMember(member *ProjectMember) {
	p.Members = append(p.Members, member)
	p.UpdatedAt = time.Now()
	p.Version++
}

// RemoveMember removes a member from the project
func (p *Project) RemoveMember(userID uuid.UUID) {
	for i, member := range p.Members {
		if member.UserID == userID {
			member.IsActive = false
			now := time.Now()
			member.LeftAt = &now
			p.Members[i] = member
			break
		}
	}
	p.UpdatedAt = time.Now()
	p.Version++
}

// AddTask adds a task to the project
func (p *Project) AddTask(task *Task) {
	task.ProjectID = &p.ID
	p.Tasks = append(p.Tasks, task)
	p.UpdatedAt = time.Now()
	p.Version++
}

// AddMilestone adds a milestone to the project
func (p *Project) AddMilestone(milestone *Milestone) {
	milestone.ProjectID = p.ID
	p.Milestones = append(p.Milestones, milestone)
	p.UpdatedAt = time.Now()
	p.Version++
}

// CalculateProgress calculates the overall project progress
func (p *Project) CalculateProgress() float64 {
	if len(p.Tasks) == 0 {
		return 0
	}

	totalProgress := 0.0
	for _, task := range p.Tasks {
		totalProgress += task.ProgressPercent
	}

	return totalProgress / float64(len(p.Tasks))
}

// GetActiveTasks returns all active tasks in the project
func (p *Project) GetActiveTasks() []*Task {
	var activeTasks []*Task
	for _, task := range p.Tasks {
		if !task.IsArchived && task.Status != StatusCancelled {
			activeTasks = append(activeTasks, task)
		}
	}
	return activeTasks
}

// GetActiveMembers returns all active members of the project
func (p *Project) GetActiveMembers() []*ProjectMember {
	var activeMembers []*ProjectMember
	for _, member := range p.Members {
		if member.IsActive {
			activeMembers = append(activeMembers, member)
		}
	}
	return activeMembers
}

// IsOverdue checks if the project is overdue
func (p *Project) IsOverdue() bool {
	if p.Timeline == nil || p.Status == ProjectStatusCompleted || p.Status == ProjectStatusCancelled {
		return false
	}
	return time.Now().After(p.Timeline.EndDate)
}

// GetBudgetUtilization calculates budget utilization percentage
func (p *Project) GetBudgetUtilization() float64 {
	if p.Budget == nil || p.Budget.TotalBudget.Amount == 0 {
		return 0
	}
	return (p.Budget.SpentAmount.Amount / p.Budget.TotalBudget.Amount) * 100
}

// Archive archives the project
func (p *Project) Archive(archivedBy uuid.UUID) {
	p.IsArchived = true
	p.Status = ProjectStatusArchived
	p.UpdatedBy = archivedBy
	p.UpdatedAt = time.Now()
	p.Version++
}

// Unarchive unarchives the project
func (p *Project) Unarchive(unarchivedBy uuid.UUID) {
	p.IsArchived = false
	p.Status = ProjectStatusPlanning
	p.UpdatedBy = unarchivedBy
	p.UpdatedAt = time.Now()
	p.Version++
}
