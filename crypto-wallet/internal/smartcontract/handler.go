package smartcontract

import (
	"context"

	pb "github.com/DimaJoyti/go-coffee/crypto-wallet/api/proto/contract"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GRPCHandler handles gRPC requests for the smart contract service
type GRPCHandler struct {
	pb.UnimplementedSmartContractServiceServer
	service *Service
	logger  *logger.Logger
}

// NewGRPCHandler creates a new gRPC handler
func NewGRPCHandler(service *Service, logger *logger.Logger) *GRPCHandler {
	return &GRPCHandler{
		service: service,
		logger:  logger.Named("smartcontract-grpc-handler"),
	}
}

// DeployContract deploys a new smart contract
func (h *GRPCHandler) DeployContract(ctx context.Context, req *pb.DeployContractRequest) (*pb.DeployContractResponse, error) {
	// Convert request
	deployReq := &models.DeployContractRequest{
		UserID:    req.UserId,
		WalletID:  req.WalletId,
		Name:      req.Name,
		Chain:     models.Chain(req.Chain),
		Type:      models.ContractType(req.Type),
		ABI:       req.Abi,
		Bytecode:  req.Bytecode,
		Arguments: req.Arguments,
		Gas:       req.Gas,
		GasPrice:  req.GasPrice,
		Passphrase: req.Passphrase,
	}

	// Call service
	resp, err := h.service.DeployContract(ctx, deployReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to deploy contract: %v", err)
	}

	// Convert response
	return &pb.DeployContractResponse{
		Contract: &pb.Contract{
			Id:        resp.Contract.ID,
			UserId:    resp.Contract.UserID,
			Name:      resp.Contract.Name,
			Address:   resp.Contract.Address,
			Chain:     resp.Contract.Chain,
			Abi:       resp.Contract.ABI,
			Bytecode:  resp.Contract.Bytecode,
			CreatedAt: timestamppb.New(resp.Contract.CreatedAt),
			UpdatedAt: timestamppb.New(resp.Contract.UpdatedAt),
		},
		Transaction: &pb.Transaction{
			Id:       resp.Transaction.ID,
			Hash:     resp.Transaction.Hash,
			From:     resp.Transaction.From,
			To:       resp.Transaction.To,
			Value:    resp.Transaction.Value,
			Gas:      resp.Transaction.Gas,
			GasPrice: resp.Transaction.GasPrice,
			Status:   resp.Transaction.Status,
		},
	}, nil
}

// ImportContract imports an existing contract
func (h *GRPCHandler) ImportContract(ctx context.Context, req *pb.ImportContractRequest) (*pb.ImportContractResponse, error) {
	// Convert request
	importReq := &models.ImportContractRequest{
		UserID:  req.UserId,
		Name:    req.Name,
		Address: req.Address,
		Chain:   models.Chain(req.Chain),
		Type:    models.ContractType(req.Type),
		ABI:     req.Abi,
	}

	// Call service
	resp, err := h.service.ImportContract(ctx, importReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to import contract: %v", err)
	}

	// Convert response
	return &pb.ImportContractResponse{
		Contract: &pb.Contract{
			Id:        resp.Contract.ID,
			UserId:    resp.Contract.UserID,
			Name:      resp.Contract.Name,
			Address:   resp.Contract.Address,
			Chain:     resp.Contract.Chain,
			Abi:       resp.Contract.ABI,
			Bytecode:  resp.Contract.Bytecode,
			CreatedAt: timestamppb.New(resp.Contract.CreatedAt),
			UpdatedAt: timestamppb.New(resp.Contract.UpdatedAt),
		},
	}, nil
}

// GetContract retrieves a contract by ID
func (h *GRPCHandler) GetContract(ctx context.Context, req *pb.GetContractRequest) (*pb.GetContractResponse, error) {
	// Convert request
	getReq := &models.GetContractRequest{
		ID: req.Id,
	}

	// Call service
	resp, err := h.service.GetContract(ctx, getReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.NotFound, "contract not found: %v", err)
	}

	// Convert response
	return &pb.GetContractResponse{
		Contract: &pb.Contract{
			Id:        resp.Contract.ID,
			UserId:    resp.Contract.UserID,
			Name:      resp.Contract.Name,
			Address:   resp.Contract.Address,
			Chain:     resp.Contract.Chain,
			Abi:       resp.Contract.ABI,
			Bytecode:  resp.Contract.Bytecode,
			CreatedAt: timestamppb.New(resp.Contract.CreatedAt),
			UpdatedAt: timestamppb.New(resp.Contract.UpdatedAt),
		},
	}, nil
}

// GetContractByAddress retrieves a contract by address
func (h *GRPCHandler) GetContractByAddress(ctx context.Context, req *pb.GetContractByAddressRequest) (*pb.GetContractByAddressResponse, error) {
	// Convert request
	getReq := &models.GetContractByAddressRequest{
		Address: req.Address,
		Chain:   models.Chain(req.Chain),
	}

	// Call service
	resp, err := h.service.GetContractByAddress(ctx, getReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.NotFound, "contract not found: %v", err)
	}

	// Convert response
	return &pb.GetContractByAddressResponse{
		Contract: &pb.Contract{
			Id:        resp.Contract.ID,
			UserId:    resp.Contract.UserID,
			Name:      resp.Contract.Name,
			Address:   resp.Contract.Address,
			Chain:     resp.Contract.Chain,
			Abi:       resp.Contract.ABI,
			Bytecode:  resp.Contract.Bytecode,
			CreatedAt: timestamppb.New(resp.Contract.CreatedAt),
			UpdatedAt: timestamppb.New(resp.Contract.UpdatedAt),
		},
	}, nil
}

// ListContracts lists all contracts for a user
func (h *GRPCHandler) ListContracts(ctx context.Context, req *pb.ListContractsRequest) (*pb.ListContractsResponse, error) {
	// Convert request
	listReq := &models.ListContractsRequest{
		UserID: req.UserId,
		Chain:  models.Chain(req.Chain),
		Type:   models.ContractType(req.Type),
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	}

	// Call service
	resp, err := h.service.ListContracts(ctx, listReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to list contracts: %v", err)
	}

	// Convert response
	contracts := make([]*pb.Contract, len(resp.Contracts))
	for i, contract := range resp.Contracts {
		contracts[i] = &pb.Contract{
			Id:        contract.ID,
			UserId:    contract.UserID,
			Name:      contract.Name,
			Address:   contract.Address,
			Chain:     contract.Chain,
			Abi:       contract.ABI,
			Bytecode:  contract.Bytecode,
			CreatedAt: timestamppb.New(contract.CreatedAt),
			UpdatedAt: timestamppb.New(contract.UpdatedAt),
		}
	}

	return &pb.ListContractsResponse{
		Contracts: contracts,
		Total:     int32(resp.Total),
	}, nil
}

// CallContract calls a contract method (read-only)
func (h *GRPCHandler) CallContract(ctx context.Context, req *pb.CallContractRequest) (*pb.CallContractResponse, error) {
	// This is a placeholder implementation
	// In a real implementation, you would call the contract method using the blockchain client
	// and return the result

	// Create a dummy result
	result, err := structpb.NewValue("Contract method call result")
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to create result: %v", err)
	}

	return &pb.CallContractResponse{
		Result: result,
	}, nil
}

// SendContractTransaction sends a contract transaction (state-changing)
func (h *GRPCHandler) SendContractTransaction(ctx context.Context, req *pb.SendContractTransactionRequest) (*pb.SendContractTransactionResponse, error) {
	// This is a placeholder implementation
	// In a real implementation, you would send a transaction to the contract using the blockchain client
	// and return the transaction details

	return &pb.SendContractTransactionResponse{
		Transaction: &pb.Transaction{
			Id:       "transaction-id",
			Hash:     "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			From:     "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			To:       req.ContractId,
			Value:    req.Value,
			Gas:      req.Gas,
			GasPrice: req.GasPrice,
			Status:   "pending",
		},
	}, nil
}

// GetContractEvents retrieves events emitted by a contract
func (h *GRPCHandler) GetContractEvents(ctx context.Context, req *pb.GetContractEventsRequest) (*pb.GetContractEventsResponse, error) {
	// This is a placeholder implementation
	// In a real implementation, you would retrieve contract events from the database
	// and return them

	return &pb.GetContractEventsResponse{
		Events: []*pb.ContractEvent{},
		Total:  0,
	}, nil
}

// GetTokenInfo retrieves information about a token contract
func (h *GRPCHandler) GetTokenInfo(ctx context.Context, req *pb.GetTokenInfoRequest) (*pb.GetTokenInfoResponse, error) {
	// This is a placeholder implementation
	// In a real implementation, you would retrieve token information from the blockchain
	// and return it

	return &pb.GetTokenInfoResponse{
		Name:        "Example Token",
		Symbol:      "EXT",
		Decimals:    18,
		TotalSupply: "1000000000000000000000000",
		Type:        "erc20",
	}, nil
}
