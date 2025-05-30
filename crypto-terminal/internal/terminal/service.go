package terminal

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/alerts"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/market"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/models"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/orderflow"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/portfolio"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/websocket"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// Service represents the main terminal service
type Service struct {
	config           *config.Config
	db               *sql.DB
	redis            *redis.Client
	marketService    *market.Service
	portfolioService *portfolio.Service
	alertsService    *alerts.Service
	orderflowService *orderflow.Service
	wsHub            *websocket.Hub
	startTime        time.Time
}

// NewService creates a new terminal service
func NewService(cfg *config.Config) (*Service, error) {
	// Initialize database connection
	db, err := sql.Open("postgres", cfg.Database.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure database connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Test database connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize Redis connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.GetRedisAddr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		MaxRetries:   cfg.Redis.MaxRetries,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Initialize market service
	marketService, err := market.NewService(cfg, db, redisClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create market service: %w", err)
	}

	// Initialize portfolio service
	portfolioService, err := portfolio.NewService(cfg, db, redisClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create portfolio service: %w", err)
	}

	// Initialize alerts service
	alertsService, err := alerts.NewService(cfg, db, redisClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create alerts service: %w", err)
	}

	// Initialize order flow service
	orderflowService, err := orderflow.NewService(cfg, db, redisClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create order flow service: %w", err)
	}

	// Initialize WebSocket hub
	wsHub := websocket.NewHub(cfg)

	return &Service{
		config:           cfg,
		db:               db,
		redis:            redisClient,
		marketService:    marketService,
		portfolioService: portfolioService,
		alertsService:    alertsService,
		orderflowService: orderflowService,
		wsHub:            wsHub,
		startTime:        time.Now(),
	}, nil
}

// Start starts the terminal service
func (s *Service) Start(ctx context.Context) error {
	logrus.Info("Starting crypto terminal service")

	// Start WebSocket hub
	go s.wsHub.Run(ctx)

	// Start market data service
	if err := s.marketService.Start(ctx); err != nil {
		return fmt.Errorf("failed to start market service: %w", err)
	}

	// Start portfolio service
	if err := s.portfolioService.Start(ctx); err != nil {
		return fmt.Errorf("failed to start portfolio service: %w", err)
	}

	// Start alerts service
	if err := s.alertsService.Start(ctx); err != nil {
		return fmt.Errorf("failed to start alerts service: %w", err)
	}

	// Start order flow service
	if err := s.orderflowService.Start(ctx); err != nil {
		return fmt.Errorf("failed to start order flow service: %w", err)
	}

	logrus.Info("Crypto terminal service started successfully")
	return nil
}

// Stop stops the terminal service
func (s *Service) Stop() error {
	logrus.Info("Stopping crypto terminal service")

	// Stop services
	if err := s.orderflowService.Stop(); err != nil {
		logrus.Errorf("Failed to stop order flow service: %v", err)
	}

	if err := s.alertsService.Stop(); err != nil {
		logrus.Errorf("Failed to stop alerts service: %v", err)
	}

	if err := s.portfolioService.Stop(); err != nil {
		logrus.Errorf("Failed to stop portfolio service: %v", err)
	}

	if err := s.marketService.Stop(); err != nil {
		logrus.Errorf("Failed to stop market service: %v", err)
	}

	// Close database connection
	if err := s.db.Close(); err != nil {
		logrus.Errorf("Failed to close database connection: %v", err)
	}

	// Close Redis connection
	if err := s.redis.Close(); err != nil {
		logrus.Errorf("Failed to close Redis connection: %v", err)
	}

	logrus.Info("Crypto terminal service stopped")
	return nil
}

// GetHealthStatus returns the health status of the service
func (s *Service) GetHealthStatus() models.HealthCheck {
	health := models.HealthCheck{
		Status:    "ok",
		Timestamp: time.Now(),
		Services:  make(map[string]string),
		Version:   "1.0.0",
		Uptime:    time.Since(s.startTime),
	}

	// Check database health
	if err := s.db.Ping(); err != nil {
		health.Services["database"] = "unhealthy"
		health.Status = "degraded"
	} else {
		health.Services["database"] = "healthy"
	}

	// Check Redis health
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.redis.Ping(ctx).Err(); err != nil {
		health.Services["redis"] = "unhealthy"
		health.Status = "degraded"
	} else {
		health.Services["redis"] = "healthy"
	}

	// Check market service health
	if s.marketService.IsHealthy() {
		health.Services["market_data"] = "healthy"
	} else {
		health.Services["market_data"] = "unhealthy"
		health.Status = "degraded"
	}

	// Check portfolio service health
	if s.portfolioService.IsHealthy() {
		health.Services["portfolio"] = "healthy"
	} else {
		health.Services["portfolio"] = "unhealthy"
		health.Status = "degraded"
	}

	// Check alerts service health
	if s.alertsService.IsHealthy() {
		health.Services["alerts"] = "healthy"
	} else {
		health.Services["alerts"] = "unhealthy"
		health.Status = "degraded"
	}

	// Check order flow service health
	if s.orderflowService.IsHealthy() {
		health.Services["order_flow"] = "healthy"
	} else {
		health.Services["order_flow"] = "unhealthy"
		health.Status = "degraded"
	}

	return health
}

// Market Data Handlers
func (s *Service) GetPrices(c *gin.Context) {
	prices, err := s.marketService.GetAllPrices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      prices,
		Timestamp: time.Now(),
	})
}

func (s *Service) GetPrice(c *gin.Context) {
	symbol := c.Param("symbol")
	price, err := s.marketService.GetPrice(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      price,
		Timestamp: time.Now(),
	})
}

func (s *Service) GetPriceHistory(c *gin.Context) {
	symbol := c.Param("symbol")
	timeframe := c.DefaultQuery("timeframe", "1h")
	limit := c.DefaultQuery("limit", "100")

	history, err := s.marketService.GetPriceHistory(c.Request.Context(), symbol, timeframe, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      history,
		Timestamp: time.Now(),
	})
}

func (s *Service) GetTechnicalIndicators(c *gin.Context) {
	symbol := c.Param("symbol")
	timeframe := c.DefaultQuery("timeframe", "1h")

	indicators, err := s.marketService.GetTechnicalIndicators(c.Request.Context(), symbol, timeframe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      indicators,
		Timestamp: time.Now(),
	})
}

func (s *Service) GetMarketOverview(c *gin.Context) {
	overview, err := s.marketService.GetMarketOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      overview,
		Timestamp: time.Now(),
	})
}

func (s *Service) GetTopGainers(c *gin.Context) {
	gainers, err := s.marketService.GetTopGainers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      gainers,
		Timestamp: time.Now(),
	})
}

func (s *Service) GetTopLosers(c *gin.Context) {
	losers, err := s.marketService.GetTopLosers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      losers,
		Timestamp: time.Now(),
	})
}

func (s *Service) GetTrendingCoins(c *gin.Context) {
	trending, err := s.marketService.GetTrendingCoins(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      trending,
		Timestamp: time.Now(),
	})
}

// Portfolio Handlers
func (s *Service) GetPortfolio(c *gin.Context) {
	userID := c.GetHeader("X-User-ID") // In real app, extract from JWT
	if userID == "" {
		userID = "default-user" // For demo purposes
	}

	portfolios, err := s.portfolioService.GetUserPortfolios(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      portfolios,
		Timestamp: time.Now(),
	})
}

func (s *Service) CreatePortfolio(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) UpdatePortfolio(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) DeletePortfolio(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) GetPortfolioPerformance(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) GetPortfolioHoldings(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) AddHolding(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) UpdateHolding(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) RemoveHolding(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) SyncPortfolio(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) GetRiskMetrics(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) GetDiversificationAnalysis(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

// Alert Handlers
func (s *Service) GetAlerts(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "default-user"
	}

	alerts, err := s.alertsService.GetUserAlerts(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      alerts,
		Timestamp: time.Now(),
	})
}

func (s *Service) CreateAlert(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) UpdateAlert(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) DeleteAlert(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) ActivateAlert(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) DeactivateAlert(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) GetAlertTemplates(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) GetAlertStatistics(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

// DeFi Handlers
func (s *Service) GetLiquidityPools(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) GetYieldOpportunities(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) GetArbitrageOpportunities(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) GetDeFiProtocols(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

// Trading Signals Handlers
func (s *Service) GetTradingSignals(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) GetSignalsForSymbol(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) BacktestSignal(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

// News and Sentiment Handlers
func (s *Service) GetNews(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

func (s *Service) GetSentiment(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success:   false,
		Error:     "Not implemented yet",
		Timestamp: time.Now(),
	})
}

// WebSocket Handlers
func (s *Service) HandleWebSocket(c *gin.Context) {
	s.wsHub.HandleWebSocket(c.Writer, c.Request)
}

func (s *Service) HandleMarketWebSocket(c *gin.Context) {
	s.wsHub.HandleMarketWebSocket(c.Writer, c.Request)
}

func (s *Service) HandlePortfolioWebSocket(c *gin.Context) {
	s.wsHub.HandlePortfolioWebSocket(c.Writer, c.Request)
}

func (s *Service) HandleAlertsWebSocket(c *gin.Context) {
	s.wsHub.HandleAlertsWebSocket(c.Writer, c.Request)
}

// Order Flow Handlers
func (s *Service) GetFootprintData(c *gin.Context) {
	symbol := c.Param("symbol")
	timeframe := c.DefaultQuery("timeframe", "1h")
	startTime := c.DefaultQuery("start_time", "")
	endTime := c.DefaultQuery("end_time", "")

	// Parse time parameters
	var start, end time.Time
	var err error

	if startTime != "" {
		start, err = time.Parse(time.RFC3339, startTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success:   false,
				Error:     "Invalid start_time format",
				Timestamp: time.Now(),
			})
			return
		}
	} else {
		start = time.Now().Add(-24 * time.Hour) // Default to last 24 hours
	}

	if endTime != "" {
		end, err = time.Parse(time.RFC3339, endTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success:   false,
				Error:     "Invalid end_time format",
				Timestamp: time.Now(),
			})
			return
		}
	} else {
		end = time.Now()
	}

	// Create default config
	config := models.OrderFlowConfig{
		Symbol:                symbol,
		TickAggregationMethod: "TIME",
		TimePerRow:           time.Hour,
		PriceTickSize:        decimal.NewFromFloat(0.01),
		ImbalanceThreshold:   decimal.NewFromFloat(70),
		ImbalanceMinVolume:   decimal.NewFromFloat(1000),
		ValueAreaPercentage:  decimal.NewFromFloat(70),
	}

	footprintData, err := s.orderflowService.GetFootprintData(c.Request.Context(), symbol, timeframe, start, end, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      footprintData,
		Timestamp: time.Now(),
	})
}

func (s *Service) GetVolumeProfile(c *gin.Context) {
	symbol := c.Param("symbol")
	profileType := c.DefaultQuery("type", "VPVR")
	startTime := c.DefaultQuery("start_time", "")
	endTime := c.DefaultQuery("end_time", "")

	// Parse time parameters
	var start, end time.Time
	var err error

	if startTime != "" {
		start, err = time.Parse(time.RFC3339, startTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success:   false,
				Error:     "Invalid start_time format",
				Timestamp: time.Now(),
			})
			return
		}
	} else {
		start = time.Now().Add(-24 * time.Hour)
	}

	if endTime != "" {
		end, err = time.Parse(time.RFC3339, endTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success:   false,
				Error:     "Invalid end_time format",
				Timestamp: time.Now(),
			})
			return
		}
	} else {
		end = time.Now()
	}

	// Create default config
	config := models.OrderFlowConfig{
		Symbol:              symbol,
		PriceTickSize:       decimal.NewFromFloat(0.01),
		ValueAreaPercentage: decimal.NewFromFloat(70),
		HVNThreshold:        decimal.NewFromFloat(1.5),
		LVNThreshold:        decimal.NewFromFloat(0.5),
	}

	volumeProfile, err := s.orderflowService.GetVolumeProfile(c.Request.Context(), symbol, profileType, start, end, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      volumeProfile,
		Timestamp: time.Now(),
	})
}

func (s *Service) GetDeltaAnalysis(c *gin.Context) {
	symbol := c.Param("symbol")
	timeframe := c.DefaultQuery("timeframe", "1h")
	startTime := c.DefaultQuery("start_time", "")
	endTime := c.DefaultQuery("end_time", "")

	// Parse time parameters
	var start, end time.Time
	var err error

	if startTime != "" {
		start, err = time.Parse(time.RFC3339, startTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success:   false,
				Error:     "Invalid start_time format",
				Timestamp: time.Now(),
			})
			return
		}
	} else {
		start = time.Now().Add(-24 * time.Hour)
	}

	if endTime != "" {
		end, err = time.Parse(time.RFC3339, endTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success:   false,
				Error:     "Invalid end_time format",
				Timestamp: time.Now(),
			})
			return
		}
	} else {
		end = time.Now()
	}

	// Create default config
	config := models.OrderFlowConfig{
		Symbol:               symbol,
		DeltaSmoothingPeriod: 10,
		EnableDeltaDivergence: true,
	}

	deltaAnalysis, err := s.orderflowService.GetDeltaAnalysis(c.Request.Context(), symbol, timeframe, start, end, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      deltaAnalysis,
		Timestamp: time.Now(),
	})
}

func (s *Service) GetOrderFlowMetrics(c *gin.Context) {
	symbol := c.Param("symbol")

	metrics, err := s.orderflowService.GetOrderFlowMetrics(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      metrics,
		Timestamp: time.Now(),
	})
}

func (s *Service) GetActiveImbalances(c *gin.Context) {
	symbol := c.Param("symbol")

	imbalances, err := s.orderflowService.GetActiveImbalances(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      imbalances,
		Timestamp: time.Now(),
	})
}
