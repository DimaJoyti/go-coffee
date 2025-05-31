package domain

import (
	"encoding/json"
	"errors"
	"regexp"
	"time"
)

// SubscriptionType represents different types of subscriptions
type SubscriptionType int32

const (
	SubscriptionTypeEvent       SubscriptionType = 0 // Event-based subscription
	SubscriptionTypeMessage     SubscriptionType = 1 // Message-based subscription
	SubscriptionTypeWebSocket   SubscriptionType = 2 // WebSocket subscription
	SubscriptionTypeWebhook     SubscriptionType = 3 // Webhook subscription
)

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus int32

const (
	SubscriptionStatusActive    SubscriptionStatus = 0
	SubscriptionStatusPaused    SubscriptionStatus = 1
	SubscriptionStatusInactive  SubscriptionStatus = 2
	SubscriptionStatusError     SubscriptionStatus = 3
)

// FilterOperator represents filter operators
type FilterOperator string

const (
	FilterOperatorEquals       FilterOperator = "eq"
	FilterOperatorNotEquals    FilterOperator = "ne"
	FilterOperatorContains     FilterOperator = "contains"
	FilterOperatorStartsWith   FilterOperator = "starts_with"
	FilterOperatorEndsWith     FilterOperator = "ends_with"
	FilterOperatorGreaterThan  FilterOperator = "gt"
	FilterOperatorLessThan     FilterOperator = "lt"
	FilterOperatorIn           FilterOperator = "in"
	FilterOperatorNotIn        FilterOperator = "not_in"
	FilterOperatorRegex        FilterOperator = "regex"
)

// Subscription represents a subscription to messages or events
type Subscription struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	Type            SubscriptionType   `json:"type"`
	Status          SubscriptionStatus `json:"status"`
	SubscriberID    string             `json:"subscriber_id"`    // Service or user ID
	SubscriberType  string             `json:"subscriber_type"`  // "service", "user", "webhook"
	Topic           string             `json:"topic"`            // Topic or pattern to subscribe to
	Filters         []*Filter          `json:"filters"`          // Message/event filters
	DeadLetterTopic string             `json:"dead_letter_topic,omitempty"` // Dead letter queue topic
	RetryPolicy     *RetryPolicy       `json:"retry_policy,omitempty"`
	RateLimit       *RateLimit         `json:"rate_limit,omitempty"`
	Endpoint        string             `json:"endpoint,omitempty"`        // For webhook subscriptions
	Headers         map[string]string  `json:"headers,omitempty"`         // For webhook subscriptions
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	LastDeliveryAt  *time.Time         `json:"last_delivery_at,omitempty"`
	DeliveryCount   int64              `json:"delivery_count"`
	ErrorCount      int64              `json:"error_count"`
	Metadata        map[string]string  `json:"metadata,omitempty"`
}

// Filter represents a message/event filter
type Filter struct {
	Field    string         `json:"field"`    // Field to filter on (e.g., "type", "source", "data.order_id")
	Operator FilterOperator `json:"operator"` // Filter operator
	Value    interface{}    `json:"value"`    // Filter value
}

// RetryPolicy represents retry configuration
type RetryPolicy struct {
	MaxRetries      int32         `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	ExponentialBase float64       `json:"exponential_base"`
}

// RateLimit represents rate limiting configuration
type RateLimit struct {
	RequestsPerSecond int32         `json:"requests_per_second"`
	BurstSize         int32         `json:"burst_size"`
	WindowSize        time.Duration `json:"window_size"`
}

// NewSubscription creates a new subscription
func NewSubscription(name, subscriberID, subscriberType, topic string, subType SubscriptionType) (*Subscription, error) {
	if name == "" {
		return nil, errors.New("subscription name is required")
	}
	
	if subscriberID == "" {
		return nil, errors.New("subscriber ID is required")
	}
	
	if subscriberType == "" {
		return nil, errors.New("subscriber type is required")
	}
	
	if topic == "" {
		return nil, errors.New("topic is required")
	}

	return &Subscription{
		ID:             generateSubscriptionID(),
		Name:           name,
		Type:           subType,
		Status:         SubscriptionStatusActive,
		SubscriberID:   subscriberID,
		SubscriberType: subscriberType,
		Topic:          topic,
		Filters:        make([]*Filter, 0),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		DeliveryCount:  0,
		ErrorCount:     0,
		Metadata:       make(map[string]string),
	}, nil
}

// AddFilter adds a filter to the subscription
func (s *Subscription) AddFilter(field string, operator FilterOperator, value interface{}) error {
	if field == "" {
		return errors.New("filter field is required")
	}

	filter := &Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	}

	s.Filters = append(s.Filters, filter)
	s.UpdatedAt = time.Now()

	return nil
}

// RemoveFilter removes a filter by index
func (s *Subscription) RemoveFilter(index int) error {
	if index < 0 || index >= len(s.Filters) {
		return errors.New("invalid filter index")
	}

	s.Filters = append(s.Filters[:index], s.Filters[index+1:]...)
	s.UpdatedAt = time.Now()

	return nil
}

// SetRetryPolicy sets the retry policy
func (s *Subscription) SetRetryPolicy(maxRetries int32, initialDelay, maxDelay time.Duration, backoffFactor, exponentialBase float64) {
	s.RetryPolicy = &RetryPolicy{
		MaxRetries:      maxRetries,
		InitialDelay:    initialDelay,
		MaxDelay:        maxDelay,
		BackoffFactor:   backoffFactor,
		ExponentialBase: exponentialBase,
	}
	s.UpdatedAt = time.Now()
}

// SetRateLimit sets the rate limit
func (s *Subscription) SetRateLimit(requestsPerSecond, burstSize int32, windowSize time.Duration) {
	s.RateLimit = &RateLimit{
		RequestsPerSecond: requestsPerSecond,
		BurstSize:         burstSize,
		WindowSize:        windowSize,
	}
	s.UpdatedAt = time.Now()
}

// SetWebhookEndpoint sets the webhook endpoint and headers
func (s *Subscription) SetWebhookEndpoint(endpoint string, headers map[string]string) error {
	if s.Type != SubscriptionTypeWebhook {
		return errors.New("webhook endpoint can only be set for webhook subscriptions")
	}

	s.Endpoint = endpoint
	s.Headers = headers
	s.UpdatedAt = time.Now()

	return nil
}

// SetStatus sets the subscription status
func (s *Subscription) SetStatus(status SubscriptionStatus) {
	s.Status = status
	s.UpdatedAt = time.Now()
}

// Pause pauses the subscription
func (s *Subscription) Pause() {
	s.SetStatus(SubscriptionStatusPaused)
}

// Resume resumes the subscription
func (s *Subscription) Resume() {
	s.SetStatus(SubscriptionStatusActive)
}

// Deactivate deactivates the subscription
func (s *Subscription) Deactivate() {
	s.SetStatus(SubscriptionStatusInactive)
}

// IncrementDeliveryCount increments the delivery count
func (s *Subscription) IncrementDeliveryCount() {
	s.DeliveryCount++
	now := time.Now()
	s.LastDeliveryAt = &now
	s.UpdatedAt = now
}

// IncrementErrorCount increments the error count
func (s *Subscription) IncrementErrorCount() {
	s.ErrorCount++
	s.UpdatedAt = time.Now()
}

// IsActive checks if the subscription is active
func (s *Subscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive
}

// MatchesMessage checks if a message matches the subscription filters
func (s *Subscription) MatchesMessage(message *Message) bool {
	// Check topic match first
	if !s.matchesTopic(string(message.Type)) {
		return false
	}

	// Apply filters
	for _, filter := range s.Filters {
		if !s.applyFilter(filter, message) {
			return false
		}
	}

	return true
}

// MatchesEvent checks if an event matches the subscription filters
func (s *Subscription) MatchesEvent(event *Event) bool {
	// Check topic match first
	if !s.matchesTopic(string(event.Type)) {
		return false
	}

	// Apply filters
	for _, filter := range s.Filters {
		if !s.applyEventFilter(filter, event) {
			return false
		}
	}

	return true
}

// Helper methods

// matchesTopic checks if a topic matches the subscription topic pattern
func (s *Subscription) matchesTopic(topic string) bool {
	// Support wildcard patterns
	// * matches any single level
	// ** matches any number of levels
	// Example: "order.*" matches "order.created", "order.updated"
	// Example: "order.**" matches "order.created", "order.payment.completed"

	if s.Topic == "*" || s.Topic == "**" {
		return true
	}

	if s.Topic == topic {
		return true
	}

	// Convert wildcard pattern to regex
	pattern := s.Topic
	pattern = regexp.QuoteMeta(pattern)
	pattern = regexp.MustCompile(`\\\*\\\*`).ReplaceAllString(pattern, `.*`)
	pattern = regexp.MustCompile(`\\\*`).ReplaceAllString(pattern, `[^.]*`)
	pattern = "^" + pattern + "$"

	matched, _ := regexp.MatchString(pattern, topic)
	return matched
}

// applyFilter applies a filter to a message
func (s *Subscription) applyFilter(filter *Filter, message *Message) bool {
	value := s.getMessageFieldValue(filter.Field, message)
	return s.compareValues(filter.Operator, value, filter.Value)
}

// applyEventFilter applies a filter to an event
func (s *Subscription) applyEventFilter(filter *Filter, event *Event) bool {
	value := s.getEventFieldValue(filter.Field, event)
	return s.compareValues(filter.Operator, value, filter.Value)
}

// getMessageFieldValue gets a field value from a message
func (s *Subscription) getMessageFieldValue(field string, message *Message) interface{} {
	switch field {
	case "type":
		return string(message.Type)
	case "source":
		return message.Source
	case "target":
		return message.Target
	case "priority":
		return int32(message.Priority)
	default:
		// Check payload fields (e.g., "payload.order_id")
		if len(field) > 8 && field[:8] == "payload." {
			payloadField := field[8:]
			return message.Payload[payloadField]
		}
		// Check headers (e.g., "headers.content_type")
		if len(field) > 8 && field[:8] == "headers." {
			headerField := field[8:]
			return message.Headers[headerField]
		}
		// Check metadata (e.g., "metadata.service")
		if len(field) > 9 && field[:9] == "metadata." {
			metadataField := field[9:]
			return message.Metadata[metadataField]
		}
	}
	return nil
}

// getEventFieldValue gets a field value from an event
func (s *Subscription) getEventFieldValue(field string, event *Event) interface{} {
	switch field {
	case "type":
		return string(event.Type)
	case "source":
		return event.Source
	case "aggregate_id":
		return event.AggregateID
	case "aggregate_type":
		return event.AggregateType
	case "version":
		return event.Version
	default:
		// Check data fields (e.g., "data.order_id")
		if len(field) > 5 && field[:5] == "data." {
			dataField := field[5:]
			return event.Data[dataField]
		}
		// Check metadata (e.g., "metadata.service")
		if len(field) > 9 && field[:9] == "metadata." {
			metadataField := field[9:]
			return event.Metadata[metadataField]
		}
	}
	return nil
}

// compareValues compares two values using the specified operator
func (s *Subscription) compareValues(operator FilterOperator, actual, expected interface{}) bool {
	if actual == nil {
		return operator == FilterOperatorNotEquals
	}

	switch operator {
	case FilterOperatorEquals:
		return actual == expected
	case FilterOperatorNotEquals:
		return actual != expected
	case FilterOperatorContains:
		if actualStr, ok := actual.(string); ok {
			if expectedStr, ok := expected.(string); ok {
				return regexp.MustCompile(regexp.QuoteMeta(expectedStr)).MatchString(actualStr)
			}
		}
	case FilterOperatorStartsWith:
		if actualStr, ok := actual.(string); ok {
			if expectedStr, ok := expected.(string); ok {
				return regexp.MustCompile("^"+regexp.QuoteMeta(expectedStr)).MatchString(actualStr)
			}
		}
	case FilterOperatorEndsWith:
		if actualStr, ok := actual.(string); ok {
			if expectedStr, ok := expected.(string); ok {
				return regexp.MustCompile(regexp.QuoteMeta(expectedStr)+"$").MatchString(actualStr)
			}
		}
	case FilterOperatorRegex:
		if actualStr, ok := actual.(string); ok {
			if expectedStr, ok := expected.(string); ok {
				matched, _ := regexp.MatchString(expectedStr, actualStr)
				return matched
			}
		}
	}

	return false
}

// ToJSON converts the subscription to JSON
func (s *Subscription) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// FromJSON creates a subscription from JSON
func SubscriptionFromJSON(data []byte) (*Subscription, error) {
	var subscription Subscription
	err := json.Unmarshal(data, &subscription)
	return &subscription, err
}

// generateSubscriptionID generates a unique subscription ID
func generateSubscriptionID() string {
	return "sub_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}
