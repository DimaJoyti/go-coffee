package examples

import (
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/orchestration-engine/internal/domain/entities"
)

// CreateBeverageInnovationWorkflow creates a comprehensive beverage innovation and testing workflow
func CreateBeverageInnovationWorkflow(createdBy uuid.UUID) *entities.Workflow {
	workflow := entities.NewWorkflow(
		"AI-Powered Beverage Innovation & Testing Pipeline",
		"Complete workflow for inventing, testing, and launching new coffee beverages with AI assistance, market analysis, and customer feedback integration",
		entities.WorkflowTypeHybrid,
		createdBy,
	)

	workflow.Category = entities.WorkflowCategoryOperations
	workflow.Priority = entities.WorkflowPriorityHigh
	workflow.Tags = []string{"ai", "beverage", "innovation", "testing", "market-research", "customer-feedback"}

	// Configure workflow settings
	workflow.Configuration = &entities.WorkflowConfig{
		MaxConcurrentExecutions: 3, // Limited for quality control
		ExecutionTimeout:        2 * time.Hour, // Long process
		RetentionPeriod:         90 * 24 * time.Hour, // 90 days for R&D
		NotificationSettings: &entities.NotificationSettings{
			OnStart:    true,
			OnComplete: true,
			OnFailure:  true,
			Recipients: []string{"innovation-team@gocoffee.com", "product-team@gocoffee.com"},
			Channels:   []string{"email", "slack", "teams"},
		},
		SecuritySettings: &entities.SecuritySettings{
			RequireApproval: true,
			AllowedRoles:    []string{"innovation-manager", "product-manager", "head-barista"},
			EncryptData:     true,
			AuditLevel:      "detailed",
		},
		Monitoring: &entities.MonitoringConfig{
			EnableMetrics: true,
			EnableTracing: true,
			EnableLogging: true,
			LogLevel:      "debug", // Detailed logging for R&D
		},
	}

	// Define workflow variables
	workflow.Variables = map[string]interface{}{
		"innovation_request_id":    "",
		"target_market":           "premium_coffee",
		"flavor_profile":          "",
		"dietary_requirements":    []string{},
		"seasonal_preferences":    "",
		"budget_range":           "medium",
		"target_launch_date":     "",
		"require_taste_testing":  true,
		"require_market_analysis": true,
		"require_cost_analysis":  true,
		"minimum_rating_threshold": 4.0,
		"max_iterations":         3,
		"auto_approve_high_scores": false,
	}

	// Create workflow definition
	definition := &entities.WorkflowDefinition{
		StartStep: "validate_innovation_request",
		EndSteps:  []string{"innovation_complete", "innovation_failed", "innovation_cancelled"},
		Steps:     make(map[string]*entities.StepDefinition),
		Connections: []*entities.StepConnection{},
		ErrorHandling: &entities.ErrorHandlingConfig{
			Strategy:        entities.ErrorHandlingStrategyFallback,
			FallbackStep:    "handle_innovation_error",
			NotifyOnError:   true,
			ContinueOnError: false,
		},
		Timeouts: map[string]time.Duration{
			"beverage_invention":    10 * time.Minute,
			"market_analysis":       15 * time.Minute,
			"cost_analysis":         5 * time.Minute,
			"taste_testing":         30 * time.Minute,
			"feedback_analysis":     10 * time.Minute,
		},
		RetryPolicies: map[string]*entities.RetryPolicy{
			"default": {
				MaxAttempts:   3,
				InitialDelay:  2 * time.Second,
				MaxDelay:      60 * time.Second,
				BackoffFactor: 2.0,
				RetryableErrors: []string{"timeout", "connection_error", "temporary_failure"},
			},
			"ai_generation": {
				MaxAttempts:   5,
				InitialDelay:  1 * time.Second,
				MaxDelay:      30 * time.Second,
				BackoffFactor: 1.5,
				RetryableErrors: []string{"ai_service_busy", "rate_limit", "temporary_failure"},
			},
		},
	}

	// Step 1: Validate Innovation Request
	definition.Steps["validate_innovation_request"] = &entities.StepDefinition{
		ID:          "validate_innovation_request",
		Name:        "Validate Innovation Request",
		Description: "Validate all required parameters for beverage innovation",
		Type:        entities.StepTypeValidation,
		Parameters: map[string]interface{}{
			"required_fields": []string{"target_market", "flavor_profile"},
			"optional_fields": []string{"dietary_requirements", "seasonal_preferences"},
		},
		Timeout:     30 * time.Second,
		RetryPolicy: definition.RetryPolicies["default"],
	}

	// Step 2: Market Research & Trend Analysis
	definition.Steps["market_research"] = &entities.StepDefinition{
		ID:          "market_research",
		Name:        "Market Research & Trend Analysis",
		Description: "Analyze current market trends and customer preferences",
		Type:        entities.StepTypeAgent,
		AgentType:   "market-research",
		Action:      "analyze_trends",
		InputMapping: map[string]string{
			"target_market":        "target_market",
			"flavor_profile":       "flavor_profile",
			"seasonal_preferences": "seasonal_preferences",
		},
		Conditions: []*entities.Condition{
			{
				Expression: "require_market_analysis",
				Operator:   entities.ConditionOperatorEquals,
				Value:      true,
			},
		},
		Timeout:     definition.Timeouts["market_analysis"],
		RetryPolicy: definition.RetryPolicies["default"],
	}

	// Step 3: AI Beverage Invention
	definition.Steps["invent_beverage"] = &entities.StepDefinition{
		ID:          "invent_beverage",
		Name:        "AI Beverage Invention",
		Description: "Generate innovative beverage recipes using AI",
		Type:        entities.StepTypeAgent,
		AgentType:   "beverage-inventor",
		Action:      "invent_beverage",
		InputMapping: map[string]string{
			"flavor_profile":       "flavor_profile",
			"dietary_requirements": "dietary_requirements",
			"market_trends":        "market_research.trends",
			"seasonal_preferences": "seasonal_preferences",
			"target_market":        "target_market",
		},
		Timeout:     definition.Timeouts["beverage_invention"],
		RetryPolicy: definition.RetryPolicies["ai_generation"],
	}

	// Step 4: Recipe Optimization
	definition.Steps["optimize_recipe"] = &entities.StepDefinition{
		ID:          "optimize_recipe",
		Name:        "Recipe Optimization",
		Description: "Optimize recipe for taste, cost, and production feasibility",
		Type:        entities.StepTypeAgent,
		AgentType:   "beverage-inventor",
		Action:      "optimize_recipe",
		InputMapping: map[string]string{
			"recipe":       "invent_beverage.recipe",
			"budget_range": "budget_range",
			"constraints":  "invent_beverage.constraints",
		},
		Timeout:     5 * time.Minute,
		RetryPolicy: definition.RetryPolicies["ai_generation"],
	}

	// Step 5: Cost Analysis
	definition.Steps["cost_analysis"] = &entities.StepDefinition{
		ID:          "cost_analysis",
		Name:        "Cost Analysis & Pricing",
		Description: "Analyze production costs and suggest pricing strategy",
		Type:        entities.StepTypeAgent,
		AgentType:   "inventory",
		Action:      "calculate_costs",
		InputMapping: map[string]string{
			"recipe":        "optimize_recipe.optimized_recipe",
			"ingredients":   "optimize_recipe.ingredients",
			"target_margin": "budget_range",
		},
		Conditions: []*entities.Condition{
			{
				Expression: "require_cost_analysis",
				Operator:   entities.ConditionOperatorEquals,
				Value:      true,
			},
		},
		Timeout:     definition.Timeouts["cost_analysis"],
		RetryPolicy: definition.RetryPolicies["default"],
	}

	// Step 6: Feasibility Assessment
	definition.Steps["feasibility_check"] = &entities.StepDefinition{
		ID:          "feasibility_check",
		Name:        "Production Feasibility Assessment",
		Description: "Assess production feasibility and equipment requirements",
		Type:        entities.StepTypeCondition,
		Conditions: []*entities.Condition{
			{
				Expression: "cost_analysis.production_cost",
				Operator:   entities.ConditionOperatorLessOrEqual,
				Value:      100.0, // Max cost per unit
			},
			{
				Expression: "optimize_recipe.complexity_score",
				Operator:   entities.ConditionOperatorLessOrEqual,
				Value:      7.0, // Max complexity (1-10 scale)
			},
		},
	}

	// Step 7: Recipe Refinement (if feasibility fails)
	definition.Steps["refine_recipe"] = &entities.StepDefinition{
		ID:          "refine_recipe",
		Name:        "Recipe Refinement",
		Description: "Refine recipe to meet feasibility constraints",
		Type:        entities.StepTypeAgent,
		AgentType:   "beverage-inventor",
		Action:      "refine_recipe",
		InputMapping: map[string]string{
			"recipe":           "optimize_recipe.optimized_recipe",
			"cost_constraints": "cost_analysis.constraints",
			"feedback":         "feasibility_check.issues",
		},
		Timeout:     definition.Timeouts["beverage_invention"],
		RetryPolicy: definition.RetryPolicies["ai_generation"],
	}

	// Step 8: Iteration Counter
	definition.Steps["check_iterations"] = &entities.StepDefinition{
		ID:          "check_iterations",
		Name:        "Check Iteration Count",
		Description: "Check if maximum iterations reached",
		Type:        entities.StepTypeCondition,
		Conditions: []*entities.Condition{
			{
				Expression: "iteration_count",
				Operator:   entities.ConditionOperatorLessThan,
				Value:      3, // Max iterations
			},
		},
	}

	// Step 9: Prepare Test Batch
	definition.Steps["prepare_test_batch"] = &entities.StepDefinition{
		ID:          "prepare_test_batch",
		Name:        "Prepare Test Batch",
		Description: "Generate detailed instructions for test batch preparation",
		Type:        entities.StepTypeAgent,
		AgentType:   "beverage-inventor",
		Action:      "generate_batch_instructions",
		InputMapping: map[string]string{
			"final_recipe":    "optimize_recipe.optimized_recipe",
			"batch_size":      "small", // Test batch
			"equipment_list":  "cost_analysis.equipment",
		},
		Timeout:     3 * time.Minute,
		RetryPolicy: definition.RetryPolicies["default"],
	}

	// Step 10: Taste Testing Coordination
	definition.Steps["coordinate_taste_testing"] = &entities.StepDefinition{
		ID:          "coordinate_taste_testing",
		Name:        "Coordinate Taste Testing",
		Description: "Schedule and coordinate taste testing sessions",
		Type:        entities.StepTypeNotification,
		Parameters: map[string]interface{}{
			"message":    "New beverage ready for taste testing",
			"recipients": []string{"taste-testing-team", "barista-team"},
			"include_recipe": true,
			"include_instructions": true,
		},
		Conditions: []*entities.Condition{
			{
				Expression: "require_taste_testing",
				Operator:   entities.ConditionOperatorEquals,
				Value:      true,
			},
		},
		Timeout: 1 * time.Minute,
	}

	// Step 11: Wait for Taste Testing Results
	definition.Steps["wait_taste_results"] = &entities.StepDefinition{
		ID:          "wait_taste_results",
		Name:        "Wait for Taste Testing Results",
		Description: "Wait for taste testing completion and results",
		Type:        entities.StepTypeWait,
		Parameters: map[string]interface{}{
			"duration":     "24h",
			"max_wait":     "72h",
			"check_interval": "1h",
		},
		Conditions: []*entities.Condition{
			{
				Expression: "require_taste_testing",
				Operator:   entities.ConditionOperatorEquals,
				Value:      true,
			},
		},
		Timeout: 72 * time.Hour,
	}

	// Step 12: Analyze Taste Testing Feedback
	definition.Steps["analyze_taste_feedback"] = &entities.StepDefinition{
		ID:          "analyze_taste_feedback",
		Name:        "Analyze Taste Testing Feedback",
		Description: "Analyze feedback from taste testing sessions",
		Type:        entities.StepTypeAgent,
		AgentType:   "feedback-analyst",
		Action:      "analyze_feedback",
		InputMapping: map[string]string{
			"feedback_data": "wait_taste_results.feedback",
			"recipe_id":     "optimize_recipe.recipe_id",
			"test_type":     "taste_testing",
		},
		Conditions: []*entities.Condition{
			{
				Expression: "require_taste_testing",
				Operator:   entities.ConditionOperatorEquals,
				Value:      true,
			},
		},
		Timeout:     definition.Timeouts["feedback_analysis"],
		RetryPolicy: definition.RetryPolicies["default"],
	}

	// Step 13: Quality Gate - Rating Check
	definition.Steps["quality_gate"] = &entities.StepDefinition{
		ID:          "quality_gate",
		Name:        "Quality Gate - Rating Assessment",
		Description: "Assess if beverage meets quality standards",
		Type:        entities.StepTypeCondition,
		Conditions: []*entities.Condition{
			{
				Expression: "analyze_taste_feedback.average_rating",
				Operator:   entities.ConditionOperatorGreaterOrEqual,
				Value:      4.0,
			},
			{
				Expression: "analyze_taste_feedback.approval_rate",
				Operator:   entities.ConditionOperatorGreaterOrEqual,
				Value:      0.75, // 75% approval rate
			},
		},
	}

	// Step 14: Generate Improvement Suggestions
	definition.Steps["generate_improvements"] = &entities.StepDefinition{
		ID:          "generate_improvements",
		Name:        "Generate Improvement Suggestions",
		Description: "Generate specific improvement suggestions based on feedback",
		Type:        entities.StepTypeAgent,
		AgentType:   "beverage-inventor",
		Action:      "suggest_improvements",
		InputMapping: map[string]string{
			"recipe":           "optimize_recipe.optimized_recipe",
			"feedback_analysis": "analyze_taste_feedback.analysis",
			"low_scores":       "analyze_taste_feedback.negative_feedback",
		},
		Timeout:     5 * time.Minute,
		RetryPolicy: definition.RetryPolicies["ai_generation"],
	}

	// Step 15: Final Recipe Documentation
	definition.Steps["document_recipe"] = &entities.StepDefinition{
		ID:          "document_recipe",
		Name:        "Document Final Recipe",
		Description: "Create comprehensive recipe documentation",
		Type:        entities.StepTypeTransform,
		InputMapping: map[string]string{
			"recipe":           "optimize_recipe.optimized_recipe",
			"cost_analysis":    "cost_analysis.results",
			"taste_feedback":   "analyze_taste_feedback.summary",
			"market_research":  "market_research.insights",
		},
		Timeout: 2 * time.Minute,
	}

	// Step 16: Launch Preparation
	definition.Steps["prepare_launch"] = &entities.StepDefinition{
		ID:          "prepare_launch",
		Name:        "Prepare Product Launch",
		Description: "Prepare marketing materials and launch strategy",
		Type:        entities.StepTypeAgent,
		AgentType:   "social-media-content",
		Action:      "create_launch_campaign",
		InputMapping: map[string]string{
			"product_info":     "document_recipe.final_documentation",
			"target_market":    "target_market",
			"unique_features":  "analyze_taste_feedback.highlights",
			"launch_date":      "target_launch_date",
		},
		Timeout:     10 * time.Minute,
		RetryPolicy: definition.RetryPolicies["default"],
	}

	// Step 17: Innovation Complete
	definition.Steps["innovation_complete"] = &entities.StepDefinition{
		ID:          "innovation_complete",
		Name:        "Innovation Process Complete",
		Description: "Mark innovation process as successfully completed",
		Type:        entities.StepTypeNotification,
		Parameters: map[string]interface{}{
			"message": "Beverage innovation process completed successfully",
			"include_documentation": true,
			"include_launch_plan": true,
			"severity": "success",
		},
	}

	// Error Handling Steps
	definition.Steps["handle_innovation_error"] = &entities.StepDefinition{
		ID:          "handle_innovation_error",
		Name:        "Handle Innovation Error",
		Description: "Handle innovation process errors and notify stakeholders",
		Type:        entities.StepTypeNotification,
		Parameters: map[string]interface{}{
			"message": "Beverage innovation process encountered an error",
			"severity": "error",
			"include_logs": true,
		},
	}

	definition.Steps["innovation_failed"] = &entities.StepDefinition{
		ID:          "innovation_failed",
		Name:        "Innovation Process Failed",
		Description: "Mark innovation process as failed",
		Type:        entities.StepTypeNotification,
		Parameters: map[string]interface{}{
			"message": "Beverage innovation process failed after maximum attempts",
			"severity": "critical",
			"require_manual_review": true,
		},
	}

	definition.Steps["innovation_cancelled"] = &entities.StepDefinition{
		ID:          "innovation_cancelled",
		Name:        "Innovation Process Cancelled",
		Description: "Mark innovation process as cancelled",
		Type:        entities.StepTypeNotification,
		Parameters: map[string]interface{}{
			"message": "Beverage innovation process was cancelled",
			"severity": "warning",
		},
	}

	// Define step connections (workflow flow)
	connections := []*entities.StepConnection{
		// Main validation flow
		{FromStep: "validate_innovation_request", ToStep: "market_research", IsDefault: true},
		
		// Market research branch
		{
			FromStep: "market_research",
			ToStep:   "invent_beverage",
			Condition: &entities.Condition{
				Expression: "require_market_analysis",
				Operator:   entities.ConditionOperatorEquals,
				Value:      true,
			},
		},
		{
			FromStep: "validate_innovation_request",
			ToStep:   "invent_beverage",
			Condition: &entities.Condition{
				Expression: "require_market_analysis",
				Operator:   entities.ConditionOperatorEquals,
				Value:      false,
			},
		},
		
		// Core invention flow
		{FromStep: "invent_beverage", ToStep: "optimize_recipe", IsDefault: true},
		{FromStep: "optimize_recipe", ToStep: "cost_analysis", IsDefault: true},
		
		// Cost analysis branch
		{
			FromStep: "cost_analysis",
			ToStep:   "feasibility_check",
			Condition: &entities.Condition{
				Expression: "require_cost_analysis",
				Operator:   entities.ConditionOperatorEquals,
				Value:      true,
			},
		},
		{
			FromStep: "optimize_recipe",
			ToStep:   "feasibility_check",
			Condition: &entities.Condition{
				Expression: "require_cost_analysis",
				Operator:   entities.ConditionOperatorEquals,
				Value:      false,
			},
		},
		
		// Feasibility branches
		{
			FromStep: "feasibility_check",
			ToStep:   "prepare_test_batch",
			Condition: &entities.Condition{
				Expression: "feasibility_check.passed",
				Operator:   entities.ConditionOperatorEquals,
				Value:      true,
			},
		},
		{
			FromStep: "feasibility_check",
			ToStep:   "refine_recipe",
			Condition: &entities.Condition{
				Expression: "feasibility_check.passed",
				Operator:   entities.ConditionOperatorEquals,
				Value:      false,
			},
		},
		
		// Refinement loop
		{FromStep: "refine_recipe", ToStep: "check_iterations", IsDefault: true},
		{
			FromStep: "check_iterations",
			ToStep:   "cost_analysis",
			Condition: &entities.Condition{
				Expression: "iteration_count",
				Operator:   entities.ConditionOperatorLessThan,
				Value:      3,
			},
		},
		{
			FromStep: "check_iterations",
			ToStep:   "innovation_failed",
			Condition: &entities.Condition{
				Expression: "iteration_count",
				Operator:   entities.ConditionOperatorGreaterOrEqual,
				Value:      3,
			},
		},
		
		// Testing flow
		{FromStep: "prepare_test_batch", ToStep: "coordinate_taste_testing", IsDefault: true},
		{
			FromStep: "coordinate_taste_testing",
			ToStep:   "wait_taste_results",
			Condition: &entities.Condition{
				Expression: "require_taste_testing",
				Operator:   entities.ConditionOperatorEquals,
				Value:      true,
			},
		},
		{
			FromStep: "prepare_test_batch",
			ToStep:   "document_recipe",
			Condition: &entities.Condition{
				Expression: "require_taste_testing",
				Operator:   entities.ConditionOperatorEquals,
				Value:      false,
			},
		},
		
		// Feedback analysis flow
		{FromStep: "wait_taste_results", ToStep: "analyze_taste_feedback", IsDefault: true},
		{FromStep: "analyze_taste_feedback", ToStep: "quality_gate", IsDefault: true},
		
		// Quality gate branches
		{
			FromStep: "quality_gate",
			ToStep:   "document_recipe",
			Condition: &entities.Condition{
				Expression: "quality_gate.passed",
				Operator:   entities.ConditionOperatorEquals,
				Value:      true,
			},
		},
		{
			FromStep: "quality_gate",
			ToStep:   "generate_improvements",
			Condition: &entities.Condition{
				Expression: "quality_gate.passed",
				Operator:   entities.ConditionOperatorEquals,
				Value:      false,
			},
		},
		
		// Improvement loop
		{FromStep: "generate_improvements", ToStep: "refine_recipe", IsDefault: true},
		
		// Final steps
		{FromStep: "document_recipe", ToStep: "prepare_launch", IsDefault: true},
		{FromStep: "prepare_launch", ToStep: "innovation_complete", IsDefault: true},
		
		// Error handling
		{FromStep: "handle_innovation_error", ToStep: "innovation_failed", IsDefault: true},
	}

	definition.Connections = connections

	// Add triggers
	workflow.Triggers = []*entities.WorkflowTrigger{
		{
			ID:         uuid.New(),
			WorkflowID: workflow.ID,
			Name:       "Manual Innovation Request",
			Type:       entities.TriggerTypeManual,
			IsActive:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			WorkflowID: workflow.ID,
			Name:       "Seasonal Innovation Schedule",
			Type:       entities.TriggerTypeSchedule,
			Configuration: map[string]interface{}{
				"schedule": "0 0 1 */3 *", // First day of every quarter
				"timezone": "UTC",
			},
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:         uuid.New(),
			WorkflowID: workflow.ID,
			Name:       "Market Trend Innovation Trigger",
			Type:       entities.TriggerTypeEvent,
			Configuration: map[string]interface{}{
				"event_type": "market.trend_detected",
				"source":     "market-research-agent",
				"conditions": map[string]interface{}{
					"trend_strength": ">= 0.8",
					"market_segment": "coffee",
				},
			},
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:         uuid.New(),
			WorkflowID: workflow.ID,
			Name:       "Customer Request Innovation",
			Type:       entities.TriggerTypeEvent,
			Configuration: map[string]interface{}{
				"event_type": "customer.innovation_request",
				"source":     "customer-service",
				"conditions": map[string]interface{}{
					"request_count": ">= 10",
					"category":      "beverage",
				},
			},
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	workflow.Definition = definition
	workflow.Status = entities.WorkflowStatusActive
	workflow.IsActive = true

	return workflow
}
