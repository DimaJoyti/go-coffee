package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Test logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	fmt.Println("🔧 Starting test server...")
	log.Println("🔧 Test server initializing...")
	
	// Set up Gin
	gin.SetMode(gin.DebugMode)
	router := gin.New()
	
	// Add middleware
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	router.Use(gin.Recovery())
	
	// Add test endpoints
	router.GET("/health", func(c *gin.Context) {
		log.Println("📍 Health endpoint called")
		c.JSON(200, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
			"service":   "test-server",
		})
	})
	
	router.GET("/test", func(c *gin.Context) {
		log.Println("📍 Test endpoint called")
		c.JSON(200, gin.H{
			"message":   "Test successful",
			"timestamp": time.Now(),
		})
	})
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8096"
	}
	
	fmt.Printf("🚀 Test server starting on port %s\n", port)
	log.Printf("🚀 Test server starting on port %s", port)
	
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	
	log.Printf("🌐 Server listening on http://localhost:%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
