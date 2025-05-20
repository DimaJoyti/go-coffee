package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pb "github.com/yourusername/web3-wallet-backend/api/proto/supply"
	"github.com/yourusername/web3-wallet-backend/internal/supply"
	"github.com/yourusername/web3-wallet-backend/pkg/config"
	"github.com/yourusername/web3-wallet-backend/pkg/kafka"
	"github.com/yourusername/web3-wallet-backend/pkg/logger"
	"github.com/yourusername/web3-wallet-backend/pkg/redis"
)

func main() {
	// Load configuration
	configPath := os.Getenv("CONFIG_FILE")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Create logger
	log := logger.NewLogger(cfg.Logging)
	defer log.Sync()

	log.Info("Starting supply service")

	// Connect to database
	db, err := connectToDatabase(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}
	defer db.Close()

	// Create Redis client
	redisClient, err := redis.NewClient(&redis.Config{
		Addresses:           []string{fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)},
		Password:            cfg.Redis.Password,
		DB:                  cfg.Redis.DB,
		PoolSize:            cfg.Redis.PoolSize,
		MinIdleConns:        10,
		DialTimeout:         5 * time.Second,
		ReadTimeout:         3 * time.Second,
		WriteTimeout:        3 * time.Second,
		PoolTimeout:         4 * time.Second,
		IdleTimeout:         10 * time.Minute,
		IdleCheckFrequency:  1 * time.Minute,
		MaxRetries:          3,
		MinRetryBackoff:     8 * time.Millisecond,
		MaxRetryBackoff:     512 * time.Millisecond,
		EnableCluster:       false,
		RouteByLatency:      false,
		RouteRandomly:       false,
		EnableReadFromReplicas: false,
	})
	if err != nil {
		log.Fatal("Failed to create Redis client", err)
	}
	defer redisClient.Close()

	// Create Kafka producer
	kafkaProducer, err := kafka.NewProducer(&kafka.Config{
		Brokers:            []string{"localhost:9092"},
		RequiredAcks:       "all",
		RetryMax:           5,
		RetryBackoff:       100 * time.Millisecond,
		Compression:        "snappy",
		BatchSize:          100,
		BatchTimeout:       1 * time.Millisecond,
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       10 * time.Second,
		DialTimeout:        30 * time.Second,
		KeepAlive:          30 * time.Second,
	})
	if err != nil {
		log.Fatal("Failed to create Kafka producer", err)
	}
	defer kafkaProducer.Close()

	// Create repository
	repo := supply.NewPostgresRepository(db, log)

	// Create service
	service := supply.NewService(repo, redisClient, kafkaProducer, log)

	// Create gRPC server
	server := supply.NewServer(service, log)

	// Create gRPC health server
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterSupplyServiceServer(grpcServer, server)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	reflection.Register(grpcServer)

	// Start gRPC server
	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatal("Failed to listen", err)
	}

	go func() {
		log.Info("Starting gRPC server", "address", ":50055")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("Failed to serve", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Stop gRPC server
	grpcServer.GracefulStop()

	log.Info("Server stopped")
}

// connectToDatabase connects to the database
func connectToDatabase(cfg config.DatabaseConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
