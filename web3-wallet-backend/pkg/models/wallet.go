package models

import (
	"time"
)

// Wallet represents a blockchain wallet
type Wallet struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	Address   string    `json:"address" db:"address"`
	Chain     string    `json:"chain" db:"chain"`
	Type      string    `json:"type" db:"type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// WalletType represents the type of wallet
type WalletType string

const (
	// WalletTypeHD represents a hierarchical deterministic wallet
	WalletTypeHD WalletType = "hd"
	// WalletTypeImported represents an imported wallet
	WalletTypeImported WalletType = "imported"
	// WalletTypeMultisig represents a multi-signature wallet
	WalletTypeMultisig WalletType = "multisig"
)

// Chain represents a blockchain network
type Chain string

const (
	// ChainEthereum represents the Ethereum network
	ChainEthereum Chain = "ethereum"
	// ChainBSC represents the Binance Smart Chain network
	ChainBSC Chain = "bsc"
	// ChainPolygon represents the Polygon network
	ChainPolygon Chain = "polygon"
	// ChainSolana represents the Solana network
	ChainSolana Chain = "solana"
)

// CreateWalletRequest represents a request to create a wallet
type CreateWalletRequest struct {
	UserID string     `json:"user_id" validate:"required"`
	Name   string     `json:"name" validate:"required"`
	Chain  Chain      `json:"chain" validate:"required,oneof=ethereum bsc polygon solana"`
	Type   WalletType `json:"type" validate:"required,oneof=hd imported multisig"`
}

// CreateWalletResponse represents a response to a create wallet request
type CreateWalletResponse struct {
	Wallet         Wallet `json:"wallet"`
	Mnemonic       string `json:"mnemonic,omitempty"`
	PrivateKey     string `json:"private_key,omitempty"`
	DerivationPath string `json:"derivation_path,omitempty"`
}

// ImportWalletRequest represents a request to import a wallet
type ImportWalletRequest struct {
	UserID     string `json:"user_id" validate:"required"`
	Name       string `json:"name" validate:"required"`
	Chain      Chain  `json:"chain" validate:"required,oneof=ethereum bsc polygon solana"`
	PrivateKey string `json:"private_key" validate:"required"`
}

// ImportWalletResponse represents a response to an import wallet request
type ImportWalletResponse struct {
	Wallet Wallet `json:"wallet"`
}

// GetWalletRequest represents a request to get a wallet
type GetWalletRequest struct {
	ID string `json:"id" validate:"required"`
}

// GetWalletResponse represents a response to a get wallet request
type GetWalletResponse struct {
	Wallet Wallet `json:"wallet"`
}

// ListWalletsRequest represents a request to list wallets
type ListWalletsRequest struct {
	UserID string     `json:"user_id" validate:"required"`
	Chain  Chain      `json:"chain"`
	Type   WalletType `json:"type"`
	Limit  int        `json:"limit" validate:"min=1,max=100"`
	Offset int        `json:"offset" validate:"min=0"`
}

// ListWalletsResponse represents a response to a list wallets request
type ListWalletsResponse struct {
	Wallets []Wallet `json:"wallets"`
	Total   int      `json:"total"`
}

// GetBalanceRequest represents a request to get a wallet balance
type GetBalanceRequest struct {
	WalletID     string `json:"wallet_id" validate:"required"`
	TokenAddress string `json:"token_address"`
}

// GetBalanceResponse represents a response to a get balance request
type GetBalanceResponse struct {
	Balance      string `json:"balance"`
	Symbol       string `json:"symbol"`
	Decimals     int    `json:"decimals"`
	TokenAddress string `json:"token_address,omitempty"`
}

// ExportWalletRequest represents a request to export a wallet
type ExportWalletRequest struct {
	WalletID   string `json:"wallet_id" validate:"required"`
	Passphrase string `json:"passphrase" validate:"required"`
}

// ExportWalletResponse represents a response to an export wallet request
type ExportWalletResponse struct {
	PrivateKey string `json:"private_key"`
	Keystore   string `json:"keystore,omitempty"`
}

// DeleteWalletRequest represents a request to delete a wallet
type DeleteWalletRequest struct {
	WalletID string `json:"wallet_id" validate:"required"`
}

// DeleteWalletResponse represents a response to a delete wallet request
type DeleteWalletResponse struct {
	Success bool `json:"success"`
}
