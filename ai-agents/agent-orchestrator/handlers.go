package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	eventsourcing "github.com/DimaJoyti/go-coffee/pkg/event-sourcing"
	redisadvanced "github.com/DimaJoyti/go-coffee/pkg/redis-advanced"
)

// Task management handlers

func (ao *AgentOrchestrator) createTask(c *gin.Context) {
	var taskReq TaskRequest
	if err := c.ShouldBindJSON(&taskReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task request", "details": err.Error()})
		return
	}

	// Validate task request
	if taskReq.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task type is required"})
		return
	}

	// Generate task ID if not provided
	if taskReq.ID == "" {
		taskReq.ID = fmt.Sprintf("task_%d", time.Now().UnixNano())
	}

	// Find suitable agents
	suitableAgents := ao.findSuitableAgents(taskReq.RequiredCapabilities)
	if len(suitableAgents) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No suitable agents found"})
		return
	}

	// Select best agent (simple round-robin for now)
	selectedAgent := suitableAgents[0]

	// Create task response
	taskResponse := &TaskResponse{
		TaskID:    taskReq.ID,
		AgentID:   selectedAgent.ID,
		Status:    "assigned",
		Timestamp: time.Now(),
	}

	// Store task in Redis JSON
	if err := ao.jsonClient.Set(c.Request.Context(), 
		fmt.Sprintf("task:%s", taskReq.ID), 
		taskReq, 
		24*time.Hour); err != nil {
		ao.logger.Error("Failed to store task", zap.Error(err))
	}

	// Record task creation event
	event := &eventsourcing.Event{
		AggregateID:   taskReq.ID,
		AggregateType: "task",
		EventType:     "task_created",
		EventVersion:  1,
		Data: map[string]interface{}{
			"task_id":     taskReq.ID,
			"task_type":   taskReq.Type,
			"agent_id":    selectedAgent.ID,
			"priority":    taskReq.Priority,
		},
	}
	
	if err := ao.eventStore.AppendEvent(c.Request.Context(), event); err != nil {
		ao.logger.Error("Failed to record task creation event", zap.Error(err))
	}

	ao.logger.Info("Task created successfully", 
		zap.String("task_id", taskReq.ID),
		zap.String("assigned_agent", selectedAgent.ID),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Task created successfully",
		"task":    taskReq,
		"response": taskResponse,
	})
}

func (ao *AgentOrchestrator) getTask(c *gin.Context) {
	taskID := c.Param("id")
	
	// Get task from Redis
	task, err := ao.jsonClient.Get(c.Request.Context(), fmt.Sprintf("task:%s", taskID), "$")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (ao *AgentOrchestrator) listTasks(c *gin.Context) {
	// Get all tasks from Redis
	tasks, err := ao.jsonClient.Search(c.Request.Context(), "task:*", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

func (ao *AgentOrchestrator) cancelTask(c *gin.Context) {
	taskID := c.Param("id")
	
	// Update task status
	updates := map[string]interface{}{
		"status": "cancelled",
		"cancelled_at": time.Now(),
	}
	
	if err := ao.jsonClient.UpdateDocument(c.Request.Context(), taskID, updates); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Record cancellation event
	event := &eventsourcing.Event{
		AggregateID:   taskID,
		AggregateType: "task",
		EventType:     "task_cancelled",
		EventVersion:  2,
		Data: map[string]interface{}{
			"task_id": taskID,
			"cancelled_at": time.Now(),
		},
	}
	
	if err := ao.eventStore.AppendEvent(c.Request.Context(), event); err != nil {
		ao.logger.Error("Failed to record task cancellation event", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task cancelled successfully"})
}

// Agent discovery handlers

func (ao *AgentOrchestrator) searchAgents(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	// Search agents using Redis Search
	searchReq := &redisadvanced.VectorSearchRequest{
		TopK: 10,
	}

	results, err := ao.searchClient.SemanticSearch(c.Request.Context(), "agent_index", query, 10)
	if err != nil {
		ao.logger.Error("Agent search failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results.Results,
		"total":   results.Total,
		"duration": results.Duration.String(),
	})
}

func (ao *AgentOrchestrator) getCapabilities(c *gin.Context) {
	capabilities := make(map[string][]string)
	
	for _, agent := range ao.agents {
		for _, capability := range agent.Capabilities {
			if _, exists := capabilities[capability]; !exists {
				capabilities[capability] = []string{}
			}
			capabilities[capability] = append(capabilities[capability], agent.ID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"capabilities": capabilities,
		"total_agents": len(ao.agents),
	})
}

func (ao *AgentOrchestrator) matchAgents(c *gin.Context) {
	var matchReq struct {
		RequiredCapabilities []string `json:"required_capabilities"`
		PreferredType        string   `json:"preferred_type,omitempty"`
		MaxAgents           int      `json:"max_agents,omitempty"`
	}

	if err := c.ShouldBindJSON(&matchReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid match request"})
		return
	}

	if matchReq.MaxAgents == 0 {
		matchReq.MaxAgents = 10
	}

	// Find matching agents
	matchingAgents := ao.findSuitableAgents(matchReq.RequiredCapabilities)
	
	// Filter by preferred type if specified
	if matchReq.PreferredType != "" {
		filtered := []*AgentInfo{}
		for _, agent := range matchingAgents {
			if agent.Type == matchReq.PreferredType {
				filtered = append(filtered, agent)
			}
		}
		matchingAgents = filtered
	}

	// Limit results
	if len(matchingAgents) > matchReq.MaxAgents {
		matchingAgents = matchingAgents[:matchReq.MaxAgents]
	}

	c.JSON(http.StatusOK, gin.H{
		"matching_agents": matchingAgents,
		"count":          len(matchingAgents),
		"criteria":       matchReq,
	})
}

// Monitoring handlers

func (ao *AgentOrchestrator) getSystemHealth(c *gin.Context) {
	activeAgents := 0
	healthyAgents := 0
	
	for _, agent := range ao.agents {
		if agent.Status == "active" {
			activeAgents++
		}
		if agent.Health != nil && agent.Health.Status == "healthy" {
			healthyAgents++
		}
	}

	health := map[string]interface{}{
		"status":         "healthy",
		"total_agents":   len(ao.agents),
		"active_agents":  activeAgents,
		"healthy_agents": healthyAgents,
		"timestamp":      time.Now(),
		"uptime":         time.Since(time.Now().Add(-time.Hour)), // Placeholder
	}

	// Determine overall health
	if activeAgents == 0 {
		health["status"] = "critical"
	} else if float64(healthyAgents)/float64(activeAgents) < 0.8 {
		health["status"] = "degraded"
	}

	c.JSON(http.StatusOK, health)
}

func (ao *AgentOrchestrator) getMetrics(c *gin.Context) {
	// Get metrics from Redis TimeSeries
	metrics := map[string]interface{}{
		"agent_registrations": 0,
		"task_completions":    0,
		"error_rate":         0.0,
		"avg_response_time":  0.0,
	}

	// This would query actual metrics from Redis TimeSeries
	// For now, return placeholder data
	c.JSON(http.StatusOK, metrics)
}

func (ao *AgentOrchestrator) getStats(c *gin.Context) {
	stats := map[string]interface{}{
		"agents": map[string]interface{}{
			"total":   len(ao.agents),
			"by_type": ao.getAgentsByType(),
			"by_status": ao.getAgentsByStatus(),
		},
		"circuit_breaker": ao.circuitBreaker.GetStats(),
		"rate_limiter":    ao.rateLimiter.GetStats(),
		"timestamp":       time.Now(),
	}

	c.JSON(http.StatusOK, stats)
}

// Event sourcing handlers

func (ao *AgentOrchestrator) getEventStream(c *gin.Context) {
	aggregateType := c.Param("aggregate_type")
	aggregateID := c.Param("aggregate_id")
	
	fromVersionStr := c.Query("from_version")
	fromVersion := int64(0)
	if fromVersionStr != "" {
		if v, err := strconv.ParseInt(fromVersionStr, 10, 64); err == nil {
			fromVersion = v
		}
	}

	events, err := ao.eventStore.GetEvents(c.Request.Context(), aggregateType, aggregateID, fromVersion)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event stream not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"aggregate_type": aggregateType,
		"aggregate_id":   aggregateID,
		"events":         events,
		"count":          len(events),
		"from_version":   fromVersion,
	})
}

func (ao *AgentOrchestrator) replayEvents(c *gin.Context) {
	var replayReq struct {
		AggregateType string `json:"aggregate_type"`
		AggregateID   string `json:"aggregate_id"`
		FromVersion   int64  `json:"from_version"`
	}

	if err := c.ShouldBindJSON(&replayReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid replay request"})
		return
	}

	// Define event handler for replay
	handler := func(ctx context.Context, event *eventsourcing.Event) error {
		ao.logger.Info("Replaying event", 
			zap.String("event_id", event.ID),
			zap.String("event_type", event.EventType),
		)
		return nil
	}

	err := ao.eventStore.ReplayEvents(c.Request.Context(), 
		replayReq.AggregateType, 
		replayReq.AggregateID, 
		replayReq.FromVersion, 
		handler)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Replay failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Events replayed successfully"})
}

// WebSocket handler

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

func (ao *AgentOrchestrator) handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		ao.logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}
	defer conn.Close()

	ao.logger.Info("WebSocket connection established")

	// Send periodic updates
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send system status
			status := map[string]interface{}{
				"type":         "status_update",
				"total_agents": len(ao.agents),
				"timestamp":    time.Now(),
			}

			if err := conn.WriteJSON(status); err != nil {
				ao.logger.Error("WebSocket write failed", zap.Error(err))
				return
			}
		}
	}
}
