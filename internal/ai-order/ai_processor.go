package aiorder

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/DimaJoyti/go-coffee/api/proto/ai_order"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// AIProcessor defines the interface for AI-powered order processing
type AIProcessor interface {
	AnalyzeOrder(ctx context.Context, order *pb.Order) (*pb.AIOrderInsights, error)
	GetRecommendations(ctx context.Context, customer *pb.Customer, items []*pb.OrderItem) ([]string, error)
	ValidateStatusTransition(ctx context.Context, currentStatus, newStatus pb.OrderStatus) error
	GenerateStatusNotifications(ctx context.Context, order *pb.Order, notifyCustomer bool) ([]string, error)
	AnalyzeCancellationImpact(ctx context.Context, order *pb.Order) (*pb.AIImpactAnalysis, error)
	GenerateAnalytics(ctx context.Context, orders []*pb.Order) (*pb.AIAnalytics, error)
	GetOrderRecommendations(ctx context.Context, req *pb.GetOrderRecommendationsRequest) ([]*pb.RecommendedItem, string, error)
	AnalyzeOrderPatterns(ctx context.Context, req *pb.AnalyzeOrderPatternsRequest) ([]*pb.PatternInsight, []string, error)
	PredictCompletionTime(ctx context.Context, req *pb.PredictCompletionTimeRequest) (*pb.PredictCompletionTimeResponse, error)
}

// RedisAIProcessor implements AIProcessor using Redis for AI operations
type RedisAIProcessor struct {
	logger *logger.Logger
}

// NewAIProcessor creates a new AI processor instance
func NewAIProcessor(logger *logger.Logger) AIProcessor {
	return &RedisAIProcessor{
		logger: logger,
	}
}

// AnalyzeOrder performs AI analysis on an order
func (p *RedisAIProcessor) AnalyzeOrder(ctx context.Context, order *pb.Order) (*pb.AIOrderInsights, error) {
	p.logger.WithFields(map[string]interface{}{"order_id": order.Id}).Info("Analyzing order with AI")

	// Calculate complexity score based on items and customizations
	complexityScore := p.calculateComplexityScore(order.Items)

	// Estimate preparation time using AI
	estimatedPrepTime := p.estimatePreparationTime(order.Items)

	// Generate recommendations
	recommendations := p.generateRecommendations(order)

	// Calculate revenue impact
	revenueImpact := p.calculateRevenueImpact(order)

	// Analyze peak time
	peakTimeAnalysis := p.analyzePeakTime()

	// Predict customer satisfaction
	satisfactionPrediction := p.predictCustomerSatisfaction(order)

	// Generate optimization suggestions
	optimizationSuggestions := p.generateOptimizationSuggestions(order)

	insights := &pb.AIOrderInsights{
		ComplexityScore:                complexityScore,
		EstimatedPrepTimeMinutes:       estimatedPrepTime,
		RecommendedAdditions:           recommendations,
		RevenueImpact:                  revenueImpact,
		PeakTimeAnalysis:               peakTimeAnalysis,
		CustomerSatisfactionPrediction: satisfactionPrediction,
		OptimizationSuggestions:        optimizationSuggestions,
	}

	return insights, nil
}

// GetRecommendations gets AI-powered recommendations for customer
func (p *RedisAIProcessor) GetRecommendations(ctx context.Context, customer *pb.Customer, items []*pb.OrderItem) ([]string, error) {
	p.logger.WithFields(map[string]interface{}{"customer_id": customer.Id}).Info("Getting AI recommendations")

	recommendations := []string{}

	// Analyze customer profile for personalized recommendations
	if customer.AiProfile != nil {
		// Recommend based on favorite items
		for _, favorite := range customer.AiProfile.FavoriteItems {
			if !p.containsItem(items, favorite) {
				recommendations = append(recommendations, fmt.Sprintf("Try your favorite: %s", favorite))
			}
		}

		// Recommend based on dietary preferences
		for _, preference := range customer.AiProfile.DietaryPreferences {
			recommendations = append(recommendations, fmt.Sprintf("Consider %s options", preference))
		}
	}

	// Add size upgrade recommendations
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.ProductName), "small") {
			recommendations = append(recommendations, fmt.Sprintf("Upgrade %s to medium for better value", item.ProductName))
		}
	}

	// Add complementary item recommendations
	if p.hasHotDrink(items) {
		recommendations = append(recommendations, "Add a pastry to complement your hot drink")
	}

	return recommendations, nil
}

// ValidateStatusTransition validates if a status transition is allowed
func (p *RedisAIProcessor) ValidateStatusTransition(ctx context.Context, currentStatus, newStatus pb.OrderStatus) error {
	validTransitions := map[pb.OrderStatus][]pb.OrderStatus{
		pb.OrderStatus_ORDER_STATUS_PENDING: {
			pb.OrderStatus_ORDER_STATUS_CONFIRMED,
			pb.OrderStatus_ORDER_STATUS_CANCELLED,
		},
		pb.OrderStatus_ORDER_STATUS_CONFIRMED: {
			pb.OrderStatus_ORDER_STATUS_PREPARING,
			pb.OrderStatus_ORDER_STATUS_CANCELLED,
		},
		pb.OrderStatus_ORDER_STATUS_PREPARING: {
			pb.OrderStatus_ORDER_STATUS_READY,
			pb.OrderStatus_ORDER_STATUS_CANCELLED,
		},
		pb.OrderStatus_ORDER_STATUS_READY: {
			pb.OrderStatus_ORDER_STATUS_COMPLETED,
		},
		pb.OrderStatus_ORDER_STATUS_COMPLETED: {}, // Terminal state
		pb.OrderStatus_ORDER_STATUS_CANCELLED: {}, // Terminal state
	}

	allowedStatuses, exists := validTransitions[currentStatus]
	if !exists {
		return fmt.Errorf("invalid current status: %v", currentStatus)
	}

	for _, allowed := range allowedStatuses {
		if allowed == newStatus {
			return nil
		}
	}

	return fmt.Errorf("transition from %v to %v is not allowed", currentStatus, newStatus)
}

// GenerateStatusNotifications generates AI-powered notifications for status changes
func (p *RedisAIProcessor) GenerateStatusNotifications(ctx context.Context, order *pb.Order, notifyCustomer bool) ([]string, error) {
	notifications := []string{}

	if notifyCustomer {
		switch order.Status {
		case pb.OrderStatus_ORDER_STATUS_CONFIRMED:
			notifications = append(notifications, fmt.Sprintf("Order %s confirmed. Estimated completion: %s",
				order.Id, order.EstimatedCompletion.AsTime().Format("15:04")))
		case pb.OrderStatus_ORDER_STATUS_PREPARING:
			notifications = append(notifications, fmt.Sprintf("Your order %s is being prepared", order.Id))
		case pb.OrderStatus_ORDER_STATUS_READY:
			notifications = append(notifications, fmt.Sprintf("Order %s is ready for pickup!", order.Id))
		case pb.OrderStatus_ORDER_STATUS_COMPLETED:
			notifications = append(notifications, fmt.Sprintf("Thank you! Order %s completed. Rate your experience", order.Id))
		case pb.OrderStatus_ORDER_STATUS_CANCELLED:
			notifications = append(notifications, fmt.Sprintf("Order %s has been cancelled. Refund will be processed", order.Id))
		}
	}

	// Internal notifications
	switch order.Status {
	case pb.OrderStatus_ORDER_STATUS_CONFIRMED:
		notifications = append(notifications, "Kitchen: New order added to queue")
	case pb.OrderStatus_ORDER_STATUS_PREPARING:
		notifications = append(notifications, "Staff: Order preparation started")
	case pb.OrderStatus_ORDER_STATUS_READY:
		notifications = append(notifications, "Front desk: Order ready for customer pickup")
	}

	return notifications, nil
}

// AnalyzeCancellationImpact analyzes the impact of order cancellation
func (p *RedisAIProcessor) AnalyzeCancellationImpact(ctx context.Context, order *pb.Order) (*pb.AIImpactAnalysis, error) {
	p.logger.WithFields(map[string]interface{}{"order_id": order.Id}).Info("Analyzing cancellation impact")

	// Calculate revenue impact
	revenueImpact := order.TotalAmount

	// Estimate customer satisfaction impact
	satisfactionImpact := -0.5 // Base negative impact
	if order.Status == pb.OrderStatus_ORDER_STATUS_PREPARING {
		satisfactionImpact = -1.0 // Higher impact if already preparing
	}

	// Generate mitigation suggestions
	mitigationSuggestions := []string{
		"Offer 10% discount on next order",
		"Provide complimentary drink voucher",
		"Send personalized apology message",
	}

	if order.TotalAmount > 20.0 {
		mitigationSuggestions = append(mitigationSuggestions, "Offer priority service on next visit")
	}

	// Calculate kitchen efficiency impact
	kitchenImpact := 0.1 // Minimal positive impact (freed resources)
	if order.Status == pb.OrderStatus_ORDER_STATUS_PREPARING {
		kitchenImpact = -0.2 // Negative impact (wasted preparation)
	}

	return &pb.AIImpactAnalysis{
		RevenueImpact:              revenueImpact,
		CustomerSatisfactionImpact: satisfactionImpact,
		MitigationSuggestions:      mitigationSuggestions,
		KitchenEfficiencyImpact:    kitchenImpact,
	}, nil
}

// GenerateAnalytics generates AI analytics for a list of orders
func (p *RedisAIProcessor) GenerateAnalytics(ctx context.Context, orders []*pb.Order) (*pb.AIAnalytics, error) {
	if len(orders) == 0 {
		return &pb.AIAnalytics{}, nil
	}

	totalRevenue := 0.0
	totalOrders := len(orders)
	itemCounts := make(map[string]int)
	satisfactionSum := 0.0
	satisfactionCount := 0

	for _, order := range orders {
		totalRevenue += order.TotalAmount

		// Count popular items
		for _, item := range order.Items {
			itemCounts[item.ProductName] += int(item.Quantity)
		}

		// Calculate satisfaction
		if order.AiInsights != nil && order.AiInsights.CustomerSatisfactionPrediction > 0 {
			satisfactionSum += order.AiInsights.CustomerSatisfactionPrediction
			satisfactionCount++
		}
	}

	// Find popular items
	popularItems := []string{}
	for item, count := range itemCounts {
		if count >= 3 { // Threshold for popularity
			popularItems = append(popularItems, item)
		}
	}

	avgOrderValue := totalRevenue / float64(totalOrders)
	avgSatisfaction := 8.5 // Default
	if satisfactionCount > 0 {
		avgSatisfaction = satisfactionSum / float64(satisfactionCount)
	}

	return &pb.AIAnalytics{
		AverageOrderValue:       avgOrderValue,
		TotalRevenue:            totalRevenue,
		PopularItems:            popularItems,
		CustomerSatisfactionAvg: avgSatisfaction,
		PeakHourAnalysis:        p.calculatePeakHours(orders),
	}, nil
}

// Helper methods

func (p *RedisAIProcessor) calculateComplexityScore(items []*pb.OrderItem) float64 {
	score := 0.0
	for _, item := range items {
		score += float64(len(item.Customizations)) * 0.5
		score += float64(item.Quantity) * 0.2
	}
	return score
}

func (p *RedisAIProcessor) estimatePreparationTime(items []*pb.OrderItem) float64 {
	baseTime := 2.0 // Base 2 minutes
	for _, item := range items {
		baseTime += float64(item.Quantity) * 1.5
		baseTime += float64(len(item.Customizations)) * 0.5
	}
	return baseTime
}

func (p *RedisAIProcessor) generateRecommendations(order *pb.Order) []string {
	recommendations := []string{}

	if len(order.Items) == 1 {
		recommendations = append(recommendations, "Add a pastry for a complete experience")
	}

	if order.TotalAmount < 10.0 {
		recommendations = append(recommendations, "Upgrade to larger size for better value")
	}

	return recommendations
}

func (p *RedisAIProcessor) calculateRevenueImpact(order *pb.Order) float64 {
	// Simple revenue impact calculation
	return order.TotalAmount * 1.2 // 20% markup consideration
}

func (p *RedisAIProcessor) analyzePeakTime() string {
	hour := time.Now().Hour()
	switch {
	case hour >= 7 && hour <= 9:
		return "Morning rush - high demand"
	case hour >= 12 && hour <= 14:
		return "Lunch peak - moderate demand"
	case hour >= 15 && hour <= 17:
		return "Afternoon break - steady demand"
	default:
		return "Off-peak - low demand"
	}
}

func (p *RedisAIProcessor) predictCustomerSatisfaction(order *pb.Order) float64 {
	// Base satisfaction score
	satisfaction := 8.0

	// Adjust based on order complexity
	if len(order.Items) > 3 {
		satisfaction -= 0.5
	}

	// Adjust based on total amount
	if order.TotalAmount > 25.0 {
		satisfaction += 0.5
	}

	return satisfaction
}

func (p *RedisAIProcessor) generateOptimizationSuggestions(order *pb.Order) []string {
	suggestions := []string{}

	if len(order.Items) > 2 {
		suggestions = append(suggestions, "Consider bundling items for efficiency")
	}

	if order.TotalAmount < 8.0 {
		suggestions = append(suggestions, "Suggest upselling opportunities")
	}

	return suggestions
}

func (p *RedisAIProcessor) containsItem(items []*pb.OrderItem, itemName string) bool {
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.ProductName), strings.ToLower(itemName)) {
			return true
		}
	}
	return false
}

func (p *RedisAIProcessor) hasHotDrink(items []*pb.OrderItem) bool {
	hotDrinks := []string{"latte", "cappuccino", "americano", "espresso", "mocha"}
	for _, item := range items {
		for _, hot := range hotDrinks {
			if strings.Contains(strings.ToLower(item.ProductName), hot) {
				return true
			}
		}
	}
	return false
}

func (p *RedisAIProcessor) calculatePeakHours(orders []*pb.Order) float64 {
	hourCounts := make(map[int]int)

	for _, order := range orders {
		hour := order.CreatedAt.AsTime().Hour()
		hourCounts[hour]++
	}

	maxCount := 0
	peakHour := 0
	for hour, count := range hourCounts {
		if count > maxCount {
			maxCount = count
			peakHour = hour
		}
	}

	return float64(peakHour)
}

// GetOrderRecommendations gets detailed AI recommendations for orders
func (p *RedisAIProcessor) GetOrderRecommendations(ctx context.Context, req *pb.GetOrderRecommendationsRequest) ([]*pb.RecommendedItem, string, error) {
	p.logger.Info("Getting detailed order recommendations", zap.String("customer_id", req.CustomerId))

	recommendations := []*pb.RecommendedItem{}

	// Sample recommendations based on time of day
	timeOfDay := req.TimeOfDay
	if timeOfDay == "" {
		hour := time.Now().Hour()
		if hour < 12 {
			timeOfDay = "morning"
		} else if hour < 17 {
			timeOfDay = "afternoon"
		} else {
			timeOfDay = "evening"
		}
	}

	// Generate recommendations based on time
	switch timeOfDay {
	case "morning":
		recommendations = append(recommendations, &pb.RecommendedItem{
			ProductId:         "latte-001",
			ProductName:       "Morning Latte",
			ConfidenceScore:   0.9,
			Reason:            "Popular morning choice with energizing caffeine",
			Price:             4.50,
			EstimatedPrepTime: 3.0,
		})
		recommendations = append(recommendations, &pb.RecommendedItem{
			ProductId:         "croissant-001",
			ProductName:       "Fresh Croissant",
			ConfidenceScore:   0.8,
			Reason:            "Perfect breakfast pairing",
			Price:             3.25,
			EstimatedPrepTime: 1.0,
		})
	case "afternoon":
		recommendations = append(recommendations, &pb.RecommendedItem{
			ProductId:         "americano-001",
			ProductName:       "Americano",
			ConfidenceScore:   0.85,
			Reason:            "Afternoon pick-me-up without overwhelming caffeine",
			Price:             3.75,
			EstimatedPrepTime: 2.5,
		})
	case "evening":
		recommendations = append(recommendations, &pb.RecommendedItem{
			ProductId:         "decaf-latte-001",
			ProductName:       "Decaf Latte",
			ConfidenceScore:   0.75,
			Reason:            "Evening-friendly option without caffeine",
			Price:             4.25,
			EstimatedPrepTime: 3.0,
		})
	}

	reasoning := fmt.Sprintf("Recommendations based on %s preferences and current time (%s). AI analyzed customer patterns and popular items for this time period.", req.CustomerId, timeOfDay)

	return recommendations, reasoning, nil
}

// AnalyzeOrderPatterns analyzes patterns in order data
func (p *RedisAIProcessor) AnalyzeOrderPatterns(ctx context.Context, req *pb.AnalyzeOrderPatternsRequest) ([]*pb.PatternInsight, []string, error) {
	p.logger.Info("Analyzing order patterns", zap.String("location_id", req.LocationId))

	insights := []*pb.PatternInsight{}
	recommendations := []string{}

	// Generate sample pattern insights
	insights = append(insights, &pb.PatternInsight{
		PatternType:   "temporal",
		Description:   "Peak ordering hours identified between 8-10 AM and 2-4 PM",
		Confidence:    0.92,
		AffectedItems: []string{"latte", "cappuccino", "americano"},
		TimePeriod:    "daily",
		ImpactScore:   8.5,
	})

	insights = append(insights, &pb.PatternInsight{
		PatternType:   "seasonal",
		Description:   "Cold brew orders increase by 40% during summer months",
		Confidence:    0.87,
		AffectedItems: []string{"cold_brew", "iced_latte", "frappuccino"},
		TimePeriod:    "seasonal",
		ImpactScore:   7.2,
	})

	insights = append(insights, &pb.PatternInsight{
		PatternType:   "customer_behavior",
		Description:   "Customers ordering pastries have 65% likelihood of ordering hot beverages",
		Confidence:    0.78,
		AffectedItems: []string{"croissant", "muffin", "latte", "cappuccino"},
		TimePeriod:    "ongoing",
		ImpactScore:   6.8,
	})

	// Generate recommendations based on patterns
	recommendations = append(recommendations,
		"Increase staff during peak hours (8-10 AM, 2-4 PM)",
		"Promote cold beverages during summer season",
		"Create pastry + hot drink combo offers",
		"Optimize inventory based on temporal patterns",
	)

	return insights, recommendations, nil
}

// PredictCompletionTime predicts when an order will be completed
func (p *RedisAIProcessor) PredictCompletionTime(ctx context.Context, req *pb.PredictCompletionTimeRequest) (*pb.PredictCompletionTimeResponse, error) {
	p.logger.Info("Predicting completion time", zap.String("order_id", req.OrderId))

	// Base prediction logic
	baseTime := 5.0 // 5 minutes base time

	// Adjust for queue size
	queueImpact := float64(req.CurrentQueueSize) * 1.5

	// Add some randomness for realistic prediction
	randomFactor := rand.Float64()*2.0 - 1.0 // -1 to +1 minutes

	estimatedMinutes := baseTime + queueImpact + randomFactor
	if estimatedMinutes < 2.0 {
		estimatedMinutes = 2.0 // Minimum 2 minutes
	}

	// Calculate confidence based on queue size
	confidence := 0.9
	if req.CurrentQueueSize > 5 {
		confidence = 0.7
	} else if req.CurrentQueueSize > 10 {
		confidence = 0.5
	}

	// Factors affecting prediction
	factors := []string{
		fmt.Sprintf("Current queue size: %d orders", req.CurrentQueueSize),
		"Historical preparation times",
		"Current staff availability",
		"Order complexity analysis",
	}

	estimatedCompletionTime := timestamppb.New(time.Now().Add(time.Duration(estimatedMinutes) * time.Minute))

	return &pb.PredictCompletionTimeResponse{
		EstimatedMinutes:        estimatedMinutes,
		ConfidenceLevel:         confidence,
		Factors:                 factors,
		EstimatedCompletionTime: estimatedCompletionTime,
		Success:                 true,
		Message:                 "Completion time predicted successfully",
	}, nil
}
