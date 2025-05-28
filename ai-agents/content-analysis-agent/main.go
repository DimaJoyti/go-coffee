package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
	"gopkg.in/yaml.v2"
)

// Config defines the structure for the agent's configuration
type Config struct {
	AgentName string `yaml:"agent_name"`
	LogLevel  string `yaml:"log_level"`
	
	Reddit struct {
		Enabled      bool     `yaml:"enabled"`
		ClientID     string   `yaml:"client_id"`
		ClientSecret string   `yaml:"client_secret"`
		UserAgent    string   `yaml:"user_agent"`
		Username     string   `yaml:"username"`
		Password     string   `yaml:"password"`
		Subreddits   []string `yaml:"subreddits"`
		PollInterval string   `yaml:"poll_interval"`
	} `yaml:"reddit"`
	
	AI struct {
		Provider    string  `yaml:"provider"`
		APIKey      string  `yaml:"api_key"`
		Model       string  `yaml:"model"`
		Temperature float64 `yaml:"temperature"`
		MaxTokens   int     `yaml:"max_tokens"`
	} `yaml:"ai"`
	
	RAG struct {
		Enabled       bool   `yaml:"enabled"`
		VectorDBURL   string `yaml:"vector_db_url"`
		IndexName     string `yaml:"index_name"`
		EmbeddingModel string `yaml:"embedding_model"`
	} `yaml:"rag"`
	
	Kafka struct {
		BrokerAddress                string `yaml:"broker_address"`
		InputTopicRedditContent      string `yaml:"input_topic_reddit_content"`
		OutputTopicContentAnalysis   string `yaml:"output_topic_content_analysis"`
		OutputTopicClassification    string `yaml:"output_topic_classification"`
		OutputTopicTrendAnalysis     string `yaml:"output_topic_trend_analysis"`
		ConsumerGroup                string `yaml:"consumer_group"`
	} `yaml:"kafka"`
	
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Name     string `yaml:"name"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"database"`
	
	Processing struct {
		BatchSize           int    `yaml:"batch_size"`
		ProcessingInterval  string `yaml:"processing_interval"`
		MaxRetries          int    `yaml:"max_retries"`
		EnableSentiment     bool   `yaml:"enable_sentiment"`
		EnableTopicModeling bool   `yaml:"enable_topic_modeling"`
		EnableTrendAnalysis bool   `yaml:"enable_trend_analysis"`
		MinConfidence       float64 `yaml:"min_confidence"`
	} `yaml:"processing"`
}

// RedditContent represents Reddit content for processing
type RedditContent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // post, comment
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Author      string                 `json:"author"`
	Subreddit   string                 `json:"subreddit"`
	Score       int                    `json:"score"`
	CreatedUTC  time.Time              `json:"created_utc"`
	URL         string                 `json:"url"`
	Metadata    map[string]interface{} `json:"metadata"`
	ProcessedAt time.Time              `json:"processed_at"`
}

// ContentAnalysisResult represents the result of content analysis
type ContentAnalysisResult struct {
	ID             string                 `json:"id"`
	ContentID      string                 `json:"content_id"`
	ContentType    string                 `json:"content_type"`
	Classification Classification         `json:"classification"`
	Sentiment      SentimentAnalysis      `json:"sentiment"`
	Topics         []TopicAnalysis        `json:"topics"`
	TrendIndicators TrendIndicators       `json:"trend_indicators"`
	Confidence     float64                `json:"confidence"`
	ProcessedAt    time.Time              `json:"processed_at"`
	ModelUsed      string                 `json:"model_used"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// Classification represents content classification
type Classification struct {
	Category    string   `json:"category"`
	Subcategory string   `json:"subcategory"`
	Tags        []string `json:"tags"`
	Confidence  float64  `json:"confidence"`
}

// SentimentAnalysis represents sentiment analysis results
type SentimentAnalysis struct {
	Label        string  `json:"label"` // positive, negative, neutral
	Score        float64 `json:"score"`
	Magnitude    float64 `json:"magnitude"`
	Subjectivity float64 `json:"subjectivity"`
}

// TopicAnalysis represents topic modeling results
type TopicAnalysis struct {
	Topic       string   `json:"topic"`
	Keywords    []string `json:"keywords"`
	Probability float64  `json:"probability"`
	Relevance   float64  `json:"relevance"`
}

// TrendIndicators represents trend analysis indicators
type TrendIndicators struct {
	ViralPotential   float64 `json:"viral_potential"`
	EngagementScore  float64 `json:"engagement_score"`
	TrendingKeywords []string `json:"trending_keywords"`
	EmergingTopics   []string `json:"emerging_topics"`
	SentimentTrend   string   `json:"sentiment_trend"`
}

func main() {
	fmt.Println("Starting Content Analysis Agent...")

	// Load configuration
	configPath := "config.yaml"
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	fmt.Printf("Agent Name: %s, Log Level: %s\n", config.AgentName, config.LogLevel)

	// Initialize context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initialize Kafka consumer
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{config.Kafka.BrokerAddress},
		Topic:    config.Kafka.InputTopicRedditContent,
		GroupID:  config.Kafka.ConsumerGroup,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer consumer.Close()

	// Initialize Kafka producer
	producer := &kafka.Writer{
		Addr:     kafka.TCP(config.Kafka.BrokerAddress),
		Balancer: &kafka.LeastBytes{},
	}
	defer producer.Close()

	// Start content processing goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := processRedditContent(ctx, consumer, producer, config); err != nil {
					log.Printf("Error processing Reddit content: %v", err)
					time.Sleep(5 * time.Second) // Wait before retrying
				}
			}
		}
	}()

	// Start periodic trend analysis
	if config.Processing.EnableTrendAnalysis {
		go func() {
			ticker := time.NewTicker(1 * time.Hour) // Run trend analysis every hour
			defer ticker.Stop()
			
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if err := performTrendAnalysis(ctx, producer, config); err != nil {
						log.Printf("Error performing trend analysis: %v", err)
					}
				}
			}
		}()
	}

	fmt.Println("Content Analysis Agent started successfully.")
	fmt.Println("Monitoring Reddit content for analysis...")
	fmt.Println("Press Ctrl+C to stop the agent.")

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down Content Analysis Agent...")
	cancel()
	
	// Give some time for graceful shutdown
	time.Sleep(2 * time.Second)
	fmt.Println("Content Analysis Agent stopped.")
}

// loadConfig loads the agent configuration from a YAML file
func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// processRedditContent processes incoming Reddit content
func processRedditContent(ctx context.Context, consumer *kafka.Reader, producer *kafka.Writer, config *Config) error {
	// Read message from Kafka
	message, err := consumer.FetchMessage(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch message: %w", err)
	}

	// Parse Reddit content
	var content RedditContent
	if err := json.Unmarshal(message.Value, &content); err != nil {
		log.Printf("Failed to parse Reddit content: %v", err)
		consumer.CommitMessages(ctx, message)
		return nil
	}

	log.Printf("Processing %s: %s (from r/%s)", content.Type, content.ID, content.Subreddit)

	// Perform content analysis
	result, err := analyzeContent(ctx, &content, config)
	if err != nil {
		log.Printf("Failed to analyze content %s: %v", content.ID, err)
		consumer.CommitMessages(ctx, message)
		return nil
	}

	// Send analysis result to Kafka
	if err := sendAnalysisResult(ctx, producer, result, config); err != nil {
		log.Printf("Failed to send analysis result: %v", err)
		return err
	}

	// Commit the message
	if err := consumer.CommitMessages(ctx, message); err != nil {
		return fmt.Errorf("failed to commit message: %w", err)
	}

	log.Printf("Successfully analyzed content %s (category: %s, confidence: %.2f)", 
		content.ID, result.Classification.Category, result.Confidence)

	return nil
}

// analyzeContent performs comprehensive content analysis
func analyzeContent(ctx context.Context, content *RedditContent, config *Config) (*ContentAnalysisResult, error) {
	result := &ContentAnalysisResult{
		ID:          generateAnalysisID(),
		ContentID:   content.ID,
		ContentType: content.Type,
		ProcessedAt: time.Now(),
		ModelUsed:   config.AI.Model,
		Metadata:    make(map[string]interface{}),
	}

	// Prepare text for analysis
	text := prepareTextForAnalysis(content)
	if text == "" {
		return nil, fmt.Errorf("no text content to analyze")
	}

	// Perform classification using AI
	classification, err := classifyContent(ctx, text, config)
	if err != nil {
		return nil, fmt.Errorf("failed to classify content: %w", err)
	}
	result.Classification = *classification

	// Perform sentiment analysis if enabled
	if config.Processing.EnableSentiment {
		sentiment, err := analyzeSentiment(ctx, text, config)
		if err != nil {
			log.Printf("Failed to analyze sentiment: %v", err)
		} else {
			result.Sentiment = *sentiment
		}
	}

	// Perform topic modeling if enabled
	if config.Processing.EnableTopicModeling {
		topics, err := extractTopics(ctx, text, config)
		if err != nil {
			log.Printf("Failed to extract topics: %v", err)
		} else {
			result.Topics = topics
		}
	}

	// Calculate trend indicators
	trendIndicators := calculateTrendIndicators(content, result)
	result.TrendIndicators = *trendIndicators

	// Calculate overall confidence
	result.Confidence = calculateOverallConfidence(result)

	// Store metadata
	result.Metadata["subreddit"] = content.Subreddit
	result.Metadata["author"] = content.Author
	result.Metadata["score"] = content.Score
	result.Metadata["created_utc"] = content.CreatedUTC.Format(time.RFC3339)

	return result, nil
}

// Helper functions (simplified implementations)
func prepareTextForAnalysis(content *RedditContent) string {
	if content.Title != "" && content.Content != "" {
		return content.Title + "\n\n" + content.Content
	} else if content.Title != "" {
		return content.Title
	}
	return content.Content
}

func classifyContent(ctx context.Context, text string, config *Config) (*Classification, error) {
	// Simplified classification - in real implementation, this would call AI service
	return &Classification{
		Category:    "general",
		Subcategory: "discussion",
		Tags:        []string{"reddit", "content"},
		Confidence:  0.8,
	}, nil
}

func analyzeSentiment(ctx context.Context, text string, config *Config) (*SentimentAnalysis, error) {
	// Simplified sentiment analysis - in real implementation, this would call AI service
	return &SentimentAnalysis{
		Label:        "neutral",
		Score:        0.6,
		Magnitude:    0.5,
		Subjectivity: 0.4,
	}, nil
}

func extractTopics(ctx context.Context, text string, config *Config) ([]TopicAnalysis, error) {
	// Simplified topic extraction - in real implementation, this would call AI service
	return []TopicAnalysis{
		{
			Topic:       "technology",
			Keywords:    []string{"tech", "software", "development"},
			Probability: 0.7,
			Relevance:   0.8,
		},
	}, nil
}

func calculateTrendIndicators(content *RedditContent, result *ContentAnalysisResult) *TrendIndicators {
	// Simplified trend calculation based on engagement metrics
	engagementScore := float64(content.Score) / 100.0 // Normalize score
	if engagementScore > 1.0 {
		engagementScore = 1.0
	}

	return &TrendIndicators{
		ViralPotential:   engagementScore * 0.8,
		EngagementScore:  engagementScore,
		TrendingKeywords: []string{"trending", "popular"},
		EmergingTopics:   []string{"emerging", "new"},
		SentimentTrend:   result.Sentiment.Label,
	}
}

func calculateOverallConfidence(result *ContentAnalysisResult) float64 {
	// Calculate weighted average of all confidence scores
	totalWeight := 0.0
	totalScore := 0.0

	// Classification confidence (weight: 0.4)
	totalScore += result.Classification.Confidence * 0.4
	totalWeight += 0.4

	// Sentiment confidence (weight: 0.3)
	totalScore += result.Sentiment.Score * 0.3
	totalWeight += 0.3

	// Topic confidence (weight: 0.3)
	if len(result.Topics) > 0 {
		avgTopicConfidence := 0.0
		for _, topic := range result.Topics {
			avgTopicConfidence += topic.Probability
		}
		avgTopicConfidence /= float64(len(result.Topics))
		totalScore += avgTopicConfidence * 0.3
		totalWeight += 0.3
	}

	if totalWeight > 0 {
		return totalScore / totalWeight
	}
	return 0.5 // Default confidence
}

func sendAnalysisResult(ctx context.Context, producer *kafka.Writer, result *ContentAnalysisResult, config *Config) error {
	// Marshal result to JSON
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal analysis result: %w", err)
	}

	// Send to content analysis topic
	message := kafka.Message{
		Topic: config.Kafka.OutputTopicContentAnalysis,
		Key:   []byte(result.ContentID),
		Value: data,
		Time:  time.Now(),
	}

	return producer.WriteMessages(ctx, message)
}

func performTrendAnalysis(ctx context.Context, producer *kafka.Writer, config *Config) error {
	log.Println("Performing periodic trend analysis...")
	
	// This would implement comprehensive trend analysis across all processed content
	// For now, it's a placeholder
	
	trendData := map[string]interface{}{
		"timestamp":        time.Now(),
		"trending_topics":  []string{"ai", "technology", "coffee"},
		"sentiment_trend":  "positive",
		"engagement_trend": "increasing",
	}

	data, err := json.Marshal(trendData)
	if err != nil {
		return fmt.Errorf("failed to marshal trend data: %w", err)
	}

	message := kafka.Message{
		Topic: config.Kafka.OutputTopicTrendAnalysis,
		Key:   []byte("trend_analysis"),
		Value: data,
		Time:  time.Now(),
	}

	return producer.WriteMessages(ctx, message)
}

func generateAnalysisID() string {
	return fmt.Sprintf("analysis_%d", time.Now().UnixNano())
}
