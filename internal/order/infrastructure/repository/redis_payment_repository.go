package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/DimaJoyti/go-coffee/internal/order/application"
	"github.com/DimaJoyti/go-coffee/internal/order/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

// RedisPaymentRepository implements PaymentRepository using Redis
type RedisPaymentRepository struct {
	client *redis.Client
	logger *logger.Logger
}

// NewRedisPaymentRepository creates a new Redis-based payment repository
func NewRedisPaymentRepository(client *redis.Client, logger *logger.Logger) *RedisPaymentRepository {
	return &RedisPaymentRepository{
		client: client,
		logger: logger,
	}
}

// Create creates a new payment in Redis
func (r *RedisPaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	// Serialize payment to JSON
	paymentData, err := json.Marshal(payment)
	if err != nil {
		return fmt.Errorf("failed to marshal payment: %w", err)
	}

	// Redis keys
	paymentKey := r.getPaymentKey(payment.ID)
	orderPaymentKey := r.getOrderPaymentKey(payment.OrderID)
	customerPaymentsKey := r.getCustomerPaymentsKey(payment.CustomerID)
	paymentsByStatusKey := r.getPaymentsByStatusKey(payment.Status)
	paymentsByMethodKey := r.getPaymentsByMethodKey(payment.PaymentMethod)

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Store payment data
	pipe.Set(ctx, paymentKey, paymentData, 0)

	// Map order to payment (one-to-one)
	pipe.Set(ctx, orderPaymentKey, payment.ID, 0)

	// Add to customer payments (sorted set by creation time)
	pipe.ZAdd(ctx, customerPaymentsKey, &redis.Z{
		Score:  float64(payment.CreatedAt.Unix()),
		Member: payment.ID,
	})

	// Add to payments by status (sorted set by creation time)
	pipe.ZAdd(ctx, paymentsByStatusKey, &redis.Z{
		Score:  float64(payment.CreatedAt.Unix()),
		Member: payment.ID,
	})

	// Add to payments by method (sorted set by creation time)
	pipe.ZAdd(ctx, paymentsByMethodKey, &redis.Z{
		Score:  float64(payment.CreatedAt.Unix()),
		Member: payment.ID,
	})

	// Set expiration for customer payments (90 days)
	pipe.Expire(ctx, customerPaymentsKey, 90*24*time.Hour)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create payment in Redis: %w", err)
	}

	r.logger.WithFields(map[string]interface{}{
		"payment_id":     payment.ID,
		"order_id":       payment.OrderID,
		"customer_id":    payment.CustomerID,
		"amount":         payment.Amount,
		"payment_method": payment.PaymentMethod,
		"status":         payment.Status.String(),
	}).Info("Payment created in Redis")

	return nil
}

// GetByID retrieves a payment by ID
func (r *RedisPaymentRepository) GetByID(ctx context.Context, paymentID string) (*domain.Payment, error) {
	paymentKey := r.getPaymentKey(paymentID)

	// Get payment data
	paymentData, err := r.client.Get(ctx, paymentKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("payment not found: %s", paymentID)
		}
		return nil, fmt.Errorf("failed to get payment from Redis: %w", err)
	}

	// Deserialize payment
	var payment domain.Payment
	if err := json.Unmarshal([]byte(paymentData), &payment); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment: %w", err)
	}

	return &payment, nil
}

// GetByOrderID retrieves a payment by order ID
func (r *RedisPaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	orderPaymentKey := r.getOrderPaymentKey(orderID)

	// Get payment ID
	paymentID, err := r.client.Get(ctx, orderPaymentKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("payment not found for order: %s", orderID)
		}
		return nil, fmt.Errorf("failed to get payment ID from Redis: %w", err)
	}

	// Get payment by ID
	return r.GetByID(ctx, paymentID)
}

// Update updates an existing payment
func (r *RedisPaymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	// Get existing payment to check status change
	existingPayment, err := r.GetByID(ctx, payment.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing payment: %w", err)
	}

	// Serialize payment to JSON
	paymentData, err := json.Marshal(payment)
	if err != nil {
		return fmt.Errorf("failed to marshal payment: %w", err)
	}

	// Redis keys
	paymentKey := r.getPaymentKey(payment.ID)

	// Use pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Update payment data
	pipe.Set(ctx, paymentKey, paymentData, 0)

	// If status changed, update status indexes
	if existingPayment.Status != payment.Status {
		// Remove from old status set
		oldStatusKey := r.getPaymentsByStatusKey(existingPayment.Status)
		pipe.ZRem(ctx, oldStatusKey, payment.ID)

		// Add to new status set
		newStatusKey := r.getPaymentsByStatusKey(payment.Status)
		pipe.ZAdd(ctx, newStatusKey, &redis.Z{
			Score:  float64(payment.UpdatedAt.Unix()),
			Member: payment.ID,
		})
	}

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update payment in Redis: %w", err)
	}

	r.logger.WithFields(map[string]interface{}{
		"payment_id": payment.ID,
		"order_id":   payment.OrderID,
		"status":     payment.Status.String(),
	}).Info("Payment updated in Redis")

	return nil
}

// List retrieves payments based on filters
func (r *RedisPaymentRepository) List(ctx context.Context, filters application.PaymentFilters) ([]*domain.Payment, error) {
	var paymentIDs []string
	var err error

	// Determine which index to use based on filters
	if filters.CustomerID != "" {
		// Use customer-specific index
		customerPaymentsKey := r.getCustomerPaymentsKey(filters.CustomerID)
		paymentIDs, err = r.client.ZRevRange(ctx, customerPaymentsKey, int64(filters.Offset), int64(filters.Offset+filters.Limit-1)).Result()
	} else if filters.Status != domain.PaymentStatusUnknown {
		// Use status-specific index
		paymentsByStatusKey := r.getPaymentsByStatusKey(filters.Status)
		paymentIDs, err = r.client.ZRevRange(ctx, paymentsByStatusKey, int64(filters.Offset), int64(filters.Offset+filters.Limit-1)).Result()
	} else if filters.PaymentMethod != domain.PaymentMethodUnknown {
		// Use payment method-specific index
		paymentsByMethodKey := r.getPaymentsByMethodKey(filters.PaymentMethod)
		paymentIDs, err = r.client.ZRevRange(ctx, paymentsByMethodKey, int64(filters.Offset), int64(filters.Offset+filters.Limit-1)).Result()
	} else {
		// Use scan to get all payments (less efficient, but works for admin queries)
		return r.scanAllPayments(ctx, filters)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get payment IDs from Redis: %w", err)
	}

	// Get payments and apply additional filters
	payments := make([]*domain.Payment, 0, len(paymentIDs))
	for _, paymentID := range paymentIDs {
		payment, err := r.GetByID(ctx, paymentID)
		if err != nil {
			r.logger.WithError(err).WithField("payment_id", paymentID).Warn("Failed to get payment")
			continue
		}

		// Apply additional filters
		if r.matchesFilters(payment, filters) {
			payments = append(payments, payment)
		}
	}

	return payments, nil
}

// Helper methods

// scanAllPayments scans all payments (used when no specific index can be used)
func (r *RedisPaymentRepository) scanAllPayments(ctx context.Context, filters application.PaymentFilters) ([]*domain.Payment, error) {
	var payments []*domain.Payment
	var cursor uint64
	count := 0

	for {
		// Scan for payment keys
		keys, nextCursor, err := r.client.Scan(ctx, cursor, "order:payments:*", 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment keys: %w", err)
		}

		// Get payments
		for _, key := range keys {
			if count >= filters.Offset+filters.Limit {
				break
			}

			paymentData, err := r.client.Get(ctx, key).Result()
			if err != nil {
				continue
			}

			var payment domain.Payment
			if err := json.Unmarshal([]byte(paymentData), &payment); err != nil {
				continue
			}

			// Apply filters
			if r.matchesFilters(&payment, filters) {
				if count >= filters.Offset {
					payments = append(payments, &payment)
				}
				count++
			}
		}

		cursor = nextCursor
		if cursor == 0 || count >= filters.Offset+filters.Limit {
			break
		}
	}

	return payments, nil
}

// matchesFilters checks if a payment matches the given filters
func (r *RedisPaymentRepository) matchesFilters(payment *domain.Payment, filters application.PaymentFilters) bool {
	// Order ID filter
	if filters.OrderID != "" && payment.OrderID != filters.OrderID {
		return false
	}

	// Date range filter
	if filters.StartDate != nil && payment.CreatedAt.Before(*filters.StartDate) {
		return false
	}
	if filters.EndDate != nil && payment.CreatedAt.After(*filters.EndDate) {
		return false
	}

	// Amount range filter
	if filters.MinAmount != nil && payment.Amount < *filters.MinAmount {
		return false
	}
	if filters.MaxAmount != nil && payment.Amount > *filters.MaxAmount {
		return false
	}

	return true
}

// Redis key generators

func (r *RedisPaymentRepository) getPaymentKey(paymentID string) string {
	return fmt.Sprintf("order:payments:%s", paymentID)
}

func (r *RedisPaymentRepository) getOrderPaymentKey(orderID string) string {
	return fmt.Sprintf("order:order_payment:%s", orderID)
}

func (r *RedisPaymentRepository) getCustomerPaymentsKey(customerID string) string {
	return fmt.Sprintf("order:customer_payments:%s", customerID)
}

func (r *RedisPaymentRepository) getPaymentsByStatusKey(status domain.PaymentStatus) string {
	return fmt.Sprintf("order:payments_by_status:%s", status.String())
}

func (r *RedisPaymentRepository) getPaymentsByMethodKey(method domain.PaymentMethod) string {
	return fmt.Sprintf("order:payments_by_method:%d", int(method))
}

// GetPaymentStats retrieves payment statistics (bonus method)
func (r *RedisPaymentRepository) GetPaymentStats(ctx context.Context, customerID string) (*PaymentStats, error) {
	customerPaymentsKey := r.getCustomerPaymentsKey(customerID)

	// Get total payment count
	totalPayments, err := r.client.ZCard(ctx, customerPaymentsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get payment count: %w", err)
	}

	// Get recent payments to calculate stats
	recentPaymentIDs, err := r.client.ZRevRange(ctx, customerPaymentsKey, 0, 99).Result() // Last 100 payments
	if err != nil {
		return nil, fmt.Errorf("failed to get recent payments: %w", err)
	}

	var totalAmount int64
	statusCounts := make(map[string]int)
	methodCounts := make(map[string]int)

	for _, paymentID := range recentPaymentIDs {
		payment, err := r.GetByID(ctx, paymentID)
		if err != nil {
			continue
		}

		totalAmount += payment.Amount
		statusCounts[payment.Status.String()]++
		methodCounts[fmt.Sprintf("%d", int(payment.PaymentMethod))]++
	}

	var averageAmount int64
	if len(recentPaymentIDs) > 0 {
		averageAmount = totalAmount / int64(len(recentPaymentIDs))
	}

	return &PaymentStats{
		TotalPayments: totalPayments,
		TotalAmount:   totalAmount,
		AverageAmount: averageAmount,
		StatusCounts:  statusCounts,
		MethodCounts:  methodCounts,
	}, nil
}

// PaymentStats represents payment statistics
type PaymentStats struct {
	TotalPayments int64          `json:"total_payments"`
	TotalAmount   int64          `json:"total_amount"`
	AverageAmount int64          `json:"average_amount"`
	StatusCounts  map[string]int `json:"status_counts"`
	MethodCounts  map[string]int `json:"method_counts"`
}
