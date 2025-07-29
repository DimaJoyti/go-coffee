package reporting

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

// Mock implementations for testing and demonstration

// MockPnLTracker provides mock P&L tracking
type MockPnLTracker struct{}

func (m *MockPnLTracker) CalculatePnL(ctx context.Context, userID string, timeRange TimeRange) (*PnLData, error) {
	return &PnLData{
		UserID:        userID,
		TimeRange:     timeRange,
		RealizedPnL:   decimal.NewFromFloat(5000),
		UnrealizedPnL: decimal.NewFromFloat(2500),
		TotalPnL:      decimal.NewFromFloat(7500),
		TotalFees:     decimal.NewFromFloat(150),
		NetPnL:        decimal.NewFromFloat(7350),
		ROI:           decimal.NewFromFloat(15.5),
		AssetBreakdown: map[string]AssetPnL{
			"BTC": {
				Asset:         "BTC",
				RealizedPnL:   decimal.NewFromFloat(3000),
				UnrealizedPnL: decimal.NewFromFloat(1500),
				TotalPnL:      decimal.NewFromFloat(4500),
				Fees:          decimal.NewFromFloat(100),
				ROI:           decimal.NewFromFloat(18.0),
				HoldingPeriod: 180 * 24 * time.Hour,
				Transactions:  15,
			},
			"ETH": {
				Asset:         "ETH",
				RealizedPnL:   decimal.NewFromFloat(2000),
				UnrealizedPnL: decimal.NewFromFloat(1000),
				TotalPnL:      decimal.NewFromFloat(3000),
				Fees:          decimal.NewFromFloat(50),
				ROI:           decimal.NewFromFloat(12.5),
				HoldingPeriod: 120 * 24 * time.Hour,
				Transactions:  8,
			},
		},
		Currency:    "USD",
		GeneratedAt: time.Now(),
	}, nil
}

func (m *MockPnLTracker) GetAssetPnL(ctx context.Context, userID string, asset string, timeRange TimeRange) (*AssetPnL, error) {
	return &AssetPnL{
		Asset:         asset,
		RealizedPnL:   decimal.NewFromFloat(1000),
		UnrealizedPnL: decimal.NewFromFloat(500),
		TotalPnL:      decimal.NewFromFloat(1500),
		Fees:          decimal.NewFromFloat(25),
		ROI:           decimal.NewFromFloat(10.0),
		HoldingPeriod: 90 * 24 * time.Hour,
		Transactions:  5,
	}, nil
}

func (m *MockPnLTracker) CalculateTaxImplications(ctx context.Context, userID string, timeRange TimeRange) (*TaxData, error) {
	return &TaxData{
		TaxableEvents: []TaxableEvent{
			{
				Type:            "sale",
				Asset:           "BTC",
				Amount:          decimal.NewFromFloat(0.5),
				CostBasis:       decimal.NewFromFloat(20000),
				FairMarketValue: decimal.NewFromFloat(25000),
				Gain:            decimal.NewFromFloat(5000),
				HoldingPeriod:   400 * 24 * time.Hour,
				IsLongTerm:      true,
				Timestamp:       time.Now().AddDate(0, -1, 0),
			},
		},
		ShortTermGains:        decimal.NewFromFloat(1000),
		LongTermGains:         decimal.NewFromFloat(4000),
		TotalTaxableGains:     decimal.NewFromFloat(5000),
		EstimatedTaxLiability: decimal.NewFromFloat(1200),
		TaxOptimizationTips: []string{
			"Consider tax loss harvesting",
			"Hold assets for more than one year for long-term capital gains",
			"Use tax-advantaged accounts when possible",
		},
	}, nil
}

// MockPerformanceAnalyzer provides mock performance analysis
type MockPerformanceAnalyzer struct{}

func (m *MockPerformanceAnalyzer) AnalyzePerformance(ctx context.Context, userID string, timeRange TimeRange) (*PerformanceData, error) {
	return &PerformanceData{
		UserID:           userID,
		TimeRange:        timeRange,
		TotalReturn:      decimal.NewFromFloat(25.5),
		AnnualizedReturn: decimal.NewFromFloat(18.2),
		Volatility:       decimal.NewFromFloat(35.0),
		SharpeRatio:      decimal.NewFromFloat(1.25),
		SortinoRatio:     decimal.NewFromFloat(1.45),
		MaxDrawdown:      decimal.NewFromFloat(-15.2),
		CalmarRatio:      decimal.NewFromFloat(1.2),
		WinRate:          decimal.NewFromFloat(65.0),
		ProfitFactor:     decimal.NewFromFloat(1.8),
		BenchmarkComparison: map[string]BenchmarkData{
			"BTC": {
				Name:             "Bitcoin",
				Return:           decimal.NewFromFloat(20.0),
				Volatility:       decimal.NewFromFloat(40.0),
				SharpeRatio:      decimal.NewFromFloat(1.1),
				Correlation:      decimal.NewFromFloat(0.85),
				Beta:             decimal.NewFromFloat(1.2),
				Alpha:            decimal.NewFromFloat(5.5),
				TrackingError:    decimal.NewFromFloat(8.5),
				InformationRatio: decimal.NewFromFloat(0.65),
			},
		},
		RiskMetrics: RiskMetrics{
			VaR95:             decimal.NewFromFloat(-5000),
			VaR99:             decimal.NewFromFloat(-8000),
			CVaR95:            decimal.NewFromFloat(-6500),
			CVaR99:            decimal.NewFromFloat(-10000),
			DownsideDeviation: decimal.NewFromFloat(25.0),
			UpsideDeviation:   decimal.NewFromFloat(30.0),
			SkewnessRatio:     decimal.NewFromFloat(0.5),
			KurtosisRatio:     decimal.NewFromFloat(3.2),
			TailRatio:         decimal.NewFromFloat(1.1),
		},
		PerformanceAttribution: PerformanceAttribution{
			AssetAllocation:   decimal.NewFromFloat(3.2),
			SecuritySelection: decimal.NewFromFloat(2.3),
			InteractionEffect: decimal.NewFromFloat(0.5),
			TotalActiveReturn: decimal.NewFromFloat(6.0),
			AssetContributions: map[string]decimal.Decimal{
				"BTC": decimal.NewFromFloat(4.5),
				"ETH": decimal.NewFromFloat(1.5),
			},
		},
		GeneratedAt: time.Now(),
	}, nil
}

func (m *MockPerformanceAnalyzer) CompareToBenchmarks(ctx context.Context, userID string, benchmarks []string, timeRange TimeRange) (map[string]*BenchmarkData, error) {
	result := make(map[string]*BenchmarkData)
	
	for _, benchmark := range benchmarks {
		result[benchmark] = &BenchmarkData{
			Name:             benchmark,
			Return:           decimal.NewFromFloat(15.0),
			Volatility:       decimal.NewFromFloat(30.0),
			SharpeRatio:      decimal.NewFromFloat(1.0),
			Correlation:      decimal.NewFromFloat(0.75),
			Beta:             decimal.NewFromFloat(1.1),
			Alpha:            decimal.NewFromFloat(3.0),
			TrackingError:    decimal.NewFromFloat(5.0),
			InformationRatio: decimal.NewFromFloat(0.6),
		}
	}
	
	return result, nil
}

func (m *MockPerformanceAnalyzer) CalculateRiskMetrics(ctx context.Context, userID string, timeRange TimeRange) (*RiskMetrics, error) {
	return &RiskMetrics{
		VaR95:             decimal.NewFromFloat(-5000),
		VaR99:             decimal.NewFromFloat(-8000),
		CVaR95:            decimal.NewFromFloat(-6500),
		CVaR99:            decimal.NewFromFloat(-10000),
		DownsideDeviation: decimal.NewFromFloat(25.0),
		UpsideDeviation:   decimal.NewFromFloat(30.0),
		SkewnessRatio:     decimal.NewFromFloat(0.5),
		KurtosisRatio:     decimal.NewFromFloat(3.2),
		TailRatio:         decimal.NewFromFloat(1.1),
	}, nil
}

// MockPortfolioVisualizer provides mock portfolio visualization
type MockPortfolioVisualizer struct{}

func (m *MockPortfolioVisualizer) GeneratePortfolioCharts(ctx context.Context, userID string, chartTypes []string) ([]Chart, error) {
	var charts []Chart
	
	for _, chartType := range chartTypes {
		chart := Chart{
			ID:          fmt.Sprintf("chart_%s_%d", chartType, time.Now().Unix()),
			Type:        chartType,
			Title:       fmt.Sprintf("Portfolio %s Chart", chartType),
			Description: fmt.Sprintf("Mock %s chart for portfolio visualization", chartType),
			Data:        map[string]interface{}{"mock": "data"},
			Config: ChartConfig{
				Width:  800,
				Height: 400,
				Theme:  "default",
			},
		}
		charts = append(charts, chart)
	}
	
	return charts, nil
}

func (m *MockPortfolioVisualizer) CreateAllocationChart(ctx context.Context, userID string) (*Chart, error) {
	return &Chart{
		ID:          "allocation_chart",
		Type:        "pie",
		Title:       "Portfolio Allocation",
		Description: "Current portfolio asset allocation",
		Data: map[string]interface{}{
			"labels": []string{"BTC", "ETH", "ADA", "DOT"},
			"values": []float64{45.0, 30.0, 15.0, 10.0},
		},
		Config: ChartConfig{
			Width:  600,
			Height: 400,
			Theme:  "default",
			Colors: []string{"#f7931a", "#627eea", "#0033ad", "#e6007a"},
		},
	}, nil
}

func (m *MockPortfolioVisualizer) CreatePerformanceChart(ctx context.Context, userID string, timeRange TimeRange) (*Chart, error) {
	return &Chart{
		ID:          "performance_chart",
		Type:        "line",
		Title:       "Portfolio Performance",
		Description: "Portfolio value over time",
		Data: map[string]interface{}{
			"timestamps": []string{"2023-01-01", "2023-02-01", "2023-03-01", "2023-04-01"},
			"values":     []float64{100000, 105000, 98000, 115000},
		},
		Config: ChartConfig{
			Width:  800,
			Height: 400,
			Theme:  "default",
			Axes: map[string]AxisConfig{
				"x": {Label: "Date", Type: "time"},
				"y": {Label: "Portfolio Value ($)", Type: "linear"},
			},
		},
	}, nil
}

// MockRiskAnalyzer provides mock risk analysis
type MockRiskAnalyzer struct{}

func (m *MockRiskAnalyzer) AnalyzeRisk(ctx context.Context, userID string) (*RiskAnalysis, error) {
	return &RiskAnalysis{
		OverallRiskScore: decimal.NewFromFloat(6.5),
		RiskFactors: map[string]decimal.Decimal{
			"market_risk":       decimal.NewFromFloat(7.0),
			"liquidity_risk":    decimal.NewFromFloat(5.0),
			"concentration_risk": decimal.NewFromFloat(8.0),
			"counterparty_risk": decimal.NewFromFloat(4.0),
		},
		ConcentrationRisk: decimal.NewFromFloat(8.0),
		LiquidityRisk:     decimal.NewFromFloat(5.0),
		MarketRisk:        decimal.NewFromFloat(7.0),
		CounterpartyRisk:  decimal.NewFromFloat(4.0),
		Recommendations: []string{
			"Consider diversifying portfolio to reduce concentration risk",
			"Monitor market conditions for increased volatility",
			"Maintain adequate liquidity reserves",
		},
	}, nil
}

func (m *MockRiskAnalyzer) RunStressTests(ctx context.Context, userID string, scenarios []string) (map[string]*StressTestResult, error) {
	results := make(map[string]*StressTestResult)
	
	for _, scenario := range scenarios {
		results[scenario] = &StressTestResult{
			Scenario:        scenario,
			PortfolioImpact: decimal.NewFromFloat(-15.0),
			AssetImpacts: map[string]decimal.Decimal{
				"BTC": decimal.NewFromFloat(-20.0),
				"ETH": decimal.NewFromFloat(-18.0),
				"ADA": decimal.NewFromFloat(-12.0),
			},
			RecoveryTime: 90 * 24 * time.Hour,
			Probability:  decimal.NewFromFloat(0.15),
		}
	}
	
	return results, nil
}

func (m *MockRiskAnalyzer) CalculateCorrelations(ctx context.Context, userID string) (*CorrelationMatrix, error) {
	return &CorrelationMatrix{
		Assets: []string{"BTC", "ETH", "ADA"},
		Matrix: [][]decimal.Decimal{
			{decimal.NewFromFloat(1.0), decimal.NewFromFloat(0.75), decimal.NewFromFloat(0.65)},
			{decimal.NewFromFloat(0.75), decimal.NewFromFloat(1.0), decimal.NewFromFloat(0.70)},
			{decimal.NewFromFloat(0.65), decimal.NewFromFloat(0.70), decimal.NewFromFloat(1.0)},
		},
		TimeRange: TimeRange{
			Start:       time.Now().AddDate(0, -3, 0),
			End:         time.Now(),
			Granularity: "day",
		},
		CalculatedAt: time.Now(),
	}, nil
}

// MockDataAggregator provides mock data aggregation
type MockDataAggregator struct{}

func (m *MockDataAggregator) AggregateData(ctx context.Context, userID string, timeRange TimeRange, granularity string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"portfolio_values": []float64{100000, 105000, 98000, 115000},
		"timestamps":       []string{"2023-01-01", "2023-02-01", "2023-03-01", "2023-04-01"},
		"granularity":      granularity,
		"user_id":          userID,
	}, nil
}

func (m *MockDataAggregator) GetHistoricalData(ctx context.Context, userID string, dataType string, timeRange TimeRange) (interface{}, error) {
	switch dataType {
	case "portfolio":
		return []PortfolioSnapshot{
			{
				UserID:     userID,
				Timestamp:  time.Now().AddDate(0, -1, 0),
				TotalValue: decimal.NewFromFloat(100000),
				Holdings: map[string]decimal.Decimal{
					"BTC": decimal.NewFromFloat(2.0),
					"ETH": decimal.NewFromFloat(50.0),
				},
			},
		}, nil
	case "transactions":
		return []Transaction{
			{
				Hash:        common.HexToHash("0x123"),
				Type:        "buy",
				Asset:       "BTC",
				Amount:      decimal.NewFromFloat(1.0),
				Price:       decimal.NewFromFloat(50000),
				Fee:         decimal.NewFromFloat(25),
				Timestamp:   time.Now().AddDate(0, -1, 0),
				BlockNumber: 12345,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported data type: %s", dataType)
	}
}

func (m *MockDataAggregator) RefreshData(ctx context.Context, userID string) error {
	return nil
}

// MockMetricCalculator provides mock metric calculation
type MockMetricCalculator struct{}

func (m *MockMetricCalculator) CalculateMetrics(ctx context.Context, userID string, metricTypes []string) (map[string]decimal.Decimal, error) {
	metrics := make(map[string]decimal.Decimal)
	
	for _, metricType := range metricTypes {
		switch metricType {
		case "total_return":
			metrics[metricType] = decimal.NewFromFloat(25.5)
		case "sharpe_ratio":
			metrics[metricType] = decimal.NewFromFloat(1.25)
		case "max_drawdown":
			metrics[metricType] = decimal.NewFromFloat(-15.2)
		case "volatility":
			metrics[metricType] = decimal.NewFromFloat(35.0)
		case "roi":
			metrics[metricType] = decimal.NewFromFloat(18.2)
		default:
			metrics[metricType] = decimal.NewFromFloat(10.0)
		}
	}
	
	return metrics, nil
}

func (m *MockMetricCalculator) CalculateCustomMetric(ctx context.Context, userID string, formula string) (decimal.Decimal, error) {
	// Mock custom metric calculation
	return decimal.NewFromFloat(15.5), nil
}

func (m *MockMetricCalculator) GetMetricHistory(ctx context.Context, userID string, metric string, timeRange TimeRange) ([]MetricDataPoint, error) {
	return []MetricDataPoint{
		{
			Timestamp: time.Now().AddDate(0, -3, 0),
			Value:     decimal.NewFromFloat(10.0),
			Metadata:  map[string]interface{}{"source": "calculated"},
		},
		{
			Timestamp: time.Now().AddDate(0, -2, 0),
			Value:     decimal.NewFromFloat(12.5),
			Metadata:  map[string]interface{}{"source": "calculated"},
		},
		{
			Timestamp: time.Now().AddDate(0, -1, 0),
			Value:     decimal.NewFromFloat(15.0),
			Metadata:  map[string]interface{}{"source": "calculated"},
		},
	}, nil
}

// MockReportGenerator provides mock report generation
type MockReportGenerator struct{}

func (m *MockReportGenerator) GenerateReport(ctx context.Context, userID string, reportType string, config ReportConfig) (*Report, error) {
	return &Report{
		ID:          fmt.Sprintf("report_%d", time.Now().Unix()),
		UserID:      userID,
		Type:        reportType,
		Title:       fmt.Sprintf("%s Report", reportType),
		Description: fmt.Sprintf("Comprehensive %s analysis report", reportType),
		TimeRange:   config.TimeRange,
		Data:        make(map[string]interface{}),
		Charts:      []Chart{},
		Metrics:     make(map[string]decimal.Decimal),
		Insights:    []Insight{},
		Recommendations: []Recommendation{},
		Format:      config.Format,
		Status:      "completed",
		GeneratedAt: time.Now(),
		Metadata:    map[string]interface{}{"version": "1.0"},
	}, nil
}

func (m *MockReportGenerator) GenerateScheduledReports(ctx context.Context) error {
	return nil
}

func (m *MockReportGenerator) ExportReport(ctx context.Context, reportID string, format string) ([]byte, error) {
	switch format {
	case "pdf":
		return []byte("mock pdf content"), nil
	case "csv":
		return []byte("mock csv content"), nil
	case "json":
		return []byte(`{"mock": "json content"}`), nil
	default:
		return []byte("mock content"), nil
	}
}

// MockChartGenerator provides mock chart generation
type MockChartGenerator struct{}

func (m *MockChartGenerator) GenerateChart(ctx context.Context, chartType string, data interface{}, config ChartConfig) (*Chart, error) {
	return &Chart{
		ID:          fmt.Sprintf("chart_%d", time.Now().Unix()),
		Type:        chartType,
		Title:       fmt.Sprintf("Mock %s Chart", chartType),
		Description: fmt.Sprintf("Generated %s chart", chartType),
		Data:        data,
		Config:      config,
	}, nil
}

func (m *MockChartGenerator) ExportChart(ctx context.Context, chart *Chart, format string, resolution string) ([]byte, error) {
	return []byte(fmt.Sprintf("mock chart export in %s format at %s resolution", format, resolution)), nil
}

func (m *MockChartGenerator) CreateInteractiveChart(ctx context.Context, chartType string, data interface{}) (*Chart, error) {
	chart, err := m.GenerateChart(ctx, chartType, data, ChartConfig{})
	if err != nil {
		return nil, err
	}
	
	chart.InteractiveElements = []InteractiveElement{
		{
			Type:    "zoom",
			Enabled: true,
			Config:  map[string]interface{}{"zoomType": "xy"},
		},
		{
			Type:    "pan",
			Enabled: true,
			Config:  map[string]interface{}{"panKey": "shift"},
		},
	}
	
	return chart, nil
}

// MockTransactionProvider provides mock transaction data
type MockTransactionProvider struct{}

func (m *MockTransactionProvider) GetTransactions(ctx context.Context, userID string, timeRange TimeRange) ([]Transaction, error) {
	return []Transaction{
		{
			Hash:        common.HexToHash("0x123"),
			Type:        "buy",
			Asset:       "BTC",
			Amount:      decimal.NewFromFloat(1.0),
			Price:       decimal.NewFromFloat(50000),
			Fee:         decimal.NewFromFloat(25),
			Timestamp:   time.Now().AddDate(0, -1, 0),
			BlockNumber: 12345,
		},
		{
			Hash:        common.HexToHash("0x456"),
			Type:        "sell",
			Asset:       "ETH",
			Amount:      decimal.NewFromFloat(10.0),
			Price:       decimal.NewFromFloat(3000),
			Fee:         decimal.NewFromFloat(15),
			Timestamp:   time.Now().AddDate(0, 0, -15),
			BlockNumber: 12350,
		},
	}, nil
}

func (m *MockTransactionProvider) GetTransactionsByAsset(ctx context.Context, userID string, asset string, timeRange TimeRange) ([]Transaction, error) {
	transactions, err := m.GetTransactions(ctx, userID, timeRange)
	if err != nil {
		return nil, err
	}
	
	var filtered []Transaction
	for _, tx := range transactions {
		if tx.Asset == asset {
			filtered = append(filtered, tx)
		}
	}
	
	return filtered, nil
}

// MockPriceProvider provides mock price data
type MockPriceProvider struct{}

func (m *MockPriceProvider) GetHistoricalPrices(ctx context.Context, assets []string, timeRange TimeRange) (map[string][]PricePoint, error) {
	result := make(map[string][]PricePoint)
	
	for _, asset := range assets {
		result[asset] = []PricePoint{
			{
				Timestamp: time.Now().AddDate(0, -1, 0),
				Price:     decimal.NewFromFloat(50000),
				Volume:    decimal.NewFromFloat(1000000),
			},
			{
				Timestamp: time.Now().AddDate(0, 0, -15),
				Price:     decimal.NewFromFloat(52000),
				Volume:    decimal.NewFromFloat(1200000),
			},
		}
	}
	
	return result, nil
}

func (m *MockPriceProvider) GetCurrentPrices(ctx context.Context, assets []string) (map[string]decimal.Decimal, error) {
	result := make(map[string]decimal.Decimal)
	
	for _, asset := range assets {
		switch asset {
		case "BTC":
			result[asset] = decimal.NewFromFloat(55000)
		case "ETH":
			result[asset] = decimal.NewFromFloat(3200)
		default:
			result[asset] = decimal.NewFromFloat(100)
		}
	}
	
	return result, nil
}

// MockPortfolioProvider provides mock portfolio data
type MockPortfolioProvider struct{}

func (m *MockPortfolioProvider) GetPortfolioSnapshot(ctx context.Context, userID string, timestamp time.Time) (*PortfolioSnapshot, error) {
	return &PortfolioSnapshot{
		UserID:     userID,
		Timestamp:  timestamp,
		TotalValue: decimal.NewFromFloat(115000),
		Holdings: map[string]decimal.Decimal{
			"BTC": decimal.NewFromFloat(2.0),
			"ETH": decimal.NewFromFloat(50.0),
			"ADA": decimal.NewFromFloat(10000),
		},
		Allocations: map[string]decimal.Decimal{
			"BTC": decimal.NewFromFloat(45.0),
			"ETH": decimal.NewFromFloat(35.0),
			"ADA": decimal.NewFromFloat(20.0),
		},
	}, nil
}

func (m *MockPortfolioProvider) GetPortfolioHistory(ctx context.Context, userID string, timeRange TimeRange) ([]PortfolioSnapshot, error) {
	return []PortfolioSnapshot{
		{
			UserID:     userID,
			Timestamp:  time.Now().AddDate(0, -3, 0),
			TotalValue: decimal.NewFromFloat(100000),
		},
		{
			UserID:     userID,
			Timestamp:  time.Now().AddDate(0, -2, 0),
			TotalValue: decimal.NewFromFloat(105000),
		},
		{
			UserID:     userID,
			Timestamp:  time.Now().AddDate(0, -1, 0),
			TotalValue: decimal.NewFromFloat(110000),
		},
		{
			UserID:     userID,
			Timestamp:  time.Now(),
			TotalValue: decimal.NewFromFloat(115000),
		},
	}, nil
}

// MockReportCache provides mock report caching
type MockReportCache struct{}

func (m *MockReportCache) GetReport(ctx context.Context, key string) (*Report, error) {
	// Mock cache miss
	return nil, fmt.Errorf("cache miss")
}

func (m *MockReportCache) SetReport(ctx context.Context, key string, report *Report, ttl time.Duration) error {
	// Mock cache set
	return nil
}

func (m *MockReportCache) InvalidateUserReports(ctx context.Context, userID string) error {
	// Mock cache invalidation
	return nil
}

// MockDataWarehouse provides mock data warehouse functionality
type MockDataWarehouse struct{}

func (m *MockDataWarehouse) StoreReportData(ctx context.Context, data interface{}) error {
	// Mock data storage
	return nil
}

func (m *MockDataWarehouse) QueryReportData(ctx context.Context, query string, params map[string]interface{}) (interface{}, error) {
	// Mock query execution
	return []*Report{}, nil
}

func (m *MockDataWarehouse) ArchiveOldData(ctx context.Context, cutoffDate time.Time) error {
	// Mock data archival
	return nil
}
