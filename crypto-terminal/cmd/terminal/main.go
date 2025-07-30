package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/api"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/brightdata"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/config"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/health"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/middleware"
	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/terminal"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
	gitCommit = "unknown"
	tracer    trace.Tracer
)

// initTelemetry initializes OpenTelemetry tracing and metrics
func initTelemetry(ctx context.Context, serviceName string) (func(), error) {
	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(),
	)
	if err != nil {
		return nil, err
	}

	// Initialize tracing
	traceExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		logrus.Warnf("Failed to create trace exporter: %v", err)
	}

	var tracerProvider *sdktrace.TracerProvider
	if traceExporter != nil {
		tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(res),
		)
		otel.SetTracerProvider(tracerProvider)
	}

	// Initialize metrics
	promExporter, err := prometheus.New()
	if err != nil {
		logrus.Warnf("Failed to create prometheus exporter: %v", err)
	}

	var meterProvider *sdkmetric.MeterProvider
	if promExporter != nil {
		meterProvider = sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(promExporter),
		)
		otel.SetMeterProvider(meterProvider)
	}

	// Set global tracer
	tracer = otel.Tracer(serviceName)

	// Return cleanup function
	return func() {
		if tracerProvider != nil {
			tracerProvider.Shutdown(ctx)
		}
		if meterProvider != nil {
			meterProvider.Shutdown(ctx)
		}
	}, nil
}

func main() {
	// Load configuration
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logging
	setupLogging(cfg.Logging)

	// Initialize telemetry
	ctx := context.Background()
	cleanup, err := initTelemetry(ctx, "crypto-terminal")
	if err != nil {
		logrus.Warnf("Failed to initialize telemetry: %v", err)
	}
	defer cleanup()

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

	// Create health service
	healthService, err := health.NewService(cfg, nil, nil)
	if err != nil {
		logrus.Fatalf("Failed to create health service: %v", err)
	}

	// Initialize telemetry middleware
	if err := middleware.InitTelemetryMiddleware(); err != nil {
		logrus.Warnf("Failed to initialize telemetry middleware: %v", err)
	}

	// Setup Gin router
	if cfg.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add telemetry and observability middleware
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.TracingMiddleware())
	router.Use(middleware.MetricsMiddleware())
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.ErrorHandlingMiddleware())

	// Setup CORS with proper middleware
	corsConfig := cors.Config{
		AllowOrigins:     cfg.Security.CORS.AllowedOrigins,
		AllowMethods:     cfg.Security.CORS.AllowedMethods,
		AllowHeaders:     cfg.Security.CORS.AllowedHeaders,
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	// If no origins specified, allow all for development
	if len(corsConfig.AllowOrigins) == 0 {
		corsConfig.AllowAllOrigins = true
	}

	router.Use(cors.New(corsConfig))

	// Setup routes
	setupRoutes(router, terminalService, healthService)

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

	// Start services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start health service
	go func() {
		if err := healthService.Start(ctx); err != nil {
			logrus.Errorf("Health service error: %v", err)
		}
	}()

	// Start terminal service
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

func setupRoutes(router *gin.Engine, service *terminal.Service, healthService *health.Service) {
	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		systemHealth := healthService.GetSystemHealth()

		// Return appropriate status code based on health
		statusCode := http.StatusOK
		if systemHealth.Status == health.StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		} else if systemHealth.Status == health.StatusDegraded {
			statusCode = http.StatusPartialContent
		}

		c.JSON(statusCode, systemHealth)
	})

	// Simple liveness probe
	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "alive",
			"timestamp": time.Now(),
		})
	})

	// Readiness probe
	router.GET("/health/ready", func(c *gin.Context) {
		systemHealth := healthService.GetSystemHealth()

		if systemHealth.Status == health.StatusUnhealthy {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "not ready",
				"timestamp": time.Now(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now(),
		})
	})

	// Version info
	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":    version,
			"build_time": buildTime,
			"git_commit": gitCommit,
		})
	})

	// Initialize Bright Data service for TradingView integration
	brightDataService := brightdata.NewService(&brightdata.BrightDataConfig{
		Enabled:         true,
		UpdateInterval:  30 * time.Second,
		MaxConcurrent:   5,
		CacheTTL:        5 * time.Minute,
		RateLimitRPS:    10,
		EnableSentiment: true,
		EnableNews:      true,
		EnableSocial:    true,
		EnableEvents:    true,
	}, nil, logrus.StandardLogger()) // nil for redis client for now

	// Initialize TradingView handlers
	tradingViewHandlers := api.NewTradingViewHandlers(brightDataService, logrus.StandardLogger())

	// Initialize Enhanced Trading handlers
	enhancedTradingHandlers := api.NewEnhancedTradingHandlers(logrus.StandardLogger())

	// API v2 routes for TradingView integration
	v2 := router.Group("/api/v2")
	{
		tradingViewHandlers.RegisterRoutes(v2)
	}

	// API v3 routes for Enhanced Trading interface
	v3 := router.Group("/api/v3")
	{
		enhancedTradingHandlers.RegisterRoutes(v3)
	}

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
