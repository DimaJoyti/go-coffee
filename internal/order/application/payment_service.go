package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/DimaJoyti/go-coffee/internal/order/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// PaymentProcessor defines the interface for payment processing
type PaymentProcessor interface {
	ProcessPayment(ctx context.Context, payment *domain.Payment) (*ProcessPaymentResult, error)
	RefundPayment(ctx context.Context, payment *domain.Payment, amount int64, reason string) (*RefundResult, error)
	GetPaymentStatus(ctx context.Context, processorRef string) (*PaymentStatusResult, error)
}

// CryptoPaymentProcessor defines the interface for cryptocurrency payment processing
type CryptoPaymentProcessor interface {
	CreatePaymentAddress(ctx context.Context, payment *domain.Payment) (*CryptoAddressResult, error)
	VerifyPayment(ctx context.Context, txHash string, network domain.CryptoNetwork) (*CryptoVerificationResult, error)
	GetTransactionStatus(ctx context.Context, txHash string, network domain.CryptoNetwork) (*CryptoTransactionStatus, error)
}

// LoyaltyService defines the interface for loyalty token operations
type LoyaltyService interface {
	GetTokenBalance(ctx context.Context, customerID string) (int64, error)
	RedeemTokens(ctx context.Context, customerID string, amount int64) error
	EarnTokens(ctx context.Context, customerID string, amount int64, reason string) error
	GetExchangeRate(ctx context.Context) (float64, error) // tokens per dollar
}

// PaymentService implements payment processing use cases
type PaymentService struct {
	paymentRepo      PaymentRepository
	orderRepo        OrderRepository
	eventPublisher   EventPublisher
	paymentProcessor PaymentProcessor
	cryptoProcessor  CryptoPaymentProcessor
	loyaltyService   LoyaltyService
	logger           *logger.Logger
}

// NewPaymentService creates a new payment service
func NewPaymentService(
	paymentRepo PaymentRepository,
	orderRepo OrderRepository,
	eventPublisher EventPublisher,
	paymentProcessor PaymentProcessor,
	cryptoProcessor CryptoPaymentProcessor,
	loyaltyService LoyaltyService,
	logger *logger.Logger,
) *PaymentService {
	return &PaymentService{
		paymentRepo:      paymentRepo,
		orderRepo:        orderRepo,
		eventPublisher:   eventPublisher,
		paymentProcessor: paymentProcessor,
		cryptoProcessor:  cryptoProcessor,
		loyaltyService:   loyaltyService,
		logger:           logger,
	}
}

// CreatePayment creates a new payment for an order
func (s *PaymentService) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// Get the order
	order, err := s.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Validate order status
	if order.Status != domain.OrderStatusPending && order.Status != domain.OrderStatusConfirmed {
		return nil, errors.New("order is not in a payable status")
	}

	// Create payment
	payment, err := domain.NewPayment(req.OrderID, order.CustomerID, order.TotalAmount, order.Currency, req.PaymentMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Handle different payment methods
	switch req.PaymentMethod {
	case domain.PaymentMethodCrypto:
		return s.handleCryptoPayment(ctx, payment, req)
	case domain.PaymentMethodLoyaltyToken:
		return s.handleLoyaltyTokenPayment(ctx, payment, req)
	case domain.PaymentMethodCreditCard, domain.PaymentMethodDebitCard:
		return s.handleCardPayment(ctx, payment, req)
	default:
		return s.handleTraditionalPayment(ctx, payment, req)
	}
}

// ProcessPayment processes a payment
func (s *PaymentService) ProcessPayment(ctx context.Context, req *ProcessPaymentRequest) (*ProcessPaymentResponse, error) {
	payment, err := s.paymentRepo.GetByID(ctx, req.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	// Check if payment can be processed
	if !payment.CanTransitionTo(domain.PaymentStatusProcessing) {
		return nil, errors.New("payment cannot be processed in current status")
	}

	// Update payment status to processing
	previousStatus := payment.Status
	if err := payment.UpdateStatus(domain.PaymentStatusProcessing); err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	// Save payment
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	// Publish payment processing event
	event := domain.NewPaymentStatusChangedEvent(payment, previousStatus)
	if err := s.eventPublisher.Publish(ctx, event); err != nil {
		s.logger.WithError(err).Error("Failed to publish payment processing event")
	}

	// Process payment based on method
	var result *ProcessPaymentResult
	if payment.IsCryptoPayment() {
		result, err = s.processCryptoPayment(ctx, payment)
	} else {
		result, err = s.paymentProcessor.ProcessPayment(ctx, payment)
	}

	if err != nil {
		// Update payment status to failed
		payment.SetFailureReason(err.Error())
		if updateErr := payment.UpdateStatus(domain.PaymentStatusFailed); updateErr == nil {
			s.paymentRepo.Update(ctx, payment)
		}
		return nil, fmt.Errorf("payment processing failed: %w", err)
	}

	// Update payment with processor details
	payment.SetProcessorDetails(result.ProcessorID, result.ProcessorRef)

	// Update payment status to completed
	if err := payment.UpdateStatus(domain.PaymentStatusCompleted); err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	// Save payment
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	// Publish payment completed event
	event = domain.NewPaymentStatusChangedEvent(payment, domain.PaymentStatusProcessing)
	if err := s.eventPublisher.Publish(ctx, event); err != nil {
		s.logger.WithError(err).Error("Failed to publish payment completed event")
	}

	s.logger.WithFields(map[string]any{
		"payment_id": payment.ID,
		"order_id":   payment.OrderID,
		"amount":     payment.Amount,
	}).Info("Payment processed successfully")

	return &ProcessPaymentResponse{
		PaymentID:    payment.ID,
		Status:       payment.Status.String(),
		ProcessorRef: payment.ProcessorRef,
		UpdatedAt:    payment.UpdatedAt,
	}, nil
}

// RefundPayment processes a payment refund
func (s *PaymentService) RefundPayment(ctx context.Context, req *RefundPaymentRequest) (*RefundPaymentResponse, error) {
	payment, err := s.paymentRepo.GetByID(ctx, req.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	// Check if payment can be refunded
	if payment.Status != domain.PaymentStatusCompleted {
		return nil, errors.New("only completed payments can be refunded")
	}

	// Create refund
	refund, err := domain.NewRefund(payment.ID, payment.OrderID, req.Amount, payment.Currency, req.Reason)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	// Process refund
	refundResult, err := s.paymentProcessor.RefundPayment(ctx, payment, req.Amount, req.Reason)
	if err != nil {
		return nil, fmt.Errorf("refund processing failed: %w", err)
	}

	// Update refund with processor details
	refund.ProcessorRef = refundResult.RefundRef
	refund.Status = domain.PaymentStatusCompleted

	// Update payment status to refunded if full refund
	if req.Amount >= payment.Amount {
		if err := payment.UpdateStatus(domain.PaymentStatusRefunded); err != nil {
			return nil, fmt.Errorf("failed to update payment status: %w", err)
		}

		if err := s.paymentRepo.Update(ctx, payment); err != nil {
			return nil, fmt.Errorf("failed to save payment: %w", err)
		}
	}

	s.logger.WithFields(map[string]any{
		"payment_id": payment.ID,
		"refund_id":  refund.ID,
		"amount":     req.Amount,
		"reason":     req.Reason,
	}).Info("Payment refunded successfully")

	return &RefundPaymentResponse{
		RefundID:     refund.ID,
		PaymentID:    payment.ID,
		Amount:       refund.Amount,
		Status:       refund.Status.String(),
		ProcessorRef: refund.ProcessorRef,
		CreatedAt:    refund.CreatedAt,
	}, nil
}

// Helper methods for different payment types

func (s *PaymentService) handleCryptoPayment(ctx context.Context, payment *domain.Payment, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// Create payment address for crypto payment
	addressResult, err := s.cryptoProcessor.CreatePaymentAddress(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to create crypto payment address: %w", err)
	}

	// Set crypto details
	payment.SetCryptoDetails(req.CryptoNetwork, req.CryptoToken, addressResult.Address, "")

	// Save payment
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	// Publish payment created event
	event := domain.NewPaymentCreatedEvent(payment)
	if err := s.eventPublisher.Publish(ctx, event); err != nil {
		s.logger.WithError(err).Error("Failed to publish payment created event")
	}

	return &CreatePaymentResponse{
		PaymentID:      payment.ID,
		Status:         payment.Status.String(),
		Amount:         payment.Amount,
		Currency:       payment.Currency,
		PaymentAddress: addressResult.Address,
		ExpiresAt:      addressResult.ExpiresAt,
		CreatedAt:      payment.CreatedAt,
	}, nil
}

func (s *PaymentService) handleLoyaltyTokenPayment(ctx context.Context, payment *domain.Payment, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	_ = req // Parameter reserved for future use

	// Get exchange rate
	exchangeRate, err := s.loyaltyService.GetExchangeRate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	// Calculate tokens needed
	tokensNeeded := int64(float64(payment.Amount) * exchangeRate / 100) // Convert cents to dollars

	// Check token balance
	balance, err := s.loyaltyService.GetTokenBalance(ctx, payment.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token balance: %w", err)
	}

	if balance < tokensNeeded {
		return nil, errors.New("insufficient loyalty tokens")
	}

	// Redeem tokens
	if err := s.loyaltyService.RedeemTokens(ctx, payment.CustomerID, tokensNeeded); err != nil {
		return nil, fmt.Errorf("failed to redeem tokens: %w", err)
	}

	// Update payment status to completed
	if err := payment.UpdateStatus(domain.PaymentStatusCompleted); err != nil {
		return nil, fmt.Errorf("failed to update payment status: %w", err)
	}

	// Save payment
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	// Publish events
	events := []*domain.DomainEvent{
		domain.NewPaymentCreatedEvent(payment),
		domain.NewPaymentStatusChangedEvent(payment, domain.PaymentStatusPending),
		domain.NewLoyaltyRedeemedEvent(payment.CustomerID, payment.OrderID, tokensNeeded, payment.Amount),
	}

	if err := s.eventPublisher.PublishBatch(ctx, events); err != nil {
		s.logger.WithError(err).Error("Failed to publish payment events")
	}

	return &CreatePaymentResponse{
		PaymentID:    payment.ID,
		Status:       payment.Status.String(),
		Amount:       payment.Amount,
		Currency:     payment.Currency,
		TokensUsed:   tokensNeeded,
		ExchangeRate: exchangeRate,
		CreatedAt:    payment.CreatedAt,
	}, nil
}

func (s *PaymentService) handleCardPayment(ctx context.Context, payment *domain.Payment, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// Set card details
	payment.SetCardDetails(req.CardLast4, req.CardBrand)

	// Save payment
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	// Publish payment created event
	event := domain.NewPaymentCreatedEvent(payment)
	if err := s.eventPublisher.Publish(ctx, event); err != nil {
		s.logger.WithError(err).Error("Failed to publish payment created event")
	}

	return &CreatePaymentResponse{
		PaymentID: payment.ID,
		Status:    payment.Status.String(),
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		CreatedAt: payment.CreatedAt,
	}, nil
}

func (s *PaymentService) handleTraditionalPayment(ctx context.Context, payment *domain.Payment, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	_ = req // Parameter reserved for future use

	// Save payment
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	// Publish payment created event
	event := domain.NewPaymentCreatedEvent(payment)
	if err := s.eventPublisher.Publish(ctx, event); err != nil {
		s.logger.WithError(err).Error("Failed to publish payment created event")
	}

	return &CreatePaymentResponse{
		PaymentID: payment.ID,
		Status:    payment.Status.String(),
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		CreatedAt: payment.CreatedAt,
	}, nil
}

func (s *PaymentService) processCryptoPayment(ctx context.Context, payment *domain.Payment) (*ProcessPaymentResult, error) {
	// Verify the crypto transaction
	verificationResult, err := s.cryptoProcessor.VerifyPayment(ctx, payment.TransactionHash, payment.CryptoNetwork)
	if err != nil {
		return nil, fmt.Errorf("failed to verify crypto payment: %w", err)
	}

	if !verificationResult.IsValid {
		return nil, errors.New("crypto payment verification failed")
	}

	// Update payment with blockchain details
	payment.BlockNumber = verificationResult.BlockNumber
	payment.GasUsed = verificationResult.GasUsed
	payment.GasPrice = verificationResult.GasPrice

	return &ProcessPaymentResult{
		ProcessorID:  "crypto-processor",
		ProcessorRef: payment.TransactionHash,
		Success:      true,
	}, nil
}
