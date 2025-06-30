package repositories

import (
	"context"
	"time"

	"go-coffee-ai-agents/social-media-content-agent/internal/domain/entities"

	"github.com/google/uuid"
)

// ContentRepository defines the interface for content data access
type ContentRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, content *entities.Content) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Content, error)
	Update(ctx context.Context, content *entities.Content) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Listing and filtering
	List(ctx context.Context, filter *ContentFilter) ([]*entities.Content, error)
	ListByBrand(ctx context.Context, brandID uuid.UUID, filter *ContentFilter) ([]*entities.Content, error)
	ListByCampaign(ctx context.Context, campaignID uuid.UUID, filter *ContentFilter) ([]*entities.Content, error)
	ListByCreator(ctx context.Context, creatorID uuid.UUID, filter *ContentFilter) ([]*entities.Content, error)
	ListByStatus(ctx context.Context, status entities.ContentStatus, filter *ContentFilter) ([]*entities.Content, error)
	ListByPlatform(ctx context.Context, platform entities.PlatformType, filter *ContentFilter) ([]*entities.Content, error)

	// Advanced queries
	GetScheduledContent(ctx context.Context, from, to time.Time) ([]*entities.Content, error)
	GetContentDueForPublishing(ctx context.Context) ([]*entities.Content, error)
	GetExpiredContent(ctx context.Context) ([]*entities.Content, error)
	GetContentByHashtag(ctx context.Context, hashtag string, filter *ContentFilter) ([]*entities.Content, error)
	GetTrendingContent(ctx context.Context, period time.Duration, limit int) ([]*entities.Content, error)
	GetTopPerformingContent(ctx context.Context, brandID uuid.UUID, period time.Duration, limit int) ([]*entities.Content, error)

	// Media assets
	AddMediaAsset(ctx context.Context, asset *entities.MediaAsset) error
	GetMediaAssets(ctx context.Context, contentID uuid.UUID) ([]*entities.MediaAsset, error)
	UpdateMediaAsset(ctx context.Context, asset *entities.MediaAsset) error
	DeleteMediaAsset(ctx context.Context, assetID uuid.UUID) error

	// Content variations
	AddVariation(ctx context.Context, variation *entities.ContentVariation) error
	GetVariations(ctx context.Context, contentID uuid.UUID) ([]*entities.ContentVariation, error)
	UpdateVariation(ctx context.Context, variation *entities.ContentVariation) error
	DeleteVariation(ctx context.Context, variationID uuid.UUID) error

	// Analytics and reporting
	GetContentMetrics(ctx context.Context, filter *ContentMetricsFilter) (*ContentMetrics, error)
	GetPerformanceReport(ctx context.Context, brandID uuid.UUID, period time.Duration) (*PerformanceReport, error)
	GetEngagementTrends(ctx context.Context, brandID uuid.UUID, period time.Duration) (*EngagementTrends, error)
	GetHashtagAnalytics(ctx context.Context, brandID uuid.UUID, period time.Duration) ([]*HashtagAnalytics, error)

	// Search and advanced queries
	Search(ctx context.Context, query string, filter *ContentFilter) ([]*entities.Content, error)
	GetContentByTags(ctx context.Context, tags []string, filter *ContentFilter) ([]*entities.Content, error)
	GetContentByKeywords(ctx context.Context, keywords []string, filter *ContentFilter) ([]*entities.Content, error)
	GetSimilarContent(ctx context.Context, contentID uuid.UUID, limit int) ([]*entities.Content, error)

	// Bulk operations
	BulkCreate(ctx context.Context, contents []*entities.Content) error
	BulkUpdate(ctx context.Context, contents []*entities.Content) error
	BulkUpdateStatus(ctx context.Context, contentIDs []uuid.UUID, status entities.ContentStatus, updatedBy uuid.UUID) error
	BulkSchedule(ctx context.Context, contentIDs []uuid.UUID, scheduledAt time.Time, updatedBy uuid.UUID) error

	// Archiving and cleanup
	Archive(ctx context.Context, contentID uuid.UUID, archivedBy uuid.UUID) error
	Unarchive(ctx context.Context, contentID uuid.UUID, unarchivedBy uuid.UUID) error
	GetArchivedContent(ctx context.Context, filter *ContentFilter) ([]*entities.Content, error)
	CleanupExpiredContent(ctx context.Context, olderThan time.Time) error

	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context, repo ContentRepository) error) error
}

// ContentFilter defines filtering options for content queries
type ContentFilter struct {
	BrandIDs      []uuid.UUID                 `json:"brand_ids,omitempty"`
	CampaignIDs   []uuid.UUID                 `json:"campaign_ids,omitempty"`
	CreatorIDs    []uuid.UUID                 `json:"creator_ids,omitempty"`
	ApproverIDs   []uuid.UUID                 `json:"approver_ids,omitempty"`
	Types         []entities.ContentType      `json:"types,omitempty"`
	Formats       []entities.ContentFormat    `json:"formats,omitempty"`
	Statuses      []entities.ContentStatus    `json:"statuses,omitempty"`
	Priorities    []entities.ContentPriority  `json:"priorities,omitempty"`
	Categories    []entities.ContentCategory  `json:"categories,omitempty"`
	Platforms     []entities.PlatformType     `json:"platforms,omitempty"`
	Tones         []entities.ContentTone      `json:"tones,omitempty"`
	Sentiments    []entities.ContentSentiment `json:"sentiments,omitempty"`
	Languages     []string                    `json:"languages,omitempty"`
	Tags          []string                    `json:"tags,omitempty"`
	Keywords      []string                    `json:"keywords,omitempty"`
	Hashtags      []string                    `json:"hashtags,omitempty"`
	ScheduledFrom *time.Time                  `json:"scheduled_from,omitempty"`
	ScheduledTo   *time.Time                  `json:"scheduled_to,omitempty"`
	PublishedFrom *time.Time                  `json:"published_from,omitempty"`
	PublishedTo   *time.Time                  `json:"published_to,omitempty"`
	CreatedAfter  *time.Time                  `json:"created_after,omitempty"`
	CreatedBefore *time.Time                  `json:"created_before,omitempty"`
	UpdatedAfter  *time.Time                  `json:"updated_after,omitempty"`
	UpdatedBefore *time.Time                  `json:"updated_before,omitempty"`
	AIGenerated   *bool                       `json:"ai_generated,omitempty"`
	IsTemplate    *bool                       `json:"is_template,omitempty"`
	IsArchived    *bool                       `json:"is_archived,omitempty"`
	HasMedia      *bool                       `json:"has_media,omitempty"`
	HasVariations *bool                       `json:"has_variations,omitempty"`
	MinEngagement *float64                    `json:"min_engagement,omitempty"`
	MaxEngagement *float64                    `json:"max_engagement,omitempty"`
	SortBy        string                      `json:"sort_by,omitempty"`
	SortOrder     string                      `json:"sort_order,omitempty"`
	Limit         int                         `json:"limit,omitempty"`
	Offset        int                         `json:"offset,omitempty"`
}

// ContentMetricsFilter defines filtering options for content metrics
type ContentMetricsFilter struct {
	BrandIDs      []uuid.UUID             `json:"brand_ids,omitempty"`
	CampaignIDs   []uuid.UUID             `json:"campaign_ids,omitempty"`
	Platforms     []entities.PlatformType `json:"platforms,omitempty"`
	ContentTypes  []entities.ContentType  `json:"content_types,omitempty"`
	Period        time.Duration           `json:"period"`
	StartDate     time.Time               `json:"start_date"`
	EndDate       time.Time               `json:"end_date"`
	GroupBy       string                  `json:"group_by,omitempty"`
	IncludeTrends bool                    `json:"include_trends"`
}

// ContentMetrics contains content metrics and analytics
type ContentMetrics struct {
	Period                string                          `json:"period"`
	TotalContent          int                             `json:"total_content"`
	PublishedContent      int                             `json:"published_content"`
	ScheduledContent      int                             `json:"scheduled_content"`
	DraftContent          int                             `json:"draft_content"`
	TotalImpressions      int64                           `json:"total_impressions"`
	TotalEngagement       int64                           `json:"total_engagement"`
	TotalReach            int64                           `json:"total_reach"`
	AverageEngagementRate float64                         `json:"average_engagement_rate"`
	TopPerformingContent  []*ContentPerformance           `json:"top_performing_content"`
	ContentByType         map[entities.ContentType]int    `json:"content_by_type"`
	ContentByStatus       map[entities.ContentStatus]int  `json:"content_by_status"`
	ContentByPlatform     map[entities.PlatformType]int   `json:"content_by_platform"`
	EngagementByPlatform  map[entities.PlatformType]int64 `json:"engagement_by_platform"`
	TrendData             map[string][]float64            `json:"trend_data,omitempty"`
	GeneratedAt           time.Time                       `json:"generated_at"`
}

// ContentPerformance represents performance data for content
type ContentPerformance struct {
	ContentID      uuid.UUID             `json:"content_id"`
	Title          string                `json:"title"`
	Type           entities.ContentType  `json:"type"`
	Platform       entities.PlatformType `json:"platform"`
	PublishedAt    time.Time             `json:"published_at"`
	Impressions    int64                 `json:"impressions"`
	Engagement     int64                 `json:"engagement"`
	Reach          int64                 `json:"reach"`
	Clicks         int64                 `json:"clicks"`
	Shares         int64                 `json:"shares"`
	Comments       int64                 `json:"comments"`
	Likes          int64                 `json:"likes"`
	Saves          int64                 `json:"saves"`
	EngagementRate float64               `json:"engagement_rate"`
	CTR            float64               `json:"ctr"`
	ViralityScore  float64               `json:"virality_score"`
	QualityScore   float64               `json:"quality_score"`
}

// PerformanceReport represents a comprehensive performance report
type PerformanceReport struct {
	BrandID              uuid.UUID                                  `json:"brand_id"`
	Period               string                                     `json:"period"`
	Overview             *PerformanceOverview                       `json:"overview"`
	PlatformBreakdown    map[entities.PlatformType]*PlatformStats   `json:"platform_breakdown"`
	ContentTypeBreakdown map[entities.ContentType]*ContentTypeStats `json:"content_type_breakdown"`
	TopContent           []*ContentPerformance                      `json:"top_content"`
	TopHashtags          []*HashtagPerformance                      `json:"top_hashtags"`
	AudienceInsights     *AudienceInsights                          `json:"audience_insights"`
	Recommendations      []string                                   `json:"recommendations"`
	GeneratedAt          time.Time                                  `json:"generated_at"`
}

// PerformanceOverview represents overall performance metrics
type PerformanceOverview struct {
	TotalContent          int                   `json:"total_content"`
	TotalImpressions      int64                 `json:"total_impressions"`
	TotalEngagement       int64                 `json:"total_engagement"`
	TotalReach            int64                 `json:"total_reach"`
	AverageEngagementRate float64               `json:"average_engagement_rate"`
	GrowthRate            float64               `json:"growth_rate"`
	ViralContent          int                   `json:"viral_content"`
	TopPerformingPlatform entities.PlatformType `json:"top_performing_platform"`
}

// PlatformStats represents platform-specific statistics
type PlatformStats struct {
	Platform        entities.PlatformType `json:"platform"`
	ContentCount    int                   `json:"content_count"`
	Impressions     int64                 `json:"impressions"`
	Engagement      int64                 `json:"engagement"`
	Reach           int64                 `json:"reach"`
	EngagementRate  float64               `json:"engagement_rate"`
	GrowthRate      float64               `json:"growth_rate"`
	BestPostingTime string                `json:"best_posting_time"`
	TopContentType  entities.ContentType  `json:"top_content_type"`
}

// ContentTypeStats represents content type statistics
type ContentTypeStats struct {
	ContentType      entities.ContentType `json:"content_type"`
	Count            int                  `json:"count"`
	Impressions      int64                `json:"impressions"`
	Engagement       int64                `json:"engagement"`
	EngagementRate   float64              `json:"engagement_rate"`
	AverageReach     int64                `json:"average_reach"`
	PerformanceScore float64              `json:"performance_score"`
}

// HashtagPerformance represents hashtag performance metrics
type HashtagPerformance struct {
	Hashtag           string  `json:"hashtag"`
	UsageCount        int     `json:"usage_count"`
	TotalImpressions  int64   `json:"total_impressions"`
	TotalEngagement   int64   `json:"total_engagement"`
	AverageEngagement float64 `json:"average_engagement"`
	TrendingScore     float64 `json:"trending_score"`
	RecommendedUse    bool    `json:"recommended_use"`
}

// EngagementTrends represents engagement trends over time
type EngagementTrends struct {
	BrandID     uuid.UUID                                 `json:"brand_id"`
	Period      string                                    `json:"period"`
	Daily       map[string]*DailyEngagement               `json:"daily"`
	Weekly      map[string]*WeeklyEngagement              `json:"weekly"`
	Monthly     map[string]*MonthlyEngagement             `json:"monthly"`
	Platforms   map[entities.PlatformType]*PlatformTrends `json:"platforms"`
	GeneratedAt time.Time                                 `json:"generated_at"`
}

// DailyEngagement represents daily engagement metrics
type DailyEngagement struct {
	Date        string  `json:"date"`
	Impressions int64   `json:"impressions"`
	Engagement  int64   `json:"engagement"`
	Reach       int64   `json:"reach"`
	Posts       int     `json:"posts"`
	Rate        float64 `json:"rate"`
}

// WeeklyEngagement represents weekly engagement metrics
type WeeklyEngagement struct {
	Week        string  `json:"week"`
	Impressions int64   `json:"impressions"`
	Engagement  int64   `json:"engagement"`
	Reach       int64   `json:"reach"`
	Posts       int     `json:"posts"`
	Rate        float64 `json:"rate"`
	Growth      float64 `json:"growth"`
}

// MonthlyEngagement represents monthly engagement metrics
type MonthlyEngagement struct {
	Month       string  `json:"month"`
	Impressions int64   `json:"impressions"`
	Engagement  int64   `json:"engagement"`
	Reach       int64   `json:"reach"`
	Posts       int     `json:"posts"`
	Rate        float64 `json:"rate"`
	Growth      float64 `json:"growth"`
}

// PlatformTrends represents platform-specific trends
type PlatformTrends struct {
	Platform   entities.PlatformType `json:"platform"`
	Trend      string                `json:"trend"`
	GrowthRate float64               `json:"growth_rate"`
	BestTimes  []string              `json:"best_times"`
	TopContent []string              `json:"top_content"`
	Insights   []string              `json:"insights"`
}

// HashtagAnalytics represents hashtag analytics
type HashtagAnalytics struct {
	Hashtag           string                        `json:"hashtag"`
	TotalUses         int                           `json:"total_uses"`
	UniqueContent     int                           `json:"unique_content"`
	TotalImpressions  int64                         `json:"total_impressions"`
	TotalEngagement   int64                         `json:"total_engagement"`
	AverageEngagement float64                       `json:"average_engagement"`
	TrendingScore     float64                       `json:"trending_score"`
	Sentiment         entities.ContentSentiment     `json:"sentiment"`
	RelatedHashtags   []string                      `json:"related_hashtags"`
	TopContent        []uuid.UUID                   `json:"top_content"`
	PlatformBreakdown map[entities.PlatformType]int `json:"platform_breakdown"`
	TimeDistribution  map[string]int                `json:"time_distribution"`
	LastUpdated       time.Time                     `json:"last_updated"`
}

// AudienceInsights represents audience insights
type AudienceInsights struct {
	Demographics          *entities.Demographics                `json:"demographics,omitempty"`
	TopInterests          []string                              `json:"top_interests,omitempty"`
	TopLocations          []string                              `json:"top_locations,omitempty"`
	DeviceBreakdown       map[string]float64                    `json:"device_breakdown,omitempty"`
	TimeOfDay             map[string]float64                    `json:"time_of_day,omitempty"`
	DayOfWeek             map[string]float64                    `json:"day_of_week,omitempty"`
	EngagementPatterns    map[string]float64                    `json:"engagement_patterns,omitempty"`
	ContentPreferences    map[entities.ContentType]float64      `json:"content_preferences,omitempty"`
	PlatformPreferences   map[entities.PlatformType]float64     `json:"platform_preferences,omitempty"`
	SentimentDistribution map[entities.ContentSentiment]float64 `json:"sentiment_distribution,omitempty"`
	GrowthTrends          map[string]float64                    `json:"growth_trends,omitempty"`
	LastUpdated           time.Time                             `json:"last_updated"`
}

// CampaignRepository defines the interface for campaign data access
type CampaignRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, campaign *entities.Campaign) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Campaign, error)
	Update(ctx context.Context, campaign *entities.Campaign) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Listing and filtering
	List(ctx context.Context, filter *CampaignFilter) ([]*entities.Campaign, error)
	ListByBrand(ctx context.Context, brandID uuid.UUID, filter *CampaignFilter) ([]*entities.Campaign, error)
	ListByManager(ctx context.Context, managerID uuid.UUID, filter *CampaignFilter) ([]*entities.Campaign, error)
	ListByStatus(ctx context.Context, status entities.CampaignStatus, filter *CampaignFilter) ([]*entities.Campaign, error)
	ListActive(ctx context.Context, filter *CampaignFilter) ([]*entities.Campaign, error)

	// Members
	AddMember(ctx context.Context, member *entities.CampaignMember) error
	UpdateMember(ctx context.Context, member *entities.CampaignMember) error
	RemoveMember(ctx context.Context, campaignID, userID uuid.UUID) error
	GetMembers(ctx context.Context, campaignID uuid.UUID) ([]*entities.CampaignMember, error)

	// Content
	GetCampaignContent(ctx context.Context, campaignID uuid.UUID, filter *ContentFilter) ([]*entities.Content, error)
	AddContentToCampaign(ctx context.Context, campaignID, contentID uuid.UUID) error
	RemoveContentFromCampaign(ctx context.Context, campaignID, contentID uuid.UUID) error

	// A/B Tests
	CreateABTest(ctx context.Context, test *entities.ABTest) error
	GetABTests(ctx context.Context, campaignID uuid.UUID) ([]*entities.ABTest, error)
	UpdateABTest(ctx context.Context, test *entities.ABTest) error
	GetABTestResults(ctx context.Context, testID uuid.UUID) (*entities.ABTestResults, error)

	// Analytics
	GetCampaignMetrics(ctx context.Context, campaignID uuid.UUID, period time.Duration) (*CampaignMetrics, error)
	GetCampaignPerformance(ctx context.Context, campaignID uuid.UUID) (*CampaignPerformance, error)

	// Search
	Search(ctx context.Context, query string, filter *CampaignFilter) ([]*entities.Campaign, error)

	// Bulk operations
	BulkUpdate(ctx context.Context, campaigns []*entities.Campaign) error
	BulkUpdateStatus(ctx context.Context, campaignIDs []uuid.UUID, status entities.CampaignStatus, updatedBy uuid.UUID) error

	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context, repo CampaignRepository) error) error
}

// CampaignFilter defines filtering options for campaign queries
type CampaignFilter struct {
	BrandIDs      []uuid.UUID                 `json:"brand_ids,omitempty"`
	ManagerIDs    []uuid.UUID                 `json:"manager_ids,omitempty"`
	Types         []entities.CampaignType     `json:"types,omitempty"`
	Statuses      []entities.CampaignStatus   `json:"statuses,omitempty"`
	Priorities    []entities.CampaignPriority `json:"priorities,omitempty"`
	Categories    []entities.CampaignCategory `json:"categories,omitempty"`
	Platforms     []entities.PlatformType     `json:"platforms,omitempty"`
	Tags          []string                    `json:"tags,omitempty"`
	Keywords      []string                    `json:"keywords,omitempty"`
	StartDateFrom *time.Time                  `json:"start_date_from,omitempty"`
	StartDateTo   *time.Time                  `json:"start_date_to,omitempty"`
	EndDateFrom   *time.Time                  `json:"end_date_from,omitempty"`
	EndDateTo     *time.Time                  `json:"end_date_to,omitempty"`
	CreatedAfter  *time.Time                  `json:"created_after,omitempty"`
	CreatedBefore *time.Time                  `json:"created_before,omitempty"`
	IsTemplate    *bool                       `json:"is_template,omitempty"`
	IsArchived    *bool                       `json:"is_archived,omitempty"`
	MinBudget     *float64                    `json:"min_budget,omitempty"`
	MaxBudget     *float64                    `json:"max_budget,omitempty"`
	SortBy        string                      `json:"sort_by,omitempty"`
	SortOrder     string                      `json:"sort_order,omitempty"`
	Limit         int                         `json:"limit,omitempty"`
	Offset        int                         `json:"offset,omitempty"`
}

// CampaignMetrics contains campaign-specific metrics
type CampaignMetrics struct {
	CampaignID           uuid.UUID                                `json:"campaign_id"`
	Period               string                                   `json:"period"`
	TotalContent         int                                      `json:"total_content"`
	PublishedContent     int                                      `json:"published_content"`
	TotalImpressions     int64                                    `json:"total_impressions"`
	TotalEngagement      int64                                    `json:"total_engagement"`
	TotalReach           int64                                    `json:"total_reach"`
	TotalClicks          int64                                    `json:"total_clicks"`
	TotalConversions     int64                                    `json:"total_conversions"`
	EngagementRate       float64                                  `json:"engagement_rate"`
	CTR                  float64                                  `json:"ctr"`
	ConversionRate       float64                                  `json:"conversion_rate"`
	CostPerClick         float64                                  `json:"cost_per_click"`
	CostPerConversion    float64                                  `json:"cost_per_conversion"`
	ROI                  float64                                  `json:"roi"`
	ROAS                 float64                                  `json:"roas"`
	BudgetUtilization    float64                                  `json:"budget_utilization"`
	ProgressPercentage   float64                                  `json:"progress_percentage"`
	TopPerformingContent []*ContentPerformance                    `json:"top_performing_content"`
	PlatformBreakdown    map[entities.PlatformType]*PlatformStats `json:"platform_breakdown"`
	GeneratedAt          time.Time                                `json:"generated_at"`
}

// CampaignPerformance contains campaign performance data
type CampaignPerformance struct {
	CampaignID       uuid.UUID               `json:"campaign_id"`
	Name             string                  `json:"name"`
	Status           entities.CampaignStatus `json:"status"`
	StartDate        time.Time               `json:"start_date"`
	EndDate          time.Time               `json:"end_date"`
	Progress         float64                 `json:"progress"`
	Budget           entities.Money          `json:"budget"`
	Spent            entities.Money          `json:"spent"`
	Remaining        entities.Money          `json:"remaining"`
	ContentCount     int                     `json:"content_count"`
	PublishedCount   int                     `json:"published_count"`
	Impressions      int64                   `json:"impressions"`
	Engagement       int64                   `json:"engagement"`
	Reach            int64                   `json:"reach"`
	Clicks           int64                   `json:"clicks"`
	Conversions      int64                   `json:"conversions"`
	EngagementRate   float64                 `json:"engagement_rate"`
	CTR              float64                 `json:"ctr"`
	ConversionRate   float64                 `json:"conversion_rate"`
	ROI              float64                 `json:"roi"`
	ROAS             float64                 `json:"roas"`
	QualityScore     float64                 `json:"quality_score"`
	PerformanceGrade string                  `json:"performance_grade"`
	Insights         []string                `json:"insights"`
	Recommendations  []string                `json:"recommendations"`
	LastUpdated      time.Time               `json:"last_updated"`
}
