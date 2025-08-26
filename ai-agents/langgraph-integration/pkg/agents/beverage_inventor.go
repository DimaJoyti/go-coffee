// Package agents provides specific agent implementations
package agents

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/ai-agents/langgraph-integration/pkg/graph"
)

// BeverageInventorAgent creates innovative coffee beverages and recipes
type BeverageInventorAgent struct {
	*BaseAgent
	recipeDatabase []Recipe
	flavorProfiles map[string][]string
	seasonalIngredients map[string][]string
}

// Recipe represents a coffee recipe
type Recipe struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Ingredients  []Ingredient      `json:"ingredients"`
	Instructions []string          `json:"instructions"`
	PrepTime     int               `json:"prep_time_minutes"`
	Difficulty   string            `json:"difficulty"`
	Season       string            `json:"season"`
	Tags         []string          `json:"tags"`
	Nutrition    NutritionInfo     `json:"nutrition"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Ingredient represents a recipe ingredient
type Ingredient struct {
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	Unit     string  `json:"unit"`
	Type     string  `json:"type"` // "base", "flavor", "topping", "sweetener"
	Optional bool    `json:"optional"`
}

// NutritionInfo represents nutritional information
type NutritionInfo struct {
	Calories     int     `json:"calories"`
	Caffeine     float64 `json:"caffeine_mg"`
	Sugar        float64 `json:"sugar_g"`
	Fat          float64 `json:"fat_g"`
	Protein      float64 `json:"protein_g"`
	Carbs        float64 `json:"carbs_g"`
}

// BeverageRequest represents a request for beverage creation
type BeverageRequest struct {
	Season           string   `json:"season"`
	FlavorProfile    string   `json:"flavor_profile"`
	DietaryRestrictions []string `json:"dietary_restrictions"`
	CaffeineLevel    string   `json:"caffeine_level"`
	Temperature      string   `json:"temperature"`
	Occasion         string   `json:"occasion"`
	CustomerPreferences map[string]interface{} `json:"customer_preferences"`
}

// NewBeverageInventorAgent creates a new beverage inventor agent
func NewBeverageInventorAgent() *BeverageInventorAgent {
	config := &AgentConfig{
		Name:        "Beverage Inventor",
		Description: "Creates innovative coffee beverages and recipes",
		Version:     "1.0.0",
		Timeout:     45 * time.Second,
		MaxRetries:  3,
		Tools:       []string{"recipe_generator", "flavor_analyzer", "nutrition_calculator"},
		Config: map[string]interface{}{
			"creativity_level": "high",
			"seasonal_focus":   true,
		},
	}

	agent := &BeverageInventorAgent{
		BaseAgent: NewBaseAgent(graph.AgentTypeBeverageInventor, config),
		recipeDatabase: make([]Recipe, 0),
		flavorProfiles: map[string][]string{
			"sweet": {"vanilla", "caramel", "chocolate", "honey", "maple"},
			"spicy": {"cinnamon", "nutmeg", "cardamom", "ginger", "clove"},
			"fruity": {"orange", "lemon", "berry", "apple", "cherry"},
			"nutty": {"almond", "hazelnut", "pecan", "walnut", "coconut"},
			"floral": {"lavender", "rose", "jasmine", "elderflower"},
		},
		seasonalIngredients: map[string][]string{
			"spring": {"lavender", "lemon", "mint", "green tea", "honey"},
			"summer": {"iced", "cold brew", "fruit", "coconut", "mint"},
			"fall": {"pumpkin", "cinnamon", "nutmeg", "apple", "caramel"},
			"winter": {"peppermint", "chocolate", "gingerbread", "eggnog", "spices"},
		},
	}

	// Initialize with some base recipes
	agent.initializeRecipeDatabase()

	return agent
}

// Execute executes the beverage inventor agent
func (a *BeverageInventorAgent) Execute(ctx context.Context, state *graph.AgentState) (*AgentExecutionResult, error) {
	return a.ExecuteWithMetrics(ctx, state, a.executeCore)
}

// executeCore contains the core business logic for beverage invention
func (a *BeverageInventorAgent) executeCore(ctx context.Context, state *graph.AgentState) (*AgentExecutionResult, error) {
	log.Printf("BeverageInventorAgent: Starting recipe creation for workflow %s", state.WorkflowID)

	// Validate input
	if err := a.ValidateInput(state); err != nil {
		return a.CreateErrorResult(err), err
	}

	// Parse beverage request from input data
	request, err := a.parseBeverageRequest(state.InputData)
	if err != nil {
		return a.CreateErrorResult(fmt.Errorf("failed to parse beverage request: %w", err)), err
	}

	// Create recipe based on request
	recipe, err := a.createRecipe(ctx, request)
	if err != nil {
		return a.CreateErrorResult(fmt.Errorf("failed to create recipe: %w", err)), err
	}

	// Analyze recipe for quality and feasibility
	analysis, err := a.analyzeRecipe(ctx, recipe)
	if err != nil {
		return a.CreateErrorResult(fmt.Errorf("failed to analyze recipe: %w", err)), err
	}

	// Generate variations if requested
	variations := make([]Recipe, 0)
	if shouldGenerateVariations(state.InputData) {
		variations, err = a.generateRecipeVariations(ctx, recipe, 3)
		if err != nil {
			log.Printf("Warning: Failed to generate variations: %v", err)
		}
	}

	// Create result
	result := a.CreateSuccessResult(map[string]interface{}{
		"recipe":     recipe,
		"analysis":   analysis,
		"variations": variations,
		"request":    request,
	})

	// Add metadata
	result.AddMetadata("recipe_id", recipe.ID)
	result.AddMetadata("season", request.Season)
	result.AddMetadata("flavor_profile", request.FlavorProfile)
	result.AddMetadata("variations_count", len(variations))

	// Determine next agent based on workflow
	if needsInventoryCheck(state.InputData) {
		result.SetNextAgent(graph.AgentTypeInventoryManager)
	}

	// Add message to state
	state.AddAIMessage(
		fmt.Sprintf("Created new recipe: %s (%s)", recipe.Name, recipe.Description),
		graph.AgentTypeBeverageInventor,
	)

	log.Printf("BeverageInventorAgent: Successfully created recipe %s", recipe.Name)
	return result, nil
}

// parseBeverageRequest parses the beverage request from input data
func (a *BeverageInventorAgent) parseBeverageRequest(inputData map[string]interface{}) (*BeverageRequest, error) {
	request := &BeverageRequest{
		Season:              getString(inputData, "season", "spring"),
		FlavorProfile:       getString(inputData, "flavor_profile", "balanced"),
		DietaryRestrictions: getStringSlice(inputData, "dietary_restrictions"),
		CaffeineLevel:       getString(inputData, "caffeine_level", "medium"),
		Temperature:         getString(inputData, "temperature", "hot"),
		Occasion:           getString(inputData, "occasion", "daily"),
		CustomerPreferences: getMap(inputData, "customer_preferences"),
	}

	return request, nil
}

// createRecipe creates a new recipe based on the request
func (a *BeverageInventorAgent) createRecipe(ctx context.Context, request *BeverageRequest) (*Recipe, error) {
	// Generate recipe name
	name := a.generateRecipeName(request)
	
	// Select base ingredients
	baseIngredients := a.selectBaseIngredients(request)
	
	// Add flavor ingredients
	flavorIngredients := a.selectFlavorIngredients(request)
	
	// Add seasonal ingredients
	seasonalIngredients := a.selectSeasonalIngredients(request)
	
	// Combine all ingredients
	allIngredients := append(baseIngredients, flavorIngredients...)
	allIngredients = append(allIngredients, seasonalIngredients...)
	
	// Generate instructions
	instructions := a.generateInstructions(allIngredients, request)
	
	// Calculate nutrition
	nutrition := a.calculateNutrition(allIngredients)
	
	recipe := &Recipe{
		ID:           fmt.Sprintf("recipe_%d", time.Now().Unix()),
		Name:         name,
		Description:  a.generateDescription(request, allIngredients),
		Ingredients:  allIngredients,
		Instructions: instructions,
		PrepTime:     a.estimatePrepTime(allIngredients),
		Difficulty:   a.assessDifficulty(allIngredients, instructions),
		Season:       request.Season,
		Tags:         a.generateTags(request, allIngredients),
		Nutrition:    nutrition,
		Metadata: map[string]interface{}{
			"created_at":     time.Now().UTC(),
			"agent_version":  a.config.Version,
			"request_id":     fmt.Sprintf("req_%d", time.Now().UnixNano()),
		},
	}

	// Add to database
	a.recipeDatabase = append(a.recipeDatabase, *recipe)
	
	return recipe, nil
}

// generateRecipeName generates a creative name for the recipe
func (a *BeverageInventorAgent) generateRecipeName(request *BeverageRequest) string {
	seasonalPrefixes := map[string][]string{
		"spring": {"Blooming", "Fresh", "Garden", "Morning"},
		"summer": {"Tropical", "Iced", "Refreshing", "Sunny"},
		"fall": {"Autumn", "Spiced", "Harvest", "Cozy"},
		"winter": {"Winter", "Warm", "Holiday", "Festive"},
	}

	flavorSuffixes := map[string][]string{
		"sweet": {"Delight", "Dream", "Bliss", "Heaven"},
		"spicy": {"Kick", "Warmth", "Spice", "Fire"},
		"fruity": {"Burst", "Splash", "Twist", "Fusion"},
		"nutty": {"Crunch", "Smooth", "Rich", "Velvet"},
	}

	prefixes := seasonalPrefixes[request.Season]
	suffixes := flavorSuffixes[request.FlavorProfile]

	if len(prefixes) == 0 {
		prefixes = []string{"Signature", "Special", "Artisan"}
	}
	if len(suffixes) == 0 {
		suffixes = []string{"Brew", "Blend", "Creation"}
	}

	prefix := prefixes[rand.Intn(len(prefixes))]
	suffix := suffixes[rand.Intn(len(suffixes))]

	return fmt.Sprintf("%s Coffee %s", prefix, suffix)
}

// selectBaseIngredients selects base coffee ingredients
func (a *BeverageInventorAgent) selectBaseIngredients(request *BeverageRequest) []Ingredient {
	ingredients := []Ingredient{
		{Name: "Espresso", Amount: 2, Unit: "shots", Type: "base"},
	}

	// Add milk based on temperature preference
	if request.Temperature == "hot" {
		ingredients = append(ingredients, Ingredient{
			Name: "Steamed Milk", Amount: 6, Unit: "oz", Type: "base",
		})
	} else {
		ingredients = append(ingredients, Ingredient{
			Name: "Cold Milk", Amount: 6, Unit: "oz", Type: "base",
		})
	}

	return ingredients
}

// selectFlavorIngredients selects flavor ingredients based on profile
func (a *BeverageInventorAgent) selectFlavorIngredients(request *BeverageRequest) []Ingredient {
	flavors := a.flavorProfiles[request.FlavorProfile]
	if len(flavors) == 0 {
		return []Ingredient{}
	}

	// Select 1-2 flavors randomly
	numFlavors := 1 + rand.Intn(2)
	selectedFlavors := make([]string, 0, numFlavors)
	
	for i := 0; i < numFlavors && i < len(flavors); i++ {
		flavor := flavors[rand.Intn(len(flavors))]
		selectedFlavors = append(selectedFlavors, flavor)
	}

	ingredients := make([]Ingredient, 0, len(selectedFlavors))
	for _, flavor := range selectedFlavors {
		ingredients = append(ingredients, Ingredient{
			Name:   fmt.Sprintf("%s Syrup", strings.Title(flavor)),
			Amount: 0.5,
			Unit:   "oz",
			Type:   "flavor",
		})
	}

	return ingredients
}

// selectSeasonalIngredients selects seasonal ingredients
func (a *BeverageInventorAgent) selectSeasonalIngredients(request *BeverageRequest) []Ingredient {
	seasonalItems := a.seasonalIngredients[request.Season]
	if len(seasonalItems) == 0 {
		return []Ingredient{}
	}

	// Select one seasonal ingredient
	seasonal := seasonalItems[rand.Intn(len(seasonalItems))]
	
	return []Ingredient{
		{
			Name:     fmt.Sprintf("%s Extract", strings.Title(seasonal)),
			Amount:   0.25,
			Unit:     "tsp",
			Type:     "flavor",
			Optional: true,
		},
	}
}

// generateInstructions generates preparation instructions
func (a *BeverageInventorAgent) generateInstructions(ingredients []Ingredient, request *BeverageRequest) []string {
	instructions := []string{
		"Prepare espresso shots using freshly ground coffee beans",
	}

	if request.Temperature == "hot" {
		instructions = append(instructions, "Steam milk to 150-160°F with microfoam")
	} else {
		instructions = append(instructions, "Chill milk and prepare cold brewing setup")
	}

	// Add flavor instructions
	for _, ingredient := range ingredients {
		if ingredient.Type == "flavor" {
			instructions = append(instructions, 
				fmt.Sprintf("Add %g %s of %s to the cup", ingredient.Amount, ingredient.Unit, ingredient.Name))
		}
	}

	instructions = append(instructions, 
		"Pour espresso into cup",
		"Add milk slowly while stirring",
		"Garnish as desired and serve immediately")

	return instructions
}

// calculateNutrition calculates nutritional information
func (a *BeverageInventorAgent) calculateNutrition(ingredients []Ingredient) NutritionInfo {
	// Simplified nutrition calculation
	nutrition := NutritionInfo{
		Calories: 150,
		Caffeine: 150.0,
		Sugar:    12.0,
		Fat:      3.5,
		Protein:  8.0,
		Carbs:    18.0,
	}

	// Adjust based on ingredients
	for _, ingredient := range ingredients {
		if strings.Contains(strings.ToLower(ingredient.Name), "syrup") {
			nutrition.Calories += int(ingredient.Amount * 20)
			nutrition.Sugar += ingredient.Amount * 5
		}
	}

	return nutrition
}

// estimatePrepTime estimates preparation time
func (a *BeverageInventorAgent) estimatePrepTime(ingredients []Ingredient) int {
	baseTime := 3 // minutes
	
	for _, ingredient := range ingredients {
		if ingredient.Type == "flavor" {
			baseTime += 1
		}
	}

	return baseTime
}

// assessDifficulty assesses recipe difficulty
func (a *BeverageInventorAgent) assessDifficulty(ingredients []Ingredient, instructions []string) string {
	if len(ingredients) <= 3 && len(instructions) <= 4 {
		return "easy"
	} else if len(ingredients) <= 6 && len(instructions) <= 7 {
		return "medium"
	}
	return "hard"
}

// generateTags generates tags for the recipe
func (a *BeverageInventorAgent) generateTags(request *BeverageRequest, ingredients []Ingredient) []string {
	tags := []string{request.Season, request.FlavorProfile, request.Temperature}
	
	for _, ingredient := range ingredients {
		if ingredient.Type == "flavor" {
			tags = append(tags, strings.ToLower(strings.Split(ingredient.Name, " ")[0]))
		}
	}

	return tags
}

// generateDescription generates a description for the recipe
func (a *BeverageInventorAgent) generateDescription(request *BeverageRequest, ingredients []Ingredient) string {
	return fmt.Sprintf("A delightful %s coffee blend perfect for %s, featuring %s flavors and crafted for %s enjoyment.",
		request.FlavorProfile, request.Season, request.Temperature, request.Occasion)
}

// analyzeRecipe analyzes the recipe for quality and feasibility
func (a *BeverageInventorAgent) analyzeRecipe(ctx context.Context, recipe *Recipe) (map[string]interface{}, error) {
	analysis := map[string]interface{}{
		"quality_score":    a.calculateQualityScore(recipe),
		"feasibility":      a.assessFeasibility(recipe),
		"cost_estimate":    a.estimateCost(recipe),
		"allergen_info":    a.identifyAllergens(recipe),
		"dietary_compliance": a.checkDietaryCompliance(recipe),
	}

	return analysis, nil
}

// calculateQualityScore calculates a quality score for the recipe
func (a *BeverageInventorAgent) calculateQualityScore(recipe *Recipe) float64 {
	score := 70.0 // base score

	// Bonus for balanced ingredients
	if len(recipe.Ingredients) >= 3 && len(recipe.Ingredients) <= 6 {
		score += 10.0
	}

	// Bonus for clear instructions
	if len(recipe.Instructions) >= 4 {
		score += 10.0
	}

	// Bonus for reasonable prep time
	if recipe.PrepTime >= 3 && recipe.PrepTime <= 10 {
		score += 10.0
	}

	return min(score, 100.0)
}

// assessFeasibility assesses recipe feasibility
func (a *BeverageInventorAgent) assessFeasibility(recipe *Recipe) string {
	if recipe.PrepTime <= 5 && recipe.Difficulty == "easy" {
		return "high"
	} else if recipe.PrepTime <= 10 && recipe.Difficulty != "hard" {
		return "medium"
	}
	return "low"
}

// estimateCost estimates the cost of the recipe
func (a *BeverageInventorAgent) estimateCost(recipe *Recipe) float64 {
	baseCost := 2.50 // base coffee cost
	
	for _, ingredient := range recipe.Ingredients {
		if ingredient.Type == "flavor" {
			baseCost += 0.50
		}
	}

	return baseCost
}

// identifyAllergens identifies potential allergens
func (a *BeverageInventorAgent) identifyAllergens(recipe *Recipe) []string {
	allergens := []string{}
	
	for _, ingredient := range recipe.Ingredients {
		name := strings.ToLower(ingredient.Name)
		if strings.Contains(name, "milk") {
			allergens = append(allergens, "dairy")
		}
		if strings.Contains(name, "nut") {
			allergens = append(allergens, "nuts")
		}
	}

	return allergens
}

// checkDietaryCompliance checks dietary compliance
func (a *BeverageInventorAgent) checkDietaryCompliance(recipe *Recipe) map[string]bool {
	compliance := map[string]bool{
		"vegan":       true,
		"vegetarian":  true,
		"gluten_free": true,
		"dairy_free":  true,
	}

	for _, ingredient := range recipe.Ingredients {
		name := strings.ToLower(ingredient.Name)
		if strings.Contains(name, "milk") {
			compliance["vegan"] = false
			compliance["dairy_free"] = false
		}
	}

	return compliance
}

// generateRecipeVariations generates variations of the base recipe
func (a *BeverageInventorAgent) generateRecipeVariations(ctx context.Context, baseRecipe *Recipe, count int) ([]Recipe, error) {
	variations := make([]Recipe, 0, count)

	for i := 0; i < count; i++ {
		variation := *baseRecipe // copy
		variation.ID = fmt.Sprintf("%s_var_%d", baseRecipe.ID, i+1)
		variation.Name = fmt.Sprintf("%s - Variation %d", baseRecipe.Name, i+1)
		
		// Modify ingredients slightly
		if len(variation.Ingredients) > 2 {
			// Change one flavor ingredient
			for j, ingredient := range variation.Ingredients {
				if ingredient.Type == "flavor" {
					variation.Ingredients[j].Amount *= (0.8 + rand.Float64()*0.4) // 0.8 to 1.2 multiplier
					break
				}
			}
		}

		variations = append(variations, variation)
	}

	return variations, nil
}

// initializeRecipeDatabase initializes the recipe database with some base recipes
func (a *BeverageInventorAgent) initializeRecipeDatabase() {
	// Add some base recipes for reference
	baseRecipes := []Recipe{
		{
			ID:          "base_latte",
			Name:        "Classic Latte",
			Description: "Traditional espresso and steamed milk",
			Ingredients: []Ingredient{
				{Name: "Espresso", Amount: 2, Unit: "shots", Type: "base"},
				{Name: "Steamed Milk", Amount: 6, Unit: "oz", Type: "base"},
			},
			Instructions: []string{
				"Prepare espresso shots",
				"Steam milk to 150°F",
				"Pour espresso into cup",
				"Add steamed milk",
			},
			PrepTime:   3,
			Difficulty: "easy",
			Season:     "all",
			Tags:       []string{"classic", "milk", "espresso"},
		},
	}

	a.recipeDatabase = append(a.recipeDatabase, baseRecipes...)
}

// Helper functions
func getString(data map[string]interface{}, key, defaultValue string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return defaultValue
}

func getStringSlice(data map[string]interface{}, key string) []string {
	if val, ok := data[key].([]interface{}); ok {
		result := make([]string, len(val))
		for i, v := range val {
			if s, ok := v.(string); ok {
				result[i] = s
			}
		}
		return result
	}
	return []string{}
}

func getMap(data map[string]interface{}, key string) map[string]interface{} {
	if val, ok := data[key].(map[string]interface{}); ok {
		return val
	}
	return make(map[string]interface{})
}

func shouldGenerateVariations(data map[string]interface{}) bool {
	if val, ok := data["generate_variations"].(bool); ok {
		return val
	}
	return false
}

func needsInventoryCheck(data map[string]interface{}) bool {
	if val, ok := data["check_inventory"].(bool); ok {
		return val
	}
	return true // default to checking inventory
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
