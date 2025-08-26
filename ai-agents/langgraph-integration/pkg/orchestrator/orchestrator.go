// Package orchestrator provides LangGraph-inspired orchestration for Go Coffee AI agents
package orchestrator

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/ai-agents/langgraph-integration/pkg/agents"
	"github.com/DimaJoyti/go-coffee/ai-agents/langgraph-integration/pkg/graph"
	"github.com/google/uuid"
)

// LangGraphOrchestrator orchestrates AI agents using LangGraph-inspired patterns
type LangGraphOrchestrator struct {
	// Core components
	agentRegistry map[graph.AgentType]agents.Agent
	graphs        map[string]*graph.Graph
	
	// State management
	activeExecutions map[uuid.UUID]*ExecutionContext
	stateStore       StateStore
	
	// Configuration
	config *OrchestratorConfig
	
	// Synchronization
	mu sync.RWMutex
	
	// Metrics and monitoring
	executionCount int
	successCount   int
	errorCount     int
	lastExecution  *time.Time
}

// OrchestratorConfig represents configuration for the orchestrator
type OrchestratorConfig struct {
	MaxConcurrentExecutions int           `json:"max_concurrent_executions"`
	DefaultTimeout          time.Duration `json:"default_timeout"`
	RetryPolicy             RetryPolicy   `json:"retry_policy"`
	StateStorageConfig      map[string]interface{} `json:"state_storage_config"`
	MonitoringEnabled       bool          `json:"monitoring_enabled"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries      int           `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	MaxDelay        time.Duration `json:"max_delay"`
}

// ExecutionContext represents the context of a workflow execution
type ExecutionContext struct {
	WorkflowID    uuid.UUID           `json:"workflow_id"`
	ExecutionID   uuid.UUID           `json:"execution_id"`
	GraphID       string              `json:"graph_id"`
	State         *graph.AgentState   `json:"state"`
	StartTime     time.Time           `json:"start_time"`
	LastUpdate    time.Time           `json:"last_update"`
	Status        graph.WorkflowStatus `json:"status"`
	CurrentNode   string              `json:"current_node"`
	ExecutedNodes []string            `json:"executed_nodes"`
	Context       context.Context     `json:"-"`
	Cancel        context.CancelFunc  `json:"-"`
}

// StateStore interface for state persistence
type StateStore interface {
	SaveState(ctx context.Context, executionID uuid.UUID, state *graph.AgentState) error
	LoadState(ctx context.Context, executionID uuid.UUID) (*graph.AgentState, error)
	DeleteState(ctx context.Context, executionID uuid.UUID) error
	ListActiveExecutions(ctx context.Context) ([]uuid.UUID, error)
}

// MemoryStateStore is an in-memory implementation of StateStore
type MemoryStateStore struct {
	states map[uuid.UUID]*graph.AgentState
	mu     sync.RWMutex
}

// NewMemoryStateStore creates a new in-memory state store
func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{
		states: make(map[uuid.UUID]*graph.AgentState),
	}
}

// SaveState saves state to memory
func (s *MemoryStateStore) SaveState(ctx context.Context, executionID uuid.UUID, state *graph.AgentState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Clone state to avoid mutations
	clonedState, err := state.Clone()
	if err != nil {
		return fmt.Errorf("failed to clone state: %w", err)
	}
	
	s.states[executionID] = clonedState
	return nil
}

// LoadState loads state from memory
func (s *MemoryStateStore) LoadState(ctx context.Context, executionID uuid.UUID) (*graph.AgentState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	state, exists := s.states[executionID]
	if !exists {
		return nil, fmt.Errorf("state not found for execution %s", executionID)
	}
	
	// Clone state to avoid mutations
	return state.Clone()
}

// DeleteState deletes state from memory
func (s *MemoryStateStore) DeleteState(ctx context.Context, executionID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.states, executionID)
	return nil
}

// ListActiveExecutions lists all active executions
func (s *MemoryStateStore) ListActiveExecutions(ctx context.Context) ([]uuid.UUID, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	executions := make([]uuid.UUID, 0, len(s.states))
	for id := range s.states {
		executions = append(executions, id)
	}
	
	return executions, nil
}

// NewLangGraphOrchestrator creates a new LangGraph orchestrator
func NewLangGraphOrchestrator(config *OrchestratorConfig) *LangGraphOrchestrator {
	if config == nil {
		config = &OrchestratorConfig{
			MaxConcurrentExecutions: 100,
			DefaultTimeout:          30 * time.Minute,
			RetryPolicy: RetryPolicy{
				MaxRetries:    3,
				InitialDelay:  1 * time.Second,
				BackoffFactor: 2.0,
				MaxDelay:      30 * time.Second,
			},
			MonitoringEnabled: true,
		}
	}

	return &LangGraphOrchestrator{
		agentRegistry:    make(map[graph.AgentType]agents.Agent),
		graphs:           make(map[string]*graph.Graph),
		activeExecutions: make(map[uuid.UUID]*ExecutionContext),
		stateStore:       NewMemoryStateStore(),
		config:           config,
	}
}

// RegisterAgent registers an agent with the orchestrator
func (o *LangGraphOrchestrator) RegisterAgent(agentType graph.AgentType, agent agents.Agent) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if _, exists := o.agentRegistry[agentType]; exists {
		return fmt.Errorf("agent %s is already registered", agentType)
	}

	// Initialize the agent
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := agent.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize agent %s: %w", agentType, err)
	}

	o.agentRegistry[agentType] = agent
	log.Printf("Registered agent: %s", agentType)
	
	return nil
}

// RegisterGraph registers a workflow graph with the orchestrator
func (o *LangGraphOrchestrator) RegisterGraph(graphID string, workflowGraph *graph.Graph) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if _, exists := o.graphs[graphID]; exists {
		return fmt.Errorf("graph %s is already registered", graphID)
	}

	// Compile the graph
	if err := workflowGraph.Compile(); err != nil {
		return fmt.Errorf("failed to compile graph %s: %w", graphID, err)
	}

	o.graphs[graphID] = workflowGraph
	log.Printf("Registered graph: %s", graphID)
	
	return nil
}

// ExecuteWorkflow executes a workflow using the specified graph
func (o *LangGraphOrchestrator) ExecuteWorkflow(
	ctx context.Context,
	graphID string,
	initialState *graph.AgentState,
) (*graph.AgentState, error) {
	// Get the graph
	o.mu.RLock()
	workflowGraph, exists := o.graphs[graphID]
	o.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("graph %s not found", graphID)
	}

	// Check concurrent execution limit
	o.mu.RLock()
	activeCount := len(o.activeExecutions)
	o.mu.RUnlock()

	if activeCount >= o.config.MaxConcurrentExecutions {
		return nil, fmt.Errorf("maximum concurrent executions reached (%d)", o.config.MaxConcurrentExecutions)
	}

	// Create execution context
	execCtx, cancel := context.WithTimeout(ctx, o.config.DefaultTimeout)
	executionContext := &ExecutionContext{
		WorkflowID:    initialState.WorkflowID,
		ExecutionID:   initialState.ExecutionID,
		GraphID:       graphID,
		State:         initialState,
		StartTime:     time.Now().UTC(),
		LastUpdate:    time.Now().UTC(),
		Status:        graph.WorkflowStatusRunning,
		ExecutedNodes: make([]string, 0),
		Context:       execCtx,
		Cancel:        cancel,
	}

	// Register execution
	o.mu.Lock()
	o.activeExecutions[initialState.ExecutionID] = executionContext
	o.executionCount++
	now := time.Now()
	o.lastExecution = &now
	o.mu.Unlock()

	// Ensure cleanup
	defer func() {
		cancel()
		o.mu.Lock()
		delete(o.activeExecutions, initialState.ExecutionID)
		o.mu.Unlock()
	}()

	log.Printf("Starting workflow execution %s using graph %s", 
		initialState.ExecutionID, graphID)

	// Execute the graph
	finalState, err := o.executeGraphWithAgents(execCtx, workflowGraph, initialState)
	
	// Update metrics
	o.mu.Lock()
	if err != nil {
		o.errorCount++
	} else {
		o.successCount++
	}
	o.mu.Unlock()

	// Save final state
	if err := o.stateStore.SaveState(ctx, initialState.ExecutionID, finalState); err != nil {
		log.Printf("Warning: Failed to save final state: %v", err)
	}

	if err != nil {
		log.Printf("Workflow execution %s failed: %v", initialState.ExecutionID, err)
		return finalState, err
	}

	log.Printf("Workflow execution %s completed successfully", initialState.ExecutionID)
	return finalState, nil
}

// executeGraphWithAgents executes the graph with agent integration
func (o *LangGraphOrchestrator) executeGraphWithAgents(
	ctx context.Context,
	workflowGraph *graph.Graph,
	initialState *graph.AgentState,
) (*graph.AgentState, error) {
	// Create agent execution nodes
	agentNodes := o.createAgentNodes()
	
	// Add agent nodes to graph
	for nodeID, nodeFunc := range agentNodes {
		node := &graph.Node{
			ID:          nodeID,
			Type:        graph.NodeTypeAgent,
			Name:        fmt.Sprintf("Agent: %s", nodeID),
			Description: fmt.Sprintf("Execute %s agent", nodeID),
			Function:    nodeFunc,
			Timeout:     30 * time.Second,
			Retries:     3,
			Metadata:    make(map[string]interface{}),
		}
		
		if err := workflowGraph.AddNode(node); err != nil {
			// Node might already exist, which is fine
			log.Printf("Note: Node %s already exists in graph", nodeID)
		}
	}

	// Recompile graph with agent nodes
	if err := workflowGraph.Compile(); err != nil {
		return nil, fmt.Errorf("failed to recompile graph with agent nodes: %w", err)
	}

	// Execute the graph
	return workflowGraph.Execute(ctx, initialState)
}

// createAgentNodes creates node functions for all registered agents
func (o *LangGraphOrchestrator) createAgentNodes() map[string]graph.NodeFunc {
	o.mu.RLock()
	defer o.mu.RUnlock()

	nodes := make(map[string]graph.NodeFunc)
	
	for agentType, agent := range o.agentRegistry {
		// Create a closure to capture the agent
		agentInstance := agent
		agentTypeVal := agentType
		
		nodeFunc := func(ctx context.Context, state *graph.AgentState) (*graph.AgentState, error) {
			// Execute the agent
			result, err := agentInstance.Execute(ctx, state)
			if err != nil {
				state.MarkAgentFailed(agentTypeVal)
				state.IncrementError(err.Error())
				return state, err
			}

			// Update state with agent result
			state.MarkAgentCompleted(agentTypeVal)
			state.SetAgentOutput(agentTypeVal, result.Output)

			// Handle next agent routing
			if result.NextAgent != nil {
				state.NextAgent = result.NextAgent
			}

			// Handle human approval requirement
			if result.HumanApprovalRequired {
				state.HumanApprovalRequired = true
			}

			return state, nil
		}
		
		nodes[string(agentType)] = nodeFunc
	}

	return nodes
}

// GetExecutionStatus returns the status of a workflow execution
func (o *LangGraphOrchestrator) GetExecutionStatus(executionID uuid.UUID) (*ExecutionContext, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	execution, exists := o.activeExecutions[executionID]
	if !exists {
		return nil, fmt.Errorf("execution %s not found", executionID)
	}

	return execution, nil
}

// ListActiveExecutions returns all active executions
func (o *LangGraphOrchestrator) ListActiveExecutions() []*ExecutionContext {
	o.mu.RLock()
	defer o.mu.RUnlock()

	executions := make([]*ExecutionContext, 0, len(o.activeExecutions))
	for _, execution := range o.activeExecutions {
		executions = append(executions, execution)
	}

	return executions
}

// CancelExecution cancels a running workflow execution
func (o *LangGraphOrchestrator) CancelExecution(executionID uuid.UUID) error {
	o.mu.RLock()
	execution, exists := o.activeExecutions[executionID]
	o.mu.RUnlock()

	if !exists {
		return fmt.Errorf("execution %s not found", executionID)
	}

	execution.Cancel()
	execution.Status = graph.WorkflowStatusCancelled
	
	log.Printf("Cancelled workflow execution %s", executionID)
	return nil
}

// GetStats returns orchestrator statistics
func (o *LangGraphOrchestrator) GetStats() map[string]interface{} {
	o.mu.RLock()
	defer o.mu.RUnlock()

	stats := map[string]interface{}{
		"registered_agents":        len(o.agentRegistry),
		"registered_graphs":        len(o.graphs),
		"active_executions":        len(o.activeExecutions),
		"total_executions":         o.executionCount,
		"successful_executions":    o.successCount,
		"failed_executions":        o.errorCount,
		"success_rate":             float64(o.successCount) / float64(max(o.executionCount, 1)),
		"max_concurrent_executions": o.config.MaxConcurrentExecutions,
	}

	if o.lastExecution != nil {
		stats["last_execution"] = o.lastExecution.Format(time.RFC3339)
	}

	return stats
}

// HealthCheck performs a health check of the orchestrator and all agents
func (o *LangGraphOrchestrator) HealthCheck(ctx context.Context) error {
	o.mu.RLock()
	agents := make([]agents.Agent, 0, len(o.agentRegistry))
	for _, agent := range o.agentRegistry {
		agents = append(agents, agent)
	}
	o.mu.RUnlock()

	// Check all agents
	for _, agent := range agents {
		if err := agent.HealthCheck(ctx); err != nil {
			return fmt.Errorf("agent %s health check failed: %w", agent.GetType(), err)
		}
	}

	return nil
}

// Shutdown gracefully shuts down the orchestrator
func (o *LangGraphOrchestrator) Shutdown(ctx context.Context) error {
	log.Printf("Shutting down LangGraph orchestrator...")

	// Cancel all active executions
	o.mu.RLock()
	executions := make([]*ExecutionContext, 0, len(o.activeExecutions))
	for _, execution := range o.activeExecutions {
		executions = append(executions, execution)
	}
	o.mu.RUnlock()

	for _, execution := range executions {
		execution.Cancel()
	}

	// Shutdown all agents
	o.mu.RLock()
	agents := make([]agents.Agent, 0, len(o.agentRegistry))
	for _, agent := range o.agentRegistry {
		agents = append(agents, agent)
	}
	o.mu.RUnlock()

	for _, agent := range agents {
		if err := agent.Shutdown(ctx); err != nil {
			log.Printf("Warning: Failed to shutdown agent %s: %v", agent.GetType(), err)
		}
	}

	log.Printf("LangGraph orchestrator shutdown complete")
	return nil
}

// Helper function to get max of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
