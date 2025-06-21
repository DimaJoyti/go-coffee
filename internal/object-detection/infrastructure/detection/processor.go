//go:build opencv
// +build opencv

package detection

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"go.uber.org/zap"
)

// ResultProcessor processes and filters detection results
type ResultProcessor struct {
	logger              *zap.Logger
	confidenceThreshold float64
	nmsThreshold        float64
	maxDetections       int
	classFilter         map[string]bool
	minObjectSize       int
	maxObjectSize       int
	mutex               sync.RWMutex
}

// ProcessorConfig configures the result processor
type ProcessorConfig struct {
	ConfidenceThreshold float64
	NMSThreshold        float64
	MaxDetections       int
	ClassFilter         []string // Empty means all classes allowed
	MinObjectSize       int      // Minimum bounding box area
	MaxObjectSize       int      // Maximum bounding box area (0 = no limit)
}

// DefaultProcessorConfig returns default processor configuration
func DefaultProcessorConfig() ProcessorConfig {
	return ProcessorConfig{
		ConfidenceThreshold: 0.5,
		NMSThreshold:        0.4,
		MaxDetections:       100,
		ClassFilter:         []string{}, // Allow all classes
		MinObjectSize:       100,        // Minimum 10x10 pixels
		MaxObjectSize:       0,          // No maximum limit
	}
}

// NewResultProcessor creates a new result processor
func NewResultProcessor(logger *zap.Logger, config ProcessorConfig) *ResultProcessor {
	classFilter := make(map[string]bool)
	for _, class := range config.ClassFilter {
		classFilter[class] = true
	}

	return &ResultProcessor{
		logger:              logger.With(zap.String("component", "result_processor")),
		confidenceThreshold: config.ConfidenceThreshold,
		nmsThreshold:        config.NMSThreshold,
		maxDetections:       config.MaxDetections,
		classFilter:         classFilter,
		minObjectSize:       config.MinObjectSize,
		maxObjectSize:       config.MaxObjectSize,
	}
}

// ProcessDetections processes raw detections and applies filtering
func (rp *ResultProcessor) ProcessDetections(ctx context.Context, detections []domain.DetectedObject) ([]domain.DetectedObject, error) {
	if len(detections) == 0 {
		return detections, nil
	}

	rp.logger.Debug("Processing detections", zap.Int("input_count", len(detections)))

	// Apply confidence filtering
	filtered := rp.filterByConfidence(detections)
	rp.logger.Debug("After confidence filtering", zap.Int("count", len(filtered)))

	// Apply class filtering
	filtered = rp.filterByClass(filtered)
	rp.logger.Debug("After class filtering", zap.Int("count", len(filtered)))

	// Apply size filtering
	filtered = rp.filterBySize(filtered)
	rp.logger.Debug("After size filtering", zap.Int("count", len(filtered)))

	// Apply Non-Maximum Suppression
	filtered = rp.applyNMS(filtered)
	rp.logger.Debug("After NMS", zap.Int("count", len(filtered)))

	// Limit number of detections
	if len(filtered) > rp.maxDetections {
		// Sort by confidence and take top N
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Confidence > filtered[j].Confidence
		})
		filtered = filtered[:rp.maxDetections]
		rp.logger.Debug("After max detections limit", zap.Int("count", len(filtered)))
	}

	// Assign detection IDs and update timestamps
	for i := range filtered {
		if filtered[i].ID == "" {
			filtered[i].ID = generateDetectionID()
		}
		filtered[i].Timestamp = time.Now()
	}

	rp.logger.Debug("Detection processing completed", 
		zap.Int("final_count", len(filtered)),
		zap.Int("original_count", len(detections)))

	return filtered, nil
}

// filterByConfidence filters detections by confidence threshold
func (rp *ResultProcessor) filterByConfidence(detections []domain.DetectedObject) []domain.DetectedObject {
	rp.mutex.RLock()
	threshold := rp.confidenceThreshold
	rp.mutex.RUnlock()

	var filtered []domain.DetectedObject
	for _, detection := range detections {
		if detection.Confidence >= threshold {
			filtered = append(filtered, detection)
		}
	}
	return filtered
}

// filterByClass filters detections by allowed classes
func (rp *ResultProcessor) filterByClass(detections []domain.DetectedObject) []domain.DetectedObject {
	rp.mutex.RLock()
	classFilter := rp.classFilter
	rp.mutex.RUnlock()

	// If no class filter is set, allow all classes
	if len(classFilter) == 0 {
		return detections
	}

	var filtered []domain.DetectedObject
	for _, detection := range detections {
		if classFilter[detection.Class] {
			filtered = append(filtered, detection)
		}
	}
	return filtered
}

// filterBySize filters detections by bounding box size
func (rp *ResultProcessor) filterBySize(detections []domain.DetectedObject) []domain.DetectedObject {
	rp.mutex.RLock()
	minSize := rp.minObjectSize
	maxSize := rp.maxObjectSize
	rp.mutex.RUnlock()

	var filtered []domain.DetectedObject
	for _, detection := range detections {
		area := detection.BoundingBox.Width * detection.BoundingBox.Height
		
		if area < minSize {
			continue
		}
		
		if maxSize > 0 && area > maxSize {
			continue
		}
		
		filtered = append(filtered, detection)
	}
	return filtered
}

// applyNMS applies Non-Maximum Suppression to remove overlapping detections
func (rp *ResultProcessor) applyNMS(detections []domain.DetectedObject) []domain.DetectedObject {
	if len(detections) <= 1 {
		return detections
	}

	rp.mutex.RLock()
	nmsThreshold := rp.nmsThreshold
	rp.mutex.RUnlock()

	// Group detections by class
	classBuckets := make(map[string][]domain.DetectedObject)
	for _, detection := range detections {
		classBuckets[detection.Class] = append(classBuckets[detection.Class], detection)
	}

	var result []domain.DetectedObject

	// Apply NMS per class
	for _, classDetections := range classBuckets {
		nmsResult := rp.nmsPerClass(classDetections, nmsThreshold)
		result = append(result, nmsResult...)
	}

	return result
}

// nmsPerClass applies NMS for a single class
func (rp *ResultProcessor) nmsPerClass(detections []domain.DetectedObject, threshold float64) []domain.DetectedObject {
	if len(detections) <= 1 {
		return detections
	}

	// Sort by confidence (descending)
	sort.Slice(detections, func(i, j int) bool {
		return detections[i].Confidence > detections[j].Confidence
	})

	var result []domain.DetectedObject
	suppressed := make([]bool, len(detections))

	for i := 0; i < len(detections); i++ {
		if suppressed[i] {
			continue
		}

		result = append(result, detections[i])

		// Suppress overlapping detections
		for j := i + 1; j < len(detections); j++ {
			if suppressed[j] {
				continue
			}

			iou := rp.calculateIOU(detections[i].BoundingBox, detections[j].BoundingBox)
			if iou > threshold {
				suppressed[j] = true
			}
		}
	}

	return result
}

// calculateIOU calculates Intersection over Union between two bounding boxes
func (rp *ResultProcessor) calculateIOU(box1, box2 domain.Rectangle) float64 {
	// Calculate intersection coordinates
	x1 := max(box1.X, box2.X)
	y1 := max(box1.Y, box2.Y)
	x2 := min(box1.X+box1.Width, box2.X+box2.Width)
	y2 := min(box1.Y+box1.Height, box2.Y+box2.Height)

	// Check if there's no intersection
	if x2 <= x1 || y2 <= y1 {
		return 0.0
	}

	// Calculate intersection area
	intersection := float64((x2 - x1) * (y2 - y1))

	// Calculate union area
	area1 := float64(box1.Width * box1.Height)
	area2 := float64(box2.Width * box2.Height)
	union := area1 + area2 - intersection

	if union <= 0 {
		return 0.0
	}

	return intersection / union
}

// UpdateConfig updates the processor configuration
func (rp *ResultProcessor) UpdateConfig(config ProcessorConfig) {
	rp.mutex.Lock()
	defer rp.mutex.Unlock()

	rp.confidenceThreshold = config.ConfidenceThreshold
	rp.nmsThreshold = config.NMSThreshold
	rp.maxDetections = config.MaxDetections
	rp.minObjectSize = config.MinObjectSize
	rp.maxObjectSize = config.MaxObjectSize

	// Update class filter
	rp.classFilter = make(map[string]bool)
	for _, class := range config.ClassFilter {
		rp.classFilter[class] = true
	}

	rp.logger.Info("Processor configuration updated",
		zap.Float64("confidence_threshold", config.ConfidenceThreshold),
		zap.Float64("nms_threshold", config.NMSThreshold),
		zap.Int("max_detections", config.MaxDetections),
		zap.Strings("class_filter", config.ClassFilter))
}

// GetConfig returns the current processor configuration
func (rp *ResultProcessor) GetConfig() ProcessorConfig {
	rp.mutex.RLock()
	defer rp.mutex.RUnlock()

	classFilter := make([]string, 0, len(rp.classFilter))
	for class := range rp.classFilter {
		classFilter = append(classFilter, class)
	}

	return ProcessorConfig{
		ConfidenceThreshold: rp.confidenceThreshold,
		NMSThreshold:        rp.nmsThreshold,
		MaxDetections:       rp.maxDetections,
		ClassFilter:         classFilter,
		MinObjectSize:       rp.minObjectSize,
		MaxObjectSize:       rp.maxObjectSize,
	}
}

// SetConfidenceThreshold updates the confidence threshold
func (rp *ResultProcessor) SetConfidenceThreshold(threshold float64) {
	if threshold < 0 || threshold > 1 {
		rp.logger.Warn("Invalid confidence threshold", zap.Float64("threshold", threshold))
		return
	}

	rp.mutex.Lock()
	defer rp.mutex.Unlock()

	rp.confidenceThreshold = threshold
	rp.logger.Info("Confidence threshold updated", zap.Float64("threshold", threshold))
}

// SetNMSThreshold updates the NMS threshold
func (rp *ResultProcessor) SetNMSThreshold(threshold float64) {
	if threshold < 0 || threshold > 1 {
		rp.logger.Warn("Invalid NMS threshold", zap.Float64("threshold", threshold))
		return
	}

	rp.mutex.Lock()
	defer rp.mutex.Unlock()

	rp.nmsThreshold = threshold
	rp.logger.Info("NMS threshold updated", zap.Float64("threshold", threshold))
}

// SetClassFilter updates the class filter
func (rp *ResultProcessor) SetClassFilter(classes []string) {
	rp.mutex.Lock()
	defer rp.mutex.Unlock()

	rp.classFilter = make(map[string]bool)
	for _, class := range classes {
		rp.classFilter[class] = true
	}

	rp.logger.Info("Class filter updated", zap.Strings("classes", classes))
}

// GetStats returns processing statistics
func (rp *ResultProcessor) GetStats() map[string]interface{} {
	rp.mutex.RLock()
	defer rp.mutex.RUnlock()

	classFilter := make([]string, 0, len(rp.classFilter))
	for class := range rp.classFilter {
		classFilter = append(classFilter, class)
	}

	return map[string]interface{}{
		"confidence_threshold": rp.confidenceThreshold,
		"nms_threshold":        rp.nmsThreshold,
		"max_detections":       rp.maxDetections,
		"class_filter":         classFilter,
		"min_object_size":      rp.minObjectSize,
		"max_object_size":      rp.maxObjectSize,
	}
}

// generateDetectionID generates a unique detection ID
func generateDetectionID() string {
	return fmt.Sprintf("det_%d", time.Now().UnixNano())
}

// Helper functions for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
