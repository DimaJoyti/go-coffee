package multichain

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// MultichainManager manages wallets across multiple blockchain networks
type MultichainManager struct {
	logger *logger.Logger
	config MultichainConfig

	// Chain managers
	chainManagers map[string]*ChainManager
	chainMutex    sync.RWMutex

	// Unified state
	unifiedBalances map[string]*UnifiedBalance
	balanceMutex    sync.RWMutex

	// Cross-chain operations
	bridgeManager    *BridgeManager
	gasTracker       *GasTracker
	priceOracle      *PriceOracle
	portfolioTracker *PortfolioTracker

	// State management
	isRunning    bool
	lastUpdate   time.Time
	updateTicker *time.Ticker
	stopChan     chan struct{}
	mutex        sync.RWMutex
}

// MultichainConfig holds configuration for multichain wallet management
type MultichainConfig struct {
	Enabled           bool                   `json:"enabled" yaml:"enabled"`
	SupportedChains   []string               `json:"supported_chains" yaml:"supported_chains"`
	DefaultChain      string                 `json:"default_chain" yaml:"default_chain"`
	UpdateInterval    time.Duration          `json:"update_interval" yaml:"update_interval"`
	BalanceThreshold  decimal.Decimal        `json:"balance_threshold" yaml:"balance_threshold"`
	ChainConfigs      map[string]ChainConfig `json:"chain_configs" yaml:"chain_configs"`
	BridgeConfig      BridgeConfig           `json:"bridge_config" yaml:"bridge_config"`
	GasConfig         GasConfig              `json:"gas_config" yaml:"gas_config"`
	PriceOracleConfig PriceOracleConfig      `json:"price_oracle_config" yaml:"price_oracle_config"`
	PortfolioConfig   PortfolioConfig        `json:"portfolio_config" yaml:"portfolio_config"`
}

// ChainConfig holds configuration for individual chains
type ChainConfig struct {
	ChainID            int64           `json:"chain_id" yaml:"chain_id"`
	Name               string          `json:"name" yaml:"name"`
	Symbol             string          `json:"symbol" yaml:"symbol"`
	RPCEndpoints       []string        `json:"rpc_endpoints" yaml:"rpc_endpoints"`
	WSEndpoints        []string        `json:"ws_endpoints" yaml:"ws_endpoints"`
	ExplorerURL        string          `json:"explorer_url" yaml:"explorer_url"`
	NativeToken        TokenConfig     `json:"native_token" yaml:"native_token"`
	Tokens             []TokenConfig   `json:"tokens" yaml:"tokens"`
	GasMultiplier      decimal.Decimal `json:"gas_multiplier" yaml:"gas_multiplier"`
	MaxGasPrice        decimal.Decimal `json:"max_gas_price" yaml:"max_gas_price"`
	ConfirmationBlocks int             `json:"confirmation_blocks" yaml:"confirmation_blocks"`
	Enabled            bool            `json:"enabled" yaml:"enabled"`
	Priority           int             `json:"priority" yaml:"priority"`
}

// TokenConfig holds token configuration
type TokenConfig struct {
	Address     string          `json:"address" yaml:"address"`
	Symbol      string          `json:"symbol" yaml:"symbol"`
	Name        string          `json:"name" yaml:"name"`
	Decimals    int             `json:"decimals" yaml:"decimals"`
	LogoURI     string          `json:"logo_uri" yaml:"logo_uri"`
	CoingeckoID string          `json:"coingecko_id" yaml:"coingecko_id"`
	IsNative    bool            `json:"is_native" yaml:"is_native"`
	IsStable    bool            `json:"is_stable" yaml:"is_stable"`
	MinBalance  decimal.Decimal `json:"min_balance" yaml:"min_balance"`
	Enabled     bool            `json:"enabled" yaml:"enabled"`
}

// UnifiedBalance represents a unified view of balances across chains
type UnifiedBalance struct {
	Token         TokenConfig              `json:"token"`
	TotalBalance  decimal.Decimal          `json:"total_balance"`
	TotalValueUSD decimal.Decimal          `json:"total_value_usd"`
	ChainBalances map[string]*ChainBalance `json:"chain_balances"`
	LastUpdated   time.Time                `json:"last_updated"`
	PriceUSD      decimal.Decimal          `json:"price_usd"`
	Change24h     decimal.Decimal          `json:"change_24h"`
	Metadata      map[string]interface{}   `json:"metadata"`
}

// ChainBalance represents balance on a specific chain
type ChainBalance struct {
	Chain       string          `json:"chain"`
	Balance     decimal.Decimal `json:"balance"`
	ValueUSD    decimal.Decimal `json:"value_usd"`
	Address     string          `json:"address"`
	LastUpdated time.Time       `json:"last_updated"`
	BlockNumber uint64          `json:"block_number"`
	Pending     decimal.Decimal `json:"pending"`
	Locked      decimal.Decimal `json:"locked"`
	Available   decimal.Decimal `json:"available"`
}

// WalletInfo represents wallet information across chains
type WalletInfo struct {
	Address       common.Address             `json:"address"`
	Chains        []string                   `json:"chains"`
	TotalValueUSD decimal.Decimal            `json:"total_value_usd"`
	TokenCount    int                        `json:"token_count"`
	Balances      map[string]*UnifiedBalance `json:"balances"`
	Transactions  []*CrossChainTransaction   `json:"transactions"`
	LastActivity  time.Time                  `json:"last_activity"`
	CreatedAt     time.Time                  `json:"created_at"`
	Metadata      map[string]interface{}     `json:"metadata"`
}

// CrossChainTransaction represents a cross-chain transaction
type CrossChainTransaction struct {
	ID             string                 `json:"id"`
	Type           string                 `json:"type"`
	Status         string                 `json:"status"`
	SourceChain    string                 `json:"source_chain"`
	DestChain      string                 `json:"dest_chain"`
	SourceTxHash   string                 `json:"source_tx_hash"`
	DestTxHash     string                 `json:"dest_tx_hash"`
	Token          TokenConfig            `json:"token"`
	Amount         decimal.Decimal        `json:"amount"`
	Fee            decimal.Decimal        `json:"fee"`
	BridgeProtocol string                 `json:"bridge_protocol"`
	EstimatedTime  time.Duration          `json:"estimated_time"`
	ActualTime     time.Duration          `json:"actual_time"`
	CreatedAt      time.Time              `json:"created_at"`
	CompletedAt    *time.Time             `json:"completed_at"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// PortfolioSummary represents a portfolio summary across all chains
type PortfolioSummary struct {
	TotalValueUSD     decimal.Decimal            `json:"total_value_usd"`
	Change24h         decimal.Decimal            `json:"change_24h"`
	Change24hPercent  decimal.Decimal            `json:"change_24h_percent"`
	ChainDistribution map[string]decimal.Decimal `json:"chain_distribution"`
	TokenDistribution map[string]decimal.Decimal `json:"token_distribution"`
	TopTokens         []*UnifiedBalance          `json:"top_tokens"`
	RiskMetrics       *RiskMetrics               `json:"risk_metrics"`
	LastUpdated       time.Time                  `json:"last_updated"`
}

// RiskMetrics represents portfolio risk metrics
type RiskMetrics struct {
	ConcentrationRisk decimal.Decimal `json:"concentration_risk"`
	ChainRisk         decimal.Decimal `json:"chain_risk"`
	TokenRisk         decimal.Decimal `json:"token_risk"`
	LiquidityRisk     decimal.Decimal `json:"liquidity_risk"`
	OverallRisk       string          `json:"overall_risk"`
	Recommendations   []string        `json:"recommendations"`
}

// NewMultichainManager creates a new multichain wallet manager
func NewMultichainManager(logger *logger.Logger, config MultichainConfig) *MultichainManager {
	manager := &MultichainManager{
		logger:          logger.Named("multichain-manager"),
		config:          config,
		chainManagers:   make(map[string]*ChainManager),
		unifiedBalances: make(map[string]*UnifiedBalance),
		stopChan:        make(chan struct{}),
	}

	// Initialize chain managers
	for _, chainName := range config.SupportedChains {
		if chainConfig, exists := config.ChainConfigs[chainName]; exists && chainConfig.Enabled {
			chainManager := NewChainManager(logger, chainName, chainConfig)
			manager.chainManagers[chainName] = chainManager
		}
	}

	// Initialize components
	manager.bridgeManager = NewBridgeManager(logger, config.BridgeConfig)
	manager.gasTracker = NewGasTracker(logger, config.GasConfig)
	manager.priceOracle = NewPriceOracle(logger, config.PriceOracleConfig)
	manager.portfolioTracker = NewPortfolioTracker(logger, config.PortfolioConfig)

	return manager
}

// Start starts the multichain manager
func (mm *MultichainManager) Start(ctx context.Context) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if mm.isRunning {
		return fmt.Errorf("multichain manager is already running")
	}

	if !mm.config.Enabled {
		mm.logger.Info("Multichain manager is disabled")
		return nil
	}

	mm.logger.Info("Starting multichain manager",
		zap.Strings("supported_chains", mm.config.SupportedChains),
		zap.String("default_chain", mm.config.DefaultChain),
		zap.Duration("update_interval", mm.config.UpdateInterval))

	// Start chain managers
	for chainName, chainManager := range mm.chainManagers {
		if err := chainManager.Start(ctx); err != nil {
			mm.logger.Error("Failed to start chain manager",
				zap.String("chain", chainName),
				zap.Error(err))
			continue
		}
	}

	// Start components
	if err := mm.bridgeManager.Start(ctx); err != nil {
		mm.logger.Error("Failed to start bridge manager", zap.Error(err))
	}

	if err := mm.gasTracker.Start(ctx); err != nil {
		mm.logger.Error("Failed to start gas tracker", zap.Error(err))
	}

	if err := mm.priceOracle.Start(ctx); err != nil {
		mm.logger.Error("Failed to start price oracle", zap.Error(err))
	}

	if err := mm.portfolioTracker.Start(ctx); err != nil {
		mm.logger.Error("Failed to start portfolio tracker", zap.Error(err))
	}

	// Start update ticker
	mm.updateTicker = time.NewTicker(mm.config.UpdateInterval)
	go mm.updateLoop(ctx)

	mm.isRunning = true
	mm.lastUpdate = time.Now()

	mm.logger.Info("Multichain manager started successfully")
	return nil
}

// Stop stops the multichain manager
func (mm *MultichainManager) Stop() error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if !mm.isRunning {
		return nil
	}

	mm.logger.Info("Stopping multichain manager")

	// Stop update ticker
	if mm.updateTicker != nil {
		mm.updateTicker.Stop()
	}

	// Signal stop
	close(mm.stopChan)

	// Stop components
	if mm.portfolioTracker != nil {
		mm.portfolioTracker.Stop()
	}
	if mm.priceOracle != nil {
		mm.priceOracle.Stop()
	}
	if mm.gasTracker != nil {
		mm.gasTracker.Stop()
	}
	if mm.bridgeManager != nil {
		mm.bridgeManager.Stop()
	}

	// Stop chain managers
	for chainName, chainManager := range mm.chainManagers {
		if err := chainManager.Stop(); err != nil {
			mm.logger.Error("Failed to stop chain manager",
				zap.String("chain", chainName),
				zap.Error(err))
		}
	}

	mm.isRunning = false
	mm.logger.Info("Multichain manager stopped")
	return nil
}

// GetUnifiedBalances returns unified balances across all chains
func (mm *MultichainManager) GetUnifiedBalances(ctx context.Context, address common.Address) (map[string]*UnifiedBalance, error) {
	mm.logger.Debug("Getting unified balances", zap.String("address", address.Hex()))

	// Get balances from all chains
	allBalances := make(map[string]map[string]*ChainBalance)

	for chainName, chainManager := range mm.chainManagers {
		balances, err := chainManager.GetBalances(ctx, address)
		if err != nil {
			mm.logger.Warn("Failed to get balances from chain",
				zap.String("chain", chainName),
				zap.Error(err))
			continue
		}
		allBalances[chainName] = balances
	}

	// Aggregate balances by token
	unifiedBalances := make(map[string]*UnifiedBalance)

	for chainName, chainBalances := range allBalances {
		for tokenSymbol, chainBalance := range chainBalances {
			if unified, exists := unifiedBalances[tokenSymbol]; exists {
				// Add to existing unified balance
				unified.TotalBalance = unified.TotalBalance.Add(chainBalance.Balance)
				unified.TotalValueUSD = unified.TotalValueUSD.Add(chainBalance.ValueUSD)
				unified.ChainBalances[chainName] = chainBalance
				if chainBalance.LastUpdated.After(unified.LastUpdated) {
					unified.LastUpdated = chainBalance.LastUpdated
				}
			} else {
				// Create new unified balance
				token := mm.getTokenConfig(chainName, tokenSymbol)
				price := mm.priceOracle.GetPrice(token.CoingeckoID)
				change24h := mm.priceOracle.GetChange24h(token.CoingeckoID)

				unifiedBalances[tokenSymbol] = &UnifiedBalance{
					Token:         token,
					TotalBalance:  chainBalance.Balance,
					TotalValueUSD: chainBalance.ValueUSD,
					ChainBalances: map[string]*ChainBalance{
						chainName: chainBalance,
					},
					LastUpdated: chainBalance.LastUpdated,
					PriceUSD:    price,
					Change24h:   change24h,
					Metadata:    make(map[string]interface{}),
				}
			}
		}
	}

	// Update cache
	mm.balanceMutex.Lock()
	mm.unifiedBalances = unifiedBalances
	mm.balanceMutex.Unlock()

	mm.logger.Info("Retrieved unified balances",
		zap.String("address", address.Hex()),
		zap.Int("token_count", len(unifiedBalances)))

	return unifiedBalances, nil
}

// GetWalletInfo returns comprehensive wallet information
func (mm *MultichainManager) GetWalletInfo(ctx context.Context, address common.Address) (*WalletInfo, error) {
	mm.logger.Debug("Getting wallet info", zap.String("address", address.Hex()))

	// Get unified balances
	balances, err := mm.GetUnifiedBalances(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get unified balances: %w", err)
	}

	// Calculate total value
	totalValueUSD := decimal.Zero
	tokenCount := 0
	for _, balance := range balances {
		if balance.TotalBalance.GreaterThan(mm.config.BalanceThreshold) {
			totalValueUSD = totalValueUSD.Add(balance.TotalValueUSD)
			tokenCount++
		}
	}

	// Get recent transactions
	transactions, err := mm.getRecentTransactions(ctx, address, 10)
	if err != nil {
		mm.logger.Warn("Failed to get recent transactions", zap.Error(err))
		transactions = []*CrossChainTransaction{}
	}

	// Find last activity
	lastActivity := time.Time{}
	for _, tx := range transactions {
		if tx.CreatedAt.After(lastActivity) {
			lastActivity = tx.CreatedAt
		}
	}

	walletInfo := &WalletInfo{
		Address:       address,
		Chains:        mm.config.SupportedChains,
		TotalValueUSD: totalValueUSD,
		TokenCount:    tokenCount,
		Balances:      balances,
		Transactions:  transactions,
		LastActivity:  lastActivity,
		CreatedAt:     time.Now(), // Simplified
		Metadata:      make(map[string]interface{}),
	}

	return walletInfo, nil
}

// GetPortfolioSummary returns portfolio summary across all chains
func (mm *MultichainManager) GetPortfolioSummary(ctx context.Context, address common.Address) (*PortfolioSummary, error) {
	mm.logger.Debug("Getting portfolio summary", zap.String("address", address.Hex()))

	// Get unified balances
	balances, err := mm.GetUnifiedBalances(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get unified balances: %w", err)
	}

	// Calculate totals
	totalValueUSD := decimal.Zero
	change24h := decimal.Zero
	chainDistribution := make(map[string]decimal.Decimal)
	tokenDistribution := make(map[string]decimal.Decimal)

	for _, balance := range balances {
		totalValueUSD = totalValueUSD.Add(balance.TotalValueUSD)
		change24h = change24h.Add(balance.Change24h.Mul(balance.TotalValueUSD))

		// Token distribution
		tokenDistribution[balance.Token.Symbol] = balance.TotalValueUSD

		// Chain distribution
		for chainName, chainBalance := range balance.ChainBalances {
			if existing, exists := chainDistribution[chainName]; exists {
				chainDistribution[chainName] = existing.Add(chainBalance.ValueUSD)
			} else {
				chainDistribution[chainName] = chainBalance.ValueUSD
			}
		}
	}

	// Calculate percentage change
	change24hPercent := decimal.Zero
	if totalValueUSD.GreaterThan(decimal.Zero) {
		change24hPercent = change24h.Div(totalValueUSD).Mul(decimal.NewFromInt(100))
	}

	// Get top tokens (sorted by value)
	topTokens := make([]*UnifiedBalance, 0, len(balances))
	for _, balance := range balances {
		topTokens = append(topTokens, balance)
	}

	// Sort by value (simplified)
	// In production, implement proper sorting

	// Calculate risk metrics
	riskMetrics := mm.calculateRiskMetrics(balances, totalValueUSD)

	summary := &PortfolioSummary{
		TotalValueUSD:     totalValueUSD,
		Change24h:         change24h,
		Change24hPercent:  change24hPercent,
		ChainDistribution: chainDistribution,
		TokenDistribution: tokenDistribution,
		TopTokens:         topTokens[:min(len(topTokens), 10)], // Top 10
		RiskMetrics:       riskMetrics,
		LastUpdated:       time.Now(),
	}

	return summary, nil
}

// TransferCrossChain initiates a cross-chain transfer
func (mm *MultichainManager) TransferCrossChain(ctx context.Context, req *CrossChainTransferRequest) (*CrossChainTransaction, error) {
	mm.logger.Info("Initiating cross-chain transfer",
		zap.String("from_chain", req.SourceChain),
		zap.String("to_chain", req.DestChain),
		zap.String("token", req.Token.Symbol),
		zap.String("amount", req.Amount.String()))

	// Validate request
	if err := mm.validateTransferRequest(req); err != nil {
		return nil, fmt.Errorf("invalid transfer request: %w", err)
	}

	// Get optimal bridge
	bridge, err := mm.bridgeManager.GetOptimalBridge(req.SourceChain, req.DestChain, req.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to find bridge: %w", err)
	}

	// Execute transfer
	transaction, err := mm.bridgeManager.ExecuteTransfer(ctx, bridge, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute transfer: %w", err)
	}

	mm.logger.Info("Cross-chain transfer initiated",
		zap.String("transaction_id", transaction.ID),
		zap.String("bridge", transaction.BridgeProtocol))

	return transaction, nil
}

// Helper methods

// updateLoop runs the main update loop
func (mm *MultichainManager) updateLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-mm.stopChan:
			return
		case <-mm.updateTicker.C:
			mm.performUpdate()
		}
	}
}

// performUpdate performs periodic updates
func (mm *MultichainManager) performUpdate() {
	mm.mutex.Lock()
	mm.lastUpdate = time.Now()
	mm.mutex.Unlock()

	// Update health status of chain managers
	mm.updateChainHealth()

	// Cleanup expired cache entries
	mm.cleanupCache()
}

// updateChainHealth updates health status of chain managers
func (mm *MultichainManager) updateChainHealth() {
	for chainName, chainManager := range mm.chainManagers {
		// Simplified health check
		isHealthy := chainManager.isRunning
		mm.logger.Debug("Chain health check",
			zap.String("chain", chainName),
			zap.Bool("healthy", isHealthy))
	}
}

// cleanupCache cleans up expired cache entries
func (mm *MultichainManager) cleanupCache() {
	mm.balanceMutex.Lock()
	defer mm.balanceMutex.Unlock()

	now := time.Now()
	for tokenSymbol, balance := range mm.unifiedBalances {
		if now.Sub(balance.LastUpdated) > 10*time.Minute {
			delete(mm.unifiedBalances, tokenSymbol)
		}
	}
}

// getTokenConfig returns token configuration for a chain and symbol
func (mm *MultichainManager) getTokenConfig(chainName, tokenSymbol string) TokenConfig {
	if chainConfig, exists := mm.config.ChainConfigs[chainName]; exists {
		// Check native token
		if chainConfig.NativeToken.Symbol == tokenSymbol {
			return chainConfig.NativeToken
		}

		// Check configured tokens
		for _, token := range chainConfig.Tokens {
			if token.Symbol == tokenSymbol {
				return token
			}
		}
	}

	// Return default token config
	return TokenConfig{
		Symbol:   tokenSymbol,
		Name:     tokenSymbol,
		Decimals: 18,
		Enabled:  true,
	}
}

// getRecentTransactions returns recent cross-chain transactions
func (mm *MultichainManager) getRecentTransactions(ctx context.Context, address common.Address, limit int) ([]*CrossChainTransaction, error) {
	// Mock implementation - in production, fetch from transaction history
	return []*CrossChainTransaction{}, nil
}

// calculateRiskMetrics calculates portfolio risk metrics
func (mm *MultichainManager) calculateRiskMetrics(balances map[string]*UnifiedBalance, totalValue decimal.Decimal) *RiskMetrics {
	if totalValue.IsZero() {
		return &RiskMetrics{
			ConcentrationRisk: decimal.Zero,
			ChainRisk:         decimal.Zero,
			TokenRisk:         decimal.Zero,
			LiquidityRisk:     decimal.Zero,
			OverallRisk:       "low",
			Recommendations:   []string{},
		}
	}

	// Calculate concentration risk (largest token percentage)
	maxTokenPercentage := decimal.Zero
	for _, balance := range balances {
		percentage := balance.TotalValueUSD.Div(totalValue)
		if percentage.GreaterThan(maxTokenPercentage) {
			maxTokenPercentage = percentage
		}
	}

	// Calculate chain risk (chain distribution)
	chainValues := make(map[string]decimal.Decimal)
	for _, balance := range balances {
		for chainName, chainBalance := range balance.ChainBalances {
			if existing, exists := chainValues[chainName]; exists {
				chainValues[chainName] = existing.Add(chainBalance.ValueUSD)
			} else {
				chainValues[chainName] = chainBalance.ValueUSD
			}
		}
	}

	maxChainPercentage := decimal.Zero
	for _, value := range chainValues {
		percentage := value.Div(totalValue)
		if percentage.GreaterThan(maxChainPercentage) {
			maxChainPercentage = percentage
		}
	}

	// Generate recommendations
	var recommendations []string
	if maxTokenPercentage.GreaterThan(decimal.NewFromFloat(0.5)) {
		recommendations = append(recommendations, "Consider diversifying token holdings")
	}
	if maxChainPercentage.GreaterThan(decimal.NewFromFloat(0.8)) {
		recommendations = append(recommendations, "Consider diversifying across more chains")
	}

	// Determine overall risk
	overallRisk := "low"
	if maxTokenPercentage.GreaterThan(decimal.NewFromFloat(0.7)) || maxChainPercentage.GreaterThan(decimal.NewFromFloat(0.9)) {
		overallRisk = "high"
	} else if maxTokenPercentage.GreaterThan(decimal.NewFromFloat(0.4)) || maxChainPercentage.GreaterThan(decimal.NewFromFloat(0.7)) {
		overallRisk = "medium"
	}

	return &RiskMetrics{
		ConcentrationRisk: maxTokenPercentage,
		ChainRisk:         maxChainPercentage,
		TokenRisk:         maxTokenPercentage,
		LiquidityRisk:     decimal.NewFromFloat(0.1), // Simplified
		OverallRisk:       overallRisk,
		Recommendations:   recommendations,
	}
}

// validateTransferRequest validates a cross-chain transfer request
func (mm *MultichainManager) validateTransferRequest(req *CrossChainTransferRequest) error {
	if req.SourceChain == "" {
		return fmt.Errorf("source chain is required")
	}
	if req.DestChain == "" {
		return fmt.Errorf("destination chain is required")
	}
	if req.SourceChain == req.DestChain {
		return fmt.Errorf("source and destination chains cannot be the same")
	}
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount must be positive")
	}
	if req.FromAddress == (common.Address{}) {
		return fmt.Errorf("from address is required")
	}
	if req.ToAddress == (common.Address{}) {
		return fmt.Errorf("to address is required")
	}
	if req.Slippage.LessThan(decimal.Zero) || req.Slippage.GreaterThan(decimal.NewFromFloat(1)) {
		return fmt.Errorf("slippage must be between 0 and 1")
	}
	return nil
}

// GetSupportedChains returns supported chains
func (mm *MultichainManager) GetSupportedChains() []string {
	return mm.config.SupportedChains
}

// GetChainStatus returns status of all chains
func (mm *MultichainManager) GetChainStatus() map[string]*ChainStatus {
	status := make(map[string]*ChainStatus)

	for chainName, chainManager := range mm.chainManagers {
		status[chainName] = &ChainStatus{
			Chain:       chainName,
			ChainID:     mm.config.ChainConfigs[chainName].ChainID,
			BlockNumber: chainManager.lastBlock,
			IsHealthy:   chainManager.isRunning,
			LastUpdated: time.Now(),
			SyncStatus:  "synced",
		}
	}

	return status
}

// GetNetworkStats returns network statistics
func (mm *MultichainManager) GetNetworkStats() *NetworkStats {
	chainStats := mm.GetChainStatus()

	activeChains := 0
	for _, status := range chainStats {
		if status.IsHealthy {
			activeChains++
		}
	}

	return &NetworkStats{
		TotalChains:  len(mm.config.SupportedChains),
		ActiveChains: activeChains,
		ChainStats:   chainStats,
		LastUpdated:  time.Now(),
	}
}

// IsRunning returns whether the manager is running
func (mm *MultichainManager) IsRunning() bool {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	return mm.isRunning
}
