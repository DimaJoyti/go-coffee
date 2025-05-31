package domain

import (
	"errors"
	"time"
)

// PaymentStatus represents the status of a payment
type PaymentStatus int32

const (
	PaymentStatusUnknown   PaymentStatus = 0
	PaymentStatusPending   PaymentStatus = 1
	PaymentStatusProcessing PaymentStatus = 2
	PaymentStatusCompleted PaymentStatus = 3
	PaymentStatusFailed    PaymentStatus = 4
	PaymentStatusCancelled PaymentStatus = 5
	PaymentStatusRefunded  PaymentStatus = 6
)

// String returns the string representation of PaymentStatus
func (s PaymentStatus) String() string {
	switch s {
	case PaymentStatusPending:
		return "PENDING"
	case PaymentStatusProcessing:
		return "PROCESSING"
	case PaymentStatusCompleted:
		return "COMPLETED"
	case PaymentStatusFailed:
		return "FAILED"
	case PaymentStatusCancelled:
		return "CANCELLED"
	case PaymentStatusRefunded:
		return "REFUNDED"
	default:
		return "UNKNOWN"
	}
}

// CryptoNetwork represents different blockchain networks
type CryptoNetwork int32

const (
	CryptoNetworkUnknown  CryptoNetwork = 0
	CryptoNetworkEthereum CryptoNetwork = 1
	CryptoNetworkPolygon  CryptoNetwork = 2
	CryptoNetworkBSC      CryptoNetwork = 3
	CryptoNetworkArbitrum CryptoNetwork = 4
)

// Payment represents a payment transaction
type Payment struct {
	ID              string        `json:"id"`
	OrderID         string        `json:"order_id"`
	CustomerID      string        `json:"customer_id"`
	Amount          int64         `json:"amount"` // in cents or smallest unit
	Currency        string        `json:"currency"`
	PaymentMethod   PaymentMethod `json:"payment_method"`
	Status          PaymentStatus `json:"status"`
	
	// Traditional payment fields
	CardLast4       string `json:"card_last4,omitempty"`
	CardBrand       string `json:"card_brand,omitempty"`
	
	// Crypto payment fields
	CryptoNetwork   CryptoNetwork `json:"crypto_network,omitempty"`
	CryptoToken     string        `json:"crypto_token,omitempty"`
	WalletAddress   string        `json:"wallet_address,omitempty"`
	TransactionHash string        `json:"transaction_hash,omitempty"`
	BlockNumber     int64         `json:"block_number,omitempty"`
	GasUsed         int64         `json:"gas_used,omitempty"`
	GasPrice        int64         `json:"gas_price,omitempty"`
	
	// Payment processor fields
	ProcessorID     string `json:"processor_id,omitempty"`
	ProcessorRef    string `json:"processor_ref,omitempty"`
	
	// Timestamps
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	FailedAt        *time.Time `json:"failed_at,omitempty"`
	
	// Additional data
	FailureReason   string            `json:"failure_reason,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// Refund represents a payment refund
type Refund struct {
	ID              string        `json:"id"`
	PaymentID       string        `json:"payment_id"`
	OrderID         string        `json:"order_id"`
	Amount          int64         `json:"amount"` // in cents or smallest unit
	Currency        string        `json:"currency"`
	Reason          string        `json:"reason"`
	Status          PaymentStatus `json:"status"`
	ProcessorRef    string        `json:"processor_ref,omitempty"`
	TransactionHash string        `json:"transaction_hash,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	CompletedAt     *time.Time    `json:"completed_at,omitempty"`
	FailedAt        *time.Time    `json:"failed_at,omitempty"`
	FailureReason   string        `json:"failure_reason,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// LoyaltyTokenPayment represents a payment using loyalty tokens
type LoyaltyTokenPayment struct {
	ID              string    `json:"id"`
	PaymentID       string    `json:"payment_id"`
	CustomerID      string    `json:"customer_id"`
	TokensUsed      int64     `json:"tokens_used"`
	TokenValue      int64     `json:"token_value"` // value in cents
	ExchangeRate    float64   `json:"exchange_rate"` // tokens per dollar
	TransactionHash string    `json:"transaction_hash,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// Business Rules and Validation

// NewPayment creates a new payment with validation
func NewPayment(orderID, customerID string, amount int64, currency string, method PaymentMethod) (*Payment, error) {
	if orderID == "" {
		return nil, errors.New("order ID is required")
	}
	
	if customerID == "" {
		return nil, errors.New("customer ID is required")
	}
	
	if amount <= 0 {
		return nil, errors.New("payment amount must be positive")
	}
	
	if currency == "" {
		return nil, errors.New("currency is required")
	}

	payment := &Payment{
		ID:            generatePaymentID(),
		OrderID:       orderID,
		CustomerID:    customerID,
		Amount:        amount,
		Currency:      currency,
		PaymentMethod: method,
		Status:        PaymentStatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Metadata:      make(map[string]string),
	}

	return payment, nil
}

// CanTransitionTo checks if the payment can transition to the given status
func (p *Payment) CanTransitionTo(newStatus PaymentStatus) bool {
	switch p.Status {
	case PaymentStatusPending:
		return newStatus == PaymentStatusProcessing || newStatus == PaymentStatusCancelled
	case PaymentStatusProcessing:
		return newStatus == PaymentStatusCompleted || newStatus == PaymentStatusFailed
	case PaymentStatusCompleted:
		return newStatus == PaymentStatusRefunded
	case PaymentStatusFailed:
		return newStatus == PaymentStatusPending // Allow retry
	case PaymentStatusCancelled:
		return false // Terminal state
	case PaymentStatusRefunded:
		return false // Terminal state
	default:
		return false
	}
}

// UpdateStatus updates the payment status with validation
func (p *Payment) UpdateStatus(newStatus PaymentStatus) error {
	if !p.CanTransitionTo(newStatus) {
		return errors.New("invalid payment status transition")
	}
	
	p.Status = newStatus
	p.UpdatedAt = time.Now()
	
	// Set timestamps for specific status changes
	now := time.Now()
	switch newStatus {
	case PaymentStatusProcessing:
		p.ProcessedAt = &now
	case PaymentStatusCompleted:
		p.CompletedAt = &now
	case PaymentStatusFailed:
		p.FailedAt = &now
	}
	
	return nil
}

// SetCryptoDetails sets cryptocurrency payment details
func (p *Payment) SetCryptoDetails(network CryptoNetwork, token, walletAddress, txHash string) {
	p.CryptoNetwork = network
	p.CryptoToken = token
	p.WalletAddress = walletAddress
	p.TransactionHash = txHash
	p.UpdatedAt = time.Now()
}

// SetCardDetails sets credit/debit card payment details
func (p *Payment) SetCardDetails(last4, brand string) {
	p.CardLast4 = last4
	p.CardBrand = brand
	p.UpdatedAt = time.Now()
}

// SetProcessorDetails sets payment processor details
func (p *Payment) SetProcessorDetails(processorID, processorRef string) {
	p.ProcessorID = processorID
	p.ProcessorRef = processorRef
	p.UpdatedAt = time.Now()
}

// SetFailureReason sets the failure reason for failed payments
func (p *Payment) SetFailureReason(reason string) {
	p.FailureReason = reason
	p.UpdatedAt = time.Now()
}

// IsExpired checks if the payment has expired (pending for too long)
func (p *Payment) IsExpired(timeout time.Duration) bool {
	if p.Status != PaymentStatusPending {
		return false
	}
	return time.Since(p.CreatedAt) > timeout
}

// IsCryptoPayment checks if this is a cryptocurrency payment
func (p *Payment) IsCryptoPayment() bool {
	return p.PaymentMethod == PaymentMethodCrypto
}

// IsLoyaltyTokenPayment checks if this is a loyalty token payment
func (p *Payment) IsLoyaltyTokenPayment() bool {
	return p.PaymentMethod == PaymentMethodLoyaltyToken
}

// NewRefund creates a new refund with validation
func NewRefund(paymentID, orderID string, amount int64, currency, reason string) (*Refund, error) {
	if paymentID == "" {
		return nil, errors.New("payment ID is required")
	}
	
	if orderID == "" {
		return nil, errors.New("order ID is required")
	}
	
	if amount <= 0 {
		return nil, errors.New("refund amount must be positive")
	}
	
	if reason == "" {
		return nil, errors.New("refund reason is required")
	}

	refund := &Refund{
		ID:        generateRefundID(),
		PaymentID: paymentID,
		OrderID:   orderID,
		Amount:    amount,
		Currency:  currency,
		Reason:    reason,
		Status:    PaymentStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]string),
	}

	return refund, nil
}

// Helper functions

// generatePaymentID generates a unique payment ID
func generatePaymentID() string {
	return "pay_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

// generateRefundID generates a unique refund ID
func generateRefundID() string {
	return "ref_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}
