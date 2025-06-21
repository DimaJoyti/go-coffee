package tracking

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// IDManager manages unique ID assignment for tracked objects
type IDManager struct {
	logger       *zap.Logger
	nextID       uint64
	usedIDs      map[string]bool
	recycledIDs  []string
	config       IDManagerConfig
	mutex        sync.RWMutex
	stats        *IDManagerStats
}

// IDManagerConfig configures the ID manager
type IDManagerConfig struct {
	IDPrefix        string        // Prefix for generated IDs
	EnableRecycling bool          // Enable ID recycling for deleted tracks
	MaxRecycledIDs  int           // Maximum number of recycled IDs to keep
	IDFormat        IDFormat      // Format for ID generation
	CustomGenerator func() string // Custom ID generator function
}

// IDFormat defines the format for ID generation
type IDFormat int

const (
	IDFormatSequential IDFormat = iota // Sequential numeric IDs
	IDFormatUUID                       // UUID-based IDs
	IDFormatTimestamp                  // Timestamp-based IDs
	IDFormatRandom                     // Random hex IDs
	IDFormatCustom                     // Custom format using provided function
)

// IDManagerStats tracks ID manager statistics
type IDManagerStats struct {
	TotalIDsGenerated uint64
	ActiveIDs         int
	RecycledIDs       int
	MaxActiveIDs      int
	StartTime         time.Time
	LastIDGenerated   time.Time
	mutex             sync.RWMutex
}

// DefaultIDManagerConfig returns default ID manager configuration
func DefaultIDManagerConfig() IDManagerConfig {
	return IDManagerConfig{
		IDPrefix:        "track",
		EnableRecycling: true,
		MaxRecycledIDs:  100,
		IDFormat:        IDFormatSequential,
	}
}

// NewIDManager creates a new ID manager
func NewIDManager(logger *zap.Logger, config IDManagerConfig) *IDManager {
	return &IDManager{
		logger:      logger.With(zap.String("component", "id_manager")),
		nextID:      1,
		usedIDs:     make(map[string]bool),
		recycledIDs: make([]string, 0),
		config:      config,
		stats: &IDManagerStats{
			StartTime: time.Now(),
		},
	}
}

// GenerateID generates a new unique ID
func (im *IDManager) GenerateID() string {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	var id string

	// Try to use recycled ID first if enabled
	if im.config.EnableRecycling && len(im.recycledIDs) > 0 {
		id = im.recycledIDs[0]
		im.recycledIDs = im.recycledIDs[1:]
		im.logger.Debug("Reusing recycled ID", zap.String("id", id))
	} else {
		// Generate new ID based on format
		id = im.generateNewID()
	}

	// Mark ID as used
	im.usedIDs[id] = true

	// Update statistics
	im.updateStats(id)

	im.logger.Debug("Generated ID", zap.String("id", id))
	return id
}

// ReleaseID releases an ID for potential recycling
func (im *IDManager) ReleaseID(id string) {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	// Check if ID exists
	if !im.usedIDs[id] {
		im.logger.Warn("Attempting to release non-existent ID", zap.String("id", id))
		return
	}

	// Remove from used IDs
	delete(im.usedIDs, id)

	// Add to recycled IDs if recycling is enabled
	if im.config.EnableRecycling {
		if len(im.recycledIDs) < im.config.MaxRecycledIDs {
			im.recycledIDs = append(im.recycledIDs, id)
			im.logger.Debug("ID added to recycling pool", zap.String("id", id))
		} else {
			im.logger.Debug("Recycling pool full, discarding ID", zap.String("id", id))
		}
	}

	im.logger.Debug("Released ID", zap.String("id", id))
}

// IsIDUsed checks if an ID is currently in use
func (im *IDManager) IsIDUsed(id string) bool {
	im.mutex.RLock()
	defer im.mutex.RUnlock()
	return im.usedIDs[id]
}

// GetActiveIDs returns all currently active IDs
func (im *IDManager) GetActiveIDs() []string {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	ids := make([]string, 0, len(im.usedIDs))
	for id := range im.usedIDs {
		ids = append(ids, id)
	}
	return ids
}

// GetRecycledIDs returns all recycled IDs
func (im *IDManager) GetRecycledIDs() []string {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	recycled := make([]string, len(im.recycledIDs))
	copy(recycled, im.recycledIDs)
	return recycled
}

// GetStats returns ID manager statistics
func (im *IDManager) GetStats() *IDManagerStats {
	im.stats.mutex.RLock()
	defer im.stats.mutex.RUnlock()

	im.mutex.RLock()
	activeIDs := len(im.usedIDs)
	recycledIDs := len(im.recycledIDs)
	im.mutex.RUnlock()

	return &IDManagerStats{
		TotalIDsGenerated: im.stats.TotalIDsGenerated,
		ActiveIDs:         activeIDs,
		RecycledIDs:       recycledIDs,
		MaxActiveIDs:      im.stats.MaxActiveIDs,
		StartTime:         im.stats.StartTime,
		LastIDGenerated:   im.stats.LastIDGenerated,
	}
}

// ClearRecycledIDs clears all recycled IDs
func (im *IDManager) ClearRecycledIDs() {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	cleared := len(im.recycledIDs)
	im.recycledIDs = im.recycledIDs[:0]

	im.logger.Info("Cleared recycled IDs", zap.Int("count", cleared))
}

// UpdateConfig updates the ID manager configuration
func (im *IDManager) UpdateConfig(config IDManagerConfig) {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	im.config = config

	// Trim recycled IDs if max limit changed
	if len(im.recycledIDs) > config.MaxRecycledIDs {
		im.recycledIDs = im.recycledIDs[:config.MaxRecycledIDs]
	}

	im.logger.Info("ID manager configuration updated",
		zap.String("prefix", config.IDPrefix),
		zap.Bool("recycling", config.EnableRecycling),
		zap.Int("max_recycled", config.MaxRecycledIDs))
}

// generateNewID generates a new ID based on the configured format
func (im *IDManager) generateNewID() string {
	switch im.config.IDFormat {
	case IDFormatSequential:
		return im.generateSequentialID()
	case IDFormatUUID:
		return im.generateUUIDID()
	case IDFormatTimestamp:
		return im.generateTimestampID()
	case IDFormatRandom:
		return im.generateRandomID()
	case IDFormatCustom:
		if im.config.CustomGenerator != nil {
			return im.config.CustomGenerator()
		}
		// Fallback to sequential if custom generator is not provided
		return im.generateSequentialID()
	default:
		return im.generateSequentialID()
	}
}

// generateSequentialID generates a sequential numeric ID
func (im *IDManager) generateSequentialID() string {
	id := fmt.Sprintf("%s_%d", im.config.IDPrefix, im.nextID)
	im.nextID++
	return id
}

// generateUUIDID generates a UUID-based ID
func (im *IDManager) generateUUIDID() string {
	// Simple UUID v4 implementation
	bytes := make([]byte, 16)
	rand.Read(bytes)
	
	// Set version (4) and variant bits
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
	
	return fmt.Sprintf("%s_%s", im.config.IDPrefix, uuid)
}

// generateTimestampID generates a timestamp-based ID
func (im *IDManager) generateTimestampID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d", im.config.IDPrefix, timestamp)
}

// generateRandomID generates a random hex ID
func (im *IDManager) generateRandomID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	randomHex := hex.EncodeToString(bytes)
	return fmt.Sprintf("%s_%s", im.config.IDPrefix, randomHex)
}

// updateStats updates ID manager statistics
func (im *IDManager) updateStats(id string) {
	im.stats.mutex.Lock()
	defer im.stats.mutex.Unlock()

	im.stats.TotalIDsGenerated++
	im.stats.LastIDGenerated = time.Now()

	// Update max active IDs
	activeCount := len(im.usedIDs)
	if activeCount > im.stats.MaxActiveIDs {
		im.stats.MaxActiveIDs = activeCount
	}
}

// ValidateID validates an ID format
func (im *IDManager) ValidateID(id string) bool {
	if id == "" {
		return false
	}

	// Basic validation - check if it starts with the configured prefix
	expectedPrefix := im.config.IDPrefix + "_"
	if len(id) < len(expectedPrefix) {
		return false
	}

	return id[:len(expectedPrefix)] == expectedPrefix
}

// GetIDInfo returns information about an ID
func (im *IDManager) GetIDInfo(id string) map[string]interface{} {
	im.mutex.RLock()
	defer im.mutex.RUnlock()

	info := map[string]interface{}{
		"id":         id,
		"valid":      im.ValidateID(id),
		"active":     im.usedIDs[id],
		"recycled":   false,
		"format":     im.config.IDFormat.String(),
	}

	// Check if ID is in recycled pool
	for _, recycledID := range im.recycledIDs {
		if recycledID == id {
			info["recycled"] = true
			break
		}
	}

	return info
}

// String returns string representation of ID format
func (f IDFormat) String() string {
	switch f {
	case IDFormatSequential:
		return "sequential"
	case IDFormatUUID:
		return "uuid"
	case IDFormatTimestamp:
		return "timestamp"
	case IDFormatRandom:
		return "random"
	case IDFormatCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// Close closes the ID manager and cleans up resources
func (im *IDManager) Close() error {
	im.mutex.Lock()
	defer im.mutex.Unlock()

	im.logger.Info("Closing ID manager",
		zap.Int("active_ids", len(im.usedIDs)),
		zap.Int("recycled_ids", len(im.recycledIDs)))

	// Clear all data
	im.usedIDs = make(map[string]bool)
	im.recycledIDs = im.recycledIDs[:0]

	im.logger.Info("ID manager closed")
	return nil
}
