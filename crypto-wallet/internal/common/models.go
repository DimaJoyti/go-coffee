package common

import (
	"time"
)

// RedditPost represents a Reddit post
type RedditPost struct {
	ID              string            `json:"id" db:"id"`
	Title           string            `json:"title" db:"title"`
	Content         string            `json:"content" db:"content"`
	Author          string            `json:"author" db:"author"`
	Subreddit       string            `json:"subreddit" db:"subreddit"`
	URL             string            `json:"url" db:"url"`
	Score           int               `json:"score" db:"score"`
	UpvoteRatio     float64           `json:"upvote_ratio" db:"upvote_ratio"`
	NumComments     int               `json:"num_comments" db:"num_comments"`
	CreatedUTC      time.Time         `json:"created_utc" db:"created_utc"`
	IsVideo         bool              `json:"is_video" db:"is_video"`
	IsSelf          bool              `json:"is_self" db:"is_self"`
	Permalink       string            `json:"permalink" db:"permalink"`
	Flair           string            `json:"flair" db:"flair"`
	NSFW            bool              `json:"nsfw" db:"nsfw"`
	Spoiler         bool              `json:"spoiler" db:"spoiler"`
	Locked          bool              `json:"locked" db:"locked"`
	Stickied        bool              `json:"stickied" db:"stickied"`
	Metadata        map[string]string `json:"metadata" db:"metadata"`
	ProcessedAt     time.Time         `json:"processed_at" db:"processed_at"`
	Classification  string            `json:"classification" db:"classification"`
	Sentiment       string            `json:"sentiment" db:"sentiment"`
	Topics          []string          `json:"topics" db:"topics"`
	Confidence      float64           `json:"confidence" db:"confidence"`
	EmbeddingVector []float64         `json:"embedding_vector" db:"embedding_vector"`
}

// RedditComment represents a Reddit comment
type RedditComment struct {
	ID              string            `json:"id" db:"id"`
	PostID          string            `json:"post_id" db:"post_id"`
	ParentID        string            `json:"parent_id" db:"parent_id"`
	Content         string            `json:"content" db:"content"`
	Author          string            `json:"author" db:"author"`
	Score           int               `json:"score" db:"score"`
	CreatedUTC      time.Time         `json:"created_utc" db:"created_utc"`
	IsSubmitter     bool              `json:"is_submitter" db:"is_submitter"`
	Depth           int               `json:"depth" db:"depth"`
	Permalink       string            `json:"permalink" db:"permalink"`
	Metadata        map[string]string `json:"metadata" db:"metadata"`
	ProcessedAt     time.Time         `json:"processed_at" db:"processed_at"`
	Classification  string            `json:"classification" db:"classification"`
	Sentiment       string            `json:"sentiment" db:"sentiment"`
	Topics          []string          `json:"topics" db:"topics"`
	Confidence      float64           `json:"confidence" db:"confidence"`
	EmbeddingVector []float64         `json:"embedding_vector" db:"embedding_vector"`
}

// ContentClassification represents content classification result
type ContentClassification struct {
	ID          string            `json:"id"`
	ContentID   string            `json:"content_id"`
	ContentType string            `json:"content_type"` // post, comment
	Category    string            `json:"category"`
	Subcategory string            `json:"subcategory"`
	Tags        []string          `json:"tags"`
	Sentiment   SentimentAnalysis `json:"sentiment"`
	Topics      []TopicAnalysis   `json:"topics"`
	Confidence  float64           `json:"confidence"`
	ModelUsed   string            `json:"model_used"`
	ProcessedAt time.Time         `json:"processed_at"`
	Metadata    map[string]string `json:"metadata"`
}

// SentimentAnalysis represents sentiment analysis result
type SentimentAnalysis struct {
	Label        string  `json:"label"`      // positive, negative, neutral
	Score        float64 `json:"score"`      // confidence score
	Magnitude    float64 `json:"magnitude"`  // intensity of sentiment
	Subjectivity float64 `json:"subjectivity"` // objective vs subjective
}

// TopicAnalysis represents topic modeling result
type TopicAnalysis struct {
	Topic       string  `json:"topic"`
	Keywords    []string `json:"keywords"`
	Probability float64 `json:"probability"`
	Relevance   float64 `json:"relevance"`
}

// TrendAnalysis represents trend analysis result
type TrendAnalysis struct {
	ID          string                 `json:"id"`
	Timeframe   string                 `json:"timeframe"` // hourly, daily, weekly
	Subreddit   string                 `json:"subreddit"`
	Category    string                 `json:"category"`
	TrendType   string                 `json:"trend_type"` // rising, declining, stable
	Metrics     TrendMetrics           `json:"metrics"`
	Keywords    []KeywordTrend         `json:"keywords"`
	Sentiment   SentimentTrend         `json:"sentiment"`
	Topics      []TopicTrend           `json:"topics"`
	Predictions []TrendPrediction      `json:"predictions"`
	GeneratedAt time.Time              `json:"generated_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TrendMetrics represents trend metrics
type TrendMetrics struct {
	PostCount       int     `json:"post_count"`
	CommentCount    int     `json:"comment_count"`
	AvgScore        float64 `json:"avg_score"`
	AvgComments     float64 `json:"avg_comments"`
	EngagementRate  float64 `json:"engagement_rate"`
	GrowthRate      float64 `json:"growth_rate"`
	VelocityScore   float64 `json:"velocity_score"`
}

// KeywordTrend represents keyword trend analysis
type KeywordTrend struct {
	Keyword     string  `json:"keyword"`
	Frequency   int     `json:"frequency"`
	Growth      float64 `json:"growth"`
	Sentiment   float64 `json:"sentiment"`
	Relevance   float64 `json:"relevance"`
	TrendScore  float64 `json:"trend_score"`
}

// SentimentTrend represents sentiment trend over time
type SentimentTrend struct {
	Overall     float64            `json:"overall"`
	Positive    float64            `json:"positive"`
	Negative    float64            `json:"negative"`
	Neutral     float64            `json:"neutral"`
	Volatility  float64            `json:"volatility"`
	Trajectory  string             `json:"trajectory"` // improving, declining, stable
	Breakdown   map[string]float64 `json:"breakdown"`
}

// TopicTrend represents topic trend analysis
type TopicTrend struct {
	Topic       string  `json:"topic"`
	Frequency   int     `json:"frequency"`
	Growth      float64 `json:"growth"`
	Engagement  float64 `json:"engagement"`
	Sentiment   float64 `json:"sentiment"`
	TrendScore  float64 `json:"trend_score"`
}

// TrendPrediction represents trend predictions
type TrendPrediction struct {
	Timeframe   string  `json:"timeframe"` // next_hour, next_day, next_week
	Metric      string  `json:"metric"`
	Prediction  float64 `json:"prediction"`
	Confidence  float64 `json:"confidence"`
	Direction   string  `json:"direction"` // up, down, stable
}

// ContentFilter represents content filtering criteria
type ContentFilter struct {
	Subreddits    []string          `json:"subreddits"`
	Keywords      []string          `json:"keywords"`
	Authors       []string          `json:"authors"`
	MinScore      int               `json:"min_score"`
	MaxScore      int               `json:"max_score"`
	MinComments   int               `json:"min_comments"`
	MaxComments   int               `json:"max_comments"`
	TimeRange     TimeRange         `json:"time_range"`
	ContentTypes  []string          `json:"content_types"` // text, image, video, link
	Flairs        []string          `json:"flairs"`
	ExcludeNSFW   bool              `json:"exclude_nsfw"`
	ExcludeSpoiler bool             `json:"exclude_spoiler"`
	Metadata      map[string]string `json:"metadata"`
}

// TimeRange represents a time range filter
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ProcessingJob represents a content processing job
type ProcessingJob struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // classification, sentiment, topic_modeling, trend_analysis
	Status      string                 `json:"status"` // pending, processing, completed, failed
	ContentIDs  []string               `json:"content_ids"`
	Filter      ContentFilter          `json:"filter"`
	Config      map[string]interface{} `json:"config"`
	Progress    float64                `json:"progress"`
	Results     map[string]interface{} `json:"results"`
	Error       string                 `json:"error"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at"`
	Metadata    map[string]string      `json:"metadata"`
}

// APIResponse represents Reddit API response structure
type APIResponse struct {
	Kind string      `json:"kind"`
	Data interface{} `json:"data"`
}

// ListingData represents Reddit listing data
type ListingData struct {
	After    string        `json:"after"`
	Before   string        `json:"before"`
	Children []interface{} `json:"children"`
	Dist     int           `json:"dist"`
	ModHash  string        `json:"modhash"`
}

// PostData represents Reddit post data from API
type PostData struct {
	ID              string      `json:"id"`
	Title           string      `json:"title"`
	Selftext        string      `json:"selftext"`
	Author          string      `json:"author"`
	Subreddit       string      `json:"subreddit"`
	URL             string      `json:"url"`
	Score           int         `json:"score"`
	UpvoteRatio     float64     `json:"upvote_ratio"`
	NumComments     int         `json:"num_comments"`
	CreatedUTC      float64     `json:"created_utc"`
	IsVideo         bool        `json:"is_video"`
	IsSelf          bool        `json:"is_self"`
	Permalink       string      `json:"permalink"`
	LinkFlairText   string      `json:"link_flair_text"`
	Over18          bool        `json:"over_18"`
	Spoiler         bool        `json:"spoiler"`
	Locked          bool        `json:"locked"`
	Stickied        bool        `json:"stickied"`
	Distinguished   interface{} `json:"distinguished"`
	PostHint        string      `json:"post_hint"`
	Preview         interface{} `json:"preview"`
	Media           interface{} `json:"media"`
	MediaEmbed      interface{} `json:"media_embed"`
	SecureMediaEmbed interface{} `json:"secure_media_embed"`
}

// CommentData represents Reddit comment data from API
type CommentData struct {
	ID           string      `json:"id"`
	ParentID     string      `json:"parent_id"`
	Body         string      `json:"body"`
	Author       string      `json:"author"`
	Score        int         `json:"score"`
	CreatedUTC   float64     `json:"created_utc"`
	IsSubmitter  bool        `json:"is_submitter"`
	Depth        int         `json:"depth"`
	Permalink    string      `json:"permalink"`
	Replies      interface{} `json:"replies"`
	Distinguished interface{} `json:"distinguished"`
	ScoreHidden  bool        `json:"score_hidden"`
	Edited       interface{} `json:"edited"`
	Gilded       int         `json:"gilded"`
	Archived     bool        `json:"archived"`
	NoFollow     bool        `json:"no_follow"`
}

// SearchRequest represents a Reddit search request
type SearchRequest struct {
	Query      string        `json:"query"`
	Subreddit  string        `json:"subreddit"`
	Sort       string        `json:"sort"`       // relevance, hot, top, new, comments
	Time       string        `json:"time"`       // hour, day, week, month, year, all
	Type       string        `json:"type"`       // link, comment
	Limit      int           `json:"limit"`
	After      string        `json:"after"`
	Before     string        `json:"before"`
	Filter     ContentFilter `json:"filter"`
}

// SearchResponse represents a Reddit search response
type SearchResponse struct {
	Posts    []RedditPost    `json:"posts"`
	Comments []RedditComment `json:"comments"`
	After    string          `json:"after"`
	Before   string          `json:"before"`
	Count    int             `json:"count"`
	HasMore  bool            `json:"has_more"`
}
