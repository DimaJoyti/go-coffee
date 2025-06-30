package entities

import (
	"time"

	"github.com/google/uuid"
)

// Beverage represents a beverage recipe in the domain
type Beverage struct {
	ID          uuid.UUID        `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Ingredients []Ingredient     `json:"ingredients"`
	Theme       string           `json:"theme"`
	CreatedAt   time.Time        `json:"created_at"`
	CreatedBy   string           `json:"created_by"`
	Status      BeverageStatus   `json:"status"`
	Metadata    BeverageMetadata `json:"metadata"`
}

// Ingredient represents an ingredient used in a beverage
type Ingredient struct {
	Name        string          `json:"name"`
	Quantity    float64         `json:"quantity"`
	Unit        string          `json:"unit"`
	Source      string          `json:"source"`
	Cost        float64         `json:"cost"`
	Nutritional NutritionalInfo `json:"nutritional"`
}

// NutritionalInfo contains nutritional information for an ingredient
type NutritionalInfo struct {
	Calories  int      `json:"calories"`
	Protein   float64  `json:"protein"`
	Carbs     float64  `json:"carbs"`
	Fat       float64  `json:"fat"`
	Sugar     float64  `json:"sugar"`
	Caffeine  float64  `json:"caffeine"`
	Allergens []string `json:"allergens"`
}

// BeverageStatus represents the status of a beverage recipe
type BeverageStatus string

const (
	StatusDraft      BeverageStatus = "draft"
	StatusPending    BeverageStatus = "pending"
	StatusApproved   BeverageStatus = "approved"
	StatusRejected   BeverageStatus = "rejected"
	StatusTesting    BeverageStatus = "testing"
	StatusProduction BeverageStatus = "production"
)

// BeverageMetadata contains additional metadata for a beverage
type BeverageMetadata struct {
	EstimatedCost        float64  `json:"estimated_cost"`
	PreparationTime      int      `json:"preparation_time_minutes"`
	Difficulty           string   `json:"difficulty"`
	Tags                 []string `json:"tags"`
	TargetAudience       []string `json:"target_audience"`
	SeasonalAvailability []string `json:"seasonal_availability"`
}

// NewBeverage creates a new beverage with default values
func NewBeverage(name, description, theme string, ingredients []Ingredient) *Beverage {
	return &Beverage{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Ingredients: ingredients,
		Theme:       theme,
		CreatedAt:   time.Now(),
		Status:      StatusDraft,
		Metadata:    BeverageMetadata{},
	}
}

// IsValid validates the beverage entity
func (b *Beverage) IsValid() bool {
	if b.Name == "" || b.Description == "" {
		return false
	}
	if len(b.Ingredients) == 0 {
		return false
	}
	return true
}

// CalculateTotalCost calculates the total cost of all ingredients
func (b *Beverage) CalculateTotalCost() float64 {
	total := 0.0
	for _, ingredient := range b.Ingredients {
		total += ingredient.Cost * ingredient.Quantity
	}
	b.Metadata.EstimatedCost = total
	return total
}

// GetTotalCalories calculates total calories for the beverage
func (b *Beverage) GetTotalCalories() int {
	total := 0
	for _, ingredient := range b.Ingredients {
		total += int(float64(ingredient.Nutritional.Calories) * ingredient.Quantity)
	}
	return total
}

// GetAllAllergens returns all unique allergens in the beverage
func (b *Beverage) GetAllAllergens() []string {
	allergenMap := make(map[string]bool)
	for _, ingredient := range b.Ingredients {
		for _, allergen := range ingredient.Nutritional.Allergens {
			allergenMap[allergen] = true
		}
	}

	allergens := make([]string, 0, len(allergenMap))
	for allergen := range allergenMap {
		allergens = append(allergens, allergen)
	}
	return allergens
}

// UpdateStatus updates the beverage status
func (b *Beverage) UpdateStatus(status BeverageStatus) {
	b.Status = status
}

// ParseUUID parses a string into a UUID
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
