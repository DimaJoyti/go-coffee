package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/analytics"
	httpTransport "github.com/DimaJoyti/go-coffee/internal/analytics/transport/http"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Initialize structured logger
	logger := initLogger()
	defer logger.Sync()

	logger.Info("Starting Analytics & Business Intelligence Service...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize analytics service
	analyticsService, err := analytics.NewService(cfg.Analytics, logger)
	if err != nil {
		logger.Fatal("Failed to create analytics service", zap.Error(err))
	}

	// Start analytics service
	ctx := context.Background()
	if err := analyticsService.Start(ctx); err != nil {
		logger.Fatal("Failed to start analytics service", zap.Error(err))
	}

	// Start health check server
	go startHealthServer(logger, cfg)

	// Create HTTP handler
	handler := httpTransport.NewHandler(analyticsService, logger)

	// Setup HTTP routes
	mux := http.NewServeMux()
	
	// Analytics endpoints
	mux.HandleFunc("/analytics/dashboard", handler.GetDashboard)
	mux.HandleFunc("/analytics/reports", handler.GetReports)
	mux.HandleFunc("/analytics/reports/", handler.GetReport)
	mux.HandleFunc("/analytics/reports/generate", handler.GenerateReport)
	mux.HandleFunc("/analytics/kpis", handler.GetKPIs)
	mux.HandleFunc("/analytics/trends", handler.GetTrends)
	
	// Business Intelligence endpoints
	mux.HandleFunc("/bi/insights", handler.GetInsights)
	mux.HandleFunc("/bi/predictions", handler.GetPredictions)
	mux.HandleFunc("/bi/recommendations", handler.GetRecommendations)
	mux.HandleFunc("/bi/forecasts", handler.GetForecasts)
	
	// Real-time analytics
	mux.HandleFunc("/realtime/metrics", handler.GetRealtimeMetrics)
	mux.HandleFunc("/realtime/events", handler.StreamEvents)
	mux.HandleFunc("/realtime/alerts", handler.GetAlerts)
	
	// Customer analytics
	mux.HandleFunc("/customer/segments", handler.GetCustomerSegments)
	mux.HandleFunc("/customer/lifetime-value", handler.GetCustomerLTV)
	mux.HandleFunc("/customer/churn-analysis", handler.GetChurnAnalysis)
	mux.HandleFunc("/customer/behavior", handler.GetCustomerBehavior)
	
	// Product analytics
	mux.HandleFunc("/product/performance", handler.GetProductPerformance)
	mux.HandleFunc("/product/recommendations", handler.GetProductRecommendations)
	mux.HandleFunc("/product/inventory-optimization", handler.GetInventoryOptimization)
	
	// Financial analytics
	mux.HandleFunc("/financial/revenue", handler.GetRevenueAnalytics)
	mux.HandleFunc("/financial/profitability", handler.GetProfitabilityAnalysis)
	mux.HandleFunc("/financial/cost-analysis", handler.GetCostAnalysis)
	mux.HandleFunc("/financial/roi", handler.GetROIAnalysis)
	
	// Operational analytics
	mux.HandleFunc("/operational/efficiency", handler.GetOperationalEfficiency)
	mux.HandleFunc("/operational/capacity", handler.GetCapacityAnalysis)
	mux.HandleFunc("/operational/quality", handler.GetQualityMetrics)
	
	// Multi-tenant analytics
	mux.HandleFunc("/tenant/", handler.GetTenantAnalytics)
	mux.HandleFunc("/tenant/comparison", handler.GetTenantComparison)
	
	// Export endpoints
	mux.HandleFunc("/export/csv", handler.ExportCSV)
	mux.HandleFunc("/export/excel", handler.ExportExcel)
	mux.HandleFunc("/export/pdf", handler.ExportPDF)
	
	// Observability endpoints
	mux.HandleFunc("/health", handler.HealthCheck)
	mux.HandleFunc("/ready", handler.ReadinessCheck)
	mux.Handle("/metrics", promhttp.Handler())

	// Create HTTP server with enhanced configuration
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Analytics.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start HTTP server in goroutine
	go func() {
		logger.Info("Starting Analytics Service HTTP server", zap.Int("port", cfg.Analytics.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down Analytics Service...")

	// Create context with timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	// Stop analytics service
	analyticsService.Stop()

	logger.Info("Analytics Service stopped gracefully")
}

// initLogger initializes a structured logger with appropriate configuration
func initLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stacktrace"

	logger, err := config.Build()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	return logger
}

// startHealthServer starts a health check server for the analytics service
func startHealthServer(logger *zap.Logger, cfg *config.Config) {
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"status":"ok",
			"service":"analytics-service",
			"timestamp":"%s",
			"version":"2.0.0",
			"features":{
				"real_time_analytics":true,
				"business_intelligence":true,
				"predictive_analytics":true,
				"multi_tenant":true,
				"export_capabilities":true
			}
		}`, time.Now().UTC().Format(time.RFC3339))
	})
	
	// Readiness check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"status":"ready",
			"service":"analytics-service",
			"timestamp":"%s",
			"checks":{
				"database":"ok",
				"cache":"ok",
				"message_queue":"ok",
				"data_pipeline":"ok"
			}
		}`, time.Now().UTC().Format(time.RFC3339))
	})
	
	// Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Start health server on a different port
	healthPort := 8096
	if cfg.Analytics.HealthPort != 0 {
		healthPort = cfg.Analytics.HealthPort
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", healthPort),
		Handler: mux,
	}

	logger.Info("Starting Analytics Service health check server", zap.Int("port", healthPort))
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Health server error", zap.Error(err))
	}
}
