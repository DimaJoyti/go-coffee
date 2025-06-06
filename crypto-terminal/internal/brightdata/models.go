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

// TradingViewData represents scraped data from TradingView
type TradingViewData struct {
	Coins           []TradingViewCoin `json:"coins"`
	MarketOverview  MarketOverview    `json:"market_overview"`
	TrendingCoins   []TrendingCoin    `json:"trending_coins"`
	Gainers         []TradingViewCoin `json:"gainers"`
	Losers          []TradingViewCoin `json:"losers"`
	MarketCap       MarketCapData     `json:"market_cap"`
	LastUpdated     time.Time         `json:"last_updated"`
	DataQuality     float64           `json:"data_quality"`
}

// TradingViewCoin represents a cryptocurrency from TradingView
type TradingViewCoin struct {
	Symbol          string          `json:"symbol"`
	Name            string          `json:"name"`
	Price           decimal.Decimal `json:"price"`
	Change24h       decimal.Decimal `json:"change_24h"`
	ChangePercent   decimal.Decimal `json:"change_percent"`
	MarketCap       decimal.Decimal `json:"market_cap"`
	Volume24h       decimal.Decimal `json:"volume_24h"`
	CircSupply      decimal.Decimal `json:"circ_supply"`
	VolMarketCap    decimal.Decimal `json:"vol_market_cap"`
	SocialDominance decimal.Decimal `json:"social_dominance"`
	Category        []string        `json:"category"`
	TechRating      string          `json:"tech_rating"`
	Rank            int             `json:"rank"`
	LogoURL         string          `json:"logo_url"`
	LastUpdated     time.Time       `json:"last_updated"`
}

// MarketOverview represents overall market statistics
type MarketOverview struct {
	TotalMarketCap  decimal.Decimal `json:"total_market_cap"`
	TotalVolume24h  decimal.Decimal `json:"total_volume_24h"`
	BTCDominance    decimal.Decimal `json:"btc_dominance"`
	ETHDominance    decimal.Decimal `json:"eth_dominance"`
	ActiveCoins     int             `json:"active_coins"`
	MarketSentiment string          `json:"market_sentiment"`
	FearGreedIndex  int             `json:"fear_greed_index"`
	LastUpdated     time.Time       `json:"last_updated"`
}

// TrendingCoin represents a trending cryptocurrency
type TrendingCoin struct {
	Symbol      string          `json:"symbol"`
	Name        string          `json:"name"`
	Price       decimal.Decimal `json:"price"`
	Change24h   decimal.Decimal `json:"change_24h"`
	Volume24h   decimal.Decimal `json:"volume_24h"`
	TrendScore  decimal.Decimal `json:"trend_score"`
	Mentions    int64           `json:"mentions"`
	LogoURL     string          `json:"logo_url"`
	LastUpdated time.Time       `json:"last_updated"`
}

// MarketCapData represents market capitalization data
type MarketCapData struct {
	TotalMarketCap decimal.Decimal            `json:"total_market_cap"`
	Dominance      map[string]decimal.Decimal `json:"dominance"`
	TopCoins       []MarketCapCoin            `json:"top_coins"`
	LastUpdated    time.Time                  `json:"last_updated"`
}

// MarketCapCoin represents a coin in market cap ranking
type MarketCapCoin struct {
	Symbol      string          `json:"symbol"`
	Name        string          `json:"name"`
	MarketCap   decimal.Decimal `json:"market_cap"`
	Price       decimal.Decimal `json:"price"`
	Rank        int             `json:"rank"`
	LogoURL     string          `json:"logo_url"`
	LastUpdated time.Time       `json:"last_updated"`
}

// PortfolioAnalytics represents comprehensive portfolio analytics
type PortfolioAnalytics struct {
	PortfolioID     string                    `json:"portfolio_id"`
	TotalValue      decimal.Decimal           `json:"total_value"`
	TotalReturn     decimal.Decimal           `json:"total_return"`
	TotalReturnPct  decimal.Decimal           `json:"total_return_pct"`
	DayReturn       decimal.Decimal           `json:"day_return"`
	DayReturnPct    decimal.Decimal           `json:"day_return_pct"`
	Holdings        []PortfolioHolding        `json:"holdings"`
	Allocation      map[string]decimal.Decimal `json:"allocation"`
	Performance     PerformanceMetrics        `json:"performance"`
	Risk            RiskMetrics               `json:"risk"`
	Diversification DiversificationMetrics    `json:"diversification"`
	LastUpdated     time.Time                 `json:"last_updated"`
}

// PortfolioHolding represents a single holding in the portfolio
type PortfolioHolding struct {
	Symbol          string          `json:"symbol"`
	Name            string          `json:"name"`
	Quantity        decimal.Decimal `json:"quantity"`
	AvgCost         decimal.Decimal `json:"avg_cost"`
	CurrentPrice    decimal.Decimal `json:"current_price"`
	MarketValue     decimal.Decimal `json:"market_value"`
	UnrealizedPnL   decimal.Decimal `json:"unrealized_pnl"`
	UnrealizedPct   decimal.Decimal `json:"unrealized_pct"`
	Weight          decimal.Decimal `json:"weight"`
	DayChange       decimal.Decimal `json:"day_change"`
	DayChangePct    decimal.Decimal `json:"day_change_pct"`
	LastUpdated     time.Time       `json:"last_updated"`
}

// PerformanceMetrics represents portfolio performance metrics
type PerformanceMetrics struct {
	SharpeRatio     decimal.Decimal `json:"sharpe_ratio"`
	SortinoRatio    decimal.Decimal `json:"sortino_ratio"`
	CalmarRatio     decimal.Decimal `json:"calmar_ratio"`
	MaxDrawdown     decimal.Decimal `json:"max_drawdown"`
	Volatility      decimal.Decimal `json:"volatility"`
	Alpha           decimal.Decimal `json:"alpha"`
	Beta            decimal.Decimal `json:"beta"`
	TrackingError   decimal.Decimal `json:"tracking_error"`
	InformationRatio decimal.Decimal `json:"information_ratio"`
	WinRate         decimal.Decimal `json:"win_rate"`
	AvgWin          decimal.Decimal `json:"avg_win"`
	AvgLoss         decimal.Decimal `json:"avg_loss"`
	ProfitFactor    decimal.Decimal `json:"profit_factor"`
}

// RiskMetrics represents comprehensive risk metrics
type RiskMetrics struct {
	VaR95           decimal.Decimal            `json:"var_95"`
	VaR99           decimal.Decimal            `json:"var_99"`
	CVaR95          decimal.Decimal            `json:"cvar_95"`
	CVaR99          decimal.Decimal            `json:"cvar_99"`
	PortfolioVol    decimal.Decimal            `json:"portfolio_vol"`
	Correlation     map[string]decimal.Decimal `json:"correlation"`
	ConcentrationRisk decimal.Decimal          `json:"concentration_risk"`
	LiquidityRisk   decimal.Decimal            `json:"liquidity_risk"`
	CounterpartyRisk decimal.Decimal           `json:"counterparty_risk"`
	RiskScore       decimal.Decimal            `json:"risk_score"`
	StressTests     []StressTestResult         `json:"stress_tests"`
	LastUpdated     time.Time                  `json:"last_updated"`
}

// StressTestResult represents results of a stress test scenario
type StressTestResult struct {
	Scenario        string          `json:"scenario"`
	Description     string          `json:"description"`
	PnLImpact       decimal.Decimal `json:"pnl_impact"`
	PnLImpactPct    decimal.Decimal `json:"pnl_impact_pct"`
	WorstHolding    string          `json:"worst_holding"`
	WorstImpact     decimal.Decimal `json:"worst_impact"`
	RecoveryTime    string          `json:"recovery_time"`
	Probability     decimal.Decimal `json:"probability"`
}

// DiversificationMetrics represents portfolio diversification metrics
type DiversificationMetrics struct {
	HerfindahlIndex     decimal.Decimal            `json:"herfindahl_index"`
	EffectiveAssets     decimal.Decimal            `json:"effective_assets"`
	ConcentrationRatio  decimal.Decimal            `json:"concentration_ratio"`
	SectorDiversification map[string]decimal.Decimal `json:"sector_diversification"`
	GeoDiversification  map[string]decimal.Decimal `json:"geo_diversification"`
	MarketCapDiversification map[string]decimal.Decimal `json:"market_cap_diversification"`
	DiversificationScore decimal.Decimal           `json:"diversification_score"`
}

// MarketHeatmap represents market heatmap data
type MarketHeatmap struct {
	Sectors         []SectorData    `json:"sectors"`
	TopMovers       []HeatmapCoin   `json:"top_movers"`
	MarketSentiment string          `json:"market_sentiment"`
	TotalMarketCap  decimal.Decimal `json:"total_market_cap"`
	LastUpdated     time.Time       `json:"last_updated"`
}

// SectorData represents data for a market sector
type SectorData struct {
	Name            string          `json:"name"`
	MarketCap       decimal.Decimal `json:"market_cap"`
	Change24h       decimal.Decimal `json:"change_24h"`
	Volume24h       decimal.Decimal `json:"volume_24h"`
	CoinCount       int             `json:"coin_count"`
	TopCoins        []HeatmapCoin   `json:"top_coins"`
	Performance     string          `json:"performance"`
}

// HeatmapCoin represents a coin in the heatmap
type HeatmapCoin struct {
	Symbol        string          `json:"symbol"`
	Name          string          `json:"name"`
	Price         decimal.Decimal `json:"price"`
	Change24h     decimal.Decimal `json:"change_24h"`
	MarketCap     decimal.Decimal `json:"market_cap"`
	Volume24h     decimal.Decimal `json:"volume_24h"`
	Color         string          `json:"color"`
	Size          decimal.Decimal `json:"size"`
	LogoURL       string          `json:"logo_url"`
	LastUpdated   time.Time       `json:"last_updated"`
}
