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

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/ai"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/common"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/content"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/rag"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/reddit"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/kafka"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/redis"
)

// ContentAnalysisService represents the main service
type ContentAnalysisService struct {
	config        *config.Config
	logger        *logger.Logger
	aiService     ai.Service
	redditService *reddit.Service
	ragService    *rag.Service
	analyzer      *content.Analyzer
	cache         redis.Client
	producer      kafka.Producer
	httpServer    *http.Server
}

func main() {
	fmt.Println("Starting Content Analysis Service...")

	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := logger.NewLogger(cfg.Logging)
	defer logger.Sync()

	// Initialize Redis client
	redisClient, err := redis.NewClientFromConfig(&cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to create Redis client", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize Kafka producer
	kafkaProducer, err := kafka.NewProducer(kafka.Config{
		Brokers: []string{"localhost:9092"}, // Use from config
	}, logger)
	if err != nil {
		logger.Fatal("Failed to create Kafka producer", zap.Error(err))
	}
	defer kafkaProducer.Close()

	// Initialize AI service
	aiService, err := ai.NewService(cfg.AI, logger, redisClient)
	if err != nil {
		logger.Fatal("Failed to create AI service", zap.Error(err))
	}

	// Initialize content analyzer
	analyzer := content.NewAnalyzer(cfg.AI.RAG.ContentAnalysis, logger, aiService, redisClient)

	// Initialize Reddit service
	redditService, err := reddit.NewService(cfg.AI.Reddit, logger, analyzer, redisClient, kafkaProducer)
	if err != nil {
		logger.Fatal("Failed to create Reddit service", zap.Error(err))
	}

	// Initialize RAG service (placeholder - would need vector store implementation)
	// ragService := rag.NewService(cfg.AI.RAG, logger, aiService, redisClient, vectorStore, embeddings)

	// Create main service
	service := &ContentAnalysisService{
		config:        cfg,
		logger:        logger,
		aiService:     aiService,
		redditService: redditService,
		analyzer:      analyzer,
		cache:         redisClient,
		producer:      kafkaProducer,
	}

	// Initialize HTTP server
	service.initHTTPServer()

	// Start services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Reddit service
	if err := service.redditService.Start(ctx); err != nil {
		logger.Fatal("Failed to start Reddit service", zap.Error(err))
	}

	// Start HTTP server
	go func() {
		logger.Info("Starting HTTP server", zap.Int("port", cfg.Server.Port))
		if err := service.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	logger.Info("Content Analysis Service started successfully")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down Content Analysis Service...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop Reddit service
	if err := service.redditService.Stop(); err != nil {
		logger.Error("Error stopping Reddit service", zap.Error(err))
	}

	// Stop HTTP server
	if err := service.httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error stopping HTTP server", zap.Error(err))
	}

	logger.Info("Content Analysis Service stopped")
}

// initHTTPServer initializes the HTTP server with routes
func (s *ContentAnalysisService) initHTTPServer() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Health check endpoint
	router.GET("/health", s.healthCheck)

	// Service status endpoint
	router.GET("/status", s.getStatus)

	// Reddit service endpoints
	redditGroup := router.Group("/api/v1/reddit")
	{
		redditGroup.GET("/stats", s.getRedditStats)
		redditGroup.POST("/analyze/post", s.analyzePost)
		redditGroup.POST("/analyze/comment", s.analyzeComment)
		redditGroup.GET("/search", s.searchContent)
	}

	// Content analysis endpoints
	analysisGroup := router.Group("/api/v1/analysis")
	{
		analysisGroup.POST("/classify", s.classifyContent)
		analysisGroup.POST("/sentiment", s.analyzeSentiment)
		analysisGroup.POST("/topics", s.extractTopics)
	}

	// RAG endpoints (placeholder)
	ragGroup := router.Group("/api/v1/rag")
	{
		ragGroup.POST("/query", s.ragQuery)
		ragGroup.POST("/index", s.indexDocuments)
		ragGroup.GET("/search", s.searchDocuments)
	}

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Server.Port),
		Handler: router,
	}
}

// HTTP handlers

func (s *ContentAnalysisService) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"service":   "content-analysis",
		"version":   "1.0.0",
	})
}

func (s *ContentAnalysisService) getStatus(c *gin.Context) {
	redditStats := s.redditService.GetStats()
	
	c.JSON(http.StatusOK, gin.H{
		"service": gin.H{
			"name":    "Content Analysis Service",
			"version": "1.0.0",
			"uptime":  time.Since(redditStats.StartedAt).String(),
			"status":  "running",
		},
		"reddit": gin.H{
			"running":            s.redditService.IsRunning(),
			"posts_collected":    redditStats.PostsCollected,
			"comments_collected": redditStats.CommentsCollected,
			"posts_analyzed":     redditStats.PostsAnalyzed,
			"comments_analyzed":  redditStats.CommentsAnalyzed,
			"errors_count":       redditStats.ErrorsCount,
			"last_collection":    redditStats.LastCollectionAt,
			"last_analysis":      redditStats.LastAnalysisAt,
		},
		"timestamp": time.Now(),
	})
}

func (s *ContentAnalysisService) getRedditStats(c *gin.Context) {
	stats := s.redditService.GetStats()
	c.JSON(http.StatusOK, stats)
}

func (s *ContentAnalysisService) analyzePost(c *gin.Context) {
	var post common.RedditPost
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	classification, err := s.analyzer.AnalyzePost(c.Request.Context(), &post)
	if err != nil {
		s.logger.Error("Failed to analyze post", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze post"})
		return
	}

	c.JSON(http.StatusOK, classification)
}

func (s *ContentAnalysisService) analyzeComment(c *gin.Context) {
	var comment common.RedditComment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	classification, err := s.analyzer.AnalyzeComment(c.Request.Context(), &comment)
	if err != nil {
		s.logger.Error("Failed to analyze comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze comment"})
		return
	}

	c.JSON(http.StatusOK, classification)
}

func (s *ContentAnalysisService) searchContent(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	// This would implement Reddit content search
	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"results": []interface{}{},
		"message": "Search functionality not yet implemented",
	})
}

func (s *ContentAnalysisService) classifyContent(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// This would implement direct content classification
	c.JSON(http.StatusOK, gin.H{
		"text":           req.Text,
		"classification": "general",
		"confidence":     0.8,
		"message":        "Direct classification not yet implemented",
	})
}

func (s *ContentAnalysisService) analyzeSentiment(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// This would implement direct sentiment analysis
	c.JSON(http.StatusOK, gin.H{
		"text":      req.Text,
		"sentiment": "neutral",
		"score":     0.5,
		"message":   "Direct sentiment analysis not yet implemented",
	})
}

func (s *ContentAnalysisService) extractTopics(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// This would implement direct topic extraction
	c.JSON(http.StatusOK, gin.H{
		"text":    req.Text,
		"topics":  []string{},
		"message": "Direct topic extraction not yet implemented",
	})
}

func (s *ContentAnalysisService) ragQuery(c *gin.Context) {
	var req struct {
		Query string `json:"query" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// This would implement RAG query
	c.JSON(http.StatusOK, gin.H{
		"query":    req.Query,
		"response": "RAG functionality not yet implemented",
		"sources":  []interface{}{},
	})
}

func (s *ContentAnalysisService) indexDocuments(c *gin.Context) {
	var req struct {
		Documents []interface{} `json:"documents" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// This would implement document indexing
	c.JSON(http.StatusOK, gin.H{
		"indexed":   len(req.Documents),
		"message":   "Document indexing not yet implemented",
		"job_id":    fmt.Sprintf("job_%d", time.Now().UnixNano()),
	})
}

func (s *ContentAnalysisService) searchDocuments(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	// This would implement document search
	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"results": []interface{}{},
		"message": "Document search not yet implemented",
	})
}
