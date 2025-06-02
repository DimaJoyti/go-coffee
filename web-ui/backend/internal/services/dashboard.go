package services

import (
	"time"
)

type DashboardMetrics struct {
	TotalOrders     int     `json:"totalOrders"`
	TotalRevenue    float64 `json:"totalRevenue"`
	PortfolioValue  float64 `json:"portfolioValue"`
	ActiveAgents    int     `json:"activeAgents"`
	OrdersChange    float64 `json:"ordersChange"`
	RevenueChange   float64 `json:"revenueChange"`
	PortfolioChange float64 `json:"portfolioChange"`
	AgentsChange    float64 `json:"agentsChange"`
}

type Activity struct {
	ID      int       `json:"id"`
	Type    string    `json:"type"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
	Icon    string    `json:"icon"`
	Color   string    `json:"color"`
}

type DashboardService struct {
	// In a real implementation, this would have database connections, etc.
}

func NewDashboardService() *DashboardService {
	return &DashboardService{}
}

func (s *DashboardService) GetMetrics() (*DashboardMetrics, error) {
	// In a real implementation, this would fetch data from various sources
	return &DashboardMetrics{
		TotalOrders:     1247,
		TotalRevenue:    45678.90,
		PortfolioValue:  123456.78,
		ActiveAgents:    9,
		OrdersChange:    12.5,
		RevenueChange:   8.3,
		PortfolioChange: 15.7,
		AgentsChange:    0,
	}, nil
}

func (s *DashboardService) GetActivity() ([]Activity, error) {
	// Mock data - in real implementation, this would come from various services
	activities := []Activity{
		{
			ID:      1,
			Type:    "order",
			Message: "New coffee order #1247",
			Time:    time.Now().Add(-2 * time.Minute),
			Icon:    "Coffee",
			Color:   "text-coffee-500",
		},
		{
			ID:      2,
			Type:    "trade",
			Message: "DeFi arbitrage executed",
			Time:    time.Now().Add(-5 * time.Minute),
			Icon:    "TrendingUp",
			Color:   "text-green-500",
		},
		{
			ID:      3,
			Type:    "agent",
			Message: "Inventory agent updated stock",
			Time:    time.Now().Add(-10 * time.Minute),
			Icon:    "Bot",
			Color:   "text-blue-500",
		},
		{
			ID:      4,
			Type:    "user",
			Message: "New customer registered",
			Time:    time.Now().Add(-15 * time.Minute),
			Icon:    "Users",
			Color:   "text-purple-500",
		},
		{
			ID:      5,
			Type:    "system",
			Message: "Market data updated",
			Time:    time.Now().Add(-20 * time.Minute),
			Icon:    "Globe",
			Color:   "text-orange-500",
		},
	}

	return activities, nil
}
