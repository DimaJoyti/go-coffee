package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/social-media-content-agent/internal/domain/entities"
	"go-coffee-ai-agents/social-media-content-agent/internal/domain/repositories"
	"go-coffee-ai-agents/social-media-content-agent/internal/domain/services"
)

// ContentManagementUseCase provides application-level content management operations
type ContentManagementUseCase struct {
	contentService      *services.ContentGenerationService
	schedulingService   *services.ContentSchedulingService
	publishingService   *services.ContentPublishingService
	analyticsService    *services.AnalyticsService
	contentRepo         repositories.ContentRepository
	brandRepo           repositories.BrandRepository
	campaignRepo        repositories.CampaignRepository
	postRepo            repositories.PostRepository
	userRepo            repositories.UserRepository
	logger              services.Logger
}

// NewContentManagementUseCase creates a new content management use case
func NewContentManagementUseCase(
	contentService *services.ContentGenerationService,
	schedulingService *services.ContentSchedulingService,
	publishingService *services.ContentPublishingService,
	analyticsService *services.AnalyticsService,
	contentRepo repositories.ContentRepository,
	brandRepo repositories.BrandRepository,
	campaignRepo repositories.CampaignRepository,
	postRepo repositories.PostRepository,
	userRepo repositories.UserRepository,
	logger services.Logger,
) *ContentManagementUseCase {
	return &ContentManagementUseCase{
		contentService:    contentService,
		schedulingService: schedulingService,
		publishingService: publishingService,
		analyticsService:  analyticsService,
		contentRepo:       contentRepo,
		brandRepo:         brandRepo,
		campaignRepo:      campaignRepo,
		postRepo:          postRepo,
		userRepo:          userRepo,
		logger:            logger,
	}
}

// CreateContentRequest represents a request to create content
type CreateContentRequest struct {
	Title             string                     `json:"title" validate:"required,min=1,max=200"`
	Body              string                     `json:"body,omitempty" validate:"max=5000"`
	Type              entities.ContentType       `json:"type" validate:"required"`
	Format            entities.ContentFormat     `json:"format,omitempty"`
	Priority          entities.ContentPriority   `json:"priority,omitempty"`
	Category          entities.ContentCategory   `json:"category,omitempty"`
	BrandID           uuid.UUID                  `json:"brand_id" validate:"required"`
	CampaignID        *uuid.UUID                 `json:"campaign_id,omitempty"`
	Platforms         []entities.PlatformType    `json:"platforms,omitempty"`
	Hashtags          []string                   `json:"hashtags,omitempty"`
	Mentions          []string                   `json:"mentions,omitempty"`
	Tags              []string                   `json:"tags,omitempty"`
	Keywords          []string                   `json:"keywords,omitempty"`
	TargetAudience    *entities.TargetAudience   `json:"target_audience,omitempty"`
	Tone              entities.ContentTone       `json:"tone,omitempty"`
	Language          string                     `json:"language,omitempty"`
	ScheduledAt       *time.Time                 `json:"scheduled_at,omitempty"`
	ExpiresAt         *time.Time                 `json:"expires_at,omitempty"`
	MediaAssets       []*MediaAssetRequest       `json:"media_assets,omitempty"`
	CustomFields      map[string]interface{}     `json:"custom_fields,omitempty"`
	IsTemplate        bool                       `json:"is_template"`
	TemplateID        *uuid.UUID                 `json:"template_id,omitempty"`
	GenerateVariations bool                      `json:"generate_variations"`
	VariationCount    int                        `json:"variation_count,omitempty"`
	AutoOptimize      bool                       `json:"auto_optimize"`
	CreatedBy         uuid.UUID                  `json:"created_by" validate:"required"`
}

// MediaAssetRequest represents a request to add media asset
type MediaAssetRequest struct {
	Type        entities.MediaType `json:"type" validate:"required"`
	URL         string             `json:"url,omitempty"`
	FileName    string             `json:"file_name,omitempty"`
	AltText     string             `json:"alt_text,omitempty"`
	Caption     string             `json:"caption,omitempty"`
	Order       int                `json:"order"`
	AIGenerated bool               `json:"ai_generated"`
	AIPrompt    string             `json:"ai_prompt,omitempty"`
}

// CreateContentResponse represents the response from creating content
type CreateContentResponse struct {
	Content       *entities.Content             `json:"content"`
	Variations    []*entities.ContentVariation  `json:"variations,omitempty"`
	MediaAssets   []*entities.MediaAsset        `json:"media_assets,omitempty"`
	Quality       *services.ContentQualityAnalysis `json:"quality,omitempty"`
	Suggestions   []string                      `json:"suggestions,omitempty"`
	Warnings      []string                      `json:"warnings,omitempty"`
}

// CreateContent creates new content with advanced features
func (uc *ContentManagementUseCase) CreateContent(ctx context.Context, req *CreateContentRequest) (*CreateContentResponse, error) {
	uc.logger.Info("Creating content", "title", req.Title, "type", req.Type, "brand_id", req.BrandID)

	// Validate brand access
	brand, err := uc.brandRepo.GetByID(ctx, req.BrandID)
	if err != nil {
		uc.logger.Error("Failed to get brand", err, "brand_id", req.BrandID)
		return nil, err
	}

	if !brand.IsActive() {
		return nil, fmt.Errorf("brand is not active")
	}

	// Validate campaign if specified
	var campaign *entities.Campaign
	if req.CampaignID != nil {
		campaign, err = uc.campaignRepo.GetByID(ctx, *req.CampaignID)
		if err != nil {
			uc.logger.Error("Failed to get campaign", err, "campaign_id", *req.CampaignID)
			return nil, err
		}
		if !campaign.IsActive() {
			return nil, fmt.Errorf("campaign is not active")
		}
	}

	// Create content entity
	content := entities.NewContent(req.Title, req.Body, req.Type, req.BrandID, req.CreatedBy)
	
	// Set optional fields
	if req.Format != "" {
		content.Format = req.Format
	}
	if req.Priority != "" {
		content.Priority = req.Priority
	}
	if req.Category != "" {
		content.Category = req.Category
	}
	if req.CampaignID != nil {
		content.CampaignID = req.CampaignID
	}
	if req.Tone != "" {
		content.Tone = req.Tone
	}
	if req.Language != "" {
		content.Language = req.Language
	}
	if req.ScheduledAt != nil {
		content.ScheduledAt = req.ScheduledAt
		content.Status = entities.StatusScheduled
	}
	if req.ExpiresAt != nil {
		content.ExpiresAt = req.ExpiresAt
	}
	
	content.Platforms = req.Platforms
	content.Hashtags = req.Hashtags
	content.Mentions = req.Mentions
	content.Tags = req.Tags
	content.Keywords = req.Keywords
	content.TargetAudience = req.TargetAudience
	content.CustomFields = req.CustomFields
	content.IsTemplate = req.IsTemplate
	content.TemplateID = req.TemplateID

	// Apply brand voice and guidelines using content service
	if err := uc.contentService.ApplyBrandGuidelines(content, brand); err != nil {
		uc.logger.Warn("Failed to apply brand guidelines", "content_id", content.ID, "error", err)
	}

	// Auto-optimize content if requested
	if req.AutoOptimize {
		if err := uc.contentService.EnhanceContent(ctx, content, brand); err != nil {
			uc.logger.Warn("Failed to auto-optimize content", "content_id", content.ID, "error", err)
		}
	}

	// Create content in repository
	if err := uc.contentRepo.Create(ctx, content); err != nil {
		uc.logger.Error("Failed to create content", err, "content_id", content.ID)
		return nil, err
	}

	response := &CreateContentResponse{
		Content: content,
	}

	// Add media assets
	for _, assetReq := range req.MediaAssets {
		asset := &entities.MediaAsset{
			ID:          uuid.New(),
			ContentID:   content.ID,
			Type:        assetReq.Type,
			URL:         assetReq.URL,
			FileName:    assetReq.FileName,
			AltText:     assetReq.AltText,
			Caption:     assetReq.Caption,
			Order:       assetReq.Order,
			AIGenerated: assetReq.AIGenerated,
			AIPrompt:    assetReq.AIPrompt,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := uc.contentRepo.AddMediaAsset(ctx, asset); err != nil {
			uc.logger.Error("Failed to add media asset", err, "asset_id", asset.ID)
			continue
		}

		response.MediaAssets = append(response.MediaAssets, asset)
	}

	// Generate variations if requested
	if req.GenerateVariations {
		count := req.VariationCount
		if count == 0 {
			count = 3 // Default variation count
		}

		variations, err := uc.contentService.GenerateVariations(ctx, content, count)
		if err != nil {
			uc.logger.Warn("Failed to generate variations", "content_id", content.ID, "error", err)
		} else {
			for _, variation := range variations {
				if err := uc.contentRepo.AddVariation(ctx, variation); err != nil {
					uc.logger.Error("Failed to save variation", err, "variation_id", variation.ID)
					continue
				}
				response.Variations = append(response.Variations, variation)
			}
		}
	}

	// Analyze content quality
	quality, err := uc.contentService.AnalyzeContentQuality(ctx, content)
	if err != nil {
		uc.logger.Warn("Failed to analyze content quality", "content_id", content.ID, "error", err)
	} else {
		response.Quality = quality
		response.Suggestions = quality.Recommendations
		response.Warnings = quality.Warnings
	}

	// Add to campaign if specified
	if campaign != nil {
		campaign.AddContent(content)
		if err := uc.campaignRepo.Update(ctx, campaign); err != nil {
			uc.logger.Error("Failed to update campaign with content", err, "campaign_id", campaign.ID)
		}
	}

	uc.logger.Info("Content created successfully", "content_id", content.ID, "variations", len(response.Variations))
	return response, nil
}

// UpdateContentRequest represents a request to update content
type UpdateContentRequest struct {
	ID                uuid.UUID                  `json:"id" validate:"required"`
	Title             *string                    `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Body              *string                    `json:"body,omitempty" validate:"omitempty,max=5000"`
	Status            *entities.ContentStatus    `json:"status,omitempty"`
	Priority          *entities.ContentPriority  `json:"priority,omitempty"`
	Category          *entities.ContentCategory  `json:"category,omitempty"`
	Platforms         []entities.PlatformType    `json:"platforms,omitempty"`
	Hashtags          []string                   `json:"hashtags,omitempty"`
	Mentions          []string                   `json:"mentions,omitempty"`
	Tags              []string                   `json:"tags,omitempty"`
	Keywords          []string                   `json:"keywords,omitempty"`
	TargetAudience    *entities.TargetAudience   `json:"target_audience,omitempty"`
	Tone              *entities.ContentTone      `json:"tone,omitempty"`
	Language          *string                    `json:"language,omitempty"`
	ScheduledAt       *time.Time                 `json:"scheduled_at,omitempty"`
	ExpiresAt         *time.Time                 `json:"expires_at,omitempty"`
	CustomFields      map[string]interface{}     `json:"custom_fields,omitempty"`
	UpdatedBy         uuid.UUID                  `json:"updated_by" validate:"required"`
	ReanalyzeQuality  bool                       `json:"reanalyze_quality"`
}

// UpdateContent updates existing content
func (uc *ContentManagementUseCase) UpdateContent(ctx context.Context, req *UpdateContentRequest) (*entities.Content, error) {
	uc.logger.Info("Updating content", "content_id", req.ID, "updated_by", req.UpdatedBy)

	// Get existing content
	content, err := uc.contentRepo.GetByID(ctx, req.ID)
	if err != nil {
		uc.logger.Error("Failed to get content", err, "content_id", req.ID)
		return nil, err
	}

	// Check if content is archived
	if content.IsArchived {
		return nil, fmt.Errorf("cannot update archived content")
	}

	// Update fields
	if req.Title != nil {
		content.Title = *req.Title
	}
	if req.Body != nil {
		content.Body = *req.Body
	}
	if req.Status != nil {
		content.UpdateStatus(*req.Status, req.UpdatedBy)
	}
	if req.Priority != nil {
		content.Priority = *req.Priority
	}
	if req.Category != nil {
		content.Category = *req.Category
	}
	if req.Tone != nil {
		content.Tone = *req.Tone
	}
	if req.Language != nil {
		content.Language = *req.Language
	}
	if req.ScheduledAt != nil {
		content.ScheduledAt = req.ScheduledAt
	}
	if req.ExpiresAt != nil {
		content.ExpiresAt = req.ExpiresAt
	}
	
	if req.Platforms != nil {
		content.Platforms = req.Platforms
	}
	if req.Hashtags != nil {
		content.Hashtags = req.Hashtags
	}
	if req.Mentions != nil {
		content.Mentions = req.Mentions
	}
	if req.Tags != nil {
		content.Tags = req.Tags
	}
	if req.Keywords != nil {
		content.Keywords = req.Keywords
	}
	if req.TargetAudience != nil {
		content.TargetAudience = req.TargetAudience
	}
	if req.CustomFields != nil {
		content.CustomFields = req.CustomFields
	}

	content.UpdatedBy = req.UpdatedBy
	content.UpdatedAt = time.Now()
	content.Version++

	// Update content in repository
	if err := uc.contentRepo.Update(ctx, content); err != nil {
		uc.logger.Error("Failed to update content", err, "content_id", req.ID)
		return nil, err
	}

	// Reanalyze quality if requested
	if req.ReanalyzeQuality {
		brand, err := uc.brandRepo.GetByID(ctx, content.BrandID)
		if err != nil {
			uc.logger.Warn("Failed to get brand for quality reanalysis", "content_id", content.ID, "brand_id", content.BrandID, "error", err)
		} else if !brand.IsActive() {
			uc.logger.Warn("Cannot reanalyze quality for inactive brand", "content_id", content.ID, "brand_id", content.BrandID)
		} else {
			// Brand is valid and active, proceed with quality analysis
			if quality, err := uc.contentService.AnalyzeContentQuality(ctx, content); err != nil {
				uc.logger.Error("Failed to reanalyze content quality", err, "content_id", content.ID)
			} else {
				uc.logger.Info("Content quality reanalyzed", "content_id", content.ID, "score", quality.OverallScore)
			}
		}
	}

	uc.logger.Info("Content updated successfully", "content_id", content.ID)
	return content, nil
}

// ScheduleContentRequest represents a request to schedule content
type ScheduleContentRequest struct {
	ContentID   uuid.UUID                `json:"content_id" validate:"required"`
	Platforms   []entities.PlatformType  `json:"platforms" validate:"required,min=1"`
	ScheduledAt time.Time                `json:"scheduled_at" validate:"required"`
	TimeZone    string                   `json:"time_zone,omitempty"`
	Recurring   *RecurringSchedule       `json:"recurring,omitempty"`
	ScheduledBy uuid.UUID                `json:"scheduled_by" validate:"required"`
}

// RecurringSchedule represents recurring schedule settings
type RecurringSchedule struct {
	Frequency   string    `json:"frequency"` // daily, weekly, monthly
	Interval    int       `json:"interval"`  // every N frequency
	DaysOfWeek  []int     `json:"days_of_week,omitempty"` // 0=Sunday, 1=Monday, etc.
	EndDate     *time.Time `json:"end_date,omitempty"`
	MaxOccurrences *int   `json:"max_occurrences,omitempty"`
}

// ScheduleContent schedules content for publishing
func (uc *ContentManagementUseCase) ScheduleContent(ctx context.Context, req *ScheduleContentRequest) ([]*entities.Post, error) {
	uc.logger.Info("Scheduling content", "content_id", req.ContentID, "platforms", req.Platforms, "scheduled_at", req.ScheduledAt)

	// Get content
	content, err := uc.contentRepo.GetByID(ctx, req.ContentID)
	if err != nil {
		uc.logger.Error("Failed to get content", err, "content_id", req.ContentID)
		return nil, err
	}

	// Validate content can be scheduled
	if !content.CanBePublished() {
		return nil, fmt.Errorf("content cannot be published: status=%s, platforms=%d", content.Status, len(content.Platforms))
	}

	// Validate scheduled time
	if req.ScheduledAt.Before(time.Now()) {
		return nil, fmt.Errorf("scheduled time must be in the future")
	}

	// Use scheduling service to create posts
	posts, err := uc.schedulingService.ScheduleContent(ctx, content, req.Platforms, req.ScheduledAt, req.ScheduledBy)
	if err != nil {
		uc.logger.Error("Failed to schedule content", err, "content_id", req.ContentID)
		return nil, err
	}

	// Update content status
	content.UpdateStatus(entities.StatusScheduled, req.ScheduledBy)
	content.ScheduledAt = &req.ScheduledAt

	if err := uc.contentRepo.Update(ctx, content); err != nil {
		uc.logger.Error("Failed to update content status", err, "content_id", req.ContentID)
	}

	uc.logger.Info("Content scheduled successfully", "content_id", req.ContentID, "posts", len(posts))
	return posts, nil
}

// GetContentRequest represents a request to get content with filtering
type GetContentRequest struct {
	Filter         *repositories.ContentFilter `json:"filter,omitempty"`
	BrandID        *uuid.UUID                  `json:"brand_id,omitempty"`
	CampaignID     *uuid.UUID                  `json:"campaign_id,omitempty"`
	CreatorID      *uuid.UUID                  `json:"creator_id,omitempty"`
	IncludeMedia   bool                        `json:"include_media"`
	IncludeVariations bool                     `json:"include_variations"`
	IncludeAnalytics bool                      `json:"include_analytics"`
}

// GetContent retrieves content with filtering and additional data
func (uc *ContentManagementUseCase) GetContent(ctx context.Context, req *GetContentRequest) ([]*entities.Content, error) {
	uc.logger.Info("Getting content", "brand_id", req.BrandID, "campaign_id", req.CampaignID)

	var contents []*entities.Content
	var err error

	// Apply different query strategies based on request
	if req.BrandID != nil {
		contents, err = uc.contentRepo.ListByBrand(ctx, *req.BrandID, req.Filter)
	} else if req.CampaignID != nil {
		contents, err = uc.contentRepo.ListByCampaign(ctx, *req.CampaignID, req.Filter)
	} else if req.CreatorID != nil {
		contents, err = uc.contentRepo.ListByCreator(ctx, *req.CreatorID, req.Filter)
	} else {
		contents, err = uc.contentRepo.List(ctx, req.Filter)
	}

	if err != nil {
		uc.logger.Error("Failed to get content", err)
		return nil, err
	}

	// Load additional data if requested
	for _, content := range contents {
		if req.IncludeMedia {
			mediaAssets, err := uc.contentRepo.GetMediaAssets(ctx, content.ID)
			if err == nil {
				content.MediaAssets = mediaAssets
			}
		}

		if req.IncludeVariations {
			variations, err := uc.contentRepo.GetVariations(ctx, content.ID)
			if err == nil {
				content.Variations = variations
			}
		}

		if req.IncludeAnalytics {
			// Load analytics data from posts
			posts, err := uc.postRepo.ListByContent(ctx, content.ID, nil)
			if err == nil && len(posts) > 0 {
				// Aggregate analytics from all posts
				analytics := uc.aggregateContentAnalytics(posts)
				content.Analytics = analytics
			}
		}
	}

	uc.logger.Info("Retrieved content", "count", len(contents))
	return contents, nil
}

// Helper methods


func (uc *ContentManagementUseCase) aggregateContentAnalytics(posts []*entities.Post) *entities.ContentAnalytics {
	analytics := &entities.ContentAnalytics{
		PlatformMetrics: make(map[entities.PlatformType]*entities.PlatformMetrics),
		LastUpdated:     time.Now(),
	}

	for _, post := range posts {
		if post.Analytics != nil {
			analytics.Impressions += post.Analytics.Impressions
			analytics.Reach += post.Analytics.Reach
			analytics.Clicks += post.Analytics.Clicks
			analytics.Likes += post.Analytics.Likes
			analytics.Shares += post.Analytics.Shares
			analytics.Comments += post.Analytics.Comments
			analytics.Saves += post.Analytics.Saves

			// Platform-specific metrics
			if analytics.PlatformMetrics[post.Platform] == nil {
				analytics.PlatformMetrics[post.Platform] = &entities.PlatformMetrics{
					Platform: post.Platform,
				}
			}
			
			platformMetrics := analytics.PlatformMetrics[post.Platform]
			platformMetrics.Impressions += post.Analytics.Impressions
			platformMetrics.Reach += post.Analytics.Reach
			platformMetrics.Engagement += post.Analytics.Engagement
			platformMetrics.Clicks += post.Analytics.Clicks
			platformMetrics.Shares += post.Analytics.Shares
			platformMetrics.Comments += post.Analytics.Comments
			platformMetrics.Likes += post.Analytics.Likes
		}
	}

	// Calculate rates
	if analytics.Impressions > 0 {
		totalEngagement := analytics.Likes + analytics.Comments + analytics.Shares + analytics.Saves
		analytics.EngagementRate = float64(totalEngagement) / float64(analytics.Impressions) * 100
		analytics.CTR = float64(analytics.Clicks) / float64(analytics.Impressions) * 100
	}

	return analytics
}
