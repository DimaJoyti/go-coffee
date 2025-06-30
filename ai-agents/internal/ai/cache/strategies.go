package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go-coffee-ai-agents/internal/ai/providers"
)

// BasicStrategy implements a basic caching strategy
type BasicStrategy struct {
	config         *Config
	strategyConfig *StrategyConfig
}

// ShouldCache determines if a response should be cached
func (s *BasicStrategy) ShouldCache(req *providers.GenerateRequest, resp *providers.GenerateResponse) bool {
	// Don't cache if response has errors
	if resp.FinishReason == "error" && !s.shouldCacheErrors() {
		return false
	}
	
	// Don't cache if response is too short or too long
	if s.strategyConfig != nil {
		if resp.Usage != nil {
			tokens := resp.Usage.TotalTokens
			if s.strategyConfig.MinTokens > 0 && tokens < s.strategyConfig.MinTokens {
				return false
			}
			if s.strategyConfig.MaxTokens > 0 && tokens > s.strategyConfig.MaxTokens {
				return false
			}
		}
	}
	
	// Don't cache if cost is too low (not worth caching)
	if s.strategyConfig != nil && s.strategyConfig.MinCostToCache > 0 {
		if resp.Cost == nil || resp.Cost.TotalCost < s.strategyConfig.MinCostToCache {
			return false
		}
	}
	
	// Don't cache streaming responses by default
	if req.Stream {
		return false
	}
	
	// Don't cache user-specific content unless configured
	if req.UserID != "" && s.strategyConfig != nil && !s.strategyConfig.UserSpecific {
		return false
	}
	
	return true
}

// GetTTL returns the TTL for a cached response
func (s *BasicStrategy) GetTTL(req *providers.GenerateRequest, resp *providers.GenerateResponse) time.Duration {
	if s.strategyConfig != nil && s.strategyConfig.TTL > 0 {
		return s.strategyConfig.TTL
	}
	
	return s.config.DefaultTTL
}

// GetKey generates a cache key for a request
func (s *BasicStrategy) GetKey(req *providers.GenerateRequest) string {
	return GenerateStandardKey(req)
}

// shouldCacheErrors returns whether errors should be cached
func (s *BasicStrategy) shouldCacheErrors() bool {
	return s.strategyConfig != nil && s.strategyConfig.CacheErrors
}

// ContentBasedStrategy implements content-based caching
type ContentBasedStrategy struct {
	config         *Config
	strategyConfig *StrategyConfig
}

// ShouldCache determines if a response should be cached based on content
func (s *ContentBasedStrategy) ShouldCache(req *providers.GenerateRequest, resp *providers.GenerateResponse) bool {
	// First check basic conditions
	basic := &BasicStrategy{config: s.config, strategyConfig: s.strategyConfig}
	if !basic.ShouldCache(req, resp) {
		return false
	}
	
	// Check content patterns if configured
	if s.strategyConfig != nil && len(s.strategyConfig.ContentPatterns) > 0 {
		content := s.getContentToCheck(req, resp)
		return s.matchesContentPatterns(content)
	}
	
	// Check if content is deterministic (good for caching)
	return s.isContentDeterministic(req, resp)
}

// GetTTL returns TTL based on content characteristics
func (s *ContentBasedStrategy) GetTTL(req *providers.GenerateRequest, resp *providers.GenerateResponse) time.Duration {
	baseTTL := s.config.DefaultTTL
	if s.strategyConfig != nil && s.strategyConfig.TTL > 0 {
		baseTTL = s.strategyConfig.TTL
	}
	
	// Adjust TTL based on content type
	if s.isFactualContent(resp) {
		// Factual content can be cached longer
		return baseTTL * 2
	}
	
	if s.isCreativeContent(resp) {
		// Creative content should have shorter TTL
		return baseTTL / 2
	}
	
	if s.isCodeContent(resp) {
		// Code can be cached longer if it's deterministic
		return baseTTL * 3
	}
	
	return baseTTL
}

// GetKey generates a content-aware cache key
func (s *ContentBasedStrategy) GetKey(req *providers.GenerateRequest) string {
	// Use content-specific key generation
	hasher := sha256.New()
	
	// Include normalized content
	content := s.normalizeContent(req)
	hasher.Write([]byte(content))
	
	// Include model and provider
	hasher.Write([]byte(req.Model))
	hasher.Write([]byte(req.Provider))
	
	// Include relevant parameters only
	if req.Temperature > 0 {
		hasher.Write([]byte(fmt.Sprintf("temp:%.1f", req.Temperature)))
	}
	
	hash := hex.EncodeToString(hasher.Sum(nil))
	return fmt.Sprintf("ai_cache:content:%s:%s:%s", req.Provider, req.Model, hash[:16])
}

// getContentToCheck extracts content for pattern matching
func (s *ContentBasedStrategy) getContentToCheck(req *providers.GenerateRequest, resp *providers.GenerateResponse) string {
	var content strings.Builder
	
	// Include request content
	if req.Prompt != "" {
		content.WriteString(req.Prompt)
	}
	
	for _, msg := range req.Messages {
		content.WriteString(msg.Content)
	}
	
	// Include response content
	content.WriteString(resp.Text)
	
	return content.String()
}

// matchesContentPatterns checks if content matches configured patterns
func (s *ContentBasedStrategy) matchesContentPatterns(content string) bool {
	for _, pattern := range s.strategyConfig.ContentPatterns {
		matched, err := regexp.MatchString(pattern, content)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// isContentDeterministic checks if content is likely to be deterministic
func (s *ContentBasedStrategy) isContentDeterministic(req *providers.GenerateRequest, resp *providers.GenerateResponse) bool {
	// Low temperature suggests deterministic output
	if req.Temperature <= 0.3 {
		return true
	}
	
	// Factual queries are often deterministic
	if s.isFactualQuery(req) {
		return true
	}
	
	// Code generation is often deterministic
	if s.isCodeQuery(req) {
		return true
	}
	
	return false
}

// isFactualContent checks if response contains factual information
func (s *ContentBasedStrategy) isFactualContent(resp *providers.GenerateResponse) bool {
	content := strings.ToLower(resp.Text)
	
	// Look for factual indicators
	factualIndicators := []string{
		"according to", "research shows", "studies indicate",
		"data suggests", "statistics show", "evidence indicates",
		"definition", "formula", "equation", "theorem",
	}
	
	for _, indicator := range factualIndicators {
		if strings.Contains(content, indicator) {
			return true
		}
	}
	
	return false
}

// isCreativeContent checks if response contains creative content
func (s *ContentBasedStrategy) isCreativeContent(resp *providers.GenerateResponse) bool {
	content := strings.ToLower(resp.Text)
	
	// Look for creative indicators
	creativeIndicators := []string{
		"once upon a time", "imagine", "creative", "story",
		"poem", "poetry", "metaphor", "artistic",
		"brainstorm", "innovative", "original",
	}
	
	for _, indicator := range creativeIndicators {
		if strings.Contains(content, indicator) {
			return true
		}
	}
	
	return false
}

// isCodeContent checks if response contains code
func (s *ContentBasedStrategy) isCodeContent(resp *providers.GenerateResponse) bool {
	content := resp.Text
	
	// Look for code indicators
	codeIndicators := []string{
		"```", "function", "class", "import", "export",
		"def ", "func ", "var ", "let ", "const ",
		"public class", "private ", "protected ",
		"#include", "package ", "namespace",
	}
	
	for _, indicator := range codeIndicators {
		if strings.Contains(content, indicator) {
			return true
		}
	}
	
	return false
}

// isFactualQuery checks if request is asking for factual information
func (s *ContentBasedStrategy) isFactualQuery(req *providers.GenerateRequest) bool {
	content := strings.ToLower(req.Prompt)
	for _, msg := range req.Messages {
		content += " " + strings.ToLower(msg.Content)
	}
	
	factualQuestions := []string{
		"what is", "what are", "define", "explain",
		"how does", "how do", "when did", "where is",
		"who is", "who was", "why does", "calculate",
		"formula for", "equation for", "convert",
	}
	
	for _, question := range factualQuestions {
		if strings.Contains(content, question) {
			return true
		}
	}
	
	return false
}

// isCodeQuery checks if request is asking for code
func (s *ContentBasedStrategy) isCodeQuery(req *providers.GenerateRequest) bool {
	content := strings.ToLower(req.Prompt)
	for _, msg := range req.Messages {
		content += " " + strings.ToLower(msg.Content)
	}
	
	codeQuestions := []string{
		"write code", "write a function", "implement",
		"code example", "programming", "algorithm",
		"debug", "fix this code", "optimize",
		"refactor", "write a script", "create a class",
	}
	
	for _, question := range codeQuestions {
		if strings.Contains(content, question) {
			return true
		}
	}
	
	return false
}

// normalizeContent normalizes content for consistent caching
func (s *ContentBasedStrategy) normalizeContent(req *providers.GenerateRequest) string {
	var content strings.Builder
	
	if req.Prompt != "" {
		content.WriteString(strings.TrimSpace(req.Prompt))
	}
	
	for _, msg := range req.Messages {
		content.WriteString(strings.TrimSpace(msg.Content))
	}
	
	// Normalize whitespace
	normalized := regexp.MustCompile(`\s+`).ReplaceAllString(content.String(), " ")
	return strings.TrimSpace(normalized)
}

// CostOptimizedStrategy implements cost-optimized caching
type CostOptimizedStrategy struct {
	config         *Config
	strategyConfig *StrategyConfig
}

// ShouldCache determines if a response should be cached based on cost optimization
func (s *CostOptimizedStrategy) ShouldCache(req *providers.GenerateRequest, resp *providers.GenerateResponse) bool {
	// First check basic conditions
	basic := &BasicStrategy{config: s.config, strategyConfig: s.strategyConfig}
	if !basic.ShouldCache(req, resp) {
		return false
	}
	
	// Always cache expensive responses
	if resp.Cost != nil && resp.Cost.TotalCost > 0.01 { // Cache responses costing more than 1 cent
		return true
	}
	
	// Cache responses with high token usage
	if resp.Usage != nil && resp.Usage.TotalTokens > 1000 {
		return true
	}
	
	// Cache responses from expensive models
	if s.isExpensiveModel(req.Model) {
		return true
	}
	
	return false
}

// GetTTL returns TTL based on cost considerations
func (s *CostOptimizedStrategy) GetTTL(req *providers.GenerateRequest, resp *providers.GenerateResponse) time.Duration {
	baseTTL := s.config.DefaultTTL
	if s.strategyConfig != nil && s.strategyConfig.TTL > 0 {
		baseTTL = s.strategyConfig.TTL
	}
	
	// Longer TTL for expensive responses
	if resp.Cost != nil {
		if resp.Cost.TotalCost > 0.10 { // Very expensive
			return baseTTL * 5
		} else if resp.Cost.TotalCost > 0.05 { // Expensive
			return baseTTL * 3
		} else if resp.Cost.TotalCost > 0.01 { // Moderate cost
			return baseTTL * 2
		}
	}
	
	// Longer TTL for high token usage
	if resp.Usage != nil {
		if resp.Usage.TotalTokens > 5000 {
			return baseTTL * 4
		} else if resp.Usage.TotalTokens > 2000 {
			return baseTTL * 2
		}
	}
	
	return baseTTL
}

// GetKey generates a cost-aware cache key
func (s *CostOptimizedStrategy) GetKey(req *providers.GenerateRequest) string {
	// Use standard key but with cost prefix
	standardKey := GenerateStandardKey(req)
	return strings.Replace(standardKey, "ai_cache:", "ai_cache:cost:", 1)
}

// isExpensiveModel checks if a model is expensive
func (s *CostOptimizedStrategy) isExpensiveModel(model string) bool {
	expensiveModels := []string{
		"gpt-4", "gpt-4-32k", "gpt-4-turbo",
		"claude-3-opus", "claude-2",
		"gemini-pro", "gemini-ultra",
	}
	
	modelLower := strings.ToLower(model)
	for _, expensive := range expensiveModels {
		if strings.Contains(modelLower, expensive) {
			return true
		}
	}
	
	return false
}

// UserSpecificStrategy implements user-specific caching
type UserSpecificStrategy struct {
	config         *Config
	strategyConfig *StrategyConfig
}

// ShouldCache determines if a response should be cached for a specific user
func (s *UserSpecificStrategy) ShouldCache(req *providers.GenerateRequest, resp *providers.GenerateResponse) bool {
	// Only cache if user ID is present
	if req.UserID == "" {
		return false
	}
	
	// Check basic conditions
	basic := &BasicStrategy{config: s.config, strategyConfig: s.strategyConfig}
	return basic.ShouldCache(req, resp)
}

// GetTTL returns TTL for user-specific content
func (s *UserSpecificStrategy) GetTTL(req *providers.GenerateRequest, resp *providers.GenerateResponse) time.Duration {
	// User-specific content typically has shorter TTL
	baseTTL := s.config.DefaultTTL
	if s.strategyConfig != nil && s.strategyConfig.TTL > 0 {
		baseTTL = s.strategyConfig.TTL
	}
	
	return baseTTL / 2 // Shorter TTL for user-specific content
}

// GetKey generates a user-specific cache key
func (s *UserSpecificStrategy) GetKey(req *providers.GenerateRequest) string {
	// Include user ID in the key
	hasher := sha256.New()
	
	// Include user ID
	hasher.Write([]byte(req.UserID))
	
	// Include standard content
	standardKey := GenerateStandardKey(req)
	hasher.Write([]byte(standardKey))
	
	hash := hex.EncodeToString(hasher.Sum(nil))
	return fmt.Sprintf("ai_cache:user:%s:%s:%s", req.UserID, req.Provider, hash[:16])
}
