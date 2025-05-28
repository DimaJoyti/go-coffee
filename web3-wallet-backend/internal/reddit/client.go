package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/config"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
)

// Client represents a Reddit API client
type Client struct {
	config      config.RedditConfig
	logger      *logger.Logger
	httpClient  *http.Client
	baseURL     string
	accessToken string
	tokenExpiry time.Time
	rateLimiter *RateLimiter
}

// RateLimiter handles Reddit API rate limiting
type RateLimiter struct {
	requests  chan struct{}
	resetTime time.Time
	remaining int
	limit     int
}

// NewClient creates a new Reddit API client
func NewClient(cfg config.RedditConfig, logger *logger.Logger) (*Client, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("reddit client is disabled")
	}

	client := &Client{
		config:  cfg,
		logger:  logger,
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: &RateLimiter{
			requests: make(chan struct{}, cfg.RateLimit.RequestsPerMinute),
			limit:    cfg.RateLimit.RequestsPerMinute,
		},
	}

	// Initialize rate limiter
	go client.rateLimiter.refillTokens(cfg.RateLimit.RequestsPerMinute)

	// Authenticate with Reddit API
	if err := client.authenticate(); err != nil {
		return nil, fmt.Errorf("failed to authenticate with Reddit API: %w", err)
	}

	logger.Info("Reddit API client initialized successfully")
	return client, nil
}

// authenticate authenticates with Reddit API using OAuth2
func (c *Client) authenticate() error {
	c.logger.Info("Authenticating with Reddit API")

	// Prepare authentication request
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", c.config.Username)
	data.Set("password", c.config.Password)

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.config.UserAgent)
	req.SetBasicAuth(c.config.ClientID, c.config.ClientSecret)

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var authResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	// Store access token
	c.accessToken = authResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)

	c.logger.Info("Successfully authenticated with Reddit API")
	return nil
}

// makeRequest makes an authenticated request to Reddit API
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, params url.Values) (*http.Response, error) {
	// Check if token needs refresh
	if time.Now().After(c.tokenExpiry.Add(-5 * time.Minute)) {
		if err := c.authenticate(); err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
	}

	// Wait for rate limit
	select {
	case <-c.rateLimiter.requests:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Build URL
	reqURL := c.baseURL + endpoint
	if params != nil && len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("User-Agent", c.config.UserAgent)
	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	// Handle rate limiting
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				c.logger.Warn(fmt.Sprintf("Rate limited, waiting %d seconds", seconds))
				time.Sleep(time.Duration(seconds) * time.Second)
				return c.makeRequest(ctx, method, endpoint, params)
			}
		}
		return nil, fmt.Errorf("rate limited")
	}

	return resp, nil
}

// GetSubredditPosts retrieves posts from a subreddit
func (c *Client) GetSubredditPosts(ctx context.Context, subreddit string, sort string, limit int, after string) ([]RedditPost, string, error) {
	c.logger.Info(fmt.Sprintf("Getting posts from r/%s", subreddit))

	// Prepare parameters
	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))
	if after != "" {
		params.Set("after", after)
	}

	// Make request
	endpoint := fmt.Sprintf("/r/%s/%s", subreddit, sort)
	resp, err := c.makeRequest(ctx, "GET", endpoint, params)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get subreddit posts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract posts
	posts, nextAfter, err := c.parsePostsFromResponse(apiResp)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse posts: %w", err)
	}

	c.logger.Info(fmt.Sprintf("Retrieved %d posts from r/%s", len(posts), subreddit))
	return posts, nextAfter, nil
}

// GetPostComments retrieves comments for a specific post
func (c *Client) GetPostComments(ctx context.Context, subreddit, postID string, sort string, limit int) ([]RedditComment, error) {
	c.logger.Info(fmt.Sprintf("Getting comments for post %s", postID))

	// Prepare parameters
	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))
	params.Set("sort", sort)

	// Make request
	endpoint := fmt.Sprintf("/r/%s/comments/%s", subreddit, postID)
	resp, err := c.makeRequest(ctx, "GET", endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get post comments: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp []APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract comments (comments are in the second element of the array)
	if len(apiResp) < 2 {
		return []RedditComment{}, nil
	}

	comments, err := c.parseCommentsFromResponse(apiResp[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse comments: %w", err)
	}

	c.logger.Info(fmt.Sprintf("Retrieved %d comments for post %s", len(comments), postID))
	return comments, nil
}

// SearchContent searches Reddit content
func (c *Client) SearchContent(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	c.logger.Info(fmt.Sprintf("Searching Reddit content: %s", req.Query))

	// Prepare parameters
	params := url.Values{}
	params.Set("q", req.Query)
	params.Set("sort", req.Sort)
	params.Set("t", req.Time)
	params.Set("type", req.Type)
	params.Set("limit", strconv.Itoa(req.Limit))
	if req.After != "" {
		params.Set("after", req.After)
	}
	if req.Before != "" {
		params.Set("before", req.Before)
	}

	// Build endpoint
	endpoint := "/search"
	if req.Subreddit != "" {
		endpoint = fmt.Sprintf("/r/%s/search", req.Subreddit)
		params.Set("restrict_sr", "true")
	}

	// Make request
	resp, err := c.makeRequest(ctx, "GET", endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to search content: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// Extract search results
	searchResp, err := c.parseSearchResponse(apiResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	c.logger.Info(fmt.Sprintf("Found %d results for query: %s", searchResp.Count, req.Query))
	return searchResp, nil
}

// Close closes the Reddit client
func (c *Client) Close() error {
	c.logger.Info("Closing Reddit API client")
	return nil
}

// parsePostsFromResponse parses posts from Reddit API response
func (c *Client) parsePostsFromResponse(apiResp APIResponse) ([]RedditPost, string, error) {
	listingData, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return nil, "", fmt.Errorf("invalid response data format")
	}

	children, ok := listingData["children"].([]interface{})
	if !ok {
		return nil, "", fmt.Errorf("invalid children format")
	}

	var posts []RedditPost
	for _, child := range children {
		childMap, ok := child.(map[string]interface{})
		if !ok {
			continue
		}

		postData, ok := childMap["data"].(map[string]interface{})
		if !ok {
			continue
		}

		post := c.parsePostData(postData)
		posts = append(posts, post)
	}

	// Get pagination token
	after, _ := listingData["after"].(string)

	return posts, after, nil
}

// parseCommentsFromResponse parses comments from Reddit API response
func (c *Client) parseCommentsFromResponse(apiResp APIResponse) ([]RedditComment, error) {
	listingData, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response data format")
	}

	children, ok := listingData["children"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid children format")
	}

	var comments []RedditComment
	for _, child := range children {
		childMap, ok := child.(map[string]interface{})
		if !ok {
			continue
		}

		commentData, ok := childMap["data"].(map[string]interface{})
		if !ok {
			continue
		}

		comment := c.parseCommentData(commentData)
		comments = append(comments, comment)

		// Parse nested replies
		if replies, ok := commentData["replies"].(map[string]interface{}); ok {
			nestedComments, _ := c.parseCommentsFromResponse(APIResponse{Data: replies})
			comments = append(comments, nestedComments...)
		}
	}

	return comments, nil
}

// parseSearchResponse parses search results from Reddit API response
func (c *Client) parseSearchResponse(apiResp APIResponse) (*SearchResponse, error) {
	listingData, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response data format")
	}

	children, ok := listingData["children"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid children format")
	}

	searchResp := &SearchResponse{
		Posts:    []RedditPost{},
		Comments: []RedditComment{},
	}

	for _, child := range children {
		childMap, ok := child.(map[string]interface{})
		if !ok {
			continue
		}

		kind, _ := childMap["kind"].(string)
		data, ok := childMap["data"].(map[string]interface{})
		if !ok {
			continue
		}

		switch kind {
		case "t3": // Post
			post := c.parsePostData(data)
			searchResp.Posts = append(searchResp.Posts, post)
		case "t1": // Comment
			comment := c.parseCommentData(data)
			searchResp.Comments = append(searchResp.Comments, comment)
		}
	}

	// Set pagination and count
	searchResp.After, _ = listingData["after"].(string)
	searchResp.Before, _ = listingData["before"].(string)
	searchResp.Count = len(searchResp.Posts) + len(searchResp.Comments)
	searchResp.HasMore = searchResp.After != ""

	return searchResp, nil
}

// parsePostData parses individual post data
func (c *Client) parsePostData(data map[string]interface{}) RedditPost {
	post := RedditPost{
		ID:          getString(data, "id"),
		Title:       getString(data, "title"),
		Content:     getString(data, "selftext"),
		Author:      getString(data, "author"),
		Subreddit:   getString(data, "subreddit"),
		URL:         getString(data, "url"),
		Score:       getInt(data, "score"),
		UpvoteRatio: getFloat64(data, "upvote_ratio"),
		NumComments: getInt(data, "num_comments"),
		IsVideo:     getBool(data, "is_video"),
		IsSelf:      getBool(data, "is_self"),
		Permalink:   getString(data, "permalink"),
		Flair:       getString(data, "link_flair_text"),
		NSFW:        getBool(data, "over_18"),
		Spoiler:     getBool(data, "spoiler"),
		Locked:      getBool(data, "locked"),
		Stickied:    getBool(data, "stickied"),
		ProcessedAt: time.Now(),
	}

	// Parse created time
	if createdUTC := getFloat64(data, "created_utc"); createdUTC > 0 {
		post.CreatedUTC = time.Unix(int64(createdUTC), 0)
	}

	return post
}

// parseCommentData parses individual comment data
func (c *Client) parseCommentData(data map[string]interface{}) RedditComment {
	comment := RedditComment{
		ID:          getString(data, "id"),
		ParentID:    getString(data, "parent_id"),
		Content:     getString(data, "body"),
		Author:      getString(data, "author"),
		Score:       getInt(data, "score"),
		IsSubmitter: getBool(data, "is_submitter"),
		Depth:       getInt(data, "depth"),
		Permalink:   getString(data, "permalink"),
		ProcessedAt: time.Now(),
	}

	// Parse created time
	if createdUTC := getFloat64(data, "created_utc"); createdUTC > 0 {
		comment.CreatedUTC = time.Unix(int64(createdUTC), 0)
	}

	return comment
}

// Helper functions for type conversion
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getInt(data map[string]interface{}, key string) int {
	if val, ok := data[key].(float64); ok {
		return int(val)
	}
	return 0
}

func getFloat64(data map[string]interface{}, key string) float64 {
	if val, ok := data[key].(float64); ok {
		return val
	}
	return 0
}

func getBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key].(bool); ok {
		return val
	}
	return false
}

// refillTokens refills rate limiter tokens
func (rl *RateLimiter) refillTokens(requestsPerMinute int) {
	ticker := time.NewTicker(time.Minute / time.Duration(requestsPerMinute))
	defer ticker.Stop()

	for range ticker.C {
		select {
		case rl.requests <- struct{}{}:
		default:
			// Channel is full, skip
		}
	}
}
