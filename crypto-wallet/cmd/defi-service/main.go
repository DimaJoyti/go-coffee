package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/defi"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/blockchain"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/redis"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
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

	// Initialize blockchain clients
	ethClient, err := blockchain.NewEthereumClient(cfg.Blockchain.Ethereum, logger)
	if err != nil {
		logger.Fatal("Failed to create Ethereum client", zap.Error(err))
	}
	defer ethClient.Close()

	bscClient, err := blockchain.NewEthereumClient(cfg.Blockchain.BSC, logger)
	if err != nil {
		logger.Fatal("Failed to create BSC client", zap.Error(err))
	}
	defer bscClient.Close()

	polygonClient, err := blockchain.NewEthereumClient(cfg.Blockchain.Polygon, logger)
	if err != nil {
		logger.Fatal("Failed to create Polygon client", zap.Error(err))
	}
	defer polygonClient.Close()

	// Initialize Solana client
	solanaClient, err := blockchain.NewSolanaClient(cfg.Blockchain.Solana, logger)
	if err != nil {
		logger.Fatal("Failed to create Solana client", zap.Error(err))
	}
	defer solanaClient.Close()

	// Initialize Solana DeFi clients
	raydiumClient, err := defi.NewRaydiumClient(cfg.Blockchain.Solana.RPCURL, logger)
	if err != nil {
		logger.Fatal("Failed to create Raydium client", zap.Error(err))
	}
	defer raydiumClient.Close()

	jupiterClient := defi.NewJupiterClient(logger)

	// Initialize DeFi service
	_ = defi.NewService(
		ethClient,
		bscClient,
		polygonClient,
		solanaClient,
		raydiumClient,
		jupiterClient,
		redisClient,
		logger,
		cfg.DeFi,
	)

	// Initialize gRPC handler
	// handler := defi.NewGRPCHandler(defiService, logger)

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	// pb.RegisterDeFiServiceServer(grpcServer, handler)

	// Enable reflection for development
	reflection.Register(grpcServer)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Services.DeFiService.GRPCPort))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

	go func() {
		logger.Info("Starting DeFi gRPC server", zap.Int("port", cfg.Services.DeFiService.GRPCPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("Failed to serve gRPC", zap.Error(err))
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

	done := make(chan bool, 1)
	go func() {
		grpcServer.GracefulStop()
		done <- true
	}()

	select {
	case <-done:
		logger.Info("DeFi service stopped gracefully")
	case <-ctx.Done():
		logger.Warn("DeFi service shutdown timeout exceeded")
		grpcServer.Stop()
	}
}
