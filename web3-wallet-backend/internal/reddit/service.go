package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/internal/common" // New import for common models
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/internal/content"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/config"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/kafka"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/redis"
)

// Service provides Reddit content collection and analysis
type Service struct {
	config    config.RedditConfig
	logger    *logger.Logger
	client    *Client
	analyzer  *content.Analyzer
	cache     redis.Client
	producer  kafka.Producer
	
	// State management
	mutex     sync.RWMutex
	running   bool
	stopChan  chan struct{}
	
	// Statistics
	stats     ServiceStats
}

// ServiceStats represents service statistics
type ServiceStats struct {
	PostsCollected    int64     `json:"posts_collected"`
	CommentsCollected int64     `json:"comments_collected"`
	PostsAnalyzed     int64     `json:"posts_analyzed"`
	CommentsAnalyzed  int64     `json:"comments_analyzed"`
	ErrorsCount       int64     `json:"errors_count"`
	LastCollectionAt  time.Time `json:"last_collection_at"`
	LastAnalysisAt    time.Time `json:"last_analysis_at"`
	Uptime            time.Duration `json:"uptime"`
	StartedAt         time.Time `json:"started_at"`
}

// NewService creates a new Reddit service
func NewService(
	cfg config.RedditConfig,
	logger *logger.Logger,
	analyzer *content.Analyzer,
	cache redis.Client,
	producer kafka.Producer,
) (*Service, error) {
	// Initialize Reddit client
	client, err := NewClient(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Reddit client: %w", err)
	}

	service := &Service{
		config:   cfg,
		logger:   logger,
		client:   client,
		analyzer: analyzer,
		cache:    cache,
		producer: producer,
		stopChan: make(chan struct{}),
		stats: ServiceStats{
			StartedAt: time.Now(),
		},
	}

	logger.Info("Reddit service initialized successfully")
	return service, nil
}

// Start starts the Reddit content collection and analysis
func (s *Service) Start(ctx context.Context) error {
	s.mutex.Lock()
	if s.running {
		s.mutex.Unlock()
		return fmt.Errorf("service is already running")
	}
	s.running = true
	s.mutex.Unlock()

	s.logger.Info("Starting Reddit content collection service")

	// Start collection workers for each subreddit
	for _, subreddit := range s.config.Subreddits {
		go s.collectSubredditContent(ctx, subreddit)
	}

	// Start periodic trend analysis
	go s.runPeriodicTrendAnalysis(ctx)

	// Start statistics updater
	go s.updateStatistics(ctx)

	s.logger.Info("Reddit service started successfully")
	return nil
}

// Stop stops the Reddit service
func (s *Service) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return fmt.Errorf("service is not running")
	}

	s.logger.Info("Stopping Reddit service")
	close(s.stopChan)
	s.running = false

	// Close Reddit client
	if err := s.client.Close(); err != nil {
		s.logger.Error(fmt.Sprintf("Error closing Reddit client: %v", err))
	}

	s.logger.Info("Reddit service stopped")
	return nil
}

// collectSubredditContent collects content from a specific subreddit
func (s *Service) collectSubredditContent(ctx context.Context, subreddit string) {
	s.logger.Info(fmt.Sprintf("Starting content collection for r/%s", subreddit))

	ticker := time.NewTicker(5 * time.Minute) // Collect every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.collectAndProcessSubreddit(ctx, subreddit); err != nil {
				s.logger.Error(fmt.Sprintf("Error collecting from r/%s: %v", subreddit, err))
				s.incrementErrorCount()
			}
		}
	}
}

// collectAndProcessSubreddit collects and processes content from a subreddit
func (s *Service) collectAndProcessSubreddit(ctx context.Context, subreddit string) error {
	s.logger.Debug(fmt.Sprintf("Collecting content from r/%s", subreddit))

	// Collect hot posts
	posts, _, err := s.client.GetSubredditPosts(ctx, subreddit, "hot", 25, "")
	if err != nil {
		return fmt.Errorf("failed to get posts from r/%s: %w", subreddit, err)
	}

	// Process each post
	for _, post := range posts {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.stopChan:
			return nil
		default:
			if err := s.processPost(ctx, &post); err != nil {
				s.logger.Error(fmt.Sprintf("Error processing post %s: %v", post.ID, err))
				continue
			}
			s.incrementPostsCollected()
		}
	}

	s.stats.LastCollectionAt = time.Now()
	s.logger.Debug(fmt.Sprintf("Collected %d posts from r/%s", len(posts), subreddit))

	return nil
}

// processPost processes a single Reddit post
func (s *Service) processPost(ctx context.Context, post *common.RedditPost) error {
	// Check if already processed
	cacheKey := fmt.Sprintf("reddit:processed:post:%s", post.ID)
	if exists, _ := s.cache.Exists(ctx, cacheKey); exists {
		return nil // Already processed
	}

	// Analyze post content
	classification, err := s.analyzer.AnalyzePost(ctx, post)
	if err != nil {
		return fmt.Errorf("failed to analyze post: %w", err)
	}

	// Update post with analysis results
	post.Classification = classification.Category
	post.Sentiment = classification.Sentiment.Label
	post.Confidence = classification.Confidence
	post.ProcessedAt = time.Now()

	// Extract topics
	if len(classification.Topics) > 0 {
		topics := make([]string, len(classification.Topics))
		for i, topic := range classification.Topics {
			topics[i] = topic.Topic
		}
		post.Topics = topics
	}

	// Send to Kafka for further processing
	if err := s.sendToKafka(ctx, "reddit_posts", post); err != nil {
		return fmt.Errorf("failed to send post to Kafka: %w", err)
	}

	// Mark as processed
	if err := s.cache.Set(ctx, cacheKey, "processed", 24*time.Hour); err != nil {
		s.logger.Warn(fmt.Sprintf("Failed to cache processed status: %v", err))
	}

	// Collect comments for high-engagement posts
	if post.Score > 100 || post.NumComments > 50 {
		go s.collectPostComments(ctx, post)
	}

	s.incrementPostsAnalyzed()
	s.stats.LastAnalysisAt = time.Now()

	return nil
}

// collectPostComments collects and processes comments for a post
func (s *Service) collectPostComments(ctx context.Context, post *common.RedditPost) {
	comments, err := s.client.GetPostComments(ctx, post.Subreddit, post.ID, "top", 50)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get comments for post %s: %v", post.ID, err))
		return
	}

	for _, comment := range comments {
		if err := s.processComment(ctx, &comment, post.ID); err != nil {
			s.logger.Error(fmt.Sprintf("Error processing comment %s: %v", comment.ID, err))
			continue
		}
		s.incrementCommentsCollected()
	}
}

// processComment processes a single Reddit comment
func (s *Service) processComment(ctx context.Context, comment *common.RedditComment, postID string) error {
	// Check if already processed
	cacheKey := fmt.Sprintf("reddit:processed:comment:%s", comment.ID)
	if exists, _ := s.cache.Exists(ctx, cacheKey); exists {
		return nil // Already processed
	}

	// Set post ID for context
	comment.PostID = postID

	// Analyze comment content
	classification, err := s.analyzer.AnalyzeComment(ctx, comment)
	if err != nil {
		return fmt.Errorf("failed to analyze comment: %w", err)
	}

	// Update comment with analysis results
	comment.Classification = classification.Category
	comment.Sentiment = classification.Sentiment.Label
	comment.Confidence = classification.Confidence
	comment.ProcessedAt = time.Now()

	// Extract topics
	if len(classification.Topics) > 0 {
		topics := make([]string, len(classification.Topics))
		for i, topic := range classification.Topics {
			topics[i] = topic.Topic
		}
		comment.Topics = topics
	}

	// Send to Kafka for further processing
	if err := s.sendToKafka(ctx, "reddit_comments", comment); err != nil {
		return fmt.Errorf("failed to send comment to Kafka: %w", err)
	}

	// Mark as processed
	if err := s.cache.Set(ctx, cacheKey, "processed", 24*time.Hour); err != nil {
		s.logger.Warn(fmt.Sprintf("Failed to cache processed status: %v", err))
	}

	s.incrementCommentsAnalyzed()
	return nil
}

// runPeriodicTrendAnalysis runs periodic trend analysis
func (s *Service) runPeriodicTrendAnalysis(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // Run every hour
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.performTrendAnalysis(ctx); err != nil {
				s.logger.Error(fmt.Sprintf("Error performing trend analysis: %v", err))
			}
		}
	}
}

// performTrendAnalysis performs trend analysis across collected content
func (s *Service) performTrendAnalysis(ctx context.Context) error {
	s.logger.Info("Performing trend analysis")

	// This would implement comprehensive trend analysis
	// For now, it's a simplified version
	trendData := common.TrendAnalysis{
		ID:          generateTrendID(),
		Timeframe:   "hourly",
		GeneratedAt: time.Now(),
		Metrics: common.TrendMetrics{
			PostCount:      int(s.stats.PostsCollected),
			CommentCount:   int(s.stats.CommentsCollected),
			EngagementRate: 0.75, // Placeholder
		},
	}

	// Send trend analysis to Kafka
	return s.sendToKafka(ctx, "trend_analysis", &trendData)
}

// sendToKafka sends data to Kafka topic
func (s *Service) sendToKafka(ctx context.Context, topic string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	return s.producer.Produce(topic, nil, jsonData)
}

// updateStatistics updates service statistics
func (s *Service) updateStatistics(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.mutex.Lock()
			s.stats.Uptime = time.Since(s.stats.StartedAt)
			s.mutex.Unlock()
		}
	}
}

// GetStats returns service statistics
func (s *Service) GetStats() ServiceStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.stats
}

// IsRunning returns whether the service is running
func (s *Service) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.running
}

// Helper methods for statistics
func (s *Service) incrementPostsCollected() {
	s.mutex.Lock()
	s.stats.PostsCollected++
	s.mutex.Unlock()
}

func (s *Service) incrementCommentsCollected() {
	s.mutex.Lock()
	s.stats.CommentsCollected++
	s.mutex.Unlock()
}

func (s *Service) incrementPostsAnalyzed() {
	s.mutex.Lock()
	s.stats.PostsAnalyzed++
	s.mutex.Unlock()
}

func (s *Service) incrementCommentsAnalyzed() {
	s.mutex.Lock()
	s.stats.CommentsAnalyzed++
	s.mutex.Unlock()
}

func (s *Service) incrementErrorCount() {
	s.mutex.Lock()
	s.stats.ErrorsCount++
	s.mutex.Unlock()
}

// generateTrendID generates a unique trend analysis ID
func generateTrendID() string {
	return fmt.Sprintf("trend_%d", time.Now().UnixNano())
}
