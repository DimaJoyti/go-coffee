package domain

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// TokenType represents different types of tokens
type TokenType int32

const (
	TokenTypeUnknown TokenType = 0
	TokenTypeERC20   TokenType = 1
	TokenTypeERC721  TokenType = 2 // NFT
	TokenTypeERC1155 TokenType = 3 // Multi-token
	TokenTypeLoyalty TokenType = 4 // Loyalty token
	TokenTypeReward  TokenType = 5 // Reward token
	TokenTypeUtility TokenType = 6 // Utility token
)

// TokenStandard represents token standards
type TokenStandard string

const (
	TokenStandardERC20   TokenStandard = "ERC20"
	TokenStandardERC721  TokenStandard = "ERC721"
	TokenStandardERC1155 TokenStandard = "ERC1155"
)

// Token represents a cryptocurrency token
type Token struct {
	ID                string        `json:"id"`
	Address           string        `json:"address"`
	Network           NetworkType   `json:"network"`
	Type              TokenType     `json:"type"`
	Standard          TokenStandard `json:"standard"`
	Name              string        `json:"name"`
	Symbol            string        `json:"symbol"`
	Decimals          int32         `json:"decimals"`
	TotalSupply       string        `json:"total_supply"` // Raw amount in smallest unit
	CirculatingSupply string        `json:"circulating_supply,omitempty"`
	MaxSupply         string        `json:"max_supply,omitempty"`

	// Token metadata
	Description   string `json:"description,omitempty"`
	LogoURL       string `json:"logo_url,omitempty"`
	WebsiteURL    string `json:"website_url,omitempty"`
	WhitepaperURL string `json:"whitepaper_url,omitempty"`

	// Contract details
	ContractOwner string `json:"contract_owner,omitempty"`
	IsMintable    bool   `json:"is_mintable"`
	IsBurnable    bool   `json:"is_burnable"`
	IsPausable    bool   `json:"is_pausable"`
	IsUpgradeable bool   `json:"is_upgradeable"`

	// Business logic
	ExchangeRate   float64 `json:"exchange_rate,omitempty"`   // Rate to USD
	RewardRate     float64 `json:"reward_rate,omitempty"`     // Tokens per dollar spent
	MinimumBalance string  `json:"minimum_balance,omitempty"` // Minimum balance to hold

	// Timestamps
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	LaunchedAt *time.Time `json:"launched_at,omitempty"`

	// Additional data
	Tags     []string          `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// TokenBalance represents a user's token balance
type TokenBalance struct {
	ID               string            `json:"id"`
	UserID           string            `json:"user_id"`
	WalletID         string            `json:"wallet_id"`
	TokenID          string            `json:"token_id"`
	TokenAddress     string            `json:"token_address"`
	Network          NetworkType       `json:"network"`
	Balance          string            `json:"balance"`                   // Raw balance in smallest unit
	BalanceFormatted string            `json:"balance_formatted"`         // Human-readable balance
	LockedBalance    string            `json:"locked_balance,omitempty"`  // Locked/staked balance
	PendingBalance   string            `json:"pending_balance,omitempty"` // Pending transactions
	USDValue         float64           `json:"usd_value,omitempty"`
	LastUpdatedAt    time.Time         `json:"last_updated_at"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// TokenTransfer represents a token transfer event
type TokenTransfer struct {
	ID              string            `json:"id"`
	TokenID         string            `json:"token_id"`
	TokenAddress    string            `json:"token_address"`
	Network         NetworkType       `json:"network"`
	TransactionHash string            `json:"transaction_hash"`
	BlockNumber     int64             `json:"block_number"`
	LogIndex        int32             `json:"log_index"`
	FromAddress     string            `json:"from_address"`
	ToAddress       string            `json:"to_address"`
	Value           string            `json:"value"`           // Raw amount
	ValueFormatted  string            `json:"value_formatted"` // Human-readable amount
	Timestamp       time.Time         `json:"timestamp"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// LoyaltyProgram represents a loyalty program configuration
type LoyaltyProgram struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	TokenID      string      `json:"token_id"`
	TokenAddress string      `json:"token_address"`
	Network      NetworkType `json:"network"`

	// Earning rules
	EarnRate     float64 `json:"earn_rate"`     // Tokens per dollar spent
	MinimumSpend int64   `json:"minimum_spend"` // Minimum spend to earn (in cents)
	MaximumEarn  int64   `json:"maximum_earn"`  // Maximum tokens per transaction

	// Redemption rules
	RedemptionRate float64 `json:"redemption_rate"` // Dollars per token
	MinimumRedeem  int64   `json:"minimum_redeem"`  // Minimum tokens to redeem
	MaximumRedeem  int64   `json:"maximum_redeem"`  // Maximum tokens per redemption

	// Program settings
	IsActive  bool       `json:"is_active"`
	StartDate time.Time  `json:"start_date"`
	EndDate   *time.Time `json:"end_date,omitempty"`

	// Bonus multipliers
	BonusMultipliers map[string]float64 `json:"bonus_multipliers,omitempty"` // Category -> multiplier

	// Timestamps
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// NewToken creates a new token
func NewToken(address string, network NetworkType, tokenType TokenType, name, symbol string, decimals int32) (*Token, error) {
	if !IsValidAddress(address) {
		return nil, errors.New("invalid token address")
	}

	if name == "" {
		return nil, errors.New("token name is required")
	}

	if symbol == "" {
		return nil, errors.New("token symbol is required")
	}

	if decimals < 0 || decimals > 18 {
		return nil, errors.New("decimals must be between 0 and 18")
	}

	var standard TokenStandard
	switch tokenType {
	case TokenTypeERC20, TokenTypeLoyalty, TokenTypeReward, TokenTypeUtility:
		standard = TokenStandardERC20
	case TokenTypeERC721:
		standard = TokenStandardERC721
	case TokenTypeERC1155:
		standard = TokenStandardERC1155
	default:
		standard = TokenStandardERC20
	}

	return &Token{
		ID:            generateTokenID(),
		Address:       strings.ToLower(address),
		Network:       network,
		Type:          tokenType,
		Standard:      standard,
		Name:          name,
		Symbol:        symbol,
		Decimals:      decimals,
		TotalSupply:   "0",
		IsMintable:    true,
		IsBurnable:    true,
		IsPausable:    false,
		IsUpgradeable: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Tags:          make([]string, 0),
		Metadata:      make(map[string]string),
	}, nil
}

// SetTotalSupply sets the total supply
func (t *Token) SetTotalSupply(totalSupply string) error {
	// Validate that totalSupply is a valid number
	if _, ok := new(big.Int).SetString(totalSupply, 10); !ok {
		return errors.New("invalid total supply format")
	}

	t.TotalSupply = totalSupply
	t.UpdatedAt = time.Now()

	return nil
}

// SetExchangeRate sets the exchange rate to USD
func (t *Token) SetExchangeRate(rate float64) error {
	if rate < 0 {
		return errors.New("exchange rate cannot be negative")
	}

	t.ExchangeRate = rate
	t.UpdatedAt = time.Now()

	return nil
}

// SetRewardRate sets the reward rate (tokens per dollar spent)
func (t *Token) SetRewardRate(rate float64) error {
	if rate < 0 {
		return errors.New("reward rate cannot be negative")
	}

	t.RewardRate = rate
	t.UpdatedAt = time.Now()

	return nil
}

// AddTag adds a tag to the token
func (t *Token) AddTag(tag string) {
	for _, existingTag := range t.Tags {
		if existingTag == tag {
			return // Already exists
		}
	}

	t.Tags = append(t.Tags, tag)
	t.UpdatedAt = time.Now()
}

// RemoveTag removes a tag from the token
func (t *Token) RemoveTag(tag string) {
	for i, existingTag := range t.Tags {
		if existingTag == tag {
			t.Tags = append(t.Tags[:i], t.Tags[i+1:]...)
			t.UpdatedAt = time.Now()
			break
		}
	}
}

// SetLaunched marks the token as launched
func (t *Token) SetLaunched() {
	now := time.Now()
	t.LaunchedAt = &now
	t.UpdatedAt = now
}

// IsLaunched checks if the token is launched
func (t *Token) IsLaunched() bool {
	return t.LaunchedAt != nil
}

// NewTokenBalance creates a new token balance
func NewTokenBalance(userID, walletID, tokenID, tokenAddress string, network NetworkType) (*TokenBalance, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	if walletID == "" {
		return nil, errors.New("wallet ID is required")
	}

	if tokenID == "" {
		return nil, errors.New("token ID is required")
	}

	if !IsValidAddress(tokenAddress) {
		return nil, errors.New("invalid token address")
	}

	return &TokenBalance{
		ID:               generateTokenBalanceID(),
		UserID:           userID,
		WalletID:         walletID,
		TokenID:          tokenID,
		TokenAddress:     strings.ToLower(tokenAddress),
		Network:          network,
		Balance:          "0",
		BalanceFormatted: "0",
		LockedBalance:    "0",
		PendingBalance:   "0",
		LastUpdatedAt:    time.Now(),
		Metadata:         make(map[string]string),
	}, nil
}

// UpdateBalance updates the token balance
func (tb *TokenBalance) UpdateBalance(balance string, decimals int32) error {
	// Validate that balance is a valid number
	balanceBig, ok := new(big.Int).SetString(balance, 10)
	if !ok {
		return errors.New("invalid balance format")
	}

	tb.Balance = balance

	// Format the balance for human readability
	if decimals > 0 {
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
		quotient := new(big.Int).Div(balanceBig, divisor)
		remainder := new(big.Int).Mod(balanceBig, divisor)

		if remainder.Cmp(big.NewInt(0)) == 0 {
			tb.BalanceFormatted = quotient.String()
		} else {
			// Convert to float for formatting
			balanceFloat := new(big.Float).SetInt(balanceBig)
			divisorFloat := new(big.Float).SetInt(divisor)
			result := new(big.Float).Quo(balanceFloat, divisorFloat)
			tb.BalanceFormatted = result.Text('f', int(decimals))
		}
	} else {
		tb.BalanceFormatted = balance
	}

	tb.LastUpdatedAt = time.Now()

	return nil
}

// NewLoyaltyProgram creates a new loyalty program
func NewLoyaltyProgram(name, description, tokenID, tokenAddress string, network NetworkType) (*LoyaltyProgram, error) {
	if name == "" {
		return nil, errors.New("program name is required")
	}

	if tokenID == "" {
		return nil, errors.New("token ID is required")
	}

	if !IsValidAddress(tokenAddress) {
		return nil, errors.New("invalid token address")
	}

	return &LoyaltyProgram{
		ID:               generateLoyaltyProgramID(),
		Name:             name,
		Description:      description,
		TokenID:          tokenID,
		TokenAddress:     strings.ToLower(tokenAddress),
		Network:          network,
		EarnRate:         1.0,   // 1 token per dollar by default
		MinimumSpend:     100,   // $1.00 minimum
		MaximumEarn:      10000, // 100 tokens maximum per transaction
		RedemptionRate:   0.01,  // $0.01 per token
		MinimumRedeem:    100,   // 100 tokens minimum
		MaximumRedeem:    10000, // 100 tokens maximum per redemption
		IsActive:         true,
		StartDate:        time.Now(),
		BonusMultipliers: make(map[string]float64),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Metadata:         make(map[string]string),
	}, nil
}

// CalculateEarnedTokens calculates tokens earned for a purchase
func (lp *LoyaltyProgram) CalculateEarnedTokens(amountSpent int64, category string) int64 {
	if !lp.IsActive || amountSpent < lp.MinimumSpend {
		return 0
	}

	// Convert cents to dollars
	dollarsSpent := float64(amountSpent) / 100.0

	// Apply base earn rate
	tokensEarned := dollarsSpent * lp.EarnRate

	// Apply category bonus multiplier if exists
	if multiplier, exists := lp.BonusMultipliers[category]; exists {
		tokensEarned *= multiplier
	}

	// Convert to integer (tokens are typically whole numbers)
	tokensInt := int64(tokensEarned)

	// Apply maximum earn limit
	if tokensInt > lp.MaximumEarn {
		tokensInt = lp.MaximumEarn
	}

	return tokensInt
}

// CalculateRedemptionValue calculates the USD value of redeemed tokens
func (lp *LoyaltyProgram) CalculateRedemptionValue(tokensRedeemed int64) int64 {
	if !lp.IsActive || tokensRedeemed < lp.MinimumRedeem {
		return 0
	}

	// Apply maximum redeem limit
	if tokensRedeemed > lp.MaximumRedeem {
		tokensRedeemed = lp.MaximumRedeem
	}

	// Calculate USD value in cents
	usdValue := float64(tokensRedeemed) * lp.RedemptionRate * 100

	return int64(usdValue)
}

// SetBonusMultiplier sets a bonus multiplier for a category
func (lp *LoyaltyProgram) SetBonusMultiplier(category string, multiplier float64) error {
	if multiplier < 0 {
		return errors.New("multiplier cannot be negative")
	}

	lp.BonusMultipliers[category] = multiplier
	lp.UpdatedAt = time.Now()

	return nil
}

// Helper functions

// generateTokenID generates a unique token ID
func generateTokenID() string {
	return "token_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

// generateTokenBalanceID generates a unique token balance ID
func generateTokenBalanceID() string {
	return "balance_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

// generateLoyaltyProgramID generates a unique loyalty program ID
func generateLoyaltyProgramID() string {
	return "loyalty_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

// Token Factory Functions

// NewLoyaltyToken creates a new loyalty token
func NewLoyaltyToken(address string, network NetworkType, name, symbol string) (*Token, error) {
	token, err := NewToken(address, network, TokenTypeLoyalty, name, symbol, 18)
	if err != nil {
		return nil, err
	}

	token.Description = "Loyalty token for Go Coffee rewards program"
	token.AddTag("loyalty")
	token.AddTag("rewards")
	if err := token.SetRewardRate(1.0); err != nil { // 1 token per dollar spent
		return nil, fmt.Errorf("failed to set reward rate: %w", err)
	}

	return token, nil
}

// NewCoffeeToken creates the main Go Coffee token
func NewCoffeeToken(address string, network NetworkType) (*Token, error) {
	token, err := NewToken(address, network, TokenTypeUtility, "Go Coffee Token", "COFFEE", 18)
	if err != nil {
		return nil, err
	}

	token.Description = "The official utility token of the Go Coffee ecosystem"
	token.AddTag("utility")
	token.AddTag("coffee")
	token.AddTag("defi")

	return token, nil
}
