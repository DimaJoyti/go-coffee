package entities

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the task management system
type User struct {
	ID                uuid.UUID              `json:"id" redis:"id"`
	Username          string                 `json:"username" redis:"username"`
	Email             string                 `json:"email" redis:"email"`
	FirstName         string                 `json:"first_name" redis:"first_name"`
	LastName          string                 `json:"last_name" redis:"last_name"`
	DisplayName       string                 `json:"display_name" redis:"display_name"`
	Avatar            string                 `json:"avatar" redis:"avatar"`
	Title             string                 `json:"title" redis:"title"`
	Department        string                 `json:"department" redis:"department"`
	Location          string                 `json:"location" redis:"location"`
	Phone             string                 `json:"phone" redis:"phone"`
	TimeZone          string                 `json:"time_zone" redis:"time_zone"`
	Language          string                 `json:"language" redis:"language"`
	Role              UserRole               `json:"role" redis:"role"`
	Status            UserStatus             `json:"status" redis:"status"`
	Skills            []string               `json:"skills" redis:"skills"`
	Certifications    []string               `json:"certifications" redis:"certifications"`
	Teams             []*TeamMembership      `json:"teams,omitempty"`
	WorkingHours      *WorkingHours          `json:"working_hours,omitempty"`
	Capacity          *UserCapacity          `json:"capacity,omitempty"`
	Preferences       *UserPreferences       `json:"preferences,omitempty"`
	Notifications     *NotificationSettings  `json:"notifications,omitempty"`
	CustomFields      map[string]interface{} `json:"custom_fields" redis:"custom_fields"`
	Metadata          map[string]interface{} `json:"metadata" redis:"metadata"`
	ExternalIDs       map[string]string      `json:"external_ids" redis:"external_ids"`
	LastLoginAt       *time.Time             `json:"last_login_at,omitempty" redis:"last_login_at"`
	LastActiveAt      *time.Time             `json:"last_active_at,omitempty" redis:"last_active_at"`
	IsActive          bool                   `json:"is_active" redis:"is_active"`
	CreatedAt         time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" redis:"updated_at"`
	CreatedBy         uuid.UUID              `json:"created_by" redis:"created_by"`
	UpdatedBy         uuid.UUID              `json:"updated_by" redis:"updated_by"`
	Version           int64                  `json:"version" redis:"version"`
}

// UserRole defines the role of a user
type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleManager   UserRole = "manager"
	RoleTeamLead  UserRole = "team_lead"
	RoleEmployee  UserRole = "employee"
	RoleContractor UserRole = "contractor"
	RoleGuest     UserRole = "guest"
)

// UserStatus defines the status of a user
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusPending   UserStatus = "pending"
)

// WorkingHours represents a user's working hours
type WorkingHours struct {
	Monday    *DaySchedule `json:"monday,omitempty"`
	Tuesday   *DaySchedule `json:"tuesday,omitempty"`
	Wednesday *DaySchedule `json:"wednesday,omitempty"`
	Thursday  *DaySchedule `json:"thursday,omitempty"`
	Friday    *DaySchedule `json:"friday,omitempty"`
	Saturday  *DaySchedule `json:"saturday,omitempty"`
	Sunday    *DaySchedule `json:"sunday,omitempty"`
	TimeZone  string       `json:"time_zone" redis:"time_zone"`
}

// DaySchedule represents working hours for a specific day
type DaySchedule struct {
	IsWorkingDay bool     `json:"is_working_day" redis:"is_working_day"`
	StartTime    string   `json:"start_time" redis:"start_time"`
	EndTime      string   `json:"end_time" redis:"end_time"`
	Breaks       []*Break `json:"breaks,omitempty"`
}

// Break represents a break period
type Break struct {
	StartTime string `json:"start_time" redis:"start_time"`
	EndTime   string `json:"end_time" redis:"end_time"`
	Type      string `json:"type" redis:"type"`
}

// UserCapacity represents a user's capacity and workload
type UserCapacity struct {
	MaxHoursPerDay   float64 `json:"max_hours_per_day" redis:"max_hours_per_day"`
	MaxHoursPerWeek  float64 `json:"max_hours_per_week" redis:"max_hours_per_week"`
	CurrentWorkload  float64 `json:"current_workload" redis:"current_workload"`
	AvailableHours   float64 `json:"available_hours" redis:"available_hours"`
	UtilizationRate  float64 `json:"utilization_rate" redis:"utilization_rate"`
	OverloadThreshold float64 `json:"overload_threshold" redis:"overload_threshold"`
	LastUpdated      time.Time `json:"last_updated" redis:"last_updated"`
}

// UserPreferences represents user preferences
type UserPreferences struct {
	Theme               string            `json:"theme" redis:"theme"`
	Language            string            `json:"language" redis:"language"`
	DateFormat          string            `json:"date_format" redis:"date_format"`
	TimeFormat          string            `json:"time_format" redis:"time_format"`
	StartOfWeek         time.Weekday      `json:"start_of_week" redis:"start_of_week"`
	DefaultView         string            `json:"default_view" redis:"default_view"`
	TaskSortOrder       string            `json:"task_sort_order" redis:"task_sort_order"`
	AutoAssignTasks     bool              `json:"auto_assign_tasks" redis:"auto_assign_tasks"`
	ShowCompletedTasks  bool              `json:"show_completed_tasks" redis:"show_completed_tasks"`
	EmailDigestFrequency string           `json:"email_digest_frequency" redis:"email_digest_frequency"`
	CustomSettings      map[string]interface{} `json:"custom_settings" redis:"custom_settings"`
}

// NotificationSettings represents notification preferences
type NotificationSettings struct {
	EmailNotifications    bool              `json:"email_notifications" redis:"email_notifications"`
	PushNotifications     bool              `json:"push_notifications" redis:"push_notifications"`
	SlackNotifications    bool              `json:"slack_notifications" redis:"slack_notifications"`
	TaskAssigned          bool              `json:"task_assigned" redis:"task_assigned"`
	TaskDue               bool              `json:"task_due" redis:"task_due"`
	TaskCompleted         bool              `json:"task_completed" redis:"task_completed"`
	TaskOverdue           bool              `json:"task_overdue" redis:"task_overdue"`
	ProjectUpdates        bool              `json:"project_updates" redis:"project_updates"`
	MentionedInComments   bool              `json:"mentioned_in_comments" redis:"mentioned_in_comments"`
	WeeklyDigest          bool              `json:"weekly_digest" redis:"weekly_digest"`
	QuietHoursStart       string            `json:"quiet_hours_start" redis:"quiet_hours_start"`
	QuietHoursEnd         string            `json:"quiet_hours_end" redis:"quiet_hours_end"`
	CustomNotifications   map[string]bool   `json:"custom_notifications" redis:"custom_notifications"`
}

// Team represents a team in the organization
type Team struct {
	ID           uuid.UUID              `json:"id" redis:"id"`
	Name         string                 `json:"name" redis:"name"`
	Description  string                 `json:"description" redis:"description"`
	Type         TeamType               `json:"type" redis:"type"`
	Status       TeamStatus             `json:"status" redis:"status"`
	LeaderID     *uuid.UUID             `json:"leader_id,omitempty" redis:"leader_id"`
	Leader       *User                  `json:"leader,omitempty"`
	ParentTeamID *uuid.UUID             `json:"parent_team_id,omitempty" redis:"parent_team_id"`
	ParentTeam   *Team                  `json:"parent_team,omitempty"`
	SubTeams     []*Team                `json:"sub_teams,omitempty"`
	Members      []*TeamMembership      `json:"members,omitempty"`
	Projects     []*Project             `json:"projects,omitempty"`
	Skills       []string               `json:"skills" redis:"skills"`
	Location     string                 `json:"location" redis:"location"`
	CustomFields map[string]interface{} `json:"custom_fields" redis:"custom_fields"`
	Metadata     map[string]interface{} `json:"metadata" redis:"metadata"`
	IsActive     bool                   `json:"is_active" redis:"is_active"`
	CreatedAt    time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" redis:"updated_at"`
	CreatedBy    uuid.UUID              `json:"created_by" redis:"created_by"`
	UpdatedBy    uuid.UUID              `json:"updated_by" redis:"updated_by"`
	Version      int64                  `json:"version" redis:"version"`
}

// TeamType defines the type of team
type TeamType string

const (
	TeamTypeDepartment TeamType = "department"
	TeamTypeProject    TeamType = "project"
	TeamTypeFunctional TeamType = "functional"
	TeamTypeCrossFunc  TeamType = "cross_functional"
	TeamTypeTemporary  TeamType = "temporary"
)

// TeamStatus defines the status of a team
type TeamStatus string

const (
	TeamStatusActive   TeamStatus = "active"
	TeamStatusInactive TeamStatus = "inactive"
	TeamStatusDisbanded TeamStatus = "disbanded"
)

// TeamMembership represents a user's membership in a team
type TeamMembership struct {
	ID           uuid.UUID      `json:"id" redis:"id"`
	TeamID       uuid.UUID      `json:"team_id" redis:"team_id"`
	Team         *Team          `json:"team,omitempty"`
	UserID       uuid.UUID      `json:"user_id" redis:"user_id"`
	User         *User          `json:"user,omitempty"`
	Role         TeamMemberRole `json:"role" redis:"role"`
	Permissions  []string       `json:"permissions" redis:"permissions"`
	JoinedAt     time.Time      `json:"joined_at" redis:"joined_at"`
	LeftAt       *time.Time     `json:"left_at,omitempty" redis:"left_at"`
	IsActive     bool           `json:"is_active" redis:"is_active"`
	Notes        string         `json:"notes" redis:"notes"`
}

// TeamMemberRole defines the role of a team member
type TeamMemberRole string

const (
	TeamRoleLeader     TeamMemberRole = "leader"
	TeamRoleAssistant  TeamMemberRole = "assistant"
	TeamRoleMember     TeamMemberRole = "member"
	TeamRoleSpecialist TeamMemberRole = "specialist"
	TeamRoleIntern     TeamMemberRole = "intern"
)

// NewUser creates a new user with default values
func NewUser(username, email, firstName, lastName string, role UserRole, createdBy uuid.UUID) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		DisplayName:  firstName + " " + lastName,
		Role:         role,
		Status:       UserStatusActive,
		Skills:       []string{},
		Certifications: []string{},
		CustomFields: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
		ExternalIDs:  make(map[string]string),
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    createdBy,
		UpdatedBy:    createdBy,
		Version:      1,
	}
}

// NewTeam creates a new team with default values
func NewTeam(name, description string, teamType TeamType, createdBy uuid.UUID) *Team {
	now := time.Now()
	return &Team{
		ID:           uuid.New(),
		Name:         name,
		Description:  description,
		Type:         teamType,
		Status:       TeamStatusActive,
		Skills:       []string{},
		CustomFields: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    createdBy,
		UpdatedBy:    createdBy,
		Version:      1,
	}
}

// UpdateStatus updates the user status
func (u *User) UpdateStatus(newStatus UserStatus, updatedBy uuid.UUID) {
	u.Status = newStatus
	u.UpdatedBy = updatedBy
	u.UpdatedAt = time.Now()
	u.Version++
}

// UpdateLastActive updates the user's last active timestamp
func (u *User) UpdateLastActive() {
	now := time.Now()
	u.LastActiveAt = &now
	u.UpdatedAt = now
	u.Version++
}

// AddSkill adds a skill to the user
func (u *User) AddSkill(skill string) {
	for _, existingSkill := range u.Skills {
		if existingSkill == skill {
			return // Skill already exists
		}
	}
	u.Skills = append(u.Skills, skill)
	u.UpdatedAt = time.Now()
	u.Version++
}

// RemoveSkill removes a skill from the user
func (u *User) RemoveSkill(skill string) {
	for i, existingSkill := range u.Skills {
		if existingSkill == skill {
			u.Skills = append(u.Skills[:i], u.Skills[i+1:]...)
			u.UpdatedAt = time.Now()
			u.Version++
			return
		}
	}
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// IsAvailable checks if the user is available for work
func (u *User) IsAvailable() bool {
	return u.IsActive && u.Status == UserStatusActive
}

// GetCurrentWorkload returns the user's current workload percentage
func (u *User) GetCurrentWorkload() float64 {
	if u.Capacity == nil {
		return 0
	}
	return u.Capacity.UtilizationRate
}

// IsOverloaded checks if the user is overloaded
func (u *User) IsOverloaded() bool {
	if u.Capacity == nil {
		return false
	}
	return u.Capacity.UtilizationRate > u.Capacity.OverloadThreshold
}

// AddToTeam adds the user to a team
func (t *Team) AddMember(membership *TeamMembership) {
	t.Members = append(t.Members, membership)
	t.UpdatedAt = time.Now()
	t.Version++
}

// RemoveMember removes a member from the team
func (t *Team) RemoveMember(userID uuid.UUID) {
	for i, member := range t.Members {
		if member.UserID == userID {
			member.IsActive = false
			now := time.Now()
			member.LeftAt = &now
			t.Members[i] = member
			break
		}
	}
	t.UpdatedAt = time.Now()
	t.Version++
}

// GetActiveMembers returns all active members of the team
func (t *Team) GetActiveMembers() []*TeamMembership {
	var activeMembers []*TeamMembership
	for _, member := range t.Members {
		if member.IsActive {
			activeMembers = append(activeMembers, member)
		}
	}
	return activeMembers
}

// GetMemberCount returns the number of active members
func (t *Team) GetMemberCount() int {
	return len(t.GetActiveMembers())
}

// HasMember checks if a user is a member of the team
func (t *Team) HasMember(userID uuid.UUID) bool {
	for _, member := range t.Members {
		if member.UserID == userID && member.IsActive {
			return true
		}
	}
	return false
}

// GetLeader returns the team leader
func (t *Team) GetLeader() *User {
	if t.LeaderID == nil {
		return nil
	}
	
	for _, member := range t.Members {
		if member.UserID == *t.LeaderID && member.IsActive {
			return member.User
		}
	}
	return nil
}

// SetLeader sets the team leader
func (t *Team) SetLeader(userID uuid.UUID, updatedBy uuid.UUID) {
	t.LeaderID = &userID
	t.UpdatedBy = updatedBy
	t.UpdatedAt = time.Now()
	t.Version++
}

// Deactivate deactivates the team
func (t *Team) Deactivate(updatedBy uuid.UUID) {
	t.IsActive = false
	t.Status = TeamStatusInactive
	t.UpdatedBy = updatedBy
	t.UpdatedAt = time.Now()
	t.Version++
}

// Activate activates the team
func (t *Team) Activate(updatedBy uuid.UUID) {
	t.IsActive = true
	t.Status = TeamStatusActive
	t.UpdatedBy = updatedBy
	t.UpdatedAt = time.Now()
	t.Version++
}
