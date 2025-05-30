package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	aiorder "github.com/DimaJoyti/go-coffee/internal/ai-order"
)

const (
	defaultPort     = "50051"
	defaultRedisURL = "redis://localhost:6379"
	serviceName     = "ai-order-service"
)

func main() {
	log.Println("🚀 Starting AI Order Service...")

	// Get configuration from environment
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = defaultPort
	}

	log.Printf("✅ Configuration loaded - Port: %s", port)

	// Initialize order service (simplified version)
	orderService := aiorder.NewSimpleService()

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register AI Order Service
	aiorder.RegisterAIOrderServiceServer(grpcServer, orderService)

	// Enable reflection for development
	reflection.Register(grpcServer)

	// Start gRPC server
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Start server in goroutine
	go func() {
		log.Printf("🌐 AI Order Service listening on port %s", port)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	log.Println("🎯 AI Order Service is running. Press Ctrl+C to stop.")
	<-c

	log.Println("🛑 Shutting down AI Order Service...")

	// Graceful shutdown
	grpcServer.GracefulStop()

	log.Println("✅ AI Order Service stopped gracefully")
}






