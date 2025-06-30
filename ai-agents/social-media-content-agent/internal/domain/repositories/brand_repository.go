package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/social-media-content-agent/internal/domain/entities"
)

// BrandRepository defines the interface for brand data access
type BrandRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, brand *entities.Brand) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Brand, error)
	GetByName(ctx context.Context, name string) (*entities.Brand, error)
	Update(ctx context.Context, brand *entities.Brand) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Listing and filtering
	List(ctx context.Context, filter *BrandFilter) ([]*entities.Brand, error)
	ListByStatus(ctx context.Context, status entities.BrandStatus, filter *BrandFilter) ([]*entities.Brand, error)
	ListByIndustry(ctx context.Context, industry string, filter *BrandFilter) ([]*entities.Brand, error)
	
	// Social profiles
	AddSocialProfile(ctx context.Context, profile *entities.SocialProfile) error
	UpdateSocialProfile(ctx context.Context, profile *entities.SocialProfile) error
	DeleteSocialProfile(ctx context.Context, profileID uuid.UUID) error
	GetSocialProfiles(ctx context.Context, brandID uuid.UUID) ([]*entities.SocialProfile, error)
	GetSocialProfileByPlatform(ctx context.Context, brandID uuid.UUID, platform entities.PlatformType) (*entities.SocialProfile, error)
	
	// Brand guidelines and assets
	UpdateGuidelines(ctx context.Context, brandID uuid.UUID, guidelines *entities.BrandGuidelines) error
	GetGuidelines(ctx context.Context, brandID uuid.UUID) (*entities.BrandGuidelines, error)
	
	// Content templates
	AddContentTemplate(ctx context.Context, template *entities.ContentTemplate) error
	UpdateContentTemplate(ctx context.Context, template *entities.ContentTemplate) error
	DeleteContentTemplate(ctx context.Context, templateID uuid.UUID) error
	GetContentTemplates(ctx context.Context, brandID uuid.UUID, filter *TemplateFilter) ([]*entities.ContentTemplate, error)
	
	// Hashtag sets
	AddHashtagSet(ctx context.Context, hashtagSet *entities.HashtagSet) error
	UpdateHashtagSet(ctx context.Context, hashtagSet *entities.HashtagSet) error
	DeleteHashtagSet(ctx context.Context, hashtagSetID uuid.UUID) error
	GetHashtagSets(ctx context.Context, brandID uuid.UUID, filter *HashtagSetFilter) ([]*entities.HashtagSet, error)
	
	// Compliance rules
	AddComplianceRule(ctx context.Context, rule *entities.ComplianceRule) error
	UpdateComplianceRule(ctx context.Context, rule *entities.ComplianceRule) error
	DeleteComplianceRule(ctx context.Context, ruleID uuid.UUID) error
	GetComplianceRules(ctx context.Context, brandID uuid.UUID) ([]*entities.ComplianceRule, error)
	
	// Search
	Search(ctx context.Context, query string, filter *BrandFilter) ([]*entities.Brand, error)
	
	// Bulk operations
	BulkUpdate(ctx context.Context, brands []*entities.Brand) error
	
	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context, repo BrandRepository) error) error
}

// BrandFilter defines filtering options for brand queries
type BrandFilter struct {
	Statuses     []entities.BrandStatus `json:"statuses,omitempty"`
	Industries   []string               `json:"industries,omitempty"`
	Languages    []string               `json:"languages,omitempty"`
	Countries    []string               `json:"countries,omitempty"`
	IsActive     *bool                  `json:"is_active,omitempty"`
	CreatedAfter *time.Time             `json:"created_after,omitempty"`
	CreatedBefore *time.Time            `json:"created_before,omitempty"`
	SortBy       string                 `json:"sort_by,omitempty"`
	SortOrder    string                 `json:"sort_order,omitempty"`
	Limit        int                    `json:"limit,omitempty"`
	Offset       int                    `json:"offset,omitempty"`
}

// TemplateFilter defines filtering options for content template queries
type TemplateFilter struct {
	Types      []entities.ContentType     `json:"types,omitempty"`
	Categories []entities.ContentCategory `json:"categories,omitempty"`
	Platforms  []entities.PlatformType    `json:"platforms,omitempty"`
	Tags       []string                   `json:"tags,omitempty"`
	IsActive   *bool                      `json:"is_active,omitempty"`
	SortBy     string                     `json:"sort_by,omitempty"`
	SortOrder  string                     `json:"sort_order,omitempty"`
	Limit      int                        `json:"limit,omitempty"`
	Offset     int                        `json:"offset,omitempty"`
}

// HashtagSetFilter defines filtering options for hashtag set queries
type HashtagSetFilter struct {
	Categories []string                `json:"categories,omitempty"`
	Platforms  []entities.PlatformType `json:"platforms,omitempty"`
	IsActive   *bool                   `json:"is_active,omitempty"`
	SortBy     string                  `json:"sort_by,omitempty"`
	SortOrder  string                  `json:"sort_order,omitempty"`
	Limit      int                     `json:"limit,omitempty"`
	Offset     int                     `json:"offset,omitempty"`
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
	ListByBrandAccess(ctx context.Context, brandID uuid.UUID, filter *UserFilter) ([]*entities.User, error)
	ListByStatus(ctx context.Context, status entities.UserStatus, filter *UserFilter) ([]*entities.User, error)
	
	// Brand access management
	GrantBrandAccess(ctx context.Context, userID, brandID uuid.UUID) error
	RevokeBrandAccess(ctx context.Context, userID, brandID uuid.UUID) error
	GetUserBrands(ctx context.Context, userID uuid.UUID) ([]*entities.Brand, error)
	
	// Preferences
	UpdatePreferences(ctx context.Context, userID uuid.UUID, preferences *entities.UserPreferences) error
	GetPreferences(ctx context.Context, userID uuid.UUID) (*entities.UserPreferences, error)
	
	// Search
	Search(ctx context.Context, query string, filter *UserFilter) ([]*entities.User, error)
	
	// Bulk operations
	BulkUpdate(ctx context.Context, users []*entities.User) error
	BulkGrantBrandAccess(ctx context.Context, userIDs []uuid.UUID, brandID uuid.UUID) error
	BulkRevokeBrandAccess(ctx context.Context, userIDs []uuid.UUID, brandID uuid.UUID) error
	
	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context, repo UserRepository) error) error
}

// UserFilter defines filtering options for user queries
type UserFilter struct {
	Roles        []entities.UserRole   `json:"roles,omitempty"`
	Statuses     []entities.UserStatus `json:"statuses,omitempty"`
	BrandIDs     []uuid.UUID           `json:"brand_ids,omitempty"`
	TimeZones    []string              `json:"time_zones,omitempty"`
	Languages    []string              `json:"languages,omitempty"`
	IsActive     *bool                 `json:"is_active,omitempty"`
	CreatedAfter *time.Time            `json:"created_after,omitempty"`
	CreatedBefore *time.Time           `json:"created_before,omitempty"`
	LastLoginAfter *time.Time          `json:"last_login_after,omitempty"`
	LastLoginBefore *time.Time         `json:"last_login_before,omitempty"`
	SortBy       string                `json:"sort_by,omitempty"`
	SortOrder    string                `json:"sort_order,omitempty"`
	Limit        int                   `json:"limit,omitempty"`
	Offset       int                   `json:"offset,omitempty"`
}

// PostRepository defines the interface for post data access
type PostRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, post *entities.Post) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Post, error)
	GetByPlatformPostID(ctx context.Context, platform entities.PlatformType, platformPostID string) (*entities.Post, error)
	Update(ctx context.Context, post *entities.Post) error
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Listing and filtering
	List(ctx context.Context, filter *PostFilter) ([]*entities.Post, error)
	ListByContent(ctx context.Context, contentID uuid.UUID, filter *PostFilter) ([]*entities.Post, error)
	ListByPlatform(ctx context.Context, platform entities.PlatformType, filter *PostFilter) ([]*entities.Post, error)
	ListByStatus(ctx context.Context, status entities.PostStatus, filter *PostFilter) ([]*entities.Post, error)
	
	// Scheduling and publishing
	GetScheduledPosts(ctx context.Context, from, to time.Time) ([]*entities.Post, error)
	GetPostsDueForPublishing(ctx context.Context) ([]*entities.Post, error)
	GetFailedPosts(ctx context.Context, retryable bool) ([]*entities.Post, error)
	
	// Analytics and engagement
	UpdateAnalytics(ctx context.Context, postID uuid.UUID, analytics *entities.PostAnalytics) error
	UpdateEngagement(ctx context.Context, postID uuid.UUID, engagement *entities.PostEngagement) error
	GetPostAnalytics(ctx context.Context, postID uuid.UUID) (*entities.PostAnalytics, error)
	GetPostEngagement(ctx context.Context, postID uuid.UUID) (*entities.PostEngagement, error)
	
	// Comments and interactions
	AddComment(ctx context.Context, comment *entities.PostComment) error
	GetComments(ctx context.Context, postID uuid.UUID, filter *CommentFilter) ([]*entities.PostComment, error)
	UpdateComment(ctx context.Context, comment *entities.PostComment) error
	DeleteComment(ctx context.Context, commentID string) error
	
	AddInteraction(ctx context.Context, interaction *entities.PostInteraction) error
	GetInteractions(ctx context.Context, postID uuid.UUID, filter *InteractionFilter) ([]*entities.PostInteraction, error)
	
	// Performance metrics
	GetPostMetrics(ctx context.Context, filter *PostMetricsFilter) (*PostMetrics, error)
	GetTopPerformingPosts(ctx context.Context, brandID uuid.UUID, platform entities.PlatformType, period time.Duration, limit int) ([]*entities.Post, error)
	GetEngagementTrends(ctx context.Context, brandID uuid.UUID, period time.Duration) (*PostEngagementTrends, error)
	
	// Search
	Search(ctx context.Context, query string, filter *PostFilter) ([]*entities.Post, error)
	SearchByHashtag(ctx context.Context, hashtag string, filter *PostFilter) ([]*entities.Post, error)
	
	// Bulk operations
	BulkCreate(ctx context.Context, posts []*entities.Post) error
	BulkUpdate(ctx context.Context, posts []*entities.Post) error
	BulkUpdateStatus(ctx context.Context, postIDs []uuid.UUID, status entities.PostStatus, updatedBy uuid.UUID) error
	BulkSchedule(ctx context.Context, postIDs []uuid.UUID, scheduledAt time.Time, updatedBy uuid.UUID) error
	
	// Cleanup
	CleanupOldPosts(ctx context.Context, olderThan time.Time) error
	
	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context, repo PostRepository) error) error
}

// PostFilter defines filtering options for post queries
type PostFilter struct {
	ContentIDs     []uuid.UUID             `json:"content_ids,omitempty"`
	Platforms      []entities.PlatformType `json:"platforms,omitempty"`
	Statuses       []entities.PostStatus   `json:"statuses,omitempty"`
	Types          []entities.PostType     `json:"types,omitempty"`
	Hashtags       []string                `json:"hashtags,omitempty"`
	Mentions       []string                `json:"mentions,omitempty"`
	ScheduledFrom  *time.Time              `json:"scheduled_from,omitempty"`
	ScheduledTo    *time.Time              `json:"scheduled_to,omitempty"`
	PublishedFrom  *time.Time              `json:"published_from,omitempty"`
	PublishedTo    *time.Time              `json:"published_to,omitempty"`
	CreatedAfter   *time.Time              `json:"created_after,omitempty"`
	CreatedBefore  *time.Time              `json:"created_before,omitempty"`
	HasMedia       *bool                   `json:"has_media,omitempty"`
	HasLocation    *bool                   `json:"has_location,omitempty"`
	IsRepost       *bool                   `json:"is_repost,omitempty"`
	MinEngagement  *int64                  `json:"min_engagement,omitempty"`
	MaxEngagement  *int64                  `json:"max_engagement,omitempty"`
	MinImpressions *int64                  `json:"min_impressions,omitempty"`
	MaxImpressions *int64                  `json:"max_impressions,omitempty"`
	SortBy         string                  `json:"sort_by,omitempty"`
	SortOrder      string                  `json:"sort_order,omitempty"`
	Limit          int                     `json:"limit,omitempty"`
	Offset         int                     `json:"offset,omitempty"`
}

// CommentFilter defines filtering options for comment queries
type CommentFilter struct {
	Platforms       []entities.PlatformType    `json:"platforms,omitempty"`
	Sentiments      []entities.ContentSentiment `json:"sentiments,omitempty"`
	IsVerified      *bool                      `json:"is_verified,omitempty"`
	IsInfluencer    *bool                      `json:"is_influencer,omitempty"`
	RequiresResponse *bool                     `json:"requires_response,omitempty"`
	IsResponded     *bool                      `json:"is_responded,omitempty"`
	CreatedAfter    *time.Time                 `json:"created_after,omitempty"`
	CreatedBefore   *time.Time                 `json:"created_before,omitempty"`
	MinLikes        *int64                     `json:"min_likes,omitempty"`
	MaxLikes        *int64                     `json:"max_likes,omitempty"`
	SortBy          string                     `json:"sort_by,omitempty"`
	SortOrder       string                     `json:"sort_order,omitempty"`
	Limit           int                        `json:"limit,omitempty"`
	Offset          int                        `json:"offset,omitempty"`
}

// InteractionFilter defines filtering options for interaction queries
type InteractionFilter struct {
	Types         []entities.InteractionType `json:"types,omitempty"`
	Platforms     []entities.PlatformType    `json:"platforms,omitempty"`
	IsVerified    *bool                      `json:"is_verified,omitempty"`
	IsInfluencer  *bool                      `json:"is_influencer,omitempty"`
	MinFollowers  *int64                     `json:"min_followers,omitempty"`
	MaxFollowers  *int64                     `json:"max_followers,omitempty"`
	TimestampFrom *time.Time                 `json:"timestamp_from,omitempty"`
	TimestampTo   *time.Time                 `json:"timestamp_to,omitempty"`
	SortBy        string                     `json:"sort_by,omitempty"`
	SortOrder     string                     `json:"sort_order,omitempty"`
	Limit         int                        `json:"limit,omitempty"`
	Offset        int                        `json:"offset,omitempty"`
}

// PostMetricsFilter defines filtering options for post metrics
type PostMetricsFilter struct {
	BrandIDs     []uuid.UUID             `json:"brand_ids,omitempty"`
	ContentIDs   []uuid.UUID             `json:"content_ids,omitempty"`
	CampaignIDs  []uuid.UUID             `json:"campaign_ids,omitempty"`
	Platforms    []entities.PlatformType `json:"platforms,omitempty"`
	PostTypes    []entities.PostType     `json:"post_types,omitempty"`
	Period       time.Duration           `json:"period"`
	StartDate    time.Time               `json:"start_date"`
	EndDate      time.Time               `json:"end_date"`
	GroupBy      string                  `json:"group_by,omitempty"`
	IncludeTrends bool                   `json:"include_trends"`
}

// PostMetrics contains post metrics and analytics
type PostMetrics struct {
	Period              string                                    `json:"period"`
	TotalPosts          int                                       `json:"total_posts"`
	PublishedPosts      int                                       `json:"published_posts"`
	ScheduledPosts      int                                       `json:"scheduled_posts"`
	FailedPosts         int                                       `json:"failed_posts"`
	TotalImpressions    int64                                     `json:"total_impressions"`
	TotalEngagement     int64                                     `json:"total_engagement"`
	TotalReach          int64                                     `json:"total_reach"`
	TotalClicks         int64                                     `json:"total_clicks"`
	AverageEngagementRate float64                                 `json:"average_engagement_rate"`
	AverageCTR          float64                                   `json:"average_ctr"`
	TopPerformingPosts  []*PostPerformance                       `json:"top_performing_posts"`
	PostsByPlatform     map[entities.PlatformType]int            `json:"posts_by_platform"`
	PostsByType         map[entities.PostType]int                `json:"posts_by_type"`
	PostsByStatus       map[entities.PostStatus]int              `json:"posts_by_status"`
	EngagementByPlatform map[entities.PlatformType]int64         `json:"engagement_by_platform"`
	TrendData           map[string][]float64                     `json:"trend_data,omitempty"`
	GeneratedAt         time.Time                                 `json:"generated_at"`
}

// PostPerformance represents performance data for a post
type PostPerformance struct {
	PostID          uuid.UUID             `json:"post_id"`
	ContentID       uuid.UUID             `json:"content_id"`
	Platform        entities.PlatformType `json:"platform"`
	Type            entities.PostType     `json:"type"`
	PublishedAt     time.Time             `json:"published_at"`
	Impressions     int64                 `json:"impressions"`
	Engagement      int64                 `json:"engagement"`
	Reach           int64                 `json:"reach"`
	Clicks          int64                 `json:"clicks"`
	Shares          int64                 `json:"shares"`
	Comments        int64                 `json:"comments"`
	Likes           int64                 `json:"likes"`
	Saves           int64                 `json:"saves"`
	EngagementRate  float64               `json:"engagement_rate"`
	CTR             float64               `json:"ctr"`
	ViralityScore   float64               `json:"virality_score"`
	QualityScore    float64               `json:"quality_score"`
	PerformanceGrade string               `json:"performance_grade"`
}

// PostEngagementTrends represents engagement trends for posts
type PostEngagementTrends struct {
	BrandID     uuid.UUID                                     `json:"brand_id"`
	Period      string                                        `json:"period"`
	Daily       map[string]*DailyPostEngagement               `json:"daily"`
	Weekly      map[string]*WeeklyPostEngagement              `json:"weekly"`
	Monthly     map[string]*MonthlyPostEngagement             `json:"monthly"`
	Platforms   map[entities.PlatformType]*PlatformPostTrends `json:"platforms"`
	GeneratedAt time.Time                                     `json:"generated_at"`
}

// DailyPostEngagement represents daily post engagement metrics
type DailyPostEngagement struct {
	Date        string  `json:"date"`
	Posts       int     `json:"posts"`
	Impressions int64   `json:"impressions"`
	Engagement  int64   `json:"engagement"`
	Reach       int64   `json:"reach"`
	Clicks      int64   `json:"clicks"`
	Rate        float64 `json:"rate"`
}

// WeeklyPostEngagement represents weekly post engagement metrics
type WeeklyPostEngagement struct {
	Week        string  `json:"week"`
	Posts       int     `json:"posts"`
	Impressions int64   `json:"impressions"`
	Engagement  int64   `json:"engagement"`
	Reach       int64   `json:"reach"`
	Clicks      int64   `json:"clicks"`
	Rate        float64 `json:"rate"`
	Growth      float64 `json:"growth"`
}

// MonthlyPostEngagement represents monthly post engagement metrics
type MonthlyPostEngagement struct {
	Month       string  `json:"month"`
	Posts       int     `json:"posts"`
	Impressions int64   `json:"impressions"`
	Engagement  int64   `json:"engagement"`
	Reach       int64   `json:"reach"`
	Clicks      int64   `json:"clicks"`
	Rate        float64 `json:"rate"`
	Growth      float64 `json:"growth"`
}

// PlatformPostTrends represents platform-specific post trends
type PlatformPostTrends struct {
	Platform    entities.PlatformType `json:"platform"`
	Trend       string                `json:"trend"`
	GrowthRate  float64               `json:"growth_rate"`
	BestTimes   []string              `json:"best_times"`
	TopHashtags []string              `json:"top_hashtags"`
	Insights    []string              `json:"insights"`
}
