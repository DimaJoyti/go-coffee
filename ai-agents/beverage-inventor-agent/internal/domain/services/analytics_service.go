package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/entities"
)

// AnalyticsService provides comprehensive analytics and reporting for beverages
type AnalyticsService struct {
	dataStore    AnalyticsDataStore
	aiProvider   AnalyticsAIProvider
	logger       Logger
}

// AnalyticsDataStore defines the interface for analytics data storage
type AnalyticsDataStore interface {
	StoreBeverageMetrics(ctx context.Context, metrics *BeverageMetrics) error
	GetBeverageMetrics(ctx context.Context, beverageID string, timeframe time.Duration) ([]*BeverageMetrics, error)
	GetPopularityTrends(ctx context.Context, timeframe time.Duration) (*PopularityTrends, error)
	GetPerformanceMetrics(ctx context.Context, beverageIDs []string, timeframe time.Duration) ([]*PerformanceMetrics, error)
	GetCustomerSegmentData(ctx context.Context, timeframe time.Duration) (*CustomerSegmentData, error)
	GetSeasonalTrends(ctx context.Context, beverageType string, years int) (*SeasonalTrends, error)
}

// AnalyticsAIProvider defines AI capabilities for analytics
type AnalyticsAIProvider interface {
	PredictBeverageSuccess(ctx context.Context, beverage *entities.Beverage, marketContext *MarketContext) (*SuccessPrediction, error)
	AnalyzeCustomerPreferences(ctx context.Context, customerData *CustomerSegmentData) (*PreferenceAnalysis, error)
	GenerateInsights(ctx context.Context, metrics []*BeverageMetrics) (*AnalyticsInsights, error)
	ForecastDemand(ctx context.Context, beverage *entities.Beverage, timeframe time.Duration) (*DemandForecast, error)
}

// BeverageMetrics contains comprehensive metrics for a beverage
type BeverageMetrics struct {
	BeverageID       string                 `json:"beverage_id"`
	Timestamp        time.Time              `json:"timestamp"`
	SalesVolume      float64                `json:"sales_volume"`
	Revenue          float64                `json:"revenue"`
	CustomerRating   float64                `json:"customer_rating"`
	OrderCount       int                    `json:"order_count"`
	ReturnRate       float64                `json:"return_rate"`
	ProductionCost   float64                `json:"production_cost"`
	ProfitMargin     float64                `json:"profit_margin"`
	QualityScore     float64                `json:"quality_score"`
	InventoryTurnover float64               `json:"inventory_turnover"`
	CustomerSegments map[string]float64     `json:"customer_segments"` // segment -> percentage
	Channels         map[string]float64     `json:"channels"`          // channel -> sales percentage
	Regions          map[string]float64     `json:"regions"`           // region -> sales percentage
	Seasonality      float64                `json:"seasonality"`       // seasonal factor
	CompetitorIndex  float64                `json:"competitor_index"`  // vs competitors
	CustomMetrics    map[string]interface{} `json:"custom_metrics"`
}

// PopularityTrends contains popularity trend data
type PopularityTrends struct {
	Timeframe     time.Duration           `json:"timeframe"`
	TopBeverages  []BeverageRanking       `json:"top_beverages"`
	TrendingUp    []BeverageRanking       `json:"trending_up"`
	TrendingDown  []BeverageRanking       `json:"trending_down"`
	NewEntries    []BeverageRanking       `json:"new_entries"`
	Categories    map[string]CategoryTrend `json:"categories"`
	GeneratedAt   time.Time               `json:"generated_at"`
}

// BeverageRanking contains ranking information for a beverage
type BeverageRanking struct {
	BeverageID    string  `json:"beverage_id"`
	Name          string  `json:"name"`
	Rank          int     `json:"rank"`
	Score         float64 `json:"score"`
	Change        float64 `json:"change"`        // change from previous period
	ChangePercent float64 `json:"change_percent"`
	Trend         string  `json:"trend"`         // up, down, stable
}

// CategoryTrend contains trend data for a beverage category
type CategoryTrend struct {
	Category      string  `json:"category"`
	Growth        float64 `json:"growth"`        // percentage growth
	MarketShare   float64 `json:"market_share"`  // percentage of total market
	TopPerformers []string `json:"top_performers"`
	Trend         string  `json:"trend"`
}

// PerformanceMetrics contains performance metrics for beverages
type PerformanceMetrics struct {
	BeverageID      string                 `json:"beverage_id"`
	Name            string                 `json:"name"`
	Timeframe       time.Duration          `json:"timeframe"`
	TotalRevenue    float64                `json:"total_revenue"`
	TotalVolume     float64                `json:"total_volume"`
	AverageRating   float64                `json:"average_rating"`
	CustomerCount   int                    `json:"customer_count"`
	RepeatRate      float64                `json:"repeat_rate"`
	ConversionRate  float64                `json:"conversion_rate"`
	ChurnRate       float64                `json:"churn_rate"`
	LifetimeValue   float64                `json:"lifetime_value"`
	AcquisitionCost float64                `json:"acquisition_cost"`
	ROI             float64                `json:"roi"`
	Benchmarks      map[string]float64     `json:"benchmarks"`      // vs industry benchmarks
	KPIs            map[string]interface{} `json:"kpis"`
}

// CustomerSegmentData contains customer segmentation data
type CustomerSegmentData struct {
	Timeframe time.Duration                    `json:"timeframe"`
	Segments  map[string]*CustomerSegment      `json:"segments"`
	Behaviors map[string]*CustomerBehavior     `json:"behaviors"`
	Preferences map[string]*CustomerPreference `json:"preferences"`
	Demographics map[string]*DemographicData   `json:"demographics"`
	GeneratedAt time.Time                      `json:"generated_at"`
}

// CustomerSegment contains data for a customer segment
type CustomerSegment struct {
	Name            string             `json:"name"`
	Size            int                `json:"size"`
	Percentage      float64            `json:"percentage"`
	Revenue         float64            `json:"revenue"`
	AverageSpend    float64            `json:"average_spend"`
	Frequency       float64            `json:"frequency"`
	Loyalty         float64            `json:"loyalty"`
	Characteristics map[string]string  `json:"characteristics"`
	TopBeverages    []string           `json:"top_beverages"`
	Growth          float64            `json:"growth"`
}

// CustomerBehavior contains customer behavior patterns
type CustomerBehavior struct {
	Segment           string             `json:"segment"`
	PurchaseFrequency float64            `json:"purchase_frequency"`
	AverageOrderSize  float64            `json:"average_order_size"`
	PreferredChannels []string           `json:"preferred_channels"`
	PeakTimes         []string           `json:"peak_times"`
	SeasonalPatterns  map[string]float64 `json:"seasonal_patterns"`
	Pricesensitivity float64            `json:"price_sensitivity"`
	BrandLoyalty      float64            `json:"brand_loyalty"`
}

// CustomerPreference contains customer preference data
type CustomerPreference struct {
	Segment         string             `json:"segment"`
	FlavorProfiles  map[string]float64 `json:"flavor_profiles"`  // flavor -> preference score
	Ingredients     map[string]float64 `json:"ingredients"`      // ingredient -> preference score
	Themes          map[string]float64 `json:"themes"`           // theme -> preference score
	PriceRanges     map[string]float64 `json:"price_ranges"`     // range -> preference score
	HealthFactors   map[string]float64 `json:"health_factors"`   // factor -> importance score
	Occasions       map[string]float64 `json:"occasions"`        // occasion -> preference score
}

// DemographicData contains demographic information
type DemographicData struct {
	AgeGroups    map[string]float64 `json:"age_groups"`
	Genders      map[string]float64 `json:"genders"`
	Incomes      map[string]float64 `json:"incomes"`
	Locations    map[string]float64 `json:"locations"`
	Lifestyles   map[string]float64 `json:"lifestyles"`
	Occupations  map[string]float64 `json:"occupations"`
}

// SeasonalTrends contains seasonal trend analysis
type SeasonalTrends struct {
	BeverageType    string                    `json:"beverage_type"`
	Years           int                       `json:"years"`
	MonthlyPatterns map[string]float64        `json:"monthly_patterns"`  // month -> sales factor
	SeasonalPeaks   []SeasonalPeak            `json:"seasonal_peaks"`
	YearOverYear    map[string]float64        `json:"year_over_year"`    // year -> growth
	Predictions     map[string]float64        `json:"predictions"`       // future months
	Confidence      float64                   `json:"confidence"`
	Factors         []string                  `json:"factors"`
}

// SeasonalPeak represents a seasonal peak period
type SeasonalPeak struct {
	Period      string  `json:"period"`      // e.g., "December", "Summer"
	StartMonth  int     `json:"start_month"`
	EndMonth    int     `json:"end_month"`
	PeakFactor  float64 `json:"peak_factor"` // multiplier vs baseline
	Reliability float64 `json:"reliability"` // how consistent this peak is
}

// SuccessPrediction contains AI prediction of beverage success
type SuccessPrediction struct {
	BeverageID       string             `json:"beverage_id"`
	SuccessScore     float64            `json:"success_score"`     // 0-100
	MarketFit        float64            `json:"market_fit"`        // 0-100
	RevenueProjection float64           `json:"revenue_projection"`
	VolumeProjection float64            `json:"volume_projection"`
	RiskFactors      []string           `json:"risk_factors"`
	SuccessFactors   []string           `json:"success_factors"`
	Recommendations  []string           `json:"recommendations"`
	Confidence       float64            `json:"confidence"`
	Timeframe        time.Duration      `json:"timeframe"`
	Scenarios        map[string]float64 `json:"scenarios"`         // scenario -> probability
}

// PreferenceAnalysis contains analysis of customer preferences
type PreferenceAnalysis struct {
	Timeframe       time.Duration          `json:"timeframe"`
	KeyInsights     []string               `json:"key_insights"`
	TrendingFlavors []string               `json:"trending_flavors"`
	EmergingSegments []string              `json:"emerging_segments"`
	ShiftingPreferences map[string]string  `json:"shifting_preferences"`
	Opportunities   []string               `json:"opportunities"`
	Threats         []string               `json:"threats"`
	Recommendations []string               `json:"recommendations"`
	Confidence      float64                `json:"confidence"`
}

// AnalyticsInsights contains AI-generated insights from metrics
type AnalyticsInsights struct {
	Timeframe       time.Duration          `json:"timeframe"`
	KeyFindings     []string               `json:"key_findings"`
	PerformanceGaps []string               `json:"performance_gaps"`
	Opportunities   []string               `json:"opportunities"`
	Risks           []string               `json:"risks"`
	Recommendations []string               `json:"recommendations"`
	Predictions     []string               `json:"predictions"`
	ActionItems     []ActionItem           `json:"action_items"`
	Confidence      float64                `json:"confidence"`
	GeneratedAt     time.Time              `json:"generated_at"`
}

// ActionItem represents a recommended action
type ActionItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    string    `json:"priority"`    // high, medium, low
	Category    string    `json:"category"`    // marketing, product, operations
	Impact      string    `json:"impact"`      // high, medium, low
	Effort      string    `json:"effort"`      // high, medium, low
	Timeline    string    `json:"timeline"`    // immediate, short-term, long-term
	Owner       string    `json:"owner"`
	DueDate     time.Time `json:"due_date"`
}

// DemandForecast contains demand forecasting data
type DemandForecast struct {
	BeverageID      string                 `json:"beverage_id"`
	Timeframe       time.Duration          `json:"timeframe"`
	Forecast        []ForecastPoint        `json:"forecast"`
	Seasonality     map[string]float64     `json:"seasonality"`
	TrendFactors    []string               `json:"trend_factors"`
	ExternalFactors []string               `json:"external_factors"`
	Scenarios       map[string][]ForecastPoint `json:"scenarios"`
	Confidence      float64                `json:"confidence"`
	Accuracy        float64                `json:"accuracy"`        // historical accuracy
	GeneratedAt     time.Time              `json:"generated_at"`
}

// ForecastPoint represents a single point in the demand forecast
type ForecastPoint struct {
	Date       time.Time `json:"date"`
	Demand     float64   `json:"demand"`
	LowerBound float64   `json:"lower_bound"`
	UpperBound float64   `json:"upper_bound"`
	Confidence float64   `json:"confidence"`
}

// AnalyticsReport contains a comprehensive analytics report
type AnalyticsReport struct {
	ReportID        string                 `json:"report_id"`
	Title           string                 `json:"title"`
	Timeframe       time.Duration          `json:"timeframe"`
	GeneratedAt     time.Time              `json:"generated_at"`
	Summary         *ReportSummary         `json:"summary"`
	PopularityTrends *PopularityTrends     `json:"popularity_trends"`
	Performance     []*PerformanceMetrics  `json:"performance"`
	CustomerInsights *CustomerSegmentData  `json:"customer_insights"`
	Predictions     *SuccessPrediction     `json:"predictions"`
	Insights        *AnalyticsInsights     `json:"insights"`
	Recommendations []string               `json:"recommendations"`
	ActionItems     []ActionItem           `json:"action_items"`
	Appendices      map[string]interface{} `json:"appendices"`
}

// ReportSummary contains a summary of the analytics report
type ReportSummary struct {
	TotalBeverages    int     `json:"total_beverages"`
	TotalRevenue      float64 `json:"total_revenue"`
	TotalVolume       float64 `json:"total_volume"`
	AverageRating     float64 `json:"average_rating"`
	TopPerformer      string  `json:"top_performer"`
	FastestGrowing    string  `json:"fastest_growing"`
	MostProfitable    string  `json:"most_profitable"`
	KeyTrends         []string `json:"key_trends"`
	CriticalIssues    []string `json:"critical_issues"`
	OverallHealth     string  `json:"overall_health"`   // excellent, good, fair, poor
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(dataStore AnalyticsDataStore, aiProvider AnalyticsAIProvider, logger Logger) *AnalyticsService {
	return &AnalyticsService{
		dataStore:  dataStore,
		aiProvider: aiProvider,
		logger:     logger,
	}
}

// RecordBeverageMetrics records metrics for a beverage
func (as *AnalyticsService) RecordBeverageMetrics(ctx context.Context, metrics *BeverageMetrics) error {
	as.logger.Info("Recording beverage metrics", "beverage_id", metrics.BeverageID)
	
	return as.dataStore.StoreBeverageMetrics(ctx, metrics)
}

// GeneratePopularityReport generates a popularity trends report
func (as *AnalyticsService) GeneratePopularityReport(ctx context.Context, timeframe time.Duration) (*PopularityTrends, error) {
	as.logger.Info("Generating popularity report", "timeframe", timeframe)
	
	trends, err := as.dataStore.GetPopularityTrends(ctx, timeframe)
	if err != nil {
		as.logger.Error("Failed to get popularity trends", err)
		return nil, err
	}
	
	// Enhance with AI insights if available
	if as.aiProvider != nil {
		// Could add AI-enhanced trend analysis here
	}
	
	return trends, nil
}

// GeneratePerformanceReport generates a performance report for specific beverages
func (as *AnalyticsService) GeneratePerformanceReport(ctx context.Context, beverageIDs []string, timeframe time.Duration) ([]*PerformanceMetrics, error) {
	as.logger.Info("Generating performance report", 
		"beverage_count", len(beverageIDs), 
		"timeframe", timeframe)
	
	metrics, err := as.dataStore.GetPerformanceMetrics(ctx, beverageIDs, timeframe)
	if err != nil {
		as.logger.Error("Failed to get performance metrics", err)
		return nil, err
	}
	
	// Calculate benchmarks and additional insights
	as.enhancePerformanceMetrics(metrics)
	
	return metrics, nil
}

// PredictBeverageSuccess predicts the success of a new beverage
func (as *AnalyticsService) PredictBeverageSuccess(ctx context.Context, beverage *entities.Beverage, marketContext *MarketContext) (*SuccessPrediction, error) {
	as.logger.Info("Predicting beverage success", "beverage_id", beverage.ID)
	
	if as.aiProvider == nil {
		return nil, fmt.Errorf("AI provider not available for success prediction")
	}
	
	prediction, err := as.aiProvider.PredictBeverageSuccess(ctx, beverage, marketContext)
	if err != nil {
		as.logger.Error("Failed to predict beverage success", err, "beverage_id", beverage.ID)
		return nil, err
	}
	
	as.logger.Info("Success prediction completed", 
		"beverage_id", beverage.ID,
		"success_score", prediction.SuccessScore,
		"confidence", prediction.Confidence)
	
	return prediction, nil
}

// GenerateComprehensiveReport generates a comprehensive analytics report
func (as *AnalyticsService) GenerateComprehensiveReport(ctx context.Context, timeframe time.Duration) (*AnalyticsReport, error) {
	as.logger.Info("Generating comprehensive analytics report", "timeframe", timeframe)
	
	report := &AnalyticsReport{
		ReportID:    fmt.Sprintf("report_%d", time.Now().Unix()),
		Title:       fmt.Sprintf("Beverage Analytics Report - %s", timeframe.String()),
		Timeframe:   timeframe,
		GeneratedAt: time.Now(),
		Appendices:  make(map[string]interface{}),
	}
	
	// Get popularity trends
	trends, err := as.dataStore.GetPopularityTrends(ctx, timeframe)
	if err == nil {
		report.PopularityTrends = trends
	}
	
	// Get customer insights
	customerData, err := as.dataStore.GetCustomerSegmentData(ctx, timeframe)
	if err == nil {
		report.CustomerInsights = customerData
	}
	
	// Generate AI insights if available
	if as.aiProvider != nil && len(report.PopularityTrends.TopBeverages) > 0 {
		// Get metrics for top beverages
		beverageIDs := make([]string, len(report.PopularityTrends.TopBeverages))
		for i, ranking := range report.PopularityTrends.TopBeverages {
			beverageIDs[i] = ranking.BeverageID
		}
		
		metrics, err := as.dataStore.GetPerformanceMetrics(ctx, beverageIDs, timeframe)
		if err == nil {
			report.Performance = metrics
			
			// Convert to BeverageMetrics for AI analysis
			beverageMetrics := as.convertToAnalyticsMetrics(metrics)
			insights, err := as.aiProvider.GenerateInsights(ctx, beverageMetrics)
			if err == nil {
				report.Insights = insights
				report.Recommendations = insights.Recommendations
				report.ActionItems = insights.ActionItems
			}
		}
	}
	
	// Generate summary
	report.Summary = as.generateReportSummary(report)
	
	as.logger.Info("Comprehensive report generated", "report_id", report.ReportID)
	
	return report, nil
}

// enhancePerformanceMetrics adds benchmarks and additional calculations
func (as *AnalyticsService) enhancePerformanceMetrics(metrics []*PerformanceMetrics) {
	if len(metrics) == 0 {
		return
	}
	
	// Calculate industry benchmarks
	totalRevenue := 0.0
	totalVolume := 0.0
	totalRating := 0.0
	
	for _, metric := range metrics {
		totalRevenue += metric.TotalRevenue
		totalVolume += metric.TotalVolume
		totalRating += metric.AverageRating
	}
	
	avgRevenue := totalRevenue / float64(len(metrics))
	avgVolume := totalVolume / float64(len(metrics))
	avgRating := totalRating / float64(len(metrics))
	
	// Add benchmarks to each metric
	for _, metric := range metrics {
		if metric.Benchmarks == nil {
			metric.Benchmarks = make(map[string]float64)
		}
		
		metric.Benchmarks["avg_revenue"] = avgRevenue
		metric.Benchmarks["avg_volume"] = avgVolume
		metric.Benchmarks["avg_rating"] = avgRating
		metric.Benchmarks["revenue_vs_avg"] = (metric.TotalRevenue / avgRevenue) * 100
		metric.Benchmarks["volume_vs_avg"] = (metric.TotalVolume / avgVolume) * 100
		metric.Benchmarks["rating_vs_avg"] = (metric.AverageRating / avgRating) * 100
	}
}

// convertToAnalyticsMetrics converts performance metrics to analytics metrics
func (as *AnalyticsService) convertToAnalyticsMetrics(performance []*PerformanceMetrics) []*BeverageMetrics {
	metrics := make([]*BeverageMetrics, len(performance))
	
	for i, perf := range performance {
		metrics[i] = &BeverageMetrics{
			BeverageID:     perf.BeverageID,
			Timestamp:      time.Now(),
			SalesVolume:    perf.TotalVolume,
			Revenue:        perf.TotalRevenue,
			CustomerRating: perf.AverageRating,
			OrderCount:     perf.CustomerCount,
			ReturnRate:     perf.ChurnRate,
			ProfitMargin:   perf.ROI,
		}
	}
	
	return metrics
}

// generateReportSummary generates a summary for the analytics report
func (as *AnalyticsService) generateReportSummary(report *AnalyticsReport) *ReportSummary {
	summary := &ReportSummary{
		KeyTrends:      []string{},
		CriticalIssues: []string{},
	}
	
	if report.Performance != nil && len(report.Performance) > 0 {
		summary.TotalBeverages = len(report.Performance)
		
		// Calculate totals
		for _, perf := range report.Performance {
			summary.TotalRevenue += perf.TotalRevenue
			summary.TotalVolume += perf.TotalVolume
			summary.AverageRating += perf.AverageRating
		}
		
		summary.AverageRating /= float64(len(report.Performance))
		
		// Find top performers
		sort.Slice(report.Performance, func(i, j int) bool {
			return report.Performance[i].TotalRevenue > report.Performance[j].TotalRevenue
		})
		
		if len(report.Performance) > 0 {
			summary.TopPerformer = report.Performance[0].Name
		}
		
		// Determine overall health
		if summary.AverageRating >= 4.5 {
			summary.OverallHealth = "excellent"
		} else if summary.AverageRating >= 4.0 {
			summary.OverallHealth = "good"
		} else if summary.AverageRating >= 3.5 {
			summary.OverallHealth = "fair"
		} else {
			summary.OverallHealth = "poor"
		}
	}
	
	// Add insights from AI analysis
	if report.Insights != nil {
		summary.KeyTrends = append(summary.KeyTrends, report.Insights.KeyFindings...)
		summary.CriticalIssues = append(summary.CriticalIssues, report.Insights.Risks...)
	}
	
	return summary
}
