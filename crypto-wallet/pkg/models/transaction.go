package models

import (
	"time"
)

// Transaction represents a blockchain transaction
type Transaction struct {
	ID            string          `json:"id" db:"id"`
	UserID        string          `json:"user_id" db:"user_id"`
	WalletID      string          `json:"wallet_id" db:"wallet_id"`
	Hash          string          `json:"hash" db:"hash"`
	From          string          `json:"from" db:"from_address"`
	To            string          `json:"to" db:"to_address"`
	Value         string          `json:"value" db:"value"`
	Gas           uint64          `json:"gas" db:"gas"`
	GasPrice      string          `json:"gas_price" db:"gas_price"`
	Nonce         uint64          `json:"nonce" db:"nonce"`
	Data          string          `json:"data" db:"data"`
	Chain         Chain           `json:"chain" db:"chain"`
	Status        TransactionStatus `json:"status" db:"status"`
	BlockNumber   uint64          `json:"block_number" db:"block_number"`
	BlockHash     string          `json:"block_hash" db:"block_hash"`
	Confirmations uint64          `json:"confirmations" db:"confirmations"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	// TransactionStatusPending represents a pending transaction
	TransactionStatusPending TransactionStatus = "pending"
	// TransactionStatusConfirmed represents a confirmed transaction
	TransactionStatusConfirmed TransactionStatus = "confirmed"
	// TransactionStatusFailed represents a failed transaction
	TransactionStatusFailed TransactionStatus = "failed"
)

// CreateTransactionRequest represents a request to create a transaction
type CreateTransactionRequest struct {
	WalletID    string `json:"wallet_id" validate:"required"`
	To          string `json:"to" validate:"required"`
	Value       string `json:"value" validate:"required"`
	Gas         uint64 `json:"gas"`
	GasPrice    string `json:"gas_price"`
	Data        string `json:"data"`
	Nonce       uint64 `json:"nonce"`
	Passphrase  string `json:"passphrase" validate:"required"`
}

// CreateTransactionResponse represents a response to a create transaction request
type CreateTransactionResponse struct {
	Transaction Transaction `json:"transaction"`
}

// GetTransactionRequest represents a request to get a transaction
type GetTransactionRequest struct {
	ID string `json:"id" validate:"required"`
}

// GetTransactionResponse represents a response to a get transaction request
type GetTransactionResponse struct {
	Transaction Transaction `json:"transaction"`
}

// GetTransactionByHashRequest represents a request to get a transaction by hash
type GetTransactionByHashRequest struct {
	Hash  string `json:"hash" validate:"required"`
	Chain Chain  `json:"chain" validate:"required,oneof=ethereum bsc polygon"`
}

// GetTransactionByHashResponse represents a response to a get transaction by hash request
type GetTransactionByHashResponse struct {
	Transaction Transaction `json:"transaction"`
}

// ListTransactionsRequest represents a request to list transactions
type ListTransactionsRequest struct {
	UserID   string            `json:"user_id" validate:"required"`
	WalletID string            `json:"wallet_id"`
	Status   TransactionStatus `json:"status"`
	Chain    Chain             `json:"chain"`
	Limit    int               `json:"limit" validate:"min=1,max=100"`
	Offset   int               `json:"offset" validate:"min=0"`
}

// ListTransactionsResponse represents a response to a list transactions request
type ListTransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
	Total        int           `json:"total"`
}

// EstimateGasRequest represents a request to estimate gas
type EstimateGasRequest struct {
	From  string `json:"from" validate:"required"`
	To    string `json:"to" validate:"required"`
	Value string `json:"value"`
	Data  string `json:"data"`
	Chain Chain  `json:"chain" validate:"required,oneof=ethereum bsc polygon"`
}

// EstimateGasResponse represents a response to an estimate gas request
type EstimateGasResponse struct {
	Gas uint64 `json:"gas"`
}

// GetGasPriceRequest represents a request to get gas price
type GetGasPriceRequest struct {
	Chain Chain `json:"chain" validate:"required,oneof=ethereum bsc polygon"`
}

// GetGasPriceResponse represents a response to a get gas price request
type GetGasPriceResponse struct {
	GasPrice string `json:"gas_price"`
	Slow     string `json:"slow"`
	Average  string `json:"average"`
	Fast     string `json:"fast"`
}

// GetTransactionReceiptRequest represents a request to get a transaction receipt
type GetTransactionReceiptRequest struct {
	Hash  string `json:"hash" validate:"required"`
	Chain Chain  `json:"chain" validate:"required,oneof=ethereum bsc polygon"`
}

// GetTransactionReceiptResponse represents a response to a get transaction receipt request
type GetTransactionReceiptResponse struct {
	BlockHash         string `json:"block_hash"`
	BlockNumber       uint64 `json:"block_number"`
	ContractAddress   string `json:"contract_address"`
	CumulativeGasUsed uint64 `json:"cumulative_gas_used"`
	From              string `json:"from"`
	GasUsed           uint64 `json:"gas_used"`
	Status            bool   `json:"status"`
	To                string `json:"to"`
	TransactionHash   string `json:"transaction_hash"`
	TransactionIndex  uint   `json:"transaction_index"`
	Logs              []Log  `json:"logs"`
}

// Log represents a transaction log
type Log struct {
	Address     string   `json:"address"`
	Topics      []string `json:"topics"`
	Data        string   `json:"data"`
	BlockNumber uint64   `json:"block_number"`
	TxHash      string   `json:"tx_hash"`
	TxIndex     uint     `json:"tx_index"`
	BlockHash   string   `json:"block_hash"`
	Index       uint     `json:"index"`
	Removed     bool     `json:"removed"`
}
