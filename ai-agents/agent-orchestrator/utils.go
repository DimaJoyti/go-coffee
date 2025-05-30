package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	redisadvanced "github.com/DimaJoyti/go-coffee/pkg/redis-advanced"
	eventsourcing "github.com/DimaJoyti/go-coffee/pkg/event-sourcing"
)

// initializeAgentIndex creates the search index for agents
func (ao *AgentOrchestrator) initializeAgentIndex() error {
	ao.logger.Info("Initializing agent search index")

	// Define schema for agent index
	schema := map[string]interface{}{
		"name":         "TEXT",
		"type":         "TAG",
		"capabilities": "TAG",
		"status":       "TAG",
		"metadata":     "TEXT",
	}

	// Create vector index for agent search
	if err := ao.searchClient.CreateVectorIndex(context.Background(), "agent_index", schema); err != nil {
		return fmt.Errorf("failed to create agent index: %w", err)
	}

	ao.logger.Info("Agent search index initialized successfully")
	return nil
}

// indexAgent adds an agent to the search index
func (ao *AgentOrchestrator) indexAgent(agent *AgentInfo) error {
	// Prepare agent data for indexing
	fields := map[string]interface{}{
		"id":           agent.ID,
		"name":         agent.Name,
		"type":         agent.Type,
		"capabilities": fmt.Sprintf("%v", agent.Capabilities),
		"status":       agent.Status,
		"endpoint":     agent.Endpoint,
		"version":      agent.Version,
	}

	// Add metadata as searchable text
	if agent.Metadata != nil {
		metadataText := ""
		for k, v := range agent.Metadata {
			metadataText += fmt.Sprintf("%s:%v ", k, v)
		}
		fields["metadata"] = metadataText
	}

	// Generate a simple vector for the agent (in real implementation, this would be an embedding)
	vector := make([]float32, 128)
	for i := range vector {
		vector[i] = float32(i) * 0.01 // Placeholder vector
	}

	return ao.searchClient.AddDocument(context.Background(), agent.ID, fields, vector)
}

// unregisterAgentInternal removes an agent from the system
func (ao *AgentOrchestrator) unregisterAgentInternal(agentID string) error {
	agent, exists := ao.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	// Remove from memory
	delete(ao.agents, agentID)

	// Remove from Redis
	ctx := context.Background()
	if err := ao.redis.Del(ctx, fmt.Sprintf("agent:%s", agentID)).Err(); err != nil {
		ao.logger.Error("Failed to remove agent from Redis", zap.Error(err))
	}

	// Record unregistration event
	event := &eventsourcing.Event{
		AggregateID:   agentID,
		AggregateType: "agent",
		EventType:     "agent_unregistered",
		EventVersion:  2,
		Data: map[string]interface{}{
			"agent_id":   agentID,
			"agent_name": agent.Name,
			"reason":     "manual_unregistration",
		},
	}
	
	if err := ao.eventStore.AppendEvent(ctx, event); err != nil {
		ao.logger.Error("Failed to record unregistration event", zap.Error(err))
	}

	ao.logger.Info("Agent unregistered successfully", 
		zap.String("agent_id", agentID),
		zap.String("agent_name", agent.Name),
	)

	return nil
}

// findSuitableAgents finds agents that match the required capabilities
func (ao *AgentOrchestrator) findSuitableAgents(requiredCapabilities []string) []*AgentInfo {
	var suitableAgents []*AgentInfo

	for _, agent := range ao.agents {
		if agent.Status != "active" {
			continue
		}

		// Check if agent has all required capabilities
		hasAllCapabilities := true
		for _, required := range requiredCapabilities {
			hasCapability := false
			for _, agentCap := range agent.Capabilities {
				if agentCap == required {
					hasCapability = true
					break
				}
			}
			if !hasCapability {
				hasAllCapabilities = false
				break
			}
		}

		if hasAllCapabilities {
			suitableAgents = append(suitableAgents, agent)
		}
	}

	return suitableAgents
}

// getAgentsByType returns agents grouped by type
func (ao *AgentOrchestrator) getAgentsByType() map[string]int {
	byType := make(map[string]int)
	
	for _, agent := range ao.agents {
		byType[agent.Type]++
	}
	
	return byType
}

// getAgentsByStatus returns agents grouped by status
func (ao *AgentOrchestrator) getAgentsByStatus() map[string]int {
	byStatus := make(map[string]int)
	
	for _, agent := range ao.agents {
		byStatus[agent.Status]++
	}
	
	return byStatus
}

// Background services

// startHealthMonitoring starts the health monitoring service
func (ao *AgentOrchestrator) startHealthMonitoring() {
	ao.logger.Info("Starting health monitoring service")
	
	ticker := time.NewTicker(ao.config.HealthInterval)
	defer ticker.Stop()

	for range ticker.C {
		ao.performHealthChecks()
	}
}

// performHealthChecks checks the health of all registered agents
func (ao *AgentOrchestrator) performHealthChecks() {
	ao.logger.Debug("Performing health checks")
	
	now := time.Now()
	unhealthyThreshold := 2 * ao.config.HealthInterval

	for agentID, agent := range ao.agents {
		// Check if agent has been silent for too long
		if now.Sub(agent.LastSeen) > unhealthyThreshold {
			ao.logger.Warn("Agent appears unhealthy", 
				zap.String("agent_id", agentID),
				zap.Duration("last_seen", now.Sub(agent.LastSeen)),
			)
			
			agent.Status = "unhealthy"
			if agent.Health != nil {
				agent.Health.Status = "unhealthy"
				agent.Health.LastCheck = now
				agent.Health.ErrorCount++
			}

			// Record health check metric
			if err := ao.timeSeriesClient.RecordMetric(context.Background(), 
				"agent_health_check", 
				0.0, // 0 for unhealthy
				map[string]string{
					"agent_id": agentID,
					"status":   "unhealthy",
				}); err != nil {
				ao.logger.Error("Failed to record health metric", zap.Error(err))
			}
		} else {
			// Agent is healthy
			if agent.Health != nil {
				agent.Health.Status = "healthy"
				agent.Health.LastCheck = now
			}

			// Record health check metric
			if err := ao.timeSeriesClient.RecordMetric(context.Background(), 
				"agent_health_check", 
				1.0, // 1 for healthy
				map[string]string{
					"agent_id": agentID,
					"status":   "healthy",
				}); err != nil {
				ao.logger.Error("Failed to record health metric", zap.Error(err))
			}
		}
	}
}

// startMetricsCollection starts the metrics collection service
func (ao *AgentOrchestrator) startMetricsCollection() {
	ao.logger.Info("Starting metrics collection service")
	
	ticker := time.NewTicker(ao.config.MetricsInterval)
	defer ticker.Stop()

	for range ticker.C {
		ao.collectMetrics()
	}
}

// collectMetrics collects system metrics
func (ao *AgentOrchestrator) collectMetrics() {
	ao.logger.Debug("Collecting system metrics")
	
	ctx := context.Background()
	now := time.Now()

	// Collect agent metrics
	totalAgents := float64(len(ao.agents))
	activeAgents := 0.0
	healthyAgents := 0.0

	for _, agent := range ao.agents {
		if agent.Status == "active" {
			activeAgents++
		}
		if agent.Health != nil && agent.Health.Status == "healthy" {
			healthyAgents++
		}
	}

	// Record metrics
	metrics := map[string]float64{
		"total_agents":   totalAgents,
		"active_agents":  activeAgents,
		"healthy_agents": healthyAgents,
	}

	for metricName, value := range metrics {
		if err := ao.timeSeriesClient.RecordMetric(ctx, metricName, value, nil); err != nil {
			ao.logger.Error("Failed to record metric", 
				zap.String("metric", metricName), 
				zap.Error(err),
			)
		}
	}

	// Calculate and record health ratio
	healthRatio := 0.0
	if totalAgents > 0 {
		healthRatio = healthyAgents / totalAgents
	}

	if err := ao.timeSeriesClient.RecordMetric(ctx, "health_ratio", healthRatio, nil); err != nil {
		ao.logger.Error("Failed to record health ratio metric", zap.Error(err))
	}

	ao.logger.Debug("Metrics collected successfully", 
		zap.Float64("total_agents", totalAgents),
		zap.Float64("active_agents", activeAgents),
		zap.Float64("healthy_agents", healthyAgents),
		zap.Float64("health_ratio", healthRatio),
	)
}

// startTaskProcessor starts the task processing service
func (ao *AgentOrchestrator) startTaskProcessor() {
	ao.logger.Info("Starting task processor service")
	
	// Subscribe to task events stream
	ctx := context.Background()
	eventChan, err := ao.streamsClient.Subscribe(ctx, []string{"events:task"}, "0")
	if err != nil {
		ao.logger.Error("Failed to subscribe to task events", zap.Error(err))
		return
	}

	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				ao.logger.Info("Task event channel closed")
				return
			}
			
			ao.processTaskEvent(event)

		case <-ctx.Done():
			ao.logger.Info("Task processor stopped")
			return
		}
	}
}

// processTaskEvent processes a task event
func (ao *AgentOrchestrator) processTaskEvent(event *redisadvanced.StreamEvent) {
	ao.logger.Debug("Processing task event", 
		zap.String("event_id", event.ID),
		zap.Any("fields", event.Fields),
	)

	// Extract event type and task ID from event fields
	eventType, ok := event.Fields["event_type"].(string)
	if !ok {
		ao.logger.Warn("Invalid event type in task event")
		return
	}

	taskID, ok := event.Fields["task_id"].(string)
	if !ok {
		ao.logger.Warn("Missing task ID in task event")
		return
	}

	// Process different event types
	switch eventType {
	case "task_created":
		ao.logger.Info("Task created", zap.String("task_id", taskID))
		
	case "task_completed":
		ao.logger.Info("Task completed", zap.String("task_id", taskID))
		
	case "task_failed":
		ao.logger.Warn("Task failed", zap.String("task_id", taskID))
		
	case "task_cancelled":
		ao.logger.Info("Task cancelled", zap.String("task_id", taskID))
		
	default:
		ao.logger.Debug("Unknown task event type", zap.String("event_type", eventType))
	}

	// Record task event metric
	if err := ao.timeSeriesClient.RecordMetric(context.Background(), 
		"task_events", 
		1.0, 
		map[string]string{
			"event_type": eventType,
			"task_id":    taskID,
		}); err != nil {
		ao.logger.Error("Failed to record task event metric", zap.Error(err))
	}
}
