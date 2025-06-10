package content

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/ai"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/common"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/redis"
)

// Analyzer provides content analysis capabilities
type Analyzer struct {
	config    config.ContentAnalysisConfig
	logger    *logger.Logger
	aiService ai.Service
	cache     redis.Client
}

// NewAnalyzer creates a new content analyzer
func NewAnalyzer(cfg config.ContentAnalysisConfig, logger *logger.Logger, aiService ai.Service, cache redis.Client) *Analyzer {
	return &Analyzer{
		config:    cfg,
		logger:    logger,
		aiService: aiService,
		cache:     cache,
	}
}

// AnalyzePost analyzes a Reddit post for classification, sentiment, and topics
func (a *Analyzer) AnalyzePost(ctx context.Context, post *common.RedditPost) (*common.ContentClassification, error) {
	a.logger.Info("Analyzing post", zap.String("post_id", post.ID))

	// Check cache first
	cacheKey := fmt.Sprintf("analysis:post:%s", post.ID)
	if cached, err := a.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var classification common.ContentClassification
		if err := json.Unmarshal([]byte(cached), &classification); err == nil {
			a.logger.Debug("Cache hit for post analysis", zap.String("post_id", post.ID))
			return &classification, nil
		}
	}

	// Prepare content for analysis
	content := a.prepareContentForAnalysis(post.Title, post.Content)
	if content == "" {
		return nil, fmt.Errorf("no content to analyze")
	}

	// Perform analysis
	classification := &common.ContentClassification{
		ID:          generateID(),
		ContentID:   post.ID,
		ContentType: "post",
		ProcessedAt: time.Now(),
		ModelUsed:   a.config.ClassificationModel,
		Metadata:    make(map[string]string),
	}

	// Classify content
	if err := a.classifyContent(ctx, content, classification); err != nil {
		a.logger.Error("Failed to classify content", zap.Error(err))
		return nil, err
	}

	// Analyze sentiment
	if a.config.SentimentAnalysis {
		if err := a.analyzeSentiment(ctx, content, classification); err != nil {
			a.logger.Error("Failed to analyze sentiment", zap.Error(err))
		}
	}

	// Extract topics
	if a.config.TopicModeling {
		if err := a.extractTopics(ctx, content, classification); err != nil {
			a.logger.Error("Failed to extract topics", zap.Error(err))
		}
	}

	// Cache result
	if data, err := json.Marshal(classification); err == nil {
		if err := a.cache.Set(ctx, cacheKey, data, 24*time.Hour); err != nil {
			a.logger.Warn("Failed to cache analysis result", zap.Error(err))
		}
	}

	a.logger.Info("Post analysis completed",
		zap.String("post_id", post.ID),
		zap.String("category", classification.Category),
		zap.Float64("confidence", classification.Confidence))

	return classification, nil
}

// AnalyzeComment analyzes a Reddit comment
func (a *Analyzer) AnalyzeComment(ctx context.Context, comment *common.RedditComment) (*common.ContentClassification, error) {
	a.logger.Info("Analyzing comment", zap.String("comment_id", comment.ID))

	// Check cache first
	cacheKey := fmt.Sprintf("analysis:comment:%s", comment.ID)
	if cached, err := a.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var classification common.ContentClassification
		if err := json.Unmarshal([]byte(cached), &classification); err == nil {
			a.logger.Debug("Cache hit for comment analysis", zap.String("comment_id", comment.ID))
			return &classification, nil
		}
	}

	// Prepare content for analysis
	content := strings.TrimSpace(comment.Content)
	if content == "" {
		return nil, fmt.Errorf("no content to analyze")
	}

	// Perform analysis
	classification := &common.ContentClassification{
		ID:          generateID(),
		ContentID:   comment.ID,
		ContentType: "comment",
		ProcessedAt: time.Now(),
		ModelUsed:   a.config.ClassificationModel,
		Metadata:    make(map[string]string),
	}

	// Classify content
	if err := a.classifyContent(ctx, content, classification); err != nil {
		a.logger.Error("Failed to classify content", zap.Error(err))
		return nil, err
	}

	// Analyze sentiment
	if a.config.SentimentAnalysis {
		if err := a.analyzeSentiment(ctx, content, classification); err != nil {
			a.logger.Error("Failed to analyze sentiment", zap.Error(err))
		}
	}

	// Extract topics
	if a.config.TopicModeling {
		if err := a.extractTopics(ctx, content, classification); err != nil {
			a.logger.Error("Failed to extract topics", zap.Error(err))
		}
	}

	// Cache result
	if data, err := json.Marshal(classification); err == nil {
		if err := a.cache.Set(ctx, cacheKey, data, 24*time.Hour); err != nil {
			a.logger.Warn("Failed to cache analysis result", zap.Error(err))
		}
	}

	a.logger.Info("Comment analysis completed",
		zap.String("comment_id", comment.ID),
		zap.String("category", classification.Category),
		zap.Float64("confidence", classification.Confidence))

	return classification, nil
}

// classifyContent classifies content into categories
func (a *Analyzer) classifyContent(ctx context.Context, content string, classification *common.ContentClassification) error {
	prompt := a.buildClassificationPrompt(content)

	req := &ai.GenerateRequest{
		UserID:      "content_analyzer",
		Message:     prompt,
		Temperature: 0.1, // Low temperature for consistent classification
		MaxTokens:   500,
		Metadata: map[string]string{
			"task":         "classification",
			"content_type": classification.ContentType,
		},
	}

	resp, err := a.aiService.GenerateResponse(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to generate classification: %w", err)
	}

	// Parse classification response
	if err := a.parseClassificationResponse(resp.Text, classification); err != nil {
		return fmt.Errorf("failed to parse classification response: %w", err)
	}

	return nil
}

// analyzeSentiment analyzes sentiment of content
func (a *Analyzer) analyzeSentiment(ctx context.Context, content string, classification *common.ContentClassification) error {
	prompt := a.buildSentimentPrompt(content)

	req := &ai.GenerateRequest{
		UserID:      "sentiment_analyzer",
		Message:     prompt,
		Temperature: 0.1,
		MaxTokens:   200,
		Metadata: map[string]string{
			"task": "sentiment_analysis",
		},
	}

	resp, err := a.aiService.GenerateResponse(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to generate sentiment analysis: %w", err)
	}

	// Parse sentiment response
	if err := a.parseSentimentResponse(resp.Text, classification); err != nil {
		return fmt.Errorf("failed to parse sentiment response: %w", err)
	}

	return nil
}

// extractTopics extracts topics from content
func (a *Analyzer) extractTopics(ctx context.Context, content string, classification *common.ContentClassification) error {
	prompt := a.buildTopicExtractionPrompt(content)

	req := &ai.GenerateRequest{
		UserID:      "topic_extractor",
		Message:     prompt,
		Temperature: 0.2,
		MaxTokens:   300,
		Metadata: map[string]string{
			"task": "topic_extraction",
		},
	}

	resp, err := a.aiService.GenerateResponse(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to generate topic extraction: %w", err)
	}

	// Parse topics response
	if err := a.parseTopicsResponse(resp.Text, classification); err != nil {
		return fmt.Errorf("failed to parse topics response: %w", err)
	}

	return nil
}

// prepareContentForAnalysis prepares content for analysis
func (a *Analyzer) prepareContentForAnalysis(title, content string) string {
	var parts []string

	if title != "" {
		parts = append(parts, "Title: "+title)
	}

	if content != "" {
		parts = append(parts, "Content: "+content)
	}

	return strings.Join(parts, "\n\n")
}

// buildClassificationPrompt builds prompt for content classification
func (a *Analyzer) buildClassificationPrompt(content string) string {
	categories := strings.Join(a.config.Categories, ", ")

	return fmt.Sprintf(`Analyze and classify the following content into one of these categories: %s

Content:
%s

Please respond with a JSON object containing:
- "category": the main category
- "subcategory": a more specific subcategory if applicable
- "tags": array of relevant tags
- "confidence": confidence score between 0 and 1
- "reasoning": brief explanation of the classification

Response:`, categories, content)
}

// buildSentimentPrompt builds prompt for sentiment analysis
func (a *Analyzer) buildSentimentPrompt(content string) string {
	return fmt.Sprintf(`Analyze the sentiment of the following content:

Content:
%s

Please respond with a JSON object containing:
- "label": sentiment label (positive, negative, neutral)
- "score": confidence score between 0 and 1
- "magnitude": intensity of sentiment between 0 and 1
- "subjectivity": objectivity vs subjectivity between 0 and 1

Response:`, content)
}

// buildTopicExtractionPrompt builds prompt for topic extraction
func (a *Analyzer) buildTopicExtractionPrompt(content string) string {
	return fmt.Sprintf(`Extract the main topics and themes from the following content:

Content:
%s

Please respond with a JSON object containing:
- "topics": array of topic objects with "topic", "keywords", "probability", and "relevance" fields

Response:`, content)
}

// parseClassificationResponse parses AI classification response
func (a *Analyzer) parseClassificationResponse(response string, classification *common.ContentClassification) error {
	// Try to extract JSON from response
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}") + 1

	if jsonStart == -1 || jsonEnd <= jsonStart {
		// Fallback to simple parsing
		return a.parseClassificationFallback(response, classification)
	}

	jsonStr := response[jsonStart:jsonEnd]

	var result struct {
		Category    string   `json:"category"`
		Subcategory string   `json:"subcategory"`
		Tags        []string `json:"tags"`
		Confidence  float64  `json:"confidence"`
		Reasoning   string   `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return a.parseClassificationFallback(response, classification)
	}

	classification.Category = result.Category
	classification.Subcategory = result.Subcategory
	classification.Tags = result.Tags
	classification.Confidence = result.Confidence
	classification.Metadata["reasoning"] = result.Reasoning

	return nil
}

// parseSentimentResponse parses AI sentiment analysis response
func (a *Analyzer) parseSentimentResponse(response string, classification *common.ContentClassification) error {
	// Try to extract JSON from response
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}") + 1

	if jsonStart == -1 || jsonEnd <= jsonStart {
		return a.parseSentimentFallback(response, classification)
	}

	jsonStr := response[jsonStart:jsonEnd]

	var result struct {
		Label        string  `json:"label"`
		Score        float64 `json:"score"`
		Magnitude    float64 `json:"magnitude"`
		Subjectivity float64 `json:"subjectivity"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return a.parseSentimentFallback(response, classification)
	}

	classification.Sentiment = common.SentimentAnalysis{
		Label:        result.Label,
		Score:        result.Score,
		Magnitude:    result.Magnitude,
		Subjectivity: result.Subjectivity,
	}

	return nil
}

// parseTopicsResponse parses AI topic extraction response
func (a *Analyzer) parseTopicsResponse(response string, classification *common.ContentClassification) error {
	// Try to extract JSON from response
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}") + 1

	if jsonStart == -1 || jsonEnd <= jsonStart {
		return a.parseTopicsFallback(response, classification)
	}

	jsonStr := response[jsonStart:jsonEnd]

	var result struct {
		Topics []struct {
			Topic       string   `json:"topic"`
			Keywords    []string `json:"keywords"`
			Probability float64  `json:"probability"`
			Relevance   float64  `json:"relevance"`
		} `json:"topics"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return a.parseTopicsFallback(response, classification)
	}

	classification.Topics = make([]common.TopicAnalysis, len(result.Topics))
	for i, topic := range result.Topics {
		classification.Topics[i] = common.TopicAnalysis{
			Topic:       topic.Topic,
			Keywords:    topic.Keywords,
			Probability: topic.Probability,
			Relevance:   topic.Relevance,
		}
	}

	return nil
}

// Fallback parsing methods for when JSON parsing fails
func (a *Analyzer) parseClassificationFallback(response string, classification *common.ContentClassification) error {
	// Simple keyword-based classification
	response = strings.ToLower(response)

	// Default values
	classification.Category = "general"
	classification.Confidence = 0.5
	classification.Tags = []string{}

	// Check for common categories
	for _, category := range a.config.Categories {
		if strings.Contains(response, strings.ToLower(category)) {
			classification.Category = category
			classification.Confidence = 0.7
			break
		}
	}

	return nil
}

func (a *Analyzer) parseSentimentFallback(response string, classification *common.ContentClassification) error {
	response = strings.ToLower(response)
 
	sentiment := common.SentimentAnalysis{
		Label:        "neutral",
		Score:        0.5,
		Magnitude:    0.5,
		Subjectivity: 0.5,
	}

	// Simple keyword-based sentiment detection
	if strings.Contains(response, "positive") {
		sentiment.Label = "positive"
		sentiment.Score = 0.7
	} else if strings.Contains(response, "negative") {
		sentiment.Label = "negative"
		sentiment.Score = 0.7
	}

	classification.Sentiment = sentiment
	return nil
}

func (a *Analyzer) parseTopicsFallback(response string, classification *common.ContentClassification) error {
	// Extract simple topics from response
	words := strings.Fields(strings.ToLower(response))
	topicMap := make(map[string]int)

	// Count word frequencies
	for _, word := range words {
		if len(word) > 3 { // Only consider words longer than 3 characters
			topicMap[word]++
		}
	}

	// Convert to topics
	classification.Topics = []common.TopicAnalysis{}
	for word, count := range topicMap {
		if count > 1 { // Only include words that appear more than once
			classification.Topics = append(classification.Topics, common.TopicAnalysis{
				Topic:       word,
				Keywords:    []string{word},
				Probability: float64(count) / float64(len(words)),
				Relevance:   0.5,
			})
		}
	}

	return nil
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("analysis_%d", time.Now().UnixNano())
}
