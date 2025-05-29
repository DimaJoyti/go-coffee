package cards

import (
	"time"
)

// Card represents a payment card (virtual or physical)
type Card struct {
	ID                string                 `json:"id" db:"id"`
	AccountID         string                 `json:"account_id" db:"account_id"`
	WalletID          string                 `json:"wallet_id" db:"wallet_id"`
	CardNumber        string                 `json:"card_number" db:"card_number"`
	MaskedNumber      string                 `json:"masked_number" db:"masked_number"`
	CardType          CardType               `json:"card_type" db:"card_type"`
	CardBrand         CardBrand              `json:"card_brand" db:"card_brand"`
	CardNetwork       CardNetwork            `json:"card_network" db:"card_network"`
	Status            CardStatus             `json:"status" db:"status"`
	Currency          string                 `json:"currency" db:"currency"`
	Balance           string                 `json:"balance" db:"balance"`
	AvailableBalance  string                 `json:"available_balance" db:"available_balance"`
	SpendingLimits    SpendingLimits         `json:"spending_limits" db:"spending_limits"`
	SecuritySettings  SecuritySettings       `json:"security_settings" db:"security_settings"`
	ExpiryMonth       int                    `json:"expiry_month" db:"expiry_month"`
	ExpiryYear        int                    `json:"expiry_year" db:"expiry_year"`
	CVV               string                 `json:"cvv" db:"cvv"`
	PIN               string                 `json:"pin" db:"pin"`
	HolderName        string                 `json:"holder_name" db:"holder_name"`
	BillingAddress    BillingAddress         `json:"billing_address" db:"billing_address"`
	ShippingAddress   *ShippingAddress       `json:"shipping_address" db:"shipping_address"`
	DesignID          string                 `json:"design_id" db:"design_id"`
	IssuedAt          *time.Time             `json:"issued_at" db:"issued_at"`
	ActivatedAt       *time.Time             `json:"activated_at" db:"activated_at"`
	LastUsedAt        *time.Time             `json:"last_used_at" db:"last_used_at"`
	BlockedAt         *time.Time             `json:"blocked_at" db:"blocked_at"`
	ExpiresAt         time.Time              `json:"expires_at" db:"expires_at"`
	RewardsProgram    *RewardsProgram        `json:"rewards_program" db:"rewards_program"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// CardType represents the type of card
type CardType string

const (
	CardTypeVirtual  CardType = "virtual"
	CardTypePhysical CardType = "physical"
)

// CardBrand represents the card brand
type CardBrand string

const (
	CardBrandVisa       CardBrand = "visa"
	CardBrandMastercard CardBrand = "mastercard"
	CardBrandAmex       CardBrand = "amex"
	CardBrandDiscover   CardBrand = "discover"
	CardBrandUnionPay   CardBrand = "unionpay"
)

// CardNetwork represents the card network
type CardNetwork string

const (
	CardNetworkVisa       CardNetwork = "visa"
	CardNetworkMastercard CardNetwork = "mastercard"
	CardNetworkAmex       CardNetwork = "amex"
	CardNetworkDiscover   CardNetwork = "discover"
)

// CardStatus represents the status of a card
type CardStatus string

const (
	CardStatusPending   CardStatus = "pending"
	CardStatusActive    CardStatus = "active"
	CardStatusInactive  CardStatus = "inactive"
	CardStatusBlocked   CardStatus = "blocked"
	CardStatusSuspended CardStatus = "suspended"
	CardStatusExpired   CardStatus = "expired"
	CardStatusCancelled CardStatus = "cancelled"
)

// SpendingLimits represents spending limits for a card
type SpendingLimits struct {
	DailyLimit        string                 `json:"daily_limit" db:"daily_limit"`
	WeeklyLimit       string                 `json:"weekly_limit" db:"weekly_limit"`
	MonthlyLimit      string                 `json:"monthly_limit" db:"monthly_limit"`
	TransactionLimit  string                 `json:"transaction_limit" db:"transaction_limit"`
	ATMLimit          string                 `json:"atm_limit" db:"atm_limit"`
	OnlineLimit       string                 `json:"online_limit" db:"online_limit"`
	ContactlessLimit  string                 `json:"contactless_limit" db:"contactless_limit"`
	MerchantCategories []string              `json:"merchant_categories" db:"merchant_categories"`
	BlockedMerchants  []string               `json:"blocked_merchants" db:"blocked_merchants"`
	AllowedCountries  []string               `json:"allowed_countries" db:"allowed_countries"`
	BlockedCountries  []string               `json:"blocked_countries" db:"blocked_countries"`
	TimeRestrictions  map[string]interface{} `json:"time_restrictions" db:"time_restrictions"`
}

// SecuritySettings represents security settings for a card
type SecuritySettings struct {
	PINRequired         bool   `json:"pin_required" db:"pin_required"`
	CVVRequired         bool   `json:"cvv_required" db:"cvv_required"`
	BiometricRequired   bool   `json:"biometric_required" db:"biometric_required"`
	TokenizationEnabled bool   `json:"tokenization_enabled" db:"tokenization_enabled"`
	CVVRotationEnabled  bool   `json:"cvv_rotation_enabled" db:"cvv_rotation_enabled"`
	CVVRotationInterval string `json:"cvv_rotation_interval" db:"cvv_rotation_interval"`
	FraudDetection      bool   `json:"fraud_detection" db:"fraud_detection"`
	VelocityChecks      bool   `json:"velocity_checks" db:"velocity_checks"`
	GeofencingEnabled   bool   `json:"geofencing_enabled" db:"geofencing_enabled"`
	NotificationsEnabled bool  `json:"notifications_enabled" db:"notifications_enabled"`
}

// BillingAddress represents a billing address
type BillingAddress struct {
	FirstName   string `json:"first_name" db:"first_name"`
	LastName    string `json:"last_name" db:"last_name"`
	Company     string `json:"company" db:"company"`
	AddressLine1 string `json:"address_line1" db:"address_line1"`
	AddressLine2 string `json:"address_line2" db:"address_line2"`
	City        string `json:"city" db:"city"`
	State       string `json:"state" db:"state"`
	PostalCode  string `json:"postal_code" db:"postal_code"`
	Country     string `json:"country" db:"country"`
}

// ShippingAddress represents a shipping address for physical cards
type ShippingAddress struct {
	FirstName    string `json:"first_name" db:"first_name"`
	LastName     string `json:"last_name" db:"last_name"`
	Company      string `json:"company" db:"company"`
	AddressLine1 string `json:"address_line1" db:"address_line1"`
	AddressLine2 string `json:"address_line2" db:"address_line2"`
	City         string `json:"city" db:"city"`
	State        string `json:"state" db:"state"`
	PostalCode   string `json:"postal_code" db:"postal_code"`
	Country      string `json:"country" db:"country"`
	Phone        string `json:"phone" db:"phone"`
}

// RewardsProgram represents a card rewards program
type RewardsProgram struct {
	Enabled             bool                   `json:"enabled" db:"enabled"`
	ProgramType         RewardsProgramType     `json:"program_type" db:"program_type"`
	CashbackRate        float64                `json:"cashback_rate" db:"cashback_rate"`
	PointsMultiplier    float64                `json:"points_multiplier" db:"points_multiplier"`
	CryptoRewards       bool                   `json:"crypto_rewards" db:"crypto_rewards"`
	RewardsToken        string                 `json:"rewards_token" db:"rewards_token"`
	CategoryMultipliers map[string]float64     `json:"category_multipliers" db:"category_multipliers"`
	TotalEarned         string                 `json:"total_earned" db:"total_earned"`
	AvailableBalance    string                 `json:"available_balance" db:"available_balance"`
	LastRewardDate      *time.Time             `json:"last_reward_date" db:"last_reward_date"`
}

// RewardsProgramType represents the type of rewards program
type RewardsProgramType string

const (
	RewardsProgramTypeCashback RewardsProgramType = "cashback"
	RewardsProgramTypePoints   RewardsProgramType = "points"
	RewardsProgramTypeCrypto   RewardsProgramType = "crypto"
	RewardsProgramTypeMiles    RewardsProgramType = "miles"
)

// Transaction represents a card transaction
type Transaction struct {
	ID                string                 `json:"id" db:"id"`
	CardID            string                 `json:"card_id" db:"card_id"`
	AccountID         string                 `json:"account_id" db:"account_id"`
	MerchantName      string                 `json:"merchant_name" db:"merchant_name"`
	MerchantCategory  string                 `json:"merchant_category" db:"merchant_category"`
	MerchantID        string                 `json:"merchant_id" db:"merchant_id"`
	Amount            string                 `json:"amount" db:"amount"`
	Currency          string                 `json:"currency" db:"currency"`
	OriginalAmount    string                 `json:"original_amount" db:"original_amount"`
	OriginalCurrency  string                 `json:"original_currency" db:"original_currency"`
	ExchangeRate      string                 `json:"exchange_rate" db:"exchange_rate"`
	Fee               string                 `json:"fee" db:"fee"`
	TransactionType   TransactionType        `json:"transaction_type" db:"transaction_type"`
	Status            TransactionStatus      `json:"status" db:"status"`
	AuthorizationCode string                 `json:"authorization_code" db:"authorization_code"`
	ProcessorResponse string                 `json:"processor_response" db:"processor_response"`
	DeclineReason     string                 `json:"decline_reason" db:"decline_reason"`
	Location          TransactionLocation    `json:"location" db:"location"`
	PaymentMethod     PaymentMethod          `json:"payment_method" db:"payment_method"`
	IsRecurring       bool                   `json:"is_recurring" db:"is_recurring"`
	RiskScore         float64                `json:"risk_score" db:"risk_score"`
	FraudFlags        []string               `json:"fraud_flags" db:"fraud_flags"`
	Rewards           *TransactionRewards    `json:"rewards" db:"rewards"`
	AuthorizedAt      time.Time              `json:"authorized_at" db:"authorized_at"`
	SettledAt         *time.Time             `json:"settled_at" db:"settled_at"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypePurchase    TransactionType = "purchase"
	TransactionTypeWithdrawal  TransactionType = "withdrawal"
	TransactionTypeRefund      TransactionType = "refund"
	TransactionTypeReversal    TransactionType = "reversal"
	TransactionTypeAdjustment  TransactionType = "adjustment"
	TransactionTypeFee         TransactionType = "fee"
	TransactionTypeReward      TransactionType = "reward"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusAuthorized TransactionStatus = "authorized"
	TransactionStatusSettled   TransactionStatus = "settled"
	TransactionStatusDeclined  TransactionStatus = "declined"
	TransactionStatusReversed  TransactionStatus = "reversed"
	TransactionStatusRefunded  TransactionStatus = "refunded"
)

// TransactionLocation represents the location of a transaction
type TransactionLocation struct {
	Country   string  `json:"country" db:"country"`
	City      string  `json:"city" db:"city"`
	Latitude  float64 `json:"latitude" db:"latitude"`
	Longitude float64 `json:"longitude" db:"longitude"`
	Timezone  string  `json:"timezone" db:"timezone"`
}

// PaymentMethod represents the payment method used
type PaymentMethod string

const (
	PaymentMethodChip        PaymentMethod = "chip"
	PaymentMethodContactless PaymentMethod = "contactless"
	PaymentMethodMagStripe   PaymentMethod = "mag_stripe"
	PaymentMethodOnline      PaymentMethod = "online"
	PaymentMethodATM         PaymentMethod = "atm"
	PaymentMethodMobile      PaymentMethod = "mobile"
)

// TransactionRewards represents rewards earned from a transaction
type TransactionRewards struct {
	Amount       string  `json:"amount" db:"amount"`
	Currency     string  `json:"currency" db:"currency"`
	Rate         float64 `json:"rate" db:"rate"`
	Category     string  `json:"category" db:"category"`
	Multiplier   float64 `json:"multiplier" db:"multiplier"`
	EarnedAt     time.Time `json:"earned_at" db:"earned_at"`
}

// CreateCardRequest represents a request to create a card
type CreateCardRequest struct {
	AccountID        string                 `json:"account_id" validate:"required"`
	WalletID         string                 `json:"wallet_id" validate:"required"`
	CardType         CardType               `json:"card_type" validate:"required"`
	Currency         string                 `json:"currency" validate:"required"`
	HolderName       string                 `json:"holder_name" validate:"required"`
	BillingAddress   BillingAddress         `json:"billing_address" validate:"required"`
	ShippingAddress  *ShippingAddress       `json:"shipping_address,omitempty"`
	SpendingLimits   *SpendingLimits        `json:"spending_limits,omitempty"`
	SecuritySettings *SecuritySettings      `json:"security_settings,omitempty"`
	DesignID         string                 `json:"design_id,omitempty"`
	RewardsProgram   *RewardsProgram        `json:"rewards_program,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateCardRequest represents a request to update a card
type UpdateCardRequest struct {
	Status           *CardStatus            `json:"status,omitempty"`
	SpendingLimits   *SpendingLimits        `json:"spending_limits,omitempty"`
	SecuritySettings *SecuritySettings      `json:"security_settings,omitempty"`
	BillingAddress   *BillingAddress        `json:"billing_address,omitempty"`
	PIN              *string                `json:"pin,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ActivateCardRequest represents a request to activate a card
type ActivateCardRequest struct {
	CardID string `json:"card_id" validate:"required"`
	PIN    string `json:"pin" validate:"required,len=4"`
}

// BlockCardRequest represents a request to block a card
type BlockCardRequest struct {
	CardID string `json:"card_id" validate:"required"`
	Reason string `json:"reason" validate:"required"`
}

// UnblockCardRequest represents a request to unblock a card
type UnblockCardRequest struct {
	CardID string `json:"card_id" validate:"required"`
}

// LoadCardRequest represents a request to load funds onto a card
type LoadCardRequest struct {
	CardID   string                 `json:"card_id" validate:"required"`
	Amount   string                 `json:"amount" validate:"required"`
	Currency string                 `json:"currency" validate:"required"`
	Source   string                 `json:"source" validate:"required"` // wallet_id or external source
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CardListRequest represents a request to list cards
type CardListRequest struct {
	Page      int        `json:"page" validate:"min=1"`
	Limit     int        `json:"limit" validate:"min=1,max=100"`
	AccountID string     `json:"account_id,omitempty"`
	CardType  CardType   `json:"card_type,omitempty"`
	Status    CardStatus `json:"status,omitempty"`
	Currency  string     `json:"currency,omitempty"`
}

// CardListResponse represents a response to list cards
type CardListResponse struct {
	Cards      []Card `json:"cards"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	TotalPages int    `json:"total_pages"`
}

// TransactionListRequest represents a request to list transactions
type TransactionListRequest struct {
	Page             int               `json:"page" validate:"min=1"`
	Limit            int               `json:"limit" validate:"min=1,max=100"`
	CardID           string            `json:"card_id,omitempty"`
	AccountID        string            `json:"account_id,omitempty"`
	Status           TransactionStatus `json:"status,omitempty"`
	TransactionType  TransactionType   `json:"transaction_type,omitempty"`
	MerchantCategory string            `json:"merchant_category,omitempty"`
	DateFrom         *time.Time        `json:"date_from,omitempty"`
	DateTo           *time.Time        `json:"date_to,omitempty"`
	MinAmount        string            `json:"min_amount,omitempty"`
	MaxAmount        string            `json:"max_amount,omitempty"`
}

// TransactionListResponse represents a response to list transactions
type TransactionListResponse struct {
	Transactions []Transaction `json:"transactions"`
	Total        int           `json:"total"`
	Page         int           `json:"page"`
	Limit        int           `json:"limit"`
	TotalPages   int           `json:"total_pages"`
}
