package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/pkg/security/monitoring"
)

// EventStoreAdapter adapts RedisEventStore to monitoring.EventStore interface
type EventStoreAdapter struct {
	redisEventStore *RedisEventStore
}

// NewEventStoreAdapter creates a new event store adapter
func NewEventStoreAdapter(redisEventStore *RedisEventStore) monitoring.EventStore {
	return &EventStoreAdapter{
		redisEventStore: redisEventStore,
	}
}

// Store implements monitoring.EventStore
func (a *EventStoreAdapter) Store(ctx context.Context, event *monitoring.SecurityEvent) error {
	// Convert monitoring.SecurityEvent to infrastructure.SecurityEvent
	infraEvent := &SecurityEvent{
		ID:          event.ID,
		EventType:   string(event.EventType),
		Severity:    string(event.Severity),
		UserID:      event.UserID,
		IPAddress:   event.IPAddress,
		UserAgent:   event.UserAgent,
		Timestamp:   event.Timestamp,
		ThreatLevel: string(event.ThreatLevel),
		Details:     event.Metadata,
		Source:      event.Source,
		Action:      event.Description,
		Description: event.Description,
		Mitigated:   event.Mitigated,
	}

	return a.redisEventStore.Store(ctx, infraEvent)
}

// Query implements monitoring.EventStore
func (a *EventStoreAdapter) Query(ctx context.Context, filter monitoring.EventFilter) ([]*monitoring.SecurityEvent, error) {
	// Convert monitoring.EventFilter to infrastructure.EventFilter
	infraFilter := EventFilter{
		StartTime:   filter.StartTime,
		EndTime:     filter.EndTime,
		EventTypes:  convertEventTypes(filter.EventTypes),
		Severities:  convertSeverities(filter.Severities),
		UserID:      filter.UserID,
		IPAddress:   filter.IPAddress,
		ThreatLevel: convertThreatLevelToString(filter.ThreatLevel),
		Limit:       filter.Limit,
	}

	result, err := a.redisEventStore.Query(ctx, infraFilter)
	if err != nil {
		return nil, err
	}

	// Convert result back to monitoring.SecurityEvent
	infraEvents, ok := result.([]*SecurityEvent)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	monitoringEvents := make([]*monitoring.SecurityEvent, len(infraEvents))
	for i, infraEvent := range infraEvents {
		monitoringEvents[i] = &monitoring.SecurityEvent{
			ID:          infraEvent.ID,
			Timestamp:   infraEvent.Timestamp,
			EventType:   monitoring.SecurityEventType(infraEvent.EventType),
			Severity:    monitoring.SecuritySeverity(infraEvent.Severity),
			Source:      infraEvent.Source,
			UserID:      infraEvent.UserID,
			IPAddress:   infraEvent.IPAddress,
			UserAgent:   infraEvent.UserAgent,
			Description: infraEvent.Description,
			Metadata:    infraEvent.Details,
			ThreatLevel: monitoring.ThreatLevel(infraEvent.ThreatLevel),
			Mitigated:   infraEvent.Mitigated,
		}
	}

	return monitoringEvents, nil
}

// Delete implements monitoring.EventStore
func (a *EventStoreAdapter) Delete(ctx context.Context, olderThan time.Time) error {
	return a.redisEventStore.Delete(ctx, olderThan)
}

// AlertManagerAdapter adapts RedisAlertManager to monitoring.AlertManager interface
type AlertManagerAdapter struct {
	redisAlertManager *RedisAlertManager
}

// NewAlertManagerAdapter creates a new alert manager adapter
func NewAlertManagerAdapter(redisAlertManager *RedisAlertManager) monitoring.AlertManager {
	return &AlertManagerAdapter{
		redisAlertManager: redisAlertManager,
	}
}

// CreateAlert implements monitoring.AlertManager
func (a *AlertManagerAdapter) CreateAlert(ctx context.Context, alert *monitoring.SecurityAlert) error {
	// Convert monitoring.SecurityAlert to infrastructure.SecurityAlert
	infraAlert := &SecurityAlert{
		ID:          alert.ID,
		Title:       alert.Title,
		Description: alert.Description,
		Severity:    string(alert.Severity),
		Status:      string(alert.Status),
		CreatedAt:   alert.Timestamp,
		UpdatedAt:   alert.Timestamp,
		ResolvedAt:  nil, // Will be set when resolved
		EventIDs:    extractEventIDs(alert.Events),
		Metadata:    alert.Metadata,
	}

	return a.redisAlertManager.CreateAlert(ctx, infraAlert)
}

// UpdateAlert implements monitoring.AlertManager
func (a *AlertManagerAdapter) UpdateAlert(ctx context.Context, alertID string, updates map[string]interface{}) error {
	return a.redisAlertManager.UpdateAlert(ctx, alertID, updates)
}

// GetActiveAlerts implements monitoring.AlertManager
func (a *AlertManagerAdapter) GetActiveAlerts(ctx context.Context) ([]*monitoring.SecurityAlert, error) {
	infraAlerts, err := a.redisAlertManager.GetActiveAlerts(ctx)
	if err != nil {
		return nil, err
	}

	monitoringAlerts := make([]*monitoring.SecurityAlert, len(infraAlerts))
	for i, infraAlert := range infraAlerts {
		monitoringAlerts[i] = &monitoring.SecurityAlert{
			ID:          infraAlert.ID,
			Timestamp:   infraAlert.CreatedAt,
			AlertType:   monitoring.AlertTypeThreatDetected, // Default type
			Severity:    monitoring.SecuritySeverity(infraAlert.Severity),
			Title:       infraAlert.Title,
			Description: infraAlert.Description,
			Events:      []monitoring.SecurityEvent{}, // Would need to fetch events by IDs
			Metadata:    infraAlert.Metadata,
			Status:      monitoring.AlertStatus(infraAlert.Status),
			AssignedTo:  "", // Not supported in infrastructure type
		}
	}

	return monitoringAlerts, nil
}

// ResolveAlert implements monitoring.AlertManager
func (a *AlertManagerAdapter) ResolveAlert(ctx context.Context, alertID string, reason string) error {
	return a.redisAlertManager.ResolveAlert(ctx, alertID, reason)
}

// ThreatDetectorAdapter adapts to monitoring.ThreatDetector interface
type ThreatDetectorAdapter struct {
	// This would wrap an actual threat detector implementation
}

// NewThreatDetectorAdapter creates a new threat detector adapter
func NewThreatDetectorAdapter() monitoring.ThreatDetector {
	return &ThreatDetectorAdapter{}
}

// AnalyzeEvent implements monitoring.ThreatDetector
func (a *ThreatDetectorAdapter) AnalyzeEvent(ctx context.Context, event *monitoring.SecurityEvent) (*monitoring.ThreatAnalysis, error) {
	// Simple threat analysis implementation
	// In a real implementation, this would use ML models, threat intelligence, etc.
	
	threatLevel := monitoring.ThreatLevelNone
	confidence := 0.0
	indicators := []string{}
	recommendations := []string{}
	shouldBlock := false

	// Basic threat detection logic
	switch event.EventType {
	case monitoring.EventTypeMaliciousActivity:
		threatLevel = monitoring.ThreatLevelHigh
		confidence = 0.9
		indicators = append(indicators, "malicious_activity_detected")
		recommendations = append(recommendations, "Block IP address", "Investigate user activity")
		shouldBlock = true
	case monitoring.EventTypeAuthentication:
		if strings.Contains(event.Description, "failed") {
			threatLevel = monitoring.ThreatLevelMedium
			confidence = 0.6
			indicators = append(indicators, "failed_authentication")
			recommendations = append(recommendations, "Monitor for brute force attacks")
		}
	case monitoring.EventTypePrivilegeEscalation:
		threatLevel = monitoring.ThreatLevelHigh
		confidence = 0.8
		indicators = append(indicators, "privilege_escalation_attempt")
		recommendations = append(recommendations, "Immediate investigation required", "Review user permissions")
		shouldBlock = true
	}

	return &monitoring.ThreatAnalysis{
		ThreatLevel:     threatLevel,
		Confidence:      confidence,
		Indicators:      indicators,
		Recommendations: recommendations,
		ShouldBlock:     shouldBlock,
		Metadata: map[string]interface{}{
			"analyzed_at": time.Now(),
			"analyzer":    "basic_threat_detector",
		},
	}, nil
}

// UpdateThreatIntelligence implements monitoring.ThreatDetector
func (a *ThreatDetectorAdapter) UpdateThreatIntelligence(ctx context.Context) error {
	// Placeholder implementation
	// In a real implementation, this would update threat intelligence feeds
	return nil
}

// GetThreatLevel implements monitoring.ThreatDetector
func (a *ThreatDetectorAdapter) GetThreatLevel(ctx context.Context, indicators []string) (monitoring.ThreatLevel, error) {
	// Simple threat level calculation based on indicators
	if len(indicators) == 0 {
		return monitoring.ThreatLevelNone, nil
	}

	highRiskIndicators := []string{"malicious_activity", "privilege_escalation", "data_exfiltration"}
	mediumRiskIndicators := []string{"failed_authentication", "suspicious_ip", "unusual_access_pattern"}

	for _, indicator := range indicators {
		for _, highRisk := range highRiskIndicators {
			if strings.Contains(indicator, highRisk) {
				return monitoring.ThreatLevelHigh, nil
			}
		}
		for _, mediumRisk := range mediumRiskIndicators {
			if strings.Contains(indicator, mediumRisk) {
				return monitoring.ThreatLevelMedium, nil
			}
		}
	}

	return monitoring.ThreatLevelLow, nil
}

// Helper functions for type conversion

func convertEventTypes(monitoringTypes []monitoring.SecurityEventType) []string {
	result := make([]string, len(monitoringTypes))
	for i, t := range monitoringTypes {
		result[i] = string(t)
	}
	return result
}

func convertSeverities(monitoringSeverities []monitoring.SecuritySeverity) []string {
	result := make([]string, len(monitoringSeverities))
	for i, s := range monitoringSeverities {
		result[i] = string(s)
	}
	return result
}

func convertThreatLevelToString(monitoringThreatLevel *monitoring.ThreatLevel) *string {
	if monitoringThreatLevel == nil {
		return nil
	}
	level := string(*monitoringThreatLevel)
	return &level
}

func convertThreatLevelToInt(threatLevel monitoring.ThreatLevel) int {
	switch threatLevel {
	case monitoring.ThreatLevelNone:
		return 0
	case monitoring.ThreatLevelLow:
		return 1
	case monitoring.ThreatLevelMedium:
		return 2
	case monitoring.ThreatLevelHigh:
		return 3
	case monitoring.ThreatLevelCritical:
		return 4
	default:
		return 0
	}
}

func extractEventIDs(events []monitoring.SecurityEvent) []string {
	eventIDs := make([]string, len(events))
	for i, event := range events {
		eventIDs[i] = event.ID
	}
	return eventIDs
}

// convertIntToThreatLevel is no longer needed since we use string representation
