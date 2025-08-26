// Package agents provides the base agent interface and implementations for Go Coffee AI agents
package agents

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/DimaJoyti/go-coffee/ai-agents/langgraph-integration/pkg/graph"
	"github.com/google/uuid"
)

// AgentConfig represents configuration for an agent
type AgentConfig struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Timeout     time.Duration          `json:"timeout"`
	MaxRetries  int                    `json:"max_retries"`
	Tools       []string               `json:"tools"`
	LLMConfig   map[string]interface{} `json:"llm_config"`
	Config      map[string]interface{} `json:"config"`
}

// AgentExecutionResult represents the result of an agent execution
type AgentExecutionResult struct {
	AgentType             graph.AgentType        `json:"agent_type"`
	Status                graph.WorkflowStatus   `json:"status"`
	Output                map[string]interface{} `json:"output"`
	Error                 string                 `json:"error,omitempty"`
	ExecutionTime         time.Duration          `json:"execution_time"`
	NextAgent             *graph.AgentType       `json:"next_agent,omitempty"`
	HumanApprovalRequired bool                   `json:"human_approval_required"`
	Metadata              map[string]interface{} `json:"metadata"`
}

// Agent represents the interface that all AI agents must implement
type Agent interface {
	// GetType returns the agent type
	GetType() graph.AgentType

	// GetConfig returns the agent configuration
	GetConfig() *AgentConfig

	// Execute executes the agent with the given state
	Execute(ctx context.Context, state *graph.AgentState) (*AgentExecutionResult, error)

	// HealthCheck performs a health check of the agent
	HealthCheck(ctx context.Context) error

	// GetTools returns the list of available tools
	GetTools() []string

	// Initialize initializes the agent
	Initialize(ctx context.Context) error

	// Shutdown gracefully shuts down the agent
	Shutdown(ctx context.Context) error
}

// BaseAgent provides common functionality for all agents
type BaseAgent struct {
	agentType       graph.AgentType
	config          *AgentConfig
	isInitialized   bool
	executionCount  int
	errorCount      int
	lastExecution   *time.Time
	lastError       *string
}

// NewBaseAgent creates a new base agent
func NewBaseAgent(agentType graph.AgentType, config *AgentConfig) *BaseAgent {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.Version == "" {
		config.Version = "1.0.0"
	}

	return &BaseAgent{
		agentType: agentType,
		config:    config,
	}
}

// GetType returns the agent type
func (a *BaseAgent) GetType() graph.AgentType {
	return a.agentType
}

// GetConfig returns the agent configuration
func (a *BaseAgent) GetConfig() *AgentConfig {
	return a.config
}

// Initialize initializes the base agent
func (a *BaseAgent) Initialize(ctx context.Context) error {
	if a.isInitialized {
		return nil
	}

	log.Printf("Initializing agent: %s", a.agentType)
	
	// Perform base initialization
	a.isInitialized = true
	
	log.Printf("Agent %s initialized successfully", a.agentType)
	return nil
}

// Shutdown gracefully shuts down the base agent
func (a *BaseAgent) Shutdown(ctx context.Context) error {
	log.Printf("Shutting down agent: %s", a.agentType)
	
	a.isInitialized = false
	
	log.Printf("Agent %s shutdown complete", a.agentType)
	return nil
}

// HealthCheck performs a basic health check
func (a *BaseAgent) HealthCheck(ctx context.Context) error {
	if !a.isInitialized {
		return fmt.Errorf("agent %s is not initialized", a.agentType)
	}
	return nil
}

// GetTools returns the list of available tools
func (a *BaseAgent) GetTools() []string {
	return a.config.Tools
}

// ExecuteWithMetrics wraps agent execution with metrics and error handling
func (a *BaseAgent) ExecuteWithMetrics(
	ctx context.Context,
	state *graph.AgentState,
	executeFunc func(ctx context.Context, state *graph.AgentState) (*AgentExecutionResult, error),
) (*AgentExecutionResult, error) {
	if !a.isInitialized {
		return nil, fmt.Errorf("agent %s is not initialized", a.agentType)
	}

	start := time.Now()
	
	log.Printf("Executing agent %s for workflow %s", a.agentType, state.WorkflowID)

	// Set current agent in state
	state.SetCurrentAgent(a.agentType)
	defer state.ClearCurrentAgent()

	// Create context with timeout
	execCtx := ctx
	if a.config.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, a.config.Timeout)
		defer cancel()
	}

	// Execute the agent
	result, err := executeFunc(execCtx, state)
	
	executionTime := time.Since(start)
	now := time.Now()
	a.lastExecution = &now
	a.executionCount++

	if err != nil {
		a.errorCount++
		errorMsg := err.Error()
		a.lastError = &errorMsg
		
		log.Printf("Agent %s failed after %v: %s", a.agentType, executionTime, err.Error())
		
		// Create error result
		return &AgentExecutionResult{
			AgentType:     a.agentType,
			Status:        graph.WorkflowStatusFailed,
			Error:         err.Error(),
			ExecutionTime: executionTime,
			Metadata: map[string]interface{}{
				"execution_count": a.executionCount,
				"error_count":     a.errorCount,
			},
		}, err
	}

	// Ensure result has correct metadata
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	
	result.AgentType = a.agentType
	result.ExecutionTime = executionTime
	result.Metadata["execution_count"] = a.executionCount
	result.Metadata["error_count"] = a.errorCount
	result.Metadata["agent_version"] = a.config.Version

	// Update state with agent completion
	state.MarkAgentCompleted(a.agentType)
	state.SetAgentOutput(a.agentType, result.Output)

	// Add success message to state
	state.AddAIMessage(
		fmt.Sprintf("Agent %s completed successfully", a.agentType),
		a.agentType,
	)

	log.Printf("Agent %s completed successfully in %v", a.agentType, executionTime)
	
	return result, nil
}

// GetStats returns execution statistics for the agent
func (a *BaseAgent) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"agent_type":      a.agentType,
		"version":         a.config.Version,
		"is_initialized":  a.isInitialized,
		"execution_count": a.executionCount,
		"error_count":     a.errorCount,
		"error_rate":      float64(a.errorCount) / float64(max(a.executionCount, 1)),
		"tools_count":     len(a.config.Tools),
	}

	if a.lastExecution != nil {
		stats["last_execution"] = a.lastExecution.Format(time.RFC3339)
	}

	if a.lastError != nil {
		stats["last_error"] = *a.lastError
	}

	return stats
}

// ValidateInput validates the input state for agent execution
func (a *BaseAgent) ValidateInput(state *graph.AgentState) error {
	if state == nil {
		return fmt.Errorf("agent state cannot be nil")
	}

	if state.WorkflowID == uuid.Nil {
		return fmt.Errorf("workflow ID is required")
	}

	if state.ExecutionID == uuid.Nil {
		return fmt.Errorf("execution ID is required")
	}

	return nil
}

// CreateSuccessResult creates a successful execution result
func (a *BaseAgent) CreateSuccessResult(output map[string]interface{}) *AgentExecutionResult {
	return &AgentExecutionResult{
		AgentType: a.agentType,
		Status:    graph.WorkflowStatusCompleted,
		Output:    output,
		Metadata:  make(map[string]interface{}),
	}
}

// CreateErrorResult creates an error execution result
func (a *BaseAgent) CreateErrorResult(err error) *AgentExecutionResult {
	return &AgentExecutionResult{
		AgentType: a.agentType,
		Status:    graph.WorkflowStatusFailed,
		Error:     err.Error(),
		Output:    make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
	}
}

// SetNextAgent sets the next agent to execute
func (result *AgentExecutionResult) SetNextAgent(agentType graph.AgentType) {
	result.NextAgent = &agentType
}

// RequireHumanApproval marks that human approval is required
func (result *AgentExecutionResult) RequireHumanApproval() {
	result.HumanApprovalRequired = true
}

// AddMetadata adds metadata to the execution result
func (result *AgentExecutionResult) AddMetadata(key string, value interface{}) {
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata[key] = value
}

// AddOutput adds output data to the execution result
func (result *AgentExecutionResult) AddOutput(key string, value interface{}) {
	if result.Output == nil {
		result.Output = make(map[string]interface{})
	}
	result.Output[key] = value
}

// Helper function to get max of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
