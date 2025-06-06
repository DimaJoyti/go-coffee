package brightdata

import (
	"time"

	"github.com/shopspring/decimal"
)

// NewsArticle represents a crypto news article
type NewsArticle struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Content     string    `json:"content"`
	URL         string    `json:"url"`
	Source      string    `json:"source"`
	Author      string    `json:"author"`
	PublishedAt time.Time `json:"published_at"`
	Sentiment   float64   `json:"sentiment"` // -1 to 1 (negative to positive)
	Relevance   float64   `json:"relevance"` // 0 to 1 (relevance to crypto)
	Symbols     []string  `json:"symbols"`   // Related crypto symbols
	Tags        []string  `json:"tags"`      // Article tags
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at"`
}

// SocialPost represents a social media post
type SocialPost struct {
	ID          string    `json:"id"`
	Platform    string    `json:"platform"`    // twitter, reddit, telegram
	Content     string    `json:"content"`
	Author      string    `json:"author"`
	AuthorID    string    `json:"author_id"`
	URL         string    `json:"url"`
	PostedAt    time.Time `json:"posted_at"`
	Sentiment   float64   `json:"sentiment"`   // -1 to 1
	Engagement  int64     `json:"engagement"`  // likes, retweets, comments
	Reach       int64     `json:"reach"`       // followers, views
	Symbols     []string  `json:"symbols"`     // Mentioned crypto symbols
	Hashtags    []string  `json:"hashtags"`
	Mentions    []string  `json:"mentions"`
	IsInfluencer bool     `json:"is_influencer"`
	CreatedAt   time.Time `json:"created_at"`
}

// SentimentAnalysis represents aggregated sentiment data
type SentimentAnalysis struct {
	Symbol           string                    `json:"symbol"`
	OverallSentiment float64                   `json:"overall_sentiment"` // -1 to 1
	SentimentScore   int                       `json:"sentiment_score"`   // 0-100
	Confidence       float64                   `json:"confidence"`        // 0 to 1
	TotalMentions    int64                     `json:"total_mentions"`
	PositiveMentions int64                     `json:"positive_mentions"`
	NegativeMentions int64                     `json:"negative_mentions"`
	NeutralMentions  int64                     `json:"neutral_mentions"`
	PlatformBreakdown map[string]PlatformSentiment `json:"platform_breakdown"`
	TrendingTopics   []string                  `json:"trending_topics"`
	InfluencerPosts  []SocialPost              `json:"influencer_posts"`
	TimeRange        string                    `json:"time_range"` // 1h, 24h, 7d
	LastUpdated      time.Time                 `json:"last_updated"`
}

// PlatformSentiment represents sentiment data for a specific platform
type PlatformSentiment struct {
	Platform         string  `json:"platform"`
	Sentiment        float64 `json:"sentiment"`
	Mentions         int64   `json:"mentions"`
	PositiveMentions int64   `json:"positive_mentions"`
	NegativeMentions int64   `json:"negative_mentions"`
	NeutralMentions  int64   `json:"neutral_mentions"`
	TopPosts         []SocialPost `json:"top_posts"`
}

// MarketInsight represents market intelligence data
type MarketInsight struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`        // news, social, technical, fundamental
	Category    string                 `json:"category"`    // bullish, bearish, neutral
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Impact      string                 `json:"impact"`      // high, medium, low
	Confidence  float64                `json:"confidence"`  // 0 to 1
	Symbols     []string               `json:"symbols"`
	Source      string                 `json:"source"`
	URL         string                 `json:"url"`
	Data        map[string]interface{} `json:"data"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   *time.Time             `json:"expires_at"`
}

// TrendingTopic represents a trending topic in crypto
type TrendingTopic struct {
	Topic       string    `json:"topic"`
	Mentions    int64     `json:"mentions"`
	Sentiment   float64   `json:"sentiment"`
	Growth      float64   `json:"growth"`      // percentage growth in mentions
	Symbols     []string  `json:"symbols"`     // Related symbols
	Platforms   []string  `json:"platforms"`   // Where it's trending
	LastUpdated time.Time `json:"last_updated"`
}

// NewsSource represents a news source configuration
type NewsSource struct {
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Selectors   Selectors `json:"selectors"`
	Enabled     bool     `json:"enabled"`
	Priority    int      `json:"priority"`    // 1-10, higher is more important
	UpdateFreq  string   `json:"update_freq"` // 5m, 15m, 30m, 1h
}

// Selectors represents CSS selectors for scraping
type Selectors struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	Author      string `json:"author"`
	PublishedAt string `json:"published_at"`
	ImageURL    string `json:"image_url"`
}

// SocialSource represents a social media source configuration
type SocialSource struct {
	Platform    string   `json:"platform"`
	Keywords    []string `json:"keywords"`
	Hashtags    []string `json:"hashtags"`
	Accounts    []string `json:"accounts"`    // Specific accounts to monitor
	Enabled     bool     `json:"enabled"`
	UpdateFreq  string   `json:"update_freq"`
}

// DataQualityMetrics represents quality metrics for Bright Data
type DataQualityMetrics struct {
	Source           string        `json:"source"`
	LastUpdate       time.Time     `json:"last_update"`
	UpdateFrequency  float64       `json:"update_frequency"` // updates per hour
	SuccessRate      float64       `json:"success_rate"`     // 0 to 1
	AverageLatency   time.Duration `json:"average_latency"`
	ErrorCount       int64         `json:"error_count"`
	DataFreshness    time.Duration `json:"data_freshness"`
	QualityScore     float64       `json:"quality_score"` // 0 to 1
}

// SearchResult represents a search result from Bright Data
type SearchResult struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Source      string    `json:"source"`
	Timestamp   time.Time `json:"timestamp"`
	Relevance   float64   `json:"relevance"`
}

// ScrapedContent represents scraped content from a webpage
type ScrapedContent struct {
	URL         string                 `json:"url"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata"`
	ScrapedAt   time.Time              `json:"scraped_at"`
	ContentType string                 `json:"content_type"` // article, post, comment
	WordCount   int                    `json:"word_count"`
	Language    string                 `json:"language"`
}

// InfluencerProfile represents a crypto influencer
type InfluencerProfile struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Platform     string    `json:"platform"`
	DisplayName  string    `json:"display_name"`
	Bio          string    `json:"bio"`
	Followers    int64     `json:"followers"`
	Following    int64     `json:"following"`
	PostCount    int64     `json:"post_count"`
	Verified     bool      `json:"verified"`
	InfluenceScore float64 `json:"influence_score"` // 0 to 100
	Specialties  []string  `json:"specialties"`     // DeFi, NFT, Trading, etc.
	AvatarURL    string    `json:"avatar_url"`
	ProfileURL   string    `json:"profile_url"`
	LastActive   time.Time `json:"last_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// MarketEvent represents a significant market event
type MarketEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`        // announcement, partnership, hack, regulation
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Impact      string                 `json:"impact"`      // bullish, bearish, neutral
	Severity    string                 `json:"severity"`    // critical, high, medium, low
	Symbols     []string               `json:"symbols"`
	Sources     []string               `json:"sources"`
	EventTime   time.Time              `json:"event_time"`
	DetectedAt  time.Time              `json:"detected_at"`
	Metadata    map[string]interface{} `json:"metadata"`
	PriceImpact *PriceImpact           `json:"price_impact,omitempty"`
}

// PriceImpact represents the price impact of a market event
type PriceImpact struct {
	Symbol           string          `json:"symbol"`
	PriceBefore      decimal.Decimal `json:"price_before"`
	PriceAfter       decimal.Decimal `json:"price_after"`
	PercentChange    decimal.Decimal `json:"percent_change"`
	VolumeIncrease   decimal.Decimal `json:"volume_increase"`
	TimeToImpact     time.Duration   `json:"time_to_impact"`
	MaxPriceChange   decimal.Decimal `json:"max_price_change"`
	RecoveryTime     *time.Duration  `json:"recovery_time,omitempty"`
}

// AlertRule represents a rule for generating alerts from Bright Data
type AlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`        // sentiment, news, social, event
	Conditions  map[string]interface{} `json:"conditions"`
	Symbols     []string               `json:"symbols"`
	Enabled     bool                   `json:"enabled"`
	Priority    string                 `json:"priority"`    // critical, high, medium, low
	Actions     []string               `json:"actions"`     // email, webhook, push
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	LastTriggered *time.Time           `json:"last_triggered,omitempty"`
}

// BrightDataConfig represents configuration for Bright Data service
type BrightDataConfig struct {
	Enabled         bool           `json:"enabled"`
	UpdateInterval  time.Duration  `json:"update_interval"`
	NewsSources     []NewsSource   `json:"news_sources"`
	SocialSources   []SocialSource `json:"social_sources"`
	MaxConcurrent   int            `json:"max_concurrent"`
	CacheTTL        time.Duration  `json:"cache_ttl"`
	RateLimitRPS    int            `json:"rate_limit_rps"`
	EnableSentiment bool           `json:"enable_sentiment"`
	EnableNews      bool           `json:"enable_news"`
	EnableSocial    bool           `json:"enable_social"`
	EnableEvents    bool           `json:"enable_events"`
}
