package services

import (
	"context"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/task-manager-agent/internal/domain/entities"
	"go-coffee-ai-agents/task-manager-agent/internal/domain/repositories"
)

// AITaskAutomationService provides AI-powered task automation and intelligence
type AITaskAutomationService struct {
	taskRepo         repositories.TaskRepository
	userRepo         repositories.UserRepository
	projectRepo      repositories.ProjectRepository
	workflowRepo     repositories.WorkflowRepository
	mlService        MachineLearningService
	nlpService       NaturalLanguageService
	eventPublisher   EventPublisher
	logger           Logger
}

// MachineLearningService defines the interface for ML operations
type MachineLearningService interface {
	PredictTaskDuration(ctx context.Context, task *entities.Task) (float64, error)
	PredictTaskComplexity(ctx context.Context, task *entities.Task) (entities.TaskComplexity, error)
	RecommendAssignees(ctx context.Context, task *entities.Task, availableUsers []*entities.User) ([]*UserRecommendation, error)
	OptimizeSchedule(ctx context.Context, tasks []*entities.Task, users []*entities.User) (*ScheduleOptimization, error)
	DetectRisks(ctx context.Context, task *entities.Task) ([]*TaskRisk, error)
	AnalyzeProductivity(ctx context.Context, userID uuid.UUID, period time.Duration) (*ProductivityAnalysis, error)
	PredictProjectCompletion(ctx context.Context, projectID uuid.UUID) (*ProjectCompletionPrediction, error)
}

// NaturalLanguageService defines the interface for NLP operations
type NaturalLanguageService interface {
	ExtractTaskDetails(ctx context.Context, description string) (*TaskDetails, error)
	GenerateTaskSummary(ctx context.Context, task *entities.Task) (string, error)
	AnalyzeSentiment(ctx context.Context, text string) (*SentimentAnalysis, error)
	ExtractKeywords(ctx context.Context, text string) ([]string, error)
	SuggestTags(ctx context.Context, task *entities.Task) ([]string, error)
	TranslateText(ctx context.Context, text, targetLanguage string) (string, error)
}

// Supporting types for AI services

// UserRecommendation represents a user recommendation for task assignment
type UserRecommendation struct {
	UserID           uuid.UUID `json:"user_id"`
	User             *entities.User `json:"user,omitempty"`
	Score            float64   `json:"score"`
	Confidence       float64   `json:"confidence"`
	Reasoning        string    `json:"reasoning"`
	SkillMatch       float64   `json:"skill_match"`
	AvailabilityScore float64  `json:"availability_score"`
	WorkloadScore    float64   `json:"workload_score"`
	PerformanceScore float64   `json:"performance_score"`
	EstimatedHours   float64   `json:"estimated_hours"`
}

// ScheduleOptimization represents an optimized schedule
type ScheduleOptimization struct {
	OptimizedTasks   []*OptimizedTask `json:"optimized_tasks"`
	TotalDuration    time.Duration    `json:"total_duration"`
	ResourceUtilization float64       `json:"resource_utilization"`
	ConflictCount    int              `json:"conflict_count"`
	OptimizationScore float64         `json:"optimization_score"`
	Recommendations  []string         `json:"recommendations"`
	GeneratedAt      time.Time        `json:"generated_at"`
}

// OptimizedTask represents a task with optimized scheduling
type OptimizedTask struct {
	TaskID           uuid.UUID     `json:"task_id"`
	AssignedUserID   uuid.UUID     `json:"assigned_user_id"`
	ScheduledStart   time.Time     `json:"scheduled_start"`
	ScheduledEnd     time.Time     `json:"scheduled_end"`
	EstimatedDuration time.Duration `json:"estimated_duration"`
	Priority         int           `json:"priority"`
	Dependencies     []uuid.UUID   `json:"dependencies"`
	Conflicts        []string      `json:"conflicts"`
}

// TaskRisk represents a potential risk for a task
type TaskRisk struct {
	Type         RiskType  `json:"type"`
	Severity     RiskLevel `json:"severity"`
	Probability  float64   `json:"probability"`
	Impact       float64   `json:"impact"`
	Description  string    `json:"description"`
	Mitigation   string    `json:"mitigation"`
	DetectedAt   time.Time `json:"detected_at"`
}

// RiskType defines types of task risks
type RiskType string

const (
	RiskTypeDelay        RiskType = "delay"
	RiskTypeOverrun      RiskType = "overrun"
	RiskTypeQuality      RiskType = "quality"
	RiskTypeResource     RiskType = "resource"
	RiskTypeDependency   RiskType = "dependency"
	RiskTypeComplexity   RiskType = "complexity"
	RiskTypeSkillGap     RiskType = "skill_gap"
)

// RiskLevel defines risk severity levels
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// ProductivityAnalysis represents productivity analysis results
type ProductivityAnalysis struct {
	UserID              uuid.UUID `json:"user_id"`
	Period              time.Duration `json:"period"`
	TasksCompleted      int       `json:"tasks_completed"`
	AverageTaskTime     time.Duration `json:"average_task_time"`
	ProductivityScore   float64   `json:"productivity_score"`
	EfficiencyTrend     string    `json:"efficiency_trend"`
	StrengthAreas       []string  `json:"strength_areas"`
	ImprovementAreas    []string  `json:"improvement_areas"`
	Recommendations     []string  `json:"recommendations"`
	ComparisonToTeam    float64   `json:"comparison_to_team"`
	BurnoutRisk         float64   `json:"burnout_risk"`
	OptimalWorkload     float64   `json:"optimal_workload"`
}

// ProjectCompletionPrediction represents project completion prediction
type ProjectCompletionPrediction struct {
	ProjectID           uuid.UUID `json:"project_id"`
	PredictedCompletion time.Time `json:"predicted_completion"`
	Confidence          float64   `json:"confidence"`
	RiskFactors         []string  `json:"risk_factors"`
	CriticalPath        []uuid.UUID `json:"critical_path"`
	ResourceBottlenecks []string  `json:"resource_bottlenecks"`
	Recommendations     []string  `json:"recommendations"`
}

// TaskDetails represents extracted task details from NLP
type TaskDetails struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Priority    entities.TaskPriority  `json:"priority"`
	Urgency     entities.TaskUrgency   `json:"urgency"`
	Complexity  entities.TaskComplexity `json:"complexity"`
	Type        entities.TaskType      `json:"type"`
	DueDate     *time.Time             `json:"due_date,omitempty"`
	EstimatedHours float64             `json:"estimated_hours"`
	Skills      []string               `json:"skills"`
	Tags        []string               `json:"tags"`
	Assignees   []string               `json:"assignees"`
	Dependencies []string              `json:"dependencies"`
}

// SentimentAnalysis represents sentiment analysis results
type SentimentAnalysis struct {
	Sentiment   string  `json:"sentiment"`
	Score       float64 `json:"score"`
	Confidence  float64 `json:"confidence"`
	Emotions    map[string]float64 `json:"emotions"`
}

// NewAITaskAutomationService creates a new AI task automation service
func NewAITaskAutomationService(
	taskRepo repositories.TaskRepository,
	userRepo repositories.UserRepository,
	projectRepo repositories.ProjectRepository,
	workflowRepo repositories.WorkflowRepository,
	mlService MachineLearningService,
	nlpService NaturalLanguageService,
	eventPublisher EventPublisher,
	logger Logger,
) *AITaskAutomationService {
	return &AITaskAutomationService{
		taskRepo:       taskRepo,
		userRepo:       userRepo,
		projectRepo:    projectRepo,
		workflowRepo:   workflowRepo,
		mlService:      mlService,
		nlpService:     nlpService,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// EnhanceTask enhances a task using AI analysis
func (ai *AITaskAutomationService) EnhanceTask(ctx context.Context, task *entities.Task) error {
	ai.logger.Info("Enhancing task with AI", "task_id", task.ID, "title", task.Title)

	// Extract additional details from description using NLP
	if task.Description != "" {
		details, err := ai.nlpService.ExtractTaskDetails(ctx, task.Description)
		if err != nil {
			ai.logger.Warn("Failed to extract task details", "task_id", task.ID, "error", err)
		} else {
			// Apply extracted details if not already set
			if task.Priority == "" && details.Priority != "" {
				task.Priority = details.Priority
			}
			if task.Urgency == "" && details.Urgency != "" {
				task.Urgency = details.Urgency
			}
			if task.Complexity == "" && details.Complexity != "" {
				task.Complexity = details.Complexity
			}
			if task.EstimatedHours == 0 && details.EstimatedHours > 0 {
				task.EstimatedHours = details.EstimatedHours
			}
			if len(task.Skills) == 0 && len(details.Skills) > 0 {
				task.Skills = details.Skills
			}
			if len(task.Tags) == 0 && len(details.Tags) > 0 {
				task.Tags = details.Tags
			}
		}
	}

	// Predict task duration if not estimated
	if task.EstimatedHours == 0 {
		duration, err := ai.mlService.PredictTaskDuration(ctx, task)
		if err != nil {
			ai.logger.Warn("Failed to predict task duration", "task_id", task.ID, "error", err)
		} else {
			task.EstimatedHours = duration
		}
	}

	// Predict complexity if not set
	if task.Complexity == "" {
		complexity, err := ai.mlService.PredictTaskComplexity(ctx, task)
		if err != nil {
			ai.logger.Warn("Failed to predict task complexity", "task_id", task.ID, "error", err)
		} else {
			task.Complexity = complexity
		}
	}

	// Suggest tags using NLP
	suggestedTags, err := ai.nlpService.SuggestTags(ctx, task)
	if err != nil {
		ai.logger.Warn("Failed to suggest tags", "task_id", task.ID, "error", err)
	} else {
		// Merge with existing tags
		tagSet := make(map[string]bool)
		for _, tag := range task.Tags {
			tagSet[tag] = true
		}
		for _, tag := range suggestedTags {
			if !tagSet[tag] {
				task.Tags = append(task.Tags, tag)
				tagSet[tag] = true
			}
		}
	}

	// Detect potential risks
	risks, err := ai.mlService.DetectRisks(ctx, task)
	if err != nil {
		ai.logger.Warn("Failed to detect task risks", "task_id", task.ID, "error", err)
	} else {
		// Store risks in metadata
		if task.Metadata == nil {
			task.Metadata = make(map[string]interface{})
		}
		task.Metadata["ai_detected_risks"] = risks
	}

	ai.logger.Info("Task enhanced successfully", "task_id", task.ID)
	return nil
}

// RecommendAssignees recommends the best assignees for a task
func (ai *AITaskAutomationService) RecommendAssignees(ctx context.Context, task *entities.Task) ([]uuid.UUID, error) {
	ai.logger.Info("Recommending assignees", "task_id", task.ID)

	// Get available users
	filter := &repositories.UserFilter{
		IsActive:    &[]bool{true}[0],
		MaxWorkload: &[]float64{80.0}[0], // Users with less than 80% workload
	}
	
	users, err := ai.userRepo.List(ctx, filter)
	if err != nil {
		ai.logger.Error("Failed to get available users", err)
		return nil, err
	}

	// Get ML recommendations
	recommendations, err := ai.mlService.RecommendAssignees(ctx, task, users)
	if err != nil {
		ai.logger.Error("Failed to get ML recommendations", err, "task_id", task.ID)
		return nil, err
	}

	// Sort by score and return top recommendations
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	var assigneeIDs []uuid.UUID
	maxRecommendations := 3 // Limit to top 3 recommendations
	for i, rec := range recommendations {
		if i >= maxRecommendations {
			break
		}
		if rec.Score > 0.6 { // Only recommend if score is above threshold
			assigneeIDs = append(assigneeIDs, rec.UserID)
		}
	}

	ai.logger.Info("Generated assignee recommendations", "task_id", task.ID, "count", len(assigneeIDs))
	return assigneeIDs, nil
}

// OptimizeTaskSchedule optimizes the schedule for a set of tasks
func (ai *AITaskAutomationService) OptimizeTaskSchedule(ctx context.Context, taskIDs []uuid.UUID) (*ScheduleOptimization, error) {
	ai.logger.Info("Optimizing task schedule", "task_count", len(taskIDs))

	// Get tasks
	var tasks []*entities.Task
	for _, taskID := range taskIDs {
		task, err := ai.taskRepo.GetByID(ctx, taskID)
		if err != nil {
			ai.logger.Warn("Failed to get task for optimization", "task_id", taskID, "error", err)
			continue
		}
		tasks = append(tasks, task)
	}

	// Get available users
	users, err := ai.userRepo.List(ctx, &repositories.UserFilter{
		IsActive: &[]bool{true}[0],
	})
	if err != nil {
		ai.logger.Error("Failed to get users for optimization", err)
		return nil, err
	}

	// Use ML service to optimize schedule
	optimization, err := ai.mlService.OptimizeSchedule(ctx, tasks, users)
	if err != nil {
		ai.logger.Error("Failed to optimize schedule", err)
		return nil, err
	}

	ai.logger.Info("Schedule optimization completed", 
		"total_duration", optimization.TotalDuration,
		"resource_utilization", optimization.ResourceUtilization,
		"conflicts", optimization.ConflictCount)

	return optimization, nil
}

// AnalyzeUserProductivity analyzes user productivity using AI
func (ai *AITaskAutomationService) AnalyzeUserProductivity(ctx context.Context, userID uuid.UUID, period time.Duration) (*ProductivityAnalysis, error) {
	ai.logger.Info("Analyzing user productivity", "user_id", userID, "period", period)

	analysis, err := ai.mlService.AnalyzeProductivity(ctx, userID, period)
	if err != nil {
		ai.logger.Error("Failed to analyze productivity", err, "user_id", userID)
		return nil, err
	}

	ai.logger.Info("Productivity analysis completed", 
		"user_id", userID,
		"productivity_score", analysis.ProductivityScore,
		"tasks_completed", analysis.TasksCompleted)

	return analysis, nil
}

// PredictProjectCompletion predicts when a project will be completed
func (ai *AITaskAutomationService) PredictProjectCompletion(ctx context.Context, projectID uuid.UUID) (*ProjectCompletionPrediction, error) {
	ai.logger.Info("Predicting project completion", "project_id", projectID)

	prediction, err := ai.mlService.PredictProjectCompletion(ctx, projectID)
	if err != nil {
		ai.logger.Error("Failed to predict project completion", err, "project_id", projectID)
		return nil, err
	}

	ai.logger.Info("Project completion prediction completed", 
		"project_id", projectID,
		"predicted_completion", prediction.PredictedCompletion,
		"confidence", prediction.Confidence)

	return prediction, nil
}

// GetTaskRecommendations provides AI-powered recommendations for a task
func (ai *AITaskAutomationService) GetTaskRecommendations(ctx context.Context, task *entities.Task) ([]string, error) {
	ai.logger.Info("Getting task recommendations", "task_id", task.ID)

	var recommendations []string

	// Check for potential issues and provide recommendations
	if task.EstimatedHours == 0 {
		recommendations = append(recommendations, "Consider adding time estimation for better planning")
	}

	if len(task.Skills) == 0 {
		recommendations = append(recommendations, "Add required skills to help with assignment")
	}

	if task.DueDate == nil {
		recommendations = append(recommendations, "Set a due date to improve prioritization")
	}

	if task.Priority == entities.PriorityMedium {
		recommendations = append(recommendations, "Review task priority based on business impact")
	}

	// Check for complexity vs estimation mismatch
	if task.Complexity == entities.ComplexityComplex && task.EstimatedHours < 8 {
		recommendations = append(recommendations, "Complex tasks typically require more time - consider increasing estimation")
	}

	// Check for overdue risk
	if task.DueDate != nil && task.DueDate.Before(time.Now().Add(24*time.Hour)) && task.Status == entities.StatusTodo {
		recommendations = append(recommendations, "Task is due soon - consider starting immediately or adjusting timeline")
	}

	// Analyze task description for clarity
	if len(task.Description) < 50 {
		recommendations = append(recommendations, "Add more detailed description for better understanding")
	}

	// Check for dependency issues
	if len(task.Dependencies) > 3 {
		recommendations = append(recommendations, "Task has many dependencies - consider breaking it down")
	}

	ai.logger.Info("Generated task recommendations", "task_id", task.ID, "count", len(recommendations))
	return recommendations, nil
}

// AutoAssignTasks automatically assigns tasks based on AI recommendations
func (ai *AITaskAutomationService) AutoAssignTasks(ctx context.Context, taskIDs []uuid.UUID) error {
	ai.logger.Info("Auto-assigning tasks", "task_count", len(taskIDs))

	for _, taskID := range taskIDs {
		task, err := ai.taskRepo.GetByID(ctx, taskID)
		if err != nil {
			ai.logger.Error("Failed to get task for auto-assignment", err, "task_id", taskID)
			continue
		}

		// Skip if already assigned
		if len(task.Assignments) > 0 {
			continue
		}

		// Get recommendations
		assigneeIDs, err := ai.RecommendAssignees(ctx, task)
		if err != nil {
			ai.logger.Error("Failed to get assignee recommendations", err, "task_id", taskID)
			continue
		}

		// Assign to top recommendation
		if len(assigneeIDs) > 0 {
			assignment := &entities.TaskAssignment{
				ID:         uuid.New(),
				TaskID:     taskID,
				UserID:     assigneeIDs[0],
				Role:       entities.RoleAssignee,
				Allocation: 100.0,
				AssignedAt: time.Now(),
				AssignedBy: uuid.New(), // System assignment
				Status:     entities.AssignmentActive,
				IsActive:   true,
			}

			if err := ai.taskRepo.AddAssignment(ctx, assignment); err != nil {
				ai.logger.Error("Failed to create auto-assignment", err, "task_id", taskID)
				continue
			}

			ai.logger.Info("Task auto-assigned", "task_id", taskID, "assignee_id", assigneeIDs[0])
		}
	}

	return nil
}

// GenerateTaskSummary generates an AI-powered summary of a task
func (ai *AITaskAutomationService) GenerateTaskSummary(ctx context.Context, task *entities.Task) (string, error) {
	ai.logger.Info("Generating task summary", "task_id", task.ID)

	summary, err := ai.nlpService.GenerateTaskSummary(ctx, task)
	if err != nil {
		ai.logger.Error("Failed to generate task summary", err, "task_id", task.ID)
		return "", err
	}

	ai.logger.Info("Task summary generated", "task_id", task.ID, "length", len(summary))
	return summary, nil
}

// AnalyzeTaskSentiment analyzes the sentiment of task comments and descriptions
func (ai *AITaskAutomationService) AnalyzeTaskSentiment(ctx context.Context, task *entities.Task) (*SentimentAnalysis, error) {
	ai.logger.Info("Analyzing task sentiment", "task_id", task.ID)

	// Combine task description and recent comments
	var textToAnalyze strings.Builder
	textToAnalyze.WriteString(task.Description)

	// Add recent comments
	comments, err := ai.taskRepo.GetTaskComments(ctx, task.ID)
	if err == nil {
		for _, comment := range comments {
			if comment.CreatedAt.After(time.Now().Add(-7 * 24 * time.Hour)) { // Last 7 days
				textToAnalyze.WriteString(" ")
				textToAnalyze.WriteString(comment.Content)
			}
		}
	}

	sentiment, err := ai.nlpService.AnalyzeSentiment(ctx, textToAnalyze.String())
	if err != nil {
		ai.logger.Error("Failed to analyze sentiment", err, "task_id", task.ID)
		return nil, err
	}

	ai.logger.Info("Task sentiment analyzed", "task_id", task.ID, "sentiment", sentiment.Sentiment, "score", sentiment.Score)
	return sentiment, nil
}

// Helper methods for ML service implementation (simplified versions)

// SimplePredictTaskDuration provides a simple duration prediction
func (ai *AITaskAutomationService) SimplePredictTaskDuration(ctx context.Context, task *entities.Task) (float64, error) {
	// Simple heuristic-based prediction
	baseHours := 4.0 // Default base hours

	// Adjust based on complexity
	switch task.Complexity {
	case entities.ComplexityTrivial:
		baseHours = 1.0
	case entities.ComplexitySimple:
		baseHours = 2.0
	case entities.ComplexityMedium:
		baseHours = 4.0
	case entities.ComplexityComplex:
		baseHours = 8.0
	case entities.ComplexityExpert:
		baseHours = 16.0
	}

	// Adjust based on type
	switch task.Type {
	case entities.TaskTypeBug:
		baseHours *= 0.8
	case entities.TaskTypeFeature:
		baseHours *= 1.5
	case entities.TaskTypeResearch:
		baseHours *= 2.0
	}

	// Adjust based on priority (higher priority might need more careful work)
	switch task.Priority {
	case entities.PriorityCritical:
		baseHours *= 1.2
	case entities.PriorityHigh:
		baseHours *= 1.1
	}

	return baseHours, nil
}

// SimpleRecommendAssignees provides simple assignee recommendations
func (ai *AITaskAutomationService) SimpleRecommendAssignees(ctx context.Context, task *entities.Task, users []*entities.User) ([]*UserRecommendation, error) {
	var recommendations []*UserRecommendation

	for _, user := range users {
		if !user.IsAvailable() {
			continue
		}

		score := ai.calculateUserScore(user, task)
		if score > 0.3 { // Minimum threshold
			recommendations = append(recommendations, &UserRecommendation{
				UserID:           user.ID,
				User:             user,
				Score:            score,
				Confidence:       0.8,
				Reasoning:        ai.generateRecommendationReasoning(user, task, score),
				SkillMatch:       ai.calculateSkillMatch(user, task),
				AvailabilityScore: ai.calculateAvailabilityScore(user),
				WorkloadScore:    ai.calculateWorkloadScore(user),
				PerformanceScore: 0.8, // Simplified
			})
		}
	}

	return recommendations, nil
}

func (ai *AITaskAutomationService) calculateUserScore(user *entities.User, task *entities.Task) float64 {
	skillScore := ai.calculateSkillMatch(user, task)
	availabilityScore := ai.calculateAvailabilityScore(user)
	workloadScore := ai.calculateWorkloadScore(user)

	// Weighted average
	return (skillScore*0.4 + availabilityScore*0.3 + workloadScore*0.3)
}

func (ai *AITaskAutomationService) calculateSkillMatch(user *entities.User, task *entities.Task) float64 {
	if len(task.Skills) == 0 {
		return 0.8 // Default score if no specific skills required
	}

	userSkillsMap := make(map[string]bool)
	for _, skill := range user.Skills {
		userSkillsMap[strings.ToLower(skill)] = true
	}

	matchCount := 0
	for _, skill := range task.Skills {
		if userSkillsMap[strings.ToLower(skill)] {
			matchCount++
		}
	}

	return float64(matchCount) / float64(len(task.Skills))
}

func (ai *AITaskAutomationService) calculateAvailabilityScore(user *entities.User) float64 {
	if !user.IsAvailable() {
		return 0.0
	}
	return 1.0 // Simplified - could check working hours, time zones, etc.
}

func (ai *AITaskAutomationService) calculateWorkloadScore(user *entities.User) float64 {
	workload := user.GetCurrentWorkload()
	if workload >= 100 {
		return 0.0
	}
	return math.Max(0, (100-workload)/100)
}

func (ai *AITaskAutomationService) generateRecommendationReasoning(user *entities.User, task *entities.Task, score float64) string {
	reasons := []string{}

	skillMatch := ai.calculateSkillMatch(user, task)
	if skillMatch > 0.8 {
		reasons = append(reasons, "excellent skill match")
	} else if skillMatch > 0.6 {
		reasons = append(reasons, "good skill match")
	}

	workload := user.GetCurrentWorkload()
	if workload < 50 {
		reasons = append(reasons, "low current workload")
	} else if workload < 80 {
		reasons = append(reasons, "moderate workload")
	}

	if len(reasons) == 0 {
		return "general availability and capability"
	}

	return strings.Join(reasons, ", ")
}
