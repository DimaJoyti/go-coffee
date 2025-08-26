package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/analytics"
	"github.com/DimaJoyti/go-coffee/pkg/monitoring"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

const (
	DefaultPort = "8090"
	DefaultHost = "0.0.0.0"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

type Server struct {
	analytics       *analytics.Service
	metrics         *monitoring.PrometheusMetrics
	businessMetrics *monitoring.BusinessMetrics
	clients         map[*websocket.Conn]bool
	broadcast       chan []byte
}

func main() {
	// Initialize monitoring
	metrics := monitoring.NewPrometheusMetrics()
	businessMetrics := monitoring.NewBusinessMetrics(metrics)

	// Initialize analytics service
	analyticsService, err := analytics.NewService(nil, nil)
	if err != nil {
		log.Fatalf("❌ Failed to initialize analytics service: %v", err)
	}

	// Create server instance
	server := &Server{
		analytics:       analyticsService,
		metrics:         metrics,
		businessMetrics: businessMetrics,
		clients:         make(map[*websocket.Conn]bool),
		broadcast:       make(chan []byte),
	}

	// Setup Gin router
	router := gin.Default()

	// Add CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	router.Use(gin.WrapH(c.Handler(router)))

	// Add metrics middleware
	router.Use(func(c *gin.Context) {
		middleware := server.metrics.MetricsMiddleware("analytics-dashboard")
		middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	})

	// Health check
	router.GET("/health", server.healthCheck)

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes
	api := router.Group("/api/v1")
	{
		// Real-time analytics
		api.GET("/realtime", server.realtimeAnalytics)
		api.GET("/ws", server.websocketHandler)

		// Business metrics
		api.GET("/business/overview", server.businessOverview)
		api.GET("/business/revenue", server.revenueAnalytics)
		api.GET("/business/orders", server.orderAnalytics)
		api.GET("/business/products", server.productAnalytics)
		api.GET("/business/locations", server.locationAnalytics)
		api.GET("/business/customers", server.customerAnalytics)

		// DeFi analytics
		api.GET("/defi/portfolio", server.defiPortfolio)
		api.GET("/defi/trading", server.tradingAnalytics)
		api.GET("/defi/yield", server.yieldAnalytics)
		api.GET("/defi/arbitrage", server.arbitrageAnalytics)

		// Technical metrics
		api.GET("/technical/performance", server.performanceMetrics)
		api.GET("/technical/infrastructure", server.infrastructureMetrics)
		api.GET("/technical/ai", server.aiMetrics)
		api.GET("/technical/security", server.securityMetrics)

		// Predictive analytics
		api.GET("/predictions/demand", server.demandPredictions)
		api.GET("/predictions/revenue", server.revenuePredictions)
		api.GET("/predictions/market", server.marketPredictions)

		// Custom dashboards
		api.GET("/dashboards", server.listDashboards)
		api.POST("/dashboards", server.createDashboard)
		api.GET("/dashboards/:id", server.getDashboard)
		api.PUT("/dashboards/:id", server.updateDashboard)
		api.DELETE("/dashboards/:id", server.deleteDashboard)

		// Export functionality
		api.GET("/export/csv", server.exportCSV)
		api.GET("/export/pdf", server.exportPDF)
		api.GET("/export/excel", server.exportExcel)

		// Alerts and notifications
		api.GET("/alerts", server.getAlerts)
		api.POST("/alerts", server.createAlert)
		api.PUT("/alerts/:id", server.updateAlert)
		api.DELETE("/alerts/:id", server.deleteAlert)

		// Comparative analytics
		api.GET("/compare/periods", server.comparePeriods)
		api.GET("/compare/locations", server.compareLocations)
		api.GET("/compare/products", server.compareProducts)
	}

	// Static file serving for dashboard UI
	router.Static("/static", "./web/static")
	router.StaticFile("/", "./web/index.html")

	// Start WebSocket broadcaster
	go server.handleWebSocketMessages()

	// Start background analytics processing
	go server.startAnalyticsProcessing()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = DefaultHost
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Analytics Dashboard Server starting on http://%s", addr)
		log.Printf("📊 Dashboard UI: http://%s", addr)
		log.Printf("📈 API Docs: http://%s/api/v1", addr)
		log.Printf("🔍 Metrics: http://%s/metrics", addr)
		log.Printf("⚡ WebSocket: ws://%s/api/v1/ws", addr)
		
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down analytics dashboard server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Analytics dashboard server stopped")
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "analytics-dashboard",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
		"uptime":    time.Since(time.Now()).String(),
	})
}

func (s *Server) realtimeAnalytics(c *gin.Context) {
	data := s.analytics.GetRealtimeData()
	c.JSON(http.StatusOK, data)
}

func (s *Server) websocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Register client
	s.clients[conn] = true
	log.Printf("🔌 WebSocket client connected (total: %d)", len(s.clients))

	// Handle client disconnection
	defer func() {
		delete(s.clients, conn)
		log.Printf("🔌 WebSocket client disconnected (total: %d)", len(s.clients))
	}()

	// Send initial data
	initialData := s.analytics.GetRealtimeData()
	if data, err := json.Marshal(initialData); err == nil {
		conn.WriteMessage(websocket.TextMessage, data)
	}

	// Keep connection alive and handle messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (s *Server) handleWebSocketMessages() {
	for {
		select {
		case message := <-s.broadcast:
			for client := range s.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					client.Close()
					delete(s.clients, client)
				}
			}
		}
	}
}

func (s *Server) startAnalyticsProcessing() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Update real-time data
			data := s.analytics.GetRealtimeData()
			if jsonData, err := json.Marshal(data); err == nil {
				s.broadcast <- jsonData
			}

			// Record business metrics
			s.recordBusinessMetrics()
		}
	}
}

func (s *Server) recordBusinessMetrics() {
	// Simulate business metrics recording
	metrics := s.analytics.GetCurrentMetrics()
	
	s.businessMetrics.RecordOrder("completed", "crypto", metrics.Revenue)
	s.businessMetrics.UpdateActiveOrders("processing", float64(metrics.ActiveOrders))
	s.businessMetrics.RecordAIPrediction("demand-forecast", "order-prediction", 0.85)
}

func (s *Server) businessOverview(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "24h")
	data := s.analytics.GetBusinessOverview(timeRange)
	c.JSON(http.StatusOK, data)
}

func (s *Server) revenueAnalytics(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "7d")
	granularity := c.DefaultQuery("granularity", "daily")
	data := s.analytics.GetRevenueAnalytics(timeRange, granularity)
	c.JSON(http.StatusOK, data)
}

func (s *Server) orderAnalytics(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "7d")
	data := s.analytics.GetOrderAnalytics(timeRange)
	c.JSON(http.StatusOK, data)
}

func (s *Server) productAnalytics(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "30d")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	data := s.analytics.GetProductAnalytics(timeRange, limit)
	c.JSON(http.StatusOK, data)
}

func (s *Server) locationAnalytics(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "30d")
	data := s.analytics.GetLocationAnalytics(timeRange)
	c.JSON(http.StatusOK, data)
}

func (s *Server) customerAnalytics(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "30d")
	data := s.analytics.GetCustomerAnalytics(timeRange)
	c.JSON(http.StatusOK, data)
}

func (s *Server) defiPortfolio(c *gin.Context) {
	data := s.analytics.GetDeFiPortfolio()
	c.JSON(http.StatusOK, data)
}

func (s *Server) tradingAnalytics(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "7d")
	data := s.analytics.GetTradingAnalytics(timeRange)
	c.JSON(http.StatusOK, data)
}

func (s *Server) yieldAnalytics(c *gin.Context) {
	data := s.analytics.GetYieldAnalytics()
	c.JSON(http.StatusOK, data)
}

func (s *Server) arbitrageAnalytics(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "24h")
	data := s.analytics.GetArbitrageAnalytics(timeRange)
	c.JSON(http.StatusOK, data)
}

func (s *Server) performanceMetrics(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "1h")
	data := s.analytics.GetPerformanceMetrics(timeRange)
	c.JSON(http.StatusOK, data)
}

func (s *Server) infrastructureMetrics(c *gin.Context) {
	data := s.analytics.GetInfrastructureMetrics()
	c.JSON(http.StatusOK, data)
}

func (s *Server) aiMetrics(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "24h")
	data := s.analytics.GetAIMetrics(timeRange)
	c.JSON(http.StatusOK, data)
}

func (s *Server) securityMetrics(c *gin.Context) {
	timeRange := c.DefaultQuery("range", "24h")
	data := s.analytics.GetSecurityMetrics(timeRange)
	c.JSON(http.StatusOK, data)
}

func (s *Server) demandPredictions(c *gin.Context) {
	horizon := c.DefaultQuery("horizon", "7d")
	data := s.analytics.GetDemandPredictions(horizon)
	c.JSON(http.StatusOK, data)
}

func (s *Server) revenuePredictions(c *gin.Context) {
	horizon := c.DefaultQuery("horizon", "30d")
	data := s.analytics.GetRevenuePredictions(horizon)
	c.JSON(http.StatusOK, data)
}

func (s *Server) marketPredictions(c *gin.Context) {
	horizon := c.DefaultQuery("horizon", "24h")
	assets := strings.Split(c.DefaultQuery("assets", "BTC,ETH"), ",")
	data := s.analytics.GetMarketPredictions(horizon, assets)
	c.JSON(http.StatusOK, data)
}

func (s *Server) listDashboards(c *gin.Context) {
	dashboards := s.analytics.GetDashboards()
	c.JSON(http.StatusOK, dashboards)
}

func (s *Server) createDashboard(c *gin.Context) {
	var dashboard analytics.Dashboard
	if err := c.ShouldBindJSON(&dashboard); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created := s.analytics.CreateDashboard(dashboard)
	c.JSON(http.StatusCreated, created)
}

func (s *Server) getDashboard(c *gin.Context) {
	id := c.Param("id")
	dashboard, exists := s.analytics.GetDashboardWithExists(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dashboard not found"})
		return
	}
	c.JSON(http.StatusOK, dashboard)
}

func (s *Server) updateDashboard(c *gin.Context) {
	id := c.Param("id")
	var dashboard analytics.Dashboard
	if err := c.ShouldBindJSON(&dashboard); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, exists := s.analytics.UpdateDashboard(id, dashboard)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dashboard not found"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (s *Server) deleteDashboard(c *gin.Context) {
	id := c.Param("id")
	if s.analytics.DeleteDashboard(id) {
		c.JSON(http.StatusOK, gin.H{"message": "Dashboard deleted"})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dashboard not found"})
	}
}

func (s *Server) getAlerts(c *gin.Context) {
	alerts := s.analytics.GetAlerts()
	c.JSON(http.StatusOK, alerts)
}

func (s *Server) createAlert(c *gin.Context) {
	var alert analytics.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created := s.analytics.CreateAlert(alert)
	c.JSON(http.StatusCreated, created)
}

func (s *Server) updateAlert(c *gin.Context) {
	id := c.Param("id")
	var alert analytics.Alert
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, exists := s.analytics.UpdateAlert(id, alert)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (s *Server) deleteAlert(c *gin.Context) {
	id := c.Param("id")
	if s.analytics.DeleteAlert(id) {
		c.JSON(http.StatusOK, gin.H{"message": "Alert deleted"})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
	}
}

func (s *Server) exportCSV(c *gin.Context) {
	dataType := c.Query("type")
	timeRange := c.DefaultQuery("range", "30d")
	
	data := s.analytics.ExportData(dataType, timeRange, "csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=analytics_%s_%s.csv", dataType, timeRange))
	c.Header("Content-Type", "text/csv")
	c.String(http.StatusOK, data)
}

func (s *Server) exportPDF(c *gin.Context) {
	reportType := c.Query("type")
	timeRange := c.DefaultQuery("range", "30d")
	
	data := s.analytics.ExportData(reportType, timeRange, "pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=report_%s_%s.pdf", reportType, timeRange))
	c.Header("Content-Type", "application/pdf")
	c.String(http.StatusOK, data)
}

func (s *Server) exportExcel(c *gin.Context) {
	dataType := c.Query("type")
	timeRange := c.DefaultQuery("range", "30d")
	
	data := s.analytics.ExportData(dataType, timeRange, "excel")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=analytics_%s_%s.xlsx", dataType, timeRange))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.String(http.StatusOK, data)
}

func (s *Server) comparePeriods(c *gin.Context) {
	period1 := c.Query("period1")
	period2 := c.Query("period2")
	metric := c.Query("metric")
	
	data := s.analytics.ComparePeriods(period1, period2, metric)
	c.JSON(http.StatusOK, data)
}

func (s *Server) compareLocations(c *gin.Context) {
	locations := strings.Split(c.Query("locations"), ",")
	metric := c.Query("metric")
	timeRange := c.DefaultQuery("range", "30d")
	
	data := s.analytics.CompareLocations(locations, metric, timeRange)
	c.JSON(http.StatusOK, data)
}

func (s *Server) compareProducts(c *gin.Context) {
	products := strings.Split(c.Query("products"), ",")
	metric := c.Query("metric")
	timeRange := c.DefaultQuery("range", "30d")
	
	data := s.analytics.CompareProducts(products, metric, timeRange)
	c.JSON(http.StatusOK, data)
}