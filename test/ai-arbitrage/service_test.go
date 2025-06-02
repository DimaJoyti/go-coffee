package aiarbitrage_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	aiarbitrage "github.com/DimaJoyti/go-coffee/internal/ai-arbitrage"
	pb "github.com/DimaJoyti/go-coffee/api/proto"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	redismcp "github.com/DimaJoyti/go-coffee/pkg/redis-mcp"
)

func TestAIArbitrageService_CreateOpportunity(t *testing.T) {
	// Setup test configuration
	cfg := &config.Config{
		Redis: config.RedisConfig{
			URL: "redis://localhost:6379",
		},
		AI: config.AIConfig{
			GeminiAPIKey: "test-key",
			OllamaURL:    "http://localhost:11434",
		},
	}

	// Create logger
	testLogger := logger.New("test-ai-arbitrage")

	// Create Redis client (mock for testing)
	redisClient, err := redismcp.NewRedisClient(cfg.Redis)
	if err != nil {
		t.Skipf("Redis not available for testing: %v", err)
	}
	defer redisClient.Close()

	// Create AI service
	aiService, err := redismcp.NewAIService(cfg.AI, testLogger, redisClient)
	require.NoError(t, err)

	// Create arbitrage service
	service, err := aiarbitrage.NewService(redisClient, aiService, testLogger, cfg)
	require.NoError(t, err)

	// Test creating an opportunity
	ctx := context.Background()
	req := &pb.CreateOpportunityRequest{
		AssetSymbol: "BTC",
		BuyPrice:    50000.0,
		SellPrice:   50500.0,
		BuyMarket:   "exchange_a",
		SellMarket:  "exchange_b",
		Volume:      1.0,
		ExpiresAt:   timestamppb.New(time.Now().Add(time.Hour)),
	}

	resp, err := service.CreateOpportunity(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotEmpty(t, resp.Opportunity.Id)
	assert.Equal(t, "BTC", resp.Opportunity.AssetSymbol)
	assert.Equal(t, 50000.0, resp.Opportunity.BuyPrice)
	assert.Equal(t, 50500.0, resp.Opportunity.SellPrice)
	assert.Greater(t, resp.Opportunity.ProfitMargin, 0.0)
}

func TestAIArbitrageService_GetOpportunities(t *testing.T) {
	// Setup test configuration
	cfg := &config.Config{
		Redis: config.RedisConfig{
			URL: "redis://localhost:6379",
		},
		AI: config.AIConfig{
			GeminiAPIKey: "test-key",
			OllamaURL:    "http://localhost:11434",
		},
	}

	// Create logger
	testLogger := logger.New("test-ai-arbitrage")

	// Create Redis client (mock for testing)
	redisClient, err := redismcp.NewRedisClient(cfg.Redis)
	if err != nil {
		t.Skipf("Redis not available for testing: %v", err)
	}
	defer redisClient.Close()

	// Create AI service
	aiService, err := redismcp.NewAIService(cfg.AI, testLogger, redisClient)
	require.NoError(t, err)

	// Create arbitrage service
	service, err := aiarbitrage.NewService(redisClient, aiService, testLogger, cfg)
	require.NoError(t, err)

	// Test getting opportunities
	ctx := context.Background()
	req := &pb.GetOpportunitiesRequest{
		AssetSymbol:     "BTC",
		MinProfitMargin: 0.5,
		Limit:           10,
	}

	resp, err := service.GetOpportunities(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Opportunities)
}

func TestAIArbitrageService_GetMarketAnalysis(t *testing.T) {
	// Setup test configuration
	cfg := &config.Config{
		Redis: config.RedisConfig{
			URL: "redis://localhost:6379",
		},
		AI: config.AIConfig{
			GeminiAPIKey: "test-key",
			OllamaURL:    "http://localhost:11434",
		},
	}

	// Create logger
	testLogger := logger.New("test-ai-arbitrage")

	// Create Redis client (mock for testing)
	redisClient, err := redismcp.NewRedisClient(cfg.Redis)
	if err != nil {
		t.Skipf("Redis not available for testing: %v", err)
	}
	defer redisClient.Close()

	// Create AI service
	aiService, err := redismcp.NewAIService(cfg.AI, testLogger, redisClient)
	require.NoError(t, err)

	// Create arbitrage service
	service, err := aiarbitrage.NewService(redisClient, aiService, testLogger, cfg)
	require.NoError(t, err)

	// Test getting market analysis
	ctx := context.Background()
	req := &pb.GetMarketAnalysisRequest{
		AssetSymbol: "BTC",
		Markets:     []string{"exchange_a", "exchange_b"},
		Timeframe:   "1h",
	}

	resp, err := service.GetMarketAnalysis(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Analysis)
}

func TestAIArbitrageService_ServiceLifecycle(t *testing.T) {
	// Setup test configuration
	cfg := &config.Config{
		Redis: config.RedisConfig{
			URL: "redis://localhost:6379",
		},
		AI: config.AIConfig{
			GeminiAPIKey: "test-key",
			OllamaURL:    "http://localhost:11434",
		},
	}

	// Create logger
	testLogger := logger.New("test-ai-arbitrage")

	// Create Redis client (mock for testing)
	redisClient, err := redismcp.NewRedisClient(cfg.Redis)
	if err != nil {
		t.Skipf("Redis not available for testing: %v", err)
	}
	defer redisClient.Close()

	// Create AI service
	aiService, err := redismcp.NewAIService(cfg.AI, testLogger, redisClient)
	require.NoError(t, err)

	// Create arbitrage service
	service, err := aiarbitrage.NewService(redisClient, aiService, testLogger, cfg)
	require.NoError(t, err)

	// Test service start and stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start service in background
	go func() {
		err := service.Start(ctx)
		if err != nil && err != context.Canceled {
			t.Errorf("Service start failed: %v", err)
		}
	}()

	// Let it run for a short time
	time.Sleep(1 * time.Second)

	// Cancel context to stop service
	cancel()

	// Wait a bit for graceful shutdown
	time.Sleep(500 * time.Millisecond)
}

func TestAIArbitrageService_MatchParticipants(t *testing.T) {
	// Setup test configuration
	cfg := &config.Config{
		Redis: config.RedisConfig{
			URL: "redis://localhost:6379",
		},
		AI: config.AIConfig{
			GeminiAPIKey: "test-key",
			OllamaURL:    "http://localhost:11434",
		},
	}

	// Create logger
	testLogger := logger.New("test-ai-arbitrage")

	// Create Redis client (mock for testing)
	redisClient, err := redismcp.NewRedisClient(cfg.Redis)
	if err != nil {
		t.Skipf("Redis not available for testing: %v", err)
	}
	defer redisClient.Close()

	// Create AI service
	aiService, err := redismcp.NewAIService(cfg.AI, testLogger, redisClient)
	require.NoError(t, err)

	// Create arbitrage service
	service, err := aiarbitrage.NewService(redisClient, aiService, testLogger, cfg)
	require.NoError(t, err)

	// First create an opportunity
	ctx := context.Background()
	createReq := &pb.CreateOpportunityRequest{
		AssetSymbol: "ETH",
		BuyPrice:    3000.0,
		SellPrice:   3050.0,
		BuyMarket:   "exchange_a",
		SellMarket:  "exchange_b",
		Volume:      2.0,
		ExpiresAt:   timestamppb.New(time.Now().Add(time.Hour)),
	}

	createResp, err := service.CreateOpportunity(ctx, createReq)
	require.NoError(t, err)
	require.True(t, createResp.Success)

	// Test matching participants
	matchReq := &pb.MatchParticipantsRequest{
		OpportunityId:  createResp.Opportunity.Id,
		ParticipantIds: []string{"participant_1", "participant_2"},
	}

	matchResp, err := service.MatchParticipants(ctx, matchReq)
	require.NoError(t, err)
	assert.True(t, matchResp.Success)
	assert.NotNil(t, matchResp.Matches)
}
