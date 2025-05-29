package redismcp

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/internal/ai"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
)

// QueryParser parses natural language queries into Redis commands
type QueryParser struct {
	aiService *ai.Service
	logger    *logger.Logger
	patterns  map[string]*QueryPattern
}

// QueryPattern represents a pattern for matching queries
type QueryPattern struct {
	Regex       *regexp.Regexp
	QueryType   QueryType
	Operation   string
	Confidence  float64
	RedisCmd    string
	Transformer func(matches []string, context map[string]interface{}) (*ParsedQuery, error)
}

// NewQueryParser creates a new query parser
func NewQueryParser(aiService *ai.Service, logger *logger.Logger) *QueryParser {
	parser := &QueryParser{
		aiService: aiService,
		logger:    logger,
		patterns:  make(map[string]*QueryPattern),
	}

	parser.initializePatterns()
	return parser
}

// initializePatterns sets up the query patterns for different operations
func (p *QueryParser) initializePatterns() {
	// Coffee shop specific patterns
	p.patterns["get_menu"] = &QueryPattern{
		Regex:       regexp.MustCompile(`(?i)(?:get|show|list)\s+(?:the\s+)?menu\s+(?:for\s+)?(?:shop\s+)?(.+)`),
		QueryType:   QueryTypeRead,
		Operation:   "HGETALL",
		Confidence:  0.9,
		RedisCmd:    "HGETALL",
		Transformer: p.transformMenuQuery,
	}

	p.patterns["get_inventory"] = &QueryPattern{
		Regex:       regexp.MustCompile(`(?i)(?:get|show|check)\s+inventory\s+(?:for\s+)?(.+)`),
		QueryType:   QueryTypeRead,
		Operation:   "HGETALL",
		Confidence:  0.9,
		RedisCmd:    "HGETALL",
		Transformer: p.transformInventoryQuery,
	}

	p.patterns["add_ingredient"] = &QueryPattern{
		Regex:       regexp.MustCompile(`(?i)add\s+(.+)\s+to\s+(?:the\s+)?ingredients?`),
		QueryType:   QueryTypeWrite,
		Operation:   "SADD",
		Confidence:  0.85,
		RedisCmd:    "SADD",
		Transformer: p.transformAddIngredientQuery,
	}

	p.patterns["get_orders"] = &QueryPattern{
		Regex:       regexp.MustCompile(`(?i)(?:get|show|list)\s+(?:top\s+(\d+)\s+)?(?:coffee\s+)?orders?\s+(?:for\s+)?(?:today|(\w+))?`),
		QueryType:   QueryTypeRead,
		Operation:   "ZREVRANGE",
		Confidence:  0.9,
		RedisCmd:    "ZREVRANGE",
		Transformer: p.transformOrdersQuery,
	}

	p.patterns["set_customer_data"] = &QueryPattern{
		Regex:       regexp.MustCompile(`(?i)(?:set|update)\s+customer\s+(\w+)\s+(.+)\s+(?:to|=)\s+(.+)`),
		QueryType:   QueryTypeWrite,
		Operation:   "HSET",
		Confidence:  0.8,
		RedisCmd:    "HSET",
		Transformer: p.transformCustomerDataQuery,
	}

	p.patterns["get_customer_data"] = &QueryPattern{
		Regex:       regexp.MustCompile(`(?i)(?:get|show)\s+customer\s+(\w+)\s+(.+)`),
		QueryType:   QueryTypeRead,
		Operation:   "HGET",
		Confidence:  0.85,
		RedisCmd:    "HGET",
		Transformer: p.transformGetCustomerDataQuery,
	}

	p.patterns["search_products"] = &QueryPattern{
		Regex:       regexp.MustCompile(`(?i)(?:search|find)\s+(?:products?\s+)?(?:with|containing)\s+(.+)`),
		QueryType:   QueryTypeSearch,
		Operation:   "SCAN",
		Confidence:  0.75,
		RedisCmd:    "SCAN",
		Transformer: p.transformSearchQuery,
	}

	p.patterns["get_analytics"] = &QueryPattern{
		Regex:       regexp.MustCompile(`(?i)(?:get|show)\s+analytics?\s+(?:for\s+)?(.+)`),
		QueryType:   QueryTypeRead,
		Operation:   "ZRANGE",
		Confidence:  0.8,
		RedisCmd:    "ZRANGE",
		Transformer: p.transformAnalyticsQuery,
	}
}

// Parse parses a natural language query into a Redis command
func (p *QueryParser) Parse(ctx context.Context, query string, context map[string]interface{}) (*ParsedQuery, error) {
	p.logger.Info("Parsing query", zap.String("query", query))

	// First try pattern matching
	for name, pattern := range p.patterns {
		if matches := pattern.Regex.FindStringSubmatch(query); matches != nil {
			p.logger.Info("Pattern matched", zap.String("pattern", name))
			return pattern.Transformer(matches, context)
		}
	}

	// If no pattern matches, use AI to parse the query
	return p.parseWithAI(ctx, query, context)
}

// parseWithAI uses AI to parse complex queries
func (p *QueryParser) parseWithAI(ctx context.Context, query string, context map[string]interface{}) (*ParsedQuery, error) {
	prompt := p.buildAIPrompt(query, context)
	
	response, err := p.aiService.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI parsing failed: %w", err)
	}

	var parsedQuery ParsedQuery
	if err := json.Unmarshal([]byte(response), &parsedQuery); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &parsedQuery, nil
}

// buildAIPrompt creates a prompt for AI query parsing
func (p *QueryParser) buildAIPrompt(query string, context map[string]interface{}) string {
	return fmt.Sprintf(`
Parse the following natural language query into a Redis command for a coffee shop system:

Query: "%s"
Context: %v

Available Redis data structures:
- coffee:menu:{shop_id} (hash) - coffee menu items
- coffee:inventory:{shop_id} (hash) - ingredient inventory
- coffee:orders:{date} (sorted set) - daily orders by popularity
- customer:{id} (hash) - customer profile data
- ingredients:available (set) - available ingredients
- feedback:{shop_id} (stream) - customer feedback stream

Return a JSON object with:
{
  "type": "read|write|delete|stream|search",
  "operation": "Redis operation name",
  "key": "Redis key",
  "value": "value if applicable",
  "fields": {"field": "value"} if hash operation,
  "redis_cmd": ["COMMAND", "key", "args..."],
  "confidence": 0.0-1.0
}

Example:
Query: "get menu for shop downtown"
Response: {
  "type": "read",
  "operation": "HGETALL", 
  "key": "coffee:menu:downtown",
  "redis_cmd": ["HGETALL", "coffee:menu:downtown"],
  "confidence": 0.9
}
`, query, context)
}

// Transformer functions for different query types

func (p *QueryParser) transformMenuQuery(matches []string, context map[string]interface{}) (*ParsedQuery, error) {
	shopID := strings.TrimSpace(matches[1])
	if shopID == "" {
		shopID = "default"
	}

	return &ParsedQuery{
		Type:       QueryTypeRead,
		Operation:  "HGETALL",
		Key:        fmt.Sprintf("coffee:menu:%s", shopID),
		RedisCmd:   []interface{}{"HGETALL", fmt.Sprintf("coffee:menu:%s", shopID)},
		Confidence: 0.9,
	}, nil
}

func (p *QueryParser) transformInventoryQuery(matches []string, context map[string]interface{}) (*ParsedQuery, error) {
	shopID := strings.TrimSpace(matches[1])
	if shopID == "" {
		shopID = "default"
	}

	return &ParsedQuery{
		Type:       QueryTypeRead,
		Operation:  "HGETALL",
		Key:        fmt.Sprintf("coffee:inventory:%s", shopID),
		RedisCmd:   []interface{}{"HGETALL", fmt.Sprintf("coffee:inventory:%s", shopID)},
		Confidence: 0.9,
	}, nil
}

func (p *QueryParser) transformAddIngredientQuery(matches []string, context map[string]interface{}) (*ParsedQuery, error) {
	ingredient := strings.TrimSpace(matches[1])

	return &ParsedQuery{
		Type:       QueryTypeWrite,
		Operation:  "SADD",
		Key:        "ingredients:available",
		Value:      ingredient,
		RedisCmd:   []interface{}{"SADD", "ingredients:available", ingredient},
		Confidence: 0.85,
	}, nil
}

func (p *QueryParser) transformOrdersQuery(matches []string, context map[string]interface{}) (*ParsedQuery, error) {
	limit := 10 // default
	date := "today"

	if len(matches) > 1 && matches[1] != "" {
		if l, err := strconv.Atoi(matches[1]); err == nil {
			limit = l
		}
	}

	if len(matches) > 2 && matches[2] != "" {
		date = matches[2]
	}

	key := fmt.Sprintf("coffee:orders:%s", date)
	
	return &ParsedQuery{
		Type:       QueryTypeRead,
		Operation:  "ZREVRANGE",
		Key:        key,
		Limit:      limit,
		RedisCmd:   []interface{}{"ZREVRANGE", key, 0, limit-1, "WITHSCORES"},
		Confidence: 0.9,
	}, nil
}

func (p *QueryParser) transformCustomerDataQuery(matches []string, context map[string]interface{}) (*ParsedQuery, error) {
	customerID := strings.TrimSpace(matches[1])
	field := strings.TrimSpace(matches[2])
	value := strings.TrimSpace(matches[3])

	return &ParsedQuery{
		Type:      QueryTypeWrite,
		Operation: "HSET",
		Key:       fmt.Sprintf("customer:%s", customerID),
		Fields:    map[string]interface{}{field: value},
		RedisCmd:  []interface{}{"HSET", fmt.Sprintf("customer:%s", customerID), field, value},
		Confidence: 0.8,
	}, nil
}

func (p *QueryParser) transformGetCustomerDataQuery(matches []string, context map[string]interface{}) (*ParsedQuery, error) {
	customerID := strings.TrimSpace(matches[1])
	field := strings.TrimSpace(matches[2])

	return &ParsedQuery{
		Type:       QueryTypeRead,
		Operation:  "HGET",
		Key:        fmt.Sprintf("customer:%s", customerID),
		RedisCmd:   []interface{}{"HGET", fmt.Sprintf("customer:%s", customerID), field},
		Confidence: 0.85,
	}, nil
}

func (p *QueryParser) transformSearchQuery(matches []string, context map[string]interface{}) (*ParsedQuery, error) {
	searchTerm := strings.TrimSpace(matches[1])
	pattern := fmt.Sprintf("*%s*", searchTerm)

	return &ParsedQuery{
		Type:       QueryTypeSearch,
		Operation:  "SCAN",
		Key:        pattern,
		RedisCmd:   []interface{}{"SCAN", "0", "MATCH", pattern, "COUNT", "100"},
		Confidence: 0.75,
	}, nil
}

func (p *QueryParser) transformAnalyticsQuery(matches []string, context map[string]interface{}) (*ParsedQuery, error) {
	metric := strings.TrimSpace(matches[1])
	key := fmt.Sprintf("analytics:%s", metric)

	return &ParsedQuery{
		Type:       QueryTypeRead,
		Operation:  "ZRANGE",
		Key:        key,
		RedisCmd:   []interface{}{"ZRANGE", key, 0, -1, "WITHSCORES"},
		Confidence: 0.8,
	}, nil
}
