package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
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

// MinimalOrchestrator manages AI agents with minimal dependencies
type MinimalOrchestrator struct {
	agents map[string]*AgentInfo
	port   string
}

func main() {
	log.Println("🚀 Starting Minimal Agent Orchestrator...")

	orchestrator := &MinimalOrchestrator{
		agents: make(map[string]*AgentInfo),
		port:   getEnv("PORT", "8095"),
	}

	// Setup HTTP routes
	http.HandleFunc("/", orchestrator.handleRoot)
	http.HandleFunc("/api/v1/agents/register", orchestrator.handleRegisterAgent)
	http.HandleFunc("/api/v1/agents/", orchestrator.handleAgents)
	http.HandleFunc("/api/v1/agents", orchestrator.handleListAgents)
	http.HandleFunc("/api/v1/tasks", orchestrator.handleTasks)
	http.HandleFunc("/api/v1/discovery/search", orchestrator.handleSearch)
	http.HandleFunc("/api/v1/monitoring/health", orchestrator.handleHealth)
	http.HandleFunc("/api/v1/monitoring/stats", orchestrator.handleStats)

	// Start health monitoring
	go orchestrator.startHealthMonitoring()

	log.Printf("🌐 Server starting on port %s", orchestrator.port)
	log.Printf("📋 Available endpoints:")
	log.Printf("  GET  / - Service info")
	log.Printf("  POST /api/v1/agents/register - Register agent")
	log.Printf("  GET  /api/v1/agents - List agents")
	log.Printf("  POST /api/v1/tasks - Create task")
	log.Printf("  GET  /api/v1/discovery/search?q=<query> - Search agents")
	log.Printf("  GET  /api/v1/monitoring/health - Health check")
	log.Printf("  GET  /api/v1/monitoring/stats - Statistics")

	if err := http.ListenAndServe(":"+orchestrator.port, nil); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

func (mo *MinimalOrchestrator) handleRoot(w http.ResponseWriter, r *http.Request) {
	mo.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	response := map[string]interface{}{
		"service": "Minimal Agent Orchestrator",
		"version": "1.0.0",
		"status":  "running",
		"agents":  len(mo.agents),
		"endpoints": []string{
			"POST /api/v1/agents/register",
			"GET /api/v1/agents",
			"POST /api/v1/tasks",
			"GET /api/v1/discovery/search",
			"GET /api/v1/monitoring/health",
			"GET /api/v1/monitoring/stats",
		},
	}

	mo.writeJSONResponse(w, 200, response)
}

func (mo *MinimalOrchestrator) handleRegisterAgent(w http.ResponseWriter, r *http.Request) {
	mo.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		mo.writeJSONResponse(w, 405, map[string]string{"error": "Method not allowed"})
		return
	}

	var agent AgentInfo
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		mo.writeJSONResponse(w, 400, map[string]string{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	if agent.ID == "" || agent.Name == "" || agent.Type == "" {
		mo.writeJSONResponse(w, 400, map[string]string{"error": "Missing required fields: id, name, type"})
		return
	}

	// Set registration time
	agent.RegisteredAt = time.Now()
	agent.LastSeen = time.Now()
	agent.Status = "active"

	// Store agent
	mo.agents[agent.ID] = &agent

	log.Printf("✅ Agent registered: %s (%s)", agent.Name, agent.ID)

	response := map[string]interface{}{
		"message": "Agent registered successfully",
		"agent":   agent,
	}

	mo.writeJSONResponse(w, 201, response)
}

func (mo *MinimalOrchestrator) handleAgents(w http.ResponseWriter, r *http.Request) {
	mo.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	// Extract agent ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/agents/")
	if path == "" {
		mo.handleListAgents(w, r)
		return
	}

	// Handle specific agent operations
	agentID := strings.Split(path, "/")[0]

	switch r.Method {
	case "GET":
		agent, exists := mo.agents[agentID]
		if !exists {
			mo.writeJSONResponse(w, 404, map[string]string{"error": "Agent not found"})
			return
		}
		mo.writeJSONResponse(w, 200, agent)

	case "DELETE":
		if _, exists := mo.agents[agentID]; !exists {
			mo.writeJSONResponse(w, 404, map[string]string{"error": "Agent not found"})
			return
		}
		delete(mo.agents, agentID)
		log.Printf("🗑️  Agent unregistered: %s", agentID)
		mo.writeJSONResponse(w, 200, map[string]string{"message": "Agent unregistered successfully"})

	case "POST":
		// Handle heartbeat
		if strings.HasSuffix(path, "/heartbeat") {
			agent, exists := mo.agents[agentID]
			if !exists {
				mo.writeJSONResponse(w, 404, map[string]string{"error": "Agent not found"})
				return
			}
			agent.LastSeen = time.Now()
			agent.Status = "active"
			mo.writeJSONResponse(w, 200, map[string]string{"message": "Heartbeat received"})
		} else {
			mo.writeJSONResponse(w, 404, map[string]string{"error": "Endpoint not found"})
		}

	default:
		mo.writeJSONResponse(w, 405, map[string]string{"error": "Method not allowed"})
	}
}

func (mo *MinimalOrchestrator) handleListAgents(w http.ResponseWriter, r *http.Request) {
	mo.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "GET" {
		mo.writeJSONResponse(w, 405, map[string]string{"error": "Method not allowed"})
		return
	}

	agents := make([]*AgentInfo, 0, len(mo.agents))
	for _, agent := range mo.agents {
		agents = append(agents, agent)
	}

	response := map[string]interface{}{
		"agents": agents,
		"count":  len(agents),
	}

	mo.writeJSONResponse(w, 200, response)
}

func (mo *MinimalOrchestrator) handleTasks(w http.ResponseWriter, r *http.Request) {
	mo.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		mo.writeJSONResponse(w, 405, map[string]string{"error": "Method not allowed"})
		return
	}

	var task TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		mo.writeJSONResponse(w, 400, map[string]string{"error": "Invalid JSON", "details": err.Error()})
		return
	}

	if task.Type == "" {
		mo.writeJSONResponse(w, 400, map[string]string{"error": "Task type is required"})
		return
	}

	if task.ID == "" {
		task.ID = fmt.Sprintf("task_%d", time.Now().UnixNano())
	}

	// Find suitable agents
	suitableAgents := mo.findSuitableAgents(task.RequiredCapabilities)
	if len(suitableAgents) == 0 {
		mo.writeJSONResponse(w, 404, map[string]string{"error": "No suitable agents found"})
		return
	}

	// Select first available agent
	selectedAgent := suitableAgents[0]

	response := map[string]interface{}{
		"message": "Task created successfully",
		"task_id": task.ID,
		"assigned_agent": map[string]interface{}{
			"id":   selectedAgent.ID,
			"name": selectedAgent.Name,
			"type": selectedAgent.Type,
		},
		"status":    "assigned",
		"timestamp": time.Now(),
	}

	log.Printf("📋 Task created: %s assigned to %s", task.ID, selectedAgent.ID)

	mo.writeJSONResponse(w, 201, response)
}

func (mo *MinimalOrchestrator) handleSearch(w http.ResponseWriter, r *http.Request) {
	mo.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "GET" {
		mo.writeJSONResponse(w, 405, map[string]string{"error": "Method not allowed"})
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		mo.writeJSONResponse(w, 400, map[string]string{"error": "Query parameter 'q' is required"})
		return
	}

	var results []*AgentInfo
	for _, agent := range mo.agents {
		if strings.Contains(strings.ToLower(agent.Name), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(agent.Type), strings.ToLower(query)) {
			results = append(results, agent)
		}
	}

	response := map[string]interface{}{
		"results": results,
		"count":   len(results),
		"query":   query,
	}

	mo.writeJSONResponse(w, 200, response)
}

func (mo *MinimalOrchestrator) handleHealth(w http.ResponseWriter, r *http.Request) {
	mo.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "GET" {
		mo.writeJSONResponse(w, 405, map[string]string{"error": "Method not allowed"})
		return
	}

	activeAgents := 0
	for _, agent := range mo.agents {
		if agent.Status == "active" && time.Since(agent.LastSeen) < 2*time.Minute {
			activeAgents++
		}
	}

	status := "healthy"
	if len(mo.agents) == 0 {
		status = "no_agents"
	} else if activeAgents == 0 {
		status = "critical"
	}

	response := map[string]interface{}{
		"status":        status,
		"total_agents":  len(mo.agents),
		"active_agents": activeAgents,
		"timestamp":     time.Now(),
	}

	mo.writeJSONResponse(w, 200, response)
}

func (mo *MinimalOrchestrator) handleStats(w http.ResponseWriter, r *http.Request) {
	mo.setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "GET" {
		mo.writeJSONResponse(w, 405, map[string]string{"error": "Method not allowed"})
		return
	}

	response := map[string]interface{}{
		"agents": map[string]interface{}{
			"total":     len(mo.agents),
			"by_type":   mo.getAgentsByType(),
			"by_status": mo.getAgentsByStatus(),
		},
		"system": map[string]interface{}{
			"version": "1.0.0",
			"uptime":  "running", // Placeholder
		},
		"timestamp": time.Now(),
	}

	mo.writeJSONResponse(w, 200, response)
}

// Helper methods

func (mo *MinimalOrchestrator) findSuitableAgents(requiredCapabilities []string) []*AgentInfo {
	var suitable []*AgentInfo

	for _, agent := range mo.agents {
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

func (mo *MinimalOrchestrator) getAgentsByType() map[string]int {
	byType := make(map[string]int)
	for _, agent := range mo.agents {
		byType[agent.Type]++
	}
	return byType
}

func (mo *MinimalOrchestrator) getAgentsByStatus() map[string]int {
	byStatus := make(map[string]int)
	for _, agent := range mo.agents {
		byStatus[agent.Status]++
	}
	return byStatus
}

func (mo *MinimalOrchestrator) startHealthMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		unhealthyCount := 0
		for agentID, agent := range mo.agents {
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

func (mo *MinimalOrchestrator) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func (mo *MinimalOrchestrator) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
