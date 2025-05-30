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
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	circuitbreaker "github.com/DimaJoyti/go-coffee/pkg/circuit-breaker"
	eventsourcing "github.com/DimaJoyti/go-coffee/pkg/event-sourcing"
	ratelimiter "github.com/DimaJoyti/go-coffee/pkg/rate-limiter"
	redisadvanced "github.com/DimaJoyti/go-coffee/pkg/redis-advanced"
)

// AgentOrchestrator manages and coordinates AI agents
type AgentOrchestrator struct {
	redis           *redis.Client
	logger          *zap.Logger
	router          *gin.Engine
	searchClient    *redisadvanced.SearchClient
	jsonClient      *redisadvanced.JSONClient
	timeSeriesClient *redisadvanced.TimeSeriesClient
	streamsClient   *redisadvanced.StreamsClient
	eventStore      *eventsourcing.EventStore
	circuitBreaker  *circuitbreaker.HTTPMiddleware
	rateLimiter     *ratelimiter.HTTPMiddleware
	agents          map[string]*AgentInfo
	config          *OrchestratorConfig
}

// OrchestratorConfig contains orchestrator configuration
type OrchestratorConfig struct {
	Port            string        `json:"port"`
	RedisURL        string        `json:"redis_url"`
	LogLevel        string        `json:"log_level"`
	HealthInterval  time.Duration `json:"health_interval"`
	MetricsInterval time.Duration `json:"metrics_interval"`
	MaxAgents       int           `json:"max_agents"`
}

// AgentInfo represents information about a registered agent
type AgentInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Version     string                 `json:"version"`
	Endpoint    string                 `json:"endpoint"`
	Status      string                 `json:"status"`
	Capabilities []string              `json:"capabilities"`
	Metadata    map[string]interface{} `json:"metadata"`
	LastSeen    time.Time              `json:"last_seen"`
	RegisteredAt time.Time             `json:"registered_at"`
	Health      *AgentHealth           `json:"health"`
}

// AgentHealth represents agent health information
type AgentHealth struct {
	Status      string    `json:"status"`
	LastCheck   time.Time `json:"last_check"`
	ResponseTime time.Duration `json:"response_time"`
	ErrorCount  int       `json:"error_count"`
	Uptime      time.Duration `json:"uptime"`
}

// TaskRequest represents a task request for agents
type TaskRequest struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Priority    int                    `json:"priority"`
	Data        map[string]interface{} `json:"data"`
	RequiredCapabilities []string      `json:"required_capabilities"`
	Timeout     time.Duration          `json:"timeout"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TaskResponse represents a task response from agents
type TaskResponse struct {
	TaskID    string                 `json:"task_id"`
	AgentID   string                 `json:"agent_id"`
	Status    string                 `json:"status"`
	Result    map[string]interface{} `json:"result"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
}

func main() {
	log.Println("🚀 Starting Agent Orchestrator...")

	// Load configuration
	config := loadConfig()

	// Initialize logger
	logger, err := initLogger(config.LogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize Redis client
	redisClient, err := initRedis(config.RedisURL)
	if err != nil {
		logger.Fatal("Failed to initialize Redis", zap.Error(err))
	}

	// Initialize orchestrator
	orchestrator, err := NewAgentOrchestrator(redisClient, logger, config)
	if err != nil {
		logger.Fatal("Failed to initialize orchestrator", zap.Error(err))
	}

	// Start orchestrator
	go func() {
		logger.Info("🌐 Starting Agent Orchestrator", zap.String("port", config.Port))
		if err := orchestrator.Start(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start orchestrator", zap.Error(err))
		}
	}()

	// Start background services
	go orchestrator.startHealthMonitoring()
	go orchestrator.startMetricsCollection()
	go orchestrator.startTaskProcessor()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	logger.Info("🎯 Agent Orchestrator is running. Press Ctrl+C to stop.")
	<-c

	logger.Info("🛑 Shutting down Agent Orchestrator...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := orchestrator.Shutdown(ctx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}

	logger.Info("✅ Agent Orchestrator stopped gracefully")
}

// NewAgentOrchestrator creates a new agent orchestrator
func NewAgentOrchestrator(redisClient *redis.Client, logger *zap.Logger, config *OrchestratorConfig) (*AgentOrchestrator, error) {
	// Initialize Redis advanced clients
	searchClient := redisadvanced.NewSearchClient(redisClient, logger, nil)
	jsonClient := redisadvanced.NewJSONClient(redisClient, logger, nil)
	timeSeriesClient := redisadvanced.NewTimeSeriesClient(redisClient, logger, nil)
	streamsClient := redisadvanced.NewStreamsClient(redisClient, logger, nil)

	// Initialize event store
	eventStore := eventsourcing.NewEventStore(streamsClient, logger, nil)

	// Initialize circuit breaker
	cbConfig := &circuitbreaker.MiddlewareConfig{
		DefaultConfig: circuitbreaker.DefaultConfig("orchestrator"),
	}
	circuitBreakerMiddleware := circuitbreaker.NewHTTPMiddleware(logger, cbConfig)

	// Initialize rate limiter
	rateLimiterClient := ratelimiter.NewRedisLimiter(redisClient, logger, nil)
	rlConfig := &ratelimiter.MiddlewareConfig{
		DefaultRule: &ratelimiter.RateLimitRule{
			Limit:     1000,
			Window:    time.Minute,
			Algorithm: "sliding_window",
		},
	}
	rateLimiterMiddleware := ratelimiter.NewHTTPMiddleware(rateLimiterClient, logger, rlConfig)

	orchestrator := &AgentOrchestrator{
		redis:            redisClient,
		logger:           logger,
		searchClient:     searchClient,
		jsonClient:       jsonClient,
		timeSeriesClient: timeSeriesClient,
		streamsClient:    streamsClient,
		eventStore:       eventStore,
		circuitBreaker:   circuitBreakerMiddleware,
		rateLimiter:      rateLimiterMiddleware,
		agents:           make(map[string]*AgentInfo),
		config:           config,
	}

	// Setup routes
	orchestrator.setupRoutes()

	// Initialize agent registry index
	if err := orchestrator.initializeAgentIndex(); err != nil {
		return nil, err
	}

	return orchestrator, nil
}

// setupRoutes configures HTTP routes
func (ao *AgentOrchestrator) setupRoutes() {
	ao.router = gin.New()
	ao.router.Use(gin.Recovery())
	ao.router.Use(ao.loggingMiddleware())
	ao.router.Use(ao.circuitBreaker.Middleware())
	ao.router.Use(ao.rateLimiter.Middleware())

	// API routes
	api := ao.router.Group("/api/v1")
	{
		// Agent management
		agents := api.Group("/agents")
		{
			agents.POST("/register", ao.registerAgent)
			agents.DELETE("/:id", ao.unregisterAgent)
			agents.GET("/:id", ao.getAgent)
			agents.GET("", ao.listAgents)
			agents.POST("/:id/heartbeat", ao.agentHeartbeat)
			agents.GET("/:id/health", ao.getAgentHealth)
		}

		// Task management
		tasks := api.Group("/tasks")
		{
			tasks.POST("", ao.createTask)
			tasks.GET("/:id", ao.getTask)
			tasks.GET("", ao.listTasks)
			tasks.POST("/:id/cancel", ao.cancelTask)
		}

		// Agent discovery and search
		discovery := api.Group("/discovery")
		{
			discovery.GET("/search", ao.searchAgents)
			discovery.GET("/capabilities", ao.getCapabilities)
			discovery.POST("/match", ao.matchAgents)
		}

		// Metrics and monitoring
		monitoring := api.Group("/monitoring")
		{
			monitoring.GET("/health", ao.getSystemHealth)
			monitoring.GET("/metrics", ao.getMetrics)
			monitoring.GET("/stats", ao.getStats)
		}

		// Event sourcing
		events := api.Group("/events")
		{
			events.GET("/stream/:aggregate_type/:aggregate_id", ao.getEventStream)
			events.POST("/replay", ao.replayEvents)
		}
	}

	// WebSocket endpoint for real-time updates
	ao.router.GET("/ws", ao.handleWebSocket)
}

// Start starts the orchestrator HTTP server
func (ao *AgentOrchestrator) Start() error {
	return ao.router.Run(":" + ao.config.Port)
}

// Shutdown gracefully shuts down the orchestrator
func (ao *AgentOrchestrator) Shutdown(ctx context.Context) error {
	ao.logger.Info("Shutting down orchestrator...")

	// Unregister all agents
	for agentID := range ao.agents {
		if err := ao.unregisterAgentInternal(agentID); err != nil {
			ao.logger.Error("Failed to unregister agent during shutdown", 
				zap.String("agent_id", agentID), 
				zap.Error(err),
			)
		}
	}

	// Close Redis connection
	if err := ao.redis.Close(); err != nil {
		ao.logger.Error("Error closing Redis connection", zap.Error(err))
	}

	return nil
}

// Helper functions

func loadConfig() *OrchestratorConfig {
	return &OrchestratorConfig{
		Port:            getEnv("ORCHESTRATOR_PORT", "8095"),
		RedisURL:        getEnv("REDIS_URL", "redis://localhost:6379"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		HealthInterval:  30 * time.Second,
		MetricsInterval: 60 * time.Second,
		MaxAgents:       100,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initLogger(level string) (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	
	if level == "debug" {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	return config.Build()
}

func initRedis(redisURL string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

func (ao *AgentOrchestrator) loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		ao.logger.Info("HTTP Request",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("client_ip", param.ClientIP),
		)
		return ""
	})
}

// Agent management handlers

func (ao *AgentOrchestrator) registerAgent(c *gin.Context) {
	var agentInfo AgentInfo
	if err := c.ShouldBindJSON(&agentInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent info", "details": err.Error()})
		return
	}

	// Validate agent info
	if agentInfo.ID == "" || agentInfo.Name == "" || agentInfo.Type == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields: id, name, type"})
		return
	}

	// Check if agent already exists
	if _, exists := ao.agents[agentInfo.ID]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Agent already registered"})
		return
	}

	// Set registration time and status
	agentInfo.RegisteredAt = time.Now()
	agentInfo.LastSeen = time.Now()
	agentInfo.Status = "active"
	agentInfo.Health = &AgentHealth{
		Status:    "healthy",
		LastCheck: time.Now(),
	}

	// Store agent info
	ao.agents[agentInfo.ID] = &agentInfo

	// Store in Redis JSON
	if err := ao.jsonClient.Set(c.Request.Context(),
		fmt.Sprintf("agent:%s", agentInfo.ID),
		agentInfo,
		24*time.Hour); err != nil {
		ao.logger.Error("Failed to store agent in Redis", zap.Error(err))
	}

	// Index agent for search
	if err := ao.indexAgent(&agentInfo); err != nil {
		ao.logger.Error("Failed to index agent", zap.Error(err))
	}

	// Record registration event
	event := &eventsourcing.Event{
		AggregateID:   agentInfo.ID,
		AggregateType: "agent",
		EventType:     "agent_registered",
		EventVersion:  1,
		Data: map[string]interface{}{
			"agent_id":   agentInfo.ID,
			"agent_name": agentInfo.Name,
			"agent_type": agentInfo.Type,
		},
	}

	if err := ao.eventStore.AppendEvent(c.Request.Context(), event); err != nil {
		ao.logger.Error("Failed to record registration event", zap.Error(err))
	}

	ao.logger.Info("Agent registered successfully",
		zap.String("agent_id", agentInfo.ID),
		zap.String("agent_name", agentInfo.Name),
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Agent registered successfully",
		"agent":   agentInfo,
	})
}

func (ao *AgentOrchestrator) unregisterAgent(c *gin.Context) {
	agentID := c.Param("id")

	if err := ao.unregisterAgentInternal(agentID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agent unregistered successfully"})
}

func (ao *AgentOrchestrator) getAgent(c *gin.Context) {
	agentID := c.Param("id")

	agent, exists := ao.agents[agentID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	c.JSON(http.StatusOK, agent)
}

func (ao *AgentOrchestrator) listAgents(c *gin.Context) {
	agents := make([]*AgentInfo, 0, len(ao.agents))
	for _, agent := range ao.agents {
		agents = append(agents, agent)
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": agents,
		"count":  len(agents),
	})
}

func (ao *AgentOrchestrator) agentHeartbeat(c *gin.Context) {
	agentID := c.Param("id")

	agent, exists := ao.agents[agentID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Update last seen time
	agent.LastSeen = time.Now()
	agent.Status = "active"

	// Record heartbeat metric
	if err := ao.timeSeriesClient.RecordMetric(c.Request.Context(),
		"agent_heartbeat",
		1.0,
		map[string]string{"agent_id": agentID}); err != nil {
		ao.logger.Error("Failed to record heartbeat metric", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{"message": "Heartbeat received"})
}

func (ao *AgentOrchestrator) getAgentHealth(c *gin.Context) {
	agentID := c.Param("id")

	agent, exists := ao.agents[agentID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	c.JSON(http.StatusOK, agent.Health)
}
