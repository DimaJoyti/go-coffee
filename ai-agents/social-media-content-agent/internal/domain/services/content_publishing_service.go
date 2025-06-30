package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/social-media-content-agent/internal/domain/entities"
	"go-coffee-ai-agents/social-media-content-agent/internal/domain/repositories"
)

// ContentPublishingService provides content publishing functionality
type ContentPublishingService struct {
	postRepo         repositories.PostRepository
	contentRepo      repositories.ContentRepository
	brandRepo        repositories.BrandRepository
	platformClients  map[entities.PlatformType]PlatformClient
	eventPublisher   EventPublisher
	logger           Logger
}

// PlatformClient defines the interface for platform-specific publishing
type PlatformClient interface {
	PublishPost(ctx context.Context, post *entities.Post, credentials *PlatformCredentials) (*PublishResult, error)
	UpdatePost(ctx context.Context, postID string, content string, credentials *PlatformCredentials) (*PublishResult, error)
	DeletePost(ctx context.Context, postID string, credentials *PlatformCredentials) error
	GetPostMetrics(ctx context.Context, postID string, credentials *PlatformCredentials) (*PostMetrics, error)
	ValidateCredentials(ctx context.Context, credentials *PlatformCredentials) error
	GetPlatformLimits() *PlatformLimits
}

// PlatformCredentials represents platform-specific authentication credentials
type PlatformCredentials struct {
	Platform    entities.PlatformType  `json:"platform"`
	AccessToken string                 `json:"access_token"`
	Secret      string                 `json:"secret,omitempty"`
	APIKey      string                 `json:"api_key,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Scope       []string               `json:"scope,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// PublishResult represents the result of publishing a post
type PublishResult struct {
	PlatformPostID string                 `json:"platform_post_id"`
	PublishedAt    time.Time              `json:"published_at"`
	URL            string                 `json:"url,omitempty"`
	Metrics        *PostMetrics           `json:"metrics,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Success        bool                   `json:"success"`
	Error          string                 `json:"error,omitempty"`
}

// PostMetrics represents post performance metrics
type PostMetrics struct {
	Impressions   int64     `json:"impressions"`
	Reach         int64     `json:"reach"`
	Engagement    int64     `json:"engagement"`
	Likes         int64     `json:"likes"`
	Comments      int64     `json:"comments"`
	Shares        int64     `json:"shares"`
	Saves         int64     `json:"saves"`
	Clicks        int64     `json:"clicks"`
	VideoViews    int64     `json:"video_views,omitempty"`
	FetchedAt     time.Time `json:"fetched_at"`
}

// PlatformLimits represents platform-specific limits
type PlatformLimits struct {
	MaxTextLength    int      `json:"max_text_length"`
	MaxHashtags      int      `json:"max_hashtags"`
	MaxMentions      int      `json:"max_mentions"`
	MaxImages        int      `json:"max_images"`
	MaxVideos        int      `json:"max_videos"`
	SupportedFormats []string `json:"supported_formats"`
	RateLimit        int      `json:"rate_limit"` // posts per hour
}

// NewContentPublishingService creates a new content publishing service
func NewContentPublishingService(
	postRepo repositories.PostRepository,
	contentRepo repositories.ContentRepository,
	brandRepo repositories.BrandRepository,
	platformClients map[entities.PlatformType]PlatformClient,
	eventPublisher EventPublisher,
	logger Logger,
) *ContentPublishingService {
	return &ContentPublishingService{
		postRepo:        postRepo,
		contentRepo:     contentRepo,
		brandRepo:       brandRepo,
		platformClients: platformClients,
		eventPublisher:  eventPublisher,
		logger:          logger,
	}
}

// PublishPost publishes a post to its target platform
func (cps *ContentPublishingService) PublishPost(ctx context.Context, postID uuid.UUID, publishedBy uuid.UUID) (*PublishResult, error) {
	cps.logger.Info("Publishing post", "post_id", postID, "published_by", publishedBy)

	// Get post
	post, err := cps.postRepo.GetByID(ctx, postID)
	if err != nil {
		cps.logger.Error("Failed to get post", err, "post_id", postID)
		return nil, err
	}

	// Validate post can be published
	if post.Status != entities.PostStatusScheduled && post.Status != entities.PostStatusDraft {
		return nil, fmt.Errorf("post cannot be published: current status=%s", post.Status)
	}

	// Get brand for platform credentials
	brand, err := cps.brandRepo.GetByID(ctx, post.ContentID) // Assuming content has brand info
	if err != nil {
		cps.logger.Error("Failed to get brand", err, "post_id", postID)
		return nil, err
	}

	// Get platform client
	client, exists := cps.platformClients[post.Platform]
	if !exists {
		return nil, fmt.Errorf("no client available for platform: %s", post.Platform)
	}

	// Get platform credentials for brand
	credentials, err := cps.getPlatformCredentials(brand, post.Platform)
	if err != nil {
		cps.logger.Error("Failed to get platform credentials", err, "brand_id", brand.ID, "platform", post.Platform)
		return nil, err
	}

	// Validate content against platform limits
	if err := cps.validatePostContent(post, client.GetPlatformLimits()); err != nil {
		cps.logger.Error("Post content validation failed", err, "post_id", postID)
		return nil, err
	}

	// Update post status to publishing
	post.Status = entities.PostStatusPublishing
	post.UpdatedBy = publishedBy
	post.UpdatedAt = time.Now()

	if err := cps.postRepo.Update(ctx, post); err != nil {
		cps.logger.Error("Failed to update post status to publishing", err, "post_id", postID)
		return nil, err
	}

	// Publish to platform
	result, err := client.PublishPost(ctx, post, credentials)
	if err != nil {
		cps.logger.Error("Failed to publish post to platform", err, "post_id", postID, "platform", post.Platform)
		
		// Update post status to failed
		post.Status = entities.PostStatusFailed
		post.ErrorMessage = err.Error()
		if updateErr := cps.postRepo.Update(ctx, post); updateErr != nil {
			cps.logger.Error("Failed to update post status to failed", updateErr, "post_id", postID)
		}

		return nil, err
	}

	// Update post with publish results
	post.UpdateStatus(entities.PostStatusPublished, publishedBy)
	post.SetPlatformPostID(result.PlatformPostID, publishedBy)
	post.PublishedAt = &result.PublishedAt
	post.ErrorMessage = ""

	if err := cps.postRepo.Update(ctx, post); err != nil {
		cps.logger.Error("Failed to update post with publish results", err, "post_id", postID)
	}

	// Store initial metrics if available
	if result.Metrics != nil {
		analytics := &entities.PostAnalytics{
			Impressions: result.Metrics.Impressions,
			Reach:       result.Metrics.Reach,
			Engagement:  result.Metrics.Engagement,
			Likes:       result.Metrics.Likes,
			Comments:    result.Metrics.Comments,
			Shares:      result.Metrics.Shares,
			Saves:       result.Metrics.Saves,
			Clicks:      result.Metrics.Clicks,
			VideoViews:  result.Metrics.VideoViews,
		}

		// Update post with analytics
		post.Analytics = analytics
		if err := cps.postRepo.Update(ctx, post); err != nil {
			cps.logger.Error("Failed to save initial post analytics", err, "post_id", postID)
		}
	}

	// Publish post published event
	event := NewPostPublishedEvent(post, result, publishedBy)
	if err := cps.eventPublisher.PublishEvent(ctx, event); err != nil {
		cps.logger.Error("Failed to publish post published event", err, "post_id", postID)
	}

	cps.logger.Info("Post published successfully", "post_id", postID, "platform_post_id", result.PlatformPostID)
	return result, nil
}

// PublishContent immediately publishes content to specified platforms
func (cps *ContentPublishingService) PublishContent(ctx context.Context, contentID uuid.UUID, platforms []entities.PlatformType, publishedBy uuid.UUID) ([]*PublishResult, error) {
	cps.logger.Info("Publishing content immediately", "content_id", contentID, "platforms", platforms)

	// Get content
	content, err := cps.contentRepo.GetByID(ctx, contentID)
	if err != nil {
		cps.logger.Error("Failed to get content", err, "content_id", contentID)
		return nil, err
	}

	// Validate content can be published
	if !content.CanBePublished() {
		return nil, fmt.Errorf("content cannot be published: status=%s", content.Status)
	}

	var results []*PublishResult

	// Create and publish posts for each platform
	for _, platform := range platforms {
		// Create post
		post := entities.NewPost(contentID, platform, content.Body, publishedBy)
		post.Hashtags = content.Hashtags
		post.Mentions = content.Mentions
		post.Status = entities.PostStatusDraft

		// Save post
		if err := cps.postRepo.Create(ctx, post); err != nil {
			cps.logger.Error("Failed to create post for immediate publishing", err, "content_id", contentID, "platform", platform)
			continue
		}

		// Publish post
		result, err := cps.PublishPost(ctx, post.ID, publishedBy)
		if err != nil {
			cps.logger.Error("Failed to publish post immediately", err, "post_id", post.ID)
			continue
		}

		results = append(results, result)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("failed to publish content on any platform")
	}

	cps.logger.Info("Content published immediately", "content_id", contentID, "successful_publishes", len(results))
	return results, nil
}

// UpdatePost updates an already published post
func (cps *ContentPublishingService) UpdatePost(ctx context.Context, postID uuid.UUID, newContent string, updatedBy uuid.UUID) (*PublishResult, error) {
	cps.logger.Info("Updating published post", "post_id", postID)

	// Get post
	post, err := cps.postRepo.GetByID(ctx, postID)
	if err != nil {
		cps.logger.Error("Failed to get post", err, "post_id", postID)
		return nil, err
	}

	// Validate post can be updated
	if post.Status != entities.PostStatusPublished {
		return nil, fmt.Errorf("post cannot be updated: current status=%s", post.Status)
	}

	if post.PlatformPostID == "" {
		return nil, fmt.Errorf("post has no platform post ID")
	}

	// Get brand and credentials
	brand, err := cps.brandRepo.GetByID(ctx, post.ContentID)
	if err != nil {
		cps.logger.Error("Failed to get brand", err, "post_id", postID)
		return nil, err
	}

	credentials, err := cps.getPlatformCredentials(brand, post.Platform)
	if err != nil {
		cps.logger.Error("Failed to get platform credentials", err, "brand_id", brand.ID, "platform", post.Platform)
		return nil, err
	}

	// Get platform client
	client, exists := cps.platformClients[post.Platform]
	if !exists {
		return nil, fmt.Errorf("no client available for platform: %s", post.Platform)
	}

	// Update on platform
	result, err := client.UpdatePost(ctx, post.PlatformPostID, newContent, credentials)
	if err != nil {
		cps.logger.Error("Failed to update post on platform", err, "post_id", postID, "platform", post.Platform)
		return nil, err
	}

	// Update post in database
	post.Text = newContent
	post.UpdatedBy = updatedBy
	post.UpdatedAt = time.Now()

	if err := cps.postRepo.Update(ctx, post); err != nil {
		cps.logger.Error("Failed to update post in database", err, "post_id", postID)
	}

	// Publish post updated event
	event := NewPostUpdatedEvent(post, updatedBy)
	if err := cps.eventPublisher.PublishEvent(ctx, event); err != nil {
		cps.logger.Error("Failed to publish post updated event", err, "post_id", postID)
	}

	cps.logger.Info("Post updated successfully", "post_id", postID)
	return result, nil
}

// DeletePost deletes a published post
func (cps *ContentPublishingService) DeletePost(ctx context.Context, postID uuid.UUID, deletedBy uuid.UUID) error {
	cps.logger.Info("Deleting published post", "post_id", postID)

	// Get post
	post, err := cps.postRepo.GetByID(ctx, postID)
	if err != nil {
		cps.logger.Error("Failed to get post", err, "post_id", postID)
		return err
	}

	// Validate post can be deleted
	if post.Status != entities.PostStatusPublished {
		return fmt.Errorf("post cannot be deleted: current status=%s", post.Status)
	}

	if post.PlatformPostID == "" {
		return fmt.Errorf("post has no platform post ID")
	}

	// Get brand and credentials
	brand, err := cps.brandRepo.GetByID(ctx, post.ContentID)
	if err != nil {
		cps.logger.Error("Failed to get brand", err, "post_id", postID)
		return err
	}

	credentials, err := cps.getPlatformCredentials(brand, post.Platform)
	if err != nil {
		cps.logger.Error("Failed to get platform credentials", err, "brand_id", brand.ID, "platform", post.Platform)
		return err
	}

	// Get platform client
	client, exists := cps.platformClients[post.Platform]
	if !exists {
		return fmt.Errorf("no client available for platform: %s", post.Platform)
	}

	// Delete from platform
	if err := client.DeletePost(ctx, post.PlatformPostID, credentials); err != nil {
		cps.logger.Error("Failed to delete post from platform", err, "post_id", postID, "platform", post.Platform)
		return err
	}

	// Update post status
	post.UpdateStatus(entities.PostStatusDeleted, deletedBy)
	post.UpdatedBy = deletedBy
	post.UpdatedAt = time.Now()

	if err := cps.postRepo.Update(ctx, post); err != nil {
		cps.logger.Error("Failed to update post status to deleted", err, "post_id", postID)
	}

	// Publish post deleted event
	event := NewPostDeletedEvent(post, deletedBy)
	if err := cps.eventPublisher.PublishEvent(ctx, event); err != nil {
		cps.logger.Error("Failed to publish post deleted event", err, "post_id", postID)
	}

	cps.logger.Info("Post deleted successfully", "post_id", postID)
	return nil
}

// RefreshPostMetrics fetches updated metrics for a published post
func (cps *ContentPublishingService) RefreshPostMetrics(ctx context.Context, postID uuid.UUID) (*PostMetrics, error) {
	cps.logger.Info("Refreshing post metrics", "post_id", postID)

	// Get post
	post, err := cps.postRepo.GetByID(ctx, postID)
	if err != nil {
		cps.logger.Error("Failed to get post", err, "post_id", postID)
		return nil, err
	}

	// Validate post is published
	if post.Status != entities.PostStatusPublished || post.PlatformPostID == "" {
		return nil, fmt.Errorf("post is not published or has no platform post ID")
	}

	// Get brand and credentials
	brand, err := cps.brandRepo.GetByID(ctx, post.ContentID)
	if err != nil {
		cps.logger.Error("Failed to get brand", err, "post_id", postID)
		return nil, err
	}

	credentials, err := cps.getPlatformCredentials(brand, post.Platform)
	if err != nil {
		cps.logger.Error("Failed to get platform credentials", err, "brand_id", brand.ID, "platform", post.Platform)
		return nil, err
	}

	// Get platform client
	client, exists := cps.platformClients[post.Platform]
	if !exists {
		return nil, fmt.Errorf("no client available for platform: %s", post.Platform)
	}

	// Fetch metrics from platform
	metrics, err := client.GetPostMetrics(ctx, post.PlatformPostID, credentials)
	if err != nil {
		cps.logger.Error("Failed to fetch post metrics from platform", err, "post_id", postID, "platform", post.Platform)
		return nil, err
	}

	// Save metrics to database
	analytics := &entities.PostAnalytics{
		Impressions: metrics.Impressions,
		Reach:       metrics.Reach,
		Engagement:  metrics.Engagement,
		Likes:       metrics.Likes,
		Comments:    metrics.Comments,
		Shares:      metrics.Shares,
		Saves:       metrics.Saves,
		Clicks:      metrics.Clicks,
		VideoViews:  metrics.VideoViews,
	}

	// Update post with refreshed analytics
	post.Analytics = analytics
	if err := cps.postRepo.Update(ctx, post); err != nil {
		cps.logger.Error("Failed to save refreshed post analytics", err, "post_id", postID)
	}

	cps.logger.Info("Post metrics refreshed successfully", "post_id", postID)
	return metrics, nil
}

// Helper methods

func (cps *ContentPublishingService) getPlatformCredentials(brand *entities.Brand, platform entities.PlatformType) (*PlatformCredentials, error) {
	// This would typically fetch credentials from a secure store
	// For now, we'll return a placeholder
	return &PlatformCredentials{
		Platform:    platform,
		AccessToken: "placeholder_token",
	}, nil
}

func (cps *ContentPublishingService) validatePostContent(post *entities.Post, limits *PlatformLimits) error {
	if limits == nil {
		return nil // No limits to validate against
	}

	if len(post.Text) > limits.MaxTextLength {
		return fmt.Errorf("post content exceeds maximum length: %d > %d", len(post.Text), limits.MaxTextLength)
	}

	if len(post.Hashtags) > limits.MaxHashtags {
		return fmt.Errorf("post has too many hashtags: %d > %d", len(post.Hashtags), limits.MaxHashtags)
	}

	if len(post.Mentions) > limits.MaxMentions {
		return fmt.Errorf("post has too many mentions: %d > %d", len(post.Mentions), limits.MaxMentions)
	}

	return nil
}

// Event creation helpers

func NewPostPublishedEvent(post *entities.Post, result *PublishResult, publishedBy uuid.UUID) DomainEvent {
	return &PostPublishedEvent{
		AggregateID:      post.ID,
		PostID:           post.ID,
		PlatformPostID:   result.PlatformPostID,
		Platform:         post.Platform,
		PublishedBy:      publishedBy,
		Timestamp:        time.Now(),
		Version:          1,
	}
}

func NewPostUpdatedEvent(post *entities.Post, updatedBy uuid.UUID) DomainEvent {
	return &PostUpdatedEvent{
		AggregateID: post.ID,
		PostID:      post.ID,
		UpdatedBy:   updatedBy,
		Timestamp:   time.Now(),
		Version:     1,
	}
}

func NewPostDeletedEvent(post *entities.Post, deletedBy uuid.UUID) DomainEvent {
	return &PostDeletedEvent{
		AggregateID: post.ID,
		PostID:      post.ID,
		DeletedBy:   deletedBy,
		Timestamp:   time.Now(),
		Version:     1,
	}
}

// Event types

type PostPublishedEvent struct {
	AggregateID    uuid.UUID              `json:"aggregate_id"`
	PostID         uuid.UUID              `json:"post_id"`
	PlatformPostID string                 `json:"platform_post_id"`
	Platform       entities.PlatformType  `json:"platform"`
	PublishedBy    uuid.UUID              `json:"published_by"`
	Timestamp      time.Time              `json:"timestamp"`
	Version        int                    `json:"version"`
}

func (e *PostPublishedEvent) GetEventType() string        { return "post.published" }
func (e *PostPublishedEvent) GetAggregateID() uuid.UUID   { return e.AggregateID }
func (e *PostPublishedEvent) GetTimestamp() time.Time     { return e.Timestamp }
func (e *PostPublishedEvent) GetVersion() int             { return e.Version }
func (e *PostPublishedEvent) GetEventData() map[string]interface{} {
	return map[string]interface{}{
		"post_id":           e.PostID,
		"platform_post_id":  e.PlatformPostID,
		"platform":          e.Platform,
		"published_by":      e.PublishedBy,
	}
}

type PostUpdatedEvent struct {
	AggregateID uuid.UUID `json:"aggregate_id"`
	PostID      uuid.UUID `json:"post_id"`
	UpdatedBy   uuid.UUID `json:"updated_by"`
	Timestamp   time.Time `json:"timestamp"`
	Version     int       `json:"version"`
}

func (e *PostUpdatedEvent) GetEventType() string        { return "post.updated" }
func (e *PostUpdatedEvent) GetAggregateID() uuid.UUID   { return e.AggregateID }
func (e *PostUpdatedEvent) GetTimestamp() time.Time     { return e.Timestamp }
func (e *PostUpdatedEvent) GetVersion() int             { return e.Version }
func (e *PostUpdatedEvent) GetEventData() map[string]interface{} {
	return map[string]interface{}{
		"post_id":    e.PostID,
		"updated_by": e.UpdatedBy,
	}
}

type PostDeletedEvent struct {
	AggregateID uuid.UUID `json:"aggregate_id"`
	PostID      uuid.UUID `json:"post_id"`
	DeletedBy   uuid.UUID `json:"deleted_by"`
	Timestamp   time.Time `json:"timestamp"`
	Version     int       `json:"version"`
}

func (e *PostDeletedEvent) GetEventType() string        { return "post.deleted" }
func (e *PostDeletedEvent) GetAggregateID() uuid.UUID   { return e.AggregateID }
func (e *PostDeletedEvent) GetTimestamp() time.Time     { return e.Timestamp }
func (e *PostDeletedEvent) GetVersion() int             { return e.Version }
func (e *PostDeletedEvent) GetEventData() map[string]interface{} {
	return map[string]interface{}{
		"post_id":    e.PostID,
		"deleted_by": e.DeletedBy,
	}
}