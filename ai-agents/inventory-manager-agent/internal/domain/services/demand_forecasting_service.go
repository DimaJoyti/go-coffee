package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/inventory-manager-agent/internal/domain/entities"
	"go-coffee-ai-agents/inventory-manager-agent/internal/domain/repositories"
)

// DemandForecastingService provides AI-powered demand forecasting capabilities
type DemandForecastingService struct {
	inventoryRepo repositories.InventoryRepository
	movementRepo  repositories.StockMovementRepository
	mlService     MachineLearningService
	logger        Logger
}

// MachineLearningService defines the interface for ML operations
type MachineLearningService interface {
	TrainForecastModel(ctx context.Context, data *TimeSeriesData) (*ForecastModel, error)
	PredictDemand(ctx context.Context, model *ForecastModel, horizon int) (*DemandPrediction, error)
	DetectSeasonality(ctx context.Context, data *TimeSeriesData) (*SeasonalityAnalysis, error)
	AnalyzeTrends(ctx context.Context, data *TimeSeriesData) (*TrendAnalysis, error)
	CalculateConfidenceInterval(ctx context.Context, prediction *DemandPrediction, confidence float64) (*ConfidenceInterval, error)
}

// TimeSeriesData represents time series data for forecasting
type TimeSeriesData struct {
	ItemID      uuid.UUID      `json:"item_id"`
	DataPoints  []*DataPoint   `json:"data_points"`
	Frequency   string         `json:"frequency"`   // daily, weekly, monthly
	StartDate   time.Time      `json:"start_date"`
	EndDate     time.Time      `json:"end_date"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// DataPoint represents a single data point in time series
type DataPoint struct {
	Date     time.Time `json:"date"`
	Value    float64   `json:"value"`
	Features map[string]float64 `json:"features"` // external factors
}

// ForecastModel represents a trained forecasting model
type ForecastModel struct {
	ID           uuid.UUID              `json:"id"`
	ItemID       uuid.UUID              `json:"item_id"`
	Algorithm    string                 `json:"algorithm"`
	Parameters   map[string]interface{} `json:"parameters"`
	Accuracy     float64                `json:"accuracy"`
	TrainedAt    time.Time              `json:"trained_at"`
	ValidUntil   time.Time              `json:"valid_until"`
	ModelData    []byte                 `json:"model_data"`
}

// DemandPrediction represents a demand prediction
type DemandPrediction struct {
	ItemID          uuid.UUID        `json:"item_id"`
	PredictionDate  time.Time        `json:"prediction_date"`
	Horizon         int              `json:"horizon"`        // days
	PredictedValues []*PredictedValue `json:"predicted_values"`
	Confidence      float64          `json:"confidence"`
	Algorithm       string           `json:"algorithm"`
	ModelVersion    string           `json:"model_version"`
}

// PredictedValue represents a single predicted value
type PredictedValue struct {
	Date       time.Time `json:"date"`
	Value      float64   `json:"value"`
	LowerBound float64   `json:"lower_bound"`
	UpperBound float64   `json:"upper_bound"`
}

// SeasonalityAnalysis represents seasonality analysis results
type SeasonalityAnalysis struct {
	HasSeasonality    bool               `json:"has_seasonality"`
	SeasonalPeriod    int                `json:"seasonal_period"`
	SeasonalStrength  float64            `json:"seasonal_strength"`
	SeasonalFactors   map[string]float64 `json:"seasonal_factors"`
	DayOfWeekFactors  map[string]float64 `json:"day_of_week_factors"`
	MonthlyFactors    map[string]float64 `json:"monthly_factors"`
	HolidayEffects    map[string]float64 `json:"holiday_effects"`
}

// TrendAnalysis represents trend analysis results
type TrendAnalysis struct {
	HasTrend      bool    `json:"has_trend"`
	TrendSlope    float64 `json:"trend_slope"`
	TrendStrength float64 `json:"trend_strength"`
	Direction     string  `json:"direction"` // increasing, decreasing, stable
	ChangePoints  []time.Time `json:"change_points"`
}

// ConfidenceInterval represents a confidence interval
type ConfidenceInterval struct {
	Lower      float64 `json:"lower"`
	Upper      float64 `json:"upper"`
	Confidence float64 `json:"confidence"`
}

// DemandForecast represents the final demand forecast
type DemandForecast struct {
	ItemID             uuid.UUID              `json:"item_id"`
	ForecastDate       time.Time              `json:"forecast_date"`
	Period             time.Duration          `json:"period"`
	PredictedDemand    float64                `json:"predicted_demand"`
	ConfidenceInterval *ConfidenceInterval    `json:"confidence_interval"`
	SeasonalFactors    map[string]float64     `json:"seasonal_factors"`
	TrendFactors       map[string]float64     `json:"trend_factors"`
	ForecastPoints     []*ForecastPoint       `json:"forecast_points"`
	Accuracy           float64                `json:"accuracy"`
	Algorithm          string                 `json:"algorithm"`
	ModelVersion       string                 `json:"model_version"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// ForecastPoint represents a single forecast point
type ForecastPoint struct {
	Date       time.Time `json:"date"`
	Value      float64   `json:"value"`
	LowerBound float64   `json:"lower_bound"`
	UpperBound float64   `json:"upper_bound"`
	Confidence float64   `json:"confidence"`
}

// ConsumptionPattern represents consumption pattern analysis
type ConsumptionPattern struct {
	ItemID             uuid.UUID          `json:"item_id"`
	AnalysisPeriod     time.Duration      `json:"analysis_period"`
	AverageDaily       float64            `json:"average_daily"`
	AverageWeekly      float64            `json:"average_weekly"`
	AverageMonthly     float64            `json:"average_monthly"`
	Volatility         float64            `json:"volatility"`
	Seasonality        map[string]float64 `json:"seasonality"`
	DayOfWeekPattern   map[string]float64 `json:"day_of_week_pattern"`
	MonthlyPattern     map[string]float64 `json:"monthly_pattern"`
	TrendDirection     string             `json:"trend_direction"`
	TrendStrength      float64            `json:"trend_strength"`
	PeakDays           []string           `json:"peak_days"`
	LowDays            []string           `json:"low_days"`
	Outliers           []*OutlierPoint    `json:"outliers"`
}

// OutlierPoint represents an outlier in consumption data
type OutlierPoint struct {
	Date     time.Time `json:"date"`
	Value    float64   `json:"value"`
	Expected float64   `json:"expected"`
	Deviation float64  `json:"deviation"`
	Reason   string    `json:"reason"`
}

// StockoutPrediction represents stockout prediction
type StockoutPrediction struct {
	ItemID             uuid.UUID  `json:"item_id"`
	CurrentStock       float64    `json:"current_stock"`
	ProbabilityPercent float64    `json:"probability_percent"`
	PredictedDate      *time.Time `json:"predicted_date,omitempty"`
	DaysUntilStockout  *int       `json:"days_until_stockout,omitempty"`
	RecommendedAction  string     `json:"recommended_action"`
	UrgencyLevel       string     `json:"urgency_level"`
	Confidence         float64    `json:"confidence"`
	Factors            []string   `json:"factors"`
}

// NewDemandForecastingService creates a new demand forecasting service
func NewDemandForecastingService(
	inventoryRepo repositories.InventoryRepository,
	movementRepo repositories.StockMovementRepository,
	mlService MachineLearningService,
	logger Logger,
) *DemandForecastingService {
	return &DemandForecastingService{
		inventoryRepo: inventoryRepo,
		movementRepo:  movementRepo,
		mlService:     mlService,
		logger:        logger,
	}
}

// ForecastDemand generates demand forecast for an inventory item
func (dfs *DemandForecastingService) ForecastDemand(ctx context.Context, itemID uuid.UUID, period time.Duration) (*DemandForecast, error) {
	dfs.logger.Info("Generating demand forecast", "item_id", itemID, "period", period)

	// Get historical data
	timeSeriesData, err := dfs.getTimeSeriesData(ctx, itemID, period)
	if err != nil {
		dfs.logger.Error("Failed to get time series data", err, "item_id", itemID)
		return nil, err
	}

	// Train or get existing model
	model, err := dfs.getOrTrainModel(ctx, itemID, timeSeriesData)
	if err != nil {
		dfs.logger.Error("Failed to get forecast model", err, "item_id", itemID)
		return nil, err
	}

	// Generate prediction
	horizon := int(period.Hours() / 24) // Convert to days
	prediction, err := dfs.mlService.PredictDemand(ctx, model, horizon)
	if err != nil {
		dfs.logger.Error("Failed to predict demand", err, "item_id", itemID)
		return nil, err
	}

	// Analyze seasonality and trends
	seasonality, err := dfs.mlService.DetectSeasonality(ctx, timeSeriesData)
	if err != nil {
		dfs.logger.Warn("Failed to detect seasonality", "item_id", itemID)
		seasonality = &SeasonalityAnalysis{}
	}

	trends, err := dfs.mlService.AnalyzeTrends(ctx, timeSeriesData)
	if err != nil {
		dfs.logger.Warn("Failed to analyze trends", "item_id", itemID)
		trends = &TrendAnalysis{}
	}

	// Calculate confidence interval
	confidenceInterval, err := dfs.mlService.CalculateConfidenceInterval(ctx, prediction, 0.95)
	if err != nil {
		dfs.logger.Warn("Failed to calculate confidence interval", "item_id", itemID)
		confidenceInterval = &ConfidenceInterval{Confidence: 0.95}
	}

	// Build forecast response
	forecast := &DemandForecast{
		ItemID:             itemID,
		ForecastDate:       time.Now(),
		Period:             period,
		PredictedDemand:    dfs.calculateTotalDemand(prediction.PredictedValues),
		ConfidenceInterval: confidenceInterval,
		SeasonalFactors:    seasonality.SeasonalFactors,
		TrendFactors:       map[string]float64{"slope": trends.TrendSlope, "strength": trends.TrendStrength},
		ForecastPoints:     dfs.convertToForecastPoints(prediction.PredictedValues),
		Accuracy:           model.Accuracy,
		Algorithm:          model.Algorithm,
		ModelVersion:       fmt.Sprintf("v%d", model.TrainedAt.Unix()),
		Metadata: map[string]interface{}{
			"has_seasonality": seasonality.HasSeasonality,
			"has_trend":       trends.HasTrend,
			"data_points":     len(timeSeriesData.DataPoints),
		},
	}

	dfs.logger.Info("Demand forecast generated successfully", 
		"item_id", itemID,
		"predicted_demand", forecast.PredictedDemand,
		"accuracy", forecast.Accuracy)

	return forecast, nil
}

// GetConsumptionPattern analyzes consumption patterns for an item
func (dfs *DemandForecastingService) GetConsumptionPattern(ctx context.Context, itemID uuid.UUID, period time.Duration) (*ConsumptionPattern, error) {
	dfs.logger.Info("Analyzing consumption pattern", "item_id", itemID, "period", period)

	// Get historical movement data
	endDate := time.Now()
	startDate := endDate.Add(-period)

	filter := &repositories.MovementFilter{
		ItemIDs:      []uuid.UUID{itemID},
		Types:        []entities.MovementType{entities.MovementTypeIssue, entities.MovementTypeConsumption},
		CreatedAfter: &startDate,
		CreatedBefore: &endDate,
		SortBy:       "created_at",
		SortOrder:    "asc",
		Limit:        10000,
	}

	movements, err := dfs.movementRepo.List(ctx, filter)
	if err != nil {
		dfs.logger.Error("Failed to get movement data", err, "item_id", itemID)
		return nil, err
	}

	// Analyze consumption patterns
	pattern := dfs.analyzeConsumptionPattern(movements, period)
	pattern.ItemID = itemID
	pattern.AnalysisPeriod = period

	dfs.logger.Info("Consumption pattern analyzed", 
		"item_id", itemID,
		"average_daily", pattern.AverageDaily,
		"volatility", pattern.Volatility)

	return pattern, nil
}

// PredictStockout predicts when an item will stock out
func (dfs *DemandForecastingService) PredictStockout(ctx context.Context, itemID uuid.UUID) (*StockoutPrediction, error) {
	dfs.logger.Info("Predicting stockout", "item_id", itemID)

	// Get current inventory item
	item, err := dfs.inventoryRepo.GetByID(ctx, itemID)
	if err != nil {
		dfs.logger.Error("Failed to get inventory item", err, "item_id", itemID)
		return nil, err
	}

	// Get consumption pattern for the last 30 days
	pattern, err := dfs.GetConsumptionPattern(ctx, itemID, 30*24*time.Hour)
	if err != nil {
		dfs.logger.Error("Failed to get consumption pattern", err, "item_id", itemID)
		return nil, err
	}

	// Calculate stockout prediction
	prediction := dfs.calculateStockoutPrediction(item, pattern)

	dfs.logger.Info("Stockout prediction completed", 
		"item_id", itemID,
		"probability", prediction.ProbabilityPercent,
		"days_until_stockout", prediction.DaysUntilStockout)

	return prediction, nil
}

// Helper methods

func (dfs *DemandForecastingService) getTimeSeriesData(ctx context.Context, itemID uuid.UUID, period time.Duration) (*TimeSeriesData, error) {
	endDate := time.Now()
	startDate := endDate.Add(-period)

	// Get outbound movements (consumption)
	filter := &repositories.MovementFilter{
		ItemIDs:       []uuid.UUID{itemID},
		Directions:    []entities.MovementDirection{entities.DirectionOut},
		CreatedAfter:  &startDate,
		CreatedBefore: &endDate,
		SortBy:        "created_at",
		SortOrder:     "asc",
		Limit:         10000,
	}

	movements, err := dfs.movementRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Aggregate data by day
	dailyData := make(map[string]float64)
	for _, movement := range movements {
		dateKey := movement.CreatedAt.Format("2006-01-02")
		dailyData[dateKey] += movement.Quantity
	}

	// Convert to data points
	var dataPoints []*DataPoint
	current := startDate
	for current.Before(endDate) {
		dateKey := current.Format("2006-01-02")
		value := dailyData[dateKey]
		
		dataPoints = append(dataPoints, &DataPoint{
			Date:  current,
			Value: value,
			Features: map[string]float64{
				"day_of_week": float64(current.Weekday()),
				"day_of_month": float64(current.Day()),
				"month": float64(current.Month()),
			},
		})
		
		current = current.AddDate(0, 0, 1)
	}

	return &TimeSeriesData{
		ItemID:     itemID,
		DataPoints: dataPoints,
		Frequency:  "daily",
		StartDate:  startDate,
		EndDate:    endDate,
		Metadata: map[string]interface{}{
			"total_movements": len(movements),
			"data_points":     len(dataPoints),
		},
	}, nil
}

func (dfs *DemandForecastingService) getOrTrainModel(ctx context.Context, itemID uuid.UUID, data *TimeSeriesData) (*ForecastModel, error) {
	// For now, create a simple model
	// In production, this would check for existing models and retrain if necessary
	model, err := dfs.mlService.TrainForecastModel(ctx, data)
	if err != nil {
		return nil, err
	}

	return model, nil
}

func (dfs *DemandForecastingService) calculateTotalDemand(values []*PredictedValue) float64 {
	total := 0.0
	for _, value := range values {
		total += value.Value
	}
	return total
}

func (dfs *DemandForecastingService) convertToForecastPoints(values []*PredictedValue) []*ForecastPoint {
	points := make([]*ForecastPoint, len(values))
	for i, value := range values {
		points[i] = &ForecastPoint{
			Date:       value.Date,
			Value:      value.Value,
			LowerBound: value.LowerBound,
			UpperBound: value.UpperBound,
			Confidence: 0.95, // Default confidence
		}
	}
	return points
}

func (dfs *DemandForecastingService) analyzeConsumptionPattern(movements []*entities.StockMovement, period time.Duration) *ConsumptionPattern {
	if len(movements) == 0 {
		return &ConsumptionPattern{}
	}

	// Calculate basic statistics
	totalConsumption := 0.0
	dailyConsumption := make(map[string]float64)
	weekdayConsumption := make(map[time.Weekday][]float64)
	monthlyConsumption := make(map[time.Month][]float64)

	for _, movement := range movements {
		quantity := movement.Quantity
		totalConsumption += quantity
		
		dateKey := movement.CreatedAt.Format("2006-01-02")
		dailyConsumption[dateKey] += quantity
		
		weekday := movement.CreatedAt.Weekday()
		weekdayConsumption[weekday] = append(weekdayConsumption[weekday], quantity)
		
		month := movement.CreatedAt.Month()
		monthlyConsumption[month] = append(monthlyConsumption[month], quantity)
	}

	days := float64(period.Hours() / 24)
	averageDaily := totalConsumption / days
	averageWeekly := averageDaily * 7
	averageMonthly := averageDaily * 30

	// Calculate volatility (standard deviation)
	volatility := dfs.calculateVolatility(dailyConsumption, averageDaily)

	// Build day of week pattern
	dayOfWeekPattern := make(map[string]float64)
	for weekday, values := range weekdayConsumption {
		if len(values) > 0 {
			avg := dfs.calculateAverage(values)
			dayOfWeekPattern[weekday.String()] = avg
		}
	}

	// Build monthly pattern
	monthlyPattern := make(map[string]float64)
	for month, values := range monthlyConsumption {
		if len(values) > 0 {
			avg := dfs.calculateAverage(values)
			monthlyPattern[month.String()] = avg
		}
	}

	// Determine trend direction
	trendDirection := dfs.determineTrendDirection(movements)

	return &ConsumptionPattern{
		AverageDaily:     averageDaily,
		AverageWeekly:    averageWeekly,
		AverageMonthly:   averageMonthly,
		Volatility:       volatility,
		DayOfWeekPattern: dayOfWeekPattern,
		MonthlyPattern:   monthlyPattern,
		TrendDirection:   trendDirection,
		TrendStrength:    0.5, // Simplified calculation
	}
}

func (dfs *DemandForecastingService) calculateStockoutPrediction(item *entities.InventoryItem, pattern *ConsumptionPattern) *StockoutPrediction {
	currentStock := item.AvailableStock
	averageDaily := pattern.AverageDaily

	prediction := &StockoutPrediction{
		ItemID:       item.ID,
		CurrentStock: currentStock,
		Confidence:   0.8, // Default confidence
	}

	if averageDaily <= 0 {
		prediction.ProbabilityPercent = 0
		prediction.RecommendedAction = "Monitor consumption patterns"
		prediction.UrgencyLevel = "low"
		return prediction
	}

	// Calculate days until stockout
	daysUntilStockout := int(currentStock / averageDaily)
	prediction.DaysUntilStockout = &daysUntilStockout

	if daysUntilStockout > 0 {
		predictedDate := time.Now().AddDate(0, 0, daysUntilStockout)
		prediction.PredictedDate = &predictedDate
	}

	// Calculate probability based on volatility and current stock level
	volatilityFactor := pattern.Volatility / averageDaily
	stockRatio := currentStock / (averageDaily * 30) // 30-day supply ratio

	if stockRatio < 0.1 {
		prediction.ProbabilityPercent = 95
		prediction.UrgencyLevel = "critical"
		prediction.RecommendedAction = "Immediate reorder required"
	} else if stockRatio < 0.3 {
		prediction.ProbabilityPercent = 75
		prediction.UrgencyLevel = "high"
		prediction.RecommendedAction = "Reorder soon"
	} else if stockRatio < 0.5 {
		prediction.ProbabilityPercent = 50
		prediction.UrgencyLevel = "medium"
		prediction.RecommendedAction = "Plan reorder"
	} else {
		prediction.ProbabilityPercent = 25
		prediction.UrgencyLevel = "low"
		prediction.RecommendedAction = "Monitor stock levels"
	}

	// Adjust for volatility
	prediction.ProbabilityPercent += volatilityFactor * 20
	if prediction.ProbabilityPercent > 100 {
		prediction.ProbabilityPercent = 100
	}

	prediction.Factors = []string{
		fmt.Sprintf("Average daily consumption: %.2f", averageDaily),
		fmt.Sprintf("Current stock: %.2f", currentStock),
		fmt.Sprintf("Volatility factor: %.2f", volatilityFactor),
	}

	return prediction
}

func (dfs *DemandForecastingService) calculateVolatility(dailyData map[string]float64, average float64) float64 {
	if len(dailyData) <= 1 {
		return 0
	}

	sumSquaredDiff := 0.0
	for _, value := range dailyData {
		diff := value - average
		sumSquaredDiff += diff * diff
	}

	variance := sumSquaredDiff / float64(len(dailyData)-1)
	return math.Sqrt(variance)
}

func (dfs *DemandForecastingService) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func (dfs *DemandForecastingService) determineTrendDirection(movements []*entities.StockMovement) string {
	if len(movements) < 2 {
		return "stable"
	}

	// Sort movements by date
	sort.Slice(movements, func(i, j int) bool {
		return movements[i].CreatedAt.Before(movements[j].CreatedAt)
	})

	// Compare first half with second half
	midpoint := len(movements) / 2
	firstHalfSum := 0.0
	secondHalfSum := 0.0

	for i := 0; i < midpoint; i++ {
		firstHalfSum += movements[i].Quantity
	}

	for i := midpoint; i < len(movements); i++ {
		secondHalfSum += movements[i].Quantity
	}

	firstHalfAvg := firstHalfSum / float64(midpoint)
	secondHalfAvg := secondHalfSum / float64(len(movements)-midpoint)

	if secondHalfAvg > firstHalfAvg*1.1 {
		return "increasing"
	} else if secondHalfAvg < firstHalfAvg*0.9 {
		return "decreasing"
	}

	return "stable"
}

// Quality Management Types
// These types are placeholders for the quality management service functionality

// ExpiringItem represents an item that is expiring soon
type ExpiringItem struct {
	ItemID         uuid.UUID  `json:"item_id"`
	SKU            string     `json:"sku"`
	Name           string     `json:"name"`
	CurrentStock   float64    `json:"current_stock"`
	ExpirationDate time.Time  `json:"expiration_date"`
	DaysUntilExpiry int       `json:"days_until_expiry"`
	BatchNumber    string     `json:"batch_number"`
	LocationID     uuid.UUID  `json:"location_id"`
	UrgencyLevel   string     `json:"urgency_level"`
}

// ExpirationProcessingResult represents the result of processing expired items
type ExpirationProcessingResult struct {
	ProcessedAt     time.Time       `json:"processed_at"`
	LocationID      *uuid.UUID      `json:"location_id,omitempty"`
	TotalItems      int             `json:"total_items"`
	DisposedItems   int             `json:"disposed_items"`
	TransferredItems int            `json:"transferred_items"`
	ProcessedItems  []*ProcessedItem `json:"processed_items"`
	TotalLoss       entities.Money  `json:"total_loss"`
	ProcessedBy     string          `json:"processed_by"`
}

// ProcessedItem represents an individual processed expired item
type ProcessedItem struct {
	ItemID      uuid.UUID      `json:"item_id"`
	SKU         string         `json:"sku"`
	Name        string         `json:"name"`
	Quantity    float64        `json:"quantity"`
	Action      string         `json:"action"` // "disposed", "transferred", "discounted"
	Reason      string         `json:"reason"`
	BatchNumber string         `json:"batch_number"`
	Loss        entities.Money `json:"loss"`
}

// QualityMetrics represents quality management metrics
type QualityMetrics struct {
	Period              time.Duration       `json:"period"`
	TotalQualityChecks  int                `json:"total_quality_checks"`
	PassedChecks        int                `json:"passed_checks"`
	FailedChecks        int                `json:"failed_checks"`
	QualityScore        float64            `json:"quality_score"`
	ExpirationRate      float64            `json:"expiration_rate"`
	WastageRate         float64            `json:"wastage_rate"`
	TotalWastageValue   entities.Money     `json:"total_wastage_value"`
	TopWastageItems     []*WastageItem     `json:"top_wastage_items"`
	QualityTrends       map[string]float64 `json:"quality_trends"`
	ImprovementAreas    []string           `json:"improvement_areas"`
	GeneratedAt         time.Time          `json:"generated_at"`
}

// WastageItem represents an item with high wastage
type WastageItem struct {
	ItemID       uuid.UUID      `json:"item_id"`
	SKU          string         `json:"sku"`
	Name         string         `json:"name"`
	WastageQty   float64        `json:"wastage_quantity"`
	WastageValue entities.Money `json:"wastage_value"`
	WastageRate  float64        `json:"wastage_rate"`
	Reason       string         `json:"reason"`
}
