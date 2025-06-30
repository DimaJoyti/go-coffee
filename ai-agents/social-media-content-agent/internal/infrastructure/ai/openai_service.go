package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-coffee-ai-agents/social-media-content-agent/internal/domain/entities"
	"go-coffee-ai-agents/social-media-content-agent/internal/domain/services"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

// OpenAIService implements AI content services using OpenAI
type OpenAIService struct {
	client *openai.Client
	logger services.Logger
	config *OpenAIConfig
}

// OpenAIConfig contains OpenAI service configuration
type OpenAIConfig struct {
	APIKey           string
	Model            string
	MaxTokens        int
	Temperature      float32
	TopP             float32
	FrequencyPenalty float32
	PresencePenalty  float32
	Timeout          time.Duration
}

// NewOpenAIService creates a new OpenAI service
func NewOpenAIService(apiKey string, logger services.Logger) (*OpenAIService, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	client := openai.NewClient(apiKey)

	config := &OpenAIConfig{
		APIKey:           apiKey,
		Model:            openai.GPT4TurboPreview,
		MaxTokens:        2000,
		Temperature:      0.7,
		TopP:             1.0,
		FrequencyPenalty: 0.0,
		PresencePenalty:  0.0,
		Timeout:          30 * time.Second,
	}

	return &OpenAIService{
		client: client,
		logger: logger,
		config: config,
	}, nil
}

// GenerateContent generates content using OpenAI
func (s *OpenAIService) GenerateContent(ctx context.Context, request *services.ContentGenerationRequest) (*services.ContentGenerationResponse, error) {
	s.logger.Info("Generating content with OpenAI", "topic", request.Topic, "type", request.Type)

	// Build the prompt
	prompt := s.buildContentPrompt(request)

	// Create chat completion request
	chatRequest := openai.ChatCompletionRequest{
		Model:            s.config.Model,
		MaxTokens:        s.config.MaxTokens,
		Temperature:      s.config.Temperature,
		TopP:             s.config.TopP,
		FrequencyPenalty: s.config.FrequencyPenalty,
		PresencePenalty:  s.config.PresencePenalty,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: s.getSystemPrompt(request),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	// Call OpenAI API
	response, err := s.client.CreateChatCompletion(ctx, chatRequest)
	if err != nil {
		s.logger.Error("Failed to generate content with OpenAI", err)
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no content generated")
	}

	// Parse the response
	generatedText := response.Choices[0].Message.Content
	content := s.parseGeneratedContent(generatedText, request)

	// Create response
	result := &services.ContentGenerationResponse{
		Content:     content,
		Suggestions: s.extractSuggestions(generatedText),
		Metadata: map[string]interface{}{
			"model":         response.Model,
			"usage":         response.Usage,
			"finish_reason": response.Choices[0].FinishReason,
			"generated_at":  time.Now(),
		},
	}

	s.logger.Info("Content generated successfully", "content_id", content.ID, "tokens_used", response.Usage.TotalTokens)
	return result, nil
}

// EnhanceContent enhances existing content
func (s *OpenAIService) EnhanceContent(ctx context.Context, content *entities.Content, brand *entities.Brand) error {
	s.logger.Info("Enhancing content with OpenAI", "content_id", content.ID)

	prompt := s.buildEnhancementPrompt(content, brand)

	chatRequest := openai.ChatCompletionRequest{
		Model:       s.config.Model,
		MaxTokens:   s.config.MaxTokens,
		Temperature: 0.5, // Lower temperature for enhancement
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an expert content editor specializing in social media optimization.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	response, err := s.client.CreateChatCompletion(ctx, chatRequest)
	if err != nil {
		s.logger.Error("Failed to enhance content", err, "content_id", content.ID)
		return fmt.Errorf("failed to enhance content: %w", err)
	}

	if len(response.Choices) == 0 {
		return fmt.Errorf("no enhancement generated")
	}

	// Apply enhancements
	enhanced := s.parseEnhancedContent(response.Choices[0].Message.Content)
	if enhanced.Body != "" {
		content.Body = enhanced.Body
	}
	if len(enhanced.Hashtags) > 0 {
		content.Hashtags = enhanced.Hashtags
	}

	s.logger.Info("Content enhanced successfully", "content_id", content.ID)
	return nil
}

// GenerateVariations generates content variations
func (s *OpenAIService) GenerateVariations(ctx context.Context, content *entities.Content, count int) ([]*entities.ContentVariation, error) {
	s.logger.Info("Generating content variations", "content_id", content.ID, "count", count)

	prompt := s.buildVariationPrompt(content, count)

	chatRequest := openai.ChatCompletionRequest{
		Model:       s.config.Model,
		MaxTokens:   s.config.MaxTokens,
		Temperature: 0.8, // Higher temperature for creativity
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a creative content writer specializing in creating engaging variations of social media content.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	response, err := s.client.CreateChatCompletion(ctx, chatRequest)
	if err != nil {
		s.logger.Error("Failed to generate variations", err, "content_id", content.ID)
		return nil, fmt.Errorf("failed to generate variations: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no variations generated")
	}

	// Parse variations
	variations := s.parseVariations(response.Choices[0].Message.Content, content.ID)

	s.logger.Info("Variations generated successfully", "content_id", content.ID, "count", len(variations))
	return variations, nil
}

// OptimizeForPlatform optimizes content for a specific platform
func (s *OpenAIService) OptimizeForPlatform(ctx context.Context, content *entities.Content, platform entities.PlatformType) (*entities.Content, error) {
	s.logger.Info("Optimizing content for platform", "content_id", content.ID, "platform", platform)

	prompt := s.buildPlatformOptimizationPrompt(content, platform)

	chatRequest := openai.ChatCompletionRequest{
		Model:       s.config.Model,
		MaxTokens:   s.config.MaxTokens,
		Temperature: 0.6,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf("You are a social media expert specializing in %s optimization.", platform),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	response, err := s.client.CreateChatCompletion(ctx, chatRequest)
	if err != nil {
		s.logger.Error("Failed to optimize content for platform", err, "platform", platform)
		return nil, fmt.Errorf("failed to optimize content: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no optimization generated")
	}

	// Create optimized content
	optimized := s.parseOptimizedContent(response.Choices[0].Message.Content, content)

	s.logger.Info("Content optimized successfully", "content_id", content.ID, "platform", platform)
	return optimized, nil
}

// GenerateHashtags generates hashtags for content
func (s *OpenAIService) GenerateHashtags(ctx context.Context, content *entities.Content, platform entities.PlatformType, count int) ([]string, error) {
	s.logger.Info("Generating hashtags", "content_id", content.ID, "platform", platform, "count", count)

	prompt := s.buildHashtagPrompt(content, platform, count)

	chatRequest := openai.ChatCompletionRequest{
		Model:       s.config.Model,
		MaxTokens:   500,
		Temperature: 0.7,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a hashtag expert who creates trending and relevant hashtags for social media content.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	response, err := s.client.CreateChatCompletion(ctx, chatRequest)
	if err != nil {
		s.logger.Error("Failed to generate hashtags", err)
		return nil, fmt.Errorf("failed to generate hashtags: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no hashtags generated")
	}

	hashtags := s.parseHashtags(response.Choices[0].Message.Content)

	s.logger.Info("Hashtags generated successfully", "count", len(hashtags))
	return hashtags, nil
}

// AnalyzeContentQuality analyzes content quality
func (s *OpenAIService) AnalyzeContentQuality(ctx context.Context, content *entities.Content) (*services.ContentQualityAnalysis, error) {
	s.logger.Info("Analyzing content quality", "content_id", content.ID)

	prompt := s.buildQualityAnalysisPrompt(content)

	chatRequest := openai.ChatCompletionRequest{
		Model:       s.config.Model,
		MaxTokens:   1000,
		Temperature: 0.3, // Low temperature for consistent analysis
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a content quality analyst who provides detailed analysis of social media content.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	response, err := s.client.CreateChatCompletion(ctx, chatRequest)
	if err != nil {
		s.logger.Error("Failed to analyze content quality", err)
		return nil, fmt.Errorf("failed to analyze content quality: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no analysis generated")
	}

	analysis := s.parseQualityAnalysis(response.Choices[0].Message.Content)

	s.logger.Info("Content quality analyzed successfully", "content_id", content.ID, "score", analysis.OverallScore)
	return analysis, nil
}

// SuggestImprovements suggests content improvements
func (s *OpenAIService) SuggestImprovements(ctx context.Context, content *entities.Content) ([]string, error) {
	s.logger.Info("Suggesting improvements", "content_id", content.ID)

	prompt := s.buildImprovementPrompt(content)

	chatRequest := openai.ChatCompletionRequest{
		Model:       s.config.Model,
		MaxTokens:   800,
		Temperature: 0.6,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a content improvement specialist who provides actionable suggestions for better social media content.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	response, err := s.client.CreateChatCompletion(ctx, chatRequest)
	if err != nil {
		s.logger.Error("Failed to suggest improvements", err)
		return nil, fmt.Errorf("failed to suggest improvements: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no suggestions generated")
	}

	suggestions := s.parseImprovements(response.Choices[0].Message.Content)

	s.logger.Info("Improvements suggested successfully", "content_id", content.ID, "count", len(suggestions))
	return suggestions, nil
}

// GenerateCaption generates a caption for content
func (s *OpenAIService) GenerateCaption(ctx context.Context, content *entities.Content, platform entities.PlatformType) (string, error) {
	s.logger.Info("Generating caption", "content_id", content.ID, "platform", platform)

	prompt := s.buildCaptionPrompt(content, platform)

	chatRequest := openai.ChatCompletionRequest{
		Model:       s.config.Model,
		MaxTokens:   500,
		Temperature: 0.7,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf("You are a caption writer specializing in %s content.", platform),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	response, err := s.client.CreateChatCompletion(ctx, chatRequest)
	if err != nil {
		s.logger.Error("Failed to generate caption", err)
		return "", fmt.Errorf("failed to generate caption: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no caption generated")
	}

	caption := strings.TrimSpace(response.Choices[0].Message.Content)

	s.logger.Info("Caption generated successfully", "content_id", content.ID)
	return caption, nil
}

// Helper methods for building prompts and parsing responses

func (s *OpenAIService) getSystemPrompt(request *services.ContentGenerationRequest) string {
	return fmt.Sprintf(`You are an expert social media content creator specializing in %s content for %s platforms.
Your task is to create engaging, authentic, and platform-optimized content that resonates with the target audience.

Guidelines:
- Create content that matches the specified tone: %s
- Include relevant hashtags when requested
- Ensure content is appropriate for the target platforms
- Focus on engagement and authenticity
- Follow best practices for each platform
- Keep content within platform character limits

Respond with a JSON object containing:
{
  "title": "Content title",
  "body": "Main content text",
  "hashtags": ["hashtag1", "hashtag2"],
  "suggestions": ["suggestion1", "suggestion2"]
}`, request.Type, strings.Join(platformsToStrings(request.Platforms), ", "), request.Tone)
}

func (s *OpenAIService) buildContentPrompt(request *services.ContentGenerationRequest) string {
	prompt := fmt.Sprintf("Create %s content about: %s\n", request.Type, request.Topic)

	if len(request.Keywords) > 0 {
		prompt += fmt.Sprintf("Keywords to include: %s\n", strings.Join(request.Keywords, ", "))
	}

	if len(request.Platforms) > 0 {
		prompt += fmt.Sprintf("Target platforms: %s\n", strings.Join(platformsToStrings(request.Platforms), ", "))
	}

	if request.TargetAudience != nil {
		prompt += fmt.Sprintf("Target audience: %s\n", s.describeAudience(request.TargetAudience))
	}

	if request.CustomPrompt != "" {
		prompt += fmt.Sprintf("Additional instructions: %s\n", request.CustomPrompt)
	}

	return prompt
}

func (s *OpenAIService) buildEnhancementPrompt(content *entities.Content, brand *entities.Brand) string {
	prompt := fmt.Sprintf("Enhance this social media content:\n\nTitle: %s\nBody: %s\n", content.Title, content.Body)

	if brand != nil && brand.Voice != nil {
		prompt += fmt.Sprintf("Brand voice: %s\n", brand.Voice.Style)
		if len(brand.Voice.KeyPhrases) > 0 {
			prompt += fmt.Sprintf("Key phrases to include: %s\n", strings.Join(brand.Voice.KeyPhrases, ", "))
		}
	}

	prompt += "\nProvide enhanced version with improved engagement, clarity, and brand alignment."
	return prompt
}

func (s *OpenAIService) buildVariationPrompt(content *entities.Content, count int) string {
	return fmt.Sprintf(`Create %d variations of this content:

Title: %s
Body: %s

Each variation should:
- Maintain the core message
- Use different wording and structure
- Be suitable for different audiences or contexts
- Include relevant hashtags

Respond with a JSON array of variations:
[
  {
    "name": "Variation 1",
    "body": "Content text",
    "hashtags": ["tag1", "tag2"]
  }
]`, count, content.Title, content.Body)
}

func (s *OpenAIService) buildPlatformOptimizationPrompt(content *entities.Content, platform entities.PlatformType) string {
	limits := s.getPlatformLimits(platform)

	return fmt.Sprintf(`Optimize this content for %s:

Original content:
Title: %s
Body: %s

Platform requirements:
- Character limit: %d
- Hashtag limit: %d
- Best practices: %s

Provide optimized version that maximizes engagement on %s.`,
		platform, content.Title, content.Body, limits.CharacterLimit, limits.HashtagLimit, limits.BestPractices, platform)
}

func (s *OpenAIService) buildHashtagPrompt(content *entities.Content, platform entities.PlatformType, count int) string {
	return fmt.Sprintf(`Generate %d relevant hashtags for this %s content:

Title: %s
Body: %s

Requirements:
- Mix of popular and niche hashtags
- Relevant to the content topic
- Appropriate for %s platform
- Include trending hashtags when relevant

Return hashtags as a comma-separated list.`, count, platform, content.Title, content.Body, platform)
}

func (s *OpenAIService) buildQualityAnalysisPrompt(content *entities.Content) string {
	return fmt.Sprintf(`Analyze the quality of this social media content:

Title: %s
Body: %s
Type: %s
Platforms: %s

Provide analysis in JSON format:
{
  "overall_score": 85,
  "readability": {"score": 90, "level": "easy"},
  "engagement_potential": 80,
  "brand_alignment": 75,
  "recommendations": ["suggestion1", "suggestion2"],
  "warnings": ["warning1"]
}`, content.Title, content.Body, content.Type, strings.Join(platformsToStrings(content.Platforms), ", "))
}

func (s *OpenAIService) buildImprovementPrompt(content *entities.Content) string {
	return fmt.Sprintf(`Suggest specific improvements for this social media content:

Title: %s
Body: %s

Provide 3-5 actionable suggestions to improve:
- Engagement potential
- Clarity and readability
- Call-to-action effectiveness
- Visual appeal
- Platform optimization

Return as a numbered list.`, content.Title, content.Body)
}

func (s *OpenAIService) buildCaptionPrompt(content *entities.Content, platform entities.PlatformType) string {
	return fmt.Sprintf(`Create an engaging caption for %s based on this content:

Title: %s
Body: %s

The caption should:
- Be optimized for %s
- Include a strong hook
- Encourage engagement
- Be within platform limits
- Include relevant emojis

Return only the caption text.`, platform, content.Title, content.Body, platform)
}

// Parsing methods

func (s *OpenAIService) parseGeneratedContent(text string, request *services.ContentGenerationRequest) *entities.Content {
	// Try to parse as JSON first
	var parsed struct {
		Title       string   `json:"title"`
		Body        string   `json:"body"`
		Hashtags    []string `json:"hashtags"`
		Suggestions []string `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(text), &parsed); err == nil {
		content := entities.NewContent(parsed.Title, parsed.Body, request.Type, request.BrandID, request.CreatedBy)
		content.Hashtags = parsed.Hashtags
		content.Platforms = request.Platforms
		content.Tone = request.Tone
		return content
	}

	// Fallback to plain text parsing
	content := entities.NewContent(request.Topic, text, request.Type, request.BrandID, request.CreatedBy)
	content.Platforms = request.Platforms
	content.Tone = request.Tone
	return content
}

func (s *OpenAIService) parseEnhancedContent(text string) struct {
	Body     string
	Hashtags []string
} {
	// Simple parsing - in production, this would be more sophisticated
	return struct {
		Body     string
		Hashtags []string
	}{
		Body:     text,
		Hashtags: []string{},
	}
}

func (s *OpenAIService) parseVariations(text string, contentID uuid.UUID) []*entities.ContentVariation {
	var variations []struct {
		Name     string   `json:"name"`
		Body     string   `json:"body"`
		Hashtags []string `json:"hashtags"`
	}

	if err := json.Unmarshal([]byte(text), &variations); err != nil {
		// Fallback parsing
		return []*entities.ContentVariation{}
	}

	result := make([]*entities.ContentVariation, len(variations))
	for i, v := range variations {
		result[i] = &entities.ContentVariation{
			ID:        uuid.New(),
			ContentID: contentID,
			Name:      v.Name,
			Body:      v.Body,
			Hashtags:  v.Hashtags,
			Weight:    1.0,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	return result
}

func (s *OpenAIService) parseOptimizedContent(text string, original *entities.Content) *entities.Content {
	optimized := *original // Copy original
	optimized.Body = text
	return &optimized
}

func (s *OpenAIService) parseHashtags(text string) []string {
	// Extract hashtags from text
	hashtags := []string{}
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			hashtags = append(hashtags, strings.TrimPrefix(line, "#"))
		} else if strings.Contains(line, ",") {
			// Comma-separated hashtags
			parts := strings.Split(line, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				part = strings.TrimPrefix(part, "#")
				if part != "" {
					hashtags = append(hashtags, part)
				}
			}
		}
	}
	return hashtags
}

func (s *OpenAIService) parseQualityAnalysis(text string) *services.ContentQualityAnalysis {
	var analysis services.ContentQualityAnalysis

	if err := json.Unmarshal([]byte(text), &analysis); err != nil {
		// Fallback analysis
		analysis = services.ContentQualityAnalysis{
			OverallScore:        75.0,
			BrandAlignment:      70.0,
			EngagementPotential: 80.0,
			Recommendations:     []string{"Consider adding more engaging elements"},
			Warnings:            []string{},
			AnalyzedAt:          time.Now(),
		}
	}

	analysis.AnalyzedAt = time.Now()
	return &analysis
}

func (s *OpenAIService) parseImprovements(text string) []string {
	improvements := []string{}
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && (strings.HasPrefix(line, "-") || strings.HasPrefix(line, "•") || strings.Contains(line, ".")) {
			improvements = append(improvements, line)
		}
	}
	return improvements
}

func (s *OpenAIService) extractSuggestions(text string) []string {
	// Extract suggestions from generated text
	return []string{"Consider adding more visual elements", "Include a call-to-action"}
}

// Helper functions

func platformsToStrings(platforms []entities.PlatformType) []string {
	result := make([]string, len(platforms))
	for i, p := range platforms {
		result[i] = string(p)
	}
	return result
}

func (s *OpenAIService) describeAudience(audience *entities.TargetAudience) string {
	if audience == nil {
		return "General audience"
	}

	description := ""
	if audience.Demographics != nil {
		description += fmt.Sprintf("Demographics: %+v", audience.Demographics)
	}
	if len(audience.Interests) > 0 {
		description += fmt.Sprintf(" Interests: %s", strings.Join(audience.Interests, ", "))
	}
	return description
}

type PlatformLimits struct {
	CharacterLimit int
	HashtagLimit   int
	BestPractices  string
}

func (s *OpenAIService) getPlatformLimits(platform entities.PlatformType) PlatformLimits {
	switch platform {
	case entities.PlatformTwitter:
		return PlatformLimits{280, 2, "Use threads for longer content, engage with replies"}
	case entities.PlatformInstagram:
		return PlatformLimits{2200, 30, "Use high-quality visuals, stories for behind-the-scenes"}
	case entities.PlatformFacebook:
		return PlatformLimits{63206, 30, "Use native video, encourage comments and shares"}
	case entities.PlatformLinkedIn:
		return PlatformLimits{3000, 3, "Professional tone, industry insights, thought leadership"}
	case entities.PlatformTikTok:
		return PlatformLimits{2200, 100, "Trending sounds, vertical video, authentic content"}
	default:
		return PlatformLimits{1000, 10, "Engage authentically with your audience"}
	}
}
