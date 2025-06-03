package application

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/security-gateway/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/DimaJoyti/go-coffee/pkg/security/monitoring"
	"github.com/DimaJoyti/go-coffee/pkg/security/validation"
)

// WAFService provides Web Application Firewall functionality
type WAFService struct {
	config            *WAFConfig
	validationService *validation.ValidationService
	monitoringService *monitoring.SecurityMonitoringService
	logger            *logger.Logger
	rules             []WAFRule
	ipWhitelist       map[string]bool
	ipBlacklist       map[string]bool
	countryBlacklist  map[string]bool
}

// WAFConfig represents WAF configuration
type WAFConfig struct {
	Enabled           bool     `yaml:"enabled"`
	BlockSuspiciousIP bool     `yaml:"block_suspicious_ip"`
	AllowedCountries  []string `yaml:"allowed_countries"`
	BlockedCountries  []string `yaml:"blocked_countries"`
	MaxRequestSize    int64    `yaml:"max_request_size"`
	EnableGeoBlocking bool     `yaml:"enable_geo_blocking"`
	EnableBotDetection bool    `yaml:"enable_bot_detection"`
	EnableRateLimiting bool    `yaml:"enable_rate_limiting"`
}

// WAFRule represents a WAF rule
type WAFRule struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Type        WAFRuleType      `json:"type"`
	Pattern     *regexp.Regexp   `json:"-"`
	PatternStr  string           `json:"pattern"`
	Action      WAFAction        `json:"action"`
	Severity    domain.SecuritySeverity `json:"severity"`
	Enabled     bool             `json:"enabled"`
	Score       int              `json:"score"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WAFRuleType represents the type of WAF rule
type WAFRuleType string

const (
	WAFRuleTypeSQLInjection    WAFRuleType = "sql_injection"
	WAFRuleTypeXSS             WAFRuleType = "xss"
	WAFRuleTypePathTraversal   WAFRuleType = "path_traversal"
	WAFRuleTypeCommandInjection WAFRuleType = "command_injection"
	WAFRuleTypeLDAPInjection   WAFRuleType = "ldap_injection"
	WAFRuleTypeXMLInjection    WAFRuleType = "xml_injection"
	WAFRuleTypeSSRF            WAFRuleType = "ssrf"
	WAFRuleTypeFileUpload      WAFRuleType = "file_upload"
	WAFRuleTypeUserAgent       WAFRuleType = "user_agent"
	WAFRuleTypeCustom          WAFRuleType = "custom"
)

// WAFAction represents the action to take when a rule matches
type WAFAction string

const (
	WAFActionAllow     WAFAction = "allow"
	WAFActionBlock     WAFAction = "block"
	WAFActionChallenge WAFAction = "challenge"
	WAFActionLog       WAFAction = "log"
)

// NewWAFService creates a new WAF service
func NewWAFService(
	config *WAFConfig,
	validationService *validation.ValidationService,
	monitoringService *monitoring.SecurityMonitoringService,
	logger *logger.Logger,
) *WAFService {
	service := &WAFService{
		config:            config,
		validationService: validationService,
		monitoringService: monitoringService,
		logger:            logger,
		ipWhitelist:       make(map[string]bool),
		ipBlacklist:       make(map[string]bool),
		countryBlacklist:  make(map[string]bool),
	}

	// Initialize rules
	service.initializeRules()

	// Initialize blacklists
	for _, country := range config.BlockedCountries {
		service.countryBlacklist[strings.ToUpper(country)] = true
	}

	return service
}

// CheckRequest checks a request against WAF rules
func (w *WAFService) CheckRequest(ctx context.Context, req *domain.SecurityRequest) (*domain.WAFResult, error) {
	if !w.config.Enabled {
		return &domain.WAFResult{
			Allowed: true,
			Blocked: false,
			Score:   0,
		}, nil
	}

	result := &domain.WAFResult{
		Allowed: true,
		Blocked: false,
		Score:   0,
		Details: make(map[string]interface{}),
	}

	// Check request size
	if w.config.MaxRequestSize > 0 && int64(len(req.Body)) > w.config.MaxRequestSize {
		result.Allowed = false
		result.Blocked = true
		result.Reason = "Request size exceeds maximum allowed"
		result.Score = 100
		
		w.logWAFEvent(ctx, req, "request_size_exceeded", domain.SeverityMedium)
		return result, nil
	}

	// Check IP whitelist/blacklist
	if blocked, reason := w.checkIPRestrictions(req.IPAddress); blocked {
		result.Allowed = false
		result.Blocked = true
		result.Reason = reason
		result.Score = 100
		
		w.logWAFEvent(ctx, req, "ip_blocked", domain.SeverityHigh)
		return result, nil
	}

	// Check geo-blocking
	if w.config.EnableGeoBlocking {
		if blocked, reason := w.checkGeoBlocking(ctx, req.IPAddress); blocked {
			result.Allowed = false
			result.Blocked = true
			result.Reason = reason
			result.Score = 80
			
			w.logWAFEvent(ctx, req, "geo_blocked", domain.SeverityMedium)
			return result, nil
		}
	}

	// Check bot detection
	if w.config.EnableBotDetection {
		if isBot, confidence := w.detectBot(req); isBot {
			result.Score += int(confidence * 50)
			result.Details["bot_detection"] = map[string]interface{}{
				"is_bot":     true,
				"confidence": confidence,
			}
			
			if confidence > 0.8 {
				result.Allowed = false
				result.Blocked = true
				result.Reason = "Malicious bot detected"
				
				w.logWAFEvent(ctx, req, "bot_detected", domain.SeverityHigh)
				return result, nil
			}
		}
	}

	// Check WAF rules
	for _, rule := range w.rules {
		if !rule.Enabled {
			continue
		}

		matched, matchDetails := w.checkRule(rule, req)
		if matched {
			result.Score += rule.Score
			result.RuleMatched = rule.ID
			result.Details[rule.ID] = matchDetails

			w.logWAFEvent(ctx, req, fmt.Sprintf("rule_matched_%s", rule.ID), rule.Severity)

			if rule.Action == WAFActionBlock {
				result.Allowed = false
				result.Blocked = true
				result.Reason = fmt.Sprintf("WAF rule violation: %s", rule.Name)
				return result, nil
			}
		}
	}

	// Check overall score
	if result.Score >= 100 {
		result.Allowed = false
		result.Blocked = true
		result.Reason = "High risk score threshold exceeded"
		
		w.logWAFEvent(ctx, req, "high_risk_score", domain.SeverityHigh)
	}

	return result, nil
}

// Initialize WAF rules
func (w *WAFService) initializeRules() {
	w.rules = []WAFRule{
		// SQL Injection rules
		{
			ID:          "sql_001",
			Name:        "SQL Injection - Union Select",
			Description: "Detects UNION SELECT SQL injection attempts",
			Type:        WAFRuleTypeSQLInjection,
			PatternStr:  `(?i)(union\s+select|union\s+all\s+select)`,
			Action:      WAFActionBlock,
			Severity:    domain.SeverityHigh,
			Enabled:     true,
			Score:       80,
		},
		{
			ID:          "sql_002",
			Name:        "SQL Injection - Comments",
			Description: "Detects SQL comment injection attempts",
			Type:        WAFRuleTypeSQLInjection,
			PatternStr:  `(?i)(--|\#|\/\*|\*\/|;)`,
			Action:      WAFActionBlock,
			Severity:    domain.SeverityHigh,
			Enabled:     true,
			Score:       70,
		},
		
		// XSS rules
		{
			ID:          "xss_001",
			Name:        "XSS - Script Tags",
			Description: "Detects script tag XSS attempts",
			Type:        WAFRuleTypeXSS,
			PatternStr:  `(?i)<script[^>]*>.*?</script>`,
			Action:      WAFActionBlock,
			Severity:    domain.SeverityHigh,
			Enabled:     true,
			Score:       80,
		},
		{
			ID:          "xss_002",
			Name:        "XSS - Event Handlers",
			Description: "Detects event handler XSS attempts",
			Type:        WAFRuleTypeXSS,
			PatternStr:  `(?i)on\w+\s*=`,
			Action:      WAFActionBlock,
			Severity:    domain.SeverityMedium,
			Enabled:     true,
			Score:       60,
		},
		
		// Path Traversal rules
		{
			ID:          "path_001",
			Name:        "Path Traversal - Directory Traversal",
			Description: "Detects directory traversal attempts",
			Type:        WAFRuleTypePathTraversal,
			PatternStr:  `(?i)(\.\.\/|\.\.\\|%2e%2e%2f|%2e%2e%5c)`,
			Action:      WAFActionBlock,
			Severity:    domain.SeverityHigh,
			Enabled:     true,
			Score:       80,
		},
		
		// Command Injection rules
		{
			ID:          "cmd_001",
			Name:        "Command Injection - System Commands",
			Description: "Detects system command injection attempts",
			Type:        WAFRuleTypeCommandInjection,
			PatternStr:  `(?i)(;|\||&|&&|\|\||` + "`" + `|\$\(|\$\{)`,
			Action:      WAFActionBlock,
			Severity:    domain.SeverityHigh,
			Enabled:     true,
			Score:       80,
		},
		
		// User Agent rules
		{
			ID:          "ua_001",
			Name:        "Malicious User Agent",
			Description: "Detects known malicious user agents",
			Type:        WAFRuleTypeUserAgent,
			PatternStr:  `(?i)(sqlmap|nikto|nmap|masscan|zap|burp|acunetix|nessus|openvas)`,
			Action:      WAFActionBlock,
			Severity:    domain.SeverityHigh,
			Enabled:     true,
			Score:       90,
		},
	}

	// Compile regex patterns
	for i := range w.rules {
		if pattern, err := regexp.Compile(w.rules[i].PatternStr); err == nil {
			w.rules[i].Pattern = pattern
		} else {
			w.logger.WithError(err).Error("Failed to compile WAF rule pattern", map[string]any{
				"rule_id": w.rules[i].ID,
				"pattern": w.rules[i].PatternStr,
			})
		}
	}
}

// Check a specific rule against the request
func (w *WAFService) checkRule(rule WAFRule, req *domain.SecurityRequest) (bool, map[string]interface{}) {
	if rule.Pattern == nil {
		return false, nil
	}

	details := make(map[string]interface{})
	
	// Check URL
	if rule.Pattern.MatchString(req.URL) {
		details["matched_in"] = "url"
		details["matched_value"] = req.URL
		return true, details
	}

	// Check headers
	for name, value := range req.Headers {
		if rule.Type == WAFRuleTypeUserAgent && name == "User-Agent" {
			if rule.Pattern.MatchString(value) {
				details["matched_in"] = "user_agent"
				details["matched_value"] = value
				return true, details
			}
		} else if rule.Pattern.MatchString(value) {
			details["matched_in"] = fmt.Sprintf("header_%s", name)
			details["matched_value"] = value
			return true, details
		}
	}

	// Check body
	if len(req.Body) > 0 {
		bodyStr := string(req.Body)
		if rule.Pattern.MatchString(bodyStr) {
			details["matched_in"] = "body"
			details["matched_value"] = bodyStr[:min(len(bodyStr), 100)] // Truncate for logging
			return true, details
		}
	}

	return false, nil
}

// Check IP restrictions
func (w *WAFService) checkIPRestrictions(ipAddress string) (bool, string) {
	// Check whitelist first
	if len(w.ipWhitelist) > 0 && !w.ipWhitelist[ipAddress] {
		return true, "IP not in whitelist"
	}

	// Check blacklist
	if w.ipBlacklist[ipAddress] {
		return true, "IP in blacklist"
	}

	// Check for private/local IPs in production
	if ip := net.ParseIP(ipAddress); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() {
			// Allow private IPs in development, block in production
			// This would be configurable based on environment
		}
	}

	return false, ""
}

// Check geo-blocking
func (w *WAFService) checkGeoBlocking(ctx context.Context, ipAddress string) (bool, string) {
	// TODO: Implement actual geo-location lookup
	// For now, this is a placeholder
	
	// In a real implementation, you would:
	// 1. Use a geo-location service (MaxMind, IP2Location, etc.)
	// 2. Look up the country for the IP address
	// 3. Check against allowed/blocked countries
	
	return false, ""
}

// Detect bots
func (w *WAFService) detectBot(req *domain.SecurityRequest) (bool, float64) {
	userAgent := req.Headers["User-Agent"]
	if userAgent == "" {
		return true, 0.8 // Missing user agent is suspicious
	}

	// Check for known bot patterns
	botPatterns := []string{
		"bot", "crawler", "spider", "scraper", "curl", "wget", "python", "java",
		"go-http-client", "okhttp", "apache-httpclient",
	}

	lowerUA := strings.ToLower(userAgent)
	for _, pattern := range botPatterns {
		if strings.Contains(lowerUA, pattern) {
			// Check if it's a legitimate bot
			legitimateBots := []string{
				"googlebot", "bingbot", "slurp", "duckduckbot", "baiduspider",
				"yandexbot", "facebookexternalhit", "twitterbot",
			}
			
			for _, legitBot := range legitimateBots {
				if strings.Contains(lowerUA, legitBot) {
					return false, 0.0 // Legitimate bot
				}
			}
			
			return true, 0.7 // Suspicious bot
		}
	}

	// Check for suspicious patterns
	if len(userAgent) < 10 || len(userAgent) > 500 {
		return true, 0.6
	}

	return false, 0.0
}

// Log WAF events
func (w *WAFService) logWAFEvent(ctx context.Context, req *domain.SecurityRequest, eventType string, severity domain.SecuritySeverity) {
	event := &monitoring.SecurityEvent{
		EventType:   monitoring.SecurityEventTypeMaliciousActivity,
		Severity:    monitoring.SecuritySeverity(severity),
		Source:      "waf",
		UserID:      req.UserID,
		IPAddress:   req.IPAddress,
		UserAgent:   req.UserAgent,
		Description: fmt.Sprintf("WAF event: %s", eventType),
		Metadata: map[string]interface{}{
			"request_id":     req.ID,
			"correlation_id": req.CorrelationID,
			"method":         req.Method,
			"url":            req.URL,
			"event_type":     eventType,
		},
	}

	w.monitoringService.LogSecurityEvent(ctx, event)
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
