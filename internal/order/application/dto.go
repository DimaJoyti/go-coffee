package application

import (
	"time"

	"github.com/DimaJoyti/go-coffee/internal/order/domain"
)

// Request DTOs

// CreateOrderRequest represents a request to create a new order
type CreateOrderRequest struct {
	CustomerID          string                     `json:"customer_id" validate:"required"`
	Items               []*CreateOrderItemRequest  `json:"items" validate:"required,min=1"`
	SpecialInstructions string                     `json:"special_instructions,omitempty"`
	DeliveryAddress     *CreateAddressRequest      `json:"delivery_address,omitempty"`
}

// CreateOrderItemRequest represents an item in a create order request
type CreateOrderItemRequest struct {
	ProductID      string                        `json:"product_id" validate:"required"`
	Name           string                        `json:"name" validate:"required"`
	Description    string                        `json:"description,omitempty"`
	Quantity       int32                         `json:"quantity" validate:"required,min=1"`
	UnitPrice      int64                         `json:"unit_price" validate:"required,min=0"`
	Customizations []*CreateCustomizationRequest `json:"customizations,omitempty"`
	Metadata       map[string]string             `json:"metadata,omitempty"`
}

// CreateCustomizationRequest represents a customization in a create order request
type CreateCustomizationRequest struct {
	ID         string `json:"id" validate:"required"`
	Name       string `json:"name" validate:"required"`
	Value      string `json:"value" validate:"required"`
	ExtraPrice int64  `json:"extra_price" validate:"min=0"`
}

// CreateAddressRequest represents an address in a create order request
type CreateAddressRequest struct {
	Street     string  `json:"street" validate:"required"`
	City       string  `json:"city" validate:"required"`
	State      string  `json:"state" validate:"required"`
	PostalCode string  `json:"postal_code" validate:"required"`
	Country    string  `json:"country" validate:"required"`
	Latitude   float64 `json:"latitude,omitempty"`
	Longitude  float64 `json:"longitude,omitempty"`
}

// GetOrderRequest represents a request to get an order
type GetOrderRequest struct {
	OrderID    string `json:"order_id" validate:"required"`
	CustomerID string `json:"customer_id,omitempty"` // For authorization
}

// ConfirmOrderRequest represents a request to confirm an order
type ConfirmOrderRequest struct {
	OrderID       string                `json:"order_id" validate:"required"`
	PaymentMethod domain.PaymentMethod  `json:"payment_method,omitempty"`
}

// UpdateOrderStatusRequest represents a request to update order status
type UpdateOrderStatusRequest struct {
	OrderID string `json:"order_id" validate:"required"`
	Status  string `json:"status" validate:"required,oneof=PREPARING READY COMPLETED CANCELLED"`
}

// CancelOrderRequest represents a request to cancel an order
type CancelOrderRequest struct {
	OrderID    string `json:"order_id" validate:"required"`
	CustomerID string `json:"customer_id,omitempty"` // For authorization
	Reason     string `json:"reason,omitempty"`
}

// ListOrdersRequest represents a request to list orders
type ListOrdersRequest struct {
	CustomerID string            `json:"customer_id,omitempty"`
	Status     string            `json:"status,omitempty"`
	Limit      int               `json:"limit,omitempty"`
	Offset     int               `json:"offset,omitempty"`
	Filters    map[string]string `json:"filters,omitempty"`
}

// Response DTOs

// CreateOrderResponse represents a response to create order request
type CreateOrderResponse struct {
	OrderID       string    `json:"order_id"`
	TotalAmount   int64     `json:"total_amount"`
	Currency      string    `json:"currency"`
	EstimatedTime int32     `json:"estimated_time"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// GetOrderResponse represents a response to get order request
type GetOrderResponse struct {
	OrderID             string                  `json:"order_id"`
	CustomerID          string                  `json:"customer_id"`
	Items               []*OrderItemResponse    `json:"items"`
	Status              string                  `json:"status"`
	Priority            domain.OrderPriority    `json:"priority"`
	TotalAmount         int64                   `json:"total_amount"`
	Currency            string                  `json:"currency"`
	PaymentMethod       domain.PaymentMethod    `json:"payment_method"`
	EstimatedTime       int32                   `json:"estimated_time"`
	ActualTime          int32                   `json:"actual_time"`
	SpecialInstructions string                  `json:"special_instructions,omitempty"`
	DeliveryAddress     *AddressResponse        `json:"delivery_address,omitempty"`
	IsDelivery          bool                    `json:"is_delivery"`
	CreatedAt           time.Time               `json:"created_at"`
	UpdatedAt           time.Time               `json:"updated_at"`
	ConfirmedAt         *time.Time              `json:"confirmed_at,omitempty"`
	CompletedAt         *time.Time              `json:"completed_at,omitempty"`
}

// OrderItemResponse represents an order item in responses
type OrderItemResponse struct {
	ID             string                     `json:"id"`
	ProductID      string                     `json:"product_id"`
	Name           string                     `json:"name"`
	Description    string                     `json:"description,omitempty"`
	Quantity       int32                      `json:"quantity"`
	UnitPrice      int64                      `json:"unit_price"`
	TotalPrice     int64                      `json:"total_price"`
	Customizations []*CustomizationResponse   `json:"customizations,omitempty"`
}

// CustomizationResponse represents a customization in responses
type CustomizationResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Value      string `json:"value"`
	ExtraPrice int64  `json:"extra_price"`
}

// AddressResponse represents an address in responses
type AddressResponse struct {
	Street     string  `json:"street"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	PostalCode string  `json:"postal_code"`
	Country    string  `json:"country"`
	Latitude   float64 `json:"latitude,omitempty"`
	Longitude  float64 `json:"longitude,omitempty"`
}

// ConfirmOrderResponse represents a response to confirm order request
type ConfirmOrderResponse struct {
	OrderID       string    `json:"order_id"`
	Status        string    `json:"status"`
	EstimatedTime int32     `json:"estimated_time"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UpdateOrderStatusResponse represents a response to update order status request
type UpdateOrderStatusResponse struct {
	OrderID   string    `json:"order_id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CancelOrderResponse represents a response to cancel order request
type CancelOrderResponse struct {
	OrderID   string    `json:"order_id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListOrdersResponse represents a response to list orders request
type ListOrdersResponse struct {
	Orders     []*GetOrderResponse `json:"orders"`
	Total      int                 `json:"total"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
	HasMore    bool                `json:"has_more"`
}

// Filter DTOs

// OrderFilters represents filters for order queries
type OrderFilters struct {
	CustomerID string
	Status     domain.OrderStatus
	StartDate  *time.Time
	EndDate    *time.Time
	MinAmount  *int64
	MaxAmount  *int64
	IsDelivery *bool
	Limit      int
	Offset     int
}

// PaymentFilters represents filters for payment queries
type PaymentFilters struct {
	CustomerID    string
	OrderID       string
	Status        domain.PaymentStatus
	PaymentMethod domain.PaymentMethod
	StartDate     *time.Time
	EndDate       *time.Time
	MinAmount     *int64
	MaxAmount     *int64
	Limit         int
	Offset        int
}

// External Service DTOs

// UserInfo represents user information from auth service
type UserInfo struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// UserPreferences represents user preferences from auth service
type UserPreferences struct {
	UserID              string            `json:"user_id"`
	DefaultPaymentMethod domain.PaymentMethod `json:"default_payment_method"`
	DeliveryAddress     *AddressResponse  `json:"delivery_address,omitempty"`
	Preferences         map[string]string `json:"preferences"`
}

// Error DTOs

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// ValidationErrorResponse represents a validation error response
type ValidationErrorResponse struct {
	Code   string             `json:"code"`
	Message string            `json:"message"`
	Errors []*ValidationError `json:"errors"`
}

// Payment Request DTOs

// CreatePaymentRequest represents a request to create a payment
type CreatePaymentRequest struct {
	OrderID       string                `json:"order_id" validate:"required"`
	PaymentMethod domain.PaymentMethod  `json:"payment_method" validate:"required"`

	// Card payment fields
	CardLast4     string `json:"card_last4,omitempty"`
	CardBrand     string `json:"card_brand,omitempty"`

	// Crypto payment fields
	CryptoNetwork domain.CryptoNetwork `json:"crypto_network,omitempty"`
	CryptoToken   string               `json:"crypto_token,omitempty"`
}

// ProcessPaymentRequest represents a request to process a payment
type ProcessPaymentRequest struct {
	PaymentID       string `json:"payment_id" validate:"required"`
	TransactionHash string `json:"transaction_hash,omitempty"` // For crypto payments
}

// RefundPaymentRequest represents a request to refund a payment
type RefundPaymentRequest struct {
	PaymentID string `json:"payment_id" validate:"required"`
	Amount    int64  `json:"amount" validate:"required,min=1"`
	Reason    string `json:"reason" validate:"required"`
}

// Payment Response DTOs

// CreatePaymentResponse represents a response to create payment request
type CreatePaymentResponse struct {
	PaymentID       string     `json:"payment_id"`
	Status          string     `json:"status"`
	PaymentAddress  string     `json:"payment_address,omitempty"` // For crypto payments
	TokensUsed      int64      `json:"tokens_used,omitempty"`     // For loyalty token payments
	ExchangeRate    float64    `json:"exchange_rate,omitempty"`   // For loyalty token payments
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`      // For crypto payments
	CreatedAt       time.Time  `json:"created_at"`
}

// ProcessPaymentResponse represents a response to process payment request
type ProcessPaymentResponse struct {
	PaymentID    string    `json:"payment_id"`
	Status       string    `json:"status"`
	ProcessorRef string    `json:"processor_ref,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RefundPaymentResponse represents a response to refund payment request
type RefundPaymentResponse struct {
	RefundID     string    `json:"refund_id"`
	PaymentID    string    `json:"payment_id"`
	Amount       int64     `json:"amount"`
	Status       string    `json:"status"`
	ProcessorRef string    `json:"processor_ref,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// Payment Processor DTOs

// ProcessPaymentResult represents the result of payment processing
type ProcessPaymentResult struct {
	ProcessorID  string `json:"processor_id"`
	ProcessorRef string `json:"processor_ref"`
	Success      bool   `json:"success"`
	Message      string `json:"message,omitempty"`
}

// RefundResult represents the result of payment refund
type RefundResult struct {
	RefundRef string `json:"refund_ref"`
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
}

// PaymentStatusResult represents payment status from processor
type PaymentStatusResult struct {
	Status    string `json:"status"`
	Reference string `json:"reference"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Crypto Payment DTOs

// CryptoAddressResult represents crypto payment address creation result
type CryptoAddressResult struct {
	Address   string     `json:"address"`
	Network   string     `json:"network"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// CryptoVerificationResult represents crypto payment verification result
type CryptoVerificationResult struct {
	IsValid     bool   `json:"is_valid"`
	Amount      int64  `json:"amount"`
	BlockNumber int64  `json:"block_number"`
	GasUsed     int64  `json:"gas_used"`
	GasPrice    int64  `json:"gas_price"`
	Confirmations int  `json:"confirmations"`
}

// CryptoTransactionStatus represents crypto transaction status
type CryptoTransactionStatus struct {
	Status        string `json:"status"`
	Confirmations int    `json:"confirmations"`
	BlockNumber   int64  `json:"block_number"`
	IsConfirmed   bool   `json:"is_confirmed"`
}
