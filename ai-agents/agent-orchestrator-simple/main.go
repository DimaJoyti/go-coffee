package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// AgentInfo represents information about a registered agent
type AgentInfo struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Version      string                 `json:"version"`
	Endpoint     string                 `json:"endpoint"`
	Status       string                 `json:"status"`
	Capabilities []string               `json:"capabilities"`
	Metadata     map[string]interface{} `json:"metadata"`
	LastSeen     time.Time              `json:"last_seen"`
	RegisteredAt time.Time              `json:"registered_at"`
}

// TaskRequest represents a task request for agents
type TaskRequest struct {
	ID                   string                 `json:"id"`
	Type                 string                 `json:"type"`
	Priority             int                    `json:"priority"`
	Data                 map[string]interface{} `json:"data"`
	RequiredCapabilities []string               `json:"required_capabilities"`
	Timeout              time.Duration          `json:"timeout"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// TaskResponse represents a task response from agents
type TaskResponse struct {
	TaskID    string                 `json:"task_id"`
	AgentID   string                 `json:"agent_id"`
	Status    string                 `json:"status"`
	Result    map[string]interface{} `json:"result"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
}

// SimpleOrchestrator manages AI agents
type SimpleOrchestrator struct {
	redis  *redis.Client
	router *gin.Engine
	agents map[string]*AgentInfo
	port   string
}

func main() {
	log.Println("🚀 Starting Simple Agent Orchestrator...")

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_URL", "localhost:6379"),
		Password: "",
		DB:       0,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️  Redis connection failed: %v", err)
		log.Println("📝 Continuing without Redis (in-memory mode)")
		redisClient = nil
	} else {
		log.Println("✅ Redis connected successfully")
	}

	// Initialize orchestrator
	orchestrator := &SimpleOrchestrator{
		redis:  redisClient,
		agents: make(map[string]*AgentInfo),
		port:   getEnv("PORT", "8095"),
	}

	// Setup routes
	orchestrator.setupRoutes()

	// Start server
	go func() {
		log.Printf("🌐 Server starting on port %s", orchestrator.port)
		if err := orchestrator.router.Run(":" + orchestrator.port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Start background services
	go orchestrator.startHealthMonitoring()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	log.Println("🎯 Simple Agent Orchestrator is running. Press Ctrl+C to stop.")
	<-c

	log.Println("🛑 Shutting down...")
	log.Println("✅ Shutdown complete")
}

func (so *SimpleOrchestrator) setupRoutes() {
	so.router = gin.Default()

	// Add CORS middleware
	so.router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API routes
	api := so.router.Group("/api/v1")
	{
		// Agent management
		agents := api.Group("/agents")
		{
			agents.POST("/register", so.registerAgent)
			agents.DELETE("/:id", so.unregisterAgent)
			agents.GET("/:id", so.getAgent)
			agents.GET("", so.listAgents)
			agents.POST("/:id/heartbeat", so.agentHeartbeat)
		}

		// Task management
		tasks := api.Group("/tasks")
		{
			tasks.POST("", so.createTask)
			tasks.GET("/:id", so.getTask)
			tasks.GET("", so.listTasks)
		}

		// Discovery
		discovery := api.Group("/discovery")
		{
			discovery.GET("/search", so.searchAgents)
			discovery.GET("/capabilities", so.getCapabilities)
		}

		// Monitoring
		monitoring := api.Group("/monitoring")
		{
			monitoring.GET("/health", so.getHealth)
			monitoring.GET("/stats", so.getStats)
		}
	}

	// Root endpoint
	so.router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "Simple Agent Orchestrator",
			"version": "1.0.0",
			"status":  "running",
			"agents":  len(so.agents),
		})
	})
}

func (so *SimpleOrchestrator) registerAgent(c *gin.Context) {
	var agent AgentInfo
	if err := c.ShouldBindJSON(&agent); err != nil {
		c.JSON(400, gin.H{"error": "Invalid agent info", "details": err.Error()})
		return
	}

	if agent.ID == "" || agent.Name == "" || agent.Type == "" {
		c.JSON(400, gin.H{"error": "Missing required fields: id, name, type"})
		return
	}

	// Set registration time
	agent.RegisteredAt = time.Now()
	agent.LastSeen = time.Now()
	agent.Status = "active"

	// Store agent
	so.agents[agent.ID] = &agent

	// Store in Redis if available
	if so.redis != nil {
		agentJSON, _ := json.Marshal(agent)
		so.redis.Set(context.Background(), fmt.Sprintf("agent:%s", agent.ID), agentJSON, 24*time.Hour)
	}

	log.Printf("✅ Agent registered: %s (%s)", agent.Name, agent.ID)

	c.JSON(201, gin.H{
		"message": "Agent registered successfully",
		"agent":   agent,
	})
}

func (so *SimpleOrchestrator) unregisterAgent(c *gin.Context) {
	agentID := c.Param("id")

	if _, exists := so.agents[agentID]; !exists {
		c.JSON(404, gin.H{"error": "Agent not found"})
		return
	}

	delete(so.agents, agentID)

	if so.redis != nil {
		so.redis.Del(context.Background(), fmt.Sprintf("agent:%s", agentID))
	}

	log.Printf("🗑️  Agent unregistered: %s", agentID)

	c.JSON(200, gin.H{"message": "Agent unregistered successfully"})
}

func (so *SimpleOrchestrator) getAgent(c *gin.Context) {
	agentID := c.Param("id")

	agent, exists := so.agents[agentID]
	if !exists {
		c.JSON(404, gin.H{"error": "Agent not found"})
		return
	}

	c.JSON(200, agent)
}

func (so *SimpleOrchestrator) listAgents(c *gin.Context) {
	agents := make([]*AgentInfo, 0, len(so.agents))
	for _, agent := range so.agents {
		agents = append(agents, agent)
	}

	c.JSON(200, gin.H{
		"agents": agents,
		"count":  len(agents),
	})
}

func (so *SimpleOrchestrator) agentHeartbeat(c *gin.Context) {
	agentID := c.Param("id")

	agent, exists := so.agents[agentID]
	if !exists {
		c.JSON(404, gin.H{"error": "Agent not found"})
		return
	}

	agent.LastSeen = time.Now()
	agent.Status = "active"

	c.JSON(200, gin.H{"message": "Heartbeat received"})
}

func (so *SimpleOrchestrator) createTask(c *gin.Context) {
	var task TaskRequest
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(400, gin.H{"error": "Invalid task request", "details": err.Error()})
		return
	}

	if task.Type == "" {
		c.JSON(400, gin.H{"error": "Task type is required"})
		return
	}

	if task.ID == "" {
		task.ID = fmt.Sprintf("task_%d", time.Now().UnixNano())
	}

	// Find suitable agents
	suitableAgents := so.findSuitableAgents(task.RequiredCapabilities)
	if len(suitableAgents) == 0 {
		c.JSON(404, gin.H{"error": "No suitable agents found"})
		return
	}

	// Select first available agent
	selectedAgent := suitableAgents[0]

	response := &TaskResponse{
		TaskID:    task.ID,
		AgentID:   selectedAgent.ID,
		Status:    "assigned",
		Timestamp: time.Now(),
	}

	log.Printf("📋 Task created: %s assigned to %s", task.ID, selectedAgent.ID)

	c.JSON(201, gin.H{
		"message":  "Task created successfully",
		"task":     task,
		"response": response,
	})
}

func (so *SimpleOrchestrator) getTask(c *gin.Context) {
	taskID := c.Param("id")
	c.JSON(200, gin.H{
		"task_id": taskID,
		"status":  "not_implemented",
		"message": "Task storage not implemented in simple version",
	})
}

func (so *SimpleOrchestrator) listTasks(c *gin.Context) {
	c.JSON(200, gin.H{
		"tasks":   []interface{}{},
		"count":   0,
		"message": "Task storage not implemented in simple version",
	})
}

func (so *SimpleOrchestrator) searchAgents(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(400, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	var results []*AgentInfo
	for _, agent := range so.agents {
		if contains(agent.Name, query) || contains(agent.Type, query) {
			results = append(results, agent)
		}
	}

	c.JSON(200, gin.H{
		"results": results,
		"count":   len(results),
		"query":   query,
	})
}

func (so *SimpleOrchestrator) getCapabilities(c *gin.Context) {
	capabilities := make(map[string][]string)

	for _, agent := range so.agents {
		for _, capability := range agent.Capabilities {
			if _, exists := capabilities[capability]; !exists {
				capabilities[capability] = []string{}
			}
			capabilities[capability] = append(capabilities[capability], agent.ID)
		}
	}

	c.JSON(200, gin.H{
		"capabilities": capabilities,
		"total_agents": len(so.agents),
	})
}

func (so *SimpleOrchestrator) getHealth(c *gin.Context) {
	activeAgents := 0
	for _, agent := range so.agents {
		if agent.Status == "active" && time.Since(agent.LastSeen) < 2*time.Minute {
			activeAgents++
		}
	}

	status := "healthy"
	if len(so.agents) == 0 {
		status = "no_agents"
	} else if activeAgents == 0 {
		status = "critical"
	}

	c.JSON(200, gin.H{
		"status":        status,
		"total_agents":  len(so.agents),
		"active_agents": activeAgents,
		"timestamp":     time.Now(),
		"redis_connected": so.redis != nil,
	})
}

func (so *SimpleOrchestrator) getStats(c *gin.Context) {
	stats := gin.H{
		"agents": gin.H{
			"total":    len(so.agents),
			"by_type":  so.getAgentsByType(),
			"by_status": so.getAgentsByStatus(),
		},
		"system": gin.H{
			"redis_connected": so.redis != nil,
			"uptime":         time.Since(time.Now().Add(-time.Hour)), // Placeholder
		},
		"timestamp": time.Now(),
	}

	c.JSON(200, stats)
}

// Helper methods

func (so *SimpleOrchestrator) findSuitableAgents(requiredCapabilities []string) []*AgentInfo {
	var suitable []*AgentInfo

	for _, agent := range so.agents {
		if agent.Status != "active" {
			continue
		}

		hasAll := true
		for _, required := range requiredCapabilities {
			hasCapability := false
			for _, agentCap := range agent.Capabilities {
				if agentCap == required {
					hasCapability = true
					break
				}
			}
			if !hasCapability {
				hasAll = false
				break
			}
		}

		if hasAll {
			suitable = append(suitable, agent)
		}
	}

	return suitable
}

func (so *SimpleOrchestrator) getAgentsByType() map[string]int {
	byType := make(map[string]int)
	for _, agent := range so.agents {
		byType[agent.Type]++
	}
	return byType
}

func (so *SimpleOrchestrator) getAgentsByStatus() map[string]int {
	byStatus := make(map[string]int)
	for _, agent := range so.agents {
		byStatus[agent.Status]++
	}
	return byStatus
}

func (so *SimpleOrchestrator) startHealthMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		unhealthyCount := 0
		for agentID, agent := range so.agents {
			if time.Since(agent.LastSeen) > 2*time.Minute {
				agent.Status = "unhealthy"
				unhealthyCount++
				log.Printf("⚠️  Agent %s (%s) appears unhealthy", agent.Name, agentID)
			}
		}

		if unhealthyCount > 0 {
			log.Printf("🏥 Health check: %d unhealthy agents", unhealthyCount)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
		 (len(s) > len(substr) &&
		  (s[:len(substr)] == substr ||
		   s[len(s)-len(substr):] == substr)))
}
