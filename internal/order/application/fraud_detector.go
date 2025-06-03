package application

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/order/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// MLFraudDetector implements machine learning-based fraud detection
type MLFraudDetector struct {
	paymentRepo PaymentRepository
	logger      *logger.Logger
	config      *FraudDetectionConfig
	riskProfiles map[string]*CustomerRiskProfile
}

// FraudDetectionConfig represents fraud detection configuration
type FraudDetectionConfig struct {
	EnableVelocityChecks    bool          `yaml:"enable_velocity_checks"`
	EnableAmountAnalysis    bool          `yaml:"enable_amount_analysis"`
	EnableLocationAnalysis  bool          `yaml:"enable_location_analysis"`
	EnableDeviceAnalysis    bool          `yaml:"enable_device_analysis"`
	EnableBehaviorAnalysis  bool          `yaml:"enable_behavior_analysis"`
	VelocityWindow          time.Duration `yaml:"velocity_window"`
	MaxTransactionsPerHour  int           `yaml:"max_transactions_per_hour"`
	SuspiciousAmountMultiplier float64    `yaml:"suspicious_amount_multiplier"`
	MaxDistanceKm           float64       `yaml:"max_distance_km"`
	MinTimeBetweenLocations time.Duration `yaml:"min_time_between_locations"`
}

// CustomerRiskProfile represents a customer's risk profile
type CustomerRiskProfile struct {
	CustomerID           string                 `json:"customer_id"`
	BaseRiskScore        float64                `json:"base_risk_score"`
	TransactionHistory   []PaymentPattern       `json:"transaction_history"`
	LocationHistory      []LocationPattern      `json:"location_history"`
	DeviceHistory        []DevicePattern        `json:"device_history"`
	BehaviorProfile      *BehaviorProfile       `json:"behavior_profile"`
	LastUpdated          time.Time              `json:"last_updated"`
	SuspiciousActivities []SuspiciousActivity   `json:"suspicious_activities"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// LocationPattern represents location usage patterns
type LocationPattern struct {
	Country   string    `json:"country"`
	Region    string    `json:"region"`
	City      string    `json:"city"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Frequency int       `json:"frequency"`
	LastUsed  time.Time `json:"last_used"`
	Trusted   bool      `json:"trusted"`
}

// DevicePattern represents device usage patterns
type DevicePattern struct {
	DeviceID    string    `json:"device_id"`
	UserAgent   string    `json:"user_agent"`
	Fingerprint string    `json:"fingerprint"`
	Frequency   int       `json:"frequency"`
	LastUsed    time.Time `json:"last_used"`
	Trusted     bool      `json:"trusted"`
}

// SuspiciousActivity represents a suspicious activity
type SuspiciousActivity struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewMLFraudDetector creates a new ML-based fraud detector
func NewMLFraudDetector(
	paymentRepo PaymentRepository,
	config *FraudDetectionConfig,
	logger *logger.Logger,
) *MLFraudDetector {
	return &MLFraudDetector{
		paymentRepo:  paymentRepo,
		config:       config,
		logger:       logger,
		riskProfiles: make(map[string]*CustomerRiskProfile),
	}
}

// AnalyzePayment performs comprehensive fraud analysis on a payment
func (f *MLFraudDetector) AnalyzePayment(ctx context.Context, payment *domain.Payment, metadata PaymentMetadata) (*FraudAnalysis, error) {
	analysis := &FraudAnalysis{
		RiskScore:       0.0,
		RiskLevel:       RiskLevelLow,
		Indicators:      []FraudIndicator{},
		Recommendations: []string{},
		ShouldBlock:     false,
		RequiresMFA:     false,
		RequiresReview:  false,
		Confidence:      0.0,
		Metadata:        make(map[string]interface{}),
	}

	// Get customer risk profile
	profile, err := f.getOrCreateRiskProfile(ctx, payment.CustomerID)
	if err != nil {
		f.logger.WithError(err).Error("Failed to get risk profile")
		// Continue with analysis using default profile
		profile = f.createDefaultRiskProfile(payment.CustomerID)
	}

	// Perform various fraud checks
	checks := []func(context.Context, *domain.Payment, PaymentMetadata, *CustomerRiskProfile, *FraudAnalysis) error{
		f.checkVelocity,
		f.checkAmount,
		f.checkLocation,
		f.checkDevice,
		f.checkBehavior,
		f.checkCardTesting,
		f.checkStolenCard,
	}

	for _, check := range checks {
		if err := check(ctx, payment, metadata, profile, analysis); err != nil {
			f.logger.WithError(err).Error("Fraud check failed")
			// Continue with other checks
		}
	}

	// Calculate final risk score and level
	f.calculateFinalRisk(analysis)

	// Determine actions based on risk
	f.determineActions(analysis)

	// Update analysis metadata
	analysis.Metadata["customer_id"] = payment.CustomerID
	analysis.Metadata["payment_id"] = payment.ID
	analysis.Metadata["analysis_time"] = time.Now()
	analysis.Metadata["profile_age"] = time.Since(profile.LastUpdated).Hours()

	return analysis, nil
}

// Velocity check - detects rapid-fire transactions
func (f *MLFraudDetector) checkVelocity(ctx context.Context, payment *domain.Payment, metadata PaymentMetadata, profile *CustomerRiskProfile, analysis *FraudAnalysis) error {
	if !f.config.EnableVelocityChecks {
		return nil
	}

	// Count recent transactions
	recentCount := 0
	cutoff := time.Now().Add(-f.config.VelocityWindow)
	
	for _, txn := range profile.TransactionHistory {
		if txn.Timestamp.After(cutoff) {
			recentCount++
		}
	}

	if recentCount > f.config.MaxTransactionsPerHour {
		indicator := FraudIndicator{
			Type:        FraudIndicatorVelocity,
			Severity:    "high",
			Description: fmt.Sprintf("High transaction velocity: %d transactions in %v", recentCount, f.config.VelocityWindow),
			Value:       recentCount,
			Threshold:   f.config.MaxTransactionsPerHour,
		}
		analysis.Indicators = append(analysis.Indicators, indicator)
		analysis.RiskScore += 0.3
	}

	return nil
}

// Amount check - detects unusual amounts
func (f *MLFraudDetector) checkAmount(ctx context.Context, payment *domain.Payment, metadata PaymentMetadata, profile *CustomerRiskProfile, analysis *FraudAnalysis) error {
	if !f.config.EnableAmountAnalysis {
		return nil
	}

	// Calculate typical amount from history
	if len(profile.TransactionHistory) > 0 {
		var totalAmount int64
		for _, txn := range profile.TransactionHistory {
			totalAmount += txn.Amount
		}
		avgAmount := float64(totalAmount) / float64(len(profile.TransactionHistory))
		
		// Check if current amount is significantly higher
		if float64(payment.Amount) > avgAmount*f.config.SuspiciousAmountMultiplier {
			indicator := FraudIndicator{
				Type:        FraudIndicatorAmount,
				Severity:    "medium",
				Description: fmt.Sprintf("Amount significantly higher than typical: %d vs avg %0.2f", payment.Amount, avgAmount),
				Value:       payment.Amount,
				Threshold:   avgAmount * f.config.SuspiciousAmountMultiplier,
			}
			analysis.Indicators = append(analysis.Indicators, indicator)
			analysis.RiskScore += 0.2
		}
	}

	// Check for round amounts (often used in fraud)
	if payment.Amount%10000 == 0 && payment.Amount >= 10000 { // Round amounts >= $100
		indicator := FraudIndicator{
			Type:        FraudIndicatorAmount,
			Severity:    "low",
			Description: "Round amount transaction",
			Value:       payment.Amount,
		}
		analysis.Indicators = append(analysis.Indicators, indicator)
		analysis.RiskScore += 0.1
	}

	return nil
}

// Location check - detects impossible travel
func (f *MLFraudDetector) checkLocation(ctx context.Context, payment *domain.Payment, metadata PaymentMetadata, profile *CustomerRiskProfile, analysis *FraudAnalysis) error {
	if !f.config.EnableLocationAnalysis || metadata.GeoLocation == nil {
		return nil
	}

	currentLocation := metadata.GeoLocation

	// Check against known locations
	isKnownLocation := false
	for _, loc := range profile.LocationHistory {
		distance := f.calculateDistance(
			currentLocation.Latitude, currentLocation.Longitude,
			loc.Latitude, loc.Longitude,
		)
		
		if distance < 50 { // Within 50km of known location
			isKnownLocation = true
			break
		}
	}

	if !isKnownLocation && len(profile.LocationHistory) > 0 {
		// Check for impossible travel
		lastLocation := profile.LocationHistory[len(profile.LocationHistory)-1]
		distance := f.calculateDistance(
			currentLocation.Latitude, currentLocation.Longitude,
			lastLocation.Latitude, lastLocation.Longitude,
		)
		
		timeDiff := time.Since(lastLocation.LastUsed)
		maxPossibleDistance := timeDiff.Hours() * 900 // Assume max 900 km/h (commercial flight)
		
		if distance > maxPossibleDistance && distance > f.config.MaxDistanceKm {
			indicator := FraudIndicator{
				Type:        FraudIndicatorLocation,
				Severity:    "high",
				Description: fmt.Sprintf("Impossible travel: %.2f km in %v", distance, timeDiff),
				Value:       distance,
				Threshold:   maxPossibleDistance,
			}
			analysis.Indicators = append(analysis.Indicators, indicator)
			analysis.RiskScore += 0.4
		}
	}

	// Check for VPN usage
	if currentLocation.VPNDetected {
		indicator := FraudIndicator{
			Type:        FraudIndicatorLocation,
			Severity:    "medium",
			Description: "VPN or proxy detected",
			Value:       true,
		}
		analysis.Indicators = append(analysis.Indicators, indicator)
		analysis.RiskScore += 0.15
	}

	return nil
}

// Device check - detects new or suspicious devices
func (f *MLFraudDetector) checkDevice(ctx context.Context, payment *domain.Payment, metadata PaymentMetadata, profile *CustomerRiskProfile, analysis *FraudAnalysis) error {
	if !f.config.EnableDeviceAnalysis {
		return nil
	}

	// Check if device is known
	isKnownDevice := false
	for _, device := range profile.DeviceHistory {
		if device.DeviceID == metadata.DeviceID || device.Fingerprint == metadata.Headers["X-Device-Fingerprint"] {
			isKnownDevice = true
			break
		}
	}

	if !isKnownDevice {
		indicator := FraudIndicator{
			Type:        FraudIndicatorDevice,
			Severity:    "medium",
			Description: "New device detected",
			Value:       metadata.DeviceID,
		}
		analysis.Indicators = append(analysis.Indicators, indicator)
		analysis.RiskScore += 0.2
	}

	// Check for suspicious user agents
	suspiciousAgents := []string{"curl", "wget", "python", "bot", "crawler", "scraper"}
	userAgent := strings.ToLower(metadata.UserAgent)
	
	for _, suspicious := range suspiciousAgents {
		if strings.Contains(userAgent, suspicious) {
			indicator := FraudIndicator{
				Type:        FraudIndicatorDevice,
				Severity:    "high",
				Description: "Suspicious user agent detected",
				Value:       metadata.UserAgent,
			}
			analysis.Indicators = append(analysis.Indicators, indicator)
			analysis.RiskScore += 0.3
			break
		}
	}

	return nil
}

// Behavior check - detects unusual behavior patterns
func (f *MLFraudDetector) checkBehavior(ctx context.Context, payment *domain.Payment, metadata PaymentMetadata, profile *CustomerRiskProfile, analysis *FraudAnalysis) error {
	if !f.config.EnableBehaviorAnalysis || profile.BehaviorProfile == nil {
		return nil
	}

	behavior := profile.BehaviorProfile

	// Check time patterns
	currentHour := time.Now().Hour()
	if behavior.TypicalTime != 0 {
		typicalHour := int(behavior.TypicalTime.Hours()) % 24
		hourDiff := int(math.Abs(float64(currentHour - typicalHour)))
		if hourDiff > 12 {
			hourDiff = 24 - hourDiff
		}
		
		if hourDiff > 6 { // More than 6 hours from typical time
			indicator := FraudIndicator{
				Type:        FraudIndicatorBehavior,
				Severity:    "low",
				Description: "Transaction at unusual time",
				Value:       currentHour,
				Threshold:   typicalHour,
			}
			analysis.Indicators = append(analysis.Indicators, indicator)
			analysis.RiskScore += 0.1
		}
	}

	// Check payment method patterns
	isPreferredMethod := false
	for _, method := range behavior.PreferredMethods {
		if method == string(payment.Method) {
			isPreferredMethod = true
			break
		}
	}

	if !isPreferredMethod && len(behavior.PreferredMethods) > 0 {
		indicator := FraudIndicator{
			Type:        FraudIndicatorBehavior,
			Severity:    "low",
			Description: "Unusual payment method",
			Value:       payment.Method,
		}
		analysis.Indicators = append(analysis.Indicators, indicator)
		analysis.RiskScore += 0.1
	}

	return nil
}

// Card testing check - detects card testing patterns
func (f *MLFraudDetector) checkCardTesting(ctx context.Context, payment *domain.Payment, metadata PaymentMetadata, profile *CustomerRiskProfile, analysis *FraudAnalysis) error {
	// Look for patterns of small amounts followed by larger amounts
	recentSmallAmounts := 0
	cutoff := time.Now().Add(-1 * time.Hour)
	
	for _, txn := range profile.TransactionHistory {
		if txn.Timestamp.After(cutoff) && txn.Amount < 500 { // Less than $5
			recentSmallAmounts++
		}
	}

	if recentSmallAmounts >= 3 && payment.Amount > 5000 { // More than $50
		indicator := FraudIndicator{
			Type:        FraudIndicatorCardTesting,
			Severity:    "high",
			Description: "Possible card testing pattern detected",
			Value:       recentSmallAmounts,
			Threshold:   3,
		}
		analysis.Indicators = append(analysis.Indicators, indicator)
		analysis.RiskScore += 0.4
	}

	return nil
}

// Stolen card check - detects stolen card patterns
func (f *MLFraudDetector) checkStolenCard(ctx context.Context, payment *domain.Payment, metadata PaymentMetadata, profile *CustomerRiskProfile, analysis *FraudAnalysis) error {
	// Check for multiple failed attempts followed by success
	recentFailures := 0
	cutoff := time.Now().Add(-30 * time.Minute)
	
	for _, txn := range profile.TransactionHistory {
		if txn.Timestamp.After(cutoff) && !txn.Success {
			recentFailures++
		}
	}

	if recentFailures >= 3 {
		indicator := FraudIndicator{
			Type:        FraudIndicatorStolenCard,
			Severity:    "high",
			Description: "Multiple failed attempts before success",
			Value:       recentFailures,
			Threshold:   3,
		}
		analysis.Indicators = append(analysis.Indicators, indicator)
		analysis.RiskScore += 0.3
	}

	return nil
}

// Calculate final risk score and level
func (f *MLFraudDetector) calculateFinalRisk(analysis *FraudAnalysis) {
	// Normalize risk score
	if analysis.RiskScore > 1.0 {
		analysis.RiskScore = 1.0
	}

	// Determine risk level
	switch {
	case analysis.RiskScore >= 0.8:
		analysis.RiskLevel = RiskLevelCritical
	case analysis.RiskScore >= 0.6:
		analysis.RiskLevel = RiskLevelHigh
	case analysis.RiskScore >= 0.4:
		analysis.RiskLevel = RiskLevelMedium
	default:
		analysis.RiskLevel = RiskLevelLow
	}

	// Calculate confidence based on number of indicators
	indicatorCount := len(analysis.Indicators)
	analysis.Confidence = math.Min(float64(indicatorCount)*0.2, 1.0)
}

// Determine actions based on risk analysis
func (f *MLFraudDetector) determineActions(analysis *FraudAnalysis) {
	switch analysis.RiskLevel {
	case RiskLevelCritical:
		analysis.ShouldBlock = true
		analysis.RequiresReview = true
		analysis.Recommendations = append(analysis.Recommendations, "Block transaction immediately", "Manual review required")
	case RiskLevelHigh:
		analysis.RequiresMFA = true
		analysis.RequiresReview = true
		analysis.Recommendations = append(analysis.Recommendations, "Require additional authentication", "Flag for review")
	case RiskLevelMedium:
		analysis.RequiresMFA = true
		analysis.Recommendations = append(analysis.Recommendations, "Require additional authentication")
	case RiskLevelLow:
		analysis.Recommendations = append(analysis.Recommendations, "Allow transaction")
	}
}

// Helper methods

func (f *MLFraudDetector) getOrCreateRiskProfile(ctx context.Context, customerID string) (*CustomerRiskProfile, error) {
	// Check in-memory cache first
	if profile, exists := f.riskProfiles[customerID]; exists {
		return profile, nil
	}

	// TODO: Load from persistent storage
	// For now, create a new profile
	profile := f.createDefaultRiskProfile(customerID)
	f.riskProfiles[customerID] = profile
	
	return profile, nil
}

func (f *MLFraudDetector) createDefaultRiskProfile(customerID string) *CustomerRiskProfile {
	return &CustomerRiskProfile{
		CustomerID:           customerID,
		BaseRiskScore:        0.1, // Low default risk
		TransactionHistory:   []PaymentPattern{},
		LocationHistory:      []LocationPattern{},
		DeviceHistory:        []DevicePattern{},
		BehaviorProfile:      &BehaviorProfile{TrustScore: 0.5},
		LastUpdated:          time.Now(),
		SuspiciousActivities: []SuspiciousActivity{},
		Metadata:             make(map[string]interface{}),
	}
}

func (f *MLFraudDetector) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Haversine formula for calculating distance between two points
	const R = 6371 // Earth's radius in kilometers

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// UpdateRiskProfile updates the customer's risk profile
func (f *MLFraudDetector) UpdateRiskProfile(ctx context.Context, customerID string, result *FraudAnalysis) error {
	profile, err := f.getOrCreateRiskProfile(ctx, customerID)
	if err != nil {
		return err
	}

	// Update base risk score based on analysis
	if result.RiskScore > profile.BaseRiskScore {
		profile.BaseRiskScore = (profile.BaseRiskScore + result.RiskScore) / 2
	}

	// Add suspicious activities if any
	if len(result.Indicators) > 0 {
		activity := SuspiciousActivity{
			Type:        "fraud_analysis",
			Description: fmt.Sprintf("Risk level: %s, Score: %.2f", result.RiskLevel, result.RiskScore),
			Severity:    string(result.RiskLevel),
			Timestamp:   time.Now(),
			Metadata:    result.Metadata,
		}
		profile.SuspiciousActivities = append(profile.SuspiciousActivities, activity)
	}

	profile.LastUpdated = time.Now()
	f.riskProfiles[customerID] = profile

	return nil
}

// GetCustomerRiskScore returns the customer's current risk score
func (f *MLFraudDetector) GetCustomerRiskScore(ctx context.Context, customerID string) (float64, error) {
	profile, err := f.getOrCreateRiskProfile(ctx, customerID)
	if err != nil {
		return 0.0, err
	}

	return profile.BaseRiskScore, nil
}
