// Package graph provides LangGraph-inspired state management for Go Coffee AI agents
package graph

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/google/uuid"
)

// WorkflowStatus represents the current status of a workflow execution
type WorkflowStatus string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
	WorkflowStatusPaused    WorkflowStatus = "paused"
)

// AgentType represents the different types of agents in the system
type AgentType string

const (
	AgentTypeBeverageInventor   AgentType = "beverage_inventor"
	AgentTypeInventoryManager   AgentType = "inventory_manager"
	AgentTypeSocialMedia        AgentType = "social_media"
	AgentTypeCustomerService    AgentType = "customer_service"
	AgentTypeFeedbackAnalyst    AgentType = "feedback_analyst"
	AgentTypeTaskManager        AgentType = "task_manager"
	AgentTypeNotifier           AgentType = "notifier"
	AgentTypeScheduler          AgentType = "scheduler"
	AgentTypeCoordinator        AgentType = "coordinator"
	AgentTypeTastingCoordinator AgentType = "tasting_coordinator"
)

// Priority represents the priority level of a workflow
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityUrgent   Priority = "urgent"
	PriorityCritical Priority = "critical"
)

// Message represents a message in the agent conversation
type Message struct {
	ID        uuid.UUID      `json:"id"`
	Type      string         `json:"type"` // "human", "ai", "system"
	Content   string         `json:"content"`
	AgentType AgentType      `json:"agent_type,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// NewMessage creates a new message
func NewMessage(msgType, content string, agentType AgentType) *Message {
	return &Message{
		ID:        uuid.New(),
		Type:      msgType,
		Content:   content,
		AgentType: agentType,
		Metadata:  make(map[string]any),
		Timestamp: time.Now().UTC(),
	}
}

// AgentState represents the shared state across all agents in a workflow
type AgentState struct {
	// Core workflow information
	WorkflowID  uuid.UUID      `json:"workflow_id"`
	ExecutionID uuid.UUID      `json:"execution_id"`
	Status      WorkflowStatus `json:"status"`
	Priority    Priority       `json:"priority"`

	// Message history
	Messages []*Message `json:"messages"`

	// Agent execution tracking
	CurrentAgent    *AgentType  `json:"current_agent,omitempty"`
	CompletedAgents []AgentType `json:"completed_agents"`
	FailedAgents    []AgentType `json:"failed_agents"`

	// Workflow data and context
	InputData    map[string]any `json:"input_data"`
	SharedData   map[string]any `json:"shared_data"`
	AgentOutputs map[string]any `json:"agent_outputs"`

	// Business context
	CustomerID *uuid.UUID `json:"customer_id,omitempty"`
	OrderID    *uuid.UUID `json:"order_id,omitempty"`
	LocationID *uuid.UUID `json:"location_id,omitempty"`
	BrandID    *uuid.UUID `json:"brand_id,omitempty"`

	// Workflow control
	NextAgent             *AgentType     `json:"next_agent,omitempty"`
	RoutingConditions     map[string]any `json:"routing_conditions"`
	HumanApprovalRequired bool           `json:"human_approval_required"`
	ApprovalStatus        *string        `json:"approval_status,omitempty"`

	// Error handling and retry
	ErrorCount int     `json:"error_count"`
	LastError  *string `json:"last_error,omitempty"`
	RetryCount int     `json:"retry_count"`
	MaxRetries int     `json:"max_retries"`

	// Timing and performance
	CreatedAt     time.Time  `json:"created_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	ExecutionTime *float64   `json:"execution_time,omitempty"`

	// Metadata and configuration
	Metadata map[string]any `json:"metadata"`
	Config   map[string]any `json:"config"`
}

// NewAgentState creates a new agent state
func NewAgentState(workflowID, executionID uuid.UUID) *AgentState {
	return &AgentState{
		WorkflowID:        workflowID,
		ExecutionID:       executionID,
		Status:            WorkflowStatusPending,
		Priority:          PriorityMedium,
		Messages:          make([]*Message, 0),
		CompletedAgents:   make([]AgentType, 0),
		FailedAgents:      make([]AgentType, 0),
		InputData:         make(map[string]any),
		SharedData:        make(map[string]any),
		AgentOutputs:      make(map[string]any),
		RoutingConditions: make(map[string]any),
		MaxRetries:        3,
		CreatedAt:         time.Now().UTC(),
		Metadata:          make(map[string]any),
		Config:            make(map[string]any),
	}
}

// AddMessage adds a message to the state
func (s *AgentState) AddMessage(message *Message) {
	s.Messages = append(s.Messages, message)
}

// AddHumanMessage adds a human message to the state
func (s *AgentState) AddHumanMessage(content string) {
	message := NewMessage("human", content, "")
	s.AddMessage(message)
}

// AddAIMessage adds an AI message to the state
func (s *AgentState) AddAIMessage(content string, agentType AgentType) {
	message := NewMessage("ai", content, agentType)
	s.AddMessage(message)
}

// AddSystemMessage adds a system message to the state
func (s *AgentState) AddSystemMessage(content string) {
	message := NewMessage("system", content, "")
	s.AddMessage(message)
}

// SetCurrentAgent sets the currently executing agent
func (s *AgentState) SetCurrentAgent(agentType AgentType) {
	s.CurrentAgent = &agentType
}

// ClearCurrentAgent clears the currently executing agent
func (s *AgentState) ClearCurrentAgent() {
	s.CurrentAgent = nil
}

// MarkAgentCompleted marks an agent as completed
func (s *AgentState) MarkAgentCompleted(agentType AgentType) {
	s.CompletedAgents = append(s.CompletedAgents, agentType)
}

// MarkAgentFailed marks an agent as failed
func (s *AgentState) MarkAgentFailed(agentType AgentType) {
	s.FailedAgents = append(s.FailedAgents, agentType)
}

// IsAgentCompleted checks if an agent has completed
func (s *AgentState) IsAgentCompleted(agentType AgentType) bool {
	return slices.Contains(s.CompletedAgents, agentType)
}

// IsAgentFailed checks if an agent has failed
func (s *AgentState) IsAgentFailed(agentType AgentType) bool {
	return slices.Contains(s.FailedAgents, agentType)
}

// SetAgentOutput sets the output for a specific agent
func (s *AgentState) SetAgentOutput(agentType AgentType, output any) {
	s.AgentOutputs[string(agentType)] = output
}

// GetAgentOutput gets the output for a specific agent
func (s *AgentState) GetAgentOutput(agentType AgentType) (any, bool) {
	output, exists := s.AgentOutputs[string(agentType)]
	return output, exists
}

// SetSharedData sets shared data that can be accessed by all agents
func (s *AgentState) SetSharedData(key string, value any) {
	s.SharedData[key] = value
}

// GetSharedData gets shared data
func (s *AgentState) GetSharedData(key string) (any, bool) {
	value, exists := s.SharedData[key]
	return value, exists
}

// SetMetadata sets metadata for the workflow
func (s *AgentState) SetMetadata(key string, value any) {
	s.Metadata[key] = value
}

// GetMetadata gets metadata for the workflow
func (s *AgentState) GetMetadata(key string) (any, bool) {
	value, exists := s.Metadata[key]
	return value, exists
}

// StartExecution marks the workflow as started
func (s *AgentState) StartExecution() {
	now := time.Now().UTC()
	s.StartedAt = &now
	s.Status = WorkflowStatusRunning
}

// CompleteExecution marks the workflow as completed
func (s *AgentState) CompleteExecution() {
	now := time.Now().UTC()
	s.CompletedAt = &now
	s.Status = WorkflowStatusCompleted

	if s.StartedAt != nil {
		duration := now.Sub(*s.StartedAt).Seconds()
		s.ExecutionTime = &duration
	}
}

// FailExecution marks the workflow as failed
func (s *AgentState) FailExecution(errorMsg string) {
	now := time.Now().UTC()
	s.CompletedAt = &now
	s.Status = WorkflowStatusFailed
	s.LastError = &errorMsg

	if s.StartedAt != nil {
		duration := now.Sub(*s.StartedAt).Seconds()
		s.ExecutionTime = &duration
	}
}

// IncrementError increments the error count
func (s *AgentState) IncrementError(errorMsg string) {
	s.ErrorCount++
	s.LastError = &errorMsg
}

// IncrementRetry increments the retry count
func (s *AgentState) IncrementRetry() {
	s.RetryCount++
}

// CanRetry checks if the workflow can be retried
func (s *AgentState) CanRetry() bool {
	return s.RetryCount < s.MaxRetries
}

// Clone creates a deep copy of the agent state
func (s *AgentState) Clone() (*AgentState, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	var clone AgentState
	if err := json.Unmarshal(data, &clone); err != nil {
		return nil, err
	}

	return &clone, nil
}

// ToJSON converts the state to JSON
func (s *AgentState) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// FromJSON creates a state from JSON
func FromJSON(data []byte) (*AgentState, error) {
	var state AgentState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}
