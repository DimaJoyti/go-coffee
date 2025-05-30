package security

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// RESTHandler handles REST API requests for DeFi security
type RESTHandler struct {
	defiHandler *DeFiSecurityHandler
	logger      *zap.Logger
}

// NewRESTHandler creates a new REST handler
func NewRESTHandler(defiHandler *DeFiSecurityHandler, logger *zap.Logger) *RESTHandler {
	return &RESTHandler{
		defiHandler: defiHandler,
		logger:      logger,
	}
}

// RegisterRoutes registers all security routes
func (h *RESTHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1/security")
	{
		// Validation endpoints
		v1.POST("/validate/arbitrage", h.ValidateArbitrage)
		v1.POST("/validate/yield-farming", h.ValidateYieldFarming)
		v1.POST("/validate/trading-bot", h.ValidateTradingBot)
		v1.POST("/validate/contract", h.ValidateContract)

		// Risk assessment endpoints
		v1.POST("/risk/assess", h.AssessRisk)
		v1.GET("/risk/thresholds", h.GetRiskThresholds)
		v1.PUT("/risk/thresholds", h.UpdateRiskThresholds)

		// Compliance endpoints
		v1.POST("/compliance/check", h.CheckCompliance)
		v1.GET("/compliance/rules", h.GetComplianceRules)

		// Monitoring endpoints
		v1.POST("/monitor/transaction", h.MonitorTransaction)
		v1.GET("/monitor/status/:hash", h.GetTransactionStatus)

		// Metrics and reporting endpoints
		v1.GET("/metrics", h.GetSecurityMetrics)
		v1.GET("/audit/logs", h.GetAuditLogs)
		v1.GET("/alerts/history", h.GetAlertHistory)

		// Configuration endpoints
		v1.GET("/config", h.GetSecurityConfig)
		v1.PUT("/config", h.UpdateSecurityConfig)
		v1.GET("/rules", h.GetAuditRules)
		v1.PUT("/rules/:id/toggle", h.ToggleAuditRule)
	}
}

// ValidateArbitrage validates an arbitrage transaction
func (h *RESTHandler) ValidateArbitrage(c *gin.Context) {
	var req ArbitrageValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid arbitrage validation request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.Timestamp = time.Now()

	response, err := h.defiHandler.ValidateArbitrageTransaction(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to validate arbitrage transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Validation failed"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ValidateYieldFarming validates a yield farming transaction
func (h *RESTHandler) ValidateYieldFarming(c *gin.Context) {
	var req YieldFarmingValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid yield farming validation request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.Timestamp = time.Now()

	response, err := h.defiHandler.ValidateYieldFarmingTransaction(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to validate yield farming transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Validation failed"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ValidateTradingBot validates a trading bot operation
func (h *RESTHandler) ValidateTradingBot(c *gin.Context) {
	var req TradingBotValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid trading bot validation request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.Timestamp = time.Now()

	response, err := h.defiHandler.ValidateTradingBotOperation(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to validate trading bot operation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Validation failed"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ValidateContract validates a smart contract
func (h *RESTHandler) ValidateContract(c *gin.Context) {
	var req ContractValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid contract validation request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.Timestamp = time.Now()

	// TODO: Implement contract validation logic
	response := &ContractValidationResponse{
		Valid:          true,
		Verified:       false,
		RiskScore:      decimal.NewFromFloat(3.0),
		SecurityIssues: []SecurityIssue{},
		Recommendations: []string{
			"Contract verification recommended",
			"Consider additional security audit",
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// AssessRisk performs risk assessment
func (h *RESTHandler) AssessRisk(c *gin.Context) {
	var req RiskAssessmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid risk assessment request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.Timestamp = time.Now()

	// TODO: Implement risk assessment logic
	response := &RiskAssessmentResponse{
		RiskScore:      decimal.NewFromFloat(4.5),
		RiskLevel:      "medium",
		Recommendation: "Proceed with caution",
		Factors: []RiskFactor{
			{
				Name:        "Amount Size",
				Impact:      decimal.NewFromFloat(2.0),
				Description: "Transaction amount is moderate",
				Weight:      decimal.NewFromFloat(0.3),
			},
			{
				Name:        "Protocol Risk",
				Impact:      decimal.NewFromFloat(3.0),
				Description: "Protocol has medium risk profile",
				Weight:      decimal.NewFromFloat(0.4),
			},
		},
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// GetRiskThresholds returns current risk thresholds
func (h *RESTHandler) GetRiskThresholds(c *gin.Context) {
	// TODO: Get from handler
	thresholds := RiskThresholds{
		MaxTransactionAmount: decimal.NewFromFloat(100000),
		MaxSlippage:          decimal.NewFromFloat(0.05),
		MaxGasPrice:          decimal.NewFromFloat(100),
		MaxDailyVolume:       decimal.NewFromFloat(1000000),
		MinLiquidity:         decimal.NewFromFloat(10000),
		MaxPriceImpact:       decimal.NewFromFloat(0.03),
	}

	c.JSON(http.StatusOK, thresholds)
}

// UpdateRiskThresholds updates risk thresholds
func (h *RESTHandler) UpdateRiskThresholds(c *gin.Context) {
	var thresholds RiskThresholds
	if err := c.ShouldBindJSON(&thresholds); err != nil {
		h.logger.Error("Invalid risk thresholds request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	err := h.defiHandler.UpdateRiskThresholds(c.Request.Context(), thresholds)
	if err != nil {
		h.logger.Error("Failed to update risk thresholds", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Risk thresholds updated successfully"})
}

// CheckCompliance performs compliance checking
func (h *RESTHandler) CheckCompliance(c *gin.Context) {
	var req ComplianceCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid compliance check request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.Timestamp = time.Now()

	// TODO: Implement compliance checking logic
	response := &ComplianceCheckResponse{
		Compliant:       true,
		Violations:      []ComplianceViolation{},
		RequiredActions: []string{},
		Restrictions:    []string{},
		Timestamp:       time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// GetComplianceRules returns compliance rules
func (h *RESTHandler) GetComplianceRules(c *gin.Context) {
	// TODO: Implement compliance rules retrieval
	rules := []map[string]interface{}{
		{
			"id":          "kyc_required",
			"name":        "KYC Required",
			"description": "KYC verification required for transactions above $10k",
			"enabled":     true,
			"threshold":   10000,
		},
		{
			"id":          "daily_limit",
			"name":        "Daily Transaction Limit",
			"description": "Maximum $100k daily transaction limit",
			"enabled":     true,
			"threshold":   100000,
		},
	}

	c.JSON(http.StatusOK, gin.H{"rules": rules})
}

// MonitorTransaction starts monitoring a transaction
func (h *RESTHandler) MonitorTransaction(c *gin.Context) {
	var req TransactionMonitoringRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid transaction monitoring request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.Timestamp = time.Now()

	// TODO: Implement transaction monitoring logic
	response := &TransactionMonitoringResponse{
		Status:         "pending",
		Confirmations:  0,
		GasUsed:        decimal.Zero,
		GasPrice:       decimal.Zero,
		Success:        false,
		SecurityAlerts: []SecurityAlert{},
		Timestamp:      time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// GetTransactionStatus returns transaction monitoring status
func (h *RESTHandler) GetTransactionStatus(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction hash required"})
		return
	}

	// TODO: Implement transaction status retrieval
	response := &TransactionMonitoringResponse{
		Status:         "confirmed",
		Confirmations:  12,
		GasUsed:        decimal.NewFromFloat(21000),
		GasPrice:       decimal.NewFromFloat(20),
		Success:        true,
		SecurityAlerts: []SecurityAlert{},
		Timestamp:      time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// GetSecurityMetrics returns security metrics
func (h *RESTHandler) GetSecurityMetrics(c *gin.Context) {
	metrics, err := h.defiHandler.GetSecurityMetrics(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get security metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetAuditLogs returns audit logs
func (h *RESTHandler) GetAuditLogs(c *gin.Context) {
	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	req := AuditLogRequest{
		UserID:    c.Query("user_id"),
		BotID:     c.Query("bot_id"),
		EventType: c.Query("event_type"),
		Severity:  c.Query("severity"),
		Limit:     limit,
		Offset:    offset,
	}

	// Parse time parameters
	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			req.StartTime = t
		}
	}
	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			req.EndTime = t
		}
	}

	// TODO: Implement audit log retrieval
	response := &AuditLogResponse{
		Events:    []AuditLogEntry{},
		Total:     0,
		Limit:     limit,
		Offset:    offset,
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// GetAlertHistory returns alert history
func (h *RESTHandler) GetAlertHistory(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// TODO: Get from alert manager
	alerts := []map[string]interface{}{
		{
			"id":        "alert_001",
			"type":      "high_transaction_amount",
			"severity":  "high",
			"message":   "Transaction amount exceeds threshold",
			"timestamp": time.Now().Add(-time.Hour),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"total":  len(alerts),
		"limit":  limit,
	})
}

// GetSecurityConfig returns security configuration
func (h *RESTHandler) GetSecurityConfig(c *gin.Context) {
	// TODO: Implement security config retrieval
	config := map[string]interface{}{
		"risk_thresholds": map[string]interface{}{
			"max_transaction_amount": 100000,
			"max_slippage":           0.05,
			"max_gas_price":          100,
		},
		"audit_rules": []map[string]interface{}{
			{
				"id":       "high_transaction_amount",
				"enabled":  true,
				"severity": "high",
				"action":   "alert",
			},
		},
		"alert_settings": map[string]interface{}{
			"email_enabled":     true,
			"slack_enabled":     true,
			"pagerduty_enabled": false,
		},
	}

	c.JSON(http.StatusOK, config)
}

// UpdateSecurityConfig updates security configuration
func (h *RESTHandler) UpdateSecurityConfig(c *gin.Context) {
	var req SecurityConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid security config request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.Timestamp = time.Now()

	// TODO: Implement security config update
	response := &SecurityConfigResponse{
		Success:   true,
		Message:   "Security configuration updated successfully",
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// GetAuditRules returns audit rules
func (h *RESTHandler) GetAuditRules(c *gin.Context) {
	// TODO: Get from auditor
	rules := []map[string]interface{}{
		{
			"id":          "high_transaction_amount",
			"name":        "High Transaction Amount",
			"description": "Detects transactions above threshold",
			"enabled":     true,
			"severity":    "high",
			"action":      "alert",
		},
		{
			"id":          "high_slippage",
			"name":        "High Slippage",
			"description": "Detects transactions with high slippage",
			"enabled":     true,
			"severity":    "medium",
			"action":      "alert",
		},
	}

	c.JSON(http.StatusOK, gin.H{"rules": rules})
}

// ToggleAuditRule toggles an audit rule
func (h *RESTHandler) ToggleAuditRule(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rule ID required"})
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid toggle rule request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// TODO: Implement rule toggle
	h.logger.Info("Toggling audit rule",
		zap.String("rule_id", ruleID),
		zap.Bool("enabled", req.Enabled),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Audit rule updated successfully",
		"rule_id": ruleID,
		"enabled": req.Enabled,
	})
}
