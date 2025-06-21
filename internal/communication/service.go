package communication

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/DimaJoyti/go-coffee/internal/communication/domain"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
	redismcp "github.com/DimaJoyti/go-coffee/pkg/redis-mcp"
)

// Service implements the Communication Service
type Service struct {
	hub    *Hub
	router *Router
	logger *logger.Logger
	UnimplementedCommunicationServiceServer
}

// NewService creates a new communication service
func NewService(hub *Hub, router *Router, logger *logger.Logger) *Service {
	return &Service{
		hub:    hub,
		router: router,
		logger: logger,
	}
}

// SendMessage sends a message through the communication hub
func (s *Service) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	s.logger.WithFields(map[string]interface{}{
		"type":   req.Type,
		"source": req.Source,
		"target": req.Target,
	}).Info("Sending message")

	// Create domain message
	payload := make(map[string]interface{})
	for k, v := range req.Payload {
		payload[k] = v
	}

	message, err := domain.NewMessage(
		domain.MessageType(req.Type),
		req.Source,
		req.Target,
		payload,
	)
	if err != nil {
		s.logger.WithField("error", err).Error("Failed to create message")
		return nil, status.Errorf(codes.InvalidArgument, "invalid message: %v", err)
	}

	// Set message properties
	if req.Priority != "" {
		switch req.Priority {
		case "LOW":
			message.SetPriority(domain.MessagePriorityLow)
		case "NORMAL":
			message.SetPriority(domain.MessagePriorityNormal)
		case "HIGH":
			message.SetPriority(domain.MessagePriorityHigh)
		case "CRITICAL":
			message.SetPriority(domain.MessagePriorityCritical)
		}
	}

	if req.CorrelationId != "" {
		message.SetCorrelationID(req.CorrelationId)
	}

	if req.ExpirationSeconds > 0 {
		message.SetExpiration(time.Duration(req.ExpirationSeconds) * time.Second)
	}

	// Add headers
	for k, v := range req.Headers {
		message.AddHeader(k, v)
	}

	// Send message through hub
	if err := s.hub.SendMessage(ctx, message); err != nil {
		s.logger.WithField("error", err).Error("Failed to send message")
		return nil, status.Errorf(codes.Internal, "failed to send message: %v", err)
	}

	return &SendMessageResponse{
		MessageId: message.ID,
		Status:    "SENT",
		Timestamp: time.Now().Unix(),
	}, nil
}

// PublishEvent publishes an event through the communication hub
func (s *Service) PublishEvent(ctx context.Context, req *PublishEventRequest) (*PublishEventResponse, error) {
	s.logger.WithFields(map[string]interface{}{
		"type":         req.Type,
		"source":       req.Source,
		"aggregate_id": req.AggregateId,
	}).Info("Publishing event")

	// Create domain event
	data := make(map[string]interface{})
	for k, v := range req.Data {
		data[k] = v
	}

	event, err := domain.NewEvent(
		domain.EventType(req.Type),
		req.Source,
		req.AggregateId,
		req.AggregateType,
		data,
	)
	if err != nil {
		s.logger.WithField("error", err).Error("Failed to create event")
		return nil, status.Errorf(codes.InvalidArgument, "invalid event: %v", err)
	}

	// Set event properties
	if req.Version > 0 {
		event.SetVersion(req.Version)
	}

	if req.CorrelationId != "" {
		event.SetCorrelationID(req.CorrelationId)
	}

	if req.ExpirationSeconds > 0 {
		event.SetExpiration(time.Duration(req.ExpirationSeconds) * time.Second)
	}

	// Add metadata
	for k, v := range req.Metadata {
		event.AddMetadata(k, v)
	}

	// Publish event through hub
	if err := s.hub.PublishEvent(ctx, event); err != nil {
		s.logger.WithField("error", err).Error("Failed to publish event")
		return nil, status.Errorf(codes.Internal, "failed to publish event: %v", err)
	}

	return &PublishEventResponse{
		EventId:   event.ID,
		Status:    "PUBLISHED",
		Timestamp: time.Now().Unix(),
	}, nil
}

// Subscribe creates a subscription for messages or events
func (s *Service) Subscribe(ctx context.Context, req *SubscribeRequest) (*SubscribeResponse, error) {
	s.logger.WithFields(map[string]interface{}{
		"name":          req.Name,
		"subscriber_id": req.SubscriberId,
		"topic":         req.Topic,
	}).Info("Creating subscription")

	// Determine subscription type
	var subType domain.SubscriptionType
	switch req.Type {
	case "MESSAGE":
		subType = domain.SubscriptionTypeMessage
	case "EVENT":
		subType = domain.SubscriptionTypeEvent
	case "WEBSOCKET":
		subType = domain.SubscriptionTypeWebSocket
	case "WEBHOOK":
		subType = domain.SubscriptionTypeWebhook
	default:
		subType = domain.SubscriptionTypeMessage
	}

	// Create subscription
	subscription, err := domain.NewSubscription(
		req.Name,
		req.SubscriberId,
		req.SubscriberType,
		req.Topic,
		subType,
	)
	if err != nil {
		s.logger.WithField("error", err).Error("Failed to create subscription")
		return nil, status.Errorf(codes.InvalidArgument, "invalid subscription: %v", err)
	}

	// Add filters
	for _, filter := range req.Filters {
		err := subscription.AddFilter(
			filter.Field,
			domain.FilterOperator(filter.Operator),
			filter.Value,
		)
		if err != nil {
			s.logger.WithField("error", err).Error("Failed to add filter")
			return nil, status.Errorf(codes.InvalidArgument, "invalid filter: %v", err)
		}
	}

	// Set webhook endpoint if provided
	if req.WebhookUrl != "" {
		headers := make(map[string]string)
		for k, v := range req.WebhookHeaders {
			headers[k] = v
		}
		if err := subscription.SetWebhookEndpoint(req.WebhookUrl, headers); err != nil {
			s.logger.WithField("error", err).Error("Failed to set webhook endpoint")
			return nil, status.Errorf(codes.InvalidArgument, "invalid webhook: %v", err)
		}
	}

	// Register subscription with hub
	if err := s.hub.RegisterSubscription(ctx, subscription); err != nil {
		s.logger.WithField("error", err).Error("Failed to register subscription")
		return nil, status.Errorf(codes.Internal, "failed to register subscription: %v", err)
	}

	return &SubscribeResponse{
		SubscriptionId: subscription.ID,
		Status:         "ACTIVE",
		Timestamp:      time.Now().Unix(),
	}, nil
}

// GetHealth returns the health status of the communication service
func (s *Service) GetHealth(ctx context.Context, req *HealthRequest) (*HealthResponse, error) {
	// Check hub health
	hubHealth := s.hub.Health(ctx)
	routerHealth := s.router.Health(ctx)

	status := "HEALTHY"
	if hubHealth != nil || routerHealth != nil {
		status = "UNHEALTHY"
	}

	return &HealthResponse{
		Status:    status,
		Timestamp: time.Now().Unix(),
		Details: map[string]string{
			"hub_status":    healthStatusString(hubHealth),
			"router_status": healthStatusString(routerHealth),
		},
	}, nil
}

// GetStats returns communication statistics
func (s *Service) GetStats(ctx context.Context, req *StatsRequest) (*StatsResponse, error) {
	hubStats := s.hub.GetStats(ctx)
	routerStats := s.router.GetStats(ctx)

	return &StatsResponse{
		TotalMessages:    hubStats.TotalMessages,
		TotalEvents:      hubStats.TotalEvents,
		ActiveSubscriptions: hubStats.ActiveSubscriptions,
		MessageRate:      routerStats.MessageRate,
		ErrorRate:        routerStats.ErrorRate,
		Timestamp:        time.Now().Unix(),
	}, nil
}

// Helper functions

func healthStatusString(err error) string {
	if err == nil {
		return "healthy"
	}
	return fmt.Sprintf("unhealthy: %v", err)
}

// Hub represents the communication hub
type Hub struct {
	redisClient *redis.Client
	aiService   *redismcp.AIService
	logger      *logger.Logger
	stats       *HubStats
}

// HubStats represents hub statistics
type HubStats struct {
	TotalMessages       int64
	TotalEvents         int64
	ActiveSubscriptions int64
}

// NewHub creates a new communication hub
func NewHub(redisClient *redis.Client, aiService *redismcp.AIService, logger *logger.Logger) *Hub {
	return &Hub{
		redisClient: redisClient,
		aiService:   aiService,
		logger:      logger,
		stats:       &HubStats{},
	}
}

// Start starts the communication hub
func (h *Hub) Start(ctx context.Context) {
	h.logger.Info("Starting communication hub")
	// Implementation for starting background processes
}

// Stop stops the communication hub
func (h *Hub) Stop() {
	h.logger.Info("Stopping communication hub")
	// Implementation for stopping background processes
}

// SendMessage sends a message through the hub
func (h *Hub) SendMessage(ctx context.Context, message *domain.Message) error {
	h.logger.WithField("message_id", message.ID).Debug("Sending message through hub")
	
	// Store message in Redis
	messageData, err := message.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	key := fmt.Sprintf("comm:messages:%s", message.ID)
	if err := h.redisClient.Set(ctx, key, messageData, time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store message: %w", err)
	}

	// Publish to Redis pub/sub
	channel := fmt.Sprintf("comm:channel:%s", message.Target)
	if err := h.redisClient.Publish(ctx, channel, messageData).Err(); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	h.stats.TotalMessages++
	return nil
}

// PublishEvent publishes an event through the hub
func (h *Hub) PublishEvent(ctx context.Context, event *domain.Event) error {
	h.logger.WithField("event_id", event.ID).Debug("Publishing event through hub")
	
	// Store event in Redis
	eventData, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	key := fmt.Sprintf("comm:events:%s", event.ID)
	if err := h.redisClient.Set(ctx, key, eventData, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store event: %w", err)
	}

	// Publish to Redis pub/sub
	channel := fmt.Sprintf("comm:events:%s", event.Type)
	if err := h.redisClient.Publish(ctx, channel, eventData).Err(); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	h.stats.TotalEvents++
	return nil
}

// RegisterSubscription registers a subscription
func (h *Hub) RegisterSubscription(ctx context.Context, subscription *domain.Subscription) error {
	h.logger.WithField("subscription_id", subscription.ID).Debug("Registering subscription")
	
	// Store subscription in Redis
	subscriptionData, err := subscription.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize subscription: %w", err)
	}

	key := fmt.Sprintf("comm:subscriptions:%s", subscription.ID)
	if err := h.redisClient.Set(ctx, key, subscriptionData, 0).Err(); err != nil {
		return fmt.Errorf("failed to store subscription: %w", err)
	}

	h.stats.ActiveSubscriptions++
	return nil
}

// Health checks the health of the hub
func (h *Hub) Health(ctx context.Context) error {
	_, err := h.redisClient.Ping(ctx).Result()
	return err
}

// GetStats returns hub statistics
func (h *Hub) GetStats(ctx context.Context) *HubStats {
	return h.stats
}

// Router represents the message router
type Router struct {
	redisClient *redis.Client
	aiService   *redismcp.AIService
	logger      *logger.Logger
	stats       *RouterStats
}

// RouterStats represents router statistics
type RouterStats struct {
	MessageRate float64
	ErrorRate   float64
}

// NewRouter creates a new message router
func NewRouter(redisClient *redis.Client, aiService *redismcp.AIService, logger *logger.Logger) *Router {
	return &Router{
		redisClient: redisClient,
		aiService:   aiService,
		logger:      logger,
		stats:       &RouterStats{},
	}
}

// Start starts the message router
func (r *Router) Start(ctx context.Context) {
	r.logger.Info("Starting message router")
	// Implementation for starting routing processes
}

// Stop stops the message router
func (r *Router) Stop() {
	r.logger.Info("Stopping message router")
	// Implementation for stopping routing processes
}

// Health checks the health of the router
func (r *Router) Health(ctx context.Context) error {
	return r.aiService.Health(ctx)
}

// GetStats returns router statistics
func (r *Router) GetStats(ctx context.Context) *RouterStats {
	return r.stats
}
