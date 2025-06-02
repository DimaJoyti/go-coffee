package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DimaJoyti/go-coffee/web-ui/backend/internal/handlers"
	"github.com/DimaJoyti/go-coffee/web-ui/backend/internal/services"
	"github.com/DimaJoyti/go-coffee/web-ui/backend/internal/websocket"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize services
	dashboardService := services.NewDashboardService()
	coffeeService := services.NewCoffeeService()
	defiService := services.NewDefiService()
	agentsService := services.NewAgentsService()
	scrapingService := services.NewScrapingService()
	analyticsService := services.NewAnalyticsService()

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Initialize handlers
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	coffeeHandler := handlers.NewCoffeeHandler(coffeeService)
	defiHandler := handlers.NewDefiHandler(defiService)
	agentsHandler := handlers.NewAgentsHandler(agentsService)
	scrapingHandler := handlers.NewScrapingHandler(scrapingService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	wsHandler := handlers.NewWebSocketHandler(wsHub)

	// Setup router
	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
			"service":   "go-coffee-web-ui",
			"version":   "1.0.0",
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Dashboard routes
		dashboard := api.Group("/dashboard")
		{
			dashboard.GET("/metrics", dashboardHandler.GetMetrics)
			dashboard.GET("/activity", dashboardHandler.GetActivity)
		}

		// Coffee routes
		coffee := api.Group("/coffee")
		{
			coffee.GET("/orders", coffeeHandler.GetOrders)
			coffee.POST("/orders", coffeeHandler.CreateOrder)
			coffee.PUT("/orders/:id", coffeeHandler.UpdateOrder)
			coffee.GET("/inventory", coffeeHandler.GetInventory)
		}

		// DeFi routes
		defi := api.Group("/defi")
		{
			defi.GET("/portfolio", defiHandler.GetPortfolio)
			defi.GET("/assets", defiHandler.GetAssets)
			defi.GET("/strategies", defiHandler.GetStrategies)
			defi.POST("/strategies/:id/toggle", defiHandler.ToggleStrategy)
		}

		// AI Agents routes
		agents := api.Group("/agents")
		{
			agents.GET("/status", agentsHandler.GetAgentsStatus)
			agents.POST("/agents/:id/toggle", agentsHandler.ToggleAgent)
			agents.GET("/agents/:id/logs", agentsHandler.GetAgentLogs)
		}

		// Scraping routes (Bright Data)
		scraping := api.Group("/scraping")
		{
			scraping.GET("/data", scrapingHandler.GetMarketData)
			scraping.POST("/refresh", scrapingHandler.RefreshData)
			scraping.GET("/sources", scrapingHandler.GetDataSources)
		}

		// Analytics routes
		analytics := api.Group("/analytics")
		{
			analytics.GET("/sales", analyticsHandler.GetSalesData)
			analytics.GET("/revenue", analyticsHandler.GetRevenueData)
			analytics.GET("/products", analyticsHandler.GetTopProducts)
			analytics.GET("/locations", analyticsHandler.GetLocationPerformance)
		}
	}

	// WebSocket endpoint
	router.GET("/ws/realtime", wsHandler.HandleWebSocket)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		log.Printf("🚀 Go Coffee Web UI Server starting on port %s", port)
		log.Printf("📊 Dashboard: http://localhost:%s", port)
		log.Printf("🔗 WebSocket: ws://localhost:%s/ws/realtime", port)
		log.Printf("❤️  Health: http://localhost:%s/health", port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("❌ Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited")
}
