package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/orchestration-engine/internal/domain/entities"
	"go-coffee-ai-agents/orchestration-engine/internal/domain/services"
)

// MockWorkflowRepository implements WorkflowRepository for testing/demo purposes
type MockWorkflowRepository struct {
	workflows map[string]*entities.Workflow
	mutex     sync.RWMutex
	logger    Logger
}

// NewMockWorkflowRepository creates a new mock workflow repository
func NewMockWorkflowRepository(logger Logger) *MockWorkflowRepository {
	return &MockWorkflowRepository{
		workflows: make(map[string]*entities.Workflow),
		logger:    logger,
	}
}

func (r *MockWorkflowRepository) Create(ctx context.Context, workflow *entities.Workflow) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	workflowID := workflow.ID.String()
	r.workflows[workflowID] = workflow
	r.logger.Info("Workflow created", "workflow_id", workflowID, "name", workflow.Name)
	return nil
}

func (r *MockWorkflowRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Workflow, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	workflow, exists := r.workflows[id.String()]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", id.String())
	}
	return workflow, nil
}

func (r *MockWorkflowRepository) Update(ctx context.Context, workflow *entities.Workflow) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	workflowID := workflow.ID.String()
	r.workflows[workflowID] = workflow
	r.logger.Info("Workflow updated", "workflow_id", workflowID)
	return nil
}

func (r *MockWorkflowRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.workflows, id.String())
	r.logger.Info("Workflow deleted", "workflow_id", id.String())
	return nil
}

func (r *MockWorkflowRepository) List(ctx context.Context, filter *services.WorkflowFilter) ([]*entities.Workflow, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*entities.Workflow
	count := 0
	
	for _, workflow := range r.workflows {
		// Apply filters
		if filter != nil {
			// Status filter
			if len(filter.Status) > 0 {
				statusMatch := false
				for _, status := range filter.Status {
					if workflow.Status == status {
						statusMatch = true
						break
					}
				}
				if !statusMatch {
					continue
				}
			}

			// Type filter
			if len(filter.Type) > 0 {
				typeMatch := false
				for _, workflowType := range filter.Type {
					if workflow.Type == workflowType {
						typeMatch = true
						break
					}
				}
				if !typeMatch {
					continue
				}
			}

			// Active filter
			if filter.IsActive != nil && workflow.IsActive != *filter.IsActive {
				continue
			}

			// Template filter
			if filter.IsTemplate != nil && workflow.IsTemplate != *filter.IsTemplate {
				continue
			}

			// Skip if offset not reached
			if filter.Offset > 0 && count < filter.Offset {
				count++
				continue
			}

			// Stop if limit reached
			if filter.Limit > 0 && len(result) >= filter.Limit {
				break
			}
		}

		result = append(result, workflow)
		count++
	}

	return result, nil
}

func (r *MockWorkflowRepository) GetActiveWorkflows(ctx context.Context) ([]*entities.Workflow, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*entities.Workflow
	for _, workflow := range r.workflows {
		if workflow.IsActive && workflow.Status == entities.WorkflowStatusActive {
			result = append(result, workflow)
		}
	}
	return result, nil
}

func (r *MockWorkflowRepository) GetWorkflowsByTrigger(ctx context.Context, triggerType entities.TriggerType) ([]*entities.Workflow, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*entities.Workflow
	for _, workflow := range r.workflows {
		for _, trigger := range workflow.Triggers {
			if trigger.Type == triggerType && trigger.IsActive {
				result = append(result, workflow)
				break
			}
		}
	}
	return result, nil
}

// MockExecutionRepository implements ExecutionRepository for testing/demo purposes
type MockExecutionRepository struct {
	executions map[string]*entities.WorkflowExecution
	mutex      sync.RWMutex
	logger     Logger
}

// NewMockExecutionRepository creates a new mock execution repository
func NewMockExecutionRepository(logger Logger) *MockExecutionRepository {
	return &MockExecutionRepository{
		executions: make(map[string]*entities.WorkflowExecution),
		logger:     logger,
	}
}

func (r *MockExecutionRepository) Create(ctx context.Context, execution *entities.WorkflowExecution) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	executionID := execution.ID.String()
	r.executions[executionID] = execution
	r.logger.Info("Execution created", "execution_id", executionID, "workflow_id", execution.WorkflowID)
	return nil
}

func (r *MockExecutionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.WorkflowExecution, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	execution, exists := r.executions[id.String()]
	if !exists {
		return nil, fmt.Errorf("execution not found: %s", id.String())
	}
	return execution, nil
}

func (r *MockExecutionRepository) Update(ctx context.Context, execution *entities.WorkflowExecution) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	executionID := execution.ID.String()
	r.executions[executionID] = execution
	r.logger.Info("Execution updated", "execution_id", executionID, "status", execution.Status)
	return nil
}

func (r *MockExecutionRepository) List(ctx context.Context, filter *services.ExecutionFilter) ([]*entities.WorkflowExecution, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*entities.WorkflowExecution
	count := 0

	for _, execution := range r.executions {
		// Apply filters
		if filter != nil {
			// Workflow ID filter
			if len(filter.WorkflowID) > 0 {
				workflowMatch := false
				for _, workflowID := range filter.WorkflowID {
					if execution.WorkflowID == workflowID {
						workflowMatch = true
						break
					}
				}
				if !workflowMatch {
					continue
				}
			}

			// Status filter
			if len(filter.Status) > 0 {
				statusMatch := false
				for _, status := range filter.Status {
					if execution.Status == status {
						statusMatch = true
						break
					}
				}
				if !statusMatch {
					continue
				}
			}

			// Date filters
			if filter.StartedAfter != nil && execution.StartedAt.Before(*filter.StartedAfter) {
				continue
			}
			if filter.StartedBefore != nil && execution.StartedAt.After(*filter.StartedBefore) {
				continue
			}

			// Skip if offset not reached
			if filter.Offset > 0 && count < filter.Offset {
				count++
				continue
			}

			// Stop if limit reached
			if filter.Limit > 0 && len(result) >= filter.Limit {
				break
			}
		}

		result = append(result, execution)
		count++
	}

	return result, nil
}

func (r *MockExecutionRepository) GetActiveExecutions(ctx context.Context, workflowID uuid.UUID) ([]*entities.WorkflowExecution, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*entities.WorkflowExecution
	for _, execution := range r.executions {
		if execution.WorkflowID == workflowID && execution.Status == entities.WorkflowStatusRunning {
			result = append(result, execution)
		}
	}
	return result, nil
}

func (r *MockExecutionRepository) GetExecutionHistory(ctx context.Context, workflowID uuid.UUID, limit int) ([]*entities.WorkflowExecution, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*entities.WorkflowExecution
	count := 0

	// Sort by creation time (newest first) and limit results
	for _, execution := range r.executions {
		if execution.WorkflowID == workflowID && count < limit {
			result = append(result, execution)
			count++
		}
	}

	return result, nil
}

// MockEventPublisher implements EventPublisher for testing/demo purposes
type MockEventPublisher struct {
	logger Logger
	events []interface{}
	mutex  sync.RWMutex
}

// NewMockEventPublisher creates a new mock event publisher
func NewMockEventPublisher(logger Logger) *MockEventPublisher {
	return &MockEventPublisher{
		logger: logger,
		events: make([]interface{}, 0),
	}
}

func (p *MockEventPublisher) PublishWorkflowEvent(ctx context.Context, event *services.WorkflowEvent) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.events = append(p.events, event)
	p.logger.Info("Workflow event published", 
		"type", event.Type, 
		"workflow_id", event.WorkflowID,
		"timestamp", event.Timestamp)
	return nil
}

func (p *MockEventPublisher) PublishExecutionEvent(ctx context.Context, event *services.ExecutionEvent) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.events = append(p.events, event)
	p.logger.Info("Execution event published", 
		"type", event.Type, 
		"execution_id", event.ExecutionID,
		"timestamp", event.Timestamp)
	return nil
}

func (p *MockEventPublisher) PublishStepEvent(ctx context.Context, event *services.StepEvent) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.events = append(p.events, event)
	p.logger.Info("Step event published", 
		"type", event.Type, 
		"step_id", event.StepID,
		"timestamp", event.Timestamp)
	return nil
}

// GetEvents returns all published events (for testing)
func (p *MockEventPublisher) GetEvents() []interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	result := make([]interface{}, len(p.events))
	copy(result, p.events)
	return result
}

// GetEventCount returns the number of published events
func (p *MockEventPublisher) GetEventCount() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return len(p.events)
}

// ClearEvents clears all events (for testing)
func (p *MockEventPublisher) ClearEvents() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.events = p.events[:0]
}

// GetEventsByType returns events of a specific type
func (p *MockEventPublisher) GetEventsByType(eventType string) []interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var result []interface{}
	for _, event := range p.events {
		switch e := event.(type) {
		case *services.WorkflowEvent:
			if e.Type == eventType {
				result = append(result, e)
			}
		case *services.ExecutionEvent:
			if e.Type == eventType {
				result = append(result, e)
			}
		case *services.StepEvent:
			if e.Type == eventType {
				result = append(result, e)
			}
		}
	}
	return result
}

// GetLatestEvent returns the most recently published event
func (p *MockEventPublisher) GetLatestEvent() interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if len(p.events) == 0 {
		return nil
	}
	return p.events[len(p.events)-1]
}

// SimulateEventProcessing simulates processing events with delays
func (p *MockEventPublisher) SimulateEventProcessing(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.mutex.RLock()
			eventCount := len(p.events)
			p.mutex.RUnlock()

			if eventCount > 0 {
				p.logger.Info("Processing events", "event_count", eventCount)
				
				// Simulate some event processing
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}
