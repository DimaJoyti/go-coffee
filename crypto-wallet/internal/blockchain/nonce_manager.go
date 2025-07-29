package blockchain

import (
	"context"
	"fmt"
	"sync"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/logger"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// NonceManager manages transaction nonces for different addresses and chains
type NonceManager struct {
	logger *logger.Logger

	// Nonce tracking per chain and address
	nonces map[string]map[common.Address]uint64
	mutex  sync.RWMutex
}

// NewNonceManager creates a new nonce manager
func NewNonceManager(logger *logger.Logger) *NonceManager {
	return &NonceManager{
		logger: logger.Named("nonce-manager"),
		nonces: make(map[string]map[common.Address]uint64),
	}
}

// GetNonce returns the next nonce for an address on a chain
func (nm *NonceManager) GetNonce(ctx context.Context, chain string, address common.Address) (uint64, error) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	// Initialize chain map if it doesn't exist
	if nm.nonces[chain] == nil {
		nm.nonces[chain] = make(map[common.Address]uint64)
	}

	// Get current nonce for address
	currentNonce, exists := nm.nonces[chain][address]
	if !exists {
		// Initialize with a mock nonce (in production, would query the blockchain)
		currentNonce = nm.fetchNonceFromChain(ctx, chain, address)
		nm.nonces[chain][address] = currentNonce
	}

	// Increment and return
	nextNonce := currentNonce
	nm.nonces[chain][address] = currentNonce + 1

	nm.logger.Debug("Retrieved nonce",
		zap.String("chain", chain),
		zap.String("address", address.Hex()),
		zap.Uint64("nonce", nextNonce))

	return nextNonce, nil
}

// SetNonce sets the nonce for an address on a chain
func (nm *NonceManager) SetNonce(chain string, address common.Address, nonce uint64) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	// Initialize chain map if it doesn't exist
	if nm.nonces[chain] == nil {
		nm.nonces[chain] = make(map[common.Address]uint64)
	}

	nm.nonces[chain][address] = nonce

	nm.logger.Debug("Set nonce",
		zap.String("chain", chain),
		zap.String("address", address.Hex()),
		zap.Uint64("nonce", nonce))
}

// ResetNonce resets the nonce for an address on a chain
func (nm *NonceManager) ResetNonce(ctx context.Context, chain string, address common.Address) error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	// Fetch fresh nonce from chain
	freshNonce := nm.fetchNonceFromChain(ctx, chain, address)

	// Initialize chain map if it doesn't exist
	if nm.nonces[chain] == nil {
		nm.nonces[chain] = make(map[common.Address]uint64)
	}

	nm.nonces[chain][address] = freshNonce

	nm.logger.Info("Reset nonce",
		zap.String("chain", chain),
		zap.String("address", address.Hex()),
		zap.Uint64("fresh_nonce", freshNonce))

	return nil
}

// GetCurrentNonce returns the current nonce without incrementing
func (nm *NonceManager) GetCurrentNonce(chain string, address common.Address) (uint64, bool) {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()

	if chainNonces, exists := nm.nonces[chain]; exists {
		if nonce, exists := chainNonces[address]; exists {
			return nonce, true
		}
	}

	return 0, false
}

// IncrementNonce manually increments the nonce for an address
func (nm *NonceManager) IncrementNonce(chain string, address common.Address) uint64 {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	// Initialize chain map if it doesn't exist
	if nm.nonces[chain] == nil {
		nm.nonces[chain] = make(map[common.Address]uint64)
	}

	currentNonce := nm.nonces[chain][address]
	nm.nonces[chain][address] = currentNonce + 1

	nm.logger.Debug("Incremented nonce",
		zap.String("chain", chain),
		zap.String("address", address.Hex()),
		zap.Uint64("new_nonce", currentNonce+1))

	return currentNonce + 1
}

// DecrementNonce manually decrements the nonce for an address (for failed transactions)
func (nm *NonceManager) DecrementNonce(chain string, address common.Address) uint64 {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	// Initialize chain map if it doesn't exist
	if nm.nonces[chain] == nil {
		nm.nonces[chain] = make(map[common.Address]uint64)
	}

	currentNonce := nm.nonces[chain][address]
	if currentNonce > 0 {
		nm.nonces[chain][address] = currentNonce - 1
	}

	nm.logger.Debug("Decremented nonce",
		zap.String("chain", chain),
		zap.String("address", address.Hex()),
		zap.Uint64("new_nonce", nm.nonces[chain][address]))

	return nm.nonces[chain][address]
}

// GetNonceGap checks for nonce gaps that might indicate failed transactions
func (nm *NonceManager) GetNonceGap(ctx context.Context, chain string, address common.Address) (uint64, error) {
	nm.mutex.RLock()
	localNonce, exists := nm.nonces[chain][address]
	nm.mutex.RUnlock()

	if !exists {
		return 0, fmt.Errorf("no local nonce found for address %s on chain %s", address.Hex(), chain)
	}

	// Get nonce from chain
	chainNonce := nm.fetchNonceFromChain(ctx, chain, address)

	if localNonce > chainNonce {
		gap := localNonce - chainNonce
		nm.logger.Warn("Nonce gap detected",
			zap.String("chain", chain),
			zap.String("address", address.Hex()),
			zap.Uint64("local_nonce", localNonce),
			zap.Uint64("chain_nonce", chainNonce),
			zap.Uint64("gap", gap))
		return gap, nil
	}

	return 0, nil
}

// SyncNonce synchronizes local nonce with chain nonce
func (nm *NonceManager) SyncNonce(ctx context.Context, chain string, address common.Address) error {
	chainNonce := nm.fetchNonceFromChain(ctx, chain, address)

	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	// Initialize chain map if it doesn't exist
	if nm.nonces[chain] == nil {
		nm.nonces[chain] = make(map[common.Address]uint64)
	}

	oldNonce := nm.nonces[chain][address]
	nm.nonces[chain][address] = chainNonce

	nm.logger.Info("Synchronized nonce",
		zap.String("chain", chain),
		zap.String("address", address.Hex()),
		zap.Uint64("old_nonce", oldNonce),
		zap.Uint64("new_nonce", chainNonce))

	return nil
}

// GetAllNonces returns all tracked nonces
func (nm *NonceManager) GetAllNonces() map[string]map[common.Address]uint64 {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()

	// Create a deep copy
	result := make(map[string]map[common.Address]uint64)
	for chain, chainNonces := range nm.nonces {
		result[chain] = make(map[common.Address]uint64)
		for address, nonce := range chainNonces {
			result[chain][address] = nonce
		}
	}

	return result
}

// ClearNonces clears all tracked nonces
func (nm *NonceManager) ClearNonces() {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	nm.nonces = make(map[string]map[common.Address]uint64)
	nm.logger.Info("Cleared all nonces")
}

// ClearChainNonces clears nonces for a specific chain
func (nm *NonceManager) ClearChainNonces(chain string) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	delete(nm.nonces, chain)
	nm.logger.Info("Cleared nonces for chain", zap.String("chain", chain))
}

// GetNonceStats returns statistics about tracked nonces
func (nm *NonceManager) GetNonceStats() map[string]interface{} {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()

	chainCount := len(nm.nonces)
	totalAddresses := 0
	chainStats := make(map[string]int)

	for chain, chainNonces := range nm.nonces {
		addressCount := len(chainNonces)
		totalAddresses += addressCount
		chainStats[chain] = addressCount
	}

	return map[string]interface{}{
		"total_chains":     chainCount,
		"total_addresses":  totalAddresses,
		"addresses_per_chain": chainStats,
	}
}

// fetchNonceFromChain fetches nonce from blockchain (mock implementation)
func (nm *NonceManager) fetchNonceFromChain(ctx context.Context, chain string, address common.Address) uint64 {
	// Mock implementation - in production, this would query the actual blockchain
	// For demonstration, we'll return a deterministic value based on address
	addressBytes := address.Bytes()
	nonce := uint64(addressBytes[19]) // Use last byte of address as base nonce
	
	nm.logger.Debug("Fetched nonce from chain (mock)",
		zap.String("chain", chain),
		zap.String("address", address.Hex()),
		zap.Uint64("nonce", nonce))

	return nonce
}

// ValidateNonce validates if a nonce is valid for an address
func (nm *NonceManager) ValidateNonce(ctx context.Context, chain string, address common.Address, nonce uint64) error {
	chainNonce := nm.fetchNonceFromChain(ctx, chain, address)

	if nonce < chainNonce {
		return fmt.Errorf("nonce %d is too low, chain nonce is %d", nonce, chainNonce)
	}

	// Check if nonce is too far ahead (potential issue)
	if nonce > chainNonce+100 {
		nm.logger.Warn("Nonce is far ahead of chain nonce",
			zap.String("chain", chain),
			zap.String("address", address.Hex()),
			zap.Uint64("provided_nonce", nonce),
			zap.Uint64("chain_nonce", chainNonce))
	}

	return nil
}
