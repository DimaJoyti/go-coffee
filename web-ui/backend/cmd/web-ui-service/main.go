package main

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/DimaJoyti/go-coffee/web-ui/backend/internal/handlers"
	"github.com/DimaJoyti/go-coffee/web-ui/backend/internal/services"
	"github.com/DimaJoyti/go-coffee/web-ui/backend/internal/websocket"
)

// loadEnv loads environment variables from .env file
func loadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "43200")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Health check handler
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"service":   "go-coffee-web-ui",
		"version":   "1.0.0",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Load environment variables from .env file
	envPath := filepath.Join("..", "..", ".env")
	if err := loadEnv(envPath); err != nil {
		log.Printf("Warning: Could not load .env file from %s: %v", envPath, err)
		// Try alternative path
		if err := loadEnv("../../.env"); err != nil {
			log.Printf("Warning: Could not load .env file from alternative path: %v", err)
		}
	} else {
		log.Printf("✅ Loaded environment variables from %s", envPath)
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
	router := mux.NewRouter()

	// Apply CORS middleware
	router.Use(corsMiddleware)

	// Health check
	router.HandleFunc("/health", healthHandler).Methods("GET")

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Dashboard routes
	dashboard := api.PathPrefix("/dashboard").Subrouter()
	dashboard.HandleFunc("/metrics", dashboardHandler.GetMetrics).Methods("GET")
	dashboard.HandleFunc("/activity", dashboardHandler.GetActivity).Methods("GET")

	// Coffee routes
	coffee := api.PathPrefix("/coffee").Subrouter()
	coffee.HandleFunc("/orders", coffeeHandler.GetOrders).Methods("GET")
	coffee.HandleFunc("/orders", coffeeHandler.CreateOrder).Methods("POST")
	coffee.HandleFunc("/orders/{id}", coffeeHandler.UpdateOrder).Methods("PUT")
	coffee.HandleFunc("/inventory", coffeeHandler.GetInventory).Methods("GET")

	// DeFi routes
	defi := api.PathPrefix("/defi").Subrouter()
	defi.HandleFunc("/portfolio", defiHandler.GetPortfolio).Methods("GET")
	defi.HandleFunc("/assets", defiHandler.GetAssets).Methods("GET")
	defi.HandleFunc("/strategies", defiHandler.GetStrategies).Methods("GET")
	defi.HandleFunc("/strategies/{id}/toggle", defiHandler.ToggleStrategy).Methods("POST")

	// AI Agents routes
	agents := api.PathPrefix("/agents").Subrouter()
	agents.HandleFunc("/status", agentsHandler.GetAgentsStatus).Methods("GET")
	agents.HandleFunc("/agents/{id}/toggle", agentsHandler.ToggleAgent).Methods("POST")
	agents.HandleFunc("/agents/{id}/logs", agentsHandler.GetAgentLogs).Methods("GET")

	// Scraping routes (Bright Data)
	scraping := api.PathPrefix("/scraping").Subrouter()
	scraping.HandleFunc("/data", scrapingHandler.GetMarketData).Methods("GET")
	scraping.HandleFunc("/refresh", scrapingHandler.RefreshData).Methods("POST")
	scraping.HandleFunc("/sources", scrapingHandler.GetDataSources).Methods("GET")
	scraping.HandleFunc("/competitors", scrapingHandler.GetCompetitorData).Methods("GET")
	scraping.HandleFunc("/news", scrapingHandler.GetMarketNews).Methods("GET")
	scraping.HandleFunc("/futures", scrapingHandler.GetCoffeeFutures).Methods("GET")
	scraping.HandleFunc("/social", scrapingHandler.GetSocialTrends).Methods("GET")
	scraping.HandleFunc("/stats", scrapingHandler.GetSessionStats).Methods("GET")
	scraping.HandleFunc("/url", scrapingHandler.ScrapeURL).Methods("POST")
	scraping.HandleFunc("/search", scrapingHandler.SearchEngine).Methods("POST")

	// Analytics routes
	analytics := api.PathPrefix("/analytics").Subrouter()
	analytics.HandleFunc("/sales", analyticsHandler.GetSalesData).Methods("GET")
	analytics.HandleFunc("/revenue", analyticsHandler.GetRevenueData).Methods("GET")
	analytics.HandleFunc("/products", analyticsHandler.GetTopProducts).Methods("GET")
	analytics.HandleFunc("/locations", analyticsHandler.GetLocationPerformance).Methods("GET")

	// WebSocket endpoint
	router.HandleFunc("/ws/realtime", wsHandler.HandleWebSocket).Methods("GET")

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
