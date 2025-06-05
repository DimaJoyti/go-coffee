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

	"github.com/DimaJoyti/go-coffee/internal/defi"
	"github.com/DimaJoyti/go-coffee/pkg/blockchain"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create logger
	logger := logger.New("defi-service")

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal("Failed to connect to Redis: %v", err)
	}

	// Create Redis client wrapper
	redisWrapper := &RedisClientWrapper{client: redisClient}

	// Create blockchain clients (mock implementations for now)
	ethClient := blockchain.NewMockEthereumClient()
	bscClient := blockchain.NewMockEthereumClient()
	polygonClient := blockchain.NewMockEthereumClient()

	// Create DeFi service
	defiService := defi.NewService(
		ethClient,
		bscClient,
		polygonClient,
		redisWrapper,
		logger,
		cfg.Web3.DeFi,
	)

	// Start DeFi service
	if err := defiService.Start(ctx); err != nil {
		logger.Fatal("Failed to start DeFi service: %v", err)
	}

	// Create HTTP server
	router := gin.Default()
	
	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "defi-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// DeFi API endpoints
	api := router.Group("/api/v1")
	{
		// Token price endpoints
		api.POST("/tokens/price", handleGetTokenPrice(defiService))
		
		// Swap endpoints
		api.POST("/swaps/quote", handleGetSwapQuote(defiService))
		api.POST("/swaps/execute", handleExecuteSwap(defiService))
		
		// Liquidity pool endpoints
		api.GET("/pools", handleGetLiquidityPools(defiService))
		
		// Arbitrage endpoints
		api.GET("/arbitrage/opportunities", handleGetArbitrageOpportunities(defiService))
		
		// Yield farming endpoints
		api.GET("/yield/opportunities", handleGetYieldOpportunities(defiService))
		
		// Trading bot endpoints
		bots := api.Group("/bots")
		{
			bots.POST("", handleCreateTradingBot(defiService))
			bots.GET("", handleGetAllTradingBots(defiService))
			bots.GET("/:id", handleGetTradingBot(defiService))
			bots.POST("/:id/start", handleStartTradingBot(defiService))
			bots.POST("/:id/stop", handleStopTradingBot(defiService))
			bots.DELETE("/:id", handleDeleteTradingBot(defiService))
			bots.GET("/:id/performance", handleGetTradingBotPerformance(defiService))
		}
		
		// On-chain analysis endpoints
		analysis := api.Group("/analysis")
		{
			analysis.GET("/signals", handleGetMarketSignals(defiService))
			analysis.GET("/whales", handleGetWhaleActivity(defiService))
			analysis.GET("/tokens/:address", handleGetTokenAnalysis(defiService))
		}
	}

	// Start HTTP server
	port := 8093 // DeFi service port
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting DeFi service HTTP server on port %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down DeFi service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("HTTP server forced to shutdown: %v", err)
	}

	// Stop DeFi service
	defiService.Stop()

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		logger.Error("Failed to close Redis connection: %v", err)
	}

	logger.Info("DeFi service stopped")
}

// RedisClientWrapper wraps the go-redis client to implement our Redis interface
type RedisClientWrapper struct {
	client *redis.Client
}

func (r *RedisClientWrapper) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClientWrapper) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClientWrapper) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisClientWrapper) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

func (r *RedisClientWrapper) HSet(ctx context.Context, key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

func (r *RedisClientWrapper) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

func (r *RedisClientWrapper) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

func (r *RedisClientWrapper) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

func (r *RedisClientWrapper) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

func (r *RedisClientWrapper) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.RPush(ctx, key, values...).Err()
}

func (r *RedisClientWrapper) LPop(ctx context.Context, key string) (string, error) {
	return r.client.LPop(ctx, key).Result()
}

func (r *RedisClientWrapper) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

func (r *RedisClientWrapper) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.LRange(ctx, key, start, stop).Result()
}

func (r *RedisClientWrapper) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

func (r *RedisClientWrapper) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

func (r *RedisClientWrapper) SRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SRem(ctx, key, members...).Err()
}

func (r *RedisClientWrapper) ZAdd(ctx context.Context, key string, members ...interface{}) error {
	// Convert interface{} to redis.Z for ZAdd
	zMembers := make([]*redis.Z, len(members)/2)
	for i := 0; i < len(members); i += 2 {
		if i+1 < len(members) {
			score, _ := members[i].(float64)
			member := members[i+1]
			zMembers[i/2] = &redis.Z{Score: score, Member: member}
		}
	}
	return r.client.ZAdd(ctx, key, zMembers...).Err()
}

func (r *RedisClientWrapper) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(ctx, key, start, stop).Result()
}

func (r *RedisClientWrapper) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.ZRem(ctx, key, members...).Err()
}

func (r *RedisClientWrapper) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

func (r *RedisClientWrapper) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

func (r *RedisClientWrapper) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisClientWrapper) Close() error {
	return r.client.Close()
}
