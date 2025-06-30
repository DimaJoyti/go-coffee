package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SerializationFormat represents different serialization formats
type SerializationFormat string

const (
	FormatProtobuf SerializationFormat = "protobuf"
	FormatJSON     SerializationFormat = "json"
	FormatAvro     SerializationFormat = "avro"
)

// EventSerializer handles serialization and deserialization of events
type EventSerializer struct {
	registry *EventRegistry
	format   SerializationFormat
}

// NewEventSerializer creates a new event serializer
func NewEventSerializer(registry *EventRegistry, format SerializationFormat) *EventSerializer {
	return &EventSerializer{
		registry: registry,
		format:   format,
	}
}

// SerializeEvent serializes an event to bytes
func (s *EventSerializer) SerializeEvent(eventType, version string, message proto.Message, metadata map[string]string) ([]byte, error) {
	// Validate the event
	if err := s.registry.ValidateEvent(eventType, version, message); err != nil {
		return nil, fmt.Errorf("event validation failed: %w", err)
	}

	switch s.format {
	case FormatProtobuf:
		return s.serializeProtobuf(eventType, version, message, metadata)
	case FormatJSON:
		return s.serializeJSON(eventType, version, message, metadata)
	default:
		return nil, fmt.Errorf("unsupported serialization format: %s", s.format)
	}
}

// DeserializeEvent deserializes bytes to an event
func (s *EventSerializer) DeserializeEvent(data []byte) (string, string, proto.Message, map[string]string, error) {
	switch s.format {
	case FormatProtobuf:
		return s.deserializeProtobuf(data)
	case FormatJSON:
		return s.deserializeJSON(data)
	default:
		return "", "", nil, nil, fmt.Errorf("unsupported serialization format: %s", s.format)
	}
}

// serializeProtobuf serializes an event using Protocol Buffers
func (s *EventSerializer) serializeProtobuf(eventType, version string, message proto.Message, metadata map[string]string) ([]byte, error) {
	// Create event envelope
	envelope := &EventEnvelope{
		EventId:       uuid.New().String(),
		EventType:     eventType,
		Version:       version,
		Timestamp:     timestamppb.New(time.Now()),
		Source:        "ai-agents",
		CorrelationId: metadata["correlation_id"],
		TraceId:       metadata["trace_id"],
		SpanId:        metadata["span_id"],
		Metadata:      metadata,
	}

	// Pack the message into Any
	payload, err := anypb.New(message)
	if err != nil {
		return nil, fmt.Errorf("failed to pack message: %w", err)
	}
	envelope.Payload = payload

	// Serialize the envelope
	return proto.Marshal(envelope)
}

// deserializeProtobuf deserializes a Protocol Buffer event
func (s *EventSerializer) deserializeProtobuf(data []byte) (string, string, proto.Message, map[string]string, error) {
	// Unmarshal the envelope
	envelope := &EventEnvelope{}
	if err := proto.Unmarshal(data, envelope); err != nil {
		return "", "", nil, nil, fmt.Errorf("failed to unmarshal envelope: %w", err)
	}

	// Get event type info (commented out unused variable)
	// info, err := s.registry.GetEventType(envelope.EventType, envelope.Version)
	// if err != nil {
	// 	return "", "", nil, nil, fmt.Errorf("unknown event type: %w", err)
	// }

	// Create message instance
	message, err := s.registry.CreateEvent(envelope.EventType, envelope.Version)
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("failed to create event instance: %w", err)
	}

	// Unpack the payload
	if err := envelope.Payload.UnmarshalTo(message); err != nil {
		return "", "", nil, nil, fmt.Errorf("failed to unpack payload: %w", err)
	}

	return envelope.EventType, envelope.Version, message, envelope.Metadata, nil
}

// serializeJSON serializes an event using JSON
func (s *EventSerializer) serializeJSON(eventType, version string, message proto.Message, metadata map[string]string) ([]byte, error) {
	// Create a JSON envelope
	envelope := map[string]interface{}{
		"event_id":       uuid.New().String(),
		"event_type":     eventType,
		"version":        version,
		"timestamp":      time.Now().Format(time.RFC3339),
		"source":         "ai-agents",
		"correlation_id": metadata["correlation_id"],
		"trace_id":       metadata["trace_id"],
		"span_id":        metadata["span_id"],
		"metadata":       metadata,
	}

	// Convert protobuf message to JSON
	messageJSON, err := s.protoToJSON(message)
	if err != nil {
		return nil, fmt.Errorf("failed to convert message to JSON: %w", err)
	}
	envelope["payload"] = messageJSON

	return json.Marshal(envelope)
}

// deserializeJSON deserializes a JSON event
func (s *EventSerializer) deserializeJSON(data []byte) (string, string, proto.Message, map[string]string, error) {
	// Parse JSON envelope
	var envelope map[string]interface{}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return "", "", nil, nil, fmt.Errorf("failed to unmarshal JSON envelope: %w", err)
	}

	eventType, ok := envelope["event_type"].(string)
	if !ok {
		return "", "", nil, nil, fmt.Errorf("missing or invalid event_type")
	}

	version, ok := envelope["version"].(string)
	if !ok {
		return "", "", nil, nil, fmt.Errorf("missing or invalid version")
	}

	// Get metadata
	metadata := make(map[string]string)
	if metaData, ok := envelope["metadata"].(map[string]interface{}); ok {
		for k, v := range metaData {
			if str, ok := v.(string); ok {
				metadata[k] = str
			}
		}
	}

	// Create message instance
	message, err := s.registry.CreateEvent(eventType, version)
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("failed to create event instance: %w", err)
	}

	// Convert JSON payload to protobuf
	payload, ok := envelope["payload"]
	if !ok {
		return "", "", nil, nil, fmt.Errorf("missing payload")
	}

	if err := s.jsonToProto(payload, message); err != nil {
		return "", "", nil, nil, fmt.Errorf("failed to convert JSON to protobuf: %w", err)
	}

	return eventType, version, message, metadata, nil
}

// protoToJSON converts a protobuf message to JSON
func (s *EventSerializer) protoToJSON(message proto.Message) (interface{}, error) {
	// Use protobuf's JSON marshaling
	jsonBytes, err := protojson.Marshal(message)
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// jsonToProto converts JSON to a protobuf message
func (s *EventSerializer) jsonToProto(jsonData interface{}, message proto.Message) error {
	// Convert to JSON bytes first
	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}

	// Use protobuf's JSON unmarshaling
	return protojson.Unmarshal(jsonBytes, message)
}

// EventEnvelope represents the common envelope for all events
// This should match the protobuf definition in common.proto
type EventEnvelope struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EventId       string                 `protobuf:"bytes,1,opt,name=event_id,json=eventId,proto3" json:"event_id,omitempty"`
	EventType     string                 `protobuf:"bytes,2,opt,name=event_type,json=eventType,proto3" json:"event_type,omitempty"`
	Version       string                 `protobuf:"bytes,3,opt,name=version,proto3" json:"version,omitempty"`
	Timestamp     *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Source        string                 `protobuf:"bytes,5,opt,name=source,proto3" json:"source,omitempty"`
	CorrelationId string                 `protobuf:"bytes,6,opt,name=correlation_id,json=correlationId,proto3" json:"correlation_id,omitempty"`
	TraceId       string                 `protobuf:"bytes,7,opt,name=trace_id,json=traceId,proto3" json:"trace_id,omitempty"`
	SpanId        string                 `protobuf:"bytes,8,opt,name=span_id,json=spanId,proto3" json:"span_id,omitempty"`
	Metadata      map[string]string      `protobuf:"bytes,9,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Payload       *anypb.Any             `protobuf:"bytes,10,opt,name=payload,proto3" json:"payload,omitempty"`
}

// Implement proto.Message interface
func (e *EventEnvelope) Reset() {
	*e = EventEnvelope{}
}

func (e *EventEnvelope) String() string {
	return protoimpl.X.MessageStringOf(e)
}

func (*EventEnvelope) ProtoMessage() {}

// Implement protoreflect.ProtoMessage interface
func (e *EventEnvelope) ProtoReflect() protoreflect.Message {
	// Return a minimal implementation for now
	// In a real implementation, this should return the generated descriptor
	return protoimpl.X.MessageOf(e)
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (*EventEnvelope) Descriptor() ([]byte, []int) {
	// Return empty descriptor for now - this should be generated by protoc
	return []byte{}, []int{0}
}

// GetEventId returns the EventId field value
func (e *EventEnvelope) GetEventId() string {
	if e != nil {
		return e.EventId
	}
	return ""
}

// GetEventType returns the EventType field value
func (e *EventEnvelope) GetEventType() string {
	if e != nil {
		return e.EventType
	}
	return ""
}

// GetVersion returns the Version field value
func (e *EventEnvelope) GetVersion() string {
	if e != nil {
		return e.Version
	}
	return ""
}

// GetTimestamp returns the Timestamp field value
func (e *EventEnvelope) GetTimestamp() *timestamppb.Timestamp {
	if e != nil {
		return e.Timestamp
	}
	return nil
}

// GetSource returns the Source field value
func (e *EventEnvelope) GetSource() string {
	if e != nil {
		return e.Source
	}
	return ""
}

// GetCorrelationId returns the CorrelationId field value
func (e *EventEnvelope) GetCorrelationId() string {
	if e != nil {
		return e.CorrelationId
	}
	return ""
}

// GetTraceId returns the TraceId field value
func (e *EventEnvelope) GetTraceId() string {
	if e != nil {
		return e.TraceId
	}
	return ""
}

// GetSpanId returns the SpanId field value
func (e *EventEnvelope) GetSpanId() string {
	if e != nil {
		return e.SpanId
	}
	return ""
}

// GetMetadata returns the Metadata field value
func (e *EventEnvelope) GetMetadata() map[string]string {
	if e != nil {
		return e.Metadata
	}
	return nil
}

// GetPayload returns the Payload field value
func (e *EventEnvelope) GetPayload() *anypb.Any {
	if e != nil {
		return e.Payload
	}
	return nil
}

// EventBatch represents a batch of events for efficient processing
type EventBatch struct {
	Events    []*EventEnvelope `json:"events"`
	BatchId   string           `json:"batch_id"`
	Timestamp time.Time        `json:"timestamp"`
	Source    string           `json:"source"`
}

// BatchSerializer handles batch serialization
type BatchSerializer struct {
	eventSerializer *EventSerializer
}

// NewBatchSerializer creates a new batch serializer
func NewBatchSerializer(eventSerializer *EventSerializer) *BatchSerializer {
	return &BatchSerializer{
		eventSerializer: eventSerializer,
	}
}

// SerializeBatch serializes a batch of events
func (bs *BatchSerializer) SerializeBatch(events []proto.Message, eventTypes []string, versions []string, metadata []map[string]string) ([]byte, error) {
	if len(events) != len(eventTypes) || len(events) != len(versions) || len(events) != len(metadata) {
		return nil, fmt.Errorf("mismatched array lengths")
	}

	batch := &EventBatch{
		BatchId:   uuid.New().String(),
		Timestamp: time.Now(),
		Source:    "ai-agents",
		Events:    make([]*EventEnvelope, len(events)),
	}

	for i, event := range events {
		data, err := bs.eventSerializer.SerializeEvent(eventTypes[i], versions[i], event, metadata[i])
		if err != nil {
			return nil, fmt.Errorf("failed to serialize event %d: %w", i, err)
		}

		envelope := &EventEnvelope{}
		if err := proto.Unmarshal(data, envelope); err != nil {
			return nil, fmt.Errorf("failed to unmarshal envelope for event %d: %w", i, err)
		}

		batch.Events[i] = envelope
	}

	return json.Marshal(batch)
}

// DeserializeBatch deserializes a batch of events
func (bs *BatchSerializer) DeserializeBatch(data []byte) ([]proto.Message, []string, []string, []map[string]string, error) {
	var batch EventBatch
	if err := json.Unmarshal(data, &batch); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to unmarshal batch: %w", err)
	}

	events := make([]proto.Message, len(batch.Events))
	eventTypes := make([]string, len(batch.Events))
	versions := make([]string, len(batch.Events))
	metadata := make([]map[string]string, len(batch.Events))

	for i, envelope := range batch.Events {
		envelopeData, err := proto.Marshal(envelope)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("failed to marshal envelope %d: %w", i, err)
		}

		eventType, version, message, meta, err := bs.eventSerializer.DeserializeEvent(envelopeData)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("failed to deserialize event %d: %w", i, err)
		}

		events[i] = message
		eventTypes[i] = eventType
		versions[i] = version
		metadata[i] = meta
	}

	return events, eventTypes, versions, metadata, nil
}
