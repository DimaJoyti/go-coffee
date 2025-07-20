package brightdata

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// SocialScraperConfig represents configuration for social media scraping
type SocialScraperConfig struct {
	UpdateInterval    time.Duration `json:"update_interval"`
	MaxConcurrent     int           `json:"max_concurrent"`
	RateLimitRPS      int           `json:"rate_limit_rps"`
	EnableTwitter     bool          `json:"enable_twitter"`
	EnableReddit      bool          `json:"enable_reddit"`
	EnableTelegram    bool          `json:"enable_telegram"`
	TwitterKeywords   []string      `json:"twitter_keywords"`
	RedditSubreddits  []string      `json:"reddit_subreddits"`
	TelegramChannels  []string      `json:"telegram_channels"`
	InfluencerAccounts []string     `json:"influencer_accounts"`
	MinFollowers      int64         `json:"min_followers"`
}

// SocialScraper handles scraping of social media data using Bright Data MCP
type SocialScraper struct {
	config *SocialScraperConfig
	logger *logrus.Logger
	
	// Bright Data MCP functions
	scrapeTwitterPosts func(ctx context.Context, url string) ([]SocialPost, error)
	scrapeRedditPosts  func(ctx context.Context, url string) ([]SocialPost, error)
	searchSocial       func(ctx context.Context, platform, query string) ([]SocialPost, error)
}

// NewSocialScraper creates a new social media scraper
func NewSocialScraper(config *SocialScraperConfig, logger *logrus.Logger) *SocialScraper {
	return &SocialScraper{
		config: config,
		logger: logger,
		// In real implementation, these would be initialized with actual MCP function calls
		scrapeTwitterPosts: mockScrapeTwitterPosts,
		scrapeRedditPosts:  mockScrapeRedditPosts,
		searchSocial:       mockSearchSocial,
	}
}

// ScrapeCryptoSentiment scrapes crypto sentiment from all enabled social platforms
func (ss *SocialScraper) ScrapeCryptoSentiment(ctx context.Context, symbols []string) (map[string]SentimentAnalysis, error) {
	ss.logger.Info("Scraping crypto sentiment from social media")
	
	results := make(map[string]SentimentAnalysis)
	
	for _, symbol := range symbols {
		sentiment, err := ss.scrapeSentimentForSymbol(ctx, symbol)
		if err != nil {
			ss.logger.Warnf("Failed to scrape sentiment for %s: %v", symbol, err)
			continue
		}
		
		results[symbol] = *sentiment
	}
	
	ss.logger.Infof("Successfully scraped sentiment for %d symbols", len(results))
	return results, nil
}

// ScrapeInfluencerPosts scrapes posts from crypto influencers
func (ss *SocialScraper) ScrapeInfluencerPosts(ctx context.Context) ([]SocialPost, error) {
	ss.logger.Info("Scraping posts from crypto influencers")
	
	var allPosts []SocialPost
	
	for _, account := range ss.config.InfluencerAccounts {
		posts, err := ss.scrapeInfluencerAccount(ctx, account)
		if err != nil {
			ss.logger.Warnf("Failed to scrape influencer %s: %v", account, err)
			continue
		}
		
		allPosts = append(allPosts, posts...)
	}
	
	ss.logger.Infof("Successfully scraped %d influencer posts", len(allPosts))
	return allPosts, nil
}

// ScrapeTrendingTopics scrapes trending crypto topics from social media
func (ss *SocialScraper) ScrapeTrendingTopics(ctx context.Context) ([]TrendingTopic, error) {
	ss.logger.Info("Scraping trending crypto topics")
	
	var topics []TrendingTopic
	
	// Scrape from Twitter
	if ss.config.EnableTwitter {
		twitterTopics, err := ss.scrapeTrendingFromTwitter(ctx)
		if err != nil {
			ss.logger.Warnf("Failed to scrape trending from Twitter: %v", err)
		} else {
			topics = append(topics, twitterTopics...)
		}
	}
	
	// Scrape from Reddit
	if ss.config.EnableReddit {
		redditTopics, err := ss.scrapeTrendingFromReddit(ctx)
		if err != nil {
			ss.logger.Warnf("Failed to scrape trending from Reddit: %v", err)
		} else {
			topics = append(topics, redditTopics...)
		}
	}
	
	ss.logger.Infof("Successfully scraped %d trending topics", len(topics))
	return topics, nil
}

// scrapeSentimentForSymbol scrapes sentiment for a specific crypto symbol
func (ss *SocialScraper) scrapeSentimentForSymbol(ctx context.Context, symbol string) (*SentimentAnalysis, error) {
	sentiment := &SentimentAnalysis{
		Symbol:            symbol,
		TimeRange:         "24h",
		LastUpdated:       time.Now(),
		PlatformBreakdown: make(map[string]PlatformSentiment),
	}
	
	var allPosts []SocialPost
	
	// Search Twitter
	if ss.config.EnableTwitter {
		query := fmt.Sprintf("%s OR $%s", symbol, symbol)
		posts, err := ss.searchSocial(ctx, "twitter", query)
		if err != nil {
			ss.logger.Warnf("Failed to search Twitter for %s: %v", symbol, err)
		} else {
			allPosts = append(allPosts, posts...)
			sentiment.PlatformBreakdown["twitter"] = ss.calculatePlatformSentiment(posts, "twitter")
		}
	}
	
	// Search Reddit
	if ss.config.EnableReddit {
		posts, err := ss.searchSocial(ctx, "reddit", symbol)
		if err != nil {
			ss.logger.Warnf("Failed to search Reddit for %s: %v", symbol, err)
		} else {
			allPosts = append(allPosts, posts...)
			sentiment.PlatformBreakdown["reddit"] = ss.calculatePlatformSentiment(posts, "reddit")
		}
	}
	
	// Calculate overall sentiment
	sentiment.OverallSentiment, sentiment.SentimentScore = ss.calculateOverallSentiment(allPosts)
	sentiment.TotalMentions = int64(len(allPosts))
	sentiment.Confidence = ss.calculateConfidence(allPosts)
	
	// Extract trending topics and influencer posts
	sentiment.TrendingTopics = ss.extractTrendingTopics(allPosts)
	sentiment.InfluencerPosts = ss.filterInfluencerPosts(allPosts)
	
	// Count sentiment breakdown
	sentiment.PositiveMentions, sentiment.NegativeMentions, sentiment.NeutralMentions = ss.countSentimentBreakdown(allPosts)
	
	return sentiment, nil
}

// scrapeInfluencerAccount scrapes posts from a specific influencer account
func (ss *SocialScraper) scrapeInfluencerAccount(ctx context.Context, account string) ([]SocialPost, error) {
	// Determine platform from account format
	platform := "twitter"
	if strings.Contains(account, "reddit.com") {
		platform = "reddit"
	} else if strings.Contains(account, "t.me") {
		platform = "telegram"
	}
	
	switch platform {
	case "twitter":
		return ss.scrapeTwitterPosts(ctx, fmt.Sprintf("https://twitter.com/%s", account))
	case "reddit":
		return ss.scrapeRedditPosts(ctx, account)
	default:
		return []SocialPost{}, nil
	}
}

// scrapeTrendingFromTwitter scrapes trending crypto topics from Twitter
func (ss *SocialScraper) scrapeTrendingFromTwitter(ctx context.Context) ([]TrendingTopic, error) {
	var topics []TrendingTopic
	
	// Search for trending crypto hashtags
	cryptoHashtags := []string{"#Bitcoin", "#Ethereum", "#Crypto", "#DeFi", "#NFT", "#Web3"}
	
	for _, hashtag := range cryptoHashtags {
		posts, err := ss.searchSocial(ctx, "twitter", hashtag)
		if err != nil {
			continue
		}
		
		if len(posts) > 10 { // Threshold for trending
			topic := TrendingTopic{
				Topic:       hashtag,
				Mentions:    int64(len(posts)),
				Sentiment:   ss.calculateAverageSentiment(posts),
				Growth:      ss.calculateGrowthRate(hashtag, posts),
				Platforms:   []string{"twitter"},
				LastUpdated: time.Now(),
			}
			
			// Extract related symbols
			topic.Symbols = ss.extractSymbolsFromPosts(posts)
			
			topics = append(topics, topic)
		}
	}
	
	return topics, nil
}

// scrapeTrendingFromReddit scrapes trending crypto topics from Reddit
func (ss *SocialScraper) scrapeTrendingFromReddit(ctx context.Context) ([]TrendingTopic, error) {
	var topics []TrendingTopic
	
	for _, subreddit := range ss.config.RedditSubreddits {
		url := fmt.Sprintf("https://reddit.com/r/%s/hot", subreddit)
		posts, err := ss.scrapeRedditPosts(ctx, url)
		if err != nil {
			continue
		}
		
		// Group posts by topic/keyword
		topicMap := ss.groupPostsByTopic(posts)
		
		for topic, topicPosts := range topicMap {
			if len(topicPosts) > 5 { // Threshold for trending
				trendingTopic := TrendingTopic{
					Topic:       topic,
					Mentions:    int64(len(topicPosts)),
					Sentiment:   ss.calculateAverageSentiment(topicPosts),
					Growth:      ss.calculateGrowthRate(topic, topicPosts),
					Platforms:   []string{"reddit"},
					LastUpdated: time.Now(),
				}
				
				trendingTopic.Symbols = ss.extractSymbolsFromPosts(topicPosts)
				topics = append(topics, trendingTopic)
			}
		}
	}
	
	return topics, nil
}

// calculatePlatformSentiment calculates sentiment for a specific platform
func (ss *SocialScraper) calculatePlatformSentiment(posts []SocialPost, platform string) PlatformSentiment {
	if len(posts) == 0 {
		return PlatformSentiment{Platform: platform}
	}
	
	var totalSentiment float64
	var positive, negative, neutral int64
	var topPosts []SocialPost
	
	for _, post := range posts {
		totalSentiment += post.Sentiment
		
		if post.Sentiment > 0.1 {
			positive++
		} else if post.Sentiment < -0.1 {
			negative++
		} else {
			neutral++
		}
		
		// Collect top posts by engagement
		if len(topPosts) < 5 || post.Engagement > topPosts[len(topPosts)-1].Engagement {
			topPosts = append(topPosts, post)
			if len(topPosts) > 5 {
				// Sort and keep top 5
				topPosts = topPosts[:5]
			}
		}
	}
	
	return PlatformSentiment{
		Platform:         platform,
		Sentiment:        totalSentiment / float64(len(posts)),
		Mentions:         int64(len(posts)),
		PositiveMentions: positive,
		NegativeMentions: negative,
		NeutralMentions:  neutral,
		TopPosts:         topPosts,
	}
}

// calculateOverallSentiment calculates overall sentiment from all posts
func (ss *SocialScraper) calculateOverallSentiment(posts []SocialPost) (float64, int) {
	if len(posts) == 0 {
		return 0, 50
	}
	
	var totalSentiment float64
	var totalWeight float64
	
	for _, post := range posts {
		// Weight by engagement and reach
		weight := float64(post.Engagement + post.Reach/10)
		if weight < 1 {
			weight = 1
		}
		
		totalSentiment += post.Sentiment * weight
		totalWeight += weight
	}
	
	avgSentiment := totalSentiment / totalWeight
	
	// Convert to 0-100 scale
	score := int((avgSentiment + 1) * 50)
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}
	
	return avgSentiment, score
}

// calculateConfidence calculates confidence in sentiment analysis
func (ss *SocialScraper) calculateConfidence(posts []SocialPost) float64 {
	if len(posts) < 10 {
		return 0.3 // Low confidence with few posts
	} else if len(posts) < 50 {
		return 0.6 // Medium confidence
	} else if len(posts) < 100 {
		return 0.8 // High confidence
	}
	return 0.9 // Very high confidence
}

// extractTrendingTopics extracts trending topics from posts
func (ss *SocialScraper) extractTrendingTopics(posts []SocialPost) []string {
	hashtagMap := make(map[string]int)
	
	for _, post := range posts {
		for _, hashtag := range post.Hashtags {
			hashtagMap[hashtag]++
		}
	}
	
	var topics []string
	for hashtag, count := range hashtagMap {
		if count >= 3 { // Minimum threshold
			topics = append(topics, hashtag)
		}
	}
	
	return topics
}

// filterInfluencerPosts filters posts from influencers
func (ss *SocialScraper) filterInfluencerPosts(posts []SocialPost) []SocialPost {
	var influencerPosts []SocialPost
	
	for _, post := range posts {
		if post.IsInfluencer || post.Reach > ss.config.MinFollowers {
			influencerPosts = append(influencerPosts, post)
		}
	}
	
	return influencerPosts
}

// countSentimentBreakdown counts positive, negative, and neutral mentions
func (ss *SocialScraper) countSentimentBreakdown(posts []SocialPost) (int64, int64, int64) {
	var positive, negative, neutral int64
	
	for _, post := range posts {
		if post.Sentiment > 0.1 {
			positive++
		} else if post.Sentiment < -0.1 {
			negative++
		} else {
			neutral++
		}
	}
	
	return positive, negative, neutral
}

// calculateAverageSentiment calculates average sentiment from posts
func (ss *SocialScraper) calculateAverageSentiment(posts []SocialPost) float64 {
	if len(posts) == 0 {
		return 0
	}
	
	var total float64
	for _, post := range posts {
		total += post.Sentiment
	}
	
	return total / float64(len(posts))
}

// calculateGrowthRate calculates growth rate for a topic
func (ss *SocialScraper) calculateGrowthRate(topic string, posts []SocialPost) float64 {
	// Simplified growth calculation based on recent posts
	now := time.Now()
	recentPosts := 0
	
	for _, post := range posts {
		if now.Sub(post.PostedAt) < time.Hour*6 {
			recentPosts++
		}
	}
	
	if len(posts) == 0 {
		return 0
	}
	
	return float64(recentPosts) / float64(len(posts)) * 100
}

// extractSymbolsFromPosts extracts crypto symbols mentioned in posts
func (ss *SocialScraper) extractSymbolsFromPosts(posts []SocialPost) []string {
	symbolMap := make(map[string]bool)
	symbolRegex := regexp.MustCompile(`(?i)\$?(BTC|ETH|ADA|DOT|LINK|UNI|AAVE|SOL|MATIC|AVAX|DOGE|SHIB|XRP|LTC|BCH)`)
	
	for _, post := range posts {
		matches := symbolRegex.FindAllStringSubmatch(post.Content, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				symbolMap[strings.ToUpper(match[1])] = true
			}
		}
	}
	
	var symbols []string
	for symbol := range symbolMap {
		symbols = append(symbols, symbol)
	}
	
	return symbols
}

// groupPostsByTopic groups posts by detected topics/keywords
func (ss *SocialScraper) groupPostsByTopic(posts []SocialPost) map[string][]SocialPost {
	topicMap := make(map[string][]SocialPost)
	
	cryptoKeywords := []string{"bitcoin", "ethereum", "defi", "nft", "crypto", "blockchain", "trading"}
	
	for _, post := range posts {
		content := strings.ToLower(post.Content)
		for _, keyword := range cryptoKeywords {
			if strings.Contains(content, keyword) {
				topicMap[keyword] = append(topicMap[keyword], post)
			}
		}
	}
	
	return topicMap
}

// Mock functions for demonstration
func mockScrapeTwitterPosts(ctx context.Context, url string) ([]SocialPost, error) {
	return []SocialPost{
		{
			ID:          "twitter_1",
			Platform:    "twitter",
			Content:     "Bitcoin is looking bullish! $BTC to the moon 🚀",
			Author:      "crypto_trader",
			PostedAt:    time.Now().Add(-time.Hour),
			Sentiment:   0.8,
			Engagement:  150,
			Reach:       5000,
			Symbols:     []string{"BTC"},
			Hashtags:    []string{"#Bitcoin", "#Crypto"},
			IsInfluencer: true,
		},
	}, nil
}

func mockScrapeRedditPosts(ctx context.Context, url string) ([]SocialPost, error) {
	return []SocialPost{
		{
			ID:          "reddit_1",
			Platform:    "reddit",
			Content:     "Ethereum 2.0 staking rewards are amazing",
			Author:      "eth_holder",
			PostedAt:    time.Now().Add(-time.Hour * 2),
			Sentiment:   0.6,
			Engagement:  75,
			Reach:       2000,
			Symbols:     []string{"ETH"},
			Hashtags:    []string{},
			IsInfluencer: false,
		},
	}, nil
}

func mockSearchSocial(ctx context.Context, platform, query string) ([]SocialPost, error) {
	return []SocialPost{
		{
			ID:          fmt.Sprintf("%s_search_1", platform),
			Platform:    platform,
			Content:     fmt.Sprintf("Search result for %s", query),
			Author:      "user123",
			PostedAt:    time.Now(),
			Sentiment:   0.5,
			Engagement:  50,
			Reach:       1000,
			Symbols:     []string{"BTC"},
			Hashtags:    []string{},
			IsInfluencer: false,
		},
	}, nil
}
