// Package graph provides LangGraph-inspired graph execution for Go Coffee AI agents
package graph

import (
	"context"
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

// NodeType represents different types of nodes in the graph
type NodeType string

const (
	NodeTypeAgent     NodeType = "agent"
	NodeTypeCondition NodeType = "condition"
	NodeTypeControl   NodeType = "control"
	NodeTypeEnd       NodeType = "end"
)

// NodeFunc represents a function that processes a node
type NodeFunc func(ctx context.Context, state *AgentState) (*AgentState, error)

// ConditionFunc represents a function that evaluates a condition
type ConditionFunc func(ctx context.Context, state *AgentState) (string, error)

// Node represents a node in the execution graph
type Node struct {
	ID          string         `json:"id"`
	Type        NodeType       `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Function    NodeFunc       `json:"-"`
	Condition   ConditionFunc  `json:"-"`
	Timeout     time.Duration  `json:"timeout"`
	Retries     int            `json:"retries"`
	Metadata    map[string]any `json:"metadata"`
}

// Edge represents a connection between nodes
type Edge struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Condition string `json:"condition,omitempty"`
	Weight    int    `json:"weight"`
}

// ExecutionResult represents the result of a node execution
type ExecutionResult struct {
	NodeID        string         `json:"node_id"`
	Success       bool           `json:"success"`
	Output        map[string]any `json:"output"`
	Error         string         `json:"error,omitempty"`
	ExecutionTime time.Duration  `json:"execution_time"`
	NextNode      string         `json:"next_node,omitempty"`
}

// Graph represents the execution graph for AI agents
type Graph struct {
	ID          uuid.UUID        `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Version     string           `json:"version"`
	Nodes       map[string]*Node `json:"nodes"`
	Edges       []*Edge          `json:"edges"`
	StartNode   string           `json:"start_node"`
	EndNodes    []string         `json:"end_nodes"`
	Config      map[string]any   `json:"config"`

	// Internal state
	mu               sync.RWMutex
	executionCount   int
	lastExecution    *time.Time
	isCompiled       bool
	adjacencyList    map[string][]string
	conditionalEdges map[string]map[string]string
}

// NewGraph creates a new execution graph
func NewGraph(name, description string) *Graph {
	return &Graph{
		ID:               uuid.New(),
		Name:             name,
		Description:      description,
		Version:          "1.0.0",
		Nodes:            make(map[string]*Node),
		Edges:            make([]*Edge, 0),
		EndNodes:         make([]string, 0),
		Config:           make(map[string]any),
		adjacencyList:    make(map[string][]string),
		conditionalEdges: make(map[string]map[string]string),
	}
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(node *Node) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.Nodes[node.ID]; exists {
		return fmt.Errorf("node with ID %s already exists", node.ID)
	}

	g.Nodes[node.ID] = node
	g.isCompiled = false
	return nil
}

// AddEdge adds an edge to the graph
func (g *Graph) AddEdge(edge *Edge) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Validate that nodes exist
	if _, exists := g.Nodes[edge.From]; !exists {
		return fmt.Errorf("source node %s does not exist", edge.From)
	}
	if _, exists := g.Nodes[edge.To]; !exists {
		return fmt.Errorf("target node %s does not exist", edge.To)
	}

	g.Edges = append(g.Edges, edge)
	g.isCompiled = false
	return nil
}

// AddConditionalEdge adds a conditional edge to the graph
func (g *Graph) AddConditionalEdge(from string, conditions map[string]string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.Nodes[from]; !exists {
		return fmt.Errorf("source node %s does not exist", from)
	}

	// Validate that all target nodes exist
	for _, to := range conditions {
		if to != "END" {
			if _, exists := g.Nodes[to]; !exists {
				return fmt.Errorf("target node %s does not exist", to)
			}
		}
	}

	g.conditionalEdges[from] = conditions
	g.isCompiled = false
	return nil
}

// SetStartNode sets the starting node for the graph
func (g *Graph) SetStartNode(nodeID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.Nodes[nodeID]; !exists {
		return fmt.Errorf("start node %s does not exist", nodeID)
	}

	g.StartNode = nodeID
	g.isCompiled = false
	return nil
}

// AddEndNode adds an end node to the graph
func (g *Graph) AddEndNode(nodeID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.Nodes[nodeID]; !exists {
		return fmt.Errorf("end node %s does not exist", nodeID)
	}

	g.EndNodes = append(g.EndNodes, nodeID)
	g.isCompiled = false
	return nil
}

// Compile compiles the graph for execution
func (g *Graph) Compile() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.StartNode == "" {
		return fmt.Errorf("start node not set")
	}

	// Build adjacency list
	g.adjacencyList = make(map[string][]string)
	for _, edge := range g.Edges {
		g.adjacencyList[edge.From] = append(g.adjacencyList[edge.From], edge.To)
	}

	// Validate graph connectivity
	if err := g.validateGraph(); err != nil {
		return fmt.Errorf("graph validation failed: %w", err)
	}

	g.isCompiled = true
	log.Printf("Graph %s compiled successfully with %d nodes and %d edges",
		g.Name, len(g.Nodes), len(g.Edges))

	return nil
}

// validateGraph validates the graph structure
func (g *Graph) validateGraph() error {
	// Check if start node exists
	if _, exists := g.Nodes[g.StartNode]; !exists {
		return fmt.Errorf("start node %s does not exist", g.StartNode)
	}

	// Check if all end nodes exist
	for _, endNode := range g.EndNodes {
		if _, exists := g.Nodes[endNode]; !exists {
			return fmt.Errorf("end node %s does not exist", endNode)
		}
	}

	// Check for cycles (simplified check)
	visited := make(map[string]bool)
	if g.hasCycle(g.StartNode, visited, make(map[string]bool)) {
		return fmt.Errorf("graph contains cycles")
	}

	return nil
}

// hasCycle checks for cycles in the graph using DFS
func (g *Graph) hasCycle(nodeID string, visited, recStack map[string]bool) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	// Check regular edges
	for _, neighbor := range g.adjacencyList[nodeID] {
		if !visited[neighbor] {
			if g.hasCycle(neighbor, visited, recStack) {
				return true
			}
		} else if recStack[neighbor] {
			return true
		}
	}

	// Check conditional edges
	if conditions, exists := g.conditionalEdges[nodeID]; exists {
		for _, neighbor := range conditions {
			if neighbor != "END" {
				if !visited[neighbor] {
					if g.hasCycle(neighbor, visited, recStack) {
						return true
					}
				} else if recStack[neighbor] {
					return true
				}
			}
		}
	}

	recStack[nodeID] = false
	return false
}

// Execute executes the graph with the given initial state
func (g *Graph) Execute(ctx context.Context, initialState *AgentState) (*AgentState, error) {
	if !g.isCompiled {
		if err := g.Compile(); err != nil {
			return nil, fmt.Errorf("failed to compile graph: %w", err)
		}
	}

	g.mu.Lock()
	g.executionCount++
	now := time.Now()
	g.lastExecution = &now
	g.mu.Unlock()

	log.Printf("Starting graph execution %s for workflow %s",
		g.ID, initialState.WorkflowID)

	// Start execution
	initialState.StartExecution()
	currentState := initialState
	currentNodeID := g.StartNode

	// Execution loop
	for currentNodeID != "" && currentNodeID != "END" {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return currentState, ctx.Err()
		default:
		}

		// Check if we've reached an end node
		if g.isEndNode(currentNodeID) {
			break
		}

		// Execute current node
		result, err := g.executeNode(ctx, currentNodeID, currentState)
		if err != nil {
			currentState.FailExecution(err.Error())
			return currentState, fmt.Errorf("node execution failed at %s: %w", currentNodeID, err)
		}

		// Update state with node result
		currentState.SetSharedData(fmt.Sprintf("node_%s_result", currentNodeID), result)

		// Determine next node
		nextNodeID, err := g.determineNextNode(ctx, currentNodeID, currentState)
		if err != nil {
			currentState.FailExecution(err.Error())
			return currentState, fmt.Errorf("failed to determine next node from %s: %w", currentNodeID, err)
		}

		currentNodeID = nextNodeID
	}

	// Complete execution
	currentState.CompleteExecution()
	log.Printf("Graph execution completed for workflow %s in %.2fs",
		currentState.WorkflowID, *currentState.ExecutionTime)

	return currentState, nil
}

// executeNode executes a specific node
func (g *Graph) executeNode(ctx context.Context, nodeID string, state *AgentState) (*ExecutionResult, error) {
	node, exists := g.Nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}

	log.Printf("Executing node %s (%s) for workflow %s",
		nodeID, node.Name, state.WorkflowID)

	start := time.Now()

	// Create node context with timeout
	nodeCtx := ctx
	if node.Timeout > 0 {
		var cancel context.CancelFunc
		nodeCtx, cancel = context.WithTimeout(ctx, node.Timeout)
		defer cancel()
	}

	// Execute node function
	var resultState *AgentState
	var err error

	if node.Function != nil {
		resultState, err = node.Function(nodeCtx, state)
	} else {
		// Default behavior for nodes without functions
		resultState = state
	}

	executionTime := time.Since(start)

	// Create execution result
	result := &ExecutionResult{
		NodeID:        nodeID,
		Success:       err == nil,
		Output:        make(map[string]any),
		ExecutionTime: executionTime,
	}

	if err != nil {
		result.Error = err.Error()
		log.Printf("Node %s failed after %v: %s", nodeID, executionTime, err.Error())
	} else {
		log.Printf("Node %s completed successfully in %v", nodeID, executionTime)
		// Copy result state back to original state
		*state = *resultState
	}

	return result, err
}

// determineNextNode determines the next node to execute
func (g *Graph) determineNextNode(ctx context.Context, currentNodeID string, state *AgentState) (string, error) {
	// Check for conditional edges first
	if conditions, exists := g.conditionalEdges[currentNodeID]; exists {
		node := g.Nodes[currentNodeID]
		if node.Condition != nil {
			condition, err := node.Condition(ctx, state)
			if err != nil {
				return "", fmt.Errorf("condition evaluation failed: %w", err)
			}

			if nextNode, exists := conditions[condition]; exists {
				return nextNode, nil
			}

			// Default condition
			if defaultNode, exists := conditions["default"]; exists {
				return defaultNode, nil
			}
		}
	}

	// Check regular edges
	if neighbors, exists := g.adjacencyList[currentNodeID]; exists && len(neighbors) > 0 {
		// For simplicity, take the first neighbor
		// In a more complex implementation, you might have routing logic here
		return neighbors[0], nil
	}

	// No next node found
	return "END", nil
}

// isEndNode checks if a node is an end node
func (g *Graph) isEndNode(nodeID string) bool {
	return slices.Contains(g.EndNodes, nodeID)
}

// GetExecutionStats returns execution statistics for the graph
func (g *Graph) GetExecutionStats() map[string]any {
	g.mu.RLock()
	defer g.mu.RUnlock()

	stats := map[string]any{
		"graph_id":        g.ID,
		"name":            g.Name,
		"version":         g.Version,
		"execution_count": g.executionCount,
		"node_count":      len(g.Nodes),
		"edge_count":      len(g.Edges),
		"is_compiled":     g.isCompiled,
	}

	if g.lastExecution != nil {
		stats["last_execution"] = g.lastExecution.Format(time.RFC3339)
	}

	return stats
}

// GetNodeInfo returns information about a specific node
func (g *Graph) GetNodeInfo(nodeID string) (map[string]any, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	node, exists := g.Nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}

	return map[string]any{
		"id":          node.ID,
		"type":        node.Type,
		"name":        node.Name,
		"description": node.Description,
		"timeout":     node.Timeout,
		"retries":     node.Retries,
		"metadata":    node.Metadata,
	}, nil
}
