package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/terminal"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	// Load configuration
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logging
	setupLogging(cfg.Logging)

	logrus.WithFields(logrus.Fields{
		"version":    version,
		"build_time": buildTime,
		"git_commit": gitCommit,
	}).Info("Starting Crypto Market Terminal")

	// Create terminal service
	terminalService, err := terminal.NewService(cfg)
	if err != nil {
		logrus.Fatalf("Failed to create terminal service: %v", err)
	}

	// Setup Gin router
	if cfg.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Setup CORS
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		for _, allowedOrigin := range cfg.Security.CORS.AllowedOrigins {
			if origin == allowedOrigin {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup routes
	setupRoutes(router, terminalService)

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.Server.GetServerAddr(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logrus.WithField("address", server.Addr).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Start terminal service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := terminalService.Start(ctx); err != nil {
			logrus.Errorf("Terminal service error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// Cancel context to stop terminal service
	cancel()

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logrus.Errorf("Server forced to shutdown: %v", err)
	}

	// Stop terminal service
	if err := terminalService.Stop(); err != nil {
		logrus.Errorf("Failed to stop terminal service: %v", err)
	}

	logrus.Info("Server exited")
}

func setupLogging(cfg config.LoggingConfig) {
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	if cfg.Format == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	if cfg.Output == "stdout" {
		logrus.SetOutput(os.Stdout)
	}
}

func setupRoutes(router *gin.Engine, service *terminal.Service) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		health := service.GetHealthStatus()
		c.JSON(http.StatusOK, health)
	})

	// Version info
	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":    version,
			"build_time": buildTime,
			"git_commit": gitCommit,
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Market data routes
		market := v1.Group("/market")
		{
			market.GET("/prices", service.GetPrices)
			market.GET("/prices/:symbol", service.GetPrice)
			market.GET("/history/:symbol", service.GetPriceHistory)
			market.GET("/indicators/:symbol", service.GetTechnicalIndicators)
			market.GET("/overview", service.GetMarketOverview)
			market.GET("/gainers", service.GetTopGainers)
			market.GET("/losers", service.GetTopLosers)
			market.GET("/trending", service.GetTrendingCoins)
		}

		// Portfolio routes
		portfolio := v1.Group("/portfolio")
		{
			portfolio.GET("", service.GetPortfolio)
			portfolio.POST("", service.CreatePortfolio)
			portfolio.PUT("/:id", service.UpdatePortfolio)
			portfolio.DELETE("/:id", service.DeletePortfolio)
			portfolio.GET("/:id/performance", service.GetPortfolioPerformance)
			portfolio.GET("/:id/holdings", service.GetPortfolioHoldings)
			portfolio.POST("/:id/holdings", service.AddHolding)
			portfolio.PUT("/:id/holdings/:holding_id", service.UpdateHolding)
			portfolio.DELETE("/:id/holdings/:holding_id", service.RemoveHolding)
			portfolio.POST("/:id/sync", service.SyncPortfolio)
			portfolio.GET("/:id/risk", service.GetRiskMetrics)
			portfolio.GET("/:id/diversification", service.GetDiversificationAnalysis)
		}

		// Alerts routes
		alerts := v1.Group("/alerts")
		{
			alerts.GET("", service.GetAlerts)
			alerts.POST("", service.CreateAlert)
			alerts.PUT("/:id", service.UpdateAlert)
			alerts.DELETE("/:id", service.DeleteAlert)
			alerts.POST("/:id/activate", service.ActivateAlert)
			alerts.POST("/:id/deactivate", service.DeactivateAlert)
			alerts.GET("/templates", service.GetAlertTemplates)
			alerts.GET("/statistics", service.GetAlertStatistics)
		}

		// DeFi routes
		defi := v1.Group("/defi")
		{
			defi.GET("/pools", service.GetLiquidityPools)
			defi.GET("/yield", service.GetYieldOpportunities)
			defi.GET("/arbitrage", service.GetArbitrageOpportunities)
			defi.GET("/protocols", service.GetDeFiProtocols)
		}

		// Trading signals routes
		signals := v1.Group("/signals")
		{
			signals.GET("", service.GetTradingSignals)
			signals.GET("/:symbol", service.GetSignalsForSymbol)
			signals.POST("/backtest", service.BacktestSignal)
		}

		// News and sentiment routes
		news := v1.Group("/news")
		{
			news.GET("", service.GetNews)
			news.GET("/sentiment/:symbol", service.GetSentiment)
		}

		// Order flow routes
		orderflow := v1.Group("/orderflow")
		{
			orderflow.GET("/footprint/:symbol", service.GetFootprintData)
			orderflow.GET("/volume-profile/:symbol", service.GetVolumeProfile)
			orderflow.GET("/delta/:symbol", service.GetDeltaAnalysis)
			orderflow.GET("/metrics/:symbol", service.GetOrderFlowMetrics)
			orderflow.GET("/imbalances/:symbol", service.GetActiveImbalances)
		}

		// HFT routes
		hft := v1.Group("/hft")
		{
			// HFT status and metrics
			hft.GET("/status", service.GetHFTStatus)
			hft.GET("/latency", service.GetHFTLatencyStats)

			// Strategy management
			strategies := hft.Group("/strategies")
			{
				strategies.GET("", service.GetHFTStrategies)
				strategies.POST("/:strategyId/start", service.StartHFTStrategy)
				strategies.POST("/:strategyId/stop", service.StopHFTStrategy)
			}

			// Order management
			orders := hft.Group("/orders")
			{
				orders.GET("", service.GetHFTOrders)
				orders.POST("", service.PlaceHFTOrder)
				orders.DELETE("/:orderId", service.CancelHFTOrder)
			}

			// Position management
			positions := hft.Group("/positions")
			{
				positions.GET("", service.GetHFTPositions)
			}

			// Risk management
			risk := hft.Group("/risk")
			{
				risk.GET("/events", service.GetHFTRiskEvents)
			}
		}
	}

	// WebSocket endpoint
	router.GET("/ws", service.HandleWebSocket)
	router.GET("/ws/market", service.HandleMarketWebSocket)
	router.GET("/ws/portfolio", service.HandlePortfolioWebSocket)
	router.GET("/ws/alerts", service.HandleAlertsWebSocket)
	router.GET("/ws/hft", service.HandleHFTWebSocket)

	// Static files for frontend
	router.Static("/static", "./web/build/static")
	router.StaticFile("/", "./web/build/index.html")
	router.StaticFile("/favicon.ico", "./web/build/favicon.ico")

	// Catch-all for SPA routing
	router.NoRoute(func(c *gin.Context) {
		c.File("./web/build/index.html")
	})
}
