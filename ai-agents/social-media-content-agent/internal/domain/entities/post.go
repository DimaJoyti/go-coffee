package entities

import (
	"time"

	"github.com/google/uuid"
)

// Post represents a published social media post
type Post struct {
	ID                uuid.UUID              `json:"id" redis:"id"`
	ContentID         uuid.UUID              `json:"content_id" redis:"content_id"`
	Content           *Content               `json:"content,omitempty"`
	Platform          PlatformType           `json:"platform" redis:"platform"`
	PlatformPostID    string                 `json:"platform_post_id" redis:"platform_post_id"`
	Status            PostStatus             `json:"status" redis:"status"`
	Type              PostType               `json:"type" redis:"type"`
	Text              string                 `json:"text" redis:"text"`
	MediaURLs         []string               `json:"media_urls" redis:"media_urls"`
	Hashtags          []string               `json:"hashtags" redis:"hashtags"`
	Mentions          []string               `json:"mentions" redis:"mentions"`
	Link              string                 `json:"link" redis:"link"`
	Location          *PostLocation          `json:"location,omitempty"`
	ScheduledAt       *time.Time             `json:"scheduled_at,omitempty" redis:"scheduled_at"`
	PublishedAt       *time.Time             `json:"published_at,omitempty" redis:"published_at"`
	LastModifiedAt    *time.Time             `json:"last_modified_at,omitempty" redis:"last_modified_at"`
	ExpiresAt         *time.Time             `json:"expires_at,omitempty" redis:"expires_at"`
	Analytics         *PostAnalytics         `json:"analytics,omitempty"`
	Engagement        *PostEngagement        `json:"engagement,omitempty"`
	Comments          []*PostComment         `json:"comments,omitempty"`
	Interactions      []*PostInteraction     `json:"interactions,omitempty"`
	BoostSettings     *BoostSettings         `json:"boost_settings,omitempty"`
	TargetAudience    *TargetAudience        `json:"target_audience,omitempty"`
	ABTestVariant     *uuid.UUID             `json:"ab_test_variant,omitempty" redis:"ab_test_variant"`
	ParentPostID      *uuid.UUID             `json:"parent_post_id,omitempty" redis:"parent_post_id"`
	ThreadPosition    int                    `json:"thread_position" redis:"thread_position"`
	IsRepost          bool                   `json:"is_repost" redis:"is_repost"`
	OriginalPostID    *uuid.UUID             `json:"original_post_id,omitempty" redis:"original_post_id"`
	ErrorMessage      string                 `json:"error_message" redis:"error_message"`
	RetryCount        int                    `json:"retry_count" redis:"retry_count"`
	MaxRetries        int                    `json:"max_retries" redis:"max_retries"`
	CustomFields      map[string]interface{} `json:"custom_fields" redis:"custom_fields"`
	Metadata          map[string]interface{} `json:"metadata" redis:"metadata"`
	ExternalIDs       map[string]string      `json:"external_ids" redis:"external_ids"`
	CreatedAt         time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" redis:"updated_at"`
	CreatedBy         uuid.UUID              `json:"created_by" redis:"created_by"`
	UpdatedBy         uuid.UUID              `json:"updated_by" redis:"updated_by"`
	Version           int64                  `json:"version" redis:"version"`
}

// PostStatus defines the status of a post
type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusScheduled PostStatus = "scheduled"
	PostStatusPublishing PostStatus = "publishing"
	PostStatusPublished PostStatus = "published"
	PostStatusFailed    PostStatus = "failed"
	PostStatusDeleted   PostStatus = "deleted"
	PostStatusExpired   PostStatus = "expired"
	PostStatusArchived  PostStatus = "archived"
)

// PostType defines the type of post
type PostType string

const (
	PostTypeText      PostType = "text"
	PostTypeImage     PostType = "image"
	PostTypeVideo     PostType = "video"
	PostTypeCarousel  PostType = "carousel"
	PostTypeStory     PostType = "story"
	PostTypeReel      PostType = "reel"
	PostTypeLive      PostType = "live"
	PostTypePoll      PostType = "poll"
	PostTypeEvent     PostType = "event"
	PostTypeProduct   PostType = "product"
)

// PostLocation represents location information for a post
type PostLocation struct {
	Name      string  `json:"name" redis:"name"`
	Address   string  `json:"address" redis:"address"`
	City      string  `json:"city" redis:"city"`
	Country   string  `json:"country" redis:"country"`
	Latitude  float64 `json:"latitude" redis:"latitude"`
	Longitude float64 `json:"longitude" redis:"longitude"`
	PlaceID   string  `json:"place_id" redis:"place_id"`
}

// PostAnalytics represents analytics data for a post
type PostAnalytics struct {
	Impressions        int64                    `json:"impressions"`
	Reach              int64                    `json:"reach"`
	Engagement         int64                    `json:"engagement"`
	Likes              int64                    `json:"likes"`
	Comments           int64                    `json:"comments"`
	Shares             int64                    `json:"shares"`
	Saves              int64                    `json:"saves"`
	Clicks             int64                    `json:"clicks"`
	VideoViews         int64                    `json:"video_views"`
	VideoCompletions   int64                    `json:"video_completions"`
	ProfileVisits      int64                    `json:"profile_visits"`
	WebsiteClicks      int64                    `json:"website_clicks"`
	EmailClicks        int64                    `json:"email_clicks"`
	PhoneClicks        int64                    `json:"phone_clicks"`
	DirectionClicks    int64                    `json:"direction_clicks"`
	EngagementRate     float64                  `json:"engagement_rate"`
	CTR                float64                  `json:"ctr"`
	VideoViewRate      float64                  `json:"video_view_rate"`
	CompletionRate     float64                  `json:"completion_rate"`
	Demographics       *AnalyticsDemographics   `json:"demographics,omitempty"`
	TopCountries       []CountryMetric          `json:"top_countries,omitempty"`
	TopCities          []CityMetric             `json:"top_cities,omitempty"`
	DeviceBreakdown    map[string]int64         `json:"device_breakdown,omitempty"`
	TimeOfDay          map[string]int64         `json:"time_of_day,omitempty"`
	DayOfWeek          map[string]int64         `json:"day_of_week,omitempty"`
	ReferralSources    map[string]int64         `json:"referral_sources,omitempty"`
	HashtagPerformance map[string]*HashtagMetric `json:"hashtag_performance,omitempty"`
	LastUpdated        time.Time                `json:"last_updated"`
	SyncedAt           time.Time                `json:"synced_at"`
}

// AnalyticsDemographics represents demographic breakdown
type AnalyticsDemographics struct {
	AgeGroups map[string]float64 `json:"age_groups,omitempty"`
	Genders   map[string]float64 `json:"genders,omitempty"`
	Interests map[string]float64 `json:"interests,omitempty"`
}

// CountryMetric represents country-specific metrics
type CountryMetric struct {
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Impressions int64   `json:"impressions"`
	Engagement  int64   `json:"engagement"`
	Percentage  float64 `json:"percentage"`
}

// CityMetric represents city-specific metrics
type CityMetric struct {
	City        string  `json:"city"`
	Country     string  `json:"country"`
	Impressions int64   `json:"impressions"`
	Engagement  int64   `json:"engagement"`
	Percentage  float64 `json:"percentage"`
}

// HashtagMetric represents hashtag performance metrics
type HashtagMetric struct {
	Hashtag     string  `json:"hashtag"`
	Impressions int64   `json:"impressions"`
	Engagement  int64   `json:"engagement"`
	Reach       int64   `json:"reach"`
	Posts       int64   `json:"posts"`
	Trending    bool    `json:"trending"`
	Score       float64 `json:"score"`
}

// PostEngagement represents engagement data for a post
type PostEngagement struct {
	TotalEngagement    int64                    `json:"total_engagement"`
	LikeCount          int64                    `json:"like_count"`
	CommentCount       int64                    `json:"comment_count"`
	ShareCount         int64                    `json:"share_count"`
	SaveCount          int64                    `json:"save_count"`
	ReactionBreakdown  map[string]int64         `json:"reaction_breakdown,omitempty"`
	TopComments        []*PostComment           `json:"top_comments,omitempty"`
	InfluencerMentions []*InfluencerMention     `json:"influencer_mentions,omitempty"`
	UserGeneratedContent []*UGCReference        `json:"user_generated_content,omitempty"`
	SentimentAnalysis  *SentimentAnalysis       `json:"sentiment_analysis,omitempty"`
	LastUpdated        time.Time                `json:"last_updated"`
}

// PostComment represents a comment on a post
type PostComment struct {
	ID               string             `json:"id"`
	PostID           uuid.UUID          `json:"post_id"`
	Platform         PlatformType       `json:"platform"`
	PlatformCommentID string            `json:"platform_comment_id"`
	AuthorID         string             `json:"author_id"`
	AuthorUsername   string             `json:"author_username"`
	AuthorName       string             `json:"author_name"`
	AuthorAvatar     string             `json:"author_avatar"`
	Text             string             `json:"text"`
	Likes            int64              `json:"likes"`
	Replies          int64              `json:"replies"`
	ParentCommentID  *string            `json:"parent_comment_id,omitempty"`
	Sentiment        ContentSentiment   `json:"sentiment"`
	IsVerified       bool               `json:"is_verified"`
	IsInfluencer     bool               `json:"is_influencer"`
	RequiresResponse bool               `json:"requires_response"`
	IsResponded      bool               `json:"is_responded"`
	ResponseID       *string            `json:"response_id,omitempty"`
	Tags             []string           `json:"tags"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

// PostInteraction represents an interaction with a post
type PostInteraction struct {
	ID           uuid.UUID        `json:"id"`
	PostID       uuid.UUID        `json:"post_id"`
	Platform     PlatformType     `json:"platform"`
	Type         InteractionType  `json:"type"`
	UserID       string           `json:"user_id"`
	Username     string           `json:"username"`
	UserName     string           `json:"user_name"`
	UserAvatar   string           `json:"user_avatar"`
	IsVerified   bool             `json:"is_verified"`
	IsInfluencer bool             `json:"is_influencer"`
	FollowerCount int64           `json:"follower_count"`
	Timestamp    time.Time        `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// InteractionType defines the type of interaction
type InteractionType string

const (
	InteractionTypeLike    InteractionType = "like"
	InteractionTypeComment InteractionType = "comment"
	InteractionTypeShare   InteractionType = "share"
	InteractionTypeSave    InteractionType = "save"
	InteractionTypeClick   InteractionType = "click"
	InteractionTypeView    InteractionType = "view"
	InteractionTypeFollow  InteractionType = "follow"
	InteractionTypeMention InteractionType = "mention"
)

// InfluencerMention represents a mention by an influencer
type InfluencerMention struct {
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	Name          string    `json:"name"`
	FollowerCount int64     `json:"follower_count"`
	VerifiedStatus bool     `json:"verified_status"`
	InfluencerTier string   `json:"influencer_tier"`
	Engagement    int64     `json:"engagement"`
	Reach         int64     `json:"reach"`
	MentionType   string    `json:"mention_type"`
	Content       string    `json:"content"`
	Timestamp     time.Time `json:"timestamp"`
}

// UGCReference represents user-generated content reference
type UGCReference struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	ContentType string    `json:"content_type"`
	ContentURL  string    `json:"content_url"`
	Caption     string    `json:"caption"`
	Hashtags    []string  `json:"hashtags"`
	Mentions    []string  `json:"mentions"`
	Engagement  int64     `json:"engagement"`
	Quality     float64   `json:"quality"`
	Timestamp   time.Time `json:"timestamp"`
}

// SentimentAnalysis represents sentiment analysis of post engagement
type SentimentAnalysis struct {
	OverallSentiment ContentSentiment       `json:"overall_sentiment"`
	SentimentScore   float64                `json:"sentiment_score"`
	Confidence       float64                `json:"confidence"`
	SentimentBreakdown map[ContentSentiment]float64 `json:"sentiment_breakdown"`
	KeyPhrases       []string               `json:"key_phrases"`
	Emotions         map[string]float64     `json:"emotions"`
	Topics           []string               `json:"topics"`
	LanguageDetected string                 `json:"language_detected"`
	LastAnalyzed     time.Time              `json:"last_analyzed"`
}

// BoostSettings represents post boost/promotion settings
type BoostSettings struct {
	IsPromoted      bool                   `json:"is_promoted"`
	Budget          Money                  `json:"budget"`
	Duration        int                    `json:"duration"` // Days
	Objective       string                 `json:"objective"`
	TargetAudience  *TargetAudience        `json:"target_audience,omitempty"`
	Placements      []string               `json:"placements"`
	BidStrategy     string                 `json:"bid_strategy"`
	OptimizationGoal string                `json:"optimization_goal"`
	CallToAction    string                 `json:"call_to_action"`
	StartDate       time.Time              `json:"start_date"`
	EndDate         time.Time              `json:"end_date"`
	Status          string                 `json:"status"`
	Performance     *PromotionPerformance  `json:"performance,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// PromotionPerformance represents promotion performance metrics
type PromotionPerformance struct {
	Spend           Money   `json:"spend"`
	Impressions     int64   `json:"impressions"`
	Reach           int64   `json:"reach"`
	Clicks          int64   `json:"clicks"`
	Conversions     int64   `json:"conversions"`
	CTR             float64 `json:"ctr"`
	CPC             float64 `json:"cpc"`
	CPM             float64 `json:"cpm"`
	ConversionRate  float64 `json:"conversion_rate"`
	CostPerConversion float64 `json:"cost_per_conversion"`
	ROAS            float64 `json:"roas"`
	LastUpdated     time.Time `json:"last_updated"`
}

// NewPost creates a new post with default values
func NewPost(contentID uuid.UUID, platform PlatformType, text string, createdBy uuid.UUID) *Post {
	now := time.Now()
	return &Post{
		ID:           uuid.New(),
		ContentID:    contentID,
		Platform:     platform,
		Status:       PostStatusDraft,
		Type:         PostTypeText,
		Text:         text,
		MediaURLs:    []string{},
		Hashtags:     []string{},
		Mentions:     []string{},
		RetryCount:   0,
		MaxRetries:   3,
		CustomFields: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
		ExternalIDs:  make(map[string]string),
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    createdBy,
		UpdatedBy:    createdBy,
		Version:      1,
	}
}

// UpdateStatus updates the post status
func (p *Post) UpdateStatus(newStatus PostStatus, updatedBy uuid.UUID) {
	p.Status = newStatus
	p.UpdatedBy = updatedBy
	p.UpdatedAt = time.Now()
	p.Version++

	// Handle status-specific logic
	switch newStatus {
	case PostStatusPublished:
		if p.PublishedAt == nil {
			now := time.Now()
			p.PublishedAt = &now
		}
	case PostStatusFailed:
		p.RetryCount++
	}
}

// SetPlatformPostID sets the platform-specific post ID
func (p *Post) SetPlatformPostID(platformPostID string, updatedBy uuid.UUID) {
	p.PlatformPostID = platformPostID
	p.UpdatedBy = updatedBy
	p.UpdatedAt = time.Now()
	p.Version++
}

// AddMediaURL adds a media URL to the post
func (p *Post) AddMediaURL(url string) {
	p.MediaURLs = append(p.MediaURLs, url)
	p.UpdatedAt = time.Now()
	p.Version++
}

// AddHashtag adds a hashtag to the post
func (p *Post) AddHashtag(hashtag string) {
	// Remove # if present
	if len(hashtag) > 0 && hashtag[0] == '#' {
		hashtag = hashtag[1:]
	}
	
	// Check if hashtag already exists
	for _, existing := range p.Hashtags {
		if existing == hashtag {
			return
		}
	}
	
	p.Hashtags = append(p.Hashtags, hashtag)
	p.UpdatedAt = time.Now()
	p.Version++
}

// AddMention adds a mention to the post
func (p *Post) AddMention(mention string) {
	// Remove @ if present
	if len(mention) > 0 && mention[0] == '@' {
		mention = mention[1:]
	}
	
	// Check if mention already exists
	for _, existing := range p.Mentions {
		if existing == mention {
			return
		}
	}
	
	p.Mentions = append(p.Mentions, mention)
	p.UpdatedAt = time.Now()
	p.Version++
}

// IsScheduled checks if the post is scheduled for future publishing
func (p *Post) IsScheduled() bool {
	return p.Status == PostStatusScheduled && p.ScheduledAt != nil && p.ScheduledAt.After(time.Now())
}

// IsPublished checks if the post is published
func (p *Post) IsPublished() bool {
	return p.Status == PostStatusPublished && p.PublishedAt != nil
}

// IsExpired checks if the post has expired
func (p *Post) IsExpired() bool {
	return p.ExpiresAt != nil && p.ExpiresAt.Before(time.Now())
}

// CanRetry checks if the post can be retried
func (p *Post) CanRetry() bool {
	return p.Status == PostStatusFailed && p.RetryCount < p.MaxRetries
}

// GetEngagementRate calculates the engagement rate
func (p *Post) GetEngagementRate() float64 {
	if p.Analytics == nil || p.Analytics.Impressions == 0 {
		return 0
	}
	return float64(p.Analytics.Engagement) / float64(p.Analytics.Impressions) * 100
}

// GetCTR calculates the click-through rate
func (p *Post) GetCTR() float64 {
	if p.Analytics == nil || p.Analytics.Impressions == 0 {
		return 0
	}
	return float64(p.Analytics.Clicks) / float64(p.Analytics.Impressions) * 100
}

// UpdateAnalytics updates the post analytics
func (p *Post) UpdateAnalytics(analytics *PostAnalytics) {
	p.Analytics = analytics
	p.Analytics.LastUpdated = time.Now()
	p.UpdatedAt = time.Now()
	p.Version++
}

// UpdateEngagement updates the post engagement
func (p *Post) UpdateEngagement(engagement *PostEngagement) {
	p.Engagement = engagement
	p.Engagement.LastUpdated = time.Now()
	p.UpdatedAt = time.Now()
	p.Version++
}

// AddComment adds a comment to the post
func (p *Post) AddComment(comment *PostComment) {
	comment.PostID = p.ID
	p.Comments = append(p.Comments, comment)
	p.UpdatedAt = time.Now()
	p.Version++
}

// AddInteraction adds an interaction to the post
func (p *Post) AddInteraction(interaction *PostInteraction) {
	interaction.PostID = p.ID
	p.Interactions = append(p.Interactions, interaction)
	p.UpdatedAt = time.Now()
	p.Version++
}

// SetError sets an error message for the post
func (p *Post) SetError(errorMessage string, updatedBy uuid.UUID) {
	p.ErrorMessage = errorMessage
	p.Status = PostStatusFailed
	p.RetryCount++
	p.UpdatedBy = updatedBy
	p.UpdatedAt = time.Now()
	p.Version++
}

// ClearError clears the error message
func (p *Post) ClearError(updatedBy uuid.UUID) {
	p.ErrorMessage = ""
	p.UpdatedBy = updatedBy
	p.UpdatedAt = time.Now()
	p.Version++
}
