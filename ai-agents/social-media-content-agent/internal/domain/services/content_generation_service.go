package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/social-media-content-agent/internal/domain/entities"
	"go-coffee-ai-agents/social-media-content-agent/internal/domain/repositories"
)

// ContentGenerationService provides AI-powered content generation and enhancement
type ContentGenerationService struct {
	contentRepo      repositories.ContentRepository
	brandRepo        repositories.BrandRepository
	campaignRepo     repositories.CampaignRepository
	aiService        AIContentService
	nlpService       NLPService
	imageService     ImageGenerationService
	videoService     VideoGenerationService
	eventPublisher   EventPublisher
	logger           Logger
}

// AIContentService defines the interface for AI content operations
type AIContentService interface {
	GenerateContent(ctx context.Context, request *ContentGenerationRequest) (*ContentGenerationResponse, error)
	EnhanceContent(ctx context.Context, content *entities.Content, brand *entities.Brand) error
	GenerateVariations(ctx context.Context, content *entities.Content, count int) ([]*entities.ContentVariation, error)
	OptimizeForPlatform(ctx context.Context, content *entities.Content, platform entities.PlatformType) (*entities.Content, error)
	GenerateHashtags(ctx context.Context, content *entities.Content, platform entities.PlatformType, count int) ([]string, error)
	AnalyzeContentQuality(ctx context.Context, content *entities.Content) (*ContentQualityAnalysis, error)
	SuggestImprovements(ctx context.Context, content *entities.Content) ([]string, error)
	GenerateCaption(ctx context.Context, content *entities.Content, platform entities.PlatformType) (string, error)
}

// NLPService defines the interface for natural language processing
type NLPService interface {
	AnalyzeSentiment(ctx context.Context, text string) (*SentimentAnalysis, error)
	ExtractKeywords(ctx context.Context, text string) ([]string, error)
	DetectLanguage(ctx context.Context, text string) (string, float64, error)
	TranslateText(ctx context.Context, text, targetLanguage string) (string, error)
	SummarizeText(ctx context.Context, text string, maxLength int) (string, error)
	CheckGrammar(ctx context.Context, text string) (*GrammarCheck, error)
	AnalyzeReadability(ctx context.Context, text string) (*ReadabilityAnalysis, error)
	ExtractEntities(ctx context.Context, text string) ([]*NamedEntity, error)
}

// ImageGenerationService defines the interface for image generation
type ImageGenerationService interface {
	GenerateImage(ctx context.Context, request *ImageGenerationRequest) (*ImageGenerationResponse, error)
	EditImage(ctx context.Context, request *ImageEditRequest) (*ImageEditResponse, error)
	GenerateVariations(ctx context.Context, imageURL string, count int) ([]string, error)
	OptimizeForPlatform(ctx context.Context, imageURL string, platform entities.PlatformType) (string, error)
	AddBrandElements(ctx context.Context, imageURL string, brand *entities.Brand) (string, error)
	GenerateThumbnail(ctx context.Context, imageURL string, size string) (string, error)
}

// VideoGenerationService defines the interface for video generation
type VideoGenerationService interface {
	GenerateVideo(ctx context.Context, request *VideoGenerationRequest) (*VideoGenerationResponse, error)
	EditVideo(ctx context.Context, request *VideoEditRequest) (*VideoEditResponse, error)
	GenerateThumbnail(ctx context.Context, videoURL string) (string, error)
	AddSubtitles(ctx context.Context, videoURL string, subtitles []Subtitle) (string, error)
	OptimizeForPlatform(ctx context.Context, videoURL string, platform entities.PlatformType) (string, error)
	ExtractFrames(ctx context.Context, videoURL string, timestamps []float64) ([]string, error)
}

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	PublishEvent(ctx context.Context, event DomainEvent) error
	PublishEvents(ctx context.Context, events []DomainEvent) error
}

// DomainEvent represents a domain event
type DomainEvent interface {
	GetEventType() string
	GetAggregateID() uuid.UUID
	GetEventData() map[string]interface{}
	GetTimestamp() time.Time
	GetVersion() int
}

// Logger defines the interface for logging
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// Supporting types for content generation

// ContentGenerationRequest represents a request to generate content
type ContentGenerationRequest struct {
	BrandID        uuid.UUID                  `json:"brand_id"`
	CampaignID     *uuid.UUID                 `json:"campaign_id,omitempty"`
	Type           entities.ContentType       `json:"type"`
	Format         entities.ContentFormat     `json:"format"`
	Platforms      []entities.PlatformType    `json:"platforms"`
	Topic          string                     `json:"topic"`
	Keywords       []string                   `json:"keywords"`
	Tone           entities.ContentTone       `json:"tone"`
	Style          string                     `json:"style"`
	TargetAudience *entities.TargetAudience   `json:"target_audience,omitempty"`
	Length         ContentLength              `json:"length"`
	IncludeHashtags bool                      `json:"include_hashtags"`
	IncludeEmojis  bool                       `json:"include_emojis"`
	IncludeCTA     bool                       `json:"include_cta"`
	CustomPrompt   string                     `json:"custom_prompt,omitempty"`
	References     []string                   `json:"references,omitempty"`
	Constraints    *ContentConstraints        `json:"constraints,omitempty"`
	CreatedBy      uuid.UUID                  `json:"created_by"`
}

// ContentLength defines the length of content
type ContentLength string

const (
	ContentLengthShort  ContentLength = "short"
	ContentLengthMedium ContentLength = "medium"
	ContentLengthLong   ContentLength = "long"
	ContentLengthCustom ContentLength = "custom"
)

// ContentConstraints represents constraints for content generation
type ContentConstraints struct {
	MaxCharacters    int      `json:"max_characters,omitempty"`
	MaxWords         int      `json:"max_words,omitempty"`
	RequiredKeywords []string `json:"required_keywords,omitempty"`
	ForbiddenWords   []string `json:"forbidden_words,omitempty"`
	RequiredHashtags []string `json:"required_hashtags,omitempty"`
	MaxHashtags      int      `json:"max_hashtags,omitempty"`
	RequiredMentions []string `json:"required_mentions,omitempty"`
	IncludeLink      bool     `json:"include_link"`
	LinkURL          string   `json:"link_url,omitempty"`
	CallToAction     string   `json:"call_to_action,omitempty"`
}

// ContentGenerationResponse represents the response from content generation
type ContentGenerationResponse struct {
	Content     *entities.Content           `json:"content"`
	Variations  []*entities.ContentVariation `json:"variations,omitempty"`
	Suggestions []string                    `json:"suggestions,omitempty"`
	Quality     *ContentQualityAnalysis     `json:"quality,omitempty"`
	Metadata    map[string]interface{}      `json:"metadata,omitempty"`
}

// ContentQualityAnalysis represents content quality analysis
type ContentQualityAnalysis struct {
	OverallScore     float64                 `json:"overall_score"`
	Readability      *ReadabilityAnalysis    `json:"readability,omitempty"`
	Sentiment        *SentimentAnalysis      `json:"sentiment,omitempty"`
	Grammar          *GrammarCheck           `json:"grammar,omitempty"`
	BrandAlignment   float64                 `json:"brand_alignment"`
	EngagementPotential float64              `json:"engagement_potential"`
	PlatformOptimization map[entities.PlatformType]float64 `json:"platform_optimization,omitempty"`
	Keywords         []string                `json:"keywords,omitempty"`
	Entities         []*NamedEntity          `json:"entities,omitempty"`
	Recommendations  []string                `json:"recommendations,omitempty"`
	Warnings         []string                `json:"warnings,omitempty"`
	AnalyzedAt       time.Time               `json:"analyzed_at"`
}

// SentimentAnalysis represents sentiment analysis results
type SentimentAnalysis struct {
	Sentiment   entities.ContentSentiment `json:"sentiment"`
	Score       float64                   `json:"score"`
	Confidence  float64                   `json:"confidence"`
	Emotions    map[string]float64        `json:"emotions,omitempty"`
	Subjectivity float64                  `json:"subjectivity"`
	Polarity    float64                   `json:"polarity"`
}

// ReadabilityAnalysis represents readability analysis results
type ReadabilityAnalysis struct {
	FleschScore      float64 `json:"flesch_score"`
	FleschKincaidGrade float64 `json:"flesch_kincaid_grade"`
	GunningFogIndex  float64 `json:"gunning_fog_index"`
	SMOGIndex        float64 `json:"smog_index"`
	ReadingLevel     string  `json:"reading_level"`
	WordCount        int     `json:"word_count"`
	SentenceCount    int     `json:"sentence_count"`
	SyllableCount    int     `json:"syllable_count"`
	ComplexWords     int     `json:"complex_words"`
}

// GrammarCheck represents grammar check results
type GrammarCheck struct {
	IsCorrect    bool            `json:"is_correct"`
	Errors       []*GrammarError `json:"errors,omitempty"`
	Suggestions  []string        `json:"suggestions,omitempty"`
	CorrectedText string         `json:"corrected_text,omitempty"`
	Score        float64         `json:"score"`
}

// GrammarError represents a grammar error
type GrammarError struct {
	Type        string `json:"type"`
	Message     string `json:"message"`
	Offset      int    `json:"offset"`
	Length      int    `json:"length"`
	Suggestions []string `json:"suggestions,omitempty"`
	Severity    string `json:"severity"`
}

// NamedEntity represents a named entity
type NamedEntity struct {
	Text       string  `json:"text"`
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
	StartPos   int     `json:"start_pos"`
	EndPos     int     `json:"end_pos"`
}

// Image generation types

// ImageGenerationRequest represents a request to generate an image
type ImageGenerationRequest struct {
	Prompt      string                  `json:"prompt"`
	Style       string                  `json:"style,omitempty"`
	Size        string                  `json:"size,omitempty"`
	Quality     string                  `json:"quality,omitempty"`
	Platform    entities.PlatformType   `json:"platform,omitempty"`
	BrandID     uuid.UUID               `json:"brand_id,omitempty"`
	Count       int                     `json:"count,omitempty"`
	Seed        *int64                  `json:"seed,omitempty"`
	Guidance    float64                 `json:"guidance,omitempty"`
	Steps       int                     `json:"steps,omitempty"`
	Model       string                  `json:"model,omitempty"`
}

// ImageGenerationResponse represents the response from image generation
type ImageGenerationResponse struct {
	Images      []GeneratedImage        `json:"images"`
	Metadata    map[string]interface{}  `json:"metadata,omitempty"`
	Cost        float64                 `json:"cost,omitempty"`
	GeneratedAt time.Time               `json:"generated_at"`
}

// GeneratedImage represents a generated image
type GeneratedImage struct {
	URL         string                  `json:"url"`
	Width       int                     `json:"width"`
	Height      int                     `json:"height"`
	Format      string                  `json:"format"`
	Size        int64                   `json:"size"`
	Seed        *int64                  `json:"seed,omitempty"`
	Prompt      string                  `json:"prompt"`
	RevisedPrompt string                `json:"revised_prompt,omitempty"`
	Quality     float64                 `json:"quality,omitempty"`
}

// ImageEditRequest represents a request to edit an image
type ImageEditRequest struct {
	ImageURL    string                  `json:"image_url"`
	Prompt      string                  `json:"prompt"`
	MaskURL     string                  `json:"mask_url,omitempty"`
	Operation   string                  `json:"operation"`
	Strength    float64                 `json:"strength,omitempty"`
	Platform    entities.PlatformType   `json:"platform,omitempty"`
	BrandID     uuid.UUID               `json:"brand_id,omitempty"`
}

// ImageEditResponse represents the response from image editing
type ImageEditResponse struct {
	EditedImage GeneratedImage          `json:"edited_image"`
	Original    string                  `json:"original"`
	Changes     []string                `json:"changes,omitempty"`
	Metadata    map[string]interface{}  `json:"metadata,omitempty"`
	Cost        float64                 `json:"cost,omitempty"`
	EditedAt    time.Time               `json:"edited_at"`
}

// Video generation types

// VideoGenerationRequest represents a request to generate a video
type VideoGenerationRequest struct {
	Prompt      string                  `json:"prompt"`
	Duration    int                     `json:"duration"` // seconds
	Style       string                  `json:"style,omitempty"`
	Quality     string                  `json:"quality,omitempty"`
	Platform    entities.PlatformType   `json:"platform,omitempty"`
	BrandID     uuid.UUID               `json:"brand_id,omitempty"`
	AudioTrack  string                  `json:"audio_track,omitempty"`
	Subtitles   []Subtitle              `json:"subtitles,omitempty"`
	Transitions []string                `json:"transitions,omitempty"`
}

// VideoGenerationResponse represents the response from video generation
type VideoGenerationResponse struct {
	Video       GeneratedVideo          `json:"video"`
	Thumbnail   string                  `json:"thumbnail,omitempty"`
	Metadata    map[string]interface{}  `json:"metadata,omitempty"`
	Cost        float64                 `json:"cost,omitempty"`
	GeneratedAt time.Time               `json:"generated_at"`
}

// GeneratedVideo represents a generated video
type GeneratedVideo struct {
	URL         string                  `json:"url"`
	Width       int                     `json:"width"`
	Height      int                     `json:"height"`
	Duration    int                     `json:"duration"`
	Format      string                  `json:"format"`
	Size        int64                   `json:"size"`
	Bitrate     int                     `json:"bitrate,omitempty"`
	FrameRate   float64                 `json:"frame_rate,omitempty"`
	Prompt      string                  `json:"prompt"`
	Quality     float64                 `json:"quality,omitempty"`
}

// VideoEditRequest represents a request to edit a video
type VideoEditRequest struct {
	VideoURL    string                  `json:"video_url"`
	Operations  []VideoOperation        `json:"operations"`
	Platform    entities.PlatformType   `json:"platform,omitempty"`
	BrandID     uuid.UUID               `json:"brand_id,omitempty"`
}

// VideoOperation represents a video editing operation
type VideoOperation struct {
	Type        string                  `json:"type"`
	StartTime   float64                 `json:"start_time,omitempty"`
	EndTime     float64                 `json:"end_time,omitempty"`
	Parameters  map[string]interface{}  `json:"parameters,omitempty"`
}

// VideoEditResponse represents the response from video editing
type VideoEditResponse struct {
	EditedVideo GeneratedVideo          `json:"edited_video"`
	Original    string                  `json:"original"`
	Operations  []VideoOperation        `json:"operations"`
	Metadata    map[string]interface{}  `json:"metadata,omitempty"`
	Cost        float64                 `json:"cost,omitempty"`
	EditedAt    time.Time               `json:"edited_at"`
}

// Subtitle represents a video subtitle
type Subtitle struct {
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Text      string  `json:"text"`
	Language  string  `json:"language,omitempty"`
}

// NewContentGenerationService creates a new content generation service
func NewContentGenerationService(
	contentRepo repositories.ContentRepository,
	brandRepo repositories.BrandRepository,
	campaignRepo repositories.CampaignRepository,
	aiService AIContentService,
	nlpService NLPService,
	imageService ImageGenerationService,
	videoService VideoGenerationService,
	eventPublisher EventPublisher,
	logger Logger,
) *ContentGenerationService {
	return &ContentGenerationService{
		contentRepo:    contentRepo,
		brandRepo:      brandRepo,
		campaignRepo:   campaignRepo,
		aiService:      aiService,
		nlpService:     nlpService,
		imageService:   imageService,
		videoService:   videoService,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// GenerateContent generates new content using AI
func (cgs *ContentGenerationService) GenerateContent(ctx context.Context, request *ContentGenerationRequest) (*ContentGenerationResponse, error) {
	cgs.logger.Info("Generating content", "brand_id", request.BrandID, "type", request.Type, "topic", request.Topic)

	// Get brand information for context
	brand, err := cgs.brandRepo.GetByID(ctx, request.BrandID)
	if err != nil {
		cgs.logger.Error("Failed to get brand", err, "brand_id", request.BrandID)
		return nil, err
	}

	// Get campaign information if specified
	var campaign *entities.Campaign
	if request.CampaignID != nil {
		campaign, err = cgs.campaignRepo.GetByID(ctx, *request.CampaignID)
		if err != nil {
			cgs.logger.Error("Failed to get campaign", err, "campaign_id", *request.CampaignID)
			return nil, err
		}
	}

	// Enhance request with brand context
	enhancedRequest := cgs.enhanceRequestWithBrandContext(request, brand, campaign)

	// Generate content using AI service
	response, err := cgs.aiService.GenerateContent(ctx, enhancedRequest)
	if err != nil {
		cgs.logger.Error("Failed to generate content", err, "brand_id", request.BrandID)
		return nil, err
	}

	// Enhance generated content
	if err := cgs.aiService.EnhanceContent(ctx, response.Content, brand); err != nil {
		cgs.logger.Warn("Failed to enhance content", "content_id", response.Content.ID, "error", err)
	}

	// Analyze content quality
	quality, err := cgs.aiService.AnalyzeContentQuality(ctx, response.Content)
	if err != nil {
		cgs.logger.Warn("Failed to analyze content quality", "content_id", response.Content.ID, "error", err)
	} else {
		response.Quality = quality
	}

	// Generate platform-specific optimizations
	for _, platform := range request.Platforms {
		optimized, err := cgs.aiService.OptimizeForPlatform(ctx, response.Content, platform)
		if err != nil {
			cgs.logger.Warn("Failed to optimize for platform", "platform", platform, "error", err)
			continue
		}

		// Create variation for platform optimization
		variation := &entities.ContentVariation{
			ID:        uuid.New(),
			ContentID: response.Content.ID,
			Name:      fmt.Sprintf("%s Optimized", strings.Title(string(platform))),
			Body:      optimized.Body,
			Hashtags:  optimized.Hashtags,
			Weight:    1.0,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		response.Variations = append(response.Variations, variation)
	}

	// Save content to repository
	if err := cgs.contentRepo.Create(ctx, response.Content); err != nil {
		cgs.logger.Error("Failed to save generated content", err, "content_id", response.Content.ID)
		return nil, err
	}

	// Save variations
	for _, variation := range response.Variations {
		if err := cgs.contentRepo.AddVariation(ctx, variation); err != nil {
			cgs.logger.Error("Failed to save content variation", err, "variation_id", variation.ID)
		}
	}

	// Publish content generated event
	event := NewContentGeneratedEvent(response.Content, request.CreatedBy)
	if err := cgs.eventPublisher.PublishEvent(ctx, event); err != nil {
		cgs.logger.Error("Failed to publish content generated event", err, "content_id", response.Content.ID)
	}

	cgs.logger.Info("Content generated successfully", "content_id", response.Content.ID, "variations", len(response.Variations))
	return response, nil
}

// enhanceRequestWithBrandContext enhances the generation request with brand context
func (cgs *ContentGenerationService) enhanceRequestWithBrandContext(request *ContentGenerationRequest, brand *entities.Brand, campaign *entities.Campaign) *ContentGenerationRequest {
	enhanced := *request

	// Add brand voice and tone
	if brand.Voice != nil {
		if enhanced.Tone == "" {
			enhanced.Tone = brand.Voice.Tone
		}
		if enhanced.Style == "" {
			enhanced.Style = string(brand.Voice.Style)
		}
		
		// Add brand keywords
		enhanced.Keywords = append(enhanced.Keywords, brand.Keywords...)
		
		// Add brand-specific constraints
		if enhanced.Constraints == nil {
			enhanced.Constraints = &ContentConstraints{}
		}
		
		if brand.Voice.KeyPhrases != nil {
			enhanced.Constraints.RequiredKeywords = append(enhanced.Constraints.RequiredKeywords, brand.Voice.KeyPhrases...)
		}
		
		if brand.Voice.AvoidPhrases != nil {
			enhanced.Constraints.ForbiddenWords = append(enhanced.Constraints.ForbiddenWords, brand.Voice.AvoidPhrases...)
		}
	}

	// Add campaign context
	if campaign != nil {
		enhanced.Keywords = append(enhanced.Keywords, campaign.Keywords...)
		if campaign.TargetAudience != nil && enhanced.TargetAudience == nil {
			enhanced.TargetAudience = campaign.TargetAudience
		}
		if enhanced.Tone == "" {
			enhanced.Tone = campaign.Tone
		}
	}

	return &enhanced
}

// GenerateVariations generates variations of existing content
func (cgs *ContentGenerationService) GenerateVariations(ctx context.Context, content *entities.Content, count int) ([]*entities.ContentVariation, error) {
	cgs.logger.Info("Generating content variations", "content_id", content.ID, "count", count)

	// Get brand for context
	brand, err := cgs.brandRepo.GetByID(ctx, content.BrandID)
	if err != nil {
		cgs.logger.Error("Failed to get brand for variations", err, "brand_id", content.BrandID)
		return nil, err
	}

	// Use AI service to generate variations
	variations, err := cgs.aiService.GenerateVariations(ctx, content, count)
	if err != nil {
		cgs.logger.Error("Failed to generate variations using AI service", err, "content_id", content.ID)
		return nil, err
	}

	// Enhance each variation with brand context
	for _, variation := range variations {
		if err := cgs.applyBrandGuidelinesToVariation(variation, brand); err != nil {
			cgs.logger.Warn("Failed to apply brand guidelines to variation", "variation_id", variation.ID, "error", err)
		}
	}

	cgs.logger.Info("Content variations generated successfully", "content_id", content.ID, "variations_count", len(variations))
	return variations, nil
}

// EnhanceContent enhances existing content using AI
func (cgs *ContentGenerationService) EnhanceContent(ctx context.Context, content *entities.Content, brand *entities.Brand) error {
	cgs.logger.Info("Enhancing content", "content_id", content.ID, "brand_id", brand.ID)

	// Use AI service to enhance content
	if err := cgs.aiService.EnhanceContent(ctx, content, brand); err != nil {
		cgs.logger.Error("Failed to enhance content using AI service", err, "content_id", content.ID)
		return err
	}

	// Apply brand guidelines after enhancement
	if err := cgs.ApplyBrandGuidelines(content, brand); err != nil {
		cgs.logger.Warn("Failed to apply brand guidelines after enhancement", "content_id", content.ID, "error", err)
	}

	cgs.logger.Info("Content enhanced successfully", "content_id", content.ID)
	return nil
}

// AnalyzeContentQuality analyzes the quality of content
func (cgs *ContentGenerationService) AnalyzeContentQuality(ctx context.Context, content *entities.Content) (*ContentQualityAnalysis, error) {
	cgs.logger.Info("Analyzing content quality", "content_id", content.ID)

	// Use AI service to analyze quality
	quality, err := cgs.aiService.AnalyzeContentQuality(ctx, content)
	if err != nil {
		cgs.logger.Error("Failed to analyze content quality", err, "content_id", content.ID)
		return nil, err
	}

	cgs.logger.Info("Content quality analyzed successfully", "content_id", content.ID, "overall_score", quality.OverallScore)
	return quality, nil
}

// OptimizeForPlatform optimizes content for a specific platform
func (cgs *ContentGenerationService) OptimizeForPlatform(ctx context.Context, content *entities.Content, platform entities.PlatformType) (*entities.Content, error) {
	cgs.logger.Info("Optimizing content for platform", "content_id", content.ID, "platform", platform)

	// Use AI service to optimize for platform
	optimized, err := cgs.aiService.OptimizeForPlatform(ctx, content, platform)
	if err != nil {
		cgs.logger.Error("Failed to optimize content for platform", err, "content_id", content.ID, "platform", platform)
		return nil, err
	}

	cgs.logger.Info("Content optimized for platform successfully", "content_id", content.ID, "platform", platform)
	return optimized, nil
}

// GenerateHashtags generates hashtags for content
func (cgs *ContentGenerationService) GenerateHashtags(ctx context.Context, content *entities.Content, platform entities.PlatformType, count int) ([]string, error) {
	cgs.logger.Info("Generating hashtags", "content_id", content.ID, "platform", platform, "count", count)

	// Use AI service to generate hashtags
	hashtags, err := cgs.aiService.GenerateHashtags(ctx, content, platform, count)
	if err != nil {
		cgs.logger.Error("Failed to generate hashtags", err, "content_id", content.ID, "platform", platform)
		return nil, err
	}

	cgs.logger.Info("Hashtags generated successfully", "content_id", content.ID, "hashtags_count", len(hashtags))
	return hashtags, nil
}

// Helper methods

// ApplyBrandGuidelines applies brand guidelines to content
func (cgs *ContentGenerationService) ApplyBrandGuidelines(content *entities.Content, brand *entities.Brand) error {
	cgs.logger.Debug("Applying brand guidelines", "content_id", content.ID, "brand_id", brand.ID)

	// Apply brand voice
	if brand.Voice != nil {
		if content.Tone == "" {
			content.Tone = brand.Voice.Tone
		}
		
		// Apply brand key phrases if content body doesn't already contain them
		if len(brand.Voice.KeyPhrases) > 0 {
			cgs.integrateKeyPhrases(content, brand.Voice.KeyPhrases)
		}
		
		// Ensure avoid phrases are not used in content
		if len(brand.Voice.AvoidPhrases) > 0 {
			cgs.removeAvoidPhrases(content, brand.Voice.AvoidPhrases)
		}
	}

	// Apply brand keywords
	if len(brand.Keywords) > 0 {
		// Deduplicate keywords before adding
		existingKeywords := make(map[string]bool)
		for _, keyword := range content.Keywords {
			existingKeywords[keyword] = true
		}
		
		for _, keyword := range brand.Keywords {
			if !existingKeywords[keyword] {
				content.Keywords = append(content.Keywords, keyword)
				existingKeywords[keyword] = true
			}
		}
	}

	// Apply brand language
	if content.Language == "" && len(brand.Languages) > 0 {
		content.Language = brand.Languages[0]
	}

	cgs.logger.Debug("Brand guidelines applied successfully", "content_id", content.ID)
	return nil
}

// integrateKeyPhrases integrates brand key phrases into content naturally
func (cgs *ContentGenerationService) integrateKeyPhrases(content *entities.Content, keyPhrases []string) {
	// Simple implementation - could be enhanced with AI for better integration
	for _, phrase := range keyPhrases {
		if !cgs.containsPhrase(content.Body, phrase) {
			// Add key phrases to content keywords for now
			// In a real implementation, this could use AI to naturally integrate phrases
			content.Keywords = append(content.Keywords, phrase)
		}
	}
}

// removeAvoidPhrases removes brand avoid phrases from content
func (cgs *ContentGenerationService) removeAvoidPhrases(content *entities.Content, avoidPhrases []string) {
	for _, phrase := range avoidPhrases {
		if cgs.containsPhrase(content.Body, phrase) {
			cgs.logger.Warn("Content contains avoid phrase", "content_id", content.ID, "phrase", phrase)
			// In a real implementation, this could use AI to rephrase content
			// For now, just log the warning
		}
	}
}

// containsPhrase checks if content contains a specific phrase (case-insensitive)
func (cgs *ContentGenerationService) containsPhrase(text, phrase string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(phrase))
}

func (cgs *ContentGenerationService) applyBrandGuidelinesToVariation(variation *entities.ContentVariation, brand *entities.Brand) error {
	// Apply brand voice to variation by incorporating brand keywords into hashtags
	if brand.Voice != nil && len(brand.Keywords) > 0 {
		// Add brand keywords as hashtags if they're not already present
		existingHashtags := make(map[string]bool)
		for _, hashtag := range variation.Hashtags {
			existingHashtags[hashtag] = true
		}
		
		// Add brand keywords as hashtags
		for _, keyword := range brand.Keywords {
			hashtagKeyword := "#" + keyword
			if !existingHashtags[hashtagKeyword] && !existingHashtags[keyword] {
				variation.Hashtags = append(variation.Hashtags, keyword)
			}
		}
	}

	return nil
}
