package services

import (
	"context"
	"fmt"
	"time"

	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/entities"
)

// ExternalIntegrationService handles integrations with external systems
type ExternalIntegrationService struct {
	googleSheetsAPI GoogleSheetsAPI
	socialMediaAPI  SocialMediaAPI
	inventoryAPI    InventoryAPI
	qualityAPI      QualityAPI
	logger          Logger
}

// GoogleSheetsAPI defines the interface for Google Sheets integration
type GoogleSheetsAPI interface {
	CreateRecipeSheet(ctx context.Context, beverage *entities.Beverage, analysis *RecipeAnalysisData) (*SheetInfo, error)
	UpdateRecipeSheet(ctx context.Context, sheetID string, updates *RecipeUpdates) error
	GetRecipeData(ctx context.Context, sheetID string) (*RecipeData, error)
	ShareSheet(ctx context.Context, sheetID string, emails []string, permission string) error
}

// SocialMediaAPI defines the interface for social media integration
type SocialMediaAPI interface {
	PostRecipe(ctx context.Context, platform string, content *SocialMediaContent) (*PostResult, error)
	SchedulePost(ctx context.Context, platform string, content *SocialMediaContent, scheduledTime time.Time) (*ScheduledPost, error)
	GetEngagementMetrics(ctx context.Context, postID string) (*EngagementMetrics, error)
	GenerateHashtags(ctx context.Context, beverage *entities.Beverage) ([]string, error)
}

// InventoryAPI defines the interface for inventory management
type InventoryAPI interface {
	CheckIngredientStock(ctx context.Context, ingredients []string) (*StockStatus, error)
	ReserveIngredients(ctx context.Context, reservation *IngredientReservation) (*ReservationResult, error)
	UpdateStock(ctx context.Context, updates []StockUpdate) error
	GetLowStockAlerts(ctx context.Context) ([]LowStockAlert, error)
	PredictStockNeeds(ctx context.Context, recipes []*entities.Beverage, timeframe time.Duration) (*StockPrediction, error)
}

// QualityAPI defines the interface for quality management
type QualityAPI interface {
	CreateQualityTest(ctx context.Context, beverage *entities.Beverage, testType string) (*QualityTest, error)
	RecordTestResults(ctx context.Context, testID string, results *TestResults) error
	GetQualityHistory(ctx context.Context, beverageID string) (*QualityHistory, error)
	GenerateQualityReport(ctx context.Context, timeframe time.Duration) (*QualityReport, error)
}

// Logger interface for external integrations
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, err error, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// Data structures for external integrations

// RecipeAnalysisData contains comprehensive recipe analysis for external sharing
type RecipeAnalysisData struct {
	Beverage              *entities.Beverage           `json:"beverage"`
	NutritionalAnalysis   *NutritionalAnalysisResult   `json:"nutritional_analysis"`
	CostAnalysis          *CostBreakdown               `json:"cost_analysis"`
	CompatibilityAnalysis *CompatibilityAnalysisResult `json:"compatibility_analysis"`
	OptimizationResult    *OptimizationResult          `json:"optimization_result"`
	MarketAnalysis        *MarketFitAnalysis           `json:"market_analysis"`
	QualityMetrics        *QualityMetrics              `json:"quality_metrics"`
	CreatedAt             time.Time                    `json:"created_at"`
	Version               string                       `json:"version"`
}

// SheetInfo contains information about a created Google Sheet
type SheetInfo struct {
	SheetID     string    `json:"sheet_id"`
	URL         string    `json:"url"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	Permissions []string  `json:"permissions"`
}

// RecipeUpdates contains updates to be made to a recipe sheet
type RecipeUpdates struct {
	TestResults      *TestResults      `json:"test_results,omitempty"`
	CostUpdates      *CostBreakdown    `json:"cost_updates,omitempty"`
	QualityMetrics   *QualityMetrics   `json:"quality_metrics,omitempty"`
	CustomerFeedback *CustomerFeedback `json:"customer_feedback,omitempty"`
	Status           string            `json:"status,omitempty"`
	Notes            string            `json:"notes,omitempty"`
}

// RecipeData contains data retrieved from a recipe sheet
type RecipeData struct {
	Beverage         *entities.Beverage  `json:"beverage"`
	TestResults      []*TestResults      `json:"test_results"`
	QualityMetrics   *QualityMetrics     `json:"quality_metrics"`
	CustomerFeedback []*CustomerFeedback `json:"customer_feedback"`
	Status           string              `json:"status"`
	LastUpdated      time.Time           `json:"last_updated"`
}

// SocialMediaContent contains content for social media posting
type SocialMediaContent struct {
	Text         string   `json:"text"`
	Images       []string `json:"images"`
	Hashtags     []string `json:"hashtags"`
	BeverageID   string   `json:"beverage_id"`
	CallToAction string   `json:"call_to_action"`
	Location     string   `json:"location,omitempty"`
}

// PostResult contains the result of a social media post
type PostResult struct {
	PostID     string    `json:"post_id"`
	Platform   string    `json:"platform"`
	URL        string    `json:"url"`
	PostedAt   time.Time `json:"posted_at"`
	Reach      int       `json:"reach"`
	Engagement int       `json:"engagement"`
}

// ScheduledPost contains information about a scheduled social media post
type ScheduledPost struct {
	ScheduleID   string              `json:"schedule_id"`
	Platform     string              `json:"platform"`
	Content      *SocialMediaContent `json:"content"`
	ScheduledFor time.Time           `json:"scheduled_for"`
	Status       string              `json:"status"`
}

// EngagementMetrics contains social media engagement metrics
type EngagementMetrics struct {
	PostID       string    `json:"post_id"`
	Platform     string    `json:"platform"`
	Likes        int       `json:"likes"`
	Shares       int       `json:"shares"`
	Comments     int       `json:"comments"`
	Reach        int       `json:"reach"`
	Impressions  int       `json:"impressions"`
	ClickThrough int       `json:"click_through"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// StockStatus contains current stock status for ingredients
type StockStatus struct {
	Ingredients []IngredientStock `json:"ingredients"`
	CheckedAt   time.Time         `json:"checked_at"`
	Warnings    []string          `json:"warnings"`
}

// IngredientStock contains stock information for a single ingredient
type IngredientStock struct {
	Name          string    `json:"name"`
	CurrentStock  float64   `json:"current_stock"`
	Unit          string    `json:"unit"`
	MinThreshold  float64   `json:"min_threshold"`
	MaxCapacity   float64   `json:"max_capacity"`
	LastRestocked time.Time `json:"last_restocked"`
	Supplier      string    `json:"supplier"`
	Status        string    `json:"status"` // available, low, out_of_stock, expired
}

// IngredientReservation contains a reservation request for ingredients
type IngredientReservation struct {
	ReservationID string                      `json:"reservation_id"`
	BeverageID    string                      `json:"beverage_id"`
	Ingredients   []IngredientReservationItem `json:"ingredients"`
	ReservedFor   string                      `json:"reserved_for"`
	ExpiresAt     time.Time                   `json:"expires_at"`
	Purpose       string                      `json:"purpose"`
}

// IngredientReservationItem contains details for reserving a specific ingredient
type IngredientReservationItem struct {
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
	Priority string  `json:"priority"` // high, medium, low
}

// ReservationResult contains the result of an ingredient reservation
type ReservationResult struct {
	ReservationID string                      `json:"reservation_id"`
	Status        string                      `json:"status"` // confirmed, partial, failed
	Reserved      []IngredientReservationItem `json:"reserved"`
	Unavailable   []IngredientReservationItem `json:"unavailable"`
	ExpiresAt     time.Time                   `json:"expires_at"`
	Notes         string                      `json:"notes"`
}

// StockUpdate contains an update to ingredient stock levels
type StockUpdate struct {
	Name      string    `json:"name"`
	Change    float64   `json:"change"` // positive for additions, negative for usage
	Unit      string    `json:"unit"`
	Reason    string    `json:"reason"`
	UpdatedBy string    `json:"updated_by"`
	Timestamp time.Time `json:"timestamp"`
}

// LowStockAlert contains information about low stock items
type LowStockAlert struct {
	Name             string    `json:"name"`
	CurrentStock     float64   `json:"current_stock"`
	MinThreshold     float64   `json:"min_threshold"`
	Unit             string    `json:"unit"`
	DaysRemaining    int       `json:"days_remaining"`
	RecommendedOrder float64   `json:"recommended_order"`
	Supplier         string    `json:"supplier"`
	Priority         string    `json:"priority"`
	CreatedAt        time.Time `json:"created_at"`
}

// StockPrediction contains predicted stock needs
type StockPrediction struct {
	Timeframe   time.Duration          `json:"timeframe"`
	Predictions []IngredientPrediction `json:"predictions"`
	Confidence  float64                `json:"confidence"`
	Assumptions []string               `json:"assumptions"`
	GeneratedAt time.Time              `json:"generated_at"`
}

// IngredientPrediction contains prediction for a single ingredient
type IngredientPrediction struct {
	Name             string  `json:"name"`
	CurrentStock     float64 `json:"current_stock"`
	PredictedUsage   float64 `json:"predicted_usage"`
	RecommendedOrder float64 `json:"recommended_order"`
	Unit             string  `json:"unit"`
	ConfidenceLevel  float64 `json:"confidence_level"`
	SeasonalFactor   float64 `json:"seasonal_factor"`
}

// QualityTest contains information about a quality test
type QualityTest struct {
	TestID     string                 `json:"test_id"`
	BeverageID string                 `json:"beverage_id"`
	TestType   string                 `json:"test_type"`
	Parameters map[string]interface{} `json:"parameters"`
	CreatedAt  time.Time              `json:"created_at"`
	Status     string                 `json:"status"`
	AssignedTo string                 `json:"assigned_to"`
	DueDate    time.Time              `json:"due_date"`
}

// TestResults contains the results of a quality test
type TestResults struct {
	TestID          string                 `json:"test_id"`
	BeverageID      string                 `json:"beverage_id"`
	TestType        string                 `json:"test_type"`
	Results         map[string]interface{} `json:"results"`
	Score           float64                `json:"score"`     // Overall quality score 0-100
	PassFail        string                 `json:"pass_fail"` // pass, fail, conditional
	Notes           string                 `json:"notes"`
	TestedBy        string                 `json:"tested_by"`
	TestedAt        time.Time              `json:"tested_at"`
	Recommendations []string               `json:"recommendations"`
}

// QualityHistory contains historical quality data for a beverage
type QualityHistory struct {
	BeverageID   string         `json:"beverage_id"`
	TestResults  []*TestResults `json:"test_results"`
	AverageScore float64        `json:"average_score"`
	Trend        string         `json:"trend"` // improving, declining, stable
	LastTested   time.Time      `json:"last_tested"`
	TestCount    int            `json:"test_count"`
}

// QualityReport contains a comprehensive quality report
type QualityReport struct {
	Timeframe       time.Duration  `json:"timeframe"`
	TotalTests      int            `json:"total_tests"`
	PassRate        float64        `json:"pass_rate"`
	AverageScore    float64        `json:"average_score"`
	TopPerformers   []string       `json:"top_performers"`
	IssuesSummary   []QualityIssue `json:"issues_summary"`
	Recommendations []string       `json:"recommendations"`
	GeneratedAt     time.Time      `json:"generated_at"`
}

// QualityIssue represents a quality issue found during testing
type QualityIssue struct {
	IssueType   string    `json:"issue_type"`
	Description string    `json:"description"`
	Frequency   int       `json:"frequency"`
	Severity    string    `json:"severity"`
	BeverageIDs []string  `json:"beverage_ids"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
}

// QualityMetrics contains quality metrics for a beverage
type QualityMetrics struct {
	OverallScore     float64            `json:"overall_score"`
	TasteScore       float64            `json:"taste_score"`
	AppearanceScore  float64            `json:"appearance_score"`
	AromaScore       float64            `json:"aroma_score"`
	TextureScore     float64            `json:"texture_score"`
	ConsistencyScore float64            `json:"consistency_score"`
	Attributes       map[string]float64 `json:"attributes"`
	Defects          []string           `json:"defects"`
	Strengths        []string           `json:"strengths"`
	TestDate         time.Time          `json:"test_date"`
	Tester           string             `json:"tester"`
}

// CustomerFeedback contains customer feedback for a beverage
type CustomerFeedback struct {
	FeedbackID   string                `json:"feedback_id"`
	BeverageID   string                `json:"beverage_id"`
	CustomerID   string                `json:"customer_id"`
	Rating       float64               `json:"rating"` // 1-5 stars
	Comments     string                `json:"comments"`
	Attributes   map[string]float64    `json:"attributes"` // taste, appearance, etc.
	Demographics *CustomerDemographics `json:"demographics"`
	Location     string                `json:"location"`
	Channel      string                `json:"channel"` // in-store, online, app
	CreatedAt    time.Time             `json:"created_at"`
	Verified     bool                  `json:"verified"`
}

// NewExternalIntegrationService creates a new external integration service
func NewExternalIntegrationService(
	googleSheetsAPI GoogleSheetsAPI,
	socialMediaAPI SocialMediaAPI,
	inventoryAPI InventoryAPI,
	qualityAPI QualityAPI,
	logger Logger,
) *ExternalIntegrationService {
	return &ExternalIntegrationService{
		googleSheetsAPI: googleSheetsAPI,
		socialMediaAPI:  socialMediaAPI,
		inventoryAPI:    inventoryAPI,
		qualityAPI:      qualityAPI,
		logger:          logger,
	}
}

// CreateRecipeDocumentation creates comprehensive documentation for a recipe
func (eis *ExternalIntegrationService) CreateRecipeDocumentation(ctx context.Context, analysisData *RecipeAnalysisData) (*SheetInfo, error) {
	eis.logger.Info("Creating recipe documentation", "beverage_id", analysisData.Beverage.ID)

	// Create Google Sheet with comprehensive analysis
	sheetInfo, err := eis.googleSheetsAPI.CreateRecipeSheet(ctx, analysisData.Beverage, analysisData)
	if err != nil {
		eis.logger.Error("Failed to create recipe sheet", err, "beverage_id", analysisData.Beverage.ID)
		return nil, err
	}

	eis.logger.Info("Recipe documentation created successfully",
		"beverage_id", analysisData.Beverage.ID,
		"sheet_id", sheetInfo.SheetID,
		"url", sheetInfo.URL)

	return sheetInfo, nil
}

// ShareRecipeWithTeam shares a recipe with team members
func (eis *ExternalIntegrationService) ShareRecipeWithTeam(ctx context.Context, sheetID string, teamEmails []string) error {
	eis.logger.Info("Sharing recipe with team", "sheet_id", sheetID, "team_size", len(teamEmails))

	err := eis.googleSheetsAPI.ShareSheet(ctx, sheetID, teamEmails, "edit")
	if err != nil {
		eis.logger.Error("Failed to share recipe sheet", err, "sheet_id", sheetID)
		return err
	}

	eis.logger.Info("Recipe shared successfully", "sheet_id", sheetID)
	return nil
}

// PublishRecipeToSocialMedia publishes a recipe to social media platforms
func (eis *ExternalIntegrationService) PublishRecipeToSocialMedia(ctx context.Context, beverage *entities.Beverage, platforms []string, images []string) ([]*PostResult, error) {
	eis.logger.Info("Publishing recipe to social media",
		"beverage_id", beverage.ID,
		"platforms", platforms)

	// Generate hashtags
	hashtags, err := eis.socialMediaAPI.GenerateHashtags(ctx, beverage)
	if err != nil {
		eis.logger.Warn("Failed to generate hashtags, using defaults", "error", err)
		hashtags = []string{"#beverage", "#recipe", "#innovation"}
	}

	// Create social media content
	content := &SocialMediaContent{
		Text:         eis.generateSocialMediaText(beverage),
		Images:       images,
		Hashtags:     hashtags,
		BeverageID:   beverage.ID.String(),
		CallToAction: "Try this amazing new recipe! 🥤",
	}

	results := make([]*PostResult, 0, len(platforms))

	// Post to each platform
	for _, platform := range platforms {
		result, err := eis.socialMediaAPI.PostRecipe(ctx, platform, content)
		if err != nil {
			eis.logger.Error("Failed to post to platform", err,
				"platform", platform,
				"beverage_id", beverage.ID)
			continue
		}

		results = append(results, result)
		eis.logger.Info("Posted to social media successfully",
			"platform", platform,
			"post_id", result.PostID,
			"url", result.URL)
	}

	return results, nil
}

// CheckIngredientAvailability checks if ingredients are available for production
func (eis *ExternalIntegrationService) CheckIngredientAvailability(ctx context.Context, beverage *entities.Beverage, batchSize int) (*AvailabilityReport, error) {
	eis.logger.Info("Checking ingredient availability",
		"beverage_id", beverage.ID,
		"batch_size", batchSize)

	// Extract ingredient names
	ingredientNames := make([]string, len(beverage.Ingredients))
	for i, ingredient := range beverage.Ingredients {
		ingredientNames[i] = ingredient.Name
	}

	// Check stock status
	stockStatus, err := eis.inventoryAPI.CheckIngredientStock(ctx, ingredientNames)
	if err != nil {
		eis.logger.Error("Failed to check ingredient stock", err, "beverage_id", beverage.ID)
		return nil, err
	}

	// Analyze availability for the requested batch size
	report := &AvailabilityReport{
		BeverageID:       beverage.ID.String(),
		RequestedBatches: batchSize,
		CheckedAt:        time.Now(),
		Available:        true,
		MaxBatches:       batchSize,
		Constraints:      []string{},
		Recommendations:  []string{},
	}

	for _, ingredient := range beverage.Ingredients {
		requiredAmount := ingredient.Quantity * float64(batchSize)

		// Find stock info for this ingredient
		var stock *IngredientStock
		for _, s := range stockStatus.Ingredients {
			if s.Name == ingredient.Name {
				stock = &s
				break
			}
		}

		if stock == nil {
			report.Available = false
			report.Constraints = append(report.Constraints,
				fmt.Sprintf("No stock information for %s", ingredient.Name))
			continue
		}

		// Check if we have enough stock
		if stock.CurrentStock < requiredAmount {
			report.Available = false
			maxPossible := int(stock.CurrentStock / ingredient.Quantity)
			if maxPossible < report.MaxBatches {
				report.MaxBatches = maxPossible
			}
			report.Constraints = append(report.Constraints,
				fmt.Sprintf("Insufficient %s: need %.1f%s, have %.1f%s",
					ingredient.Name, requiredAmount, ingredient.Unit,
					stock.CurrentStock, stock.Unit))
		}

		// Check for low stock warnings
		if stock.CurrentStock <= stock.MinThreshold {
			report.Recommendations = append(report.Recommendations,
				fmt.Sprintf("Reorder %s - below minimum threshold", ingredient.Name))
		}
	}

	eis.logger.Info("Ingredient availability check completed",
		"beverage_id", beverage.ID,
		"available", report.Available,
		"max_batches", report.MaxBatches)

	return report, nil
}

// ReserveIngredientsForProduction reserves ingredients for beverage production
func (eis *ExternalIntegrationService) ReserveIngredientsForProduction(ctx context.Context, beverage *entities.Beverage, batchSize int, reservedFor string) (*ReservationResult, error) {
	eis.logger.Info("Reserving ingredients for production",
		"beverage_id", beverage.ID,
		"batch_size", batchSize,
		"reserved_for", reservedFor)

	// Create reservation request
	reservationItems := make([]IngredientReservationItem, len(beverage.Ingredients))
	for i, ingredient := range beverage.Ingredients {
		reservationItems[i] = IngredientReservationItem{
			Name:     ingredient.Name,
			Quantity: ingredient.Quantity * float64(batchSize),
			Unit:     ingredient.Unit,
			Priority: "high",
		}
	}

	reservation := &IngredientReservation{
		ReservationID: fmt.Sprintf("res_%s_%d", beverage.ID.String()[:8], time.Now().Unix()),
		BeverageID:    beverage.ID.String(),
		Ingredients:   reservationItems,
		ReservedFor:   reservedFor,
		ExpiresAt:     time.Now().Add(24 * time.Hour), // 24 hour reservation
		Purpose:       "beverage_production",
	}

	// Make reservation
	result, err := eis.inventoryAPI.ReserveIngredients(ctx, reservation)
	if err != nil {
		eis.logger.Error("Failed to reserve ingredients", err, "beverage_id", beverage.ID)
		return nil, err
	}

	eis.logger.Info("Ingredients reserved successfully",
		"beverage_id", beverage.ID,
		"reservation_id", result.ReservationID,
		"status", result.Status)

	return result, nil
}

// CreateQualityTestPlan creates a comprehensive quality test plan for a beverage
func (eis *ExternalIntegrationService) CreateQualityTestPlan(ctx context.Context, beverage *entities.Beverage, testTypes []string) ([]*QualityTest, error) {
	eis.logger.Info("Creating quality test plan",
		"beverage_id", beverage.ID,
		"test_types", testTypes)

	tests := make([]*QualityTest, 0, len(testTypes))

	for _, testType := range testTypes {
		test, err := eis.qualityAPI.CreateQualityTest(ctx, beverage, testType)
		if err != nil {
			eis.logger.Error("Failed to create quality test", err,
				"beverage_id", beverage.ID,
				"test_type", testType)
			continue
		}

		tests = append(tests, test)
		eis.logger.Info("Quality test created",
			"test_id", test.TestID,
			"test_type", testType)
	}

	return tests, nil
}

// generateSocialMediaText generates engaging text for social media posts
func (eis *ExternalIntegrationService) generateSocialMediaText(beverage *entities.Beverage) string {
	return fmt.Sprintf(`🎉 Introducing our latest creation: %s!

%s

This %s-themed beverage combines the perfect blend of flavors for an unforgettable experience.

What do you think? Would you try this innovative recipe? Let us know in the comments! 👇`,
		beverage.Name,
		beverage.Description,
		beverage.Theme)
}

// AvailabilityReport contains the result of ingredient availability check
type AvailabilityReport struct {
	BeverageID       string    `json:"beverage_id"`
	RequestedBatches int       `json:"requested_batches"`
	Available        bool      `json:"available"`
	MaxBatches       int       `json:"max_batches"`
	Constraints      []string  `json:"constraints"`
	Recommendations  []string  `json:"recommendations"`
	CheckedAt        time.Time `json:"checked_at"`
}
