package domain

import (
	"errors"
	"math/big"
	"strings"
	"time"
)

// TransactionType represents different types of blockchain transactions
type TransactionType int32

const (
	TransactionTypeUnknown      TransactionType = 0
	TransactionTypePayment      TransactionType = 1
	TransactionTypeTokenTransfer TransactionType = 2
	TransactionTypeContractCall TransactionType = 3
	TransactionTypeTokenMint    TransactionType = 4
	TransactionTypeTokenBurn    TransactionType = 5
	TransactionTypeStaking      TransactionType = 6
	TransactionTypeUnstaking    TransactionType = 7
	TransactionTypeSwap         TransactionType = 8
	TransactionTypeLoyaltyReward TransactionType = 9
)

// TransactionStatus represents the status of a blockchain transaction
type TransactionStatus int32

const (
	TransactionStatusUnknown   TransactionStatus = 0
	TransactionStatusPending   TransactionStatus = 1
	TransactionStatusConfirming TransactionStatus = 2
	TransactionStatusConfirmed TransactionStatus = 3
	TransactionStatusFailed    TransactionStatus = 4
	TransactionStatusDropped   TransactionStatus = 5
	TransactionStatusReplaced  TransactionStatus = 6
)

// String returns the string representation of TransactionStatus
func (s TransactionStatus) String() string {
	switch s {
	case TransactionStatusPending:
		return "PENDING"
	case TransactionStatusConfirming:
		return "CONFIRMING"
	case TransactionStatusConfirmed:
		return "CONFIRMED"
	case TransactionStatusFailed:
		return "FAILED"
	case TransactionStatusDropped:
		return "DROPPED"
	case TransactionStatusReplaced:
		return "REPLACED"
	default:
		return "UNKNOWN"
	}
}

// Transaction represents a blockchain transaction
type Transaction struct {
	ID              string            `json:"id"`
	Hash            string            `json:"hash"`
	Network         NetworkType       `json:"network"`
	Type            TransactionType   `json:"type"`
	Status          TransactionStatus `json:"status"`
	
	// Transaction details
	FromAddress     string            `json:"from_address"`
	ToAddress       string            `json:"to_address"`
	Value           string            `json:"value"`           // Amount in wei
	ValueFormatted  string            `json:"value_formatted"` // Human-readable amount
	TokenAddress    string            `json:"token_address,omitempty"`
	TokenSymbol     string            `json:"token_symbol,omitempty"`
	TokenDecimals   int32             `json:"token_decimals,omitempty"`
	
	// Gas details
	GasLimit        int64             `json:"gas_limit"`
	GasUsed         int64             `json:"gas_used,omitempty"`
	GasPrice        string            `json:"gas_price"`        // In wei
	MaxFeePerGas    string            `json:"max_fee_per_gas,omitempty"`    // EIP-1559
	MaxPriorityFee  string            `json:"max_priority_fee,omitempty"`   // EIP-1559
	
	// Block details
	BlockNumber     int64             `json:"block_number,omitempty"`
	BlockHash       string            `json:"block_hash,omitempty"`
	TransactionIndex int32            `json:"transaction_index,omitempty"`
	Confirmations   int32             `json:"confirmations"`
	
	// Contract interaction
	ContractAddress string            `json:"contract_address,omitempty"`
	InputData       string            `json:"input_data,omitempty"`
	MethodName      string            `json:"method_name,omitempty"`
	
	// Business context
	OrderID         string            `json:"order_id,omitempty"`
	PaymentID       string            `json:"payment_id,omitempty"`
	UserID          string            `json:"user_id,omitempty"`
	WalletID        string            `json:"wallet_id,omitempty"`
	
	// Timestamps
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	SubmittedAt     *time.Time        `json:"submitted_at,omitempty"`
	ConfirmedAt     *time.Time        `json:"confirmed_at,omitempty"`
	
	// Additional data
	Nonce           int64             `json:"nonce"`
	ErrorMessage    string            `json:"error_message,omitempty"`
	ReplacedByHash  string            `json:"replaced_by_hash,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// TransactionReceipt represents a transaction receipt
type TransactionReceipt struct {
	TransactionHash   string            `json:"transaction_hash"`
	TransactionIndex  int32             `json:"transaction_index"`
	BlockHash         string            `json:"block_hash"`
	BlockNumber       int64             `json:"block_number"`
	From              string            `json:"from"`
	To                string            `json:"to"`
	GasUsed           int64             `json:"gas_used"`
	CumulativeGasUsed int64             `json:"cumulative_gas_used"`
	ContractAddress   string            `json:"contract_address,omitempty"`
	Status            int32             `json:"status"` // 1 for success, 0 for failure
	Logs              []*TransactionLog `json:"logs"`
	LogsBloom         string            `json:"logs_bloom"`
	EffectiveGasPrice string            `json:"effective_gas_price,omitempty"`
}

// TransactionLog represents a transaction log/event
type TransactionLog struct {
	Address          string   `json:"address"`
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	BlockNumber      int64    `json:"block_number"`
	TransactionHash  string   `json:"transaction_hash"`
	TransactionIndex int32    `json:"transaction_index"`
	BlockHash        string   `json:"block_hash"`
	LogIndex         int32    `json:"log_index"`
	Removed          bool     `json:"removed"`
}

// NewTransaction creates a new transaction
func NewTransaction(network NetworkType, txType TransactionType, fromAddress, toAddress, value string) (*Transaction, error) {
	if !IsValidAddress(fromAddress) {
		return nil, errors.New("invalid from address")
	}
	
	if !IsValidAddress(toAddress) {
		return nil, errors.New("invalid to address")
	}
	
	// Validate value is a valid number
	if _, ok := new(big.Int).SetString(value, 10); !ok {
		return nil, errors.New("invalid value format")
	}

	return &Transaction{
		ID:             generateTransactionID(),
		Network:        network,
		Type:           txType,
		Status:         TransactionStatusPending,
		FromAddress:    strings.ToLower(fromAddress),
		ToAddress:      strings.ToLower(toAddress),
		Value:          value,
		Confirmations:  0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Metadata:       make(map[string]string),
	}, nil
}

// SetHash sets the transaction hash
func (t *Transaction) SetHash(hash string) {
	t.Hash = hash
	t.UpdatedAt = time.Now()
}

// SetStatus sets the transaction status
func (t *Transaction) SetStatus(status TransactionStatus) {
	t.Status = status
	t.UpdatedAt = time.Now()
	
	// Set timestamps for specific status changes
	now := time.Now()
	switch status {
	case TransactionStatusConfirming:
		if t.SubmittedAt == nil {
			t.SubmittedAt = &now
		}
	case TransactionStatusConfirmed:
		if t.ConfirmedAt == nil {
			t.ConfirmedAt = &now
		}
	}
}

// SetGasDetails sets gas-related details
func (t *Transaction) SetGasDetails(gasLimit int64, gasPrice string) error {
	if gasLimit <= 0 {
		return errors.New("gas limit must be positive")
	}
	
	// Validate gas price is a valid number
	if _, ok := new(big.Int).SetString(gasPrice, 10); !ok {
		return errors.New("invalid gas price format")
	}
	
	t.GasLimit = gasLimit
	t.GasPrice = gasPrice
	t.UpdatedAt = time.Now()
	
	return nil
}

// SetEIP1559Gas sets EIP-1559 gas details
func (t *Transaction) SetEIP1559Gas(gasLimit int64, maxFeePerGas, maxPriorityFee string) error {
	if gasLimit <= 0 {
		return errors.New("gas limit must be positive")
	}
	
	// Validate gas prices are valid numbers
	if _, ok := new(big.Int).SetString(maxFeePerGas, 10); !ok {
		return errors.New("invalid max fee per gas format")
	}
	
	if _, ok := new(big.Int).SetString(maxPriorityFee, 10); !ok {
		return errors.New("invalid max priority fee format")
	}
	
	t.GasLimit = gasLimit
	t.MaxFeePerGas = maxFeePerGas
	t.MaxPriorityFee = maxPriorityFee
	t.UpdatedAt = time.Now()
	
	return nil
}

// SetBlockDetails sets block-related details
func (t *Transaction) SetBlockDetails(blockNumber int64, blockHash string, txIndex int32) {
	t.BlockNumber = blockNumber
	t.BlockHash = blockHash
	t.TransactionIndex = txIndex
	t.UpdatedAt = time.Now()
}

// SetTokenDetails sets token-related details
func (t *Transaction) SetTokenDetails(tokenAddress, tokenSymbol string, tokenDecimals int32) error {
	if tokenAddress != "" && !IsValidAddress(tokenAddress) {
		return errors.New("invalid token address")
	}
	
	t.TokenAddress = strings.ToLower(tokenAddress)
	t.TokenSymbol = tokenSymbol
	t.TokenDecimals = tokenDecimals
	t.UpdatedAt = time.Now()
	
	return nil
}

// SetContractDetails sets contract interaction details
func (t *Transaction) SetContractDetails(contractAddress, inputData, methodName string) error {
	if contractAddress != "" && !IsValidAddress(contractAddress) {
		return errors.New("invalid contract address")
	}
	
	t.ContractAddress = strings.ToLower(contractAddress)
	t.InputData = inputData
	t.MethodName = methodName
	t.UpdatedAt = time.Now()
	
	return nil
}

// SetBusinessContext sets business-related context
func (t *Transaction) SetBusinessContext(orderID, paymentID, userID, walletID string) {
	t.OrderID = orderID
	t.PaymentID = paymentID
	t.UserID = userID
	t.WalletID = walletID
	t.UpdatedAt = time.Now()
}

// SetNonce sets the transaction nonce
func (t *Transaction) SetNonce(nonce int64) {
	t.Nonce = nonce
	t.UpdatedAt = time.Now()
}

// SetError sets an error message for failed transactions
func (t *Transaction) SetError(errorMessage string) {
	t.ErrorMessage = errorMessage
	t.SetStatus(TransactionStatusFailed)
}

// SetReplacedBy sets the hash of the transaction that replaced this one
func (t *Transaction) SetReplacedBy(replacementHash string) {
	t.ReplacedByHash = replacementHash
	t.SetStatus(TransactionStatusReplaced)
}

// UpdateConfirmations updates the confirmation count
func (t *Transaction) UpdateConfirmations(confirmations int32) {
	t.Confirmations = confirmations
	t.UpdatedAt = time.Now()
	
	// Auto-update status based on confirmations
	if t.Status == TransactionStatusConfirming && confirmations >= 12 {
		t.SetStatus(TransactionStatusConfirmed)
	}
}

// AddMetadata adds metadata to the transaction
func (t *Transaction) AddMetadata(key, value string) {
	t.Metadata[key] = value
	t.UpdatedAt = time.Now()
}

// IsConfirmed checks if the transaction is confirmed
func (t *Transaction) IsConfirmed() bool {
	return t.Status == TransactionStatusConfirmed
}

// IsPending checks if the transaction is pending
func (t *Transaction) IsPending() bool {
	return t.Status == TransactionStatusPending || t.Status == TransactionStatusConfirming
}

// IsFailed checks if the transaction failed
func (t *Transaction) IsFailed() bool {
	return t.Status == TransactionStatusFailed
}

// IsTokenTransfer checks if this is a token transfer
func (t *Transaction) IsTokenTransfer() bool {
	return t.TokenAddress != ""
}

// IsContractCall checks if this is a contract call
func (t *Transaction) IsContractCall() bool {
	return t.ContractAddress != ""
}

// GetTotalGasCost calculates the total gas cost
func (t *Transaction) GetTotalGasCost() *big.Int {
	if t.GasUsed == 0 || t.GasPrice == "" {
		return big.NewInt(0)
	}
	
	gasUsed := big.NewInt(t.GasUsed)
	gasPrice, _ := new(big.Int).SetString(t.GasPrice, 10)
	
	return new(big.Int).Mul(gasUsed, gasPrice)
}

// GetAge returns the age of the transaction
func (t *Transaction) GetAge() time.Duration {
	return time.Since(t.CreatedAt)
}

// GetConfirmationTime returns the time it took to confirm
func (t *Transaction) GetConfirmationTime() time.Duration {
	if t.ConfirmedAt == nil || t.SubmittedAt == nil {
		return 0
	}
	return t.ConfirmedAt.Sub(*t.SubmittedAt)
}

// Helper functions

// generateTransactionID generates a unique transaction ID
func generateTransactionID() string {
	return "tx_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

// Transaction Factory Functions

// NewPaymentTransaction creates a new payment transaction
func NewPaymentTransaction(network NetworkType, fromAddress, toAddress, value, orderID, paymentID string) (*Transaction, error) {
	tx, err := NewTransaction(network, TransactionTypePayment, fromAddress, toAddress, value)
	if err != nil {
		return nil, err
	}
	
	tx.OrderID = orderID
	tx.PaymentID = paymentID
	tx.AddMetadata("transaction_type", "payment")
	
	return tx, nil
}

// NewTokenTransferTransaction creates a new token transfer transaction
func NewTokenTransferTransaction(network NetworkType, fromAddress, toAddress, value, tokenAddress, tokenSymbol string, tokenDecimals int32) (*Transaction, error) {
	tx, err := NewTransaction(network, TransactionTypeTokenTransfer, fromAddress, toAddress, value)
	if err != nil {
		return nil, err
	}
	
	if err := tx.SetTokenDetails(tokenAddress, tokenSymbol, tokenDecimals); err != nil {
		return nil, err
	}
	
	tx.AddMetadata("transaction_type", "token_transfer")
	
	return tx, nil
}

// NewLoyaltyRewardTransaction creates a new loyalty reward transaction
func NewLoyaltyRewardTransaction(network NetworkType, toAddress, value, tokenAddress, userID string) (*Transaction, error) {
	// Use zero address as from for minting
	zeroAddress := "0x0000000000000000000000000000000000000000"
	
	tx, err := NewTransaction(network, TransactionTypeLoyaltyReward, zeroAddress, toAddress, value)
	if err != nil {
		return nil, err
	}
	
	tx.TokenAddress = strings.ToLower(tokenAddress)
	tx.UserID = userID
	tx.AddMetadata("transaction_type", "loyalty_reward")
	tx.AddMetadata("reward_type", "coffee_purchase")
	
	return tx, nil
}

// NewContractCallTransaction creates a new contract call transaction
func NewContractCallTransaction(network NetworkType, fromAddress, contractAddress, inputData, methodName string) (*Transaction, error) {
	tx, err := NewTransaction(network, TransactionTypeContractCall, fromAddress, contractAddress, "0")
	if err != nil {
		return nil, err
	}
	
	if err := tx.SetContractDetails(contractAddress, inputData, methodName); err != nil {
		return nil, err
	}
	
	tx.AddMetadata("transaction_type", "contract_call")
	
	return tx, nil
}
