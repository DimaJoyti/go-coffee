package main

import (
	"context"
	"fmt"
	"log"

	"go-coffee-ai-agents/beverage-inventor-agent/internal/application/usecases"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/entities"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/services"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/infrastructure/factory"
	"go-coffee-ai-agents/beverage-inventor-agent/internal/infrastructure/logger"
)

// Enhanced Beverage Inventor Demo
// This demonstrates the comprehensive capabilities of the enhanced beverage inventor agent

func main() {
	fmt.Println("🚀 Enhanced Beverage Inventor Agent Demo")
	fmt.Println("=========================================")

	// Initialize logger
	appLogger := logger.New()

	// Create enhanced services
	servicesFactory := factory.NewEnhancedServicesFactory(appLogger)
	enhancedServices := servicesFactory.CreateEnhancedServices(nil)

	// Create use case with enhanced services
	beverageUseCase := usecases.NewBeverageInventorUseCase(
		nil, nil, nil, nil, nil,
		enhancedServices.NutritionalAnalyzer,
		enhancedServices.CostCalculator,
		enhancedServices.CompatibilityAnalyzer,
		enhancedServices.RecipeOptimizer,
		appLogger,
	)

	ctx := context.Background()

	// Demo 1: Enhanced Beverage Creation
	fmt.Println("\n📊 Demo 1: Enhanced Beverage Creation")
	fmt.Println("=====================================")

	enhancedRequest := &usecases.InventBeverageRequest{
		Ingredients:    []string{"espresso", "whole milk", "cinnamon", "honey"},
		Theme:          "comfort",
		UseAI:          true,
		CreatedBy:      "demo_user",
		ServingSize:    350,
		BatchSize:      1,
		AnalysisLevel:  "comprehensive",
		OptimizationGoals: &usecases.OptimizationGoals{
			PrioritizeNutrition: true,
			TargetMargin:        25.0,
			MaxCostPerServing:   4.00,
			HealthScore:         75.0,
		},
		DietaryProfile: &services.DietaryProfile{
			Allergies:           []string{},
			DietaryRestrictions: []string{"vegetarian"},
			HealthGoals:         []string{"energy", "wellness"},
			PreferredFlavors:    []string{"sweet", "warm spices"},
			MaxCalories:         func() *float64 { f := 200.0; return &f }(),
			MaxSugar:           func() *float64 { f := 15.0; return &f }(),
		},
	}

	response, err := beverageUseCase.InventBeverage(ctx, enhancedRequest)
	if err != nil {
		log.Fatalf("Failed to create enhanced beverage: %v", err)
	}

	printBeverageAnalysis(response)

	// Demo 2: Beverage Analysis
	fmt.Println("\n🔬 Demo 2: Comprehensive Beverage Analysis")
	fmt.Println("==========================================")

	if response.Beverage != nil {
		analysisResult, err := beverageUseCase.AnalyzeBeverage(ctx, response.Beverage.ID.String(), "comprehensive")
		if err != nil {
			log.Printf("Analysis failed: %v", err)
		} else {
			printDetailedAnalysis(analysisResult)
		}
	}

	// Demo 3: Recipe Optimization
	fmt.Println("\n⚡ Demo 3: Recipe Optimization")
	fmt.Println("==============================")

	if response.Beverage != nil {
		optimizationGoals := &usecases.OptimizationGoals{
			PrioritizeCost:    true,
			TargetMargin:      30.0,
			MaxCostPerServing: 3.50,
		}

		optimizationResult, err := beverageUseCase.OptimizeBeverage(ctx, response.Beverage.ID.String(), optimizationGoals)
		if err != nil {
			log.Printf("Optimization failed: %v", err)
		} else {
			printOptimizationResults(optimizationResult)
		}
	}

	// Demo 4: Individual Service Capabilities
	fmt.Println("\n🧪 Demo 4: Individual Service Capabilities")
	fmt.Println("==========================================")

	demonstrateIndividualServices(ctx, enhancedServices)

	fmt.Println("\n✅ Enhanced Beverage Inventor Demo Complete!")
	fmt.Println("The agent now provides comprehensive analysis, optimization, and intelligence.")
}

func printBeverageAnalysis(response *usecases.InventBeverageResponse) {
	fmt.Printf("🥤 Created Beverage: %s\n", response.Beverage.Name)
	fmt.Printf("📝 Description: %s\n", response.Beverage.Description)
	fmt.Printf("🎯 Theme: %s\n", response.Beverage.Theme)

	if response.HealthScore > 0 {
		fmt.Printf("🏥 Health Score: %.1f/100\n", response.HealthScore)
	}

	if response.EstimatedCost > 0 {
		fmt.Printf("💰 Estimated Cost: $%.2f per serving\n", response.EstimatedCost)
	}

	if response.ProfitMargin > 0 {
		fmt.Printf("📈 Target Profit Margin: %.1f%%\n", response.ProfitMargin)
	}

	if len(response.Recommendations) > 0 {
		fmt.Println("💡 Recommendations:")
		for _, rec := range response.Recommendations {
			fmt.Printf("  • %s\n", rec)
		}
	}

	if len(response.Warnings) > 0 {
		fmt.Println("⚠️ Warnings:")
		for _, warning := range response.Warnings {
			fmt.Printf("  • %s\n", warning)
		}
	}
}

func printDetailedAnalysis(result *usecases.InventBeverageResponse) {
	fmt.Println("📊 Detailed Analysis Results:")

	if result.NutritionalAnalysis != nil {
		fmt.Printf("🍎 Nutritional Score: %.1f/100\n", result.NutritionalAnalysis.Score)
		if result.NutritionalAnalysis.BasicNutrition != nil {
			nutrition := result.NutritionalAnalysis.BasicNutrition
			fmt.Printf("  • Calories: %.0f\n", nutrition.Calories)
			fmt.Printf("  • Protein: %.1fg\n", nutrition.Protein)
			fmt.Printf("  • Sugar: %.1fg\n", nutrition.Sugar)
			if nutrition.Caffeine > 0 {
				fmt.Printf("  • Caffeine: %.0fmg\n", nutrition.Caffeine)
			}
		}
	}

	if result.CostAnalysis != nil {
		fmt.Printf("💵 Cost Analysis:\n")
		fmt.Printf("  • Total Cost: $%.2f\n", result.CostAnalysis.TotalCost)
		fmt.Printf("  • Cost per Serving: $%.2f\n", result.CostAnalysis.CostPerServing)
		fmt.Printf("  • Confidence: %.1f%%\n", result.CostAnalysis.Confidence*100)
	}

	if result.CompatibilityAnalysis != nil {
		fmt.Printf("🔬 Compatibility Analysis:\n")
		fmt.Printf("  • Overall Compatibility: %.1f/100\n", result.CompatibilityAnalysis.OverallCompatibility)
		fmt.Printf("  • Conflicts: %d\n", len(result.CompatibilityAnalysis.Conflicts))
		fmt.Printf("  • Synergies: %d\n", len(result.CompatibilityAnalysis.Synergies))
	}
}

func printOptimizationResults(result *services.OptimizationResult) {
	fmt.Println("⚡ Optimization Results:")

	if result.BestRecipe != nil {
		fmt.Printf("🎯 Optimization Score: %.1f/100\n", result.BestRecipe.OptimizationScore)
		fmt.Printf("🔄 Iterations: %d\n", result.BestRecipe.Iterations)
		fmt.Printf("✅ Converged at: %d\n", result.BestRecipe.ConvergedAt)

		if result.BestRecipe.Objectives != nil {
			obj := result.BestRecipe.Objectives
			fmt.Printf("📊 Objective Scores:\n")
			fmt.Printf("  • Taste: %.1f/100\n", obj.TasteScore)
			fmt.Printf("  • Cost: %.1f/100\n", obj.CostScore)
			fmt.Printf("  • Nutrition: %.1f/100\n", obj.NutritionScore)
			fmt.Printf("  • Compatibility: %.1f/100\n", obj.CompatibilityScore)
		}
	}

	if len(result.Recommendations) > 0 {
		fmt.Println("💡 Optimization Recommendations:")
		for _, rec := range result.Recommendations {
			fmt.Printf("  • %s\n", rec)
		}
	}
}

func demonstrateIndividualServices(ctx context.Context, enhancedServices *factory.EnhancedServices) {
	// Create a sample beverage for testing
	sampleBeverage := &entities.Beverage{
		Name:        "Test Latte",
		Description: "A test latte for demonstration",
		Theme:       "coffee",
		Ingredients: []entities.Ingredient{
			{Name: "espresso", Quantity: 30, Unit: "ml"},
			{Name: "whole milk", Quantity: 200, Unit: "ml"},
			{Name: "sugar", Quantity: 5, Unit: "g"},
		},
	}

	// Nutritional Analysis Demo
	fmt.Println("🍎 Nutritional Analysis Service:")
	nutritionReq := &services.NutritionalAnalysisRequest{
		Beverage:      sampleBeverage,
		ServingSize:   250,
		AnalysisLevel: services.AnalysisLevelDetailed,
	}

	nutritionResult, err := enhancedServices.NutritionalAnalyzer.AnalyzeNutrition(ctx, nutritionReq)
	if err == nil && nutritionResult.BasicNutrition != nil {
		fmt.Printf("  • Calories: %.0f\n", nutritionResult.BasicNutrition.Calories)
		fmt.Printf("  • Health Score: %.1f/100\n", nutritionResult.Score)
	}

	// Cost Calculation Demo
	fmt.Println("💰 Cost Calculation Service:")
	costReq := &services.CostCalculationRequest{
		Beverage:    sampleBeverage,
		ServingSize: 250,
		BatchSize:   1,
	}

	costResult, err := enhancedServices.CostCalculator.CalculateCost(ctx, costReq)
	if err == nil {
		fmt.Printf("  • Cost per serving: $%.2f\n", costResult.CostPerServing)
		fmt.Printf("  • Confidence: %.1f%%\n", costResult.Confidence*100)
	}

	// Compatibility Analysis Demo
	fmt.Println("🔬 Compatibility Analysis Service:")
	compatReq := &services.CompatibilityAnalysisRequest{
		Ingredients:   sampleBeverage.Ingredients,
		BeverageType:  sampleBeverage.Theme,
		AnalysisLevel: services.CompatibilityAnalysisBasic,
	}

	compatResult, err := enhancedServices.CompatibilityAnalyzer.AnalyzeCompatibility(ctx, compatReq)
	if err == nil {
		fmt.Printf("  • Compatibility Score: %.1f/100\n", compatResult.OverallCompatibility)
		fmt.Printf("  • Conflicts: %d\n", len(compatResult.Conflicts))
		fmt.Printf("  • Synergies: %d\n", len(compatResult.Synergies))
	}

	// Recipe Optimization Demo
	fmt.Println("⚡ Recipe Optimization Service:")
	objectives := &services.OptimizationObjectives{
		TasteWeight:     0.4,
		CostWeight:      0.3,
		NutritionWeight: 0.3,
		MaxIterations:   10,
	}

	optReq := &services.OptimizationRequest{
		BaseRecipe:       sampleBeverage,
		Objectives:       objectives,
		ServingSize:      250,
		BatchSize:        1,
		OptimizationType: services.OptimizationTypeMultiObjective,
	}

	optResult, err := enhancedServices.RecipeOptimizer.OptimizeRecipe(ctx, optReq)
	if err == nil && optResult.BestRecipe != nil {
		fmt.Printf("  • Optimization Score: %.1f/100\n", optResult.BestRecipe.OptimizationScore)
		fmt.Printf("  • Confidence: %.1f%%\n", optResult.BestRecipe.Confidence*100)
	}
}
