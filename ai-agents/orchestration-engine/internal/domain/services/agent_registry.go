package services

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// DefaultAgentRegistry implements the AgentRegistry interface
type DefaultAgentRegistry struct {
	agents      map[string]Agent
	agentsMutex sync.RWMutex
	healthData  map[string]*AgentHealth
	healthMutex sync.RWMutex
	logger      Logger
}

// NewDefaultAgentRegistry creates a new agent registry
func NewDefaultAgentRegistry(logger Logger) *DefaultAgentRegistry {
	return &DefaultAgentRegistry{
		agents:     make(map[string]Agent),
		healthData: make(map[string]*AgentHealth),
		logger:     logger,
	}
}

// RegisterAgent registers an agent with the registry
func (ar *DefaultAgentRegistry) RegisterAgent(agentType string, agent Agent) error {
	ar.agentsMutex.Lock()
	defer ar.agentsMutex.Unlock()

	if _, exists := ar.agents[agentType]; exists {
		return fmt.Errorf("agent type %s already registered", agentType)
	}

	ar.agents[agentType] = agent
	
	// Initialize health data
	ar.healthMutex.Lock()
	ar.healthData[agentType] = &AgentHealth{
		Status:   AgentStatusOnline,
		LastSeen: time.Now(),
	}
	ar.healthMutex.Unlock()

	ar.logger.Info("Agent registered", "agent_type", agentType)
	return nil
}

// GetAgent retrieves an agent by type
func (ar *DefaultAgentRegistry) GetAgent(agentType string) (Agent, error) {
	ar.agentsMutex.RLock()
	defer ar.agentsMutex.RUnlock()

	agent, exists := ar.agents[agentType]
	if !exists {
		return nil, fmt.Errorf("agent type %s not found", agentType)
	}

	// Update last seen
	ar.healthMutex.Lock()
	if health, exists := ar.healthData[agentType]; exists {
		health.LastSeen = time.Now()
	}
	ar.healthMutex.Unlock()

	return agent, nil
}

// ListAgents returns all registered agents
func (ar *DefaultAgentRegistry) ListAgents() map[string]Agent {
	ar.agentsMutex.RLock()
	defer ar.agentsMutex.RUnlock()

	result := make(map[string]Agent)
	for agentType, agent := range ar.agents {
		result[agentType] = agent
	}

	return result
}

// IsAgentAvailable checks if an agent is available
func (ar *DefaultAgentRegistry) IsAgentAvailable(agentType string) bool {
	ar.agentsMutex.RLock()
	agent, exists := ar.agents[agentType]
	ar.agentsMutex.RUnlock()

	if !exists {
		return false
	}

	status := agent.GetStatus()
	return status == AgentStatusOnline
}

// GetAgentHealth retrieves health information for an agent
func (ar *DefaultAgentRegistry) GetAgentHealth(agentType string) (*AgentHealth, error) {
	ar.healthMutex.RLock()
	defer ar.healthMutex.RUnlock()

	health, exists := ar.healthData[agentType]
	if !exists {
		return nil, fmt.Errorf("health data not found for agent type %s", agentType)
	}

	// Create a copy to avoid race conditions
	healthCopy := *health
	return &healthCopy, nil
}

// UpdateAgentHealth updates health information for an agent
func (ar *DefaultAgentRegistry) UpdateAgentHealth(agentType string, health *AgentHealth) {
	ar.healthMutex.Lock()
	defer ar.healthMutex.Unlock()

	ar.healthData[agentType] = health
}

// MonitorAgentHealth starts monitoring agent health
func (ar *DefaultAgentRegistry) MonitorAgentHealth(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ar.checkAgentHealth(ctx)
		}
	}
}

// checkAgentHealth checks the health of all registered agents
func (ar *DefaultAgentRegistry) checkAgentHealth(ctx context.Context) {
	ar.agentsMutex.RLock()
	agents := make(map[string]Agent)
	for agentType, agent := range ar.agents {
		agents[agentType] = agent
	}
	ar.agentsMutex.RUnlock()

	for agentType, agent := range agents {
		go ar.checkSingleAgentHealth(ctx, agentType, agent)
	}
}

// checkSingleAgentHealth checks the health of a single agent
func (ar *DefaultAgentRegistry) checkSingleAgentHealth(ctx context.Context, agentType string, agent Agent) {
	startTime := time.Now()
	
	// Get agent status
	status := agent.GetStatus()
	responseTime := time.Since(startTime)

	// Get agent metrics
	metrics := agent.GetMetrics()
	
	// Calculate error rate
	errorRate := 0.0
	if metrics.TotalRequests > 0 {
		errorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests) * 100
	}

	// Update health data
	health := &AgentHealth{
		Status:       status,
		LastSeen:     time.Now(),
		ResponseTime: responseTime,
		ErrorRate:    errorRate,
		Load:         metrics.CurrentLoad,
	}

	ar.UpdateAgentHealth(agentType, health)

	// Log health issues
	if status != AgentStatusOnline {
		ar.logger.Warn("Agent health issue detected", 
			"agent_type", agentType, 
			"status", status, 
			"error_rate", errorRate)
	}
}

// Agent implementations for different agent types

// SocialMediaContentAgent implements the Agent interface for social media content operations
type SocialMediaContentAgent struct {
	baseURL string
	client  HTTPClient
	metrics *AgentMetrics
	mutex   sync.RWMutex
	logger  Logger
}

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	Post(ctx context.Context, url string, data interface{}) (map[string]interface{}, error)
	Get(ctx context.Context, url string) (map[string]interface{}, error)
	Put(ctx context.Context, url string, data interface{}) (map[string]interface{}, error)
	Delete(ctx context.Context, url string) error
}

// NewSocialMediaContentAgent creates a new social media content agent
func NewSocialMediaContentAgent(baseURL string, client HTTPClient, logger Logger) *SocialMediaContentAgent {
	return &SocialMediaContentAgent{
		baseURL: baseURL,
		client:  client,
		metrics: &AgentMetrics{
			LastUpdated: time.Now(),
		},
		logger: logger,
	}
}

// Execute executes an action on the social media content agent
func (smca *SocialMediaContentAgent) Execute(ctx context.Context, action string, input map[string]interface{}) (map[string]interface{}, error) {
	smca.updateMetrics(true, time.Now())
	defer func(start time.Time) {
		smca.updateResponseTime(time.Since(start))
	}(time.Now())

	switch action {
	case "create_content":
		return smca.createContent(ctx, input)
	case "schedule_content":
		return smca.scheduleContent(ctx, input)
	case "publish_content":
		return smca.publishContent(ctx, input)
	case "analyze_content":
		return smca.analyzeContent(ctx, input)
	case "generate_variations":
		return smca.generateVariations(ctx, input)
	default:
		smca.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

// GetCapabilities returns the capabilities of the agent
func (smca *SocialMediaContentAgent) GetCapabilities() []string {
	return []string{
		"create_content",
		"schedule_content", 
		"publish_content",
		"analyze_content",
		"generate_variations",
	}
}

// GetStatus returns the current status of the agent
func (smca *SocialMediaContentAgent) GetStatus() AgentStatus {
	// Simple health check - in production this would be more sophisticated
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := smca.client.Get(ctx, smca.baseURL+"/health")
	if err != nil {
		return AgentStatusOffline
	}

	return AgentStatusOnline
}

// Validate validates the input for an action
func (smca *SocialMediaContentAgent) Validate(action string, input map[string]interface{}) error {
	switch action {
	case "create_content":
		return smca.validateCreateContent(input)
	case "schedule_content":
		return smca.validateScheduleContent(input)
	case "publish_content":
		return smca.validatePublishContent(input)
	case "analyze_content":
		return smca.validateAnalyzeContent(input)
	case "generate_variations":
		return smca.validateGenerateVariations(input)
	default:
		return fmt.Errorf("unsupported action: %s", action)
	}
}

// GetMetrics returns the current metrics for the agent
func (smca *SocialMediaContentAgent) GetMetrics() *AgentMetrics {
	smca.mutex.RLock()
	defer smca.mutex.RUnlock()

	// Return a copy to avoid race conditions
	metricsCopy := *smca.metrics
	return &metricsCopy
}

// Action implementations

func (smca *SocialMediaContentAgent) createContent(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	smca.logger.Info("Creating content", "input", input)
	
	result, err := smca.client.Post(ctx, smca.baseURL+"/api/v1/content", input)
	if err != nil {
		smca.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to create content: %w", err)
	}

	smca.updateMetrics(true, time.Now())
	return result, nil
}

func (smca *SocialMediaContentAgent) scheduleContent(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	smca.logger.Info("Scheduling content", "input", input)

	contentID, exists := input["content_id"]
	if !exists {
		return nil, fmt.Errorf("content_id is required")
	}

	url := fmt.Sprintf("%s/api/v1/content/%v/schedule", smca.baseURL, contentID)
	result, err := smca.client.Post(ctx, url, input)
	if err != nil {
		smca.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to schedule content: %w", err)
	}

	smca.updateMetrics(true, time.Now())
	return result, nil
}

func (smca *SocialMediaContentAgent) publishContent(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	smca.logger.Info("Publishing content", "input", input)

	contentID, exists := input["content_id"]
	if !exists {
		return nil, fmt.Errorf("content_id is required")
	}

	url := fmt.Sprintf("%s/api/v1/content/%v/publish", smca.baseURL, contentID)
	result, err := smca.client.Post(ctx, url, input)
	if err != nil {
		smca.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to publish content: %w", err)
	}

	smca.updateMetrics(true, time.Now())
	return result, nil
}

func (smca *SocialMediaContentAgent) analyzeContent(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	smca.logger.Info("Analyzing content", "input", input)

	result, err := smca.client.Post(ctx, smca.baseURL+"/api/v1/ai/analyze", input)
	if err != nil {
		smca.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to analyze content: %w", err)
	}

	smca.updateMetrics(true, time.Now())
	return result, nil
}

func (smca *SocialMediaContentAgent) generateVariations(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	smca.logger.Info("Generating content variations", "input", input)

	contentID, exists := input["content_id"]
	if !exists {
		return nil, fmt.Errorf("content_id is required")
	}

	url := fmt.Sprintf("%s/api/v1/content/%v/variations", smca.baseURL, contentID)
	result, err := smca.client.Post(ctx, url, input)
	if err != nil {
		smca.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to generate variations: %w", err)
	}

	smca.updateMetrics(true, time.Now())
	return result, nil
}

// Validation methods

func (smca *SocialMediaContentAgent) validateCreateContent(input map[string]interface{}) error {
	required := []string{"title", "type", "brand_id", "created_by"}
	for _, field := range required {
		if _, exists := input[field]; !exists {
			return fmt.Errorf("required field missing: %s", field)
		}
	}
	return nil
}

func (smca *SocialMediaContentAgent) validateScheduleContent(input map[string]interface{}) error {
	required := []string{"content_id", "platforms", "scheduled_at"}
	for _, field := range required {
		if _, exists := input[field]; !exists {
			return fmt.Errorf("required field missing: %s", field)
		}
	}
	return nil
}

func (smca *SocialMediaContentAgent) validatePublishContent(input map[string]interface{}) error {
	required := []string{"content_id", "platforms"}
	for _, field := range required {
		if _, exists := input[field]; !exists {
			return fmt.Errorf("required field missing: %s", field)
		}
	}
	return nil
}

func (smca *SocialMediaContentAgent) validateAnalyzeContent(input map[string]interface{}) error {
	if _, exists := input["content"]; !exists {
		if _, exists := input["content_id"]; !exists {
			return fmt.Errorf("either content or content_id is required")
		}
	}
	return nil
}

func (smca *SocialMediaContentAgent) validateGenerateVariations(input map[string]interface{}) error {
	required := []string{"content_id"}
	for _, field := range required {
		if _, exists := input[field]; !exists {
			return fmt.Errorf("required field missing: %s", field)
		}
	}
	return nil
}

// Helper methods

func (smca *SocialMediaContentAgent) updateMetrics(success bool, timestamp time.Time) {
	smca.mutex.Lock()
	defer smca.mutex.Unlock()

	smca.metrics.TotalRequests++
	if success {
		smca.metrics.SuccessfulRequests++
	} else {
		smca.metrics.FailedRequests++
	}
	smca.metrics.LastUpdated = timestamp
}

func (smca *SocialMediaContentAgent) updateResponseTime(duration time.Duration) {
	smca.mutex.Lock()
	defer smca.mutex.Unlock()

	// Simple moving average
	if smca.metrics.TotalRequests == 1 {
		smca.metrics.AverageResponseTime = duration
	} else {
		smca.metrics.AverageResponseTime = (smca.metrics.AverageResponseTime + duration) / 2
	}
}

// FeedbackAnalystAgent implements the Agent interface for feedback analysis operations
type FeedbackAnalystAgent struct {
	baseURL string
	client  HTTPClient
	metrics *AgentMetrics
	mutex   sync.RWMutex
	logger  Logger
}

// NewFeedbackAnalystAgent creates a new feedback analyst agent
func NewFeedbackAnalystAgent(baseURL string, client HTTPClient, logger Logger) *FeedbackAnalystAgent {
	return &FeedbackAnalystAgent{
		baseURL: baseURL,
		client:  client,
		metrics: &AgentMetrics{
			LastUpdated: time.Now(),
		},
		logger: logger,
	}
}

// Execute executes an action on the feedback analyst agent
func (faa *FeedbackAnalystAgent) Execute(ctx context.Context, action string, input map[string]interface{}) (map[string]interface{}, error) {
	faa.updateMetrics(true, time.Now())
	defer func(start time.Time) {
		faa.updateResponseTime(time.Since(start))
	}(time.Now())

	switch action {
	case "analyze_feedback":
		return faa.analyzeFeedback(ctx, input)
	case "generate_response":
		return faa.generateResponse(ctx, input)
	case "categorize_feedback":
		return faa.categorizeFeedback(ctx, input)
	case "extract_insights":
		return faa.extractInsights(ctx, input)
	default:
		faa.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

// GetCapabilities returns the capabilities of the agent
func (faa *FeedbackAnalystAgent) GetCapabilities() []string {
	return []string{
		"analyze_feedback",
		"generate_response",
		"categorize_feedback",
		"extract_insights",
	}
}

// GetStatus returns the current status of the agent
func (faa *FeedbackAnalystAgent) GetStatus() AgentStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := faa.client.Get(ctx, faa.baseURL+"/health")
	if err != nil {
		return AgentStatusOffline
	}

	return AgentStatusOnline
}

// Validate validates the input for an action
func (faa *FeedbackAnalystAgent) Validate(action string, input map[string]interface{}) error {
	switch action {
	case "analyze_feedback":
		return faa.validateAnalyzeFeedback(input)
	case "generate_response":
		return faa.validateGenerateResponse(input)
	case "categorize_feedback":
		return faa.validateCategorizeFeedback(input)
	case "extract_insights":
		return faa.validateExtractInsights(input)
	default:
		return fmt.Errorf("unsupported action: %s", action)
	}
}

// GetMetrics returns the current metrics for the agent
func (faa *FeedbackAnalystAgent) GetMetrics() *AgentMetrics {
	faa.mutex.RLock()
	defer faa.mutex.RUnlock()

	metricsCopy := *faa.metrics
	return &metricsCopy
}

// Action implementations for FeedbackAnalystAgent

func (faa *FeedbackAnalystAgent) analyzeFeedback(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	faa.logger.Info("Analyzing feedback", "input", input)
	
	result, err := faa.client.Post(ctx, faa.baseURL+"/api/v1/feedback/analyze", input)
	if err != nil {
		faa.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to analyze feedback: %w", err)
	}

	faa.updateMetrics(true, time.Now())
	return result, nil
}

func (faa *FeedbackAnalystAgent) generateResponse(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	faa.logger.Info("Generating response", "input", input)

	result, err := faa.client.Post(ctx, faa.baseURL+"/api/v1/feedback/response", input)
	if err != nil {
		faa.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	faa.updateMetrics(true, time.Now())
	return result, nil
}

func (faa *FeedbackAnalystAgent) categorizeFeedback(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	faa.logger.Info("Categorizing feedback", "input", input)

	result, err := faa.client.Post(ctx, faa.baseURL+"/api/v1/feedback/categorize", input)
	if err != nil {
		faa.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to categorize feedback: %w", err)
	}

	faa.updateMetrics(true, time.Now())
	return result, nil
}

func (faa *FeedbackAnalystAgent) extractInsights(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	faa.logger.Info("Extracting insights", "input", input)

	result, err := faa.client.Post(ctx, faa.baseURL+"/api/v1/feedback/insights", input)
	if err != nil {
		faa.updateMetrics(false, time.Now())
		return nil, fmt.Errorf("failed to extract insights: %w", err)
	}

	faa.updateMetrics(true, time.Now())
	return result, nil
}

// Validation methods for FeedbackAnalystAgent

func (faa *FeedbackAnalystAgent) validateAnalyzeFeedback(input map[string]interface{}) error {
	if _, exists := input["feedback"]; !exists {
		if _, exists := input["feedback_id"]; !exists {
			return fmt.Errorf("either feedback or feedback_id is required")
		}
	}
	return nil
}

func (faa *FeedbackAnalystAgent) validateGenerateResponse(input map[string]interface{}) error {
	required := []string{"feedback_id"}
	for _, field := range required {
		if _, exists := input[field]; !exists {
			return fmt.Errorf("required field missing: %s", field)
		}
	}
	return nil
}

func (faa *FeedbackAnalystAgent) validateCategorizeFeedback(input map[string]interface{}) error {
	if _, exists := input["feedback"]; !exists {
		if _, exists := input["feedback_id"]; !exists {
			return fmt.Errorf("either feedback or feedback_id is required")
		}
	}
	return nil
}

func (faa *FeedbackAnalystAgent) validateExtractInsights(input map[string]interface{}) error {
	// Insights can be extracted from multiple feedback items or a time period
	return nil
}

// Helper methods for FeedbackAnalystAgent

func (faa *FeedbackAnalystAgent) updateMetrics(success bool, timestamp time.Time) {
	faa.mutex.Lock()
	defer faa.mutex.Unlock()

	faa.metrics.TotalRequests++
	if success {
		faa.metrics.SuccessfulRequests++
	} else {
		faa.metrics.FailedRequests++
	}
	faa.metrics.LastUpdated = timestamp
}

func (faa *FeedbackAnalystAgent) updateResponseTime(duration time.Duration) {
	faa.mutex.Lock()
	defer faa.mutex.Unlock()

	if faa.metrics.TotalRequests == 1 {
		faa.metrics.AverageResponseTime = duration
	} else {
		faa.metrics.AverageResponseTime = (faa.metrics.AverageResponseTime + duration) / 2
	}
}
