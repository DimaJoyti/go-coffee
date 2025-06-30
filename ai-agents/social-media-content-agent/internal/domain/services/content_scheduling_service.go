package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/social-media-content-agent/internal/domain/entities"
	"go-coffee-ai-agents/social-media-content-agent/internal/domain/repositories"
)

// ContentSchedulingService provides content scheduling functionality
type ContentSchedulingService struct {
	postRepo         repositories.PostRepository
	contentRepo      repositories.ContentRepository
	brandRepo        repositories.BrandRepository
	scheduler        SchedulerService
	eventPublisher   EventPublisher
	logger           Logger
}

// SchedulerService defines the interface for scheduling operations
type SchedulerService interface {
	SchedulePost(ctx context.Context, post *entities.Post, scheduledAt time.Time) error
	CancelScheduledPost(ctx context.Context, postID uuid.UUID) error
	UpdateScheduledPost(ctx context.Context, postID uuid.UUID, scheduledAt time.Time) error
	GetScheduledPosts(ctx context.Context, startTime, endTime time.Time) ([]*entities.Post, error)
	ProcessScheduledPosts(ctx context.Context) error
}

// NewContentSchedulingService creates a new content scheduling service
func NewContentSchedulingService(
	postRepo repositories.PostRepository,
	contentRepo repositories.ContentRepository,
	brandRepo repositories.BrandRepository,
	scheduler SchedulerService,
	eventPublisher EventPublisher,
	logger Logger,
) *ContentSchedulingService {
	return &ContentSchedulingService{
		postRepo:       postRepo,
		contentRepo:    contentRepo,
		brandRepo:      brandRepo,
		scheduler:      scheduler,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// ScheduleContent schedules content for publishing across multiple platforms
func (css *ContentSchedulingService) ScheduleContent(
	ctx context.Context,
	content *entities.Content,
	platforms []entities.PlatformType,
	scheduledAt time.Time,
	scheduledBy uuid.UUID,
) ([]*entities.Post, error) {
	css.logger.Info("Scheduling content", "content_id", content.ID, "platforms", platforms, "scheduled_at", scheduledAt)

	// Validate content can be scheduled
	if !content.CanBePublished() {
		return nil, fmt.Errorf("content cannot be published: status=%s", content.Status)
	}

	// Validate scheduled time
	if scheduledAt.Before(time.Now()) {
		return nil, fmt.Errorf("scheduled time must be in the future")
	}

	// Get brand information for platform-specific settings
	brand, err := css.brandRepo.GetByID(ctx, content.BrandID)
	if err != nil {
		css.logger.Error("Failed to get brand", err, "brand_id", content.BrandID)
		return nil, err
	}

	var posts []*entities.Post

	// Create posts for each platform
	for _, platform := range platforms {
		// Check if brand has configuration for this platform
		if brand.GetSocialProfile(platform) == nil {
			css.logger.Warn("Brand has no configuration for platform", "brand_id", brand.ID, "platform", platform)
			continue
		}

		// Create post entity
		post := entities.NewPost(content.ID, platform, content.Body, scheduledBy)
		post.UpdateStatus(entities.PostStatusScheduled, scheduledBy)
		post.ScheduledAt = &scheduledAt
		
		// Customize content for platform
		if err := css.customizeContentForPlatform(post, content, platform, brand); err != nil {
			css.logger.Error("Failed to customize content for platform", err, "platform", platform)
			continue
		}

		// Save post to repository
		if err := css.postRepo.Create(ctx, post); err != nil {
			css.logger.Error("Failed to create scheduled post", err, "post_id", post.ID)
			continue
		}

		// Schedule post with scheduler service
		if err := css.scheduler.SchedulePost(ctx, post, scheduledAt); err != nil {
			css.logger.Error("Failed to schedule post", err, "post_id", post.ID)
			// Try to delete the created post since scheduling failed
			if deleteErr := css.postRepo.Delete(ctx, post.ID); deleteErr != nil {
				css.logger.Error("Failed to cleanup failed post", deleteErr, "post_id", post.ID)
			}
			continue
		}

		posts = append(posts, post)
		css.logger.Info("Post scheduled successfully", "post_id", post.ID, "platform", platform)
	}

	if len(posts) == 0 {
		return nil, fmt.Errorf("failed to schedule content on any platform")
	}

	// Publish content scheduled event
	event := NewContentScheduledEvent(content, posts, scheduledBy)
	if err := css.eventPublisher.PublishEvent(ctx, event); err != nil {
		css.logger.Error("Failed to publish content scheduled event", err, "content_id", content.ID)
	}

	css.logger.Info("Content scheduled successfully", "content_id", content.ID, "posts_created", len(posts))
	return posts, nil
}

// ReschedulePost reschedules an existing post
func (css *ContentSchedulingService) ReschedulePost(ctx context.Context, postID uuid.UUID, newScheduledAt time.Time, rescheduledBy uuid.UUID) error {
	css.logger.Info("Rescheduling post", "post_id", postID, "new_scheduled_at", newScheduledAt)

	// Get existing post
	post, err := css.postRepo.GetByID(ctx, postID)
	if err != nil {
		css.logger.Error("Failed to get post", err, "post_id", postID)
		return err
	}

	// Validate post can be rescheduled
	if post.Status != entities.PostStatusScheduled {
		return fmt.Errorf("post cannot be rescheduled: current status=%s", post.Status)
	}

	// Validate new scheduled time
	if newScheduledAt.Before(time.Now()) {
		return fmt.Errorf("new scheduled time must be in the future")
	}

	// Update scheduler
	if err := css.scheduler.UpdateScheduledPost(ctx, postID, newScheduledAt); err != nil {
		css.logger.Error("Failed to update scheduled post", err, "post_id", postID)
		return err
	}

	// Update post
	post.ScheduledAt = &newScheduledAt
	post.UpdatedBy = rescheduledBy
	post.UpdatedAt = time.Now()

	if err := css.postRepo.Update(ctx, post); err != nil {
		css.logger.Error("Failed to update post", err, "post_id", postID)
		return err
	}

	// Publish post rescheduled event
	event := NewPostRescheduledEvent(post, rescheduledBy)
	if err := css.eventPublisher.PublishEvent(ctx, event); err != nil {
		css.logger.Error("Failed to publish post rescheduled event", err, "post_id", postID)
	}

	css.logger.Info("Post rescheduled successfully", "post_id", postID)
	return nil
}

// CancelScheduledPost cancels a scheduled post
func (css *ContentSchedulingService) CancelScheduledPost(ctx context.Context, postID uuid.UUID, cancelledBy uuid.UUID) error {
	css.logger.Info("Cancelling scheduled post", "post_id", postID)

	// Get existing post
	post, err := css.postRepo.GetByID(ctx, postID)
	if err != nil {
		css.logger.Error("Failed to get post", err, "post_id", postID)
		return err
	}

	// Validate post can be cancelled
	if post.Status != entities.PostStatusScheduled {
		return fmt.Errorf("post cannot be cancelled: current status=%s", post.Status)
	}

	// Cancel with scheduler
	if err := css.scheduler.CancelScheduledPost(ctx, postID); err != nil {
		css.logger.Error("Failed to cancel scheduled post", err, "post_id", postID)
		return err
	}

	// Update post status
	post.UpdateStatus(entities.PostStatusDeleted, cancelledBy)
	post.UpdatedBy = cancelledBy
	post.UpdatedAt = time.Now()

	if err := css.postRepo.Update(ctx, post); err != nil {
		css.logger.Error("Failed to update cancelled post", err, "post_id", postID)
		return err
	}

	// Publish post cancelled event
	event := NewPostCancelledEvent(post, cancelledBy)
	if err := css.eventPublisher.PublishEvent(ctx, event); err != nil {
		css.logger.Error("Failed to publish post cancelled event", err, "post_id", postID)
	}

	css.logger.Info("Post cancelled successfully", "post_id", postID)
	return nil
}

// GetScheduledPosts retrieves scheduled posts within a time range
func (css *ContentSchedulingService) GetScheduledPosts(ctx context.Context, startTime, endTime time.Time) ([]*entities.Post, error) {
	css.logger.Info("Getting scheduled posts", "start_time", startTime, "end_time", endTime)

	posts, err := css.scheduler.GetScheduledPosts(ctx, startTime, endTime)
	if err != nil {
		css.logger.Error("Failed to get scheduled posts", err)
		return nil, err
	}

	css.logger.Info("Retrieved scheduled posts", "count", len(posts))
	return posts, nil
}

// ProcessScheduledPosts processes posts that are ready to be published
func (css *ContentSchedulingService) ProcessScheduledPosts(ctx context.Context) error {
	css.logger.Info("Processing scheduled posts")

	if err := css.scheduler.ProcessScheduledPosts(ctx); err != nil {
		css.logger.Error("Failed to process scheduled posts", err)
		return err
	}

	css.logger.Info("Scheduled posts processed successfully")
	return nil
}

// Helper methods

func (css *ContentSchedulingService) customizeContentForPlatform(post *entities.Post, content *entities.Content, platform entities.PlatformType, brand *entities.Brand) error {
	// Set basic content
	post.Text = content.Body
	post.Hashtags = content.Hashtags
	post.Mentions = content.Mentions

	// Platform-specific customizations
	switch platform {
	case entities.PlatformTwitter:
		// Twitter has character limits
		if len(post.Text) > 280 {
			post.Text = post.Text[:277] + "..."
		}
		// Limit hashtags for Twitter
		if len(post.Hashtags) > 2 {
			post.Hashtags = post.Hashtags[:2]
		}

	case entities.PlatformInstagram:
		// Instagram allows more hashtags
		if len(post.Hashtags) > 30 {
			post.Hashtags = post.Hashtags[:30]
		}

	case entities.PlatformLinkedIn:
		// LinkedIn prefers professional tone
		// Could add tone adjustments here

	case entities.PlatformFacebook:
		// Facebook allows longer content
		// No specific limitations

	case entities.PlatformYouTube:
		// YouTube needs video content
		// This might require different handling

	case entities.PlatformTikTok:
		// TikTok is video-focused
		// This might require different handling
	}

	return nil
}

// Event creation helpers

func NewContentScheduledEvent(content *entities.Content, posts []*entities.Post, scheduledBy uuid.UUID) DomainEvent {
	return &ContentScheduledEvent{
		AggregateID: content.ID,
		ContentID:   content.ID,
		Posts:       posts,
		ScheduledBy: scheduledBy,
		Timestamp:   time.Now(),
		Version:     1,
	}
}

func NewPostRescheduledEvent(post *entities.Post, rescheduledBy uuid.UUID) DomainEvent {
	return &PostRescheduledEvent{
		AggregateID:   post.ID,
		PostID:        post.ID,
		RescheduledBy: rescheduledBy,
		Timestamp:     time.Now(),
		Version:       1,
	}
}

func NewPostCancelledEvent(post *entities.Post, cancelledBy uuid.UUID) DomainEvent {
	return &PostCancelledEvent{
		AggregateID: post.ID,
		PostID:      post.ID,
		CancelledBy: cancelledBy,
		Timestamp:   time.Now(),
		Version:     1,
	}
}

// Event types

type ContentScheduledEvent struct {
	AggregateID uuid.UUID         `json:"aggregate_id"`
	ContentID   uuid.UUID         `json:"content_id"`
	Posts       []*entities.Post  `json:"posts"`
	ScheduledBy uuid.UUID         `json:"scheduled_by"`
	Timestamp   time.Time         `json:"timestamp"`
	Version     int               `json:"version"`
}

func (e *ContentScheduledEvent) GetEventType() string        { return "content.scheduled" }
func (e *ContentScheduledEvent) GetAggregateID() uuid.UUID   { return e.AggregateID }
func (e *ContentScheduledEvent) GetTimestamp() time.Time     { return e.Timestamp }
func (e *ContentScheduledEvent) GetVersion() int             { return e.Version }
func (e *ContentScheduledEvent) GetEventData() map[string]interface{} {
	return map[string]interface{}{
		"content_id":   e.ContentID,
		"posts":        e.Posts,
		"scheduled_by": e.ScheduledBy,
	}
}

type PostRescheduledEvent struct {
	AggregateID   uuid.UUID `json:"aggregate_id"`
	PostID        uuid.UUID `json:"post_id"`
	RescheduledBy uuid.UUID `json:"rescheduled_by"`
	Timestamp     time.Time `json:"timestamp"`
	Version       int       `json:"version"`
}

func (e *PostRescheduledEvent) GetEventType() string        { return "post.rescheduled" }
func (e *PostRescheduledEvent) GetAggregateID() uuid.UUID   { return e.AggregateID }
func (e *PostRescheduledEvent) GetTimestamp() time.Time     { return e.Timestamp }
func (e *PostRescheduledEvent) GetVersion() int             { return e.Version }
func (e *PostRescheduledEvent) GetEventData() map[string]interface{} {
	return map[string]interface{}{
		"post_id":        e.PostID,
		"rescheduled_by": e.RescheduledBy,
	}
}

type PostCancelledEvent struct {
	AggregateID uuid.UUID `json:"aggregate_id"`
	PostID      uuid.UUID `json:"post_id"`
	CancelledBy uuid.UUID `json:"cancelled_by"`
	Timestamp   time.Time `json:"timestamp"`
	Version     int       `json:"version"`
}

func (e *PostCancelledEvent) GetEventType() string        { return "post.cancelled" }
func (e *PostCancelledEvent) GetAggregateID() uuid.UUID   { return e.AggregateID }
func (e *PostCancelledEvent) GetTimestamp() time.Time     { return e.Timestamp }
func (e *PostCancelledEvent) GetVersion() int             { return e.Version }
func (e *PostCancelledEvent) GetEventData() map[string]interface{} {
	return map[string]interface{}{
		"post_id":      e.PostID,
		"cancelled_by": e.CancelledBy,
	}
}