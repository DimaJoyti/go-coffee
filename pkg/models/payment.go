package models

import "time"

// Wallet represents a cryptocurrency wallet
type Wallet struct {
	ID         string    `json:"id,omitempty" db:"id"`
	Address    string    `json:"address" db:"address"`
	PrivateKey string    `json:"private_key,omitempty" db:"private_key"` // Should be encrypted in production
	PublicKey  string    `json:"public_key" db:"public_key"`
	Network    string    `json:"network" db:"network"` // mainnet, testnet
	Type       string    `json:"type" db:"type"`       // bitcoin, ethereum, etc.
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// AddressValidation represents the result of address validation
type AddressValidation struct {
	Address string `json:"address"`
	Valid   bool   `json:"valid"`
	Type    string `json:"type,omitempty"`    // P2PKH, P2SH, etc.
	Network string `json:"network,omitempty"` // mainnet, testnet
}

// MultisigRequest represents a request to create a multisig address
type MultisigRequest struct {
	PublicKeys []string `json:"public_keys" binding:"required,min=2"`
	Threshold  int      `json:"threshold" binding:"required,min=1"`
	Testnet    bool     `json:"testnet"`
}

// MultisigAddress represents a multisig address
type MultisigAddress struct {
	ID         string    `json:"id,omitempty" db:"id"`
	Address    string    `json:"address" db:"address"`
	Threshold  int       `json:"threshold" db:"threshold"`
	PublicKeys []string  `json:"public_keys" db:"public_keys"`
	Network    string    `json:"network" db:"network"`
	Type       string    `json:"type" db:"type"` // P2SH, P2WSH, etc.
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// SignMessageRequest represents a request to sign a message
type SignMessageRequest struct {
	Message    string `json:"message" binding:"required"`
	PrivateKey string `json:"private_key" binding:"required"`
}

// SignMessageResponse represents the response from message signing
type SignMessageResponse struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
	Address   string `json:"address"`
}

// VerifyMessageRequest represents a request to verify a message signature
type VerifyMessageRequest struct {
	Message   string `json:"message" binding:"required"`
	Signature string `json:"signature" binding:"required"`
	Address   string `json:"address" binding:"required"`
}

// VerifyMessageResponse represents the response from message verification
type VerifyMessageResponse struct {
	Message string `json:"message"`
	Address string `json:"address"`
	Valid   bool   `json:"valid"`
}

// Transaction represents a cryptocurrency transaction
type Transaction struct {
	ID            string    `json:"id" db:"id"`
	Hash          string    `json:"hash" db:"hash"`
	FromAddress   string    `json:"from_address" db:"from_address"`
	ToAddress     string    `json:"to_address" db:"to_address"`
	Amount        int64     `json:"amount" db:"amount"` // Amount in satoshis
	Fee           int64     `json:"fee" db:"fee"`       // Fee in satoshis
	Status        string    `json:"status" db:"status"` // pending, confirmed, failed
	Network       string    `json:"network" db:"network"`
	Type          string    `json:"type" db:"type"` // bitcoin, ethereum, etc.
	BlockHeight   int64     `json:"block_height,omitempty" db:"block_height"`
	Confirmations int       `json:"confirmations" db:"confirmations"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// UTXO represents an unspent transaction output
type UTXO struct {
	TxHash       string `json:"tx_hash" db:"tx_hash"`
	OutputIndex  int    `json:"output_index" db:"output_index"`
	Amount       int64  `json:"amount" db:"amount"` // Amount in satoshis
	ScriptPubKey string `json:"script_pub_key" db:"script_pub_key"`
	Address      string `json:"address" db:"address"`
	Confirmations int   `json:"confirmations" db:"confirmations"`
}

// CreateTransactionRequest represents a request to create a transaction
type CreateTransactionRequest struct {
	FromAddress string  `json:"from_address" binding:"required"`
	ToAddress   string  `json:"to_address" binding:"required"`
	Amount      int64   `json:"amount" binding:"required,min=1"`
	FeeRate     int64   `json:"fee_rate,omitempty"` // Satoshis per byte
	PrivateKey  string  `json:"private_key" binding:"required"`
	UTXOs       []UTXO  `json:"utxos,omitempty"`
}

// CreateTransactionResponse represents the response from transaction creation
type CreateTransactionResponse struct {
	TransactionID   string `json:"transaction_id"`
	Hash           string `json:"hash"`
	RawTransaction string `json:"raw_transaction"`
	Fee            int64  `json:"fee"`
	Size           int    `json:"size"`
}

// PaymentStatus represents the status of a payment
type PaymentStatus struct {
	ID            string    `json:"id"`
	Status        string    `json:"status"` // pending, processing, completed, failed
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	TransactionID string    `json:"transaction_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PaymentRequest represents a payment request
type PaymentRequest struct {
	Amount      int64  `json:"amount" binding:"required,min=1"`
	Currency    string `json:"currency" binding:"required"`
	ToAddress   string `json:"to_address" binding:"required"`
	FromAddress string `json:"from_address,omitempty"`
	Description string `json:"description,omitempty"`
	OrderID     string `json:"order_id,omitempty"`
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
	PaymentID     string `json:"payment_id"`
	Status        string `json:"status"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	Address       string `json:"address"`
	QRCode        string `json:"qr_code,omitempty"`
	ExpiresAt     time.Time `json:"expires_at,omitempty"`
}

// Balance represents account balance
type Balance struct {
	Address      string `json:"address"`
	Balance      int64  `json:"balance"`      // Confirmed balance in satoshis
	Unconfirmed  int64  `json:"unconfirmed"`  // Unconfirmed balance in satoshis
	Currency     string `json:"currency"`
	Network      string `json:"network"`
	LastUpdated  time.Time `json:"last_updated"`
}
