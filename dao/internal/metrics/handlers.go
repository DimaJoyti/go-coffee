package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// HTTP Handlers for Metrics operations

// GetTVLMetricsHandler handles GET /api/v1/tvl
func (s *Service) GetTVLMetricsHandler(c *gin.Context) {
	// Parse query parameters
	protocol := c.Query("protocol")
	chain := c.Query("chain")
	period := c.Query("period")

	req := &GetTVLMetricsRequest{
		Protocol: protocol,
		Chain:    chain,
		Period:   period,
	}

	// Get TVL metrics
	response, err := s.GetTVLMetrics(c.Request.Context(), req)
	if err != nil {
		s.logger.Error("Failed to get TVL metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get TVL metrics"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RecordTVLHandler handles POST /api/v1/tvl/record
func (s *Service) RecordTVLHandler(c *gin.Context) {
	var req RecordTVLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if req.Protocol == "" || req.Chain == "" || req.Source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	// Record TVL
	response, err := s.RecordTVL(c.Request.Context(), &req)
	if err != nil {
		s.logger.Error("Failed to record TVL", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetTVLHistoryHandler handles GET /api/v1/tvl/history
func (s *Service) GetTVLHistoryHandler(c *gin.Context) {
	protocol := c.Query("protocol")
	chain := c.Query("chain")
	daysStr := c.Query("days")

	days := 30 // default
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	// Get historical TVL data
	history, err := s.defiLlamaClient.GetHistoricalTVL(c.Request.Context(), protocol, days)
	if err != nil {
		s.logger.Error("Failed to get TVL history", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get TVL history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"protocol": protocol,
		"chain":    chain,
		"days":     days,
		"history":  history,
	})
}

// GetTVLByProtocolHandler handles GET /api/v1/tvl/by-protocol
func (s *Service) GetTVLByProtocolHandler(c *gin.Context) {
	// Parse pagination
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 && limitInt <= 100 {
			limit = limitInt
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offsetInt, err := strconv.Atoi(offsetStr); err == nil && offsetInt >= 0 {
			offset = offsetInt
		}
	}

	// Get protocols list
	protocols, err := s.defiLlamaClient.GetProtocolList(c.Request.Context())
	if err != nil {
		s.logger.Error("Failed to get protocols", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get protocols"})
		return
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(protocols) {
		start = len(protocols)
	}
	if end > len(protocols) {
		end = len(protocols)
	}

	paginatedProtocols := protocols[start:end]

	c.JSON(http.StatusOK, gin.H{
		"protocols": paginatedProtocols,
		"total":     len(protocols),
		"limit":     limit,
		"offset":    offset,
	})
}

// GetTVLGrowthHandler handles GET /api/v1/tvl/growth
func (s *Service) GetTVLGrowthHandler(c *gin.Context) {
	protocol := c.Query("protocol")
	chain := c.Query("chain")

	// Mock growth data
	growth := gin.H{
		"protocol":   protocol,
		"chain":      chain,
		"growth_24h": "2.5%",
		"growth_7d":  "15.2%",
		"growth_30d": "45.8%",
		"trend":      "bullish",
		"timestamp":  time.Now(),
	}

	c.JSON(http.StatusOK, growth)
}

// GetMAUMetricsHandler handles GET /api/v1/mau
func (s *Service) GetMAUMetricsHandler(c *gin.Context) {
	// Parse query parameters
	feature := c.Query("feature")
	period := c.Query("period")

	req := &GetMAUMetricsRequest{
		Feature: feature,
		Period:  period,
	}

	// Get MAU metrics
	response, err := s.GetMAUMetrics(c.Request.Context(), req)
	if err != nil {
		s.logger.Error("Failed to get MAU metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get MAU metrics"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RecordMAUHandler handles POST /api/v1/mau/record
func (s *Service) RecordMAUHandler(c *gin.Context) {
	var req RecordMAURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if req.Feature == "" || req.Period == "" || req.Source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	// Record MAU
	response, err := s.RecordMAU(c.Request.Context(), &req)
	if err != nil {
		s.logger.Error("Failed to record MAU", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetMAUHistoryHandler handles GET /api/v1/mau/history
func (s *Service) GetMAUHistoryHandler(c *gin.Context) {
	feature := c.Query("feature")
	monthsStr := c.Query("months")

	months := 12 // default
	if monthsStr != "" {
		if m, err := strconv.Atoi(monthsStr); err == nil && m > 0 && m <= 24 {
			months = m
		}
	}

	// Mock MAU history
	var history []gin.H
	baseMAU := 20000
	for i := 0; i < months; i++ {
		growth := float64(i) * 0.05 // 5% monthly growth
		mau := int(float64(baseMAU) * (1 + growth))

		history = append(history, gin.H{
			"month": time.Now().AddDate(0, -months+i+1, 0).Format("2006-01"),
			"mau":   mau,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"feature": feature,
		"months":  months,
		"history": history,
	})
}

// GetMAUByFeatureHandler handles GET /api/v1/mau/by-feature
func (s *Service) GetMAUByFeatureHandler(c *gin.Context) {
	// Mock MAU by feature
	features := []gin.H{
		{"feature": "defi_trading", "mau": 15000, "growth": "8.5%"},
		{"feature": "nft_marketplace", "mau": 8000, "growth": "12.3%"},
		{"feature": "dao_governance", "mau": 5000, "growth": "6.7%"},
		{"feature": "analytics_dashboard", "mau": 3000, "growth": "15.2%"},
		{"feature": "developer_tools", "mau": 2000, "growth": "22.1%"},
	}

	c.JSON(http.StatusOK, gin.H{
		"features": features,
		"total":    len(features),
	})
}

// GetMAUGrowthHandler handles GET /api/v1/mau/growth
func (s *Service) GetMAUGrowthHandler(c *gin.Context) {
	feature := c.Query("feature")

	// Mock growth data
	growth := gin.H{
		"feature":    feature,
		"growth_30d": "8.5%",
		"growth_90d": "25.3%",
		"retention":  "75.2%",
		"trend":      "growing",
		"timestamp":  time.Now(),
	}

	c.JSON(http.StatusOK, growth)
}

// GetPerformanceDashboardHandler handles GET /api/v1/performance/dashboard
func (s *Service) GetPerformanceDashboardHandler(c *gin.Context) {
	// Mock performance dashboard data
	dashboard := &PerformanceDashboard{
		TVLMetrics: &TVLMetricsResponse{
			CurrentTVL: decimal.NewFromFloat(5000000),
			Growth24h:  decimal.NewFromFloat(2.5),
			Growth7d:   decimal.NewFromFloat(15.2),
			Growth30d:  decimal.NewFromFloat(45.8),
			Timestamp:  time.Now(),
		},
		MAUMetrics: &MAUMetricsResponse{
			CurrentMAU: 25000,
			Growth30d:  decimal.NewFromFloat(8.5),
			Growth90d:  decimal.NewFromFloat(25.3),
			Retention:  decimal.NewFromFloat(75.2),
			Timestamp:  time.Now(),
		},
		TopContributors: []ImpactLeaderboard{
			{
				EntityID:   "dev_001",
				EntityType: "developer",
				Name:       "Alice Developer",
				TVLImpact:  decimal.NewFromFloat(500000),
				MAUImpact:  2500,
				TotalScore: decimal.NewFromFloat(95.5),
				Rank:       1,
			},
		},
		RecentAlerts: []Alert{
			{
				ID:      1,
				Name:    "TVL Growth Alert",
				Type:    AlertTypeGrowthRate,
				Status:  AlertStatusActive,
				Message: "TVL growth exceeded 20% in 24h",
			},
		},
		TrendingProtocols: []ProtocolTrend{
			{
				Protocol:   "go-coffee-defi",
				Chain:      "ethereum",
				CurrentTVL: decimal.NewFromFloat(5000000),
				Growth24h:  decimal.NewFromFloat(2.5),
				Growth7d:   decimal.NewFromFloat(15.2),
				TrendScore: decimal.NewFromFloat(8.7),
			},
		},
		LastUpdated: time.Now(),
	}

	c.JSON(http.StatusOK, dashboard)
}

// GetAttributionAnalysisHandler handles GET /api/v1/performance/attribution
func (s *Service) GetAttributionAnalysisHandler(c *gin.Context) {
	entityID := c.Query("entity_id")
	period := c.Query("period")

	if entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing entity_id parameter"})
		return
	}

	// Get attribution analysis
	attribution, err := s.impactRepo.GetAttribution(c.Request.Context(), entityID, period)
	if err != nil {
		s.logger.Error("Failed to get attribution analysis", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get attribution analysis"})
		return
	}

	c.JSON(http.StatusOK, attribution)
}

// RecordImpactHandler handles POST /api/v1/performance/impact
func (s *Service) RecordImpactHandler(c *gin.Context) {
	var req RecordImpactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if req.EntityID == "" || req.EntityType == "" || req.Source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	// Record impact
	response, err := s.RecordImpact(c.Request.Context(), &req)
	if err != nil {
		s.logger.Error("Failed to record impact", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetImpactLeaderboardHandler handles GET /api/v1/performance/leaderboard
func (s *Service) GetImpactLeaderboardHandler(c *gin.Context) {
	entityType := c.Query("entity_type")
	period := c.Query("period")

	// Parse limit
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 && limitInt <= 50 {
			limit = limitInt
		}
	}

	// Get leaderboard
	leaderboard, err := s.impactRepo.GetLeaderboard(c.Request.Context(), entityType, period, limit)
	if err != nil {
		s.logger.Error("Failed to get impact leaderboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get leaderboard"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_type": entityType,
		"period":      period,
		"leaderboard": leaderboard,
		"limit":       limit,
	})
}

// GetAnalyticsOverviewHandler handles GET /api/v1/analytics/overview
func (s *Service) GetAnalyticsOverviewHandler(c *gin.Context) {
	// Mock analytics overview
	overview := gin.H{
		"total_tvl":        "$8,500,000",
		"total_mau":        33000,
		"active_protocols": 5,
		"growth_rate_24h":  "3.2%",
		"growth_rate_30d":  "28.5%",
		"top_performing":   "DeFi Trading",
		"alerts_active":    3,
		"data_sources":     8,
		"last_updated":     time.Now(),
	}

	c.JSON(http.StatusOK, overview)
}

// GetTrendsAnalysisHandler handles GET /api/v1/analytics/trends
func (s *Service) GetTrendsAnalysisHandler(c *gin.Context) {
	_ = c.Query("metric_type") // metric type parameter (unused in mock)
	period := c.Query("period")

	// Mock trends analysis
	trends := &TrendsAnalysis{
		MetricType: MetricTypeTVL,
		Period:     period,
		Trend:      "bullish",
		GrowthRate: decimal.NewFromFloat(15.2),
		Seasonality: map[string]decimal.Decimal{
			"Q1": decimal.NewFromFloat(12.5),
			"Q2": decimal.NewFromFloat(18.3),
			"Q3": decimal.NewFromFloat(15.7),
			"Q4": decimal.NewFromFloat(22.1),
		},
		Forecast: map[string]decimal.Decimal{
			"next_week":    decimal.NewFromFloat(5200000),
			"next_month":   decimal.NewFromFloat(5800000),
			"next_quarter": decimal.NewFromFloat(6500000),
		},
		Confidence:  decimal.NewFromFloat(0.85),
		GeneratedAt: time.Now(),
	}

	c.JSON(http.StatusOK, trends)
}

// GetForecastsHandler handles GET /api/v1/analytics/forecasts
func (s *Service) GetForecastsHandler(c *gin.Context) {
	metricType := c.Query("metric_type")
	horizon := c.Query("horizon")

	// Mock forecasts
	forecasts := gin.H{
		"metric_type": metricType,
		"horizon":     horizon,
		"forecasts": []gin.H{
			{"period": "next_week", "value": "5,200,000", "confidence": "85%"},
			{"period": "next_month", "value": "5,800,000", "confidence": "78%"},
			{"period": "next_quarter", "value": "6,500,000", "confidence": "65%"},
		},
		"generated_at": time.Now(),
	}

	c.JSON(http.StatusOK, forecasts)
}

// GetAlertsHandler handles GET /api/v1/analytics/alerts
func (s *Service) GetAlertsHandler(c *gin.Context) {
	// Parse pagination
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 && limitInt <= 100 {
			limit = limitInt
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offsetInt, err := strconv.Atoi(offsetStr); err == nil && offsetInt >= 0 {
			offset = offsetInt
		}
	}

	// Get alerts
	alerts, err := s.alertRepo.GetActive(c.Request.Context(), limit, offset)
	if err != nil {
		s.logger.Error("Failed to get alerts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"limit":  limit,
		"offset": offset,
	})
}

// CreateAlertHandler handles POST /api/v1/analytics/alerts
func (s *Service) CreateAlertHandler(c *gin.Context) {
	var req CreateAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if req.Name == "" || req.Condition == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	// Create alert
	alert := &Alert{
		Name:       req.Name,
		Type:       req.Type,
		MetricType: req.MetricType,
		Threshold:  req.Threshold,
		Condition:  req.Condition,
		Status:     AlertStatusActive,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Metadata:   req.Metadata,
	}

	if err := s.alertRepo.Create(c.Request.Context(), alert); err != nil {
		s.logger.Error("Failed to create alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create alert"})
		return
	}

	response := &CreateAlertResponse{
		AlertID:   alert.ID,
		Status:    "created",
		CreatedAt: alert.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetDailyReportHandler handles GET /api/v1/reports/daily
func (s *Service) GetDailyReportHandler(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	// Mock daily report
	report := gin.H{
		"date":         date,
		"tvl_start":    "$4,850,000",
		"tvl_end":      "$5,000,000",
		"tvl_change":   "+3.1%",
		"mau_active":   24500,
		"new_users":    450,
		"transactions": 1250,
		"top_protocols": []gin.H{
			{"name": "DeFi Trading", "tvl": "$2,500,000", "change": "+2.8%"},
			{"name": "Lending", "tvl": "$1,800,000", "change": "+4.2%"},
		},
		"alerts_triggered": 2,
		"generated_at":     time.Now(),
	}

	c.JSON(http.StatusOK, report)
}

// GetWeeklyReportHandler handles GET /api/v1/reports/weekly
func (s *Service) GetWeeklyReportHandler(c *gin.Context) {
	week := c.Query("week")
	if week == "" {
		week = time.Now().Format("2006-W02")
	}

	// Mock weekly report
	report := gin.H{
		"week":          week,
		"tvl_growth":    "+15.2%",
		"mau_growth":    "+8.5%",
		"avg_daily_txs": 1150,
		"new_protocols": 1,
		"top_performers": []gin.H{
			{"entity": "Alice Developer", "impact": "$500,000", "type": "developer"},
			{"entity": "DeFi Analytics", "impact": "$300,000", "type": "solution"},
		},
		"generated_at": time.Now(),
	}

	c.JSON(http.StatusOK, report)
}

// GetMonthlyReportHandler handles GET /api/v1/reports/monthly
func (s *Service) GetMonthlyReportHandler(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	// Mock monthly report
	report := gin.H{
		"month":              month,
		"tvl_growth":         "+45.8%",
		"mau_growth":         "+25.3%",
		"total_transactions": 35000,
		"revenue_generated":  "$125,000",
		"developer_rewards":  "$37,500",
		"community_rewards":  "$12,500",
		"platform_revenue":   "$75,000",
		"generated_at":       time.Now(),
	}

	c.JSON(http.StatusOK, report)
}

// GenerateCustomReportHandler handles POST /api/v1/reports/generate
func (s *Service) GenerateCustomReportHandler(c *gin.Context) {
	var req GenerateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if req.Name == "" || req.Type == "" || req.Period == "" || req.CreatedBy == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	// Generate report
	report := &Report{
		Name:        req.Name,
		Type:        req.Type,
		Period:      req.Period,
		Data:        map[string]interface{}{"status": "generated"},
		GeneratedAt: time.Now(),
		CreatedBy:   req.CreatedBy,
	}

	if err := s.reportRepo.Create(c.Request.Context(), report); err != nil {
		s.logger.Error("Failed to create report", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report"})
		return
	}

	response := &GenerateReportResponse{
		ReportID:    report.ID,
		Status:      "generated",
		GeneratedAt: report.GeneratedAt,
		DownloadURL: fmt.Sprintf("/api/v1/reports/%d/download", report.ID),
	}

	c.JSON(http.StatusCreated, response)
}

// HandleWebhookHandler handles POST /api/v1/integrations/webhook
func (s *Service) HandleWebhookHandler(c *gin.Context) {
	var payload WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		s.logger.Error("Invalid webhook payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload"})
		return
	}

	s.logger.Info("Received webhook",
		zap.String("source", payload.Source),
		zap.String("type", payload.Type))

	// Process webhook data
	// In real implementation, this would validate signature and process data

	c.JSON(http.StatusOK, gin.H{
		"status":    "received",
		"timestamp": time.Now(),
	})
}

// GetDataSourcesHandler handles GET /api/v1/integrations/sources
func (s *Service) GetDataSourcesHandler(c *gin.Context) {
	sources, err := s.dataSourceRepo.GetAll(c.Request.Context())
	if err != nil {
		s.logger.Error("Failed to get data sources", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get data sources"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data_sources": sources,
		"total":        len(sources),
	})
}

// AddDataSourceHandler handles POST /api/v1/integrations/sources
func (s *Service) AddDataSourceHandler(c *gin.Context) {
	var req AddDataSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate required fields
	if req.Name == "" || req.Type == "" || req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	// Create data source
	source := &DataSource{
		Name:      req.Name,
		Type:      req.Type,
		URL:       req.URL,
		APIKey:    req.APIKey,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Config:    req.Config,
	}

	if err := s.dataSourceRepo.Create(c.Request.Context(), source); err != nil {
		s.logger.Error("Failed to create data source", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create data source"})
		return
	}

	response := &AddDataSourceResponse{
		DataSourceID: source.ID,
		Status:       "created",
		CreatedAt:    source.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}
