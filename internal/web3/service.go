package web3

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/web3/payment"
	"github.com/DimaJoyti/go-coffee/pkg/config"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// Service provides Web3 and DeFi operations for the coffee platform
type Service struct {
	paymentProcessor *payment.Processor
	logger           *zap.Logger
	config           config.Web3Config

	// State management
	mutex   sync.RWMutex
	running bool

	// Payment tracking
	payments map[string]*Payment
	orders   map[string]*CoffeeOrder
}

// Payment represents a crypto payment for coffee
type Payment struct {
	ID              string                 `json:"id"`
	OrderID         string                 `json:"order_id"`
	CustomerAddress string                 `json:"customer_address"`
	Amount          decimal.Decimal        `json:"amount"`
	Currency        string                 `json:"currency"`
	Chain           string                 `json:"chain"`
	Status          PaymentStatus          `json:"status"`
	TransactionHash string                 `json:"transaction_hash,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	ExpiresAt       time.Time              `json:"expires_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// Implement payment.Payment interface
func (p *Payment) GetID() string {
	return p.ID
}

func (p *Payment) GetChain() string {
	return p.Chain
}

func (p *Payment) GetCurrency() string {
	return p.Currency
}

func (p *Payment) GetAmount() decimal.Decimal {
	return p.Amount
}

func (p *Payment) GetCustomerAddress() string {
	return p.CustomerAddress
}

func (p *Payment) SetStatus(status string) {
	p.Status = PaymentStatus(status)
}

func (p *Payment) SetTransactionHash(hash string) {
	p.TransactionHash = hash
}

// CoffeeOrder represents a coffee order with crypto payment
type CoffeeOrder struct {
	ID           string          `json:"id"`
	CustomerName string          `json:"customer_name"`
	CoffeeType   string          `json:"coffee_type"`
	Quantity     int             `json:"quantity"`
	TotalAmount  decimal.Decimal `json:"total_amount"`
	Currency     string          `json:"currency"`
	PaymentID    string          `json:"payment_id,omitempty"`
	Status       OrderStatus     `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// PaymentStatus represents the status of a crypto payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusConfirmed PaymentStatus = "confirmed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusExpired   PaymentStatus = "expired"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

// OrderStatus represents the status of a coffee order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusPaid       OrderStatus = "paid"
	OrderStatusPreparing  OrderStatus = "preparing"
	OrderStatusReady      OrderStatus = "ready"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

// PaymentRequest represents a request to create a crypto payment
type PaymentRequest struct {
	OrderID         string                 `json:"order_id"`
	CustomerAddress string                 `json:"customer_address"`
	Amount          decimal.Decimal        `json:"amount"`
	Currency        string                 `json:"currency"`
	Chain           string                 `json:"chain"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// PaymentResponse represents the response after creating a payment
type PaymentResponse struct {
	Payment         *Payment `json:"payment"`
	PaymentAddress  string   `json:"payment_address"`
	QRCode          string   `json:"qr_code,omitempty"`
	ExpirationTime  int64    `json:"expiration_time"`
	EstimatedGasFee string   `json:"estimated_gas_fee,omitempty"`
}

// NewService creates a new Web3 service
func NewService(
	paymentProcessor *payment.Processor,
	logger *zap.Logger,
	config config.Web3Config,
) *Service {
	return &Service{
		paymentProcessor: paymentProcessor,
		logger:           logger,
		config:           config,
		payments:         make(map[string]*Payment),
		orders:           make(map[string]*CoffeeOrder),
	}
}

// Start starts the Web3 service
func (s *Service) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return fmt.Errorf("service is already running")
	}

	s.logger.Info("Starting Web3 service...")

	// Start payment processor
	if err := s.paymentProcessor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start payment processor: %w", err)
	}

	// Start background tasks
	go s.monitorPayments(ctx)
	go s.cleanupExpiredPayments(ctx)

	s.running = true
	s.logger.Info("Web3 service started successfully")

	return nil
}

// Stop stops the Web3 service
func (s *Service) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return
	}

	s.logger.Info("Stopping Web3 service...")

	// Stop payment processor
	s.paymentProcessor.Stop()

	s.running = false
	s.logger.Info("Web3 service stopped")
}

// CreatePayment creates a new crypto payment for a coffee order
func (s *Service) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Validate request
	if err := s.validatePaymentRequest(req); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}

	// Create payment
	payment := &Payment{
		ID:              uuid.New().String(),
		OrderID:         req.OrderID,
		CustomerAddress: req.CustomerAddress,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Chain:           req.Chain,
		Status:          PaymentStatusPending,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		ExpiresAt:       time.Now().Add(15 * time.Minute), // 15 minutes to pay
		Metadata:        req.Metadata,
	}

	// Generate payment address
	paymentAddress, err := s.paymentProcessor.GeneratePaymentAddress(ctx, req.Chain, req.Currency)
	if err != nil {
		return nil, fmt.Errorf("failed to generate payment address: %w", err)
	}

	// Store payment
	s.payments[payment.ID] = payment

	// Generate QR code for payment
	qrCode, err := s.generatePaymentQRCode(payment, paymentAddress)
	if err != nil {
		s.logger.Warn("Failed to generate QR code", zap.Error(err))
	}

	// Estimate gas fee
	gasFee, err := s.paymentProcessor.EstimateGasFee(ctx, req.Chain, req.Currency)
	if err != nil {
		s.logger.Warn("Failed to estimate gas fee", zap.Error(err))
	}

	s.logger.Info("Payment created",
		zap.String("payment_id", payment.ID),
		zap.String("order_id", req.OrderID),
		zap.String("chain", req.Chain),
		zap.String("currency", req.Currency),
		zap.String("amount", req.Amount.String()),
	)

	return &PaymentResponse{
		Payment:         payment,
		PaymentAddress:  paymentAddress,
		QRCode:          qrCode,
		ExpirationTime:  payment.ExpiresAt.Unix(),
		EstimatedGasFee: gasFee,
	}, nil
}

// GetPaymentStatus retrieves the status of a payment
func (s *Service) GetPaymentStatus(ctx context.Context, paymentID string) (*Payment, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	payment, exists := s.payments[paymentID]
	if !exists {
		return nil, fmt.Errorf("payment not found: %s", paymentID)
	}

	// Check for updates from blockchain
	if payment.Status == PaymentStatusPending {
		updated, err := s.paymentProcessor.CheckPaymentStatus(ctx, payment)
		if err != nil {
			s.logger.Error("Failed to check payment status", zap.Error(err))
		} else if updated {
			payment.UpdatedAt = time.Now()
		}
	}

	return payment, nil
}

// ConfirmPayment manually confirms a payment (for testing or admin purposes)
func (s *Service) ConfirmPayment(ctx context.Context, paymentID, transactionHash string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	payment, exists := s.payments[paymentID]
	if !exists {
		return fmt.Errorf("payment not found: %s", paymentID)
	}

	if payment.Status != PaymentStatusPending {
		return fmt.Errorf("payment is not in pending status: %s", payment.Status)
	}

	// Verify transaction on blockchain
	verified, err := s.paymentProcessor.VerifyTransaction(ctx, payment.Chain, transactionHash, payment.Amount, payment.Currency)
	if err != nil {
		return fmt.Errorf("failed to verify transaction: %w", err)
	}

	if !verified {
		return fmt.Errorf("transaction verification failed")
	}

	// Update payment status
	payment.Status = PaymentStatusConfirmed
	payment.TransactionHash = transactionHash
	payment.UpdatedAt = time.Now()

	s.logger.Info("Payment confirmed",
		zap.String("payment_id", paymentID),
		zap.String("transaction_hash", transactionHash),
	)

	return nil
}

// CancelPayment cancels a pending payment
func (s *Service) CancelPayment(ctx context.Context, paymentID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	payment, exists := s.payments[paymentID]
	if !exists {
		return fmt.Errorf("payment not found: %s", paymentID)
	}

	if payment.Status != PaymentStatusPending {
		return fmt.Errorf("payment cannot be cancelled, current status: %s", payment.Status)
	}

	payment.Status = PaymentStatusCancelled
	payment.UpdatedAt = time.Now()

	s.logger.Info("Payment cancelled", zap.String("payment_id", paymentID))

	return nil
}

// validatePaymentRequest validates a payment request
func (s *Service) validatePaymentRequest(req *PaymentRequest) error {
	if req.OrderID == "" {
		return fmt.Errorf("order_id is required")
	}
	if req.CustomerAddress == "" {
		return fmt.Errorf("customer_address is required")
	}
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount must be greater than zero")
	}
	if req.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if req.Chain == "" {
		return fmt.Errorf("chain is required")
	}

	// Validate supported chains and currencies
	if !s.isSupportedChain(req.Chain) {
		return fmt.Errorf("unsupported chain: %s", req.Chain)
	}
	if !s.isSupportedCurrency(req.Currency) {
		return fmt.Errorf("unsupported currency: %s", req.Currency)
	}

	return nil
}

// isSupportedChain checks if a blockchain is supported
func (s *Service) isSupportedChain(chain string) bool {
	supportedChains := []string{"ethereum", "bsc", "polygon", "solana"}
	for _, supported := range supportedChains {
		if chain == supported {
			return true
		}
	}
	return false
}

// isSupportedCurrency checks if a currency is supported
func (s *Service) isSupportedCurrency(currency string) bool {
	supportedCurrencies := []string{"ETH", "BNB", "MATIC", "SOL", "USDC", "USDT", "COFFEE"}
	for _, supported := range supportedCurrencies {
		if currency == supported {
			return true
		}
	}
	return false
}

// generatePaymentQRCode generates a QR code for payment
func (s *Service) generatePaymentQRCode(payment *Payment, address string) (string, error) {
	// This would generate a QR code with payment details
	// For now, return a placeholder
	return fmt.Sprintf("qr_code_for_payment_%s", payment.ID), nil
}

// monitorPayments monitors payment status in the background
func (s *Service) monitorPayments(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkPendingPayments(ctx)
		}
	}
}

// checkPendingPayments checks the status of pending payments
func (s *Service) checkPendingPayments(ctx context.Context) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, payment := range s.payments {
		if payment.Status == PaymentStatusPending {
			updated, err := s.paymentProcessor.CheckPaymentStatus(ctx, payment)
			if err != nil {
				s.logger.Error("Failed to check payment status",
					zap.String("payment_id", payment.ID),
					zap.Error(err),
				)
				continue
			}
			if updated {
				payment.UpdatedAt = time.Now()
				s.logger.Info("Payment status updated",
					zap.String("payment_id", payment.ID),
					zap.String("status", string(payment.Status)),
				)
			}
		}
	}
}

// cleanupExpiredPayments removes expired payments
func (s *Service) cleanupExpiredPayments(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.expireOldPayments()
		}
	}
}

// expireOldPayments marks expired payments as expired
func (s *Service) expireOldPayments() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	for _, payment := range s.payments {
		if payment.Status == PaymentStatusPending && now.After(payment.ExpiresAt) {
			payment.Status = PaymentStatusExpired
			payment.UpdatedAt = now
			s.logger.Info("Payment expired", zap.String("payment_id", payment.ID))
		}
	}
}
