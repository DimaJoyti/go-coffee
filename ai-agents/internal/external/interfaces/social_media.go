package interfaces

import (
	"context"
	"time"
)

// SocialMediaProvider defines the interface for social media platforms
type SocialMediaProvider interface {
	// Post operations
	CreatePost(ctx context.Context, req *CreatePostRequest) (*Post, error)
	GetPost(ctx context.Context, postID string) (*Post, error)
	UpdatePost(ctx context.Context, postID string, req *UpdatePostRequest) (*Post, error)
	DeletePost(ctx context.Context, postID string) error
	
	// Media operations
	UploadMedia(ctx context.Context, req *UploadMediaRequest) (*Media, error)
	GetMedia(ctx context.Context, mediaID string) (*Media, error)
	DeleteMedia(ctx context.Context, mediaID string) error
	
	// Engagement operations
	LikePost(ctx context.Context, postID string) error
	UnlikePost(ctx context.Context, postID string) error
	CommentOnPost(ctx context.Context, postID string, req *SocialCommentRequest) (*SocialComment, error)
	ReplyToComment(ctx context.Context, commentID string, req *SocialCommentRequest) (*SocialComment, error)
	SharePost(ctx context.Context, postID string, req *SocialShareRequest) (*Post, error)
	
	// User operations
	GetUser(ctx context.Context, userID string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	FollowUser(ctx context.Context, userID string) error
	UnfollowUser(ctx context.Context, userID string) error
	GetFollowers(ctx context.Context, userID string, req *PaginationRequest) (*SocialUserList, error)
	GetFollowing(ctx context.Context, userID string, req *PaginationRequest) (*SocialUserList, error)
	
	// Timeline and feed operations
	GetUserTimeline(ctx context.Context, userID string, req *TimelineRequest) (*PostList, error)
	GetHomeFeed(ctx context.Context, req *FeedRequest) (*PostList, error)
	GetTrendingPosts(ctx context.Context, req *TrendingRequest) (*PostList, error)
	
	// Search operations
	SearchPosts(ctx context.Context, req *SocialSearchRequest) (*PostSearchResult, error)
	SearchUsers(ctx context.Context, req *UserSearchRequest) (*UserSearchResult, error)
	SearchHashtags(ctx context.Context, req *HashtagSearchRequest) (*HashtagSearchResult, error)
	
	// Analytics operations
	GetPostAnalytics(ctx context.Context, postID string, req *AnalyticsRequest) (*PostAnalytics, error)
	GetUserAnalytics(ctx context.Context, userID string, req *AnalyticsRequest) (*UserAnalytics, error)
	GetHashtagAnalytics(ctx context.Context, hashtag string, req *AnalyticsRequest) (*HashtagAnalytics, error)
	
	// Hashtag operations
	GetHashtagPosts(ctx context.Context, hashtag string, req *HashtagPostsRequest) (*PostList, error)
	GetTrendingHashtags(ctx context.Context, req *TrendingHashtagsRequest) (*HashtagList, error)
	
	// Direct messaging (if supported)
	SendDirectMessage(ctx context.Context, req *DirectMessageRequest) (*DirectMessage, error)
	GetDirectMessages(ctx context.Context, req *DirectMessageListRequest) (*DirectMessageList, error)
	
	// Scheduling operations
	SchedulePost(ctx context.Context, req *SchedulePostRequest) (*ScheduledPost, error)
	GetScheduledPosts(ctx context.Context, req *ScheduledPostsRequest) (*ScheduledPostList, error)
	UpdateScheduledPost(ctx context.Context, scheduledPostID string, req *UpdateScheduledPostRequest) (*ScheduledPost, error)
	CancelScheduledPost(ctx context.Context, scheduledPostID string) error
	
	// Webhook operations
	RegisterWebhook(ctx context.Context, req *SocialWebhookRequest) (*SocialWebhook, error)
	UnregisterWebhook(ctx context.Context, webhookID string) error
	
	// Provider info
	GetProviderInfo() *ProviderInfo
}

// Post represents a social media post
type Post struct {
	ID              string                 `json:"id"`
	Text            string                 `json:"text"`
	AuthorID        string                 `json:"author_id"`
	Author          *User                  `json:"author,omitempty"`
	
	// Media attachments
	Media           []*Media               `json:"media,omitempty"`
	
	// Post metadata
	Type            PostType               `json:"type"`
	Language        string                 `json:"language,omitempty"`
	Source          string                 `json:"source,omitempty"`
	
	// Engagement metrics
	LikeCount       int                    `json:"like_count"`
	CommentCount    int                    `json:"comment_count"`
	ShareCount      int                    `json:"share_count"`
	ViewCount       int                    `json:"view_count,omitempty"`
	
	// Interaction flags
	IsLiked         bool                   `json:"is_liked"`
	IsShared        bool                   `json:"is_shared"`
	IsSaved         bool                   `json:"is_saved"`
	
	// Post properties
	IsPromoted      bool                   `json:"is_promoted"`
	IsVerified      bool                   `json:"is_verified"`
	IsSensitive     bool                   `json:"is_sensitive"`
	
	// Hashtags and mentions
	Hashtags        []string               `json:"hashtags,omitempty"`
	Mentions        []string               `json:"mentions,omitempty"`
	URLs            []string               `json:"urls,omitempty"`
	
	// Location
	Location        *Location              `json:"location,omitempty"`
	
	// Timestamps
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	PublishedAt     *time.Time             `json:"published_at,omitempty"`
	
	// Thread information
	InReplyToID     string                 `json:"in_reply_to_id,omitempty"`
	ConversationID  string                 `json:"conversation_id,omitempty"`
	
	// External references
	ExternalID      string                 `json:"external_id,omitempty"`
	URL             string                 `json:"url,omitempty"`
	PermalinkURL    string                 `json:"permalink_url,omitempty"`
}

// User represents a social media user
type User struct {
	ID              string                 `json:"id"`
	Username        string                 `json:"username"`
	DisplayName     string                 `json:"display_name"`
	Bio             string                 `json:"bio,omitempty"`
	
	// Profile information
	ProfileImageURL string                 `json:"profile_image_url,omitempty"`
	BannerImageURL  string                 `json:"banner_image_url,omitempty"`
	Website         string                 `json:"website,omitempty"`
	Location        string                 `json:"location,omitempty"`
	
	// User metrics
	FollowersCount  int                    `json:"followers_count"`
	FollowingCount  int                    `json:"following_count"`
	PostsCount      int                    `json:"posts_count"`
	LikesCount      int                    `json:"likes_count,omitempty"`
	
	// User properties
	IsVerified      bool                   `json:"is_verified"`
	IsPrivate       bool                   `json:"is_private"`
	IsFollowing     bool                   `json:"is_following"`
	IsFollowedBy    bool                   `json:"is_followed_by"`
	IsBlocked       bool                   `json:"is_blocked"`
	IsMuted         bool                   `json:"is_muted"`
	
	// Account information
	AccountType     AccountType            `json:"account_type"`
	Language        string                 `json:"language,omitempty"`
	Timezone        string                 `json:"timezone,omitempty"`
	
	// Timestamps
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	LastActiveAt    *time.Time             `json:"last_active_at,omitempty"`
	
	// External references
	ExternalID      string                 `json:"external_id,omitempty"`
	URL             string                 `json:"url,omitempty"`
}

// Media represents media content
type Media struct {
	ID              string                 `json:"id"`
	Type            MediaType              `json:"type"`
	URL             string                 `json:"url"`
	ThumbnailURL    string                 `json:"thumbnail_url,omitempty"`
	
	// Media properties
	Width           int                    `json:"width,omitempty"`
	Height          int                    `json:"height,omitempty"`
	Duration        *time.Duration         `json:"duration,omitempty"`
	Size            int64                  `json:"size,omitempty"`
	MimeType        string                 `json:"mime_type,omitempty"`
	
	// Media metadata
	AltText         string                 `json:"alt_text,omitempty"`
	Caption         string                 `json:"caption,omitempty"`
	
	// Upload information
	UploaderID      string                 `json:"uploader_id"`
	UploadedAt      time.Time              `json:"uploaded_at"`
	
	// External references
	ExternalID      string                 `json:"external_id,omitempty"`
}

// SocialComment represents a comment on a post
type SocialComment struct {
	ID              string                 `json:"id"`
	PostID          string                 `json:"post_id"`
	AuthorID        string                 `json:"author_id"`
	Author          *User                  `json:"author,omitempty"`
	Text            string                 `json:"text"`
	
	// Comment metadata
	LikeCount       int                    `json:"like_count"`
	ReplyCount      int                    `json:"reply_count"`
	IsLiked         bool                   `json:"is_liked"`
	
	// Thread information
	ParentID        string                 `json:"parent_id,omitempty"`
	
	// Timestamps
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	
	// External references
	ExternalID      string                 `json:"external_id,omitempty"`
}

// DirectMessage represents a direct message
type DirectMessage struct {
	ID              string                 `json:"id"`
	SenderID        string                 `json:"sender_id"`
	RecipientID     string                 `json:"recipient_id"`
	Text            string                 `json:"text"`
	
	// Message metadata
	IsRead          bool                   `json:"is_read"`
	Media           []*Media               `json:"media,omitempty"`
	
	// Timestamps
	CreatedAt       time.Time              `json:"created_at"`
	ReadAt          *time.Time             `json:"read_at,omitempty"`
	
	// External references
	ExternalID      string                 `json:"external_id,omitempty"`
}

// ScheduledPost represents a scheduled post
type ScheduledPost struct {
	ID              string                 `json:"id"`
	PostContent     *CreatePostRequest     `json:"post_content"`
	ScheduledAt     time.Time              `json:"scheduled_at"`
	Status          ScheduledPostStatus    `json:"status"`
	
	// Timestamps
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	PublishedAt     *time.Time             `json:"published_at,omitempty"`
	
	// Result
	PublishedPostID string                 `json:"published_post_id,omitempty"`
	Error           string                 `json:"error,omitempty"`
}

// Analytics structures
type PostAnalytics struct {
	PostID          string                 `json:"post_id"`
	Period          AnalyticsPeriod        `json:"period"`
	
	// Engagement metrics
	Impressions     int                    `json:"impressions"`
	Reach           int                    `json:"reach"`
	Likes           int                    `json:"likes"`
	Comments        int                    `json:"comments"`
	Shares          int                    `json:"shares"`
	Saves           int                    `json:"saves"`
	Clicks          int                    `json:"clicks"`
	
	// Demographic data
	Demographics    *Demographics          `json:"demographics,omitempty"`
	
	// Time series data
	TimeSeries      []*TimeSeriesPoint     `json:"time_series,omitempty"`
	
	// Generated at
	GeneratedAt     time.Time              `json:"generated_at"`
}

type UserAnalytics struct {
	UserID          string                 `json:"user_id"`
	Period          AnalyticsPeriod        `json:"period"`
	
	// Growth metrics
	FollowersGained int                    `json:"followers_gained"`
	FollowersLost   int                    `json:"followers_lost"`
	NetFollowerGrowth int                  `json:"net_follower_growth"`
	
	// Content metrics
	PostsPublished  int                    `json:"posts_published"`
	TotalImpressions int                   `json:"total_impressions"`
	TotalEngagement int                    `json:"total_engagement"`
	EngagementRate  float64                `json:"engagement_rate"`
	
	// Top content
	TopPosts        []*Post                `json:"top_posts,omitempty"`
	
	// Demographics
	Demographics    *Demographics          `json:"demographics,omitempty"`
	
	// Generated at
	GeneratedAt     time.Time              `json:"generated_at"`
}

type HashtagAnalytics struct {
	Hashtag         string                 `json:"hashtag"`
	Period          AnalyticsPeriod        `json:"period"`
	
	// Usage metrics
	PostCount       int                    `json:"post_count"`
	UniqueUsers     int                    `json:"unique_users"`
	TotalImpressions int                   `json:"total_impressions"`
	TotalEngagement int                    `json:"total_engagement"`
	
	// Trend data
	TrendScore      float64                `json:"trend_score"`
	TrendDirection  TrendDirection         `json:"trend_direction"`
	
	// Related hashtags
	RelatedHashtags []string               `json:"related_hashtags,omitempty"`
	
	// Top posts
	TopPosts        []*Post                `json:"top_posts,omitempty"`
	
	// Generated at
	GeneratedAt     time.Time              `json:"generated_at"`
}

// Request types
type CreatePostRequest struct {
	Text            string                 `json:"text"`
	MediaIDs        []string               `json:"media_ids,omitempty"`
	Location        *Location              `json:"location,omitempty"`
	InReplyToID     string                 `json:"in_reply_to_id,omitempty"`
	IsSensitive     bool                   `json:"is_sensitive,omitempty"`
	ScheduledAt     *time.Time             `json:"scheduled_at,omitempty"`
}

type UpdatePostRequest struct {
	Text            *string                `json:"text,omitempty"`
	IsSensitive     *bool                  `json:"is_sensitive,omitempty"`
}

type UploadMediaRequest struct {
	Content         []byte                 `json:"content"`
	Filename        string                 `json:"filename"`
	MimeType        string                 `json:"mime_type"`
	AltText         string                 `json:"alt_text,omitempty"`
	Caption         string                 `json:"caption,omitempty"`
}

type SocialCommentRequest struct {
	Text            string                 `json:"text"`
}

type SocialShareRequest struct {
	Text            string                 `json:"text,omitempty"`
}

type TimelineRequest struct {
	Count           int                    `json:"count,omitempty"`
	SinceID         string                 `json:"since_id,omitempty"`
	MaxID           string                 `json:"max_id,omitempty"`
	IncludeReplies  bool                   `json:"include_replies,omitempty"`
	IncludeRetweets bool                   `json:"include_retweets,omitempty"`
}

type FeedRequest struct {
	Count           int                    `json:"count,omitempty"`
	SinceID         string                 `json:"since_id,omitempty"`
	MaxID           string                 `json:"max_id,omitempty"`
}

type TrendingRequest struct {
	Count           int                    `json:"count,omitempty"`
	Period          TrendingPeriod         `json:"period,omitempty"`
	Location        string                 `json:"location,omitempty"`
}

type SocialSearchRequest struct {
	Query           string                 `json:"query"`
	Count           int                    `json:"count,omitempty"`
	SinceID         string                 `json:"since_id,omitempty"`
	MaxID           string                 `json:"max_id,omitempty"`
	ResultType      SearchResultType       `json:"result_type,omitempty"`
	Language        string                 `json:"language,omitempty"`
	Geocode         string                 `json:"geocode,omitempty"`
	Since           *time.Time             `json:"since,omitempty"`
	Until           *time.Time             `json:"until,omitempty"`
}

type UserSearchRequest struct {
	Query           string                 `json:"query"`
	Count           int                    `json:"count,omitempty"`
	Page            int                    `json:"page,omitempty"`
}

type HashtagSearchRequest struct {
	Query           string                 `json:"query"`
	Count           int                    `json:"count,omitempty"`
	TrendingOnly    bool                   `json:"trending_only,omitempty"`
}

type AnalyticsRequest struct {
	Period          AnalyticsPeriod        `json:"period"`
	StartDate       *time.Time             `json:"start_date,omitempty"`
	EndDate         *time.Time             `json:"end_date,omitempty"`
	Metrics         []string               `json:"metrics,omitempty"`
	IncludeDemographics bool               `json:"include_demographics,omitempty"`
	IncludeTimeSeries bool                 `json:"include_time_series,omitempty"`
}

type HashtagPostsRequest struct {
	Count           int                    `json:"count,omitempty"`
	SinceID         string                 `json:"since_id,omitempty"`
	MaxID           string                 `json:"max_id,omitempty"`
	ResultType      SearchResultType       `json:"result_type,omitempty"`
}

type TrendingHashtagsRequest struct {
	Count           int                    `json:"count,omitempty"`
	Location        string                 `json:"location,omitempty"`
}

type DirectMessageRequest struct {
	RecipientID     string                 `json:"recipient_id"`
	Text            string                 `json:"text"`
	MediaIDs        []string               `json:"media_ids,omitempty"`
}

type DirectMessageListRequest struct {
	Count           int                    `json:"count,omitempty"`
	SinceID         string                 `json:"since_id,omitempty"`
	MaxID           string                 `json:"max_id,omitempty"`
}

type SchedulePostRequest struct {
	PostContent     *CreatePostRequest     `json:"post_content"`
	ScheduledAt     time.Time              `json:"scheduled_at"`
}

type UpdateScheduledPostRequest struct {
	PostContent     *CreatePostRequest     `json:"post_content,omitempty"`
	ScheduledAt     *time.Time             `json:"scheduled_at,omitempty"`
}

type ScheduledPostsRequest struct {
	Status          []ScheduledPostStatus  `json:"status,omitempty"`
	Count           int                    `json:"count,omitempty"`
	Page            int                    `json:"page,omitempty"`
}

type PaginationRequest struct {
	Count           int                    `json:"count,omitempty"`
	Cursor          string                 `json:"cursor,omitempty"`
}

type SocialWebhookRequest struct {
	URL             string                 `json:"url"`
	Events          []string               `json:"events"`
	Secret          string                 `json:"secret,omitempty"`
}

// Response types
type PostList struct {
	Posts           []*Post                `json:"posts"`
	NextCursor      string                 `json:"next_cursor,omitempty"`
	HasMore         bool                   `json:"has_more"`
}

type SocialUserList struct {
	Users           []*User                `json:"users"`
	NextCursor      string                 `json:"next_cursor,omitempty"`
	HasMore         bool                   `json:"has_more"`
}

type PostSearchResult struct {
	Posts           []*Post                `json:"posts"`
	Total           int                    `json:"total"`
	NextCursor      string                 `json:"next_cursor,omitempty"`
	HasMore         bool                   `json:"has_more"`
}

type UserSearchResult struct {
	Users           []*User                `json:"users"`
	Total           int                    `json:"total"`
	NextCursor      string                 `json:"next_cursor,omitempty"`
	HasMore         bool                   `json:"has_more"`
}

type HashtagSearchResult struct {
	Hashtags        []*Hashtag             `json:"hashtags"`
	Total           int                    `json:"total"`
	NextCursor      string                 `json:"next_cursor,omitempty"`
	HasMore         bool                   `json:"has_more"`
}

type HashtagList struct {
	Hashtags        []*Hashtag             `json:"hashtags"`
	NextCursor      string                 `json:"next_cursor,omitempty"`
	HasMore         bool                   `json:"has_more"`
}

type DirectMessageList struct {
	Messages        []*DirectMessage       `json:"messages"`
	NextCursor      string                 `json:"next_cursor,omitempty"`
	HasMore         bool                   `json:"has_more"`
}

type ScheduledPostList struct {
	ScheduledPosts  []*ScheduledPost       `json:"scheduled_posts"`
	Total           int                    `json:"total"`
	Page            int                    `json:"page"`
	HasMore         bool                   `json:"has_more"`
}

// Supporting types
type Location struct {
	Name            string                 `json:"name"`
	Latitude        float64                `json:"latitude,omitempty"`
	Longitude       float64                `json:"longitude,omitempty"`
	PlaceID         string                 `json:"place_id,omitempty"`
}

type Hashtag struct {
	Name            string                 `json:"name"`
	PostCount       int                    `json:"post_count"`
	TrendScore      float64                `json:"trend_score,omitempty"`
	TrendDirection  TrendDirection         `json:"trend_direction,omitempty"`
}

type Demographics struct {
	AgeGroups       map[string]int         `json:"age_groups,omitempty"`
	Genders         map[string]int         `json:"genders,omitempty"`
	Locations       map[string]int         `json:"locations,omitempty"`
	Languages       map[string]int         `json:"languages,omitempty"`
	Interests       map[string]int         `json:"interests,omitempty"`
}

type TimeSeriesPoint struct {
	Timestamp       time.Time              `json:"timestamp"`
	Value           int                    `json:"value"`
	Metric          string                 `json:"metric"`
}

type SocialWebhook struct {
	ID              string                 `json:"id"`
	URL             string                 `json:"url"`
	Events          []string               `json:"events"`
	Secret          string                 `json:"secret,omitempty"`
	Active          bool                   `json:"active"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// Enums
type PostType string

const (
	PostTypeText    PostType = "text"
	PostTypeImage   PostType = "image"
	PostTypeVideo   PostType = "video"
	PostTypeLink    PostType = "link"
	PostTypePoll    PostType = "poll"
	PostTypeStory   PostType = "story"
)

type MediaType string

const (
	MediaTypeImage  MediaType = "image"
	MediaTypeVideo  MediaType = "video"
	MediaTypeGIF    MediaType = "gif"
	MediaTypeAudio  MediaType = "audio"
)

type AccountType string

const (
	AccountTypePersonal AccountType = "personal"
	AccountTypeBusiness AccountType = "business"
	AccountTypeCreator  AccountType = "creator"
)

type ScheduledPostStatus string

const (
	ScheduledPostStatusPending   ScheduledPostStatus = "pending"
	ScheduledPostStatusPublished ScheduledPostStatus = "published"
	ScheduledPostStatusFailed    ScheduledPostStatus = "failed"
	ScheduledPostStatusCancelled ScheduledPostStatus = "cancelled"
)

type AnalyticsPeriod string

const (
	AnalyticsPeriodDay   AnalyticsPeriod = "day"
	AnalyticsPeriodWeek  AnalyticsPeriod = "week"
	AnalyticsPeriodMonth AnalyticsPeriod = "month"
	AnalyticsPeriodYear  AnalyticsPeriod = "year"
)

type TrendDirection string

const (
	TrendDirectionUp    TrendDirection = "up"
	TrendDirectionDown  TrendDirection = "down"
	TrendDirectionFlat  TrendDirection = "flat"
)

type TrendingPeriod string

const (
	TrendingPeriodHour  TrendingPeriod = "hour"
	TrendingPeriodDay   TrendingPeriod = "day"
	TrendingPeriodWeek  TrendingPeriod = "week"
)

type SearchResultType string

const (
	SearchResultTypeMixed   SearchResultType = "mixed"
	SearchResultTypeRecent  SearchResultType = "recent"
	SearchResultTypePopular SearchResultType = "popular"
)
