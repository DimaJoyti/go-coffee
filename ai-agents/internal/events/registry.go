package events

import (
	"fmt"
	"reflect"
	"sync"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// EventRegistry manages event types and their schemas
type EventRegistry struct {
	mu           sync.RWMutex
	eventTypes   map[string]EventTypeInfo
	eventsByName map[string]EventTypeInfo
}

// EventTypeInfo contains information about an event type
type EventTypeInfo struct {
	Name        string
	Version     string
	MessageType protoreflect.MessageType
	GoType      reflect.Type
	Description string
	Schema      string
}

// NewEventRegistry creates a new event registry
func NewEventRegistry() *EventRegistry {
	return &EventRegistry{
		eventTypes:   make(map[string]EventTypeInfo),
		eventsByName: make(map[string]EventTypeInfo),
	}
}

// RegisterEvent registers an event type in the registry
func (r *EventRegistry) RegisterEvent(eventType, version, description string, message proto.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s:%s", eventType, version)
	
	// Check if already registered
	if _, exists := r.eventTypes[key]; exists {
		return fmt.Errorf("event type %s version %s already registered", eventType, version)
	}

	msgType := message.ProtoReflect().Type()
	goType := reflect.TypeOf(message)

	info := EventTypeInfo{
		Name:        eventType,
		Version:     version,
		MessageType: msgType,
		GoType:      goType,
		Description: description,
		Schema:      r.generateSchema(msgType),
	}

	r.eventTypes[key] = info
	r.eventsByName[eventType] = info

	return nil
}

// GetEventType retrieves event type information
func (r *EventRegistry) GetEventType(eventType, version string) (EventTypeInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", eventType, version)
	info, exists := r.eventTypes[key]
	if !exists {
		return EventTypeInfo{}, fmt.Errorf("event type %s version %s not found", eventType, version)
	}

	return info, nil
}

// GetLatestEventType retrieves the latest version of an event type
func (r *EventRegistry) GetLatestEventType(eventType string) (EventTypeInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, exists := r.eventsByName[eventType]
	if !exists {
		return EventTypeInfo{}, fmt.Errorf("event type %s not found", eventType)
	}

	return info, nil
}

// ListEventTypes returns all registered event types
func (r *EventRegistry) ListEventTypes() []EventTypeInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var types []EventTypeInfo
	for _, info := range r.eventTypes {
		types = append(types, info)
	}

	return types
}

// ValidateEvent validates an event against its registered schema
func (r *EventRegistry) ValidateEvent(eventType, version string, message proto.Message) error {
	info, err := r.GetEventType(eventType, version)
	if err != nil {
		return err
	}

	// Check if the message type matches
	if message.ProtoReflect().Type() != info.MessageType {
		return fmt.Errorf("message type mismatch for event %s:%s", eventType, version)
	}

	// Validate required fields
	return r.validateRequiredFields(message)
}

// CreateEvent creates a new event instance
func (r *EventRegistry) CreateEvent(eventType, version string) (proto.Message, error) {
	info, err := r.GetEventType(eventType, version)
	if err != nil {
		return nil, err
	}

	// Create new instance using reflection
	msgValue := reflect.New(info.GoType.Elem())
	message, ok := msgValue.Interface().(proto.Message)
	if !ok {
		return nil, fmt.Errorf("failed to create event instance for %s:%s", eventType, version)
	}

	return message, nil
}

// generateSchema generates a JSON schema representation of the protobuf message
func (r *EventRegistry) generateSchema(msgType protoreflect.MessageType) string {
	// This is a simplified schema generation
	// In a real implementation, you might want to use a proper JSON schema generator
	descriptor := msgType.Descriptor()
	return fmt.Sprintf("Message: %s, Fields: %d", descriptor.FullName(), descriptor.Fields().Len())
}

// validateRequiredFields validates that all required fields are present
func (r *EventRegistry) validateRequiredFields(message proto.Message) error {
	reflection := message.ProtoReflect()
	descriptor := reflection.Descriptor()

	fields := descriptor.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		
		// Check if field is required (in proto3, all fields are optional by default)
		// This is a simplified validation - you might want more sophisticated rules
		if field.Cardinality() == protoreflect.Required {
			if !reflection.Has(field) {
				return fmt.Errorf("required field %s is missing", field.Name())
			}
		}
	}

	return nil
}

// EventVersionManager manages event versioning and compatibility
type EventVersionManager struct {
	registry *EventRegistry
}

// NewEventVersionManager creates a new event version manager
func NewEventVersionManager(registry *EventRegistry) *EventVersionManager {
	return &EventVersionManager{
		registry: registry,
	}
}

// IsCompatible checks if two event versions are compatible
func (vm *EventVersionManager) IsCompatible(eventType, fromVersion, toVersion string) (bool, error) {
	fromInfo, err := vm.registry.GetEventType(eventType, fromVersion)
	if err != nil {
		return false, err
	}

	toInfo, err := vm.registry.GetEventType(eventType, toVersion)
	if err != nil {
		return false, err
	}

	// Simple version compatibility check
	// In a real implementation, you'd have more sophisticated compatibility rules
	return vm.checkSchemaCompatibility(fromInfo, toInfo), nil
}

// checkSchemaCompatibility performs basic schema compatibility checking
func (vm *EventVersionManager) checkSchemaCompatibility(from, to EventTypeInfo) bool {
	// This is a simplified compatibility check
	// In practice, you'd want to check field additions, removals, type changes, etc.
	return from.MessageType.Descriptor().FullName() == to.MessageType.Descriptor().FullName()
}

// MigrateEvent migrates an event from one version to another
func (vm *EventVersionManager) MigrateEvent(eventType, fromVersion, toVersion string, message proto.Message) (proto.Message, error) {
	compatible, err := vm.IsCompatible(eventType, fromVersion, toVersion)
	if err != nil {
		return nil, err
	}

	if !compatible {
		return nil, fmt.Errorf("incompatible versions: %s -> %s for event type %s", fromVersion, toVersion, eventType)
	}

	// Create new event instance
	newEvent, err := vm.registry.CreateEvent(eventType, toVersion)
	if err != nil {
		return nil, err
	}

	// Copy compatible fields
	err = vm.copyCompatibleFields(message, newEvent)
	if err != nil {
		return nil, err
	}

	return newEvent, nil
}

// copyCompatibleFields copies fields between compatible event versions
func (vm *EventVersionManager) copyCompatibleFields(from, to proto.Message) error {
	fromReflection := from.ProtoReflect()
	toReflection := to.ProtoReflect()

	fromDescriptor := fromReflection.Descriptor()
	toDescriptor := toReflection.Descriptor()

	// Copy fields that exist in both versions
	fromFields := fromDescriptor.Fields()
	for i := 0; i < fromFields.Len(); i++ {
		fromField := fromFields.Get(i)
		toField := toDescriptor.Fields().ByName(fromField.Name())

		if toField != nil && fromReflection.Has(fromField) {
			// Check if field types are compatible
			if fromField.Kind() == toField.Kind() {
				value := fromReflection.Get(fromField)
				toReflection.Set(toField, value)
			}
		}
	}

	return nil
}

// DefaultEventRegistry is the global event registry instance
var DefaultEventRegistry = NewEventRegistry()

// RegisterEvent registers an event type in the default registry
func RegisterEvent(eventType, version, description string, message proto.Message) error {
	return DefaultEventRegistry.RegisterEvent(eventType, version, description, message)
}

// GetEventType retrieves event type information from the default registry
func GetEventType(eventType, version string) (EventTypeInfo, error) {
	return DefaultEventRegistry.GetEventType(eventType, version)
}

// ValidateEvent validates an event against the default registry
func ValidateEvent(eventType, version string, message proto.Message) error {
	return DefaultEventRegistry.ValidateEvent(eventType, version, message)
}
