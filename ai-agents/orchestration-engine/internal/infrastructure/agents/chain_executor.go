package agents

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ChainExecutor executes agent chains with advanced orchestration
type ChainExecutor struct {
	agentRegistry    *AgentRegistry
	conditionEngine  *ConditionEngine
	retryManager     *RetryManager
	timeoutManager   *TimeoutManager
	logger           Logger
	mutex            sync.RWMutex
}

// ConditionEngine evaluates conditional logic for agent chains
type ConditionEngine struct {
	evaluators map[string]ConditionEvaluator
	logger     Logger
}

// RetryManager handles retry logic for failed agent operations
type RetryManager struct {
	policies map[string]*RetryPolicy
	logger   Logger
}

// TimeoutManager manages timeouts for agent operations
type TimeoutManager struct {
	timeouts map[string]*TimeoutConfig
	logger   Logger
}

// AgentRegistry maintains a registry of available agents
type AgentRegistry struct {
	agents    map[string]*RegisteredAgent
	instances map[string]*AgentInstance
	mutex     sync.RWMutex
	logger    Logger
}

// ConditionEvaluator interface for condition evaluation
type ConditionEvaluator interface {
	Evaluate(ctx context.Context, condition string, data map[string]interface{}) (bool, error)
	GetType() string
}

// RegisteredAgent represents a registered agent type
type RegisteredAgent struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Operations  []string               `json:"operations"`
	Config      *AgentConfig           `json:"config"`
	Factory     AgentFactory           `json:"-"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// AgentInstance represents an active agent instance
type AgentInstance struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	State       InstanceState          `json:"state"`
	CurrentOp   string                 `json:"current_operation"`
	StartTime   time.Time              `json:"start_time"`
	LastUsed    time.Time              `json:"last_used"`
	UsageCount  int64                  `json:"usage_count"`
	Performance *InstancePerformance   `json:"performance"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AgentFactory creates agent instances
type AgentFactory interface {
	CreateAgent(config *AgentConfig) (Agent, error)
	GetAgentType() string
}

// Agent interface for all agents
type Agent interface {
	Execute(ctx context.Context, operation string, parameters map[string]interface{}) (*AgentResult, error)
	GetType() string
	GetCapabilities() []string
	Health(ctx context.Context) error
	Stop(ctx context.Context) error
}

// ChainExecutionContext contains execution context for a chain
type ChainExecutionContext struct {
	ChainID     string                 `json:"chain_id"`
	ExecutionID string                 `json:"execution_id"`
	StartTime   time.Time              `json:"start_time"`
	CurrentStep int                    `json:"current_step"`
	StepResults map[string]interface{} `json:"step_results"`
	Variables   map[string]interface{} `json:"variables"`
	Metadata    map[string]interface{} `json:"metadata"`
	Cancelled   bool                   `json:"cancelled"`
}

// StepExecutionResult represents the result of executing a chain step
type StepExecutionResult struct {
	StepID      string                 `json:"step_id"`
	Success     bool                   `json:"success"`
	Result      interface{}            `json:"result"`
	Error       error                  `json:"error"`
	Duration    time.Duration          `json:"duration"`
	RetryCount  int                    `json:"retry_count"`
	Metadata    map[string]interface{} `json:"metadata"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
}

// AgentResult represents the result of an agent operation
type AgentResult struct {
	Success   bool                   `json:"success"`
	Data      interface{}            `json:"data"`
	Error     error                  `json:"error"`
	Duration  time.Duration          `json:"duration"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// Supporting types
type InstanceState string
const (
	InstanceStateIdle    InstanceState = "idle"
	InstanceStateBusy    InstanceState = "busy"
	InstanceStateError   InstanceState = "error"
	InstanceStateStopped InstanceState = "stopped"
)

type AgentConfig struct {
	Type        string                 `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Timeout     time.Duration          `json:"timeout"`
	RetryPolicy *RetryPolicy           `json:"retry_policy"`
	Resources   *ResourceConfig        `json:"resources"`
}

type ResourceConfig struct {
	CPULimit    string `json:"cpu_limit"`
	MemoryLimit string `json:"memory_limit"`
	Timeout     time.Duration `json:"timeout"`
}

type InstancePerformance struct {
	AvgResponseTime time.Duration `json:"avg_response_time"`
	SuccessRate     float64       `json:"success_rate"`
	ErrorRate       float64       `json:"error_rate"`
	TotalRequests   int64         `json:"total_requests"`
	LastUpdated     time.Time     `json:"last_updated"`
}

type TimeoutConfig struct {
	Default   time.Duration `json:"default"`
	Operation map[string]time.Duration `json:"operation"`
	Global    time.Duration `json:"global"`
}

type RetryPolicy struct {
	MaxAttempts   int           `json:"max_attempts"`
	BackoffType   string        `json:"backoff_type"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	Multiplier    float64       `json:"multiplier"`
	RetryableErrors []string    `json:"retryable_errors"`
}

// Logger interface
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// NewChainExecutor creates a new chain executor
func NewChainExecutor(logger Logger) *ChainExecutor {
	return &ChainExecutor{
		agentRegistry:   NewAgentRegistry(logger),
		conditionEngine: NewConditionEngine(logger),
		retryManager:    NewRetryManager(logger),
		timeoutManager:  NewTimeoutManager(logger),
		logger:          logger,
	}
}

// ExecuteChain executes an agent chain
func (ce *ChainExecutor) ExecuteChain(ctx context.Context, chain *AgentChain, parameters map[string]interface{}) (*ChainExecutionResult, error) {
	ce.logger.Info("Starting chain execution", "chain_id", chain.ID, "name", chain.Name)

	// Create execution context
	execCtx := &ChainExecutionContext{
		ChainID:     chain.ID,
		ExecutionID: fmt.Sprintf("exec_%d", time.Now().UnixNano()),
		StartTime:   time.Now(),
		CurrentStep: 0,
		StepResults: make(map[string]interface{}),
		Variables:   parameters,
		Metadata:    make(map[string]interface{}),
	}

	// Execute chain steps
	results := make([]*StepExecutionResult, 0, len(chain.Steps))
	
	for i, step := range chain.Steps {
		execCtx.CurrentStep = i
		
		// Check if execution was cancelled
		if execCtx.Cancelled {
			break
		}
		
		// Check step dependencies
		if !ce.checkStepDependencies(step, execCtx) {
			ce.logger.Warn("Step dependencies not met", "step_id", step.ID, "chain_id", chain.ID)
			continue
		}
		
		// Evaluate step conditions
		if !ce.evaluateStepConditions(ctx, step, execCtx) {
			ce.logger.Info("Step conditions not met, skipping", "step_id", step.ID, "chain_id", chain.ID)
			continue
		}
		
		// Execute step
		stepResult, err := ce.executeStep(ctx, step, execCtx)
		if err != nil {
			ce.logger.Error("Step execution failed", err, "step_id", step.ID, "chain_id", chain.ID)
			
			// Handle step failure based on chain config
			if chain.Config != nil && chain.Config.FailurePolicy == "stop" {
				return &ChainExecutionResult{
					ChainID:     chain.ID,
					ExecutionID: execCtx.ExecutionID,
					Success:     false,
					Error:       err,
					Results:     results,
					StartTime:   execCtx.StartTime,
					EndTime:     time.Now(),
				}, err
			}
		}
		
		results = append(results, stepResult)
		
		// Store step result for future steps
		execCtx.StepResults[step.ID] = stepResult.Result
		
		// Execute step actions based on result
		if stepResult.Success && step.OnSuccess != nil {
			ce.executeStepAction(ctx, step.OnSuccess, execCtx)
		} else if !stepResult.Success && step.OnFailure != nil {
			ce.executeStepAction(ctx, step.OnFailure, execCtx)
		}
	}

	// Create final result
	success := true
	for _, result := range results {
		if !result.Success {
			success = false
			break
		}
	}

	chainResult := &ChainExecutionResult{
		ChainID:     chain.ID,
		ExecutionID: execCtx.ExecutionID,
		Success:     success,
		Results:     results,
		StartTime:   execCtx.StartTime,
		EndTime:     time.Now(),
		Metadata:    execCtx.Metadata,
	}

	ce.logger.Info("Chain execution completed", 
		"chain_id", chain.ID, 
		"execution_id", execCtx.ExecutionID,
		"success", success,
		"duration", chainResult.EndTime.Sub(chainResult.StartTime),
	)

	return chainResult, nil
}

// executeStep executes a single chain step
func (ce *ChainExecutor) executeStep(ctx context.Context, step *ChainStep, execCtx *ChainExecutionContext) (*StepExecutionResult, error) {
	startTime := time.Now()
	
	ce.logger.Debug("Executing step", "step_id", step.ID, "agent_type", step.AgentType, "operation", step.Operation)

	// Get agent instance
	agent, err := ce.agentRegistry.GetAgent(step.AgentType)
	if err != nil {
		return &StepExecutionResult{
			StepID:    step.ID,
			Success:   false,
			Error:     err,
			StartTime: startTime,
			EndTime:   time.Now(),
		}, err
	}

	// Prepare parameters
	parameters := ce.prepareStepParameters(step, execCtx)

	// Execute with retry logic
	var result *AgentResult
	var execErr error
	
	retryPolicy := step.RetryPolicy
	if retryPolicy == nil {
		retryPolicy = ce.retryManager.GetDefaultPolicy()
	}

	for attempt := 0; attempt < retryPolicy.MaxAttempts; attempt++ {
		// Create timeout context
		stepCtx := ctx
		if step.Timeout > 0 {
			var cancel context.CancelFunc
			stepCtx, cancel = context.WithTimeout(ctx, step.Timeout)
			defer cancel()
		}

		// Execute agent operation
		result, execErr = agent.Execute(stepCtx, step.Operation, parameters)
		
		if execErr == nil {
			break
		}

		// Check if error is retryable
		if !ce.retryManager.IsRetryableError(execErr, retryPolicy) {
			break
		}

		// Wait before retry
		if attempt < retryPolicy.MaxAttempts-1 {
			delay := ce.retryManager.CalculateDelay(attempt, retryPolicy)
			time.Sleep(delay)
		}
	}

	endTime := time.Now()
	success := execErr == nil && result != nil && result.Success

	stepResult := &StepExecutionResult{
		StepID:     step.ID,
		Success:    success,
		Result:     result,
		Error:      execErr,
		Duration:   endTime.Sub(startTime),
		RetryCount: 0, // Would track actual retry count
		StartTime:  startTime,
		EndTime:    endTime,
		Metadata:   make(map[string]interface{}),
	}

	if result != nil {
		stepResult.Result = result.Data
		stepResult.Metadata = result.Metadata
	}

	return stepResult, execErr
}

// checkStepDependencies checks if step dependencies are satisfied
func (ce *ChainExecutor) checkStepDependencies(step *ChainStep, execCtx *ChainExecutionContext) bool {
	for _, depID := range step.Dependencies {
		if _, exists := execCtx.StepResults[depID]; !exists {
			return false
		}
	}
	return true
}

// evaluateStepConditions evaluates step conditions
func (ce *ChainExecutor) evaluateStepConditions(ctx context.Context, step *ChainStep, execCtx *ChainExecutionContext) bool {
	if len(step.Conditions) == 0 {
		return true
	}

	for _, condition := range step.Conditions {
		result, err := ce.conditionEngine.EvaluateCondition(ctx, condition, execCtx.Variables)
		if err != nil {
			ce.logger.Error("Failed to evaluate condition", err, "step_id", step.ID)
			return false
		}
		
		if !result {
			return false
		}
	}

	return true
}

// prepareStepParameters prepares parameters for step execution
func (ce *ChainExecutor) prepareStepParameters(step *ChainStep, execCtx *ChainExecutionContext) map[string]interface{} {
	parameters := make(map[string]interface{})

	// Copy step parameters
	for k, v := range step.Parameters {
		parameters[k] = v
	}

	// Add execution context variables
	for k, v := range execCtx.Variables {
		parameters[k] = v
	}

	// Add step results from dependencies
	for _, depID := range step.Dependencies {
		if result, exists := execCtx.StepResults[depID]; exists {
			parameters[fmt.Sprintf("dep_%s", depID)] = result
		}
	}

	return parameters
}

// executeStepAction executes a step action
func (ce *ChainExecutor) executeStepAction(ctx context.Context, action *StepAction, execCtx *ChainExecutionContext) {
	ce.logger.Debug("Executing step action", "type", action.Type, "target", action.Target)

	switch action.Type {
	case "set_variable":
		if key, ok := action.Parameters["key"].(string); ok {
			if value, exists := action.Parameters["value"]; exists {
				execCtx.Variables[key] = value
			}
		}
	case "log":
		if message, ok := action.Parameters["message"].(string); ok {
			ce.logger.Info("Step action log", "message", message, "chain_id", execCtx.ChainID)
		}
	case "cancel":
		execCtx.Cancelled = true
	default:
		ce.logger.Warn("Unknown step action type", "type", action.Type)
	}
}

// NewAgentRegistry creates a new agent registry
func NewAgentRegistry(logger Logger) *AgentRegistry {
	return &AgentRegistry{
		agents:    make(map[string]*RegisteredAgent),
		instances: make(map[string]*AgentInstance),
		logger:    logger,
	}
}

// RegisterAgent registers a new agent type
func (ar *AgentRegistry) RegisterAgent(agent *RegisteredAgent) error {
	ar.mutex.Lock()
	defer ar.mutex.Unlock()

	ar.agents[agent.Type] = agent
	ar.logger.Info("Agent registered", "type", agent.Type, "name", agent.Name)

	return nil
}

// GetAgent gets an agent instance
func (ar *AgentRegistry) GetAgent(agentType string) (Agent, error) {
	ar.mutex.RLock()
	registeredAgent, exists := ar.agents[agentType]
	ar.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("agent type %s not registered", agentType)
	}

	// Create new agent instance using factory
	agent, err := registeredAgent.Factory.CreateAgent(registeredAgent.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent instance: %w", err)
	}

	return agent, nil
}

// NewConditionEngine creates a new condition engine
func NewConditionEngine(logger Logger) *ConditionEngine {
	ce := &ConditionEngine{
		evaluators: make(map[string]ConditionEvaluator),
		logger:     logger,
	}

	// Register default evaluators
	ce.RegisterEvaluator(&SimpleConditionEvaluator{})
	ce.RegisterEvaluator(&JSONPathEvaluator{})

	return ce
}

// RegisterEvaluator registers a condition evaluator
func (ce *ConditionEngine) RegisterEvaluator(evaluator ConditionEvaluator) {
	ce.evaluators[evaluator.GetType()] = evaluator
}

// EvaluateCondition evaluates a condition
func (ce *ConditionEngine) EvaluateCondition(ctx context.Context, condition *StepCondition, data map[string]interface{}) (bool, error) {
	evaluator, exists := ce.evaluators[condition.Type]
	if !exists {
		return false, fmt.Errorf("condition evaluator %s not found", condition.Type)
	}

	conditionStr := fmt.Sprintf("%s %s %v", condition.Field, condition.Operator, condition.Value)
	return evaluator.Evaluate(ctx, conditionStr, data)
}

// SimpleConditionEvaluator implements basic condition evaluation
type SimpleConditionEvaluator struct{}

func (sce *SimpleConditionEvaluator) Evaluate(ctx context.Context, condition string, data map[string]interface{}) (bool, error) {
	// Simplified condition evaluation
	// In a real implementation, this would parse and evaluate complex conditions
	return true, nil
}

func (sce *SimpleConditionEvaluator) GetType() string {
	return "simple"
}

// JSONPathEvaluator implements JSONPath-based condition evaluation
type JSONPathEvaluator struct{}

func (jpe *JSONPathEvaluator) Evaluate(ctx context.Context, condition string, data map[string]interface{}) (bool, error) {
	// Simplified JSONPath evaluation
	// In a real implementation, this would use a JSONPath library
	return true, nil
}

func (jpe *JSONPathEvaluator) GetType() string {
	return "jsonpath"
}

// NewRetryManager creates a new retry manager
func NewRetryManager(logger Logger) *RetryManager {
	return &RetryManager{
		policies: make(map[string]*RetryPolicy),
		logger:   logger,
	}
}

// GetDefaultPolicy returns the default retry policy
func (rm *RetryManager) GetDefaultPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:  3,
		BackoffType:  "exponential",
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// IsRetryableError checks if an error is retryable
func (rm *RetryManager) IsRetryableError(err error, policy *RetryPolicy) bool {
	// Simplified implementation - would check against retryable error patterns
	return true
}

// CalculateDelay calculates retry delay
func (rm *RetryManager) CalculateDelay(attempt int, policy *RetryPolicy) time.Duration {
	switch policy.BackoffType {
	case "exponential":
		delay := time.Duration(float64(policy.InitialDelay) * pow(policy.Multiplier, float64(attempt)))
		if delay > policy.MaxDelay {
			return policy.MaxDelay
		}
		return delay
	case "linear":
		delay := policy.InitialDelay * time.Duration(attempt+1)
		if delay > policy.MaxDelay {
			return policy.MaxDelay
		}
		return delay
	default:
		return policy.InitialDelay
	}
}

// NewTimeoutManager creates a new timeout manager
func NewTimeoutManager(logger Logger) *TimeoutManager {
	return &TimeoutManager{
		timeouts: make(map[string]*TimeoutConfig),
		logger:   logger,
	}
}

// Helper function for power calculation
func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// ChainExecutionResult represents the result of chain execution
type ChainExecutionResult struct {
	ChainID     string                   `json:"chain_id"`
	ExecutionID string                   `json:"execution_id"`
	Success     bool                     `json:"success"`
	Error       error                    `json:"error"`
	Results     []*StepExecutionResult   `json:"results"`
	StartTime   time.Time                `json:"start_time"`
	EndTime     time.Time                `json:"end_time"`
	Duration    time.Duration            `json:"duration"`
	Metadata    map[string]interface{}   `json:"metadata"`
}

// Import types from services package
type AgentChain struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Steps       []*ChainStep           `json:"steps"`
	Config      *ChainConfig           `json:"config"`
	State       ChainState             `json:"state"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type ChainStep struct {
	ID            string                 `json:"id"`
	AgentType     string                 `json:"agent_type"`
	Operation     string                 `json:"operation"`
	Parameters    map[string]interface{} `json:"parameters"`
	Conditions    []*StepCondition       `json:"conditions"`
	Timeout       time.Duration          `json:"timeout"`
	RetryPolicy   *RetryPolicy           `json:"retry_policy"`
	OnSuccess     *StepAction            `json:"on_success"`
	OnFailure     *StepAction            `json:"on_failure"`
	Dependencies  []string               `json:"dependencies"`
	Metadata      map[string]interface{} `json:"metadata"`
}

type StepCondition struct {
	Type      string      `json:"type"`
	Field     string      `json:"field"`
	Operator  string      `json:"operator"`
	Value     interface{} `json:"value"`
	LogicOp   string      `json:"logic_op"`
}

type StepAction struct {
	Type       string                 `json:"type"`
	Target     string                 `json:"target"`
	Parameters map[string]interface{} `json:"parameters"`
}

type ChainConfig struct {
	MaxRetries      int           `json:"max_retries"`
	Timeout         time.Duration `json:"timeout"`
	FailurePolicy   string        `json:"failure_policy"`
	ParallelSteps   bool          `json:"parallel_steps"`
	EnableMetrics   bool          `json:"enable_metrics"`
	EnableTracing   bool          `json:"enable_tracing"`
}

type ChainState string
const (
	ChainStateIdle       ChainState = "idle"
	ChainStateRunning    ChainState = "running"
	ChainStateCompleted  ChainState = "completed"
	ChainStateFailed     ChainState = "failed"
	ChainStatePaused     ChainState = "paused"
)
