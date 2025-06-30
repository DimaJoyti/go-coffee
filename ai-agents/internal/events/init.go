package events

import (
	"log"
)

// InitializeEventRegistry initializes the default event registry with all known event types
func InitializeEventRegistry() error {
	// Note: In a real implementation, these would import the generated protobuf types
	// For now, we'll create placeholder registrations
	
	// Beverage Events
	if err := registerBeverageEvents(); err != nil {
		return err
	}
	
	// Task Events
	if err := registerTaskEvents(); err != nil {
		return err
	}
	
	// Notification Events
	if err := registerNotificationEvents(); err != nil {
		return err
	}
	
	// Social Media Events
	if err := registerSocialMediaEvents(); err != nil {
		return err
	}
	
	// Common Events
	if err := registerCommonEvents(); err != nil {
		return err
	}
	
	log.Println("Event registry initialized successfully")
	return nil
}

// registerBeverageEvents registers all beverage-related events
func registerBeverageEvents() error {
	// Note: These would use actual protobuf message types in a real implementation
	
	// Register BeverageCreatedEvent
	// err := RegisterEvent("beverage.created", "1.0", "Event fired when a new beverage is created", &events.BeverageCreatedEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register BeverageUpdatedEvent
	// err = RegisterEvent("beverage.updated", "1.0", "Event fired when a beverage is updated", &events.BeverageUpdatedEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register BeverageStatusChangedEvent
	// err = RegisterEvent("beverage.status_changed", "1.0", "Event fired when a beverage status changes", &events.BeverageStatusChangedEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register RecipeRequestEvent
	// err = RegisterEvent("recipe.request", "1.0", "Event fired when a recipe is requested", &events.RecipeRequestEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register IngredientDiscoveredEvent
	// err = RegisterEvent("ingredient.discovered", "1.0", "Event fired when a new ingredient is discovered", &events.IngredientDiscoveredEvent{})
	// if err != nil {
	//     return err
	// }
	
	return nil
}

// registerTaskEvents registers all task-related events
func registerTaskEvents() error {
	// Register TaskCreatedEvent
	// err := RegisterEvent("task.created", "1.0", "Event fired when a new task is created", &events.TaskCreatedEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register TaskUpdatedEvent
	// err = RegisterEvent("task.updated", "1.0", "Event fired when a task is updated", &events.TaskUpdatedEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register TaskStatusChangedEvent
	// err = RegisterEvent("task.status_changed", "1.0", "Event fired when a task status changes", &events.TaskStatusChangedEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register TaskAssignedEvent
	// err = RegisterEvent("task.assigned", "1.0", "Event fired when a task is assigned", &events.TaskAssignedEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register TaskCompletedEvent
	// err = RegisterEvent("task.completed", "1.0", "Event fired when a task is completed", &events.TaskCompletedEvent{})
	// if err != nil {
	//     return err
	// }
	
	return nil
}

// registerNotificationEvents registers all notification-related events
func registerNotificationEvents() error {
	// Register NotificationSentEvent
	// err := RegisterEvent("notification.sent", "1.0", "Event fired when a notification is sent", &events.NotificationSentEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register NotificationFailedEvent
	// err = RegisterEvent("notification.failed", "1.0", "Event fired when a notification fails", &events.NotificationFailedEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register SlackMessageSentEvent
	// err = RegisterEvent("slack.message_sent", "1.0", "Event fired when a Slack message is sent", &events.SlackMessageSentEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register EmailSentEvent
	// err = RegisterEvent("email.sent", "1.0", "Event fired when an email is sent", &events.EmailSentEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register WebhookSentEvent
	// err = RegisterEvent("webhook.sent", "1.0", "Event fired when a webhook is sent", &events.WebhookSentEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register AlertTriggeredEvent
	// err = RegisterEvent("alert.triggered", "1.0", "Event fired when an alert is triggered", &events.AlertTriggeredEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register AlertResolvedEvent
	// err = RegisterEvent("alert.resolved", "1.0", "Event fired when an alert is resolved", &events.AlertResolvedEvent{})
	// if err != nil {
	//     return err
	// }
	
	return nil
}

// registerSocialMediaEvents registers all social media-related events
func registerSocialMediaEvents() error {
	// Register SocialMediaPostCreatedEvent
	// err := RegisterEvent("social_media.post_created", "1.0", "Event fired when a social media post is created", &events.SocialMediaPostCreatedEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register SocialMediaPostPublishedEvent
	// err = RegisterEvent("social_media.post_published", "1.0", "Event fired when a social media post is published", &events.SocialMediaPostPublishedEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register SocialMediaEngagementEvent
	// err = RegisterEvent("social_media.engagement", "1.0", "Event fired when there's engagement on a social media post", &events.SocialMediaEngagementEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register ContentGenerationRequestEvent
	// err = RegisterEvent("content.generation_request", "1.0", "Event fired when content generation is requested", &events.ContentGenerationRequestEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register ContentGeneratedEvent
	// err = RegisterEvent("content.generated", "1.0", "Event fired when content is generated", &events.ContentGeneratedEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register InfluencerMentionEvent
	// err = RegisterEvent("influencer.mention", "1.0", "Event fired when an influencer mentions the brand", &events.InfluencerMentionEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register SocialMediaAnalyticsEvent
	// err = RegisterEvent("social_media.analytics", "1.0", "Event fired with social media analytics data", &events.SocialMediaAnalyticsEvent{})
	// if err != nil {
	//     return err
	// }
	
	return nil
}

// registerCommonEvents registers all common system events
func registerCommonEvents() error {
	// Register ErrorEvent
	// err := RegisterEvent("system.error", "1.0", "Event fired when a system error occurs", &events.ErrorEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register HealthCheckEvent
	// err = RegisterEvent("system.health_check", "1.0", "Event fired with health check results", &events.HealthCheckEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register MetricEvent
	// err = RegisterEvent("system.metric", "1.0", "Event fired with metric measurements", &events.MetricEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register AuditEvent
	// err = RegisterEvent("system.audit", "1.0", "Event fired for audit logging", &events.AuditEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register ConfigurationChangedEvent
	// err = RegisterEvent("system.config_changed", "1.0", "Event fired when configuration changes", &events.ConfigurationChangedEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register ServiceStartedEvent
	// err = RegisterEvent("system.service_started", "1.0", "Event fired when a service starts", &events.ServiceStartedEvent{})
	// if err != nil {
	//     return err
	// }
	
	// Register ServiceStoppedEvent
	// err = RegisterEvent("system.service_stopped", "1.0", "Event fired when a service stops", &events.ServiceStoppedEvent{})
	// if err != nil {
	//     return err
	// }
	
	return nil
}

// GetEventTopicMapping returns the mapping of event types to Kafka topics
func GetEventTopicMapping() map[string]string {
	return map[string]string{
		// Beverage Events
		"beverage.created":        "beverage.created",
		"beverage.updated":        "beverage.updated",
		"beverage.status_changed": "beverage.updated",
		"recipe.request":          "recipe.requests",
		"ingredient.discovered":   "ingredient.discovered",
		
		// Task Events
		"task.created":        "task.created",
		"task.updated":        "task.updated",
		"task.status_changed": "task.updated",
		"task.assigned":       "task.assigned",
		"task.completed":      "task.completed",
		
		// Notification Events
		"notification.sent":    "notifications",
		"notification.failed":  "notifications",
		"slack.message_sent":   "notifications.slack",
		"email.sent":          "notifications.email",
		"webhook.sent":        "notifications.webhook",
		"alert.triggered":     "alerts",
		"alert.resolved":      "alerts",
		
		// Social Media Events
		"social_media.post_created":    "social_media.posts",
		"social_media.post_published":  "social_media.posts",
		"social_media.engagement":      "social_media.engagement",
		"content.generation_request":   "content.requests",
		"content.generated":           "content.generated",
		"influencer.mention":          "social_media.mentions",
		"social_media.analytics":      "social_media.analytics",
		
		// Common Events
		"system.error":          "system.errors",
		"system.health_check":   "system.health",
		"system.metric":         "system.metrics",
		"system.audit":          "system.audit",
		"system.config_changed": "system.config",
		"system.service_started": "system.lifecycle",
		"system.service_stopped": "system.lifecycle",
	}
}

// GetEventVersions returns the supported versions for each event type
func GetEventVersions() map[string][]string {
	return map[string][]string{
		// Beverage Events
		"beverage.created":        {"1.0"},
		"beverage.updated":        {"1.0"},
		"beverage.status_changed": {"1.0"},
		"recipe.request":          {"1.0"},
		"ingredient.discovered":   {"1.0"},
		
		// Task Events
		"task.created":        {"1.0"},
		"task.updated":        {"1.0"},
		"task.status_changed": {"1.0"},
		"task.assigned":       {"1.0"},
		"task.completed":      {"1.0"},
		
		// Notification Events
		"notification.sent":    {"1.0"},
		"notification.failed":  {"1.0"},
		"slack.message_sent":   {"1.0"},
		"email.sent":          {"1.0"},
		"webhook.sent":        {"1.0"},
		"alert.triggered":     {"1.0"},
		"alert.resolved":      {"1.0"},
		
		// Social Media Events
		"social_media.post_created":    {"1.0"},
		"social_media.post_published":  {"1.0"},
		"social_media.engagement":      {"1.0"},
		"content.generation_request":   {"1.0"},
		"content.generated":           {"1.0"},
		"influencer.mention":          {"1.0"},
		"social_media.analytics":      {"1.0"},
		
		// Common Events
		"system.error":          {"1.0"},
		"system.health_check":   {"1.0"},
		"system.metric":         {"1.0"},
		"system.audit":          {"1.0"},
		"system.config_changed": {"1.0"},
		"system.service_started": {"1.0"},
		"system.service_stopped": {"1.0"},
	}
}
