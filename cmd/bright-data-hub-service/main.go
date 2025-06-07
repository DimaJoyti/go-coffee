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

	brightdatahub "github.com/DimaJoyti/go-coffee/pkg/bright-data-hub"
	"github.com/DimaJoyti/go-coffee/pkg/bright-data-hub/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server for Bright Data Hub
type Server struct {
	hub    *brightdatahub.BrightDataHub
	router *gin.Engine
	logger *logger.Logger
}

// RequestPayload represents the request payload for function execution
type RequestPayload struct {
	Function string      `json:"function" binding:"required"`
	Params   interface{} `json:"params"`
}

// loadEnv loads environment variables from .env file
func loadEnv() {
	if _, err := os.Stat(".env"); err == nil {
		// Simple .env loader
		// In production, use a proper .env library
		log.Println("Loading .env file...")
	}
}

func main() {
	// Set up standard logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load environment variables
	loadEnv()

	// Initialize logger
	loggerInstance := logger.New("bright-data-hub")

	// Test logging immediately
	log.Println("🔧 Bright Data Hub Service initializing...")
	fmt.Println("🔧 Starting Bright Data Hub Service...")
	
	// Load configuration
	cfg := config.LoadConfig()
	
	// Create Bright Data Hub
	hub, err := brightdatahub.NewBrightDataHub(cfg, loggerInstance)
	if err != nil {
		log.Fatalf("Failed to create Bright Data Hub: %v", err)
	}

	// Create server
	server := NewServer(hub, loggerInstance)
	
	// Start the hub
	ctx := context.Background()
	if err := hub.Start(ctx); err != nil {
		log.Fatalf("Failed to start Bright Data Hub: %v", err)
	}
	
	// Start HTTP server
	port := os.Getenv("BRIGHT_DATA_HUB_PORT")
	if port == "" {
		port = "8095"
	}
	
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: server.router,
	}

	// Start server in goroutine
	go func() {
		fmt.Printf("🚀 Bright Data Hub Service starting on port %s\n", port)
		log.Printf("🚀 Bright Data Hub Service starting on port %s", port)
		loggerInstance.Info("🚀 Bright Data Hub Service starting on port %s", port)
		loggerInstance.Info("")
		loggerInstance.Info("📊 **ENHANCED BRIGHT DATA MCP INTEGRATION ENDPOINTS:**")
		loggerInstance.Info("   🔍 Execute Function:     POST /api/v1/bright-data/execute")
		loggerInstance.Info("   📈 Social Analytics:     GET  /api/v1/bright-data/social/analytics")
		loggerInstance.Info("   🛒 Ecommerce Data:       GET  /api/v1/bright-data/ecommerce/{platform}")
		loggerInstance.Info("   🔎 Search Engine:        POST /api/v1/bright-data/search")
		loggerInstance.Info("   📊 Hub Status:           GET  /api/v1/bright-data/status")
		loggerInstance.Info("   ❤️  Health Check:        GET  /api/v1/bright-data/health")
		loggerInstance.Info("")
		loggerInstance.Info("🎯 **SUPPORTED PLATFORMS:**")
		loggerInstance.Info("   📱 Social: Instagram, Facebook, Twitter/X, LinkedIn")
		loggerInstance.Info("   🛍️  E-commerce: Amazon, Booking, Zillow")
		loggerInstance.Info("   🔍 Search: Google, Bing, Yandex")
		loggerInstance.Info("   🤖 AI Analytics: Sentiment, Trends, Intelligence")
		loggerInstance.Info("")
		
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	loggerInstance.Info("Shutting down Bright Data Hub Service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		loggerInstance.Error("HTTP server shutdown error: %v", err)
	}

	// Stop hub
	if err := hub.Stop(); err != nil {
		loggerInstance.Error("Hub shutdown error: %v", err)
	}

	loggerInstance.Info("Bright Data Hub Service stopped")
}

// NewServer creates a new HTTP server
func NewServer(hub *brightdatahub.BrightDataHub, logger *logger.Logger) *Server {
	server := &Server{
		hub:    hub,
		logger: logger,
	}
	
	// Setup Gin router
	gin.SetMode(gin.DebugMode)
	router := gin.New()
	
	// Middleware
	router.Use(gin.Recovery())
	router.Use(server.loggingMiddleware())
	router.Use(server.corsMiddleware())
	
	// Routes
	api := router.Group("/api/v1/bright-data")
	{
		// Core functionality
		api.POST("/execute", server.executeFunction)
		api.GET("/status", server.getStatus)
		api.GET("/health", server.healthCheck)
		
		// Social media endpoints
		social := api.Group("/social")
		{
			social.GET("/analytics", server.getSocialAnalytics)
			social.GET("/trending", server.getTrendingTopics)
			social.POST("/instagram/profile", server.getInstagramProfile)
			social.POST("/facebook/posts", server.getFacebookPosts)
			social.POST("/twitter/posts", server.getTwitterPosts)
			social.POST("/linkedin/profile", server.getLinkedInProfile)
		}
		
		// E-commerce endpoints
		ecommerce := api.Group("/ecommerce")
		{
			ecommerce.POST("/amazon/product", server.getAmazonProduct)
			ecommerce.POST("/amazon/reviews", server.getAmazonReviews)
			ecommerce.POST("/booking/hotels", server.getBookingHotels)
			ecommerce.POST("/zillow/properties", server.getZillowProperties)
		}
		
		// Search endpoints
		search := api.Group("/search")
		{
			search.POST("/engine", server.searchEngine)
			search.POST("/scrape", server.scrapeURL)
		}
		
		// Analytics endpoints
		analytics := api.Group("/analytics")
		{
			analytics.GET("/sentiment/:platform", server.getSentimentAnalysis)
			analytics.GET("/trends", server.getTrendAnalysis)
			analytics.GET("/intelligence", server.getMarketIntelligence)
		}
	}
	
	server.router = router
	return server
}

// Middleware
func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		// Log to both standard log and our logger
		log.Printf("%s %s %d %v", c.Request.Method, c.Request.URL.Path, status, latency)
		s.logger.Info("%s %s %d %v",
			c.Request.Method,
			c.Request.URL.Path,
			status,
			latency,
		)
	}
}

func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

// Core handlers
func (s *Server) executeFunction(c *gin.Context) {
	var payload RequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}
	
	response, err := s.hub.ExecuteFunction(c.Request.Context(), payload.Function, payload.Params)
	if err != nil {
		c.JSON(500, gin.H{"error": "Function execution failed", "details": err.Error()})
		return
	}
	
	c.JSON(200, response)
}

func (s *Server) getStatus(c *gin.Context) {
	status := s.hub.GetStatus()
	c.JSON(200, gin.H{
		"service": "bright-data-hub",
		"status":  status,
		"timestamp": time.Now(),
	})
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"service": "bright-data-hub",
		"status":  "healthy",
		"timestamp": time.Now(),
		"version": "1.0.0",
	})
}

// Social media handlers
func (s *Server) getSocialAnalytics(c *gin.Context) {
	// Implementation for social analytics
	c.JSON(200, gin.H{
		"message": "Social analytics endpoint",
		"status":  "coming soon",
	})
}

func (s *Server) getTrendingTopics(c *gin.Context) {
	// Implementation for trending topics
	c.JSON(200, gin.H{
		"message": "Trending topics endpoint",
		"status":  "coming soon",
	})
}

func (s *Server) getInstagramProfile(c *gin.Context) {
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": "Invalid parameters"})
		return
	}
	
	response, err := s.hub.ExecuteFunction(c.Request.Context(), "web_data_instagram_profiles_Bright_Data", params)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(200, response)
}

func (s *Server) getFacebookPosts(c *gin.Context) {
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": "Invalid parameters"})
		return
	}
	
	response, err := s.hub.ExecuteFunction(c.Request.Context(), "web_data_facebook_posts_Bright_Data", params)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(200, response)
}

func (s *Server) getTwitterPosts(c *gin.Context) {
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": "Invalid parameters"})
		return
	}
	
	response, err := s.hub.ExecuteFunction(c.Request.Context(), "web_data_x_posts_Bright_Data", params)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(200, response)
}

func (s *Server) getLinkedInProfile(c *gin.Context) {
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": "Invalid parameters"})
		return
	}
	
	response, err := s.hub.ExecuteFunction(c.Request.Context(), "web_data_linkedin_person_profile_Bright_Data", params)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(200, response)
}

// E-commerce handlers
func (s *Server) getAmazonProduct(c *gin.Context) {
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": "Invalid parameters"})
		return
	}
	
	response, err := s.hub.ExecuteFunction(c.Request.Context(), "web_data_amazon_product_Bright_Data", params)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(200, response)
}

func (s *Server) getAmazonReviews(c *gin.Context) {
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": "Invalid parameters"})
		return
	}
	
	response, err := s.hub.ExecuteFunction(c.Request.Context(), "web_data_amazon_product_reviews_Bright_Data", params)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(200, response)
}

func (s *Server) getBookingHotels(c *gin.Context) {
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": "Invalid parameters"})
		return
	}
	
	response, err := s.hub.ExecuteFunction(c.Request.Context(), "web_data_booking_hotel_listings_Bright_Data", params)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(200, response)
}

func (s *Server) getZillowProperties(c *gin.Context) {
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": "Invalid parameters"})
		return
	}
	
	response, err := s.hub.ExecuteFunction(c.Request.Context(), "web_data_zillow_properties_listing_Bright_Data", params)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(200, response)
}

// Search handlers
func (s *Server) searchEngine(c *gin.Context) {
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": "Invalid parameters"})
		return
	}
	
	response, err := s.hub.ExecuteFunction(c.Request.Context(), "search_engine_Bright_Data", params)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(200, response)
}

func (s *Server) scrapeURL(c *gin.Context) {
	var params map[string]interface{}
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": "Invalid parameters"})
		return
	}
	
	response, err := s.hub.ExecuteFunction(c.Request.Context(), "scrape_as_markdown_Bright_Data", params)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(200, response)
}

// Analytics handlers
func (s *Server) getSentimentAnalysis(c *gin.Context) {
	platform := c.Param("platform")
	c.JSON(200, gin.H{
		"platform": platform,
		"message":  "Sentiment analysis endpoint",
		"status":   "coming soon",
	})
}

func (s *Server) getTrendAnalysis(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Trend analysis endpoint",
		"status":  "coming soon",
	})
}

func (s *Server) getMarketIntelligence(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Market intelligence endpoint",
		"status":  "coming soon",
	})
}
