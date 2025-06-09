package ai

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/kitchen/application"
	"github.com/DimaJoyti/go-coffee/internal/kitchen/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// OptimizerServiceImpl implements application.OptimizerService
type OptimizerServiceImpl struct {
	logger *logger.Logger
}

// NewOptimizerService creates a new AI optimizer service
func NewOptimizerService(logger *logger.Logger) application.OptimizerService {
	return &OptimizerServiceImpl{
		logger: logger,
	}
}

// OptimizeWorkflow optimizes workflow for given orders
func (s *OptimizerServiceImpl) OptimizeWorkflow(ctx context.Context, orders []*domain.KitchenOrder) (*domain.WorkflowOptimization, error) {
	if len(orders) == 0 {
		return nil, fmt.Errorf("no orders to optimize")
	}

	// For now, optimize the first order as an example
	order := orders[0]
	
	s.logger.WithField("order_id", order.ID()).Info("Optimizing workflow")

	optimization := domain.NewWorkflowOptimization(order.ID())
	
	// Analyze required stations
	requiredStations := order.GetRequiredStations()
	
	// Create optimized steps based on required stations
	stepID := 1
	totalTime := int32(0)
	
	for _, stationType := range requiredStations {
		step := &domain.WorkflowStep{
			StepID:        fmt.Sprintf("step_%d", stepID),
			StationType:   stationType,
			EstimatedTime: s.estimateStationTime(stationType, order),
			RequiredSkill: s.getRequiredSkillForStation(stationType),
			Dependencies:  s.getDependencies(stepID, requiredStations),
			CanParallelize: s.canParallelize(stationType),
		}
		
		optimization.AddStep(step)
		totalTime += step.EstimatedTime
		stepID++
	}
	
	optimization.CalculateEstimatedTime()
	
	// Calculate efficiency gain (simplified)
	baselineTime := int32(len(requiredStations) * 120) // 2 minutes per station baseline
	if baselineTime > 0 {
		optimization.EfficiencyGain = float32(baselineTime-totalTime) / float32(baselineTime) * 100
	}
	
	// Add recommendations
	optimization.AddRecommendation("Prioritize parallel processing where possible")
	optimization.AddRecommendation("Ensure staff specialization matches station requirements")
	
	s.logger.WithFields(map[string]interface{}{
		"order_id":        order.ID(),
		"estimated_time":  optimization.EstimatedTime,
		"efficiency_gain": optimization.EfficiencyGain,
	}).Info("Workflow optimization completed")
	
	return optimization, nil
}

// PredictPreparationTime predicts preparation time for an order
func (s *OptimizerServiceImpl) PredictPreparationTime(ctx context.Context, order *domain.KitchenOrder) (int32, error) {
	s.logger.WithField("order_id", order.ID()).Info("Predicting preparation time")

	baseTime := int32(60) // 1 minute base time
	
	// Add time based on items
	for _, item := range order.Items() {
		itemTime := s.estimateItemTime(item)
		baseTime += itemTime * item.Quantity()
	}
	
	// Add complexity factor based on number of required stations
	requiredStations := order.GetRequiredStations()
	complexityFactor := float32(len(requiredStations)) * 0.2
	adjustedTime := float32(baseTime) * (1.0 + complexityFactor)
	
	// Add priority adjustment
	switch order.Priority() {
	case domain.OrderPriorityUrgent:
		adjustedTime *= 0.8 // 20% faster for urgent orders
	case domain.OrderPriorityHigh:
		adjustedTime *= 0.9 // 10% faster for high priority
	case domain.OrderPriorityLow:
		adjustedTime *= 1.2 // 20% slower for low priority
	}
	
	predictedTime := int32(adjustedTime)
	
	s.logger.WithFields(map[string]interface{}{
		"order_id":       order.ID(),
		"predicted_time": predictedTime,
		"base_time":      baseTime,
		"complexity":     complexityFactor,
	}).Info("Preparation time predicted")
	
	return predictedTime, nil
}

// AllocateStaff allocates staff to orders optimally
func (s *OptimizerServiceImpl) AllocateStaff(ctx context.Context, orders []*domain.KitchenOrder, staff []*domain.Staff) (*domain.StaffAllocation, error) {
	s.logger.WithFields(map[string]interface{}{
		"orders_count": len(orders),
		"staff_count":  len(staff),
	}).Info("Allocating staff to orders")

	allocation := &domain.StaffAllocation{
		Allocations:     []*domain.StaffOrderAllocation{},
		LoadBalance:     make(map[string]float32),
		Recommendations: []string{},
	}
	
	// Sort orders by priority
	sortedOrders := make([]*domain.KitchenOrder, len(orders))
	copy(sortedOrders, orders)
	sort.Slice(sortedOrders, func(i, j int) bool {
		return sortedOrders[i].Priority() > sortedOrders[j].Priority()
	})
	
	// Sort staff by availability and skill
	availableStaff := make([]*domain.Staff, 0)
	for _, s := range staff {
		if s.CanAcceptOrder() {
			availableStaff = append(availableStaff, s)
		}
	}
	
	sort.Slice(availableStaff, func(i, j int) bool {
		return availableStaff[i].GetWorkload() < availableStaff[j].GetWorkload()
	})
	
	// Allocate staff to orders
	for _, order := range sortedOrders {
		if order.Status() != domain.OrderStatusPending {
			continue
		}
		
		requiredStations := order.GetRequiredStations()
		bestStaff := s.findBestStaffForOrder(order, availableStaff, requiredStations)
		
		if bestStaff != nil {
			estimatedTime, _ := s.PredictPreparationTime(ctx, order)
			efficiency := bestStaff.GetEfficiencyForStation(requiredStations[0])
			
			orderAllocation := domain.NewStaffOrderAllocation(
				bestStaff.ID(),
				order.ID(),
				requiredStations[0],
				estimatedTime,
				efficiency,
				"Optimal skill match",
			)
			
			allocation.Allocations = append(allocation.Allocations, orderAllocation)
			allocation.LoadBalance[bestStaff.ID()] = bestStaff.GetWorkload()
		}
	}
	
	// Calculate utilization rate
	if len(staff) > 0 {
		totalWorkload := float32(0)
		for _, s := range staff {
			totalWorkload += s.GetWorkload()
		}
		allocation.UtilizationRate = totalWorkload / float32(len(staff))
	}
	
	// Add recommendations
	if allocation.UtilizationRate > 0.8 {
		allocation.Recommendations = append(allocation.Recommendations, "Consider adding more staff - high utilization detected")
	}
	if len(allocation.Allocations) < len(orders) {
		allocation.Recommendations = append(allocation.Recommendations, "Some orders could not be allocated - check staff availability")
	}
	
	s.logger.WithFields(map[string]interface{}{
		"allocations_count": len(allocation.Allocations),
		"utilization_rate":  allocation.UtilizationRate,
	}).Info("Staff allocation completed")
	
	return allocation, nil
}

// OptimizeStaffSchedule optimizes staff schedule
func (s *OptimizerServiceImpl) OptimizeStaffSchedule(ctx context.Context, staff []*domain.Staff, timeWindow time.Duration) (*application.StaffScheduleOptimization, error) {
	s.logger.WithFields(map[string]interface{}{
		"staff_count":  len(staff),
		"time_window": timeWindow.String(),
	}).Info("Optimizing staff schedule")

	optimization := &application.StaffScheduleOptimization{
		Schedule:        make(map[string]*application.StaffShift),
		Recommendations: []string{},
		CreatedAt:       time.Now(),
	}
	
	// Create shifts for each staff member
	for _, s := range staff {
		if !s.IsAvailable() {
			continue
		}
		
		// Assign to their primary specialization
		primaryStation := s.Specializations()[0]
		
		shift := &application.StaffShift{
			StaffID:   s.ID(),
			StartTime: time.Now(),
			EndTime:   time.Now().Add(timeWindow),
			Station:   primaryStation,
			Load:      s.GetWorkload(),
		}
		
		optimization.Schedule[s.ID()] = shift
	}
	
	// Calculate efficiency gain (simplified)
	optimization.EfficiencyGain = 15.0 // 15% efficiency gain estimate
	optimization.CostReduction = 10.0  // 10% cost reduction estimate
	
	optimization.Recommendations = append(optimization.Recommendations, "Balance workload across all staff members")
	optimization.Recommendations = append(optimization.Recommendations, "Consider cross-training for better flexibility")
	
	return optimization, nil
}

// OptimizeEquipmentUsage optimizes equipment usage
func (s *OptimizerServiceImpl) OptimizeEquipmentUsage(ctx context.Context, equipment []*domain.Equipment, orders []*domain.KitchenOrder) (*application.EquipmentOptimization, error) {
	s.logger.WithFields(map[string]interface{}{
		"equipment_count": len(equipment),
		"orders_count":    len(orders),
	}).Info("Optimizing equipment usage")

	optimization := &application.EquipmentOptimization{
		Allocations:     make(map[string]*application.EquipmentAllocation),
		Recommendations: []string{},
		CreatedAt:       time.Now(),
	}
	
	// Allocate equipment to orders
	for _, eq := range equipment {
		if !eq.IsAvailable() {
			continue
		}
		
		// Find orders that need this equipment type
		var suitableOrders []string
		for _, order := range orders {
			requiredStations := order.GetRequiredStations()
			for _, station := range requiredStations {
				if station == eq.StationType() {
					suitableOrders = append(suitableOrders, order.ID())
					break
				}
			}
		}
		
		if len(suitableOrders) > 0 {
			allocation := &application.EquipmentAllocation{
				EquipmentID: eq.ID(),
				OrderIDs:    suitableOrders,
				StartTime:   time.Now(),
				EndTime:     time.Now().Add(2 * time.Hour), // 2 hour window
				Load:        eq.GetUtilizationRate(),
			}
			
			optimization.Allocations[eq.ID()] = allocation
		}
	}
	
	// Calculate utilization gain
	optimization.UtilizationGain = 20.0 // 20% utilization improvement estimate
	
	optimization.Recommendations = append(optimization.Recommendations, "Schedule maintenance during low-demand periods")
	optimization.Recommendations = append(optimization.Recommendations, "Monitor equipment efficiency scores regularly")
	
	return optimization, nil
}

// PredictEquipmentLoad predicts equipment load
func (s *OptimizerServiceImpl) PredictEquipmentLoad(ctx context.Context, equipment *domain.Equipment, timeWindow time.Duration) (float32, error) {
	// Simplified prediction based on current load and time window
	currentLoad := equipment.GetUtilizationRate()
	
	// Predict slight increase during peak hours
	hours := timeWindow.Hours()
	if hours > 0 && hours <= 4 {
		// Peak hours - increase load
		return math.Min(float64(currentLoad*1.3), 1.0), nil
	}
	
	// Off-peak hours - maintain or slightly decrease
	return currentLoad * 0.9, nil
}

// PredictCapacity predicts kitchen capacity
func (s *OptimizerServiceImpl) PredictCapacity(ctx context.Context, timeWindow time.Duration) (*application.CapacityPrediction, error) {
	s.logger.WithField("time_window", timeWindow.String()).Info("Predicting kitchen capacity")

	prediction := &application.CapacityPrediction{
		TimeWindow:       timeWindow,
		CurrentCapacity:  50, // Assume current capacity of 50 orders
		RequiredCapacity: 60, // Predict need for 60 orders
		Confidence:       0.85,
		Recommendations:  []string{},
		PredictedAt:      time.Now(),
	}
	
	// Calculate predicted orders based on time window
	hours := timeWindow.Hours()
	prediction.PredictedOrders = int32(hours * 10) // 10 orders per hour estimate
	
	prediction.CapacityGap = prediction.RequiredCapacity - prediction.CurrentCapacity
	
	if prediction.CapacityGap > 0 {
		prediction.Recommendations = append(prediction.Recommendations, "Consider adding temporary staff")
		prediction.Recommendations = append(prediction.Recommendations, "Optimize current workflows")
	}
	
	return prediction, nil
}

// AnalyzeBottlenecks analyzes system bottlenecks
func (s *OptimizerServiceImpl) AnalyzeBottlenecks(ctx context.Context) (*application.BottleneckAnalysis, error) {
	s.logger.Info("Analyzing system bottlenecks")

	analysis := &application.BottleneckAnalysis{
		Bottlenecks:   []*application.Bottleneck{},
		OverallImpact: 0.3, // 30% overall impact
		Recommendations: []string{
			"Monitor equipment utilization rates",
			"Balance staff workloads",
			"Optimize order queue management",
		},
		AnalyzedAt: time.Now(),
	}
	
	// Example bottlenecks
	bottlenecks := []*application.Bottleneck{
		{
			Type:       "equipment",
			ResourceID: "espresso_machine_1",
			Severity:   0.7,
			Impact:     "High utilization causing delays",
			Suggestions: []string{"Add backup equipment", "Schedule maintenance"},
		},
		{
			Type:       "staff",
			ResourceID: "barista_1",
			Severity:   0.5,
			Impact:     "Overloaded during peak hours",
			Suggestions: []string{"Add staff during peak hours", "Cross-train other staff"},
		},
	}
	
	analysis.Bottlenecks = bottlenecks
	
	return analysis, nil
}

// AnalyzePerformance analyzes performance
func (s *OptimizerServiceImpl) AnalyzePerformance(ctx context.Context, period *application.TimePeriod) (*application.PerformanceAnalysis, error) {
	s.logger.WithField("period", period).Info("Analyzing performance")

	analysis := &application.PerformanceAnalysis{
		Period:          period,
		Metrics:         make(map[string]float32),
		Trends:          make(map[string][]float32),
		Comparisons:     make(map[string]float32),
		Insights:        []string{},
		Recommendations: []*application.Recommendation{},
		AnalyzedAt:      time.Now(),
	}
	
	// Example metrics
	analysis.Metrics["efficiency"] = 0.85
	analysis.Metrics["throughput"] = 45.0
	analysis.Metrics["quality"] = 0.92
	
	// Example trends
	analysis.Trends["efficiency"] = []float32{0.80, 0.82, 0.85, 0.87, 0.85}
	analysis.Trends["throughput"] = []float32{40.0, 42.0, 45.0, 47.0, 45.0}
	
	// Example insights
	analysis.Insights = append(analysis.Insights, "Efficiency has improved by 5% over the period")
	analysis.Insights = append(analysis.Insights, "Throughput shows steady growth with occasional dips")
	
	return analysis, nil
}

// GenerateRecommendations generates optimization recommendations
func (s *OptimizerServiceImpl) GenerateRecommendations(ctx context.Context, context *application.OptimizationContext) ([]*application.Recommendation, error) {
	s.logger.Info("Generating optimization recommendations")

	recommendations := []*application.Recommendation{
		{
			ID:          "rec_1",
			Type:        "efficiency",
			Priority:    1,
			Title:       "Optimize Staff Allocation",
			Description: "Reallocate staff based on current queue and specializations",
			Impact:      "High",
			Effort:      "Medium",
			Data:        make(map[string]interface{}),
			CreatedAt:   time.Now(),
		},
		{
			ID:          "rec_2",
			Type:        "capacity",
			Priority:    2,
			Title:       "Equipment Maintenance",
			Description: "Schedule maintenance for underperforming equipment",
			Impact:      "Medium",
			Effort:      "Low",
			Data:        make(map[string]interface{}),
			CreatedAt:   time.Now(),
		},
	}
	
	return recommendations, nil
}

// Helper methods

func (s *OptimizerServiceImpl) estimateStationTime(stationType domain.StationType, order *domain.KitchenOrder) int32 {
	baseTime := map[domain.StationType]int32{
		domain.StationTypeEspresso: 90,  // 1.5 minutes
		domain.StationTypeGrinder:  30,  // 30 seconds
		domain.StationTypeSteamer:  60,  // 1 minute
		domain.StationTypeAssembly: 45,  // 45 seconds
	}
	
	if time, exists := baseTime[stationType]; exists {
		// Adjust for order complexity
		complexity := float32(order.GetTotalQuantity()) * 0.1
		return int32(float32(time) * (1.0 + complexity))
	}
	
	return 60 // Default 1 minute
}

func (s *OptimizerServiceImpl) estimateItemTime(item *domain.OrderItem) int32 {
	// Base time per item type (simplified)
	baseTime := int32(30) // 30 seconds per item
	
	// Adjust based on complexity (number of requirements)
	complexity := len(item.Requirements())
	return baseTime + int32(complexity*10)
}

func (s *OptimizerServiceImpl) getRequiredSkillForStation(stationType domain.StationType) float32 {
	skillRequirements := map[domain.StationType]float32{
		domain.StationTypeEspresso: 8.0, // High skill required
		domain.StationTypeGrinder:  5.0, // Medium skill
		domain.StationTypeSteamer:  7.0, // High skill
		domain.StationTypeAssembly: 6.0, // Medium-high skill
	}
	
	if skill, exists := skillRequirements[stationType]; exists {
		return skill
	}
	
	return 5.0 // Default medium skill
}

func (s *OptimizerServiceImpl) getDependencies(stepID int, stations []domain.StationType) []string {
	// Simplified dependency logic
	if stepID == 1 {
		return []string{} // First step has no dependencies
	}
	
	return []string{fmt.Sprintf("step_%d", stepID-1)} // Depends on previous step
}

func (s *OptimizerServiceImpl) canParallelize(stationType domain.StationType) bool {
	// Some stations can work in parallel
	parallelizable := map[domain.StationType]bool{
		domain.StationTypeGrinder:  true,
		domain.StationTypeAssembly: true,
		domain.StationTypeEspresso: false, // Usually sequential
		domain.StationTypeSteamer:  false, // Usually sequential
	}
	
	if canParallel, exists := parallelizable[stationType]; exists {
		return canParallel
	}
	
	return false
}

func (s *OptimizerServiceImpl) findBestStaffForOrder(order *domain.KitchenOrder, staff []*domain.Staff, requiredStations []domain.StationType) *domain.Staff {
	var bestStaff *domain.Staff
	bestScore := float32(-1)
	
	for _, s := range staff {
		if !s.CanAcceptOrder() {
			continue
		}
		
		// Calculate score based on specialization match and workload
		score := float32(0)
		for _, station := range requiredStations {
			if s.CanHandleStation(station) {
				score += s.GetEfficiencyForStation(station)
				break
			}
		}
		
		// Prefer less loaded staff
		score = score * (1.0 - s.GetWorkload())
		
		if score > bestScore {
			bestScore = score
			bestStaff = s
		}
	}
	
	return bestStaff
}
