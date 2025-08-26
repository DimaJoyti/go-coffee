package agents

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/config"
	"go.uber.org/zap"
)

// BeverageInventorAgent creates new coffee recipes and analyzes beverage trends
type BeverageInventorAgent struct {
	id           string
	config       config.AgentConfig
	logger       *zap.Logger
	status       AgentStatus
	capabilities []string
	
	// Agent-specific state
	recipes      map[string]*Recipe
	trends       map[string]*Trend
	lastAnalysis time.Time
}

// Recipe represents a coffee recipe
type Recipe struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Ingredients []Ingredient           `json:"ingredients"`
	Steps       []string               `json:"steps"`
	Difficulty  string                 `json:"difficulty"`
	PrepTime    int                    `json:"prep_time_minutes"`
	Tags        []string               `json:"tags"`
	Nutrition   map[string]interface{} `json:"nutrition"`
	CreatedAt   time.Time              `json:"created_at"`
}

// Ingredient represents a recipe ingredient
type Ingredient struct {
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	Unit     string  `json:"unit"`
	Optional bool    `json:"optional"`
}

// Trend represents a beverage trend
type Trend struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Popularity  float64   `json:"popularity"`
	Growth      float64   `json:"growth_rate"`
	Season      string    `json:"season"`
	Demographics []string `json:"demographics"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewBeverageInventorAgent creates a new beverage inventor agent
func NewBeverageInventorAgent(config config.AgentConfig, logger *zap.Logger) Agent {
	return &BeverageInventorAgent{
		id:     "beverage-inventor",
		config: config,
		logger: logger,
		status: AgentStatusOffline,
		capabilities: []string{
			"analyze_order",
			"create_recipe",
			"suggest_modifications",
			"analyze_trends",
			"generate_seasonal_menu",
			"optimize_ingredients",
		},
		recipes: make(map[string]*Recipe),
		trends:  make(map[string]*Trend),
	}
}

// GetID returns the agent ID
func (a *BeverageInventorAgent) GetID() string {
	return a.id
}

// GetType returns the agent type
func (a *BeverageInventorAgent) GetType() string {
	return "beverage-inventor"
}

// GetCapabilities returns the agent capabilities
func (a *BeverageInventorAgent) GetCapabilities() []string {
	return a.capabilities
}

// GetStatus returns the agent status
func (a *BeverageInventorAgent) GetStatus() AgentStatus {
	return a.status
}

// IsHealthy returns whether the agent is healthy
func (a *BeverageInventorAgent) IsHealthy() bool {
	return a.status == AgentStatusOnline
}

// Start starts the agent
func (a *BeverageInventorAgent) Start(ctx context.Context) error {
	a.logger.Info("Starting Beverage Inventor Agent")
	
	// Initialize with some default recipes
	a.initializeDefaultRecipes()
	a.initializeDefaultTrends()
	
	a.status = AgentStatusOnline
	a.logger.Info("Beverage Inventor Agent started successfully")
	
	return nil
}

// Stop stops the agent
func (a *BeverageInventorAgent) Stop() error {
	a.logger.Info("Stopping Beverage Inventor Agent")
	a.status = AgentStatusOffline
	return nil
}

// ExecuteAction executes an action
func (a *BeverageInventorAgent) ExecuteAction(ctx context.Context, action string, inputs map[string]interface{}) (map[string]interface{}, error) {
	a.logger.Info("Executing action",
		zap.String("action", action),
		zap.Any("inputs", inputs),
	)

	switch action {
	case "analyze_order":
		return a.analyzeOrder(ctx, inputs)
	case "create_recipe":
		return a.createRecipe(ctx, inputs)
	case "suggest_modifications":
		return a.suggestModifications(ctx, inputs)
	case "analyze_trends":
		return a.analyzeTrends(ctx, inputs)
	case "generate_seasonal_menu":
		return a.generateSeasonalMenu(ctx, inputs)
	case "optimize_ingredients":
		return a.optimizeIngredients(ctx, inputs)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// ReceiveMessage receives a message from another agent
func (a *BeverageInventorAgent) ReceiveMessage(ctx context.Context, message *AgentMessage) error {
	a.logger.Info("Received message",
		zap.String("from", message.FromAgent),
		zap.String("type", message.Type),
	)

	// Process message based on type
	switch message.Type {
	case "recipe_request":
		return a.handleRecipeRequest(ctx, message)
	case "trend_analysis_request":
		return a.handleTrendAnalysisRequest(ctx, message)
	default:
		a.logger.Debug("Unhandled message type", zap.String("type", message.Type))
	}

	return nil
}

// SendMessage sends a message to another agent
func (a *BeverageInventorAgent) SendMessage(ctx context.Context, message *AgentMessage) error {
	// This would typically be handled by the orchestrator
	a.logger.Info("Sending message",
		zap.String("to", message.ToAgent),
		zap.String("type", message.Type),
	)
	return nil
}

// analyzeOrder analyzes a coffee order and provides recommendations
func (a *BeverageInventorAgent) analyzeOrder(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	orderType, _ := inputs["order_type"].(string)
	
	analysis := map[string]interface{}{
		"order_complexity": "medium",
		"estimated_prep_time": 5,
		"recommended_modifications": []string{
			"Add extra shot for stronger flavor",
			"Consider oat milk for creamier texture",
		},
		"seasonal_suggestions": []string{
			"Pumpkin spice variant for fall",
			"Iced version for summer",
		},
		"nutritional_info": map[string]interface{}{
			"calories": 150,
			"caffeine_mg": 95,
			"sugar_g": 12,
		},
	}

	a.logger.Info("Order analysis completed",
		zap.String("order_type", orderType),
		zap.Any("analysis", analysis),
	)

	return analysis, nil
}

// createRecipe creates a new coffee recipe
func (a *BeverageInventorAgent) createRecipe(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	recipeName, _ := inputs["name"].(string)
	if recipeName == "" {
		recipeName = a.generateRandomRecipeName()
	}

	recipe := &Recipe{
		ID:          fmt.Sprintf("recipe_%d", time.Now().Unix()),
		Name:        recipeName,
		Description: "AI-generated coffee recipe",
		Ingredients: []Ingredient{
			{Name: "Espresso", Amount: 2, Unit: "shots", Optional: false},
			{Name: "Steamed Milk", Amount: 6, Unit: "oz", Optional: false},
			{Name: "Vanilla Syrup", Amount: 1, Unit: "pump", Optional: true},
		},
		Steps: []string{
			"Pull 2 shots of espresso",
			"Steam milk to 150°F",
			"Add vanilla syrup if desired",
			"Pour steamed milk into espresso",
			"Create latte art",
		},
		Difficulty: "Medium",
		PrepTime:   5,
		Tags:       []string{"coffee", "milk", "espresso"},
		Nutrition: map[string]interface{}{
			"calories": 150,
			"caffeine": 95,
			"protein": 8,
		},
		CreatedAt: time.Now(),
	}

	a.recipes[recipe.ID] = recipe

	return map[string]interface{}{
		"recipe": recipe,
		"status": "created",
	}, nil
}

// suggestModifications suggests modifications to existing recipes
func (a *BeverageInventorAgent) suggestModifications(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	modifications := []map[string]interface{}{
		{
			"type": "ingredient_substitution",
			"suggestion": "Replace regular milk with oat milk for dairy-free option",
			"impact": "Creamier texture, nutty flavor",
		},
		{
			"type": "flavor_enhancement",
			"suggestion": "Add cinnamon powder for warming spice",
			"impact": "Enhanced aroma and taste complexity",
		},
		{
			"type": "temperature_variant",
			"suggestion": "Create iced version for summer menu",
			"impact": "Seasonal appeal, refreshing alternative",
		},
	}

	return map[string]interface{}{
		"modifications": modifications,
		"total_suggestions": len(modifications),
	}, nil
}

// analyzeTrends analyzes current beverage trends
func (a *BeverageInventorAgent) analyzeTrends(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Simulate trend analysis
	trends := []map[string]interface{}{
		{
			"name": "Plant-based milk alternatives",
			"popularity": 85.5,
			"growth_rate": 12.3,
			"recommendation": "Expand oat and almond milk options",
		},
		{
			"name": "Cold brew variations",
			"popularity": 78.2,
			"growth_rate": 8.7,
			"recommendation": "Introduce nitro cold brew and flavored variants",
		},
		{
			"name": "Functional beverages",
			"popularity": 72.1,
			"growth_rate": 15.4,
			"recommendation": "Add adaptogens and wellness-focused ingredients",
		},
	}

	a.lastAnalysis = time.Now()

	return map[string]interface{}{
		"trends": trends,
		"analysis_date": a.lastAnalysis,
		"market_insights": "Strong demand for health-conscious and sustainable options",
	}, nil
}

// generateSeasonalMenu generates a seasonal menu
func (a *BeverageInventorAgent) generateSeasonalMenu(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	season, _ := inputs["season"].(string)
	if season == "" {
		season = a.getCurrentSeason()
	}

	var menuItems []map[string]interface{}

	switch season {
	case "spring":
		menuItems = []map[string]interface{}{
			{"name": "Lavender Honey Latte", "description": "Floral and sweet spring blend"},
			{"name": "Matcha Cherry Blossom", "description": "Japanese-inspired green tea drink"},
		}
	case "summer":
		menuItems = []map[string]interface{}{
			{"name": "Iced Coconut Cold Brew", "description": "Refreshing tropical cold coffee"},
			{"name": "Watermelon Mint Refresher", "description": "Hydrating fruit-based beverage"},
		}
	case "fall":
		menuItems = []map[string]interface{}{
			{"name": "Pumpkin Spice Latte", "description": "Classic autumn favorite"},
			{"name": "Apple Cinnamon Chai", "description": "Warming spiced tea blend"},
		}
	case "winter":
		menuItems = []map[string]interface{}{
			{"name": "Peppermint Mocha", "description": "Festive chocolate and mint combination"},
			{"name": "Gingerbread Latte", "description": "Holiday-spiced coffee drink"},
		}
	}

	return map[string]interface{}{
		"season": season,
		"menu_items": menuItems,
		"total_items": len(menuItems),
	}, nil
}

// optimizeIngredients optimizes ingredient usage
func (a *BeverageInventorAgent) optimizeIngredients(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	optimizations := []map[string]interface{}{
		{
			"ingredient": "Milk",
			"current_usage": "High",
			"optimization": "Implement portion control to reduce waste",
			"potential_savings": "15%",
		},
		{
			"ingredient": "Syrups",
			"current_usage": "Medium",
			"optimization": "Standardize pump counts across all drinks",
			"potential_savings": "8%",
		},
	}

	return map[string]interface{}{
		"optimizations": optimizations,
		"total_potential_savings": "23%",
	}, nil
}

// Helper methods
func (a *BeverageInventorAgent) initializeDefaultRecipes() {
	// Initialize with some default recipes
	defaultRecipes := []*Recipe{
		{
			ID:          "classic_latte",
			Name:        "Classic Latte",
			Description: "Traditional espresso and steamed milk",
			Ingredients: []Ingredient{
				{Name: "Espresso", Amount: 2, Unit: "shots"},
				{Name: "Steamed Milk", Amount: 6, Unit: "oz"},
			},
			Steps:     []string{"Pull espresso", "Steam milk", "Combine"},
			Difficulty: "Easy",
			PrepTime:  4,
			CreatedAt: time.Now(),
		},
	}

	for _, recipe := range defaultRecipes {
		a.recipes[recipe.ID] = recipe
	}
}

func (a *BeverageInventorAgent) initializeDefaultTrends() {
	// Initialize with some default trends
	a.trends["plant_milk"] = &Trend{
		ID:          "plant_milk",
		Name:        "Plant-based Milk",
		Popularity:  85.0,
		Growth:      12.0,
		UpdatedAt:   time.Now(),
	}
}

func (a *BeverageInventorAgent) generateRandomRecipeName() string {
	adjectives := []string{"Smooth", "Rich", "Creamy", "Bold", "Delicate"}
	nouns := []string{"Latte", "Cappuccino", "Mocha", "Macchiato", "Americano"}
	
	return fmt.Sprintf("%s %s", 
		adjectives[rand.Intn(len(adjectives))], 
		nouns[rand.Intn(len(nouns))])
}

func (a *BeverageInventorAgent) getCurrentSeason() string {
	month := time.Now().Month()
	switch {
	case month >= 3 && month <= 5:
		return "spring"
	case month >= 6 && month <= 8:
		return "summer"
	case month >= 9 && month <= 11:
		return "fall"
	default:
		return "winter"
	}
}

func (a *BeverageInventorAgent) handleRecipeRequest(ctx context.Context, message *AgentMessage) error {
	// Handle recipe request from other agents
	a.logger.Info("Handling recipe request", zap.String("from", message.FromAgent))
	return nil
}

func (a *BeverageInventorAgent) handleTrendAnalysisRequest(ctx context.Context, message *AgentMessage) error {
	// Handle trend analysis request from other agents
	a.logger.Info("Handling trend analysis request", zap.String("from", message.FromAgent))
	return nil
}

// Placeholder agent constructors for the other 8 agents
func NewInventoryManagerAgent(config config.AgentConfig, logger *zap.Logger) Agent {
	return &BeverageInventorAgent{id: "inventory-manager", config: config, logger: logger, status: AgentStatusOffline, capabilities: []string{"check_availability", "forecast_demand", "manage_stock"}}
}

func NewTaskManagerAgent(config config.AgentConfig, logger *zap.Logger) Agent {
	return &BeverageInventorAgent{id: "task-manager", config: config, logger: logger, status: AgentStatusOffline, capabilities: []string{"create_task", "assign_task", "track_progress"}}
}

func NewSocialMediaAgent(config config.AgentConfig, logger *zap.Logger) Agent {
	return &BeverageInventorAgent{id: "social-media", config: config, logger: logger, status: AgentStatusOffline, capabilities: []string{"generate_daily_content", "post_updates", "analyze_engagement"}}
}

func NewFeedbackAnalystAgent(config config.AgentConfig, logger *zap.Logger) Agent {
	return &BeverageInventorAgent{id: "feedback-analyst", config: config, logger: logger, status: AgentStatusOffline, capabilities: []string{"analyze_feedback", "generate_insights", "track_sentiment"}}
}

func NewSchedulerAgent(config config.AgentConfig, logger *zap.Logger) Agent {
	return &BeverageInventorAgent{id: "scheduler", config: config, logger: logger, status: AgentStatusOffline, capabilities: []string{"schedule_task", "optimize_schedule", "manage_calendar"}}
}

func NewInterLocationCoordinatorAgent(config config.AgentConfig, logger *zap.Logger) Agent {
	return &BeverageInventorAgent{id: "inter-location-coordinator", config: config, logger: logger, status: AgentStatusOffline, capabilities: []string{"coordinate_operations", "sync_inventory", "manage_transfers"}}
}

func NewNotifierAgent(config config.AgentConfig, logger *zap.Logger) Agent {
	return &BeverageInventorAgent{id: "notifier", config: config, logger: logger, status: AgentStatusOffline, capabilities: []string{"send_notification", "manage_alerts", "broadcast_updates"}}
}

func NewTastingCoordinatorAgent(config config.AgentConfig, logger *zap.Logger) Agent {
	return &BeverageInventorAgent{id: "tasting-coordinator", config: config, logger: logger, status: AgentStatusOffline, capabilities: []string{"schedule_tasting", "coordinate_feedback", "manage_sessions"}}
}
