package zones

import (
	"fmt"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service implements the zone service
type Service struct {
	logger     *zap.Logger
	repository domain.ZoneRepository
	config     ServiceConfig

	// In-memory tracking for performance
	zoneCache    map[string]*domain.DetectionZone
	occupancy    map[string]map[string]*ObjectPresence // zoneID -> objectID -> presence
	statistics   map[string]*domain.ZoneStatistics
	mutex        sync.RWMutex
}

// Ensure Service implements domain.ZoneService
var _ domain.ZoneService = (*Service)(nil)

// ServiceConfig configures the zone service
type ServiceConfig struct {
	EnableCaching          bool          `yaml:"enable_caching"`
	CacheRefreshInterval   time.Duration `yaml:"cache_refresh_interval"`
	StatisticsUpdateInterval time.Duration `yaml:"statistics_update_interval"`
	MaxEventHistory        int           `yaml:"max_event_history"`
	EnableRealTimeProcessing bool        `yaml:"enable_realtime_processing"`
}

// ObjectPresence tracks an object's presence in a zone
type ObjectPresence struct {
	ObjectID     string
	ZoneID       string
	EntryTime    time.Time
	LastSeen     time.Time
	Position     domain.Point
	BoundingBox  domain.Rectangle
	ObjectClass  string
	Confidence   float64
	IsStationary bool
	DwellTime    time.Duration
}

// DefaultServiceConfig returns default configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		EnableCaching:            true,
		CacheRefreshInterval:     30 * time.Second,
		StatisticsUpdateInterval: 10 * time.Second,
		MaxEventHistory:          1000,
		EnableRealTimeProcessing: true,
	}
}

// NewService creates a new zone service
func NewService(logger *zap.Logger, repository domain.ZoneRepository, config ServiceConfig) *Service {
	return &Service{
		logger:     logger.With(zap.String("component", "zone_service")),
		repository: repository,
		config:     config,
		zoneCache:  make(map[string]*domain.DetectionZone),
		occupancy:  make(map[string]map[string]*ObjectPresence),
		statistics: make(map[string]*domain.ZoneStatistics),
	}
}

// CreateZone creates a new detection zone
func (s *Service) CreateZone(zone *domain.DetectionZone) error {
	// Validate zone
	if err := zone.Validate(); err != nil {
		return fmt.Errorf("invalid zone: %w", err)
	}

	// Set timestamps
	now := time.Now()
	zone.CreatedAt = now
	zone.UpdatedAt = now

	// Generate ID if not provided
	if zone.ID == "" {
		zone.ID = uuid.New().String()
	}

	// Create in repository
	if err := s.repository.CreateZone(zone); err != nil {
		return fmt.Errorf("failed to create zone: %w", err)
	}

	// Update cache
	s.mutex.Lock()
	s.zoneCache[zone.ID] = zone
	s.occupancy[zone.ID] = make(map[string]*ObjectPresence)
	s.statistics[zone.ID] = &domain.ZoneStatistics{
		ZoneID:       zone.ID,
		StreamID:     zone.StreamID,
		ObjectCounts: make(map[string]int64),
		HourlyStats:  make(map[int]int64),
		DailyStats:   make(map[string]int64),
		StartTime:    now,
	}
	s.mutex.Unlock()

	s.logger.Info("Zone created",
		zap.String("zone_id", zone.ID),
		zap.String("stream_id", zone.StreamID),
		zap.String("name", zone.Name),
		zap.String("type", string(zone.Type)))

	return nil
}

// GetZone retrieves a zone by ID
func (s *Service) GetZone(id string) (*domain.DetectionZone, error) {
	// Check cache first
	if s.config.EnableCaching {
		s.mutex.RLock()
		if zone, exists := s.zoneCache[id]; exists {
			s.mutex.RUnlock()
			return zone, nil
		}
		s.mutex.RUnlock()
	}

	// Get from repository
	zone, err := s.repository.GetZone(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone: %w", err)
	}

	// Update cache
	if s.config.EnableCaching {
		s.mutex.Lock()
		s.zoneCache[id] = zone
		s.mutex.Unlock()
	}

	return zone, nil
}

// GetZonesByStream retrieves all zones for a stream
func (s *Service) GetZonesByStream(streamID string) ([]*domain.DetectionZone, error) {
	zones, err := s.repository.GetZonesByStream(streamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get zones for stream: %w", err)
	}

	// Update cache
	if s.config.EnableCaching {
		s.mutex.Lock()
		for _, zone := range zones {
			s.zoneCache[zone.ID] = zone
			if _, exists := s.occupancy[zone.ID]; !exists {
				s.occupancy[zone.ID] = make(map[string]*ObjectPresence)
			}
		}
		s.mutex.Unlock()
	}

	return zones, nil
}

// UpdateZone updates an existing zone
func (s *Service) UpdateZone(zone *domain.DetectionZone) error {
	// Validate zone
	if err := zone.Validate(); err != nil {
		return fmt.Errorf("invalid zone: %w", err)
	}

	// Set update timestamp
	zone.UpdatedAt = time.Now()

	// Update in repository
	if err := s.repository.UpdateZone(zone); err != nil {
		return fmt.Errorf("failed to update zone: %w", err)
	}

	// Update cache
	s.mutex.Lock()
	s.zoneCache[zone.ID] = zone
	s.mutex.Unlock()

	s.logger.Info("Zone updated",
		zap.String("zone_id", zone.ID),
		zap.String("name", zone.Name))

	return nil
}

// DeleteZone deletes a zone
func (s *Service) DeleteZone(id string) error {
	// Delete from repository
	if err := s.repository.DeleteZone(id); err != nil {
		return fmt.Errorf("failed to delete zone: %w", err)
	}

	// Remove from cache
	s.mutex.Lock()
	delete(s.zoneCache, id)
	delete(s.occupancy, id)
	delete(s.statistics, id)
	s.mutex.Unlock()

	s.logger.Info("Zone deleted", zap.String("zone_id", id))
	return nil
}

// ListZones lists all zones with pagination
func (s *Service) ListZones(limit, offset int) ([]*domain.DetectionZone, error) {
	return s.repository.ListZones(limit, offset)
}

// ProcessDetection processes a detection against all zones for a stream
func (s *Service) ProcessDetection(streamID string, detection *domain.DetectedObject) ([]*domain.ZoneEvent, error) {
	if !s.config.EnableRealTimeProcessing {
		return nil, nil
	}

	// Get zones for stream
	zones, err := s.GetZonesByStream(streamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get zones: %w", err)
	}

	var events []*domain.ZoneEvent
	detectionCenter := domain.Point{
		X: float64(detection.BoundingBox.X + detection.BoundingBox.Width/2),
		Y: float64(detection.BoundingBox.Y + detection.BoundingBox.Height/2),
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, zone := range zones {
		if !zone.IsActive {
			continue
		}

		// Check if detection is in zone
		inZone := zone.Polygon.Contains(detectionCenter)
		
		// Process based on zone type
		switch zone.Type {
		case domain.ZoneTypeInclude:
			if !inZone {
				continue // Skip detections outside include zones
			}
		case domain.ZoneTypeExclude:
			if inZone {
				continue // Skip detections inside exclude zones
			}
		}

		// Check object class filter
		if len(zone.Rules.ObjectClasses) > 0 {
			classMatch := false
			for _, class := range zone.Rules.ObjectClasses {
				if class == detection.Class {
					classMatch = true
					break
				}
			}
			if !classMatch {
				continue
			}
		}

		// Check confidence threshold
		if detection.Confidence < zone.Rules.MinConfidence {
			continue
		}

		// Check time restrictions
		if zone.Rules.IsTimeRestricted(time.Now()) {
			continue
		}

		// Process zone events
		zoneEvents := s.processZoneDetection(zone, detection, detectionCenter, inZone)
		events = append(events, zoneEvents...)
	}

	return events, nil
}

// processZoneDetection processes a detection for a specific zone
func (s *Service) processZoneDetection(zone *domain.DetectionZone, detection *domain.DetectedObject, center domain.Point, inZone bool) []*domain.ZoneEvent {
	var events []*domain.ZoneEvent
	now := time.Now()

	// Get or create object presence
	zoneOccupancy, exists := s.occupancy[zone.ID]
	if !exists {
		zoneOccupancy = make(map[string]*ObjectPresence)
		s.occupancy[zone.ID] = zoneOccupancy
	}

	presence, wasPresent := zoneOccupancy[detection.ID]

	if inZone {
		if !wasPresent {
			// Object entered zone
			presence = &ObjectPresence{
				ObjectID:    detection.ID,
				ZoneID:      zone.ID,
				EntryTime:   now,
				LastSeen:    now,
				Position:    center,
				BoundingBox: detection.BoundingBox,
				ObjectClass: detection.Class,
				Confidence:  detection.Confidence,
			}
			zoneOccupancy[detection.ID] = presence

			// Generate entry event
			if zone.Rules.AlertOnEntry {
				events = append(events, &domain.ZoneEvent{
					ID:          uuid.New().String(),
					ZoneID:      zone.ID,
					StreamID:    zone.StreamID,
					EventType:   domain.ZoneEventEntry,
					ObjectID:    detection.ID,
					ObjectClass: detection.Class,
					Confidence:  detection.Confidence,
					Position:    center,
					BoundingBox: detection.BoundingBox,
					Timestamp:   now,
				})
			}

			// Update statistics
			s.updateZoneStatistics(zone.ID, domain.ZoneEventEntry, detection.Class)

		} else {
			// Object still in zone - update presence
			presence.LastSeen = now
			presence.Position = center
			presence.BoundingBox = detection.BoundingBox
			presence.Confidence = detection.Confidence
			presence.DwellTime = now.Sub(presence.EntryTime)

			// Check for loitering
			if zone.Rules.AlertOnLoitering && 
			   zone.Rules.MinDwellTime > 0 && 
			   presence.DwellTime >= zone.Rules.MinDwellTime {
				events = append(events, &domain.ZoneEvent{
					ID:          uuid.New().String(),
					ZoneID:      zone.ID,
					StreamID:    zone.StreamID,
					EventType:   domain.ZoneEventLoitering,
					ObjectID:    detection.ID,
					ObjectClass: detection.Class,
					Confidence:  detection.Confidence,
					Position:    center,
					BoundingBox: detection.BoundingBox,
					DwellTime:   presence.DwellTime,
					Timestamp:   now,
				})
			}

			// Check for max dwell time violation
			if zone.Rules.MaxDwellTime > 0 && 
			   presence.DwellTime >= zone.Rules.MaxDwellTime {
				events = append(events, &domain.ZoneEvent{
					ID:          uuid.New().String(),
					ZoneID:      zone.ID,
					StreamID:    zone.StreamID,
					EventType:   domain.ZoneEventViolation,
					ObjectID:    detection.ID,
					ObjectClass: detection.Class,
					Confidence:  detection.Confidence,
					Position:    center,
					BoundingBox: detection.BoundingBox,
					DwellTime:   presence.DwellTime,
					Timestamp:   now,
					Metadata: map[string]interface{}{
						"violation_type": "max_dwell_time_exceeded",
						"max_dwell_time": zone.Rules.MaxDwellTime.String(),
					},
				})
			}
		}
	} else if wasPresent {
		// Object exited zone
		dwellTime := now.Sub(presence.EntryTime)

		// Generate exit event
		if zone.Rules.AlertOnExit {
			events = append(events, &domain.ZoneEvent{
				ID:          uuid.New().String(),
				ZoneID:      zone.ID,
				StreamID:    zone.StreamID,
				EventType:   domain.ZoneEventExit,
				ObjectID:    detection.ID,
				ObjectClass: detection.Class,
				Confidence:  detection.Confidence,
				Position:    center,
				BoundingBox: detection.BoundingBox,
				DwellTime:   dwellTime,
				Timestamp:   now,
			})
		}

		// Remove from occupancy
		delete(zoneOccupancy, detection.ID)

		// Update statistics
		s.updateZoneStatistics(zone.ID, domain.ZoneEventExit, detection.Class)
	}

	// Check for crowding
	if zone.Rules.AlertOnCrowding && 
	   zone.Rules.MaxObjects > 0 && 
	   len(zoneOccupancy) > zone.Rules.MaxObjects {
		events = append(events, &domain.ZoneEvent{
			ID:          uuid.New().String(),
			ZoneID:      zone.ID,
			StreamID:    zone.StreamID,
			EventType:   domain.ZoneEventCrowding,
			ObjectID:    detection.ID,
			ObjectClass: detection.Class,
			Confidence:  detection.Confidence,
			Position:    center,
			BoundingBox: detection.BoundingBox,
			Timestamp:   now,
			Metadata: map[string]interface{}{
				"current_occupancy": len(zoneOccupancy),
				"max_objects":       zone.Rules.MaxObjects,
			},
		})
	}

	// Store events in repository
	for _, event := range events {
		if err := s.repository.CreateZoneEvent(event); err != nil {
			s.logger.Error("Failed to store zone event",
				zap.String("zone_id", zone.ID),
				zap.String("event_type", string(event.EventType)),
				zap.Error(err))
		}
	}

	return events
}

// updateZoneStatistics updates zone statistics
func (s *Service) updateZoneStatistics(zoneID string, eventType domain.ZoneEventType, objectClass string) {
	stats, exists := s.statistics[zoneID]
	if !exists {
		return
	}

	now := time.Now()
	hour := now.Hour()
	day := now.Format("2006-01-02")

	switch eventType {
	case domain.ZoneEventEntry:
		stats.TotalEntries++
		stats.CurrentOccupancy++
		if stats.CurrentOccupancy > stats.MaxOccupancy {
			stats.MaxOccupancy = stats.CurrentOccupancy
		}
	case domain.ZoneEventExit:
		stats.TotalExits++
		if stats.CurrentOccupancy > 0 {
			stats.CurrentOccupancy--
		}
	}

	// Update object class counts
	if stats.ObjectCounts == nil {
		stats.ObjectCounts = make(map[string]int64)
	}
	stats.ObjectCounts[objectClass]++

	// Update hourly stats
	if stats.HourlyStats == nil {
		stats.HourlyStats = make(map[int]int64)
	}
	stats.HourlyStats[hour]++

	// Update daily stats
	if stats.DailyStats == nil {
		stats.DailyStats = make(map[string]int64)
	}
	stats.DailyStats[day]++

	stats.LastActivity = now
}

// CheckZoneViolations checks for zone violations across all detections
func (s *Service) CheckZoneViolations(streamID string, detections []*domain.DetectedObject) ([]*domain.ZoneEvent, error) {
	var allEvents []*domain.ZoneEvent

	for _, detection := range detections {
		events, err := s.ProcessDetection(streamID, detection)
		if err != nil {
			s.logger.Error("Failed to process detection for zones",
				zap.String("stream_id", streamID),
				zap.String("detection_id", detection.ID),
				zap.Error(err))
			continue
		}
		allEvents = append(allEvents, events...)
	}

	return allEvents, nil
}

// UpdateZoneOccupancy updates zone occupancy information
func (s *Service) UpdateZoneOccupancy(streamID string) error {
	// This would typically be called periodically to clean up stale presence data
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	staleThreshold := 30 * time.Second // Objects not seen for 30 seconds are considered gone

	for zoneID, occupancy := range s.occupancy {
		for objectID, presence := range occupancy {
			if now.Sub(presence.LastSeen) > staleThreshold {
				// Object is stale, remove it
				delete(occupancy, objectID)
				
				// Update statistics
				s.updateZoneStatistics(zoneID, domain.ZoneEventExit, presence.ObjectClass)
				
				s.logger.Debug("Removed stale object from zone",
					zap.String("zone_id", zoneID),
					zap.String("object_id", objectID),
					zap.Duration("last_seen", now.Sub(presence.LastSeen)))
			}
		}
	}

	return nil
}

// GetZoneStatistics retrieves statistics for a zone
func (s *Service) GetZoneStatistics(zoneID string) (*domain.ZoneStatistics, error) {
	s.mutex.RLock()
	stats, exists := s.statistics[zoneID]
	s.mutex.RUnlock()

	if exists {
		return stats, nil
	}

	// Get from repository
	return s.repository.GetZoneStatistics(zoneID)
}

// GetZoneAnalytics generates analytics for a zone
func (s *Service) GetZoneAnalytics(zoneID string, timeRange domain.TimeRange) (*domain.ZoneAnalytics, error) {
	// Get zone events for the time range
	events, err := s.repository.GetZoneEventsByTimeRange(zoneID, timeRange.Start, timeRange.End)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone events: %w", err)
	}

	analytics := &domain.ZoneAnalytics{
		ZoneID:                zoneID,
		TimeRange:             timeRange,
		ObjectClassBreakdown:  make(map[string]int64),
		HourlyDistribution:    make(map[int]int64),
		DwellTimeDistribution: make(map[string]int64),
	}

	// Process events to generate analytics
	uniqueObjects := make(map[string]bool)
	var totalOccupancy int64
	var occupancyCount int64

	for _, event := range events {
		analytics.TotalDetections++
		uniqueObjects[event.ObjectID] = true

		// Track object classes
		analytics.ObjectClassBreakdown[event.ObjectClass]++

		// Track hourly distribution
		hour := event.Timestamp.Hour()
		analytics.HourlyDistribution[hour]++

		// Track violations and alerts
		switch event.EventType {
		case domain.ZoneEventViolation:
			analytics.ViolationCount++
		case domain.ZoneEventEntry, domain.ZoneEventExit, domain.ZoneEventLoitering:
			analytics.AlertCount++
		}

		// Track dwell time distribution
		if event.DwellTime > 0 {
			dwellBucket := s.getDwellTimeBucket(event.DwellTime)
			analytics.DwellTimeDistribution[dwellBucket]++
		}
	}

	analytics.UniqueObjects = int64(len(uniqueObjects))

	// Calculate averages
	if occupancyCount > 0 {
		analytics.AverageOccupancy = float64(totalOccupancy) / float64(occupancyCount)
	}

	// Calculate entry/exit ratio
	var entries, exits int64
	for _, event := range events {
		switch event.EventType {
		case domain.ZoneEventEntry:
			entries++
		case domain.ZoneEventExit:
			exits++
		}
	}

	if exits > 0 {
		analytics.EntryExitRatio = float64(entries) / float64(exits)
	}

	return analytics, nil
}

// getDwellTimeBucket categorizes dwell time into buckets
func (s *Service) getDwellTimeBucket(dwellTime time.Duration) string {
	minutes := int(dwellTime.Minutes())
	
	switch {
	case minutes < 1:
		return "< 1 min"
	case minutes < 5:
		return "1-5 min"
	case minutes < 15:
		return "5-15 min"
	case minutes < 30:
		return "15-30 min"
	case minutes < 60:
		return "30-60 min"
	default:
		return "> 1 hour"
	}
}

// GenerateZoneReport generates a report for a zone
func (s *Service) GenerateZoneReport(zoneID string, reportType domain.ReportType, timeRange domain.TimeRange) (*domain.ZoneReport, error) {
	analytics, err := s.GetZoneAnalytics(zoneID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %w", err)
	}

	report := &domain.ZoneReport{
		ID:          uuid.New().String(),
		ZoneID:      zoneID,
		ReportType:  reportType,
		TimeRange:   timeRange,
		GeneratedAt: time.Now(),
		Data:        make(map[string]interface{}),
	}

	// Generate report data based on type
	switch reportType {
	case domain.ReportTypeOccupancy:
		report.Data["average_occupancy"] = analytics.AverageOccupancy
		report.Data["peak_occupancy"] = analytics.PeakOccupancy
		report.Data["peak_occupancy_time"] = analytics.PeakOccupancyTime
		report.Data["hourly_distribution"] = analytics.HourlyDistribution

	case domain.ReportTypeTraffic:
		report.Data["total_detections"] = analytics.TotalDetections
		report.Data["unique_objects"] = analytics.UniqueObjects
		report.Data["object_class_breakdown"] = analytics.ObjectClassBreakdown
		report.Data["entry_exit_ratio"] = analytics.EntryExitRatio

	case domain.ReportTypeViolations:
		report.Data["violation_count"] = analytics.ViolationCount
		report.Data["alert_count"] = analytics.AlertCount

	case domain.ReportTypeDwellTime:
		report.Data["dwell_time_distribution"] = analytics.DwellTimeDistribution

	case domain.ReportTypeSummary:
		report.Data = map[string]interface{}{
			"total_detections":       analytics.TotalDetections,
			"unique_objects":         analytics.UniqueObjects,
			"average_occupancy":      analytics.AverageOccupancy,
			"peak_occupancy":         analytics.PeakOccupancy,
			"violation_count":        analytics.ViolationCount,
			"alert_count":            analytics.AlertCount,
			"object_class_breakdown": analytics.ObjectClassBreakdown,
			"entry_exit_ratio":       analytics.EntryExitRatio,
		}
	}

	return report, nil
}
