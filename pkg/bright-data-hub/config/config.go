package config

import (
	"os"
	"strconv"
	"time"
)

// BrightDataHubConfig contains all configuration for the Bright Data Hub
type BrightDataHubConfig struct {
	// Core settings
	Enabled         bool          `json:"enabled"`
	MCPServerURL    string        `json:"mcp_server_url"`
	MaxConcurrent   int           `json:"max_concurrent"`
	RequestTimeout  time.Duration `json:"request_timeout"`
	
	// Rate limiting
	RateLimitRPS    int           `json:"rate_limit_rps"`
	RateLimitBurst  int           `json:"rate_limit_burst"`
	
	// Caching
	CacheTTL        time.Duration `json:"cache_ttl"`
	CacheMaxSize    int           `json:"cache_max_size"`
	RedisURL        string        `json:"redis_url"`
	
	// Features
	EnableSocial    bool          `json:"enable_social"`
	EnableEcommerce bool          `json:"enable_ecommerce"`
	EnableSearch    bool          `json:"enable_search"`
	EnableAnalytics bool          `json:"enable_analytics"`
	
	// AI Analytics
	SentimentEnabled    bool    `json:"sentiment_enabled"`
	TrendDetectionEnabled bool  `json:"trend_detection_enabled"`
	ConfidenceThreshold float64 `json:"confidence_threshold"`
	
	// Monitoring
	MetricsEnabled  bool   `json:"metrics_enabled"`
	TracingEnabled  bool   `json:"tracing_enabled"`
	LogLevel        string `json:"log_level"`
	
	// Security
	APIKeyRequired  bool   `json:"api_key_required"`
	MaxRequestSize  int64  `json:"max_request_size"`
	
	// Platform-specific settings
	Social    SocialConfig    `json:"social"`
	Ecommerce EcommerceConfig `json:"ecommerce"`
	Search    SearchConfig    `json:"search"`
}

type SocialConfig struct {
	Instagram InstagramConfig `json:"instagram"`
	Facebook  FacebookConfig  `json:"facebook"`
	Twitter   TwitterConfig   `json:"twitter"`
	LinkedIn  LinkedInConfig  `json:"linkedin"`
}

type InstagramConfig struct {
	Enabled           bool `json:"enabled"`
	MaxPostsPerQuery  int  `json:"max_posts_per_query"`
	MaxCommentsPerPost int `json:"max_comments_per_post"`
}

type FacebookConfig struct {
	Enabled              bool `json:"enabled"`
	MaxPostsPerQuery     int  `json:"max_posts_per_query"`
	MaxReviewsPerCompany int  `json:"max_reviews_per_company"`
}

type TwitterConfig struct {
	Enabled          bool `json:"enabled"`
	MaxPostsPerQuery int  `json:"max_posts_per_query"`
}

type LinkedInConfig struct {
	Enabled bool `json:"enabled"`
}

type EcommerceConfig struct {
	Amazon  AmazonConfig  `json:"amazon"`
	Booking BookingConfig `json:"booking"`
	Zillow  ZillowConfig  `json:"zillow"`
}

type AmazonConfig struct {
	Enabled           bool `json:"enabled"`
	MaxReviewsPerProduct int `json:"max_reviews_per_product"`
}

type BookingConfig struct {
	Enabled bool `json:"enabled"`
}

type ZillowConfig struct {
	Enabled bool `json:"enabled"`
}

type SearchConfig struct {
	DefaultEngine    string   `json:"default_engine"`
	EnabledEngines   []string `json:"enabled_engines"`
	MaxResultsPerQuery int    `json:"max_results_per_query"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *BrightDataHubConfig {
	config := &BrightDataHubConfig{
		// Core settings
		Enabled:        getEnvBool("BRIGHT_DATA_HUB_ENABLED", true),
		MCPServerURL:   getEnvString("MCP_SERVER_URL", "http://localhost:3001"),
		MaxConcurrent:  getEnvInt("BRIGHT_DATA_MAX_CONCURRENT", 10),
		RequestTimeout: getEnvDuration("BRIGHT_DATA_REQUEST_TIMEOUT", 30*time.Second),
		
		// Rate limiting
		RateLimitRPS:   getEnvInt("BRIGHT_DATA_RATE_LIMIT_RPS", 10),
		RateLimitBurst: getEnvInt("BRIGHT_DATA_RATE_LIMIT_BURST", 20),
		
		// Caching
		CacheTTL:     getEnvDuration("BRIGHT_DATA_CACHE_TTL", 5*time.Minute),
		CacheMaxSize: getEnvInt("BRIGHT_DATA_CACHE_MAX_SIZE", 1000),
		RedisURL:     getEnvString("REDIS_URL", "redis://localhost:6379"),
		
		// Features
		EnableSocial:    getEnvBool("BRIGHT_DATA_ENABLE_SOCIAL", true),
		EnableEcommerce: getEnvBool("BRIGHT_DATA_ENABLE_ECOMMERCE", true),
		EnableSearch:    getEnvBool("BRIGHT_DATA_ENABLE_SEARCH", true),
		EnableAnalytics: getEnvBool("BRIGHT_DATA_ENABLE_ANALYTICS", true),
		
		// AI Analytics
		SentimentEnabled:      getEnvBool("BRIGHT_DATA_SENTIMENT_ENABLED", true),
		TrendDetectionEnabled: getEnvBool("BRIGHT_DATA_TREND_DETECTION_ENABLED", true),
		ConfidenceThreshold:   getEnvFloat("BRIGHT_DATA_CONFIDENCE_THRESHOLD", 0.7),
		
		// Monitoring
		MetricsEnabled: getEnvBool("BRIGHT_DATA_METRICS_ENABLED", true),
		TracingEnabled: getEnvBool("BRIGHT_DATA_TRACING_ENABLED", true),
		LogLevel:       getEnvString("BRIGHT_DATA_LOG_LEVEL", "info"),
		
		// Security
		APIKeyRequired: getEnvBool("BRIGHT_DATA_API_KEY_REQUIRED", false),
		MaxRequestSize: getEnvInt64("BRIGHT_DATA_MAX_REQUEST_SIZE", 10*1024*1024), // 10MB
		
		// Platform-specific settings
		Social: SocialConfig{
			Instagram: InstagramConfig{
				Enabled:            getEnvBool("BRIGHT_DATA_INSTAGRAM_ENABLED", true),
				MaxPostsPerQuery:   getEnvInt("BRIGHT_DATA_INSTAGRAM_MAX_POSTS", 50),
				MaxCommentsPerPost: getEnvInt("BRIGHT_DATA_INSTAGRAM_MAX_COMMENTS", 100),
			},
			Facebook: FacebookConfig{
				Enabled:              getEnvBool("BRIGHT_DATA_FACEBOOK_ENABLED", true),
				MaxPostsPerQuery:     getEnvInt("BRIGHT_DATA_FACEBOOK_MAX_POSTS", 50),
				MaxReviewsPerCompany: getEnvInt("BRIGHT_DATA_FACEBOOK_MAX_REVIEWS", 100),
			},
			Twitter: TwitterConfig{
				Enabled:          getEnvBool("BRIGHT_DATA_TWITTER_ENABLED", true),
				MaxPostsPerQuery: getEnvInt("BRIGHT_DATA_TWITTER_MAX_POSTS", 50),
			},
			LinkedIn: LinkedInConfig{
				Enabled: getEnvBool("BRIGHT_DATA_LINKEDIN_ENABLED", true),
			},
		},
		Ecommerce: EcommerceConfig{
			Amazon: AmazonConfig{
				Enabled:              getEnvBool("BRIGHT_DATA_AMAZON_ENABLED", true),
				MaxReviewsPerProduct: getEnvInt("BRIGHT_DATA_AMAZON_MAX_REVIEWS", 100),
			},
			Booking: BookingConfig{
				Enabled: getEnvBool("BRIGHT_DATA_BOOKING_ENABLED", true),
			},
			Zillow: ZillowConfig{
				Enabled: getEnvBool("BRIGHT_DATA_ZILLOW_ENABLED", true),
			},
		},
		Search: SearchConfig{
			DefaultEngine:      getEnvString("BRIGHT_DATA_DEFAULT_SEARCH_ENGINE", "google"),
			EnabledEngines:     []string{"google", "bing", "yandex"},
			MaxResultsPerQuery: getEnvInt("BRIGHT_DATA_MAX_SEARCH_RESULTS", 20),
		},
	}
	
	return config
}

// Helper functions for environment variables
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
