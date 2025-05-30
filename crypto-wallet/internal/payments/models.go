package payments

import (
	"time"
)

// Payment represents a payment transaction in the system
type Payment struct {
	ID                string                 `json:"id" db:"id"`
	AccountID         string                 `json:"account_id" db:"account_id"`
	WalletID          string                 `json:"wallet_id" db:"wallet_id"`
	MerchantID        string                 `json:"merchant_id" db:"merchant_id"`
	OrderID           string                 `json:"order_id" db:"order_id"`
	PaymentMethod     PaymentMethod          `json:"payment_method" db:"payment_method"`
	Currency          string                 `json:"currency" db:"currency"`
	Amount            string                 `json:"amount" db:"amount"`
	FeeAmount         string                 `json:"fee_amount" db:"fee_amount"`
	NetAmount         string                 `json:"net_amount" db:"net_amount"`
	ExchangeRate      string                 `json:"exchange_rate" db:"exchange_rate"`
	Status            PaymentStatus          `json:"status" db:"status"`
	PaymentType       PaymentType            `json:"payment_type" db:"payment_type"`
	Description       string                 `json:"description" db:"description"`
	Reference         string                 `json:"reference" db:"reference"`
	TransactionHash   string                 `json:"transaction_hash" db:"transaction_hash"`
	BlockNumber       int64                  `json:"block_number" db:"block_number"`
	Confirmations     int                    `json:"confirmations" db:"confirmations"`
	Network           string                 `json:"network" db:"network"`
	FromAddress       string                 `json:"from_address" db:"from_address"`
	ToAddress         string                 `json:"to_address" db:"to_address"`
	GasUsed           string                 `json:"gas_used" db:"gas_used"`
	GasPrice          string                 `json:"gas_price" db:"gas_price"`
	RiskScore         float64                `json:"risk_score" db:"risk_score"`
	FraudFlags        []string               `json:"fraud_flags" db:"fraud_flags"`
	ProcessedAt       *time.Time             `json:"processed_at" db:"processed_at"`
	SettledAt         *time.Time             `json:"settled_at" db:"settled_at"`
	ExpiresAt         *time.Time             `json:"expires_at" db:"expires_at"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// PaymentMethod represents the method used for payment
type PaymentMethod string

const (
	PaymentMethodCrypto     PaymentMethod = "crypto"
	PaymentMethodCard       PaymentMethod = "card"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodWallet     PaymentMethod = "wallet"
	PaymentMethodStablecoin PaymentMethod = "stablecoin"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted  PaymentStatus = "completed"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusCancelled  PaymentStatus = "cancelled"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusExpired    PaymentStatus = "expired"
	PaymentStatusOnHold     PaymentStatus = "on_hold"
)

// PaymentType represents the type of payment
type PaymentType string

const (
	PaymentTypeInbound  PaymentType = "inbound"
	PaymentTypeOutbound PaymentType = "outbound"
	PaymentTypeInternal PaymentType = "internal"
	PaymentTypeRefund   PaymentType = "refund"
)

// PaymentIntent represents a payment intent before processing
type PaymentIntent struct {
	ID              string                 `json:"id" db:"id"`
	AccountID       string                 `json:"account_id" db:"account_id"`
	MerchantID      string                 `json:"merchant_id" db:"merchant_id"`
	Amount          string                 `json:"amount" db:"amount"`
	Currency        string                 `json:"currency" db:"currency"`
	PaymentMethods  []PaymentMethod        `json:"payment_methods" db:"payment_methods"`
	Description     string                 `json:"description" db:"description"`
	ReturnURL       string                 `json:"return_url" db:"return_url"`
	CancelURL       string                 `json:"cancel_url" db:"cancel_url"`
	WebhookURL      string                 `json:"webhook_url" db:"webhook_url"`
	Status          PaymentIntentStatus    `json:"status" db:"status"`
	ClientSecret    string                 `json:"client_secret" db:"client_secret"`
	ExpiresAt       time.Time              `json:"expires_at" db:"expires_at"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// PaymentIntentStatus represents the status of a payment intent
type PaymentIntentStatus string

const (
	PaymentIntentStatusCreated    PaymentIntentStatus = "created"
	PaymentIntentStatusProcessing PaymentIntentStatus = "processing"
	PaymentIntentStatusSucceeded  PaymentIntentStatus = "succeeded"
	PaymentIntentStatusFailed     PaymentIntentStatus = "failed"
	PaymentIntentStatusCancelled  PaymentIntentStatus = "cancelled"
	PaymentIntentStatusExpired    PaymentIntentStatus = "expired"
)

// Refund represents a payment refund
type Refund struct {
	ID          string                 `json:"id" db:"id"`
	PaymentID   string                 `json:"payment_id" db:"payment_id"`
	AccountID   string                 `json:"account_id" db:"account_id"`
	Amount      string                 `json:"amount" db:"amount"`
	Currency    string                 `json:"currency" db:"currency"`
	Reason      RefundReason           `json:"reason" db:"reason"`
	Status      RefundStatus           `json:"status" db:"status"`
	Description string                 `json:"description" db:"description"`
	ProcessedAt *time.Time             `json:"processed_at" db:"processed_at"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// RefundReason represents the reason for a refund
type RefundReason string

const (
	RefundReasonRequested     RefundReason = "requested"
	RefundReasonFraud         RefundReason = "fraud"
	RefundReasonDuplicate     RefundReason = "duplicate"
	RefundReasonError         RefundReason = "error"
	RefundReasonChargeback    RefundReason = "chargeback"
	RefundReasonCancellation  RefundReason = "cancellation"
)

// RefundStatus represents the status of a refund
type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusProcessing RefundStatus = "processing"
	RefundStatusCompleted RefundStatus = "completed"
	RefundStatusFailed    RefundStatus = "failed"
	RefundStatusCancelled RefundStatus = "cancelled"
)

// FeeStructure represents the fee structure for payments
type FeeStructure struct {
	ID              string                 `json:"id" db:"id"`
	Name            string                 `json:"name" db:"name"`
	PaymentMethod   PaymentMethod          `json:"payment_method" db:"payment_method"`
	Currency        string                 `json:"currency" db:"currency"`
	FeeType         FeeType                `json:"fee_type" db:"fee_type"`
	FixedFee        string                 `json:"fixed_fee" db:"fixed_fee"`
	PercentageFee   float64                `json:"percentage_fee" db:"percentage_fee"`
	MinFee          string                 `json:"min_fee" db:"min_fee"`
	MaxFee          string                 `json:"max_fee" db:"max_fee"`
	TierRules       []FeeTier              `json:"tier_rules" db:"tier_rules"`
	IsActive        bool                   `json:"is_active" db:"is_active"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// FeeType represents the type of fee calculation
type FeeType string

const (
	FeeTypeFixed      FeeType = "fixed"
	FeeTypePercentage FeeType = "percentage"
	FeeTypeTiered     FeeType = "tiered"
	FeeTypeHybrid     FeeType = "hybrid"
)

// FeeTier represents a fee tier for tiered pricing
type FeeTier struct {
	MinAmount     string  `json:"min_amount"`
	MaxAmount     string  `json:"max_amount"`
	FixedFee      string  `json:"fixed_fee"`
	PercentageFee float64 `json:"percentage_fee"`
}

// Settlement represents a payment settlement
type Settlement struct {
	ID            string                 `json:"id" db:"id"`
	MerchantID    string                 `json:"merchant_id" db:"merchant_id"`
	Currency      string                 `json:"currency" db:"currency"`
	Amount        string                 `json:"amount" db:"amount"`
	FeeAmount     string                 `json:"fee_amount" db:"fee_amount"`
	NetAmount     string                 `json:"net_amount" db:"net_amount"`
	PaymentCount  int                    `json:"payment_count" db:"payment_count"`
	Status        SettlementStatus       `json:"status" db:"status"`
	SettledAt     *time.Time             `json:"settled_at" db:"settled_at"`
	PeriodStart   time.Time              `json:"period_start" db:"period_start"`
	PeriodEnd     time.Time              `json:"period_end" db:"period_end"`
	Metadata      map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// SettlementStatus represents the status of a settlement
type SettlementStatus string

const (
	SettlementStatusPending   SettlementStatus = "pending"
	SettlementStatusProcessing SettlementStatus = "processing"
	SettlementStatusCompleted SettlementStatus = "completed"
	SettlementStatusFailed    SettlementStatus = "failed"
	SettlementStatusOnHold    SettlementStatus = "on_hold"
)

// CreatePaymentRequest represents a request to create a payment
type CreatePaymentRequest struct {
	AccountID     string                 `json:"account_id" validate:"required"`
	WalletID      string                 `json:"wallet_id,omitempty"`
	MerchantID    string                 `json:"merchant_id,omitempty"`
	OrderID       string                 `json:"order_id,omitempty"`
	PaymentMethod PaymentMethod          `json:"payment_method" validate:"required"`
	Currency      string                 `json:"currency" validate:"required"`
	Amount        string                 `json:"amount" validate:"required"`
	Description   string                 `json:"description"`
	Reference     string                 `json:"reference"`
	ToAddress     string                 `json:"to_address,omitempty"`
	Network       string                 `json:"network,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CreatePaymentIntentRequest represents a request to create a payment intent
type CreatePaymentIntentRequest struct {
	AccountID      string                 `json:"account_id" validate:"required"`
	MerchantID     string                 `json:"merchant_id,omitempty"`
	Amount         string                 `json:"amount" validate:"required"`
	Currency       string                 `json:"currency" validate:"required"`
	PaymentMethods []PaymentMethod        `json:"payment_methods" validate:"required"`
	Description    string                 `json:"description"`
	ReturnURL      string                 `json:"return_url"`
	CancelURL      string                 `json:"cancel_url"`
	WebhookURL     string                 `json:"webhook_url"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ConfirmPaymentRequest represents a request to confirm a payment
type ConfirmPaymentRequest struct {
	PaymentIntentID string                 `json:"payment_intent_id" validate:"required"`
	PaymentMethod   PaymentMethod          `json:"payment_method" validate:"required"`
	WalletID        string                 `json:"wallet_id,omitempty"`
	FromAddress     string                 `json:"from_address,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// CreateRefundRequest represents a request to create a refund
type CreateRefundRequest struct {
	PaymentID   string                 `json:"payment_id" validate:"required"`
	Amount      string                 `json:"amount,omitempty"` // If empty, full refund
	Reason      RefundReason           `json:"reason" validate:"required"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PaymentListRequest represents a request to list payments
type PaymentListRequest struct {
	Page          int           `json:"page" validate:"min=1"`
	Limit         int           `json:"limit" validate:"min=1,max=100"`
	AccountID     string        `json:"account_id,omitempty"`
	MerchantID    string        `json:"merchant_id,omitempty"`
	Status        PaymentStatus `json:"status,omitempty"`
	PaymentMethod PaymentMethod `json:"payment_method,omitempty"`
	Currency      string        `json:"currency,omitempty"`
	DateFrom      *time.Time    `json:"date_from,omitempty"`
	DateTo        *time.Time    `json:"date_to,omitempty"`
	MinAmount     string        `json:"min_amount,omitempty"`
	MaxAmount     string        `json:"max_amount,omitempty"`
}

// PaymentListResponse represents a response to list payments
type PaymentListResponse struct {
	Payments   []Payment `json:"payments"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}

// PaymentWebhook represents a payment webhook event
type PaymentWebhook struct {
	ID        string                 `json:"id"`
	Event     string                 `json:"event"`
	Payment   *Payment               `json:"payment,omitempty"`
	Refund    *Refund                `json:"refund,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
