package cost

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go-coffee-ai-agents/internal/ai/providers"
)

// Tracker tracks AI usage costs across providers
type Tracker struct {
	usage      map[string]*ProviderUsage
	budgets    map[string]*Budget
	alerts     []AlertRule
	mutex      sync.RWMutex
	
	// Callbacks
	onBudgetExceeded func(provider string, budget *Budget, usage *ProviderUsage)
	onAlertTriggered func(alert AlertRule, usage *ProviderUsage)
}

// ProviderUsage tracks usage for a specific provider
type ProviderUsage struct {
	Provider         string            `json:"provider"`
	TotalRequests    int64             `json:"total_requests"`
	TotalTokens      int64             `json:"total_tokens"`
	InputTokens      int64             `json:"input_tokens"`
	OutputTokens     int64             `json:"output_tokens"`
	TotalCost        float64           `json:"total_cost"`
	
	// Time-based usage
	DailyUsage       map[string]*DailyUsage `json:"daily_usage"`
	MonthlyUsage     map[string]*MonthlyUsage `json:"monthly_usage"`
	
	// Model-specific usage
	ModelUsage       map[string]*ModelUsage `json:"model_usage"`
	
	// User-specific usage
	UserUsage        map[string]*UserUsage `json:"user_usage"`
	
	// Last updated
	LastUpdated      time.Time         `json:"last_updated"`
}

// DailyUsage tracks daily usage
type DailyUsage struct {
	Date         string    `json:"date"`
	Requests     int64     `json:"requests"`
	Tokens       int64     `json:"tokens"`
	Cost         float64   `json:"cost"`
	LastUpdated  time.Time `json:"last_updated"`
}

// MonthlyUsage tracks monthly usage
type MonthlyUsage struct {
	Month        string    `json:"month"`
	Requests     int64     `json:"requests"`
	Tokens       int64     `json:"tokens"`
	Cost         float64   `json:"cost"`
	LastUpdated  time.Time `json:"last_updated"`
}

// ModelUsage tracks usage per model
type ModelUsage struct {
	Model        string    `json:"model"`
	Requests     int64     `json:"requests"`
	InputTokens  int64     `json:"input_tokens"`
	OutputTokens int64     `json:"output_tokens"`
	TotalTokens  int64     `json:"total_tokens"`
	Cost         float64   `json:"cost"`
	LastUsed     time.Time `json:"last_used"`
}

// UserUsage tracks usage per user
type UserUsage struct {
	UserID       string    `json:"user_id"`
	Requests     int64     `json:"requests"`
	Tokens       int64     `json:"tokens"`
	Cost         float64   `json:"cost"`
	LastUsed     time.Time `json:"last_used"`
}

// Budget represents a cost budget
type Budget struct {
	Provider     string    `json:"provider"`
	Type         BudgetType `json:"type"`
	Limit        float64   `json:"limit"`
	Period       BudgetPeriod `json:"period"`
	Currency     string    `json:"currency"`
	
	// Current usage
	CurrentUsage float64   `json:"current_usage"`
	Remaining    float64   `json:"remaining"`
	
	// Alert thresholds
	AlertThresholds []float64 `json:"alert_thresholds"`
	
	// Status
	Exceeded     bool      `json:"exceeded"`
	LastReset    time.Time `json:"last_reset"`
	NextReset    time.Time `json:"next_reset"`
}

// BudgetType represents the type of budget
type BudgetType string

const (
	BudgetTypeCost     BudgetType = "cost"
	BudgetTypeTokens   BudgetType = "tokens"
	BudgetTypeRequests BudgetType = "requests"
)

// BudgetPeriod represents the budget period
type BudgetPeriod string

const (
	BudgetPeriodDaily   BudgetPeriod = "daily"
	BudgetPeriodWeekly  BudgetPeriod = "weekly"
	BudgetPeriodMonthly BudgetPeriod = "monthly"
	BudgetPeriodYearly  BudgetPeriod = "yearly"
)

// AlertRule represents an alert rule
type AlertRule struct {
	ID           string    `json:"id"`
	Provider     string    `json:"provider"`
	Type         AlertType `json:"type"`
	Threshold    float64   `json:"threshold"`
	Period       BudgetPeriod `json:"period"`
	Enabled      bool      `json:"enabled"`
	
	// Alert configuration
	Message      string    `json:"message"`
	Channels     []string  `json:"channels"`
	
	// State
	LastTriggered time.Time `json:"last_triggered"`
	TriggerCount  int64     `json:"trigger_count"`
}

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeCostThreshold     AlertType = "cost_threshold"
	AlertTypeTokenThreshold    AlertType = "token_threshold"
	AlertTypeRequestThreshold  AlertType = "request_threshold"
	AlertTypeBudgetExceeded    AlertType = "budget_exceeded"
	AlertTypeUnusualUsage      AlertType = "unusual_usage"
)

// NewTracker creates a new cost tracker
func NewTracker() *Tracker {
	return &Tracker{
		usage:   make(map[string]*ProviderUsage),
		budgets: make(map[string]*Budget),
		alerts:  []AlertRule{},
	}
}

// TrackUsage tracks usage for a provider
func (t *Tracker) TrackUsage(ctx context.Context, provider string, req *providers.GenerateRequest, resp *providers.GenerateResponse) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	// Get or create provider usage
	usage, exists := t.usage[provider]
	if !exists {
		usage = &ProviderUsage{
			Provider:     provider,
			DailyUsage:   make(map[string]*DailyUsage),
			MonthlyUsage: make(map[string]*MonthlyUsage),
			ModelUsage:   make(map[string]*ModelUsage),
			UserUsage:    make(map[string]*UserUsage),
		}
		t.usage[provider] = usage
	}
	
	// Update total usage
	usage.TotalRequests++
	if resp.Usage != nil {
		usage.TotalTokens += int64(resp.Usage.TotalTokens)
		usage.InputTokens += int64(resp.Usage.PromptTokens)
		usage.OutputTokens += int64(resp.Usage.CompletionTokens)
	}
	
	if resp.Cost != nil {
		usage.TotalCost += resp.Cost.TotalCost
	}
	
	// Update daily usage
	today := time.Now().Format("2006-01-02")
	dailyUsage, exists := usage.DailyUsage[today]
	if !exists {
		dailyUsage = &DailyUsage{Date: today}
		usage.DailyUsage[today] = dailyUsage
	}
	dailyUsage.Requests++
	if resp.Usage != nil {
		dailyUsage.Tokens += int64(resp.Usage.TotalTokens)
	}
	if resp.Cost != nil {
		dailyUsage.Cost += resp.Cost.TotalCost
	}
	dailyUsage.LastUpdated = time.Now()
	
	// Update monthly usage
	month := time.Now().Format("2006-01")
	monthlyUsage, exists := usage.MonthlyUsage[month]
	if !exists {
		monthlyUsage = &MonthlyUsage{Month: month}
		usage.MonthlyUsage[month] = monthlyUsage
	}
	monthlyUsage.Requests++
	if resp.Usage != nil {
		monthlyUsage.Tokens += int64(resp.Usage.TotalTokens)
	}
	if resp.Cost != nil {
		monthlyUsage.Cost += resp.Cost.TotalCost
	}
	monthlyUsage.LastUpdated = time.Now()
	
	// Update model usage
	model := req.Model
	if model == "" {
		model = "unknown"
	}
	modelUsage, exists := usage.ModelUsage[model]
	if !exists {
		modelUsage = &ModelUsage{Model: model}
		usage.ModelUsage[model] = modelUsage
	}
	modelUsage.Requests++
	if resp.Usage != nil {
		modelUsage.InputTokens += int64(resp.Usage.PromptTokens)
		modelUsage.OutputTokens += int64(resp.Usage.CompletionTokens)
		modelUsage.TotalTokens += int64(resp.Usage.TotalTokens)
	}
	if resp.Cost != nil {
		modelUsage.Cost += resp.Cost.TotalCost
	}
	modelUsage.LastUsed = time.Now()
	
	// Update user usage
	userID := req.UserID
	if userID == "" {
		userID = "anonymous"
	}
	userUsage, exists := usage.UserUsage[userID]
	if !exists {
		userUsage = &UserUsage{UserID: userID}
		usage.UserUsage[userID] = userUsage
	}
	userUsage.Requests++
	if resp.Usage != nil {
		userUsage.Tokens += int64(resp.Usage.TotalTokens)
	}
	if resp.Cost != nil {
		userUsage.Cost += resp.Cost.TotalCost
	}
	userUsage.LastUsed = time.Now()
	
	usage.LastUpdated = time.Now()
	
	// Check budgets and alerts
	t.checkBudgets(provider, usage)
	t.checkAlerts(provider, usage)
	
	return nil
}

// SetBudget sets a budget for a provider
func (t *Tracker) SetBudget(budget *Budget) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	key := fmt.Sprintf("%s_%s_%s", budget.Provider, budget.Type, budget.Period)
	t.budgets[key] = budget
}

// GetBudget gets a budget for a provider
func (t *Tracker) GetBudget(provider string, budgetType BudgetType, period BudgetPeriod) *Budget {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	
	key := fmt.Sprintf("%s_%s_%s", provider, budgetType, period)
	return t.budgets[key]
}

// AddAlertRule adds an alert rule
func (t *Tracker) AddAlertRule(rule AlertRule) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	t.alerts = append(t.alerts, rule)
}

// GetUsage gets usage for a provider
func (t *Tracker) GetUsage(provider string) *ProviderUsage {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	
	return t.usage[provider]
}

// GetAllUsage gets usage for all providers
func (t *Tracker) GetAllUsage() map[string]*ProviderUsage {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	
	// Return a copy to prevent external modifications
	result := make(map[string]*ProviderUsage)
	for k, v := range t.usage {
		result[k] = v
	}
	
	return result
}

// GetTotalCost gets total cost across all providers
func (t *Tracker) GetTotalCost() float64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	
	total := 0.0
	for _, usage := range t.usage {
		total += usage.TotalCost
	}
	
	return total
}

// GetDailyCost gets daily cost for a specific date
func (t *Tracker) GetDailyCost(provider, date string) float64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	
	usage, exists := t.usage[provider]
	if !exists {
		return 0.0
	}
	
	dailyUsage, exists := usage.DailyUsage[date]
	if !exists {
		return 0.0
	}
	
	return dailyUsage.Cost
}

// GetMonthlyCost gets monthly cost for a specific month
func (t *Tracker) GetMonthlyCost(provider, month string) float64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	
	usage, exists := t.usage[provider]
	if !exists {
		return 0.0
	}
	
	monthlyUsage, exists := usage.MonthlyUsage[month]
	if !exists {
		return 0.0
	}
	
	return monthlyUsage.Cost
}

// checkBudgets checks if any budgets are exceeded
func (t *Tracker) checkBudgets(provider string, usage *ProviderUsage) {
	for _, budget := range t.budgets {
		if budget.Provider != provider && budget.Provider != "*" {
			continue
		}
		
		// Check if budget period needs reset
		if time.Now().After(budget.NextReset) {
			t.resetBudget(budget)
		}
		
		// Get current usage for budget period
		currentUsage := t.getCurrentUsageForBudget(budget, usage)
		budget.CurrentUsage = currentUsage
		budget.Remaining = budget.Limit - currentUsage
		
		// Check if budget is exceeded
		if currentUsage > budget.Limit && !budget.Exceeded {
			budget.Exceeded = true
			if t.onBudgetExceeded != nil {
				t.onBudgetExceeded(provider, budget, usage)
			}
		}
		
		// Check alert thresholds
		for _, threshold := range budget.AlertThresholds {
			if currentUsage >= budget.Limit*threshold/100 {
				// Trigger alert
				if t.onAlertTriggered != nil {
					alert := AlertRule{
						Type:      AlertTypeBudgetExceeded,
						Threshold: threshold,
						Message:   fmt.Sprintf("Budget threshold %.1f%% exceeded for %s", threshold, provider),
					}
					t.onAlertTriggered(alert, usage)
				}
			}
		}
	}
}

// checkAlerts checks if any alert rules are triggered
func (t *Tracker) checkAlerts(provider string, usage *ProviderUsage) {
	for _, alert := range t.alerts {
		if !alert.Enabled || (alert.Provider != provider && alert.Provider != "*") {
			continue
		}
		
		triggered := false
		
		switch alert.Type {
		case AlertTypeCostThreshold:
			if usage.TotalCost >= alert.Threshold {
				triggered = true
			}
		case AlertTypeTokenThreshold:
			if float64(usage.TotalTokens) >= alert.Threshold {
				triggered = true
			}
		case AlertTypeRequestThreshold:
			if float64(usage.TotalRequests) >= alert.Threshold {
				triggered = true
			}
		}
		
		if triggered && t.onAlertTriggered != nil {
			t.onAlertTriggered(alert, usage)
		}
	}
}

// resetBudget resets a budget for a new period
func (t *Tracker) resetBudget(budget *Budget) {
	budget.CurrentUsage = 0
	budget.Remaining = budget.Limit
	budget.Exceeded = false
	budget.LastReset = time.Now()
	
	// Calculate next reset time
	switch budget.Period {
	case BudgetPeriodDaily:
		budget.NextReset = time.Now().AddDate(0, 0, 1)
	case BudgetPeriodWeekly:
		budget.NextReset = time.Now().AddDate(0, 0, 7)
	case BudgetPeriodMonthly:
		budget.NextReset = time.Now().AddDate(0, 1, 0)
	case BudgetPeriodYearly:
		budget.NextReset = time.Now().AddDate(1, 0, 0)
	}
}

// getCurrentUsageForBudget gets current usage for a budget period
func (t *Tracker) getCurrentUsageForBudget(budget *Budget, usage *ProviderUsage) float64 {
	switch budget.Period {
	case BudgetPeriodDaily:
		today := time.Now().Format("2006-01-02")
		if dailyUsage, exists := usage.DailyUsage[today]; exists {
			switch budget.Type {
			case BudgetTypeCost:
				return dailyUsage.Cost
			case BudgetTypeTokens:
				return float64(dailyUsage.Tokens)
			case BudgetTypeRequests:
				return float64(dailyUsage.Requests)
			}
		}
	case BudgetPeriodMonthly:
		month := time.Now().Format("2006-01")
		if monthlyUsage, exists := usage.MonthlyUsage[month]; exists {
			switch budget.Type {
			case BudgetTypeCost:
				return monthlyUsage.Cost
			case BudgetTypeTokens:
				return float64(monthlyUsage.Tokens)
			case BudgetTypeRequests:
				return float64(monthlyUsage.Requests)
			}
		}
	}
	
	return 0.0
}

// SetBudgetExceededCallback sets the callback for budget exceeded events
func (t *Tracker) SetBudgetExceededCallback(callback func(provider string, budget *Budget, usage *ProviderUsage)) {
	t.onBudgetExceeded = callback
}

// SetAlertTriggeredCallback sets the callback for alert triggered events
func (t *Tracker) SetAlertTriggeredCallback(callback func(alert AlertRule, usage *ProviderUsage)) {
	t.onAlertTriggered = callback
}
