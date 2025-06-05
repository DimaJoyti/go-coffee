package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/integration"
	"github.com/go-redis/redis/v8"
)

func main() {
	log.Println("🚀 Starting MCP-AI Integration - The Ultimate Coffee AI Experience!")
	log.Println("🧠 Combining Redis MCP + AI Search for Blazingly Fast Intelligence")

	// Configuration
	redisURL := "redis://localhost:6379"
	aiSearchURL := "http://localhost:8092"
	mcpServerURL := "http://localhost:8090"
	integrationPort := "8093"

	// Override with environment variables
	if url := os.Getenv("REDIS_URL"); url != "" {
		redisURL = url
	}
	if url := os.Getenv("AI_SEARCH_URL"); url != "" {
		aiSearchURL = url
	}
	if url := os.Getenv("MCP_SERVER_URL"); url != "" {
		mcpServerURL = url
	}
	if port := os.Getenv("INTEGRATION_PORT"); port != "" {
		integrationPort = port
	}

	// Initialize Redis client
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	redisClient := redis.NewClient(opt)

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Printf("✅ Connected to Redis successfully")

	// Initialize MCP-AI Integration
	integration := integration.NewMCPAIIntegration(redisClient, aiSearchURL, mcpServerURL)

	// Start server in a goroutine
	go func() {
		if err := integration.Start(integrationPort); err != nil {
			log.Fatalf("Failed to start MCP-AI Integration: %v", err)
		}
	}()

	log.Printf("🎯 MCP-AI Integration is running on http://localhost:%s", integrationPort)
	log.Println("")
	log.Println("🔥 **ULTIMATE AI COFFEE EXPERIENCE ENDPOINTS:**")
	log.Println("   🧠 Enhanced Query:       POST /api/v1/mcp-ai/query")
	log.Println("   🎯 Smart Search:         POST /api/v1/mcp-ai/smart-search")
	log.Println("   👤 Recommendations:      GET  /api/v1/mcp-ai/recommendations/{user_id}")
	log.Println("   📈 Trending:             GET  /api/v1/mcp-ai/trending")
	log.Println("   ❤️  Health Check:         GET  /api/v1/mcp-ai/health")
	log.Println("   📋 Demo & Examples:      GET  /api/v1/mcp-ai/demo")
	log.Println("")
	log.Println("🚀 **INTEGRATION FEATURES:**")
	log.Println("   • Smart Query Routing (AI vs MCP)")
	log.Println("   • Semantic + Vector + Hybrid Search")
	log.Println("   • Fallback to Traditional MCP")
	log.Println("   • AI-Powered Recommendations")
	log.Println("   • Real-time Performance Optimization")
	log.Println("   • Blazingly Fast Response Times")
	log.Println("")
	log.Println("🎯 **EXAMPLE SMART QUERIES:**")
	log.Println(`   # AI-Powered Semantic Search:`)
	log.Println(`   curl -X POST http://localhost:8093/api/v1/mcp-ai/smart-search \`)
	log.Println(`     -H "Content-Type: application/json" \`)
	log.Println(`     -d '{"query": "I want something refreshing and not too strong", "agent_id": "demo"}'`)
	log.Println("")
	log.Println(`   # Traditional MCP Query:`)
	log.Println(`   curl -X POST http://localhost:8093/api/v1/mcp-ai/smart-search \`)
	log.Println(`     -H "Content-Type: application/json" \`)
	log.Println(`     -d '{"query": "get menu for shop downtown", "agent_id": "demo"}'`)
	log.Println("")
	log.Println("Press Ctrl+C to stop...")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🛑 Shutting down MCP-AI Integration...")

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}

	log.Println("✅ MCP-AI Integration stopped gracefully")
}
