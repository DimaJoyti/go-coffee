package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"google.golang.org/grpc"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	l := logger.NewLogger(cfg.Logging)
	l.Info("Starting Wallet Service")

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50051)) // Use a different port for gRPC
	if err != nil {
		l.Fatal(fmt.Sprintf("Failed to listen: %v", err))
	}

	// Create a gRPC server object
	grpcServer := grpc.NewServer()

	// Register services
	// TODO: Register wallet service implementation
	// pb.RegisterWalletServiceServer(grpcServer, &walletService{})

	// Start gRPC server in a goroutine
	go func() {
		l.Info("Wallet Service listening on port 50051")
		if err := grpcServer.Serve(lis); err != nil {
			l.Fatal(fmt.Sprintf("Failed to serve: %v", err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	l.Info("Shutting down Wallet Service...")

	// Gracefully stop the gRPC server
	grpcServer.GracefulStop()
	l.Info("Wallet Service stopped")
}

// walletService implements the WalletService gRPC service
type walletService struct {
	// TODO: Add dependencies like database, blockchain client, etc.
	// pb.UnimplementedWalletServiceServer
}

// CreateWallet creates a new wallet
func (s *walletService) CreateWallet(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement wallet creation
	return nil, nil
}

// GetWallet retrieves a wallet by ID
func (s *walletService) GetWallet(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement wallet retrieval
	return nil, nil
}

// ListWallets lists all wallets for a user
func (s *walletService) ListWallets(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement wallet listing
	return nil, nil
}

// GetBalance retrieves the balance of a wallet
func (s *walletService) GetBalance(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement balance retrieval
	return nil, nil
}

// ImportWallet imports an existing wallet
func (s *walletService) ImportWallet(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement wallet import
	return nil, nil
}

// ExportWallet exports a wallet (private key or keystore)
func (s *walletService) ExportWallet(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement wallet export
	return nil, nil
}
