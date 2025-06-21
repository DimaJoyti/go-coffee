package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPolygon_Contains(t *testing.T) {
	// Create a simple rectangle polygon
	polygon := Polygon{
		Points: []Point{
			{X: 0, Y: 0},
			{X: 10, Y: 0},
			{X: 10, Y: 10},
			{X: 0, Y: 10},
		},
	}

	tests := []struct {
		name     string
		point    Point
		expected bool
	}{
		{
			name:     "point inside polygon",
			point:    Point{X: 5, Y: 5},
			expected: true,
		},
		{
			name:     "point outside polygon",
			point:    Point{X: 15, Y: 15},
			expected: false,
		},
		{
			name:     "point on edge",
			point:    Point{X: 0, Y: 5},
			expected: false, // Edge points are typically considered outside
		},
		{
			name:     "point at corner",
			point:    Point{X: 0, Y: 0},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := polygon.Contains(tt.point)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPolygon_Area(t *testing.T) {
	tests := []struct {
		name     string
		polygon  Polygon
		expected float64
	}{
		{
			name: "rectangle",
			polygon: Polygon{
				Points: []Point{
					{X: 0, Y: 0},
					{X: 10, Y: 0},
					{X: 10, Y: 10},
					{X: 0, Y: 10},
				},
			},
			expected: 100.0,
		},
		{
			name: "triangle",
			polygon: Polygon{
				Points: []Point{
					{X: 0, Y: 0},
					{X: 10, Y: 0},
					{X: 5, Y: 10},
				},
			},
			expected: 50.0,
		},
		{
			name: "empty polygon",
			polygon: Polygon{
				Points: []Point{},
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			area := tt.polygon.Area()
			assert.InDelta(t, tt.expected, area, 0.001)
		})
	}
}

func TestPolygon_Centroid(t *testing.T) {
	polygon := Polygon{
		Points: []Point{
			{X: 0, Y: 0},
			{X: 10, Y: 0},
			{X: 10, Y: 10},
			{X: 0, Y: 10},
		},
	}

	centroid := polygon.Centroid()
	assert.Equal(t, 5.0, centroid.X)
	assert.Equal(t, 5.0, centroid.Y)
}

func TestPolygon_BoundingBox(t *testing.T) {
	polygon := Polygon{
		Points: []Point{
			{X: 2, Y: 3},
			{X: 8, Y: 1},
			{X: 12, Y: 7},
			{X: 4, Y: 9},
		},
	}

	bbox := polygon.BoundingBox()
	assert.Equal(t, 2, bbox.X)
	assert.Equal(t, 1, bbox.Y)
	assert.Equal(t, 10, bbox.Width)  // 12 - 2
	assert.Equal(t, 8, bbox.Height)  // 9 - 1
}

func TestPolygon_Validate(t *testing.T) {
	tests := []struct {
		name      string
		polygon   Polygon
		expectErr bool
	}{
		{
			name: "valid polygon",
			polygon: Polygon{
				Points: []Point{
					{X: 0, Y: 0},
					{X: 10, Y: 0},
					{X: 10, Y: 10},
					{X: 0, Y: 10},
				},
			},
			expectErr: false,
		},
		{
			name: "too few points",
			polygon: Polygon{
				Points: []Point{
					{X: 0, Y: 0},
					{X: 10, Y: 0},
				},
			},
			expectErr: true,
		},
		{
			name: "duplicate consecutive points",
			polygon: Polygon{
				Points: []Point{
					{X: 0, Y: 0},
					{X: 0, Y: 0}, // Duplicate
					{X: 10, Y: 10},
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.polygon.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPoint_Distance(t *testing.T) {
	p1 := Point{X: 0, Y: 0}
	p2 := Point{X: 3, Y: 4}

	distance := p1.Distance(p2)
	assert.InDelta(t, 5.0, distance, 0.001) // 3-4-5 triangle
}

func TestDetectionZone_Validate(t *testing.T) {
	validZone := &DetectionZone{
		ID:       "zone-1",
		StreamID: "stream-1",
		Name:     "Test Zone",
		Type:     ZoneTypeAlert,
		Polygon: Polygon{
			Points: []Point{
				{X: 0, Y: 0},
				{X: 10, Y: 0},
				{X: 10, Y: 10},
				{X: 0, Y: 10},
			},
		},
		Rules: ZoneRules{
			MinConfidence: 0.5,
			MaxObjects:   10,
		},
	}

	tests := []struct {
		name      string
		zone      *DetectionZone
		expectErr bool
	}{
		{
			name:      "valid zone",
			zone:      validZone,
			expectErr: false,
		},
		{
			name: "missing ID",
			zone: &DetectionZone{
				StreamID: "stream-1",
				Name:     "Test Zone",
				Type:     ZoneTypeAlert,
				Polygon:  validZone.Polygon,
				Rules:    validZone.Rules,
			},
			expectErr: true,
		},
		{
			name: "missing stream ID",
			zone: &DetectionZone{
				ID:      "zone-1",
				Name:    "Test Zone",
				Type:    ZoneTypeAlert,
				Polygon: validZone.Polygon,
				Rules:   validZone.Rules,
			},
			expectErr: true,
		},
		{
			name: "missing name",
			zone: &DetectionZone{
				ID:       "zone-1",
				StreamID: "stream-1",
				Type:     ZoneTypeAlert,
				Polygon:  validZone.Polygon,
				Rules:    validZone.Rules,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.zone.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestZoneRules_Validate(t *testing.T) {
	tests := []struct {
		name      string
		rules     ZoneRules
		expectErr bool
	}{
		{
			name: "valid rules",
			rules: ZoneRules{
				MinConfidence: 0.5,
				MaxObjects:   10,
				MinDwellTime:  5 * time.Second,
				MaxDwellTime:  60 * time.Second,
			},
			expectErr: false,
		},
		{
			name: "invalid confidence - too low",
			rules: ZoneRules{
				MinConfidence: -0.1,
			},
			expectErr: true,
		},
		{
			name: "invalid confidence - too high",
			rules: ZoneRules{
				MinConfidence: 1.1,
			},
			expectErr: true,
		},
		{
			name: "invalid max objects",
			rules: ZoneRules{
				MaxObjects: -2,
			},
			expectErr: true,
		},
		{
			name: "invalid dwell time - negative min",
			rules: ZoneRules{
				MinDwellTime: -5 * time.Second,
			},
			expectErr: true,
		},
		{
			name: "invalid dwell time - max less than min",
			rules: ZoneRules{
				MinDwellTime: 60 * time.Second,
				MaxDwellTime: 30 * time.Second,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rules.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestZoneRules_IsTimeRestricted(t *testing.T) {
	rules := ZoneRules{
		TimeRestrictions: []TimeWindow{
			{
				StartTime: "09:00",
				EndTime:   "17:00",
				Days:      []int{1, 2, 3, 4, 5}, // Monday to Friday
			},
			{
				StartTime: "22:00",
				EndTime:   "06:00",
				Days:      []int{0, 6}, // Sunday and Saturday
			},
		},
	}

	tests := []struct {
		name     string
		time     time.Time
		expected bool
	}{
		{
			name:     "weekday during business hours",
			time:     time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC), // Monday 14:30
			expected: true,
		},
		{
			name:     "weekday outside business hours",
			time:     time.Date(2024, 1, 15, 20, 30, 0, 0, time.UTC), // Monday 20:30
			expected: false,
		},
		{
			name:     "weekend during night restriction",
			time:     time.Date(2024, 1, 14, 23, 30, 0, 0, time.UTC), // Sunday 23:30
			expected: true,
		},
		{
			name:     "weekend outside restrictions",
			time:     time.Date(2024, 1, 14, 12, 30, 0, 0, time.UTC), // Sunday 12:30
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rules.IsTimeRestricted(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestZoneTypes(t *testing.T) {
	// Test that all zone types are defined
	types := []ZoneType{
		ZoneTypeInclude,
		ZoneTypeExclude,
		ZoneTypeAlert,
		ZoneTypeCount,
		ZoneTypeRestricted,
		ZoneTypeMonitor,
	}

	for _, zoneType := range types {
		assert.NotEmpty(t, string(zoneType))
	}
}

func TestZoneEventTypes(t *testing.T) {
	// Test that all event types are defined
	eventTypes := []ZoneEventType{
		ZoneEventEntry,
		ZoneEventExit,
		ZoneEventLoitering,
		ZoneEventCrowding,
		ZoneEventViolation,
		ZoneEventCount,
		ZoneEventDwellTime,
	}

	for _, eventType := range eventTypes {
		assert.NotEmpty(t, string(eventType))
	}
}

func TestReportTypes(t *testing.T) {
	// Test that all report types are defined
	reportTypes := []ReportType{
		ReportTypeOccupancy,
		ReportTypeTraffic,
		ReportTypeViolations,
		ReportTypeDwellTime,
		ReportTypeHeatMap,
		ReportTypeSummary,
	}

	for _, reportType := range reportTypes {
		assert.NotEmpty(t, string(reportType))
	}
}

func TestCountDirection(t *testing.T) {
	// Test that all count directions are defined
	directions := []CountDirection{
		CountDirectionBoth,
		CountDirectionIn,
		CountDirectionOut,
	}

	for _, direction := range directions {
		assert.NotEmpty(t, string(direction))
	}
}
