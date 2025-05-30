package main

import (
	"time"
)

// AgentInfo represents information about a registered agent
type AgentInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Version     string                 `json:"version"`
	Endpoint    string                 `json:"endpoint"`
	Status      string                 `json:"status"`
	Capabilities []string              `json:"capabilities"`
	Metadata    map[string]interface{} `json:"metadata"`
	LastSeen    time.Time              `json:"last_seen"`
	RegisteredAt time.Time             `json:"registered_at"`
	Health      *AgentHealth           `json:"health"`
}

// AgentHealth represents agent health information
type AgentHealth struct {
	Status      string        `json:"status"`
	LastCheck   time.Time     `json:"last_check"`
	ResponseTime time.Duration `json:"response_time"`
	ErrorCount  int           `json:"error_count"`
	Uptime      time.Duration `json:"uptime"`
}

// TaskRequest represents a task request for agents
type TaskRequest struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Priority    int                    `json:"priority"`
	Data        map[string]interface{} `json:"data"`
	RequiredCapabilities []string      `json:"required_capabilities"`
	Timeout     time.Duration          `json:"timeout"`
	Metadata    map[string]interface{} `json:"metadata"`
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

// OrchestratorConfig contains orchestrator configuration
type OrchestratorConfig struct {
	Port            string        `json:"port"`
	RedisURL        string        `json:"redis_url"`
	LogLevel        string        `json:"log_level"`
	HealthInterval  time.Duration `json:"health_interval"`
	MetricsInterval time.Duration `json:"metrics_interval"`
	MaxAgents       int           `json:"max_agents"`
}
