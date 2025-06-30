package entities

import (
	"time"

	"github.com/google/uuid"
)

// Location represents a physical location where inventory is stored
type Location struct {
	ID                uuid.UUID              `json:"id" redis:"id"`
	Code              string                 `json:"code" redis:"code"`
	Name              string                 `json:"name" redis:"name"`
	Type              LocationType           `json:"type" redis:"type"`
	Status            LocationStatus         `json:"status" redis:"status"`
	Address           *Address               `json:"address,omitempty"`
	Coordinates       *Coordinates           `json:"coordinates,omitempty"`
	ParentLocationID  *uuid.UUID             `json:"parent_location_id,omitempty" redis:"parent_location_id"`
	ParentLocation    *Location              `json:"parent_location,omitempty"`
	ChildLocations    []*Location            `json:"child_locations,omitempty"`
	StorageZones      []*StorageZone         `json:"storage_zones,omitempty"`
	Capacity          *LocationCapacity      `json:"capacity,omitempty"`
	Environment       *EnvironmentConditions `json:"environment,omitempty"`
	Security          *SecurityInformation   `json:"security,omitempty"`
	Staff             []*LocationStaff       `json:"staff,omitempty"`
	Equipment         []*LocationEquipment   `json:"equipment,omitempty"`
	OperatingHours    *OperatingHours        `json:"operating_hours,omitempty"`
	ContactInfo       *ContactInformation    `json:"contact_info,omitempty"`
	Attributes        map[string]interface{} `json:"attributes" redis:"attributes"`
	Tags              []string               `json:"tags" redis:"tags"`
	Notes             string                 `json:"notes" redis:"notes"`
	IsActive          bool                   `json:"is_active" redis:"is_active"`
	IsDefault         bool                   `json:"is_default" redis:"is_default"`
	CreatedAt         time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" redis:"updated_at"`
	CreatedBy         string                 `json:"created_by" redis:"created_by"`
	UpdatedBy         string                 `json:"updated_by" redis:"updated_by"`
	Version           int64                  `json:"version" redis:"version"`
}

// LocationType defines the type of location
type LocationType string

const (
	LocationTypeWarehouse    LocationType = "warehouse"
	LocationTypeStore        LocationType = "store"
	LocationTypeKitchen      LocationType = "kitchen"
	LocationTypeStorage      LocationType = "storage"
	LocationTypeFreezer      LocationType = "freezer"
	LocationTypeRefrigerator LocationType = "refrigerator"
	LocationTypeDryStorage   LocationType = "dry_storage"
	LocationTypeOffice       LocationType = "office"
	LocationTypeProduction   LocationType = "production"
	LocationTypeDistribution LocationType = "distribution"
	LocationTypeRetail       LocationType = "retail"
	LocationTypeVirtual      LocationType = "virtual"
)

// LocationStatus defines the status of a location
type LocationStatus string

const (
	LocationStatusActive      LocationStatus = "active"
	LocationStatusInactive    LocationStatus = "inactive"
	LocationStatusMaintenance LocationStatus = "maintenance"
	LocationStatusClosed      LocationStatus = "closed"
	LocationStatusTemporary   LocationStatus = "temporary"
)

// Coordinates represents geographical coordinates
type Coordinates struct {
	Latitude  float64 `json:"latitude" redis:"latitude"`
	Longitude float64 `json:"longitude" redis:"longitude"`
	Altitude  float64 `json:"altitude" redis:"altitude"`
	Accuracy  float64 `json:"accuracy" redis:"accuracy"`
}

// StorageZone represents a storage zone within a location
type StorageZone struct {
	ID                uuid.UUID              `json:"id" redis:"id"`
	Code              string                 `json:"code" redis:"code"`
	Name              string                 `json:"name" redis:"name"`
	Type              StorageZoneType        `json:"type" redis:"type"`
	Status            string                 `json:"status" redis:"status"`
	Capacity          *ZoneCapacity          `json:"capacity,omitempty"`
	Environment       *EnvironmentConditions `json:"environment,omitempty"`
	AccessRestrictions []string              `json:"access_restrictions" redis:"access_restrictions"`
	StorageRules      []string               `json:"storage_rules" redis:"storage_rules"`
	Attributes        map[string]interface{} `json:"attributes" redis:"attributes"`
	IsActive          bool                   `json:"is_active" redis:"is_active"`
	CreatedAt         time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" redis:"updated_at"`
}

// StorageZoneType defines the type of storage zone
type StorageZoneType string

const (
	ZoneTypeDry         StorageZoneType = "dry"
	ZoneTypeRefrigerated StorageZoneType = "refrigerated"
	ZoneTypeFrozen      StorageZoneType = "frozen"
	ZoneTypeHazardous   StorageZoneType = "hazardous"
	ZoneTypeQuarantine  StorageZoneType = "quarantine"
	ZoneTypeReceiving   StorageZoneType = "receiving"
	ZoneTypeShipping    StorageZoneType = "shipping"
	ZoneTypeProduction  StorageZoneType = "production"
	ZoneTypeDisplay     StorageZoneType = "display"
)

// LocationCapacity represents the capacity information for a location
type LocationCapacity struct {
	TotalArea         float64                `json:"total_area" redis:"total_area"`
	StorageArea       float64                `json:"storage_area" redis:"storage_area"`
	UsableArea        float64                `json:"usable_area" redis:"usable_area"`
	MaxWeight         float64                `json:"max_weight" redis:"max_weight"`
	MaxVolume         float64                `json:"max_volume" redis:"max_volume"`
	MaxPallets        int                    `json:"max_pallets" redis:"max_pallets"`
	MaxSKUs           int                    `json:"max_skus" redis:"max_skus"`
	CurrentUtilization float64               `json:"current_utilization" redis:"current_utilization"`
	AreaUnit          string                 `json:"area_unit" redis:"area_unit"`
	WeightUnit        string                 `json:"weight_unit" redis:"weight_unit"`
	VolumeUnit        string                 `json:"volume_unit" redis:"volume_unit"`
	Attributes        map[string]interface{} `json:"attributes" redis:"attributes"`
}

// ZoneCapacity represents the capacity information for a storage zone
type ZoneCapacity struct {
	MaxWeight      float64 `json:"max_weight" redis:"max_weight"`
	MaxVolume      float64 `json:"max_volume" redis:"max_volume"`
	MaxItems       int     `json:"max_items" redis:"max_items"`
	CurrentWeight  float64 `json:"current_weight" redis:"current_weight"`
	CurrentVolume  float64 `json:"current_volume" redis:"current_volume"`
	CurrentItems   int     `json:"current_items" redis:"current_items"`
	WeightUnit     string  `json:"weight_unit" redis:"weight_unit"`
	VolumeUnit     string  `json:"volume_unit" redis:"volume_unit"`
	Utilization    float64 `json:"utilization" redis:"utilization"`
}

// EnvironmentConditions represents environmental conditions for storage
type EnvironmentConditions struct {
	Temperature     *TemperatureRange      `json:"temperature,omitempty"`
	Humidity        *HumidityRange         `json:"humidity,omitempty"`
	AirPressure     *PressureRange         `json:"air_pressure,omitempty"`
	LightLevel      *LightRange            `json:"light_level,omitempty"`
	AirQuality      *AirQualityInfo        `json:"air_quality,omitempty"`
	Ventilation     string                 `json:"ventilation" redis:"ventilation"`
	ClimateControl  bool                   `json:"climate_control" redis:"climate_control"`
	Monitoring      bool                   `json:"monitoring" redis:"monitoring"`
	Sensors         []*EnvironmentSensor   `json:"sensors,omitempty"`
	Attributes      map[string]interface{} `json:"attributes" redis:"attributes"`
	LastUpdated     time.Time              `json:"last_updated" redis:"last_updated"`
}

// PressureRange defines pressure range
type PressureRange struct {
	MinPressure float64 `json:"min_pressure" redis:"min_pressure"`
	MaxPressure float64 `json:"max_pressure" redis:"max_pressure"`
	Unit        string  `json:"unit" redis:"unit"`
}

// LightRange defines light level range
type LightRange struct {
	MinLux float64 `json:"min_lux" redis:"min_lux"`
	MaxLux float64 `json:"max_lux" redis:"max_lux"`
	Unit   string  `json:"unit" redis:"unit"`
}

// AirQualityInfo represents air quality information
type AirQualityInfo struct {
	CO2Level    float64 `json:"co2_level" redis:"co2_level"`
	O2Level     float64 `json:"o2_level" redis:"o2_level"`
	Particulates float64 `json:"particulates" redis:"particulates"`
	VOCs        float64 `json:"vocs" redis:"vocs"`
	AQI         int     `json:"aqi" redis:"aqi"`
	Status      string  `json:"status" redis:"status"`
}

// EnvironmentSensor represents an environmental monitoring sensor
type EnvironmentSensor struct {
	ID           uuid.UUID              `json:"id" redis:"id"`
	Type         string                 `json:"type" redis:"type"`
	Name         string                 `json:"name" redis:"name"`
	Location     string                 `json:"location" redis:"location"`
	Status       string                 `json:"status" redis:"status"`
	LastReading  float64                `json:"last_reading" redis:"last_reading"`
	Unit         string                 `json:"unit" redis:"unit"`
	MinThreshold float64                `json:"min_threshold" redis:"min_threshold"`
	MaxThreshold float64                `json:"max_threshold" redis:"max_threshold"`
	Calibrated   bool                   `json:"calibrated" redis:"calibrated"`
	LastCalibration *time.Time          `json:"last_calibration,omitempty" redis:"last_calibration"`
	Attributes   map[string]interface{} `json:"attributes" redis:"attributes"`
	IsActive     bool                   `json:"is_active" redis:"is_active"`
	CreatedAt    time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" redis:"updated_at"`
}

// SecurityInformation represents security information for a location
type SecurityInformation struct {
	SecurityLevel     string                 `json:"security_level" redis:"security_level"`
	AccessControl     bool                   `json:"access_control" redis:"access_control"`
	Surveillance      bool                   `json:"surveillance" redis:"surveillance"`
	AlarmSystem       bool                   `json:"alarm_system" redis:"alarm_system"`
	FireSafety        bool                   `json:"fire_safety" redis:"fire_safety"`
	SecurityPersonnel bool                   `json:"security_personnel" redis:"security_personnel"`
	AccessMethods     []string               `json:"access_methods" redis:"access_methods"`
	SecurityZones     []string               `json:"security_zones" redis:"security_zones"`
	EmergencyContacts []*Contact             `json:"emergency_contacts,omitempty"`
	Attributes        map[string]interface{} `json:"attributes" redis:"attributes"`
	LastAudit         *time.Time             `json:"last_audit,omitempty" redis:"last_audit"`
	NextAudit         *time.Time             `json:"next_audit,omitempty" redis:"next_audit"`
}

// LocationStaff represents staff assigned to a location
type LocationStaff struct {
	ID           uuid.UUID `json:"id" redis:"id"`
	UserID       uuid.UUID `json:"user_id" redis:"user_id"`
	Name         string    `json:"name" redis:"name"`
	Role         string    `json:"role" redis:"role"`
	Department   string    `json:"department" redis:"department"`
	Permissions  []string  `json:"permissions" redis:"permissions"`
	ShiftPattern string    `json:"shift_pattern" redis:"shift_pattern"`
	ContactInfo  *Contact  `json:"contact_info,omitempty"`
	IsActive     bool      `json:"is_active" redis:"is_active"`
	StartDate    time.Time `json:"start_date" redis:"start_date"`
	EndDate      *time.Time `json:"end_date,omitempty" redis:"end_date"`
	CreatedAt    time.Time `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" redis:"updated_at"`
}

// LocationEquipment represents equipment at a location
type LocationEquipment struct {
	ID             uuid.UUID              `json:"id" redis:"id"`
	EquipmentID    uuid.UUID              `json:"equipment_id" redis:"equipment_id"`
	Name           string                 `json:"name" redis:"name"`
	Type           string                 `json:"type" redis:"type"`
	Model          string                 `json:"model" redis:"model"`
	SerialNumber   string                 `json:"serial_number" redis:"serial_number"`
	Status         string                 `json:"status" redis:"status"`
	Zone           string                 `json:"zone" redis:"zone"`
	InstallDate    time.Time              `json:"install_date" redis:"install_date"`
	LastMaintenance *time.Time            `json:"last_maintenance,omitempty" redis:"last_maintenance"`
	NextMaintenance *time.Time            `json:"next_maintenance,omitempty" redis:"next_maintenance"`
	Specifications map[string]interface{} `json:"specifications" redis:"specifications"`
	IsActive       bool                   `json:"is_active" redis:"is_active"`
	CreatedAt      time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" redis:"updated_at"`
}

// OperatingHours represents operating hours for a location
type OperatingHours struct {
	Monday    *DayHours `json:"monday,omitempty"`
	Tuesday   *DayHours `json:"tuesday,omitempty"`
	Wednesday *DayHours `json:"wednesday,omitempty"`
	Thursday  *DayHours `json:"thursday,omitempty"`
	Friday    *DayHours `json:"friday,omitempty"`
	Saturday  *DayHours `json:"saturday,omitempty"`
	Sunday    *DayHours `json:"sunday,omitempty"`
	Holidays  []*HolidayHours `json:"holidays,omitempty"`
	TimeZone  string    `json:"time_zone" redis:"time_zone"`
	IsActive  bool      `json:"is_active" redis:"is_active"`
}

// DayHours represents operating hours for a specific day
type DayHours struct {
	IsOpen    bool   `json:"is_open" redis:"is_open"`
	OpenTime  string `json:"open_time" redis:"open_time"`
	CloseTime string `json:"close_time" redis:"close_time"`
	Breaks    []*BreakPeriod `json:"breaks,omitempty"`
}

// BreakPeriod represents a break period during operating hours
type BreakPeriod struct {
	StartTime string `json:"start_time" redis:"start_time"`
	EndTime   string `json:"end_time" redis:"end_time"`
	Type      string `json:"type" redis:"type"`
}

// HolidayHours represents special hours for holidays
type HolidayHours struct {
	Date        time.Time `json:"date" redis:"date"`
	Name        string    `json:"name" redis:"name"`
	IsClosed    bool      `json:"is_closed" redis:"is_closed"`
	SpecialHours *DayHours `json:"special_hours,omitempty"`
}

// NewLocation creates a new location with default values
func NewLocation(name, code string, locationType LocationType) *Location {
	now := time.Now()
	return &Location{
		ID:         uuid.New(),
		Code:       code,
		Name:       name,
		Type:       locationType,
		Status:     LocationStatusActive,
		Attributes: make(map[string]interface{}),
		Tags:       []string{},
		IsActive:   true,
		IsDefault:  false,
		CreatedAt:  now,
		UpdatedAt:  now,
		Version:    1,
	}
}

// Activate activates the location
func (l *Location) Activate() {
	l.IsActive = true
	l.Status = LocationStatusActive
	l.UpdatedAt = time.Now()
	l.Version++
}

// Deactivate deactivates the location
func (l *Location) Deactivate() {
	l.IsActive = false
	l.Status = LocationStatusInactive
	l.UpdatedAt = time.Now()
	l.Version++
}

// SetMaintenance sets the location to maintenance mode
func (l *Location) SetMaintenance() {
	l.Status = LocationStatusMaintenance
	l.UpdatedAt = time.Now()
	l.Version++
}

// AddStorageZone adds a storage zone to the location
func (l *Location) AddStorageZone(zone *StorageZone) {
	l.StorageZones = append(l.StorageZones, zone)
	l.UpdatedAt = time.Now()
	l.Version++
}

// GetStorageZoneByCode returns a storage zone by its code
func (l *Location) GetStorageZoneByCode(code string) *StorageZone {
	for _, zone := range l.StorageZones {
		if zone.Code == code && zone.IsActive {
			return zone
		}
	}
	return nil
}

// GetActiveStorageZones returns all active storage zones
func (l *Location) GetActiveStorageZones() []*StorageZone {
	var activeZones []*StorageZone
	for _, zone := range l.StorageZones {
		if zone.IsActive {
			activeZones = append(activeZones, zone)
		}
	}
	return activeZones
}

// AddStaff adds staff to the location
func (l *Location) AddStaff(staff *LocationStaff) {
	l.Staff = append(l.Staff, staff)
	l.UpdatedAt = time.Now()
	l.Version++
}

// GetActiveStaff returns all active staff at the location
func (l *Location) GetActiveStaff() []*LocationStaff {
	var activeStaff []*LocationStaff
	for _, staff := range l.Staff {
		if staff.IsActive {
			activeStaff = append(activeStaff, staff)
		}
	}
	return activeStaff
}

// AddEquipment adds equipment to the location
func (l *Location) AddEquipment(equipment *LocationEquipment) {
	l.Equipment = append(l.Equipment, equipment)
	l.UpdatedAt = time.Now()
	l.Version++
}

// GetActiveEquipment returns all active equipment at the location
func (l *Location) GetActiveEquipment() []*LocationEquipment {
	var activeEquipment []*LocationEquipment
	for _, equipment := range l.Equipment {
		if equipment.IsActive {
			activeEquipment = append(activeEquipment, equipment)
		}
	}
	return activeEquipment
}

// IsOperational checks if the location is operational
func (l *Location) IsOperational() bool {
	return l.IsActive && l.Status == LocationStatusActive
}

// CanStore checks if the location can store a specific item type
func (l *Location) CanStore(itemCategory ItemCategory, storageRequirements *StorageRequirements) bool {
	if !l.IsOperational() {
		return false
	}
	
	// Check if we have appropriate storage zones
	for _, zone := range l.GetActiveStorageZones() {
		if l.isZoneCompatible(zone, itemCategory, storageRequirements) {
			return true
		}
	}
	
	return false
}

// isZoneCompatible checks if a storage zone is compatible with item requirements
func (l *Location) isZoneCompatible(zone *StorageZone, itemCategory ItemCategory, requirements *StorageRequirements) bool {
	if requirements == nil {
		return true
	}
	
	// Check temperature requirements
	if requirements.Temperature != nil && zone.Environment != nil && zone.Environment.Temperature != nil {
		if requirements.Temperature.MinCelsius < zone.Environment.Temperature.MinCelsius ||
		   requirements.Temperature.MaxCelsius > zone.Environment.Temperature.MaxCelsius {
			return false
		}
	}
	
	// Check humidity requirements
	if requirements.Humidity != nil && zone.Environment != nil && zone.Environment.Humidity != nil {
		if requirements.Humidity.MinPercent < zone.Environment.Humidity.MinPercent ||
		   requirements.Humidity.MaxPercent > zone.Environment.Humidity.MaxPercent {
			return false
		}
	}
	
	// Check hazardous material requirements
	if requirements.Hazardous && zone.Type != ZoneTypeHazardous {
		return false
	}
	
	return true
}

// GetCapacityUtilization returns the current capacity utilization percentage
func (l *Location) GetCapacityUtilization() float64 {
	if l.Capacity == nil {
		return 0.0
	}
	return l.Capacity.CurrentUtilization
}

// HasCapacityFor checks if the location has capacity for additional items
func (l *Location) HasCapacityFor(weight, volume float64, itemCount int) bool {
	if l.Capacity == nil {
		return true // Assume unlimited capacity if not specified
	}
	
	// Check weight capacity
	if l.Capacity.MaxWeight > 0 {
		currentWeight := l.Capacity.MaxWeight * (l.Capacity.CurrentUtilization / 100)
		if currentWeight + weight > l.Capacity.MaxWeight {
			return false
		}
	}
	
	// Check volume capacity
	if l.Capacity.MaxVolume > 0 {
		currentVolume := l.Capacity.MaxVolume * (l.Capacity.CurrentUtilization / 100)
		if currentVolume + volume > l.Capacity.MaxVolume {
			return false
		}
	}
	
	return true
}

// UpdateCapacityUtilization updates the capacity utilization
func (l *Location) UpdateCapacityUtilization(utilizationPercent float64) {
	if l.Capacity == nil {
		l.Capacity = &LocationCapacity{}
	}
	
	l.Capacity.CurrentUtilization = utilizationPercent
	l.UpdatedAt = time.Now()
	l.Version++
}
