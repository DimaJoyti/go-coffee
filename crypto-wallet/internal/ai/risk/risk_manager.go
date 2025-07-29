package risk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// RiskManager is the central AI-powered risk management system
type RiskManager struct {
	logger *logger.Logger
	config RiskManagerConfig

	// AI Components
	transactionRiskScorer *TransactionRiskScorer
	volatilityAnalyzer    *VolatilityAnalyzer
	contractAuditor       *ContractAuditor
	portfolioAssessor     *PortfolioAssessor
	marketPredictor       *MarketPredictor

	// Risk State
	riskCache   map[string]*RiskAssessment
	alertsCache map[string]*RiskAlert
	cacheMutex  sync.RWMutex

	// Monitoring
	isRunning    bool
	updateTicker *time.Ticker
	stopChan     chan struct{}
	mutex        sync.RWMutex
}

// RiskManagerConfig holds configuration for the risk management system
type RiskManagerConfig struct {
	Enabled                bool                        `json:"enabled" yaml:"enabled"`
	UpdateInterval         time.Duration               `json:"update_interval" yaml:"update_interval"`
	CacheTimeout           time.Duration               `json:"cache_timeout" yaml:"cache_timeout"`
	AlertThresholds        AlertThresholds             `json:"alert_thresholds" yaml:"alert_thresholds"`
	TransactionRiskConfig  TransactionRiskConfig       `json:"transaction_risk_config" yaml:"transaction_risk_config"`
	VolatilityConfig       VolatilityConfig            `json:"volatility_config" yaml:"volatility_config"`
	ContractAuditConfig    ContractAuditConfig         `json:"contract_audit_config" yaml:"contract_audit_config"`
	PortfolioConfig        PortfolioRiskConfig         `json:"portfolio_config" yaml:"portfolio_config"`
	MarketPredictionConfig MarketPredictionConfig      `json:"market_prediction_config" yaml:"market_prediction_config"`
	MLModelPaths           map[string]string           `json:"ml_model_paths" yaml:"ml_model_paths"`
	DataSources            map[string]DataSourceConfig `json:"data_sources" yaml:"data_sources"`
}

// AlertThresholds defines risk alert thresholds
type AlertThresholds struct {
	TransactionRisk   decimal.Decimal `json:"transaction_risk" yaml:"transaction_risk"`
	PortfolioRisk     decimal.Decimal `json:"portfolio_risk" yaml:"portfolio_risk"`
	VolatilityRisk    decimal.Decimal `json:"volatility_risk" yaml:"volatility_risk"`
	ContractRisk      decimal.Decimal `json:"contract_risk" yaml:"contract_risk"`
	MarketRisk        decimal.Decimal `json:"market_risk" yaml:"market_risk"`
	ConcentrationRisk decimal.Decimal `json:"concentration_risk" yaml:"concentration_risk"`
	LiquidityRisk     decimal.Decimal `json:"liquidity_risk" yaml:"liquidity_risk"`
}

// DataSourceConfig holds configuration for external data sources
type DataSourceConfig struct {
	Enabled   bool              `json:"enabled" yaml:"enabled"`
	URL       string            `json:"url" yaml:"url"`
	APIKey    string            `json:"api_key" yaml:"api_key"`
	RateLimit int               `json:"rate_limit" yaml:"rate_limit"`
	Timeout   time.Duration     `json:"timeout" yaml:"timeout"`
	Headers   map[string]string `json:"headers" yaml:"headers"`
	Priority  int               `json:"priority" yaml:"priority"`
}

// RiskAssessment represents a comprehensive risk assessment
type RiskAssessment struct {
	ID               string                  `json:"id"`
	Address          common.Address          `json:"address"`
	Timestamp        time.Time               `json:"timestamp"`
	OverallRiskScore decimal.Decimal         `json:"overall_risk_score"`
	RiskLevel        string                  `json:"risk_level"`
	TransactionRisk  *TransactionRiskResult  `json:"transaction_risk"`
	PortfolioRisk    *PortfolioRiskResult    `json:"portfolio_risk"`
	VolatilityRisk   *VolatilityResult       `json:"volatility_risk"`
	ContractRisk     *ContractAuditResult    `json:"contract_risk"`
	MarketRisk       *MarketPredictionResult `json:"market_risk"`
	Recommendations  []string                `json:"recommendations"`
	Alerts           []*RiskAlert            `json:"alerts"`
	Confidence       decimal.Decimal         `json:"confidence"`
	ExpiresAt        time.Time               `json:"expires_at"`
	Metadata         map[string]interface{}  `json:"metadata"`
}

// RiskAlert represents a risk alert
type RiskAlert struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Severity     string                 `json:"severity"`
	Title        string                 `json:"title"`
	Message      string                 `json:"message"`
	RiskScore    decimal.Decimal        `json:"risk_score"`
	Threshold    decimal.Decimal        `json:"threshold"`
	Address      common.Address         `json:"address"`
	CreatedAt    time.Time              `json:"created_at"`
	ExpiresAt    time.Time              `json:"expires_at"`
	Acknowledged bool                   `json:"acknowledged"`
	Actions      []string               `json:"actions"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// RiskMetrics represents aggregated risk metrics
type RiskMetrics struct {
	TotalAssessments int                        `json:"total_assessments"`
	HighRiskCount    int                        `json:"high_risk_count"`
	MediumRiskCount  int                        `json:"medium_risk_count"`
	LowRiskCount     int                        `json:"low_risk_count"`
	AverageRiskScore decimal.Decimal            `json:"average_risk_score"`
	AlertCounts      map[string]int             `json:"alert_counts"`
	RiskDistribution map[string]decimal.Decimal `json:"risk_distribution"`
	TrendAnalysis    *RiskTrendAnalysis         `json:"trend_analysis"`
	LastUpdated      time.Time                  `json:"last_updated"`
}

// RiskTrendAnalysis represents risk trend analysis
type RiskTrendAnalysis struct {
	Direction          string          `json:"direction"`
	ChangeRate         decimal.Decimal `json:"change_rate"`
	Volatility         decimal.Decimal `json:"volatility"`
	PredictedRisk24h   decimal.Decimal `json:"predicted_risk_24h"`
	PredictedRisk7d    decimal.Decimal `json:"predicted_risk_7d"`
	ConfidenceInterval decimal.Decimal `json:"confidence_interval"`
}

// NewRiskManager creates a new AI-powered risk manager
func NewRiskManager(logger *logger.Logger, config RiskManagerConfig) *RiskManager {
	manager := &RiskManager{
		logger:      logger.Named("risk-manager"),
		config:      config,
		riskCache:   make(map[string]*RiskAssessment),
		alertsCache: make(map[string]*RiskAlert),
		stopChan:    make(chan struct{}),
	}

	// Initialize AI components
	manager.transactionRiskScorer = NewTransactionRiskScorer(logger, config.TransactionRiskConfig)
	manager.volatilityAnalyzer = NewVolatilityAnalyzer(logger, config.VolatilityConfig)
	manager.contractAuditor = NewContractAuditor(logger, config.ContractAuditConfig)
	manager.portfolioAssessor = NewPortfolioAssessor(logger, config.PortfolioConfig)
	manager.marketPredictor = NewMarketPredictor(logger, config.MarketPredictionConfig)

	return manager
}

// Start starts the risk management system
func (rm *RiskManager) Start(ctx context.Context) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if rm.isRunning {
		return fmt.Errorf("risk manager is already running")
	}

	if !rm.config.Enabled {
		rm.logger.Info("Risk manager is disabled")
		return nil
	}

	rm.logger.Info("Starting AI-powered risk management system",
		zap.Duration("update_interval", rm.config.UpdateInterval),
		zap.Duration("cache_timeout", rm.config.CacheTimeout))

	// Start AI components
	if err := rm.transactionRiskScorer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start transaction risk scorer: %w", err)
	}

	if err := rm.volatilityAnalyzer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start volatility analyzer: %w", err)
	}

	if err := rm.contractAuditor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start contract auditor: %w", err)
	}

	if err := rm.portfolioAssessor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start portfolio assessor: %w", err)
	}

	if err := rm.marketPredictor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start market predictor: %w", err)
	}

	// Start monitoring loop
	rm.updateTicker = time.NewTicker(rm.config.UpdateInterval)
	go rm.monitoringLoop(ctx)

	rm.isRunning = true
	rm.logger.Info("AI-powered risk management system started successfully")
	return nil
}

// Stop stops the risk management system
func (rm *RiskManager) Stop() error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if !rm.isRunning {
		return nil
	}

	rm.logger.Info("Stopping AI-powered risk management system")

	// Stop monitoring
	if rm.updateTicker != nil {
		rm.updateTicker.Stop()
	}
	close(rm.stopChan)

	// Stop AI components
	if rm.marketPredictor != nil {
		rm.marketPredictor.Stop()
	}
	if rm.portfolioAssessor != nil {
		rm.portfolioAssessor.Stop()
	}
	if rm.contractAuditor != nil {
		rm.contractAuditor.Stop()
	}
	if rm.volatilityAnalyzer != nil {
		rm.volatilityAnalyzer.Stop()
	}
	if rm.transactionRiskScorer != nil {
		rm.transactionRiskScorer.Stop()
	}

	rm.isRunning = false
	rm.logger.Info("AI-powered risk management system stopped")
	return nil
}

// AssessRisk performs comprehensive risk assessment
func (rm *RiskManager) AssessRisk(ctx context.Context, req *RiskAssessmentRequest) (*RiskAssessment, error) {
	rm.logger.Debug("Performing comprehensive risk assessment",
		zap.String("address", req.Address.Hex()),
		zap.String("type", req.AssessmentType))

	// Check cache first
	cacheKey := rm.generateCacheKey(req)
	if cached := rm.getCachedAssessment(cacheKey); cached != nil {
		rm.logger.Debug("Returning cached risk assessment")
		return cached, nil
	}

	// Perform comprehensive assessment
	assessment := &RiskAssessment{
		ID:        rm.generateAssessmentID(),
		Address:   req.Address,
		Timestamp: time.Now(),
		ExpiresAt: time.Now().Add(rm.config.CacheTimeout),
		Metadata:  make(map[string]interface{}),
	}

	// Run all risk assessments in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)

	// Transaction Risk Assessment
	if req.IncludeTransactionRisk {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := rm.transactionRiskScorer.AssessTransactionRisk(ctx, &TransactionRiskRequest{
				Address:     req.Address,
				Transaction: req.Transaction,
			})
			mu.Lock()
			if err != nil {
				errors = append(errors, fmt.Errorf("transaction risk: %w", err))
			} else {
				assessment.TransactionRisk = result
			}
			mu.Unlock()
		}()
	}

	// Portfolio Risk Assessment
	if req.IncludePortfolioRisk {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := rm.portfolioAssessor.AssessPortfolioRisk(ctx, &PortfolioRiskRequest{
				Address:   req.Address,
				Portfolio: req.Portfolio,
			})
			mu.Lock()
			if err != nil {
				errors = append(errors, fmt.Errorf("portfolio risk: %w", err))
			} else {
				assessment.PortfolioRisk = result
			}
			mu.Unlock()
		}()
	}

	// Volatility Risk Assessment
	if req.IncludeVolatilityRisk {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := rm.volatilityAnalyzer.AnalyzeVolatility(ctx, &VolatilityRequest{
				Assets:    req.Assets,
				TimeFrame: req.TimeFrame,
			})
			mu.Lock()
			if err != nil {
				errors = append(errors, fmt.Errorf("volatility risk: %w", err))
			} else {
				assessment.VolatilityRisk = result
			}
			mu.Unlock()
		}()
	}

	// Contract Risk Assessment
	if req.IncludeContractRisk && req.ContractAddress != (common.Address{}) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := rm.contractAuditor.AuditContract(ctx, &ContractAuditRequest{
				ContractAddress: req.ContractAddress,
				SourceCode:      req.SourceCode,
			})
			mu.Lock()
			if err != nil {
				errors = append(errors, fmt.Errorf("contract risk: %w", err))
			} else {
				assessment.ContractRisk = result
			}
			mu.Unlock()
		}()
	}

	// Market Risk Assessment
	if req.IncludeMarketRisk {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := rm.marketPredictor.PredictMarketRisk(ctx, &MarketPredictionRequest{
				Assets:    req.Assets,
				TimeFrame: req.TimeFrame,
			})
			mu.Lock()
			if err != nil {
				errors = append(errors, fmt.Errorf("market risk: %w", err))
			} else {
				assessment.MarketRisk = result
			}
			mu.Unlock()
		}()
	}

	// Wait for all assessments to complete
	wg.Wait()

	// Check for errors
	if len(errors) > 0 {
		rm.logger.Warn("Some risk assessments failed", zap.Int("error_count", len(errors)))
		for _, err := range errors {
			rm.logger.Warn("Risk assessment error", zap.Error(err))
		}
	}

	// Calculate overall risk score
	assessment.OverallRiskScore = rm.calculateOverallRiskScore(assessment)
	assessment.RiskLevel = rm.determineRiskLevel(assessment.OverallRiskScore)
	assessment.Confidence = rm.calculateConfidence(assessment)

	// Generate recommendations
	assessment.Recommendations = rm.generateRecommendations(assessment)

	// Generate alerts
	assessment.Alerts = rm.generateAlerts(assessment)

	// Cache the assessment
	rm.cacheAssessment(cacheKey, assessment)

	rm.logger.Info("Risk assessment completed",
		zap.String("address", req.Address.Hex()),
		zap.String("overall_risk_score", assessment.OverallRiskScore.String()),
		zap.String("risk_level", assessment.RiskLevel),
		zap.Int("alert_count", len(assessment.Alerts)))

	return assessment, nil
}

// GetRiskMetrics returns aggregated risk metrics
func (rm *RiskManager) GetRiskMetrics(ctx context.Context) (*RiskMetrics, error) {
	rm.cacheMutex.RLock()
	defer rm.cacheMutex.RUnlock()

	totalAssessments := len(rm.riskCache)
	highRiskCount := 0
	mediumRiskCount := 0
	lowRiskCount := 0
	totalRiskScore := decimal.Zero

	alertCounts := make(map[string]int)
	riskDistribution := make(map[string]decimal.Decimal)

	for _, assessment := range rm.riskCache {
		totalRiskScore = totalRiskScore.Add(assessment.OverallRiskScore)

		switch assessment.RiskLevel {
		case "high":
			highRiskCount++
		case "medium":
			mediumRiskCount++
		case "low":
			lowRiskCount++
		}

		for _, alert := range assessment.Alerts {
			alertCounts[alert.Type]++
		}
	}

	averageRiskScore := decimal.Zero
	if totalAssessments > 0 {
		averageRiskScore = totalRiskScore.Div(decimal.NewFromInt(int64(totalAssessments)))
	}

	// Calculate risk distribution
	if totalAssessments > 0 {
		riskDistribution["high"] = decimal.NewFromInt(int64(highRiskCount)).Div(decimal.NewFromInt(int64(totalAssessments)))
		riskDistribution["medium"] = decimal.NewFromInt(int64(mediumRiskCount)).Div(decimal.NewFromInt(int64(totalAssessments)))
		riskDistribution["low"] = decimal.NewFromInt(int64(lowRiskCount)).Div(decimal.NewFromInt(int64(totalAssessments)))
	}

	// Generate trend analysis
	trendAnalysis := rm.generateTrendAnalysis()

	return &RiskMetrics{
		TotalAssessments: totalAssessments,
		HighRiskCount:    highRiskCount,
		MediumRiskCount:  mediumRiskCount,
		LowRiskCount:     lowRiskCount,
		AverageRiskScore: averageRiskScore,
		AlertCounts:      alertCounts,
		RiskDistribution: riskDistribution,
		TrendAnalysis:    trendAnalysis,
		LastUpdated:      time.Now(),
	}, nil
}

// GetActiveAlerts returns active risk alerts
func (rm *RiskManager) GetActiveAlerts(ctx context.Context, address *common.Address) ([]*RiskAlert, error) {
	rm.cacheMutex.RLock()
	defer rm.cacheMutex.RUnlock()

	var alerts []*RiskAlert
	now := time.Now()

	for _, alert := range rm.alertsCache {
		// Filter by address if specified
		if address != nil && alert.Address != *address {
			continue
		}

		// Only return active (non-expired, non-acknowledged) alerts
		if !alert.Acknowledged && now.Before(alert.ExpiresAt) {
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// AcknowledgeAlert acknowledges a risk alert
func (rm *RiskManager) AcknowledgeAlert(ctx context.Context, alertID string) error {
	rm.cacheMutex.Lock()
	defer rm.cacheMutex.Unlock()

	alert, exists := rm.alertsCache[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	alert.Acknowledged = true
	rm.logger.Info("Risk alert acknowledged", zap.String("alert_id", alertID))

	return nil
}

// Helper methods

// monitoringLoop runs the main monitoring loop
func (rm *RiskManager) monitoringLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-rm.stopChan:
			return
		case <-rm.updateTicker.C:
			rm.performMaintenance()
		}
	}
}

// performMaintenance performs periodic maintenance tasks
func (rm *RiskManager) performMaintenance() {
	rm.logger.Debug("Performing risk manager maintenance")

	// Clean up expired cache entries
	rm.cleanupExpiredCache()

	// Update risk trends
	rm.updateRiskTrends()

	// Generate system-wide alerts
	rm.generateSystemAlerts()
}

// cleanupExpiredCache removes expired cache entries
func (rm *RiskManager) cleanupExpiredCache() {
	rm.cacheMutex.Lock()
	defer rm.cacheMutex.Unlock()

	now := time.Now()

	// Clean up risk assessments
	for key, assessment := range rm.riskCache {
		if now.After(assessment.ExpiresAt) {
			delete(rm.riskCache, key)
		}
	}

	// Clean up alerts
	for key, alert := range rm.alertsCache {
		if now.After(alert.ExpiresAt) {
			delete(rm.alertsCache, key)
		}
	}
}

// calculateOverallRiskScore calculates the overall risk score
func (rm *RiskManager) calculateOverallRiskScore(assessment *RiskAssessment) decimal.Decimal {
	weights := map[string]decimal.Decimal{
		"transaction": decimal.NewFromFloat(0.25),
		"portfolio":   decimal.NewFromFloat(0.25),
		"volatility":  decimal.NewFromFloat(0.20),
		"contract":    decimal.NewFromFloat(0.15),
		"market":      decimal.NewFromFloat(0.15),
	}

	totalScore := decimal.Zero
	totalWeight := decimal.Zero

	if assessment.TransactionRisk != nil {
		totalScore = totalScore.Add(assessment.TransactionRisk.RiskScore.Mul(weights["transaction"]))
		totalWeight = totalWeight.Add(weights["transaction"])
	}

	if assessment.PortfolioRisk != nil {
		totalScore = totalScore.Add(assessment.PortfolioRisk.OverallRiskScore.Mul(weights["portfolio"]))
		totalWeight = totalWeight.Add(weights["portfolio"])
	}

	if assessment.VolatilityRisk != nil {
		totalScore = totalScore.Add(assessment.VolatilityRisk.RiskScore.Mul(weights["volatility"]))
		totalWeight = totalWeight.Add(weights["volatility"])
	}

	if assessment.ContractRisk != nil {
		totalScore = totalScore.Add(assessment.ContractRisk.SecurityScore.Mul(weights["contract"]))
		totalWeight = totalWeight.Add(weights["contract"])
	}

	if assessment.MarketRisk != nil {
		totalScore = totalScore.Add(assessment.MarketRisk.RiskScore.Mul(weights["market"]))
		totalWeight = totalWeight.Add(weights["market"])
	}

	if totalWeight.IsZero() {
		return decimal.Zero
	}

	return totalScore.Div(totalWeight)
}

// determineRiskLevel determines the risk level based on score
func (rm *RiskManager) determineRiskLevel(score decimal.Decimal) string {
	if score.GreaterThan(decimal.NewFromFloat(70)) {
		return "high"
	} else if score.GreaterThan(decimal.NewFromFloat(40)) {
		return "medium"
	}
	return "low"
}

// calculateConfidence calculates confidence in the assessment
func (rm *RiskManager) calculateConfidence(assessment *RiskAssessment) decimal.Decimal {
	// Simplified confidence calculation based on available data
	dataPoints := 0
	totalConfidence := decimal.Zero

	if assessment.TransactionRisk != nil {
		dataPoints++
		totalConfidence = totalConfidence.Add(assessment.TransactionRisk.Confidence)
	}

	if assessment.PortfolioRisk != nil {
		dataPoints++
		totalConfidence = totalConfidence.Add(decimal.NewFromFloat(0.8)) // Mock confidence
	}

	if assessment.VolatilityRisk != nil {
		dataPoints++
		totalConfidence = totalConfidence.Add(assessment.VolatilityRisk.Confidence)
	}

	if assessment.ContractRisk != nil {
		dataPoints++
		totalConfidence = totalConfidence.Add(assessment.ContractRisk.Confidence)
	}

	if assessment.MarketRisk != nil {
		dataPoints++
		totalConfidence = totalConfidence.Add(assessment.MarketRisk.Confidence)
	}

	if dataPoints == 0 {
		return decimal.NewFromFloat(0.5)
	}

	return totalConfidence.Div(decimal.NewFromInt(int64(dataPoints)))
}

// generateRecommendations generates risk mitigation recommendations
func (rm *RiskManager) generateRecommendations(assessment *RiskAssessment) []string {
	var recommendations []string

	// Transaction risk recommendations
	if assessment.TransactionRisk != nil && assessment.TransactionRisk.RiskScore.GreaterThan(decimal.NewFromFloat(70)) {
		recommendations = append(recommendations, "Consider using a hardware wallet for high-value transactions")
		recommendations = append(recommendations, "Verify recipient addresses carefully")
	}

	// Portfolio risk recommendations
	if assessment.PortfolioRisk != nil {
		if assessment.PortfolioRisk.ConcentrationRisk.GreaterThan(decimal.NewFromFloat(0.7)) {
			recommendations = append(recommendations, "Diversify portfolio to reduce concentration risk")
		}
		if assessment.PortfolioRisk.CorrelationRisk.GreaterThan(decimal.NewFromFloat(0.8)) {
			recommendations = append(recommendations, "Add uncorrelated assets to reduce portfolio correlation")
		}
	}

	// Volatility risk recommendations
	if assessment.VolatilityRisk != nil && assessment.VolatilityRisk.RiskScore.GreaterThan(decimal.NewFromFloat(60)) {
		recommendations = append(recommendations, "Consider reducing position sizes due to high volatility")
		recommendations = append(recommendations, "Implement stop-loss orders to limit downside risk")
	}

	// Contract risk recommendations
	if assessment.ContractRisk != nil && assessment.ContractRisk.SecurityScore.LessThan(decimal.NewFromFloat(70)) {
		recommendations = append(recommendations, "Exercise caution when interacting with this contract")
		recommendations = append(recommendations, "Consider waiting for additional security audits")
	}

	// Market risk recommendations
	if assessment.MarketRisk != nil && assessment.MarketRisk.RiskScore.GreaterThan(decimal.NewFromFloat(70)) {
		recommendations = append(recommendations, "Consider reducing exposure during high market risk periods")
		recommendations = append(recommendations, "Monitor market conditions closely")
	}

	return recommendations
}

// generateAlerts generates risk alerts based on assessment
func (rm *RiskManager) generateAlerts(assessment *RiskAssessment) []*RiskAlert {
	var alerts []*RiskAlert

	// Transaction risk alerts
	if assessment.TransactionRisk != nil && assessment.TransactionRisk.RiskScore.GreaterThan(rm.config.AlertThresholds.TransactionRisk) {
		alert := &RiskAlert{
			ID:        rm.generateAlertID(),
			Type:      "transaction_risk",
			Severity:  rm.determineSeverity(assessment.TransactionRisk.RiskScore),
			Title:     "High Transaction Risk Detected",
			Message:   fmt.Sprintf("Transaction risk score: %s", assessment.TransactionRisk.RiskScore.String()),
			RiskScore: assessment.TransactionRisk.RiskScore,
			Threshold: rm.config.AlertThresholds.TransactionRisk,
			Address:   assessment.Address,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Actions:   []string{"Review transaction details", "Consider using hardware wallet"},
			Metadata:  make(map[string]interface{}),
		}
		alerts = append(alerts, alert)
		rm.cacheAlert(alert)
	}

	// Portfolio risk alerts
	if assessment.PortfolioRisk != nil && assessment.PortfolioRisk.OverallRiskScore.GreaterThan(rm.config.AlertThresholds.PortfolioRisk) {
		alert := &RiskAlert{
			ID:        rm.generateAlertID(),
			Type:      "portfolio_risk",
			Severity:  rm.determineSeverity(assessment.PortfolioRisk.OverallRiskScore),
			Title:     "High Portfolio Risk Detected",
			Message:   fmt.Sprintf("Portfolio risk score: %s", assessment.PortfolioRisk.OverallRiskScore.String()),
			RiskScore: assessment.PortfolioRisk.OverallRiskScore,
			Threshold: rm.config.AlertThresholds.PortfolioRisk,
			Address:   assessment.Address,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Actions:   []string{"Diversify portfolio", "Reduce position sizes"},
			Metadata:  make(map[string]interface{}),
		}
		alerts = append(alerts, alert)
		rm.cacheAlert(alert)
	}

	// Volatility risk alerts
	if assessment.VolatilityRisk != nil && assessment.VolatilityRisk.RiskScore.GreaterThan(rm.config.AlertThresholds.VolatilityRisk) {
		alert := &RiskAlert{
			ID:        rm.generateAlertID(),
			Type:      "volatility_risk",
			Severity:  rm.determineSeverity(assessment.VolatilityRisk.RiskScore),
			Title:     "High Volatility Risk Detected",
			Message:   fmt.Sprintf("Volatility risk score: %s", assessment.VolatilityRisk.RiskScore.String()),
			RiskScore: assessment.VolatilityRisk.RiskScore,
			Threshold: rm.config.AlertThresholds.VolatilityRisk,
			Address:   assessment.Address,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(12 * time.Hour),
			Actions:   []string{"Reduce exposure", "Implement stop-loss orders"},
			Metadata:  make(map[string]interface{}),
		}
		alerts = append(alerts, alert)
		rm.cacheAlert(alert)
	}

	return alerts
}

// generateTrendAnalysis generates risk trend analysis
func (rm *RiskManager) generateTrendAnalysis() *RiskTrendAnalysis {
	// Simplified trend analysis - in production, implement proper time series analysis
	return &RiskTrendAnalysis{
		Direction:          "stable",
		ChangeRate:         decimal.NewFromFloat(0.02),
		Volatility:         decimal.NewFromFloat(0.15),
		PredictedRisk24h:   decimal.NewFromFloat(45),
		PredictedRisk7d:    decimal.NewFromFloat(48),
		ConfidenceInterval: decimal.NewFromFloat(0.85),
	}
}

// updateRiskTrends updates risk trends
func (rm *RiskManager) updateRiskTrends() {
	rm.logger.Debug("Updating risk trends")
	// Implementation for updating risk trends
}

// generateSystemAlerts generates system-wide alerts
func (rm *RiskManager) generateSystemAlerts() {
	rm.logger.Debug("Generating system alerts")
	// Implementation for generating system-wide alerts
}

// generateCacheKey generates a cache key for risk assessment
func (rm *RiskManager) generateCacheKey(req *RiskAssessmentRequest) string {
	return fmt.Sprintf("%s_%s_%d", req.Address.Hex(), req.AssessmentType, time.Now().Unix()/3600)
}

// generateAssessmentID generates a unique assessment ID
func (rm *RiskManager) generateAssessmentID() string {
	return fmt.Sprintf("risk_%d", time.Now().UnixNano())
}

// generateAlertID generates a unique alert ID
func (rm *RiskManager) generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

// getCachedAssessment retrieves a cached risk assessment
func (rm *RiskManager) getCachedAssessment(key string) *RiskAssessment {
	rm.cacheMutex.RLock()
	defer rm.cacheMutex.RUnlock()

	assessment, exists := rm.riskCache[key]
	if !exists {
		return nil
	}

	if time.Now().After(assessment.ExpiresAt) {
		delete(rm.riskCache, key)
		return nil
	}

	return assessment
}

// cacheAssessment caches a risk assessment
func (rm *RiskManager) cacheAssessment(key string, assessment *RiskAssessment) {
	rm.cacheMutex.Lock()
	defer rm.cacheMutex.Unlock()
	rm.riskCache[key] = assessment
}

// cacheAlert caches a risk alert
func (rm *RiskManager) cacheAlert(alert *RiskAlert) {
	rm.cacheMutex.Lock()
	defer rm.cacheMutex.Unlock()
	rm.alertsCache[alert.ID] = alert
}

// determineSeverity determines alert severity based on risk score
func (rm *RiskManager) determineSeverity(score decimal.Decimal) string {
	if score.GreaterThan(decimal.NewFromFloat(80)) {
		return "critical"
	} else if score.GreaterThan(decimal.NewFromFloat(60)) {
		return "high"
	} else if score.GreaterThan(decimal.NewFromFloat(40)) {
		return "medium"
	}
	return "low"
}

// IsRunning returns whether the risk manager is running
func (rm *RiskManager) IsRunning() bool {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	return rm.isRunning
}

// GetSystemHealth returns system health status
func (rm *RiskManager) GetSystemHealth() *SystemHealth {
	componentStatuses := make(map[string]string)

	if rm.transactionRiskScorer != nil {
		if rm.transactionRiskScorer.IsHealthy() {
			componentStatuses["transaction_risk_scorer"] = "healthy"
		} else {
			componentStatuses["transaction_risk_scorer"] = "unhealthy"
		}
	}

	if rm.volatilityAnalyzer != nil {
		if rm.volatilityAnalyzer.IsHealthy() {
			componentStatuses["volatility_analyzer"] = "healthy"
		} else {
			componentStatuses["volatility_analyzer"] = "unhealthy"
		}
	}

	if rm.contractAuditor != nil {
		if rm.contractAuditor.IsHealthy() {
			componentStatuses["contract_auditor"] = "healthy"
		} else {
			componentStatuses["contract_auditor"] = "unhealthy"
		}
	}

	if rm.portfolioAssessor != nil {
		if rm.portfolioAssessor.IsHealthy() {
			componentStatuses["portfolio_assessor"] = "healthy"
		} else {
			componentStatuses["portfolio_assessor"] = "unhealthy"
		}
	}

	if rm.marketPredictor != nil {
		if rm.marketPredictor.IsHealthy() {
			componentStatuses["market_predictor"] = "healthy"
		} else {
			componentStatuses["market_predictor"] = "unhealthy"
		}
	}

	// Determine overall status
	overallStatus := "healthy"
	for _, status := range componentStatuses {
		if status == "unhealthy" {
			overallStatus = "degraded"
			break
		}
	}

	return &SystemHealth{
		OverallStatus:     overallStatus,
		ComponentStatuses: componentStatuses,
		LastHealthCheck:   time.Now(),
		Uptime:            time.Since(time.Now().Add(-24 * time.Hour)), // Mock uptime
		ErrorRate:         decimal.NewFromFloat(0.01),
		PerformanceMetrics: map[string]interface{}{
			"assessments_per_minute": 10,
			"cache_hit_rate":         0.85,
			"average_response_time":  "250ms",
		},
	}
}
