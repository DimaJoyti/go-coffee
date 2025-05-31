package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// WalletType represents different types of wallets
type WalletType int32

const (
	WalletTypeUnknown     WalletType = 0
	WalletTypeMetaMask    WalletType = 1
	WalletTypeWalletConnect WalletType = 2
	WalletTypeCoinbase    WalletType = 3
	WalletTypeTrustWallet WalletType = 4
	WalletTypeHardware    WalletType = 5
	WalletTypeCustodial   WalletType = 6
)

// WalletStatus represents the status of a wallet
type WalletStatus int32

const (
	WalletStatusUnknown     WalletStatus = 0
	WalletStatusConnected   WalletStatus = 1
	WalletStatusDisconnected WalletStatus = 2
	WalletStatusBlocked     WalletStatus = 3
	WalletStatusVerifying   WalletStatus = 4
)

// NetworkType represents different blockchain networks
type NetworkType int32

const (
	NetworkTypeUnknown  NetworkType = 0
	NetworkTypeEthereum NetworkType = 1
	NetworkTypePolygon  NetworkType = 2
	NetworkTypeBSC      NetworkType = 3
	NetworkTypeArbitrum NetworkType = 4
	NetworkTypeOptimism NetworkType = 5
	NetworkTypeAvalanche NetworkType = 6
)

// String returns the string representation of NetworkType
func (n NetworkType) String() string {
	switch n {
	case NetworkTypeEthereum:
		return "ethereum"
	case NetworkTypePolygon:
		return "polygon"
	case NetworkTypeBSC:
		return "bsc"
	case NetworkTypeArbitrum:
		return "arbitrum"
	case NetworkTypeOptimism:
		return "optimism"
	case NetworkTypeAvalanche:
		return "avalanche"
	default:
		return "unknown"
	}
}

// GetChainID returns the chain ID for the network
func (n NetworkType) GetChainID() int64 {
	switch n {
	case NetworkTypeEthereum:
		return 1
	case NetworkTypePolygon:
		return 137
	case NetworkTypeBSC:
		return 56
	case NetworkTypeArbitrum:
		return 42161
	case NetworkTypeOptimism:
		return 10
	case NetworkTypeAvalanche:
		return 43114
	default:
		return 0
	}
}

// Wallet represents a user's cryptocurrency wallet
type Wallet struct {
	ID              string            `json:"id"`
	UserID          string            `json:"user_id"`
	Address         string            `json:"address"`
	Type            WalletType        `json:"type"`
	Status          WalletStatus      `json:"status"`
	Networks        []NetworkType     `json:"networks"`
	Label           string            `json:"label,omitempty"`
	IsDefault       bool              `json:"is_default"`
	LastUsedAt      *time.Time        `json:"last_used_at,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	
	// Security features
	RequireSignature bool              `json:"require_signature"`
	TrustedContracts []string          `json:"trusted_contracts,omitempty"`
	SpendingLimits   map[string]int64  `json:"spending_limits,omitempty"` // token -> limit in wei
}

// WalletBalance represents the balance of tokens in a wallet
type WalletBalance struct {
	WalletID      string      `json:"wallet_id"`
	Network       NetworkType `json:"network"`
	TokenAddress  string      `json:"token_address"`
	TokenSymbol   string      `json:"token_symbol"`
	TokenDecimals int32       `json:"token_decimals"`
	Balance       string      `json:"balance"`       // Raw balance in wei/smallest unit
	BalanceFormatted string   `json:"balance_formatted"` // Human-readable balance
	USDValue      float64     `json:"usd_value,omitempty"`
	LastUpdatedAt time.Time   `json:"last_updated_at"`
}

// WalletConnection represents a wallet connection session
type WalletConnection struct {
	ID            string            `json:"id"`
	WalletID      string            `json:"wallet_id"`
	UserID        string            `json:"user_id"`
	SessionID     string            `json:"session_id"`
	IPAddress     string            `json:"ip_address"`
	UserAgent     string            `json:"user_agent"`
	ConnectedAt   time.Time         `json:"connected_at"`
	LastActiveAt  time.Time         `json:"last_active_at"`
	ExpiresAt     *time.Time        `json:"expires_at,omitempty"`
	IsActive      bool              `json:"is_active"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// NewWallet creates a new wallet
func NewWallet(userID, address string, walletType WalletType) (*Wallet, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	
	if !IsValidAddress(address) {
		return nil, errors.New("invalid wallet address")
	}

	return &Wallet{
		ID:               generateWalletID(),
		UserID:           userID,
		Address:          strings.ToLower(address),
		Type:             walletType,
		Status:           WalletStatusVerifying,
		Networks:         []NetworkType{NetworkTypeEthereum}, // Default to Ethereum
		IsDefault:        false,
		RequireSignature: true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Metadata:         make(map[string]string),
		TrustedContracts: make([]string, 0),
		SpendingLimits:   make(map[string]int64),
	}, nil
}

// SetStatus sets the wallet status
func (w *Wallet) SetStatus(status WalletStatus) {
	w.Status = status
	w.UpdatedAt = time.Now()
}

// AddNetwork adds a supported network
func (w *Wallet) AddNetwork(network NetworkType) {
	for _, n := range w.Networks {
		if n == network {
			return // Already exists
		}
	}
	w.Networks = append(w.Networks, network)
	w.UpdatedAt = time.Now()
}

// RemoveNetwork removes a supported network
func (w *Wallet) RemoveNetwork(network NetworkType) {
	for i, n := range w.Networks {
		if n == network {
			w.Networks = append(w.Networks[:i], w.Networks[i+1:]...)
			w.UpdatedAt = time.Now()
			break
		}
	}
}

// SetAsDefault sets this wallet as the default for the user
func (w *Wallet) SetAsDefault() {
	w.IsDefault = true
	w.UpdatedAt = time.Now()
}

// UnsetAsDefault removes the default status
func (w *Wallet) UnsetAsDefault() {
	w.IsDefault = false
	w.UpdatedAt = time.Now()
}

// SetLabel sets a custom label for the wallet
func (w *Wallet) SetLabel(label string) {
	w.Label = label
	w.UpdatedAt = time.Now()
}

// AddTrustedContract adds a trusted contract address
func (w *Wallet) AddTrustedContract(contractAddress string) error {
	if !IsValidAddress(contractAddress) {
		return errors.New("invalid contract address")
	}
	
	contractAddress = strings.ToLower(contractAddress)
	for _, addr := range w.TrustedContracts {
		if addr == contractAddress {
			return nil // Already exists
		}
	}
	
	w.TrustedContracts = append(w.TrustedContracts, contractAddress)
	w.UpdatedAt = time.Now()
	
	return nil
}

// RemoveTrustedContract removes a trusted contract address
func (w *Wallet) RemoveTrustedContract(contractAddress string) {
	contractAddress = strings.ToLower(contractAddress)
	for i, addr := range w.TrustedContracts {
		if addr == contractAddress {
			w.TrustedContracts = append(w.TrustedContracts[:i], w.TrustedContracts[i+1:]...)
			w.UpdatedAt = time.Now()
			break
		}
	}
}

// SetSpendingLimit sets a spending limit for a token
func (w *Wallet) SetSpendingLimit(tokenAddress string, limit int64) error {
	if !IsValidAddress(tokenAddress) {
		return errors.New("invalid token address")
	}
	
	if limit < 0 {
		return errors.New("spending limit cannot be negative")
	}
	
	w.SpendingLimits[strings.ToLower(tokenAddress)] = limit
	w.UpdatedAt = time.Now()
	
	return nil
}

// GetSpendingLimit gets the spending limit for a token
func (w *Wallet) GetSpendingLimit(tokenAddress string) (int64, bool) {
	limit, exists := w.SpendingLimits[strings.ToLower(tokenAddress)]
	return limit, exists
}

// SupportsNetwork checks if the wallet supports a specific network
func (w *Wallet) SupportsNetwork(network NetworkType) bool {
	for _, n := range w.Networks {
		if n == network {
			return true
		}
	}
	return false
}

// IsConnected checks if the wallet is connected
func (w *Wallet) IsConnected() bool {
	return w.Status == WalletStatusConnected
}

// IsBlocked checks if the wallet is blocked
func (w *Wallet) IsBlocked() bool {
	return w.Status == WalletStatusBlocked
}

// MarkAsUsed marks the wallet as recently used
func (w *Wallet) MarkAsUsed() {
	now := time.Now()
	w.LastUsedAt = &now
	w.UpdatedAt = now
}

// NewWalletConnection creates a new wallet connection
func NewWalletConnection(walletID, userID, sessionID, ipAddress, userAgent string) (*WalletConnection, error) {
	if walletID == "" {
		return nil, errors.New("wallet ID is required")
	}
	
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	
	if sessionID == "" {
		return nil, errors.New("session ID is required")
	}

	return &WalletConnection{
		ID:           generateConnectionID(),
		WalletID:     walletID,
		UserID:       userID,
		SessionID:    sessionID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ConnectedAt:  time.Now(),
		LastActiveAt: time.Now(),
		IsActive:     true,
		Metadata:     make(map[string]string),
	}, nil
}

// SetExpiration sets the connection expiration time
func (wc *WalletConnection) SetExpiration(duration time.Duration) {
	expiresAt := wc.ConnectedAt.Add(duration)
	wc.ExpiresAt = &expiresAt
}

// UpdateActivity updates the last active time
func (wc *WalletConnection) UpdateActivity() {
	wc.LastActiveAt = time.Now()
}

// Disconnect disconnects the wallet connection
func (wc *WalletConnection) Disconnect() {
	wc.IsActive = false
}

// IsExpired checks if the connection has expired
func (wc *WalletConnection) IsExpired() bool {
	if wc.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*wc.ExpiresAt)
}

// Validation functions

// IsValidAddress validates an Ethereum-style address
func IsValidAddress(address string) bool {
	if len(address) != 42 {
		return false
	}
	
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	
	// Check if the rest are valid hex characters
	hexPart := address[2:]
	matched, _ := regexp.MatchString("^[0-9a-fA-F]+$", hexPart)
	return matched
}

// IsValidPrivateKey validates a private key format
func IsValidPrivateKey(privateKey string) bool {
	if len(privateKey) == 64 {
		// Raw hex format
		matched, _ := regexp.MatchString("^[0-9a-fA-F]+$", privateKey)
		return matched
	}
	
	if len(privateKey) == 66 && strings.HasPrefix(privateKey, "0x") {
		// Prefixed hex format
		hexPart := privateKey[2:]
		matched, _ := regexp.MatchString("^[0-9a-fA-F]+$", hexPart)
		return matched
	}
	
	return false
}

// Helper functions

// generateWalletID generates a unique wallet ID
func generateWalletID() string {
	return "wallet_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

// generateConnectionID generates a unique connection ID
func generateConnectionID() string {
	return "conn_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

// generateRandomString generates a random string of given length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

// Network utility functions

// GetNetworkByChainID returns the network type for a given chain ID
func GetNetworkByChainID(chainID int64) NetworkType {
	switch chainID {
	case 1:
		return NetworkTypeEthereum
	case 137:
		return NetworkTypePolygon
	case 56:
		return NetworkTypeBSC
	case 42161:
		return NetworkTypeArbitrum
	case 10:
		return NetworkTypeOptimism
	case 43114:
		return NetworkTypeAvalanche
	default:
		return NetworkTypeUnknown
	}
}

// GetNetworkRPCURL returns the RPC URL for a network
func GetNetworkRPCURL(network NetworkType) string {
	switch network {
	case NetworkTypeEthereum:
		return "https://mainnet.infura.io/v3/"
	case NetworkTypePolygon:
		return "https://polygon-rpc.com"
	case NetworkTypeBSC:
		return "https://bsc-dataseed.binance.org"
	case NetworkTypeArbitrum:
		return "https://arb1.arbitrum.io/rpc"
	case NetworkTypeOptimism:
		return "https://mainnet.optimism.io"
	case NetworkTypeAvalanche:
		return "https://api.avax.network/ext/bc/C/rpc"
	default:
		return ""
	}
}

// GetNativeTokenSymbol returns the native token symbol for a network
func GetNativeTokenSymbol(network NetworkType) string {
	switch network {
	case NetworkTypeEthereum:
		return "ETH"
	case NetworkTypePolygon:
		return "MATIC"
	case NetworkTypeBSC:
		return "BNB"
	case NetworkTypeArbitrum:
		return "ETH"
	case NetworkTypeOptimism:
		return "ETH"
	case NetworkTypeAvalanche:
		return "AVAX"
	default:
		return ""
	}
}
