package entities

import (
	"time"

	"github.com/google/uuid"
)

// Campaign represents a comprehensive social media campaign entity
type Campaign struct {
	ID                uuid.UUID              `json:"id" redis:"id"`
	Name              string                 `json:"name" redis:"name"`
	Description       string                 `json:"description" redis:"description"`
	Type              CampaignType           `json:"type" redis:"type"`
	Status            CampaignStatus         `json:"status" redis:"status"`
	Priority          CampaignPriority       `json:"priority" redis:"priority"`
	Category          CampaignCategory       `json:"category" redis:"category"`
	BrandID           uuid.UUID              `json:"brand_id" redis:"brand_id"`
	Brand             *Brand                 `json:"brand,omitempty"`
	ManagerID         uuid.UUID              `json:"manager_id" redis:"manager_id"`
	Manager           *User                  `json:"manager,omitempty"`
	TeamMembers       []*CampaignMember      `json:"team_members,omitempty"`
	Content           []*Content             `json:"content,omitempty"`
	Platforms         []PlatformType         `json:"platforms" redis:"platforms"`
	TargetAudience    *TargetAudience        `json:"target_audience,omitempty"`
	Budget            *CampaignBudget        `json:"budget,omitempty"`
	Timeline          *CampaignTimeline      `json:"timeline,omitempty"`
	Objectives        []*CampaignObjective   `json:"objectives,omitempty"`
	KPIs              []*CampaignKPI         `json:"kpis,omitempty"`
	Analytics         *CampaignAnalytics     `json:"analytics,omitempty"`
	ABTests           []*ABTest              `json:"ab_tests,omitempty"`
	Hashtags          []string               `json:"hashtags" redis:"hashtags"`
	Keywords          []string               `json:"keywords" redis:"keywords"`
	Tags              []string               `json:"tags" redis:"tags"`
	Tone              ContentTone            `json:"tone" redis:"tone"`
	Language          string                 `json:"language" redis:"language"`
	ApprovalWorkflow  *ApprovalWorkflow      `json:"approval_workflow,omitempty"`
	Compliance        *ComplianceInfo        `json:"compliance,omitempty"`
	CustomFields      map[string]interface{} `json:"custom_fields" redis:"custom_fields"`
	Metadata          map[string]interface{} `json:"metadata" redis:"metadata"`
	ExternalIDs       map[string]string      `json:"external_ids" redis:"external_ids"`
	IsTemplate        bool                   `json:"is_template" redis:"is_template"`
	TemplateID        *uuid.UUID             `json:"template_id,omitempty" redis:"template_id"`
	IsArchived        bool                   `json:"is_archived" redis:"is_archived"`
	CreatedAt         time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" redis:"updated_at"`
	CreatedBy         uuid.UUID              `json:"created_by" redis:"created_by"`
	UpdatedBy         uuid.UUID              `json:"updated_by" redis:"updated_by"`
	Version           int64                  `json:"version" redis:"version"`
}

// CampaignType defines the type of campaign
type CampaignType string

const (
	CampaignTypeMarketing    CampaignType = "marketing"
	CampaignTypePromotion    CampaignType = "promotion"
	CampaignTypeBranding     CampaignType = "branding"
	CampaignTypeProduct      CampaignType = "product"
	CampaignTypeEvent        CampaignType = "event"
	CampaignTypeSeasonal     CampaignType = "seasonal"
	CampaignTypeInfluencer   CampaignType = "influencer"
	CampaignTypeUserGenerated CampaignType = "user_generated"
	CampaignTypeEducational  CampaignType = "educational"
	CampaignTypeAwareness    CampaignType = "awareness"
)

// CampaignStatus defines the status of campaign
type CampaignStatus string

const (
	CampaignStatusDraft     CampaignStatus = "draft"
	CampaignStatusPlanning  CampaignStatus = "planning"
	CampaignStatusReview    CampaignStatus = "review"
	CampaignStatusApproved  CampaignStatus = "approved"
	CampaignStatusActive    CampaignStatus = "active"
	CampaignStatusPaused    CampaignStatus = "paused"
	CampaignStatusCompleted CampaignStatus = "completed"
	CampaignStatusCancelled CampaignStatus = "cancelled"
	CampaignStatusArchived  CampaignStatus = "archived"
)

// CampaignPriority defines the priority of campaign
type CampaignPriority string

const (
	CampaignPriorityLow      CampaignPriority = "low"
	CampaignPriorityMedium   CampaignPriority = "medium"
	CampaignPriorityHigh     CampaignPriority = "high"
	CampaignPriorityCritical CampaignPriority = "critical"
)

// CampaignCategory defines the category of campaign
type CampaignCategory string

const (
	CategoryBrandAwareness   CampaignCategory = "brand_awareness"
	CategoryLeadGeneration   CampaignCategory = "lead_generation"
	CategorySales            CampaignCategory = "sales"
	CategoryEngagement       CampaignCategory = "engagement"
	CategoryRetention        CampaignCategory = "retention"
	CategoryRecruitment      CampaignCategory = "recruitment"
	CategoryCommunity        CampaignCategory = "community"
	CategoryEducation        CampaignCategory = "education"
)

// CampaignMember represents a team member in a campaign
type CampaignMember struct {
	ID           uuid.UUID    `json:"id" redis:"id"`
	CampaignID   uuid.UUID    `json:"campaign_id" redis:"campaign_id"`
	UserID       uuid.UUID    `json:"user_id" redis:"user_id"`
	User         *User        `json:"user,omitempty"`
	Role         MemberRole   `json:"role" redis:"role"`
	Permissions  []string     `json:"permissions" redis:"permissions"`
	JoinedAt     time.Time    `json:"joined_at" redis:"joined_at"`
	LeftAt       *time.Time   `json:"left_at,omitempty" redis:"left_at"`
	IsActive     bool         `json:"is_active" redis:"is_active"`
}

// MemberRole defines the role of a campaign member
type MemberRole string

const (
	RoleManager      MemberRole = "manager"
	RoleCreator      MemberRole = "creator"
	RoleEditor       MemberRole = "editor"
	RoleReviewer     MemberRole = "reviewer"
	RoleApprover     MemberRole = "approver"
	RoleAnalyst      MemberRole = "analyst"
	RoleContributor  MemberRole = "contributor"
	RoleViewer       MemberRole = "viewer"
)

// CampaignBudget represents campaign budget information
type CampaignBudget struct {
	TotalBudget     Money                          `json:"total_budget" redis:"total_budget"`
	SpentAmount     Money                          `json:"spent_amount" redis:"spent_amount"`
	RemainingBudget Money                          `json:"remaining_budget" redis:"remaining_budget"`
	Currency        string                         `json:"currency" redis:"currency"`
	BudgetItems     []*BudgetItem                  `json:"budget_items,omitempty"`
	PlatformBudgets map[PlatformType]*Money        `json:"platform_budgets,omitempty"`
	DailyBudget     *Money                         `json:"daily_budget,omitempty"`
	LastUpdated     time.Time                      `json:"last_updated" redis:"last_updated"`
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

// CampaignTimeline represents campaign timeline information
type CampaignTimeline struct {
	StartDate       time.Time              `json:"start_date" redis:"start_date"`
	EndDate         time.Time              `json:"end_date" redis:"end_date"`
	ActualStart     *time.Time             `json:"actual_start,omitempty" redis:"actual_start"`
	ActualEnd       *time.Time             `json:"actual_end,omitempty" redis:"actual_end"`
	Duration        int                    `json:"duration" redis:"duration"` // Days
	Milestones      []*CampaignMilestone   `json:"milestones,omitempty"`
	Phases          []*CampaignPhase       `json:"phases,omitempty"`
	IsOnTrack       bool                   `json:"is_on_track" redis:"is_on_track"`
}

// CampaignMilestone represents a campaign milestone
type CampaignMilestone struct {
	ID          uuid.UUID  `json:"id" redis:"id"`
	CampaignID  uuid.UUID  `json:"campaign_id" redis:"campaign_id"`
	Name        string     `json:"name" redis:"name"`
	Description string     `json:"description" redis:"description"`
	DueDate     time.Time  `json:"due_date" redis:"due_date"`
	CompletedAt *time.Time `json:"completed_at,omitempty" redis:"completed_at"`
	IsCompleted bool       `json:"is_completed" redis:"is_completed"`
	Order       int        `json:"order" redis:"order"`
}

// CampaignPhase represents a phase in the campaign
type CampaignPhase struct {
	ID          uuid.UUID  `json:"id" redis:"id"`
	CampaignID  uuid.UUID  `json:"campaign_id" redis:"campaign_id"`
	Name        string     `json:"name" redis:"name"`
	Description string     `json:"description" redis:"description"`
	StartDate   time.Time  `json:"start_date" redis:"start_date"`
	EndDate     time.Time  `json:"end_date" redis:"end_date"`
	Status      string     `json:"status" redis:"status"`
	Order       int        `json:"order" redis:"order"`
}

// CampaignObjective represents a campaign objective
type CampaignObjective struct {
	ID          uuid.UUID `json:"id" redis:"id"`
	CampaignID  uuid.UUID `json:"campaign_id" redis:"campaign_id"`
	Name        string    `json:"name" redis:"name"`
	Description string    `json:"description" redis:"description"`
	Type        string    `json:"type" redis:"type"`
	Target      float64   `json:"target" redis:"target"`
	Current     float64   `json:"current" redis:"current"`
	Unit        string    `json:"unit" redis:"unit"`
	Priority    int       `json:"priority" redis:"priority"`
	IsAchieved  bool      `json:"is_achieved" redis:"is_achieved"`
}

// CampaignKPI represents a campaign key performance indicator
type CampaignKPI struct {
	ID          uuid.UUID `json:"id" redis:"id"`
	CampaignID  uuid.UUID `json:"campaign_id" redis:"campaign_id"`
	Name        string    `json:"name" redis:"name"`
	Description string    `json:"description" redis:"description"`
	Metric      string    `json:"metric" redis:"metric"`
	Target      float64   `json:"target" redis:"target"`
	Current     float64   `json:"current" redis:"current"`
	Unit        string    `json:"unit" redis:"unit"`
	Trend       string    `json:"trend" redis:"trend"`
	LastUpdated time.Time `json:"last_updated" redis:"last_updated"`
}

// CampaignAnalytics represents analytics data for a campaign
type CampaignAnalytics struct {
	TotalImpressions   int64                                `json:"total_impressions"`
	TotalReach         int64                                `json:"total_reach"`
	TotalEngagement    int64                                `json:"total_engagement"`
	TotalClicks        int64                                `json:"total_clicks"`
	TotalConversions   int64                                `json:"total_conversions"`
	TotalSpend         Money                                `json:"total_spend"`
	AvgEngagementRate  float64                              `json:"avg_engagement_rate"`
	AvgCTR             float64                              `json:"avg_ctr"`
	ConversionRate     float64                              `json:"conversion_rate"`
	CostPerClick       float64                              `json:"cost_per_click"`
	CostPerConversion  float64                              `json:"cost_per_conversion"`
	ROI                float64                              `json:"roi"`
	ROAS               float64                              `json:"roas"` // Return on Ad Spend
	PlatformMetrics    map[PlatformType]*PlatformMetrics    `json:"platform_metrics,omitempty"`
	ContentPerformance []*ContentPerformance                `json:"content_performance,omitempty"`
	AudienceInsights   *AudienceInsights                    `json:"audience_insights,omitempty"`
	LastUpdated        time.Time                            `json:"last_updated"`
}

// ContentPerformance represents performance data for content within a campaign
type ContentPerformance struct {
	ContentID       uuid.UUID `json:"content_id"`
	ContentTitle    string    `json:"content_title"`
	Impressions     int64     `json:"impressions"`
	Engagement      int64     `json:"engagement"`
	Clicks          int64     `json:"clicks"`
	Conversions     int64     `json:"conversions"`
	EngagementRate  float64   `json:"engagement_rate"`
	CTR             float64   `json:"ctr"`
	ConversionRate  float64   `json:"conversion_rate"`
	Score           float64   `json:"score"`
}

// AudienceInsights represents audience insights for a campaign
type AudienceInsights struct {
	Demographics    *Demographics              `json:"demographics,omitempty"`
	TopInterests    []string                   `json:"top_interests,omitempty"`
	TopLocations    []string                   `json:"top_locations,omitempty"`
	DeviceBreakdown map[string]float64         `json:"device_breakdown,omitempty"`
	TimeOfDay       map[string]float64         `json:"time_of_day,omitempty"`
	DayOfWeek       map[string]float64         `json:"day_of_week,omitempty"`
	Sentiment       map[ContentSentiment]float64 `json:"sentiment,omitempty"`
}

// ABTest represents an A/B test within a campaign
type ABTest struct {
	ID              uuid.UUID              `json:"id" redis:"id"`
	CampaignID      uuid.UUID              `json:"campaign_id" redis:"campaign_id"`
	Name            string                 `json:"name" redis:"name"`
	Description     string                 `json:"description" redis:"description"`
	Status          ABTestStatus           `json:"status" redis:"status"`
	Type            ABTestType             `json:"type" redis:"type"`
	Variants        []*ABTestVariant       `json:"variants,omitempty"`
	TrafficSplit    map[string]float64     `json:"traffic_split" redis:"traffic_split"`
	Metric          string                 `json:"metric" redis:"metric"`
	Hypothesis      string                 `json:"hypothesis" redis:"hypothesis"`
	Results         *ABTestResults         `json:"results,omitempty"`
	StartDate       time.Time              `json:"start_date" redis:"start_date"`
	EndDate         time.Time              `json:"end_date" redis:"end_date"`
	MinSampleSize   int                    `json:"min_sample_size" redis:"min_sample_size"`
	ConfidenceLevel float64                `json:"confidence_level" redis:"confidence_level"`
	CreatedAt       time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" redis:"updated_at"`
}

// ABTestStatus defines the status of an A/B test
type ABTestStatus string

const (
	ABTestStatusDraft    ABTestStatus = "draft"
	ABTestStatusRunning  ABTestStatus = "running"
	ABTestStatusPaused   ABTestStatus = "paused"
	ABTestStatusCompleted ABTestStatus = "completed"
	ABTestStatusCancelled ABTestStatus = "cancelled"
)

// ABTestType defines the type of A/B test
type ABTestType string

const (
	ABTestTypeContent   ABTestType = "content"
	ABTestTypeAudience  ABTestType = "audience"
	ABTestTypeTiming    ABTestType = "timing"
	ABTestTypeCreative  ABTestType = "creative"
	ABTestTypePlatform  ABTestType = "platform"
)

// ABTestVariant represents a variant in an A/B test
type ABTestVariant struct {
	ID          uuid.UUID              `json:"id" redis:"id"`
	TestID      uuid.UUID              `json:"test_id" redis:"test_id"`
	Name        string                 `json:"name" redis:"name"`
	Description string                 `json:"description" redis:"description"`
	ContentID   *uuid.UUID             `json:"content_id,omitempty" redis:"content_id"`
	Content     *Content               `json:"content,omitempty"`
	Config      map[string]interface{} `json:"config" redis:"config"`
	TrafficShare float64               `json:"traffic_share" redis:"traffic_share"`
	Performance *VariantPerformance    `json:"performance,omitempty"`
	IsControl   bool                   `json:"is_control" redis:"is_control"`
}

// VariantPerformance represents performance data for an A/B test variant
type VariantPerformance struct {
	Impressions    int64   `json:"impressions"`
	Clicks         int64   `json:"clicks"`
	Conversions    int64   `json:"conversions"`
	CTR            float64 `json:"ctr"`
	ConversionRate float64 `json:"conversion_rate"`
	Cost           Money   `json:"cost"`
	Revenue        Money   `json:"revenue"`
	ROI            float64 `json:"roi"`
	StatSignificance float64 `json:"statistical_significance"`
}

// ABTestResults represents the results of an A/B test
type ABTestResults struct {
	Winner              *uuid.UUID         `json:"winner,omitempty"`
	WinnerConfidence    float64            `json:"winner_confidence"`
	StatSignificance    float64            `json:"statistical_significance"`
	PerformanceGain     float64            `json:"performance_gain"`
	Recommendation      string             `json:"recommendation"`
	Summary             string             `json:"summary"`
	VariantComparison   []*VariantComparison `json:"variant_comparison,omitempty"`
	CompletedAt         time.Time          `json:"completed_at"`
}

// VariantComparison represents a comparison between variants
type VariantComparison struct {
	VariantA        uuid.UUID `json:"variant_a"`
	VariantB        uuid.UUID `json:"variant_b"`
	Metric          string    `json:"metric"`
	Difference      float64   `json:"difference"`
	Significance    float64   `json:"significance"`
	ConfidenceLevel float64   `json:"confidence_level"`
}

// ApprovalWorkflow represents the approval workflow for a campaign
type ApprovalWorkflow struct {
	ID          uuid.UUID              `json:"id" redis:"id"`
	CampaignID  uuid.UUID              `json:"campaign_id" redis:"campaign_id"`
	Steps       []*ApprovalStep        `json:"steps,omitempty"`
	CurrentStep int                    `json:"current_step" redis:"current_step"`
	Status      ApprovalStatus         `json:"status" redis:"status"`
	CreatedAt   time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" redis:"updated_at"`
}

// ApprovalStep represents a step in the approval workflow
type ApprovalStep struct {
	ID          uuid.UUID      `json:"id" redis:"id"`
	WorkflowID  uuid.UUID      `json:"workflow_id" redis:"workflow_id"`
	Name        string         `json:"name" redis:"name"`
	Description string         `json:"description" redis:"description"`
	ApproverID  uuid.UUID      `json:"approver_id" redis:"approver_id"`
	Approver    *User          `json:"approver,omitempty"`
	Status      ApprovalStatus `json:"status" redis:"status"`
	Comments    string         `json:"comments" redis:"comments"`
	ApprovedAt  *time.Time     `json:"approved_at,omitempty" redis:"approved_at"`
	Order       int            `json:"order" redis:"order"`
	IsRequired  bool           `json:"is_required" redis:"is_required"`
}

// ApprovalStatus defines the status of approval
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
	ApprovalStatusSkipped  ApprovalStatus = "skipped"
)

// NewCampaign creates a new campaign with default values
func NewCampaign(name, description string, campaignType CampaignType, brandID, managerID uuid.UUID) *Campaign {
	now := time.Now()
	return &Campaign{
		ID:           uuid.New(),
		Name:         name,
		Description:  description,
		Type:         campaignType,
		Status:       CampaignStatusDraft,
		Priority:     CampaignPriorityMedium,
		BrandID:      brandID,
		ManagerID:    managerID,
		Platforms:    []PlatformType{},
		Hashtags:     []string{},
		Keywords:     []string{},
		Tags:         []string{},
		Tone:         ToneFriendly,
		Language:     "en",
		CustomFields: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
		ExternalIDs:  make(map[string]string),
		IsTemplate:   false,
		IsArchived:   false,
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    managerID,
		UpdatedBy:    managerID,
		Version:      1,
	}
}

// UpdateStatus updates the campaign status
func (c *Campaign) UpdateStatus(newStatus CampaignStatus, updatedBy uuid.UUID) {
	c.Status = newStatus
	c.UpdatedBy = updatedBy
	c.UpdatedAt = time.Now()
	c.Version++

	// Handle status-specific logic
	switch newStatus {
	case CampaignStatusActive:
		if c.Timeline != nil && c.Timeline.ActualStart == nil {
			now := time.Now()
			c.Timeline.ActualStart = &now
		}
	case CampaignStatusCompleted:
		if c.Timeline != nil {
			now := time.Now()
			c.Timeline.ActualEnd = &now
		}
	}
}

// AddMember adds a member to the campaign
func (c *Campaign) AddMember(member *CampaignMember) {
	member.CampaignID = c.ID
	c.TeamMembers = append(c.TeamMembers, member)
	c.UpdatedAt = time.Now()
	c.Version++
}

// AddContent adds content to the campaign
func (c *Campaign) AddContent(content *Content) {
	content.CampaignID = &c.ID
	c.Content = append(c.Content, content)
	c.UpdatedAt = time.Now()
	c.Version++
}

// IsActive checks if the campaign is currently active
func (c *Campaign) IsActive() bool {
	return c.Status == CampaignStatusActive && !c.IsArchived
}

// IsCompleted checks if the campaign is completed
func (c *Campaign) IsCompleted() bool {
	return c.Status == CampaignStatusCompleted
}

// GetProgress calculates the campaign progress percentage
func (c *Campaign) GetProgress() float64 {
	if c.Timeline == nil {
		return 0
	}

	now := time.Now()
	if now.Before(c.Timeline.StartDate) {
		return 0
	}
	if now.After(c.Timeline.EndDate) {
		return 100
	}

	total := c.Timeline.EndDate.Sub(c.Timeline.StartDate)
	elapsed := now.Sub(c.Timeline.StartDate)
	
	return float64(elapsed) / float64(total) * 100
}

// GetBudgetUtilization calculates budget utilization percentage
func (c *Campaign) GetBudgetUtilization() float64 {
	if c.Budget == nil || c.Budget.TotalBudget.Amount == 0 {
		return 0
	}
	return (c.Budget.SpentAmount.Amount / c.Budget.TotalBudget.Amount) * 100
}

// Archive archives the campaign
func (c *Campaign) Archive(archivedBy uuid.UUID) {
	c.IsArchived = true
	c.Status = CampaignStatusArchived
	c.UpdatedBy = archivedBy
	c.UpdatedAt = time.Now()
	c.Version++
}
