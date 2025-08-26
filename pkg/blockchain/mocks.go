package blockchain

import (
	"context"
	"math/big"

	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
)

// Constructor functions for blockchain clients
func NewEthereumClient(config config.EthereumConfig) (EthereumClient, error) {
	return &MockEthereumClient{}, nil
}

func NewBSCClient(config config.EthereumConfig) (EthereumClient, error) {
	return &MockEthereumClient{}, nil
}

func NewPolygonClient(config config.EthereumConfig) (EthereumClient, error) {
	return &MockEthereumClient{}, nil
}

func NewSolanaClient(config config.SolanaConfig) (SolanaClient, error) {
	return &MockSolanaClient{}, nil
}

// MockEthereumClient is a mock implementation of EthereumClient
type MockEthereumClient struct {
	mock.Mock
}

func (m *MockEthereumClient) Connect(ctx context.Context, rpcURL string) error {
	args := m.Called(ctx, rpcURL)
	return args.Error(0)
}

func (m *MockEthereumClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEthereumClient) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockEthereumClient) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEthereumClient) GetBlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	args := m.Called(ctx, number)
	return args.Get(0).(*types.Block), args.Error(1)
}

func (m *MockEthereumClient) GetTransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockEthereumClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockEthereumClient) EstimateGas(ctx context.Context, msg interface{}) (uint64, error) {
	args := m.Called(ctx, msg)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEthereumClient) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	args := m.Called(ctx, address)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEthereumClient) GetNonce(ctx context.Context, address common.Address) (uint64, error) {
	args := m.Called(ctx, address)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEthereumClient) CallContract(ctx context.Context, call interface{}, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, call, blockNumber)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEthereumClient) GetTokenBalance(ctx context.Context, tokenAddress, walletAddress common.Address) (*big.Int, error) {
	args := m.Called(ctx, tokenAddress, walletAddress)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEthereumClient) GetTokenInfo(ctx context.Context, tokenAddress common.Address) (*TokenInfo, error) {
	args := m.Called(ctx, tokenAddress)
	return args.Get(0).(*TokenInfo), args.Error(1)
}

func (m *MockEthereumClient) GetTokenPrice(ctx context.Context, tokenAddress common.Address) (decimal.Decimal, error) {
	args := m.Called(ctx, tokenAddress)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

// MockSolanaClient is a mock implementation of SolanaClient
type MockSolanaClient struct {
	mock.Mock
}

func (m *MockSolanaClient) Connect(ctx context.Context, rpcURL string) error {
	args := m.Called(ctx, rpcURL)
	return args.Error(0)
}

func (m *MockSolanaClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSolanaClient) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockSolanaClient) GetBalance(ctx context.Context, address string) (uint64, error) {
	args := m.Called(ctx, address)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockSolanaClient) GetTokenBalance(ctx context.Context, tokenMint, walletAddress string) (uint64, error) {
	args := m.Called(ctx, tokenMint, walletAddress)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockSolanaClient) SendTransaction(ctx context.Context, transaction []byte) (string, error) {
	args := m.Called(ctx, transaction)
	return args.String(0), args.Error(1)
}

func (m *MockSolanaClient) GetTransaction(ctx context.Context, signature string) (*SolanaTransaction, error) {
	args := m.Called(ctx, signature)
	return args.Get(0).(*SolanaTransaction), args.Error(1)
}

func (m *MockSolanaClient) GetTokenInfo(ctx context.Context, mintAddress string) (*SolanaTokenInfo, error) {
	args := m.Called(ctx, mintAddress)
	return args.Get(0).(*SolanaTokenInfo), args.Error(1)
}

func (m *MockSolanaClient) GetTokenPrice(ctx context.Context, mintAddress string) (decimal.Decimal, error) {
	args := m.Called(ctx, mintAddress)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockSolanaClient) CallProgram(ctx context.Context, programID string, data []byte) ([]byte, error) {
	args := m.Called(ctx, programID, data)
	return args.Get(0).([]byte), args.Error(1)
}

// MockClientFactory is a mock implementation of ClientFactory
type MockClientFactory struct {
	mock.Mock
}

func (m *MockClientFactory) CreateEthereumClient(config NetworkConfig) (EthereumClient, error) {
	args := m.Called(config)
	return args.Get(0).(EthereumClient), args.Error(1)
}

func (m *MockClientFactory) CreateSolanaClient(config NetworkConfig) (SolanaClient, error) {
	args := m.Called(config)
	return args.Get(0).(SolanaClient), args.Error(1)
}

// NewMockEthereumClient creates a new mock Ethereum client with default behaviors
func NewMockEthereumClient() *MockEthereumClient {
	client := &MockEthereumClient{}
	
	// Set up default mock behaviors
	client.On("IsConnected").Return(true).Maybe()
	client.On("GetTokenPrice", mock.Anything, mock.Anything).Return(decimal.NewFromFloat(2500.0), nil).Maybe()
	
	return client
}

// NewMockSolanaClient creates a new mock Solana client with default behaviors
func NewMockSolanaClient() *MockSolanaClient {
	client := &MockSolanaClient{}
	
	// Set up default mock behaviors
	client.On("IsConnected").Return(true).Maybe()
	client.On("GetTokenPrice", mock.Anything, mock.Anything).Return(decimal.NewFromFloat(100.0), nil).Maybe()
	
	return client
}
