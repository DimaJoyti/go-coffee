package entities

import (
	"time"

	"github.com/google/uuid"
)

// Content represents a comprehensive social media content entity
type Content struct {
	ID                uuid.UUID              `json:"id" redis:"id"`
	Title             string                 `json:"title" redis:"title"`
	Body              string                 `json:"body" redis:"body"`
	Type              ContentType            `json:"type" redis:"type"`
	Format            ContentFormat          `json:"format" redis:"format"`
	Status            ContentStatus          `json:"status" redis:"status"`
	Priority          ContentPriority        `json:"priority" redis:"priority"`
	Category          ContentCategory        `json:"category" redis:"category"`
	BrandID           uuid.UUID              `json:"brand_id" redis:"brand_id"`
	Brand             *Brand                 `json:"brand,omitempty"`
	CampaignID        *uuid.UUID             `json:"campaign_id,omitempty" redis:"campaign_id"`
	Campaign          *Campaign              `json:"campaign,omitempty"`
	CreatorID         uuid.UUID              `json:"creator_id" redis:"creator_id"`
	Creator           *User                  `json:"creator,omitempty"`
	ApproverID        *uuid.UUID             `json:"approver_id,omitempty" redis:"approver_id"`
	Approver          *User                  `json:"approver,omitempty"`
	Platforms         []PlatformType         `json:"platforms" redis:"platforms"`
	Posts             []*Post                `json:"posts,omitempty"`
	MediaAssets       []*MediaAsset          `json:"media_assets,omitempty"`
	Hashtags          []string               `json:"hashtags" redis:"hashtags"`
	Mentions          []string               `json:"mentions" redis:"mentions"`
	Tags              []string               `json:"tags" redis:"tags"`
	Keywords          []string               `json:"keywords" redis:"keywords"`
	TargetAudience    *TargetAudience        `json:"target_audience,omitempty"`
	Tone              ContentTone            `json:"tone" redis:"tone"`
	Language          string                 `json:"language" redis:"language"`
	Sentiment         ContentSentiment       `json:"sentiment" redis:"sentiment"`
	AIGenerated       bool                   `json:"ai_generated" redis:"ai_generated"`
	AIPrompt          string                 `json:"ai_prompt" redis:"ai_prompt"`
	AIModel           string                 `json:"ai_model" redis:"ai_model"`
	Variations        []*ContentVariation    `json:"variations,omitempty"`
	ScheduledAt       *time.Time             `json:"scheduled_at,omitempty" redis:"scheduled_at"`
	PublishedAt       *time.Time             `json:"published_at,omitempty" redis:"published_at"`
	ExpiresAt         *time.Time             `json:"expires_at,omitempty" redis:"expires_at"`
	Analytics         *ContentAnalytics      `json:"analytics,omitempty"`
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

// ContentType defines the type of content
type ContentType string

const (
	ContentTypePost        ContentType = "post"
	ContentTypeStory       ContentType = "story"
	ContentTypeReel        ContentType = "reel"
	ContentTypeVideo       ContentType = "video"
	ContentTypeImage       ContentType = "image"
	ContentTypeCarousel    ContentType = "carousel"
	ContentTypePoll        ContentType = "poll"
	ContentTypeEvent       ContentType = "event"
	ContentTypePromotion   ContentType = "promotion"
	ContentTypeAnnouncement ContentType = "announcement"
	ContentTypeBlog        ContentType = "blog"
	ContentTypeNewsletter  ContentType = "newsletter"
)

// ContentFormat defines the format of content
type ContentFormat string

const (
	FormatText     ContentFormat = "text"
	FormatImage    ContentFormat = "image"
	FormatVideo    ContentFormat = "video"
	FormatAudio    ContentFormat = "audio"
	FormatGIF      ContentFormat = "gif"
	FormatCarousel ContentFormat = "carousel"
	FormatLive     ContentFormat = "live"
	FormatStory    ContentFormat = "story"
)

// ContentStatus defines the status of content
type ContentStatus string

const (
	StatusDraft      ContentStatus = "draft"
	StatusReview     ContentStatus = "review"
	StatusApproved   ContentStatus = "approved"
	StatusRejected   ContentStatus = "rejected"
	StatusScheduled  ContentStatus = "scheduled"
	StatusPublished  ContentStatus = "published"
	StatusFailed     ContentStatus = "failed"
	StatusArchived   ContentStatus = "archived"
	StatusExpired    ContentStatus = "expired"
)

// ContentPriority defines the priority of content
type ContentPriority string

const (
	PriorityLow      ContentPriority = "low"
	PriorityMedium   ContentPriority = "medium"
	PriorityHigh     ContentPriority = "high"
	PriorityCritical ContentPriority = "critical"
	PriorityUrgent   ContentPriority = "urgent"
)

// ContentCategory defines the category of content
type ContentCategory string

const (
	CategoryMarketing    ContentCategory = "marketing"
	CategoryPromotion    ContentCategory = "promotion"
	CategoryEducational  ContentCategory = "educational"
	CategoryEntertainment ContentCategory = "entertainment"
	CategoryNews         ContentCategory = "news"
	CategoryProduct      ContentCategory = "product"
	CategoryBehindScenes ContentCategory = "behind_scenes"
	CategoryUserGenerated ContentCategory = "user_generated"
	CategorySeasonal     ContentCategory = "seasonal"
	CategoryEvent        ContentCategory = "event"
)

// ContentTone defines the tone of content
type ContentTone string

const (
	ToneFriendly     ContentTone = "friendly"
	ToneProfessional ContentTone = "professional"
	ToneCasual       ContentTone = "casual"
	ToneExcited      ContentTone = "excited"
	ToneInformative  ContentTone = "informative"
	ToneHumorous     ContentTone = "humorous"
	ToneInspiring    ContentTone = "inspiring"
	ToneUrgent       ContentTone = "urgent"
)

// ContentSentiment defines the sentiment of content
type ContentSentiment string

const (
	SentimentPositive ContentSentiment = "positive"
	SentimentNeutral  ContentSentiment = "neutral"
	SentimentNegative ContentSentiment = "negative"
)

// PlatformType defines social media platforms
type PlatformType string

const (
	PlatformInstagram PlatformType = "instagram"
	PlatformFacebook  PlatformType = "facebook"
	PlatformTwitter   PlatformType = "twitter"
	PlatformLinkedIn  PlatformType = "linkedin"
	PlatformTikTok    PlatformType = "tiktok"
	PlatformYouTube   PlatformType = "youtube"
	PlatformPinterest PlatformType = "pinterest"
	PlatformSnapchat  PlatformType = "snapchat"
)

// MediaAsset represents a media asset attached to content
type MediaAsset struct {
	ID          uuid.UUID   `json:"id" redis:"id"`
	ContentID   uuid.UUID   `json:"content_id" redis:"content_id"`
	Type        MediaType   `json:"type" redis:"type"`
	URL         string      `json:"url" redis:"url"`
	FileName    string      `json:"file_name" redis:"file_name"`
	FileSize    int64       `json:"file_size" redis:"file_size"`
	MimeType    string      `json:"mime_type" redis:"mime_type"`
	Width       int         `json:"width" redis:"width"`
	Height      int         `json:"height" redis:"height"`
	Duration    int         `json:"duration" redis:"duration"` // For videos/audio in seconds
	AltText     string      `json:"alt_text" redis:"alt_text"`
	Caption     string      `json:"caption" redis:"caption"`
	AIGenerated bool        `json:"ai_generated" redis:"ai_generated"`
	AIPrompt    string      `json:"ai_prompt" redis:"ai_prompt"`
	Order       int         `json:"order" redis:"order"`
	IsActive    bool        `json:"is_active" redis:"is_active"`
	CreatedAt   time.Time   `json:"created_at" redis:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" redis:"updated_at"`
}

// MediaType defines the type of media asset
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
	MediaTypeAudio MediaType = "audio"
	MediaTypeGIF   MediaType = "gif"
	MediaTypeDocument MediaType = "document"
)

// ContentVariation represents a variation of content for A/B testing
type ContentVariation struct {
	ID          uuid.UUID `json:"id" redis:"id"`
	ContentID   uuid.UUID `json:"content_id" redis:"content_id"`
	Name        string    `json:"name" redis:"name"`
	Body        string    `json:"body" redis:"body"`
	Hashtags    []string  `json:"hashtags" redis:"hashtags"`
	MediaAssets []*MediaAsset `json:"media_assets,omitempty"`
	Weight      float64   `json:"weight" redis:"weight"` // For A/B testing distribution
	Performance *VariationPerformance `json:"performance,omitempty"`
	IsActive    bool      `json:"is_active" redis:"is_active"`
	CreatedAt   time.Time `json:"created_at" redis:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" redis:"updated_at"`
}

// VariationPerformance represents performance metrics for a content variation
type VariationPerformance struct {
	Impressions   int64   `json:"impressions"`
	Clicks        int64   `json:"clicks"`
	Likes         int64   `json:"likes"`
	Shares        int64   `json:"shares"`
	Comments      int64   `json:"comments"`
	CTR           float64 `json:"ctr"`           // Click-through rate
	EngagementRate float64 `json:"engagement_rate"`
	ConversionRate float64 `json:"conversion_rate"`
	LastUpdated   time.Time `json:"last_updated"`
}

// TargetAudience represents the target audience for content
type TargetAudience struct {
	Demographics *Demographics `json:"demographics,omitempty"`
	Interests    []string      `json:"interests,omitempty"`
	Behaviors    []string      `json:"behaviors,omitempty"`
	Locations    []string      `json:"locations,omitempty"`
	Languages    []string      `json:"languages,omitempty"`
	CustomAudiences []string   `json:"custom_audiences,omitempty"`
}

// Demographics represents demographic information
type Demographics struct {
	AgeMin    int      `json:"age_min,omitempty"`
	AgeMax    int      `json:"age_max,omitempty"`
	Genders   []string `json:"genders,omitempty"`
	Education []string `json:"education,omitempty"`
	Income    []string `json:"income,omitempty"`
}

// ContentAnalytics represents analytics data for content
type ContentAnalytics struct {
	Impressions     int64     `json:"impressions"`
	Reach           int64     `json:"reach"`
	Clicks          int64     `json:"clicks"`
	Likes           int64     `json:"likes"`
	Shares          int64     `json:"shares"`
	Comments        int64     `json:"comments"`
	Saves           int64     `json:"saves"`
	CTR             float64   `json:"ctr"`
	EngagementRate  float64   `json:"engagement_rate"`
	ConversionRate  float64   `json:"conversion_rate"`
	CostPerClick    float64   `json:"cost_per_click"`
	CostPerImpression float64 `json:"cost_per_impression"`
	ROI             float64   `json:"roi"`
	Sentiment       ContentSentiment `json:"sentiment"`
	LastUpdated     time.Time `json:"last_updated"`
	PlatformMetrics map[PlatformType]*PlatformMetrics `json:"platform_metrics,omitempty"`
}

// PlatformMetrics represents platform-specific metrics
type PlatformMetrics struct {
	Platform       PlatformType `json:"platform"`
	Impressions    int64        `json:"impressions"`
	Reach          int64        `json:"reach"`
	Engagement     int64        `json:"engagement"`
	Clicks         int64        `json:"clicks"`
	Shares         int64        `json:"shares"`
	Comments       int64        `json:"comments"`
	Likes          int64        `json:"likes"`
	EngagementRate float64      `json:"engagement_rate"`
	LastUpdated    time.Time    `json:"last_updated"`
}

// ComplianceInfo represents compliance and legal information
type ComplianceInfo struct {
	IsCompliant     bool      `json:"is_compliant"`
	ReviewedBy      *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	ComplianceNotes string    `json:"compliance_notes"`
	LegalApproval   bool      `json:"legal_approval"`
	RequiresDisclaimer bool   `json:"requires_disclaimer"`
	Disclaimer      string    `json:"disclaimer"`
	Regulations     []string  `json:"regulations"`
	Warnings        []string  `json:"warnings"`
}

// NewContent creates a new content with default values
func NewContent(title, body string, contentType ContentType, brandID, creatorID uuid.UUID) *Content {
	now := time.Now()
	return &Content{
		ID:           uuid.New(),
		Title:        title,
		Body:         body,
		Type:         contentType,
		Format:       FormatText,
		Status:       StatusDraft,
		Priority:     PriorityMedium,
		BrandID:      brandID,
		CreatorID:    creatorID,
		Platforms:    []PlatformType{},
		Hashtags:     []string{},
		Mentions:     []string{},
		Tags:         []string{},
		Keywords:     []string{},
		Tone:         ToneFriendly,
		Language:     "en",
		Sentiment:    SentimentNeutral,
		AIGenerated:  false,
		CustomFields: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
		ExternalIDs:  make(map[string]string),
		IsTemplate:   false,
		IsArchived:   false,
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    creatorID,
		UpdatedBy:    creatorID,
		Version:      1,
	}
}

// UpdateStatus updates the content status
func (c *Content) UpdateStatus(newStatus ContentStatus, updatedBy uuid.UUID) {
	c.Status = newStatus
	c.UpdatedBy = updatedBy
	c.UpdatedAt = time.Now()
	c.Version++

	// Handle status-specific logic
	switch newStatus {
	case StatusPublished:
		if c.PublishedAt == nil {
			now := time.Now()
			c.PublishedAt = &now
		}
	case StatusScheduled:
		// Ensure scheduled time is set
		if c.ScheduledAt == nil {
			future := time.Now().Add(1 * time.Hour)
			c.ScheduledAt = &future
		}
	}
}

// AddMediaAsset adds a media asset to the content
func (c *Content) AddMediaAsset(asset *MediaAsset) {
	asset.ContentID = c.ID
	c.MediaAssets = append(c.MediaAssets, asset)
	c.UpdatedAt = time.Now()
	c.Version++
}

// AddVariation adds a content variation for A/B testing
func (c *Content) AddVariation(variation *ContentVariation) {
	variation.ContentID = c.ID
	c.Variations = append(c.Variations, variation)
	c.UpdatedAt = time.Now()
	c.Version++
}

// AddHashtag adds a hashtag to the content
func (c *Content) AddHashtag(hashtag string) {
	// Remove # if present and add it
	if hashtag[0] == '#' {
		hashtag = hashtag[1:]
	}
	
	// Check if hashtag already exists
	for _, existing := range c.Hashtags {
		if existing == hashtag {
			return
		}
	}
	
	c.Hashtags = append(c.Hashtags, hashtag)
	c.UpdatedAt = time.Now()
	c.Version++
}

// AddPlatform adds a platform to the content
func (c *Content) AddPlatform(platform PlatformType) {
	// Check if platform already exists
	for _, existing := range c.Platforms {
		if existing == platform {
			return
		}
	}
	
	c.Platforms = append(c.Platforms, platform)
	c.UpdatedAt = time.Now()
	c.Version++
}

// IsScheduled checks if the content is scheduled for future publishing
func (c *Content) IsScheduled() bool {
	return c.Status == StatusScheduled && c.ScheduledAt != nil && c.ScheduledAt.After(time.Now())
}

// IsPublished checks if the content is published
func (c *Content) IsPublished() bool {
	return c.Status == StatusPublished && c.PublishedAt != nil
}

// IsExpired checks if the content has expired
func (c *Content) IsExpired() bool {
	return c.ExpiresAt != nil && c.ExpiresAt.Before(time.Now())
}

// CanBePublished checks if the content can be published
func (c *Content) CanBePublished() bool {
	return c.Status == StatusApproved && len(c.Platforms) > 0 && !c.IsArchived
}

// GetEngagementRate calculates the engagement rate
func (c *Content) GetEngagementRate() float64 {
	if c.Analytics == nil || c.Analytics.Impressions == 0 {
		return 0
	}
	
	totalEngagement := c.Analytics.Likes + c.Analytics.Comments + c.Analytics.Shares + c.Analytics.Saves
	return float64(totalEngagement) / float64(c.Analytics.Impressions) * 100
}

// Archive archives the content
func (c *Content) Archive(archivedBy uuid.UUID) {
	c.IsArchived = true
	c.Status = StatusArchived
	c.UpdatedBy = archivedBy
	c.UpdatedAt = time.Now()
	c.Version++
}

// Unarchive unarchives the content
func (c *Content) Unarchive(unarchivedBy uuid.UUID) {
	c.IsArchived = false
	c.Status = StatusDraft
	c.UpdatedBy = unarchivedBy
	c.UpdatedAt = time.Now()
	c.Version++
}

// Domain errors
var (
	ErrContentNotFound       = NewDomainError("CONTENT_NOT_FOUND", "Content not found")
	ErrInvalidStatus         = NewDomainError("INVALID_STATUS", "Invalid content status")
	ErrContentNotApproved    = NewDomainError("CONTENT_NOT_APPROVED", "Content is not approved for publishing")
	ErrContentExpired        = NewDomainError("CONTENT_EXPIRED", "Content has expired")
	ErrInvalidPlatform       = NewDomainError("INVALID_PLATFORM", "Invalid platform specified")
	ErrContentArchived       = NewDomainError("CONTENT_ARCHIVED", "Cannot modify archived content")
	ErrInvalidScheduleTime   = NewDomainError("INVALID_SCHEDULE_TIME", "Schedule time must be in the future")
)

// DomainError represents a domain-specific error
type DomainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *DomainError) Error() string {
	return e.Message
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}
