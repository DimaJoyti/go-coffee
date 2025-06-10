package security

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/auth/application"
	"github.com/DimaJoyti/go-coffee/internal/auth/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// SecurityService implements the SecurityService interface
type SecurityService struct {
	cacheService application.CacheService
	logger       *logger.Logger
	config       *Config
}

// Config represents security service configuration
type Config struct {
	// Rate limiting
	MaxLoginAttempts    int           `yaml:"max_login_attempts"`
	LoginAttemptWindow  time.Duration `yaml:"login_attempt_window"`
	AccountLockDuration time.Duration `yaml:"account_lock_duration"`

	// IP-based rate limiting
	MaxRequestsPerIP  int           `yaml:"max_requests_per_ip"`
	IPRateLimitWindow time.Duration `yaml:"ip_rate_limit_window"`
	IPBlockDuration   time.Duration `yaml:"ip_block_duration"`

	// Suspicious activity detection
	MaxFailedMFA          int           `yaml:"max_failed_mfa"`
	MFAFailureWindow      time.Duration `yaml:"mfa_failure_window"`
	SuspiciousIPThreshold int           `yaml:"suspicious_ip_threshold"`

	// Trusted networks
	TrustedNetworks []string `yaml:"trusted_networks"`

	// Security events retention
	EventRetentionPeriod time.Duration `yaml:"event_retention_period"`
}

// NewSecurityService creates a new security service
func NewSecurityService(
	cacheService application.CacheService,
	config *Config,
	logger *logger.Logger,
) application.SecurityService {
	return &SecurityService{
		cacheService: cacheService,
		config:       config,
		logger:       logger,
	}
}

// LogSecurityEvent logs a security event
func (s *SecurityService) LogSecurityEvent(ctx context.Context, userID string, eventType domain.SecurityEventType, severity domain.SecuritySeverity, description string, metadata map[string]string) error {
	event := &domain.SecurityEvent{
		ID:          generateEventID(),
		UserID:      userID,
		Type:        eventType,
		Severity:    severity,
		Description: description,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
	}

	// Store event in cache with expiration
	eventKey := fmt.Sprintf("security_event:%s:%s", userID, event.ID)
	if err := s.cacheService.Set(ctx, eventKey, event, s.config.EventRetentionPeriod); err != nil {
		s.logger.WithError(err).WithField("event_id", event.ID).Error("Failed to store security event")
		return fmt.Errorf("failed to store security event: %w", err)
	}

	// Add to user's event list
	userEventsKey := fmt.Sprintf("user_security_events:%s", userID)
	eventIDs := []string{}

	// Get existing events
	if data, err := s.cacheService.Get(ctx, userEventsKey); err != nil && err != application.ErrCacheKeyNotFound {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user security events")
	} else if err == nil {
		if existingEventIDs, ok := data.([]string); ok {
			eventIDs = existingEventIDs
		}
	}

	// Add new event ID
	eventIDs = append([]string{event.ID}, eventIDs...)

	// Keep only recent events (limit to 100)
	if len(eventIDs) > 100 {
		eventIDs = eventIDs[:100]
	}

	// Store updated list
	if err := s.cacheService.Set(ctx, userEventsKey, eventIDs, s.config.EventRetentionPeriod); err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to update user security events")
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id":     userID,
		"event_type":  eventType,
		"severity":    severity,
		"description": description,
	}).Info("Security event logged")

	return nil
}

// GetSecurityEvents retrieves security events for a user
func (s *SecurityService) GetSecurityEvents(ctx context.Context, userID string, limit int) ([]*application.SecurityEventDTO, error) {
	userEventsKey := fmt.Sprintf("user_security_events:%s", userID)

	// Get event IDs
	data, err := s.cacheService.Get(ctx, userEventsKey)
	if err != nil {
		if err == application.ErrCacheKeyNotFound {
			return []*application.SecurityEventDTO{}, nil
		}
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user security events")
		return nil, fmt.Errorf("failed to get security events: %w", err)
	}

	eventIDs, ok := data.([]string)
	if !ok {
		return []*application.SecurityEventDTO{}, nil
	}

	// Limit the number of events
	if limit > 0 && len(eventIDs) > limit {
		eventIDs = eventIDs[:limit]
	}

	// Retrieve individual events
	events := make([]*application.SecurityEventDTO, 0, len(eventIDs))
	for _, eventID := range eventIDs {
		eventKey := fmt.Sprintf("security_event:%s:%s", userID, eventID)
		eventData, err := s.cacheService.Get(ctx, eventKey)
		if err != nil {
			continue // Skip missing events
		}

		if event, ok := eventData.(*domain.SecurityEvent); ok {
			eventDTO := &application.SecurityEventDTO{
				ID:          event.ID,
				UserID:      event.UserID,
				Type:        string(event.Type),
				Severity:    string(event.Severity),
				Description: event.Description,
				Metadata:    event.Metadata,
				CreatedAt:   event.CreatedAt,
			}
			events = append(events, eventDTO)
		}
	}

	return events, nil
}

// CheckRateLimit checks if an action is rate limited
func (s *SecurityService) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	rateLimitKey := fmt.Sprintf("rate_limit:%s", key)

	// Get current count
	count, err := s.cacheService.IncrementWithExpiry(ctx, rateLimitKey, window)
	if err != nil {
		s.logger.WithError(err).WithField("key", key).Error("Failed to check rate limit")
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	exceeded := count > int64(limit)

	if exceeded {
		s.logger.WithFields(map[string]interface{}{
			"key":   key,
			"count": count,
			"limit": limit,
		}).Warn("Rate limit exceeded")
	}

	return exceeded, nil
}

// IncrementRateLimit increments the rate limit counter for a key
func (s *SecurityService) IncrementRateLimit(ctx context.Context, key string) error {
	rateLimitKey := fmt.Sprintf("rate_limit:%s", key)

	_, err := s.cacheService.IncrementWithExpiry(ctx, rateLimitKey, time.Hour)
	if err != nil {
		s.logger.WithError(err).WithField("key", key).Error("Failed to increment rate limit")
		return fmt.Errorf("failed to increment rate limit: %w", err)
	}

	return nil
}

// IsIPBlocked checks if an IP address is blocked
func (s *SecurityService) IsIPBlocked(ctx context.Context, ipAddress string) (bool, error) {
	// Check if IP is in trusted networks
	if s.isIPTrusted(ipAddress) {
		return false, nil
	}

	blockKey := fmt.Sprintf("ip_blocked:%s", ipAddress)
	blocked, err := s.cacheService.Exists(ctx, blockKey)
	if err != nil {
		s.logger.WithError(err).WithField("ip", ipAddress).Error("Failed to check IP block status")
		return false, fmt.Errorf("failed to check IP block status: %w", err)
	}

	return blocked, nil
}

// BlockIP blocks an IP address
func (s *SecurityService) BlockIP(ctx context.Context, ipAddress string, reason string) error {
	// Don't block trusted IPs
	if s.isIPTrusted(ipAddress) {
		s.logger.WithField("ip", ipAddress).Warn("Attempted to block trusted IP")
		return nil
	}

	blockKey := fmt.Sprintf("ip_blocked:%s", ipAddress)
	blockInfo := map[string]interface{}{
		"reason":     reason,
		"blocked_at": time.Now(),
	}

	if err := s.cacheService.Set(ctx, blockKey, blockInfo, s.config.IPBlockDuration); err != nil {
		s.logger.WithError(err).WithField("ip", ipAddress).Error("Failed to block IP")
		return fmt.Errorf("failed to block IP: %w", err)
	}

	s.logger.WithFields(map[string]interface{}{
		"ip":     ipAddress,
		"reason": reason,
	}).Warn("IP address blocked")

	return nil
}

// UnblockIP unblocks an IP address
func (s *SecurityService) UnblockIP(ctx context.Context, ipAddress string) error {
	blockKey := fmt.Sprintf("ip_blocked:%s", ipAddress)

	if err := s.cacheService.Delete(ctx, blockKey); err != nil {
		s.logger.WithError(err).WithField("ip", ipAddress).Error("Failed to unblock IP")
		return fmt.Errorf("failed to unblock IP: %w", err)
	}

	s.logger.WithField("ip", ipAddress).Info("IP address unblocked")
	return nil
}

// RecordFailedLogin records a failed login attempt
func (s *SecurityService) RecordFailedLogin(ctx context.Context, userID, ipAddress string) error {
	// Record for user
	userKey := fmt.Sprintf("failed_login:%s", userID)
	userCount, err := s.cacheService.IncrementWithExpiry(ctx, userKey, s.config.LoginAttemptWindow)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to record user failed login")
		return fmt.Errorf("failed to record user failed login: %w", err)
	}

	// Record for IP
	ipKey := fmt.Sprintf("failed_login_ip:%s", ipAddress)
	ipCount, err := s.cacheService.IncrementWithExpiry(ctx, ipKey, s.config.IPRateLimitWindow)
	if err != nil {
		s.logger.WithError(err).WithField("ip", ipAddress).Error("Failed to record IP failed login")
		return fmt.Errorf("failed to record IP failed login: %w", err)
	}

	// Check if user should be locked
	if userCount >= int64(s.config.MaxLoginAttempts) {
		if err := s.lockAccount(ctx, userID); err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Failed to lock account")
		}
	}

	// Check if IP should be blocked
	if ipCount >= int64(s.config.MaxRequestsPerIP) {
		if err := s.BlockIP(ctx, ipAddress, "Too many failed login attempts"); err != nil {
			s.logger.WithError(err).WithField("ip", ipAddress).Error("Failed to block IP")
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id":    userID,
		"ip":         ipAddress,
		"user_count": userCount,
		"ip_count":   ipCount,
	}).Warn("Failed login recorded")

	return nil
}

// IsAccountLocked checks if an account is locked
func (s *SecurityService) IsAccountLocked(ctx context.Context, userID string) (bool, error) {
	lockKey := fmt.Sprintf("account_locked:%s", userID)
	locked, err := s.cacheService.Exists(ctx, lockKey)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to check account lock status")
		return false, fmt.Errorf("failed to check account lock status: %w", err)
	}

	return locked, nil
}

// CheckAccountSecurity performs comprehensive account security checks
func (s *SecurityService) CheckAccountSecurity(ctx context.Context, userID string) error {
	// Check if account is locked
	locked, err := s.IsAccountLocked(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check account lock status: %w", err)
	}
	if locked {
		return fmt.Errorf("account is locked")
	}

	// Check for suspicious activity
	// This could include checking for unusual login patterns, etc.
	// For now, we'll just return nil (account is secure)

	return nil
}

// LockAccount locks a user account with a specific reason
func (s *SecurityService) LockAccount(ctx context.Context, userID string, reason string) error {
	return s.lockAccount(ctx, userID)
}

// TrackFailedLogin tracks a failed login attempt by email
func (s *SecurityService) TrackFailedLogin(ctx context.Context, email string) error {
	// Convert email to a consistent key format
	failedKey := fmt.Sprintf("failed_login_email:%s", email)

	_, err := s.cacheService.IncrementWithExpiry(ctx, failedKey, s.config.LoginAttemptWindow)
	if err != nil {
		s.logger.WithError(err).WithField("email", email).Error("Failed to track failed login")
		return fmt.Errorf("failed to track failed login: %w", err)
	}

	s.logger.WithField("email", email).Warn("Failed login tracked")
	return nil
}

// ResetFailedLoginCount resets the failed login count for an email
func (s *SecurityService) ResetFailedLoginCount(ctx context.Context, email string) error {
	failedKey := fmt.Sprintf("failed_login_email:%s", email)

	if err := s.cacheService.Delete(ctx, failedKey); err != nil {
		s.logger.WithError(err).WithField("email", email).Error("Failed to reset failed login count")
		return fmt.Errorf("failed to reset failed login count: %w", err)
	}

	s.logger.WithField("email", email).Debug("Failed login count reset")
	return nil
}

// UnlockAccount unlocks a user account
func (s *SecurityService) UnlockAccount(ctx context.Context, userID string) error {
	lockKey := fmt.Sprintf("account_locked:%s", userID)

	if err := s.cacheService.Delete(ctx, lockKey); err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to unlock account")
		return fmt.Errorf("failed to unlock account: %w", err)
	}

	// Also clear failed login attempts
	userKey := fmt.Sprintf("failed_login:%s", userID)
	s.cacheService.Delete(ctx, userKey)

	s.logger.WithField("user_id", userID).Info("Account unlocked")
	return nil
}

// RecordMFAFailure records a failed MFA attempt
func (s *SecurityService) RecordMFAFailure(ctx context.Context, userID string) error {
	mfaKey := fmt.Sprintf("mfa_failed:%s", userID)
	count, err := s.cacheService.IncrementWithExpiry(ctx, mfaKey, s.config.MFAFailureWindow)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to record MFA failure")
		return fmt.Errorf("failed to record MFA failure: %w", err)
	}

	// Check if account should be locked due to MFA failures
	if count >= int64(s.config.MaxFailedMFA) {
		if err := s.lockAccount(ctx, userID); err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Failed to lock account due to MFA failures")
		}
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id": userID,
		"count":   count,
	}).Warn("MFA failure recorded")

	return nil
}

// AnalyzeSuspiciousActivity analyzes user activity for suspicious patterns
func (s *SecurityService) AnalyzeSuspiciousActivity(ctx context.Context, userID, ipAddress, userAgent string) (*application.SecurityAnalysis, error) {
	analysis := &application.SecurityAnalysis{
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		RiskScore: 0.0,
		Factors:   []string{},
		Timestamp: time.Now(),
	}

	// Check if IP is from a different location than usual
	// (This would require a geolocation service in a real implementation)

	// Check if user agent is different than usual
	// (This would require storing user agent history)

	// Check login frequency
	loginKey := fmt.Sprintf("login_frequency:%s", userID)
	loginCount, err := s.cacheService.IncrementWithExpiry(ctx, loginKey, 24*time.Hour)
	if err == nil && loginCount > 10 {
		analysis.RiskScore += 0.2
		analysis.Factors = append(analysis.Factors, "High login frequency")
	}

	// Check if IP has been used by multiple users
	ipUsersKey := fmt.Sprintf("ip_users:%s", ipAddress)
	var ipUsers []string
	if data, err := s.cacheService.Get(ctx, ipUsersKey); err == nil {
		if userList, ok := data.([]string); ok {
			ipUsers = userList
			uniqueUsers := make(map[string]bool)
			for _, user := range ipUsers {
				uniqueUsers[user] = true
			}
			if len(uniqueUsers) > s.config.SuspiciousIPThreshold {
				analysis.RiskScore += 0.3
				analysis.Factors = append(analysis.Factors, "IP used by multiple users")
			}
		}
	}

	// Add current user to IP users list
	ipUsers = append(ipUsers, userID)
	s.cacheService.Set(ctx, ipUsersKey, ipUsers, 7*24*time.Hour)

	// Determine risk level
	switch {
	case analysis.RiskScore >= 0.7:
		analysis.RiskLevel = "HIGH"
	case analysis.RiskScore >= 0.4:
		analysis.RiskLevel = "MEDIUM"
	default:
		analysis.RiskLevel = "LOW"
	}

	s.logger.WithFields(map[string]interface{}{
		"user_id":    userID,
		"ip":         ipAddress,
		"risk_score": analysis.RiskScore,
		"risk_level": analysis.RiskLevel,
	}).Debug("Security analysis completed")

	return analysis, nil
}

// Helper methods

// lockAccount locks a user account
func (s *SecurityService) lockAccount(ctx context.Context, userID string) error {
	lockKey := fmt.Sprintf("account_locked:%s", userID)
	lockInfo := map[string]interface{}{
		"locked_at": time.Now(),
		"reason":    "Too many failed attempts",
	}

	if err := s.cacheService.Set(ctx, lockKey, lockInfo, s.config.AccountLockDuration); err != nil {
		return fmt.Errorf("failed to lock account: %w", err)
	}

	s.logger.WithField("user_id", userID).Warn("Account locked")
	return nil
}

// isIPTrusted checks if an IP is in trusted networks
func (s *SecurityService) isIPTrusted(ipAddress string) bool {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return false
	}

	for _, network := range s.config.TrustedNetworks {
		if strings.Contains(network, "/") {
			// CIDR notation
			_, cidr, err := net.ParseCIDR(network)
			if err != nil {
				continue
			}
			if cidr.Contains(ip) {
				return true
			}
		} else {
			// Single IP
			if ipAddress == network {
				return true
			}
		}
	}

	return false
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

// DefaultConfig returns default security service configuration
func DefaultConfig() *Config {
	return &Config{
		MaxLoginAttempts:      5,
		LoginAttemptWindow:    15 * time.Minute,
		AccountLockDuration:   30 * time.Minute,
		MaxRequestsPerIP:      100,
		IPRateLimitWindow:     time.Hour,
		IPBlockDuration:       24 * time.Hour,
		MaxFailedMFA:          3,
		MFAFailureWindow:      15 * time.Minute,
		SuspiciousIPThreshold: 5,
		TrustedNetworks: []string{
			"127.0.0.1",
			"::1",
			"10.0.0.0/8",
			"172.16.0.0/12",
			"192.168.0.0/16",
		},
		EventRetentionPeriod: 30 * 24 * time.Hour, // 30 days
	}
}
