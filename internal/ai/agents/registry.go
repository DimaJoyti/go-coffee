package agents

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Agent defines the interface that all AI agents must implement
type Agent interface {
	// Basic agent information
	GetID() string
	GetType() string
	GetCapabilities() []string
	GetStatus() AgentStatus

	// Health and lifecycle
	IsHealthy() bool
	Start(ctx context.Context) error
	Stop() error

	// Action execution
	ExecuteAction(ctx context.Context, action string, inputs map[string]interface{}) (map[string]interface{}, error)
	
	// Communication
	ReceiveMessage(ctx context.Context, message *AgentMessage) error
	SendMessage(ctx context.Context, message *AgentMessage) error
}

// AgentStatus represents the status of an agent
type AgentStatus string

const (
	AgentStatusOnline  AgentStatus = "online"
	AgentStatusOffline AgentStatus = "offline"
	AgentStatusBusy    AgentStatus = "busy"
	AgentStatusError   AgentStatus = "error"
)

// AgentMessage represents a message between agents
type AgentMessage struct {
	ID        string                 `json:"id"`
	FromAgent string                 `json:"from_agent"`
	ToAgent   string                 `json:"to_agent"`
	Type      string                 `json:"type"`
	Content   map[string]interface{} `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
}

// Registry manages all registered AI agents
type Registry struct {
	agents map[string]Agent
	mutex  sync.RWMutex
	logger *zap.Logger
}

// NewRegistry creates a new agent registry
func NewRegistry(logger *zap.Logger) *Registry {
	return &Registry{
		agents: make(map[string]Agent),
		logger: logger,
	}
}

// Register registers an agent with the registry
func (r *Registry) Register(agentID string, agent Agent) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.agents[agentID]; exists {
		return fmt.Errorf("agent already registered: %s", agentID)
	}

	r.agents[agentID] = agent
	r.logger.Info("Agent registered",
		zap.String("agent_id", agentID),
		zap.String("agent_type", agent.GetType()),
	)

	return nil
}

// Unregister removes an agent from the registry
func (r *Registry) Unregister(agentID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	// Stop the agent
	if err := agent.Stop(); err != nil {
		r.logger.Warn("Error stopping agent during unregistration",
			zap.String("agent_id", agentID),
			zap.Error(err),
		)
	}

	delete(r.agents, agentID)
	r.logger.Info("Agent unregistered", zap.String("agent_id", agentID))

	return nil
}

// GetAgent returns an agent by ID
func (r *Registry) GetAgent(agentID string) (Agent, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	return agent, nil
}

// ListAgents returns all registered agents
func (r *Registry) ListAgents() []Agent {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	agents := make([]Agent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}

	return agents
}

// GetAgentsByType returns all agents of a specific type
func (r *Registry) GetAgentsByType(agentType string) []Agent {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var agents []Agent
	for _, agent := range r.agents {
		if agent.GetType() == agentType {
			agents = append(agents, agent)
		}
	}

	return agents
}

// GetHealthyAgents returns all healthy agents
func (r *Registry) GetHealthyAgents() []Agent {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var agents []Agent
	for _, agent := range r.agents {
		if agent.IsHealthy() {
			agents = append(agents, agent)
		}
	}

	return agents
}

// StartAllAgents starts all registered agents
func (r *Registry) StartAllAgents(ctx context.Context) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for agentID, agent := range r.agents {
		if err := agent.Start(ctx); err != nil {
			r.logger.Error("Failed to start agent",
				zap.String("agent_id", agentID),
				zap.Error(err),
			)
			return fmt.Errorf("failed to start agent %s: %w", agentID, err)
		}
	}

	r.logger.Info("All agents started successfully")
	return nil
}

// StopAllAgents stops all registered agents
func (r *Registry) StopAllAgents() error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var errors []error
	for agentID, agent := range r.agents {
		if err := agent.Stop(); err != nil {
			r.logger.Error("Failed to stop agent",
				zap.String("agent_id", agentID),
				zap.Error(err),
			)
			errors = append(errors, fmt.Errorf("failed to stop agent %s: %w", agentID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors stopping agents: %v", errors)
	}

	r.logger.Info("All agents stopped successfully")
	return nil
}

// BroadcastMessage sends a message to all agents
func (r *Registry) BroadcastMessage(ctx context.Context, message *AgentMessage) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var errors []error
	for agentID, agent := range r.agents {
		if err := agent.ReceiveMessage(ctx, message); err != nil {
			r.logger.Error("Failed to deliver broadcast message",
				zap.String("agent_id", agentID),
				zap.Error(err),
			)
			errors = append(errors, fmt.Errorf("failed to deliver to agent %s: %w", agentID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors broadcasting message: %v", errors)
	}

	return nil
}

// GetAgentStats returns statistics about registered agents
func (r *Registry) GetAgentStats() map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_agents":   len(r.agents),
		"healthy_agents": 0,
		"online_agents":  0,
		"offline_agents": 0,
		"busy_agents":    0,
		"error_agents":   0,
		"agent_types":    make(map[string]int),
	}

	agentTypes := make(map[string]int)
	for _, agent := range r.agents {
		// Count by status
		switch agent.GetStatus() {
		case AgentStatusOnline:
			stats["online_agents"] = stats["online_agents"].(int) + 1
		case AgentStatusOffline:
			stats["offline_agents"] = stats["offline_agents"].(int) + 1
		case AgentStatusBusy:
			stats["busy_agents"] = stats["busy_agents"].(int) + 1
		case AgentStatusError:
			stats["error_agents"] = stats["error_agents"].(int) + 1
		}

		// Count healthy agents
		if agent.IsHealthy() {
			stats["healthy_agents"] = stats["healthy_agents"].(int) + 1
		}

		// Count by type
		agentType := agent.GetType()
		agentTypes[agentType]++
	}

	stats["agent_types"] = agentTypes
	return stats
}

// FindAgentByCapability finds agents that have a specific capability
func (r *Registry) FindAgentByCapability(capability string) []Agent {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var agents []Agent
	for _, agent := range r.agents {
		capabilities := agent.GetCapabilities()
		for _, cap := range capabilities {
			if cap == capability {
				agents = append(agents, agent)
				break
			}
		}
	}

	return agents
}

// GetBestAgentForAction finds the best agent to handle a specific action
func (r *Registry) GetBestAgentForAction(action string) (Agent, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Simple strategy: find the first healthy agent that has the capability
	for _, agent := range r.agents {
		if !agent.IsHealthy() {
			continue
		}

		capabilities := agent.GetCapabilities()
		for _, cap := range capabilities {
			if cap == action {
				return agent, nil
			}
		}
	}

	return nil, fmt.Errorf("no suitable agent found for action: %s", action)
}

// ValidateAgent validates that an agent implements the required interface correctly
func (r *Registry) ValidateAgent(agent Agent) error {
	if agent.GetID() == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	if agent.GetType() == "" {
		return fmt.Errorf("agent type cannot be empty")
	}

	capabilities := agent.GetCapabilities()
	if len(capabilities) == 0 {
		return fmt.Errorf("agent must have at least one capability")
	}

	return nil
}
