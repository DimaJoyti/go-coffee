package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/config"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
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
	l.Info("Starting Transaction Service")

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50052)) // Use a different port for gRPC
	if err != nil {
		l.Fatal(fmt.Sprintf("Failed to listen: %v", err))
	}

	// Create a gRPC server object
	grpcServer := grpc.NewServer()

	// Register services
	// TODO: Register transaction service implementation
	// pb.RegisterTransactionServiceServer(grpcServer, &transactionService{})

	// Start gRPC server in a goroutine
	go func() {
		l.Info("Transaction Service listening on port 50052")
		if err := grpcServer.Serve(lis); err != nil {
			l.Fatal(fmt.Sprintf("Failed to serve: %v", err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	l.Info("Shutting down Transaction Service...")

	// Gracefully stop the gRPC server
	grpcServer.GracefulStop()
	l.Info("Transaction Service stopped")
}

// transactionService implements the TransactionService gRPC service
type transactionService struct {
	// TODO: Add dependencies like database, blockchain client, etc.
	// pb.UnimplementedTransactionServiceServer
}

// CreateTransaction creates a new transaction
func (s *transactionService) CreateTransaction(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement transaction creation
	return nil, nil
}

// GetTransaction retrieves a transaction by ID
func (s *transactionService) GetTransaction(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement transaction retrieval
	return nil, nil
}

// ListTransactions lists all transactions for a wallet
func (s *transactionService) ListTransactions(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement transaction listing
	return nil, nil
}

// GetTransactionStatus retrieves the status of a transaction
func (s *transactionService) GetTransactionStatus(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement transaction status retrieval
	return nil, nil
}

// EstimateGas estimates the gas required for a transaction
func (s *transactionService) EstimateGas(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement gas estimation
	return nil, nil
}

// GetGasPrice retrieves the current gas price
func (s *transactionService) GetGasPrice(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement gas price retrieval
	return nil, nil
}
