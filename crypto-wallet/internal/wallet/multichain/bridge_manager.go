package multichain

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// BridgeManager manages cross-chain bridge operations
type BridgeManager struct {
	logger *logger.Logger
	config BridgeConfig

	// Bridge protocols
	bridges map[string]*BridgeProtocol

	// State management
	isRunning bool
	routes    map[string][]*BridgeRoute // chain_pair -> routes
}

// BridgeProtocol represents a bridge protocol implementation
type BridgeProtocol struct {
	Name     string
	Config   BridgeProtocolConfig
	Client   BridgeClient
	IsActive bool
}

// BridgeClient interface for bridge protocol clients
type BridgeClient interface {
	GetQuote(ctx context.Context, req *BridgeQuoteRequest) (*BridgeQuote, error)
	ExecuteTransfer(ctx context.Context, req *BridgeTransferRequest) (*CrossChainTransaction, error)
	GetTransactionStatus(ctx context.Context, txID string) (*TransactionStatus, error)
	GetSupportedRoutes() []*BridgeRoute
}

// BridgeQuoteRequest represents a bridge quote request
type BridgeQuoteRequest struct {
	SourceChain string          `json:"source_chain"`
	DestChain   string          `json:"dest_chain"`
	Token       TokenConfig     `json:"token"`
	Amount      decimal.Decimal `json:"amount"`
	Slippage    decimal.Decimal `json:"slippage"`
}

// BridgeQuote represents a bridge quote
type BridgeQuote struct {
	ID            string          `json:"id"`
	Protocol      string          `json:"protocol"`
	SourceChain   string          `json:"source_chain"`
	DestChain     string          `json:"dest_chain"`
	Token         TokenConfig     `json:"token"`
	AmountIn      decimal.Decimal `json:"amount_in"`
	AmountOut     decimal.Decimal `json:"amount_out"`
	Fee           decimal.Decimal `json:"fee"`
	EstimatedTime time.Duration   `json:"estimated_time"`
	ExpiresAt     time.Time       `json:"expires_at"`
	Route         *BridgeRoute    `json:"route"`
}

// BridgeTransferRequest represents a bridge transfer request
type BridgeTransferRequest struct {
	Quote       *BridgeQuote `json:"quote"`
	FromAddress string       `json:"from_address"`
	ToAddress   string       `json:"to_address"`
	PrivateKey  string       `json:"private_key"`
}

// TransactionStatus represents transaction status
type TransactionStatus struct {
	ID           string    `json:"id"`
	Status       string    `json:"status"`
	Progress     int       `json:"progress"`
	UpdatedAt    time.Time `json:"updated_at"`
	SourceTxHash string    `json:"source_tx_hash"`
	DestTxHash   string    `json:"dest_tx_hash"`
}

// NewBridgeManager creates a new bridge manager
func NewBridgeManager(logger *logger.Logger, config BridgeConfig) *BridgeManager {
	manager := &BridgeManager{
		logger:  logger.Named("bridge-manager"),
		config:  config,
		bridges: make(map[string]*BridgeProtocol),
		routes:  make(map[string][]*BridgeRoute),
	}

	// Initialize bridge protocols
	for name, bridgeConfig := range config.BridgeConfigs {
		if bridgeConfig.Enabled {
			protocol := &BridgeProtocol{
				Name:     name,
				Config:   bridgeConfig,
				Client:   manager.createBridgeClient(name, bridgeConfig),
				IsActive: true,
			}
			manager.bridges[name] = protocol
		}
	}

	return manager
}

// Start starts the bridge manager
func (bm *BridgeManager) Start(ctx context.Context) error {
	if bm.isRunning {
		return fmt.Errorf("bridge manager is already running")
	}

	if !bm.config.Enabled {
		bm.logger.Info("Bridge manager is disabled")
		return nil
	}

	bm.logger.Info("Starting bridge manager",
		zap.Strings("supported_bridges", bm.config.SupportedBridges),
		zap.String("default_bridge", bm.config.DefaultBridge))

	// Load routes from all bridges
	bm.loadRoutes()

	bm.isRunning = true
	bm.logger.Info("Bridge manager started successfully")
	return nil
}

// Stop stops the bridge manager
func (bm *BridgeManager) Stop() error {
	if !bm.isRunning {
		return nil
	}

	bm.logger.Info("Stopping bridge manager")
	bm.isRunning = false
	bm.logger.Info("Bridge manager stopped")
	return nil
}

// GetOptimalBridge finds the optimal bridge for a transfer
func (bm *BridgeManager) GetOptimalBridge(sourceChain, destChain string, token TokenConfig) (*BridgeRoute, error) {
	bm.logger.Debug("Finding optimal bridge",
		zap.String("source_chain", sourceChain),
		zap.String("dest_chain", destChain),
		zap.String("token", token.Symbol))

	chainPair := fmt.Sprintf("%s_%s", sourceChain, destChain)
	routes, exists := bm.routes[chainPair]
	if !exists || len(routes) == 0 {
		return nil, fmt.Errorf("no bridge routes available for %s -> %s", sourceChain, destChain)
	}

	// Filter routes by token
	var validRoutes []*BridgeRoute
	for _, route := range routes {
		if route.Token.Symbol == token.Symbol && route.Enabled {
			validRoutes = append(validRoutes, route)
		}
	}

	if len(validRoutes) == 0 {
		return nil, fmt.Errorf("no bridge routes available for token %s", token.Symbol)
	}

	// Sort by score (fee, time, success rate)
	sort.Slice(validRoutes, func(i, j int) bool {
		scoreI := bm.calculateRouteScore(validRoutes[i])
		scoreJ := bm.calculateRouteScore(validRoutes[j])
		return scoreI.GreaterThan(scoreJ)
	})

	bestRoute := validRoutes[0]
	bm.logger.Info("Selected optimal bridge",
		zap.String("protocol", bestRoute.Protocol),
		zap.String("fee", bestRoute.Fee.String()),
		zap.Duration("estimated_time", bestRoute.EstimatedTime))

	return bestRoute, nil
}

// ExecuteTransfer executes a cross-chain transfer
func (bm *BridgeManager) ExecuteTransfer(ctx context.Context, route *BridgeRoute, req *CrossChainTransferRequest) (*CrossChainTransaction, error) {
	bm.logger.Info("Executing cross-chain transfer",
		zap.String("protocol", route.Protocol),
		zap.String("source_chain", req.SourceChain),
		zap.String("dest_chain", req.DestChain))

	// Get bridge protocol
	bridge, exists := bm.bridges[route.Protocol]
	if !exists {
		return nil, fmt.Errorf("bridge protocol %s not found", route.Protocol)
	}

	// Get quote
	quoteReq := &BridgeQuoteRequest{
		SourceChain: req.SourceChain,
		DestChain:   req.DestChain,
		Token:       req.Token,
		Amount:      req.Amount,
		Slippage:    req.Slippage,
	}

	quote, err := bridge.Client.GetQuote(ctx, quoteReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	// Execute transfer
	transferReq := &BridgeTransferRequest{
		Quote:       quote,
		FromAddress: req.FromAddress.Hex(),
		ToAddress:   req.ToAddress.Hex(),
		PrivateKey:  req.PrivateKey,
	}

	transaction, err := bridge.Client.ExecuteTransfer(ctx, transferReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute transfer: %w", err)
	}

	bm.logger.Info("Cross-chain transfer executed",
		zap.String("transaction_id", transaction.ID),
		zap.String("source_tx_hash", transaction.SourceTxHash))

	return transaction, nil
}

// GetTransferStatus gets the status of a cross-chain transfer
func (bm *BridgeManager) GetTransferStatus(ctx context.Context, protocol, txID string) (*TransactionStatus, error) {
	bridge, exists := bm.bridges[protocol]
	if !exists {
		return nil, fmt.Errorf("bridge protocol %s not found", protocol)
	}

	return bridge.Client.GetTransactionStatus(ctx, txID)
}

// Helper methods

// loadRoutes loads routes from all bridge protocols
func (bm *BridgeManager) loadRoutes() {
	for _, bridge := range bm.bridges {
		routes := bridge.Client.GetSupportedRoutes()
		for _, route := range routes {
			chainPair := fmt.Sprintf("%s_%s", route.SourceChain, route.DestChain)
			bm.routes[chainPair] = append(bm.routes[chainPair], route)
		}
	}

	bm.logger.Info("Loaded bridge routes",
		zap.Int("total_routes", bm.getTotalRoutes()))
}

// calculateRouteScore calculates a score for a bridge route
func (bm *BridgeManager) calculateRouteScore(route *BridgeRoute) decimal.Decimal {
	// Score based on fee (lower is better), time (faster is better), success rate (higher is better)
	feeScore := decimal.NewFromFloat(1.0).Sub(route.Fee.Div(decimal.NewFromFloat(100)))                                                      // Assume max 100% fee
	timeScore := decimal.NewFromFloat(1.0).Sub(decimal.NewFromFloat(float64(route.EstimatedTime.Minutes())).Div(decimal.NewFromFloat(1440))) // Assume max 24 hours
	successScore := route.Success

	// Weighted average
	score := feeScore.Mul(decimal.NewFromFloat(0.4)).
		Add(timeScore.Mul(decimal.NewFromFloat(0.3))).
		Add(successScore.Mul(decimal.NewFromFloat(0.3)))

	return score
}

// getTotalRoutes returns the total number of routes
func (bm *BridgeManager) getTotalRoutes() int {
	total := 0
	for _, routes := range bm.routes {
		total += len(routes)
	}
	return total
}

// createBridgeClient creates a bridge client for a protocol
func (bm *BridgeManager) createBridgeClient(name string, config BridgeProtocolConfig) BridgeClient {
	switch name {
	case "stargate":
		return NewStargateBridgeClient(bm.logger, config)
	case "hop":
		return NewHopBridgeClient(bm.logger, config)
	case "across":
		return NewAcrossBridgeClient(bm.logger, config)
	case "cbridge":
		return NewCBridgeClient(bm.logger, config)
	default:
		return NewMockBridgeClient(bm.logger, config)
	}
}

// Mock bridge client for demonstration
type MockBridgeClient struct {
	logger *logger.Logger
	config BridgeProtocolConfig
}

func NewMockBridgeClient(logger *logger.Logger, config BridgeProtocolConfig) *MockBridgeClient {
	return &MockBridgeClient{
		logger: logger.Named("mock-bridge"),
		config: config,
	}
}

func (mbc *MockBridgeClient) GetQuote(ctx context.Context, req *BridgeQuoteRequest) (*BridgeQuote, error) {
	return &BridgeQuote{
		ID:            uuid.New().String(),
		Protocol:      mbc.config.Name,
		SourceChain:   req.SourceChain,
		DestChain:     req.DestChain,
		Token:         req.Token,
		AmountIn:      req.Amount,
		AmountOut:     req.Amount.Mul(decimal.NewFromFloat(0.995)), // 0.5% fee
		Fee:           req.Amount.Mul(decimal.NewFromFloat(0.005)),
		EstimatedTime: mbc.config.EstimatedTime,
		ExpiresAt:     time.Now().Add(5 * time.Minute),
	}, nil
}

func (mbc *MockBridgeClient) ExecuteTransfer(ctx context.Context, req *BridgeTransferRequest) (*CrossChainTransaction, error) {
	return &CrossChainTransaction{
		ID:             uuid.New().String(),
		Type:           "bridge",
		Status:         "pending",
		SourceChain:    req.Quote.SourceChain,
		DestChain:      req.Quote.DestChain,
		SourceTxHash:   "0x" + uuid.New().String(),
		Token:          req.Quote.Token,
		Amount:         req.Quote.AmountIn,
		Fee:            req.Quote.Fee,
		BridgeProtocol: req.Quote.Protocol,
		EstimatedTime:  req.Quote.EstimatedTime,
		CreatedAt:      time.Now(),
		Metadata:       make(map[string]interface{}),
	}, nil
}

func (mbc *MockBridgeClient) GetTransactionStatus(ctx context.Context, txID string) (*TransactionStatus, error) {
	return &TransactionStatus{
		ID:        txID,
		Status:    "completed",
		Progress:  100,
		UpdatedAt: time.Now(),
	}, nil
}

func (mbc *MockBridgeClient) GetSupportedRoutes() []*BridgeRoute {
	return []*BridgeRoute{
		{
			ID:            "ethereum_polygon_usdc",
			Protocol:      mbc.config.Name,
			SourceChain:   "ethereum",
			DestChain:     "polygon",
			Token:         TokenConfig{Symbol: "USDC", Name: "USD Coin", Decimals: 6},
			MinAmount:     decimal.NewFromFloat(1),
			MaxAmount:     decimal.NewFromFloat(1000000),
			Fee:           decimal.NewFromFloat(0.005),
			EstimatedTime: 10 * time.Minute,
			Liquidity:     decimal.NewFromFloat(10000000),
			Success:       decimal.NewFromFloat(0.99),
			Enabled:       true,
		},
		{
			ID:            "ethereum_arbitrum_eth",
			Protocol:      mbc.config.Name,
			SourceChain:   "ethereum",
			DestChain:     "arbitrum",
			Token:         TokenConfig{Symbol: "ETH", Name: "Ethereum", Decimals: 18},
			MinAmount:     decimal.NewFromFloat(0.01),
			MaxAmount:     decimal.NewFromFloat(1000),
			Fee:           decimal.NewFromFloat(0.003),
			EstimatedTime: 15 * time.Minute,
			Liquidity:     decimal.NewFromFloat(5000),
			Success:       decimal.NewFromFloat(0.98),
			Enabled:       true,
		},
	}
}

// Placeholder implementations for other bridge clients
func NewStargateBridgeClient(logger *logger.Logger, config BridgeProtocolConfig) BridgeClient {
	return NewMockBridgeClient(logger, config)
}

func NewHopBridgeClient(logger *logger.Logger, config BridgeProtocolConfig) BridgeClient {
	return NewMockBridgeClient(logger, config)
}

func NewAcrossBridgeClient(logger *logger.Logger, config BridgeProtocolConfig) BridgeClient {
	return NewMockBridgeClient(logger, config)
}

func NewCBridgeClient(logger *logger.Logger, config BridgeProtocolConfig) BridgeClient {
	return NewMockBridgeClient(logger, config)
}
