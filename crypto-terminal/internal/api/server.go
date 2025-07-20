package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-terminal/internal/trading"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Server represents the API server
type Server struct {
	router              *gin.Engine
	httpServer          *http.Server
	coffeeTradingHandler *CoffeeTradingHandler
	webSocketHandler    *WebSocketHandler
	strategyEngine      *trading.StrategyEngine
	logger              *logrus.Logger
	config              *ServerConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	EnableCORS   bool          `json:"enable_cors"`
	EnableTLS    bool          `json:"enable_tls"`
	CertFile     string        `json:"cert_file"`
	KeyFile      string        `json:"key_file"`
}

// NewServer creates a new API server
func NewServer(strategyEngine *trading.StrategyEngine, config *ServerConfig, logger *logrus.Logger) *Server {
	if config == nil {
		config = &ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
			EnableCORS:   true,
			EnableTLS:    false,
		}
	}

	// Set Gin mode based on environment
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Enable CORS if configured
	if config.EnableCORS {
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
		corsConfig.AllowWebSockets = true
		router.Use(cors.New(corsConfig))
	}

	// Create handlers
	coffeeTradingHandler := NewCoffeeTradingHandler(strategyEngine, logger)
	webSocketHandler := NewWebSocketHandler(strategyEngine, logger)

	server := &Server{
		router:               router,
		coffeeTradingHandler: coffeeTradingHandler,
		webSocketHandler:     webSocketHandler,
		strategyEngine:       strategyEngine,
		logger:               logger,
		config:               config,
	}

	// Setup routes
	server.setupRoutes()

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return server
}

// setupRoutes sets up all API routes
func (s *Server) setupRoutes() {
	// Health check endpoint
	s.router.GET("/health", s.healthCheck)
	s.router.GET("/", s.welcomeMessage)

	// API info endpoint
	s.router.GET("/api/v1/info", s.apiInfo)

	// Register coffee trading routes
	s.coffeeTradingHandler.RegisterRoutes(s.router)

	// Register WebSocket routes
	s.webSocketHandler.RegisterRoutes(s.router)

	// Static files for documentation (optional)
	s.router.Static("/docs", "./docs")
}

// Start starts the API server
func (s *Server) Start(ctx context.Context) error {
	s.logger.Infof("Starting Coffee Trading API Server on %s", s.httpServer.Addr)

	// Start WebSocket handler
	if err := s.webSocketHandler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start WebSocket handler: %w", err)
	}

	// Start HTTP server
	go func() {
		var err error
		if s.config.EnableTLS {
			err = s.httpServer.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
		} else {
			err = s.httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			s.logger.Errorf("HTTP server error: %v", err)
		}
	}()

	s.logger.Info("Coffee Trading API Server started successfully")
	s.logger.Infof("API Documentation available at: http://%s/docs", s.httpServer.Addr)
	s.logger.Infof("WebSocket endpoint: ws://%s/ws/coffee-trading", s.httpServer.Addr)

	return nil
}

// Stop stops the API server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping Coffee Trading API Server")

	// Stop WebSocket handler
	if err := s.webSocketHandler.Stop(); err != nil {
		s.logger.Errorf("Error stopping WebSocket handler: %v", err)
	}

	// Stop HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Errorf("Error stopping HTTP server: %v", err)
		return err
	}

	s.logger.Info("Coffee Trading API Server stopped successfully")
	return nil
}

// Health check endpoint
func (s *Server) healthCheck(c *gin.Context) {
	strategies := s.strategyEngine.GetAllStrategies()
	portfolio := s.strategyEngine.GetPortfolio()

	activeStrategies := 0
	for _, strategy := range strategies {
		if strategy.Status == trading.StatusActive {
			activeStrategies++
		}
	}

	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"service":   "coffee-trading-api",
		"uptime":    time.Since(time.Now()).String(), // This would be calculated from start time
		"coffee_theme": gin.H{
			"message": "☕ Coffee Trading API is brewing perfectly!",
			"emoji":   "☕",
		},
		"components": gin.H{
			"strategy_engine": gin.H{
				"status":            "healthy",
				"total_strategies":  len(strategies),
				"active_strategies": activeStrategies,
			},
			"portfolio": gin.H{
				"status":      "healthy",
				"total_value": portfolio.TotalValue,
				"positions":   len(portfolio.Positions),
			},
			"websocket": gin.H{
				"status": "healthy",
			},
		},
	}

	c.JSON(http.StatusOK, health)
}

// Welcome message endpoint
func (s *Server) welcomeMessage(c *gin.Context) {
	welcome := gin.H{
		"message": "☕ Welcome to the Coffee Trading API!",
		"description": "A delightful blend of cryptocurrency trading and coffee culture. " +
			"Our API serves up fresh trading strategies with the perfect aroma of profit.",
		"version": "1.0.0",
		"coffee_menu": gin.H{
			"espresso":   "High-frequency scalping - quick and intense ☕",
			"latte":      "Smooth swing trading - balanced and satisfying 🥛",
			"cold_brew":  "Patient position trading - slow extraction, rich rewards 🧊",
			"cappuccino": "Frothy momentum trading - light, airy, and dynamic 🫖",
		},
		"endpoints": gin.H{
			"health":           "/health",
			"api_info":         "/api/v1/info",
			"strategies":       "/api/v1/coffee-trading/strategies",
			"coffee_menu":      "/api/v1/coffee-trading/coffee/menu",
			"portfolio":        "/api/v1/coffee-trading/portfolio",
			"dashboard":        "/api/v1/coffee-trading/analytics/dashboard",
			"websocket":        "/ws/coffee-trading",
			"documentation":    "/docs",
		},
		"getting_started": gin.H{
			"1": "Check the coffee menu: GET /api/v1/coffee-trading/coffee/menu",
			"2": "Create a strategy: POST /api/v1/coffee-trading/strategies",
			"3": "Start trading: POST /api/v1/coffee-trading/strategies/{id}/start",
			"4": "Monitor via WebSocket: ws://localhost:8080/ws/coffee-trading",
		},
		"support": gin.H{
			"documentation": "Visit /docs for detailed API documentation",
			"websocket":     "Connect to /ws/coffee-trading for real-time updates",
			"health":        "Monitor service health at /health",
		},
	}

	c.JSON(http.StatusOK, welcome)
}

// API info endpoint
func (s *Server) apiInfo(c *gin.Context) {
	info := gin.H{
		"name":        "Coffee Trading API",
		"version":     "1.0.0",
		"description": "A cryptocurrency trading platform with coffee-themed strategies",
		"author":      "Go Coffee Team",
		"license":     "MIT",
		"repository":  "https://github.com/DimaJoyti/go-coffee",
		"coffee_theme": gin.H{
			"philosophy": "Just like brewing the perfect cup of coffee, successful trading requires patience, timing, and the right blend of strategies.",
			"strategies": gin.H{
				"espresso":   "For traders who like their profits quick and intense",
				"latte":      "For those who prefer a smooth, balanced approach",
				"cold_brew":  "For patient traders who believe good things come to those who wait",
				"cappuccino": "For dynamic traders who love riding the market's momentum",
			},
		},
		"features": []string{
			"Coffee-themed trading strategies",
			"Real-time WebSocket updates",
			"Advanced risk management",
			"Portfolio analytics",
			"Binance integration",
			"TradingView data analysis",
			"Performance tracking",
			"Coffee shop correlation analysis",
		},
		"api_endpoints": gin.H{
			"strategies": gin.H{
				"base_url":    "/api/v1/coffee-trading/strategies",
				"methods":     []string{"GET", "POST", "PUT", "DELETE"},
				"description": "Manage trading strategies",
			},
			"coffee": gin.H{
				"base_url":    "/api/v1/coffee-trading/coffee",
				"methods":     []string{"GET", "POST"},
				"description": "Coffee-themed strategy endpoints",
			},
			"portfolio": gin.H{
				"base_url":    "/api/v1/coffee-trading/portfolio",
				"methods":     []string{"GET"},
				"description": "Portfolio management and analytics",
			},
			"signals": gin.H{
				"base_url":    "/api/v1/coffee-trading/signals",
				"methods":     []string{"GET", "POST"},
				"description": "Trading signal management",
			},
			"analytics": gin.H{
				"base_url":    "/api/v1/coffee-trading/analytics",
				"methods":     []string{"GET"},
				"description": "Performance and risk analytics",
			},
		},
		"websocket": gin.H{
			"endpoint":    "/ws/coffee-trading",
			"protocol":    "WebSocket",
			"description": "Real-time updates for prices, signals, trades, and portfolio",
			"channels": []string{
				"price_updates",
				"signal_alerts",
				"trade_executions",
				"portfolio_updates",
				"risk_alerts",
			},
		},
		"rate_limits": gin.H{
			"requests_per_minute": 1000,
			"burst_limit":         100,
			"websocket_connections": 100,
		},
		"timestamp": time.Now(),
	}

	c.JSON(http.StatusOK, info)
}

// GetRouter returns the Gin router (useful for testing)
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// GetHTTPServer returns the HTTP server
func (s *Server) GetHTTPServer() *http.Server {
	return s.httpServer
}

// GetWebSocketHandler returns the WebSocket handler
func (s *Server) GetWebSocketHandler() *WebSocketHandler {
	return s.webSocketHandler
}
