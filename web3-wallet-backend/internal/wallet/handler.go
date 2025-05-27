package wallet

import (
	"context"
	"time"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/logger"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/models"
	pb "github.com/DimaJoyti/go-coffee/web3-wallet-backend/api/proto/wallet"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GRPCHandler handles gRPC requests for the wallet service
type GRPCHandler struct {
	pb.UnimplementedWalletServiceServer
	service *Service
	logger  *logger.Logger
}

// NewGRPCHandler creates a new gRPC handler
func NewGRPCHandler(service *Service, logger *logger.Logger) *GRPCHandler {
	return &GRPCHandler{
		service: service,
		logger:  logger.Named("wallet-grpc-handler"),
	}
}

// CreateWallet creates a new wallet
func (h *GRPCHandler) CreateWallet(ctx context.Context, req *pb.CreateWalletRequest) (*pb.CreateWalletResponse, error) {
	// Convert request
	createReq := &models.CreateWalletRequest{
		UserID: req.UserId,
		Name:   req.Name,
		Chain:  models.Chain(req.Chain),
		Type:   models.WalletType(req.Type),
	}

	// Call service
	resp, err := h.service.CreateWallet(ctx, createReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to create wallet: %v", err)
	}

	// Convert response
	return &pb.CreateWalletResponse{
		Wallet: &pb.Wallet{
			Id:        resp.Wallet.ID,
			UserId:    resp.Wallet.UserID,
			Name:      resp.Wallet.Name,
			Address:   resp.Wallet.Address,
			Chain:     resp.Wallet.Chain,
			Type:      resp.Wallet.Type,
			CreatedAt: timestamppb.New(resp.Wallet.CreatedAt),
			UpdatedAt: timestamppb.New(resp.Wallet.UpdatedAt),
		},
		Mnemonic:       resp.Mnemonic,
		PrivateKey:     resp.PrivateKey,
		DerivationPath: resp.DerivationPath,
	}, nil
}

// GetWallet retrieves a wallet by ID
func (h *GRPCHandler) GetWallet(ctx context.Context, req *pb.GetWalletRequest) (*pb.GetWalletResponse, error) {
	// Convert request
	getReq := &models.GetWalletRequest{
		ID: req.Id,
	}

	// Call service
	resp, err := h.service.GetWallet(ctx, getReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.NotFound, "wallet not found: %v", err)
	}

	// Convert response
	return &pb.GetWalletResponse{
		Wallet: &pb.Wallet{
			Id:        resp.Wallet.ID,
			UserId:    resp.Wallet.UserID,
			Name:      resp.Wallet.Name,
			Address:   resp.Wallet.Address,
			Chain:     resp.Wallet.Chain,
			Type:      resp.Wallet.Type,
			CreatedAt: timestamppb.New(resp.Wallet.CreatedAt),
			UpdatedAt: timestamppb.New(resp.Wallet.UpdatedAt),
		},
	}, nil
}

// ListWallets lists all wallets for a user
func (h *GRPCHandler) ListWallets(ctx context.Context, req *pb.ListWalletsRequest) (*pb.ListWalletsResponse, error) {
	// Convert request
	listReq := &models.ListWalletsRequest{
		UserID: req.UserId,
		Chain:  models.Chain(req.Chain),
		Type:   models.WalletType(req.Type),
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	}

	// Call service
	resp, err := h.service.ListWallets(ctx, listReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to list wallets: %v", err)
	}

	// Convert response
	wallets := make([]*pb.Wallet, len(resp.Wallets))
	for i, wallet := range resp.Wallets {
		wallets[i] = &pb.Wallet{
			Id:        wallet.ID,
			UserId:    wallet.UserID,
			Name:      wallet.Name,
			Address:   wallet.Address,
			Chain:     wallet.Chain,
			Type:      wallet.Type,
			CreatedAt: timestamppb.New(wallet.CreatedAt),
			UpdatedAt: timestamppb.New(wallet.UpdatedAt),
		}
	}

	return &pb.ListWalletsResponse{
		Wallets: wallets,
		Total:   int32(resp.Total),
	}, nil
}

// GetBalance retrieves the balance of a wallet
func (h *GRPCHandler) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	// Convert request
	balanceReq := &models.GetBalanceRequest{
		WalletID:     req.WalletId,
		TokenAddress: req.TokenAddress,
	}

	// Call service
	resp, err := h.service.GetBalance(ctx, balanceReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to get balance: %v", err)
	}

	// Convert response
	return &pb.GetBalanceResponse{
		Balance:      resp.Balance,
		Symbol:       resp.Symbol,
		Decimals:     int32(resp.Decimals),
		TokenAddress: resp.TokenAddress,
	}, nil
}

// ImportWallet imports an existing wallet
func (h *GRPCHandler) ImportWallet(ctx context.Context, req *pb.ImportWalletRequest) (*pb.ImportWalletResponse, error) {
	// Convert request
	importReq := &models.ImportWalletRequest{
		UserID:     req.UserId,
		Name:       req.Name,
		Chain:      models.Chain(req.Chain),
		PrivateKey: req.PrivateKey,
	}

	// Call service
	resp, err := h.service.ImportWallet(ctx, importReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to import wallet: %v", err)
	}

	// Convert response
	return &pb.ImportWalletResponse{
		Wallet: &pb.Wallet{
			Id:        resp.Wallet.ID,
			UserId:    resp.Wallet.UserID,
			Name:      resp.Wallet.Name,
			Address:   resp.Wallet.Address,
			Chain:     resp.Wallet.Chain,
			Type:      resp.Wallet.Type,
			CreatedAt: timestamppb.New(resp.Wallet.CreatedAt),
			UpdatedAt: timestamppb.New(resp.Wallet.UpdatedAt),
		},
	}, nil
}

// ExportWallet exports a wallet (private key or keystore)
func (h *GRPCHandler) ExportWallet(ctx context.Context, req *pb.ExportWalletRequest) (*pb.ExportWalletResponse, error) {
	// Convert request
	exportReq := &models.ExportWalletRequest{
		WalletID:   req.WalletId,
		Passphrase: req.Passphrase,
	}

	// Call service
	resp, err := h.service.ExportWallet(ctx, exportReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to export wallet: %v", err)
	}

	// Convert response
	return &pb.ExportWalletResponse{
		PrivateKey: resp.PrivateKey,
		Keystore:   resp.Keystore,
	}, nil
}

// DeleteWallet deletes a wallet
func (h *GRPCHandler) DeleteWallet(ctx context.Context, req *pb.DeleteWalletRequest) (*pb.DeleteWalletResponse, error) {
	// Convert request
	deleteReq := &models.DeleteWalletRequest{
		WalletID: req.WalletId,
	}

	// Call service
	resp, err := h.service.DeleteWallet(ctx, deleteReq)
	if err != nil {
		h.logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "failed to delete wallet: %v", err)
	}

	// Convert response
	return &pb.DeleteWalletResponse{
		Success: resp.Success,
	}, nil
}
