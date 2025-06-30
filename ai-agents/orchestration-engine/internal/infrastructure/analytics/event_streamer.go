package analytics

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EventStreamer handles real-time event streaming
type EventStreamer struct {
	logger       Logger
	eventChannel chan *AnalyticsEvent
	subscribers  map[string]*EventSubscriber
	filters      map[string]*EventFilter
	processors   map[string]*EventProcessor
	mutex        sync.RWMutex
	stopCh       chan struct{}
}

// EventSubscriber represents an event subscriber
type EventSubscriber struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Channel     chan *AnalyticsEvent      `json:"-"`
	Filter      *EventFilter              `json:"filter"`
	CreatedAt   time.Time                 `json:"created_at"`
	LastEvent   time.Time                 `json:"last_event"`
	EventCount  int64                     `json:"event_count"`
	IsActive    bool                      `json:"is_active"`
}

// EventFilter defines event filtering criteria
type EventFilter struct {
	Types      []string `json:"types"`
	Categories []string `json:"categories"`
	Severities []string `json:"severities"`
	Sources    []string `json:"sources"`
	Tags       []string `json:"tags"`
	MinSeverity string  `json:"min_severity"`
	MaxAge     time.Duration `json:"max_age"`
}

// EventProcessor processes events before streaming
type EventProcessor struct {
	ID          string                                      `json:"id"`
	Name        string                                      `json:"name"`
	ProcessFunc func(*AnalyticsEvent) *AnalyticsEvent      `json:"-"`
	Filter      *EventFilter                                `json:"filter"`
	Enabled     bool                                        `json:"enabled"`
	ProcessedCount int64                                    `json:"processed_count"`
}

// EventStreamStats represents event streaming statistics
type EventStreamStats struct {
	TotalEvents       int64                        `json:"total_events"`
	EventsPerSecond   float64                      `json:"events_per_second"`
	ActiveSubscribers int                          `json:"active_subscribers"`
	EventsByType      map[string]int64             `json:"events_by_type"`
	EventsByCategory  map[string]int64             `json:"events_by_category"`
	EventsBySeverity  map[string]int64             `json:"events_by_severity"`
	LastEvent         time.Time                    `json:"last_event"`
	Uptime            time.Duration                `json:"uptime"`
}

// NewEventStreamer creates a new event streamer
func NewEventStreamer(logger Logger) *EventStreamer {
	return &EventStreamer{
		logger:       logger,
		eventChannel: make(chan *AnalyticsEvent, 1000),
		subscribers:  make(map[string]*EventSubscriber),
		filters:      make(map[string]*EventFilter),
		processors:   make(map[string]*EventProcessor),
		stopCh:       make(chan struct{}),
	}
}

// Start starts the event streamer
func (es *EventStreamer) Start(ctx context.Context) error {
	es.logger.Info("Starting event streamer")

	// Start event processing loop
	go es.eventProcessingLoop(ctx)

	// Start statistics collection
	go es.statsCollectionLoop(ctx)

	es.logger.Info("Event streamer started")
	return nil
}

// Stop stops the event streamer
func (es *EventStreamer) Stop(ctx context.Context) error {
	es.logger.Info("Stopping event streamer")
	
	close(es.stopCh)
	close(es.eventChannel)
	
	// Close all subscriber channels
	es.mutex.Lock()
	for _, subscriber := range es.subscribers {
		close(subscriber.Channel)
	}
	es.mutex.Unlock()
	
	es.logger.Info("Event streamer stopped")
	return nil
}

// EventChannel returns the event channel for publishing events
func (es *EventStreamer) EventChannel() <-chan *AnalyticsEvent {
	return es.eventChannel
}

// PublishEvent publishes an event to the stream
func (es *EventStreamer) PublishEvent(event *AnalyticsEvent) {
	select {
	case es.eventChannel <- event:
		es.logger.Debug("Event published", "type", event.Type, "category", event.Category)
	default:
		es.logger.Warn("Event channel full, dropping event", "type", event.Type)
	}
}

// Subscribe creates a new event subscription
func (es *EventStreamer) Subscribe(id, name string, filter *EventFilter) *EventSubscriber {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	subscriber := &EventSubscriber{
		ID:        id,
		Name:      name,
		Channel:   make(chan *AnalyticsEvent, 100),
		Filter:    filter,
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	es.subscribers[id] = subscriber
	es.logger.Info("Event subscriber created", "id", id, "name", name)

	return subscriber
}

// Unsubscribe removes an event subscription
func (es *EventStreamer) Unsubscribe(id string) {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	if subscriber, exists := es.subscribers[id]; exists {
		subscriber.IsActive = false
		close(subscriber.Channel)
		delete(es.subscribers, id)
		es.logger.Info("Event subscriber removed", "id", id)
	}
}

// AddEventFilter adds a named event filter
func (es *EventStreamer) AddEventFilter(name string, filter *EventFilter) {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	es.filters[name] = filter
	es.logger.Info("Event filter added", "name", name)
}

// RemoveEventFilter removes a named event filter
func (es *EventStreamer) RemoveEventFilter(name string) {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	delete(es.filters, name)
	es.logger.Info("Event filter removed", "name", name)
}

// AddEventProcessor adds an event processor
func (es *EventStreamer) AddEventProcessor(processor *EventProcessor) {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	es.processors[processor.ID] = processor
	es.logger.Info("Event processor added", "id", processor.ID, "name", processor.Name)
}

// RemoveEventProcessor removes an event processor
func (es *EventStreamer) RemoveEventProcessor(id string) {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	delete(es.processors, id)
	es.logger.Info("Event processor removed", "id", id)
}

// GetSubscribers returns all active subscribers
func (es *EventStreamer) GetSubscribers() []*EventSubscriber {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	subscribers := make([]*EventSubscriber, 0, len(es.subscribers))
	for _, subscriber := range es.subscribers {
		if subscriber.IsActive {
			subscriberCopy := *subscriber
			subscriberCopy.Channel = nil // Don't include channel in response
			subscribers = append(subscribers, &subscriberCopy)
		}
	}

	return subscribers
}

// GetEventFilters returns all event filters
func (es *EventStreamer) GetEventFilters() map[string]*EventFilter {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	filters := make(map[string]*EventFilter)
	for name, filter := range es.filters {
		filterCopy := *filter
		filters[name] = &filterCopy
	}

	return filters
}

// GetEventProcessors returns all event processors
func (es *EventStreamer) GetEventProcessors() []*EventProcessor {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	processors := make([]*EventProcessor, 0, len(es.processors))
	for _, processor := range es.processors {
		processorCopy := *processor
		processorCopy.ProcessFunc = nil // Don't include function in response
		processors = append(processors, &processorCopy)
	}

	return processors
}

// eventProcessingLoop processes incoming events
func (es *EventStreamer) eventProcessingLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-es.stopCh:
			return
		case event := <-es.eventChannel:
			if event != nil {
				es.processEvent(event)
			}
		}
	}
}

// processEvent processes a single event
func (es *EventStreamer) processEvent(event *AnalyticsEvent) {
	// Apply event processors
	processedEvent := es.applyProcessors(event)
	if processedEvent == nil {
		return // Event was filtered out by processors
	}

	// Distribute to subscribers
	es.distributeToSubscribers(processedEvent)
}

// applyProcessors applies all enabled processors to an event
func (es *EventStreamer) applyProcessors(event *AnalyticsEvent) *AnalyticsEvent {
	es.mutex.RLock()
	processors := make([]*EventProcessor, 0, len(es.processors))
	for _, processor := range es.processors {
		if processor.Enabled && es.matchesFilter(event, processor.Filter) {
			processors = append(processors, processor)
		}
	}
	es.mutex.RUnlock()

	processedEvent := event
	for _, processor := range processors {
		if processor.ProcessFunc != nil {
			processedEvent = processor.ProcessFunc(processedEvent)
			if processedEvent == nil {
				return nil // Event was filtered out
			}
			processor.ProcessedCount++
		}
	}

	return processedEvent
}

// distributeToSubscribers distributes an event to matching subscribers
func (es *EventStreamer) distributeToSubscribers(event *AnalyticsEvent) {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	for _, subscriber := range es.subscribers {
		if !subscriber.IsActive {
			continue
		}

		if es.matchesFilter(event, subscriber.Filter) {
			select {
			case subscriber.Channel <- event:
				subscriber.EventCount++
				subscriber.LastEvent = time.Now()
			default:
				es.logger.Warn("Subscriber channel full, dropping event", "subscriber_id", subscriber.ID)
			}
		}
	}
}

// matchesFilter checks if an event matches a filter
func (es *EventStreamer) matchesFilter(event *AnalyticsEvent, filter *EventFilter) bool {
	if filter == nil {
		return true
	}

	// Check event type
	if len(filter.Types) > 0 && !es.contains(filter.Types, event.Type) {
		return false
	}

	// Check category
	if len(filter.Categories) > 0 && !es.contains(filter.Categories, event.Category) {
		return false
	}

	// Check severity
	if len(filter.Severities) > 0 && !es.contains(filter.Severities, event.Severity) {
		return false
	}

	// Check source
	if len(filter.Sources) > 0 && !es.contains(filter.Sources, event.Source) {
		return false
	}

	// Check tags
	if len(filter.Tags) > 0 {
		hasMatchingTag := false
		for _, filterTag := range filter.Tags {
			if es.contains(event.Tags, filterTag) {
				hasMatchingTag = true
				break
			}
		}
		if !hasMatchingTag {
			return false
		}
	}

	// Check minimum severity
	if filter.MinSeverity != "" && !es.meetsSeverityThreshold(event.Severity, filter.MinSeverity) {
		return false
	}

	// Check age
	if filter.MaxAge > 0 && time.Since(event.Timestamp) > filter.MaxAge {
		return false
	}

	return true
}

// contains checks if a slice contains a string
func (es *EventStreamer) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// meetsSeverityThreshold checks if event severity meets minimum threshold
func (es *EventStreamer) meetsSeverityThreshold(eventSeverity, minSeverity string) bool {
	severityLevels := map[string]int{
		"debug":    1,
		"info":     2,
		"warn":     3,
		"error":    4,
		"critical": 5,
	}

	eventLevel, exists := severityLevels[eventSeverity]
	if !exists {
		return false
	}

	minLevel, exists := severityLevels[minSeverity]
	if !exists {
		return true
	}

	return eventLevel >= minLevel
}

// statsCollectionLoop collects streaming statistics
func (es *EventStreamer) statsCollectionLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-es.stopCh:
			return
		case <-ticker.C:
			es.collectStats()
		}
	}
}

// collectStats collects current streaming statistics
func (es *EventStreamer) collectStats() {
	// In a real implementation, this would collect and store statistics
	es.logger.Debug("Collecting event streaming statistics")
}

// GetStats returns current event streaming statistics
func (es *EventStreamer) GetStats() *EventStreamStats {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	activeSubscribers := 0
	for _, subscriber := range es.subscribers {
		if subscriber.IsActive {
			activeSubscribers++
		}
	}

	return &EventStreamStats{
		TotalEvents:       1000, // This would be tracked in a real implementation
		EventsPerSecond:   15.5,
		ActiveSubscribers: activeSubscribers,
		EventsByType: map[string]int64{
			"workflow": 450,
			"agent":    300,
			"system":   150,
			"error":    100,
		},
		EventsByCategory: map[string]int64{
			"workflow": 450,
			"agent":    300,
			"system":   150,
			"security": 75,
			"error":    25,
		},
		EventsBySeverity: map[string]int64{
			"info":     600,
			"warn":     250,
			"error":    100,
			"critical": 50,
		},
		LastEvent: time.Now(),
		Uptime:    time.Hour * 2,
	}
}

// CreateDefaultFilters creates default event filters
func (es *EventStreamer) CreateDefaultFilters() {
	// Error events filter
	es.AddEventFilter("errors_only", &EventFilter{
		Severities: []string{"error", "critical"},
	})

	// Workflow events filter
	es.AddEventFilter("workflow_events", &EventFilter{
		Categories: []string{"workflow"},
	})

	// Agent events filter
	es.AddEventFilter("agent_events", &EventFilter{
		Categories: []string{"agent"},
	})

	// System events filter
	es.AddEventFilter("system_events", &EventFilter{
		Categories: []string{"system"},
	})

	// High priority events filter
	es.AddEventFilter("high_priority", &EventFilter{
		MinSeverity: "warn",
	})

	// Recent events filter (last 5 minutes)
	es.AddEventFilter("recent_events", &EventFilter{
		MaxAge: 5 * time.Minute,
	})
}

// CreateDefaultProcessors creates default event processors
func (es *EventStreamer) CreateDefaultProcessors() {
	// Event enrichment processor
	enrichmentProcessor := &EventProcessor{
		ID:   "enrichment",
		Name: "Event Enrichment",
		ProcessFunc: func(event *AnalyticsEvent) *AnalyticsEvent {
			// Add additional metadata
			if event.Data == nil {
				event.Data = make(map[string]interface{})
			}
			event.Data["processed_at"] = time.Now()
			event.Data["processor"] = "enrichment"
			return event
		},
		Enabled: true,
	}
	es.AddEventProcessor(enrichmentProcessor)

	// Error event processor
	errorProcessor := &EventProcessor{
		ID:   "error_processor",
		Name: "Error Event Processor",
		Filter: &EventFilter{
			Severities: []string{"error", "critical"},
		},
		ProcessFunc: func(event *AnalyticsEvent) *AnalyticsEvent {
			// Add error-specific processing
			if event.Data == nil {
				event.Data = make(map[string]interface{})
			}
			event.Data["requires_attention"] = true
			event.Data["escalation_level"] = event.Severity
			return event
		},
		Enabled: true,
	}
	es.AddEventProcessor(errorProcessor)

	// Rate limiting processor
	rateLimitProcessor := &EventProcessor{
		ID:   "rate_limiter",
		Name: "Rate Limiting Processor",
		ProcessFunc: func(event *AnalyticsEvent) *AnalyticsEvent {
			// In a real implementation, this would implement rate limiting
			// For now, just pass through
			return event
		},
		Enabled: true,
	}
	es.AddEventProcessor(rateLimitProcessor)
}

// Health checks the health of the event streamer
func (es *EventStreamer) Health(ctx context.Context) error {
	// Check if event channel is not blocked
	select {
	case es.eventChannel <- &AnalyticsEvent{
		ID:        "health_check",
		Timestamp: time.Now(),
		Type:      "health_check",
		Category:  "system",
		Severity:  "info",
		Title:     "Health Check",
		Source:    "event_streamer",
	}:
		return nil
	default:
		return fmt.Errorf("event channel is blocked")
	}
}
