package redisadvanced

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// StreamsClient provides event streaming operations using Redis Streams
type StreamsClient struct {
	redis  *redis.Client
	logger *zap.Logger
	config *StreamsConfig
}

// StreamsConfig contains configuration for streams operations
type StreamsConfig struct {
	MaxLength       int64
	ApproximateMaxLen bool
	BlockTimeout    time.Duration
	ConsumerGroup   string
	ConsumerName    string
}

// StreamEvent represents an event in a stream
type StreamEvent struct {
	ID        string                 `json:"id"`
	Stream    string                 `json:"stream"`
	Fields    map[string]interface{} `json:"fields"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// StreamMessage represents a message for publishing
type StreamMessage struct {
	Stream   string                 `json:"stream"`
	Fields   map[string]interface{} `json:"fields"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ConsumerGroupInfo represents consumer group information
type ConsumerGroupInfo struct {
	Name         string    `json:"name"`
	Consumers    int       `json:"consumers"`
	Pending      int       `json:"pending"`
	LastDelivery time.Time `json:"last_delivery"`
}

// StreamInfo represents stream information
type StreamInfo struct {
	Name           string              `json:"name"`
	Length         int64               `json:"length"`
	Groups         []ConsumerGroupInfo `json:"groups"`
	FirstEntry     *StreamEvent        `json:"first_entry,omitempty"`
	LastEntry      *StreamEvent        `json:"last_entry,omitempty"`
	MaxDeletedID   string              `json:"max_deleted_id,omitempty"`
}

// NewStreamsClient creates a new streams client
func NewStreamsClient(redisClient *redis.Client, logger *zap.Logger, config *StreamsConfig) *StreamsClient {
	if config == nil {
		config = &StreamsConfig{
			MaxLength:         10000,
			ApproximateMaxLen: true,
			BlockTimeout:      5 * time.Second,
			ConsumerGroup:     "default-group",
			ConsumerName:      "default-consumer",
		}
	}

	return &StreamsClient{
		redis:  redisClient,
		logger: logger,
		config: config,
	}
}

// Publish publishes an event to a stream
func (sc *StreamsClient) Publish(ctx context.Context, message *StreamMessage) (string, error) {
	sc.logger.Info("Publishing event to stream", 
		zap.String("stream", message.Stream),
		zap.Any("fields", message.Fields),
	)

	// Prepare fields for Redis
	fields := make(map[string]interface{})
	for k, v := range message.Fields {
		fields[k] = v
	}

	// Add metadata
	if message.Metadata != nil {
		for k, v := range message.Metadata {
			fields[fmt.Sprintf("_meta_%s", k)] = v
		}
	}

	// Add timestamp
	fields["_timestamp"] = time.Now().Unix()

	// Build XADD command
	args := &redis.XAddArgs{
		Stream: message.Stream,
		Values: fields,
	}

	// Add max length if configured
	if sc.config.MaxLength > 0 {
		args.MaxLen = sc.config.MaxLength
		args.Approx = sc.config.ApproximateMaxLen
	}

	eventID, err := sc.redis.XAdd(ctx, args).Result()
	if err != nil {
		return "", fmt.Errorf("failed to publish event: %w", err)
	}

	sc.logger.Info("Event published successfully", 
		zap.String("stream", message.Stream),
		zap.String("event_id", eventID),
	)

	return eventID, nil
}

// Subscribe subscribes to events from streams
func (sc *StreamsClient) Subscribe(ctx context.Context, streams []string, fromID string) (<-chan *StreamEvent, error) {
	sc.logger.Info("Subscribing to streams", zap.Strings("streams", streams))

	eventChan := make(chan *StreamEvent, 100)

	go func() {
		defer close(eventChan)

		// Prepare stream arguments
		streamArgs := make([]string, 0, len(streams)*2)
		for _, stream := range streams {
			streamArgs = append(streamArgs, stream)
			if fromID != "" {
				streamArgs = append(streamArgs, fromID)
			} else {
				streamArgs = append(streamArgs, "$") // Latest
			}
		}

		for {
			select {
			case <-ctx.Done():
				sc.logger.Info("Stream subscription cancelled")
				return
			default:
				// Read from streams
				result, err := sc.redis.XRead(ctx, &redis.XReadArgs{
					Streams: streamArgs,
					Block:   sc.config.BlockTimeout,
				}).Result()

				if err != nil {
					if err != redis.Nil {
						sc.logger.Error("Failed to read from streams", zap.Error(err))
					}
					continue
				}

				// Process messages
				for _, stream := range result {
					for _, message := range stream.Messages {
						event := sc.parseStreamMessage(stream.Stream, message)
						select {
						case eventChan <- event:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}()

	return eventChan, nil
}

// CreateConsumerGroup creates a consumer group for a stream
func (sc *StreamsClient) CreateConsumerGroup(ctx context.Context, stream, group, startID string) error {
	sc.logger.Info("Creating consumer group", 
		zap.String("stream", stream),
		zap.String("group", group),
	)

	if startID == "" {
		startID = "0" // From beginning
	}

	err := sc.redis.XGroupCreate(ctx, stream, group, startID).Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	sc.logger.Info("Consumer group created successfully", 
		zap.String("stream", stream),
		zap.String("group", group),
	)

	return nil
}

// ConsumeGroup consumes events from a stream using consumer group
func (sc *StreamsClient) ConsumeGroup(ctx context.Context, stream, group, consumer string) (<-chan *StreamEvent, error) {
	sc.logger.Info("Starting group consumption", 
		zap.String("stream", stream),
		zap.String("group", group),
		zap.String("consumer", consumer),
	)

	eventChan := make(chan *StreamEvent, 100)

	go func() {
		defer close(eventChan)

		for {
			select {
			case <-ctx.Done():
				sc.logger.Info("Group consumption cancelled")
				return
			default:
				// Read from consumer group
				result, err := sc.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
					Group:    group,
					Consumer: consumer,
					Streams:  []string{stream, ">"},
					Block:    sc.config.BlockTimeout,
					Count:    10,
				}).Result()

				if err != nil {
					if err != redis.Nil {
						sc.logger.Error("Failed to read from consumer group", zap.Error(err))
					}
					continue
				}

				// Process messages
				for _, streamResult := range result {
					for _, message := range streamResult.Messages {
						event := sc.parseStreamMessage(streamResult.Stream, message)
						event.Metadata = map[string]interface{}{
							"consumer_group": group,
							"consumer":       consumer,
						}

						select {
						case eventChan <- event:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}()

	return eventChan, nil
}

// AckMessage acknowledges a message in a consumer group
func (sc *StreamsClient) AckMessage(ctx context.Context, stream, group, messageID string) error {
	err := sc.redis.XAck(ctx, stream, group, messageID).Err()
	if err != nil {
		return fmt.Errorf("failed to acknowledge message: %w", err)
	}

	sc.logger.Debug("Message acknowledged", 
		zap.String("stream", stream),
		zap.String("group", group),
		zap.String("message_id", messageID),
	)

	return nil
}

// GetPendingMessages gets pending messages for a consumer group
func (sc *StreamsClient) GetPendingMessages(ctx context.Context, stream, group string) ([]string, error) {
	result, err := sc.redis.XPending(ctx, stream, group).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get pending messages: %w", err)
	}

	// Get detailed pending info
	pendingExt, err := sc.redis.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: stream,
		Group:  group,
		Start:  "-",
		End:    "+",
		Count:  result.Count,
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to get pending messages details: %w", err)
	}

	messageIDs := make([]string, len(pendingExt))
	for i, pending := range pendingExt {
		messageIDs[i] = pending.ID
	}

	return messageIDs, nil
}

// GetStreamInfo gets information about a stream
func (sc *StreamsClient) GetStreamInfo(ctx context.Context, stream string) (*StreamInfo, error) {
	// Get basic stream info
	info, err := sc.redis.XInfoStream(ctx, stream).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get stream info: %w", err)
	}

	streamInfo := &StreamInfo{
		Name:   stream,
		Length: info.Length,
	}

	// Parse first and last entries
	if len(info.FirstEntry.Values) > 0 {
		streamInfo.FirstEntry = sc.parseStreamMessage(stream, info.FirstEntry)
	}
	if len(info.LastEntry.Values) > 0 {
		streamInfo.LastEntry = sc.parseStreamMessage(stream, info.LastEntry)
	}

	// Get consumer groups info
	groups, err := sc.redis.XInfoGroups(ctx, stream).Result()
	if err == nil {
		streamInfo.Groups = make([]ConsumerGroupInfo, len(groups))
		for i, group := range groups {
			streamInfo.Groups[i] = ConsumerGroupInfo{
				Name:      group.Name,
				Consumers: int(group.Consumers),
				Pending:   int(group.Pending),
			}
		}
	}

	return streamInfo, nil
}

// TrimStream trims a stream to a maximum length
func (sc *StreamsClient) TrimStream(ctx context.Context, stream string, maxLen int64, approximate bool) (int64, error) {
	sc.logger.Info("Trimming stream", 
		zap.String("stream", stream),
		zap.Int64("max_len", maxLen),
	)

	var trimmed int64
	var err error

	if approximate {
		trimmed, err = sc.redis.XTrimMaxLenApprox(ctx, stream, maxLen, 0).Result()
	} else {
		trimmed, err = sc.redis.XTrimMaxLen(ctx, stream, maxLen).Result()
	}

	if err != nil {
		return 0, fmt.Errorf("failed to trim stream: %w", err)
	}

	sc.logger.Info("Stream trimmed successfully", 
		zap.String("stream", stream),
		zap.Int64("trimmed_count", trimmed),
	)

	return trimmed, nil
}

// ReadRange reads events from a stream within a range
func (sc *StreamsClient) ReadRange(ctx context.Context, stream, start, end string, count int64) ([]*StreamEvent, error) {
	sc.logger.Info("Reading stream range", 
		zap.String("stream", stream),
		zap.String("start", start),
		zap.String("end", end),
	)

	result, err := sc.redis.XRange(ctx, stream, start, end).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read stream range: %w", err)
	}

	events := make([]*StreamEvent, len(result))
	for i, message := range result {
		events[i] = sc.parseStreamMessage(stream, message)
	}

	sc.logger.Info("Stream range read successfully", 
		zap.String("stream", stream),
		zap.Int("events_count", len(events)),
	)

	return events, nil
}

// Helper methods

func (sc *StreamsClient) parseStreamMessage(stream string, message redis.XMessage) *StreamEvent {
	event := &StreamEvent{
		ID:     message.ID,
		Stream: stream,
		Fields: make(map[string]interface{}),
	}

	// Parse timestamp from ID
	if parts := strings.Split(message.ID, "-"); len(parts) > 0 {
		if timestampMs, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
			event.Timestamp = time.UnixMilli(timestampMs)
		}
	}

	// Parse fields and metadata
	metadata := make(map[string]interface{})
	for key, value := range message.Values {
		if strings.HasPrefix(key, "_meta_") {
			metaKey := strings.TrimPrefix(key, "_meta_")
			metadata[metaKey] = value
		} else if key == "_timestamp" {
			if timestamp, err := strconv.ParseInt(fmt.Sprintf("%v", value), 10, 64); err == nil {
				event.Timestamp = time.Unix(timestamp, 0)
			}
		} else {
			event.Fields[key] = value
		}
	}

	if len(metadata) > 0 {
		event.Metadata = metadata
	}

	return event
}
