package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/order/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/DimaJoyti/go-coffee/pkg/security/encryption"
	"github.com/DimaJoyti/go-coffee/pkg/security/monitoring"
	"github.com/DimaJoyti/go-coffee/pkg/security/validation"
)

// SecurePaymentService provides enhanced payment processing with security features
type SecurePaymentService struct {
	paymentRepo       PaymentRepository
	orderRepo         OrderRepository
	eventPublisher    EventPublisher
	paymentProcessor  PaymentProcessor
	cryptoProcessor   CryptoPaymentProcessor
	loyaltyService    LoyaltyService
	encryptionService *encryption.EncryptionService
	validationService *validation.ValidationService
	monitoringService *monitoring.SecurityMonitoringService
	fraudDetector     FraudDetector
	logger            *logger.Logger
	config            *SecurePaymentConfig
}

// SecurePaymentConfig represents secure payment configuration
type SecurePaymentConfig struct {
	EnableFraudDetection     bool          `yaml:"enable_fraud_detection"`
	EnableEncryption         bool          `yaml:"enable_encryption"`
	EnableRealTimeMonitoring bool          `yaml:"enable_real_time_monitoring"`
	MaxPaymentAmount         int64         `yaml:"max_payment_amount"`
	RequireMFAAmount         int64         `yaml:"require_mfa_amount"`
	SuspiciousAmountThreshold int64        `yaml:"suspicious_amount_threshold"`
	PaymentTimeout           time.Duration `yaml:"payment_timeout"`
	EnablePCICompliance      bool          `yaml:"enable_pci_compliance"`
	AllowedCountries         []string      `yaml:"allowed_countries"`
	BlockedCountries         []string      `yaml:"blocked_countries"`
}

// FraudDetector interface for fraud detection
type FraudDetector interface {
	AnalyzePayment(ctx context.Context, payment *domain.Payment, metadata PaymentMetadata) (*FraudAnalysis, error)
	UpdateRiskProfile(ctx context.Context, customerID string, result *FraudAnalysis) error
	GetCustomerRiskScore(ctx context.Context, customerID string) (float64, error)
}

// PaymentMetadata contains additional payment context for security analysis
type PaymentMetadata struct {
	IPAddress       string            `json:"ip_address"`
	UserAgent       string            `json:"user_agent"`
	DeviceID        string            `json:"device_id"`
	SessionID       string            `json:"session_id"`
	GeoLocation     *GeoLocation      `json:"geo_location,omitempty"`
	PaymentHistory  []PaymentPattern  `json:"payment_history,omitempty"`
	BehaviorProfile *BehaviorProfile  `json:"behavior_profile,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
}

// GeoLocation represents geographical location
type GeoLocation struct {
	Country     string  `json:"country"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ISP         string  `json:"isp"`
	VPNDetected bool    `json:"vpn_detected"`
}

// PaymentPattern represents payment behavior patterns
type PaymentPattern struct {
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	Method        string    `json:"method"`
	Timestamp     time.Time `json:"timestamp"`
	MerchantType  string    `json:"merchant_type"`
	Success       bool      `json:"success"`
	IPAddress     string    `json:"ip_address"`
	DeviceID      string    `json:"device_id"`
}

// BehaviorProfile represents user behavior profile
type BehaviorProfile struct {
	TypicalAmount       int64         `json:"typical_amount"`
	TypicalTime         time.Duration `json:"typical_time"`
	PreferredMethods    []string      `json:"preferred_methods"`
	FrequentLocations   []string      `json:"frequent_locations"`
	AverageSessionTime  time.Duration `json:"average_session_time"`
	LastActivityTime    time.Time     `json:"last_activity_time"`
	SuspiciousActivity  int           `json:"suspicious_activity"`
	TrustScore          float64       `json:"trust_score"`
}

// FraudAnalysis represents the result of fraud analysis
type FraudAnalysis struct {
	RiskScore       float64                `json:"risk_score"`
	RiskLevel       RiskLevel              `json:"risk_level"`
	Indicators      []FraudIndicator       `json:"indicators"`
	Recommendations []string               `json:"recommendations"`
	ShouldBlock     bool                   `json:"should_block"`
	RequiresMFA     bool                   `json:"requires_mfa"`
	RequiresReview  bool                   `json:"requires_review"`
	Confidence      float64                `json:"confidence"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// RiskLevel represents the risk level
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// FraudIndicator represents a fraud indicator
type FraudIndicator struct {
	Type        FraudIndicatorType `json:"type"`
	Severity    string             `json:"severity"`
	Description string             `json:"description"`
	Value       interface{}        `json:"value"`
	Threshold   interface{}        `json:"threshold"`
}

// FraudIndicatorType represents the type of fraud indicator
type FraudIndicatorType string

const (
	FraudIndicatorVelocity        FraudIndicatorType = "velocity"
	FraudIndicatorAmount          FraudIndicatorType = "amount"
	FraudIndicatorLocation        FraudIndicatorType = "location"
	FraudIndicatorDevice          FraudIndicatorType = "device"
	FraudIndicatorBehavior        FraudIndicatorType = "behavior"
	FraudIndicatorCardTesting     FraudIndicatorType = "card_testing"
	FraudIndicatorStolenCard      FraudIndicatorType = "stolen_card"
	FraudIndicatorSyntheticID     FraudIndicatorType = "synthetic_id"
)

// SecureCreatePaymentRequest extends the original request with security fields
type SecureCreatePaymentRequest struct {
	*CreatePaymentRequest
	Metadata         PaymentMetadata `json:"metadata"`
	RequireMFA       bool            `json:"require_mfa"`
	MFAToken         string          `json:"mfa_token,omitempty"`
	DeviceFingerprint string         `json:"device_fingerprint"`
	SecurityToken    string          `json:"security_token"`
}

// SecureCreatePaymentResponse extends the original response with security fields
type SecureCreatePaymentResponse struct {
	*CreatePaymentResponse
	SecurityCheck   *SecurityCheckResult `json:"security_check"`
	FraudAnalysis   *FraudAnalysis       `json:"fraud_analysis,omitempty"`
	RequiresMFA     bool                 `json:"requires_mfa"`
	MFAChallenge    *MFAChallenge        `json:"mfa_challenge,omitempty"`
	RiskScore       float64              `json:"risk_score"`
	SecurityToken   string               `json:"security_token"`
}

// SecurityCheckResult represents the result of security checks
type SecurityCheckResult struct {
	Passed      bool                   `json:"passed"`
	Checks      []SecurityCheck        `json:"checks"`
	Warnings    []string               `json:"warnings,omitempty"`
	BlockReason string                 `json:"block_reason,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityCheck represents a single security check
type SecurityCheck struct {
	Name     string                 `json:"name"`
	Status   SecurityCheckStatus    `json:"status"`
	Message  string                 `json:"message,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SecurityCheckStatus represents the status of a security check
type SecurityCheckStatus string

const (
	SecurityCheckStatusPassed  SecurityCheckStatus = "passed"
	SecurityCheckStatusFailed  SecurityCheckStatus = "failed"
	SecurityCheckStatusWarning SecurityCheckStatus = "warning"
)

// MFAChallenge represents an MFA challenge
type MFAChallenge struct {
	Type        string    `json:"type"`
	Challenge   string    `json:"challenge"`
	ExpiresAt   time.Time `json:"expires_at"`
	Attempts    int       `json:"attempts"`
	MaxAttempts int       `json:"max_attempts"`
}

// NewSecurePaymentService creates a new secure payment service
func NewSecurePaymentService(
	paymentRepo PaymentRepository,
	orderRepo OrderRepository,
	eventPublisher EventPublisher,
	paymentProcessor PaymentProcessor,
	cryptoProcessor CryptoPaymentProcessor,
	loyaltyService LoyaltyService,
	encryptionService *encryption.EncryptionService,
	validationService *validation.ValidationService,
	monitoringService *monitoring.SecurityMonitoringService,
	fraudDetector FraudDetector,
	config *SecurePaymentConfig,
	logger *logger.Logger,
) *SecurePaymentService {
	return &SecurePaymentService{
		paymentRepo:       paymentRepo,
		orderRepo:         orderRepo,
		eventPublisher:    eventPublisher,
		paymentProcessor:  paymentProcessor,
		cryptoProcessor:   cryptoProcessor,
		loyaltyService:    loyaltyService,
		encryptionService: encryptionService,
		validationService: validationService,
		monitoringService: monitoringService,
		fraudDetector:     fraudDetector,
		config:            config,
		logger:            logger,
	}
}

// CreateSecurePayment creates a payment with enhanced security checks
func (s *SecurePaymentService) CreateSecurePayment(ctx context.Context, req *SecureCreatePaymentRequest) (*SecureCreatePaymentResponse, error) {
	startTime := time.Now()
	
	// Get order for logging
	order, _ := s.orderRepo.GetByID(ctx, req.OrderID)
	amount := int64(0)
	if order != nil {
		amount = order.TotalAmount
	}

	// Log security event
	s.logSecurityEvent(ctx, "payment_creation_started", monitoring.SeverityInfo, map[string]interface{}{
		"order_id":     req.OrderID,
		"ip_address":   req.Metadata.IPAddress,
		"amount":       amount,
		"method":       req.PaymentMethod,
	})

	// 1. Input validation
	securityCheck := &SecurityCheckResult{
		Passed: true,
		Checks: []SecurityCheck{},
	}

	if err := s.validatePaymentRequest(ctx, req, securityCheck); err != nil {
		return nil, fmt.Errorf("payment validation failed: %w", err)
	}

	// 2. Fraud detection
	var fraudAnalysis *FraudAnalysis
	if s.config.EnableFraudDetection {
		// Get the order first to create payment for fraud analysis
		order, err := s.orderRepo.GetByID(ctx, req.OrderID)
		if err != nil {
			return nil, fmt.Errorf("failed to get order: %w", err)
		}

		// Create temporary payment for fraud analysis
		tempPayment, err := domain.NewPayment(req.OrderID, order.CustomerID, order.TotalAmount, order.Currency, req.PaymentMethod)
		if err != nil {
			return nil, fmt.Errorf("failed to create payment for analysis: %w", err)
		}

		fraudAnalysis, err = s.fraudDetector.AnalyzePayment(ctx, tempPayment, req.Metadata)
		if err != nil {
			s.logger.WithError(err).Error("Fraud analysis failed")
			// Continue with payment but log the error
		} else {
			// Check if payment should be blocked
			if fraudAnalysis.ShouldBlock {
				s.logSecurityEvent(ctx, "payment_blocked_fraud", monitoring.SeverityHigh, map[string]interface{}{
					"order_id":    req.OrderID,
					"risk_score":  fraudAnalysis.RiskScore,
					"risk_level":  fraudAnalysis.RiskLevel,
					"indicators":  fraudAnalysis.Indicators,
				})

				securityCheck.Passed = false
				securityCheck.BlockReason = "Payment blocked due to fraud detection"
				
				return &SecureCreatePaymentResponse{
					SecurityCheck: securityCheck,
					FraudAnalysis: fraudAnalysis,
					RiskScore:     fraudAnalysis.RiskScore,
				}, nil
			}

			// Check if MFA is required
			if fraudAnalysis.RequiresMFA || s.requiresMFA(req) {
				if req.MFAToken == "" {
					mfaChallenge := s.generateMFAChallenge(ctx, order.CustomerID)
					
					return &SecureCreatePaymentResponse{
						SecurityCheck: securityCheck,
						FraudAnalysis: fraudAnalysis,
						RequiresMFA:   true,
						MFAChallenge:  mfaChallenge,
						RiskScore:     fraudAnalysis.RiskScore,
					}, nil
				}

				// Verify MFA token
				if !s.verifyMFAToken(ctx, order.CustomerID, req.MFAToken) {
					s.logSecurityEvent(ctx, "payment_mfa_failed", monitoring.SeverityMedium, map[string]interface{}{
						"order_id":    req.OrderID,
						"customer_id": order.CustomerID,
					})

					securityCheck.Passed = false
					securityCheck.BlockReason = "MFA verification failed"
					
					return &SecureCreatePaymentResponse{
						SecurityCheck: securityCheck,
						FraudAnalysis: fraudAnalysis,
						RiskScore:     fraudAnalysis.RiskScore,
					}, nil
				}
			}
		}
	}

	// 3. Create payment using original service logic
	originalReq := req.CreatePaymentRequest
	originalResp, err := s.createPaymentInternal(ctx, originalReq)
	if err != nil {
		s.logSecurityEvent(ctx, "payment_creation_failed", monitoring.SeverityHigh, map[string]interface{}{
			"order_id": req.OrderID,
			"error":    err.Error(),
		})
		return nil, err
	}

	// 4. Generate security token for payment tracking
	securityToken, err := s.generateSecurityToken()
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate security token")
	}

	// 5. Log successful payment creation
	s.logSecurityEvent(ctx, "payment_created_successfully", monitoring.SeverityInfo, map[string]interface{}{
		"payment_id":      originalResp.PaymentID,
		"order_id":        req.OrderID,
		"amount":          originalResp.Amount,
		"processing_time": time.Since(startTime).Milliseconds(),
		"risk_score":      getRiskScore(fraudAnalysis),
	})

	// 6. Update fraud profile if enabled
	if s.config.EnableFraudDetection && fraudAnalysis != nil {
		order, _ := s.orderRepo.GetByID(ctx, req.OrderID)
		if order != nil {
			s.fraudDetector.UpdateRiskProfile(ctx, order.CustomerID, fraudAnalysis)
		}
	}

	return &SecureCreatePaymentResponse{
		CreatePaymentResponse: originalResp,
		SecurityCheck:         securityCheck,
		FraudAnalysis:         fraudAnalysis,
		RequiresMFA:           false,
		RiskScore:             getRiskScore(fraudAnalysis),
		SecurityToken:         securityToken,
	}, nil
}

// Helper methods

func (s *SecurePaymentService) validatePaymentRequest(ctx context.Context, req *SecureCreatePaymentRequest, securityCheck *SecurityCheckResult) error {
	checks := []SecurityCheck{}

	// Validate request structure
	if req.CreatePaymentRequest == nil {
		checks = append(checks, SecurityCheck{
			Name:    "request_validation",
			Status:  SecurityCheckStatusFailed,
			Message: "Invalid payment request structure",
		})
		securityCheck.Checks = checks
		securityCheck.Passed = false
		return errors.New("invalid payment request: missing required fields")
	}

	// Get order to validate amount
	order, err := s.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		checks = append(checks, SecurityCheck{
			Name:    "order_validation",
			Status:  SecurityCheckStatusFailed,
			Message: "Failed to retrieve order for validation",
		})
	} else {
		// Validate amount
		if order.TotalAmount <= 0 {
			checks = append(checks, SecurityCheck{
				Name:    "amount_validation",
				Status:  SecurityCheckStatusFailed,
				Message: "Invalid payment amount",
			})
		} else if order.TotalAmount > s.config.MaxPaymentAmount {
			checks = append(checks, SecurityCheck{
				Name:    "amount_limit",
				Status:  SecurityCheckStatusFailed,
				Message: "Payment amount exceeds maximum allowed",
			})
		} else {
			checks = append(checks, SecurityCheck{
				Name:   "amount_validation",
				Status: SecurityCheckStatusPassed,
			})
		}
	}

	// Validate IP address
	if req.Metadata.IPAddress != "" {
		ipResult := s.validationService.ValidateIP(req.Metadata.IPAddress)
		if !ipResult.IsValid {
			checks = append(checks, SecurityCheck{
				Name:    "ip_validation",
				Status:  SecurityCheckStatusFailed,
				Message: "Invalid IP address",
			})
		} else {
			checks = append(checks, SecurityCheck{
				Name:   "ip_validation",
				Status: SecurityCheckStatusPassed,
			})
		}
	}

	// Check geo-restrictions
	if req.Metadata.GeoLocation != nil {
		if s.isCountryBlocked(req.Metadata.GeoLocation.Country) {
			checks = append(checks, SecurityCheck{
				Name:    "geo_restriction",
				Status:  SecurityCheckStatusFailed,
				Message: "Payment not allowed from this country",
			})
		} else {
			checks = append(checks, SecurityCheck{
				Name:   "geo_restriction",
				Status: SecurityCheckStatusPassed,
			})
		}
	}

	securityCheck.Checks = checks

	// Check if any critical checks failed
	for _, check := range checks {
		if check.Status == SecurityCheckStatusFailed {
			securityCheck.Passed = false
			return errors.New("security validation failed")
		}
	}

	return nil
}

func (s *SecurePaymentService) requiresMFA(req *SecureCreatePaymentRequest) bool {
	// Get order to check amount
	order, err := s.orderRepo.GetByID(context.Background(), req.OrderID)
	if err != nil {
		// If we can't get the order, err on the side of caution and require MFA
		return true
	}
	return order.TotalAmount >= s.config.RequireMFAAmount || req.RequireMFA
}

func (s *SecurePaymentService) generateMFAChallenge(ctx context.Context, customerID string) *MFAChallenge {
	// Generate random challenge
	challenge := make([]byte, 16)
	rand.Read(challenge)
	
	return &MFAChallenge{
		Type:        "totp",
		Challenge:   hex.EncodeToString(challenge),
		ExpiresAt:   time.Now().Add(5 * time.Minute),
		Attempts:    0,
		MaxAttempts: 3,
	}
}

func (s *SecurePaymentService) verifyMFAToken(ctx context.Context, customerID, token string) bool {
	// TODO: Implement actual MFA verification
	// This would integrate with the auth service
	return token != ""
}

func (s *SecurePaymentService) generateSecurityToken() (string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}

func (s *SecurePaymentService) isCountryBlocked(country string) bool {
	for _, blocked := range s.config.BlockedCountries {
		if blocked == country {
			return true
		}
	}
	return false
}

func (s *SecurePaymentService) logSecurityEvent(ctx context.Context, eventType string, severity monitoring.SecuritySeverity, metadata map[string]interface{}) {
	event := &monitoring.SecurityEvent{
		EventType:   monitoring.EventTypeDataAccess,
		Severity:    severity,
		Source:      "secure-payment-service",
		Description: eventType,
		Metadata:    metadata,
	}

	s.monitoringService.LogSecurityEvent(ctx, event)
}

func (s *SecurePaymentService) createPaymentInternal(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// This would call the original payment service logic
	// For now, we'll create a basic implementation
	
	order, err := s.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	payment, err := domain.NewPayment(req.OrderID, order.CustomerID, order.TotalAmount, order.Currency, req.PaymentMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	return &CreatePaymentResponse{
		PaymentID: payment.ID,
		Status:    payment.Status.String(),
		Amount:    payment.Amount,
		CreatedAt: payment.CreatedAt,
	}, nil
}

func getRiskScore(analysis *FraudAnalysis) float64 {
	if analysis == nil {
		return 0.0
	}
	return analysis.RiskScore
}
