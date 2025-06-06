package brightdata

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// SentimentAnalyzer analyzes social media sentiment using Bright Data MCP
type SentimentAnalyzer struct {
	service *Service
	logger  *logrus.Logger
	
	// Sentiment analysis patterns
	symbolRegex     *regexp.Regexp
	hashtagRegex    *regexp.Regexp
	mentionRegex    *regexp.Regexp
	
	// Crypto symbols to track
	trackedSymbols  []string
	
	// Influencer accounts to monitor
	influencers     map[string]*InfluencerProfile
}

// NewSentimentAnalyzer creates a new sentiment analyzer
func NewSentimentAnalyzer(service *Service, logger *logrus.Logger) *SentimentAnalyzer {
	trackedSymbols := []string{
		"BTC", "ETH", "BNB", "ADA", "SOL", "XRP", "DOT", "DOGE",
		"AVAX", "MATIC", "LINK", "UNI", "LTC", "ATOM", "XLM",
	}
	
	// Initialize known crypto influencers
	influencers := map[string]*InfluencerProfile{
		"elonmusk": {
			ID:           "elonmusk",
			Username:     "elonmusk",
			Platform:     "twitter",
			DisplayName:  "Elon Musk",
			Followers:    150000000,
			Verified:     true,
			InfluenceScore: 95.0,
			Specialties:  []string{"DOGE", "BTC"},
		},
		"VitalikButerin": {
			ID:           "VitalikButerin",
			Username:     "VitalikButerin",
			Platform:     "twitter",
			DisplayName:  "Vitalik Buterin",
			Followers:    5000000,
			Verified:     true,
			InfluenceScore: 98.0,
			Specialties:  []string{"ETH", "DeFi"},
		},
		"cz_binance": {
			ID:           "cz_binance",
			Username:     "cz_binance",
			Platform:     "twitter",
			DisplayName:  "CZ Binance",
			Followers:    8000000,
			Verified:     true,
			InfluenceScore: 90.0,
			Specialties:  []string{"BNB", "Trading"},
		},
	}
	
	return &SentimentAnalyzer{
		service:        service,
		logger:         logger,
		symbolRegex:    regexp.MustCompile(`\$([A-Z]{3,5})\b`),
		hashtagRegex:   regexp.MustCompile(`#(\w+)`),
		mentionRegex:   regexp.MustCompile(`@(\w+)`),
		trackedSymbols: trackedSymbols,
		influencers:    influencers,
	}
}

// AnalyzeSentiment analyzes sentiment across social media platforms
func (sa *SentimentAnalyzer) AnalyzeSentiment(ctx context.Context) error {
	sa.logger.Info("Starting sentiment analysis")
	
	for _, symbol := range sa.trackedSymbols {
		sentiment, err := sa.analyzeSentimentForSymbol(ctx, symbol)
		if err != nil {
			sa.logger.Errorf("Failed to analyze sentiment for %s: %v", symbol, err)
			continue
		}
		
		sa.service.updateSentiment(symbol, sentiment)
	}
	
	sa.logger.Infof("Completed sentiment analysis for %d symbols", len(sa.trackedSymbols))
	return nil
}

// analyzeSentimentForSymbol analyzes sentiment for a specific symbol
func (sa *SentimentAnalyzer) analyzeSentimentForSymbol(ctx context.Context, symbol string) (*SentimentAnalysis, error) {
	// Collect posts from different platforms
	twitterPosts, err := sa.collectTwitterPosts(ctx, symbol)
	if err != nil {
		sa.logger.Warnf("Failed to collect Twitter posts for %s: %v", symbol, err)
	}
	
	redditPosts, err := sa.collectRedditPosts(ctx, symbol)
	if err != nil {
		sa.logger.Warnf("Failed to collect Reddit posts for %s: %v", symbol, err)
	}
	
	// Combine all posts
	allPosts := append(twitterPosts, redditPosts...)
	
	if len(allPosts) == 0 {
		return nil, fmt.Errorf("no social posts found for symbol %s", symbol)
	}
	
	// Analyze sentiment
	sentiment := sa.calculateAggregatedSentiment(allPosts)
	sentiment.Symbol = symbol
	sentiment.TimeRange = "24h"
	sentiment.LastUpdated = time.Now()
	
	// Calculate platform breakdown
	sentiment.PlatformBreakdown = sa.calculatePlatformBreakdown(allPosts)
	
	// Extract trending topics
	sentiment.TrendingTopics = sa.extractTrendingTopics(allPosts)
	
	// Get influencer posts
	sentiment.InfluencerPosts = sa.filterInfluencerPosts(allPosts)
	
	return sentiment, nil
}

// collectTwitterPosts collects Twitter/X posts about a symbol
func (sa *SentimentAnalyzer) collectTwitterPosts(ctx context.Context, symbol string) ([]*SocialPost, error) {
	// This would use web_data_x_posts_Bright_Data MCP function
	// For now, we'll simulate the data
	
	sa.logger.Infof("Collecting Twitter posts for %s", symbol)
	
	// Simulate Twitter posts
	posts := []*SocialPost{
		{
			ID:          "twitter_1",
			Platform:    "twitter",
			Content:     fmt.Sprintf("$%s is looking bullish! Great fundamentals and strong community support. #crypto #%s", symbol, strings.ToLower(symbol)),
			Author:      "crypto_trader_123",
			AuthorID:    "123456789",
			PostedAt:    time.Now().Add(-1 * time.Hour),
			Sentiment:   0.7,
			Engagement:  150,
			Reach:       5000,
			Symbols:     []string{symbol},
			Hashtags:    []string{"crypto", strings.ToLower(symbol)},
			IsInfluencer: false,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "twitter_2",
			Platform:    "twitter",
			Content:     fmt.Sprintf("Concerned about %s price action. Market seems uncertain. #crypto", symbol),
			Author:      "market_analyst",
			AuthorID:    "987654321",
			PostedAt:    time.Now().Add(-2 * time.Hour),
			Sentiment:   -0.3,
			Engagement:  75,
			Reach:       2000,
			Symbols:     []string{symbol},
			Hashtags:    []string{"crypto"},
			IsInfluencer: false,
			CreatedAt:   time.Now(),
		},
	}
	
	return posts, nil
}

// collectRedditPosts collects Reddit posts about a symbol
func (sa *SentimentAnalyzer) collectRedditPosts(ctx context.Context, symbol string) ([]*SocialPost, error) {
	// This would use web_data_reddit_posts_Bright_Data MCP function
	// For now, we'll simulate the data
	
	sa.logger.Infof("Collecting Reddit posts for %s", symbol)
	
	// Simulate Reddit posts
	posts := []*SocialPost{
		{
			ID:          "reddit_1",
			Platform:    "reddit",
			Content:     fmt.Sprintf("Just bought more %s. This dip is a great opportunity for long-term holders.", symbol),
			Author:      "hodler_2021",
			AuthorID:    "reddit_user_1",
			PostedAt:    time.Now().Add(-3 * time.Hour),
			Sentiment:   0.5,
			Engagement:  200,
			Reach:       10000,
			Symbols:     []string{symbol},
			Hashtags:    []string{},
			IsInfluencer: false,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "reddit_2",
			Platform:    "reddit",
			Content:     fmt.Sprintf("Technical analysis suggests %s might see a correction soon. Be careful out there.", symbol),
			Author:      "ta_expert",
			AuthorID:    "reddit_user_2",
			PostedAt:    time.Now().Add(-4 * time.Hour),
			Sentiment:   -0.2,
			Engagement:  120,
			Reach:       8000,
			Symbols:     []string{symbol},
			Hashtags:    []string{},
			IsInfluencer: false,
			CreatedAt:   time.Now(),
		},
	}
	
	return posts, nil
}

// calculateAggregatedSentiment calculates overall sentiment from posts
func (sa *SentimentAnalyzer) calculateAggregatedSentiment(posts []*SocialPost) *SentimentAnalysis {
	if len(posts) == 0 {
		return &SentimentAnalysis{
			OverallSentiment: 0.0,
			SentimentScore:   50,
			Confidence:       0.0,
			TotalMentions:    0,
		}
	}
	
	var totalSentiment float64
	var totalWeight float64
	var positiveMentions, negativeMentions, neutralMentions int64
	
	for _, post := range posts {
		// Weight by engagement and reach
		weight := float64(post.Engagement + post.Reach/10)
		if weight == 0 {
			weight = 1
		}
		
		totalSentiment += post.Sentiment * weight
		totalWeight += weight
		
		// Count mentions by sentiment
		if post.Sentiment > 0.1 {
			positiveMentions++
		} else if post.Sentiment < -0.1 {
			negativeMentions++
		} else {
			neutralMentions++
		}
	}
	
	overallSentiment := totalSentiment / totalWeight
	sentimentScore := int((overallSentiment + 1) * 50) // Convert -1,1 to 0,100
	
	// Calculate confidence based on number of posts and consistency
	confidence := float64(len(posts)) / 100.0 // More posts = higher confidence
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return &SentimentAnalysis{
		OverallSentiment: overallSentiment,
		SentimentScore:   sentimentScore,
		Confidence:       confidence,
		TotalMentions:    int64(len(posts)),
		PositiveMentions: positiveMentions,
		NegativeMentions: negativeMentions,
		NeutralMentions:  neutralMentions,
	}
}

// calculatePlatformBreakdown calculates sentiment breakdown by platform
func (sa *SentimentAnalyzer) calculatePlatformBreakdown(posts []*SocialPost) map[string]PlatformSentiment {
	breakdown := make(map[string]PlatformSentiment)
	platformPosts := make(map[string][]*SocialPost)
	
	// Group posts by platform
	for _, post := range posts {
		platformPosts[post.Platform] = append(platformPosts[post.Platform], post)
	}
	
	// Calculate sentiment for each platform
	for platform, posts := range platformPosts {
		sentiment := sa.calculateAggregatedSentiment(posts)
		
		var positiveMentions, negativeMentions, neutralMentions int64
		var topPosts []*SocialPost
		
		for _, post := range posts {
			if post.Sentiment > 0.1 {
				positiveMentions++
			} else if post.Sentiment < -0.1 {
				negativeMentions++
			} else {
				neutralMentions++
			}
			
			// Select top posts by engagement
			if len(topPosts) < 3 {
				topPosts = append(topPosts, post)
			}
		}
		
		breakdown[platform] = PlatformSentiment{
			Platform:         platform,
			Sentiment:        sentiment.OverallSentiment,
			Mentions:         int64(len(posts)),
			PositiveMentions: positiveMentions,
			NegativeMentions: negativeMentions,
			NeutralMentions:  neutralMentions,
			TopPosts:         topPosts,
		}
	}
	
	return breakdown
}

// extractTrendingTopics extracts trending topics from posts
func (sa *SentimentAnalyzer) extractTrendingTopics(posts []*SocialPost) []string {
	hashtagCount := make(map[string]int)
	
	for _, post := range posts {
		for _, hashtag := range post.Hashtags {
			hashtagCount[hashtag]++
		}
		
		// Also extract hashtags from content
		matches := sa.hashtagRegex.FindAllStringSubmatch(post.Content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				hashtag := strings.ToLower(match[1])
				hashtagCount[hashtag]++
			}
		}
	}
	
	// Sort by frequency and return top topics
	var topics []string
	for hashtag, count := range hashtagCount {
		if count >= 2 { // Minimum threshold
			topics = append(topics, hashtag)
		}
	}
	
	// Limit to top 10
	if len(topics) > 10 {
		topics = topics[:10]
	}
	
	return topics
}

// filterInfluencerPosts filters posts from known influencers
func (sa *SentimentAnalyzer) filterInfluencerPosts(posts []*SocialPost) []SocialPost {
	var influencerPosts []SocialPost
	
	for _, post := range posts {
		if _, isInfluencer := sa.influencers[post.Author]; isInfluencer {
			post.IsInfluencer = true
			influencerPosts = append(influencerPosts, *post)
		}
	}
	
	return influencerPosts
}
