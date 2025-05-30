package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	log.Println("🚀 Starting Redis 8 AI Search Engine - Blazingly Fast!")
	log.Println("⚡ Powered by Vector Similarity Search & Semantic Understanding")

	// Get Redis URL from environment
	redisURL := "redis://localhost:6379"
	if url := os.Getenv("REDIS_URL"); url != "" {
		redisURL = url
	}

	serverPort := "8092"
	if port := os.Getenv("AI_SEARCH_PORT"); port != "" {
		serverPort = port
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

	log.Printf("✅ Connected to Redis 8 successfully")

	// Initialize Redis 8 AI Search Engine
	searchEngine := NewRedis8AISearchEngine(redisClient)

	// Start server in a goroutine
	go func() {
		if err := searchEngine.Start(serverPort); err != nil {
			log.Fatalf("Failed to start Redis 8 AI Search Engine: %v", err)
		}
	}()

	log.Printf("🎯 Redis 8 AI Search Engine is running on http://localhost:%s", serverPort)
	log.Println("")
	log.Println("🔥 **BLAZINGLY FAST AI SEARCH ENDPOINTS:**")
	log.Println("   🧠 Semantic Search:     POST /api/v1/ai-search/semantic")
	log.Println("   🎯 Vector Search:       POST /api/v1/ai-search/vector")
	log.Println("   🚀 Hybrid Search:       POST /api/v1/ai-search/hybrid")
	log.Println("   💡 Suggestions:         GET  /api/v1/ai-search/suggestions/{query}")
	log.Println("   📈 Trending:            GET  /api/v1/ai-search/trending")
	log.Println("   👤 Personalized:        GET  /api/v1/ai-search/personalized/{user_id}")
	log.Println("   ❤️  Health Check:        GET  /api/v1/ai-search/health")
	log.Println("   📊 Statistics:          GET  /api/v1/ai-search/stats")
	log.Println("")
	log.Println("⚡ **PERFORMANCE FEATURES:**")
	log.Println("   • Sub-millisecond vector similarity search")
	log.Println("   • Real-time AI embeddings")
	log.Println("   • Semantic understanding")
	log.Println("   • Hybrid search algorithms")
	log.Println("   • Personalized recommendations")
	log.Println("   • Redis 8 optimized indexing")
	log.Println("")
	log.Println("🎯 **EXAMPLE QUERIES:**")
	log.Println(`   curl -X POST http://localhost:8092/api/v1/ai-search/semantic \`)
	log.Println(`     -H "Content-Type: application/json" \`)
	log.Println(`     -d '{"query": "strong coffee with milk", "limit": 5}'`)
	log.Println("")
	log.Println("Press Ctrl+C to stop...")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🛑 Shutting down Redis 8 AI Search Engine...")

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}

	log.Println("✅ Redis 8 AI Search Engine stopped gracefully")
}
