package domain

import (
	"fmt"
	"math"
	"time"
)

// DetectionZone represents a region of interest for object detection
type DetectionZone struct {
	ID          string    `json:"id" db:"id"`
	StreamID    string    `json:"stream_id" db:"stream_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        ZoneType  `json:"type" db:"type"`
	Polygon     Polygon   `json:"polygon" db:"polygon"`
	Rules       ZoneRules `json:"rules" db:"rules"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ZoneType defines the type of detection zone
type ZoneType string

const (
	ZoneTypeInclude    ZoneType = "include"    // Only detect objects within this zone
	ZoneTypeExclude    ZoneType = "exclude"    // Exclude objects within this zone
	ZoneTypeAlert      ZoneType = "alert"      // Generate alerts for objects in this zone
	ZoneTypeCount      ZoneType = "count"      // Count objects entering/exiting this zone
	ZoneTypeRestricted ZoneType = "restricted" // Restricted area - generate alerts for unauthorized access
	ZoneTypeMonitor    ZoneType = "monitor"    // Monitor for specific behaviors in this zone
)

// Polygon represents a polygon defined by a series of points
type Polygon struct {
	Points []Point `json:"points"`
}

// Point represents a 2D coordinate
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ZoneRules defines detection rules for a zone
type ZoneRules struct {
	ObjectClasses      []string      `json:"object_classes"`       // Classes to detect (empty = all)
	MinConfidence      float64       `json:"min_confidence"`       // Minimum confidence threshold
	MaxObjects         int           `json:"max_objects"`          // Maximum objects allowed (-1 = unlimited)
	MinDwellTime       time.Duration `json:"min_dwell_time"`       // Minimum time to trigger alert
	MaxDwellTime       time.Duration `json:"max_dwell_time"`       // Maximum allowed dwell time
	AlertOnEntry       bool          `json:"alert_on_entry"`       // Alert when object enters zone
	AlertOnExit        bool          `json:"alert_on_exit"`        // Alert when object exits zone
	AlertOnLoitering   bool          `json:"alert_on_loitering"`   // Alert on loitering behavior
	AlertOnCrowding    bool          `json:"alert_on_crowding"`    // Alert when too many objects
	CountDirection     CountDirection `json:"count_direction"`     // Direction for counting
	IgnoreStationaryObjects bool     `json:"ignore_stationary"`   // Ignore objects that don't move
	TimeRestrictions   []TimeWindow  `json:"time_restrictions"`   // Time-based restrictions
}

// CountDirection defines counting direction for zones
type CountDirection string

const (
	CountDirectionBoth CountDirection = "both"
	CountDirectionIn   CountDirection = "in"
	CountDirectionOut  CountDirection = "out"
)

// TimeWindow defines a time restriction window
type TimeWindow struct {
	StartTime string `json:"start_time"` // HH:MM format
	EndTime   string `json:"end_time"`   // HH:MM format
	Days      []int  `json:"days"`       // 0=Sunday, 1=Monday, etc.
}

// ZoneEvent represents an event that occurred in a detection zone
type ZoneEvent struct {
	ID           string          `json:"id"`
	ZoneID       string          `json:"zone_id"`
	StreamID     string          `json:"stream_id"`
	EventType    ZoneEventType   `json:"event_type"`
	ObjectID     string          `json:"object_id"`
	ObjectClass  string          `json:"object_class"`
	Confidence   float64         `json:"confidence"`
	Position     Point           `json:"position"`
	BoundingBox  Rectangle       `json:"bounding_box"`
	DwellTime    time.Duration   `json:"dwell_time"`
	Metadata     map[string]interface{} `json:"metadata"`
	Timestamp    time.Time       `json:"timestamp"`
}

// ZoneEventType defines types of zone events
type ZoneEventType string

const (
	ZoneEventEntry      ZoneEventType = "entry"
	ZoneEventExit       ZoneEventType = "exit"
	ZoneEventLoitering  ZoneEventType = "loitering"
	ZoneEventCrowding   ZoneEventType = "crowding"
	ZoneEventViolation  ZoneEventType = "violation"
	ZoneEventCount      ZoneEventType = "count"
	ZoneEventDwellTime  ZoneEventType = "dwell_time"
)

// ZoneStatistics tracks statistics for a detection zone
type ZoneStatistics struct {
	ZoneID           string            `json:"zone_id"`
	StreamID         string            `json:"stream_id"`
	TotalEntries     int64             `json:"total_entries"`
	TotalExits       int64             `json:"total_exits"`
	CurrentOccupancy int               `json:"current_occupancy"`
	MaxOccupancy     int               `json:"max_occupancy"`
	AverageDwellTime time.Duration     `json:"average_dwell_time"`
	ObjectCounts     map[string]int64  `json:"object_counts"`
	HourlyStats      map[int]int64     `json:"hourly_stats"`
	DailyStats       map[string]int64  `json:"daily_stats"`
	LastActivity     time.Time         `json:"last_activity"`
	StartTime        time.Time         `json:"start_time"`
}

// ZoneRepository defines the interface for zone data access
type ZoneRepository interface {
	// Zone management
	CreateZone(zone *DetectionZone) error
	GetZone(id string) (*DetectionZone, error)
	GetZonesByStream(streamID string) ([]*DetectionZone, error)
	UpdateZone(zone *DetectionZone) error
	DeleteZone(id string) error
	ListZones(limit, offset int) ([]*DetectionZone, error)

	// Zone events
	CreateZoneEvent(event *ZoneEvent) error
	GetZoneEvents(zoneID string, limit, offset int) ([]*ZoneEvent, error)
	GetZoneEventsByTimeRange(zoneID string, start, end time.Time) ([]*ZoneEvent, error)

	// Zone statistics
	GetZoneStatistics(zoneID string) (*ZoneStatistics, error)
	UpdateZoneStatistics(stats *ZoneStatistics) error
}

// ZoneService defines the interface for zone business logic
type ZoneService interface {
	// Zone management
	CreateZone(zone *DetectionZone) error
	GetZone(id string) (*DetectionZone, error)
	GetZonesByStream(streamID string) ([]*DetectionZone, error)
	UpdateZone(zone *DetectionZone) error
	DeleteZone(id string) error
	ListZones(limit, offset int) ([]*DetectionZone, error)

	// Zone processing
	ProcessDetection(streamID string, detection *DetectedObject) ([]*ZoneEvent, error)
	CheckZoneViolations(streamID string, detections []*DetectedObject) ([]*ZoneEvent, error)
	UpdateZoneOccupancy(streamID string) error

	// Zone analytics
	GetZoneStatistics(zoneID string) (*ZoneStatistics, error)
	GetZoneAnalytics(zoneID string, timeRange TimeRange) (*ZoneAnalytics, error)
	GenerateZoneReport(zoneID string, reportType ReportType, timeRange TimeRange) (*ZoneReport, error)
}

// ZoneAnalytics provides analytical data for zones
type ZoneAnalytics struct {
	ZoneID              string                    `json:"zone_id"`
	TimeRange           TimeRange                 `json:"time_range"`
	TotalDetections     int64                     `json:"total_detections"`
	UniqueObjects       int64                     `json:"unique_objects"`
	AverageOccupancy    float64                   `json:"average_occupancy"`
	PeakOccupancy       int                       `json:"peak_occupancy"`
	PeakOccupancyTime   time.Time                 `json:"peak_occupancy_time"`
	ObjectClassBreakdown map[string]int64         `json:"object_class_breakdown"`
	HourlyDistribution  map[int]int64             `json:"hourly_distribution"`
	DwellTimeDistribution map[string]int64        `json:"dwell_time_distribution"`
	EntryExitRatio      float64                   `json:"entry_exit_ratio"`
	ViolationCount      int64                     `json:"violation_count"`
	AlertCount          int64                     `json:"alert_count"`
}

// ZoneReport represents a generated report for a zone
type ZoneReport struct {
	ID          string                 `json:"id"`
	ZoneID      string                 `json:"zone_id"`
	ReportType  ReportType             `json:"report_type"`
	TimeRange   TimeRange              `json:"time_range"`
	Data        map[string]interface{} `json:"data"`
	GeneratedAt time.Time              `json:"generated_at"`
	GeneratedBy string                 `json:"generated_by"`
}

// ReportType defines types of zone reports
type ReportType string

const (
	ReportTypeOccupancy   ReportType = "occupancy"
	ReportTypeTraffic     ReportType = "traffic"
	ReportTypeViolations  ReportType = "violations"
	ReportTypeDwellTime   ReportType = "dwell_time"
	ReportTypeHeatMap     ReportType = "heat_map"
	ReportTypeSummary     ReportType = "summary"
)

// TimeRange defines a time range for analytics
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Polygon methods

// Contains checks if a point is inside the polygon using ray casting algorithm
func (p *Polygon) Contains(point Point) bool {
	if len(p.Points) < 3 {
		return false
	}

	inside := false
	j := len(p.Points) - 1

	for i := 0; i < len(p.Points); i++ {
		xi, yi := p.Points[i].X, p.Points[i].Y
		xj, yj := p.Points[j].X, p.Points[j].Y

		if ((yi > point.Y) != (yj > point.Y)) &&
			(point.X < (xj-xi)*(point.Y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}

	return inside
}

// Area calculates the area of the polygon using the shoelace formula
func (p *Polygon) Area() float64 {
	if len(p.Points) < 3 {
		return 0
	}

	area := 0.0
	j := len(p.Points) - 1

	for i := 0; i < len(p.Points); i++ {
		area += (p.Points[j].X + p.Points[i].X) * (p.Points[j].Y - p.Points[i].Y)
		j = i
	}

	return math.Abs(area) / 2.0
}

// Centroid calculates the centroid of the polygon
func (p *Polygon) Centroid() Point {
	if len(p.Points) == 0 {
		return Point{0, 0}
	}

	var cx, cy float64
	for _, point := range p.Points {
		cx += point.X
		cy += point.Y
	}

	return Point{
		X: cx / float64(len(p.Points)),
		Y: cy / float64(len(p.Points)),
	}
}

// BoundingBox returns the bounding box of the polygon
func (p *Polygon) BoundingBox() Rectangle {
	if len(p.Points) == 0 {
		return Rectangle{}
	}

	minX, minY := p.Points[0].X, p.Points[0].Y
	maxX, maxY := p.Points[0].X, p.Points[0].Y

	for _, point := range p.Points[1:] {
		if point.X < minX {
			minX = point.X
		}
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}

	return Rectangle{
		X:      int(minX),
		Y:      int(minY),
		Width:  int(maxX - minX),
		Height: int(maxY - minY),
	}
}

// Validate validates the polygon
func (p *Polygon) Validate() error {
	if len(p.Points) < 3 {
		return fmt.Errorf("polygon must have at least 3 points")
	}

	// Check for duplicate consecutive points
	for i := 0; i < len(p.Points); i++ {
		next := (i + 1) % len(p.Points)
		if p.Points[i].X == p.Points[next].X && p.Points[i].Y == p.Points[next].Y {
			return fmt.Errorf("polygon has duplicate consecutive points at index %d", i)
		}
	}

	return nil
}

// Point methods

// Distance calculates the Euclidean distance between two points
func (p Point) Distance(other Point) float64 {
	dx := p.X - other.X
	dy := p.Y - other.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// DetectionZone methods

// Validate validates the detection zone
func (z *DetectionZone) Validate() error {
	if z.ID == "" {
		return fmt.Errorf("zone ID is required")
	}
	if z.StreamID == "" {
		return fmt.Errorf("stream ID is required")
	}
	if z.Name == "" {
		return fmt.Errorf("zone name is required")
	}

	if err := z.Polygon.Validate(); err != nil {
		return fmt.Errorf("invalid polygon: %w", err)
	}

	return z.Rules.Validate()
}

// ZoneRules methods

// Validate validates the zone rules
func (r *ZoneRules) Validate() error {
	if r.MinConfidence < 0 || r.MinConfidence > 1 {
		return fmt.Errorf("min confidence must be between 0 and 1")
	}

	if r.MaxObjects < -1 {
		return fmt.Errorf("max objects must be -1 (unlimited) or positive")
	}

	if r.MinDwellTime < 0 {
		return fmt.Errorf("min dwell time cannot be negative")
	}

	if r.MaxDwellTime > 0 && r.MaxDwellTime < r.MinDwellTime {
		return fmt.Errorf("max dwell time must be greater than min dwell time")
	}

	return nil
}

// IsTimeRestricted checks if the zone is restricted at the given time
func (r *ZoneRules) IsTimeRestricted(t time.Time) bool {
	if len(r.TimeRestrictions) == 0 {
		return false
	}

	weekday := int(t.Weekday())
	timeStr := t.Format("15:04")

	for _, window := range r.TimeRestrictions {
		// Check if current day is in the restriction
		dayMatch := false
		for _, day := range window.Days {
			if day == weekday {
				dayMatch = true
				break
			}
		}

		if dayMatch {
			// Check if current time is in the restriction window
			if timeStr >= window.StartTime && timeStr <= window.EndTime {
				return true
			}
		}
	}

	return false
}
