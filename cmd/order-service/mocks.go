package main

import (
	"context"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/order/application"
	"github.com/DimaJoyti/go-coffee/internal/order/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// MockEventPublisher implements EventPublisher interface
type MockEventPublisher struct {
	logger *logger.Logger
}

func (m *MockEventPublisher) Publish(ctx context.Context, event *domain.DomainEvent) error {
	m.logger.WithFields(map[string]interface{}{
		"event_type":    event.Type,
		"aggregate_id":  event.AggregateID,
		"event_id":      event.ID,
	}).Info("Event published")
	return nil
}

func (m *MockEventPublisher) PublishBatch(ctx context.Context, events []*domain.DomainEvent) error {
	for _, event := range events {
		if err := m.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// MockKitchenService implements KitchenService interface
type MockKitchenService struct {
	logger *logger.Logger
}

func (m *MockKitchenService) SubmitOrder(ctx context.Context, order *domain.Order) error {
	m.logger.WithFields(map[string]interface{}{
		"order_id":    order.ID,
		"customer_id": order.CustomerID,
		"item_count":  len(order.Items),
	}).Info("Order submitted to kitchen")
	return nil
}

func (m *MockKitchenService) UpdateOrderStatus(ctx context.Context, orderID string, status domain.OrderStatus) error {
	m.logger.WithFields(map[string]interface{}{
		"order_id": orderID,
		"status":   status.String(),
	}).Info("Order status updated in kitchen")
	return nil
}

func (m *MockKitchenService) GetEstimatedTime(ctx context.Context, items []*domain.OrderItem) (int32, error) {
	// Simple estimation: 2 minutes per item + 3 minutes base time
	estimatedTime := int32(180 + len(items)*120) // in seconds
	
	m.logger.WithFields(map[string]interface{}{
		"item_count":     len(items),
		"estimated_time": estimatedTime,
	}).Info("Estimated preparation time calculated")
	
	return estimatedTime, nil
}

// MockAuthService implements AuthService interface
type MockAuthService struct {
	logger *logger.Logger
}

func (m *MockAuthService) ValidateUser(ctx context.Context, userID string) (*application.UserInfo, error) {
	m.logger.WithField("user_id", userID).Info("User validated")
	
	return &application.UserInfo{
		UserID:    userID,
		Email:     "user@example.com",
		Role:      "customer",
		Status:    "active",
		CreatedAt: time.Now().Add(-30 * 24 * time.Hour), // 30 days ago
	}, nil
}

func (m *MockAuthService) GetUserPreferences(ctx context.Context, userID string) (*application.UserPreferences, error) {
	m.logger.WithField("user_id", userID).Info("User preferences retrieved")
	
	return &application.UserPreferences{
		UserID:              userID,
		DefaultPaymentMethod: domain.PaymentMethodCreditCard,
		Preferences: map[string]string{
			"notification_email": "true",
			"notification_sms":   "false",
			"default_size":       "medium",
		},
	}, nil
}

// MockPaymentProcessor implements PaymentProcessor interface
type MockPaymentProcessor struct {
	logger *logger.Logger
}

func (m *MockPaymentProcessor) ProcessPayment(ctx context.Context, payment *domain.Payment) (*application.ProcessPaymentResult, error) {
	m.logger.WithFields(map[string]interface{}{
		"payment_id":     payment.ID,
		"amount":         payment.Amount,
		"payment_method": payment.PaymentMethod,
	}).Info("Payment processed")
	
	return &application.ProcessPaymentResult{
		ProcessorID:  "mock-processor",
		ProcessorRef: "mock_" + payment.ID,
		Success:      true,
		Message:      "Payment processed successfully",
	}, nil
}

func (m *MockPaymentProcessor) RefundPayment(ctx context.Context, payment *domain.Payment, amount int64, reason string) (*application.RefundResult, error) {
	m.logger.WithFields(map[string]interface{}{
		"payment_id": payment.ID,
		"amount":     amount,
		"reason":     reason,
	}).Info("Payment refunded")
	
	return &application.RefundResult{
		RefundRef: "refund_" + payment.ID,
		Success:   true,
		Message:   "Refund processed successfully",
	}, nil
}

func (m *MockPaymentProcessor) GetPaymentStatus(ctx context.Context, processorRef string) (*application.PaymentStatusResult, error) {
	m.logger.WithField("processor_ref", processorRef).Info("Payment status retrieved")
	
	return &application.PaymentStatusResult{
		Status:    "completed",
		Reference: processorRef,
		UpdatedAt: time.Now(),
	}, nil
}

// MockCryptoProcessor implements CryptoPaymentProcessor interface
type MockCryptoProcessor struct {
	logger *logger.Logger
}

func (m *MockCryptoProcessor) CreatePaymentAddress(ctx context.Context, payment *domain.Payment) (*application.CryptoAddressResult, error) {
	m.logger.WithFields(map[string]interface{}{
		"payment_id":     payment.ID,
		"crypto_network": payment.CryptoNetwork,
		"crypto_token":   payment.CryptoToken,
	}).Info("Crypto payment address created")
	
	expiresAt := time.Now().Add(30 * time.Minute)
	
	return &application.CryptoAddressResult{
		Address:   "0x1234567890abcdef1234567890abcdef12345678",
		Network:   "ethereum",
		ExpiresAt: &expiresAt,
	}, nil
}

func (m *MockCryptoProcessor) VerifyPayment(ctx context.Context, txHash string, network domain.CryptoNetwork) (*application.CryptoVerificationResult, error) {
	m.logger.WithFields(map[string]interface{}{
		"tx_hash": txHash,
		"network": network,
	}).Info("Crypto payment verified")
	
	return &application.CryptoVerificationResult{
		IsValid:       true,
		Amount:        1000000, // 1 USDC (6 decimals)
		BlockNumber:   18500000,
		GasUsed:       21000,
		GasPrice:      20000000000, // 20 gwei
		Confirmations: 12,
	}, nil
}

func (m *MockCryptoProcessor) GetTransactionStatus(ctx context.Context, txHash string, network domain.CryptoNetwork) (*application.CryptoTransactionStatus, error) {
	m.logger.WithFields(map[string]interface{}{
		"tx_hash": txHash,
		"network": network,
	}).Info("Crypto transaction status retrieved")
	
	return &application.CryptoTransactionStatus{
		Status:        "confirmed",
		Confirmations: 12,
		BlockNumber:   18500000,
		IsConfirmed:   true,
	}, nil
}

// MockLoyaltyService implements LoyaltyService interface
type MockLoyaltyService struct {
	logger *logger.Logger
}

func (m *MockLoyaltyService) GetTokenBalance(ctx context.Context, customerID string) (int64, error) {
	m.logger.WithField("customer_id", customerID).Info("Token balance retrieved")
	
	// Mock balance: 1000 tokens
	return 1000, nil
}

func (m *MockLoyaltyService) RedeemTokens(ctx context.Context, customerID string, amount int64) error {
	m.logger.WithFields(map[string]interface{}{
		"customer_id": customerID,
		"amount":      amount,
	}).Info("Tokens redeemed")
	
	return nil
}

func (m *MockLoyaltyService) EarnTokens(ctx context.Context, customerID string, amount int64, reason string) error {
	m.logger.WithFields(map[string]interface{}{
		"customer_id": customerID,
		"amount":      amount,
		"reason":      reason,
	}).Info("Tokens earned")
	
	return nil
}

func (m *MockLoyaltyService) GetExchangeRate(ctx context.Context) (float64, error) {
	m.logger.Info("Exchange rate retrieved")
	
	// Mock exchange rate: 100 tokens per dollar
	return 100.0, nil
}
