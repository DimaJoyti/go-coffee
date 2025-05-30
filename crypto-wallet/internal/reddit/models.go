package reddit

import (
	"time"
)

// RedditPost represents a Reddit post
type RedditPost struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Author       string    `json:"author"`
	Subreddit    string    `json:"subreddit"`
	Score        int       `json:"score"`
	Upvotes      int       `json:"upvotes"`
	Downvotes    int       `json:"downvotes"`
	Comments     int       `json:"comments"`
	URL          string    `json:"url"`
	Permalink    string    `json:"permalink"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedUTC   time.Time `json:"created_utc"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsStickied   bool      `json:"is_stickied"`
	IsLocked     bool      `json:"is_locked"`
	IsNSFW       bool      `json:"is_nsfw"`
	Flair        string    `json:"flair"`
	Domain       string    `json:"domain"`
	MediaURL     string    `json:"media_url"`
	ThumbnailURL string    `json:"thumbnail_url"`
	UpvoteRatio  float64   `json:"upvote_ratio"`
	NumComments  int       `json:"num_comments"`
	IsVideo      bool      `json:"is_video"`
	IsSelf       bool      `json:"is_self"`
	NSFW         bool      `json:"nsfw"`
	Spoiler      bool      `json:"spoiler"`
	Locked       bool      `json:"locked"`
	Stickied     bool      `json:"stickied"`
	ProcessedAt  time.Time `json:"processed_at"`
}

// RedditComment represents a Reddit comment
type RedditComment struct {
	ID          string    `json:"id"`
	ParentID    string    `json:"parent_id"`
	PostID      string    `json:"post_id"`
	Content     string    `json:"content"`
	Author      string    `json:"author"`
	Subreddit   string    `json:"subreddit"`
	Score       int       `json:"score"`
	Upvotes     int       `json:"upvotes"`
	Downvotes   int       `json:"downvotes"`
	Depth       int       `json:"depth"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedUTC  time.Time `json:"created_utc"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsStickied  bool      `json:"is_stickied"`
	IsLocked    bool      `json:"is_locked"`
	Permalink   string    `json:"permalink"`
	IsSubmitter bool      `json:"is_submitter"`
	ProcessedAt time.Time `json:"processed_at"`
}

// APIResponse represents a Reddit API response
type APIResponse struct {
	Kind string      `json:"kind"`
	Data interface{} `json:"data"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query      string   `json:"query"`
	Subreddit  string   `json:"subreddit"`
	Sort       string   `json:"sort"`
	Time       string   `json:"time"`
	Type       string   `json:"type"`
	Limit      int      `json:"limit"`
	After      string   `json:"after"`
	Before     string   `json:"before"`
	Categories []string `json:"categories"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	Posts    []RedditPost    `json:"posts"`
	Comments []RedditComment `json:"comments"`
	After    string          `json:"after"`
	Before   string          `json:"before"`
	Total    int             `json:"total"`
	Count    int             `json:"count"`
	HasMore  bool            `json:"has_more"`
}

// ContentClassification represents content analysis results
type ContentClassification struct {
	ID          string           `json:"id"`
	Type        string           `json:"type"` // "post" or "comment"
	Category    string           `json:"category"`
	Sentiment   SentimentAnalysis `json:"sentiment"`
	Topics      []TopicAnalysis  `json:"topics"`
	Keywords    []string         `json:"keywords"`
	Language    string           `json:"language"`
	Confidence  float64          `json:"confidence"`
	Toxicity    ToxicityAnalysis `json:"toxicity"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// SentimentAnalysis represents sentiment analysis results
type SentimentAnalysis struct {
	Score      float64 `json:"score"`      // -1.0 to 1.0
	Label      string  `json:"label"`      // "positive", "negative", "neutral"
	Confidence float64 `json:"confidence"` // 0.0 to 1.0
}

// TopicAnalysis represents topic analysis results
type TopicAnalysis struct {
	Topic       string   `json:"topic"`
	Keywords    []string `json:"keywords"`
	Probability float64  `json:"probability"`
	Relevance   float64  `json:"relevance"`
}

// ToxicityAnalysis represents toxicity analysis results
type ToxicityAnalysis struct {
	Score      float64 `json:"score"`      // 0.0 to 1.0
	IsToxic    bool    `json:"is_toxic"`
	Categories []string `json:"categories"` // harassment, hate_speech, etc.
	Confidence float64 `json:"confidence"`
}
