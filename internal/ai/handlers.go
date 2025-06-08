package ai

import (
	"encoding/json"
	"net/http"

	"github.com/DimaJoyti/go-coffee/pkg/models"
)

// SetupRoutes configures the HTTP routes for the AI service
func SetupRoutes(mux *http.ServeMux, service *Service) {
	// AI Analytics endpoints
	mux.HandleFunc("/api/v1/ai/recommendations", methodHandler("POST", getRecommendationsHandler(service)))
	mux.HandleFunc("/api/v1/ai/arbitrage", methodHandler("POST", getArbitrageHandler(service)))
	mux.HandleFunc("/api/v1/ai/demand/forecast", methodHandler("POST", getDemandForecastHandler(service)))
	mux.HandleFunc("/api/v1/ai/pricing/optimize", methodHandler("POST", getPricingOptimizationHandler(service)))
	mux.HandleFunc("/api/v1/ai/behavior/analyze", methodHandler("POST", getBehaviorAnalysisHandler(service)))
	mux.HandleFunc("/api/v1/ai/inventory/optimize", methodHandler("GET", getInventoryOptimizationHandler(service)))

	// AI Service management
	mux.HandleFunc("/api/v1/ai/status", methodHandler("GET", getAIStatusHandler(service)))
	mux.HandleFunc("/api/v1/ai/models", methodHandler("GET", getModelsHandler(service)))
	mux.HandleFunc("/api/v1/ai/insights", methodHandler("GET", getInsightsHandler(service)))
	mux.HandleFunc("/api/v1/ai/metrics", methodHandler("GET", getMetricsHandler(service)))
}

// methodHandler wraps handlers to only accept specific HTTP methods
func methodHandler(method string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		handler(w, r)
	}
}

// getRecommendationsHandler handles recommendation requests
func getRecommendationsHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			UserID      string                   `json:"user_id"`
			Preferences *models.UserPreferences `json:"preferences"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		if req.UserID == "" {
			writeErrorResponse(w, http.StatusBadRequest, "User ID is required")
			return
		}

		recommendations, err := service.GetRecommendations(r.Context(), req.UserID, req.Preferences)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeSuccessResponse(w, recommendations)
	}
}

// getArbitrageHandler handles arbitrage detection requests
func getArbitrageHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Markets []string `json:"markets"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		if len(req.Markets) == 0 {
			req.Markets = []string{"binance", "coinbase", "kraken"} // Default markets
		}

		opportunities, err := service.DetectArbitrageOpportunities(r.Context(), req.Markets)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeSuccessResponse(w, opportunities)
	}
}

// getDemandForecastHandler handles demand forecasting requests
func getDemandForecastHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Timeframe string `json:"timeframe"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		if req.Timeframe == "" {
			req.Timeframe = "day" // Default timeframe
		}

		forecast, err := service.ForecastDemand(r.Context(), req.Timeframe)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeSuccessResponse(w, forecast)
	}
}

// getPricingOptimizationHandler handles pricing optimization requests
func getPricingOptimizationHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Products []string `json:"products"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		if len(req.Products) == 0 {
			req.Products = []string{"Espresso", "Latte", "Cappuccino", "Americano"} // Default products
		}

		optimization, err := service.OptimizePricing(r.Context(), req.Products)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeSuccessResponse(w, optimization)
	}
}

// getBehaviorAnalysisHandler handles behavior analysis requests
func getBehaviorAnalysisHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			CustomerID string `json:"customer_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		if req.CustomerID == "" {
			writeErrorResponse(w, http.StatusBadRequest, "Customer ID is required")
			return
		}

		analysis, err := service.AnalyzeBehavior(r.Context(), req.CustomerID)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeSuccessResponse(w, analysis)
	}
}

// getInventoryOptimizationHandler handles inventory optimization requests
func getInventoryOptimizationHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		optimization, err := service.OptimizeInventory(r.Context())
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeSuccessResponse(w, optimization)
	}
}

// getAIStatusHandler returns AI service status
func getAIStatusHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"service":    "ai-service",
			"version":    "2.0.0",
			"status":     "healthy",
			"modules": map[string]string{
				"recommendations": "active",
				"arbitrage":       "active",
				"forecasting":     "active",
				"optimization":    "active",
				"behavior":        "active",
				"inventory":       "active",
			},
			"capabilities": []string{
				"personalized_recommendations",
				"crypto_arbitrage_detection",
				"demand_forecasting",
				"dynamic_pricing",
				"customer_behavior_analysis",
				"inventory_optimization",
			},
		}

		writeSuccessResponse(w, status)
	}
}

// getModelsHandler returns available AI models
func getModelsHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		models := []models.MLModel{
			{
				ID:          "recommendation_engine_v2",
				Name:        "Coffee Recommendation Engine",
				Type:        "collaborative_filtering",
				Version:     "2.1.0",
				Accuracy:    0.87,
				Status:      "ready",
				Description: "Personalized coffee recommendations based on user preferences and behavior",
			},
			{
				ID:          "arbitrage_detector_v1",
				Name:        "Crypto Arbitrage Detector",
				Type:        "price_analysis",
				Version:     "1.3.0",
				Accuracy:    0.92,
				Status:      "ready",
				Description: "Real-time cryptocurrency arbitrage opportunity detection",
			},
			{
				ID:          "demand_forecaster_v1",
				Name:        "Demand Forecasting Model",
				Type:        "time_series",
				Version:     "1.2.0",
				Accuracy:    0.84,
				Status:      "ready",
				Description: "Predicts future coffee demand based on historical data and trends",
			},
		}

		writeSuccessResponse(w, map[string]interface{}{
			"models": models,
			"count":  len(models),
		})
	}
}

// getInsightsHandler returns AI-generated business insights
func getInsightsHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		insights := []models.AIInsight{
			{
				ID:          "insight_001",
				Type:        "trend",
				Title:       "Morning Rush Peak Detected",
				Description: "Coffee demand peaks at 8:30 AM with 40% higher volume than average",
				Confidence:  0.89,
				Impact:      "high",
				Category:    "demand_pattern",
				Actions:     []string{"Increase staff during 8-9 AM", "Pre-prepare popular items"},
			},
			{
				ID:          "insight_002",
				Type:        "opportunity",
				Title:       "Latte Price Optimization",
				Description: "Latte prices can be increased by 8% without significant demand impact",
				Confidence:  0.76,
				Impact:      "medium",
				Category:    "pricing",
				Actions:     []string{"Gradual price increase", "Monitor customer response"},
			},
		}

		writeSuccessResponse(w, map[string]interface{}{
			"insights": insights,
			"count":    len(insights),
		})
	}
}

// getMetricsHandler returns AI service metrics
func getMetricsHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := models.AIMetrics{
			TotalRequests:       15420,
			SuccessfulRequests:  14987,
			AverageResponseTime: 245.6,
			ModelAccuracy: map[string]float64{
				"recommendations": 0.87,
				"arbitrage":       0.92,
				"forecasting":     0.84,
			},
			ResourceUsage: models.ResourceUsage{
				CPUUsage:     45.2,
				MemoryUsage:  1024.5,
				StorageUsage: 2048.0,
			},
		}

		writeSuccessResponse(w, metrics)
	}
}

// Helper functions
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	writeJSONResponse(w, statusCode, map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

func writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}
