package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
)

// Sample data structures
type User struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Age      int       `json:"age"`
	City     string    `json:"city"`
	JoinDate time.Time `json:"join_date"`
}

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Description string  `json:"description"`
}

type Order struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	ProductID  string    `json:"product_id"`
	Quantity   int       `json:"quantity"`
	Total      float64   `json:"total"`
	Status     string    `json:"status"`
	OrderDate  time.Time `json:"order_date"`
}

func main() {
	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	// Test connection
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	fmt.Printf("Connected to Redis: %s\n", pong)

	// Seed data
	fmt.Println("🌱 Seeding Redis with sample data...")

	// Clear existing data (optional)
	fmt.Println("🧹 Clearing existing data...")
	rdb.FlushDB(ctx)

	// Seed users
	seedUsers(ctx, rdb)

	// Seed products
	seedProducts(ctx, rdb)

	// Seed orders
	seedOrders(ctx, rdb)

	// Seed analytics data
	seedAnalytics(ctx, rdb)

	// Seed real-time data
	seedRealTimeData(ctx, rdb)

	// Seed search indices
	seedSearchData(ctx, rdb)

	fmt.Println("✅ Data seeding completed successfully!")
	fmt.Println("\n📊 Summary:")
	
	// Print summary
	printSummary(ctx, rdb)
}

func seedUsers(ctx context.Context, rdb *redis.Client) {
	fmt.Println("👥 Seeding users...")
	
	users := []User{
		{ID: "1", Name: "John Doe", Email: "john@example.com", Age: 30, City: "New York", JoinDate: time.Now().AddDate(0, -6, 0)},
		{ID: "2", Name: "Jane Smith", Email: "jane@example.com", Age: 25, City: "Los Angeles", JoinDate: time.Now().AddDate(0, -4, 0)},
		{ID: "3", Name: "Bob Johnson", Email: "bob@example.com", Age: 35, City: "Chicago", JoinDate: time.Now().AddDate(0, -8, 0)},
		{ID: "4", Name: "Alice Brown", Email: "alice@example.com", Age: 28, City: "Houston", JoinDate: time.Now().AddDate(0, -2, 0)},
		{ID: "5", Name: "Charlie Wilson", Email: "charlie@example.com", Age: 32, City: "Phoenix", JoinDate: time.Now().AddDate(0, -10, 0)},
	}

	for _, user := range users {
		// Store as hash
		userKey := fmt.Sprintf("user:%s", user.ID)
		rdb.HSet(ctx, userKey, map[string]interface{}{
			"id":        user.ID,
			"name":      user.Name,
			"email":     user.Email,
			"age":       user.Age,
			"city":      user.City,
			"join_date": user.JoinDate.Format(time.RFC3339),
		})

		// Store as JSON (for RedisJSON module)
		userJSON, _ := json.Marshal(user)
		rdb.Set(ctx, fmt.Sprintf("user:json:%s", user.ID), userJSON, 0)

		// Add to sets
		rdb.SAdd(ctx, "users:all", user.ID)
		rdb.SAdd(ctx, fmt.Sprintf("users:city:%s", user.City), user.ID)
		rdb.ZAdd(ctx, "users:by_age", &redis.Z{Score: float64(user.Age), Member: user.ID})

		// Set expiration for some keys (demo TTL)
		if user.ID == "5" {
			rdb.Expire(ctx, userKey, 1*time.Hour)
		}
	}
}

func seedProducts(ctx context.Context, rdb *redis.Client) {
	fmt.Println("🛍️ Seeding products...")
	
	products := []Product{
		{ID: "p1", Name: "Laptop", Category: "Electronics", Price: 999.99, Stock: 50, Description: "High-performance laptop"},
		{ID: "p2", Name: "Coffee Mug", Category: "Kitchen", Price: 15.99, Stock: 200, Description: "Ceramic coffee mug"},
		{ID: "p3", Name: "Book", Category: "Education", Price: 29.99, Stock: 100, Description: "Programming book"},
		{ID: "p4", Name: "Headphones", Category: "Electronics", Price: 199.99, Stock: 75, Description: "Wireless headphones"},
		{ID: "p5", Name: "Desk Chair", Category: "Furniture", Price: 299.99, Stock: 25, Description: "Ergonomic office chair"},
	}

	for _, product := range products {
		// Store as hash
		productKey := fmt.Sprintf("product:%s", product.ID)
		rdb.HSet(ctx, productKey, map[string]interface{}{
			"id":          product.ID,
			"name":        product.Name,
			"category":    product.Category,
			"price":       product.Price,
			"stock":       product.Stock,
			"description": product.Description,
		})

		// Add to category sets
		rdb.SAdd(ctx, fmt.Sprintf("products:category:%s", product.Category), product.ID)
		rdb.ZAdd(ctx, "products:by_price", &redis.Z{Score: product.Price, Member: product.ID})
		rdb.ZAdd(ctx, "products:by_stock", &redis.Z{Score: float64(product.Stock), Member: product.ID})
	}
}

func seedOrders(ctx context.Context, rdb *redis.Client) {
	fmt.Println("📦 Seeding orders...")
	
	statuses := []string{"pending", "processing", "shipped", "delivered", "cancelled"}
	
	for i := 1; i <= 20; i++ {
		order := Order{
			ID:        fmt.Sprintf("o%d", i),
			UserID:    fmt.Sprintf("%d", rand.Intn(5)+1),
			ProductID: fmt.Sprintf("p%d", rand.Intn(5)+1),
			Quantity:  rand.Intn(5) + 1,
			Total:     float64(rand.Intn(1000) + 10),
			Status:    statuses[rand.Intn(len(statuses))],
			OrderDate: time.Now().AddDate(0, 0, -rand.Intn(30)),
		}

		// Store as hash
		orderKey := fmt.Sprintf("order:%s", order.ID)
		rdb.HSet(ctx, orderKey, map[string]interface{}{
			"id":         order.ID,
			"user_id":    order.UserID,
			"product_id": order.ProductID,
			"quantity":   order.Quantity,
			"total":      order.Total,
			"status":     order.Status,
			"order_date": order.OrderDate.Format(time.RFC3339),
		})

		// Add to lists and sets
		rdb.LPush(ctx, fmt.Sprintf("user:%s:orders", order.UserID), order.ID)
		rdb.SAdd(ctx, fmt.Sprintf("orders:status:%s", order.Status), order.ID)
		rdb.ZAdd(ctx, "orders:by_total", &redis.Z{Score: order.Total, Member: order.ID})
	}
}

func seedAnalytics(ctx context.Context, rdb *redis.Client) {
	fmt.Println("📊 Seeding analytics data...")
	
	// Page views counter
	for i := 0; i < 1000; i++ {
		rdb.Incr(ctx, "analytics:page_views")
	}

	// Daily active users
	for i := 0; i < 7; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		users := rand.Intn(100) + 50
		rdb.Set(ctx, fmt.Sprintf("analytics:dau:%s", date), users, 0)
	}

	// Hourly metrics
	for i := 0; i < 24; i++ {
		hour := fmt.Sprintf("%02d", i)
		requests := rand.Intn(500) + 100
		rdb.Set(ctx, fmt.Sprintf("metrics:requests:hour:%s", hour), requests, 0)
	}
}

func seedRealTimeData(ctx context.Context, rdb *redis.Client) {
	fmt.Println("⚡ Seeding real-time data...")
	
	// Active sessions
	for i := 1; i <= 10; i++ {
		sessionID := fmt.Sprintf("sess_%d", i)
		rdb.Set(ctx, fmt.Sprintf("session:%s", sessionID), fmt.Sprintf("user_%d", rand.Intn(5)+1), 30*time.Minute)
		rdb.SAdd(ctx, "sessions:active", sessionID)
	}

	// Recent activities (stream)
	activities := []string{"login", "logout", "purchase", "view_product", "add_to_cart"}
	for i := 0; i < 50; i++ {
		activity := activities[rand.Intn(len(activities))]
		userID := fmt.Sprintf("%d", rand.Intn(5)+1)
		
		rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: "activities",
			Values: map[string]interface{}{
				"user_id":   userID,
				"activity":  activity,
				"timestamp": time.Now().Unix(),
				"ip":        fmt.Sprintf("192.168.1.%d", rand.Intn(255)),
			},
		})
	}

	// Cache some frequently accessed data
	popularProducts := []string{"p1", "p2", "p4"}
	for _, productID := range popularProducts {
		cacheKey := fmt.Sprintf("cache:product:%s", productID)
		productData := fmt.Sprintf(`{"id":"%s","cached_at":"%s"}`, productID, time.Now().Format(time.RFC3339))
		rdb.Set(ctx, cacheKey, productData, 10*time.Minute)
	}
}

func seedSearchData(ctx context.Context, rdb *redis.Client) {
	fmt.Println("🔍 Seeding search data...")
	
	// Search queries (for autocomplete)
	searchQueries := []string{
		"laptop", "coffee", "book", "headphones", "chair",
		"electronics", "kitchen", "furniture", "programming",
		"wireless", "ergonomic", "ceramic", "high-performance",
	}

	for _, query := range searchQueries {
		// Add to search suggestions with scores
		rdb.ZAdd(ctx, "search:suggestions", &redis.Z{
			Score:  float64(rand.Intn(100) + 1),
			Member: query,
		})
		
		// Track search frequency
		rdb.Incr(ctx, fmt.Sprintf("search:count:%s", query))
	}

	// Popular searches
	popularSearches := []string{"laptop", "coffee", "headphones"}
	for i, search := range popularSearches {
		rdb.ZAdd(ctx, "search:popular", &redis.Z{
			Score:  float64(len(popularSearches) - i),
			Member: search,
		})
	}
}

func printSummary(ctx context.Context, rdb *redis.Client) {
	// Count different data types
	keys := rdb.Keys(ctx, "*").Val()
	
	counts := map[string]int{
		"users":     0,
		"products":  0,
		"orders":    0,
		"analytics": 0,
		"sessions":  0,
		"cache":     0,
		"search":    0,
		"other":     0,
	}

	for _, key := range keys {
		switch {
		case contains(key, "user:"):
			counts["users"]++
		case contains(key, "product:"):
			counts["products"]++
		case contains(key, "order:"):
			counts["orders"]++
		case contains(key, "analytics:") || contains(key, "metrics:"):
			counts["analytics"]++
		case contains(key, "session:"):
			counts["sessions"]++
		case contains(key, "cache:"):
			counts["cache"]++
		case contains(key, "search:"):
			counts["search"]++
		default:
			counts["other"]++
		}
	}

	fmt.Printf("  • Users: %d keys\n", counts["users"])
	fmt.Printf("  • Products: %d keys\n", counts["products"])
	fmt.Printf("  • Orders: %d keys\n", counts["orders"])
	fmt.Printf("  • Analytics: %d keys\n", counts["analytics"])
	fmt.Printf("  • Sessions: %d keys\n", counts["sessions"])
	fmt.Printf("  • Cache: %d keys\n", counts["cache"])
	fmt.Printf("  • Search: %d keys\n", counts["search"])
	fmt.Printf("  • Other: %d keys\n", counts["other"])
	fmt.Printf("  • Total: %d keys\n", len(keys))

	// Show stream info
	streamInfo := rdb.XInfoStream(ctx, "activities").Val()
	fmt.Printf("  • Activities stream: %d entries\n", streamInfo.Length)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
