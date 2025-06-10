package transaction

import (
	"context"

	pb "github.com/DimaJoyti/go-coffee/crypto-wallet/api/proto/transaction"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GRPCHandler handles gRPC requests for the transaction service
type GRPCHandler struct {
	pb.UnimplementedTransactionServiceServer
	service *Service
	logger  *logger.Logger
}

// NewGRPCHandler creates a new gRPC handler
func NewGRPCHandler(service *Service, logger *logger.Logger) *GRPCHandler {
	return &GRPCHandler{
		service: service,
		logger:  logger.Named("transaction-grpc-handler"),
	}
}

// CreateTransaction creates a new transaction
func (h *GRPCHandler) CreateTransaction(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.CreateTransactionResponse, error) {
	// Convert request
	createReq := &models.CreateTransactionRequest{
		WalletID:   req.WalletId,
		To:         req.To,
		Value:      req.Value,
		Gas:        req.Gas,
		GasPrice:   req.GasPrice,
		Data:       req.Data,
		Nonce:      req.Nonce,
		Passphrase: req.Passphrase,
	}

	// Call service
	resp, err := h.service.CreateTransaction(ctx, createReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to create transaction: %v", err)
	}

	// Convert response
	return &pb.CreateTransactionResponse{
		Transaction: convertTransactionToProto(&resp.Transaction),
	}, nil
}

// GetTransaction retrieves a transaction by ID
func (h *GRPCHandler) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.GetTransactionResponse, error) {
	// Convert request
	getReq := &models.GetTransactionRequest{
		ID: req.Id,
	}

	// Call service
	resp, err := h.service.GetTransaction(ctx, getReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.NotFound, "transaction not found: %v", err)
	}

	// Convert response
	return &pb.GetTransactionResponse{
		Transaction: convertTransactionToProto(&resp.Transaction),
	}, nil
}

// GetTransactionByHash retrieves a transaction by hash
func (h *GRPCHandler) GetTransactionByHash(ctx context.Context, req *pb.GetTransactionByHashRequest) (*pb.GetTransactionByHashResponse, error) {
	// Convert request
	getReq := &models.GetTransactionByHashRequest{
		Hash:  req.Hash,
		Chain: models.Chain(req.Chain),
	}

	// Call service
	resp, err := h.service.GetTransactionByHash(ctx, getReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.NotFound, "transaction not found: %v", err)
	}

	// Convert response
	return &pb.GetTransactionByHashResponse{
		Transaction: convertTransactionToProto(&resp.Transaction),
	}, nil
}

// ListTransactions lists all transactions for a wallet
func (h *GRPCHandler) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	// Convert request
	listReq := &models.ListTransactionsRequest{
		UserID:   req.UserId,
		WalletID: req.WalletId,
		Status:   models.TransactionStatus(req.Status),
		Chain:    models.Chain(req.Chain),
		Limit:    int(req.Limit),
		Offset:   int(req.Offset),
	}

	// Call service
	resp, err := h.service.ListTransactions(ctx, listReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to list transactions: %v", err)
	}

	// Convert response
	transactions := make([]*pb.Transaction, len(resp.Transactions))
	for i, tx := range resp.Transactions {
		transactions[i] = convertTransactionToProto(&tx)
	}

	return &pb.ListTransactionsResponse{
		Transactions: transactions,
		Total:        int32(resp.Total),
	}, nil
}

// GetTransactionStatus retrieves the status of a transaction
func (h *GRPCHandler) GetTransactionStatus(ctx context.Context, req *pb.GetTransactionStatusRequest) (*pb.GetTransactionStatusResponse, error) {
	// Convert request
	getReq := &models.GetTransactionRequest{
		ID: req.Id,
	}

	// Call service
	resp, err := h.service.GetTransaction(ctx, getReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.NotFound, "transaction not found: %v", err)
	}

	// Convert response
	return &pb.GetTransactionStatusResponse{
		Status:       resp.Transaction.Status,
		Confirmations: resp.Transaction.Confirmations,
		BlockNumber:  resp.Transaction.BlockNumber,
	}, nil
}

// EstimateGas estimates the gas required for a transaction
func (h *GRPCHandler) EstimateGas(ctx context.Context, req *pb.EstimateGasRequest) (*pb.EstimateGasResponse, error) {
	// Convert request
	estimateReq := &models.EstimateGasRequest{
		From:  req.From,
		To:    req.To,
		Value: req.Value,
		Data:  req.Data,
		Chain: models.Chain(req.Chain),
	}

	// Call service
	resp, err := h.service.EstimateGas(ctx, estimateReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to estimate gas: %v", err)
	}

	// Convert response
	return &pb.EstimateGasResponse{
		Gas: resp.Gas,
	}, nil
}

// GetGasPrice retrieves the current gas price
func (h *GRPCHandler) GetGasPrice(ctx context.Context, req *pb.GetGasPriceRequest) (*pb.GetGasPriceResponse, error) {
	// Convert request
	gasPriceReq := &models.GetGasPriceRequest{
		Chain: models.Chain(req.Chain),
	}

	// Call service
	resp, err := h.service.GetGasPrice(ctx, gasPriceReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to get gas price: %v", err)
	}

	// Convert response
	return &pb.GetGasPriceResponse{
		GasPrice: resp.GasPrice,
		Slow:     resp.Slow,
		Average:  resp.Average,
		Fast:     resp.Fast,
	}, nil
}

// GetTransactionReceipt retrieves a transaction receipt
func (h *GRPCHandler) GetTransactionReceipt(ctx context.Context, req *pb.GetTransactionReceiptRequest) (*pb.GetTransactionReceiptResponse, error) {
	// Convert request
	receiptReq := &models.GetTransactionReceiptRequest{
		Hash:  req.Hash,
		Chain: models.Chain(req.Chain),
	}

	// Call service
	resp, err := h.service.GetTransactionReceipt(ctx, receiptReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to get transaction receipt: %v", err)
	}

	// Convert logs
	logs := make([]*pb.Log, len(resp.Logs))
	for i, log := range resp.Logs {
		logs[i] = &pb.Log{
			Address:     log.Address,
			Topics:      log.Topics,
			Data:        log.Data,
			BlockNumber: log.BlockNumber,
			TxHash:      log.TxHash,
			TxIndex:     log.TxIndex,
			BlockHash:   log.BlockHash,
			Index:       log.Index,
			Removed:     log.Removed,
		}
	}

	// Convert response
	return &pb.GetTransactionReceiptResponse{
		BlockHash:         resp.BlockHash,
		BlockNumber:       resp.BlockNumber,
		ContractAddress:   resp.ContractAddress,
		CumulativeGasUsed: resp.CumulativeGasUsed,
		From:              resp.From,
		GasUsed:           resp.GasUsed,
		Status:            resp.Status,
		To:                resp.To,
		TransactionHash:   resp.TransactionHash,
		TransactionIndex:  resp.TransactionIndex,
		Logs:              logs,
	}, nil
}

// Helper function to convert a transaction model to a protobuf transaction
func convertTransactionToProto(tx *models.Transaction) *pb.Transaction {
	return &pb.Transaction{
		Id:           tx.ID,
		UserId:       tx.UserID,
		WalletId:     tx.WalletID,
		Hash:         tx.Hash,
		From:         tx.From,
		To:           tx.To,
		Value:        tx.Value,
		Gas:          tx.Gas,
		GasPrice:     tx.GasPrice,
		Nonce:        tx.Nonce,
		Data:         tx.Data,
		Chain:        tx.Chain,
		Status:       tx.Status,
		BlockNumber:  tx.BlockNumber,
		BlockHash:    tx.BlockHash,
		Confirmations: tx.Confirmations,
		CreatedAt:    timestamppb.New(tx.CreatedAt),
		UpdatedAt:    timestamppb.New(tx.UpdatedAt),
	}
}
