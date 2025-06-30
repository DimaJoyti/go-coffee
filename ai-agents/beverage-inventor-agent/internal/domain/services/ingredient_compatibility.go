package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"go-coffee-ai-agents/beverage-inventor-agent/internal/domain/entities"
)

// IngredientCompatibilityAnalyzer analyzes ingredient compatibility and suggests substitutions
type IngredientCompatibilityAnalyzer struct {
	knowledgeBase IngredientKnowledgeBase
	aiProvider    CompatibilityAIProvider
	flavorDB      FlavorDatabase
}

// IngredientKnowledgeBase defines the interface for ingredient knowledge
type IngredientKnowledgeBase interface {
	GetIngredientProfile(ctx context.Context, ingredient string) (*IngredientProfile, error)
	GetCompatibilityRules(ctx context.Context) ([]*CompatibilityRule, error)
	GetSubstitutions(ctx context.Context, ingredient string) ([]*Substitution, error)
	GetFlavorProfile(ctx context.Context, ingredient string) (*FlavorProfile, error)
	GetChemicalProperties(ctx context.Context, ingredient string) (*ChemicalProperties, error)
}

// CompatibilityAIProvider defines AI capabilities for compatibility analysis
type CompatibilityAIProvider interface {
	AnalyzeFlavorHarmony(ctx context.Context, ingredients []entities.Ingredient) (*FlavorHarmonyAnalysis, error)
	PredictTasteProfile(ctx context.Context, ingredients []entities.Ingredient) (*TasteProfile, error)
	SuggestFlavorEnhancements(ctx context.Context, currentProfile *TasteProfile, targetProfile *TasteProfile) ([]string, error)
	AnalyzeTextureInteractions(ctx context.Context, ingredients []entities.Ingredient) (*TextureAnalysis, error)
}

// FlavorDatabase defines the interface for flavor data
type FlavorDatabase interface {
	GetFlavorCompounds(ctx context.Context, ingredient string) ([]string, error)
	GetFlavorIntensity(ctx context.Context, ingredient string) (float64, error)
	GetComplementaryFlavors(ctx context.Context, ingredient string) ([]string, error)
	GetConflictingFlavors(ctx context.Context, ingredient string) ([]string, error)
}

// IngredientProfile represents comprehensive ingredient information
type IngredientProfile struct {
	Name            string              `json:"name"`
	Category        string              `json:"category"`        // dairy, fruit, spice, etc.
	FlavorProfile   *FlavorProfile      `json:"flavor_profile"`
	ChemicalProps   *ChemicalProperties `json:"chemical_properties"`
	NutritionalInfo *entities.NutritionalInfo `json:"nutritional_info"`
	Allergens       []string            `json:"allergens"`
	Substitutions   []*Substitution     `json:"substitutions"`
	StorageReqs     *StorageRequirements `json:"storage_requirements"`
	SeasonalInfo    *SeasonalInfo       `json:"seasonal_info"`
}

// FlavorProfile represents the flavor characteristics of an ingredient
type FlavorProfile struct {
	Primary         []string  `json:"primary"`         // main flavor notes
	Secondary       []string  `json:"secondary"`       // secondary flavor notes
	Intensity       float64   `json:"intensity"`       // 0-10 scale
	Sweetness       float64   `json:"sweetness"`       // 0-10 scale
	Acidity         float64   `json:"acidity"`         // 0-10 scale
	Bitterness      float64   `json:"bitterness"`      // 0-10 scale
	Saltiness       float64   `json:"saltiness"`       // 0-10 scale
	Umami           float64   `json:"umami"`           // 0-10 scale
	Astringency     float64   `json:"astringency"`     // 0-10 scale
	Aromatics       []string  `json:"aromatics"`       // aromatic compounds
	Mouthfeel       string    `json:"mouthfeel"`       // creamy, light, thick, etc.
	AfterTaste      string    `json:"after_taste"`     // lingering flavors
}

// ChemicalProperties represents chemical characteristics
type ChemicalProperties struct {
	PH              float64   `json:"ph"`
	Solubility      string    `json:"solubility"`      // water, oil, alcohol
	Stability       string    `json:"stability"`       // heat, light, oxygen
	ReactiveGroups  []string  `json:"reactive_groups"` // compounds that may react
	Preservatives   []string  `json:"preservatives"`   // natural preservatives
	Emulsifiers     []string  `json:"emulsifiers"`     // emulsifying properties
	Antioxidants    []string  `json:"antioxidants"`    // antioxidant compounds
}

// Substitution represents a possible ingredient substitution
type Substitution struct {
	Original        string    `json:"original"`
	Substitute      string    `json:"substitute"`
	Ratio           float64   `json:"ratio"`           // substitution ratio
	FlavorImpact    string    `json:"flavor_impact"`   // how it affects flavor
	TextureImpact   string    `json:"texture_impact"`  // how it affects texture
	NutritionalImpact string  `json:"nutritional_impact"` // nutritional changes
	Confidence      float64   `json:"confidence"`      // 0-1 confidence score
	Notes           string    `json:"notes"`
}

// StorageRequirements represents storage requirements
type StorageRequirements struct {
	Temperature     string    `json:"temperature"`     // room, cold, frozen
	Humidity        string    `json:"humidity"`        // low, medium, high
	Light           string    `json:"light"`           // dark, normal, bright
	ShelfLife       string    `json:"shelf_life"`      // duration
	Container       string    `json:"container"`       // glass, plastic, metal
}

// SeasonalInfo represents seasonal availability and quality
type SeasonalInfo struct {
	PeakSeason      []string  `json:"peak_season"`     // months
	OffSeason       []string  `json:"off_season"`      // months
	QualityVariation string   `json:"quality_variation"` // seasonal quality changes
	PriceVariation  string    `json:"price_variation"`   // seasonal price changes
}

// CompatibilityRule represents a compatibility rule
type CompatibilityRule struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"`            // positive, negative, neutral
	Ingredient1     string    `json:"ingredient1"`
	Ingredient2     string    `json:"ingredient2"`
	Category1       string    `json:"category1"`       // can be category instead of specific ingredient
	Category2       string    `json:"category2"`
	Reason          string    `json:"reason"`
	Confidence      float64   `json:"confidence"`
	Source          string    `json:"source"`          // culinary tradition, science, etc.
}

// FlavorHarmonyAnalysis represents AI analysis of flavor harmony
type FlavorHarmonyAnalysis struct {
	OverallHarmony  float64            `json:"overall_harmony"`  // 0-100 score
	FlavorBalance   *FlavorBalance     `json:"flavor_balance"`
	Conflicts       []FlavorConflict   `json:"conflicts"`
	Synergies       []FlavorSynergy    `json:"synergies"`
	Recommendations []string           `json:"recommendations"`
	Confidence      float64            `json:"confidence"`
}

// FlavorBalance represents the balance of flavor elements
type FlavorBalance struct {
	Sweet           float64   `json:"sweet"`
	Sour            float64   `json:"sour"`
	Bitter          float64   `json:"bitter"`
	Salty           float64   `json:"salty"`
	Umami           float64   `json:"umami"`
	Overall         string    `json:"overall"`         // balanced, sweet-heavy, etc.
	Recommendations []string  `json:"recommendations"`
}

// FlavorConflict represents a flavor conflict
type FlavorConflict struct {
	Ingredient1     string    `json:"ingredient1"`
	Ingredient2     string    `json:"ingredient2"`
	ConflictType    string    `json:"conflict_type"`   // chemical, flavor, texture
	Severity        string    `json:"severity"`        // low, medium, high
	Description     string    `json:"description"`
	Mitigation      []string  `json:"mitigation"`      // ways to resolve conflict
}

// FlavorSynergy represents a positive flavor interaction
type FlavorSynergy struct {
	Ingredients     []string  `json:"ingredients"`
	SynergyType     string    `json:"synergy_type"`    // complementary, enhancing, masking
	Effect          string    `json:"effect"`          // what the synergy creates
	Strength        float64   `json:"strength"`        // 0-10 strength of synergy
	Description     string    `json:"description"`
}

// TasteProfile represents the overall taste profile
type TasteProfile struct {
	DominantFlavors []string  `json:"dominant_flavors"`
	FlavorNotes     []string  `json:"flavor_notes"`
	Intensity       float64   `json:"intensity"`
	Complexity      float64   `json:"complexity"`
	Balance         float64   `json:"balance"`
	Uniqueness      float64   `json:"uniqueness"`
	Appeal          float64   `json:"appeal"`
	Description     string    `json:"description"`
}

// TextureAnalysis represents texture interaction analysis
type TextureAnalysis struct {
	OverallTexture  string            `json:"overall_texture"`
	Consistency     string            `json:"consistency"`
	Mouthfeel       string            `json:"mouthfeel"`
	Interactions    []TextureInteraction `json:"interactions"`
	Issues          []string          `json:"issues"`
	Recommendations []string          `json:"recommendations"`
}

// TextureInteraction represents how ingredients affect texture
type TextureInteraction struct {
	Ingredients     []string  `json:"ingredients"`
	Effect          string    `json:"effect"`
	Mechanism       string    `json:"mechanism"`
	Impact          string    `json:"impact"`          // positive, negative, neutral
}

// CompatibilityAnalysisRequest represents a request for compatibility analysis
type CompatibilityAnalysisRequest struct {
	Ingredients     []entities.Ingredient `json:"ingredients"`
	BeverageType    string               `json:"beverage_type"`
	TargetProfile   *TasteProfile        `json:"target_profile,omitempty"`
	Constraints     *CompatibilityConstraints `json:"constraints,omitempty"`
	AnalysisLevel   CompatibilityAnalysisLevel `json:"analysis_level"`
}

// CompatibilityConstraints represents constraints for compatibility analysis
type CompatibilityConstraints struct {
	Allergens       []string  `json:"allergens"`       // allergens to avoid
	DietaryRestrictions []string `json:"dietary_restrictions"`
	FlavorPreferences []string `json:"flavor_preferences"`
	TexturePreferences []string `json:"texture_preferences"`
	MaxIngredients  int       `json:"max_ingredients"`
	Budget          float64   `json:"budget"`
}

// CompatibilityAnalysisLevel defines the depth of analysis
type CompatibilityAnalysisLevel string

const (
	CompatibilityAnalysisBasic       CompatibilityAnalysisLevel = "basic"
	CompatibilityAnalysisDetailed    CompatibilityAnalysisLevel = "detailed"
	CompatibilityAnalysisComprehensive CompatibilityAnalysisLevel = "comprehensive"
)

// CompatibilityAnalysisResult represents the result of compatibility analysis
type CompatibilityAnalysisResult struct {
	OverallCompatibility float64                `json:"overall_compatibility"` // 0-100 score
	FlavorHarmony       *FlavorHarmonyAnalysis `json:"flavor_harmony,omitempty"`
	TasteProfile        *TasteProfile          `json:"taste_profile,omitempty"`
	TextureAnalysis     *TextureAnalysis       `json:"texture_analysis,omitempty"`
	Conflicts           []FlavorConflict       `json:"conflicts"`
	Synergies           []FlavorSynergy        `json:"synergies"`
	Substitutions       []*SubstitutionSuggestion `json:"substitutions"`
	Enhancements        []string               `json:"enhancements"`
	Warnings            []string               `json:"warnings"`
	Recommendations     []string               `json:"recommendations"`
	Confidence          float64                `json:"confidence"`
}

// SubstitutionSuggestion represents a suggested substitution
type SubstitutionSuggestion struct {
	Original        string    `json:"original"`
	Suggestions     []*Substitution `json:"suggestions"`
	Reason          string    `json:"reason"`
	Priority        string    `json:"priority"`        // high, medium, low
}

// NewIngredientCompatibilityAnalyzer creates a new compatibility analyzer
func NewIngredientCompatibilityAnalyzer(kb IngredientKnowledgeBase, ai CompatibilityAIProvider, flavorDB FlavorDatabase) *IngredientCompatibilityAnalyzer {
	return &IngredientCompatibilityAnalyzer{
		knowledgeBase: kb,
		aiProvider:    ai,
		flavorDB:      flavorDB,
	}
}

// AnalyzeCompatibility performs comprehensive ingredient compatibility analysis
func (ica *IngredientCompatibilityAnalyzer) AnalyzeCompatibility(ctx context.Context, req *CompatibilityAnalysisRequest) (*CompatibilityAnalysisResult, error) {
	result := &CompatibilityAnalysisResult{
		Conflicts:       []FlavorConflict{},
		Synergies:       []FlavorSynergy{},
		Substitutions:   []*SubstitutionSuggestion{},
		Enhancements:    []string{},
		Warnings:        []string{},
		Recommendations: []string{},
	}
	
	// Basic compatibility check
	basicScore, conflicts, synergies := ica.performBasicCompatibilityCheck(ctx, req.Ingredients)
	result.OverallCompatibility = basicScore
	result.Conflicts = conflicts
	result.Synergies = synergies
	
	// Detailed analysis if requested
	if req.AnalysisLevel == CompatibilityAnalysisDetailed || req.AnalysisLevel == CompatibilityAnalysisComprehensive {
		// AI-powered flavor harmony analysis
		flavorHarmony, err := ica.aiProvider.AnalyzeFlavorHarmony(ctx, req.Ingredients)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Flavor harmony analysis failed: %v", err))
		} else {
			result.FlavorHarmony = flavorHarmony
		}
		
		// Taste profile prediction
		tasteProfile, err := ica.aiProvider.PredictTasteProfile(ctx, req.Ingredients)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Taste profile prediction failed: %v", err))
		} else {
			result.TasteProfile = tasteProfile
		}
		
		// Texture analysis
		textureAnalysis, err := ica.aiProvider.AnalyzeTextureInteractions(ctx, req.Ingredients)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Texture analysis failed: %v", err))
		} else {
			result.TextureAnalysis = textureAnalysis
		}
	}
	
	// Generate substitution suggestions
	substitutions, err := ica.generateSubstitutionSuggestions(ctx, req.Ingredients, req.Constraints)
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Substitution generation failed: %v", err))
	} else {
		result.Substitutions = substitutions
	}
	
	// Generate enhancement suggestions
	if req.TargetProfile != nil && result.TasteProfile != nil {
		enhancements, err := ica.aiProvider.SuggestFlavorEnhancements(ctx, result.TasteProfile, req.TargetProfile)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Enhancement suggestion failed: %v", err))
		} else {
			result.Enhancements = enhancements
		}
	}
	
	// Generate recommendations
	result.Recommendations = ica.generateRecommendations(result)
	
	// Calculate confidence
	result.Confidence = ica.calculateConfidence(result)
	
	return result, nil
}

// performBasicCompatibilityCheck performs basic rule-based compatibility checking
func (ica *IngredientCompatibilityAnalyzer) performBasicCompatibilityCheck(ctx context.Context, ingredients []entities.Ingredient) (float64, []FlavorConflict, []FlavorSynergy) {
	conflicts := []FlavorConflict{}
	synergies := []FlavorSynergy{}
	
	// Get compatibility rules
	rules, err := ica.knowledgeBase.GetCompatibilityRules(ctx)
	if err != nil {
		return 50.0, conflicts, synergies // Default score if rules unavailable
	}
	
	totalScore := 0.0
	totalChecks := 0
	
	// Check each pair of ingredients
	for i, ing1 := range ingredients {
		for j, ing2 := range ingredients {
			if i >= j {
				continue // Avoid duplicate checks
			}
			
			totalChecks++
			pairScore := 50.0 // Neutral score
			
			// Check rules for this pair
			for _, rule := range rules {
				if ica.ruleApplies(rule, ing1.Name, ing2.Name) {
					switch rule.Type {
					case "positive":
						pairScore += 20 * rule.Confidence
						synergies = append(synergies, FlavorSynergy{
							Ingredients: []string{ing1.Name, ing2.Name},
							SynergyType: "complementary",
							Effect:      rule.Reason,
							Strength:    rule.Confidence * 10,
							Description: rule.Reason,
						})
					case "negative":
						pairScore -= 30 * rule.Confidence
						conflicts = append(conflicts, FlavorConflict{
							Ingredient1:  ing1.Name,
							Ingredient2:  ing2.Name,
							ConflictType: "flavor",
							Severity:     ica.getSeverity(rule.Confidence),
							Description:  rule.Reason,
							Mitigation:   []string{"Consider reducing quantities", "Add balancing ingredients"},
						})
					}
				}
			}
			
			// Ensure score is within bounds
			if pairScore < 0 {
				pairScore = 0
			}
			if pairScore > 100 {
				pairScore = 100
			}
			
			totalScore += pairScore
		}
	}
	
	// Calculate average score
	averageScore := 50.0 // Default if no checks
	if totalChecks > 0 {
		averageScore = totalScore / float64(totalChecks)
	}
	
	return math.Round(averageScore*10) / 10, conflicts, synergies
}

// generateSubstitutionSuggestions generates substitution suggestions
func (ica *IngredientCompatibilityAnalyzer) generateSubstitutionSuggestions(ctx context.Context, ingredients []entities.Ingredient, constraints *CompatibilityConstraints) ([]*SubstitutionSuggestion, error) {
	suggestions := []*SubstitutionSuggestion{}
	
	for _, ingredient := range ingredients {
		// Get possible substitutions
		substitutions, err := ica.knowledgeBase.GetSubstitutions(ctx, ingredient.Name)
		if err != nil {
			continue
		}
		
		// Filter substitutions based on constraints
		if constraints != nil {
			substitutions = ica.filterSubstitutions(substitutions, constraints)
		}
		
		if len(substitutions) > 0 {
			// Sort by confidence
			sort.Slice(substitutions, func(i, j int) bool {
				return substitutions[i].Confidence > substitutions[j].Confidence
			})
			
			// Take top 3 suggestions
			maxSuggestions := 3
			if len(substitutions) < maxSuggestions {
				maxSuggestions = len(substitutions)
			}
			
			suggestion := &SubstitutionSuggestion{
				Original:    ingredient.Name,
				Suggestions: substitutions[:maxSuggestions],
				Reason:      ica.getSubstitutionReason(ingredient.Name, substitutions),
				Priority:    ica.getSubstitutionPriority(substitutions[0].Confidence),
			}
			
			suggestions = append(suggestions, suggestion)
		}
	}
	
	return suggestions, nil
}

// generateRecommendations generates overall recommendations
func (ica *IngredientCompatibilityAnalyzer) generateRecommendations(result *CompatibilityAnalysisResult) []string {
	recommendations := []string{}
	
	// Compatibility-based recommendations
	if result.OverallCompatibility < 60 {
		recommendations = append(recommendations, "Consider ingredient substitutions to improve compatibility")
	} else if result.OverallCompatibility > 80 {
		recommendations = append(recommendations, "Excellent ingredient compatibility - recipe should work well")
	}
	
	// Conflict-based recommendations
	if len(result.Conflicts) > 0 {
		highSeverityConflicts := 0
		for _, conflict := range result.Conflicts {
			if conflict.Severity == "high" {
				highSeverityConflicts++
			}
		}
		
		if highSeverityConflicts > 0 {
			recommendations = append(recommendations, fmt.Sprintf("Address %d high-severity flavor conflicts", highSeverityConflicts))
		}
	}
	
	// Synergy-based recommendations
	if len(result.Synergies) > 2 {
		recommendations = append(recommendations, "Great flavor synergies detected - consider highlighting these combinations")
	}
	
	// Texture-based recommendations
	if result.TextureAnalysis != nil && len(result.TextureAnalysis.Issues) > 0 {
		recommendations = append(recommendations, "Review texture interactions for potential improvements")
	}
	
	return recommendations
}

// Helper methods
func (ica *IngredientCompatibilityAnalyzer) ruleApplies(rule *CompatibilityRule, ing1, ing2 string) bool {
	// Check direct ingredient matches
	if (rule.Ingredient1 == ing1 && rule.Ingredient2 == ing2) ||
		(rule.Ingredient1 == ing2 && rule.Ingredient2 == ing1) {
		return true
	}
	
	// TODO: Check category matches
	// This would require getting ingredient categories and matching against rule categories
	
	return false
}

func (ica *IngredientCompatibilityAnalyzer) getSeverity(confidence float64) string {
	if confidence > 0.8 {
		return "high"
	} else if confidence > 0.5 {
		return "medium"
	}
	return "low"
}

func (ica *IngredientCompatibilityAnalyzer) filterSubstitutions(substitutions []*Substitution, constraints *CompatibilityConstraints) []*Substitution {
	filtered := []*Substitution{}
	
	for _, sub := range substitutions {
		// Check allergen constraints
		hasAllergen := false
		for _, allergen := range constraints.Allergens {
			if strings.Contains(strings.ToLower(sub.Substitute), strings.ToLower(allergen)) {
				hasAllergen = true
				break
			}
		}
		
		if !hasAllergen {
			filtered = append(filtered, sub)
		}
	}
	
	return filtered
}

func (ica *IngredientCompatibilityAnalyzer) getSubstitutionReason(original string, substitutions []*Substitution) string {
	if len(substitutions) == 0 {
		return "No specific reason"
	}
	
	// Use the reason from the highest confidence substitution
	return substitutions[0].Notes
}

func (ica *IngredientCompatibilityAnalyzer) getSubstitutionPriority(confidence float64) string {
	if confidence > 0.8 {
		return "high"
	} else if confidence > 0.6 {
		return "medium"
	}
	return "low"
}

func (ica *IngredientCompatibilityAnalyzer) calculateConfidence(result *CompatibilityAnalysisResult) float64 {
	confidence := 0.8 // Base confidence
	
	// Reduce confidence based on warnings
	confidence -= float64(len(result.Warnings)) * 0.1
	
	// Increase confidence if we have detailed analysis
	if result.FlavorHarmony != nil {
		confidence += 0.1
	}
	if result.TasteProfile != nil {
		confidence += 0.1
	}
	if result.TextureAnalysis != nil {
		confidence += 0.1
	}
	
	// Ensure confidence is within bounds
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}
	
	return math.Round(confidence*100) / 100
}
