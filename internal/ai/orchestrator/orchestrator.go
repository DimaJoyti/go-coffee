package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/ai/agents"
	"github.com/DimaJoyti/go-coffee/internal/ai/messaging"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Orchestrator manages AI agents and coordinates their interactions
type Orchestrator struct {
	agentRegistry *agents.Registry
	kafkaManager  *messaging.KafkaManager
	logger        *zap.Logger
	config        config.OrchestratorConfig

	// State management
	mutex     sync.RWMutex
	running   bool
	workflows map[string]*Workflow
	tasks     map[string]*Task

	// Communication channels
	agentMessages chan *AgentMessage
	taskQueue     chan *Task
	stopChan      chan struct{}
}

// Workflow represents an AI workflow
type Workflow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Steps       []*WorkflowStep        `json:"steps"`
	Status      WorkflowStatus         `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowStep represents a step in a workflow
type WorkflowStep struct {
	ID       string                 `json:"id"`
	AgentID  string                 `json:"agent_id"`
	Action   string                 `json:"action"`
	Inputs   map[string]interface{} `json:"inputs"`
	Outputs  map[string]interface{} `json:"outputs,omitempty"`
	Status   StepStatus             `json:"status"`
	Duration time.Duration          `json:"duration,omitempty"`
}

// Task represents an AI task
type Task struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflow_id,omitempty"`
	AgentID     string                 `json:"agent_id"`
	Action      string                 `json:"action"`
	Inputs      map[string]interface{} `json:"inputs"`
	Outputs     map[string]interface{} `json:"outputs,omitempty"`
	Status      TaskStatus             `json:"status"`
	Priority    TaskPriority           `json:"priority"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// AgentMessage represents a message between agents
type AgentMessage = agents.AgentMessage

// Status enums
type WorkflowStatus string
type StepStatus string
type TaskStatus string
type TaskPriority string
type MessageType string

const (
	WorkflowStatusPending   WorkflowStatus = "pending"
	WorkflowStatusRunning   WorkflowStatus = "running"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"

	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"

	TaskStatusPending   TaskStatus = "pending"
	TaskStatusAssigned  TaskStatus = "assigned"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"

	TaskPriorityLow      TaskPriority = "low"
	TaskPriorityMedium   TaskPriority = "medium"
	TaskPriorityHigh     TaskPriority = "high"
	TaskPriorityCritical TaskPriority = "critical"

	MessageTypeRequest      MessageType = "request"
	MessageTypeResponse     MessageType = "response"
	MessageTypeNotification MessageType = "notification"
	MessageTypeBroadcast    MessageType = "broadcast"
)

// NewOrchestrator creates a new AI orchestrator
func NewOrchestrator(
	agentRegistry *agents.Registry,
	kafkaManager *messaging.KafkaManager,
	logger *zap.Logger,
	config config.OrchestratorConfig,
) *Orchestrator {
	return &Orchestrator{
		agentRegistry: agentRegistry,
		kafkaManager:  kafkaManager,
		logger:        logger,
		config:        config,
		workflows:     make(map[string]*Workflow),
		tasks:         make(map[string]*Task),
		agentMessages: make(chan *AgentMessage, 1000),
		taskQueue:     make(chan *Task, 1000),
		stopChan:      make(chan struct{}),
	}
}

// Start starts the orchestrator
func (o *Orchestrator) Start(ctx context.Context) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.running {
		return fmt.Errorf("orchestrator is already running")
	}

	o.logger.Info("Starting AI orchestrator...")

	// Start Kafka manager
	if err := o.kafkaManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start Kafka manager: %w", err)
	}

	// Start background workers
	go o.messageProcessor(ctx)
	go o.taskProcessor(ctx)
	go o.workflowMonitor(ctx)
	go o.healthMonitor(ctx)

	// Initialize predefined workflows
	if err := o.initializePredefinedWorkflows(); err != nil {
		o.logger.Error("Failed to initialize predefined workflows", zap.Error(err))
	}

	o.running = true
	o.logger.Info("AI orchestrator started successfully")

	return nil
}

// Stop stops the orchestrator
func (o *Orchestrator) Stop() {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if !o.running {
		return
	}

	o.logger.Info("Stopping AI orchestrator...")

	// Signal stop to all workers
	close(o.stopChan)

	// Stop Kafka manager
	o.kafkaManager.Stop()

	o.running = false
	o.logger.Info("AI orchestrator stopped")
}

// CreateWorkflow creates a new workflow
func (o *Orchestrator) CreateWorkflow(workflow *Workflow) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if workflow.ID == "" {
		workflow.ID = uuid.New().String()
	}

	workflow.Status = WorkflowStatusPending
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()

	o.workflows[workflow.ID] = workflow

	o.logger.Info("Workflow created",
		zap.String("workflow_id", workflow.ID),
		zap.String("name", workflow.Name),
	)

	return nil
}

// ExecuteWorkflow executes a workflow
func (o *Orchestrator) ExecuteWorkflow(workflowID string) error {
	o.mutex.Lock()
	workflow, exists := o.workflows[workflowID]
	if !exists {
		o.mutex.Unlock()
		return fmt.Errorf("workflow not found: %s", workflowID)
	}
	o.mutex.Unlock()

	o.logger.Info("Executing workflow",
		zap.String("workflow_id", workflowID),
		zap.String("name", workflow.Name),
	)

	// Update workflow status
	workflow.Status = WorkflowStatusRunning
	workflow.UpdatedAt = time.Now()

	// Execute workflow steps
	go o.executeWorkflowSteps(workflow)

	return nil
}

// executeWorkflowSteps executes the steps of a workflow
func (o *Orchestrator) executeWorkflowSteps(workflow *Workflow) {
	for _, step := range workflow.Steps {
		if err := o.executeWorkflowStep(workflow, step); err != nil {
			o.logger.Error("Workflow step failed",
				zap.String("workflow_id", workflow.ID),
				zap.String("step_id", step.ID),
				zap.Error(err),
			)
			workflow.Status = WorkflowStatusFailed
			return
		}
	}

	workflow.Status = WorkflowStatusCompleted
	workflow.UpdatedAt = time.Now()

	o.logger.Info("Workflow completed",
		zap.String("workflow_id", workflow.ID),
		zap.String("name", workflow.Name),
	)
}

// executeWorkflowStep executes a single workflow step
func (o *Orchestrator) executeWorkflowStep(workflow *Workflow, step *WorkflowStep) error {
	step.Status = StepStatusRunning
	startTime := time.Now()

	// Get agent
	agent, err := o.agentRegistry.GetAgent(step.AgentID)
	if err != nil {
		step.Status = StepStatusFailed
		return fmt.Errorf("agent not found: %s", step.AgentID)
	}

	// Execute action
	outputs, err := agent.ExecuteAction(context.Background(), step.Action, step.Inputs)
	if err != nil {
		step.Status = StepStatusFailed
		return fmt.Errorf("action execution failed: %w", err)
	}

	step.Outputs = outputs
	step.Status = StepStatusCompleted
	step.Duration = time.Since(startTime)

	o.logger.Info("Workflow step completed",
		zap.String("workflow_id", workflow.ID),
		zap.String("step_id", step.ID),
		zap.String("agent_id", step.AgentID),
		zap.String("action", step.Action),
		zap.Duration("duration", step.Duration),
	)

	return nil
}

// AssignTask assigns a task to an agent
func (o *Orchestrator) AssignTask(task *Task) error {
	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	task.Status = TaskStatusAssigned
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	o.mutex.Lock()
	o.tasks[task.ID] = task
	o.mutex.Unlock()

	// Add to task queue
	select {
	case o.taskQueue <- task:
		o.logger.Info("Task assigned",
			zap.String("task_id", task.ID),
			zap.String("agent_id", task.AgentID),
			zap.String("action", task.Action),
		)
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}

// SendMessage sends a message between agents
func (o *Orchestrator) SendMessage(message *AgentMessage) error {
	if message.ID == "" {
		message.ID = uuid.New().String()
	}
	message.Timestamp = time.Now()

	// Add to message queue
	select {
	case o.agentMessages <- message:
		o.logger.Info("Message sent",
			zap.String("message_id", message.ID),
			zap.String("from_agent", message.FromAgent),
			zap.String("to_agent", message.ToAgent),
			zap.String("type", string(message.Type)),
		)
		return nil
	default:
		return fmt.Errorf("message queue is full")
	}
}

// messageProcessor processes agent messages
func (o *Orchestrator) messageProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-o.stopChan:
			return
		case message := <-o.agentMessages:
			o.processMessage(message)
		}
	}
}

// taskProcessor processes tasks
func (o *Orchestrator) taskProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-o.stopChan:
			return
		case task := <-o.taskQueue:
			o.processTask(task)
		}
	}
}

// processMessage processes an agent message
func (o *Orchestrator) processMessage(message *AgentMessage) {
	// Publish message to Kafka
	if err := o.kafkaManager.PublishAgentMessage(context.Background(), message); err != nil {
		o.logger.Error("Failed to publish agent message", zap.Error(err))
	}

	// Route message to target agent
	if message.ToAgent != "" {
		agent, err := o.agentRegistry.GetAgent(message.ToAgent)
		if err != nil {
			o.logger.Error("Target agent not found",
				zap.String("agent_id", message.ToAgent),
				zap.Error(err),
			)
			return
		}

		if err := agent.ReceiveMessage(context.Background(), message); err != nil {
			o.logger.Error("Failed to deliver message to agent",
				zap.String("agent_id", message.ToAgent),
				zap.Error(err),
			)
		}
	}
}

// processTask processes a task
func (o *Orchestrator) processTask(task *Task) {
	task.Status = TaskStatusRunning
	task.UpdatedAt = time.Now()

	// Get agent
	agent, err := o.agentRegistry.GetAgent(task.AgentID)
	if err != nil {
		task.Status = TaskStatusFailed
		task.Error = fmt.Sprintf("agent not found: %s", task.AgentID)
		return
	}

	// Execute task
	outputs, err := agent.ExecuteAction(context.Background(), task.Action, task.Inputs)
	if err != nil {
		task.Status = TaskStatusFailed
		task.Error = err.Error()
		o.logger.Error("Task execution failed",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
		return
	}

	task.Outputs = outputs
	task.Status = TaskStatusCompleted
	task.UpdatedAt = time.Now()
	now := time.Now()
	task.CompletedAt = &now

	o.logger.Info("Task completed",
		zap.String("task_id", task.ID),
		zap.String("agent_id", task.AgentID),
		zap.String("action", task.Action),
	)
}

// workflowMonitor monitors workflow execution
func (o *Orchestrator) workflowMonitor(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-o.stopChan:
			return
		case <-ticker.C:
			o.monitorWorkflows()
		}
	}
}

// healthMonitor monitors agent health
func (o *Orchestrator) healthMonitor(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-o.stopChan:
			return
		case <-ticker.C:
			o.checkAgentHealth()
		}
	}
}

// monitorWorkflows monitors running workflows
func (o *Orchestrator) monitorWorkflows() {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, workflow := range o.workflows {
		if workflow.Status == WorkflowStatusRunning {
			// Check for timeout or other issues
			if time.Since(workflow.UpdatedAt) > 30*time.Minute {
				o.logger.Warn("Workflow appears to be stuck",
					zap.String("workflow_id", workflow.ID),
					zap.String("name", workflow.Name),
				)
			}
		}
	}
}

// checkAgentHealth checks the health of all registered agents
func (o *Orchestrator) checkAgentHealth() {
	agents := o.agentRegistry.ListAgents()
	for _, agent := range agents {
		if !agent.IsHealthy() {
			o.logger.Warn("Agent is unhealthy",
				zap.String("agent_id", agent.GetID()),
				zap.String("agent_type", agent.GetType()),
			)
		}
	}
}

// initializePredefinedWorkflows initializes common workflows
func (o *Orchestrator) initializePredefinedWorkflows() error {
	// Coffee Order Processing Workflow
	orderWorkflow := &Workflow{
		ID:          "coffee-order-processing",
		Name:        "Coffee Order Processing",
		Description: "Automated coffee order processing with AI assistance",
		Steps: []*WorkflowStep{
			{
				ID:      "analyze-order",
				AgentID: "beverage-inventor",
				Action:  "analyze_order",
				Inputs:  map[string]interface{}{"order_type": "analysis"},
			},
			{
				ID:      "check-inventory",
				AgentID: "inventory-manager",
				Action:  "check_availability",
				Inputs:  map[string]interface{}{"check_type": "ingredients"},
			},
			{
				ID:      "schedule-preparation",
				AgentID: "scheduler",
				Action:  "schedule_task",
				Inputs:  map[string]interface{}{"task_type": "preparation"},
			},
			{
				ID:      "notify-customer",
				AgentID: "notifier",
				Action:  "send_notification",
				Inputs:  map[string]interface{}{"notification_type": "order_status"},
			},
		},
	}

	if err := o.CreateWorkflow(orderWorkflow); err != nil {
		return fmt.Errorf("failed to create order workflow: %w", err)
	}

	// Daily Operations Workflow
	dailyOpsWorkflow := &Workflow{
		ID:          "daily-operations",
		Name:        "Daily Operations Optimization",
		Description: "Daily operational tasks and optimizations",
		Steps: []*WorkflowStep{
			{
				ID:      "forecast-demand",
				AgentID: "inventory-manager",
				Action:  "forecast_demand",
				Inputs:  map[string]interface{}{"period": "daily"},
			},
			{
				ID:      "optimize-schedule",
				AgentID: "scheduler",
				Action:  "optimize_schedule",
				Inputs:  map[string]interface{}{"optimization_type": "daily"},
			},
			{
				ID:      "coordinate-locations",
				AgentID: "inter-location-coordinator",
				Action:  "coordinate_operations",
				Inputs:  map[string]interface{}{"scope": "all_locations"},
			},
			{
				ID:      "generate-content",
				AgentID: "social-media",
				Action:  "generate_daily_content",
				Inputs:  map[string]interface{}{"content_type": "daily_special"},
			},
		},
	}

	if err := o.CreateWorkflow(dailyOpsWorkflow); err != nil {
		return fmt.Errorf("failed to create daily ops workflow: %w", err)
	}

	o.logger.Info("Predefined workflows initialized successfully")
	return nil
}

// GetWorkflow returns a workflow by ID
func (o *Orchestrator) GetWorkflow(workflowID string) (*Workflow, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	workflow, exists := o.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	return workflow, nil
}

// ListWorkflows returns all workflows
func (o *Orchestrator) ListWorkflows() []*Workflow {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	workflows := make([]*Workflow, 0, len(o.workflows))
	for _, workflow := range o.workflows {
		workflows = append(workflows, workflow)
	}

	return workflows
}

// GetTask returns a task by ID
func (o *Orchestrator) GetTask(taskID string) (*Task, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	task, exists := o.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return task, nil
}

// ListTasks returns all tasks
func (o *Orchestrator) ListTasks() []*Task {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	tasks := make([]*Task, 0, len(o.tasks))
	for _, task := range o.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}
