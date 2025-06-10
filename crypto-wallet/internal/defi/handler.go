package defi

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCHandler handles gRPC requests for DeFi service
type GRPCHandler struct {
	service *Service
	logger  *logger.Logger
	// pb.UnimplementedDeFiServiceServer
}

// NewGRPCHandler creates a new gRPC handler
func NewGRPCHandler(service *Service, logger *logger.Logger) *GRPCHandler {
	return &GRPCHandler{
		service: service,
		logger:  logger.Named("grpc-handler"),
	}
}

// GetTokenPrice handles get token price requests
func (h *GRPCHandler) GetTokenPrice(ctx context.Context, req interface{}) (interface{}, error) {
	h.logger.Info("Handling GetTokenPrice request")

	// In a real implementation, you would:
	// 1. Parse the gRPC request
	// 2. Validate the request
	// 3. Call the service method
	// 4. Return the gRPC response

	// Mock implementation
	mockReq := &GetTokenPriceRequest{
		TokenAddress: "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // WETH
		Chain:        ChainEthereum,
	}

	resp, err := h.service.GetTokenPrice(ctx, mockReq)
	if err != nil {
		h.logger.Error("Failed to get token price", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get token price: %v", err)
	}

	return resp, nil
}

// GetSwapQuote handles get swap quote requests
func (h *GRPCHandler) GetSwapQuote(ctx context.Context, req interface{}) (interface{}, error) {
	h.logger.Info("Handling GetSwapQuote request")

	// Mock implementation
	mockReq := &GetSwapQuoteRequest{
		TokenIn:  "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // WETH
		TokenOut: "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1", // USDC
		AmountIn: decimal.NewFromFloat(1.0),
		Chain:    ChainEthereum,
	}

	resp, err := h.service.GetSwapQuote(ctx, mockReq)
	if err != nil {
		h.logger.Error("Failed to get swap quote", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get swap quote: %v", err)
	}

	return resp, nil
}

// ExecuteSwap handles execute swap requests
func (h *GRPCHandler) ExecuteSwap(ctx context.Context, req interface{}) (interface{}, error) {
	h.logger.Info("Handling ExecuteSwap request")

	// Mock implementation
	mockReq := &ExecuteSwapRequest{
		QuoteID:    "mock-quote-id",
		UserID:     "mock-user-id",
		WalletID:   "mock-wallet-id",
		Passphrase: "mock-passphrase",
	}

	resp, err := h.service.ExecuteSwap(ctx, mockReq)
	if err != nil {
		h.logger.Error("Failed to execute swap", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to execute swap: %v", err)
	}

	return resp, nil
}

// GetLiquidityPools handles get liquidity pools requests
func (h *GRPCHandler) GetLiquidityPools(ctx context.Context, req interface{}) (interface{}, error) {
	h.logger.Info("Handling GetLiquidityPools request")

	// Mock implementation
	mockReq := &GetLiquidityPoolsRequest{
		Chain:    ChainEthereum,
		Protocol: ProtocolTypeUniswap,
		Limit:    10,
		Offset:   0,
	}

	resp, err := h.service.GetLiquidityPools(ctx, mockReq)
	if err != nil {
		h.logger.Error("Failed to get liquidity pools", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get liquidity pools: %v", err)
	}

	return resp, nil
}

// AddLiquidity handles add liquidity requests
func (h *GRPCHandler) AddLiquidity(ctx context.Context, req interface{}) (interface{}, error) {
	h.logger.Info("Handling AddLiquidity request")

	// Mock implementation
	mockReq := &AddLiquidityRequest{
		UserID:     "mock-user-id",
		WalletID:   "mock-wallet-id",
		PoolID:     "mock-pool-id",
		Amount0:    decimal.NewFromFloat(1.0),
		Amount1:    decimal.NewFromFloat(2500.0),
		Passphrase: "mock-passphrase",
	}

	resp, err := h.service.AddLiquidity(ctx, mockReq)
	if err != nil {
		h.logger.Error("Failed to add liquidity", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to add liquidity: %v", err)
	}

	return resp, nil
}

// GetYieldFarms handles get yield farms requests
func (h *GRPCHandler) GetYieldFarms(ctx context.Context, req interface{}) (interface{}, error) {
	h.logger.Info("Handling GetYieldFarms request")

	farms, err := h.service.GetYieldFarms(ctx, ChainEthereum)
	if err != nil {
		h.logger.Error("Failed to get yield farms", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get yield farms: %v", err)
	}

	return farms, nil
}

// StakeTokens handles stake tokens requests
func (h *GRPCHandler) StakeTokens(ctx context.Context, req interface{}) (interface{}, error) {
	h.logger.Info("Handling StakeTokens request")

	err := h.service.StakeTokens(ctx, "mock-user-id", "mock-farm-id", decimal.NewFromFloat(100.0))
	if err != nil {
		h.logger.Error("Failed to stake tokens", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to stake tokens: %v", err)
	}

	return map[string]string{"status": "success"}, nil
}

// GetLendingPositions handles get lending positions requests
func (h *GRPCHandler) GetLendingPositions(ctx context.Context, req interface{}) (interface{}, error) {
	h.logger.Info("Handling GetLendingPositions request")

	positions, err := h.service.GetLendingPositions(ctx, "mock-user-id")
	if err != nil {
		h.logger.Error("Failed to get lending positions", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get lending positions: %v", err)
	}

	return positions, nil
}

// LendTokens handles lend tokens requests
func (h *GRPCHandler) LendTokens(ctx context.Context, req interface{}) (interface{}, error) {
	h.logger.Info("Handling LendTokens request")

	err := h.service.LendTokens(ctx, "mock-user-id", "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1", decimal.NewFromFloat(1000.0))
	if err != nil {
		h.logger.Error("Failed to lend tokens", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to lend tokens: %v", err)
	}

	return map[string]string{"status": "success"}, nil
}

// BorrowTokens handles borrow tokens requests
func (h *GRPCHandler) BorrowTokens(ctx context.Context, req interface{}) (interface{}, error) {
	h.logger.Info("Handling BorrowTokens request")

	err := h.service.BorrowTokens(ctx, "mock-user-id", "0xA0b86a33E6441b8C4505B6B8C0E4F7c3C4b5C8E1", decimal.NewFromFloat(500.0))
	if err != nil {
		h.logger.Error("Failed to borrow tokens", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to borrow tokens: %v", err)
	}

	return map[string]string{"status": "success"}, nil
}

// GetAaveAccountData handles get Aave account data requests
func (h *GRPCHandler) GetAaveAccountData(ctx context.Context, req interface{}) (interface{}, error) {
	h.logger.Info("Handling GetAaveAccountData request")

	accountData, err := h.service.aaveClient.GetUserAccountData(ctx, "0x742d35Cc6634C0532925a3b8D4C9db96590e4265")
	if err != nil {
		h.logger.Error("Failed to get Aave account data", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get Aave account data: %v", err)
	}

	return accountData, nil
}

// GetChainlinkPrice handles get Chainlink price requests
func (h *GRPCHandler) GetChainlinkPrice(ctx context.Context, req interface{}) (interface{}, error) {
	h.logger.Info("Handling GetChainlinkPrice request")

	price, err := h.service.chainlinkClient.GetTokenPrice(ctx, "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	if err != nil {
		h.logger.Error("Failed to get Chainlink price", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get Chainlink price: %v", err)
	}

	return map[string]interface{}{
		"token_address": "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
		"price":         price,
	}, nil
}

// GetOneInchTokens handles get 1inch supported tokens requests
func (h *GRPCHandler) GetOneInchTokens(ctx context.Context, req interface{}) (interface{}, error) {
	h.logger.Info("Handling GetOneInchTokens request")

	tokens, err := h.service.oneInchClient.GetSupportedTokens(ctx, ChainEthereum)
	if err != nil {
		h.logger.Error("Failed to get 1inch tokens", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get 1inch tokens: %v", err)
	}

	return tokens, nil
}

// Health check endpoint
func (h *GRPCHandler) HealthCheck(ctx context.Context, req interface{}) (interface{}, error) {
	h.logger.Info("Handling HealthCheck request")

	return map[string]interface{}{
		"status":    "healthy",
		"service":   "defi-service",
		"timestamp": time.Now().Unix(),
		"protocols": []string{
			string(ProtocolTypeUniswap),
			string(ProtocolTypeAave),
			string(ProtocolTypeChainlink),
			string(ProtocolType1inch),
		},
	}, nil
}

// Helper methods

// validateRequest validates incoming requests
func (h *GRPCHandler) validateRequest(req interface{}) error {
	// In a real implementation, you would validate the request structure
	// and required fields
	return nil
}

// logRequest logs incoming requests for debugging
func (h *GRPCHandler) logRequest(method string, req interface{}) {
	// Always log debug messages - let the logger configuration handle filtering
	reqJSON, _ := json.Marshal(req)
	h.logger.Debug("Incoming request",
		zap.String("method", method),
		zap.String("request", string(reqJSON)))
}

// logResponse logs outgoing responses for debugging
func (h *GRPCHandler) logResponse(method string, resp interface{}) {
	// Always log debug messages - let the logger configuration handle filtering
	respJSON, _ := json.Marshal(resp)
	h.logger.Debug("Outgoing response",
		zap.String("method", method),
		zap.String("response", string(respJSON)))
}
