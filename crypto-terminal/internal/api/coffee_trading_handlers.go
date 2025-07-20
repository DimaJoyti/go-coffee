package api

import (
	"net/http"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/trading"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// CoffeeTradingHandler handles coffee-themed trading API endpoints
type CoffeeTradingHandler struct {
	strategyEngine  *trading.StrategyEngine
	strategyFactory *trading.CoffeeStrategyFactory
	logger          *logrus.Logger
}

// NewCoffeeTradingHandler creates a new coffee trading handler
func NewCoffeeTradingHandler(strategyEngine *trading.StrategyEngine, logger *logrus.Logger) *CoffeeTradingHandler {
	return &CoffeeTradingHandler{
		strategyEngine:  strategyEngine,
		strategyFactory: trading.NewCoffeeStrategyFactory(),
		logger:          logger,
	}
}

// RegisterRoutes registers all coffee trading routes
func (cth *CoffeeTradingHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1/coffee-trading")
	{
		// Strategy Management
		strategies := api.Group("/strategies")
		{
			strategies.GET("", cth.GetAllStrategies)
			strategies.POST("", cth.CreateStrategy)
			strategies.GET("/:id", cth.GetStrategy)
			strategies.PUT("/:id", cth.UpdateStrategy)
			strategies.DELETE("/:id", cth.DeleteStrategy)
			strategies.POST("/:id/start", cth.StartStrategy)
			strategies.POST("/:id/stop", cth.StopStrategy)
			strategies.GET("/:id/performance", cth.GetStrategyPerformance)
		}

		// Coffee-Themed Strategy Endpoints
		coffee := api.Group("/coffee")
		{
			coffee.POST("/espresso/start", cth.StartEspressoStrategy)
			coffee.POST("/latte/start", cth.StartLatteStrategy)
			coffee.POST("/cold-brew/start", cth.StartColdBrewStrategy)
			coffee.POST("/cappuccino/start", cth.StartCappuccinoStrategy)
			coffee.GET("/menu", cth.GetCoffeeMenu)
			coffee.GET("/recommendations", cth.GetStrategyRecommendations)
		}

		// Portfolio Management
		portfolio := api.Group("/portfolio")
		{
			portfolio.GET("", cth.GetPortfolio)
			portfolio.GET("/performance", cth.GetPortfolioPerformance)
			portfolio.GET("/risk-metrics", cth.GetPortfolioRiskMetrics)
			portfolio.GET("/positions", cth.GetPositions)
			portfolio.POST("/positions/:symbol/close", cth.ClosePosition)
		}

		// Signal Management
		signals := api.Group("/signals")
		{
			signals.GET("", cth.GetRecentSignals)
			signals.POST("", cth.CreateManualSignal)
			signals.GET("/:id", cth.GetSignal)
			signals.POST("/:id/execute", cth.ExecuteSignal)
		}

		// Analytics
		analytics := api.Group("/analytics")
		{
			analytics.GET("/dashboard", cth.GetDashboard)
			analytics.GET("/performance", cth.GetPerformanceAnalytics)
			analytics.GET("/risk", cth.GetRiskAnalytics)
			analytics.GET("/coffee-correlation", cth.GetCoffeeCorrelation)
		}
	}
}

// Strategy Management Handlers

// GetAllStrategies gets all trading strategies
func (cth *CoffeeTradingHandler) GetAllStrategies(c *gin.Context) {
	strategies := cth.strategyEngine.GetAllStrategies()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"strategies": strategies,
			"count":      len(strategies),
		},
	})
}

// CreateStrategy creates a new trading strategy
func (cth *CoffeeTradingHandler) CreateStrategy(c *gin.Context) {
	var req struct {
		Name        string                     `json:"name" binding:"required"`
		Type        trading.CoffeeStrategyType `json:"type" binding:"required"`
		Symbol      string                     `json:"symbol" binding:"required"`
		Config      *trading.StrategyConfig    `json:"config,omitempty"`
		Description string                     `json:"description,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	var strategy *trading.CoffeeStrategy

	// Create strategy based on type
	switch req.Type {
	case trading.StrategyEspresso:
		strategy = cth.strategyFactory.CreateEspressoStrategy(req.Symbol)
	case trading.StrategyLatte:
		strategy = cth.strategyFactory.CreateLatteStrategy(req.Symbol)
	case trading.StrategyColdBrew:
		strategy = cth.strategyFactory.CreateColdBrewStrategy(req.Symbol)
	case trading.StrategyCappuccino:
		strategy = cth.strategyFactory.CreateCappuccinoStrategy(req.Symbol)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid strategy type",
		})
		return
	}

	// Override with custom config if provided
	if req.Config != nil {
		strategy.Config = *req.Config
	}

	if req.Name != "" {
		strategy.Name = req.Name
	}

	if req.Description != "" {
		strategy.Description = req.Description
	}

	if err := cth.strategyEngine.AddStrategy(strategy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"strategy": strategy,
			"message":  "Strategy created successfully",
		},
	})
}

// GetStrategy gets a specific strategy
func (cth *CoffeeTradingHandler) GetStrategy(c *gin.Context) {
	strategyID := c.Param("id")

	strategy, err := cth.strategyEngine.GetStrategy(strategyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"strategy": strategy,
		},
	})
}

// StartStrategy starts a specific strategy
func (cth *CoffeeTradingHandler) StartStrategy(c *gin.Context) {
	strategyID := c.Param("id")

	if err := cth.strategyEngine.StartStrategy(strategyID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Strategy started successfully",
	})
}

// StopStrategy stops a specific strategy
func (cth *CoffeeTradingHandler) StopStrategy(c *gin.Context) {
	strategyID := c.Param("id")

	if err := cth.strategyEngine.StopStrategy(strategyID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Strategy stopped successfully",
	})
}

// Coffee-Themed Strategy Handlers

// StartEspressoStrategy starts an espresso scalping strategy
func (cth *CoffeeTradingHandler) StartEspressoStrategy(c *gin.Context) {
	var req struct {
		Symbol string `json:"symbol" binding:"required"`
		Name   string `json:"name,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	strategy := cth.strategyFactory.CreateEspressoStrategy(req.Symbol)
	if req.Name != "" {
		strategy.Name = req.Name
	}

	if err := cth.strategyEngine.AddStrategy(strategy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := cth.strategyEngine.StartStrategy(strategy.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"strategy":    strategy,
			"message":     "☕ Espresso strategy brewing! High-frequency scalping activated.",
			"description": trading.GetCoffeeThemeDescription(trading.StrategyEspresso),
		},
	})
}

// StartLatteStrategy starts a latte swing trading strategy
func (cth *CoffeeTradingHandler) StartLatteStrategy(c *gin.Context) {
	var req struct {
		Symbol string `json:"symbol" binding:"required"`
		Name   string `json:"name,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	strategy := cth.strategyFactory.CreateLatteStrategy(req.Symbol)
	if req.Name != "" {
		strategy.Name = req.Name
	}

	if err := cth.strategyEngine.AddStrategy(strategy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := cth.strategyEngine.StartStrategy(strategy.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"strategy":    strategy,
			"message":     "🥛 Latte strategy steaming! Smooth swing trading in progress.",
			"description": trading.GetCoffeeThemeDescription(trading.StrategyLatte),
		},
	})
}

// GetCoffeeMenu gets the coffee strategy menu
func (cth *CoffeeTradingHandler) GetCoffeeMenu(c *gin.Context) {
	menu := map[string]interface{}{
		"espresso": gin.H{
			"name":        "Espresso Scalper",
			"description": trading.GetCoffeeThemeDescription(trading.StrategyEspresso),
			"timeframes":  trading.GetOptimalTimeframes(trading.StrategyEspresso),
			"indicators":  trading.GetRecommendedIndicators(trading.StrategyEspresso),
			"risk_level":  "High",
			"emoji":       "☕",
		},
		"latte": gin.H{
			"name":        "Latte Swing",
			"description": trading.GetCoffeeThemeDescription(trading.StrategyLatte),
			"timeframes":  trading.GetOptimalTimeframes(trading.StrategyLatte),
			"indicators":  trading.GetRecommendedIndicators(trading.StrategyLatte),
			"risk_level":  "Medium",
			"emoji":       "🥛",
		},
		"cold_brew": gin.H{
			"name":        "Cold Brew Position",
			"description": trading.GetCoffeeThemeDescription(trading.StrategyColdBrew),
			"timeframes":  trading.GetOptimalTimeframes(trading.StrategyColdBrew),
			"indicators":  trading.GetRecommendedIndicators(trading.StrategyColdBrew),
			"risk_level":  "Low",
			"emoji":       "🧊",
		},
		"cappuccino": gin.H{
			"name":        "Cappuccino Momentum",
			"description": trading.GetCoffeeThemeDescription(trading.StrategyCappuccino),
			"timeframes":  trading.GetOptimalTimeframes(trading.StrategyCappuccino),
			"indicators":  trading.GetRecommendedIndicators(trading.StrategyCappuccino),
			"risk_level":  "Medium-High",
			"emoji":       "🫖",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"menu":        menu,
			"message":     "☕ Welcome to the Coffee Trading Menu! Choose your perfect strategy blend.",
			"total_items": len(menu),
		},
	})
}

// StartColdBrewStrategy starts a cold brew position trading strategy
func (cth *CoffeeTradingHandler) StartColdBrewStrategy(c *gin.Context) {
	var req struct {
		Symbol string `json:"symbol" binding:"required"`
		Name   string `json:"name,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	strategy := cth.strategyFactory.CreateColdBrewStrategy(req.Symbol)
	if req.Name != "" {
		strategy.Name = req.Name
	}

	if err := cth.strategyEngine.AddStrategy(strategy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := cth.strategyEngine.StartStrategy(strategy.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"strategy":    strategy,
			"message":     "🧊 Cold Brew strategy chilling! Patient position trading activated.",
			"description": trading.GetCoffeeThemeDescription(trading.StrategyColdBrew),
		},
	})
}

// StartCappuccinoStrategy starts a cappuccino momentum trading strategy
func (cth *CoffeeTradingHandler) StartCappuccinoStrategy(c *gin.Context) {
	var req struct {
		Symbol string `json:"symbol" binding:"required"`
		Name   string `json:"name,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	strategy := cth.strategyFactory.CreateCappuccinoStrategy(req.Symbol)
	if req.Name != "" {
		strategy.Name = req.Name
	}

	if err := cth.strategyEngine.AddStrategy(strategy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := cth.strategyEngine.StartStrategy(strategy.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"strategy":    strategy,
			"message":     "🫖 Cappuccino strategy frothing! Dynamic momentum trading engaged.",
			"description": trading.GetCoffeeThemeDescription(trading.StrategyCappuccino),
		},
	})
}

// GetStrategyRecommendations gets strategy recommendations based on market conditions
func (cth *CoffeeTradingHandler) GetStrategyRecommendations(c *gin.Context) {
	marketCondition := c.Query("market_condition")
	volatilityStr := c.Query("volatility")

	var volatility decimal.Decimal
	if volatilityStr != "" {
		var err error
		volatility, err = decimal.NewFromString(volatilityStr)
		if err != nil {
			volatility = decimal.NewFromFloat(0.03) // Default 3%
		}
	} else {
		volatility = decimal.NewFromFloat(0.03)
	}

	if marketCondition == "" {
		marketCondition = "sideways"
	}

	recommendedType := cth.strategyFactory.GetStrategyRecommendation(marketCondition, volatility)

	recommendations := gin.H{
		"recommended_strategy": recommendedType,
		"market_condition":     marketCondition,
		"volatility":           volatility,
		"description":          trading.GetCoffeeThemeDescription(recommendedType),
		"timeframes":           trading.GetOptimalTimeframes(recommendedType),
		"indicators":           trading.GetRecommendedIndicators(recommendedType),
		"alternatives": []gin.H{
			{
				"type":        trading.StrategyEspresso,
				"description": trading.GetCoffeeThemeDescription(trading.StrategyEspresso),
				"suitability": "High volatility, trending markets",
			},
			{
				"type":        trading.StrategyLatte,
				"description": trading.GetCoffeeThemeDescription(trading.StrategyLatte),
				"suitability": "Moderate volatility, balanced approach",
			},
			{
				"type":        trading.StrategyColdBrew,
				"description": trading.GetCoffeeThemeDescription(trading.StrategyColdBrew),
				"suitability": "Low volatility, long-term positions",
			},
			{
				"type":        trading.StrategyCappuccino,
				"description": trading.GetCoffeeThemeDescription(trading.StrategyCappuccino),
				"suitability": "Momentum markets, trend following",
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    recommendations,
		"message": "☕ Here's your personalized coffee strategy recommendation!",
	})
}

// GetPortfolio gets the current portfolio
func (cth *CoffeeTradingHandler) GetPortfolio(c *gin.Context) {
	portfolio := cth.strategyEngine.GetPortfolio()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"portfolio": portfolio,
		},
	})
}

// GetDashboard gets the trading dashboard data
func (cth *CoffeeTradingHandler) GetDashboard(c *gin.Context) {
	strategies := cth.strategyEngine.GetAllStrategies()
	portfolio := cth.strategyEngine.GetPortfolio()

	activeStrategies := 0
	totalPnL := decimal.Zero

	for _, strategy := range strategies {
		if strategy.Status == trading.StatusActive {
			activeStrategies++
		}
		totalPnL = totalPnL.Add(strategy.Performance.TotalPnL)
	}

	dashboard := gin.H{
		"portfolio": gin.H{
			"total_value":       portfolio.TotalValue,
			"available_balance": portfolio.AvailableBalance,
			"total_pnl":         portfolio.TotalPnL,
			"daily_pnl":         portfolio.DailyPnL,
			"unrealized_pnl":    portfolio.UnrealizedPnL,
		},
		"strategies": gin.H{
			"total":  len(strategies),
			"active": activeStrategies,
			"paused": len(strategies) - activeStrategies,
		},
		"performance": gin.H{
			"total_pnl":    totalPnL,
			"win_rate":     calculateOverallWinRate(strategies),
			"total_trades": calculateTotalTrades(strategies),
		},
		"coffee_theme": gin.H{
			"message": "☕ Welcome to your Coffee Trading Dashboard!",
			"tip":     "Remember: Good coffee and good trades both require patience and timing.",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dashboard,
	})
}

// Helper functions
func calculateOverallWinRate(strategies []*trading.CoffeeStrategy) decimal.Decimal {
	totalTrades := 0
	totalWins := 0

	for _, strategy := range strategies {
		totalTrades += strategy.Performance.TotalTrades
		totalWins += strategy.Performance.WinningTrades
	}

	if totalTrades == 0 {
		return decimal.Zero
	}

	return decimal.NewFromInt(int64(totalWins)).Div(decimal.NewFromInt(int64(totalTrades))).Mul(decimal.NewFromInt(100))
}

func calculateTotalTrades(strategies []*trading.CoffeeStrategy) int {
	total := 0
	for _, strategy := range strategies {
		total += strategy.Performance.TotalTrades
	}
	return total
}

// Placeholder handlers for remaining endpoints
func (cth *CoffeeTradingHandler) UpdateStrategy(c *gin.Context)          { /* TODO */ }
func (cth *CoffeeTradingHandler) DeleteStrategy(c *gin.Context)          { /* TODO */ }
func (cth *CoffeeTradingHandler) GetStrategyPerformance(c *gin.Context)  { /* TODO */ }
func (cth *CoffeeTradingHandler) GetPortfolioPerformance(c *gin.Context) { /* TODO */ }
func (cth *CoffeeTradingHandler) GetPortfolioRiskMetrics(c *gin.Context) { /* TODO */ }
func (cth *CoffeeTradingHandler) GetPositions(c *gin.Context)            { /* TODO */ }
func (cth *CoffeeTradingHandler) ClosePosition(c *gin.Context)           { /* TODO */ }
func (cth *CoffeeTradingHandler) GetRecentSignals(c *gin.Context)        { /* TODO */ }
func (cth *CoffeeTradingHandler) CreateManualSignal(c *gin.Context)      { /* TODO */ }
func (cth *CoffeeTradingHandler) GetSignal(c *gin.Context)               { /* TODO */ }
func (cth *CoffeeTradingHandler) ExecuteSignal(c *gin.Context)           { /* TODO */ }
func (cth *CoffeeTradingHandler) GetPerformanceAnalytics(c *gin.Context) { /* TODO */ }
func (cth *CoffeeTradingHandler) GetRiskAnalytics(c *gin.Context)        { /* TODO */ }
func (cth *CoffeeTradingHandler) GetCoffeeCorrelation(c *gin.Context)    { /* TODO */ }
