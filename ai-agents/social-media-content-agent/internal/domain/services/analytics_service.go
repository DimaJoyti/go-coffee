package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/social-media-content-agent/internal/domain/entities"
	"go-coffee-ai-agents/social-media-content-agent/internal/domain/repositories"
)

// AnalyticsService provides analytics and reporting functionality
type AnalyticsService struct {
	postRepo        repositories.PostRepository
	contentRepo     repositories.ContentRepository
	brandRepo       repositories.BrandRepository
	campaignRepo    repositories.CampaignRepository
	userRepo        repositories.UserRepository
	dataProcessor   AnalyticsDataProcessor
	eventPublisher  EventPublisher
	logger          Logger
}

// AnalyticsDataProcessor defines the interface for processing analytics data
type AnalyticsDataProcessor interface {
	ProcessPostAnalytics(ctx context.Context, posts []*entities.Post) (*ProcessedAnalytics, error)
	ProcessContentAnalytics(ctx context.Context, contents []*entities.Content) (*ProcessedAnalytics, error)
	ProcessBrandAnalytics(ctx context.Context, brandID uuid.UUID, timeRange TimeRange) (*BrandAnalytics, error)
	ProcessCampaignAnalytics(ctx context.Context, campaignID uuid.UUID) (*CampaignAnalytics, error)
	GenerateInsights(ctx context.Context, analytics *ProcessedAnalytics) ([]*AnalyticsInsight, error)
	CalculateTrends(ctx context.Context, metrics []*MetricPoint) (*TrendAnalysis, error)
	PredictPerformance(ctx context.Context, content *entities.Content) (*PerformancePrediction, error)
}

// TimeRange represents a time range for analytics queries
type TimeRange struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// ProcessedAnalytics represents processed analytics data
type ProcessedAnalytics struct {
	TotalPosts      int64                 `json:"total_posts"`
	TotalReach      int64                 `json:"total_reach"`
	TotalImpressions int64                `json:"total_impressions"`
	TotalEngagement int64                 `json:"total_engagement"`
	EngagementRate  float64               `json:"engagement_rate"`
	AverageReach    float64               `json:"average_reach"`
	TopPosts        []*entities.Post      `json:"top_posts"`
	PlatformBreakdown map[entities.PlatformType]*PlatformAnalytics `json:"platform_breakdown"`
	TimeSeriesData  []*MetricPoint        `json:"time_series_data"`
	ProcessedAt     time.Time             `json:"processed_at"`
}

// PlatformAnalytics represents analytics for a specific platform
type PlatformAnalytics struct {
	Platform        entities.PlatformType `json:"platform"`
	PostCount       int64                 `json:"post_count"`
	Reach           int64                 `json:"reach"`
	Impressions     int64                 `json:"impressions"`
	Engagement      int64                 `json:"engagement"`
	EngagementRate  float64               `json:"engagement_rate"`
	AverageReach    float64               `json:"average_reach"`
	TopPost         *entities.Post        `json:"top_post,omitempty"`
	GrowthRate      float64               `json:"growth_rate"`
}

// MetricPoint represents a point in time-series data
type MetricPoint struct {
	Timestamp   time.Time              `json:"timestamp"`
	Metrics     map[string]float64     `json:"metrics"`
	Platform    entities.PlatformType  `json:"platform,omitempty"`
	ContentID   *uuid.UUID             `json:"content_id,omitempty"`
}

// BrandAnalytics represents comprehensive brand analytics
type BrandAnalytics struct {
	BrandID         uuid.UUID                                     `json:"brand_id"`
	TimeRange       TimeRange                                     `json:"time_range"`
	Overview        *ProcessedAnalytics                           `json:"overview"`
	PlatformMetrics map[entities.PlatformType]*PlatformAnalytics  `json:"platform_metrics"`
	TopContent      []*entities.Content                           `json:"top_content"`
	Insights        []*AnalyticsInsight                           `json:"insights"`
	Trends          *TrendAnalysis                                `json:"trends"`
	Benchmarks      *BrandBenchmarks                              `json:"benchmarks"`
	GeneratedAt     time.Time                                     `json:"generated_at"`
}

// CampaignAnalytics represents campaign-specific analytics
type CampaignAnalytics struct {
	CampaignID      uuid.UUID                                     `json:"campaign_id"`
	Overview        *ProcessedAnalytics                           `json:"overview"`
	ContentMetrics  map[uuid.UUID]*ContentMetrics                 `json:"content_metrics"`
	PlatformMetrics map[entities.PlatformType]*PlatformAnalytics  `json:"platform_metrics"`
	Performance     *CampaignPerformance                          `json:"performance"`
	Insights        []*AnalyticsInsight                           `json:"insights"`
	Recommendations []*ActionableRecommendation                   `json:"recommendations"`
	GeneratedAt     time.Time                                     `json:"generated_at"`
}

// ContentMetrics represents metrics for individual content
type ContentMetrics struct {
	ContentID       uuid.UUID                                     `json:"content_id"`
	PostMetrics     map[uuid.UUID]*entities.PostAnalytics        `json:"post_metrics"`
	AggregatedMetrics *AggregatedContentMetrics                  `json:"aggregated_metrics"`
	Performance     *ContentPerformance                           `json:"performance"`
	Ranking         int                                           `json:"ranking"`
}

// AggregatedContentMetrics represents aggregated metrics across all posts for content
type AggregatedContentMetrics struct {
	TotalImpressions   int64   `json:"total_impressions"`
	TotalReach         int64   `json:"total_reach"`
	TotalEngagement    int64   `json:"total_engagement"`
	TotalLikes         int64   `json:"total_likes"`
	TotalComments      int64   `json:"total_comments"`
	TotalShares        int64   `json:"total_shares"`
	TotalClicks        int64   `json:"total_clicks"`
	AverageEngagementRate float64 `json:"average_engagement_rate"`
	BestPerformingPost *uuid.UUID `json:"best_performing_post,omitempty"`
}

// AnalyticsInsight represents an analytics insight
type AnalyticsInsight struct {
	ID          uuid.UUID               `json:"id"`
	Type        InsightType             `json:"type"`
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Severity    InsightSeverity         `json:"severity"`
	Category    InsightCategory         `json:"category"`
	Data        map[string]interface{}  `json:"data"`
	Action      *RecommendedAction      `json:"action,omitempty"`
	GeneratedAt time.Time               `json:"generated_at"`
}

// InsightType represents the type of insight
type InsightType string

const (
	InsightTypeTrend         InsightType = "trend"
	InsightTypeAnomaly       InsightType = "anomaly"
	InsightTypeOpportunity   InsightType = "opportunity"
	InsightTypeWarning       InsightType = "warning"
	InsightTypeBenchmark     InsightType = "benchmark"
	InsightTypeRecommendation InsightType = "recommendation"
)

// InsightSeverity represents the severity of an insight
type InsightSeverity string

const (
	InsightSeverityLow    InsightSeverity = "low"
	InsightSeverityMedium InsightSeverity = "medium"
	InsightSeverityHigh   InsightSeverity = "high"
	InsightSeverityCritical InsightSeverity = "critical"
)

// InsightCategory represents the category of an insight
type InsightCategory string

const (
	InsightCategoryPerformance InsightCategory = "performance"
	InsightCategoryEngagement  InsightCategory = "engagement"
	InsightCategoryReach       InsightCategory = "reach"
	InsightCategoryTiming      InsightCategory = "timing"
	InsightCategoryContent     InsightCategory = "content"
	InsightCategoryAudience    InsightCategory = "audience"
)

// RecommendedAction represents a recommended action
type RecommendedAction struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Priority    int                    `json:"priority"`
	Impact      string                 `json:"impact"`
	Effort      string                 `json:"effort"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// TrendAnalysis represents trend analysis results
type TrendAnalysis struct {
	OverallTrend    TrendDirection         `json:"overall_trend"`
	TrendStrength   float64                `json:"trend_strength"`
	PlatformTrends  map[entities.PlatformType]TrendDirection `json:"platform_trends"`
	MetricTrends    map[string]TrendDirection `json:"metric_trends"`
	Predictions     []*MetricPrediction    `json:"predictions"`
	AnalyzedPeriod  TimeRange              `json:"analyzed_period"`
	ComparisonPeriod TimeRange             `json:"comparison_period,omitempty"`
}

// TrendDirection represents the direction of a trend
type TrendDirection string

const (
	TrendDirectionUp    TrendDirection = "up"
	TrendDirectionDown  TrendDirection = "down"
	TrendDirectionFlat  TrendDirection = "flat"
	TrendDirectionVolatile TrendDirection = "volatile"
)

// MetricPrediction represents a prediction for a metric
type MetricPrediction struct {
	Metric      string    `json:"metric"`
	CurrentValue float64  `json:"current_value"`
	PredictedValue float64 `json:"predicted_value"`
	Confidence  float64   `json:"confidence"`
	TimeFrame   string    `json:"time_frame"`
	PredictedAt time.Time `json:"predicted_at"`
}

// PerformancePrediction represents performance prediction for content
type PerformancePrediction struct {
	ContentID       uuid.UUID              `json:"content_id"`
	PredictedMetrics map[string]float64    `json:"predicted_metrics"`
	Confidence      float64                `json:"confidence"`
	Factors         []*PredictionFactor    `json:"factors"`
	Recommendations []*OptimizationSuggestion `json:"recommendations"`
	PredictedAt     time.Time              `json:"predicted_at"`
}

// PredictionFactor represents a factor influencing the prediction
type PredictionFactor struct {
	Factor      string  `json:"factor"`
	Impact      float64 `json:"impact"`
	Description string  `json:"description"`
}

// OptimizationSuggestion represents a suggestion for optimization
type OptimizationSuggestion struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Impact      float64 `json:"expected_impact"`
	Confidence  float64 `json:"confidence"`
}

// BrandBenchmarks represents brand performance benchmarks
type BrandBenchmarks struct {
	Industry        string                 `json:"industry"`
	IndustryAverage map[string]float64     `json:"industry_average"`
	BrandMetrics    map[string]float64     `json:"brand_metrics"`
	Comparison      map[string]float64     `json:"comparison"` // percentage difference
	Ranking         *IndustryRanking       `json:"ranking,omitempty"`
}

// IndustryRanking represents brand ranking within industry
type IndustryRanking struct {
	OverallRank     int     `json:"overall_rank"`
	TotalBrands     int     `json:"total_brands"`
	Percentile      float64 `json:"percentile"`
	TopPerformers   []string `json:"top_performers,omitempty"`
}

// CampaignPerformance represents campaign performance metrics
type CampaignPerformance struct {
	Goals           map[string]float64     `json:"goals"`
	Achieved        map[string]float64     `json:"achieved"`
	Progress        map[string]float64     `json:"progress"` // percentage
	ROI             float64                `json:"roi"`
	CostPerEngagement float64             `json:"cost_per_engagement"`
	Status          CampaignStatus         `json:"status"`
	Recommendations []*ActionableRecommendation `json:"recommendations"`
}

// CampaignStatus represents campaign status
type CampaignStatus string

const (
	CampaignStatusOnTrack     CampaignStatus = "on_track"
	CampaignStatusAtRisk      CampaignStatus = "at_risk"
	CampaignStatusBehind      CampaignStatus = "behind"
	CampaignStatusExceeding   CampaignStatus = "exceeding"
)

// ContentPerformance represents content performance metrics
type ContentPerformance struct {
	Score           float64                `json:"score"`
	Rank            int                    `json:"rank"`
	BenchmarkComparison float64           `json:"benchmark_comparison"`
	PerformanceLevel PerformanceLevel     `json:"performance_level"`
	KeyStrengths    []string             `json:"key_strengths"`
	ImprovementAreas []string            `json:"improvement_areas"`
}

// PerformanceLevel represents performance level
type PerformanceLevel string

const (
	PerformanceLevelExcellent PerformanceLevel = "excellent"
	PerformanceLevelGood      PerformanceLevel = "good"
	PerformanceLevelAverage   PerformanceLevel = "average"
	PerformanceLevelPoor      PerformanceLevel = "poor"
)

// ActionableRecommendation represents an actionable recommendation
type ActionableRecommendation struct {
	ID          uuid.UUID              `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Type        RecommendationType     `json:"type"`
	Priority    RecommendationPriority `json:"priority"`
	Impact      float64                `json:"expected_impact"`
	Effort      float64                `json:"effort_required"`
	Actions     []*RecommendedAction   `json:"actions"`
	GeneratedAt time.Time              `json:"generated_at"`
}

// RecommendationType represents the type of recommendation
type RecommendationType string

const (
	RecommendationTypeContent    RecommendationType = "content"
	RecommendationTypeTiming     RecommendationType = "timing"
	RecommendationTypePlatform   RecommendationType = "platform"
	RecommendationTypeAudience   RecommendationType = "audience"
	RecommendationTypeEngagement RecommendationType = "engagement"
)

// RecommendationPriority represents recommendation priority
type RecommendationPriority string

const (
	RecommendationPriorityLow    RecommendationPriority = "low"
	RecommendationPriorityMedium RecommendationPriority = "medium"
	RecommendationPriorityHigh   RecommendationPriority = "high"
	RecommendationPriorityCritical RecommendationPriority = "critical"
)

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(
	postRepo repositories.PostRepository,
	contentRepo repositories.ContentRepository,
	brandRepo repositories.BrandRepository,
	campaignRepo repositories.CampaignRepository,
	userRepo repositories.UserRepository,
	dataProcessor AnalyticsDataProcessor,
	eventPublisher EventPublisher,
	logger Logger,
) *AnalyticsService {
	return &AnalyticsService{
		postRepo:       postRepo,
		contentRepo:    contentRepo,
		brandRepo:      brandRepo,
		campaignRepo:   campaignRepo,
		userRepo:       userRepo,
		dataProcessor:  dataProcessor,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// GenerateBrandAnalytics generates comprehensive analytics for a brand
func (as *AnalyticsService) GenerateBrandAnalytics(ctx context.Context, brandID uuid.UUID, timeRange TimeRange) (*BrandAnalytics, error) {
	as.logger.Info("Generating brand analytics", "brand_id", brandID, "time_range", timeRange)

	// Validate brand exists
	brand, err := as.brandRepo.GetByID(ctx, brandID)
	if err != nil {
		as.logger.Error("Failed to get brand", err, "brand_id", brandID)
		return nil, err
	}
	
	// Validate brand is active
	if !brand.IsActive() {
		return nil, fmt.Errorf("brand is not active: %s", brandID)
	}

	// Process brand analytics using data processor
	analytics, err := as.dataProcessor.ProcessBrandAnalytics(ctx, brandID, timeRange)
	if err != nil {
		as.logger.Error("Failed to process brand analytics", err, "brand_id", brandID)
		return nil, err
	}

	as.logger.Info("Brand analytics generated successfully", "brand_id", brandID)
	return analytics, nil
}

// GenerateCampaignAnalytics generates analytics for a campaign
func (as *AnalyticsService) GenerateCampaignAnalytics(ctx context.Context, campaignID uuid.UUID) (*CampaignAnalytics, error) {
	as.logger.Info("Generating campaign analytics", "campaign_id", campaignID)

	// Validate campaign exists
	campaign, err := as.campaignRepo.GetByID(ctx, campaignID)
	if err != nil {
		as.logger.Error("Failed to get campaign", err, "campaign_id", campaignID)
		return nil, err
	}
	
	// Validate campaign is active
	if !campaign.IsActive() {
		return nil, fmt.Errorf("campaign is not active: %s", campaignID)
	}

	// Process campaign analytics using data processor
	analytics, err := as.dataProcessor.ProcessCampaignAnalytics(ctx, campaignID)
	if err != nil {
		as.logger.Error("Failed to process campaign analytics", err, "campaign_id", campaignID)
		return nil, err
	}

	as.logger.Info("Campaign analytics generated successfully", "campaign_id", campaignID)
	return analytics, nil
}

// GenerateContentAnalytics generates analytics for specific content
func (as *AnalyticsService) GenerateContentAnalytics(ctx context.Context, contentIDs []uuid.UUID) (*ProcessedAnalytics, error) {
	as.logger.Info("Generating content analytics", "content_count", len(contentIDs))

	// Get content
	var contents []*entities.Content
	for _, contentID := range contentIDs {
		content, err := as.contentRepo.GetByID(ctx, contentID)
		if err != nil {
			as.logger.Error("Failed to get content", err, "content_id", contentID)
			continue
		}
		contents = append(contents, content)
	}

	if len(contents) == 0 {
		return nil, fmt.Errorf("no valid content found")
	}

	// Process content analytics
	analytics, err := as.dataProcessor.ProcessContentAnalytics(ctx, contents)
	if err != nil {
		as.logger.Error("Failed to process content analytics", err)
		return nil, err
	}

	as.logger.Info("Content analytics generated successfully", "content_count", len(contents))
	return analytics, nil
}

// GeneratePostAnalytics generates analytics for specific posts
func (as *AnalyticsService) GeneratePostAnalytics(ctx context.Context, postIDs []uuid.UUID) (*ProcessedAnalytics, error) {
	as.logger.Info("Generating post analytics", "post_count", len(postIDs))

	// Get posts
	var posts []*entities.Post
	for _, postID := range postIDs {
		post, err := as.postRepo.GetByID(ctx, postID)
		if err != nil {
			as.logger.Error("Failed to get post", err, "post_id", postID)
			continue
		}
		posts = append(posts, post)
	}

	if len(posts) == 0 {
		return nil, fmt.Errorf("no valid posts found")
	}

	// Process post analytics
	analytics, err := as.dataProcessor.ProcessPostAnalytics(ctx, posts)
	if err != nil {
		as.logger.Error("Failed to process post analytics", err)
		return nil, err
	}

	as.logger.Info("Post analytics generated successfully", "post_count", len(posts))
	return analytics, nil
}

// GenerateInsights generates insights from analytics data
func (as *AnalyticsService) GenerateInsights(ctx context.Context, analytics *ProcessedAnalytics) ([]*AnalyticsInsight, error) {
	as.logger.Info("Generating insights from analytics")

	insights, err := as.dataProcessor.GenerateInsights(ctx, analytics)
	if err != nil {
		as.logger.Error("Failed to generate insights", err)
		return nil, err
	}

	as.logger.Info("Insights generated successfully", "insight_count", len(insights))
	return insights, nil
}

// PredictContentPerformance predicts performance for content
func (as *AnalyticsService) PredictContentPerformance(ctx context.Context, contentID uuid.UUID) (*PerformancePrediction, error) {
	as.logger.Info("Predicting content performance", "content_id", contentID)

	// Get content
	content, err := as.contentRepo.GetByID(ctx, contentID)
	if err != nil {
		as.logger.Error("Failed to get content", err, "content_id", contentID)
		return nil, err
	}

	// Generate prediction
	prediction, err := as.dataProcessor.PredictPerformance(ctx, content)
	if err != nil {
		as.logger.Error("Failed to predict content performance", err, "content_id", contentID)
		return nil, err
	}

	as.logger.Info("Content performance predicted successfully", "content_id", contentID)
	return prediction, nil
}

// CalculateTrends calculates trends from metric data
func (as *AnalyticsService) CalculateTrends(ctx context.Context, metrics []*MetricPoint) (*TrendAnalysis, error) {
	as.logger.Info("Calculating trends", "metric_points", len(metrics))

	trends, err := as.dataProcessor.CalculateTrends(ctx, metrics)
	if err != nil {
		as.logger.Error("Failed to calculate trends", err)
		return nil, err
	}

	as.logger.Info("Trends calculated successfully")
	return trends, nil
}

// GetTopContent gets top performing content for a brand
func (as *AnalyticsService) GetTopContent(ctx context.Context, brandID uuid.UUID, timeRange TimeRange, limit int) ([]*entities.Content, error) {
	as.logger.Info("Getting top content", "brand_id", brandID, "limit", limit)

	// Get content for brand in time range
	filter := &repositories.ContentFilter{
		BrandIDs:      []uuid.UUID{brandID},
		CreatedAfter:  &timeRange.StartTime,
		CreatedBefore: &timeRange.EndTime,
		Limit:         limit * 2, // Get more to account for filtering
	}

	contents, err := as.contentRepo.ListByBrand(ctx, brandID, filter)
	if err != nil {
		as.logger.Error("Failed to get content for brand", err, "brand_id", brandID)
		return nil, err
	}

	// Calculate performance scores and sort
	type contentWithScore struct {
		content *entities.Content
		score   float64
	}

	var contentScores []contentWithScore
	for _, content := range contents {
		score := as.calculateContentPerformanceScore(content)
		contentScores = append(contentScores, contentWithScore{
			content: content,
			score:   score,
		})
	}

	// Sort by score descending
	sort.Slice(contentScores, func(i, j int) bool {
		return contentScores[i].score > contentScores[j].score
	})

	// Extract top content
	var topContent []*entities.Content
	for i, cs := range contentScores {
		if i >= limit {
			break
		}
		topContent = append(topContent, cs.content)
	}

	as.logger.Info("Top content retrieved", "brand_id", brandID, "count", len(topContent))
	return topContent, nil
}

// Helper methods

func (as *AnalyticsService) calculateContentPerformanceScore(content *entities.Content) float64 {
	// Simple scoring algorithm - can be enhanced with ML
	score := 0.0

	// Base score from engagement metrics if available
	if content.Analytics != nil {
		// Calculate total engagement from individual metrics
		totalEngagement := content.Analytics.Likes + content.Analytics.Comments + content.Analytics.Shares + content.Analytics.Saves
		
		if content.Analytics.Impressions > 0 {
			score += float64(totalEngagement) / float64(content.Analytics.Impressions) * 100
		}
		score += math.Log(float64(content.Analytics.Reach+1)) * 10
		score += float64(content.Analytics.Shares) * 5
		score += float64(content.Analytics.Comments) * 3
		score += float64(content.Analytics.Likes) * 1
		score += float64(content.Analytics.Clicks) * 2
	}

	// Penalty for older content
	daysSinceCreation := time.Since(content.CreatedAt).Hours() / 24
	if daysSinceCreation > 30 {
		score *= 0.9 // 10% penalty for content older than 30 days
	}

	return score
}

// NewContentGeneratedEvent creates a content generated event
func NewContentGeneratedEvent(content *entities.Content, createdBy uuid.UUID) DomainEvent {
	return &ContentGeneratedEvent{
		AggregateID: content.ID,
		ContentID:   content.ID,
		CreatedBy:   createdBy,
		Timestamp:   time.Now(),
		Version:     1,
	}
}

// ContentGeneratedEvent represents a content generated event
type ContentGeneratedEvent struct {
	AggregateID uuid.UUID `json:"aggregate_id"`
	ContentID   uuid.UUID `json:"content_id"`
	CreatedBy   uuid.UUID `json:"created_by"`
	Timestamp   time.Time `json:"timestamp"`
	Version     int       `json:"version"`
}

func (e *ContentGeneratedEvent) GetEventType() string        { return "content.generated" }
func (e *ContentGeneratedEvent) GetAggregateID() uuid.UUID   { return e.AggregateID }
func (e *ContentGeneratedEvent) GetTimestamp() time.Time     { return e.Timestamp }
func (e *ContentGeneratedEvent) GetVersion() int             { return e.Version }
func (e *ContentGeneratedEvent) GetEventData() map[string]interface{} {
	return map[string]interface{}{
		"content_id": e.ContentID,
		"created_by": e.CreatedBy,
	}
}