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

// PortfolioRiskAnalyzer provides comprehensive portfolio risk assessment
type PortfolioRiskAnalyzer struct {
	logger *logger.Logger
	config PortfolioRiskAnalyzerConfig

	// Analysis components
	correlationAnalyzer   *CorrelationAnalyzer
	diversificationEngine *DiversificationEngine
	varCalculator         *VaRCalculator
	riskMetricsEngine     *RiskMetricsEngine

	// Data management
	portfolioCache map[string]*PortfolioRiskAnalysis
	priceHistory   map[string][]PricePoint
	cacheMutex     sync.RWMutex
	dataMutex      sync.RWMutex

	// State management
	isRunning    bool
	updateTicker *time.Ticker
	stopChan     chan struct{}
	mutex        sync.RWMutex
}

// PortfolioRiskAnalyzerConfig holds configuration for portfolio risk analysis
type PortfolioRiskAnalyzerConfig struct {
	Enabled               bool                     `json:"enabled" yaml:"enabled"`
	UpdateInterval        time.Duration            `json:"update_interval" yaml:"update_interval"`
	CacheTimeout          time.Duration            `json:"cache_timeout" yaml:"cache_timeout"`
	HistoryWindow         int                      `json:"history_window" yaml:"history_window"`
	ConfidenceLevel       decimal.Decimal          `json:"confidence_level" yaml:"confidence_level"`
	RiskFreeRate          decimal.Decimal          `json:"risk_free_rate" yaml:"risk_free_rate"`
	CorrelationConfig     CorrelationConfig        `json:"correlation_config" yaml:"correlation_config"`
	DiversificationConfig DiversificationConfig    `json:"diversification_config" yaml:"diversification_config"`
	VaRConfig             VaRConfig                `json:"var_config" yaml:"var_config"`
	RiskMetricsConfig     RiskMetricsConfig        `json:"risk_metrics_config" yaml:"risk_metrics_config"`
	AlertThresholds       PortfolioAlertThresholds `json:"alert_thresholds" yaml:"alert_thresholds"`
	DataSources           []string                 `json:"data_sources" yaml:"data_sources"`
	RebalancingThresholds RebalancingThresholds    `json:"rebalancing_thresholds" yaml:"rebalancing_thresholds"`
}

// CorrelationConfig holds correlation analysis configuration
type CorrelationConfig struct {
	Enabled            bool          `json:"enabled" yaml:"enabled"`
	WindowSize         int           `json:"window_size" yaml:"window_size"`
	UpdateInterval     time.Duration `json:"update_interval" yaml:"update_interval"`
	MinDataPoints      int           `json:"min_data_points" yaml:"min_data_points"`
	CorrelationMethods []string      `json:"correlation_methods" yaml:"correlation_methods"`
}

// DiversificationConfig holds diversification analysis configuration
type DiversificationConfig struct {
	Enabled          bool                       `json:"enabled" yaml:"enabled"`
	MaxConcentration decimal.Decimal            `json:"max_concentration" yaml:"max_concentration"`
	MinAssets        int                        `json:"min_assets" yaml:"min_assets"`
	SectorLimits     map[string]decimal.Decimal `json:"sector_limits" yaml:"sector_limits"`
	ChainLimits      map[string]decimal.Decimal `json:"chain_limits" yaml:"chain_limits"`
	ProtocolLimits   map[string]decimal.Decimal `json:"protocol_limits" yaml:"protocol_limits"`
}

// VaRConfig holds Value at Risk calculation configuration
type VaRConfig struct {
	Enabled          bool              `json:"enabled" yaml:"enabled"`
	Methods          []string          `json:"methods" yaml:"methods"`
	ConfidenceLevels []decimal.Decimal `json:"confidence_levels" yaml:"confidence_levels"`
	TimeHorizons     []int             `json:"time_horizons" yaml:"time_horizons"`
	MonteCarloSims   int               `json:"monte_carlo_sims" yaml:"monte_carlo_sims"`
}

// RiskMetricsConfig holds risk metrics calculation configuration
type RiskMetricsConfig struct {
	Enabled          bool   `json:"enabled" yaml:"enabled"`
	CalculateSharpе  bool   `json:"calculate_sharpe" yaml:"calculate_sharpe"`
	CalculateSortino bool   `json:"calculate_sortino" yaml:"calculate_sortino"`
	CalculateTreynor bool   `json:"calculate_treynor" yaml:"calculate_treynor"`
	CalculateAlpha   bool   `json:"calculate_alpha" yaml:"calculate_alpha"`
	CalculateBeta    bool   `json:"calculate_beta" yaml:"calculate_beta"`
	BenchmarkAsset   string `json:"benchmark_asset" yaml:"benchmark_asset"`
}

// PortfolioAlertThresholds defines alert thresholds for portfolio metrics
type PortfolioAlertThresholds struct {
	MaxConcentration   decimal.Decimal `json:"max_concentration" yaml:"max_concentration"`
	MaxCorrelation     decimal.Decimal `json:"max_correlation" yaml:"max_correlation"`
	MaxVaR             decimal.Decimal `json:"max_var" yaml:"max_var"`
	MinSharpeRatio     decimal.Decimal `json:"min_sharpe_ratio" yaml:"min_sharpe_ratio"`
	MaxDrawdown        decimal.Decimal `json:"max_drawdown" yaml:"max_drawdown"`
	MinDiversification decimal.Decimal `json:"min_diversification" yaml:"min_diversification"`
}

// RebalancingThresholds defines thresholds for portfolio rebalancing
type RebalancingThresholds struct {
	Enabled              bool            `json:"enabled" yaml:"enabled"`
	DeviationThreshold   decimal.Decimal `json:"deviation_threshold" yaml:"deviation_threshold"`
	TimeThreshold        time.Duration   `json:"time_threshold" yaml:"time_threshold"`
	VolatilityThreshold  decimal.Decimal `json:"volatility_threshold" yaml:"volatility_threshold"`
	CorrelationThreshold decimal.Decimal `json:"correlation_threshold" yaml:"correlation_threshold"`
}

// PortfolioRiskAnalysis represents comprehensive portfolio risk analysis
type PortfolioRiskAnalysis struct {
	PortfolioID            string                  `json:"portfolio_id"`
	Address                common.Address          `json:"address"`
	AnalysisID             string                  `json:"analysis_id"`
	Timestamp              time.Time               `json:"timestamp"`
	Portfolio              *Portfolio              `json:"portfolio"`
	OverallRiskScore       decimal.Decimal         `json:"overall_risk_score"`
	RiskLevel              string                  `json:"risk_level"`
	CorrelationAnalysis    *CorrelationAnalysis    `json:"correlation_analysis"`
	DiversificationMetrics *DiversificationMetrics `json:"diversification_metrics"`
	VaRAnalysis            *VaRAnalysis            `json:"var_analysis"`
	RiskMetrics            *PortfolioRiskMetrics   `json:"risk_metrics"`
	PerformanceMetrics     *PerformanceMetrics     `json:"performance_metrics"`
	RebalancingAdvice      *RebalancingAdvice      `json:"rebalancing_advice"`
	RiskAlerts             []*PortfolioRiskAlert   `json:"risk_alerts"`
	Recommendations        []string                `json:"recommendations"`
	Confidence             decimal.Decimal         `json:"confidence"`
	AnalysisDuration       time.Duration           `json:"analysis_duration"`
	Metadata               map[string]interface{}  `json:"metadata"`
}

// Portfolio represents a portfolio for analysis
type Portfolio struct {
	ID          string                 `json:"id"`
	Address     common.Address         `json:"address"`
	TotalValue  decimal.Decimal        `json:"total_value"`
	Assets      []*PortfolioAsset      `json:"assets"`
	LastUpdated time.Time              `json:"last_updated"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PortfolioAsset represents an asset in a portfolio
type PortfolioAsset struct {
	Symbol      string          `json:"symbol"`
	Name        string          `json:"name"`
	Amount      decimal.Decimal `json:"amount"`
	Value       decimal.Decimal `json:"value"`
	Weight      decimal.Decimal `json:"weight"`
	Price       decimal.Decimal `json:"price"`
	Chain       string          `json:"chain"`
	Protocol    string          `json:"protocol"`
	Sector      string          `json:"sector"`
	AssetType   string          `json:"asset_type"`
	LastUpdated time.Time       `json:"last_updated"`
}

// CorrelationAnalysis represents correlation analysis results
type CorrelationAnalysis struct {
	CorrelationMatrix     map[string]map[string]decimal.Decimal `json:"correlation_matrix"`
	AverageCorrelation    decimal.Decimal                       `json:"average_correlation"`
	MaxCorrelation        decimal.Decimal                       `json:"max_correlation"`
	MinCorrelation        decimal.Decimal                       `json:"min_correlation"`
	HighlyCorrelatedPairs []CorrelationPair                     `json:"highly_correlated_pairs"`
	CorrelationRisk       decimal.Decimal                       `json:"correlation_risk"`
	DiversificationRatio  decimal.Decimal                       `json:"diversification_ratio"`
	EffectiveAssets       decimal.Decimal                       `json:"effective_assets"`
	Metadata              map[string]interface{}                `json:"metadata"`
}

// CorrelationPair represents a pair of correlated assets
type CorrelationPair struct {
	Asset1       string          `json:"asset1"`
	Asset2       string          `json:"asset2"`
	Correlation  decimal.Decimal `json:"correlation"`
	Significance decimal.Decimal `json:"significance"`
}

// DiversificationMetrics represents diversification analysis results
type DiversificationMetrics struct {
	ConcentrationRisk         decimal.Decimal            `json:"concentration_risk"`
	HerfindahlIndex           decimal.Decimal            `json:"herfindahl_index"`
	EffectiveAssetCount       decimal.Decimal            `json:"effective_asset_count"`
	SectorDiversification     decimal.Decimal            `json:"sector_diversification"`
	ChainDiversification      decimal.Decimal            `json:"chain_diversification"`
	ProtocolDiversification   decimal.Decimal            `json:"protocol_diversification"`
	GeographicDiversification decimal.Decimal            `json:"geographic_diversification"`
	DiversificationScore      decimal.Decimal            `json:"diversification_score"`
	ConcentrationBreakdown    map[string]decimal.Decimal `json:"concentration_breakdown"`
	Recommendations           []string                   `json:"recommendations"`
	Metadata                  map[string]interface{}     `json:"metadata"`
}

// VaRAnalysis represents Value at Risk analysis results
type VaRAnalysis struct {
	HistoricalVaR     map[string]decimal.Decimal `json:"historical_var"`
	ParametricVaR     map[string]decimal.Decimal `json:"parametric_var"`
	MonteCarloVaR     map[string]decimal.Decimal `json:"monte_carlo_var"`
	ConditionalVaR    map[string]decimal.Decimal `json:"conditional_var"`
	ExpectedShortfall map[string]decimal.Decimal `json:"expected_shortfall"`
	MaxDrawdown       decimal.Decimal            `json:"max_drawdown"`
	WorstCaseScenario decimal.Decimal            `json:"worst_case_scenario"`
	StressTestResults map[string]decimal.Decimal `json:"stress_test_results"`
	BacktestResults   *VaRBacktestResults        `json:"backtest_results"`
	Metadata          map[string]interface{}     `json:"metadata"`
}

// VaRBacktestResults represents VaR backtesting results
type VaRBacktestResults struct {
	Violations         int             `json:"violations"`
	ExpectedViolations int             `json:"expected_violations"`
	ViolationRate      decimal.Decimal `json:"violation_rate"`
	KupiecTest         decimal.Decimal `json:"kupiec_test"`
	ChristoffersenTest decimal.Decimal `json:"christoffersen_test"`
	IsValid            bool            `json:"is_valid"`
}

// PortfolioRiskMetrics represents portfolio risk metrics
type PortfolioRiskMetrics struct {
	SharpeRatio      decimal.Decimal        `json:"sharpe_ratio"`
	SortinoRatio     decimal.Decimal        `json:"sortino_ratio"`
	TreynorRatio     decimal.Decimal        `json:"treynor_ratio"`
	Alpha            decimal.Decimal        `json:"alpha"`
	Beta             decimal.Decimal        `json:"beta"`
	TrackingError    decimal.Decimal        `json:"tracking_error"`
	InformationRatio decimal.Decimal        `json:"information_ratio"`
	CalmarRatio      decimal.Decimal        `json:"calmar_ratio"`
	Volatility       decimal.Decimal        `json:"volatility"`
	Skewness         decimal.Decimal        `json:"skewness"`
	Kurtosis         decimal.Decimal        `json:"kurtosis"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// PerformanceMetrics represents portfolio performance metrics
type PerformanceMetrics struct {
	TotalReturn      decimal.Decimal        `json:"total_return"`
	AnnualizedReturn decimal.Decimal        `json:"annualized_return"`
	DailyReturns     []decimal.Decimal      `json:"daily_returns"`
	MonthlyReturns   []decimal.Decimal      `json:"monthly_returns"`
	YearlyReturns    []decimal.Decimal      `json:"yearly_returns"`
	BestDay          decimal.Decimal        `json:"best_day"`
	WorstDay         decimal.Decimal        `json:"worst_day"`
	PositiveDays     int                    `json:"positive_days"`
	NegativeDays     int                    `json:"negative_days"`
	WinRate          decimal.Decimal        `json:"win_rate"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// RebalancingAdvice represents portfolio rebalancing recommendations
type RebalancingAdvice struct {
	ShouldRebalance     bool                       `json:"should_rebalance"`
	RebalanceReason     string                     `json:"rebalance_reason"`
	TargetAllocations   map[string]decimal.Decimal `json:"target_allocations"`
	CurrentAllocations  map[string]decimal.Decimal `json:"current_allocations"`
	RebalanceActions    []*RebalanceAction         `json:"rebalance_actions"`
	ExpectedImprovement decimal.Decimal            `json:"expected_improvement"`
	RebalanceCost       decimal.Decimal            `json:"rebalance_cost"`
	NetBenefit          decimal.Decimal            `json:"net_benefit"`
	Metadata            map[string]interface{}     `json:"metadata"`
}

// RebalanceAction represents a specific rebalancing action
type RebalanceAction struct {
	Asset         string          `json:"asset"`
	Action        string          `json:"action"` // "buy", "sell", "hold"
	CurrentWeight decimal.Decimal `json:"current_weight"`
	TargetWeight  decimal.Decimal `json:"target_weight"`
	AmountChange  decimal.Decimal `json:"amount_change"`
	ValueChange   decimal.Decimal `json:"value_change"`
	Priority      int             `json:"priority"`
	Reason        string          `json:"reason"`
}

// PortfolioRiskAlert represents a portfolio risk alert
type PortfolioRiskAlert struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Severity  string                 `json:"severity"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Metric    string                 `json:"metric"`
	Value     decimal.Decimal        `json:"value"`
	Threshold decimal.Decimal        `json:"threshold"`
	CreatedAt time.Time              `json:"created_at"`
	Actions   []string               `json:"actions"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// PricePoint represents a price data point
type PricePoint struct {
	Timestamp time.Time       `json:"timestamp"`
	Price     decimal.Decimal `json:"price"`
	Volume    decimal.Decimal `json:"volume"`
}

// NewPortfolioRiskAnalyzer creates a new portfolio risk analyzer
func NewPortfolioRiskAnalyzer(logger *logger.Logger, config PortfolioRiskAnalyzerConfig) *PortfolioRiskAnalyzer {
	analyzer := &PortfolioRiskAnalyzer{
		logger:         logger.Named("portfolio-risk-analyzer"),
		config:         config,
		portfolioCache: make(map[string]*PortfolioRiskAnalysis),
		priceHistory:   make(map[string][]PricePoint),
		stopChan:       make(chan struct{}),
	}

	// Initialize analysis components
	analyzer.correlationAnalyzer = NewCorrelationAnalyzer(logger, config.CorrelationConfig)
	analyzer.diversificationEngine = NewDiversificationEngine(logger, config.DiversificationConfig)
	analyzer.varCalculator = NewVaRCalculator(logger, config.VaRConfig)
	analyzer.riskMetricsEngine = NewRiskMetricsEngine(logger, config.RiskMetricsConfig)

	return analyzer
}

// Start starts the portfolio risk analyzer
func (pra *PortfolioRiskAnalyzer) Start(ctx context.Context) error {
	pra.mutex.Lock()
	defer pra.mutex.Unlock()

	if pra.isRunning {
		return fmt.Errorf("portfolio risk analyzer is already running")
	}

	if !pra.config.Enabled {
		pra.logger.Info("Portfolio risk analyzer is disabled")
		return nil
	}

	pra.logger.Info("Starting portfolio risk analyzer",
		zap.Duration("update_interval", pra.config.UpdateInterval),
		zap.Int("history_window", pra.config.HistoryWindow))

	// Start analysis components
	if err := pra.correlationAnalyzer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start correlation analyzer: %w", err)
	}

	if err := pra.diversificationEngine.Start(ctx); err != nil {
		return fmt.Errorf("failed to start diversification engine: %w", err)
	}

	if err := pra.varCalculator.Start(ctx); err != nil {
		return fmt.Errorf("failed to start VaR calculator: %w", err)
	}

	if err := pra.riskMetricsEngine.Start(ctx); err != nil {
		return fmt.Errorf("failed to start risk metrics engine: %w", err)
	}

	// Start monitoring loop
	pra.updateTicker = time.NewTicker(pra.config.UpdateInterval)
	go pra.monitoringLoop(ctx)

	pra.isRunning = true
	pra.logger.Info("Portfolio risk analyzer started successfully")
	return nil
}

// Stop stops the portfolio risk analyzer
func (pra *PortfolioRiskAnalyzer) Stop() error {
	pra.mutex.Lock()
	defer pra.mutex.Unlock()

	if !pra.isRunning {
		return nil
	}

	pra.logger.Info("Stopping portfolio risk analyzer")

	// Stop monitoring
	if pra.updateTicker != nil {
		pra.updateTicker.Stop()
	}
	close(pra.stopChan)

	// Stop analysis components
	if pra.riskMetricsEngine != nil {
		pra.riskMetricsEngine.Stop()
	}
	if pra.varCalculator != nil {
		pra.varCalculator.Stop()
	}
	if pra.diversificationEngine != nil {
		pra.diversificationEngine.Stop()
	}
	if pra.correlationAnalyzer != nil {
		pra.correlationAnalyzer.Stop()
	}

	pra.isRunning = false
	pra.logger.Info("Portfolio risk analyzer stopped")
	return nil
}

// AnalyzePortfolioRisk performs comprehensive portfolio risk analysis
func (pra *PortfolioRiskAnalyzer) AnalyzePortfolioRisk(ctx context.Context, portfolio *Portfolio) (*PortfolioRiskAnalysis, error) {
	startTime := time.Now()
	pra.logger.Info("Starting portfolio risk analysis",
		zap.String("portfolio_id", portfolio.ID),
		zap.String("address", portfolio.Address.Hex()),
		zap.Int("asset_count", len(portfolio.Assets)),
		zap.String("total_value", portfolio.TotalValue.String()))

	// Check cache first
	cacheKey := pra.generateCacheKey(portfolio)
	if cached := pra.getCachedAnalysis(cacheKey); cached != nil {
		pra.logger.Debug("Returning cached portfolio analysis")
		return cached, nil
	}

	// Initialize analysis result
	analysis := &PortfolioRiskAnalysis{
		PortfolioID:     portfolio.ID,
		Address:         portfolio.Address,
		AnalysisID:      pra.generateAnalysisID(),
		Timestamp:       time.Now(),
		Portfolio:       portfolio,
		RiskAlerts:      []*PortfolioRiskAlert{},
		Recommendations: []string{},
		Metadata:        make(map[string]interface{}),
	}

	// Run analysis components in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)

	// Correlation analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		correlationAnalysis, err := pra.correlationAnalyzer.AnalyzeCorrelations(ctx, portfolio)
		mu.Lock()
		if err != nil {
			errors = append(errors, fmt.Errorf("correlation analysis: %w", err))
		} else {
			analysis.CorrelationAnalysis = correlationAnalysis
		}
		mu.Unlock()
	}()

	// Diversification analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		diversificationMetrics, err := pra.diversificationEngine.AnalyzeDiversification(ctx, portfolio)
		mu.Lock()
		if err != nil {
			errors = append(errors, fmt.Errorf("diversification analysis: %w", err))
		} else {
			analysis.DiversificationMetrics = diversificationMetrics
		}
		mu.Unlock()
	}()

	// VaR analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		varAnalysis, err := pra.varCalculator.CalculateVaR(ctx, portfolio)
		mu.Lock()
		if err != nil {
			errors = append(errors, fmt.Errorf("VaR analysis: %w", err))
		} else {
			analysis.VaRAnalysis = varAnalysis
		}
		mu.Unlock()
	}()

	// Risk metrics calculation
	wg.Add(1)
	go func() {
		defer wg.Done()
		riskMetrics, err := pra.riskMetricsEngine.CalculateRiskMetrics(ctx, portfolio)
		mu.Lock()
		if err != nil {
			errors = append(errors, fmt.Errorf("risk metrics: %w", err))
		} else {
			analysis.RiskMetrics = riskMetrics
		}
		mu.Unlock()
	}()

	// Performance metrics calculation
	wg.Add(1)
	go func() {
		defer wg.Done()
		performanceMetrics := pra.calculatePerformanceMetrics(portfolio)
		mu.Lock()
		analysis.PerformanceMetrics = performanceMetrics
		mu.Unlock()
	}()

	// Wait for all analyses to complete
	wg.Wait()

	// Check for errors
	if len(errors) > 0 {
		pra.logger.Warn("Some portfolio analyses failed", zap.Int("error_count", len(errors)))
		for _, err := range errors {
			pra.logger.Warn("Portfolio analysis error", zap.Error(err))
		}
	}

	// Calculate overall risk score and metrics
	analysis.OverallRiskScore = pra.calculateOverallRiskScore(analysis)
	analysis.RiskLevel = pra.determineRiskLevel(analysis.OverallRiskScore)
	analysis.Confidence = pra.calculateConfidence(analysis)

	// Generate rebalancing advice
	analysis.RebalancingAdvice = pra.generateRebalancingAdvice(analysis)

	// Generate risk alerts
	analysis.RiskAlerts = pra.generateRiskAlerts(analysis)

	// Generate recommendations
	analysis.Recommendations = pra.generateRecommendations(analysis)

	analysis.AnalysisDuration = time.Since(startTime)

	// Cache the analysis
	pra.cacheAnalysis(cacheKey, analysis)

	pra.logger.Info("Portfolio risk analysis completed",
		zap.String("portfolio_id", portfolio.ID),
		zap.String("overall_risk_score", analysis.OverallRiskScore.String()),
		zap.String("risk_level", analysis.RiskLevel),
		zap.Int("alert_count", len(analysis.RiskAlerts)),
		zap.Duration("duration", analysis.AnalysisDuration))

	return analysis, nil
}

// Helper methods

// calculateOverallRiskScore calculates the overall portfolio risk score
func (pra *PortfolioRiskAnalyzer) calculateOverallRiskScore(analysis *PortfolioRiskAnalysis) decimal.Decimal {
	weights := map[string]decimal.Decimal{
		"concentration":   decimal.NewFromFloat(0.25),
		"correlation":     decimal.NewFromFloat(0.20),
		"volatility":      decimal.NewFromFloat(0.20),
		"var":             decimal.NewFromFloat(0.15),
		"diversification": decimal.NewFromFloat(0.10),
		"performance":     decimal.NewFromFloat(0.10),
	}

	totalScore := decimal.Zero
	totalWeight := decimal.Zero

	// Concentration risk (from diversification metrics)
	if analysis.DiversificationMetrics != nil {
		concentrationScore := decimal.NewFromFloat(100).Sub(analysis.DiversificationMetrics.ConcentrationRisk.Mul(decimal.NewFromFloat(100)))
		totalScore = totalScore.Add(concentrationScore.Mul(weights["concentration"]))
		totalWeight = totalWeight.Add(weights["concentration"])
	}

	// Correlation risk
	if analysis.CorrelationAnalysis != nil {
		correlationScore := decimal.NewFromFloat(100).Sub(analysis.CorrelationAnalysis.CorrelationRisk.Mul(decimal.NewFromFloat(100)))
		totalScore = totalScore.Add(correlationScore.Mul(weights["correlation"]))
		totalWeight = totalWeight.Add(weights["correlation"])
	}

	// Volatility risk (from risk metrics)
	if analysis.RiskMetrics != nil {
		volatilityScore := decimal.NewFromFloat(100).Sub(analysis.RiskMetrics.Volatility.Mul(decimal.NewFromFloat(100)))
		if volatilityScore.LessThan(decimal.Zero) {
			volatilityScore = decimal.Zero
		}
		totalScore = totalScore.Add(volatilityScore.Mul(weights["volatility"]))
		totalWeight = totalWeight.Add(weights["volatility"])
	}

	// VaR risk
	if analysis.VaRAnalysis != nil {
		if var95, exists := analysis.VaRAnalysis.HistoricalVaR["95"]; exists {
			varScore := decimal.NewFromFloat(100).Sub(var95.Abs().Mul(decimal.NewFromFloat(100)))
			if varScore.LessThan(decimal.Zero) {
				varScore = decimal.Zero
			}
			totalScore = totalScore.Add(varScore.Mul(weights["var"]))
			totalWeight = totalWeight.Add(weights["var"])
		}
	}

	// Diversification score
	if analysis.DiversificationMetrics != nil {
		totalScore = totalScore.Add(analysis.DiversificationMetrics.DiversificationScore.Mul(weights["diversification"]))
		totalWeight = totalWeight.Add(weights["diversification"])
	}

	// Performance score (based on Sharpe ratio)
	if analysis.RiskMetrics != nil {
		performanceScore := decimal.NewFromFloat(50) // Neutral base
		if analysis.RiskMetrics.SharpeRatio.GreaterThan(decimal.Zero) {
			performanceScore = performanceScore.Add(analysis.RiskMetrics.SharpeRatio.Mul(decimal.NewFromFloat(10)))
		}
		if performanceScore.GreaterThan(decimal.NewFromFloat(100)) {
			performanceScore = decimal.NewFromFloat(100)
		}
		totalScore = totalScore.Add(performanceScore.Mul(weights["performance"]))
		totalWeight = totalWeight.Add(weights["performance"])
	}

	if totalWeight.IsZero() {
		return decimal.NewFromFloat(50) // Default neutral score
	}

	return totalScore.Div(totalWeight)
}

// determineRiskLevel determines the risk level based on score
func (pra *PortfolioRiskAnalyzer) determineRiskLevel(score decimal.Decimal) string {
	if score.GreaterThanOrEqual(decimal.NewFromFloat(80)) {
		return "low"
	} else if score.GreaterThanOrEqual(decimal.NewFromFloat(60)) {
		return "medium"
	} else if score.GreaterThanOrEqual(decimal.NewFromFloat(40)) {
		return "high"
	} else {
		return "critical"
	}
}

// calculateConfidence calculates confidence in the analysis
func (pra *PortfolioRiskAnalyzer) calculateConfidence(analysis *PortfolioRiskAnalysis) decimal.Decimal {
	confidenceFactors := []decimal.Decimal{}

	// Data availability confidence
	if len(analysis.Portfolio.Assets) >= 3 {
		confidenceFactors = append(confidenceFactors, decimal.NewFromFloat(0.9))
	} else {
		confidenceFactors = append(confidenceFactors, decimal.NewFromFloat(0.6))
	}

	// Analysis completeness confidence
	completedAnalyses := 0
	totalAnalyses := 5

	if analysis.CorrelationAnalysis != nil {
		completedAnalyses++
	}
	if analysis.DiversificationMetrics != nil {
		completedAnalyses++
	}
	if analysis.VaRAnalysis != nil {
		completedAnalyses++
	}
	if analysis.RiskMetrics != nil {
		completedAnalyses++
	}
	if analysis.PerformanceMetrics != nil {
		completedAnalyses++
	}

	completenessConfidence := decimal.NewFromInt(int64(completedAnalyses)).Div(decimal.NewFromInt(int64(totalAnalyses)))
	confidenceFactors = append(confidenceFactors, completenessConfidence)

	// Calculate average confidence
	total := decimal.Zero
	for _, factor := range confidenceFactors {
		total = total.Add(factor)
	}

	return total.Div(decimal.NewFromInt(int64(len(confidenceFactors))))
}

// calculatePerformanceMetrics calculates portfolio performance metrics
func (pra *PortfolioRiskAnalyzer) calculatePerformanceMetrics(portfolio *Portfolio) *PerformanceMetrics {
	// Mock performance metrics - in production, use actual historical data
	return &PerformanceMetrics{
		TotalReturn:      decimal.NewFromFloat(0.15),  // 15% total return
		AnnualizedReturn: decimal.NewFromFloat(0.12),  // 12% annualized
		DailyReturns:     []decimal.Decimal{},         // Would contain daily returns
		MonthlyReturns:   []decimal.Decimal{},         // Would contain monthly returns
		YearlyReturns:    []decimal.Decimal{},         // Would contain yearly returns
		BestDay:          decimal.NewFromFloat(0.08),  // 8% best day
		WorstDay:         decimal.NewFromFloat(-0.06), // -6% worst day
		PositiveDays:     180,                         // 180 positive days
		NegativeDays:     120,                         // 120 negative days
		WinRate:          decimal.NewFromFloat(0.6),   // 60% win rate
		Metadata:         make(map[string]interface{}),
	}
}

// generateRebalancingAdvice generates portfolio rebalancing advice
func (pra *PortfolioRiskAnalyzer) generateRebalancingAdvice(analysis *PortfolioRiskAnalysis) *RebalancingAdvice {
	shouldRebalance := false
	rebalanceReason := ""

	// Check if rebalancing is needed based on thresholds
	if pra.config.RebalancingThresholds.Enabled {
		// Check concentration risk
		if analysis.DiversificationMetrics != nil &&
			analysis.DiversificationMetrics.ConcentrationRisk.GreaterThan(pra.config.RebalancingThresholds.DeviationThreshold) {
			shouldRebalance = true
			rebalanceReason = "High concentration risk detected"
		}

		// Check correlation risk
		if analysis.CorrelationAnalysis != nil &&
			analysis.CorrelationAnalysis.CorrelationRisk.GreaterThan(pra.config.RebalancingThresholds.CorrelationThreshold) {
			shouldRebalance = true
			if rebalanceReason != "" {
				rebalanceReason += "; High correlation risk"
			} else {
				rebalanceReason = "High correlation risk detected"
			}
		}
	}

	// Generate target allocations (simplified)
	targetAllocations := make(map[string]decimal.Decimal)
	currentAllocations := make(map[string]decimal.Decimal)

	for _, asset := range analysis.Portfolio.Assets {
		currentAllocations[asset.Symbol] = asset.Weight
		// Simple equal weight target for demonstration
		targetAllocations[asset.Symbol] = decimal.NewFromFloat(1.0).Div(decimal.NewFromInt(int64(len(analysis.Portfolio.Assets))))
	}

	// Generate rebalance actions
	rebalanceActions := []*RebalanceAction{}
	if shouldRebalance {
		for _, asset := range analysis.Portfolio.Assets {
			currentWeight := asset.Weight
			targetWeight := targetAllocations[asset.Symbol]

			if !currentWeight.Equal(targetWeight) {
				action := "hold"
				if currentWeight.GreaterThan(targetWeight) {
					action = "sell"
				} else if currentWeight.LessThan(targetWeight) {
					action = "buy"
				}

				rebalanceActions = append(rebalanceActions, &RebalanceAction{
					Asset:         asset.Symbol,
					Action:        action,
					CurrentWeight: currentWeight,
					TargetWeight:  targetWeight,
					AmountChange:  targetWeight.Sub(currentWeight).Mul(analysis.Portfolio.TotalValue),
					ValueChange:   targetWeight.Sub(currentWeight).Mul(analysis.Portfolio.TotalValue),
					Priority:      1,
					Reason:        "Portfolio rebalancing",
				})
			}
		}
	}

	return &RebalancingAdvice{
		ShouldRebalance:     shouldRebalance,
		RebalanceReason:     rebalanceReason,
		TargetAllocations:   targetAllocations,
		CurrentAllocations:  currentAllocations,
		RebalanceActions:    rebalanceActions,
		ExpectedImprovement: decimal.NewFromFloat(0.02),  // Mock 2% improvement
		RebalanceCost:       decimal.NewFromFloat(0.001), // Mock 0.1% cost
		NetBenefit:          decimal.NewFromFloat(0.019), // Mock net benefit
		Metadata:            make(map[string]interface{}),
	}
}

// generateRiskAlerts generates portfolio risk alerts
func (pra *PortfolioRiskAnalyzer) generateRiskAlerts(analysis *PortfolioRiskAnalysis) []*PortfolioRiskAlert {
	var alerts []*PortfolioRiskAlert

	// Concentration risk alert
	if analysis.DiversificationMetrics != nil &&
		analysis.DiversificationMetrics.ConcentrationRisk.GreaterThan(pra.config.AlertThresholds.MaxConcentration) {
		alerts = append(alerts, &PortfolioRiskAlert{
			ID:        fmt.Sprintf("concentration_%d", time.Now().Unix()),
			Type:      "concentration_risk",
			Severity:  "high",
			Title:     "High Concentration Risk",
			Message:   fmt.Sprintf("Portfolio concentration risk: %s%%", analysis.DiversificationMetrics.ConcentrationRisk.Mul(decimal.NewFromFloat(100)).String()),
			Metric:    "concentration_risk",
			Value:     analysis.DiversificationMetrics.ConcentrationRisk,
			Threshold: pra.config.AlertThresholds.MaxConcentration,
			CreatedAt: time.Now(),
			Actions:   []string{"Diversify holdings", "Reduce largest position"},
			Metadata:  make(map[string]interface{}),
		})
	}

	// Correlation risk alert
	if analysis.CorrelationAnalysis != nil &&
		analysis.CorrelationAnalysis.CorrelationRisk.GreaterThan(pra.config.AlertThresholds.MaxCorrelation) {
		alerts = append(alerts, &PortfolioRiskAlert{
			ID:        fmt.Sprintf("correlation_%d", time.Now().Unix()),
			Type:      "correlation_risk",
			Severity:  "medium",
			Title:     "High Correlation Risk",
			Message:   fmt.Sprintf("Portfolio correlation risk: %s%%", analysis.CorrelationAnalysis.CorrelationRisk.Mul(decimal.NewFromFloat(100)).String()),
			Metric:    "correlation_risk",
			Value:     analysis.CorrelationAnalysis.CorrelationRisk,
			Threshold: pra.config.AlertThresholds.MaxCorrelation,
			CreatedAt: time.Now(),
			Actions:   []string{"Add uncorrelated assets", "Reduce correlated positions"},
			Metadata:  make(map[string]interface{}),
		})
	}

	// VaR alert
	if analysis.VaRAnalysis != nil {
		if var95, exists := analysis.VaRAnalysis.HistoricalVaR["95"]; exists &&
			var95.Abs().GreaterThan(pra.config.AlertThresholds.MaxVaR) {
			alerts = append(alerts, &PortfolioRiskAlert{
				ID:        fmt.Sprintf("var_%d", time.Now().Unix()),
				Type:      "var_risk",
				Severity:  "high",
				Title:     "High Value at Risk",
				Message:   fmt.Sprintf("95%% VaR: %s%%", var95.Abs().Mul(decimal.NewFromFloat(100)).String()),
				Metric:    "var_95",
				Value:     var95.Abs(),
				Threshold: pra.config.AlertThresholds.MaxVaR,
				CreatedAt: time.Now(),
				Actions:   []string{"Reduce position sizes", "Add hedging instruments"},
				Metadata:  make(map[string]interface{}),
			})
		}
	}

	return alerts
}

// generateRecommendations generates portfolio recommendations
func (pra *PortfolioRiskAnalyzer) generateRecommendations(analysis *PortfolioRiskAnalysis) []string {
	var recommendations []string

	// Diversification recommendations
	if analysis.DiversificationMetrics != nil {
		if analysis.DiversificationMetrics.ConcentrationRisk.GreaterThan(decimal.NewFromFloat(0.5)) {
			recommendations = append(recommendations, "Reduce concentration in largest holdings")
		}

		if analysis.DiversificationMetrics.SectorDiversification.LessThan(decimal.NewFromFloat(0.6)) {
			recommendations = append(recommendations, "Improve sector diversification")
		}

		if analysis.DiversificationMetrics.ChainDiversification.LessThan(decimal.NewFromFloat(0.7)) {
			recommendations = append(recommendations, "Diversify across more blockchain networks")
		}
	}

	// Correlation recommendations
	if analysis.CorrelationAnalysis != nil {
		if analysis.CorrelationAnalysis.AverageCorrelation.GreaterThan(decimal.NewFromFloat(0.7)) {
			recommendations = append(recommendations, "Add assets with lower correlation to existing holdings")
		}

		if len(analysis.CorrelationAnalysis.HighlyCorrelatedPairs) > 2 {
			recommendations = append(recommendations, "Consider reducing positions in highly correlated assets")
		}
	}

	// Performance recommendations
	if analysis.RiskMetrics != nil {
		if analysis.RiskMetrics.SharpeRatio.LessThan(decimal.NewFromFloat(1.0)) {
			recommendations = append(recommendations, "Consider strategies to improve risk-adjusted returns")
		}

		if analysis.RiskMetrics.Volatility.GreaterThan(decimal.NewFromFloat(0.3)) {
			recommendations = append(recommendations, "Consider reducing portfolio volatility through diversification")
		}
	}

	// VaR recommendations
	if analysis.VaRAnalysis != nil {
		if analysis.VaRAnalysis.MaxDrawdown.GreaterThan(decimal.NewFromFloat(0.2)) {
			recommendations = append(recommendations, "Implement risk management strategies to limit drawdowns")
		}
	}

	return recommendations
}

// Utility methods

// monitoringLoop runs the main monitoring loop
func (pra *PortfolioRiskAnalyzer) monitoringLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-pra.stopChan:
			return
		case <-pra.updateTicker.C:
			pra.performMaintenance()
		}
	}
}

// performMaintenance performs periodic maintenance tasks
func (pra *PortfolioRiskAnalyzer) performMaintenance() {
	pra.logger.Debug("Performing portfolio risk analyzer maintenance")

	// Clean up expired cache entries
	pra.cleanupExpiredCache()

	// Update price history
	pra.updatePriceHistory()
}

// cleanupExpiredCache removes expired cache entries
func (pra *PortfolioRiskAnalyzer) cleanupExpiredCache() {
	pra.cacheMutex.Lock()
	defer pra.cacheMutex.Unlock()

	now := time.Now()
	for key, analysis := range pra.portfolioCache {
		if now.Sub(analysis.Timestamp) > pra.config.CacheTimeout {
			delete(pra.portfolioCache, key)
		}
	}
}

// updatePriceHistory updates price history data
func (pra *PortfolioRiskAnalyzer) updatePriceHistory() {
	pra.dataMutex.Lock()
	defer pra.dataMutex.Unlock()

	// Mock price history update - in production, fetch from data sources
	pra.logger.Debug("Updating price history data")
}

// generateCacheKey generates a cache key for portfolio analysis
func (pra *PortfolioRiskAnalyzer) generateCacheKey(portfolio *Portfolio) string {
	return fmt.Sprintf("%s_%s_%d", portfolio.ID, portfolio.Address.Hex(), portfolio.LastUpdated.Unix())
}

// generateAnalysisID generates a unique analysis ID
func (pra *PortfolioRiskAnalyzer) generateAnalysisID() string {
	return fmt.Sprintf("portfolio_analysis_%d", time.Now().UnixNano())
}

// getCachedAnalysis retrieves cached portfolio analysis
func (pra *PortfolioRiskAnalyzer) getCachedAnalysis(key string) *PortfolioRiskAnalysis {
	pra.cacheMutex.RLock()
	defer pra.cacheMutex.RUnlock()

	analysis, exists := pra.portfolioCache[key]
	if !exists {
		return nil
	}

	// Check if cache entry is still valid
	if time.Since(analysis.Timestamp) > pra.config.CacheTimeout {
		delete(pra.portfolioCache, key)
		return nil
	}

	return analysis
}

// cacheAnalysis caches portfolio analysis
func (pra *PortfolioRiskAnalyzer) cacheAnalysis(key string, analysis *PortfolioRiskAnalysis) {
	pra.cacheMutex.Lock()
	defer pra.cacheMutex.Unlock()
	pra.portfolioCache[key] = analysis
}

// IsRunning returns whether the analyzer is running
func (pra *PortfolioRiskAnalyzer) IsRunning() bool {
	pra.mutex.RLock()
	defer pra.mutex.RUnlock()
	return pra.isRunning
}

// GetAnalysisMetrics returns analysis metrics
func (pra *PortfolioRiskAnalyzer) GetAnalysisMetrics() map[string]interface{} {
	pra.cacheMutex.RLock()
	defer pra.cacheMutex.RUnlock()

	return map[string]interface{}{
		"cached_analyses":      len(pra.portfolioCache),
		"is_running":           pra.IsRunning(),
		"correlation_analyzer": pra.correlationAnalyzer != nil,
		"diversification_engine": pra.diversificationEngine != nil,
		"var_calculator":       pra.varCalculator != nil,
		"risk_metrics_engine":  pra.riskMetricsEngine != nil,
		"price_history_assets": len(pra.priceHistory),
	}
}

// GetPortfolioSummary returns a summary of portfolio analysis
func (pra *PortfolioRiskAnalyzer) GetPortfolioSummary(ctx context.Context, portfolio *Portfolio) (*PortfolioSummary, error) {
	analysis, err := pra.AnalyzePortfolioRisk(ctx, portfolio)
	if err != nil {
		return nil, err
	}

	return &PortfolioSummary{
		PortfolioID:      portfolio.ID,
		TotalValue:       portfolio.TotalValue,
		AssetCount:       len(portfolio.Assets),
		OverallRiskScore: analysis.OverallRiskScore,
		RiskLevel:        analysis.RiskLevel,
		AlertCount:       len(analysis.RiskAlerts),
		LastAnalyzed:     analysis.Timestamp,
		Confidence:       analysis.Confidence,
	}, nil
}

// PortfolioSummary represents a portfolio summary
type PortfolioSummary struct {
	PortfolioID      string          `json:"portfolio_id"`
	TotalValue       decimal.Decimal `json:"total_value"`
	AssetCount       int             `json:"asset_count"`
	OverallRiskScore decimal.Decimal `json:"overall_risk_score"`
	RiskLevel        string          `json:"risk_level"`
	AlertCount       int             `json:"alert_count"`
	LastAnalyzed     time.Time       `json:"last_analyzed"`
	Confidence       decimal.Decimal `json:"confidence"`
}
