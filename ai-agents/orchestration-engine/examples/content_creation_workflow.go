package examples

import (
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/orchestration-engine/internal/domain/entities"
)

// CreateContentCreationWorkflow creates a comprehensive content creation workflow
func CreateContentCreationWorkflow(createdBy uuid.UUID) *entities.Workflow {
	workflow := entities.NewWorkflow(
		"AI-Powered Content Creation & Publishing",
		"Complete workflow for creating, optimizing, and publishing social media content with AI assistance and feedback analysis",
		entities.WorkflowTypeHybrid,
		createdBy,
	)

	workflow.Category = entities.WorkflowCategoryContentCreation
	workflow.Priority = entities.WorkflowPriorityHigh
	workflow.Tags = []string{"ai", "content", "social-media", "automation", "feedback"}

	// Configure workflow settings
	workflow.Configuration = &entities.WorkflowConfig{
		MaxConcurrentExecutions: 5,
		ExecutionTimeout:        30 * time.Minute,
		RetentionPeriod:         30 * 24 * time.Hour, // 30 days
		NotificationSettings: &entities.NotificationSettings{
			OnStart:    true,
			OnComplete: true,
			OnFailure:  true,
			Recipients: []string{"content-team@gocoffee.com"},
			Channels:   []string{"email", "slack"},
		},
		SecuritySettings: &entities.SecuritySettings{
			RequireApproval: true,
			AllowedRoles:    []string{"content-manager", "marketing-manager"},
			EncryptData:     true,
			AuditLevel:      "detailed",
		},
		Monitoring: &entities.MonitoringConfig{
			EnableMetrics: true,
			EnableTracing: true,
			EnableLogging: true,
			LogLevel:      "info",
		},
	}

	// Define workflow variables
	workflow.Variables = map[string]interface{}{
		"brand_id":           "",
		"campaign_id":        "",
		"content_topic":      "",
		"target_platforms":   []string{"instagram", "facebook", "twitter"},
		"content_type":       "post",
		"tone":              "friendly",
		"auto_publish":      false,
		"require_approval":  true,
		"generate_variations": true,
		"analyze_feedback":  true,
	}

	// Create workflow definition
	definition := &entities.WorkflowDefinition{
		StartStep: "validate_input",
		EndSteps:  []string{"workflow_complete", "workflow_failed"},
		Steps:     make(map[string]*entities.StepDefinition),
		Connections: []*entities.StepConnection{},
		ErrorHandling: &entities.ErrorHandlingConfig{
			Strategy:        entities.ErrorHandlingStrategyFallback,
			FallbackStep:    "handle_error",
			NotifyOnError:   true,
			ContinueOnError: false,
		},
		Timeouts: map[string]time.Duration{
			"ai_content_generation": 5 * time.Minute,
			"content_optimization":  3 * time.Minute,
			"feedback_analysis":     2 * time.Minute,
		},
		RetryPolicies: map[string]*entities.RetryPolicy{
			"default": {
				MaxAttempts:   3,
				InitialDelay:  1 * time.Second,
				MaxDelay:      30 * time.Second,
				BackoffFactor: 2.0,
				RetryableErrors: []string{"timeout", "connection_error", "temporary_failure"},
			},
		},
	}

	// Step 1: Validate Input
	definition.Steps["validate_input"] = &entities.StepDefinition{
		ID:          "validate_input",
		Name:        "Validate Input Parameters",
		Description: "Validate all required input parameters for content creation",
		Type:        entities.StepTypeValidation,
		Parameters: map[string]interface{}{
			"required_fields": []string{"brand_id", "content_topic", "target_platforms"},
		},
		Timeout:     30 * time.Second,
		RetryPolicy: definition.RetryPolicies["default"],
	}

	// Step 2: Generate Initial Content
	definition.Steps["generate_content"] = &entities.StepDefinition{
		ID:          "generate_content",
		Name:        "AI Content Generation",
		Description: "Generate initial content using AI based on topic and brand guidelines",
		Type:        entities.StepTypeAgent,
		AgentType:   "social-media-content",
		Action:      "create_content",
		InputMapping: map[string]string{
			"title":       "content_topic",
			"type":        "content_type",
			"brand_id":    "brand_id",
			"tone":        "tone",
			"created_by":  "workflow_user",
		},
		Timeout:     definition.Timeouts["ai_content_generation"],
		RetryPolicy: definition.RetryPolicies["default"],
	}

	// Step 3: Analyze Content Quality
	definition.Steps["analyze_quality"] = &entities.StepDefinition{
		ID:          "analyze_quality",
		Name:        "Content Quality Analysis",
		Description: "Analyze content quality and provide optimization suggestions",
		Type:        entities.StepTypeAgent,
		AgentType:   "social-media-content",
		Action:      "analyze_content",
		InputMapping: map[string]string{
			"content_id": "generate_content.content.id",
		},
		Timeout:     2 * time.Minute,
		RetryPolicy: definition.RetryPolicies["default"],
	}

	// Step 4: Quality Gate Decision
	definition.Steps["quality_gate"] = &entities.StepDefinition{
		ID:          "quality_gate",
		Name:        "Quality Gate Decision",
		Description: "Decide if content quality meets standards",
		Type:        entities.StepTypeCondition,
		Conditions: []*entities.Condition{
			{
				Expression: "analyze_quality.quality.overall_score",
				Operator:   entities.ConditionOperatorGreaterOrEqual,
				Value:      75.0,
			},
		},
	}

	// Step 5: Optimize Content (if quality is low)
	definition.Steps["optimize_content"] = &entities.StepDefinition{
		ID:          "optimize_content",
		Name:        "Content Optimization",
		Description: "Optimize content based on quality analysis suggestions",
		Type:        entities.StepTypeAgent,
		AgentType:   "social-media-content",
		Action:      "enhance_content",
		InputMapping: map[string]string{
			"content_id":   "generate_content.content.id",
			"suggestions":  "analyze_quality.suggestions",
		},
		Timeout:     definition.Timeouts["content_optimization"],
		RetryPolicy: definition.RetryPolicies["default"],
	}

	// Step 6: Generate Platform Variations
	definition.Steps["generate_variations"] = &entities.StepDefinition{
		ID:          "generate_variations",
		Name:        "Generate Platform Variations",
		Description: "Generate platform-specific content variations",
		Type:        entities.StepTypeAgent,
		AgentType:   "social-media-content",
		Action:      "generate_variations",
		InputMapping: map[string]string{
			"content_id": "generate_content.content.id",
			"platforms":  "target_platforms",
			"count":      "3",
		},
		Conditions: []*entities.Condition{
			{
				Expression: "generate_variations",
				Operator:   entities.ConditionOperatorEquals,
				Value:      true,
			},
		},
		Timeout:     3 * time.Minute,
		RetryPolicy: definition.RetryPolicies["default"],
	}

	// Step 7: Content Approval (if required)
	definition.Steps["content_approval"] = &entities.StepDefinition{
		ID:          "content_approval",
		Name:        "Content Approval",
		Description: "Wait for content approval from designated approvers",
		Type:        entities.StepTypeWait,
		Parameters: map[string]interface{}{
			"approval_required": true,
			"approvers":        []string{"content-manager", "brand-manager"},
			"timeout":          "24h",
		},
		Conditions: []*entities.Condition{
			{
				Expression: "require_approval",
				Operator:   entities.ConditionOperatorEquals,
				Value:      true,
			},
		},
		Timeout: 24 * time.Hour,
	}

	// Step 8: Schedule Content
	definition.Steps["schedule_content"] = &entities.StepDefinition{
		ID:          "schedule_content",
		Name:        "Schedule Content Publishing",
		Description: "Schedule content for optimal publishing times",
		Type:        entities.StepTypeAgent,
		AgentType:   "social-media-content",
		Action:      "schedule_content",
		InputMapping: map[string]string{
			"content_id": "generate_content.content.id",
			"platforms":  "target_platforms",
		},
		Timeout:     1 * time.Minute,
		RetryPolicy: definition.RetryPolicies["default"],
	}

	// Step 9: Publish Content (if auto-publish enabled)
	definition.Steps["publish_content"] = &entities.StepDefinition{
		ID:          "publish_content",
		Name:        "Publish Content",
		Description: "Publish content to selected platforms",
		Type:        entities.StepTypeAgent,
		AgentType:   "social-media-content",
		Action:      "publish_content",
		InputMapping: map[string]string{
			"content_id": "generate_content.content.id",
			"platforms":  "target_platforms",
		},
		Conditions: []*entities.Condition{
			{
				Expression: "auto_publish",
				Operator:   entities.ConditionOperatorEquals,
				Value:      true,
			},
		},
		Timeout:     2 * time.Minute,
		RetryPolicy: definition.RetryPolicies["default"],
	}

	// Step 10: Monitor Initial Engagement
	definition.Steps["monitor_engagement"] = &entities.StepDefinition{
		ID:          "monitor_engagement",
		Name:        "Monitor Initial Engagement",
		Description: "Monitor initial engagement metrics for published content",
		Type:        entities.StepTypeWait,
		Parameters: map[string]interface{}{
			"duration": "2h",
		},
		Timeout: 3 * time.Hour,
	}

	// Step 11: Collect Feedback
	definition.Steps["collect_feedback"] = &entities.StepDefinition{
		ID:          "collect_feedback",
		Name:        "Collect User Feedback",
		Description: "Collect and analyze user feedback on published content",
		Type:        entities.StepTypeAgent,
		AgentType:   "feedback-analyst",
		Action:      "extract_insights",
		InputMapping: map[string]string{
			"content_id": "generate_content.content.id",
			"time_range": "2h",
		},
		Conditions: []*entities.Condition{
			{
				Expression: "analyze_feedback",
				Operator:   entities.ConditionOperatorEquals,
				Value:      true,
			},
		},
		Timeout:     definition.Timeouts["feedback_analysis"],
		RetryPolicy: definition.RetryPolicies["default"],
	}

	// Step 12: Generate Performance Report
	definition.Steps["generate_report"] = &entities.StepDefinition{
		ID:          "generate_report",
		Name:        "Generate Performance Report",
		Description: "Generate comprehensive performance report with insights",
		Type:        entities.StepTypeTransform,
		InputMapping: map[string]string{
			"content_data":    "generate_content.content",
			"engagement_data": "monitor_engagement.metrics",
			"feedback_data":   "collect_feedback.insights",
		},
		Timeout: 1 * time.Minute,
	}

	// Step 13: Workflow Complete
	definition.Steps["workflow_complete"] = &entities.StepDefinition{
		ID:          "workflow_complete",
		Name:        "Workflow Complete",
		Description: "Mark workflow as successfully completed",
		Type:        entities.StepTypeNotification,
		Parameters: map[string]interface{}{
			"message": "Content creation workflow completed successfully",
			"include_report": true,
		},
	}

	// Error Handling Step
	definition.Steps["handle_error"] = &entities.StepDefinition{
		ID:          "handle_error",
		Name:        "Handle Workflow Error",
		Description: "Handle workflow errors and notify stakeholders",
		Type:        entities.StepTypeNotification,
		Parameters: map[string]interface{}{
			"message": "Content creation workflow encountered an error",
			"severity": "error",
		},
	}

	// Step 14: Workflow Failed
	definition.Steps["workflow_failed"] = &entities.StepDefinition{
		ID:          "workflow_failed",
		Name:        "Workflow Failed",
		Description: "Mark workflow as failed",
		Type:        entities.StepTypeNotification,
		Parameters: map[string]interface{}{
			"message": "Content creation workflow failed",
			"severity": "critical",
		},
	}

	// Define step connections (workflow flow)
	connections := []*entities.StepConnection{
		// Main flow
		{FromStep: "validate_input", ToStep: "generate_content", IsDefault: true},
		{FromStep: "generate_content", ToStep: "analyze_quality", IsDefault: true},
		{FromStep: "analyze_quality", ToStep: "quality_gate", IsDefault: true},
		
		// Quality gate branches
		{
			FromStep: "quality_gate",
			ToStep:   "generate_variations",
			Condition: &entities.Condition{
				Expression: "analyze_quality.quality.overall_score",
				Operator:   entities.ConditionOperatorGreaterOrEqual,
				Value:      75.0,
			},
		},
		{
			FromStep: "quality_gate",
			ToStep:   "optimize_content",
			Condition: &entities.Condition{
				Expression: "analyze_quality.quality.overall_score",
				Operator:   entities.ConditionOperatorLessThan,
				Value:      75.0,
			},
		},
		
		// After optimization, re-analyze
		{FromStep: "optimize_content", ToStep: "analyze_quality", IsDefault: true},
		
		// Variations to approval/scheduling
		{
			FromStep: "generate_variations",
			ToStep:   "content_approval",
			Condition: &entities.Condition{
				Expression: "require_approval",
				Operator:   entities.ConditionOperatorEquals,
				Value:      true,
			},
		},
		{
			FromStep: "generate_variations",
			ToStep:   "schedule_content",
			Condition: &entities.Condition{
				Expression: "require_approval",
				Operator:   entities.ConditionOperatorEquals,
				Value:      false,
			},
		},
		
		// After approval
		{FromStep: "content_approval", ToStep: "schedule_content", IsDefault: true},
		
		// Publishing flow
		{
			FromStep: "schedule_content",
			ToStep:   "publish_content",
			Condition: &entities.Condition{
				Expression: "auto_publish",
				Operator:   entities.ConditionOperatorEquals,
				Value:      true,
			},
		},
		{
			FromStep: "schedule_content",
			ToStep:   "workflow_complete",
			Condition: &entities.Condition{
				Expression: "auto_publish",
				Operator:   entities.ConditionOperatorEquals,
				Value:      false,
			},
		},
		
		// Monitoring and feedback flow
		{FromStep: "publish_content", ToStep: "monitor_engagement", IsDefault: true},
		{
			FromStep: "monitor_engagement",
			ToStep:   "collect_feedback",
			Condition: &entities.Condition{
				Expression: "analyze_feedback",
				Operator:   entities.ConditionOperatorEquals,
				Value:      true,
			},
		},
		{
			FromStep: "monitor_engagement",
			ToStep:   "generate_report",
			Condition: &entities.Condition{
				Expression: "analyze_feedback",
				Operator:   entities.ConditionOperatorEquals,
				Value:      false,
			},
		},
		
		// Final steps
		{FromStep: "collect_feedback", ToStep: "generate_report", IsDefault: true},
		{FromStep: "generate_report", ToStep: "workflow_complete", IsDefault: true},
		
		// Error handling
		{FromStep: "handle_error", ToStep: "workflow_failed", IsDefault: true},
	}

	definition.Connections = connections

	// Add triggers
	workflow.Triggers = []*entities.WorkflowTrigger{
		{
			ID:         uuid.New(),
			WorkflowID: workflow.ID,
			Name:       "Manual Content Creation",
			Type:       entities.TriggerTypeManual,
			IsActive:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			WorkflowID: workflow.ID,
			Name:       "Scheduled Content Creation",
			Type:       entities.TriggerTypeSchedule,
			Configuration: map[string]interface{}{
				"schedule": "0 9 * * 1-5", // 9 AM on weekdays
				"timezone": "UTC",
			},
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:         uuid.New(),
			WorkflowID: workflow.ID,
			Name:       "Campaign Content Request",
			Type:       entities.TriggerTypeEvent,
			Configuration: map[string]interface{}{
				"event_type": "campaign.content_requested",
				"source":     "campaign-manager",
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
