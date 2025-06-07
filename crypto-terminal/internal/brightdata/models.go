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

// TradingSignal represents a trading signal from 3commas or other sources
type TradingSignal struct {
	ID              string                 `json:"id"`
	Source          string                 `json:"source"`          // 3commas, tradingview, custom
	Type            string                 `json:"type"`            // buy, sell, hold
	Symbol          string                 `json:"symbol"`
	Exchange        string                 `json:"exchange"`
	Price           decimal.Decimal        `json:"price"`
	TargetPrice     *decimal.Decimal       `json:"target_price,omitempty"`
	StopLoss        *decimal.Decimal       `json:"stop_loss,omitempty"`
	Confidence      decimal.Decimal        `json:"confidence"`      // 0 to 100
	Strength        string                 `json:"strength"`        // weak, moderate, strong
	TimeFrame       string                 `json:"time_frame"`      // 1m, 5m, 15m, 1h, 4h, 1d
	Strategy        string                 `json:"strategy"`        // RSI, MACD, Bollinger, etc.
	Indicators      map[string]interface{} `json:"indicators"`
	RiskLevel       string                 `json:"risk_level"`      // low, medium, high
	ExpectedReturn  decimal.Decimal        `json:"expected_return"` // percentage
	MaxDrawdown     decimal.Decimal        `json:"max_drawdown"`    // percentage
	Description     string                 `json:"description"`
	Tags            []string               `json:"tags"`
	CreatedAt       time.Time              `json:"created_at"`
	ExpiresAt       *time.Time             `json:"expires_at,omitempty"`
	Status          string                 `json:"status"`          // active, expired, executed
	Performance     *SignalPerformance     `json:"performance,omitempty"`
}

// SignalPerformance tracks the performance of a trading signal
type SignalPerformance struct {
	SignalID        string          `json:"signal_id"`
	EntryPrice      decimal.Decimal `json:"entry_price"`
	ExitPrice       *decimal.Decimal `json:"exit_price,omitempty"`
	CurrentPrice    decimal.Decimal `json:"current_price"`
	UnrealizedPnL   decimal.Decimal `json:"unrealized_pnl"`
	RealizedPnL     *decimal.Decimal `json:"realized_pnl,omitempty"`
	ReturnPct       decimal.Decimal `json:"return_pct"`
	MaxGain         decimal.Decimal `json:"max_gain"`
	MaxLoss         decimal.Decimal `json:"max_loss"`
	Duration        time.Duration   `json:"duration"`
	IsActive        bool            `json:"is_active"`
	ExecutedAt      *time.Time      `json:"executed_at,omitempty"`
	ClosedAt        *time.Time      `json:"closed_at,omitempty"`
	LastUpdated     time.Time       `json:"last_updated"`
}

// TradingBot represents a 3commas trading bot
type TradingBot struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Type            string                 `json:"type"`            // simple, composite, grid
	Status          string                 `json:"status"`          // enabled, disabled, archived
	Exchange        string                 `json:"exchange"`
	Pairs           []string               `json:"pairs"`
	Strategy        string                 `json:"strategy"`
	BaseOrderSize   decimal.Decimal        `json:"base_order_size"`
	SafetyOrderSize decimal.Decimal        `json:"safety_order_size"`
	MaxSafetyOrders int                    `json:"max_safety_orders"`
	TakeProfit      decimal.Decimal        `json:"take_profit"`
	StopLoss        *decimal.Decimal       `json:"stop_loss,omitempty"`
	TrailingEnabled bool                   `json:"trailing_enabled"`
	TrailingDeviation decimal.Decimal      `json:"trailing_deviation"`
	ActiveDeals     int                    `json:"active_deals"`
	CompletedDeals  int                    `json:"completed_deals"`
	TotalProfit     decimal.Decimal        `json:"total_profit"`
	TotalProfitPct  decimal.Decimal        `json:"total_profit_pct"`
	WinRate         decimal.Decimal        `json:"win_rate"`
	AvgDealTime     time.Duration          `json:"avg_deal_time"`
	MaxDrawdown     decimal.Decimal        `json:"max_drawdown"`
	CreatedAt       time.Time              `json:"created_at"`
	LastUpdated     time.Time              `json:"last_updated"`
	Settings        map[string]interface{} `json:"settings"`
	Performance     *BotPerformance        `json:"performance,omitempty"`
}

// BotPerformance represents detailed bot performance metrics
type BotPerformance struct {
	BotID           string          `json:"bot_id"`
	TotalTrades     int             `json:"total_trades"`
	WinningTrades   int             `json:"winning_trades"`
	LosingTrades    int             `json:"losing_trades"`
	WinRate         decimal.Decimal `json:"win_rate"`
	AvgWin          decimal.Decimal `json:"avg_win"`
	AvgLoss         decimal.Decimal `json:"avg_loss"`
	ProfitFactor    decimal.Decimal `json:"profit_factor"`
	SharpeRatio     decimal.Decimal `json:"sharpe_ratio"`
	MaxDrawdown     decimal.Decimal `json:"max_drawdown"`
	MaxDrawdownDate time.Time       `json:"max_drawdown_date"`
	TotalVolume     decimal.Decimal `json:"total_volume"`
	AvgDealTime     time.Duration   `json:"avg_deal_time"`
	BestTrade       decimal.Decimal `json:"best_trade"`
	WorstTrade      decimal.Decimal `json:"worst_trade"`
	LastTradeAt     *time.Time      `json:"last_trade_at,omitempty"`
	LastUpdated     time.Time       `json:"last_updated"`
}

// TradingDeal represents an active or completed trading deal
type TradingDeal struct {
	ID              string          `json:"id"`
	BotID           string          `json:"bot_id"`
	BotName         string          `json:"bot_name"`
	Symbol          string          `json:"symbol"`
	Exchange        string          `json:"exchange"`
	Status          string          `json:"status"`          // active, completed, cancelled, failed
	Type            string          `json:"type"`            // long, short
	BaseOrderSize   decimal.Decimal `json:"base_order_size"`
	SafetyOrders    int             `json:"safety_orders"`
	CompletedOrders int             `json:"completed_orders"`
	AveragePrice    decimal.Decimal `json:"average_price"`
	CurrentPrice    decimal.Decimal `json:"current_price"`
	TakeProfit      decimal.Decimal `json:"take_profit"`
	StopLoss        *decimal.Decimal `json:"stop_loss,omitempty"`
	UnrealizedPnL   decimal.Decimal `json:"unrealized_pnl"`
	RealizedPnL     *decimal.Decimal `json:"realized_pnl,omitempty"`
	TotalInvested   decimal.Decimal `json:"total_invested"`
	MaxFunds        decimal.Decimal `json:"max_funds"`
	CreatedAt       time.Time       `json:"created_at"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty"`
	LastUpdated     time.Time       `json:"last_updated"`
}

// TechnicalIndicator represents a technical analysis indicator
type TechnicalIndicator struct {
	Name        string                 `json:"name"`        // RSI, MACD, SMA, EMA, etc.
	Symbol      string                 `json:"symbol"`
	TimeFrame   string                 `json:"time_frame"`  // 1m, 5m, 15m, 1h, 4h, 1d
	Value       decimal.Decimal        `json:"value"`
	Signal      string                 `json:"signal"`      // buy, sell, neutral
	Strength    decimal.Decimal        `json:"strength"`    // 0 to 100
	Parameters  map[string]interface{} `json:"parameters"`
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source"`      // tradingview, taapi, custom
}

// TechnicalAnalysis represents comprehensive technical analysis for a symbol
type TechnicalAnalysis struct {
	Symbol          string               `json:"symbol"`
	Exchange        string               `json:"exchange"`
	TimeFrame       string               `json:"time_frame"`
	OverallSignal   string               `json:"overall_signal"`   // strong_buy, buy, neutral, sell, strong_sell
	OverallScore    decimal.Decimal      `json:"overall_score"`    // 0 to 100
	Indicators      []TechnicalIndicator `json:"indicators"`
	SupportLevels   []decimal.Decimal    `json:"support_levels"`
	ResistanceLevels []decimal.Decimal   `json:"resistance_levels"`
	TrendDirection  string               `json:"trend_direction"`  // bullish, bearish, sideways
	TrendStrength   decimal.Decimal      `json:"trend_strength"`   // 0 to 100
	Volatility      decimal.Decimal      `json:"volatility"`
	Volume          VolumeAnalysis       `json:"volume"`
	Patterns        []ChartPattern       `json:"patterns"`
	LastUpdated     time.Time            `json:"last_updated"`
	Source          string               `json:"source"`
}

// VolumeAnalysis represents volume analysis data
type VolumeAnalysis struct {
	CurrentVolume   decimal.Decimal `json:"current_volume"`
	AverageVolume   decimal.Decimal `json:"average_volume"`
	VolumeRatio     decimal.Decimal `json:"volume_ratio"`
	VolumeProfile   string          `json:"volume_profile"`   // accumulation, distribution, neutral
	VolumeSignal    string          `json:"volume_signal"`    // bullish, bearish, neutral
	OnBalanceVolume decimal.Decimal `json:"on_balance_volume"`
}

// ChartPattern represents a detected chart pattern
type ChartPattern struct {
	Name        string          `json:"name"`        // head_and_shoulders, triangle, flag, etc.
	Type        string          `json:"type"`        // bullish, bearish, neutral
	Confidence  decimal.Decimal `json:"confidence"`  // 0 to 100
	Target      *decimal.Decimal `json:"target,omitempty"`
	StopLoss    *decimal.Decimal `json:"stop_loss,omitempty"`
	Breakout    *decimal.Decimal `json:"breakout,omitempty"`
	DetectedAt  time.Time       `json:"detected_at"`
	Status      string          `json:"status"`      // forming, confirmed, invalidated
}
