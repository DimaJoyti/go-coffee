package application

import (
	"context"
	"testing"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/order/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// Mock implementations for testing

type mockPaymentRepository struct {
	payments map[string]*domain.Payment
}

func newMockPaymentRepository() *mockPaymentRepository {
	return &mockPaymentRepository{
		payments: make(map[string]*domain.Payment),
	}
}

func (m *mockPaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	m.payments[payment.ID] = payment
	return nil
}

func (m *mockPaymentRepository) GetByID(ctx context.Context, paymentID string) (*domain.Payment, error) {
	if payment, exists := m.payments[paymentID]; exists {
		return payment, nil
	}
	return nil, domain.ErrPaymentNotFound
}

func (m *mockPaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	for _, payment := range m.payments {
		if payment.OrderID == orderID {
			return payment, nil
		}
	}
	return nil, domain.ErrPaymentNotFound
}

func (m *mockPaymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	m.payments[payment.ID] = payment
	return nil
}

func (m *mockPaymentRepository) List(ctx context.Context, filters PaymentFilters) ([]*domain.Payment, error) {
	var result []*domain.Payment
	for _, payment := range m.payments {
		result = append(result, payment)
	}
	return result, nil
}

type mockOrderRepository struct {
	orders map[string]*domain.Order
}

func newMockOrderRepository() *mockOrderRepository {
	return &mockOrderRepository{
		orders: make(map[string]*domain.Order),
	}
}

func (m *mockOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	m.orders[order.ID] = order
	return nil
}

func (m *mockOrderRepository) GetByID(ctx context.Context, orderID string) (*domain.Order, error) {
	if order, exists := m.orders[orderID]; exists {
		return order, nil
	}
	return nil, domain.ErrOrderNotFound
}

func (m *mockOrderRepository) GetByCustomerID(ctx context.Context, customerID string, limit, offset int) ([]*domain.Order, error) {
	var result []*domain.Order
	for _, order := range m.orders {
		if order.CustomerID == customerID {
			result = append(result, order)
		}
	}
	return result, nil
}

func (m *mockOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	m.orders[order.ID] = order
	return nil
}

func (m *mockOrderRepository) Delete(ctx context.Context, orderID string) error {
	delete(m.orders, orderID)
	return nil
}

func (m *mockOrderRepository) List(ctx context.Context, filters OrderFilters) ([]*domain.Order, error) {
	var result []*domain.Order
	for _, order := range m.orders {
		result = append(result, order)
	}
	return result, nil
}

type mockEventPublisher struct{}

func (m *mockEventPublisher) Publish(ctx context.Context, event *domain.DomainEvent) error {
	return nil
}

func (m *mockEventPublisher) PublishBatch(ctx context.Context, events []*domain.DomainEvent) error {
	return nil
}

type mockPaymentProcessor struct{}

func (m *mockPaymentProcessor) ProcessPayment(ctx context.Context, payment *domain.Payment) (*ProcessPaymentResult, error) {
	return &ProcessPaymentResult{
		ProcessorID:  "test-processor",
		ProcessorRef: "test-ref-123",
		Success:      true,
		Message:      "Payment processed successfully",
	}, nil
}

func (m *mockPaymentProcessor) RefundPayment(ctx context.Context, payment *domain.Payment, amount int64, reason string) (*RefundResult, error) {
	return &RefundResult{
		RefundRef: "refund-ref-123",
		Success:   true,
		Message:   "Refund processed successfully",
	}, nil
}

func (m *mockPaymentProcessor) GetPaymentStatus(ctx context.Context, processorRef string) (*PaymentStatusResult, error) {
	return &PaymentStatusResult{
		Status:    "completed",
		Reference: processorRef,
		UpdatedAt: time.Now(),
	}, nil
}

type mockCryptoPaymentProcessor struct{}

func (m *mockCryptoPaymentProcessor) CreatePaymentAddress(ctx context.Context, payment *domain.Payment) (*CryptoAddressResult, error) {
	return &CryptoAddressResult{
		Address: "0x1234567890abcdef",
		Network: "ethereum",
	}, nil
}

func (m *mockCryptoPaymentProcessor) VerifyPayment(ctx context.Context, txHash string, network domain.CryptoNetwork) (*CryptoVerificationResult, error) {
	return &CryptoVerificationResult{
		IsValid:       true,
		Amount:        1000,
		BlockNumber:   12345,
		GasUsed:       21000,
		GasPrice:      20000000000,
		Confirmations: 6,
	}, nil
}

func (m *mockCryptoPaymentProcessor) GetTransactionStatus(ctx context.Context, txHash string, network domain.CryptoNetwork) (*CryptoTransactionStatus, error) {
	return &CryptoTransactionStatus{
		Status:        "confirmed",
		Confirmations: 6,
		BlockNumber:   12345,
		IsConfirmed:   true,
	}, nil
}

type mockLoyaltyService struct{}

func (m *mockLoyaltyService) GetTokenBalance(ctx context.Context, customerID string) (int64, error) {
	return 1000, nil
}

func (m *mockLoyaltyService) RedeemTokens(ctx context.Context, customerID string, amount int64) error {
	return nil
}

func (m *mockLoyaltyService) EarnTokens(ctx context.Context, customerID string, amount int64, reason string) error {
	return nil
}

func (m *mockLoyaltyService) GetExchangeRate(ctx context.Context) (float64, error) {
	return 0.01, nil // 1 cent per token
}

func TestPaymentService_CreatePayment(t *testing.T) {
	// Setup
	paymentRepo := newMockPaymentRepository()
	orderRepo := newMockOrderRepository()
	eventPublisher := &mockEventPublisher{}
	paymentProcessor := &mockPaymentProcessor{}
	cryptoProcessor := &mockCryptoPaymentProcessor{}
	loyaltyService := &mockLoyaltyService{}
	logger := logger.New("test")

	service := NewPaymentService(
		paymentRepo,
		orderRepo,
		eventPublisher,
		paymentProcessor,
		cryptoProcessor,
		loyaltyService,
		logger,
	)

	// Create a test order
	items := []*domain.OrderItem{
		{
			ID:         "item1",
			ProductID:  "product1",
			Name:       "Coffee",
			Quantity:   1,
			UnitPrice:  500, // $5.00
			TotalPrice: 500,
		},
	}
	order, err := domain.NewOrder("customer1", items)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}
	order.Status = domain.OrderStatusConfirmed

	// Store the order
	orderRepo.Create(context.Background(), order)

	// Test creating a credit card payment
	req := &CreatePaymentRequest{
		OrderID:       order.ID,
		PaymentMethod: domain.PaymentMethodCreditCard,
		CardLast4:     "1234",
		CardBrand:     "Visa",
	}

	ctx := context.Background()
	resp, err := service.CreatePayment(ctx, req)

	// Assertions
	if err != nil {
		t.Fatalf("CreatePayment failed: %v", err)
	}

	if resp.PaymentID == "" {
		t.Error("PaymentID should not be empty")
	}

	if resp.Status != "PENDING" {
		t.Errorf("Expected status PENDING, got %s", resp.Status)
	}

	if resp.Amount != order.TotalAmount {
		t.Errorf("Expected amount %d, got %d", order.TotalAmount, resp.Amount)
	}

	if resp.Currency != order.Currency {
		t.Errorf("Expected currency %s, got %s", order.Currency, resp.Currency)
	}

	// Verify payment was stored
	payment, err := paymentRepo.GetByID(ctx, resp.PaymentID)
	if err != nil {
		t.Fatalf("Failed to get payment: %v", err)
	}

	if payment.OrderID != order.ID {
		t.Errorf("Expected OrderID %s, got %s", order.ID, payment.OrderID)
	}

	if payment.PaymentMethod != domain.PaymentMethodCreditCard {
		t.Errorf("Expected PaymentMethod %v, got %v", domain.PaymentMethodCreditCard, payment.PaymentMethod)
	}
}

func TestPaymentService_CreateCryptoPayment(t *testing.T) {
	// Setup
	paymentRepo := newMockPaymentRepository()
	orderRepo := newMockOrderRepository()
	eventPublisher := &mockEventPublisher{}
	paymentProcessor := &mockPaymentProcessor{}
	cryptoProcessor := &mockCryptoPaymentProcessor{}
	loyaltyService := &mockLoyaltyService{}
	logger := logger.New("test")

	service := NewPaymentService(
		paymentRepo,
		orderRepo,
		eventPublisher,
		paymentProcessor,
		cryptoProcessor,
		loyaltyService,
		logger,
	)

	// Create a test order
	items := []*domain.OrderItem{
		{
			ID:         "item1",
			ProductID:  "product1",
			Name:       "Coffee",
			Quantity:   1,
			UnitPrice:  1000, // $10.00
			TotalPrice: 1000,
		},
	}
	order, err := domain.NewOrder("customer1", items)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}
	order.Status = domain.OrderStatusConfirmed

	// Store the order
	orderRepo.Create(context.Background(), order)

	// Test creating a crypto payment
	req := &CreatePaymentRequest{
		OrderID:       order.ID,
		PaymentMethod: domain.PaymentMethodCrypto,
		CryptoNetwork: domain.CryptoNetworkEthereum,
		CryptoToken:   "ETH",
	}

	ctx := context.Background()
	resp, err := service.CreatePayment(ctx, req)

	// Assertions
	if err != nil {
		t.Fatalf("CreatePayment failed: %v", err)
	}

	if resp.PaymentID == "" {
		t.Error("PaymentID should not be empty")
	}

	if resp.Status != "PENDING" {
		t.Errorf("Expected status PENDING, got %s", resp.Status)
	}

	if resp.PaymentAddress == "" {
		t.Error("PaymentAddress should not be empty for crypto payments")
	}

	// Verify payment was stored
	payment, err := paymentRepo.GetByID(ctx, resp.PaymentID)
	if err != nil {
		t.Fatalf("Failed to get payment: %v", err)
	}

	if payment.PaymentMethod != domain.PaymentMethodCrypto {
		t.Errorf("Expected PaymentMethod %v, got %v", domain.PaymentMethodCrypto, payment.PaymentMethod)
	}

	if payment.CryptoNetwork != domain.CryptoNetworkEthereum {
		t.Errorf("Expected CryptoNetwork %v, got %v", domain.CryptoNetworkEthereum, payment.CryptoNetwork)
	}
}
