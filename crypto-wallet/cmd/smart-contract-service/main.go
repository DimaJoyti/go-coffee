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
	l.Info("Starting Smart Contract Service")

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50053)) // Use a different port for gRPC
	if err != nil {
		l.Fatal(fmt.Sprintf("Failed to listen: %v", err))
	}

	// Create a gRPC server object
	grpcServer := grpc.NewServer()

	// Register services
	// TODO: Register smart contract service implementation
	// pb.RegisterSmartContractServiceServer(grpcServer, &smartContractService{})

	// Start gRPC server in a goroutine
	go func() {
		l.Info("Smart Contract Service listening on port 50053")
		if err := grpcServer.Serve(lis); err != nil {
			l.Fatal(fmt.Sprintf("Failed to serve: %v", err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	l.Info("Shutting down Smart Contract Service...")

	// Gracefully stop the gRPC server
	grpcServer.GracefulStop()
	l.Info("Smart Contract Service stopped")
}

// smartContractService implements the SmartContractService gRPC service
type smartContractService struct {
	// TODO: Add dependencies like database, blockchain client, etc.
	// pb.UnimplementedSmartContractServiceServer
}

// DeployContract deploys a new smart contract
func (s *smartContractService) DeployContract(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement contract deployment
	return nil, nil
}

// CallContract calls a smart contract method
func (s *smartContractService) CallContract(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement contract method call
	return nil, nil
}

// GetContractEvents retrieves events emitted by a smart contract
func (s *smartContractService) GetContractEvents(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement contract event retrieval
	return nil, nil
}

// GetContractABI retrieves the ABI of a smart contract
func (s *smartContractService) GetContractABI(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement contract ABI retrieval
	return nil, nil
}

// VerifyContract verifies a smart contract on the blockchain
func (s *smartContractService) VerifyContract(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement contract verification
	return nil, nil
}

// GetTokenInfo retrieves information about a token contract
func (s *smartContractService) GetTokenInfo(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: Implement token info retrieval
	return nil, nil
}
