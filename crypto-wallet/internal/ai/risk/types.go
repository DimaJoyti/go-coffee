package risk

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

// RiskAssessmentRequest represents a request for risk assessment
type RiskAssessmentRequest struct {
	Address                common.Address         `json:"address"`
	AssessmentType         string                 `json:"assessment_type"`
	Transaction            *TransactionData       `json:"transaction,omitempty"`
	Portfolio              *PortfolioData         `json:"portfolio,omitempty"`
	Assets                 []string               `json:"assets,omitempty"`
	ContractAddress        common.Address         `json:"contract_address,omitempty"`
	SourceCode             string                 `json:"source_code,omitempty"`
	TimeFrame              time.Duration          `json:"time_frame"`
	IncludeTransactionRisk bool                   `json:"include_transaction_risk"`
	IncludePortfolioRisk   bool                   `json:"include_portfolio_risk"`
	IncludeVolatilityRisk  bool                   `json:"include_volatility_risk"`
	IncludeContractRisk    bool                   `json:"include_contract_risk"`
	IncludeMarketRisk      bool                   `json:"include_market_risk"`
	Metadata               map[string]interface{} `json:"metadata"`
}

// TransactionData represents transaction data for risk assessment
type TransactionData struct {
	Hash        string          `json:"hash"`
	From        common.Address  `json:"from"`
	To          common.Address  `json:"to"`
	Value       decimal.Decimal `json:"value"`
	GasPrice    decimal.Decimal `json:"gas_price"`
	GasLimit    uint64          `json:"gas_limit"`
	Data        []byte          `json:"data"`
	Timestamp   time.Time       `json:"timestamp"`
	BlockNumber uint64          `json:"block_number"`
}

// PortfolioData represents portfolio data for risk assessment
type PortfolioData struct {
	TotalValue  decimal.Decimal          `json:"total_value"`
	Assets      map[string]*AssetHolding `json:"assets"`
	Chains      []string                 `json:"chains"`
	LastUpdated time.Time                `json:"last_updated"`
}

// AssetHolding represents an asset holding in a portfolio
type AssetHolding struct {
	Symbol      string          `json:"symbol"`
	Amount      decimal.Decimal `json:"amount"`
	Value       decimal.Decimal `json:"value"`
	Percentage  decimal.Decimal `json:"percentage"`
	Chain       string          `json:"chain"`
	LastUpdated time.Time       `json:"last_updated"`
}

// TransactionRiskConfig holds configuration for transaction risk scoring
type TransactionRiskConfig struct {
	Enabled          bool               `json:"enabled" yaml:"enabled"`
	ModelPath        string             `json:"model_path" yaml:"model_path"`
	FeatureWeights   map[string]float64 `json:"feature_weights" yaml:"feature_weights"`
	AddressWhitelist []string           `json:"address_whitelist" yaml:"address_whitelist"`
	AddressBlacklist []string           `json:"address_blacklist" yaml:"address_blacklist"`
	ThresholdHigh    decimal.Decimal    `json:"threshold_high" yaml:"threshold_high"`
	ThresholdMedium  decimal.Decimal    `json:"threshold_medium" yaml:"threshold_medium"`
	UpdateInterval   time.Duration      `json:"update_interval" yaml:"update_interval"`
	DataSources      []string           `json:"data_sources" yaml:"data_sources"`
}

// VolatilityConfig holds configuration for volatility analysis
type VolatilityConfig struct {
	Enabled          bool                       `json:"enabled" yaml:"enabled"`
	WindowSize       int                        `json:"window_size" yaml:"window_size"`
	ConfidenceLevel  decimal.Decimal            `json:"confidence_level" yaml:"confidence_level"`
	UpdateInterval   time.Duration              `json:"update_interval" yaml:"update_interval"`
	DataSources      []string                   `json:"data_sources" yaml:"data_sources"`
	VolatilityModels []string                   `json:"volatility_models" yaml:"volatility_models"`
	RiskThresholds   map[string]decimal.Decimal `json:"risk_thresholds" yaml:"risk_thresholds"`
}

// ContractAuditConfig holds configuration for contract auditing
type ContractAuditConfig struct {
	Enabled          bool          `json:"enabled" yaml:"enabled"`
	AuditRules       []string      `json:"audit_rules" yaml:"audit_rules"`
	SecurityPatterns []string      `json:"security_patterns" yaml:"security_patterns"`
	VulnerabilityDB  string        `json:"vulnerability_db" yaml:"vulnerability_db"`
	UpdateInterval   time.Duration `json:"update_interval" yaml:"update_interval"`
	TimeoutDuration  time.Duration `json:"timeout_duration" yaml:"timeout_duration"`
	MaxCodeSize      int           `json:"max_code_size" yaml:"max_code_size"`
}

// PortfolioRiskConfig holds configuration for portfolio risk assessment
type PortfolioRiskConfig struct {
	Enabled           bool            `json:"enabled" yaml:"enabled"`
	RiskModels        []string        `json:"risk_models" yaml:"risk_models"`
	CorrelationWindow int             `json:"correlation_window" yaml:"correlation_window"`
	VaRConfidence     decimal.Decimal `json:"var_confidence" yaml:"var_confidence"`
	UpdateInterval    time.Duration   `json:"update_interval" yaml:"update_interval"`
	DataSources       []string        `json:"data_sources" yaml:"data_sources"`
	RiskFactors       []string        `json:"risk_factors" yaml:"risk_factors"`
}

// MarketPredictionConfig holds configuration for market prediction
type MarketPredictionConfig struct {
	Enabled             bool            `json:"enabled" yaml:"enabled"`
	ModelTypes          []string        `json:"model_types" yaml:"model_types"`
	PredictionHorizon   time.Duration   `json:"prediction_horizon" yaml:"prediction_horizon"`
	UpdateInterval      time.Duration   `json:"update_interval" yaml:"update_interval"`
	DataSources         []string        `json:"data_sources" yaml:"data_sources"`
	Features            []string        `json:"features" yaml:"features"`
	ConfidenceThreshold decimal.Decimal `json:"confidence_threshold" yaml:"confidence_threshold"`
}

// TransactionRiskRequest represents a transaction risk assessment request
type TransactionRiskRequest struct {
	Address     common.Address   `json:"address"`
	Transaction *TransactionData `json:"transaction"`
}

// TransactionRiskResult represents transaction risk assessment result
type TransactionRiskResult struct {
	RiskScore       decimal.Decimal        `json:"risk_score"`
	RiskLevel       string                 `json:"risk_level"`
	RiskFactors     []string               `json:"risk_factors"`
	AddressRisk     decimal.Decimal        `json:"address_risk"`
	TransactionRisk decimal.Decimal        `json:"transaction_risk"`
	Confidence      decimal.Decimal        `json:"confidence"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// VolatilityRequest represents a volatility analysis request
type VolatilityRequest struct {
	Assets    []string      `json:"assets"`
	TimeFrame time.Duration `json:"time_frame"`
}

// VolatilityResult represents volatility analysis result
type VolatilityResult struct {
	RiskScore         decimal.Decimal                       `json:"risk_score"`
	Volatility        decimal.Decimal                       `json:"volatility"`
	VaR               decimal.Decimal                       `json:"var"`
	ExpectedShortfall decimal.Decimal                       `json:"expected_shortfall"`
	VolatilityTrend   string                                `json:"volatility_trend"`
	Confidence        decimal.Decimal                       `json:"confidence"`
	AssetVolatilities map[string]decimal.Decimal            `json:"asset_volatilities"`
	Correlations      map[string]map[string]decimal.Decimal `json:"correlations"`
	Metadata          map[string]interface{}                `json:"metadata"`
}

// ContractAuditRequest represents a contract audit request
type ContractAuditRequest struct {
	ContractAddress common.Address `json:"contract_address"`
	SourceCode      string         `json:"source_code"`
}

// ContractAuditResult represents contract audit result
type ContractAuditResult struct {
	SecurityScore    decimal.Decimal        `json:"security_score"`
	RiskLevel        string                 `json:"risk_level"`
	Vulnerabilities  []Vulnerability        `json:"vulnerabilities"`
	SecurityPatterns []SecurityPattern      `json:"security_patterns"`
	GasOptimization  []GasOptimization      `json:"gas_optimization"`
	Confidence       decimal.Decimal        `json:"confidence"`
	AuditTimestamp   time.Time              `json:"audit_timestamp"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Severity    string          `json:"severity"`
	Description string          `json:"description"`
	Location    string          `json:"location"`
	Impact      decimal.Decimal `json:"impact"`
	Confidence  decimal.Decimal `json:"confidence"`
	Remediation string          `json:"remediation"`
}

// Note: SecurityPattern type is defined in contract_analyzer.go

// GasOptimization represents a gas optimization suggestion
type GasOptimization struct {
	Type        string          `json:"type"`
	Description string          `json:"description"`
	Savings     decimal.Decimal `json:"savings"`
	Confidence  decimal.Decimal `json:"confidence"`
}

// PortfolioRiskRequest represents a portfolio risk assessment request
type PortfolioRiskRequest struct {
	Address   common.Address `json:"address"`
	Portfolio *PortfolioData `json:"portfolio"`
}

// PortfolioRiskResult represents portfolio risk assessment result
type PortfolioRiskResult struct {
	OverallRiskScore     decimal.Decimal            `json:"overall_risk_score"`
	RiskLevel            string                     `json:"risk_level"`
	ConcentrationRisk    decimal.Decimal            `json:"concentration_risk"`
	CorrelationRisk      decimal.Decimal            `json:"correlation_risk"`
	LiquidityRisk        decimal.Decimal            `json:"liquidity_risk"`
	VaR                  decimal.Decimal            `json:"var"`
	ExpectedShortfall    decimal.Decimal            `json:"expected_shortfall"`
	SharpeRatio          decimal.Decimal            `json:"sharpe_ratio"`
	MaxDrawdown          decimal.Decimal            `json:"max_drawdown"`
	DiversificationRatio decimal.Decimal            `json:"diversification_ratio"`
	RiskContributions    map[string]decimal.Decimal `json:"risk_contributions"`
	Recommendations      []string                   `json:"recommendations"`
	Metadata             map[string]interface{}     `json:"metadata"`
}

// MarketPredictionRequest represents a market prediction request
type MarketPredictionRequest struct {
	Assets    []string      `json:"assets"`
	TimeFrame time.Duration `json:"time_frame"`
}

// MarketPredictionResult represents market prediction result
type MarketPredictionResult struct {
	RiskScore           decimal.Decimal            `json:"risk_score"`
	MarketDirection     string                     `json:"market_direction"`
	PredictedReturns    map[string]decimal.Decimal `json:"predicted_returns"`
	PredictedVolatility map[string]decimal.Decimal `json:"predicted_volatility"`
	SentimentScore      decimal.Decimal            `json:"sentiment_score"`
	TechnicalScore      decimal.Decimal            `json:"technical_score"`
	FundamentalScore    decimal.Decimal            `json:"fundamental_score"`
	Confidence          decimal.Decimal            `json:"confidence"`
	PredictionHorizon   time.Duration              `json:"prediction_horizon"`
	Metadata            map[string]interface{}     `json:"metadata"`
}

// RiskComponent interface for all risk assessment components
type RiskComponent interface {
	Start(ctx context.Context) error
	Stop() error
	IsHealthy() bool
	GetMetrics() map[string]interface{}
}

// MLModel interface for machine learning models
type MLModel interface {
	LoadModel(path string) error
	Predict(features map[string]interface{}) (decimal.Decimal, error)
	UpdateModel(data []map[string]interface{}) error
	GetModelInfo() map[string]interface{}
}

// DataProvider interface for data providers
type DataProvider interface {
	GetData(request DataRequest) (*DataResponse, error)
	IsAvailable() bool
	GetRateLimit() int
}

// DataRequest represents a data request
type DataRequest struct {
	Type      string                 `json:"type"`
	Assets    []string               `json:"assets"`
	TimeFrame time.Duration          `json:"time_frame"`
	Params    map[string]interface{} `json:"params"`
}

// DataResponse represents a data response
type DataResponse struct {
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Quality   decimal.Decimal        `json:"quality"`
}

// RiskEvent represents a risk event
type RiskEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Address     common.Address         `json:"address"`
	Description string                 `json:"description"`
	RiskScore   decimal.Decimal        `json:"risk_score"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// SystemHealth represents system health status
type SystemHealth struct {
	OverallStatus      string                 `json:"overall_status"`
	ComponentStatuses  map[string]string      `json:"component_statuses"`
	LastHealthCheck    time.Time              `json:"last_health_check"`
	Uptime             time.Duration          `json:"uptime"`
	ErrorRate          decimal.Decimal        `json:"error_rate"`
	PerformanceMetrics map[string]interface{} `json:"performance_metrics"`
}
