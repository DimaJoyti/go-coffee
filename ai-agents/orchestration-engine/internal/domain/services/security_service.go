package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go-coffee-ai-agents/orchestration-engine/internal/common"
	"go-coffee-ai-agents/orchestration-engine/internal/config"
	"go-coffee-ai-agents/orchestration-engine/internal/infrastructure/security"
)

// SecurityService provides comprehensive security management
type SecurityService struct {
	authService       *security.AuthService
	jwtManager        *security.JWTManager
	rateLimiter       *security.RateLimiter
	inputValidator    *security.InputValidator
	auditLogger       *security.AuditLoggerImpl
	tokenBlacklist    *security.TokenBlacklist
	securityMiddleware *security.SecurityMiddleware
	
	config            *config.SecurityConfig
	logger            common.Logger
	
	// Security monitoring
	securityMetrics   *SecurityMetrics
	threatDetector    *ThreatDetector
	
	// Control
	mutex             sync.RWMutex
	stopCh            chan struct{}
}

// SecurityMetrics tracks security-related metrics
type SecurityMetrics struct {
	AuthenticationAttempts int64     `json:"authentication_attempts"`
	SuccessfulLogins       int64     `json:"successful_logins"`
	FailedLogins           int64     `json:"failed_logins"`
	BlockedRequests        int64     `json:"blocked_requests"`
	RateLimitViolations    int64     `json:"rate_limit_violations"`
	SecurityViolations     int64     `json:"security_violations"`
	ActiveSessions         int64     `json:"active_sessions"`
	TokensIssued           int64     `json:"tokens_issued"`
	TokensRevoked          int64     `json:"tokens_revoked"`
	LastSecurityEvent      time.Time `json:"last_security_event"`
	ThreatLevel            string    `json:"threat_level"`
	LastUpdated            time.Time `json:"last_updated"`
}

// ThreatDetector detects security threats
type ThreatDetector struct {
	suspiciousIPs     map[string]*IPThreatInfo
	attackPatterns    []AttackPattern
	config            *ThreatDetectionConfig
	logger            Logger
	mutex             sync.RWMutex
}

// IPThreatInfo contains threat information for an IP address
type IPThreatInfo struct {
	IP                string    `json:"ip"`
	FailedAttempts    int       `json:"failed_attempts"`
	LastFailedAttempt time.Time `json:"last_failed_attempt"`
	ThreatScore       float64   `json:"threat_score"`
	IsBlocked         bool      `json:"is_blocked"`
	BlockedUntil      time.Time `json:"blocked_until"`
	AttackTypes       []string  `json:"attack_types"`
}

// AttackPattern represents a known attack pattern
type AttackPattern struct {
	Name        string   `json:"name"`
	Patterns    []string `json:"patterns"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
}

// ThreatDetectionConfig contains threat detection configuration
type ThreatDetectionConfig struct {
	Enabled                bool          `json:"enabled"`
	MaxFailedAttempts      int           `json:"max_failed_attempts"`
	ThreatScoreThreshold   float64       `json:"threat_score_threshold"`
	BlockDuration          time.Duration `json:"block_duration"`
	MonitoringWindow       time.Duration `json:"monitoring_window"`
	EnableBehaviorAnalysis bool          `json:"enable_behavior_analysis"`
	EnableGeoBlocking      bool          `json:"enable_geo_blocking"`
	AllowedCountries       []string      `json:"allowed_countries"`
}

// SecurityReport represents a comprehensive security report
type SecurityReport struct {
	Timestamp           time.Time                    `json:"timestamp"`
	OverallThreatLevel  string                       `json:"overall_threat_level"`
	SecurityMetrics     *SecurityMetrics             `json:"security_metrics"`
	ActiveThreats       []*IPThreatInfo              `json:"active_threats"`
	RecentSecurityEvents []*security.AuditEntry      `json:"recent_security_events"`
	Recommendations     []string                     `json:"recommendations"`
	SystemHealth        *SecurityHealthStatus        `json:"system_health"`
}

// SecurityHealthStatus represents the health status of security components
type SecurityHealthStatus struct {
	AuthServiceHealth      string `json:"auth_service_health"`
	RateLimiterHealth      string `json:"rate_limiter_health"`
	AuditLoggerHealth      string `json:"audit_logger_health"`
	ThreatDetectorHealth   string `json:"threat_detector_health"`
	OverallHealth          string `json:"overall_health"`
}

// NewSecurityService creates a new security service
func NewSecurityService(
	authService *security.AuthService,
	jwtManager *security.JWTManager,
	rateLimiter *security.RateLimiter,
	inputValidator *security.InputValidator,
	auditLogger *security.AuditLoggerImpl,
	tokenBlacklist *security.TokenBlacklist,
	config *config.SecurityConfig,
	logger common.Logger,
) *SecurityService {
	
	ss := &SecurityService{
		authService:    authService,
		jwtManager:     jwtManager,
		rateLimiter:    rateLimiter,
		inputValidator: inputValidator,
		auditLogger:    auditLogger,
		tokenBlacklist: tokenBlacklist,
		config:         config,
		logger:         logger,
		securityMetrics: &SecurityMetrics{
			LastUpdated: time.Now(),
		},
		stopCh: make(chan struct{}),
	}

	// Initialize threat detector
	ss.threatDetector = NewThreatDetector(&ThreatDetectionConfig{
		Enabled:                true,
		MaxFailedAttempts:      5,
		ThreatScoreThreshold:   0.7,
		BlockDuration:          time.Hour,
		MonitoringWindow:       time.Hour,
		EnableBehaviorAnalysis: true,
		EnableGeoBlocking:      false,
		AllowedCountries:       []string{},
	}, logger)

	// Initialize security middleware
	ss.securityMiddleware = security.NewSecurityMiddleware(
		authService,
		rateLimiter,
		inputValidator,
		tokenBlacklist,
		&security.SecurityConfig{
			EnableAuthentication:  config.EnableAuthentication,
			EnableRateLimit:      config.EnableRateLimit,
			EnableInputValidation: config.EnableInputValidation,
			EnableSecurityHeaders: config.EnableSecurityHeaders,
			EnableCORS:           config.EnableCORS,
		},
		logger,
	)

	return ss
}

// Start starts the security service
func (ss *SecurityService) Start(ctx context.Context) error {
	ss.logger.Info("Starting security service")

	// Start monitoring routines
	go ss.metricsCollectionLoop(ctx)
	go ss.threatMonitoringLoop(ctx)
	go ss.cleanupLoop(ctx)

	ss.logger.Info("Security service started")
	return nil
}

// Stop stops the security service
func (ss *SecurityService) Stop(ctx context.Context) error {
	ss.logger.Info("Stopping security service")
	
	close(ss.stopCh)
	
	// Stop audit logger
	if ss.auditLogger != nil {
		ss.auditLogger.Stop()
	}
	
	ss.logger.Info("Security service stopped")
	return nil
}

// Authenticate authenticates a user
func (ss *SecurityService) Authenticate(ctx context.Context, req *security.AuthRequest) (*security.AuthResponse, error) {
	// Check for threats
	if ss.threatDetector.IsBlocked(req.IPAddress) {
		ss.recordSecurityViolation("blocked_ip_attempt", req.IPAddress, req.Username)
		return &security.AuthResponse{
			Success: false,
			Message: "Access denied",
		}, nil
	}

	// Perform authentication
	response, err := ss.authService.Authenticate(ctx, req)
	
	// Update metrics
	ss.mutex.Lock()
	ss.securityMetrics.AuthenticationAttempts++
	if response != nil && response.Success {
		ss.securityMetrics.SuccessfulLogins++
		ss.securityMetrics.TokensIssued++
		ss.securityMetrics.ActiveSessions++
	} else {
		ss.securityMetrics.FailedLogins++
		// Report failed attempt to threat detector
		ss.threatDetector.ReportFailedAttempt(req.IPAddress, req.Username, "authentication_failed")
	}
	ss.securityMetrics.LastUpdated = time.Now()
	ss.mutex.Unlock()

	return response, err
}

// ValidateToken validates a JWT token
func (ss *SecurityService) ValidateToken(ctx context.Context, token string) (*security.User, error) {
	return ss.authService.ValidateToken(ctx, token)
}

// RevokeToken revokes a JWT token
func (ss *SecurityService) RevokeToken(ctx context.Context, token string) error {
	// Parse token to get claims
	claims, err := ss.jwtManager.ValidateToken(token)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// Add to blacklist
	ss.tokenBlacklist.BlacklistToken(claims.ID, claims.ExpiresAt.Time)

	// Update metrics
	ss.mutex.Lock()
	ss.securityMetrics.TokensRevoked++
	ss.securityMetrics.ActiveSessions--
	ss.securityMetrics.LastUpdated = time.Now()
	ss.mutex.Unlock()

	// Log revocation
	if ss.auditLogger != nil {
		ss.auditLogger.LogSecurityEvent(ctx, &security.SecurityEvent{
			Type:      "token_revoked",
			UserID:    claims.UserID,
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"token_id": claims.ID,
			},
		})
	}

	return nil
}

// GetSecurityReport generates a comprehensive security report
func (ss *SecurityService) GetSecurityReport(ctx context.Context) (*SecurityReport, error) {
	ss.mutex.RLock()
	metrics := *ss.securityMetrics
	ss.mutex.RUnlock()

	// Get active threats
	activeThreats := ss.threatDetector.GetActiveThreats()

	// Get recent security events
	recentEvents, _ := ss.auditLogger.Query(ctx, &security.AuditFilter{
		Category: "security",
		Limit:    10,
	})

	// Generate recommendations
	recommendations := ss.generateSecurityRecommendations(&metrics, activeThreats)

	// Check system health
	systemHealth := ss.checkSystemHealth(ctx)

	report := &SecurityReport{
		Timestamp:            time.Now(),
		OverallThreatLevel:   ss.calculateOverallThreatLevel(&metrics, activeThreats),
		SecurityMetrics:      &metrics,
		ActiveThreats:        activeThreats,
		RecentSecurityEvents: recentEvents,
		Recommendations:      recommendations,
		SystemHealth:         systemHealth,
	}

	return report, nil
}

// GetSecurityMetrics returns current security metrics
func (ss *SecurityService) GetSecurityMetrics() *SecurityMetrics {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	
	metricsCopy := *ss.securityMetrics
	return &metricsCopy
}

// GetMiddleware returns the security middleware
func (ss *SecurityService) GetMiddleware() *security.SecurityMiddleware {
	return ss.securityMiddleware
}

// metricsCollectionLoop collects security metrics
func (ss *SecurityService) metricsCollectionLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ss.stopCh:
			return
		case <-ticker.C:
			ss.collectMetrics(ctx)
		}
	}
}

// threatMonitoringLoop monitors for security threats
func (ss *SecurityService) threatMonitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ss.stopCh:
			return
		case <-ticker.C:
			ss.threatDetector.AnalyzeThreats()
		}
	}
}

// cleanupLoop performs periodic cleanup
func (ss *SecurityService) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ss.stopCh:
			return
		case <-ticker.C:
			ss.performCleanup(ctx)
		}
	}
}

// collectMetrics collects current security metrics
func (ss *SecurityService) collectMetrics(ctx context.Context) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	// Update rate limiter metrics
	if ss.rateLimiter != nil {
		stats := ss.rateLimiter.GetStats()
		ss.securityMetrics.RateLimitViolations += int64(stats.BlockedTrackers)
	}

	// Update blacklisted tokens count
	if ss.tokenBlacklist != nil {
		// This would be implemented based on the blacklist interface
	}

	ss.securityMetrics.LastUpdated = time.Now()
}

// performCleanup performs periodic security cleanup
func (ss *SecurityService) performCleanup(ctx context.Context) {
	// Cleanup expired blacklisted tokens
	if ss.tokenBlacklist != nil {
		ss.tokenBlacklist.CleanupExpired()
	}

	// Cleanup old audit logs
	if ss.auditLogger != nil {
		if err := ss.auditLogger.Cleanup(ctx); err != nil {
			ss.logger.Error("Failed to cleanup audit logs", err)
		}
	}

	// Cleanup threat detector
	ss.threatDetector.Cleanup()

	ss.logger.Debug("Security cleanup completed")
}

// recordSecurityViolation records a security violation
func (ss *SecurityService) recordSecurityViolation(violationType, ipAddress, username string) {
	ss.mutex.Lock()
	ss.securityMetrics.SecurityViolations++
	ss.securityMetrics.LastSecurityEvent = time.Now()
	ss.securityMetrics.LastUpdated = time.Now()
	ss.mutex.Unlock()

	ss.logger.Warn("Security violation detected",
		"type", violationType,
		"ip", ipAddress,
		"username", username)
}

// generateSecurityRecommendations generates security recommendations
func (ss *SecurityService) generateSecurityRecommendations(metrics *SecurityMetrics, threats []*IPThreatInfo) []string {
	var recommendations []string

	// Check authentication metrics
	if metrics.FailedLogins > 0 && metrics.SuccessfulLogins > 0 {
		failureRate := float64(metrics.FailedLogins) / float64(metrics.AuthenticationAttempts) * 100
		if failureRate > 20 {
			recommendations = append(recommendations, "High authentication failure rate detected. Consider implementing additional security measures.")
		}
	}

	// Check active threats
	if len(threats) > 10 {
		recommendations = append(recommendations, "Multiple active threats detected. Consider implementing IP blocking or geographic restrictions.")
	}

	// Check rate limiting
	if metrics.RateLimitViolations > 100 {
		recommendations = append(recommendations, "High rate limit violations detected. Consider adjusting rate limiting thresholds.")
	}

	// Check security violations
	if metrics.SecurityViolations > 50 {
		recommendations = append(recommendations, "Multiple security violations detected. Review security policies and monitoring.")
	}

	return recommendations
}

// calculateOverallThreatLevel calculates the overall threat level
func (ss *SecurityService) calculateOverallThreatLevel(metrics *SecurityMetrics, threats []*IPThreatInfo) string {
	score := 0.0

	// Factor in failed logins
	if metrics.AuthenticationAttempts > 0 {
		failureRate := float64(metrics.FailedLogins) / float64(metrics.AuthenticationAttempts)
		score += failureRate * 30
	}

	// Factor in active threats
	score += float64(len(threats)) * 5

	// Factor in security violations
	score += float64(metrics.SecurityViolations) * 2

	// Factor in rate limit violations
	score += float64(metrics.RateLimitViolations) * 0.1

	if score >= 80 {
		return "critical"
	} else if score >= 60 {
		return "high"
	} else if score >= 40 {
		return "medium"
	} else if score >= 20 {
		return "low"
	} else {
		return "minimal"
	}
}

// checkSystemHealth checks the health of security components
func (ss *SecurityService) checkSystemHealth(ctx context.Context) *SecurityHealthStatus {
	health := &SecurityHealthStatus{
		AuthServiceHealth:     "healthy",
		RateLimiterHealth:     "healthy",
		AuditLoggerHealth:     "healthy",
		ThreatDetectorHealth:  "healthy",
		OverallHealth:         "healthy",
	}

	// Check each component
	// This would be implemented based on actual health check methods

	return health
}

// NewThreatDetector creates a new threat detector
func NewThreatDetector(config *ThreatDetectionConfig, logger Logger) *ThreatDetector {
	return &ThreatDetector{
		suspiciousIPs:  make(map[string]*IPThreatInfo),
		attackPatterns: getDefaultAttackPatterns(),
		config:         config,
		logger:         logger,
	}
}

// ReportFailedAttempt reports a failed authentication attempt
func (td *ThreatDetector) ReportFailedAttempt(ip, username, attackType string) {
	td.mutex.Lock()
	defer td.mutex.Unlock()

	info, exists := td.suspiciousIPs[ip]
	if !exists {
		info = &IPThreatInfo{
			IP:           ip,
			AttackTypes:  make([]string, 0),
		}
		td.suspiciousIPs[ip] = info
	}

	info.FailedAttempts++
	info.LastFailedAttempt = time.Now()
	info.ThreatScore += 0.1

	// Add attack type if not already present
	found := false
	for _, existing := range info.AttackTypes {
		if existing == attackType {
			found = true
			break
		}
	}
	if !found {
		info.AttackTypes = append(info.AttackTypes, attackType)
	}

	// Check if should be blocked
	if info.FailedAttempts >= td.config.MaxFailedAttempts || info.ThreatScore >= td.config.ThreatScoreThreshold {
		info.IsBlocked = true
		info.BlockedUntil = time.Now().Add(td.config.BlockDuration)
		td.logger.Warn("IP blocked due to suspicious activity", "ip", ip, "failed_attempts", info.FailedAttempts, "threat_score", info.ThreatScore)
	}
}

// IsBlocked checks if an IP is blocked
func (td *ThreatDetector) IsBlocked(ip string) bool {
	td.mutex.RLock()
	defer td.mutex.RUnlock()

	info, exists := td.suspiciousIPs[ip]
	if !exists {
		return false
	}

	if info.IsBlocked && time.Now().Before(info.BlockedUntil) {
		return true
	}

	// Unblock if time has passed
	if info.IsBlocked && time.Now().After(info.BlockedUntil) {
		info.IsBlocked = false
		info.BlockedUntil = time.Time{}
	}

	return false
}

// GetActiveThreats returns currently active threats
func (td *ThreatDetector) GetActiveThreats() []*IPThreatInfo {
	td.mutex.RLock()
	defer td.mutex.RUnlock()

	var threats []*IPThreatInfo
	for _, info := range td.suspiciousIPs {
		if info.IsBlocked || info.ThreatScore > 0.5 {
			threatCopy := *info
			threats = append(threats, &threatCopy)
		}
	}

	return threats
}

// AnalyzeThreats analyzes current threats
func (td *ThreatDetector) AnalyzeThreats() {
	td.mutex.Lock()
	defer td.mutex.Unlock()

	now := time.Now()
	for ip, info := range td.suspiciousIPs {
		// Decay threat score over time
		if now.Sub(info.LastFailedAttempt) > td.config.MonitoringWindow {
			info.ThreatScore *= 0.9
			if info.ThreatScore < 0.1 {
				delete(td.suspiciousIPs, ip)
			}
		}
	}
}

// Cleanup removes old threat information
func (td *ThreatDetector) Cleanup() {
	td.mutex.Lock()
	defer td.mutex.Unlock()

	now := time.Now()
	for ip, info := range td.suspiciousIPs {
		if now.Sub(info.LastFailedAttempt) > 24*time.Hour && !info.IsBlocked {
			delete(td.suspiciousIPs, ip)
		}
	}
}

// getDefaultAttackPatterns returns default attack patterns
func getDefaultAttackPatterns() []AttackPattern {
	return []AttackPattern{
		{
			Name:        "SQL Injection",
			Patterns:    []string{"union select", "drop table", "' or 1=1"},
			Severity:    "high",
			Description: "SQL injection attack attempt",
		},
		{
			Name:        "XSS Attack",
			Patterns:    []string{"<script>", "javascript:", "onerror="},
			Severity:    "medium",
			Description: "Cross-site scripting attack attempt",
		},
		{
			Name:        "Path Traversal",
			Patterns:    []string{"../", "..\\", "%2e%2e"},
			Severity:    "medium",
			Description: "Path traversal attack attempt",
		},
	}
}
