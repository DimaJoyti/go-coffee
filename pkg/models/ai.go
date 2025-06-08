package models

import "time"

// AI Service Models

// UserPreferences represents user preferences for recommendations
type UserPreferences struct {
	FavoriteType    string            `json:"favorite_type"`
	Allergies       []string          `json:"allergies"`
	PriceRange      PriceRange        `json:"price_range"`
	Customizations  []string          `json:"customizations"`
	PreferredTimes  []string          `json:"preferred_times"`
	Metadata        map[string]string `json:"metadata"`
}

// PriceRange represents a price range preference
type PriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// RecommendationResponse represents AI-generated recommendations
type RecommendationResponse struct {
	UserID          string                 `json:"user_id"`
	Recommendations []CoffeeRecommendation `json:"recommendations"`
	GeneratedAt     time.Time              `json:"generated_at"`
	Confidence      float64                `json:"confidence"`
	Algorithm       string                 `json:"algorithm"`
}

// CoffeeRecommendation represents a single coffee recommendation
type CoffeeRecommendation struct {
	CoffeeType      string   `json:"coffee_type"`
	Confidence      float64  `json:"confidence"`
	Reason          string   `json:"reason"`
	Price           float64  `json:"price"`
	Customizations  []string `json:"customizations"`
	EstimatedTime   int      `json:"estimated_time_minutes"`
	NutritionalInfo string   `json:"nutritional_info,omitempty"`
}

// ArbitrageResponse represents arbitrage analysis results
type ArbitrageResponse struct {
	Opportunities []ArbitrageOpportunity `json:"opportunities"`
	AnalyzedAt    time.Time              `json:"analyzed_at"`
	Markets       []string               `json:"markets"`
	TotalProfit   float64                `json:"total_profit"`
	RiskLevel     string                 `json:"risk_level"`
}

// ArbitrageOpportunity represents a single arbitrage opportunity
type ArbitrageOpportunity struct {
	Asset        string  `json:"asset"`
	BuyMarket    string  `json:"buy_market"`
	SellMarket   string  `json:"sell_market"`
	BuyPrice     float64 `json:"buy_price"`
	SellPrice    float64 `json:"sell_price"`
	ProfitMargin float64 `json:"profit_margin"`
	Volume       float64 `json:"volume"`
	Confidence   float64 `json:"confidence"`
	TimeWindow   int     `json:"time_window_seconds"`
	RiskScore    float64 `json:"risk_score"`
}

// DemandForecast represents demand forecasting results
type DemandForecast struct {
	Timeframe   string              `json:"timeframe"`
	Predictions []DemandPrediction  `json:"predictions"`
	GeneratedAt time.Time           `json:"generated_at"`
	Algorithm   string              `json:"algorithm"`
	Accuracy    float64             `json:"accuracy"`
	Factors     []string            `json:"factors"`
}

// DemandPrediction represents a single demand prediction
type DemandPrediction struct {
	Timestamp  time.Time `json:"timestamp"`
	Demand     float64   `json:"demand"`
	Confidence float64   `json:"confidence"`
	Category   string    `json:"category,omitempty"`
	Factors    []string  `json:"factors,omitempty"`
}

// PricingOptimization represents pricing optimization results
type PricingOptimization struct {
	Suggestions []PriceSuggestion `json:"suggestions"`
	GeneratedAt time.Time         `json:"generated_at"`
	Strategy    string            `json:"strategy"`
	TotalImpact float64           `json:"total_impact"`
	Timeframe   string            `json:"timeframe"`
}

// PriceSuggestion represents a single pricing suggestion
type PriceSuggestion struct {
	Product         string  `json:"product"`
	CurrentPrice    float64 `json:"current_price"`
	OptimizedPrice  float64 `json:"optimized_price"`
	ExpectedRevenue float64 `json:"expected_revenue"`
	Confidence      float64 `json:"confidence"`
	Reason          string  `json:"reason"`
	Impact          string  `json:"impact"`
}

// BehaviorAnalysis represents customer behavior analysis
type BehaviorAnalysis struct {
	CustomerID      string                 `json:"customer_id"`
	Patterns        []string               `json:"patterns"`
	Preferences     map[string]interface{} `json:"preferences"`
	LoyaltyScore    float64                `json:"loyalty_score"`
	ChurnRisk       float64                `json:"churn_risk"`
	LifetimeValue   float64                `json:"lifetime_value"`
	AnalyzedAt      time.Time              `json:"analyzed_at"`
	Recommendations []string               `json:"recommendations"`
	Segments        []string               `json:"segments"`
}

// InventoryOptimization represents inventory optimization results
type InventoryOptimization struct {
	Suggestions []InventorySuggestion `json:"suggestions"`
	GeneratedAt time.Time             `json:"generated_at"`
	Algorithm   string                `json:"algorithm"`
	CostSavings float64               `json:"cost_savings"`
	Efficiency  float64               `json:"efficiency_improvement"`
}

// InventorySuggestion represents a single inventory suggestion
type InventorySuggestion struct {
	Item             string  `json:"item"`
	CurrentStock     int     `json:"current_stock"`
	RecommendedStock int     `json:"recommended_stock"`
	ReorderPoint     int     `json:"reorder_point"`
	Reason           string  `json:"reason"`
	Priority         int     `json:"priority"`
	CostImpact       float64 `json:"cost_impact"`
}

// AIAnalyticsRequest represents a request for AI analytics
type AIAnalyticsRequest struct {
	Type       string                 `json:"type"` // recommendation, arbitrage, demand, pricing, behavior, inventory
	Parameters map[string]interface{} `json:"parameters"`
	UserID     string                 `json:"user_id,omitempty"`
	Timeframe  string                 `json:"timeframe,omitempty"`
}

// AIAnalyticsResponse represents a generic AI analytics response
type AIAnalyticsResponse struct {
	Type        string      `json:"type"`
	Data        interface{} `json:"data"`
	GeneratedAt time.Time   `json:"generated_at"`
	Success     bool        `json:"success"`
	Message     string      `json:"message,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MLModel represents a machine learning model
type MLModel struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Version     string                 `json:"version"`
	Accuracy    float64                `json:"accuracy"`
	TrainedAt   time.Time              `json:"trained_at"`
	Parameters  map[string]interface{} `json:"parameters"`
	Status      string                 `json:"status"` // training, ready, deprecated
	Description string                 `json:"description"`
}

// TrainingData represents training data for ML models
type TrainingData struct {
	ID          string                 `json:"id"`
	ModelType   string                 `json:"model_type"`
	Features    map[string]interface{} `json:"features"`
	Labels      map[string]interface{} `json:"labels"`
	CreatedAt   time.Time              `json:"created_at"`
	Source      string                 `json:"source"`
	Quality     float64                `json:"quality_score"`
}

// PredictionRequest represents a request for ML prediction
type PredictionRequest struct {
	ModelID string                 `json:"model_id"`
	Input   map[string]interface{} `json:"input"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// PredictionResponse represents a ML prediction response
type PredictionResponse struct {
	ModelID     string                 `json:"model_id"`
	Prediction  interface{}            `json:"prediction"`
	Confidence  float64                `json:"confidence"`
	Explanation string                 `json:"explanation,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	PredictedAt time.Time              `json:"predicted_at"`
}

// AIAgent represents an AI agent configuration
type AIAgent struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"` // recommender, arbitrage, forecaster, etc.
	Status       string                 `json:"status"` // active, inactive, training
	Capabilities []string               `json:"capabilities"`
	Config       map[string]interface{} `json:"config"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Performance  AgentPerformance       `json:"performance"`
}

// AgentPerformance represents AI agent performance metrics
type AgentPerformance struct {
	Accuracy        float64   `json:"accuracy"`
	ResponseTime    float64   `json:"response_time_ms"`
	RequestsHandled int64     `json:"requests_handled"`
	ErrorRate       float64   `json:"error_rate"`
	LastEvaluated   time.Time `json:"last_evaluated"`
	Uptime          float64   `json:"uptime_percentage"`
}

// AIInsight represents an AI-generated business insight
type AIInsight struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // trend, anomaly, opportunity, risk
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Impact      string                 `json:"impact"` // high, medium, low
	Category    string                 `json:"category"`
	Data        map[string]interface{} `json:"data"`
	Actions     []string               `json:"recommended_actions"`
	GeneratedAt time.Time              `json:"generated_at"`
	ExpiresAt   time.Time              `json:"expires_at,omitempty"`
}

// AIConfiguration represents AI service configuration
type AIConfiguration struct {
	EnabledModules    []string               `json:"enabled_modules"`
	ModelSettings     map[string]interface{} `json:"model_settings"`
	PerformanceThresholds map[string]float64 `json:"performance_thresholds"`
	UpdateFrequency   string                 `json:"update_frequency"`
	DataRetention     string                 `json:"data_retention"`
	PrivacySettings   map[string]bool        `json:"privacy_settings"`
}

// AIMetrics represents AI service metrics
type AIMetrics struct {
	TotalRequests     int64     `json:"total_requests"`
	SuccessfulRequests int64    `json:"successful_requests"`
	AverageResponseTime float64 `json:"average_response_time_ms"`
	ModelAccuracy     map[string]float64 `json:"model_accuracy"`
	ResourceUsage     ResourceUsage      `json:"resource_usage"`
	LastUpdated       time.Time          `json:"last_updated"`
}

// ResourceUsage represents resource usage metrics
type ResourceUsage struct {
	CPUUsage    float64 `json:"cpu_usage_percent"`
	MemoryUsage float64 `json:"memory_usage_mb"`
	GPUUsage    float64 `json:"gpu_usage_percent,omitempty"`
	StorageUsage float64 `json:"storage_usage_mb"`
}
