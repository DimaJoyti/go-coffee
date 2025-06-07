package social

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/config"
	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/core"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Service handles all social media data collection and analysis
type Service struct {
	client *core.MCPClient
	config *config.BrightDataHubConfig
	logger *logger.Logger
	
	// Platform handlers
	instagram *InstagramHandler
	facebook  *FacebookHandler
	twitter   *TwitterHandler
	linkedin  *LinkedInHandler
}

// SocialData represents aggregated social media data
type SocialData struct {
	Platform    string                 `json:"platform"`
	Type        string                 `json:"type"` // profile, post, comment, etc.
	Content     map[string]interface{} `json:"content"`
	Metadata    *SocialMetadata        `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// SocialMetadata contains metadata about social data
type SocialMetadata struct {
	URL           string    `json:"url"`
	Author        string    `json:"author,omitempty"`
	Engagement    *Engagement `json:"engagement,omitempty"`
	Sentiment     *Sentiment  `json:"sentiment,omitempty"`
	DataQuality   float64   `json:"data_quality"`
	ProcessedAt   time.Time `json:"processed_at"`
}

// Engagement represents social media engagement metrics
type Engagement struct {
	Likes     int64 `json:"likes"`
	Comments  int64 `json:"comments"`
	Shares    int64 `json:"shares"`
	Views     int64 `json:"views,omitempty"`
	Followers int64 `json:"followers,omitempty"`
}

// Sentiment represents sentiment analysis results
type Sentiment struct {
	Score      float64 `json:"score"`      // -1 to 1
	Label      string  `json:"label"`      // positive, negative, neutral
	Confidence float64 `json:"confidence"` // 0 to 1
}

// SocialAnalytics represents aggregated social analytics
type SocialAnalytics struct {
	Platform        string                 `json:"platform"`
	TotalPosts      int64                  `json:"total_posts"`
	TotalEngagement int64                  `json:"total_engagement"`
	AvgSentiment    float64                `json:"avg_sentiment"`
	TopHashtags     []string               `json:"top_hashtags"`
	TopMentions     []string               `json:"top_mentions"`
	TrendingTopics  []string               `json:"trending_topics"`
	TimeRange       *TimeRange             `json:"time_range"`
	Demographics    map[string]interface{} `json:"demographics,omitempty"`
}

// TimeRange represents a time period for analytics
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// NewService creates a new social media service
func NewService(client *core.MCPClient, cfg *config.BrightDataHubConfig, log *logger.Logger) (*Service, error) {
	service := &Service{
		client: client,
		config: cfg,
		logger: log,
	}
	
	// Initialize platform handlers
	if cfg.Social.Instagram.Enabled {
		service.instagram = NewInstagramHandler(client, cfg, log)
	}
	
	if cfg.Social.Facebook.Enabled {
		service.facebook = NewFacebookHandler(client, cfg, log)
	}
	
	if cfg.Social.Twitter.Enabled {
		service.twitter = NewTwitterHandler(client, cfg, log)
	}
	
	if cfg.Social.LinkedIn.Enabled {
		service.linkedin = NewLinkedInHandler(client, cfg, log)
	}
	
	return service, nil
}

// Start starts the social media service
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting social media service")
	
	// Start background data collection if enabled
	// This could include periodic trending analysis, sentiment monitoring, etc.
	
	return nil
}

// ExecuteFunction executes a social media function
func (s *Service) ExecuteFunction(ctx context.Context, function string, params interface{}) (interface{}, error) {
	s.logger.Debug("Executing social function: %s", function)
	
	switch function {
	// Instagram functions
	case "web_data_instagram_profiles_Bright_Data":
		if s.instagram == nil {
			return nil, fmt.Errorf("Instagram handler not enabled")
		}
		return s.instagram.GetProfile(ctx, params)
		
	case "web_data_instagram_posts_Bright_Data":
		if s.instagram == nil {
			return nil, fmt.Errorf("Instagram handler not enabled")
		}
		return s.instagram.GetPosts(ctx, params)
		
	case "web_data_instagram_reels_Bright_Data":
		if s.instagram == nil {
			return nil, fmt.Errorf("Instagram handler not enabled")
		}
		return s.instagram.GetReels(ctx, params)
		
	case "web_data_instagram_comments_Bright_Data":
		if s.instagram == nil {
			return nil, fmt.Errorf("Instagram handler not enabled")
		}
		return s.instagram.GetComments(ctx, params)
		
	// Facebook functions
	case "web_data_facebook_posts_Bright_Data":
		if s.facebook == nil {
			return nil, fmt.Errorf("Facebook handler not enabled")
		}
		return s.facebook.GetPosts(ctx, params)
		
	case "web_data_facebook_marketplace_listings_Bright_Data":
		if s.facebook == nil {
			return nil, fmt.Errorf("Facebook handler not enabled")
		}
		return s.facebook.GetMarketplaceListings(ctx, params)
		
	case "web_data_facebook_company_reviews_Bright_Data":
		if s.facebook == nil {
			return nil, fmt.Errorf("Facebook handler not enabled")
		}
		return s.facebook.GetCompanyReviews(ctx, params)
		
	// Twitter/X functions
	case "web_data_x_posts_Bright_Data":
		if s.twitter == nil {
			return nil, fmt.Errorf("Twitter handler not enabled")
		}
		return s.twitter.GetPosts(ctx, params)
		
	// LinkedIn functions
	case "web_data_linkedin_person_profile_Bright_Data":
		if s.linkedin == nil {
			return nil, fmt.Errorf("LinkedIn handler not enabled")
		}
		return s.linkedin.GetPersonProfile(ctx, params)
		
	case "web_data_linkedin_company_profile_Bright_Data":
		if s.linkedin == nil {
			return nil, fmt.Errorf("LinkedIn handler not enabled")
		}
		return s.linkedin.GetCompanyProfile(ctx, params)
		
	default:
		return nil, fmt.Errorf("unsupported social function: %s", function)
	}
}

// GetAggregatedAnalytics returns aggregated analytics across all platforms
func (s *Service) GetAggregatedAnalytics(ctx context.Context, timeRange *TimeRange) (*SocialAnalytics, error) {
	analytics := &SocialAnalytics{
		Platform:  "aggregated",
		TimeRange: timeRange,
	}
	
	// Collect analytics from all enabled platforms
	var totalPosts int64
	var totalEngagement int64
	var sentimentSum float64
	var sentimentCount int64
	
	allHashtags := make(map[string]int)
	allMentions := make(map[string]int)
	
	// Instagram analytics
	if s.instagram != nil {
		instagramAnalytics, err := s.instagram.GetAnalytics(ctx, timeRange)
		if err != nil {
			s.logger.Warn("Failed to get Instagram analytics: %v", err)
		} else {
			totalPosts += instagramAnalytics.TotalPosts
			totalEngagement += instagramAnalytics.TotalEngagement
			if instagramAnalytics.AvgSentiment != 0 {
				sentimentSum += instagramAnalytics.AvgSentiment
				sentimentCount++
			}
			
			// Aggregate hashtags and mentions
			for _, hashtag := range instagramAnalytics.TopHashtags {
				allHashtags[hashtag]++
			}
			for _, mention := range instagramAnalytics.TopMentions {
				allMentions[mention]++
			}
		}
	}
	
	// Facebook analytics
	if s.facebook != nil {
		facebookAnalytics, err := s.facebook.GetAnalytics(ctx, timeRange)
		if err != nil {
			s.logger.Warn("Failed to get Facebook analytics: %v", err)
		} else {
			totalPosts += facebookAnalytics.TotalPosts
			totalEngagement += facebookAnalytics.TotalEngagement
			if facebookAnalytics.AvgSentiment != 0 {
				sentimentSum += facebookAnalytics.AvgSentiment
				sentimentCount++
			}
		}
	}
	
	// Twitter analytics
	if s.twitter != nil {
		twitterAnalytics, err := s.twitter.GetAnalytics(ctx, timeRange)
		if err != nil {
			s.logger.Warn("Failed to get Twitter analytics: %v", err)
		} else {
			totalPosts += twitterAnalytics.TotalPosts
			totalEngagement += twitterAnalytics.TotalEngagement
			if twitterAnalytics.AvgSentiment != 0 {
				sentimentSum += twitterAnalytics.AvgSentiment
				sentimentCount++
			}
			
			// Aggregate hashtags and mentions
			for _, hashtag := range twitterAnalytics.TopHashtags {
				allHashtags[hashtag]++
			}
			for _, mention := range twitterAnalytics.TopMentions {
				allMentions[mention]++
			}
		}
	}
	
	// Set aggregated values
	analytics.TotalPosts = totalPosts
	analytics.TotalEngagement = totalEngagement
	
	if sentimentCount > 0 {
		analytics.AvgSentiment = sentimentSum / float64(sentimentCount)
	}
	
	// Get top hashtags and mentions
	analytics.TopHashtags = getTopItems(allHashtags, 10)
	analytics.TopMentions = getTopItems(allMentions, 10)
	
	return analytics, nil
}

// GetTrendingTopics returns trending topics across all platforms
func (s *Service) GetTrendingTopics(ctx context.Context) ([]string, error) {
	var allTopics []string
	
	// Collect trending topics from all platforms
	if s.instagram != nil {
		topics, err := s.instagram.GetTrendingTopics(ctx)
		if err != nil {
			s.logger.Warn("Failed to get Instagram trending topics: %v", err)
		} else {
			allTopics = append(allTopics, topics...)
		}
	}
	
	if s.twitter != nil {
		topics, err := s.twitter.GetTrendingTopics(ctx)
		if err != nil {
			s.logger.Warn("Failed to get Twitter trending topics: %v", err)
		} else {
			allTopics = append(allTopics, topics...)
		}
	}
	
	// Remove duplicates and return top topics
	uniqueTopics := removeDuplicates(allTopics)
	if len(uniqueTopics) > 20 {
		uniqueTopics = uniqueTopics[:20]
	}
	
	return uniqueTopics, nil
}

// Helper functions
func getTopItems(items map[string]int, limit int) []string {
	type item struct {
		name  string
		count int
	}
	
	var sorted []item
	for name, count := range items {
		sorted = append(sorted, item{name, count})
	}
	
	// Simple bubble sort (for small datasets)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].count < sorted[j+1].count {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	
	var result []string
	for i, item := range sorted {
		if i >= limit {
			break
		}
		result = append(result, item.name)
	}
	
	return result
}

func removeDuplicates(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, item := range items {
		if !seen[strings.ToLower(item)] {
			seen[strings.ToLower(item)] = true
			result = append(result, item)
		}
	}
	
	return result
}
