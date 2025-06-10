package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/defi"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/internal/wallet"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/blockchain"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/config"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/crypto"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

// SolanaIntegrationTestSuite contains integration tests for Solana functionality
type SolanaIntegrationTestSuite struct {
	suite.Suite
	ctx           context.Context
	logger        *logger.Logger
	keyManager    *crypto.KeyManager
	solanaClient  *blockchain.SolanaClient
	raydiumClient *defi.RaydiumClient
	jupiterClient *defi.JupiterClient
	walletService *wallet.Service
}

// SetupSuite runs before all tests in the suite
func (s *SolanaIntegrationTestSuite) SetupSuite() {
	// Skip if not running integration tests
	if testing.Short() {
		s.T().Skip("Skipping integration tests")
	}

	// Check if Solana cluster is specified
	cluster := os.Getenv("SOLANA_CLUSTER")
	if cluster == "" {
		cluster = "devnet"
	}

	s.ctx = context.Background()
	s.logger = logger.New("integration-test")

	// Create key manager
	s.keyManager = crypto.NewKeyManager("./test-keystore")

	// Create Solana client
	solanaConfig := config.SolanaNetworkConfig{
		Network:            cluster,
		RPCURL:             getSolanaRPCURL(cluster),
		WSURL:              getSolanaWSURL(cluster),
		Cluster:            cluster,
		Commitment:         "confirmed",
		Timeout:            "30s",
		MaxRetries:         3,
		ConfirmationBlocks: 32,
	}

	var err error
	s.solanaClient, err = blockchain.NewSolanaClient(solanaConfig, s.logger)
	s.Require().NoError(err)

	// Create DeFi clients
	s.raydiumClient, err = defi.NewRaydiumClient(solanaConfig.RPCURL, s.logger)
	s.Require().NoError(err)

	s.jupiterClient = defi.NewJupiterClient(s.logger)

	// Create mock wallet repository for testing
	mockRepo := &MockWalletRepository{}

	// Create wallet service
	s.walletService = wallet.NewService(
		mockRepo,
		nil, // ethClient
		nil, // bscClient
		nil, // polygonClient
		s.solanaClient,
		s.keyManager,
		s.logger,
		"./test-keystore",
	)
}

// TearDownSuite runs after all tests in the suite
func (s *SolanaIntegrationTestSuite) TearDownSuite() {
	if s.solanaClient != nil {
		s.solanaClient.Close()
	}
	if s.raydiumClient != nil {
		s.raydiumClient.Close()
	}
}

// TestSolanaWalletCreation tests creating a Solana wallet
func (s *SolanaIntegrationTestSuite) TestSolanaWalletCreation() {
	// Create wallet request
	req := &models.CreateWalletRequest{
		UserID: "test-user-1",
		Name:   "Test Solana Wallet",
		Chain:  models.ChainSolana,
		Type:   models.WalletTypeHD,
	}

	// Create wallet
	resp, err := s.walletService.CreateWallet(s.ctx, req)
	s.Require().NoError(err)
	s.Assert().NotNil(resp)

	// Verify wallet properties
	s.Assert().Equal(req.UserID, resp.Wallet.UserID)
	s.Assert().Equal(req.Name, resp.Wallet.Name)
	s.Assert().Equal(string(models.ChainSolana), resp.Wallet.Chain)
	s.Assert().NotEmpty(resp.Wallet.Address)
	s.Assert().NotEmpty(resp.PrivateKey)
	s.Assert().NotEmpty(resp.Mnemonic)
	s.Assert().Equal("m/44'/501'/0'/0'", resp.DerivationPath)

	// Verify address format (Solana addresses are base58)
	s.Assert().True(len(resp.Wallet.Address) >= 32)
	s.Assert().True(len(resp.Wallet.Address) <= 44)
}

// TestSolanaWalletImport tests importing a Solana wallet
func (s *SolanaIntegrationTestSuite) TestSolanaWalletImport() {
	// First create a wallet to get a valid private key
	privateKey, _, address, err := s.keyManager.GenerateSolanaKeyPair()
	s.Require().NoError(err)

	// Import wallet request
	req := &models.ImportWalletRequest{
		UserID:     "test-user-2",
		Name:       "Imported Solana Wallet",
		Chain:      models.ChainSolana,
		PrivateKey: privateKey,
	}

	// Import wallet
	resp, err := s.walletService.ImportWallet(s.ctx, req)
	s.Require().NoError(err)
	s.Assert().NotNil(resp)

	// Verify wallet properties
	s.Assert().Equal(req.UserID, resp.Wallet.UserID)
	s.Assert().Equal(req.Name, resp.Wallet.Name)
	s.Assert().Equal(string(models.ChainSolana), resp.Wallet.Chain)
	s.Assert().Equal(address, resp.Wallet.Address)
}

// TestSolanaBalanceQuery tests querying SOL balance
func (s *SolanaIntegrationTestSuite) TestSolanaBalanceQuery() {
	// Create a test wallet
	wallet := &models.Wallet{
		ID:      "test-wallet-1",
		Address: "11111111111111111111111111111112", // System program address
		Chain:   string(models.ChainSolana),
	}

	// Get balance request
	req := &models.GetBalanceRequest{
		WalletID: wallet.ID,
	}

	// Mock the repository to return our test wallet
	mockRepo := s.walletService.(*wallet.Service)
	// Note: In a real test, you'd need to properly mock the repository

	// Get balance (this might fail if the address has no funds, but should not error on format)
	balance, err := s.solanaClient.GetBalance(s.ctx, wallet.Address)
	if err != nil {
		// Should not be a format error
		s.Assert().NotContains(err.Error(), "invalid address")
		s.T().Skipf("Network error: %v", err)
	} else {
		s.Assert().True(balance.GreaterThanOrEqual(decimal.Zero))
	}
}

// TestRaydiumIntegration tests Raydium DeFi operations
func (s *SolanaIntegrationTestSuite) TestRaydiumIntegration() {
	// Get available pools
	pools, err := s.raydiumClient.GetPools(s.ctx)
	s.Require().NoError(err)
	s.Assert().NotEmpty(pools)

	// Test swap quote
	quote, err := s.raydiumClient.GetSwapQuote(s.ctx,
		"So11111111111111111111111111111111111111112",  // SOL
		"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
		decimal.NewFromFloat(1.0))
	s.Require().NoError(err)
	s.Assert().NotNil(quote)
	s.Assert().True(quote.OutputAmount.GreaterThan(decimal.Zero))

	// Test pool info
	if len(pools) > 0 {
		poolInfo, err := s.raydiumClient.GetPoolInfo(s.ctx, pools[0].ID)
		s.Require().NoError(err)
		s.Assert().NotNil(poolInfo)
		s.Assert().Equal(pools[0].ID, poolInfo.ID)
	}
}

// TestJupiterIntegration tests Jupiter aggregator operations
func (s *SolanaIntegrationTestSuite) TestJupiterIntegration() {
	// Skip if no internet connection
	if testing.Short() {
		s.T().Skip("Skipping Jupiter integration test")
	}

	// Get supported tokens
	tokens, err := s.jupiterClient.GetSupportedTokens(s.ctx)
	s.Require().NoError(err)
	s.Assert().NotEmpty(tokens)

	// Test token price
	price, err := s.jupiterClient.GetTokenPrice(s.ctx, "So11111111111111111111111111111111111111112")
	s.Require().NoError(err)
	s.Assert().True(price.GreaterThan(decimal.Zero))
}

// TestSolanaKeyManagement tests key management operations
func (s *SolanaIntegrationTestSuite) TestSolanaKeyManagement() {
	// Test key generation
	privateKey, publicKey, address, err := s.keyManager.GenerateSolanaKeyPair()
	s.Require().NoError(err)
	s.Assert().NotEmpty(privateKey)
	s.Assert().NotEmpty(publicKey)
	s.Assert().NotEmpty(address)
	s.Assert().Equal(publicKey, address) // In Solana, address equals public key

	// Test key import
	importedAddress, err := s.keyManager.ImportSolanaPrivateKey(privateKey)
	s.Require().NoError(err)
	s.Assert().Equal(address, importedAddress)

	// Test mnemonic generation and conversion
	mnemonic, err := s.keyManager.GenerateMnemonic()
	s.Require().NoError(err)
	s.Assert().NotEmpty(mnemonic)

	mnemonicPrivateKey, err := s.keyManager.SolanaMnemonicToPrivateKey(mnemonic, "m/44'/501'/0'/0'")
	s.Require().NoError(err)
	s.Assert().NotEmpty(mnemonicPrivateKey)
}

// TestSolanaNetworkOperations tests network-level operations
func (s *SolanaIntegrationTestSuite) TestSolanaNetworkOperations() {
	// Test getting recent blockhash
	blockhash, err := s.solanaClient.GetRecentBlockhash(s.ctx)
	if err != nil {
		s.T().Skipf("Network error: %v", err)
	} else {
		s.Assert().NotEmpty(blockhash.String())
	}

	// Test getting minimum balance for rent exemption
	minBalance, err := s.solanaClient.GetMinimumBalanceForRentExemption(s.ctx, 165)
	if err != nil {
		s.T().Skipf("Network error: %v", err)
	} else {
		s.Assert().Greater(minBalance, uint64(0))
	}
}

// TestSolanaErrorHandling tests error handling scenarios
func (s *SolanaIntegrationTestSuite) TestSolanaErrorHandling() {
	// Test invalid address
	_, err := s.solanaClient.GetBalance(s.ctx, "invalid-address")
	s.Assert().Error(err)
	s.Assert().Contains(err.Error(), "invalid address")

	// Test invalid private key import
	_, err = s.keyManager.ImportSolanaPrivateKey("invalid-key")
	s.Assert().Error(err)

	// Test context cancellation
	ctx, cancel := context.WithTimeout(s.ctx, 1*time.Millisecond)
	defer cancel()

	_, err = s.solanaClient.GetBalance(ctx, "11111111111111111111111111111112")
	s.Assert().Error(err)
}

// Helper functions

func getSolanaRPCURL(cluster string) string {
	switch cluster {
	case "mainnet-beta":
		return "https://api.mainnet-beta.solana.com"
	case "testnet":
		return "https://api.testnet.solana.com"
	default:
		return "https://api.devnet.solana.com"
	}
}

func getSolanaWSURL(cluster string) string {
	switch cluster {
	case "mainnet-beta":
		return "wss://api.mainnet-beta.solana.com"
	case "testnet":
		return "wss://api.testnet.solana.com"
	default:
		return "wss://api.devnet.solana.com"
	}
}

// MockWalletRepository is a mock implementation for testing
type MockWalletRepository struct{}

func (m *MockWalletRepository) CreateWallet(ctx context.Context, wallet *models.Wallet) error {
	return nil
}

func (m *MockWalletRepository) GetWallet(ctx context.Context, id string) (*models.Wallet, error) {
	return &models.Wallet{
		ID:      id,
		Address: "11111111111111111111111111111112",
		Chain:   string(models.ChainSolana),
	}, nil
}

func (m *MockWalletRepository) ListWallets(ctx context.Context, userID, chain, walletType string, limit, offset int) ([]*models.Wallet, int, error) {
	return []*models.Wallet{}, 0, nil
}

func (m *MockWalletRepository) UpdateWallet(ctx context.Context, wallet *models.Wallet) error {
	return nil
}

func (m *MockWalletRepository) DeleteWallet(ctx context.Context, id string) error {
	return nil
}

func (m *MockWalletRepository) SaveKeystore(ctx context.Context, walletID, keystore string) error {
	return nil
}

func (m *MockWalletRepository) GetKeystore(ctx context.Context, walletID string) (string, error) {
	return "mock-keystore", nil
}

func (m *MockWalletRepository) DeleteKeystore(ctx context.Context, walletID string) error {
	return nil
}

// TestSolanaIntegration runs the integration test suite
func TestSolanaIntegration(t *testing.T) {
	suite.Run(t, new(SolanaIntegrationTestSuite))
}
