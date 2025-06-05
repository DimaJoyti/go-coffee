package application

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/security-gateway/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/DimaJoyti/go-coffee/pkg/security/encryption"
	"github.com/DimaJoyti/go-coffee/pkg/security/monitoring"
	"github.com/DimaJoyti/go-coffee/pkg/security/validation"
)

// SecurityGatewayService provides security gateway functionality
type SecurityGatewayService struct {
	config            *GatewayConfig
	encryptionService *encryption.EncryptionService
	validationService *validation.ValidationService
	monitoringService *monitoring.SecurityMonitoringService
	rateLimitService  *RateLimitService
	wafService        *WAFService
	logger            *logger.Logger
	proxies           map[string]*httputil.ReverseProxy
}

// GatewayConfig represents gateway configuration
type GatewayConfig struct {
	Services          map[string]string `yaml:"services"`
	Timeout           time.Duration     `yaml:"timeout" default:"30s"`
	MaxRetries        int               `yaml:"max_retries" default:"3"`
	RetryDelay        time.Duration     `yaml:"retry_delay" default:"1s"`
	EnableCircuitBreaker bool           `yaml:"enable_circuit_breaker" default:"true"`
	EnableLoadBalancing  bool           `yaml:"enable_load_balancing" default:"false"`
}

// NewSecurityGatewayService creates a new security gateway service
func NewSecurityGatewayService(
	config *GatewayConfig,
	encryptionService *encryption.EncryptionService,
	validationService *validation.ValidationService,
	monitoringService *monitoring.SecurityMonitoringService,
	rateLimitService *RateLimitService,
	wafService *WAFService,
	logger *logger.Logger,
) *SecurityGatewayService {
	service := &SecurityGatewayService{
		config:            config,
		encryptionService: encryptionService,
		validationService: validationService,
		monitoringService: monitoringService,
		rateLimitService:  rateLimitService,
		wafService:        wafService,
		logger:            logger,
		proxies:           make(map[string]*httputil.ReverseProxy),
	}

	// Initialize reverse proxies for each service
	for serviceName, serviceURL := range config.Services {
		if targetURL, err := url.Parse(serviceURL); err == nil {
			service.proxies[serviceName] = httputil.NewSingleHostReverseProxy(targetURL)
		} else {
			logger.WithError(err).Error("Failed to parse service URL", map[string]any{
				"service": serviceName,
				"url":     serviceURL,
			})
		}
	}

	return service
}

// ProcessRequest processes an incoming request through the security gateway
func (s *SecurityGatewayService) ProcessRequest(ctx context.Context, r *http.Request) (*domain.SecurityResponse, error) {
	startTime := time.Now()

	// Create security request
	securityRequest := domain.NewSecurityRequest(r)

	// Log security event
	s.logSecurityEvent(ctx, securityRequest, domain.SecurityEventTypeNetworkActivity, domain.SeverityInfo, "Request received")

	// Perform security checks
	securityChecks := []domain.SecurityCheck{}

	// 1. Rate limiting check
	rateLimitCheck, err := s.performRateLimitCheck(ctx, securityRequest)
	if err != nil {
		s.logger.WithError(err).Error("Rate limit check failed")
	}
	securityChecks = append(securityChecks, *rateLimitCheck)

	if rateLimitCheck.Result == domain.SecurityCheckResultBlock {
		return s.createBlockedResponse(securityRequest, "Rate limit exceeded", securityChecks, startTime), nil
	}

	// 2. WAF check
	wafCheck, err := s.performWAFCheck(ctx, securityRequest)
	if err != nil {
		s.logger.WithError(err).Error("WAF check failed")
	}
	securityChecks = append(securityChecks, *wafCheck)

	if wafCheck.Result == domain.SecurityCheckResultBlock {
		return s.createBlockedResponse(securityRequest, "WAF rule violation", securityChecks, startTime), nil
	}

	// 3. Input validation check
	validationCheck, err := s.performValidationCheck(ctx, securityRequest)
	if err != nil {
		s.logger.WithError(err).Error("Validation check failed")
	}
	securityChecks = append(securityChecks, *validationCheck)

	if validationCheck.Result == domain.SecurityCheckResultBlock {
		return s.createBlockedResponse(securityRequest, "Input validation failed", securityChecks, startTime), nil
	}

	// 4. Authentication check (if required)
	authCheck, err := s.performAuthenticationCheck(ctx, securityRequest)
	if err != nil {
		s.logger.WithError(err).Error("Authentication check failed")
	}
	if authCheck != nil {
		securityChecks = append(securityChecks, *authCheck)

		if authCheck.Result == domain.SecurityCheckResultBlock {
			return s.createBlockedResponse(securityRequest, "Authentication failed", securityChecks, startTime), nil
		}
	}

	// 5. Authorization check (if required)
	authzCheck, err := s.performAuthorizationCheck(ctx, securityRequest)
	if err != nil {
		s.logger.WithError(err).Error("Authorization check failed")
	}
	if authzCheck != nil {
		securityChecks = append(securityChecks, *authzCheck)

		if authzCheck.Result == domain.SecurityCheckResultBlock {
			return s.createBlockedResponse(securityRequest, "Authorization failed", securityChecks, startTime), nil
		}
	}

	// All checks passed - create successful response
	response := &domain.SecurityResponse{
		RequestID:      securityRequest.ID,
		StatusCode:     http.StatusOK,
		Headers:        make(map[string]string),
		ProcessingTime: time.Since(startTime),
		Timestamp:      time.Now(),
		SecurityChecks: securityChecks,
	}

	// Add security headers
	response.Headers["X-Request-ID"] = securityRequest.ID
	response.Headers["X-Correlation-ID"] = securityRequest.CorrelationID
	response.Headers["X-Security-Gateway"] = "go-coffee-security-gateway"

	return response, nil
}

// ProxyRequest proxies a request to the target service
func (s *SecurityGatewayService) ProxyRequest(serviceName string, w http.ResponseWriter, r *http.Request) error {
	proxy, exists := s.proxies[serviceName]
	if !exists {
		return fmt.Errorf("service not found: %s", serviceName)
	}

	// Add security headers
	r.Header.Set("X-Gateway", "security-gateway")
	r.Header.Set("X-Gateway-Version", "1.0.0")

	// Proxy the request
	proxy.ServeHTTP(w, r)

	return nil
}

// Security check methods

func (s *SecurityGatewayService) performRateLimitCheck(ctx context.Context, req *domain.SecurityRequest) (*domain.SecurityCheck, error) {
	startTime := time.Now()
	
	check := &domain.SecurityCheck{
		Name:     "rate_limit",
		Type:     domain.SecurityCheckTypeRateLimit,
		Status:   domain.SecurityCheckStatusPassed,
		Result:   domain.SecurityCheckResultAllow,
		Duration: time.Since(startTime),
	}

	// Check rate limit
	allowed, info, err := s.rateLimitService.CheckRateLimit(ctx, req.IPAddress)
	if err != nil {
		check.Status = domain.SecurityCheckStatusFailed
		check.Message = fmt.Sprintf("Rate limit check failed: %v", err)
		return check, err
	}

	check.Metadata = map[string]interface{}{
		"limit":     info.Limit,
		"remaining": info.Remaining,
		"reset":     info.Reset,
	}

	if !allowed {
		check.Status = domain.SecurityCheckStatusFailed
		check.Result = domain.SecurityCheckResultBlock
		check.Message = "Rate limit exceeded"
		
		// Log security event
		s.logSecurityEvent(ctx, req, domain.SecurityEventTypeMaliciousActivity, domain.SeverityMedium, "Rate limit exceeded")
	}

	check.Duration = time.Since(startTime)
	return check, nil
}

func (s *SecurityGatewayService) performWAFCheck(ctx context.Context, req *domain.SecurityRequest) (*domain.SecurityCheck, error) {
	startTime := time.Now()
	
	check := &domain.SecurityCheck{
		Name:     "waf",
		Type:     domain.SecurityCheckTypeWAF,
		Status:   domain.SecurityCheckStatusPassed,
		Result:   domain.SecurityCheckResultAllow,
		Duration: time.Since(startTime),
	}

	// Perform WAF check
	wafResult, err := s.wafService.CheckRequest(ctx, req)
	if err != nil {
		check.Status = domain.SecurityCheckStatusFailed
		check.Message = fmt.Sprintf("WAF check failed: %v", err)
		return check, err
	}

	check.Metadata = map[string]interface{}{
		"score":        wafResult.Score,
		"rule_matched": wafResult.RuleMatched,
		"details":      wafResult.Details,
	}

	if wafResult.Blocked {
		check.Status = domain.SecurityCheckStatusFailed
		check.Result = domain.SecurityCheckResultBlock
		check.Message = wafResult.Reason
		
		// Log security event
		s.logSecurityEvent(ctx, req, domain.SecurityEventTypeMaliciousActivity, domain.SeverityHigh, fmt.Sprintf("WAF block: %s", wafResult.Reason))
	}

	check.Duration = time.Since(startTime)
	return check, nil
}

func (s *SecurityGatewayService) performValidationCheck(ctx context.Context, req *domain.SecurityRequest) (*domain.SecurityCheck, error) {
	startTime := time.Now()
	
	check := &domain.SecurityCheck{
		Name:     "input_validation",
		Type:     domain.SecurityCheckTypeValidation,
		Status:   domain.SecurityCheckStatusPassed,
		Result:   domain.SecurityCheckResultAllow,
		Duration: time.Since(startTime),
	}

	// Validate URL
	urlResult := s.validationService.ValidateURL(req.URL)
	if !urlResult.IsValid {
		check.Status = domain.SecurityCheckStatusFailed
		check.Result = domain.SecurityCheckResultBlock
		check.Message = fmt.Sprintf("Invalid URL: %v", urlResult.Errors)
		check.Metadata = map[string]interface{}{
			"url_validation": urlResult,
		}
		
		// Log security event
		s.logSecurityEvent(ctx, req, domain.SecurityEventTypeMaliciousActivity, domain.SeverityMedium, "URL validation failed")
		
		check.Duration = time.Since(startTime)
		return check, nil
	}

	// Validate headers
	for name, value := range req.Headers {
		headerResult := s.validationService.ValidateInput(value)
		if !headerResult.IsValid {
			check.Status = domain.SecurityCheckStatusFailed
			check.Result = domain.SecurityCheckResultBlock
			check.Message = fmt.Sprintf("Invalid header %s: %v", name, headerResult.Errors)
			check.Metadata = map[string]interface{}{
				"header_validation": headerResult,
				"header_name":       name,
			}
			
			// Log security event
			s.logSecurityEvent(ctx, req, domain.SecurityEventTypeMaliciousActivity, domain.SeverityMedium, fmt.Sprintf("Header validation failed: %s", name))
			
			check.Duration = time.Since(startTime)
			return check, nil
		}
	}

	// Validate body if present
	if len(req.Body) > 0 {
		bodyResult := s.validationService.ValidateInput(string(req.Body))
		if !bodyResult.IsValid {
			check.Status = domain.SecurityCheckStatusFailed
			check.Result = domain.SecurityCheckResultBlock
			check.Message = fmt.Sprintf("Invalid request body: %v", bodyResult.Errors)
			check.Metadata = map[string]interface{}{
				"body_validation": bodyResult,
			}
			
			// Log security event
			s.logSecurityEvent(ctx, req, domain.SecurityEventTypeMaliciousActivity, domain.SeverityHigh, "Request body validation failed")
			
			check.Duration = time.Since(startTime)
			return check, nil
		}
	}

	check.Duration = time.Since(startTime)
	return check, nil
}

func (s *SecurityGatewayService) performAuthenticationCheck(ctx context.Context, req *domain.SecurityRequest) (*domain.SecurityCheck, error) {
	// Skip authentication for public endpoints
	if s.isPublicEndpoint(req.URL) {
		return nil, nil
	}

	startTime := time.Now()
	
	check := &domain.SecurityCheck{
		Name:     "authentication",
		Type:     domain.SecurityCheckTypeAuthentication,
		Status:   domain.SecurityCheckStatusPassed,
		Result:   domain.SecurityCheckResultAllow,
		Duration: time.Since(startTime),
	}

	// Check for Authorization header
	authHeader, exists := req.Headers["Authorization"]
	if !exists || authHeader == "" {
		check.Status = domain.SecurityCheckStatusFailed
		check.Result = domain.SecurityCheckResultBlock
		check.Message = "Missing authorization header"
		
		// Log security event
		s.logSecurityEvent(ctx, req, domain.SecurityEventTypeAuthentication, domain.SeverityMedium, "Missing authorization header")
		
		check.Duration = time.Since(startTime)
		return check, nil
	}

	// TODO: Implement JWT token validation
	// For now, just check if token is present
	if len(authHeader) < 10 {
		check.Status = domain.SecurityCheckStatusFailed
		check.Result = domain.SecurityCheckResultBlock
		check.Message = "Invalid authorization token"
		
		// Log security event
		s.logSecurityEvent(ctx, req, domain.SecurityEventTypeAuthentication, domain.SeverityMedium, "Invalid authorization token")
	}

	check.Duration = time.Since(startTime)
	return check, nil
}

func (s *SecurityGatewayService) performAuthorizationCheck(ctx context.Context, req *domain.SecurityRequest) (*domain.SecurityCheck, error) {
	// Skip authorization for public endpoints
	if s.isPublicEndpoint(req.URL) {
		return nil, nil
	}

	startTime := time.Now()
	
	check := &domain.SecurityCheck{
		Name:     "authorization",
		Type:     domain.SecurityCheckTypeAuthorization,
		Status:   domain.SecurityCheckStatusPassed,
		Result:   domain.SecurityCheckResultAllow,
		Duration: time.Since(startTime),
	}

	// TODO: Implement proper authorization logic
	// For now, just pass if authentication passed

	check.Duration = time.Since(startTime)
	return check, nil
}

// Helper methods

func (s *SecurityGatewayService) createBlockedResponse(req *domain.SecurityRequest, reason string, checks []domain.SecurityCheck, startTime time.Time) *domain.SecurityResponse {
	return &domain.SecurityResponse{
		RequestID:      req.ID,
		StatusCode:     http.StatusForbidden,
		Headers:        map[string]string{
			"X-Request-ID":     req.ID,
			"X-Correlation-ID": req.CorrelationID,
			"X-Block-Reason":   reason,
		},
		ProcessingTime: time.Since(startTime),
		Timestamp:      time.Now(),
		SecurityChecks: checks,
	}
}

func (s *SecurityGatewayService) isPublicEndpoint(url string) bool {
	publicEndpoints := []string{
		"/health",
		"/metrics",
		"/api/v1/auth/login",
		"/api/v1/auth/register",
	}

	for _, endpoint := range publicEndpoints {
		if url == endpoint {
			return true
		}
	}
	return false
}

func (s *SecurityGatewayService) logSecurityEvent(ctx context.Context, req *domain.SecurityRequest, eventType domain.SecurityEventType, severity domain.SecuritySeverity, description string) {
	event := &monitoring.SecurityEvent{
		EventType:   s.mapEventType(eventType),
		Severity:    s.mapSeverity(severity),
		Source:      "security-gateway",
		UserID:      req.UserID,
		IPAddress:   req.IPAddress,
		UserAgent:   req.UserAgent,
		Description: description,
		Metadata: map[string]interface{}{
			"request_id":     req.ID,
			"correlation_id": req.CorrelationID,
			"method":         req.Method,
			"url":            req.URL,
		},
	}

	s.monitoringService.LogSecurityEvent(ctx, event)
}

// mapEventType maps domain event types to monitoring event types
func (s *SecurityGatewayService) mapEventType(eventType domain.SecurityEventType) monitoring.SecurityEventType {
	switch eventType {
	case domain.SecurityEventTypeAuthentication:
		return monitoring.EventTypeAuthentication
	case domain.SecurityEventTypeAuthorization:
		return monitoring.EventTypeAuthorization
	case domain.SecurityEventTypeNetworkActivity:
		return monitoring.EventTypeNetworkActivity
	case domain.SecurityEventTypeMaliciousActivity:
		return monitoring.EventTypeMaliciousActivity
	case domain.SecurityEventTypePrivilegeEscalation:
		return monitoring.EventTypePrivilegeEscalation
	case domain.SecurityEventTypeDataAccess:
		return monitoring.EventTypeDataAccess
	case domain.SecurityEventTypeSystemAccess:
		return monitoring.EventTypeSystemAccess
	default:
		return monitoring.EventTypeNetworkActivity
	}
}

// mapSeverity maps domain severity to monitoring severity
func (s *SecurityGatewayService) mapSeverity(severity domain.SecuritySeverity) monitoring.SecuritySeverity {
	switch severity {
	case domain.SeverityInfo:
		return monitoring.SeverityInfo
	case domain.SeverityLow:
		return monitoring.SeverityLow
	case domain.SeverityMedium:
		return monitoring.SeverityMedium
	case domain.SeverityHigh:
		return monitoring.SeverityHigh
	case domain.SeverityCritical:
		return monitoring.SeverityCritical
	default:
		return monitoring.SeverityInfo
	}
}
