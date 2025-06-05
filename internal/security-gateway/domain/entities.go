package domain

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// SecurityEventType represents the type of security event
type SecurityEventType string

const (
	SecurityEventTypeAuthentication     SecurityEventType = "authentication"
	SecurityEventTypeAuthorization     SecurityEventType = "authorization"
	SecurityEventTypeNetworkActivity   SecurityEventType = "network_activity"
	SecurityEventTypeMaliciousActivity SecurityEventType = "malicious_activity"
	SecurityEventTypePrivilegeEscalation SecurityEventType = "privilege_escalation"
	SecurityEventTypeDataAccess        SecurityEventType = "data_access"
	SecurityEventTypeSystemAccess      SecurityEventType = "system_access"
)

// SecuritySeverity represents the severity level of a security event
type SecuritySeverity string

const (
	SeverityInfo     SecuritySeverity = "info"
	SeverityLow      SecuritySeverity = "low"
	SeverityMedium   SecuritySeverity = "medium"
	SeverityHigh     SecuritySeverity = "high"
	SeverityCritical SecuritySeverity = "critical"
)

// SecurityRequest represents a request being processed by the security gateway
type SecurityRequest struct {
	ID            string            `json:"id"`
	Method        string            `json:"method"`
	URL           string            `json:"url"`
	Headers       map[string]string `json:"headers"`
	Body          []byte            `json:"body,omitempty"`
	IPAddress     string            `json:"ip_address"`
	UserAgent     string            `json:"user_agent"`
	Timestamp     time.Time         `json:"timestamp"`
	UserID        string            `json:"user_id,omitempty"`
	TenantID      string            `json:"tenant_id,omitempty"`
	SessionID     string            `json:"session_id,omitempty"`
	CorrelationID string            `json:"correlation_id"`
	
	// Security context
	ThreatLevel   ThreatLevel       `json:"threat_level"`
	RiskScore     float64           `json:"risk_score"`
	Blocked       bool              `json:"blocked"`
	BlockReason   string            `json:"block_reason,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityResponse represents a response from the security gateway
type SecurityResponse struct {
	RequestID     string            `json:"request_id"`
	StatusCode    int               `json:"status_code"`
	Headers       map[string]string `json:"headers"`
	Body          []byte            `json:"body,omitempty"`
	ProcessingTime time.Duration    `json:"processing_time"`
	Timestamp     time.Time         `json:"timestamp"`
	
	// Security context
	SecurityChecks []SecurityCheck   `json:"security_checks"`
	Warnings       []string          `json:"warnings,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityCheck represents a security check performed on a request
type SecurityCheck struct {
	Name        string                 `json:"name"`
	Type        SecurityCheckType      `json:"type"`
	Status      SecurityCheckStatus    `json:"status"`
	Result      SecurityCheckResult    `json:"result"`
	Message     string                 `json:"message,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityCheckType represents the type of security check
type SecurityCheckType string

const (
	SecurityCheckTypeAuthentication SecurityCheckType = "authentication"
	SecurityCheckTypeAuthorization  SecurityCheckType = "authorization"
	SecurityCheckTypeValidation     SecurityCheckType = "validation"
	SecurityCheckTypeRateLimit      SecurityCheckType = "rate_limit"
	SecurityCheckTypeWAF            SecurityCheckType = "waf"
	SecurityCheckTypeThreatDetection SecurityCheckType = "threat_detection"
	SecurityCheckTypeEncryption     SecurityCheckType = "encryption"
	SecurityCheckTypeAudit          SecurityCheckType = "audit"
)

// SecurityCheckStatus represents the status of a security check
type SecurityCheckStatus string

const (
	SecurityCheckStatusPassed  SecurityCheckStatus = "passed"
	SecurityCheckStatusFailed  SecurityCheckStatus = "failed"
	SecurityCheckStatusWarning SecurityCheckStatus = "warning"
	SecurityCheckStatusSkipped SecurityCheckStatus = "skipped"
)

// SecurityCheckResult represents the result of a security check
type SecurityCheckResult string

const (
	SecurityCheckResultAllow SecurityCheckResult = "allow"
	SecurityCheckResultBlock SecurityCheckResult = "block"
	SecurityCheckResultWarn  SecurityCheckResult = "warn"
)

// ThreatLevel represents the threat level of a request
type ThreatLevel string

const (
	ThreatLevelNone     ThreatLevel = "none"
	ThreatLevelLow      ThreatLevel = "low"
	ThreatLevelMedium   ThreatLevel = "medium"
	ThreatLevelHigh     ThreatLevel = "high"
	ThreatLevelCritical ThreatLevel = "critical"
)

// SecurityPolicy represents a security policy
type SecurityPolicy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Rules       []SecurityRule         `json:"rules"`
	Actions     []SecurityAction       `json:"actions"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// SecurityRule represents a security rule within a policy
type SecurityRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        SecurityRuleType       `json:"type"`
	Condition   SecurityCondition      `json:"condition"`
	Enabled     bool                   `json:"enabled"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityRuleType represents the type of security rule
type SecurityRuleType string

const (
	SecurityRuleTypeIPWhitelist    SecurityRuleType = "ip_whitelist"
	SecurityRuleTypeIPBlacklist    SecurityRuleType = "ip_blacklist"
	SecurityRuleTypeRateLimit      SecurityRuleType = "rate_limit"
	SecurityRuleTypeGeoBlocking    SecurityRuleType = "geo_blocking"
	SecurityRuleTypeUserAgent      SecurityRuleType = "user_agent"
	SecurityRuleTypeRequestSize    SecurityRuleType = "request_size"
	SecurityRuleTypeContentType    SecurityRuleType = "content_type"
	SecurityRuleTypeCustom         SecurityRuleType = "custom"
)

// SecurityCondition represents a condition for a security rule
type SecurityCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Values   []interface{} `json:"values,omitempty"`
}

// SecurityAction represents an action to take when a rule matches
type SecurityAction struct {
	Type     SecurityActionType     `json:"type"`
	Config   map[string]interface{} `json:"config,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityActionType represents the type of security action
type SecurityActionType string

const (
	SecurityActionTypeAllow     SecurityActionType = "allow"
	SecurityActionTypeBlock     SecurityActionType = "block"
	SecurityActionTypeChallenge SecurityActionType = "challenge"
	SecurityActionTypeLog       SecurityActionType = "log"
	SecurityActionTypeAlert     SecurityActionType = "alert"
	SecurityActionTypeRedirect  SecurityActionType = "redirect"
	SecurityActionTypeRateLimit SecurityActionType = "rate_limit"
)

// RateLimitInfo represents rate limiting information
type RateLimitInfo struct {
	Limit     int           `json:"limit"`
	Remaining int           `json:"remaining"`
	Reset     time.Time     `json:"reset"`
	Window    time.Duration `json:"window"`
	Blocked   bool          `json:"blocked"`
}

// WAFResult represents the result of WAF processing
type WAFResult struct {
	Allowed     bool                   `json:"allowed"`
	Blocked     bool                   `json:"blocked"`
	Reason      string                 `json:"reason,omitempty"`
	RuleMatched string                 `json:"rule_matched,omitempty"`
	Score       float64                `json:"score"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// ThreatIntelligence represents threat intelligence data
type ThreatIntelligence struct {
	IPAddress    string                 `json:"ip_address"`
	Reputation   ReputationScore        `json:"reputation"`
	Categories   []ThreatCategory       `json:"categories"`
	LastSeen     time.Time              `json:"last_seen"`
	Confidence   float64                `json:"confidence"`
	Source       string                 `json:"source"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ReputationScore represents the reputation score of an IP or domain
type ReputationScore string

const (
	ReputationScoreGood       ReputationScore = "good"
	ReputationScoreNeutral    ReputationScore = "neutral"
	ReputationScoreSuspicious ReputationScore = "suspicious"
	ReputationScoreMalicious  ReputationScore = "malicious"
)

// ThreatCategory represents a category of threat
type ThreatCategory string

const (
	ThreatCategoryMalware    ThreatCategory = "malware"
	ThreatCategoryPhishing   ThreatCategory = "phishing"
	ThreatCategoryBotnet     ThreatCategory = "botnet"
	ThreatCategorySpam       ThreatCategory = "spam"
	ThreatCategoryScanning   ThreatCategory = "scanning"
	ThreatCategoryBruteForce ThreatCategory = "brute_force"
	ThreatCategoryDDoS       ThreatCategory = "ddos"
)

// GeoLocation represents geographical location information
type GeoLocation struct {
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ISP         string  `json:"isp"`
	Organization string `json:"organization"`
}

// SecurityMetrics represents security metrics
type SecurityMetrics struct {
	TotalRequests       int64                    `json:"total_requests"`
	BlockedRequests     int64                    `json:"blocked_requests"`
	AllowedRequests     int64                    `json:"allowed_requests"`
	ThreatDetections    int64                    `json:"threat_detections"`
	RateLimitViolations int64                    `json:"rate_limit_violations"`
	WAFBlocks           int64                    `json:"waf_blocks"`
	RequestsByCountry   map[string]int64         `json:"requests_by_country"`
	ThreatsByCategory   map[ThreatCategory]int64 `json:"threats_by_category"`
	AverageResponseTime time.Duration            `json:"average_response_time"`
	LastUpdated         time.Time                `json:"last_updated"`
}

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID          string                 `json:"id"`
	Type        SecurityAlertType      `json:"type"`
	Severity    SecuritySeverity       `json:"severity"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Source      string                 `json:"source"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Status      SecurityAlertStatus    `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityAlertType represents the type of security alert
type SecurityAlertType string

const (
	SecurityAlertTypeThreatDetected    SecurityAlertType = "threat_detected"
	SecurityAlertTypeRateLimitExceeded SecurityAlertType = "rate_limit_exceeded"
	SecurityAlertTypeWAFBlock          SecurityAlertType = "waf_block"
	SecurityAlertTypeSuspiciousActivity SecurityAlertType = "suspicious_activity"
	SecurityAlertTypeSystemAnomaly     SecurityAlertType = "system_anomaly"
)



// SecurityAlertStatus represents the status of a security alert
type SecurityAlertStatus string

const (
	SecurityAlertStatusOpen       SecurityAlertStatus = "open"
	SecurityAlertStatusInvestigating SecurityAlertStatus = "investigating"
	SecurityAlertStatusResolved   SecurityAlertStatus = "resolved"
	SecurityAlertStatusFalsePositive SecurityAlertStatus = "false_positive"
)

// NewSecurityRequest creates a new security request from HTTP request
func NewSecurityRequest(r *http.Request) *SecurityRequest {
	headers := make(map[string]string)
	for name, values := range r.Header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}

	return &SecurityRequest{
		ID:            generateRequestID(),
		Method:        r.Method,
		URL:           r.URL.String(),
		Headers:       headers,
		IPAddress:     getClientIP(r),
		UserAgent:     r.UserAgent(),
		Timestamp:     time.Now(),
		CorrelationID: getCorrelationID(r),
		ThreatLevel:   ThreatLevelNone,
		RiskScore:     0.0,
		Blocked:       false,
		Metadata:      make(map[string]interface{}),
	}
}

// Helper functions
func generateRequestID() string {
	// Implementation to generate unique request ID
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func getCorrelationID(r *http.Request) string {
	if cid := r.Header.Get("X-Correlation-ID"); cid != "" {
		return cid
	}
	return generateRequestID()
}
