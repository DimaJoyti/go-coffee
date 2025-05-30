package ai

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"
)

// GenerateRequest represents a request to generate AI response
type GenerateRequest struct {
	UserID      string            `json:"user_id"`
	Message     string            `json:"message"`
	Context     string            `json:"context,omitempty"`
	Temperature float64           `json:"temperature,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// MessageHash generates a hash of the message for caching
func (r *GenerateRequest) MessageHash() string {
	data := fmt.Sprintf("%s:%s:%s:%.2f", r.UserID, r.Message, r.Context, r.Temperature)
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

// GenerateResponse represents a response from AI generation
type GenerateResponse struct {
	Text        string            `json:"text"`
	Provider    string            `json:"provider"`
	Confidence  float64           `json:"confidence,omitempty"`
	Suggestions []string          `json:"suggestions,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Cached      bool              `json:"cached,omitempty"`
	GeneratedAt time.Time         `json:"generated_at"`
}

// CoffeeOrderRequest represents a coffee order processing request
type CoffeeOrderRequest struct {
	UserID    string `json:"user_id"`
	Message   string `json:"message"`
	ChatID    int64  `json:"chat_id"`
	MessageID int    `json:"message_id"`
}

// CoffeeOrderResponse represents a coffee order processing response
type CoffeeOrderResponse struct {
	AIResponse  string    `json:"ai_response"`
	Provider    string    `json:"provider"`
	ProcessedAt time.Time `json:"processed_at"`
	Confidence  float64   `json:"confidence,omitempty"`
	Suggestions []string  `json:"suggestions,omitempty"`

	// Parsed order details (if successfully extracted)
	OrderDetails *ParsedCoffeeOrder `json:"order_details,omitempty"`
}

// ParsedCoffeeOrder represents parsed coffee order details
type ParsedCoffeeOrder struct {
	CoffeeType        string   `json:"coffee_type"`
	Size              string   `json:"size"`
	Extras            []string `json:"extras"`
	Quantity          int      `json:"quantity"`
	EstimatedPriceUSD float64  `json:"estimated_price_usd"`
	PaymentSuggestion string   `json:"payment_suggestion"`
}

// PaymentQueryRequest represents a payment query request
type PaymentQueryRequest struct {
	UserID          string            `json:"user_id"`
	Message         string            `json:"message"`
	WalletBalance   string            `json:"wallet_balance,omitempty"`
	AvailableTokens []string          `json:"available_tokens,omitempty"`
	CurrentPrices   map[string]string `json:"current_prices,omitempty"`
	ChatID          int64             `json:"chat_id"`
	MessageID       int               `json:"message_id"`
}

// PaymentQueryResponse represents a payment query response
type PaymentQueryResponse struct {
	AIResponse  string    `json:"ai_response"`
	Provider    string    `json:"provider"`
	ProcessedAt time.Time `json:"processed_at"`
	Confidence  float64   `json:"confidence,omitempty"`

	// Structured payment recommendations
	Recommendations *PaymentRecommendations `json:"recommendations,omitempty"`
}

// PaymentRecommendations represents structured payment recommendations
type PaymentRecommendations struct {
	RecommendedToken string            `json:"recommended_token"`
	EstimatedFees    map[string]string `json:"estimated_fees"`
	Alternatives     []string          `json:"alternatives"`
	SecurityTips     []string          `json:"security_tips"`
}

// WalletQueryRequest represents a wallet-related query request
type WalletQueryRequest struct {
	UserID    string `json:"user_id"`
	Message   string `json:"message"`
	ChatID    int64  `json:"chat_id"`
	MessageID int    `json:"message_id"`
}

// WalletQueryResponse represents a wallet query response
type WalletQueryResponse struct {
	AIResponse  string    `json:"ai_response"`
	Provider    string    `json:"provider"`
	ProcessedAt time.Time `json:"processed_at"`
	Confidence  float64   `json:"confidence,omitempty"`
}

// MenuQueryRequest represents a menu query request
type MenuQueryRequest struct {
	UserID    string `json:"user_id"`
	Message   string `json:"message"`
	ChatID    int64  `json:"chat_id"`
	MessageID int    `json:"message_id"`
}

// MenuQueryResponse represents a menu query response
type MenuQueryResponse struct {
	AIResponse  string    `json:"ai_response"`
	Provider    string    `json:"provider"`
	ProcessedAt time.Time `json:"processed_at"`
	Confidence  float64   `json:"confidence,omitempty"`

	// Structured menu information
	MenuItems []MenuItem `json:"menu_items,omitempty"`
}

// MenuItem represents a coffee menu item
type MenuItem struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	PriceUSD    float64 `json:"price_usd"`
	Available   bool    `json:"available"`
	Category    string  `json:"category"`
}

// GeneralQueryRequest represents a general query request
type GeneralQueryRequest struct {
	UserID    string `json:"user_id"`
	Message   string `json:"message"`
	Context   string `json:"context,omitempty"`
	ChatID    int64  `json:"chat_id"`
	MessageID int    `json:"message_id"`
}

// GeneralQueryResponse represents a general query response
type GeneralQueryResponse struct {
	AIResponse  string    `json:"ai_response"`
	Provider    string    `json:"provider"`
	ProcessedAt time.Time `json:"processed_at"`
	Confidence  float64   `json:"confidence,omitempty"`
	Suggestions []string  `json:"suggestions,omitempty"`
}

// ConversationContext represents conversation context for better AI responses
type ConversationContext struct {
	UserID       string             `json:"user_id"`
	ChatID       int64              `json:"chat_id"`
	LastMessages []Message          `json:"last_messages"`
	UserProfile  UserProfile        `json:"user_profile"`
	CurrentOrder *ParsedCoffeeOrder `json:"current_order,omitempty"`
	Preferences  map[string]string  `json:"preferences"`
	SessionStart time.Time          `json:"session_start"`
	LastActivity time.Time          `json:"last_activity"`
}

// Message represents a conversation message
type Message struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	IsUser    bool      `json:"is_user"`
	Timestamp time.Time `json:"timestamp"`
	Context   string    `json:"context,omitempty"`
}

// UserProfile represents user profile information
type UserProfile struct {
	UserID            string   `json:"user_id"`
	PreferredLanguage string   `json:"preferred_language"`
	FavoriteCoffees   []string `json:"favorite_coffees"`
	PaymentMethods    []string `json:"payment_methods"`
	OrderHistory      int      `json:"order_history"`
	LoyaltyLevel      string   `json:"loyalty_level"`
}

// AIProviderInterface defines the interface for AI providers
type AIProviderInterface interface {
	GenerateResponse(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
	IsHealthy(ctx context.Context) bool
	Close() error
}

// ProviderConfig represents configuration for an AI provider
type ProviderConfig struct {
	Name      string            `json:"name"`
	Enabled   bool              `json:"enabled"`
	Priority  int               `json:"priority"`
	Config    map[string]string `json:"config"`
	RateLimit RateLimit         `json:"rate_limit"`
	Timeout   time.Duration     `json:"timeout"`
}

// RateLimit represents rate limiting configuration
type RateLimit struct {
	RequestsPerMinute int           `json:"requests_per_minute"`
	BurstSize         int           `json:"burst_size"`
	WindowSize        time.Duration `json:"window_size"`
}

// AIMetrics represents AI service metrics
type AIMetrics struct {
	TotalRequests      int64            `json:"total_requests"`
	SuccessfulRequests int64            `json:"successful_requests"`
	FailedRequests     int64            `json:"failed_requests"`
	AverageLatency     time.Duration    `json:"average_latency"`
	ProviderStats      map[string]int64 `json:"provider_stats"`
	CacheHitRate       float64          `json:"cache_hit_rate"`
	LastUpdated        time.Time        `json:"last_updated"`
}

// ErrorResponse represents an error response from AI service
type ErrorResponse struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Provider  string    `json:"provider,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
}

// Error implements the error interface
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("[%s] %s (provider: %s)", e.Code, e.Message, e.Provider)
}
